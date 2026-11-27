import { Fragment, useLayoutEffect, useMemo, useRef, useState } from "react";
import { expandedRowEnd, galleryColumns } from "./publishedGalleryLayout";
import PublishedListingEditor from "./PublishedListingEditor";
import "./publishedListings.css";

export type Attribute = { id: string; name?: string; value_name?: string | null };
type Picture = { id?: string; secure_url?: string; url?: string };
type Variation = {
  id?: string | number;
  attributes?: Attribute[];
  attribute_combinations?: Attribute[];
  seller_custom_field?: string | null;
  available_quantity?: number | null;
  price?: number | null;
  picture_ids?: string[];
};
export type Listing = {
  id: string; title?: string; category_id?: string; status?: string;
  family_id?: number | string | null; siteless_family_id?: number | string | null; family_name?: string;
  detail_unavailable_at?: string; partial?: boolean;
  price?: number | null; currency_id?: string | null; available_quantity?: number | null;
  cbt_item_id?: string; user_product_id?: string; siteless_user_product_id?: string;
  cbt_category_path?: { original: string[]; zh?: string[] };
  secure_thumbnail?: string; pictures?: Picture[];
  seller_custom_field?: string | null; attributes?: Attribute[]; variations?: Variation[];
  description?: string;
};
type Group = { key: string; familyId?: string; members: Listing[] };
type Props = { items: Listing[]; selectedId: string; onSelect: (id: string) => void; site: string; logistic: string; store: string; allowDiscard: () => boolean; onDirtyChange: (dirty: boolean) => void };
type SkuRow = { key: string; sku?: string; name: string; pictures: Picture[]; stock?: number | null; fromItem: Listing };

function getFamily(item: Listing): { key: string; id?: string } {
  if (item.family_id != null && String(item.family_id) !== "") {
    return { key: "family:" + String(item.family_id), id: String(item.family_id) };
  }
  if (item.siteless_family_id != null && String(item.siteless_family_id) !== "") {
    return { key: "siteless-family:" + String(item.siteless_family_id), id: String(item.siteless_family_id) };
  }
  return { key: "listing:" + item.id };
}

function getSKU(item: Listing, variation?: Variation): string | undefined {
  const sku = variation?.seller_custom_field
    || variation?.attributes?.find((a) => a.id === "SELLER_SKU")?.value_name
    || item.attributes?.find((a) => a.id === "SELLER_SKU")?.value_name
    || item.seller_custom_field || undefined;
  // Some older platform payloads copy the entire title into the custom field.
  return sku?.trim() && sku.trim() !== item.title?.trim() ? sku.trim() : undefined;
}
const optionField = /^(COLOR|MAIN_COLOR|SECONDARY_COLOR|SIZE|SIZE_GRID|MODEL|PATTERN|DESIGN)$/;
function variantName(item: Listing, variation?: Variation): string | undefined {
  const combination = variation?.attribute_combinations?.filter((a) => a.value_name);
  if (combination?.length) return combination.map((a) => a.value_name).join(" / ");
  const options = (variation?.attributes || item.attributes || []).filter((a) => optionField.test(a.id) && a.value_name);
  if (options.length) return options.map((a) => a.value_name).join(" / ");
}

function safePictures(item: Listing, ids?: string[]): Picture[] {
  const all = (item.pictures || []).filter((pic) => {
    const url = pic.secure_url || pic.url || "";
    return url.startsWith("https://");
  });
  const matched = ids?.length ? all.filter((pic) => pic.id && ids.includes(pic.id)) : [];
  const pictures = matched.length ? matched : all;
  if (!pictures.length && item.secure_thumbnail?.startsWith("https://")) {
    return [{ secure_url: item.secure_thumbnail }];
  }
  return pictures.filter((pic, index, list) =>
    list.findIndex((other) => (other.secure_url || other.url) === (pic.secure_url || pic.url)) === index);
}

