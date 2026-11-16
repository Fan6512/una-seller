# 05 Una Seller 可借鉴功能清单

> 用途：把智赢 Windows 客户端的静态观察，翻译成 Una Seller 的**功能要求 / 数据契约要求 / 通用技术思路**。
> **只提炼要求与思路，不复制第三方专有代码、接口、签名或反编译产物；自研实现必须独立编写。**
> 本轮仅研究，**不修改 `ERP-MVP.md` / `BROWSER-COLLECTION-PLUGIN-MVP.md` / `FOUNDATION.md` 等正式开发任务书，不创建业务 PR，不自动合并。**
>
> **证据等级**：表中「智赢静态证据」列来自 `01`–`04` 的 **OBSERVED** 结论；「价值 / 独立实现要点 / 不建议采用理由」属于**基于自研约束的设计判断（INFERRED）**；无法确认的实现细节见 `06` 的 **UNKNOWN** 清单。本轮无 **RUNTIME_VERIFIED** 结论。

---

## 0. 对照基线

| 文档 | 来源 | 关键约束 |
|---|---|---|
| `docs/architecture/FOUNDATION.md`（Foundation v2） | 本地工作区 | 简单优先；Tauri 2 + React + TS；单后端；PostgreSQL；本地不建全量业务库；平台 Token 只在服务端；不提前引入复杂抽象 |
| `docs/architecture/ERP-MVP.md` v0.3 | `Fan6512/una-seller` GitHub（本轮取回） | 首版只到「采集库 + 商品详情」；`source_products` / `source_product_images` / `source_product_variants` 三表；`POST /api/v1/collections` 一次幂等入库；去重键 = 用户 + 平台 + 站点 + 规范化 URL；状态区分新建/更新/部分/失败 |
| `docs/product/BROWSER-COLLECTION-PLUGIN-MVP.md` v0.3 | `Fan6512/una-seller` GitHub（本轮取回） | 人工点击采集；Amazon.com US 优先；MLM/MLB/MLA；`description` 与 `bullet_points` 独立；P1–P5 五项实施要求；不上传整页 HTML；不补造变体 |

> **已有结论不重复**：插件侧的 P1–P5（两阶段采集、变体补全、描述/卖点独立、页面状态分类、一次幂等入库）已由前 3 份插件报告沉淀进 v0.3 任务书。本轮 EXE 研究的**增量价值**在「客户端如何组织、展示、编辑、缓存与通信」，以及「哪些商业级设计不该照搬」。

---

## 1. 可借鉴功能清单

优先级定义：**P0 = 首版验收必需**；**P1 = 首版之后紧接着做**；**P2 = 有真实使用频率后再评估**；**X = 不建议采用**。

### 1.1 P0：数据契约与展示结构（首版必需，且任务书已覆盖大部分）

