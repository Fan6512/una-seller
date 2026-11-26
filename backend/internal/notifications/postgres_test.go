package notifications

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	testApp      = "2069392825111111"
	testOrder    = "marketplace_orders"
	testShipment = "marketplace_shipments"
	testOwnerA   = "99000000-0000-4000-8000-000000000001"
	testOwnerB   = "99000000-0000-4000-8000-000000000002"
	testStoreA   = "99000000-0000-4000-8000-000000000011"
	testStoreB   = "99000000-0000-4000-8000-000000000012"
	testParentA  = "99001"
	testParentB  = "99002"
	testChildA   = "99011"
	testChildB   = "99012"
)

func testDatabase(t *testing.T) (context.Context, PostgresInbox) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires an isolated TEST_DATABASE_URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	t.Cleanup(cancel)
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return ctx, PostgresInbox{DB: db}
}

func seedStores(t *testing.T, ctx context.Context, p PostgresInbox) {
	t.Helper()
	if err := migrations.Apply(ctx, p.DB); err != nil {
		t.Fatal(err)
	}
	cleanStores(t, ctx, p)
	mustExec(t, ctx, p, `INSERT INTO app.users(id) VALUES($1),($2)`, testOwnerA, testOwnerB)
	for _, row := range [][4]string{
		{testStoreA, testOwnerA, testParentA, testChildA},
		{testStoreB, testOwnerB, testParentB, testChildB},
	} {
		mustExec(t, ctx, p, `INSERT INTO app.meli_store_connections
			(id,owner_user_id,seller_id,marketplaces,marketplaces_status,
			 access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at)
			VALUES($1,$2,$3,jsonb_build_array(jsonb_build_object('user_id',$4::text)),
			       'verified',decode('aa','hex'),decode('aa','hex'),now()+interval '1 hour')`,
			row[0], row[1], row[2], row[3])
	}
}

func cleanStores(t *testing.T, ctx context.Context, p PostgresInbox) {
	t.Helper()
	mustExec(t, ctx, p, `DELETE FROM app.meli_store_connections WHERE id IN ($1,$2)`, testStoreA, testStoreB)
	mustExec(t, ctx, p, `DELETE FROM app.users WHERE id IN ($1,$2)`, testOwnerA, testOwnerB)
}

func mustExec(t *testing.T, ctx context.Context, p PostgresInbox, query string, args ...any) {
	t.Helper()
	if _, err := p.DB.Exec(ctx, query, args...); err != nil {
		t.Fatal(err)
	}
}

func assertCount(t *testing.T, ctx context.Context, p PostgresInbox, want int, query string, args ...any) {
	t.Helper()
	var got int
	if err := p.DB.QueryRow(ctx, query, args...).Scan(&got); err != nil || got != want {
		t.Fatalf("count=%d want=%d err=%v\n%s", got, want, err, query)
	}
}

func eventBody(topic, resource, seller, event string) string {
	path := "orders"
	switch topic {
	case testShipment:
		path = "shipments"
	case "marketplace_questions":
		path = "questions"
	case "marketplace_messages":
		path = "messages"
	case "marketplace_claims":
		path = "claims"
	case "marketplace_items":
		path = "items"
	}
	return fmt.Sprintf(`{"_id":%q,"user_id":%q,"application_id":%q,"topic":%q,"resource":"/marketplace/%s/%s","received":"2026-01-01T01:00:00Z"}`,
		event, seller, testApp, topic, path, resource)
}

