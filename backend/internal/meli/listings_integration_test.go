package meli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestListingWritePoolAcquireTimeout(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	acquireCtx, cancelAcquire := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelAcquire()
	for i := int32(0); i < db.Config().MaxConns; i++ {
		conn, err := db.Acquire(acquireCtx)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Release()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	var logs bytes.Buffer
	trace := slog.New(slog.NewTextHandler(&logs, nil))
	err := s.saveListingBatch(ctx, owner, store, ListingScope{Site: "CBT"}, "", tokenSnapshot{}, listingCursor{}, ListingSync{}, nil, trace)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected connection acquisition timeout, got %v", err)
	}
	if line := logs.String(); !strings.Contains(line, "stage=db_acquire") || !strings.Contains(line, "outcome=timeout") ||
		!strings.Contains(line, "new_conns_delta=0") || !strings.Contains(line, "empty_acquires_delta=0") ||
		!strings.Contains(line, "empty_acquire_wait_ms_delta=0") || !strings.Contains(line, "canceled_acquires_delta=1") ||
		strings.Contains(line, "stage=db_begin") || strings.Contains(line, owner) || strings.Contains(line, store) {
		t.Fatalf("unexpected timing log: %s", line)
	}
}

func TestPublishedListingsBatchAtomicity(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections
		SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	scope := ListingScope{Site: "CBT"}
	row, seller, err := s.listingScope(ctx, owner, store, scope)
	if err != nil {
		t.Fatal(err)
	}
	trace := slog.New(slog.DiscardHandler)
	details := make(map[string]listingWrite, ListingBatchSize)
	for i := 1; i <= ListingBatchSize; i++ {
		id := fmt.Sprintf("CBT%d", i)
		details[id] = listingWrite{detail: json.RawMessage(fmt.Sprintf(`{"id":%q,"seller_id":%q}`, id, seller))}
	}
	next := ListingBatchSize
	result := ListingSync{Saved: len(details), NextOffset: &next}
	if err := s.saveListingBatch(ctx, owner, store, scope, seller, row, listingCursor{}, result, details, trace); err != nil {
		t.Fatal(err)
	}
	assertState := func(wantCount, wantOffset int) {
		t.Helper()
		if inUse := db.Stat().AcquiredConns(); inUse != 0 {
			t.Fatalf("connection still checked out: %d", inUse)
		}
		page, err := s.ListPublished(ctx, owner, store, scope, 0)
		if err != nil || len(page.Items) != wantCount || page.NextOffset == nil || *page.NextOffset != wantOffset {
			t.Fatalf("read-back: items=%d next=%v error=%v", len(page.Items), page.NextOffset, err)
		}
	}
	assertState(ListingBatchSize, next)

	following := map[string]listingWrite{"CBT6": {detail: json.RawMessage(fmt.Sprintf(`{"id":"CBT6","seller_id":%q}`, seller))}}
	followResult := ListingSync{Saved: 1, NextOffset: new(next + 1)}
	if err := s.saveListingBatch(ctx, owner, store, scope, seller, row, listingCursor{}, followResult, following, trace); !errors.Is(err, ErrListingCursor) {
		t.Fatalf("expected stale cursor rejection, got %v", err)
	}
	assertState(ListingBatchSize, next)

	cursor := listingCursor{Offset: next}
	following["CBT7"] = listingWrite{} // SQL NULL violates details NOT NULL inside the Batch.
	followResult = ListingSync{Saved: 2, NextOffset: new(next + 2)}
	err = s.saveListingBatch(ctx, owner, store, scope, seller, row, cursor, followResult, following, trace)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23502" || pgErr.ColumnName != "details" {
		t.Fatalf("expected Batch NOT NULL failure, got %v", err)
	}
	assertState(ListingBatchSize, next)

	delete(following, "CBT7")
	followResult = ListingSync{Saved: 1, NextOffset: new(1000)}
	err = s.saveListingBatch(ctx, owner, store, scope, seller, row, cursor, followResult, following, trace)
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("expected cursor CHECK failure, got %v", err)
	}
	assertState(ListingBatchSize, next)

	followResult.NextOffset = new(next + 1)
	if err := s.saveListingBatch(ctx, owner, store, scope, seller, row, cursor, followResult, following, trace); err != nil {
		t.Fatal(err)
	}
	assertState(next+1, next+1)
}

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
		sync, err := s.SyncPublished(ctx, owner, store, global, nil, i != 0)
		if err != nil || sync.Saved != 1 || sync.Partial != 1 || sync.Skipped != 0 {
			t.Fatalf("global sync: %+v %v", sync, err)
		}
	}
	sync, err := s.SyncPublished(ctx, owner, store, local, nil, false)
	if err != nil || sync.Saved != 1 || sync.Interruption == nil || sync.Interruption.PlatformStatus != 429 || sync.NextOffset == nil || *sync.NextOffset != 1 {
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
	if _, err := s.SyncPublished(ctx, owner, store, ListingScope{Site: "MLB", Logistic: "fulfillment"}, nil, false); !errors.Is(err, ErrSiteNotAuthorized) {
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
			var timing listingGETTiming
			status, err := s.listingGET(context.Background(), "synthetic", "/detail", &item, &timing)
			if status != code || !errors.As(err, &remote) || int(remote) != code || timing.failedPhase != "http_status" {
				t.Fatalf("status=%d error=%v", status, err)
			}
		})
	}
}

