"use strict";
const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const source = fs.readFileSync(path.join(__dirname, "../content.js"), "utf8");
const css = fs.readFileSync(path.join(__dirname, "../content.css"), "utf8");

function fixture({ hasTitle = true, hasSection = true, hasCard = false } = {}) {
  const controls = [];
  let observer;
  let timer;
  const url = "https://www.amazon.com/dp/B000000001";
  const location = { href: url };
  function element() {
    return {
      children: [], attrs: {}, dataset: {},
      setAttribute(key, value) { this.attrs[key] = value; },
      getAttribute(key) { return this.attrs[key] ?? null; },
      append(...nodes) { nodes.forEach(node => this.appendChild(node)); },
      appendChild(node) { this.children.push(node); node.parentElement = this; },
      remove() { const parent = this.parentElement; parent.children.splice(parent.children.indexOf(this), 1); this.parentElement = null; },
      addEventListener() {}
    };
  }
  const holder = element();
  const card = element();
  const root = element();
  const title = { closest: selector => selector === "#titleSection" && hasSection ? holder : null };
  card.querySelector = selector => selector === "h2" ? {} :
    selector === ".una-collector-button" ? card.children.find(node => node.className?.includes("una-collector-card")) : null;
  card.querySelectorAll = () => [{ href: url, closest: () => null }];
  const createElement = () => { const node = element(); controls.push(node); return node; };
  const document = {
    body: {},
    querySelector: selector => selector === "#productTitle" && hasTitle ? title :
      selector === "#titleSection" && hasSection ? holder : selector === "#dp-container" ? root : null,
    querySelectorAll: selector => selector === ".una-collector-detail" ?
      controls.filter(node => node.parentElement && node.className?.split(" ").includes("una-collector-detail")) :
      selector.includes("s-search-result") && hasCard ? [card] : [],
    createElement
  };
  const context = {
    document, location,
    UnaAmazonCore: { canonicalAmazonURL: value => value },
    MutationObserver: class { constructor(callback) { observer = callback; } observe() {} },
    setTimeout(callback) { timer = callback; }
  };
  vm.runInNewContext(source, context);
  return { holder, card, root, location, rescan() { observer(); timer(); } };
}

test("detail control is scoped to the title section and does not duplicate", () => {
  const f = fixture();
  assert.equal(f.holder.children.length, 1);
  assert.match(f.holder.children[0].className, /una-collector-detail/);
  const first = f.holder.children[0];
  f.rescan();
  assert.equal(f.holder.children.length, 1);
  assert.equal(f.holder.children[0], first);
  f.location.href = "https://www.amazon.com/dp/B000000002";
  f.rescan();
  assert.equal(f.holder.children.length, 1);
  assert.notEqual(f.holder.children[0], first);
  assert.equal(f.holder.children[0].getAttribute("data-una-product"), f.location.href);
});

test("no unsafe fallback into the detail root", () => {
  const f = fixture({ hasSection: false });
  assert.equal(f.holder.children.length, 0);
  assert.equal(f.root.children.length, 0);
  f.rescan();
  assert.equal(f.root.children.length, 0);
});

test("search card retains its own control and never duplicates", () => {
  const f = fixture({ hasTitle: false, hasCard: true });
  assert.equal(f.card.children.length, 1);
  assert.match(f.card.children[0].className, /una-collector-card/);
  f.rescan();
  assert.equal(f.card.children.length, 1);
});

test("float clearing and full-row layout are scoped to search cards", () => {
  const rule = selector => css.split(selector + " {")[1].split("}")[0];
  const common = rule(".una-collector-wrapper");
  const card = rule(".una-collector-wrapper.una-collector-card");
  const detail = rule(".una-collector-wrapper.una-collector-detail");
  assert.doesNotMatch(common, /(?:clear|flex|grid-column|align-self)\s*:/);
  assert.match(card, /clear:\s*both/);
  assert.match(card, /flex:\s*0 0 100%/);
  assert.match(detail, /clear:\s*none/);
  assert.doesNotMatch(detail, /(?:flex|grid-column|align-self)\s*:/);
});