func testReceiver(t *testing.T, p PostgresInbox) http.Handler {
	t.Helper()
	h, err := NewReceiver(testApp, token, p)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func assertACK(t *testing.T, h http.Handler, want int, topic, resource, seller, event string) {
	t.Helper()
	r := httptest.NewRequest(http.MethodPost, Path, strings.NewReader(eventBody(topic, resource, seller, event)))
	r.Header.Set(ingressHeader, token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != want {
		t.Fatalf("event %s received HTTP %d, want %d", event, w.Code, want)
	}
}

func TestNotificationPostgresACKAndCapacity(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second

	measure := func(name string, distinct bool) {
		t.Helper()
		const requests, concurrency = 96, 6
		durations := make([]time.Duration, requests)
		var wg sync.WaitGroup
		for worker := 0; worker < concurrency; worker++ {
			wg.Go(func() {
				for i := worker; i < requests; i += concurrency {
					event := name
					if distinct {
						event = fmt.Sprintf("%s-%d", name, i)
					}
					req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL+Path,
						strings.NewReader(eventBody(testOrder, "999", testChildA, event)))
					if err != nil {
						t.Error(err)
						return
					}
					req.Header.Set(ingressHeader, token)
					start := time.Now()
					resp, err := client.Do(req)
					durations[i] = time.Since(start)
					if err != nil {
						t.Error(err)
						return
					}
					resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						t.Errorf("event %s received HTTP %d", event, resp.StatusCode)
					}
				}
			})
		}
		wg.Wait()
		slices.Sort(durations)
		t.Logf("loopback HTTP + PostgreSQL ACK %s n=%d concurrency=%d p50=%s p95=%s p99=%s max=%s",
			name, requests, concurrency, durations[(requests*50+99)/100-1], durations[(requests*95+99)/100-1], durations[(requests*99+99)/100-1], durations[requests-1])
		if durations[requests-1] >= 500*time.Millisecond {
			t.Error("isolated ACK exceeded 500 ms")
		}
	}
	measure("concurrent-retry", false)
	measure("distinct-receipt", true)
	assertACK(t, h, 200, testOrder, "999", testParentA, "parent")
	assertACK(t, h, 200, testOrder, "999", testChildB, "concurrent-retry")
	assertCount(t, ctx, p, 98, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreB)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE store_id=$1 AND owner_user_id=$2 AND generation=98 AND notification_user_id=$3`, testStoreA, testOwnerA, testParentA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE store_id=$1 AND owner_user_id=$2 AND generation=1`, testStoreB, testOwnerB)

	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces='[{"user_id":"99011"}]' WHERE id=$1`, testStoreB)
	assertACK(t, h, 403, testOrder, "999", testChildA, "ambiguous")
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces='[{"user_id":"99012"}]' WHERE id=$1`, testStoreB)

	// A busy admission lock must not ACK or persist an event before commit.
	lockTx, err := p.DB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer lockTx.Rollback(ctx)
	if _, err := lockTx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, int64(queueAdmissionLock)); err != nil {
		t.Fatal(err)
	}
	assertACK(t, h, 503, testShipment, "991", testChildA, "admission-blocked")
	if err := lockTx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND resource_id='991'`, testStoreA)
	assertACK(t, h, 200, testShipment, "991", testChildA, "admission-blocked")
	assertACK(t, h, 200, testShipment, "991", testChildA, "admission-blocked")
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1 AND resource_id='991'`, testStoreA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='991' AND generation=1`, testStoreA)
	// Keep the existing capacity fixture's baseline unchanged.
	mustExec(t, ctx, p, `DELETE FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='991'`, testStoreA)
	mustExec(t, ctx, p, `DELETE FROM app.meli_notification_inbox WHERE store_id=$1 AND resource_id='991'`, testStoreA)

	// Let the admission precheck read the queue, then block its write transaction.
	tx, err := p.DB.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `LOCK app.meli_notification_jobs IN SHARE MODE`); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	assertACK(t, h, 503, testOrder, "999", testChildA, "blocked")
	t.Logf("PostgreSQL lock failure HTTP 503 after %s", time.Since(start))
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 98, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)

	mustExec(t, ctx, p, `INSERT INTO app.meli_notification_inbox
		(store_id,owner_user_id,parent_seller_id,application_id,notification_user_id,topic,resource_id,event_key,source_received_at,enqueued_at)
		SELECT $1,$2,'99001',$3,'99011','marketplace_orders',(1000000+n)::text,repeat('a',64),now(),now()-interval '8 days'
		FROM generate_series(1,$4-(SELECT count(*)::int FROM app.meli_notification_inbox)) n`,
		testStoreA, testOwnerA, testApp, queueCapacity)
	assertACK(t, h, 200, testOrder, "999", testChildA, "concurrent-retry")
	assertACK(t, h, 503, testOrder, "998", testChildA, "over-capacity")
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='998'`, testStoreA)
	if err := p.prune(ctx); err != nil {
		t.Fatal(err)
	}
	assertACK(t, h, 200, testOrder, "998", testChildA, "over-capacity")
	assertCount(t, ctx, p, queueCapacity-499, `SELECT count(*) FROM app.meli_notification_inbox`)
	mustExec(t, ctx, p, `INSERT INTO app.meli_notification_jobs
	    (store_id,owner_user_id,parent_seller_id,application_id,notification_user_id,topic,resource_id,dead_at)
	    SELECT $1,$2,'99001',$3,'99011','marketplace_orders',(1000000+n)::text,now()
	    FROM generate_series(1,$4-(SELECT count(*)::int FROM app.meli_notification_jobs)) n`,
		testStoreA, testOwnerA, testApp, queueCapacity)
	assertACK(t, h, 200, testOrder, "998", testChildA, "join-active-at-capacity")
	assertACK(t, h, 503, testOrder, "997", testChildA, "new-job-over-capacity")
	assertCount(t, ctx, p, queueCapacity, `SELECT count(*) FROM app.meli_notification_jobs`)
	assertCount(t, ctx, p, queueCapacity-498, `SELECT count(*) FROM app.meli_notification_inbox`)
}

