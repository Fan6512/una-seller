package meli

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var ErrCBTPredictorInput = errors.New("CBT category prediction requires a descriptive edited title")
var ErrCBTPredictorStore = errors.New("connect a Mercado Libre CBT store before requesting official category predictions")

var cbtCategoryID = regexp.MustCompile(`^CBT[0-9]{1,15}$`)

// A token-expired store remains authorized to refresh via RequireStoreConnection.
// Never skip it and incorrectly claim that no store is connected.
func preferredCBTStore(stores []Store) string {
	for _, store := range stores {
		if store.Status == "connected" || store.Status == "token_expired" {
			return store.ID
		}
	}
	return ""
}

// CBTCategorySuggestion is a suggestion, not a validated marketplace listing
// category. User must confirm and site-specific fields are checked at upload.
type CBTCategorySuggestion struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	DomainID     string `json:"domain_id"`
	DomainName   string `json:"domain_name"`
}

func (s *Service) PredictCBTCategories(ctx context.Context, owner, editedEnglishTitle string) ([]CBTCategorySuggestion, error) {
	title := strings.TrimSpace(editedEnglishTitle)
	if utf8.RuneCountInString(title) < 4 || utf8.RuneCountInString(title) > 500 ||
		len(strings.Fields(title)) < 2 {
		return nil, ErrCBTPredictorInput
	}
	readCtx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()
	stores, err := s.List(readCtx, owner)
	if err != nil {
		return nil, err
	}
	storeID := preferredCBTStore(stores)
	if storeID == "" {
		return nil, ErrCBTPredictorStore
	}
	auth, err := s.RequireStoreConnection(readCtx, owner, storeID)
	if err != nil {
		return nil, err
	}
	var raw []CBTCategorySuggestion
	// Global Selling official CBT predictor: no country/marketplace selection,
	// no mutation, no access token or supplier information returned to desktop.
	_, err = s.listingGET(readCtx, auth.AccessToken,
		"/marketplace/domain_discovery/search?q="+url.QueryEscape(title), &raw, nil)
	if err != nil {
		return nil, err
	}
	suggestions := make([]CBTCategorySuggestion, 0, 8)
	seen := make(map[string]bool)
	for _, item := range raw {
		if !cbtCategoryID.MatchString(item.CategoryID) || seen[item.CategoryID] {
			continue
		}
		seen[item.CategoryID] = true
		suggestions = append(suggestions, item)
		if len(suggestions) == 8 {
			break
		}
	}
	return suggestions, nil
}
