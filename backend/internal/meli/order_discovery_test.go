package meli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	c := newNotificationCase(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/marketplace/orders/8000011" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, orderPayload("8000011", "987654", "MLB", "700", "801", "paid", "2026-11-25T01:00:00Z"))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Shared TEST_DATABASE_URL hygiene: foreign stores would win the oldest-due
	// discovery reservation and resolveScope uniqueness.
	if _, err := c.db.Exec(ctx, `DELETE FROM app.meli_store_connections WHERE id<>$1::uuid`, c.store); err != nil {
		t.Fatal(err)
	}
	inbox := notifications.PostgresInbox{DB: c.db}
	search := func(_ context.Context, scope notifications.DiscoveryScope, from, to time.Time, offset int) (notifications.DiscoveryPage, error) {
		if scope.ChildID != "987654" || offset != 0 || !from.Before(to) {
			t.Errorf("unexpected discovery search scope %+v offset=%d", scope, offset)
			return notifications.DiscoveryPage{}, notifications.ErrScope
		}
		return notifications.DiscoveryPage{IDs: []string{"8000011"}, Count: 1, Total: 1}, nil
	}
	if err := inbox.DiscoverySweep(ctx, "2069392825111111", search); err != nil {
		t.Fatal(err)
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_notification_inbox WHERE store_id=$1
AND topic='marketplace_orders' AND resource_id='8000011' AND source='compensation'`, c.store); got != 1 {
		t.Fatal("discovered order not durably admitted")
	}
	worker := notifications.NewWorker(inbox, "2069392825111111", c.service.NotificationHandlers())
	if err := worker.ValidateActivation(); err != nil {
		t.Fatal(err)
	}
	runCtx, stop := context.WithCancel(context.Background())
	defer stop()
	go worker.Run(runCtx, log.New(io.Discard, "", 0))
	deadline := time.Now().Add(15 * time.Second)
	for {
		var stored int
		if err := c.db.QueryRow(ctx, `SELECT count(*) FROM app.meli_orders WHERE store_id=$1::uuid
AND order_id='8000011' AND status='paid'`, c.store).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		if stored == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("worker never read back the discovered order")
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(300 * time.Millisecond):
		}
	}
	stop()
	if got := c.count(t, `SELECT count(*) FROM app.meli_notification_jobs WHERE store_id=$1
AND topic='marketplace_orders' AND resource_id='8000011'`, c.store); got != 0 {
		t.Fatal("discovered job not completed")
	}
	if got := c.count(t, `SELECT count(*) FROM app.meli_order_discovery WHERE store_id=$1::uuid
AND completed_to IS NOT NULL AND window_to IS NULL`, c.store); got != 1 {
		t.Fatal("discovery watermark not advanced after completion")
	}
}
