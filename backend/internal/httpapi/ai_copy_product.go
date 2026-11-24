package httpapi

import (
	"context"
	"strings"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"golang.org/x/text/unicode/norm"
)

type savedAICopyProduct struct {
	title       string
	legacyTitle string
	fields      collections.ListingFields
	revision    int64
	bullets     []string
	variants    []collections.Variant
}

func (a collectionAPI) readSavedAICopyProduct(ctx context.Context, owner string, input aiCopyInput) (savedAICopyProduct, error) {
	var result savedAICopyProduct
	if input.OwnProductID != "" {
		p, err := a.store.GetOwnProduct(ctx, owner, input.OwnProductID)
		return savedAICopyProduct{title: p.SourceTitle, legacyTitle: p.Title, fields: p.ListingFields,
			revision: p.Revision, bullets: p.BulletPoints, variants: p.Variants}, err
	}
	draft, err := a.store.GetWorkingDraft(ctx, owner, input.SourceProductID)
	if err != nil {
		return result, err
	}
	if draft != nil {
		result.fields, result.revision, result.legacyTitle = draft.ListingFields, draft.Revision, draft.Title
	}
	if input.Operation != "translate" {
		p, err := a.store.Get(ctx, owner, input.SourceProductID)
		if err != nil {
			return result, err
		}
		result.title, result.bullets, result.variants = p.Title, p.BulletPoints, p.Variants
	}
	return result, nil
}

type aiCopyVariant struct {
	SKU        string            `json:"seller_sku,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type aiCopyCBTAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

// Only explicitly selected product facts cross the provider boundary.
type aiCopyFacts struct {
	Title         string                         `json:"source_title,omitempty"`
	EnglishTitle  string                         `json:"english_title,omitempty"`
	LegacyTitle   string                         `json:"saved_title,omitempty"`
	Category      string                         `json:"cbt_category_path,omitempty"`
	CategoryID    string                         `json:"cbt_category_id,omitempty"`
	Brand         string                         `json:"brand,omitempty"`
	Model         string                         `json:"model,omitempty"`
	SKU           string                         `json:"seller_sku,omitempty"`
	WeightG       string                         `json:"weight_g,omitempty"`
	LengthCM      string                         `json:"length_cm,omitempty"`
	WidthCM       string                         `json:"width_cm,omitempty"`
	HeightCM      string                         `json:"height_cm,omitempty"`
	BulletPoints  []string                       `json:"bullet_points,omitempty"`
	Attributes    []collections.ListingAttribute `json:"attributes,omitempty"`
	CBTAttributes []aiCopyCBTAttribute           `json:"confirmed_cbt_attributes,omitempty"`
	Variants      []aiCopyVariant                `json:"variants,omitempty"`
}

func aiCopyFact(value string, limit int) string {
	value = strings.TrimSpace(norm.NFC.String(value))
	if !validAICopyText(value, limit, false) {
		return ""
	}
	return value
}

func (p savedAICopyProduct) facts() aiCopyFacts {
	f := p.fields
	result := aiCopyFacts{
		Title: aiCopyFact(p.title, 500), EnglishTitle: aiCopyFact(f.Titles.English, 500),
		Brand: aiCopyFact(f.Brand, 120), Model: aiCopyFact(f.Model, 120), SKU: aiCopyFact(f.SellerSKU, 120),
		WeightG: aiCopyFact(f.WeightG, 20), LengthCM: aiCopyFact(f.LengthCM, 20),
		WidthCM: aiCopyFact(f.WidthCM, 20), HeightCM: aiCopyFact(f.HeightCM, 20),
	}
	if f.Titles.Preferred() == "" && p.legacyTitle != p.title {
		result.LegacyTitle = aiCopyFact(p.legacyTitle, 500)
	}
	if f.CategoryID != "" {
		result.Category = aiCopyFact(f.CategoryPath, 400)
		if result.Category == "" {
			result.CategoryID = aiCopyFact(f.CategoryID, 60)
		}
		if f.CategoryAttributes != nil {
			for _, attr := range *f.CategoryAttributes {
				name, value := aiCopyFact(attr.ID, 80), aiCopyFact(attr.Value, 500)
				if attr.CategoryID == f.CategoryID && name != "" && value != "" {
					result.CBTAttributes = append(result.CBTAttributes, aiCopyCBTAttribute{
						Name: name, Value: value, Unit: aiCopyFact(attr.UnitID, 80),
					})
				}
			}
		}
	}
	bullets := f.BulletPoints
	if bullets == nil {
		bullets = p.bullets
	}
	for _, bullet := range bullets {
		if text := aiCopyFact(bullet, 500); text != "" {
			result.BulletPoints = append(result.BulletPoints, text)
		}
	}
	for _, attr := range f.Attributes {
		name, value := aiCopyFact(attr.Name, 120), aiCopyFact(attr.Value, 500)
		if name != "" && value != "" {
			result.Attributes = append(result.Attributes, collections.ListingAttribute{Name: name, Value: value})
		}
	}
	variants := f.Variants
	if variants == nil {
		captured := make([]collections.ListingVariant, 0, len(p.variants))
		for _, variant := range p.variants {
			captured = append(captured, collections.ListingVariant{Attributes: variant.Attributes})
		}
		variants = &captured
	}
	for _, variant := range *variants {
		item := aiCopyVariant{SKU: aiCopyFact(variant.SellerSKU, 120), Attributes: make(map[string]string)}
		for name, value := range variant.Attributes {
			name, value = aiCopyFact(name, 80), aiCopyFact(value, 200)
			if name != "" && value != "" {
				item.Attributes[name] = value
			}
		}
		if item.SKU != "" || len(item.Attributes) > 0 {
			result.Variants = append(result.Variants, item)
		}
	}
	return result
}
