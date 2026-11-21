import { authLifetime, getAccessToken } from "./auth";
import { apiBase } from "./collections";
import { ORDER_WRITE_TIMEOUT_MS, OrderRequestError, orderRead, type OrderScope } from "./orders";
import { beginRequestDiagnostic } from "./DevelopmentDiagnostics";

export type OrderPurchase = {
  order_id: string;
  purchase_order_id: string;
  cost_cny: string | null;
  source: "manual" | "extension";
  warehouse_address?: string;
  purchase_tracking_number?: string;
  updated_at: string;
};

function purchasePath(scope: OrderScope, orderId: string) {
  if (!/^\d{1,30}$/.test(orderId)) throw new Error("美客多订单号不正确");
  return "/api/v1/stores/" + encodeURIComponent(scope.storeId) +
    "/orders/" + encodeURIComponent(orderId) + "/purchase?" +
    new URLSearchParams({ site: scope.site, logistic_type: scope.logistic });
}

export async function loadPurchase(scope: OrderScope, orderId: string, signal?: AbortSignal) {
  return orderRead<{ purchase: OrderPurchase | null; warehouse_addresses?: string[] }>(purchasePath(scope, orderId), signal);
}

// Explicit user action only. Do not automatically repeat a POST after ambiguous network failure.
export async function savePurchase(scope: OrderScope, orderId: string, purchaseId: string,
  costCNY: string, updatedAt: string | null, signal: AbortSignal, warehouse = "", tracking = "") {
  if (!/^\d{8,40}$/.test(purchaseId)) throw new Error("1688 采购订单号应为 8–40 位数字");
  if (costCNY && !/^(0|[1-9]\d{0,11})(\.\d{1,2})?$/.test(costCNY)) {
    throw new Error("采购金额应为非负人民币金额，最多保留两位小数");
  }
  if (warehouse.length > 300 || /[\r\n\0]/.test(warehouse)) {
    throw new Error("货代仓地址最长 300 字，不能包含换行");
  }
  if (tracking && !/^[A-Za-z0-9][A-Za-z0-9_/-]{2,79}$/.test(tracking)) {
    throw new Error("国内采购运单号应为 3–80 位字母、数字、横线或斜线");
  }
  const diagnostic = beginRequestDiagnostic(purchasePath(scope, orderId), "POST");
  const lifetime = authLifetime();
  const timedSignal = AbortSignal.any([signal, AbortSignal.timeout(ORDER_WRITE_TIMEOUT_MS)]);
  lifetime.throwIfAborted();
  let token: string;
  try {
    token = await getAccessToken(false, diagnostic
      ? (cached) => diagnostic("token_preflight", undefined, undefined, cached) : undefined);
  } catch (error) {
    diagnostic?.("failed", undefined, "token_preflight");
    throw error;
  }
  lifetime.throwIfAborted();
  timedSignal.throwIfAborted();
  let response: Response;
  try {
    response = await fetch(apiBase + purchasePath(scope, orderId), {
      method: "POST", cache: "no-store", signal: timedSignal,
      headers: { Authorization: "Bearer " + token, "Content-Type": "application/json" },
      body: JSON.stringify({ purchase_order_id: purchaseId, cost_cny: costCNY,
        warehouse_address: warehouse, purchase_tracking_number: tracking, updated_at: updatedAt }),
    });
    diagnostic?.("response_headers", response);
  } catch (error) {
    diagnostic?.("failed", undefined, timedSignal.reason?.name === "TimeoutError" ? "timeout"
      : signal.aborted || lifetime.aborted ? "aborted" : "network");
    if (lifetime.aborted || signal.aborted) throw error;
    if (timedSignal.aborted) throw new Error("采购关联保存超时，结果尚未确认；请重新读取订单，不要盲目重复保存");
    throw new Error("采购关联写入结果未确认，请重新打开订单核对，不要盲目重复操作");
  }
  lifetime.throwIfAborted();
  timedSignal.throwIfAborted();
  const result = await response.json().catch(() => null) as {
    purchase?: OrderPurchase; error?: string;
  } | null;
  lifetime.throwIfAborted();
  timedSignal.throwIfAborted();
  if (!response.ok) {
    diagnostic?.("failed", response, result?.error);
    if (response.status === 409) throw new OrderRequestError("order_purchase_conflict", 409);
    if (response.status === 401) {
      await getAccessToken(true);
      throw new Error("会话已刷新，请检查采购关联是否保存后再操作");
    }
    if (response.status === 404) throw new Error("当前订单已不存在或子卖家映射已变化");
    if (response.status === 400) throw new Error("采购单号或成本金额格式不正确");
    throw new Error("保存采购信息失败（HTTP " + response.status + "）");
  }
  if (!result?.purchase?.purchase_order_id) {
    diagnostic?.("failed", response, "invalid_body");
    throw new Error("采购保存结果无法确认，请重新读取订单后核对");
  }
  diagnostic?.("response_body", response);
  return result.purchase;
}
