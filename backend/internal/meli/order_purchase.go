package meli

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var (
	ErrOrderPurchaseInvalid  = errors.New("invalid order purchase details")
	ErrOrderPurchaseConflict = errors.New("order purchase version changed")
)

var purchaseIDPattern = regexp.MustCompile("^[0-9]{8,40}$")
var purchaseTrackingPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_/-]{2,79}package meli

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var (
	ErrOrderPurchaseInvalid  = errors.New("invalid order purchase details")
	ErrOrderPurchaseConflict = errors.New("order purchase version changed")
)

)
var purchaseCostPattern = regexp.MustCompile("^(0|[1-9][0-9]{0,11})(\\.[0-9]{1,2})?$")

// OrderPurchase is a local procurement note. It never changes a Mercado Libre order.
type OrderPurchase struct {
	OrderID         string    `json:"order_id"`
	PurchaseOrderID string    `json:"purchase_order_id"`
	CostCNY         *string   `json:"cost_cny"`
	Source          string    `json:"source"`
	WarehouseAddress string `json:"warehouse_address"`
	PurchaseTrackingNumber string `json:"purchase_tracking_number"`
	UpdatedAt       time.Time `json:"updated_at"`
}

const purchaseSelect = `WITH connection AS (` + orderConnection + `)
 SELECT o.order_id, p.purchase_order_id,
 CASE WHEN p.cost_cny IS NULL THEN NULL ELSE trim_scale(p.cost_cny)::text END,
 p.cost_source,p.updated_at,
 COALESCE(p.warehouse_address,''),COALESCE(p.purchase_tracking_number,'')
 FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 LEFT JOIN app.meli_order_purchases p ON
 (p.store_id,p.site_id,p.logistic_type,p.seller_id,p.order_id)=
 (o.store_id,o.site_id,o.logistic_type,o.seller_id,o.order_id)
 WHERE c.status<>'disconnected' AND o.site_id=$3 AND o.logistic_type=$4 AND o.order_id=$5`

// GetOrderPurchase returns nil when an existing cached order has no procurement link.
// It uses the *current* verified child-seller mapping in the same read statement.
func (s *Service) GetOrderPurchase(ctx context.Context, owner, store string, scope OrderScope, id string) (*OrderPurchase, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !orderNumber(id) {
		return nil, ErrOrderNotFound
	}
	var orderID string
	var purchaseID, amount, source sql.NullString
	var updated sql.NullTime
	var warehouse, tracking string
	err := s.db.QueryRow(ctx, purchaseSelect, store, owner, scope.Site, scope.Logistic, id).
		Scan(&orderID, &purchaseID, &amount, &source, &updated, &warehouse, &tracking)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if !purchaseID.Valid {
		return nil, nil
	}
	record := &OrderPurchase{OrderID: orderID, PurchaseOrderID: purchaseID.String, Source: source.String, UpdatedAt: updated.Time, WarehouseAddress: warehouse, PurchaseTrackingNumber: tracking}
	if amount.Valid {
		record.CostCNY = &amount.String
	}
	return record, nil
}

// SaveOrderPurchase accepts manually verified 1688 order details only.
// The service never receives a client-supplied seller/owner ID.
func (s *Service) SaveOrderPurchase(ctx context.Context, owner, store string, scope OrderScope,
	id, purchaseID, costCNY string, expected *time.Time) (*OrderPurchase, error) {
	return s.SaveOrderPurchaseWithShipping(ctx, owner, store, scope, id, purchaseID, costCNY, "", "", expected)
}

