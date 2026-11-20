package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

var (
	ErrOrderBusy                 = errors.New("order sync already running")
	ErrOrderInvalidResponse      = errors.New("invalid order response")
	ErrOrderAuthorizationChanged = errors.New("order authorization changed")
	ErrOrderNotFound             = errors.New("order not found")
	orderNumberPattern           = regexp.MustCompile(`^[0-9]{1,30}$`)
	orderAmountPattern           = regexp.MustCompile(`^[0-9]{1,16}(\.[0-9]{1,8})?$`)
)

type OrderScope struct {
	Site     string
	Logistic string
}

type OrderLine struct {
	ItemID      string `json:"item_id"`
	Title       string `json:"title"`
	VariationID string `json:"variation_id,omitempty"`
	Quantity    int    `json:"quantity"`
	UnitPrice   string `json:"unit_price"`
	CurrencyID  string `json:"currency_id"`
}

type OrderRecord struct {
	ID          string      `json:"order_id"`
	PackID      string      `json:"pack_id,omitempty"`
	Status      string      `json:"status"`
	CurrencyID  string      `json:"currency_id"`
	TotalAmount string      `json:"total_amount"`
	DateCreated time.Time   `json:"date_created"`
	LastUpdated time.Time   `json:"last_updated"`
	ShipmentID  string      `json:"shipment_id,omitempty"`
	Items       []OrderLine `json:"order_items"`
}

type OrderPage struct {
	Items   []OrderRecord `json:"items"`
	HasMore bool          `json:"has_more"`
}

type OrderSyncResult struct {
	Saved      int  `json:"saved"`
	Processed  int  `json:"processed"`
	HasMore    bool `json:"has_more"`
	NextOffset int  `json:"next_offset"`
}

type orderCursor struct {
	LeaseID      string
	From         time.Time
	To           time.Time
	Offset       int
	PendingIDs   []string
	PendingCount int
	PendingMore  bool
}

func orderScopeValid(scope OrderScope) bool {
	return scope.Site != "" && scope.Site != "CBT" && scope.Logistic != "" &&
		len(scope.Site) >= 3 && len(scope.Site) <= 8 && len(scope.Logistic) <= 40 &&
		strings.ToUpper(scope.Site) == scope.Site && strings.TrimSpace(scope.Site) == scope.Site &&
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
	if payload.SiteID == nil && payload.Context.Site == nil {
		return OrderRecord{}, ErrOrderInvalidResponse
	}
	for _, field := range []json.RawMessage{payload.SiteID, payload.Context.Site} {
		if field != nil {
			var reportedSite string
			if json.Unmarshal(field, &reportedSite) != nil || reportedSite == "" || reportedSite != site {
				return OrderRecord{}, ErrOrderInvalidResponse
			}
		}
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
		record.Items = append(record.Items, line)
	}
	return record, nil
}

const orderConnection = `SELECT c.id,c.status,m.value->>'user_id' AS child_id
 FROM app.meli_store_connections c
 LEFT JOIN LATERAL jsonb_array_elements(c.marketplaces) m(value)
   ON c.marketplaces_status='verified' AND m.value->>'site_id'=$3 AND m.value->>'logistic_type'=$4
 WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid`

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
     pending_has_more=sync.seller_id=EXCLUDED.seller_id AND sync.pending_has_more
 WHERE sync.lease_until IS NULL OR sync.lease_until<now()
