package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var ErrOrderBillingInvalid = errors.New("invalid billing request or response")
var billingMoneyPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,15})(\.[0-9]{1,8})?$`)

// These types deliberately exclude buyers, payer IDs, shipping addresses, and raw billing JSON.
// A billing charge, payment taxable base, or a Brazil-only sale_fee.net is NOT net receipts.
type BillingCharge struct {
	DetailID             string `json:"detail_id,omitempty"`
	Type                 string `json:"type"`
	Subtype              string `json:"subtype,omitempty"`
	Amount               string `json:"amount"`
	Currency             string `json:"currency"`
	DebitedFromOperation bool   `json:"debited_from_operation"`
}
type BillingPayment struct {
	Status             string `json:"status"`
	MoneyReleaseStatus string `json:"money_release_status"`
	TaxableBaseUSD     string `json:"taxable_base_usd,omitempty"`
}
type BillingTax struct {
	Type        string `json:"type"`
	OriginalUSD string `json:"original_usd,omitempty"`
	RefundedUSD string `json:"refunded_usd,omitempty"`
}
type BillingOrderPreview struct {
	OrderID        string           `json:"order_id"`
	Available      bool             `json:"available"`
	Payments       []BillingPayment `json:"payments"`
	Taxes          []BillingTax     `json:"taxes"`
	Charges        []BillingCharge  `json:"charges"`
	Reconciliation string           `json:"reconciliation"`
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
					DebitedFromOperation string      `json:"debited_from_operation"`
				} `json:"charge_info"`
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
			if err != nil || len(pay.Status) > 60 || len(pay.MoneyReleaseStatus) > 60 || len(pay.TaxDetails) > 100 {
				return BillingPreviewPage{}, ErrOrderBillingInvalid
			}
			out.Payments = append(out.Payments, BillingPayment{Status: pay.Status,
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
		for _, line := range bill.Details {
			c := line.ChargeInfo
			amount, err := billingAmount(c.DetailAmount)
			if err != nil || amount == "" || len(c.DetailType) > 80 || len(c.DetailSubtype) > 80 ||
				len(line.CurrencyInfo.CurrencyID) > 8 || len(c.DetailID) > 30 {
				return BillingPreviewPage{}, ErrOrderBillingInvalid
			}
			out.Charges = append(out.Charges, BillingCharge{DetailID: c.DetailID.String(),
				Type: c.DetailType, Subtype: c.DetailSubtype, Amount: amount,
				Currency:             line.CurrencyInfo.CurrencyID,
				DebitedFromOperation: strings.EqualFold(c.DebitedFromOperation, "YES")})
		}
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

// PreviewOrderBilling is an explicitly triggered, read-only batch (max 60).
// Every order must currently belong to the owner's verified marketplace child.
func (s *Service) PreviewOrderBilling(ctx context.Context, owner, store string, scope OrderScope,
	ids []string) (BillingPreviewPage, error) {
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
	rows, err := s.db.Query(ctx, `WITH connection AS (`+orderConnection+`)
 SELECT o.order_id FROM app.meli_orders o
 JOIN connection c ON c.id=o.store_id AND c.child_id=o.seller_id
 WHERE c.status<>'disconnected' AND o.site_id=$3 AND o.logistic_type=$4
 AND o.order_id=ANY($5::text[])`, store, owner, scope.Site, scope.Logistic, ids)
	if err != nil {
		return BillingPreviewPage{}, err
	}
	found := map[string]bool{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			break
		}
		found[id] = true
	}
	if err == nil {
		err = rows.Err()
	}
	rows.Close()
	if err != nil {
		return BillingPreviewPage{}, err
	}
	if len(found) != len(ids) {
		return BillingPreviewPage{}, ErrOrderNotFound
	}
	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	var raw json.RawMessage
	_, err = s.orderGET(fetchCtx, auth.AccessToken, path, &raw)
	if err != nil {
		return BillingPreviewPage{}, err
	}
	return parseBillingPreview(raw, ids)
}

// No billing data is persisted or classified as final in this preview slice.
