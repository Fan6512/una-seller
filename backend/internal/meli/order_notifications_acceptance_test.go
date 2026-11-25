package meli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WH-03 acceptance: order-first, pack fan-out, duplicate/out-of-order delivery,
// tenancy, ambiguous/>10/mixed/stale shipments, partial responses, revoked
// authorization between read and commit, manual concurrency and shared limits.

type upstreamCall struct {
	path   string
	header http.Header
}

type notificationCase struct {
	service *Service
	db      *pgxpool.Pool
	owner   string
	store   string
	parent  string
	calls   func() []upstreamCall
}

func newNotificationCase(t *testing.T, serve http.HandlerFunc) notificationCase {
	t.Helper()
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var parent string
	if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1 RETURNING seller_id`, store).Scan(&parent); err != nil {
		t.Fatal(err)
	}
	var mu sync.Mutex
	var recorded []upstreamCall
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		recorded = append(recorded, upstreamCall{path: r.URL.Path, header: r.Header.Clone()})
		mu.Unlock()
		serve(w, r)
	}))
	t.Cleanup(server.Close)
	s.apiURL = server.URL
	return notificationCase{service: s, db: db, owner: owner, store: store, parent: parent,
		calls: func() []upstreamCall {
			mu.Lock()
			defer mu.Unlock()
			return append([]upstreamCall(nil), recorded...)
		}}
}

func (c notificationCase) job(topic, resource, user string) notifications.Job {
	return notifications.Job{StoreID: c.store, OwnerID: c.owner, ParentSellerID: c.parent,
		NotificationUserID: user, Topic: topic, ResourceID: resource}
}

func (c notificationCase) prepare(t *testing.T, job notifications.Job) (notifications.VerifiedResource, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return c.service.NotificationHandlers()[job.Topic](ctx, job)
}

func (c notificationCase) apply(t *testing.T, resource notifications.VerifiedResource) error {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return pgx.BeginFunc(ctx, c.db, func(tx pgx.Tx) error { return resource.Apply(ctx, tx) })
}

func (c notificationCase) count(t *testing.T, query string, args ...any) int {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var got int
	if err := c.db.QueryRow(ctx, query, args...).Scan(&got); err != nil {
		t.Fatalf("%v\n%s", err, query)
	}
	return got
}

func orderPayload(id, seller, site, pack, shipment, status, updated string) string {
	return fmt.Sprintf(`{"id":%s,"seller":{"id":%s},"site_id":%q,"pack_id":%s,"status":%q,
"currency_id":"USD","total_amount":29.99,"shipping":{"id":%s},
"date_created":"2026-11-20T00:00:00Z","last_updated":%q,
"order_items":[{"item":{"id":"MLB100","title":"Item"},"quantity":1,"unit_price":29.99,"currency_id":"USD"}],
"buyer":{"email":"private@example.invalid"}}`, id, seller, site, pack, status, shipment, updated)
}

func shipmentPayload(id, sender, site, status, updated string) string {
	return fmt.Sprintf(`{"id":%s,"sender_id":%s,"site_id":%q,"status":%q,"tracking_number":"TRACK%s",
"logistic":{"type":"drop_off"},"last_updated":%q}`, id, sender, site, status, id, updated)
}

// linkRow mirrors the documented association shape: order_id/pack_id are
// strings, seller_id is a number, and each purchased item repeats the order.
func linkRow(orderID, packID, seller string) string {
	return fmt.Sprintf(`{"order_id":%q,"pack_id":%q,"item_id":"MLB100","variation_id":null,
"user_product_id":"MLAU147563159","seller_id":%s,"requested_quantity":1}`, orderID, packID, seller)
}

func linksBody(rows ...string) string { return "[" + strings.Join(rows, ",") + "]" }

func assertHeader(t *testing.T, call upstreamCall, key, want string) {
	t.Helper()
	if got := call.header.Get(key); got != want {
		t.Fatalf("%s: header %s=%q want %q", call.path, key, got, want)
	}
}

func TestNotificationOrderFirstThenShipmentPackPostgres(t *testing.T) {
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/marketplace/orders/10001", "/marketplace/orders/10002":
			id := strings.TrimPrefix(r.URL.Path, "/marketplace/orders/")
			fmt.Fprint(w, notifiedOrderFixture(id, "2026-11-21T01:00:00Z", "paid", "801"))
		case "/marketplace/shipments/801":
			fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "delivered", "2026-11-22T00:00:00Z"))
		case "/shipments/801/orders":
			fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654"), linkRow("10002", "700", "987654")))
		default:
			http.NotFound(w, r)
		}
	})

	orderJob := c.job("marketplace_orders", "10001", "987654")
	resource, err := c.prepare(t, orderJob)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders
WHERE store_id=$1 AND order_id='10001' AND pack_id='700' AND shipment_id='801' AND shipment_snapshot IS NULL`, c.store); got != 1 {
		t.Fatalf("order-first row missing or snapshot written early: %d", got)
	}

	shipmentJob := c.job("marketplace_shipments", "801", "987654")
	resource, err = c.prepare(t, shipmentJob)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders
