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
	// Owner A has one row of each counted state plus rows that must be excluded.
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,updated_at)
		VALUES($1,'MLB','remote','111','901','paid','USD',10,now(),now(),now()),
		($1,'MLB','remote','111','902','cancelled','USD',10,now(),now(),now())`, storeA); err != nil {
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
	if err != nil || len(foreign.Activity) != 0 || len(foreign.Todos) != 0 {
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
