package notifications

import "testing"

func TestKnownResourceCompensation(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces=
		'[{"user_id":"99011","site_id":"MLB","logistic_type":"remote"}]' WHERE id=$1`, testStoreA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,updated_at)
		VALUES ($1,'MLB','remote',$2,'901','paid','USD',10,now()-interval '1 day',now()-interval '2 hours','991',now()-interval '2 hours'),
		($1,'MLB','remote',$2,'902','paid','USD',10,now()-interval '1 day',now()-interval '2 hours','991',now()-interval '2 hours'),
		($1,'MLB','remote','99999','903','paid','USD',10,now()-interval '1 day',now()-interval '2 hours','992',now()-interval '2 hours'),
		($1,'MLB','remote',$2,'904','cancelled','USD',10,now()-interval '1 day',now()-interval '2 hours','993',now()-interval '2 hours')`, testStoreA, testChildA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_order_read_coordination(application_id,retry_until)
		VALUES($1,now()+interval '1 hour') ON CONFLICT(application_id) DO UPDATE SET retry_until=EXCLUDED.retry_until`, testApp)
	defer mustExec(t, ctx, p, `DELETE FROM app.meli_order_read_coordination WHERE application_id=$1`, testApp)
	if err := p.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `UPDATE app.meli_order_read_coordination SET retry_until=now()-interval '1 second' WHERE application_id=$1`, testApp)
	for i := 0; i < 2; i++ {
		if err := p.CompensationSweep(ctx, testApp); err != nil {
			t.Fatal(err)
		}
	}
	assertCount(t, ctx, p, 3, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='compensation'`, testStoreA)
	assertCount(t, ctx, p, 3, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND generation=1`, testStoreA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND topic='marketplace_shipments' AND resource_id='991'`, testStoreA)
	mustExec(t, ctx, p, `DELETE FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
	if err := p.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
}

func TestCompensationCooldownBacklogAndRestart(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces=
		'[{"user_id":"99011","site_id":"MLB","logistic_type":"remote"}]' WHERE id=$1`, testStoreA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,updated_at)
		VALUES ($1,'MLB','remote',$2,'910','paid','USD',10,now()-interval '1 day',now()-interval '2 hours','',now()-interval '2 hours')`,
		testStoreA, testChildA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_notification_cooldowns
		(store_id,owner_user_id,parent_seller_id,application_id,available_at)
		VALUES($1,$2,$3,$4,now()+interval '1 hour')`, testStoreA, testOwnerA, testParentA, testApp)
	if err := p.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `DELETE FROM app.meli_notification_cooldowns WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_notification_jobs
		(store_id,owner_user_id,parent_seller_id,application_id,notification_user_id,topic,resource_id)
		SELECT $1,$2,$3,$4,$5,'marketplace_orders',(7000+g)::text FROM generate_series(1,32) g`,
		testStoreA, testOwnerA, testParentA, testApp, testChildA)
	if err := p.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `DELETE FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
	// A fresh process value has no in-memory cursor; the same persisted rows are due.
	restarted := PostgresInbox{DB: p.DB}
	if err := restarted.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND resource_id='910' AND source='compensation'`, testStoreA)
	if err := restarted.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='910'`, testStoreA)
}

func TestKnownPaymentPendingOrderCompensation(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces=
		'[{"user_id":"99011","site_id":"MLB","logistic_type":"remote"}]' WHERE id=$1`, testStoreA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,updated_at)
		VALUES($1,'MLB','remote',$2,'901','payment_required','USD',10,now()-interval '2 hours',
		 now()-interval '2 hours','991',now()-interval '2 hours')`, testStoreA, testChildA)
	if err := p.CompensationSweep(ctx, testApp); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1
		AND topic='marketplace_orders' AND resource_id='901'`, testStoreA)
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1
		AND topic='marketplace_shipments'`, testStoreA)
}
