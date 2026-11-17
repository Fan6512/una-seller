import { useCallback, useEffect, useState } from "react";
import { restoreSession, signOut, type AuthSession } from "./auth";
import Login from "./Login";
import {
  apiBase, formatDate, formatShortDate, formatMoney, getCollection, getCollectionPage, getHealth, PAGE_SIZE,
  type CollectionDetail, type CollectionSummary,
} from "./collections";

type ServiceStatus = "checking" | "online" | "offline";
type IconName = "search" | "refresh" | "box" | "image" | "external" | "left" | "right" | "warning";

function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
      {name === "search" && <><circle cx="11" cy="11" r="7" /><path d="m16 16 4 4" /></>}
      {name === "refresh" && <><path d="M20 11a8 8 0 0 0-14-5L4 8" /><path d="M4 4v4h4" /><path d="M4 13a8 8 0 0 0 14 5l2-2" /><path d="M20 20v-4h-4" /></>}
      {name === "box" && <><path d="m12 3 9 5-9 5-9-5 9-5Z" /><path d="M3 8v9l9 5 9-5V8M12 13v9" /></>}
      {name === "image" && <><rect x="3" y="4" width="18" height="16" rx="2" /><circle cx="8.5" cy="9" r="1.5" /><path d="m3 17 5-5 4 4 3-3 6 5" /></>}
      {name === "external" && <><path d="M13 5h6v6M19 5l-9 9" /><path d="M19 14v4a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h4" /></>}
      {name === "left" && <path d="m15 18-6-6 6-6" />}
      {name === "right" && <path d="m9 18 6-6-6-6" />}
      {name === "warning" && <><path d="m12 3 10 18H2L12 3Z" /><path d="M12 9v5M12 17h.01" /></>}
    </svg>
  );
}

function siteName(platform: string, site: string) {
  const region: Record<string, string> = { US: "美国站", MLM: "墨西哥站", MLB: "巴西站", MLA: "阿根廷站" };
  const brand = platform === "mercadolibre" ? "Mercado Libre" : platform === "amazon" ? "Amazon" : platform;
  return brand + " · " + (region[site] ?? site);
}

function CapturePill({ status }: { status: string }) {
  const label = status === "complete" ? "已采集" : status === "partial" ? "部分采集" : "状态待确认";
  const style = status === "complete" ? "complete" : status === "partial" ? "partial" : "unknown";
  return <span className={"capture-pill " + style}>{label}</span>;
}

function ProductImage({ src, alt, className = "", eager = false }: { src: string | null; alt: string; className?: string; eager?: boolean }) {
  const [failed, setFailed] = useState(false);
  useEffect(() => setFailed(false), [src]);
  if (!src || failed) {
    return <span className={"image-fallback " + className} role="img" aria-label="暂无商品图片"><Icon name="image" size={24} /></span>;
  }
  return <img className={className} src={src} alt={alt} loading={eager ? "eager" : "lazy"} referrerPolicy="no-referrer" onError={() => setFailed(true)} />;
}

function ProductGallery({ images, title }: { images: CollectionDetail["images"]; title: string }) {
  const [active, setActive] = useState(0);
  if (images.length === 0) {
    return <div className="gallery"><ProductImage src={null} alt={title} className="gallery-main" /></div>;
  }
  const selected = images[active] ?? images[0];
  return (
    <div className="gallery">
      <div className="gallery-stage">
        <ProductImage key={selected.url} src={selected.url} alt={title + "，图片 " + (active + 1)} className="gallery-main" eager />
        {images.length > 1 && (
          <>
            <span className="gallery-count">{active + 1} / {images.length}</span>
            <div className="gallery-arrows">
              <button type="button" aria-label="上一张图片" disabled={active === 0} onClick={() => setActive((n) => Math.max(0, n - 1))}><Icon name="left" size={17} /></button>
              <button type="button" aria-label="下一张图片" disabled={active === images.length - 1} onClick={() => setActive((n) => Math.min(images.length - 1, n + 1))}><Icon name="right" size={17} /></button>
            </div>
          </>
        )}
      </div>
      {images.length > 1 && (
        <div className="gallery-thumbs" aria-label="商品图片列表">
          {images.map((image, index) => (
            <button key={image.url + index} type="button" className={"gallery-thumb " + (index === active ? "active" : "")} aria-label={"查看第 " + (index + 1) + " 张图片"} aria-pressed={index === active} onClick={() => setActive(index)}>
              <ProductImage src={image.url} alt={"图片 " + (index + 1)} />
            </button>
          ))}
        </div>
      )}
    </div>
  );
}

