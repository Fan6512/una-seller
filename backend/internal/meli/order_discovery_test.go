package meli

import (
	"errors"
	"fmt"
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
