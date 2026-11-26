import { useCallback, useEffect, useRef, useState } from "react";
import DesktopSidebar, { desktopSections, type DesktopSection } from "./DesktopSidebar";
import "./desktopShell.css";
import { restoreSession, signOut, type AuthSession } from "./auth";
import Login from "./Login";
import DevelopmentDiagnostics, { diagnosticsEnabled } from "./DevelopmentDiagnostics";
import StoresPage from "./StoresPage";
import AccountSettingsPage from "./AccountSettingsPage";
import TitleModeSettingsPage from "./TitleModeSettingsPage";
import { cachedStores, checkStoreConnection, listStores, watchStoreList } from "./stores";
import { checkStoresOnLogin, resolveManualStoreCheck, storeHealthFromList, type StoreHealth } from "./storeLoginCheck";
import PublishedListingsPage from "./PublishedListingsPage";
import OrdersPage from "./OrdersPage";
import OperationsPage from "./OperationsPage";
import WorkingDraftEditor from "./WorkingDraftEditor";
import OwnProductsPage from "./OwnProductsPage";
import { approveCollection } from "./ownProducts";
import { approvalBlockReason, countApprovalOutcomes, type ApprovalOutcome } from "./workbenchRules";
import DesktopUpdates, { useDesktopUpdates } from "./DesktopUpdates";
import {
  apiBase, deleteCollection, formatDate, formatShortDate, formatMoney, getCollection, getCollectionPage, getWorkingDraft, getHealth, PAGE_SIZE,
  type CollectionDetail, type CollectionSummary,
} from "./collections";

type ServiceStatus = "checking" | "online" | "offline";
type IconName = "search" | "refresh" | "box" | "image" | "external" | "left" | "right" | "warning" | "bell";

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
      {name === "bell" && <><path d="M18 8a6 6 0 0 0-12 0c0 7-3 7-3 9h18c0-2-3-2-3-9M10 21h4" /></>}
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

