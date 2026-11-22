import { authLifetime, getAccessToken } from "./auth";
import { apiBase } from "./collections";
import { storeRequest } from "./stores";
import { beginRequestDiagnostic } from "./DevelopmentDiagnostics";

export type OrderScope = { storeId: string; site: string; logistic: string };
export type OrderLine = {
  item_id: string;
  title: string;
  variation_id?: string;
  image_url?: string;
  variation_attributes?: { name: string; value_name: string }[];
  quantity: number;
  unit_price: string;
  currency_id: string;
};
export type OrderShipment = {
  shipment_id: string;
  status: string;
  substatus?: string;
  tracking_number?: string;
  logistic_type?: string;
  last_updated?: string;
  checked_at: string;
};

export type OrderRecord = {
  order_id: string;
  pack_id?: string;
  status: string;
  currency_id: string;
  total_amount: string;
  date_created: string;
  last_updated: string;
  shipment_id?: string;
  shipment?: OrderShipment;
  order_items: OrderLine[];
  purchase_order_id?: string;
  purchase_cost_cny?: string | null;
  purchase_tracking_number?: string;
  // Populated only by the global cached search endpoint.
  store_id?: string;
  site_id?: string;
  logistic_type?: string;
  // Official shipment SLA; never infer deadline from order timestamps.
  dispatch_deadline?: string;
  dispatch_sla_status?: string;
  dispatch_checked_at?: string;
};
export type OrderPage = { items: OrderRecord[]; has_more: boolean; purchase_filter?: string };
export type OrderSearchPage = OrderPage & {
  query: string;
  site_filter: string;
  store_filter: string;
};
export function orderRecordScope(record: OrderRecord, fallback: OrderScope | null): OrderScope | null {
  return record.store_id && record.site_id && record.logistic_type
    ? { storeId: record.store_id, site: record.site_id, logistic: record.logistic_type } : fallback;
}
export function orderRecordKey(record: OrderRecord, fallback: OrderScope | null) {
  const scope = orderRecordScope(record, fallback);
  return [scope?.storeId ?? "", scope?.site ?? "", scope?.logistic ?? "", record.order_id].join("|");
}
export type OrderProgress = { saved: number; processed: number; has_more: boolean; next_offset: number };
export type OrderSyncState = { next_offset: number; can_reset: boolean; review?: OrderReview };
export type OrderReview = { order_id: string; review_id: string; offset: number; reason: string };

export class OrderRequestError extends Error {
  readonly code: string;
  readonly status: number;
  readonly review?: OrderReview;
  readonly progress?: OrderProgress;
  constructor(code: string, status: number, review?: OrderReview, progress?: OrderProgress) {
    super(orderErrorLabel(code, status));
    this.code = code;
    this.status = status;
    this.review = review;
    this.progress = progress;
  }
}

const messages: Record<string, string> = {
  order_purchase_conflict: "采购关联已被修改，请重新读取后再保存。",
  order_purchase_filter_invalid: "采购关联筛选无效，请选择全部、待关联或已关联。",
  order_image_unavailable: "官方未返回可信的商品主图；请先在已刊登商品中同步该商品。",
  order_unavailable: "订单服务暂不可用，请稍后重新读取。",
  order_not_found: "当前订单不存在，或此店铺无权访问。",
  order_sla_not_applicable: "已取消或 Fulfillment 订单不适用此发货 SLA。",
  order_sla_no_shipment: "当前订单没有可查询的官方 Shipment 或 SLA 已不存在。",
  order_sla_invalid: "官方发货截止时间格式异常，请查看诊断日志。",
  order_scope_required: "请先选择已验证的站点和物流类型。",
  order_reset_invalid: "重扫请求无效，请重新读取服务器同步状态。",
  order_needs_review: "有一笔订单详情异常，同步已暂停，请核实后再决定是否跳过。",
  order_review_conflict: "异常订单或同步位置已变化，请重新检查，不会重复跳过。",
  order_invalid_response: "美客多返回的订单数据无法解析，本次没有自动跳过订单。",
  order_platform_unavailable: "美客多订单接口暂不可用，请核对店铺权限或稍后重试。",
  order_shipment_missing: "平台未找到该运单或订单尚无 Shipment ID。",
  order_shipment_denied: "平台拒绝读取物流，请核对店铺授权及物流权限。",
  order_site_not_authorized: "当前站点或物流类型尚未通过店铺映射验证。",
  order_reauth_required: "店铺已断开或需要重新授权，历史缓存订单仍可在允许时查看。",
  order_refresh_unknown: "店铺授权续签结果待确认，请先到店铺管理检查。",
  order_busy: "订单服务正在处理其他请求，请稍后重试（非平台限流）。",
  order_auth_changed: "同步期间店铺授权或子卖家映射发生变化，未继续写入。",
  order_rate_limited: "美客多限制请求频率，请稍后继续同步。",
  order_timeout: "请求超时，已提交的订单和同步位置不会丢失。",
  order_canceled: "已停止本次同步，请刷新后继续。",
  order_reset_conflict: "当前不允许重扫：游标已变化、仍有待处理订单或存在有效租约。",
};

