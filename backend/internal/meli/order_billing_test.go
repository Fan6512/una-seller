package meli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestBillingPreviewSanitizesAndDoesNotInventNet(t *testing.T) {
	raw := []byte(`{"results":[{"order_id":123456789,"payer_nickname":"SECRET_BUYER",
	"payment_info":[{"payment_id":998877,"status":"approved","money_release_status":"released","base_amount_usd":25.19,
	"payer_id":999,"tax_details":[{"mov_financial_entity":"iva","original_amount_usd":4.03,"refunded_amount_usd":0}]}],
	"sale_fee":{"net":22},"details":[{"charge_info":{"detail_id":345,"detail_amount":0.53,
	"detail_type":"CHARGE","detail_sub_type":"CVMPI","debited_from_operation":"YES","transaction_detail":"SECRET_BUYER address"},
	"currency_info":{"currency_id":"USD"},"shipping_info":{"pack_id":"222333"},"sales_info":[{"order_id":123456789,"operation_id":998877,"transaction_amount_usd":29.19,"payer_nickname":"SECRET_BUYER"},{"order_id":123456788,"operation_id":998878}]}]}]}`)
	page, err := parseBillingPreview(raw, []string{"123456789", "123456788"})
	if err != nil || len(page.Items) != 2 || !page.Items[0].Available ||
		page.Items[0].Payments[0].TaxableBaseUSD != "25.19" ||
		page.Items[0].Charges[0].Amount != "0.53" || page.Items[0].Payments[0].PaymentID != "998877" ||
		page.Items[0].Charges[0].PackID != "222333" || len(page.Items[0].Charges[0].SalesRefs) != 2 ||
		page.Items[0].Charges[0].SalesRefs[0] != (BillingSaleRef{OrderID: "123456789", OperationID: "998877", TransactionUSD: "29.19"}) ||
		page.Items[0].Charges[0].SalesRefs[1].OrderID != "123456788" ||
		page.Items[1].Available || page.NetIncomeUSD != nil {
		t.Fatalf("unexpected sanitized billing page %+v %v", page, err)
	}
	encoded, _ := json.Marshal(page)
	if strings.Contains(string(encoded), "SECRET_BUYER") || strings.Contains(string(encoded), "payer_id") ||
		strings.Contains(string(encoded), "net_income_usd\":\"25.19") {
		t.Fatal("PII or fake net payout leaked")
	}
}
func TestBillingRequestLimitsAndMapping(t *testing.T) {
	path, err := billingPath([]string{"11", "22"}, "987654")
	if err != nil || !strings.Contains(path, "seller_id=987654") || !strings.Contains(path, "order_ids=11%2C22") {
		t.Fatalf("official billing path=%s, %v", path, err)
	}
	for _, ids := range [][]string{{}, {"1", "1"}, {"01"}, {strings.Repeat("1", 31)}} {
		if _, err = billingPath(ids, "987654"); !errors.Is(err, ErrOrderBillingInvalid) {
			t.Fatalf("invalid batch accepted: %+v", ids)
		}
	}
	tooMany := make([]string, 61)
	for i := range tooMany {
		tooMany[i] = strings.Repeat("1", i%30+1)
	}
	if _, err = billingPath(tooMany, "987654"); !errors.Is(err, ErrOrderBillingInvalid) {
		t.Fatal("61 IDs accepted")
	}
	for _, bad := range []string{
		`{"results":[{"order_id":999}]}`, `{"results":[{"order_id":11},{"order_id":11}]}`,
		`{"results":null}`, `{"results":[{"order_id":11,"details":[{"charge_info":{"detail_amount":"bad"}}]}]}`,
	} {
		if _, err = parseBillingPreview([]byte(bad), []string{"11"}); !errors.Is(err, ErrOrderBillingInvalid) {
			t.Fatalf("invalid or foreign bill accepted: %s: %v", bad, err)
		}
	}
}
func TestBillingGroupsRetainNativeCurrenciesAndDuplicateIdentity(t *testing.T) {
	raw := []byte(`{"results":[{"order_id":11,"payment_info":[{"status":"approved","money_release_status":"pending",
 "tax_details":[{"mov_financial_entity":"iva","original_amount_usd":0.3,"refunded_amount_usd":0.15}]}],
 "details":[
 {"charge_info":{"detail_id":1,"detail_type":"CHARGE","detail_sub_type":"COMMISSION","detail_amount":0.10000001},"currency_info":{"currency_id":"USD"}},
 {"charge_info":{"detail_id":1,"detail_type":"CHARGE","detail_sub_type":"COMMISSION","detail_amount":0.10000001},"currency_info":{"currency_id":"USD"}},
 {"charge_info":{"detail_id":2,"detail_type":"CHARGE","detail_sub_type":"COMMISSION","detail_amount":0.2},"currency_info":{"currency_id":"USD"}},
 {"charge_info":{"detail_id":3,"detail_type":"CHARGE","detail_sub_type":"COMMISSION","detail_amount":0.9},"currency_info":{"currency_id":"BRL"}},
 {"charge_info":{"detail_id":4,"detail_type":"BONUS","detail_sub_type":"COMMISSION","detail_amount":-0.10000001,"charge_bonified_id":1},"currency_info":{"currency_id":"USD"}}
 ]}]}`)
	page, err := parseBillingPreview(raw, []string{"11"})
	if err != nil || page.NetIncomeUSD != nil || len(page.Items[0].Charges) != 4 || len(page.Items[0].ChargeGroups) != 3 {
		t.Fatalf("failed to preserve original billing identity: %+v %v", page, err)
	}
	order := page.Items[0]
	if order.Payments[0].MoneyReleaseStatus != "pending" || order.Taxes[0].RefundedUSD != "0.15" ||
		order.Charges[3].ReversalOf != "1" || order.ChargeGroups[0].OriginalTotal != "0.30000001" ||
		order.ChargeGroups[1].Currency != "BRL" || order.ChargeGroups[1].OriginalTotal != "0.9" ||
		order.ChargeGroups[2].OriginalTotal != "-0.10000001" {
		t.Fatalf("native currency or adjustment changed: %+v", order)
	}
	conflicting := []byte(strings.Replace(string(raw), `"detail_id":2`, `"detail_id":1`, 1))
	if _, err = parseBillingPreview(conflicting, []string{"11"}); !errors.Is(err, ErrOrderBillingInvalid) {
		t.Fatal("conflicting charge identity accepted")
	}
}

