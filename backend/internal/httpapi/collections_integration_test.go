package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func TestCollectionHTTPWithPostgres(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing in CI")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.DB.Close()
	if err = migrations.Apply(ctx, store.DB); err != nil {
		t.Fatal(err)
	}
	owner := "2a9c0000-0000-4000-8000-000000000013"
	defer func() {
		store.DB.Exec(context.Background(), `DELETE FROM app.source_products WHERE owner_user_id=$1`, owner)
		store.DB.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1`, owner)
	}()
	h := HandlerWithCollections(store, stubVerifier{id: owner})
	submit := func(body string) (*httptest.ResponseRecorder, map[string]any) {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/collections", strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer verified-test-session")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		return w, response
	}
	full := `{"source_platform":"amazon","source_site":"US","source_url":"https://amazon.com/dp/B777777777?tag=one","title":"Synthetic test product","price":"17.2500","currency":"USD","description":"Original details","bullet_points":["First feature","Second feature"],"images":[{"url":"https://example.com/main.jpg","type":"main"}],"variants":[{"attributes":{"Color":"Blue"},"price":"18.25","image_urls":["https://example.com/blue.jpg"]}],"collection_status":"complete"}`
	w, result := submit(full)
	if w.Code != http.StatusCreated || result["result"] != "created" {
		t.Fatalf("first status=%d: %s", w.Code, w.Body.String())
	}
	obj, ok := result["collection"].(map[string]any)
	if !ok {
		t.Fatal("missing collection")
	}
	id, ok := obj["id"].(string)
	if !ok || !collections.ValidUserID(id) {
		t.Fatalf("invalid id %v", obj["id"])
	}
	w, result = submit(full)
	if w.Code != http.StatusOK || result["result"] != "updated" {
		t.Fatalf("duplicate status=%d: %s", w.Code, w.Body.String())
	}
	fragment := `{"source_platform":"amazon","source_site":"US","source_url":"https://amazon.com/dp/B777777777?tag=two","collection_status":"partial","price":null,"description":null,"images":[],"variants":[]}`
	w, _ = submit(fragment)
	if w.Code != http.StatusOK {
		t.Fatalf("partial status=%d: %s", w.Code, w.Body.String())
	}
	for _, path := range []string{"/api/v1/collections?q=Synthetic", "/api/v1/collections/" + id} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("Authorization", "Bearer verified-test-session")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Original details") && strings.Contains(path, id) {
			t.Fatalf("GET %s status=%d: %s", path, w.Code, w.Body.String())
		}
	}
	// Authenticated users cannot see a record owned by another user.
	otherOwner := "3a9c0000-0000-4000-8000-000000000014"
	other := HandlerWithCollections(store, stubVerifier{id: otherOwner})
	for _, path := range []string{"/api/v1/collections", "/api/v1/collections/" + id} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("Authorization", "Bearer verified-other-session")
		w := httptest.NewRecorder()
		other.ServeHTTP(w, r)
		if strings.Contains(path, id) {
			if w.Code != http.StatusNotFound {
				t.Fatalf("other user read product: status=%d", w.Code)
			}
		} else if w.Code != http.StatusOK || strings.Contains(w.Body.String(), id) {
			t.Fatalf("other user listed product: status=%d body=%s", w.Code, w.Body.String())
		}
	}

	w, _ = submit(`{"source_platform":"amazon","source_site":"US","source_url":"https://evil.example/dp/B777777777","collection_status":"complete"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid host status=%d", w.Code)
	}
}
