package meli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Marketplace is a site-specific seller CHILD of one authenticated CBT parent.
// It never owns its own OAuth token; the parent seller_id owns the token.
type Marketplace struct {
	UserID       json.Number `json:"user_id"`
	SiteID       string      `json:"site_id"`
	LogisticType string      `json:"logistic_type"`
}

// getMarketplaces reads the account's real site mapping, without assuming every
// country advertised by Global Selling is enabled on the current seller.
func (s *Service) getMarketplaces(ctx context.Context, parentID, accessToken string) ([]Marketplace, error) {
	path := s.apiURL + "/marketplace/users/" + url.PathEscape(parentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, errors.New("invalid marketplace identity request")
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, errors.New("marketplace mapping unavailable")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("marketplace mapping not available for this account")
	}
	var mapping struct {
		UserID       json.Number   `json:"user_id"`
		SiteID       string        `json:"site_id"`
		Marketplaces []Marketplace `json:"marketplaces"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<17)).Decode(&mapping); err != nil {
		return nil, errors.New("invalid marketplace mapping response")
	}
	if mapping.UserID.String() != parentID || mapping.SiteID != "CBT" ||
		mapping.Marketplaces == nil || len(mapping.Marketplaces) > 32 {
		return nil, errors.New("marketplace mapping does not match CBT parent")
	}
	seen := make(map[[2]string]bool, len(mapping.Marketplaces))
	for _, m := range mapping.Marketplaces {
		site, mode := m.SiteID, m.LogisticType
		key := [2]string{site, mode}
		child := m.UserID.String()
		if len(site) < 3 || len(site) > 8 || strings.ToUpper(site) != site ||
			site != strings.TrimSpace(site) || child == "" || child == "0" || len(child) > 64 ||
			mode == "" || mode != strings.TrimSpace(mode) || len(mode) > 40 || seen[key] {
			return nil, errors.New("invalid marketplace child identity")
		}
		seen[key] = true
	}
	return mapping.Marketplaces, nil
}