func TestPublishedListingGETTimeoutPhases(t *testing.T) {
	for _, phase := range []string{"response_headers", "response_body"} {
		t.Run(phase, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if phase == "response_body" {
					fmt.Fprint(w, "{")
					w.(http.Flusher).Flush()
				}
				<-r.Context().Done()
			}))
			defer server.Close()
			s := Service{apiURL: server.URL, client: server.Client()}
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			var logs bytes.Buffer
			trace := slog.New(slog.NewJSONHandler(&logs, nil))
			_, _, err := s.fetchListingBatch(ctx, "synthetic_secret", "123456", ListingScope{Site: "CBT"}, 0, "", trace)
			if !errors.Is(err, context.DeadlineExceeded) || listingInterruption(err).Code != "listing_timeout" {
				t.Fatalf("search returned %v", err)
			}
			var entry struct {
				Stage           string `json:"stage"`
				FailedPhase     string `json:"failed_phase"`
				Outcome         string `json:"outcome"`
				BudgetExhausted bool   `json:"platform_budget_exhausted"`
			}
			if err := json.Unmarshal(logs.Bytes(), &entry); err != nil ||
				entry.Stage != "platform_search_get" || entry.FailedPhase != phase || entry.Outcome != "timeout" || !entry.BudgetExhausted {
				t.Fatalf("unexpected timeout trace %+v, decode error %v", entry, err)
			}
			for _, sensitive := range []string{"synthetic_secret", "123456", "/marketplace/users/"} {
				if strings.Contains(logs.String(), sensitive) {
					t.Fatal("timeout trace contains request data")
				}
			}
		})
	}
}