function skuRows(items: Listing[]): SkuRow[] {
  return items.flatMap((item) => {
    if (!item.variations?.length) {
      return [{ key: item.id, sku: getSKU(item),
        name: variantName(item) || (items.length > 1 ? item.title?.trim() || item.id : "单规格"),
        pictures: safePictures(item), stock: item.available_quantity, fromItem: item }];
    }
    return item.variations.map((variant, index) => ({
      key: item.id + ":" + (variant.id ?? index),
      sku: getSKU(item, variant),
      name: variantName(item, variant) || "款式 " + (index + 1),
      pictures: safePictures(item, variant.picture_ids), stock: variant.available_quantity, fromItem: item,
    }));
  });
}

const STATUS_LABELS: Record<string, string> = { active: "已发布", paused: "已暂停", closed: "已关闭", under_review: "审核中" };
function statusName(status?: string): string {
  return status ? STATUS_LABELS[status] || status : "状态未返回";
}

function thumb(item: Listing): string | undefined {
  if (item.secure_thumbnail?.startsWith("https://")) return item.secure_thumbnail;
  const pic = safePictures(item)[0];
  return pic?.secure_url || pic?.url;
}

function pictureURL(picture: Picture): string | undefined {
  return picture.secure_url || picture.url;
}

function SkuGallery({ members }: { members: Listing[] }) {
  const rows = skuRows(members);
  return <section className="published-block">
    <div className="published-block-head"><h3>款式与图片</h3><span>{rows.length} 个款式</span></div>
    <div className="published-sku-list">
      {rows.map((row) => <div className="published-sku-row" key={row.key}>
        <div className="published-sku-meta"><strong>{row.name}</strong>
          {members.length > 1 && !row.fromItem.variations?.length && row.name !== row.fromItem.id &&
            <small>刊登 {row.fromItem.id}</small>}
          {row.sku && <span>商家 SKU：{row.sku}</span>}
          <small>库存 {row.stock ?? "未获取"}</small></div>
        <div className="published-photo-scroll">
          {row.pictures.length ? row.pictures.map((pic, index) => <div className="published-photo-item" key={row.key + ":" + index}>
            <img src={pictureURL(pic)} alt={row.name + " 图片 " + (index + 1)}
              referrerPolicy="no-referrer" loading="lazy" />
            <span>{index === 0 ? "主图" : "图片 " + (index + 1)}</span>
          </div>) : <div className="published-no-photo">平台未返回图片</div>}
        </div>
      </div>)}
    </div>
    <p className="published-hint">仅展示平台已返回的图片；没有独立 SKU 图片时显示刊登图片，不执行图片编辑。</p>
  </section>;
}

function Detail({ group, selected, site, logistic, store, familyMode, onDirtyChange }: { group: Group; selected: Listing; site: string; logistic: string; store: string; familyMode: boolean; onDirtyChange: (dirty: boolean) => void }) {
  return <article className="published-detail-card">
    <header className="published-detail-header">
      <div><span className="published-eyebrow">{site} · 已刊登商品</span>
        <h2>{selected.title || selected.family_name || selected.id}</h2>
        {familyMode && <p>{group.key.startsWith("siteless-family:") ? "Siteless Family ID" : "Family ID"}：{group.familyId || "平台未返回"} · 本页 {group.members.length} 条刊登</p>}
        <p>刊登 ID：{selected.id}</p>
      </div>
      <span className={"published-state " + (selected.status === "active" ? "is-active" : "")}>{statusName(selected.status)}</span>
    </header>
    {selected.detail_unavailable_at && <p className="published-warning">平台当前无法取得详情，下方展示上次保存的信息。</p>}
    <SkuGallery members={[selected]}/>
    <PublishedListingEditor key={selected.id + ":" + site} listing={selected} site={site} logistic={logistic} store={store} members={group.members} onDirtyChange={onDirtyChange}/>
  </article>;
}

