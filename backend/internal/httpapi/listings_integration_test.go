package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/migrations"
)

type listingTestTransport func(*http.Request) (*http.Response, error)

func (f listingTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPublishedListingHandler(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing in CI")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	collectionStore, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	db := collectionStore.DB
	defer db.Close()
	if err := migrations.Apply(ctx, db); err != nil {
		t.Fatal(err)
	}
	var owner, store string
	if err := db.QueryRow(ctx, "INSERT INTO app.users (id) VALUES (gen_random_uuid()) RETURNING id::text").Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(context.Background(), "DELETE FROM app.meli_store_connections WHERE owner_user_id=$1::uuid", owner)
		_, _ = db.Exec(context.Background(), "DELETE FROM app.users WHERE id=$1::uuid", owner)
	}()
	const seller = "678901"
	const access = "synthetic_listing_access"
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	cipher, err := meli.NewTokenCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := cipher.Seal(access, []byte("meli|"+owner+"|"+seller+"|access"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
        (owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at)
        VALUES ($1::uuid,$2,$3,$3,now()+interval '30 minutes') RETURNING id::text`, owner, seller, encrypted).Scan(&store); err != nil {
		t.Fatal(err)
	}

	detailCode := 200
	originalTransport := http.DefaultTransport
	http.DefaultTransport = listingTestTransport(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer "+access {
			t.Error("unexpected platform method or authorization")
		}
		code, body := 200, ""
		switch r.URL.Path {
		case "/marketplace/users/" + seller + "/items/search":
			body = `{"seller_id":678901,"results":["CBT1"],"paging":{"total":1,"offset":0,"limit":5}}`
		case "/marketplace/items/CBT1":
			code = detailCode
			body = `{"id":"CBT1","seller_id":678901,"site_id":"CBT","price":0}`
		default:
			t.Error("unexpected platform path")
			code = 404
		}
		return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	defer func() { http.DefaultTransport = originalTransport }()
	service, err := meli.NewService(db, meli.Config{ClientID: "synthetic", ClientSecret: "synthetic",
		RedirectURI: "https://example.invalid" + meli.CallbackPath, TokenEncryptionKey: key})
	if err != nil {
		t.Fatal(err)
	}
	handler := HandlerWithCollections(collectionStore, stubVerifier{id: owner}, service)
	path := "/api/v1/stores/" + store + "/published-listings"
	call := func(method, query string, status int, code string) map[string]any {
		t.Helper()
		req := httptest.NewRequest(method, path+query, nil)
		req.Header.Set("Authorization", "Bearer synthetic-session")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		var result map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if w.Code != status || (code != "" && result["code"] != code) || w.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("%s %s: HTTP %d %s", method, query, w.Code, w.Body.String())
		}
		if status == 405 && w.Header().Get("Allow") != "GET, POST" {
			t.Fatal("wrong allowed methods")
		}
		return result
	}
	call("DELETE", "", 405, "listing_method")
	call("POST", "?offset=1000", 400, "listing_offset")
	call("POST", "?offset=-1", 400, "listing_offset")
	call("GET", "?offset=100001", 400, "listing_offset")
	call("POST", "?offset=1", 409, "listing_cursor_changed")
	call("GET", "?site=MLB&logistic_type=remote", 403, "listing_site_unverified")
	for _, check := range []struct {
		upstream, status int
		code             string
	}{
		{401, 403, "listing_platform_auth"}, {403, 403, "listing_platform_auth"},
		{429, 429, "listing_rate_limited"}, {500, 502, "listing_platform_error"},
	} {
		detailCode = check.upstream
		result := call("POST", "", check.status, check.code)
		if result["platform_status"] != float64(check.upstream) {
			t.Fatal("missing platform status")
		}
	}
	blank := call("GET", "", 200, "")
	if len(blank["items"].([]any)) != 0 || blank["next_offset"] != float64(0) {
		t.Fatalf("interruption mutated cursor or details: %+v", blank)
	}
	detailCode = 200
	call("POST", "?offset=0", 200, "")
	first := call("GET", "", 200, "")
	last := call("GET", "?offset=20", 200, "")
	if len(first["items"].([]any)) != 1 || len(last["items"].([]any)) != 0 || first["next_offset"] != nil {
		t.Fatal("unexpected listing pages")
	}
	req := httptest.NewRequest("GET", path, nil)
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	if out.Code != 401 {
		t.Fatalf("unauthenticated status %d", out.Code)
	}
}
