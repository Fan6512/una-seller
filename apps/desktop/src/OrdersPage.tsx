import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { cachedStores, listStores, type StoreConnection } from "./stores";
import {
  getOrder, listOrders, orderDate, orderMoney, orderStatusLabel,
  OrderRequestError, resetOrderCursor, skipOrder, syncOrders,
  type OrderPage, type OrderProgress, type OrderRecord, type OrderReview, type OrderScope,
} from "./orders";
import { loadPurchase, savePurchase, type OrderPurchase } from "./orderPurchases";
import "./orders.css";

const PAGE_SIZE = 20;
const SITE_NAMES: Record<string, string> = {
  MLB: "巴西", MLM: "墨西哥", MLA: "阿根廷", MLC: "智利", MCO: "哥伦比亚",
};
const FILTERS = [
  { value: "", label: "全部订单" },
  { value: "paid", label: "已付款" },
  { value: "payment_required", label: "待付款" },
  { value: "cancelled", label: "已取消" },
];
const REVIEW_LABELS: Record<string, string> = {
  invalid_detail: "平台订单详情字段异常",
  detail_not_found: "搜索到订单，但详情接口返回 404",
};
type BeginSync = () => (() => void) | undefined;

function displayError(error: unknown) {
  if (error instanceof Error && error.name === "AbortError") return "已停止操作，请刷新以确认服务端最终状态";
  if (error instanceof Error && error.name === "TimeoutError") return "请求超时，请刷新后继续";
  return error instanceof Error ? error.message : "订单请求失败";
}
function orderCount(order: OrderRecord) {
  return order.order_items.reduce((sum, item) => sum + item.quantity, 0);
}
function safeOrderStatus(status: string) {
  return status === "paid" ? "paid" : status === "cancelled" ? "cancelled" : "other";
}

