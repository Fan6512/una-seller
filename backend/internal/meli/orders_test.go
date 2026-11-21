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
	for _, scope := range []OrderScope{{"CBT", ""}, {"MLB", ""}, {"mlb", "remote"}, {"MLB", "remote/sub"}, {"ML1", "remote"},
		{"ML A", "remote"}, {"M!A", "remote"}, {"MÉX", "remote"}, {"AA", "remote"}, {"ABCDEFGHI", "remote"}} {
		if orderScopeValid(scope) {
			t.Fatalf("accepted unsupported scope: %+v", scope)
		}
	}
	if !orderScopeValid(OrderScope{Site: "MPE", Logistic: "fulfillment"}) ||
		!orderScopeValid(OrderScope{Site: "SITEABCD", Logistic: "remote"}) {
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
		{"null top level", `,"site_id":null,"context":{"site":"MLB"}`, true},
		{"null context site", `,"site_id":"MLB","context":{"site":null}`, true},
		{"both null", `,"site_id":null,"context":{"site":null}`, false},
		{"invalid top level type", `,"site_id":123,"context":{"site":"MLB"}`, false},
		{"null top level, wrong context", `,"site_id":null,"context":{"site":"MLM"}`, false},
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

func TestOrderLocalRequestContentionIsNotUpstreamRateLimit(t *testing.T) {
	s := Service{}
	s.orderRequests.Lock()
	defer s.orderRequests.Unlock()
	var raw json.RawMessage
	status, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/search", &raw)
	if status != http.StatusConflict || !errors.Is(err, ErrOrderBusy) ||
		OrderErrorCode(err) != "order_busy" {
		t.Fatalf("local contention incorrectly reported as provider 429: status=%d err=%v", status, err)
	}
}

func TestOrderProviderPacing(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, `{}`)
	}))
	defer server.Close()
	s := Service{apiURL: server.URL, client: server.Client()}
	var raw json.RawMessage
	if _, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/search", &raw); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	next := s.orderNextRequest
	// Reserve a known future slot even on a loaded test runner.
	s.orderNextRequest = started.Add(orderRequestInterval)
	if next.IsZero() {
		t.Fatal("provider request did not reserve the next slot")
	}
	if _, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/101", &raw); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 || time.Since(started) < orderRequestInterval {
		t.Fatal("second provider request was not paced")
	}
}

// The search endpoint has a separate 100/minute quota. Detail GETs remain
// on the faster shared pacing and must not consume/reserve a search slot.
func TestOrderSearchPacingRespectsQuota(t *testing.T) {
	requests := make(chan time.Time, 3)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests <- time.Now()
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()
	s := Service{apiURL: server.URL, client: server.Client()}
	var raw json.RawMessage

	if _, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/search?seller.id=1", &raw); err != nil {
		t.Fatal(err)
	}
	first := <-requests
	nextSearch := s.orderNextSearch
	if nextSearch.IsZero() {
		t.Fatal("search did not reserve a quota slot")
	}
	if _, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/101", &raw); err != nil {
		t.Fatal(err)
	}
	detail := <-requests
	if !s.orderNextSearch.Equal(nextSearch) {
		t.Fatal("detail request consumed a search quota slot")
	}
	if detail.Sub(first) < orderRequestInterval-20*time.Millisecond {
		t.Fatal("detail request bypassed general pacing")
	}
	if _, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/search", &raw); err != nil {
		t.Fatal(err)
	}
	second := <-requests
	if elapsed := second.Sub(first); elapsed < orderSearchRequestInterval-20*time.Millisecond {
		t.Fatalf("search requests too close: %s, want at least %s", elapsed, orderSearchRequestInterval)
	}
}

