import type { ListingFields, ListingVariant } from "./listingFields";
import { detectTitleLanguage, validateListingFields, preferredTitle } from "./listingFields";
import { getCBTCategorySuggestions, type CBTCategorySuggestion } from "./collections";
import { useEffect, useRef, useState } from "react";

type Props={value:ListingFields;onChange:(next:ListingFields)=>void;disabled?:boolean};
 export default function ListingFieldsEditor({value,onChange,disabled=false}:Props) {
  const set=(key:keyof ListingFields,val:ListingFields[typeof key])=>onChange({...value,[key]:val});
  const [suggestions,setSuggestions]=useState<CBTCategorySuggestion[]>([]);
  const [categoryTranslations, setCategoryTranslations] = useState<Record<string, string>>({});
  const [chosenPath, setChosenPath] = useState<string[]>([]);
  const [predicting,setPredicting]=useState(false);
  const [predictMessage,setPredictMessage]=useState("");
  const predictionRequest = useRef<AbortController|null>(null);
  const [copyMessage,setCopyMessage]=useState("");
  useEffect(()=>()=>predictionRequest.current?.abort(),[]);
  const chosenTitle=preferredTitle(value.titles);
  const setLocalizedTitle=(lang:"en"|"es"|"pt",title:string)=>{
    predictionRequest.current?.abort();
    setPredicting(false);
    setSuggestions([]);
    const titles={...value.titles,[lang]:title};
    // CBT predictions depend on the English title, not the country languages.
    // Never leave a category that was chosen for a different English title.
    if(lang==="en" && title!==value.titles.en && value.category_id) {
      onChange({...value,titles,category_id:"",category_path:""});
      setPredictMessage("英文标题已修改，请重新选择或手动填写 CBT 类目。");
    } else {
      set("titles",titles);
      setPredictMessage("");
    }
  };
  const copyTitle=async(lang:"en"|"es"|"pt")=>{
    const title=value.titles[lang];
    if(!title)return;
    try {
      if(!navigator.clipboard?.writeText)throw new Error("clipboard unavailable");
      await navigator.clipboard.writeText(title);
      setCopyMessage(lang+" 已复制");
    }catch{
      setCopyMessage("复制未完成，请手动选择标题文本");
    }
  };
  const predict=async()=>{
    const english=value.titles.en.trim();
    if(!english || english.split(/\s+/u).length<2)return;
    predictionRequest.current?.abort();
    const controller=new AbortController();
    predictionRequest.current=controller;
    setPredicting(true);setPredictMessage("");setSuggestions([]);
    try{
      const result=await getCBTCategorySuggestions(english,controller.signal);
      if(controller.signal.aborted)return;
      setSuggestions(result.items);
      setCategoryTranslations(result.translations_zh ?? {});
      setPredictMessage(result.items.length?"请核对分类并点击确认候选；不会自动选中。":"官方未返回候选，请手动填写 CBT 类目。");
    }catch(e){
      if(!controller.signal.aborted)
        setPredictMessage(e instanceof Error?e.message:"暂时无法获取官方 CBT 分类推荐，请手动填写。");
    }finally{
      if(predictionRequest.current===controller) {
        predictionRequest.current=null;
        setPredicting(false);
      }
    }
  };
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
    set("images",next.map((im,i)=>({...im,type:(i===0?"main":"additional") as "main" | "additional"})));
  };
  const updateVariant=(index:number,patch:Partial<ListingVariant>)=>{
    set("variants",variants.map((v,i)=>i===index?{...v,...patch}:v));
  };
  return <section className="listing-fields">
    <h3>上架资料整理</h3>
    <p className="listing-fields-note">采集数据只是参考；此处编辑的内容会保存到当前工作稿，转入产品箱后仍可继续修改。站点类目及属性须在刊登前用美客多官方规则验证。</p>

    <fieldset><legend>三语商品标题</legend>
      <p className="listing-fields-note">采集时仅把原标题放入识别的语言栏；不自动翻译。无法确定语言时暂放英文栏，可手动调整。编辑完成后再使用 CBT 分类推荐。</p>
      <div className="localized-title-grid">
        {([
          {key:"en",label:"英文标题（源标题 / 类目推荐）",target:"建议 45–55 字符"},
          {key:"es",label:"西班牙语标题（墨西哥买家）",target:"建议不超过 60 字符"},
          {key:"pt",label:"葡萄牙语标题（巴西买家）",target:"建议不超过 60 字符"},
        ] as const).map(item=><label className="listing-field" key={item.key}>
          {item.label}
          <textarea rows={2} disabled={disabled} maxLength={500} value={value.titles[item.key]}
            onChange={e=>setLocalizedTitle(item.key,e.target.value)} placeholder="可人工填写或后续 AI 生成；不编造规格"/>
          <div className="localized-title-actions">
            <small className={"localized-title-count "+(item.key!=="en"&&Array.from(value.titles[item.key]).length>60?"title-over-limit":"")}>
              {Array.from(value.titles[item.key]).length} 字符（含空格及符号）· {item.target}
            </small>
            <button type="button" disabled={!value.titles[item.key]} onClick={()=>void copyTitle(item.key)}>
              复制{item.key==="en"?"英文":item.key==="es"?"西语":"葡语"}
            </button>
          </div>
        </label>)}
      </div>
      {copyMessage&&<p className="listing-fields-note" role="status">{copyMessage}</p>}
      {!chosenTitle&&<p className="listing-validation">至少需要一个语言标题，才能保存。</p>}
      <p className="listing-fields-note">三语都属于同一个产品，站点与店铺在实际刊登时选择；英文标题优先作为 CBT 推荐依据。</p>
    </fieldset>

    <fieldset><legend>分类建议与基本属性</legend>
      <p className="listing-fields-note">采集后不立即分类。先修改好标题，再按需点击“推荐 CBT 类目”。结果由 Mercado Libre Global Selling 官方接口返回，必须由你确认，站点专属必填属性在上传阶段另行验证。</p>
      <div className="category-predict-row">
        <button type="button" disabled={disabled||predicting||!value.titles.en.trim()||value.titles.en.trim().split(/\s+/).length<2}
          onClick={()=>void predict()}>{predicting?"查询中…":"根据英文标题推荐 CBT 类目"}</button>
        <small>来源语言识别：{detectTitleLanguage(chosenTitle)==="uncertain"?"不确定":detectTitleLanguage(chosenTitle)==="en"?"英语":detectTitleLanguage(chosenTitle)==="es"?"西语":"葡语"}（可手动调整）</small>
      </div>
      {predictMessage&&<p className="listing-fields-note" role="status">{predictMessage}</p>}
      {suggestions.length>0&&<div className="category-suggestion-list">
        {suggestions.map(item => {
          const names = item.path?.length ? item.path.map(node => node.name) : [item.category_name];
          const official = names.join(" / ");
          const translated = names.every(name => Object.hasOwn(categoryTranslations, name))
            ? names.map(name => categoryTranslations[name]).join(" / ") : "";
          return <button key={item.category_id} type="button" disabled={disabled}
            onClick={() => {
              onChange({...value,category_id:item.category_id,category_path:official.length <= 400 ? official : item.category_name});
              setChosenPath(names);
              setSuggestions([]);
              setPredictMessage("已填入所选 CBT 类目，仍需在刊登前核对官方必填属性。");
            }}>
            {translated && <strong>{translated}</strong>}
            <span>{official}</span><small>{item.category_id}</small>
          </button>;
        })}
      </div>}
      {chosenPath.join(" / ") === value.category_path && chosenPath.length > 0 &&
        chosenPath.every(name => Object.hasOwn(categoryTranslations, name)) &&
        <p className="listing-fields-note">{chosenPath.map(name => categoryTranslations[name]).join(" / ")}</p>}
      <div className="listing-field-grid">
        {field("CBT 分类名称 / 路径（可手动修改）","category_path","整理完标题后再填写或选择官方推荐",400)}
        {field("CBT 通用分类 ID（非 MLB/MLM）","category_id","例如 CBT12345；可手动填写",60)}
        {field("品牌","brand","无品牌请留空")}
        {field("型号","model")}
        {field("主商品 SKU（内部编码）","seller_sku","例如：FLOOR-001")}
        {field("条码 GTIN / EAN（有则填）","gtin","不编造条码",50)}
      </div>
    </fieldset>

    <fieldset><legend>采购成本</legend>
      <div className="listing-field-grid">
        {field("采购成本（不是来源竞品售价）","purchase_price","留空表示未知")}
        {field("采购币种","purchase_currency","如 CNY、USD",3)}
      </div>
      <p className="listing-fields-note">不在这里选择国家站点，也不填写各站点刊登售价。上传商品时再选择目标店铺、国家站点、对应售价及站点属性。</p>
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
            <button type="button" disabled={disabled} onClick={()=>set("images",images.filter((_,j)=>j!==i).map((x,j)=>({...x,type:(j===0?"main":"additional") as "main" | "additional"})))}>移除</button>
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
