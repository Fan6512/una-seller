import { authLifetime, getAccessToken, rejectSession } from "./auth";
import { apiBase } from "./collections";
import { beginRequestDiagnostic } from "./DevelopmentDiagnostics";
import type { Listing } from "./PublishedListingView";

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

export type Review = { listing_id?: string; offset: number; reason: string; skippable: boolean };
export type ListingPage = { items: Listing[]; has_more: boolean; synced_at: string | null; next_offset: number | null; scan_required: boolean; review?: Review };

export class StoreRequestError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(message: string, status: number, code = "") {
    super(message);
    this.status = status;
    this.code = code;
  }
}

const cacheTTL = 60_000;
const refreshAhead = 60_000;
type StoreSession = {
  stores?: { items: StoreConnection[]; expires: number };
  storesRead?: Promise<StoreConnection[]>;
  storesRevision: number;
  storeListeners: Set<(items: StoreConnection[]) => void>;
  checks: Map<string, Promise<{ status: string; marketplaces_verified: boolean }>>;
  unconfirmedChecks: Set<string>;
  pages: Map<string, { page: ListingPage; expires: number }>;
  listingRevision: number;
};
const storeSessions = new WeakMap<AbortSignal, StoreSession>();

function storeSession(lifetime = authLifetime()) {
  lifetime.throwIfAborted();
  let session = storeSessions.get(lifetime);
  if (!session) {
    const fresh: StoreSession = {
      storesRevision: 0, storeListeners: new Set(), checks: new Map(), unconfirmedChecks: new Set(), pages: new Map(), listingRevision: 0,
    };
    storeSessions.set(lifetime, fresh);
    lifetime.addEventListener("abort", () => {
      fresh.stores = undefined;
      fresh.pages.clear();
      fresh.checks.clear();
      fresh.unconfirmedChecks.clear();
      fresh.storeListeners.clear();
      storeSessions.delete(lifetime);
    }, { once: true });
    session = fresh;
  }
  return session;
}

export function listingErrorMessage(code: string, platformStatus?: number): string {
  switch (code) {
    case "description_unavailable": return "平台暂未提供该商品的完整描述";
    case "description_link_missing": return "这条商品缺少可读取描述的关联标识";
    case "listing_unavailable": return "商品服务暂不可用，请稍后重试";
    case "listing_method": return "此接口不支持当前请求方式";
    case "category_invalid": return "类目 ID 与当前站点不匹配";
    case "category_unavailable": return "无法读取该类目的平台属性规则，请核对分类或稍后重试";
    case "listing_store_missing": return "店铺不存在或无权访问";
    case "listing_site_unverified": return "该站点尚未验证";
    case "listing_reauthorization": return "此店铺需要重新授权";
    case "listing_platform_auth": return "美客多拒绝商品访问，请在店铺管理验证连接和权限";
    case "listing_token_expired": return "店铺访问令牌已到期，请重新发起操作以先检查续签状态";
    case "listing_refresh_unknown": return "店铺 Token 续签状态待确认，请在店铺管理处理";
    case "listing_auth_changed": return "店铺授权在请求期间变化，请重试";
    case "listing_cursor_changed": return "同步位置已更新，请按最新位置继续";
    case "listing_busy": return "此店铺已有同步在执行，请稍后继续";
    case "listing_rate_limited": return "美客多限制请求频率，请稍后继续";
    case "listing_timeout": return "商品请求超时，请刷新后以服务端同步位置为准";
    case "listing_commit_unknown": return "同步提交结果待确认，请刷新后以服务端同步位置为准";
    case "listing_transport": return "连接美客多时网络中断，请重试";
    case "listing_identity_mismatch": return "刊登归属与授权范围不一致，需要人工核实";
    case "listing_review_changed": return "待核实商品或位置已变化，请重新核实后继续";
    case "listing_canceled": return "已停止读取，可从已保存位置继续";
    case "listing_invalid_response": return "美客多返回了无法识别的刊登数据，需要人工核实";
    default: return platformStatus ? "美客多商品请求失败（HTTP " + platformStatus + "）" : "商品请求失败";
  }
}

