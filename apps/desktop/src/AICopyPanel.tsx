import { useEffect, useRef, useState } from "react";
import { generateAICopy, type AICopyProduct, type AICopyResults } from "./collections";
import { isPlainListingText, listingTextCount, localizedTitleLimit,
  type ListingFields, type LocalizedTitles, type TitleLanguage } from "./listingFields";

const languages = { en: "英语 EN", es: "西语 MX", pt: "葡语 BR" };
const descriptionLabels = { en: "英文详情", es: "西语详情", pt: "葡语详情" };
type CopyField = "titles" | "descriptions";
type Operation = keyof AICopyResults;

type Props = {
  product: AICopyProduct;
  value: ListingFields;
  saved?: Partial<ListingFields>;
  disabled: boolean;
  onAdopt: (field: CopyField, language: TitleLanguage, text: string, sourceDigest?: string) => void;
};

export default function AICopyPanel({ product, value, saved, disabled, onAdopt }: Props) {
  const [titles, setTitles] = useState<LocalizedTitles | null>(null);
  const [english, setEnglish] = useState<AICopyResults["english"] | null>(null);
  const [regional, setRegional] = useState<(AICopyResults["translate"] & { revision: number }) | null>(null);
  const [language, setLanguage] = useState<TitleLanguage>("en");
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const active = useRef<AbortController | null>(null);
  const savedEnglish = saved?.descriptions?.en?.normalize("NFC") ?? "";
  const englishDirty = (value.descriptions?.en?.normalize("NFC") ?? "") !== savedEnglish;
  const savedDigest = saved?.description_translation?.en_digest;
  const regionalIsStale = !!regional && (englishDirty || !savedEnglish.trim() ||
    (savedDigest ? regional.source_digest !== savedDigest : regional.revision !== product.expected_revision));

  useEffect(() => () => active.current?.abort(), []);
  useEffect(() => {
    if (!disabled || !active.current) return;
    active.current.abort();
    active.current = null;
    setGenerating(false);
  }, [disabled]);

  const generate = async (operation: Operation) => {
    if (active.current) return;
    const controller = new AbortController();
    active.current = controller;
    setGenerating(true);
    setError("");
    setMessage("");
    try {
      switch (operation) {
        case "titles": {
          const result = await generateAICopy(product, operation, controller.signal);
          if (!controller.signal.aborted) setTitles(result.titles);
          break;
        }
        case "english": {
          const result = await generateAICopy(product, operation, controller.signal);
          if (!controller.signal.aborted) {
            setEnglish({ ...result, description: result.selling_points.join("\n") + "\n\n" + result.description });
            setLanguage("en");
          }
          break;
        }
        case "translate": {
          const result = await generateAICopy(product, operation, controller.signal);
          if (!controller.signal.aborted) {
            setRegional({ ...result, revision: product.expected_revision });
            setLanguage(current => current === "en" ? "es" : current);
          }
          break;
        }
      }
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

  const renderField = (field: CopyField, label: string, text: string, edit: (text: string) => void, sourceDigest?: string) => {
    const current = value[field]?.[language] ?? "";
    const count = listingTextCount(text);
    const title = field === "titles";
    const limit = title ? (language === "en" ? 500 : localizedTitleLimit) : 20000;
    const longEnglishTitle = title && language === "en" && count > localizedTitleLimit;
    const formatValid = isPlainListingText(text) && !(title && /[\r\n\t]/.test(text));
    const sameSource = !sourceDigest || language === "en" ||
      value.description_translation?.[`${language}_source_digest`] === sourceDigest;
    const eligible = text.trim() !== "" && formatValid && count <= limit &&
      (!sourceDigest || !regionalIsStale) && (text.normalize("NFC") !== current.normalize("NFC") || !sameSource);
    const rows = title ? 3 : 8;
    return <div>
      <div className="listing-field-grid">
        <label className="listing-field">当前{label}
          <textarea rows={rows} value={current} readOnly />
        </label>
        <label className="listing-field">候选{label}
          <textarea rows={rows} value={text} disabled={disabled || generating}
            onChange={e => edit(e.target.value)} />
        </label>
      </div>
      <div className="localized-title-actions">
        <small className={"localized-title-count " + (count > limit || longEnglishTitle || !formatValid ? "title-over-limit" : "")}>
          {count} / {limit} 字符{title && language === "en" ? " · 建议 45–55 字符" : ""}
          {count > limit ? " · 请精简后采纳" : !formatValid ? " · 请改为纯文本" : ""}
          {longEnglishTitle ? " · 已超过 60 字符" : ""}
        </small>
        <div className="localized-title-actions">
          <button type="button" disabled={!text} onClick={() => void copy(text, label)}>复制候选{label}</button>
          <button type="button" disabled={disabled || generating || !eligible}
            onClick={() => {
              onAdopt(field, language, text.normalize("NFC"), sourceDigest);
              setMessage("");
            }}>采纳{label}</button>
        </div>
      </div>
    </div>;
  };

  return <fieldset className="listing-editor-panel">
    <legend>AI 文案候选</legend>
    <div className="localized-title-actions">
      <div className="localized-title-actions">
        <button type="button" disabled={disabled || generating} onClick={() => void generate("titles")}>AI 优化标题</button>
        <button type="button" disabled={disabled || generating} onClick={() => void generate("english")}>AI 生成英文详情</button>
        <button type="button" disabled={disabled || generating || !savedEnglish.trim() || englishDirty}
          onClick={() => void generate("translate")}>{regional ? "重新翻译详情至 MX/BR" : "翻译详情至 MX/BR"}</button>
        {generating && <button type="button" onClick={cancel}>取消生成</button>}
      </div>
    </div>
    <p className="listing-fields-note">生成使用已保存的商品资料。采纳或编辑英文详情后，先保存，再翻译为西语和葡语。</p>
    {(Object.keys(descriptionLabels) as TitleLanguage[]).filter(lang => lang !== "en").map(lang => {
      const current = value.descriptions?.[lang];
      const digest = value.description_translation?.[`${lang}_source_digest`];
      if (!current?.trim()) return null;
      if (!digest) return <p className="listing-fields-note" key={lang}>{descriptionLabels[lang]}的英文来源未记录。</p>;
      if (englishDirty || (digest !== savedDigest && !(regional && !regionalIsStale && digest === regional.source_digest))) {
        return <p className="listing-validation" key={lang}>{descriptionLabels[lang]}可能已过期。可保留现有内容或按已保存英文重新翻译。</p>;
      }
      return null;
    })}
    {regionalIsStale && <p className="listing-validation">英文母版已变化。请保存后重新翻译再采纳。</p>}
    {generating && <p className="listing-fields-note" role="status">正在生成文案…</p>}
    {error && <p className="listing-validation" role="alert">{error}</p>}
    {message && <p className="listing-fields-note" role="status">{message}</p>}
    {(titles || english || regional) && <>
      <div className="listing-language-tabs" role="group" aria-label="候选文案语言">
        {(Object.keys(languages) as TitleLanguage[]).map(lang => <button key={lang} type="button"
          aria-pressed={language === lang} onClick={() => setLanguage(lang)}>{languages[lang]}</button>)}
      </div>
      {titles && renderField("titles", "标题", titles[language], text => setTitles({ ...titles, [language]: text }))}
      {language === "en" && english && <>
        {english.missing_facts.length > 0 && <div>
          <strong>待补充的商品信息</strong>
          <ul>{english.missing_facts.map(fact => <li key={fact}>{fact}</li>)}</ul>
        </div>}
        {renderField("descriptions", descriptionLabels.en, english.description, text => setEnglish({ ...english, description: text }))}
      </>}
      {language !== "en" && regional && <>
        {regional.review_terms.length > 0 && <div>
          <strong>请核对译文术语</strong>
          <ul>{regional.review_terms.map(term => <li key={term}>{term}</li>)}</ul>
        </div>}
        {renderField("descriptions", descriptionLabels[language], regional[language],
          text => setRegional({ ...regional, [language]: text }), regional.source_digest)}
      </>}
    </>}
  </fieldset>;
}
