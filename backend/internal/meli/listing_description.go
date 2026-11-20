package meli

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrListingDescriptionMissing     = errors.New("published listing was not found for this store and site")
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
	findListing := func(snapshot tokenSnapshot, child string) error {
		return s.db.QueryRow(readCtx, `SELECT item.details->>'cbt_item_id', item.details->>'parent_item_id'
		FROM app.meli_published_listings item
		JOIN app.meli_store_connections c ON c.id=item.store_id
		WHERE (item.store_id,item.site_id,item.logistic_type,item.listing_id)=($1::uuid,$2,$3,$4)
		  AND item.details->>'seller_id'=$5 AND c.owner_user_id=$6::uuid AND c.status='connected'
		  AND c.marketplaces=$7::jsonb AND c.marketplaces_status=$8 AND c.access_token_ciphertext=$9`,
			store, scope.Site, scope.Logistic, listing, child, owner, snapshot.marketsJSON, snapshot.marketsStatus, snapshot.accessCipher).Scan(&global, &parent)
	}
	err = findListing(row, seller)
	if errors.Is(err, pgx.ErrNoRows) {
		current, child, scopeErr := s.listingScope(readCtx, owner, store, scope)
		if scopeErr != nil {
			return "", scopeErr
		}
		if bytes.Equal(row.accessCipher, current.accessCipher) && bytes.Equal(row.marketsJSON, current.marketsJSON) {
			return "", ErrListingDescriptionMissing
		}
		// Report a changed snapshot only if the listing still belongs to the current scope.
		switch err = findListing(current, child); {
		case errors.Is(err, pgx.ErrNoRows):
			return "", ErrListingDescriptionMissing
		case err != nil:
			return "", err
		default:
			return "", ErrListingAuthorizationChanged
		}
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
		Text                string `json:"text"`
		PlainText           string `json:"plain_text"`
		TranslatedPlainText string `json:"translated_plain_text"`
	}
	if _, err := s.listingGET(readCtx, auth.AccessToken, path, &content); err != nil {
		if errors.Is(err, ErrListingResponse) {
			return "", errors.New("invalid marketplace description response")
		}
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
