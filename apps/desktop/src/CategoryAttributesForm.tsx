import { useEffect, useState } from "react";
import { getCBTCategoryRules, type CBTCategoryRules, type CBTAttributeRule } from "./collections";
import type { CategoryAttributeValue } from "./listingFields";

const decimal = /^-?\d+(?:\.\d+)?$/u;
const supportedTypes = new Set(["string", "number", "number_unit", "list"]);
const deferredTags = ["conditional_required", "variation_attribute", "allow_variations", "size_grid_required",
  "multivalued", "read_only", "fixed", "hidden"];

const choicesFor = (rule: CBTAttributeRule) => Array.isArray(rule.values) ? rule.values.filter(v => v?.id != null && v.name) : [];
const unitsFor = (rule: CBTAttributeRule) => Array.isArray(rule.allowed_units) ? rule.allowed_units.filter(u => u?.id && u.name) : [];

function validation(rule: CBTAttributeRule, field?: CategoryAttributeValue) {
  if (!field?.value.trim()) return rule.tags?.required || rule.tags?.new_required ? "待填写必填值" : "";
  if (Array.from(field.value).length > Math.min(rule.value_max_length || 500, 500)) return "超出官方长度上限";
  if (rule.value_type === "list" && !choicesFor(rule).some(choice => String(choice.id) === field.value_id)) {
    return "请选择该类目的官方选项";
  }
  if (rule.value_type === "number" || rule.value_type === "number_unit") {
    if (!decimal.test(field.value)) return "请输入数字";
    if (rule.value_type === "number_unit" && !unitsFor(rule).some(unit => unit.id === field.unit_id)) return "请选择官方单位";
  }
  if (rule.value_type !== "list" && field.value_id) return "请按当前规则重新填写值";
  if (rule.value_type !== "number_unit" && field.unit_id) return "当前字段不接受单位";
  return "";
}

