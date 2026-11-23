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
import CategoryAttributesForm from "/src/CategoryAttributesForm.tsx";
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
  const id = new URL(path, location.href).searchParams.get("category_id");
  if (id === "CBT4") return Promise.resolve(Response.json({ error: "upstream failed" }, { status: 502 }));
  if (!options.headers.Authorization) throw Error("missing bearer");
  return Promise.resolve(Response.json({
    category_id: id, technical_specs_verified: true,
    translations_zh: id === "CBT2" ? { Home: "家居", Other: "其他", Brand: "品牌" } : undefined,
    path: [{ id: "CBT1", name: "Home" }, { id, name: id === "CBT1" ? "Jugs" : "Other" }],
    attributes: id === "CBT1" ? fields : id === "CBT2" ? [
      fields[0],
      { ...fields[2], allowed_units: [{ id: "ml", name: "mL" }] },
      { ...fields[3], values: [{ id: 300, name: "Plastic" }] },
    ] : [],
  }));
};
let switchCategory, working;
function Fixture() {
  const [category, setCategory] = useState("CBT1");
  const [values, setValues] = useState([]);
  switchCategory = id => flushSync(() => setCategory(id));
  working = values;
  return React.createElement(CategoryAttributesForm, { category, values, onChange: setValues, disabled: false });
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
  const setter = Object.getOwnPropertyDescriptor(node instanceof HTMLSelectElement
    ? HTMLSelectElement.prototype : HTMLInputElement.prototype, "value").set;
  setter.call(node, value);
  node.dispatchEvent(new Event(node instanceof HTMLSelectElement ? "change" : "input", { bubbles: true }));
});
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
  switchCategory("CBT2");
  await loaded("small", "历史参考（CBT1）：Sample Brand");
  const newBrand = (await loaded("label", "品牌 · Brand")).querySelector("input");
  await loaded("p", "家居 / 其他");
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
  await loaded("summary", "当前未采用的已填属性");
  switchCategory("CBT3");
  await loaded("p", "官方未返回可用属性");
  switchCategory("CBT4");
  await loaded('[role="alert"]', "仍可保存工作稿");
  document.body.dataset.result = "pass";
} catch (error) { document.body.dataset.result = "fail"; document.body.append(error.message); }
`;

test("official CBT fields keep compatible history and allow draft saves without translation", { timeout: 45_000 }, async () => {
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
        if (id === "\0cbt-auth") return 'export const getAccessToken = async () => "fixture"; export const rejectSession = () => {};';
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