// SaveOrderPurchaseWithShipping saves manual forwarding notes, without shipping anything.
func (s *Service) SaveOrderPurchaseWithShipping(ctx context.Context, owner, store string, scope OrderScope,
	id, purchaseID, costCNY, warehouse, tracking string, expected *time.Time) (*OrderPurchase, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !orderNumber(id) {
		return nil, ErrOrderNotFound
	}
	if !purchaseIDPattern.MatchString(purchaseID) || (costCNY != "" && !purchaseCostPattern.MatchString(costCNY)) ||
		!utf8.ValidString(warehouse) || utf8.RuneCountInString(warehouse) > 300 ||
		(warehouse != "" && (strings.TrimSpace(warehouse) != warehouse || strings.ContainsAny(warehouse, "\r\n\x00"))) ||
		(tracking != "" && !purchaseTrackingPattern.MatchString(tracking)) {
		return nil, ErrOrderPurchaseInvalid
	}
	var record OrderPurchase
	var amount sql.NullString
	err := pgx.BeginTxFunc(ctx, s.db, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		// Lock the mapping and order before reading the purchase. The order also
		// serializes concurrent first writes, when there is no purchase row yet.
		var seller string
		err := tx.QueryRow(ctx, `WITH connection AS (`+orderConnection+` FOR SHARE OF c)
 SELECT o.seller_id FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 WHERE c.status<>'disconnected' AND o.site_id=$3 AND o.logistic_type=$4 AND o.order_id=$5
 FOR NO KEY UPDATE OF o`, store, owner, scope.Site, scope.Logistic, id).Scan(&seller)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx, `WITH supplied AS (
 SELECT $6::text AS purchase_order_id,NULLIF($7,'')::numeric AS cost_cny,
 $9::text AS warehouse_address,$10::text AS purchase_tracking_number
 ), existing AS (
 SELECT p.*,p.purchase_order_id=s.purchase_order_id AND p.cost_cny IS NOT DISTINCT FROM s.cost_cny
   AND p.warehouse_address=s.warehouse_address AND p.purchase_tracking_number=s.purchase_tracking_number
   AND p.cost_source='manual' AS same
 FROM app.meli_order_purchases p CROSS JOIN supplied s
 WHERE (p.store_id,p.site_id,p.logistic_type,p.seller_id,p.order_id)=($1::uuid,$2,$3,$4,$5)
 ), written AS (
 INSERT INTO app.meli_order_purchases AS p
 (store_id,site_id,logistic_type,seller_id,order_id,purchase_order_id,cost_cny,cost_source,warehouse_address,purchase_tracking_number)
 SELECT $1::uuid,$2,$3,$4,$5,s.purchase_order_id,s.cost_cny,'manual',
 s.warehouse_address,s.purchase_tracking_number FROM supplied s
 WHERE ($8::timestamptz IS NULL AND NOT EXISTS(SELECT 1 FROM existing))
    OR EXISTS(SELECT 1 FROM existing WHERE NOT same AND updated_at=$8)
 ON CONFLICT (store_id,site_id,logistic_type,seller_id,order_id) DO UPDATE
 SET purchase_order_id=EXCLUDED.purchase_order_id,cost_cny=EXCLUDED.cost_cny,
 warehouse_address=EXCLUDED.warehouse_address,purchase_tracking_number=EXCLUDED.purchase_tracking_number,
 cost_source='manual',updated_at=GREATEST(clock_timestamp(),p.updated_at+interval '1 microsecond')
 RETURNING order_id,purchase_order_id,cost_cny,cost_source,updated_at,warehouse_address,purchase_tracking_number
 )
 SELECT order_id,purchase_order_id,trim_scale(cost_cny)::text,cost_source,updated_at,warehouse_address,purchase_tracking_number FROM written
 UNION ALL
 SELECT order_id,purchase_order_id,trim_scale(cost_cny)::text,cost_source,updated_at,warehouse_address,purchase_tracking_number FROM existing WHERE same`,
			store, scope.Site, scope.Logistic, seller, id, purchaseID, costCNY, expected, warehouse, tracking).
			Scan(&record.OrderID, &record.PurchaseOrderID, &amount, &record.Source,
				&record.UpdatedAt, &record.WarehouseAddress, &record.PurchaseTrackingNumber)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderPurchaseConflict
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if amount.Valid {
		record.CostCNY = &amount.String
	}
	return &record, nil
}

// ListPurchaseWarehouses returns recent manually entered addresses within the
// verified seller scope. It is not a freight forwarder directory or API.
func (s *Service) ListPurchaseWarehouses(ctx context.Context, owner, store string, scope OrderScope) ([]string, error) {
	if !orderScopeValid(scope) {
		return nil, ErrStoreNotFound
	}
	if err := s.authorizeOrderRead(ctx, owner, store, scope); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `WITH connection AS (`+orderConnection+`)
 SELECT DISTINCT p.warehouse_address FROM app.meli_order_purchases p
 JOIN connection c ON c.id=p.store_id AND c.child_id=p.seller_id
 WHERE c.status<>'disconnected' AND p.site_id=$3 AND p.logistic_type=$4
 AND p.warehouse_address<>'' ORDER BY p.warehouse_address LIMIT 20`,
		store, owner, scope.Site, scope.Logistic)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var address string
		if err := rows.Scan(&address); err != nil {
			return nil, err
		}
		values = append(values, address)
	}
	return values, rows.Err()
}