WHERE store_id=$1 AND pack_id='700' AND shipment_id='801'
AND shipment_snapshot->>'status'='delivered' AND shipment_snapshot->>'tracking_number'='TRACK801'`, c.store); got != 2 {
		t.Fatalf("pack fan-out did not attach both snapshots: %d", got)
	}

	// Replaying the same delivery must stay idempotent and must not fan out
	// another order read per duplicate association row.
	resource, err = c.prepare(t, shipmentJob)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1`, c.store); got != 2 {
		t.Fatalf("duplicate replay changed row count: %d", got)
	}

	var association, detail upstreamCall
	for _, call := range c.calls() {
		switch call.path {
		case "/shipments/801/orders":
			association = call
		case "/marketplace/shipments/801":
			detail = call
		}
	}
	assertHeader(t, association, "X-New-Domain", "true")
	if association.header.Get("x-format-new") != "" {
		t.Fatal("association endpoint must not inherit the shipment detail format header")
	}
	assertHeader(t, detail, "x-format-new", "true")
	assertHeader(t, detail, "X-New-Domain", "")
}

func TestNotificationShipmentFailsClosedOnUnsafeLists(t *testing.T) {
	eleven := make([]string, 0, 11)
	for i := 1; i <= 11; i++ {
		eleven = append(eleven, linkRow(fmt.Sprintf("100%02d", i), "700", "987654"))
	}
	cases := []struct {
		name   string
		serve  func(w http.ResponseWriter, r *http.Request)
		scope  bool
		mutate func(t *testing.T, c notificationCase)
	}{
		{name: "more than ten unique orders", scope: false, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprint(w, linksBody(eleven...))
			default:
				id := strings.TrimPrefix(r.URL.Path, "/marketplace/orders/")
				fmt.Fprint(w, notifiedOrderFixture(id, "2026-11-21T01:00:00Z", "paid", "801"))
			}
		}},
		{name: "mixed seller association", scope: true, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654"), linkRow("10002", "700", "987655")))
			default:
				id := strings.TrimPrefix(r.URL.Path, "/marketplace/orders/")
				seller := "987654"
				if id == "10002" {
					seller = "987655"
				}
				fmt.Fprint(w, orderPayload(id, seller, "MLB", "700", "801", "paid", "2026-11-21T01:00:00Z"))
			}
		}},
		{name: "association seller differs from order", scope: true, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprint(w, linksBody(linkRow("10001", "700", "424242")))
			default:
				fmt.Fprint(w, notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "paid", "801"))
			}
		}},
		{name: "shipment sender outside the mapping", scope: true, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "424242", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654")))
			default:
				fmt.Fprint(w, notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "paid", "801"))
			}
		}},
		{name: "shipment site conflicts with order site", scope: true, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLA", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654")))
			default:
				fmt.Fprint(w, notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "paid", "801"))
			}
		}},
		{name: "no orders yet", scope: false, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				w.WriteHeader(http.StatusNoContent)
			default:
				http.NotFound(w, r)
			}
		}},
		{name: "partial content array", scope: false, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				w.WriteHeader(http.StatusPartialContent)
				fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654")))
			default:
				fmt.Fprint(w, notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "paid", "801"))
			}
		}},
		{name: "wrapped collection instead of an array", scope: false, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprintf(w, `{"results":%s}`, linksBody(linkRow("10001", "700", "987654")))
			default:
				http.NotFound(w, r)
			}
		}},
		{name: "conflicting duplicate rows", scope: false, serve: func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/marketplace/shipments/801":
				fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "shipped", "2026-11-22T00:00:00Z"))
			case "/shipments/801/orders":
				fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654"), linkRow("10001", "701", "987654")))
			default:
				http.NotFound(w, r)
			}
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newNotificationCase(t, tc.serve)
			resource, err := c.prepare(t, c.job("marketplace_shipments", "801", "987654"))
			if err == nil {
				t.Fatalf("unsafe association accepted: %+v", resource)
			}
			if got := errors.Is(err, notifications.ErrScope); got != tc.scope {
				t.Fatalf("scope classification=%v want %v: %v", got, tc.scope, err)
			}
			if got := c.count(t, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1`, c.store); got != 0 {
				t.Fatalf("failed prepare wrote %d orders", got)
			}
		})
	}
}

func TestNotificationStaleShipmentDoesNotReplaceSnapshot(t *testing.T) {
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/marketplace/shipments/801":
			fmt.Fprint(w, shipmentPayload("801", "987654", "MLB", "delivered", "2026-11-22T00:00:00Z"))
		case "/shipments/801/orders":
			fmt.Fprint(w, linksBody(linkRow("10001", "700", "987654")))
		case "/marketplace/orders/10001":
			// The order already moved to the replacement shipment.
			fmt.Fprint(w, notifiedOrderFixture("10001", "2026-11-23T01:00:00Z", "paid", "900"))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	if err := pgx.BeginFunc(ctx, c.db, func(tx pgx.Tx) error {
		return upsertOrderRecord(ctx, tx, c.store, OrderScope{Site: "MLB", Logistic: "remote"}, "987654",
			OrderRecord{ID: "10001", PackID: "700", Status: "paid", CurrencyID: "USD", TotalAmount: "29.99", Items: []OrderLine{},
				DateCreated: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC),
				LastUpdated: time.Date(2026, 11, 23, 1, 0, 0, 0, time.UTC), ShipmentID: "900"})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.prepare(t, c.job("marketplace_shipments", "801", "987654")); !errors.Is(err, ErrOrderShipmentChanged) {
		t.Fatalf("replaced shipment accepted: %v", err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders
WHERE store_id=$1 AND order_id='10001' AND shipment_id='900' AND shipment_snapshot IS NULL`, c.store); got != 1 {
		t.Fatalf("stale notification touched the replacement shipment: %d", got)
	}
}

