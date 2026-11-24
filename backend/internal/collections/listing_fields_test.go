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
