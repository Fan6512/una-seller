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
	draftBody := `{"title":"Curated title","description":"Curated description","notes":"Internal supplier","expected_revision":0,
"listing_fields":{"titles":{"en":"Curated English Title","es":"Título en español","pt":"Título em português"},"category_id":"CBT123","category_path":"Beleza / Cuidados",
"brand":"Example Brand","seller_sku":"P-001","purchase_price":"25.50","purchase_currency":"CNY",
"weight_g":"850","length_cm":"20.5","width_cm":"15","height_cm":"5",
"site_prices":[{"site":"MLB","price":"89.90","currency":"BRL"}],
"images":[{"url":"https://example.com/edited-main.jpg","type":"main"}],
"variants":[{"seller_sku":"P-BLACK","attributes":{"颜色":"黑色"},"price":"89.90","currency":"BRL","image_urls":[]}]}}`
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
	if !collections.ValidUserID(id) || response.Product.Title != "Curated English Title" ||
		response.Product.SourceTitle != "Captured original" ||
		response.Product.Notes != "Internal supplier" ||
		len(response.Product.Images) != 1 ||
		response.Product.ListingFields.WeightG != "850" ||
		response.Product.ListingFields.SellerSKU != "P-001" ||
		response.Product.ListingFields.CategoryID != "CBT123" || response.Product.ListingFields.Titles.Spanish != "Título en español" ||
		response.Product.ListingFields.Images == nil ||
		len(*response.Product.ListingFields.Images) != 1 {
		t.Fatalf("wrong snapshot or listing fields: %#v", response.Product)
	}
	again := call(owner, "POST", approvePath, `{"expected_revision":1}`)
	if again.Code != 200 || !strings.Contains(again.Body.String(), "already_approved") {
		t.Fatalf("duplicate approval %d", again.Code)
	}
	// Converted working drafts are permanently read-only; users must edit
	// app.own_products, never a stale working draft that will not be consumed.
	if w := call(owner, "PUT", draftPath,
		`{"title":"Invisible after approval","description":"wrong","notes":"wrong","expected_revision":1}`); w.Code != 409 || !strings.Contains(w.Body.String(), "read-only after product-box approval") {
		t.Fatalf("converted draft still writable: %d %s", w.Code, w.Body.String())
	}
	if w := call(other, "PUT", draftPath, draftBody); w.Code != 404 {
		t.Fatalf("foreign converted draft write: %d", w.Code)
	}
	if w := call(owner, "GET", "/api/v1/collections?inbox=1", ""); w.Code != 200 || strings.Contains(w.Body.String(), source.ID) {
		t.Fatalf("approved item leaked into inbox %d", w.Code)
	}
	if w := call(owner, "GET", "/api/v1/collections?inbox=0", ""); w.Code != 200 || !strings.Contains(w.Body.String(), source.ID) {
		t.Fatalf("source provenance disappeared %d", w.Code)
	}
	if w := call(owner, "GET", "/api/v1/collections?inbox=0", ""); w.Code != 200 ||
		!strings.Contains(w.Body.String(), `"own_product_id":"`+id+`"`) {
		t.Fatalf("converted source summary has no own product reference: %d %s", w.Code, w.Body.String())
	}
	if w := call(owner, "GET", "/api/v1/collections/"+source.ID, ""); w.Code != 200 ||
		!strings.Contains(w.Body.String(), `"own_product_id":"`+id+`"`) {
		t.Fatalf("converted source detail has no own product reference: %d %s", w.Code, w.Body.String())
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
	// The own-product editor can update merchant fields without changing source facts.
	updated := call(owner, "PUT", "/api/v1/products/"+id,
		`{"title":"Edited product","description":"Own description","notes":"Own notes","expected_revision":2,
		"listing_fields":{"seller_sku":"P-002","weight_g":"1200","length_cm":"23","width_cm":"19","height_cm":"9",
		"titles":{"en":"Edited English Title","es":"Otro título en español","pt":"Outro título em português"},"category_id":"CBT999","site_prices":[{"site":"MLM","price":"199.90","currency":"MXN"}]}}`)
	if updated.Code != 200 {
		t.Fatalf("listing fields edit %d %s", updated.Code, updated.Body.String())
	}
	var editedResponse struct {
		Product collections.OwnProduct `json:"product"`
	}
	if err = json.Unmarshal(updated.Body.Bytes(), &editedResponse); err != nil {
		t.Fatal(err)
	}
	if editedResponse.Product.ListingFields.WeightG != "1200" ||
		editedResponse.Product.ListingFields.SellerSKU != "P-002" ||
		editedResponse.Product.ListingFields.CategoryID != "CBT999" || editedResponse.Product.ListingFields.Titles.Portuguese != "Outro título em português" {
		t.Fatalf("merchant listing edit not persisted: %#v", editedResponse.Product.ListingFields)
	}
	if w := call(owner, "PUT", "/api/v1/products/"+id,
		`{"title":"Invalid fields","description":"","notes":"","expected_revision":3,
		"listing_fields":{"weight_g":"-20"}}`); w.Code != 400 {
		t.Fatalf("invalid listing weight accepted: %d", w.Code)
	}

	input.Status = "partial"
	input.Title.Value = "Recaptured source"
	if _, _, err = db.Save(ctx, owner, input); err != nil {
		t.Fatal(err)
	}
	p, err := db.GetOwnProduct(ctx, owner, id)
	if err != nil || p.Title != "Edited English Title" || p.SourceTitle != "Captured original" || p.Revision != 3 || p.ListingFields.WeightG != "1200" {
		t.Fatalf("recapture polluted own product: %#v %v", p, err)
	}
	if w := call(owner, "DELETE", "/api/v1/collections/"+source.ID, ""); w.Code != 204 {
		t.Fatalf("source deletion %d", w.Code)
	}
	p, err = db.GetOwnProduct(ctx, owner, id)
	if err != nil || p.SourceProductID != nil || p.Title != "Edited English Title" {
		t.Fatalf("source deletion removed own product: %#v %v", p, err)
	}
	list = call(owner, "GET", "/api/v1/products", "")
	if list.Code != 200 || !strings.Contains(list.Body.String(), id) {
		t.Fatalf("orphaned own product disappeared: %d", list.Code)
	}
}
