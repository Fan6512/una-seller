import { getAccessToken, rejectSession } from "./auth";
import { apiBase, type CollectionImage, type CollectionVariant } from "./collections";
import type { ListingFields } from "./listingFields";

export type OwnProductSummary = {
  id: string;
  source_product_id: string | null;
  source_platform: string;
  source_site: string;
  title: string;
  price: string | null;
  currency: string | null;
  main_image_url: string | null;
  updated_at: string;
};
export type OwnProduct = OwnProductSummary & {
  source_url: string;
  source_title: string;
  description: string;
  notes: string;
  category: string | null;
  bullet_points: string[];
  images: CollectionImage[];
  variants: CollectionVariant[];
  listing_fields: Partial<ListingFields>;
  revision: number;
  created_at: string;
};
export class ProductRequestError extends Error {
  constructor(public readonly status: number, message: string) {
    super(message);
  }
}
async function request<T>(path:string,signal:AbortSignal,method:"GET"|"POST"|"PUT"="GET",body?:unknown):Promise<T>{
  const fetchToken=async(force=false)=>{
    const token=await getAccessToken(force);
    return fetch(apiBase+path,{
      method,signal,cache:"no-store",
      headers:{Authorization:"Bearer "+token,...(body===undefined?{}:{"Content-Type":"application/json"})},
      body:body===undefined?undefined:JSON.stringify(body),
    });
  };
  let res:Response;
  try {
    res=await fetchToken();
    if(res.status===401&&method==="GET") {
      res=await fetchToken(true);
      if(res.status===401) {rejectSession();throw new Error("登录已失效，请重新登录");}
    }
    if(res.status===401&&method!=="GET") {
      // A lost POST/PUT must be read back before the user retries.
      throw new ProductRequestError(401,"登录状态已刷新或写入结果未知，请重新读取后确认");
    }
  } catch (error) {
    if(error instanceof TypeError) throw new Error("无法连接 Go API，请检查网络");
    throw error;
  }
  if(res.status===202) throw new ProductRequestError(202,"服务器已接收更改，但读回未确认；请刷新列表核实");
  if(!res.ok) throw new ProductRequestError(res.status,res.status===409
    ?"数据已在另一窗口变化（409），请重新加载后操作"
    :res.status===404?"商品不存在或无法访问":"操作失败（HTTP "+res.status+"）");
  return await res.json() as T;
}
export const OWN_PRODUCT_PAGE_SIZE = 24;
export async function getOwnProductPage(q:string,offset:number,signal:AbortSignal) {
  const query=new URLSearchParams({q:q.trim().slice(0,200),limit:String(OWN_PRODUCT_PAGE_SIZE+1),offset:String(offset)});
  const data=await request<{items:OwnProductSummary[]}>("/api/v1/products?"+query,signal);
  return {items:data.items.slice(0,OWN_PRODUCT_PAGE_SIZE),hasMore:data.items.length>OWN_PRODUCT_PAGE_SIZE};
}
export function getOwnProduct(id:string,signal:AbortSignal) {
  return request<OwnProduct>("/api/v1/products/"+encodeURIComponent(id),signal);
}
export function updateOwnProduct(id:string,data:Pick<OwnProduct,"title"|"description"|"notes"|"revision"|"listing_fields">,signal:AbortSignal) {
  return request<{product:OwnProduct}>("/api/v1/products/"+encodeURIComponent(id),signal,"PUT",{
    title:data.title,description:data.description,notes:data.notes,
    listing_fields:data.listing_fields,expected_revision:data.revision,
  });
}
export function approveCollection(id:string,revision:number,signal:AbortSignal) {
  return request<{product:OwnProduct;result:"created"|"already_approved"}>(
    "/api/v1/collections/"+encodeURIComponent(id)+"/approve",signal,"POST",{expected_revision:revision});
}
