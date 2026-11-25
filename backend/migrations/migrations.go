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

//go:embed 0007_collection_working_drafts.sql
var collectionWorkingDraftSQL string

//go:embed 0008_own_products.sql
var ownProductsSQL string

//go:embed 0009_listing_editor_fields.sql
var listingEditorFieldsSQL string

//go:embed 0010_order_shipment_snapshot.sql
var orderShipmentSnapshotSQL string

//go:embed 0011_user_service_credentials.sql
var userServiceCredentialsSQL string

//go:embed 0012_cbt_c0_evidence.sql
var cbtC0EvidenceSQL string

//go:embed 0013_order_billing_snapshot.sql
var orderBillingSnapshotSQL string

//go:embed 0014_meli_notification_inbox.sql
var notificationInboxSQL string

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
	_, err = conn.Conn().PgConn().Exec(ctx, collectionWorkingDraftSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply collection working draft schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, ownProductsSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply own product-box schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, listingEditorFieldsSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply listing editor fields schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, orderShipmentSnapshotSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply order shipment snapshot schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, userServiceCredentialsSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply user service credentials schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, cbtC0EvidenceSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply CBT C0 evidence schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, orderBillingSnapshotSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply order billing snapshot schema: %w", err)
	}
	_, err = conn.Conn().PgConn().Exec(ctx, notificationInboxSQL).ReadAll()
	if err != nil {
		return fmt.Errorf("apply notification inbox schema: %w", err)
	}
	return nil
}
