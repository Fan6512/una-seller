import { useEffect, useState } from "react";
import { storeRequest } from "./stores";
import "./operations.css";

type StoreSummary = {
  store_id: string; nickname: string; connected: boolean;
  paid_orders: number; cached_unanswered: number; incoming_messages_7d: number;
  open_claims: number; paused_items: number; pending_jobs: number; dead_jobs: number;
};
type Summary = { stores: StoreSummary[]; checked_at: string; truncated: boolean; cache_only: boolean };

const metrics = [
  ["paid_orders", "已付款订单"], ["cached_unanswered", "缓存未回答问答"],
  ["incoming_messages_7d", "近7日收到消息"], ["open_claims", "缓存进行中投诉"],
  ["paused_items", "缓存暂停商品"], ["pending_jobs", "待处理通知"],
  ["dead_jobs", "待核实失败通知"],
] as const;

export default function OperationsPage() {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [refresh, setRefresh] = useState(0);
  useEffect(() => {
    let stopped = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    let controller: AbortController | undefined;
    const load = async () => {
      if (stopped || document.hidden || controller) return;
      const request = new AbortController();
      controller = request;
      const timeout = setTimeout(() => request.abort(), 10000);
      setLoading(true);
      try {
        const next = await storeRequest<Summary>("/api/v1/stores/operations-summary", "GET", request.signal);
        if (!stopped && !request.signal.aborted) {
          setSummary(next);
          setError("");
        }
      } catch {
        if (!stopped && !document.hidden) setError("暂时无法读取运营摘要，可能尚未完成服务端升级。");
      } finally {
        clearTimeout(timeout);
        if (controller === request) controller = undefined;
        if (!stopped) {
          setLoading(false);
          if (!document.hidden) timer = setTimeout(load, 30000);
        }
      }
    };
    const visibility = () => {
      clearTimeout(timer);
      if (document.hidden) controller?.abort();
      else void load();
    };
    document.addEventListener("visibilitychange", visibility);
    void load();
    return () => {
      stopped = true;
      clearTimeout(timer);
      controller?.abort();
      document.removeEventListener("visibilitychange", visibility);
    };
  }, [refresh]);

  return <section className="operations-page">
    <div className="page-heading"><div><h1>运营中心</h1><p>本地同步摘要 · 不代表平台全部数据或实时未读数</p></div>
      <button type="button" className="refresh-button" disabled={loading} onClick={() => setRefresh(n => n + 1)}>刷新摘要</button>
    </div>
    <p className="operations-note" role="status">{loading ? "正在读取本地摘要…" : summary ? "读取时间（北京时间）：" + new Date(summary.checked_at).toLocaleString("zh-CN", {timeZone:"Asia/Shanghai", hour12:false}) : "暂无摘要"}</p>
    {error && <p role="alert" className="operations-error">{error}{summary ? " 下方保留上次读取结果。" : ""}</p>}
    {summary?.truncated && <p className="operations-note">仅显示前100家店铺。</p>}
    {summary && summary.stores.length === 0 && <p className="operations-note">尚未绑定店铺。</p>}
    {summary?.stores.map(store => <article className="operations-store" key={store.store_id}>
      <div className="operations-store-header"><h2>{store.nickname || "店铺"}</h2><span>{store.connected ? "已验证连接" : "连接待确认"}</span></div>
      {store.connected ? <div className="operations-metrics">{metrics.map(([key,label]) =>
        <div className="operations-metric" key={key}><span>{label}</span><strong>{store[key]}</strong></div>
      )}</div> : <p className="operations-note">当前授权不可用，暂不展示缓存指标。</p>}
    </article>)}
  </section>;
}
