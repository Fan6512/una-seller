"use strict";
// Everything that knows credentials lives in the isolated MV3 service worker.
// Tokens are never returned to a content script or injected into a web page.
const AUTH = "https://rwktwfxzfkzglxwaulcm.supabase.co";
const API = "https://mkd.186512.xyz";
const PUBLISHABLE_KEY = "sb_publishable_KUDh1isf3-EknWx-rviV0A_w3lm-u0z";
const AMAZON = new Set(["amazon.com", "www.amazon.com"]);
const sessionStorageReady = chrome.storage.local.setAccessLevel({ accessLevel: "TRUSTED_CONTEXTS" });
let refreshInFlight = null;

function amazonProductURL(raw) {
  try {
    const u = new URL(raw);
    if (u.protocol !== "https:" || !AMAZON.has(u.hostname) || u.port || u.username || u.password) return null;
    const m = u.pathname.match(/\/(?:dp|gp\/product)\/([a-z0-9]{10})(?:\/|$)/i);
    return m ? "https://www.amazon.com/dp/" + m[1].toUpperCase() : null;
  } catch { return null; }
}
function fromAmazonContent(sender) {
  try { return Boolean(sender.tab && sender.url && AMAZON.has(new URL(sender.url).hostname) && new URL(sender.url).protocol === "https:"); }
  catch { return false; }
}
function fromPopup(sender) {
  return typeof sender.url === "string" && sender.url === chrome.runtime.getURL("popup.html");
}
async function loadSession() {
  await sessionStorageReady;
  const data = await chrome.storage.local.get("unaSession");
  return data.unaSession || null;
}
async function saveSession(s) {
  await sessionStorageReady;
  await chrome.storage.local.set({ unaSession: s });
}
async function tokenRequest(type, data) {
  const response = await fetch(AUTH + "/auth/v1/token?grant_type=" + type, {
    method: "POST", headers: { apikey: PUBLISHABLE_KEY, "Content-Type": "application/json" },
    body: JSON.stringify(data), cache: "no-store"
  });
  if (!response.ok) {
    if (response.status === 429) throw new Error("登录请求过于频繁，请稍后重试");
    if (type === "password") throw new Error("邮箱、密码错误或账号尚未验证");
    throw new Error("会话已过期，请在插件弹窗重新登录");
  }
  const value = await response.json();
  if (!value.access_token || !value.refresh_token || !value.user?.id || !Number.isFinite(value.expires_in)) {
    throw new Error("身份服务返回无效会话");
  }
  return {
    access_token: value.access_token, refresh_token: value.refresh_token,
    expires_at: Math.floor(Date.now() / 1000) + value.expires_in,
    email: value.user.email || ""
  };
}
async function getAccessToken(forceRefresh = false) {
  const current = await loadSession();
  if (!current) throw new Error("请先在 Una Seller 插件弹窗登录");
  if (!forceRefresh && current.expires_at > Math.floor(Date.now() / 1000) + 60) return current.access_token;
  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      const fresh = await tokenRequest("refresh_token", { refresh_token: current.refresh_token });
      // Do not restore a session that was signed out while a refresh was in progress.
      if ((await loadSession())?.refresh_token !== current.refresh_token) throw new Error("已退出登录，请重新登录");
      await saveSession(fresh);
      return fresh.access_token;
    })().finally(() => { refreshInFlight = null; });
  }
  return refreshInFlight;
}
async function readAmazonDetail(rawURL, sender) {
  if (!fromAmazonContent(sender)) throw new Error("只允许 Amazon.com 页面发起采集");
  const productURL = amazonProductURL(rawURL);
  if (!productURL) throw new Error("不是受支持的 Amazon.com 商品详情链接");
  await getAccessToken(); // No remote detail fetch before extension login.
  const ctrl = new AbortController();
  const timeout = setTimeout(() => ctrl.abort(), 12000);
  try {
    const res = await fetch(productURL, {
      method: "GET", credentials: "include", redirect: "follow", signal: ctrl.signal,
      headers: { Accept: "text/html" }
    });
    if (!res.ok) throw new Error("Amazon 详情请求失败（HTTP " + res.status + "），请手动查看页面后重试");
    if (!amazonProductURL(res.url) || !String(res.headers.get("content-type") || "").includes("text/html")) {
      throw new Error("商品详情被重定向或不是网页，请手动检查");
    }
    if (Number(res.headers.get("content-length") || "0") > 6 * 1024 * 1024) throw new Error("商品详情过大，已停止采集");
    const html = await res.text();
    if (html.length > 6 * 1024 * 1024) throw new Error("商品详情过大，已停止采集");
    return { html };
  } finally { clearTimeout(timeout); }
}
async function saveCollection(collection, sender) {
  if (!fromAmazonContent(sender)) throw new Error("只允许 Amazon.com 页面发起采集");
  if (!collection || collection.source_platform !== "amazon" || collection.source_site !== "US" ||
      !amazonProductURL(collection.source_url)) throw new Error("采集来源校验失败");
  const body = JSON.stringify(collection);
  if (body.length > 800000) throw new Error("采集数据过大");
  async function submit(token) {
    return fetch(API + "/api/v1/collections", {
      method: "POST", cache: "no-store",
      headers: { "Content-Type": "application/json", Authorization: "Bearer " + token },
      body
    });
  }
  let response = await submit(await getAccessToken());
  if (response.status === 401) response = await submit(await getAccessToken(true));
  if (response.status === 401) throw new Error("账号登录已失效，请在插件弹窗重新登录");
  if (response.status === 202) return { ok: true, status: "unverified", message: "服务端已提交，回读待确认；请在采集箱核对" };
  if (!response.ok) throw new Error("商品未确认保存（HTTP " + response.status + "），请检查后重试");
  const result = await response.json();
  if ((result.result !== "created" && result.result !== "updated") || !result.collection?.id) {
    throw new Error("服务端返回缺少保存确认，暂不标记成功");
  }
  return { ok: true, status: result.collection.collection_status || collection.collection_status, id: result.collection.id };
}

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  (async () => {
    if (!msg || typeof msg !== "object") throw new Error("无效请求");
    switch (msg.type) {
      case "SESSION_STATUS": {
        if (!fromPopup(sender)) throw new Error("操作不允许");
        const current = await loadSession();
        return { ok: true, email: current?.email || "" };
      }
      case "SIGN_IN": {
        if (!fromPopup(sender)) throw new Error("操作不允许");
        if (typeof msg.email !== "string" || typeof msg.password !== "string" ||
            msg.email.length > 254 || msg.password.length > 1000) throw new Error("登录资料无效");
        const session = await tokenRequest("password", { email: msg.email.trim(), password: msg.password });
        await saveSession(session);
        return { ok: true, email: session.email };
      }
      case "SIGN_OUT": {
        if (!fromPopup(sender)) throw new Error("操作不允许");
        await sessionStorageReady;
        await chrome.storage.local.remove("unaSession");
        return { ok: true };
      }
      case "FETCH_AMAZON_HTML":
        return { ok: true, ...(await readAmazonDetail(msg.url, sender)) };
      case "SAVE_COLLECTION":
        return await saveCollection(msg.collection, sender);
      default: throw new Error("未知操作");
    }
  })().then(sendResponse).catch(error => sendResponse({ ok: false, error: error instanceof Error ? error.message : "插件操作失败" }));
  return true; // Keep MV3 async message channel open.
});
