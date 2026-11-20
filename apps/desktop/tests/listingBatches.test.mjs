import assert from "node:assert/strict";
import { registerHooks } from "node:module";
import test from "node:test";
import { runListingBatches } from "../src/listingBatches.ts";

const storesURL = new URL("../src/stores.ts", import.meta.url).href;
const hooks = registerHooks({
  resolve(specifier, context, nextResolve) {
    if (context.parentURL === storesURL) {
      if (specifier === "./auth") {
        return { url: "data:text/javascript,export const getAccessToken = async () => 'fixture'; export const rejectSession = () => {};", shortCircuit: true };
      }
      if (specifier === "./collections") {
        return { url: "data:text/javascript,export const apiBase = 'https://fixture.invalid';", shortCircuit: true };
      }
    }
    return nextResolve(specifier, context);
  },
});
const { storeRequest } = await import(storesURL);
hooks.deregister();

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
