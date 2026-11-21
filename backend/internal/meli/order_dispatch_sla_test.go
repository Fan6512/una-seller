package meli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRefreshOrderDispatchSLAUsesOfficialExpectedDateAndScope(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections
 SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
 date_created,last_updated,shipment_id,order_items)
 VALUES($1::uuid,'MLB','remote','987654','10001','paid','USD',29.99,
 '2026-11-20T00:00:00Z','2026-11-20T00:01:00Z','987651234','[]'::jsonb)`, store); err != nil {
		t.Fatal(err)
	}
	reply := `{"status":"on_time","expected_date":"2026-11-23T08:30:00-03:00"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-format-new") != "true" ||
			r.Header.Get("Authorization") == "" || r.URL.Path != "/marketplace/shipments/987651234/sla" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(reply))
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	got, err := s.RefreshOrderDispatchSLA(ctx, owner, store, scope, "10001")
	if err != nil || got.DispatchDeadline == nil || got.DispatchSLAStatus != "on_time" ||
		got.DispatchDeadline.UTC().Format(time.RFC3339) != "2026-11-23T11:30:00Z" ||
		got.DispatchCheckedAt == nil {
		t.Fatalf("official SLA missing or wrong: %+v, %v", got, err)
	}
	// A new shipment never inherits the previous shipment's SLA.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='987651235'
 WHERE store_id=$1::uuid AND site_id='MLB' AND logistic_type='remote' AND seller_id='987654'
 AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	noOldSLA, err := s.GetOrder(ctx, owner, store, scope, "10001")
	if err != nil || noOldSLA.DispatchDeadline != nil || noOldSLA.DispatchCheckedAt != nil {
		t.Fatalf("old shipment deadline leaked to new shipment: %+v, %v", noOldSLA, err)
	}
	// Never write a deadline for an order that belongs to another owner.
	if _, err := s.RefreshOrderDispatchSLA(ctx,
		"00000000-0000-4000-8000-000000000099", store, scope, "10001"); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("cross-owner SLA lookup: %v", err)
	}
}

func TestDispatchSLAInvalidResponseDoesNotWrite(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections
 SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
 date_created,last_updated,shipment_id,order_items)
 VALUES($1::uuid,'MLB','remote','987654','10001','paid','USD',29.99,
 '2026-11-20T00:00:00Z','2026-11-20T00:01:00Z','987651234','[]'::jsonb)`, store); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"on_time","expected_date":"not-a-date"}`))
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	if _, err := s.RefreshOrderDispatchSLA(ctx, owner, store, scope, "10001"); !errors.Is(err, ErrOrderSLAInvalid) {
		t.Fatalf("invalid official cutoff accepted: %v", err)
	}
	var total int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM app.meli_order_dispatch_sla
 WHERE store_id=$1::uuid`, store).Scan(&total); err != nil || total != 0 {
		t.Fatalf("invalid SLA wrote persisted data: total=%d error=%v", total, err)
	}
}

func TestDispatchSLARejectsCancelledWithoutRemoteRequest(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
	(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
	date_created,last_updated,shipment_id,order_items)
	VALUES ($1::uuid,'MLB','remote','987654','10001','cancelled','USD',12.00,now(),now(),'123456789','[]'::jsonb)`, store); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("cancelled order must not reach official SLA endpoint")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	s.apiURL = server.URL
	_, err := s.RefreshOrderDispatchSLA(ctx, owner, store, OrderScope{Site: "MLB", Logistic: "remote"}, "10001")
	if !errors.Is(err, ErrOrderSLANotApplicable) {
		t.Fatalf("cancelled shipment SLA: %v", err)
	}
}
