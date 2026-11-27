package meli

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOperationsSummaryOwnerScopedAndAccuratePostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("PostgreSQL integration requires TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	var ownerA, ownerB, storeA, storeB string
	for _, target := range []*string{&ownerA, &ownerB, &storeA, &storeB} {
		if err := db.QueryRow(ctx, `SELECT gen_random_uuid()::text`).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.users(id) VALUES($1),($2)`, ownerA, ownerB); err != nil {
		t.Fatal(err)
	}
	defer db.Exec(ctx, `DELETE FROM app.users WHERE id IN ($1,$2)`, ownerA, ownerB)
	defer db.Exec(ctx, `DELETE FROM app.meli_store_connections WHERE id IN ($1,$2)`, storeA, storeB)
	mapping := `[{"user_id":"111","site_id":"MLB","logistic_type":"remote"}]`
	for _, row := range [][4]string{{storeA, ownerA, "900", "connected"}, {storeB, ownerB, "901", "connected"}} {
		if _, err := db.Exec(ctx, `INSERT INTO app.meli_store_connections
			(id,owner_user_id,seller_id,nickname,marketplaces,marketplaces_status,status,
			 access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at)
			VALUES($1,$2,$3,$4,$5::jsonb,'verified',$6,decode('aa','hex'),decode('aa','hex'),now()+interval '1 hour')`,
			row[0], row[1], row[2], "store-"+row[2], mapping, row[3]); err != nil {
			t.Fatal(err)
		}
	}
	// Fix the order creation time and the sales read clock to the same Beijing day.
	// The other operations assertions intentionally keep their live timestamps.
	salesTime := time.Date(2026, 11, 26, 16, 0, 0, 0, time.UTC)
	// Owner A has one row of each counted state plus rows that must be excluded.
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,updated_at)
		VALUES($1,'MLB','remote','111','901','paid','USD',10,$2,now(),now()),
		($1,'MLB','remote','111','902','cancelled','USD',10,$2,now(),now())`, storeA, salesTime); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_customer_notifications
		(store_id,owner_user_id,parent_seller_id,seller_id,site_id,topic,resource_id,related_id,status,direction,created_at,checked_at)
		VALUES($1,$2,'900','111','MLB','marketplace_questions','1','CBT1','UNANSWERED','',now(),now()),
		($1,$2,'900','111','MLB','marketplace_questions','2','CBT1','ANSWERED','',now(),now()),
		($1,$2,'900','111','MLB','marketplace_messages','aaaaaaaaaaaaaaaa','700','available','incoming',now()-interval '2 days',now()),
		($1,$2,'900','111','MLB','marketplace_messages','bbbbbbbbbbbbbbbb','700','available','outgoing',now()-interval '1 day',now())`,
		storeA, ownerA); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_claim_notifications
		(store_id,owner_user_id,parent_seller_id,seller_id,site_id,claim_id,resource_type,related_id,status,stage,claim_type,reason_code,last_updated,checked_at)
		VALUES($1,$2,'900','111','MLB','1','order','1','opened','claim','mediations','PNR1',now(),now()),
		($1,$2,'900','111','MLB','2','order','2','closed','claim','mediations','PNR1',now(),now())`, storeA, ownerA); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_item_notifications
		(store_id,owner_user_id,parent_seller_id,seller_id,site_id,item_id,status,last_updated,checked_at)
		VALUES($1,$2,'900','111','MLB','MLB1','paused',now(),now()),
		($1,$2,'900','111','MLB','MLB2','active',now(),now())`, storeA, ownerA); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_notification_jobs
		(store_id,owner_user_id,parent_seller_id,application_id,notification_user_id,topic,resource_id,dead_at)
		VALUES($1,$2,'900','test-client','900','marketplace_orders','1',NULL),
		($1,$2,'900','test-client','111','marketplace_orders','2',NULL),
		($1,$2,'900','test-client','111','marketplace_orders','3',now())`, storeA, ownerA); err != nil {
		t.Fatal(err)
	}
	svc := &Service{db: db, cfg: Config{ClientID: "test-client"}}
	got, err := svc.OperationsSummary(ctx, ownerA)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Stores) != 1 {
		t.Fatalf("stores=%d", len(got.Stores))
	}
	s := got.Stores[0]
	if !s.Connected || s.PaidOrders != 1 || s.CachedUnanswered == nil || *s.CachedUnanswered != 1 ||
		s.PendingMessages == nil || *s.PendingMessages != 0 || s.OpenClaims == nil || *s.OpenClaims != 1 ||
		s.PausedItems == nil || *s.PausedItems != 1 || s.PendingJobs != 2 || s.DeadJobs != 1 {
		t.Fatalf("summary mismatch: %+v", s)
	}
	other, err := svc.OperationsSummary(ctx, ownerB)
	if err != nil {
		t.Fatal(err)
	}
	if len(other.Stores) != 1 || other.Stores[0].StoreID != storeB || other.Stores[0].PaidOrders != 0 {
		t.Fatalf("owner isolation failed: %+v", other.Stores)
	}
	empty, err := svc.OperationsSummary(ctx, "00000000-0000-4000-8000-000000000099")
	if err != nil {
		t.Fatal(err)
	}
	if len(empty.Stores) != 0 || !empty.CacheOnly {
		t.Fatalf("empty state failed: %+v", empty)
	}
	if other.Stores[0].CachedUnanswered != nil || other.Stores[0].OpenClaims != nil {
		t.Fatal("no observations were displayed as zero")
	}
	overview, err := svc.OperationsOverview(ctx, ownerA, storeA, "MLB")
	if err != nil {
		t.Fatal(err)
	}
	// Deliberately different raw total: the official USD merchandise line wins.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET order_items='[{"item_id":"MLB1","quantity":2,"unit_price":"3.25","currency_id":"USD"}]'::jsonb WHERE store_id=$1 AND order_id='901'`, storeA); err != nil {
		t.Fatal(err)
	}
	// The public overview still wires sales to its own current read date.
	day, _, _ := salesDay(overview.CheckedAt)
	if overview.Sales.Day != day || !overview.Sales.ScopeAvailable {
		t.Fatalf("overview sales scope or day: %+v", overview.Sales)
	}
	// Test the same production sales query at a fixed instant without adding
	// a production clock override or relying on a test not crossing midnight.
	fixedSummary := got
	fixedSummary.CheckedAt = salesTime
	sales, err := svc.operationsSales(ctx, ownerA, storeA, "MLB", fixedSummary)
	if err != nil || sales.AmountUSD == nil || *sales.AmountUSD != "6.50" || sales.IncludedOrders != 1 || sales.CancelledOrders != 1 {
		t.Fatalf("daily cached merchandise sales: %+v %v", sales, err)
	}
	// Synthetic confirmed net: production never substitutes the fee subtotal.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET billing_snapshot=jsonb_build_object(
		'shipment_id','','preview',jsonb_build_object('order_id','901','available',true,'checked_at',now()))
		WHERE store_id=$1 AND order_id='901'`, storeA); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_order_purchases
		(store_id,site_id,logistic_type,seller_id,order_id,purchase_order_id,cost_cny)
		VALUES($1,'MLB','remote','111','901','12345678',30)`, storeA); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.daily_usd_cny_rates
		(accounting_day,observation_day,raw_rate,locked_rate,source)
		VALUES('2026-11-27','2026-11-26',6.68,6.6,'frankfurter_v2_ecb') ON CONFLICT DO NOTHING`); err != nil {
		t.Fatal(err)
	}
	ps := OrderScope{Site:"MLB",Logistic:"remote"}
	initial, err := svc.GetExpectedOrderProfit(ctx,ownerA,storeA,ps,"901")
	if err!=nil || initial.Profit!=nil || !initial.CanConfirm || initial.FX==nil {
		t.Fatalf("initial expected profit: %+v %v",initial,err)
	}
	confirmed, err := svc.ConfirmExpectedOrderProfit(ctx,ownerA,storeA,ps,"901",initial.Revision,"7.90","2",false)
	if err!=nil || confirmed.Profit==nil || confirmed.Profit.Kind!="expected" {
		t.Fatalf("confirmed expected profit: %+v %v",confirmed,err)
	}
	rollup, err := svc.operationsSales(ctx,ownerA,storeA,"MLB",fixedSummary)
	if err!=nil || rollup.ProfitOrders!=1 || rollup.ExpectedProfitCNY==nil || *rollup.ExpectedProfitCNY!=confirmed.Profit.ProfitCNY {
		t.Fatalf("profit rollup: %+v %v",rollup,err)
	}
	if _, err := svc.GetExpectedOrderProfit(ctx,ownerB,storeA,ps,"901"); err==nil {
		t.Fatal("foreign owner read profit")
	}
	if _, err := svc.ConfirmExpectedOrderProfit(ctx,ownerB,storeA,ps,"901",confirmed.Revision,"9","0",false); err==nil {
		t.Fatal("foreign owner wrote confirmation")
	}
	if _, err := svc.ConfirmExpectedOrderProfit(ctx,ownerA,storeA,ps,"901",initial.Revision,"9","0",false); err==nil {
		t.Fatal("stale confirmation revision accepted")
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET billing_snapshot=jsonb_set(billing_snapshot,'{preview,charges}','[{"type":"CHARGE","amount":"1"}]') WHERE store_id=$1 AND order_id='901'`,storeA); err!=nil {
		t.Fatal(err)
	}
	stale,err:=svc.GetExpectedOrderProfit(ctx,ownerA,storeA,ps,"901")
	if err!=nil || stale.Profit!=nil || stale.Reason!="confirmation_stale" {
		t.Fatalf("changed evidence reused confirmation: %+v %v",stale,err)
	}
	if _, err := svc.ConfirmExpectedOrderProfit(ctx,ownerA,storeA,ps,"901",confirmed.Revision,"9","0",false); err==nil {
		t.Fatal("changed billing evidence accepted old revision")
	}
	if len(overview.Activity) != 9 || len(overview.Todos) != 3 {
		t.Fatalf("filtered activity/todos: %+v", overview)
	}
	for _, row := range overview.Todos {
		if row.StoreID != storeA || row.Site != "MLB" {
			t.Fatalf("out-of-scope todo: %+v", row)
		}
	}
	// A cached reply closes the conversation; a new incoming message reopens it.
	// A delayed older incoming observation cannot change the latest message time.
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_customer_notifications
		(store_id,owner_user_id,parent_seller_id,seller_id,site_id,topic,resource_id,related_id,status,direction,created_at,checked_at)
		VALUES($1,$2,'900','111','MLB','marketplace_messages','cccccccccccccccc','700','available','incoming',now()-interval '3 days',now())`, storeA, ownerA); err != nil {
		t.Fatal(err)
	}
	processed, err := svc.OperationsOverview(ctx, ownerA, storeA, "MLB")
	if err != nil || len(processed.Todos) != 3 || processed.Stores[0].PendingMessages == nil || *processed.Stores[0].PendingMessages != 0 {
		t.Fatalf("late old message reopened replied conversation: %+v %v", processed, err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_customer_notifications
		(store_id,owner_user_id,parent_seller_id,seller_id,site_id,topic,resource_id,related_id,status,direction,created_at,checked_at)
		VALUES($1,$2,'900','111','MLB','marketplace_messages','dddddddddddddddd','700','available','incoming',now(),now()),
		($1,$2,'900','111','MLB','marketplace_messages','eeeeeeeeeeeeeeee','700','available','incoming',now(),now())`, storeA, ownerA); err != nil {
		t.Fatal(err)
	}
	reopened, err := svc.OperationsOverview(ctx, ownerA, storeA, "MLB")
	if err != nil || len(reopened.Todos) != 4 || reopened.Stores[0].PendingMessages == nil || *reopened.Stores[0].PendingMessages != 1 {
		t.Fatalf("new messages did not reopen exactly one conversation: %+v %v", reopened, err)
	}
	for _, row := range reopened.Todos {
		if row.Kind == "message" && (row.ResourceID != "700" || row.Status != "pending_reply") {
			t.Fatalf("incorrect conversation todo: %+v", row)
		}
	}
	foreign, err := svc.OperationsOverview(ctx, ownerB, storeA, "")
	if err != nil || len(foreign.Activity) != 0 || len(foreign.Todos) != 0 || foreign.Sales.ScopeAvailable || foreign.Sales.AmountUSD != nil {
		t.Fatalf("foreign operations rows: %+v %v", foreign, err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, storeA); err != nil {
		t.Fatal(err)
	}
	revoked, err := svc.OperationsSummary(ctx, ownerA)
	if err != nil {
		t.Fatal(err)
	}
	s = revoked.Stores[0]
	if s.Connected || s.PaidOrders != 0 || s.CachedUnanswered != nil || s.PendingMessages != nil ||
		s.OpenClaims != nil || s.PausedItems != nil || s.PendingJobs != 0 || s.DeadJobs != 0 {
		t.Fatalf("revoked authorization leaked cached metrics: %+v", s)
	}
}
