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
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func TestBillingPreviewRejectsInvalidRequestsBeforeDatabase(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	const path = "/api/v1/stores/00000000-0000-4000-8000-000000000001/orders/billing-preview"
	const scope = "?site=MLB&logistic_type=remote"
	for _, tc := range []struct {
		method string
		query  string
		body   string
		auth   bool
		want   int
	}{
		{http.MethodGet, scope, "", true, http.StatusMethodNotAllowed},
		{http.MethodPost, scope, `{"order_ids":["11"]}`, false, http.StatusUnauthorized},
		{http.MethodPost, "?site=MLB", `{"order_ids":["11"]}`, true, http.StatusBadRequest},
		{http.MethodPost, scope + "&site=MLM", `{"order_ids":["11"]}`, true, http.StatusBadRequest},
		{http.MethodPost, scope, `{"order_ids":["11"],"unknown":true}`, true, http.StatusBadRequest},
		{http.MethodPost, scope, `{"order_ids":["11"]} {}`, true, http.StatusBadRequest},
		{http.MethodPost, scope, `{"order_ids":["11"]}` + strings.Repeat(" ", 4096), true, http.StatusBadRequest},
	} {
		r := httptest.NewRequest(tc.method, path+tc.query, strings.NewReader(tc.body))
		if tc.auth {
			r.Header.Set("Authorization", "Bearer test-only")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("billing %s: status=%d want=%d body=%s", tc.method, w.Code, tc.want, w.Body.String())
		}
		if tc.want == http.StatusMethodNotAllowed && w.Header().Get("Allow") != http.MethodPost {
			t.Fatal("missing POST Allow header")
		}
	}
}

func TestBillingPreviewErrorMapping(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
		code string
	}{
		{meli.ErrOrderBillingInvalid, http.StatusBadRequest, "order_billing_invalid"},
		{meli.ErrOrderNotFound, http.StatusNotFound, "order_not_found"},
		{meli.ErrSiteNotAuthorized, http.StatusForbidden, "order_site_not_authorized"},
		{meli.ErrOrderAuthorizationChanged, http.StatusConflict, "order_auth_changed"},
		{meli.ErrOrderBusy, http.StatusConflict, "order_busy"},
		{meli.ListingPlatformError(http.StatusTooManyRequests), http.StatusTooManyRequests, "order_rate_limited"},
		{meli.ListingPlatformError(http.StatusNotFound), http.StatusBadGateway, "order_platform_unavailable"},
	} {
		w := httptest.NewRecorder()
		writeOrderBillingError(w, tc.err)
		if tc.want == http.StatusTooManyRequests && w.Header().Get("Access-Control-Expose-Headers") != "Retry-After" {
			t.Fatal("desktop cannot read the retry header over CORS")
		}
		if w.Code != tc.want || !strings.Contains(w.Body.String(), tc.code) {
			t.Fatalf("billing error %v: status=%d body=%s", tc.err, w.Code, w.Body.String())
		}
	}
}

