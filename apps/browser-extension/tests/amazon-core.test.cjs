const test = require("node:test");
const assert = require("node:assert/strict");
require("../amazon-core.js");
const core = globalThis.UnaAmazonCore;
test("reject unrelated hosts and normalize Amazon ASIN", () => {
  assert.equal(core.canonicalAmazonURL("https://www.amazon.com/name/dp/B0ABC12345?ref=xyz"), "https://www.amazon.com/dp/B0ABC12345");
  assert.equal(core.canonicalAmazonURL("https://amazon.com/gp/product/B0ABC12345"), "https://www.amazon.com/dp/B0ABC12345");
  assert.equal(core.canonicalAmazonURL("https://amazon.co.uk/dp/B0ABC12345"), null);
});
test("parse exact prices, not price ranges", () => {
  assert.equal(core.parsePrice("$18.59"), "18.59");
  assert.equal(core.parsePrice("$18"), "18.00");
  assert.equal(core.parsePrice("$1,234.56"), "1234.56");
  for (const raw of ["$1,23", "$18.590", "$18.", "$.59", "$1,234.5", "$18–20", "USD $18.59", "$18.59abc"]) {
    assert.equal(core.parsePrice(raw), null, raw);
  }
  assert.equal(core.parsePrice("$9.50 - $15.00"), null);
});
function fakeImage(attrs = {}) {
  return { getAttribute: k => attrs[k] || null };
}
function fakePage() {
  const hero = fakeImage({
    "data-old-hires":"https://m.media-amazon.com/images/I/61ABC._AC_SL1500_.jpg",
    "data-a-dynamic-image":JSON.stringify({
      "https://m.media-amazon.com/images/I/61ABC._AC_SY300_.jpg":[300,300],
      "https://m.media-amazon.com/images/I/61ABC._AC_SL900_.jpg":[900,900]
    })
  });
  const imageThumb = fakeImage({ src:"https://m.media-amazon.com/images/I/71SECOND._AC_US40_.jpg" });
  const videoThumb = fakeImage({ src:"https://m.media-amazon.com/images/I/VIDEO._AC_US40_.jpg" });
  const photos = [
    { className:"imageThumbnail", querySelector: selector => selector==="img" ? imageThumb : null },
    { className:"videoThumbnail", querySelector: selector => selector==="img" ? videoThumb : null }
  ];
  return {
    querySelector: s => s==="#landingImage" ? hero : null,
    querySelectorAll: s => s==="#altImages li, #imageBlockThumbs li" ? photos : []
  };
}
test("dedupe Amazon image sizes, upgrade thumbnails, exclude video thumbnails", () => {
  const images=core.collectImages(fakePage());
  assert.equal(images.length,2);
  assert.deepEqual(images.map(x=>x.type),["main","additional"]);
  assert.equal(images[0].url,"https://m.media-amazon.com/images/I/61ABC._AC_SL1500_.jpg");
  assert.equal(images[1].url,"https://m.media-amazon.com/images/I/71SECOND.jpg");
  assert.ok(!images.some(x=>x.url.includes("VIDEO")));
});
test("only permitted Amazon image domains",()=>{
  assert.equal(core.imageInfo("https://m.media-amazon.com/images/I/video-poster.jpg"), null);
  assert.equal(core.imageInfo("https://m.media-amazon.com/images/I/ABC._SS40_PKdp-play-icon-overlay__.jpg"), null);
  assert.equal(core.imageInfo("https://evil.example/images/I/61ABC._AC_SL1500_.jpg"),null);
  assert.equal(core.imageInfo("https://m.media-amazon.com/images/I/61ABC._AC_SL200_.jpg").key, "61ABC");
});
test("blocked and missing product information never pretend complete",()=>{
  const blocked={title:"Robot Check",querySelector:()=>null};
  assert.throws(()=>core.parseProduct(blocked,"https://www.amazon.com/dp/B0ABC12345"),/验证码/);
  const noInfo={title:"Example",querySelector:s=>s==="#productTitle"?{textContent:"A product"}:s===".a-price .a-offscreen"?{textContent:"$12.99"}:null,querySelectorAll:()=>[]};
  const result=core.parseProduct(noInfo,"https://www.amazon.com/dp/B0ABC12345");
  assert.equal(result.collection_status,"partial");
  assert.deepEqual(result.missing_fields,["price","images","bullet_points","description"]);
});
test("search controls mount inside the matched card, never as siblings of its heading",()=>{
  const fs=require("node:fs");
  const content=fs.readFileSync(require("node:path").join(__dirname,"../content.js"),"utf8");
  assert.match(content,/mount\.appendChild\(createControl\(url\)\)/);
  assert.doesNotMatch(content,/insertAdjacentElement\(["']afterend/);
});

const fs = require("node:fs");
const path = require("node:path");
const source = fs.readFileSync(path.join(__dirname, "image-block-atf.fixture.txt"), "utf8");
const url = "https://www.amazon.com/dp/B0ABC12345?th=1&psc=1";
const heroURL = "https://m.media-amazon.com/images/I/MainA._AC_SY395_.jpg";
const page = (heroAttributes, script = source) => ({
  querySelector: selector => selector === "#landingImage" ? fakeImage(heroAttributes) :
    selector === "#productTitle" ? {textContent: "Synthetic product"} : null,
  querySelectorAll: selector => selector === "#imageBlock_feature_div script" ? [{textContent: script}] : []
});

test("reads only the active ImageBlockATF gallery when the DOM has no thumbnails", () => {
  const product = core.parseProduct(page({src: heroURL}), url);
  assert.equal(product.source_url, "https://www.amazon.com/dp/B0ABC12345");
  assert.deepEqual(product.images, [
    {url: "https://m.media-amazon.com/images/I/MainA._AC_SY695_.jpg", type: "main"},
    {url: "https://m.media-amazon.com/images/I/PhotoB._AC_SL1500_.jpg", type: "additional"}
  ]);
  assert.ok(!product.missing_fields.includes("additional_images"));

  const wrongMedia = core.parseProduct(page({src: heroURL}, source.replace('"physicalIdForMedia":"PhotoB"', '"physicalIdForMedia":"WrongColor"')), url);
  assert.equal(wrongMedia.images[1].url, "https://m.media-amazon.com/images/I/PhotoB._AC_SY200_.jpg");

  const partial = core.parseProduct(page({src: heroURL}, ""), url);
  assert.equal(partial.images.length, 1);
  assert.ok(partial.missing_fields.includes("additional_images"));
});

test("uses the same hero element's hires or dynamic image for a placeholder src", () => {
  const src = "data:image/gif;base64,R0lGODlh";
  for (const attrs of [
    {"data-old-hires": heroURL},
    {"data-image-hires": heroURL},
    {"data-a-dynamic-image": JSON.stringify({
      [heroURL]: [395,395],
      [heroURL.replace("395", "695")]: [695,695]
    })}
  ]) {
    const product = core.parseProduct(page({src, ...attrs}), url);
    assert.deepEqual(product.images.map(image => image.type), ["main", "additional"]);
    assert.ok(!product.missing_fields.includes("additional_images"));
  }
});

test("rejects the gallery when the lazy hero identity is different or ambiguous", () => {
  const src = "data:image/gif;base64,R0lGODlh";
  const otherHero = heroURL.replace("MainA", "mainA");
  const product = core.parseProduct(page({src, "data-old-hires": otherHero}), url);
  assert.deepEqual(product.images, [{url: otherHero, type: "main"}]);
  assert.ok(product.missing_fields.includes("additional_images"));
  assert.equal(product.collection_status, "partial");

  const ambiguous = core.parseProduct(page({src, "data-a-dynamic-image": JSON.stringify({
    [heroURL]: [395,395], [otherHero]: [395,395]
  })}), url);
  assert.equal(ambiguous.images, undefined);
  assert.ok(ambiguous.missing_fields.includes("images"));
});

test("omits images when a thumbnail is present but the hero identity is ambiguous", () => {
  const doc = page({
    src: "data:image/gif;base64,R0lGODlh",
    "data-a-dynamic-image": JSON.stringify({
      [heroURL]: [395,395],
      [heroURL.replace("MainA", "OtherHero")]: [395,395]
    })
  });
  const query = doc.querySelectorAll;
  const thumb = fakeImage({src: "https://m.media-amazon.com/images/I/Thumbnail._AC_US40_.jpg"});
  doc.querySelectorAll = selector => selector === "#altImages li, #imageBlockThumbs li"
    ? [{className: "imageThumbnail", querySelector: name => name === "img" ? thumb : null}]
    : query(selector);

  const product = core.parseProduct(doc, url);
  assert.equal(Object.hasOwn(product, "images"), false);
  assert.ok(product.missing_fields.includes("images"));
  assert.equal(product.collection_status, "partial");
});

test("a seven-photo ATF fixture preserves one main and six ordered additional photos", () => {
  const gallery = fs.readFileSync(path.join(__dirname, "image-block-atf-seven.fixture.txt"), "utf8");
  const images = core.collectImages(page({src: heroURL}, gallery));
  assert.equal(images.length, 7);
  assert.deepEqual(images.map(image => image.type), ["main", ...Array(6).fill("additional")]);
  assert.deepEqual(images.map(image => core.imageInfo(image.url)?.key),
    ["MainA", "Photo1", "Photo2", "Photo3", "Photo4", "Photo5", "Photo6"]);
  assert.equal(new Set(images.map(image => core.imageInfo(image.url)?.key)).size, 7);
});

test("storage isolation failure blocks session operations and network requests", async () => {
  const vm = require("node:vm");
  const worker = fs.readFileSync(path.join(__dirname, "../background.js"), "utf8");
  const popup = {url: "chrome-extension://fixture/popup.html"};
  const amazon = {tab: {}, url};
  for (const setAccessLevel of [undefined, async () => { throw Error("unavailable"); }]) {
    let listener;
    const calls = [];
    const chrome = {
      storage: {local: {
        setAccessLevel,
        get: () => calls.push("get"),
        set: () => calls.push("set"),
        remove: () => calls.push("remove")
      }},
      runtime: {onMessage: {addListener: fn => { listener = fn; }}}
    };
    vm.runInNewContext(worker, {chrome, fetch: () => calls.push("fetch")});
    for (const [type, sender] of [["SESSION_STATUS", popup], ["SIGN_IN", popup], ["SIGN_OUT", popup],
      ["FETCH_AMAZON_HTML", amazon], ["SAVE_COLLECTION", amazon]]) {
      const response = await new Promise(resolve => listener({type, url}, sender, resolve));
      assert.equal(response.ok, false);
      assert.match(response.error, /无法保护扩展会话存储/);
    }
    assert.deepEqual(calls, []);
  }
});
