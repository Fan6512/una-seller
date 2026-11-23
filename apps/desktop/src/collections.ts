import { getAccessToken, rejectSession } from "./auth";
import { readAPIError } from "./apiErrors";
import type { ListingFields, LocalizedTitles } from "./listingFields";

export type CaptureStatus = "complete" | "partial";
export const PAGE_SIZE = 24;

export type CollectionImage = {
  url: string;
  type: "main" | "additional";
};

export type CollectionVariant = {
  attributes: Record<string, string>;
  price: string | null;
  currency: string | null;
  image_urls: string[];
};

export type CollectionSummary = {
  id: string;
  own_product_id: string | null;
  title: string;
  source_platform: string;
  source_site: string;
  price: string | null;
  currency: string | null;
  main_image_url: string | null;
  collection_status: CaptureStatus;
  last_collected_at: string;
};

export type CollectionDetail = {
  id: string;
  own_product_id: string | null;
  title: string;
  source_platform: string;
  source_site: string;
  source_url: string;
  normalized_url: string;
  price: string | null;
  currency: string | null;
  category: string | null;
  description: string | null;
  bullet_points: string[];
  images: CollectionImage[];
  variants: CollectionVariant[];
  collection_status: CaptureStatus;
  collection_warnings: string[];
  first_collected_at: string;
  last_collected_at: string;
};

export type CollectionPage = {
  items: CollectionSummary[];
  limit: number;
  offset: number;
};

// Packaged desktop builds use the cloud API; source dev stays on loopback.
const defaultApiBase = import.meta.env.DEV ? "http://127.0.0.1:8787" : "https://mkd.186512.xyz";
export const apiBase = (import.meta.env.VITE_API_BASE_URL || defaultApiBase).replace(/\/$/, "");

class CollectionRequestError extends Error {
  constructor(readonly status: number, message: string, readonly code?: string) {
    super(message);
    this.name = "CollectionRequestError";
  }
}

async function requestJSON<T>(path: string, signal: AbortSignal, requiresAuth = true,
  method: "GET" | "DELETE" | "PUT" | "POST" = "GET", body?: string): Promise<T> {
  let response: Response;
  const fetchWithToken = (token: string | null) => fetch(apiBase + path, {
    method,
    signal,
    cache: "no-store",
    headers: { ...(token ? { "Authorization": "Bearer " + token } : {}),
      ...(body !== undefined ? { "Content-Type": "application/json" } : {}) },
    body,
  });

  try {
    const token = requiresAuth ? await getAccessToken() : null;
    response = await fetchWithToken(token);
    if (requiresAuth && response.status === 401) {
      // Retry once with a fresh session; never fall back to the legacy fixed owner.
      const renewed = await getAccessToken(true);
      response = await fetchWithToken(renewed);
      if (response.status === 401) {
        rejectSession();
        throw new Error("登录状态已失效，请重新登录");
      }
    }
  } catch (error) {
    if (signal.aborted) throw error;
    if (error instanceof TypeError) {
      throw new Error("无法连接 Go API（" + apiBase + "），请检查服务及网络");
    }
    throw error;
  }

  if (!response.ok) {
    const message = response.status === 503
      ? "采集服务暂时不可用，请检查 Go API、数据库及登录配置"
      : response.status === 404
        ? "商品不存在或不可访问"
        : "请求失败（HTTP " + response.status + "）";
    const details=await readAPIError(response,message);
    throw new CollectionRequestError(response.status,details.message,details.code);
  }
  if (response.status === 204) return undefined as T;
  return (await response.json()) as T;
}

export async function getCollectionPage(q: string, offset: number, signal: AbortSignal, inboxOnly = false) {
  const params = new URLSearchParams({
    q,
    // One extra row tells the desktop whether another page exists.
    limit: String(PAGE_SIZE + 1),
    offset: String(offset),
    inbox: inboxOnly ? "1" : "0",
  });
  const page = await requestJSON<CollectionPage>("/api/v1/collections?" + params.toString(), signal);
  return { items: page.items.slice(0, PAGE_SIZE), hasMore: page.items.length > PAGE_SIZE };
}

export type CBTAttributeRule = {
  id: string;
  name: string;
  value_type?: string;
  value_max_length?: number;
  tags?: Record<string, boolean>;
  values?: { id?: string | number; name: string }[];
  allowed_units?: { id: string; name: string }[];
};
export type CBTCategoryRules = {
  category_id: string;
  attributes: CBTAttributeRule[];
  path?: { id: string; name: string }[];
  translations_zh?: Record<string, string>;
  technical_specs_verified?: boolean;
  technical_requirements?: string[];
  technical_requirement_names?: Record<string, string>;
  special_requirements?: boolean;
};
export function getCBTCategoryRules(category: string, signal: AbortSignal) {
  return requestJSON<CBTCategoryRules>("/api/v1/stores/cbt-category-attributes?" +
    new URLSearchParams({ category_id: category }), signal);
}

