package collections

import "testing"

func TestStableVariantIdentity(t *testing.T) {
	q, zero := 5, 0
	cost := "12.50"
	rows := []ListingVariant{
		{SellerSKU: "RED", AvailableQuantity: &q, PurchaseCostCNY: &cost},
		{SellerSKU: "BLUE", AvailableQuantity: &zero},
	}
	next := ListingFields{VariantIdentityVersion: 1, Variants: &rows}
	if err := prepareVariantIdentity(&next, ListingFields{}); err != nil {
		t.Fatal(err)
	}
	saved := *next.Variants
	if saved[0].VariantID == "" || saved[0].VariantID == saved[1].VariantID || rows[0].VariantID != "" {
		t.Fatal("server identity must be distinct and must not mutate input")
	}
	reordered := []ListingVariant{saved[1], saved[0]}
	reordered[1].SellerSKU = "RED-NEW"
	reordered[1].PurchaseCostCNY = nil
	edit := ListingFields{VariantIdentityVersion: 1, Variants: &reordered}
	if err := prepareVariantIdentity(&edit, next); err != nil {
		t.Fatal(err)
	}
	got := *edit.Variants
	if got[0].VariantID != saved[1].VariantID || *got[0].AvailableQuantity != 0 ||
		got[1].VariantID != saved[0].VariantID || *got[1].PurchaseCostCNY != cost {
		t.Fatal("reordering or SKU editing changed identity or metadata")
	}
	for _, bad := range [][]ListingVariant{
		{saved[0], saved[0]},
		{{VariantID: "00000000000000000000000000000000"}},
		{{SellerSKU: saved[0].SellerSKU}},
	} {
		in := ListingFields{VariantIdentityVersion: 1, Variants: &bad}
		if err := prepareVariantIdentity(&in, next); err != ErrVariantUpgradeRequired {
			t.Fatalf("ambiguous identity must be rejected: %v", err)
		}
	}
	legacyRows := []ListingVariant{{SellerSKU: "BLUE"}, {SellerSKU: "RED"}}
	legacy := ListingFields{Variants: &legacyRows}
	if err := prepareVariantIdentity(&legacy, next); err != nil {
		t.Fatal(err)
	}
	if legacy.VariantIdentityVersion != 1 || (*legacy.Variants)[1].VariantID != saved[0].VariantID ||
		*(*legacy.Variants)[1].AvailableQuantity != q {
		t.Fatal("legacy projection lost identity or stock")
	}
}
