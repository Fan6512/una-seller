package meli

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupRefreshTest(t *testing.T) (*Service, *pgxpool.Pool, string, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	var owner, storeID string
	if err = db.QueryRow(ctx, "SELECT gen_random_uuid()::text").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	s, err := NewService(db, Config{
		ClientID: "synthetic_app", ClientSecret: "synthetic_secret",
		RedirectURI:        "https://example.invalid" + CallbackPath,
		TokenEncryptionKey: key,
	})
	if err != nil {
		t.Fatal(err)
	}
	seller := strconv.FormatInt(time.Now().UnixNano(), 10)
	access, err := s.cipher.Seal("synthetic_old_access", tokenAAD(owner, seller, "access"))
	if err != nil {
		t.Fatal(err)
	}
	refresh, err := s.cipher.Seal("synthetic_old_refresh", tokenAAD(owner, seller, "refresh"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(ctx, "INSERT INTO app.users(id) VALUES($1::uuid)", owner); err != nil {
		t.Fatal(err)
	}
	const marketJSON = `[{"user_id":987654,"site_id":"MLB","logistic_type":"remote"}]`
	err = db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
 (owner_user_id,seller_id,nickname,access_token_ciphertext,refresh_token_ciphertext,
 access_token_expires_at,marketplaces,marketplaces_status)
 VALUES ($1::uuid,$2,'Synthetic seller',$3,$4,now()-interval '2 minutes',$5::jsonb,'verified')
 RETURNING id::text`, owner, seller, access, refresh, marketJSON).Scan(&storeID)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_, _ = db.Exec(c, "DELETE FROM app.meli_store_connections WHERE id=$1::uuid", storeID)
		_, _ = db.Exec(c, "DELETE FROM app.meli_oauth_states WHERE owner_user_id=$1::uuid", owner)
		_, _ = db.Exec(c, "DELETE FROM app.users WHERE id=$1::uuid", owner)
	})
	return s, db, owner, storeID
}

func TestRefreshSingleflightAndLease(t *testing.T) {
	s, db, owner, storeID := setupRefreshTest(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth/token" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		if err := r.ParseForm(); err != nil || r.PostForm.Get("grant_type") != "refresh_token" ||
			r.PostForm.Get("refresh_token") != "synthetic_old_refresh" {
			http.Error(w, "unexpected refresh request", http.StatusBadRequest)
			return
		}
		calls.Add(1)
		time.Sleep(80 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"synthetic_new_access","refresh_token":"synthetic_new_refresh","expires_in":10800}`))
	}))
	defer server.Close()
	s.apiURL = server.URL
	s2, err := NewService(db, s.cfg)
	if err != nil {
		t.Fatal(err)
	}
	s2.apiURL = server.URL
	const count = 10
	var wg sync.WaitGroup
	errs := make(chan error, count)
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			current := s
			if i%2 == 1 {
				current = s2
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			auth, err := current.RequireStoreConnection(ctx, owner, storeID)
			if err == nil && auth.AccessToken != "synthetic_new_access" {
				err = fmt.Errorf("got stale access token")
			}
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if n := calls.Load(); n != 1 {
		t.Fatalf("refresh token used %d times, want once", n)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	auth, childID, err := s.RequireSiteConnection(ctx, owner, storeID, "MLB", "remote")
	if err != nil || childID != "987654" || auth.SellerID == "" {
		t.Fatalf("verified site mapping not preserved: %s, %v", childID, err)
	}
	if _, _, err := s.RequireSiteConnection(ctx, owner, storeID, "MLA", "remote"); !errors.Is(err, ErrSiteNotAuthorized) {
		t.Fatalf("unmapped site must be rejected: %v", err)
	}
	otherOwner := "550e8400-e29b-41d4-a716-446655440000"
	if _, err := s.RequireStoreConnection(ctx, otherOwner, storeID); !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("cross-owner read was permitted: %v", err)
	}
	var status string
	var refreshed []byte
	if err := db.QueryRow(ctx, `SELECT status,refresh_token_ciphertext FROM app.meli_store_connections
 WHERE id=$1::uuid`, storeID).Scan(&status, &refreshed); err != nil {
		t.Fatal(err)
	}
	if status != "connected" {
		t.Fatalf("unexpected connection state: %s", status)
	}
	oldSeller := auth.SellerID
	next, err := s.cipher.Open(refreshed, tokenAAD(owner, oldSeller, "refresh"))
	if err != nil || next != "synthetic_new_refresh" {
		t.Fatal("rotated refresh token was not atomically saved")
	}
}

func TestInvalidGrantStopsRefresh(t *testing.T) {
	s, db, owner, id := setupRefreshTest(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()
	s.apiURL = server.URL
	ctx := context.Background()
	if _, err := s.RequireStoreConnection(ctx, owner, id); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("invalid_grant not handled: %v", err)
	}
	if _, err := s.RequireStoreConnection(ctx, owner, id); !errors.Is(err, ErrReauthorizationRequired) {
		t.Fatalf("reauthorization must be required: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("invalid_grant was retried %d times", calls.Load())
	}
	var status string
	if err := db.QueryRow(ctx, "SELECT status FROM app.meli_store_connections WHERE id=$1::uuid", id).Scan(&status); err != nil || status != "reauth_required" {
		t.Fatalf("wrong state: %s (%v)", status, err)
	}
}

func TestUncertainRefreshNeverReusesOldToken(t *testing.T) {
	s, db, owner, id := setupRefreshTest(t)
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "temporary problem", http.StatusBadGateway)
	}))
	defer server.Close()
	s.apiURL = server.URL
	for i := 0; i < 2; i++ {
		if _, err := s.RequireStoreConnection(context.Background(), owner, id); !errors.Is(err, ErrRefreshUnknown) {
			t.Fatalf("refresh failure must not be silently retried: %v", err)
		}
	}
	if calls.Load() != 1 {
		t.Fatalf("uncertain refresh attempted %d times", calls.Load())
	}
	var status string
	if err := db.QueryRow(context.Background(), "SELECT status FROM app.meli_store_connections WHERE id=$1::uuid", id).Scan(&status); err != nil || status != "refresh_unknown" {
		t.Fatalf("uncertain refresh not recorded: %s (%v)", status, err)
	}
}

func TestStaleRefreshLeaseFailsClosed(t *testing.T) {
	s, db, owner, id := setupRefreshTest(t)
	ctx := context.Background()
	_, err := db.Exec(ctx, `UPDATE app.meli_store_connections
 SET refresh_lock=gen_random_uuid(),refresh_lease_until=now()-interval '1 second'
 WHERE id=$1::uuid`, id)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.RequireStoreConnection(ctx, owner, id); !errors.Is(err, ErrRefreshUnknown) {
		t.Fatalf("stale lease must fail closed: %v", err)
	}
	var status string
	if err := db.QueryRow(ctx, "SELECT status FROM app.meli_store_connections WHERE id=$1::uuid", id).Scan(&status); err != nil || status != "refresh_unknown" {
		t.Fatalf("stale lease state was not persisted: %s (%v)", status, err)
	}
}
