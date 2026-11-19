package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

const ListingBatchSize = 5

var (
	ErrListingBusy         = errors.New("listing sync is already running")
	ErrListingTokenExpired = errors.New("store Access Token expired")
	ErrListingResponse     = errors.New("unexpected marketplace listing response")
	listingID              = regexp.MustCompile(`^[A-Z]{3}[0-9]{1,30}$`)
	sellerID               = regexp.MustCompile(`^[0-9]{1,30}$`)
)

type ListingPlatformError int

func (e ListingPlatformError) Error() string {
	return fmt.Sprintf("marketplace listing request returned HTTP %d", e)
}

type ListingScope struct {
	Site     string
	Logistic string
}

type ListingPage struct {
	Items    []json.RawMessage `json:"items"`
	HasMore  bool              `json:"has_more"`
	SyncedAt *time.Time        `json:"synced_at"`
}

type ListingSync struct {
	Saved          int  `json:"saved"`
	Failed         int  `json:"failed"`
	Partial        int  `json:"partial"`
	NextOffset     *int `json:"next_offset"`
	ScanRequired   bool `json:"scan_required"`
	PlatformStatus int  `json:"platform_status,omitempty"`
}

func (s *Service) listingScope(ctx context.Context, owner, store string, scope ListingScope) (tokenSnapshot, string, error) {
	row, err := s.snapshot(ctx, owner, store)
	if err != nil {
		return row, "", err
	}
	if err := tokenStatusError(row.status); err != nil {
		return row, "", err
	}
	if scope.Site == "CBT" && scope.Logistic == "" {
		return row, row.sellerID, nil
	}
	if row.marketsStatus == "verified" && scope.Logistic != "" {
		var markets []Marketplace
		if err := json.Unmarshal(row.marketsJSON, &markets); err != nil {
			return row, "", err
		}
		for _, market := range markets {
			if market.SiteID == scope.Site && market.LogisticType == scope.Logistic {
				return row, market.UserID.String(), nil
			}
		}
	}
	return row, "", ErrSiteNotAuthorized
}

const ListingPageSize = 20

func (s *Service) ListPublished(ctx context.Context, owner, store string, scope ListingScope, offset int) (ListingPage, error) {
	row, seller, err := s.listingScope(ctx, owner, store, scope)
	if err != nil {
		return ListingPage{}, err
	}
	page := ListingPage{Items: []json.RawMessage{}}
	// Recheck the connection in the same SQL statement that reads its private rows.
	rows, err := s.db.Query(ctx, `SELECT sync.synced_at, item.details
      FROM app.meli_store_connections c
      LEFT JOIN app.meli_listing_syncs sync ON (sync.store_id,sync.site_id,sync.logistic_type)=(c.id,$3,$4)
      LEFT JOIN LATERAL (
        SELECT details,listing_id FROM app.meli_published_listings
        WHERE (store_id,site_id,logistic_type)=(c.id,$3,$4)
          AND details->>'seller_id'=$9
        ORDER BY listing_id OFFSET $5 LIMIT $6
      ) item ON true
      WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid AND c.status='connected'
        AND c.marketplaces=$7::jsonb AND c.marketplaces_status=$8
      ORDER BY item.listing_id`, store, owner, scope.Site, scope.Logistic, offset, ListingPageSize+1, row.marketsJSON, row.marketsStatus, seller)
	if err != nil {
		return ListingPage{}, err
	}
	defer rows.Close()
	found := false
	for rows.Next() {
		var detail json.RawMessage
		if err = rows.Scan(&page.SyncedAt, &detail); err != nil {
			return ListingPage{}, err
		}
		found = true
		if detail != nil {
			page.Items = append(page.Items, detail)
		}
	}
	if err = rows.Err(); err != nil {
		return ListingPage{}, err
	}
	if !found {
		return ListingPage{}, ErrStoreNotFound
	}
	if len(page.Items) > ListingPageSize {
		page.HasMore = true
		page.Items = page.Items[:ListingPageSize]
	}
	return page, nil
}

