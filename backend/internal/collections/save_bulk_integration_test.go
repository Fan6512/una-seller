package collections_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/jackc/pgx/v5/pgxpool"
)

func imageBatchInput(t *testing.T, url string, count int, status, prefix string) collections.Input {
	t.Helper()
	in := inputJSON(t, fmt.Sprintf(`{"source_platform":"amazon","source_site":"US","source_url":%q,"title":"Synthetic image batch test","collection_status":%q}`, url, status))
	images := make([]collections.ImageInput, count)
	variants := make([]collections.VariantInput, 2)
	for i := range images {
		images[i] = collections.ImageInput{
			URL:  fmt.Sprintf("https://example.com/%s/main-%03d.jpg", prefix, i),
			Type: "additional",
		}
	}
	for v := range variants {
		urls := make([]string, count)
		for j := range urls {
			urls[j] = fmt.Sprintf("https://example.com/%s/variant-%d-%03d.jpg", prefix, v, j)
		}
		variants[v] = collections.VariantInput{
			Attributes: map[string]string{"Color": fmt.Sprintf("Color-%d", v)},
			ImageURLs:  collections.Field[[]string]{Present: true, Value: urls},
		}
	}
	in.Images = collections.Field[[]collections.ImageInput]{Present: true, Value: images}
	in.Variants = collections.Field[[]collections.VariantInput]{Present: true, Value: variants}
	return in
}

// A larger image batch must not add database round trips. The existing
// variant-row lookup/update count is held constant at two variants.
func TestCollectionSaveImageQueryCountIndependentOfImageVolume(t *testing.T) {
	const owner = "2a9c0000-0000-4000-8000-000000000080"
	store, ctx := disposableDB(t, owner)
	for _, mode := range []string{"complete", "partial"} {
		var firstQueries int64
		for i, n := range []int{1, 20} {
			t.Run(fmt.Sprintf("%s_%d_images_per_group", mode, n), func(t *testing.T) {
				url := fmt.Sprintf("https://www.amazon.com/dp/B%09d", 800000000+i+100*len(mode))
				if mode == "partial" {
					seed := imageBatchInput(t, url, 1, "complete", fmt.Sprintf("seed-%d-%s", i, mode))
					if _, _, err := store.Save(ctx, owner, seed); err != nil {
						t.Fatal(err)
					}
				}
				cfg, err := pgxpool.ParseConfig(store.DB.Config().ConnString())
				if err != nil {
					t.Fatal(err)
				}
				tracer := &getQueryTracer{}
				cfg.ConnConfig.Tracer = tracer
				cfg.HealthCheckPeriod = time.Hour
				tracedPool, err := pgxpool.NewWithConfig(ctx, cfg)
				if err != nil {
					t.Fatal(err)
				}
				defer tracedPool.Close()
				if err := tracedPool.Ping(ctx); err != nil {
					t.Fatal(err)
				}
				tracer.count.Store(0)
				candidate := imageBatchInput(t, url, n, mode, fmt.Sprintf("new-%d-%s", i, mode))
				started := time.Now()
				p, created, err := (&collections.Store{DB: tracedPool}).Save(ctx, owner, candidate)
				elapsed := time.Since(started)
				if err != nil {
					t.Fatal(err)
				}
				if created != (mode == "complete") {
					t.Fatalf("created=%v, mode=%s", created, mode)
				}
				want := n
				if mode == "partial" {
					want++
				}
				if len(p.Images) != want || len(p.Variants) != 2 {
					t.Fatalf("incorrect batch result: main=%d variants=%d want=%d", len(p.Images), len(p.Variants), want)
				}
				for _, v := range p.Variants {
					if len(v.ImageURLs) != want {
						t.Fatalf("variant %s: %d images want %d", v.Attributes["Color"], len(v.ImageURLs), want)
					}
				}
				queries := tracer.count.Load()
				t.Logf("%s %d images/group: %d SQL commands, Save latency %s", mode, n, queries, elapsed)
				if queries == 0 {
					t.Fatal("tracer did not count database statements")
				}
				if i == 0 {
					firstQueries = queries
				} else if queries != firstQueries {
					t.Fatalf("image count increased SQL commands: 1 image=%d, %d images=%d", firstQueries, n, queries)
				}
			})
		}
	}
}

