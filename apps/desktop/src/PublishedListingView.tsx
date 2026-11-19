import { useMemo, useState } from "react";
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
type Props = { items: Listing[]; selectedId: string; onSelect: (id: string) => void; site: string };
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

function statusName(status?: string): string {
  const names: Record<string, string> = { active: "已发布", paused: "已暂停", closed: "已关闭", under_review: "审核中" };
  return status ? names[status] || status : "状态未返回";
}

function priceName(item: Listing): string {
  return item.price != null ? item.price.toFixed(2) + (item.currency_id ? " " + item.currency_id : "") : "未返回";
}

function thumb(item: Listing): string | undefined {
  const pic = safePictures(item)[0];
  return pic?.secure_url || pic?.url;
}

function Field({ label, value }: { label: string; value?: string | number | null }) {
  return <div className="published-field"><span>{label}</span><strong>{value === "" || value == null ? "未获取" : value}</strong></div>;
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

function Detail({ group, selected, site }: { group: Group; selected: Listing; site: string }) {
  const logistics = (selected.attributes || []).filter((attr) =>
    /^(PACKAGE_(WEIGHT|HEIGHT|LENGTH|WIDTH)|WEIGHT|HEIGHT|LENGTH|WIDTH)$/.test(attr.id) && attr.value_name != null);
  const attributes = (selected.attributes || []).filter((attr) =>
    !logistics.includes(attr) && attr.id !== "SELLER_SKU" && attr.value_name != null && attr.value_name !== "");
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
    <section className="published-block">
      <div className="published-block-head"><h3>商品分类</h3></div>
      <div className="published-category"><strong>{selected.category_id || "类目未返回"}</strong>
        <span>完整分类路径尚未获取，暂显示平台类目 ID</span></div>
    </section>
    <section className="published-block">
      <div className="published-block-head"><h3>分类细项与属性</h3><span>按当前刊登返回的真实属性显示</span></div>
      {attributes.length ? <div className="published-field-grid">
        {attributes.map((a, i) => <Field key={a.id + i} label={a.name || a.id} value={a.value_name} />)}
      </div> : <p className="published-hint">平台暂未返回可展示的分类属性。</p>}
      <p className="published-hint">当前为只读展示。必填／选填规则和编辑控件将在接入类目属性接口后提供。</p>
    </section>
    <section className="published-block">
      <div className="published-block-head"><h3>重量与尺寸</h3></div>
      {logistics.length ? <div className="published-field-grid">
        {logistics.map((a, i) => <Field key={a.id + i} label={a.name || a.id} value={a.value_name}/>)}
      </div> : <p className="published-hint">重量和尺寸尚未同步到此刊登的属性中。</p>}
    </section>
    <section className="published-block">
      <div className="published-block-head"><h3>价格与库存</h3></div>
      <div className="published-price-grid">
        <Field label="站点售价" value={priceName(selected)}/>
        <Field label="可售库存" value={selected.available_quantity}/>
        <Field label="预期净回款" value="待接入"/>
        <Field label="优惠价格" value="待接入"/>
      </div>
      {group.members.length > 1 && <details className="published-detail-expand">
        <summary>查看同家族各刊登价格（{group.members.length} 条）</summary>
        <div className="published-family-prices">
          {group.members.map((member) => <div key={member.id}><span>{getSKU(member)} · {member.id}</span>
            <strong>{priceName(member)} · 库存 {member.available_quantity ?? "—"}</strong></div>)}
        </div>
      </details>}
    </section>
    <section className="published-block">
      <div className="published-block-head"><h3>商品标题</h3></div>
      <p className="published-title-text">{selected.title || "平台未返回标题"}</p>
    </section>
    <section className="published-block">
      <div className="published-block-head"><h3>商品详情</h3></div>
      <p className="published-description">{selected.description || "当前商品同步接口尚未获取完整描述，后续接入详情接口后显示。"}</p>
      <details className="published-detail-expand"><summary>查看更多平台标识与状态</summary>
        <div className="published-field-grid">
          <Field label="刊登 ID" value={selected.id}/>
          <Field label="User Product ID" value={selected.user_product_id}/>
          <Field label="全球关联 ID" value={selected.cbt_item_id}/>
          <Field label="Siteless User Product" value={selected.siteless_user_product_id}/>
          <Field label="来源状态" value={selected.partial ? "部分字段" : "已保存"}/>
          <Field label="检查时间" value={selected.detail_unavailable_at ? new Date(selected.detail_unavailable_at).toLocaleString() : "暂无异常"}/>
        </div>
      </details>
    </section>
  </article>;
}

export default function PublishedListingView({ items, selectedId, onSelect, site }: Props) {
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
        ? <Detail key={active.key + ":" + selected.id} group={active} selected={selected} site={site}/>
        : <p className="published-empty">选择左侧商品查看详情。</p>}</div>
    </div>
  </div>;
}