func TestNotificationPostgresAdmissionScopeAndStats(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)

	assertACK(t, h, 403, testOrder, "998", "999999", "unknown")
	for _, status := range []string{"reauth_required", "refresh_unknown", "disconnected"} {
		mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status=$1 WHERE id=$2`, status, testStoreA)
		assertACK(t, h, 503, testOrder, "998", testParentA, "retry")
		assertACK(t, h, 503, testOrder, "998", testChildA, "retry")
	}
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='connected',marketplaces_status='unverified' WHERE id=$1`, testStoreA)
	assertACK(t, h, 503, testOrder, "998", testChildA, "retry")
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET marketplaces_status='verified' WHERE id=$1`, testStoreA)
	assertACK(t, h, 200, testOrder, "998", testChildA, "retry")

	const otherApp = "2069392825111112"
	event, err := parse(strings.NewReader(strings.Replace(eventBody(testOrder, "998", testChildA, "other-app"), testApp, otherApp, 1)), otherApp)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Enqueue(ctx, event); err != nil {
		t.Fatal(err)
	}
	for _, app := range []string{testApp, otherApp} {
		stats, err := p.stats(ctx, app, []string{testOrder, testShipment})
		if err != nil || stats.Receipts != 1 || stats.GlobalReceipts < 2 || stats.Pending != 1 {
			t.Fatalf("application receipt counts not isolated: %+v %v", stats, err)
		}
	}
}

func TestNotificationPostgresWorker(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	var applied, prepared int
	result := func(_ context.Context, j Job) (VerifiedResource, error) {
		prepared++
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { applied++; return nil }}, nil
	}
	orderWorker := NewWorker(p, testApp, map[string]Prepare{testOrder: result})
	bothWorker := NewWorker(p, testApp, map[string]Prepare{testOrder: result, testShipment: result})
	run := func(w *Worker) {
		t.Helper()
		if err := w.runOnce(ctx); err != nil {
			t.Fatal(err)
		}
	}
	assertNoClaim := func(w *Worker) {
		t.Helper()
		if _, err := w.claim(ctx); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("expected no eligible job, got %v", err)
		}
	}
	assertACK(t, h, 200, testShipment, "897", testChildA, "unregistered")
	assertNoClaim(NewWorker(p, testApp, nil))
	assertNoClaim(orderWorker)
	stats, err := p.stats(ctx, testApp, orderWorker.topics)
	if err != nil || stats.Unregistered != 1 || stats.Pending != 1 {
		t.Fatalf("unexpected backlog: %+v, %v", stats, err)
	}

	assertACK(t, h, 200, testOrder, "898", testChildA, "lease")
	var wg sync.WaitGroup
	leases := make(chan Job, 2)
	for range 2 {
		wg.Go(func() {
			j, err := orderWorker.claim(ctx)
			if err == nil {
				leases <- j
			} else if !errors.Is(err, pgx.ErrNoRows) {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	close(leases)
	if len(leases) != 1 {
		t.Fatalf("concurrent claims=%d", len(leases))
	}
	old := <-leases
	assertACK(t, h, 200, testOrder, "898", testParentA, "during-lease")
	mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET lease_until=now()-interval '1 second' WHERE id=$1`, old.ID)
	fresh, err := orderWorker.claim(ctx)
	if err != nil || fresh.ID != old.ID || fresh.leaseToken == old.leaseToken || fresh.generation != 2 || fresh.attempts != 1 {
		t.Fatalf("lease not recovered: %+v %v", fresh, err)
	}
	verified, _ := result(ctx, old)
	if err := orderWorker.finish(ctx, old, verified); !errors.Is(err, errLeaseLost) || applied != 0 {
		t.Fatalf("stale fence allowed writes: %d %v", applied, err)
	}
	if err := orderWorker.retry(ctx, old, &RateLimited{RetryAfter: time.Hour}); !errors.Is(err, errLeaseLost) {
		t.Fatalf("stale fence allowed retries: %v", err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_cooldowns WHERE store_id=$1`, testStoreA)

	// A failed fence at commit rolls back writes made by Apply.
	verified.Apply = func(ctx context.Context, tx LocalWriter) error {
		_, err := tx.Exec(ctx, `INSERT INTO app.users(id) VALUES('99000000-0000-4000-8000-000000000099')`)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE app.meli_notification_jobs SET lease_until=now()-interval '1 second' WHERE id=$1`, fresh.ID)
		return err
	}
	if err := orderWorker.finish(ctx, fresh, verified); !errors.Is(err, errLeaseLost) {
		t.Fatalf("expected fenced rollback, got %v", err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.users WHERE id='99000000-0000-4000-8000-000000000099'`)

	assertACK(t, h, 200, testOrder, "898", testChildA, "during-current-lease")
	verified, _ = result(ctx, fresh)
	if err := orderWorker.finish(ctx, fresh, verified); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE id=$1 AND generation=3 AND attempts=0 AND lease_token IS NULL`, fresh.ID)
	run(orderWorker)
	assertACK(t, h, 200, testOrder, "898", testChildA, "lease")
	if err := migrations.Apply(ctx, p.DB); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE id=$1 OR (store_id=$2 AND topic=$3)`, fresh.ID, testStoreA, testOrder)

	assertACK(t, h, 200, testOrder, "899", testChildA, "auth-scope")
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, testStoreA)
	before := prepared
	run(orderWorker)
	if prepared != before {
		t.Fatal("revoked store reached handler")
	}
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='connected' WHERE id=$1`, testStoreA)
	ready := func() {
		t.Helper()
		mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET available_at=now()
			WHERE store_id=$1 AND resource_id='899'`, testStoreA)
	}
	ready()
	before = applied
	foreign := NewWorker(p, testApp, map[string]Prepare{testOrder: func(ctx context.Context, j Job) (VerifiedResource, error) {
		resource, err := result(ctx, j)
		resource.SellerID = testChildB
		return resource, err
	}})
	run(foreign)
	if applied != before {
		t.Fatal("foreign seller reached write")
	}
	ready()
	revokedDuringRead := NewWorker(p, testApp, map[string]Prepare{testOrder: func(ctx context.Context, j Job) (VerifiedResource, error) {
		mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='disconnected' WHERE id=$1`, testStoreA)
		return result(ctx, j)
	}})
	run(revokedDuringRead)
	if applied != before {
		t.Fatal("revoked seller reached write after read")
	}
	mustExec(t, ctx, p, `UPDATE app.meli_store_connections SET status='connected' WHERE id=$1`, testStoreA)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE store_id=$1 AND resource_id='899' AND attempts=1 AND last_failure='scope_unavailable'
		  AND available_at BETWEEN now()+interval '58 seconds' AND now()+interval '61 seconds'`, testStoreA)

	ready()
	limited := NewWorker(p, testApp, map[string]Prepare{testOrder: func(context.Context, Job) (VerifiedResource, error) {
		return VerifiedResource{}, &RateLimited{RetryAfter: 2 * time.Minute}
	}})
	run(limited)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_cooldowns
		WHERE store_id=$1 AND available_at >= now()+interval '118 seconds'`, testStoreA)
	assertNoClaim(bothWorker)
	assertACK(t, h, 200, testShipment, "897", testChildB, "other-store")
	run(bothWorker)
	if applied != before+1 {
		t.Fatal("cooldown blocked another store")
	}
	mustExec(t, ctx, p, `UPDATE app.meli_notification_cooldowns SET available_at=now() WHERE store_id=$1`, testStoreA)
	for range 7 {
		ready()
		run(NewWorker(p, testApp, map[string]Prepare{testOrder: func(context.Context, Job) (VerifiedResource, error) {
			return VerifiedResource{}, errors.New("synthetic failure")
		}}))
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE store_id=$1 AND dead_at IS NOT NULL AND attempts=8 AND last_failure='handler_failed'`, testStoreA)
	assertNoClaim(orderWorker)
	assertACK(t, h, 200, testOrder, "899", testChildA, "auth-scope")
	assertNoClaim(orderWorker)
	assertACK(t, h, 200, testOrder, "899", testChildA, "after-dead")
	run(orderWorker)
	if applied != before+2 {
		t.Fatal("fresh delivery after dead letter was not processed")
	}
	assertACK(t, h, 200, testOrder, "999", testChildA, "exhausted-lease")
	last, err := orderWorker.claim(ctx)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs
		SET attempts=8,lease_until=now()-interval '1 second' WHERE id=$1`, last.ID)
	assertNoClaim(orderWorker)
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE id=$1 AND dead_at IS NOT NULL AND last_failure='lease_expired'`, last.ID)
	mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET dead_at=now()-interval '31 days' WHERE store_id=$1 AND dead_at IS NOT NULL`, testStoreA)
	if err := p.prune(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND topic=$2 AND attempts=0`, testStoreA, testShipment)
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND topic=$2`, testStoreA, testOrder)
}

