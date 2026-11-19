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
	CategoryID string             `json:"category_id"`
	Attributes json.RawMessage    `json:"attributes"`
	Path       []CategoryPathNode `json:"path,omitempty"`
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

	// Fixed host/path; category must match the authenticated marketplace site.
	// Keep the full platform attribute schema so value lists, tags and units
	// remain available to the category-driven editor without manual mapping.
	var raw json.RawMessage
	if _, err = s.listingGET(readCtx, auth.AccessToken, "/categories/"+category+"/attributes", &raw); err != nil {
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
	if _, e := s.listingGET(readCtx, auth.AccessToken, "/categories/"+category, &info); e == nil &&
		info.ID == category && len(info.PathFromRoot) > 0 {
		result.Path = info.PathFromRoot
	}
	return result, nil
}
