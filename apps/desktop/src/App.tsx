import { useCallback, useEffect, useState, type FormEvent } from "react";
import {
  apiBase, formatDate, formatMoney, getCollection, getCollectionPage, getHealth, PAGE_SIZE,
  type CollectionDetail, type CollectionSummary,
} from "./collections";

type ServiceStatus = "checking" | "online" | "offline";

function CapturePill({ status }: { status: string }) {
  const label = status === "complete" ? "完整采集" : status === "partial" ? "部分采集" : "状态未知";
  const style = status === "complete" || status === "partial" ? status : "";
  return <span className={"capture-pill " + style}>{label}</span>;
}

function Thumbnail({ src, title, large = false }: { src: string | null; title: string; large?: boolean }) {
  return src
    ? <img className={large ? "product-image product-image-large" : "product-image"} src={src} alt={title} loading="lazy" referrerPolicy="no-referrer" />
    : <span className={large ? "product-image product-image-large image-empty" : "product-image image-empty"} role="img" aria-label="没有商品图片">暂无图片</span>;
}

function Detail({ product }: { product: CollectionDetail }) {
  return (
    <article className="detail">
      <header className="detail-head">
        <div className="detail-head-text">
          <span className="detail-kicker">{product.source_platform.toUpperCase()} · {product.source_site}</span>
          <h2>{product.title}</h2>
          <div className="detail-meta">
            <CapturePill status={product.collection_status} />
            <span>{formatMoney(product.price, product.currency)}</span>
          </div>
        </div>
        <a href={product.source_url} target="_blank" rel="noopener noreferrer">查看来源页面</a>
      </header>

      {product.collection_status === "partial" && <p className="notice">此商品为部分采集，未采集到的字段可能缺失。</p>}
      {product.collection_warnings.length > 0 && (
        <section className="warning-list" aria-label="采集警告">
          <strong>采集提示</strong>
          <ul>{product.collection_warnings.map((warning, index) => <li key={index}>{warning}</li>)}</ul>
        </section>
      )}

      <section aria-labelledby="images-title" className="detail-section">
        <h3 id="images-title">商品图片 <span>{product.images.length}</span></h3>
        {product.images.length ? (
          <div className="image-gallery">
            {product.images.map((image, index) => <Thumbnail key={image.url + index} src={image.url} title={index === 0 ? product.title : product.title + " 图片 " + (index + 1)} large />)}
          </div>
        ) : <p className="empty-inline">尚未采集到商品图片</p>}
      </section>

      <section aria-labelledby="description-title" className="detail-section">
        <h3 id="description-title">商品详情</h3>
        {product.description ? <p className="description-text">{product.description}</p> : <p className="empty-inline">尚未采集到商品详情</p>}
      </section>

      <section aria-labelledby="bullets-title" className="detail-section">
        <h3 id="bullets-title">商品卖点 <span>{product.bullet_points.length}</span></h3>
        {product.bullet_points.length ? <ol className="bullets">{product.bullet_points.map((text, i) => <li key={i}>{text}</li>)}</ol> : <p className="empty-inline">尚未采集到商品卖点</p>}
      </section>

      <section aria-labelledby="variants-title" className="detail-section">
        <h3 id="variants-title">实际采集规格 <span>{product.variants.length}</span></h3>
        {product.variants.length ? (
          <div className="variant-list">
            {product.variants.map((variant, index) => (
              <div className="variant-card" key={index}>
                <div className="variant-head">
                  <strong>{Object.entries(variant.attributes).map(([key, value]) => key + "：" + value).join(" / ")}</strong>
                  <span>{formatMoney(variant.price, variant.currency)}</span>
                </div>
                {variant.image_urls.length > 0 && (
                  <div className="variant-images">
                    {variant.image_urls.map((src, imageIndex) => <Thumbnail key={src + imageIndex} src={src} title={"规格图片 " + (imageIndex + 1)} />)}
                  </div>
                )}
              </div>
            ))}
          </div>
        ) : <p className="empty-inline">未确认任何规格组合，不会自动构造变体</p>}
      </section>
      <footer className="detail-foot">首次采集：{formatDate(product.first_collected_at)}　最近采集：{formatDate(product.last_collected_at)}</footer>
    </article>
  );
}

