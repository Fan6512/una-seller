package collections_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Count the database commands made by Get, rather than assuming a fixed count
// from the implementation. This regression test runs against disposable PG17.
type getQueryTracer struct {
	count atomic.Int64
}

func (t *getQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	t.count.Add(1)
	return ctx
}

func (t *getQueryTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestCollectionGetFixedQueryCountAndImageOrdering(t *testing.T) {
	for _, count := range []int{0, 1, 8, 30} {
		t.Run(fmt.Sprintf("%d_variants", count), func(t *testing.T) {
			owner := fmt.Sprintf("2a9c0000-0000-4000-8000-%012d", 200+count)
			store, ctx := disposableDB(t, owner)
			if _, err := store.DB.Exec(ctx, `INSERT INTO app.users(id) VALUES($1) ON CONFLICT DO NOTHING`, owner); err != nil {
				t.Fatal(err)
			}
			var productID string
			url := fmt.Sprintf("https://example.com/get-bulk/%d", count)
			err := store.DB.QueryRow(ctx, `INSERT INTO app.source_products(owner_user_id,source_platform,source_site,source_url,normalized_url,title,collection_status)
				VALUES($1,'amazon','US',$2,$2,'Synthetic bulk-read product','complete') RETURNING id::text`, owner, url).Scan(&productID)
			if err != nil {
				t.Fatal(err)
			}

			// Insert images out of order; both main and variant image order must
			// still be determined by image_order,id, never by insertion order.
			for _, order := range []int{2, 0, 1} {
				typ := "additional"
				if order == 0 {
					typ = "main"
				}
				imageURL := fmt.Sprintf("https://example.com/main-%d.jpg", order)
				if _, err := store.DB.Exec(ctx, `INSERT INTO app.source_product_images(source_product_id,image_url,image_type,image_order)
					VALUES($1,$2,$3,$4)`, productID, imageURL, typ, order); err != nil {
					t.Fatal(err)
				}
			}

			// Insert variants in reverse order to ensure sort_order wins.
			for i := count - 1; i >= 0; i-- {
				var variantID string
				attrs := fmt.Sprintf(`{"Color":"Color-%02d"}`, i)
				err := store.DB.QueryRow(ctx, `INSERT INTO app.source_product_variants(source_product_id,variant_key,attributes,sort_order)
					VALUES($1,$2,$3::jsonb,$4) RETURNING id::text`, productID, attrs, attrs, i).Scan(&variantID)
				if err != nil {
					t.Fatal(err)
				}
				orders := []int{}
				switch i % 3 {
				case 0:
					orders = []int{2, 0, 1}
				case 2:
					orders = []int{0}
				}
				for _, order := range orders {
					imageURL := fmt.Sprintf("https://example.com/variant-%02d-%d.jpg", i, order)
					if _, err := store.DB.Exec(ctx, `INSERT INTO app.source_product_images(source_product_id,variant_id,image_url,image_type,image_order)
						VALUES($1,$2,$3,'variant',$4)`, productID, variantID, imageURL, order); err != nil {
						t.Fatal(err)
					}
				}
			}

			cfg, err := pgxpool.ParseConfig(store.DB.Config().ConnString())
			if err != nil {
				t.Fatal(err)
			}
			tracer := &getQueryTracer{}
			cfg.ConnConfig.Tracer = tracer
			readPool, err := pgxpool.NewWithConfig(ctx, cfg)
			if err != nil {
				t.Fatal(err)
			}
			defer readPool.Close()
			if err := readPool.Ping(ctx); err != nil {
				t.Fatal(err)
			}
			tracer.count.Store(0)
			got, err := (&collections.Store{DB: readPool}).Get(ctx, owner, productID)
			if err != nil {
				t.Fatal(err)
			}
			if queries := tracer.count.Load(); queries != 3 {
				t.Fatalf("Get sent %d queries for %d variants, want fixed 3", queries, count)
			}
			if got.ID != productID || len(got.Variants) != count || len(got.Images) != 3 {
				t.Fatalf("incorrect product: %+v", got)
			}
			for order, image := range got.Images {
				want := fmt.Sprintf("https://example.com/main-%d.jpg", order)
				if image.URL != want {
					t.Fatalf("main image %d: got %q want %q", order, image.URL, want)
				}
			}
			if got.Images[0].Type != "main" || got.Images[1].Type != "additional" {
				t.Fatalf("image types changed: %+v", got.Images)
			}
			if got.Images == nil || got.BulletPoints == nil || got.Warnings == nil || got.Variants == nil {
				t.Fatal("empty arrays must remain non-nil")
			}
			for i, variant := range got.Variants {
				wantColor := fmt.Sprintf("Color-%02d", i)
				if variant.Attributes["Color"] != wantColor || variant.Price != nil || variant.Currency != nil {
					t.Fatalf("variant %d fields or order changed: %+v", i, variant)
				}
				if variant.ImageURLs == nil {
					t.Fatalf("variant %d has nil image_urls rather than []", i)
				}
				var wantedOrders []int
				switch i % 3 {
				case 0:
					wantedOrders = []int{0, 1, 2}
				case 2:
					wantedOrders = []int{0}
				}
				if len(variant.ImageURLs) != len(wantedOrders) {
					t.Fatalf("variant %d has %d images, want %d", i, len(variant.ImageURLs), len(wantedOrders))
				}
				for j, order := range wantedOrders {
					want := fmt.Sprintf("https://example.com/variant-%02d-%d.jpg", i, order)
					if variant.ImageURLs[j] != want {
						t.Fatalf("variant %d image %d: got %q want %q", i, j, variant.ImageURLs[j], want)
					}
				}
			}
			payload, err := json.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var shape struct {
				Images   []json.RawMessage `json:"images"`
				Variants []struct {
					ImageURLs []string `json:"image_urls"`
				} `json:"variants"`
			}
			if err := json.Unmarshal(payload, &shape); err != nil {
				t.Fatal(err)
			}
			if len(shape.Images) != 3 || len(shape.Variants) != count {
				t.Fatalf("unexpected JSON response: %s", payload)
			}
			if _, err := (&collections.Store{DB: readPool}).Get(ctx, "2a9c0000-0000-4000-8000-000000000099", productID); !errors.Is(err, collections.ErrNotFound) {
				t.Fatalf("cross-owner lookup: got %v, want ErrNotFound", err)
			}
		})
	}
}
