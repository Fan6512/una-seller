package collections

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// insertCollectionImages writes a complete or partial main-image batch in one
// statement. The caller owns the recapture row lock and the transaction. For
// complete captures Save has already deleted the previous main images.
func insertCollectionImages(ctx context.Context, tx pgx.Tx, productID string, images []ImageInput, partial bool) error {
	if len(images) == 0 {
		return nil
	}
	urls := make([]string, 0, len(images))
	types := make([]string, 0, len(images))
	for _, image := range images {
		urls = append(urls, image.URL)
		types = append(types, image.Type)
	}

	// Complete recaptures preserve incoming duplicate rows and input order.
	// Partial recaptures skip existing URLs and later duplicates in this batch.
	// MAX(image_order) is read once from the statement snapshot; row_number
	// allocates contiguous append positions after the existing highest order.
	_, err := tx.Exec(ctx, `WITH incoming AS (
  SELECT image_url,image_type,ordinality AS ord
  FROM unnest($2::text[],$3::text[]) WITH ORDINALITY AS u(image_url,image_type,ordinality)
), ranked AS (
  SELECT image_url,image_type,ord,
         ROW_NUMBER() OVER (PARTITION BY image_url ORDER BY ord) AS occurrence
  FROM incoming
), selected AS (
  SELECT r.image_url,r.image_type,r.ord
  FROM ranked r
  WHERE NOT $4::boolean OR (
    r.occurrence=1 AND NOT EXISTS (
      SELECT 1 FROM app.source_product_images existing
      WHERE existing.source_product_id=$1::uuid
        AND existing.variant_id IS NULL AND existing.image_url=r.image_url
    )
  )
), numbered AS (
  SELECT image_url,image_type,ord,ROW_NUMBER() OVER (ORDER BY ord)-1 AS image_offset
  FROM selected
)
INSERT INTO app.source_product_images(source_product_id,image_url,image_type,image_order)
SELECT $1::uuid,n.image_url,n.image_type,
       (COALESCE((SELECT MAX(image_order) FROM app.source_product_images
                  WHERE source_product_id=$1::uuid AND variant_id IS NULL),-1)+n.image_offset+1)::integer
FROM numbered n
ORDER BY n.ord`, productID, urls, types, partial)
	return err
}

// insertVariantImages writes all image groups for this recapture in one SQL
// command, even when the input contains many variants. Variant IDs are resolved
// by Save's existing upsert logic before this function is called.
func insertVariantImages(ctx context.Context, tx pgx.Tx, productID string, variantIDs, urls []string, partial bool) error {
	if len(urls) == 0 {
		return nil
	}
	// The position produced by WITH ORDINALITY is global. Per-variant
	// row_number restores each image group's original input order.
	_, err := tx.Exec(ctx, `WITH incoming AS (
  SELECT variant_id::uuid AS variant_id,image_url,ordinality AS ord
  FROM unnest($2::text[],$3::text[]) WITH ORDINALITY AS u(variant_id,image_url,ordinality)
), ranked AS (
  SELECT variant_id,image_url,ord,
         ROW_NUMBER() OVER (PARTITION BY variant_id,image_url ORDER BY ord) AS occurrence
  FROM incoming
), selected AS (
  SELECT r.variant_id,r.image_url,r.ord
  FROM ranked r
  WHERE NOT $4::boolean OR (
    r.occurrence=1 AND NOT EXISTS (
      SELECT 1 FROM app.source_product_images existing
      WHERE existing.variant_id=r.variant_id AND existing.image_url=r.image_url
    )
  )
), numbered AS (
  SELECT variant_id,image_url,ord,
         ROW_NUMBER() OVER (PARTITION BY variant_id ORDER BY ord)-1 AS image_offset
  FROM selected
), max_orders AS (
  SELECT existing.variant_id,MAX(existing.image_order) AS max_order
  FROM app.source_product_images existing
  WHERE existing.source_product_id=$1::uuid
    AND existing.variant_id IN (SELECT DISTINCT variant_id FROM numbered)
  GROUP BY existing.variant_id
)
INSERT INTO app.source_product_images(source_product_id,variant_id,image_url,image_type,image_order)
SELECT $1::uuid,n.variant_id,n.image_url,'variant',
       (COALESCE(m.max_order,-1)+n.image_offset+1)::integer
FROM numbered n
LEFT JOIN max_orders m ON m.variant_id=n.variant_id
ORDER BY n.ord`, productID, variantIDs, urls, partial)
	return err
}
