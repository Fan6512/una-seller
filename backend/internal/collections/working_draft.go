package collections

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var ErrDraftConflict = errors.New("working draft has been updated")
var ErrDraftConverted = errors.New("source already moved to product box")

type WorkingDraftInput struct {
	Title            string `json:"title"`
	Description      string `json:"description"`
	Notes            string `json:"notes"`
	ListingFields    *ListingFields `json:"listing_fields,omitempty"`
	ExpectedRevision int64 `json:"expected_revision"`
}

type WorkingDraft struct {
	SourceProductID string    `json:"source_product_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Notes           string    `json:"notes"`
	ListingFields   ListingFields `json:"listing_fields"`
	Revision        int64     `json:"revision"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (in WorkingDraftInput) Validate() error {
	if strings.TrimSpace(in.Title) == "" || utf8.RuneCountInString(in.Title) > 500 ||
		utf8.RuneCountInString(in.Description) > 20000 || utf8.RuneCountInString(in.Notes) > 5000 ||
		in.ExpectedRevision < 0 {
		return errors.New("invalid working draft fields")
	}
	if in.ListingFields != nil {
		return in.ListingFields.Validate()
	}
	return nil
}

// A missing draft returns nil, but a source owned by someone else returns ErrNotFound.
func (s *Store) GetWorkingDraft(ctx context.Context, owner, id string) (*WorkingDraft, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return nil, ErrNotFound
	}
	var sourceID string
	var title, description, notes sql.NullString
	var updated sql.NullTime
	var revision sql.NullInt64
	var fields []byte
	err := s.DB.QueryRow(ctx, `SELECT p.id::text,d.title,d.description,d.notes,d.revision,d.updated_at,d.listing_fields
		FROM app.source_products p
		LEFT JOIN app.collection_working_drafts d
			ON d.source_product_id=p.id AND d.owner_user_id=p.owner_user_id
		WHERE p.id=$1::uuid AND p.owner_user_id=$2::uuid`, id, owner).
		Scan(&sourceID, &title, &description, &notes, &revision, &updated, &fields)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !updated.Valid {
		return nil, nil
	}
	var listing ListingFields
	if len(fields) > 0 {
		if err = json.Unmarshal(fields, &listing); err != nil {
			return nil, err
		}
	}
	return &WorkingDraft{SourceProductID: sourceID, Title: title.String, Description: description.String,
		Notes: notes.String, ListingFields: listing, Revision: revision.Int64, UpdatedAt: updated.Time}, nil
}

// SaveWorkingDraft uses revision=0 only for first creation, then CAS revisions 1,2,...
// Stale writes return ErrDraftConflict without changing another window's edits.
func (s *Store) SaveWorkingDraft(ctx context.Context, owner, id string, in WorkingDraftInput) (WorkingDraft, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return WorkingDraft{}, ErrNotFound
	}
	if err := in.Validate(); err != nil {
		return WorkingDraft{}, err
	}
	// Lock the owned source row first, matching Promote's lock order.
	// Conversion and draft writes can therefore never commit out of order.
	tx, err := s.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return WorkingDraft{}, err
	}
	defer tx.Rollback(ctx)
	var sourceID string
	err = tx.QueryRow(ctx, `SELECT id::text FROM app.source_products
		WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`, id, owner).Scan(&sourceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkingDraft{}, ErrNotFound
	}
	if err != nil {
		return WorkingDraft{}, err
	}
	var converted bool
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app.own_products
		WHERE source_product_id=$1::uuid AND owner_user_id=$2::uuid)`, id, owner).Scan(&converted)
	if err != nil {
		return WorkingDraft{}, err
	}
	if converted {
		return WorkingDraft{}, ErrDraftConverted
	}
	var listingPayload any
	if in.ListingFields != nil {
		raw, marshalErr := json.Marshal(in.ListingFields)
		if marshalErr != nil {
			return WorkingDraft{}, marshalErr
		}
		listingPayload = string(raw)
	}
	var draft WorkingDraft
	var savedFields []byte
	if in.ExpectedRevision == 0 {
		err = tx.QueryRow(ctx, `INSERT INTO app.collection_working_drafts
			(source_product_id,owner_user_id,title,description,notes,revision,listing_fields)
			SELECT p.id,p.owner_user_id,$3,$4,$5,1,COALESCE($6::jsonb,'{}'::jsonb) FROM app.source_products p
			WHERE p.id=$2::uuid AND p.owner_user_id=$1::uuid
			ON CONFLICT (source_product_id) DO NOTHING
			RETURNING source_product_id::text,title,description,notes,revision,updated_at,listing_fields`,
			owner, id, in.Title, in.Description, in.Notes, listingPayload).
			Scan(&draft.SourceProductID, &draft.Title, &draft.Description, &draft.Notes,
				&draft.Revision, &draft.UpdatedAt, &savedFields)
	} else {
		err = tx.QueryRow(ctx, `UPDATE app.collection_working_drafts d
			SET title=$3,description=$4,notes=$5,listing_fields=COALESCE($7::jsonb,d.listing_fields),
				revision=d.revision+1,updated_at=now()
			FROM app.source_products p
			WHERE d.source_product_id=p.id AND p.id=$2::uuid AND p.owner_user_id=$1::uuid
				AND d.owner_user_id=p.owner_user_id AND d.revision=$6
			RETURNING d.source_product_id::text,d.title,d.description,d.notes,d.revision,d.updated_at,d.listing_fields`,
			owner, id, in.Title, in.Description, in.Notes, in.ExpectedRevision, listingPayload).
			Scan(&draft.SourceProductID, &draft.Title, &draft.Description, &draft.Notes,
				&draft.Revision, &draft.UpdatedAt, &savedFields)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if readErr := tx.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM app.source_products WHERE id=$1::uuid AND owner_user_id=$2::uuid
		)`, id, owner).Scan(&exists); readErr != nil {
			return WorkingDraft{}, readErr
		}
		if exists {
			return WorkingDraft{}, ErrDraftConflict
		}
		return WorkingDraft{}, ErrNotFound
	}
	if err != nil {
		return WorkingDraft{}, err
	}
	if err = json.Unmarshal(savedFields, &draft.ListingFields); err != nil {
		return WorkingDraft{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return WorkingDraft{}, err
	}
	return draft, nil
}
