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

// CI calls this test twice, with a real Docker restart of PostgreSQL between
// invocations. The fixture is intentionally NOT cleaned up in the "before" phase.
func TestPostgresProcessRestartPersistence(t *testing.T) {
	phase := os.Getenv("TEST_RESTART_PHASE")
	if phase == "" {
		t.Skip("executed separately in CI as before/restart/after")
	}
	if phase != "before" && phase != "after" {
		t.Fatalf("invalid TEST_RESTART_PHASE %q", phase)
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL is required for PostgreSQL restart test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.DB.Close()

	const owner = "2a9c0000-0000-4000-8000-000000000099"
	if phase == "before" {
		if err := migrations.Apply(ctx, store.DB); err != nil {
			t.Fatal(err)
		}
		var input collections.Input
		const sample = `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B765432109","title":"Synthetic restart fixture","price":"10.1250","currency":"USD","description":"Survives PostgreSQL process restart","bullet_points":["Persistence fixture"],"images":[{"url":"https://example.com/restart.jpg","type":"main"}],"variants":[{"attributes":{"Size":"S"},"price":"10.1250","currency":"USD","image_urls":["https://example.com/restart-variant.jpg"]}],"collection_status":"complete"}`
		if err := json.Unmarshal([]byte(sample), &input); err != nil {
			t.Fatal(err)
		}
		p, created, err := store.Save(ctx, owner, input)
		if err != nil || !created || len(p.Variants) != 1 || len(p.Images) != 1 {
			t.Fatalf("restart fixture seed: created=%v err=%v product=%+v", created, err, p)
		}
		t.Logf("PostgreSQL pre-restart fixture persisted: id=%s", p.ID)
		return
	}

	// Only runs after the CI runner restarts the actual PostgreSQL service
	// container; no migration or Save is called in this verification phase.
	items, err := store.List(ctx, owner, "Synthetic restart fixture", 10, 0)
	if err != nil || len(items) != 1 {
		t.Fatalf("restart fixture missing after process restart: err=%v items=%+v", err, items)
	}
	p, err := store.Get(ctx, owner, items[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if p.Price == nil || *p.Price != "10.1250" || p.Description == nil || *p.Description != "Survives PostgreSQL process restart" || len(p.Images) != 1 || len(p.BulletPoints) != 1 || len(p.Variants) != 1 || len(p.Variants[0].ImageURLs) != 1 {
		t.Fatalf("fixture content changed after PostgreSQL restart: %+v", p)
	}
	if _, err := store.DB.Exec(ctx, `DELETE FROM app.source_products WHERE owner_user_id=$1`, owner); err != nil {
		t.Fatal(err)
	}
	if _, err := store.DB.Exec(ctx, `DELETE FROM app.users WHERE id=$1`, owner); err != nil {
		t.Fatal(err)
	}
	t.Logf("PostgreSQL process-restart persistence verified: id=%s", p.ID)
}
