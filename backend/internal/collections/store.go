package collections

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct{ DB *pgxpool.Pool }

var ErrNotFound = errors.New("collection not found")
var ErrReadback = errors.New("collection committed, but readback failed")

// Save is a single transaction. The unique key and row lock serialize recaptures.
// A partial recapture merges known values only, never deleting previously saved subitems.
func (s *Store) Save(ctx context.Context, owner string, in Input) (Product, bool, error) {
	norm, err := in.Validate()
	if err != nil {
		return Product{}, false, err
	}
	if !ValidUserID(owner) {
		return Product{}, false, errors.New("invalid configured owner")
	}
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Product{}, false, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `INSERT INTO app.users(id) VALUES ($1) ON CONFLICT DO NOTHING`, owner); err != nil {
		return Product{}, false, err
	}
	var id, title string
	var existingStatus string
	err = tx.QueryRow(ctx, `SELECT id::text,title,collection_status FROM app.source_products WHERE owner_user_id=$1 AND source_platform=$2 AND source_site=$3 AND normalized_url=$4 FOR UPDATE`, owner, in.Platform, in.Site, norm).Scan(&id, &title, &existingStatus)
	created := false
	if errors.Is(err, pgx.ErrNoRows) {
		title, err = validateTitle(in, "")
		if err != nil {
			return Product{}, false, err
		}
		err = tx.QueryRow(ctx, `INSERT INTO app.source_products(owner_user_id,source_platform,source_site,source_url,normalized_url,title,collection_status) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING RETURNING id::text`, owner, in.Platform, in.Site, in.SourceURL, norm, title, in.Status).Scan(&id)
		if err == nil {
			created = true
			existingStatus = in.Status
		}
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, `SELECT id::text,title,collection_status FROM app.source_products WHERE owner_user_id=$1 AND source_platform=$2 AND source_site=$3 AND normalized_url=$4 FOR UPDATE`, owner, in.Platform, in.Site, norm).Scan(&id, &title, &existingStatus)
		}
	}
	if err != nil {
		return Product{}, false, err
	}
	title, err = validateTitle(in, title)
	if err != nil {
		return Product{}, false, err
	}
	partial := in.Status == "partial"
	var oldPrice, oldCurrency, oldCategory, oldDescription *string
	var oldBullets []byte
	if err = tx.QueryRow(ctx, `SELECT price::text,currency,category,description,bullet_points FROM app.source_products WHERE id=$1`, id).Scan(&oldPrice, &oldCurrency, &oldCategory, &oldDescription, &oldBullets); err != nil {
		return Product{}, false, err
	}
	price := nullable(in.Price, partial, oldPrice)
	currency := nullable(in.Currency, partial, oldCurrency)
	category := nullable(in.Category, partial, oldCategory)
	description := nullable(in.Description, partial, oldDescription)
	bullets := oldBullets
	if in.BulletPoints.Present && !in.BulletPoints.Null && (!partial || len(in.BulletPoints.Value) > 0) {
		bullets, _ = json.Marshal(in.BulletPoints.Value)
	} else if in.BulletPoints.Present && in.BulletPoints.Null && !partial {
		bullets = []byte(`[]`)
	}
	status := in.Status
	if existingStatus == "complete" && partial {
		status = "complete"
	}
	warnings := append(append([]string{}, in.Warnings...), in.MissingFields...)
	if warnings == nil {
		warnings = []string{}
	}
	warningsJSON, _ := json.Marshal(warnings)
	_, err = tx.Exec(ctx, `UPDATE app.source_products SET source_url=$2,title=$3,price=$4,currency=$5,category=$6,description=$7,bullet_points=$8,collection_status=$9,collection_warnings=$10,last_collected_at=now() WHERE id=$1`, id, in.SourceURL, title, price, currency, category, description, bullets, status, warningsJSON)
	if err != nil {
		return Product{}, false, err
	}
	if in.Images.Present && (!partial || (!in.Images.Null && len(in.Images.Value) > 0)) {
		if !partial {
			if _, err = tx.Exec(ctx, `DELETE FROM app.source_product_images WHERE source_product_id=$1 AND variant_id IS NULL`, id); err != nil {
				return Product{}, false, err
			}
		}
		for i, image := range in.Images.Value {
			if partial {
				var exists bool
				err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app.source_product_images WHERE source_product_id=$1 AND variant_id IS NULL AND image_url=$2)`, id, image.URL).Scan(&exists)
				if err != nil {
					return Product{}, false, err
				}
				if exists {
					continue
				}
			}
			order := i
			if partial {
				if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(image_order),-1)+1 FROM app.source_product_images WHERE source_product_id=$1 AND variant_id IS NULL`, id).Scan(&order); err != nil {
					return Product{}, false, err
				}
			}
			if _, err = tx.Exec(ctx, `INSERT INTO app.source_product_images(source_product_id,image_url,image_type,image_order) VALUES($1,$2,$3,$4)`, id, image.URL, image.Type, order); err != nil {
				return Product{}, false, err
			}
		}
	}
	if in.Variants.Present && (!partial || (!in.Variants.Null && len(in.Variants.Value) > 0)) {
		if !partial {
			if _, err = tx.Exec(ctx, `DELETE FROM app.source_product_variants WHERE source_product_id=$1`, id); err != nil {
				return Product{}, false, err
			}
		}
		for i, v := range in.Variants.Value {
			key, _ := json.Marshal(v.Attributes)
			var variantID string
			var oldVPrice, oldVCurrency *string
			err = tx.QueryRow(ctx, `SELECT id::text,price::text,currency FROM app.source_product_variants WHERE source_product_id=$1 AND variant_key=$2`, id, string(key)).Scan(&variantID, &oldVPrice, &oldVCurrency)
			if errors.Is(err, pgx.ErrNoRows) {
				variantID = ""
			} else if err != nil {
				return Product{}, false, err
			}
			vPrice := nullable(v.Price, partial, oldVPrice)
			vCurrency := nullable(v.Currency, partial, oldVCurrency)
			if variantID == "" {
				order := i
				if partial {
					if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order),-1)+1 FROM app.source_product_variants WHERE source_product_id=$1`, id).Scan(&order); err != nil {
						return Product{}, false, err
					}
				}
				err = tx.QueryRow(ctx, `INSERT INTO app.source_product_variants(source_product_id,variant_key,attributes,price,currency,sort_order) VALUES($1,$2,$3,$4,$5,$6) RETURNING id::text`, id, string(key), key, vPrice, vCurrency, order).Scan(&variantID)
			} else {
				_, err = tx.Exec(ctx, `UPDATE app.source_product_variants SET price=$2,currency=$3 WHERE id=$1`, variantID, vPrice, vCurrency)
			}
			if err != nil {
				return Product{}, false, err
			}
			if v.ImageURLs.Present && (!partial || (!v.ImageURLs.Null && len(v.ImageURLs.Value) > 0)) {
				if !partial {
					if _, err = tx.Exec(ctx, `DELETE FROM app.source_product_images WHERE variant_id=$1`, variantID); err != nil {
						return Product{}, false, err
					}
				}
				for j, imageURL := range v.ImageURLs.Value {
					if partial {
						var exists bool
						if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app.source_product_images WHERE variant_id=$1 AND image_url=$2)`, variantID, imageURL).Scan(&exists); err != nil {
							return Product{}, false, err
						}
						if exists {
							continue
						}
					}
					order := j
					if partial {
						if err = tx.QueryRow(ctx, `SELECT COALESCE(MAX(image_order),-1)+1 FROM app.source_product_images WHERE variant_id=$1`, variantID).Scan(&order); err != nil {
							return Product{}, false, err
						}
					}
					if _, err = tx.Exec(ctx, `INSERT INTO app.source_product_images(source_product_id,variant_id,image_url,image_type,image_order) VALUES($1,$2,$3,'variant',$4)`, id, variantID, imageURL, order); err != nil {
						return Product{}, false, err
					}
				}
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Product{}, false, err
	}
	result, err := s.Get(ctx, owner, id)
	if err != nil {
		return Product{ID: id, Status: status}, created, fmt.Errorf("%w: %v", ErrReadback, err)
	}
	return result, created, nil
}

