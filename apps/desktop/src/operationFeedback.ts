// Business state only. The renderer owns display duration and animation.
export type FeedbackKind = "success" | "error" | "warning" | "info" | "loading";
export type FeedbackView = {
  id: number;
  kind: FeedbackKind;
  title: string;
  description?: string;
  count?: number;
  progress?: number;
};
type Outcome = "success" | "failed" | "skipped";
type Identity = { scope: string; key: string };
type BatchCounts = { success: number; failed: number; skipped: number };

// One visible owner; older operations still finish in their business pages,
// but cannot overwrite newer global feedback. No business retries are implied.
export function createFeedback(render: (view: FeedbackView | null) => void, now = Date.now) {
  let sequence = 0;
  let active: { view: FeedbackView; identity: Identity; at: number; single: boolean } | null = null;
  function start(identity: Identity, view: Omit<FeedbackView, "id">, single = false) {
    const id = ++sequence;
    active = { view: { ...view, id }, identity, at: now(), single };
    render(active.view);
    return id;
  }
  function update(id: number, view: Omit<FeedbackView, "id">) {
    if (active?.view.id !== id) return;
    active.view = { ...view, id };
    render(active.view);
  }
  function notify(identity: Identity, kind: FeedbackKind, title: string) {
    if (active?.single && active.identity.scope === identity.scope &&
        active.identity.key === identity.key && active.view.kind === kind &&
        active.view.title === title && now() - active.at <= 1500) {
      active.at = now();
      update(active.view.id, { kind, title, count: (active.view.count ?? 1) + 1 });
      return;
    }
    start(identity, { kind, title }, true);
  }
  function task(identity: Identity, title: string) {
    const id = start(identity, { kind: "loading", title });
    let finished = false;
    return {
      update(title: string, progress?: number, description?: string) {
        if (finished) return;
        if (progress !== undefined && (!Number.isFinite(progress) || progress < 0 || progress > 1)) return;
        update(id, { kind: "loading", title, progress, description });
      },
      finish(kind: Exclude<FeedbackKind, "loading">, title: string, description?: string) {
        if (finished) return;
        finished = true;
        update(id, { kind, title, description });
      },
    };
  }
  function batch(identity: Identity, title: string, total: number) {
    if (!Number.isSafeInteger(total) || total <= 0) throw new RangeError("invalid feedback batch size");
    const id = start(identity, { kind: "loading", title, description: "0/" + total, progress: 0 });
    const seen = new Set<string>();
    const counts: BatchCounts = { success: 0, failed: 0, skipped: 0 };
    let finished = false;
    const description = () => "成功 " + counts.success + " · 失败 " + counts.failed +
      " · 跳过 " + counts.skipped + " · 已处理 " + seen.size + "/" + total;
    return {
      record(itemId: string, outcome: Outcome) {
        if (finished || !itemId || seen.has(itemId) || seen.size >= total) return;
        seen.add(itemId);
        counts[outcome]++;
        update(id, { kind: "loading", title, description: description(), progress: seen.size / total });
      },
      finish(finalTitle = title) {
        if (finished) return;
        finished = true;
        const kind = seen.size < total || counts.skipped > 0 || counts.failed > 0
          ? (counts.failed === total ? "error" : "warning") : "success";
        update(id, { kind, title: finalTitle, description: description() });
      },
    };
  }
  function clear() {
    active = null;
    render(null);
  }
  return { notify, task, batch, clear };
}
