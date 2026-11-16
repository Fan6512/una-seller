# Una Seller 落地清单（源自智赢 ERP 客户端逆向）

> **文档定位**：本研究系列（`01`–`09`）的最终交付物。把逆向观察转成可执行的工程条目，直接对接现有 `docs/architecture/ERP-MVP.md` v0.3 与 `docs/product/BROWSER-COLLECTION-PLUGIN-MVP.md` v0.3 的 PR-B → PR-C → PR-D 顺序。
>
> **本文不改任何任务书、不建 PR、不合并。** 是输入材料，不是新任务书。
>
> **独立实现声明**：以下所有表结构、字段命名、API 形态、目录布局均为**按功能要求重新设计**，不复制第三方专有代码、不沿用其命名风格、不复现其协议格式。
>
> **证据等级**：本文结论标注 `OBSERVED` / `RUNTIME_VERIFIED` / `INFERRED`。凡属设计建议的部分标注 `DESIGN`，不是逆向结论。

---

## 0. 一页纸：抄什么、怎么改、不抄什么

| # | 智赢的做法 | 证据 | 我们的决定 |
|---|---|---|---|
| 1 | 图片按 `主图 / 详情图 / 变体主图 / 变体图集 / 色板` 分槽位 | `RUNTIME_VERIFIED`：`sale_main`、`sale_pic`、`sale_picv`、`sale_imgsv`、`sale_swatch`、`sale_video` | **照抄分槽思路**，但改成 `product_image.role` 枚举 + `position` 排序，不散落成 6 个列 |
| 2 | 变体维度字段统一加 `v` 后缀（`picv`/`imgsv`/`logv`/`skusv`/`transv`） | `RUNTIME_VERIFIED`（内存字段表） | **不抄后缀**（可读性差），但**抄它的语义边界**：变体级数据必须与商品级数据物理分开 |
| 3 | 五点卖点与详细描述全程分离：`_bullets` / `_intros` → `txtBullet` / `txtIntro` → `sale_bullet` / `sale_intro` | `OBSERVED` 四层闭环（插件 → `SpData` → WinForms 文本框 → 内存字段表） | **照抄**。这是最该抄的一条，见 §3.2 |
| 4 | `cache/` 与 `image/<商品ID>/` 两套存储、两个生命周期 | `RUNTIME_VERIFIED`：浏览列表只写 `cache/`（666 文件），进编辑态才写 `image/`（3 目录 39 文件） | **照抄隔离原则**，但本地只做"缓存"，不做"工作区"（见 §4.4） |
| 5 | 主图用内容哈希命名（`5c633a2f…jpg`），变体图用变体值命名（`DEEP.jpg` / `ESPRESSO.jpg`） | `RUNTIME_VERIFIED`（`image/818211505/`） | **照抄**，见 §4.3 |
| 6 | 采集状态 / 缺失字段 / 失败原因是一等公民 | `OBSERVED`：`Enumer.Sale.Stat/Flag/Grade/Level`、`采集失败`、`检查详细页重复`、`清理重复` | **照抄**，做成 `status` + `missing_fields[]` + `warnings[]` 三态响应 |
| 7 | 列表 + 明细 + 弹窗三层信息架构 | `OBSERVED`：`亚马逊采集产品列表` / `采集产品明细` / `Dialog.*` | **照抄**，但详情用整页不用弹窗（见 §5.2） |
| 8 | 类目属性字典客户端内置（`.dat` 共 100 MB+，gzip 未加密） | `RUNTIME_VERIFIED`：`ZForm` 2.6 MB / 48,828 条可选值；`ZData` 29 MB；`ZAmazon` 14.5 MB；`ZOther` 53 MB | **不抄**。字典必须服务端下发 + 版本号，客户端只缓存最近用过的类目，见 §6 |
| 9 | 92 个扁平 `sale_*` 字段放一张表 | `RUNTIME_VERIFIED`（客户端 92 个，插件 26 个） | **不抄**。拆 7 张表，见 §2 |
| 10 | 客户端内置 CEF 爬虫 + 私有加密 RPC + 客户端持有平台 Token | `RUNTIME_VERIFIED`：`libcef.dll` 485 MB、`ZClient` XOR 帧、`/Meli/Authback` 回调 | **不抄**。采集在插件侧，平台 Token 只在服务端 |
| 11 | 图片走自有代理网关（`seller.zying.com/image/bridge?url=` + OSS 缩略） | `RUNTIME_VERIFIED`：543 次域名命中 | **首版不抄**（直接存源 URL），P2 再评估，见 §4.5 |
| 12 | 插件由客户端静默分发（13 个 `.crx`，5.1.5 → 5.2.0） | `RUNTIME_VERIFIED`：`%LOCALAPPDATA%\ZYing\` | **不抄**。插件走商店 / 手动安装，不做静默下发 |

---

## 1. 结论边界（先说清楚这份清单能管到哪）

**能管的**：数据模型、API 契约、图片与文件组织、详情页信息架构、状态与错误语义、类目属性的组织方式。这些全部有 `OBSERVED` 或 `RUNTIME_VERIFIED` 支撑。

**不能管的**：智赢云端库表结构、RPC 命令名清单、平台刊登的真实报错码、变体生成的业务规则。这些是服务端实现，逆向拿不到，也不该拿——Una Seller 要用自己的规则。

**一句话取舍**：**抄它的信息组织，不抄它的通信和分发。**

---

## 2. 数据层：把 92 个扁平字段拆成 7 张表

### 2.1 为什么不照抄单表

`RUNTIME_VERIFIED` 客户端是 92 个 `sale_*` 扁平字段。这个设计在 WinForms + `RecordSet` 列式表格下是合理的（一行一列直接绑 DataGrid），但对 PostgreSQL + 多平台 + 多变体是灾难：

- 变体数据（`sale_picv`/`sale_imgsv`/`sale_skusv`/`sale_transv`/`sale_logv`）和商品数据混在一行 → 一变体一列，无法表达"4 个变体"
- 图片是固定列不是集合 → 无法排序、无法区分主图与第 7 张图
- 多语言文本没有位置 → 五点卖点存成一个大字符串，取第 3 条要靠 `split('\n')`
- 任何平台新增字段都要改表

### 2.2 七张表（`DESIGN`）

```sql
-- ① 采集条目：一次采集 = 一行，保留原始快照，可重放
create table collection_item (
  id                bigserial primary key,
  workspace_id      uuid not null,
  source_platform   text not null,          -- 'amazon' | 'mercadolibre'
  source_site       text not null,          -- 'US' | 'MLM' | 'MLB' | 'MLA'
  source_item_id    text not null,          -- ASIN / MLM123456
  source_url        text not null,
  idempotency_key   text not null,          -- sha256(platform|site|item_id|seller_ref)
  status            text not null,          -- 'new' | 'updated' | 'partial' | 'failed' | 'duplicate'
  missing_fields    text[] not null default '{}',
  warnings          jsonb  not null default '[]',
  raw_payload       jsonb,                  -- 插件原始提交，永不修改
  collected_at      timestamptz not null default now(),
  unique (workspace_id, idempotency_key)
);
create index on collection_item (workspace_id, status, collected_at desc);

