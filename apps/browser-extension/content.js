"use strict";
(() => {
  const core = globalThis.UnaAmazonCore;
  if (!core) return;

  const CARD = '[data-component-type="s-search-result"], .s-result-item.s-asin[data-asin], [class*="_c2Itd_asinContainer_"][data-asin], [data-asin][data-uuid], [data-asin][data-component-id]';
  const BUTTON = ".una-collector-button";
  let pending = false;

  function humanError(error) {
    const raw = String(error?.message || error || "");
    if (/登录|session|401/i.test(raw)) return "登录状态过期，请点插件图标重新登录";
    if (/captcha|验证码|robot|blocked/i.test(raw)) return "亚马逊要求验证码或登录，先处理后再采集";
    if (/403|origin not allowed/i.test(raw)) return "云端暂未允许插件访问，请检查 Go API 是否更新";
    if (/404|不是受支持|不是标准/i.test(raw)) return "无法确认商品详情地址，请打开商品页再试";
    if (/429/i.test(raw)) return "请求太频繁了，请稍后再试";
    if (/502|503|504|500/i.test(raw)) return "云端暂时不可用，请稍后再试";
    if (/超时|timeout|abort/i.test(raw)) return "获取详情超时，请刷新商品页重试";
    if (/标题|完整加载/i.test(raw)) return "页面商品信息还没加载完，请刷新后再试";
    return raw ? "采集失败：" + raw.slice(0, 120) : "采集失败，请重试";
  }
  function missingText(product) {
    const labels = {
      price: "价格", images: "图片", additional_images: "附图", bullet_points: "卖点",
      description: "详细描述", variants_not_exhaustive: "完整规格"
    };
    const missing = (product?.missing_fields || []).map(x => labels[x]).filter(Boolean);
    return missing.length ? "缺少：" + missing.join("、") : "部分字段未获取";
  }
  function partialMessage(product, action) {
    const missing = product?.missing_fields || [];
    // A partial update with no new images does not overwrite earlier saved images.
    if (action === "updated" && missing.includes("images") && !Object.hasOwn(product, "images")) {
      const other = missing.filter(field => field !== "images");
      return "本次未确认主图，之前保存的图片未被覆盖（如有）" +
        (other.length ? "；" + missingText({ missing_fields: other }) : "");
    }
    return (action === "updated" ? "已更新采集箱记录，" : "已保存到采集箱，") + missingText(product);
  }
  function createControl(url, context = "card") {
    const wrapper = document.createElement("div");
    wrapper.className = "una-collector-wrapper una-collector-" + context;
    wrapper.setAttribute("data-una-product", url);
    const button = document.createElement("button");
    button.type = "button";
    button.className = "una-collector-button";
    button.textContent = "采集";
    const message = document.createElement("span");
    message.className = "una-collector-note";
    message.textContent = "保存到采集箱";
    wrapper.append(button, message);
    function update(text, state = "") {
      message.textContent = text;
      message.dataset.state = state;
    }
    button.addEventListener("click", async event => {
      event.preventDefault();
      event.stopPropagation();
      if (button.disabled) return;
      button.disabled = true;
      button.textContent = "采集中…";
      update("正在读取商品详情…", "progress");
      try {
        let page = document;
        if (core.canonicalAmazonURL(location.href) !== url || !document.querySelector("#productTitle")) {
          const result = await chrome.runtime.sendMessage({ type: "FETCH_AMAZON_HTML", url });
          if (!result?.ok || typeof result.html !== "string") throw new Error(result?.error || "获取商品详情失败");
          page = new DOMParser().parseFromString(result.html, "text/html");
        }
        const product = core.parseProduct(page, url);
        update("正在保存到云端采集箱…", "progress");
        const result = await chrome.runtime.sendMessage({ type: "SAVE_COLLECTION", collection: product });
        if (!result?.ok) throw new Error(result?.error || "服务器没有确认保存");
        if (result.status === "unverified") {
          button.textContent = "待核对";
          update("已提交，请到采集箱确认是否保存", "warning");
        } else if (result.status === "partial" || product.collection_status === "partial") {
          button.textContent = "部分成功";
          update(partialMessage(product, result.action), "warning");
        } else if (result.action === "updated") {
          button.textContent = "已更新";
          update("已更新采集箱记录", "success");
        } else {
          button.textContent = "已采集";
          update("采集成功，已保存到采集箱", "success");
        }
      } catch (error) {
        button.textContent = "重试";
        update(humanError(error), "error");
      } finally {
        button.disabled = false;
      }
    });
    return wrapper;
  }
  function findLink(card) {
    const anchors = card.querySelectorAll('h2 a[href], a.a-link-normal.s-no-outline[href], a[href*="/dp/"], a[href*="/gp/product/"]');
    for (const anchor of anchors) {
      if (anchor.closest(".una-collector-wrapper")) continue;
      const normalized = core.canonicalAmazonURL(anchor.href);
      if (normalized) return normalized;
    }
    // A valid ASIN can be the only reliable identifier on revised search cards.
    const asin = card.getAttribute("data-asin");
    if (card.querySelector("h2") && /^[A-Z0-9]{10}$/i.test(asin || "")) return "https://www.amazon.com/dp/" + asin.toUpperCase();
    return null;
  }
  function scanCards() {
    const cards = [...document.querySelectorAll(CARD)];
    for (const card of cards) {
      if (card.querySelector(BUTTON)) continue;
      // Prefer the outside item card; reject nested ASIN widgets representing the same item.
      const parent = card.parentElement?.closest(CARD);
      if (parent && parent.getAttribute("data-asin") === card.getAttribute("data-asin") &&
          parent.querySelector("h2")) continue;
      if (!card.querySelector("h2")) continue;
      const url = findLink(card);
      if (!url) continue;
      // Crucially, mount INSIDE the card, after its content (not after a heading
      // or beside a grid item where the button may drift into the next row).
      const mount = card.querySelector(".s-card-container > .a-section, .puis-card-container > .a-section") ||
        card.querySelector(".puis-card-container, .s-card-container") || card;
      if (!mount.querySelector(BUTTON)) mount.appendChild(createControl(url));
    }
  }
  function scanDetail() {
    const title = document.querySelector("#productTitle");
    // Only mount inside the known title section, never the entire detail grid.
    const holder = title?.closest("#titleSection");
    const url = core.canonicalAmazonURL(location.href);
    let existing = null;
    for (const control of document.querySelectorAll(".una-collector-detail")) {
      if (!existing && holder && url && control.parentElement === holder &&
          control.getAttribute("data-una-product") === url) {
        existing = control;
      } else {
        control.remove();
      }
    }
    if (!holder || !url || existing) return;
    holder.appendChild(createControl(url, "detail"));
  }
  function scan() {
    pending = false;
    scanCards();
    scanDetail();
  }
  function schedule() {
    if (pending) return;
    pending = true;
    setTimeout(scan, 250);
  }
  scan();
  if (document.body) new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true });
})();