package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgxpool"
)

type listingTestTransport func(*http.Request) (*http.Response, error)

func (f listingTestTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type listingQueryStart func(context.Context, pgx.TraceQueryStartData)

func (f listingQueryStart) TraceQueryStart(ctx context.Context, _ *pgx.Conn, query pgx.TraceQueryStartData) context.Context {
	f(ctx, query)
	return ctx
}

func (listingQueryStart) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

// Drop the server's COMMIT acknowledgment only, after PostgreSQL has executed it.
type listingLostCommitConn struct {
	net.Conn
	commit, ack []byte
	waiting     atomic.Bool
	ackDropped  *atomic.Bool
}

func (c *listingLostCommitConn) Write(data []byte) (int, error) {
	if bytes.Equal(data, c.commit) {
		c.waiting.Store(true)
	}
	return c.Conn.Write(data)
}

func (c *listingLostCommitConn) Read(data []byte) (int, error) {
	if !c.waiting.Swap(false) {
		return c.Conn.Read(data)
	}
	ack := make([]byte, len(c.ack))
	if _, err := io.ReadFull(c.Conn, ack); err != nil {
		return 0, err
	}
	if !bytes.Equal(ack, c.ack) {
		return 0, fmt.Errorf("expected a PostgreSQL COMMIT acknowledgment")
	}
	c.ackDropped.Store(true)
	return 0, io.ErrUnexpectedEOF
}

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
	var descriptionCalls atomic.Int32
	var descriptionFailure atomic.Int32
	searchTimeout, invalidSearch := false, false
	detailBody := `{"id":"CBT1","seller_id":678901,"site_id":"CBT","price":0,"status":"active"}`
	ids, total := `"CBT1"`, 1
	originalTransport := http.DefaultTransport
	http.DefaultTransport = listingTestTransport(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer "+access {
			t.Error("unexpected platform method or authorization")
		}
		code, body := 200, ""
		switch r.URL.Path {
		case "/marketplace/users/" + seller + "/items/search":
			if searchTimeout {
				return nil, context.DeadlineExceeded
			}
			if r.URL.Query().Get("orders") != "start_time_desc" {
				t.Error("published search did not use the fixed documented sort")
			}
			body = fmt.Sprintf(`{"seller_id":678901,"results":[%s],"paging":{"total":%d,"offset":%s,"limit":%s}}`,
				ids, total, r.URL.Query().Get("offset"), r.URL.Query().Get("limit"))
			if invalidSearch {
				body = `{"seller_id":678901,"results":[],"paging":{}}`
			}
		case "/marketplace/items/CBT1":
			code = detailCode
			body = detailBody
		case "/marketplace/items/CBT2":
			body = `{"id":"CBT2","seller_id":678901,"site_id":"CBT"}`
		case "/marketplace/items/CBT1/description":
			descriptionCalls.Add(1)
			if r.URL.RawQuery != "logistic_type=remote&site_id=MLB" {
				t.Error("wrong platform description scope")
			}
			body = `{"plain_text":"Platform description"}`
			switch descriptionFailure.Load() {
			case 1:
				return nil, io.ErrUnexpectedEOF
			case 2:
				code = http.StatusServiceUnavailable
			case 3:
				body = "{"
			}
		case "/marketplace/items/CBT3/description":
			descriptionCalls.Add(1)
			code = http.StatusNotFound
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
	call := func(method, query string, status int, code string, body ...string) map[string]any {
		t.Helper()
		var payload io.Reader
		if len(body) > 0 {
			payload = strings.NewReader(body[0])
		}
		req := httptest.NewRequest(method, path+query, payload)
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
	call("POST", "?offset=0&restart=0", 200, "")
	call("POST", "?restart=0", 409, "listing_cursor_changed")
	first := call("GET", "", 200, "")
	last := call("GET", "?offset=20", 200, "")
	if len(first["items"].([]any)) != 1 || len(last["items"].([]any)) != 0 || first["next_offset"] != nil {
		t.Fatal("unexpected listing pages")
	}

	detailCode = 404
	gone := call("POST", "?restart=1", 200, "")
	historical := call("GET", "", 200, "")["items"].([]any)[0].(map[string]any)
	if gone["saved"] != float64(0) || gone["skipped"] != float64(1) ||
		historical["status"] != "active" || historical["price"] != float64(0) || historical["detail_unavailable_at"] == nil {
		t.Fatal("404 did not retain and mark historical details")
	}
	detailCode = 206
	call("POST", "?restart=1", 200, "")
	restored := call("GET", "", 200, "")["items"].([]any)[0].(map[string]any)
	if restored["detail_unavailable_at"] != nil {
		t.Fatal("success did not clear the last 404")
	}

	detailCode = 200
	ids, total = `"CBT1","CBT2"`, 2
	var bad = `{"id":"CBT1","seller_id":678901,"site_id":"CBT","pictures":{}}`
	detailBody = bad
	response := call("POST", "?restart=1", 502, "listing_invalid_response")
	review := response["review"].(map[string]any)
	if review["listing_id"] != "CBT1" || review["offset"] != float64(0) ||
		review["reason"] != "detail_field_format" || review["skippable"] != true {
		t.Fatal("missing precise review location and reason")
	}
	if persisted := call("GET", "", 200, "")["review"].(map[string]any); persisted["listing_id"] != review["listing_id"] {
		t.Fatal("review not persisted for a new client")
	}

	skip := "?skip_invalid=1&offset=0"
	confirmation := `{"listing_id":"CBT1"}`
	call("POST", skip, 400, "listing_review_invalid")
	call("POST", skip, 409, "listing_review_changed", `{"listing_id":"CBT2"}`)
	ids = `"CBT2","CBT1"`
	call("POST", skip, 409, "listing_review_changed", confirmation)
	ids = `"CBT1","CBT2"`
	// The original review was eligible. The fresh check must still reject a changed identity.
	detailBody = `{"id":"CBT1","seller_id":678902,"site_id":"CBT","pictures":{}}`
	interrupted := call("POST", skip, 502, "listing_identity_mismatch", confirmation)
	if interrupted["review"].(map[string]any)["skippable"] != false {
		t.Fatal("changed seller was skippable")
	}
	call("POST", skip, 409, "listing_review_changed", confirmation)
	detailBody = "{"
	unreadable := call("POST", "?restart=0&offset=0", 502, "listing_invalid_response")["review"].(map[string]any)
	if unreadable["reason"] != "detail_payload_unreadable" || unreadable["skippable"] != false {
		t.Fatal("unreadable identity was confirmed for skipping")
	}
	detailBody = bad
	call("POST", "", 502, "listing_invalid_response")
	resumed := call("POST", skip, 200, "", confirmation)
	if resumed["saved"] != float64(1) || resumed["skipped"] != float64(1) || resumed["next_offset"] != nil {
		t.Fatal("confirmed skip did not advance and save the following record")
	}
	var skipped meli.ListingReview
	var skippedAt time.Time
	if err := db.QueryRow(ctx, `SELECT last_skipped_review, (last_skipped_review->>'confirmed_at')::timestamptz
        FROM app.meli_listing_syncs WHERE store_id=$1::uuid`, store).Scan(&skipped, &skippedAt); err != nil ||
		skipped.ListingID != "CBT1" || skipped.Reason != "detail_field_format" || skippedAt.IsZero() {
		t.Fatalf("confirmed skip was not recorded: %v", err)
	}
	settled := call("GET", "", 200, "")
	if settled["review"] != nil || len(settled["items"].([]any)) != 2 {
		t.Fatal("review remained pending after confirmed skip")
	}

	canceled, stop := context.WithCancel(ctx)
	stop()
	cancelResponse := httptest.NewRecorder()
	(storeAPI{service: service}).publishedListings(cancelResponse,
		httptest.NewRequest("POST", path+"?restart=1", nil).WithContext(canceled), owner, store)
	if cancelResponse.Code != http.StatusRequestTimeout || !strings.Contains(cancelResponse.Body.String(), "listing_canceled") {
		t.Fatalf("cancellation returned HTTP %d", cancelResponse.Code)
	}
	if after := call("GET", "", http.StatusOK, ""); !reflect.DeepEqual(after, settled) {
		t.Fatal("cancellation changed persisted listing state")
	}

	req := httptest.NewRequest("GET", path, nil)
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, req)
	if out.Code != 401 {
		t.Fatalf("unauthenticated status %d", out.Code)
	}

	detailBody = `{"id":"CBT1","seller_id":678901,"site_id":"CBT","price":0,"status":"active"}`
	ids, total = `"CBT1"`, 3
	call("POST", "?restart=1", http.StatusOK, "")
	saved := call("GET", "", http.StatusOK, "")
	if saved["next_offset"] != float64(1) {
		t.Fatal("expected a nonzero cursor before testing a failed restart")
	}
	searchTimeout = true
	call("POST", "?offset=0&restart=1", http.StatusGatewayTimeout, "listing_timeout")
	searchTimeout = false
	if after := call("GET", "", http.StatusOK, ""); !reflect.DeepEqual(after, saved) {
		t.Fatal("search timeout on restart changed persisted listing state")
	}
	invalidSearch = true
	call("POST", "?offset=0&restart=1", http.StatusBadGateway, "listing_invalid_response")
	invalidSearch = false
	if after := call("GET", "", http.StatusOK, ""); !reflect.DeepEqual(after, saved) {
		t.Fatal("invalid search response on restart changed persisted listing state")
	}

	faultConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	commit, err := (&pgproto3.Query{String: "commit"}).Encode(nil)
	if err != nil {
		t.Fatal(err)
	}
	ack, err := (&pgproto3.CommandComplete{CommandTag: []byte("COMMIT")}).Encode(nil)
	if err != nil {
		t.Fatal(err)
	}
	// The test pool talks only to the disposable CI database.
	var ackDropped atomic.Bool
	dial := faultConfig.ConnConfig.DialFunc
	faultConfig.ConnConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		conn, err := dial(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return &listingLostCommitConn{Conn: conn, commit: commit, ack: ack, ackDropped: &ackDropped}, nil
	}
	faultDB, err := pgxpool.NewWithConfig(ctx, faultConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer faultDB.Close()
	faultService, err := meli.NewService(faultDB, meli.Config{ClientID: "synthetic", ClientSecret: "synthetic",
		RedirectURI: "https://example.invalid" + meli.CallbackPath, TokenEncryptionKey: key})
	if err != nil {
		t.Fatal(err)
	}
	healthyHandler := handler
	handler = HandlerWithCollections(collectionStore, stubVerifier{id: owner}, faultService)
	ids = `"CBT2"`
	unknown := call("POST", "?offset=1", http.StatusBadGateway, "listing_commit_unknown")
	handler = healthyHandler
	if !ackDropped.Load() {
		t.Fatal("the PostgreSQL COMMIT acknowledgment was not observed and dropped")
	}
	if unknown["retryable"] != false {
		t.Fatal("an unknown commit was advertised as retryable")
	}
	if after := call("GET", "", http.StatusOK, ""); after["next_offset"] != float64(2) || len(after["items"].([]any)) != 2 {
		t.Fatal("a committed batch did not advance the persisted cursor after losing its acknowledgment")
	}
	call("POST", "?offset=1", http.StatusConflict, "listing_cursor_changed")

	t.Run("description HTTP route", func(t *testing.T) {
		const markets = `[{"site_id":"MLB","logistic_type":"remote","user_id":987654},
			{"site_id":"MLM","logistic_type":"remote","user_id":987655}]`
		if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces=$2::jsonb,
			marketplaces_status='verified' WHERE id=$1::uuid`, store, markets); err != nil {
			t.Fatal(err)
		}
		const otherSeller = "678902"
		otherCipher, err := cipher.Seal(access, []byte("meli|"+owner+"|"+otherSeller+"|access"))
		if err != nil {
			t.Fatal(err)
		}
		var otherStore string
		if err := db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
			(owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at,marketplaces,marketplaces_status)
			VALUES ($1::uuid,$2,$3,$3,now()+interval '30 minutes',$4::jsonb,'verified') RETURNING id::text`,
			owner, otherSeller, otherCipher, markets).Scan(&otherStore); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(ctx, `INSERT INTO app.meli_published_listings
			(store_id,site_id,logistic_type,listing_id,details) VALUES
			($1::uuid,'MLB','remote','MLB1','{"seller_id":"987654","cbt_item_id":"CBT1"}'),
			($1::uuid,'MLB','remote','MLB2','{"seller_id":"987655","cbt_item_id":"CBT1"}'),
			($1::uuid,'MLB','remote','MLB3','{"seller_id":"987654","cbt_item_id":"CBT3"}')`, store); err != nil {
			t.Fatal(err)
		}

		server := httptest.NewServer(handler)
		defer server.Close()
		otherOwner := httptest.NewServer(HandlerWithCollections(collectionStore,
			stubVerifier{id: "00000000-0000-4000-8000-000000000001"}, service))
		defer otherOwner.Close()

		rotated, err := cipher.Seal("synthetic_rotated_access", []byte("meli|"+owner+"|"+seller+"|access"))
		if err != nil {
			t.Fatal(err)
		}
		raceConfig := db.Config()
		var rotations atomic.Int32
		// Change credentials after listingScope but before the atomic listing lookup.
		raceConfig.ConnConfig.Tracer = listingQueryStart(func(ctx context.Context, query pgx.TraceQueryStartData) {
			if !strings.Contains(query.SQL, "item.details->>'cbt_item_id'") {
				return
			}
			tag, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_ciphertext=$2
				WHERE id=$1::uuid AND access_token_ciphertext=$3`, store, rotated, encrypted)
			if err != nil {
				t.Errorf("rotate synthetic credentials: %v", err)
			}
			rotations.Add(int32(tag.RowsAffected()))
		})
		raceDB, err := pgxpool.NewWithConfig(ctx, raceConfig)
		if err != nil {
			t.Fatal(err)
		}
		defer raceDB.Close()
		raceService, err := meli.NewService(raceDB, meli.Config{ClientID: "synthetic", ClientSecret: "synthetic",
			RedirectURI: "https://example.invalid" + meli.CallbackPath, TokenEncryptionKey: key})
		if err != nil {
			t.Fatal(err)
		}
		raceServer := httptest.NewServer(HandlerWithCollections(collectionStore, stubVerifier{id: owner}, raceService))
		defer raceServer.Close()

		const scope = "?site=MLB&logistic_type=remote"
		for _, check := range []struct {
			name, base, store, listing, query, method, token string
			status                                           int
			code, description                                string
			calls                                            int32
		}{
			{"success", server.URL, store, "MLB1", scope, "GET", "synthetic-session", 200, "", "Platform description", 1},
			{"no session", server.URL, store, "MLB1", scope, "GET", "", 401, "", "", 0},
			{"wrong owner", otherOwner.URL, store, "MLB1", scope, "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"other store", server.URL, otherStore, "MLB1", scope, "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"other verified site", server.URL, store, "MLB1", "?site=MLM&logistic_type=remote", "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"unverified logistics", server.URL, store, "MLB1", "?site=MLB&logistic_type=fulfillment", "GET", "synthetic-session", 403, "listing_site_unverified", "", 0},
			{"wrong child seller", server.URL, store, "MLB2", scope, "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"missing listing", server.URL, store, "MLB999", scope, "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"malformed listing ID", server.URL, store, "MLB-1", scope, "GET", "synthetic-session", 400, "listing_invalid_response", "", 0},
			{"duplicate site", server.URL, store, "MLB1", scope + "&site=MLM", "GET", "synthetic-session", 400, "listing_invalid_response", "", 0},
			{"platform 404", server.URL, store, "MLB3", scope, "GET", "synthetic-session", 404, "description_unavailable", "", 1},
			{"wrong method", server.URL, store, "MLB1", scope, "POST", "synthetic-session", 405, "listing_method", "", 0},
			{"platform transport", server.URL, store, "MLB1", scope, "GET", "synthetic-session", 502, "listing_transport", "", 1},
			{"platform 5xx", server.URL, store, "MLB1", scope, "GET", "synthetic-session", 502, "listing_unavailable", "", 1},
			{"malformed platform body", server.URL, store, "MLB1", scope, "GET", "synthetic-session", 502, "listing_unavailable", "", 1},
			{"authorization rotated", raceServer.URL, store, "MLB1", scope, "GET", "synthetic-session", 409, "listing_auth_changed", "", 0},
			{"listing missing despite rotation", raceServer.URL, store, "MLB999", scope, "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"wrong child despite rotation", raceServer.URL, store, "MLB2", scope, "GET", "synthetic-session", 404, "listing_store_missing", "", 0},
			{"database unavailable", raceServer.URL, store, "MLB1", scope, "GET", "synthetic-session", 502, "listing_unavailable", "", 0},
		} {
			t.Run(check.name, func(t *testing.T) {
				switch check.name {
				case "platform transport":
					descriptionFailure.Store(1)
				case "platform 5xx":
					descriptionFailure.Store(2)
				case "malformed platform body":
					descriptionFailure.Store(3)
				}
				defer descriptionFailure.Store(0)
				if check.name == "database unavailable" {
					raceDB.Close()
				} else if check.base == raceServer.URL {
					rotations.Store(0)
					if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_ciphertext=$2 WHERE id=$1::uuid`, store, encrypted); err != nil {
						t.Fatal(err)
					}
				}
				before := descriptionCalls.Load()
				req, err := http.NewRequestWithContext(ctx, check.method,
					check.base+"/api/v1/stores/"+check.store+"/published-listings/"+check.listing+"/description"+check.query, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("Origin", "http://tauri.localhost")
				if check.token != "" {
					req.Header.Set("Authorization", "Bearer "+check.token)
				}
				resp, err := server.Client().Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()
				var result struct {
					Code        string `json:"code"`
					Description string `json:"description"`
				}
				if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}
				if resp.StatusCode != check.status || result.Code != check.code || result.Description != check.description {
					t.Fatalf("HTTP %d, body %+v", resp.StatusCode, result)
				}
				if resp.Header.Get("Access-Control-Allow-Origin") != "http://tauri.localhost" ||
					(check.status == 200 && resp.Header.Get("Cache-Control") != "no-store") ||
					(check.status == 405 && resp.Header.Get("Allow") != "GET") {
					t.Fatalf("unexpected response headers: %v", resp.Header)
				}
				if got := descriptionCalls.Load() - before; got != check.calls {
					t.Fatalf("platform calls = %d, want %d", got, check.calls)
				}
				if check.base == raceServer.URL && check.name != "database unavailable" && rotations.Load() != 1 {
					t.Fatal("credential rotation was not exercised")
				}
			})
		}
	})
}
