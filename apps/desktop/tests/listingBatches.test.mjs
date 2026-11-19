import assert from "node:assert/strict";
import test from "node:test";
import { runListingBatches } from "../src/listingBatches.ts";

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
