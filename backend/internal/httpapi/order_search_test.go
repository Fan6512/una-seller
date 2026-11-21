package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/meli"
)

func TestOrderSearchAuthAndInputValidation(t *testing.T) {
	h:=HandlerWithCollections(&collections.Store{},stubVerifier{
		id:"6cbe41df-69cb-4594-9bad-a5473aa1a5fb",
	},&meli.Service{})
	const root="/api/v1/stores/all/orders/search"
	for _,tc:=range []struct{path string;method string;status int}{
		{root+"?site=*&logistic_type=*&q=yellow",http.MethodGet,http.StatusUnauthorized},
		{root, http.MethodGet, http.StatusBadRequest},
		{root+"?site=*&logistic_type=*&q=a",http.MethodGet,http.StatusBadRequest},
		{root+"?site=*&site=MLB&logistic_type=*",http.MethodGet,http.StatusBadRequest},
		{root+"?site=*&logistic_type=*&purchase=invalid",http.MethodGet,http.StatusBadRequest},
		{root+"?site=*&logistic_type=*&offset=-2",http.MethodGet,http.StatusBadRequest},
		{root+"?site=MLB&logistic_type=*&q=hello",http.MethodPost,http.StatusMethodNotAllowed},
	} {
		req:=httptest.NewRequest(tc.method,tc.path,nil)
		if tc.status!=http.StatusUnauthorized { req.Header.Set("Authorization","Bearer synthetic-session") }
		rec:=httptest.NewRecorder()
		h.ServeHTTP(rec,req)
		if rec.Code!=tc.status {
			t.Fatalf("%s %s status=%d want=%d body=%s",tc.method,tc.path,rec.Code,tc.status,rec.Body.String())
		}
		if tc.status==http.StatusMethodNotAllowed && rec.Header().Get("Allow")!="GET" {
			t.Fatalf("missing Allow: %s",rec.Header().Get("Allow"))
		}
	}
	if !meli.OrderSearchFilter{Store:"all",Site:"*",Logistic:"*",Query:"Test title"}.Valid() {
		t.Fatal("ordinary title search rejected")
	}
	for _,q:=range []string{"a",strings.Repeat("A",121)} {
		if (meli.OrderSearchFilter{Store:"all",Site:"*",Logistic:"*",Query:q}).Valid() {
			t.Fatalf("invalid query accepted %q",q)
		}
	}
}
