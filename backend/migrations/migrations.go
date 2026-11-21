package migrations

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed 0001_source_collections.sql
var initialSQL string

//go:embed 0002_meli_store_connections.sql
var meliOauthSQL string

//go:embed 0003_published_listings.sql
var publishedListingsSQL string

//go:embed 0004_meli_orders.sql
var ordersSQL string

//go:embed 0005_meli_order_purchases.sql
var orderPurchasesSQL string

//go:embed 0006_order_dispatch_sla.sql
var orderDispatchSLASQL string

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
	_, err = conn.Conn().PgConn().Exec(ctx, meliOauthSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply Mercado Libre OAuth schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, publishedListingsSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply published listings schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, ordersSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply orders schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, orderPurchasesSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply order purchases schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, orderDispatchSLASQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply order dispatch SLA schema: %w", err)
	}
	return nil
}