// The cloud database retains original source URLs for auditability. Displaying a
// low-resolution transform and its high-resolution sibling as separate photos
// makes an image gallery misleading. Collapse display-only duplicates.
function distinctGalleryImages(images: CollectionDetail["images"]): CollectionDetail["images"] {
  const selected: Array<{ image: CollectionDetail["images"][number]; score: number }> = [];
  const seen = new Map<string, number>();
  for (const image of images) {
    let identity = image.url;
    let score = 1;
    try {
      const url = new URL(image.url);
      if (/\.(mp4|webm|m3u8)$/i.test(url.pathname)) continue;
      const match = url.pathname.match(/\/images\/I\/([^/]+?)\.(jpe?g|png|webp)$/i);
      if (match && /^(?:m\.media-amazon\.com|images-(?:na|eu|fe)\.ssl-images-amazon\.com)$/i.test(url.hostname)) {
        if (/(?:play[-_]?(?:button|icon)|video[-_]?(?:icon|thumb|poster)|sprite)/i.test(match[1])) continue;
        const photo = match[1].split("._", 1)[0];
        identity = "/images/I/" + photo;
        const transform = match[1].slice(photo.length);
        const size = transform.match(/_(?:SL|SX|SY|SS|US|UL|UX|UY|SR)(\d+)/i);
        score = size ? Number(size[1]) : transform ? 0 : Number.POSITIVE_INFINITY;
      }
    } catch { /* Keep non-Amazon URLs untouched. */ }
    const position = seen.get(identity);
    if (position === undefined) {
      seen.set(identity, selected.length);
      selected.push({ image, score });
    } else if (score > selected[position].score) {
      selected[position] = { image, score };
    }
  }
  return selected.map(x => x.image);
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

function ProductDetail({ product, onDelete, deleting, deleteError, onDraftDirtyChange, approving }: {
  product: CollectionDetail; onDelete: () => void; deleting: boolean; deleteError: string | null;
  onDraftDirtyChange: (dirty:boolean)=>void; approving: boolean;
}) {
  return (
    <article className="product-detail">
      <div className="detail-toolbar">
        <span className="detail-eyebrow">商品详情</span>
        <div className="detail-actions">
          <button type="button" className="delete-source-button" onClick={onDelete} disabled={deleting || approving}>{deleting ? "删除中…" : "删除商品"}</button>
          <a href={product.source_url} target="_blank" rel="noopener noreferrer" className="source-link">查看原商品 <Icon name="external" size={15} /></a>
        </div>
      </div>
      {deleteError && <p className="delete-source-error" role="alert">{deleteError}</p>}
      <div className="product-top">
        <ProductGallery key={product.id + ":" + product.last_collected_at} images={distinctGalleryImages(product.images)} title={product.title} />
        <div className="product-summary">
          <p className="source-name">{siteName(product.source_platform, product.source_site)}</p>
          <h2>{product.title}</h2>
          <div className="product-price">{formatMoney(product.price, product.currency)}</div>
          {product.category && <p className="date-note">采集来源分类：{product.category}（仅供参考，非美客多正式类目）</p>}
          <div className="product-status"><CapturePill status={product.collection_status} /></div>
          <p className="date-note">最近更新 {formatDate(product.last_collected_at)}</p>
        </div>
      </div>

      {product.own_product_id
        ? <p className="converted-draft-hint" role="status">此商品已移入产品箱，原工作稿停止编辑。请切换到「产品箱」继续修改商品。</p>
        : <WorkingDraftEditor key={product.id} collectionId={product.id}
            sourceTitle={product.title} sourceDescription={product.description} sourceProduct={product} locked={approving} onDirtyChange={onDraftDirtyChange} />}

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

      {product.own_product_id && product.bullet_points.length > 0 && (
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

function CollectionWorkspace({ email, onLogout, updates, beginSync, syncCount }: {
  email: string; onLogout: () => void; updates: ReturnType<typeof useDesktopUpdates>;
  beginSync: () => (() => void) | undefined; syncCount: number;
}) {
  const [section, setSection] = useState<DesktopSection>("operations");
  const [workspaceTab,setWorkspaceTab] = useState<"inbox"|"products">("inbox");
  const [showTransferred,setShowTransferred] = useState(false);
  const [checked,setChecked] = useState<Set<string>>(()=>new Set());
  const [approving,setApproving] = useState(false);
  const [approveNotice,setApproveNotice] = useState("");
  const [draftDirty,setDraftDirty] = useState(false);
  const [ownDirty,setOwnDirty] = useState(false);
  const [settings, setSettings] = useState(false);
  const [settingsSection, setSettingsSection] = useState<"updates" | "stores" | "account" | "titles">("updates");
  const [storeHealth, setStoreHealth] = useState<StoreHealth>({
    status: "checking", message: "正在核实店铺连接状态",
  });
  const failedStoreChecks = useRef(new Set<string>());
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
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const detailCache = useRef(new Map<string, { value: CollectionDetail; savedAt: number }>());

  // One nonblocking verification round per authenticated workspace mount.
  // A later plain GET must never turn a failed POST verification green.
  useEffect(() => {
    let active = true;
    let finished = false;
    const unwatch = watchStoreList((items) => {
      if (active && finished) setStoreHealth(storeHealthFromList(items, failedStoreChecks.current));
    });
    void checkStoresOnLogin(listStores, checkStoreConnection).then(({ health, failedStoreIds }) => {
      if (!active) return;
      failedStoreChecks.current = failedStoreIds;
      finished = true;
      setStoreHealth(health);
    }).catch(() => {
      if (!active) return;
      finished = true;
      failedStoreChecks.current.add("__initial_read_failed__");
      setStoreHealth({
        status: "attention", message: "店铺连接状态暂时无法确认，请检查网络或进入设置重试",
      });
    });
    return () => { active = false; unwatch(); };
  }, []);

  const leavingDirty = () => {
    if (approving) return false;
    if ((workspaceTab==="inbox"&&draftDirty)||(workspaceTab==="products"&&ownDirty)) {
      return window.confirm("当前工作台存在未保存的修改，离开将丢失这些修改。确定继续吗？");
    }
    return true;
  };
  const chooseSource = (id:string) => {
    if (id !== selectedId && !leavingDirty()) return;
    setSelectedId(id);
  };
  const approveSelected=async()=>{
    if (approving || listLoading || deletingId) return;
    const reason=approvalBlockReason(checked.size,draftDirty);
    if(reason){setApproveNotice(reason);return;}
    setApproving(true);setApproveNotice("");
    const outcomes:ApprovalOutcome[]=[];
    const errors:string[]=[];
    const failed=new Set<string>();
    for(const id of checked) {
      try {
        // Read the revision immediately before POST; unknown saves are never
        // retried automatically and a concurrent edit is rejected by the API.
        const {draft}=await getWorkingDraft(id,AbortSignal.timeout(15_000));
        const result=await approveCollection(id,draft?.revision??0,AbortSignal.timeout(20_000));
        outcomes.push(result.result);
      } catch(e) {
        outcomes.push("failed");
        failed.add(id);
        const item=items.find(x=>x.id===id);
        errors.push((item?.title.slice(0,24)??id)+": "+(e instanceof Error && e.name !== "TimeoutError" ? e.message : "转入结果未确认，请刷新产品箱后再决定是否重试"));
      }
    }
    const summary=countApprovalOutcomes(outcomes);
    setChecked(failed);
    setApproveNotice(`本次新转入 ${summary.created} 件；此前已转入 ${summary.alreadyApproved} 件；${summary.failed} 件未确认。`+(errors.length?" "+errors.join("；"):""));
    setApproving(false);
    if(summary.created>0||summary.alreadyApproved>0)reload();
  };
  const reload = () => {
    detailCache.current.clear();
    setRefresh((n) => n + 1);
  };

  async function removeCollection(product: CollectionDetail) {
    if (deletingId) return;
    if (!window.confirm("确定永久删除采集商品“" + product.title + "”吗？采集图片、规格以及该商品尚未转库的人工整理工作稿都会一并永久删除，无法恢复。")) return;
    setDeletingId(product.id);
    setDeleteError(null);
    try {
      await deleteCollection(product.id, new AbortController().signal);
      detailCache.current.delete(product.id);
      setItems((previous) => previous.filter((item) => item.id !== product.id));
      setSelectedId((current) => current === product.id ? null : current);
      setDetail((current) => current?.id === product.id ? null : current);
      if (items.length <= 1 && offset > 0) {
        setOffset(Math.max(0, offset - PAGE_SIZE));
      } else {
        setRefresh((n) => n + 1);
      }
    } catch (error) {
      setDeleteError(error instanceof Error ? error.message : "删除商品失败，请重试");
    } finally {
      setDeletingId(null);
    }
  }

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
    if (isComposing || approving || draftDirty) return;
    const next = searchDraft.trim();
    if (next === search) return;
    const timer = window.setTimeout(() => {
      setOffset(0);
      setChecked(new Set());
      setSearch(next);
    }, 280);
    return () => window.clearTimeout(timer);
  }, [searchDraft, isComposing, search, approving, draftDirty]);

  useEffect(() => {
    const controller = new AbortController();
    setListLoading(true);
    setListError(null);
    void getCollectionPage(search, offset, controller.signal,!showTransferred).then((page) => {
      if (controller.signal.aborted) return;
      if (offset > 0 && page.items.length === 0) {
        setOffset(Math.max(0, offset - PAGE_SIZE));
        return;
      }
      // A new collection timestamp invalidates any previously viewed detail.
      for (const item of page.items) {
        const cached = detailCache.current.get(item.id);
        if (cached && cached.value.last_collected_at !== item.last_collected_at) detailCache.current.delete(item.id);
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
  }, [search, offset, refresh, showTransferred]);

  useEffect(() => {
    setDeleteError(null);
    if (!selectedId) {
      setDetail(null);
      setDetailError(null);
      setDetailLoading(false);
      return;
    }
    const cached = detailCache.current.get(selectedId);
    setDetail(cached?.value ?? null);
    setDetailError(null);
    setDetailLoading(!cached);
    // No new network/Auth round trip when returning to a recently viewed product.
    if (cached && Date.now() - cached.savedAt < 90_000) return;
    const controller = new AbortController();
    void getCollection(selectedId, controller.signal).then((product) => {
      if (controller.signal.aborted) return;
      detailCache.current.set(selectedId, { value: product, savedAt: Date.now() });
      // Bound memory usage: a small, per-login in-memory cache, never persisted.
      if (detailCache.current.size > 16) {
        const oldest = detailCache.current.keys().next().value;
        if (oldest) detailCache.current.delete(oldest);
      }
      setDetail(product);
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
    <div className="app-shell desktop-shell">
      <DesktopSidebar email={email} section={settings ? null : section}
        accountSelected={settings && settingsSection === "account"}
        settingsSelected={settings && settingsSection !== "account"}
        updateAvailable={Boolean(updates.update)}
        onNavigate={target => { if (!leavingDirty()) return; setSettings(false); setSection(target); }}
        onAccount={() => { if (!leavingDirty()) return; setSettings(true); setSettingsSection("account"); }}
        onSettings={() => { if (!leavingDirty()) return; setSettings(true); setSettingsSection("updates"); }} />
      <div className="desktop-content">
      <header className="topbar desktop-toolbar">
        <span className="desktop-toolbar-label">{settings ? (settingsSection === "account" ? "个人中心" : "设置") : desktopSections.find(item => item.id === section)?.label}</span>
        <div className="topbar-spacer" />
        <span className={"connection-indicator " + connectionStatus} role="status">
          <span className="connection-dot" aria-hidden="true" />
          {connectionStatus === "checking" ? "连接中" : connectionStatus === "online" ? "服务正常" : "连接异常"}
        </span>
        <button type="button"
          className={"store-health-button " + (storeHealth.status === "connected" ? "online" : storeHealth.status)}
          title={storeHealth.message}
          onClick={() => { if(!leavingDirty())return;setSettings(true); setSettingsSection("stores"); }}>
          <span className="connection-dot" aria-hidden="true" />
          {storeHealth.status === "checking" ? "店铺检查中"
            : storeHealth.status === "connected" ? "店铺已连接"
              : storeHealth.status === "empty" ? "未绑定店铺" : "店铺需检查"}
        </button>
      </header>

      <main className="main-content">
        {settings && <>
          <nav className="settings-navigation" aria-label="设置分类">
            <button type="button" aria-pressed={settingsSection === "updates"}
              onClick={() => setSettingsSection("updates")}>客户端与更新</button>
            <button type="button" aria-pressed={settingsSection === "stores"}
              onClick={() => setSettingsSection("stores")}>店铺管理</button>
            <button type="button" aria-pressed={settingsSection === "account"}
              onClick={() => setSettingsSection("account")}>账号管理</button>
            <button type="button" aria-pressed={settingsSection === "titles"}
              onClick={() => setSettingsSection("titles")}>刊登与标题</button>
          </nav>
          {settingsSection === "updates" && <>
            <DesktopUpdates updates={updates} syncCount={syncCount} onLater={() => setSettings(false)} />
            {diagnosticsEnabled && <DevelopmentDiagnostics />}
          </>}
          {settingsSection === "account" && <>
            <AccountSettingsPage email={email} />
            <div className="desktop-account-actions"><button type="button" className="logout-action" onClick={() => { if(leavingDirty())onLogout(); }}>退出登录</button></div>
          </>}
          {settingsSection === "titles" && <TitleModeSettingsPage />}
          {settingsSection === "stores" && <StoresPage onConnectionChecked={(id) => {
            const latest = cachedStores();
            if (latest) {
              failedStoreChecks.current = resolveManualStoreCheck(failedStoreChecks.current, latest, id);
              setStoreHealth(storeHealthFromList(latest, failedStoreChecks.current));
            }
          }} />}
        </>}
        {section === "published" ? <div hidden={settings}><PublishedListingsPage beginSync={beginSync} installingUpdate={updates.phase === "installing"} /></div>
          : section === "orders" ? <div hidden={settings}><OrdersPage beginSync={beginSync} installingUpdate={updates.phase === "installing"} visible={!settings} /></div>
          : section === "operations" ? (!settings && <OperationsPage onNavigate={target => { if (!leavingDirty()) return; setSection(target); }} />)
          : section === "afterSales" ? (!settings && <section className="desktop-module-preview">
              <div className="page-heading"><h1>售后管理</h1></div>
              <div className="desktop-module-panel">
                <h2>客服与售后</h2>
                <p>售前问答、买家消息和投诉纠纷管理将逐步开放。目前可在运营中心查看已支持的待办与动态。</p>
                <button type="button" className="refresh-button" onClick={() => { if(leavingDirty())setSection("operations"); }}>查看运营中心</button>
              </div>
            </section>)
          : !settings && <>
        <div className="page-heading">
          <div><h1>工作台</h1></div>
          {workspaceTab==="inbox" && <button type="button" className="refresh-button" onClick={reload} disabled={listLoading || draftDirty || approving} aria-label="刷新采集箱"><Icon name="refresh" size={17} /> 刷新</button>}
        </div>
        <div className="workbench-tabs" role="tablist" aria-label="工作台分区">
          <button type="button" role="tab" aria-selected={workspaceTab==="inbox"} onClick={()=>{if(!leavingDirty())return;setChecked(new Set());setWorkspaceTab("inbox");}}>采集箱</button>
          <button type="button" role="tab" aria-selected={workspaceTab==="products"} onClick={()=>{if(!leavingDirty())return;setChecked(new Set());setWorkspaceTab("products");}}>产品箱</button>
        </div>
        {workspaceTab==="products"? <OwnProductsPage onDirtyChange={setOwnDirty} /> : <>
        <div className="workbench-approval">
          <label><input type="checkbox" checked={showTransferred} disabled={approving}
            onClick={e=>{if(draftDirty&&!leavingDirty())e.preventDefault();}}
            onChange={e=>{setShowTransferred(e.target.checked);setOffset(0);setChecked(new Set());}} /> 包含已转入来源</label>
          <span>已勾选 {checked.size} 件</span>
          <button type="button" className="approval-select-current" disabled={!selectedId||approving||!items.some(item=>item.id===selectedId)}
            aria-pressed={selectedId!==null&&checked.has(selectedId)}
            onClick={()=>{
              if(!selectedId)return;
              setChecked(prev=>{const next=new Set(prev);if(next.has(selectedId))next.delete(selectedId);else next.add(selectedId);return next;});
              setApproveNotice("");
            }}>{selectedId&&checked.has(selectedId)?"取消勾选当前商品":"勾选当前商品"}</button>
          <button type="button" className="primary-button" disabled={approving || listLoading || deletingId !== null} onClick={()=>void approveSelected()}>{approving?"转入中…":"通过并移入产品箱"}</button>
        </div>
        {approveNotice&&<p className="workspace-feedback" role="status">{approveNotice}</p>}
        <div className={"workspace " + (!showDetail ? "workspace-single" : "")}>
          <section className="list-pane" aria-label="采集商品列表">
            <div className="list-toolbar">
              <label className="search-box">
                <Icon name="search" size={18} />
                <input type="search" placeholder="搜索商品名称" aria-label="搜索商品名称" value={searchDraft} disabled={approving || draftDirty} onChange={(event) => setSearchDraft(event.target.value)} onCompositionStart={() => setIsComposing(true)} onCompositionEnd={(event) => { setIsComposing(false); setSearchDraft(event.currentTarget.value); }} maxLength={200} />
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
                <div className="product-card-grid">
                  {items.map((item) => (
                    <article key={item.id} className={"product-card-tile " + (selectedId===item.id?"selected":"")}>
                      <label className="product-card-check" title="勾选待通过商品">
                        <input type="checkbox" aria-label={"选择 "+item.title} checked={checked.has(item.id)} disabled={approving}
                          onChange={event=>{setChecked(prev=>{const next=new Set(prev);if(event.target.checked)next.add(item.id);else next.delete(item.id);return next;});setApproveNotice("");}} />
                      </label>
                      <button type="button" onClick={()=>chooseSource(item.id)} aria-label={"查看采集商品 "+item.title}>
                        <ProductImage src={item.main_image_url} alt={item.title} className="card-image" />
                        <strong title={item.title}>{item.title}</strong>
                        <small>{formatMoney(item.price,item.currency)}</small>
                        <small>{siteName(item.source_platform,item.source_site)} · {formatShortDate(item.last_collected_at)}</small>
                        {item.own_product_id && <span className="converted-card-label">已转入产品箱</span>}
                        {item.collection_status!=="complete"&&<CapturePill status={item.collection_status}/>}
                      </button>
                    </article>
                  ))}
                </div>
              )}
            </div>
            {items.length > 0 && !listError && (
              <div className="list-footer">
                <span>第 {pageNumber} 页</span>
                <div className="page-actions">
                  <button type="button" disabled={offset === 0 || listLoading || draftDirty || approving} aria-label="上一页" onClick={() => {setChecked(new Set());setOffset((n) => Math.max(0, n - PAGE_SIZE));}}><Icon name="left" size={17} />上一页</button>
                  <button type="button" disabled={!hasMore || listLoading || draftDirty || approving} aria-label="下一页" onClick={() => {setChecked(new Set());setOffset((n) => n + PAGE_SIZE);}}>下一页<Icon name="right" size={17} /></button>
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
                <ProductDetail key={detail.id} product={detail} onDelete={() => void removeCollection(detail)} deleting={deletingId !== null} deleteError={deleteError} onDraftDirtyChange={setDraftDirty} approving={approving} />
              )}
            </section>
          )}
        </div>
        </>}
        </>}
      </main>
      </div>
    </div>
  );
}

export default function App() {
  const [session, setSession] = useState<AuthSession | null>(null);
  const [restoring, setRestoring] = useState(true);
  const activity = useRef({ syncs: 0, installing: false });
  const [syncCount, setSyncCount] = useState(0);
  const updates = useDesktopUpdates(activity);

  function beginSync() {
    if (activity.current.installing) return;
    setSyncCount(++activity.current.syncs);
    return () => setSyncCount(--activity.current.syncs);
  }

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
  return <CollectionWorkspace key={session.user.id} email={session.user.email}
    updates={updates} beginSync={beginSync} syncCount={syncCount} onLogout={() => {
    setSession(null);
    void signOut();
  }} />;
}