func TestBillingShipmentCostsFailClosedOnMissingCurrencyOrAmbiguousSeller(t *testing.T) {
	counterpartyCosts := `{"currency_id":"USD","senders":[{"user_id":111,"cost":3.38},{"user_id":222,"cost":0}],
		"receiver":{"cost":10,"cost_details":[{"sender_id":222,"amount":2.28},{"sender_id":111,"amount":7.72}]}}`
	buyer, seller := parseBillingShipping([]byte(counterpartyCosts), "111")
	if buyer != "7.72" || seller != "3.38" {
		t.Fatalf("billing shipping selected wrong seller: buyer=%s seller=%s", buyer, seller)
	}
	for _, raw := range []string{
		strings.Replace(counterpartyCosts, "USD", "BRL", 1),
		strings.Replace(counterpartyCosts, "111", "222", 1),
		strings.Replace(counterpartyCosts, "222", "111", 1),
		strings.Replace(counterpartyCosts, `"cost":3.38`, `"different":3.38`, 1),
	} {
		if buyer, seller = parseBillingShipping([]byte(raw), "111"); buyer != "" || seller != "" {
			t.Fatalf("unconfirmed shipping costs escaped: buyer=%s seller=%s", buyer, seller)
		}
	}
	zero := []byte(`{"currency_id":"USD","senders":[{"user_id":111,"cost":0}],"receiver":{"cost":0}}`)
	if buyer, seller = parseBillingShipping(zero, "111"); buyer != "0" || seller != "0" {
		t.Fatalf("explicit zero shipping costs lost: buyer=%s seller=%s", buyer, seller)
	}
}