func TestNotificationPostgresRestart(t *testing.T) {
	phase := os.Getenv("TEST_RESTART_PHASE")
	if phase == "" {
		t.Skip("run before, down and after a real PostgreSQL restart")
	}
	ctx, p := testDatabase(t)
	h := testReceiver(t, p)
	w := NewWorker(p, testApp, map[string]Prepare{testOrder: func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { return nil }}, nil
	}})
	switch phase {
	case "before":
		seedStores(t, ctx, p)
		assertACK(t, h, 200, testOrder, "888", testChildA, "crash")
		j, err := w.claim(ctx)
		if err != nil || j.attempts != 1 {
			t.Fatalf("initial lease: %+v %v", j, err)
		}
		assertACK(t, h, 200, testShipment, "889", testChildA, "unregistered")
		assertACK(t, h, 200, testOrder, "890", testChildB, "cooldown")
		cooling, err := w.claim(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := w.retry(ctx, cooling, &RateLimited{RetryAfter: time.Hour}); err != nil {
			t.Fatal(err)
		}
		// Preserve an actual partially drained discovery page across the same
		// database process restart. No provider call is needed to resume IDs.
		mustExec(t, ctx, p, `INSERT INTO app.meli_order_discovery
			(store_id,owner_user_id,parent_seller_id,application_id,site_id,logistic_type,seller_id,
			 completed_to,window_from,window_to,next_offset,pending_ids,pending_count,pending_more)
			VALUES($1,$2,$3,$4,'MLB','remote',$5,'2026-11-20T00:00:00Z','2026-11-19T23:30:00Z',
			 '2026-11-20T01:00:00Z',3,ARRAY['901','902'],2,true)`, testStoreA, testOwnerA, testParentA, testApp, testChildA)
		// The shared application-wide coordinator must also survive a true restart.
		mustExec(t, ctx, p, `INSERT INTO app.meli_order_read_coordination(application_id,retry_until)
VALUES($1,now()+interval '1 hour') ON CONFLICT(application_id) DO UPDATE SET retry_until=EXCLUDED.retry_until`, testApp)
	case "down":
		if err := p.DB.Ping(ctx); err == nil {
			t.Fatal("database must be stopped")
		}
		start := time.Now()
		assertACK(t, h, 503, testOrder, "888", testChildA, "database-down")
		t.Logf("stopped PostgreSQL HTTP 503 after %s", time.Since(start))
	case "after":
		defer cleanStores(t, ctx, p)
		defer mustExec(t, ctx, p, `DELETE FROM app.meli_order_read_coordination WHERE application_id=$1`, testApp)
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
			AND next_offset=3 AND pending_ids=ARRAY['901','902'] AND pending_count=2 AND pending_more
			AND completed_to='2026-11-20T00:00:00Z'::timestamptz
			AND window_to='2026-11-20T01:00:00Z'::timestamptz`, testStoreA)
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_order_read_coordination WHERE application_id=$1 AND retry_until>now()`, testApp)
		assertCount(t, ctx, p, 2, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
		assertACK(t, h, 200, testOrder, "888", testChildA, "crash")
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='888' AND attempts=1 AND generation=1 AND lease_token IS NOT NULL`, testStoreA)
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_cooldowns WHERE store_id=$1 AND available_at > now()`, testStoreB)
		// Advance the lease to avoid a wall-clock wait in CI.
		mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET lease_until=now()-interval '1 second' WHERE store_id=$1 AND resource_id='888'`, testStoreA)
		if _, err := w.claim(ctx); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("shared application cooldown did not block claim after restart: %v", err)
		}
		mustExec(t, ctx, p, `UPDATE app.meli_order_read_coordination SET retry_until=now()-interval '1 second' WHERE application_id=$1`, testApp)
		if err := w.runOnce(ctx); err != nil {
			t.Fatal(err)
		}
		assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='888'`, testStoreA)
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='889' AND attempts=0`, testStoreA)
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='890' AND last_failure='rate_limited'`, testStoreB)
	default:
		t.Fatalf("unknown TEST_RESTART_PHASE %q", phase)
	}
}