function CardImage({ url, className = "" }: { url?: string; className?: string }) {
  const [failed, setFailed] = useState(false);
  return <span className={"published-card-image " + className}>
    {url && !failed ? <img src={url} alt="" loading="lazy" decoding="async" referrerPolicy="no-referrer"
      onError={() => setFailed(true)}/> : <span>暂无图片</span>}
  </span>;
}

export default function PublishedListingView({ items, selectedId, onSelect, site, logistic, store, allowDiscard, onDirtyChange }: Props) {
  const [mode, setMode] = useState<"family" | "listing">("family");
  const [search, setSearch] = useState("");
  const [expandedKey, setExpandedKey] = useState("");
  const [variantKey, setVariantKey] = useState("");
  const gridRef = useRef<HTMLDivElement>(null);
  const detailRef = useRef<HTMLElement>(null);
  const [columns, setColumns] = useState(1);
  const groups = useMemo<Group[]>(() => {
    const map = new Map<string, Group>();
    for (const item of items) {
      const family = mode === "family" ? getFamily(item) : { key: "listing:" + item.id };
      const group = map.get(family.key);
      if (group) group.members.push(item);
      else map.set(family.key, { key: family.key, familyId: "id" in family ? family.id : undefined, members: [item] });
    }
    return Array.from(map.values());
  }, [items, mode]);
  const visibleGroups = useMemo(() => {
    const query = search.trim().toLocaleLowerCase();
    return groups.filter((group) => !query || group.members.some((item) =>
      [item.title, item.family_name, item.id, item.category_id, getSKU(item), item.user_product_id,
        getFamily(item).id, ...(item.variations || []).map((v) => getSKU(item, v))]
        .some((field) => String(field || "").toLocaleLowerCase().includes(query))));
  }, [groups, search]);
  const active = groups.find((group) => group.members.some((member) => member.id === selectedId));
  const selected = active?.members.find((item) => item.id === selectedId);
  const expanded = mode === "family" ? visibleGroups.find((group) => group.key === expandedKey) : undefined;
  const rowEnd = expandedRowEnd(visibleGroups.findIndex((group) => group === expanded), columns, visibleGroups.length);

  useLayoutEffect(() => {
    const grid = gridRef.current;
    if (!grid) return;
    const measure = () => setColumns(galleryColumns(grid.clientWidth));
    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(grid);
    return () => observer.disconnect();
  }, []);
  useLayoutEffect(() => {
    if (detailRef.current) detailRef.current.scrollTop = 0;
  }, [selectedId]);

  function select(item: Listing, key = item.id) {
    if (item.id !== selectedId && !allowDiscard()) return;
    setVariantKey(key);
    onSelect(item.id);
  }
  function closeDetail() {
    if (!allowDiscard()) return;
    onSelect("");
    setVariantKey("");
  }
  function switchMode(next: "family" | "listing") {
    if (next === mode || !allowDiscard()) return;
    setMode(next);
    setExpandedKey("");
    setVariantKey("");
    onSelect("");
  }

  return <div className="published-browser">
    <div className="published-browser-toolbar">
      <div className="published-view-switch" role="group" aria-label="商品查看方式">
        <button type="button" className={mode === "family" ? "active" : ""} aria-pressed={mode === "family"}
          onClick={() => switchMode("family")}>家族视图</button>
        <button type="button" className={mode === "listing" ? "active" : ""} aria-pressed={mode === "listing"}
          onClick={() => switchMode("listing")}>刊登视图</button>
      </div>
      <label className="published-search"><span>搜索当前页</span>
        <input value={search} onChange={(e) => setSearch(e.target.value)}
          placeholder="标题 / SKU / Family ID / 刊登 ID" /></label>
      <span className="published-browser-count">当前页 {items.length} 条刊登{mode === "family" ? " · " + groups.length + " 组" : ""}</span>
    </div>
    <div className={"published-browser-main" + (selected ? " has-detail" : "")}>
      <section className="published-wall" aria-label="商品图片墙">
        {mode === "family" && <p className="published-side-note">仅按当前页实际 Family ID 归组；未取得关联的独立显示。跨页家族可能不完整。</p>}
        <div className="published-gallery-grid" ref={gridRef} style={{ gridTemplateColumns: "repeat(" + columns + ", minmax(0, 1fr))" }}>
          {visibleGroups.map((group, index) => {
            const item = group.members[0];
            const title = (mode === "family" ? item.family_name : item.title) || item.title || item.id;
            const rows = mode === "family" ? skuRows(group.members) : [];
            const isExpanded = expandedKey === group.key && mode === "family";
            return <Fragment key={group.key}>
              <article className={"published-gallery-card" + (group === active ? " selected" : "") + (isExpanded ? " expanded" : "")}>
                <button type="button" className="published-cover-button" aria-pressed={group === active} onClick={() => select(item)}>
                  <CardImage key={thumb(item)} url={thumb(item)}/>
                  <strong className="published-card-title" title={title}>{title}</strong>
                </button>
                {mode === "family" && <>
                  <div className="published-variant-preview">
                    {rows.slice(0, 4).map((row) => <button type="button" key={row.key}
                      title={row.name} aria-label={"查看 " + row.name} onClick={() => select(row.fromItem, row.key)}>
                      <CardImage key={row.pictures[0] && pictureURL(row.pictures[0])} url={row.pictures[0] && pictureURL(row.pictures[0])}/>
                    </button>)}
                    {rows.length > 4 && <span>+{rows.length - 4}</span>}
                  </div>
                  <button type="button" className="published-expand-family" aria-expanded={isExpanded}
                    onClick={() => setExpandedKey(isExpanded ? "" : group.key)}>
                    {rows.length} 个规格 · {group.members.length} 条刊登 <span aria-hidden="true">{isExpanded ? "▴" : "▾"}</span>
                  </button>
                </>}
              </article>
              {index === rowEnd && expanded && <section className="published-family-expansion" aria-label="当前家族规格">
                <header><div><strong>{expanded.members[0].family_name || expanded.members[0].title || expanded.members[0].id}</strong>
                  <p>{expanded.familyId ? "Family " + expanded.familyId + " · " : ""}当前页 {skuRows(expanded.members).length} 个规格</p></div>
                  <button type="button" onClick={() => setExpandedKey("")}>收起 ×</button></header>
                <div className="published-variant-grid">
                  {skuRows(expanded.members).map((row) => <button type="button" key={row.key}
                    className={"published-variant-card" + (selectedId === row.fromItem.id && (variantKey === row.key || variantKey === row.fromItem.id && !row.fromItem.variations?.length) ? " selected" : "")}
                    onClick={() => select(row.fromItem, row.key)} title={row.name}>
                    <CardImage key={row.pictures[0] && pictureURL(row.pictures[0])} url={row.pictures[0] && pictureURL(row.pictures[0])}/>
                    <strong>{row.name}</strong><small>{row.sku || row.fromItem.id}</small>
                  </button>)}
                </div>
                {expanded.members.some((member) => member.variations?.length) && <p className="published-hint">同一传统刊登内的多个规格共用刊登级详情和本地草稿。</p>}
              </section>}
            </Fragment>;
          })}
        </div>
        {visibleGroups.length === 0 && <p className="published-empty">当前页没有匹配的商品。</p>}
      </section>
      {active && selected && <aside className="published-detail-wrap" ref={detailRef} aria-label="商品详情">
        <div className="published-detail-toolbar"><strong>商品详情</strong>
          <button type="button" aria-label="关闭商品详情" onClick={closeDetail}>关闭 ×</button></div>
        <Detail key={selected.id + ":" + site} group={active} selected={selected} site={site} logistic={logistic} store={store}
          familyMode={mode === "family"} onDirtyChange={onDirtyChange}/>
      </aside>}
    </div>
  </div>;
}
