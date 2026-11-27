package meli

import (
	"testing"
	"time"
)

func billingCandidateFixture() billingExpectedNetEvidence {
	checked := time.Date(2026, time.November, 27, 0, 0, 0, 0, time.UTC)
	const order = "2000000000000001"
	const pack = "2000000000000002"
	const shipment = "10000000001"
	return billingExpectedNetEvidence{
		Scope: OrderScope{Site: "MLB", Logistic: "remote"}, OrderID: order, PackID: pack,
		ShipmentID: shipment, Status: "paid",
		Lines:          []OrderLine{{Quantity: 1, UnitPrice: "8.74", CurrencyID: "USD"}},
		PackOrders:     []string{order},
		ShipmentOrders: []string{order},
		Billing: BillingOrderPreview{
			OrderID: order, Available: true, CheckedAt: &checked,
			Payments: []BillingPayment{{Status: "approved", MoneyReleaseStatus: "released"}},
			Taxes:    []BillingTax{},
			Shipping: &BillingShipping{BuyerCostUSD: "7.63", SellerChargeUSD: "10.78", SellerNetCostUSD: "3.15", SellerChargeDetailID: "104"},
			Charges: []BillingCharge{
				{DetailID: "101", Type: "CHARGE", Subtype: "CVMPI", Amount: "0.16", Currency: "USD", DebitedFromOperation: "YES", SalesRefs: []BillingSaleRef{{OrderID: order}}},
				{DetailID: "102", Type: "CHARGE", Subtype: "CVMPCB", Amount: "0.15", Currency: "USD", DebitedFromOperation: "NO", SalesRefs: []BillingSaleRef{{OrderID: order}}},
				{DetailID: "103", Type: "CHARGE", Subtype: "CVMLCB", Amount: "0.78", Currency: "USD", DebitedFromOperation: "NO", SalesRefs: []BillingSaleRef{{OrderID: order}}},
				{DetailID: "104", Type: "CHARGE", Subtype: "CCSI", Amount: "10.78", Currency: "USD", DebitedFromOperation: "YES", ShipmentID: shipment},
			},
		},
	}
}

func TestBillingExpectedNetCandidate(t *testing.T) {
	// Amounts are from the observed MLB sample. Relationship and charge IDs
	// are synthetic; this test is not evidence of official completeness.
	e := billingCandidateFixture()
	if got, ok := billingExpectedNetCandidate(e); !ok || got != "4.50" {
		t.Fatalf("candidate = %q, %v; want 4.50, true", got, ok)
	}
	for _, tt := range []struct {
		name   string
		mutate func(*billingExpectedNetEvidence)
	}{
		{"unsupported site", func(e *billingExpectedNetEvidence) { e.Scope.Site = "MLC" }},
		{"unknown official pack", func(e *billingExpectedNetEvidence) { e.PackOrders = nil }},
		{"shared shipment", func(e *billingExpectedNetEvidence) { e.ShipmentOrders = append(e.ShipmentOrders, "2000000000000003") }},
		{"refund", func(e *billingExpectedNetEvidence) { e.Status = "partially_refunded" }},
		{"missing tax information", func(e *billingExpectedNetEvidence) { e.Billing.Taxes = nil }},
		{"tax", func(e *billingExpectedNetEvidence) { e.Billing.Taxes = []BillingTax{{OriginalUSD: "0.20"}} }},
		{"duplicate fee", func(e *billingExpectedNetEvidence) {
			e.Billing.Charges = append(e.Billing.Charges, e.Billing.Charges[0])
		}},
		{"unattributed fee", func(e *billingExpectedNetEvidence) { e.Billing.Charges[0].SalesRefs = nil }},
		{"unknown fee", func(e *billingExpectedNetEvidence) { e.Billing.Charges[0].Subtype = "UNKNOWN" }},
		{"reversed fee", func(e *billingExpectedNetEvidence) { e.Billing.Charges[0].ReversalOf = "105" }},
		{"wrong shipping", func(e *billingExpectedNetEvidence) { e.Billing.Shipping.SellerNetCostUSD = "3.16" }},
		{"missing buyer freight", func(e *billingExpectedNetEvidence) { e.Billing.Shipping.BuyerCostUSD = "" }},
		{"non USD line", func(e *billingExpectedNetEvidence) { e.Lines[0].CurrencyID = "BRL" }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := billingCandidateFixture()
			tt.mutate(&e)
			if got, ok := billingExpectedNetCandidate(e); ok || got != "" {
				t.Fatalf("unexpected candidate %q, %v", got, ok)
			}
		})
	}
}

func TestBillingExpectedNetCandidateMLM(t *testing.T) {
	// Observed amounts and fee codes; all identities and complete membership
	// are synthetic. The real shipment membership returned 401, so it is unknown.
	e := billingCandidateFixture()
	e.Scope.Site = "MLM"
	e.Lines = []OrderLine{{Quantity: 2, UnitPrice: "11.89", CurrencyID: "USD"}}
	e.Billing.Payments[0].MoneyReleaseStatus = "pending"
	e.Billing.Shipping = &BillingShipping{BuyerCostUSD: "0", SellerChargeUSD: "5.54", SellerNetCostUSD: "5.54", SellerChargeDetailID: "104"}
	e.Billing.Charges = []BillingCharge{
		{DetailID: "101", Type: "CHARGE", Subtype: "CVMPI", Amount: "0.42", Currency: "USD", DebitedFromOperation: "YES", SalesRefs: []BillingSaleRef{{OrderID: e.OrderID}}},
		{DetailID: "102", Type: "CHARGE", Subtype: "CV", Amount: "4.34", Currency: "USD", DebitedFromOperation: "YES", SalesRefs: []BillingSaleRef{{OrderID: e.OrderID}}},
		{DetailID: "104", Type: "CHARGE", Subtype: "CCSI", Amount: "5.54", Currency: "USD", DebitedFromOperation: "YES", ShipmentID: e.ShipmentID},
	}
	if got, ok := billingExpectedNetCandidate(e); !ok || got != "13.48" {
		t.Fatalf("MLM candidate = %q, %v", got, ok)
	}
	e.ShipmentOrders = nil
	if got, ok := billingExpectedNetCandidate(e); ok || got != "" {
		t.Fatal("unknown official shipment membership produced a candidate")
	}
	e.ShipmentOrders = []string{e.OrderID}
	e.Billing.Charges[1].DebitedFromOperation = "NO"
	if _, ok := billingExpectedNetCandidate(e); ok {
		t.Fatal("unobserved MLM fee mode accepted")
	}
	e.Billing.Charges[1].DebitedFromOperation = "YES"
	e.Scope.Site = "MLB"
	if _, ok := billingExpectedNetCandidate(e); ok {
		t.Fatal("MLM fee code accepted as MLB")
	}
}
