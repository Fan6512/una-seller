package collections

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

type OwnProduct struct {
	ID              string        `json:"id"`
	SourceProductID *string       `json:"source_product_id"`
	SourcePlatform  string        `json:"source_platform"`
	SourceSite      string        `json:"source_site"`
	SourceURL       string        `json:"source_url"`
	SourceTitle     string        `json:"source_title"`
	Title           string        `json:"title"`
	Description     string        `json:"description"`
	Notes           string        `json:"notes"`
	Price           *string       `json:"price"`
	Currency        *string       `json:"currency"`
	Category        *string       `json:"category"`
	BulletPoints    []string      `json:"bullet_points"`
	Images          []Image       `json:"images"`
	Variants        []Variant     `json:"variants"`
	ListingFields   ListingFields `json:"listing_fields"`
	Revision        int64         `json:"revision"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type OwnProductSummary struct {
	ID              string    `json:"id"`
	SourceProductID *string   `json:"source_product_id"`
	Title           string    `json:"title"`
	SourcePlatform  string    `json:"source_platform"`
	SourceSite      string    `json:"source_site"`
	Price           *string   `json:"price"`
	Currency        *string   `json:"currency"`
	MainImageURL    *string   `json:"main_image_url"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type OwnProductEdit struct {
	Title            string         `json:"title"`
	Description      string         `json:"description"`
	Notes            string         `json:"notes"`
	ListingFields    *ListingFields `json:"listing_fields,omitempty"`
	ExpectedRevision int64          `json:"expected_revision"`
}

func (in OwnProductEdit) Validate() error {
	if strings.TrimSpace(in.Title) == "" || utf8.RuneCountInString(in.Title) > 500 ||
		utf8.RuneCountInString(in.Description) > 20000 || utf8.RuneCountInString(in.Notes) > 5000 ||
		in.ExpectedRevision < 1 {
		return errors.New("invalid product fields")
	}
	if in.ListingFields != nil {
		return in.ListingFields.Validate()
	}
	return nil
}

func (s *Store) GetOwnProduct(ctx context.Context, owner, id string) (OwnProduct, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return OwnProduct{}, ErrNotFound
	}
	var p OwnProduct
	var bullets, images, variants, listingJSON []byte
	err := s.DB.QueryRow(ctx, `SELECT id::text,source_product_id::text,source_platform,source_site,source_url,source_title,
    title,description,notes,price::text,currency,category,bullet_points,images,variants,revision,created_at,updated_at,listing_fields
    FROM app.own_products WHERE id=$1::uuid AND owner_user_id=$2::uuid`, id, owner).
		Scan(&p.ID, &p.SourceProductID, &p.SourcePlatform, &p.SourceSite, &p.SourceURL, &p.SourceTitle,
			&p.Title, &p.Description, &p.Notes, &p.Price, &p.Currency, &p.Category, &bullets, &images, &variants,
			&p.Revision, &p.CreatedAt, &p.UpdatedAt, &listingJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return OwnProduct{}, ErrNotFound
	}
	if err != nil {
		return OwnProduct{}, err
	}
	if err = json.Unmarshal(bullets, &p.BulletPoints); err != nil {
		return OwnProduct{}, err
	}
	if err = json.Unmarshal(images, &p.Images); err != nil {
		return OwnProduct{}, err
	}
	if err = json.Unmarshal(variants, &p.Variants); err != nil {
		return OwnProduct{}, err
	}
	if err = json.Unmarshal(listingJSON, &p.ListingFields); err != nil {
		return OwnProduct{}, err
	}
	return p, nil
}

