import { useEffect, useRef, useState } from "react";
import { listingErrorMessage, listStores, storeRequest, type StoreConnection } from "./stores";
import { runListingBatches, type ListingBatch } from "./listingBatches";
import PublishedListingView, { type Listing } from "./PublishedListingView";

type Review = { listing_id?: string; offset: number; reason: string; skippable: boolean };
type Page = { items: Listing[]; has_more: boolean; synced_at: string | null; next_offset: number | null; scan_required: boolean; review?: Review };
const reviewReasons: Record<string, string> = {
  detail_field_format: "详情字段格式异常",
  detail_payload_unreadable: "详情响应无法解析，归属待核实",
  detail_identity_mismatch: "详情归属与授权范围不一致",
  search_response_invalid: "搜索响应格式异常",
  search_identity_mismatch: "搜索归属与授权范围不一致",
};
type Progress = Pick<ListingBatch, "processed" | "saved" | "skipped" | "needs_review" | "partial">;
const emptyProgress: Progress = { processed: 0, saved: 0, skipped: 0, needs_review: 0, partial: 0 };
const pageSize = 20;
type BeginSync = () => (() => void) | undefined;

function message(error: unknown) {
  if (error instanceof TypeError || (error instanceof Error && error.message === "Failed to fetch")) {
    return "网络请求未完成，已保存的商品不会丢失。可检查网络后继续同步。";
  }
  return error instanceof Error ? error.message : "无法读取已刊登商品";
}

export default function PublishedListingsPage({ beginSync, installingUpdate }: {
  beginSync: BeginSync; installingUpdate: boolean;
}) {
  const [stores, setStores] = useState<StoreConnection[]>();
  const [storeId, setStoreId] = useState("");
  const [scope, setScope] = useState("");
  const [error, setError] = useState("");
  useEffect(() => {
    const controller = new AbortController();
    listStores(controller.signal).then((items) => {
      if (controller.signal.aborted) return;
      const eligible = items.filter((store) => store.status === "connected" || store.status === "token_expired");
      setStores(eligible);
      setStoreId(eligible[0]?.id ?? "");
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) setError(message(e));
    });
    return () => controller.abort();
  }, []);

  const store = stores?.find((item) => item.id === storeId);
  const markets = store?.marketplaces_status === "verified"
    ? store.marketplaces.filter((market) => market.site_id && market.logistic_type) : [];
  const scopeValue = markets.some((market) => market.site_id + ":" + market.logistic_type === scope)
    ? scope : markets[0] ? markets[0].site_id + ":" + markets[0].logistic_type : "";
  const siteLabels: Record<string, string> = { MLM: "墨西哥 MX", MLB: "巴西 BR", MLA: "阿根廷 AR", MLC: "智利 CL", MCO: "哥伦比亚 CO" };
  return <div className="published-page">
    <div className="page-heading">
      <div><h1>已刊登商品</h1><p>按站点管理已同步的商品与 SKU</p></div>
    </div>
    {error && <p role="alert" className="published-alert">{error}</p>}
    {!stores && !error && <p role="status">正在读取店铺…</p>}
    {stores?.length === 0 && <p className="published-empty">暂无已连接店铺。请先在店铺管理检查授权状态。</p>}
    {store && <>
      <div className="published-top-filters">
        <label>店铺
          <select value={storeId} onChange={(event) => { setStoreId(event.target.value); setScope(""); }}>
            {stores?.map((item) => <option key={item.id} value={item.id}>{item.nickname || item.seller_id}</option>)}
          </select>
        </label>
        <label>站点
          <select value={scopeValue} onChange={(event) => setScope(event.target.value)} disabled={markets.length === 0}>
            {markets.map((market) => {
              const value = market.site_id + ":" + market.logistic_type;
              return <option key={value} value={value}>{siteLabels[market.site_id] || market.site_id} · {market.logistic_type}</option>;
            })}
          </select>
        </label>
        <span className="published-filter-hint">每个站点的数据独立显示</span>
      </div>
      {scopeValue
        ? <ListingScope key={store.id + ":" + scopeValue} store={store.id} scope={scopeValue}
            beginSync={beginSync} installingUpdate={installingUpdate} />
        : <p className="published-empty">该店铺尚无已验证的站点，请到店铺管理检查站点关联。</p>}
    </>}
  </div>;
}

