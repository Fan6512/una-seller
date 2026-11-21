import { useEffect, useRef, useState } from "react";
import { getWorkingDraft, isDraftConflict, saveWorkingDraft, type WorkingDraftInput } from "./collections";

type DraftFields = Pick<WorkingDraftInput, "title" | "description" | "notes">;
type LoadState = "loading" | "ready" | "error";
const charCount = (value: string) => Array.from(value).length;

export default function WorkingDraftEditor({ collectionId, sourceTitle, sourceDescription }: {
  collectionId: string; sourceTitle: string; sourceDescription: string | null;
}) {
  const [fields, setFields] = useState<DraftFields>({
    title: sourceTitle, description: sourceDescription ?? "", notes: "",
  });
  const [saved, setSaved] = useState<DraftFields | null>(null);
  const [revision, setRevision] = useState(0);
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [reload, setReload] = useState(0);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const active = useRef<AbortController | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    active.current = controller;
    setLoadState("loading");
    setError("");
    setNotice("");
    void getWorkingDraft(collectionId, controller.signal).then(({ draft }) => {
      if (controller.signal.aborted) return;
      const next = draft
        ? { title: draft.title, description: draft.description, notes: draft.notes }
        : { title: sourceTitle, description: sourceDescription ?? "", notes: "" };
      setFields(next);
      setSaved(draft ? next : null);
      setRevision(draft?.revision ?? 0);
      setLoadState("ready");
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) {
        setLoadState("error");
        setError((e instanceof Error ? e.message : "无法读取工作稿") + "。当前已禁用保存。");
      }
    });
    return () => controller.abort();
  }, [collectionId, sourceTitle, sourceDescription, reload]);

  // The GET effect owns its request; this cleanup also cancels any active PUT.
  useEffect(() => () => active.current?.abort(), []);

  const dirty = !saved || fields.title !== saved.title ||
    fields.description !== saved.description || fields.notes !== saved.notes;
  const titleCount = charCount(fields.title);
  const descriptionCount = charCount(fields.description);
  const notesCount = charCount(fields.notes);
  const valid = fields.title.trim().length > 0 && titleCount <= 500 &&
    descriptionCount <= 20000 && notesCount <= 5000;

  const change = (key: keyof DraftFields, value: string) => {
    setFields((current) => ({ ...current, [key]: value }));
    setNotice("");
  };

  const refreshDraft = () => {
    if (saved && dirty && !window.confirm(
      "重新加载将丢弃此窗口尚未保存的修改。请先复制需要保留的内容，确定继续吗？",
    )) return;
    setReload((n) => n + 1);
  };

  const save = async () => {
    if (!valid || loadState !== "ready" || saving || !dirty) return;
    const controller = new AbortController();
    active.current?.abort();
    active.current = controller;
    setSaving(true);
    setError("");
    setNotice("");
    try {
      const { draft } = await saveWorkingDraft(collectionId, {
        ...fields, expected_revision: revision,
      }, controller.signal);
      if (controller.signal.aborted) return;
      const next = { title: draft.title, description: draft.description, notes: draft.notes };
      setFields(next);
      setSaved(next);
      setRevision(draft.revision);
      setNotice("已保存到云端整理工作稿，关闭窗口后可恢复。");
    } catch (e) {
      if (!controller.signal.aborted) {
        // A timed-out PUT may have committed. Never blindly retry the same revision.
        setLoadState("error");
        setError(isDraftConflict(e)
          ? "工作稿已被另一窗口修改（409），本次没有覆盖服务器内容。请复制需要保留的文本，再重新加载。"
          : (e instanceof Error ? e.message : "保存失败") + "。保存状态未确认，请重新加载后再试。");
      }
    } finally {
      if (!controller.signal.aborted) setSaving(false);
    }
  };

  return <section className="detail-section working-draft">
    <div className="working-draft-heading">
      <h3>整理工作稿</h3>
      <span>{loadState === "loading" ? "读取中…" : loadState === "error" ? "状态未知 · 保存禁用"
        : saved ? (dirty ? "有未保存修改" : "已保存") : "尚未保存"}</span>
    </div>
    <p>这里是转产品库前的暂存稿，原始采集标题、描述和图片不会被覆盖。</p>
    <label>商品标题
      <input value={fields.title} disabled={loadState !== "ready" || saving}
        onChange={(e) => change("title", e.target.value)} />
      <small>{titleCount} / 500 字符</small>
    </label>
    <label>整理后的商品描述
      <textarea rows={5} value={fields.description} disabled={loadState !== "ready" || saving}
        onChange={(e) => change("description", e.target.value)} />
      <small>{descriptionCount} / 20000 字符</small>
    </label>
    <label>供应商 / 规格 / 包装等整理备注（仅内部使用，不直接刊登）
      <textarea rows={3} value={fields.notes} disabled={loadState !== "ready" || saving}
        onChange={(e) => change("notes", e.target.value)} />
      <small>{notesCount} / 5000 字符</small>
    </label>
    {loadState === "ready" && !valid && <p className="working-draft-error" role="alert">
      请填写商品标题，并将标题、描述、备注分别控制在 500 / 20000 / 5000 字符以内。
    </p>}
    {error && <p className="working-draft-error" role="alert">{error}</p>}
    {notice && <p className="working-draft-success" role="status">{notice}</p>}
    {loadState === "error" && <button type="button" className="refresh-button"
      onClick={refreshDraft}>重新加载工作稿</button>}
    <button type="button" className="primary-button" onClick={() => void save()}
      disabled={!valid || loadState !== "ready" || saving || !dirty}>{saving ? "保存中…" : "保存整理工作稿"}</button>
    <p className="working-draft-hint">下一阶段才会将工作稿复制为独立产品库草稿；当前不会自动刊登。</p>
  </section>;
}
