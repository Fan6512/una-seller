import { useEffect, useRef, useState } from "react";
import { listingErrorMessage, listStores, storeRequest, type StoreConnection } from "./stores";
import { runListingBatches, type ListingBatch } from "./listingBatches";

type Attribute = { id: string; name?: string; value_name?: string | null };
type Listing = {
  id: string; partial?: boolean; title?: string; category_id?: string; status?: string;
  price?: number | null; currency_id?: string | null; available_quantity?: number | null;
  cbt_item_id?: string; user_product_id?: string; siteless_user_product_id?: string;
  secure_thumbnail?: string; pictures?: { secure_url?: string }[];
  seller_custom_field?: string | null; attributes?: Attribute[];
  variations?: { attribute_combinations?: Attribute[]; attributes?: Attribute[]; seller_custom_field?: string | null; available_quantity?: number | null }[];
};
type Page = { items: Listing[]; has_more: boolean; synced_at: string | null; next_offset: number | null; scan_required: boolean };
type Progress = Pick<ListingBatch, "processed" | "saved" | "skipped" | "needs_review" | "partial">;
const emptyProgress: Progress = { processed: 0, saved: 0, skipped: 0, needs_review: 0, partial: 0 };
const pageSize = 20;

function message(error: unknown) {
  return error instanceof Error ? error.message : "无法读取已刊登商品";
}

function specs(attributes?: Attribute[]) {
  return attributes?.filter((a) => a.value_name != null).map((a) => (a.name || a.id) + " " + a.value_name).join(" · ");
}

export default function PublishedListingsPage() {
  const [stores, setStores] = useState<StoreConnection[]>();
  const [storeId, setStoreId] = useState("");
  const [scope, setScope] = useState("CBT");
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
  return <>
    <div className="page-heading"><h1>已刊登商品</h1></div>
    {error && <p role="alert">{error}</p>}
    {!stores && !error && <p role="status">正在读取店铺…</p>}
    {stores?.length === 0 && <p>暂无已连接店铺。可在店铺管理查看授权状态。</p>}
    {store && <>
      <label>店铺 <select value={storeId} onChange={(event) => { setStoreId(event.target.value); setScope("CBT"); }}>
        {stores?.map((item) => <option key={item.id} value={item.id}>{item.nickname || item.seller_id}</option>)}
      </select></label>{" "}
      <label>站点 <select value={scope} onChange={(event) => setScope(event.target.value)}>
        <option value="CBT">CBT 全局商品</option>
        {store.marketplaces_status === "verified" && store.marketplaces.map((market) => {
          const value = market.site_id + ":" + market.logistic_type;
          return <option key={value} value={value}>{market.site_id} · {market.logistic_type}</option>;
        })}
      </select></label>
      <ListingScope key={store.id + ":" + scope} store={store.id} scope={scope} />
    </>}
  </>;
}

