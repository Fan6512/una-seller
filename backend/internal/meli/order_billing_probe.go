package meli

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// BillingProbeReport is an operator-only, redacted diagnostic. It is not a
// confirmed expected-net source, and a local singleton pack is not proof of
// the official pack or shipment membership.
type BillingProbeReport struct {
	Site               string                  `json:"site"`
	Logistic           string                  `json:"logistic"`
	OrderStatus        string                  `json:"order_status"`
	MerchandiseUSD     string                  `json:"merchandise_usd,omitempty"`
	Available          bool                    `json:"billing_available"`
	Payments           []billingProbePayment   `json:"payments"`
	Charges            []billingProbeCharge    `json:"charges"`
	Taxes              []BillingTax            `json:"taxes"`
	Shipping           *billingProbeShipping   `json:"shipping,omitempty"`
	OfficialMembership string                  `json:"official_membership"`
	PackMembership     billingProbeMembers     `json:"pack_membership"`
	ShipmentMembership billingProbeMembers     `json:"shipment_membership"`
	BillingEstimate    *BillingEstimatedUSDNet `json:"billing_estimate,omitempty"`
}

type billingProbeMembers struct {
	HTTPStatus int  `json:"http_status"`
	Rows       int  `json:"rows"`
	Matches    bool `json:"single_order_matches"`
}

type billingProbePayment struct {
	Status             string `json:"status"`
	MoneyReleaseStatus string `json:"money_release_status"`
	HasPaymentID       bool   `json:"has_payment_id"`
}

type billingProbeCharge struct {
	Type                 string `json:"type"`
	Subtype              string `json:"subtype"`
	Currency             string `json:"currency"`
	Amount               string `json:"amount"`
	DebitedFromOperation string `json:"debited_from_operation"`
	Status               string `json:"status"`
	HasDetailID          bool   `json:"has_detail_id"`
	HasReversal          bool   `json:"has_reversal"`
	SalesRefCount        int    `json:"sales_ref_count"`
	AllRefsMatchOrder    bool   `json:"all_refs_match_order"`
	HasOperationID       bool   `json:"has_operation_id"`
	PackRef              string `json:"pack_ref"`
}

type billingProbeShipping struct {
	BuyerCostUSD     string `json:"buyer_cost_usd,omitempty"`
	SellerChargeUSD  string `json:"seller_charge_usd,omitempty"`
	SellerNetCostUSD string `json:"seller_net_cost_usd,omitempty"`
}

func redactBillingProbe(scope OrderScope, orderID, packID, status string, lines []OrderLine, item BillingOrderPreview) BillingProbeReport {
	out := BillingProbeReport{
		Site: scope.Site, Logistic: scope.Logistic, OrderStatus: status,
		MerchandiseUSD: orderSalesAmountUSD(lines), Available: item.Available,
		Payments: []billingProbePayment{}, Charges: []billingProbeCharge{},
		Taxes: item.Taxes, OfficialMembership: "not_verified",
	}
	for _, p := range item.Payments {
		out.Payments = append(out.Payments, billingProbePayment{
			Status: p.Status, MoneyReleaseStatus: p.MoneyReleaseStatus, HasPaymentID: p.PaymentID != "",
		})
	}
	for _, c := range item.Charges {
		x := billingProbeCharge{
			Type: c.Type, Subtype: c.Subtype, Currency: c.Currency, Amount: c.Amount,
			DebitedFromOperation: c.DebitedFromOperation, Status: c.Status,
			HasDetailID: c.DetailID != "", HasReversal: c.ReversalOf != "",
			SalesRefCount: len(c.SalesRefs), AllRefsMatchOrder: len(c.SalesRefs) > 0,
			PackRef: "missing",
		}
		if c.PackID != "" {
			x.PackRef = "different"
			if c.PackID == packID {
				x.PackRef = "matches"
			}
		}
		for _, ref := range c.SalesRefs {
			x.AllRefsMatchOrder = x.AllRefsMatchOrder && ref.OrderID == orderID
			x.HasOperationID = x.HasOperationID || ref.OperationID != ""
		}
		out.Charges = append(out.Charges, x)
	}
	if item.Shipping != nil {
		out.Shipping = &billingProbeShipping{
			BuyerCostUSD: item.Shipping.BuyerCostUSD, SellerChargeUSD: item.Shipping.SellerChargeUSD,
			SellerNetCostUSD: item.Shipping.SellerNetCostUSD,
		}
	}
	return out
}

