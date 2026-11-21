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
import { displayOrderStatus, hasProcurement, orderMoney, orderStatusLabel, orderVariationText,
  shouldApplyOrderRefresh, shouldStopSLABatch } from "/src/orders.ts";

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
const order = { order_id: "102", status: "paid", order_items: [], total_amount: "12", currency_id: "USD",
  shipment_id: "87123456" };
let slaFetchCount = 0;
window.fetch = (path, options) => {
  const { pathname } = new URL(path, location.href);
  if (pathname === "/api/v1/stores") return Promise.resolve(Response.json({ items: stores }));
  if (pathname.endsWith("/orders/102/sla")) {
    slaFetchCount++;
    order.dispatch_deadline = new Date(Date.now() + 3 * 86_400_000).toISOString();
    order.dispatch_sla_status = "on_time";
    order.dispatch_checked_at = new Date().toISOString();
    return Promise.resolve(Response.json(order));
  }
  if (pathname.endsWith("/orders/102/refresh")) {
    order.status = "partially_refunded";
    return Promise.resolve(Response.json(order));
  }
  if (pathname.endsWith("/orders/search")) {
    const uri = new URL(path, location.href);
    const q = uri.searchParams.get("q") ?? "";
    const storeFilter = pathname.split("/")[4];
    const site = uri.searchParams.get("site") ?? "";
    const purchaseFilter = uri.searchParams.get("purchase") ?? "";
    searchRequests.push({ q, storeFilter, site, uri });
    const lemon = { ...order, order_id: "20002", store_id: "one", site_id: "MLM", logistic_type: "remote",
      purchase_order_id: "", order_items: [{ item_id: "MLM100", title: "Lemon massage tool",
        quantity: 1, unit_price: "12", currency_id: "USD" }] };
    const another = { ...lemon, store_id: "two", site_id: "MLB",
      order_items: [{ ...lemon.order_items[0], item_id: "MLB200" }] };
    const rows = q === "20002" || q.toLowerCase() === "lemon"
      ? storeFilter === "all" ? [lemon, another] : [lemon]
      : site === "*" ? [{ ...lemon, order_id: "10001", site_id: "MLB" }, lemon] : [lemon];
    return Promise.resolve(Response.json({ items: rows, has_more: false, purchase_filter: purchaseFilter,
      query: q, site_filter: site, store_filter: storeFilter }));
  }
  if (pathname.endsWith("/orders")) {
    const filter = new URL(path, location.href).searchParams.get("purchase") ?? "";
    purchaseFilters.push(filter);
    const linked = { ...order, order_id: "103", shipment_id: "",
      purchase_order_id: "12345678", purchase_cost_cny: "16" };
    const rows = filter === "linked" ? [linked] : filter === "unlinked" ? [order] : [order, linked];
    return Promise.resolve(Response.json({ items: rows, has_more: false, purchase_filter: filter }));
  }
   if (pathname.endsWith("/sync-state")) return Promise.resolve(Response.json({ next_offset: 0, can_reset: false }));
  if (pathname.endsWith("/orders/102")) return Promise.resolve(Response.json(order));
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
  createRoot(document.querySelector("main")).render(React.createElement(OrdersPage, {
    installingUpdate: false, beginSync() { active++; return () => active--; },
  }));
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
  assert(shouldStopSLABatch(502, "order_platform_unavailable"), "upstream 502 must stop");
  assert(shouldStopSLABatch(504, "order_timeout"), "timeout must stop");
  assert(shouldStopSLABatch(429, "order_rate_limited"), "429 must stop");
  assert(shouldStopSLABatch(0, ""), "network loss must stop");
  assert(!shouldStopSLABatch(404, "order_sla_no_shipment"), "one no-shipment row may be skipped");
  const now = Date.UTC(2026, 10, 21, 0, 0);
  const sample = (hours, status = "on_time") => ({
    ...order, status: "paid", dispatch_sla_status: status,
    dispatch_deadline: new Date(now + hours * 3600_000).toISOString(),
  });
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
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector("dialog.orders-modal:not(.orders-purchase-modal)")?.open,
    "real order detail modal did not open after purchase modal closed");
  flushSync(() => [...document.querySelectorAll(".orders-detail-refresh")][0].click());
  await waitFor(() => document.querySelector(".orders-detail-summary")?.textContent.includes("部分退款"),
    "manual order refresh did not display refunded state");
  assert(!document.querySelector(".orders-money-net") &&
    ![...document.querySelectorAll("button")].some((item) => item.textContent === "订单设置"),
    "nonexistent finance editor should not be visible");

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
    nativeSetter.call(input, "20002");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  flushSync(() => submitSearch().click());
  await waitFor(() => searchRequests.some((q) => q.q === "20002" && q.site === "*") &&
    document.querySelector('select[aria-label="站点和物流模式"]')?.value === "all" &&
    document.querySelector(".orders-id")?.textContent === "20002",
    "cross-site Order ID must change to all sites and show exactly the matching order");
  assert(document.querySelectorAll(".orders-id").length === 1, "numeric search returned extra orders");
  flushSync(() => document.querySelector(".orders-row-purchase").click());
  const crossSitePurchase = await takePurchase("GET");
  assert(new URL(crossSitePurchase.path, location.href).searchParams.get("site") === "MLM" &&
    crossSitePurchase.path.includes("/stores/one/orders/20002/purchase"),
    "cross-site purchase opened under the wrong site or store");
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
  assert(secondStorePurchase.path.includes("/stores/two/orders/20002/purchase") &&
    new URL(secondStorePurchase.path, location.href).searchParams.get("site") === "MLB",
    "second-store row must not reuse first-store purchase scope");
  secondStorePurchase.resolve(Response.json({ purchase: null }));
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