function Description({ value }: { value: string }) {
  const [expanded, setExpanded] = useState(false);
  const long = value.length > 500;
  return (
    <section className="detail-section">
      <h3>商品描述</h3>
      <p className="description-text">{long && !expanded ? value.slice(0, 500) + "…" : value}</p>
      {long && <button type="button" className="text-action" onClick={() => setExpanded((state) => !state)}>{expanded ? "收起" : "展开全部"}</button>}
    </section>
  );
}

function ProductDetail({ product }: { product: CollectionDetail }) {
  return (
    <article className="product-detail">
      <div className="detail-toolbar">
        <span className="detail-eyebrow">商品详情</span>
        <a href={product.source_url} target="_blank" rel="noopener noreferrer" className="source-link">查看原商品 <Icon name="external" size={15} /></a>
      </div>
      <div className="product-top">
        <ProductGallery images={product.images} title={product.title} />
        <div className="product-summary">
          <p className="source-name">{siteName(product.source_platform, product.source_site)}</p>
          <h2>{product.title}</h2>
          <div className="product-price">{formatMoney(product.price, product.currency)}</div>
          <div className="product-status"><CapturePill status={product.collection_status} /></div>
          <p className="date-note">最近更新 {formatDate(product.last_collected_at)}</p>
        </div>
      </div>

      {product.collection_warnings.length > 0 && (
        <details className="capture-notice">
          <summary><Icon name="warning" size={17} /><span>有 {product.collection_warnings.length} 项信息需要留意</span><span className="notice-hint">查看</span></summary>
          <ul>{product.collection_warnings.map((warning, index) => <li key={index}>{warning}</li>)}</ul>
        </details>
      )}

      {product.variants.length > 0 && (
        <section className="detail-section">
          <h3>商品规格 <span className="section-count">{product.variants.length}</span></h3>
          <div className="variant-list">
            {product.variants.map((variant, index) => (
              <div className="variant-card" key={index}>
                <div className="variant-head">
                  <div className="variant-attrs">
                    {Object.entries(variant.attributes).map(([key, value]) => (
                      <span className="variant-attribute" key={key}><small>{key}</small><strong>{value}</strong></span>
                    ))}
                  </div>
                  <strong className="variant-price">{formatMoney(variant.price, variant.currency)}</strong>
                </div>
                {variant.image_urls.length > 0 && (
                  <div className="variant-images">
                    {variant.image_urls.map((src, imageIndex) => <ProductImage key={src + imageIndex} src={src} alt={"规格图片 " + (imageIndex + 1)} />)}
                  </div>
                )}
              </div>
            ))}
          </div>
        </section>
      )}

      {product.bullet_points.length > 0 && (
        <section className="detail-section">
          <h3>商品卖点 <span className="section-count">{product.bullet_points.length}</span></h3>
          <ol className="bullet-list">{product.bullet_points.map((text, index) => <li key={index}>{text}</li>)}</ol>
        </section>
      )}
      {product.description && <Description value={product.description} />}
      <p className="detail-footer">首次采集 {formatDate(product.first_collected_at)}</p>
    </article>
  );
}

function EmptyState({ searching }: { searching: boolean }) {
  return (
    <div className="empty-state">
      <div className="empty-illustration"><Icon name={searching ? "search" : "box"} size={28} /></div>
      <h2>{searching ? "没有找到匹配的商品" : "这里还没有商品"}</h2>
      <p>{searching ? "试试其他关键词，或清空搜索条件。" : "采集的商品会显示在这里。"}</p>
    </div>
  );
}

