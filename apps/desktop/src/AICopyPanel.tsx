import { useEffect, useRef, useState } from "react";
import { generateAICopy, type AICopyCandidate, type AICopyProduct } from "./collections";
import { isPlainListingText, listingTextCount, localizedTitleLimit, preferredTitle,
  type ListingFields, type TitleLanguage } from "./listingFields";

const languages = { en: "英语 EN", es: "西语 MX", pt: "葡语 BR" };
type CopyField = "titles" | "descriptions";

type Props = {
  product: AICopyProduct;
  value: ListingFields;
  disabled: boolean;
  onAdopt: (field: CopyField, language: TitleLanguage, text: string) => void;
};

export default function AICopyPanel({ product, value, disabled, onAdopt }: Props) {
  const [candidate, setCandidate] = useState<AICopyCandidate | null>(null);
  const [language, setLanguage] = useState<TitleLanguage>("en");
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const active = useRef<AbortController | null>(null);
  useEffect(() => () => active.current?.abort(), []);
  useEffect(() => {
    if (!disabled || !active.current) return;
    active.current.abort();
    active.current = null;
    setGenerating(false);
  }, [disabled]);

  const generate = async () => {
    if (active.current) return;
    const controller = new AbortController();
    active.current = controller;
    setGenerating(true);
    setError("");
    setMessage("");
    try {
      const result = await generateAICopy(product, value, controller.signal);
      if (controller.signal.aborted) return;
      const descriptions = { ...result.descriptions };
      for (const lang of Object.keys(languages) as TitleLanguage[]) {
        descriptions[lang] = result.selling_points[lang].join("\n") + "\n\n" + descriptions[lang];
      }
      setCandidate({ ...result, descriptions });
    } catch (cause) {
      if (!controller.signal.aborted) {
        setError(cause instanceof Error ? cause.message : "AI 文案生成失败，请重试");
      }
    } finally {
      if (active.current === controller) {
        active.current = null;
        setGenerating(false);
      }
    }
  };

  const cancel = () => {
    active.current?.abort();
    active.current = null;
    setGenerating(false);
    setMessage("已取消生成");
  };

  const copy = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setMessage("已复制候选" + label);
    } catch {
      setMessage("复制未完成，请手动选择候选文本");
    }
  };

  const renderField = (field: CopyField, label: string) => {
    if (!candidate) return null;
    const text = candidate[field][language];
    const current = value[field]?.[language] ?? "";
    const count = listingTextCount(text);
    const title = field === "titles";
    const limit = title ? (language === "en" ? 500 : localizedTitleLimit) : 20000;
    const formatValid = isPlainListingText(text) && !(title && /[\r\n\t]/.test(text));
    const eligible = text.trim() !== "" && formatValid && count <= limit;
    const rows = title ? 3 : 8;
    return <div>
      <div className="listing-field-grid">
        <label className="listing-field">当前{label}
          <textarea rows={rows} value={current} readOnly />
        </label>
        <label className="listing-field">候选{label}
          <textarea rows={rows} value={text} disabled={disabled || generating}
            onChange={e => setCandidate({ ...candidate, [field]: { ...candidate[field], [language]: e.target.value } })} />
        </label>
      </div>
      <div className="localized-title-actions">
        <small className={"localized-title-count " + (count > limit || !formatValid ? "title-over-limit" : "")}>
          {count} / {limit} 字符{title && language === "en" ? " · 建议 45–55 字符" : ""}
          {count > limit ? " · 请精简后采纳" : !formatValid ? " · 请改为纯文本" : ""}
        </small>
        <div className="localized-title-actions">
          <button type="button" disabled={!text} onClick={() => void copy(text, label)}>复制候选{label}</button>
          <button type="button" disabled={disabled || generating || !eligible || text === current}
            onClick={() => {
              onAdopt(field, language, text.normalize("NFC"));
              setMessage("");
            }}>采纳{label}</button>
        </div>
      </div>
    </div>;
  };

  return <fieldset className="listing-editor-panel">
    <legend>AI 文案候选</legend>
    <div className="localized-title-actions">
      <p className="listing-fields-note">使用当前商品标题、规格和卖点生成。核对后可逐项采纳，随商品一起保存。</p>
      <div className="localized-title-actions">
        <button type="button" disabled={disabled || generating || !preferredTitle(value.titles)}
          onClick={() => void generate()}>{candidate ? "重新生成" : "生成 AI 文案"}</button>
        {generating && <button type="button" onClick={cancel}>取消生成</button>}
      </div>
    </div>
    {generating && <p className="listing-fields-note" role="status">正在生成三语文案…</p>}
    {error && <p className="listing-validation" role="alert">{error}</p>}
    {message && <p className="listing-fields-note" role="status">{message}</p>}
    {candidate && <>
      {candidate.missing_facts.length > 0 && <div>
        <strong>待补充的商品信息</strong>
        <ul>{candidate.missing_facts.map(fact => <li key={fact}>{fact}</li>)}</ul>
      </div>}
      <div className="listing-language-tabs" role="group" aria-label="候选文案语言">
        {(Object.keys(languages) as TitleLanguage[]).map(lang => <button key={lang} type="button"
          aria-pressed={language === lang} onClick={() => setLanguage(lang)}>{languages[lang]}</button>)}
      </div>
      {renderField("titles", "标题")}
      {renderField("descriptions", "描述（含卖点）")}
    </>}
  </fieldset>;
}
