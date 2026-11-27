package meli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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

func TestBillingProbeMembershipAndNoCacheWrite(t *testing.T) {
	for _, code := range []int{http.StatusOK, http.StatusUnauthorized, http.StatusTooManyRequests} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			s, db, owner, store := setupRefreshTest(t)
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
				t.Fatal(err)
			}
			if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
			(store_id,site_id,logistic_type,seller_id,order_id,pack_id,shipment_id,status,currency_id,total_amount,date_created,last_updated,order_items,billing_snapshot)
			VALUES($1,'MLB','remote','987654','10001','20001','80001','paid','USD',8.74,now(),now(),
			'[{"quantity":1,"unit_price":"8.74","currency_id":"USD"}]','{"sentinel":true}')`, store); err != nil {
				t.Fatal(err)
			}
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.Method != http.MethodGet || r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
					t.Error("unexpected authorization or method")
					w.WriteHeader(403)
					return
				}
				switch r.URL.Path {
				case "/billing/integration/group/ML/order/details":
					fmt.Fprint(w, `{"results":[{"order_id":10001,"payment_info":[],"details":[]}]}`)
				case "/marketplace/shipments/80001/costs":
					fmt.Fprint(w, `{"currency_id":"USD","senders":[{"user_id":987654,"cost":0}],"receiver":{"cost":0}}`)
				case "/marketplace/orders/pack/20001":
					fmt.Fprint(w, `{"id":20001,"shipment":{"id":80001},"orders":[{"id":10001}]}`)
				case "/shipments/80001/orders":
					if r.Header.Get("X-New-Domain") != "true" {
						t.Error("missing official relationship header")
					}
					w.WriteHeader(code)
					if code == http.StatusOK {
						fmt.Fprint(w, `[{"order_id":"10001","pack_id":"20001","seller_id":987654}]`)
					}
				default:
					t.Error("unexpected provider endpoint")
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			s.apiURL = server.URL
			if _, err := s.ProbeBillingByPack(ctx, "00000000-0000-4000-8000-000000000099", store, OrderScope{Site: "MLB", Logistic: "remote"}, "20001"); err == nil || calls != 0 {
				t.Fatal("foreign owner reached provider")
			}
			report, err := s.ProbeBillingByPack(ctx, owner, store, OrderScope{Site: "MLB", Logistic: "remote"}, "20001")
			if code == http.StatusTooManyRequests {
				if err == nil {
					t.Fatal("rate limit accepted")
				}
			} else {
				if err != nil || !report.PackMembership.Matches || report.ShipmentMembership.HTTPStatus != code || report.MerchandiseUSD != "8.74" {
					t.Fatalf("probe report: %+v %v", report, err)
				}
				if (report.OfficialMembership == "single_order_verified") != (code == http.StatusOK) {
					t.Fatal("unknown membership promoted")
				}
			}
			var unchanged bool
			if err := db.QueryRow(ctx, `SELECT billing_snapshot='{"sentinel":true}'::jsonb FROM app.meli_orders WHERE store_id=$1 AND order_id='10001'`, store).Scan(&unchanged); err != nil || !unchanged {
				t.Fatalf("probe wrote financial cache: %v", err)
			}
			if calls != 4 {
				t.Fatalf("unexpected retries or requests: %d", calls)
			}
		})
	}
}
