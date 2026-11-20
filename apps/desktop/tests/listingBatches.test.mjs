import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { registerHooks } from "node:module";
import test from "node:test";
import ts from "typescript";
import { runListingBatches } from "../src/listingBatches.ts";

const storesURL = new URL("../src/stores.ts", import.meta.url).href;
const diagnosticsSource = readFileSync(new URL("../src/DevelopmentDiagnostics.tsx", import.meta.url), "utf8");
const { outputText } = ts.transpileModule(
  diagnosticsSource.replace("import.meta.env.VITE_UNA_SELLER_DEV_DIAGNOSTICS", '"1"'),
  { compilerOptions: { module: ts.ModuleKind.ESNext, target: ts.ScriptTarget.ES2022, jsx: ts.JsxEmit.ReactJSX } },
);
const diagnosticsURL = "data:text/javascript," + encodeURIComponent(outputText
  .replace('"react"', JSON.stringify(import.meta.resolve("react")))
  .replace('"react/jsx-runtime"', JSON.stringify(import.meta.resolve("react/jsx-runtime"))));
const authURL = "data:text/javascript," + encodeURIComponent(`
  let controller = new AbortController();
  export const authLifetime = () => controller.signal;
  export const getAccessToken = async (_refresh, report) => { report?.(true); return 'fixture'; };
  export function rejectSession() { controller.abort(); controller = new AbortController(); }
`);
const hooks = registerHooks({
  resolve(specifier, context, nextResolve) {
    if (context.parentURL === storesURL) {
      if (specifier === "./auth") {
        return { url: authURL, shortCircuit: true };
      }
      if (specifier === "./collections") {
        return { url: "data:text/javascript,export const apiBase = 'https://fixture.invalid';", shortCircuit: true };
      }
      if (specifier === "./DevelopmentDiagnostics") return { url: diagnosticsURL, shortCircuit: true };
    }
    return nextResolve(specifier, context);
  },
});
const { cachedStores, listStores, checkStoreConnection, ensureStoreConnection, cachedListingPage, loadListingPage,
  disconnectStore, invalidateListingPages, storeRequest } = await import(storesURL);
const { rejectSession } = await import(authURL);
const { beginRequestDiagnostic, clearDiagnostics, diagnosticsJSON } = await import(diagnosticsURL);
hooks.deregister();

test("diagnostics retain only allowlisted fields and rotate within the current account", async (t) => {
  clearDiagnostics();
  t.mock.method(globalThis, "fetch", async () => Response.json({ code: "fixture-secret" }, {
    status: 409, headers: { "Server-Timing": 'api;dur=1.2, private;dur=3, read;dur=5;desc="fixture-secret"' },
  }));
  const path = "/api/v1/stores/fixture-secret/published-listings/fixture-secret/description?state=fixture-secret";
  await assert.rejects(storeRequest(path, "GET", undefined, true));
  let events = JSON.parse(diagnosticsJSON());
  assert.deepEqual(events.map(({ operation, phase, cache_hit }) => [operation, phase, cache_hit]), [
    ["description_get", "token_preflight", true],
    ["description_get", "response_headers", false],
    ["description_get", "failed", false],
  ]);
  assert.equal(events[2].http_status, 409);
  assert.equal(events[2].error_code, "http_409");
  assert.equal(events[2].server_timing, "api;dur=1.2");
  assert.doesNotMatch(diagnosticsJSON(), /fixture-secret|Authorization/i);

  const previousAccount = beginRequestDiagnostic("/api/v1/stores", "GET");
  clearDiagnostics();
  previousAccount("failed");
  assert.deepEqual(JSON.parse(diagnosticsJSON()), []);
  const report = beginRequestDiagnostic("/api/v1/stores/example/published-listings", "POST");
  report("response_body", undefined, "listing_commit_unknown");
  assert.equal(JSON.parse(diagnosticsJSON())[0].error_code, "listing_commit_unknown");
  for (let i = 0; i < 200; i++) report("response_headers", new Response(null, { status: i < 199 ? 200 : 204 }));
  events = JSON.parse(diagnosticsJSON());
  assert.equal(events.length, 200);
  assert.equal(events[0].phase, "response_headers");
  assert.equal(events.at(-1).http_status, 204);
  clearDiagnostics();
});

