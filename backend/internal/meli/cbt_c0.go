package meli

import (
 "context"
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "strings"
 "time"
)

// PublicationEvidence contains only sanitized C0 account metadata.
// It never carries a bearer token, seller/child IDs, or original API payload.
type PublicationEvidence struct {
 Mode string `json:"publication_mode"`
 EvidenceLevel string `json:"mode_evidence_level"`
 SellAllowed string `json:"sell_allowed"`
 UserProductSeller string `json:"user_product_seller"`
 WarehouseManagement string `json:"warehouse_management"`
 Sites []PublicationSiteEvidence `json:"sites"`
 CheckedAt time.Time `json:"checked_at"`
}

type PublicationSiteEvidence struct {
 SiteID string `json:"site_id"`
 LogisticType string `json:"logistic_type"`
 BusinessModel string `json:"business_model"`
 PricingModel string `json:"pricing_model"`
 UserProduct string `json:"user_product"`
}

const publicationUnverified = "undetermined"

// ProbePublicationMode uses the existing owner-scoped OAuth refresh. Its
// marketplace requests are exclusively GET. The only write is local,
// owner-scoped storage of the sanitized model evidence.
func (s *Service) ProbePublicationMode(ctx context.Context, owner, storeID string) (PublicationEvidence, error) {
 auth, err := s.RequireStoreConnection(ctx, owner, storeID)
 if err != nil {
  return PublicationEvidence{}, err
 }
 if !auth.Verified || len(auth.Markets) == 0 {
  return PublicationEvidence{}, ErrSiteNotAuthorized
 }
 ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
 defer cancel()
 evidence, err := s.collectPublicationEvidence(ctx, auth)
 if err != nil {
  return PublicationEvidence{}, err
 }
 packed, err := json.Marshal(evidence)
 if err != nil { return PublicationEvidence{}, err }
 command, err := s.db.Exec(ctx, `INSERT INTO app.meli_cbt_c0_evidence
   (store_id,owner_user_id,publication_mode,evidence,checked_at)
 SELECT id,owner_user_id,$4,$5::jsonb,$6
 FROM app.meli_store_connections
 WHERE id=$1::uuid AND owner_user_id=$2::uuid AND seller_id=$3 AND status='connected'
 ON CONFLICT (store_id) DO UPDATE SET
 publication_mode=EXCLUDED.publication_mode,evidence=EXCLUDED.evidence,checked_at=EXCLUDED.checked_at
 WHERE app.meli_cbt_c0_evidence.owner_user_id=EXCLUDED.owner_user_id`,
  storeID, owner, auth.SellerID, evidence.Mode, string(packed), evidence.CheckedAt)
 if err != nil { return PublicationEvidence{}, errors.New("could not save CBT account evidence") }
 if command.RowsAffected()!=1 { return PublicationEvidence{}, ErrStoreNotFound }
 return evidence, nil
}

func (s *Service) collectPublicationEvidence(ctx context.Context, auth AuthorizedStore) (PublicationEvidence, error) {
 var self struct {
  ID json.Number `json:"id"`
  SiteID string `json:"site_id"`
  Status struct { Sell struct { Allow *bool `json:"allow"` } `json:"sell"` } `json:"status"`
 }
 if err := s.readC0JSON(ctx,"/users/me",auth.AccessToken,&self); err!=nil { return PublicationEvidence{},err }
 if self.ID.String()!=auth.SellerID || self.SiteID!="CBT" { return PublicationEvidence{},errors.New("CBT identity mismatch") }

 var user struct {
  ID json.Number `json:"id"`
  SiteID string `json:"site_id"`
  Tags []string `json:"tags"`
 }
 sellerPath := url.PathEscape(auth.SellerID)
 if err := s.readC0JSON(ctx,"/users/"+sellerPath,auth.AccessToken,&user); err!=nil { return PublicationEvidence{},err }
 if (user.ID!="" && user.ID.String()!=auth.SellerID) || (user.SiteID!="" && user.SiteID!="CBT") { return PublicationEvidence{},errors.New("CBT tags identity mismatch") }

 var mapping struct {
  UserID json.Number `json:"user_id"`
  SiteID string `json:"site_id"`
  Marketplaces []struct {
   UserID json.Number `json:"user_id"`
   SiteID string `json:"site_id"`
   LogisticType string `json:"logistic_type"`
   BusinessModel string `json:"business_model"`
   PricingModel string `json:"pricing_model"`
   UserProduct *bool `json:"user_product"`
  } `json:"marketplaces"`
 }
 if err := s.readC0JSON(ctx,"/marketplace/users/"+sellerPath,auth.AccessToken,&mapping); err!=nil { return PublicationEvidence{},err }
 if mapping.UserID.String()!=auth.SellerID || mapping.SiteID!="CBT" || mapping.Marketplaces==nil || len(mapping.Marketplaces)!=len(auth.Markets) || len(mapping.Marketplaces)>32 {
  return PublicationEvidence{},errors.New("CBT marketplace identity or mapping changed: verify store authorization")
 }
 expected:=make(map[string]string,len(auth.Markets))
 for _,m:=range auth.Markets { expected[m.SiteID+"|"+m.LogisticType]=m.UserID.String() }
 seen:=make(map[string]bool,len(mapping.Marketplaces))
 evidence:=PublicationEvidence{
  EvidenceLevel:"c0_read_only_not_publish_approved",
  SellAllowed:publicationCapability(self.Status.Sell.Allow),
  UserProductSeller:publicationTag(user.Tags,"user_product_seller"),
  WarehouseManagement:publicationTag(user.Tags,"warehouse_management"),
  CheckedAt:time.Now().UTC(),
  Sites:make([]PublicationSiteEvidence,0,len(mapping.Marketplaces)),
 }
 for _,m:=range mapping.Marketplaces {
  key:=m.SiteID+"|"+m.LogisticType
  child,ok:=expected[key]
  if !ok || seen[key] || child!=m.UserID.String() { return PublicationEvidence{},errors.New("CBT marketplace mapping differs from authorized store") }
  seen[key]=true
  evidence.Sites=append(evidence.Sites,PublicationSiteEvidence{
   SiteID:m.SiteID,LogisticType:m.LogisticType,
   BusinessModel:publicationBusinessModel(m.BusinessModel),
   PricingModel:publicationPricingModel(m.PricingModel),
   UserProduct:publicationCapability(m.UserProduct),
  })
 }
 evidence.Mode=classifyPublicationEvidence(evidence)
 return evidence,nil
}

