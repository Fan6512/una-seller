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

func disposableDB(t *testing.T, owner string) (*collections.Store, context.Context) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing in CI")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	store, err := collections.Open(ctx, dsn)
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	if err := migrations.Apply(ctx, store.DB); err != nil {
		store.DB.Close()
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_, _ = store.DB.Exec(cleanupCtx, `DELETE FROM app.source_products WHERE owner_user_id=$1`, owner)
		_, _ = store.DB.Exec(cleanupCtx, `DELETE FROM app.users WHERE id=$1`, owner)
		store.DB.Close()
		cancel()
	})
	return store, ctx
}

func inputJSON(t *testing.T, raw string) collections.Input {
	t.Helper()
	var in collections.Input
	if err := json.Unmarshal([]byte(raw), &in); err != nil {
		t.Fatal(err)
	}
	return in
}

func TestCompleteNullClearsArrays(t *testing.T) {
	const owner = "2a9c0000-0000-4000-8000-000000000014"
	store, ctx := disposableDB(t, owner)
	first := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B111111114","title":"Synthetic complete null test","description":"Old description","bullet_points":["Old bullet"],"images":[{"url":"https://example.com/old.jpg","type":"main"}],"variants":[{"attributes":{"Color":"Red"},"price":"1.00","image_urls":["https://example.com/variant.jpg"]}],"collection_status":"complete"}`)
	p, created, err := store.Save(ctx, owner, first)
	if err != nil || !created || len(p.Images) != 1 || len(p.Variants) != 1 {
		t.Fatalf("seed: created=%v err=%v product=%+v", created, err, p)
	}
	clear := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B111111114","collection_status":"complete","description":null,"bullet_points":null,"images":null,"variants":null}`)
	got, created, err := store.Save(ctx, owner, clear)
	if err != nil || created {
		t.Fatalf("clear: created=%v err=%v", created, err)
	}
	if got.ID != p.ID || got.Description != nil || len(got.BulletPoints) != 0 || len(got.Images) != 0 || len(got.Variants) != 0 {
		t.Fatalf("explicit null did not clear in complete capture: %+v", got)
	}
	// Omitted arrays after explicit clear must remain empty on a later recapture.
	again := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B111111114","collection_status":"partial","images":null,"variants":null,"bullet_points":null}`)
	got, _, err = store.Save(ctx, owner, again)
	if err != nil || len(got.Images) != 0 || len(got.Variants) != 0 || len(got.BulletPoints) != 0 {
		t.Fatalf("partial null unexpectedly changed data: err=%v product=%+v", err, got)
	}
}

func TestPartialVariantAppendKeepsOrder(t *testing.T) {
	const owner = "2a9c0000-0000-4000-8000-000000000015"
	store, ctx := disposableDB(t, owner)
	seed := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B111111115","title":"Synthetic variant order test","collection_status":"complete","variants":[{"attributes":{"Color":"A"}},{"attributes":{"Color":"B"}}]}`)
	p, created, err := store.Save(ctx, owner, seed)
	if err != nil || !created || len(p.Variants) != 2 {
		t.Fatalf("seed: created=%v err=%v product=%+v", created, err, p)
	}
	partial := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B111111115","collection_status":"partial","variants":[{"attributes":{"Color":"C"}}]}`)
	p, created, err = store.Save(ctx, owner, partial)
	if err != nil || created || len(p.Variants) != 3 {
		t.Fatalf("append: created=%v err=%v product=%+v", created, err, p)
	}
	for i, want := range []string{"A", "B", "C"} {
		if p.Variants[i].Attributes["Color"] != want {
			t.Fatalf("variant index %d: got %+v, want %s", i, p.Variants[i].Attributes, want)
		}
	}
	rows, err := store.DB.Query(ctx, `SELECT sort_order FROM app.source_product_variants WHERE source_product_id=$1 ORDER BY sort_order,id`, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	order := 0
	for rows.Next() {
		var found int
		if err := rows.Scan(&found); err != nil {
			t.Fatal(err)
		}
		if found != order {
			t.Fatalf("variant sort_order collision at index %d: %d", order, found)
		}
		order++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if order != 3 {
		t.Fatalf("expected 3 variant rows, got %d", order)
	}
}