export function readableStoreError(error: unknown, fallback: string): string {
  if (error instanceof Error && error.name === "TimeoutError") {
    return "请求超时，请检查网络后重试";
  }
  return error instanceof Error ? error.message : fallback;
}

export async function storeRequest<T>(path: string, method: "GET" | "POST", signal?: AbortSignal, listing = false, body?: string,
  errorLabel?: (code: string, status: number) => string): Promise<T> {
  const lifetime = authLifetime();
  const diagnostic = beginRequestDiagnostic(path, method);
  const assertActive = () => {
    lifetime.throwIfAborted();
    signal?.throwIfAborted();
  };
  const unreadable = () => new Error(method === "POST"
    ? "未收到可读取的同步响应，请刷新后以服务端同步位置为准"
    : "无法读取服务端商品和同步位置，请重试读取");
  const request = async (refresh = false) => {
    let duringPreflight = true;
    try {
      assertActive();
      const token = await getAccessToken(refresh, diagnostic
        ? (cached) => diagnostic("token_preflight", undefined, undefined, cached) : undefined);
      duringPreflight = false;
      assertActive();
      const response = await fetch(apiBase + path, {
        method,
        signal,
        cache: "no-store",
        headers: body === undefined ? { Authorization: "Bearer " + token } : {
          Authorization: "Bearer " + token, "Content-Type": "application/json",
        },
        body,
      });
      assertActive();
      diagnostic?.("response_headers", response);
      return response;
    } catch (error) {
      diagnostic?.("failed", undefined,
        signal?.reason?.name === "TimeoutError" ? "timeout" :
        lifetime.aborted || signal?.aborted ? "aborted" :
        duringPreflight ? "token_preflight" : "network");
      assertActive();
      if (listing && error instanceof TypeError) throw unreadable();
      if (error instanceof TypeError) throw new Error("无法连接店铺服务，请检查网络后重试");
      throw error;
    }
  };
  let response = await request();
  assertActive();
  if (response.status === 401) {
    if (method === "POST") {
      // A POST cannot be retried automatically after an unknown commit.
      diagnostic?.("failed", response);
      await getAccessToken(true, diagnostic
        ? (cached) => diagnostic("token_preflight", undefined, undefined, cached) : undefined);
      assertActive();
      throw new StoreRequestError("登录令牌已刷新，请确认状态后重新发起操作", 401);
    }
    response = await request(true);
    assertActive();
    if (response.status === 401) {
      diagnostic?.("failed", response);
      rejectSession();
      throw new Error("登录已过期，请重新登录");
    }
  }
  if (!response.ok) {
    if (listing) {
      const detail = await response.json().catch((error: unknown) => {
        assertActive();
        if (error instanceof Error && error.name === "AbortError") throw error;
        return null;
      }) as { code?: string; platform_status?: number; skipped?: number } | null;
      diagnostic?.("failed", response, detail?.code);
      const reason = detail?.code ? listingErrorMessage(detail.code, detail.platform_status)
        : "商品请求失败（HTTP " + response.status + "）";
      assertActive();
      throw new StoreRequestError(detail?.skipped ? reason + "，该批已跳过 " + detail.skipped + " 条" : reason, response.status);
    }
    if (errorLabel) {
      const detail = await response.json().catch(() => null) as { error?: string } | null;
      assertActive();
      diagnostic?.("failed", response, detail?.error);
      const code = detail?.error || "order_unavailable";
      throw new StoreRequestError(errorLabel(code, response.status), response.status, code);
    }
    diagnostic?.("failed", response);
    const reason = ({
      503: "暂时无法检查店铺连接，请稍后验证",
      404: "店铺不存在或无权访问",
      403: "此店铺需要重新授权",
      409: "续签结果待确认，请在店铺管理检查后处理",
    } as Record<number, string>)[response.status];
    throw new StoreRequestError(reason ?? "店铺操作失败（HTTP " + response.status + "）", response.status);
  }
  try {
    const result = await response.json() as T;
    assertActive();
    diagnostic?.("response_body", response, (result as { interruption?: { code?: unknown } } | null)?.interruption?.code);
    return result;
  } catch (error) {
    diagnostic?.("failed", response, signal?.reason?.name === "TimeoutError" ? "timeout" :
      lifetime.aborted || signal?.aborted ? "aborted" : "invalid_body");
    assertActive();
    if (listing && !(error instanceof Error && error.name === "AbortError")) throw unreadable();
    throw error;
  }
}

