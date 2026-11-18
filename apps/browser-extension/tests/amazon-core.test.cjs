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
  assert.equal(core.parsePrice("$1,299.00"), "1299.00");
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

test("reads only the active ImageBlockATF gallery when the DOM has no thumbnails", () => {
  const fs = require("node:fs");
  const path = require("node:path");
  const source = fs.readFileSync(path.join(__dirname, "image-block-atf.fixture.txt"), "utf8");
  const url = "https://www.amazon.com/dp/B0ABC12345?th=1&psc=1";
  const heroURL = "https://m.media-amazon.com/images/I/MainA._AC_SY395_.jpg";
  const page = (src, script = source) => ({
    querySelector: selector => selector === "#landingImage" ? fakeImage({src}) :
      selector === "#productTitle" ? {textContent: "Synthetic product"} : null,
    querySelectorAll: selector => selector === "#imageBlock_feature_div script" ? [{textContent: script}] : []
  });

  const product = core.parseProduct(page(heroURL), url);
  assert.equal(product.source_url, "https://www.amazon.com/dp/B0ABC12345");
  assert.deepEqual(product.images, [
    {url: "https://m.media-amazon.com/images/I/MainA._AC_SY695_.jpg", type: "main"},
    {url: "https://m.media-amazon.com/images/I/PhotoB._AC_SL1500_.jpg", type: "additional"}
  ]);
  assert.ok(!product.missing_fields.includes("additional_images"));

  const wrongMedia = core.parseProduct(page(heroURL, source.replace('"physicalIdForMedia":"PhotoB"', '"physicalIdForMedia":"WrongColor"')), url);
  assert.equal(wrongMedia.images[1].url, "https://m.media-amazon.com/images/I/PhotoB._AC_SY200_.jpg");

  for (const doc of [page(heroURL, ""), page(heroURL.replace("MainA", "mainA"))]) {
    const partial = core.parseProduct(doc, url);
    assert.equal(partial.images.length, 1);
    assert.ok(partial.missing_fields.includes("additional_images"));
    assert.equal(partial.collection_status, "partial");
  }
});
