package publication

import (
	"fmt"
	"regexp"
	"strings"
)

type ExpectedVariant struct {
	Key, SKU string
}

type ReadbackScope struct {
	ParentSellerID, SiteSellerID string
	Site, Logistic               string
}

// VerifiedListing is normalized evidence from authenticated official GETs.
// Only candidates bound to this creation attempt may enter this matcher;
// a SKU search alone cannot establish that provenance. The adapter must
// establish the UP/item mapping and family membership before constructing it.
// Family and seller identifiers are strings to preserve their exact precision.
type VerifiedListing struct {
	Scope                          ReadbackScope
	SKU                            string
	SitelessUPID, SitelessFamilyID string
	SiteUPID, SiteFamilyID, ItemID string
	SiteFamilyName                 string
}

type MatchedVariant struct {
	Key     string
	Listing VerifiedListing
	// Identity can be confirmed even if a custom title needs verification.
	TitleVerified bool
}

type ReadbackResult struct {
	Matched []MatchedVariant
	Pending []string
}

var (
	numericID  = regexp.MustCompile(`^[1-9][0-9]{0,29}$`)
	globalUPID = regexp.MustCompile(`^U[1-9][0-9]{0,29}$`)
)

// MatchReadback never guesses by response order, never creates anything and
// never concludes that missing evidence means the remote create failed.
// Conflicting or foreign evidence fails closed; incomplete evidence leaves
// unmatched identities pending. Callers retain all returned candidate IDs.
func MatchReadback(p Prepared, scope ReadbackScope, listings []VerifiedListing) (ReadbackResult, error) {
	if !numericID.MatchString(scope.ParentSellerID) || !numericID.MatchString(scope.SiteSellerID) ||
		scope.Site != p.Target.Site || scope.Logistic != p.Target.Logistic ||
		!sitePattern.MatchString(scope.Site) || scope.Logistic != "remote" || len(p.Variants) == 0 ||
		(p.Copy.Mode != Localized && p.Copy.Mode != PlatformAuto) {
		return ReadbackResult{}, fmt.Errorf("invalid readback scope")
	}
	bySKU := make(map[string]VerifiedListing, len(listings))
	expected := make(map[string]bool, len(p.Variants))
	keys := make(map[string]bool, len(p.Variants))
	for _, v := range p.Variants {
		if v.SKU == "" || v.Key == "" || expected[v.SKU] || keys[v.Key] {
			return ReadbackResult{}, fmt.Errorf("ambiguous expected variant")
		}
		expected[v.SKU], keys[v.Key] = true, true
	}
	var globalFamily, siteFamily string
	seenUP, seenSiteUP, seenItem := map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, row := range listings {
		if row.Scope != scope || !expected[row.SKU] || !globalUPID.MatchString(row.SitelessUPID) ||
			!numericID.MatchString(row.SitelessFamilyID) || !numericID.MatchString(row.SiteFamilyID) ||
			!prefixedNumeric(row.SiteUPID, scope.Site+"U") || !prefixedNumeric(row.ItemID, scope.Site) {
			return ReadbackResult{}, fmt.Errorf("unverified or out-of-scope listing evidence")
		}
		if _, exists := bySKU[row.SKU]; exists || seenUP[row.SitelessUPID] || seenSiteUP[row.SiteUPID] || seenItem[row.ItemID] {
			return ReadbackResult{}, fmt.Errorf("ambiguous variant mapping")
		}
		if globalFamily != "" && (globalFamily != row.SitelessFamilyID || siteFamily != row.SiteFamilyID) {
			return ReadbackResult{}, fmt.Errorf("family mismatch")
		}
		globalFamily, siteFamily = row.SitelessFamilyID, row.SiteFamilyID
		bySKU[row.SKU] = row
		seenUP[row.SitelessUPID], seenSiteUP[row.SiteUPID], seenItem[row.ItemID] = true, true, true
	}
	var result ReadbackResult
	for _, v := range p.Variants {
		row, ok := bySKU[v.SKU]
		if !ok {
			result.Pending = append(result.Pending, v.Key)
			continue
		}
		// Compare the site family name, not the generated item title (which
		// may contain appended variation attributes). No text fallback.
		titleOK := p.Copy.Mode == PlatformAuto && strings.TrimSpace(row.SiteFamilyName) != "" ||
			p.Copy.Mode == Localized && row.SiteFamilyName == p.Copy.SiteFamilyName
		result.Matched = append(result.Matched, MatchedVariant{Key: v.Key, Listing: row, TitleVerified: titleOK})
	}
	return result, nil
}

func prefixedNumeric(id, prefix string) bool {
	return strings.HasPrefix(id, prefix) && numericID.MatchString(strings.TrimPrefix(id, prefix))
}