// Pure control decisions keep React state and network timing out of error policy tests.
export function shouldStopSLABatch(status: number, code: string) {
  // Avoid a 20-request storm during network failure, upstream outage or timeout.
  // A per-order missing Shipment (404) may be skipped without aborting the batch.
  return status === 0 || status === 401 || status === 403 || status === 429 ||
    status >= 500 ||
    ["order_rate_limited", "order_platform_unavailable", "order_reauth_required",
      "order_site_not_authorized", "order_shipment_denied", "order_refresh_unknown", "order_busy",
      "order_timeout", "order_unavailable"].includes(code);
}
export function shouldApplyOrderRefresh(requestId: string, currentId: string, aborted: boolean) {
  return !aborted && requestId === currentId;
}
export function orderErrorLabel(code: string, httpStatus?: number) {
  return messages[code] || (httpStatus ? "订单操作失败（HTTP " + httpStatus + "）" : "订单操作失败");
}
export function orderStatusLabel(status: string) {
  const labels: Record<string, string> = {
    paid: "已付款", payment_required: "待付款", confirmed: "已确认",
    cancelled: "已取消", cancellation_requested: "申请取消", partially_paid: "部分付款",
    partially_refunded: "部分退款", refunded: "已退款", payment_in_process: "付款处理中", invalid: "无效订单",
  };
  return labels[status] || status || "未知状态";
}
// Platform payment status is authoritative; procurement is a separate manually verified fact.
// Shipment ID and dispatch SLA cannot establish "shipped" or "delivered".
export function displayOrderStatus(order: OrderRecord) {
  return orderStatusLabel(order.status);
}

export function shipmentStatusLabel(shipment?: OrderShipment) {
  if (!shipment) return "尚未读取";
  const { status, substatus } = shipment;
  if (status === "shipped") {
    if (substatus === "in_transit") return "运输中";
    if (["out_for_delivery", "soon_deliver", "at_the_door"].includes(substatus ?? "")) return "派送中";
    if (["receiver_absent", "bad_address", "delivery_blocked", "impassable_zone"].includes(substatus ?? "")) return "物流异常";
    return "已发货";
  }
  const labels: Record<string, string> = {
    pending: "待处理", handling: "待发货", ready_to_ship: "待发货", delivered: "已签收",
    not_delivered: "物流异常", not_verified: "状态待核实", cancelled: "物流已取消",
  };
  return Object.hasOwn(labels, status) ? labels[status] : "未知状态";
}
export function hasProcurement(order: OrderRecord) {
  return Boolean(order.purchase_order_id || order.purchase_tracking_number);
}
export function orderVariationText(line?: OrderLine) {
  const attributes = line?.variation_attributes?.filter((attr) => attr.name && attr.value_name)
    .map((attr) => attr.name + ":" + attr.value_name);
  if (attributes?.length) return attributes.join(" · ");
  return line?.variation_id ? "变体 " + line.variation_id + "（属性未返回）" : "";
}

export function orderDate(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? "—" : date.toLocaleString("zh-CN", {
    year: "numeric", month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit",
  });
}
export function orderMoney(value: string, currency: string) {
  const symbols: Record<string, string> = {
    USD: "$", BRL: "R$", MXN: "MX$", ARS: "AR$", CLP: "CLP$", CNY: "¥",
  };
  const prefix = currency + (symbols[currency] ? " " + symbols[currency] : "") + " ";
  return prefix + value;
}
export const ORDER_READ_TIMEOUT_MS = 8_000;
export const ORDER_WRITE_TIMEOUT_MS = 25_000;

