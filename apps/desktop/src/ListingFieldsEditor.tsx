import CategoryAttributesForm from "./CategoryAttributesForm";
import AICopyPanel from "./AICopyPanel";
import type { ListingFields, ListingVariant } from "./listingFields";
import { detectTitleLanguage, listingTextCount, localizedTitleLimit, preferredTitle } from "./listingFields";
import { getCBTCategorySuggestions, type AICopyProduct, type CBTCategorySuggestion } from "./collections";
import { useEffect, useRef, useState } from "react";
import { moveGalleryImage, removeDuplicateGalleryLinks, toggleVariantImage, renameVariantAttribute } from "./editorUx";

type Props={value:ListingFields;onChange:(next:ListingFields)=>void;disabled?:boolean;aiProduct?:AICopyProduct;aiSaved?:Partial<ListingFields>};
export default function ListingFieldsEditor({value,onChange,disabled=false,aiProduct,aiSaved}:Props) {
  const set=(key:keyof ListingFields,val:ListingFields[typeof key])=>onChange({...value,[key]:val});
  const [suggestions,setSuggestions]=useState<CBTCategorySuggestion[]>([]);
  const [categoryTranslations, setCategoryTranslations] = useState<Record<string, string>>({});
  const [chosenPath, setChosenPath] = useState<string[]>([]);
  const [predicting,setPredicting]=useState(false);
  const [predictMessage,setPredictMessage]=useState("");
  const predictionRequest = useRef<AbortController|null>(null);
  const [copyMessage,setCopyMessage]=useState("");
  const [activeLanguage,setActiveLanguage]=useState<"en"|"es"|"pt">("en");
  const [imagePreview,setImagePreview]=useState<number|null>(null);
  const [editingImage,setEditingImage]=useState<number|null>(null);
  const [variantImageIndex,setVariantImageIndex]=useState<number|null>(null);
  const [variantImageUrlDraft,setVariantImageUrlDraft]=useState("");
  useEffect(()=>()=>predictionRequest.current?.abort(),[]);
  const chosenTitle=preferredTitle(value.titles);
  const legacyPurchase = Boolean(value.purchase_price && value.purchase_currency !== "CNY");
  const setLocalizedTitle=(lang:"en"|"es"|"pt",title:string)=>{
    predictionRequest.current?.abort();
    setPredicting(false);
    setSuggestions([]);
    const titles={...value.titles,[lang]:title};
    // Invalidate the category selected for the old English title, keeping its attribute history.
    if(lang==="en" && title.normalize("NFC")!==value.titles.en.normalize("NFC") && value.category_id) {
      onChange({...value,titles,category_id:"",category_path:""});
      setPredictMessage("英文标题已修改，请重新选择或手动填写 CBT 类目。");
    } else {
      set("titles",titles);
      setPredictMessage("");
    }
  };
  const setLocalizedDescription=(lang:"en"|"es"|"pt", text:string, sourceDigest?:string)=>{
    const descriptions={en:"",es:"",pt:"",...value.descriptions,[lang]:text};
    if(lang!=="en" && sourceDigest) {
      onChange({...value,descriptions,description_translation:{
        ...value.description_translation,[`${lang}_source_digest`]:sourceDigest,
      }});
    } else {
      set("descriptions",descriptions);
    }
  };
  const copyListingText=async(text:string,label:string)=>{
    if(!text)return;
    try {
      if(!navigator.clipboard?.writeText)throw new Error("clipboard unavailable");
      await navigator.clipboard.writeText(text);
      setCopyMessage("已复制"+label);
    }catch{
      setCopyMessage("复制未完成，请手动选择文本");
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
  const duplicateCount=images.length-removeDuplicateGalleryLinks(images).length;
  const languages=[
    {key:"en" as const,label:"英语 EN",full:"英文标题（源标题 / CBT 分类推荐）",target:"建议 45–55 字符"},
    {key:"es" as const,label:"西语 MX",full:"墨西哥西语标题",target:"不超过 60 字符"},
    {key:"pt" as const,label:"葡语 BR",full:"巴西葡语标题",target:"不超过 60 字符"},
  ];
  const selectedLanguage=languages.find(item=>item.key===activeLanguage)!;
  const field=(label:string,key:keyof ListingFields,placeholder="",maxLength=120)=>
    <label className="listing-field">{label}
      <input value={String(value[key]??"")} onChange={e=>set(key,e.target.value)}
        placeholder={placeholder} maxLength={maxLength} disabled={disabled} />
    </label>;
  const updateImage=(index:number,url:string)=>{
    set("images",images.map((im,i)=>i===index?{...im,url}:im));
  };
  const resetImageFocus=()=>{setImagePreview(null);setEditingImage(null);};
  const moveMain=(index:number)=>{set("images",moveGalleryImage(images,index,0));resetImageFocus();};
  const moveImage=(from:number,to:number)=>{set("images",moveGalleryImage(images,from,to));resetImageFocus();};
  const removeImage=(index:number)=>{
    set("images",images.filter((_,i)=>i!==index).map((im,i)=>({...im,type:(i===0?"main":"additional") as "main"|"additional"})));
    setImagePreview(null);
    setEditingImage(null);
  };
  const updateVariant=(index:number,patch:Partial<ListingVariant>)=>{
    set("variants",variants.map((v,i)=>i===index?{...v,...patch}:v));
  };
  return <section className="listing-fields">
    <fieldset className="listing-editor-panel">
      <legend>多语言标题与描述</legend>
      <div className="listing-language-tabs" role="group" aria-label="标题语言">
        {languages.map(item=><button key={item.key} type="button"
          aria-pressed={activeLanguage===item.key} onClick={()=>setActiveLanguage(item.key)}>
          {item.label} <small>{listingTextCount(value.titles[item.key])}</small>
        </button>)}
      </div>
      <label className="listing-field listing-active-title">
        <span>{selectedLanguage.full}</span>
        <textarea rows={3} disabled={disabled}
          value={value.titles[activeLanguage]}
          onChange={e=>setLocalizedTitle(activeLanguage,e.target.value)}
          placeholder="填写商品标题"/>
      </label>
      <div className="localized-title-actions">
        <small className={"localized-title-count "+(listingTextCount(value.titles[activeLanguage])>localizedTitleLimit?"title-over-limit":"")}>
          {listingTextCount(value.titles[activeLanguage])} 字符（含空格及符号） · {selectedLanguage.target}
          {activeLanguage==="en"&&listingTextCount(value.titles.en)>localizedTitleLimit?" · 已超过 60 字符":""}
        </small>
        <button type="button" disabled={!value.titles[activeLanguage]} onClick={()=>void copyListingText(value.titles[activeLanguage],"标题")}>
          复制{activeLanguage==="en"?"英文":activeLanguage==="es"?"西语":"葡语"}标题
        </button>
      </div>
      {copyMessage&&<p className="listing-fields-note" role="status">{copyMessage}</p>}
      {!chosenTitle&&<p className="listing-validation">至少需要一个语言标题，才能保存。</p>}
      <label className="listing-field">{selectedLanguage.label}商品描述
        <textarea rows={5} disabled={disabled} value={value.descriptions?.[activeLanguage]??""}
          onChange={e=>setLocalizedDescription(activeLanguage,e.target.value)} placeholder="填写纯文本商品描述"/>
        <small>{listingTextCount(value.descriptions?.[activeLanguage]??"")} / 20000 字符</small>
      </label>
      <div className="localized-title-actions">
        <button type="button" disabled={!value.descriptions?.[activeLanguage]}
          onClick={()=>void copyListingText(value.descriptions?.[activeLanguage]??"","详情")}>复制详情</button>
      </div>
    </fieldset>
    {aiProduct&&<AICopyPanel product={aiProduct} value={value} saved={aiSaved} disabled={disabled} onAdopt={(field,lang,text,sourceDigest)=>{
      if(field==="descriptions"){
        setLocalizedDescription(lang,text,sourceDigest);
        return;
      }
      if(lang==="en"&&value.category_id&&text!==value.titles.en.normalize("NFC")&&!window.confirm(
        "采纳新的英文标题会清空当前 CBT 类目及路径，但已填写属性会保留为历史参考。"
      ))return;
      setLocalizedTitle(lang,text);
    }}/>}
    <h4 className="listing-editor-section-title">基础信息</h4>
    <fieldset><legend>分类建议与基本属性</legend>
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
              onChange({...value, category_id: item.category_id,
                category_path: official.length <= 400 ? official : item.category_name});
              setChosenPath(names);
              setSuggestions([]);
              setPredictMessage("");
            }}>
            {translated ? <><strong>{translated}</strong><span>{official}</span></> : <strong>{official}</strong>}
            <small>{item.category_id} · {item.domain_name}</small>
          </button>;
        })}
      </div>}
      {chosenPath.join(" / ") === value.category_path && chosenPath.length > 0 &&
        chosenPath.every(name => Object.hasOwn(categoryTranslations, name)) &&
        <p className="listing-fields-note">{chosenPath.map(name => categoryTranslations[name]).join(" / ")}</p>}
      <div className="listing-field-grid">
        {field("CBT 分类名称 / 路径（可手动修改）","category_path","整理完标题后再填写或选择官方推荐",400)}
        <label className="listing-field">CBT 通用分类 ID
          <input value={value.category_id} maxLength={60} disabled={disabled} placeholder="例如 CBT12345"
            onChange={event => {
              predictionRequest.current?.abort();
              setPredicting(false);
              setSuggestions([]);
              setPredictMessage("");
              onChange({...value,category_id:event.target.value.trim().toUpperCase(),category_path:""});
            }}/>
        </label>
        {field("品牌","brand","无品牌请留空")}
        {field("型号","model")}
        {field("主商品 SKU（内部编码）","seller_sku","例如：FLOOR-001")}
        {field("条码 GTIN / EAN（有则填）","gtin","不编造条码",50)}
      </div>
    </fieldset>

    <CategoryAttributesForm category={value.category_id} values={value.category_attributes ?? []}
      onChange={fields => set("category_attributes",fields)} disabled={disabled}/>

    <fieldset className="listing-editor-plain"><legend>采购成本（人民币 CNY）</legend>
      <label className="listing-field">采购成本（¥）
        <input value={legacyPurchase ? "" : value.purchase_price} disabled={disabled}
          placeholder={legacyPurchase ? "核实后输入人民币金额" : "人民币金额"}
          onChange={event=>onChange({...value,purchase_price:event.target.value,purchase_currency:"CNY"})}/>
      </label>
      {legacyPurchase && <small className="listing-currency-legacy" role="status">
        历史采购成本 {value.purchase_price}（{value.purchase_currency || "币种未确认"}）。核实实际人民币金额后再填写。
      </small>}
    </fieldset>

    <fieldset className="listing-editor-plain"><legend>重量与包装尺寸</legend>
      <div className="listing-dimension-row">
        {field("重量（g）","weight_g","克",15)}
        {field("长（cm）","length_cm","厘米",15)}
        {field("宽（cm）","width_cm","厘米",15)}
        {field("高（cm）","height_cm","厘米",15)}
      </div>
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
    </fieldset>
    <fieldset className="listing-editor-panel">
      <legend>图片管理</legend>
      <div className="listing-gallery-toolbar">
        <strong>{images.length} 张图片</strong>
        <button type="button" disabled={disabled||duplicateCount===0}
          onClick={()=>{set("images",removeDuplicateGalleryLinks(images));setImagePreview(null);setEditingImage(null);}}>
          清理重复链接 {duplicateCount>0?"("+duplicateCount+")":""}
        </button>
        <button type="button" disabled={disabled||images.length>=60}
          onClick={()=>{set("images",[...images,{url:"",type:images.length?"additional":"main"}]);setEditingImage(images.length);}}>
          + 添加图片链接
        </button>
      </div>
      <div className="listing-gallery-wall">
        {images.map((im,i)=><article className="listing-gallery-tile" key={i}>
          <button type="button" className="listing-gallery-thumb" onClick={()=>setImagePreview(i)}
            aria-label={"预览第 "+(i+1)+" 张商品图片"}>
            {/^https:\/\//i.test(im.url)?<img src={im.url} alt={"第 "+(i+1)+" 张商品图片"} loading="lazy" referrerPolicy="no-referrer"/>:
              <span className="image-fallback">待填图片</span>}
            <span className="listing-gallery-number">{i===0?"主图":String(i+1)}</span>
          </button>
          <div className="listing-gallery-actions">
            <label className="listing-gallery-position">展示顺序
              <select disabled={disabled||images.length<2} aria-label={"第 "+(i+1)+" 张图片的展示位置"} value={i}
                onChange={e=>moveImage(i,Number(e.target.value))}>
                {images.map((_,position)=><option key={position} value={position}>第 {position+1} 张</option>)}
              </select>
            </label>
            <button type="button" disabled={disabled||i===0} onClick={()=>moveMain(i)}>设为主图</button>
            <button type="button" disabled={disabled||i===0} onClick={()=>moveImage(i,i-1)} aria-label={"图片 "+(i+1)+" 前移"}>←</button>
            <button type="button" disabled={disabled||i===images.length-1} onClick={()=>moveImage(i,i+1)} aria-label={"图片 "+(i+1)+" 后移"}>→</button>
            <button type="button" disabled={disabled} onClick={()=>setEditingImage(editingImage===i?null:i)}>链接</button>
            <button type="button" disabled={disabled} onClick={()=>removeImage(i)}>删除</button>
          </div>
          {editingImage===i&&<label className="listing-gallery-url">图片 HTTPS URL
            <input disabled={disabled} value={im.url} aria-label={"第 "+(i+1)+" 张图片 URL"}
              onChange={e=>updateImage(i,e.target.value)} placeholder="https://..." />
          </label>}
        </article>)}
      </div>
      {images.length===0&&<p className="listing-fields-note">暂无图片，可以手动添加 HTTPS 图片链接。</p>}
      {imagePreview!==null&&images[imagePreview]&&<div className="listing-gallery-preview" role="region" aria-label="大图预览">
        <div><strong>预览图片 {imagePreview+1}</strong>
          <button type="button" onClick={()=>setImagePreview(null)}>关闭预览</button>
        </div>
        {/^https:\/\//i.test(images[imagePreview].url)?
          <img src={images[imagePreview].url} alt={"图片 "+(imagePreview+1)+" 大图"} referrerPolicy="no-referrer"/>:
          <span>该图片还没有有效的 HTTPS 链接。</span>}
      </div>}
    </fieldset>
    <fieldset className="listing-editor-panel">
      <legend>变体 SKU</legend>
      <div className="listing-variant-toolbar">
        <strong>现有 {variants.length} 个变体</strong>
        <button type="button" disabled={disabled||variants.length>=100}
          onClick={()=>set("variants",[...variants,{seller_sku:"",attributes:{"属性1":""},price:"",currency:"",image_urls:[]}])}>
          + 添加变体
        </button>
      </div>
      <div className="listing-variant-table-scroll" role="region" aria-label="可横向滚动的变体表格" tabIndex={0}>
        <table className="listing-variant-table">
          <thead><tr><th scope="col">规格属性</th><th scope="col">SKU</th><th scope="col">价格</th><th scope="col">币种</th><th scope="col">图片</th><th scope="col">操作</th></tr></thead>
          <tbody>{variants.map((v,i)=><tr key={i}>
            <td className="listing-variant-spec">
              {Object.entries(v.attributes).map(([key,val],j)=><div className="listing-variant-attribute" key={j}>
                <input disabled={disabled} aria-label={"变体 "+(i+1)+" 属性 "+(j+1)+" 名称"} value={key} placeholder="颜色 / 尺寸"
                  onChange={e=>updateVariant(i,{attributes:renameVariantAttribute(v.attributes,key,e.target.value)})}/>
                <input disabled={disabled} aria-label={"变体 "+(i+1)+" 属性 "+(j+1)+" 值"} value={val} placeholder="黑色 / M"
                  onChange={e=>updateVariant(i,{attributes:{...v.attributes,[key]:e.target.value}})}/>
                <button type="button" disabled={disabled} onClick={()=>{
                  const next={...v.attributes};delete next[key];updateVariant(i,{attributes:next});
                }} aria-label={"删除变体 "+(i+1)+" 属性 "+(j+1)}>×</button>
              </div>)}
              <button type="button" disabled={disabled||Object.keys(v.attributes).length>=12}
                onClick={()=>{
                  let j=1;while(Object.prototype.hasOwnProperty.call(v.attributes,"属性"+j))j++;
                  updateVariant(i,{attributes:{...v.attributes,["属性"+j]:""}});
                }}>+ 属性</button>
            </td>
            <td><input disabled={disabled} aria-label={"变体 "+(i+1)+" SKU"} value={v.seller_sku}
              maxLength={120} onChange={e=>updateVariant(i,{seller_sku:e.target.value})} placeholder="自定义 SKU"/></td>
            <td><input disabled={disabled} aria-label={"变体 "+(i+1)+" 价格"} value={v.price}
              onChange={e=>updateVariant(i,{price:e.target.value})} placeholder="可留空"/></td>
            <td><input disabled={disabled} aria-label={"变体 "+(i+1)+" 币种"} value={v.currency}
              maxLength={3} onChange={e=>updateVariant(i,{currency:e.target.value})} placeholder="USD"/></td>
            <td><button type="button" className="listing-variant-images-button" disabled={disabled}
                onClick={()=>{setVariantImageIndex(variantImageIndex===i?null:i);setVariantImageUrlDraft("");}}>
                选择图片 ({v.image_urls.length})
              </button>
              <div className="listing-variant-mini-images">
                {v.image_urls.slice(0,3).map((url,j)=><img key={j} src={url} alt={"变体 "+(i+1)+" 图片 "+(j+1)} loading="lazy" referrerPolicy="no-referrer"/>)}
              </div>
            </td>
            <td><button type="button" disabled={disabled} onClick={()=>{
              set("variants",variants.filter((_,j)=>j!==i));
              setVariantImageIndex(null);
            }}>删除</button></td>
          </tr>)}</tbody>
        </table>
      </div>
      {variants.length===0&&<p className="listing-fields-note">未采集到变体，可点击“添加变体”手动填写真实规格。</p>}
      {variantImageIndex!==null&&variants[variantImageIndex]&&<section className="listing-variant-image-picker" aria-label={"变体 "+(variantImageIndex+1)+" 图片选择"}>
        <div className="listing-variant-image-picker-head">
          <strong>变体 {variantImageIndex+1} · 选择图片</strong>
          <button type="button" onClick={()=>setVariantImageIndex(null)}>完成</button>
        </div>
        <div className="listing-variant-image-options">
          {Array.from(new Set([...images.map(im=>im.url),...variants[variantImageIndex].image_urls]))
            .filter(url=>url.trim()!=="").map(url=><label key={url}>
              <input type="checkbox" disabled={disabled} checked={variants[variantImageIndex].image_urls.includes(url)}
                onChange={e=>updateVariant(variantImageIndex,{
                  image_urls:toggleVariantImage(variants[variantImageIndex].image_urls,url,e.target.checked)
                })}/>
              <img src={url} alt="可选变体图片" loading="lazy" referrerPolicy="no-referrer"/>
            </label>)}
        </div>
        <div className="listing-variant-url-row">
          <input disabled={disabled} aria-label="追加变体图片 URL" placeholder="https://..." value={variantImageUrlDraft}
            onChange={e=>setVariantImageUrlDraft(e.target.value)}/>
          <button type="button" disabled={disabled||!/^https:\/\/\S+$/i.test(variantImageUrlDraft.trim())}
            onClick={()=>{
              updateVariant(variantImageIndex,{image_urls:toggleVariantImage(
                variants[variantImageIndex].image_urls,variantImageUrlDraft.trim(),true)});
              setVariantImageUrlDraft("");
            }}>添加链接</button>
        </div>
      </section>}
    </fieldset>
  </section>;
}