func TestOfficialBillingChecksOwnerAndMappingBeforeAndAfterRead(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,order_items)
 VALUES ($1::uuid,'MLB','remote','987654','10001','paid','USD',20,now(),now(),'80001','[]')`, store); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var revoke atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method == http.MethodGet && r.URL.Path == "/marketplace/shipments/80001/costs" {
			if r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
				t.Error("shipping query did not use scoped store credential")
			}
			if revoke.Load() {
				if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces_status='unverified' WHERE id=$1::uuid`, store); err != nil {
					t.Error(err)
				}
			}
			_, _ = w.Write([]byte(`{"currency_id":"USD","gross_amount":11.09,"senders":[{"user_id":987654,"cost":3.38}],
			 "receiver":{"user_id":998877,"cost":7.72,"cost_details":[{"sender_id":987654,"amount":7.72}]}}`))
			return
		}
		if r.Method != http.MethodGet || r.URL.Path != "/billing/integration/group/ML/order/details" ||
			r.URL.Query().Get("seller_id") != "987654" || r.URL.Query().Get("order_ids") != "10001" ||
			r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("unexpected fiscal request or mapping")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		_, _ = w.Write([]byte(`{"results":[{"order_id":10001,"payment_info":[{"status":"approved","money_release_status":"pending"}],"details":[
		  {"charge_info":{"detail_id":91,"detail_amount":11.10,"detail_type":"CHARGE","detail_sub_type":"CCSI","debited_from_operation":"YES"},
		   "currency_info":{"currency_id":"USD"},"shipping_info":{"shipping_id":80001}},
		  {"charge_info":{"detail_id":92,"detail_amount":-1.12,"detail_type":"BONUS","detail_sub_type":"CCSI","charge_bonified_id":91},
		   "currency_info":{"currency_id":"USD"},"shipping_info":{"shipping_id":80001}}
		]}]}`))
	}))
	defer server.Close()
	service.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	_, err := service.PreviewOrderBilling(ctx, "00000000-0000-4000-8000-000000000099", store, scope, []string{"10001"})
	if !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("foreign owner: %v", err)
	}
	_, err = service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10002"})
	if !errors.Is(err, ErrOrderNotFound) || calls.Load() != 0 {
		t.Fatalf("foreign order reached platform: %v", err)
	}
	cache, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || len(cache.Items) != 1 || cache.Items[0].Available || cache.Items[0].Reconciliation != "not_cached" || calls.Load() != 0 {
		t.Fatalf("empty cache contacted platform or returned a bill: %+v %v", cache, err)
	}
	page, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if err != nil || len(page.Items) != 1 || page.NetIncomeUSD != nil || calls.Load() != 2 {
		t.Fatalf("unexpected billing preview: %+v %v", page, err)
	}
	item := page.Items[0]
	if item.Shipping == nil || item.Shipping.BuyerCostUSD != "7.72" || item.Shipping.SellerChargeUSD != "11.10" ||
		item.Shipping.SellerNetCostUSD != "3.38" || item.Shipping.SellerChargeDetailID != "91" ||
		len(item.Charges) != 2 || item.Charges[1].ReversalOf != "91" || item.Charges[1].Amount != "-1.12" {
		t.Fatalf("source amounts or reversal were misclassified: %+v", item)
	}
	encoded, _ := json.Marshal(page)
	if strings.Contains(string(encoded), "998877") {
		t.Fatal("shipping receiver identity escaped sanitization")
	}
	cache, err = service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || len(cache.Items) != 1 || cache.NetIncomeUSD != nil || calls.Load() != 2 ||
		cache.Items[0].CheckedAt == nil || cache.Items[0].Shipping == nil ||
		cache.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("successful bill not persisted or cache contacted upstream: %+v %v", cache, err)
	}
	if _, err := service.GetCachedOrderBilling(ctx, "00000000-0000-4000-8000-000000000099", store, scope, "10001"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("foreign owner cached billing: %v", err)
	}
	revoke.Store(true)
	page, err = service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if !errors.Is(err, ErrOrderAuthorizationChanged) || len(page.Items) != 0 {
		t.Fatalf("revoked mapping leaked billing: %+v %v", page, err)
	}
	if _, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001"); !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("revoked mapping cached billing: %v", err)
	}
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces_status='verified' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	restored, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || len(restored.Items) != 1 || restored.Items[0].CheckedAt == nil ||
		!restored.Items[0].CheckedAt.Equal(*cache.Items[0].CheckedAt) {
		t.Fatalf("failed refresh did not preserve successful snapshot: %+v %v", restored, err)
	}
}

