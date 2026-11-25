package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
	"github.com/jackc/pgx/v5"
)

func notifiedOrderFixture(id, updated, status, shipment string) string {
	return fmt.Sprintf(`{"id":%s,"seller":{"id":987654},"site_id":"MLB","pack_id":700,
"status":%q,"currency_id":"USD","total_amount":29.99,"shipping":{"id":%s},
"date_created":"2026-11-20T00:00:00Z","last_updated":%q,
"order_items":[{"item":{"id":"MLB100","title":"Item"},"quantity":1,"unit_price":29.99,"currency_id":"USD"}],
"buyer":{"email":"private@example.invalid"}}`, id, status, shipment, updated)
}

func TestNotificationOrderScopeRejectsAmbiguity(t *testing.T) {
	auth := AuthorizedStore{SellerID: "123", Verified: true, Markets: []Marketplace{
		{UserID: "987654", SiteID: "MLB", LogisticType: "remote"},
	}}
	raw := json.RawMessage(notifiedOrderFixture("10001", "2026-11-21T01:00:00Z", "paid", "801"))
	got, err := resolveNotifiedOrder(raw, "10001", "123", auth)
	if err != nil || got.scope != (OrderScope{Site: "MLB", Logistic: "remote"}) || got.seller != "987654" {
		t.Fatalf("parent scope: %+v %v", got, err)
	}
	if _, err := resolveNotifiedOrder(raw, "10001", "987655", auth); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("other child accepted: %v", err)
	}
	if _, err := resolveNotifiedOrder(raw, "700", "123", auth); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("pack ID accepted as order: %v", err)
	}
	auth.Markets = append(auth.Markets, Marketplace{UserID: "987654", SiteID: "MLB", LogisticType: "fulfillment"})
	if _, err := resolveNotifiedOrder(raw, "10001", "123", auth); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("ambiguous logistic mode accepted: %v", err)
	}
}

func TestNotificationShipmentFirstAndOrderReplayPostgres(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var parent string
	if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour'
WHERE id=$1 RETURNING seller_id`, store).Scan(&parent); err != nil {
		t.Fatal(err)
	}
	var old atomic.Bool
	var limited atomic.Bool
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("notification must use scoped official GET")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if limited.Load() {
			w.Header().Set("Retry-After", "120")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		switch r.URL.Path {
		case "/marketplace/shipments/801":
			if r.Header.Get("x-format-new") != "true" {
				t.Error("missing shipment format header")
			}
			fmt.Fprint(w, `{"id":801,"sender_id":987654,"site_id":"MLB","status":"delivered","tracking_number":"TRACK801","logistic":{"type":"drop_off"},"last_updated":"2026-11-22T00:00:00Z"}`)
		case "/shipments/801/orders":
			if r.Header.Get("X-New-Domain") != "true" {
				t.Error("missing hosted association header")
			}
			fmt.Fprint(w, `[{"order_id":"10001","pack_id":"700","seller_id":987654},{"order_id":"10001","pack_id":"700","seller_id":987654},{"order_id":"10002","pack_id":"700","seller_id":987654}]`)
		case "/marketplace/orders/10001", "/marketplace/orders/10002":
			id := r.URL.Path[len("/marketplace/orders/"):]
			if old.Load() {
				fmt.Fprint(w, notifiedOrderFixture(id, "2026-11-20T00:01:00Z", "cancelled", "802"))
			} else {
				fmt.Fprint(w, notifiedOrderFixture(id, "2026-11-21T01:00:00Z", "paid", "801"))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	handlers := s.NotificationHandlers()
	job := notifications.Job{StoreID: store, OwnerID: owner, ParentSellerID: parent,
		NotificationUserID: "987654", Topic: "marketplace_shipments", ResourceID: "801"}
	resource, err := handlers[job.Topic](ctx, job)
	if err != nil || resource.SellerID != "987654" || resource.ResourceID != "801" {
		t.Fatalf("shipment prepare: %+v %v", resource, err)
	}
	apply := func(resource notifications.VerifiedResource) error {
		return pgx.BeginFunc(ctx, db, func(tx pgx.Tx) error {
			return resource.Apply(ctx, tx)
		})
	}
	var count int
	if err := db.QueryRow(ctx, "SELECT count(*) FROM app.meli_orders WHERE store_id=$1", store).Scan(&count); err != nil || count != 0 {
		t.Fatalf("prepare wrote orders before fenced apply: %d %v", count, err)
	}
	if err := apply(resource); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1 AND site_id='MLB' AND logistic_type='remote'
AND seller_id='987654' AND pack_id='700' AND shipment_id='801' AND shipment_snapshot->>'status'='delivered'`, store).Scan(&count); err != nil || count != 2 {
		t.Fatalf("shipment-first did not create both orders: %d %v", count, err)
	}
	if calls.Load() != 4 {
		t.Fatalf("duplicate association caused extra official calls: %d", calls.Load())
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_order_purchases
(store_id,site_id,logistic_type,seller_id,order_id,purchase_order_id,cost_cny)
VALUES($1,'MLB','remote','987654','10001','12345678',15.50)`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET billing_snapshot='{"sentinel":true}'::jsonb
WHERE store_id=$1 AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	old.Store(true)
	job.Topic, job.ResourceID = "marketplace_orders", "10001"
	resource, err = handlers[job.Topic](ctx, job)
	if err != nil {
		t.Fatal(err)
	}
	if err := apply(resource); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `SELECT count(*) FROM app.meli_orders o JOIN app.meli_order_purchases p
USING(store_id,site_id,logistic_type,seller_id,order_id)
WHERE o.store_id=$1 AND o.order_id='10001' AND o.status='paid' AND o.shipment_id='801'
AND o.shipment_snapshot->>'status'='delivered' AND o.billing_snapshot->>'sentinel'='true'
AND p.purchase_order_id='12345678'`, store).Scan(&count); err != nil || count != 1 {
		t.Fatalf("old order replay changed protected local data: %d %v", count, err)
	}
	if err := db.QueryRow(ctx, "SELECT count(*) FROM app.meli_order_syncs WHERE store_id=$1", store).Scan(&count); err != nil || count != 0 {
		t.Fatalf("notification advanced manual cursor: %d %v", count, err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces='[{"user_id":987654,"site_id":"MLB","logistic_type":"fulfillment"}]' WHERE id=$1`, store); err != nil {
		t.Fatal(err)
	}
	if err := apply(resource); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("changed logistics scope accepted: %v", err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces='[{"user_id":987654,"site_id":"MLB","logistic_type":"remote"}]' WHERE id=$1`, store); err != nil {
		t.Fatal(err)
	}
	limited.Store(true)
	if _, err := handlers[job.Topic](ctx, job); err == nil {
		t.Fatal("expected upstream rate limit")
	} else {
		var rate *notifications.RateLimited
		if !errors.As(err, &rate) || rate.RetryAfter < 119*time.Second {
			t.Fatalf("notification rate limit not translated: %v", err)
		}
	}
	before := calls.Load()
	var raw json.RawMessage
	if _, err := s.orderGET(ctx, "synthetic_old_access", "/marketplace/orders/10002", &raw); err == nil || calls.Load() != before {
		t.Fatal("manual official read bypassed shared cooldown")
	}
}
