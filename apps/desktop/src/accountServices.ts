import { authLifetime, getAccessToken, rejectSession } from "./auth";
import { apiBase } from "./collections";

export type AccountServiceKind = "ai" | "translation";
export type AccountServiceStatus = {
  kind: AccountServiceKind;
  provider: string;
  base_url?: string;
  model?: string;
  region?: string;
  app_id?: string;
  configured: boolean;
  needs_rebind?: boolean;
};

export type AccountServiceInput = {
  provider: string;
  base_url?: string;
  model?: string;
  region?: string;
  app_id?: string;
  api_key: string;
};

const root = "/api/v1/account/service-configs";

async function accountRequest<T>(
  path: string, method: "GET" | "PUT" | "DELETE" | "POST", body?: AccountServiceInput,
  signal?: AbortSignal,
): Promise<T> {
  // Only the service backend handles and persists Key material. This module
  // does not log, cache, retain or export the body after the explicit request.
  const lifetime = authLifetime();
  lifetime.throwIfAborted();
  signal?.throwIfAborted();
  const token = await getAccessToken();
  lifetime.throwIfAborted();
  const send = (bearer: string) => fetch(apiBase + path, {
    method,
    cache: "no-store",
    signal,
    headers: {
      Authorization: "Bearer " + bearer,
      ...(body === undefined ? {} : { "Content-Type": "application/json" }),
    },
    ...(body === undefined ? {} : { body: JSON.stringify(body) }),
  });
  let response = await send(token);
  lifetime.throwIfAborted();
  if (response.status === 401) {
    // Only a read is idempotent. Do not resubmit Key changes or paid tests on
    // 401: the server might have received and committed a prior request.
    const renewed = await getAccessToken(true);
    lifetime.throwIfAborted();
    if (method !== "GET") {
      throw new Error("登录令牌已更新。请刷新服务状态，确认操作结果后再试。");
    }
    response = await send(renewed);
    if (response.status === 401) {
      rejectSession();
      throw new Error("登录已过期，请重新登录");
    }
  }
  signal?.throwIfAborted();
  if (!response.ok) {
    const details = await response.json().catch(() => null) as { error?: string } | null;
    const message = response.status === 503 ? "云端密钥功能尚未启用或暂不可用"
      : response.status === 412 ? "请先保存服务配置再测试连接"
        : response.status === 422 ? "服务商、地址、模型或密钥配置不符合规则"
          : response.status === 429 ? "请求过于频繁，请稍后再试"
            : details?.error && details.error.length < 150 ? details.error
              : "服务配置操作失败（HTTP " + response.status + "）";
    throw new Error(message);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

export async function listAccountServices(signal?: AbortSignal) {
  const statuses = await accountRequest<AccountServiceStatus[]>(root, "GET", undefined, signal);
  if (!Array.isArray(statuses)) throw new Error("账号服务配置格式不正确");
  return statuses;
}

export function saveAccountService(kind: AccountServiceKind, value: AccountServiceInput, signal?: AbortSignal) {
  return accountRequest<AccountServiceStatus>(root + "/" + kind, "PUT", value, signal);
}

export function deleteAccountService(kind: AccountServiceKind, signal?: AbortSignal) {
  return accountRequest<void>(root + "/" + kind, "DELETE", undefined, signal);
}

export function testAccountService(kind: AccountServiceKind, signal?: AbortSignal) {
  return accountRequest<{ ok: boolean }>(root + "/" + kind + "/test", "POST", undefined, signal);
}