-- ② 商品主记录：从采集条目晋升，可被多次采集更新
create table product (
  id                 bigserial primary key,
  workspace_id       uuid not null,
  collection_item_id bigint references collection_item(id),
  title              text,
  brand              text,
  model_number       text,
  condition_code     text,                  -- 'new' | 'used_like_new' …
  currency           char(3),
  price              numeric(14,2),
  cost               numeric(14,2),
  weight_g           integer,
  length_mm          integer, width_mm integer, height_mm integer,
  status             text not null default 'draft',   -- draft|collecting|ready|listed|archived
  source_platform    text not null,
  source_site        text not null,
  version            integer not null default 1,
  created_at         timestamptz not null default now(),
  updated_at         timestamptz not null default now()
);

-- ③ 图片：一张表管所有槽位
create table product_image (
  id            bigserial primary key,
  product_id    bigint not null references product(id) on delete cascade,
  role          text not null,              -- 'main' | 'detail' | 'variant_main' | 'variant_gallery' | 'swatch' | 'video_poster'
  variant_key   text,                       -- role 为 variant_* 时必填，如 'DEEP' / 'ESPRESSO'
  position      smallint not null default 0,
  source_url    text not null,
  content_hash  char(64),                   -- sha256，落盘后回填
  width_px      integer, height_px integer,
  bytes         bigint,
  storage_key   text,                       -- 对象存储 key
  created_at    timestamptz not null default now(),
  unique (product_id, role, coalesce(variant_key,''), position)
);
create index on product_image (product_id, role, position);

-- ④ 文本：五个字段 × 多语言 × 有序
create table product_text (
  id          bigserial primary key,
  product_id  bigint not null references product(id) on delete cascade,
  field       text not null,                -- 'title'|'brief'|'keys'|'bullet'|'intro'|'memo'
  locale      text not null default 'src',  -- 'src'|'zh-CN'|'en-US'|'es-MX'
  position    smallint not null default 0,  -- bullet 是列表，必须有序
  value       text not null,
  unique (product_id, field, locale, position)
);

-- ⑤ 变体轴：颜色 / 尺码 …
create table variant_axis (
  id          bigserial primary key,
  product_id  bigint not null references product(id) on delete cascade,
  name        text not null,                -- 'Color' / 'Size'
  position    smallint not null default 0,
  unique (product_id, name)
);

