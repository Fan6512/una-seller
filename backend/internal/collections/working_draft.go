package collections

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

type WorkingDraftInput struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Notes       string `json:"notes"`
}

type WorkingDraft struct {
	SourceProductID string    `json:"source_product_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Notes           string    `json:"notes"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (in WorkingDraftInput) Validate() error {
	if strings.TrimSpace(in.Title) == "" || utf8.RuneCountInString(in.Title) > 500 ||
		utf8.RuneCountInString(in.Description) > 20000 || utf8.RuneCountInString(in.Notes) > 5000 {
		return errors.New("invalid working draft fields")
	}
	return nil
}

// GetWorkingDraft distinguishes an unstarted draft (nil) from a missing or foreign source.
func (s *Store) GetWorkingDraft(ctx context.Context, owner, id string) (*WorkingDraft, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return nil, ErrNotFound
	}
	var title, description, notes sql.NullString
	var updated sql.NullTime
	err := s.DB.QueryRow(ctx, `SELECT d.title,d.description,d.notes,d.updated_at
		FROM app.source_products p
		LEFT JOIN app.collection_working_drafts d ON d.source_product_id=p.id AND d.owner_user_id=p.owner_user_id
		WHERE p.id=$1::uuid AND p.owner_user_id=$2::uuid`, id, owner).
		Scan(&title, &description, &notes, &updated)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !updated.Valid {
		return nil, nil
	}
	return &WorkingDraft{SourceProductID: id, Title: title.String, Description: description.String,
		Notes: notes.String, UpdatedAt: updated.Time}, nil
}

// SaveWorkingDraft is an owner-scoped upsert; browser recapture cannot change this copy.
func (s *Store) SaveWorkingDraft(ctx context.Context, owner, id string, in WorkingDraftInput) (WorkingDraft, error) {
	if !ValidUserID(owner) || !ValidUserID(id) {
		return WorkingDraft{}, ErrNotFound
	}
	if err := in.Validate(); err != nil {
		return WorkingDraft{}, err
	}
	var draft WorkingDraft
	err := s.DB.QueryRow(ctx, `INSERT INTO app.collection_working_drafts
		(source_product_id,owner_user_id,title,description,notes)
		SELECT p.id,p.owner_user_id,$3,$4,$5 FROM app.source_products p
		WHERE p.id=$2::uuid AND p.owner_user_id=$1::uuid
		ON CONFLICT (source_product_id) DO UPDATE SET
			title=EXCLUDED.title,description=EXCLUDED.description,
			notes=EXCLUDED.notes,updated_at=now()
		WHERE app.collection_working_drafts.owner_user_id=EXCLUDED.owner_user_id
		RETURNING source_product_id::text,title,description,notes,updated_at`,
		owner, id, in.Title, in.Description, in.Notes).
		Scan(&draft.SourceProductID, &draft.Title, &draft.Description, &draft.Notes, &draft.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return WorkingDraft{}, ErrNotFound
	}
	return draft, err
}
