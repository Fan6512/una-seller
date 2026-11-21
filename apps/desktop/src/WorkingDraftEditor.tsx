import { useEffect, useRef, useState } from "react";
import { getWorkingDraft, saveWorkingDraft, type WorkingDraftInput } from "./collections";

export default function WorkingDraftEditor({ collectionId, sourceTitle, sourceDescription }: {
  collectionId: string; sourceTitle: string; sourceDescription: string | null;
}) {
  const [fields, setFields] = useState<WorkingDraftInput>({
    title: sourceTitle, description: sourceDescription ?? "", notes: "",
  });
  const [saved, setSaved] = useState<WorkingDraftInput | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const active = useRef<AbortController | null>(null);
  useEffect(() => {
    const controller = new AbortController();
    active.current = controller;
    setLoading(true);
    setError("");
    setNotice("");
    void getWorkingDraft(collectionId, controller.signal).then(({ draft }) => {
      if (controller.signal.aborted) return;
      const next = draft
        ? { title: draft.title, description: draft.description, notes: draft.notes }
        : { title: sourceTitle, description: sourceDescription ?? "", notes: "" };
      setFields(next);
      setSaved(draft ? next : null);
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "整理工作稿读取失败");
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, [collectionId, sourceTitle, sourceDescription]);

  const dirty = !saved || fields.title !== saved.title ||
    fields.description !== saved.description || fields.notes !== saved.notes;
  const valid = fields.title.trim().length > 0 && fields.title.length <= 500 &&
    fields.description.length <= 20000 && fields.notes.length <= 5000;

  const save = async () => {
    if (!valid || loading || saving) return;
    const controller = new AbortController();
    active.current?.abort();
    active.current = controller;
    setSaving(true);
    setError("");
    setNotice("");
    try {
      const { draft } = await saveWorkingDraft(collectionId, fields, controller.signal);
      if (controller.signal.aborted) return;
      const next = { title: draft.title, description: draft.description, notes: draft.notes };
      setFields(next);
      setSaved(next);
      setNotice("已保存到云端整理工作稿，关闭窗口后可恢复。");
    } catch (e) {
      if (!controller.signal.aborted) setError(e instanceof Error ? e.message : "保存失败");
    } finally {
      if (!controller.signal.aborted) setSaving(false);
    }
  };

  return <section className="detail-section working-draft">
    <div className="working-draft-heading">
      <h3>整理工作稿</h3>
      <span>{loading ? "读取中…" : saved ? (dirty ? "有未保存修改" : "已保存") : "尚未保存"}</span>
    </div>
    <p>这里编辑的是独立工作稿，原始采集标题、描述和图片不会被覆盖。</p>
    <label>商品标题
      <input value={fields.title} maxLength={500} disabled={loading || saving}
        onChange={(e) => setFields({ ...fields, title: e.target.value })} />
      <small>{fields.title.length} / 500</small>
    </label>
    <label>整理后的商品描述
      <textarea rows={5} value={fields.description} maxLength={20000} disabled={loading || saving}
        onChange={(e) => setFields({ ...fields, description: e.target.value })} />
      <small>{fields.description.length} / 20000</small>
    </label>
    <label>供应商 / 规格 / 包装等整理备注
      <textarea rows={3} value={fields.notes} maxLength={5000} disabled={loading || saving}
        onChange={(e) => setFields({ ...fields, notes: e.target.value })} />
      <small>{fields.notes.length} / 5000</small>
    </label>
    {error && <p className="working-draft-error" role="alert">{error}</p>}
    {notice && <p className="working-draft-success" role="status">{notice}</p>}
    <button type="button" className="primary-button" onClick={() => void save()}
      disabled={!valid || loading || saving || !dirty}>{saving ? "保存中…" : "保存整理工作稿"}</button>
    <p className="working-draft-hint">下一阶段才会将工作稿转为独立产品库草稿；当前不会自动刊登。</p>
  </section>;
}
