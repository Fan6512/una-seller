package collections

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// ListingFields are editable preparation data, not official platform attributes.
// Workbench drafts are independent of listing markets and site prices.
type LocalizedTitles struct {
	English    string `json:"en"`
	Spanish    string `json:"es"`
	Portuguese string `json:"pt"`
}

type LocalizedDescriptions = LocalizedTitles

type DescriptionTranslation struct {
	EnglishDigest          string `json:"en_digest,omitempty"`
	SpanishSourceDigest    string `json:"es_source_digest,omitempty"`
	PortugueseSourceDigest string `json:"pt_source_digest,omitempty"`
	SpanishManualEdited    bool   `json:"es_manual_edited,omitempty"`
	PortugueseManualEdited bool   `json:"pt_manual_edited,omitempty"`
}

func EnglishDescriptionDigest(text string) string {
	if strings.TrimSpace(text) == "" {
		return ""
	}
	digest := sha256.Sum256([]byte(norm.NFC.String(text)))
	return hex.EncodeToString(digest[:])
}

var ErrInvalidTranslationSource = errors.New("description translation source does not match the saved English description")

const LocalizedTitleLimit = 60

var ErrLocalizedTitleTooLong = errors.New("MX Spanish and BR Portuguese titles must be at most 60 characters")

func (titles LocalizedTitles) NFC() LocalizedTitles {
	return LocalizedTitles{norm.NFC.String(titles.English), norm.NFC.String(titles.Spanish), norm.NFC.String(titles.Portuguese)}
}

func (titles LocalizedTitles) exceedsRegionalLimits() bool {
	return utf8.RuneCountInString(titles.Spanish) > LocalizedTitleLimit ||
		utf8.RuneCountInString(titles.Portuguese) > LocalizedTitleLimit
}