export type CBTCategorySuggestion = {
  category_id: string;
  category_name: string;
  domain_id: string;
  domain_name: string;
  path?: { id: string; name: string }[];
};
export function getCBTCategorySuggestions(editedEnglishTitle:string,signal:AbortSignal){
  const q=editedEnglishTitle.trim();
  if(!q||q.split(/\s+/u).length<2)throw new Error("请先填写完整英文标题，再请求 CBT 类目推荐");
  return requestJSON<{items:CBTCategorySuggestion[];source:string;confirmed:boolean;translations_zh?:Record<string,string>}>(
    "/api/v1/categories/predict?q="+encodeURIComponent(q),signal);
}

export function getCollection(id: string, signal: AbortSignal) {
  return requestJSON<CollectionDetail>("/api/v1/collections/" + encodeURIComponent(id), signal);
}

export type AICopyProduct = ({ source_product_id: string; own_product_id?: never } |
  { own_product_id: string; source_product_id?: never }) & { expected_revision: number };

export type AICopyResults = {
  titles: { titles: LocalizedTitles };
  english: { description: string; selling_points: string[]; missing_facts: string[] };
  translate: { es: string; pt: string; review_terms: string[]; source_digest: string };
};

export function generateAICopy<Operation extends keyof AICopyResults>(
  product: AICopyProduct, operation: Operation, signal: AbortSignal,
) {
  return requestJSON<AICopyResults[Operation]>("/api/v1/products/ai-copy", signal, true, "POST",
    JSON.stringify({ ...product, operation }));
}

export type CollectionWorkingDraft = {
  source_product_id: string;
  revision: number;
  title: string;
  description: string;
  notes: string;
  listing_fields: Partial<ListingFields>;
  updated_at: string;
};
export type WorkingDraftInput = Pick<CollectionWorkingDraft, "title" | "description" | "notes"> & {
  listing_fields?: ListingFields;
  expected_revision: number;
};

export function isDraftConverted(error: unknown) {
  return error instanceof CollectionRequestError && error.status === 409 &&
    error.code === "WORKING_DRAFT_CONVERTED";
}

export function isDraftConflict(error: unknown) {
  return error instanceof CollectionRequestError && error.status === 409 &&
    !isDraftConverted(error);
}

export function isDraftRejected(error: unknown) {
  return error instanceof CollectionRequestError && [400, 413, 415].includes(error.status);
}

// Old API and missing source both return 404; check the source before diagnosing.
async function draftRequest<T>(id: string, method: "GET" | "PUT", signal: AbortSignal, body?: string) {
  const path = "/api/v1/collections/" + encodeURIComponent(id) + "/working-draft";
  try {
    return await requestJSON<T>(path, signal, true, method, body);
  } catch (error) {
    if (error instanceof CollectionRequestError && error.status === 404) {
      await getCollection(id, signal);
      throw new Error("Go 服务端尚未支持整理工作稿，请先升级 Go API 后重试。");
    }
    throw error;
  }
}

export function getWorkingDraft(id: string, signal: AbortSignal) {
  return draftRequest<{ draft: CollectionWorkingDraft | null }>(id, "GET", signal);
}

export function saveWorkingDraft(id: string, input: WorkingDraftInput, signal: AbortSignal) {
  return draftRequest<{ draft: CollectionWorkingDraft }>(id, "PUT", signal, JSON.stringify(input));
}

export function deleteCollection(id: string, signal: AbortSignal) {
  return requestJSON<void>("/api/v1/collections/" + encodeURIComponent(id), signal, true, "DELETE");
}

export function getHealth(signal: AbortSignal) {
  return requestJSON<{ status: string; service: string }>("/healthz", signal, false);
}

export function formatMoney(value: string | null, currency: string | null) {
  if (value === null) return "价格未采到";
  // DB keeps up to 4 fractional digits. Trim only trailing zeros beyond cents,
  // preserving meaningful precision without using floating-point conversion.
  const match = /^(-?\d+)\.(\d+)$/.exec(value);
  const amount = match
    ? match[1] + "." + match[2].slice(0, Math.max(2, match[2].replace(/0+$/, "").length)).padEnd(2, "0")
    : value;
  return (currency ? currency + " " : "") + amount;
}

function parseCollectedDate(value: string) {
  // Go may emit RFC3339Nano timestamps; older WebKit needs millisecond precision.
  const compatible = value.replace(/\.(\d{3})\d+(?=Z|[+-]\d{2}:\d{2}$)/, ".$1");
  return new Date(compatible);
}

export function formatDate(value: string) {
  const date = parseCollectedDate(value);
  return Number.isNaN(date.getTime()) ? "时间未知" : date.toLocaleString("zh-CN");
}

export function formatShortDate(value: string) {
  const date = parseCollectedDate(value);
  return Number.isNaN(date.getTime()) ? "时间未知" : date.toLocaleDateString("zh-CN", { month: "numeric", day: "numeric" });
}
