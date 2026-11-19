package meli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const ListingBatchSize = 5

var (
	ErrListingAuthorizationChanged = errors.New("listing authorization snapshot changed")
	ErrListingCursor               = errors.New("listing cursor changed or completed")
	ErrListingIdentity             = errors.New("marketplace listing identity mismatch")
	ErrListingTransport            = errors.New("marketplace listing transport failed")
	ErrListingBusy                 = errors.New("listing sync is already running")
	ErrListingTokenExpired         = errors.New("store Access Token expired")
	ErrListingResponse             = errors.New("unexpected marketplace listing response")
	listingID                      = regexp.MustCompile(`^[A-Z]{3}[0-9]{1,30}$`)
	sellerID                       = regexp.MustCompile(`^[0-9]{1,30}$`)
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
	Items        []json.RawMessage `json:"items"`
	HasMore      bool              `json:"has_more"`
	SyncedAt     *time.Time        `json:"synced_at"`
	NextOffset   *int              `json:"next_offset"`
	ScanRequired bool              `json:"scan_required"`
}

type ListingSync struct {
	Saved        int                  `json:"saved"`
	Processed    int                  `json:"processed"`
	Skipped      int                  `json:"skipped"`
	NeedsReview  int                  `json:"needs_review"`
	Partial      int                  `json:"partial"`
	NextOffset   *int                 `json:"next_offset"`
	ScanRequired bool                 `json:"scan_required"`
	Interruption *ListingInterruption `json:"interruption,omitempty"`
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
	rows, err := s.db.Query(ctx, `SELECT sync.synced_at, CASE WHEN sync.store_id IS NULL THEN 0 ELSE sync.next_offset END, COALESCE(sync.scan_required,false), item.details
      FROM app.meli_store_connections c
      LEFT JOIN app.meli_listing_syncs sync ON (sync.store_id,sync.site_id,sync.logistic_type,sync.seller_id)=(c.id,$3,$4,$9)
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
		if err = rows.Scan(&page.SyncedAt, &page.NextOffset, &page.ScanRequired, &detail); err != nil {
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
		return ListingPage{}, ErrListingAuthorizationChanged
	}
	if len(page.Items) > ListingPageSize {
		page.HasMore = true
		page.Items = page.Items[:ListingPageSize]
	}
	return page, nil
}

func (s *Service) SyncPublished(ctx context.Context, owner, store string, scope ListingScope, expected *int, restart bool) (ListingSync, error) {
	// Leave room in the proxy's 20-second budget for validation and the final commit.
	return s.syncPublished(ctx, owner, store, scope, expected, restart, 10*time.Second)
}

func (s *Service) syncPublished(ctx context.Context, owner, store string, scope ListingScope, expected *int, restart bool, budget time.Duration) (ListingSync, error) {
	readCtx, endRead := context.WithTimeout(ctx, 2*time.Second)
	defer endRead()
	row, seller, err := s.listingScope(readCtx, owner, store, scope)
	if err != nil {
		return ListingSync{}, err
	}
	// Use the current token without starting the Issue #21 refresh flow.
	if row.expires == nil || !row.expires.After(time.Now().Add(budget)) {
		return ListingSync{}, ErrListingTokenExpired
	}
	auth, err := s.accessFromSnapshot(owner, row)
	if err != nil {
		return ListingSync{}, err
	}
	if !sellerID.MatchString(seller) {
		return ListingSync{}, ErrListingIdentity
	}
	key := strings.ToLower(store)
	if _, busy := s.listingRuns.LoadOrStore(key, struct{}{}); busy {
		return ListingSync{}, ErrListingBusy
	}
	defer s.listingRuns.Delete(key)
	cursor, err := readListingCursor(readCtx, s.db, store, scope, seller)
	endRead()
	if err != nil {
		return ListingSync{}, err
	}
	offset := cursor
	if restart {
		offset = 0
	}
	if offset < 0 || (expected != nil && *expected != offset) {
		return ListingSync{}, ErrListingCursor
	}
	fetchCtx, endFetch := context.WithTimeout(ctx, budget)
	result, details, err := s.fetchListingBatch(fetchCtx, auth.AccessToken, seller, scope, offset)
	endFetch()
	if ctx.Err() != nil {
		return ListingSync{}, ctx.Err()
	}
	if err != nil {
		interruption := listingInterruption(err)
		if !restart {
			return ListingSync{}, interruption
		}
		result = ListingSync{NextOffset: &offset, Interruption: interruption}
	}
	// The platform budget can expire while the original request is still live.
	writeCtx, endWrite := context.WithTimeout(ctx, 3*time.Second)
	defer endWrite()
	if err = s.saveListingBatch(writeCtx, owner, store, scope, seller, row, cursor, result, details); err != nil {
		return ListingSync{}, err
	}
	if result.Saved == 0 && result.Interruption != nil {
		return result, result.Interruption
	}
	return result, nil
}

func readListingCursor(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, store string, scope ListingScope, seller string) (int, error) {
	var cursor int
	err := db.QueryRow(ctx, `SELECT COALESCE((
        SELECT COALESCE(next_offset,-1) FROM app.meli_listing_syncs
        WHERE (store_id,site_id,logistic_type,seller_id)=($1::uuid,$2,$3,$4)),0)`,
		store, scope.Site, scope.Logistic, seller).Scan(&cursor)
	return cursor, err
}

type listingWrite struct {
	detail  json.RawMessage
	partial bool
}

func (s *Service) fetchListingBatch(ctx context.Context, token, seller string, scope ListingScope, offset int) (ListingSync, map[string]listingWrite, error) {
	var search struct {
		Seller json.Number `json:"seller_id"`
		IDs    []string    `json:"results"`
		Paging struct {
			Total  *int `json:"total"`
			Offset *int `json:"offset"`
			Limit  *int `json:"limit"`
		} `json:"paging"`
	}
	limit := min(ListingBatchSize, 1000-offset)
	query := url.Values{"offset": {strconv.Itoa(offset)}, "limit": {strconv.Itoa(limit)}}
	_, err := s.listingGET(ctx, token, "/marketplace/users/"+seller+"/items/search?"+query.Encode(), &search)
	if err != nil {
		return ListingSync{}, nil, err
	}
	if search.Seller.String() != seller {
		return ListingSync{}, nil, ErrListingIdentity
	}
	p := search.Paging
	if search.IDs == nil || p.Total == nil || *p.Total < 0 || p.Offset == nil || *p.Offset != offset ||
		p.Limit == nil || *p.Limit != limit || len(search.IDs) > limit || (len(search.IDs) == 0 && offset < *p.Total) {
		return ListingSync{}, nil, ErrListingResponse
	}

	var result ListingSync
	details := make(map[string]listingWrite, len(search.IDs))
	for _, id := range search.IDs {
		if !listingID.MatchString(id) {
			result.Skipped++
			result.Processed++
			continue
		}
		select {
		case <-ctx.Done():
			result.Interruption = listingInterruption(ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
		if result.Interruption != nil {
			break
		}
		var item listingDetail
		status, fetchErr := s.listingGET(ctx, token, "/marketplace/items/"+id+"?include_attributes=all", &item)
		if errors.Is(fetchErr, ListingPlatformError(http.StatusNotFound)) {
			result.Skipped++
			result.Processed++
			continue
		}
		if fetchErr == nil && (item.ID != id || item.Site != scope.Site || item.Seller.String() != seller ||
			(scope.Logistic != "" && item.Logistic != "" && item.Logistic != scope.Logistic)) {
			fetchErr = ErrListingIdentity
		}
		if fetchErr != nil {
			result.Interruption = listingInterruption(fetchErr)
			if result.Interruption.Code == "listing_identity_mismatch" || result.Interruption.Code == "listing_invalid_response" {
				result.NeedsReview++
			}
			break
		}
		item.Partial = status == http.StatusPartialContent
		detail, err := json.Marshal(item)
		if err != nil {
			return ListingSync{}, nil, err
		}
		details[id] = listingWrite{detail: detail, partial: item.Partial}
		result.Processed++
	}
	result.Saved = len(details)
	for _, detail := range details {
		if detail.partial {
			result.Partial++
		}
	}
	next := offset + result.Processed
	if result.Interruption != nil || next < *p.Total {
		if next < 1000 {
			result.NextOffset = &next
		} else {
			result.ScanRequired = true
		}
	}
	return result, details, nil
}

func (s *Service) saveListingBatch(ctx context.Context, owner, store string, scope ListingScope, seller string, row tokenSnapshot, cursor int, result ListingSync, details map[string]listingWrite) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // pgx discards the connection if cancellation prevents rollback.

	var valid bool
	err = tx.QueryRow(ctx, `SELECT COALESCE(status='connected' AND seller_id=$3
        AND access_token_ciphertext=$4 AND marketplaces=$5::jsonb AND marketplaces_status=$6
        AND access_token_expires_at > clock_timestamp(),false)
        FROM app.meli_store_connections
        WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`,
		store, owner, row.sellerID, row.accessCipher, row.marketsJSON, row.marketsStatus).Scan(&valid)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrStoreNotFound
	}
	if err != nil {
		return err
	}
	if !valid {
		return ErrListingAuthorizationChanged
	}
	current, err := readListingCursor(ctx, tx, store, scope, seller)
	if err != nil {
		return err
	}
	if current != cursor {
		return ErrListingCursor
	}
	for id, item := range details {
		detail := item.detail
		if item.partial {
			var previous json.RawMessage
			err = tx.QueryRow(ctx, `SELECT details FROM app.meli_published_listings
              WHERE (store_id,site_id,logistic_type,listing_id)=($1::uuid,$2,$3,$4)
                AND details->>'seller_id'=$5`, store, scope.Site, scope.Logistic, id, seller).Scan(&previous)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			if err == nil {
				detail, err = mergePartialListing(previous, detail)
				if err != nil {
					return err
				}
			}
		}
		_, err = tx.Exec(ctx, `INSERT INTO app.meli_published_listings
            (store_id,site_id,logistic_type,listing_id,details) VALUES ($1::uuid,$2,$3,$4,$5)
            ON CONFLICT (store_id,site_id,logistic_type,listing_id)
            DO UPDATE SET details=EXCLUDED.details`, store, scope.Site, scope.Logistic, id, detail)
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.meli_listing_syncs AS state
      (store_id,site_id,logistic_type,seller_id,synced_at,next_offset,scan_required)
      VALUES ($1::uuid,$2,$3,$4,CASE WHEN $5 THEN now() END,$6,$7)
      ON CONFLICT (store_id,site_id,logistic_type) DO UPDATE SET
        seller_id=EXCLUDED.seller_id, next_offset=EXCLUDED.next_offset, scan_required=EXCLUDED.scan_required,
        synced_at=COALESCE(EXCLUDED.synced_at, CASE WHEN state.seller_id=EXCLUDED.seller_id THEN state.synced_at END)`,
		store, scope.Site, scope.Logistic, seller, result.Saved > 0 || result.Interruption == nil, result.NextOffset, result.ScanRequired)
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type ListingInterruption struct {
	Code           string `json:"code"`
	PlatformStatus int    `json:"platform_status,omitempty"`
	Retryable      bool   `json:"retryable"`
}

func (e *ListingInterruption) Error() string { return e.Code }

func listingInterruption(err error) *ListingInterruption {
	var remote ListingPlatformError
	switch {
	case errors.Is(err, ErrListingIdentity):
		return &ListingInterruption{Code: "listing_identity_mismatch"}
	case errors.Is(err, context.DeadlineExceeded):
		return &ListingInterruption{Code: "listing_timeout", Retryable: true}
	case errors.Is(err, ErrListingTransport):
		return &ListingInterruption{Code: "listing_transport", Retryable: true}
	case errors.As(err, &remote):
		e := &ListingInterruption{Code: "listing_platform_error", PlatformStatus: int(remote)}
		switch {
		case remote == 401 || remote == 403:
			e.Code = "listing_platform_auth"
		case remote == 429:
			e.Code, e.Retryable = "listing_rate_limited", true
		case remote >= 500:
			e.Retryable = true
		}
		return e
	default:
		return &ListingInterruption{Code: "listing_invalid_response"}
	}
}

func mergePartialListing(previous, incoming json.RawMessage) (json.RawMessage, error) {
	var old, patch map[string]any
	for _, doc := range []struct {
		raw    json.RawMessage
		fields *map[string]any
	}{{previous, &old}, {incoming, &patch}} {
		decoder := json.NewDecoder(bytes.NewReader(doc.raw))
		decoder.UseNumber()
		if err := decoder.Decode(doc.fields); err != nil {
			return nil, err
		}
	}
	return json.Marshal(mergeListingFields(old, patch))
}

func mergeListingFields(old, patch map[string]any) map[string]any {
	for key, value := range patch {
		if value == nil {
			continue
		}
		switch next := value.(type) {
		case map[string]any:
			if previous, ok := old[key].(map[string]any); ok {
				value = mergeListingFields(previous, next)
			}
		case []any:
			// Empty arrays are explicit. Nonempty collections merge by ID.
			switch key {
			case "pictures", "attributes", "variations", "attribute_combinations":
				if previous, ok := old[key].([]any); ok && len(next) > 0 {
					value = mergeListingEntries(previous, next)
				}
			}
		}
		old[key] = value
	}
	return old
}

func mergeListingEntries(previous, next []any) []any {
	indexes := make(map[string]int, len(previous))
	for i, entry := range previous {
		if id, ok := listingEntryID(entry); ok {
			indexes[id] = i
		}
	}
	for _, entry := range next {
		id, ok := listingEntryID(entry)
		if !ok {
			// A partial entry without a stable key cannot safely replace another.
			continue
		}
		if index, exists := indexes[id]; exists {
			previous[index] = mergeListingFields(previous[index].(map[string]any), entry.(map[string]any))
		} else {
			indexes[id] = len(previous)
			previous = append(previous, entry)
		}
	}
	return previous
}

func listingEntryID(entry any) (string, bool) {
	fields, ok := entry.(map[string]any)
	if !ok {
		return "", false
	}
	switch id := fields["id"].(type) {
	case string:
		return id, id != ""
	case json.Number:
		return id.String(), id != ""
	default:
		return "", false
	}
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
		return 0, listingReadError(ctx, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return resp.StatusCode, ListingPlatformError(resp.StatusCode)
	}
	const maxBody = 2 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return resp.StatusCode, listingReadError(ctx, err)
	}
	if len(data) > maxBody || json.Unmarshal(data, target) != nil {
		return resp.StatusCode, ErrListingResponse
	}
	return resp.StatusCode, nil
}

func listingReadError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return context.DeadlineExceeded
	}
	return ErrListingTransport
}
