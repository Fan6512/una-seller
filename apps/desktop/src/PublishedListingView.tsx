import { useMemo, useState } from "react";
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
  secure_thumbnail?: string; pictures?: Picture[];
  seller_custom_field?: string | null; attributes?: Attribute[]; variations?: Variation[];
  description?: string;
};
type Group = { key: string; familyId?: string; members: Listing[] };
type Props = { items: Listing[]; selectedId: string; onSelect: (id: string) => void; site: string; logistic: string; store: string };
type SkuRow = { key: string; sku: string; name: string; pictures: Picture[]; stock?: number | null; fromItem: Listing };

function getFamily(item: Listing): { key: string; id?: string } {
  if (item.family_id != null && String(item.family_id) !== "") {
    return { key: "family:" + String(item.family_id), id: String(item.family_id) };
  }
  if (item.siteless_family_id != null && String(item.siteless_family_id) !== "") {
    return { key: "siteless-family:" + String(item.siteless_family_id), id: String(item.siteless_family_id) };
  }
  return { key: "listing:" + item.id };
}

function getSKU(item: Listing, variation?: Variation): string {
  return variation?.seller_custom_field
    || variation?.attributes?.find((a) => a.id === "SELLER_SKU")?.value_name
    || item.attributes?.find((a) => a.id === "SELLER_SKU")?.value_name
    || item.seller_custom_field || "SKU 未返回";
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
      return [{ key: item.id, sku: getSKU(item), name: item.title || "未命名款式",
        pictures: safePictures(item), stock: item.available_quantity, fromItem: item }];
    }
    return item.variations.map((variant, index) => ({
      key: item.id + ":" + (variant.id ?? index),
      sku: getSKU(item, variant),
      name: (variant.attribute_combinations || []).map((attr) => attr.value_name).filter(Boolean).join(" / ")
        || "规格 " + (index + 1),
      pictures: safePictures(item, variant.picture_ids), stock: variant.available_quantity, fromItem: item,
    }));
  });
}

const STATUS_LABELS: Record<string, string> = { active: "已发布", paused: "已暂停", closed: "已关闭", under_review: "审核中" };
function statusName(status?: string): string {
  return status ? STATUS_LABELS[status] || status : "状态未返回";
}

function priceName(item: Listing): string {
  return item.price != null ? item.price.toFixed(2) + (item.currency_id ? " " + item.currency_id : "") : "未返回";
}

function thumb(item: Listing): string | undefined {
  const pic = safePictures(item)[0];
  return pic?.secure_url || pic?.url;
}

function pictureURL(picture: Picture): string | undefined {
  return picture.secure_url || picture.url;
}

function SkuGallery({ members }: { members: Listing[] }) {
  const rows = skuRows(members);
  return <section className="published-block">
    <div className="published-block-head"><h3>SKU 图片</h3><span>{rows.length} 个 SKU / 款式</span></div>
    <div className="published-sku-list">
      {rows.map((row) => <div className="published-sku-row" key={row.key}>
        <div className="published-sku-meta"><strong>{row.name}</strong><span>{row.sku}</span>
          <small>库存 {row.stock ?? "未获取"}</small></div>
        <div className="published-photo-scroll">
          {row.pictures.length ? row.pictures.map((pic, index) => <div className="published-photo-item" key={row.key + ":" + index}>
            <img src={pictureURL(pic)} alt={(row.name || row.sku) + " 图片 " + (index + 1)}
              referrerPolicy="no-referrer" loading="lazy" />
            <span>{index === 0 ? "主图" : "图片 " + (index + 1)}</span>
          </div>) : <div className="published-no-photo">平台未返回图片</div>}
        </div>
      </div>)}
    </div>
    <p className="published-hint">仅展示平台已返回的图片；没有独立 SKU 图片时显示刊登图片，不执行图片编辑。</p>
  </section>;
}

