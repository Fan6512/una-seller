package meli

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestBillingEstimatedNetShipmentRisk(t *testing.T) {
	e := billingCandidateFixture()
	for i := range e.Billing.Charges {
		e.Billing.Charges[i].PackID = e.PackID
		e.Billing.Charges[i].SalesRefs = []BillingSaleRef{{OrderID: e.OrderID, OperationID: "10001"}}
	}
	e.ShipmentOrders = nil
	if _, ok := billingExpectedNetCandidate(e); ok {
		t.Fatal("strict candidate accepted unknown membership")
	}
	got, ok := billingEstimatedNet(e, http.StatusUnauthorized, 1)
	if !ok || got.Amount != "4.50" || got.Source != ExpectedNetBillingEstimate || got.Risk != BillingEstimateShipmentUnknown {
		t.Fatalf("unexpected estimate: %+v", got)
	}
	if _, ok := billingEstimatedNet(e, http.StatusOK, 1); ok {
		t.Fatal("empty successful membership accepted")
	}
	if _, ok := billingEstimatedNet(e, http.StatusUnauthorized, 2); ok {
		t.Fatal("known local sharing accepted")
	}
	e.ShipmentOrders = []string{e.OrderID, "10002"}
	if _, ok := billingEstimatedNet(e, http.StatusOK, 1); ok {
		t.Fatal("known official sharing accepted")
	}
	e.ShipmentOrders = []string{e.OrderID}
	got, ok = billingEstimatedNet(e, http.StatusOK, 1)
	if !ok || got.Risk != "" {
		t.Fatal("verified membership did not retain estimate-only source")
	}
	e.Billing.Charges[0].SalesRefs[0].OperationID = ""
	if _, ok := billingEstimatedNet(e, http.StatusOK, 1); ok {
		t.Fatal("unattributed operation accepted")
	}
}

func TestBillingEstimateProfitSourceAndMissingCost(t *testing.T) {
	now := time.Date(2026, 11, 27, 0, 0, 0, 0, time.UTC)
	net := BillingEstimatedUSDNet{Amount: "13.48", ObservedAt: now, Source: ExpectedNetBillingEstimate, Risk: BillingEstimateShipmentUnknown}
	fx := ProfitFXRate{CNYPerUSD: "6.6", Source: dailyFXSource, AsOf: now}
	got, err := EstimateProfitFromBillingNet(net, fx, "55", "5")
	if err != nil || got.ProfitCNY != "28.97" || got.Kind != "billing_estimate" || got.NetSource != ExpectedNetBillingEstimate || got.NetRisk != BillingEstimateShipmentUnknown {
		t.Fatalf("estimated profit: %+v %v", got, err)
	}
	if _, err := EstimateProfitFromExpectedNet(ConfirmedExpectedUSDNet{Amount: net.Amount, Source: net.Source, ObservedAt: now}, fx, "55", "5"); err == nil {
		t.Fatal("estimate accepted by confirmed contract")
	}
	raw, err := json.Marshal(storedBillingEstimate{BillingEstimatedUSDNet: net, Evidence: "current"})
	if err != nil {
		t.Fatal(err)
	}
	ns := func(v string) sql.NullString { return sql.NullString{String: v, Valid: true} }
	i := expectedProfitInput{
		created: now, estimateEligible: true, evidence: "current", estimate: raw,
		purchase: ns("55"), rate: ns("6.6"), source: ns(dailyFXSource), observation: ns("2026-11-26"),
	}
	out := i.result()
	if out.NetUSD == nil || out.NetSource != ExpectedNetBillingEstimate || out.Profit != nil || out.Reason != "extra_missing" {
		t.Fatalf("missing cost: %+v", out)
	}
	i.extra = ns("5")
	out = i.result()
	if out.Profit == nil || out.Profit.ProfitCNY != "28.97" || out.Reason != "ready_estimate" {
		t.Fatalf("estimated runtime: %+v", out)
	}
	i.evidence = "changed"
	if out := i.result(); out.NetUSD != nil || out.Profit != nil || out.Reason != "estimate_stale" {
		t.Fatalf("stale estimate used: %+v", out)
	}
	i.evidence = "current"
	i.estimateEligible = false
	if out := i.result(); out.NetUSD != nil || out.Profit != nil {
		t.Fatal("ineligible order used estimate")
	}
	i.estimateEligible, i.eligible = true, true
	i.confirmation, err = json.Marshal(expectedProfitConfirmation{NetUSD: "14.00", ExtraCNY: "0", Source: ExpectedNetSellerCenter, ObservedAt: now, Evidence: "current"})
	if err != nil {
		t.Fatal(err)
	}
	out = i.result()
	if out.Profit == nil || out.NetSource != ExpectedNetSellerCenter || out.Profit.Kind != "expected" || out.Profit.ProfitCNY != "32.40" {
		t.Fatalf("confirmed precedence and explicit cost: %+v", out)
	}
}