// A persisted cooldown protects one store across topics, never another store,
// and releases exactly at its boundary instead of waiting for a new receipt.
func TestNotificationCooldownBoundaryPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	ready := func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { return nil }}, nil
	}
	worker := NewWorker(p, testApp, map[string]Prepare{testOrder: ready, testShipment: ready})
	assertACK(t, h, 200, testOrder, "997", testChildA, "cooldown-a")
	assertACK(t, h, 200, testShipment, "897", testChildA, "cooldown-b")
	j, err := worker.claim(ctx)
	if err != nil || j.Topic != testOrder || j.ResourceID != "997" {
		t.Fatalf("unexpected first claim: %+v %v", j, err)
	}
	if err := worker.retry(ctx, j, &RateLimited{RetryAfter: 90 * time.Second}); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_cooldowns
		WHERE store_id=$1 AND available_at BETWEEN now()+interval '85 seconds' AND now()+interval '91 seconds'`, testStoreA)
	if _, err := worker.claim(ctx); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("cooldown leaked another topic of the same store: %v", err)
	}

	assertACK(t, h, 200, testShipment, "896", testChildB, "cooldown-c")
	other, err := worker.claim(ctx)
	if err != nil || other.ResourceID != "896" {
		t.Fatalf("cooldown blocked another store: %+v %v", other, err)
	}

	// Advance the persisted boundary instead of sleeping in CI.
	mustExec(t, ctx, p, `UPDATE app.meli_notification_cooldowns SET available_at=now()-interval '1 second' WHERE store_id=$1`, testStoreA)
	released, err := worker.claim(ctx)
	if err != nil || released.ResourceID != "897" || released.StoreID != testStoreA {
		t.Fatalf("cooldown did not release at its boundary: %+v %v", released, err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE store_id=$1 AND resource_id='997' AND last_failure='rate_limited' AND dead_at IS NULL`, testStoreA)
}