| # | 可借鉴点 | 智赢静态证据 | 对 Una Seller 的价值 | 独立实现要点 |
|---|---|---|---|---|
| P0-1 | **图片按「主图 / 详情图」分集合，且保留顺序** | `SpData.Pics` / `DetailImgs` 均为 `List<string>`；`SpThumb(Title,Url,Image)`；资源「产品图片」「图片目录」「替换变体图片」 | 详情页能正确分区展示，不变体错图 | `source_product_images.image_type` + `image_order` + `variant_id`；只存 URL，不塞二进制进 JSON |
| P0-2 | **普通描述与五点卖点全程分离**（采集层 `_intros`/`_bullets` → 属性 `Intro`/`Bullet` → 编辑控件 `BulletTemplate`/`BulletTextInput`） | `ZYing.Work.dll` REA 字段/方法；`ZYing.User.dll` 资源 | 避免「把描述切成五点」这类数据污染，直接对应验收项 | `description` 文本 + `bullet_points` 有序数组；页面无卖点时留空，不自动拆分、不 AI 补造 |
| P0-3 | **变体组合、SKU、库存三者分离** | `SpData.Pairs` / `Stocks` / `Size`；`SpStock.Sku` / `Num` / `IsOffline` / `Items` | 保证「真实存在的组合」不被笛卡儿积伪造 | `source_product_variants.attributes` JSON + `price`/`currency` + 图片关联；只保存页面实际提供的组合 |
| P0-4 | **采集状态与缺失信息是一等公民** | `SpData.Error` / `CheckValid()`；`采集失败`、`待审核`、`部分`类状态词 | 支撑 `collection_status` / `collection_warnings` / `page_state` 三态分离 | 后端返回明确「新建 / 已有更新 / 部分保存 / 失败未保存」；插件仅在持久化成功后改按钮状态 |
| P0-5 | **采集库 = 列表 + 明细 + 独立操作弹窗** | `ZYing.Channel.ProductList/ProductDetail`、`ZYing.Detail.Sale`、`ZYing.Dialog.*`；资源「亚马逊采集产品列表/明细」 | 验证「4 个首版菜单」的信息架构是合理的最小集 | React 列表页 + 详情页 + 少量 Modal；不引入复杂窗口系统 |
| P0-6 | **重复判断独立成功能** | `检查详细页重复`、`清理重复`、`IsRepeat` / `RemoveRepeat`、`采集产品黑名单` | 支撑规范化 URL 去重 | 唯一键 + PostgreSQL 事务；不照搬「远端防重 + 二次入库」两段式 |
| P0-7 | **图片与商品记录解耦**（`ImageURL` / `ImageBase64` / OSS 域名 / `IO.Cache.GetImage`） | `ZYing.Core.dll` | 图片失败不拖垮商品字段 | 保留来源 URL + 可选本地缓存路径；下载失败记录状态而非丢字段 |

### 1.2 P1：首版之后紧接着做（有明确运营价值）

| # | 可借鉴点 | 证据 | 说明 |
|---|---|---|---|
| P1-1 | **从采集库到草稿的显式中间态** | `把产品加入到草稿箱` / `修改草稿箱的产品` / `QueryProductDraft` / `Drafts` / `ProductData` | 若后续要做刊登，草稿态能隔离「原始采集资料」与「可编辑待刊登资料」；首版不做 |
| P1-2 | **采集任务列表 + 重试 + 二次筛选** | `亚马逊采集任务列表/明细`、`采集二次筛选`、`批量重试发布`、`SpiderSync` / `SpiderBatch` | 对应插件 MVP 的「重新采集 / 部分成功 / 状态回显」；先做单商品闭环再考虑任务化 |
| P1-3 | **多语言文本容器** | `SpData._titles/_keys/_bullets/_intros` 同构 + `GetAllLangs()` | MLM（西语）/ MLB（葡语）/ MLA（西语）多语言场景可参考；首版不做 AI 翻译与改写 |
| P1-4 | **变体字典 / 映射 / 词替换** | `变体字典`、`变体映射列表`、`替换变体词`、`采集变体替换` | 跨平台刊登时的规格名归一化；有真实数据后再建 |
| P1-5 | **统一请求状态层** | `ZClient` 的 `UI/Loader/Loading/IsActive/Busy`、`DoError/ShowErr/Throw` | React Query 的 `isPending/isError` 即可覆盖；避免把网络状态写进组件树深处 |
| P1-6 | **批量编辑能力** | `批量修改价格`、`批量修改库存`、`批量设置店铺`、`批量更改产品属性` | 采集库条目变多后才有意义，首版单条编辑足够 |

### 1.3 P2：观察但不建议排期

| 点 | 证据 | 暂缓理由 |
|---|---|---|
| 标题模板 / 标题词 / 词替换 / 翻译标题 | `产品标题模板`、`标题词导入`、`翻译标题` | 属于 AI/运营增强，`ERP-MVP.md` 首版明确不做 AI 改写与翻译 |
| 图片差异处理 / 图片放大 / 拍照上传 | `图片差异处理`、`任务图片放大`、`上传拍照图片` | 图片加工链路成本高，首版只需预览与顺序 |
| 价格计算器 / 国家加价 / 国家运费 | `价格计算器`、`国家定价`、`国家运费` | 属于刊登定价，首版不做刊登 |
| 店铺归组 / 绩效 / 水印 / 多店铺主机 | `店铺归组`、`店铺绩效`、`店铺水印` | 属于成熟多店 SaaS 运营，超出「首期自用单用户」 |
| 订单 / 客服 / 面单 | `ZYing.Detail.Order`、`DMail`、`PrintWayBill` | `FOUNDATION.md` 列为 M3 按需添加 |

