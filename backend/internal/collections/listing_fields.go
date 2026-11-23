package collections

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ListingFields are editable preparation data, not official platform attributes.
// Workbench drafts are independent of listing markets and site prices.
type LocalizedTitles struct {
	English    string `json:"en"`
	Spanish    string `json:"es"`
	Portuguese string `json:"pt"`
}

func (titles LocalizedTitles) Preferred() string {
	for _, title := range []string{titles.English, titles.Spanish, titles.Portuguese} {
		if strings.TrimSpace(title) != "" {
			return strings.TrimSpace(title)
		}
	}
	return ""
}

type ListingFields struct {
	Titles             LocalizedTitles             `json:"titles"`
	CategoryID         string                      `json:"category_id"`
	CategoryPath       string                      `json:"category_path"`
	Brand              string                      `json:"brand"`
	Model              string                      `json:"model"`
	SellerSKU          string                      `json:"seller_sku"`
	GTIN               string                      `json:"gtin"`
	PurchasePrice      string                      `json:"purchase_price"`
	PurchaseCurrency   string                      `json:"purchase_currency"`
	WeightG            string                      `json:"weight_g"`
	LengthCM           string                      `json:"length_cm"`
	WidthCM            string                      `json:"width_cm"`
	HeightCM           string                      `json:"height_cm"`
	BulletPoints       []string                    `json:"bullet_points"`
	Attributes         []ListingAttribute          `json:"attributes"`
	CategoryAttributes *[]ListingCategoryAttribute `json:"category_attributes,omitempty"`
	Images             *[]Image                    `json:"images,omitempty"`
	Variants           *[]ListingVariant           `json:"variants,omitempty"`
}

type ListingAttribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type ListingCategoryAttribute struct {
	CategoryID string `json:"category_id"`
	ID         string `json:"id"`
	ValueID    string `json:"value_id,omitempty"`
	Value      string `json:"value"`
	UnitID     string `json:"unit_id,omitempty"`
}

var listingCategoryAttributeID = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,79}$`)

type ListingVariant struct {
	SellerSKU  string            `json:"seller_sku"`
	Attributes map[string]string `json:"attributes"`
	Price      string            `json:"price"`
	Currency   string            `json:"currency"`
	ImageURLs  []string          `json:"image_urls"`
}

func allNumeric(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(value) <= 15
}

func listingMoney(value string) bool {
	if value == "" {
		return true
	}
	return moneyPattern.MatchString(value)
}
func listingDecimal(value string, max float64, integer bool) bool {
	if value == "" {
		return true
	}
	n, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsInf(n, 0) || math.IsNaN(n) || n <= 0 || n > max {
		return false
	}
	if integer && math.Trunc(n) != n {
		return false
	}
	if len(value) > 20 || strings.ContainsAny(value, "eE+-") {
		return false
	}
	return true
}
func validListingText(text string, max int) bool {
	return utf8.RuneCountInString(text) <= max
}

func (f ListingFields) Validate() error {
	if !validListingText(f.Titles.English, 500) || !validListingText(f.Titles.Spanish, 500) ||
		!validListingText(f.Titles.Portuguese, 500) ||
		!validListingText(f.CategoryID, 60) ||
		!validListingText(f.CategoryPath, 400) || !validListingText(f.Brand, 120) ||
		!validListingText(f.Model, 120) || !validListingText(f.SellerSKU, 120) ||
		!validListingText(f.GTIN, 50) ||
		!listingMoney(f.PurchasePrice) ||
		!listingDecimal(f.WeightG, 1000000, true) ||
		!listingDecimal(f.LengthCM, 10000, false) ||
		!listingDecimal(f.WidthCM, 10000, false) ||
		!listingDecimal(f.HeightCM, 10000, false) {
		return errors.New("invalid listing preparation fields")
	}
	if f.PurchaseCurrency != "" && !currencyPattern.MatchString(f.PurchaseCurrency) {
		return errors.New("purchase_currency must contain 3 uppercase letters, e.g. CNY or USD")
	}
	if f.CategoryID != "" && (!strings.HasPrefix(f.CategoryID, "CBT") ||
		len(f.CategoryID) < 4 || !allNumeric(f.CategoryID[3:])) {
		return errors.New("category draft must use CBT general category; select marketplace at publishing")
	}
	if len(f.BulletPoints) > 20 || len(f.Attributes) > 60 {
		return errors.New("too many listing attributes")
	}
	for _, bullet := range f.BulletPoints {
		if !validListingText(bullet, 500) {
			return errors.New("bullet point too long")
		}
	}
	for _, attribute := range f.Attributes {
		if !validListingText(attribute.Name, 120) || !validListingText(attribute.Value, 500) {
			return errors.New("attribute too long")
		}
	}
	if f.CategoryAttributes != nil {
		if len(*f.CategoryAttributes) > 180 {
			return errors.New("too many category attribute fields")
		}
		seen := make(map[[2]string]bool)
		for _, field := range *f.CategoryAttributes {
			key := [2]string{field.CategoryID, field.ID}
			if !strings.HasPrefix(field.CategoryID, "CBT") || !allNumeric(strings.TrimPrefix(field.CategoryID, "CBT")) ||
				!listingCategoryAttributeID.MatchString(field.ID) || seen[key] ||
				!validListingText(field.Value, 500) || !validListingText(field.ValueID, 80) ||
				!validListingText(field.UnitID, 80) || (field.ValueID != "" && strings.TrimSpace(field.Value) == "") {
				return errors.New("invalid category attribute field")
			}
			seen[key] = true
		}
	}
	if f.Images != nil {
		if len(*f.Images) > 60 {
			return errors.New("too many images")
		}
		for _, im := range *f.Images {
			if !validImageURL(im.URL) || (im.Type != "main" && im.Type != "additional") {
				return errors.New("invalid product image")
			}
		}
	}
	if f.Variants != nil {
		if len(*f.Variants) > 100 {
			return errors.New("too many variants")
		}
		for _, v := range *f.Variants {
			if !validListingText(v.SellerSKU, 120) || len(v.Attributes) > 12 ||
				!listingMoney(v.Price) || (v.Currency != "" && !currencyPattern.MatchString(v.Currency)) ||
				len(v.ImageURLs) > 30 {
				return errors.New("invalid variant fields")
			}
			for name, value := range v.Attributes {
				if name == "" || value == "" || !validListingText(name, 80) || !validListingText(value, 200) {
					return errors.New("invalid variant attribute")
				}
			}
			for _, u := range v.ImageURLs {
				if !validImageURL(u) {
					return errors.New("invalid variant image")
				}
			}
		}
	}
	return nil
}
