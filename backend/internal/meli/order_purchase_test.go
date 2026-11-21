package meli

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestOrderPurchaseManualLinkAndSellerIsolation(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	_, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated)
 VALUES($1::uuid,'MLB','remote','987654','401','paid','USD',12.34,now(),now())`, store)
	if err != nil {
		t.Fatal(err)
	}
	missing, err := s.GetOrderPurchase(ctx, owner, store, scope, "401")
	if err != nil || missing != nil {
		t.Fatalf("new order purchase=%+v error=%v", missing, err)
	}
	// A distinct unlinked order proves the filter is applied before pagination.
	if _, err = db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated)
 VALUES($1::uuid,'MLB','remote','987654','402','paid','USD',9.99,now(),now())`, store); err != nil {
		t.Fatal(err)
	}
	const purchaseID = "1234567890123456"
	if _, err = s.SaveOrderPurchase(ctx, owner, store, scope, "401", purchaseID, "16.00", nil); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetOrderPurchase(ctx, owner, store, scope, "401")
	if err != nil || got == nil || got.PurchaseOrderID != purchaseID || got.CostCNY == nil || *got.CostCNY != "16" || got.Source != "manual" {
		t.Fatalf("purchase=%+v error=%v", got, err)
	}
	linked, err := s.ListOrders(ctx, owner, store, scope, "", "linked", 0)
	if err != nil || len(linked.Items) != 1 || linked.Items[0].PurchaseOrderID != purchaseID ||
		linked.Items[0].PurchaseCostCNY == nil || *linked.Items[0].PurchaseCostCNY != "16" {
		t.Fatalf("linked purchase filter returned %+v, %v", linked, err)
	}
	unlinked, err := s.ListOrders(ctx, owner, store, scope, "", "unlinked", 0)
	if err != nil || len(unlinked.Items) != 1 || unlinked.Items[0].ID != "402" ||
		unlinked.Items[0].PurchaseOrderID != "" || unlinked.Items[0].PurchaseCostCNY != nil {
		t.Fatalf("unlinked purchase filter returned %+v, %v", unlinked, err)
	}
	detail, err := s.GetOrder(ctx, owner, store, scope, "401")
	if err != nil || detail.PurchaseOrderID != purchaseID {
		t.Fatalf("order purchase detail missing: %+v, %v", detail, err)
	}
	for _, version := range []*time.Time{nil, &got.UpdatedAt} {
		replayed, err := s.SaveOrderPurchase(ctx, owner, store, scope, "401", purchaseID, "16.0", version)
		if err != nil || !replayed.UpdatedAt.Equal(got.UpdatedAt) {
			t.Fatalf("replay changed purchase version: %+v, %v", replayed, err)
		}
	}
	type saveResult struct {
		purchase *OrderPurchase
		err      error
	}
	results := make(chan saveResult, 2)
	for _, cost := range []string{"17", "18"} {
		go func() {
			purchase, err := s.SaveOrderPurchase(ctx, owner, store, scope, "401", purchaseID, cost, &got.UpdatedAt)
			results <- saveResult{purchase, err}
		}()
	}
	var winner *OrderPurchase
	conflicts := 0
	for range 2 {
		result := <-results
		switch {
		case result.err == nil:
			winner = result.purchase
		case errors.Is(result.err, ErrOrderPurchaseConflict):
			conflicts++
		default:
			t.Fatal(result.err)
		}
	}
	latest, err := s.GetOrderPurchase(ctx, owner, store, scope, "401")
	if err != nil || latest == nil || conflicts != 1 || winner == nil || !winner.UpdatedAt.After(got.UpdatedAt) ||
		!latest.UpdatedAt.Equal(winner.UpdatedAt) || *latest.CostCNY != *winner.CostCNY {
		t.Fatalf("purchase after concurrent writes: latest=%+v winner=%+v conflicts=%d error=%v", latest, winner, conflicts, err)
	}
	if _, err = s.SaveOrderPurchase(ctx, owner, store, scope, "401", purchaseID, "16.123", nil); !errors.Is(err, ErrOrderPurchaseInvalid) {
		t.Fatalf("invalid cost accepted: %v", err)
	}
	if _, err = s.SaveOrderPurchase(ctx, owner, store, scope, "401", "bad-number", "", nil); !errors.Is(err, ErrOrderPurchaseInvalid) {
		t.Fatalf("invalid purchase ID accepted: %v", err)
	}
	if _, err = s.SaveOrderPurchase(ctx, "550e8400-e29b-41d4-a716-446655440000", store, scope, "401", purchaseID, "99", nil); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("cross owner write: %v", err)
	}
	var otherStore string
	err = db.QueryRow(ctx, `INSERT INTO app.meli_store_connections(owner_user_id,seller_id,status,marketplaces,marketplaces_status)
 SELECT owner_user_id,seller_id||'0','reauth_required',marketplaces,'verified'
 FROM app.meli_store_connections WHERE id=$1::uuid RETURNING id::text`, store).Scan(&otherStore)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Exec(ctx, "DELETE FROM app.meli_store_connections WHERE id=$1::uuid", otherStore)
	checkIsolation := func(storeID string, scope OrderScope) {
		t.Helper()
		if _, err := s.GetOrderPurchase(ctx, owner, storeID, scope, "401"); !errors.Is(err, ErrOrderNotFound) {
			t.Fatalf("purchase read for %s %+v: %v", storeID, scope, err)
		}
		if _, err := s.SaveOrderPurchase(ctx, owner, storeID, scope, "401", purchaseID, "99", &winner.UpdatedAt); !errors.Is(err, ErrOrderNotFound) {
			t.Fatalf("purchase write for %s %+v: %v", storeID, scope, err)
		}
	}
	checkIsolation(otherStore, scope)
	checkIsolation(store, OrderScope{Site: "MLA", Logistic: "remote"})
	checkIsolation(store, OrderScope{Site: "MLB", Logistic: "fulfillment"})
	_, err = db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces=
 '[{"user_id":987655,"site_id":"MLB","logistic_type":"remote"}]'::jsonb WHERE id=$1::uuid`, store)
	if err != nil {
		t.Fatal(err)
	}
	checkIsolation(store, scope)
	if _, err = db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated)
 VALUES($1::uuid,'MLB','remote','987655','401','paid','USD',12.34,now(),now())`, store); err != nil {
		t.Fatal(err)
	}
	remapped, err := s.GetOrderPurchase(ctx, owner, store, scope, "401")
	if err != nil || remapped != nil {
		t.Fatalf("previous seller purchase leaked after remapping: %+v, %v", remapped, err)
	}
	newSellerOrder, err := s.GetOrder(ctx, owner, store, scope, "401")
	if err != nil || newSellerOrder.PurchaseOrderID != "" || newSellerOrder.PurchaseCostCNY != nil {
		t.Fatalf("remapped seller can see previous purchase: %+v, %v", newSellerOrder, err)
	}
	remappedLinked, err := s.ListOrders(ctx, owner, store, scope, "", "linked", 0)
	if err != nil || len(remappedLinked.Items) != 0 {
		t.Fatalf("remapped seller linked results: %+v, %v", remappedLinked, err)
	}
	if _, err = s.SaveOrderPurchase(ctx, owner, store, scope, "401", purchaseID, "99", &winner.UpdatedAt); !errors.Is(err, ErrOrderPurchaseConflict) {
		t.Fatalf("previous seller version accepted after remapping: %v", err)
	}
}
