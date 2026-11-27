import { useEffect, useMemo, useRef, useState } from "react";
import { ensureStoreConnection, storeRequest } from "./stores";
import type { Listing } from "./PublishedListingView";

type Choice = { id?: string | number; name?: string };
type Rule = {
  id: string; name: string; value_type?: string; value_max_length?: number;
  tags?: Record<string, boolean>; values?: Choice[];
  attribute_group_id?: string; attribute_group_name?: string; allowed_units?: unknown;
};
type RulesResponse = { category_id: string; attributes: Rule[]; path?: { id: string; name: string }[]; translations_zh?: Record<string, string> };
type Draft = {
  category_id: string;
  attributes: Record<string, string>;
  title: string; price: string; quantity: string; description: string;
};
type Props = { listing: Listing; site: string; logistic: string; store: string; members: Listing[]; onDirtyChange: (dirty: boolean) => void };

const defaultLabel = "未返回";
const titleLimit = 255;
// Read-only public category metadata; keep cache small and short-lived.
const CATEGORY_RULE_TTL_MS = 3 * 60 * 1000;
const CATEGORY_RULE_MAX_ENTRIES = 32;
const categoryRulesCache = new Map<string, { payload: RulesResponse; expiresAt: number }>();
function getCachedRules(key: string): RulesResponse | undefined {
  const cached = categoryRulesCache.get(key);
  if (!cached) return;
  if (cached.expiresAt <= Date.now()) { categoryRulesCache.delete(key); return; }
  return cached.payload;
}
function rememberRules(key: string, payload: RulesResponse): void {
  categoryRulesCache.delete(key);
  if (categoryRulesCache.size >= CATEGORY_RULE_MAX_ENTRIES) {
    const oldestKey = categoryRulesCache.keys().next().value;
    if (oldestKey !== undefined) categoryRulesCache.delete(oldestKey);
  }
  categoryRulesCache.set(key, { payload, expiresAt: Date.now() + CATEGORY_RULE_TTL_MS });
}
const unitField = /^(PACKAGE_(WEIGHT|HEIGHT|LENGTH|WIDTH)|WEIGHT|HEIGHT|LENGTH|WIDTH)$/;
const charCount = (value: string) => Array.from(value).length;
// Convert only explicit, unambiguous metric units; never invent missing units.
function metricNumber(raw: string): { amount: number; unit: string } | undefined {
  const match = /^(\d+(?:[.,]\d{1,2})?)\s*(kg|g|mm|cm|m)$/i.exec(raw.trim());
  if (!match) return;
  const amount = Number(match[1].replace(",", "."));
  return Number.isFinite(amount) ? { amount, unit: match[2].toLowerCase() } : undefined;
}
const formatMetric = (n: number) => new Intl.NumberFormat("zh-CN", { maximumFractionDigits: 2 }).format(n);
function compactWeight(value?: string): string {
  if (!value) return "未返回";
  const parsed = metricNumber(value);
  if (!parsed || (parsed.unit !== "g" && parsed.unit !== "kg")) return value;
  const grams = parsed.amount * (parsed.unit === "kg" ? 1000 : 1);
  if (grams < 1000) return formatMetric(grams) + " g";
  return new Intl.NumberFormat("zh-CN", { maximumFractionDigits: 5 }).format(grams / 1000) + " kg";
}
function compactDimensions(values: (string | undefined)[]): string {
  if (values.every((v) => !v)) return "未返回";
  const parsed = values.map((v) => v && metricNumber(v));
  if (parsed.every((p, i) => !values[i] || (p && ["mm", "cm", "m"].includes(p.unit)))) {
    const length = parsed.map((p) => {
      if (!p) return "—";
      return formatMetric(p.amount * (p.unit === "m" ? 100 : p.unit === "mm" ? 0.1 : 1));
    });
    return length.join(" × ") + " cm";
  }
  return values.map((v) => v || "—").join(" × ");
}
function initialDraft(item: Listing): Draft {
  return {
    category_id: item.category_id || "",
    attributes: Object.fromEntries((item.attributes || []).filter((a) => a.id && a.value_name != null)
      .map((a) => [a.id, String(a.value_name)])),
    title: item.title || "",
    price: item.price == null ? "" : String(item.price),
    quantity: item.available_quantity == null ? "" : String(item.available_quantity),
    description: item.description || "",
  };
}
function keyFor(store: string, site: string, logistic: string, item: Listing): string {
  return "una-seller:published-draft:v1:" + [store, site, logistic, item.id].join(":");
}
function loadDraft(key: string, item: Listing): Draft {
  const base = initialDraft(item);
  try {
    const raw = localStorage.getItem(key);
    if (!raw) return base;
    const saved = JSON.parse(raw) as Partial<Draft>;
    if (!saved || typeof saved !== "object" || typeof saved.category_id !== "string" ||
        typeof saved.title !== "string" || typeof saved.price !== "string" ||
        typeof saved.quantity !== "string" || typeof saved.description !== "string" ||
        !saved.attributes || typeof saved.attributes !== "object" || Array.isArray(saved.attributes)) return base;
    const attributes = Object.fromEntries(Object.entries(saved.attributes)
      .filter(([id, value]) => /^[A-Z0-9_]{1,80}$/.test(id) && typeof value === "string"));
    return { ...base, ...saved, attributes: { ...base.attributes, ...attributes } };
  } catch { return base; }
}
function normalizedChoices(rule: Rule): Choice[] {
  return (rule.values || []).filter((v) => v && v.name != null);
}
function ruleRequired(rule: Rule): boolean { return rule.tags?.required === true; }
function currentValue(draft: Draft, rule: Rule): string { return draft.attributes[rule.id] || ""; }
function isFixed(rule: Rule): boolean { return rule.tags?.fixed === true || rule.tags?.variation_attribute === true; }

