import { useEffect, useState } from "react";
import ListingFieldsEditor from "./ListingFieldsEditor";
import { initialListingFields,listingTextCount,preferredTitle,validateListingFields,type ListingFields } from "./listingFields";
import { formatMoney } from "./collections";
import { getOwnProduct, getOwnProductPage, updateOwnProduct, OWN_PRODUCT_PAGE_SIZE, ProductRequestError, type OwnProduct, type OwnProductSummary } from "./ownProducts";
import { confirmPublication, getPublicationTargets, type ProductPublicationTarget } from "./ownProducts";
import { getSiteTitleMode, type SiteTitleMode } from "./titleModePreference";
import { emptyProductPageFallback, productSaveNeedsReadback } from "./workbenchRules";

type Editor = Pick<OwnProduct, "title" | "description" | "notes"> & { listing_fields: ListingFields };
const BLANK:Editor={title:"",description:"",notes:"",listing_fields:initialListingFields()};
const count=(value:string)=>Array.from(value).length;

export default function OwnProductsPage({ onDirtyChange }: { onDirtyChange?: (dirty:boolean)=>void }) {
  const [query,setQuery]=useState("");
  const [queryDraft,setQueryDraft]=useState("");
  const [composing,setComposing]=useState(false);
  const [offset,setOffset]=useState(0);
  const [reload,setReload]=useState(0);
  const [items,setItems]=useState<OwnProductSummary[]>([]);
  const [hasMore,setHasMore]=useState(false);
  const [loading,setLoading]=useState(true);
  const [listError,setListError]=useState("");
  const [selected,setSelected]=useState<string|null>(null);
  const [product,setProduct]=useState<OwnProduct|null>(null);
  const [detailError,setDetailError]=useState("");
  const [detailLoading,setDetailLoading]=useState(false);
  const [fields,setFields]=useState<Editor>(BLANK);
  const [saving,setSaving]=useState(false);
  const [status,setStatus]=useState("");
  const [unknown,setUnknown]=useState(false);
  const dirty=!!product&&(fields.title!==product.title||fields.description!==product.description||fields.notes!==product.notes||
    JSON.stringify(fields.listing_fields)!==JSON.stringify(initialListingFields(product.listing_fields,product,product.title)));
  useEffect(()=>onDirtyChange?.(dirty),[dirty,onDirtyChange]);
  useEffect(()=>()=>onDirtyChange?.(false),[onDirtyChange]);
  useEffect(()=>{
    if(composing||dirty)return;
    const normalized=queryDraft.trim();
    if(normalized===query)return;
    const timer=window.setTimeout(()=>{
       setQuery(normalized);
      setOffset(0);
    },280);
    return ()=>window.clearTimeout(timer);
  },[queryDraft,composing,query,dirty]);
  useEffect(()=>{
    const c=new AbortController();
    setLoading(true);
    setListError("");
    void getOwnProductPage(query,offset,c.signal).then(page=>{
      if(c.signal.aborted)return;
      const previous=emptyProductPageFallback(offset,page.items.length,OWN_PRODUCT_PAGE_SIZE);
      if(previous!==null){setOffset(previous);return;}
      setItems(page.items);setHasMore(page.hasMore);
      setSelected(prev=>page.items.some(p=>p.id===prev)?prev:page.items[0]?.id??null);
    }).catch(e=>{if(!c.signal.aborted){setListError(e instanceof Error?e.message:"产品箱读取失败");setItems([]);setSelected(null);}})
      .finally(()=>{if(!c.signal.aborted)setLoading(false);});
    return ()=>c.abort();
  },[query,offset,reload]);

  useEffect(()=>{
    if(!selected){setProduct(null);setFields(BLANK);setDetailError("");return;}
    const c=new AbortController();setDetailLoading(true);setDetailError("");setStatus("");setUnknown(false);
    void getOwnProduct(selected,c.signal).then(p=>{
      if(c.signal.aborted)return;
      setProduct(p);setFields({title:p.title,description:p.description,notes:p.notes,
        listing_fields:initialListingFields(p.listing_fields,p,p.title)});
    }).catch(e=>{if(!c.signal.aborted){setDetailError(e instanceof Error?e.message:"详情读取失败");setProduct(null);}})
      .finally(()=>{if(!c.signal.aborted)setDetailLoading(false);});
    return ()=>c.abort();
  },[selected,reload]);

  const changeSelection=(id:string)=>{
    if(id===selected)return;
    if(dirty&&!window.confirm("当前产品箱商品有尚未保存的修改。确定丢弃并切换吗？"))return;
    setSelected(id);
  };
  const refresh=()=>{
    if(dirty&&!window.confirm("重新加载会丢失尚未保存的修改。确定继续吗？"))return;
    setReload(n=>n+1);
  };
  const save=async()=>{
    if(!product||saving||unknown||!dirty||!valid)return;
    setSaving(true);setStatus("");
    try {
      const {product:next}=await updateOwnProduct(product.id,{...fields,revision:product.revision},new AbortController().signal);
      setProduct(next);setFields({title:next.title,description:next.description,notes:next.notes,
        listing_fields:initialListingFields(next.listing_fields,next,next.title)});
      setItems(old=>old.map(x=>x.id===next.id?{...x,title:next.title,updated_at:next.updated_at,
        main_image_url:next.listing_fields.images!==undefined
          ?next.listing_fields.images[0]?.url??null
          :next.images[0]?.url??null}:x)
        .sort((a,b)=>b.updated_at.localeCompare(a.updated_at)||b.id.localeCompare(a.id)));
      setStatus("已保存到产品箱，重启后可恢复");
    } catch(e) {
      const message=e instanceof Error?e.message:"保存失败";
      setStatus(message);
      // A timeout or 202 may have committed. Do not replay stale revision.
      const code=e instanceof ProductRequestError?e.status:undefined;
      if(productSaveNeedsReadback(code))setUnknown(true);
    } finally {setSaving(false);}
  };
  const legacyDescriptionOversize = count(fields.description) > 20000;
  const validation = legacyDescriptionOversize ? "历史描述超过 20000 字符，请先在下方处理历史数据"
    : count(fields.notes) > 5000 ? "整理备注不能超过 5000 字符"
    : listingTextCount(fields.title) > 500 ? "商品标题不能超过 500 字符"
    : validateListingFields(fields.listing_fields, product?.listing_fields.titles);
  const valid = !validation && fields.title.trim().length > 0;
  return <div className="workspace">
    <section className="list-pane" aria-label="产品箱商品卡片">
      <div className="list-toolbar">
        <label className="search-box">
          <input type="search" placeholder={dirty?"请先保存商品再搜索":"搜索产品箱"} value={queryDraft} maxLength={200} readOnly={dirty}
            onChange={e=>setQueryDraft(e.target.value)}
            onCompositionStart={()=>setComposing(true)}
            onCompositionEnd={e=>{setComposing(false);setQueryDraft(e.currentTarget.value);}}
            aria-label="搜索产品箱" />
        </label>
        <div className="list-meta"><span>{loading?"读取中…":listError?"读取失败":"已整理产品"}</span><span>{items.length} 件（当前页）</span></div>
      </div>
      {listError?<p role="alert" className="workspace-feedback">{listError} <button type="button" onClick={refresh}>重试</button></p>
        :loading&&!items.length?<div className="loading-state">正在加载产品箱…</div>
        :!items.length?<div className="empty-state"><h2>产品箱暂无商品</h2><p>在采集箱勾选整理好的商品，点击“通过并移入产品箱”。</p></div>
        :<div className="product-card-grid">
          {items.map(item=><article key={item.id} className={"product-card-tile "+(item.id===selected?"selected":"")}>
            <button type="button" onClick={()=>changeSelection(item.id)} aria-label={"打开产品箱商品 "+item.title}>
              {item.main_image_url?<img src={item.main_image_url} alt={item.title} loading="lazy" referrerPolicy="no-referrer" />
                :<span className="image-fallback">暂无图片</span>}
              <strong title={item.title}>{item.title}</strong>
              <small>来源参考价：{formatMoney(item.price,item.currency)}</small>
            </button>
          </article>)}
        </div>}
      {(items.length>0||offset>0)&&<div className="list-footer"><span>第 {Math.floor(offset/OWN_PRODUCT_PAGE_SIZE)+1} 页</span>
        <div className="page-actions"><button disabled={loading||offset===0||dirty} onClick={()=>setOffset(Math.max(0,offset-OWN_PRODUCT_PAGE_SIZE))}>上一页</button>
          <button disabled={loading||!hasMore||dirty} onClick={()=>setOffset(offset+OWN_PRODUCT_PAGE_SIZE)}>下一页</button></div></div>}
    </section>
    <section className="detail-pane" aria-label="产品箱编辑详情">
      {detailError?<p className="workspace-feedback" role="alert">{detailError}<button onClick={refresh}>重新加载</button></p>
        :detailLoading?<div className="loading-state">正在加载详情…</div>
        :!product?<div className="empty-state">选择商品查看与编辑</div>
        :<div className="product-detail own-product-editor">
          <div className="detail-toolbar"><span className="detail-eyebrow">产品箱 · 独立商品草稿</span>
            <button className="refresh-button" type="button" onClick={refresh}>重新加载</button></div>
          <div className="own-product-preview">
            {fields.listing_fields.images?.[0]?.url
              ?<img src={fields.listing_fields.images[0].url} alt={fields.title} referrerPolicy="no-referrer" />
              :<span>暂无图片</span>}
            <div><strong>{product.source_title}</strong><small>来源：{product.source_platform} · {product.source_site}</small>
              <small>来源参考价：{formatMoney(product.price,product.currency)}</small>
              <a href={product.source_url} target="_blank" rel="noopener noreferrer">查看原商品</a></div>
          </div>
          <div className="working-draft">
            <ListingFieldsEditor key={product.id} value={fields.listing_fields} disabled={saving||unknown}
              aiProduct={{own_product_id:product.id,expected_revision:product.revision}} aiSaved={product.listing_fields}
              onChange={listing_fields=>setFields(v=>({...v,title:preferredTitle(listing_fields.titles),listing_fields}))}/>
            {legacyDescriptionOversize&&<details className="listing-fields-note">
               <summary>历史描述过长 · 查看及处理（非刊登描述）</summary>
               <textarea readOnly rows={4} value={fields.description} aria-label="超长历史描述，只读"/>
               <p>请先复制需要保留的历史文本。清空仅修改本地工作稿，保存后才写入；正常历史描述不会显示或被删除。</p>
               <button type="button" disabled={saving||unknown} onClick={()=>{
                 if(window.confirm("确认清空超长历史描述？正式多语言商品描述不受影响。"))setFields(v=>({...v,description:""}));
               }}>清空超长历史描述</button>
             </details>}
             <label>内部整理备注<textarea rows={4} value={fields.notes} onChange={e=>setFields(v=>({...v,notes:e.target.value}))} disabled={saving||unknown}/><small>{count(fields.notes)} / 5000</small></label>
            {validation && <p className="working-draft-error" role="alert">{validation}</p>}
            <div className="own-editor-actions">
              <span>{dirty?"有未保存修改":"内容已保存"} · 修订 {product.revision}</span>
              <button type="button" className="primary-button" disabled={!dirty||!valid||saving||unknown} onClick={()=>void save()}>{saving?"保存中…":"保存产品"}</button>
            </div>
            {status&&<p role="status" className="workspace-feedback">{status}</p>}
            {unknown&&<p>保存结果未知，请使用“重新加载”核实，勿重复提交。</p>}
          </div>
          <ProductPublication key={product.id+":"+product.revision} product={product} disabled={dirty||saving||unknown} />
          <p className="listing-fields-note">来源图片、原价格及原始规格始终保留在源快照中；上方编辑区修改的是你自己的产品数据。</p>
        </div>}
    </section>
  </div>;
}

