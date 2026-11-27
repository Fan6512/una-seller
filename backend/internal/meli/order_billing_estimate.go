package meli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const ExpectedNetBillingEstimate = "billing_estimate"
const BillingEstimateShipmentUnknown = "shipment_membership_unknown"

// BillingEstimatedUSDNet is a computed reference amount, not a confirmed
// Seller Center net or an actual settlement. It carries its own risk marker.
type BillingEstimatedUSDNet struct {
	Amount      string    `json:"amount_usd"`
	ObservedAt  time.Time `json:"observed_at"`
	Source      string    `json:"source"`
	Risk        string    `json:"risk"`
	calculation *billingEstimateCalculation
}

// billingEstimatedNet accepts only an explicitly denied shipment membership
// read (or verified single membership), a verified single-order Pack and no
// known local sharing. All charges, including shipping, must identify the
// same order and pack. An empty successful response is not permission denial.
func billingEstimatedNet(e billingExpectedNetEvidence, shipmentHTTPStatus, localShipmentOrders int) (*BillingEstimatedUSDNet, bool) {
	if localShipmentOrders != 1 || e.Billing.Shipping == nil ||
		e.Billing.Shipping.SenderScope != billingShippingSingleSender {
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
	out.calculation = &billingEstimateCalculation{
		Scope: e.Scope, OrderID: e.OrderID, PackID: e.PackID, ShipmentID: e.ShipmentID,
		MerchandiseUSD: orderSalesAmountUSD(e.Lines), Billing: e.Billing,
		PackOrders: e.PackOrders, ShipmentOrders: e.ShipmentOrders,
		ShipmentHTTPStatus: shipmentHTTPStatus, LocalShipmentOrders: localShipmentOrders,
		VerifiedAt: time.Now().UTC(),
	}
	return out, true
}

// Stored privately alongside the estimate; only whitelisted billing fields,
// never raw provider bodies, order titles, buyer details or credentials.
type billingEstimateCalculation struct {
	Scope               OrderScope          `json:"scope"`
	OrderID             string              `json:"order_id"`
	PackID              string              `json:"pack_id"`
	ShipmentID          string              `json:"shipment_id"`
	MerchandiseUSD      string              `json:"merchandise_usd"`
	Billing             BillingOrderPreview `json:"billing"`
	PackOrders          []string            `json:"pack_orders"`
	ShipmentOrders      []string            `json:"shipment_orders"`
	ShipmentHTTPStatus  int                 `json:"shipment_http_status"`
	LocalShipmentOrders int                 `json:"local_shipment_orders"`
	VerifiedAt          time.Time           `json:"verified_at"`
}

func (e storedBillingEstimate) calculationDigest() (string, error) {
	if e.Calculation == nil {
		return "", errors.New("missing billing calculation evidence")
	}
	raw, err := json.Marshal(struct {
		Net         BillingEstimatedUSDNet
		Calculation *billingEstimateCalculation
	}{e.BillingEstimatedUSDNet, e.Calculation})
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

func newStoredBillingEstimate(net BillingEstimatedUSDNet, evidence string) (storedBillingEstimate, error) {
	stored := storedBillingEstimate{BillingEstimatedUSDNet: net, Evidence: evidence, Calculation: net.calculation}
	digest, err := stored.calculationDigest()
	stored.CalculationSHA256 = digest
	return stored, err
}

func (e storedBillingEstimate) calculationValid() bool {
	digest, err := e.calculationDigest()
	return err == nil && digest == e.CalculationSHA256
}
