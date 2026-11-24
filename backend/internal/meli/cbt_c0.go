package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"log"
	"time"
)

// PublicationEvidence contains only sanitized C0 account metadata.
// It never carries a bearer token, seller/child IDs, or original API payload.
type PublicationEvidence struct {
	Mode                string                    `json:"publication_mode"`
	EvidenceLevel       string                    `json:"mode_evidence_level"`
	SellAllowed         string                    `json:"sell_allowed"`
	UserProductSeller   string                    `json:"user_product_seller"`
	WarehouseManagement string                    `json:"warehouse_management"`
	Sites               []PublicationSiteEvidence `json:"sites"`
	CheckedAt           time.Time                 `json:"checked_at"`
}

type PublicationSiteEvidence struct {
	SiteID        string `json:"site_id"`
	LogisticType  string `json:"logistic_type"`
	BusinessModel string `json:"business_model"`
	PricingModel  string `json:"pricing_model"`
	UserProduct   string `json:"user_product"`
}

const publicationUnverified = "undetermined"

var ErrCBTProbeCooldown = errors.New("CBT C0 probe cooldown is active")
var ErrCBTProbeStale = errors.New("CBT C0 store authorization or identity changed")

// Each authorized store permits at most one manual C0 probe per 10 seconds
// within this Go process. No Redis or background scheduler is needed.
func (s *Service) claimC0Probe(storeID string) error {
	s.probeMu.Lock()
	defer s.probeMu.Unlock()
	if s.probeNext == nil {
		s.probeNext = make(map[string]time.Time)
	}
	now := time.Now()
	if deadline, ok := s.probeNext[storeID]; ok && now.Before(deadline) {
		return ErrCBTProbeCooldown
	}
	if len(s.probeNext) > 512 {
		for key, deadline := range s.probeNext {
			if !now.Before(deadline) {
				delete(s.probeNext, key)
			}
		}
	}
	s.probeNext[storeID] = now.Add(10 * time.Second)
	return nil
}

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
	if err := s.claimC0Probe(storeID); err != nil {
		return PublicationEvidence{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 40*time.Second)
	defer cancel()
	evidence, err := s.collectPublicationEvidence(ctx, auth)
	if err != nil {
		return PublicationEvidence{}, err
	}
	packed, err := json.Marshal(evidence)
	if err != nil {
		return PublicationEvidence{}, err
	}
	command, err := s.db.Exec(ctx, `INSERT INTO app.meli_cbt_c0_evidence
   (store_id,owner_user_id,publication_mode,evidence,checked_at)
 SELECT id,owner_user_id,$4,$5::jsonb,$6
 FROM app.meli_store_connections
 WHERE id=$1::uuid AND owner_user_id=$2::uuid AND seller_id=$3 AND status='connected'
 ON CONFLICT (store_id) DO UPDATE SET
 publication_mode=EXCLUDED.publication_mode,evidence=EXCLUDED.evidence,checked_at=EXCLUDED.checked_at
 WHERE app.meli_cbt_c0_evidence.owner_user_id=EXCLUDED.owner_user_id`,
		storeID, owner, auth.SellerID, evidence.Mode, string(packed), evidence.CheckedAt)
	if err != nil {
		return PublicationEvidence{}, errors.New("could not save CBT account evidence")
	}
	if command.RowsAffected() != 1 {
		return PublicationEvidence{}, ErrCBTProbeStale
	}
	return evidence, nil
}

