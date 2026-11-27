package meli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collectors/billing"
)

func TestBillingCollectorBatchResumeAndProfit(t *testing.T) {
	s, db, owner, store := setupRefreshTest(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Historical synthetic date isolates the collector's date window from
	// other tests sharing the PG instance. Provider requests use httptest.
	now := time.Date(2001, 2, 3, 12, 0, 0, 0, time.UTC)
	created := now.Add(-3 * time.Hour)
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
		(store_id,site_id,logistic_type,seller_id,order_id,pack_id,shipment_id,status,currency_id,total_amount,date_created,last_updated,order_items)
		VALUES($1,'MLB','remote','987654','10001','20001','80001','paid','USD',48.07,$2,$2,
		'[{"quantity":1,"unit_price":"8.74","currency_id":"USD"}]'),
		($1,'MLB','remote','987654','10002','20002','80002','paid','USD',48.07,$2,$2,
		'[{"quantity":1,"unit_price":"8.74","currency_id":"USD"}]')`, store, created); err != nil {
		t.Fatal(err)
	}
	billingCalls, relationCalls := 0, 0
	failOnce := true
	record := func(order, pack, shipment, feeID, shippingID int) string {
		return fmt.Sprintf(`{"order_id":%d,"payment_info":[{"payment_id":90001,"status":"approved","money_release_status":"released","tax_details":[]}],
		"details":[
		{"charge_info":{"detail_id":%d,"detail_type":"CHARGE","detail_sub_type":"CVMPI","detail_amount":1.09,"debited_from_operation":"YES"},
		"shipping_info":{"pack_id":%d},"sales_info":[{"order_id":%d,"operation_id":90001}],"currency_info":{"currency_id":"USD"}},
		{"charge_info":{"detail_id":%d,"detail_type":"CHARGE","detail_sub_type":"CCSI","detail_amount":10.78,"debited_from_operation":"YES"},
		"shipping_info":{"pack_id":%d,"shipping_id":%d},"sales_info":[{"order_id":%d,"operation_id":90001}],"currency_info":{"currency_id":"USD"}}]}`,
			order, feeID, pack, order, shippingID, pack, shipment, order)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("unexpected provider request")
			w.WriteHeader(403)
			return
		}
		path := r.URL.Path
		switch {
		case path == "/billing/integration/group/ML/order/details":
			billingCalls++
			if r.URL.Query().Get("order_ids") != "10001,10002" || r.URL.Query().Get("seller_id") != "987654" {
				t.Errorf("not a scoped batch: %s", r.URL.RawQuery)
			}
			fmt.Fprintf(w, `{"results":[%s,%s]}`, record(10001, 20001, 80001, 101, 102), record(10002, 20002, 80002, 103, 104))
		case strings.HasPrefix(path, "/marketplace/shipments/") && strings.HasSuffix(path, "/costs"):
			relationCalls++
			if failOnce {
				failOnce = false
				w.WriteHeader(429)
				return
			}
			fmt.Fprint(w, `{"currency_id":"USD","senders":[{"user_id":987654,"cost":3.15}],"receiver":{"cost":7.63}}`)
		case path == "/marketplace/orders/pack/20001":
			relationCalls++
			fmt.Fprint(w, `{"id":20001,"shipment":{"id":80001},"orders":[{"id":10001}]}`)
		case path == "/marketplace/orders/pack/20002":
			relationCalls++
			fmt.Fprint(w, `{"id":20002,"shipment":{"id":80002},"orders":[{"id":10002}]}`)
		case strings.HasPrefix(path, "/shipments/") && strings.HasSuffix(path, "/orders"):
			relationCalls++
			w.WriteHeader(401)
		default:
			t.Errorf("unexpected endpoint %s", path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	s.apiURL = server.URL
	firstPass := billingPassStats{}
	if err := s.refreshBillingEstimatesPass(ctx, now, &firstPass); err == nil || billingCalls != 1 || relationCalls != 1 || firstPass.targets != 2 || firstPass.done != 0 || firstPass.review != 0 {
		t.Fatalf("expected one batch then rate limit: %v, %d/%d", err, billingCalls, relationCalls)
	}
	// The simulated scheduler advances two hours; the existing HTTP gate
	// uses wall time, so expire its synthetic five-second cooldown as well.
	s.orderRetryUntil = time.Time{}
	secondPass := billingPassStats{}
	if err := s.refreshBillingEstimatesPass(ctx, now.Add(billing.Interval), &secondPass); err != nil || billingCalls != 1 || secondPass.targets != 2 || secondPass.done != 2 || secondPass.review != 0 {
		t.Fatalf("resume refetched Billing: %v, calls=%d", err, billingCalls)
	}
	after := relationCalls
	if err := s.refreshBillingEstimates(ctx, now.Add(2*billing.Interval)); err != nil || billingCalls != 1 || relationCalls != after {
		t.Fatalf("processed batch repeated: %v", err)
	}
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	out, err := s.GetExpectedOrderProfit(ctx, owner, store, scope, "10001")
	if err != nil || out.NetUSD == nil || *out.NetUSD != "4.50" || out.NetSource != ExpectedNetBillingEstimate || out.NetRisk != BillingEstimateShipmentUnknown || out.Profit != nil || out.HasConfirmation {
		t.Fatalf("estimated net: %+v %v", out, err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_order_purchases
		(store_id,site_id,logistic_type,seller_id,order_id,purchase_order_id,cost_cny)
		VALUES($1,'MLB','remote','987654','10001','12345678',20)`, store); err != nil {
		t.Fatal(err)
	}
	rateWrite, err := db.Exec(ctx, `INSERT INTO app.daily_usd_cny_rates
		(accounting_day,observation_day,raw_rate,locked_rate,source)
		VALUES('2001-02-03','2001-02-02',6.68,6.6,'frankfurter_v2_ecb') ON CONFLICT DO NOTHING`)
	if err != nil {
		t.Fatal(err)
	}
	if rateWrite.RowsAffected() == 1 {
		defer db.Exec(context.Background(), `DELETE FROM app.daily_usd_cny_rates WHERE accounting_day='2001-02-03'`)
	}
	out, err = s.SetExpectedProfitExtraCost(ctx, owner, store, scope, "10001", out.Revision, "0")
	if err != nil || out.Profit == nil || out.Profit.ProfitCNY != "9.70" || out.HasConfirmation {
		t.Fatalf("cost-only estimated profit: %+v %v", out, err)
	}
	summary, err := s.OperationsSummary(ctx, owner)
	if err != nil {
		t.Fatal(err)
	}
	summary.CheckedAt = now
	sales, err := s.operationsSales(ctx, owner, store, "MLB", summary)
	if err != nil || sales.ProfitOrders != 0 || sales.ExpectedProfitCNY != nil || sales.BillingEstimatedOrders != 1 ||
		sales.BillingEstimatedProfitCNY == nil || *sales.BillingEstimatedProfitCNY != "9.70" {
		t.Fatalf("estimated rollup: %+v %v", sales, err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='80001' WHERE store_id=$1 AND order_id='10002'`, store); err != nil {
		t.Fatal(err)
	}
	out, err = s.GetExpectedOrderProfit(ctx, owner, store, scope, "10001")
	if err != nil || out.NetUSD != nil || out.Profit != nil {
		t.Fatalf("known sharing did not invalidate estimate: %+v %v", out, err)
	}

	// Simulate a process that stopped after reserving its final attempt.
	// Both phases must become visible review states without another GET.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='80002'
		WHERE store_id=$1 AND order_id='10002'`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders o SET
		expected_profit_billing_estimate=NULL,
		expected_profit_billing_attempt=jsonb_build_object('evidence',`+expectedProfitEvidence+`,
			'state',CASE WHEN o.order_id='10001' THEN 'pending' ELSE 'billing_read' END,
			'billing_attempts',3,'relation_attempts',3,'retry_unix',0)
		FROM app.meli_store_connections c WHERE c.id=o.store_id AND o.store_id=$1
		AND o.order_id IN ('10001','10002')`, store); err != nil {
		t.Fatal(err)
	}
	exhaustedPass := billingPassStats{}
	if err := s.refreshBillingEstimatesPass(ctx, now.Add(3*billing.Interval), &exhaustedPass); err != nil || billingCalls != 1 || relationCalls != after || exhaustedPass.targets != 2 || exhaustedPass.done != 0 || exhaustedPass.review != 2 {
		t.Fatalf("exhausted attempts made a request: %v", err)
	}
	for _, id := range []string{"10001", "10002"} {
		got, err := s.GetExpectedOrderProfit(ctx, owner, store, scope, id)
		if err != nil || got.Reason != "billing_review" || got.CollectionReason != "attempts_exhausted" || got.Profit != nil {
			t.Fatalf("exhausted attempt not visible: %+v %v", got, err)
		}
	}
	// Local contention before a GET must not consume the Billing budget.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET expected_profit_billing_attempt=NULL
		WHERE store_id=$1 AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	s.orderRequests.Lock()
	busyErr := s.refreshBillingEstimates(ctx, now.Add(4*billing.Interval))
	s.orderRequests.Unlock()
	if busyErr == nil || billingCalls != 1 || relationCalls != after {
		t.Fatalf("local busy request: %v", busyErr)
	}
	var attempts int
	var reason string
	if err := db.QueryRow(ctx, `SELECT (expected_profit_billing_attempt->>'billing_attempts')::int,
		expected_profit_billing_attempt->>'reason' FROM app.meli_orders
		WHERE store_id=$1 AND order_id='10001'`, store).Scan(&attempts, &reason); err != nil || attempts != 0 || reason != "request_busy" {
		t.Fatalf("busy consumed budget: %d/%s %v", attempts, reason, err)
	}
}
