import { getAccessToken, rejectSession } from "./auth";
import { apiBase } from "./collections";

export type StoreMarketplace = {
  user_id: number | string;
  site_id: string;
  logistic_type: string;
};

export type StoreConnection = {
  id: string;
  seller_id: string;
  nickname: string;
  status: "connected" | "token_expired" | "reauth_required" | "refresh_unknown" | "disconnected";
  token_expires_at?: string | null;
  updated_at: string;
  marketplaces_status: "verified" | "unverified";
  marketplaces: StoreMarketplace[];
};

export async function storeRequest<T>(path: string, method: "GET" | "POST", signal?: AbortSignal, listing = false): Promise<T> {
  const request = (token: string) => fetch(apiBase + path, {
    method,
    signal,
    cache: "no-store",
    headers: { Authorization: "Bearer " + token },
  });
  let response = await request(await getAccessToken());
  if (response.status === 401) {
    response = await request(await getAccessToken(true));
    if (response.status === 401) {
      rejectSession();
      throw new Error("登录已过期，请重新登录");
    }
  }
  if (!response.ok) {
    if (listing) {
      const detail = await response.json().catch(() => null) as { error?: string } | null;
      const known: Record<string, string> = {
        "store not found": "店铺不存在或无权访问",
        "marketplace site is not verified": "该站点尚未验证",
        "store needs reauthorization": "此店铺需要重新授权",
        "store token expired; check connection in store management": "店铺 Token 已过期，请在店铺管理验证连接",
        "store token refresh outcome unknown": "店铺 Token 续签状态待确认",
        "listing sync already running": "此店铺已有同步在执行",
        "Mercado Libre rate limited listing sync": "美客多限制请求频率，请稍后重试",
      };
      const platform = detail?.error?.match(/^marketplace listing request returned HTTP (\d{3})$/);
      if (platform) throw new Error("美客多拒绝商品请求（HTTP " + platform[1] + "）");
      throw new Error((detail?.error && known[detail.error]) || "商品请求失败（HTTP " + response.status + "）");
    }
    if (response.status === 503) throw new Error("美客多授权尚未在服务器上配置或服务不可用");
    if (response.status === 404) throw new Error("店铺不存在或无权访问");
    if (response.status === 403) throw new Error("此店铺需要重新授权");
    if (response.status === 409) throw new Error("Token 续签结果不确定，需要重新授权");
    throw new Error("店铺操作失败（HTTP " + response.status + "）");
  }
  return await response.json() as T;
}

export async function listStores(signal?: AbortSignal) {
  const response = await storeRequest<{ items: StoreConnection[] }>("/api/v1/stores", "GET", signal);
  return response.items;
}

export async function startStoreAuthorization() {
  const response = await storeRequest<{ authorization_url: string }>("/api/v1/stores/oauth/start", "POST");
  const url = new URL(response.authorization_url);
  if (url.origin !== "https://global-selling.mercadolibre.com" || url.pathname !== "/authorization") {
    throw new Error("服务器返回的授权地址无效");
  }
  return response.authorization_url;
}

export function disconnectStore(id: string) {
  return storeRequest<{ status: string }>("/api/v1/stores/" + encodeURIComponent(id) + "/disconnect", "POST");
}

export function checkStoreConnection(id: string) {
  return storeRequest<{ status: string; seller_id: string; marketplaces_verified: boolean }>(
    "/api/v1/stores/" + encodeURIComponent(id) + "/connection-check", "POST"
  );
}
