import type { ListingFields, ListingVariant } from "./listingFields";
import { SITE_NAMES, validateListingFields } from "./listingFields";

type Props={value:ListingFields;onChange:(next:ListingFields)=>void;disabled?:boolean};
const CURRENCIES:Record<string,string>={MLB:"BRL",MLM:"MXN",MLA:"ARS",MLC:"CLP",MCO:"COP"};
const targetSites=Object.keys(SITE_NAMES);

export default function ListingFieldsEditor({value,onChange,disabled=false}:Props) {
  const set=(key:keyof ListingFields,val:ListingFields[typeof key])=>onChange({...value,[key]:val});
  const prices=value.site_prices??[];
  const images=value.images??[];
  const variants=value.variants??[];
  const error=validateListingFields(value);
  const field=(label:string,key:keyof ListingFields,placeholder="",maxLength=120)=>
    <label className="listing-field">{label}
      <input value={String(value[key]??"")} onChange={e=>set(key,e.target.value)}
        placeholder={placeholder} maxLength={maxLength} disabled={disabled} />
    </label>;
  const updateImage=(index:number,url:string)=>{
    set("images",images.map((im,i)=>i===index?{...im,url}:im));
  };
  const moveMain=(index:number)=>{
    const next=images.slice();
    const [pick]=next.splice(index,1);
    next.unshift(pick);
    set("images",next.map((im,i)=>({...im,type:i===0?"main":"additional"})));
  };
  const updateVariant=(index:number,patch:Partial<ListingVariant>)=>{
    set("variants",variants.map((v,i)=>i===index?{...v,...patch}:v));
  };
  return <section className="listing-fields">
    <h3>上架资料整理</h3>
    <p className="listing-fields-note">采集数据只是参考；此处编辑的内容会保存到当前工作稿，转入产品箱后仍可继续修改。站点类目及属性须在刊登前用美客多官方规则验证。</p>

    <fieldset><legend>分类及基本信息</legend>
      <div className="listing-field-grid">
        <label className="listing-field">目标站点
          <select value={value.category_site} disabled={disabled} onChange={e=>set("category_site",e.target.value)}>
            <option value="">待选择</option>
            {targetSites.map(site=><option key={site} value={site}>{SITE_NAMES[site]}</option>)}
          </select>
        </label>
        {field("分类路径（中文或官方分类名称）","category_path","例如：家居 / 地板涂料",400)}
        {field("官方分类 ID（待核实）","category_id","例如：MLB…",60)}
        {field("品牌","brand","无品牌请留空")}
        {field("型号","model")}
        {field("主商品 SKU（内部编码）","seller_sku","例如：FLOOR-001")}
        {field("条码 GTIN / EAN（有则填）","gtin","不编造条码",50)}
      </div>
    </fieldset>

    <fieldset><legend>价格与站点售价</legend>
      <div className="listing-field-grid">
        {field("采购成本（非采集来源售价）","purchase_price","留空表示未知")}
        {field("采购币种","purchase_currency","如 CNY、USD",3)}
      </div>
      <div className="site-price-list">
        {prices.map((p,i)=><div className="site-price-row" key={p.site}>
          <strong>{SITE_NAMES[p.site]??p.site}</strong>
          <label>刊登目标价
            <input value={p.price} placeholder="尚未定价" disabled={disabled}
              onChange={e=>set("site_prices",prices.map((x,j)=>j===i?{...x,price:e.target.value}:x))}/>
          </label>
          <span>{p.currency||CURRENCIES[p.site]}</span>
          <button type="button" disabled={disabled} onClick={()=>set("site_prices",prices.filter((_,j)=>i!==j))}>移除</button>
        </div>)}
        {prices.length<5&&<div className="site-price-add">
          <span>增加目标站点：</span>
          {targetSites.filter(site=>!prices.some(p=>p.site===site)).map(site=>
            <button type="button" key={site} disabled={disabled} onClick={()=>
              set("site_prices",[...prices,{site,price:"",currency:CURRENCIES[site]}])}>+ {SITE_NAMES[site]}</button>)}
        </div>}
      </div>
      <p className="listing-fields-note">站点价格是草稿字段，不会因保存而提交到美客多。不同国家、币种分别填写。</p>
    </fieldset>

    <fieldset><legend>重量与包装尺寸</legend>
      <div className="listing-dimension-row">
        {field("重量（g）","weight_g","克",15)}
        {field("长（cm）","length_cm","厘米",15)}
        {field("宽（cm）","width_cm","厘米",15)}
        {field("高（cm）","height_cm","厘米",15)}
      </div>
      <p className="listing-fields-note">填写商品的实际包装重量和尺寸，留空表示待测量，不根据采集售价、图片估算。</p>
    </fieldset>

    <fieldset><legend>卖点与分类属性</legend>
      <label className="listing-field">商品卖点（每行一条）
        <textarea rows={4} disabled={disabled} value={(value.bullet_points??[]).join("\n")}
          onChange={e=>set("bullet_points",e.target.value.split("\n"))}
          placeholder="产品特点、适用场景等，仅录入真实信息"/>
      </label>
      {(value.attributes??[]).map((a,i)=><div className="listing-attribute-row" key={i}>
        <input disabled={disabled} aria-label={"属性名称 "+(i+1)} value={a.name} placeholder="属性名称"
          onChange={e=>set("attributes",value.attributes.map((x,j)=>i===j?{...x,name:e.target.value}:x))}/>
        <input disabled={disabled} aria-label={"属性值 "+(i+1)} value={a.value} placeholder="属性值"
          onChange={e=>set("attributes",value.attributes.map((x,j)=>i===j?{...x,value:e.target.value}:x))}/>
        <button type="button" disabled={disabled} onClick={()=>set("attributes",value.attributes.filter((_,j)=>j!==i))}>删除</button>
      </div>)}
      <button type="button" disabled={disabled||value.attributes.length>=60}
        onClick={()=>set("attributes",[...value.attributes,{name:"",value:""}])}>+ 添加属性</button>
      <p className="listing-fields-note">分类必填属性会在接入该站点官方分类规则后明确标识；目前不能把手工填写当作官方校验通过。</p>
    </fieldset>

    <fieldset><legend>商品图片</legend>
      <div className="listing-image-grid">
        {images.map((im,i)=><div className="listing-image-card" key={i}>
          {/^https:\/\//.test(im.url)?<img src={im.url} alt={"商品图片 "+(i+1)} loading="lazy" referrerPolicy="no-referrer"/>:
            <div className="image-fallback">待填图片 URL</div>}
          <input value={im.url} disabled={disabled} aria-label={"第 "+(i+1)+" 张图片 URL"} placeholder="HTTPS 图片地址"
            onChange={e=>updateImage(i,e.target.value)}/>
          <div>
            <button type="button" disabled={disabled||i===0} onClick={()=>moveMain(i)}>{i===0?"首图":"设为首图"}</button>
            <button type="button" disabled={disabled} onClick={()=>set("images",images.filter((_,j)=>j!==i).map((x,j)=>({...x,type:j===0?"main":"additional"})))}>移除</button>
          </div>
        </div>)}
      </div>
      <button type="button" disabled={disabled||images.length>=60}
        onClick={()=>set("images",[...images,{url:"",type:images.length?"additional":"main"}])}>+ 添加图片 URL</button>
      <p className="listing-fields-note">这里支持调整顺序、移除和替换图片链接；真正上传本地图片、AI 修图、图片永久托管需要后续图片资产功能。仅有来源 URL 不保证美客多可直接使用。</p>
    </fieldset>

    <fieldset><legend>规格与 SKU</legend>
      {variants.map((v,i)=><div className="listing-variant" key={i}>
        <div className="listing-variant-heading"><strong>规格 {i+1}</strong>
          <button type="button" disabled={disabled} onClick={()=>set("variants",variants.filter((_,j)=>j!==i))}>删除规格</button>
        </div>
        <div className="listing-field-grid">
          <label className="listing-field">SKU
            <input disabled={disabled} value={v.seller_sku} maxLength={120}
              onChange={e=>updateVariant(i,{seller_sku:e.target.value})}/>
          </label>
          <label className="listing-field">规格价格（可选）
            <input disabled={disabled} value={v.price} placeholder="留空" onChange={e=>updateVariant(i,{price:e.target.value})}/>
          </label>
          <label className="listing-field">币种
            <input disabled={disabled} value={v.currency} maxLength={3} onChange={e=>updateVariant(i,{currency:e.target.value})}/>
          </label>
        </div>
        {Object.entries(v.attributes).map(([key,val],j)=><div className="listing-attribute-row" key={j}>
          <input disabled={disabled} value={key} aria-label={"规格 "+(i+1)+" 的属性名"} placeholder="如颜色"
            onChange={e=>{
              const pairs=Object.entries(v.attributes);
              pairs[j]=[e.target.value,val];
              updateVariant(i,{attributes:Object.fromEntries(pairs)});
            }}/>
          <input disabled={disabled} value={val} aria-label={"规格 "+(i+1)+" 的属性值"} placeholder="如黑色"
            onChange={e=>updateVariant(i,{attributes:{...v.attributes,[key]:e.target.value}})}/>
          <button type="button" disabled={disabled} onClick={e=>{e.preventDefault();const next={...v.attributes};delete next[key];updateVariant(i,{attributes:next});}}>删除</button>
        </div>)}
        <button type="button" disabled={disabled||Object.keys(v.attributes).length>=12}
          onClick={()=>updateVariant(i,{attributes:{...v.attributes,["新属性"+(Object.keys(v.attributes).length+1)]:""}})}>+ 规格属性</button>
      </div>)}
      <button type="button" disabled={disabled||variants.length>=100}
        onClick={()=>set("variants",[...variants,{seller_sku:"",attributes:{颜色:""},price:"",currency:"",image_urls:[]}])}>+ 添加规格</button>
    </fieldset>

    {error&&<p className="listing-validation" role="alert">{error}</p>}
  </section>;
}