function Field({ label, value }: { label: string; value?: string | number | null }) {
  return <div className="published-field"><span>{label}</span><strong>{value == null || value === "" ? defaultLabel : value}</strong></div>;
}
function AttributeInput({ rule, value, change, display }: { rule: Rule; value: string; change: (v: string) => void; display: (name: string) => string }) {
  const choices = normalizedChoices(rule);
  const locked = isFixed(rule) || rule.tags?.hidden === true;
  if (choices.length > 0) {
    return <select disabled={locked} value={value} onChange={(e) => change(e.target.value)}>
      <option value="">请选择</option>
      {choices.map((choice, i) => <option key={String(choice.id ?? i)} value={choice.name || ""}>{display(choice.name || "")}</option>)}
      {value && !choices.some((v) => v.name === value) && <option value={value}>{value}（原始值）</option>}
    </select>;
  }
  return <input disabled={locked} type={rule.value_type === "number" ? "number" : "text"}
    maxLength={rule.value_max_length && rule.value_max_length > 0 ? rule.value_max_length : undefined}
    value={value} onChange={(e) => change(e.target.value)}
    placeholder={rule.value_type === "number_unit" ? "数值与单位，例如 10 kg" : "按真实产品资料填写"}/>;
}
function classify(rule: Rule): string {
  if (ruleRequired(rule)) return "必填";
  if (rule.tags?.conditional_required) return "条件必填 · 待平台判断";
  return "选填";
}
function AttributeGroup({ title, fields, draft, onChange, display }: {
  title: string; fields: Rule[]; draft: Draft; onChange: (id: string, value: string) => void; display: (name: string) => string;
}) {
  const [expanded, setExpanded] = useState(false);
  const visible = expanded ? fields : fields.filter((a) =>
    ruleRequired(a) || a.tags?.conditional_required || Boolean(currentValue(draft, a))).slice(0, 18);
  return <div className="published-attr-group">
    <div className="published-attr-heading"><strong>{title}</strong><small>{fields.length} 项属性</small></div>
    <div className="published-edit-grid">
      {visible.map((rule) => <label className="published-edit-field" key={rule.id}>
        <span>{display(rule.name || rule.id)}{ruleRequired(rule) && <b className="published-required"> *</b>}</span>
        <AttributeInput rule={rule} value={currentValue(draft, rule)} change={(v) => onChange(rule.id, v)} display={display}/>
        <small>{classify(rule)}{isFixed(rule) ? " · 固定或 SKU 级属性，不在此编辑" : ""}
          {rule.value_max_length ? " · 最长 " + rule.value_max_length + " 字符" : ""}</small>
      </label>)}
    </div>
    {expanded
      ? <button type="button" className="published-inline-action"
          onClick={() => setExpanded(false)}>收起选填属性</button>
      : fields.length > visible.length && <button type="button" className="published-inline-action"
          onClick={() => setExpanded(true)}>显示其余 {fields.length - visible.length} 项选填属性</button>}
  </div>;
}
export default function PublishedListingEditor({ listing, site, logistic, store, members, onDirtyChange }: Props) {
  const localKey = keyFor(store, site, logistic, listing);
  const [draft, setDraft] = useState<Draft>(() => loadDraft(localKey, listing));
  const [saved, setSaved] = useState("");
  const [rules, setRules] = useState<Rule[] | null>(null);
  const [categoryPath, setCategoryPath] = useState<string[]>([]);
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const display = (name: string) => Object.hasOwn(translations, name) ? translations[name] + " · " + name : name;
  const [rulesError, setRulesError] = useState("");
  const [loading, setLoading] = useState(false);
  const [version, setVersion] = useState(0);
  const [descriptionReload, setDescriptionReload] = useState(0);
  const [remoteDescription, setRemoteDescription] = useState("");
  const [descriptionState, setDescriptionState] = useState<"idle" | "loading" | "loaded" | "empty" | "error">("idle");
  const [descriptionError, setDescriptionError] = useState("");
  const descriptionTouched = useRef(false);

  useEffect(() => {
    descriptionTouched.current = false;
    setDraft(loadDraft(localKey, listing)); setSaved("");
  }, [localKey]);

  useEffect(() => {
    if (listing.description) {
      setRemoteDescription(listing.description);
      setDescriptionState("loaded");
      return;
    }
    const controller = new AbortController();
    setDescriptionState("loading");
    setDescriptionError("");
    const path = "/api/v1/stores/" + encodeURIComponent(store) + "/published-listings/" +
      encodeURIComponent(listing.id) + "/description?" +
      new URLSearchParams({ site, logistic_type: logistic }).toString();
    ensureStoreConnection(store, controller.signal).then(() =>
      storeRequest<{ description: string }>(path, "GET", controller.signal, true),
    ).then((response) => {
      if (controller.signal.aborted) return;
      if (!response || typeof response.description !== "string" || !response.description.trim()) {
        setDescriptionState("empty");
        return;
      }
      setRemoteDescription(response.description);
      setDescriptionState("loaded");
      // Keep a previously saved local draft untouched; also never overwrite active typing.
      let hasSavedDraft = true;
      try { hasSavedDraft = localStorage.getItem(localKey) !== null; } catch { /* conservative */ }
      if (!hasSavedDraft && !descriptionTouched.current) {
        setDraft((old) => old.description ? old : { ...old, description: response.description });
      }
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) {
        setDescriptionState("error");
        setDescriptionError(error instanceof Error ? error.message : "无法读取平台商品描述");
      }
    });
    return () => controller.abort();
  }, [localKey, store, site, logistic, listing.id, listing.description, descriptionReload]);
  const category = draft.category_id.trim();
  const cbtPath = category === listing.category_id ? listing.cbt_category_path : undefined;
  const [settledCategory, setSettledCategory] = useState(category);
  const validCategory = new RegExp("^" + site + "[0-9]{1,15}$").test(category);
  useEffect(() => {
    const timer = setTimeout(() => setSettledCategory(category), 500);
    return () => clearTimeout(timer);
  }, [category]);
  useEffect(() => {
    const controller = new AbortController();
    setRules(null); setCategoryPath([]); setTranslations({}); setRulesError(""); setLoading(false);
    if (!validCategory || settledCategory !== category) return () => controller.abort();
    const cacheKey = [store, site, logistic, category].join(":");
    const applyRules = (response: RulesResponse) => {
      setRules(response.attributes.filter((r) => r && /^[A-Z0-9_]+$/.test(r.id)));
      setTranslations(response.translations_zh ?? {});
      setCategoryPath((response.path || []).filter((p) => p?.name).map((p) => p.name));
    };
    const cached = getCachedRules(cacheKey);
    if (cached) {
      applyRules(cached);
      return () => controller.abort();
    }
    setLoading(true);
    const path = "/api/v1/stores/" + encodeURIComponent(store) + "/category-attributes?" +
      new URLSearchParams({ site, logistic_type: logistic, category_id: category }).toString();
    storeRequest<RulesResponse>(path, "GET", controller.signal, true).then((response) => {
      if (!controller.signal.aborted) {
        if (response.category_id !== category || !Array.isArray(response.attributes)) {
          setRulesError("类目属性响应与当前分类不一致。"); return;
        }
        rememberRules(cacheKey, response);
        applyRules(response);
      }
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setRulesError(error instanceof Error ? error.message : "无法读取在线类目属性");
    }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [store, site, logistic, category, settledCategory, version]);

  const groups = useMemo(() => {
    const values = (rules || []).filter((r) => !unitField.test(r.id) && (!r.tags?.hidden || draft.attributes[r.id]));
    const byGroup = new Map<string, Rule[]>();
    for (const rule of values) {
      const key = rule.attribute_group_name || "其他属性";
      const group = byGroup.get(key);
      if (group) group.push(rule); else byGroup.set(key, [rule]);
    }
    return [...byGroup].map(([name, list]) => ({ name, fields: list.sort((a,b) =>
      Number(ruleRequired(b)) - Number(ruleRequired(a)) || a.name.localeCompare(b.name)) }));
  }, [rules, draft.attributes]);

  const missing = (rules || []).filter((r) => ruleRequired(r) &&
    !isFixed(r) && !r.tags?.hidden && !currentValue(draft, r).trim()).map((r) => r.name || r.id);
  const platformRequired = (rules || []).filter((r) => ruleRequired(r) &&
    (isFixed(r) || r.tags?.hidden) && !currentValue(draft, r).trim()).map((r) => r.name || r.id);
  const priceValid = draft.price === "" || (Number.isFinite(Number(draft.price)) && Number(draft.price) > 0);
  const stockValid = draft.quantity === "" || (/^[0-9]+$/.test(draft.quantity) && Number.isSafeInteger(Number(draft.quantity)));
  const attrs = listing.attributes || [];
  const fallbackAttributes = attrs.filter((a) => a.value_name != null && a.id !== "SELLER_SKU" && !unitField.test(a.id));
  const metricRules = (rules || []).filter((rule) => unitField.test(rule.id) && !rule.tags?.hidden);
  const metricPrefix = Object.entries(draft.attributes).some(([id, value]) =>
    unitField.test(id) && id.startsWith("PACKAGE_") && value.trim()) ? "PACKAGE_" : "";
  const metric = (type: string): string | undefined => {
    const id = metricPrefix + type;
    return draft.attributes[id]?.trim() || undefined;
  };

  function update(field: keyof Omit<Draft, "attributes">, value: string) {
    onDirtyChange(true);
    if (field === "description") descriptionTouched.current = true;
    setDraft((old) => ({ ...old, [field]: value })); setSaved("");
  }
  function changeAttr(id: string, value: string) {
    onDirtyChange(true);
    setDraft((old) => ({ ...old, attributes: { ...old.attributes, [id]: value } })); setSaved("");
  }
  function save() {
    try {
      localStorage.setItem(localKey, JSON.stringify(draft));
      onDirtyChange(false);
      setSaved("草稿已保存在本机。不会修改 Mercado Libre 线上商品。");
    } catch { setSaved("本地存储失败，请检查客户端存储空间。"); }
  }
  function reset() {
    if (!window.confirm("清除当前刊登在本机的编辑草稿并恢复已同步资料？")) return;
    try { localStorage.removeItem(localKey); } catch { /* still reset visible form */ }
    descriptionTouched.current = false;
    setDraft({ ...initialDraft(listing), description: remoteDescription || listing.description || "" });
    onDirtyChange(false);
    setSaved("已恢复已读取的商品资料。");
  }
  return <>
    <section className="published-block">
      <div className="published-block-head"><h3>商品分类</h3><span>仅编辑本地草稿</span></div>
      {cbtPath && <>
        <p className="published-category-path">{site !== "CBT" && "CBT 分类 · "}{(cbtPath.zh ?? cbtPath.original).join(" / ")}</p>
        {cbtPath.zh && <p className="published-hint">{cbtPath.original.join(" / ")}</p>}
      </>}
      {site !== "CBT" || !cbtPath ? <p className="published-category-path">
        {categoryPath.length ? categoryPath.join(" / ") : loading ? "正在加载分类名称…" : "官方分类名称暂未返回"}
      </p> : null}
      <details className="published-detail-expand">
        <summary>调整类目（高级）</summary>
        <label className="published-edit-field">
          <span>站点类目 ID</span>
          <input value={draft.category_id} onChange={(e) => update("category_id", e.target.value.toUpperCase())}
            placeholder={site + " 类目 ID"} aria-label="类目 ID"/>
          <small>仅修改本地草稿。更换类目后将重新加载官方属性，不改变平台商品。</small>
        </label>
      </details>
      {!validCategory && <p className="published-form-warning">请输入当前站点的有效类目后加载规则。</p>}
    </section>

    <section className="published-block">
      <div className="published-block-head"><h3>分类细项与属性</h3>
        <span>{loading ? "正在加载 Mercado Libre 类目规则…" : settledCategory !== category ? "输入结束后加载…" : rules ? "官方规则已加载 · " + rules.length + " 项" : "等待规则"}</span>
      </div>
      {rulesError && <>
        <p role="alert" className="published-form-warning">{rulesError}。无法验证必填属性，当前仅能保留本地草稿。</p>
        <button type="button" className="published-inline-action" onClick={() => setVersion((v) => v + 1)}>重新加载</button>
      </>}
      {rules && <>
        <p className="published-hint">红色 * 为官方硬性必填。条件必填需要平台结合商品信息进一步判定。
          数值及单位属性目前仅保存在草稿中，后续提交平台前必须转换为对应的 number / unit 结构。</p>
        {missing.length > 0 && <p role="status" className="published-form-warning">
          目前缺少 {missing.length} 项可编辑必填：{missing.slice(0, 5).join("、")}{missing.length > 5 ? "等" : ""}。
        </p>}
        {platformRequired.length > 0 && <p role="status" className="published-form-warning">
          另有 {platformRequired.length} 项平台固定或 SKU 级必填未返回值，
          需在对应平台 / SKU 维护，不能在此表单直接修改。
        </p>}
        {groups.map((group) => <AttributeGroup key={group.name} title={group.name}
          fields={group.fields} draft={draft} onChange={changeAttr} display={display}/>)}
        {groups.length === 0 && metricRules.length === 0 && <p className="published-hint">该类目未返回可编辑属性。</p>}
      </>}
      {!rules && !loading && fallbackAttributes.length > 0 && <div className="published-field-grid">
        {fallbackAttributes.map((a) => <Field key={a.id} label={a.name || a.id} value={a.value_name}/>)}
      </div>}
    </section>

    <section className="published-block published-metrics-block">
      <div className="published-metrics-inline">
        <span>{metricPrefix ? "包装" : "商品"}重量：<strong>{compactWeight(metric("WEIGHT"))}</strong></span>
        <span>{metricPrefix ? "包装" : "商品"}尺寸（长 × 宽 × 高）：<strong>{compactDimensions([metric("LENGTH"), metric("WIDTH"), metric("HEIGHT")])}</strong></span>
      </div>
      {metricRules.length > 0 && <details className="published-detail-expand">
        <summary>编辑重量与尺寸</summary>
        <p className="published-hint">商品与包装分别填写，使用实际单位。修改仅保存在本机草稿。</p>
        <div className="published-edit-grid">
          {metricRules.map((rule) => <label className="published-edit-field" key={rule.id}>
            <span>{rule.id.startsWith("PACKAGE_") ? "包装 · " : "商品 · "}{display(rule.name || rule.id)}
              {ruleRequired(rule) && <b className="published-required"> *</b>}</span>
            <AttributeInput rule={rule} value={currentValue(draft, rule)} change={(value) => changeAttr(rule.id, value)} display={display} />
            <small>{classify(rule)}{isFixed(rule) ? " · 固定或 SKU 级属性" : ""}</small>
          </label>)}
        </div>
      </details>}
    </section>

    <section className="published-block">
      <div className="published-block-head"><h3>价格与库存</h3><span>{site} 当前站点 · 本地草稿</span></div>
      <div className="published-edit-grid">
        <label className="published-edit-field"><span>站点售价 {listing.currency_id || ""}</span>
          <input type="number" min="0.01" step="0.01" value={draft.price} onChange={(e) => update("price", e.target.value)} />
          {!priceValid && <small className="published-form-error">售价必须大于 0</small>}
        </label>
        <label className="published-edit-field"><span>可售库存</span>
          <input type="number" min="0" step="1" value={draft.quantity} onChange={(e) => update("quantity", e.target.value)} />
          {!stockValid && <small className="published-form-error">库存必须为非负整数</small>}
        </label>
      </div>
      <div className="published-field-grid published-readonly-metrics">
        <Field label="预期净回款" value="尚未接入平台回款数据"/>
        <Field label="促销优惠" value="尚未接入平台促销数据"/>
      </div>
      {members.length > 1 && <details className="published-detail-expand">
        <summary>查看本页同家族各款式（{members.length} 条）</summary>
        <div className="published-family-prices">
          {members.map((item) => <div key={item.id}><span>{item.seller_custom_field || item.id}</span>
            <strong>{item.price ?? "—"} {item.currency_id || ""} · 库存 {item.available_quantity ?? "—"}</strong></div>)}
        </div>
      </details>}
    </section>

    <section className="published-block">
      <div className="published-block-head"><h3>商品标题</h3><span>当前站点语言，不做自动翻译</span></div>
      <div className="published-counter-wrap">
        <textarea className="published-edit-textarea" rows={3} value={draft.title}
          onChange={(e) => update("title", e.target.value)} maxLength={titleLimit}
          placeholder="根据真实商品信息填写标题" />
        <span className="published-char-counter">{draft.title.length} / {titleLimit} 字符</span>
      </div>
      <p className="published-hint">计数包含空格、标点及符号。仅保存到本机，不会更新线上刊登。</p>
    </section>

    <section className="published-block">
      <div className="published-block-head"><h3>商品详情</h3>
        <span>{descriptionState === "loading" ? "正在读取平台描述…" :
          descriptionState === "loaded" ? "已读取平台描述 · 本地可编辑" :
          descriptionState === "empty" ? "平台未返回完整描述" : "完整描述"}</span>
      </div>
      <div className="published-counter-wrap">
        <textarea className="published-edit-textarea" rows={6} value={draft.description}
          onChange={(e) => update("description", e.target.value)}
          placeholder="平台描述尚未读取；可先录入本地草稿" />
        <span className="published-char-counter">{charCount(draft.description)} 字符</span>
      </div>
      <div role="status">
        {descriptionState === "error" && <p className="published-form-warning">{descriptionError}。本地草稿不受影响。</p>}
      </div>
      {descriptionState === "error" &&
        <button type="button" className="published-inline-action" onClick={() => setDescriptionReload((v) => v + 1)}>重新读取</button>
      }
      {remoteDescription && remoteDescription !== draft.description && <button type="button"
        className="published-inline-action" onClick={() => {
          if (window.confirm("使用平台描述替换当前草稿中的描述？")) update("description", remoteDescription);
        }}>用平台描述填入当前草稿</button>}
      <details className="published-detail-expand"><summary>平台标识与同步信息</summary>
        <div className="published-field-grid">
          <Field label="刊登 ID" value={listing.id}/>
          <Field label="User Product" value={listing.user_product_id}/>
          <Field label="Siteless User Product" value={listing.siteless_user_product_id}/>
          <Field label="最近检查时间" value={listing.detail_unavailable_at
            ? new Date(listing.detail_unavailable_at).toLocaleString() : "平台未返回异常"}/>
          <Field label="全球关联" value={listing.cbt_item_id}/>
          <Field label="部分响应" value={listing.partial ? "是" : "否"}/>
        </div>
      </details>
    </section>
    <div className="published-draft-toolbar">
      <div><strong>商品编辑草稿</strong>
        <p>只存储在当前电脑；不调用商品刊登或修改接口。</p>
        <p>{rules ? missing.length ? "可编辑硬性必填缺失：" + missing.length + " 项。" : "当前可编辑硬性必填无缺失。"
          : "平台类目规则未加载，无法判断是否符合发布要求。"}
          {platformRequired.length ? " 另有 " + platformRequired.length + " 项平台 / SKU 级必填待处理。" : ""}
          {" "}条件必填仍需服务端与平台进一步验证。</p>
        {saved && <p role="status">{saved}</p>}
      </div>
      <div className="published-draft-actions">
        <button type="button" onClick={reset} className="published-light-button">恢复同步值</button>
        <button type="button" onClick={save} className="published-save-button">保存本机草稿</button>
      </div>
    </div>
    {(!priceValid || !stockValid) && <p role="alert" className="published-form-warning">存在基础校验错误，请修正后再用于后续提交。</p>}
  </>;
}
