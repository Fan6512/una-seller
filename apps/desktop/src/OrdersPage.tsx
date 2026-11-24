import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { cachedStores, listStores, watchStoreList, StoreRequestError, type StoreConnection } from "./stores";
import {
  getOrder, getOrderSyncState, listOrders, searchOrders, orderRecordScope, orderRecordKey, orderDate, orderMoney,
  OrderRequestError, refreshOrderImage, refreshOneOrder, refreshOrderLogistics, resetOrderCursor, skipOrder, syncOrders,
  shouldApplyOrderRefresh, shouldStopLogisticsBatch, displayOrderStatus, hasProcurement, orderVariationText, shipmentStatusLabel,
  type OrderPage, type OrderProgress, type OrderRecord, type OrderReview, type OrderScope, type OrderSyncState,
} from "./orders";
import { loadPurchase, savePurchase, type OrderPurchase } from "./orderPurchases";
import TranslatedOrderTitle, { useOrderTitleTranslations } from "./TranslatedOrderTitle";
import "./orders.css";

const PAGE_SIZE = 20;
const SITE_NAMES: Record<string, string> = {
  MLB: "巴西", MLM: "墨西哥", MLA: "阿根廷", MLC: "智利", MCO: "哥伦比亚",
};
function logisticLabel(value?: string) {
  return value === "remote" ? "自发货" : value === "fulfillment" ? "海外仓" : value || "—";
}
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
// The absolute cutoff is in China time; the countdown uses epoch milliseconds.
// Never infer that a shipment was dispatched from the Shipment ID or deadline.
export function dispatchDeadlineLabel(order: OrderRecord, now = Date.now(), logistic = "") {
  if (order.status === "cancelled" || order.status === "cancellation_requested")
    return { date: "不适用", countdown: "不适用", hint: "已取消订单", state: "neutral" };
  if (logistic.toLowerCase() === "fulfillment" || logistic.toLowerCase() === "fulfilment")
    return { date: "不适用", countdown: "不适用", hint: "Fulfillment 不适用", state: "neutral" };
  if (!order.shipment_id)
    return { date: "无平台物流", countdown: "无平台物流", hint: "", state: "neutral" };
  if (!order.dispatch_deadline) {
    if (order.dispatch_sla_status === "delayed")
      return { date: "暂无截止时间", countdown: "平台标记逾期", hint: "请核实是否已发货", state: "late" };
    const checked = Boolean(order.dispatch_checked_at);
    return {
      date: checked ? "暂无截止时间" : "待获取",
      countdown: checked ? "暂无截止时间" : "待获取",
      hint: checked ? "官方未提供，请后台核实" : "点击更新本页时限",
      state: "neutral",
    };
  }
  const date = new Date(order.dispatch_deadline);
  if (!Number.isFinite(date.getTime()))
    return { date: "时间格式异常", countdown: "时间格式异常", hint: "请重新获取", state: "neutral" };
  const dateLabel = date.toLocaleString("zh-CN", {
    timeZone: "Asia/Shanghai", year: "numeric", month: "2-digit", day: "2-digit",
    hour: "2-digit", minute: "2-digit", hour12: false,
  });
  const remaining = date.getTime() - now;
  const platformLate = order.dispatch_sla_status === "delayed";
  const countdown = platformLate ? "平台标记逾期"
    : remaining <= 0 ? "截止已过"
      : remaining < 3_600_000 ? "不足1小时"
        : remaining < 86_400_000 ? "剩余" + Math.floor(remaining / 3_600_000) + "小时"
          : "约剩余" + Math.floor(remaining / 86_400_000) + "天";
  const hint = platformLate || remaining <= 0 ? "请核实是否已发货" : "";
  // Red only in the final 24 hours (and for overdue/officially delayed).
  const state = platformLate || remaining < 86_400_000 ? "late" : "neutral";
  return { date: dateLabel, countdown, hint, state };
}
function DispatchCountdown({ order, now, logistic }: {
  order: OrderRecord; now: number; logistic: string;
}) {
  const result = dispatchDeadlineLabel(order, now, logistic);
  return <div className={"orders-dispatch orders-dispatch-" + result.state}
    title={order.dispatch_checked_at ? "官方 SLA 最近核验：" + order.dispatch_checked_at : "此订单尚未读取官方发货 SLA"}>
    <strong>{result.countdown}</strong>
    {result.hint && <small>{result.hint}</small>}
  </div>;
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
  const [searchText, setSearchText] = useState("");
  const [activeQuery, setActiveQuery] = useState("");
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
  const allStores = storeChoice === "all";
  const store = allStores ? undefined : availableStores.find((item) => item.id === storeChoice) ?? availableStores[0];
  const marketplaces = useMemo(() => {
    if (!allStores) return store?.marketplaces ?? [];
    const unique = new Map<string, StoreConnection["marketplaces"][number]>();
    for (const item of availableStores) for (const market of item.marketplaces)
      unique.set(market.site_id + "|" + market.logistic_type, market);
    return [...unique.values()];
  }, [allStores, store, availableStores]);
  const scopeKey = scopeChoice === "all" || allStores && scopeChoice === "" ? "all"
    : marketplaces.some((m) => m.site_id + "|" + m.logistic_type === scopeChoice)
      ? scopeChoice : marketplaces[0] ? marketplaces[0].site_id + "|" + marketplaces[0].logistic_type : "";
  const selectedMarket = marketplaces.find((m) => m.site_id + "|" + m.logistic_type === scopeKey);
  const scope = useMemo<OrderScope | null>(() => store && selectedMarket
    ? { storeId: store.id, site: selectedMarket.site_id, logistic: selectedMarket.logistic_type } : null,
    [store?.id, selectedMarket?.site_id, selectedMarket?.logistic_type]);

  return <ScopedOrders key={JSON.stringify([scope, selectedMarket?.user_id, scopeKey, allStores])}
    scope={scope} scopeKey={scopeKey} store={store} stores={availableStores} storeError={storeError}
    statusFilter={statusFilter} setStatusFilter={setStatusFilter}
    searchText={searchText} setSearchText={setSearchText} activeQuery={activeQuery} setActiveQuery={setActiveQuery}
    storeFilter={allStores ? "all" : store?.id ?? ""}
    siteFilter={scopeKey === "all" ? "*" : selectedMarket?.site_id ?? "*"}
    logisticFilter={scopeKey === "all" ? "*" : selectedMarket?.logistic_type ?? "*"}
    marketplaces={marketplaces}
    selectStore={(id) => { setStoreChoice(id); setScopeChoice(id === "all" ? "all" : ""); }}
    selectScope={setScopeChoice}
    beginSync={beginSync} installingUpdate={installingUpdate}
    reloadStores={reloadStores} storeRefreshing={storeRefreshing} />;
}

