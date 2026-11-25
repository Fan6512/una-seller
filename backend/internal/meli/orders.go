package meli

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const OrderBatchSize = 3
const orderPageSize = 20
const orderDetailBatchSize = 3
const orderRequestInterval = 100 * time.Millisecond

// Official /marketplace/orders/search quota: 100 requests/minute.
// Use 650ms between searches for headroom; details keep the 100ms pacing.
const orderSearchRequestInterval = 650 * time.Millisecond
const orderDefaultCooldown = 5 * time.Second

var (
	ErrOrderBusy                 = errors.New("order sync already running")
	ErrOrderInvalidResponse      = errors.New("invalid order response")
	ErrOrderAuthorizationChanged = errors.New("order authorization changed")
	ErrOrderReviewConflict       = errors.New("order review no longer matches")
	ErrOrderResetConflict        = errors.New("order reset no longer matches")
	ErrOrderResetInvalid         = errors.New("invalid order reset request")
	ErrOrderSkipInvalid          = errors.New("invalid order skip request")
	ErrOrderNotFound             = errors.New("order not found")
	orderSitePattern             = regexp.MustCompile(`^[A-Z]{3,8}$`)
	orderNumberPattern           = regexp.MustCompile(`^[0-9]{1,30}$`)
	orderAmountPattern           = regexp.MustCompile(`^[0-9]{1,16}(\.[0-9]{1,8})?$`)
)

type OrderScope struct {
	Site     string
	Logistic string
}

