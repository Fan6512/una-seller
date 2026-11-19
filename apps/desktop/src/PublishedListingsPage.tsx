import { useEffect, useState } from "react";
import { listStores, storeRequest, type StoreConnection } from "./stores";

type Attribute = { id: string; name?: string; value_name?: string | null };
type Listing = {
  id: string; partial?: boolean; title?: string; category_id?: string; status?: string;
  price?: number | null; currency_id?: string | null; available_quantity?: number | null;
  cbt_item_id?: string; user_product_id?: string; siteless_user_product_id?: string;
  secure_thumbnail?: string; pictures?: { secure_url?: string }[];
  seller_custom_field?: string | null; attributes?: Attribute[];
  variations?: { attribute_combinations?: Attribute[]; attributes?: Attribute[]; seller_custom_field?: string | null; available_quantity?: number | null }[];
};
type Page = { items: Listing[]; has_more: boolean; synced_at: string | null };
type Sync = { saved: number; failed: number; partial: number; next_offset: number | null; scan_required: boolean; platform_status?: number };
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
  const [result, setResult] = useState<Sync>();
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

  async function sync(remoteOffset: number | null | undefined) {
    if (remoteOffset == null) return;
    setBusy(true);
    setError("");
    try {
      setResult(await storeRequest<Sync>(path + "&offset=" + remoteOffset, "POST", undefined, true));
      setRevision((old) => old + 1);
    } catch (e) {
      setError(message(e));
    } finally {
      setBusy(false);
    }
  }

  const selected = page?.items.find((item) => item.id === selectedId);
  const image = selected?.pictures?.[0]?.secure_url ?? selected?.secure_thumbnail;
  return <>
    <p>
      <button className="primary-action" type="button" onClick={() => void sync(0)} disabled={busy}>同步商品</button>{" "}
      {result?.next_offset != null && <button type="button" disabled={busy} onClick={() => void sync(result.next_offset)}>{result.failed ? "重试本批" : "同步下一批"}</button>}
    </p>
    {busy && <p role="status">正在同步此范围的一批商品…</p>}
    {result && !busy && !error && <p role="status">本批保存 {result.saved} 条，失败或跳过 {result.failed} 条。
      {result.partial > 0 && " " + result.partial + " 条来自平台部分响应。"}
      {result.platform_status != null && " 美客多 HTTP " + result.platform_status + "。"}
      {result.scan_required && " 平台要求使用扫描分页继续。"}
    </p>}
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
