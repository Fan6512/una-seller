package meli

import (
 "context"
 "encoding/json"
 "fmt"
 "net/http"
 "net/http/httptest"
 "reflect"
 "strings"
 "testing"
)

func TestCBTC0ReadOnlySanitizedAndManaged(t *testing.T) {
 paths:=make([]string,0,3)
 token:="synthetic-only-access"
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
  paths=append(paths,r.URL.Path)
  if r.Method!=http.MethodGet || r.Header.Get("Authorization")!="Bearer "+token || r.URL.RawQuery!="" {
   t.Error("CBT probe must issue only bearer-authenticated GETs")
  }
  switch r.URL.Path {
  case "/users/me": fmt.Fprint(w,`{"id":7000,"site_id":"CBT","status":{"sell":{"allow":true}},"email":"private-email"}`)
  case "/users/7000": fmt.Fprint(w,`{"id":7000,"site_id":"CBT","tags":["user_product_seller","warehouse_management","secret-tag"]}`)
  case "/marketplace/users/7000": fmt.Fprint(w,`{"user_id":7000,"site_id":"CBT","marketplaces":[{"user_id":8000,"site_id":"MLM","logistic_type":"remote","business_model":"CBT CN Fulfillment Managed","pricing_model":"net_proceeds","user_product":true,"token":"private"}]}`)
  default:w.WriteHeader(http.StatusNotFound)
  }
 }))
 defer server.Close()
 service:=&Service{client:server.Client(),apiURL:server.URL}
 auth:=AuthorizedStore{SellerID:"7000",AccessToken:token,Verified:true,Markets:[]Marketplace{{UserID:json.Number("8000"),SiteID:"MLM",LogisticType:"remote"}}}
 e,err:=service.collectPublicationEvidence(context.Background(),auth)
 if err!=nil {t.Fatal(err)}
 if e.Mode!="fully_managed" || e.EvidenceLevel!="c0_read_only_not_publish_approved" || e.UserProductSeller!="present" || e.Sites[0].PricingModel!="net_proceeds" {t.Fatalf("incorrect evidence: %+v",e)}
 if !reflect.DeepEqual(paths,[]string{"/users/me","/users/7000","/marketplace/users/7000"}) {t.Fatalf("paths: %v",paths)}
 raw,err:=json.Marshal(e);if err!=nil{t.Fatal(err)}
 for _,private:=range []string{"7000","8000","synthetic-only-access","private-email","secret-tag"} {
  if strings.Contains(string(raw),private){t.Fatalf("private identity leaked: %s",private)}
 }
}

func TestCBTC0ClassificationDoesNotInventLegacyEligibility(t *testing.T) {
 cases:=[]struct{
  name string
  tags string
  sell string
  sites []PublicationSiteEvidence
  want string
 }{
  {"managed", "present","enabled",[]PublicationSiteEvidence{{BusinessModel:"CBT CN Fulfillment Managed"}},"fully_managed"},
  {"standard", "present","enabled",[]PublicationSiteEvidence{{BusinessModel:"unverified",UserProduct:"enabled"}},"standard_up"},
  {"standard_when_flag_missing", "present","enabled",[]PublicationSiteEvidence{{BusinessModel:"unverified",UserProduct:"unverified"}},"standard_up"},
  {"contradictory_up", "present","enabled",[]PublicationSiteEvidence{{BusinessModel:"unverified",UserProduct:"disabled"}},"undetermined"},
  {"legacy_is_not_approval", "absent","enabled",[]PublicationSiteEvidence{{BusinessModel:"unverified",UserProduct:"disabled"}},"legacy_candidate"},
  {"unknown_business", "present","enabled",[]PublicationSiteEvidence{{BusinessModel:"other",UserProduct:"enabled"}},"undetermined"},
  {"disabled_sale", "present","disabled",[]PublicationSiteEvidence{{BusinessModel:"CBT CN Fulfillment Managed"}},"undetermined"},
  {"missing_tag", "unverified","enabled",[]PublicationSiteEvidence{{BusinessModel:"unverified"}},"undetermined"},
  {"no_sites", "present","enabled",nil,"undetermined"},
 }
 for _,tc:=range cases { t.Run(tc.name,func(t *testing.T){
  got:=classifyPublicationEvidence(PublicationEvidence{UserProductSeller:tc.tags,SellAllowed:tc.sell,Sites:tc.sites})
  if got!=tc.want {t.Fatalf("got %s, want %s",got,tc.want)}
 }) }
}

func TestCBTC0RejectsMismatchedChildAndDoesNotExposeBody(t *testing.T) {
 server:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  switch r.URL.Path{
  case "/users/me":fmt.Fprint(w,`{"id":7000,"site_id":"CBT","status":{"sell":{"allow":true}}}`)
  case "/users/7000":fmt.Fprint(w,`{"tags":[]}`)
  case "/marketplace/users/7000":fmt.Fprint(w,`{"user_id":7000,"site_id":"CBT","marketplaces":[{"user_id":8999,"site_id":"MLM","logistic_type":"remote","business_model":"secret"}]}`)
  }
 }))
 defer server.Close()
 service:=&Service{client:server.Client(),apiURL:server.URL}
 auth:=AuthorizedStore{SellerID:"7000",AccessToken:"secret-token",Verified:true,Markets:[]Marketplace{{UserID:json.Number("8000"),SiteID:"MLM",LogisticType:"remote"}}}
 _,err:=service.collectPublicationEvidence(context.Background(),auth)
 if err==nil||strings.Contains(err.Error(),"8999")||strings.Contains(err.Error(),"secret") {t.Fatalf("unsafe validation: %v",err)}
}

func TestCBTC0FiltersUnknownBusinessAndPricing(t *testing.T) {
 if got:=publicationBusinessModel("provider-secret");got!="other"{t.Fatalf("business leaked: %s",got)}
 if got:=publicationPricingModel("provider-secret");got!="other"{t.Fatalf("pricing leaked: %s",got)}
 if got:=publicationCapability(nil);got!="unverified"{t.Fatalf("capability: %s",got)}
 if got:=publicationTag(nil,"user_product_seller");got!="unverified"{t.Fatalf("tag: %s",got)}
}