const batch = (offset) => ({
  saved: 4, processed: 5, skipped: 1, needs_review: 0, partial: 0,
  next_offset: offset + 5, scan_required: false,
});

test("one click runs at most ten batches and restarts only the first", async () => {
  const calls = [];
  let saved = 0;
  await runListingBatches(0, true, new AbortController().signal, async (offset, restart) => {
    calls.push([offset, restart]);
    return batch(offset);
  }, (result) => { saved += result.saved; });
  assert.deepEqual(calls, Array.from({ length: 10 }, (_, i) => [i * 5, i === 0]));
  assert.equal(saved, 40);
});

test("stop aborts the active request, then the next click resumes the committed offset", async () => {
  const controller = new AbortController();
  let committed = 0;
  const calls = [];
  await assert.rejects(runListingBatches(0, false, controller.signal, async (offset, restart, signal) => {
    calls.push([offset, restart]);
    if (offset === 5) {
      controller.abort();
      signal.throwIfAborted();
    }
    committed = offset + 5;
    return batch(offset);
  }, () => {}), { name: "AbortError" });
  await runListingBatches(committed, false, new AbortController().signal, async (offset, restart) => {
    calls.push([offset, restart]);
    return { ...batch(offset), next_offset: null };
  }, () => {});
  assert.deepEqual(calls, [[0, false], [5, false], [5, false]]);
});

test("a partial response with an interruption stops automatically", async () => {
  let calls = 0;
  await runListingBatches(5, false, new AbortController().signal, async () => {
    calls++;
    return { ...batch(5), interruption: { code: "listing_rate_limited", retryable: true } };
  }, () => {});
  assert.equal(calls, 1);
});

test("a lost restart response stops the loop and a later click uses the server cursor", async (t) => {
  const calls = [];
  let committed = 0;
  t.mock.method(globalThis, "fetch", async (url, { method }) => {
    const query = new URL(url).searchParams;
    const offset = Number(query.get("offset"));
    const restart = query.get("restart") === "1";
    calls.push([method, offset, restart]);
    if (method === "GET") return Response.json({ next_offset: committed });
    if (restart) {
      committed = 5;
      throw new TypeError("private network details");
    }
    return Response.json({ ...batch(offset), next_offset: null });
  });
  const signal = new AbortController().signal;
  const request = (offset, restart, abort) => storeRequest(
    "/listings?offset=" + offset + (restart ? "&restart=1" : ""), "POST", abort, true,
  );
  await assert.rejects(runListingBatches(0, true, signal, request, () => assert.fail("unexpected progress")), {
    message: "未收到可读取的同步响应，请刷新后以服务端同步位置为准",
  });
  assert.deepEqual(calls, [["POST", 0, true]]);
  const page = await storeRequest("/listings?offset=0", "GET", signal, true);
  await runListingBatches(page.next_offset, false, signal, request, () => {});
  assert.deepEqual(calls, [["POST", 0, true], ["GET", 0, false], ["POST", 5, false]]);
});

