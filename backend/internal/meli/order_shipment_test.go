package meli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRefreshOrderShipmentOfficialSnapshotAndScope(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
 date_created,last_updated,shipment_id,order_items)
 VALUES($1::uuid,'MLB','remote','987654','10001','paid','USD',29.99,now(),now(),'987651234','[]'::jsonb),
 ($1::uuid,'MLB','remote','987654','10002','paid','USD',29.99,now(),now(),NULL,'[]'::jsonb)`, store); err != nil {
		t.Fatal(err)
	}

	type response struct {
		code          int
		body          string
		carrierCode   int
		carrierBody   string
		historyBody   string
		beforeHistory func()
		retryAfter    string
		waitForCancel bool
	}
	var reply atomic.Pointer[response]
	var calls atomic.Int32
	var detailCalls atomic.Int32
	const shipmentPath = "/marketplace/shipments/987651234"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet ||
			r.Header.Get("x-format-new") != "true" || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Errorf("unexpected official shipment request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		current := reply.Load()
		switch r.URL.Path {
		case shipmentPath + "/carrier":
			detailCalls.Add(1)
			w.Header().Set("Retry-After", current.retryAfter)
			w.WriteHeader(current.carrierCode)
			_, _ = w.Write([]byte(current.carrierBody))
			return
		case shipmentPath + "/history":
			detailCalls.Add(1)
			if current.beforeHistory != nil {
				current.beforeHistory()
			}
			if current.historyBody == "" {
				w.WriteHeader(http.StatusNotFound)
			} else {
				_, _ = w.Write([]byte(current.historyBody))
			}
			return
		case shipmentPath:
		default:
			t.Errorf("unexpected shipment path: %s", r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if current.waitForCancel {
			<-r.Context().Done()
			return
		}
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(current.code)
		_, _ = w.Write([]byte(current.body))
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	for _, request := range []struct {
		owner, orderID string
		scope          OrderScope
		want           error
	}{
		{owner, "10002", scope, ErrOrderShipmentMissing},
		{"00000000-0000-4000-8000-000000000099", "10001", scope, ErrStoreNotFound},
		{owner, "10001", OrderScope{Site: "MLM", Logistic: "remote"}, ErrSiteNotAuthorized},
	} {
		_, err := s.RefreshOrderShipment(ctx, request.owner, store, request.scope, request.orderID, true)
		if !errors.Is(err, request.want) {
			t.Fatalf("invalid shipment scope: got %v, want %v", err, request.want)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("an unauthorized or missing shipment triggered an official request")
	}

	for i, state := range []struct{ status, substatus string }{
		{"ready_to_ship", "printed"}, {"shipped", ""}, {"shipped", "in_transit"},
		{"shipped", "out_for_delivery"}, {"delivered", ""}, {"cancelled", ""}, {"new_official_status", ""},
	} {
		body := fmt.Sprintf(`{"id":987651234,"status":%q,"substatus":%q,"tracking_number":"TRACK-001",
 "logistic":{"type":"drop_off"},"last_updated":"2026-11-20T08:%02d:00-03:00",
 "destination":{"receiver_name":"NEVER PERSIST THIS"}}`, state.status, state.substatus, i)
		reply.Store(&response{code: http.StatusOK, body: body})
		got, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", false)
		if err != nil || got.Status != "paid" || got.Shipment == nil || got.Shipment.ID != "987651234" ||
			got.Shipment.Status != state.status || got.Shipment.Substatus != state.substatus ||
			got.Shipment.TrackingNumber != "TRACK-001" || got.Shipment.LogisticType != "drop_off" ||
			got.Shipment.CheckedAt.IsZero() || got.Shipment.LastUpdated == nil {
			t.Fatalf("wrong official shipment snapshot: %+v, %v", got.Shipment, err)
		}
	}
	if detailCalls.Load() != 0 {
		t.Fatal("summary refresh fetched carrier or history")
	}
	var persisted string
	if err := db.QueryRow(ctx, `SELECT shipment_snapshot::text FROM app.meli_orders WHERE store_id=$1::uuid AND order_id='10001'`, store).Scan(&persisted); err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct {
		res  response
		want error
	}{
		{response{code: http.StatusNotFound}, ErrOrderShipmentMissing},
		{response{code: http.StatusUnauthorized}, ErrOrderShipmentDenied},
		{response{code: http.StatusForbidden}, ErrOrderShipmentDenied},
		{response{code: http.StatusTooManyRequests}, ListingPlatformError(http.StatusTooManyRequests)},
		{response{code: http.StatusOK, body: `{"id":100,"status":"delivered"}`}, ErrOrderInvalidResponse},
		{response{code: http.StatusOK, waitForCancel: true}, context.DeadlineExceeded},
	} {
		reply.Store(&item.res)
		requestCtx := ctx
		if item.res.waitForCancel {
			var stop context.CancelFunc
			requestCtx, stop = context.WithTimeout(ctx, 250*time.Millisecond)
			defer stop()
		}
		_, err := s.RefreshOrderShipment(requestCtx, owner, store, scope, "10001", false)
		if !errors.Is(err, item.want) {
			t.Fatalf("upstream failure: got %v, want %v for %d", err, item.want, item.res.code)
		}
	}
	var after string
	if err := db.QueryRow(ctx, `SELECT shipment_snapshot::text FROM app.meli_orders WHERE store_id=$1::uuid AND order_id='10001'`, store).Scan(&after); err != nil || persisted != after {
		t.Fatal("failed refresh replaced the most recent confirmed shipment")
	}

	_, _, otherOwner, otherStore := setupRefreshTest(t)
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
 date_created,last_updated,shipment_id,order_items)
 VALUES($1::uuid,'MLB','remote','987654','10001','paid','USD',19,now(),now(),'987651236','[]'::jsonb)`, otherStore); err != nil {
		t.Fatal(err)
	}
	before := calls.Load()
	if _, err := s.RefreshOrderShipment(ctx, owner, otherStore, scope, "10001", true); !errors.Is(err, ErrStoreNotFound) || calls.Load() != before {
		t.Fatalf("cross-store detail called the platform: %v", err)
	}
	const delivered = `{"id":987651234,"status":"delivered","fm_carrier":"First Mile",
 "last_updated":"2026-11-20T08:07:00-03:00","status_history":{"date_shipped":"2026-11-18T04:00:00-03:00"}}`
	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusOK,
		carrierBody: `{"name":"Official Carrier","url":"https://tracking.example/trip/123","receiver":"private"}`,
		historyBody: `[{"status":"delivered","date":"2026-11-19T09:30:00-03:00"},
 {"status":"shipped","date":"2026-11-18T10:00:00Z"}, {"status":"handling","date":"2026-11-17T10:00:00Z"},
 {"status":"pending","date":null}, {"status":"delivered","date":"invalid"}]`})
	result, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
	if err != nil || result.Shipment == nil || result.Shipment.Details == nil {
		t.Fatalf("shipment details unavailable: %+v, %v", result.Shipment, err)
	}
	shipment := result.Shipment
	if shipment.FirstMileCarrier != "First Mile" || shipment.Details.CarrierName != "Official Carrier" ||
		shipment.Details.TrackingURL != "https://tracking.example/trip/123" || len(shipment.Details.History) != 3 ||
		shipment.Details.History[0].Status != "handling" || shipment.Details.History[2].Status != "delivered" ||
		shipment.ShippedAt == nil || shipment.ShippedAt.UTC().Format(time.RFC3339) != "2026-11-18T07:00:00Z" ||
		shipment.DeliveredAt == nil || shipment.DeliveredAt.UTC().Format(time.RFC3339) != "2026-11-19T12:30:00Z" {
		t.Fatalf("wrong shipment details: %+v, %+v", shipment, shipment.Details)
	}
	before = calls.Load()
	cached, err := s.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || cached.Shipment == nil || cached.Shipment.Details == nil || cached.Shipment.DeliveredAt == nil ||
		!cached.Shipment.DeliveredAt.Equal(*shipment.DeliveredAt) ||
		cached.Shipment.Details.CarrierName != shipment.Details.CarrierName ||
		cached.Shipment.Details.TrackingURL != shipment.Details.TrackingURL ||
		len(cached.Shipment.Details.History) != 3 ||
		!cached.Shipment.Details.CheckedAt.Equal(shipment.CheckedAt) || calls.Load() != before {
		t.Fatalf("cached detail read lost official history, carrier or dates: %+v, %v", cached.Shipment, err)
	}
	// Non-detail consumers use the same scoped summary projection.
	summary, err := s.GetOrderSummary(ctx, owner, store, scope, "10001")
	if err != nil || summary.Shipment == nil || summary.Shipment.Details != nil || summary.Shipment.DeliveredAt == nil {
		t.Fatalf("summary read returned wrong projection: %+v, %v", summary.Shipment, err)
	}
	if _, err := s.GetOrderSummary(ctx, owner, otherStore, scope, "10001"); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("summary read bypassed owner scope: %v", err)
	}
	// List and global search must not expose cached details.
	page, err := s.ListOrders(ctx, owner, store, scope, "", "", 0)
	if err != nil || len(page.Items) != 2 {
		t.Fatalf("cached list: %+v, %v", page, err)
	}
	for _, row := range page.Items {
		if row.ID == "10001" && (row.Shipment == nil || row.Shipment.Details != nil) {
			t.Fatalf("list exposed details or lost summary: %+v", row.Shipment)
		}
	}
	search, err := s.SearchOrders(ctx, owner, OrderSearchFilter{Store: store, Site: "MLB", Logistic: "remote"})
	if err != nil || len(search.Items) != 2 {
		t.Fatalf("cached search: %+v, %v", search, err)
	}
	for _, row := range search.Items {
		if row.ID == "10001" && (row.Shipment == nil || row.Shipment.Details != nil) {
			t.Fatalf("search exposed details or lost summary: %+v", row.Shipment)
		}
	}
	for _, filter := range []string{"fulfillment_delivered", "fulfillment_paid", "paid"} {
		page, err := s.ListOrders(ctx, owner, store, scope, filter, "", 0)
		search, searchErr := s.SearchOrders(ctx, owner, OrderSearchFilter{Store: store, Site: "MLB", Logistic: "remote", Status: filter})
		if err != nil || searchErr != nil || len(page.Items) != len(search.Items) {
			t.Fatalf("inconsistent fulfillment filters %s: %v / %v", filter, err, searchErr)
		}
		found := false
		for _, row := range page.Items {
			if row.ID == "10001" {
				found = true
			}
		}
		if found != (filter != "fulfillment_paid") {
			t.Fatalf("incorrect delivered order inclusion for %s: %+v", filter, page.Items)
		}
	}
	if calls.Load() != before {
		t.Fatal("cached reads requested official API")
	}
	// Summary refreshes must preserve verified dates and details when omitted.
	reply.Store(&response{code: http.StatusOK, body: delivered})
	result, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", false)
	if err != nil || result.Shipment == nil || result.Shipment.Details != nil ||
		result.Shipment.DeliveredAt == nil || !result.Shipment.DeliveredAt.Equal(*shipment.DeliveredAt) {
		t.Fatalf("summary refresh lost the verified delivery date: %+v, %v", result.Shipment, err)
	}
	cached, err = s.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || cached.Shipment == nil || cached.Shipment.Details == nil ||
		len(cached.Shipment.Details.History) != 3 ||
		!cached.Shipment.Details.CheckedAt.Equal(shipment.CheckedAt) {
		t.Fatalf("summary refresh erased cached details or changed their verification time: %+v, %v", cached.Shipment, err)
	}
	cached, err = s.GetOrder(ctx, otherOwner, otherStore, scope, "10001")
	if err != nil || cached.ShipmentID != "987651236" || cached.Shipment != nil {
		t.Fatalf("detail was attached to another store: %+v, %v", cached.Shipment, err)
	}

	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusOK,
		carrierBody: `{"url":"javascript:alert(1)"}`})
	result, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
	if err != nil || result.Shipment == nil || result.Shipment.Details == nil ||
		result.Shipment.Details.TrackingURL != "" || result.Shipment.Details.History != nil ||
		result.Shipment.DeliveredAt == nil || !result.Shipment.DeliveredAt.Equal(*shipment.DeliveredAt) {
		t.Fatalf("missing history or unsafe tracking URL accepted: %+v, %v", result.Shipment, err)
	}
	// Explicit negative current status must not inherit a prior delivered date.
	for _, status := range []string{"not_delivered", "cancelled"} {
		reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusOK,
			carrierBody: `{"name":"Official Carrier"}`,
			historyBody: `[{"status":"delivered","date":"2026-11-19T09:30:00-03:00"}]`})
		if restored, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true); err != nil ||
			restored.Shipment == nil || restored.Shipment.DeliveredAt == nil {
			t.Fatalf("restore confirmed delivery before reversal: %+v, %v", restored.Shipment, err)
		}
		body := fmt.Sprintf(`{"id":987651234,"status":%q,"last_updated":"2026-11-20T08:07:00-03:00"}`, status)
		reply.Store(&response{code: http.StatusOK, body: body})
		reversed, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", false)
		if err != nil || reversed.Shipment == nil || reversed.Shipment.Status != status || reversed.Shipment.DeliveredAt != nil {
			t.Fatalf("negative status inherited obsolete delivery date: %+v, %v", reversed.Shipment, err)
		}
		reply.Store(&response{code: http.StatusOK, body: body, carrierCode: http.StatusOK,
			carrierBody: `{"name":"Official Carrier"}`,
			historyBody: `[{"status":"delivered","date":"2026-11-19T09:30:00-03:00"}]`})
		reversed, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
		if err != nil || reversed.Shipment == nil || reversed.Shipment.DeliveredAt != nil ||
			reversed.Shipment.Details == nil || reversed.Shipment.Details.CarrierName != "Official Carrier" ||
			len(reversed.Shipment.Details.History) != 1 {
			t.Fatalf("negative status reused historical date or erased history: %+v, %v", reversed.Shipment, err)
		}
		// An explicit official date is still a historical fact, even in a negative state.
		explicit := fmt.Sprintf(`{"id":987651234,"status":%q,"last_updated":"2026-11-20T08:07:00-03:00",
 "status_history":{"date_delivered":"2026-11-19T09:30:00-03:00"}}`, status)
		reply.Store(&response{code: http.StatusOK, body: explicit, carrierCode: http.StatusOK,
			carrierBody: `{"name":"Official Carrier"}`,
			historyBody: `[{"status":"delivered","date":"2026-11-19T09:30:00-03:00"}]`})
		confirmed, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
		if err != nil || confirmed.Shipment == nil || confirmed.Shipment.Status != status ||
			confirmed.Shipment.DeliveredAt == nil || !confirmed.Shipment.DeliveredAt.Equal(*shipment.DeliveredAt) {
			t.Fatalf("negative status lost explicit official delivery date: %+v, %v", confirmed.Shipment, err)
		}
		confirmed, err = s.GetOrderSummary(ctx, owner, store, scope, "10001")
		if err != nil || confirmed.Shipment == nil || confirmed.Shipment.Status != status ||
			confirmed.Shipment.DeliveredAt == nil || !confirmed.Shipment.DeliveredAt.Equal(*shipment.DeliveredAt) ||
			confirmed.Shipment.Details != nil {
			t.Fatalf("cached summary lost explicit historical delivery date: %+v, %v", confirmed.Shipment, err)
		}
	}
	// With no prior confirmed delivery date, a substatus alone cannot create one.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_snapshot=shipment_snapshot-'delivered_at'
 WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusNotFound,
		historyBody: `[{"status":"delivered","substatus":"updated","date":"2026-11-19T09:30:00-03:00"}]`})
	result, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
	if err != nil || result.Shipment == nil || result.Shipment.Details == nil ||
		result.Shipment.Details.CarrierName != "" || len(result.Shipment.Details.History) != 1 || result.Shipment.DeliveredAt != nil {
		t.Fatalf("missing carrier or delivery substatus handled incorrectly: %+v, %v", result.Shipment, err)
	}

	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusNotFound, historyBody: "[]",
		beforeHistory: func() {
			if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_snapshot=jsonb_set(shipment_snapshot,
 '{checked_at}',to_jsonb(clock_timestamp()+interval '1 second')) WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
				t.Error(err)
			}
		}})
	if _, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true); !errors.Is(err, ErrOrderShipmentChanged) {
		t.Fatalf("late shipment details accepted for superseded snapshot: %v", err)
	}

	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusTooManyRequests, retryAfter: "15"})
	before = calls.Load()
	_, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
	var limit *OrderRateLimitError
	if !errors.As(err, &limit) || limit.RetryAfter() == "0" || calls.Load() != before+2 {
		t.Fatalf("carrier 429 did not stop before requesting history: %v", err)
	}
	if _, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", false); !errors.As(err, &limit) || calls.Load() != before+2 {
		t.Fatalf("summary bypassed detail rate-limit cooldown: %v", err)
	}
	s.orderRetryUntil = time.Time{}

	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusNotFound, historyBody: "[]",
		beforeHistory: func() {
			if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='987651235' WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
				t.Error(err)
			}
		}})
	_, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
	if !errors.Is(err, ErrOrderShipmentChanged) {
		t.Fatalf("replaced shipment must return shipment conflict: %v", err)
	}
	// A changed authorization must not be misclassified as a shipment conflict.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='987651234' WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusNotFound, historyBody: "[]",
		beforeHistory: func() {
			if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces_status='unverified'
 WHERE id=$1::uuid`, store); err != nil {
				t.Error(err)
			}
		}})
	if _, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true); !errors.Is(err, ErrOrderAuthorizationChanged) {
		t.Fatalf("revoked mapping must return authorization conflict: %v", err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces_status='verified' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='987651235' WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || got.ShipmentID != "987651235" || got.Shipment != nil {
		t.Fatalf("replacement inherited a prior shipment snapshot: %+v, %v", got, err)
	}
}