func TestProfitFromVerifiedNetUSD(t *testing.T) {
	asOf := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	fx := ProfitFXRate{CNYPerUSD: "7", Source: "test bank receipt", AsOf: asOf}
	net := func(amount string) VerifiedUSDNet {
		return VerifiedUSDNet{Amount: amount, Source: "test settlement", SettledAt: asOf}
	}
	x, err := EstimateProfitFromVerifiedNet(net("20"), fx, "85", "5")
	if err != nil || x.NetUSD != "20.00" || x.ProfitUSD != "7.14" || x.FX != fx ||
		x.MarginPercent == nil || *x.MarginPercent != "35.7" {
		t.Fatalf("verified-net calculation %+v %v", x, err)
	}
	x, err = EstimateProfitFromVerifiedNet(net("0"), fx, "85", "5")
	if err != nil || x.ProfitUSD != "-12.86" || x.MarginPercent != nil {
		t.Fatalf("zero net %+v %v", x, err)
	}
	x, err = EstimateProfitFromVerifiedNet(net("-0.125"), fx, "0", "0")
	if err != nil || x.ProfitUSD != "-0.13" || x.MarginPercent != nil {
		t.Fatalf("refund net %+v %v", x, err)
	}
	for _, bad := range []string{"", "NaN", "5e3", "1.333333333", " 30", "100000000000000000"} {
		if _, err = EstimateProfitFromVerifiedNet(net(bad), fx, "1", "0"); !errors.Is(err, ErrProfitInputsInvalid) {
			t.Fatalf("invalid net accepted %q", bad)
		}
	}
	for _, rate := range []ProfitFXRate{
		{CNYPerUSD: "0", Source: "bank receipt", AsOf: asOf},
		{CNYPerUSD: "7", AsOf: asOf},
		{CNYPerUSD: "7", Source: "bank receipt"},
	} {
		if _, err := EstimateProfitFromVerifiedNet(net("20"), rate, "5", "0"); err == nil {
			t.Fatalf("unverified rate accepted: %+v", rate)
		}
	}
	if _, err := EstimateProfitFromVerifiedNet(net("20"), fx, "", "0"); err == nil {
		t.Fatal("missing cost accepted")
	}
	if _, err := EstimateProfitFromVerifiedNet(VerifiedUSDNet{Amount: "20"}, fx, "5", "0"); err == nil {
		t.Fatal("unverified receipt accepted")
	}
}

func TestOrderReadResourceUsesFixedLabels(t *testing.T) {
	for path, want := range map[string]string{
		"/billing/integration/group/ML/order/details?order_ids=11": "billing",
		"/marketplace/shipments/11/costs":                          "shipment",
		"/marketplace/orders/search?offset=20":                     "order_search",
		"/marketplace/orders/11":                                   "order_detail",
		"/unrecognized/secret":                                     "other",
	} {
		if got := orderReadResource(path); got != want {
			t.Fatalf("resource label=%q want=%q", got, want)
		}
	}
}

