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

test("deferred copy results aggregate and cannot replace a newer task or session", () => {
  let view;
  const feedback = createFeedback(v => { view = v; });
  const copy = {scope:"store-a", key:"copy"};
  const first = feedback.notify(copy);
  const second = feedback.notify(copy);
  first("success", "已复制");
  second("success", "已复制");
  assert.equal(view.count, 2);
  const late = feedback.notify(copy);
  feedback.task({scope:"store-b", key:"sync"}, "同步中");
  late("success", "已复制");
  assert.equal(view.title, "同步中");
  const oldSession = feedback.notify(copy);
  feedback.clear();
  feedback.notify(copy);
  oldSession("success", "已复制");
  assert.equal(view, null);
});

test("renderer failure is isolated from business callers and state stays recoverable", () => {
  let broken = true;
  let view;
  const feedback = createFeedback(v => {
    if (broken) throw new Error("renderer unavailable");
    view = v;
  });
  const identity = {scope:"store-a", key:"sync"};
  const task = feedback.task(identity, "准备同步");
  task.update("同步中", 0.5);
  task.finish("success", "同步完成");
  const copy = feedback.notify({scope:"store-a", key:"copy"});
  copy("success", "已复制");
  feedback.notify({scope:"store-a", key:"copy"}, "success", "已复制");
  const batch = feedback.batch({scope:"store-a", key:"images"}, "补图", 2);
  batch.record("one", "success");
  broken = false;
  batch.record("two", "failed");
  assert.equal(view.progress, 1);
  batch.finish("补图结束");
  assert.equal(view.kind, "warning");
  assert.match(view.description, /成功 1 · 失败 1 · 跳过 0/);
  const stale = feedback.task(identity, "旧任务");
  const staleCopy = feedback.notify({scope:"store-a", key:"copy"});
  broken = true;
  feedback.clear();
  broken = false;
  feedback.notify({scope:"store-b", key:"copy"}, "success", "新会话");
  stale.finish("error", "旧任务失败");
  staleCopy("success", "旧账号复制");
  assert.equal(view.title, "新会话");
  feedback.clear();
  assert.equal(view, null);
});
