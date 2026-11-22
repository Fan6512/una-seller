package collections

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Field keeps missing and explicit null distinct. A missing field never clears saved data.
type Field[T any] struct {
	Present bool `json:"-"`
	Null    bool `json:"-"`
	Value   T    `json:"-"`
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Null = true
		return nil
	}
	return json.Unmarshal(data, &f.Value)
}

type ImageInput struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type VariantInput struct {
	Attributes map[string]string `json:"attributes"`
	Price      Field[string]     `json:"price"`
	Currency   Field[string]     `json:"currency"`
	ImageURLs  Field[[]string]   `json:"image_urls"`
}

type Input struct {
	Platform      string                `json:"source_platform"`
	Site          string                `json:"source_site"`
	SourceURL     string                `json:"source_url"`
	Title         Field[string]         `json:"title"`
	Price         Field[string]         `json:"price"`
	Currency      Field[string]         `json:"currency"`
	Category      Field[string]         `json:"category"`
	Description   Field[string]         `json:"description"`
	BulletPoints  Field[[]string]       `json:"bullet_points"`
	Images        Field[[]ImageInput]   `json:"images"`
	Variants      Field[[]VariantInput] `json:"variants"`
	Status        string                `json:"collection_status"`
	Warnings      []string              `json:"warnings"`
	MissingFields []string              `json:"missing_fields"`
	PageState     string                `json:"page_state"`
}

type Image struct {
	URL  string `json:"url"`
	Type string `json:"type"`
}

type Variant struct {
	Attributes map[string]string `json:"attributes"`
	Price      *string           `json:"price"`
	Currency   *string           `json:"currency"`
	ImageURLs  []string          `json:"image_urls"`
}

type Product struct {
	ID               string    `json:"id"`
	OwnProductID     *string   `json:"own_product_id"`
	Platform         string    `json:"source_platform"`
	Site             string    `json:"source_site"`
	SourceURL        string    `json:"source_url"`
	NormalizedURL    string    `json:"normalized_url"`
	Title            string    `json:"title"`
	Price            *string   `json:"price"`
	Currency         *string   `json:"currency"`
	Category         *string   `json:"category"`
	Description      *string   `json:"description"`
	BulletPoints     []string  `json:"bullet_points"`
	Images           []Image   `json:"images"`
	Variants         []Variant `json:"variants"`
	Status           string    `json:"collection_status"`
	Warnings         []string  `json:"collection_warnings"`
	FirstCollectedAt string    `json:"first_collected_at"`
	LastCollectedAt  string    `json:"last_collected_at"`
}

type Summary struct {
	ID              string  `json:"id"`
	OwnProductID    *string `json:"own_product_id"`
	Title           string  `json:"title"`
	Platform        string  `json:"source_platform"`
	Site            string  `json:"source_site"`
	Price           *string `json:"price"`
	Currency        *string `json:"currency"`
	MainImageURL    *string `json:"main_image_url"`
	Status          string  `json:"collection_status"`
	LastCollectedAt string  `json:"last_collected_at"`
}

var moneyPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,15})(\.[0-9]{1,4})?$`)
var uuidPattern = regexp.MustCompile(`(?i)^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

func ValidUserID(s string) bool { return uuidPattern.MatchString(s) }

func validMoney(f Field[string]) bool {
	return !f.Present || f.Null || moneyPattern.MatchString(f.Value)
}
func validCurrency(f Field[string]) bool {
	return !f.Present || f.Null || currencyPattern.MatchString(f.Value)
}

func normalizedURL(platform, site, raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Hostname() == "" || u.Port() != "" {
		return "", errors.New("source_url must be an HTTPS product URL")
	}
	host := strings.ToLower(u.Hostname())
	if platform == "amazon" {
		if site != "US" || (host != "amazon.com" && host != "www.amazon.com") {
			return "", errors.New("amazon only supports US / amazon.com")
		}
		host = "www.amazon.com"
		parts := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")
		for i := 0; i+1 < len(parts); i++ {
			if (parts[i] == "dp" || parts[i] == "product") && len(parts[i+1]) == 10 {
				u.Path = "/dp/" + strings.ToUpper(parts[i+1])
				break
			}
		}
	} else if platform == "mercadolibre" {
		suffix := map[string]string{"MLM": "mercadolibre.com.mx", "MLB": "mercadolivre.com.br", "MLA": "mercadolibre.com.ar"}[site]
		if suffix == "" || (host != suffix && !strings.HasSuffix(host, "."+suffix)) {
			return "", errors.New("source_site does not match Mercado Libre URL")
		}
		host = suffix
	} else {
		return "", errors.New("unsupported source_platform")
	}
	if u.Path == "" || u.Path == "/" {
		return "", errors.New("missing product path")
	}
	u.Host = host
	u.RawQuery, u.Fragment, u.RawFragment = "", "", ""
	u.Path = strings.TrimRight(u.Path, "/")
	return u.String(), nil
}