func (s *Store) ListOwnProducts(ctx context.Context, owner, q string, limit, offset int) ([]OwnProductSummary, error) {
	if !ValidUserID(owner) {
		return nil, ErrNotFound
	}
	if limit < 1 || limit > 100 || offset < 0 || offset > 100000 || len(q) > 200 {
		return nil, fmt.Errorf("invalid paging")
	}
	rows, err := s.DB.Query(ctx, `SELECT id::text,source_product_id::text,title,source_platform,source_site,price::text,
    currency,COALESCE(NULLIF(listing_fields->'images'->0->>'url',''),images->0->>'url'),updated_at FROM app.own_products
    WHERE owner_user_id=$1::uuid AND ($2='' OR title ILIKE '%'||$2||'%')
    ORDER BY updated_at DESC,id DESC LIMIT $3 OFFSET $4`, owner, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []OwnProductSummary{}
	for rows.Next() {
		var item OwnProductSummary
		if err = rows.Scan(&item.ID, &item.SourceProductID, &item.Title, &item.SourcePlatform, &item.SourceSite,
			&item.Price, &item.Currency, &item.MainImageURL, &item.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

// Promote owns one source per own-product for this MVP. Transaction + source
// row lock serializes recaptures and competing approvals. Existing product
// reads are idempotent and never rewrite an already-approved edit.
func (s *Store) Promote(ctx context.Context, owner, sourceID string, expectedRevision int64) (OwnProduct, bool, error) {
	if !ValidUserID(owner) || !ValidUserID(sourceID) {
		return OwnProduct{}, false, ErrNotFound
	}
	if expectedRevision < 0 {
		return OwnProduct{}, false, ErrDraftConflict
	}
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return OwnProduct{}, false, err
	}
	defer tx.Rollback(ctx)
	var platform, site, url, title string
	var price, currency, category, description *string
	var bullets []byte
	err = tx.QueryRow(ctx, `SELECT source_platform,source_site,source_url,title,price::text,currency,category,description,bullet_points
      FROM app.source_products WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`, sourceID, owner).
		Scan(&platform, &site, &url, &title, &price, &currency, &category, &description, &bullets)
	if errors.Is(err, pgx.ErrNoRows) {
		return OwnProduct{}, false, ErrNotFound
	}
	if err != nil {
		return OwnProduct{}, false, err
	}

	var existingID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM app.own_products WHERE source_product_id=$1::uuid AND owner_user_id=$2::uuid`, sourceID, owner).Scan(&existingID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return OwnProduct{}, false, err
	}
	if existingID != "" {
		if err = tx.Commit(ctx); err != nil {
			return OwnProduct{}, false, err
		}
		existing, err := s.GetOwnProduct(ctx, owner, existingID)
		return existing, false, err
	}

	var draftTitle, draftDescription, draftNotes string
	var revision int64
	listingPayload := []byte("{}")
	err = tx.QueryRow(ctx, `SELECT title,description,notes,revision,listing_fields FROM app.collection_working_drafts
      WHERE source_product_id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`, sourceID, owner).
		Scan(&draftTitle, &draftDescription, &draftNotes, &revision, &listingPayload)
	if errors.Is(err, pgx.ErrNoRows) {
		revision = 0
		err = nil
	}
	if err != nil {
		return OwnProduct{}, false, err
	}
	if revision != expectedRevision {
		return OwnProduct{}, false, ErrDraftConflict
	}
	sourceTitle := title
	if revision > 0 {
		title = draftTitle
		description = &draftDescription
	}
	if description == nil {
		empty := ""
		description = &empty
	}
	if strings.TrimSpace(title) == "" || utf8.RuneCountInString(title) > 500 {
		return OwnProduct{}, false, errors.New("invalid source title")
	}

	var images, variants []byte
	err = tx.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object('url',image_url,'type',image_type)
      ORDER BY CASE WHEN image_type='main' THEN 0 ELSE 1 END,image_order,id),'[]'::jsonb)
      FROM app.source_product_images WHERE source_product_id=$1::uuid AND variant_id IS NULL`, sourceID).Scan(&images)
	if err != nil {
		return OwnProduct{}, false, err
	}
	err = tx.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object(
      'attributes',v.attributes,'price',v.price::text,'currency',v.currency,
      'image_urls',COALESCE((SELECT jsonb_agg(i.image_url ORDER BY i.image_order,i.id)
       FROM app.source_product_images i WHERE i.variant_id=v.id),'[]'::jsonb)
     ) ORDER BY v.sort_order,v.id),'[]'::jsonb)
     FROM app.source_product_variants v WHERE v.source_product_id=$1::uuid`, sourceID).Scan(&variants)
	if err != nil {
		return OwnProduct{}, false, err
	}

	var productID string
	err = tx.QueryRow(ctx, `INSERT INTO app.own_products
    (owner_user_id,source_product_id,source_platform,source_site,source_url,source_title,
     title,description,notes,price,currency,category,bullet_points,images,variants,listing_fields)
    VALUES($1::uuid,$2::uuid,$3,$4,$5,$6,$7,$8,$9,$10::numeric,$11,$12,$13::jsonb,$14::jsonb,$15::jsonb,$16::jsonb)
    RETURNING id::text`, owner, sourceID, platform, site, url, sourceTitle, title, *description, draftNotes,
		price, currency, category, string(bullets), string(images), string(variants), string(listingPayload)).Scan(&productID)
	if err != nil {
		return OwnProduct{}, false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return OwnProduct{}, false, err
	}
	result, err := s.GetOwnProduct(ctx, owner, productID)
	if err != nil {
		return OwnProduct{ID: productID}, true, fmt.Errorf("%w: %v", ErrReadback, err)
	}
	return result, true, nil
}

func (s *Store) UpdateOwnProduct(ctx context.Context, owner, id string, in OwnProductEdit) (OwnProduct, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return OwnProduct{}, ErrNotFound
	}
	if err := in.Validate(); err != nil {
		return OwnProduct{}, err
	}
	var listingValue any
	if in.ListingFields != nil {
		raw, err := json.Marshal(in.ListingFields)
		if err != nil {
			return OwnProduct{}, err
		}
		listingValue = string(raw)
	}
	var revision int64
	err := s.DB.QueryRow(ctx, `UPDATE app.own_products
    SET title=$3,description=$4,notes=$5,
      listing_fields=COALESCE($7::jsonb,listing_fields),
      revision=revision+1,updated_at=now()
    WHERE id=$1::uuid AND owner_user_id=$2::uuid AND revision=$6 RETURNING revision`,
		id, owner, in.Title, in.Description, in.Notes, in.ExpectedRevision, listingValue).Scan(&revision)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		e := s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app.own_products WHERE id=$1::uuid AND owner_user_id=$2::uuid)`,
			id, owner).Scan(&exists)
		if e != nil {
			return OwnProduct{}, e
		}
		if exists {
			return OwnProduct{}, ErrDraftConflict
		}
		return OwnProduct{}, ErrNotFound
	}
	if err != nil {
		return OwnProduct{}, err
	}
	p, err := s.GetOwnProduct(ctx, owner, id)
	if err != nil {
		return OwnProduct{ID: id, Revision: revision}, fmt.Errorf("%w: %v", ErrReadback, err)
	}
	return p, nil
}
