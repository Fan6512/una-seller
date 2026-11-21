import type { CollectionImage, CollectionVariant } from "./collections";

export type ListingVariant = {
  seller_sku: string;
  attributes: Record<string,string>;
  price: string;
  currency: string;
  image_urls: string[];
};
export type ListingFields = {
  category_site: string;
  category_id: string;
  category_path: string;
  brand: string;
  model: string;
  seller_sku: string;
  gtin: string;
  purchase_price: string;
  purchase_currency: string;
  weight_g: string;
  length_cm: string;
  width_cm: string;
  height_cm: string;
  site_prices: {site:string;price:string;currency:string}[];
  bullet_points: string[];
  attributes: {name:string;value:string}[];
  images?: CollectionImage[];
  variants?: ListingVariant[];
};
export type ListingSource = {
  images: CollectionImage[];
  variants: CollectionVariant[];
  bullet_points: string[];
};

const siteDefaultPrices = [
  { site:"MLB",price:"",currency:"BRL" },
  { site:"MLM",price:"",currency:"MXN" },
];
export const SITE_NAMES:Record<string,string> = {
  MLB:"巴西",MLM:"墨西哥",MLA:"阿根廷",MLC:"智利",MCO:"哥伦比亚",
};

export function initialListingFields(saved?: Partial<ListingFields> | null, source?: ListingSource): ListingFields {
  const defaults: ListingFields = {
    category_site:"",category_id:"",category_path:"",brand:"",model:"",
    seller_sku:"",gtin:"",purchase_price:"",purchase_currency:"",
    weight_g:"",length_cm:"",width_cm:"",height_cm:"",
    site_prices:siteDefaultPrices.map(p=>({...p})),
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
  const result:ListingFields={...defaults,...saved};
  result.site_prices=saved.site_prices!==undefined?saved.site_prices.map(p=>({...p})):defaults.site_prices;
  result.bullet_points=saved.bullet_points?.slice()??defaults.bullet_points;
  result.attributes=saved.attributes?.map(a=>({...a}))??[];
  if(saved.images!==undefined)result.images=saved.images.map(x=>({...x}));
  if(saved.variants!==undefined)result.variants=saved.variants.map(v=>({...v,attributes:{...v.attributes},image_urls:v.image_urls.slice()}));
  return result;
}
const isMoney=(v:string)=>v===""||/^(0|[1-9][0-9]{0,15})(\.[0-9]{1,4})?$/.test(v);
const isDecimal=(v:string,max:number,integer:boolean)=>v===""||
  (new RegExp(integer?"^[0-9]+$":"^[0-9]+(?:\\.[0-9]+)?$").test(v)&&
    Number(v)>0&&Number(v)<=max);
export function validateListingFields(v:ListingFields):string {
  if(v.category_path.length>400||v.category_id.length>60||v.brand.length>120||
    v.model.length>120||v.seller_sku.length>120||v.gtin.length>50) return "类目或基础属性过长";
  if(!isMoney(v.purchase_price)||v.site_prices.some(p=>!isMoney(p.price)))return "采购价或站点售价格式不正确，最多保留四位小数";
  if(!isDecimal(v.weight_g,1000000,true))return "重量请输入正整数，单位为克";
  for(const [n,k] of [["长","length_cm"],["宽","width_cm"],["高","height_cm"]] as const) {
    if(!isDecimal(v[k],10000,false))return n+"度必须是大于零的厘米数字";
  }
  if(v.images&&v.images.some(i=>!/^https:\/\/[^\s]+$/i.test(i.url)||i.url.length>2000))return "图片 URL 必须为有效 HTTPS 地址";
  if(v.images&&v.images.length>60)return "图片不能超过 60 张";
  if(v.variants&&v.variants.length>100)return "规格不能超过 100 项";
  if(v.variants?.some(x=>!isMoney(x.price)||x.seller_sku.length>120||Object.entries(x.attributes).some(([k,val])=>!k||!val)))return "请补充规格属性和正确的 SKU、价格";
  if(v.bullet_points.length>20||v.bullet_points.some(x=>x.length>500))return "商品卖点最多 20 条，每条 500 字符";
  if(v.attributes.length>60)return "附加属性过多";
  return "";
}
