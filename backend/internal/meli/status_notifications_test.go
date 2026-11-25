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
			fmt.Fprintf(w, `{"id":123,"resource_id":700,"resource":"order","status":%q,
"stage":"dispute","type":"mediations","reason_id":"PDD9939","site_id":"MLB","last_updated":%q,
"players":[{"type":"seller","role":"respondent","user_id":987654},{"type":"buyer","role":"complainant","user_id":111}]}`, status, updated)
		case "/marketplace/items/MLB123":
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"paused",
"available_quantity":0,"last_updated":"2026-11-22T00:00:00Z"}`)
		case "/marketplace/items/MLM123":
			fmt.Fprint(w, `{"id":"MLM123","site_id":"MLM","seller_id":987654,"status":"active",
"available_quantity":3,"last_updated":"2026-11-22T00:00:00Z"}`)
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
AND claim_id='123' AND resource_type='order' AND related_id='700' AND status='closed'`, c.store); got != 1 {
		t.Fatal("claim status regressed")
	}
	item, err := c.prepare(t, c.job("marketplace_items", "MLB123", "987654"))
	if err != nil {
		t.Fatal(err)
	}
	if err := c.apply(t, item); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_item_notifications WHERE store_id=$1
AND item_id='MLB123' AND status='paused' AND available_quantity=0`, c.store); got != 1 {
		t.Fatal("item projection missing")
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
