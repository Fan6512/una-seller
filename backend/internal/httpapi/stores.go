package httpapi

import (
 "errors"
 "net/http"
 "strings"

 "github.com/Fan6512/una-seller/backend/internal/meli"
)

type storeAPI struct {
 auth collectionAPI
 service *meli.Service
}

func(a storeAPI) requireOwner(w http.ResponseWriter,r *http.Request)(string,bool){
 owner,ok:=a.auth.owner(w,r)
 if !ok{return "",false}
 if a.service==nil{writeError(w,http.StatusServiceUnavailable,"Mercado Libre OAuth is not configured");return "",false}
 return owner,true
}

func(a storeAPI) stores(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodGet{w.Header().Set("Allow","GET");writeError(w,405,"method not allowed");return}
 owner,ok:=a.requireOwner(w,r);if !ok{return}
 items,err:=a.service.List(r.Context(),owner)
 if err!=nil{writeError(w,500,"stores unavailable");return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]any{"items":items})
}

func(a storeAPI) start(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");writeError(w,405,"method not allowed");return}
 owner,ok:=a.requireOwner(w,r);if !ok{return}
 link,err:=a.service.Start(r.Context(),owner)
 if err!=nil{writeError(w,503,"could not create authorization");return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]string{"authorization_url":link})
}

func(a storeAPI) disconnect(w http.ResponseWriter,r *http.Request){
 if r.Method!=http.MethodPost{w.Header().Set("Allow","POST");writeError(w,405,"method not allowed");return}
 owner,ok:=a.requireOwner(w,r);if !ok{return}
 part:=strings.TrimPrefix(r.URL.Path,"/api/v1/stores/")
 if !strings.HasSuffix(part,"/disconnect"){writeError(w,404,"store not found");return}
 id:=strings.TrimSuffix(part,"/disconnect")
 if strings.Contains(id,"/"){writeError(w,404,"store not found");return}
 err:=a.service.Disconnect(r.Context(),owner,id)
 if errors.Is(err,meli.ErrStoreNotFound){writeError(w,404,"store not found");return}
 if err!=nil{writeError(w,500,"could not disconnect store");return}
 w.Header().Set("Cache-Control","no-store")
 writeJSON(w,200,map[string]string{"status":"disconnected"})
}

// The callback is intentionally public. Its high-entropy, single-use state
// binds it to the authenticated owner who initiated the flow.
func(a storeAPI) callback(w http.ResponseWriter,r *http.Request){
 w.Header().Set("Cache-Control","no-store")
 w.Header().Set("Referrer-Policy","no-referrer")
 w.Header().Set("X-Content-Type-Options","nosniff")
 w.Header().Set("Content-Security-Policy","default-src 'none'")
 w.Header().Set("Content-Type","text/html; charset=utf-8")
 if r.Method!=http.MethodGet{w.Header().Set("Allow","GET");http.Error(w,"method not allowed",405);return}
 if a.service==nil{http.Error(w,"OAuth not configured",503);return}
 q:=r.URL.Query()
 if len(q["state"])!=1 || len(q["code"])>1 || len(q["error"])>1 || len(q.Get("code"))>4096 {
  http.Error(w,"invalid authorization callback",400);return
 }
 _,err:=a.service.Complete(r.Context(),q.Get("state"),q.Get("code"),q.Get("error"))
 if err!=nil{
  status:=http.StatusBadRequest
  if errors.Is(err,meli.ErrSellerOwned){status=http.StatusConflict}
  // Never put the provider's error text, access token, code or state in HTML.
  w.WriteHeader(status)
  _,_=w.Write([]byte("<!doctype html><html lang='zh'><meta charset='utf-8'><title>授权未完成</title><h2>授权未完成</h2><p>请关闭此页面，在 Una Seller 店铺管理中重新授权。</p></html>"))
  return
 }
 w.WriteHeader(http.StatusOK)
 _,_=w.Write([]byte("<!doctype html><html lang='zh'><meta charset='utf-8'><title>授权成功</title><h2>美客多店铺授权成功</h2><p>可以关闭此页面，返回 Una Seller 点击刷新店铺。</p></html>"))
}
