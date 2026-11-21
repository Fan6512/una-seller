package httpapi

import (
  "encoding/json"
  "errors"
  "io"
  "net/http"
  "strconv"
  "strings"

  "github.com/Fan6512/una-seller/backend/internal/collections"
)

func (a collectionAPI) ownProducts(w http.ResponseWriter,r *http.Request) {
  if r.Method!=http.MethodGet {
    w.Header().Set("Allow","GET")
    writeError(w,http.StatusMethodNotAllowed,"method not allowed")
    return
  }
  owner,ok:=a.owner(w,r);if !ok{return}
  q:=strings.TrimSpace(r.URL.Query().Get("q"))
  limit,offset:=30,0
  var err error
  if raw:=r.URL.Query().Get("limit");raw!="" {limit,err=strconv.Atoi(raw);if err!=nil {writeError(w,400,"invalid limit");return}}
  if raw:=r.URL.Query().Get("offset");raw!="" {offset,err=strconv.Atoi(raw);if err!=nil {writeError(w,400,"invalid offset");return}}
  if len(q)>200||limit<1||limit>100||offset<0||offset>100000 {writeError(w,400,"invalid query");return}
  items,err:=a.store.ListOwnProducts(r.Context(),owner,q,limit,offset)
  if err!=nil {writeError(w,500,"product box unavailable");return}
  w.Header().Set("Cache-Control","no-store")
  writeJSON(w,200,map[string]any{"items":items,"limit":limit,"offset":offset})
}

func (a collectionAPI) ownProductDetail(w http.ResponseWriter,r *http.Request) {
  owner,ok:=a.owner(w,r);if !ok{return}
  id:=strings.TrimPrefix(r.URL.Path,"/api/v1/products/")
  if !collections.ValidUserID(id) {writeError(w,404,"product not found");return}
  switch r.Method {
  case http.MethodGet:
    p,err:=a.store.GetOwnProduct(r.Context(),owner,id)
    if errors.Is(err,collections.ErrNotFound){writeError(w,404,"product not found");return}
    if err!=nil {writeError(w,500,"product unavailable");return}
    w.Header().Set("Cache-Control","no-store")
    writeJSON(w,200,p)
  case http.MethodPut:
    if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")),"application/json") {
      writeError(w,415,"expected application/json");return
    }
    r.Body=http.MaxBytesReader(w,r.Body,256<<10)
    dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
    var input collections.OwnProductEdit
    if err:=dec.Decode(&input);err!=nil {writeError(w,400,"invalid product edit JSON");return}
    if err:=dec.Decode(new(any));err!=io.EOF {writeError(w,400,"trailing product edit JSON");return}
    if err:=input.Validate();err!=nil {writeError(w,400,"invalid product edit fields");return}
    p,err:=a.store.UpdateOwnProduct(r.Context(),owner,id,input)
    if errors.Is(err,collections.ErrNotFound){writeError(w,404,"product not found");return}
    if errors.Is(err,collections.ErrDraftConflict){writeError(w,409,"product changed; reload before saving");return}
    if errors.Is(err,collections.ErrReadback){
      writeJSON(w,http.StatusAccepted,map[string]any{"result":"saved_unverified","id":p.ID,"revision":p.Revision});return
    }
    if err!=nil {writeError(w,500,"product not saved");return}
    w.Header().Set("Cache-Control","no-store")
    writeJSON(w,200,map[string]any{"product":p})
  default:
    w.Header().Set("Allow","GET, PUT")
    writeError(w,405,"method not allowed")
  }
}

// Explicit desktop-only promotion. A missing/incorrect draft revision
// receives 409 rather than copying data with unknown saved state.
func (a collectionAPI) approveCollection(w http.ResponseWriter,r *http.Request,owner,id string) {
  if origin:=r.Header.Get("Origin");origin!=""&&!allowedAppOrigin(origin){
    writeError(w,403,"origin not allowed");return
  }
  if r.Method!=http.MethodPost {
    w.Header().Set("Allow","POST")
    writeError(w,405,"method not allowed");return
  }
  if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")),"application/json") {
    writeError(w,415,"expected application/json");return
  }
  r.Body=http.MaxBytesReader(w,r.Body,4096)
  dec:=json.NewDecoder(r.Body);dec.DisallowUnknownFields()
  var input struct{ ExpectedRevision *int64 `json:"expected_revision"` }
  if err:=dec.Decode(&input);err!=nil {writeError(w,400,"invalid approval JSON");return}
  if err:=dec.Decode(new(any));err!=io.EOF {writeError(w,400,"trailing approval JSON");return}
  if input.ExpectedRevision==nil||*input.ExpectedRevision<0 {
    writeError(w,400,"expected_revision required");return
  }
  product,created,err:=a.store.Promote(r.Context(),owner,id,*input.ExpectedRevision)
  if errors.Is(err,collections.ErrNotFound){writeError(w,404,"collection not found");return}
  if errors.Is(err,collections.ErrDraftConflict){writeError(w,409,"working draft has changed; reload before approval");return}
  if errors.Is(err,collections.ErrReadback){
    writeJSON(w,http.StatusAccepted,map[string]any{"result":"saved_unverified","id":product.ID});return
  }
  if err!=nil {writeError(w,500,"approval state unknown; reread product box");return}
  w.Header().Set("Cache-Control","no-store")
  if created {writeJSON(w,201,map[string]any{"result":"created","product":product})
  } else {writeJSON(w,200,map[string]any{"result":"already_approved","product":product})}
}
