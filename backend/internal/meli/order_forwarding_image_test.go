package meli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestManualForwardingWarehouseAndPurchaseTracking(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	_, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,order_items)
 VALUES($1::uuid,'MLB','remote','987654','9001','paid','USD',16,now(),now(),'[]'::jsonb)`, store)
	if err != nil {
		t.Fatal(err)
	}
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	purchase, err := s.SaveOrderPurchaseWithShipping(ctx, owner, store, scope,
		"9001", "5127949513086296911", "16", "东莞市常平镇货代仓 A 库", "SF123456789", nil)
	if err != nil || purchase.PurchaseTrackingNumber != "SF123456789" {
		t.Fatalf("save forwarding: %+v, %v", purchase, err)
	}
	got, err := s.GetOrderPurchase(ctx, owner, store, scope, "9001")
	if err != nil || got == nil || got.WarehouseAddress != "东莞市常平镇货代仓 A 库" ||
		got.PurchaseTrackingNumber != "SF123456789" {
		t.Fatalf("persist forwarding: %+v %v", got, err)
	}
	addresses, err := s.ListPurchaseWarehouses(ctx, owner, store, scope)
	if err != nil || len(addresses) != 1 || addresses[0] != got.WarehouseAddress {
		t.Fatalf("warehouse choices: %v, %v", addresses, err)
	}
	order, err := s.GetOrder(ctx, owner, store, scope, "9001")
	if err != nil || order.PurchaseTrackingNumber != "SF123456789" {
		t.Fatalf("tracking not projected: %+v %v", order, err)
	}
	if _, err = s.SaveOrderPurchaseWithShipping(ctx, owner, store, scope,
		"9001", purchase.PurchaseOrderID, "16", "带换行\n的地址", "BAD TRACK", &purchase.UpdatedAt);
		!errors.Is(err, ErrOrderPurchaseInvalid) {
		t.Fatalf("accepted invalid forwarding data: %v", err)
	}
}

func TestManualOfficialOrderImageBackfill(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections
 SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	_, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,order_items)
 VALUES($1::uuid,'MLB','remote','987654','9002','paid','USD',20,now(),now(),
 '[{"item_id":"MLB123456","title":"Plant cleaner","quantity":1,"unit_price":"20","currency_id":"USD"}]'::jsonb)`, store)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/marketplace/items/MLB123456" || r.Header.Get("Authorization") == "" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"MLB123456","seller_id":987654,"site_id":"MLB",
	"secure_thumbnail":"https://http2.mlstatic.com/D_123-I.jpg"}`))
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	order, err := s.RefreshOrderPrimaryImage(ctx, owner, store, scope, "9002")
	if err != nil || len(order.Items) != 1 ||
		order.Items[0].ImageURL != "https://http2.mlstatic.com/D_123-I.jpg" {
		t.Fatalf("image backfill: %+v %v", order, err)
	}
	again, err := s.RefreshOrderPrimaryImage(ctx, owner, store, scope, "9002")
	if err != nil || len(again.Items) != 1 || calls.Load() != 1 {
		t.Fatalf("cached image was fetched twice: %+v calls=%d err=%v", again, calls.Load(), err)
	}
}