function ListingScope({ store, scope, beginSync, installingUpdate }: {
  store: string; scope: string; beginSync: BeginSync; installingUpdate: boolean;
}) {
  const [offset, setOffset] = useState(0);
  const [revision, setRevision] = useState(0);
  const [page, setPage] = useState<Page>();
  const [selectedId, setSelectedId] = useState("");
  const [progress, setProgress] = useState<Progress>();
  const [notice, setNotice] = useState("");
  const active = useRef<AbortController | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [readError, setReadError] = useState("");
  const [reading, setReading] = useState(true);
  const [site, logistic = ""] = scope.split(":");
  const path = "/api/v1/stores/" + encodeURIComponent(store) + "/published-listings?" +
    new URLSearchParams({ site, logistic_type: logistic }).toString();

  useEffect(() => {
    const controller = new AbortController();
    setReading(true);
    setReadError("");
    storeRequest<Page>(path + "&offset=" + offset, "GET", controller.signal, true).then((next) => {
      if (controller.signal.aborted) return;
      setPage(next);
      setSelectedId((old) => next.items.some((item) => item.id === old) ? old : next.items[0]?.id ?? "");
    }).catch((e: unknown) => {
      if (!controller.signal.aborted) {
        setPage(undefined);
        setReadError(message(e));
      }
    }).finally(() => {
      if (!controller.signal.aborted) setReading(false);
    });
    return () => controller.abort();
  }, [path, offset, revision]);

  useEffect(() => () => {
    active.current?.abort();
    active.current = null;
  }, []);

  async function sync(restart: boolean, skip?: Review) {
    if (active.current) return;
    const remoteOffset = skip?.offset ?? (restart ? 0 : page?.next_offset);
    if (remoteOffset == null) return;
    const finishSync = beginSync();
    if (!finishSync) return;
    const controller = new AbortController();
    active.current = controller;
    setBusy(true);
    setError("");
    setNotice("");
    setProgress(emptyProgress);
    try {
      await runListingBatches(remoteOffset, restart, controller.signal, (next, reset, signal) => {
        const confirmSkip = skip && next === skip.offset;
        return storeRequest<ListingBatch>(
          path + "&offset=" + next + (reset ? "&restart=1" : "") + (confirmSkip ? "&skip_invalid=1" : ""),
          "POST", signal, true, confirmSkip ? JSON.stringify({ listing_id: skip.listing_id }) : undefined,
        );
      },
      (batch) => {
        setProgress((previous) => {
          const total = previous ?? emptyProgress;
          return {
            saved: total.saved + batch.saved, processed: total.processed + batch.processed,
            skipped: total.skipped + batch.skipped, needs_review: total.needs_review + batch.needs_review,
            partial: total.partial + batch.partial,
          };
        });
        if (batch.interruption) {
          const interruption = batch.interruption;
          setError(listingErrorMessage(interruption.code, interruption.platform_status));
        } else if (batch.next_offset == null && !batch.scan_required) {
          setNotice("已到本轮搜索末尾。");
        } else {
          setNotice("已保存进度，可继续下一轮。");
        }
      });
    } catch (e) {
      if (active.current === controller) {
        if (controller.signal.aborted) setNotice("已停止，可从服务端保存的位置继续。");
        else setError(message(e));
      }
    } finally {
      finishSync();
      if (active.current === controller) {
        active.current = null;
        setBusy(false);
        setReading(true);
        setRevision((old) => old + 1);
      }
    }
  }

  const review = page?.review;
  return <>
    <div className="published-sync-bar">
    <p className="page-actions">
      <button className="primary-action" type="button" onClick={() => void sync(false)}
        disabled={installingUpdate || busy || reading || page?.next_offset == null}>{review ? "重新核实并继续" : page?.next_offset ? "继续同步" : "同步商品"}</button>
      <button type="button" onClick={() => active.current?.abort()} disabled={!busy}>停止</button>
      <button type="button" onClick={() => void sync(true)} disabled={installingUpdate || busy || reading || !page}>从头同步</button>
    </p>
    {installingUpdate && <p role="status">正在启动更新安装，重启后可同步商品。</p>}
    {review && <p role="status">
      待核实 · {review.listing_id ? "刊登 " + review.listing_id + " · " : ""}位置 {review.offset + 1} · {reviewReasons[review.reason] ?? review.reason}{" "}
      {review.skippable && review.listing_id && <button type="button" disabled={installingUpdate || busy || reading} onClick={() => {
        if (window.confirm("若刊登 " + review.listing_id + " 复核后仍有格式异常，将记录跳过并继续。确认吗？")) {
          void sync(false, review);
        }
      }}>确认跳过并继续</button>}
    </p>}
    {progress && <p role="status">
      {busy ? "同步中。" : notice} 本轮已处理 {progress.processed} 条，保存 {progress.saved} 条，跳过 {progress.skipped} 条。
      {progress.needs_review > 0 && " " + progress.needs_review + " 条待人工核实。"}
      {progress.partial > 0 && " " + progress.partial + " 条来自平台部分响应。"}
    </p>}
    {page?.scan_required && <p role="status">已达到常规分页上限，继续读取需要扫描模式。</p>}
    <p className="published-sync-time">已保存同步位置 {page?.next_offset ?? "已结束"} · 最近成功 {page?.synced_at ? new Date(page.synced_at).toLocaleString() : "暂无"}</p>
    </div>
    {error && <p role="alert">{error}</p>}
    {readError && <p role="alert">{readError}</p>}
    {reading && <p role="status">正在读取已保存商品…</p>}
    {!reading && page?.items.length === 0 && <p>此范围暂无已保存商品。</p>}
    {!!page?.items.length && <PublishedListingView
      items={page.items} selectedId={selectedId} onSelect={setSelectedId} site={site}
    />}
    {(offset > 0 || page?.has_more) && <p className="page-actions">
      <button type="button" disabled={reading || offset === 0} onClick={() => setOffset(offset - pageSize)}>上一页</button>
      <span>第 {offset / pageSize + 1} 页</span>
      <button type="button" disabled={reading || !page?.has_more} onClick={() => setOffset(offset + pageSize)}>下一页</button>
    </p>}
  </>;
}
