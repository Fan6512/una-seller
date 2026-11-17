import { getAccessToken, rejectSession } from "./auth";

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

export const apiBase = (import.meta.env.VITE_API_BASE_URL || "http://127.0.0.1:8787").replace(/\/$/, "");

async function requestJSON<T>(path: string, signal: AbortSignal, requiresAuth = true): Promise<T> {
  let response: Response;
  const fetchWithToken = (token: string | null) => fetch(apiBase + path, {
    signal,
    cache: "no-store",
    headers: token ? { "Authorization": "Bearer " + token } : undefined,
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
    throw new Error(message);
  }
  return (await response.json()) as T;
}

export async function getCollectionPage(q: string, offset: number, signal: AbortSignal) {
  const params = new URLSearchParams({
    q,
    // One extra row tells the desktop whether another page exists.
    limit: String(PAGE_SIZE + 1),
    offset: String(offset),
  });
  const page = await requestJSON<CollectionPage>("/api/v1/collections?" + params.toString(), signal);
  return { items: page.items.slice(0, PAGE_SIZE), hasMore: page.items.length > PAGE_SIZE };
}

export function getCollection(id: string, signal: AbortSignal) {
  return requestJSON<CollectionDetail>("/api/v1/collections/" + encodeURIComponent(id), signal);
}

export function getHealth(signal: AbortSignal) {
  return requestJSON<{ status: string; service: string }>("/healthz", signal, false);
}

export function formatMoney(value: string | null, currency: string | null) {
  if (value === null) return "价格未采到";
  return (currency ? currency + " " : "") + value;
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
