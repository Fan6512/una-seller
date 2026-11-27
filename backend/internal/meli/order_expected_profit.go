package meli

import (
	"math/big"
	"time"
)

// ConfirmedExpectedUSDNet is the observed order-level expected net in Seller
// Center, never a settlement receipt or a number derived from a fee subtotal.
// Before constructing it, the caller must verify the same full order scope,
// USD currency, order-level attribution and the applicable fee/refund state.
// A source adapter must supply explicit order-level evidence; never a fee subtotal.
type ConfirmedExpectedUSDNet struct {
	Amount     string
	Source     string
	ObservedAt time.Time
}

const ExpectedNetSellerCenter = "seller_center_order_net"

type ExpectedOrderProfitUSD struct {
	EstimatedProfitUSD
	NetSource     string    `json:"net_source"`
	NetObservedAt time.Time `json:"net_observed_at"`
	NetRisk       string    `json:"net_risk,omitempty"`
	Kind          string    `json:"kind"`
	ProfitCNY     string    `json:"profit_cny"`
}

// EstimateProfitFromExpectedNet requires an explicitly confirmed expected net,
// actual CNY costs and an explicit dated accounting FX rate. A blank extra
// cost is unknown, not zero. Already-netted platform fees are not deducted again.
func EstimateProfitFromExpectedNet(net ConfirmedExpectedUSDNet, fx ProfitFXRate, purchaseCNY, extraCNY string) (ExpectedOrderProfitUSD, error) {
	if net.Source != ExpectedNetSellerCenter || net.ObservedAt.IsZero() {
		return ExpectedOrderProfitUSD{}, ErrProfitInputsInvalid
	}
	return estimateProfitWithSource(net.Amount, net.Source, "", net.ObservedAt, "expected", fx, purchaseCNY, extraCNY)
}

// EstimateProfitFromBillingNet never turns an estimate into a confirmed net.
// Costs still require actual values, including an explicit zero extra cost.
func EstimateProfitFromBillingNet(net BillingEstimatedUSDNet, fx ProfitFXRate, purchaseCNY, extraCNY string) (ExpectedOrderProfitUSD, error) {
	if net.Source != ExpectedNetBillingEstimate || net.ObservedAt.IsZero() ||
		(net.Risk != "" && net.Risk != BillingEstimateShipmentUnknown) {
		return ExpectedOrderProfitUSD{}, ErrProfitInputsInvalid
	}
	return estimateProfitWithSource(net.Amount, net.Source, net.Risk, net.ObservedAt, "billing_estimate", fx, purchaseCNY, extraCNY)
}

func estimateProfitWithSource(amount, source, risk string, observedAt time.Time, kind string, fx ProfitFXRate, purchaseCNY, extraCNY string) (ExpectedOrderProfitUSD, error) {
	profit, err := calculateProfitUSD(amount, fx, purchaseCNY, extraCNY)
	if err != nil {
		return ExpectedOrderProfitUSD{}, err
	}
	// Calculate directly in CNY before rounding, not from rounded USD profit.
	usd, _ := ratOf(amount)
	rate, _ := ratOf(fx.CNYPerUSD)
	purchase, _ := ratOf(purchaseCNY)
	extra, _ := ratOf(extraCNY)
	cny := new(big.Rat).Sub(new(big.Rat).Mul(usd, rate), new(big.Rat).Add(purchase, extra))
	return ExpectedOrderProfitUSD{
		ProfitCNY:          fixedPoint(roundRatio(cny, 100), 2),
		EstimatedProfitUSD: profit,
		NetSource:          source,
		NetRisk:            risk,
		NetObservedAt:      observedAt,
		Kind:               kind,
	}, nil
}
