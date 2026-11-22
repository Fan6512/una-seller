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
		code           int
		body           string
		changeShipment bool
		waitForCancel  bool
	}
	var reply atomic.Pointer[response]
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/marketplace/shipments/987651234" ||
			r.Header.Get("x-format-new") != "true" || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Errorf("unexpected official shipment request")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		current := reply.Load()
		if current.waitForCancel {
			<-r.Context().Done()
			return
		}
		if current.changeShipment {
			if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='987651235' WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
				t.Error(err)
			}
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
		_, err := s.RefreshOrderShipment(ctx, request.owner, store, request.scope, request.orderID)
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
		got, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001")
		if err != nil || got.Status != "paid" || got.Shipment == nil || got.Shipment.ID != "987651234" ||
			got.Shipment.Status != state.status || got.Shipment.Substatus != state.substatus ||
			got.Shipment.TrackingNumber != "TRACK-001" || got.Shipment.LogisticType != "drop_off" ||
			got.Shipment.CheckedAt.IsZero() || got.Shipment.LastUpdated == nil {
			t.Fatalf("wrong official shipment snapshot: %+v, %v", got.Shipment, err)
		}
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
		_, err := s.RefreshOrderShipment(requestCtx, owner, store, scope, "10001")
		if !errors.Is(err, item.want) {
			t.Fatalf("upstream failure: got %v, want %v for %d", err, item.want, item.res.code)
		}
	}
	var after string
	if err := db.QueryRow(ctx, `SELECT shipment_snapshot::text FROM app.meli_orders WHERE store_id=$1::uuid AND order_id='10001'`, store).Scan(&after); err != nil || persisted != after {
		t.Fatal("failed refresh replaced the most recent confirmed shipment")
	}
	reply.Store(&response{code: http.StatusOK, body: `{"id":987651234,"status":"delivered"}`, changeShipment: true})
	_, err := s.RefreshOrderShipment(ctx, owner, store, scope, "10001")
	if !errors.Is(err, ErrOrderAuthorizationChanged) {
		t.Fatalf("replaced shipment accepted late update: %v", err)
	}
	got, err := s.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || got.ShipmentID != "987651235" || got.Shipment != nil {
		t.Fatalf("replacement inherited a prior shipment snapshot: %+v, %v", got, err)
	}
}
