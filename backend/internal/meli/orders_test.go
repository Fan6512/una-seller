package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestOrderSearchUsesNestedOrderIDsNotPackIDs(t *testing.T) {
	input := []byte(`{"results":[
		{"id":900,"orders":[{"id":101,"seller":{"id":456}},{"id":102,"seller":{"id":456}}]},
		{"id":901,"orders":[{"id":101,"seller":{"id":456}}]}
	],"paging":{"total":2,"limit":3}}`)
	ids, total, count, err := parseOrderSearch(input, "456", 0)
	if err != nil || total != 2 || count != 2 || len(ids) != 2 || ids[0] != "101" || ids[1] != "102" {
		t.Fatalf("ids=%v total=%d count=%d err=%v", ids, total, count, err)
	}
}

func TestOrderSearchRejectsMissingOrWrongSeller(t *testing.T) {
	for _, input := range []string{
		`{"results":[{"id":900,"orders":[]}],"paging":{"total":1}}`,
		`{"results":[{"id":900,"orders":[{"id":101,"seller":{"id":999}}]}],"paging":{"total":1}}`,
		`{"results":[{"id":900,"orders":[{"id":101,"seller":{"id":456}}]}]}`,
		`{"results":[],"paging":{"total":4}}`,
	} {
		if _, _, _, err := parseOrderSearch([]byte(input), "456", 0); err == nil {
			t.Errorf("accepted unsafe search: %s", input)
		}
	}
}

func TestOrderDetailSanitizesBuyerFieldsAndValidatesIdentity(t *testing.T) {
	const input = `{
		"id":101,"pack_id":900,"seller":{"id":456},"site_id":"MLB",
		"status":"paid","currency_id":"BRL","total_amount":123.45,
		"date_created":"2026-11-19T12:00:00Z","last_updated":"2026-11-19T12:01:00Z",
		"buyer":{"phone":"secret","email":"sensitive@example.com"},
		"shipping":{"id":801,"receiver_address":{"street_name":"private"}},
		"order_items":[{"item":{"id":"MLB1","title":"Product","variation_id":1234},
			"quantity":2,"unit_price":61.725,"currency_id":"BRL"}]
	}`
	record, err := parseOrderDetail([]byte(input), "101", "456", "MLB")
	if err != nil || record.ID != "101" || record.PackID != "900" ||
		record.ShipmentID != "801" || record.TotalAmount != "123.45" ||
		len(record.Items) != 1 || record.Items[0].VariationID != "1234" {
		t.Fatalf("parsed=%+v err=%v", record, err)
	}
	raw, _ := json.Marshal(record)
	for _, secret := range []string{"phone", "email", "receiver_address", "sensitive@example.com"} {
		if json.Valid(raw) && containsBytes(raw, []byte(secret)) {
			t.Errorf("PII persisted into order response: %s", secret)
		}
	}
	for _, test := range []struct{ order, seller, site string }{
		{"900", "456", "MLB"}, {"101", "999", "MLB"}, {"101", "456", "MLM"},
	} {
		if _, err = parseOrderDetail([]byte(input), test.order, test.seller, test.site); err == nil {
			t.Errorf("accepted mismatched order/seller/site: %+v", test)
		}
	}
}

func containsBytes(data, fragment []byte) bool {
	for i := 0; i+len(fragment) <= len(data); i++ {
		if string(data[i:i+len(fragment)]) == string(fragment) {
			return true
		}
	}
	return false
}

func TestOrderScopeAndStatusValidation(t *testing.T) {
	for _, scope := range []OrderScope{{"CBT", ""}, {"MLB", ""}, {"mlb", "remote"}, {"MLB", "remote/sub"}} {
		if orderScopeValid(scope) {
			t.Fatalf("accepted unsupported scope: %+v", scope)
		}
	}
	if !orderScopeValid(OrderScope{Site: "MLB", Logistic: "fulfillment"}) {
		t.Fatal("verified marketplace scope rejected")
	}
	if !ValidOrderStatus("paid") || !ValidOrderStatus("") || !ValidOrderStatus("payment_required") || !ValidOrderStatus("code2") ||
		ValidOrderStatus("paid;DROP TABLE") {
		t.Fatal("unexpected order status validation")
	}
}

