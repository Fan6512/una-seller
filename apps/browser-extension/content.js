"use strict";
(() => {
  const core = globalThis.UnaAmazonCore;
  if (!core) return;
  const BUTTON = "una-collector-button";
  let scanTimer = null;

  function makeButton(url) {
    const wrapper = document.createElement("div");
    wrapper.className = "una-collector-wrapper";
    const button = document.createElement("button");
    button.type = "button";
    button.className = BUTTON;
    button.textContent = "采集";
    wrapper.appendChild(button);
    button.addEventListener("click", async event => {
      event.preventDefault();
      event.stopPropagation();
      if (button.disabled) return;
      button.disabled = true;
      button.textContent = "正在采集…";
      button.dataset.status = "";
      try {
        let productDocument = document;
        if (core.canonicalAmazonURL(location.href) !== url || !document.querySelector("#productTitle")) {
          const detail = await chrome.runtime.sendMessage({ type: "FETCH_AMAZON_HTML", url });
          if (!detail?.ok || typeof detail.html !== "string") throw new Error(detail?.error || "获取详情失败");
          productDocument = new DOMParser().parseFromString(detail.html, "text/html");
        }
        const product = core.parseProduct(productDocument, url);
        const saved = await chrome.runtime.sendMessage({ type: "SAVE_COLLECTION", collection: product });
        if (!saved?.ok) throw new Error(saved?.error || "提交失败");
        button.dataset.status = saved.status === "complete" ? "success" : "partial";
        button.textContent = saved.status === "unverified" ? "待核验" :
          saved.status === "complete" ? "已采集" : "部分成功";
        button.title = saved.message || "再次点击可重新采集";
      } catch (error) {
        button.dataset.status = "error";
        button.textContent = "重新采集";
        button.title = error instanceof Error ? error.message : "采集失败";
      } finally { button.disabled = false; }
    });
    return wrapper;
  }
  function attach(card, link) {
    if (card.querySelector("." + BUTTON)) return;
    const url = core.canonicalAmazonURL(link);
    if (!url) return;
    card.appendChild(makeButton(url));
  }
  function scan() {
    scanTimer = null;
    for (const card of document.querySelectorAll('[data-component-type="s-search-result"]')) {
      const anchor = [...card.querySelectorAll('a[href*="/dp/"], a[href*="/gp/product/"]')]
        .find(a => core.canonicalAmazonURL(a.href));
      if (anchor) attach(card, anchor.href);
    }
    if (document.querySelector("#productTitle")) {
      const holder = document.querySelector("#titleSection") || document.querySelector("#dp-container");
      if (holder) attach(holder, location.href);
    }
  }
  function scheduleScan() {
    if (scanTimer !== null) return;
    scanTimer = setTimeout(scan, 250);
  }
  scan();
  // Watch dynamically inserted search cards; only buttons are injected automatically.
  const observer = new MutationObserver(scheduleScan);
  if (document.body) observer.observe(document.body, { childList: true, subtree: true });
})();
