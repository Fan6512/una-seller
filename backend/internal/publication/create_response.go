package publication

import (
	"encoding/json"
	"fmt"
)

// CreateCandidate is an identity hint from the response to a persisted create
// attempt. It is not a verified mapping to a local variant or a success result.
// Keep hints even if a target site is missing or another row is incomplete.
type CreateCandidate struct {
	SitelessUPID, SitelessFamilyID string
	SiteUPID, ItemID               string
}

type CreateHints struct {
	Candidates []CreateCandidate
	Incomplete bool
}

type createRow struct {
	SellerID         json.Number       `json:"seller_id"`
	Site             string            `json:"site_id"`
	SitelessUPID     string            `json:"siteless_user_product_id"`
	SitelessFamilyID json.Number       `json:"siteless_family_id"`
	SiteItems        []json.RawMessage `json:"site_items"`
}

type createSite struct {
	SellerID json.Number `json:"seller_id"`
	Site     string      `json:"site_id"`
	Logistic string      `json:"logistic_type"`
	UPID     string      `json:"user_product_id"`
	ItemID   string      `json:"item_id"`
}

// ExtractCreateHints accepts the documented single-object and Family-array
// responses. No positional association, failure classification, or retry
// decision is made here. The caller persists hints under the original attempt
// and uses authenticated GETs to establish VerifiedListing evidence.
func ExtractCreateHints(p Prepared, scope ReadbackScope, body []byte) (CreateHints, error) {
	if scope.Site != p.Target.Site || scope.Logistic != p.Target.Logistic ||
		!numericID.MatchString(scope.ParentSellerID) || !numericID.MatchString(scope.SiteSellerID) ||
		!sitePattern.MatchString(scope.Site) || scope.Logistic != "remote" || len(p.Variants) == 0 || len(body) > 2<<20 {
		return CreateHints{}, fmt.Errorf("invalid creation response scope or size")
	}
	var rows []json.RawMessage
	switch p.Path {
	case SinglePath:
		rows = []json.RawMessage{body}
	case FamilyPath:
		if err := json.Unmarshal(body, &rows); err != nil {
			return CreateHints{}, fmt.Errorf("unrecognized family creation response")
		}
	default:
		return CreateHints{}, fmt.Errorf("unsupported creation route")
	}
	result := CreateHints{Incomplete: len(rows) != len(p.Variants)}
	seen := make(map[string]bool)
	var family string
	for _, raw := range rows {
		var row createRow
		if err := json.Unmarshal(raw, &row); err != nil || row.Site != "CBT" || row.SellerID.String() != scope.ParentSellerID || !globalUPID.MatchString(row.SitelessUPID) {
			result.Incomplete = true
			continue
		}
		if seen[row.SitelessUPID] {
			result.Incomplete = true
			continue
		}
		seen[row.SitelessUPID] = true
		c := CreateCandidate{SitelessUPID: row.SitelessUPID}
		if id := row.SitelessFamilyID.String(); numericID.MatchString(id) {
			c.SitelessFamilyID = id
			if family != "" && family != id {
				result.Incomplete = true
			}
			family = id
		} else {
			result.Incomplete = true
		}
		targets := 0
		for _, rawSite := range row.SiteItems {
			var s createSite
			if err := json.Unmarshal(rawSite, &s); err != nil || s.Site != scope.Site || s.Logistic != scope.Logistic || s.SellerID.String() != scope.SiteSellerID ||
				!prefixedNumeric(s.UPID, scope.Site+"U") || !prefixedNumeric(s.ItemID, scope.Site) {
				result.Incomplete = true
				continue
			}
			targets++
			c.SiteUPID, c.ItemID = s.UPID, s.ItemID
		}
		if targets != 1 {
			c.SiteUPID, c.ItemID = "", ""
			result.Incomplete = true
		}
		result.Candidates = append(result.Candidates, c)
	}
	return result, nil
}
