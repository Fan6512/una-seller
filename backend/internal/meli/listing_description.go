package meli

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrListingDescriptionMissing = errors.New("published listing was not found for this store and site")
	ErrListingDescriptionLinkMissing = errors.New("published listing lacks a usable global description id")
)

// GetPublishedDescription reads one listing's platform description on demand.
// It is deliberately outside the batched sync loop to preserve its time budget.
func (s *Service) GetPublishedDescription(ctx context.Context, owner, store string, scope ListingScope, listing string) (string, error) {
	if !listingID.MatchString(listing) || scope.Site == "CBT" || scope.Logistic == "" {
		return "", ErrListingResponse
	}
	readCtx, cancel := context.WithTimeout(ctx, 9*time.Second)
	defer cancel()

	row, seller, err := s.listingScope(readCtx, owner, store, scope)
	if err != nil {
		return "", err
	}
	if row.expires == nil || !row.expires.After(time.Now().Add(9*time.Second)) {
		return "", ErrListingTokenExpired
	}

	// Do not contact the platform for a listing that does not belong to the
	// authenticated owner, store, verified site and currently linked child seller.
	var global, parent *string
	err = s.db.QueryRow(readCtx, `SELECT details->>'cbt_item_id', details->>'parent_item_id'
		FROM app.meli_published_listings
		WHERE (store_id,site_id,logistic_type,listing_id)=($1::uuid,$2,$3,$4)
		  AND details->>'seller_id'=$5`, store, scope.Site, scope.Logistic, listing, seller).Scan(&global, &parent)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrListingDescriptionMissing
	}
	if err != nil {
		return "", err
	}
	auth, err := s.accessFromSnapshot(owner, row)
	if err != nil {
		return "", err
	}

	// CBT supports a site-localized global description; ordinary site listings
	// have the documented /items/{id}/description read endpoint.
	sourceID := ""
	for _, candidate := range []*string{global, parent} {
		if candidate != nil && strings.HasPrefix(*candidate, "CBT") && listingID.MatchString(*candidate) {
			sourceID = *candidate
			break
		}
	}
	path := ""
	if sourceID != "" {
		query := url.Values{"site_id": {scope.Site}, "logistic_type": {scope.Logistic}}
		path = "/marketplace/items/" + sourceID + "/description?" + query.Encode()
	} else if strings.HasPrefix(listing, scope.Site) {
		path = "/items/" + listing + "/description"
	} else {
		return "", ErrListingDescriptionLinkMissing
	}

	var content struct {
		Text string `json:"text"`
		PlainText string `json:"plain_text"`
		TranslatedPlainText string `json:"translated_plain_text"`
	}
	if _, err := s.listingGET(readCtx, auth.AccessToken, path, &content); err != nil {
		return "", err
	}
	if content.PlainText != "" {
		return content.PlainText, nil
	}
	if content.TranslatedPlainText != "" {
		return content.TranslatedPlainText, nil
	}
	return content.Text, nil
}