type OrderLine struct {
	ElementID   string `json:"element_id,omitempty"`
	ItemID      string `json:"item_id"`
	Title       string `json:"title"`
	VariationID string `json:"variation_id,omitempty"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	CurrencyID  string `json:"currency_id"`
	ImageURL    string `json:"image_url,omitempty"`
}

type OrderRecord struct {
	ID                       string         `json:"order_id"`
	PackID                   string         `json:"pack_id,omitempty"`
	Status                   string         `json:"status"`
	CurrencyID               string         `json:"currency_id"`
	TotalAmount              string         `json:"total_amount"` // Raw upstream field; its currency label can disagree with its value.
	SalesAmountUSD           string         `json:"sales_amount_usd,omitempty"`
	OrderItemsReviewRequired bool           `json:"order_items_review_required,omitempty"`
	DateCreated              time.Time      `json:"date_created"`
	LastUpdated              time.Time      `json:"last_updated"`
	ShipmentID               string         `json:"shipment_id,omitempty"`
	Shipment                 *OrderShipment `json:"shipment,omitempty"`
	Items                    []OrderLine    `json:"order_items"`
	PurchaseOrderID          string         `json:"purchase_order_id,omitempty"`
	PurchaseCostCNY          *string        `json:"purchase_cost_cny,omitempty"`
	PurchaseTrackingNumber   string         `json:"purchase_tracking_number,omitempty"`
	DispatchDeadline         *time.Time     `json:"dispatch_deadline,omitempty"`
	DispatchSLAStatus        string         `json:"dispatch_sla_status,omitempty"`
	DispatchCheckedAt        *time.Time     `json:"dispatch_checked_at,omitempty"`
}

type OrderPage struct {
	Items          []OrderRecord `json:"items"`
	HasMore        bool          `json:"has_more"`
	PurchaseFilter string        `json:"purchase_filter"`
}

type OrderSyncResult struct {
	Saved      int  `json:"saved"`     // number of order details stored
	Processed  int  `json:"processed"` // number of search envelopes completed
	HasMore    bool `json:"has_more"`
	NextOffset int  `json:"next_offset"`
}

// Read-only persisted sync snapshot for a verified seller scope.
type OrderSyncState struct {
	NextOffset int          `json:"next_offset"`
	CanReset   bool         `json:"can_reset"`
	Review     *OrderReview `json:"review,omitempty"`
}

type OrderReview struct {
	OrderID  string `json:"order_id"`
	ReviewID string `json:"review_id"`
	Offset   int    `json:"offset"`
	Reason   string `json:"reason"`
}

func (r *OrderReview) Error() string { return "order_needs_review" }

type OrderRateLimitError struct {
	until  time.Time
	shared bool
}

func (e *OrderRateLimitError) Error() string { return "order rate limited" }
func (e *OrderRateLimitError) Unwrap() error { return ListingPlatformError(http.StatusTooManyRequests) }
func (e *OrderRateLimitError) RetryAfter() string {
	delay := time.Until(e.until)
	if delay <= 0 {
		return "0"
	}
	return strconv.FormatInt(int64((delay-1)/time.Second)+1, 10)
}

// Use a fixed diagnostic label, never a URL, query, store/order ID or token.
func orderReadResource(path string) string {
	switch {
	case strings.HasPrefix(path, "/billing/integration/group/ML/order/details?"):
		return "billing"
	case strings.HasPrefix(path, "/marketplace/shipments/"):
		return "shipment"
	case path == "/marketplace/orders/search" || strings.HasPrefix(path, "/marketplace/orders/search?"):
		return "order_search"
	case strings.HasPrefix(path, "/marketplace/orders/"):
		return "order_detail"
	default:
		return "other"
	}
}

// Pace order reads across scopes in this process and apply upstream cooldowns
// before another request can start, including concurrent syncs from other stores.
func (s *Service) orderGET(ctx context.Context, token, path string, target any) (status int, err error) {
	if s.sharedOrderCoordination {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
	}
	if err := s.orderRequests.acquire(ctx, isBackgroundOrderRead(ctx)); err != nil {
		if errors.Is(err, ErrOrderBusy) {
			return http.StatusConflict, err
		}
		return 0, err
	}
	defer s.orderRequests.Unlock()
	finish, err := s.beginSharedOrderRead(ctx)
	if err != nil {
		if errors.Is(err, ErrOrderBusy) {
			return http.StatusConflict, err
		}
		return 0, err
	}
	if finish != nil {
		defer func() {
			if finishErr := finish(); finishErr != nil {
				status, err = 0, finishErr
			}
		}()
	}
	if time.Now().Before(s.orderRetryUntil) {
		cooldown := &OrderRateLimitError{until: s.orderRetryUntil, shared: s.sharedOrderCoordination}
		if orderReadResource(path) == "billing" {
			slog.Info("order read rate limited", "source", "local_cooldown", "resource", "billing", "retry_after_seconds", cooldown.RetryAfter())
		}
		return http.StatusTooManyRequests, cooldown
	}
	// Both clocks are protected by orderRequests. Search calls have their
	// own slower quota; detail reads must not inherit the search delay.
	isSearch := path == "/marketplace/orders/search" || strings.HasPrefix(path, "/marketplace/orders/search?")
	next := s.orderNextRequest
	if isSearch && s.orderNextSearch.After(next) {
		next = s.orderNextSearch
	}
	if delay := time.Until(next); delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-timer.C:
		}
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	now := time.Now()
	s.orderNextRequest = now.Add(orderRequestInterval)
	if isSearch {
		s.orderNextSearch = now.Add(orderSearchRequestInterval)
	}
	var response listingGETTiming
	status, err = s.listingGET(ctx, token, path, target, &response)
	if status == http.StatusTooManyRequests {
		now := time.Now()
		until, dateErr := http.ParseTime(response.retryAfter)
		if dateErr != nil {
			delay := orderDefaultCooldown
			if seconds, parseErr := strconv.ParseUint(response.retryAfter, 10, 64); parseErr == nil {
				delay = time.Duration(min(seconds, math.MaxInt64/uint64(time.Second))) * time.Second
			}
			until = now.Add(delay)
		}
		if s.sharedOrderCoordination {
			until = laterOrderTime(until, now.Add(time.Minute))
		}
		s.orderRetryUntil = until
		cooldown := &OrderRateLimitError{until: until, shared: s.sharedOrderCoordination}
		slog.Warn("order read rate limited", "source", "upstream", "resource", orderReadResource(path),
			"upstream_http_status", response.httpStatus, "retry_after_seconds", cooldown.RetryAfter())
		return status, cooldown
	}
	return status, err
}

type orderCursor struct {
	LeaseID      string
	From         time.Time
	To           time.Time
	Offset       int
	PendingIDs   []string
	PendingCount int
	PendingMore  bool
	ReviewID     string
	ReviewReason string
}

func orderScopeValid(scope OrderScope) bool {
	return scope.Site != "CBT" && orderSitePattern.MatchString(scope.Site) && len(scope.Logistic) <= 40 &&
		regexp.MustCompile(`^[a-z_]+$`).MatchString(scope.Logistic)
}

func orderNumber(id string) bool {
	return orderNumberPattern.MatchString(id)
}

// Search returns pack envelopes; ONLY results[].orders[].id is an order ID.
func parseOrderSearch(data []byte, seller string, offset int) ([]string, int, int, error) {
	var response struct {
		Results []struct {
			Orders []struct {
				ID     json.Number `json:"id"`
				Seller struct {
					ID json.Number `json:"id"`
				} `json:"seller"`
			} `json:"orders"`
		} `json:"results"`
		Paging struct {
			Total *int `json:"total"`
		} `json:"paging"`
	}
	if err := json.Unmarshal(data, &response); err != nil || response.Results == nil ||
		response.Paging.Total == nil || *response.Paging.Total < 0 ||
		offset > *response.Paging.Total || len(response.Results) > OrderBatchSize ||
		(len(response.Results) == 0 && offset < *response.Paging.Total) {
		return nil, 0, 0, ErrOrderInvalidResponse
	}
	ids, seen := make([]string, 0, len(response.Results)), make(map[string]bool)
	for _, envelope := range response.Results {
		if len(envelope.Orders) == 0 || len(envelope.Orders) > 40 {
			return nil, 0, 0, ErrOrderInvalidResponse
		}
		for _, order := range envelope.Orders {
			id := order.ID.String()
			if !orderNumber(id) || (order.Seller.ID != "" && order.Seller.ID.String() != seller) {
				return nil, 0, 0, ErrOrderInvalidResponse
			}
			if !seen[id] {
				seen[id] = true
				ids = append(ids, id)
			}
		}
	}
	return ids, *response.Paging.Total, len(response.Results), nil
}

// Accept only explicit business fields; never store addresses, phone numbers,
// buyers, payment instruments, or unfiltered platform JSON.
func parseOrderDetail(data []byte, expectedID, seller, site string) (OrderRecord, error) {
	var payload struct {
		ID      json.Number     `json:"id"`
		PackID  json.Number     `json:"pack_id"`
		Status  string          `json:"status"`
		SiteID  json.RawMessage `json:"site_id"`
		Context struct {
			Site json.RawMessage `json:"site"`
		} `json:"context"`
		CurrencyID  string      `json:"currency_id"`
		TotalAmount json.Number `json:"total_amount"`
		DateCreated string      `json:"date_created"`
		LastUpdated string      `json:"last_updated"`
		Seller      struct {
			ID json.Number `json:"id"`
		} `json:"seller"`
		Shipping struct {
			ID json.Number `json:"id"`
		} `json:"shipping"`
		OrderItems []struct {
			Item struct {
				ID          string      `json:"id"`
				Title       string      `json:"title"`
				VariationID json.Number `json:"variation_id"`
			} `json:"item"`
			ElementID  json.Number `json:"element_id"`
			Quantity   int         `json:"quantity"`
			UnitPrice  json.Number `json:"unit_price"`
			CurrencyID string      `json:"currency_id"`
		} `json:"order_items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil ||
		payload.ID.String() != expectedID || !orderNumber(expectedID) ||
		payload.Seller.ID.String() != seller ||
		payload.Status == "" || len(payload.Status) > 80 ||
		payload.CurrencyID == "" || len(payload.CurrencyID) > 12 ||
		!orderAmountPattern.MatchString(payload.TotalAmount.String()) {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	matchedSite := false
	for _, field := range []json.RawMessage{payload.SiteID, payload.Context.Site} {
		if field == nil || string(field) == "null" {
			continue
		}
		var reportedSite string
		if json.Unmarshal(field, &reportedSite) != nil || reportedSite == "" || reportedSite != site {
			return OrderRecord{}, ErrOrderInvalidResponse
		}
		matchedSite = true
	}
	if !matchedSite {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	created, err := time.Parse(time.RFC3339Nano, payload.DateCreated)
	if err != nil {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	updated, err := time.Parse(time.RFC3339Nano, payload.LastUpdated)
	if err != nil {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	record := OrderRecord{
		ID: expectedID, Status: payload.Status, CurrencyID: payload.CurrencyID,
		TotalAmount: payload.TotalAmount.String(), DateCreated: created.UTC(),
		LastUpdated: updated.UTC(), Items: make([]OrderLine, 0, len(payload.OrderItems)),
	}
	if payload.PackID != "" {
		if !orderNumber(payload.PackID.String()) {
			return OrderRecord{}, ErrOrderInvalidResponse
		}
		record.PackID = payload.PackID.String()
	}
	if payload.Shipping.ID != "" {
		if !orderNumber(payload.Shipping.ID.String()) {
			return OrderRecord{}, ErrOrderInvalidResponse
		}
		record.ShipmentID = payload.Shipping.ID.String()
	}
	if len(payload.OrderItems) > 100 {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	seenElements := make(map[string]OrderLine)
	for _, item := range payload.OrderItems {
		if item.Quantity <= 0 || item.Quantity > 100000 ||
			!orderAmountPattern.MatchString(item.UnitPrice.String()) ||
			item.Item.ID == "" || len(item.Item.ID) > 100 {
			return OrderRecord{}, ErrOrderInvalidResponse
		}
		line := OrderLine{
			ItemID: item.Item.ID, Title: item.Item.Title, Quantity: item.Quantity,
			UnitPrice: item.UnitPrice.String(), CurrencyID: item.CurrencyID,
		}
		if item.Item.VariationID != "" {
			if !orderNumber(item.Item.VariationID.String()) {
				return OrderRecord{}, ErrOrderInvalidResponse
			}
			line.VariationID = item.Item.VariationID.String()
		}
		if elementID := item.ElementID.String(); elementID != "" {
			if !orderNumber(elementID) {
				return OrderRecord{}, ErrOrderInvalidResponse
			}
			line.ElementID = elementID
			if previous, found := seenElements[elementID]; found {
				if previous != line {
					return OrderRecord{}, ErrOrderInvalidResponse
				}
				continue
			}
			seenElements[elementID] = line
		}
		record.Items = append(record.Items, line)
	}
	return record, nil
}

// Legacy snapshots lack element IDs. Equal commercial lines are only candidates
// for duplication, not proof: never delete or silently merge them. A fresh
// official order read may establish distinct element IDs or deduplicate them.
func orderItemsNeedReview(items []OrderLine) bool {
	type signature struct {
		item, variation, currency, price string
		quantity                         int
	}
	seen := make(map[signature]bool, len(items))
	for _, line := range items {
		price := line.UnitPrice
		if orderAmountPattern.MatchString(price) {
			if value, ok := new(big.Rat).SetString(price); ok {
				price = value.RatString()
			}
		}
		key := signature{line.ItemID, line.VariationID, line.CurrencyID, price, line.Quantity}
		missing := line.ElementID == ""
		if previousMissing, exists := seen[key]; exists && (previousMissing || missing) {
			return true
		}
		seen[key] = missing
	}
	return false
}

// Marketplace Orders can report a local total under a USD header. Only sum
// explicit USD merchandise lines, per Order, excluding shipping and payments.
func orderSalesAmountUSD(items []OrderLine) string {
	if len(items) == 0 || orderItemsNeedReview(items) {
		return ""
	}
	total := new(big.Rat)
	for _, line := range items {
		if line.CurrencyID != "USD" || line.Quantity < 1 || !orderAmountPattern.MatchString(line.UnitPrice) {
			return ""
		}
		price, _ := new(big.Rat).SetString(line.UnitPrice)
		total.Add(total, price.Mul(price, big.NewRat(int64(line.Quantity), 1)))
	}
	amount := strings.TrimRight(total.FloatString(8), "0")
	if strings.HasSuffix(amount, ".") {
		return amount + "00"
	}
	if len(amount)-strings.IndexByte(amount, '.') == 2 {
		return amount + "0"
	}
	return amount
}

// Both order ingestion paths can replace the line snapshots. Keep a cached
// product image for any listing still present without restoring old line data.
func orderItemsWithSavedImages(incoming, stored string) string {
	return `(SELECT COALESCE(jsonb_agg(
	 CASE WHEN saved.url IS NOT NULL THEN jsonb_set(line.value,'{image_url}',to_jsonb(saved.url),true)
	      ELSE line.value END ORDER BY line.pos),'[]'::jsonb)
	 FROM jsonb_array_elements(` + incoming + `) WITH ORDINALITY AS line(value,pos)
	 LEFT JOIN LATERAL (
	   SELECT old.value->>'image_url' AS url FROM jsonb_array_elements(` + stored + `) AS old(value)
	   WHERE old.value->>'item_id'=line.value->>'item_id' AND COALESCE(old.value->>'image_url','')<>''
	   LIMIT 1
	 ) saved ON true)`
}

const orderConnection = `SELECT c.id,c.status,m.value->>'user_id' AS child_id
 FROM app.meli_store_connections c
 LEFT JOIN LATERAL jsonb_array_elements(c.marketplaces) m(value)
   ON c.marketplaces_status='verified' AND m.value->>'site_id'=$3 AND m.value->>'logistic_type'=$4
 WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid`

const orderSyncScope = ` FROM connection c
 JOIN app.meli_order_syncs sync ON sync.store_id=c.id AND sync.seller_id=c.child_id
 WHERE c.status<>'disconnected' AND sync.site_id=$3 AND sync.logistic_type=$4`

func (s *Service) currentOrderReview(ctx context.Context, owner, store string, scope OrderScope) error {
	var review OrderReview
	err := s.db.QueryRow(ctx, `WITH connection AS (`+orderConnection+`)
 SELECT sync.pending_order_ids[1],sync.review_id::text,sync.next_offset,sync.review_reason`+orderSyncScope+`
 AND sync.review_id IS NOT NULL`, store, owner, scope.Site, scope.Logistic).
		Scan(&review.OrderID, &review.ReviewID, &review.Offset, &review.Reason)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	return &review
}

func (s *Service) acquireOrderCursor(ctx context.Context, owner, store string, scope OrderScope, seller string) (orderCursor, error) {
	var cursor orderCursor
	err := s.db.QueryRow(ctx, `WITH connection AS (`+orderConnection+` FOR SHARE OF c)
INSERT INTO app.meli_order_syncs AS sync
 (store_id,site_id,logistic_type,seller_id,window_from,window_to,lease_id,lease_until)
SELECT id,$3,$4,$5,now()-interval '30 days',now(),gen_random_uuid(),now()+interval '45 seconds'
 FROM connection WHERE status='connected' AND child_id=$5
ON CONFLICT (store_id,site_id,logistic_type) DO UPDATE
 SET lease_id=EXCLUDED.lease_id,lease_until=EXCLUDED.lease_until,seller_id=EXCLUDED.seller_id,
     window_from=CASE WHEN sync.seller_id<>EXCLUDED.seller_id THEN EXCLUDED.window_from
       ELSE COALESCE(sync.window_from,sync.last_completed_to - interval '10 minutes',EXCLUDED.window_from) END,
     window_to=CASE WHEN sync.seller_id<>EXCLUDED.seller_id THEN EXCLUDED.window_to
       ELSE COALESCE(sync.window_to,EXCLUDED.window_to) END,
     next_offset=CASE WHEN sync.seller_id<>EXCLUDED.seller_id THEN 0 ELSE sync.next_offset END,
     last_completed_to=CASE WHEN sync.seller_id=EXCLUDED.seller_id THEN sync.last_completed_to END,
     last_synced_at=CASE WHEN sync.seller_id=EXCLUDED.seller_id THEN sync.last_synced_at END,
     pending_order_ids=CASE WHEN sync.seller_id<>EXCLUDED.seller_id THEN '{}' ELSE sync.pending_order_ids END,
     pending_page_count=CASE WHEN sync.seller_id<>EXCLUDED.seller_id THEN 0 ELSE sync.pending_page_count END,
     pending_has_more=sync.seller_id=EXCLUDED.seller_id AND sync.pending_has_more,
     review_id=CASE WHEN sync.seller_id=EXCLUDED.seller_id THEN sync.review_id END,
     review_reason=CASE WHEN sync.seller_id=EXCLUDED.seller_id THEN sync.review_reason END
 WHERE sync.lease_until IS NULL OR sync.lease_until<now()
RETURNING lease_id::text,window_from,window_to,next_offset,pending_order_ids,pending_page_count,pending_has_more,
 COALESCE(review_id::text,''),COALESCE(review_reason,'')`,
		store, owner, scope.Site, scope.Logistic, seller).Scan(&cursor.LeaseID, &cursor.From, &cursor.To, &cursor.Offset,
		&cursor.PendingIDs, &cursor.PendingCount, &cursor.PendingMore, &cursor.ReviewID, &cursor.ReviewReason)
	if errors.Is(err, pgx.ErrNoRows) {
		return orderCursor{}, ErrOrderBusy
	}
	return cursor, err
}

func (s *Service) releaseOrderCursor(store string, scope OrderScope, lease string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _ = s.db.Exec(ctx, `UPDATE app.meli_order_syncs SET lease_id=NULL,lease_until=NULL
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND lease_id=$4::uuid`,
		store, scope.Site, scope.Logistic, lease)
}

// wasOrderSkipped only recognizes a previously confirmed skip for the same seller,
// order and failure reason. Valid details are still fetched and saved normally.
func (s *Service) wasOrderSkipped(ctx context.Context, store string, scope OrderScope, seller, orderID, reason string) (bool, error) {
	var skipped bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(
 SELECT 1 FROM app.meli_order_skips
 WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3
 AND seller_id=$4 AND order_id=$5 AND reason=$6)`,
		store, scope.Site, scope.Logistic, seller, orderID, reason).Scan(&skipped)
	return skipped, err
}

// SyncOrders commits one detail batch and advances the search offset only after the whole page is saved.
func (s *Service) SyncOrders(ctx context.Context, owner, store string, scope OrderScope) (OrderSyncResult, error) {
	if !orderScopeValid(scope) {
		return OrderSyncResult{}, ErrSiteNotAuthorized
	}
	if !validUUID(store) {
		return OrderSyncResult{}, ErrStoreNotFound
	}
	if err := s.currentOrderReview(ctx, owner, store, scope); err != nil {
		return OrderSyncResult{}, err
	}
	auth, childID, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderSyncResult{}, err
	}
	if !orderNumber(childID) {
		return OrderSyncResult{}, ErrOrderInvalidResponse
	}
	cursor, err := s.acquireOrderCursor(ctx, owner, store, scope, childID)
	if err != nil {
		return OrderSyncResult{}, err
	}
	defer s.releaseOrderCursor(store, scope, cursor.LeaseID)
	if cursor.ReviewID != "" {
		return OrderSyncResult{}, &OrderReview{
			OrderID: cursor.PendingIDs[0], ReviewID: cursor.ReviewID, Offset: cursor.Offset, Reason: cursor.ReviewReason,
		}
	}

	fetchCtx, cancelFetch := context.WithTimeout(ctx, 15*time.Second)
	defer cancelFetch()
	ids := cursor.PendingIDs
	if len(ids) == 0 {
		query := url.Values{
			"seller.id":         {childID},
			"site":              {scope.Site},
			"last_updated.from": {cursor.From.UTC().Format(time.RFC3339Nano)},
			"last_updated.to":   {cursor.To.UTC().Format(time.RFC3339Nano)},
			"sort":              {"updated_asc"},
			"limit":             {strconv.Itoa(OrderBatchSize)},
			"offset":            {strconv.Itoa(cursor.Offset)},
		}
		var search json.RawMessage
		_, err = s.orderGET(fetchCtx, auth.AccessToken, "/marketplace/orders/search?"+query.Encode(), &search)
		if err != nil {
			return OrderSyncResult{}, err
		}
		var total int
		ids, total, cursor.PendingCount, err = parseOrderSearch(search, childID, cursor.Offset)
		if err != nil {
			return OrderSyncResult{}, err
		}
		cursor.PendingMore = cursor.Offset+cursor.PendingCount < total
	}
	batchSize := min(orderDetailBatchSize, len(ids))
	records := make([]OrderRecord, 0, batchSize)
	processedDetails := 0
detailLoop:
	for _, id := range ids[:batchSize] {
		var raw json.RawMessage
		status, detailErr := s.orderGET(fetchCtx, auth.AccessToken, "/marketplace/orders/"+id, &raw)
		reason := ""
		switch {
		case status == http.StatusNotFound:
			reason = "detail_not_found"
		case errors.Is(detailErr, ErrListingResponse):
			reason = "invalid_detail"
		case detailErr != nil:
			// Keep an already processed prefix when only the provider budget expires.
			if errors.Is(detailErr, context.DeadlineExceeded) && ctx.Err() == nil &&
				(processedDetails > 0 || len(cursor.PendingIDs) == 0) {
				break detailLoop
			}
			return OrderSyncResult{}, detailErr
		}
		var record OrderRecord
		if reason == "" {
			var parseErr error
			record, parseErr = parseOrderDetail(raw, id, childID, scope.Site)
			if parseErr != nil {
				reason = "invalid_detail"
			}
		}
		if reason != "" {
			// Reuse the operator's decision only if the same order fails again in
			// the same way. Never hide a recovered/valid order.
			skipped, checkErr := s.wasOrderSkipped(ctx, store, scope, childID, id, reason)
			if checkErr != nil {
				return OrderSyncResult{}, checkErr
			}
			if skipped {
				processedDetails++
				continue
			}
			cursor.ReviewReason = reason
			break
		}
		records = append(records, record)
		processedDetails++
	}
	if ctx.Err() != nil {
		return OrderSyncResult{}, ctx.Err()
	}
	cursor.PendingIDs = ids[processedDetails:]
	if cursor.ReviewReason != "" {
		cursor.ReviewID = cursor.LeaseID
	}
	result, err := s.saveOrderBatch(ctx, owner, store, scope, auth.SellerID, childID, cursor, records)
	if err != nil {
		return OrderSyncResult{}, err
	}
	if cursor.ReviewID != "" {
		return result, &OrderReview{
			OrderID: cursor.PendingIDs[0], ReviewID: cursor.ReviewID, Offset: result.NextOffset, Reason: cursor.ReviewReason,
		}
	}
	return result, nil
}

func (s *Service) saveOrderBatch(ctx context.Context, owner, store string, scope OrderScope,
	globalSeller, childSeller string, cursor orderCursor, records []OrderRecord) (OrderSyncResult, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrderSyncResult{}, err
	}
	defer tx.Rollback(ctx)
	var seller, status, marketStatus string
	var rawMarkets []byte
	err = tx.QueryRow(ctx, `SELECT seller_id,status,marketplaces_status,marketplaces FROM app.meli_store_connections
WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR SHARE`, store, owner).
		Scan(&seller, &status, &marketStatus, &rawMarkets)
	if err != nil {
		return OrderSyncResult{}, ErrOrderAuthorizationChanged
	}
	var markets []Marketplace
	if json.Unmarshal(rawMarkets, &markets) != nil || seller != globalSeller ||
		status != "connected" || marketStatus != "verified" {
		return OrderSyncResult{}, ErrOrderAuthorizationChanged
	}
	allowed := false
	for _, m := range markets {
		if m.SiteID == scope.Site && m.LogisticType == scope.Logistic && m.UserID.String() == childSeller {
			allowed = true
		}
	}
	if !allowed {
		return OrderSyncResult{}, ErrOrderAuthorizationChanged
	}
	var offset int
	var from, to time.Time
	err = tx.QueryRow(ctx, `SELECT next_offset,window_from,window_to FROM app.meli_order_syncs
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND seller_id=$4 AND lease_id=$5::uuid AND lease_until>now() FOR UPDATE`,
		store, scope.Site, scope.Logistic, childSeller, cursor.LeaseID).Scan(&offset, &from, &to)
	if err != nil || offset != cursor.Offset || !from.Equal(cursor.From) || !to.Equal(cursor.To) {
		return OrderSyncResult{}, ErrOrderBusy
	}
	for _, rec := range records {
		if err = upsertOrderRecord(ctx, tx, store, scope, childSeller, rec); err != nil {
			return OrderSyncResult{}, err
		}
	}
	result, err := updateOrderProgress(ctx, tx, store, scope, cursor)
	if err != nil {
		return OrderSyncResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OrderSyncResult{}, err
	}
	result.Saved = len(records)
	return result, nil
}

// updateOrderProgress requires the caller to hold the sync row lock.
func updateOrderProgress(ctx context.Context, tx pgx.Tx, store string, scope OrderScope, cursor orderCursor) (OrderSyncResult, error) {
	pageComplete := len(cursor.PendingIDs) == 0
	result := OrderSyncResult{HasMore: !pageComplete || cursor.PendingMore}
	next := cursor.Offset
	if pageComplete {
		next += cursor.PendingCount
		result.Processed = cursor.PendingCount
		cursor.PendingCount = 0
		cursor.PendingMore = false
	}
	err := tx.QueryRow(ctx, `UPDATE app.meli_order_syncs
SET next_offset=CASE WHEN $5 THEN 0 ELSE $4 END,
 last_completed_to=CASE WHEN $5 THEN window_to ELSE last_completed_to END,
 window_from=CASE WHEN NOT $5 THEN window_from END,window_to=CASE WHEN NOT $5 THEN window_to END,
 pending_order_ids=$6,pending_page_count=$7,pending_has_more=$8,
 review_id=NULLIF($9,'')::uuid,review_reason=NULLIF($10,''),
 last_synced_at=now(),lease_id=NULL,lease_until=NULL
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3
RETURNING next_offset`,
		store, scope.Site, scope.Logistic, next, !result.HasMore, cursor.PendingIDs, cursor.PendingCount, cursor.PendingMore,
		cursor.ReviewID, cursor.ReviewReason).Scan(&result.NextOffset)
	return result, err
}

func (s *Service) SkipOrder(ctx context.Context, owner, store string, scope OrderScope,
	orderID, reviewID string, offset int) (OrderSyncResult, error) {
	if !validUUID(store) {
		return OrderSyncResult{}, ErrStoreNotFound
	}
	if !orderScopeValid(scope) {
		return OrderSyncResult{}, ErrSiteNotAuthorized
	}
	if !orderNumber(orderID) || !validUUID(reviewID) || offset < 0 || offset > 1<<31-1 {
		return OrderSyncResult{}, ErrOrderSkipInvalid
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrderSyncResult{}, err
	}
	defer tx.Rollback(ctx)
	cursor := orderCursor{Offset: offset}
	var seller, reason string
	err = tx.QueryRow(ctx, `WITH connection AS (`+orderConnection+` FOR SHARE OF c)
 SELECT sync.pending_order_ids,sync.pending_page_count,sync.pending_has_more,sync.review_reason,sync.seller_id`+orderSyncScope+`
 AND sync.pending_order_ids[1]=$5 AND sync.review_id=$6::uuid AND sync.next_offset=$7
 AND (sync.lease_until IS NULL OR sync.lease_until<now()) FOR UPDATE OF sync`,
		store, owner, scope.Site, scope.Logistic, orderID, reviewID, offset).
		Scan(&cursor.PendingIDs, &cursor.PendingCount, &cursor.PendingMore, &reason, &seller)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderSyncResult{}, ErrOrderReviewConflict
	}
	if err != nil {
		return OrderSyncResult{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.meli_order_skips
 (review_id,store_id,site_id,logistic_type,seller_id,order_id,search_offset,reason)
 VALUES ($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8)`,
		reviewID, store, scope.Site, scope.Logistic, seller, orderID, offset, reason)
	if err != nil {
		return OrderSyncResult{}, err
	}
	cursor.PendingIDs = cursor.PendingIDs[1:]
	result, err := updateOrderProgress(ctx, tx, store, scope, cursor)
	if err != nil {
		return OrderSyncResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OrderSyncResult{}, err
	}
	return result, nil
}

// GetOrderSyncState reads the server-persisted cursor for the current verified
// child seller. Never reveal an old seller's state after remapping.
func (s *Service) GetOrderSyncState(ctx context.Context, owner, store string, scope OrderScope) (OrderSyncState, error) {
	var state OrderSyncState
	if !orderScopeValid(scope) {
		return state, ErrSiteNotAuthorized
	}
	if err := s.authorizeOrderRead(ctx, owner, store, scope); err != nil {
		return state, err
	}
	var reviewID, reason, pendingID *string
	err := s.db.QueryRow(ctx, `WITH connection AS (`+orderConnection+`)
SELECT sync.next_offset,
  (c.status='connected' AND sync.next_offset>0 AND sync.window_from IS NOT NULL AND sync.window_to IS NOT NULL
    AND sync.review_id IS NULL AND cardinality(sync.pending_order_ids)=0
    AND (sync.lease_until IS NULL OR sync.lease_until<now())),
  sync.review_id::text,sync.review_reason,sync.pending_order_ids[1]
FROM app.meli_order_syncs sync JOIN connection c ON c.id=sync.store_id AND c.child_id=sync.seller_id
WHERE c.status<>'disconnected' AND sync.site_id=$3 AND sync.logistic_type=$4`,
		store, owner, scope.Site, scope.Logistic).Scan(&state.NextOffset, &state.CanReset, &reviewID, &reason, &pendingID)
	if errors.Is(err, pgx.ErrNoRows) {
		return state, nil
	}
	if err != nil {
		return OrderSyncState{}, err
	}
	if reviewID != nil && reason != nil && pendingID != nil {
		state.Review = &OrderReview{OrderID: *pendingID, ReviewID: *reviewID, Offset: state.NextOffset, Reason: *reason}
	}
	return state, nil
}

// ResetOrderCursor explicitly rewinds an already advanced search page within its
// original time window. It does not delete orders, skips or change the window.
// Incomplete detail pages and active reviews must be resolved, not reset.
func (s *Service) ResetOrderCursor(ctx context.Context, owner, store string, scope OrderScope, expectedOffset int) (OrderSyncResult, error) {
	if !validUUID(store) {
		return OrderSyncResult{}, ErrStoreNotFound
	}
	if !orderScopeValid(scope) {
		return OrderSyncResult{}, ErrSiteNotAuthorized
	}
	if expectedOffset <= 0 || expectedOffset > 1<<31-1 {
		return OrderSyncResult{}, ErrOrderResetInvalid
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return OrderSyncResult{}, err
	}
	defer tx.Rollback(ctx)
	var seller string
	var oldOffset int
	var from, to time.Time
	err = tx.QueryRow(ctx, `WITH connection AS (`+orderConnection+` FOR SHARE OF c)
 SELECT sync.seller_id,sync.next_offset,sync.window_from,sync.window_to`+orderSyncScope+`
 AND c.status='connected' AND sync.next_offset=$5
 AND sync.window_from IS NOT NULL AND sync.window_to IS NOT NULL
 AND sync.review_id IS NULL AND cardinality(sync.pending_order_ids)=0
 AND (sync.lease_until IS NULL OR sync.lease_until<now()) FOR UPDATE OF sync`,
		store, owner, scope.Site, scope.Logistic, expectedOffset).
		Scan(&seller, &oldOffset, &from, &to)
	if errors.Is(err, pgx.ErrNoRows) {
		return OrderSyncResult{}, ErrOrderResetConflict
	}
	if err != nil {
		return OrderSyncResult{}, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.meli_order_resets
 (store_id,site_id,logistic_type,seller_id,previous_offset,window_from,window_to)
 VALUES ($1::uuid,$2,$3,$4,$5,$6,$7)`,
		store, scope.Site, scope.Logistic, seller, oldOffset, from, to)
	if err != nil {
		return OrderSyncResult{}, err
	}
	_, err = tx.Exec(ctx, `UPDATE app.meli_order_syncs
 SET next_offset=0, pending_order_ids='{}'::text[],pending_page_count=0,
 pending_has_more=false,lease_id=NULL,lease_until=NULL
 WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND seller_id=$4`,
		store, scope.Site, scope.Logistic, seller)
	if err != nil {
		return OrderSyncResult{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OrderSyncResult{}, err
	}
	return OrderSyncResult{HasMore: true, NextOffset: 0}, nil
}

func (s *Service) authorizeOrderRead(ctx context.Context, owner, store string, scope OrderScope) error {
	if !validUUID(store) {
		return ErrStoreNotFound
	}
	var status string
	var seller *string
	err := s.db.QueryRow(ctx, `WITH connection AS (`+orderConnection+`)
 SELECT status,child_id FROM connection`, store, owner, scope.Site, scope.Logistic).Scan(&status, &seller)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrStoreNotFound
	}
	if err != nil {
		return err
	}
	if status == "disconnected" {
		return ErrReauthorizationRequired
	}
	if seller == nil || !orderNumber(*seller) {
		return ErrSiteNotAuthorized
	}
	return nil
}

// Recheck the current mapping in the same statement that reads cached orders.
// Shared projection used by both local and cross-store cached order reads.
// Keep scanOrder's destinations in this same column order.
const orderColumnsBase = `o.order_id,COALESCE(o.pack_id,''),o.status,o.currency_id,
 trim_scale(o.total_amount)::text,o.date_created,o.last_updated,COALESCE(o.shipment_id,''),
 COALESCE((
   SELECT jsonb_agg(line.value || jsonb_build_object(
     'listing_image_url',image.url,'cached_image_url',cached.url) ORDER BY line.pos)
   FROM jsonb_array_elements(o.order_items) WITH ORDINALITY AS line(value,pos)
   LEFT JOIN LATERAL (
     SELECT COALESCE(NULLIF(p.details->>'secure_thumbnail',''),
       NULLIF(p.details->'pictures'->0->>'secure_url',''),
       NULLIF(p.details->'pictures'->0->>'url','')) AS url
     FROM app.meli_published_listings p
     WHERE p.store_id=o.store_id AND p.site_id=o.site_id
       AND p.logistic_type=o.logistic_type AND p.listing_id=line.value->>'item_id'
       AND p.details->>'seller_id'=o.seller_id
     LIMIT 1
   ) AS image ON true
   LEFT JOIN LATERAL (
     SELECT entry.value->>'image_url' AS url
     FROM app.meli_orders sibling
     CROSS JOIN LATERAL jsonb_array_elements(sibling.order_items) AS entry(value)
     WHERE image.url IS NULL AND COALESCE(line.value->>'image_url','')=''
       AND sibling.store_id=o.store_id AND sibling.site_id=o.site_id
       AND sibling.logistic_type=o.logistic_type AND sibling.seller_id=o.seller_id
       AND entry.value->>'item_id'=line.value->>'item_id'
       AND COALESCE(entry.value->>'image_url','')<>''
     LIMIT 1
   ) AS cached ON true
 ), '[]'::jsonb),
 COALESCE(p.purchase_order_id,''),
 CASE WHEN p.cost_cny IS NULL THEN NULL ELSE trim_scale(p.cost_cny)::text END,
 COALESCE(p.purchase_tracking_number,''),
 COALESCE(sla.sla_status,''),sla.expected_date,sla.checked_at,`

// List and cross-store search never return the potentially large history array.
const orderColumns = orderColumnsBase + ` CASE WHEN o.shipment_snapshot->>'shipment_id'=o.shipment_id THEN o.shipment_snapshot-'details' END`
const orderDetailColumns = orderColumnsBase + ` CASE WHEN o.shipment_snapshot->>'shipment_id'=o.shipment_id THEN o.shipment_snapshot END`

const orderSelectPrefix = `WITH connection AS (` + orderConnection + `)
 SELECT `
const orderSelectFrom = `
 FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 LEFT JOIN app.meli_order_purchases p
   ON p.store_id=o.store_id AND p.site_id=o.site_id AND p.logistic_type=o.logistic_type
   AND p.seller_id=o.seller_id AND p.order_id=o.order_id
 LEFT JOIN app.meli_order_dispatch_sla sla
   ON sla.store_id=o.store_id AND sla.site_id=o.site_id AND sla.logistic_type=o.logistic_type
   AND sla.seller_id=o.seller_id AND sla.order_id=o.order_id
   AND sla.shipment_id=o.shipment_id
 WHERE c.status<>'disconnected' AND o.site_id=$3 AND o.logistic_type=$4`

const orderSelect = orderSelectPrefix + orderColumns + orderSelectFrom
const orderDetailSelect = orderSelectPrefix + orderDetailColumns + orderSelectFrom

func scanOrder(rows pgx.Rows, extra ...any) (OrderRecord, error) {
	var o OrderRecord
	var items []byte
	var cost sql.NullString
	var deadline, checked sql.NullTime
	destinations := []any{&o.ID, &o.PackID, &o.Status, &o.CurrencyID, &o.TotalAmount,
		&o.DateCreated, &o.LastUpdated, &o.ShipmentID, &items, &o.PurchaseOrderID, &cost,
		&o.PurchaseTrackingNumber, &o.DispatchSLAStatus, &deadline, &checked, &o.Shipment}
	err := rows.Scan(append(destinations, extra...)...)
	if deadline.Valid {
		o.DispatchDeadline = &deadline.Time
	}
	if checked.Valid {
		o.DispatchCheckedAt = &checked.Time
	}
	if err != nil {
		return o, err
	}
	var lines []struct {
		OrderLine
		ListingURL string `json:"listing_image_url"`
		CachedURL  string `json:"cached_image_url"`
	}
	if err := json.Unmarshal(items, &lines); err != nil {
		return o, err
	}
	o.Items = make([]OrderLine, 0, len(lines))
	for _, line := range lines {
		storedURL := line.ImageURL
		line.ImageURL = ""
		for _, candidate := range []string{line.ListingURL, storedURL, line.CachedURL} {
			if orderImageURLAllowed(candidate) {
				line.ImageURL = candidate
				break
			}
		}
		o.Items = append(o.Items, line.OrderLine)
	}
	o.OrderItemsReviewRequired = orderItemsNeedReview(o.Items)
	o.SalesAmountUSD = orderSalesAmountUSD(o.Items)
	if cost.Valid {
		o.PurchaseCostCNY = &cost.String
	}
	return o, nil
}

func orderImageURLAllowed(raw string) bool {
	if raw == "" {
		return false
	}
	if len(raw) > 2048 {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return host == "mlstatic.com" || strings.HasSuffix(host, ".mlstatic.com")
}

// Keep this expression aligned with desktop orderFulfillmentStatus. Raw status
// filters remain backward compatible; new fulfillment_* filters are applied
// before pagination in both scoped list and global search.
const orderFulfillmentStatusSQL = `CASE
 WHEN o.status <> 'paid' OR o.shipment_id IS NULL OR
   (o.shipment_snapshot->>'shipment_id') IS DISTINCT FROM o.shipment_id THEN o.status
 WHEN o.shipment_snapshot->>'status'='ready_to_ship' AND o.shipment_snapshot->>'substatus'='printed' THEN 'label_printed'
 WHEN o.shipment_snapshot->>'status' IN ('pending','handling','ready_to_ship') THEN 'awaiting_dispatch'
 WHEN o.shipment_snapshot->>'status'='delivered' THEN 'delivered'
 WHEN o.shipment_snapshot->>'status'='cancelled' THEN 'shipment_cancelled'
 WHEN o.shipment_snapshot->>'status'='not_delivered' THEN 'logistics_exception'
 WHEN o.shipment_snapshot->>'status'='shipped' THEN CASE
   WHEN o.shipment_snapshot->>'substatus' IN ('receiver_absent','bad_address','delivery_blocked','impassable_zone') THEN 'logistics_exception'
   WHEN o.shipment_snapshot->>'substatus' IN ('out_for_delivery','soon_deliver','at_the_door') THEN 'out_for_delivery'
   WHEN o.shipment_snapshot->>'substatus'='in_transit' THEN 'in_transit'
   ELSE 'shipped' END
 ELSE o.status END`
const orderStatusFilterSQL = `($5='' OR o.status=$5 OR ('fulfillment_' || (` + orderFulfillmentStatusSQL + `))=$5)`

func (s *Service) ListOrders(ctx context.Context, owner, store string,
	scope OrderScope, status, purchaseLink string, offset int) (OrderPage, error) {
	if !orderScopeValid(scope) || offset < 0 || offset > 100000 ||
		!ValidOrderStatus(status) || !ValidPurchaseLinkFilter(purchaseLink) {
		return OrderPage{}, ErrOrderInvalidResponse
	}
	if err := s.authorizeOrderRead(ctx, owner, store, scope); err != nil {
		return OrderPage{}, err
	}
	rows, err := s.db.Query(ctx, orderSelect+` AND `+orderStatusFilterSQL+`
 AND ($8='' OR ($8='linked' AND p.purchase_order_id IS NOT NULL)
   OR ($8='unlinked' AND p.purchase_order_id IS NULL))
 ORDER BY o.date_created DESC,o.order_id DESC LIMIT $6 OFFSET $7`,
		store, owner, scope.Site, scope.Logistic, status, orderPageSize+1, offset, purchaseLink)
	if err != nil {
		return OrderPage{}, err
	}
	defer rows.Close()
	page := OrderPage{Items: []OrderRecord{}, PurchaseFilter: purchaseLink}
	for rows.Next() {
		o, e := scanOrder(rows)
		if e != nil {
			return OrderPage{}, e
		}
		page.Items = append(page.Items, o)
	}
	if err = rows.Err(); err != nil {
		return OrderPage{}, err
	}
	if len(page.Items) > orderPageSize {
		page.HasMore = true
		page.Items = page.Items[:orderPageSize]
	}
	return page, nil
}

func (s *Service) GetOrder(ctx context.Context, owner, store string,
	scope OrderScope, id string) (OrderRecord, error) {
	return s.getOrder(ctx, owner, store, scope, id, true)
}

// GetOrderSummary reads the same scoped cached order without shipment details.
// Use for non-detail consumers such as title translation and image refresh.
func (s *Service) GetOrderSummary(ctx context.Context, owner, store string,
	scope OrderScope, id string) (OrderRecord, error) {
	return s.getOrder(ctx, owner, store, scope, id, false)
}

func (s *Service) getOrder(ctx context.Context, owner, store string,
	scope OrderScope, id string, withDetails bool) (OrderRecord, error) {
	if !orderScopeValid(scope) || !orderNumber(id) {
		return OrderRecord{}, ErrOrderNotFound
	}
	if err := s.authorizeOrderRead(ctx, owner, store, scope); err != nil {
		return OrderRecord{}, err
	}
	query := orderSelect
	if withDetails {
		query = orderDetailSelect
	}
	rows, err := s.db.Query(ctx, query+` AND o.order_id=$5 LIMIT 1`,
		store, owner, scope.Site, scope.Logistic, id)
	if err != nil {
		return OrderRecord{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		if err = rows.Err(); err != nil {
			return OrderRecord{}, err
		}
		return OrderRecord{}, ErrOrderNotFound
	}
	o, err := scanOrder(rows)
	if err != nil {
		return OrderRecord{}, fmt.Errorf("read order: %w", err)
	}
	return o, nil
}

// A fixed, non-sensitive set of platform failure groups is enough for client UX.
func OrderErrorCode(err error) string {
	var platform ListingPlatformError
	var review *OrderReview
	switch {
	case errors.As(err, &review):
		return "order_needs_review"
	case errors.Is(err, ErrOrderReviewConflict):
		return "order_review_conflict"
	case errors.Is(err, ErrOrderResetConflict):
		return "order_reset_conflict"
	case errors.Is(err, ErrOrderResetInvalid):
		return "order_reset_invalid"
	case errors.Is(err, ErrOrderSkipInvalid):
		return "order_skip_invalid"
	case errors.Is(err, ErrStoreNotFound), errors.Is(err, ErrOrderNotFound):
		return "order_not_found"
	case errors.Is(err, ErrSiteNotAuthorized):
		return "order_site_not_authorized"
	case errors.Is(err, ErrReauthorizationRequired):
		return "order_reauth_required"
	case errors.Is(err, ErrRefreshUnknown):
		return "order_refresh_unknown"
	case errors.Is(err, ErrOrderBusy), errors.Is(err, ErrRefreshBusy):
		return "order_busy"
	case errors.Is(err, ErrOrderAuthorizationChanged):
		return "order_auth_changed"
	case errors.Is(err, ErrOrderInvalidResponse), errors.Is(err, ErrListingResponse):
		return "order_invalid_response"
	case errors.As(err, &platform) && int(platform) == 429:
		return "order_rate_limited"
	case errors.As(err, &platform):
		return "order_platform_unavailable"
	case errors.Is(err, context.DeadlineExceeded):
		return "order_timeout"
	case errors.Is(err, context.Canceled):
		return "order_canceled"
	default:
		return "order_unavailable"
	}
}

// Purchase filtering happens before paging.
func ValidPurchaseLinkFilter(value string) bool {
	return value == "" || value == "linked" || value == "unlinked"
}
func ValidOrderStatus(status string) bool {
	return status == "" || (len(status) <= 80 &&
		strings.IndexFunc(status, func(c rune) bool {
			return (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_'
		}) == -1)
}