// ProbeBillingByPack resolves a single locally known pack order in the
// explicit owner/store/site/logistics scope, then uses the normal authorized
// Billing read without updating the financial snapshot. OAuth refresh may
// still use the service's existing credential persistence path.
func (s *Service) ProbeBillingByPack(ctx context.Context, owner, store string, scope OrderScope, packID string) (BillingProbeReport, error) {
	return s.probeBillingByPack(ctx, owner, store, scope, packID, nil)
}

// A collector can supply a sanitized batch item instead of re-fetching Billing.
func (s *Service) probeBillingByPack(ctx context.Context, owner, store string, scope OrderScope, packID string, preloaded *BillingOrderPreview) (BillingProbeReport, error) {
	if !validUUID(owner) || !validUUID(store) || !orderScopeValid(scope) || !billingIDsValid([]string{packID}) {
		return BillingProbeReport{}, ErrOrderBillingInvalid
	}
	type scopedOrder struct {
		id, status, shipment, child, items string
		localShipmentOrders               int
	}
	lookup := func() (scopedOrder, error) {
		rows, err := s.db.Query(ctx, `WITH connection AS (`+orderConnection+`)
SELECT o.order_id, o.status, COALESCE(o.shipment_id,''), o.seller_id, o.order_items::text,
 (SELECT count(*) FROM app.meli_orders other WHERE other.store_id=o.store_id
 AND other.shipment_id=o.shipment_id)
FROM app.meli_orders o
JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
WHERE c.status='connected' AND o.site_id=$3 AND o.logistic_type=$4 AND o.pack_id=$5
LIMIT 2`, store, owner, scope.Site, scope.Logistic, packID)
		if err != nil {
			return scopedOrder{}, err
		}
		defer rows.Close()
		var order scopedOrder
		count := 0
		for rows.Next() {
			count++
			if err := rows.Scan(&order.id, &order.status, &order.shipment, &order.child, &order.items, &order.localShipmentOrders); err != nil {
				return scopedOrder{}, err
			}
		}
		if err := rows.Err(); err != nil {
			return scopedOrder{}, err
		}
		if count != 1 || !billingIDsValid([]string{order.id}) {
			return scopedOrder{}, ErrOrderNotFound
		}
		return order, nil
	}
	order, err := lookup()
	if err != nil {
		return BillingProbeReport{}, err
	}
	item := BillingOrderPreview{}
	if preloaded == nil {
		page, err := s.readOrderBilling(ctx, owner, store, scope, []string{order.id}, false)
		if err != nil {
			return BillingProbeReport{}, err
		}
		if len(page.Items) != 1 {
			return BillingProbeReport{}, ErrOrderInvalidResponse
		}
		item = page.Items[0]
	} else {
		item = *preloaded
	}
	observedAt := time.Now().UTC()
	if item.CheckedAt != nil {
		observedAt = *item.CheckedAt
	}
	if item.OrderID != order.id {
		return BillingProbeReport{}, ErrOrderInvalidResponse
	}
	report := redactBillingProbe(scope, order.id, packID, order.status, nil, item)
	if item.Available && orderSLAShipmentID.MatchString(order.shipment) {
		auth, child, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
		if err != nil {
			return BillingProbeReport{}, err
		}
		if child != order.child {
			return BillingProbeReport{}, ErrOrderAuthorizationChanged
		}
		if preloaded != nil && item.Shipping == nil {
			shipping := &BillingShipping{}
			shipping.SellerChargeUSD, shipping.SellerChargeDetailID = billingShippingDebit(item.Charges, order.shipment)
			var costs json.RawMessage
			code, err := s.orderGET(ctx, auth.AccessToken, "/marketplace/shipments/"+order.shipment+"/costs", &costs)
			if err != nil && code != http.StatusUnauthorized && code != http.StatusForbidden && code != http.StatusNotFound {
				return BillingProbeReport{}, err
			}
			if code == http.StatusOK {
				shipping.BuyerCostUSD, shipping.SellerNetCostUSD = parseBillingShipping(costs, child)
			}
			item.Shipping = shipping
			report = redactBillingProbe(scope, order.id, packID, order.status, nil, item)
		}
		// Official documented relationship endpoints only. A non-200 stops
		// relationship reads; 401/403 remains unknown, never a local fallback.
		var pack struct {
			ID       json.Number `json:"id"`
			Shipment struct {
				ID json.Number `json:"id"`
			} `json:"shipment"`
			Orders []struct {
				ID json.Number `json:"id"`
			} `json:"orders"`
		}
		code, err := s.orderGET(ctx, auth.AccessToken, "/marketplace/orders/pack/"+packID, &pack)
		if err != nil && code != http.StatusUnauthorized && code != http.StatusForbidden && code != http.StatusNotFound {
			return BillingProbeReport{}, err
		}
		report.PackMembership = billingProbeMembers{HTTPStatus: code, Rows: len(pack.Orders), Matches: code == http.StatusOK && pack.ID.String() == packID && pack.Shipment.ID.String() == order.shipment && len(pack.Orders) == 1 && pack.Orders[0].ID.String() == order.id}
		if report.PackMembership.Matches {
			var members []struct {
				Order  json.Number `json:"order_id"`
				Pack   json.Number `json:"pack_id"`
				Seller json.Number `json:"seller_id"`
			}
			code, err = s.orderGET(ctx, auth.AccessToken, "/shipments/"+order.shipment+"/orders", &members)
			if err != nil && code != http.StatusUnauthorized && code != http.StatusForbidden && code != http.StatusNotFound {
				return BillingProbeReport{}, err
			}
			matches := code == http.StatusOK && len(members) == 1
			for _, m := range members {
				matches = matches && m.Order.String() == order.id && m.Pack.String() == packID && m.Seller.String() == child
			}
			report.ShipmentMembership = billingProbeMembers{HTTPStatus: code, Rows: len(members), Matches: matches}
			if report.PackMembership.Matches && matches {
				report.OfficialMembership = "single_order_verified"
			}
		}
		current, currentChild, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
		if err != nil {
			return BillingProbeReport{}, err
		}
		if current.SellerID != auth.SellerID || currentChild != child {
			return BillingProbeReport{}, ErrOrderAuthorizationChanged
		}
	}
	currentOrder, err := lookup()
	if err != nil {
		return BillingProbeReport{}, err
	}
	if order != currentOrder {
		return BillingProbeReport{}, ErrOrderAuthorizationChanged
	}
	var lines []OrderLine
	if json.Unmarshal([]byte(order.items), &lines) != nil {
		return BillingProbeReport{}, ErrOrderInvalidResponse
	}
	report.MerchandiseUSD = orderSalesAmountUSD(lines)
	if report.PackMembership.Matches {
		var shipmentOrders []string
		if report.ShipmentMembership.HTTPStatus == http.StatusOK && report.ShipmentMembership.Matches {
			shipmentOrders = []string{order.id}
		}
		item.CheckedAt = &observedAt
		report.BillingEstimate, _ = billingEstimatedNet(billingExpectedNetEvidence{
			Scope: scope, OrderID: order.id, PackID: packID, ShipmentID: order.shipment,
			Status: order.status, Lines: lines, PackOrders: []string{order.id}, ShipmentOrders: shipmentOrders,
			Billing: item,
		}, report.ShipmentMembership.HTTPStatus, order.localShipmentOrders)
	}
	return report, nil
}
