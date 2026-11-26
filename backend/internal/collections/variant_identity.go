package collections

import (
	"crypto/rand"
	"encoding/hex"
	"maps"
	"regexp"
	"slices"
)

var variantIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// This is a local integer storage/UI bound, not a platform inventory limit.
func validAvailableQuantity(quantity *int) bool {
	return quantity == nil || (*quantity >= 0 && *quantity <= 2147483647)
}

// Identity version 1 is an explicit editor capability. Older clients can
// continue saving unrelated fields, but cannot ambiguously replace saved IDs.
// All persisted IDs originate on the server; a blank ID means a new row.
func prepareVariantIdentity(next *ListingFields, previous ListingFields) error {
	modern := next.VariantIdentityVersion == 1
	hadIdentity := previous.VariantIdentityVersion == 1
	if !modern && hadIdentity {
		next.VariantIdentityVersion = 1
		next.AvailableQuantity = previous.AvailableQuantity
	}
	if next.Variants == nil && (modern || hadIdentity) {
		next.Variants = previous.Variants
	}
	if next.Variants == nil {
		return nil
	}
	rows := slices.Clone(*next.Variants)
	var old []ListingVariant
	if previous.Variants != nil {
		old = *previous.Variants
	}
	if !modern && !hadIdentity {
		for _, row := range rows {
			if row.VariantID != "" || row.AvailableQuantity != nil {
				return ErrVariantUpgradeRequired
			}
		}
		return nil
	}

	known := make(map[string]ListingVariant)
	for _, row := range old {
		if hadIdentity {
			if !variantIDPattern.MatchString(row.VariantID) {
				return ErrVariantUpgradeRequired
			}
			if _, exists := known[row.VariantID]; exists {
				return ErrVariantUpgradeRequired
			}
			known[row.VariantID] = row
		}
	}
	if !modern {
		// No deletions or additions through a client that cannot express identity.
		if len(rows) != len(old) {
			return ErrVariantUpgradeRequired
		}
		used := make(map[string]bool)
		for i, row := range rows {
			var match ListingVariant
			n := 0
			for _, prior := range old {
				if sameLegacyVariant(row, prior) {
					match, n = prior, n+1
				}
			}
			if n != 1 || used[match.VariantID] || (row.VariantID != "" && row.VariantID != match.VariantID) {
				return ErrVariantUpgradeRequired
			}
			used[match.VariantID] = true
			rows[i].VariantID = match.VariantID
			rows[i].AvailableQuantity = match.AvailableQuantity
			preserveIdentityMetadata(&rows[i], match)
		}
		next.Variants = &rows
		return nil
	}

	used := make(map[string]bool)
	for i := range rows {
		row := &rows[i]
		if row.VariantID != "" {
			prior, exists := known[row.VariantID]
			if !exists || used[row.VariantID] {
				return ErrVariantUpgradeRequired
			}
			used[row.VariantID] = true
			preserveIdentityMetadata(row, prior)
		}
	}
	for i := range rows {
		row := &rows[i]
		if row.VariantID != "" {
			continue
		}
		// A formerly identified row must not silently acquire a new ID.
		for _, prior := range old {
			if hadIdentity && !used[prior.VariantID] &&
				(sameLegacyVariant(*row, prior) || (row.SellerSKU != "" && row.SellerSKU == prior.SellerSKU && maps.Equal(row.Attributes, prior.Attributes))) {
				return ErrVariantUpgradeRequired
			}
		}
		var raw [16]byte
		if _, err := rand.Read(raw[:]); err != nil {
			return err
		}
		row.VariantID = hex.EncodeToString(raw[:])
		if _, exists := known[row.VariantID]; exists || used[row.VariantID] {
			return ErrVariantUpgradeRequired
		}
		used[row.VariantID] = true
	}
	next.Variants = &rows
	return nil
}

func preserveIdentityMetadata(row *ListingVariant, prior ListingVariant) {
	if row.PurchaseCostCNY == nil {
		row.PurchaseCostCNY = prior.PurchaseCostCNY
	}
	if row.TargetNetProceedsUSD == nil {
		row.TargetNetProceedsUSD = prior.TargetNetProceedsUSD
	}
	if row.CategoryAttributes == nil {
		row.CategoryAttributes = prior.CategoryAttributes
	}
}