function ScopedOrders({ scope, scopeKey, store, stores, storeError, statusFilter, setStatusFilter,
  searchText, setSearchText, activeQuery, setActiveQuery, storeFilter, siteFilter, logisticFilter, marketplaces,
  selectStore, selectScope, beginSync, installingUpdate, reloadStores, storeRefreshing }: {
  scope: OrderScope | null;
  scopeKey: string;
  store: StoreConnection | undefined;
  stores: StoreConnection[];
  storeError: string;
  statusFilter: string;
  setStatusFilter: (value: string) => void;
  searchText: string;
  setSearchText: (value: string) => void;
  activeQuery: string;
  setActiveQuery: (value: string) => void;
  storeFilter: string;
  siteFilter: string;
  logisticFilter: string;
  marketplaces: StoreConnection["marketplaces"];
  selectStore: (id: string) => void;
  selectScope: (key: string) => void;
  beginSync: BeginSync;
  installingUpdate: boolean;
  reloadStores: () => Promise<void>;
  storeRefreshing: boolean;
}) {
  const [purchaseLink, setPurchaseLink] = useState<"" | "linked" | "unlinked">("");
  const [offset, setOffset] = useState(0);
  const [nonce, setNonce] = useState(0);
  const [page, setPage] = useState<OrderPage>({ items: [], has_more: false });
  const translatedTitles = useOrderTitleTranslations(page.items, scope);
  const [loading, setLoading] = useState(false);
  const [imageBusy, setImageBusy] = useState(false);
  const [listError, setListError] = useState("");
  const [selectedId, setSelectedId] = useState("");
  const selectedIdRef = useRef(selectedId);
  useLayoutEffect(() => { selectedIdRef.current = selectedId; }, [selectedId]);
  const [purchaseSelectedId, setPurchaseSelectedId] = useState("");
  const detailDialogRef = useRef<HTMLDialogElement>(null);
  const purchaseDialogRef = useRef<HTMLDialogElement>(null);
  const [detail, setDetail] = useState<OrderRecord | null>(null);
  const [detailError, setDetailError] = useState("");
  const [detailLoading, setDetailLoading] = useState(false);
  const [selectedRowScope, setSelectedRowScope] = useState<OrderScope | null>(null);
  const [purchaseRowScope, setPurchaseRowScope] = useState<OrderScope | null>(null);
  const [running, setRunning] = useState(false);
  const [logisticsAction, setLogisticsAction] = useState<"sla" | "shipment" | null>(null);
  const [slaProgress, setSLAProgress] = useState("");
  const [actionBusy, setActionBusy] = useState(false);
  const [notice, setNotice] = useState("");
  useEffect(() => {
    if (!notice.startsWith("本次尝试补全")) return;
    const timer = window.setTimeout(() => setNotice((current) => current === notice ? "" : current), 8000);
    return () => window.clearTimeout(timer);
  }, [notice]);
  const [review, setReview] = useState<OrderReview | null>(null);
  const [progress, setProgress] = useState<OrderProgress | null>(null);
  const [syncState, setSyncState] = useState<OrderSyncState | null>(null);
  const [completedBatches, setCompletedBatches] = useState(0);
  const [purchase, setPurchase] = useState<OrderPurchase | null>(null);
  const [purchaseNumber, setPurchaseNumber] = useState("");
  const [purchaseCost, setPurchaseCost] = useState("");
  const [warehouseAddress, setWarehouseAddress] = useState("");
  const [warehouseOptions, setWarehouseOptions] = useState<string[]>([]);
  const [purchaseTracking, setPurchaseTracking] = useState("");
  const [purchaseBusy, setPurchaseBusy] = useState(false);
  const [refreshingOrder, setRefreshingOrder] = useState(false);
  const [purchaseError, setPurchaseError] = useState("");
  const [purchaseMessage, setPurchaseMessage] = useState("");
  const syncAbort = useRef<AbortController | null>(null);
  const slaAbort = useRef<AbortController | null>(null);
  const imageAbort = useRef<AbortController | null>(null);
  const orderRefreshAbort = useRef<AbortController | null>(null);

  const reviewAbort = useRef<AbortController | null>(null);
  const purchaseRequest = useRef<AbortController | null>(null);
  const [purchaseReady, setPurchaseReady] = useState(false);
  const [purchaseReload, setPurchaseReload] = useState(0);
  const [clockNow, setClockNow] = useState(() => Date.now());

  useEffect(() => {
    // Local clock only; dispatch is an hourly/day-level concern, not minute-level.
    const updateClock = () => setClockNow(Date.now());
    const interval = window.setInterval(updateClock, 15 * 60_000);
    window.addEventListener("focus", updateClock);
    document.addEventListener("visibilitychange", updateClock);
    return () => {
      window.clearInterval(interval);
      window.removeEventListener("focus", updateClock);
      document.removeEventListener("visibilitychange", updateClock);
    };
  }, []);

  useLayoutEffect(() => () => {
    syncAbort.current?.abort();
    imageAbort.current?.abort();
    slaAbort.current?.abort();
    orderRefreshAbort.current?.abort();
    reviewAbort.current?.abort();
  }, []);

  // Closing or switching orders invalidates the previous detail refresh.
  useLayoutEffect(() => {
    orderRefreshAbort.current?.abort();
    orderRefreshAbort.current = null;
    setRefreshingOrder(false);
  }, [selectedId, selectedRowScope?.storeId, selectedRowScope?.site, selectedRowScope?.logistic,
    scope?.storeId, scope?.site, scope?.logistic]);

  const loadRows = useCallback(() => setNonce((x) => x + 1), []);
  const closeOrder = () => { setSelectedId(""); setSelectedRowScope(null); };
  const closePurchase = () => {
    if (purchaseBusy) return;
    if (purchaseReady && (purchaseNumber !== (purchase?.purchase_order_id ?? "") ||
        purchaseCost !== (purchase?.cost_cny ?? "") ||
        warehouseAddress !== (purchase?.warehouse_address ?? "") ||
        purchaseTracking !== (purchase?.purchase_tracking_number ?? "")) &&
        !window.confirm("采购信息尚未保存，确定关闭采购窗口吗？")) return;
    setPurchaseSelectedId("");
    setPurchaseRowScope(null);
  };
  const openPurchase = (id: string, targetScope: OrderScope | null = scope) => {
    setSelectedId("");
    setSelectedRowScope(null);
    setPurchaseRowScope(targetScope);
    setPurchaseSelectedId(id);
  };
  useLayoutEffect(() => {
    const dialog = detailDialogRef.current;
    if (!dialog) return;
    if (selectedId && !dialog.open) dialog.showModal();
    if (!selectedId && dialog.open) dialog.close();
  }, [selectedId]);
  useLayoutEffect(() => {
    const dialog = purchaseDialogRef.current;
    if (!dialog) return;
    if (purchaseSelectedId && !dialog.open) dialog.showModal();
    if (!purchaseSelectedId && dialog.open) dialog.close();
  }, [purchaseSelectedId]);
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
    const global = !scope || activeQuery !== "";
    if (!storeFilter || !stores.length) {
      setPage({ items: [], has_more: false });
      setLoading(false);
      return () => controller.abort();
    }
    setLoading(true);
    setListError("");
    const requestedSite = /^\d{1,30}$/.test(activeQuery) ? "*" : siteFilter;
    const read = global
      ? searchOrders(storeFilter, requestedSite, requestedSite === "*" ? "*" : logisticFilter,
          activeQuery, statusFilter, purchaseLink, offset, controller.signal)
      : listOrders(scope, statusFilter, purchaseLink, offset, controller.signal);
    void read.then((value) => {
      if (controller.signal.aborted) return;
      if (global && /^\d{1,30}$/.test(activeQuery) && scopeKey !== "all" &&
        value.items.some((record) => record.site_id !== siteFilter ||
          record.logistic_type !== logisticFilter)) {
        selectScope("all");
        return;
      }
      setPage(value);
    }).catch((error: unknown) => {
      if (controller.signal.aborted) return;
      setListError(displayError(error));
      setPage({ items: [], has_more: false });
    }).finally(() => {
      if (!controller.signal.aborted) setLoading(false);
    });
    return () => controller.abort();
  }, [scope?.storeId, scope?.site, scope?.logistic, storeFilter, siteFilter, logisticFilter,
    stores.length, activeQuery, scopeKey, statusFilter, purchaseLink, offset, nonce]);

  useLayoutEffect(() => {
    const controller = new AbortController();
    setDetail(null);
    setDetailError("");
    const target = selectedRowScope ?? scope;
    if (!target || !selectedId) {
      setDetailLoading(false);
      return () => controller.abort();
    }
    setDetailLoading(true);
    void getOrder(target, selectedId, controller.signal).then((result) => {
      if (!controller.signal.aborted) setDetail(result);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setDetailError(displayError(error));
    }).finally(() => {
      if (!controller.signal.aborted) setDetailLoading(false);
    });
    return () => controller.abort();
  }, [scope?.storeId, scope?.site, scope?.logistic,
    selectedRowScope?.storeId, selectedRowScope?.site, selectedRowScope?.logistic, selectedId, nonce]);

  useLayoutEffect(() => {
    const controller = new AbortController();
    purchaseRequest.current = controller;
    setPurchaseReady(false);
    setPurchaseBusy(false);
    setPurchase(null);
    setPurchaseNumber("");
    setPurchaseCost("");
    setWarehouseAddress("");
    setPurchaseTracking("");
    setWarehouseOptions([]);
    setPurchaseError("");
    setPurchaseMessage("");
    const target = purchaseRowScope ?? scope;
    if (!target || !purchaseSelectedId) return () => controller.abort();
    void loadPurchase(target, purchaseSelectedId, controller.signal).then((value) => {
      if (controller.signal.aborted) return;
      setPurchase(value.purchase);
      setPurchaseNumber(value.purchase?.purchase_order_id ?? "");
      setPurchaseCost(value.purchase?.cost_cny ?? "");
      setWarehouseAddress(value.purchase?.warehouse_address ?? "");
      setPurchaseTracking(value.purchase?.purchase_tracking_number ?? "");
      setWarehouseOptions(value.warehouse_addresses ?? []);
      setPurchaseReady(true);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setPurchaseError(displayError(error));
    });
    return () => controller.abort();
  }, [scope, purchaseRowScope?.storeId, purchaseRowScope?.site, purchaseRowScope?.logistic,
    purchaseSelectedId, purchaseReload]);

  const refreshSelectedOrder = async (action: "order" | "shipment" = "order") => {
    const target = selectedRowScope ?? scope;
    if (!target || !selectedId || refreshingOrder || running || logisticsAction || installingUpdate) return;
    const id = selectedId;
    const controller = new AbortController();
    orderRefreshAbort.current?.abort();
    orderRefreshAbort.current = controller;
    setRefreshingOrder(true);
    setDetailError("");
    try {
      const fresh = action === "shipment"
        ? await refreshOrderLogistics(target, id, "shipmentDetails", controller.signal)
        : await refreshOneOrder(target, id, controller.signal);
      if (controller.signal.aborted) return;
      if (shouldApplyOrderRefresh(id, selectedIdRef.current, controller.signal.aborted)) setDetail(fresh);
      setNotice(action === "shipment" ? "已读取订单 " + id + " 的官方物流。" :
        "已从 Mercado Libre 重新读取订单 " + id + " 的最新状态。退款金额与最终净回款仍需单独核账。");
      setPage((previous) => ({ ...previous, items: previous.items.map((record) =>
        record.order_id === fresh.order_id && orderRecordKey(record, scope) ===
          orderRecordKey(fresh, target)
          ? { ...record, ...fresh, shipment: fresh.shipment && { ...fresh.shipment, details: undefined } }
          : record) }));
    } catch (error) {
      if (!controller.signal.aborted) setDetailError(displayError(error));
    } finally {
      if (orderRefreshAbort.current === controller) orderRefreshAbort.current = null;
      if (!controller.signal.aborted) setRefreshingOrder(false);
    }
  };

  const savePurchaseInfo = async () => {
    const signal = purchaseRequest.current?.signal;
    const target = purchaseRowScope ?? scope;
    if (!target || !purchaseSelectedId || !purchaseReady || purchaseBusy || !purchaseNumber || !signal) return;
    setPurchaseBusy(true);
    setPurchaseError("");
    setPurchaseMessage("");
    try {
      const result = await savePurchase(target, purchaseSelectedId, purchaseNumber.trim(), purchaseCost.trim(),
        purchase?.updated_at ?? null, signal, warehouseAddress.trim(), purchaseTracking.trim());
      if (signal.aborted) return;
      setPurchase(result);
      setPurchaseNumber(result.purchase_order_id);
      setPurchaseCost(result.cost_cny ?? "");
      setWarehouseAddress(result.warehouse_address ?? "");
      setPurchaseTracking(result.purchase_tracking_number ?? "");
      setNotice("采购关联已保存，金额为人工录入；未自动推送至货代仓。");
      setPurchaseSelectedId("");
      loadRows();
    } catch (error) {
      if (signal.aborted) return;
      if (error instanceof OrderRequestError && error.status === 409) setPurchaseReady(false);
      setPurchaseError(displayError(error));
    } finally {
      if (!signal.aborted) setPurchaseBusy(false);
    }
  };

  // No local-page pseudo-search: global search and pagination are server-scoped.
  const visibleOrders = page.items;

  useEffect(() => () => { imageAbort.current?.abort(); }, 
    [scope?.storeId, scope?.site, scope?.logistic, statusFilter, purchaseLink, offset]);
  const refreshPageImages = async () => {
    if (!scope || activeQuery || imageBusy || running || installingUpdate || loading) return;
    const snapshot = page.items.filter((order) =>
      order.order_items.some((item) => !item.image_url)).slice(0, 5);
    if (!snapshot.length) {
      setNotice("本页已有可用主图，或平台未返回商品明细。");
      return;
    }
    const controller = new AbortController();
    imageAbort.current = controller;
    setImageBusy(true);
    let completed = 0;
    try {
      for (const order of snapshot) {
        if (controller.signal.aborted) break;
        try {
          const updated = await refreshOrderImage(scope, order.order_id, controller.signal);
          if (controller.signal.aborted) break;
          setPage((previous) => ({ ...previous, items: previous.items.map((item) =>
            item.order_id === updated.order_id ? updated : item) }));
          completed++;
        } catch (error) {
          if (controller.signal.aborted) break;
          setNotice("已补全 " + completed + " 笔订单图片；之后停止：" + displayError(error) +
            "。不自动重试，请核对服务端版本和官方刊登权限。");
          return;
        }
      }
      if (!controller.signal.aborted) {
        setNotice("本次尝试补全 " + completed + " 笔订单的主图。剩余缺图可再次点击；不存在可信图片时仍显示暂无图片。");
      }
    } finally {
      if (imageAbort.current === controller) imageAbort.current = null;
      if (!controller.signal.aborted) setImageBusy(false);
    }
  };
  const selectStatus = (value: string) => {
    setStatusFilter(value);
    setOffset(0);
  };

  useEffect(() => () => {
    slaAbort.current?.abort();
    slaAbort.current = null;
    setLogisticsAction(null);
  }, [scope?.storeId, scope?.site, scope?.logistic, activeQuery, statusFilter, purchaseLink, offset]);

  const refreshPageLogistics = async (action: "sla" | "shipment") => {
    if (slaAbort.current || running || refreshingOrder || installingUpdate || loading) return;
    const isSLA = action === "sla";
    if (isSLA && (!scope || activeQuery ||
        (store?.status !== "connected" && store?.status !== "token_expired"))) return;
    if (isSLA && scope && /^fulfil+ment$/i.test(scope.logistic)) {
      setNotice("Fulfillment 订单不适用发货截止 SLA。");
      return;
    }
    const snapshot = page.items.slice(0, PAGE_SIZE).flatMap((item) => {
      const target = orderRecordScope(item, scope);
      if (!target || !/^[1-9]\d{0,29}$/.test(item.shipment_id ?? "") ||
          (isSLA && (item.status === "cancelled" || item.status === "cancellation_requested"))) return [];
      return [{ item, target }];
    });
    const label = isSLA ? "官方发货截止时间" : "平台物流";
    if (!snapshot.length) {
      setNotice(page.items.some(item => orderRecordScope(item, scope)) || scope
        ? "本页暂无可用于查询" + label + "的 Shipment ID。"
        : "请先选择已验证的店铺、站点和物流类型，或使用跨店铺订单搜索。");
      return;
    }
    const controller = new AbortController();
    slaAbort.current = controller;
    setLogisticsAction(action);
    setSLAProgress("0/" + snapshot.length);
    let done = 0;
    let failed = 0;
    let haveDate = 0;
    let stoppedReason = "";
    try {
      // One user click, up to one page (20 orders), sequential and cancellable.
      // Never run the official SLA fetch automatically in the background.
      for (const { item, target } of snapshot) {
        if (controller.signal.aborted) break;
        try {
          const refreshed = await refreshOrderLogistics(target, item.order_id, action, controller.signal);
          if (controller.signal.aborted) break;
          setPage((previous) => ({ ...previous, items: previous.items.map((order) =>
            orderRecordKey(order, scope) === orderRecordKey(refreshed, target) &&
              order.shipment_id === item.shipment_id && refreshed.shipment_id === item.shipment_id
              ? isSLA ? { ...order, dispatch_deadline: refreshed.dispatch_deadline,
                dispatch_sla_status: refreshed.dispatch_sla_status, dispatch_checked_at: refreshed.dispatch_checked_at }
                : { ...order, shipment: refreshed.shipment }
              : order) }));
          done++;
          if (refreshed.dispatch_deadline) haveDate++;
        } catch (error) {
          if (controller.signal.aborted) break;
          failed++;
          const status = error instanceof StoreRequestError ? error.status : 0;
          const code = error instanceof StoreRequestError ? error.code : "";
          // An upstream 403 may be translated into a sanitized 502 response.
          if (shouldStopLogisticsBatch(status, code)) {
            stoppedReason = displayError(error);
            break;
          }
        }
        setSLAProgress((done + failed) + "/" + snapshot.length);
      }
      if (!controller.signal.aborted && stoppedReason) {
        const unattempted = snapshot.length - done - failed;
        setNotice(label + "读取已停止：" + stoppedReason + "。成功 " + done +
          " 笔，失败 " + failed + " 笔，剩余未尝试 " + unattempted + " 笔；不会自动重试。");
      } else if (!controller.signal.aborted && !failed) {
        setNotice("本页已核验 " + done + " 笔" + label + (isSLA ? "，其中有有效截止时间 " + haveDate + " 笔。" : "。"));
      } else if (!controller.signal.aborted) {
        setNotice(label + "读取成功 " + done + " 笔，未获取 " + failed + " 笔。请在卖家后台核实。");
      }
    } finally {
      if (slaAbort.current === controller) {
        slaAbort.current = null;
        setLogisticsAction(null);
      }
    }
  };

  const performSync = async () => {
    if (!scope || running || logisticsAction || actionBusy || installingUpdate ||
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
  const openOrder = (id: string, targetScope: OrderScope | null = scope) => {
    setSelectedRowScope(targetScope);
    setSelectedId(id);
  };
  const searchOrder = () => {
    const input = searchText.trim();
    if (/^\d{1,5}$/.test(input)) {
      setListError("请输入完整的 Mercado Libre 销售订单号（至少 6 位），数字不支持部分匹配。");
      return;
    }
    if (input && !/^\d{1,30}$/.test(input) && [...input].length < 2) {
      setListError("商品标题搜索至少输入 2 个字符。");
      return;
    }
    if ([...input].length > 120) {
      setListError("搜索内容不能超过 120 字符。");
      return;
    }
    setOffset(0);
    if (activeQuery === input) loadRows();
    else setActiveQuery(input);
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
          <button className="refresh-button" type="button" disabled={!scope || Boolean(activeQuery) || imageBusy || loading || running}
            onClick={() => void refreshPageImages()}>{imageBusy ? "补图中…" : "补全本页主图"}</button>
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
          (storeStatusText[store.status] ?? "状态未知") : "当前查看：所有已验证店铺"}
        {" · "}{marketplaces.length ? "已核实站点 / 物流组合 " + marketplaces.length + " 个" : "无可选站点"}
      </p>
      {store?.status === "token_expired" && <p className="orders-filter-note" role="status">
        美客多访问令牌已到期。点击「同步订单」后服务端将尝试续签；若失败，请到「店铺管理」验证连接。
      </p>}
      {scope && !canSync && !installingUpdate && <p className="orders-filter-note" role="status">{disabledReason}</p>}
      <div className="orders-filters">
        <label>店铺
          <select aria-label="订单店铺" value={storeFilter} onChange={(event) => selectStore(event.target.value)}>
            {stores.length > 1 && <option value="all">所有店铺</option>}
            {stores.map((item) => <option key={item.id} value={item.id}>{item.nickname || item.seller_id}</option>)}
          </select>
        </label>
        <label>站点 / 物流模式
          <select aria-label="站点和物流模式" value={scopeKey} onChange={(event) => selectScope(event.target.value)}>
            {marketplaces.length > 0 && <option value="all">所有站点</option>}
            {marketplaces.map((market) => <option key={market.site_id + market.logistic_type} value={market.site_id + "|" + market.logistic_type}>
              {(SITE_NAMES[market.site_id] ?? market.site_id) + " · " +
                logisticLabel(market.logistic_type)}
            </option>)}
          </select>
        </label>
        <label className="orders-filter-search">搜索订单或商品标题
          <span className="orders-search-inner">
            <input type="search" value={searchText} placeholder="完整销售订单号 / 商品标题（全部订单）"
              onChange={(event) => { setSearchText(event.target.value);
                if (!event.target.value) { setActiveQuery(""); setOffset(0); }
              }}
              onKeyDown={(event) => { if (event.key === "Enter") searchOrder(); }} />
            <button type="button" onClick={searchOrder} title="在所选店铺的全部已同步订单中查找">搜索</button>
          </span>
        </label>
      </div>
      <p className="orders-filter-note">输入完整 Order ID（至少 6 位）跨所选店铺全部站点精确查询；商品标题在当前站点的全部已同步订单中搜索。搜索期间暂不可批量补图或刷新发货截止。{activeQuery && " · 当前搜索：" + activeQuery}</p>

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
          <button type="button" className="orders-sla-update" disabled={!scope || Boolean(activeQuery) || Boolean(logisticsAction) || running || loading || installingUpdate ||
              scope?.logistic.toLowerCase() === "fulfillment" || scope?.logistic.toLowerCase() === "fulfilment" ||
              (store?.status !== "connected" && store?.status !== "token_expired")}
            title="按本页有 Shipment ID 的订单逐笔读取官方 SLA，最多 20 笔，不自动重复"
            onClick={() => void refreshPageLogistics("sla")}>
            {logisticsAction === "sla" ? "读取发货截止 " + slaProgress : "更新本页发货截止"}
          </button>
          <button type="button" className="orders-sla-update"
            disabled={Boolean(logisticsAction) || running || loading || installingUpdate || !page.items.some(order => Boolean(order.shipment_id))}
            onClick={() => void refreshPageLogistics("shipment")}>
            {logisticsAction === "shipment" ? "读取物流 " + slaProgress : "更新本页物流"}
          </button>
        </div>
        <div className="orders-procurement-toolbar" role="group" aria-label="1688采购关联筛选">
          <strong>采购</strong>
          {([
            { value: "", label: "全部" },
            { value: "unlinked", label: "待关联" },
            { value: "linked", label: "已关联" },
          ] as const).map((item) => <button type="button" key={item.value}
            aria-pressed={purchaseLink === item.value}
            onClick={() => { setPurchaseLink(item.value); setOffset(0); }}>
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
          <div className="orders-notice-actions">
            {syncState?.can_reset && !review && !running &&
              <button type="button" disabled={actionBusy} onClick={() => void performReset()}>重扫当前窗口</button>}
            {notice && <button type="button" aria-label="关闭提示" onClick={() => setNotice("")}>关闭</button>}
          </div>
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
                  <th>订单信息</th><th>商品标题 / 规格</th><th>销售金额</th><th>订单状态</th><th>采购 / 平台物流</th><th>发货倒计时</th><th>下单时间</th><th>操作</th>
                </tr></thead>
                <tbody>
                  {loading && page.items.length === 0 && <tr><td colSpan={8} className="orders-table-empty">正在读取订单…</td></tr>}
                  {!loading && visibleOrders.length === 0 && <tr><td colSpan={8} className="orders-table-empty">没有符合条件的订单。可调整筛选或点击同步订单。</td></tr>}
                  {visibleOrders.map((order) => <tr key={orderRecordKey(order, scope)}
                    className={selectedId === order.order_id &&
                      orderRecordKey(order, scope) === [selectedRowScope?.storeId ?? scope?.storeId ?? "",
                        selectedRowScope?.site ?? scope?.site ?? "",
                        selectedRowScope?.logistic ?? scope?.logistic ?? "", selectedId].join("|") ? "selected" : ""}>
                    <td><button type="button" className="orders-id" onClick={() =>
                        openOrder(order.order_id, orderRecordScope(order, scope))}>{order.order_id}</button>
                      <span className="orders-subline">{(
                        stores.find((item) => item.id === (order.store_id ?? scope?.storeId))?.nickname ||
                        store?.nickname || "店铺").trim().slice(0, 12)}（{
                        SITE_NAMES[order.site_id ?? scope?.site ?? ""] ?? order.site_id ?? scope?.site
                      }） · {logisticLabel(order.logistic_type ?? scope?.logistic)}</span>
                    </td>
                    <td className="orders-product-cell">
                      <div className="orders-product-preview">
                        <ProductThumbnail src={order.order_items[0]?.image_url} title={order.order_items[0]?.title || ""} />
                        <div className="orders-product-info">
                          <TranslatedOrderTitle title={order.order_items[0]?.title || "平台未返回商品标题"}
                            translated={translatedTitles.get(order.order_items[0]?.title || "")} />
                          <span className="orders-product-meta">
                            {order.order_items.length > 1 ? order.order_items.length + " 个商品行 · " : ""}
                            {orderCount(order)} 件 · {order.order_items[0]?.item_id || "商品 ID 未返回"}
                          </span>
                          {orderVariationText(order.order_items[0]) && <span className="orders-product-spec"
                            title="官方订单尚未返回可读规格时，仅保留变体 ID">
                            {orderVariationText(order.order_items[0])}
                          </span>}
                        </div>
                      </div>
                    </td>
                    <td className="orders-money-cell"><strong className="orders-money">{orderMoney(order.total_amount, order.currency_id)}</strong>
                      <small title="净收益待官方结算费用数据，尚未核算">净收益：—</small>
                      <small title="利润待实际采购成本、费用和汇率核算">利润：—</small>
                    </td>
                    <td><span className={"orders-status " + safeOrderStatus(order.status)}>{displayOrderStatus(order)}</span>
                      {hasProcurement(order) && <small className="orders-purchase-badge">已采购</small>}
                    </td>
                    <td className="orders-shipping-refs">
                      <span><small>采购运单</small>{order.purchase_tracking_number || "待填写"}</span>
                      <span><small>平台运单</small>{order.shipment?.tracking_number || (order.shipment ? "平台未提供" : "尚未读取")}</span>
                      <span><small>平台物流</small>{shipmentStatusLabel(order.shipment)}</span>
                    </td>
                    <td><DispatchCountdown order={order} now={clockNow}
                      logistic={order.logistic_type ?? scope?.logistic ?? ""} /></td>
                    <td><span className="orders-date">{orderDate(order.date_created)}</span></td>
                    <td><div className="orders-row-actions">
                      <button className="orders-table-action" type="button"
                        onClick={() => openOrder(order.order_id, orderRecordScope(order, scope))}>订单详情</button>
                      <button className="orders-table-action orders-row-purchase" type="button"
                        onClick={() => openPurchase(order.order_id, orderRecordScope(order, scope))}>{order.purchase_order_id ? "编辑采购" : "采购"}</button>
                      {order.purchase_order_id && <small className="orders-row-cost">
                        成本：{order.purchase_cost_cny == null ? "待补录" : "¥" + order.purchase_cost_cny}
                      </small>}
                    </div></td>
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
            onClick={(event) => { if (event.target === event.currentTarget) closeOrder(); }}>
            <div className="orders-detail" aria-label="订单详情">
              <header className="orders-detail-head"><div><small>订单详情 · {SITE_NAMES[selectedRowScope?.site ?? scope?.site ?? ""] ?? selectedRowScope?.site ?? scope?.site}</small>
                <h2 id="orders-detail-title">{selectedId || "订单详情"}</h2></div>
                <div className="orders-detail-head-actions">
                  <button type="button" className="orders-detail-refresh" disabled={refreshingOrder || running || Boolean(logisticsAction) || installingUpdate}
                    onClick={() => void refreshSelectedOrder()}>
                    {refreshingOrder ? "读取中…" : "更新订单状态"}
                  </button>
                  <button type="button" aria-label="关闭订单详情" onClick={closeOrder}>×</button>
                </div>
              </header>
              <div className="orders-modal-scroll">
            {detailLoading && <p className="orders-detail-loading">正在读取最新订单详情…</p>}
            {detailError && <p className="orders-alert" role="alert">{detailError}</p>}
            {detail && <>
              <div className="orders-detail-summary">
                <span className={"orders-status " + safeOrderStatus(detail.status)}>{displayOrderStatus(detail)}</span>
                {hasProcurement(detail) && <small className="orders-purchase-badge">已采购</small>}
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
                      <TranslatedOrderTitle title={line.title || line.item_id} translated={translatedTitles.get(line.title)} detail />
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
                  <dt>物流模式</dt><dd>{logisticLabel(selectedRowScope?.logistic ?? scope?.logistic)}</dd>
                  <dt>采购快递运单</dt><dd>{detail.purchase_tracking_number || "待填写"}</dd>
                  <dt>平台运单号</dt><dd>{detail.shipment?.tracking_number ? <button type="button" onClick={() => copy(detail.shipment?.tracking_number ?? "")}>{detail.shipment.tracking_number}　复制</button> : (detail.shipment ? "平台未提供" : "尚未读取")}</dd>
                  <dt>平台物流状态</dt><dd title={[detail.shipment?.status, detail.shipment?.substatus].filter(Boolean).join(" / ")}>
                    {shipmentStatusLabel(detail.shipment)}</dd>
                  {detail.shipment?.first_mile_carrier && <><dt>首程承运商</dt><dd>{detail.shipment.first_mile_carrier}</dd></>}
                  <dt>官方承运商</dt><dd>{detail.shipment?.details
                    ? (detail.shipment.details.carrier_name || "平台未提供") : "尚未读取"}</dd>
                  <dt>物流查询</dt><dd>{detail.shipment?.details
                    ? (detail.shipment.details.tracking_url
                      ? <a href={detail.shipment.details.tracking_url} target="_blank" rel="noopener noreferrer">打开官方查询链接</a>
                      : "平台未提供") : "尚未读取"}</dd>
                  <dt>官方实际发货</dt><dd>{detail.shipment?.shipped_at
                    ? <time title="本机时区" dateTime={detail.shipment.shipped_at}>{orderDate(detail.shipment.shipped_at)}</time>
                    : (detail.shipment?.details ? "平台未提供时间" : "尚未读取详情")}</dd>
                  <dt>买家签收时间</dt><dd>{detail.shipment?.delivered_at
                    ? <time title="本机时区" dateTime={detail.shipment.delivered_at}>{orderDate(detail.shipment.delivered_at)}</time>
                    : (detail.shipment?.details ? "平台未提供时间" : "尚未读取详情")}</dd>
                  <dt title="本机时区">物流最近核验</dt><dd>{detail.shipment ? orderDate(detail.shipment.checked_at) : "尚未读取"}</dd>
                  {detail.shipment?.last_updated && <><dt title="本机时区">平台物流更新时间</dt><dd>{orderDate(detail.shipment.last_updated)}</dd></>}
                  <dt>官方发货截止</dt><dd title="北京时间">{dispatchDeadlineLabel(detail, clockNow, selectedRowScope?.logistic ?? scope?.logistic).date}
                    {dispatchDeadlineLabel(detail, clockNow, selectedRowScope?.logistic ?? scope?.logistic).hint &&
                      <small> · {dispatchDeadlineLabel(detail, clockNow, selectedRowScope?.logistic ?? scope?.logistic).hint}</small>}</dd>
                </dl>
                <button type="button" className="orders-table-action"
                  disabled={!detail.shipment_id || refreshingOrder || running || Boolean(logisticsAction) || installingUpdate}
                  onClick={() => void refreshSelectedOrder("shipment")}>更新平台物流</button>
              </section>
              {detail.shipment?.details && <section className="orders-detail-block">
                <h3>官方物流时间轴</h3>
                {detail.shipment.details.checked_at && <p className="orders-muted" title="本机时区">详情最近核验：{orderDate(detail.shipment.details.checked_at)}
                  {detail.shipment.checked_at && Date.parse(detail.shipment.checked_at) > Date.parse(detail.shipment.details.checked_at) && " · 概要已更新，详情可能较旧，可手动刷新"}</p>}
                {detail.shipment.details.history?.length ? <ol className="orders-fulfillment-steps">
                  {detail.shipment.details.history.map((event, index) => <li key={index}>
                    <b>{shipmentStatusLabel(event)}</b>
                    <small><time title="本机时区" dateTime={event.date}>{orderDate(event.date)}</time> · {event.status}
                      {event.substatus && " / " + event.substatus}</small>
                  </li>)}
                </ol> : <p className="orders-muted">{detail.shipment.details.history
                  ? "平台暂无物流事件。" : "平台未提供物流历史。"}</p>}
              </section>}
              <section className="orders-detail-block">
                <div className="orders-detail-block-title">
                  <h3>采购</h3>
                  <button type="button" className="orders-table-action"
                    onClick={() => openPurchase(detail.order_id, selectedRowScope ?? scope)}>
                    {detail.purchase_order_id ? "编辑采购" : "采购"}</button>
                </div>
                <p className="orders-muted">采购单：{detail.purchase_order_id || "待关联"}；采购成本：{
                  detail.purchase_cost_cny == null ? "待补录" : "¥" + detail.purchase_cost_cny
                }（CNY）；货代签收及采购快递状态需另行核实。</p>
              </section>
              <section className="orders-detail-block">
                <h3>订单履约流程</h3>
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
                    <li><b>官方物流扫描交接</b><small>{detail.shipment?.shipped_at
                      ? "官方发货：" + orderDate(detail.shipment.shipped_at)
                      : (detail.shipment ? "平台状态：" + shipmentStatusLabel(detail.shipment) + "；交接扫描时间待核实" : "官方物流尚未读取")}</small></li>
                    <li><b>海外买家签收</b><small>{detail.shipment?.status === "delivered"
                      ? (detail.shipment.delivered_at ? orderDate(detail.shipment.delivered_at) : "平台确认已签收，未提供时间")
                      : (detail.shipment ? "平台状态：" + shipmentStatusLabel(detail.shipment) + "；尚无签收确认" : "官方物流尚未读取")}</small></li>
                  </ol>
                  <span>货代仓签收、官方物流接收、买家签收是三个不同事件；当前没有实时状态数据时不显示为已完成。</span>
                </div>
              </section>
            </>}
              </div>
            </div>
          </dialog>
        </div>
        <dialog ref={purchaseDialogRef} className="orders-modal orders-purchase-modal"
          aria-label="1688采购关联" onCancel={(event) => { event.preventDefault(); closePurchase(); }}
          onClose={() => { if (purchaseSelectedId) setPurchaseSelectedId(""); setPurchaseRowScope(null); }}>
          <header className="orders-detail-head">
            <div><small>手动采购信息 · 不自动推送货代仓</small>
              <h2>关联 1688 采购</h2></div>
            <button type="button" aria-label="关闭采购关联" disabled={purchaseBusy}
              onClick={closePurchase}>×</button>
          </header>
          <div className="orders-purchase-compact">
            <p className="orders-muted">美客多订单：{purchaseSelectedId}</p>
            <p className="orders-muted" data-testid="purchase-target-scope">
              店铺：{(() => {
                const current = stores.find((item) => item.id === (purchaseRowScope?.storeId ?? scope?.storeId));
                return current?.nickname || current?.seller_id || "未知店铺";
              })()} · 站点：{SITE_NAMES[purchaseRowScope?.site ?? scope?.site ?? ""] ??
                purchaseRowScope?.site ?? scope?.site ?? "未知站点"} · 物流：{
                logisticLabel(purchaseRowScope?.logistic ?? scope?.logistic)}
            </p>
            <p className="orders-muted">填写实际货代仓和采购运单；仅保存关联，不会向货小易自动推单。</p>
            <div className="orders-purchase-form">
              <label>货代仓收货地址
                <select aria-label="货代仓地址选择" disabled={!purchaseReady || purchaseBusy}
                  value={warehouseOptions.includes(warehouseAddress) ? warehouseAddress : warehouseAddress ? "custom" : ""}
                  onChange={(event) => setWarehouseAddress(event.target.value === "custom" ? "" : event.target.value)}>
                  <option value="">不指定仓库</option>
                  {warehouseOptions.map((address) => <option key={address} value={address}>{address}</option>)}
                  <option value="custom">手动录入新地址</option>
                </select>
                <input type="text" maxLength={300} aria-label="货代仓收货地址"
                  placeholder="填写仓库收货地址，保存后供下次选择"
                  value={warehouseAddress} disabled={!purchaseReady || purchaseBusy}
                  onChange={(event) => setWarehouseAddress(event.target.value)} />
              </label>
              <label>1688 采购订单号
                <input type="text" inputMode="numeric" maxLength={40}
                  placeholder="填写实际采购单号" aria-label="1688 采购订单号"
                  value={purchaseNumber} disabled={!purchaseReady || purchaseBusy}
                  onChange={(event) => setPurchaseNumber(event.target.value)} />
              </label>
              <label>采购金额 CNY（可选）
                <input type="text" inputMode="decimal" maxLength={20}
                  placeholder="例如 16.50" aria-label="人民币采购总金额"
                  value={purchaseCost} disabled={!purchaseReady || purchaseBusy}
                  onChange={(event) => setPurchaseCost(event.target.value)} />
              </label>
              <label>国内采购快递单号（可选）
                <input type="text" maxLength={80} aria-label="国内采购运单号"
                  placeholder="厂家发货后填写实际快递单号"
                  value={purchaseTracking} disabled={!purchaseReady || purchaseBusy}
                  onChange={(event) => setPurchaseTracking(event.target.value)} />
              </label>
              {purchaseError && <p className="orders-alert" role="alert">{purchaseError}</p>}
              {purchaseMessage && <p className="orders-purchase-success" role="status">{purchaseMessage}</p>}
              {purchase && <p className="orders-purchase-state">已关联：{purchase.purchase_order_id}
                {purchase.cost_cny !== null ? " · ¥" + purchase.cost_cny : " · 金额待补"}</p>}
              <div className="orders-purchase-actions">
                <button type="button" className="orders-main-button"
                  disabled={!purchaseReady || purchaseBusy || !/^\d{8,40}$/.test(purchaseNumber)}
                  onClick={() => void savePurchaseInfo()}>{purchaseBusy ? "保存中…" : "保存采购关联"}</button>
                {!purchaseReady && purchaseError && <button type="button"
                  onClick={() => setPurchaseReload((n) => n + 1)}>重新读取采购关联</button>}
                <button type="button" disabled={purchaseBusy} onClick={closePurchase}>取消</button>
              </div>
            </div>
          </div>
        </dialog>
      </>}
    </div>
  );
}