func (s *Store) Get(ctx context.Context, owner, id string) (Product, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return Product{}, ErrNotFound
	}
	p := Product{Images: []Image{}, Variants: []Variant{}, BulletPoints: []string{}, Warnings: []string{}}
	var rawBullets, rawWarnings []byte
	var first, last time.Time
	err := s.DB.QueryRow(ctx, `SELECT id::text,source_platform,source_site,source_url,normalized_url,title,price::text,currency,category,description,bullet_points,collection_status,collection_warnings,first_collected_at,last_collected_at FROM app.source_products WHERE id=$1 AND owner_user_id=$2`, id, owner).Scan(&p.ID, &p.Platform, &p.Site, &p.SourceURL, &p.NormalizedURL, &p.Title, &p.Price, &p.Currency, &p.Category, &p.Description, &rawBullets, &p.Status, &rawWarnings, &first, &last)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, err
	}
	p.FirstCollectedAt, p.LastCollectedAt = first.UTC().Format(time.RFC3339Nano), last.UTC().Format(time.RFC3339Nano)
	if err = json.Unmarshal(rawBullets, &p.BulletPoints); err != nil {
		return Product{}, err
	}
	if err = json.Unmarshal(rawWarnings, &p.Warnings); err != nil {
		return Product{}, err
	}
	rows, err := s.DB.Query(ctx, `SELECT image_url,image_type FROM app.source_product_images WHERE source_product_id=$1 AND variant_id IS NULL ORDER BY image_order,id`, id)
	if err != nil {
		return Product{}, err
	}
	for rows.Next() {
		var im Image
		if err = rows.Scan(&im.URL, &im.Type); err != nil {
			break
		}
		p.Images = append(p.Images, im)
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return Product{}, err
	}
	variants, err := s.DB.Query(ctx, `SELECT id::text,attributes,price::text,currency FROM app.source_product_variants WHERE source_product_id=$1 ORDER BY sort_order,id`, id)
	if err != nil {
		return Product{}, err
	}
	type variantRow struct {
		id              string
		attrs           []byte
		price, currency *string
	}
	var buffer []variantRow
	for variants.Next() {
		var v variantRow
		if err = variants.Scan(&v.id, &v.attrs, &v.price, &v.currency); err != nil {
			break
		}
		buffer = append(buffer, v)
	}
	if err == nil {
		err = variants.Err()
	}
	variants.Close()
	if err != nil {
		return Product{}, err
	}
	for _, v := range buffer {
		item := Variant{Price: v.price, Currency: v.currency, ImageURLs: []string{}}
		if err = json.Unmarshal(v.attrs, &item.Attributes); err != nil {
			return Product{}, err
		}
		irows, e := s.DB.Query(ctx, `SELECT image_url FROM app.source_product_images WHERE variant_id=$1 ORDER BY image_order,id`, v.id)
		if e != nil {
			return Product{}, e
		}
		for irows.Next() {
			var u string
			if e = irows.Scan(&u); e != nil {
				break
			}
			item.ImageURLs = append(item.ImageURLs, u)
		}
		if e == nil {
			e = irows.Err()
		}
		irows.Close()
		if e != nil {
			return Product{}, e
		}
		p.Variants = append(p.Variants, item)
	}
	return p, nil
}