func (in Input) Validate() (string, error) {
	norm, err := normalizedURL(in.Platform, in.Site, in.SourceURL)
	if err != nil {
		return "", err
	}
	if in.Status != "complete" && in.Status != "partial" {
		return "", errors.New("collection_status must be complete or partial")
	}
	if in.Title.Present && !in.Title.Null && len(strings.TrimSpace(in.Title.Value)) > 500 {
		return "", errors.New("title too long")
	}
	if !validMoney(in.Price) || !validCurrency(in.Currency) {
		return "", errors.New("invalid price or currency")
	}
	if in.BulletPoints.Present && !in.BulletPoints.Null && len(in.BulletPoints.Value) > 50 {
		return "", errors.New("too many bullet_points")
	}
	if in.Images.Present && !in.Images.Null {
		if len(in.Images.Value) > 60 {
			return "", errors.New("too many images")
		}
		for _, im := range in.Images.Value {
			if !validImageURL(im.URL) || (im.Type != "main" && im.Type != "additional") {
				return "", errors.New("invalid image")
			}
		}
	}
	if in.Variants.Present && !in.Variants.Null {
		if len(in.Variants.Value) > 100 {
			return "", errors.New("too many variants")
		}
		keys := make(map[string]bool)
		for _, v := range in.Variants.Value {
			if len(v.Attributes) == 0 || len(v.Attributes) > 12 || !validMoney(v.Price) || !validCurrency(v.Currency) {
				return "", errors.New("invalid variant")
			}
			for k, s := range v.Attributes {
				if k == "" || s == "" || len(k) > 80 || len(s) > 200 {
					return "", errors.New("invalid variant attribute")
				}
			}
			key, _ := json.Marshal(v.Attributes)
			if keys[string(key)] {
				return "", errors.New("duplicate variant attributes")
			}
			keys[string(key)] = true
			if v.ImageURLs.Present && !v.ImageURLs.Null {
				if len(v.ImageURLs.Value) > 30 {
					return "", errors.New("too many variant images")
				}
				for _, u := range v.ImageURLs.Value {
					if !validImageURL(u) {
						return "", errors.New("invalid variant image URL")
					}
				}
			}
		}
	}
	if len(in.Warnings) > 30 || len(in.MissingFields) > 30 {
		return "", errors.New("too many warnings")
	}
	for _, v := range append(append([]string{}, in.Warnings...), in.MissingFields...) {
		if len(v) > 200 {
			return "", errors.New("warning too long")
		}
	}
	if in.PageState != "" && in.PageState != "normal" { // Unreliable pages must never be presented as complete.
		if in.Status == "complete" {
			return "", errors.New("non-normal page cannot be complete")
		}
	}
	return norm, nil
}

func validImageURL(raw string) bool {
	u, e := url.Parse(raw)
	return e == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Port() == "" && len(raw) <= 2000
}

func nullable[T any](f Field[T], partial bool, existing *T) *T {
	if !f.Present {
		return existing
	}
	if f.Null {
		if partial {
			return existing
		}
		return nil
	}
	if v, ok := any(f.Value).(string); ok && v == "" && partial {
		return existing
	}
	v := f.Value
	return &v
}

func validateTitle(in Input, oldTitle string) (string, error) {
	if !in.Title.Present || in.Title.Null || strings.TrimSpace(in.Title.Value) == "" {
		if oldTitle == "" {
			return "", fmt.Errorf("title is required for a new collection")
		}
		return oldTitle, nil
	}
	return strings.TrimSpace(in.Title.Value), nil
}
