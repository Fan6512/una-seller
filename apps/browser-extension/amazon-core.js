// Isolated content-script helper; no credentials, network requests or third-party libraries.
(() => {
  "use strict";
  const AMAZON_HOSTS = new Set(["amazon.com", "www.amazon.com"]);

  function canonicalAmazonURL(raw) {
    try {
      const parsed = new URL(raw);
      if (parsed.protocol !== "https:" || !AMAZON_HOSTS.has(parsed.hostname) || parsed.port || parsed.username || parsed.password) return null;
      const match = parsed.pathname.match(/\/(?:dp|gp\/product)\/([a-z0-9]{10})(?:\/|$)/i);
      return match ? "https://www.amazon.com/dp/" + match[1].toUpperCase() : null;
    } catch { return null; }
  }

  function parsePrice(raw) {
    if (typeof raw !== "string" || !raw.includes("$")) return null;
    // Never invent a decimal point or turn a range into one price.
    if (/\$\s*[\d,.]+\s*[-–]\s*\$/.test(raw)) return null;
    const match = raw.match(/\$\s*((?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d{2})?)/);
    if (!match) return null;
    const value = match[1].replace(/,/g, "");
    return /^\d+(?:\.\d{2})?$/.test(value) ? value : null;
  }

  function pageState(doc) {
    const pageTitle = doc.title || "";
    if (doc.querySelector('form[action*="validateCaptcha"], input[name="amzn"]') ||
        /robot check|captcha/i.test(pageTitle)) return "captcha";
    if (/sign in|authentication required/i.test(pageTitle) && !doc.querySelector("#productTitle")) return "blocked";
    return "normal";
  }

  function textOf(node) { return (node?.textContent || "").replace(/\s+/g, " ").trim(); }

  function parseProduct(doc, rawURL) {
    const sourceURL = canonicalAmazonURL(rawURL);
    if (!sourceURL) throw new Error("不是受支持的 Amazon.com 商品链接");
    const state = pageState(doc);
    if (state !== "normal") throw new Error("商品页面要求验证或登录，请手动处理后重试");
    const title = textOf(doc.querySelector("#productTitle"));
    if (!title) throw new Error("商品详情未完整加载，无法确认标题");

    const missing = [], warnings = [];
    const result = {
      source_platform: "amazon", source_site: "US", source_url: sourceURL,
      title, collection_status: "partial", page_state: "normal",
      missing_fields: missing, warnings
    };
    const priceText = textOf(doc.querySelector("#corePrice_feature_div .a-price .a-offscreen") ||
      doc.querySelector("#corePriceDisplay_desktop_feature_div .a-price .a-offscreen") ||
      doc.querySelector(".a-price .a-offscreen"));
    const price = parsePrice(priceText);
    if (price) {
      result.price = price;
      result.currency = "USD"; // Only accepted when an actual US-dollar price is shown.
    } else missing.push("price");

    const imageNodes = [doc.querySelector("#landingImage"), doc.querySelector("#imgTagWrapperId img")].filter(Boolean);
    const images = [];
    const seen = new Set();
    function addImage(raw, type) {
      if (!raw || typeof raw !== "string") return;
      try {
        const u = new URL(raw, sourceURL);
        if (u.protocol !== "https:" || u.username || u.password || u.port || u.href.length > 2000 || seen.has(u.href)) return;
        seen.add(u.href);
        if (images.length < 60) images.push({ url: u.href, type: images.length ? "additional" : type });
      } catch { /* Broken source URL is not an image. */ }
    }
    for (const imageNode of imageNodes) {
      addImage(imageNode.getAttribute("data-old-hires") || imageNode.getAttribute("src"), "main");
      try {
        const dynamic = JSON.parse(imageNode.getAttribute("data-a-dynamic-image") || "{}");
        for (const url of Object.keys(dynamic)) addImage(url, "additional");
      } catch { /* No public dynamic-image JSON. */ }
    }
    for (const image of doc.querySelectorAll("#altImages img")) {
      addImage(image.getAttribute("src"), "additional");
    }
    if (images.length) result.images = images;
    else missing.push("images");

    const bullets = [...doc.querySelectorAll("#feature-bullets li .a-list-item")]
      .map(textOf).filter(v => v && !/^(?:See more|About this item)$/i.test(v)).slice(0, 50);
    if (bullets.length) result.bullet_points = bullets;
    else missing.push("bullet_points");
    const description = textOf(doc.querySelector("#productDescription"));
    if (description) result.description = description;
    else missing.push("description");
    const category = [...doc.querySelectorAll("#wayfinding-breadcrumbs_feature_div a")]
      .map(textOf).filter(Boolean).join(" > ");
    if (category) result.category = category;

    // Variant exploration is a separate follow-up PR. Do not synthesize combinations.
    if (doc.querySelector('#twister, [id^="variation_"]')) {
      missing.push("variants_not_exhaustive");
      warnings.push("检测到规格选项；当前版本尚未采集完整变体关系");
    }
    if (missing.length) warnings.push("部分商品信息在当前页面不可确认");
    result.collection_status = missing.length ? "partial" : "complete";
    return result;
  }

  globalThis.UnaAmazonCore = Object.freeze({ canonicalAmazonURL, parsePrice, parseProduct });
})();
