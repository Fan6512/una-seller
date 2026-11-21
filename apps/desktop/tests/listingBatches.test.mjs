import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { registerHooks } from "node:module";
import test from "node:test";
import ts from "typescript";
import { runListingBatches } from "../src/listingBatches.ts";
// Store login checks share this suite's Node strip-types fixture and store-session mocks.
import { checkStoresOnLogin, resolveManualStoreCheck, storeHealthFromList } from "../src/storeLoginCheck.ts";

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
  export const getAccessToken = async (refresh, report) => { report?.(!refresh); return 'fixture'; };
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
  disconnectStore, invalidateListingPages, storeRequest, readableStoreError, watchStoreList } = await import(storesURL);
const { rejectSession, authLifetime } = await import(authURL);
const { beginRequestDiagnostic, clearDiagnostics, diagnosticsJSON, recordCacheHit } = await import(diagnosticsURL);
hooks.deregister();

test("diagnostics retain only allowlisted fields and rotate within the current account", async (t) => {
  clearDiagnostics();
  t.mock.method(globalThis, "fetch", async () => Response.json({ code: "fixture-secret" }, {
    status: 409, headers: { "Server-Timing": 'api;dur=1.2, private;dur=3, read;dur=5;desc="fixture-secret"' },
  }));
  const path = "/api/v1/stores/fixture-secret/published-listings/fixture-secret/description?state=fixture-secret";
  await assert.rejects(storeRequest(path, "GET", undefined, true));
  let events = JSON.parse(diagnosticsJSON());
  assert.deepEqual(events.map(({ operation, phase, cache_hit, token_cached }) =>
    [operation, phase, cache_hit, token_cached]), [
    ["description_get", "token_preflight", undefined, true],
    ["description_get", "response_headers", undefined, undefined],
    ["description_get", "failed", undefined, undefined],
  ]);
  assert.ok(events[0].request_id > 0);
  assert.ok(events.every((event) => event.request_id === events[0].request_id));
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
  report("failed", new Response(null, { status: 400 }), "listing_review_invalid");
  report("failed", new Response(null, { status: 400 }), "listing_offset");
  assert.deepEqual(JSON.parse(diagnosticsJSON()).slice(-2).map(({ error_code }) => error_code),
    ["listing_review_invalid", "listing_offset"]);
  for (let i = 0; i < 200; i++) report("response_headers", new Response(null, { status: i < 199 ? 200 : 204 }));
  events = JSON.parse(diagnosticsJSON());
  assert.equal(events.length, 200);
  assert.equal(events[0].phase, "response_headers");
  assert.equal(events.at(-1).http_status, 204);
  clearDiagnostics();
});

test("GET 401 recovered by refresh does not count as failure", async (t) => {
  rejectSession();
  clearDiagnostics();
  let calls = 0;
  const fetch = t.mock.method(globalThis, "fetch", async () => {
    calls++;
    return calls === 1 ? new Response(null, { status: 401 }) : Response.json({ items: [] });
  });
  assert.deepEqual(await storeRequest("/api/v1/stores", "GET"), { items: [] });
  assert.equal(fetch.mock.callCount(), 2);
  const events = JSON.parse(diagnosticsJSON());
  assert.deepEqual(events.filter(({ phase }) => phase === "response_headers").map(({ http_status }) => http_status), [401, 200]);
  assert.equal(events.filter(({ phase }) => phase === "failed").length, 0);
  assert.deepEqual(events.filter(({ phase }) => phase === "token_preflight").map(({ token_cached }) => token_cached), [true, false]);
  assert.ok(events.every((event) => event.request_id === events[0].request_id));
  assert.equal(events.at(-1).cache_hit, false);
});

test("GET 401 still fails when refresh cannot authorize", async (t) => {
  rejectSession();
  clearDiagnostics();
  const fetch = t.mock.method(globalThis, "fetch", async () => new Response(null, { status: 401 }));
  await assert.rejects(storeRequest("/api/v1/stores", "GET"), /登录已过期/);
  assert.equal(fetch.mock.callCount(), 2);
  const events = JSON.parse(diagnosticsJSON());
  assert.equal(events.filter(({ phase }) => phase === "failed").length, 1);
  assert.equal(events.at(-1).error_code, "http_401");
});

test("interleaved diagnostic operations retain separate request IDs", () => {
  clearDiagnostics();
  const first = beginRequestDiagnostic("/api/v1/stores", "GET");
  const second = beginRequestDiagnostic("/api/v1/stores/sample/published-listings", "GET");
  first("response_headers", new Response(null, { status: 200 }));
  second("response_headers", new Response(null, { status: 200 }));
  first("response_body", new Response(null, { status: 200 }));
  const events = JSON.parse(diagnosticsJSON());
  assert.deepEqual(events.map(({ request_id }) => request_id), [1, 2, 1]);
  assert.deepEqual(events.map(({ cache_hit }) => cache_hit), [undefined, undefined, false]);
  recordCacheHit("/api/v1/stores");
  const cached = JSON.parse(diagnosticsJSON()).at(-1);
  assert.equal(cached.request_id, 3);
  assert.equal(cached.cache_hit, true);
  assert.equal(cached.token_cached, undefined);
});

test("8-second timeout has an actionable Chinese message", () => {
  assert.equal(readableStoreError(new DOMException("signal timed out", "TimeoutError"), "无法读取店铺"),
    "请求超时，请检查网络后重试");
  assert.equal(readableStoreError(new Error("已有中文错误"), "fallback"), "已有中文错误");
  assert.equal(readableStoreError(null, "fallback"), "fallback");
});

