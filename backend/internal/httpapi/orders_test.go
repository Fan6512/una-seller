package httpapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	svc, err := meli.NewService(cs.DB, meli.Config{
		ClientID: "synthetic", ClientSecret: "synthetic", RedirectURI: "https://example.invalid" + meli.CallbackPath,
		TokenEncryptionKey: base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901")),
		HTTPClient: &http.Client{Transport: listingTestTransport(func(*http.Request) (*http.Response, error) {
			t.Error("review or skip attempted an upstream request")
			return nil, context.Canceled
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
	w := post("sync", "")
	var blocked struct {
		Error  string
		Review meli.OrderReview
	}
	if err := json.Unmarshal(w.Body.Bytes(), &blocked); err != nil || w.Code != http.StatusConflict ||
		blocked.Error != "order_needs_review" || blocked.Review != (meli.OrderReview{OrderID: "102", ReviewID: reviewID, Offset: 3, Reason: "detail_not_found"}) {
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
}
