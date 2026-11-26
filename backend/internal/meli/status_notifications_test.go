package meli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

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

func TestClaimActionScopeAndMandatoryDeadline(t *testing.T) {
	auth := AuthorizedStore{SellerID: "1234", Verified: true, Markets: []Marketplace{
		{UserID: "987654", SiteID: "MLB", LogisticType: "remote"},
		{UserID: "987655", SiteID: "MLB", LogisticType: "remote"},
	}}
	job := notifications.Job{Topic: "marketplace_claims", ResourceID: "123", NotificationUserID: "987654"}
	// The seller is the complainant. The buyer respondent has different actions.
	payload := func(other string, sellerActions string) json.RawMessage {
		return json.RawMessage(fmt.Sprintf(`{"id":123,"resource_id":700,"resource":"shipment",
"status":"opened","stage":"claim","type":"mediations","reason_id":"PNR9503","site_id":"MLB",
"last_updated":"2026-11-22T00:00:00Z","players":[
{"type":"sender","role":"complainant","user_id":987654,"available_actions":%s},
{"type":"receiver","role":"respondent","user_id":%s,"available_actions":[{"action":"buyer_action","mandatory":true,"due_date":"2026-11-22T01:00:00Z"}]}]}`, sellerActions, other))
	}
	actions := `[{"action":"send_message_to_mediator","mandatory":false,"due_date":"2026-11-22T02:00:00Z"},
{"action":"add_shipping_evidence","mandatory":true,"due_date":"2026-11-23T09:00:00Z"}]`
	n, err := decodeMarketplaceStatusNotice(payload("111", actions), job, auth, time.Now())
	if err != nil || n.actions != "add_shipping_evidence,send_message_to_mediator" ||
		n.dueAt == nil || n.dueAt.Format(time.RFC3339) != "2026-11-23T09:00:00Z" {
		t.Fatalf("wrong seller actions or mandatory deadline: %+v %v", n, err)
	}
	n, err = decodeMarketplaceStatusNotice(payload("111", `[{"action":"send_message_to_mediator","mandatory":false,"due_date":"2026-11-22T02:00:00Z"}]`), job, auth, time.Now())
	if err != nil || n.dueAt != nil {
		t.Fatalf("optional action reported as mandatory: %+v %v", n, err)
	}
	if _, err := decodeMarketplaceStatusNotice(payload("987655", actions), job, auth, time.Now()); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("notification user disambiguated two mapped participants: %v", err)
	}
	if _, err := decodeMarketplaceStatusNotice(payload("111", `[{"action":"refund","mandatory":true,"due_date":"bad"}]`), job, auth, time.Now()); !errors.Is(err, ErrOrderInvalidResponse) {
		t.Fatalf("invalid mandatory deadline accepted: %v", err)
	}
}

func TestItemNotificationOptionalFieldsAndStaleSnapshotPostgres(t *testing.T) {
	var version atomic.Int32
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/marketplace/items/MLB123" {
			http.NotFound(w, r)
			return
		}
		switch version.Load() {
		case 0:
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"paused",
"sub_status":["under_review"],"price":29.87,"available_quantity":45,"last_updated":"2026-11-22T00:00:00Z"}`)
		case 1:
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"paused",
"last_updated":"2026-11-23T00:00:00Z"}`)
		case 2:
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"active",
"sub_status":[],"price":0,"available_quantity":0,"last_updated":"2026-11-24T00:00:00Z"}`)
		case 3:
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"paused",
"sub_status":["under_review"],"price":99,"available_quantity":99,"last_updated":"2026-11-21T00:00:00Z"}`)
		default:
			fmt.Fprint(w, `{"id":"MLB123","site_id":"MLB","seller_id":987654,"status":"paused",
"sub_status":["aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","cccccccccccccccccccccccccccccccccccccccc","dddddddddddddddddddddddddddddddddddddddd"],"last_updated":"2026-11-25T00:00:00Z"}`)
		}
	})
	job := c.job("marketplace_items", "MLB123", "987654")
	apply := func(v int32) {
		t.Helper()
		version.Store(v)
		resource, err := c.prepare(t, job)
		if err != nil {
			t.Fatal(err)
		}
		if err := c.apply(t, resource); err != nil {
			t.Fatal(err)
		}
	}
	apply(0)
	apply(1)
	if got := c.count(t, `SELECT count(*) FROM app.meli_item_notifications WHERE store_id=$1
AND item_id='MLB123' AND sub_status='under_review' AND price=29.87 AND available_quantity=45
AND last_updated='2026-11-23T00:00:00Z'::timestamptz`, c.store); got != 1 {
		t.Fatal("omitted fields erased existing values")
	}
	apply(2)
	apply(3)
	if got := c.count(t, `SELECT count(*) FROM app.meli_item_notifications WHERE store_id=$1
AND item_id='MLB123' AND status='active' AND sub_status='' AND price=0 AND available_quantity=0
AND last_updated='2026-11-24T00:00:00Z'::timestamptz`, c.store); got != 1 {
		t.Fatal("explicit clear/zero or stale-update protection failed")
	}
	version.Store(4)
	if _, err := c.prepare(t, job); !errors.Is(err, ErrOrderInvalidResponse) {
		t.Fatalf("oversized sub_status reached database write path: %v", err)
	}
}