func TestNotificationKeepsNewerManualOrderAndProtectedData(t *testing.T) {
	var delivered atomic.Value
	delivered.Store(notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "cancelled", "801"))
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/marketplace/orders/10001":
			fmt.Fprint(w, delivered.Load().(string))
		default:
			http.NotFound(w, r)
		}
	})
	ctx := context.Background()
	newer := time.Date(2026, 11, 25, 3, 0, 0, 0, time.UTC)
	if err := pgx.BeginFunc(ctx, c.db, func(tx pgx.Tx) error {
		if err := upsertOrderRecord(ctx, tx, c.store, OrderScope{Site: "MLB", Logistic: "remote"}, "987654",
			OrderRecord{ID: "10001", PackID: "700", Status: "paid", CurrencyID: "USD", TotalAmount: "29.99", Items: []OrderLine{},
				DateCreated: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), LastUpdated: newer, ShipmentID: "801"}); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE app.meli_orders SET billing_snapshot='{"sentinel":true}'::jsonb,
shipment_snapshot='{"shipment_id":"801","status":"delivered","last_updated":"2026-11-24T00:00:00Z","checked_at":"2026-11-25T02:00:00Z"}'::jsonb
WHERE store_id=$1 AND order_id='10001'`, c.store); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO app.meli_order_purchases
(store_id,site_id,logistic_type,seller_id,order_id,purchase_order_id,cost_cny)
VALUES($1,'MLB','remote','987654','10001','12345678',15.50)`, c.store); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO app.meli_order_syncs(store_id,site_id,logistic_type,seller_id,next_offset)
VALUES($1,'MLB','remote','987654',7)`, c.store)
		return err
	}); err != nil {
		t.Fatal(err)
	}

	resource, err := c.prepare(t, c.job("marketplace_orders", "10001", "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders o
JOIN app.meli_order_purchases p USING(store_id,site_id,logistic_type,seller_id,order_id)
WHERE o.store_id=$1 AND o.order_id='10001' AND o.status='paid' AND o.shipment_id='801'
AND o.last_updated='2026-11-25T03:00:00Z' AND o.billing_snapshot->>'sentinel'='true'
AND o.shipment_snapshot->>'status'='delivered' AND p.cost_cny=15.50`, c.store); got != 1 {
		t.Fatalf("older notification overwrote newer manual state or protected data: %d", got)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_order_syncs WHERE store_id=$1 AND next_offset=7`, c.store); got != 1 {
		t.Fatalf("notification advanced the manual cursor: %d", got)
	}

	// A manual batch that finishes after a newer notification must not roll the
	// order back either: the upsert guard is the shared ordering rule.
	if err := pgx.BeginFunc(ctx, c.db, func(tx pgx.Tx) error {
		return upsertOrderRecord(ctx, tx, c.store, OrderScope{Site: "MLB", Logistic: "remote"}, "987654",
			OrderRecord{ID: "10001", PackID: "700", Status: "cancelled", CurrencyID: "USD", TotalAmount: "1.00", Items: []OrderLine{},
				DateCreated: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), LastUpdated: newer.Add(-48 * time.Hour), ShipmentID: "802"})
	}); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders
WHERE store_id=$1 AND order_id='10001' AND status='paid' AND shipment_id='801' AND total_amount=29.99`, c.store); got != 1 {
		t.Fatalf("late manual batch overwrote a newer order: %d", got)
	}

	// Out-of-order deliveries of the same resource: the delayed older body must
	// not regress the row written by the newer delivery.
	for _, payload := range []string{
		notifiedOrderFixture("10001", "2026-11-25T04:00:00Z", "shipped", "801"),
		notifiedOrderFixture("10001", "2026-11-24T00:00:00Z", "paid", "801"),
	} {
		delivered.Store(payload)
		resource, err := c.prepare(t, c.job("marketplace_orders", "10001", "987654"))
		if err != nil {
			t.Fatal(err)
		}
		if err := c.apply(t, resource); err != nil {
			t.Fatal(err)
		}
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders
WHERE store_id=$1 AND order_id='10001' AND status='shipped' AND last_updated='2026-11-25T04:00:00Z'`, c.store); got != 1 {
		t.Fatalf("out-of-order delivery regressed the order: %d", got)
	}
}

func TestNotificationTenancyAndRevocationPostgres(t *testing.T) {
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/marketplace/orders/10001" {
			fmt.Fprint(w, notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "paid", "801"))
			return
		}
		http.NotFound(w, r)
	})
	ctx := context.Background()
	var otherStore string
	access, err := c.service.cipher.Seal("synthetic_other_access", tokenAAD(c.owner, "555001", "access"))
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := c.service.cipher.Seal("synthetic_other_refresh", tokenAAD(c.owner, "555001", "refresh"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
(owner_user_id,seller_id,nickname,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at,
 marketplaces,marketplaces_status,status)
VALUES($1::uuid,'555001','Other store',$2,$3,now()+interval '1 hour',
       '[{"user_id":987655,"site_id":"MLB","logistic_type":"remote"}]','verified','connected')
RETURNING id::text`, c.owner, access, refresh).Scan(&otherStore); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		clean, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, _ = c.db.Exec(clean, `DELETE FROM app.meli_store_connections WHERE id=$1`, otherStore)
	})
	// The same upstream order ID already exists in the other store.
	if err := pgx.BeginFunc(ctx, c.db, func(tx pgx.Tx) error {
		return upsertOrderRecord(ctx, tx, otherStore, OrderScope{Site: "MLB", Logistic: "remote"}, "987655",
			OrderRecord{ID: "10001", Status: "paid", CurrencyID: "USD", TotalAmount: "29.99", Items: []OrderLine{},
				DateCreated: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC),
				LastUpdated: time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC), ShipmentID: "555"})
	}); err != nil {
		t.Fatal(err)
	}

	resource, err := c.prepare(t, c.job("marketplace_orders", "10001", "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1 AND order_id='10001' AND shipment_id='555'`, otherStore); got != 1 {
		t.Fatalf("cross-store row changed: %d", got)
	}

	// A job claiming another parent's identity must be rejected before any read.
	wrongParent := c.job("marketplace_orders", "10001", "987654")
	wrongParent.ParentSellerID = "424242"
	if _, err := c.prepare(t, wrongParent); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("foreign parent accepted: %v", err)
	}
	// The second store's child cannot read this order.
	if _, err := c.prepare(t, notifications.Job{StoreID: otherStore, OwnerID: c.owner, ParentSellerID: "555001",
		NotificationUserID: "987655", Topic: "marketplace_orders", ResourceID: "10001"}); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("foreign child accepted: %v", err)
	}

	// Authorization revoked after the official read but before the commit.
	resource, err = c.prepare(t, c.job("marketplace_orders", "10001", "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.db.Exec(ctx, `DELETE FROM app.meli_orders WHERE store_id=$1 AND order_id='10001'`, c.store); err != nil {
		t.Fatal(err)
	}
	if _, err := c.db.Exec(ctx, `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, c.store); err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); err == nil {
		t.Fatal("revoked authorization still committed")
	} else if !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("revocation misclassified for the worker: %v", err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1 AND order_id='10001'`, c.store); got != 0 {
		t.Fatalf("revoked commit wrote %d rows", got)
	}
	if _, err := c.db.Exec(ctx, `UPDATE app.meli_store_connections SET status='connected' WHERE id=$1`, c.store); err != nil {
		t.Fatal(err)
	}

	// The mapping changed between read and commit: still no write.
	resource, err = c.prepare(t, c.job("marketplace_orders", "10001", "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.db.Exec(ctx, `UPDATE app.meli_store_connections
SET marketplaces='[{"user_id":987654,"site_id":"MLB","logistic_type":"fulfillment"}]' WHERE id=$1`, c.store); err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, resource); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("changed marketing mapping committed: %v", err)
	}
}

func TestNotificationSharesManualOrderLimit(t *testing.T) {
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	ctx := context.Background()
	job := c.job("marketplace_orders", "10001", "987654")
	_, err := c.prepare(t, job)
	var limited *notifications.RateLimited
	if !errors.As(err, &limited) || limited.RetryAfter < 119*time.Second {
		t.Fatalf("notification 429 not translated: %v", err)
	}
	before := len(c.calls())
	var raw map[string]any
	if _, err := c.service.orderGET(ctx, "synthetic_old_access", "/marketplace/orders/10002", &raw); err == nil {
		t.Fatal("manual read bypassed the notification cooldown")
	}
	if len(c.calls()) != before {
		t.Fatal("manual read reached the provider during cooldown")
	}

	// The reverse direction: a manual 429 must block the worker before its read.
	c.service.orderRetryUntil = time.Time{}
	if _, err := c.service.orderGET(ctx, "synthetic_old_access", "/marketplace/orders/10002", &raw); err == nil {
		t.Fatal("expected manual upstream 429")
	}
	before = len(c.calls())
	_, err = c.prepare(t, job)
	if !errors.As(err, &limited) || limited.RetryAfter < 119*time.Second {
		t.Fatalf("manual cooldown did not reach the worker: %v", err)
	}
	if len(c.calls()) != before {
		t.Fatal("worker read reached the provider during the manual cooldown")
	}

	// Both paths share one in-process mutex; neither is prioritized today.
	c.service.orderRetryUntil = time.Time{}
	c.service.orderRequests.Lock()
	defer c.service.orderRequests.Unlock()
	if _, err := c.prepare(t, job); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("worker did not report local contention: %v", err)
	}
	var target map[string]any
	if _, err := c.service.orderGET(ctx, "synthetic_old_access", "/marketplace/orders/10002", &target); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("manual read did not report local contention: %v", err)
	}
}
