package collections

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ListingFields are editable preparation data, not official platform attributes.
// Category IDs and site prices require Mercado Libre validation before publishing.
type ListingFields struct {
	CategorySite   string           `json:"category_site"`
	CategoryID     string           `json:"category_id"`
	CategoryPath   string           `json:"category_path"`
	Brand          string           `json:"brand"`
	Model          string           `json:"model"`
	SellerSKU      string           `json:"seller_sku"`
	GTIN           string           `json:"gtin"`
	PurchasePrice  string           `json:"purchase_price"`
	PurchaseCurrency string         `json:"purchase_currency"`
	WeightG        string           `json:"weight_g"`
	LengthCM       string           `json:"length_cm"`
	WidthCM        string           `json:"width_cm"`
	HeightCM       string           `json:"height_cm"`
	SitePrices     []SitePrice      `json:"site_prices"`
	BulletPoints   []string         `json:"bullet_points"`
	Attributes     []ListingAttribute `json:"attributes"`
	Images         *[]Image         `json:"images,omitempty"`
	Variants       *[]ListingVariant `json:"variants,omitempty"`
}

type SitePrice struct {
	Site     string `json:"site"`
	Price    string `json:"price"`
	Currency string `json:"currency"`
}
type ListingAttribute struct {
	Name string `json:"name"`
	Value string `json:"value"`
}
type ListingVariant struct {
	SellerSKU string            `json:"seller_sku"`
	Attributes map[string]string `json:"attributes"`
	Price string `json:"price"`
	Currency string `json:"currency"`
	ImageURLs []string `json:"image_urls"`
}

func listingMoney(value string) bool {
	if value == "" { return true }
	return moneyPattern.MatchString(value)
}
func listingDecimal(value string, max float64, integer bool) bool {
	if value == "" { return true }
	n, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsInf(n, 0) || math.IsNaN(n) || n <= 0 || n > max { return false }
	if integer && math.Trunc(n) != n { return false }
	if len(value) > 20 || strings.ContainsAny(value, "eE+-") { return false }
	return true
}
func validListingText(text string, max int) bool {
	return utf8.RuneCountInString(text) <= max
}

func (f ListingFields) Validate() error {
	if !validListingText(f.CategorySite, 4) || !validListingText(f.CategoryID, 60) ||
		!validListingText(f.CategoryPath, 400) || !validListingText(f.Brand, 120) ||
		!validListingText(f.Model, 120) || !validListingText(f.SellerSKU, 120) ||
		!validListingText(f.GTIN, 50) ||
		!listingMoney(f.PurchasePrice) ||
		(f.PurchaseCurrency != "" && !currencyPattern.MatchString(f.PurchaseCurrency)) ||
		!listingDecimal(f.WeightG, 1000000, true) ||
		!listingDecimal(f.LengthCM, 10000, false) ||
		!listingDecimal(f.WidthCM, 10000, false) ||
		!listingDecimal(f.HeightCM, 10000, false) {
		return errors.New("invalid listing preparation fields")
	}
	allowedSites:=map[string]string{"MLM":"MXN","MLB":"BRL","MLA":"ARS","MLC":"CLP","MCO":"COP"}
	if f.CategorySite != "" && allowedSites[f.CategorySite] == "" {
		return errors.New("invalid target category site")
	}
	if len(f.SitePrices)>5 || len(f.BulletPoints)>20 || len(f.Attributes)>60 {
		return errors.New("too many site prices or listing attributes")
	}
	seen:=map[string]bool{}
	for _,p:=range f.SitePrices {
		if allowedSites[p.Site]=="" || seen[p.Site] || !listingMoney(p.Price) ||
			(p.Currency!="" && p.Currency!=allowedSites[p.Site]) {
			return errors.New("invalid site price or currency")
		}
		seen[p.Site]=true
	}
	for _,bullet:=range f.BulletPoints {
		if !validListingText(bullet,500) {return errors.New("bullet point too long")}
	}
	for _,attribute:=range f.Attributes {
		if !validListingText(attribute.Name,120)||!validListingText(attribute.Value,500) {
			return errors.New("attribute too long")
		}
	}
	if f.Images!=nil {
		if len(*f.Images)>60 {return errors.New("too many images")}
		for _,im:=range *f.Images {
			if !validImageURL(im.URL)||(im.Type!="main"&&im.Type!="additional") {
				return errors.New("invalid product image")
			}
		}
	}
	if f.Variants!=nil {
		if len(*f.Variants)>100 {return errors.New("too many variants")}
		for _,v:=range *f.Variants {
			if !validListingText(v.SellerSKU,120)||len(v.Attributes)>12||
				!listingMoney(v.Price)||(v.Currency!=""&&!currencyPattern.MatchString(v.Currency))||
				len(v.ImageURLs)>30 {
				return errors.New("invalid variant fields")
			}
			for name,value:=range v.Attributes {
				if name==""||value==""||!validListingText(name,80)||!validListingText(value,200) {
					return errors.New("invalid variant attribute")
				}
			}
			for _,u:=range v.ImageURLs {
				if !validImageURL(u){return errors.New("invalid variant image")}
			}
		}
	}
	return nil
}
