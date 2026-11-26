package notifications

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func seedDiscovery(t *testing.T, ctx context.Context, p PostgresInbox) {
	t.Helper()
	seedStores(t, ctx, p)
	// TEST_DATABASE_URL is shared across packages. A sweep creates checkpoints
	// for every verified mapping in the database, so leftovers from other
	// packages would win the oldest-due reservation and break these assertions.
	mustExec(t, ctx, p, `DELETE FROM app.meli_store_connections WHERE id NOT IN ($1,$2)`,
		testStoreA, testStoreB)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces=
		'[{"user_id":"99011","site_id":"MLB","logistic_type":"remote"}]' WHERE id=$1`, testStoreA)
}

func TestDiscoveryPageCheckpointPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedDiscovery(t, ctx, p)
	defer cleanStores(t, ctx, p)
	calls := 0
	var from, to time.Time
	search := func(_ context.Context, scope DiscoveryScope, start, end time.Time, offset int) (DiscoveryPage, error) {
		calls++
		if scope.StoreID != testStoreA || scope.OwnerID != testOwnerA || scope.ParentID != testParentA ||
			scope.ChildID != testChildA || scope.Site != "MLB" || scope.Logistic != "remote" {
			t.Fatalf("wrong search scope: %+v", scope)
		}
		switch calls {
		case 1:
			from, to = start, end
			if offset != 0 || !start.Before(end) || end.Sub(start) < 23*time.Hour {
				t.Fatal("invalid first window")
			}
			ids := make([]string, 10)
			for i := range ids {
				ids[i] = fmt.Sprint(901 + i)
			}
			return DiscoveryPage{IDs: ids, Count: 3, Total: 4}, nil
		case 2:
			if offset != 3 || !start.Equal(from) || !end.Equal(to) {
				t.Fatal("page did not resume fixed window")
			}
			return DiscoveryPage{IDs: []string{"911"}, Count: 1, Total: 4}, nil
		case 3:
			if offset != 0 || !start.Equal(to.Add(-30*time.Minute)) || !end.After(to) {
				t.Fatal("completed window did not overlap")
			}
			return DiscoveryPage{IDs: []string{"911"}, Count: 1, Total: 1}, nil
		default:
			t.Fatal("unexpected search")
			return DiscoveryPage{}, nil
		}
	}
	mustExec(t, ctx, p, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated)
		VALUES($1,'MLB','remote',$2,'901','paid','USD',1,now(),now())`, testStoreA, testChildA)
	assertACK(t, testReceiver(t, p), 200, testOrder, "902", testChildA, "existing-webhook")
	if err := p.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
		AND next_offset=0 AND pending_ids=ARRAY['909','910'] AND pending_count=3 AND pending_more
		AND last_completed_at IS NULL`, testStoreA)
	assertCount(t, ctx, p, 6, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='compensation'`, testStoreA)
	// Simulate restart/retry with a fresh scheduler. A pending page does not
	// call the provider again, even if its results would have changed.
	due := func() {
		mustExec(t, ctx, p, `UPDATE app.meli_order_discovery SET next_due_at=now()-interval '1 second' WHERE store_id=$1`, testStoreA)
	}
	due()
	restarted := PostgresInbox{DB: p.DB}
	if err := restarted.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("pending page refetched")
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
		AND next_offset=3 AND cardinality(pending_ids)=0 AND window_to=$2`, testStoreA, to)
	due()
	if err := restarted.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
		AND next_offset=0 AND window_to IS NULL AND completed_to=$2 AND last_completed_at IS NOT NULL
		AND next_due_at>now()+interval '28 minutes'`, testStoreA, to)
	assertCount(t, ctx, p, 9, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='compensation'`, testStoreA)
	// Same resource in the overlap is already covered by an active job.
	due()
	// Let the second boundary pass without sleep by simulating an old watermark.
	mustExec(t, ctx, p, `UPDATE app.meli_order_discovery SET completed_to=completed_to-interval '3 minutes' WHERE store_id=$1`, testStoreA)
	to = to.Add(-3 * time.Minute)
	if err := restarted.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	if calls != 3 {
		t.Fatalf("search count=%d", calls)
	}
	assertCount(t, ctx, p, 9, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='compensation'`, testStoreA)
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_order_syncs WHERE store_id=$1`, testStoreA)
}

func TestDiscoveryBackpressureAndRevocationPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedDiscovery(t, ctx, p)
	defer cleanStores(t, ctx, p)
	defer mustExec(t, ctx, p, `DELETE FROM app.meli_order_read_coordination WHERE application_id=$1`, testApp)
	calls := 0
	search := func(_ context.Context, _ DiscoveryScope, _, _ time.Time, _ int) (DiscoveryPage, error) {
		calls++
		// Backlog becomes full AFTER the official search, not merely before it.
		mustExec(t, ctx, p, `INSERT INTO app.meli_notification_jobs
			(store_id,owner_user_id,parent_seller_id,application_id,notification_user_id,topic,resource_id)
			SELECT $1,$2,$3,$4,$5,'marketplace_orders',(7000+g)::text FROM generate_series(1,32) g`,
			testStoreA, testOwnerA, testParentA, testApp, testChildA)
		return DiscoveryPage{IDs: []string{"901", "902"}, Count: 1, Total: 1}, nil
	}
	if err := p.DiscoverySweep(ctx, testApp, search); !errors.Is(err, ErrCapacity) {
		t.Fatalf("capacity reported as success: %v", err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
		AND next_offset=0 AND pending_ids=ARRAY['901','902'] AND last_completed_at IS NULL`, testStoreA)
	mustExec(t, ctx, p, `DELETE FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `UPDATE app.meli_order_discovery SET next_due_at=now()-interval '1 second' WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `INSERT INTO app.meli_order_read_coordination(application_id,retry_until)
		VALUES($1,now()+interval '1 hour') ON CONFLICT(application_id) DO UPDATE SET retry_until=EXCLUDED.retry_until`, testApp)
	if err := p.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `UPDATE app.meli_order_read_coordination SET retry_until=now()-interval '1 second' WHERE application_id=$1`, testApp)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, testStoreA)
	if err := p.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='connected' WHERE id=$1`, testStoreA)
	if err := p.DiscoverySweep(ctx, testApp, search); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatal("pending page refetched after capacity or revocation")
	}
	assertCount(t, ctx, p, 2, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='compensation'`, testStoreA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1 AND window_to IS NULL
		AND cardinality(pending_ids)=0 AND last_completed_at IS NOT NULL`, testStoreA)
}

func TestDiscoveryRejectsInvalidPage(t *testing.T) {
	for _, p := range []DiscoveryPage{
		{IDs: []string{"1"}, Count: 0, Total: 1},
		{IDs: []string{"1"}, Count: 2, Total: 1},
		{IDs: []string{"0"}, Count: 1, Total: 1},
		{IDs: []string{"1", "1"}, Count: 1, Total: 1},
		{IDs: []string{"../1"}, Count: 1, Total: 1},
		{Count: 0, Total: 1},
	} {
		if err := validateDiscoveryPage(p, 0); err == nil {
			t.Fatalf("invalid page accepted: %+v", p)
		}
	}
}

func TestDiscoveryRewindsSameWindowPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedDiscovery(t, ctx, p)
	defer cleanStores(t, ctx, p)
	var fixedFrom, fixedTo time.Time
	first := func(_ context.Context, _ DiscoveryScope, from, to time.Time, offset int) (DiscoveryPage, error) {
		if offset != 0 {
			t.Fatal("unexpected initial offset")
		}
		fixedFrom, fixedTo = from, to
		return DiscoveryPage{IDs: []string{"901"}, Count: 1, Total: 2}, nil
	}
	if err := p.DiscoverySweep(ctx, testApp, first); err != nil {
		t.Fatal(err)
	}
	due := func() {
		mustExec(t, ctx, p, `UPDATE app.meli_order_discovery SET next_due_at=now()-interval '1 second' WHERE store_id=$1`, testStoreA)
	}
	due()
	if err := p.DiscoverySweep(ctx, testApp, func(_ context.Context, _ DiscoveryScope, from, to time.Time, offset int) (DiscoveryPage, error) {
		if offset != 1 || !from.Equal(fixedFrom) || !to.Equal(fixedTo) {
			t.Fatal("fixed search window changed")
		}
		return DiscoveryPage{}, ErrDiscoveryPageChanged
	}); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
		AND window_from=$2 AND window_to=$3 AND next_offset=0 AND last_completed_at IS NULL`, testStoreA, fixedFrom, fixedTo)
	due()
	if err := p.DiscoverySweep(ctx, testApp, first); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND source='compensation'`, testStoreA)
}

func TestDiscoveryRevokedDuringSearchPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedDiscovery(t, ctx, p)
	defer cleanStores(t, ctx, p)
	err := p.DiscoverySweep(ctx, testApp, func(_ context.Context, _ DiscoveryScope, _, _ time.Time, _ int) (DiscoveryPage, error) {
		mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, testStoreA)
		return DiscoveryPage{IDs: []string{"901"}, Count: 1, Total: 1}, nil
	})
	if !errors.Is(err, ErrScopeUnavailable) {
		t.Fatalf("revoked authorization accepted: %v", err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
		AND next_offset=0 AND window_to IS NOT NULL AND last_completed_at IS NULL`, testStoreA)
}
