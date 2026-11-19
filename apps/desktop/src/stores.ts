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

export function listingErrorMessage(code: string, platformStatus?: number): string {
  switch (code) {
    case "listing_store_missing": return "店铺不存在或无权访问";
    case "listing_site_unverified": return "该站点尚未验证";
    case "listing_reauthorization": return "此店铺需要重新授权";
    case "listing_platform_auth": return "美客多拒绝商品访问，请在店铺管理验证连接和权限";
    case "listing_token_expired": return "店铺 Token 已过期，请在店铺管理验证连接";
    case "listing_refresh_unknown": return "店铺 Token 续签状态待确认，请在店铺管理处理";
    case "listing_auth_changed": return "店铺授权在同步期间变化，请验证连接后继续";
    case "listing_cursor_changed": return "同步位置已更新，请按最新位置继续";
    case "listing_busy": return "此店铺已有同步在执行，请稍后继续";
    case "listing_rate_limited": return "美客多限制请求频率，请稍后继续";
    case "listing_timeout": return "美客多读取超时，可从已保存位置继续";
    case "listing_transport": return "美客多网络请求中断，可从已保存位置继续";
    case "listing_identity_mismatch": return "刊登归属与授权范围不一致，需要人工核实";
    case "listing_review_changed": return "待核实商品或位置已变化，请重新核实后继续";
    case "listing_canceled": return "已停止读取，可从已保存位置继续";
    case "listing_invalid_response": return "美客多返回了无法识别的刊登数据，需要人工核实";
    default: return platformStatus ? "美客多商品请求失败（HTTP " + platformStatus + "）" : "商品请求失败";
  }
}

export async function storeRequest<T>(path: string, method: "GET" | "POST", signal?: AbortSignal, listing = false, body?: string): Promise<T> {
  const request = (token: string) => fetch(apiBase + path, {
    method,
    signal,
    cache: "no-store",
    headers: body === undefined ? { Authorization: "Bearer " + token } : {
      Authorization: "Bearer " + token, "Content-Type": "application/json",
    },
    body,
  });
  let response = await request(await getAccessToken());
  if (response.status === 401) {
    signal?.throwIfAborted();
    response = await request(await getAccessToken(true));
    if (response.status === 401) {
      rejectSession();
      throw new Error("登录已过期，请重新登录");
    }
  }
  if (!response.ok) {
    if (listing) {
      const detail = await response.json().catch(() => null) as { code?: string; platform_status?: number; skipped?: number } | null;
      const reason = detail?.code ? listingErrorMessage(detail.code, detail.platform_status)
        : "商品请求失败（HTTP " + response.status + "）";
      throw new Error(detail?.skipped ? reason + "，该批已跳过 " + detail.skipped + " 条" : reason);
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
