import { useEffect, useState } from "react";
import { storeRequest } from "./stores";
import "./operations.css";

type StoreSummary = {
  store_id: string; nickname: string; connected: boolean; sites: string[];
  paid_orders: number; cached_unanswered: number | null; pending_message_conversations: number | null;
  open_claims: number | null; paused_items: number | null; pending_jobs: number; dead_jobs: number;
};
type WorkRow = { store_id: string; site: string; kind: string; resource_id: string; status: string; observed_at: string; due_at?: string };
type Sales = { day: string; amount_usd: string | null; paid_orders: number; included_orders: number; review_orders: number; refunded_orders: number; cancelled_orders: number; other_orders: number; scope_available: boolean; truncated: boolean };
type Summary = { sales?: Sales; stores: StoreSummary[]; checked_at: string; truncated: boolean; cache_only: boolean; activity: WorkRow[]; todos: WorkRow[] };
type Target = "orders" | "published";
const metrics = [
  ["paid_orders", "缓存已付款订单"], ["cached_unanswered", "缓存未回答问答"],
  ["pending_message_conversations", "待回复会话"], ["open_claims", "缓存进行中投诉"],
  ["paused_items", "缓存暂停商品"], ["pending_jobs", "待处理通知"], ["dead_jobs", "待核实失败通知"],
] as const;
const kinds: Record<string, string> = { order: "订单", question: "问答", message: "消息", claim: "投诉", item: "商品", sync_failure: "同步异常" };
const states: Record<string, string> = { paid: "已付款", cancelled: "已取消", awaiting_dispatch: "待发货", label_printed: "已打印面单",
  logistics_exception: "物流异常", shipment_cancelled: "物流取消", delivered: "已签收", shipped: "已发货", in_transit: "运输中", out_for_delivery: "派送中",
  UNANSWERED: "未回答", ANSWERED: "已回答", available: "已收录", pending_reply: "待回复", replied: "已回复 / 已处理", sent: "已发送", opened: "进行中", closed: "已关闭", paused: "暂停", active: "在售", failed: "待核实" };
const date = (value: string) => new Date(value).toLocaleString("zh-CN", { timeZone: "Asia/Shanghai", hour12: false });
const beijingMinute = new Intl.DateTimeFormat("zh-CN", { timeZone: "Asia/Shanghai", hour: "2-digit", minute: "2-digit", hourCycle: "h23" });

const beijingDate = new Intl.DateTimeFormat("sv-SE", { timeZone: "Asia/Shanghai", year: "numeric", month: "2-digit", day: "2-digit" });

function BeijingClock({ onDay }: { onDay: (day: string) => void }) {
  const [time, setTime] = useState(() => beijingMinute.format(new Date()));
  useEffect(() => {
    let timer: ReturnType<typeof setTimeout>;
    const tick = () => {
      clearTimeout(timer);
      const now = new Date();
      setTime(beijingMinute.format(now));
      onDay(beijingDate.format(now));
      if (!document.hidden) timer = setTimeout(tick, 60000 - (now.getTime() % 60000));
    };
    tick();
    document.addEventListener("visibilitychange", tick);
    return () => { clearTimeout(timer); document.removeEventListener("visibilitychange", tick); };
  }, [onDay]);
  return <span className="operations-note operations-clock">今日 · 北京时间 {time}</span>;
}