func TestBillingWrappedRateLimitUsesExistingOrderCooldown(t *testing.T) {
	for _, tc := range []struct {
		path string
		body string
		want int
	}{
		{"/billing/integration/group/ML/order/details?order_ids=11", "[REPORTING_CORE_API][ML_DETAILS_BY_ORDER][429]", http.StatusTooManyRequests},
		{"/billing/integration/group/ML/order/details?order_ids=11", "ordinary bad request [429]", http.StatusBadRequest},
		{"/marketplace/orders/11", "[REPORTING_CORE_API][ML_DETAILS_BY_ORDER][429]", http.StatusBadRequest},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(tc.body))
		}))
		s := &Service{apiURL: server.URL, client: server.Client()}
		var raw json.RawMessage
		status, err := s.orderGET(context.Background(), "synthetic", tc.path, &raw)
		server.Close()
		if status != tc.want {
			t.Fatalf("status=%d want=%d err=%v", status, tc.want, err)
		}
		if tc.want == http.StatusTooManyRequests {
			if OrderErrorCode(err) != "order_rate_limited" || time.Until(s.orderRetryUntil) < 50*time.Second {
				t.Fatalf("wrapped Billing limit not classified or cooled down: %v", err)
			}
		}
	}
}

func insertBillingTestOrder(t *testing.T, ctx context.Context, db *pgxpool.Pool, store, seller, shipment string) {
	t.Helper()
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,shipment_id,order_items)
 VALUES ($1::uuid,'MLB','remote',$2,'10001','paid','USD',20,now(),now(),$3,'[]')`, store, seller, shipment); err != nil {
		t.Fatal(err)
	}
}

func billingTestSeller(t *testing.T, ctx context.Context, db *pgxpool.Pool, store string) string {
	t.Helper()
	var seller string
	if err := db.QueryRow(ctx, `SELECT seller_id FROM app.meli_store_connections WHERE id=$1::uuid`, store).Scan(&seller); err != nil {
		t.Fatal(err)
	}
	return seller
}

func TestCachedBillingIsScopedToStoreAndShipment(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	defer stop()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	otherAccess, err := service.cipher.Seal("synthetic_other_access", tokenAAD(owner, "555001", "access"))
	if err != nil {
		t.Fatal(err)
	}
	otherRefresh, err := service.cipher.Seal("synthetic_other_refresh", tokenAAD(owner, "555001", "refresh"))
	if err != nil {
		t.Fatal(err)
	}
	var otherStore string
	if err := db.QueryRow(ctx, `INSERT INTO app.meli_store_connections
 (owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at,marketplaces,marketplaces_status)
 VALUES ($1::uuid,'555001',$2,$3,now()+interval '1 hour',
 '[{"user_id":555002,"site_id":"MLB","logistic_type":"remote"}]','verified') RETURNING id::text`,
		owner, otherAccess, otherRefresh).Scan(&otherStore); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = db.Exec(c, "DELETE FROM app.meli_store_connections WHERE id=$1::uuid", otherStore)
	})
	insertBillingTestOrder(t, ctx, db, store, "987654", "80001")
	insertBillingTestOrder(t, ctx, db, otherStore, "555002", "80001")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/order/details"):
			_, _ = w.Write([]byte(`{"results":[{"order_id":10001,"payment_info":[{"status":"approved","money_release_status":"pending"}],
			 "details":[{"charge_info":{"detail_id":91,"detail_amount":11.10,"detail_type":"CHARGE","detail_sub_type":"CCSI","debited_from_operation":"YES"},
			 "currency_info":{"currency_id":"USD"},"shipping_info":{"shipping_id":80001}}]}]}`))
		case strings.HasSuffix(r.URL.Path, "/costs") && strings.Contains(r.URL.Path, "/80002/"):
			_, _ = w.Write([]byte(`{"currency_id":"USD","senders":[{"user_id":987654,"cost":4.00}],"receiver":{"cost":7.72}}`))
		case strings.HasSuffix(r.URL.Path, "/costs"):
			_, _ = w.Write([]byte(`{"currency_id":"USD","senders":[{"user_id":987654,"cost":3.38}],"receiver":{"cost":7.72}}`))
		default:
			t.Errorf("unexpected upstream path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	service.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}

	page, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Shipping == nil || page.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("initial official bill: %+v %v", page, err)
	}
	cached, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || len(cached.Items) != 1 || !cached.Items[0].Available || cached.Items[0].CheckedAt == nil ||
		cached.Items[0].Shipping == nil || cached.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("store snapshot: %+v %v", cached, err)
	}
	// The same order number in another store must not read this snapshot.
	other, err := service.GetCachedOrderBilling(ctx, owner, otherStore, scope, "10001")
	if err != nil || len(other.Items) != 1 || other.Items[0].Available || other.Items[0].Reconciliation != "not_cached" ||
		other.Items[0].CheckedAt != nil {
		t.Fatalf("cross-store snapshot leak: %+v %v", other, err)
	}
	// A changed shipment invalidates the old snapshot instead of reusing its costs.
	if _, err := db.Exec(ctx, `UPDATE app.meli_orders SET shipment_id='80002' WHERE store_id=$1::uuid AND order_id='10001'`, store); err != nil {
		t.Fatal(err)
	}
	changed, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || len(changed.Items) != 1 || changed.Items[0].Available || changed.Items[0].Reconciliation != "not_cached" {
		t.Fatalf("stale shipment snapshot served: %+v %v", changed, err)
	}
	page, err = service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if err != nil || len(page.Items) != 1 || page.Items[0].Shipping == nil || page.Items[0].Shipping.SellerNetCostUSD != "4.00" {
		t.Fatalf("official bill after shipment change: %+v %v", page, err)
	}
	rebound, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || len(rebound.Items) != 1 || !rebound.Items[0].Available || rebound.Items[0].Shipping == nil ||
		rebound.Items[0].Shipping.SellerNetCostUSD != "4.00" || rebound.Items[0].CheckedAt == nil {
		t.Fatalf("snapshot not rebound to current shipment: %+v %v", rebound, err)
	}
}

func TestLateBillingResponseKeepsNewestSnapshot(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	insertBillingTestOrder(t, ctx, db, store, "987654", "80001")
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	started := time.Now().UTC().Add(-time.Minute)
	newest := BillingOrderPreview{OrderID: "10001", Available: true, Reconciliation: "review_required",
		Payments: []BillingPayment{}, Taxes: []BillingTax{},
		Charges: []BillingCharge{{DetailID: "91", Type: "CHARGE", Amount: "2.00", Currency: "USD"}}}
	stale := newest
	stale.Charges = []BillingCharge{{DetailID: "91", Type: "CHARGE", Amount: "1.00", Currency: "USD"}}
	parent := billingTestSeller(t, ctx, db, store)
	saved, err := service.saveOrderBillingSnapshot(ctx, owner, store, scope, "987654", parent, "80001", started.Add(time.Second), newest)
	if err != nil || saved.CheckedAt == nil {
		t.Fatalf("newest snapshot not saved: %+v %v", saved, err)
	}
	late, err := service.saveOrderBillingSnapshot(ctx, owner, store, scope, "987654", parent, "80001", started, stale)
	if err != nil || late.Charges[0].Amount != "2.00" || !late.CheckedAt.Equal(*saved.CheckedAt) {
		t.Fatalf("late response overwrote newer snapshot: %+v %v", late, err)
	}
	cached, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || cached.Items[0].Charges[0].Amount != "2.00" {
		t.Fatalf("stored snapshot changed by late response: %+v %v", cached, err)
	}
	if _, err = service.saveOrderBillingSnapshot(ctx, "00000000-0000-4000-8000-000000000099", store, scope,
		"987654", parent, "80001", started.Add(2*time.Second), stale); !errors.Is(err, ErrOrderAuthorizationChanged) {
		t.Fatalf("foreign owner wrote snapshot: %v", err)
	}
}

func TestPartialBillingResponsePreservesSnapshot(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	defer stop()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	insertBillingTestOrder(t, ctx, db, store, "987654", "80001")
	var costsDown, billingMissing atomic.Bool
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch {
		case strings.HasSuffix(r.URL.Path, "/costs"):
			if costsDown.Load() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = w.Write([]byte(`{"currency_id":"USD","senders":[{"user_id":987654,"cost":3.38}],"receiver":{"cost":7.72}}`))
		case strings.HasSuffix(r.URL.Path, "/order/details"):
			if billingMissing.Load() {
				_, _ = w.Write([]byte(`{"results":[]}`))
				return
			}
			_, _ = w.Write([]byte(`{"results":[{"order_id":10001,"payment_info":[],"details":[
			 {"charge_info":{"detail_id":91,"detail_amount":11.10,"detail_type":"CHARGE","detail_sub_type":"CCSI"},
			  "currency_info":{"currency_id":"USD"},"shipping_info":{"shipping_id":80001}}]}]}`))
		default:
			t.Errorf("unexpected upstream path %s", r.URL.Path)
		}
	}))
	defer server.Close()
	service.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	if page, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"}); err != nil ||
		page.Items[0].Shipping == nil || page.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("initial bill: %+v %v", page, err)
	}
	saved, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || !saved.Items[0].Available {
		t.Fatalf("initial snapshot: %+v %v", saved, err)
	}
	checked := *saved.Items[0].CheckedAt

	costsDown.Store(true)
	page, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if err != nil || !page.Items[0].Available || page.Items[0].Shipping == nil || page.Items[0].Shipping.SellerNetCostUSD != "" {
		t.Fatalf("partial shipping cost response: %+v %v", page, err)
	}
	afterCosts, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || !afterCosts.Items[0].CheckedAt.Equal(checked) || afterCosts.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("partial response replaced snapshot: %+v %v", afterCosts, err)
	}

	costsDown.Store(false)
	billingMissing.Store(true)
	page, err = service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if err != nil || page.Items[0].Available {
		t.Fatalf("missing billing response: %+v %v", page, err)
	}
	afterMissing, err := service.GetCachedOrderBilling(ctx, owner, store, scope, "10001")
	if err != nil || !afterMissing.Items[0].CheckedAt.Equal(checked) || afterMissing.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("missing response replaced snapshot: %+v %v", afterMissing, err)
	}
	if calls.Load() < 5 {
		t.Fatalf("partial refreshes skipped the platform: calls=%d", calls.Load())
	}
}

func TestConcurrentBillingPreviewIsSingleFlight(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	defer stop()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	insertBillingTestOrder(t, ctx, db, store, "987654", "80001")
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if strings.HasSuffix(r.URL.Path, "/order/details") {
			entered <- struct{}{}
			<-release
			_, _ = w.Write([]byte(`{"results":[{"order_id":10001,"payment_info":[],"details":[]}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"currency_id":"USD","senders":[{"user_id":987654,"cost":3.38}],"receiver":{"cost":7.72}}`))
	}))
	defer server.Close()
	service.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	done := make(chan error, 1)
	go func() {
		_, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
		done <- err
	}()
	<-entered
	if _, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"}); !errors.Is(err, ErrOrderBusy) {
		t.Fatalf("concurrent refresh not rejected: %v", err)
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("concurrent refresh reached platform %d times", calls.Load())
	}
}

func TestBillingRateLimitSetsCooldown(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	defer stop()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	insertBillingTestOrder(t, ctx, db, store, "987654", "80001")
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"too many requests"}`))
	}))
	defer server.Close()
	service.apiURL = server.URL
	scope := OrderScope{Site: "MLB", Logistic: "remote"}
	var cooldown *OrderRateLimitError
	if _, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"}); !errors.As(err, &cooldown) {
		t.Fatalf("upstream 429 not surfaced as cooldown: %v", err)
	}
	retry, err := strconv.Atoi(cooldown.RetryAfter())
	if err != nil || retry < 1 || retry > 120 {
		t.Fatalf("retry-after out of range: %q", cooldown.RetryAfter())
	}
	if _, err = service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"}); !errors.As(err, &cooldown) {
		t.Fatalf("local cooldown not enforced: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("cooldown still reached platform: calls=%d", calls.Load())
	}
}