func TestOrderDetailRequiresConsistentSite(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields string
		valid  bool
	}{
		{"CBT context only", `,"context":{"site":"MLB"}`, true},
		{"top level only", `,"site_id":"MLB"`, true},
		{"both agree", `,"site_id":"MLB","context":{"site":"MLB"}`, true},
		{"wrong CBT context", `,"context":{"site":"MLM"}`, false},
		{"wrong top level", `,"site_id":"MLM","context":{"site":"MLB"}`, false},
		{"conflicting context", `,"site_id":"MLB","context":{"site":"MLM"}`, false},
		{"neither present", "", false},
		{"context without site", `,"context":{"flows":["cbt"]}`, false},
		{"null site", `,"site_id":null,"context":{"site":"MLB"}`, false},
		{"empty context site", `,"site_id":"MLB","context":{"site":""}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const base = `{"id":101,"seller":{"id":456},"status":"paid","currency_id":"USD","total_amount":12.34,"date_created":"2026-11-01T00:00:00Z","last_updated":"2026-11-01T01:00:00Z"`
			_, err := parseOrderDetail([]byte(base+tc.fields+"}"), "101", "456", "MLB")
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v, error=%v", tc.valid, err)
			}
		})
	}
}

func TestOrderCancellationAndTimeoutAreDistinct(t *testing.T) {
	if OrderErrorCode(context.Canceled) != "order_canceled" || OrderErrorCode(context.DeadlineExceeded) != "order_timeout" {
		t.Fatal("expected distinct cancellation and deadline error codes")
	}
}

func TestOrderPostgresPackResumeAndLocalRead(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, "UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid", store); err != nil {
		t.Fatal(err)
	}
	var searches, details, unexpected atomic.Int32
	var slow atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/marketplace/orders/search":
			searches.Add(1)
			switch r.URL.Query().Get("offset") {
			case "0":
				fmt.Fprint(w, `{"results":[{"orders":[{"id":101},{"id":102},{"id":103},{"id":104},{"id":105},{"id":106},{"id":107}]}],"paging":{"total":2}}`)
			case "1":
				fmt.Fprint(w, `{"results":[{"orders":[{"id":108}]}],"paging":{"total":2}}`)
			default:
				t.Error("unexpected search offset")
				w.WriteHeader(400)
			}
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/marketplace/orders/"):
			details.Add(1)
			id := strings.TrimPrefix(r.URL.Path, "/marketplace/orders/")
			if id == "105" && slow.Load() {
				<-r.Context().Done()
				return
			}
			fmt.Fprint(w, orderTestDetail(id, "987654", "MLB"))
		default:
			unexpected.Add(1)
			w.WriteHeader(400)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}

	assertState := func(wantRows, wantOffset int, wantPending string, completed bool) {
		t.Helper()
		var rows, offset int
		var pending string
		var done, leased bool
		err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM app.meli_orders WHERE store_id=$1::uuid),
 next_offset,array_to_string(pending_order_ids,','),last_completed_to IS NOT NULL,lease_id IS NOT NULL
 FROM app.meli_order_syncs WHERE store_id=$1::uuid`, store).Scan(&rows, &offset, &pending, &done, &leased)
		if err != nil || rows != wantRows || offset != wantOffset || pending != wantPending || done != completed || leased {
			t.Fatalf("rows=%d offset=%d pending=%q completed=%t leased=%t error=%v", rows, offset, pending, done, leased, err)
		}
	}

	got, err := s.SyncOrders(ctx, owner, store, scope)
	if err != nil || got != (OrderSyncResult{Saved: 3, HasMore: true}) || searches.Load() != 1 || details.Load() != 3 {
		t.Fatalf("first batch=%+v searches=%d details=%d error=%v", got, searches.Load(), details.Load(), err)
	}
	assertState(3, 0, "104,105,106,107", false)

	slow.Store(true)
	timeout, stop := context.WithTimeout(ctx, 250*time.Millisecond)
	_, err = s.SyncOrders(timeout, owner, store, scope)
	stop()
	slow.Store(false)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected detail timeout, got %v", err)
	}
	assertState(3, 0, "104,105,106,107", false)

	const dropCheck = "ALTER TABLE app.meli_orders DROP CONSTRAINT IF EXISTS test_order_pack_rollback"
	if _, err = db.Exec(ctx, "ALTER TABLE app.meli_orders ADD CONSTRAINT test_order_pack_rollback CHECK (order_id <> '105') NOT VALID"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := db.Exec(c, dropCheck); err != nil {
			t.Error(err)
		}
	})
	_, err = s.SyncOrders(ctx, owner, store, scope)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.ConstraintName != "test_order_pack_rollback" {
		t.Fatalf("expected second insert to abort batch, got %v", err)
	}
	assertState(3, 0, "104,105,106,107", false)
	if _, err = db.Exec(ctx, dropCheck); err != nil {
		t.Fatal(err)
	}

	slow.Store(true)
	clientTimeout := s.client.Timeout
	s.client.Timeout = 250 * time.Millisecond
	got, err = s.SyncOrders(ctx, owner, store, scope)
	s.client.Timeout = clientTimeout
	slow.Store(false)
	if err != nil || got != (OrderSyncResult{Saved: 1, HasMore: true}) || searches.Load() != 1 {
		t.Fatalf("provider timeout must commit the fetched prefix, got=%+v searches=%d error=%v", got, searches.Load(), err)
	}
	assertState(4, 0, "105,106,107", false)
	got, err = s.SyncOrders(ctx, owner, store, scope)
	if err != nil || got != (OrderSyncResult{Saved: 3, Processed: 1, HasMore: true, NextOffset: 1}) || searches.Load() != 1 {
		t.Fatalf("completed pack=%+v searches=%d error=%v", got, searches.Load(), err)
	}
	assertState(7, 1, "", false)
	got, err = s.SyncOrders(ctx, owner, store, scope)
	if err != nil || got != (OrderSyncResult{Saved: 1, Processed: 1, NextOffset: 2}) || searches.Load() != 2 {
		t.Fatalf("next page=%+v searches=%d error=%v", got, searches.Load(), err)
	}
	assertState(8, 0, "", true)

	for _, status := range []string{"connected", "reauth_required", "refresh_unknown"} {
		if _, err = db.Exec(ctx, "UPDATE app.meli_store_connections SET status=$2,access_token_expires_at=now()-interval '1 day' WHERE id=$1::uuid", store, status); err != nil {
			t.Fatal(err)
		}
		page, err := s.ListOrders(ctx, owner, store, scope, "payment_required", 0)
		if err != nil || len(page.Items) != 8 {
			t.Fatalf("cached list for %s has %d results, error=%v", status, len(page.Items), err)
		}
		rec, err := s.GetOrder(ctx, owner, store, scope, "105")
		if err != nil || rec.TotalAmount != "12.34" || rec.Status != "payment_required" {
			t.Fatalf("cached detail amount=%s status=%s error=%v", rec.TotalAmount, rec.Status, err)
		}
		var actual string
		if err = db.QueryRow(ctx, "SELECT status FROM app.meli_store_connections WHERE id=$1::uuid", store).Scan(&actual); err != nil || actual != status {
			t.Fatalf("cached read changed authorization state to %s, error=%v", actual, err)
		}
	}
	if unexpected.Load() != 0 {
		t.Fatal("cached read attempted an upstream request")
	}
	if err = s.Disconnect(ctx, owner, store); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ListOrders(ctx, owner, store, scope, "", 0); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("disconnected store list should be denied, got %v", err)
	}
	if _, err = s.GetOrder(ctx, owner, store, scope, "105"); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("disconnected store detail should be denied, got %v", err)
	}
}