// A handler that outlives its budget must not commit or lose the job: the lease
// expires and the next worker pass retries the same delivery.
func TestNotificationTimeoutKeepsLeasePostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	assertACK(t, h, 200, testOrder, "995", testChildA, "timeout")
	blocked := NewWorker(p, testApp, map[string]Prepare{testOrder: func(ctx context.Context, _ Job) (VerifiedResource, error) {
		<-ctx.Done()
		return VerifiedResource{}, ctx.Err()
	}})
	short, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	if err := blocked.runOnce(short); err == nil {
		t.Fatal("expired handler budget reported as success")
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
		WHERE store_id=$1 AND resource_id='995' AND attempts=1 AND lease_until > now() AND dead_at IS NULL`, testStoreA)
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_cooldowns WHERE store_id=$1`, testStoreA)

	mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET lease_until=now()-interval '1 second'
		WHERE store_id=$1 AND resource_id='995'`, testStoreA)
	var applied int
	recovered := NewWorker(p, testApp, map[string]Prepare{testOrder: func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { applied++; return nil }}, nil
	}})
	if err := recovered.runOnce(ctx); err != nil || applied != 1 {
		t.Fatalf("expired lease was not retried: applied=%d err=%v", applied, err)
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND resource_id='995'`, testStoreA)
}