// Cache reads are safe even while a sign-out abort listener is running.
export function cachedStores() {
  const lifetime = authLifetime();
  if (lifetime.aborted) return undefined;
  const cached = storeSession(lifetime).stores;
  return cached && cached.expires > Date.now() ? cached.items : undefined;
}

export function watchStoreList(listener: (items: StoreConnection[]) => void) {
  const lifetime = authLifetime();
  if (lifetime.aborted) return () => {};
  const { storeListeners } = storeSession(lifetime);
  storeListeners.add(listener);
  return () => { storeListeners.delete(listener); };
}

export function listStores(fresh = false): Promise<StoreConnection[]> {
  const lifetime = authLifetime();
  const session = storeSession(lifetime);
  if (fresh) {
    session.storesRevision++;
    session.storesRead = undefined;
  }
  const revision = session.storesRevision;
  session.storesRead ??= storeRequest<{ items: StoreConnection[] }>(
    "/api/v1/stores", "GET", AbortSignal.any([lifetime, AbortSignal.timeout(8_000)]),
  ).then(({ items }) => {
    if (session.storesRevision !== revision) return listStores();
    for (const previous of session.stores?.items ?? []) {
      const latest = items.find((item) => item.id === previous.id);
      if (!latest || ((previous.status === "connected" || previous.status === "token_expired") &&
        latest.status !== "connected" && latest.status !== "token_expired")) {
        invalidateListingPages(previous.id, lifetime);
      }
      if (!latest || latest.status === "disconnected") session.unconfirmedChecks.delete(previous.id);
    }
    session.stores = { items, expires: Date.now() + cacheTTL };
    for (const listener of session.storeListeners) listener(items);
    return items;
  }).finally(() => {
    if (session.storesRevision === revision) session.storesRead = undefined;
  });
  return session.storesRead;
}

export async function startStoreAuthorization() {
  const response = await storeRequest<{ authorization_url: string }>("/api/v1/stores/oauth/start", "POST");
  const url = new URL(response.authorization_url);
  if (url.origin !== "https://global-selling.mercadolibre.com" || url.pathname !== "/authorization") {
    throw new Error("服务器返回的授权地址无效");
  }
  return response.authorization_url;
}

export async function disconnectStore(id: string) {
  const lifetime = authLifetime();
  const session = storeSession(lifetime);
  await storeRequest<{ status: string }>("/api/v1/stores/" + encodeURIComponent(id) + "/disconnect", "POST");
  lifetime.throwIfAborted();
  session.storesRevision++;
  session.storesRead = undefined;
  const remaining = session.stores?.items.filter((item) => item.id !== id) ?? [];
  session.stores = { items: remaining, expires: Date.now() + cacheTTL };
  session.unconfirmedChecks.delete(id);
  invalidateListingPages(id, lifetime);
  for (const listener of session.storeListeners) listener(remaining);
}

