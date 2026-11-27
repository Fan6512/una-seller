package meli

import (
	"math/big"
	"time"
)

// billingExpectedNetEvidence describes the inputs of a narrow, diagnostic
// MLB/MLM remote rule. PackOrders and ShipmentOrders must be complete lists from
// the authorized official relationship reads, never inferred from local counts.
type billingExpectedNetEvidence struct {
	Scope          OrderScope
	OrderID        string
	PackID         string
	ShipmentID     string
	Status         string
	Lines          []OrderLine
	PackOrders     []string
	ShipmentOrders []string
	Billing        BillingOrderPreview
}

// billingExpectedNetCandidate computes a candidate only. Even success is not
// a confirmed expected net: absence of additional charges/discounts/taxes
// still needs a verified source contract. Do not pass the result to the
// confirmed-profit path or expose it as actual settlement.
func billingExpectedNetCandidate(e billingExpectedNetEvidence) (string, bool) {
	return billingNetArithmetic(e, false)
}

// billingNetArithmetic is shared only with the explicitly labelled estimate
// path. allowUnknownShipment is private; callers must first prove that the
// official membership read was denied, not that it returned a different order.
func billingNetArithmetic(e billingExpectedNetEvidence, allowUnknownShipment bool) (string, bool) {
	b := e.Billing
	oneOrder := func(ids []string) bool {
		return len(ids) == 1 && ids[0] == e.OrderID
	}
	if (e.Scope.Site != "MLB" && e.Scope.Site != "MLM") || e.Scope.Logistic != "remote" ||
		e.Status != "paid" || !billingIDsValid([]string{e.OrderID}) ||
		!billingIDsValid([]string{e.PackID}) || !orderSLAShipmentID.MatchString(e.ShipmentID) ||
		!oneOrder(e.PackOrders) || (!oneOrder(e.ShipmentOrders) && !(allowUnknownShipment && e.ShipmentOrders == nil)) ||
		!b.Available || b.OrderID != e.OrderID || b.CheckedAt == nil || b.CheckedAt.Equal(time.Time{}) ||
		len(b.Payments) != 1 || b.Payments[0].Status != "approved" ||
		(b.Payments[0].MoneyReleaseStatus != "pending" && b.Payments[0].MoneyReleaseStatus != "released") ||
		b.Taxes == nil || len(b.Taxes) != 0 || b.Shipping == nil {
		return "", false
	}
	// Require exact cents; do not round uncertain charge allocations.
	cents := func(raw string) (*big.Rat, bool) {
		if !moneyUSDPattern.MatchString(raw) {
			return nil, false
		}
		v, err := ratOf(raw)
		if err != nil || v.Sign() < 0 || !new(big.Rat).Mul(v, big.NewRat(100, 1)).IsInt() {
			return nil, false
		}
		return v, true
	}
	gross, ok := cents(orderSalesAmountUSD(e.Lines))
	if !ok || gross.Sign() <= 0 {
		return "", false
	}
	buyer, ok := cents(b.Shipping.BuyerCostUSD)
	if !ok {
		return "", false
	}
	shipping, ok := cents(b.Shipping.SellerChargeUSD)
	if !ok {
		return "", false
	}
	shippingNet, ok := cents(b.Shipping.SellerNetCostUSD)
	if !ok || new(big.Rat).Sub(shipping, buyer).Cmp(shippingNet) != 0 {
		return "", false
	}
	seen := map[string]bool{}
	fees := new(big.Rat)
	shippingCount := 0
	for _, charge := range b.Charges {
		if !billingIDsValid([]string{charge.DetailID}) || seen[charge.DetailID] ||
			charge.Type != "CHARGE" || charge.Currency != "USD" || charge.ReversalOf != "" ||
			charge.Status != "" || (charge.DebitedFromOperation != "YES" && charge.DebitedFromOperation != "NO") ||
			(charge.PackID != "" && charge.PackID != e.PackID) ||
			(charge.ShipmentID != "" && charge.ShipmentID != e.ShipmentID) {
			return "", false
		}
		seen[charge.DetailID] = true
		amount, ok := cents(charge.Amount)
		if !ok {
			return "", false
		}
		for _, ref := range charge.SalesRefs {
			if ref.OrderID != e.OrderID {
				return "", false
			}
		}
		switch charge.Subtype {
		case "CCSI":
			if charge.DetailID != b.Shipping.SellerChargeDetailID ||
				charge.ShipmentID != e.ShipmentID || charge.DebitedFromOperation != "YES" ||
				amount.Cmp(shipping) != 0 {
				return "", false
			}
			shippingCount++
		case "CVMPI", "CVMPCB", "CVMLCB", "CV":
			if (e.Scope.Site == "MLB" && charge.Subtype == "CV") ||
				(e.Scope.Site == "MLM" && (charge.Subtype == "CVMPCB" || charge.Subtype == "CVMLCB" || charge.DebitedFromOperation != "YES")) {
				return "", false
			}
			// Require explicit order attribution. NO is not zero; the observed
			// Seller Center commission includes both YES and NO charges.
			if len(charge.SalesRefs) == 0 {
				return "", false
			}
			fees.Add(fees, amount)
		default:
			return "", false
		}
	}
	if shippingCount != 1 {
		return "", false
	}
	net := new(big.Rat).Sub(gross, fees)
	net.Sub(net, shippingNet) // CCSI is already reflected in shippingNet.
	if net.Sign() < 0 {
		return "", false
	}
	return net.FloatString(2), true
}
