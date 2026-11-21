package collections

import "testing"

func TestListingFieldsValidation(t *testing.T) {
	good := ListingFields{
		CategorySite:     "MLB",
		CategoryID:       "MLB123",
		PurchasePrice:    "25.50",
		PurchaseCurrency: "CNY",
		WeightG:          "850",
		LengthCM:         "20.5",
		WidthCM:          "15",
		HeightCM:         "5",
		SitePrices:       []SitePrice{{Site: "MLB", Price: "89.90", Currency: "BRL"}},
	}
	if err := good.Validate(); err != nil {
		t.Fatalf("valid listing fields rejected: %v", err)
	}
	tests := []struct {
		name   string
		change func(*ListingFields)
	}{
		{"negative weight", func(x *ListingFields) { x.WeightG = "-20" }},
		{"decimal weight", func(x *ListingFields) { x.WeightG = "1.5" }},
		{"invalid dimensions", func(x *ListingFields) { x.WidthCM = "0" }},
		{"invalid price", func(x *ListingFields) { x.PurchasePrice = "12,5" }},
		{"wrong site currency", func(x *ListingFields) { x.SitePrices[0].Currency = "USD" }},
		{"duplicate site", func(x *ListingFields) { x.SitePrices = append(x.SitePrices, x.SitePrices[0]) }},
		{"invalid image link", func(x *ListingFields) {
			images := []Image{{URL: "http://invalid.test/a.jpg", Type: "main"}}
			x.Images = &images
		}},
		{"invalid variant value", func(x *ListingFields) {
			variants := []ListingVariant{{Attributes: map[string]string{"颜色": ""}}}
			x.Variants = &variants
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			copy := good
			copy.SitePrices = append([]SitePrice(nil), good.SitePrices...)
			tc.change(&copy)
			if err := copy.Validate(); err == nil {
				t.Fatal("invalid fields accepted")
			}
		})
	}
}
