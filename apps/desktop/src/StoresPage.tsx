import { useCallback, useEffect, useRef, useState } from "react";
import { recordCacheHit } from "./DevelopmentDiagnostics";
import { cachedStores, checkStoreConnection, disconnectStore, ensureStoreConnection, listStores, probeStorePublicationMode, startStoreAuthorization, readableStoreError, type StoreConnection } from "./stores";

const statusText: Record<StoreConnection["status"], string> = {
  connected: "已连接",
  token_expired: "访问令牌已到期，系统会尝试续签",
  reauth_required: "需要重新授权",
  refresh_unknown: "续签状态待确认",
  disconnected: "已断开",
};

const publicationModeNames: Record<string, string> = {
  fully_managed: "全托管店铺",
  standard_up: "标准 UP 店铺",
  legacy_candidate: "传统刊登店铺候选（待确认）",
  undetermined: "刊登类型待确认",
};

export default function StoresPage({ onConnectionChecked }: { onConnectionChecked?: (storeId: string) => void }) {
  const [items, setItems] = useState<StoreConnection[]>(() => cachedStores() ?? []);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [authorizationURL, setAuthorizationURL] = useState("");
  const [notice, setNotice] = useState("");
  const lifecycle = useRef<AbortController>(null);
  const lastLoad = useRef(0);
  const cacheReported = useRef(false);

  const reload = useCallback(async () => {
    const signal = lifecycle.current?.signal;
    if (!signal || signal.aborted) return;
    lastLoad.current = Date.now();
    setLoading(true);
    setError("");
    try {
      const result = await listStores();
      if (signal.aborted) return;
      setItems(result);
      const expired = result.filter((item) => item.status === "token_expired");
      if (expired.length) {
        void Promise.allSettled(expired.map((item) => ensureStoreConnection(item.id, signal))).then((outcomes) => {
          if (signal.aborted) return;
          const fresh = cachedStores();
          if (fresh) setItems(fresh);
          const failure = outcomes.find((outcome) => outcome.status === "rejected");
          if (failure) setError(readableStoreError(failure.reason, "暂时无法自动续签"));
        });
      }
    } catch (e) {
      if (!signal?.aborted) setError(readableStoreError(e, "无法读取店铺"));
    } finally {
      if (!signal?.aborted) setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!cacheReported.current && cachedStores()) {
      cacheReported.current = true;
      recordCacheHit("/api/v1/stores");
    }
    const controller = new AbortController();
    lifecycle.current = controller;
    void reload();
    const onFocus = () => {
      if (Date.now() - lastLoad.current >= 15_000) void reload();
    };
    window.addEventListener("focus", onFocus);
    return () => {
      controller.abort();
      window.removeEventListener("focus", onFocus);
    };
  }, [reload]);

  async function authorize() {
    setBusy(true);
    setError("");
    setAuthorizationURL("");
    setNotice("");
    try {
      const url = await startStoreAuthorization();
      setAuthorizationURL(url);
      setNotice("请在系统浏览器打开下方官方链接，完成卖家授权后系统会自动核实刊登类型一次；返回此处刷新店铺。");
    } catch (e) {
      setError(readableStoreError(e, "无法开始授权"));
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

  async function check(item: StoreConnection) {
    setBusy(true);
    setError("");
    try {
      const result = await checkStoreConnection(item.id);
      onConnectionChecked?.(item.id);
      setNotice(result.marketplaces_verified
        ? "Token 有效；获准站点映射已核实，具体刊登能力仍需验证。"
        : "Token 有效；站点映射尚未核实，不允许直接刊登。");
    } catch (e) {
      setError(readableStoreError(e, "检查授权失败"));
    } finally {
      const fresh = cachedStores();
      if (fresh) setItems(fresh);
      setBusy(false);
    }
  }

  async function probeMode(item: StoreConnection) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      const result = await probeStorePublicationMode(item.id);
      const fresh = await listStores(true);
      setItems(fresh);
      setNotice("完成 CBT C0 只读取证："
        + (publicationModeNames[result.publication_mode] ?? "尚未确认")
        + "。仅识别账号刊登模型，官方图片上传和真实发布权限仍需单独验收。");
    } catch (e) {
      setError(readableStoreError(e, "CBT C0 只读取证失败；店铺模式未改变"));
    } finally {
      setBusy(false);
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
      setError(readableStoreError(e, "断开店铺失败"));
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

      {loading && !!items.length && <p role="status">正在更新店铺…</p>}
      {loading && !items.length ? <p className="loading-state">正在读取店铺…</p> : items.length === 0 ? (
        <div className="store-empty">暂无已绑定店铺。点击“添加美客多店铺”启动官方 OAuth 授权。</div>
      ) : (
        <div className="store-grid">
          {items.map((item) => (
            <article key={item.id} className="store-card">
              <div className="store-card-top">
                <div>
                  <strong>{item.nickname || "未提供昵称"}</strong>
                  <p>CBT 主账号 ID：{item.seller_id}</p>
                  <span className="store-type-label">
                    刊登类型：{item.publication_mode_checked_at
                      ? publicationModeNames[item.publication_mode ?? "undetermined"] ?? "刊登类型待确认"
                      : "待检测确认"}
                  </span>
                </div>
                <span className={"store-status " + (item.status === "connected" ? "store-connected" : "store-attention")}>{statusText[item.status] || "状态未知"}</span>
              </div>
              <p className="store-note">
                {item.marketplaces_status === "verified"
                  ? item.marketplaces.length === 0
                    ? "该账号目前没有返回可经营站点。"
                    : "已核实站点：" + item.marketplaces.map((site) => site.site_id + "（子账号 " + site.user_id + " · " + site.logistic_type + "）").join("、")
                  : "站点映射未验证：不开放站点刊登或订单操作。"}
              </p>
              <p className="store-note">
                {item.publication_mode_checked_at ? "最近取证：" + new Date(item.publication_mode_checked_at).toLocaleString() : "尚未取得有效刊登类型证据；可点击下方按钮手动核实。"}
              </p>
              <p className="store-note">C0 是官方只读账号模式识别，不代表具备商品创建或图片上传权限；软件不会修改官方店铺经营模式。</p>
              <div className="store-actions">
                {(item.status === "reauth_required" || item.status === "refresh_unknown" || item.status === "disconnected") &&
                  <button type="button" className="refresh-button" disabled={busy} onClick={() => void authorize()}>重新授权</button>}
                {item.status !== "disconnected" && (
                  <button type="button" className="refresh-button" disabled={busy} onClick={() => void check(item)}>验证连接</button>
                )}
                {item.status !== "disconnected" && (
                  <button type="button" className="refresh-button" disabled={busy} onClick={() => void probeMode(item)}>核实刊登模式（只读）</button>
                )}
                {item.status !== "disconnected" && <button type="button" className="refresh-button" disabled={busy} onClick={() => void disconnect(item)}>断开连接</button>}
              </div>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