export default function App() {
  const [service, setService] = useState<ServiceStatus>("checking");
  const [healthText, setHealthText] = useState("检查本地 Go 服务中");
  const [tab, setTab] = useState<"collections" | "connection">("collections");
  const [searchDraft, setSearchDraft] = useState("");
  const [search, setSearch] = useState("");
  const [offset, setOffset] = useState(0);
  const [refresh, setRefresh] = useState(0);
  const [items, setItems] = useState<CollectionSummary[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [listLoading, setListLoading] = useState(true);
  const [listError, setListError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [detail, setDetail] = useState<CollectionDetail | null>(null);
  const [detailLoading, setDetailLoading] = useState(false);
  const [detailError, setDetailError] = useState<string | null>(null);

  const checkHealth = useCallback(async (signal: AbortSignal) => {
    setService("checking");
    try {
      const result = await getHealth(signal);
      if (result.status !== "ok" || result.service !== "una-seller-api") throw new Error("服务响应格式不正确");
      if (!signal.aborted) {
        setService("online");
        setHealthText("Go 服务已连接；数据库状态以采集列表为准");
      }
    } catch (error) {
      if (!signal.aborted) {
        setService("offline");
        setHealthText(error instanceof Error ? error.message : "服务无法连接");
      }
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void checkHealth(controller.signal);
    return () => controller.abort();
  }, [checkHealth, refresh]);

  useEffect(() => {
    const controller = new AbortController();
    setListLoading(true);
    setListError(null);
    void getCollectionPage(search, offset, controller.signal).then((page) => {
      if (controller.signal.aborted) return;
      if (offset > 0 && page.items.length === 0) {
        // Records may have been removed since the last page was loaded.
        setOffset(Math.max(0, offset - PAGE_SIZE));
        return;
      }
      setItems(page.items);
      setHasMore(page.hasMore);
      setSelectedId((old) => page.items.some((item) => item.id === old) ? old : page.items[0]?.id ?? null);
    }).catch((error: unknown) => {
      if (controller.signal.aborted) return;
      setItems([]);
      setHasMore(false);
      setSelectedId(null);
      setListError(error instanceof Error ? error.message : "采集列表无法获取");
    }).finally(() => {
      if (!controller.signal.aborted) setListLoading(false);
    });
    return () => controller.abort();
  }, [search, offset, refresh]);

  useEffect(() => {
    if (!selectedId) {
      setDetail(null);
      setDetailError(null);
      setDetailLoading(false);
      return;
    }
    const controller = new AbortController();
    setDetail(null);
    setDetailError(null);
    setDetailLoading(true);
    void getCollection(selectedId, controller.signal).then((data) => {
      if (!controller.signal.aborted) setDetail(data);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setDetailError(error instanceof Error ? error.message : "商品详情无法获取");
    }).finally(() => {
      if (!controller.signal.aborted) setDetailLoading(false);
    });
    return () => controller.abort();
  }, [selectedId, refresh]);

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setOffset(0);
    setSearch(searchDraft.trim());
  }

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand"><span className="brand-icon">U</span><span>Una Seller</span></div>
        <nav aria-label="主导航">
          <button type="button" className={"nav-item " + (tab === "collections" ? "nav-item-active" : "")} onClick={() => setTab("collections")}>采集箱</button>
          <button type="button" className={"nav-item " + (tab === "connection" ? "nav-item-active" : "")} onClick={() => setTab("connection")}>连接状态</button>
          <div className="nav-item nav-item-disabled">产品库 <span>后续阶段</span></div>
          <div className="nav-item nav-item-disabled">Mercado Libre 刊登 <span>后续阶段</span></div>
        </nav>
        <div className="sidebar-foot"><span className={"status-dot " + service} /> 本地开发 · M0-03</div>
      </aside>

      <main className="content">
        <header className="page-head">
          <div>
            <p className="eyebrow">UNA SELLER · SOURCE COLLECTIONS</p>
            <h1>{tab === "collections" ? "商品采集箱" : "服务连接"}</h1>
            <p className="subtitle">{tab === "collections" ? "真实商品记录来自 Go API，尚未接入浏览器自动采集。" : "本地 API 和数据库需要单独启动与配置。"}</p>
          </div>
          <span className="phase">开发阶段</span>
        </header>
        {tab === "connection" ? (
          <section className="panel">
            <div className="service-row">
              <span className={"status-dot " + service} />
              <div className="service-copy">
                <strong>{service === "online" ? "API 在线" : service === "checking" ? "正在检查" : "API 未连接"}</strong>
                <span role="status">{healthText}</span>
                <code>{apiBase}/healthz</code>
              </div>
              <button type="button" onClick={() => setRefresh((n) => n + 1)}>重新检测</button>
            </div>
            <p className="connection-note">连接正常只说明 /healthz 可用。采集箱必须从 GET /api/v1/collections 实际读到数据库记录，才算采集入库成功。当前不支持公开互联网采集或店铺授权。</p>
          </section>
        ) : (
          <div className="collections-layout">
            <section className="panel collection-list-panel" aria-label="采集商品列表">
              <div className="collection-controls">
                <form onSubmit={submitSearch} className="search-form">
                  <input type="search" aria-label="按商品标题搜索" placeholder="搜索商品标题…" value={searchDraft} onChange={(e) => setSearchDraft(e.target.value)} maxLength={200} />
                  <button type="submit">搜索</button>
                </form>
                <button type="button" onClick={() => setRefresh((n) => n + 1)} disabled={listLoading}>刷新</button>
              </div>
              {listLoading ? <p className="empty-state" role="status">正在读取数据库记录…</p> : listError ? (
                <div className="error-state" role="alert"><strong>无法获取采集箱</strong><p>{listError}</p><small>检查本机 Go API、DATABASE_URL 与 UNA_SELLER_LOCAL_USER_ID。</small></div>
              ) : items.length === 0 ? (
                <div className="empty-state"><strong>{search ? "没有匹配的商品" : "还没有采集商品"}</strong><p>{search ? "请尝试其他标题关键词或清空搜索条件。" : "这里不会显示演示商品。待通过 Go API 保存真实采集记录后，点击刷新即可看到。"}</p></div>
              ) : (
                <div className="collection-items">
                  {items.map((item) => <button className={"collection-item " + (selectedId === item.id ? "selected" : "")} key={item.id} type="button" onClick={() => setSelectedId(item.id)} aria-current={selectedId === item.id ? "true" : undefined}>
                    <Thumbnail src={item.main_image_url} title={item.title} />
                    <span className="item-copy"><strong>{item.title}</strong><span>{item.source_platform.toUpperCase()} · {item.source_site} · {formatMoney(item.price, item.currency)}</span><small>{formatDate(item.last_collected_at)}</small></span>
                    <CapturePill status={item.collection_status} />
                  </button>)}
                </div>
              )}
              <div className="list-footer">
                <span>本页 {items.length} 条 · 每页最多 {PAGE_SIZE} 条</span>
                <div className="pager">
                  <button type="button" disabled={offset === 0 || listLoading} onClick={() => setOffset((n) => Math.max(0, n - PAGE_SIZE))}>上一页</button>
                  <button type="button" disabled={!hasMore || listLoading} onClick={() => setOffset((n) => n + PAGE_SIZE)}>下一页</button>
                </div>
              </div>
            </section>

            <section className="panel collection-detail-panel" aria-label="商品详情">
              {detailLoading ? <p className="empty-state" role="status">正在读取商品详情…</p> : detailError ? <p className="error-state" role="alert">{detailError}</p> : detail ? <Detail product={detail} /> : <p className="empty-state">从左侧选择商品以查看来源详情、图片、卖点和变体。</p>}
            </section>
          </div>
        )}
      </main>
    </div>
  );
}