func (s *Service) collectPublicationEvidence(ctx context.Context, auth AuthorizedStore) (PublicationEvidence, error) {
	var self struct {
		ID     json.Number `json:"id"`
		SiteID string      `json:"site_id"`
		Status struct {
			Sell struct {
				Allow *bool `json:"allow"`
			} `json:"sell"`
		} `json:"status"`
	}
	if err := s.readC0JSON(ctx, "/users/me", auth.AccessToken, &self); err != nil {
		return PublicationEvidence{}, fmt.Errorf("CBT users/me: %w", err)
	}
	if self.ID.String() != auth.SellerID {
		return PublicationEvidence{}, errors.New("CBT users/me: seller id mismatch")
	}
	if self.SiteID != "CBT" {
		return PublicationEvidence{}, errors.New("CBT users/me: site_id is not CBT")
	}
	if self.Status.Sell.Allow == nil {
		log.Print("CBT C0 users/me: status.sell.allow missing; publication remains unverified")
	}

	var user struct {
		ID     json.Number `json:"id"`
		SiteID string      `json:"site_id"`
		Tags   []string    `json:"tags"`
	}
	sellerPath := url.PathEscape(auth.SellerID)
	if err := s.readC0JSON(ctx, "/users/"+sellerPath, auth.AccessToken, &user); err != nil {
		return PublicationEvidence{}, fmt.Errorf("CBT seller tags: %w", err)
	}
	if (user.ID != "" && user.ID.String() != auth.SellerID) || (user.SiteID != "" && user.SiteID != "CBT") {
		return PublicationEvidence{}, errors.New("CBT seller tags: identity mismatch")
	}

	var mapping struct {
		UserID       json.Number `json:"user_id"`
		SiteID       string      `json:"site_id"`
		Marketplaces []struct {
			UserID        json.Number `json:"user_id"`
			SiteID        string      `json:"site_id"`
			LogisticType  string      `json:"logistic_type"`
			BusinessModel string      `json:"business_model"`
			PricingModel  string      `json:"pricing_model"`
			UserProduct   *bool       `json:"user_product"`
		} `json:"marketplaces"`
	}
	if err := s.readC0JSON(ctx, "/marketplace/users/"+sellerPath, auth.AccessToken, &mapping); err != nil {
		return PublicationEvidence{}, fmt.Errorf("CBT marketplace mapping: %w", err)
	}
	if mapping.UserID.String() != auth.SellerID || mapping.SiteID != "CBT" || mapping.Marketplaces == nil || len(mapping.Marketplaces) != len(auth.Markets) || len(mapping.Marketplaces) > 32 {
		return PublicationEvidence{}, errors.New("CBT marketplace identity or mapping changed: verify store authorization")
	}
	expected := make(map[string]string, len(auth.Markets))
	for _, m := range auth.Markets {
		expected[m.SiteID+"|"+m.LogisticType] = m.UserID.String()
	}
	seen := make(map[string]bool, len(mapping.Marketplaces))
	evidence := PublicationEvidence{
		EvidenceLevel:       "c0_read_only_not_publish_approved",
		SellAllowed:         publicationCapability(self.Status.Sell.Allow),
		UserProductSeller:   publicationTag(user.Tags, "user_product_seller"),
		WarehouseManagement: publicationTag(user.Tags, "warehouse_management"),
		CheckedAt:           time.Now().UTC(),
		Sites:               make([]PublicationSiteEvidence, 0, len(mapping.Marketplaces)),
	}
	for _, m := range mapping.Marketplaces {
		key := m.SiteID + "|" + m.LogisticType
		child, ok := expected[key]
		if !ok || seen[key] || child != m.UserID.String() {
			return PublicationEvidence{}, errors.New("CBT marketplace mapping differs from authorized store")
		}
		seen[key] = true
		evidence.Sites = append(evidence.Sites, PublicationSiteEvidence{
			SiteID: m.SiteID, LogisticType: m.LogisticType,
			BusinessModel: publicationBusinessModel(m.BusinessModel),
			PricingModel:  publicationPricingModel(m.PricingModel),
			UserProduct:   publicationCapability(m.UserProduct),
		})
	}
	evidence.Mode = ClassifyPublicationEvidence(evidence)
	return evidence, nil
}

func (s *Service) readC0JSON(ctx context.Context, path, token string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.apiURL+path, nil)
	if err != nil {
		return errors.New("invalid CBT read request")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	res, err := s.client.Do(req)
	if err != nil {
		return errors.New("CBT account read unavailable")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("CBT account read rejected (HTTP %d)", res.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, 128*1024+1))
	if err != nil || len(body) > 128*1024 {
		return errors.New("invalid CBT response size")
	}
	if err = json.Unmarshal(body, out); err != nil {
		return errors.New("invalid CBT response shape")
	}
	return nil
}

func publicationCapability(v *bool) string {
	if v == nil {
		return "unverified"
	}
	if *v {
		return "enabled"
	}
	return "disabled"
}
func publicationTag(tags []string, name string) string {
	if tags == nil {
		return "unverified"
	}
	for _, tag := range tags {
		if tag == name {
			return "present"
		}
	}
	return "absent"
}
func publicationBusinessModel(v string) string {
	if v == "" {
		return "unverified"
	}
	if v == "CBT CN Fulfillment Managed" {
		return v
	}
	return "other"
}
func publicationPricingModel(v string) string {
	switch v {
	case "price", "listing_price", "net_proceeds":
		return v
	case "":
		return "unverified"
	default:
		return "other"
	}
}

// These are routing candidates, not permission to POST.
// A store is classified as Fully Managed only when EVERY mapped child
// explicitly agrees. Mixed/unknown evidence stays undetermined. Legacy
// eligibility always needs separate write-permission verification.
func ClassifyPublicationEvidence(e PublicationEvidence) string {
	if e.SellAllowed != "enabled" || len(e.Sites) == 0 {
		return publicationUnverified
	}
	managedCount := 0
	hasUnknown := false
	hasDisabledUP := false
	hasEnabledUP := false
	for _, site := range e.Sites {
		if site.BusinessModel == "CBT CN Fulfillment Managed" {
			managedCount++
		}
		if site.BusinessModel == "other" {
			hasUnknown = true
		}
		if site.UserProduct == "disabled" {
			hasDisabledUP = true
		}
		if site.UserProduct == "enabled" {
			hasEnabledUP = true
		}
	}
	if managedCount > 0 {
		if managedCount == len(e.Sites) {
			return "fully_managed"
		}
		return publicationUnverified
	}
	if hasUnknown {
		return publicationUnverified
	}
	if e.UserProductSeller == "present" && !hasDisabledUP {
		return "standard_up"
	}
	if e.UserProductSeller == "absent" && !hasEnabledUP {
		return "legacy_candidate"
	}
	return publicationUnverified
}