func TestOrderPostgresResetsCursorForNewSeller(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	const oldMarket = `[{"user_id":987654,"site_id":"MPE","logistic_type":"remote"}]`
	if _, err := db.Exec(ctx, "UPDATE app.meli_store_connections SET marketplaces=$2::jsonb,access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid", store, oldMarket); err != nil {
		t.Fatal(err)
	}
	scope := OrderScope{Site: "MPE", Logistic: "remote"}
	var newSearches atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/marketplace/orders/search" {
			switch r.URL.Query().Get("seller.id") {
			case "987654":
				fmt.Fprint(w, `{"results":[{"orders":[{"id":201},{"id":202},{"id":203},{"id":204}]}],"paging":{"total":1}}`)
			case "987655":
				newSearches.Add(1)
				from, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("last_updated.from"))
				if err != nil || time.Since(from) < 29*24*time.Hour || r.URL.Query().Get("offset") != "0" {
					t.Error("new seller inherited the old search window or offset")
				}
				fmt.Fprint(w, `{"results":[{"orders":[{"id":201}]}],"paging":{"total":1}}`)
			default:
				t.Error("unexpected seller")
				w.WriteHeader(400)
			}
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/marketplace/orders/")
		if newSearches.Load() != 0 {
			// The same order ID must become readable even if the previous seller's version was newer.
			fmt.Fprint(w, strings.ReplaceAll(orderTestDetail(id, "987655", "MPE"), "2026-11-01", "2026-10-31"))
			return
		}
		fmt.Fprint(w, orderTestDetail(id, "987654", "MPE"))
	}))
	defer server.Close()
	s.apiURL = server.URL
	first, err := s.SyncOrders(ctx, owner, store, scope)
	if err != nil || first.Saved != 3 || !first.HasMore {
		t.Fatalf("old seller batch=%+v error=%v", first, err)
	}
	old, err := s.acquireOrderCursor(ctx, owner, store, scope, "987654")
	if err != nil || len(old.PendingIDs) != 1 {
		t.Fatalf("old seller pending absent, error=%v", err)
	}
	if _, err = db.Exec(ctx, "UPDATE app.meli_order_syncs SET next_offset=7,last_completed_to=now()-interval '1 hour',window_from=now()-interval '2 hours' WHERE store_id=$1::uuid", store); err != nil {
		t.Fatal(err)
	}
	const newMarket = `[{"user_id":987655,"site_id":"MPE","logistic_type":"remote"},{"user_id":987655,"site_id":"MLB","logistic_type":"remote"}]`
	if _, err = db.Exec(ctx, "UPDATE app.meli_store_connections SET marketplaces=$2::jsonb WHERE id=$1::uuid", store, newMarket); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SyncOrders(ctx, owner, store, scope); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("new seller bypassed an active lease, got %v", err)
	}
	s.releaseOrderCursor(store, scope, old.LeaseID)
	if _, err = s.acquireOrderCursor(ctx, owner, store, scope, "987654"); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("stale seller reacquired cursor, got %v", err)
	}
	if _, err = s.GetOrder(ctx, owner, store, scope, "202"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("old seller order leaked, got %v", err)
	}
	result, err := s.SyncOrders(ctx, owner, store, scope)
	if err != nil || result.Saved != 1 || result.HasMore || newSearches.Load() != 1 {
		t.Fatalf("new seller batch=%+v error=%v", result, err)
	}
	page, err := s.ListOrders(ctx, owner, store, scope, "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "201" || page.Items[0].LastUpdated.Day() != 31 {
		t.Fatalf("unexpected new seller results, rows=%d error=%v", len(page.Items), err)
	}
	if _, err = s.GetOrder(ctx, owner, store, OrderScope{Site: "MLB", Logistic: "remote"}, "201"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("order leaked to another site, got %v", err)
	}
	if _, err = s.GetOrder(ctx, owner, store, OrderScope{Site: "MPE", Logistic: "fulfillment"}, "201"); !errors.Is(err, ErrSiteNotAuthorized) {
		t.Fatalf("order leaked to another logistics mode, got %v", err)
	}
}

func orderTestDetail(id, seller, site string) string {
	return fmt.Sprintf(`{"id":%s,"seller":{"id":%s},"context":{"site":%q},"status":"payment_required","currency_id":"USD","total_amount":12.3400,"date_created":"2026-11-01T00:00:00Z","last_updated":"2026-11-01T01:00:00Z"}`, id, seller, site)
}
