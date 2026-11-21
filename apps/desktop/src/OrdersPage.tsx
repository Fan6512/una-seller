import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { cachedStores, listStores, watchStoreList, type StoreConnection } from "./stores";
import {
  getOrder, getOrderSyncState, listOrders, orderDate, orderMoney, orderStatusLabel,
  OrderRequestError, resetOrderCursor, skipOrder, syncOrders,
  type OrderPage, type OrderProgress, type OrderRecord, type OrderReview, type OrderScope, type OrderSyncState,
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
function trustedOrderImage(src?: string) {
  if (!src) return false;
  try {
    const uri = new URL(src);
    return uri.protocol === "https:" &&
      (uri.hostname === "mlstatic.com" || uri.hostname.endsWith(".mlstatic.com"));
  } catch { return false; }
}
function ProductThumbnail({ src, title, size = "list" }: {
  src?: string; title: string; size?: "list" | "detail";
}) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [src]);
  return <span className={"orders-image orders-image-" + size}>
    {trustedOrderImage(src) && !failed
      ? <img src={src} loading="lazy" decoding="async" referrerPolicy="no-referrer"
          alt={title ? "商品主图：" + title : "商品主图"} onError={() => setFailed(true)} />
      : <span className="orders-image-placeholder" aria-label="暂无商品主图">暂无图片</span>}
  </span>;
}

export default function OrdersPage({ beginSync, installingUpdate }: {
  beginSync: BeginSync;
  installingUpdate: boolean;
}) {
  const [stores, setStores] = useState<StoreConnection[]>(() => cachedStores() ?? []);
  const [storeError, setStoreError] = useState("");
  const [storeRefreshing, setStoreRefreshing] = useState(false);
  const [storeChoice, setStoreChoice] = useState("");
  const [scopeChoice, setScopeChoice] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const reloadStores = useCallback(async () => {
    setStoreRefreshing(true);
    setStoreError("");
    try {
      // Read the latest store/token state; a cached status may have changed
      // after a connection check, OAuth callback, or another page's action.
      const items = await listStores(true);
      setStores(items);
    } catch (error) {
      setStoreError(displayError(error));
    } finally {
      setStoreRefreshing(false);
    }
  }, []);

  useEffect(() => {
    // Keep the order page in sync when Store Management renews or changes
    // a connection. This does not start or repeat an order synchronization.
    const unsubscribe = watchStoreList(setStores);
    void reloadStores();
    return unsubscribe;
  }, [reloadStores]);

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

  return <ScopedOrders key={JSON.stringify([scope, selectedMarket?.user_id])}
    scope={scope} scopeKey={scopeKey} store={store} stores={availableStores} storeError={storeError}
    statusFilter={statusFilter} setStatusFilter={setStatusFilter}
    selectStore={(id) => { setStoreChoice(id); setScopeChoice(""); }} selectScope={setScopeChoice}
    beginSync={beginSync} installingUpdate={installingUpdate}
    reloadStores={reloadStores} storeRefreshing={storeRefreshing} />;
}

