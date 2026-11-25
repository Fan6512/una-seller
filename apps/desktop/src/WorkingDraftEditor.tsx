import { useEffect, useRef, useState } from "react";
import ListingFieldsEditor from "./ListingFieldsEditor";
import { initialListingFields, listingTextCount, preferredTitle, validateListingFields, type ListingFields, type ListingSource } from "./listingFields";
import { getWorkingDraft, isDraftConflict, isDraftConverted, isDraftRejected, isVariantUpgradeRequired, saveWorkingDraft, type WorkingDraftInput } from "./collections";

type DraftFields = Pick<WorkingDraftInput, "title" | "description" | "notes"> & {listing_fields: ListingFields};
type LoadState = "loading" | "ready" | "error";
const charCount = (value: string) => Array.from(value).length;

export default function WorkingDraftEditor({ collectionId, sourceTitle, sourceDescription, sourceProduct, onDirtyChange, locked }: {
  collectionId: string; sourceTitle: string; sourceDescription: string | null;
  sourceProduct: ListingSource; locked: boolean; onDirtyChange?: (dirty: boolean) => void;
}) {
  const [fields, setFields] = useState<DraftFields>({
    title: sourceTitle, description: sourceDescription ?? "", notes: "",
    listing_fields: initialListingFields(null,sourceProduct,sourceTitle),
  });
  const [saved, setSaved] = useState<DraftFields | null>(null);
  const [touched, setTouched] = useState(false);
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
        ? { title: draft.title, description: draft.description, notes: draft.notes,
            listing_fields: initialListingFields(draft.listing_fields,sourceProduct,draft.title) }
        : { title: sourceTitle, description: sourceDescription ?? "", notes: "",
            listing_fields: initialListingFields(null,sourceProduct,sourceTitle) };
      setFields(next);
      setSaved(draft ? next : null);
      setTouched(false);
      setRevision(draft?.revision ?? 0);
      setLoadState("ready");
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) {
        setLoadState("error");
        setError((e instanceof Error ? e.message : "无法读取工作稿") + "。当前已禁用保存。");
      }
    });
    return () => controller.abort();
  }, [collectionId, sourceTitle, sourceDescription, sourceProduct, reload]);

  // The GET effect owns its request; this cleanup also cancels any active PUT.
  useEffect(() => () => active.current?.abort(), []);

  const dirty = !saved || fields.title !== saved.title ||
    fields.description !== saved.description || fields.notes !== saved.notes ||
    JSON.stringify(fields.listing_fields) !== JSON.stringify(saved.listing_fields);
  // Only edited text is unsaved: an untouched source copy is not an obstacle.
  useEffect(() => {
    onDirtyChange?.(touched && dirty);
    return () => onDirtyChange?.(false);
  }, [touched,dirty,onDirtyChange]);
  const editable = !locked && loadState === "ready" && !saving;
  const titleCount = listingTextCount(fields.title);
  const legacyDescriptionOversize = charCount(fields.description) > 20000;
  const notesCount = charCount(fields.notes);
  const validation = legacyDescriptionOversize ? "历史描述超过 20000 字符，请先处理历史数据"
    : notesCount > 5000 ? "整理备注不能超过 5000 字符"
    : titleCount > 500 ? "商品标题不能超过 500 字符"
    : validateListingFields(fields.listing_fields, saved?.listing_fields.titles);
  const valid = !validation && fields.title.trim().length > 0;

  const change = (key: keyof DraftFields, value: string) => {
    setFields((current) => ({ ...current, [key]: value }));
    setTouched(true);
    setNotice("");
  };

  const refreshDraft = () => {
    if (touched && !window.confirm(
      "重新加载将丢弃此窗口尚未保存的修改。请先复制需要保留的内容，确定继续吗？",
    )) return;
    setReload((n) => n + 1);
  };

  const save = async () => {
    if (!valid || !editable || !dirty) return;
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
      const next = { title: draft.title, description: draft.description, notes: draft.notes,
        listing_fields: initialListingFields(draft.listing_fields,sourceProduct,draft.title) };
      setFields(next);
      setSaved(next);
      setTouched(false);
      setRevision(draft.revision);
      setNotice("已保存到云端整理工作稿，关闭窗口后可恢复。");
    } catch (e) {
      if (!controller.signal.aborted) {
        // A timed-out PUT may have committed. Never blindly retry the same revision.
        if (!isDraftRejected(e)) setLoadState("error");
        setError(isDraftConverted(e)
          ? "商品已在另一窗口转入产品箱，原工作稿停止编辑。请先复制当前修改，再到「工作台 → 产品箱」编辑商品。"
          : isDraftConflict(e)
          ? "工作稿已被其他窗口或上一次提交修改（409），本次没有覆盖服务器内容。请复制需要保留的文本，再重新加载。"
          : isVariantUpgradeRequired(e)
            ? "旧版变体编辑无法安全保存，请先复制未保存的修改，再升级支持新版变体字段的客户端。本次未写入。"
          : isDraftRejected(e)
            ? "服务器拒绝保存（请求格式或长度不符合要求），内容没有写入。请修改后重新保存。"
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
    <ListingFieldsEditor value={fields.listing_fields} disabled={!editable}
      aiProduct={{source_product_id:collectionId,expected_revision:revision}} aiSaved={saved?.listing_fields}
      onChange={listing_fields=>{
        setFields(v=>({...v,title:preferredTitle(listing_fields.titles),listing_fields}));
        setTouched(true);setNotice("");
      }} />
    {legacyDescriptionOversize&&<details className="listing-fields-note">
      <summary>历史描述过长 · 查看及处理（非刊登描述）</summary>
      <textarea rows={4} readOnly value={fields.description} aria-label="超长历史描述，只读"/>
      <p>请先复制需要保留的历史文本。正常历史描述不会显示或被删除。</p>
      <button type="button" disabled={!editable} onClick={()=>{
        if(window.confirm("确认清空超长历史描述？正式多语言商品描述不受影响。"))change("description","");
      }}>清空超长历史描述</button>
    </details>}
    <label>供应商 / 规格 / 包装等整理备注（仅内部使用，不直接刊登）
      <textarea rows={3} value={fields.notes} readOnly={!editable}
        onChange={(e) => change("notes", e.target.value)} />
      <small>{notesCount} / 5000 字符</small>
    </label>
    {loadState === "ready" && validation && <p className="working-draft-error" role="alert">{validation}</p>}
    {error && <p className="working-draft-error" role="alert">{error}</p>}
    {notice && <p className="working-draft-success" role="status">{notice}</p>}
    {loadState === "error" && <button type="button" className="refresh-button"
      disabled={locked} onClick={refreshDraft}>重新加载工作稿</button>}
    <button type="button" className="primary-button" onClick={() => void save()}
      disabled={!valid || !editable || !dirty}>{saving ? "保存中…" : "保存整理工作稿"}</button>
  </section>;
}