export function checkStoreConnection(id: string, manual = true) {
  const lifetime = authLifetime();
  const session = storeSession(lifetime);
  const active = session.checks.get(id);
  if (active) return active;

  const read = async () => (await listStores(true)).find((item) => item.id === id);
  const check = (async () => {
    const unconfirmed = session.unconfirmedChecks.has(id);
    if (manual || unconfirmed) {
      const item = await read();
      assertStoreCanRenew(item);
      if (unconfirmed && item.status === "connected" && Date.parse(item.token_expires_at ?? "") > Date.now() + refreshAhead) {
        session.unconfirmedChecks.delete(id);
        return { status: item.status, marketplaces_verified: item.marketplaces_status === "verified" };
      }
    }

    // The server lease guards rotation even when the client stops waiting.
    let confirmedResponse = false;
    let requestError: unknown;
    try {
      await storeRequest<{ status: string }>(
        "/api/v1/stores/" + encodeURIComponent(id) + "/connection-check", "POST",
        AbortSignal.any([lifetime, AbortSignal.timeout(25_000)]),
      );
      confirmedResponse = true;
    } catch (error) {
      lifetime.throwIfAborted();
      requestError = error;
    }
    session.unconfirmedChecks.add(id);
    const item = await read().catch(() => undefined);
    lifetime.throwIfAborted();
    if (item?.status === "connected" &&
        Date.parse(item.token_expires_at ?? "") > Date.now() + (confirmedResponse ? 10_000 : refreshAhead)) {
      session.unconfirmedChecks.delete(id);
      return { status: item.status, marketplaces_verified: item.marketplaces_status === "verified" };
    }
    if (item) assertStoreCanRenew(item);
    if (requestError instanceof StoreRequestError) throw requestError;
    throw new Error("续签结果尚未确认，请检查网络后在店铺管理点击验证连接");
  })().finally(() => session.checks.delete(id));
  session.checks.set(id, check);
  return check;
}

function assertStoreCanRenew(item: StoreConnection | undefined): asserts item is StoreConnection {
  if (!item || item.status === "disconnected") throw new Error("店铺已断开或无权访问");
  if (item.status === "reauth_required") throw new Error(listingErrorMessage("listing_reauthorization"));
  if (item.status === "refresh_unknown") throw new Error(listingErrorMessage("listing_refresh_unknown"));
}

export async function ensureStoreConnection(id: string, signal?: AbortSignal) {
  const lifetime = authLifetime();
  const session = storeSession(lifetime);
  signal?.throwIfAborted();
  const item = (session.stores?.items ?? await listStores()).find((store) => store.id === id);
  lifetime.throwIfAborted();
  signal?.throwIfAborted();
  assertStoreCanRenew(item);
  if (!session.unconfirmedChecks.has(id) && item.status === "connected" &&
      Date.parse(item.token_expires_at ?? "") > Date.now() + refreshAhead) return;

  const result = checkStoreConnection(id, false);
  if (!signal) {
    await result;
  } else {
    await new Promise<void>((resolve, reject) => {
      const onAbort = () => reject(signal.reason);
      signal.addEventListener("abort", onAbort, { once: true });
      result.then(() => resolve(), reject).finally(() => signal.removeEventListener("abort", onAbort));
    });
  }
  lifetime.throwIfAborted();
  signal?.throwIfAborted();
}

export function cachedListingPage(path: string) {
  const lifetime = authLifetime();
  if (lifetime.aborted) return undefined;
  const entry = storeSession(lifetime).pages.get(path);
  return entry && entry.expires > Date.now() ? entry.page : undefined;
}

export async function loadListingPage(path: string, signal: AbortSignal) {
  const session = storeSession();
  const revision = session.listingRevision;
  const page = await storeRequest<ListingPage>(path, "GET", signal, true);
  signal.throwIfAborted();
  if (session.listingRevision === revision) {
    session.pages.delete(path);
    session.pages.set(path, { page, expires: Date.now() + cacheTTL });
    if (session.pages.size > 24) session.pages.delete(session.pages.keys().next().value!);
  }
  return page;
}

export function invalidateListingPages(id: string, lifetime = authLifetime()) {
  if (lifetime.aborted) return;
  const session = storeSession(lifetime);
  session.listingRevision++;
  const prefix = "/api/v1/stores/" + encodeURIComponent(id) + "/";
  for (const path of session.pages.keys()) if (path.startsWith(prefix)) session.pages.delete(path);
}

export function listingAccessDenied(error: unknown) {
  return error instanceof StoreRequestError && (error.status === 403 || error.status === 404);
}
