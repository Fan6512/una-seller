import type { CollectionImage, CollectionVariant } from "./collections";

export type ListingVariant = {
  seller_sku: string;
  attributes: Record<string,string>;
  price: string;
  currency: string;
  image_urls: string[];
  purchase_cost_cny?: string;
  target_net_proceeds_usd?: string;
  category_attributes?: CategoryAttributeValue[];
};
export type LocalizedTitles = { en: string; es: string; pt: string };
export type LocalizedDescriptions = LocalizedTitles;
export type TitleLanguage = keyof LocalizedTitles;
export const localizedTitleLimit = 60;
export const listingTextCount = (text: string) => Array.from(text.normalize("NFC")).length;
const listingMarkup = new RegExp("<(?:/?[A-Za-z][A-Za-z0-9:-]*(?:[ \\t\\r\\n][^<>]*)?[ \\t\\r\\n]*/?>|[!?][^<>]*>)|^[ \\t]{0,3}#{1,6}(?:[ \\t]|$)|\\[[^\\]\\n]+\\][ \\t]*(?:\\(|\\[)|(?:\\*\\*|__|~~)[^\\n]+(?:\\*\\*|__|~~)|`", "m");
export const isPlainListingText = (text: string) =>
  !listingMarkup.test(text) && !/[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F-\u009F\p{Surrogate}]/u.test(text);

export type CategoryAttributeValue = {
  category_id: string;
  id: string;
  value: string;
  value_id?: string;
  unit_id?: string;
};

export type ListingFields = {
  titles: LocalizedTitles;
  descriptions?: LocalizedDescriptions;
  description_translation?: {
    en_digest?: string;
    es_source_digest?: string;
    pt_source_digest?: string;
    es_manual_edited?: boolean;
    pt_manual_edited?: boolean;
  };
  category_id: string;
  category_path: string;
  brand: string;
  model: string;
  seller_sku: string;
  gtin: string;
  purchase_price: string;
  purchase_currency: string;
  target_net_proceeds_usd?: string;
  weight_g: string;
  length_cm: string;
  width_cm: string;
  height_cm: string;
  bullet_points: string[];
  attributes: {name:string;value:string}[];
  category_attributes?: CategoryAttributeValue[];
  images?: CollectionImage[];
  variants?: ListingVariant[];
};
export type ListingSource = {
  images: CollectionImage[];
  variants: CollectionVariant[];
  bullet_points: string[];
};

export function initialListingFields(saved?: Partial<ListingFields> | null, source?: ListingSource, capturedTitle = ""): ListingFields {
  const defaults: ListingFields = {
    titles: titlesFromSource(capturedTitle),
    category_id:"",category_path:"",brand:"",model:"",
    seller_sku:"",gtin:"",purchase_price:"",purchase_currency:"CNY",target_net_proceeds_usd:"",
    weight_g:"",length_cm:"",width_cm:"",height_cm:"",
    bullet_points:source?.bullet_points.slice()??[],
    attributes:[],
    images:source?.images.map(x=>({...x}))??[],
    variants:source?.variants.map(v=>({
      seller_sku:"",
      attributes:{...v.attributes},
      price:v.price??"",
      currency:v.currency??"",
      image_urls:v.image_urls.slice(),
    }))??[],
  };
  if(!saved)return defaults;
  const result:ListingFields={...defaults,...saved,
    titles:saved.titles ? {...defaults.titles,...saved.titles} : defaults.titles};
  if (saved.descriptions) result.descriptions = { ...saved.descriptions };
  if (saved.description_translation) result.description_translation = { ...saved.description_translation };
  // Do not silently reinterpret a legacy, unlabelled purchase amount as CNY.
  if(saved.purchase_price && !saved.purchase_currency) result.purchase_currency="";
  result.bullet_points=saved.bullet_points?.slice()??defaults.bullet_points;
  result.attributes=saved.attributes?.map(a=>({...a}))??[];
  if (saved.category_attributes) result.category_attributes = saved.category_attributes.map(field => ({ ...field }));
  if(saved.images!==undefined)result.images=saved.images.map(x=>({...x}));
  if(saved.variants!==undefined)result.variants=saved.variants.map(v=>({...v,attributes:{...v.attributes},image_urls:v.image_urls.slice(),category_attributes:v.category_attributes?.map(a=>({...a}))}));
  return result;
}
export function detectTitleLanguage(title:string): TitleLanguage | "uncertain" {
  const text=title.toLowerCase().normalize("NFC");
  if (/[ñ¿¡]/u.test(text)) return "es";
  if (/[ãõç]/u.test(text)) return "pt";
  const words=text.match(/[\p{L}]+/gu)??[];
  const spanish=new Set(["con","sin","limpiador","limpiadora","limpieza","invierno","mujer","hombre","niños","baño","zapatos","cuidado","juguete","herramienta","talla","impermeable","pantalones","lavable"]);
  const portuguese=new Set(["com","sem","limpeza","limpador","inverno","mulher","homem","crianças","banheiro","sapatos","cuidados","brinquedo","ferramenta","tamanho","impermeável","calças","lavável","não"]);
  const english=new Set(["with","for","women","men","kids","bathroom","shoes","care","toy","tool","size","waterproof","pants","washable","and","set"]);
  const score=(set:Set<string>)=>words.reduce((n,w)=>n+(set.has(w)?1:0),0);
  const es=score(spanish),pt=score(portuguese),en=score(english);
  if(es>pt && es>en) return "es";
  if(pt>es && pt>en) return "pt";
  if(en>es && en>pt) return "en";
  return "uncertain";
}
export function titlesFromSource(title:string):LocalizedTitles {
  const language=detectTitleLanguage(title);
  return {en:language==="en"||language==="uncertain"?title:"",es:language==="es"?title:"",pt:language==="pt"?title:""};
}
export function preferredTitle(titles:LocalizedTitles):string {
  return titles.en.trim() || titles.es.trim() || titles.pt.trim();
}
const isMoney=(v:string)=>v===""||/^(0|[1-9][0-9]{0,15})(\.[0-9]{1,4})?$/.test(v);
const isDecimal=(v:string,max:number,integer:boolean)=>v===""||
  (new RegExp(integer?"^[0-9]+$":"^[0-9]+(?:\\.[0-9]+)?$").test(v)&&
    Number(v)>0&&Number(v)<=max);