test("listing timeouts, unknown commits, unreadable responses, and cancellation stay distinct", async (t) => {
  const replies = [
    () => Response.json({ code: "listing_timeout" }, { status: 504 }),
    () => Response.json({ code: "listing_commit_unknown" }, { status: 502 }),
    () => new Response("{"),
    () => { throw new DOMException("Canceled", "AbortError"); },
  ];
  const fetch = t.mock.method(globalThis, "fetch", async () => replies.shift()());
  for (const message of [
    "商品请求超时，请刷新后以服务端同步位置为准",
    "同步提交结果待确认，请刷新后以服务端同步位置为准",
    "未收到可读取的同步响应，请刷新后以服务端同步位置为准",
  ]) {
    await assert.rejects(storeRequest("/listings", "POST", undefined, true), { message });
  }
  await assert.rejects(storeRequest("/listings", "POST", undefined, true), { name: "AbortError" });
  await assert.rejects(storeRequest("/listings", "POST", AbortSignal.abort(), true), { name: "AbortError" });
  assert.equal(fetch.mock.callCount(), 4);
});

const store = { id: "store-1", status: "token_expired", token_expires_at: "2020-01-01T00:00:00Z" };
const freshStore = () => ({ ...store, status: "connected", token_expires_at: new Date(Date.now() + 3600_000).toISOString() });
const mlm = "/api/v1/stores/store-1/published-listings?site=MLM&logistic_type=remote&offset=0";

test("store reads coalesce, then one detached renewal gates platform requests and rereads authorization", async (t) => {
  rejectSession();
  const calls = [];
  let completeCheck;
  const check = new Promise((resolve) => { completeCheck = resolve; });
  t.mock.method(globalThis, "fetch", async (url, { method }) => {
    const path = new URL(url).pathname;
    calls.push([method, path]);
    if (path.endsWith("connection-check")) return check;
    if (path === "/api/v1/stores") return Response.json({ items: [calls.length === 1 ? store : freshStore()] });
    return Response.json({ description: "fixture" });
  });
  const listing = listStores();
  assert.equal(listStores(), listing);
  assert.deepEqual(await listing, [store]);
  const cancel = new AbortController();
  const stopped = ensureStoreConnection(store.id, cancel.signal);
  const sync = ensureStoreConnection(store.id).then(() => storeRequest(mlm, "POST", undefined, true));
  const desc = ensureStoreConnection(store.id).then(() => storeRequest(mlm, "GET", undefined, true));
  cancel.abort();
  const aborted = assert.rejects(stopped, { name: "AbortError" });
  await Promise.resolve();
  assert.deepEqual(calls, [["GET", "/api/v1/stores"], ["POST", "/api/v1/stores/store-1/connection-check"]]);
  completeCheck(Response.json({ status: "connected" }));
  await Promise.all([sync, desc, aborted]);
  assert.equal(cachedStores()[0].status, "connected");
  assert.deepEqual(calls.slice(2).map(([method, path]) => [method, path.split("/").at(-1)]), [
    ["GET", "stores"], ["POST", "published-listings"], ["GET", "published-listings"],
  ]);
});

test("lost response is reconciled before a later retry or manual check", async (t) => {
  for (const recovery of ["automatic", "manual", "already renewed", "timeout"]) {
    rejectSession();
    const calls = [];
    let current = store;
    const deadline = new AbortController();
    const timeout = AbortSignal.timeout.bind(AbortSignal);
    const timer = t.mock.method(AbortSignal, "timeout", (ms) =>
      recovery === "timeout" && ms === 25_000 ? deadline.signal : timeout(ms));
    const fetch = t.mock.method(globalThis, "fetch", async (_url, { method, signal }) => {
      calls.push(method);
      if (method === "GET") return Response.json({ items: [current] });
      if (calls.filter((call) => call === "POST").length === 1) {
        if (recovery === "already renewed" || recovery === "timeout") current = freshStore();
        if (recovery === "timeout") {
          return new Promise((_resolve, reject) => {
            signal.addEventListener("abort", () => reject(signal.reason), { once: true });
            deadline.abort(new DOMException("deadline", "TimeoutError"));
          });
        }
        throw new TypeError("Failed to fetch");
      }
      current = freshStore();
      return Response.json({ status: "connected" });
    });
    await listStores();
    if (recovery === "already renewed" || recovery === "timeout") {
      await ensureStoreConnection(store.id);
      assert.deepEqual(calls, ["GET", "POST", "GET"]);
    } else {
      await assert.rejects(ensureStoreConnection(store.id), /续签结果尚未确认/);
      assert.deepEqual(calls, ["GET", "POST", "GET"]);
      if (recovery === "manual") await checkStoreConnection(store.id);
      else await ensureStoreConnection(store.id);
      assert.deepEqual(calls, ["GET", "POST", "GET", "GET", "POST", "GET"]);
    }
    assert.equal(cachedStores()[0].status, "connected");
    fetch.mock.restore();
    timer.mock.restore();
  }
});