function ListingScope({ store, scope }: { store: string; scope: string }) {
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

  async function sync(restart: boolean) {
    if (active.current) return;
    const remoteOffset = restart ? 0 : page?.next_offset;
    if (remoteOffset == null) return;
    const controller = new AbortController();
    active.current = controller;
    setBusy(true);
    setError("");
    setNotice("");
    setProgress(emptyProgress);
    try {
      await runListingBatches(remoteOffset, restart, controller.signal, (next, reset, signal) =>
        storeRequest<ListingBatch>(path + "&offset=" + next + (reset ? "&restart=1" : ""), "POST", signal, true),
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
          setNotice("此范围同步完成。");
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
      if (active.current === controller) {
        active.current = null;
        setBusy(false);
        setReading(true);
        setRevision((old) => old + 1);
      }
    }
  }

  const selected = page?.items.find((item) => item.id === selectedId);
  const image = selected?.pictures?.[0]?.secure_url ?? selected?.secure_thumbnail;
  return <>
    <p className="page-actions">
      <button className="primary-action" type="button" onClick={() => void sync(false)}
        disabled={busy || reading || page?.next_offset == null}>{page?.next_offset ? "继续同步" : "同步商品"}</button>
      <button type="button" onClick={() => active.current?.abort()} disabled={!busy}>停止</button>
      <button type="button" onClick={() => void sync(true)} disabled={busy || reading || !page}>从头同步</button>
    </p>
    {progress && <p role="status">
      {busy ? "同步中。" : notice} 本轮已处理 {progress.processed} 条，保存 {progress.saved} 条，跳过 {progress.skipped} 条。
      {progress.needs_review > 0 && " " + progress.needs_review + " 条待人工核实。"}
      {progress.partial > 0 && " " + progress.partial + " 条来自平台部分响应。"}
    </p>}
    {page?.scan_required && <p role="status">已达到常规分页上限，继续读取需要扫描模式。</p>}
    <p>上次保存同步结果 {page?.synced_at ? new Date(page.synced_at).toLocaleString() : "暂无"}</p>
    {error && <p role="alert">{error}</p>}
    {readError && <p role="alert">{readError}</p>}
    {reading && <p role="status">正在读取已保存商品…</p>}
    {!reading && page?.items.length === 0 && <p>此范围暂无已保存商品。</p>}
    {!!page?.items.length && <div className="workspace">
      <section className="list-pane" aria-label="已刊登商品列表"><ul>
        {page.items.map((item) => <li key={item.id}>
          <button type="button" aria-pressed={item.id === selectedId} onClick={() => setSelectedId(item.id)}>
            {item.title || item.id} · {item.status ?? "状态未返回"}
          </button>
        </li>)}
      </ul></section>
      {selected && <section className="detail-pane" aria-label="刊登详情">
        <h2>{selected.title || selected.id}</h2>
        {image?.startsWith("https://") && <img src={image} width="160" loading="lazy" referrerPolicy="no-referrer" alt="" />}
        <dl>
          <dt>刊登编号</dt><dd>{selected.id}</dd>
          <dt>状态</dt><dd>{selected.status ?? "未返回"}</dd>
          <dt>价格</dt><dd>{selected.price != null && selected.currency_id ? selected.price + " " + selected.currency_id : "未返回完整价格"}</dd>
          <dt>库存</dt><dd>{selected.available_quantity ?? "未返回"}</dd>
          {selected.category_id && <><dt>类目</dt><dd>{selected.category_id}</dd></>}
          {site !== "CBT" && <><dt>全局关联</dt><dd>{selected.cbt_item_id || "平台未提供关联"}</dd></>}
          {selected.user_product_id && <><dt>User Product</dt><dd>{selected.user_product_id}</dd></>}
          {selected.siteless_user_product_id && <><dt>UP Siteless</dt><dd>{selected.siteless_user_product_id}</dd></>}
          {selected.partial && <><dt>详情来源</dt><dd>平台返回部分字段</dd></>}
          {selected.seller_custom_field && <><dt>SKU</dt><dd>{selected.seller_custom_field}</dd></>}
        </dl>
        {specs(selected.attributes) && <p>{specs(selected.attributes)}</p>}
        {!!selected.variations?.length && <details open><summary>规格</summary><ul>
          {selected.variations.map((variant, index) => {
            const sku = variant.attributes?.find((a) => a.id === "SELLER_SKU")?.value_name || variant.seller_custom_field;
            return <li key={index}>{specs(variant.attribute_combinations) || "规格 " + (index + 1)}
              {sku && " · SKU " + sku}{variant.available_quantity != null && " · 库存 " + variant.available_quantity}</li>;
          })}
        </ul></details>}
      </section>}
    </div>}
    {(offset > 0 || page?.has_more) && <p className="page-actions">
      <button type="button" disabled={reading || offset === 0} onClick={() => setOffset(offset - pageSize)}>上一页</button>
      <span>第 {offset / pageSize + 1} 页</span>
      <button type="button" disabled={reading || !page?.has_more} onClick={() => setOffset(offset + pageSize)}>下一页</button>
    </p>}
  </>;
}
