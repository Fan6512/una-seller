package httpapi

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func TestProductBoxPromoteSnapshotOwnerIsolationAndEdit(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.DB.Close()
	if err = migrations.Apply(ctx, db.DB); err != nil {
		t.Fatal(err)
	}
	owner := "2a9c0000-0000-4000-8000-000000000081"
	other := "2a9c0000-0000-4000-8000-000000000082"
	defer func() {
		cleanup := context.Background()
		db.DB.Exec(cleanup, "DELETE FROM app.own_products WHERE owner_user_id=$1::uuid", owner)
		db.DB.Exec(cleanup, "DELETE FROM app.source_products WHERE owner_user_id=$1::uuid", owner)
		db.DB.Exec(cleanup, "DELETE FROM app.users WHERE id=$1::uuid", owner)
	}()
	input := collections.Input{
		Platform: "amazon", Site: "US", SourceURL: "https://www.amazon.com/dp/B000000081", Status: "complete",
		Title:       collections.Field[string]{Present: true, Value: "Captured original"},
		Description: collections.Field[string]{Present: true, Value: "Captured description"},
		Price:       collections.Field[string]{Present: true, Value: "12.50"},
		Currency:    collections.Field[string]{Present: true, Value: "USD"},
		Images: collections.Field[[]collections.ImageInput]{Present: true, Value: []collections.ImageInput{
			{URL: "https://m.media-amazon.com/images/I/test081.jpg", Type: "main"},
		}},
	}
	source, _, err := db.Save(ctx, owner, input)
	if err != nil {
		t.Fatal(err)
	}
	call := func(user, method, path, body string) *httptest.ResponseRecorder {
		handler := HandlerWithCollections(db, stubVerifier{id: user})
		request := httptest.NewRequest(method, path, strings.NewReader(body))
		if user != "" {
			request.Header.Set("Authorization", "Bearer synthetic-session")
		}
		if body != "" {
			request.Header.Set("Content-Type", "application/json")
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}
	approvePath := "/api/v1/collections/" + source.ID + "/approve"
	if w := call(other, "POST", approvePath, `{"expected_revision":0}`); w.Code != 404 {
		t.Fatalf("foreign approval %d", w.Code)
	}
	if w := call("", "GET", "/api/v1/products", ""); w.Code != 401 {
		t.Fatalf("unauthenticated %d", w.Code)
	}
	draftPath := "/api/v1/collections/" + source.ID + "/working-draft"
	draftBody := `{"title":"Curated title","description":"Curated description","notes":"Internal supplier","expected_revision":0}`
	if w := call(owner, "PUT", draftPath, draftBody); w.Code != 200 {
		t.Fatalf("draft %d %s", w.Code, w.Body.String())
	}
	if w := call(owner, "POST", approvePath, `{"expected_revision":0}`); w.Code != 409 {
		t.Fatalf("stale approval %d", w.Code)
	}
	first := call(owner, "POST", approvePath, `{"expected_revision":1}`)
	if first.Code != 201 {
		t.Fatalf("approval %d %s", first.Code, first.Body.String())
	}
	var response struct {
		Result  string                 `json:"result"`
		Product collections.OwnProduct `json:"product"`
	}
	if err = json.Unmarshal(first.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	id := response.Product.ID
	if !collections.ValidUserID(id) || response.Product.Title != "Curated title" ||
		response.Product.SourceTitle != "Captured original" ||
		response.Product.Notes != "Internal supplier" ||
		len(response.Product.Images) != 1 {
		t.Fatalf("wrong snapshot: %#v", response.Product)
	}
	again := call(owner, "POST", approvePath, `{"expected_revision":1}`)
	if again.Code != 200 || !strings.Contains(again.Body.String(), "already_approved") {
		t.Fatalf("duplicate approval %d", again.Code)
	}
	if w := call(owner, "GET", "/api/v1/collections?inbox=1", ""); w.Code != 200 || strings.Contains(w.Body.String(), source.ID) {
		t.Fatalf("approved item leaked into inbox %d", w.Code)
	}
	if w := call(owner, "GET", "/api/v1/collections?inbox=0", ""); w.Code != 200 || !strings.Contains(w.Body.String(), source.ID) {
		t.Fatalf("source provenance disappeared %d", w.Code)
	}
	list := call(owner, "GET", "/api/v1/products", "")
	if list.Code != 200 || !strings.Contains(list.Body.String(), id) {
		t.Fatalf("own list %d %s", list.Code, list.Body.String())
	}
	if w := call(other, "GET", "/api/v1/products/"+id, ""); w.Code != 404 {
		t.Fatalf("foreign own read %d", w.Code)
	}
	if w := call(other, "PUT", "/api/v1/products/"+id, `{"title":"Wrong","description":"","notes":"","expected_revision":1}`); w.Code != 404 {
		t.Fatalf("foreign own write %d", w.Code)
	}
	edited := call(owner, "PUT", "/api/v1/products/"+id, `{"title":"Edited product","description":"Own description","notes":"Own notes","expected_revision":1}`)
	if edited.Code != 200 {
		t.Fatalf("own edit %d %s", edited.Code, edited.Body.String())
	}
	if w := call(owner, "PUT", "/api/v1/products/"+id, `{"title":"Stale","description":"","notes":"","expected_revision":1}`); w.Code != 409 {
		t.Fatalf("stale own edit %d", w.Code)
	}
	input.Status = "partial"
	input.Title.Value = "Recaptured source"
	if _, _, err = db.Save(ctx, owner, input); err != nil {
		t.Fatal(err)
	}
	p, err := db.GetOwnProduct(ctx, owner, id)
	if err != nil || p.Title != "Edited product" || p.SourceTitle != "Captured original" || p.Revision != 2 {
		t.Fatalf("recapture polluted own product: %#v %v", p, err)
	}
	if w := call(owner, "DELETE", "/api/v1/collections/"+source.ID, ""); w.Code != 204 {
		t.Fatalf("source deletion %d", w.Code)
	}
	p, err = db.GetOwnProduct(ctx, owner, id)
	if err != nil || p.SourceProductID != nil || p.Title != "Edited product" {
		t.Fatalf("source deletion removed own product: %#v %v", p, err)
	}
	list = call(owner, "GET", "/api/v1/products", "")
	if list.Code != 200 || !strings.Contains(list.Body.String(), id) {
		t.Fatalf("orphaned own product disappeared: %d", list.Code)
	}
}