func (titles LocalizedTitles) validateRegionalLimits(previous LocalizedTitles) error {
	if (utf8.RuneCountInString(titles.Spanish) > LocalizedTitleLimit && titles.Spanish != previous.Spanish) ||
		(utf8.RuneCountInString(titles.Portuguese) > LocalizedTitleLimit && titles.Portuguese != previous.Portuguese) {
		return ErrLocalizedTitleTooLong
	}
	return nil
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
	Titles                 LocalizedTitles             `json:"titles"`
	Descriptions           *LocalizedDescriptions      `json:"descriptions,omitempty"`
	DescriptionTranslation *DescriptionTranslation     `json:"description_translation,omitempty"`
	CategoryID             string                      `json:"category_id"`
	CategoryPath           string                      `json:"category_path"`
	Brand                  string                      `json:"brand"`
	Model                  string                      `json:"model"`
	SellerSKU              string                      `json:"seller_sku"`
	GTIN                   string                      `json:"gtin"`
	PurchasePrice          string                      `json:"purchase_price"`
	PurchaseCurrency       string                      `json:"purchase_currency"`
	TargetNetProceedsUSD   *string                     `json:"target_net_proceeds_usd,omitempty"`
	WeightG                string                      `json:"weight_g"`
	LengthCM               string                      `json:"length_cm"`
	WidthCM                string                      `json:"width_cm"`
	HeightCM               string                      `json:"height_cm"`
	BulletPoints           []string                    `json:"bullet_points"`
	Attributes             []ListingAttribute          `json:"attributes"`
	CategoryAttributes     *[]ListingCategoryAttribute `json:"category_attributes,omitempty"`
	Images                 *[]Image                    `json:"images,omitempty"`
	Variants               *[]ListingVariant           `json:"variants,omitempty"`
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
	SellerSKU            string                     `json:"seller_sku"`
	Attributes           map[string]string          `json:"attributes"`
	Price                string                     `json:"price"`
	Currency             string                     `json:"currency"`
	ImageURLs            []string                   `json:"image_urls"`
	PurchaseCostCNY      string                     `json:"purchase_cost_cny,omitempty"`
	TargetNetProceedsUSD string                     `json:"target_net_proceeds_usd,omitempty"`
	CategoryAttributes   []ListingCategoryAttribute `json:"category_attributes,omitempty"`
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

func validListingTitle(title string) bool {
	return validListingText(norm.NFC.String(title), 500)
}

var listingMarkup = regexp.MustCompile("(?m)<(?:/?[A-Za-z][A-Za-z0-9:-]*(?:[ \\t\\r\\n][^<>]*)?[ \\t\\r\\n]*/?>|[!?][^<>]*>)|^[ \\t]{0,3}#{1,6}(?:[ \\t]|$)|\\[[^\\]\\n]+\\][ \\t]*(?:\\(|\\[)|(?:\\*\\*|__|~~)[^\\n]+(?:\\*\\*|__|~~)|`")

func IsPlainListingText(text string) bool {
	return utf8.ValidString(text) && !listingMarkup.MatchString(text) && strings.IndexFunc(text, func(r rune) bool {
		return unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t'
	}) < 0
}

func (f ListingFields) normalizedCopy() *ListingFields {
	f.Titles = f.Titles.NFC()
	if f.Descriptions != nil {
		descriptions := f.Descriptions.NFC()
		f.Descriptions = &descriptions
	}
	return &f
}

func (f ListingFields) needsPrevious() bool {
	return f.Descriptions != nil || f.Titles.NFC().exceedsRegionalLimits()
}

func (f ListingFields) prepareForSave(previous ListingFields) (*ListingFields, error) {
	next := f.normalizedCopy()
	if err := next.Titles.validateRegionalLimits(previous.Titles.NFC()); err != nil {
		return nil, err
	}
	if next.Descriptions == nil {
		if next.DescriptionTranslation != nil {
			return nil, ErrInvalidTranslationSource
		}
		return next, nil
	}

	var current, proposed DescriptionTranslation
	if previous.DescriptionTranslation != nil {
		current = *previous.DescriptionTranslation
	}
	if next.DescriptionTranslation != nil {
		proposed = *next.DescriptionTranslation
	}
	var savedEnglish string
	if previous.Descriptions != nil {
		savedEnglish = previous.Descriptions.English
	}
	savedDigest := EnglishDescriptionDigest(savedEnglish)
	current.EnglishDigest = EnglishDescriptionDigest(next.Descriptions.English)
	for _, locale := range []struct {
		text, oldText, proposal string
		source                  *string
		manual                  *bool
	}{
		{next.Descriptions.Spanish, "", proposed.SpanishSourceDigest, &current.SpanishSourceDigest, &current.SpanishManualEdited},
		{next.Descriptions.Portuguese, "", proposed.PortugueseSourceDigest, &current.PortugueseSourceDigest, &current.PortugueseManualEdited},
	} {
		if previous.Descriptions != nil {
			if locale.source == &current.SpanishSourceDigest {
				locale.oldText = previous.Descriptions.Spanish
			} else {
				locale.oldText = previous.Descriptions.Portuguese
			}
		}
		if strings.TrimSpace(locale.text) == "" {
			*locale.source = ""
			*locale.manual = false
			continue
		}
		if locale.proposal != "" {
			if locale.proposal != savedDigest || locale.proposal != current.EnglishDigest {
				return nil, ErrInvalidTranslationSource
			}
			*locale.source = locale.proposal
			*locale.manual = false
		} else if locale.text != locale.oldText {
			*locale.manual = true
		}
	}
	next.DescriptionTranslation = &current
	return next, nil
}

func (f ListingFields) Validate() error {
	if !validListingTitle(f.Titles.English) || !validListingTitle(f.Titles.Spanish) ||
		!validListingTitle(f.Titles.Portuguese) ||
		!validListingText(f.CategoryID, 60) ||
		!validListingText(f.CategoryPath, 400) || !validListingText(f.Brand, 120) ||
		!validListingText(f.Model, 120) || !validListingText(f.SellerSKU, 120) ||
		!validListingText(f.GTIN, 50) ||
		!listingMoney(f.PurchasePrice) || (f.TargetNetProceedsUSD != nil && !listingMoney(*f.TargetNetProceedsUSD)) ||
		!listingDecimal(f.WeightG, 1000000, true) ||
		!listingDecimal(f.LengthCM, 10000, false) ||
		!listingDecimal(f.WidthCM, 10000, false) ||
		!listingDecimal(f.HeightCM, 10000, false) {
		return errors.New("invalid listing preparation fields")
	}
	if f.Descriptions != nil {
		descriptions := f.Descriptions.NFC()
		for _, text := range []string{descriptions.English, descriptions.Spanish, descriptions.Portuguese} {
			if !validListingText(text, 20000) || !IsPlainListingText(text) {
				return errors.New("localized descriptions must be plain text of at most 20000 characters each")
			}
		}
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
		if !validListingText(bullet, 500) || !IsPlainListingText(bullet) {
			return errors.New("bullet point too long")
		}
	}
	for _, attribute := range f.Attributes {
		if !validListingText(attribute.Name, 120) || !validListingText(attribute.Value, 500) {
			return errors.New("attribute too long")
		}
	}
	if f.CategoryAttributes != nil {
		if err := validateListingCategoryAttributes(*f.CategoryAttributes, 180); err != nil {
			return err
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
				!listingMoney(v.Price) || !listingMoney(v.PurchaseCostCNY) || !listingMoney(v.TargetNetProceedsUSD) || (v.Currency != "" && !currencyPattern.MatchString(v.Currency)) ||
				len(v.ImageURLs) > 30 {
				return errors.New("invalid variant fields")
			}
			for name, value := range v.Attributes {
				if name == "" || value == "" || !validListingText(name, 80) || !validListingText(value, 200) {
					return errors.New("invalid variant attribute")
				}
			}
			if err := validateListingCategoryAttributes(v.CategoryAttributes, 12); err != nil {
				return err
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

func validateListingCategoryAttributes(fields []ListingCategoryAttribute, limit int) error {
	if len(fields) > limit {
		return errors.New("too many category attribute fields")
	}
	seen := make(map[[2]string]bool)
	for _, field := range fields {
		key := [2]string{field.CategoryID, field.ID}
		if !strings.HasPrefix(field.CategoryID, "CBT") || !allNumeric(strings.TrimPrefix(field.CategoryID, "CBT")) ||
			!listingCategoryAttributeID.MatchString(field.ID) || seen[key] ||
			!validListingText(field.Value, 500) || !validListingText(field.ValueID, 80) ||
			!validListingText(field.UnitID, 80) || (field.ValueID != "" && strings.TrimSpace(field.Value) == "") {
			return errors.New("invalid category attribute field")
		}
		seen[key] = true
	}
	return nil
}
