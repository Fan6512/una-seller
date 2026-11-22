// Render only server-authored API error strings, with bounded length.
// Network failures and malformed/non-JSON responses retain caller fallbacks.
export async function readAPIError(response: Response, fallback: string): Promise<{message:string;code?:string}> {
  let body: unknown;
  try { body=await response.json(); } catch { return {message:fallback}; }
  if (!body || typeof body!=="object" || Array.isArray(body)) return {message:fallback};
  const payload=body as Record<string,unknown>;
  const code=typeof payload.code==="string" && /^[A-Z][A-Z0-9_]{0,63}$/.test(payload.code)
    ?payload.code:undefined;
  if (code==="CBT_STORE_REQUIRED") {
    return {message:"请先在「设置 → 店铺管理」连接 Mercado Libre CBT 店铺，再使用官方类目推荐。",code};
  }
  const serverText=typeof payload.error==="string"?payload.error.trim():"";
  return {message:serverText&&serverText.length<=300?serverText:fallback,code};
}
