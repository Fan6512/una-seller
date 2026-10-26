# 数据模型：智赢客户端看到的数据形状

## 最关键的一条结论

**客户端没有本地关系型数据库。** 整个反编译集合里没有出现 `create table` / `select … from` / ORM 映射，
列表数据全部是服务器返回的**记录集（RecordSet）**，列名由服务器决定。
客户端读到的不同列名共 377 个，其中按实体前缀成族的有 15 族。

这意味着：智赢的"数据模型"其实是**服务器视图的投影**。前缀就是视图归属——同一条商品记录在
跨境已上传明细里叫 `stat_itemid`，在全球商品里叫 `up_itemid`，在跟卖监控里叫 `mcm_itemid`，
客户端把它们统一读进 `_itemid` 这样的本地变量。**自研不要复刻这种设计**：它把 JOIN 逻辑压进了
每条 SQL 视图，字段语义随页面漂移，跨功能复用只能靠服务器再写一个视图。

## 记录列前缀族（服务器视图归属）

| 前缀 | 归属实体 | 列数 | 代表列 |
|---|---|---|---|
| `acc_` | 授权账号 | 16 | `acc_abbr`, `acc_access`, `acc_app`, `acc_color`, `acc_flag`, `acc_fresh` … |
| `stat_` | 跨境已上传明细 | 16 | `stat_abbr`, `stat_aid`, `stat_attr`, `stat_cateid`, `stat_cbtid`, `stat_fromid` … |
| `claim_` | 售后索赔 | 16 | `claim_affected`, `claim_aid`, `claim_has_incentive`, `claim_message`, `claim_message_trans`, `claim_players` … |
| `up_` | 全球商品（User Product） | 12 | `up_aid`, `up_attrs`, `up_cateid`, `up_cbtid`, `up_fromid`, `up_itemid` … |
| `mcm_` | 跟卖监控（Catalog Monitor） | 12 | `mcm_aid`, `mcm_catalog_items`, `mcm_catalog_status`, `mcm_categoryid`, `mcm_itemid`, `mcm_rank` … |
| `qa_` | 售前问答（本土） | 10 | `qa_aid`, `qa_id`, `qa_itemid`, `qa_link`, `qa_saleid`, `qa_site` … |
| `q_` | 售前问答（跨境） | 10 | `q_aid`, `q_antext`, `q_id`, `q_itemid`, `q_link`, `q_retext` … |
| `meli_` | 美客多杂项（重量/计数） | 9 | `meli_claim`, `meli_item_weight`, `meli_msg_count`, `meli_order`, `meli_pack_weight`, `meli_size` … |
| `order_` | 订单 | 7 | `order_accid`, `order_key`, `order_max`, `order_min`, `order_no`, `order_place_invoice` … |
| `msg_` | 站内信 | 7 | `msg_aid`, `msg_content`, `msg_content_trans`, `msg_flag`, `msg_site`, `msg_status` … |
| `attr_` | 类目属性 | 7 | `attr_allow_variations`, `attr_id`, `attr_multi`, `attr_required`, `attr_type`, `attr_value_type` … |
| `metrics_` | 店铺绩效 | 7 | `metrics_canceled_value`, `metrics_claims_value`, `metrics_delayed_value`, `metrics_period`, `metrics_trade_canceled`, `metrics_trade_completed` … |
| `log_` | 操作日志 | 5 | `log_aid`, `log_flag`, `log_group`, `log_stat`, `log_type` |
| `as_` | 自动服务状态 | 4 | `as_login`, `as_text`, `as_text_zh`, `as_time` |
| `catalog_` | 目录商品 | 3 | `catalog_id`, `catalog_listing`, `catalog_status` |

## 贯穿各族的连接键