func TestBillingCacheRouteRejectsInvalidRequestsBeforeDatabase(t *testing.T) {
	h := HandlerWithCollections(&collections.Store{}, stubVerifier{
		id: "6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	}, &meli.Service{})
	const path = "/api/v1/stores/not-a-uuid/orders/10001/billing-cache"
	const scope = "?site=MLB&logistic_type=remote"
	for _, tc := range []struct {
		method string
		query  string
		auth   bool
		want   int
	}{
		{http.MethodPost, scope, true, http.StatusMethodNotAllowed},
		{http.MethodGet, scope, false, http.StatusUnauthorized},
		{http.MethodGet, "?site=MLB", true, http.StatusBadRequest},
		{http.MethodGet, scope + "&logistic_type=cross_border", true, http.StatusBadRequest},
		{http.MethodGet, scope + "&site=MLM", true, http.StatusBadRequest},
		{http.MethodGet, scope, true, http.StatusBadRequest},
	} {
		r := httptest.NewRequest(tc.method, path+tc.query, nil)
		if tc.auth {
			r.Header.Set("Authorization", "Bearer test-only")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("billing cache %s %s: status=%d want=%d body=%s", tc.method, tc.query, w.Code, tc.want, w.Body.String())
		}
		if tc.want == http.StatusMethodNotAllowed && w.Header().Get("Allow") != http.MethodGet {
			t.Fatal("missing GET Allow header")
		}
		if tc.want == http.StatusBadRequest && !strings.Contains(w.Body.String(), "order_") {
			t.Fatalf("unexpected bad-request body: %s", w.Body.String())
		}
	}
}

func TestOrderBillingHTTPReadsPersistedSnapshotWithoutUpstream(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing in CI")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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
	if err := db.QueryRow(ctx, `INSERT INTO app.users(id) VALUES(gen_random_uuid()) RETURNING id::text`).Scan(&owner); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM app.meli_store_connections WHERE owner_user_id=$1::uuid`, owner)
		_, _ = db.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1::uuid`, owner)
	}()
	const seller = "777001"
	key := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	cipher, err := meli.NewTokenCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	encrypted, err := cipher.Seal("synthetic_billing_access", []byte("meli|"+owner+"|"+seller+"|access"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
 (owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at,marketplaces,marketplaces_status)
 VALUES ($1::uuid,$2,$3,$3,now()+interval '1 hour',
 '[{"user_id":777002,"site_id":"MLB","logistic_type":"remote"}]','verified') RETURNING id::text`,
		owner, seller, encrypted).Scan(&store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,order_items)
 VALUES ($1::uuid,'MLB','remote','777002','10001','paid','USD',20,now(),now(),'80001','[]')`, store); err != nil {
		t.Fatal(err)
	}
	var billingCalls atomic.Int32
	var limited atomic.Bool
	transport := listingTestTransport(func(r *http.Request) (*http.Response, error) {
		json200 := func(body string) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/order/details"):
			billingCalls.Add(1)
			if limited.Load() {
				return &http.Response{StatusCode: http.StatusTooManyRequests,
					Header: http.Header{"Retry-After": {"120"}}, Body: io.NopCloser(strings.NewReader(`{}`))}, nil
			}
			return json200(`{"results":[{"order_id":10001,"payment_info":[{"status":"approved","money_release_status":"pending"}],
			 "details":[{"charge_info":{"detail_id":91,"detail_amount":11.10,"detail_type":"CHARGE","detail_sub_type":"CCSI","debited_from_operation":"YES"},
			 "currency_info":{"currency_id":"USD"},"shipping_info":{"shipping_id":80001}}]}]}`)
		case strings.Contains(r.URL.Path, "/marketplace/shipments/") && strings.HasSuffix(r.URL.Path, "/costs"):
			return json200(`{"currency_id":"USD","senders":[{"user_id":777002,"cost":3.38}],"receiver":{"cost":7.72}}`)
		default:
			t.Errorf("unexpected upstream path %s", r.URL.Path)
			return &http.Response{StatusCode: http.StatusNotFound, Header: make(http.Header),
				Body: io.NopCloser(strings.NewReader(""))}, nil
		}
	})
	service, err := meli.NewService(db, meli.Config{ClientID: "synthetic", ClientSecret: "synthetic",
		RedirectURI: "https://example.invalid" + meli.CallbackPath, TokenEncryptionKey: key,
		HTTPClient: &http.Client{Transport: transport}})
	if err != nil {
		t.Fatal(err)
	}
	handler := HandlerWithCollections(collectionStore, stubVerifier{id: owner}, service)
	const scope = "?site=MLB&logistic_type=remote"
	base := "/api/v1/stores/" + store + "/orders"
	do := func(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer test-only")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	type billingItem struct {
		OrderID    string     `json:"order_id"`
		Available  bool       `json:"available"`
		CheckedAt  *time.Time `json:"checked_at"`
		Reconciled string     `json:"reconciliation"`
		Shipping   *struct {
			SellerNetCostUSD string `json:"seller_net_cost_usd"`
		} `json:"shipping"`
	}
	decode := func(w *httptest.ResponseRecorder) billingItem {
		t.Helper()
		var page struct {
			Items []billingItem `json:"items"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil || len(page.Items) != 1 {
			t.Fatalf("billing response: %s (%v)", w.Body.String(), err)
		}
		return page.Items[0]
	}

	official := do(handler, http.MethodPost, base+"/billing-preview"+scope, `{"order_ids":["10001"]}`)
	if official.Code != http.StatusOK {
		t.Fatalf("official billing preview: status=%d body=%s", official.Code, official.Body.String())
	}
	first := decode(official)
	if !first.Available || first.CheckedAt == nil || first.Shipping == nil || first.Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("official billing preview lost data: %+v", first)
	}
	callsAfterPreview := billingCalls.Load()

	cached := do(handler, http.MethodGet, base+"/10001/billing-cache"+scope, "")
	if cached.Code != http.StatusOK {
		t.Fatalf("billing cache: status=%d body=%s", cached.Code, cached.Body.String())
	}
	restored := decode(cached)
	if !restored.Available || restored.CheckedAt == nil || !restored.CheckedAt.Equal(*first.CheckedAt) ||
		restored.Shipping == nil || restored.Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("persisted billing not returned by cache route: %+v", restored)
	}
	if billingCalls.Load() != callsAfterPreview {
		t.Fatal("cache route contacted the official platform")
	}
	if do(handler, http.MethodPost, base+"/10001/billing-cache"+scope, "").Code != http.StatusMethodNotAllowed {
		t.Fatal("cache route accepted a non-GET method")
	}
	foreign := do(HandlerWithCollections(collectionStore, stubVerifier{id: "00000000-0000-4000-8000-0000000000ff"}, service),
		http.MethodGet, base+"/10001/billing-cache"+scope, "")
	if foreign.Code != http.StatusNotFound || !strings.Contains(foreign.Body.String(), "order_not_found") {
		t.Fatalf("foreign owner read billing cache: status=%d body=%s", foreign.Code, foreign.Body.String())
	}

	limited.Store(true)
	limitedRec := do(handler, http.MethodPost, base+"/billing-preview"+scope, `{"order_ids":["10001"]}`)
	if limitedRec.Code != http.StatusTooManyRequests || limitedRec.Header().Get("Retry-After") == "" ||
		limitedRec.Header().Get("Access-Control-Expose-Headers") != "Retry-After" ||
		!strings.Contains(limitedRec.Body.String(), "order_rate_limited") {
		t.Fatalf("rate limit mapping: status=%d headers=%v body=%s",
			limitedRec.Code, limitedRec.Header(), limitedRec.Body.String())
	}
	callsAfterLimit := billingCalls.Load()
	cooldown := do(handler, http.MethodPost, base+"/billing-preview"+scope, `{"order_ids":["10001"]}`)
	if cooldown.Code != http.StatusTooManyRequests || billingCalls.Load() != callsAfterLimit {
		t.Fatalf("cooldown not enforced locally: status=%d calls=%d", cooldown.Code, billingCalls.Load())
	}
}
