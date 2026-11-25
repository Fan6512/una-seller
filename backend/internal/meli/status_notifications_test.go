package meli

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
)

func TestMarketplaceStatusNotificationsPostgres(t *testing.T) {
	var closed atomic.Bool
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		status, updated := "opened", "2026-11-21T00:00:00Z"
		if closed.Load() {
			status, updated = "closed", "2026-11-22T00:00:00Z"
		}
		switch r.URL.Path {
		case "/marketplace/v2/claims/123":
			fmt.Fprintf(w, `{"id":123,"resource_id":700,"resource":"shipment","status":%q,
"stage":"dispute","type":"mediations","reason_id":"PDD9939","site_id":"MLB","last_updated":%q,
"players":[{"type":"sender","role":"respondent","user_id":987654,"available_actions":[
{"action":"refund","mandatory":true,"due_date":"2026-11-23T09:00:00.000Z"},
{"action":"allow_return","mandatory":false,"due_date":"2026-11-24T09:00:00.000Z"}]},
{"type":"receiver","role":"complainant","user_id":111}]}`, status, updated)
		case "/marketplace/items/MLB123":
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"paused","sub_status":["under_review"],
"price":29.87,"available_quantity":0,"last_updated":"2026-11-22T00:00:00Z"}`)
		case "/marketplace/items/MLM123":
			fmt.Fprint(w, `{"id":"MLM123","site_id":"MLM","seller_id":987654,"status":"active","sub_status":[],
"price":10.5,"available_quantity":3,"last_updated":"2026-11-22T00:00:00Z"}`)
		default:
			http.NotFound(w, r)
		}
	})
	job := c.job("marketplace_claims", "123", "987654")
	first, err := c.prepare(t, job)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, first); err != nil {
		t.Fatal(err)
	}
	closed.Store(true)
	newer, err := c.prepare(t, job)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, newer); err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, first); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_claim_notifications WHERE store_id=$1
AND claim_id='123' AND resource_type='shipment' AND related_id='700' AND status='closed'
AND due_at='2026-11-23T09:00:00Z'::timestamptz AND allowed_actions='allow_return,refund'`, c.store); got != 1 {
		t.Fatal("claim status, due date or allowed actions missing")
	}
	item, err := c.prepare(t, c.job("marketplace_items", "MLB123", "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, item); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_item_notifications WHERE store_id=$1
AND item_id='MLB123' AND status='paused' AND sub_status='under_review'
AND price=29.87 AND available_quantity=0`, c.store); got != 1 {
		t.Fatal("item projection missing price or review status")
	}
	if _, err := c.prepare(t, c.job("marketplace_items", "MLM123", "987654")); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("wrong site accepted: %v", err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_published_listings WHERE store_id=$1`, c.store); got != 0 {
		t.Fatal("notification overwrote listing cache")
	}
	if _, err := c.db.Exec(t.Context(), `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, c.store); err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, item); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("revoked store accepted: %v", err)
	}
}
