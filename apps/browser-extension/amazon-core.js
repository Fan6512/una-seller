// Only structured, publicly visible product information is extracted.
(() => {
  "use strict";
  const HOSTS = new Set(["amazon.com", "www.amazon.com"]);
  const IMAGE_HOSTS = new Set([
    "m.media-amazon.com", "images-na.ssl-images-amazon.com",
    "images-eu.ssl-images-amazon.com", "images-fe.ssl-images-amazon.com"
  ]);

  function canonicalAmazonURL(raw) {
    try {
      const url = new URL(raw);
      if (url.protocol !== "https:" || !HOSTS.has(url.hostname) || url.port || url.username || url.password) return null;
      const match = url.pathname.match(/\/(?:dp|gp\/product)\/([a-z0-9]{10})(?:\/|$)/i);
      return match ? "https://www.amazon.com/dp/" + match[1].toUpperCase() : null;
    } catch { return null; }
  }
  function parsePrice(raw) {
    if (typeof raw !== "string") return null;
    const match = raw.trim().match(/^\$\s*(\d{1,3}(?:,\d{3})+|\d+)(\.\d{2})?$/);
    return match ? match[1].replace(/,/g, "") + (match[2] || ".00") : null;
  }
  function pageState(doc) {
    const title = doc.title || "";
    if (doc.querySelector('form[action*="validateCaptcha"], input[id*="captcha"], input[name="amzn"]') || /robot check|captcha/i.test(title)) return "captcha";
    if (/sign in|authentication required/i.test(title) && !doc.querySelector("#productTitle")) return "blocked";
    return "normal";
  }
  const textOf = node => (node?.textContent || "").replace(/\s+/g, " ").trim();
  // Amazon image URLs contain presentation variants like ._AC_SL300_.jpg.
  // Remove the transform for identity comparisons without inventing separate photos.
  function imageInfo(raw) {
    try {
      if (!raw || typeof raw !== "string") return null;
      const url = new URL(raw);
      if (url.protocol !== "https:" || !IMAGE_HOSTS.has(url.hostname) || url.username || url.password || url.port) return null;
      const match = url.pathname.match(/^\/images\/I\/([^/]+?)\.(jpe?g|png|webp)$/i);
      if (!match) return null;
      const base = match[1].replace(/\._[A-Za-z0-9,+_-]+$/i, "");
      if (!base || /(?:play[-_]?(?:button|icon)|video[-_]?(?:icon|thumb|poster)|sprite)/i.test(match[1])) return null;
      const ext = match[2].toLowerCase();
      const original = url.origin + "/images/I/" + base + "." + ext;
      const transformed = base !== match[1];
      return { key: base, original, raw: url.origin + url.pathname, transformed };
    } catch { return null; }
  }
  function collectImages(doc) {
    const selected = [];
    const index = new Map();
    function add(raw, score, main = false, fromThumbnail = false) {
      const info = imageInfo(raw);
      if (!info) return;
      const url = fromThumbnail && info.transformed ? info.original : info.raw;
      const candidate = { url, score, main };
      const oldIndex = index.get(info.key);
      if (oldIndex !== undefined) {
        // Preserve a verified main flag regardless of which rendition wins.
        const previous = selected[oldIndex];
        if (candidate.score > previous.score) selected[oldIndex] = { ...candidate, main: previous.main || main };
        else previous.main ||= main;
      } else if (selected.length < 30) {
        index.set(info.key, selected.length);
        selected.push(candidate);
      }
    }
    function dynamicImages(image) {
      try { return JSON.parse(image.getAttribute("data-a-dynamic-image") || "{}") || {}; }
      catch { return {}; }
    }
    function addDynamic(image, main, expectedKey) {
      const matches = url => {
        const key = imageInfo(url)?.key;
        return key && (!expectedKey || key === expectedKey);
      };
      const hiRes = [image.getAttribute("data-old-hires"), image.getAttribute("data-image-hires")].find(matches);
      if (hiRes) add(hiRes, 20000000, main);
      const dynamic = Object.entries(dynamicImages(image)).filter(([url]) => matches(url));
      for (const [url, dims] of dynamic) {
        const area = Array.isArray(dims) && dims.length >= 2 ? Number(dims[0]) * Number(dims[1]) : 0;
        add(url, Number.isFinite(area) ? area : 0, main);
      }
      if (!hiRes && !dynamic.length) add(image.getAttribute("src"), 0, main, true);
    }
    const hero = doc.querySelector("#landingImage") || doc.querySelector("#imgTagWrapperId img") ||
      doc.querySelector("#mainImageContainer img[data-a-dynamic-image]");
    const videoMarker = 'video, [data-video], [data-action*="video" i], [class*="videoThumbnail" i], [class*="videoBlockIngress" i], [class*="playIcon" i]';
    if (hero && !hero.closest?.(videoMarker)) {
      // The initial ImageBlockATF gallery is available before thumbnail nodes are populated.
      let heroKey = imageInfo(hero.getAttribute("src"))?.key ||
        imageInfo(hero.getAttribute("data-old-hires"))?.key ||
        imageInfo(hero.getAttribute("data-image-hires"))?.key;
      if (!heroKey) {
        const keys = new Set(Object.keys(dynamicImages(hero)).map(url => imageInfo(url)?.key).filter(Boolean));
        if (keys.size === 1) [heroKey] = keys;
      }
      if (heroKey) {
        for (const script of doc.querySelectorAll("#imageBlock_feature_div script")) {
          const source = script.textContent || "";
          if (source.length > 128 * 1024 || !source.includes('register("ImageBlockATF"')) continue;
          const match = source.match(/'colorImages'\s*:\s*\{\s*'initial'\s*:\s*A\.\$\.parseJSON\('((?:\\.|[^'\\])*)'\)/);
          if (!match) continue;
          let gallery;
          try { gallery = JSON.parse(match[1].replace(/\\(['\\])/g, "$1")); } catch { continue; }
          if (!Array.isArray(gallery) || gallery[0]?.variant !== "MAIN" ||
              (imageInfo(gallery[0].large)?.key || gallery[0].physicalIdForMedia) !== heroKey) continue;

          for (const photo of gallery.slice(0, 30)) {
            if (typeof photo?.variant !== "string" || !/^(MAIN|PT\d{2})$/.test(photo.variant) ||
                photo.video || photo.videoUrl ||
                photo.videoId || photo.isVideo || /video/i.test(photo.mediaType || "")) continue;
            // Some pages use an unsupported CDN path for large; main still has sized source URLs.
            const key = imageInfo(photo.large)?.key || photo.physicalIdForMedia;
            const main = photo.variant === "MAIN" && key === heroKey;
            if (!key || (photo.variant === "MAIN" && !main)) continue;
            add(photo.large, 0, main);
            for (const [url, dims] of Object.entries(photo.main || {})) {
              if (imageInfo(url)?.key !== key || !Array.isArray(dims)) continue;
              const area = dims[0] * dims[1];
              if (Number.isFinite(area)) add(url, area, main);
            }
            if (imageInfo(photo.hiRes)?.key === key) add(photo.hiRes, 20000000, main);
          }
          if (selected[0]?.main && selected.length > 1 && selected.filter(item => item.main).length === 1) {
            const hiRes = hero.getAttribute("data-old-hires") || hero.getAttribute("data-image-hires");
            if (imageInfo(hiRes)?.key === heroKey) add(hiRes, 20000000, true);
            return selected.map(item => ({ url: item.url, type: item.main ? "main" : "additional" }));
          }
          selected.length = 0;
          index.clear();
          break;
        }
      }
      if (heroKey) addDynamic(hero, true, heroKey);
    }
    // Thumbnails are only identifiers for photos. Never keep low-res thumbnail URLs.
    for (const item of doc.querySelectorAll("#altImages li, #imageBlockThumbs li")) {
      if (/video|play|movie/i.test(String(item.className || "")) ||
          item.querySelector(videoMarker)) continue;
      const image = item.querySelector("img");
      if (!image || /video|play/i.test(image.getAttribute("alt") || "")) continue;
      addDynamic(image, false);
    }
    // Some page variants expose thumbnail images without an li wrapper.
    if (!doc.querySelector("#altImages li, #imageBlockThumbs li")) {
      for (const image of doc.querySelectorAll("#altImages img, #imageBlockThumbs img")) {
        if (image.closest(videoMarker) || /video|play/i.test(image.getAttribute("alt") || "")) continue;
        addDynamic(image, false);
      }
    }
    if (selected.filter(item => item.main).length !== 1) return [];
    return selected.map(item => ({ url: item.url, type: item.main ? "main" : "additional" }));
  }
  function parseProduct(doc, rawURL) {
    const sourceURL = canonicalAmazonURL(rawURL);
    if (!sourceURL) throw new Error("不是受支持的 Amazon.com 商品链接");
    if (pageState(doc) !== "normal") throw new Error("页面需要验证码或登录，请手动解决后再试");
    const title = textOf(doc.querySelector("#productTitle") || doc.querySelector("#title #productTitle") || doc.querySelector("#btAsinTitle"));
    if (!title) throw new Error("商品详情还没加载完整，未找到标题");

    const missing = [], warnings = [];
    const product = {
      source_platform: "amazon", source_site: "US", source_url: sourceURL,
      title, collection_status: "partial", page_state: "normal",
      warnings, missing_fields: missing
    };
    const priceNode = doc.querySelector("#corePrice_feature_div .a-price .a-offscreen") ||
      doc.querySelector("#corePriceDisplay_desktop_feature_div .a-price .a-offscreen") ||
      doc.querySelector("#priceblock_ourprice") || doc.querySelector("#priceblock_dealprice");
    const price = parsePrice(textOf(priceNode));
    if (price) { product.price = price; product.currency = "USD"; }
    else missing.push("price");

    const images = collectImages(doc);
    if (images.length) product.images = images;
    else missing.push("images");
    if (images.length === 1) missing.push("additional_images");

    const bullets = [...doc.querySelectorAll("#feature-bullets li .a-list-item, .product-facts-title + ul li")]
      .map(textOf).filter(v => v && !/^(?:See more|About this item)$/i.test(v)).slice(0, 50);
    if (bullets.length) product.bullet_points = bullets;
    else missing.push("bullet_points");

    const description = textOf(doc.querySelector("#productDescription_feature_div #productDescription") ||
      doc.querySelector("#productDescription"));
    if (description) product.description = description;
    else missing.push("description");

    const category = [...doc.querySelectorAll("#wayfinding-breadcrumbs_feature_div a")].map(textOf).filter(Boolean).join(" > ");
    if (category) product.category = category;
    if (doc.querySelector('#twister, [id^="variation_"]')) {
      missing.push("variants_not_exhaustive");
      warnings.push("该商品存在规格选项，完整变体关系尚未采集");
    }
    if (missing.length) warnings.push("部分商品信息本次未获取");
    product.collection_status = missing.length ? "partial" : "complete";
    return product;
  }
  globalThis.UnaAmazonCore = Object.freeze({ canonicalAmazonURL, parsePrice, parseProduct, collectImages, imageInfo });
})();