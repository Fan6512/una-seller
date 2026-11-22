package meli

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestBillingPreviewSanitizesAndDoesNotInventNet(t *testing.T) {
	raw := []byte(`{"results":[{"order_id":123456789,"payer_nickname":"SECRET_BUYER",
	"payment_info":[{"status":"approved","money_release_status":"released","base_amount_usd":25.19,
	"payer_id":999,"tax_details":[{"mov_financial_entity":"iva","original_amount_usd":4.03,"refunded_amount_usd":0}]}],
	"sale_fee":{"net":22},"details":[{"charge_info":{"detail_id":345,"detail_amount":0.53,
	"detail_type":"CHARGE","detail_sub_type":"CVMPI","debited_from_operation":"YES"},
	"currency_info":{"currency_id":"USD"},"sales_info":[{"payer_nickname":"SECRET_BUYER"}]}]}]}`)
	page, err := parseBillingPreview(raw, []string{"123456789", "123456788"})
	if err != nil || len(page.Items) != 2 || !page.Items[0].Available ||
		page.Items[0].Payments[0].TaxableBaseUSD != "25.19" ||
		page.Items[0].Charges[0].Amount != "0.53" ||
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

func TestOfficialBillingChecksOwnerAndMappingBeforeAndAfterRead(t *testing.T) {
	service, db, owner, store := setupRefreshTest(t)
	ctx, stop := context.WithTimeout(context.Background(), 10*time.Second)
	defer stop()
	if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET access_token_expires_at=now()+interval '1 hour' WHERE id=$1::uuid`, store); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO app.meli_orders
 (store_id,site_id,logistic_type,seller_id,order_id,status,currency_id,total_amount,date_created,last_updated,order_items)
 VALUES ($1::uuid,'MLB','remote','987654','10001','paid','USD',20,now(),now(),'[]')`, store); err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	var revoke atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.Method != http.MethodGet || r.URL.Path != "/billing/integration/group/ML/order/details" ||
			r.URL.Query().Get("seller_id") != "987654" || r.URL.Query().Get("order_ids") != "10001" ||
			r.Header.Get("Authorization") != "Bearer synthetic_old_access" {
			t.Error("unexpected fiscal request or mapping")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if revoke.Load() {
			if _, err := db.Exec(ctx, `UPDATE app.meli_store_connections SET marketplaces_status='unverified' WHERE id=$1::uuid`, store); err != nil {
				t.Error(err)
			}
		}
		_, _ = w.Write([]byte(`{"results":[{"order_id":10001,"payment_info":[{"status":"approved","money_release_status":"released"}]}]}`))
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
	page, err := service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if err != nil || len(page.Items) != 1 || page.NetIncomeUSD != nil {
		t.Fatalf("unexpected billing preview: %+v %v", page, err)
	}
	revoke.Store(true)
	page, err = service.PreviewOrderBilling(ctx, owner, store, scope, []string{"10001"})
	if !errors.Is(err, ErrOrderAuthorizationChanged) || len(page.Items) != 0 {
		t.Fatalf("revoked mapping leaked billing: %+v %v", page, err)
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