function ScopedOrders({ scope, scopeKey, store, stores, storeError, statusFilter, setStatusFilter,
  selectStore, selectScope, beginSync, installingUpdate, reloadStores, storeRefreshing }: {
  scope: OrderScope | null;
  scopeKey: string;
  store: StoreConnection | undefined;
  stores: StoreConnection[];
  storeError: string;
  statusFilter: string;
  setStatusFilter: (value: string) => void;
  selectStore: (id: string) => void;
  selectScope: (key: string) => void;
  beginSync: BeginSync;
  installingUpdate: boolean;
  reloadStores: () => Promise<void>;
  storeRefreshing: boolean;
}) {
  const marketplaces = store?.marketplaces ?? [];
  const [searchText, setSearchText] = useState("");
  const [purchaseLink, setPurchaseLink] = useState<"" | "linked" | "unlinked">("");
  const [offset, setOffset] = useState(0);
  const [nonce, setNonce] = useState(0);
  const [page, setPage] = useState<OrderPage>({ items: [], has_more: false });
  const [loading, setLoading] = useState(false);
  const [listError, setListError] = useState("");
  const [selectedId, setSelectedId] = useState("");
  const detailDialogRef = useRef<HTMLDialogElement>(null);
  const [detail, setDetail] = useState<OrderRecord | null>(null);
  const [detailError, setDetailError] = useState("");
  const [detailLoading, setDetailLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [actionBusy, setActionBusy] = useState(false);
  const [notice, setNotice] = useState("");
  const [review, setReview] = useState<OrderReview | null>(null);
  const [progress, setProgress] = useState<OrderProgress | null>(null);
  const [syncState, setSyncState] = useState<OrderSyncState | null>(null);
  const [completedBatches, setCompletedBatches] = useState(0);
  const [purchase, setPurchase] = useState<OrderPurchase | null>(null);
  const [purchaseNumber, setPurchaseNumber] = useState("");
  const [purchaseCost, setPurchaseCost] = useState("");
  const [purchaseBusy, setPurchaseBusy] = useState(false);
  const [purchaseError, setPurchaseError] = useState("");
  const [purchaseMessage, setPurchaseMessage] = useState("");
  const syncAbort = useRef<AbortController | null>(null);

  const reviewAbort = useRef<AbortController | null>(null);
  const purchaseRequest = useRef<AbortController | null>(null);
  const [purchaseReady, setPurchaseReady] = useState(false);
  const [purchaseReload, setPurchaseReload] = useState(0);

  useLayoutEffect(() => () => {
    syncAbort.current?.abort();
    reviewAbort.current?.abort();
  }, []);

  const loadRows = useCallback(() => setNonce((x) => x + 1), []);
  const closeOrder = () => {
    if (purchaseBusy) return;
    if (purchaseReady && (purchaseNumber !== (purchase?.purchase_order_id ?? "") ||
        purchaseCost !== (purchase?.cost_cny ?? "")) &&
        !window.confirm("采购信息尚未保存，确定关闭订单详情吗？")) return;
    setSelectedId("");
  };
  useLayoutEffect(() => {
    const dialog = detailDialogRef.current;
    if (!dialog) return;
    if (selectedId && !dialog.open) dialog.showModal();
    if (!selectedId && dialog.open) dialog.close();
  }, [selectedId]);
  // Server cursor and Review survive page refresh and process restarts.
  useEffect(() => {
    const controller = new AbortController();
    if (!scope) return () => controller.abort();
    void getOrderSyncState(scope, controller.signal).then((state) => {
      if (controller.signal.aborted) return;
      setSyncState(state);
      if (state.review) setReview(state.review);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setNotice("无法读取订单同步位置：" + displayError(error));
    });
    return () => controller.abort();
  }, [scope, nonce]);

  useEffect(() => {
    const controller = new AbortController();
    if (!scope) {
      setPage({ items: [], has_more: false });
      setLoading(false);
      return () => controller.abort();
    }
    setLoading(true);
    setListError("");
    void listOrders(scope, statusFilter, purchaseLink, offset, controller.signal).then((value) => {
      if (!controller.signal.aborted) setPage(value);
    }).catch((error: unknown) => {
      if (controller.signal.aborted) return;
      setListError(displayError(error));
      setPage({ items: [], has_more: false });
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, [scope?.storeId, scope?.site, scope?.logistic, statusFilter, purchaseLink, offset, nonce]);

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

  useLayoutEffect(() => {
    const controller = new AbortController();
    purchaseRequest.current = controller;
    setPurchaseReady(false);
    setPurchaseBusy(false);
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
      setPurchaseReady(true);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setPurchaseError(displayError(error));
    });
    return () => controller.abort();
  }, [scope, selectedId, purchaseReload]);

  const savePurchaseInfo = async () => {
    const signal = purchaseRequest.current?.signal;
    if (!scope || !selectedId || !purchaseReady || purchaseBusy || !purchaseNumber || !signal) return;
    setPurchaseBusy(true);
    setPurchaseError("");
    setPurchaseMessage("");
    try {
      const result = await savePurchase(scope, selectedId, purchaseNumber.trim(), purchaseCost.trim(), purchase?.updated_at ?? null, signal);
      if (signal.aborted) return;
      setPurchase(result);
      setPurchaseNumber(result.purchase_order_id);
      setPurchaseCost(result.cost_cny ?? "");
      setPurchaseMessage("采购关联已保存，列表状态已更新；金额仍为人工录入。");
      loadRows();
    } catch (error) {
      if (signal.aborted) return;
      if (error instanceof OrderRequestError && error.status === 409) setPurchaseReady(false);
      setPurchaseError(displayError(error));
    } finally {
      if (!signal.aborted) setPurchaseBusy(false);
    }
  };

  const visibleOrders = useMemo(() => {
    const needle = searchText.trim().toLowerCase();
    if (!needle) return page.items;
    return page.items.filter((order) =>
      order.order_id.includes(needle) || order.pack_id?.includes(needle) ||
      order.purchase_order_id?.includes(needle) ||
      order.shipment_id?.includes(needle) ||
      order.order_items.some((item) => item.title.toLowerCase().includes(needle) ||
        item.item_id.toLowerCase().includes(needle)));
  }, [page.items, searchText]);

  const selectStatus = (value: string) => {
    setStatusFilter(value);
    setOffset(0);
    setSearchText("");
  };

  const performSync = async () => {
    if (!scope || running || actionBusy || installingUpdate ||
        (store?.status !== "connected" && store?.status !== "token_expired")) return;
    const endSync = beginSync();
    if (!endSync) return;
    const controller = new AbortController();
    syncAbort.current = controller;
    setRunning(true);
    setCompletedBatches(0);
    setProgress(null);
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
        controller.signal.throwIfAborted();
        saved += current.saved;
        last = current;
        setCompletedBatches(batches + 1);
        setProgress({ ...current, saved });
        if (!current.has_more) break;
      }
      setNotice(last?.has_more
        ? "本轮已保存 " + saved + " 笔订单，达到 20 批安全上限。点击继续同步可从已保存位置接着处理。"
        : "订单同步完成：本轮保存 " + saved + " 笔，已同步到当前搜索窗口末尾。");
      if (store?.status === "token_expired") void reloadStores();
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
        const detail = error instanceof OrderRequestError
          ? "（HTTP " + error.status + " · " + error.code + "）" : "";
        setNotice("本轮已保存约 " + saved + " 笔。同步中断：" + displayError(error) + detail +
          "。如请求结果未知，请先刷新确认，不要连续重试。");
      }
    } finally {
      if (syncAbort.current === controller) syncAbort.current = null;
      setRunning(false);
      endSync();
      loadRows();
    }
  };

  const performSkip = async () => {
    if (!scope || !review || actionBusy || running) return;
    if (!window.confirm("仅跳过本次无法解析的订单 " + review.order_id +
      "，不操作美客多平台。该决定会被记录。确定继续吗？")) return;
    const controller = new AbortController();
    reviewAbort.current = controller;
    setActionBusy(true);
    try {
      const result = await skipOrder(scope, review, controller.signal);
      if (controller.signal.aborted) return;
      setReview(null);
      setProgress(result);
      setNotice("已记录人工跳过订单 " + review.order_id + "；其他订单仍可正常同步。");
      loadRows();
    } catch (error) {
      if (!controller.signal.aborted) setNotice(displayError(error));
    } finally {
      if (reviewAbort.current === controller) reviewAbort.current = null;
      if (!controller.signal.aborted) setActionBusy(false);
    }
  };
  const performReset = async () => {
    if (!scope || !syncState?.can_reset || syncState.next_offset <= 0 || running || actionBusy || review) return;
    if (!window.confirm("将当前订单搜索窗口从第 0 页重新扫描，不删除已保存的订单。确定重扫吗？")) return;
    const controller = new AbortController();
    reviewAbort.current = controller;
    setActionBusy(true);
    try {
      const result = await resetOrderCursor(scope, syncState.next_offset, controller.signal);
      if (controller.signal.aborted) return;
      setProgress(result);
      setSyncState({ next_offset: 0, can_reset: false });
      setNotice("重扫游标已归零，不会删除已有订单或人工 Skip 审计。");
      loadRows();
    } catch (error) {
      if (!controller.signal.aborted) setNotice(displayError(error));
    } finally {
      if (reviewAbort.current === controller) reviewAbort.current = null;
      if (!controller.signal.aborted) setActionBusy(false);
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

  const canSync = Boolean(scope &&
    (store?.status === "connected" || store?.status === "token_expired") && !installingUpdate);
  const storeStatusText: Record<StoreConnection["status"], string> = {
    connected: "已连接", token_expired: "访问令牌已到期",
    reauth_required: "需要重新授权", refresh_unknown: "续签状态待确认", disconnected: "已断开",
  };
  const disabledReason = !scope ? "请先选择已验证的店铺、站点和物流类型"
    : store?.status === "reauth_required" ? "店铺需要重新授权，请到店铺管理处理"
      : store?.status === "refresh_unknown" ? "令牌续签状态未知，请到店铺管理核实"
        : installingUpdate ? "正在安装应用更新" : "当前店铺暂不可同步，请检查连接状态";

  return (
    <div className="orders-page">
      <div className="page-heading">
        <div><h1>订单管理</h1><p>按店铺与站点查看 Mercado Libre 订单 · 订单详情不包含买家隐私资料</p></div>
        <div className="orders-heading-actions">
          <button className="refresh-button" type="button" onClick={() => void reloadStores()}
            disabled={storeRefreshing || running}>{storeRefreshing ? "读取店铺中…" : "刷新店铺状态"}</button>
          <button className="refresh-button" type="button" onClick={loadRows} disabled={loading || running}>刷新订单</button>
          {running ? (
            <button type="button" className="orders-main-button orders-stop" onClick={() => syncAbort.current?.abort()}>停止同步</button>
          ) : (
            <button type="button" className="orders-main-button" onClick={() => void performSync()}
              disabled={!canSync || actionBusy} title={!canSync ? disabledReason : undefined}>
              {installingUpdate ? "正在安装更新" : "同步订单"}</button>
          )}
        </div>
      </div>

      <p className="orders-filter-note" role="status">
        {store ? "当前店铺：" + (store.nickname || store.seller_id) + " · " +
          (storeStatusText[store.status] ?? "状态未知") : "请先连接并验证店铺"}
        {" · "}{store ? "已核实站点 " + marketplaces.length + " 个" : "无可选站点"}
      </p>
      {store?.status === "token_expired" && <p className="orders-filter-note" role="status">
        美客多访问令牌已到期。点击「同步订单」后服务端将尝试续签；若失败，请到「店铺管理」验证连接。
      </p>}
      {scope && !canSync && !installingUpdate && <p className="orders-filter-note" role="status">{disabledReason}</p>}
      <div className="orders-filters">
        <label>店铺
          <select aria-label="订单店铺" value={store?.id ?? ""} onChange={(event) => selectStore(event.target.value)}>
            {stores.map((item) => <option key={item.id} value={item.id}>{item.nickname || item.seller_id}</option>)}
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
      {!stores.length && <div className="orders-empty">
        <strong>还没有可查看订单的店铺</strong>
        <p>请先在“店铺管理”完成美客多 OAuth 授权并验证站点映射。</p>
      </div>}

      {stores.length > 0 && <>
        <div className="orders-tabs" role="group" aria-label="订单状态筛选">
          {FILTERS.map((filter) => <button type="button" key={filter.value}
            aria-pressed={statusFilter === filter.value} onClick={() => selectStatus(filter.value)}>
            {filter.label}
          </button>)}
          <span className="orders-tabs-meta">本页 {page.items.length} 笔 · 每页 {PAGE_SIZE} 笔</span>
        </div>
        <div className="orders-procurement-toolbar" role="group" aria-label="1688采购关联筛选">
          <strong>1688 采购</strong>
          {([
            { value: "", label: "全部" },
            { value: "unlinked", label: "待关联" },
            { value: "linked", label: "已关联" },
          ] as const).map((item) => <button type="button" key={item.value}
            aria-pressed={purchaseLink === item.value}
            onClick={() => { setPurchaseLink(item.value); setOffset(0); setSearchText(""); }}>
            {item.label}
          </button>)}
          <span>按数据库关联状态筛选 · 不代表厂家已发货</span>
        </div>

        {running && <div className="orders-notice" role="status">
          正在同步：已完成 {completedBatches} 批，本次已保存 {progress?.saved ?? 0} 笔。
          最近一批处理 {progress?.processed ?? 0} 个搜索结果组；单次操作最多 20 批。
        </div>}
        {(notice || syncState?.can_reset) && <div className="orders-notice" role="status">
          <span>{notice || "当前存在可重新扫描的服务器同步窗口。"}</span>
          {syncState?.can_reset && !review && !running &&
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
        <div className="orders-main-layout">
          <section className="orders-list-panel" aria-label="订单列表">
            <div className="orders-table-scroll">
              <table className="orders-table">
                <thead><tr>
                  <th>订单信息</th><th>商品 / 规格</th><th>销售金额</th><th>1688 采购关联</th><th>付款状态</th><th title="平台配送记录标识，不代表已发货，也不等于快递运单号">平台物流</th><th>下单时间</th><th>操作</th>
                </tr></thead>
                <tbody>
                  {loading && page.items.length === 0 && <tr><td colSpan={8} className="orders-table-empty">正在读取订单…</td></tr>}
                  {!loading && visibleOrders.length === 0 && <tr><td colSpan={8} className="orders-table-empty">没有符合条件的订单。可调整筛选或点击同步订单。</td></tr>}
                  {visibleOrders.map((order) => <tr key={order.order_id} className={selectedId === order.order_id ? "selected" : ""}>
                    <td><button type="button" className="orders-id" onClick={() => openOrder(order.order_id)}>{order.order_id}</button>
                      <span className="orders-subline">{SITE_NAMES[scope?.site ?? ""] ?? scope?.site} · {scope?.logistic}</span>
                      {order.pack_id && <span className="orders-subline">Pack：{order.pack_id}</span>}
                    </td>
                    <td className="orders-product-cell">
                      <div className="orders-product-preview">
                        <ProductThumbnail src={order.order_items[0]?.image_url} title={order.order_items[0]?.title || ""} />
                        <div className="orders-product-info">
                          <span className="orders-product-title">{order.order_items[0]?.title || "平台未返回商品标题"}</span>
                          <span className="orders-subline">
                            {order.order_items.length > 1 ? order.order_items.length + " 个商品行 · " : ""}
                            {orderCount(order)} 件 · {order.order_items[0]?.variation_id ? "变体 " + order.order_items[0].variation_id : order.order_items[0]?.item_id || "无 SKU 信息"}
                          </span>
                        </div>
                      </div>
                    </td>
                    <td><strong className="orders-money">{orderMoney(order.total_amount, order.currency_id)}</strong></td>
                    <td className="orders-purchase-cell">
                      {order.purchase_order_id ? <>
                        <strong>已关联</strong><span title={order.purchase_order_id}>{order.purchase_order_id}</span>
                        <small>{order.purchase_cost_cny != null ? "采购成本 ¥" + order.purchase_cost_cny : "采购成本待补录"}</small>
                      </> : <button className="orders-table-action" type="button" onClick={() => openOrder(order.order_id)}>关联采购</button>}
                    </td>
                    <td><span className={"orders-status " + safeOrderStatus(order.status)}>{orderStatusLabel(order.status)}</span></td>
                    <td>{order.shipment_id ? <span className="orders-tracking" title="仅表示已关联平台物流记录，不代表已发货"><small>Shipment ID</small>{order.shipment_id}</span> : <span className="orders-muted">暂无平台物流记录</span>}</td>
                    <td><span className="orders-date">{orderDate(order.date_created)}</span></td>
                    <td><button className="orders-table-action" type="button" onClick={() => openOrder(order.order_id)}>订单详情</button></td>
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

          <dialog ref={detailDialogRef} className="orders-modal" aria-labelledby="orders-detail-title"
            onCancel={(event) => { event.preventDefault(); closeOrder(); }}
            onClick={(event) => { if (event.target === event.currentTarget) closeOrder(); }}
            onClose={() => { if (selectedId) setSelectedId(""); }}>
            <div className="orders-detail" aria-label="订单详情">
              <header className="orders-detail-head"><div><small>订单详情 · {SITE_NAMES[scope?.site ?? ""] ?? scope?.site}</small>
                <h2 id="orders-detail-title">{selectedId || "订单详情"}</h2></div>
                <button type="button" aria-label="关闭订单详情" disabled={purchaseBusy} onClick={closeOrder}>×</button>
              </header>
              <div className="orders-modal-scroll">
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
                    <ProductThumbnail src={line.image_url} title={line.title || line.item_id} size="detail" />
                    <div className="orders-line-content">
                      <strong>{line.title || line.item_id}</strong>
                      <small>商品 ID：{line.item_id}</small>
                      {line.variation_id && <small>变体 ID：{line.variation_id}</small>}
                      <span>{orderMoney(line.unit_price, line.currency_id)} × {line.quantity}</span>
                    </div>
                  </div>)}
              </section>
              <section className="orders-detail-block">
                <h3>订单与物流</h3>
                <dl className="orders-detail-fields">
                  <dt>销售订单号</dt><dd><button type="button" onClick={() => copy(detail.order_id)}>{detail.order_id}　复制</button></dd>
                  <dt>Pack ID</dt><dd>{detail.pack_id || "—"}</dd>
                  <dt>物流模式</dt><dd>{scope?.logistic ?? "—"}</dd>
                  <dt>平台物流记录号</dt><dd>{detail.shipment_id || "未返回"} <small>（Shipment ID）</small></dd>
                </dl>
                <p className="orders-muted">此处的 Shipment ID 并非物流运单号；当前订单 API 不含追踪状态、收件人地址、电话或面单，不能据此认定订单已发货。</p>
              </section>
              <section className="orders-detail-block orders-purchase-block">
                <div className="orders-detail-block-title"><h3>1688 采购关联</h3><span>{purchase ? "已关联" : "待关联"}</span></div>
                <p>先保存实际 1688 采购单号与人民币成本。后续优先通过 1688 买家账号授权读取订单金额及采购物流；浏览器插件仅作授权接口不可用时的备选，不会虚构利润。</p>
                <div className="orders-purchase-form">
                  <label>1688 采购订单号
                    <input type="text" inputMode="numeric" maxLength={40} placeholder="输入 1688 采购单号"
                      value={purchaseNumber} disabled={!purchaseReady || purchaseBusy} onChange={(event) => setPurchaseNumber(event.target.value)}
                      aria-label="1688 采购订单号" />
                  </label>
                  <label>采购总金额 CNY（可选）
                    <input type="text" inputMode="decimal" maxLength={20} placeholder="例如 16.50，可稍后补录"
                      value={purchaseCost} disabled={!purchaseReady || purchaseBusy} onChange={(event) => setPurchaseCost(event.target.value)}
                      aria-label="人民币采购总金额" />
                  </label>
                  {purchaseError && <p className="orders-alert" role="alert">{purchaseError}</p>}
                  {purchaseMessage && <p className="orders-purchase-success" role="status">{purchaseMessage}</p>}
                  <div className="orders-purchase-actions">
                    <button type="button" className="orders-main-button" disabled={!purchaseReady || purchaseBusy || !/^\d{8,40}$/.test(purchaseNumber)}
                      onClick={() => void savePurchaseInfo()}>{purchaseBusy ? "保存中…" : "保存采购关联"}</button>
                    {!purchaseReady && purchaseError && <button type="button"
                      onClick={() => setPurchaseReload((n) => n + 1)}>重新读取采购关联</button>}
                    <a href="https://www.1688.com/" target="_blank" rel="noopener noreferrer">打开 1688</a>
                  </div>
                  {purchase && <p className="orders-purchase-state">已保存采购单：{purchase.purchase_order_id}
                    {purchase.cost_cny !== null ? " · 人民币 ¥" + purchase.cost_cny : " · 金额待获取"}</p>}
                </div>
                <div className="orders-purchase-followup">
                  <strong>采购代发货流程 · 两张订单关联</strong>
                  <span>一张是美客多买家的销售订单，一张是向 1688 厂家采购的订单。中间由货代仓接收，再交接官方或指定承运商扫描发货。</span>
                  <ol className="orders-fulfillment-steps" aria-label="采购至买家签收的履约流程">
                    <li>
                      <b>1688 厂家采购</b>
                      <small>{purchase ? "采购单号已关联；付款与发货状态待同步" : "采购单号尚未关联"}</small>
                    </li>
                    <li><b>国内快递运输</b><small>厂家发货及国内物流状态待接入</small></li>
                    <li><b>货代仓库签收</b><small>货代仓入库状态待接入</small></li>
                    <li><b>官方物流扫描交接</b><small>交接扫描与国际运输状态待接入</small></li>
                    <li><b>海外买家签收</b><small>美客多 Shipment 配送状态待接入</small></li>
                  </ol>
                  <span>货代仓签收、官方物流接收、买家签收是三个不同事件；当前没有实时状态数据时不显示为已完成。</span>
                </div>
                <button type="button" disabled className="orders-shipping-disabled">提交发货（接口待接入）</button>
              </section>
            </>}
              </div>
            </div>
          </dialog>
        </div>
      </>}
    </div>
  );
}