function LoadError({ detail, onRetry, isDetail = false }: { detail: string; onRetry: () => void; isDetail?: boolean }) {
  return (
    <div className="load-error" role="alert">
      <div className="error-illustration"><Icon name="warning" size={22} /></div>
      <h2>{isDetail ? "商品详情暂时无法打开" : "商品暂时无法加载"}</h2>
      <p>请确认本地服务正常运行，然后重试。</p>
      <button type="button" className="primary-button" onClick={onRetry}>重新加载</button>
      <details className="error-details"><summary>查看错误详情</summary><code>{detail}</code><code>服务地址：{apiBase}</code></details>
    </div>
  );
}

function CollectionWorkspace({ email, onLogout }: { email: string; onLogout: () => void }) {
  const [service, setService] = useState<ServiceStatus>("checking");
  const [serviceError, setServiceError] = useState("");
  const [searchDraft, setSearchDraft] = useState("");
  const [isComposing, setIsComposing] = useState(false);
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

  const reload = () => setRefresh((n) => n + 1);
  const checkHealth = useCallback(async (signal: AbortSignal) => {
    try {
      const result = await getHealth(signal);
      if (result.status !== "ok" || result.service !== "una-seller-api") throw new Error("服务响应格式不正确");
      if (!signal.aborted) {
        setService("online");
        setServiceError("");
      }
    } catch (error) {
      if (!signal.aborted) {
        setService("offline");
        setServiceError(error instanceof Error ? error.message : "无法连接服务");
      }
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void checkHealth(controller.signal);
    return () => controller.abort();
  }, [checkHealth, refresh]);

  useEffect(() => {
    if (isComposing) return;
    const next = searchDraft.trim();
    if (next === search) return;
    const timer = window.setTimeout(() => {
      setOffset(0);
      setSearch(next);
    }, 280);
    return () => window.clearTimeout(timer);
  }, [searchDraft, isComposing, search]);

  useEffect(() => {
    const controller = new AbortController();
    setListLoading(true);
    setListError(null);
    void getCollectionPage(search, offset, controller.signal).then((page) => {
      if (controller.signal.aborted) return;
      if (offset > 0 && page.items.length === 0) {
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
      setListError(error instanceof Error ? error.message : "商品加载失败");
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
    void getCollection(selectedId, controller.signal).then((product) => {
      if (!controller.signal.aborted) setDetail(product);
    }).catch((error: unknown) => {
      if (!controller.signal.aborted) setDetailError(error instanceof Error ? error.message : "商品详情加载失败");
    }).finally(() => {
      if (!controller.signal.aborted) setDetailLoading(false);
    });
    return () => controller.abort();
  }, [selectedId, refresh]);

  const showDetail = items.length > 0 && !listError;
  const connectionStatus: ServiceStatus = listError || serviceError ? "offline" : service;
  const pageNumber = Math.floor(offset / PAGE_SIZE) + 1;

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand"><span className="brand-mark" aria-hidden="true">U</span><span className="brand-name">Una Seller</span></div>
        <span className="topbar-divider" aria-hidden="true" />
        <span className="topbar-label">商品工作台</span>
        <div className="topbar-spacer" />
        <span className={"connection-indicator " + connectionStatus} role="status">
          <span className="connection-dot" aria-hidden="true" />
          {connectionStatus === "checking" ? "连接中" : connectionStatus === "online" ? "服务正常" : "连接异常"}
        </span>
        <span className="account-email" title={email}>{email || "当前用户"}</span>
        <button type="button" className="logout-action" onClick={onLogout}>退出登录</button>
      </header>

      <main className="main-content">
        <div className="page-heading">
          <div><h1>采集箱</h1><p>查看已保存的商品资料</p></div>
          <button type="button" className="refresh-button" onClick={reload} disabled={listLoading} aria-label="刷新商品列表"><Icon name="refresh" size={17} /> 刷新</button>
        </div>

        <div className={"workspace " + (!showDetail ? "workspace-single" : "")}>
          <section className="list-pane" aria-label="采集商品列表">
            <div className="list-toolbar">
              <label className="search-box">
                <Icon name="search" size={18} />
                <input type="search" placeholder="搜索商品名称" aria-label="搜索商品名称" value={searchDraft} onChange={(event) => setSearchDraft(event.target.value)} onCompositionStart={() => setIsComposing(true)} onCompositionEnd={(event) => { setIsComposing(false); setSearchDraft(event.currentTarget.value); }} maxLength={200} />
              </label>
              <div className="list-meta">
                <span>{listLoading ? "正在更新…" : listError ? "加载失败" : search ? "搜索结果" : "全部商品"}</span>
                {!listLoading && !listError && items.length > 0 && <span>{items.length} 件</span>}
              </div>
            </div>

            <div className="list-body" aria-busy={listLoading}>
              {listError ? (
                <LoadError detail={listError} onRetry={reload} />
              ) : listLoading && items.length === 0 ? (
                <div className="loading-state" role="status"><span className="spinner" />正在加载商品…</div>
              ) : items.length === 0 ? (
                <EmptyState searching={Boolean(search)} />
              ) : (
                <div className="product-list">
                  {items.map((item) => (
                    <button type="button" key={item.id} className={"product-row " + (selectedId === item.id ? "selected" : "")} aria-current={selectedId === item.id ? "true" : undefined} onClick={() => setSelectedId(item.id)}>
                      <ProductImage src={item.main_image_url} alt={item.title} className="row-image" />
                      <span className="row-body">
                        <strong className="row-title" title={item.title}>{item.title}</strong>
                        <span className="row-price">{formatMoney(item.price, item.currency)}</span>
                        <span className="row-meta">{siteName(item.source_platform, item.source_site)} · {formatShortDate(item.last_collected_at)}</span>
                        {item.collection_status !== "complete" && <span className="row-badge"><CapturePill status={item.collection_status} /></span>}
                      </span>
                      <span className="row-chevron"><Icon name="right" size={17} /></span>
                    </button>
                  ))}
                </div>
              )}
            </div>
            {items.length > 0 && !listError && (
              <div className="list-footer">
                <span>第 {pageNumber} 页</span>
                <div className="page-actions">
                  <button type="button" disabled={offset === 0 || listLoading} aria-label="上一页" onClick={() => setOffset((n) => Math.max(0, n - PAGE_SIZE))}><Icon name="left" size={17} />上一页</button>
                  <button type="button" disabled={!hasMore || listLoading} aria-label="下一页" onClick={() => setOffset((n) => n + PAGE_SIZE)}>下一页<Icon name="right" size={17} /></button>
                </div>
              </div>
            )}
          </section>

          {showDetail && (
            <section className="detail-pane" aria-label="选中商品的详情">
              {detailError ? (
                <LoadError detail={detailError} onRetry={reload} isDetail />
              ) : detailLoading || !detail || detail.id !== selectedId ? (
                <div className="loading-state detail-loading" role="status"><span className="spinner" />正在加载详情…</div>
              ) : (
                <ProductDetail key={detail.id} product={detail} />
              )}
            </section>
          )}
        </div>
      </main>
    </div>
  );
}

export default function App() {
  const [session, setSession] = useState<AuthSession | null>(null);
  const [restoring, setRestoring] = useState(true);

  useEffect(() => {
    let active = true;
    const sessionLost = () => { if (active) setSession(null); };
    window.addEventListener("una-seller:auth-lost", sessionLost);
    void restoreSession().then((saved) => {
      if (active) setSession(saved);
    }).finally(() => {
      if (active) setRestoring(false);
    });
    return () => {
      active = false;
      window.removeEventListener("una-seller:auth-lost", sessionLost);
    };
  }, []);

  if (restoring) {
    return <div className="login-page"><div className="loading-state" role="status"><span className="spinner" />正在恢复登录…</div></div>;
  }
  if (!session) return <Login onSignedIn={setSession} />;
  return <CollectionWorkspace key={session.user.id} email={session.user.email} onLogout={() => {
    setSession(null);
    void signOut();
  }} />;
}