test("terminal store states block automatic and manual rotation", async (t) => {
  for (const [status, finalStatus, expected] of [
    [409, "refresh_unknown", /续签状态待确认/],
    [403, "reauth_required", /需要重新授权/],
  ]) {
    rejectSession();
    let posts = 0;
    const fetch = t.mock.method(globalThis, "fetch", async (_url, { method }) => {
      if (method === "GET") return Response.json({ items: [{ ...store, status: posts ? finalStatus : store.status }] });
      posts++;
      return Response.json({}, { status });
    });
    await listStores();
    await assert.rejects(ensureStoreConnection(store.id), expected);
    await assert.rejects(ensureStoreConnection(store.id), expected);
    await assert.rejects(checkStoreConnection(store.id), expected);
    assert.equal(posts, 1);
    assert.equal(cachedStores()[0].status, finalStatus);
    fetch.mock.restore();
  }
});

test("pages are available during revalidation, isolated by scope, and invalidated on sync, disconnect or session change", async (t) => {
  rejectSession();
  const page = { items: [{ id: "one" }] };
  let release;
  const pending = new Promise((resolve) => { release = resolve; });
  let delayed = false;
  t.mock.method(globalThis, "fetch", async (_url, { method }) => {
    if (method === "POST") return Response.json({ status: "disconnected" });
    if (delayed) await pending;
    return Response.json(page);
  });
  const signal = new AbortController().signal;
  const coldStart = performance.now();
  assert.equal(cachedListingPage(mlm), undefined);
  delayed = true;
  const coldRead = loadListingPage(mlm, signal);
  setTimeout(release, 180);
  await coldRead;
  const coldMs = performance.now() - coldStart;
  const warmStart = performance.now();
  assert.deepEqual(cachedListingPage(mlm), page);
  assert.equal(JSON.parse(diagnosticsJSON()).some((x) => x.operation === "listings_get" && x.cache_hit), true);
  const warmMs = performance.now() - warmStart;
  t.diagnostic(`fixture cold data ${coldMs.toFixed(1)} ms, warm memory data ${warmMs.toFixed(2)} ms`);
  for (const other of [mlm.replace("MLM", "MLB"), mlm.replace("remote", "fulfillment"), mlm.replace("offset=0", "offset=20"), mlm.replace("store-1", "store-2")]) {
    assert.equal(cachedListingPage(other), undefined);
  }
  const oldRead = loadListingPage(mlm, signal);
  invalidateListingPages(store.id);
  await oldRead;
  assert.equal(cachedListingPage(mlm), undefined);
  await loadListingPage(mlm, signal);
  await disconnectStore(store.id);
  assert.equal(cachedListingPage(mlm), undefined);
  await loadListingPage(mlm, signal);
  const previousUser = loadListingPage(mlm, signal);
  rejectSession();
  await assert.rejects(previousUser, { name: "AbortError" });
  assert.equal(cachedListingPage(mlm), undefined);
  assert.equal(cachedStores(), undefined);
});

test("an expired login does not replay a POST", async (t) => {
  rejectSession();
  const fetch = t.mock.method(globalThis, "fetch", async () => new Response(null, { status: 401 }));
  await assert.rejects(storeRequest(mlm, "POST", undefined, true), /确认状态后重新发起/);
  assert.equal(fetch.mock.callCount(), 1);
});
