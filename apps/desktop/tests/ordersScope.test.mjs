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
import OrdersPage from "/src/OrdersPage.tsx";

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
];
const stores = ["one", "two"].map((id) => ({
  id, nickname: id, status: id === "one" ? "token_expired" : "connected",
  marketplaces_status: "verified", marketplaces: markets,
}));
const deferred = [];
const purchaseRequests = [];
const purchaseFilters = [];
const order = { order_id: "102", status: "paid", order_items: [], total_amount: "12", currency_id: "USD" };
window.fetch = (path, options) => {
  const { pathname } = new URL(path, location.href);
  if (pathname === "/api/v1/stores") return Promise.resolve(Response.json({ items: stores }));
  if (pathname.endsWith("/orders")) {
    const filter = new URL(path, location.href).searchParams.get("purchase") ?? "";
    purchaseFilters.push(filter);
    const linked = { ...order, order_id: "103", purchase_order_id: "12345678", purchase_cost_cny: "16" };
    const rows = filter === "linked" ? [linked] : filter === "unlinked" ? [order] : [order, linked];
    return Promise.resolve(Response.json({ items: rows, has_more: false, purchase_filter: filter }));
  }
   if (pathname.endsWith("/sync-state")) return Promise.resolve(Response.json({ next_offset: 0, can_reset: false }));
  if (pathname.endsWith("/orders/102")) return Promise.resolve(Response.json(order));
  if (pathname.endsWith("/orders/103")) return Promise.resolve(Response.json({
    ...order, order_id: "103", purchase_order_id: "12345678", purchase_cost_cny: "16",
  }));
  if (pathname.endsWith("/purchase")) {
    return new Promise((resolve) => purchaseRequests.push({ resolve, ...options }));
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
  flushSync(() => document.querySelector(".orders-id").click());
  await waitFor(() => document.querySelector("dialog.orders-modal")?.open,
    "order details did not open in a modal sheet");
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
  flushSync(() => filter("待关联").click());
  await waitFor(() => purchaseFilters.includes("unlinked") &&
    document.querySelector(".orders-purchase-cell button")?.textContent === "关联采购",
    "unlinked filter or purchase CTA missing");
  flushSync(() => filter("已关联").click());
  await waitFor(() => purchaseFilters.includes("linked") &&
    document.querySelector(".orders-purchase-cell strong")?.textContent === "已关联",
    "linked filter or purchase label missing");
  assert(document.querySelector(".orders-purchase-cell")?.textContent.includes("采购成本 ¥16"),
    "linked purchase cost missing");
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
    source: "manual", updated_at: "2026-11-20T00:00:00.123456Z" };
  currentRead.resolve(Response.json({ purchase }));
  staleRead.resolve(Response.json({ purchase: { ...purchase, purchase_order_id: "87654321" } }));
  const save = () => purchaseButton("保存采购关联");
  await waitFor(() => save() && !save().disabled, "purchase form unavailable after read");
  assert(staleRead.signal.aborted, "previous purchase read was not canceled");
  flushSync(() => save().click());
  const conflict = await takePurchase("POST");
  const submitted = JSON.parse(conflict.body);
  assert(submitted.purchase_order_id === purchase.purchase_order_id, "previous purchase read entered current form");
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
  const sheet = document.querySelector("dialog.orders-modal");
  assert(sheet?.open, "order sheet closed unexpectedly during purchase workflow");
  flushSync(() => document.querySelector('button[aria-label="关闭订单详情"]').click());
  await waitFor(() => !sheet.open, "order sheet did not close");
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