---

## 2. 通用技术思路（可借鉴，不涉及代码）

1. **采集与编辑解耦**：采集结果（`SpData`）先校验（`CheckValid`）、再转记录（`ToRecord`），编辑发生在后续环节。→ 对应「插件只提交结构化数据，后端负责校验/去重/持久化，桌面端只读展示与编辑」。
2. **「来源资料」与「刊登资料」分层**：`ZYing.Detail.Sale`（资料）与 `EMarketSale` / `UpdateMarketItem` / `PublishMartketResult`（刊登）分离。→ 未来做刊登时，不要在 `source_products` 上直接堆平台字段。
3. **图片 URL 与二进制缓存分离**：`ImageURL` / `ImageBase64` / `IO.Cache.GetImage` / OSS 域名。→ 与 `ERP-MVP.md`「商品图片保存为文件，不把二进制塞进数据库 JSON」一致。
4. **统一错误模型**：`Result` / `IsErr` / `Message` / `CheckValid().Error`。→ 后端统一错误响应 + 字段级 `missing_fields` / `warnings`，避免插件误报成功。
5. **多语言作为一等结构**：四个同构容器 + `GetAllLangs()`。→ 若未来支持多语言描述，用结构化容器而非拼接字符串。

---

## 3. 明确不建议采用的设计（X）

| # | 智赢设计 | 为什么不适合 Una Seller |
|---|---|---|
| X-1 | **客户端内置 CEF 采集器**（`ZYing.Crawl.SpiderBase` + CefSharp + 注入 JS） | 与「插件人工点击、不自动爬取」的产品定位冲突；增加合规与维护成本；`BROWSER-COLLECTION-PLUGIN-MVP.md` 明确不绕过反爬、不无人值守采集 |
| X-2 | **自定义加密二进制 RPC / 私有签名协议**（`ZClient` 的 `_tcp`/`Encrypt`/`Decrypt`、`SignRequest`、插件侧 WASM 签名） | 自研系统用标准 HTTPS + JSON 已足够；引入私有协议只会增加双端维护与调试成本，且涉及不可复用的第三方专有实现 |
| X-3 | **客户端直连平台 API / 持有平台凭据**（`ZYing.Work.dll` 内 `api.mercadolibre.com` / `MeliAuth` / `X-Api-Key`） | `FOUNDATION.md` 强制「美客多平台调用只通过官方 REST API，OAuth 凭据保留在服务端，桌面前端与 AI 不得持有 Token」 |
| X-4 | **两段式商业采集链路**（插件 `spider.check` → `amazon_hub.spider` → `spider.insert`） | `BROWSER-COLLECTION-PLUGIN-MVP.md` P5 已决定「一次 `POST /api/v1/collections` + 后端事务幂等」，不需要独立防重服务/商业采集参数接口 |
| X-5 | **扁平 `sale_*` 字段命名直接照搬** | 属于第三方内部契约；Una Seller 已有 `source_products` 等独立 DTO，照搬会造成语义耦合与合规风险 |
| X-6 | **Windows 专属 UI/互操作**（Win32 控件消息、MSHTML、`Microsoft.WindowsDesktop.App`） | 与 Tauri 2 跨平台路线冲突；Rust 只做系统能力，业务逻辑不下沉 |
| X-7 | **草稿箱 → 审核 → 导入主库 → 多店上架的多级工作流**（首版） | 属于成熟商业 ERP 的团队流程；`ERP-MVP.md` 首版明确不做刊登、不做订单、不做团队权限，提前建设违反 Foundation「简单优先」 |
| X-8 | **请求签名 / 设备指纹 / 机器码 / 更新配置标识** | 只记录「存在此类机制」的边界，不输出明文、不尝试复用、不写入自研代码 |
| X-9 | **整页 HTML 上行 + 服务端解析**（对照 `03_跨境ERP助手…md`） | 与已确定的「插件本地解析、结构化提交」相反，且放大数据与合规风险 |
| X-10 | **本地全量商品库 + 复杂离线同步** | `FOUNDATION.md`：桌面本地先不建全量业务数据库，需要缓存时再用 SQLite；不提前做双向同步 |