func TestBillingSnapshotPostgresRestartPersistence(t *testing.T) {
	phase := os.Getenv("TEST_RESTART_PHASE")
	if phase == "" {
		t.Skip("executed separately with a real PostgreSQL restart")
	}
	if phase != "before" && phase != "after" {
		t.Fatalf("invalid TEST_RESTART_PHASE %q", phase)
	}
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL is required for PostgreSQL restart test")
	}
	ctx, stop := context.WithTimeout(context.Background(), 20*time.Second)
	defer stop()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	const owner = "3b1c0000-0000-4000-8000-0000000000a7"
	const store = "3b1c0000-0000-4000-8000-0000000000a8"
	if phase == "before" {
		if err := migrations.Apply(ctx, db); err != nil {
			t.Fatal(err)
		}
		service, err := NewService(db, Config{ClientID: "synthetic", ClientSecret: "synthetic",
			RedirectURI:        "https://example.invalid" + CallbackPath,
			TokenEncryptionKey: base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))})
		if err != nil {
			t.Fatal(err)
		}
		access, err := service.cipher.Seal("synthetic_restart_access", tokenAAD(owner, "900777", "access"))
		if err != nil {
			t.Fatal(err)
		}
		refresh, err := service.cipher.Seal("synthetic_restart_refresh", tokenAAD(owner, "900777", "refresh"))
		if err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `DELETE FROM app.meli_store_connections WHERE id=$1::uuid`, store); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO app.users(id) VALUES($1::uuid) ON CONFLICT DO NOTHING`, owner); err != nil {
			t.Fatal(err)
		}
		if _, err = db.Exec(ctx, `INSERT INTO app.meli_store_connections
 (id,owner_user_id,seller_id,access_token_ciphertext,refresh_token_ciphertext,access_token_expires_at,marketplaces,marketplaces_status)
 VALUES ($1::uuid,$2::uuid,'900777',$3,$4,now()+interval '1 hour',
 '[{"user_id":900778,"site_id":"MLB","logistic_type":"remote"}]','verified')`, store, owner, access, refresh); err != nil {
			t.Fatal(err)
		}
		insertBillingTestOrder(t, ctx, db, store, "900778", "80001")
		saved, err := service.saveOrderBillingSnapshot(ctx, owner, store, OrderScope{Site: "MLB", Logistic: "remote"},
			"900778", "900777", "80001", time.Now().UTC(), BillingOrderPreview{OrderID: "10001", Available: true,
				Reconciliation: "review_required", Payments: []BillingPayment{}, Taxes: []BillingTax{},
				Charges:  []BillingCharge{{DetailID: "91", Type: "CHARGE", Amount: "11.10", Currency: "USD"}},
				Shipping: &BillingShipping{SellerNetCostUSD: "3.38"}})
		if err != nil || saved.CheckedAt == nil {
			t.Fatalf("restart fixture not saved: %+v %v", saved, err)
		}
		t.Logf("billing snapshot persisted before restart: checked_at=%s", saved.CheckedAt.Format(time.RFC3339Nano))
		return
	}
	service, err := NewService(db, Config{ClientID: "synthetic", ClientSecret: "synthetic",
		RedirectURI:        "https://example.invalid" + CallbackPath,
		TokenEncryptionKey: base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))})
	if err != nil {
		t.Fatal(err)
	}
	cached, err := service.GetCachedOrderBilling(ctx, owner, store, OrderScope{Site: "MLB", Logistic: "remote"}, "10001")
	if err != nil || len(cached.Items) != 1 || !cached.Items[0].Available || cached.Items[0].CheckedAt == nil ||
		cached.Items[0].Shipping == nil || cached.Items[0].Shipping.SellerNetCostUSD != "3.38" {
		t.Fatalf("snapshot missing after restart: %+v %v", cached, err)
	}
	if _, err := db.Exec(ctx, `DELETE FROM app.meli_store_connections WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `DELETE FROM app.users WHERE id=$1::uuid`, owner); err != nil {
		t.Fatal(err)
	}
	t.Logf("billing snapshot survived PostgreSQL restart: checked_at=%s", cached.Items[0].CheckedAt.Format(time.RFC3339Nano))
}
