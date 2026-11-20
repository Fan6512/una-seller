package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func TestOrdersRequireAuthenticatedDesktopSession(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	for _, route := range []struct{ method, tail string }{{http.MethodGet, ""}, {http.MethodPost, "/sync"}, {http.MethodPost, "/skip"}} {
		r := httptest.NewRequest(route.method,
			"/api/v1/stores/00000000-0000-4000-8000-000000000001/orders"+route.tail+"?site=MLB&logistic_type=remote", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("%s %s unauthenticated order access: %d", route.method, route.tail, w.Code)
		}
	}
}

func TestOrdersRejectMissingScopeAndInvalidFiltersBeforeDatabase(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders/sync?site=MLB"},
		{http.MethodGet, "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders?site=MLB&logistic_type=remote&offset=-1"},
		{http.MethodGet, "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders?site=MLB&logistic_type=remote&status=not%3Bvalid"},
	}
	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		req.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s %s: status=%d body=%s", tt.method, tt.path, w.Code, w.Body.String())
		}
	}
}

func TestOrderUnsupportedMethodsReturn405(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	for _, tc := range []struct {
		method, tail, allowed string
	}{
		{http.MethodPut, "sync", http.MethodPost},
		{http.MethodGet, "skip", http.MethodPost},
		{http.MethodGet, "reset", http.MethodPost},
		{http.MethodPost, "", http.MethodGet},
	} {
		r := httptest.NewRequest(tc.method,
			"/api/v1/stores/00000000-0000-4000-8000-000000000001/orders"+map[bool]string{true: "/" + tc.tail, false: ""}[tc.tail != ""]+"?site=MLB&logistic_type=remote", nil)
		r.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusMethodNotAllowed || w.Header().Get("Allow") != tc.allowed {
			t.Fatalf("%s %s: status=%d allow=%s body=%s", tc.method, tc.tail, w.Code, w.Header().Get("Allow"), w.Body.String())
		}
	}
}

func TestOrderSkipRequiresExplicitOffset(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	for _, body := range []string{
		`{"order_id":"102","review_id":"00000000-0000-4000-8000-000000000001"}`,
		`{"order_id":"102","review_id":"00000000-0000-4000-8000-000000000001","offset":null}`,
		`{"order_id":"102","review_id":"00000000-0000-4000-8000-000000000001","offset":0} {}`,
	} {
		r := httptest.NewRequest(http.MethodPost,
			"/api/v1/stores/00000000-0000-4000-8000-000000000001/orders/skip?site=MLB&logistic_type=remote",
			strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "order_skip_invalid") {
			t.Fatalf("unexpected response code=%d body=%s", w.Code, w.Body.String())
		}
	}
}

func TestOrderPurchaseRejectsUnsupportedMethodAndMalformedJSON(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	const path = "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders/102/purchase?site=MLB&logistic_type=remote"
	for _, tc := range []struct {
		method, body string
		status       int
	}{
		{http.MethodPut, "", http.StatusMethodNotAllowed},
		{http.MethodPost, "{", http.StatusBadRequest},
		{http.MethodPost, "{\"purchase_order_id\":\"12345678\",\"cost_cny\":\"10\",\"updated_at\":null} {}", http.StatusBadRequest},
		{http.MethodPost, "{\"purchase_order_id\":\"12345678\",\"unknown\":true}", http.StatusBadRequest},
	} {
		r := httptest.NewRequest(tc.method, path, strings.NewReader(tc.body))
		r.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.status {
			t.Errorf("%s %q: status=%d expected=%d body=%s", tc.method, tc.body, w.Code, tc.status, w.Body.String())
		}
		if tc.status == http.StatusMethodNotAllowed && w.Header().Get("Allow") != "GET, POST" {
			t.Errorf("wrong Allow header: %q", w.Header().Get("Allow"))
		}
	}
}