RETURNING lease_id::text,window_from,window_to,next_offset,pending_order_ids,pending_page_count,pending_has_more`,
		store, owner, scope.Site, scope.Logistic, seller).Scan(&cursor.LeaseID, &cursor.From, &cursor.To, &cursor.Offset,
		&cursor.PendingIDs, &cursor.PendingCount, &cursor.PendingMore)
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

// SyncOrders commits one detail batch and advances the search offset only after the whole page is saved.
func (s *Service) SyncOrders(ctx context.Context, owner, store string, scope OrderScope) (OrderSyncResult, error) {
	if !orderScopeValid(scope) {
		return OrderSyncResult{}, ErrSiteNotAuthorized
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

	fetchCtx, cancelFetch := context.WithTimeout(ctx, 15*time.Second)
	defer cancelFetch()
	ids, count, hasNextPage := cursor.PendingIDs, cursor.PendingCount, cursor.PendingMore
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
		_, err = s.listingGET(fetchCtx, auth.AccessToken, "/marketplace/orders/search?"+query.Encode(), &search, nil)
		if err != nil {
			return OrderSyncResult{}, err
		}
		var total int
		ids, total, count, err = parseOrderSearch(search, childID, cursor.Offset)
		if err != nil {
			return OrderSyncResult{}, err
		}
		hasNextPage = cursor.Offset+count < total
	}
	batchSize := min(orderDetailBatchSize, len(ids))
	records := make([]OrderRecord, 0, batchSize)
	for _, id := range ids[:batchSize] {
		var raw json.RawMessage
		_, err = s.listingGET(fetchCtx, auth.AccessToken, "/marketplace/orders/"+id, &raw, nil)
		if err != nil {
			// A provider timeout can leave a fetched prefix or a newly discovered page.
			// Preserve it only while the caller's context is still active.
			if errors.Is(err, context.DeadlineExceeded) && ctx.Err() == nil &&
				(len(records) > 0 || len(cursor.PendingIDs) == 0) {
				break
			}
			return OrderSyncResult{}, err
		}
		record, parseErr := parseOrderDetail(raw, id, childID, scope.Site)
		if parseErr != nil {
			return OrderSyncResult{}, parseErr
		}
		records = append(records, record)
	}
	if ctx.Err() != nil {
		return OrderSyncResult{}, ctx.Err()
	}
	remaining := ids[len(records):]
	pageComplete := len(remaining) == 0
	result := OrderSyncResult{Saved: len(records), HasMore: !pageComplete || hasNextPage, NextOffset: cursor.Offset}
	if pageComplete {
		result.Processed = count
		result.NextOffset += count
		count = 0
		hasNextPage = false
	}
	if err = s.saveOrderBatch(ctx, owner, store, scope, auth.SellerID, childID, cursor, records,
		result.NextOffset, pageComplete && !result.HasMore, remaining, count, hasNextPage); err != nil {
		return OrderSyncResult{}, err
	}
	return result, nil
}

func (s *Service) saveOrderBatch(ctx context.Context, owner, store string, scope OrderScope,
	globalSeller, childSeller string, cursor orderCursor, records []OrderRecord, next int, windowComplete bool,
	remaining []string, pendingCount int, pendingHasMore bool) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var seller, status, marketStatus string
	var rawMarkets []byte
	err = tx.QueryRow(ctx, `SELECT seller_id,status,marketplaces_status,marketplaces FROM app.meli_store_connections
WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR SHARE`, store, owner).
		Scan(&seller, &status, &marketStatus, &rawMarkets)
	if err != nil {
		return ErrOrderAuthorizationChanged
	}
	var markets []Marketplace
	if json.Unmarshal(rawMarkets, &markets) != nil || seller != globalSeller ||
		status != "connected" || marketStatus != "verified" {
		return ErrOrderAuthorizationChanged
	}
	allowed := false
	for _, m := range markets {
		if m.SiteID == scope.Site && m.LogisticType == scope.Logistic && m.UserID.String() == childSeller {
			allowed = true
		}
	}
	if !allowed {
		return ErrOrderAuthorizationChanged
	}
	var offset int
	var from, to time.Time
	err = tx.QueryRow(ctx, `SELECT next_offset,window_from,window_to FROM app.meli_order_syncs
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND seller_id=$4 AND lease_id=$5::uuid AND lease_until>now() FOR UPDATE`,
		store, scope.Site, scope.Logistic, childSeller, cursor.LeaseID).Scan(&offset, &from, &to)
	if err != nil || offset != cursor.Offset || !from.Equal(cursor.From) || !to.Equal(cursor.To) {
		return ErrOrderBusy
	}
	for _, rec := range records {
		items, jsonErr := json.Marshal(rec.Items)
		if jsonErr != nil {
			return jsonErr
		}
		_, err = tx.Exec(ctx, `INSERT INTO app.meli_orders
