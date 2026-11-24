import CategoryAttributesForm from "./CategoryAttributesForm";
import AICopyPanel from "./AICopyPanel";
import type { ListingFields, ListingVariant } from "./listingFields";
import { detectTitleLanguage, listingTextCount, localizedTitleLimit, preferredTitle } from "./listingFields";
import { getCBTCategorySuggestions, type AICopyProduct, type CBTCategorySuggestion, type CBTCategoryRules, type CBTAttributeRule } from "./collections";
import { useEffect, useRef, useState } from "react";
import { moveGalleryImage, removeDuplicateGalleryLinks, toggleVariantImage, renameVariantAttribute } from "./editorUx";
import { useSiteTitleMode } from "./titleModePreference";

type Props={value:ListingFields;onChange:(next:ListingFields)=>void;disabled?:boolean;aiProduct?:AICopyProduct;aiSaved?:Partial<ListingFields>};
export default function ListingFieldsEditor({value,onChange,disabled=false,aiProduct,aiSaved}:Props) {
  const set=(key:keyof ListingFields,val:ListingFields[typeof key])=>onChange({...value,[key]:val});
  const [categoryRules,setCategoryRules]=useState<CBTCategoryRules|null>(null);
  const [suggestions,setSuggestions]=useState<CBTCategorySuggestion[]>([]);
  const [predicting,setPredicting]=useState(false);
  const [predictMessage,setPredictMessage]=useState("");
  const predictionRequest = useRef<AbortController|null>(null);
  const [copyMessage,setCopyMessage]=useState("");
  const [activeLanguage,setActiveLanguage]=useState<"en"|"es"|"pt">("en");
  const siteTitleMode=useSiteTitleMode();
  const platformAuto=siteTitleMode==="platform_auto";
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
    if(lang!=="en") {
      onChange({...value,descriptions,description_translation:{
        ...value.description_translation,[`${lang}_source_digest`]:sourceDigest??"",
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
  const activeRules=categoryRules?.category_id===value.category_id?categoryRules:null;
  const childRules=(activeRules?.attributes??[]).filter(r=>r.hierarchy==="CHILD_PK"&&/^[A-Z][A-Z0-9_]{0,79}$/.test(r.id));
  const translations=activeRules?.translations_zh??{};
  const bilingual=(name:string)=>Object.hasOwn(translations,name)&&translations[name]!==name?name+"（"+translations[name]+"）":name;
  const options=(rule:CBTAttributeRule)=>(rule.values??[]).filter(v=>v.id!=null&&v.name);
  const supported=(rule:CBTAttributeRule)=>!["read_only","hidden","fixed","multivalued","conditional_required","size_grid_required"].some(t=>rule.tags?.[t])&&
    (rule.value_type==="string"||rule.value_type==="list"||rule.value_type==="number"||rule.value_type==="number_unit");
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
  const setOfficialSpec=(index:number,rule:CBTAttributeRule,entered:string,valueId?:string,unitId?:string)=>{
    const variant=variants[index];
    const history=(variant.category_attributes??[]).filter(a=>a.category_id!==value.category_id||a.id!==rule.id);
    if(entered||unitId)history.push({category_id:value.category_id,id:rule.id,value:entered,value_id:valueId,unit_id:unitId});
    updateVariant(index,{category_attributes:history});
  };
  const exactMatch=(rule:CBTAttributeRule,variant:ListingVariant)=>{
    const normalize=(s:string)=>s.trim().normalize("NFC").toLowerCase();
    const names=[rule.id,rule.name,translations[rule.name]].filter(Boolean).map(normalize);
    const raw=Object.entries(variant.attributes).filter(([name])=>names.includes(normalize(name)));
    if(raw.length!==1)return undefined;
    const matches=options(rule).filter(v=>[v.name,translations[v.name]].filter(Boolean).map(normalize).includes(normalize(raw[0][1])));
    return matches.length===1?matches[0]:undefined;
  };
  const specIssue=(variant:ListingVariant,rule:CBTAttributeRule)=>{
    const field=variant.category_attributes?.find(a=>a.category_id===value.category_id&&a.id===rule.id);
    if(!supported(rule))return "特殊规格，刊登前人工核实";
    if(!field?.value.trim())return "待填写";
    if(rule.value_type==="list"&&!options(rule).some(o=>String(o.id)===field.value_id&&o.name===field.value))return "待核实官方选项";
    if(!field.value_id&&/\p{Script=Han}/u.test(field.value))return "请填写英文规格";
    if((rule.value_type==="number"||rule.value_type==="number_unit")&&!/^-?\d+(?:\.\d+)?$/.test(field.value))return "请输入数字";
    if(rule.value_type==="number_unit"&&!(rule.allowed_units??[]).some(u=>u.id===field.unit_id))return "请选择官方单位";
    return "";
  };
  const completeKeys=variants.map(v=>childRules.length&&childRules.every(r=>!specIssue(v,r))?JSON.stringify(childRules.map(r=>{
    const f=v.category_attributes!.find(a=>a.category_id===value.category_id&&a.id===r.id)!;
    return [r.id,f.value_id||f.value.trim().toLowerCase(),f.unit_id??""];
  })):"");
  // These controls have separate identities across modes and languages so an
  // already focused Spanish/Portuguese textarea cannot become the English input.
  const renderTitle=(lang:"en"|"es"|"pt",label:string)=>{
    const language=languages.find(item=>item.key===lang)!;
    const count=listingTextCount(value.titles[lang]);
    return <>
      <label className="listing-field listing-active-title">
        <span>{label}</span>
        <textarea key={siteTitleMode+":"+lang} rows={3} disabled={disabled}
          value={value.titles[lang]}
          onChange={e=>setLocalizedTitle(lang,e.target.value)}
          placeholder="填写商品标题"/>
      </label>
      <div className="localized-title-actions">
        <small className={"localized-title-count "+(count>localizedTitleLimit?"title-over-limit":"")}>
          {count} 字符（含空格及符号） · {language.target}
          {lang==="en"&&count>localizedTitleLimit?" · 已超过 60 字符":""}
        </small>
        <button type="button" disabled={!value.titles[lang]} onClick={()=>void copyListingText(value.titles[lang],"标题")}>
          复制{lang==="en"?"英文":lang==="es"?"西语":"葡语"}标题
        </button>
      </div>
    </>;
  };
  const languageTabs=<div className="listing-language-tabs" role="group"
    aria-label="标题与详情语言">
    {languages.map(item=><button key={item.key} type="button"
      aria-pressed={activeLanguage===item.key} onClick={()=>setActiveLanguage(item.key)}>
      {item.label} <small>{listingTextCount(value.titles[item.key])}</small>
    </button>)}
  </div>;
  const renderDescription=(lang:"en"|"es"|"pt")=><>
    <label className="listing-field">{languages.find(item=>item.key===lang)!.label}商品描述
      <textarea key={siteTitleMode+":description:"+lang} rows={5} disabled={disabled} value={value.descriptions?.[lang]??""}
        onChange={e=>setLocalizedDescription(lang,e.target.value)} placeholder="填写纯文本商品描述"/>
      <small>{listingTextCount(value.descriptions?.[lang]??"")} / 20000 字符</small>
    </label>
    <div className="localized-title-actions">
      <button type="button" disabled={!value.descriptions?.[lang]}
        onClick={()=>void copyListingText(value.descriptions?.[lang]??"","详情")}>复制详情</button>
    </div>
  </>;
  return <section className="listing-fields">
    {platformAuto ? <>
      <fieldset key="platform-auto-title" className="listing-editor-panel">
        <legend>UP 英文主标题</legend>
        <p className="listing-fields-note">平台自动生成模式：仅编辑英文标题和详情。已保存的西语、葡语标题与详情会原样保留。</p>
        {renderTitle("en","英文主标题（UP Family Name 参考）")}
        {!value.titles.en.trim()&&<p className="listing-fields-note">尚未填写英文主标题。未来使用 UP 自动生成模式刊登前须填写英文标题；现有工作稿仍可按原规则保存。</p>}
      </fieldset>
      <fieldset key="platform-auto-description" className="listing-editor-panel">
        <legend>英文商品详情</legend>
        {renderDescription("en")}
      </fieldset>
    </> : <fieldset key="localized-title-description" className="listing-editor-panel">
      <legend>多语言标题与描述</legend>
      <p className="listing-fields-note">自定义站点标题：分别编辑英文、西语和葡语；最终能否提交站点级标题取决于 CBT 账号与官方接口。</p>
      {languageTabs}
      {renderTitle(activeLanguage,selectedLanguage.full)}
      {renderDescription(activeLanguage)}
    </fieldset>}
    {copyMessage&&<p className="listing-fields-note" role="status">{copyMessage}</p>}
    {!chosenTitle&&<p className="listing-validation">至少需要一个语言标题，才能保存。</p>}
    {aiProduct&&<AICopyPanel product={aiProduct} value={value} saved={aiSaved} disabled={disabled} englishOnly={platformAuto} onAdoptPoints={points=>{
      if(value.bullet_points.some(text=>text.trim()) &&
        JSON.stringify(value.bullet_points)!==JSON.stringify(points) &&
        !window.confirm("采纳 AI 卖点会替换当前全部商品卖点，是否继续？"))return;
      set("bullet_points",points);
    }} onAdopt={(field,lang,text,sourceDigest)=>{
      if(platformAuto&&lang!=="en")return;
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
          const translated = item.path_zh?.length === names.length ? item.path_zh.join(" / ") : official;
          return <button key={item.category_id} type="button" disabled={disabled}
            onClick={() => {
              onChange({...value, category_id: item.category_id,
                category_path: official.length <= 400 ? official : item.category_name});
              setSuggestions([]);
              setPredictMessage("");
            }}>
            <strong>{translated}</strong>{translated !== official && <span>{official}</span>}
            {item.domain_name && <small>{item.domain_name}</small>}
          </button>;
        })}
      </div>}
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
      onChange={fields => set("category_attributes",fields)} disabled={disabled} onRules={setCategoryRules}/>

    <fieldset className="listing-editor-plain" aria-label="采购成本与目标净收益">
      <div className="listing-field-grid">
      <label className="listing-field">采购成本（¥）
        <input value={legacyPurchase ? "" : value.purchase_price} disabled={disabled}
          placeholder={legacyPurchase ? "核实后输入人民币金额" : "人民币金额"}
          onChange={event=>onChange({...value,purchase_price:event.target.value,purchase_currency:"CNY"})}/>
      </label>
      {field("净收益（$）","target_net_proceeds_usd","目标净回款，USD",21)}
      </div>
      <small className="listing-fields-note">净收益为刊登目标净回款（美元），不是订单结算或利润。</small>
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

    <fieldset><legend>附加属性</legend>
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
      <p className="listing-fields-note">{!activeRules?"选择 CBT 类目并加载官方规则后可填写 UP 规格。":childRules.length===0?"当前规则未返回 CHILD_PK，保留原始规格，需在刊登前核实。":"官方规格使用英文及官方 ID，中文仅辅助阅读。原始规格不自动提交。"}</p>
      <div className="listing-variant-table-scroll" role="region" aria-label="可横向滚动的变体表格" tabIndex={0}>
        <table className="listing-variant-table">
          <thead><tr>{childRules.map(rule=><th key={rule.id} scope="col">{bilingual(rule.name||rule.id)}</th>)}<th scope="col">原始规格（参考）</th><th scope="col">SKU</th><th scope="col">采购成本（¥）</th><th scope="col">净收益（$）</th><th scope="col">来源价格</th><th scope="col">图片</th><th scope="col">操作</th></tr></thead>
          <tbody>{variants.map((v,i)=><tr key={i}>
            {childRules.map(rule=>{
              const f=v.category_attributes?.find(a=>a.category_id===value.category_id&&a.id===rule.id);
              const issue=specIssue(v,rule);
              const candidate=!f?.value&&exactMatch(rule,v);
              const label="变体 "+(i+1)+" "+(rule.name||rule.id);
              return <td key={rule.id}>
                {supported(rule)?rule.value_type==="list"?<select disabled={disabled} aria-label={label} value={f?.value_id??""}
                  onChange={e=>{const o=options(rule).find(o=>String(o.id)===e.target.value);setOfficialSpec(i,rule,o?.name??"",o?String(o.id):undefined);}}>
                  <option value="">请选择</option>
                  {f?.value_id&&!options(rule).some(o=>String(o.id)===f.value_id)&&<option value={f.value_id} disabled>{f.value}（待核实）</option>}
                  {options(rule).map(o=><option key={String(o.id)} value={String(o.id)}>{bilingual(o.name)}</option>)}
                </select>:<>
                  <input disabled={disabled} aria-label={label} value={f?.value??""} placeholder="English"
                    maxLength={Math.min(rule.value_max_length||500,500)} onChange={e=>setOfficialSpec(i,rule,e.target.value,undefined,f?.unit_id)}/>
                  {rule.value_type==="number_unit"&&<select disabled={disabled} aria-label={label+" unit"} value={f?.unit_id??""}
                    onChange={e=>setOfficialSpec(i,rule,f?.value??"",undefined,e.target.value||undefined)}>
                    <option value="">单位</option>{(rule.allowed_units??[]).map(u=><option key={u.id} value={u.id}>{u.name}</option>)}
                  </select>}
                </>:<small>特殊属性，保留历史值：{f?.value??"—"}</small>}
                {candidate&&supported(rule)&&rule.value_type==="list"&&<button type="button" disabled={disabled}
                  onClick={()=>setOfficialSpec(i,rule,candidate.name,String(candidate.id))}>匹配 {bilingual(candidate.name)}</button>}
                {issue&&<small className="listing-validation">{issue}</small>}
              </td>;
            })}
            <td className="listing-variant-spec">
              {completeKeys[i]&&completeKeys.filter(k=>k===completeKeys[i]).length>1&&<small className="listing-validation">重复官方规格组合 · 刊登前修正</small>}
              {Object.entries(v.attributes).map(([key,val],j)=><div className="listing-variant-attribute" key={j}>
                <input disabled={disabled} aria-label={"变体 "+(i+1)+" 属性 "+(j+1)+" 名称"} value={key} placeholder="Color / Size"
                  onChange={e=>updateVariant(i,{attributes:renameVariantAttribute(v.attributes,key,e.target.value)})}/>
                <input disabled={disabled} aria-label={"变体 "+(i+1)+" 属性 "+(j+1)+" 值"} value={val} placeholder="Black / M"
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
            <td><input disabled={disabled} aria-label={"变体 "+(i+1)+" 采购成本（¥）"} value={v.purchase_cost_cny??""}
              onChange={e=>updateVariant(i,{purchase_cost_cny:e.target.value})} placeholder="CNY"/></td>
            <td><input disabled={disabled} aria-label={"变体 "+(i+1)+" 净收益（$）"} value={v.target_net_proceeds_usd??""}
              onChange={e=>updateVariant(i,{target_net_proceeds_usd:e.target.value})} placeholder="USD"/></td>
            <td><small>{v.price? v.price+" "+(v.currency||"币种未确认"):"—"}</small></td>
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
    <fieldset className="listing-editor-panel">
      <legend>商品卖点</legend>
      <label className="listing-field">每行一条
        <textarea rows={4} disabled={disabled} value={(value.bullet_points??[]).join("\n")}
          onChange={e=>set("bullet_points",e.target.value.split("\n"))}
          placeholder="产品特点、适用场景等，仅录入真实信息"/>
      </label>
    </fieldset>
  </section>;
}
