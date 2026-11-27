package meli

import (
	"net/http"
	"time"
)

const ExpectedNetBillingEstimate = "billing_estimate"
const BillingEstimateShipmentUnknown = "shipment_membership_unknown"

// BillingEstimatedUSDNet is a computed reference amount, not a confirmed
// Seller Center net or an actual settlement. It carries its own risk marker.
type BillingEstimatedUSDNet struct {
	Amount     string    `json:"amount_usd"`
	ObservedAt time.Time `json:"observed_at"`
	Source     string    `json:"source"`
	Risk       string    `json:"risk"`
}

// billingEstimatedNet accepts only an explicitly denied shipment membership
// read (or verified single membership), a verified single-order Pack and no
// known local sharing. All charges, including shipping, must identify the
// same order and pack. An empty successful response is not permission denial.
func billingEstimatedNet(e billingExpectedNetEvidence, shipmentHTTPStatus, localShipmentOrders int) (*BillingEstimatedUSDNet, bool) {
	if localShipmentOrders != 1 {
		return nil, false
	}
	unknown := (shipmentHTTPStatus == http.StatusUnauthorized || shipmentHTTPStatus == http.StatusForbidden) && e.ShipmentOrders == nil
	if !unknown && shipmentHTTPStatus != http.StatusOK {
		return nil, false
	}
	commissionCount := 0
	for _, charge := range e.Billing.Charges {
		if charge.PackID != e.PackID || len(charge.SalesRefs) != 1 ||
			charge.SalesRefs[0].OrderID != e.OrderID ||
			!billingIDsValid([]string{charge.SalesRefs[0].OperationID}) {
			return nil, false
		}
		if charge.Subtype != "CCSI" {
			commissionCount++
		}
	}
	// A missing commission breakdown is not proof of a zero commission.
	if commissionCount == 0 {
		return nil, false
	}
	amount, ok := billingNetArithmetic(e, unknown)
	if !ok {
		return nil, false
	}
	out := &BillingEstimatedUSDNet{
		Amount: amount, ObservedAt: *e.Billing.CheckedAt, Source: ExpectedNetBillingEstimate,
	}
	if unknown {
		out.Risk = BillingEstimateShipmentUnknown
	}
	return out, true
}