(store_id,site_id,logistic_type,seller_id,order_id,pack_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,order_items)
VALUES($1::uuid,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9::numeric,$10,$11,NULLIF($12,''),$13::jsonb)
ON CONFLICT (store_id,site_id,logistic_type,order_id) DO UPDATE
SET seller_id=EXCLUDED.seller_id,pack_id=EXCLUDED.pack_id,status=EXCLUDED.status,currency_id=EXCLUDED.currency_id,
total_amount=EXCLUDED.total_amount,date_created=EXCLUDED.date_created,
last_updated=EXCLUDED.last_updated,shipment_id=EXCLUDED.shipment_id,
order_items=EXCLUDED.order_items,updated_at=now()
WHERE app.meli_orders.seller_id<>EXCLUDED.seller_id OR app.meli_orders.last_updated<=EXCLUDED.last_updated`,
			store, scope.Site, scope.Logistic, childSeller, rec.ID, rec.PackID, rec.Status,
			rec.CurrencyID, rec.TotalAmount, rec.DateCreated, rec.LastUpdated, rec.ShipmentID, string(items))
		if err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `UPDATE app.meli_order_syncs
SET next_offset=CASE WHEN $6 THEN 0 ELSE $5 END,
 last_completed_to=CASE WHEN $6 THEN window_to ELSE last_completed_to END,
 window_from=CASE WHEN NOT $6 THEN window_from END,window_to=CASE WHEN NOT $6 THEN window_to END,
 pending_order_ids=$7,pending_page_count=$8,pending_has_more=$9,
 last_synced_at=now(),lease_id=NULL,lease_until=NULL
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND lease_id=$4::uuid`,
		store, scope.Site, scope.Logistic, cursor.LeaseID, next, windowComplete, remaining, pendingCount, pendingHasMore)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
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
const orderSelect = `WITH connection AS (` + orderConnection + `)
 SELECT o.order_id,COALESCE(o.pack_id,''),o.status,o.currency_id,
 trim_scale(o.total_amount)::text,o.date_created,o.last_updated,COALESCE(o.shipment_id,''),o.order_items
 FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 WHERE c.status<>'disconnected' AND o.site_id=$3 AND o.logistic_type=$4`

func scanOrder(rows pgx.Rows) (OrderRecord, error) {
	var o OrderRecord
	var items []byte
	err := rows.Scan(&o.ID, &o.PackID, &o.Status, &o.CurrencyID, &o.TotalAmount,
		&o.DateCreated, &o.LastUpdated, &o.ShipmentID, &items)
	if err == nil {
		err = json.Unmarshal(items, &o.Items)
	}
	return o, err
}

func (s *Service) ListOrders(ctx context.Context, owner, store string,
	scope OrderScope, status string, offset int) (OrderPage, error) {
	if !orderScopeValid(scope) || offset < 0 || offset > 100000 || len(status) > 80 {
		return OrderPage{}, ErrOrderInvalidResponse
	}
	if err := s.authorizeOrderRead(ctx, owner, store, scope); err != nil {
		return OrderPage{}, err
	}
	rows, err := s.db.Query(ctx, orderSelect+` AND ($5='' OR o.status=$5)
ORDER BY o.date_created DESC,o.order_id DESC LIMIT $6 OFFSET $7`,
		store, owner, scope.Site, scope.Logistic, status, orderPageSize+1, offset)
	if err != nil {
		return OrderPage{}, err
	}
	defer rows.Close()
	page := OrderPage{Items: []OrderRecord{}}
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
	if !orderScopeValid(scope) || !orderNumber(id) {
		return OrderRecord{}, ErrOrderNotFound
	}
	if err := s.authorizeOrderRead(ctx, owner, store, scope); err != nil {
		return OrderRecord{}, err
	}
	rows, err := s.db.Query(ctx, orderSelect+` AND o.order_id=$5 LIMIT 1`,
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
	switch {
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

func ValidOrderStatus(status string) bool {
	return status == "" || (len(status) <= 80 &&
		strings.IndexFunc(status, func(c rune) bool {
			return (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_'
		}) == -1)
}
