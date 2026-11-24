package collections

import (
	"strings"
	"testing"
)

func TestListingFieldsValidation(t *testing.T) {
	good := ListingFields{
		Titles:           LocalizedTitles{English: "Winter Plant Cleaning Spray", Spanish: "Limpiador de Plantas", Portuguese: "Spray para Plantas"},
		CategoryID:       "CBT123",
		PurchasePrice:    "25.50",
		PurchaseCurrency: "CNY",
		WeightG:          "850",
		LengthCM:         "20.5",
		WidthCM:          "15",
		HeightCM:         "5",
	}
	if err := good.Validate(); err != nil {
		t.Fatalf("valid listing fields rejected: %v", err)
	}
	tests := []struct {
		name   string
		change func(*ListingFields)
	}{
		{"invalid site-specific category", func(x *ListingFields) { x.CategoryID = "MLB123" }},
		{"overlong localized title", func(x *ListingFields) { x.Titles.Spanish = strings.Repeat("a", 501) }},
		{"negative weight", func(x *ListingFields) { x.WeightG = "-20" }},
		{"decimal weight", func(x *ListingFields) { x.WeightG = "1.5" }},
		{"invalid dimensions", func(x *ListingFields) { x.WidthCM = "0" }},
		{"invalid price", func(x *ListingFields) { x.PurchasePrice = "12,5" }},
		{"invalid purchase currency", func(x *ListingFields) { x.PurchaseCurrency = "NOT-CURRENCY" }},
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
			tc.change(&copy)
			if err := copy.Validate(); err == nil {
				t.Fatal("invalid fields accepted")
			}
		})
	}
}

func TestDescriptionTranslationManualEditAndReadoption(t *testing.T) {
	english := "Glass cup for home use."
	digest := EnglishDescriptionDigest(english)
	initial := ListingFields{
		Descriptions:           &LocalizedDescriptions{English: english, Spanish: "Taza de vidrio"},
		DescriptionTranslation: &DescriptionTranslation{EnglishDigest: digest, SpanishSourceDigest: digest},
	}
	manual := initial
	manual.Descriptions = &LocalizedDescriptions{English: english, Spanish: "Taza de vidrio, uso doméstico"}
	manual.DescriptionTranslation = &DescriptionTranslation{SpanishSourceDigest: ""}
	afterManual, err := manual.prepareForSave(initial)
	if err != nil || !afterManual.DescriptionTranslation.SpanishManualEdited ||
		afterManual.DescriptionTranslation.SpanishSourceDigest != digest {
		t.Fatalf("manual edit must preserve original source and mark manual: %+v %v", afterManual, err)
	}
	readopt := *afterManual
	readopt.Descriptions = &LocalizedDescriptions{English: english, Spanish: "Nuevo texto de la IA"}
	readopt.DescriptionTranslation = &DescriptionTranslation{SpanishSourceDigest: digest}
	afterAI, err := readopt.prepareForSave(*afterManual)
	if err != nil || afterAI.DescriptionTranslation.SpanishManualEdited ||
		afterAI.DescriptionTranslation.SpanishSourceDigest != digest {
		t.Fatalf("re-adopting the same master must clear manual marker: %+v %v", afterAI, err)
	}
}

func TestTargetListingAmounts(t *testing.T) {
	proceeds := "8.1234"
	variants := []ListingVariant{{Attributes: map[string]string{"Color": "Black"}, Price: "79.90", Currency: "BRL", PurchaseCostCNY: new("25.50"), TargetNetProceedsUSD: new("8.50"), CategoryAttributes: &[]ListingCategoryAttribute{{CategoryID: "CBT123", ID: "COLOR", Value: "Black", ValueID: "1"}}}}
	fields := ListingFields{TargetNetProceedsUSD: &proceeds, Variants: &variants}
	if err := fields.Validate(); err != nil {
		t.Fatal(err)
	}
	proceeds = "-1"
	if err := fields.Validate(); err == nil {
		t.Fatal("invalid proceeds accepted")
	}
	proceeds = ""
	variants[0].PurchaseCostCNY = new("1.12345")
	if err := fields.Validate(); err == nil {
		t.Fatal("invalid cost accepted")
	}
}

func TestPreserveVariantFields(t *testing.T) {
	black := map[string]string{"Color": "Black"}
	white := map[string]string{"Color": "White"}
	blue := map[string]string{"Color": "Blue"}
	previous := []ListingVariant{
		{SellerSKU: "black", Attributes: black, PurchaseCostCNY: new("25.50"), TargetNetProceedsUSD: new("8.50")},
		{Attributes: white, PurchaseCostCNY: new("20.00"), CategoryAttributes: &[]ListingCategoryAttribute{{CategoryID: "CBT1055", ID: "COLOR", Value: "White", ValueID: "52055"}}},
	}
	current := []ListingVariant{
		{Attributes: white, PurchaseCostCNY: new(""), CategoryAttributes: &[]ListingCategoryAttribute{}},
		{SellerSKU: "black", Attributes: blue},
	}
	saved, err := preserveVariantFields(current, previous)
	if err != nil {
		t.Fatal(err)
	}
	if *saved[0].PurchaseCostCNY != "" || len(*saved[0].CategoryAttributes) != 0 ||
		*saved[1].PurchaseCostCNY != "25.50" || *saved[1].TargetNetProceedsUSD != "8.50" || current[1].PurchaseCostCNY != nil {
		t.Fatalf("old metadata was lost, copied to the wrong row, or restored after an explicit clear: %+v", saved)
	}
	if _, err := preserveVariantFields([]ListingVariant{{Attributes: blue}}, previous); err != ErrDraftConflict {
		t.Fatalf("ambiguous legacy replacement should conflict, got %v", err)
	}
}