func TestNotificationDeferredClaimsAreRefunded(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	assertACK(t, h, 200, testOrder, "993", testChildA, "deferred")
	worker := NewWorker(p, testApp, map[string]Prepare{testOrder: func(context.Context, Job) (VerifiedResource, error) {
		return VerifiedResource{}, &Deferred{Reason: DeferBusy, Delay: 2 * time.Second}
	}})
	for range 10 {
		if err := worker.runOnce(ctx); err != nil {
			t.Fatal(err)
		}
		assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
WHERE store_id=$1 AND resource_id='993' AND attempts=0 AND lease_token IS NULL
AND dead_at IS NULL AND last_failure='local_busy' AND available_at>now()`, testStoreA)
		mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET available_at=now()
WHERE store_id=$1 AND resource_id='993'`, testStoreA)
	}
	mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET pending_since=now()-interval '25 hours'
WHERE store_id=$1 AND resource_id='993'`, testStoreA)
	if err := worker.runOnce(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
WHERE store_id=$1 AND resource_id='993' AND attempts=0 AND dead_at IS NOT NULL`, testStoreA)
}

// A pending association waits 1, 2, 4, ... minutes instead of polling every
// minute, and the cap keeps the projection bounded.
func TestNotificationDeferredBackoffPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	assertACK(t, h, 200, testShipment, "991", testChildA, "backoff")
	ready := func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID}, nil
	}
	worker := NewWorker(p, testApp, map[string]Prepare{testOrder: ready, testShipment: ready})
	for _, tc := range []struct {
		age   string
		delay time.Duration
	}{
		{"0 minutes", time.Minute},
		{"40 minutes", 16 * time.Minute},
		{"3 hours", time.Hour},
	} {
		mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs
SET pending_since=now()-interval '`+tc.age+`', available_at=now()
WHERE store_id=$1 AND resource_id='991'`, testStoreA)
		j, err := worker.claim(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if err := worker.retry(ctx, j, &Deferred{Reason: DeferAssociation, Delay: time.Minute}); err != nil {
			t.Fatal(err)
		}
		var wait float64
		if err := p.DB.QueryRow(ctx, `SELECT extract(epoch from available_at-now())
FROM app.meli_notification_jobs WHERE id=$1`, j.ID).Scan(&wait); err != nil {
			t.Fatal(err)
		}
		if time.Duration(wait*float64(time.Second)) < tc.delay-30*time.Second ||
			time.Duration(wait*float64(time.Second)) > tc.delay+30*time.Second {
			t.Fatalf("age %s waited %.0fs, want %s", tc.age, wait, tc.delay)
		}
	}
}

func TestNotificationPermanentScopeDeadLettersImmediately(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	assertACK(t, h, 200, testOrder, "992", testChildA, "permanent-scope")
	worker := NewWorker(p, testApp, map[string]Prepare{testOrder: func(context.Context, Job) (VerifiedResource, error) {
		return VerifiedResource{}, ErrPermanentScope
	}})
	if err := worker.runOnce(ctx); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
WHERE store_id=$1 AND resource_id='992' AND attempts=1 AND dead_at IS NOT NULL
AND last_failure='scope_rejected' AND lease_token IS NULL`, testStoreA)
}

func TestCustomerTopicsDurableAdmissionPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	assertACK(t, h, 200, "marketplace_questions", "123", testChildA, "question")
	assertACK(t, h, 200, "marketplace_messages", "a35cf79864a845ca9a3bf6aee59bb4d7", testChildA, "message")
	assertACK(t, h, 200, "marketplace_messages", "a35cf79864a845ca9a3bf6aee59bb4d7", testChildA, "message")
	assertACK(t, h, 403, "marketplace_questions", "124", "123456789", "unmapped")
	if err := migrations.Apply(ctx, p.DB); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 2, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1`, testStoreA)
	assertCount(t, ctx, p, 2, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND generation=1`, testStoreA)
	prepare := func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { return nil }}, nil
	}
	w := NewWorker(p, testApp, map[string]Prepare{"marketplace_questions": prepare, "marketplace_messages": prepare})
	for range 2 {
		if err := w.runOnce(ctx); err != nil {
			t.Fatal(err)
		}
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
}

func TestStatusTopicsDurableAdmissionPostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	assertACK(t, h, 200, "marketplace_claims", "123", testChildA, "claim")
	assertACK(t, h, 200, "marketplace_items", "MLB123", testChildA, "item")
	assertACK(t, h, 200, "marketplace_items", "MLB123", testChildA, "item")
	assertACK(t, h, 400, "marketplace_items", "CBT123", testChildA, "parent-item")
	if err := migrations.Apply(ctx, p.DB); err != nil {
		t.Fatal(err)
	}
	assertCount(t, ctx, p, 2, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1 AND generation=1`, testStoreA)
	prepare := func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { return nil }}, nil
	}
	w := NewWorker(p, testApp, map[string]Prepare{"marketplace_claims": prepare, "marketplace_items": prepare})
	for range 2 {
		if err := w.runOnce(ctx); err != nil {
			t.Fatal(err)
		}
	}
	assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1`, testStoreA)
}

func TestNotificationNewGenerationSurvivesOldFailurePostgres(t *testing.T) {
	ctx, p := testDatabase(t)
	seedStores(t, ctx, p)
	defer cleanStores(t, ctx, p)
	h := testReceiver(t, p)
	prepare := func(_ context.Context, j Job) (VerifiedResource, error) {
		return VerifiedResource{Topic: j.Topic, ResourceID: j.ResourceID, SellerID: j.NotificationUserID,
			Apply: func(context.Context, LocalWriter) error { return nil }}, nil
	}
	w := NewWorker(p, testApp, map[string]Prepare{testOrder: prepare})
	for i, tc := range []struct {
		name  string
		cause error
	}{
		{"permanent", ErrPermanentScope},
		{"final_attempt", errors.New("synthetic failure")},
		{"expired_lease", nil},
		{"rate_limited", &RateLimited{RetryAfter: time.Minute}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id := fmt.Sprint(8100 + i)
			assertACK(t, h, 200, testOrder, id, testChildA, "first")
			mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET attempts=7
				WHERE store_id=$1 AND resource_id=$2`, testStoreA, id)
			old, err := w.claim(ctx)
			if err != nil || old.ResourceID != id || old.attempts != 8 {
				t.Fatalf("final claim: %+v %v", old, err)
			}
			assertACK(t, h, 200, testOrder, id, testChildA, "first")
			assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
				WHERE id=$1 AND generation=1 AND attempts=8`, old.ID)
			assertACK(t, h, 200, testOrder, id, testParentA, "new")
			assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
				WHERE id=$1 AND generation=2 AND attempts=0 AND dead_at IS NULL AND lease_token=$2`, old.ID, old.leaseToken)
			if tc.cause == nil {
				mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET lease_until=now()-interval '1 second' WHERE id=$1`, old.ID)
			} else {
				if err := w.retry(ctx, old, tc.cause); err != nil {
					t.Fatal(err)
				}
				assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_jobs
					WHERE id=$1 AND generation=2 AND attempts=0 AND dead_at IS NULL AND lease_token IS NULL`, old.ID)
				if tc.name == "rate_limited" {
					assertCount(t, ctx, p, 1, `SELECT count(*) FROM app.meli_notification_cooldowns
						WHERE store_id=$1 AND available_at>now()`, testStoreA)
					if _, err := w.claim(ctx); !errors.Is(err, pgx.ErrNoRows) {
						t.Fatalf("new generation bypassed real cooldown: %v", err)
					}
					mustExec(t, ctx, p, `UPDATE app.meli_notification_cooldowns SET available_at=now()-interval '1 second' WHERE store_id=$1`, testStoreA)
				}
				mustExec(t, ctx, p, `UPDATE app.meli_notification_jobs SET available_at=now()-interval '1 second' WHERE id=$1`, old.ID)
			}
			fresh, err := w.claim(ctx)
			if err != nil || fresh.ID != old.ID || fresh.generation != 2 || fresh.attempts != 1 || fresh.leaseToken == old.leaseToken {
				t.Fatalf("new generation lost: %+v %v", fresh, err)
			}
			resource, _ := prepare(ctx, fresh)
			if err := w.finish(ctx, fresh, resource); err != nil {
				t.Fatal(err)
			}
			assertCount(t, ctx, p, 0, `SELECT count(*) FROM app.meli_notification_jobs WHERE id=$1`, old.ID)
		})
	}
}
