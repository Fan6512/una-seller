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
	if err != nil || cached.Shipment == nil || cached.Shipment.Details != nil || cached.Shipment.DeliveredAt != nil || calls.Load() != before {
		t.Fatalf("detail history leaked into the cached read: %+v, %v", cached.Shipment, err)
	}
	cached, err = s.GetOrder(ctx, otherOwner, otherStore, scope, "10001")
	if err != nil || cached.ShipmentID != "987651236" || cached.Shipment != nil {
		t.Fatalf("detail was attached to another store: %+v, %v", cached.Shipment, err)
	}

	reply.Store(&response{code: http.StatusOK, body: delivered, carrierCode: http.StatusOK,
		carrierBody: `{"url":"javascript:alert(1)"}`})
	result, err = s.RefreshOrderShipment(ctx, owner, store, scope, "10001", true)
	if err != nil || result.Shipment == nil || result.Shipment.Details == nil ||
		result.Shipment.Details.TrackingURL != "" || result.Shipment.Details.History != nil || result.Shipment.DeliveredAt != nil {
		t.Fatalf("missing history or unsafe tracking URL accepted: %+v, %v", result.Shipment, err)
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
	if !errors.Is(err, ErrOrderAuthorizationChanged) {
		t.Fatalf("replaced shipment accepted late update: %v", err)
	}
	got, err := s.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || got.ShipmentID != "987651235" || got.Shipment != nil {
		t.Fatalf("replacement inherited a prior shipment snapshot: %+v, %v", got, err)
	}
}
