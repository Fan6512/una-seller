package meli

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var ErrOrderBillingInvalid = errors.New("invalid billing request or response")
var billingMoneyPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,15})(\.[0-9]{1,8})?$`)

// These types deliberately exclude buyers, payer IDs, shipping addresses, and raw billing JSON.
// A billing charge, payment taxable base, or a Brazil-only sale_fee.net is NOT net receipts.
// BillingSaleRef retains only the official financial linkage, not payer data.
// A shared charge may link multiple orders; do not infer allocation.
type BillingSaleRef struct {
	OrderID        string `json:"order_id,omitempty"`
	OperationID    string `json:"operation_id,omitempty"`
	TransactionUSD string `json:"transaction_usd,omitempty"`
}

type BillingCharge struct {
	DetailID             string           `json:"detail_id,omitempty"`
	Type                 string           `json:"type"`
	Subtype              string           `json:"subtype,omitempty"`
	Amount               string           `json:"amount"`
	Currency             string           `json:"currency"`
	DebitedFromOperation string           `json:"debited_from_operation,omitempty"`
	ReversalOf           string           `json:"reversal_of,omitempty"`
	Status               string           `json:"status,omitempty"`
	ShipmentID           string           `json:"shipment_id,omitempty"`
	PackID               string           `json:"pack_id,omitempty"`
	SalesRefs            []BillingSaleRef `json:"sales_refs,omitempty"`
}
type BillingChargeGroup struct {
	Currency      string `json:"currency"`
	Type          string `json:"type"`
	Subtype       string `json:"subtype,omitempty"`
	OriginalTotal string `json:"original_total"`
}

type BillingPayment struct {
	PaymentID          string `json:"payment_id,omitempty"`
	Status             string `json:"status"`
	MoneyReleaseStatus string `json:"money_release_status"`
	TaxableBaseUSD     string `json:"taxable_base_usd,omitempty"`
}
type BillingTax struct {
	Type        string `json:"type"`
	OriginalUSD string `json:"original_usd,omitempty"`
	RefundedUSD string `json:"refunded_usd,omitempty"`
}
type BillingShipping struct {
	BuyerCostUSD         string `json:"buyer_cost_usd,omitempty"`
	SellerChargeUSD      string `json:"seller_charge_usd,omitempty"`
	SellerNetCostUSD     string `json:"seller_net_cost_usd,omitempty"`
	SellerChargeDetailID string `json:"seller_charge_detail_id,omitempty"`
}

type BillingOrderPreview struct {
	OrderID        string               `json:"order_id"`
	Available      bool                 `json:"available"`
	Payments       []BillingPayment     `json:"payments"`
	Taxes          []BillingTax         `json:"taxes"`
	Charges        []BillingCharge      `json:"charges"`
	Reconciliation string               `json:"reconciliation"`
	ChargeGroups   []BillingChargeGroup `json:"charge_groups,omitempty"`
	Shipping       *BillingShipping     `json:"shipping,omitempty"`
	CheckedAt      *time.Time           `json:"checked_at,omitempty"`
}
type BillingPreviewPage struct {
	Items        []BillingOrderPreview `json:"items"`
	NetIncomeUSD *string               `json:"net_income_usd"`
}

func billingIDsValid(ids []string) bool {
	if len(ids) < 1 || len(ids) > 60 {
		return false
	}
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if !orderNumber(id) || id[0] == '0' || seen[id] {
			return false
		}
		seen[id] = true
	}
	return true
}

func billingPath(ids []string, seller string) (string, error) {
	if !billingIDsValid(ids) || !orderNumber(seller) || seller[0] == '0' {
		return "", ErrOrderBillingInvalid
	}
	query := url.Values{"order_ids": {strings.Join(ids, ",")}, "seller_id": {seller}}
	return "/billing/integration/group/ML/order/details?" + query.Encode(), nil
}
func optionalBillingID(value json.Number) bool {
	return value == "" || (orderNumber(value.String()) && value.String()[0] != '0')
}

func billingAmount(value json.Number) (string, error) {
	if value == "" {
		return "", nil
	}
	if !billingMoneyPattern.MatchString(value.String()) {
		return "", ErrOrderBillingInvalid
	}
	return value.String(), nil
}

// parseBillingPreview whitelists fiscal fields. It never returns a computed net payout:
// the official order/pack billing response includes charges and tax bases, not an
// independently verified per-order USD settled net amount.
func parseBillingPreview(raw []byte, requested []string) (BillingPreviewPage, error) {
	var source struct {
		Results []struct {
			OrderID     json.Number `json:"order_id"`
			PaymentInfo []struct {
				PaymentID          json.Number `json:"payment_id"`
				Status             string      `json:"status"`
				MoneyReleaseStatus string      `json:"money_release_status"`
				BaseAmountUSD      json.Number `json:"base_amount_usd"`
				TaxDetails         []struct {
					FinancialEntity string      `json:"mov_financial_entity"`
					OriginalUSD     json.Number `json:"original_amount_usd"`
					RefundedUSD     json.Number `json:"refunded_amount_usd"`
				} `json:"tax_details"`
			} `json:"payment_info"`
			Details []struct {
				ChargeInfo struct {
					DetailID             json.Number `json:"detail_id"`
					DetailType           string      `json:"detail_type"`
					DetailSubtype        string      `json:"detail_sub_type"`
					DetailAmount         json.Number `json:"detail_amount"`
					Status               string      `json:"status"`
					ReversalOf           json.Number `json:"charge_bonified_id"`
					DebitedFromOperation string      `json:"debited_from_operation"`
				} `json:"charge_info"`
				ShippingInfo struct {
					ID     json.Number `json:"shipping_id"`
					PackID json.Number `json:"pack_id"`
				} `json:"shipping_info"`
				SalesInfo []struct {
					OrderID        json.Number `json:"order_id"`
					OperationID    json.Number `json:"operation_id"`
					TransactionUSD json.Number `json:"transaction_amount_usd"`
				} `json:"sales_info"`
				CurrencyInfo struct {
					CurrencyID string `json:"currency_id"`
				} `json:"currency_info"`
			} `json:"details"`
		} `json:"results"`
	}
	if !billingIDsValid(requested) || json.Unmarshal(raw, &source) != nil ||
		source.Results == nil || len(source.Results) > len(requested) {
		return BillingPreviewPage{}, ErrOrderBillingInvalid
	}
	allowed := make(map[string]bool, len(requested))
	for _, id := range requested {
		allowed[id] = true
	}
	records := make(map[string]BillingOrderPreview, len(requested))
	for _, bill := range source.Results {
		id := bill.OrderID.String()
		if !allowed[id] || records[id].OrderID != "" || len(bill.Details) > 500 ||
			len(bill.PaymentInfo) > 100 {
			return BillingPreviewPage{}, ErrOrderBillingInvalid
		}
		out := BillingOrderPreview{OrderID: id, Available: true, Reconciliation: "review_required",
			Payments: []BillingPayment{}, Taxes: []BillingTax{}, Charges: []BillingCharge{}}
		for _, pay := range bill.PaymentInfo {
			base, err := billingAmount(pay.BaseAmountUSD)
			if err != nil || !optionalBillingID(pay.PaymentID) || len(pay.Status) > 60 || len(pay.MoneyReleaseStatus) > 60 || len(pay.TaxDetails) > 100 {
				return BillingPreviewPage{}, ErrOrderBillingInvalid
			}
			out.Payments = append(out.Payments, BillingPayment{PaymentID: pay.PaymentID.String(), Status: pay.Status,
				MoneyReleaseStatus: pay.MoneyReleaseStatus, TaxableBaseUSD: base})
			for _, tax := range pay.TaxDetails {
				original, err := billingAmount(tax.OriginalUSD)
				if err != nil {
					return BillingPreviewPage{}, err
				}
				refunded, err := billingAmount(tax.RefundedUSD)
				if err != nil || len(tax.FinancialEntity) > 80 {
					return BillingPreviewPage{}, ErrOrderBillingInvalid
				}
				out.Taxes = append(out.Taxes, BillingTax{Type: tax.FinancialEntity, OriginalUSD: original, RefundedUSD: refunded})
			}
		}
		chargeIDs := make(map[string]BillingCharge, len(bill.Details))
		for _, line := range bill.Details {
			c := line.ChargeInfo
			amount, err := billingAmount(c.DetailAmount)
			if err != nil || amount == "" || len(c.DetailType) > 80 || len(c.DetailSubtype) > 80 ||
				len(line.CurrencyInfo.CurrencyID) > 8 || len(c.DetailID) > 30 || len(c.ReversalOf) > 30 ||
				len(c.Status) > 40 || len(c.DebitedFromOperation) > 10 ||
				!optionalBillingID(line.ShippingInfo.PackID) || len(line.SalesInfo) > 60 ||
				(line.ShippingInfo.ID != "" && !orderSLAShipmentID.MatchString(line.ShippingInfo.ID.String())) {
				return BillingPreviewPage{}, ErrOrderBillingInvalid
			}
			charge := BillingCharge{DetailID: c.DetailID.String(), Type: c.DetailType,
				Subtype: c.DetailSubtype, Amount: amount, Currency: line.CurrencyInfo.CurrencyID,
				DebitedFromOperation: c.DebitedFromOperation,
				ReversalOf:           c.ReversalOf.String(), Status: c.Status, ShipmentID: line.ShippingInfo.ID.String(),
				PackID: line.ShippingInfo.PackID.String()}
			for _, sale := range line.SalesInfo {
				usd, err := billingAmount(sale.TransactionUSD)
				if err != nil || !optionalBillingID(sale.OrderID) || !optionalBillingID(sale.OperationID) {
					return BillingPreviewPage{}, ErrOrderBillingInvalid
				}
				// Missing identity remains missing; even a numeric amount is not net receipts.
				charge.SalesRefs = append(charge.SalesRefs, BillingSaleRef{
					OrderID: sale.OrderID.String(), OperationID: sale.OperationID.String(), TransactionUSD: usd,
				})
			}
			if charge.DetailID != "" {
				if previous, exists := chargeIDs[charge.DetailID]; exists {
					if !reflect.DeepEqual(previous, charge) {
						return BillingPreviewPage{}, ErrOrderBillingInvalid
					}
					continue
				}
				chargeIDs[charge.DetailID] = charge
			}
			out.Charges = append(out.Charges, charge)
		}
		out.ChargeGroups = groupBillingCharges(out.Charges)
		records[id] = out
	}
	page := BillingPreviewPage{Items: make([]BillingOrderPreview, 0, len(requested))}
	for _, id := range requested {
		if record, exists := records[id]; exists {
			page.Items = append(page.Items, record)
		} else {
			page.Items = append(page.Items, BillingOrderPreview{OrderID: id, Reconciliation: "not_available",
				Payments: []BillingPayment{}, Taxes: []BillingTax{}, Charges: []BillingCharge{}})
		}
	}
	return page, nil
}

// Group only identifiable records, using platform signs and native currencies.
func groupBillingCharges(charges []BillingCharge) []BillingChargeGroup {
	var entries []BillingChargeGroup
	var totals []*big.Rat
	index := make(map[[3]string]int)
	for _, charge := range charges {
		if charge.DetailID == "" || len(charge.Currency) != 3 || charge.Type == "" {
			return nil
		}
		key := [3]string{charge.Currency, charge.Type, charge.Subtype}
		i, found := index[key]
		if !found {
			i = len(entries)
			index[key] = i
			entries = append(entries, BillingChargeGroup{Currency: charge.Currency, Type: charge.Type, Subtype: charge.Subtype})
			totals = append(totals, new(big.Rat))
		}
		amount, err := ratOf(charge.Amount)
		if err != nil {
			return nil
		}
		totals[i].Add(totals[i], amount)
	}
	for i, total := range totals {
		entries[i].OriginalTotal = strings.TrimRight(strings.TrimRight(total.FloatString(8), "0"), ".")
	}
	return entries
}

// Shipment cost is scoped to a seller within the Shipment. Its net cost is
// already reported by the platform. Billing's original debit is a separate fact.
func parseBillingShipping(raw []byte, child string) (buyer, sellerNet string) {
	var payload struct {
		Currency string `json:"currency_id"`
		Senders  []struct {
			UserID json.Number `json:"user_id"`
			Cost   json.Number `json:"cost"`
		} `json:"senders"`
		Receiver struct {
			Cost    json.Number `json:"cost"`
			Details []struct {
				SenderID json.Number `json:"sender_id"`
				Amount   json.Number `json:"amount"`
			} `json:"cost_details"`
		} `json:"receiver"`
	}
	if json.Unmarshal(raw, &payload) != nil || payload.Currency != "USD" ||
		len(payload.Senders) > 60 || len(payload.Receiver.Details) > 60 {
		return "", ""
	}
	for _, sender := range payload.Senders {
		if sender.UserID.String() == child {
			if sellerNet != "" || !orderAmountPattern.MatchString(sender.Cost.String()) {
				return "", ""
			}
			sellerNet = sender.Cost.String()
		}
	}
	if sellerNet == "" {
		return "", ""
	}
	for _, detail := range payload.Receiver.Details {
		if detail.SenderID.String() == child {
			if buyer != "" || !orderAmountPattern.MatchString(detail.Amount.String()) {
				return "", sellerNet
			}
			buyer = detail.Amount.String()
		}
	}
	if buyer == "" && len(payload.Senders) == 1 && len(payload.Receiver.Details) == 0 &&
		orderAmountPattern.MatchString(payload.Receiver.Cost.String()) {
		buyer = payload.Receiver.Cost.String()
	}
	return buyer, sellerNet
}

func billingShippingDebit(charges []BillingCharge, shipment string) (amount, detailID string) {
	for _, charge := range charges {
		if charge.ShipmentID == shipment && charge.Type == "CHARGE" && charge.Subtype == "CCSI" &&
			charge.Currency == "USD" && charge.DebitedFromOperation == "YES" && charge.ReversalOf == "" {
			if detailID != "" || charge.DetailID == "" || !orderAmountPattern.MatchString(charge.Amount) {
				return "", ""
			}
			amount, detailID = charge.Amount, charge.DetailID
		}
	}
	return amount, detailID
}

// PreviewOrderBilling is an explicitly triggered official read (max 60).
// A successful complete single-order preview also updates the local snapshot.
// Every order must currently belong to the owner's verified marketplace child.
func (s *Service) PreviewOrderBilling(ctx context.Context, owner, store string, scope OrderScope,
	ids []string) (BillingPreviewPage, error) {
	return s.readOrderBilling(ctx, owner, store, scope, ids, true)
}

// readOrderBilling preserves the existing authorization and provider checks.
// Diagnostic callers pass persist=false to avoid updating financial snapshots.
func (s *Service) readOrderBilling(ctx context.Context, owner, store string, scope OrderScope,
	ids []string, persist bool) (BillingPreviewPage, error) {
	if !validUUID(store) || !orderScopeValid(scope) || !billingIDsValid(ids) {
		return BillingPreviewPage{}, ErrOrderBillingInvalid
	}
	// Authorization and current child mapping must be checked before platform access.
	auth, child, err := s.RequireSiteConnection(ctx, owner, store, scope.Site, scope.Logistic)
	if err != nil {
		return BillingPreviewPage{}, err
	}
	path, err := billingPath(ids, child)
	if err != nil {
		return BillingPreviewPage{}, err
	}
	checkScope := func() (bool, string, error) {
		var count int
		var shipment string
		err := s.db.QueryRow(ctx, `WITH connection AS (`+orderConnection+`)
 SELECT count(*), COALESCE(max(o.shipment_id), '') FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 JOIN app.meli_store_connections verified ON verified.id=c.id
 WHERE c.status='connected' AND o.site_id=$3 AND o.logistic_type=$4
 AND o.order_id=ANY($5::text[]) AND c.child_id=$6 AND verified.seller_id=$7`,
			store, owner, scope.Site, scope.Logistic, ids, child, auth.SellerID).Scan(&count, &shipment)
		return count == len(ids), shipment, err
	}
	scoped, shipment, err := checkScope()
	if err != nil {
		return BillingPreviewPage{}, err
	}
	if !scoped {
		return BillingPreviewPage{}, ErrOrderNotFound
	}
	startedAt := time.Now().UTC()
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var raw json.RawMessage
	status, err := s.orderGET(fetchCtx, auth.AccessToken, path, &raw)
	if err != nil {
		return BillingPreviewPage{}, err
	}
	// Partial and ambiguous documents are unsuitable for accounting previews.
	if status != http.StatusOK {
		return BillingPreviewPage{}, ErrOrderInvalidResponse
	}
	scoped, currentShipment, err := checkScope()
	if err != nil {
		return BillingPreviewPage{}, err
	}
	if !scoped || (len(ids) == 1 && currentShipment != shipment) {
		return BillingPreviewPage{}, ErrOrderAuthorizationChanged
	}
	page, err := parseBillingPreview(raw, ids)
	if err != nil {
		return BillingPreviewPage{}, ErrOrderInvalidResponse
	}
	if len(ids) == 1 && page.Items[0].Available && orderSLAShipmentID.MatchString(shipment) {
		item := &page.Items[0]
		shipping := &BillingShipping{}
		shipping.SellerChargeUSD, shipping.SellerChargeDetailID = billingShippingDebit(item.Charges, shipment)
		var costs json.RawMessage
		if code, err := s.orderGET(fetchCtx, auth.AccessToken, "/marketplace/shipments/"+shipment+"/costs", &costs); err == nil && code == http.StatusOK {
			shipping.BuyerCostUSD, shipping.SellerNetCostUSD = parseBillingShipping(costs, child)
		}
		scoped, currentShipment, err := checkScope()
		if err != nil {
			return BillingPreviewPage{}, err
		}
		if !scoped || currentShipment != shipment {
			return BillingPreviewPage{}, ErrOrderAuthorizationChanged
		}
		item.Shipping = shipping
	}
	// Preserve A1 batch behavior. The desktop requests one order at a time.
	// Never replace the last successful snapshot with missing billing or a
	// partial shipment-cost response (including a costs request failure).
	if persist && len(ids) == 1 && page.Items[0].Available && (shipment == "" ||
		(page.Items[0].Shipping != nil && page.Items[0].Shipping.SellerNetCostUSD != "")) {
		item, err := s.saveOrderBillingSnapshot(ctx, owner, store, scope, child, auth.SellerID, shipment, startedAt, page.Items[0])
		if err != nil {
			return BillingPreviewPage{}, err
		}
		page.Items[0] = item
	}
	return page, nil
}