export function orderRead<T>(path: string, signal?: AbortSignal): Promise<T> {
  const deadline = AbortSignal.timeout(ORDER_READ_TIMEOUT_MS);
  const combined = signal ? AbortSignal.any([signal, deadline]) : deadline;
  return storeRequest<T>(path, "GET", combined, false, undefined, orderErrorLabel);
}
function pathWithTail(scope: OrderScope, tail: string) {
  return "/api/v1/stores/" + encodeURIComponent(scope.storeId) + "/orders/" + encodeURIComponent(tail) +
    "?" + new URLSearchParams({ site: scope.site, logistic_type: scope.logistic }).toString();
}
export function listOrders(scope: OrderScope, status: string, purchaseLink: "" | "linked" | "unlinked",
  offset: number, signal?: AbortSignal) {
  const query = new URLSearchParams({
    site: scope.site, logistic_type: scope.logistic, offset: String(offset),
  });
  if (status) query.set("status", status);
  if (purchaseLink) query.set("purchase", purchaseLink);
  return orderRead<OrderPage>("/api/v1/stores/" + encodeURIComponent(scope.storeId) +
    "/orders?" + query.toString(), signal).then((page) => {
    if (purchaseLink && (page.purchase_filter !== purchaseLink ||
      page.items.some((order) => Boolean(order.purchase_order_id) !== (purchaseLink === "linked")))) {
      throw new Error("服务端版本不支持采购筛选，或返回了不一致的关联结果。请先升级 Go 服务端再刷新订单。");
    }
    return page;
  });
}
// Global cached search never calls upstream Mercado Libre. Numeric IDs are always
// searched across all selected-store sites; title text respects the selected site.
export function searchOrders(storeId: string, site: string, logistic: string, query: string,
  status: string, purchaseLink: "" | "linked" | "unlinked", offset: number, signal?: AbortSignal) {
  const options = new URLSearchParams({
    site, logistic_type: logistic, q: query, status, purchase: purchaseLink, offset: String(offset),
  });
  return orderRead<OrderSearchPage>("/api/v1/stores/" + encodeURIComponent(storeId) +
    "/orders/search?" + options, signal).then((page) => {
    if (page.query !== query || page.site_filter !== site || page.store_filter !== storeId ||
        page.purchase_filter !== purchaseLink ||
        page.items.some((record) => !record.store_id || !record.site_id || !record.logistic_type ||
          (purchaseLink && Boolean(record.purchase_order_id) !== (purchaseLink === "linked")))) {
      throw new Error("服务端订单搜索结果与请求范围不一致，请先升级 Go API。");
    }
    return page;
  }).catch((error: unknown) => {
    if (error instanceof OrderRequestError && error.status === 404)
      throw new Error("Go 服务端尚未支持跨站点订单搜索，请先升级 Go API。");
    throw error;
  });
}
// A user-clicked, read-only official item query which only caches a trusted image URL.
export function refreshOrderImage(scope: OrderScope, id: string, signal?: AbortSignal) {
  if (!/^\d{1,30}$/.test(id)) throw new Error("订单号必须是数字");
  return storeRequest<OrderRecord>(pathWithTail(scope, id + "/image-refresh"), "POST",
    signal ? AbortSignal.any([signal, AbortSignal.timeout(25_000)]) : AbortSignal.timeout(25_000),
    false, undefined, orderErrorLabel);
}
export function getOrder(scope: OrderScope, id: string, signal?: AbortSignal) {
  if (!/^\d{1,30}$/.test(id)) throw new Error("订单号必须是数字");
  return orderRead<OrderRecord>(pathWithTail(scope, id), signal);
}
export function refreshOrderLogistics(scope: OrderScope, id: string, action: "sla" | "shipment", signal?: AbortSignal) {
  if (!/^\d{1,30}$/.test(id)) throw new Error("订单号必须是数字");
  return storeRequest<OrderRecord>(
    "/api/v1/stores/" + encodeURIComponent(scope.storeId) + "/orders/" + encodeURIComponent(id) +
      "/" + action + "?" + new URLSearchParams({ site: scope.site, logistic_type: scope.logistic }),
    "POST", signal ? AbortSignal.any([signal, AbortSignal.timeout(25_000)])
      : AbortSignal.timeout(25_000), false, undefined, orderErrorLabel);
}