export default function OperationsPage({ onNavigate }: { onNavigate: (target: Target) => void }) {
  const [summary, setSummary] = useState<Summary | null>(null);
  const [day, setDay] = useState(() => beijingDate.format(new Date()));
  const [scope, setScope] = useState("");
  const [site, setSite] = useState("");
  const [loadedScope, setLoadedScope] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [refresh, setRefresh] = useState(0);
  const scopeKey = scope + "/" + site;
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
        const query = new URLSearchParams({ store_id: scope, site });
        const next = await storeRequest<Summary>("/api/v1/stores/operations-summary?" + query.toString(), "GET", request.signal);
        if (!Array.isArray(next.activity) || !Array.isArray(next.todos)) throw new Error("operations API not upgraded");
        if (!stopped && !request.signal.aborted) {
          setSummary(next); setLoadedScope(scopeKey); setError("");
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
  }, [refresh, scope, site, scopeKey]);

  const stores = summary?.stores ?? [];
  const selected = stores.filter(s => !scope || s.store_id === scope);
  const sites = [...new Set(selected.flatMap(s => s.sites ?? []))].sort();
  const currentRows = loadedScope === scopeKey ? summary : null;
  const sales = currentRows?.sales;
  const isCurrentSales = sales?.day === day;
  const salesNote = !sales ? "待接入订单销售汇总" : !isCurrentSales ? "等待今日缓存销售额更新" : sales.truncated ? "当日缓存超过汇总上限，暂不显示金额" :
    !sales.scope_available ? "当前店铺或站点授权不可用" :
      sales.day + " · 缓存已付款商品额 " + sales.included_orders + "/" + sales.paid_orders + " 笔" +
      (sales.refunded_orders ? " · 退款排除 " + sales.refunded_orders + " 笔" : "") +
      (sales.review_orders ? " · 金额待核实 " + sales.review_orders + " 笔" : "") +
      (sales.cancelled_orders ? " · 已取消 " + sales.cancelled_orders + " 笔" : "") +
      " · 非实收，平台覆盖未确认";
  // These existing counts are store-wide, not site-wide. Unknown or partial
  // coverage must never be silently summed into a seemingly complete total.
  const total = (key: (typeof metrics)[number][0]) => {
    if (!currentRows || selected.length === 0 || (!scope && currentRows.truncated) ||
        selected.some(store => !store.connected || store[key] == null)) return "—";
    return selected.reduce((sum, store) => sum + (store[key] ?? 0), 0).toLocaleString("zh-CN");
  };
  const list = (rows: WorkRow[], todos: boolean) => {
    const sorted = [...rows].sort((a, b) => {
      if (todos) {
        if (a.due_at && b.due_at) { const order = Date.parse(a.due_at) - Date.parse(b.due_at); if (order) return order; }
        if (a.due_at && !b.due_at) return -1;
        if (!a.due_at && b.due_at) return 1;
      }
      return Date.parse(b.observed_at) - Date.parse(a.observed_at);
    });
    return sorted.length ? <ul className="operations-list">{sorted.map((row, index) => <li key={row.store_id + row.site + row.kind + row.resource_id + index}>
      <div><strong>{kinds[row.kind] ?? row.kind} · {states[row.status] ?? row.status}</strong>
        <span>{stores.find(s => s.store_id === row.store_id)?.nickname || "店铺"} · {row.site || "站点未确认"} · {row.kind === "message" ? "会话 " : ""}{row.resource_id}</span></div>
      <div className="operations-row-time">{row.due_at && <strong>缓存期限：{date(row.due_at)}</strong>}<span>{date(row.observed_at)}</span>
        {todos && row.kind === "order" && <button type="button" onClick={() => onNavigate("orders")}>订单管理</button>}
        {todos && row.kind === "item" && <button type="button" onClick={() => onNavigate("published")}>已刊登商品</button>}
      </div>
    </li>)}</ul> : <p className="operations-note">当前范围暂无本地记录，不代表平台无待办。</p>;
  };
  return <section className="operations-page operations-dashboard">
    <div className="page-heading operations-heading">
      <div><h1>运营中心</h1><p>经营概览与待处理事项</p></div>
      <button type="button" className="refresh-button" disabled={loading} onClick={() => setRefresh(n => n + 1)}>{loading ? "刷新中…" : "刷新"}</button>
    </div>
    <div className="operations-filter-row">
      <div className="operations-filters">
        <label>店铺 <select value={scope} onChange={e => { setScope(e.target.value); setSite(""); }}><option value="">全部店铺</option>{stores.map(s => <option key={s.store_id} value={s.store_id}>{s.nickname || "店铺"}</option>)}</select></label>
        <label>待办 / 动态站点 <select value={site} onChange={e => setSite(e.target.value)}><option value="">全部站点</option>{sites.map(s => <option key={s} value={s}>{s}</option>)}</select></label>
      </div>
    </div>
    {error && <p role="alert" className="operations-error">{error}{currentRows ? " 保留上次读取的摘要。" : ""}</p>}
    {summary?.truncated && <p className="operations-note">摘要最多显示100家店铺，全部店铺汇总暂不可用。</p>}
    {summary && stores.length === 0 && <p className="operations-note">尚未绑定店铺。</p>}

    <section className="operations-section" aria-label="经营总览">
      <div className="operations-section-heading"><h2>经营总览</h2><BeijingClock onDay={setDay} /></div>
      <div className="operations-overview-cards">
        <article className="operations-kpi operations-financial"><span>今日销售额（缓存）</span><strong>{isCurrentSales && sales?.amount_usd != null ? "USD " + sales.amount_usd : "—"}</strong><small>{salesNote}</small></article>
        <article className="operations-kpi operations-financial"><span>今日利润</span><strong>—</strong><small>待真实回款与成本核算</small></article>
      </div>
      <div className="operations-section-heading operations-subheading"><h3>运营指标</h3><span className="operations-note">店铺整体缓存 · 不随站点筛选</span></div>
      <div className="operations-overview-cards operations-count-cards">
        {metrics.slice(0, 4).map(([key,label]) => <article className="operations-kpi" key={key}><span>{label}</span><strong>{total(key)}</strong></article>)}
      </div>
      <p className="operations-note operations-coverage">— 表示暂无完整可用数据；缓存指标不代表平台全部数据或实时未读数。</p>
      {selected.length > 0 && <details className="operations-breakdown">
        <summary>店铺指标明细与同步状态</summary>
        {selected.map(store => <article className="operations-store" key={store.store_id}>
          <div className="operations-store-header"><h3>{store.nickname || "店铺"}</h3></div>
          {currentRows && store.connected ? <div className="operations-metrics">{metrics.map(([key,label]) =>
            <div className="operations-metric" key={key}><span>{label}</span><strong>{store[key] ?? "—"}</strong></div>
          )}</div> : <p className="operations-note">{!store.connected ? "当前授权不可用，请前往设置检查店铺连接。" : "当前范围尚未读取摘要。"}</p>}
        </article>)}
      </details>}
    </section>

    <section className="operations-section" aria-label="待办事项">
      <div className="operations-section-heading"><h2>待办事项</h2><span className="operations-note">{currentRows ? "本地记录 " + currentRows.todos.length + " 条" : "—"}</span></div>
      <article className="operations-list-panel">
        <p className="operations-note">按已知期限优先，最多30条。回复已同步的消息自动移出待办，新买家消息才重新列入。</p>
        {currentRows ? list(currentRows.todos, true) : <p className="operations-note">当前筛选暂无摘要。</p>}
      </article>
    </section>

    <section className="operations-section" aria-label="最近动态">
      <div className="operations-section-heading"><h2>最近动态</h2><span className="operations-note">{currentRows ? "本地记录 " + currentRows.activity.length + " 条" : "—"}</span></div>
      <article className="operations-list-panel">
        <p className="operations-note">最近30条本地观察；消息按会话最新消息时间，其余为本地记录更新时间。</p>
        {currentRows ? list(currentRows.activity, false) : <p className="operations-note">当前筛选暂无摘要。</p>}
      </article>
    </section>
  </section>;
}
