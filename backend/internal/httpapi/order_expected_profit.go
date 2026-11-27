package httpapi

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func (a storeAPI) orderExpectedProfit(w http.ResponseWriter, r *http.Request, store, orderID string) {
	w.Header().Set("Cache-Control","no-store")
	if r.Method!=http.MethodGet && r.Method!=http.MethodPost {
		w.Header().Set("Allow","GET, POST")
		writeJSON(w,http.StatusMethodNotAllowed,map[string]string{"error":"order_method_not_allowed"})
		return
	}
	owner,ok:=a.requireOwner(w,r)
	if !ok { return }
	q:=r.URL.Query()
	if q.Get("site")=="" || q.Get("logistic_type")=="" || len(q["site"])!=1 || len(q["logistic_type"])!=1 {
		writeJSON(w,http.StatusBadRequest,map[string]string{"error":"order_scope_required"})
		return
	}
	scope:=meli.OrderScope{Site:q.Get("site"),Logistic:q.Get("logistic_type")}
	var out meli.OrderExpectedProfit
	var err error
	if r.Method==http.MethodGet {
		out,err=a.service.GetExpectedOrderProfit(r.Context(),owner,store,scope,orderID)
	} else {
		var input struct {
			Revision string `json:"revision"`
			NetUSD string `json:"net_usd"`
			ExtraCNY string `json:"extra_cny"`
			Clear bool `json:"clear"`
			Confirmed bool `json:"confirmed"`
		}
		decoder:=json.NewDecoder(http.MaxBytesReader(w,r.Body,1024))
		decoder.DisallowUnknownFields()
		if decoder.Decode(&input)!=nil || decoder.Decode(new(any))!=io.EOF || !input.Confirmed {
			writeJSON(w,http.StatusBadRequest,map[string]string{"error":"order_billing_invalid"})
			return
		}
		out,err=a.service.ConfirmExpectedOrderProfit(r.Context(),owner,store,scope,orderID,input.Revision,input.NetUSD,input.ExtraCNY,input.Clear)
	}
	if err!=nil {
		writeOrderBillingError(w,err)
		return
	}
	writeJSON(w,http.StatusOK,out)
}
