import assert from "node:assert/strict";
import { test } from "node:test";
import { createFeedback } from "../src/operationFeedback.ts";

test("same short-lived notification aggregates, distinct store or action replaces", () => {
  let view, clock = 0;
  const feedback = createFeedback(v => { view = v; }, () => clock);
  const a = { scope: "store-a", key: "copy" };
  feedback.notify(a, "success", "已复制");
  const first = view.id;
  feedback.notify(a, "success", "已复制");
  assert.equal(view.id, first);
  assert.equal(view.count, 2);
  clock = 2000;
  feedback.notify(a, "success", "已复制");
  assert.notEqual(view.id, first);
  feedback.notify({ ...a, scope: "store-b" }, "success", "已复制");
  assert.equal(view.count, undefined);
});

test("task lifecycle cannot overwrite a newer notification or resurrect after clear", () => {
  let view;
  const feedback = createFeedback(v => { view = v; });
  const a = feedback.task({scope:"a", key:"sync"}, "准备同步");
  a.update("同步中");
  assert.equal(view.progress, undefined);
  a.update("同步中", 0.5);
  assert.equal(view.progress, 0.5);
  const b = feedback.task({scope:"b", key:"save"}, "保存中");
  const id = view.id;
  a.finish("error", "旧任务失败");
  assert.equal(view.id, id);
  b.finish("success", "已保存");
  b.update("late", 1);
  assert.equal(view.title, "已保存");
  feedback.clear();
  assert.equal(view, null);
  b.finish("error", "late");
  assert.equal(view, null);
});

test("batch counts deduplicate and incomplete work is not reported as successful", () => {
  let view;
  const feedback = createFeedback(v => { view = v; });
  const batch = feedback.batch({scope:"a", key:"images"}, "图片处理", 5);
  const id = view.id;
  batch.record("one","success");
  batch.record("one","failed");
  batch.record("two","failed");
  batch.record("three","skipped");
  assert.equal(view.id, id);
  assert.equal(view.progress, 0.6);
  batch.finish("已停止");
  assert.equal(view.kind, "warning");
  assert.match(view.description, /成功 1 · 失败 1 · 跳过 1 · 已处理 3\/5/);
  assert.equal(view.progress, undefined);
  batch.record("four","success");
  assert.equal(view.title,"已停止");
});