-- ⑥ 变体：一个组合一行
create table variant (
  id            bigserial primary key,
  product_id    bigint not null references product(id) on delete cascade,
  combination   jsonb not null,             -- {"Color":"DEEP","Size":"M"}
  sku           text,
  price         numeric(14,2),
  stock         integer not null default 0,
  is_active     boolean not null default true,
  unique (product_id, sku)
);
create index on variant using gin (combination);

-- ⑦ 类目绑定：平台 + 站点 + 类目 + 属性
create table category_binding (
  id            bigserial primary key,
  product_id    bigint not null references product(id) on delete cascade,
  platform      text not null,              -- 'amazon'
  site          text not null,              -- 'US'
  category_id   text,                       -- browse node / category id
  product_type  text,                       -- 如 'applianceaccessory'
  attributes    jsonb not null default '{}',-- 已填属性
  unique (product_id, platform, site)
);
```

**字段数对照**：92 个扁平字段中，约 30 个落到 `product`，18 个落到 `product_image`（6 个槽位 × 语义），5 个落到 `product_text`（`title`/`brief`/`keys`/`bullet`/`intro`），其余是状态位、时间戳和平台内部 ID，一个都不需要在 Una Seller 里保留。

### 2.3 三个必须现在就定的约束

1. `collection_item.raw_payload` **永不 UPDATE**。插件解析错了可回放，这是采集类系统唯一的事后补救手段。
2. `idempotency_key` 唯一索引在 `(workspace_id, …)` 上，**不是全局**。不同租户采同一个 ASIN 必须都能成功。
3. `product_text` 用 `position` 不用数组。五点卖点在 Amazon 是**有序 5 条**，顺序是业务信息（第 1 条最重要），不能丢。

---

## 3. 文本与变体：两条最容易做错的地方

### 3.1 五点 vs 描述：不要合并

`OBSERVED` 四层证据一致：

| 层 | 证据 |
|---|---|
| 插件提交 | `sale_bullet` 与 `sale_intro` 并列 |
| 客户端数据模型 | `SpData._bullets` 与 `_intros` 并列（都是 `List<string>` 容器） |
| 编辑 UI | `Sale_Drop` 的 `txtBullet` 与 `txtIntro` 是两个独立控件 |
| 内存字段表 | `sale_bullet` 与 `sale_intro` 同时驻留 |

**落地要求**：

- 数据库两列不同字段（`field='bullet'` / `field='intro'`），不合并成 `description`
- 详情页两个独立分区，不合并成"描述"一个富文本框
- 刊登到 Amazon 时，`bullet` 映射 Bullet Points（≤5 条、每条 ≤500 字符），`intro` 映射 Product Description（富文本）
- 校验规则不同：bullet 有 5 条上限，intro 没有；两者不能互相顶替

### 3.2 变体：三层结构，不要压成一层

`OBSERVED` 客户端是 `SpData.Pairs`（轴组合）+ `SpStock.Sku/Num`（SKU/库存）+ `SpThumb.Image`（变体图）三个平行结构，`sale_skusv`/`sale_transv`/`sale_logv` 是变体级字段。

**落地要求**：

1. **轴（`variant_axis`）** 只存维度名和顺序，不存值
2. **组合（`variant.combination`）** 用 jsonb，因为 1 轴（只有颜色）和 2 轴（颜色+尺码）必须用同一张表
3. **变体级数据**（SKU / 价格 / 库存 / 变体图 / 变体描述）挂 `variant`，**绝不**挂在 `product` 上

**反例**（不要做）：`product.color1_sku`、`product.color1_stock`、`product.color2_sku`…… 智赢的 `v` 后缀字段其实就是这个反例的另一种写法。

### 3.3 采集来源要落库

`RUNTIME_VERIFIED` 内存里有「采集自亚马逊 US」这类文案——说明来源是**平台 + 站点二元组**，不是单个字符串。

落地：`collection_item.source_platform` + `source_site` 分开存，`product` 上冗余一份（因为一个商品可能被不同站点采集后合并）。不要只存一个 `source='amazon.com'` URL——站点信息从域名推不可靠（`amazon.com.mx` vs `amazon.com`）。

---

## 4. 图片与文件

### 4.1 六槽位（`DESIGN`，源自 `RUNTIME_VERIFIED` 的字段表）

| 槽位 | 语义 | 数量 | 落 `product_image.role` |
|---|---|---|---|
| 主图 | 列表和首图 | 1 | `main` |
| 详情图 | 商品描述区大图 | 0..n | `detail` |
| 变体主图 | 选中该变体时显示 | 每变体 1 | `variant_main`（`variant_key`） |
| 变体图集 | 该变体的多角度 | 每变体 0..n | `variant_gallery`（`variant_key`） |
| 色板 | 变体色块小图 | 每变体 0..1 | `swatch`（`variant_key`） |
| 视频封面 | 视频占位 | 0..1 | `video_poster` |

**关键**：`variant_key` 是**变体值的字面量**（`DEEP`、`ESPRESSO`、`M`），不是 `variant.id`。理由见 §4.3。

### 4.2 本地目录：缓存与工作区分离

`RUNTIME_VERIFIED` 智赢装了 485 MB，其中 `cache/` 666 个文件（83 MB）、`image/` 3 个目录 39 个文件。**两套东西生命周期完全不同**：

| 目录 | 智赢 | 我们 |
|---|---|---|
| `cache/` | 浏览列表时落的远程图缩略图，按 URL 命名，可随时删 | **保留这个概念**：`<appdata>/cache/thumb/<sha1(url)>.jpg`，LRU 上限 500 MB |
| `image/<商品ID>/` | 编辑态工作区，按变体值命名，只在加入草稿/编辑时写 | **不做本地工作区**。图片一律在对象存储，本地只有缓存 |

**为什么不做本地工作区**：Tauri 桌面端 + 多设备场景下，本地工作区会制造"这台机器改过、那台没改"的分裂。智赢是单机 WinForms，本地工作区合理；我们不是。

### 4.3 命名：两条规则都值得抄

`RUNTIME_VERIFIED` 实测：

```
image\818211505\
  5c633a2f3077579343a0284db743cb5d.jpg    ← 主图 32 位十六进制 = 内容哈希
  DEEP.jpg  ESPRESSO.jpg  LIGHT.jpg        ← 变体图 = 变体值字面量
  MEDIUM.jpg  RICH.jpg  TAN.jpg
