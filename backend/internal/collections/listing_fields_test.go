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
		{SellerSKU: "white", Attributes: white, PurchaseCostCNY: new("20.00"), CategoryAttributes: &[]ListingCategoryAttribute{{CategoryID: "CBT1055", ID: "COLOR", Value: "White", ValueID: "52055"}}},
	}
	current := []ListingVariant{
		{SellerSKU: "white", Attributes: white, PurchaseCostCNY: new(""), CategoryAttributes: &[]ListingCategoryAttribute{}},
		{SellerSKU: "black", Attributes: black, ImageURLs: []string{"https://example.com/new.jpg"}},
	}
	saved, err := preserveVariantFields(current, previous)
	if err != nil {
		t.Fatal(err)
	}
	if *saved[0].PurchaseCostCNY != "" || len(*saved[0].CategoryAttributes) != 0 ||
		*saved[1].PurchaseCostCNY != "25.50" || *saved[1].TargetNetProceedsUSD != "8.50" || current[1].PurchaseCostCNY != nil {
		t.Fatalf("old metadata lost or copied incorrectly: %+v", saved)
	}
	for _, rows := range [][]ListingVariant{
		{{SellerSKU: "new", Attributes: blue}, {SellerSKU: "white", Attributes: white}},
		{{SellerSKU: "black", Attributes: blue}, {SellerSKU: "white", Attributes: white}},
		{{SellerSKU: "white", Attributes: white}},
		{{SellerSKU: "black", Attributes: white}, {SellerSKU: "white", Attributes: black}},
	} {
		if _, err := preserveVariantFields(rows, previous); err != ErrVariantUpgradeRequired {
			t.Fatalf("ambiguous legacy identity edit should request upgrade, got %v", err)
		}
	}
	previous[0].SellerSKU, previous[1].SellerSKU = "", ""
	legacy := []ListingVariant{{Attributes: black}, {Attributes: white}}
	if _, err := preserveVariantFields(legacy, previous); err != nil {
		t.Fatalf("unchanged unnamed variants should retain metadata: %v", err)
	}
	legacy[0], legacy[1] = legacy[1], legacy[0]
	saved, err = preserveVariantFields(legacy, previous)
	if err != nil || *saved[0].PurchaseCostCNY != "20.00" || *saved[1].PurchaseCostCNY != "25.50" {
		t.Fatalf("unnamed unchanged projections must follow the reordered rows: %+v %v", saved, err)
	}
	legacy[0].Price = "9.00"
	if _, err := preserveVariantFields(legacy, previous); err != ErrVariantUpgradeRequired {
		t.Fatalf("changed unnamed projection must require upgrade: %v", err)
	}
}

func TestPreserveUnchangedDuplicateSKUGroup(t *testing.T) {
	black := map[string]string{"Color": "Black"}
	white := map[string]string{"Color": "White"}
	previous := []ListingVariant{
		{SellerSKU: "DUP", Attributes: black, Price: "10.00", Currency: "BRL", ImageURLs: []string{"https://example.com/black.jpg"}, PurchaseCostCNY: new("25.50"), TargetNetProceedsUSD: new("8.50")},
		{SellerSKU: "DUP", Attributes: white, Price: "20.00", Currency: "BRL", ImageURLs: []string{"https://example.com/white.jpg"}, PurchaseCostCNY: new("20.00"), CategoryAttributes: &[]ListingCategoryAttribute{{CategoryID: "CBT1055", ID: "COLOR", Value: "White", ValueID: "52055"}}},
	}
	legacy := []ListingVariant{previous[0], previous[1]}
	for i := range legacy {
		legacy[i].PurchaseCostCNY, legacy[i].TargetNetProceedsUSD, legacy[i].CategoryAttributes = nil, nil, nil
	}
	for _, rows := range [][]ListingVariant{legacy, {legacy[1], legacy[0]}} {
		saved, err := preserveVariantFields(rows, previous)
		if err != nil {
			t.Fatalf("unchanged duplicate SKU should save: %v", err)
		}
		for i, row := range saved {
			if row.Attributes["Color"] == "Black" {
				if *row.PurchaseCostCNY != "25.50" || *row.TargetNetProceedsUSD != "8.50" {
					t.Fatal("black metadata lost")
				}
			} else if *row.PurchaseCostCNY != "20.00" || (*row.CategoryAttributes)[0].ValueID != "52055" {
				t.Fatal("white metadata lost")
			}
			if rows[i].PurchaseCostCNY != nil {
				t.Fatal("input mutated")
			}
		}
	}
	swapped := []ListingVariant{legacy[0], legacy[1]}
	swapped[0].Attributes, swapped[1].Attributes = white, black
	edited := []ListingVariant{legacy[0], legacy[1]}
	edited[0].ImageURLs = []string{"https://example.com/changed.jpg"}
	for _, rows := range [][]ListingVariant{swapped, edited, {legacy[0]}, {legacy[0], legacy[0]}} {
		if _, err := preserveVariantFields(rows, previous); err != ErrVariantUpgradeRequired {
			t.Fatalf("changed or ambiguous duplicate SKU must be rejected: %v", err)
		}
	}
	identical := []ListingVariant{previous[0], previous[0]}
	identical[1].PurchaseCostCNY = new("99.00")
	if _, err := preserveVariantFields([]ListingVariant{legacy[0], legacy[0]}, identical); err != ErrVariantUpgradeRequired {
		t.Fatalf("indistinguishable rows cannot inherit different costs: %v", err)
	}
}
