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
    scope := OrderScope{Site:"MLB", Logistic:"remote"}
    _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated)
 VALUES($1::uuid,'MLB','remote','987654','401','paid','USD',12.34,now(),now())`,store)
    if err!=nil { t.Fatal(err) }
    missing,err:=s.GetOrderPurchase(ctx,owner,store,scope,"401")
    if err!=nil || missing!=nil { t.Fatalf("new order purchase=%+v error=%v",missing,err) }
    const purchaseID="1234567890123456"
    if _,err=s.SaveOrderPurchase(ctx,owner,store,scope,"401",purchaseID,"16.00");err!=nil {
        t.Fatal(err)
    }
    got,err:=s.GetOrderPurchase(ctx,owner,store,scope,"401")
    if err!=nil || got==nil || got.PurchaseOrderID!=purchaseID || got.CostCNY==nil || *got.CostCNY!="16" || got.Source!="manual" {
        t.Fatalf("purchase=%+v error=%v",got,err)
    }
    if _,err=s.SaveOrderPurchase(ctx,owner,store,scope,"401",purchaseID,"16.123");!errors.Is(err,ErrOrderPurchaseInvalid) {
        t.Fatalf("invalid cost accepted: %v",err)
    }
    if _,err=s.SaveOrderPurchase(ctx,owner,store,scope,"401","bad-number","");!errors.Is(err,ErrOrderPurchaseInvalid) {
        t.Fatalf("invalid purchase ID accepted: %v",err)
    }
    if _,err=s.SaveOrderPurchase(ctx,"550e8400-e29b-41d4-a716-446655440000",store,scope,"401",purchaseID,"99");!errors.Is(err,ErrOrderNotFound) {
        t.Fatalf("cross owner write: %v",err)
    }
    if _,err=s.GetOrderPurchase(ctx,owner,store,OrderScope{Site:"MLA",Logistic:"remote"},"401");!errors.Is(err,ErrOrderNotFound) {
        t.Fatalf("cross site read: %v",err)
    }
    _,err=db.Exec(ctx,`UPDATE app.meli_store_connections SET marketplaces=
 '[{"user_id":987655,"site_id":"MLB","logistic_type":"remote"}]'::jsonb WHERE id=$1::uuid`,store)
    if err!=nil {t.Fatal(err)}
    if _,err=s.GetOrderPurchase(ctx,owner,store,scope,"401");!errors.Is(err,ErrOrderNotFound) {
        t.Fatalf("old seller purchase visible: %v",err)
    }
    if _,err=s.SaveOrderPurchase(ctx,owner,store,scope,"401",purchaseID,"99");!errors.Is(err,ErrOrderNotFound) {
        t.Fatalf("old seller purchase overwritten: %v",err)
    }
}