func TestOrderSearchCanceledWhileThrottledDoesNotSendRequest(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()
	s := Service{apiURL: server.URL, client: server.Client()}
	var raw json.RawMessage
	if _, err := s.orderGET(t.Context(), "fixture", "/marketplace/orders/search", &raw); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := s.orderGET(ctx, "fixture", "/marketplace/orders/search", &raw); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled search returned %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("canceled search sent %d upstream requests", calls.Load())
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
	if err != nil || got != (OrderSyncResult{Saved: 1, Processed: 1, NextOffset: 0}) || searches.Load() != 2 {
		t.Fatalf("next page=%+v searches=%d error=%v", got, searches.Load(), err)
	}
	assertState(8, 0, "", true)

	for _, status := range []string{"connected", "reauth_required", "refresh_unknown"} {
		if _, err = db.Exec(ctx, "UPDATE app.meli_store_connections SET status=$2,access_token_expires_at=now()-interval '1 day' WHERE id=$1::uuid", store, status); err != nil {
			t.Fatal(err)
		}
		page, err := s.ListOrders(ctx, owner, store, scope, "payment_required", "", 0)
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
	if _, err = s.ListOrders(ctx, owner, store, scope, "", "", 0); !errors.Is(err, ErrReauthorizationRequired) {
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
			// An overlapping order ID belongs to each seller's independent history.
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
	page, err := s.ListOrders(ctx, owner, store, scope, "", "", 0)
	if err != nil || len(page.Items) != 1 || page.Items[0].ID != "201" || page.Items[0].LastUpdated.Day() != 31 {
		t.Fatalf("unexpected new seller results, rows=%d error=%v", len(page.Items), err)
	}
	var retained int
	if err = db.QueryRow(ctx, "SELECT count(*) FROM app.meli_orders WHERE store_id=$1::uuid AND seller_id='987654'", store).Scan(&retained); err != nil || retained != 3 {
		t.Fatalf("old seller history count=%d error=%v", retained, err)
	}
	if _, err = s.GetOrder(ctx, owner, store, OrderScope{Site: "MLB", Logistic: "remote"}, "201"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("order leaked to another site, got %v", err)
	}
	if _, err = s.GetOrder(ctx, owner, store, OrderScope{Site: "MPE", Logistic: "fulfillment"}, "201"); !errors.Is(err, ErrSiteNotAuthorized) {
		t.Fatalf("order leaked to another logistics mode, got %v", err)
	}
}

func TestOrderPostgresResetCursor(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	_, err := db.Exec(ctx, `INSERT INTO app.meli_order_syncs
 (store_id,site_id,logistic_type,seller_id,window_from,window_to,next_offset)
 VALUES ($1::uuid,'MLB','remote','987654',now()-interval '30 days',now(),4)`, store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ResetOrderCursor(ctx, "550e8400-e29b-41d4-a716-446655440000", store, scope, 4); !errors.Is(err, ErrOrderResetConflict) {
		t.Fatalf("cross-owner reset: %v", err)
	}
	if _, err = s.ResetOrderCursor(ctx, owner, store, scope, 3); !errors.Is(err, ErrOrderResetConflict) {
		t.Fatalf("wrong offset reset: %v", err)
	}
	if _, err = s.ResetOrderCursor(ctx, owner, store, scope, 0); !errors.Is(err, ErrOrderResetInvalid) {
		t.Fatalf("no-op reset: %v", err)
	}
	_, err = db.Exec(ctx, `UPDATE app.meli_order_syncs SET lease_id=gen_random_uuid(),lease_until=now()+interval '1 minute'
 WHERE store_id=$1::uuid`, store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ResetOrderCursor(ctx, owner, store, scope, 4); !errors.Is(err, ErrOrderResetConflict) {
		t.Fatalf("reset bypassed active lease: %v", err)
	}
	_, err = db.Exec(ctx, `UPDATE app.meli_order_syncs SET lease_id=NULL,lease_until=NULL,pending_order_ids=ARRAY['101']
 WHERE store_id=$1::uuid`, store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ResetOrderCursor(ctx, owner, store, scope, 4); !errors.Is(err, ErrOrderResetConflict) {
		t.Fatalf("reset discarded pending orders: %v", err)
	}
	_, err = db.Exec(ctx, `UPDATE app.meli_order_syncs SET pending_order_ids='{}'::text[] WHERE store_id=$1::uuid`, store)
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.ResetOrderCursor(ctx, owner, store, scope, 4)
	if err != nil || !result.HasMore || result.NextOffset != 0 {
		t.Fatalf("reset=%+v error=%v", result, err)
	}
	var offset, audit int
	if err = db.QueryRow(ctx, `SELECT sync.next_offset,(SELECT count(*) FROM app.meli_order_resets
 WHERE store_id=$1::uuid AND previous_offset=4) FROM app.meli_order_syncs sync
 WHERE sync.store_id=$1::uuid`, store).Scan(&offset, &audit); err != nil || offset != 0 || audit != 1 {
		t.Fatalf("reset state offset=%d audit=%d error=%v", offset, audit, err)
	}
	if _, err = s.ResetOrderCursor(ctx, owner, store, scope, 4); !errors.Is(err, ErrOrderResetConflict) {
		t.Fatalf("repeated reset bypassed offset fence: %v", err)
	}
}

func orderTestDetail(id, seller, site string) string {
	return fmt.Sprintf(`{"id":%s,"seller":{"id":%s},"context":{"site":%q},"status":"payment_required","currency_id":"USD","total_amount":12.3400,"date_created":"2026-11-01T00:00:00Z","last_updated":"2026-11-01T01:00:00Z"}`, id, seller, site)
}

func TestOrderPostgresReviewAndSkip(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	const original = `[{ "user_id":987654,"site_id":"MLB","logistic_type":"remote"},{"user_id":987654,"site_id":"MPE","logistic_type":"remote"}]`
	if _, err := db.Exec(ctx, "UPDATE app.meli_store_connections SET marketplaces=$2::jsonb,access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid", store, original); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var firstFailure atomic.Int32
	firstFailure.Store(http.StatusTooManyRequests)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/marketplace/orders/search":
			if r.URL.Query().Get("site") == "MPE" {
				fmt.Fprint(w, `{"results":[],"paging":{"total":0}}`)
				return
			}
			switch r.URL.Query().Get("offset") {
			case "0":
				fmt.Fprint(w, `{"results":[{"orders":[{"id":101},{"id":102}]}],"paging":{"total":2}}`)
			case "1":
				fmt.Fprint(w, `{"results":[{"orders":[{"id":103},{"id":104}]}],"paging":{"total":2}}`)
			default:
				t.Error("unexpected offset")
				w.WriteHeader(400)
			}
		case "/marketplace/orders/102":
			switch status := firstFailure.Load(); status {
			case 0:
				fmt.Fprint(w, `{"invalid_json":`)
			case -1:
				<-r.Context().Done()
			default:
				if status == http.StatusTooManyRequests {
					w.Header().Set("Retry-After", "0")
				}
				w.WriteHeader(int(status))
			}
		case "/marketplace/orders/103":
			w.WriteHeader(http.StatusNotFound)
		case "/marketplace/orders/101", "/marketplace/orders/104":
			fmt.Fprint(w, orderTestDetail(strings.TrimPrefix(r.URL.Path, "/marketplace/orders/"), "987654", "MLB"))
		default:
			t.Error("unexpected upstream request")
			w.WriteHeader(400)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	const unknownReview = "00000000-0000-4000-8000-000000000001"
	for _, transient := range []struct {
		status int32
		code   string
	}{{429, "order_rate_limited"}, {503, "order_platform_unavailable"}, {-1, "order_timeout"}} {
		firstFailure.Store(transient.status)
		// The search quota may delay a subsequent attempt by 650ms; allow
		// enough time to reach the simulated 429/503/timeout response.
		attempt, stop := context.WithTimeout(ctx, 1500*time.Millisecond)
		_, err := s.SyncOrders(attempt, owner, store, scope)
		stop()
		if OrderErrorCode(err) != transient.code {
			t.Fatalf("expected %s, got %v", transient.code, err)
		}
		var flagged bool
		if err = db.QueryRow(ctx, "SELECT review_id IS NOT NULL FROM app.meli_order_syncs WHERE store_id=$1::uuid", store).Scan(&flagged); err != nil || flagged {
			t.Fatalf("transient failure was made skippable: flagged=%v err=%v", flagged, err)
		}
		if _, err = s.SkipOrder(ctx, owner, store, scope, "102", unknownReview, 0); !errors.Is(err, ErrOrderReviewConflict) {
			t.Fatalf("transient failure skip got %v", err)
		}
	}

	firstFailure.Store(0)
	result, err := s.SyncOrders(ctx, owner, store, scope)
	var review *OrderReview
	if !errors.As(err, &review) || result.Saved != 1 || result.NextOffset != 0 ||
		review.OrderID != "102" || review.Reason != "invalid_detail" || review.Offset != 0 || !validUUID(review.ReviewID) {
		t.Fatalf("first review=%+v result=%+v error=%v", review, result, err)
	}
	if _, err = s.GetOrder(ctx, owner, store, scope, "101"); err != nil {
		t.Fatalf("valid order prefix was not persisted: %v", err)
	}
	var pending string
	var persistedOffset int
	if err = db.QueryRow(ctx, `SELECT array_to_string(pending_order_ids,','),next_offset
 FROM app.meli_order_syncs WHERE store_id=$1::uuid AND site_id='MLB'`, store).Scan(&pending, &persistedOffset); err != nil || pending != "102" || persistedOffset != 0 {
		t.Fatalf("review did not preserve the pending row and offset: pending=%q offset=%d error=%v", pending, persistedOffset, err)
	}
	before := calls.Load()
	var repeated *OrderReview
	_, err = s.SyncOrders(ctx, owner, store, scope)
	if !errors.As(err, &repeated) || *repeated != *review || calls.Load() != before {
		t.Fatalf("unconfirmed review attempted to advance: %v", err)
	}
	if other, err := s.SyncOrders(ctx, owner, store, OrderScope{Site: "MPE", Logistic: "remote"}); err != nil || other.HasMore || other.NextOffset != 0 {
		t.Fatalf("review blocked another site: %+v %v", other, err)
	}

	for _, invalid := range []struct {
		owner, orderID string
		offset         int
	}{
		{owner, "104", 0}, {owner, "102", 1}, {unknownReview, "102", 0},
	} {
		if _, err = s.SkipOrder(ctx, invalid.owner, store, scope, invalid.orderID, review.ReviewID, invalid.offset); !errors.Is(err, ErrOrderReviewConflict) {
			t.Fatalf("invalid skip got %v", err)
		}
	}
	if _, err = db.Exec(ctx, `UPDATE app.meli_store_connections
 SET marketplaces=jsonb_set(marketplaces,'{0,user_id}','987655') WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SkipOrder(ctx, owner, store, scope, "102", review.ReviewID, 0); !errors.Is(err, ErrOrderReviewConflict) {
		t.Fatalf("cross-seller skip got %v", err)
	}
	if _, err = db.Exec(ctx, "UPDATE app.meli_store_connections SET marketplaces=$2::jsonb WHERE id=$1::uuid", store, original); err != nil {
		t.Fatal(err)
	}
	result, err = s.SkipOrder(ctx, owner, store, scope, "102", review.ReviewID, 0)
	if err != nil || !result.HasMore || result.NextOffset != 1 {
		t.Fatalf("skip should complete the first page: %+v %v", result, err)
	}

	_, err = s.SyncOrders(ctx, owner, store, scope)
	var last *OrderReview
	if !errors.As(err, &last) || last.OrderID != "103" || last.Reason != "detail_not_found" || last.Offset != 1 {
		t.Fatalf("second page review=%+v error=%v", last, err)
	}
	if _, err = s.SkipOrder(ctx, owner, store, scope, "103", review.ReviewID, 1); !errors.Is(err, ErrOrderReviewConflict) {
		t.Fatalf("old review ID applied to a new order: %v", err)
	}
	result, err = s.SkipOrder(ctx, owner, store, scope, "103", last.ReviewID, 1)
	if err != nil || !result.HasMore || result.NextOffset != 1 {
		t.Fatalf("skip advanced past unprocessed order: %+v %v", result, err)
	}
	if _, err = s.SkipOrder(ctx, owner, store, scope, "103", last.ReviewID, 1); !errors.Is(err, ErrOrderReviewConflict) {
		t.Fatalf("skip replay got %v", err)
	}
	result, err = s.SyncOrders(ctx, owner, store, scope)
	if err != nil || result.Saved != 1 || result.HasMore || result.NextOffset != 0 {
		t.Fatalf("last batch=%+v error=%v", result, err)
	}
	var audit string
	if err = db.QueryRow(ctx, `SELECT next_offset, (SELECT string_agg(order_id||':'||reason,',' ORDER BY order_id)
 FROM app.meli_order_skips WHERE store_id=$1::uuid AND site_id='MLB' AND seller_id='987654')
 FROM app.meli_order_syncs WHERE store_id=$1::uuid AND site_id='MLB'`, store).Scan(&persistedOffset, &audit); err != nil ||
		persistedOffset != result.NextOffset || audit != "102:invalid_detail,103:detail_not_found" {
		t.Fatalf("offset=%d audit=%q error=%v", persistedOffset, audit, err)
	}
	page, err := s.ListOrders(ctx, owner, store, scope, "", "", 0)
	if err != nil || len(page.Items) != 2 || page.Items[0].ID != "104" || page.Items[1].ID != "101" {
		t.Fatalf("skip lost valid details, page=%+v error=%v", page, err)
	}

	// Re-overlap the same two search envelopes: only still-invalid details
	// matching an earlier explicit Skip are bypassed, never valid details.
	replayed, err := s.SyncOrders(ctx, owner, store, scope)
	if err != nil || replayed.Saved != 1 || !replayed.HasMore || replayed.NextOffset != 1 {
		t.Fatalf("repeat skipped page=%+v error=%v", replayed, err)
	}
	replayed, err = s.SyncOrders(ctx, owner, store, scope)
	if err != nil || replayed.Saved != 1 || replayed.HasMore || replayed.NextOffset != 0 {
		t.Fatalf("repeat final page=%+v error=%v", replayed, err)
	}
	var skipCount int
	if err = db.QueryRow(ctx, "SELECT count(*) FROM app.meli_order_skips WHERE store_id=$1::uuid", store).Scan(&skipCount); err != nil || skipCount != 2 {
		t.Fatalf("duplicate Skip audit count=%d error=%v", skipCount, err)
	}
}

func TestOrderImageURLRestriction(t *testing.T) {
	for _, test := range []struct {
		url  string
		want bool
	}{
		{"https://http2.mlstatic.com/D_123-I.jpg", true},
		{"https://cbt-s2-p.mlstatic.com/abc.jpg", true},
		{"http://http2.mlstatic.com/D_123-I.jpg", false},
		{"https://fake-mlstatic.com/image.jpg", false},
		{"https://mlstatic.com.evil.example/image", false},
		{"javascript:alert(1)", false},
		{"https://some-user@http2.mlstatic.com/image.jpg", false},
	} {
		if got := orderImageURLAllowed(test.url); got != test.want {
			t.Errorf("image URL allowed=%t wanted=%t for %q", got, test.want, test.url)
		}
	}
}

func TestOrderPageReusesScopedPublishedThumbnail(t *testing.T) {
	svc, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	const orderItem = "MLB100012345"
	const orderLines = `[{"item_id":"MLB100012345","title":"Test item","quantity":1,"unit_price":"12.5","currency_id":"BRL"}]`
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
	 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
	  date_created,last_updated,order_items)
	 VALUES ($1::uuid,'MLB','remote','987654','10001','paid','BRL',12.5,now(),now(),$2::jsonb)`,
		store, orderLines); err != nil {
		t.Fatal(err)
	}
	const url = "https://http2.mlstatic.com/D_123-I.jpg"
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_published_listings
	 (store_id,site_id,logistic_type,listing_id,details)
	 VALUES ($1::uuid,'MLB','remote',$2,'{"seller_id":987654,"secure_thumbnail":"https://http2.mlstatic.com/D_123-I.jpg"}'::jsonb)`,
		store, orderItem); err != nil {
		t.Fatal(err)
	}
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	page, err := svc.ListOrders(ctx, owner, store, scope, "", 0)
	if err != nil || len(page.Items) != 1 || len(page.Items[0].Items) != 1 ||
		page.Items[0].Items[0].ImageURL != url {
		t.Fatalf("scoped thumbnail is missing: page=%+v err=%v", page, err)
	}
	// A different child seller cannot attach its thumbnail to our order.
	if _, err := db.Exec(ctx, `UPDATE app.meli_published_listings SET details=
	 '{"seller_id":987655,"secure_thumbnail":"https://http2.mlstatic.com/D_123-I.jpg"}'::jsonb
	 WHERE store_id=$1::uuid AND listing_id=$2`, store, orderItem); err != nil {
		t.Fatal(err)
	}
	order, err := svc.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || len(order.Items) != 1 || order.Items[0].ImageURL != "" {
		t.Fatalf("cross-seller image leaked: order=%+v err=%v", order, err)
	}
}