---

## 4. 智赢证据 → Una Seller 数据契约对照（独立设计，不复制字段）

| 智赢观察（OBSERVED） | Una Seller 现有契约 | 是否要改 |
|---|---|---|
| `SpData.Url` / `Source` / `SiteID` / `Cur` / `Cost` / `CateId` / `Lan` | `source_products.source_url` / `normalized_url` / `source_platform` / `source_site` / `price` / `currency` / `category` | 否，已覆盖 |
| `SpData.Pics` / `DetailImgs` / `SpThumb` | `source_product_images`（`image_url` / `image_order` / `image_type` / `variant_id` / `local_path`） | 否，已覆盖；只需确保 `image_type` 区分主图/详情图/变体图 |
| `SpData.Pairs` / `Stocks` / `Size` / `SpStock` | `source_product_variants`（`attributes` / `price` / `currency` / 图片关联） | 否，已覆盖 |
| `_intros` / `_bullets` | `description` / `bullet_points`（数组） | 否，已覆盖 |
| `SpData.Error` / `CheckValid()` | `collection_status` / `collection_warnings` / `page_state` / `missing_fields` | 否，已覆盖 |
| `SaleId` / `sale_stat` / `sale_flag` 等状态编号 | 无对应 | **不新增**；Una Seller 用语义化状态，不引入第三方编号体系 |
| `Seller` / `Shop` / `Buys` / `VideoUrl` / `WAttr` | 明确「不采集卖家、销量、评价、品牌」 | **不新增**，遵守插件 MVP §4 |

---

## 5. 三个最值得先开发的功能（与最终汇总第 8 问一致）

1. **后端采集 API + PostgreSQL 三表 + 幂等去重**（对应 `ERP-MVP.md` PR-B）
   `POST /api/v1/collections` 一次事务完成去重/条件更新，`GET` 列表与详情可回读；必须能保存一个含图片、普通描述、Bullet Points、真实变体的测试商品，且失败响应不能被当成成功。
2. **桌面采集库列表 + 详情页**（对应 PR-C）
   列表支持缩略图/标题/平台/站点/价格/时间/状态/关键词搜索/筛选/分页；详情分区展示基础信息、图片、描述、五点、规格、采集记录；缺失字段留空并给出原因。
3. **插件 → 后端 → 桌面端的单商品纵向闭环**（对应 PR-D 与插件 MVP §12）
   Amazon.com US 一个真实商品：点击采集 → 结构化入库 → 桌面端立刻可见；同时验证重复采集、部分采集、失败重试三种状态。

> **不排进前三**：草稿箱/审核、多店上架、AI 翻译、批量改价、订单客服、客户端内置爬虫。理由见 §1.3 与 §3。

---

## 6. 落地建议（不修改任务书，只作为下一轮讨论输入）

1. 在 `source_product_images` 的取值范围上明确 `image_type ∈ {main, gallery, variant, detail}`，避免后续把详情图与变体图混在一起。
2. 在 API 错误响应中固定三类字段：`status`（new/updated/partial/failed）、`missing_fields[]`、`warnings[]`，与 `SpData.Error` / `CheckValid()` 的思路一致但独立命名。
3. 把「请求 loading / error / 空态」收敛到统一 API client + Query 层，不要散落在页面组件里（对应 `ZClient` 的 UI 绑定思路，但解耦）。
4. 图片本地缓存在首版保持可选：先只存来源 URL；出现真实离线/加载性能问题后再按 `FOUNDATION.md` 引入 SQLite 与文件缓存。
5. 多语言容器思路记为 ADR 备选，不在 M0/M1 实施；MLM/MLB/MLA 首版只保留来源原文。

---

## 7. 合规声明

- 本清单**只提炼功能要求、可观察数据契约与通用技术思路**；未包含任何第三方反编译代码、私有接口实现、密钥/Token/签名算法明文。
- 未修改 Una Seller 业务代码、正式开发任务书，未创建 PR，未自动合并。
- 原始 `ZInstall.exe` 只读；未上传任何第三方安装包、反编译产物或密钥到 GitHub（本轮仅从公开仓库**读取**两份自研任务书用于对照）。
