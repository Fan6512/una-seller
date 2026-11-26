import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { promisify } from "node:util";
import { createServer } from "vite";

const fixture = String.raw`
import React from "react";
import { flushSync } from "react-dom";
import { createRoot } from "react-dom/client";
import OrdersPage, { dispatchDeadlineLabel } from "/src/OrdersPage.tsx";
import { createFeedback } from "/src/operationFeedback.ts";
import { OperationFeedbackContext } from "/src/operationFeedbackContext.ts";
import { displayOrderStatus, hasProcurement, orderMoney, orderStatusLabel, orderVariationText,
  shouldApplyOrderRefresh, shouldStopLogisticsBatch, shipmentStatusLabel, orderDate, officialTrackingURL, orderUSD, orderFulfillmentStatus } from "/src/orders.ts";

const assert = (value, message) => { if (!value) throw new Error(message); };
const waitFor = async (check, message) => {
  for (let i = 0; i < 100; i++) {
    if (check()) return;
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
  throw new Error(message);
};
const markets = [
  { user_id: 1, site_id: "MLB", logistic_type: "remote" },
  { user_id: 2, site_id: "MLB", logistic_type: "fulfillment" },
  { user_id: 3, site_id: "MLA", logistic_type: "remote" },
  { user_id: 4, site_id: "MLM", logistic_type: "remote" },
];
const stores = ["one", "two"].map((id) => ({
  id, nickname: id, status: id === "one" ? "token_expired" : "connected",
  marketplaces_status: "verified", marketplaces: markets,
}));
const deferred = [];
const purchaseRequests = [];
const purchaseFilters = [];
const searchRequests = [];
const imageRequests = [];
let exerciseImages = false;
let distinctImageItems = false;
let pauseOrderRefresh = false;
const pausedRefreshes = [];
const statusReads = [];
const order = { order_id: "102", status: "paid", order_items: [
    { item_id: "MLB101", title: "Original title without translation", quantity: 1, unit_price: "12", currency_id: "USD" },
  ], total_amount: "67.23", currency_id: "USD", sales_amount_usd: "12.00",
  shipment_id: "87123456" };
let slaFetchCount = 0;
let billingFetches = 0;
let billingCacheReads = 0;
let savedBilling = null;
let billingRateLimited = false;
let shipmentFetchCount = 0;
let dueShipmentIDs = [];
let dueReads = 0;
let shipmentRateLimited = false;
let refreshPaid = false;
let shipmentDetailsCount = 0;
let cachedShipmentDetails;
const pendingShipmentDetails = [];
const shipmentDetails = {
  shipment_id: order.shipment_id, status: "delivered", tracking_number: "OFFICIAL-TRACK-102",
  checked_at: "2026-11-24T04:00:00Z", first_mile_carrier: "Fixture first mile",
  shipped_at: "2026-11-21T04:00:00Z", delivered_at: "2026-11-24T03:30:00Z",
  details: { checked_at: "2026-11-24T04:00:00Z", carrier_name: "Fixture official carrier", tracking_url: "https://www.jtexpress.com.br/track/102",
    history: [{ status: "shipped", date: "2026-11-21T04:00:00Z" },
      { status: "delivered", date: "2026-11-24T03:30:00Z" }] },
};
for (const [status, substatus, label] of [
  ["ready_to_ship", "printed", "待发货"], ["shipped", "", "已发货"], ["shipped", "in_transit", "运输中"],
  ["shipped", "out_for_delivery", "派送中"], ["delivered", "", "已签收"],
  ["cancelled", "", "物流已取消"], ["unknown_new", "", "未知状态"],
]) assert(shipmentStatusLabel({ status, substatus }) === label, "wrong official shipment label: " + status);
window.fetch = (path, options) => {
  const { pathname, searchParams } = new URL(path, location.href);
  if (pathname === "/api/v1/stores") return Promise.resolve(Response.json({ items: stores }));
  if (decodeURIComponent(pathname).endsWith("/image-refresh")) {
    if (exerciseImages) return new Promise(resolve => imageRequests.push({ resolve, path, signal: options.signal }));
    return Promise.resolve(Response.json({ error: "order_image_unavailable" }, { status: 404 }));
  }
  if (pathname.endsWith("/orders/200002/shipment")) {
    assert(searchParams.get("details") === "1", "selected order must request detail mode");
    return new Promise((resolve) => pendingShipmentDetails.push({ resolve, path, signal: options.signal }));
  }
  if (pathname.endsWith("/orders/shipment-refresh-candidates")) {
    dueReads++;
    return Promise.resolve(Response.json({order_ids:dueShipmentIDs,has_more:false}));
  }
  if (pathname.endsWith("/orders/102/shipment")) {
    shipmentFetchCount++;
    if(shipmentRateLimited)return Promise.resolve(Response.json({error:"order_rate_limited"},{status:429}));
    if (searchParams.get("details") === "1") {
      shipmentDetailsCount++;
      cachedShipmentDetails = shipmentDetails;
      order.shipment = { ...shipmentDetails, details: undefined };
      return Promise.resolve(Response.json({ ...order, shipment: shipmentDetails }));
    }
    order.shipment = { shipment_id: order.shipment_id, status: "shipped", substatus: "in_transit",
      tracking_number: "OFFICIAL-TRACK-102", checked_at: new Date().toISOString() };
    return Promise.resolve(Response.json(order));
  }
  if (pathname === "/api/v1/stores/order-title-translations") {
    const refs = JSON.parse(options.body);
    assert(refs.every(ref => Object.keys(ref).sort().join(",") === "logistic_type,order_id,site_id,store_id"),
      "translation must send scoped IDs only");
    return Promise.resolve(Response.json(refs.some(ref => ref.order_id === "200002") ? { "Lemon massage tool": "柠檬按摩工具" } : {}));
  }
  if (pathname.endsWith("/orders/102/sla")) {
    slaFetchCount++;
    order.dispatch_deadline = new Date(Date.now() + 3 * 86_400_000).toISOString();
    order.dispatch_sla_status = "on_time";
    order.dispatch_checked_at = new Date().toISOString();
    return Promise.resolve(Response.json(order));
  }
  if (pathname.endsWith("/billing-cache")) {
    assert(options.method === "GET", "billing cache must be read-only");
    const id = pathname.split("/").at(-2);
    if (id === "102" && searchParams.get("site") === "MLB" && searchParams.get("logistic_type") === "remote") {
      billingCacheReads++;
      if (savedBilling) return Promise.resolve(Response.json(savedBilling));
    }
    return Promise.resolve(Response.json({ net_income_usd: null, items: [{ order_id: id, available: false,
      reconciliation: "not_cached", payments: [], taxes: [], charges: [] }] }));
  }
  if (pathname.endsWith("/orders/billing-preview")) {
    billingFetches++;
    const q = new URL(path, location.href).searchParams;
    assert(q.get("site") === "MLB" && q.get("logistic_type") === "remote", "billing used the wrong scope");
    assert(JSON.parse(options.body).order_ids.join(",") === "102", "billing requested the wrong order");
    if (billingRateLimited) return Promise.resolve(Response.json({ error: "order_rate_limited" }, { status: 429, headers:{"Retry-After":"120"} }));
    savedBilling = { net_income_usd: null, items: [{ order_id: "102", available: true, checked_at: "2026-11-24T04:00:00Z",
      payments: [{ status: "approved", money_release_status: "pending" }], taxes: [],
      charges: [
        { detail_id: "91", type: "CHARGE", subtype: "CCSI", amount: "11.10", currency: "USD", shipment_id: "80001" },
        { detail_id: "94", type: "CHARGE", subtype: "CCSI", amount: "2.45", currency: "USD", shipment_id: "80002" },
        { detail_id: "92", type: "BONUS", subtype: "CCSI", amount: "-1.12", currency: "USD", reversal_of: "91" },
        { detail_id: "93", type: "CHARGE", subtype: "LOCAL_ONLY", amount: "1.25", currency: "BRL" },
        { detail_id: "95", type: "CHARGE", amount: "0.40", currency: "USD" },
      ],
      shipping: { buyer_cost_usd: "7.72", seller_charge_usd: "11.10", seller_net_cost_usd: "3.38",
        seller_charge_detail_id: "91" },
      charge_groups: [
        { type: "CHARGE", subtype: "CCSI", currency: "USD", original_total: "13.55" },
        { type: "CHARGE", subtype: "LOCAL_ONLY", currency: "BRL", original_total: "1.25" },
        { type: "BONUS", subtype: "CCSI", currency: "USD", original_total: "-1.12" },
        { type: "CHARGE", currency: "USD", original_total: "0.40" },
      ],
    }] };
    return Promise.resolve(Response.json(savedBilling));
  }
  if (pathname.endsWith("/orders/102/refresh")) {
    if (pauseOrderRefresh) return new Promise(resolve => pausedRefreshes.push(resolve));
    order.status = refreshPaid ? "paid" : "partially_refunded";
    return Promise.resolve(Response.json(order));
  }
  if (pathname.endsWith("/orders/search")) {
    const uri = new URL(path, location.href);
    const q = uri.searchParams.get("q") ?? "";
    const storeFilter = pathname.split("/")[4];
    const site = uri.searchParams.get("site") ?? "";
    const purchaseFilter = uri.searchParams.get("purchase") ?? "";
    searchRequests.push({ q, storeFilter, site, uri });
    const lemon = { ...order, order_id: "200002", store_id: "one", site_id: "MLM", logistic_type: "remote",
      purchase_order_id: "", order_items: [{ item_id: "MLM100", title: "Lemon massage tool",
        quantity: 1, unit_price: "12", currency_id: "USD" }] };
    const another = { ...lemon, store_id: "two", site_id: "MLB",
      order_items: [{ ...lemon.order_items[0], item_id: "MLB200" }] };
    const rows = q === "200002" || q.toLowerCase() === "lemon"
      ? storeFilter === "all" ? [lemon, another] : [lemon]
      : site === "*" ? [{ ...lemon, order_id: "10001", site_id: "MLB" }, lemon] : [lemon];
    return Promise.resolve(Response.json({ items: rows, has_more: false, purchase_filter: purchaseFilter,
      query: q, site_filter: site, store_filter: storeFilter }));
  }
  if (pathname.endsWith("/orders")) {
    const filter = new URL(path, location.href).searchParams.get("purchase") ?? "";
    purchaseFilters.push(filter);
    const status = searchParams.get("status") ?? "";
    statusReads.push(status);
    const linked = { ...order, order_id: "103", shipment_id: "",
      purchase_order_id: "12345678", purchase_cost_cny: "16",
      ...(distinctImageItems ? { order_items: [{ ...order.order_items[0], item_id: "MLB702" }] } : {}) };
    const rows = (filter === "linked" ? [linked] : filter === "unlinked" ? [order] : [order, linked])
      .filter(row => !status || (status.startsWith("fulfillment_") ?
        "fulfillment_" + orderFulfillmentStatus(row) === status : row.status === status));
    return Promise.resolve(Response.json({ items: rows, has_more: false, purchase_filter: filter }));
  }
   if (pathname.endsWith("/sync-state")) return Promise.resolve(Response.json({ next_offset: 0, can_reset: false }));
  if (pathname.endsWith("/orders/102")) return Promise.resolve(Response.json({ ...order, shipment: cachedShipmentDetails ?? order.shipment }));
  if (pathname.endsWith("/orders/200002")) return Promise.resolve(Response.json({
    ...order, order_id: "200002", order_items: [
      { item_id: "MLM100", title: "Lemon massage tool", quantity: 1, unit_price: "12", currency_id: "USD" },
    ],
  }));
  if (pathname.endsWith("/orders/103")) return Promise.resolve(Response.json({
    ...order, order_id: "103", purchase_order_id: "12345678", purchase_cost_cny: "16",
  }));
  if (pathname.endsWith("/purchase")) {
    return new Promise((resolve) => purchaseRequests.push({ resolve, path, ...options }));
  }
  if (pathname.endsWith("/sync")) {
    // Simulate a response already on its way when the scope is switched.
    return new Promise((resolve) => deferred.push({ resolve, signal: options.signal }));
  }
  throw new Error("unexpected request " + path);
};
const select = (name, value) => {
  const field = document.querySelector('select[aria-label="' + name + '"]');
  field.value = value;
  flushSync(() => field.dispatchEvent(new Event("change", { bubbles: true })));
};
const button = () => [...document.querySelectorAll("button")].find((item) => item.textContent === "同步订单");
const refreshStoreStatus = () => [...document.querySelectorAll("button")]
  .find((item) => item.textContent === "刷新店铺状态");
const purchaseButton = (text) => [...document.querySelectorAll(".orders-purchase-actions button")]
  .find((item) => item.textContent === text);
const takePurchase = async (method) => {
  await waitFor(() => purchaseRequests.length > 0, "purchase request missing");
  const request = purchaseRequests.shift();
  assert(request.method === method, "unexpected purchase method");
  return request;
};
const openPurchase = async () => {
  await waitFor(() => document.querySelector(".orders-id"), "order unavailable");
  flushSync(() => document.querySelector(".orders-row-purchase").click());
  await waitFor(() => document.querySelector("dialog.orders-purchase-modal")?.open,
    "purchase action did not open its own compact dialog");
  assert(!document.querySelector("dialog.orders-modal:not(.orders-purchase-modal)")?.open,
    "purchase and order-details must not share a dialog");
  return takePurchase("GET");
};
const review = { order_id: "102", review_id: "fixture", offset: 3, reason: "invalid_detail" };
const progress = { saved: 7, processed: 1, next_offset: 3, has_more: true };
const reviewResponse = () => Response.json({ ...progress, review, error: "order_needs_review" }, { status: 409 });
const start = async () => {
  await waitFor(() => button() && !button().disabled, "sync button unavailable");
  flushSync(() => button().click());
  await waitFor(() => deferred.length > 0, "sync did not start");
  return deferred.shift();
};
let active = 0;
try {
  const feedback = createFeedback(() => {});
  createRoot(document.querySelector("main")).render(React.createElement(OperationFeedbackContext.Provider, {
    value: feedback,
  }, React.createElement(OrdersPage, {
    installingUpdate: false, beginSync() { active++; return () => active--; },
  })));
  const filter = (label) => [...document.querySelectorAll(".orders-procurement-toolbar button")]
    .find((item) => item.textContent === label);
  await waitFor(() => filter("待关联"), "purchase filter unavailable");
  const actions = document.querySelector(".orders-row-actions");
  await waitFor(() => document.querySelector(".orders-row-actions"), "direct row actions missing");
  assert([...document.querySelector(".orders-row-actions").querySelectorAll("button")]
    .map((b) => b.textContent).join("|") === "订单详情|采购",
    "purchase must be directly below order details, not inside details modal");
  assert([...document.querySelectorAll(".orders-heading-actions button")]
    .some((btn) => btn.textContent === "补全本页主图"), "explicit image backfill missing");
  flushSync(() => filter("待关联").click());
  await waitFor(() => purchaseFilters.includes("unlinked") &&
    document.querySelector(".orders-row-purchase")?.textContent === "采购",
    "unlinked row must expose direct purchase CTA");
  flushSync(() => filter("已关联").click());
  await waitFor(() => purchaseFilters.includes("linked") &&
    document.querySelector(".orders-row-purchase")?.textContent === "编辑采购",
    "linked row must expose edit purchase CTA");
  assert(document.querySelector(".orders-row-cost")?.textContent.includes("¥16"),
    "linked procurement cost must remain visible on the list");
  assert(document.querySelector(".orders-status")?.textContent === "已付款" &&
    document.querySelector(".orders-purchase-badge")?.textContent === "已采购",
    "official paid status must remain distinct from procurement");
  flushSync(() => filter("全部").click());
  await waitFor(() => purchaseFilters.at(-1) === "", "clear purchase filter not sent");
  for (const [field, value, response] of [
    ["订单店铺", "two", reviewResponse()],
    ["站点和物流模式", "MLB|fulfillment", Response.json(progress)],
    ["站点和物流模式", "MLA|remote", reviewResponse()],
  ]) {
    const stale = await start();
    select(field, value);
    stale.resolve(response);
    await waitFor(() => active === 0 || deferred.length > 0, "previous sync did not settle");
    assert(stale.signal.aborted && deferred.length === 0, "previous sync continued after scope change");
    assert(!document.querySelector(".orders-notice, .orders-review"), "previous result appeared in current scope");
  }
  (await start()).resolve(Response.json({ ...progress, has_more: false }));
  await waitFor(() => document.querySelector(".orders-notice")?.textContent.includes("订单同步完成"), "current success missing");
  (await start()).resolve(reviewResponse());
  await waitFor(() => document.querySelector(".orders-review")?.textContent.includes("102"), "current Review missing");

  const staleRead = await openPurchase();
  select("订单店铺", "one");
  const currentRead = await openPurchase();
  const purchase = { order_id: "102", purchase_order_id: "12345678", cost_cny: "16",
    source: "manual", warehouse_address: "", purchase_tracking_number: "",
    updated_at: "2026-11-20T00:00:00.123456Z" };
  currentRead.resolve(Response.json({ purchase }));
  staleRead.resolve(Response.json({ purchase: { ...purchase, purchase_order_id: "87654321" } }));
  const save = () => purchaseButton("保存采购关联");
  await waitFor(() => save() && !save().disabled, "purchase form unavailable after read");
  assert(staleRead.signal.aborted, "previous purchase read was not canceled");
  flushSync(() => save().click());
  const conflict = await takePurchase("POST");
  const submitted = JSON.parse(conflict.body);
  assert(submitted.purchase_order_id === purchase.purchase_order_id, "previous purchase read entered current form");
  assert(submitted.warehouse_address === "" && submitted.purchase_tracking_number === "",
    "compact purchase form must submit explicit forwarding fields");
  assert(submitted.updated_at === purchase.updated_at, "save did not send the full version");
  conflict.resolve(Response.json({ error: "order_purchase_conflict" }, { status: 409 }));
  const reread = () => purchaseButton("重新读取采购关联");
  await waitFor(reread, "purchase conflict did not offer to reread");
  assert(save().disabled, "stale version can still be submitted");
  flushSync(() => reread().click());
  const latest = { ...purchase, updated_at: "2026-11-20T00:00:01.123456Z" };
  (await takePurchase("GET")).resolve(Response.json({ purchase: latest }));
  await waitFor(() => save() && !save().disabled, "reread did not enable save");
  flushSync(() => save().click());
  const staleSave = await takePurchase("POST");
  assert(JSON.parse(staleSave.body).updated_at === latest.updated_at, "reread version was not used");
  select("站点和物流模式", "MLB|fulfillment");
  (await openPurchase()).resolve(Response.json({ purchase: null }));
  staleSave.resolve(Response.json({ purchase }));
  await waitFor(() => document.querySelector('input[aria-label="1688 采购订单号"]:enabled'), "new purchase form unavailable");
  assert(staleSave.signal.aborted && !document.querySelector(".orders-purchase-success, .orders-purchase-state"),
    "previous purchase save entered current form");

  // A verified expired-token store must still be able to start one explicit
  // sync. Show the upstream failure code without leaking any request token.
  const failed = await start();
  failed.resolve(Response.json({ error: "order_platform_unavailable" }, { status: 502 }));
  await waitFor(() => document.querySelector(".orders-notice")?.textContent.includes("HTTP 502"),
    "order failure HTTP status not surfaced");
  assert(document.querySelector(".orders-notice").textContent.includes("order_platform_unavailable"),
    "upstream failure error code not surfaced");

  // Store state can change after a connection check while OrdersPage is open;
  // use a fresh status read, rather than a stale snapshot, to gate actions.
  assert(refreshStoreStatus(), "refresh connection control missing");
  stores[0].status = "reauth_required";
  flushSync(() => refreshStoreStatus().click());
  await waitFor(() => button()?.disabled && document.querySelector(".orders-filter-note")?.textContent.includes("重新授权"),
    "reauthorization status did not disable sync");
  stores[0].status = "token_expired";
  flushSync(() => refreshStoreStatus().click());
  await waitFor(() => button() && !button().disabled &&
    document.querySelector(".orders-filter-note")?.textContent.includes("访问令牌已到期"),
    "expired-token store remained disabled after refreshing status");
  const afterRefresh = await start();
  stores[0].status = "connected";
  afterRefresh.resolve(Response.json({ ...progress, has_more: false }));
  await waitFor(() => document.querySelector(".orders-notice")?.textContent.includes("订单同步完成"),
    "sync success not confirmed");
  await waitFor(() => [...document.querySelectorAll(".orders-filter-note")]
    .some((item) => item.textContent.includes("当前店铺") && item.textContent.includes("已连接")),
    "successful token renewal left the expired banner visible");
  const sheet = document.querySelector("dialog.orders-purchase-modal");
  assert(sheet?.open, "purchase dialog closed unexpectedly");
  flushSync(() => document.querySelector('button[aria-label="关闭采购关联"]').click());
  await waitFor(() => !sheet.open, "purchase dialog did not close");

  // Preserve both existing purchase filters and the new SLA/refund behavior.
  assert(orderStatusLabel("partially_refunded") === "部分退款", "partial refund label missing");
  assert(displayOrderStatus({ ...order, purchase_order_id: "12345678" }) === "已付款",
    "purchase link must not replace the official paid label");
  assert(displayOrderStatus({ ...order, shipment: shipmentDetails }) === "已签收", "delivered order must display fulfillment state");
  assert(displayOrderStatus({ ...order, status: "refunded", shipment: shipmentDetails }) === "已退款", "refund must win over delivery");
  assert(displayOrderStatus({ ...order, shipment: { ...shipmentDetails, shipment_id: "stale" } }) === "已付款", "stale shipment must not change order state");
  assert(displayOrderStatus({ ...order, shipment: { ...shipmentDetails, status: "ready_to_ship", substatus: "printed" } }) === "标签已打印", "printed label must not imply dispatch");
  assert(orderDate("2026-11-09T16:14:00-03:00").includes("2026/11/10") && orderDate("2026-11-09T16:14:00-03:00").includes("03:14"), "Beijing timezone must be independent of machine timezone");
  assert(hasProcurement({ ...order, purchase_order_id: "12345678" }), "purchase badge predicate incorrect");
  assert(!hasProcurement(order), "unlinked order must not be marked procured");
  assert(orderVariationText({ item_id: "MLM111", title: "", variation_id: "998" }) ===
    "变体 998（属性未返回）", "variation ID fallback missing");
  assert(orderVariationText({ item_id: "MLM111", title: "",
    variation_attributes: [{ name: "Color", value_name: "Amarillo" }] }) ===
    "Color:Amarillo", "official variation attributes missing");
  assert(orderVariationText({ item_id: "MLM111", title: "" }) === "", "absent variation must not render noise");
  assert(orderMoney("12", "BRL") === "BRL R$ 12", "BRL symbol regresssed");
  assert(orderMoney("23.78", "USD") === "USD $ 23.78", "USD currency symbol regresssed");
  assert(shouldApplyOrderRefresh("102", "102", false), "same order refresh must apply");
  assert(!shouldApplyOrderRefresh("102", "103", false), "stale A must not replace selected B");
  assert(!shouldApplyOrderRefresh("102", "102", true), "aborted request must not apply");
  assert(shouldStopLogisticsBatch(502, "order_platform_unavailable"), "upstream 502 must stop");
  assert(shouldStopLogisticsBatch(504, "order_timeout"), "timeout must stop");
  assert(shouldStopLogisticsBatch(429, "order_rate_limited"), "429 must stop");
  assert(shouldStopLogisticsBatch(0, ""), "network loss must stop");
  assert(!shouldStopLogisticsBatch(404, "order_sla_no_shipment"), "one no-shipment row may be skipped");
  const now = Date.UTC(2026, 10, 21, 0, 0);
  const sample = (hours, status = "on_time") => ({
    ...order, status: "paid", dispatch_sla_status: status,
    dispatch_deadline: new Date(now + hours * 3600_000).toISOString(),
  });
  assert(dispatchDeadlineLabel(sample(42), now, "remote").countdown === "剩余1天18小时", "show remaining days and hours");
  assert(dispatchDeadlineLabel({ ...sample(-2), shipment: shipmentDetails }, now, "remote").countdown === "已签收", "delivered orders must not show expired countdown");
  assert(dispatchDeadlineLabel(sample(24), now, "remote").state === "neutral",
    "24h must not be red");
  assert(dispatchDeadlineLabel(sample(23.9), now, "remote").state === "late",
    "under 24h must be red");
  assert(dispatchDeadlineLabel(sample(0.5), now, "remote").countdown === "不足1小时",
    "under 1h boundary lost");
  assert(dispatchDeadlineLabel({ ...sample(72, "delayed"), dispatch_deadline: undefined },
    now, "remote").countdown === "平台标记逾期", "official delay without date lost");
  assert(dispatchDeadlineLabel(sample(72), now, "fulfillment").countdown === "不适用",
    "fulfillment incorrectly displays SLA");
  assert(slaFetchCount === 0, "SLA must not run automatically");
  assert(document.querySelector(".orders-product-title")?.textContent === "Original title without translation",
    "an untranslated title did not fall back to the original");
  assert([...document.querySelectorAll("th")].filter((h) => h.textContent.trim()).length === 8,
    "table columns must be 8 after moving purchase action");
  assert([...document.querySelectorAll("td.orders-table-empty")].every((cell) => cell.colSpan === 8),
    "empty-state colSpan should be 8");
  const slaButton = () => [...document.querySelectorAll(".orders-tabs button")]
    .find((item) => item.textContent.includes("更新本页发货截止"));
  assert(slaButton()?.disabled, "fulfillment SLA control must be disabled");
  select("站点和物流模式", "MLB|remote");
  await waitFor(() => document.querySelector(".orders-id")?.textContent === "102" &&
    Boolean(slaButton()) && !slaButton().disabled, "remote SLA control not ready");
  flushSync(() => slaButton().click());
  await waitFor(() => slaFetchCount === 1 &&
    document.querySelector(".orders-dispatch")?.textContent.includes("剩余"),
    "manual SLA refresh did not update list");
  assert(shipmentFetchCount === 0, "opening/listing must not trigger logistics refresh");
  // Even a zero-new-order sync must refresh the server-selected stale shipment.
  dueShipmentIDs=["102"]; order.shipment={shipment_id:order.shipment_id,status:"ready_to_ship",substatus:"printed",checked_at:"2026-11-24T00:00:00Z"};
  const beforeDue=dueReads;
  (await start()).resolve(Response.json({...progress,saved:0,has_more:false}));
  await waitFor(()=>dueReads===beforeDue+1&&shipmentFetchCount===1&&document.querySelector(".orders-status")?.textContent.includes("运输中"),"zero-new-order sync failed to refresh stale logistics");
  assert(shipmentDetailsCount===0,"sync must not read full history");
  dueShipmentIDs=["102","999"]; shipmentRateLimited=true;
  (await start()).resolve(Response.json({...progress,saved:0,has_more:false}));
  await waitFor(()=>document.querySelector(".orders-notice")?.textContent.includes("已停止")&&shipmentFetchCount===2,"rate limit did not stop sync logistics pass");
  assert(document.querySelector(".orders-notice").textContent.includes("未更新 2 笔"),"partial logistics result not reported");
  dueShipmentIDs=[]; shipmentRateLimited=false;
  shipmentFetchCount=0;
  const batchShipment = [...document.querySelectorAll(".orders-tabs button")]
    .find((item) => item.textContent === "更新本页物流");
  flushSync(() => batchShipment.click());
  await waitFor(() => shipmentFetchCount === 1 &&
    document.querySelector(".orders-shipping-refs")?.textContent.includes("OFFICIAL-TRACK-102") &&
    !document.querySelector(".orders-shipping-refs")?.textContent.includes("运输中") &&
    document.querySelector(".orders-status")?.textContent.includes("运输中"),
    "main order state must follow current shipment");
  const stateFilter = label => [...document.querySelectorAll(".orders-tabs button")].find(item => item.textContent === label);
  flushSync(() => stateFilter("运输中").click());
  await waitFor(() => statusReads.at(-1) === "fulfillment_in_transit" && document.querySelectorAll(".orders-id").length === 1,
    "server transport filter did not select the expected order");
  let transportReads = statusReads.length;
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector("dialog.orders-modal:not(.orders-purchase-modal)")?.open,
    "real order detail modal did not open after purchase modal closed");
  const shipmentButton = () => [...document.querySelectorAll(".orders-detail-block button")]
    .find((item) => item.textContent === "更新平台物流");
  await waitFor(() => shipmentButton() && !shipmentButton().disabled, "order shipment button unavailable");
  assert(shipmentDetailsCount === 0 && shipmentFetchCount === 1,
    "batch/automatic order opening must not load tracking details");
  refreshPaid=true;
  flushSync(()=>document.querySelector(".orders-detail-refresh").click());
  await waitFor(()=>shipmentFetchCount===2&&!document.querySelector(".orders-detail-refresh").disabled,"single order status refresh did not read shipment summary");
  assert(shipmentDetailsCount===0,"single combined refresh must not read history");
  refreshPaid=false;
  transportReads=statusReads.length;
  flushSync(() => shipmentButton().click());
  await waitFor(() => shipmentDetailsCount === 1 &&
    document.querySelector(".orders-detail-fields time")?.dateTime === shipmentDetails.shipped_at &&
    document.querySelector(".orders-detail-fields")?.textContent.includes("Fixture official carrier"),
    "single-order shipment refresh did not show carrier and dispatch time");
  await waitFor(() => statusReads.length > transportReads && statusReads.at(-1) === "fulfillment_in_transit" &&
    document.querySelectorAll(".orders-id").length === 0 && document.querySelector("dialog.orders-modal:not(.orders-purchase-modal)")?.open,
    "single-order delivery must requery the filtered list without closing details");
  const detailsFields = document.querySelector(".orders-detail-fields");
  assert(detailsFields.textContent.includes("OFFICIAL-TRACK-102") &&
    detailsFields.textContent.includes("物流最近核验") &&
    detailsFields.textContent.includes("首程承运商Fixture first mile") &&
    [...detailsFields.querySelectorAll("time")].some(time => time.dateTime === shipmentDetails.delivered_at),
    "tracking, first-mile carrier and exact receipt time must stay distinct");
  const trackingActions = [...detailsFields.querySelectorAll("button")];
  const browserTracking = trackingActions.find(button => button.textContent === "官方地址浏览器打开（备用）");
  const opened = [];
  window.open = (...args) => opened.push(args);
  browserTracking?.click();
  assert(trackingActions.some(button => button.textContent === "官方承运商原始查询（备用）") &&
    trackingActions.some(button => button.textContent === "复制地址") &&
    opened[0]?.[0] === shipmentDetails.details.tracking_url && opened[0]?.[2] === "noopener,noreferrer",
    "carrier actions must offer popup, clipboard and safe browser fallback");
  const providerSelect = detailsFields.querySelector('select[aria-label="物流查询平台"]');
  assert(providerSelect?.value === "track718","Track718 must be the default query provider");
  const providerBrowser = trackingActions.find(button=>button.textContent==="浏览器查询");
  providerBrowser.click();
  assert(opened.at(-1)?.[0]==="https://www.track718.com/zh-CN","Track718 destination changed");
  flushSync(()=>{providerSelect.value="17track";providerSelect.dispatchEvent(new Event("change",{bubbles:true}));});
  providerBrowser.click();
  assert(opened.at(-1)?.[0]==="https://www.17track.net/zh-cn","17TRACK destination changed");
  const beforePopup = opened.length;
  flushSync(()=>trackingActions.find(button=>button.textContent==="独立窗口查询").click());
  assert(opened.length===beforePopup&&detailsFields.textContent.includes("无法打开客户端独立窗口"),"popup must not silently open a browser tab");
  assert(officialTrackingURL("file:///tmp/example") === "" &&
    officialTrackingURL("https://10.0.0.1/track") === "" &&
    officialTrackingURL("https://u@jtexpress.com.br/track") === "", "unsafe tracking address accepted");
  assert(orderUSD("11.28") === "USD 11.28" && orderUSD() === "待核实" &&
    orderUSD("22.59000003") === "USD 22.59" && orderUSD("9.995") === "USD 10.00", "USD display must round exactly and fail closed");
  const timeline = [...document.querySelectorAll(".orders-detail-block h3")]
    .find(item => item.textContent === "官方物流时间轴");
  assert(timeline?.parentElement.querySelectorAll("time").length === 2 &&
    !document.querySelector(".orders-shipping-refs")?.textContent.includes("Fixture official carrier"),
    "timeline must be rendered inside the selected detail");
  const beforeReopen = shipmentFetchCount;
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  await waitFor(() => !document.querySelector("dialog.orders-modal:not(.orders-purchase-modal)")?.open,
    "detail did not close");
  flushSync(() => stateFilter("全部订单").click());
  await waitFor(() => statusReads.at(-1) === "" && document.querySelector(".orders-id")?.textContent === "102",
    "all orders were not restored after leaving the transport filter");
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector(".orders-detail-fields")?.textContent.includes("Fixture official carrier") &&
    [...document.querySelectorAll(".orders-detail-block h3")].some(item => item.textContent === "官方物流时间轴"),
    "reopening detail did not display cached carrier and history");
  assert(shipmentFetchCount === beforeReopen && shipmentDetailsCount === 1,
    "reopening cached detail must not request official shipment");
  flushSync(() => [...document.querySelectorAll(".orders-detail-refresh")][0].click());
  await waitFor(() => document.querySelector(".orders-detail-summary")?.textContent.includes("部分退款"),
    "manual order refresh did not display refunded state");
  assert(!document.querySelector(".orders-money-net") &&
    ![...document.querySelectorAll("button")].some((item) => item.textContent === "订单设置"),
    "nonexistent finance settings should not be visible");
  const profit = () => document.querySelector('section[aria-label="订单账单与利润"]');
  assert(billingFetches === 0 && profit()?.textContent.includes("待核账"), "billing ran automatically or invented net proceeds");
  await waitFor(() => billingCacheReads > 0, "opening order must read local billing cache");
  flushSync(() => profit().querySelector("button").click());
  await waitFor(() => billingFetches === 1 && profit()?.textContent.includes("-USD 3.38") &&
    profit()?.textContent.includes("+USD 7.72") && profit()?.textContent.includes("-USD 11.10") &&
    profit()?.textContent.includes("待释放，待核账"), "billing must show scoped shipping amounts and keep net proceeds pending");
  assert(profit().textContent.includes("USD 12.00") && profit().textContent.includes("-USD 1.12") &&
    profit().textContent.match(/11\.10/g)?.length === 1 &&
    !["67.23", "1.25", "7.90"].some(amount => profit().textContent.includes(amount)),
    "wrong sale, currency, duplicate shipping charge or unverified net was displayed");
  assert(profit().textContent.includes("同类物流扣项逐项展示") && profit().querySelector("details")?.open &&
    profit().textContent.includes("USD 2.45") && !profit().textContent.includes("USD 13.55"),
    "mixed shipping group must expose remaining charges without showing a misleading subtotal");
  assert(profit().textContent.includes("不可相加或重复扣减") && profit().textContent.includes("同货件多单时未分摊") &&
    profit().textContent.includes("已批准 / 待释放"), "shipping scope and payment status must be explicit");
  const subtotal = [...profit().querySelectorAll("dl")].find(dl => dl.textContent.includes("CHARGE 小计"));
  assert(subtotal && getComputedStyle(subtotal).display === "grid" && subtotal.children.length === 4,
    "subtotal groups must be direct grid pairs");
  const pairs = [...subtotal.children];
  for (let i = 0; i < pairs.length; i += 2) {
    const label = pairs[i].getBoundingClientRect(), value = pairs[i+1].getBoundingClientRect();
    assert(pairs[i].tagName === "DT" && pairs[i+1].tagName === "DD" && Math.abs(label.top-value.top)<1 &&
      value.left >= label.right, "subtotal grid label and value must share a row");
  }
  assert(subtotal.textContent.includes("USD 0.40") && subtotal.textContent.includes("-USD 1.12") &&
    profit().textContent.includes("部分官方费用缺少已核实的美元金额，暂不展示") &&
    !["BRL", "LOCAL_ONLY", "1.25", "USD 待核实"].some(value => profit().textContent.includes(value)),
    "non-USD subtotals and line items must stay hidden while confirmed USD rows remain visible");
  const beforeBillingReopen = billingCacheReads;
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => billingCacheReads > beforeBillingReopen && profit()?.textContent.includes("-USD 3.38") &&
    profit()?.textContent.includes("账单最近成功保存") && profit()?.textContent.includes("2026/11/24 12:00"),
    "reopening detail must read the persisted server snapshot and timestamp");
  assert(billingFetches === 1, "reopening detail made another official Billing request");
  billingRateLimited = true;
  flushSync(() => profit().querySelector("button").click());
  await waitFor(() => billingFetches === 2 && profit().querySelector("button").disabled &&
    profit().querySelector("button").textContent === "限流冷却中" && profit().textContent.includes("平台暂时限流"),
    "rate limit must stop manual billing retries");
  assert(profit().textContent.includes("-USD 3.38"), "rate limited read must retain prior scoped bill");
  assert(profit().textContent.includes("120 秒")||profit().textContent.includes("119 秒"),"server Retry-After ignored");
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(()=>profit()?.querySelector("button").textContent==="限流冷却中","reopening bypassed shared billing cooldown");
  assert(billingFetches===2&&profit().querySelector("button").disabled,"reopening triggered an upstream call");
  billingRateLimited = false;

  // Issue #53: numeric Order ID crosses sites; title searches respect selected site.
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  await waitFor(() => !document.querySelector("dialog.orders-modal:not(.orders-purchase-modal)")?.open,
    "detail dialog remained open before global search");
  select("站点和物流模式", "MLB|remote");
  const searchInput = () => document.querySelector('input[type="search"]');
  const submitSearch = () => [...document.querySelectorAll(".orders-search-inner button")]
    .find((item) => item.textContent === "搜索");
  flushSync(() => {
    const input = searchInput();
    const nativeSetter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value").set;
    nativeSetter.call(input, "200002");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  flushSync(() => submitSearch().click());
  await waitFor(() => searchRequests.some((q) => q.q === "200002" && q.site === "*") &&
    document.querySelector('select[aria-label="站点和物流模式"]')?.value === "all" &&
    document.querySelector(".orders-id")?.textContent === "200002",
    "cross-site Order ID must change to all sites and show exactly the matching order");
  assert(document.querySelectorAll(".orders-id").length === 1, "numeric search returned extra orders");
  await waitFor(() => document.querySelector(".orders-product-title")?.textContent === "柠檬按摩工具",
    "Chinese translation was not the visible title");
  assert(document.querySelector(".orders-product-title")?.title === "Lemon massage tool",
    "official title is missing from the tooltip");
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector(".orders-line-content strong")?.textContent === "柠檬按摩工具" &&
    document.querySelector(".orders-line-content")?.textContent.includes("原标题：Lemon massage tool"),
    "order details must show the translation and the complete original title");
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  flushSync(() => document.querySelector(".orders-row-purchase").click());
  const crossSitePurchase = await takePurchase("GET");
  assert(new URL(crossSitePurchase.path, location.href).searchParams.get("site") === "MLM" &&
    crossSitePurchase.path.includes("/stores/one/orders/200002/purchase"),
    "cross-site purchase opened under the wrong site or store");
  assert(document.querySelector('[data-testid="purchase-target-scope"]')?.textContent.includes("店铺：one") &&
    document.querySelector('[data-testid="purchase-target-scope"]')?.textContent.includes("墨西哥"),
    "purchase modal must identify the first-store MLM target");
  crossSitePurchase.resolve(Response.json({ purchase: null }));
  await waitFor(() => document.querySelector('button[aria-label="关闭采购关联"]:enabled'),
    "cross-site purchase dialog not ready");
  flushSync(() => document.querySelector('button[aria-label="关闭采购关联"]').click());

  select("站点和物流模式", "MLM|remote");
  flushSync(() => {
    const input = searchInput();
    const nativeSetter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value").set;
    nativeSetter.call(input, "Lemon");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  flushSync(() => submitSearch().click());
  await waitFor(() => searchRequests.some((q) => q.q === "Lemon" && q.site === "MLM") &&
    document.querySelector('select[aria-label="站点和物流模式"]')?.value === "MLM|remote",
    "title search must remain in MLM, not switch sites");

  select("订单店铺", "all");
  await waitFor(() => searchRequests.some((q) => q.storeFilter === "all") &&
    document.querySelectorAll(".orders-id").length === 2, "all-store search must merge both stores");
  flushSync(() => [...document.querySelectorAll(".orders-row-purchase")][1].click());
  const secondStorePurchase = await takePurchase("GET");
  assert(secondStorePurchase.path.includes("/stores/two/orders/200002/purchase") &&
    new URL(secondStorePurchase.path, location.href).searchParams.get("site") === "MLB",
    "second-store row must not reuse first-store purchase scope");
  assert(document.querySelector('[data-testid="purchase-target-scope"]')?.textContent.includes("店铺：two") &&
    document.querySelector('[data-testid="purchase-target-scope"]')?.textContent.includes("巴西"),
    "purchase modal must identify second-store MLB target");
  secondStorePurchase.resolve(Response.json({ purchase: null }));
  await waitFor(() => document.querySelector('button[aria-label="关闭采购关联"]:enabled'),
    "second store purchase did not settle");
  flushSync(() => document.querySelector('button[aria-label="关闭采购关联"]').click());
  const openShipping = async (index) => {
    flushSync(() => [...document.querySelectorAll(".orders-id")][index].click());
    await waitFor(() => shipmentButton() && !shipmentButton().disabled, "scoped shipment button unavailable");
    flushSync(() => shipmentButton().click());
    await waitFor(() => pendingShipmentDetails.length === 1, "scoped shipment request missing");
    return pendingShipmentDetails.shift();
  };
  const oldShipment = await openShipping(0);
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  const newShipment = await openShipping(1);
  assert(oldShipment.signal.aborted && oldShipment.path.includes("/stores/one/") && newShipment.path.includes("/stores/two/"),
    "same order ID across stores reused the wrong request or did not abort the old response");
  // Chromium can deliver the previous dialog's close event after reopening.
  flushSync(() => document.querySelector('dialog[aria-labelledby="orders-detail-title"]').dispatchEvent(new Event("close")));
  newShipment.resolve(Response.json({ ...order, order_id: "200002", shipment: shipmentDetails }));
  await waitFor(() => document.querySelector(".orders-detail-fields")?.textContent.includes("Fixture official carrier"),
    "new scoped shipment did not render");
  oldShipment.resolve(Response.json({ ...order, order_id: "200002", shipment: {
    ...shipmentDetails, details: { ...shipmentDetails.details, carrier_name: "Stale carrier" },
  } }));
  await new Promise(resolve => setTimeout(resolve, 20));
  assert(!document.querySelector(".orders-detail-fields")?.textContent.includes("Stale carrier"),
    "delayed result escaped the store selection guard");
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  exerciseImages = true;
  const testingThumbnail = "https://http2.mlstatic.com/D_700-I.jpg";
  select("订单店铺", "one");
  select("站点和物流模式", "MLB|remote");
  const imageField = searchInput();
  flushSync(() => {
    const nativeSetter = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, "value").set;
    nativeSetter.call(imageField, "");
    imageField.dispatchEvent(new Event("input", { bubbles: true }));
  });
  order.order_items[0] = { ...order.order_items[0], item_id: "MLB700" };
  flushSync(() => submitSearch().click());
  await waitFor(() => imageRequests.length === 1, "new order did not start non-blocking image backfill");
  assert(document.querySelectorAll(".orders-id").length === 2, "image read blocked order rows");
  const staleImage = imageRequests.shift();
  select("订单店铺", "two");
  await waitFor(() => imageRequests.length === 1, "next store did not start independent image backfill");
  const matchingImage = imageRequests.shift();
  assert(staleImage.signal.aborted && staleImage.path.includes("/stores/one/") &&
    matchingImage.path.includes("/stores/two/"), "image work did not follow the store scope");
  staleImage.resolve(Response.json({ ...order, order_items: [{ ...order.order_items[0], image_url: "https://http2.mlstatic.com/stale.jpg" }] }));
  matchingImage.resolve(Response.json({ ...order, order_items: [{ ...order.order_items[0], image_url: testingThumbnail }] }));
  await waitFor(() => document.querySelectorAll(".orders-product-preview img").length === 2,
    "same-shop orders for one item did not share the image");
  assert([...document.querySelectorAll(".orders-product-preview img")].every(img => img.src === testingThumbnail) &&
    imageRequests.length === 0, "late image update or repeat item request escaped its scope");
  // A user operation interrupts automatic images without changing the view.
  distinctImageItems = true;
  order.order_items[0] = { ...order.order_items[0], item_id: "MLB701" };
  flushSync(() => [...document.querySelectorAll(".orders-heading-actions button")].find(btn => btn.textContent === "刷新订单").click());
  await waitFor(() => imageRequests.length === 1, "new automatic image batch missing");
  const interrupted = imageRequests.shift();
  pauseOrderRefresh = true;
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector(".orders-detail-summary") && !document.querySelector(".orders-detail-refresh").disabled,
    "order detail not ready for image interruption");
  flushSync(() => document.querySelector(".orders-detail-refresh").click());
  await waitFor(() => pausedRefreshes.length === 1 && interrupted.signal.aborted, "user refresh did not cancel images");
  interrupted.resolve(Response.json({ ...order, order_items: [{ ...order.order_items[0], image_url: "https://http2.mlstatic.com/interrupted.jpg" }] }));
  pausedRefreshes.shift()(Response.json(order));
  pauseOrderRefresh = false;
  await waitFor(() => imageRequests.length === 1, "automatic images did not resume on the same view");
  const resumed = imageRequests.shift();
  resumed.resolve(Response.json({ ...order, order_items: [{ ...order.order_items[0], image_url: testingThumbnail }] }));
  await waitFor(() => imageRequests.length === 1, "unattempted image was not processed after interruption");
  const remaining = imageRequests.shift();
  assert(decodeURIComponent(new URL(remaining.path, location.href).pathname).endsWith("/orders/103/image-refresh"),
    "remaining distinct item was not requested: " + remaining.path);
  remaining.resolve(Response.json({ error: "order_rate_limited" }, { status: 429 }));
  await waitFor(() => [...document.querySelectorAll(".orders-heading-actions button")].some(btn => btn.textContent === "补全本页主图"),
    "image batch did not stop after rate limit");
  await new Promise(resolve => setTimeout(resolve, 500));
  assert(imageRequests.length === 0, "rate-limited automatic batch retried itself");
  // Exercise actual layout and the visible money / Pack / review fields.
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  const tabs = document.querySelector(".orders-tabs");
  const originalWidth = tabs.style.width;
  for (const width of [1024, 1280, 1440]) {
    tabs.style.width = width + "px";
    const bounds = tabs.getBoundingClientRect();
    assert(getComputedStyle(tabs).flexWrap === "wrap" &&
      [...tabs.querySelectorAll("button")].every(btn => {
        const rect = btn.getBoundingClientRect();
        return rect.left >= bounds.left - 1 && rect.right <= bounds.right + 1;
      }), "order status toolbar overflow at " + width);
  }
  tabs.style.width = originalWidth;
  exerciseImages = false;
  order.pack_id = "920";
  order.sales_amount_usd = "22.59000003";
  order.order_items = [{ ...order.order_items[0], quantity: 2, image_url: testingThumbnail }];
  flushSync(() => [...document.querySelectorAll(".orders-heading-actions button")].find(btn => btn.textContent === "刷新订单").click());
  await waitFor(() => document.querySelectorAll(".orders-id").length === 2 &&
    [...document.querySelectorAll(".orders-id")].every(btn => btn.textContent === "920") &&
    document.querySelector(".orders-money")?.textContent === "USD 22.59", "Pack ID or rounded list sale missing");
  const packRows = [...document.querySelectorAll(".orders-id")].map(btn => btn.closest("tr"));
  assert(packRows[0].textContent.includes("内部 Order ID：102") && packRows[1].textContent.includes("内部 Order ID：103") &&
    packRows[0].querySelector(".orders-product-meta").textContent.includes("2 件") &&
    packRows[0].querySelector(".orders-money").title.includes("22.59000003"), "Pack rows, exact sale tooltip or count incorrect");
  flushSync(() => packRows[0].querySelector(".orders-id").click());
  await waitFor(() => document.querySelector("#orders-detail-title")?.textContent === "920" &&
    document.querySelector(".orders-detail-summary strong")?.textContent === "USD 22.59" &&
    [...document.querySelectorAll(".orders-detail-fields button")].some(btn => btn.textContent.includes("102") && btn.textContent.includes("复制")),
    "detail must expose Pack and internal Order ID copy");
  assert(document.querySelector(".orders-detail-summary strong").title.includes("22.59000003"), "detail exact sale missing");
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  delete order.sales_amount_usd;
  order.order_items_review_required = true;
  order.order_items = [...order.order_items, { ...order.order_items[0] },
    { ...order.order_items[0], item_id: "MLB703", unit_price: "" }];
  flushSync(() => [...document.querySelectorAll(".orders-heading-actions button")].find(btn => btn.textContent === "刷新订单").click());
  await waitFor(() => document.querySelector(".orders-money")?.textContent === "待核实" &&
    document.querySelector(".orders-product-meta")?.textContent.includes("件数待核实"), "legacy review list must not show unverified total");
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector(".orders-detail-summary strong")?.textContent === "待核实" &&
    document.querySelector(".orders-detail-block h3")?.textContent.includes("件数待核实") &&
    document.querySelector(".orders-alert")?.textContent.includes("无法确认是否重复") &&
    document.querySelectorAll(".orders-line-item").length === 3, "legacy review detail must retain all lines and withhold totals");
  const missingPrice = [...document.querySelectorAll(".orders-line-content span")]
    .find(span => span.textContent.includes("待核实 ×"));
  assert(missingPrice && !missingPrice.hasAttribute("title") &&
    !document.querySelector(".orders-detail-summary strong").hasAttribute("title"),
    "missing monetary values must not show an empty exact-money tooltip");
  document.body.dataset.result = "pass";
} catch (error) {
  document.body.dataset.result = "fail";
  document.body.append(error.message);
}
`;

