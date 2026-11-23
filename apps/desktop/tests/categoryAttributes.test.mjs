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

const fields = [
  { id: "BRAND", name: "Brand", value_type: "string", tags: { required: true } },
  { id: "COUNT", name: "Count", value_type: "number", tags: { new_required: true } },
  { id: "CAPACITY", name: "Capacity", value_type: "number_unit", tags: { required: true },
    allowed_units: [{ id: "l", name: "L" }] },
  { id: "MATERIAL", name: "Material", value_type: "list", values: [{ id: 200, name: "Glass" }] },
  { id: "EMPTY_LIST", name: "Empty List", value_type: "list" },
];
window.fetch = (path, options) => {
  const url = new URL(path, location.href);
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
    technical_requirements: id === "CBT2" ? ["STYLE"] : undefined,
    technical_requirement_names: id === "CBT2" ? { STYLE: "Style" } : undefined,
    path: [{ id: "CBT1", name: "Home" }, { id, name: id === "CBT1" ? "Jugs" : "Other" }],
    attributes: id === "CBT1" ? fields : id === "CBT2" ? [
      fields[0],
      { ...fields[2], allowed_units: [{ id: "ml", name: "mL" }] },
      { ...fields[3], values: [{ id: 300, name: "Plastic" }] },
    ] : [],
  }));
};
let editor, working;
function Fixture() {
  const [value, setValue] = useState(() => initialListingFields({
    category_id: "CBT1", category_path: "Home / Jugs",
    titles: { en: "Sample Jug", es: "", pt: "" },
  }));
  editor = value;
  working = value.category_attributes ?? [];
  return React.createElement(ListingFieldsEditor, { value, onChange: setValue });
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
  await loaded("p", "款式 · Style");
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

test("listing editor clears CBT category on English title edits and preserves official attribute history", { timeout: 45_000 }, async () => {
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
