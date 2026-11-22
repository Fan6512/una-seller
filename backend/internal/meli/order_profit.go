package meli

import (
	"errors"
	"math/big"
	"regexp"
	"strings"
)

var ErrProfitInputsInvalid = errors.New("profit inputs require reconciled USD net and valid CNY costs")
var moneyUSDPattern = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,15})(\.[0-9]{1,8})?$`)
var fxRatePattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,3})(\.[0-9]{1,6})?$`)

// EstimatedProfitCNY is a reference calculation, NOT a payout statement.
// Call only after independently reconciling real order-level USD net proceeds.
type EstimatedProfitCNY struct {
	NetCNY        string  `json:"net_cny"`
	ProfitCNY     string  `json:"profit_cny"`
	MarginPercent *string `json:"margin_percent"`
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

// EstimateProfitFromVerifiedNet never derives net USD from order total,
// tax base, sale_fee.net, charge totals, or listing net_proceeds.
func EstimateProfitFromVerifiedNet(netUSD, cnyPerUSD, purchaseCNY, packingCNY string) (EstimatedProfitCNY, error) {
	if !moneyUSDPattern.MatchString(netUSD) || !fxRatePattern.MatchString(cnyPerUSD) ||
		!purchaseCostPattern.MatchString(purchaseCNY) || !purchaseCostPattern.MatchString(packingCNY) {
		return EstimatedProfitCNY{}, ErrProfitInputsInvalid
	}
	net, _ := ratOf(netUSD)
	rate, _ := ratOf(cnyPerUSD)
	if rate.Sign() <= 0 {
		return EstimatedProfitCNY{}, ErrProfitInputsInvalid
	}
	purchase, _ := ratOf(purchaseCNY)
	packing, _ := ratOf(packingCNY)
	netCNY := new(big.Rat).Mul(net, rate)
	profit := new(big.Rat).Sub(netCNY, new(big.Rat).Add(purchase, packing))
	out := EstimatedProfitCNY{NetCNY: fixedPoint(roundRatio(netCNY, 100), 2),
		ProfitCNY: fixedPoint(roundRatio(profit, 100), 2)}
	if netCNY.Sign() > 0 {
		margin := new(big.Rat).Quo(profit, netCNY)
		val := fixedPoint(roundRatio(new(big.Rat).Mul(margin, big.NewRat(100, 1)), 10), 1)
		out.MarginPercent = &val
	}
	return out, nil
}
