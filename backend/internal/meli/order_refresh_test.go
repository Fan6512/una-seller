package meli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRefreshOneOrderKeepsPurchaseAndReplacesStatus(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections
	SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
	(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,
	date_created,last_updated,shipment_id,shipment_snapshot,order_items)
	VALUES ($1::uuid,'MLB','remote','987654','10001','paid','USD',29.99,
	'2026-11-20T00:00:00Z','2026-11-20T00:01:00Z','801',
	'{"shipment_id":"801","status":"delivered","checked_at":"2026-11-20T00:02:00Z"}'::jsonb,
	'[{"item_id":"MLB100","title":"Item","quantity":1,"unit_price":"29.99","currency_id":"USD","image_url":"https://http2.mlstatic.com/D_100-I.jpg"}]'::jsonb)`,
		store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_order_purchases
	(store_id,site_id,logistic_type,seller_id,order_id,purchase_order_id,cost_cny)
	VALUES ($1::uuid,'MLB','remote','987654','10001','12345678',15.50)`, store); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/marketplace/orders/10001" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
 "id":10001,"seller":{"id":987654},"site_id":"MLB",
 "status":"partially_refunded","currency_id":"USD","total_amount":29.99,"shipping":{"id":802},
 "date_created":"2026-11-20T00:00:00Z","last_updated":"2026-11-21T01:00:00Z",
 "order_items":[{"item":{"id":"MLB100","title":"Item"},"quantity":1,"unit_price":29.99,"currency_id":"USD"}]
 }`))
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	refreshed, err := s.RefreshOneOrder(ctx, owner, store, scope, "10001")
	if err != nil || refreshed.Status != "partially_refunded" || refreshed.PurchaseOrderID != "12345678" ||
		len(refreshed.Items) != 1 || refreshed.Items[0].ImageURL != "https://http2.mlstatic.com/D_100-I.jpg" ||
		refreshed.SalesAmountUSD != "29.99" {
		t.Fatalf("order refresh=%+v error=%v", refreshed, err)
	}
	var state string
	var snapshotCleared bool
	if err = db.QueryRow(ctx, "SELECT status,shipment_snapshot IS NULL FROM app.meli_orders WHERE store_id=$1::uuid AND order_id='10001'", store).
		Scan(&state, &snapshotCleared); err != nil || state != "partially_refunded" || !snapshotCleared ||
		refreshed.ShipmentID != "802" || refreshed.Shipment != nil {
		t.Fatalf("stored status=%s snapshot cleared=%t shipment=%+v error=%v", state, snapshotCleared, refreshed.Shipment, err)
	}
	if _, err = s.RefreshOneOrder(ctx, owner, store, scope, "99999"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("refresh unknown order must not create a record: %v", err)
	}
}