export function refreshOneOrder(scope: OrderScope, id: string, signal?: AbortSignal) {
  if (!/^\d{1,30}$/.test(id)) throw new Error("订单号必须是数字");
  const path = "/api/v1/stores/" + encodeURIComponent(scope.storeId) + "/orders/" +
    encodeURIComponent(id) + "/refresh?" +
    new URLSearchParams({ site: scope.site, logistic_type: scope.logistic });
  return storeRequest<OrderRecord>(path, "POST",
    signal ? AbortSignal.any([signal, AbortSignal.timeout(25_000)]) : AbortSignal.timeout(25_000),
    false, undefined, orderErrorLabel);
}
export function getOrderSyncState(scope: OrderScope, signal?: AbortSignal) {
  return orderRead<OrderSyncState>(pathWithTail(scope, "sync-state"), signal);
}
async function orderPost(scope: OrderScope, tail: string, body: object | undefined, signal: AbortSignal) {
  const diagnostic = beginRequestDiagnostic(pathWithTail(scope, tail), "POST");
  const lifetime = authLifetime();
  const timedSignal = AbortSignal.any([signal, AbortSignal.timeout(ORDER_WRITE_TIMEOUT_MS)]);
  lifetime.throwIfAborted();
  timedSignal.throwIfAborted();
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
    response = await fetch(apiBase + pathWithTail(scope, tail), {
      method: "POST", cache: "no-store", signal: timedSignal,
      headers: {
        Authorization: "Bearer " + token,
        ...(body === undefined ? {} : { "Content-Type": "application/json" }),
      },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
    diagnostic?.("response_headers", response);
  } catch (error) {
    diagnostic?.("failed", undefined, timedSignal.reason?.name === "TimeoutError" ? "timeout"
      : signal.aborted || lifetime.aborted ? "aborted" : "network");
    if (signal.aborted || lifetime.aborted) throw error;
    if (timedSignal.aborted) throw new Error("订单操作超时，服务端是否提交尚未确认，请刷新后核对，不要直接重试。");
    throw new Error("订单请求未完成，请刷新后以服务端状态为准，勿盲目重复操作。");
  }
  lifetime.throwIfAborted();
  timedSignal.throwIfAborted();
  const value = await response.json().catch(() => null) as (OrderProgress & {
    error?: string; review?: OrderReview;
  }) | null;
  lifetime.throwIfAborted();
  timedSignal.throwIfAborted();
  if (!response.ok) {
    diagnostic?.("failed", response, value?.error);
    if (response.status === 401) {
      // Never automatically retry a POST whose commit outcome is unknown.
      await getAccessToken(true);
      throw new Error("会话已刷新，请确认服务器的订单状态后再操作。");
    }
    throw new OrderRequestError(value?.error || "order_unavailable", response.status,
      value?.review, value && typeof value.saved === "number" ? {
        saved: value.saved, processed: value.processed,
        has_more: value.has_more, next_offset: value.next_offset,
      } : undefined);
  }
  if (!value || typeof value.saved !== "number" || typeof value.has_more !== "boolean" ||
    typeof value.next_offset !== "number") {
    diagnostic?.("failed", response, "invalid_body");
    throw new Error("服务端已响应但结果无法确认，请刷新订单和同步位置。");
  }
  diagnostic?.("response_body", response);
  return value;
}
export function syncOrders(scope: OrderScope, signal: AbortSignal) {
  return orderPost(scope, "sync", undefined, signal);
}
export function skipOrder(scope: OrderScope, review: OrderReview, signal: AbortSignal) {
  return orderPost(scope, "skip", {
    order_id: review.order_id, review_id: review.review_id, offset: review.offset,
  }, signal);
}
export function resetOrderCursor(scope: OrderScope, offset: number, signal: AbortSignal) {
  return orderPost(scope, "reset", { offset }, signal);
}
