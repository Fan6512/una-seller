"use strict";
const $ = id => document.getElementById(id);
async function message(payload) {
  try { return await chrome.runtime.sendMessage(payload); }
  catch { return { ok: false, error: "插件后台未就绪，请重开弹窗重试" }; }
}
function showSession(email) {
  $("signed-out").hidden = Boolean(email);
  $("signed-in").hidden = !email;
  $("account").textContent = email || "";
}
async function restore() {
  const res = await message({ type: "SESSION_STATUS" });
  showSession(res.ok ? res.email : "");
  if (!res.ok) $("message").textContent = res.error || "无法读取会话状态";
}
$("login").addEventListener("submit", async event => {
  event.preventDefault();
  $("message").textContent = "";
  $("sign-in").disabled = true;
  const password = $("password").value;
  const res = await message({ type: "SIGN_IN", email: $("email").value.trim(), password });
  $("password").value = "";
  $("sign-in").disabled = false;
  if (res.ok) showSession(res.email);
  else $("message").textContent = res.error || "登录失败";
});
$("sign-out").addEventListener("click", async () => {
  const res = await message({ type: "SIGN_OUT" });
  if (res.ok) { showSession(""); $("message").textContent = "已退出登录"; }
  else $("message").textContent = res.error || "退出失败";
});
void restore();
