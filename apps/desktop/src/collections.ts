export type CaptureStatus = "complete" | "partial";

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

async function requestJSON<T>(path: string, signal: AbortSignal): Promise<T> {
  const response = await fetch(apiBase + path, { signal, cache: "no-store" });
  if (!response.ok) {
    const message = response.status === 503
      ? "采集服务尚未配置数据库及本地用户"
      : response.status === 404
        ? "商品不存在或不可访问"
        : "请求失败（HTTP " + response.status + "）";
    throw new Error(message);
  }
  return (await response.json()) as T;
}

export function getCollectionPage(q: string, offset: number, signal: AbortSignal) {
  const params = new URLSearchParams({
    q,
    limit: "24",
    offset: String(offset),
  });
  return requestJSON<CollectionPage>("/api/v1/collections?" + params.toString(), signal);
}

export function getCollection(id: string, signal: AbortSignal) {
  return requestJSON<CollectionDetail>("/api/v1/collections/" + encodeURIComponent(id), signal);
}

export function getHealth(signal: AbortSignal) {
  return requestJSON<{ status: string; service: string }>("/healthz", signal);
}

export function formatMoney(value: string | null, currency: string | null) {
  if (value === null) return "价格未采到";
  return (currency ? currency + " " : "") + value;
}

export function formatDate(value: string) {
  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? value : new Date(timestamp).toLocaleString("zh-CN");
}