export default function CategoryAttributesForm({ category, values, onChange, disabled }: {
  category: string;
  values: CategoryAttributeValue[];
  onChange: (values: CategoryAttributeValue[]) => void;
  disabled: boolean;
}) {
  const [result, setResult] = useState<CBTCategoryRules | null>(null);
  const [requestError, setRequestError] = useState("");
  const [attempt, setAttempt] = useState(0);
  const validCategory = /^CBT[0-9]{1,15}$/.test(category);
  useEffect(() => {
    setResult(null);
    setRequestError("");
    if (!validCategory) return;
    const controller = new AbortController();
    const signal = AbortSignal.any([controller.signal, AbortSignal.timeout(15_000)]);
    const timer = window.setTimeout(() => {
      void getCBTCategoryRules(category, signal).then(data => {
        if (!controller.signal.aborted) {
          if (data.category_id !== category || !Array.isArray(data.attributes)) throw new Error("官方规则和当前类目不一致");
          setResult(data);
        }
      }).catch((error: unknown) => {
        if (!controller.signal.aborted) setRequestError(error instanceof Error ? error.message : "当前无法读取 CBT 属性");
      });
    }, 250);
    return () => { window.clearTimeout(timer); controller.abort(); };
  }, [category, validCategory, attempt]);

  const current = result?.category_id === category ? result : null;
  const translations = current?.translations_zh ?? {};
  const display = (name: string) => Object.hasOwn(translations, name) ? translations[name] + " · " + name : name;
  const categories = current?.path?.map(node => node.name) ?? [];
  const rules = (current?.attributes ?? []).filter(rule => rule && /^[A-Z][A-Z0-9_]{0,79}$/.test(rule.id));
  const editable = rules.filter(rule => supportedTypes.has(rule.value_type ?? "") &&
    !deferredTags.some(tag => rule.tags?.[tag]) &&
    (rule.value_type !== "list" || choicesFor(rule).length > 0) &&
    (rule.value_type !== "number_unit" || unitsFor(rule).length > 0));
  const newestFirst = values.slice().reverse();
  const active = new Map(values.filter(value => value.category_id === category)
    .map(value => [value.id, value]));
  const retained = current ? values.filter(value => (value.value || value.unit_id) &&
    (value.category_id !== category || !editable.some(rule => rule.id === value.id))) : values;
  const special = current && (!current.technical_specs_verified || current.special_requirements ||
    editable.length !== (current.attributes?.length ?? 0));
  const ranking = current?.technical_requirements?.filter(id =>
    !rules.some(rule => rule.id === id && (rule.tags?.required || rule.tags?.new_required))) ?? [];
  const recordKey = (value: CategoryAttributeValue) => value.category_id + ":" + value.id;

  function setField(id: string, value: string, valueID?: string, unitID?: string) {
    onChange([...values.filter(field => field.category_id !== category || field.id !== id),
      { category_id: category, id, value, value_id: valueID, unit_id: unitID }]);
  }

  return <fieldset>
    <legend>CBT 官方属性</legend>
    {!validCategory && <p className="listing-fields-note">选择 CBT 类目后加载对应属性。</p>}
    {validCategory && !current && !requestError && <p role="status">正在读取类目规则…</p>}
    {requestError && <>
      <p className="listing-validation" role="alert">{requestError}。仍可保存工作稿。</p>
      <button type="button" disabled={disabled} onClick={() => setAttempt(n => n + 1)}>重新读取类目规则</button>
    </>}
    {current && <>
      {categories.length > 0 && <>
        {categories.some(name => Object.hasOwn(translations, name) && translations[name] !== name) &&
          <p className="listing-fields-note">{categories.map(name => Object.hasOwn(translations, name) ? translations[name] : name).join(" / ")}</p>}
        <p className="listing-fields-note">{categories.join(" / ")}</p>
      </>}
      {special && <p className="listing-validation" role="status">该分类存在特殊官方要求，需要在刊登前进一步校验。
        {!current.technical_specs_verified && " 补充技术规范尚未核实。"}</p>}
      {ranking.length > 0 && <p className="listing-fields-note">搜索排序建议补充：{ranking.map(id => {
        const names = current.technical_requirement_names;
        return names && Object.hasOwn(names, id) ? display(names[id]) : id;
      }).join("、")}</p>}
      {rules.length === 0 && <p className="listing-fields-note">官方未返回可用属性。仍需在刊登前校验。</p>}
      {editable.length > 0 && <div className="listing-field-grid">
        {editable.map(rule => {
          const field = active.get(rule.id);
          const label = display(rule.name || rule.id);
          const mandatory = Boolean(rule.tags?.required || rule.tags?.new_required);
          const issue = validation(rule, field);
          const prior = !field?.value.trim() && !field?.unit_id
            ? newestFirst.find(value => value.category_id !== category && value.id === rule.id && value.value.trim())
            : undefined;
          const priorIssue = prior && validation(rule, prior);
          const priorOption = prior?.value_id && choicesFor(rule).find(choice => String(choice.id) === prior.value_id);
          return <div key={rule.id}>
            <label>{label}{mandatory && " · 必填"}
              {rule.value_type === "list" ? <select disabled={disabled} value={field?.value_id ?? ""}
                aria-required={mandatory} aria-invalid={Boolean(issue)}
                onChange={event => {
                  const id = event.target.value;
                  const choice = choicesFor(rule).find(candidate => String(candidate.id) === id);
                  setField(rule.id, choice?.name ?? "", choice ? id : undefined);
                }}>
                <option value="">请选择</option>
                {field?.value_id && !choicesFor(rule).some(choice => String(choice.id) === field.value_id) &&
                  <option value={field.value_id} disabled>{field.value}（已保存，待核实）</option>}
                {choicesFor(rule).map(choice => <option key={String(choice.id)} value={String(choice.id)}>{display(choice.name)}</option>)}
              </select> : <input disabled={disabled} value={field?.value ?? ""}
                inputMode={rule.value_type === "string" ? "text" : "decimal"}
                aria-required={mandatory} aria-invalid={Boolean(issue)}
                maxLength={Math.min(rule.value_max_length || 500, 500)}
                onChange={event => setField(rule.id, event.target.value, undefined,
                  rule.value_type === "number_unit" ? field?.unit_id : undefined)} />}
            </label>
            {rule.value_type === "number_unit" && <label>{label} 单位
              <select disabled={disabled} value={field?.unit_id ?? ""}
                onChange={event => setField(rule.id, field?.value ?? "", undefined, event.target.value || undefined)}>
                <option value="">请选择单位</option>
                {field?.unit_id && !unitsFor(rule).some(unit => unit.id === field.unit_id) &&
                  <option value={field.unit_id} disabled>{field.unit_id}（已保存，待核实）</option>}
                {unitsFor(rule).map(unit => <option key={unit.id} value={unit.id}>{unit.name}</option>)}
              </select>
            </label>}
            {issue && <small className="listing-validation">{issue} · 可保存工作稿</small>}
            {prior && <div>
              <small>历史参考（{prior.category_id}）：{prior.value}{prior.unit_id ? " " + prior.unit_id : ""}</small>
              {priorIssue
                ? <small className="listing-validation"> · 待核实：{priorIssue}</small>
                : <button type="button" disabled={disabled}
                    onClick={() => setField(rule.id, priorOption ? priorOption.name : prior.value, prior.value_id, prior.unit_id)}>
                    沿用此值
                  </button>}
            </div>}
          </div>;
        })}
      </div>}
    </>}
    {retained.length > 0 && <details>
      <summary>{current ? "当前未采用的已填属性" : "已保存属性"}（{retained.length}）</summary>
      <ul>{retained.map(value => <li key={recordKey(value)}>
        {value.category_id} · {value.id} · {value.value}{value.unit_id ? " " + value.unit_id : ""}
        {" "}<button type="button" disabled={disabled}
          onClick={() => onChange(values.filter(field => recordKey(field) !== recordKey(value)))}>移除</button>
      </li>)}</ul>
    </details>}
  </fieldset>;
}
