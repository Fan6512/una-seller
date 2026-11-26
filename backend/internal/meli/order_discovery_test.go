package meli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/notifications"
)

func TestSearchDiscoveryOrdersPostgres(t *testing.T) {
	from := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if r.Method != http.MethodGet || r.URL.Path != "/marketplace/orders/search" ||
			q.Get("seller.id") != "987654" || q.Get("site") != "MLB" ||
			q.Get("date_created.from") != from.Format(time.RFC3339) || q.Get("date_created.to") != to.Format(time.RFC3339) ||
			q.Get("sort") != "date_asc" || q.Get("limit") != "3" || q.Get("order.status") != "" {
			t.Errorf("invalid discovery search path: %s", r.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if q.Get("offset") == "3" {
			fmt.Fprint(w, `{"results":[],"paging":{"total":2}}`)
			return
		}
		fmt.Fprint(w, `{"results":[{"id":700,"orders":[{"id":10001,"seller":{"id":987654}},{"id":10002,"seller":{"id":987654}}]}],"paging":{"total":1}}`)
	})
	scope := notifications.DiscoveryScope{StoreID: c.store, OwnerID: c.owner, ParentID: c.parent,
		ChildID: "987654", Site: "MLB", Logistic: "remote"}
	page, err := c.service.SearchDiscoveryOrders(t.Context(), scope, from, to, 0)
	if err != nil || page.Count != 1 || page.Total != 1 || len(page.IDs) != 2 || page.IDs[0] != "10001" || page.IDs[1] != "10002" {
		t.Fatalf("wrong pack order IDs: %+v %v", page, err)
	}
	if _, err := c.service.SearchDiscoveryOrders(t.Context(), scope, from, to, 3); !errors.Is(err, notifications.ErrDiscoveryPageChanged) {
		t.Fatalf("shrunken page not classified: %v", err)
	}
	scope.ChildID = "999999"
	if _, err := c.service.SearchDiscoveryOrders(t.Context(), scope, from, to, 0); !errors.Is(err, notifications.ErrScope) {
		t.Fatalf("wrong child accepted: %v", err)
	}
	if len(c.calls()) != 2 {
		t.Fatal("invalid scope reached provider")
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_order_syncs WHERE store_id=$1`, c.store); got != 0 {
		t.Fatal("discovery modified manual cursor")
	}
}

func TestDiscoveryToWorkerRoundtripPostgres(t *testing.T) {
	created := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
	var reads atomic.Int32
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/marketplace/orders/search":
			q := r.URL.Query()
			from, fromErr := time.Parse(time.RFC3339, q.Get("date_created.from"))
			to, toErr := time.Parse(time.RFC3339, q.Get("date_created.to"))
			if fromErr != nil || toErr != nil || created.Before(from) || created.After(to) ||
				q.Get("seller.id") != "987654" || q.Get("site") != "MLB" ||
				q.Get("sort") != "date_asc" || q.Get("offset") != "0" || q.Get("limit") != "3" ||
				q.Get("order.status") != "" {
				t.Error("unexpected discovery search contract")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fmt.Fprint(w, `{"results":[{"id":700,"orders":[{"id":8000011,"seller":{"id":987654}}]}],"paging":{"total":1}}`)
		case "/marketplace/orders/8000011":
			if reads.Add(1) == 1 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			payload := orderPayload("8000011", "987654", "MLB", "700", "801", "paid", created.Format(time.RFC3339))
			fmt.Fprint(w, strings.Replace(payload, "2026-11-20T00:00:00Z", created.Format(time.RFC3339), 1))
		default:
			http.NotFound(w, r)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if got := c.count(t, `SELECT count(*) FROM app.meli_store_connections WHERE id<>$1::uuid`, c.store); got != 0 {
		t.Fatal("discovery requires a fresh isolated TEST_DATABASE_URL and serial package tests (-p 1); foreign store fixtures were preserved")
	}
	// A unique synthetic application avoids sharing cooldown state with other
	// tests. Enable the same coordination path used by the production worker.
	appID := c.parent
	c.service.cfg.ClientID = appID
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if _, err := c.db.Exec(cleanup, `DELETE FROM app.meli_order_read_coordination WHERE application_id=$1`, appID); err != nil {
			t.Error(err)
		}
	})
	if err := c.service.EnableSharedOrderCoordination(ctx); err != nil {
		t.Fatal(err)
	}
	inbox := notifications.PostgresInbox{DB: c.db}
	if err := inbox.DiscoverySweep(ctx, appID, c.service.SearchDiscoveryOrders); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1
AND topic='marketplace_orders' AND resource_id='8000011' AND source='compensation'`, c.store); got != 1 {
		t.Fatal("discovered order not durably admitted")
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1`, c.store); got != 0 {
		t.Fatal("search wrote an order directly")
	}
	var watermark time.Time
	if err := c.db.QueryRow(ctx, `SELECT completed_to FROM app.meli_order_discovery
WHERE store_id=$1 AND window_to IS NULL AND last_completed_at IS NOT NULL`, c.store).Scan(&watermark); err != nil {
		t.Fatal(err)
	}
	worker := notifications.NewWorker(inbox, appID, c.service.NotificationHandlers())
	if err := worker.ValidateActivation(); err != nil {
		t.Fatal(err)
	}
	runCtx, stop := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		worker.Run(runCtx, log.New(io.Discard, "", 0))
	}()
	defer func() {
		stop()
		<-done
	}()
	waitFor := func(query string) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for c.count(t, query, c.store) != 1 {
			if time.Now().After(deadline) {
				t.Fatalf("worker state not reached: %s", query)
			}
			select {
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			case <-time.After(100 * time.Millisecond):
			}
		}
	}
	waitFor(`SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1
AND resource_id='8000011' AND attempts=1 AND dead_at IS NULL AND lease_token IS NULL
AND last_failure='handler_failed'`)
	if got := c.count(t, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1`, c.store); got != 0 {
		t.Fatal("failed read unexpectedly persisted an order")
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1
AND completed_to=$2 AND window_to IS NULL`, c.store, watermark); got != 1 {
		t.Fatal("worker failure changed the admitted-search watermark")
	}
	// Avoid a wall-clock backoff wait while exercising the real retry path.
	if _, err := c.db.Exec(ctx, `UPDATE app.meli_notification_jobs SET available_at=now()
WHERE store_id=$1 AND resource_id='8000011' AND lease_token IS NULL AND dead_at IS NULL`, c.store); err != nil {
		t.Fatal(err)
	}
	waitFor(`SELECT count(*) FROM app.meli_orders WHERE store_id=$1 AND order_id='8000011' AND status='paid'`)
	stop()
	<-done
	if got := c.count(t, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1
AND topic='marketplace_orders' AND resource_id='8000011'`, c.store); got != 0 {
		t.Fatal("discovered job not completed")
	}
	if reads.Load() != 2 || len(c.calls()) != 3 {
		t.Fatal("expected one real search and two detail reads")
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_order_syncs WHERE store_id=$1`, c.store); got != 0 {
		t.Fatal("discovery modified manual cursor")
	}
}