test("scope changes discard delayed order results and a purchase conflict requires rereading", { timeout: 45_000 }, async () => {
  const profile = await mkdtemp(join(tmpdir(), "una-order-scope-"));
  const server = await createServer({
    server: { port: 0 },
    optimizeDeps: { include: ["react", "react-dom", "react-dom/client", "react/jsx-runtime"] },
    plugins: [{
      name: "orders-scope-fixture",
      enforce: "pre",
      resolveId(id) {
        if (id === "./auth") return "\0orders-test-auth";
        if (id === "/orders-scope-fixture.js") return "\0orders-scope-fixture.js";
      },
      load(id) {
        if (id === "\0orders-scope-fixture.js") return fixture;
        if (id === "\0orders-test-auth") return `
          const signal = new AbortController().signal;
          export const authLifetime = () => signal;
          export const getAccessToken = async () => "fixture";
          export const rejectSession = () => {};
        `;
      },
      configureServer(server) {
        server.middlewares.use(async (req, res, next) => {
          if (req.url === "/orders-scope-fixture") {
            res.setHeader("Content-Type", "text/html; charset=utf-8");
            res.end(await server.transformIndexHtml(req.url,
              '<!doctype html><html><head></head><body><main></main><script type="module" src="/orders-scope-fixture.js"></script></body></html>'));
          } else next();
        });
      },
    }],
  });
  try {
    await server.listen();
    const { port } = server.httpServer.address();
    const chrome = process.env.CHROME_BIN ?? (process.platform === "win32"
      ? "C:/Program Files/Google/Chrome/Application/chrome.exe" : "google-chrome");
    const { stdout } = await promisify(execFile)(chrome, [
      "--headless", "--no-sandbox", "--disable-background-networking", `--user-data-dir=${profile}`,
      "--dump-dom", "--virtual-time-budget=5000", `http://127.0.0.1:${port}/orders-scope-fixture`,
    ], { timeout: 30_000 });
    assert.ok(stdout.includes('data-result="pass"'), stdout.slice(-1000));
  } finally {
    await server.close();
    await rm(profile, { recursive: true, force: true, maxRetries: 5, retryDelay: 200 });
  }
});
