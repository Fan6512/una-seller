package meli

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
)

// OrderSearchFilter searches saved orders across only the caller's current,
// verified account/store/child-seller mappings. It never calls Mercado Libre.
type OrderSearchFilter struct {
	Store    string
	Site     string
	Logistic string
	Status   string
	Purchase string
	Query    string
	Offset   int
}

var orderSearchLogisticPattern = regexp.MustCompile(`^[a-z_]{1,40}$`)

func (f OrderSearchFilter) Valid() bool {
	if f.Store != "all" && !validUUID(f.Store) {
		return false
	}
	if f.Site != "*" && (f.Site == "CBT" || !orderSitePattern.MatchString(f.Site)) {
		return false
	}
	if f.Logistic != "*" && !orderSearchLogisticPattern.MatchString(f.Logistic) {
		return false
	}
	if !ValidOrderStatus(f.Status) || !ValidPurchaseLinkFilter(f.Purchase) ||
		f.Offset < 0 || f.Offset > 100000 || strings.TrimSpace(f.Query) != f.Query ||
		utf8.RuneCountInString(f.Query) > 120 {
		return false
	}
	if f.Query != "" && orderNumber(f.Query) && len(f.Query) < 6 {
		// An incomplete numeric Order ID must not look like a successful empty search.
		return false
	}
	if f.Query != "" && !orderNumber(f.Query) && utf8.RuneCountInString(f.Query) < 2 {
		return false
	}
	return true
}

type OrderSearchRecord struct {
	OrderRecord
	StoreID      string `json:"store_id"`
	SiteID       string `json:"site_id"`
	LogisticType string `json:"logistic_type"`
}

type OrderSearchPage struct {
	Items          []OrderSearchRecord `json:"items"`
	HasMore        bool                `json:"has_more"`
	PurchaseFilter string              `json:"purchase_filter"`
	Query          string              `json:"query"`
	SiteFilter     string              `json:"site_filter"`
	StoreFilter    string              `json:"store_filter"`
}

// At higher order volumes, profile global ORDER BY and JSONB title search with
// EXPLAIN ANALYZE before adding indexes. No schema changes belong in #55.
const globalOrderSearchSQL = `SELECT ` + orderColumns + `,
  o.store_id::text,o.site_id,o.logistic_type
  FROM app.meli_orders o
  JOIN app.meli_store_connections c ON c.id=o.store_id
    AND c.owner_user_id=$1::uuid AND c.status<>'disconnected'
    AND c.marketplaces_status='verified'
    AND EXISTS (
      SELECT 1 FROM jsonb_array_elements(c.marketplaces) AS m(value)
      WHERE m.value->>'site_id'=o.site_id
        AND m.value->>'logistic_type'=o.logistic_type
        AND m.value->>'user_id'=o.seller_id
    )
  LEFT JOIN app.meli_order_purchases p
    ON p.store_id=o.store_id AND p.site_id=o.site_id AND p.logistic_type=o.logistic_type
    AND p.seller_id=o.seller_id AND p.order_id=o.order_id
  LEFT JOIN app.meli_order_dispatch_sla sla
    ON sla.store_id=o.store_id AND sla.site_id=o.site_id AND sla.logistic_type=o.logistic_type
    AND sla.seller_id=o.seller_id AND sla.order_id=o.order_id
    AND sla.shipment_id=o.shipment_id
  WHERE ($2::uuid IS NULL OR o.store_id=$2::uuid)
    AND ($3='*' OR o.site_id=$3)
    AND ($4='*' OR o.logistic_type=$4)
    AND ($5='' OR o.status=$5)
    AND ($6='' OR ($6='linked' AND p.purchase_order_id IS NOT NULL)
      OR ($6='unlinked' AND p.purchase_order_id IS NULL))
    AND ($7='' OR ($8='id' AND o.order_id=$7)
      OR ($8='title' AND EXISTS (
        SELECT 1 FROM jsonb_array_elements(o.order_items) AS item
        WHERE strpos(lower(item->>'title'),lower($7))>0
      )))
  ORDER BY o.date_created DESC,o.order_id DESC,o.store_id,o.site_id,o.logistic_type
  LIMIT $9 OFFSET $10`

func (s *Service) SearchOrders(ctx context.Context, owner string, f OrderSearchFilter) (OrderSearchPage, error) {
	if !validUUID(owner) || !f.Valid() {
		return OrderSearchPage{}, ErrOrderInvalidResponse
	}
	var storeID any
	if f.Store != "all" {
		storeID = f.Store
	}
	kind := ""
	if f.Query != "" {
		kind = "title"
		if orderNumber(f.Query) {
			kind = "id"
		}
	}
	rows, err := s.db.Query(ctx, globalOrderSearchSQL, owner, storeID, f.Site, f.Logistic,
		f.Status, f.Purchase, f.Query, kind, orderPageSize+1, f.Offset)
	if err != nil {
		return OrderSearchPage{}, err
	}
	defer rows.Close()
	page := OrderSearchPage{Items: []OrderSearchRecord{}, PurchaseFilter: f.Purchase,
		Query: f.Query, SiteFilter: f.Site, StoreFilter: f.Store}
	for rows.Next() {
		var row OrderSearchRecord
		row.OrderRecord, err = scanOrder(rows, &row.StoreID, &row.SiteID, &row.LogisticType)
		if err != nil {
			return OrderSearchPage{}, err
		}
		page.Items = append(page.Items, row)
	}
	if err = rows.Err(); err != nil {
		return OrderSearchPage{}, err
	}
	if len(page.Items) > orderPageSize {
		page.HasMore = true
		page.Items = page.Items[:orderPageSize]
	}
	return page, nil
}