func TestPublishedListingsResume(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx := context.Background()
	var seller string
	if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
        SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&seller); err != nil {
		t.Fatal(err)
	}
	var phase atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("expected authenticated platform GET")
			w.WriteHeader(400)
			return
		}
		switch r.URL.Path {
		case "/marketplace/users/" + seller + "/items/search":
			offset := r.URL.Query().Get("offset")
			switch offset {
			case "0":
				fmt.Fprintf(w, `{"seller_id":%q,"results":["CBT1","invalid","CBT2","CBT3"],"paging":{"total":4,"offset":0,"limit":5}}`, seller)
			case "3":
				fmt.Fprintf(w, `{"seller_id":%q,"results":["CBT3"],"paging":{"total":4,"offset":3,"limit":5}}`, seller)
			default:
				t.Error("unexpected search offset ", offset)
				w.WriteHeader(400)
			}
		case "/marketplace/items/CBT1":
			fmt.Fprintf(w, `{"id":"CBT1","site_id":"CBT","seller_id":%q,"price":0}`, seller)
		case "/marketplace/items/CBT2":
			w.WriteHeader(404)
		case "/marketplace/items/CBT3":
			switch phase.Load() {
			case 0:
				w.WriteHeader(429)
			case 1:
				w.WriteHeader(500)
			case 2:
				fmt.Fprintf(w, `{"id":"CBT3","site_id":"MLB","seller_id":%q}`, seller)
			default:
				fmt.Fprintf(w, `{"id":"CBT3","site_id":"CBT","seller_id":%q}`, seller)
			}
		default:
			t.Error("unexpected path")
			w.WriteHeader(400)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := ListingScope{Site: "CBT"}
	result, err := s.SyncPublished(ctx, owner, store, scope, nil, false)
	if err != nil || result.Saved != 1 || result.Skipped != 2 || result.Processed != 3 ||
		result.Interruption == nil || result.Interruption.Code != "listing_rate_limited" || result.NextOffset == nil || *result.NextOffset != 3 {
		t.Fatalf("partial batch=%+v err=%v", result, err)
	}
	page, err := s.ListPublished(ctx, owner, store, scope, 0)
	if err != nil || page.NextOffset == nil || *page.NextOffset != 3 || len(page.Items) != 1 {
		t.Fatalf("persisted cursor=%+v err=%v", page, err)
	}
	start := 0
	if _, err := s.SyncPublished(ctx, owner, store, scope, &start, false); !errors.Is(err, ErrListingCursor) {
		t.Fatalf("stale offset: %v", err)
	}
	for i, code := range []string{"listing_platform_error", "listing_identity_mismatch"} {
		phase.Store(int32(i + 1))
		result, err = s.SyncPublished(ctx, owner, store, scope, nil, false)
		var pause *ListingInterruption
		if !errors.As(err, &pause) || pause.Code != code || pause.Retryable != (i == 0) ||
			result.NeedsReview != i || result.NextOffset == nil || *result.NextOffset != 3 {
			t.Fatalf("interruption=%+v err=%v", result, err)
		}
	}
	phase.Store(3)
	result, err = s.SyncPublished(ctx, owner, store, scope, nil, false)
	if err != nil || result.Saved != 1 || result.NextOffset != nil {
		t.Fatalf("resume=%+v err=%v", result, err)
	}
	if _, err := s.SyncPublished(ctx, owner, store, scope, nil, false); !errors.Is(err, ErrListingCursor) {
		t.Fatalf("completed cursor accepted: %v", err)
	}
}

func TestPublishedListingsTimeoutAndCancellation(t *testing.T) {
	for _, scenario := range []string{"timeout", "cancel", "complete"} {
		t.Run(scenario, func(t *testing.T) {
			s, db, owner, store := setupRefreshTest(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var seller string
			if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
                SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&seller); err != nil {
				t.Fatal(err)
			}
			var thirdRequested atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/marketplace/users/" + seller + "/items/search":
					fmt.Fprintf(w, `{"seller_id":%q,"results":["CBT1","CBT2","CBT3","CBT4","CBT5"],"paging":{"total":5,"offset":0,"limit":5}}`, seller)
				case "/marketplace/items/CBT1", "/marketplace/items/CBT2", "/marketplace/items/CBT4", "/marketplace/items/CBT5":
					fmt.Fprintf(w, `{"id":%q,"site_id":"CBT","seller_id":%q}`, strings.TrimPrefix(r.URL.Path, "/marketplace/items/"), seller)
				case "/marketplace/items/CBT3":
					thirdRequested.Store(true)
					if scenario == "complete" {
						fmt.Fprintf(w, `{"id":"CBT3","site_id":"CBT","seller_id":%q}`, seller)
						return
					}
					if scenario == "cancel" {
						cancel()
					}
					<-r.Context().Done()
				default:
					w.WriteHeader(400)
				}
			}))
			defer server.Close()
			s.apiURL = server.URL
			scope := ListingScope{Site: "CBT"}
			result, err := s.syncPublished(ctx, owner, store, scope, nil, false, "", 2500*time.Millisecond)
			if !thirdRequested.Load() {
				t.Fatal("third detail was not requested")
			}
			want := 2
			switch scenario {
			case "cancel":
				want = 0
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("cancellation: %v", err)
				}
			case "complete":
				want = 5
				if err != nil || result.Saved != want || result.Processed != want || result.NextOffset != nil || result.Interruption != nil {
					t.Fatalf("complete=%+v err=%v", result, err)
				}
			case "timeout":
				if err != nil || result.Saved != want || result.Processed != want || result.NextOffset == nil ||
					*result.NextOffset != want || result.Interruption == nil || result.Interruption.Code != "listing_timeout" {
					t.Fatalf("timeout=%+v err=%v", result, err)
				}
			}
			page, err := s.ListPublished(context.Background(), owner, store, scope, 0)
			if err != nil || len(page.Items) != want {
				t.Fatalf("read-back: %+v %v", page, err)
			}
			if scenario == "complete" && page.NextOffset != nil ||
				scenario != "complete" && (page.NextOffset == nil || *page.NextOffset != want) {
				t.Fatalf("unexpected persisted cursor: %+v", page.NextOffset)
			}
		})
	}
}

