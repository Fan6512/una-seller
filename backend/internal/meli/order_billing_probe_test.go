package meli

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRedactBillingProbe(t *testing.T) {
	const order = "2000000000000001"
	const pack = "2000000000000002"
	item := BillingOrderPreview{
		OrderID: order, Available: true,
		Payments: []BillingPayment{{PaymentID: "9876543210", Status: "approved", MoneyReleaseStatus: "released"}},
		Charges: []BillingCharge{
			{DetailID: "1234567890", PackID: pack, Amount: "1.09", Currency: "USD", SalesRefs: []BillingSaleRef{{OrderID: order, OperationID: "5678912340"}}},
			{DetailID: "1234567891", PackID: "2000000000000003", SalesRefs: []BillingSaleRef{{OrderID: "2000000000000004"}}},
		},
		Shipping: &BillingShipping{SellerChargeDetailID: "1234567892", SellerNetCostUSD: "3.15"},
	}
	report := redactBillingProbe(OrderScope{Site: "MLB", Logistic: "remote"}, order, pack, "paid",
		[]OrderLine{{Quantity: 1, UnitPrice: "8.74", CurrencyID: "USD"}}, item)
	if report.MerchandiseUSD != "8.74" || report.OfficialMembership != "not_verified" ||
		!report.Payments[0].HasPaymentID || !report.Charges[0].AllRefsMatchOrder ||
		!report.Charges[0].HasOperationID || report.Charges[0].PackRef != "matches" ||
		report.Charges[1].AllRefsMatchOrder || report.Charges[1].PackRef != "different" ||
		report.Shipping.SellerNetCostUSD != "3.15" {
		t.Fatalf("unexpected probe report: %+v", report)
	}
	raw, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{order, pack, "9876543210", "1234567890", "1234567891", "1234567892", "5678912340", "2000000000000003", "2000000000000004"} {
		if strings.Contains(string(raw), id) {
			t.Fatal("report leaked an identifier")
		}
	}
}
