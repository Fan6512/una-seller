package meli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublishedListingsSync(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx := context.Background()
	var parent string
	if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
        SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&parent); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("expected authenticated platform GET")
			w.WriteHeader(400)
			return
		}
		switch r.URL.Path {
		case "/marketplace/users/" + parent + "/items/search":
			fmt.Fprintf(w, `{"seller_id":%q,"results":["CBT1"],"paging":{"total":1,"offset":0,"limit":5}}`, parent)
		case "/marketplace/users/987654/items/search":
			fmt.Fprint(w, `{"seller_id":987654,"results":["MLB1","MLB2"],"paging":{"total":2,"offset":0,"limit":5}}`)
		case "/marketplace/items/CBT1":
			w.WriteHeader(206)
			fmt.Fprintf(w, `{"id":"CBT1","site_id":"CBT","seller_id":%s,"title":"Synthetic global item","price":0,"currency_id":"USD",
                "variations":[{"id":101,"attributes":[{"id":"SELLER_SKU","value_name":"synthetic"}]}]}`, parent)
		case "/marketplace/items/MLB1":
			fmt.Fprint(w, `{"id":"MLB1","site_id":"MLB","seller_id":987654,"user_logistic_type":"remote","cbt_item_id":"CBT1","available_quantity":0}`)
		case "/marketplace/items/MLB2":
			w.WriteHeader(429)
		default:
			t.Errorf("unexpected route %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL

	global := ListingScope{Site: "CBT"}
	local := ListingScope{Site: "MLB", Logistic: "remote"}
	for i := 0; i < 2; i++ {
		sync, err := s.SyncPublished(ctx, owner, store, global, 0)
		if err != nil || sync.Saved != 1 || sync.Partial != 1 || sync.Failed != 0 {
			t.Fatalf("global sync: %+v %v", sync, err)
		}
	}
	sync, err := s.SyncPublished(ctx, owner, store, local, 0)
	if err != nil || sync.Saved != 1 || sync.Failed != 1 || sync.PlatformStatus != 429 || sync.NextOffset == nil || *sync.NextOffset != 0 {
		t.Fatalf("site sync: %+v %v", sync, err)
	}
	for _, check := range []struct {
		scope             ListingScope
		id, parent, field string
	}{
		{global, "CBT1", "", "price"},
		{local, "MLB1", "CBT1", "available_quantity"},
	} {
		page, err := s.ListPublished(ctx, owner, store, check.scope, 0)
		if err != nil || len(page.Items) != 1 || page.SyncedAt == nil || page.HasMore {
			t.Fatalf("read %s: %+v %v", check.scope.Site, page, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(page.Items[0], &fields); err != nil ||
			string(fields["id"]) != fmt.Sprintf("%q", check.id) || string(fields[check.field]) != "0" ||
			(check.parent != "" && string(fields["cbt_item_id"]) != fmt.Sprintf("%q", check.parent)) {
			t.Fatalf("unexpected persisted fields: %v", err)
		}
	}
	if _, err := s.ListPublished(ctx, "00000000-0000-4000-8000-000000000001", store, global, 0); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("wrong owner: %v", err)
	}
	if _, err := s.SyncPublished(ctx, owner, store, ListingScope{Site: "MLB", Logistic: "fulfillment"}, 0); !errors.Is(err, ErrSiteNotAuthorized) {
		t.Fatalf("unverified mode: %v", err)
	}
}

func TestPublishedListingGETPlatformErrors(t *testing.T) {
	for _, code := range []int{401, 403, 429, 500} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(code)
			}))
			defer server.Close()
			s := Service{apiURL: server.URL, client: server.Client()}
			var item listingDetail
			var remote ListingPlatformError
			status, err := s.listingGET(context.Background(), "synthetic", "/detail", &item)
			if status != code || !errors.As(err, &remote) || int(remote) != code {
				t.Fatalf("status=%d error=%v", status, err)
			}
		})
	}
}