func TestPublishedListingsAuthorizationChangesDuringFetch(t *testing.T) {
	for _, kind := range []string{"rotation", "disconnect", "reauthorization", "site mapping"} {
		t.Run(kind, func(t *testing.T) {
			s, db, owner, store := setupRefreshTest(t)
			ctx := context.Background()
			var seller string
			if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
                SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&seller); err != nil {
				t.Fatal(err)
			}
			rotated, err := s.cipher.Seal("synthetic_rotated_access", tokenAAD(owner, seller, "access"))
			if err != nil {
				t.Fatal(err)
			}
			entered := make(chan struct{})
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/marketplace/users/"+seller+"/items/search" {
					fmt.Fprintf(w, `{"seller_id":%q,"results":["CBT1"],"paging":{"total":1,"offset":0,"limit":5}}`, seller)
					return
				}
				close(entered)
				select {
				case <-r.Context().Done():
					return
				case <-release:
					fmt.Fprintf(w, `{"id":"CBT1","site_id":"CBT","seller_id":%q}`, seller)
				}
			}))
			defer server.Close()
			defer close(release)
			s.apiURL = server.URL
			scope := ListingScope{Site: "CBT"}
			done := make(chan error, 1)
			go func() {
				_, err := s.SyncPublished(ctx, owner, store, scope, nil, false)
				done <- err
			}()
			select {
			case <-entered:
			case err := <-done:
				t.Fatal("sync finished before detail request", err)
			case <-time.After(5 * time.Second):
				t.Fatal("detail request not reached")
			}
			if _, err := s.SyncPublished(ctx, owner, store, scope, nil, false); !errors.Is(err, ErrListingBusy) {
				t.Fatalf("same store lock: %v", err)
			}
			switch kind {
			case "disconnect", "reauthorization":
				err = s.Disconnect(ctx, owner, store)
				if err != nil {
					t.Fatal(err)
				}
				if kind == "reauthorization" {
					_, err = db.Exec(ctx, `UPDATE app.meli_store_connections SET status='connected',
                      access_token_ciphertext=$2, refresh_token_ciphertext=$2,
                      access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid`, store, rotated)
				}
			case "rotation":
				_, err = db.Exec(ctx, `UPDATE app.meli_store_connections
                  SET access_token_ciphertext=$2 WHERE id=$1::uuid`, store, rotated)
			case "site mapping":
				_, err = db.Exec(ctx, `UPDATE app.meli_store_connections
                  SET marketplaces='[]'::jsonb WHERE id=$1::uuid`, store)
			}
			if err != nil {
				t.Fatal(err)
			}
			release <- struct{}{}
			if err := <-done; !errors.Is(err, ErrListingAuthorizationChanged) {
				t.Fatalf("changed snapshot: %v", err)
			}
			var rows int
			if err := db.QueryRow(ctx, `SELECT
                (SELECT count(*) FROM app.meli_published_listings WHERE store_id=$1::uuid) +
                (SELECT count(*) FROM app.meli_listing_syncs WHERE store_id=$1::uuid)`, store).Scan(&rows); err != nil || rows != 0 {
				t.Fatalf("changed auth wrote %d rows, %v", rows, err)
			}
		})
	}
}