export default function OrdersPage({ beginSync, installingUpdate }: {
  beginSync: BeginSync;
  installingUpdate: boolean;
}) {
  const [stores, setStores] = useState<StoreConnection[]>(() => cachedStores() ?? []);
  const [storeError, setStoreError] = useState("");
  const [storeChoice, setStoreChoice] = useState("");
  const [scopeChoice, setScopeChoice] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [searchText, setSearchText] = useState("");
  const [offset, setOffset] = useState(0);
  const [nonce, setNonce] = useState(0);
  const [page, setPage] = useState<OrderPage>({ items: [], has_more: false });
  const [loading, setLoading] = useState(false);
  const [listError, setListError] = useState("");
  const [selectedId, setSelectedId] = useState("");
  const [detail, setDetail] = useState<OrderRecord | null>(null);
  const [detailError, setDetailError] = useState("");
  const [detailLoading, setDetailLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [actionBusy, setActionBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [review, setReview] = useState<OrderReview | null>(null);
  const [progress, setProgress] = useState<OrderProgress | null>(null);
  const [purchase, setPurchase] = useState<OrderPurchase | null>(null);
  const [purchaseNumber, setPurchaseNumber] = useState("");
  const [purchaseCost, setPurchaseCost] = useState("");
  const [purchaseBusy, setPurchaseBusy] = useState(false);
  const [purchaseError, setPurchaseError] = useState("");
  const [purchaseMessage, setPurchaseMessage] = useState("");
  const syncAbort = useRef<AbortController | null>(null);

  useEffect(() => {
    let alive = true;
    void listStores().then((items) => {
      if (alive) setStores(items);
    }).catch((error: unknown) => {
      if (alive) setStoreError(displayError(error));
    });
    return () => { alive = false; syncAbort.current?.abort(); };
  }, []);

  const availableStores = useMemo(() =>
    stores.filter((store) => store.status !== "disconnected" &&
      store.marketplaces_status === "verified" && store.marketplaces.length > 0), [stores]);
  const store = availableStores.find((item) => item.id === storeChoice) ?? availableStores[0];
  const marketplaces = store?.marketplaces ?? [];
  const scopeKey = marketplaces.some((m) => m.site_id + "|" + m.logistic_type === scopeChoice)
    ? scopeChoice : marketplaces[0] ? marketplaces[0].site_id + "|" + marketplaces[0].logistic_type : "";
  const selectedMarket = marketplaces.find((m) => m.site_id + "|" + m.logistic_type === scopeKey);
  const scope = useMemo<OrderScope | null>(() => store && selectedMarket
    ? { storeId: store.id, site: selectedMarket.site_id, logistic: selectedMarket.logistic_type } : null,
    [store?.id, selectedMarket?.site_id, selectedMarket?.logistic_type]);

  const loadRows = useCallback(() => setNonce((x) => x + 1), []);
  useEffect(() => {
    setSelectedId("");
    setDetail(null);
    setReview(null);
    setProgress(null);
    setNotice("");
  }, [scope?.storeId, scope?.site, scope?.logistic]);

  useEffect(() => {
    const controller = new AbortController();
    if (!scope) {
      setPage({ items: [], has_more: false });
      setLoading(false);
      return () => controller.abort();
    }
    setLoading(true);
    setListError("");
    void listOrders(scope, statusFilter, offset, controller.signal).then((value) => {
      if (!controller.signal.aborted) setPage(value);
    }).catch((error: unknown) => {
      if (controller.signal.aborted) return;
      setListError(displayError(error));
      setPage({ items: [], has_more: false });
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, [scope?.storeId, scope?.site, scope?.logistic, statusFilter, offset, nonce]);

  useEffect(() => {
    const controller = new AbortController();
    setDetail(null);
    setDetailError("");
    if (!scope || !selectedId) {
      setDetailLoading(false);
      return () => controller.abort();
    }
    setDetailLoading(true);
    void getOrder(scope, selectedId, controller.signal).then((result) => {
      if (!controller.signal.aborted) setDetail(result);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setDetailError(displayError(error));
    }).finally(() => {
      if (!controller.signal.aborted) setDetailLoading(false);
    });
    return () => controller.abort();
  }, [scope?.storeId, scope?.site, scope?.logistic, selectedId, nonce]);

  useEffect(() => {
    const controller = new AbortController();
    setPurchase(null);
    setPurchaseNumber("");
    setPurchaseCost("");
    setPurchaseError("");
    setPurchaseMessage("");
    if (!scope || !selectedId) return () => controller.abort();
    void loadPurchase(scope, selectedId, controller.signal).then((value) => {
      if (controller.signal.aborted) return;
      setPurchase(value.purchase);
      setPurchaseNumber(value.purchase?.purchase_order_id ?? "");
      setPurchaseCost(value.purchase?.cost_cny ?? "");
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setPurchaseError(displayError(error));
    });
    return () => controller.abort();
  }, [scope?.storeId, scope?.site, scope?.logistic, selectedId]);

  const savePurchaseInfo = async () => {
    if (!scope || !selectedId || purchaseBusy || !purchaseNumber) return;
    const controller = new AbortController();
    setPurchaseBusy(true);
    setPurchaseError("");
    setPurchaseMessage("");
    try {
      const result = await savePurchase(scope, selectedId, purchaseNumber.trim(), purchaseCost.trim(), controller.signal);
      setPurchase(result);
      setPurchaseNumber(result.purchase_order_id);
      setPurchaseCost(result.cost_cny ?? "");
      setPurchaseMessage("采购信息已保存。后续优先接入 1688 官方买家授权与订单金额查询。");
    } catch (error) {
      setPurchaseError(displayError(error));
    } finally {
      setPurchaseBusy(false);
    }
  };

  const visibleOrders = useMemo(() => {
    const needle = searchText.trim().toLowerCase();
    if (!needle) return page.items;
    return page.items.filter((order) =>
      order.order_id.includes(needle) || order.pack_id?.includes(needle) ||
      order.shipment_id?.includes(needle) ||
      order.order_items.some((item) => item.title.toLowerCase().includes(needle) ||
        item.item_id.toLowerCase().includes(needle)));
  }, [page.items, searchText]);

  const selectStatus = (value: string) => {
    setStatusFilter(value);
    setOffset(0);
    setSearchText("");
  };
  const selectStore = (value: string) => {
    setStoreChoice(value);
    setScopeChoice("");
    setOffset(0);
    setSearchText("");
  };
  const selectScope = (value: string) => {
    setScopeChoice(value);
    setOffset(0);
    setSearchText("");
  };

  const performSync = async () => {
    if (!scope || running || actionBusy || installingUpdate || store?.status !== "connected") return;
    const controller = new AbortController();
    syncAbort.current = controller;
    const endSync = beginSync();
    setRunning(true);
    setNotice("");
    setReview(null);
    setListError("");
    let saved = 0;
    let batches = 0;
    let last: OrderProgress | null = null;
    try {
      // Bound one user action to avoid unbounded request loops/rate-limit storms.
      for (; batches < 20; batches++) {
        const current = await syncOrders(scope, controller.signal);
        saved += current.saved;
        last = current;
        setProgress({ ...current, saved });
        if (!current.has_more) break;
      }
      setNotice(last?.has_more
        ? "本轮已保存 " + saved + " 笔订单，达到 20 批安全上限。点击继续同步可从已保存位置接着处理。"
        : "订单同步完成：本轮保存 " + saved + " 笔，已同步到当前搜索窗口末尾。");
    } catch (error) {
      if (controller.signal.aborted) {
        setNotice("同步已停止。请刷新后查看服务器上的已保存订单与进度。");
      } else if (error instanceof OrderRequestError && error.review) {
        setReview(error.review);
        const committed = error.progress?.saved ?? 0;
        saved += committed;
        setProgress(error.progress ? { ...error.progress, saved } : null);
        setNotice("本轮已保存 " + saved + " 笔订单，有一笔详情异常，等待核实。");
      } else {
        setNotice("本轮已保存约 " + saved + " 笔。同步中断：" + displayError(error) +
          "。如请求结果未知，请先刷新确认，不要连续重试。");
      }
    } finally {
      if (syncAbort.current === controller) syncAbort.current = null;
      setRunning(false);
      endSync?.();
      loadRows();
    }
  };

  const performSkip = async () => {
    if (!scope || !review || actionBusy || running) return;
    if (!window.confirm("仅跳过本次无法解析的订单 " + review.order_id +
      "，不操作美客多平台。该决定会被记录。确定继续吗？")) return;
    const controller = new AbortController();
    setActionBusy(true);
    try {
      const result = await skipOrder(scope, review, controller.signal);
      setReview(null);
      setProgress(result);
      setNotice("已记录人工跳过订单 " + review.order_id + "；其他订单仍可正常同步。");
      loadRows();
    } catch (error) {
      setNotice(displayError(error));
    } finally {
      setActionBusy(false);
    }
  };
  const performReset = async () => {
    if (!scope || !progress || progress.next_offset <= 0 || running || actionBusy || review) return;
    if (!window.confirm("将当前订单搜索窗口从第 0 页重新扫描，不删除已保存的订单。确定重扫吗？")) return;
    const controller = new AbortController();
    setActionBusy(true);
    try {
      const result = await resetOrderCursor(scope, progress.next_offset, controller.signal);
      setProgress(result);
      setNotice("重扫游标已归零，不会删除已有订单或人工 Skip 审计。");
    } catch (error) {
      setNotice(displayError(error));
    } finally {
      setActionBusy(false);
    }
  };
  const openOrder = (id: string) => setSelectedId(id);
  const searchOrder = () => {
    const input = searchText.trim();
    if (/^\d{1,30}$/.test(input)) openOrder(input);
  };
  const copy = (value: string) => {
    void navigator.clipboard.writeText(value).then(() => setNotice("已复制 " + value))
      .catch(() => setNotice("复制失败，请手动选择文字复制。"));
  };

  const canSync = Boolean(scope && store?.status === "connected" && !installingUpdate);

  return (
    <div className="orders-page">
      <div className="page-heading">
        <div><h1>订单管理</h1><p>按店铺与站点查看 Mercado Libre 订单 · 订单详情不包含买家隐私资料</p></div>
        <div className="orders-heading-actions">
          <button className="refresh-button" type="button" onClick={loadRows} disabled={loading || running}>刷新订单</button>
          {running ? (
            <button type="button" className="orders-main-button orders-stop" onClick={() => syncAbort.current?.abort()}>停止同步</button>
          ) : (
            <button type="button" className="orders-main-button" onClick={() => void performSync()}
              disabled={!canSync || actionBusy}>{installingUpdate ? "正在安装更新" : "同步订单"}</button>
          )}
        </div>
      </div>

      <div className="orders-filters">
        <label>店铺
          <select aria-label="订单店铺" value={store?.id ?? ""} onChange={(event) => selectStore(event.target.value)}>
            {availableStores.map((item) => <option key={item.id} value={item.id}>{item.nickname || item.seller_id}</option>)}
          </select>
        </label>
        <label>站点 / 物流模式
          <select aria-label="站点和物流模式" value={scopeKey} onChange={(event) => selectScope(event.target.value)}>
            {marketplaces.map((market) => <option key={market.site_id + market.logistic_type} value={market.site_id + "|" + market.logistic_type}>
              {(SITE_NAMES[market.site_id] ?? market.site_id) + " · " + market.logistic_type}
            </option>)}
          </select>
        </label>
        <label className="orders-filter-search">搜索订单或商品
          <span className="orders-search-inner">
            <input type="search" value={searchText} placeholder="订单号 / Pack ID / 商品名称（本页）"
              onChange={(event) => setSearchText(event.target.value)}
              onKeyDown={(event) => { if (event.key === "Enter") searchOrder(); }} />
            <button type="button" onClick={searchOrder} title="输入完整数字订单号可跨页查找">查订单</button>
          </span>
        </label>
      </div>
      <p className="orders-filter-note">商品关键词只筛选当前页；输入完整数字订单号并点击“查订单”可直接查询该订单，不受当前分页限制。</p>

      {storeError && <p className="orders-alert" role="alert">{storeError}</p>}
      {!availableStores.length && <div className="orders-empty">
        <strong>还没有可查看订单的店铺</strong>
        <p>请先在“店铺管理”完成美客多 OAuth 授权并验证站点映射。</p>
      </div>}

      {availableStores.length > 0 && <>
        <div className="orders-tabs" role="group" aria-label="订单状态筛选">
          {FILTERS.map((filter) => <button type="button" key={filter.value}
            aria-pressed={statusFilter === filter.value} onClick={() => selectStatus(filter.value)}>
            {filter.label}
          </button>)}
          <span className="orders-tabs-meta">本页 {page.items.length} 笔 · 每页 {PAGE_SIZE} 笔</span>
        </div>

        {notice && <div className="orders-notice" role="status">
          <span>{notice}</span>
          {progress && progress.next_offset > 0 && !review && !running &&
            <button type="button" disabled={actionBusy} onClick={() => void performReset()}>重扫当前窗口</button>}
        </div>}
        {review && <div className="orders-review" role="alert">
          <div>
            <strong>订单 {review.order_id} 需要核查</strong>
            <p>{REVIEW_LABELS[review.reason] ?? review.reason}。仅在人工确认后才能跳过，跳过不会取消美客多订单。</p>
            <small>搜索页位置：{review.offset} · 审核编号：{review.review_id}</small>
          </div>
          <div className="orders-review-actions">
            <button type="button" onClick={() => copy(review.order_id)}>复制订单号</button>
            <button type="button" className="orders-main-button" disabled={actionBusy} onClick={() => void performSkip()}>确认跳过此单</button>
          </div>
        </div>}

        {listError && <p className="orders-alert" role="alert">{listError}</p>}
        <div className={"orders-main-layout " + (selectedId ? "detail-open" : "")}>
          <section className="orders-list-panel" aria-label="订单列表">
            <div className="orders-table-scroll">
              <table className="orders-table">
                <thead><tr>
                  <th>订单信息</th><th>商品 / 规格</th><th>销售金额</th><th>付款状态</th><th>美客多 Shipment</th><th>下单时间</th><th>操作</th>
                </tr></thead>
                <tbody>
                  {loading && page.items.length === 0 && <tr><td colSpan={7} className="orders-table-empty">正在读取订单…</td></tr>}
                  {!loading && visibleOrders.length === 0 && <tr><td colSpan={7} className="orders-table-empty">没有符合条件的订单。可调整筛选或点击同步订单。</td></tr>}
                  {visibleOrders.map((order) => <tr key={order.order_id} className={selectedId === order.order_id ? "selected" : ""}>
                    <td><button type="button" className="orders-id" onClick={() => openOrder(order.order_id)}>{order.order_id}</button>
                      <span className="orders-subline">{SITE_NAMES[scope?.site ?? ""] ?? scope?.site} · {scope?.logistic}</span>
                      {order.pack_id && <span className="orders-subline">Pack：{order.pack_id}</span>}
                    </td>
                    <td className="orders-product-cell">
                      <span className="orders-product-title">{order.order_items[0]?.title || "平台未返回商品标题"}</span>
                      <span className="orders-subline">
                        {order.order_items.length > 1 ? order.order_items.length + " 个商品行 · " : ""}
                        {orderCount(order)} 件 · {order.order_items[0]?.variation_id ? "变体 " + order.order_items[0].variation_id : order.order_items[0]?.item_id || "无 SKU 信息"}
                      </span>
                    </td>
                    <td><strong className="orders-money">{orderMoney(order.total_amount, order.currency_id)}</strong></td>
                    <td><span className={"orders-status " + safeOrderStatus(order.status)}>{orderStatusLabel(order.status)}</span></td>
                    <td>{order.shipment_id ? <span className="orders-tracking">{order.shipment_id}</span> : <span className="orders-muted">无 Shipment ID</span>}</td>
                    <td><span className="orders-date">{orderDate(order.date_created)}</span></td>
                    <td><button className="orders-table-action" type="button" onClick={() => openOrder(order.order_id)}>查看详情</button></td>
                  </tr>)}
                </tbody>
              </table>
            </div>
            <div className="orders-pagination">
              <span>当前第 {Math.floor(offset / PAGE_SIZE) + 1} 页{loading ? " · 读取中…" : ""}</span>
              <div><button type="button" disabled={offset === 0 || loading} onClick={() => setOffset((n) => Math.max(0, n - PAGE_SIZE))}>上一页</button>
                <button type="button" disabled={!page.has_more || loading} onClick={() => setOffset((n) => n + PAGE_SIZE)}>下一页</button></div>
            </div>
          </section>

          {selectedId && <aside className="orders-detail" aria-label="订单详情">
            <header className="orders-detail-head"><div><small>订单详情</small><h2>{selectedId}</h2></div>
              <button type="button" aria-label="关闭订单详情" onClick={() => setSelectedId("")}>×</button>
            </header>
            {detailLoading && <p className="orders-detail-loading">正在读取最新订单详情…</p>}
            {detailError && <p className="orders-alert" role="alert">{detailError}</p>}
            {detail && <>
              <div className="orders-detail-summary">
                <span className={"orders-status " + safeOrderStatus(detail.status)}>{orderStatusLabel(detail.status)}</span>
                <strong>{orderMoney(detail.total_amount, detail.currency_id)}</strong>
                <small>下单：{orderDate(detail.date_created)}</small>
                <small>更新：{orderDate(detail.last_updated)}</small>
              </div>
              <section className="orders-detail-block">
                <h3>商品清单 <small>{orderCount(detail)} 件</small></h3>
                {detail.order_items.length === 0 && <p className="orders-muted">平台未返回商品明细。</p>}
                {detail.order_items.map((line, i) =>
                  <div className="orders-line-item" key={line.item_id + ":" + (line.variation_id ?? "") + ":" + i}>
                    <strong>{line.title || line.item_id}</strong>
                    <small>商品 ID：{line.item_id}</small>
                    {line.variation_id && <small>变体 ID：{line.variation_id}</small>}
                    <span>{orderMoney(line.unit_price, line.currency_id)} × {line.quantity}</span>
                  </div>)}
              </section>
              <section className="orders-detail-block">
                <h3>订单与物流</h3>
                <dl className="orders-detail-fields">
                  <dt>销售订单号</dt><dd><button type="button" onClick={() => copy(detail.order_id)}>{detail.order_id}　复制</button></dd>
                  <dt>Pack ID</dt><dd>{detail.pack_id || "—"}</dd>
                  <dt>物流模式</dt><dd>{scope?.logistic ?? "—"}</dd>
                  <dt>平台 Shipment ID</dt><dd>{detail.shipment_id || "未返回"}</dd>
                </dl>
                <p className="orders-muted">此处的 Shipment ID 并非物流运单号；当前订单 API 不含追踪状态、收件人地址、电话或面单，不能据此认定订单已发货。</p>
              </section>
              <section className="orders-detail-block orders-purchase-block">
                <div className="orders-detail-block-title"><h3>1688 采购关联</h3><span>{purchase ? "已关联" : "待关联"}</span></div>
                <p>先保存实际 1688 采购单号与人民币成本。后续优先通过 1688 买家账号授权读取订单金额及采购物流；浏览器插件仅作授权接口不可用时的备选，不会虚构利润。</p>
                <div className="orders-purchase-form">
                  <label>1688 采购订单号
                    <input type="text" inputMode="numeric" maxLength={40} placeholder="输入 1688 采购单号"
                      value={purchaseNumber} onChange={(event) => setPurchaseNumber(event.target.value)}
                      aria-label="1688 采购订单号" />
                  </label>
                  <label>采购总金额 CNY（可选）
                    <input type="text" inputMode="decimal" maxLength={20} placeholder="例如 16.50，可稍后补录"
                      value={purchaseCost} onChange={(event) => setPurchaseCost(event.target.value)}
                      aria-label="人民币采购总金额" />
                  </label>
                  {purchaseError && <p className="orders-alert" role="alert">{purchaseError}</p>}
                  {purchaseMessage && <p className="orders-purchase-success" role="status">{purchaseMessage}</p>}
                  <div className="orders-purchase-actions">
                    <button type="button" className="orders-main-button" disabled={purchaseBusy || !/^\d{8,40}$/.test(purchaseNumber)}
                      onClick={() => void savePurchaseInfo()}>{purchaseBusy ? "保存中…" : "保存采购关联"}</button>
                    <a href="https://www.1688.com/" target="_blank" rel="noopener noreferrer">打开 1688</a>
                  </div>
                  {purchase && <p className="orders-purchase-state">已保存采购单：{purchase.purchase_order_id}
                    {purchase.cost_cny !== null ? " · 人民币 ¥" + purchase.cost_cny : " · 金额待获取"}</p>}
                </div>
                <div className="orders-purchase-followup">
                  <strong>下一步：1688 买家授权 → 采购成本/国内物流 → 货代 → 美客多 Shipment 跟踪</strong>
                  <span>1688 国内发货、货代转运、美客多海外派送须分别跟踪。现阶段仅保存采购信息，不会把采购包裹已签收误记为海外买家已签收。</span>
                </div>
                <button type="button" disabled className="orders-shipping-disabled">提交发货（接口待接入）</button>
              </section>
            </>}
          </aside>}
        </div>
      </>}
    </div>
  );
}