func (s *Service) SyncPublished(ctx context.Context, owner, store string, scope ListingScope, offset int) (ListingSync, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	row, seller, err := s.listingScope(ctx, owner, store, scope)
	if err != nil {
		return ListingSync{}, err
	}
	// This read-only platform operation must not trigger the Issue #21 refresh flow.
	if row.expires == nil || !row.expires.After(time.Now().Add(20*time.Second)) {
		return ListingSync{}, ErrListingTokenExpired
	}
	auth, err := s.accessFromSnapshot(owner, row)
	if err != nil {
		return ListingSync{}, err
	}
	if !sellerID.MatchString(seller) {
		return ListingSync{}, ErrListingResponse
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ListingSync{}, err
	}
	defer func() {
		cleanup, stop := context.WithTimeout(context.Background(), 3*time.Second)
		defer stop()
		_ = tx.Rollback(cleanup)
	}()
	var locked bool
	if err = tx.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock(hashtextextended('meli-listings:' || $1::text, 0))", store).Scan(&locked); err != nil {
		return ListingSync{}, err
	}
	if !locked {
		return ListingSync{}, ErrListingBusy
	}

	var search struct {
		Seller json.Number `json:"seller_id"`
		IDs    []string    `json:"results"`
		Paging struct {
			Total  *int `json:"total"`
			Offset *int `json:"offset"`
			Limit  *int `json:"limit"`
		} `json:"paging"`
	}
	query := url.Values{"offset": {strconv.Itoa(offset)}, "limit": {strconv.Itoa(ListingBatchSize)}}
	_, err = s.listingGET(ctx, auth.AccessToken, "/marketplace/users/"+seller+"/items/search?"+query.Encode(), &search)
	if err != nil {
		return ListingSync{}, err
	}
	p := search.Paging
	if search.Seller.String() != seller || search.IDs == nil || p.Total == nil || *p.Total < 0 ||
		p.Offset == nil || *p.Offset != offset || p.Limit == nil || *p.Limit != ListingBatchSize ||
		len(search.IDs) > ListingBatchSize || (len(search.IDs) == 0 && offset < *p.Total) {
		return ListingSync{}, ErrListingResponse
	}

	var result ListingSync
	details := make(map[string]json.RawMessage, len(search.IDs))
	for index, id := range search.IDs {
		select {
		case <-ctx.Done():
			return ListingSync{}, ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
		if !listingID.MatchString(id) {
			result.Failed++
			continue
		}
		var item listingDetail
		status, fetchErr := s.listingGET(ctx, auth.AccessToken, "/marketplace/items/"+id+"?include_attributes=all", &item)
		if fetchErr != nil || item.ID != id || item.Site != scope.Site || item.Seller.String() != seller ||
			(scope.Logistic != "" && item.Logistic != "" && item.Logistic != scope.Logistic) {
			result.Failed++
			var remote ListingPlatformError
			if errors.As(fetchErr, &remote) {
				result.PlatformStatus = int(remote)
				if remote == 401 || remote == 403 || remote == 429 || remote >= 500 {
					result.Failed += len(search.IDs) - index - 1
					break
				}
			}
			continue
		}
		item.Partial = status == http.StatusPartialContent
		if item.Partial {
			result.Partial++
		}
		detail, err := json.Marshal(item)
		if err != nil {
			return ListingSync{}, err
		}
		details[id] = detail
	}

	// Disconnection, reauthorization or a changed child mapping during the fetch invalidates the batch.
	var current string
	err = tx.QueryRow(ctx, `SELECT id::text FROM app.meli_store_connections
      WHERE id=$1::uuid AND owner_user_id=$2::uuid AND status='connected'
        AND access_token_ciphertext=$3 AND marketplaces=$4::jsonb AND marketplaces_status=$5
      FOR SHARE`, store, owner, row.accessCipher, row.marketsJSON, row.marketsStatus).Scan(&current)
	if errors.Is(err, pgx.ErrNoRows) {
		return ListingSync{}, ErrStoreNotFound
	}
	if err != nil {
		return ListingSync{}, err
	}
	for id, detail := range details {
		_, err = tx.Exec(ctx, `INSERT INTO app.meli_published_listings AS item
          (store_id,site_id,logistic_type,listing_id,details) VALUES ($1::uuid,$2,$3,$4,$5)
          ON CONFLICT (store_id,site_id,logistic_type,listing_id)
          DO UPDATE SET details=CASE WHEN EXCLUDED.details->>'partial'='true'
            THEN item.details || EXCLUDED.details ELSE EXCLUDED.details END`, store, scope.Site, scope.Logistic, id, detail)
		if err != nil {
			return ListingSync{}, err
		}
	}
	if len(details) > 0 || len(search.IDs) == 0 {
		_, err = tx.Exec(ctx, `INSERT INTO app.meli_listing_syncs (store_id,site_id,logistic_type,synced_at)
          VALUES ($1::uuid,$2,$3,now()) ON CONFLICT (store_id,site_id,logistic_type)
          DO UPDATE SET synced_at=EXCLUDED.synced_at`, store, scope.Site, scope.Logistic)
		if err != nil {
			return ListingSync{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return ListingSync{}, err
	}
	result.Saved = len(details)
	next := offset + len(search.IDs)
	if result.Failed != 0 {
		result.NextOffset = &offset
	} else if len(search.IDs) > 0 && next < *p.Total {
		if next < 1000 {
			result.NextOffset = &next
		} else {
			result.ScanRequired = true
		}
	}
	return result, nil
}

type listingDetail struct {
	Partial        bool            `json:"partial"`
	SitelessID     *string         `json:"siteless_user_product_id,omitempty"`
	SitelessFamily *json.Number    `json:"siteless_family_id,omitempty"`
	ID             string          `json:"id"`
	Site           string          `json:"site_id"`
	Seller         json.Number     `json:"seller_id"`
	Logistic       string          `json:"user_logistic_type,omitempty"`
	Global         string          `json:"cbt_item_id,omitempty"`
	Parent         string          `json:"parent_item_id,omitempty"`
	UserProduct    *string         `json:"user_product_id,omitempty"`
	Family         *json.Number    `json:"family_id,omitempty"`
	FamilyName     *string         `json:"family_name,omitempty"`
	Title          *string         `json:"title,omitempty"`
	Category       *string         `json:"category_id,omitempty"`
	Price          json.RawMessage `json:"price,omitempty"`
	Currency       json.RawMessage `json:"currency_id,omitempty"`
	Stock          json.RawMessage `json:"available_quantity,omitempty"`
	Status         *string         `json:"status,omitempty"`
	LastUpdated    *string         `json:"last_updated,omitempty"`
	Thumbnail      *string         `json:"secure_thumbnail,omitempty"`
	SKU            json.RawMessage `json:"seller_custom_field,omitempty"`
	Pictures       json.RawMessage `json:"pictures,omitempty"`
	Attributes     json.RawMessage `json:"attributes,omitempty"`
	Variations     json.RawMessage `json:"variations,omitempty"`
	Markets        json.RawMessage `json:"marketplace_items,omitempty"`
}

func (s *Service) listingGET(ctx context.Context, token, path string, target any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.apiURL+path, nil)
	if err != nil {
		return 0, ErrListingResponse
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return 0, ErrListingResponse
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return resp.StatusCode, ListingPlatformError(resp.StatusCode)
	}
	const maxBody = 2 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(data) > maxBody || json.Unmarshal(data, target) != nil {
		return resp.StatusCode, ErrListingResponse
	}
	return resp.StatusCode, nil
}