function ProductPublication({product,disabled}: {product:OwnProduct;disabled:boolean}) {
  const [open,setOpen]=useState(false);
  const [targets,setTargets]=useState<ProductPublicationTarget[]>([]);
  const [selection,setSelection]=useState("");
  const [mode,setMode]=useState<SiteTitleMode>(getSiteTitleMode);
  const [loading,setLoading]=useState(false);
  const [busy,setBusy]=useState(false);
  const [message,setMessage]=useState("");
  const [refresh,setRefresh]=useState(0);
  const [result,setResult]=useState<Awaited<ReturnType<typeof confirmPublication>>|null>(null);
  const [uncertain,setUncertain]=useState(false);
  const key=(target: ProductPublicationTarget)=>target.store_id+":"+target.site_id;
  const target=targets.find(t=>key(t)===selection);
  useEffect(()=>{
    if(!open)return;
    const controller=new AbortController();
    setLoading(true);
    void getPublicationTargets(product.id,controller.signal).then(({targets:next})=>{
      if(controller.signal.aborted)return;
      setTargets(next);
      setSelection(current=>next.some(t=>key(t)===current)?current:next[0]?key(next[0]):"");
    }).catch(e=>{
      if(!controller.signal.aborted)setMessage(e instanceof Error?e.message:"读取刊登目标失败");
    }).finally(()=>{if(!controller.signal.aborted)setLoading(false);});
    return ()=>controller.abort();
  },[product.id,open,refresh]);
  const regional=target?.site_id==="MLB"?"pt":"es";
  const fields=product.listing_fields;
  const title=fields.titles?.[regional];
  const regionalDescription=fields.descriptions?.[regional];
  const submit=async()=>{
    if(!target||busy||loading||(!target.state&&(disabled||uncertain)))return;
    if(!target.state&&!window.confirm("确认刊登已保存的产品到 "+target.store_name+" · "+target.site_id+" / Remote？"))return;
    setBusy(true);setMessage("");setResult(null);
    try{
      const next=await confirmPublication(product.id,target,new AbortController().signal,target.state?undefined:{
        expected_revision:product.revision,title_mode:mode,
      });
      setResult(next);
      setUncertain(false);
      setTargets(old=>old.map(t=>key(t)===key(target)?{...t,state:next.state}:t));
    }catch(e){
      if(!target.state&&(!(e instanceof ProductRequestError)||e.status>=500||e.status===409||e.status===401)){
        setUncertain(true);
        setRefresh(n=>n+1);
      }
      setMessage(e instanceof Error?e.message:"刊登结果待核实");
    }finally{setBusy(false);}
  };
  if(!open)return <button type="button" disabled={disabled} onClick={()=>{setMode(getSiteTitleMode());setOpen(true);}}>准备刊登</button>;
  return <div className="working-draft">
    <label>刊登目标
      <select value={selection} disabled={busy||loading} onChange={e=>{setSelection(e.target.value);setResult(null);setMessage("");}}>
        {!targets.length&&<option value="">{loading?"读取中…":"暂无获批目标"}</option>}
        {targets.map(t=><option key={key(t)} value={key(t)}>{t.store_name} · {t.site_id} / Remote {t.listing_type_id??""}</option>)}
      </select>
    </label>
    {target&&!target.state&&<>
      <label>标题策略
        <select value={mode} disabled={busy} onChange={e=>setMode(e.target.value as SiteTitleMode)}>
          <option value="platform_auto">平台自动生成站点标题</option>
          <option value="localized">使用已保存的站点标题</option>
        </select>
      </label>
      <p>英文 Family 名 · {fields.titles?.en||"待填写"}</p>
      {mode==="localized"&&<p>站点标题 · {title||"待填写"}。{regionalDescription?.trim()?"使用已保存的站点描述。":"沿用英文描述。"}</p>}
    </>}
    {target?.state&&<p role="status">{target.state==="success"?"官方商品身份已确认":target.state==="submitting"?"正在提交":"刊登结果待核实"}</p>}
    {result&&<ul>{result.variants.map(v=><li key={v.Key}>{v.Listing.SKU} · {v.Listing.ItemID}{!v.TitleVerified&&" · 站点标题待核实"}</li>)}
      {result.pending.length>0&&<li>{result.pending.length} 个规格待核实</li>}</ul>}
    {message&&<p role="alert">{message}</p>}
    {uncertain&&<p>创建结果待核实，请刷新目标并读取已有尝试。</p>}
    <div className="own-editor-actions">
      <button type="button" disabled={busy||loading} onClick={()=>{setMessage("");setRefresh(n=>n+1);}}>刷新目标</button>
      {target&&<button type="button" className="primary-button"
        disabled={busy||loading||(!target.state&&(disabled||uncertain||(mode==="localized"&&!title?.trim())))}
        onClick={()=>void submit()}>{busy?"处理中…":target.state?"读取官方结果":"确认刊登"}</button>}
    </div>
  </div>;
}
