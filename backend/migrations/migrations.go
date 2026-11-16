package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 0001_source_collections.sql
var initialSQL string

// Apply is explicit (the API never migrates automatically). The SQL is idempotent.
func Apply(ctx context.Context, db *pgxpool.Pool) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	_, err = conn.Conn().PgConn().Exec(ctx, initialSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply source-collection schema: %w", err)
	}
	return nil
}
