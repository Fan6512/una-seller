import assert from "node:assert/strict";
import test from "node:test";
import {
  approvalBlockReason,
  countApprovalOutcomes,
  emptyProductPageFallback,
  productSaveNeedsReadback,
} from "../src/workbenchRules.ts";

test("batch approval distinguishes actual transfers, existing products and failures", () => {
  assert.deepEqual(countApprovalOutcomes([
    "created","already_approved","created","failed","already_approved",
  ]), { created:2, alreadyApproved:2, failed:1 });
  assert.deepEqual(countApprovalOutcomes(["already_approved"]),
    { created:0, alreadyApproved:1, failed:0 },
    "retrying a transferred source is not a new success");
});

test("deterministic write rejection does not lock the product editor", () => {
  for (const code of [400,404,409,415,422]) {
    assert.equal(productSaveNeedsReadback(code), false, String(code));
  }
  for (const code of [undefined,202,401,500,504]) {
    assert.equal(productSaveNeedsReadback(code), true, String(code));
  }
});

test("empty last page returns to the previous page without trapping navigation", () => {
  assert.equal(emptyProductPageFallback(48,0,24),24);
  assert.equal(emptyProductPageFallback(24,0,24),0);
  assert.equal(emptyProductPageFallback(0,0,24),null);
  assert.equal(emptyProductPageFallback(48,2,24),null);
});


test("approve button gives actionable feedback without bypassing selection and save", () => {
  assert.match(approvalBlockReason(0,false),/勾选/);
  assert.match(approvalBlockReason(0,true),/勾选/);
  assert.match(approvalBlockReason(1,true),/保存/);
  assert.equal(approvalBlockReason(1,false),null);
  assert.equal(approvalBlockReason(9,false),null);
});
