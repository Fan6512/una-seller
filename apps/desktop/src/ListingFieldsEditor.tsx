import CategoryAttributesForm, { choicesFor, enumOnly, requiredAttribute, unitsFor, variantRuleEditable } from "./CategoryAttributesForm";
import AICopyPanel from "./AICopyPanel";
import type { ListingFields, ListingVariant } from "./listingFields";
import { detectTitleLanguage, listingTextCount, localizedTitleLimit, preferredTitle } from "./listingFields";
import { getCBTCategorySuggestions, type AICopyProduct, type CBTCategorySuggestion, type CBTCategoryRules, type CBTAttributeRule } from "./collections";
import { useEffect, useId, useRef, useState } from "react";
import { moveGalleryImage, removeDuplicateGalleryLinks, toggleVariantImage, renameVariantAttribute, planSecondaryImageReuse } from "./editorUx";
import { useOperationFeedback } from "./operationFeedbackContext";
import { useSiteTitleMode } from "./titleModePreference";

type Props={value:ListingFields;onChange:(next:ListingFields)=>void;disabled?:boolean;aiProduct?:AICopyProduct;aiSaved?:Partial<ListingFields>};
export default function ListingFieldsEditor({value,onChange,disabled=false,aiProduct,aiSaved}:Props) {
  const feedback = useOperationFeedback();
  const imageFeedbackScope = useId();
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
  const [reuseMode,setReuseMode]=useState<"bulk"|"single"|null>(null);
  const [reuseSource,setReuseSource]=useState("");
  const [reuseTargets,setReuseTargets]=useState<number[]>([]);
  const [reuseMessage,setReuseMessage]=useState("");
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
  const modelRule=activeRules?.attributes.find(rule=>rule.id==="MODEL");
  const modelRequired=Boolean(activeRules&&(activeRules.technical_requirements?.includes("MODEL")||
    (modelRule&&requiredAttribute(modelRule,activeRules.technical_requirements))));
  const modelHint=!activeRules?"选择类目后核实":
    modelRequired?"官方 MODEL 必填，请在官方属性或变体规格填写":
    !activeRules.technical_specs_verified||modelRule?.tags?.conditional_required?"官方要求待核实":
    "当前已读取规则未标记 MODEL 必填";
  const childRules=(activeRules?.attributes??[]).filter(r=>r?.hierarchy==="CHILD_PK"&&/^[A-Z][A-Z0-9_]{0,79}$/.test(r.id));
  // These are display-only hints; never change the platform option's value or ID.
  const glossary:Record<string,string>={Color:"颜色",Colour:"颜色",Size:"尺寸",Style:"款式",Material:"材质",Voltage:"电压",
    Capacity:"容量",Black:"黑色",White:"白色",Red:"红色",Blue:"蓝色",Green:"绿色",Yellow:"黄色",
    Pink:"粉色",Purple:"紫色",Gray:"灰色",Grey:"灰色",Brown:"棕色",Orange:"橙色",
    Beige:"米色","Navy Blue":"藏青色",Silver:"银色",Gold:"金色",Khaki:"卡其色",Burgundy:"酒红色"};
  const translations={...glossary,...activeRules?.translations_zh};
  const [specSearch,setSpecSearch]=useState<Record<string,string>>({});
  useEffect(()=>setSpecSearch({}),[value.category_id]);
  const bilingual=(name:string)=>Object.hasOwn(translations,name)&&translations[name]!==name?name+"（"+translations[name]+"）":name;
  const supported=variantRuleEditable;
  const hasChoices=(rule:CBTAttributeRule)=>supported(rule)&&rule.value_type!=="number_unit"&&choicesFor(rule).length>0;
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
  const variantLabel=(v:ListingVariant,i:number)=>{
    // Prefer confirmed specs from the current category, never another category's history.
    const specs=(v.category_attributes??[]).filter(a=>a.category_id===value.category_id&&a.value.trim()&&
      (!activeRules||childRules.some(rule=>rule.id===a.id)));
    const ordered=activeRules?childRules.flatMap(rule=>specs.filter(a=>a.id===rule.id)):specs;
    const official=ordered.map(a=>{
      const rule=childRules.find(r=>r.id===a.id);
      const unit=rule&&a.unit_id?unitsFor(rule).find(u=>u.id===a.unit_id)?.name:undefined;
      return bilingual(a.value.trim())+(a.unit_id?" "+bilingual(unit||a.unit_id):"");
    }).join(" / ");
    const raw=Object.values(v.attributes??{}).filter(s=>typeof s==="string").map(s=>s.trim()).filter(Boolean).map(bilingual).join(" / ");
    const spec=official||raw;
    return "变体 "+(i+1)+" · "+([spec,(v.seller_sku??"").trim()].filter(Boolean).join(" · ")||"未命名");
  };
  const beginReuse=(mode:"bulk"|"single",target?:number)=>{
    const index=variants.findIndex((v,i)=>i!==target&&v.image_urls.length>1);
    setReuseMode(mode);setReuseSource(index<0?"":String(index));
    setReuseTargets(mode==="single"&&target!==undefined?[target]:variants.map((_,i)=>i).filter(i=>i!==index));
    setReuseMessage("");
  };
  const applyReuse=()=>{
    if(disabled||reuseSource==="")return;
    const sourceIndex=Number(reuseSource);
    const source=variants[sourceIndex];
    if(!source)return;
    const plans=new Map(reuseTargets.filter(i=>i!==sourceIndex&&variants[i]).map(i=>[i,planSecondaryImageReuse(source.image_urls,variants[i].image_urls)] as const));
    const changed=[...plans].filter(([,p])=>p.changed);
    if(changed.some(([i])=>variants[i].image_urls.length>1)&&!window.confirm("将覆盖 "+changed.filter(([i])=>variants[i].image_urls.length>1).length+" 个变体已有副图，保留主图。是否继续？"))return;
    if(changed.length)set("variants",variants.map((v,i)=>plans.get(i)?.changed?{...v,image_urls:plans.get(i)!.urls}:v));
    if (reuseMode === "bulk" && plans.size) {
      const batch = feedback.batch({ scope: imageFeedbackScope, key: "reuse-secondary-images" }, "批量复用副图", plans.size);
      for (const [index, plan] of plans) {
        batch.record(String(index), plan.changed ? "success" : "skipped");
      }
      batch.finish("副图复用结果");
    }
    const skipped=[...plans].filter(([,p])=>p.error).map(([i,p])=>variantLabel(variants[i],i)+"："+p.error);
    setReuseMessage("已更新 "+changed.length+" 个变体；未变更 "+(plans.size-changed.length)+" 个。"+(skipped.length?" "+skipped.join("；"):""));
  };
  const reusePanel=reuseMode&&<div className="listing-variant-image-picker" aria-label="复用副图">
    <strong>{reuseMode==="bulk"?"批量复用副图":"复用副图"}</strong>
    <label>来源变体
      <select disabled={disabled} value={reuseSource} onChange={e=>{
        setReuseSource(e.target.value);
        if(reuseMode==="bulk")setReuseTargets(variants.map((_,i)=>i).filter(i=>String(i)!==e.target.value));
        setReuseMessage("");
      }}>
        <option value="">请选择</option>
        {variants.map((v,i)=>(reuseMode!=="single"||!reuseTargets.includes(i))&&<option key={i} value={i} disabled={v.image_urls.length<2}>{variantLabel(v,i)}（{Math.max(0,v.image_urls.length-1)} 张副图）</option>)}
      </select>
    </label>
    {reuseMode==="bulk"&&<>
      <button type="button" disabled={disabled} onClick={()=>setReuseTargets(variants.map((_,i)=>i).filter(i=>String(i)!==reuseSource))}>全选其他变体</button>
      <button type="button" disabled={disabled} onClick={()=>setReuseTargets([])}>清空选择</button>
      {variants.map((v,i)=>String(i)!==reuseSource&&<label key={i}>
        <input type="checkbox" disabled={disabled} checked={reuseTargets.includes(i)} onChange={e=>setReuseTargets(old=>e.target.checked?[...old,i]:old.filter(x=>x!==i))}/>
        {variantLabel(v,i)}{v.image_urls.length?"":"（尚无主图，将跳过）"}
      </label>)}
    </>}
    <small>仅复用来源第 2 张起的副图，目标主图不变。</small>
    <button type="button" disabled={disabled||reuseSource===""||!reuseTargets.some(i=>String(i)!==reuseSource)} onClick={applyReuse}>应用复用副图</button>
    <button type="button" onClick={()=>{setReuseMode(null);setReuseMessage("");}}>关闭</button>
    {reuseMessage&&<p role="status">{reuseMessage}</p>}
  </div>;
  const setOfficialSpec=(index:number,rule:CBTAttributeRule,entered:string,valueId?:string,unitId?:string)=>{
    const variant=variants[index];
    const history=(variant.category_attributes??[]).filter(a=>a.category_id!==value.category_id||a.id!==rule.id);
    if(entered||unitId)history.push({category_id:value.category_id,id:rule.id,value:entered,value_id:valueId,unit_id:unitId});
    updateVariant(index,{category_attributes:history});
  };
  const normalize=(s:string)=>s.trim().normalize("NFC").toLowerCase();
  const exactMatch=(rule:CBTAttributeRule,variant:ListingVariant)=>{
    const names=[rule.id,rule.name,translations[rule.name]].filter(Boolean).map(normalize);
    const raw=Object.entries(variant.attributes).filter(([name])=>names.includes(normalize(name)));
    if(raw.length!==1)return undefined;
    const matches=choicesFor(rule).filter(v=>[v.name,translations[v.name]].filter(Boolean).map(normalize).includes(normalize(raw[0][1])));
    return matches.length===1?matches[0]:undefined;
  };
  const currentSpec=(variant:ListingVariant,rule:CBTAttributeRule)=>
    variant.category_attributes?.find(a=>a.category_id===value.category_id&&a.id===rule.id);
  const mandatory=(rule:CBTAttributeRule)=>requiredAttribute(rule,activeRules?.technical_requirements);
  const familyChildRules=childRules.filter(rule=>supported(rule)&&
    (mandatory(rule)||variants.some(v=>currentSpec(v,rule)?.value.trim())));
  const specIssue=(field:ReturnType<typeof currentSpec>,rule:CBTAttributeRule)=>{
    if(!supported(rule))return "特殊规格，刊登前人工核实";
    if(!field?.value.trim())return mandatory(rule)||field?.unit_id?"待填写规格值":
      familyChildRules.includes(rule)?"请核对家族规格项是否一致":"";
    if(Array.from(field.value).length>Math.min(rule.value_max_length||500,500))return "超过官方长度上限";
    if((enumOnly(rule)||field.value_id)&&!choicesFor(rule).some(o=>String(o.id)===field.value_id&&o.name===field.value))return "待核实官方选项";
    if(!field.value_id&&/\p{Script=Han}/u.test(field.value))return "请填写英文规格";
    if((rule.value_type==="number"||rule.value_type==="number_unit")&&!/^-?\d+(?:\.\d+)?$/.test(field.value))return "请输入数字";
    if(rule.value_type==="number_unit"&&!unitsFor(rule).some(u=>u.id===field.unit_id))return "请选择官方单位";
    if(rule.value_type!=="number_unit"&&field.unit_id)return "当前字段不接受单位";
    return "";
  };
  const completeKeys=variants.map(v=>familyChildRules.length&&familyChildRules.every(r=>!specIssue(currentSpec(v,r),r))?JSON.stringify(familyChildRules.map(r=>{
    const f=currentSpec(v,r)!;
    let canonical=normalize(f.value);
    if(r.value_type==="number"||r.value_type==="number_unit") {
      canonical=canonical.replace(/^(-?)0+(?=\d)/,"$1").replace(/(\.\d*?)0+$/,"$1").replace(/\.$/,"");
      if(canonical==="-0")canonical="0";
    }
    const matches=choicesFor(r).filter(o=>normalize(o.name)===canonical);
    const id=r.value_type==="number_unit"?"":f.value_id||(matches.length===1?String(matches[0].id):"");
    return [r.id,id?"id:"+id:"value:"+canonical,f.unit_id??""];
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
        <label className="listing-field">型号（基础参考）
          <input disabled={disabled} value={value.model} maxLength={120} onChange={e=>set("model",e.target.value)} placeholder="如有实际型号可填写"/>
          <small>{modelHint}</small>
        </label>
        {field("主商品 SKU（内部编码）","seller_sku","例如：FLOOR-001")}
        {field("条码 GTIN / EAN（有则填）","gtin","不编造条码",50)}
      </div>
    </fieldset>

    <CategoryAttributesForm category={value.category_id} values={value.category_attributes ?? []}
      onChange={fields => set("category_attributes",fields)} disabled={disabled} onRules={setCategoryRules} hasVariants={variants.length>0}/>

    <fieldset className="listing-editor-plain" aria-label="采购成本与目标净收益">
      <div className="listing-field-grid">
      <label className="listing-field">采购成本（¥）
        <input value={legacyPurchase ? "" : value.purchase_price} disabled={disabled}
          placeholder={legacyPurchase ? "核实后输入人民币金额" : "人民币金额"}
          onChange={event=>onChange({...value,purchase_price:event.target.value,purchase_currency:"CNY"})}/>
      </label>
      {field("目标净收益（USD）","target_net_proceeds_usd","目标净回款",21)}
      </div>
      <small className="listing-fields-note">用于填写刊登时的目标净回款。订单结算金额和利润另行计算。</small>
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
        <button type="button" disabled={disabled||variants.length<2||!variants.some(v=>v.image_urls.length>1)} onClick={()=>beginReuse("bulk")}>批量复用副图</button>
        <button type="button" disabled={disabled||duplicateCount===0}
          onClick={()=>{set("images",removeDuplicateGalleryLinks(images));setImagePreview(null);setEditingImage(null);}}>
          清理重复链接 {duplicateCount>0?"("+duplicateCount+")":""}
        </button>
        <button type="button" disabled={disabled||images.length>=60}
          onClick={()=>{set("images",[...images,{url:"",type:images.length?"additional":"main"}]);setEditingImage(images.length);}}>
          + 添加图片链接
        </button>
      </div>
      {reuseMode==="bulk"&&reusePanel}
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
          onClick={()=>set("variants",[...variants,{seller_sku:"",attributes:{},price:"",currency:"",image_urls:[],
            purchase_cost_cny:"",target_net_proceeds_usd:"",category_attributes:[]}])}>
          + 添加变体
        </button>
      </div>
      {activeRules&&variants.length>0&&<p className="listing-fields-note">家族共用属性请在 CBT 官方属性中填写。刊登前核对家族名称、类目和商品状态一致。</p>}
      <p className="listing-fields-note">{!activeRules?"选择 CBT 类目并加载官方规则后可填写 UP 规格。":childRules.length===0?"当前规则未返回 CHILD_PK，保留原始规格，需在刊登前核实。":"官方规格使用英文及官方 ID，中文仅辅助阅读。原始规格不自动提交。"}</p>
      {variants.length>0&&<p className="listing-fields-note">规格警告不阻止保存工作稿；有未核实项或重复组合时，不代表已通过正式刊登校验，刊登前必须修正。</p>}
      {childRules.some(hasChoices)&&<div className="listing-variant-rule-search">
        {childRules.filter(hasChoices).map(rule=><label className="listing-field" key={rule.id}>
          {bilingual(rule.name||rule.id)}{mandatory(rule)&&" · 必填"}
          <input aria-label={"搜索规格 "+rule.id} disabled={disabled} placeholder="英文 / 中文"
            value={specSearch[rule.id]??""} onChange={e=>setSpecSearch(old=>({...old,[rule.id]:e.target.value}))}/>
        </label>)}
      </div>}
      <div className="listing-variant-list">
        {variants.map((v,i)=><article className="listing-variant-card" key={i} aria-label={"变体 "+(i+1)}>
          <div className="listing-variant-card-heading">
            <strong>{variantLabel(v,i)}</strong>
            <small>来源价格：{v.price? v.price+" "+(v.currency||"币种未确认"):"—"}</small>
            <button type="button" disabled={disabled} onClick={()=>{
              set("variants",variants.filter((_,j)=>j!==i));
              setVariantImageIndex(null);setReuseMode(null);setReuseMessage("");
            }}>删除</button>

          </div>
          <div className="listing-variant-card-fields">
            <label className="listing-field">自定义 SKU<input disabled={disabled} aria-label={"变体 "+(i+1)+" SKU"} value={v.seller_sku}
              maxLength={120} onChange={e=>updateVariant(i,{seller_sku:e.target.value})} placeholder="自定义 SKU"/></label>
            <label className="listing-field">采购成本（¥）<input disabled={disabled} aria-label={"变体 "+(i+1)+" 采购成本（¥）"} value={v.purchase_cost_cny??""}
              inputMode="decimal" maxLength={21} onChange={e=>updateVariant(i,{purchase_cost_cny:e.target.value})} placeholder="CNY"/></label>
            <label className="listing-field">目标净收益（USD）<input disabled={disabled} aria-label={"变体 "+(i+1)+" 目标净收益（USD）"} value={v.target_net_proceeds_usd??""}
              inputMode="decimal" maxLength={21} onChange={e=>updateVariant(i,{target_net_proceeds_usd:e.target.value})} placeholder="USD"/></label>
          </div>
            <div className="listing-variant-image-action"><button type="button" className="listing-variant-images-button" aria-expanded={variantImageIndex===i} disabled={disabled}
                onClick={()=>{setVariantImageIndex(variantImageIndex===i?null:i);setVariantImageUrlDraft("");setReuseMode(null);setReuseMessage("");}}>
                {variantImageIndex===i?"收起图片":"选择图片"} ({v.image_urls.length})
              </button>
              <div className="listing-variant-mini-images">
                {v.image_urls.slice(0,3).map((url,j)=><img key={j} src={url} alt={"变体 "+(i+1)+" 图片 "+(j+1)} loading="lazy" referrerPolicy="no-referrer"/>)}
              </div>
            </div>
          {variantImageIndex===i&&<section className="listing-variant-image-picker" aria-label={"变体 "+(i+1)+" 图片选择"}>
        <div className="listing-variant-image-picker-head">
          <strong>{variantLabel(v,i)} · 选择图片</strong>
          <button type="button" disabled={disabled||!v.image_urls.length||!variants.some((other,j)=>j!==i&&other.image_urls.length>1)}
            onClick={()=>beginReuse("single",i)}>复用副图</button>
          <button type="button" onClick={()=>{setVariantImageIndex(null);setReuseMode(null);setReuseMessage("");}}>完成</button>
        </div>
        {reuseMode==="single"&&reusePanel}
        <div className="listing-variant-url-row">
          <label>变体主图
            <select disabled={disabled||v.image_urls.length===0}
              value={v.image_urls[0]??""}
              onChange={e=>updateVariant(i,{image_urls:[e.target.value,...v.image_urls.filter(url=>url!==e.target.value)]})}>
              {!v.image_urls.length&&<option value="">请先选择图片</option>}
              {v.image_urls.map((url,i)=><option key={url} value={url}>已选图片 {i+1}{i===0?"（主图）":""}</option>)}
            </select>
          </label>
          {v.image_urls[0]&&<img width={64} height={64} style={{objectFit:"contain"}} src={v.image_urls[0]} alt="当前变体主图" referrerPolicy="no-referrer"/>}
        </div>
        <div className="listing-variant-image-options">
          {Array.from(new Set([...v.image_urls,...images.map(im=>im.url),...variants.flatMap(other=>other.image_urls)]))
            .filter(url=>url.trim()!=="").map(url=><label key={url}>
              <input type="checkbox" disabled={disabled||(v.image_urls.length>=30&&
                !v.image_urls.includes(url))} checked={v.image_urls.includes(url)}
                onChange={e=>updateVariant(i,{
                  image_urls:toggleVariantImage(v.image_urls,url,e.target.checked)
                })}/>
              <img src={url} alt="可选变体图片" loading="lazy" referrerPolicy="no-referrer"/>
              <small>{v.image_urls[0]===url?"主图":v.image_urls.includes(url)?"副图":"未选"}</small>
            </label>)}
        </div>
        <div className="listing-variant-url-row">
          <input disabled={disabled} aria-label="追加变体图片 URL" placeholder="https://..." value={variantImageUrlDraft}
            onChange={e=>setVariantImageUrlDraft(e.target.value)}/>
          <button type="button" disabled={disabled||v.image_urls.length>=30||
            variantImageUrlDraft.trim().length>2000||!/^https:\/\/\S+$/i.test(variantImageUrlDraft.trim())}
            onClick={()=>{
              updateVariant(i,{image_urls:toggleVariantImage(
                v.image_urls,variantImageUrlDraft.trim(),true)});
              setVariantImageUrlDraft("");
            }}>添加链接</button>
        </div>
      </section>}
          <div className="listing-variant-card-specs">
            {childRules.map(rule=>{
              const f=currentSpec(v,rule);
              const issue=specIssue(f,rule);
              const choices=choicesFor(rule);
              const candidate=hasChoices(rule)&&!f?.value&&exactMatch(rule,v);
              const label="变体 "+(i+1)+" "+(rule.name||rule.id);
              return <div className="listing-variant-official" key={rule.id}>
                <strong>{bilingual(rule.name||rule.id)}{mandatory(rule)&&" · 必填"}</strong>
                {supported(rule)?<>
                  {hasChoices(rule)&&<select disabled={disabled} aria-label={label} value={f?.value_id??""}
                    aria-required={mandatory(rule)} aria-invalid={Boolean(issue)}
                    onChange={e=>{const o=choices.find(o=>String(o.id)===e.target.value);setOfficialSpec(i,rule,o?.name??"",o?String(o.id):undefined);}}>
                    <option value="">请选择官方值</option>
                    {f?.value_id&&!choices.some(o=>String(o.id)===f.value_id)&&<option value={f.value_id} disabled>{f.value}（待核实）</option>}
                    {choices.filter(o=>String(o.id)===f?.value_id||normalize(bilingual(o.name)).includes(normalize(specSearch[rule.id]??"")))
                      .map(o=><option key={String(o.id)} value={String(o.id)}>{bilingual(o.name)}</option>)}
                  </select>}
                  {!enumOnly(rule)&&<input disabled={disabled} aria-label={hasChoices(rule)?label+" 自定义值":label}
                    aria-required={mandatory(rule)} aria-invalid={Boolean(issue)}
                    value={f?.value??""} placeholder={rule.value_type==="string"?"English":"数字"}
                    inputMode={rule.value_type==="string"?"text":"decimal"}
                    maxLength={Math.min(rule.value_max_length||500,500)} onChange={e=>setOfficialSpec(i,rule,e.target.value,undefined,f?.unit_id)}/>}
                  {rule.value_type==="number_unit"&&<select disabled={disabled} aria-label={label+" unit"} value={f?.unit_id??""}
                    onChange={e=>setOfficialSpec(i,rule,f?.value??"",undefined,e.target.value||undefined)}>
                    <option value="">单位</option>
                    {f?.unit_id&&!unitsFor(rule).some(u=>u.id===f.unit_id)&&<option value={f.unit_id} disabled>{f.unit_id}（待核实）</option>}
                    {unitsFor(rule).map(u=><option key={u.id} value={u.id}>{u.name}</option>)}
                  </select>}
                </>:<small>特殊属性，保留历史值：{f?.value??"—"}</small>}
                {candidate&&<button type="button" disabled={disabled}
                  onClick={()=>setOfficialSpec(i,rule,candidate.name,String(candidate.id))}>匹配 {bilingual(candidate.name)}</button>}
                {issue&&<small className="listing-validation">{issue} · 可保存工作稿，刊登前修正</small>}
              </div>;
            })}
            <div className="listing-variant-spec">
              <strong>原始规格（参考）</strong>
              {completeKeys[i]&&completeKeys.filter(k=>k===completeKeys[i]).length>1&&<small className="listing-validation">重复官方规格组合 · 可保存工作稿，刊登前修正</small>}
              {(v.category_attributes??[]).filter(a=>a.category_id!==value.category_id||!childRules.some(r=>r.id===a.id)).map(a=><div key={a.category_id+":"+a.id}>
                <small>历史规格：{a.category_id} · {a.id} · {a.value}</small>
                <button type="button" disabled={disabled} onClick={()=>updateVariant(i,{category_attributes:(v.category_attributes??[]).filter(x=>x.category_id!==a.category_id||x.id!==a.id)})}>删除历史规格</button>
              </div>)}
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
            </div>
          </div>
        </article>)}
      </div>
      {variants.length===0&&<p className="listing-fields-note">未采集到变体，可点击“添加变体”手动填写真实规格。</p>}
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
