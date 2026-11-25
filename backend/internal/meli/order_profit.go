package meli

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
	"time"
)

var ErrProfitInputsInvalid = errors.New("profit inputs require reconciled USD net and valid CNY costs")
var moneyUSDPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,15})(\.[0-9]{1,8})?$`)
var fxRatePattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,3})(\.[0-9]{1,6})?$`)

type VerifiedUSDNet struct {
	Amount    string
	Source    string
	SettledAt time.Time
}

type ProfitFXRate struct {
	CNYPerUSD string    `json:"rate"`
	Source    string    `json:"source"`
	AsOf      time.Time `json:"as_of"`
}

type EstimatedProfitUSD struct {
	NetUSD        string       `json:"net_usd"`
	ProfitUSD     string       `json:"profit_usd"`
	MarginPercent *string      `json:"margin_percent"`
	FX            ProfitFXRate `json:"fx"`
}

func ratOf(value string) (*big.Rat, error) {
	n, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, ErrProfitInputsInvalid
	}
	return n, nil
}

// roundRatio rounds to the nearest unit after multiplying by scale,
// with halves away from zero (including losses).
func roundRatio(value *big.Rat, scale int64) string {
	x := new(big.Rat).Mul(value, new(big.Rat).SetInt64(scale))
	num := new(big.Int).Set(x.Num())
	den := x.Denom()
	sign := num.Sign()
	num.Abs(num)
	quot, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	if new(big.Int).Mul(rem, big.NewInt(2)).Cmp(den) >= 0 {
		quot.Add(quot, big.NewInt(1))
	}
	if sign < 0 {
		quot.Neg(quot)
	}
	return quot.String()
}
func fixedPoint(whole string, places int) string {
	neg := strings.HasPrefix(whole, "-")
	whole = strings.TrimPrefix(whole, "-")
	for len(whole) <= places {
		whole = "0" + whole
	}
	if neg && strings.TrimLeft(whole, "0") != "" {
		return "-" + whole[:len(whole)-places] + "." + whole[len(whole)-places:]
	}
	return whole[:len(whole)-places] + "." + whole[len(whole)-places:]
}

// A2-only calculation contract: currently exercised by tests, not called by any
// runtime route or UI. Do not expose it before reconciliation is verified.
// EstimateProfitFromVerifiedNet never derives net USD from order total,
// tax base, sale_fee.net, charge totals, or listing net_proceeds.
func EstimateProfitFromVerifiedNet(receipt VerifiedUSDNet, fx ProfitFXRate, purchaseCNY, extraCNY string) (EstimatedProfitUSD, error) {
	if !moneyUSDPattern.MatchString(receipt.Amount) || !fxRatePattern.MatchString(fx.CNYPerUSD) ||
		strings.TrimSpace(receipt.Source) == "" || receipt.SettledAt.IsZero() ||
		strings.TrimSpace(fx.Source) == "" || fx.AsOf.IsZero() ||
		!purchaseCostPattern.MatchString(purchaseCNY) || !purchaseCostPattern.MatchString(extraCNY) {
		return EstimatedProfitUSD{}, ErrProfitInputsInvalid
	}
	net, _ := ratOf(receipt.Amount)
	rate, _ := ratOf(fx.CNYPerUSD)
	if rate.Sign() <= 0 {
		return EstimatedProfitUSD{}, ErrProfitInputsInvalid
	}
	purchase, _ := ratOf(purchaseCNY)
	extra, _ := ratOf(extraCNY)
	costUSD := new(big.Rat).Quo(new(big.Rat).Add(purchase, extra), rate)
	profit := new(big.Rat).Sub(net, costUSD)
	out := EstimatedProfitUSD{NetUSD: fixedPoint(roundRatio(net, 100), 2),
		ProfitUSD: fixedPoint(roundRatio(profit, 100), 2), FX: fx}
	if net.Sign() > 0 {
		margin := new(big.Rat).Quo(profit, net)
		val := fixedPoint(roundRatio(new(big.Rat).Mul(margin, big.NewRat(100, 1)), 10), 1)
		out.MarginPercent = &val
	}
	return out, nil
}
