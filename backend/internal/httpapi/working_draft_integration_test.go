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

func TestWorkingDraftOwnerAndRecaptureIsolation(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing")
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
	owner := "2a9c0000-0000-4000-8000-000000000071"
	other := "2a9c0000-0000-4000-8000-000000000072"
	defer func() {
		// Delete sources first: user deletion is FK-restricted.
		store.DB.Exec(context.Background(), "DELETE FROM app.source_products WHERE owner_user_id=$1::uuid", owner)
		store.DB.Exec(context.Background(), "DELETE FROM app.users WHERE id=$1::uuid", owner)
	}()

	input := collections.Input{
		Platform: "amazon", Site: "US", SourceURL: "https://www.amazon.com/dp/B000000071",
		Status: "complete", Title: collections.Field[string]{Present: true, Value: "Original capture"},
	}
	product, _, err := store.Save(ctx, owner, input)
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/collections/" + product.ID + "/working-draft"
	call := func(user, method, body string) *httptest.ResponseRecorder {
		h := HandlerWithCollections(store, stubVerifier{id: user})
		r := httptest.NewRequest(method, path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer synthetic-session")
		if method == http.MethodPut {
			r.Header.Set("Content-Type", "application/json")
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	initial := call(owner, http.MethodGet, "")
	if initial.Code != 200 || !strings.Contains(initial.Body.String(), `"draft":null`) {
		t.Fatalf("initial: %d %s", initial.Code, initial.Body.String())
	}
	payload := `{"title":"Edited draft","description":"Manual draft description","notes":"Supplier information","expected_revision":0}`
	for _, u := range []string{other} {
		if w := call(u, http.MethodGet, ""); w.Code != http.StatusNotFound {
			t.Fatalf("foreign GET=%d", w.Code)
		}
		if w := call(u, http.MethodPut, payload); w.Code != http.StatusNotFound {
			t.Fatalf("foreign PUT=%d", w.Code)
		}
	}
	if w := call(owner, http.MethodPut, `{"title":"","notes":"invalid"}`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid PUT=%d", w.Code)
	}
	if w := call(owner, http.MethodPut, payload); w.Code != http.StatusOK {
		t.Fatalf("save PUT=%d: %s", w.Code, w.Body.String())
	} else if !strings.Contains(w.Body.String(), `"revision":1`) {
		t.Fatalf("first revision missing: %s", w.Body.String())
	}
	if w := call(owner, http.MethodPut, payload); w.Code != http.StatusConflict {
		t.Fatalf("duplicate first writer should conflict: %d", w.Code)
	}
	// Source recapture must never overwrite the manual working draft.
	input.Status = "partial"
	input.Title.Value = "Recaptured original"
	if _, _, err = store.Save(ctx, owner, input); err != nil {
		t.Fatal(err)
	}
	got := call(owner, http.MethodGet, "")
	var page struct {
		Draft *collections.WorkingDraft `json:"draft"`
	}
	if got.Code != 200 || json.Unmarshal(got.Body.Bytes(), &page) != nil ||
		page.Draft == nil || page.Draft.SourceProductID != product.ID ||
		page.Draft.Revision != 1 || page.Draft.Title != "Edited draft" ||
		page.Draft.Description != "Manual draft description" {
		t.Fatalf("draft overwritten: %d %s", got.Code, got.Body.String())
	}
	source, err := store.Get(ctx, owner, product.ID)
	if err != nil || source.Title != "Recaptured original" {
		t.Fatalf("source changed incorrectly: %q err=%v", source.Title, err)
	}
	if w := call(owner, http.MethodPut, `{"title":"Second edit","description":"","notes":"","expected_revision":1}`); w.Code != 200 ||
		!strings.Contains(w.Body.String(), `"revision":2`) {
		t.Fatalf("second PUT=%d %s", w.Code, w.Body.String())
	}
	if w := call(owner, http.MethodPut, `{"title":"Stale edit","description":"wrong","notes":"","expected_revision":1}`); w.Code != http.StatusConflict {
		t.Fatalf("stale update should conflict: %d", w.Code)
	}
	if w := call(owner, http.MethodGet, ""); w.Code != 200 || !strings.Contains(w.Body.String(), "Second edit") {
		t.Fatalf("stale write replaced draft: %d", w.Code)
	}
	// A 20k-character Chinese description must fit the API byte limit.
	large := strings.Repeat("汉", 20000)
	body, _ := json.Marshal(collections.WorkingDraftInput{
		Title: "Edited draft", Description: large, ExpectedRevision: 2,
	})
	if w := call(owner, http.MethodPut, string(body)); w.Code != http.StatusOK {
		t.Fatalf("large Chinese draft rejected: status=%d", w.Code)
	}
	if w := call(owner, http.MethodGet, ""); w.Code != http.StatusOK {
		t.Fatalf("large Chinese draft lost: %d", w.Code)
	} else {
		var result struct {
			Draft *collections.WorkingDraft `json:"draft"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.Draft == nil ||
			len([]rune(result.Draft.Description)) != 20000 || result.Draft.Revision != 3 {
			t.Fatalf("large draft readback failed: %v", err)
		}
	}
	// 415, trailing JSON, unknown keys, >256KiB, and unsupported methods must fail.
	untrusted := HandlerWithCollections(store, stubVerifier{id: owner})
	request := func(method, payload, mime, origin string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, path, strings.NewReader(payload))
		r.Header.Set("Authorization", "Bearer synthetic-session")
		if mime != "" {
			r.Header.Set("Content-Type", mime)
		}
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		w := httptest.NewRecorder()
		untrusted.ServeHTTP(w, r)
		return w
	}
	if w := request(http.MethodPut, "{}", "text/plain", ""); w.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("non-JSON should be 415: %d", w.Code)
	}
	if w := request(http.MethodPut, payload+" {}", "application/json", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("trailing JSON should be 400: %d", w.Code)
	}
	if w := request(http.MethodPut, `{"title":"hello","unknown":1}`, "application/json", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("unknown key should be 400: %d", w.Code)
	}
	if w := request(http.MethodPut, `{"title":"A","description":"`+strings.Repeat("x", (256<<10)+1)+`","notes":"","expected_revision":3}`, "application/json", ""); w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized body should be 413: %d", w.Code)
	}
	if w := request(http.MethodPost, "", "", ""); w.Code != http.StatusMethodNotAllowed ||
		w.Header().Get("Allow") != "GET, PUT" {
		t.Fatalf("POST should be 405 and Allow GET PUT: %d %q", w.Code, w.Header().Get("Allow"))
	}
	if w := request(http.MethodDelete, "", "", ""); w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE should be 405: %d", w.Code)
	}
	if w := request(http.MethodPut, payload, "application/json", "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); w.Code != http.StatusForbidden {
		t.Fatalf("extension-origin PUT must be 403: %d", w.Code)
	}
	if err := store.Delete(ctx, owner, product.ID); err != nil {
		t.Fatal(err)
	}
	if w := call(owner, http.MethodGet, ""); w.Code != http.StatusNotFound {
		t.Fatalf("deleted source still readable: %d", w.Code)
	}
}