test("cached reads and watchers are safe during synchronous session abort", () => {
  const oldLifetime = authLifetime();
  let observed = false;
  oldLifetime.addEventListener("abort", () => {
    observed = true;
    assert.equal(cachedStores(), undefined);
    assert.equal(cachedListingPage("/api/v1/stores/test/published-listings"), undefined);
    assert.doesNotThrow(() => watchStoreList(() => {})());
  }, { once: true });
  rejectSession();
  assert.equal(observed, true);
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
  const countHits = () => JSON.parse(diagnosticsJSON())
    .filter((event) => event.operation === "listings_get" && event.cache_hit).length;
  const before = countHits();
  assert.deepEqual(cachedListingPage(mlm), page);
  assert.deepEqual(cachedListingPage(mlm), page);
  assert.equal(countHits(), before, "pure cache reads must not record during render");
  recordCacheHit(mlm);
  assert.equal(countHits(), before + 1);
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


test("login verifies each eligible store once and preserves a failed verification warning", async () => {
  const initial = [
    { id: "one", status: "connected", marketplaces_status: "verified" },
    { id: "two", status: "token_expired", marketplaces_status: "verified" },
  ];
  const fresh = initial.map((item) => ({ ...item, status: "connected" }));
  const reads = [];
  const checks = [];
  const result = await checkStoresOnLogin(
    async (force) => { reads.push(force); return reads.length === 1 ? initial : fresh; },
    async (id, manual) => {
      checks.push([id, manual]);
      if (id === "two") throw new Error("network error");
    },
  );
  assert.deepEqual(reads, [true, true], "only the initial and final store GETs are needed");
  assert.deepEqual(checks, [["one", false], ["two", false]], "one automatic POST per eligible store");
  assert.equal(result.health.status, "attention", "an ambiguous check must not show connected");
  assert.deepEqual([...result.failedStoreIds], ["two"]);
  assert.equal(storeHealthFromList(fresh, result.failedStoreIds).status, "attention",
    "later plain GETs must not override a failed connection check");
  result.failedStoreIds.delete("two"); // confirmed manual check for this store
  assert.equal(storeHealthFromList(fresh, result.failedStoreIds).status, "connected",
    "only a successful recheck should clear the warning");
});

test("login leaves terminal/absent stores alone instead of attempting OAuth rotation", async () => {
  const states = [
    { id: "reauth", status: "reauth_required", marketplaces_status: "verified" },
    { id: "uncertain", status: "refresh_unknown", marketplaces_status: "verified" },
    { id: "disconnected", status: "disconnected", marketplaces_status: "verified" },
  ];
  let reads = 0;
  let postCalls = 0;
  const result = await checkStoresOnLogin(async () => { reads++; return states; }, async () => { postCalls++; });
  assert.equal(reads, 1);
  assert.equal(postCalls, 0);
  assert.equal(result.health.status, "attention");
  assert.equal(result.failedStoreIds.size, 0);
  const empty = await checkStoresOnLogin(async () => [], async () => { postCalls++; });
  assert.equal(empty.health.status, "empty");
  assert.equal(postCalls, 0);
});

test("disconnected-only stores show an empty state, not a connection error", async () => {
  const disconnected = { id: "inactive", status: "disconnected", marketplaces_status: "unverified" };
  const connected = { id: "active", status: "connected", marketplaces_status: "verified" };
  const reauth = { id: "reauth", status: "reauth_required", marketplaces_status: "verified" };
  assert.deepEqual(storeHealthFromList([disconnected]), {
    status: "empty", message: "所有店铺已断开，可在设置中重新授权",
  });
  assert.equal(storeHealthFromList([connected, disconnected]).status, "connected",
    "a voluntarily disconnected store must not make a healthy account orange");
  assert.equal(storeHealthFromList([disconnected, reauth]).status, "attention",
    "a still-active store requiring reauthorization remains actionable");
  assert.equal(storeHealthFromList([disconnected], new Set(["inactive"])).status, "empty",
    "failed checks on subsequently disconnected stores are not active failures");
  assert.equal(storeHealthFromList([connected], new Set(["__initial_read_failed__"])).status, "attention",
    "an initial read failure still requires confirmation if active stores exist");
});

test("manual verification recovery keeps unverified peer stores flagged", () => {
  const stores = [
    { id: "confirmed", status: "connected", marketplaces_status: "verified" },
    { id: "other", status: "connected", marketplaces_status: "verified" },
    { id: "deliberate-disconnect", status: "disconnected", marketplaces_status: "verified" },
  ];
  const initialFailure = new Set(["__initial_read_failed__", "deliberate-disconnect"]);
  const afterFirst = resolveManualStoreCheck(initialFailure, stores, "confirmed");
  assert.deepEqual([...initialFailure].sort(), ["__initial_read_failed__", "deliberate-disconnect"],
    "input should not be mutated");
  assert.deepEqual([...afterFirst], ["other"], "other active store must remain unverified");
  assert.equal(storeHealthFromList(stores, afterFirst).status, "attention");
  const afterSecond = resolveManualStoreCheck(afterFirst, stores, "other");
  assert.deepEqual([...afterSecond], []);
  assert.equal(storeHealthFromList(stores, afterSecond).status, "connected");

  const failedOne = new Set(["confirmed", "other"]);
  const recoveredOne = resolveManualStoreCheck(failedOne, stores, "confirmed");
  assert.deepEqual([...recoveredOne], ["other"], "manual recovery must only clear the checked store");
});