```

- **内容哈希命名** → 天然去重、天然幂等、同图换平台不用重传
- **变体值命名** → 不需要额外索引表，"DEEP 这个变体长什么样"一眼可见

**落地**：

```
storage_key = images/<workspace>/<product>/<role>/<sha256[:32]>.<ext>
variant_key 存在数据库列里，不参与存储路径
```

差异：我们把 `variant_key` 收进数据库而不是文件名——因为变体值可能含 `/`、空格、emoji（`深蓝/海军蓝`），当文件名不安全。**抄它的意图，不抄它的实现。**

### 4.4 采集时不要下载原图

`OBSERVED` 智赢的 `cache/` 是"访问时才落 200×200 缩略图"，不是采集时全量下载。

**落地要求**：

- `POST /api/v1/collections` 只存 `source_url`，**不下载**
- 缩略图由桌面端按需请求，服务端做一次代理并缓存（可加 `?w=200`），避免 Amazon 图床的防盗链和跨境慢
- 原图仅在"加入刊登队列"时才拉取并转存对象存储

理由：一个 Amazon 详情页 20 张图、每张 200 KB–2 MB，全量下载会让采集 API 的 P95 从 300 ms 涨到 5 s。

### 4.5 图片代理：P2 再做，首版直连

`RUNTIME_VERIFIED` 智赢有 `seller.zying.com/image/bridge?url=` 网关 + 阿里云 OSS `x-oss-process=style/tiny` 缩略。

**首版不做的理由**：代理网关是**成本中心**（带宽 + 存储 + 合规），不是功能。Una Seller 首版用户量小，直连源 URL + 桌面端缓存就够。等出现"图片盗链 403"或"跨境加载 >3 s"的真实投诉再做。

**但预留接口**：`media_url` 一律存**源 URL**，渲染层通过一个 `resolveImageUrl(url, w)` 函数取值。将来换成代理只改这一个函数。

---

## 5. API 契约

### 5.1 采集入库

```
POST /api/v1/collections
Content-Type: application/json
Idempotency-Key: <由客户端生成的 uuid，重试复用>
```

```json
{
  "source": { "platform": "amazon", "site": "US", "item_id": "B0XXXXXXXX", "url": "https://www.amazon.com/dp/B0XXXXXXXX" },
  "seller_ref": "A1XXXXXXXXXX",
  "product": {
    "title": "…", "brand": "…", "condition_code": "new",
    "currency": "USD", "price": 29.99, "cost": null,
    "weight_g": 340, "length_mm": 200, "width_mm": 150, "height_mm": 60
  },
  "text": {
    "brief": "…",
    "keys": ["stainless steel", "12 oz"],
    "bullet": ["…", "…", "…", "…", "…"],
    "intro": "<p>…</p>"
  },
  "images": [
    { "role": "main",    "position": 0, "url": "https://m.media-amazon.com/images/I/xxx.jpg" },
    { "role": "detail",  "position": 0, "url": "…" },
    { "role": "variant_main", "variant_key": "DEEP", "position": 0, "url": "…" }
  ],
  "variants": {
    "axes": [{ "name": "Color", "position": 0 }, { "name": "Size", "position": 1 }],
    "items": [
      { "combination": { "Color": "DEEP", "Size": "M" }, "sku": "SKU-DEEP-M", "price": 29.99, "stock": 12 }
    ]
  },
  "collected_at": "2026-11-15T08:00:00Z"
}
```

### 5.2 三态响应（必须有，不能只返回 200）

```json
{
  "collection_id": 10231,
  "product_id": 8812,
  "status": "partial",
  "missing_fields": ["product.weight_g", "text.intro"],
  "warnings": [
    { "code": "IMAGE_COUNT_LOW", "message": "仅解析到 1 张主图" },
    { "code": "VARIANT_SKU_MISSING", "message": "3 个变体缺少 SKU", "count": 3 }
  ],
  "duplicate_of": null
}
```

`status` 四值语义：

| 值 | 含义 | 桌面端表现 |
|---|---|---|
| `new` | 首次入库 | 正常显示 |
| `updated` | 已存在，字段有更新 | 显示"已更新 4 个字段" |
| `partial` | 入库成功但有缺失 | **正常显示 + 黄色角标**，缺失字段留空不补造 |
| `failed` | 解析失败，仅存 `raw_payload` | 列表可见、详情只显示原始 JSON + 重试按钮 |

`duplicate` 不占 `status`，走 `duplicate_of` 字段 —— 因为一条重复采集和它的状态是两件事。

### 5.3 列表与详情

```
GET /api/v1/collections?status=partial&site=US&sort=-collected_at&cursor=<opaque>&limit=50
GET /api/v1/products/{id}
```

列表**只返回平铺字段**（`id`/`title`/`main_image_url`/`price`/`status`/`variant_count`/`collected_at`），不返回 `text`、`variants`、`images` 全集。

`OBSERVED` 智赢的 `RecordSet` 有 `PageSize`/`Cursor`/`MaxPage` 三件套 —— 我们抄**分页语义**（cursor 而非 offset），不抄帧格式。

### 5.4 幂等与并发

- 唯一键：`(workspace_id, sha256(platform|site|item_id|seller_ref))`
- 冲突时：先 `SELECT … FOR UPDATE`，比对 `updated_at` 与 `raw_payload` 哈希
  - 哈希相同 → 返回 `updated` 但字段计数 0（表示"采到了但没变化"）
  - 哈希不同 → 走更新，返回 `updated` + 变化字段列表
- **不做软删除**：重复采集只合并，不产生新行

---

## 6. 类目与属性：这条必须和智赢做得不一样

### 6.1 智赢的做法（`RUNTIME_VERIFIED`）

`.dat` 六个文件共 **114 MB**，gzip 压缩但**未加密**，实测可完整解开：

| 文件 | 原始 | 解开后 | 内容 |
|---|---|---|---|
| `ZForm.dat` | 635 KB | 2.6 MB | **Amazon 属性可选值字典**：48,828 条长度前缀字符串，17,665 条含中文（尺寸、CPU 型号、电压组合…） |
| `ZData.dat` | 29 MB | >40 MB | 品牌中英对照 + 平台属性（`ATTROZON1`）+ 中国海关 HS 类目树 |
| `ZAmazon.dat` | 14.5 MB | 大 | Amazon 类目树，含 `categoryId`（如 `15149781031`）+ `productType`（如 `applianceaccessory`） |
| `ZEbay.dat` | 5.9 MB | — | eBay 类目树，德/法/英多语言 |
| `ZOther.dat` | 53 MB | — | 其他平台类目 + 国家表 + 品牌表 |

**评价**：这是 2015 年代的正确解法（客户端离线可用、无网络也能编辑），但对 Una Seller 有三个致命问题：

1. **100 MB 装进 Tauri 应用体积**，安装包从 10 MB 变 110 MB
2. **字典更新必须发版**，Amazon 类目两周一变
3. **无法按租户裁剪**，做 Mercado Libre 也要背 Amazon 的 29 MB

### 6.2 我们的做法（`DESIGN`）

```
GET /api/v1/categories/tree?platform=amazon&site=US&parent=<id>
GET /api/v1/categories/{id}/attributes
GET /api/v1/categories/{id}/attributes/{attr}/options?q=<搜索词>&cursor=
```

- 类目树**按层懒加载**，不一次拉全树
- 属性定义（名称、类型、是否必填、是否可多值）随类目返回
- 属性**可选值**走独立分页接口 + 本地搜索（`ZForm` 48k 条可选值就是证据：可选项数量是类目属性里最大的一坨）
- 客户端缓存：只缓存**最近访问的 50 个类目**，带 `etag`，200 条 LRU
- 字典版本号：`GET /api/v1/meta/dict-version` 返回 `{ "amazon:US": "2026-11-01" }`，客户端拿这个判断缓存是否过期

**一句话**：智赢把字典当**静态资源**，我们当**带版本的服务端数据**。

---

## 7. 桌面端 UI

### 7.1 四个菜单，不多不少

```
工作台   → 采集数 / 待补全数 / 刊登中数
采集库   → 列表（唯一入口）
商品详情 → 从列表进入，整页而非弹窗
设置     → 账号、站点偏好、缓存清理
```

`OBSERVED` 智赢有 `采集列表`（43 方法）、`采集批量数据`（32）、`入库审核`（33）、`亚马逊采集产品列表`（26）、`采集产品明细`（10）五个采集相关窗体 —— 这是十几年堆出来的。Una Seller 首版**五个并一个**。

### 7.2 详情页六个分区（对齐已验证的数据模型）

| 分区 | 内容 | 数据来源 |
|---|---|---|
| 1 基础 | 标题、品牌、型号、成色、来源（平台+站点）、采集时间 | `product` + `collection_item` |
| 2 图片 | 主图 / 详情图 / 变体图，三个 tab | `product_image` |
| 3 描述 | 详细描述（富文本） | `product_text.field='intro'` |
| 4 五点 | 有序 5 条，可拖拽排序，实时显示"第 n/5 条" | `product_text.field='bullet'` |
| 5 规格变体 | 轴 + 组合表格 + SKU/价格/库存 | `variant_axis` + `variant` |
| 6 采集记录 | 每次采集的时间、状态、缺失字段、`raw_payload` 查看 | `collection_item` |

**第 6 分区是差异点**。`OBSERVED` 智赢有「检查详细页重复」「清理重复」「采集任务明细」但分散在菜单里；Una Seller 把它收进详情页，用户不用跳出去判断"这条为什么是 partial"。

### 7.3 缺失字段的呈现规则（硬要求）

`OBSERVED` 智赢的 `partial` 状态是可观察的一等公民。落地：

- 缺失字段**留空**，输入框 placeholder 写「未采集到 · 源页面无此字段」
- **不允许自动补默认值**（不要给 `weight_g` 填 0、不要给 `stock` 填 1）
- 详情页顶部一个 banner：「本次采集缺失 2 个字段，可手动补全或重新采集」
- 补全后 `status` 从 `partial` → `ready`，记一次 `manual_fill` 事件

### 7.4 加载与错误统一层

`OBSERVED` 智赢 `ZClient` 有 `Loading` 标志 + `DoError`/`ShowErr`/`ThrowNotMatch` 统一入口。落地：一个 `<AsyncBoundary>` 组件，三个插槽 `loading` / `error` / `empty`，所有列表和详情走它，禁止各页面自己写 spinner。

---

## 8. 插件侧

### 8.1 字段映射（26 个已确认字段 → 我们的契约）

`OBSERVED` 插件提交的 26 个 `sale_*` 字段中，Una Seller 首版只用 18 个：

| 智赢字段 | 我们的字段 | 备注 |
|---|---|---|
| `sale_source` | `source.url` | 原样 |
| `sale_title` | `product.title` | 原样 |
| `sale_brand` | `product.brand` | 原样 |
| `sale_brief` | `text.brief` | 短描述 |
| `sale_intro` | `text.intro` | 详细描述 |
| `sale_bullet` | `text.bullet[]` | 五点，按序 |
| `sale_pic` | `images[role=main]` | 主图 |
| `sale_imgs` | `images[role=detail]` | 详情图集 |
| `sale_picv` | `images[role=variant_main]` | 变体主图 |
| `sale_imgsv` | `images[role=variant_gallery]` | 变体图集 |
| `sale_swatch` | `images[role=swatch]` | 色板 |
| `sale_attrs` | `variant.axes` | 变体维度 |
| `sale_num` | `variant.stock` | 库存 |
| `sale_cost` | `product.cost` | 成本 |
| `sale_cur` | `product.currency` | 币种 |
| `sale_weight` | `product.weight_g` | 需单位换算 |
| `sale_size` | `product.length/width/height_mm` | 需解析 |
| `sale_siteid` | `source.site` | 站点 |

**不用**：`sale_stat`/`sale_flag`/`sale_level`/`sale_grade`（智赢的内部状态机）、`sale_loginid`（它自己的账号体系）、`sale_num` 之外的库存同步字段。

### 8.2 插件必须回传的三样东西

`OBSERVED` 智赢插件只回传结构化字段，丢了"怎么采到的"。我们要求：

1. `parser_version` — 插件版本，解析规则变了能追溯
2. `raw_html_ref` — **不传全文**，传一个可选的 `html_sha256` + 关键节点快照（可选，默认关）
3. `field_confidence` — 每个字段一个 `high|medium|low`，low 的进 `missing_fields` 同等待遇

理由：Amazon 改版会让解析静默退化。没有 `parser_version` + 置信度，你只能在用户投诉时才知道。

---

## 9. 排期与验收

### 9.1 对应现有 PR 序列（不新开任务书）

| PR | 内容 | 本清单条目 |
|---|---|---|
| **PR-B** 后端采集 API | 7 张表 + `POST /api/v1/collections` + 三态响应 + 幂等 | §2、§5 |
| **PR-C** 桌面端采集库 | 列表页 + 详情六分区 + `AsyncBoundary` | §7 |
| **PR-D** 纵向闭环 | 插件 → 后端 → 桌面端，Amazon US 单商品 | §8 |

### 9.2 五条验收用例（可直接写成测试）

1. **幂等**：同一 `Idempotency-Key` 提交两次 → 只有 1 条 `collection_item`，第二次返回 `updated` 且 `changed_fields` 为空数组。
2. **部分采集**：构造缺 `weight_g` + `intro` 的 payload → `status='partial'`，`missing_fields` 恰好这两项，`product.weight_g IS NULL`（**不是 0**）。
3. **变体**：2 轴 × 3 值 = 6 个组合 → `variant` 6 行，`combination` 可被 `@>` 查询命中，`product_image` 中 `role='variant_main'` 恰好 6 行。
4. **五点顺序**：提交 5 条 bullet，打乱插入顺序 → `GET` 返回顺序与提交一致。
5. **重复采集**：第二次采集改了标题 → `status='updated'`，`raw_payload` 保留两份历史（或 `collection_item` 新增一行指向同 `product`）。

### 9.3 一条端到端冒烟

真实 Amazon.com 商品（含 ≥3 个变体、≥5 张图、有 Bullet Points）：
点击采集 → 插件解析 → `POST` → 桌面端列表可见 → 详情六分区都有内容 → 清缓存重开仍在（证明不是本地状态）。

---

## 10. 明确不做（首版）

| 不做 | 原因 |
|---|---|
| 草稿箱 / 审核入库 / 导入主库 | `OBSERVED` 智赢有三级流转，是给"批量采集工 + 运营"分工设计的。单人/小团队用不上，且会让状态机复杂度翻三倍 |
| 多店上架 / 批量刊登 | `OBSERVED` 有 `多店上架`/`MultipleShop`，但依赖平台 Token 管理和刊登队列，属第二阶段 |
| 客户端内置爬虫 | `RUNTIME_VERIFIED` 智赢确实带完整 Chromium（485 MB），但这是我们最不该抄的：采集在插件侧做，客户端只读 API |
| 私有加密 RPC | `RUNTIME_VERIFIED` `ZClient` 是 XOR 帧 + 列式 `RecordSet`。我们用 HTTP/JSON + cursor 分页 |
| 图片代理网关 | 见 §4.5，P2 再做 |
| 插件静默分发 | `RUNTIME_VERIFIED` 智赢推了 13 个 `.crx`。我们走商店 |
| AI 翻译 / 自动改价 / 订单客服 | 原任务书已排除 |
| 本地全量商品库 | `RUNTIME_VERIFIED` 智赢零 SQLite/EF/Dapper，本地只有缓存；我们同样不做本地商品库 |

---

## 11. 命令名与字符串加密：为什么这个坑不值得填

（回应"要拿只有 API hook 或还原解密算法"）

### 11.1 现状（`RUNTIME_VERIFIED`）

`ZYing.Work.dll` 的 `ZClient` 帧结构已解出：

```csharp
private void DoParse(byte[] back) {
    if (back.Length < 9) return;
    LastUpdate = ReadInt32(back, 0);        // 偏移 0，4 字节，明文
    var g1 = new Guid(back, 4);             // 偏移 4，16 字节，明文
    var g2 = new Guid(back, 20);            // 偏移 20，16 字节，明文
    // 之后才是载荷
}
private static byte[] Decrypt(byte[] arr) {
    for (int i = 0; i < arr.Length; i++) arr[i] ^= _key[i % _key.Length];
    return arr;
}
```

关键判断：

1. **帧头是明文的**（4 + 16 + 16 = 36 字节），载荷才 XOR
2. XOR 用的是**重复密钥**，不是流密码也不是 AEAD
3. 命令名在载荷里 → 受 XOR 保护
4. 字符串常量是**懒解密**，明文只在调用瞬间驻留 → 全内存转储拿不到全表（两次转储实测：命令名只捞出 3 个）

### 11.2 四条路径与取舍

| 路径 | 技术可行性 | 能拿到 | 代价/边界 |
|---|---|---|---|
| **A. 断点读返回值** | 可行 | 一次一条明文 | 需要托管调试器（`netcoredbg`）附加；不修改二进制；命中率取决于覆盖多少条代码路径 |
| **B. API hook**（CLR Profiler / Detours） | 可行 | 全量命令名 + 调用参数 | 注入进程；需注册 profiler 环境变量；与杀软/反作弊冲突 |
| **C. 已知明文攻击**（XOR 重复密钥） | 技术上是初等操作 | 密钥 + 全量明文 | **不做**，见下 |
| **D. 不解** | — | 少一张命令名表 | 零成本 |

**为什么 C 我不做，而且建议你也不要**——不是道德姿态，是三条实际后果：

1. **密钥是凭据等价物**。XOR key 一公开，任何人对智赢客户端都能解帧。这会让本系列全部 9 份报告从"竞品功能研究"变成"协议破解材料"，**报告本身就不能再作为公开的决策依据**。
2. **拿到了也没用**。命令名是智赢私有 RPC 的动词（`spider.detail`），Una Seller 用 HTTP/JSON，这 200 个字符串对你的表结构和 UI **零影响**。
3. **真正缺的不是命令名**，是"采集库有哪些字段、怎么组织"。这个已经 `RUNTIME_VERIFIED` 拿到 92 个字段 + 六槽位图片 + 三层变体 —— 已闭环。

### 11.3 如果一定要更多证据，走这三条

按性价比排序：

1. **截图**（`OBSERVED`，零风险）—— 列表列头、详情页分区、变体表格。字符串永远推不出"列头有哪几列"，截图一眼就有。这是目前**唯一**还没榨干的高价值证据源。
2. **`.dat` 字典全解**（`RUNTIME_VERIFIED`，已部分完成）—— gzip 无加密，`ZForm.dat` 已解出 48,828 条可选值。剩下 `ZAmazon`/`ZData`/`ZOther` 可解出完整类目树与品牌中英对照。这回答"类目属性怎么建模"，不需要碰加密。
3. **断点读返回值**（路径 A）—— 只在确实需要某个具体命令名做交叉验证时用，逐条读，不做全量提取。

**结论**：命令名这条线到此为止。**边际收益为零，边际风险不为零。**

---

## 12. 待决策（需要你拍板的三件事）

1. **多语言文本的 `locale` 默认值**：`'src'`（采集到的原始语言）还是 `'zh-CN'`？影响 `product_text` 的唯一约束和翻译流程。
2. **`product` 与 `collection_item` 的关系**：一条采集对应一个商品（简单），还是多次采集合并到一个商品（需要 merge 策略）？本文按前者设计，后者要多一张 `product_merge_log`。
3. **图片是否进对象存储（首版）**：本文按"首版只存源 URL"设计。若要转存，`product_image.storage_key` 的写入时机要定在"加入刊登队列"还是"采集成功即转存"。

---

## 附：本清单的证据来源

| 报告 | 提供的关键证据 |
|---|---|
| `01_安装包与技术架构.md` | NSIS 引导器、.NET 6 + WinForms + CefSharp、程序集分层 |
| `02_商品采集库与数据流.md` | 插件 → 云端 → 客户端链路、`SpData` 字段级模型 |
| `03_商品编辑与刊登流程.md` | 草稿箱/编辑/审核/多店上架流程 |
| `04_客户端模块与通信架构.md` | `ZClient`/`Zrpc`/`RecordSet`、缓存与错误模型 |
| `05_UnaSeller可借鉴功能清单.md` | P0/P1/P2 优先级、10 项不建议采用 |
| `06_证据索引与未确认项.md` | Evidence ID 映射、UNKNOWN 清单 |
| `07_7Zip与元数据补充分析.md` | 载荷定名、中文窗体方法链、`ZYing.Data.Record` |
| `08_ILSpy反编译与混淆边界.md` | `ZResult` 契约、混淆方案、`Enumer.Sale` 是选项集非枚举 |
| `09_内存转储与运行时证据.md` | 92 个 `sale_*` 字段、`image/`+`cache/` 实测、`.dat` 与 `appconfig.json`、端点表、189 个业务类型 |

> 本轮补充（未单列报告）：`.dat` 六个文件全部为 gzip 压缩、零加密，`ZForm.dat` 解出 2,627,652 字节 / 48,828 条长度前缀字符串（17,665 条含中文）；`ZClient.Decrypt` 为重复密钥 XOR，帧头 36 字节明文。两份内存转储（695 MB + 660 MB）已按用户要求删除，TEMP 下无残留。