func TestCollectionSaveImageBatchCompletePartialSemantics(t *testing.T) {
	const owner = "2a9c0000-0000-4000-8000-000000000081"
	const url = "https://www.amazon.com/dp/B800000081"
	store, ctx := disposableDB(t, owner)
	seed := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B800000081","title":"Original batch title","collection_status":"complete","images":[{"url":"https://example.com/root-a.jpg","type":"main"},{"url":"https://example.com/root-a.jpg","type":"additional"},{"url":"https://example.com/root-b.jpg","type":"additional"}],"variants":[{"attributes":{"Color":"A"},"image_urls":["https://example.com/a-old.jpg","https://example.com/a-old.jpg"]},{"attributes":{"Color":"B"},"image_urls":[]}]}`)
	p, created, err := store.Save(ctx, owner, seed)
	if err != nil || !created || len(p.Images) != 3 || len(p.Variants) != 2 || len(p.Variants[0].ImageURLs) != 2 || len(p.Variants[1].ImageURLs) != 0 {
		t.Fatalf("complete seed retained duplicates: created=%v err=%v p=%+v", created, err, p)
	}
	if p.Images[0].Type != "main" || p.Images[1].Type != "additional" {
		t.Fatalf("complete image types changed: %+v", p.Images)
	}
	partial := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B800000081","collection_status":"partial","images":[{"url":"https://example.com/root-a.jpg","type":"additional"},{"url":"https://example.com/root-new.jpg","type":"additional"},{"url":"https://example.com/root-new.jpg","type":"main"},{"url":"https://example.com/root-end.jpg","type":"main"}],"variants":[{"attributes":{"Color":"A"},"image_urls":["https://example.com/a-old.jpg","https://example.com/a-new.jpg","https://example.com/a-new.jpg","https://example.com/a-end.jpg"]},{"attributes":{"Color":"B"},"image_urls":[]},{"attributes":{"Color":"C"},"image_urls":["https://example.com/c-new.jpg","https://example.com/c-new.jpg"]}]}`)
	p, created, err = store.Save(ctx, owner, partial)
	if err != nil || created || p.Status != "complete" {
		t.Fatalf("partial recapture: created=%v err=%v status=%s", created, err, p.Status)
	}
	wantMain := []string{"root-a.jpg", "root-a.jpg", "root-b.jpg", "root-new.jpg", "root-end.jpg"}
	if len(p.Images) != len(wantMain) {
		t.Fatalf("partial main images: %+v", p.Images)
	}
	for i, suffix := range wantMain {
		if p.Images[i].URL != "https://example.com/"+suffix {
			t.Fatalf("main image %d = %s; want suffix %s", i, p.Images[i].URL, suffix)
		}
	}
	if p.Images[3].Type != "additional" || p.Images[4].Type != "main" {
		t.Fatalf("partial URL duplicate changed first-seen image type: %+v", p.Images)
	}
	if len(p.Variants) != 3 || p.Variants[0].Attributes["Color"] != "A" || p.Variants[1].Attributes["Color"] != "B" || p.Variants[2].Attributes["Color"] != "C" {
		t.Fatalf("variant order changed: %+v", p.Variants)
	}
	wantA := []string{"a-old.jpg", "a-old.jpg", "a-new.jpg", "a-end.jpg"}
	if len(p.Variants[0].ImageURLs) != len(wantA) || len(p.Variants[1].ImageURLs) != 0 || len(p.Variants[2].ImageURLs) != 1 {
		t.Fatalf("variant group dedup incorrect: %+v", p.Variants)
	}
	for i, suffix := range wantA {
		if p.Variants[0].ImageURLs[i] != "https://example.com/"+suffix {
			t.Fatalf("variant A image %d = %s, want %s", i, p.Variants[0].ImageURLs[i], suffix)
		}
	}
	if p.Variants[1].ImageURLs == nil {
		t.Fatal("empty image_urls became nil")
	}
	if p.Variants[2].ImageURLs[0] != "https://example.com/c-new.jpg" {
		t.Fatalf("variant C image wrong: %+v", p.Variants[2])
	}
	// A complete recapture replaces old groups, preserving duplicate URLs
	// from the new complete input, and drops unmentioned variants.
	replace := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B800000081","collection_status":"complete","images":[{"url":"https://example.com/replacement.jpg","type":"main"},{"url":"https://example.com/replacement.jpg","type":"additional"}],"variants":[{"attributes":{"Color":"A"},"image_urls":["https://example.com/a-replaced.jpg","https://example.com/a-replaced.jpg"]}]}`)
	p, _, err = store.Save(ctx, owner, replace)
	if err != nil || len(p.Images) != 2 || len(p.Variants) != 1 || len(p.Variants[0].ImageURLs) != 2 {
		t.Fatalf("complete replacement: err=%v p=%+v", err, p)
	}
	noOp := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B800000081","collection_status":"partial","images":[],"variants":[{"attributes":{"Color":"A"},"image_urls":null}]}`)
	p, _, err = store.Save(ctx, owner, noOp)
	if err != nil || len(p.Images) != 2 || len(p.Variants) != 1 || len(p.Variants[0].ImageURLs) != 2 {
		t.Fatalf("partial null/empty erased images: err=%v p=%+v", err, p)
	}
	// A different authenticated owner must not access this product.
	if _, err := store.Get(ctx, "2a9c0000-0000-4000-8000-000000000090", p.ID); err != collections.ErrNotFound {
		t.Fatalf("cross-owner access: %v", err)
	}
	_ = url
}

func TestCollectionSaveBatchFailureRollsBack(t *testing.T) {
	const owner = "2a9c0000-0000-4000-8000-000000000082"
	const url = "https://www.amazon.com/dp/B800000082"
	store, ctx := disposableDB(t, owner)
	seed := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B800000082","title":"Original batch title","collection_status":"complete","images":[{"url":"https://example.com/old.jpg","type":"main"}]}`)
	original, _, err := store.Save(ctx, owner, seed)
	if err != nil {
		t.Fatal(err)
	}
	// Valid PostgreSQL int max forces image_order overflow on append, so
	// the batch INSERT fails after earlier UPDATE statements in the tx.
	if _, err := store.DB.Exec(ctx, `UPDATE app.source_product_images SET image_order=2147483647
		WHERE source_product_id=$1 AND variant_id IS NULL`, original.ID); err != nil {
		t.Fatal(err)
	}
	partial := inputJSON(t, `{"source_platform":"amazon","source_site":"US","source_url":"https://www.amazon.com/dp/B800000082","title":"Should roll back","collection_status":"partial","images":[{"url":"https://example.com/new.jpg","type":"additional"}]}`)
	if _, _, err := store.Save(ctx, owner, partial); err == nil {
		t.Fatal("batch image overflow unexpectedly committed")
	}
	got, err := store.Get(ctx, owner, original.ID)
	if err != nil || got.Title != "Original batch title" || len(got.Images) != 1 || got.Images[0].URL != "https://example.com/old.jpg" {
		t.Fatalf("transaction failed to roll back earlier updates: err=%v got=%+v", err, got)
	}
	_ = url
}
