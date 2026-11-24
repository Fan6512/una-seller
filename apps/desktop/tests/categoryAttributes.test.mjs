import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { promisify } from "node:util";
import { createServer } from "vite";

const fixture = String.raw`
import React, { useState } from "react";
import { createRoot } from "react-dom/client";
import { flushSync } from "react-dom";
import ListingFieldsEditor from "/src/ListingFieldsEditor.tsx";
import PublishedListingEditor from "/src/PublishedListingEditor.tsx";
import { initialListingFields, validateListingFields } from "/src/listingFields.ts";
import { setSiteTitleMode } from "/src/titleModePreference.ts";

const fields = [
  { id: "BRAND", name: "Brand", value_type: "string", tags: { required: true } },
  { id: "COUNT", name: "Count", value_type: "number", tags: { new_required: true } },
  { id: "CAPACITY", name: "Capacity", value_type: "number_unit", tags: { required: true },
    allowed_units: [{ id: "l", name: "L" }] },
  { id: "MATERIAL", name: "Material", value_type: "list", values: [{ id: 200, name: "Glass" }] },
  { id: "EMPTY_LIST", name: "Empty List", value_type: "list" },
];
const master = "For home use.";
const originalDigest = "a".repeat(64);
const revisedDigest = "b".repeat(64);
window.fetch = (path, options) => {
  const url = new URL(path, location.href);
  if (url.pathname === "/api/v1/products/ai-copy") {
    aiRequest = JSON.parse(options.body);
    if (options.method !== "POST" || !options.headers.Authorization) throw Error("missing AI method/auth");
    if (aiMode === "wait") return new Promise((_, reject) => options.signal.addEventListener("abort", () => {
      aiCancelled = true;
      reject(options.signal.reason);
    }, { once: true }));
    if (aiMode === "error") return Promise.resolve(Response.json({ error: "AI fixture unavailable" }, { status: 502 }));
    if (aiRequest.operation === "titles") return Promise.resolve(Response.json({
      titles: { en: "Refined Glass Pitcher " + "X".repeat(45), es: "é".repeat(61), pt: "Jarra de vidro" },
    }));
    if (aiRequest.operation === "english") return Promise.resolve(Response.json({
      description: "For home use.", selling_points: ["Glass body"], missing_facts: ["Capacity"],
    }));
    if (aiRequest.operation === "translate") return Promise.resolve(Response.json({
      es: "Cuerpo de vidrio\n\nPara el hogar.", pt: "Corpo de vidro\n\nPara uso doméstico.",
      source_digest: originalDigest, review_terms: ["Glass body"],
    }));
    throw Error("unexpected AI operation");
  }
  if (url.pathname === "/api/v1/categories/predict") {
    return Promise.resolve(Response.json({ items: [{ category_id: "CBT2", category_name: "Other", domain_name: "Mixed",
      path: [{ id: "CBT1", name: "Home" }, { id: "CBT2", name: "Other" }], path_zh: ["家居", "Other"] }] }));
  }
  const id = url.searchParams.get("category_id");
  if (id === "CBT4") return Promise.resolve(Response.json({ error: "upstream failed" }, { status: 502 }));
  if (!options.headers.Authorization) throw Error("missing bearer");
  if (id === "MLB771") {
    return Promise.resolve(Response.json({ category_id: id, attributes: [], path: [{ id, name: "Jarras" }] }));
  }
  return Promise.resolve(Response.json({
    category_id: id, technical_specs_verified: true,
    translations_zh: id === "CBT2" ? { Brand: "品牌", Plastic: "塑料", Style: "款式" } : undefined,
    path_zh: id === "CBT2" ? ["家居", "Other"] : undefined,
    technical_requirements: id === "CBT2" ? ["STYLE","READONLY_CODE","SPECIAL_GRID"] : id === "CBT1055" ? ["INTERNAL_MEMORY"] : undefined,
    technical_requirement_names: id === "CBT2" ? { STYLE: "Style" } : undefined,
    path: [{ id: "CBT1", name: "Home" }, { id, name: id === "CBT1" ? "Jugs" : "Other" }],
    attributes: id === "CBT1" ? fields : id === "CBT2" ? [
      fields[0],
      { ...fields[2], allowed_units: [{ id: "ml", name: "mL" }] },
      { ...fields[3], values: [{ id: 300, name: "Plastic" }] },
      {id:"READONLY_CODE",name:"Read-only code",value_type:"string",tags:{read_only:true}},
      {id:"SPECIAL_GRID",name:"Special grid",value_type:"string",tags:{size_grid_required:true}},
    ] : id === "CBT1055" ? [
      {id:"IS_DUAL_SIM",name:"Is Dual SIM",hierarchy:"PARENT_PK",value_type:"boolean",tags:{required:true},
        values:[{id:"242084",name:"No"},{id:"242085",name:"Yes"}]},
      {id:"COLOR",name:"Color",hierarchy:"CHILD_PK",value_type:"string",tags:{required:true,allow_variations:true},
        values:[{id:"52049",name:"Black"},{id:"52055",name:"White"},{id:"52028",name:"Blue"}]},
      {id:"INTERNAL_MEMORY",name:"Internal memory",hierarchy:"CHILD_PK",value_type:"number_unit",
        allowed_units:[{id:"GB",name:"GB"},{id:"MB",name:"MB"}]},
      {id:"MAIN_COLOR",name:"Main color",hierarchy:"CHILD_DEPENDENT",value_type:"list",values:[{id:"2450295",name:"Black"}]},
    ] : [],
  }));
};
let editor, working, aiRequest, aiCancelled, aiMode, saveAccepted, reloadSaved, setEditorDisabled, setEditorValue;
function Fixture() {
  const [value, setValue] = useState(() => initialListingFields({
    category_id: "CBT1", category_path: "Home / Jugs",
    titles: { en: "Sample Jug", es: "", pt: "" },
  }));
  const [saved, setSaved] = useState(undefined);
  const [revision, setRevision] = useState(1);
  const [editorKey, setEditorKey] = useState(1);
  const [disabled, setDisabled] = useState(false);
  setEditorDisabled = setDisabled;
  setEditorValue = setValue;
  editor = value;
  working = value.category_attributes ?? [];
  saveAccepted = () => {
    const next = JSON.parse(JSON.stringify(editor));
    if (next.descriptions) next.description_translation = { ...next.description_translation,
      en_digest: next.descriptions.en === master ? originalDigest : revisedDigest };
    setSaved(next);
    setValue(next);
    setRevision(n => n + 1);
  };
  reloadSaved = () => {
    setValue(initialListingFields(JSON.parse(JSON.stringify(saved))));
    setEditorKey(n => n + 1);
  };
  return React.createElement(ListingFieldsEditor, {
    key: editorKey, value, onChange: setValue, aiSaved: saved, disabled,
    aiProduct: { source_product_id: "fixture-product", expected_revision: revision },
  });
}
const assert = (condition, message) => { if (!condition) throw Error(message); };
const loaded = async (selector, text) => {
  for (let i = 0; i < 120; i++) {
    const result = [...document.querySelectorAll(selector)].find(node => node.textContent.includes(text));
    if (result) return result;
    await new Promise(resolve => setTimeout(resolve, 12));
  }
  throw Error("Timed out: " + text);
};
const input = (node, value) => flushSync(() => {
  const setter = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(node), "value").set;
  setter.call(node, value);
  node.dispatchEvent(new Event(node instanceof HTMLSelectElement ? "change" : "input", { bubbles: true }));
});
const switchCategory = id => input(document.querySelector('input[placeholder="例如 CBT12345"]'), id);
try {
  const root = createRoot(document.querySelector("main"));
  flushSync(() => root.render(React.createElement(Fixture)));
  const brand = await loaded("label", "Brand");
  assert(brand.querySelector("input").getAttribute("aria-required") === "true", "required text missing");
  assert((await loaded("label", "Count")).querySelector("input").getAttribute("aria-required") === "true", "new_required number missing");
  input(brand.querySelector("input"), "Sample Brand");
  input((await loaded("label", "Count")).querySelector("input"), "3");
  input((await loaded("label", "Material")).querySelector("select"), "200");
  input((await loaded("label", "Capacity")).querySelector("input"), "1.5");
  input((await loaded("label", "Capacity 单位")).querySelector("select"), "l");
  await loaded("p", "存在特殊官方要求");
  assert(![...document.querySelectorAll("label")].some(node => node.textContent.includes("Empty List")), "missing enumeration became editable");
  const draft = initialListingFields({ category_attributes: working }, undefined, "Sample Jug");
  assert(draft.category_attributes.find(v => v.id === "MATERIAL").value_id === "200", "option ID missing");
  assert(draft.category_attributes.find(v => v.id === "CAPACITY").unit_id === "l", "unit ID missing");
  assert(validateListingFields(draft) === "", "valid work draft rejected");
  assert(validateListingFields(initialListingFields(null, undefined, "Old Draft")) === "", "legacy draft rejected");
  for (const [language, key, title] of [["西语 MX", "es", "Jarra"], ["葡语 BR", "pt", "Jarro"]]) {
    const tab = await loaded("button", language);
    flushSync(() => tab.click());
    input(document.querySelector(".listing-active-title textarea"), title);
    assert(editor.titles[key] === title && editor.category_id === "CBT1" && editor.category_path === "Home / Jugs",
      "non-English title edit cleared the confirmed category");
  }
  const englishTab = await loaded("button", "英语 EN");
  flushSync(() => englishTab.click());
  input(document.querySelector(".listing-active-title textarea"), "Sample Pitcher");
  assert(editor.titles.en === "Sample Pitcher" && editor.category_id === "" && editor.category_path === "",
    "English title edit kept the old category");
  const reopened = initialListingFields(JSON.parse(JSON.stringify(editor)));
  assert(reopened.category_id === "" && reopened.category_path === "" &&
    JSON.stringify(reopened.category_attributes) === JSON.stringify(draft.category_attributes),
    "draft round trip lost historical attributes or restored the category");
  switchCategory("CBT2");
  await loaded("small", "历史参考（CBT1）：Sample Brand");
  const newBrand = (await loaded("label", "品牌 · Brand")).querySelector("input");
  await loaded("p", "家居 / Other");
  await loaded("p", "Home / Other");
  const deferredTechnical = await loaded("p", "款式 · Style");
  assert(deferredTechnical.textContent.includes("Read-only code（平台依赖／只读")&&
    deferredTechnical.textContent.includes("Special grid（暂无可编辑控件"),"required technical fields without editable controls were omitted");
  await loaded("option", "塑料 · Plastic");
  assert(newBrand.value === "" && working.length === 4 && working.every(field => field.category_id === "CBT1"),
    "changing category accepted or lost historical values");
  assert((await loaded("label", "Material")).querySelector("select").value === "", "historical enum was applied to the wrong category");
  assert((await loaded("label", "Capacity 单位")).querySelector("select").value === "", "historical unit was applied to the wrong category");
  await loaded("small", "待核实：请选择该类目的官方选项");
  await loaded("small", "待核实：请选择官方单位");
  const reuseButtons = [...document.querySelectorAll("button")].filter(node => node.textContent.trim() === "沿用此值");
  assert(reuseButtons.length === 1, "incompatible choices must not be offered for reuse");
  flushSync(() => reuseButtons[0].click());
  assert(newBrand.value === "Sample Brand" &&
    working.some(field => field.category_id === "CBT2" && field.id === "BRAND" && field.value === "Sample Brand") &&
    working.filter(field => field.category_id === "CBT1").length === 4,
    "deliberate confirmation did not save a separate current-category value");
  const translatedOption = (await loaded("label", "Material")).querySelector("select");
  input(translatedOption, "300");
  assert(working.some(field => field.category_id === "CBT2" && field.id === "MATERIAL" &&
    field.value_id === "300" && field.value === "Plastic"), "translated enum altered the official value");
  await loaded("summary", "当前未采用的已填属性");
  switchCategory("CBT3");
  await loaded("p", "官方未返回可用属性");
  switchCategory("CBT4");
  await loaded('[role="alert"]', "仍可保存工作稿");
  flushSync(() => [...document.querySelectorAll("button")].find(node => node.textContent.includes("推荐 CBT 类目")).click());
  const suggestion = await loaded(".category-suggestion-list button", "家居 / Other");
  assert(suggestion.textContent.includes("Home / Other"), "suggestion lost its official path");
  flushSync(() => suggestion.click());
  assert(editor.category_id === "CBT2" && editor.category_path === "Home / Other", "translated prediction altered confirmed official data");
  const beforeAI = JSON.stringify(editor);
  const click = async (selector, text) => {
    const button = await loaded(selector, text);
    flushSync(() => button.click());
  };
  await click("button", "AI 优化标题");
  await loaded("small", "已超过 60 字符");
  assert(JSON.stringify(editor) === beforeAI, "generation applied unapproved content");
  assert(aiRequest.source_product_id === "fixture-product" && aiRequest.operation === "titles" &&
    aiRequest.expected_revision === 1 && Object.keys(aiRequest).length === 3, "title request sent unsaved metadata");
  const candidateTab = '[aria-label="候选文案语言"] button';
  await click(candidateTab, "西语 MX");
  const adoptTitle = await loaded("button", "采纳标题");
  assert(adoptTitle.disabled, "61-char Spanish candidate eligible for adoption");
  const editableTitle = (await loaded("label", "候选标题")).querySelector("textarea");
  input(editableTitle, "e\u0301".repeat(59) + "!");
  await loaded("small", "60 / 60 字符");
  assert(!adoptTitle.disabled, "60-char NFC Spanish title unavailable");
  flushSync(() => adoptTitle.click());
  assert(editor.titles.es === "é".repeat(59) + "!" && editor.category_id === "CBT2" &&
    editor.category_path === "Home / Other", "Spanish adoption modified the category or persisted decomposed text");
  await click(candidateTab, "英语 EN");
  const englishAdoptTitle = await loaded("button", "采纳标题");
  assert(englishAdoptTitle !== adoptTitle && !adoptTitle.isConnected, "candidate input group was reused across languages");
  assert(!englishAdoptTitle.disabled, "English title longer than 60 should only warn");
  input((await loaded("label", "候选标题")).querySelector("textarea"), "Refined Glass Pitcher");
  const history = JSON.stringify(editor.category_attributes);
  let confirmed = false, dialogCount = 0;
  window.confirm = message => {
    dialogCount++;
    assert(message === "采纳新的英文标题会清空当前 CBT 类目及路径，但已填写属性会保留为历史参考。", "wrong confirmation message");
    return confirmed;
  };
  await click("button", "采纳标题");
  assert(editor.titles.en === "Sample Pitcher" && editor.category_id === "CBT2" && dialogCount === 1,
    "cancelled English candidate applied");
  confirmed = true;
  await click("button", "采纳标题");
  assert(editor.titles.en === "Refined Glass Pitcher" && editor.category_id === "" && editor.category_path === "" &&
    JSON.stringify(editor.category_attributes) === history && dialogCount === 2,
    "confirmed English candidate lost history or failed to clear category");

  assert((await loaded("button", "翻译详情至 MX/BR")).disabled, "translation enabled before saving an English master");
  await click("button", "AI 生成英文详情");
  await loaded("li", "Capacity");
  assert(aiRequest.operation === "english" && !editor.descriptions &&
    ![...document.querySelectorAll("label")].some(node => node.textContent.includes("候选西语详情")),
    "English-only generation created regional copy or overwrote saved descriptions");
  const copied = [];
  Object.defineProperty(navigator, "clipboard", { value: { writeText: async text => copied.push(text) } });
  await click("button", "单独采纳英文卖点");
  assert(editor.bullet_points.length === 1 && editor.bullet_points[0] === "Glass body" &&
    !editor.descriptions, "selling points were not adopted independently");
  await click("button", "复制候选英文详情");
  assert(copied[0] === master && !editor.descriptions, "English description candidate was combined with selling points or saved itself");
  await click("button", "采纳英文详情");
  assert(editor.descriptions.en === master && editor.descriptions.es === "" && editor.descriptions.pt === "",
    "English adoption changed regional descriptions");
  assert((await loaded("button", "翻译详情至 MX/BR")).disabled, "unsaved English could be translated");
  flushSync(() => saveAccepted());
  await click("button", "翻译详情至 MX/BR");
  await loaded("li", "Glass body");
  assert(aiRequest.operation === "translate" && aiRequest.expected_revision === 2 &&
    Object.keys(aiRequest).length === 3 && editor.descriptions.es === "",
    "translation sent client text, used an older revision or automatically adopted");
  await click("button", "复制候选西语详情");
  assert(copied[1] === "Cuerpo de vidrio\n\nPara el hogar." && editor.descriptions.es === "",
    "regional copy changed product fields");
  await click("button", "采纳西语详情");
  assert(editor.descriptions.es === copied[1] && editor.descriptions.pt === "" &&
    editor.description_translation.es_source_digest === originalDigest &&
    !editor.description_translation.pt_source_digest, "Spanish adoption affected Portuguese or lost source");
  await click(candidateTab, "葡语 BR");
  await click("button", "采纳葡语详情");
  assert(editor.descriptions.pt === "Corpo de vidro\n\nPara uso doméstico." &&
    editor.description_translation.pt_source_digest === originalDigest, "Portuguese adoption lost its source");
  // Switch while both the product editor and AI candidate are on Portuguese.
  const editorTabs = '[aria-label="标题与详情语言"] button';
  await click(editorTabs, "葡语 BR");
  const oldTitle = document.querySelector(".listing-active-title textarea");
  const oldDescription = document.querySelector('textarea[placeholder="填写纯文本商品描述"]');
  oldDescription.focus();
  const beforeModeSwitch = JSON.stringify(editor);
  const beforeAIRequest = aiRequest;
  flushSync(() => setSiteTitleMode("platform_auto"));
  await loaded("legend", "英文商品详情");
  assert(!document.querySelector('[aria-label="标题与详情语言"]') &&
    !document.querySelector('[aria-label="候选文案语言"]') &&
    ![...document.querySelectorAll("button")].some(node => node.textContent.includes("翻译详情至 MX/BR")),
    "auto mode exposed language tabs or regional translation");
  assert(!oldTitle.isConnected && !oldDescription.isConnected, "auto mode reused a regional input");
  assert(document.querySelector(".listing-active-title textarea").value === editor.titles.en &&
    document.querySelector('textarea[placeholder="填写纯文本商品描述"]').value === editor.descriptions.en,
    "auto mode did not bind both fields to English");
  assert(JSON.stringify(editor) === beforeModeSwitch && aiRequest === beforeAIRequest,
    "mode switch changed saved fields or issued an AI request");
  await loaded("label", "候选英文标题");
  await loaded("label", "候选英文详情");
  assert(![...document.querySelectorAll("label")].some(node =>
    /候选西语详情|候选葡语详情/.test(node.textContent)), "auto mode exposed regional AI candidates");
  input(document.querySelector('textarea[placeholder="填写纯文本商品描述"]'), "Auto-mode English draft.");
  const before = JSON.parse(beforeModeSwitch);
  assert(editor.descriptions.en === "Auto-mode English draft." &&
    editor.descriptions.es === before.descriptions.es && editor.descriptions.pt === before.descriptions.pt &&
    JSON.stringify(editor.titles) === JSON.stringify(before.titles),
    "English-only description edit clobbered regional data or titles");
  input(document.querySelector('textarea[placeholder="填写纯文本商品描述"]'), before.descriptions.en);
  flushSync(() => setSiteTitleMode("localized"));
  await click(editorTabs, "葡语 BR");
  assert(document.querySelector(".listing-active-title textarea").value === before.titles.pt &&
    document.querySelector('textarea[placeholder="填写纯文本商品描述"]').value === before.descriptions.pt &&
    JSON.stringify(editor) === beforeModeSwitch, "localized mode failed to restore the regional fields");
  await click(editorTabs, "英语 EN");
  await loaded("button", "重新翻译详情至 MX/BR");
  flushSync(() => saveAccepted());
  input(document.querySelector('textarea[placeholder="填写纯文本商品描述"]'), "Revised English master.");
  await loaded("p", "西语详情可能已过期");
  assert((await loaded("button", "重新翻译详情至 MX/BR")).disabled, "unsaved English revision can be translated");
  flushSync(() => saveAccepted());
  flushSync(() => reloadSaved());
  await loaded("p", "西语详情可能已过期");
  await loaded("p", "葡语详情可能已过期");
  assert(editor.descriptions.es === copied[1] && editor.descriptions.pt === "Corpo de vidro\n\nPara uso doméstico." &&
    editor.description_translation.en_digest === revisedDigest, "editing and reloading English lost region copy");
  aiMode = "wait";
  await click("button", "AI 优化标题");
  await click("button", "取消生成");
  await loaded('[role="status"]', "已取消生成");
  assert(aiCancelled, "desktop cancelled without aborting AI fetch");
  aiMode = "error";
  await click("button", "AI 优化标题");
  await loaded('[role="alert"]', "AI fixture unavailable");
  assert(editor.titles.en === "Refined Glass Pitcher", "error changed adopted title");
  switchCategory("CBT1055");
  const commonColor=(await loaded("label","Color · 必填")).querySelector("select");
  assert(commonColor.options.length>1,"single-product official string suggestions missing");
  input((await loaded("label","Is Dual SIM · 必填 · 家族共用")).querySelector("select"),"242085");
  await loaded("th", "Color（颜色）");
  assert(![...document.querySelectorAll("p")].some(p=>p.textContent.includes("存在特殊官方要求")),"optional dependent field alone produced a warning");
  await click("button", "+ 添加变体");
  await click("button", "+ 添加变体");
  const color1=await loaded("th","Color（颜色）");
  assert(color1.querySelector("input"),"bilingual search missing");
  input(color1.querySelector("input"),"黑色");
  const firstColor=document.querySelector('select[aria-label="变体 1 Color"]');
  assert([...firstColor.options].some(o=>o.textContent.includes("Black（黑色）"))&&!([...firstColor.options].some(o=>o.textContent.includes("White"))),"Chinese option filter failed");
  assert(!commonColor.isConnected,"child spec still editable as a shared attribute");
  assert(![...document.querySelectorAll("th")].some(node=>node.textContent.includes("Main color")),
    "dependent spec rendered as an independent variation axis");
  input(firstColor,"52049");
  input(color1.querySelector("input"),"");
  const secondColor=document.querySelector('select[aria-label="变体 2 Color"]');
  const secondCustom=document.querySelector('input[aria-label="变体 2 Color 自定义值"]');
  input(secondCustom,"黑色");
  await loaded("small","请填写英文规格 · 可保存工作稿，刊登前修正");
  input(secondColor,"52055");
  assert(editor.variants[0].category_attributes[0].value==="Black"&&
    editor.variants[0].category_attributes[0].value_id==="52049"&&
    editor.variants[1].category_attributes[0].value_id==="52055"&&
    working.some(f=>f.id==="IS_DUAL_SIM"&&f.value==="Yes"&&f.value_id==="242085"),
    "official child and parent options did not retain the unmodified English values and IDs");
  input(document.querySelector('input[aria-label="变体 1 Internal memory"]'),"64.0");
  input(document.querySelector('select[aria-label="变体 1 Internal memory unit"]'),"GB");
  await loaded("small","待填写规格值");
  assert(document.querySelector('input[aria-label="变体 2 Internal memory"]').getAttribute("aria-required")==="true",
    "supplemental official required rule was treated as optional");
  input(document.querySelector('input[aria-label="变体 2 Internal memory"]'),"64");
  input(document.querySelector('select[aria-label="变体 2 Internal memory unit"]'),"GB");
  input(secondColor,"52049");
  await loaded("small","重复官方规格组合");
  input(secondCustom," black ");
  await loaded("small","重复官方规格组合");
  input(secondColor,"52055");
  switchCategory("CBT2");
  await loaded("small","历史规格：CBT1055 · COLOR · Black");
  switchCategory("CBT1055");
  await loaded("th","Color（颜色）");
  assert(document.querySelector('select[aria-label="变体 1 Color"]').value==="52049","category switch lost its historical official selection");
  input(document.querySelector('input[aria-label="变体 1 采购成本（¥）"]'),"20.25");
  input(document.querySelector('input[aria-label="变体 1 目标净收益（USD）"]'),"8.50");
  const openPicker=async i=>{flushSync(()=>document.querySelectorAll(".listing-variant-images-button")[i].click());};
  const appendImage=async url=>{
    input(document.querySelector('input[aria-label="追加变体图片 URL"]'),url);
    await click("button","添加链接");
  };
  const img=s=>"https://example.com/"+s+".jpg";
  await openPicker(0);await appendImage(img("black"));await appendImage(img("shared"));await appendImage(img("shared2"));
  await click(".listing-variant-image-picker-head button","完成");
  await openPicker(1);await appendImage(img("white"));await appendImage(img("old"));
  await loaded(".listing-variant-image-picker-head strong","White（白色）");
  await click(".listing-variant-image-picker-head button","复用副图");
  await loaded('[aria-label="复用副图"] option',"Black（黑色） / 64.0 GB");
  assert(![...document.querySelectorAll('[aria-label="复用副图"] option')].some(o=>o.textContent.includes("未命名")),"official-only variants lost display names");
  let overwriteCount=0;
  window.confirm=message=>{assert(message.includes("覆盖 1 个变体已有副图"),"unexpected overwrite confirmation");overwriteCount++;return false;};
  await click("button","应用复用副图");
  assert(overwriteCount===1&&editor.variants[1].image_urls[1]===img("old"),"cancelled reuse removed existing secondaries");
  flushSync(()=>setEditorDisabled(true));
  assert((await loaded("button","应用复用副图")).disabled&&document.querySelector('select[aria-label="变体 1 Color"]').disabled,
    "locked workbench permits reuse or official spec edits");
  flushSync(()=>setEditorDisabled(false));
  window.confirm=()=>true;
  await click("button","应用复用副图");
  assert(JSON.stringify(editor.variants[1].image_urls)===JSON.stringify([img("white"),img("shared"),img("shared2")]),"single reuse changed main or lost secondaries");
  await click(".listing-variant-image-picker-head button","完成");
  await click("button","+ 添加变体");
  input(document.querySelector('select[aria-label="变体 3 Color"]'),"52028");
  await click(".listing-gallery-toolbar button","批量复用副图");
  await loaded('[aria-label="复用副图"] label',"White（白色） / 64 GB");
  await loaded('[aria-label="复用副图"] label',"Blue（蓝色）");
  await click('[aria-label="复用副图"] button',"清空选择");
  assert((await loaded("button","应用复用副图")).disabled,"empty target selection remained actionable");
  await click('[aria-label="复用副图"] button',"全选其他变体");
  await click("button","应用复用副图");
  await loaded('[role="status"]',"目标变体尚未设置有效主图");
  assert(editor.variants[2].image_urls.length===0,"bulk reuse fabricated a target main image");
  await click('[aria-label="复用副图"] button',"关闭");
  await openPicker(2);await appendImage(img("blue"));await click(".listing-variant-image-picker-head button","完成");
  await click(".listing-gallery-toolbar button","批量复用副图");
  await click("button","应用复用副图");
  assert(JSON.stringify(editor.variants[2].image_urls)===JSON.stringify([img("blue"),img("shared"),img("shared2")]),"bulk reuse changed main or lost order");
  await click('[aria-label="复用副图"] button',"关闭");
  assert(validateListingFields(editor)==="","new workbench fields not valid");
  flushSync(()=>saveAccepted());flushSync(()=>reloadSaved());
  assert(editor.variants[0].purchase_cost_cny==="20.25"&&editor.variants[0].target_net_proceeds_usd==="8.50"&&editor.variants[2].image_urls[0]===img("blue"),"variant reload lost new fields or images");
  assert(editor.variants[0].category_attributes.some(f=>f.id==="INTERNAL_MEMORY"&&f.value==="64.0"&&f.unit_id==="GB"),
    "variant reload altered official unit or numeric value");
  flushSync(()=>setEditorValue({...editor,variants:editor.variants.map((v,i)=>i===2?
    {...v,image_urls:Array.from({length:30},(_,n)=>img("cap"+n))}:v)}));
  await openPicker(2);
  input(document.querySelector('input[aria-label="追加变体图片 URL"]'),img("overflow"));
  assert((await loaded("button","添加链接")).disabled,"variant image picker permits the 31st image");
  const published = { id: "MLB1", category_id: "MLB771", description: "Fixture details",
    cbt_category_path: { original: ["Home", "Other"], zh: ["家居", "其他"] } };
  const showPublished = listing => flushSync(() => root.render(React.createElement(PublishedListingEditor, {
    key: listing.id, listing, site: "MLB", logistic: "remote", store: "fixture", members: [listing],
  })));
  showPublished(published);
  await loaded(".published-category-path", "家居 / 其他");
  await loaded(".published-hint", "Home / Other");
  await loaded(".published-category-path", "Jarras");
  showPublished({ ...published, id: "MLB2", cbt_category_path: undefined });
  await loaded(".published-category-path", "Jarras");
  assert(![...document.querySelectorAll(".published-category-path")].some(node => node.textContent.includes("家居")),
    "site-only listing reused another listing's CBT breadcrumb");
  document.body.dataset.result = "pass";
} catch (error) { document.body.dataset.result = "fail"; document.body.append(error.message); }
`;