func (s *Service) readC0JSON(ctx context.Context, path, token string, out any) error {
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,s.apiURL+path,nil)
 if err!=nil { return errors.New("invalid CBT read request") }
 req.Header.Set("Authorization","Bearer "+token)
 req.Header.Set("Accept","application/json")
 res,err:=s.client.Do(req)
 if err!=nil { return errors.New("CBT account read unavailable") }
 defer res.Body.Close()
 if res.StatusCode!=http.StatusOK { return fmt.Errorf("CBT account read rejected (HTTP %d)",res.StatusCode) }
 body,err:=io.ReadAll(io.LimitReader(res.Body,128*1024+1))
 if err!=nil || len(body)>128*1024 { return errors.New("invalid CBT response size") }
 if err=json.Unmarshal(body,out);err!=nil { return errors.New("invalid CBT response shape") }
 return nil
}

func publicationCapability(v *bool) string {
 if v==nil { return "unverified" }
 if *v { return "enabled" }
 return "disabled"
}
func publicationTag(tags []string, name string) string {
 if tags==nil { return "unverified" }
 for _,tag:=range tags {if tag==name{return "present"}}
 return "absent"
}
func publicationBusinessModel(v string) string {
 if v=="" {return "unverified"}
 if v=="CBT CN Fulfillment Managed" {return v}
 return "other"
}
func publicationPricingModel(v string) string {
 switch v {
 case "price","listing_price","net_proceeds":return v
 case "":return "unverified"
 default:return "other"
 }
}

// These are routing candidates, not permission to POST.
// Mercado Libre identifies Fully Managed if ANY child has the exact
// business_model. Legacy always requires separate explicit confirmation.
func classifyPublicationEvidence(e PublicationEvidence) string {
 if e.SellAllowed!="enabled" || len(e.Sites)==0 {return publicationUnverified}
 hasManaged:=false
 hasUnknown:=false
 hasDisabledUP:=false
 hasEnabledUP:=false
 for _,site:=range e.Sites {
  if site.BusinessModel=="CBT CN Fulfillment Managed" {hasManaged=true}
  if site.BusinessModel=="other" {hasUnknown=true}
  if site.UserProduct=="disabled" {hasDisabledUP=true}
  if site.UserProduct=="enabled" {hasEnabledUP=true}
 }
 if hasManaged {return "fully_managed"}
 if hasUnknown {return publicationUnverified}
 if e.UserProductSeller=="present" && !hasDisabledUP {return "standard_up"}
 if e.UserProductSeller=="absent" && !hasEnabledUP {return "legacy_candidate"}
 return publicationUnverified
}

// C0 evidence is invalidated when an OAuth connection is replaced.
func (s *Service) clearC0Evidence(ctx context.Context, owner, storeID string) error {
 if !validUUID(storeID) {return ErrStoreNotFound}
 _,err:=s.db.Exec(ctx,`DELETE FROM app.meli_cbt_c0_evidence WHERE store_id=$1::uuid AND owner_user_id=$2::uuid`,storeID,owner)
 return err
}

// A store may have a previous model recorded. Display it only as a historical
// C0 probe, never as proof of current write capabilities.
func publicationModeLabel(mode string) string {
 if strings.TrimSpace(mode)=="" {return publicationUnverified}
 return mode
}
