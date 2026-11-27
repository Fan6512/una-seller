package meli

import (
	"context"
	"encoding/json"
)

// BillingProbeReport is an operator-only, redacted diagnostic. It is not a
// confirmed expected-net source, and a local singleton pack is not proof of
// the official pack or shipment membership.
type BillingProbeReport struct {
	Site               string                `json:"site"`
	Logistic           string                `json:"logistic"`
	OrderStatus        string                `json:"order_status"`
	MerchandiseUSD     string                `json:"merchandise_usd,omitempty"`
	Available          bool                  `json:"billing_available"`
	Payments           []billingProbePayment `json:"payments"`
	Charges            []billingProbeCharge  `json:"charges"`
	Taxes              []BillingTax          `json:"taxes"`
	Shipping           *billingProbeShipping `json:"shipping,omitempty"`
	OfficialMembership string                `json:"official_membership"`
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
	if !validUUID(owner) || !validUUID(store) || !orderScopeValid(scope) || !billingIDsValid([]string{packID}) {
		return BillingProbeReport{}, ErrOrderBillingInvalid
	}
	lookup := func() (string, string, []byte, error) {
		rows, err := s.db.Query(ctx, `WITH connection AS (`+orderConnection+`)
SELECT o.order_id, o.status, o.order_items
FROM app.meli_orders o
JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
WHERE c.status='connected' AND o.site_id=$3 AND o.logistic_type=$4 AND o.pack_id=$5
LIMIT 2`, store, owner, scope.Site, scope.Logistic, packID)
		if err != nil {
			return "", "", nil, err
		}
		defer rows.Close()
		var id, status string
		var items []byte
		count := 0
		for rows.Next() {
			count++
			if err := rows.Scan(&id, &status, &items); err != nil {
				return "", "", nil, err
			}
		}
		if err := rows.Err(); err != nil {
			return "", "", nil, err
		}
		if count != 1 || !billingIDsValid([]string{id}) {
			return "", "", nil, ErrOrderNotFound
		}
		return id, status, items, nil
	}
	id, status, raw, err := lookup()
	if err != nil {
		return BillingProbeReport{}, err
	}
	page, err := s.readOrderBilling(ctx, owner, store, scope, []string{id}, false)
	if err != nil {
		return BillingProbeReport{}, err
	}
	currentID, currentStatus, currentRaw, err := lookup()
	if err != nil {
		return BillingProbeReport{}, err
	}
	if id != currentID || status != currentStatus || string(raw) != string(currentRaw) {
		return BillingProbeReport{}, ErrOrderAuthorizationChanged
	}
	var lines []OrderLine
	if json.Unmarshal(raw, &lines) != nil || len(page.Items) != 1 || page.Items[0].OrderID != id {
		return BillingProbeReport{}, ErrOrderInvalidResponse
	}
	return redactBillingProbe(scope, id, packID, status, lines, page.Items[0]), nil
}
