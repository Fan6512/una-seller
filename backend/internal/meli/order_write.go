package meli

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgconn"
)

// orderWriter is the local transaction boundary shared by bulk sync and the
// fenced notification worker. It cannot commit or issue upstream requests.
type orderWriter interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

// upsertOrderRecord requires a verified, locked store mapping from its caller.
// No sync cursor, procurement, SLA or billing writes occur here.
func upsertOrderRecord(ctx context.Context, tx orderWriter, store string, scope OrderScope, childSeller string, rec OrderRecord) error {
	items, err := json.Marshal(rec.Items)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO app.meli_orders
(store_id,site_id,logistic_type,seller_id,order_id,pack_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,order_items)
VALUES($1::uuid,$2,$3,$4,$5,NULLIF($6,''),$7,$8,$9::numeric,$10,$11,NULLIF($12,''),$13::jsonb)
ON CONFLICT (store_id,site_id,logistic_type,seller_id,order_id) DO UPDATE
SET pack_id=EXCLUDED.pack_id,status=EXCLUDED.status,currency_id=EXCLUDED.currency_id,
total_amount=EXCLUDED.total_amount,date_created=EXCLUDED.date_created,
last_updated=EXCLUDED.last_updated,shipment_id=EXCLUDED.shipment_id,
shipment_snapshot=CASE WHEN meli_orders.shipment_id=EXCLUDED.shipment_id THEN meli_orders.shipment_snapshot END,
order_items=`+orderItemsWithSavedImages("EXCLUDED.order_items", "meli_orders.order_items")+`,updated_at=now()
WHERE app.meli_orders.last_updated<=EXCLUDED.last_updated`,
		store, scope.Site, scope.Logistic, childSeller, rec.ID, rec.PackID, rec.Status,
		rec.CurrencyID, rec.TotalAmount, rec.DateCreated, rec.LastUpdated, rec.ShipmentID, string(items))
	return err
}