export function validateListingFields(v:ListingFields, previousTitles?: LocalizedTitles):string {
  if(Object.values(v.titles).some(t=>listingTextCount(t)>500))return "单个语言标题不能超过 500 字符";
  for (const lang of ["es", "pt"] as const) {
    const title = v.titles[lang].normalize("NFC");
    if (listingTextCount(title) > localizedTitleLimit && title !== previousTitles?.[lang].normalize("NFC")) {
      return "墨西哥西语与巴西葡语标题不能超过 60 字符";
    }
  }
  if (v.descriptions && Object.values(v.descriptions).some(text => listingTextCount(text) > 20000 || !isPlainListingText(text))) {
    return "每种语言的商品描述须为纯文本且不超过 20000 字符";
  }
  if(!preferredTitle(v.titles))return "请至少填写一个语言的商品标题";
  if(v.category_id && !/^CBT[0-9]{1,15}$/.test(v.category_id))return "请填写 CBT 通用类目 ID，站点类目应在上传时选择";
  if(v.category_path.length>400||v.category_id.length>60||v.brand.length>120||
    v.model.length>120||v.seller_sku.length>120||v.gtin.length>50) return "类目或基础属性过长";
  if(!isMoney(v.target_net_proceeds_usd??""))return "美元目标净收益格式不正确，最多保留四位小数";
  if(!isMoney(v.purchase_price))return "采购成本格式不正确，最多保留四位小数";
  if(v.purchase_currency && !/^[A-Z]{3}$/.test(v.purchase_currency))return "采购币种必须是 3 位大写字母（例如 CNY、USD）";
  if(!isDecimal(v.weight_g,1000000,true))return "重量请输入正整数，单位为克";
  for(const [n,k] of [["长","length_cm"],["宽","width_cm"],["高","height_cm"]] as const) {
    if(!isDecimal(v[k],10000,false))return n+"度必须是大于零的厘米数字";
  }
  if(v.images&&v.images.some(i=>!/^https:\/\/[^\s]+$/i.test(i.url)||i.url.length>2000))return "图片 URL 必须为有效 HTTPS 地址";
  if(v.images&&v.images.length>60)return "图片不能超过 60 张";
  if(v.variants&&v.variants.length>100)return "规格不能超过 100 项";
  if(v.variants?.some(x=>!isMoney(x.price)||!isMoney(x.purchase_cost_cny??"")||!isMoney(x.target_net_proceeds_usd??"")||Array.from(x.seller_sku).length>120||
    (x.currency!==""&&!/^[A-Z]{3}$/.test(x.currency))||
    (Object.keys(x.attributes).length===0&&!(x.category_attributes??[]).some(a=>a.category_id===v.category_id&&a.value.trim()))||
    Object.keys(x.attributes).length>12||
    Object.entries(x.attributes).some(([k,val])=>!k||!val||Array.from(k).length>80||Array.from(val).length>200)||
    (x.category_attributes?.length??0)>12||
    (x.category_attributes??[]).some(a=>!/^CBT[0-9]{1,15}$/.test(a.category_id)||!/^[A-Z][A-Z0-9_]{0,79}$/.test(a.id)||
      Array.from(a.value).length>500||(a.value_id?.length??0)>80||(a.unit_id?.length??0)>80||
      (Boolean(a.value_id)&&!a.value.trim()))||
    new Set((x.category_attributes??[]).map(a=>a.category_id+":"+a.id)).size!==(x.category_attributes?.length??0)||
    x.image_urls.length>30||x.image_urls.some(url=>!/^https:\/\/[^\s]+$/i.test(url)||url.length>2000))) {
    return "变体价格、币种（3 位大写字母）、SKU、属性或图片 URL 不符合要求";
  }
  if(v.bullet_points.length>20||v.bullet_points.some(x=>listingTextCount(x)>500||!isPlainListingText(x)))return "商品卖点最多 20 条，每条 500 字符，必须是纯文本";
  if(v.attributes.length>60||v.attributes.some(a=>Array.from(a.name).length>120||Array.from(a.value).length>500))
    return "附加属性最多 60 条，名称 120 字符、值 500 字符以内";
  const categoryAttributes = v.category_attributes ?? [];
  if (categoryAttributes.length > 180) return "官方属性超过 180 项，请移除不再使用的历史属性";
  const invalid = categoryAttributes.find(field => !/^CBT[0-9]{1,15}$/.test(field.category_id) ||
    !/^[A-Z][A-Z0-9_]{0,79}$/.test(field.id) || Array.from(field.value).length > 500 ||
    (field.value_id?.length ?? 0) > 80 || (field.unit_id?.length ?? 0) > 80);
  if (invalid) return invalid.id + " 官方属性格式错误";
  return "";
}
