package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	apiBase    = "https://api.mercadolibre.com"
	unverified = "UNVERIFIED"
	maxBody    = 128 << 10
)

var (
	userIDFormat = regexp.MustCompile("^[1-9][0-9]{0,19}$")
	siteIDFormat = regexp.MustCompile("^[A-Z]{3}$")
)

type accountEvidence struct {
	SiteID            string                `json:"site_id"`
	UserProductSeller string                `json:"user_product_seller"`
	GlobalSeller      string                `json:"global_seller"`
	WarehouseAccess   string                `json:"warehouse_management"`
	SellAllowed       string                `json:"sell_allowed"`
	Marketplaces      []marketplaceEvidence `json:"marketplaces"`
}

type marketplaceEvidence struct {
	SellerAlias   string `json:"seller_alias"`
	SiteID        string `json:"site_id"`
	LogisticType  string `json:"logistic_type"`
	BusinessModel string `json:"business_model"`
	PricingModel  string `json:"pricing_model"`
	UserProduct   string `json:"user_product"`
}

func probe(client *http.Client, base, token string) (accountEvidence, error) {
	var self struct {
		ID     json.Number `json:"id"`
		SiteID string      `json:"site_id"`
		Status struct {
			Sell struct {
				Allow *bool `json:"allow"`
			} `json:"sell"`
		} `json:"status"`
	}
	if err := readJSON(client, base, "/users/me", "account identity", token, &self); err != nil {
		return accountEvidence{}, err
	}
	merchant := self.ID.String()
	if self.SiteID != "CBT" || !userIDFormat.MatchString(merchant) {
		return accountEvidence{}, errors.New("account identity: expected a CBT merchant")
	}

	var user struct {
		ID     json.Number `json:"id"`
		SiteID string      `json:"site_id"`
		Tags   []string    `json:"tags"`
	}
	if err := readJSON(client, base, "/users/"+merchant, "seller tags", token, &user); err != nil {
		return accountEvidence{}, err
	}
	if (user.ID != "" && user.ID != self.ID) || (user.SiteID != "" && user.SiteID != "CBT") {
		return accountEvidence{}, errors.New("seller tags: account identity mismatch")
	}

	var mapping struct {
		ID           json.Number `json:"user_id"`
		SiteID       string      `json:"site_id"`
		Marketplaces []struct {
			ID            json.Number `json:"user_id"`
			SiteID        string      `json:"site_id"`
			LogisticType  string      `json:"logistic_type"`
			BusinessModel string      `json:"business_model"`
			PricingModel  string      `json:"pricing_model"`
			UserProduct   *bool       `json:"user_product"`
		} `json:"marketplaces"`
	}
	if err := readJSON(client, base, "/marketplace/users/"+merchant, "marketplace mapping", token, &mapping); err != nil {
		return accountEvidence{}, err
	}
	if mapping.ID != self.ID || mapping.SiteID != "CBT" || mapping.Marketplaces == nil {
		return accountEvidence{}, errors.New("marketplace mapping: expected the same CBT merchant and a marketplace array")
	}

	result := accountEvidence{
		SiteID:            "CBT",
		UserProductSeller: tagStatus(user.Tags, "user_product_seller"),
		GlobalSeller:      tagStatus(user.Tags, "global_seller"),
		WarehouseAccess:   tagStatus(user.Tags, "warehouse_management"),
		SellAllowed:       capabilityStatus(self.Status.Sell.Allow),
		Marketplaces:      make([]marketplaceEvidence, 0, len(mapping.Marketplaces)),
	}
	aliases := make(map[json.Number]string, len(mapping.Marketplaces))
	for _, entry := range mapping.Marketplaces {
		if !userIDFormat.MatchString(entry.ID.String()) || entry.ID == self.ID || entry.SiteID == "CBT" || !siteIDFormat.MatchString(entry.SiteID) {
			return accountEvidence{}, errors.New("marketplace mapping: invalid child identity")
		}
		alias, exists := aliases[entry.ID]
		if !exists {
			alias = fmt.Sprintf("seller-%d", len(aliases)+1)
			aliases[entry.ID] = alias
		}
		result.Marketplaces = append(result.Marketplaces, marketplaceEvidence{
			SellerAlias:   alias,
			SiteID:        entry.SiteID,
			LogisticType:  knownValue(entry.LogisticType, "remote", "fulfillment", "cross_docking"),
			BusinessModel: knownValue(entry.BusinessModel, "CBT CN Fulfillment Managed"),
			PricingModel:  knownValue(entry.PricingModel, "net_proceeds", "listing_price", "price"),
			UserProduct:   capabilityStatus(entry.UserProduct),
		})
	}
	return result, nil
}

func capabilityStatus(enabled *bool) string {
	if enabled == nil {
		return unverified
	}
	if *enabled {
		return "enabled"
	}
	return "disabled"
}

func tagStatus(tags []string, tag string) string {
	switch {
	case tags == nil:
		return unverified
	case slices.Contains(tags, tag):
		return "present"
	default:
		return "absent"
	}
}

func knownValue(value string, known ...string) string {
	if slices.Contains(known, value) {
		return value
	}
	return unverified
}

func readJSON(client *http.Client, base, path, label, token string, out any) error {
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	if err != nil {
		return fmt.Errorf("%s: invalid request", label)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%s: request failed", label)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: HTTP %d", label, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return fmt.Errorf("%s: could not read response", label)
	}
	if len(body) > maxBody || json.Unmarshal(body, out) != nil {
		return fmt.Errorf("%s: invalid or oversized JSON response", label)
	}
	return nil
}

func probeClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func main() {
	token := strings.TrimSpace(os.Getenv("MELI_CBT_PROBE_TOKEN"))
	if token == "" {
		fmt.Fprintln(os.Stderr, "MELI_CBT_PROBE_TOKEN is required from an approved CBT test session")
		os.Exit(1)
	}
	evidence, err := probe(probeClient(), apiBase, token)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	output := json.NewEncoder(os.Stdout)
	output.SetIndent("", "  ")
	if err := output.Encode(evidence); err != nil {
		fmt.Fprintln(os.Stderr, "could not write sanitized evidence")
		os.Exit(1)
	}
}
