package collections_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/migrations"
)

// Runs only against GitHub Actions' temporary PG17 service. Never points at Supabase.
func TestPostgresCollectionRoundTrip(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing in CI")
		}
		t.Skip("TEST_DATABASE_URL not set (requires disposable PostgreSQL)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	store, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err = migrations.Apply(ctx, store.DB); err != nil {
		t.Fatal(err)
	}
	owner := "2a9c0000-0000-4000-8000-000000000011"
	defer func() {
		store.DB.Exec(context.Background(), `DELETE FROM app.source_products WHERE owner_user_id=$1`, owner)
		store.DB.Exec(context.Background(), `DELETE FROM app.users WHERE id=$1`, owner)
		store.DB.Close()
	}()
	var in collections.Input
	const full = `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/gp/product/B123456789?ref_=s","title":"Synthetic twin-size lamp","price":"29.9900","currency":"USD","category":"Lighting","description":"Synthetic item, for testing only.","bullet_points":["Low heat","Touch control"],"images":[{"url":"https://example.com/a.jpg","type":"main"},{"url":"https://example.com/b.jpg","type":"additional"}],"variants":[{"attributes":{"Color":"White","Size":"Small"},"price":"24.50","currency":"USD","image_urls":["https://example.com/white.jpg"]},{"attributes":{"Color":"Black","Size":"Large"},"price":null,"currency":"USD","image_urls":[]}],"collection_status":"complete"}`
	if err = json.Unmarshal([]byte(full), &in); err != nil {
		t.Fatal(err)
	}
	first, created, err := store.Save(ctx, owner, in)
	if err != nil {
		t.Fatal(err)
	}
	if !created || len(first.Images) != 2 || len(first.BulletPoints) != 2 || len(first.Variants) != 2 || first.Variants[1].Price != nil {
		t.Fatalf("not persisted: %+v", first)
	}
	if first.Price == nil || *first.Price != "29.9900" {
		t.Fatalf("not decimal exact: %v", first.Price)
	}
	again, created, err := store.Save(ctx, owner, in)
	if err != nil || created || again.ID != first.ID {
		t.Fatalf("duplicate: created=%t err=%v", created, err)
	}
	var partial collections.Input
	const fragment = `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B123456789?tag=track","collection_status":"partial","page_state":"buyboxMissing","price":null,"description":null,"bullet_points":[],"images":[],"variants":[{"attributes":{"Color":"White","Size":"Small"},"price":null,"image_urls":[]}],"warnings":["buybox_missing"]}`
	if err = json.Unmarshal([]byte(fragment), &partial); err != nil {
		t.Fatal(err)
	}
	last, created, err := store.Save(ctx, owner, partial)
	if err != nil || created {
		t.Fatalf("partial: created=%t err=%v", created, err)
	}
	if last.ID != first.ID || last.Status != "complete" || last.Description == nil || *last.Description != *first.Description || last.Price == nil || *last.Price != *first.Price || len(last.Images) != 2 || len(last.Variants) != 2 || len(last.BulletPoints) != 2 || len(last.Variants[0].ImageURLs) != 1 {
		t.Fatalf("partial erased saved content: %+v", last)
	}
	items, err := store.List(ctx, owner, "lamp", 10, 0)
	if err != nil || len(items) != 1 || items[0].ID != first.ID {
		t.Fatalf("list: %v %+v", err, items)
	}
	store.DB.Close()
	store, err = collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Get(ctx, owner, first.ID)
	if err != nil || reopened.ID != first.ID || len(reopened.Variants) != 2 {
		t.Fatalf("reconnect read: %v %+v", err, reopened)
	}
	other := "2a9c0000-0000-4000-8000-000000000012"
	if _, err = store.Get(ctx, other, first.ID); err != collections.ErrNotFound {
		t.Fatalf("owner isolation: %v", err)
	}
}