func TestOrderReviewAndSkipHTTP(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cs, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.DB.Close()
	if err := migrations.Apply(ctx, cs.DB); err != nil {
		t.Fatal(err)
	}
	var owner, store, reviewID string
	err = cs.DB.QueryRow(ctx, `WITH owner AS (
 INSERT INTO app.users(id) VALUES(gen_random_uuid()) RETURNING id
), connection AS (
 INSERT INTO app.meli_store_connections(owner_user_id,seller_id,status,marketplaces,marketplaces_status)
 SELECT id,'987650','reauth_required','[{"user_id":987654,"site_id":"MLB","logistic_type":"remote"}]','verified' FROM owner
 RETURNING id,owner_user_id
), review AS (
 INSERT INTO app.meli_order_syncs(store_id,site_id,logistic_type,seller_id,window_from,window_to,
 next_offset,pending_order_ids,pending_page_count,review_id,review_reason)
 SELECT id,'MLB','remote','987654',now()-interval '1 day',now(),3,ARRAY['102'],1,gen_random_uuid(),'detail_not_found' FROM connection
 RETURNING review_id
)
SELECT owner_user_id::text,id::text,review_id::text FROM connection CROSS JOIN review`).Scan(&owner, &store, &reviewID)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = cs.DB.Exec(ctx, "DELETE FROM app.meli_store_connections WHERE id=$1::uuid", store)
		_, _ = cs.DB.Exec(ctx, "DELETE FROM app.users WHERE id=$1::uuid", owner)
	}()
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	upstreamCalls := 0
	svc, err := meli.NewService(cs.DB, meli.Config{
		ClientID: "synthetic", ClientSecret: "synthetic", RedirectURI: "https://example.invalid" + meli.CallbackPath,
		TokenEncryptionKey: key,
		HTTPClient: &http.Client{Transport: listingTestTransport(func(r *http.Request) (*http.Response, error) {
			upstreamCalls++
			if r.URL.Path != "/marketplace/orders/search" {
				t.Errorf("unexpected upstream request: %s", r.URL.Path)
			}
			return &http.Response{StatusCode: http.StatusTooManyRequests,
				Header: http.Header{"Retry-After": {"30"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})},
	})
	if err != nil {
		t.Fatal(err)
	}
	h := HandlerWithCollections(cs, stubVerifier{id: owner}, svc)
	post := func(tail, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/stores/"+store+"/orders/"+tail+"?site=MLB&logistic_type=remote", strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	// Recovery state survives a new client session and is scoped by the
	// current verified child seller, not by the client-provided offset.
	readState := func() meli.OrderSyncState {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/stores/"+store+"/orders/sync-state?site=MLB&logistic_type=remote", nil)
		req.Header.Set("Authorization", "Bearer test-only")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("sync state HTTP %d %s", rec.Code, rec.Body)
		}
		var state meli.OrderSyncState
		if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	state := readState()
	if state.CanReset || state.NextOffset != 3 || state.Review == nil ||
		state.Review.ReviewID != reviewID || state.Review.OrderID != "102" {
		t.Fatalf("expected persisted Review, got %+v", state)
	}
	w := post("sync", "")
	var blocked struct {
		Error  string
		Review meli.OrderReview
	}
	if err := json.Unmarshal(w.Body.Bytes(), &blocked); err != nil || w.Code != http.StatusConflict ||
		blocked.Error != "order_needs_review" || blocked.Review != (meli.OrderReview{OrderID: "102", ReviewID: reviewID, Offset: 3, Reason: "detail_not_found"}) ||
		!strings.Contains(w.Body.String(), "\"saved\":0") || !strings.Contains(w.Body.String(), "\"next_offset\":3") {
		t.Fatalf("unexpected review HTTP %d %s", w.Code, w.Body)
	}
	body := fmt.Sprintf(`{"order_id":"102","review_id":%q,"offset":3}`, reviewID)
	w = post("skip", body)
	var result meli.OrderSyncResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || w.Code != http.StatusOK ||
		result != (meli.OrderSyncResult{Processed: 1}) {
		t.Fatalf("final skip should return window completion and offset 0: HTTP %d %s", w.Code, w.Body)
	}
	w = post("skip", body)
	if w.Code != http.StatusConflict || !strings.Contains(w.Body.String(), "order_review_conflict") {
		t.Fatalf("skip replay HTTP %d %s", w.Code, w.Body)
	}
	if upstreamCalls != 0 {
		t.Fatal("review or skip attempted an upstream request")
	}

	t.Run("purchase version conflict", func(t *testing.T) {
		if _, err := cs.DB.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated)
 VALUES($1::uuid,'MLB','remote','987654','102','paid','USD',12.34,now(),now())`, store); err != nil {
			t.Fatal(err)
		}
		const initial = `{"purchase_order_id":"12345678","cost_cny":"10","updated_at":null}`
		created := post("102/purchase", initial)
		var value struct{ Purchase meli.OrderPurchase }
		if err := json.Unmarshal(created.Body.Bytes(), &value); err != nil || created.Code != http.StatusOK || value.Purchase.UpdatedAt.IsZero() {
			t.Fatalf("purchase create: %d %s", created.Code, created.Body)
		}
		for i, cost := range []string{"11", "12"} {
			edit := fmt.Sprintf(`{"purchase_order_id":"12345678","cost_cny":%q,"updated_at":%q}`, cost, value.Purchase.UpdatedAt.Format(time.RFC3339Nano))
			response := post("102/purchase", edit)
			if i == 0 && response.Code != http.StatusOK ||
				i == 1 && (response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "order_purchase_conflict")) {
				t.Fatalf("purchase client %d: %d %s", i, response.Code, response.Body)
			}
		}
	})


	t.Run("persisted reset offset", func(t *testing.T) {
		if _, err := cs.DB.Exec(ctx, `UPDATE app.meli_order_syncs
SET next_offset=3,window_from=now()-interval '1 day',window_to=now(),
pending_order_ids='{}'::text[],review_id=NULL,review_reason=NULL,lease_id=NULL,lease_until=NULL
WHERE store_id=$1::uuid`, store); err != nil {
			t.Fatal(err)
		}
		state := readState()
		if !state.CanReset || state.NextOffset != 3 || state.Review != nil {
			t.Fatalf("expected recoverable reset offset, got %+v", state)
		}
		reset := post("reset", `{"offset":3}`)
		if reset.Code != http.StatusOK {
			t.Fatalf("reset HTTP %d %s", reset.Code, reset.Body)
		}
		if state = readState(); state.CanReset || state.NextOffset != 0 {
			t.Fatalf("reset state not refreshed: %+v", state)
		}
	})

	t.Run("provider 429 and cooldown", func(t *testing.T) {
		cipher, err := meli.NewTokenCipher(key)
		if err != nil {
			t.Fatal(err)
		}
		access, err := cipher.Seal("synthetic", []byte("meli|"+owner+"|987650|access"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := cs.DB.Exec(ctx, `UPDATE app.meli_store_connections
 SET status='connected',access_token_ciphertext=$2,refresh_token_ciphertext=$2,
 access_token_expires_at=now()+interval '30 minutes' WHERE id=$1::uuid`, store, access); err != nil {
			t.Fatal(err)
		}
		for range 2 {
			response := post("sync", "")
			retry, err := strconv.Atoi(response.Header().Get("Retry-After"))
			if response.Code != http.StatusTooManyRequests || err != nil || retry < 29 || retry > 30 ||
				!strings.Contains(response.Body.String(), "order_rate_limited") {
				t.Fatalf("rate limit HTTP %d Retry-After=%q %s", response.Code, response.Header().Get("Retry-After"), response.Body)
			}
		}
		if upstreamCalls != 1 {
			t.Fatalf("cooldown allowed another upstream request, calls=%d", upstreamCalls)
		}
		var leased, reviewing bool
		if err := cs.DB.QueryRow(ctx, `SELECT lease_id IS NOT NULL, review_id IS NOT NULL
 FROM app.meli_order_syncs WHERE store_id=$1::uuid`, store).Scan(&leased, &reviewing); err != nil || leased || reviewing {
			t.Fatalf("429 left lease=%t review=%t error=%v", leased, reviewing, err)
		}
	})
}