function Detail({ group, selected, site, logistic, store }: { group: Group; selected: Listing; site: string; logistic: string; store: string }) {
  return <article className="published-detail-card">
    <header className="published-detail-header">
      <div><span className="published-eyebrow">{site} · 已刊登商品</span>
        <h2>{selected.family_name || selected.title || selected.id}</h2>
        <p>{group.key.startsWith("siteless-family:") ? "Siteless Family ID" : "Family ID"}：{group.familyId || "平台未返回"} · 当前可见 {group.members.length} 条刊登</p>
      </div>
      <span className={"published-state " + (selected.status === "active" ? "is-active" : "")}>{statusName(selected.status)}</span>
    </header>
    {selected.detail_unavailable_at && <p className="published-warning">平台当前无法取得详情，下方展示上次保存的信息。</p>}
    <SkuGallery members={group.members}/>
    <PublishedListingEditor key={selected.id + ":" + site} listing={selected} site={site} logistic={logistic} store={store} members={group.members}/>
  </article>;
}

export default function PublishedListingView({ items, selectedId, onSelect, site, logistic, store }: Props) {
  const [mode, setMode] = useState<"family" | "listing">("family");
  const [search, setSearch] = useState("");
  const groups = useMemo<Group[]>(() => {
    const map = new Map<string, Group>();
    for (const item of items) {
      const family = mode === "family" ? getFamily(item) : { key: "listing:" + item.id };
      const group = map.get(family.key);
      if (group) group.members.push(item);
      else map.set(family.key, { key: family.key, familyId: "id" in family ? family.id : undefined, members: [item] });
    }
    const query = search.trim().toLocaleLowerCase();
    return Array.from(map.values()).filter((group) => !query ||
      group.members.some((item) => [item.title, item.id, item.category_id, item.seller_custom_field,
        item.user_product_id, group.familyId].some((field) => String(field || "").toLocaleLowerCase().includes(query))));
  }, [items, mode, search]);
  const active = groups.find((group) => group.members.some((member) => member.id === selectedId)) || groups[0];
  const selected = active?.members.find((item) => item.id === selectedId) || active?.members[0];
  return <div className="published-browser">
    <div className="published-browser-toolbar">
      <div className="published-view-switch" role="group" aria-label="商品查看方式">
        <button type="button" className={mode === "family" ? "active" : ""} aria-pressed={mode === "family"}
          onClick={() => setMode("family")}>家族视图</button>
        <button type="button" className={mode === "listing" ? "active" : ""} aria-pressed={mode === "listing"}
          onClick={() => setMode("listing")}>刊登明细</button>
      </div>
      <label className="published-search"><span>搜索当前页</span>
        <input value={search} onChange={(e) => setSearch(e.target.value)}
          placeholder="标题 / SKU / Family ID / 刊登 ID" /></label>
      <span className="published-browser-count">当前页 {items.length} 条刊登</span>
    </div>
    <div className="published-browser-main">
      <aside className="published-side" aria-label="商品列表">
        <div className="published-side-title">商品列表 <span>{groups.length} 组</span></div>
        {mode === "family" && <p className="published-side-note">仅按平台实际 Family ID 归组；未取得关联的独立显示。</p>}
        <ul className="published-item-list">
          {groups.map((group) => {
            const item = group.members[0];
            const cover = thumb(item);
            return <li key={group.key}><button className={"published-item " + (group === active ? "selected" : "")}
              aria-pressed={group === active} onClick={() => onSelect(item.id)} type="button">
              <span className="published-item-image">{cover
                ? <img src={cover} alt="" loading="lazy" referrerPolicy="no-referrer" />
                : <span>暂无图片</span>}</span>
              <span className="published-item-info">
                <strong title={item.title || item.id}>{item.family_name || item.title || item.id}</strong>
                <small>{group.familyId ? (group.key.startsWith("siteless-family:") ? "Siteless Family " : "Family ") + group.familyId + " · " + group.members.length + " 条刊登"
                  : "刊登 " + item.id}</small>
                <span className="published-item-bottom"><em>{statusName(item.status)}</em><b>{priceName(item)}</b></span>
              </span>
            </button></li>;
          })}
        </ul>
        {groups.length === 0 && <p className="published-empty">当前页没有匹配的商品。</p>}
      </aside>
      <div className="published-detail-wrap">{active && selected
        ? <Detail key={active.key + ":" + selected.id} group={active} selected={selected} site={site} logistic={logistic} store={store}/>
        : <p className="published-empty">选择左侧商品查看详情。</p>}</div>
    </div>
  </div>;
}
