package meli

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
)

// Category rules are fetched from Mercado Libre for the specific authorized site.
// This endpoint never changes the publication and never exposes the platform token.
var ErrCategoryInvalid = errors.New("category is invalid for the selected marketplace")

var categoryIDPattern = regexp.MustCompile(`^[A-Z]{3}[0-9]{1,15}$`)

type CategoryPathNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CategoryRules struct {
	Translations           map[string]string  `json:"translations_zh,omitempty"`
	CategoryID             string             `json:"category_id"`
	Attributes             json.RawMessage    `json:"attributes"`
	Path                   []CategoryPathNode `json:"path,omitempty"`
	TechnicalSpecsVerified bool               `json:"technical_specs_verified,omitempty"`
	TechnicalRequirements  []string           `json:"technical_requirements,omitempty"`
	SpecialRequirements    bool               `json:"special_requirements,omitempty"`
}

func validateCategoryForSite(category, site string) bool {
	return categoryIDPattern.MatchString(category) && site != "CBT" &&
		len(site) == 3 && strings.HasPrefix(category, site)
}

func (s *Service) GetCategoryRules(ctx context.Context, owner, store string, scope ListingScope, category string) (CategoryRules, error) {
	if !validateCategoryForSite(category, scope.Site) {
		return CategoryRules{}, ErrCategoryInvalid
	}
	readCtx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()

	row, _, err := s.listingScope(readCtx, owner, store, scope)
	if err != nil {
		return CategoryRules{}, err
	}
	if row.expires == nil || !row.expires.After(time.Now().Add(9*time.Second)) {
		return CategoryRules{}, ErrListingTokenExpired
	}
	auth, err := s.accessFromSnapshot(owner, row)
	if err != nil {
		return CategoryRules{}, err
	}

	return s.readCategoryRules(readCtx, auth.AccessToken, category)
}

func (s *Service) readCategoryRules(readCtx context.Context, token, category string) (CategoryRules, error) {
	// Fixed host/path; caller already validated the category and scope.
	// Keep the full platform attribute schema so value lists, tags and units
	// remain available to the category-driven editor without manual mapping.
	var raw json.RawMessage
	if _, err := s.listingGET(readCtx, token, "/categories/"+category+"/attributes", &raw, nil); err != nil {
		return CategoryRules{}, err
	}
	if len(raw) < 2 || raw[0] != '[' {
		return CategoryRules{}, ErrListingResponse
	}
	result := CategoryRules{CategoryID: category, Attributes: raw}
	// The readable category path helps sellers recognize the chosen category.
	// Failure here must not hide the successfully fetched required-attribute rules.
	var info struct {
		ID           string             `json:"id"`
		PathFromRoot []CategoryPathNode `json:"path_from_root"`
	}
	if _, e := s.listingGET(readCtx, token, "/categories/"+category, &info, nil); e == nil &&
		info.ID == category && len(info.PathFromRoot) > 0 {
		result.Path = info.PathFromRoot
	}
	return result, nil
}

// CBT category rules are global product metadata. Store owner authorization is
// required, and the route does not accept local marketplace category IDs.
func (s *Service) GetCBTCategoryRules(ctx context.Context, owner, category string) (CategoryRules, error) {
	if !cbtCategoryID.MatchString(category) {
		return CategoryRules{}, ErrCategoryInvalid
	}
	readCtx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()
	stores, err := s.List(readCtx, owner)
	if err != nil {
		return CategoryRules{}, err
	}
	store := preferredCBTStore(stores)
	if store == "" {
		return CategoryRules{}, ErrCBTPredictorStore
	}
	auth, err := s.RequireStoreConnection(readCtx, owner, store)
	if err != nil {
		return CategoryRules{}, err
	}
	result, err := s.readCategoryRules(readCtx, auth.AccessToken, category)
	if err != nil {
		return CategoryRules{}, err
	}
	var spec struct {
		Groups *[]struct {
			Components []struct {
				Type       string `json:"component"`
				Attributes []struct {
					ID   string   `json:"id"`
					Tags []string `json:"tags"`
				} `json:"attributes"`
			} `json:"components"`
		} `json:"groups"`
	}
	if _, err := s.listingGET(readCtx, auth.AccessToken, "/categories/"+category+"/technical_specs/input", &spec, nil); err != nil || spec.Groups == nil {
		// Attribute form remains usable if the supplemental source is unavailable.
		return result, nil
	}
	result.TechnicalSpecsVerified = true
	seen := make(map[string]bool)
	for _, group := range *spec.Groups {
		for _, component := range group.Components {
			if strings.Contains(component.Type, "SIZE") || strings.Contains(component.Type, "VARIATION") {
				result.SpecialRequirements = true
			}
			for _, field := range component.Attributes {
				for _, tag := range field.Tags {
					switch tag {
					case "required":
						if field.ID != "" && !seen[field.ID] {
							result.TechnicalRequirements = append(result.TechnicalRequirements, field.ID)
							seen[field.ID] = true
						}
					case "conditional_required", "variation_attribute", "allow_variations", "size_grid_required":
						result.SpecialRequirements = true
					}
				}
			}
		}
	}
	return result, nil
}
