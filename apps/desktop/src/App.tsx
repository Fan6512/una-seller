import { useCallback, useEffect, useState } from "react";

type ServiceStatus = "checking" | "online" | "offline";
type HealthResponse = { status: string; service: string };

const apiBase = (import.meta.env.VITE_API_BASE_URL ?? "http://127.0.0.1:8787").replace(/\/$/, "");

export default function App() {
  const [status, setStatus] = useState<ServiceStatus>("checking");
  const [detail, setDetail] = useState("正在检查本机开发服务…");

  const check = useCallback(async (signal?: AbortSignal) => {
    setStatus("checking");
    setDetail("正在检查本机开发服务…");
    try {
      const response = await fetch(`${apiBase}/healthz`, { signal, cache: "no-store" });
      if (!response.ok) throw new Error(`HTTP ${response.status}`);
      const result = (await response.json()) as HealthResponse;
      if (result.status !== "ok" || result.service !== "una-seller-api") {
        throw new Error("服务响应格式不正确");
      }
      if (!signal?.aborted) {
        setStatus("online");
        setDetail("已连接到 Una Seller 开发服务");
      }
    } catch (error) {
      if (signal?.aborted) return;
      setStatus("offline");
      setDetail(error instanceof Error ? error.message : "服务不可访问");
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void check(controller.signal);
    return () => controller.abort();
  }, [check]);

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand"><span className="brand-icon">U</span><span>Una Seller</span></div>
        <nav aria-label="主导航">
          <div className="nav-item nav-item-active">工作台</div>
          <div className="nav-item nav-item-disabled">商品中心 <span>即将开发</span></div>
          <div className="nav-item nav-item-disabled">订单管理 <span>未开发</span></div>
        </nav>
        <div className="sidebar-foot">M0-01 · 最小工程</div>
      </aside>
      <main className="content">
        <header className="page-head">
          <div><p className="eyebrow">MERCADO LIBRE · AI ERP</p><h1>工作台</h1><p className="subtitle">先建立可运行的基础，再逐步实现真正需要的功能。</p></div>
          <span className="phase">开发阶段</span>
        </header>
        <section className="panel" aria-labelledby="health-title">
          <div className="panel-head"><div><h2 id="health-title">服务连接</h2><p>当前只检测本机开发 API，不代表美客多账号已经接入。</p></div></div>
          <div className="service-row">
            <span className={`status-dot ${status}`} aria-hidden="true" />
            <div className="service-copy">
              <strong>{status === "online" ? "连接正常" : status === "checking" ? "检查中" : "尚未连接"}</strong>
              <span role="status">{detail}</span>
              <code>{apiBase}/healthz</code>
            </div>
            <button type="button" disabled={status === "checking"} onClick={() => void check()}>重新检测</button>
          </div>
        </section>
        <section className="panel muted" aria-labelledby="next-title">
          <h2 id="next-title">接下来</h2>
          <p>等完成 M0 工程验证，再接入真实店铺 OAuth。当前卖家主账号 KYC 未完成，`mkd.186512.xyz` 也尚未配置 DNS/HTTPS。</p>
          <div className="chip-row"><span>Windows / macOS</span><span>Tauri 2 + React</span><span>Go API · 本机</span></div>
        </section>
      </main>
    </div>
  );
}
