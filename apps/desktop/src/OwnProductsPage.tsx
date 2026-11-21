import { useEffect, useState } from "react";
import { formatMoney } from "./collections";
import { getOwnProduct, getOwnProductPage, updateOwnProduct, type OwnProduct, type OwnProductSummary } from "./ownProducts";

type Editor = Pick<OwnProduct, "title" | "description" | "notes">;
const BLANK:Editor={title:"",description:"",notes:""};
const count=(value:string)=>Array.from(value).length;

export default function OwnProductsPage({ onDirtyChange }: { onDirtyChange?: (dirty:boolean)=>void }) {
  const [query,setQuery]=useState("");
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
  const dirty=!!product&&(fields.title!==product.title||fields.description!==product.description||fields.notes!==product.notes);
  useEffect(()=>onDirtyChange?.(dirty),[dirty,onDirtyChange]);
  useEffect(()=>{
    const c=new AbortController();
    setLoading(true);
    setListError("");
    void getOwnProductPage(query,offset,c.signal).then(page=>{
      if(c.signal.aborted)return;
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
      setProduct(p);setFields({title:p.title,description:p.description,notes:p.notes});
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
    if(!product||saving||unknown||!dirty)return;
    if(!fields.title.trim()||count(fields.title)>500||count(fields.description)>20000||count(fields.notes)>5000)return;
    setSaving(true);setStatus("");
    try {
      const {product:next}=await updateOwnProduct(product.id,{...fields,revision:product.revision},new AbortController().signal);
      setProduct(next);setFields({title:next.title,description:next.description,notes:next.notes});
      setItems(old=>old.map(x=>x.id===next.id?{...x,title:next.title,updated_at:next.updated_at}:x));
      setStatus("已保存到产品箱，重启后可恢复");
    } catch(e) {
      const message=e instanceof Error?e.message:"保存失败";
      setStatus(message);
      // A timeout or 202 may have committed. Do not replay stale revision.
      if(!message.includes("409")&&!message.includes("另一窗口"))setUnknown(true);
    } finally {setSaving(false);}
  };
  const valid=fields.title.trim().length>0&&count(fields.title)<=500&&count(fields.description)<=20000&&count(fields.notes)<=5000;
  const chooseSearch=(value:string)=>{
    if(dirty&&!window.confirm("搜索将切换商品，丢弃未保存内容。确定继续吗？"))return;
    setQuery(value);setOffset(0);
  };

  return <div className="workspace">
    <section className="list-pane" aria-label="产品箱商品卡片">
      <div className="list-toolbar">
        <label className="search-box">
          <input type="search" placeholder="搜索产品箱" value={query} onChange={e=>chooseSearch(e.target.value)} aria-label="搜索产品箱" />
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
              <small>{formatMoney(item.price,item.currency)}</small>
            </button>
          </article>)}
        </div>}
      {items.length>0&&<div className="list-footer"><span>第 {Math.floor(offset/24)+1} 页</span>
        <div className="page-actions"><button disabled={loading||offset===0||dirty} onClick={()=>setOffset(Math.max(0,offset-24))}>上一页</button>
          <button disabled={loading||!hasMore||dirty} onClick={()=>setOffset(offset+24)}>下一页</button></div></div>}
    </section>
    <section className="detail-pane" aria-label="产品箱编辑详情">
      {detailError?<p className="workspace-feedback" role="alert">{detailError}<button onClick={refresh}>重新加载</button></p>
        :detailLoading?<div className="loading-state">正在加载详情…</div>
        :!product?<div className="empty-state">选择商品查看与编辑</div>
        :<div className="product-detail own-product-editor">
          <div className="detail-toolbar"><span className="detail-eyebrow">产品箱 · 独立商品草稿</span>
            <button className="refresh-button" type="button" onClick={refresh}>重新加载</button></div>
          <div className="own-product-preview">
            {product.images[0]?.url?<img src={product.images[0].url} alt={product.title} referrerPolicy="no-referrer" />:<span>暂无图片</span>}
            <div><strong>{product.source_title}</strong><small>来源：{product.source_platform} · {product.source_site}</small>
              <small>{formatMoney(product.price,product.currency)}</small>
              <a href={product.source_url} target="_blank" rel="noopener noreferrer">查看原商品</a></div>
          </div>
          <div className="working-draft">
            <label>商品标题<input value={fields.title} onChange={e=>setFields(v=>({...v,title:e.target.value}))} disabled={saving||unknown} /><small>{count(fields.title)} / 500</small></label>
            <label>商品描述<textarea rows={8} value={fields.description} onChange={e=>setFields(v=>({...v,description:e.target.value}))} disabled={saving||unknown}/><small>{count(fields.description)} / 20000</small></label>
            <label>内部整理备注<textarea rows={4} value={fields.notes} onChange={e=>setFields(v=>({...v,notes:e.target.value}))} disabled={saving||unknown}/><small>{count(fields.notes)} / 5000</small></label>
            <div className="own-editor-actions">
              <span>{dirty?"有未保存修改":"内容已保存"} · 修订 {product.revision}</span>
              <button type="button" className="primary-button" disabled={!dirty||!valid||saving||unknown} onClick={()=>void save()}>{saving?"保存中…":"保存产品"}</button>
            </div>
            {status&&<p role="status" className="workspace-feedback">{status}</p>}
            {unknown&&<p>保存结果未知，请使用“重新加载”核实，勿重复提交。</p>}
          </div>
          {product.images.length>1&&<div className="own-product-image-strip">{product.images.slice(1).map((image,i)=><img key={i} src={image.url} alt={"附图 "+(i+1)} loading="lazy" referrerPolicy="no-referrer"/>)}</div>}
          {product.variants.length>0&&<section className="detail-section"><h3>来源规格（只读快照）</h3>{product.variants.map((v,i)=><p key={i}>{Object.entries(v.attributes).map(([k,val])=>k+"："+val).join(" / ")}</p>)}</section>}
        </div>}
    </section>
  </div>;
}
