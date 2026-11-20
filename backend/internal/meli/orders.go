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
	LeaseID string
	From    time.Time
	To      time.Time
	Offset  int
}

func orderScopeValid(scope OrderScope) bool {
	return scope.Site != "" && scope.Site != "CBT" && scope.Logistic != "" &&
		len(scope.Site) <= 8 && len(scope.Logistic) <= 40 &&
		regexp.MustCompile(`^[A-Z]{3}$`).MatchString(scope.Site) &&
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
		ID          json.Number `json:"id"`
		PackID      json.Number `json:"pack_id"`
		Status      string      `json:"status"`
		SiteID      string      `json:"site_id"`
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
		(payload.SiteID != "" && payload.SiteID != site) ||
		payload.Status == "" || len(payload.Status) > 80 ||
		payload.CurrencyID == "" || len(payload.CurrencyID) > 12 ||
		!orderAmountPattern.MatchString(payload.TotalAmount.String()) {
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

func (s *Service) acquireOrderCursor(ctx context.Context, store string, scope OrderScope, seller string) (orderCursor, error) {
	var cursor orderCursor
	err := s.db.QueryRow(ctx, `
INSERT INTO app.meli_order_syncs
 (store_id,site_id,logistic_type,seller_id,window_from,window_to,lease_id,lease_until)
VALUES ($1::uuid,$2,$3,$4,now()-interval '30 days',now(),gen_random_uuid(),now()+interval '45 seconds')
ON CONFLICT (store_id,site_id,logistic_type) DO UPDATE
 SET lease_id=EXCLUDED.lease_id,
     lease_until=EXCLUDED.lease_until,
     window_from=COALESCE(app.meli_order_syncs.window_from,
       COALESCE(app.meli_order_syncs.last_completed_to - interval '10 minutes', EXCLUDED.window_from)),
     window_to=COALESCE(app.meli_order_syncs.window_to, EXCLUDED.window_to)
 WHERE app.meli_order_syncs.seller_id=EXCLUDED.seller_id AND
       (app.meli_order_syncs.lease_until IS NULL OR app.meli_order_syncs.lease_until<now())
RETURNING lease_id::text,window_from,window_to,next_offset`,
		store, scope.Site, scope.Logistic, seller).Scan(&cursor.LeaseID, &cursor.From, &cursor.To, &cursor.Offset)
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

// SyncOrders processes one small page. No order or cursor is committed on a failed
// page; a later request safely replays it with UPSERT.
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
	cursor, err := s.acquireOrderCursor(ctx, store, scope, childID)
	if err != nil {
		return OrderSyncResult{}, err
	}
	defer s.releaseOrderCursor(store, scope, cursor.LeaseID)

	fetchCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
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
	ids, total, count, err := parseOrderSearch(search, childID, cursor.Offset)
	if err != nil {
		return OrderSyncResult{}, err
	}
	records := make([]OrderRecord, 0, len(ids))
	for _, id := range ids {
		var raw json.RawMessage
		_, err = s.listingGET(fetchCtx, auth.AccessToken, "/marketplace/orders/"+id, &raw, nil)
		if err != nil {
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
	next := cursor.Offset + count
	hasMore := next < total
	if err = s.saveOrderBatch(ctx, owner, store, scope, auth.SellerID, childID, cursor, records, next, hasMore); err != nil {
		return OrderSyncResult{}, err
	}
	return OrderSyncResult{Saved: len(records), Processed: count, HasMore: hasMore, NextOffset: next}, nil
}

func (s *Service) saveOrderBatch(ctx context.Context, owner, store string, scope OrderScope,
	globalSeller, childSeller string, cursor orderCursor, records []OrderRecord, next int, hasMore bool) error {
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
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND seller_id=$4 AND lease_id=$5::uuid FOR UPDATE`,
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
SET pack_id=EXCLUDED.pack_id,status=EXCLUDED.status,currency_id=EXCLUDED.currency_id,
total_amount=EXCLUDED.total_amount,date_created=EXCLUDED.date_created,
last_updated=EXCLUDED.last_updated,shipment_id=EXCLUDED.shipment_id,
order_items=EXCLUDED.order_items,updated_at=now()
WHERE app.meli_orders.last_updated<=EXCLUDED.last_updated`,
			store, scope.Site, scope.Logistic, childSeller, rec.ID, rec.PackID, rec.Status,
			rec.CurrencyID, rec.TotalAmount, rec.DateCreated, rec.LastUpdated, rec.ShipmentID, string(items))
		if err != nil {
			return err
		}
	}
	if hasMore {
		_, err = tx.Exec(ctx, `UPDATE app.meli_order_syncs SET next_offset=$5,last_synced_at=now(),lease_id=NULL,lease_until=NULL
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND lease_id=$4::uuid`,
			store, scope.Site, scope.Logistic, cursor.LeaseID, next)
	} else {
		_, err = tx.Exec(ctx, `UPDATE app.meli_order_syncs
SET next_offset=0,last_completed_to=window_to,window_from=NULL,window_to=NULL,
last_synced_at=now(),lease_id=NULL,lease_until=NULL
WHERE store_id=$1::uuid AND site_id=$2 AND logistic_type=$3 AND lease_id=$4::uuid`,
			store, scope.Site, scope.Logistic, cursor.LeaseID)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const orderSelect = `SELECT o.order_id,COALESCE(o.pack_id,''),o.status,o.currency_id,
 o.total_amount::text,o.date_created,o.last_updated,COALESCE(o.shipment_id,''),o.order_items
 FROM app.meli_orders o
 JOIN app.meli_store_connections c ON c.id=o.store_id
 WHERE c.id=$1::uuid AND c.owner_user_id=$2::uuid AND c.status='connected'
 AND o.site_id=$3 AND o.logistic_type=$4 AND o.seller_id=$5`

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
	_, seller, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderPage{}, err
	}
	rows, err := s.db.Query(ctx, orderSelect+` AND ($6='' OR o.status=$6)
ORDER BY o.date_created DESC,o.order_id DESC LIMIT $7 OFFSET $8`,
		store, owner, scope.Site, scope.Logistic, seller, status, orderPageSize+1, offset)
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
	_, seller, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return OrderRecord{}, err
	}
	rows, err := s.db.Query(ctx, orderSelect+` AND o.order_id=$6 LIMIT 1`,
		store, owner, scope.Site, scope.Logistic, seller, id)
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
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "order_timeout"
	default:
		return "order_unavailable"
	}
}

func ValidOrderStatus(status string) bool {
	return status == "" || (len(status) <= 80 &&
		strings.IndexFunc(status, func(c rune) bool {
			return c < 'a' || c > 'z' && (c < '0' || c > '9') && c != '_'
		}) == -1)
}
