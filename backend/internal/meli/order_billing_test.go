package meli

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestBillingPreviewSanitizesAndDoesNotInventNet(t *testing.T) {
	raw:=[]byte(`{"results":[{"order_id":123456789,"payer_nickname":"SECRET_BUYER",
	"payment_info":[{"status":"approved","money_release_status":"released","base_amount_usd":25.19,
	"payer_id":999,"tax_details":[{"mov_financial_entity":"iva","original_amount_usd":4.03,"refunded_amount_usd":0}]}],
	"sale_fee":{"net":22},"details":[{"charge_info":{"detail_id":345,"detail_amount":0.53,
	"detail_type":"CHARGE","detail_sub_type":"CVMPI","debited_from_operation":"YES"},
	"currency_info":{"currency_id":"USD"},"sales_info":[{"payer_nickname":"SECRET_BUYER"}]}]}]}`)
	page,err:=parseBillingPreview(raw,[]string{"123456789","123456788"})
	if err!=nil || len(page.Items)!=2 || !page.Items[0].Available ||
		page.Items[0].Payments[0].TaxableBaseUSD!="25.19" ||
		page.Items[0].Charges[0].Amount!="0.53" ||
		page.Items[1].Available || page.NetIncomeUSD!=nil {
		t.Fatalf("unexpected sanitized billing page %+v %v",page,err)
	}
	encoded,_:=json.Marshal(page)
	if strings.Contains(string(encoded),"SECRET_BUYER") || strings.Contains(string(encoded),"payer_id") ||
		strings.Contains(string(encoded),"net_income_usd\":\"25.19") {
		t.Fatal("PII or fake net payout leaked")
	}
}
func TestBillingRequestLimitsAndMapping(t *testing.T) {
	path,err:=billingPath([]string{"11","22"},"987654")
	if err!=nil || !strings.Contains(path,"seller_id=987654") || !strings.Contains(path,"order_ids=11%2C22") {
		t.Fatalf("official billing path=%s, %v",path,err)
	}
	for _,ids:=range [][]string{{},{"1","1"},{"01"},{strings.Repeat("1",31)}} {
		if _,err=billingPath(ids,"987654");!errors.Is(err,ErrOrderBillingInvalid) {
			t.Fatalf("invalid batch accepted: %+v",ids)
		}
	}
	tooMany:=make([]string,61)
	for i:=range tooMany { tooMany[i]=strings.Repeat("1",i%30+1) }
	if _,err=billingPath(tooMany,"987654");!errors.Is(err,ErrOrderBillingInvalid) {
		t.Fatal("61 IDs accepted")
	}
	for _,bad:=range []string{
		`{"results":[{"order_id":999}]}`,`{"results":[{"order_id":11},{"order_id":11}]}`,
		`{"results":null}`,`{"results":[{"order_id":11,"details":[{"charge_info":{"detail_amount":"bad"}}]}]}`,
	} {
		if _,err=parseBillingPreview([]byte(bad),[]string{"11"});!errors.Is(err,ErrOrderBillingInvalid) {
			t.Fatalf("invalid or foreign bill accepted: %s: %v",bad,err)
		}
	}
}
func TestProfitFromVerifiedNetUSD(t *testing.T) {
	x,err:=EstimateProfitFromVerifiedNet("20","7","85","5")
	if err!=nil || x.NetCNY!="140.00" || x.ProfitCNY!="50.00" ||
		x.MarginPercent==nil || *x.MarginPercent!="35.7" {
		t.Fatalf("verified-net calculation %+v %v",x,err)
	}
	x,err=EstimateProfitFromVerifiedNet("0","7","85","5")
	if err!=nil || x.ProfitCNY!="-90.00" || x.MarginPercent!=nil {
		t.Fatalf("zero net %+v %v",x,err)
	}
	x,err=EstimateProfitFromVerifiedNet("-0.12","7","0","0")
	if err!=nil || x.ProfitCNY!="-0.84" || x.MarginPercent!=nil {
		t.Fatalf("refund net %+v %v",x,err)
	}
	for _,net:=range []string{"","NaN","5e3","1.333333333"," 30","100000000000000000"} {
		if _,err=EstimateProfitFromVerifiedNet(net,"7","1","0");!errors.Is(err,ErrProfitInputsInvalid) {
			t.Fatalf("invalid net accepted %q",net)
		}
	}
	if _,err=EstimateProfitFromVerifiedNet("20","0","5","0");err==nil { t.Fatal("zero fx accepted") }
}
