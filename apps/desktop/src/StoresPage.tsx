import { useCallback, useEffect, useState } from "react";
import { disconnectStore, listStores, startStoreAuthorization, type StoreConnection } from "./stores";

const statusText: Record<StoreConnection["status"], string> = {
  connected: "已连接",
  token_expired: "Token 已过期（等待续签功能）",
  reauth_required: "需要重新授权",
  refresh_unknown: "续签状态待确认",
  disconnected: "已断开",
};

export default function StoresPage() {
  const [items, setItems] = useState<StoreConnection[]>([]);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [authorizationURL, setAuthorizationURL] = useState("");
  const [notice, setNotice] = useState("");

  const reload = useCallback(async (signal?: AbortSignal) => {
    setLoading(true);
    setError("");
    try {
      const result = await listStores(signal);
      if (!signal?.aborted) setItems(result);
    } catch (e) {
      if (!signal?.aborted) setError(e instanceof Error ? e.message : "无法读取店铺");
    } finally {
      if (!signal?.aborted) setLoading(false);
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void reload(controller.signal);
    return () => controller.abort();
  }, [reload]);

  async function authorize() {
    setBusy(true);
    setError("");
    setAuthorizationURL("");
    setNotice("");
    try {
      const url = await startStoreAuthorization();
      setAuthorizationURL(url);
      setNotice("请在系统浏览器打开下方官方链接，完成卖家授权后返回此处刷新店铺。");
    } catch (e) {
      setError(e instanceof Error ? e.message : "无法开始授权");
    } finally {
      setBusy(false);
    }
  }

  async function copyURL() {
    if (!authorizationURL) return;
    try {
      await navigator.clipboard.writeText(authorizationURL);
      setNotice("授权链接已复制，请粘贴到系统浏览器打开。");
    } catch {
      setNotice("复制失败，请手动选中下方链接复制到系统浏览器。");
    }
  }

  async function disconnect(item: StoreConnection) {
    if (!window.confirm("确定断开店铺“" + (item.nickname || item.seller_id) + "”？之后需重新授权才能访问该店铺数据。")) return;
    setBusy(true);
    setError("");
    try {
      await disconnectStore(item.id);
      setAuthorizationURL("");
      setNotice("已断开本系统中的店铺连接。平台侧授权可在美客多开发者中心另行撤销。");
      await reload();
    } catch (e) {
      setError(e instanceof Error ? e.message : "断开店铺失败");
    } finally {
      setBusy(false);
    }
  }

  return (
    <section className="stores-page" aria-label="美客多店铺管理">
      <div className="page-heading">
        <div><h1>店铺管理</h1><p>连接 Mercado Libre Global Selling 卖家账号。采集箱不需要店铺授权。</p></div>
        <div className="store-actions">
          <button type="button" className="refresh-button" onClick={() => void reload()} disabled={loading || busy}>刷新店铺</button>
          <button type="button" className="primary-button" onClick={() => void authorize()} disabled={busy}>添加美客多店铺</button>
        </div>
      </div>

      {error && <p className="store-feedback store-error" role="alert">{error}</p>}
      {notice && <p className="store-feedback" role="status">{notice}</p>}
      {authorizationURL && (
        <div className="store-link-panel">
          <strong>Mercado Libre 官方授权链接</strong>
          <p>在浏览器中确认当前登录的是希望绑定的卖家账号。链接包含一次性授权 state，有效期约 10 分钟，请勿公开分享。</p>
          <div className="store-actions">
            <a href={authorizationURL} target="_blank" rel="noopener noreferrer">打开官方授权页面</a>
            <button type="button" className="refresh-button" onClick={() => void copyURL()}>复制链接</button>
          </div>
          <input readOnly value={authorizationURL} onFocus={(event) => event.target.select()} aria-label="授权链接（可手动复制）" />
          <button type="button" className="refresh-button" onClick={() => void reload()} disabled={loading}>我已完成授权，刷新列表</button>
        </div>
      )}

      {loading ? <p className="loading-state">正在读取店铺…</p> : items.length === 0 ? (
        <div className="store-empty">暂无已绑定店铺。点击“添加美客多店铺”启动官方 OAuth 授权。</div>
      ) : (
        <div className="store-grid">
          {items.map((item) => (
            <article key={item.id} className="store-card">
              <div className="store-card-top">
                <div><strong>{item.nickname || "未提供昵称"}</strong><p>卖家 ID：{item.seller_id}</p></div>
                <span className={"store-status " + (item.status === "connected" ? "store-connected" : "store-attention")}>{statusText[item.status] || "状态未知"}</span>
              </div>
              <p className="store-note">此处仅确认卖家授权身份；站点范围、订单与刊登能力仍需单独验证。</p>
              <div className="store-actions">
                <button type="button" className="refresh-button" disabled={busy} onClick={() => void authorize()}>重新授权</button>
                {item.status !== "disconnected" && <button type="button" className="refresh-button" disabled={busy} onClick={() => void disconnect(item)}>断开连接</button>}
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