test("listing editor keeps official CBT data with partial translations and explicitly adopted AI copy", { timeout: 45_000 }, async () => {
  const profile = await mkdtemp(join(tmpdir(), "una-cbt-attrs-"));
  const server = await createServer({
    server: { port: 0 },
    optimizeDeps: { include: ["react", "react-dom", "react-dom/client", "react/jsx-runtime"] },
    plugins: [{
      name: "cbt-attribute-fixture",
      enforce: "pre",
      resolveId(id) {
        if (id === "./auth") return "\0cbt-auth";
        if (id === "/cbt-fixture.js") return "\0cbt-fixture.js";
      },
      load(id) {
        if (id === "\0cbt-fixture.js") return fixture;
        if (id === "\0cbt-auth") return 'const lifetime = new AbortController().signal; export const authLifetime = () => lifetime; export const getAccessToken = async () => "fixture"; export const rejectSession = () => {};';
      },
      configureServer(server) {
        server.middlewares.use(async (req, res, next) => {
          if (req.url !== "/cbt-fixture") return next();
          res.setHeader("Content-Type", "text/html; charset=utf-8");
          res.end(await server.transformIndexHtml(req.url,
            '<!doctype html><html><head></head><body><main></main><script type="module" src="/cbt-fixture.js"></script></body></html>'));
        });
      },
    }],
  });
  try {
    await server.listen();
    const { port } = server.httpServer.address();
    const chrome = process.env.CHROME_BIN ?? (process.platform === "win32"
      ? "C:/Program Files/Google/Chrome/Application/chrome.exe" : "google-chrome");
    const { stdout } = await promisify(execFile)(chrome, [
      "--headless", "--no-sandbox", "--disable-background-networking", `--user-data-dir=${profile}`,
      "--dump-dom", "--virtual-time-budget=6000", `http://127.0.0.1:${port}/cbt-fixture`,
    ], { timeout: 30_000 });
    assert.ok(stdout.includes('data-result="pass"'), stdout.slice(-1800));
  } finally {
    await server.close();
    await rm(profile, { recursive: true, force: true, maxRetries: 5, retryDelay: 200 });
  }
});
