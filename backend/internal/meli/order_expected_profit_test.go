package meli

import (
	"errors"
	"testing"
	"time"
)

func TestExpectedProfitConfirmedNet(t *testing.T) {
	observed := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	// Two observed Seller Center net amounts; costs and rate are synthetic
	// test inputs, not real purchase or conversion evidence.
	fx := ProfitFXRate{CNYPerUSD: "7", Source: "test accounting rate", AsOf: observed}
	net := ConfirmedExpectedUSDNet{Amount: "7.90", Source: ExpectedNetSellerCenter, ObservedAt: observed}
	got, err := EstimateProfitFromExpectedNet(net, fx, "35", "0")
	if err != nil || got.Kind != "expected" || got.NetUSD != "7.90" || got.ProfitUSD != "2.90" ||
		got.NetSource != ExpectedNetSellerCenter || !got.NetObservedAt.Equal(observed) ||
		got.MarginPercent == nil || *got.MarginPercent != "36.7" {
		t.Fatalf("expected profit: %+v %v", got, err)
	}
	net.Amount = "8.99"
	got, err = EstimateProfitFromExpectedNet(net, fx, "56", "7")
	if err != nil || got.ProfitUSD != "-0.01" {
		t.Fatalf("expected loss: %+v %v", got, err)
	}
	for _, tc := range []struct {
		net      ConfirmedExpectedUSDNet
		fx       ProfitFXRate
		purchase string
		extra    string
	}{
		{net, fx, "", "0"},
		{net, fx, "35", ""},
		{net, ProfitFXRate{}, "35", "0"},
		{ConfirmedExpectedUSDNet{Amount: "7.90", Source: "billing_fee_candidate", ObservedAt: observed}, fx, "35", "0"},
		{ConfirmedExpectedUSDNet{Amount: "7.90", Source: ExpectedNetSellerCenter}, fx, "35", "0"},
		{ConfirmedExpectedUSDNet{Amount: "NaN", Source: ExpectedNetSellerCenter, ObservedAt: observed}, fx, "35", "0"},
	} {
		if _, err := EstimateProfitFromExpectedNet(tc.net, tc.fx, tc.purchase, tc.extra); !errors.Is(err, ErrProfitInputsInvalid) {
			t.Fatalf("unverified input accepted: %v", err)
		}
	}
}
