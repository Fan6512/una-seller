package meli

import (
	"encoding/json"
	"testing"
)

func TestOrderSearchUsesNestedOrderIDsNotPackIDs(t *testing.T) {
	input := []byte(`{"results":[
		{"id":900,"orders":[{"id":101,"seller":{"id":456}},{"id":102,"seller":{"id":456}}]},
		{"id":901,"orders":[{"id":101,"seller":{"id":456}}]}
	],"paging":{"total":2,"limit":3}}`)
	ids, total, count, err := parseOrderSearch(input, "456", 0)
	if err != nil || total != 2 || count != 2 || len(ids) != 2 || ids[0] != "101" || ids[1] != "102" {
		t.Fatalf("ids=%v total=%d count=%d err=%v", ids, total, count, err)
	}
}

func TestOrderSearchRejectsMissingOrWrongSeller(t *testing.T) {
	for _, input := range []string{
		`{"results":[{"id":900,"orders":[]}],"paging":{"total":1}}`,
		`{"results":[{"id":900,"orders":[{"id":101,"seller":{"id":999}}]}],"paging":{"total":1}}`,
		`{"results":[{"id":900,"orders":[{"id":101,"seller":{"id":456}}]}]}`,
		`{"results":[],"paging":{"total":4}}`,
	} {
		if _, _, _, err := parseOrderSearch([]byte(input), "456", 0); err == nil {
			t.Errorf("accepted unsafe search: %s", input)
		}
	}
}

func TestOrderDetailSanitizesBuyerFieldsAndValidatesIdentity(t *testing.T) {
	const input = `{
		"id":101,"pack_id":900,"seller":{"id":456},"site_id":"MLB",
		"status":"paid","currency_id":"BRL","total_amount":123.45,
		"date_created":"2026-11-19T12:00:00Z","last_updated":"2026-11-19T12:01:00Z",
		"buyer":{"phone":"secret","email":"sensitive@example.com"},
		"shipping":{"id":801,"receiver_address":{"street_name":"private"}},
		"order_items":[{"item":{"id":"MLB1","title":"Product","variation_id":1234},
			"quantity":2,"unit_price":61.725,"currency_id":"BRL"}]
	}`
	record, err := parseOrderDetail([]byte(input), "101", "456", "MLB")
	if err != nil || record.ID != "101" || record.PackID != "900" ||
		record.ShipmentID != "801" || record.TotalAmount != "123.45" ||
		len(record.Items) != 1 || record.Items[0].VariationID != "1234" {
		t.Fatalf("parsed=%+v err=%v", record, err)
	}
	raw, _ := json.Marshal(record)
	for _, secret := range []string{"phone", "email", "receiver_address", "sensitive@example.com"} {
		if json.Valid(raw) && containsBytes(raw, []byte(secret)) {
			t.Errorf("PII persisted into order response: %s", secret)
		}
	}
	for _, test := range []struct{ order, seller, site string }{
		{"900", "456", "MLB"}, {"101", "999", "MLB"}, {"101", "456", "MLM"},
	} {
		if _, err = parseOrderDetail([]byte(input), test.order, test.seller, test.site); err == nil {
			t.Errorf("accepted mismatched order/seller/site: %+v", test)
		}
	}
}

func containsBytes(data, fragment []byte) bool {
	for i := 0; i+len(fragment) <= len(data); i++ {
		if string(data[i:i+len(fragment)]) == string(fragment) {
			return true
		}
	}
	return false
}

func TestOrderScopeAndStatusValidation(t *testing.T) {
	for _, scope := range []OrderScope{{"CBT", ""}, {"MLB", ""}, {"mlb", "remote"}, {"MLB", "remote/sub"}} {
		if orderScopeValid(scope) {
			t.Fatalf("accepted unsupported scope: %+v", scope)
		}
	}
	if !orderScopeValid(OrderScope{Site: "MLB", Logistic: "fulfillment"}) {
		t.Fatal("verified marketplace scope rejected")
	}
	if !ValidOrderStatus("paid") || !ValidOrderStatus("") ||
		ValidOrderStatus("paid;DROP TABLE") {
		t.Fatal("unexpected order status validation")
	}
}