func TestPublishedListingsPoolAndPagination(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx := context.Background()
	var parent, second string
	if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
        SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&parent); err != nil {
		t.Fatal(err)
	}
	seller := parent + "1"
	access, err := s.cipher.Seal("synthetic_second_access", tokenAAD(owner, seller, "access"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
      (owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at)
      VALUES ($1::uuid,$2,$3,$3,now()+interval '30 minutes') RETURNING id::text`, owner, seller, access).Scan(&second); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.Exec(ctx, "DELETE FROM app.meli_store_connections WHERE id=$1::uuid", second) })
	cfg := db.Config()
	cfg.MaxConns = 1
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	s.db = pool
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		segments := strings.Split(r.URL.Path, "/")
		entered <- struct{}{}
		select {
		case <-release:
			fmt.Fprintf(w, `{"seller_id":%q,"results":[],"paging":{"total":0,"offset":0,"limit":5}}`, segments[3])
		case <-r.Context().Done():
		}
	}))
	defer server.Close()
	defer close(release)
	s.apiURL = server.URL
	scope := ListingScope{Site: "CBT"}
	done := make(chan error, 2)
	for _, storeID := range []string{store, second} {
		go func() {
			_, err := s.SyncPublished(ctx, owner, storeID, scope, nil, false)
			done <- err
		}()
	}
	for range 2 {
		select {
		case <-entered:
		case err := <-done:
			t.Fatal("sync returned before release", err)
		case <-time.After(4 * time.Second):
			t.Fatal("second store fetch could not start with one DB connection")
		}
	}
	queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = pool.Exec(queryCtx, `INSERT INTO app.meli_published_listings (store_id,site_id,logistic_type,listing_id,details)
        SELECT $1::uuid,'CBT','','CBT'||n,jsonb_build_object('id','CBT'||n,'seller_id',$2::text)
        FROM generate_series(101,121) n`, store, parent)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		offset, count int
		more          bool
	}{{0, 20, true}, {20, 1, false}} {
		page, err := s.ListPublished(queryCtx, owner, store, scope, check.offset)
		if err != nil || len(page.Items) != check.count || page.HasMore != check.more {
			t.Fatalf("paging: %+v %v", page, err)
		}
	}
	release <- struct{}{}
	release <- struct{}{}
	for range 2 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
}

func TestPublishedListingsCompleteAndPartialSnapshots(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx := context.Background()
	var seller string
	if err := db.QueryRow(ctx, `UPDATE app.meli_store_connections
      SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid RETURNING seller_id`, store).Scan(&seller); err != nil {
		t.Fatal(err)
	}
	var phase atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/items/search") {
			fmt.Fprintf(w, `{"seller_id":%q,"results":["CBT1"],"paging":{"total":1,"offset":0,"limit":5}}`, seller)
			return
		}
		var fields string
		switch phase.Load() {
		case 0:
			fields = `,"price":15,"currency_id":"USD","title":"Synthetic item","available_quantity":9,"seller_custom_field":"item-sku",
                "pictures":[{"id":"p","secure_url":"https://example.invalid/p"}],"attributes":[{"id":"COLOR","value_name":"blue"}],
                "variations":[{"id":9007199254740993,"available_quantity":5,"attributes":[{"id":"SELLER_SKU","value_name":"variant-sku"}]}]`
		case 1:
			w.WriteHeader(206)
			fields = `,"price":null,"available_quantity":0,"seller_custom_field":null,"pictures":null,"attributes":null,
                "variations":[{"id":9007199254740993,"available_quantity":0,"attributes":[{"id":"SELLER_SKU","value_name":null}]}]`
		case 2:
			w.WriteHeader(206)
			fields = `,"pictures":[],"variations":null`
		default:
			fields = `,"price":null`
		}
		fmt.Fprintf(w, `{"id":"CBT1","site_id":"CBT","seller_id":%q%s}`, seller, fields)
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := ListingScope{Site: "CBT"}
	for i := range 4 {
		phase.Store(int32(i))
		result, err := s.SyncPublished(ctx, owner, store, scope, nil, i != 0)
		if err != nil || result.Saved != 1 || result.NextOffset != nil {
			t.Fatalf("phase %d: %+v %v", i, result, err)
		}
		page, err := s.ListPublished(ctx, owner, store, scope, 0)
		if err != nil || len(page.Items) != 1 {
			t.Fatalf("listing page: %+v %v", page, err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(page.Items[0], &fields); err != nil {
			t.Fatal(err)
		}
		switch i {
		case 1:
			if string(fields["price"]) != "15" || string(fields["available_quantity"]) != "0" ||
				string(fields["title"]) != "\"Synthetic item\"" || string(fields["seller_custom_field"]) != "\"item-sku\"" ||
				!strings.Contains(string(fields["pictures"]), "example.invalid") ||
				!strings.Contains(string(fields["attributes"]), "blue") ||
				!strings.Contains(string(fields["variations"]), "variant-sku") ||
				!strings.Contains(string(fields["variations"]), "9007199254740993") ||
				!strings.Contains(string(fields["variations"]), "\"available_quantity\": 0") {
				t.Fatalf("partial lost original fields or zero: %s", page.Items[0])
			}
		case 2:
			if string(fields["pictures"]) != "[]" || !strings.Contains(string(fields["variations"]), "variant-sku") {
				t.Fatalf("explicit empty and null: %s", page.Items[0])
			}
		case 3:
			if string(fields["price"]) != "null" || fields["title"] != nil || fields["variations"] != nil {
				t.Fatalf("complete snapshot: %s", page.Items[0])
			}
		}
	}
}

func TestPublishedDescriptionOnDemandIsScopedAndReadOnly(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx := context.Background()
	_, err := db.Exec(ctx, `UPDATE app.meli_store_connections
		SET access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid`, store)
	if err != nil {
		t.Fatal(err)
	}
	for _, listing := range []struct {
		id, body string
	}{
		{"MLB1", `{"id":"MLB1","seller_id":"987654","cbt_item_id":"CBT1"}`},
		{"MLB2", `{"id":"MLB2","seller_id":"987654"}`},
		{"MLB3", `{"id":"MLB3","seller_id":"987654","cbt_item_id":"CBT3"}`},
	} {
		_, err := db.Exec(ctx, `INSERT INTO app.meli_published_listings
			(store_id,site_id,logistic_type,listing_id,details)
			VALUES ($1::uuid,'MLB','remote',$2,$3::jsonb)`,
			store, listing.id, listing.body)
		if err != nil {
			t.Fatal(err)
		}
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("unexpected platform request method or credentials")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		switch r.URL.Path {
		case "/marketplace/items/CBT1/description":
			if r.URL.Query().Get("site_id") != "MLB" || r.URL.Query().Get("logistic_type") != "remote" {
				t.Error("wrong description site scope")
			}
			fmt.Fprint(w, `{"plain_text":"Localized product description"}`)
		case "/items/MLB2/description":
			fmt.Fprint(w, `{"plain_text":"Fallback local item description"}`)
		case "/marketplace/items/CBT3/description":
			w.WriteHeader(http.StatusNotFound)
		default:
			t.Errorf("unexpected platform description path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	scope := ListingScope{Site: "MLB", Logistic: "remote"}
	for _, check := range []struct {
		id, want string
	}{
		{"MLB1", "Localized product description"},
		{"MLB2", "Fallback local item description"},
	} {
		got, err := s.GetPublishedDescription(ctx, owner, store, scope, check.id)
		if err != nil || got != check.want {
			t.Fatalf("read %s: got %q err %v", check.id, got, err)
		}
	}
	before := calls.Load()
	if _, err := s.GetPublishedDescription(ctx, "00000000-0000-4000-8000-000000000001", store, scope, "MLB1"); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("wrong owner: %v", err)
	}
	if _, err := s.GetPublishedDescription(ctx, owner, store, ListingScope{Site: "MLB", Logistic: "fulfillment"}, "MLB1"); !errors.Is(err, ErrSiteNotAuthorized) {
		t.Fatalf("wrong site logistics: %v", err)
	}
	if _, err := s.GetPublishedDescription(ctx, owner, store, scope, "MLB999"); !errors.Is(err, ErrListingDescriptionMissing) {
		t.Fatalf("unowned or missing listing: %v", err)
	}
	if _, err := s.GetPublishedDescription(ctx, owner, store, scope, "../foo"); !errors.Is(err, ErrListingResponse) {
		t.Fatalf("invalid listing id: %v", err)
	}
	if calls.Load() != before {
		t.Fatal("unauthorized or invalid reads reached the platform")
	}
	var platform ListingPlatformError
	if _, err := s.GetPublishedDescription(ctx, owner, store, scope, "MLB3"); !errors.As(err, &platform) || int(platform) != http.StatusNotFound {
		t.Fatalf("platform 404 not surfaced: %v", err)
	}
}
