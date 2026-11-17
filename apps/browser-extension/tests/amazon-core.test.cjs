const test = require("node:test");
const assert = require("node:assert/strict");
require("../amazon-core.js");
const core = globalThis.UnaAmazonCore;

test("Amazon US product URLs only, strip tracking information", () => {
  assert.equal(core.canonicalAmazonURL("https://www.amazon.com/Product-Name/dp/B0ABC12345?ref_=abc"), "https://www.amazon.com/dp/B0ABC12345");
  assert.equal(core.canonicalAmazonURL("https://amazon.com/gp/product/B0ABC12345?tag=test"), "https://www.amazon.com/dp/B0ABC12345");
  for (const url of ["https://amazon.co.uk/dp/B0ABC12345", "http://www.amazon.com/dp/B0ABC12345", "https://evil.com/dp/B0ABC12345", "https://www.amazon.com.evil.com/dp/B0ABC12345", "https://www.amazon.com/s?k=toy"]) {
    assert.equal(core.canonicalAmazonURL(url), null);
  }
});
test("Price preserves visible precise dollars but never invents ranges", () => {
  assert.equal(core.parsePrice("$12.99"), "12.99");
  assert.equal(core.parsePrice("$1,299.00"), "1299.00");
  assert.equal(core.parsePrice("$9.50 - $15.00"), null);
  assert.equal(core.parsePrice("Not available"), null);
  assert.equal(core.parsePrice("19.99"), null);
});
test("Blocked pages are not saved as valid products", () => {
  const doc = {
    title: "Robot Check",
    querySelector: () => null
  };
  assert.throws(() => core.parseProduct(doc, "https://www.amazon.com/dp/B0ABC12345"), /验证或登录/);
});
test("Partial product remains partial when description, images, bullets or price are missing", () => {
  const doc = {
    title: "Example Item",
    querySelector: selector => selector === "#productTitle" ? { textContent: "Simple Example" } : null,
    querySelectorAll: () => []
  };
  const input = core.parseProduct(doc, "https://www.amazon.com/dp/B0ABC12345");
  assert.equal(input.collection_status, "partial");
  assert.deepEqual(input.missing_fields, ["price","images","bullet_points","description"]);
  assert.equal(input.source_site, "US");
  assert.equal(input.source_url, "https://www.amazon.com/dp/B0ABC12345");
});
test("Manifest is narrowly scoped to recognized hosts", () => {
  const manifest = require("../manifest.json");
  assert.equal(manifest.manifest_version, 3);
  assert.equal(manifest.permissions.join(","), "storage");
  assert.ok(!manifest.host_permissions.some(p => /<all_urls>|\*:/.test(p)));
  assert.deepEqual(manifest.content_scripts[0].matches, ["https://www.amazon.com/*","https://amazon.com/*"]);
});