| 键 | 含义（据使用位置判读） | 出现频次 |
|---|---|---|
| `aid` | 智赢侧的授权账号主键（不是美客多 user_id，后者另有字段） | 79 |
| `saleid` | 上传/刊登记录行主键，跨境与本土共用同一命名 | 69 |
| `itemid` | 美客多商品 ID（MLM/MLA… 站点对象） | 40 |
| `cbtid` | 跨境商品 ID，与 `itemid` 常成对出现 | 9 |
| `localid` | 本土站点的行主键 | 6 |
| `upid` | 全球商品（User Product）ID | 9 |
| `cateid` | 类目 ID | 13 |
| `claimid` | 索赔单 ID | 3 |
| `oid` | 订单 ID | 1 |
| `site` | 站点码（MLM/MLA/…） | 9 |
| `abbr` | 站点缩写，用于字典与筛选 | 22 |
| `acc` | 账号显示名（列表展示用，非主键） | 27 |
| `loginid` | 登录会话/设备标识，同步链路相关 | 4 |
| `group` | 分组键，用于账号或记录分组 | 12 |

规律：**账号维度用 `aid`，行维度用 `saleid`/`localid`，平台对象维度用 `itemid`/`cbtid`/`upid`**，
站点维度用 `abbr`（站点缩写）+ `site`。三者构成客户端所有列表页的过滤条件，也正好是自研系统里
`account_id` / `listing_id` / `platform_item_id` / `site` 这套主键的外向映射。

## 记录集信封约定

命令返回不是裸数组，而是带元信息的信封。高频返回键：`list`(60), `root`(33), `stat`(15), `stats`(10), `abbr`(9), `flags`(7), `areas`(7), `types`(3)。
其中 `list` 是数据体，`root` 是树形结构的父节点，`stat`/`stats` 是汇总口径，`abbr`/`areas`/`flags` 是
字典表。分页信息随记录集一起返回（记录对象上可直接取总条数），客户端不自己算总数。

## 直连官方 API 的对象模型

与上面不同，`MeliApi` 直连层使用强类型 DTO 反序列化，共 31 个：

`CatalogEligibilityResponse`, `CatalogProduct`, `ClaimExpectedResolutions`, `ClaimInfo`, `ClaimReason`, `CreateUpResult`, `FamilyUserProducts`, `InfractionPageResult`, `ItemPerformance`, `MarketItemMapInfo`, `MarketSiteInfo`, `MeliAllCategory`, `MeliCategory`, `MeliClipsInfo`, `MeliClipsResponse`, `MeliItemInfo`, `MeliMessage`, `MeliPicture`, `PackMessages`, `PartialRefundAcailableOffers`, `PriceToWinResult`, `ProductItemsResult`, `SearchClaimsResult`, `SearchProductResponse`, `SearchSiteItemsResult`, `SerachCatalogsResponse`, `ShopInfo`, `TokenInfo`, `UnReadMessageResult`, `UserInfo`, `UserProduct`

这些 DTO 与官方资源同名同形（商品、类目、索赔、消息包、绩效、用户产品），是**自研服务端真正该建模的部分**；前缀族属于智赢私有视图，不必继承。

## 建议的落库形态（对照结论）

| 智赢可见形态 | 自研对应 | 理由 |
|---|---|---|
| 服务器视图前缀族 | 规范化表 + 视图只在 API 层拼装 | 字段语义不该随页面变化 |
| 记录集分页元信息 | 表内 `updated_at` + 游标 | 换设备后要能续传，而不是重拉 |
| `aid`/`saleid`/`itemid` 三键 | `account_id`/`listing_id`/`platform_item_id` | 一一对应，且是同步与幂等的锚点 |
| 无本地库 | 客户端本地缓存（快照 + 变更游标） | 智赢每次都要回源，弱网与冷启动体验差；自研把服务器当唯一真相、客户端只作缓存与出站队列 |

## 限制

- 覆盖范围是已还原的 82 个类型，集中在美客多相关窗体；受 .NET Reactor 保护的模块尚未还原，非美客多板块的列族没有统计进来。
- 列名到实体的归属依据是**前缀 + 使用位置**，属推断，不是服务器 schema 的直接观测。
- 频次是"该列名在已还原代码里被读取的次数"，按文件内容去重后统计（反编译产物里每个窗体存在两份同名不同文件名的副本），不代表数据量或重要性。
- 全部为静态观察，未验证任何返回结构在真实账号下的实际形态。
