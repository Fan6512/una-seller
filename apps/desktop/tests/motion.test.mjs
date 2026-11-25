import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { test } from "node:test";
const read = file => readFileSync(new URL("../src/" + file, import.meta.url), "utf8");
test("motion tokens and reduced motion are shared and loaded after workbench styles", () => {
  const css = read("motion.css");
  assert.match(css, /--motion-standard:\s*160ms/);
  assert.match(css, /--motion-ease:\s*cubic-bezier/);
  assert.match(css, /prefers-reduced-motion:\s*reduce[\s\S]*animation:\s*none\s*!important[\s\S]*transition:\s*none\s*!important/);
  assert.ok(read("main.tsx").indexOf('import "./motion.css"') > read("main.tsx").indexOf('import "./workbench.css"'));
  for (const file of ["styles.css", "workbench.css", "orders.css", "publishedListings.css", "accountSettings.css", "titleModeSettings.css"]) {
    assert.doesNotMatch(read(file), /\b(?:animation|transition)\s*:/, file);
  }
});
test("no layout transitions, blanket compositor hints, or animated table rows", () => {
  const css = read("motion.css").replace(/\/\*[\s\S]*?\*\//g, "");
  const allowed = new Set(["background-color", "border-color", "color", "opacity", "transform"]);
  for (const [, properties] of css.matchAll(/transition-property:\s*([^;]+);/g)) {
    for (const property of properties.split(",").map(s => s.trim())) assert.ok(allowed.has(property), property);
  }
  assert.doesNotMatch(css, /will-change\s*:|transition:\s*all\b/);
  assert.doesNotMatch(css, /(?:tbody\s+tr|product-card-grid\s+article)[^{]*\{[^}]*animation\s*:/);
});
test("numeric cues keep exact text and cancel on reduced motion or unmount", () => {
  const source = read("motion.tsx");
  assert.match(source, /\{value\}<\/span>/);
  assert.match(source, /running\.current\?\.cancel\(\)/);
  assert.match(source, /prefers-reduced-motion: reduce/);
  assert.doesNotMatch(source, /requestAnimationFrame|setInterval|setTimeout/);
});
