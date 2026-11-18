import { getAccessToken, rejectSession } from "./auth";
import { apiBase } from "./collections";

export type StoreConnection = {
  id: string;
  seller_id: string;
  nickname: string;
  status: "connected" | "token_expired" | "reauth_required" | "refresh_unknown" | "disconnected";
  token_expires_at?: string | null;
  updated_at: string;
};

async function storeRequest<T>(path: string, method: "GET" | "POST", signal?: AbortSignal): Promise<T> {
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
    if (response.status === 503) throw new Error("美客多授权尚未在服务器上配置或服务不可用");
    if (response.status === 404) throw new Error("店铺不存在或无权访问");
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