func (s *Store) List(ctx context.Context, owner, query string, limit, offset int) ([]Summary, error) {
	if !ValidUserID(owner) {
		return nil, errors.New("invalid configured owner")
	}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 {
		return nil, fmt.Errorf("invalid paging")
	}
	rows, err := s.DB.Query(ctx, `SELECT p.id::text,p.title,p.source_platform,p.source_site,p.price::text,p.currency,(SELECT i.image_url FROM app.source_product_images i WHERE i.source_product_id=p.id AND i.variant_id IS NULL ORDER BY CASE WHEN i.image_type='main' THEN 0 ELSE 1 END,i.image_order LIMIT 1),p.collection_status,p.last_collected_at FROM app.source_products p WHERE p.owner_user_id=$1 AND ($2='' OR p.title ILIKE '%'||$2||'%') ORDER BY p.last_collected_at DESC,p.id LIMIT $3 OFFSET $4`, owner, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []Summary{}
	for rows.Next() {
		var x Summary
		var t time.Time
		if err = rows.Scan(&x.ID, &x.Title, &x.Platform, &x.Site, &x.Price, &x.Currency, &x.MainImageURL, &x.Status, &t); err != nil {
			return nil, err
		}
		x.LastCollectedAt = t.UTC().Format(time.RFC3339Nano)
		result = append(result, x)
	}
	return result, rows.Err()
}

// Open uses one small pool; the persistent Go backend does not use Supabase Data API.
func Open(ctx context.Context, dsn string) (*Store, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("DATABASE_URL is empty")
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 3
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect database: %w", err)
	}
	return &Store{DB: pool}, nil
}
