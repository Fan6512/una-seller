> **历史规划提示（2026-10-26 归档整理）**：原文 v3 内容为研究阶段构想，保留证据与演进脉络；**现行工程地基以 [FOUNDATION](../architecture/FOUNDATION.md)、[DOMAIN_AND_EXECUTION](../architecture/DOMAIN_AND_EXECUTION.md)、[CROSS_PLATFORM](../architecture/CROSS_PLATFORM.md) 及 [ADR](../decisions/README.md) 为准**。特别是原文的 NestJS/Fastify 后端、默认 Redis/Event Bus、SQLCipher 开箱可用、Windows EV 强制要求、完整同步引擎一次到位等设计，均已被新基线覆盖或改成按需决策。

# 美客多 AI ERP：合规接入清单与模块路线图

> 版本：2026-10-26（v3）  
> 本版变更：移除“官方 MCP 作为 ERP 入口”主线；新增“参考智赢的官方 API 对接实现”“批量关闭活动插件专项”与“客户端版跨设备同步审查”；明确桌面客户端优先，路线图、验收与来源随之调整。  
> 适用对象：以 Mercado Libre 卖家为目标客户、以 AI 为核心差异化的多租户 ERP/SaaS  
> 配套材料：[ZYing 逆向分析报告](../../research/zying/ZYing-美客多逆向分析报告.md)（仅作业务流与接口地图参考，不作为合规范本）  
> 说明：本文是工程与合规落地清单，不构成法律意见；涉及数据训练、跨境传输、消费者保护等事项，应由目标国法务复核。

---

## 0. 一页结论

1. **做 ERP 的“准入”是普通开发者应用 + 卖家 OAuth，不是 DPP。** DPP 是认证、曝光与支持体系，门槛是 GMV、安全评估和季度交付指标；官方文档将两者分开描述，按文档结构推断 DPP 不是调用 API 的前置条件，具体以应用后台要求为准。
2. **智赢的真实形态是“双通道”**：大部分业务走自有 ZClient 命令网关到智赢服务器，再由服务器对接美客多；少量能力（刷新 token、用户信息、CBT User Products、图片、索赔、消息、发票）由桌面端直接调用 `api.mercadolibre.com`，而授权码换 token 走服务端 `meli_acc.auth_token` 命令。新建 ERP 不应复制这套私有协议，只借鉴它的领域模型与同步策略，底层统一收敛到“官方 REST API + 服务端 OAuth”。
3. **智赢最值得借鉴的是工程模型**：以 `aid`（账号）+ `site` + `saleid`（刊登记录）为业务主键；订单按 `startTime + status` 增量拉取；刊登按字段拆分命令；图片上传对 429/500/502 做重试；token 到期前 15 分钟刷新；售后按“消息 / 索赔 / 问答”三个工作台组织。
4. **必须避开的六类实现**：把 access/refresh token 下发到客户端、`client_secret` 放查询串、把 `access_token` 拼进附件 URL、关闭 TLS 证书校验、内置多组 AppSecret、用明文私有 TCP 协议承载业务与 token。这些在智赢里存在，不代表可以照搬。
5. **“批量关闭活动”不是桌面端功能，而是 Chrome 扩展 5.1.25 的“批量关闭折扣”**（`MELI_SELLER_BATCH_STOP_DISCOUNT`）。它在美客多卖家中心 `listings/list/promos` 页面上做 DOM 自动化，不调用官方写接口，也不依赖智赢后端完成核心循环。技术上可拆，但商业上不建议；合规做法是用官方 Promotions API 重建。
6. **AI 功能的正式入口应自建**：官方 REST API + 每租户 OAuth 做工具层，再以自建 Agent Gateway 暴露给模型；模型不接触原始 token，所有写操作经过策略引擎与人工审批。
7. **推荐顺序**：P0 合规底座 → P1 只读 ERP（含促销只读）→ P2 交易闭环（含官方 API 促销退出）→ P3 AI 增强 → P4 多国与 DPP。
8. **产品形态与跨设备**：主入口是桌面客户端（Tauri 2 + React/TypeScript），业务数据以服务端为主库；客户端只保留加密只读缓存、草稿与 outbox。换机登录 ERP 账号即可恢复业务视图；同步用快照 + 游标 + 变更日志 + 幂等命令，不照搬智赢的手工时间窗。详见 4.11。

### 智赢实现与本 ERP 的取舍

| 能力 | 智赢的做法 | 证据 | 本 ERP 应采用的做法 |
|---|---|---|---|
| 账号授权 | 外部浏览器授权，授权码回填，服务端换 token，客户端保存账号记录 | [MeliApi.resolved.cs:57](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L57)、[Meli授权详情.resolved.cs:779](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L779) | 服务端 OAuth + Token Vault，客户端不接触 Meli token |
| 商品刊登 | 服务端 `meli_stat.*`/`meli.batch` + CBT 直连 `/global/user-products` | [Meli_V2.resolved.cs:655](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L655)、[MeliApi.resolved.cs:956](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L956) | 官方 `/items` 与 `/user-products`，状态机 + 幂等 + 审批 |
| 订单获取 | 服务端 `meli.order` 增量拉取，`meli.label` 返回面单 PDF | [ZOrderMeli.resolved.cs:157](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZOrderMeli.resolved.cs#L157)、[Logic.MeliLabelPrint.resolved.cs:59](file:///E:/美客多逆向/analysis/decrypted/user/Logic.MeliLabelPrint.resolved.cs#L59) | 官方 Orders/Packs/Shipments + 通知/定时对账 |
| 售后消息 | 服务端 `meli_message.*`/`meli_claim.*`/`meli_qa.*`，桌面端也直连消息/索赔接口 | [Meli消息列表.resolved.cs:573](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L573)、[MeliApi.resolved.cs:227](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L227) | 统一会话模型 + AI 草稿 + 人工发送/审批 |
| 促销退出 | Chrome 扩展 DOM 自动化点击 “Stop participating” | [promo_700k_780k.pretty.js:1070](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1070) | 官方 Promotions API：预览 + 审批 + DELETE + 复核 |

---

## 1. 参考智赢的官方 API 对接实现

> 证据边界：本节基于反编译/反混淆代码的静态分析，未运行智赢程序，也未使用任何卖家账号。凡涉及智赢服务端内部实现的部分（如服务端如何交换 token、如何聚合订单），均为从客户端命令、参数与返回字段推断，未在服务端内部验证。

### 1.1 双通道总览

```text
智赢桌面端（WinForms + CefSharp）
   ├── 通道 A：ZClient 私有 TCP 网关（默认 7735；美客多二线 17735）
   │      └── 智赢服务器 ──► 美客多官方 API / 智赢业务库
   │      业务命令：meli_acc.*、meli_stat.*、meli.order、meli_message.*、meli_claim.*、meli_qa.*
   │
   └── 通道 B：直接 HTTPS 调用 api.mercadolibre.com
          授权换 token、刷新 token、/users/me、CBT /global/user-products、
          /pictures/items/upload、/marketplace/messages/*、/marketplace/v2/claims/*、发票
```

| 通道 | 传输/鉴权 | 典型用途 | 对新建 ERP 的判断 |
|---|---|---|---|
| ZClient 私有网关 | 自定义 TCP 帧；`UUID + Member.Token`；无 TLS、无签名 | 账号、刊登、订单、消息、索赔、问答等大部分业务 | **不复制**。私有协议、token 明文、服务端黑盒，无法审计与扩展 |
| 官方 REST 直连 | HTTPS + `Authorization: Bearer`；`X-API-Version: 2` | CBT User Products、图片、用户信息、消息、索赔、发票 | **采用**。但必须改为服务端持有 token、统一网关调用 |

智赢的直连客户端把 `https://api.mercadolibre.com` 作为基址，请求前调用 `CheckToken()`，把 token 放进 `Authorization` 头；CBT User Products 额外带 `X-API-Version: 2`。见 [MeliApi.resolved.cs:1057](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1057)、[MeliApi.resolved.cs:1072](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1072)、[MeliApi.resolved.cs:1088](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1088)。

### 1.2 用户数据与授权

#### 1.2.1 授权链路

1. **拼授权 URL**：按站点选择授权主机，拼 `/authorization?response_type=code&client_id=...&redirect_uri=...&state=...`。`state` 由调用方传入，例如 `client_` + `Environment.TickCount`。见 [MeliApi.resolved.cs:57](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L57)、[Meli授权详情.resolved.cs:729](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L729)。
2. **外部浏览器授权**：桌面端打开系统浏览器或复制授权链接，卖家登录并同意后拿到 `code`，再回填到客户端。见 [Meli授权详情.resolved.cs:755](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L755)。
3. **服务端换 token**：客户端把 `{code, appid}` 发给智赢服务端的 `meli_acc.auth_token` 命令，由服务端完成 `authorization_code` 交换，返回 `token`、`userinfo`、`marketsite`。见 [Meli授权详情.resolved.cs:779](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L779)。
4. **绑定校验**：客户端用返回的 `userInfo.id` 与当前编辑的店铺做一致性校验，避免把 A 店授权绑到 B 店；随后填充到期时间、access/refresh token、店铺 ID、站点、是否 CBT、标签、公司、邮箱等字段。见 [Meli授权详情.resolved.cs:779](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L779)。
5. **持久化账号**：通过 `meli_acc.insert` / `meli_acc.update` 保存 `acc_app`、`acc_shopid`、`acc_shop`、`acc_name`、`acc_expire`、`acc_access`、`acc_fresh`、`acc_email`、`acc_corp`、`acc_siteid`、`acc_cur`、`selectedabbrs`、`groupabbrs`、`siteinfo`、`acc_iscbt`、`acc_isfull`、`acc_siteuserid`、`acc_flag`、`acc_quota`、`acc_totalitems`、`acc_listingtype`、`acc_tags`。见 [Meli授权详情.resolved.cs:1399](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L1399)。
6. **直连客户端取 token**：`MeliApiBuilder.Create(aid)` 调 `meli_acc.detail`，`GetFirst()` 调 `meli_acc.first`，把返回的记录构造成 `MeliApi`；构造函数从 `acc_*` 字段还原账号上下文。见 [MeliApiBuilder.resolved.cs:14](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs#L14)、[MeliApi.resolved.cs:1038](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1038)。
7. **到期前刷新**：每次请求前判断 `DateTime.Now.AddMinutes(15) < _expire`，不满足则用 refresh token 刷新，并覆盖保存新的 access/refresh/expire。见 [MeliApi.resolved.cs:1123](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1123)。
8. **按站点授权主机**：CBT 用 `global-selling.mercadolibre.com`；MLM/MLB/MLC/MLA/MCO 各自使用对应国家的 auth 域名。见 [MeliApi.resolved.cs:1147](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1147)。

#### 1.2.2 从官方接口拿到的用户/店铺数据

| 接口 | 用途 | 智赢落点 | 本 ERP 的建模 |
|---|---|---|---|
| `GET /users/me` | 当前授权用户身份 | 授权时校验 `userInfo.id`，刷新时同步 | `seller_account`：`meli_user_id` 为不可变主键 |
| `GET /users/{shopId}` | 店铺/用户资料 | `GetShopInfo` | `seller_profile`：昵称、邮箱、站点、国家、标签快照 |
| `GET /marketplace/users/{userId}` | 市场站点信息 | `GetMarketSite` | `seller_market`：站点、币种、CBT 标记 |
| `GET /marketplace/users/cap` | 刊登配额 | `GetListingsLimit` | `listing_quota`：配额、已用、刷新时间 |
| `meli_acc.auth_token` 返回 | token + userinfo + marketsite | 写 `acc_*` 字段 | `oauth_grant` + `seller_snapshot`，token 单独加密 |

代码位置：[MeliApi.resolved.cs:767](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L767)、[MeliApi.resolved.cs:773](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L773)、[MeliApi.resolved.cs:782](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L782)、[MeliApi.resolved.cs:791](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L791)。

#### 1.2.3 账号维度字段（智赢的 `acc_*` 模型）

| 字段 | 含义 | 是否可复用 |
|---|---|---|
| `acc_id` / `acc_group` | 账号 ID / 分组 | 可复用为 `tenant_store_id` / `store_group_id` |
| `acc_userid` / `acc_shopid` / `acc_siteuserid` | 智赢用户 / 店铺 / 站点用户 | 映射为 `owner_user_id`、`meli_seller_id` |
| `acc_siteid` / `acc_abbr` / `acc_cur` | 站点、缩写、币种 | 可复用为 `site_id`、`currency` |
| `acc_iscbt` / `acc_isfull` | 是否 CBT / 是否全托管 | 可复用为授权类型 |
| `acc_app` | 使用的应用 AppId | **不复用**；改为本 ERP 自有应用的内部引用，不落原始 AppId |
| `acc_access` / `acc_fresh` / `acc_expire` | access / refresh / 到期 | 可复用字段语义；值必须进 Token Vault |
| `acc_email` / `acc_corp` / `acc_tags` | 邮箱、公司、标签 | 最小化保存；按 DPA 与目的限制 |
| `acc_quota` / `acc_totalitems` / `acc_listingtype` | 配额、商品数、刊登类型 | 可复用为运营指标 |

#### 1.2.4 可复用与不可照搬

**可复用**

- 账号模型覆盖“人 / 店 / 站点 / CBT / 币种 / 配额 / 标签”，比只存一个 seller_id 更贴近真实运营。
- 授权时校验返回用户与当前店铺一致，防止错绑。
- token 到期前 15 分钟刷新；刷新后覆盖 refresh token。
- 按站点选择授权域名，支持 CBT 与本土站点。

**不可照搬**

- **token 下发到客户端**：`meli_acc.detail` 把 access/refresh 返回给桌面端，客户端内存与本地记录都能看到 token。见 [MeliApiBuilder.resolved.cs:30](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs#L30)。
- **`client_secret` 放查询串**：`GetToken` / `RefreshToken` 把 `client_secret` 拼进 URL。见 [MeliApi.resolved.cs:76](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L76)、[MeliApi.resolved.cs:99](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L99)。
- **内置多组 App 配置**：代码里存在 6 组应用配置（含 AppId/AppSecret/RedirectUrl），本文不展开具体值。见 [MeliAppInfo.resolved.cs:46](file:///E:/美客多逆向/analysis/decrypted/MeliAppInfo.resolved.cs#L46)。
- **刷新无并发锁**：客户端各实例可能同时刷新，单次有效的 refresh token 会互相踢掉。智赢用 `AsyncLocker` 只保护了 builder 缓存，不保护服务端刷新。

**本 ERP 的做法**

- OAuth 只在校验 `state` 后于服务端完成；`client_secret` 只从 KMS 读取，只放请求体或 HTTP Basic，不放 URL。
- token 以 AES-256 加密落库，密钥在 KMS；刷新用 `SELECT ... FOR UPDATE` 或 Redis 分布式锁串行化；刷新失败按 `invalid_grant` 触发重新授权。
- 客户端只拿本 ERP 的会话 token；任何 Meli token 不进入浏览器、桌面进程、日志、APM、LLM prompt。
- 账号资料按“授权快照 + 最小字段”保存；用户 ID、站点、CBT 标记不可变，昵称/邮箱/配额可刷新。

### 1.3 商品刊登

#### 1.3.1 智赢的刊登链路

```text
本地产品/采集数据
   → 类目预测（服务端 meli_stat.predict / 直连 domain_discovery）
   → 图片上传（meli_picture.save_pics 或直连 /pictures/items/upload）
   → 上传/创建（meli.batch、meli_log_up.*、meli_stat.publish）
   → 刊登记录（saleid）
   → 字段级修改（meli_stat.adjust_price / title / intro / attr / change_cate）
   → 上/下架（meli_stat.active / setstat / delete）
```

| 阶段 | 智赢实现 | 证据 |
|---|---|---|
| 类目预测 | 直连 `GET /marketplace/domain_discovery/search?q=`；服务端 `meli_stat.predict` | [MeliApi.resolved.cs:222](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L222)、[Meli_List.resolved.cs:1793](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli_List.resolved.cs#L1793) |
| 图片上传 | 直连 `POST /pictures/items/upload`，multipart；429/500/502 最多重试 3 次，间隔 `600ms × retry` | [MeliApi.resolved.cs:433](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L433) |
| 图片入库 | 服务端 `meli_picture.save_pics`；也直接调用 `MeliApi.UploadPic` | [Meli图片上传.resolved.cs:915](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli图片上传.resolved.cs#L915)、[Meli图片上传.resolved.cs:979](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli图片上传.resolved.cs#L979) |
| 批量上传 | 服务端 `meli.batch` | [Meli更新.resolved.cs:431](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli更新.resolved.cs#L431) |
| 上传日志 | 服务端 `meli_log_up.*` | [MeliUP_.resolved.cs:337](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.MeliUP_.resolved.cs#L337) |
| 发布 | 服务端 `meli_stat.publish`，参数 `aid` + `saleid` | [Meli_V2.resolved.cs:815](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L815) |
| 改价 | 服务端 `meli_stat.adjust_price`，支持 `price` 或 `net_proceeds`，带 `aid`、`saleid`、`isfill` | [Meli_V2.resolved.cs:655](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L655)、[Meli_V2.resolved.cs:707](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L707) |
| 改标题/描述/属性/类目 | 服务端 `meli_stat.title` / `intro` / `attr` / `change_cate` | [Meli_V2.resolved.cs:987](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L987)、[Meli_V2.resolved.cs:1041](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L1041)、[Meli_V2.resolved.cs:1095](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L1095)、[Meli_V2.resolved.cs:1199](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L1199) |
| 详情/违规/状态 | `meli_stat.detail`、`meli_stat.get_infraction`、`meli_stat.setstat` | [Meli_V2.resolved.cs:441](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L441)、[Meli_V2.resolved.cs:611](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs#L611)、[Channel.Meli_V2.resolved.cs:2051](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli_V2.resolved.cs#L2051) |

#### 1.3.2 CBT User Products 直连

智赢对 CBT（跨境）使用官方 User Products 接口：

| 操作 | 方法与路径 | 关键参数 |
|---|---|---|
| 创建全球商品 | `POST /global/user-products`，`X-API-Version: 2` | 商品模型 |
| 上架到站点 | `POST /global/user-products/{upId}` | `sites_to_sell` |
| 改价 | `PUT /global/user-products/{upId}` | 价格模型 |
| 改属性/重量尺寸 | `PUT /global/user-products/{upId}` | 属性模型 |
| 改图片 | `PUT /global/user-products/{upId}` | 图片模型 |

代码位置：[MeliApi.resolved.cs:956](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L956)、[MeliApi.resolved.cs:990](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L990)、[MeliApi.resolved.cs:1003](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1003)、[MeliApi.resolved.cs:1009](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1009)、[MeliApi.resolved.cs:1023](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1023)。

对照官方文档：普通刊登以 `POST/PUT /items` 为主；CBT/跨境以 `/global/user-products` 与 `/user-products` 组织。两者不要混用同一套字段模型。

#### 1.3.3 判断

**可复用**

- “账号 + 站点 + 刊登记录”三级主键；所有操作都带 `aid` 与 `saleid`，便于多店隔离与审计。
- 字段级操作：改价、标题、描述、属性、类目分开，适合 AI 逐项建议与人工确认。
- 图片上传的 429/500/502 重试与退避；上传与刊登分离，失败可重试。
- 上传日志与详情/违规查询分开，适合做刊登质量与失败归因。

**不可照搬**

- 服务端命令是黑盒：没有请求/响应快照、没有版本、没有幂等键，失败后难以重放。
- 客户端直接持有 token 调 CBT 接口，token 暴露面大。
- 没有 dry-run/diff：改价、改属性直接提交，AI 接入后风险更高。
- 图片重试是固定 `600ms × retry`，无抖动；高并发时容易一起打满。

**本 ERP 的做法**

- `Listing Service` 状态机：`draft → validating → publishing → active → paused → closed → failed`；每次迁移写事件与快照。
- 写操作统一 `Idempotency-Key`（租户 + 店铺 + 商品 + 操作 + 请求哈希）；重复请求返回首次结果。
- 刊登前调用官方 Validator/Moderation（如该站点/品类支持），把错误码映射成可修复项。
- AI 只产出“字段级 diff + 理由 + 影响范围”，人工确认后走同一套写入服务；批量操作支持逐条暂停/重试。
- 保存官方原始响应与请求摘要，支持回滚到上一版标题/描述/价格/属性。

### 1.4 订单获取与履约

#### 1.4.1 智赢的订单链路

| 能力 | 智赢实现 | 证据 |
|---|---|---|
| 增量拉单 | 服务端 `meli.order`，参数 `aid`、`startTime`（`yyyy-MM-dd HH:mm:ss`）、`status`；通过 `Receiving` 流式回传，`Done` 返回数量 | [ZOrderMeli.resolved.cs:157](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZOrderMeli.resolved.cs#L157) |
| 商品同步 | 服务端 `meli.sync`，参数 `aid`、`startTime`、`endTime` | [ZSyncMeli.resolved.cs:163](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZSyncMeli.resolved.cs#L163) |
| 面单 | 服务端 `meli.label`；客户端解析 base64 PDF 并打印 | [Logic.MeliLabelPrint.resolved.cs:59](file:///E:/美客多逆向/analysis/decrypted/user/Logic.MeliLabelPrint.resolved.cs#L59) |
| 批量面单 | 批量 `meli.label`，合并 PDF 后打印 | [打印Meli.resolved.cs:501](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.打印Meli.resolved.cs#L501)、[打印Meli.resolved.cs:569](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.打印Meli.resolved.cs#L569) |
| 发票 | 直连 `GET /marketplace/orders/{orderId}/invoice`，返回字节流 | [MeliApi.resolved.cs:756](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L756) |
| 拆单 | 服务端 `meli.order_split_skus`、`meli.order_split` | [Meli订单拆分.resolved.cs:197](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.Meli订单拆分.resolved.cs#L197)、[Meli订单拆分.resolved.cs:315](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.Meli订单拆分.resolved.cs#L315) |

#### 1.4.2 官方侧对应

- 订单与包裹：官方 Orders / Packs 接口；Packs 文档以 `/orders/$ORDER_ID` 组织包裹（见 Sources 的 Packs Management）。
- 面单：官方 Shipments 的面单接口；智赢把“取面单”封装成 `meli.label`，返回 base64 PDF，属于服务端代理模式。
- 发票：智赢直连 `/marketplace/orders/{id}/invoice`，说明该能力在官方 Marketplace 侧有入口；具体权限与站点支持需按文档复核。
- 拆单：属于 ERP 内部业务动作，最终仍需映射到官方订单/包裹结构；不要把智赢的拆单命令当成官方 API。

#### 1.4.3 判断

**可复用**

- 增量拉取以 `startTime + status` 为游标，适合做定时对账与断点续传。
- 流式回传（`Receiving`）适合大批量订单，避免一次性内存峰值。
- 面单按需获取并缓存 PDF；批量合并打印是高频刚需。
- 发票、面单、订单详情分开建模，避免“订单对象”过大。

**不可照搬**

- 只靠拉取、没有通知/对账：智赢的实时性依赖轮询频率，容易漏单或重复。
- 服务端聚合后只回传 ERP 字段，原始订单/包裹字段可能被裁剪，不利于纠纷取证。
- 没有幂等 upsert 的明确证据；重复拉取可能导致重复记录。
- 面单/发票经过服务端代理，若服务端日志记录 URL 或 token，会扩大泄露面。

**本 ERP 的做法**

- 订单主键 `order_id`，包裹主键 `pack_id`，物流主键 `shipment_id`；三者分开建模，保留官方原始 JSON 快照。
- 优先接入官方通知（若站点/权限开放），否则定时增量拉取 + 日终全量对账；两者都按 `order_id` 幂等 upsert。
- 面单与发票存对象存储，按租户加密、短期签名 URL 下载；保留期限与访问审计。
- 订单状态时间线：创建、支付、打包、发货、签收、退款/索赔；异常状态进入人工队列。
- 拆单/合单作为内部履约单，不直接改写官方订单；所有映射关系可追溯。

### 1.5 售后消息、索赔与问答

#### 1.5.1 三个工作台

智赢把售后拆成三个列表：

| 工作台 | 列表入口 | 主要命令 | 证据 |
|---|---|---|---|
| 售后消息 | `MeliMessge` | `meli_message.fetch` / `detail` / `send_message` / `save_trans_message` | [Meli消息列表.resolved.cs:573](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L573)、[Meli消息列表.resolved.cs:739](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L739) |
| 订单索赔 | `MeliClaim` | `meli_claim.fetch` / `detail` / `refund` / `partial_refund` / `open_dispute` / `get_partial_refund_offers` / `send_message` / `set_remark` / `delete` | [Meli索赔列表.resolved.cs:563](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L563)、[Meli索赔列表.resolved.cs:851](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L851)、[Meli索赔明细.resolved.cs:683](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli索赔明细.resolved.cs#L683) |
| 商品问答 | `MeliQa` | `meli_qa.pull` / `reply` / `save` / `memo` / `delete`；模板 `meli_qatemp.*` | [Meli本土Qa.resolved.cs:289](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli本土Qa.resolved.cs#L289)、[Meli本土Qa.resolved.cs:625](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli本土Qa.resolved.cs#L625) |

#### 1.5.2 消息发送流程

1. 从多语言编辑器取出当前语言文本；为空则提示“请翻译或输入消息内容”。
2. 上传附件，得到 `attachids`。
3. 组装 `aid`、`packid`、`message`、`attachids`。
4. 调 `meli_message.send_message`，返回消息记录。
5. 触发 `AfterSendMsg`，由调用方刷新消息与未读状态。

代码位置：[MeliMsgContainer.resolved.cs:885](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.MeliMsgContainer.resolved.cs#L885)。

直连接口侧，智赢还实现了：

- `GET /marketplace/messages/packs/{packId}`：按 pack 取消息。
- `GET /marketplace/messages/{id}`：单条消息。
- `GET /marketplace/messages/unread?...`：未读。
- `POST /marketplace/messages/attachments?site_id=`：上传附件。
- `GET /marketplace/messages/attachments/{id}?tag=post_sale&site_id=`：下载附件。
- `POST /marketplace/messages/packs/{packId}`：发送消息。

见 [MeliApi.resolved.cs:658](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L658)、[MeliApi.resolved.cs:679](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L679)、[MeliApi.resolved.cs:690](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L690)、[MeliApi.resolved.cs:704](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L704)、[MeliApi.resolved.cs:747](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L747)。

#### 1.5.3 索赔能力

智赢直连的 Claims v2 接口覆盖：

| 能力 | 路径 |
|---|---|
| 搜索索赔 | `/marketplace/v2/claims/search?` |
| 详情 | `/marketplace/v2/claims/{id}` |
| 原因 | `/marketplace/v2/claims/reasons/{reasonid}` |
| 可选解决方案 | `/marketplace/v2/claims/{id}/expected-resolutions` |
| 退款 | `POST /marketplace/v2/claims/{id}/expected-resolutions/refund` |
| 部分退款可选项 | `/marketplace/v2/claims/{id}/partial-refund/available-offers` |
| 部分退款 | `POST /marketplace/v2/claims/{id}/expected-resolutions/partial-refund`，body `{percentage}` |
| 开争议 | `/marketplace/v2/claims/{id}/actions/open-dispute` |
| 是否影响声誉 | `/marketplace/v2/claims/{id}/affects-reputation` |
| 索赔消息 | `/marketplace/v2/claims/{id}/messages` |
| 发消息 | `POST /marketplace/v2/claims/{id}/actions/send-message` |
| 附件 | `POST /marketplace/v2/claims/{id}/attachments` |

代码位置：[MeliApi.resolved.cs:227](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L227)、[MeliApi.resolved.cs:267](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L267)、[MeliApi.resolved.cs:287](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L287)、[MeliApi.resolved.cs:298](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L298)、[MeliApi.resolved.cs:318](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L318)、[MeliApi.resolved.cs:328](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L328)、[MeliApi.resolved.cs:339](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L339)。

#### 1.5.4 判断

**可复用**

- 以 pack/claim/question 为会话主键；消息、索赔、问答分开建模，避免把所有售后塞进一个“工单”表。
- 多语言草稿 + 翻译存储：跨境客服的核心需求。
- 附件两步式（先上传拿 id，再发消息带 `attachids`）。
- 索赔动作白名单 + 结果校验（退款、部分退款、开争议、声誉影响）。
- 模板（`meli_qatemp.*`）与备注（`memo`/`set_remark`）适合做 AI 知识库与人工交接。

**不可照搬**

- **token 拼进附件 URL**：`GetMessageAttachmentUrl` 在查询串追加 `&access_token=`。见 [MeliApi.resolved.cs:733](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L733)。
- **退款/争议只有单次确认**：界面弹一次 `Confirm` 后直接调 `meli_claim.refund` / `open_dispute`，没有多级审批、限额或事后复核。见 [Meli索赔明细.resolved.cs:713](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli索赔明细.resolved.cs#L713)。
- **消息与索赔删除混在一起**：消息列表里也出现 `meli_claim.delete`，边界不清。见 [Meli消息列表.resolved.cs:441](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L441)。
- **缺少完整证据链**：在已查看的代码中，未见“操作前快照 + 操作人 + 平台响应 + 结果复核”的完整模型。

**本 ERP 的做法**

- 统一 `Conversation Service`：`channel`（QA / message / claim）、`thread_key`、参与方、消息类型、附件、翻译、SLA、状态。
- 索赔状态机：`open → negotiating → escalated → resolved / closed`；每个动作有前置条件、审批级别、证据附件、平台响应快照。
- AI 只做分类、摘要、话术草稿、退款方案建议；发送/退款/开争议必须人工确认（L4 暂不开放）。
- 事件接入：官方通知（若可用）+ 定时增量；按 `message_id` / `claim_id` 去重；未读与 SLA 超时进入队列。
- 附件走对象存储 + 短期签名 URL，绝不把 Meli token 放进 URL、日志或前端。

### 1.6 汇总：可复用 / 不要照搬

#### 可复用清单

| 主题 | 具体做法 | 价值 |
|---|---|---|
| 账号模型 | `aid` + `site` + `siteuserid` + `shopid` + `isCbt` + `currency` + `quota` | 多店/多站点隔离的基础 |
| 授权校验 | 返回用户 ID 与当前店铺比对 | 防错绑、防串店 |
| Token 生命周期 | 到期前 15 分钟刷新、覆盖 refresh、失败重授权 | 降低 401/失效 |
| 刊登主键 | `aid` + `saleid` + 字段级操作 | 可审计、可逐项回滚 |
| 图片上传 | multipart + 429/500/502 重试 | 提高批量刊登成功率 |
| 订单增量 | `startTime` + `status` + 流式回传 | 断点续传、低内存 |
| 面单/发票 | 按需获取、PDF 缓存、批量合并 | 履约高频刚需 |
| 售后分层 | 消息 / 索赔 / 问答三工作台 | 业务边界清晰 |
| 多语言客服 | 草稿 + 翻译 + 模板 | 跨境场景必需 |

#### 不要照搬清单

| 反模式 | 智赢证据 | 风险 | 替代 |
|---|---|---|---|
| token 下发客户端 | [MeliApiBuilder.resolved.cs:30](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs#L30) | 桌面进程/本地记录可读 token | 服务端 Token Vault |
| secret 放查询串 | [MeliApi.resolved.cs:76](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L76) | 代理/WAF/日志留痕 | 请求体或 HTTP Basic |
| token 拼 URL | [MeliApi.resolved.cs:733](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L733) | 浏览器历史、Referer、日志泄露 | Authorization 头 + 签名 URL |
| 关闭证书校验 | 逆向报告“TLS 证书校验被禁用” | 中间人攻击 | 严格校验 + HSTS |
| 内置 AppSecret | [MeliAppInfo.resolved.cs:46](file:///E:/美客多逆向/analysis/decrypted/MeliAppInfo.resolved.cs#L46) | 密钥泄露、应用被封 | 自有应用 + KMS |
| 明文私有 TCP | ZClient 帧无 TLS/签名 | 窃听、篡改、重放 | 官方 HTTPS API |
| 无幂等/无 dry-run | 字段命令直接提交 | 重复改价、误下架 | 幂等键 + diff + 审批 |
| DOM 自动化 | 扩展“批量关闭折扣” | 违反条款、账号验证、UI 漂移 | 官方 Promotions API |

---

## 2. 批量关闭活动插件专项

> 证据边界：本节只检查扩展安装目录中的代码与切片，未读取浏览器 Cookie、历史记录或 `Local Extension Settings`，未运行扩展，也未登录任何卖家账号。

### 2.1 先排除一个误判

桌面端确实有一个叫“优惠活动”的模块，但它是 **Lazada 专用**：`selAcc.Items = Enumer.Lazada.Acc`，按钮只有“加载”“参加”，没有“关闭”。见 [ZYing.Channel.优惠活动.cs:20](file:///E:/美客多逆向/analysis/decompiled/user_plugin/ZYing.Channel.优惠活动.cs#L20)、[ZYing.Channel.优惠活动.cs:48](file:///E:/美客多逆向/analysis/decompiled/user_plugin/ZYing.Channel.优惠活动.cs#L48)。`D一键报名` 的“结束”也只是把 `stop` 置真以停止报名任务，不是关闭活动。见 [ZYing.Dialog.D一键报名.cs:103](file:///E:/美客多逆向/analysis/decompiled/user_plugin/ZYing.Dialog.D一键报名.cs#L103)。

用户看到的“批量关闭活动”，实际是智赢 Chrome 扩展 5.1.25 的功能：

| 项目 | 事实 |
|---|---|
| 扩展路径 | `...\Google\Chrome\User Data\Profile 11\Extensions\pembniiibielncgmegepmgcbkieljieh\5.1.25_0` |
| 工作副本 | [analysis/zying_ext/extension/5.1.25_0](file:///E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0) |
| 清单 | MV3；content script 注入所有 URL；host permissions 为 `http://*/*`、`https://*/*`。见 [manifest.json:9](file:///E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0/manifest.json#L9)、[manifest.json:24](file:///E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0/manifest.json#L24) |
| 功能名 | UI “批量关闭折扣”；feature id `MELI_SELLER_BATCH_STOP_DISCOUNT`。见 [promo_700k_780k.pretty.js:148](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L148)、[promo_700k_780k.pretty.js:2263](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L2263) |
| 生效页面 | `sellers.mercadolibre.com/listings/list/promos`，排除 `.../online-editing`。见 [promo_700k_780k.pretty.js:2245](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L2245) |

### 2.2 实现拆解

#### 2.2.1 入口与前置

- 在页面右下角注入悬浮按钮 `zy-batch-stop-floating-btn`，React 挂载；DOM 被重建时用 `MutationObserver` 重新挂载。见 [promo_700k_780k.pretty.js:2245](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L2245)。
- 点击前有登录门槛：调用 `getUserInfo()` 检查智赢账号；本地登录时间戳有效期 72 小时（`2592e5` 毫秒）。见 [promo_700k_780k.pretty.js:1921](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1921)。
- 点击时上报埋点 `feature_id/feature_name/platform/page_type`；这一步只用于统计，不参与关闭逻辑。见 [promo_700k_780k.pretty.js:2263](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L2263)。
- 配置项：折扣阈值默认 15、范围 0–100；可选“关闭负折扣”；操作间隔按商品数自动取值（≤500 → 1500ms，≤1500 → 2000ms，≤3000 → 2500ms，>3000 → 3000ms）。见 [promo_700k_780k.pretty.js:1511](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1511)。

#### 2.2.2 扫描与筛选

- 从商品行读取 SKU、item id、标题、活动名、折扣；多 SKU（variation family）会展开并逐行处理。见 [promo_700k_780k.pretty.js:1252](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1252)、[promo_700k_780k.pretty.js:1271](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1271)。
- 筛选规则：折扣 `< 0` 时仅在“关闭负折扣”开启时纳入；否则折扣 `>= 阈值` 才纳入。见 [promo_700k_780k.pretty.js:1287](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1287)。
- 去重键：`itemId + 活动名 + 折扣` 与 `itemId + 活动名 + 折扣 + 按钮 testid`；同一活动按钮只处理一次。见 [promo_700k_780k.pretty.js:1297](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1297)。
- 多 SKU 总数通过只读接口获取：`/listings/list/promos/api/family-items?...`，`credentials: "include"`，带卖家登录态。见 [promo_700k_780k.pretty.js:544](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L544)。

#### 2.2.3 关闭执行

对每个目标行：

1. 关闭残留弹窗，定位 `Edit` / 编辑按钮；必要时滚动到可视区。
2. 如果按钮本身就是“直接停止”（`list-direct-stop`）：点击 → 等二次确认 → 点击确认 → 等弹窗关闭 → 校验“停止”按钮消失。
3. 否则点击 `Edit` → 等活动弹窗 → 找 “Stop participating” / “停止参与” → 点击 → 等二次确认 → 点击确认 → 校验按钮消失。
4. 编辑弹窗最多重试 4 次；失败按 `ri/rr/rn/ra/rs/ro/rd` 分类记录。
5. 平台锁定、活动不支持停止、弹窗不含停止按钮等情况会跳过并记入日志。

核心代码：[promo_700k_780k.pretty.js:1070](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1070) 到 [promo_700k_780k.pretty.js:1251](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1251)。

#### 2.2.4 拟人化与安全阀

- 用合成 `pointerdown/mousedown/pointerup/mouseup/click` 事件，坐标在元素内随机偏移，避免直接 `element.click()`。
- 每次操作有等待；每 15 次操作插入 12–22.5 秒休息；按商品数增加页间延迟。
- 连续 3 个商品组无法展开规格，或连续 3 个商品组全部关闭失败，主动终止任务，日志写“疑似触发平台限制或网络异常，主动终止任务以保护账号安全”。见 [promo_700k_780k.pretty.js:1388](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1388)、[promo_700k_780k.pretty.js:1457](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1457)。
- 断点保存在 `sessionStorage["batch_stop_resume_state"]`，支持从上次页码继续；支持导出 CSV。见 [promo_700k_780k.pretty.js:1478](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L1478)。

#### 2.2.5 选择器与多语言

- 依赖卖家中心内部选择器：`cdp--promotion-box-button--*`、`.sc-list-collapsible-row__promotion-box`、`.sc-list-row-group`、`.sc-list-family-row`、`.promotion-modal`、`cdp--promotion-modal--delete`、`.sc-list-simple-modal`、`cdp--simple-modal--confirm`。见 [promo_700k_780k.pretty.js:163](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L163)。
- 停止按钮文本用多语言正则匹配：英/西/葡/中（`stop participating`、`dejar de participar`、`deixar de participar`、`停止参与` 等）。见 [promo_700k_780k.pretty.js:163](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L163)。

#### 2.2.6 网络行为

- 核心关闭流程**没有调用官方写接口**，全部是 DOM 点击。
- 唯一确认的美客多网络请求是只读的 `family-items` 列表接口，用于拿多 SKU 总数；带 `credentials: "include"`，依赖浏览器里的卖家登录态。
- 核心关闭流程未发现依赖智赢后端；智赢账号只用于登录门槛与埋点。
- 扩展还包含“批量参加折扣”“批量删除”等功能；其中“批量参加”明确声明不会自动提交，关闭折扣则会完成“停止参与 + 确认”。

### 2.3 依赖清单

| 依赖 | 说明 | 可替代性 |
|---|---|---|
| 美客多卖家登录态 | 浏览器 Cookie/会话，`family-items` 与页面操作都依赖 | 官方 OAuth 可替代 |
| 卖家中心 DOM | 选择器、按钮文本、弹窗结构、分页 | 无稳定性承诺，UI 改动即失效 |
| Chrome MV3 | content script、React、Ant Design、i18n | 可重写 |
| 智赢账号 | 登录门槛与埋点（72 小时本地时间戳） | 可移除 |
| 多语言文本 | 停止/编辑/确认按钮的正则 | 需持续维护 |
| 更新通道 | Chrome 商店自动更新 | 自建需自己做版本管理 |

### 2.4 能不能单独拿出来？

#### 2.4.1 技术可行性

**结论：可以拆出“逻辑”，但不能“拿一个文件就跑”。**

需要做的工作：

1. 从 `content.js` 中剥离“批量关闭折扣”相关模块（原始切片约 93KB，可读化后约 116KB；源码为压缩混淆代码）。
2. 重写依赖：React/Ant Design 组件、i18n、storage、埋点、登录门槛、UI 皮肤。
3. 收窄 `host_permissions` 到 `sellers.mercadolibre.com`；移除全站注入。
4. 用配置化选择器适配不同站点/语言；建立 UI 变更告警与回归测试。
5. 增加日志、失败重试、断点、导出；保留原有的安全阀（不要关闭）。
6. 明确更新与分发方式（企业内部分发 / 商店审核 / 自建 CRX）。

工作量取决于“只给自己用”还是“给别人用”：前者是脚本级改造，后者需要产品化、测试、支持与合规审查。

#### 2.4.2 合规可行性

**结论：不建议作为商业产品，也不建议作为 ERP 的生产链路。**

主要风险：

| 风险 | 说明 | 依据 |
|---|---|---|
| 非官方接口 | DOM 自动化不是官方 API；平台页面结构、按钮文本、风控随时变化 | 扩展的失败分类已覆盖“目标行已变动/编辑弹窗未弹出/Edit 未消失”等 UI 漂移场景（`ri/rr/ro`） |
| 合同红线 | 可能触及“不逆向/不抓取/不过量访问/不做竞品”等条款 | 官方合同第 7 条（见 Sources） |
| 账号安全 | 扩展自己提示“请勿同时打开多个窗口，容易触发账号安全验证” | [promo_700k_780k.pretty.js:148](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js#L148) |
| 无审计/无幂等 | 点击成功与否靠 DOM 校验；断点、重试、并发都可能造成重复/漏操作 | 代码中的失败分类与断点逻辑 |
| 版权/授权 | 代码、UI、文案属于智赢；没有开源许可 | 扩展清单与代码头部 |
| 商店政策 | 全站 host permission、自动化操作可能触发商店审核 | `manifest.json` |

#### 2.4.3 三种选择

| 方案 | 技术难度 | 合规风险 | 建议 |
|---|---|---|---|
| 直接抽取智赢扩展代码 | 中 | 高（版权 + 条款 + 账号） | **不做** |
| 参考交互，重写一个内部辅助工具 | 中高 | 中高（仍是非官方 DOM 自动化） | 仅限内部、默认人工确认、不绕过验证；不作为产品卖点 |
| 用官方 Promotions API 重建 | 中 | 低（在授权与条款内） | **推荐** |

### 2.5 推荐替代：官方 Promotions API

#### 2.5.1 能力映射

| 智赢扩展动作 | 官方 API 做法 |
|---|---|
| 列出活动/参与商品 | `GET /seller-promotions/promotions/{promotion_id}`、`GET /seller-promotions/promotions/{promotion_id}/items`（以官方文档为准） |
| 判断折扣是否达到阈值 | 用返回的参与价/折扣在 ERP 侧计算，保留快照 |
| 关闭单个商品活动 | `DELETE /seller-promotions/items/{ITEM_ID}?promotion_type=...&promotion_id=...&app_version=v2` |
| 价格折扣类型 | `DELETE /seller-promotions/items/{ITEM_ID}?promotion_type=PRICE_DISCOUNT&app_version=v2` |
| 复核 | 关闭后重新 `GET .../items`，确认商品已不在活动 |
| 批量 | ERP 侧队列逐条调用，限流 + 重试 + 逐条结果 |

官方示例（Traditional Campaigns）：`DELETE /seller-promotions/items/{ITEM_ID}?promotion_type=DEAL&promotion_id=...&app_version=v2`。价格折扣文档给出 `promotion_type=PRICE_DISCOUNT` 的变体。见 Sources。

#### 2.5.2 推荐流程

```text
每日/每小时同步活动与参与商品
        ↓
卖家设置规则（折扣阈值、是否含负折扣、品类/店铺白名单）
        ↓
ERP 生成“待关闭清单” + 影响预览（商品、原价、活动价、折扣、预计损失）
        ↓
人工审批（或低风险限额自动，需租户显式开启）
        ↓
逐条 DELETE（幂等键 + 限流 + 429 退避）
        ↓
重新拉取参与列表复核 → 成功/失败/跳过分类
        ↓
审计日志 + CSV 导出 + 失败重试
```

#### 2.5.3 关键设计

- **权限**：申请 Promotions/Coupons scope；缺失时识别 403 `PA_UNAUTHORIZED_RESULT_FROM_POLICIES`。
- **活动类型**：`promotion_type` 按活动实际类型传；不要把 `DEAL` 套到所有活动。
- **幂等**：`tenant + store + promotion_id + item_id + action` 作为幂等键；重复 DELETE 返回“已不在活动”按成功处理。
- **限流**：按租户/应用做令牌桶；429 指数退避 + 抖动；批量任务可暂停/继续。
- **审批**：关闭活动会直接影响曝光与价格，默认人工审批；AI 只能生成清单与建议。
- **覆盖缺口**：官方 API 不一定覆盖所有活动类型/站点；不支持的场景回退为“深链到卖家中心 + 人工操作”，**不要**用 DOM 自动化补洞。
- **验证**：关闭后必须回读；DOM 自动化常见的“点了但没生效”在 API 侧也会出现（异步、锁定），所以复核是必需步骤。

### 2.6 结论

- **它是什么**：智赢 Chrome 扩展的“批量关闭折扣”，基于卖家中心 DOM 自动化，不是桌面端功能，也不是官方 API。
- **怎么实现**：扫描活动行 → 按折扣阈值筛选 → 定位编辑/直接停止按钮 → 点击停止 → 二次确认 → DOM 校验；配合拟人化事件、延迟、安全阀、断点与 CSV。
- **能否单独拿出来**：技术上可拆、需重写；商业与合规上不建议。真要做“批量退出活动”，用官方 Promotions API 重建，保留同样的预览、阈值、审批、复核体验。

---

## 3. 合规接入清单

### 3.1 Gate 定义

| Gate | 时点 | 目标 |
|---|---|---|
| G0 | 设计/开发前 | 应用、合同、数据角色、AI 边界明确 |
| G1 | 内测前 | OAuth、Token、权限、审计、安全基线可用 |
| G2 | 公测/GA 前 | 生产安全、限流、灾备、AI 护栏、退出机制完成 |
| G3 | 申请 DPP 前 | GMV、安全评估、交付指标、Partner Center 资料就绪 |

### 3.2 A. 应用与品牌

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| A1 | 在目标站点注册开发者应用，确认应用归属与管理员 | G0 | 应用后台截图、应用 ID 记录（脱敏） |
| A2 | 确认 Mercado Libre 与 Mercado Pago 应用拆分要求；官方公告称 2026-08-30 起生效，落地前在后台复核 | G0 | 应用类型、scope 清单、后台提示 |
| A3 | 应用名称/Logo/描述不使用 “mercado”“libre”“pago”“envíos”“crédito” 等易混淆词 | G0 | 品牌审查记录、应用商店素材 |
| A4 | Redirect URI 使用精确 HTTPS 地址，无通配符/变量；测试与生产分离 | G0 | 应用配置、回调测试记录 |
| A5 | 发布隐私政策、服务条款、卖家数据处理协议（DPA）模板 | G0 | 法务审批版本与链接 |
| A6 | 建立安全联系人与漏洞响应流程，承诺发现漏洞 24 小时内通知 Meli | G0 | 安全联系人、SOP、演练记录 |

### 3.3 B. OAuth 与 Token

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| B1 | 使用授权码模式，授权码在服务端换取 token，不在前端暴露 client_secret | G0 | 时序图、代码评审 |
| B2 | 授权请求带 `state` 并校验回调，防 CSRF | G0 | 单测/渗透测试报告 |
| B3 | 自有 OAuth 若支持 PKCE 则启用 S256；标准 Meli API 文档未强制 PKCE，不要把它当作 Meli 的硬要求 | G0 | 配置说明、兼容性测试 |
| B4 | client_secret 存放在密钥管理服务，不写入代码、配置文件或镜像；不放查询串 | G0 | 密钥管理截图、代码扫描 |
| B5 | token 以 AES-256 加密落库，加密密钥在 KMS/Secrets Manager，不与密文同库 | G0 | 表结构、KMS 配置、渗透测试 |
| B6 | Access Token 按 6 小时有效期设计，提前刷新；每次刷新生成的新 refresh token 必须覆盖保存 | G1 | 刷新日志（脱敏）、故障演练 |
| B7 | 同一租户的刷新操作加分布式锁，避免并发刷新导致旧 refresh token 失效 | G1 | 并发测试报告 |
| B8 | 卖家解绑/卸载应用时删除 token；疑似泄露时调用撤销接口并要求重新授权 | G1 | 解绑流程、撤销测试 |
| B9 | token 不出现在日志、APM、错误上报、LLM prompt、前端配置中 | G1 | 日志抽样、DLP 扫描 |
| B10 | scope 升级触发重新授权；缺失权限时正确识别 403 `PA_UNAUTHORIZED_RESULT_FROM_POLICIES` | G1 | 权限矩阵、错误码测试 |
| B11 | 使用生产环境测试卖家账号；官方不提供 sandbox，禁止用个人主账号测试 | G1 | 测试账号清单、操作记录 |
| B12 | 桌面/浏览器客户端不持有 Meli access/refresh token；只持有本 ERP 会话 | G1 | 客户端二进制/前端检查、抓包 |
| B13 | 授权码、client_secret、token 均不进入 URL 查询串 | G1 | 代码扫描、代理日志抽样 |

### 3.4 C. 权限与数据

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| C1 | 按最小权限申请 scope：Users、Publication & Sync、Sales & Shipments、Communication、Metrics、Promotions、Advertising、Billing 逐项论证 | G0 | scope 申请理由表 |
| C2 | 与卖家签署 DPA：卖家为数据控制者，ERP 为处理者 | G0 | DPA 模板与签署记录 |
| C3 | 仅按卖家授权与合同目的处理数据，不做二次营销、不跨租户聚合训练 | G0 | 数据处理清单、目的限制说明 |
| C4 | 建立数据分类、保留期限、删除与导出机制；终止合作后按合同删除 | G1 | 数据地图、保留策略、删除演练 |
| C5 | 披露子处理者（云、LLM、日志、客服外包），签署同等保护条款 | G1 | 子处理者清单、DPA |
| C6 | 明确跨境传输机制与数据驻留选项（巴西/墨西哥/阿根廷等） | G1 | 法务意见、区域配置 |
| C7 | 对买家 PII 做最小化、脱敏、访问控制；LLM 输入默认去标识化 | G2 | 脱敏规则、抽样审计 |
| C8 | 多租户隔离：数据、token、缓存、向量库、对象存储均按租户分区 | G2 | 隔离测试、越权测试 |
| C9 | 支持卖家数据导出与账号注销；注销后 token 与业务数据按策略清除 | G2 | 注销流程、删除证明 |
| C10 | 明确“用户数据”的使用边界：授权快照、店铺资料、配额、标签只用于运营与对账，不用于跨租户画像 | G1 | 字段清单、用途映射 |

### 3.5 D. 安全基线

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| D1 | 全链路 TLS 1.2+，启用证书校验，禁止跳过证书验证；启用 HSTS | G1 | 配置扫描、TLS 报告 |
| D2 | 无硬编码密钥/固定 DES/内置 AppSecret；使用密钥管理服务 | G1 | SAST/secret scan 报告 |
| D3 | 管理员 MFA、RBAC、最小权限、离职回收 | G1 | IAM 配置、审计日志 |
| D4 | 不可篡改审计日志：谁在何时对哪个租户执行了什么操作、模型版本与审批人 | G1 | 日志样例、保留策略 |
| D5 | SAST/DAST/依赖漏洞扫描/容器扫描进入 CI，阻断高危 | G1 | CI 报告、修复记录 |
| D6 | 上线前第三方渗透测试，重点覆盖 OAuth、越权、token 泄露、提示注入 | G2 | 渗透测试报告与复测 |
| D7 | 事件响应预案：漏洞 24 小时内通知 Meli、卖家通知、取证、恢复 | G1 | IR 演练记录 |
| D8 | 备份/灾备/RTO/RPO 演练；密钥与 token 可恢复 | G2 | 演练报告 |
| D9 | API 滥用防护、限流、队列、熔断、429 退避与抖动 | G1 | 压测报告 |
| D10 | 按 Meli 安全评估要求准备证据；DPP 目标 ≥65% | G3 | 自评表、整改记录 |
| D11 | 不复制智赢的固定密钥、明文配置、私有协议；桌面端如存在本地密钥，必须进 OS 密钥库 | G1 | 代码扫描、配置检查 |
| D12 | 浏览器扩展/桌面组件纳入供应链管理：签名、权限最小化、更新校验 | G2 | 扩展清单、签名与更新记录 |

### 3.6 E. API 使用与合同红线

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| E1 | 按租户/应用做调用预算、并发限制、429 指数退避与抖动；不把限流当无限资源 | G1 | 限流配置、压测 |
| E2 | 不使用 robots、harvester、spider、scraping 或任何非官方方式获取数据 | G0 | 架构评审、代码扫描 |
| E3 | 不逆向、反编译、绕过技术限制或干扰 Meli 服务器/网络 | G0 | 合规声明、供应商审查 |
| E4 | 不转售、不再许可 API；不提供“实质等同 API”的产品 | G0 | 商业模式评审 |
| E5 | 不用 Meli 内容开发竞品或竞争服务 | G0 | 产品边界说明 |
| E6 | 不把 Meli 内容存储用于合同外目的；缓存仅限运营必要 | G1 | 数据流图、保留策略 |
| E7 | 品牌与素材合规；不暗示官方背书 | G0 | 品牌审查 |
| E8 | 建立 Meli 新闻/弃用跟踪机制；强制变更按 15 天窗口响应 | G2 | 变更日历、责任人 |
| E9 | 终止/退出预案：停止调用、删除内容与个人信息、停止使用 Meli 商标 | G2 | 退出 Runbook |
| E10 | 生产链路不使用 DOM 自动化、浏览器插件代操作卖家中心；辅助工具必须明示非官方、默认人工确认、不绕过验证 | G0 | 架构评审、产品文案 |
| E11 | 不规避账号安全验证、验证码、风控挑战；遇到验证即暂停并转人工 | G0 | 运行手册、告警记录 |
| E12 | 促销报名/退出优先用官方 Promotions API；不支持的场景回退人工，不用脚本补洞 | G0 | 功能清单、API 覆盖表 |

### 3.7 F. Developer Partner Program（可选）

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| F1 | 核对目标国月 GMV 门槛：巴西 250 万美元、阿根廷 150 万、墨西哥 100 万、智利/乌拉圭/哥伦比亚 20 万、秘鲁 10 万 | G3 | GMV 报表 |
| F2 | 完成安全评估并达到 65% 以上 | G3 | 评估报告 |
| F3 | 接受每季度最多 4 个指派任务（2 高优 + 2 中优，不含强制变更与技术债） | G3 | 任务台账 |
| F4 | 按奖牌等级满足交付时限：强制变更 15 天；紧急 8 小时 + 48 小时复盘；P0 60/50/40 天；P1 75/65/55 天；P2 80/75/70 天 | G3 | SLA 看板 |
| F5 | 建立隔离/降级应对：黄牌 40 天、红牌 40 天、仍未交付则取消认证 | G3 | 应急预案 |
| F6 | 准备 Partner Center 资料、支持 SLA、区域化申请 | G3 | 申请材料 |
| F7 | 评估区域化要求（本地计费、葡萄牙语/西班牙语等） | G3 | 区域化清单 |

> 按官方文档结构推断，DPP 是“认证/曝光/支持”体系，不应作为第一版 ERP 的前置依赖；先做好普通接入与产品价值，再按 GMV 与战略决定是否申请。

### 3.8 G. AI 专项

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| G1 | 建立 AI 用例登记表：用途、数据来源、是否写操作、风险级别、责任人 | G0 | AI 用例清单 |
| G2 | 明确训练/微调的法律依据；未经授权不使用 Meli 数据或跨租户数据训练 | G0 | 法务意见、数据授权链 |
| G3 | 截至本次检索（检索词：inteligencia artificial、entrenar、machine learning），官方合同未发现“禁止 AI 训练”的明文条款，但目的限制、竞品限制、反抓取仍适用；不得把“未禁止”当“已授权” | G0 | 法务复核记录 |
| G4 | LLM 供应商签署 DPA，承诺不训练、不留存、可区域化；敏感数据脱敏后再调用 | G1 | 供应商 DPA、配置 |
| G5 | 提示注入防护：把商品描述、买家消息、文档当不可信数据；工具白名单、参数校验、最小权限 | G2 | 红队测试报告 |
| G6 | AI 生成内容发布前经官方 Listing Validator/Moderation 或内部规则校验 | G2 | 校验通过率、拦截记录 |
| G7 | 高风险写操作 human-in-the-loop：改价、索赔、退款、下架、广告预算、批量刊登、批量退出活动 | G2 | 审批流、抽样审计 |
| G8 | 记录模型/版本/prompt 哈希/输入输出/审批人；支持回放与追责 | G2 | 审计日志 |
| G9 | 建立离线评测集与回归测试：事实性、合规性、越权、幻觉、成本 | G2 | 评测报告 |
| G10 | 提供 AI 总开关与租户级关闭；异常时回退人工流程 | G2 | 开关演练 |
| G11 | 按适用法律与平台规则做 AI 披露/买家告知 | G2 | 文案与配置 |
| G12 | 设 token/成本/调用上限，防止 Agent 循环与账单失控 | G2 | 预算告警、熔断 |

### 3.9 H. 参考实现与知识产权

| ID | 检查项 | Gate | 验收证据 |
|---|---|---|---|
| H1 | 对智赢只做行为/流程参考，采用 clean-room 方式重写；不复制反编译代码、类名、字符串与资源 | G0 | 代码评审、相似度扫描 |
| H2 | 不使用智赢的 AppId/AppSecret/RedirectUrl/固定密钥；注册自有应用 | G0 | 配置清单、secret scan |
| H3 | 不使用智赢的 Logo、图标、文案、帮助页素材；不暗示与智赢的关系 | G0 | 品牌审查 |
| H4 | 不把扩展的“批量关闭折扣”代码、选择器、UI 直接产品化；如参考交互，需重写并做法务评估 | G0 | 产品评审、法务意见 |
| H5 | 第三方开源依赖保留许可证与 NOTICE；商业组件确认授权范围 | G2 | SBOM、许可证清单 |

### 3.10 Go/No-Go 汇总

| Gate | 必须通过 | 责任人 | 证据包 |
|---|---|---|---|
| G0 | A1-A6、B1-B5、C1-C3、E2-E5、E10-E12、G1-G3、H1-H4 | 产品/法务/安全 | 应用配置、DPA、AI 用例表、合规评审 |
| G1 | B6-B13、C4-C5、C10、D1-D5、D7、D9、E1、E6、G4 | 后端/安全 | OAuth 测试、日志审计、压测 |
| G2 | C6-C9、D6、D8、D10-D12、E7-E9、G5-G12、H5 | 安全/运维/法务 | 渗透报告、演练、AI 评测、SBOM |
| G3 | F1-F7 | 商务/合规 | GMV、评估、SLA 看板 |

---

## 4. 模块路线图

### 4.1 阶段总览

| 阶段 | 周期 | 目标 | 核心模块 | 退出标准 |
|---|---|---|---|---|
| P0 合规底座 | 0-4 周 | 能安全地授权、存 token、调 API、审计 | 租户、OAuth、Token Vault、API Gateway、审计、策略引擎 | G0/G1 通过；单店铺联调成功 |
| P1 只读 ERP | 4-10 周 | 把店铺数据变成可用报表与洞察 | 连接中心、商品/订单/物流/客服/促销只读、指标、数据仓库 | 同步成功率 ≥99%；无平台写权限 |
| P2 交易闭环 | 10-18 周 | 能安全地写回业务动作 | 刊登、价格库存、履约、面单、问答/消息、售后处理、促销退出 | 幂等/回滚/审批完备；G2 通过 |
| P3 AI 增强 | 18-28 周 | AI 真正融入日常运营 | Copilot、内容工作室、客服助手、定价顾问、促销顾问、异常检测 | AI 评测达标；人工接管可用 |
| P4 规模化与 DPP | 28 周+ | 多国、多仓、生态认证 | 多站点、区域化、DPP、Partner Center、高级 Agent | 目标国 GMV/安全/交付达标 |

### 4.2 P0 合规底座（0-4 周）

**模块与交付物**

- 多租户与店铺模型：tenant、seller、site、store、role；对齐智赢的 `aid/site/siteuserid/shopid/isCbt/currency/quota` 语义，但不使用其字段名与协议。
- 设备与会话：设备注册、短期会话 + 轮换 refresh、设备列表/吊销、系统密钥库（Keychain / Credential Manager）存储。
- 桌面客户端骨架：Tauri 2 + React/TypeScript；代码签名、更新校验、SQLCipher 本地缓存与离线 outbox 雏形。
- OAuth 服务：授权、回调、刷新、撤销、重新授权；`state` 校验；服务端换 token。
- Token Vault：AES-256、KMS、租户隔离、审计、并发锁。
- API Gateway：统一鉴权、限流、重试、429 退避、幂等键、错误码映射；统一 `Authorization: Bearer`。
- 官方 API 客户端：`/users/me`、`/users/{id}`、`/marketplace/users/{id}`、`/marketplace/users/cap`、`/items`、`/user-products`、订单/消息/索赔基础读。
- 审计与可观测：结构化日志、trace、指标、告警、token 脱敏。
- 策略引擎雏形：scope 校验、租户隔离、写操作白名单。
- 合规文档：隐私政策、DPA、子处理者清单、AI 用例登记表。

**验收**

- 单店铺完成授权 → 调用 `/users/me` → 刷新 → 解绑 → token 删除全流程。
- 日志中无 token/PII；越权测试无法跨租户读取。
- 429、401、403 均有可观测指标与告警。

### 4.3 P1 只读 ERP（4-10 周）

**模块**

- 连接中心：多站点授权、店铺健康度、token 到期提醒、配额与标签。
- 商品中心（只读）：商品、类目、属性、变体、图片、Listing 质量、违规/降级。
- 订单中心（只读）：订单、打包、物流轨迹、异常订单、发票。
- 客服中心（只读）：问答、售后消息、未回复提醒、索赔列表。
- 促销中心（只读）：活动列表、参与商品、折扣分布、即将到期。
- 指标中心：流量、转化、销售、广告、声誉。
- 数据仓库：按租户分区，保留原始快照与规范化模型。
- 同步服务：快照清单 + `sync_changes` 游标 + SSE 提示 + 幂等命令通道（P1 只写 ERP 自有数据）；每日对账与死信。
- 本地读模型：SQLite + SQLCipher + FTS5，按租户隔离；媒体按签名 URL 按需下载。

**AI（仅 L1 洞察）**

- 销售趋势/异常检测、库存风险、Listing 质量评分、客服积压摘要、促销参与异常。
- 所有输出只读，不触发任何写操作。

**验收**

- 增量同步延迟 < 5 分钟（按业务可接受度调整）。
- 同步成功率 ≥99%，失败可重放。
- 数据口径与 Meli 后台抽样对账一致。

### 4.4 P2 交易闭环（10-18 周）

**模块**

- 刊登与同步：新建/修改/下架、属性补全、变体、图片；普通站点 `/items`，CBT `/user-products`。
- 价格与库存：单改/批量、多仓、价格保护、回滚；支持按净收益改价。
- 订单履约：打包、面单、发货、取消、异常处理；拆单/合单作为内部履约单。
- 客服与售后：问答回复、售后消息、Claims/Returns 处理；退款/争议强制审批。
- 促销与广告：活动报名、优惠券、**官方 API 批量退出活动**、预算与出价（先只读/建议）。

**AI（L2 草稿 + 人工确认）**

- Listing 文案/属性建议、客服回复草稿、改价建议、补货建议、促销退出清单。
- 每个写操作展示 diff、影响范围、回滚方案。

**验收**

- 幂等键覆盖所有写接口；重复请求不产生重复动作。
- 批量操作可暂停/回滚；失败逐条可重试。
- 高风险操作 100% 人工审批；审批记录可追溯。
- 促销退出有“预览 → 审批 → 执行 → 回读复核”闭环，误关闭率为 0。

### 4.5 P3 AI 增强（18-28 周）

**模块**

- AI Copilot：跨模块问答 + 工具调用（只读优先）。
- 内容工作室：标题、描述、属性、图片合规检查。
- 客服助手：分类、摘要、草稿、情绪与升级建议；多语言翻译。
- 定价顾问：竞品/成本/转化/库存综合建议。
- 促销顾问：活动 ROI、参与/退出建议、预算与折扣模拟。
- 异常中心：订单、物流、广告、财务、促销异常检测。
- 财务对账：账单、发票、费用差异（先只读）。

**AI 自主级别**

| 级别 | 行为 | 适用 |
|---|---|---|
| L0 | 无 AI | 基础规则 |
| L1 | 只读洞察 | 报表、预警、评分 |
| L2 | 生成草稿，人工确认后执行 | 文案、客服、改价建议、促销退出清单 |
| L3 | 低风险自动执行 + 抽样审计 | FAQ 自动回复、低金额库存同步 |
| L4 | 高风险全自动 | **暂不建议**（索赔、退款、批量改价、广告预算、批量退出活动） |

**验收**

- 离线评测集覆盖事实性、合规、越权、幻觉、成本。
- 人工接管率、建议采纳率、误操作率有基线。
- 任何 AI 写操作可一键关闭并回退。

### 4.6 P4 规模化与 DPP（28 周+）

- 多站点/多币种/多语言/多仓。
- 区域化：本地计费、税务/发票、葡语/西语。
- DPP：GMV、65% 安全评估、季度任务、SLA、Partner Center。
- 高级 Agent：跨店铺组合优化、供应链协同、广告组合投放。
- 生态：对外开放 API/插件，但仍需遵守不转售 API、不做竞品等条款。

### 4.7 模块-权限-风险对照

| 模块 | 核心能力 | 主要 Meli 权限 | AI 增强 | 风险 | 关键验收 |
|---|---|---|---|---|---|
| 连接中心 | 授权/解绑/多站点 | Users | 授权异常诊断 | 低 | token 轮换/撤销 |
| 商品中心 | 刊登/属性/类目/变体/图片 | Publication & Sync | 属性补全、文案、图片检查 | 中 | Listing Validator 通过率 |
| 价格库存 | 改价/库存/多仓 | Publication & Sync | 定价建议、缺货预测 | 中高 | 审批 + 回滚 |
| 订单履约 | 订单/打包/面单/物流 | Sales & Shipments | 异常检测、时效预测 | 中 | 幂等 + 对账 |
| 客服 | 问答/消息 | Pre/Post-Sales Communication | 草稿/分类/情绪 | 中 | 人工发送起步 |
| 售后 | Claims/Returns | Sales & Shipments + Communication | 归因、话术、方案建议 | 高 | 人工审批 + 证据链 |
| 营销 | 促销/优惠券/广告 | Promotions/Coupons + Advertising | ROI、预算建议、退出清单 | 中高 | 预算上限 + 复核 |
| 数据报表 | 流量/转化/利润 | Business Metrics + Sales | 预测/异常/归因 | 低 | 口径对账 |
| 财务 | 账单/发票 | Billing | 对账异常 | 高 | 人工复核 |
| AI Copilot | 跨模块问答/操作 | 复用上述 | RAG + 工具调用 | 高 | 权限继承 + 审计 |

### 4.8 参考架构

```text
卖家 / 运营人员
      │
      ▼
Desktop Client（Windows / macOS，Tauri 2）
      ├── UI：React + TypeScript
      ├── 本地：SQLite 缓存 / 草稿 / 打印 / 文件
      └── Local IPC ◄── MV3 AI Side Panel（只读上下文桥）
      │ HTTPS + SSE（客户端不持有 Meli token）
      ▼
Control Plane（云端或客户私有化）
      ├── API Gateway ─── 租户鉴权 / 限流 / 审计 / 幂等
      ├── OAuth Service ─── Token Vault（AES-256 + KMS）
      ├── Domain Services
      │     ├── 连接 / 商品 / 价格库存 / 订单 / 履约
      │     ├── 客服 / 售后 / 促销 / 报表 / 财务
      │     └── 策略引擎（金额、品类、时间窗、白名单）
      ├── Sync Workers ─── 官方 REST API（每租户 token）
      │        ├── 429 退避 / 重试 / 死信 / 对账
      │        └── 通知接收器 + 定时增量
      ├── Event Bus ─── 运营库 / 数据仓库 / 指标
      └── AI Orchestrator
            ├── 模型路由 / 脱敏 / 提示注入防护
            ├── 工具网关（自建 Tool API）
            ├── 人工审批队列
            └── 评测 / 审计 / 成本控制
```

产品主入口是桌面客户端；Control Plane 不提供面向卖家的 Web 工作台，只提供 API、同步、授权、对象存储与 AI 能力。

**架构原则**

- 产品主形态是桌面客户端：Tauri 2 + React/TypeScript；Rust 原生层负责打印、文件、系统密钥库（Keychain / Credential Manager）、签名更新与开机启动。Web 只做只读分享页与管理后台，不做卖家工作台。
- 客户端只持有本 ERP 的设备会话；Meli token、AI 供应商密钥、推送签名密钥都只在 Control Plane。客户端日志、崩溃转储、APM 不得出现这些值。
- 服务端是业务数据唯一事实源；客户端 SQLite 只是加密只读缓存 + 草稿 + 发件箱，换机后可由服务端重建。
- 同步采用快照 + 游标 + 变更日志，不依赖“用户点按钮、按时间窗拉取”；离线写入进 outbox，恢复后幂等重放。
- MV3 扩展只做 AI 侧栏的上下文桥：读取当前页面公开信息、深链到 ERP；不写 DOM 表单、不碰 cookie、不拿 token。
- 安装包与自动更新必须代码签名（Windows EV / macOS Developer ID + notarization）；更新清单签名校验，禁止静默降级。
- 只用官方 REST API；不复制 ZClient 私有网关，不做 DOM 自动化，不做抓取。
- token 只在 Token Vault 与 API Gateway 内可见，模型、前端、桌面客户端不可见。
- 所有写操作统一走策略引擎；AI 与人工使用同一套权限与审计。
- 写操作先预览 diff，再审批，再执行，再回读复核；批量操作逐条记录。
- 数据仓库按租户分区；向量库不混租户；训练/微调默认关闭。
- 所有外部输入（商品描述、买家消息、平台文档）都按不可信数据处理。

### 4.9 前 90 天执行表

| 周 | 交付 | Gate |
|---|---|---|
| 1-2 | 应用注册、scope 论证、DPA/隐私政策、AI 用例表、参考实现合规评审；Tauri 客户端骨架与签名/更新链路 | G0 |
| 3-4 | OAuth + Token Vault + API Gateway + 审计；设备注册/会话；单店铺联调 | G0/G1 |
| 5-6 | 连接中心、商品/订单只读、基础报表；快照 + 游标同步最小闭环 | G1 |
| 7-8 | 数据仓库、指标口径、异常监控、压测；增量对账、死信与重放 | G1 |
| 9-10 | 客服只读、Listing 质量、促销只读、AI L1 洞察；MV3 上下文桥 | G1 |
| 11-12 | 安全评审、渗透测试、G2 整改；多设备吊销、离线 outbox；准备交易闭环与促销退出 | G2 |

### 4.10 明确不做

- 不做纯浏览器 SaaS 作为主产品；Web 仅限只读分享、账单与管理后台。
- 不做“无服务端、单机即完整产品”的默认形态。单机受限模式仅限：用户自带 Meli 应用、token 进系统密钥库、AI 用用户自己的 key，且明确不提供跨设备同步与团队协作。
- 不做 ZClient 私有协议兼容层；不做“智赢协议代理”。
- 不做客户端直连 Meli API（单机受限模式除外），不在桌面/浏览器客户端保存 Meli token、AI key、推送密钥；不做 token 下发。
- 不做“用户手工点同步”作为主链路；后台增量同步与每日对账必须是默认能力。
- 不做无审计的跨设备登录；新设备必须注册、可见、可吊销，本地缓存必须可远程失效。
- 不做卖家中心 DOM 自动化、不抽取智赢扩展作为产品功能。
- 不做验证码/风控绕过、不做高频轮询、不做抓取。
- 不做未经审批的退款、争议、批量改价、批量退出活动。
- 不用 Meli 数据做跨租户训练或竞品分析。

### 4.11 客户端版跨设备同步：智赢机制审查与本 ERP 设计

需求前提：服务器负责缓存与账号管理；用户换电脑后只登录 ERP 账号，就能看到原有商品、订单、消息与设置。智赢实际上已经是这个模型——**服务端主库 + 瘦客户端**。本节先还原其链路，再给出可落地的跨设备设计。

#### 4.11.1 智赢的同步链路

**一句话**：业务数据都在 ZYing 服务端，客户端只做分页查询与展示；Meli 账号与 token 也挂在服务端账号体系下，所以换电脑登录 ZYing 账号即可恢复业务视图，不需要重新授权 Meli。

```text
ZYing Desktop
  │ ZClient 私有长连接（登录会话 token）
  │   ├── 查询：meli_acc.select / meli_message.list / meli_claim.list ...
  │   └── 触发：meli.order / meli.sync / meli_message.fetch / meli_claim.fetch ...
  ▼
ZYing Server ──► Meli 官方 REST API（服务端持有每账号 token）
  │
  ├── 主库：账号 / 商品 / 订单 / 消息 / 索赔 / QA / 统计数据
  ├── 推送：acc_order + mpush 响应头
  └── 同步水位：stat_synctime / up_synctime / stat_visit_time ...
  │
  ▼
Desktop：按 loginid + 分页查询 → 表格/详情
本地：配置、登录记录、PDF、消息附件缓存
```

| 层 | 智赢做法 | 证据 |
|---|---|---|
| 身份与会话 | 业务员账号登录；`Member.Token` 是会话 GUID；服务端返回上次登录 IP `iplast`，登录时以“异地登录”文案提示；未发现设备 ID、设备列表或远程吊销 | [ZYing_Member.work.resolved.cs:165](file:///E:/美客多逆向/analysis/decrypted/ZYing_Member.work.resolved.cs#L165)、[ZYing_Member.work.resolved.cs:698](file:///E:/美客多逆向/analysis/decrypted/ZYing_Member.work.resolved.cs#L698)、[ZYing_Member.work.resolved.cs:747](file:///E:/美客多逆向/analysis/decrypted/ZYing_Member.work.resolved.cs#L747)、[MLogin.resolved.cs:433](file:///E:/美客多逆向/analysis/decrypted/MLogin.resolved.cs#L433) |
| 账号归属 | Meli 账号保存在服务端；列表按 `loginid`、`role`、`pagesize=50` 查询，按账号 `id` 取详情 | [Meli授权列表.resolved.cs:727](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli授权列表.resolved.cs#L727)、[Meli授权列表.resolved.cs:799](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli授权列表.resolved.cs#L799)、[Meli授权列表.resolved.cs:801](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli授权列表.resolved.cs#L801)、[Meli授权列表.resolved.cs:809](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli授权列表.resolved.cs#L809)、[MeliApiBuilder.resolved.cs:31](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs#L31) |
| Token | 账号表存 `acc_access` / `acc_fresh` / `acc_expire`；客户端通过 `meli_acc.detail` / `meli_acc.first` 取回并构造 `MeliApi`；刷新结果默认只留在内存 | [Meli授权详情.resolved.cs:1399](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L1399)、[MeliApiBuilder.resolved.cs:31](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs#L31)、[MeliApiBuilder.resolved.cs:48](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs#L48)、[MeliApi.resolved.cs:1052](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs#L1052) |
| 商品/刊登 | 服务端 `Melistat`（`stat_aid` + `stat_saleid` 复合键）与 `Meliup`（`up_aid` + `up_saleid`）；客户端按账号、关键词、时间分页查询 | [Melistat.resolved.cs:21](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Melistat.resolved.cs#L21)、[Meliup.resolved.cs:29](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meliup.resolved.cs#L29) |
| 订单 | 订单详情从 `root` / `items` / `meli_order` 三个记录集读取，含平台订单号、履约、重量尺寸等扩展字段 | [OP美客多.resolved.cs:235](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.OP美客多.resolved.cs#L235)、[OP美客多.resolved.cs:287](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.OP美客多.resolved.cs#L287) |
| 消息/索赔/QA | 服务端表 + 列表命令 `meli_message.*` / `meli_claim.*`；拉取命令 `meli_message.fetch` / `meli_claim.fetch` | [Meli消息列表.resolved.cs:573](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L573)、[Meli消息列表.resolved.cs:739](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L739)、[Meli索赔列表.resolved.cs:563](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L563)、[Meli索赔列表.resolved.cs:851](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L851) |
| 同步水位 | 商品有 `stat_synctime`、`stat_visit_time`、`up_synctime`、`up_visittime`；是“记录级时间戳”，不是客户端可续传的全局游标 | [Melistat.resolved.cs:225](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Melistat.resolved.cs#L225)、[Meliup.resolved.cs:281](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meliup.resolved.cs#L281) |
| 推送 | 账号可勾选“接受订单推送”（`acc_order`）；ZClient 响应头 `mpush` 交给 `MPush.GetPushData`；消息/索赔仅使用 ZYing 应用授权的账号能自动推送 | [Meli授权详情.resolved.cs:2665](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs#L2665)、[ZYing_ZClient.resolved.cs:1106](file:///E:/美客多逆向/analysis/decrypted/ZYing_ZClient.resolved.cs#L1106)、[Meli消息列表.resolved.cs:1093](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L1093)、[Meli索赔列表.resolved.cs:1211](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L1211) |
| 本地文件 | PDF 放程序目录 `pdf/`，标签/面单也走同一目录；消息附件放 `Cache.Dir`；没有账号/租户命名空间，也没有 TTL | [打印Meli.resolved.cs:235](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.打印Meli.resolved.cs#L235)、[MeliLabelPrint.resolved.cs:81](file:///E:/美客多逆向/analysis/decrypted/user/Logic.MeliLabelPrint.resolved.cs#L81)、[MeliMsgContainer.resolved.cs:557](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.MeliMsgContainer.resolved.cs#L557) |
| 本地业务库 | 未发现订单/商品/消息的本地数据库；客户端拿不到数据时只能依赖服务端可用 | [ZYing-美客多逆向分析报告.md:229](file:///E:/美客多逆向/ZYing-美客多逆向分析报告.md#L229) |

**同步触发点**

| 用户动作 | 命令 | 参数/窗口 | 证据 |
|---|---|---|---|
| 手工接收订单 | `meli.order` | `aid`、`startTime`、`status`；默认从 2 天前开始，无结束时间 | [ZOrderMeli.resolved.cs:167](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZOrderMeli.resolved.cs#L167)、[任务订单Meli.resolved.cs:115](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.任务订单Meli.resolved.cs#L115) |
| 同步产品/状态 | `meli.sync` | `aid`、`startTime`、`endTime`；UI 默认 30 天 | [ZSyncMeli.resolved.cs:173](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZSyncMeli.resolved.cs#L173)、[Meli状态同步.resolved.cs:119](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli状态同步.resolved.cs#L119) |
| 拉取消息 | `meli_message.fetch` | 选中行给 `packids`；否则每账号一条 `aid` | [Meli消息列表.resolved.cs:739](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs#L739) |
| 拉取索赔 | `meli_claim.fetch` | 选中行给 `claimids`；否则按账号逐天 `aid` + `date`；无日期时默认 3 天 | [Meli索赔列表.resolved.cs:851](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L851)、[Meli索赔列表.resolved.cs:691](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs#L691) |
| 拉取问答/站内信 | `meli_qa.pull` / `meli_question.pull` | `aid`、`stat` | [Meli_Qa.resolved.cs:289](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli_Qa.resolved.cs#L289)、[Meli站内信.resolved.cs:665](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli站内信.resolved.cs#L665) |
| 拉取商品明细 | `meli_stat.sync_items` / `meli_localstat.fetch` | `itemids`，逐页回传进度 | [DMeli同步数据.resolved.cs:275](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.DMeli同步数据.resolved.cs#L275) |
| 补拉缺失商品 | `meli.syncAccProduct` | 仅拉取上传明细里没有的产品 | [Meli上传明细List.resolved.cs:1829](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli上传明细List.resolved.cs#L1829) |
| 清理无效刊登 | `meli_stat.clean` | 手工触发 | [Meli上传明细List.resolved.cs:1851](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli上传明细List.resolved.cs#L1851) |

**换一台电脑时实际发生什么**

1. 登录 ZYing 账号，服务端返回 `Member.Token` 与权限、站点、业务员范围。
2. 客户端按 `loginid` 查询账号列表、商品、订单、消息、索赔，全部从服务端主库读取。
3. Meli token 由服务端保存；只有需要客户端直连的接口（消息、索赔、图片、发票、CBT 等）才会把 token 下发到客户端。
4. 本地需要重新生成 PDF/附件缓存；未发现打印设置、窗口布局等本地配置的服务端同步机制。
5. 未发现设备列表与远程吊销能力，旧设备的会话在服务端侧仍可能继续有效。

#### 4.11.2 可复用与不可照搬

**可复用**

- 服务端作为业务数据唯一事实源，客户端不保存全量业务库——换机成本最低。
- Meli 账号绑定在 ERP 账号/租户下，不绑定设备；用户换机不需要重新授权 Meli。
- 所有列表按 `loginid`（业务员/租户范围）+ 分页查询，权限边界清晰。
- 记录级 `synctime` / `visittime` / `pubtime` 可作为对账与补拉的起点。
- 推送 + 手工补拉双通道；手工路径可作为推送失败的兜底。
- 长任务用流式进度回传，失败可按账号重试。
- PDF/附件本地缓存提升打开速度。

**不可照搬**

- **把“手工时间窗拉取”当主同步**：订单默认 -2 天、商品 -30 天、索赔 -3 天，漏跑就断档；没有断点续传与补偿任务。
- **客户端没有全局游标/变更日志/墓碑接口**：无法问“自我上次同步后变了什么”，删除与状态变更容易漏。
- **协议里没有幂等键与命令日志**：重复点击、网络重试可能重复建单/重复操作；失败后无法安全重放。
- **token 下发客户端**：`meli_acc.detail` 把 access/refresh 返回桌面端，扩大泄露面。
- **自动刷新不回写、未发现分布式锁**：刷新只改内存；refresh token 单次有效，多设备/多进程会互相踢掉。
- **未发现设备/会话管理**：只有最后登录 IP 提示，不能吊销旧设备或远程清缓存。
- **推送依赖 ZYing 应用授权**：自有应用授权时消息/索赔不会自动推送，需要卖家自己配置回调。
- **补拉只看“缺失”**：`meli.syncAccProduct` 只补没有的产品，已有产品的价格/库存/状态变化不会刷新。
- **本地文件无租户隔离**：所有账号共用 `pdf/` 与 `Cache.Dir`，换账号或多人共用电脑时容易串数据。
- **未发现原始报文快照**：服务端至少对客户端只暴露归一化结果，出问题难以回放平台原始响应。

#### 4.11.3 本 ERP 的跨设备同步设计

**目标**：换电脑只登录 ERP 账号，商品、订单、消息、售后、规则、AI 会话与审批记录全部恢复；Meli token 永不下发；离线可读、可草稿，联网后幂等同步。

注意：服务器不能只做缓存。平台镜像可以从 Meli 重建，但 ERP 自有数据、命令日志、审批记录与同步游标必须持久化；否则换机后只能恢复平台数据，恢复不了工作上下文。

**数据分层**

| 层 | 内容 | 存放 | 跨设备 |
|---|---|---|---|
| 平台镜像 | Meli 商品、订单、消息、索赔、问答、指标、原始报文 | 服务端 PostgreSQL + 对象存储 | 是，服务端事实源 |
| ERP 自有数据 | 备注、标签、任务、规则、审批、AI 会话/记忆、操作日志 | 服务端 PostgreSQL | 是 |
| 设备本地 | 窗口布局、打印模板、缓存、离线草稿、日志 | 本地 SQLite（SQLCipher）+ 文件 | 可选：设置加密同步；业务数据不落本地全量 |
| 密钥 | Meli token、AI key、推送签名密钥 | 服务端 KMS/Vault | 否，永不下发 |
| 设备会话 | 设备 ID、会话 token、刷新令牌 | 服务端 + 本地系统密钥库 | 否，每设备独立；服务端可吊销 |

**身份与设备模型**

- `tenant`：卖家/公司；`user`：ERP 账号；`membership`：用户在租户内的角色与店铺范围。
- `device`：安装时生成设备 ID + 公钥；登录时注册设备名、系统、版本。
- `session`：短期 access + 轮换 refresh，绑定 `device_id`；服务端保留设备列表、最后活跃时间、IP/地区。
- 新设备登录：密码/SSO + 二次验证（可策略强制）→ 设备注册 → 可选“旧设备仍在线”提示 → 签发会话。
- 设备吊销：立即失效该设备会话；本地缓存标记为不可解密/待清除；下次联网执行远程清除。
- 权限：所有查询与命令都带 `tenant_id` + `membership` 范围；服务端不做“客户端传什么就信什么”。

**换机恢复流程**

```text
登录（密码/SSO + 2FA）
  → 注册/识别设备，签发设备会话
  → 拉租户与店铺权限（不含 Meli token）
  → 打开本地加密缓存
      ├── 有缓存：GET /sync/changes?cursor=<last_seq>&limit=...
      └── 无缓存：GET /sync/snapshot/manifest → 分页快照 → 从 snapshot_seq 接增量
  → GET /sync/stream (SSE, Last-Event-ID)      ← 实时增量
  → 媒体按需 GET /files/{id}/signed-url
  → 离线草稿进 outbox；联网后 POST /commands 幂等重放
```

**同步协议**

- `sync_changes` 变更日志：`tenant_id`、全局递增 `seq`、`entity_type`、`entity_id`、`op`（upsert/delete）、`version`、`payload_hash`、`changed_at`、`actor`。删除写墓碑，不做物理删除。
- 快照：按实体导出分页清单 + `snapshot_id`；客户端先拿清单，再从该快照对应的 `seq` 之后接增量，避免“边导边漏”。
- 增量：`GET /sync/changes?cursor=<seq>&limit=500`，返回 `next_cursor`、`has_more`；客户端事务内更新本地 `cursor`，保证至少一次 + 幂等去重。
- 实时：`GET /sync/stream` SSE，事件带 `id=seq`；断线用 `Last-Event-ID` 续传；SSE 只做提示，最终以 `changes` 拉取为准。
- 命令：`POST /commands` 带 `Idempotency-Key`（客户端生成 UUID）+ `expected_version`；服务端命令日志去重，返回 `command_id` 与最终 `seq`。
- 冲突：平台镜像以服务端为准；ERP 自有字段用 `version` 乐观锁，冲突返回 409 + 双方 diff，交人工合并；本地草稿永不自动覆盖服务端。
- 对账：每账号维护 `watermark`（订单/消息/索赔/商品各自独立）；每日全量 diff + 时间窗补拉；失败进死信队列并告警。
- 重放：保留原始响应快照（加密、按租户分区、可配置保留期），支持按 `command_id` / `request_id` 回放与追责。

**本地库**

- SQLite + SQLCipher；每个租户独立数据库文件或强制 `tenant_id` 过滤 + 页级加密。
- 表：`entities`（JSON + 索引列）、`changes_applied`（seq 去重）、`outbox`（待发命令）、`drafts`、`files_cache`、`sync_state`。
- FTS5 做商品/订单/消息全文检索；附件只存路径与哈希，正文按需下载。
- 退出登录/设备吊销：删除数据库密钥并清空缓存；保留最小审计日志（不含业务正文）。

**服务端缓存刷新节奏**

| 数据 | 触发 | 兜底 |
|---|---|---|
| 订单 | 官方通知（若有）+ 1–5 分钟增量 | 每 15 分钟按更新时间窗补拉；每日全量对账 |
| 消息/索赔 | 通知 + 1–2 分钟增量 | 5 分钟按账号补拉；每日对账 |
| 商品/价格/库存 | 通知 + 5–15 分钟 | 30 分钟全店扫描；每日差异对账 |
| 指标/报表 | 小时级 | 每日重算 |
| 授权/账号信息 | 登录/手动刷新 | 每日校验 token 与账号状态 |

所有刷新按租户 + 账号限流；429 用带抖动的指数退避 + 优先级队列；平台维护/限流时降级为只读缓存并明确提示数据时间。

**对象存储与 AI**

- 面单、发票、附件、导出报表、AI 生成物进对象存储；服务端签发短时 URL；客户端不持有长期凭证。
- AI Orchestrator 只在服务端；客户端与 MV3 侧栏只传“实体 ID + 用户意图 + 最小上下文”，不传全量订单/消息。
- AI 记忆按 `tenant_id` + `user_id` 分区；跨租户检索默认禁止；敏感字段（地址、电话、邮箱）先脱敏再进模型。
- AI 产生的写操作走统一 `POST /commands`，带 `dry_run` 预览、审批单与回读复核；与人工操作同一套审计。

**技术栈落点**

- 客户端：Tauri 2 + React/TypeScript；Rust 负责 SQLCipher、打印、文件、系统密钥库、签名更新；本地 IPC 给 MV3 侧栏。
- Control Plane：TypeScript（NestJS/Fastify）+ PostgreSQL + Redis + S3 兼容存储；同步 worker 独立部署，按租户分片。
- 身份：OIDC/SSO；设备会话与权限服务独立；审计日志只追加。
- 可观测：同步延迟、cursor lag、命令成功率、409 冲突率、死信数、每租户 token 刷新失败率。

#### 4.11.4 结论

智赢已经验证了“服务端主库 + 账号登录即恢复”的骨架，这条路可行；它的问题不在方向，而在**同步是手工时间窗、没有游标/变更日志/幂等/设备管理**。本 ERP 保留骨架，替换同步内核：

- 保留：服务端事实源、账号绑定店铺、按业务员/租户分页、记录级时间戳、推送 + 手工兜底、流式进度、本地文件缓存。
- 替换：手工时间窗 → 快照 + 游标 + 变更日志 + SSE + 每日对账；客户端持 token → Token Vault + 设备会话；无幂等 → 命令日志 + `Idempotency-Key`；无设备管理 → 设备注册/吊销/远程清缓存。
- 结果：换机只登录 ERP 账号即可恢复业务视图；Meli 授权与 AI 密钥不落客户端；离线可读可草稿，联网后安全重放。

---

## 5. 验收指标

| 类别 | 指标 | 目标 |
|---|---|---|
| 授权 | OAuth 成功率 | ≥99%（排除用户取消） |
| Token | 刷新失败率 | <0.5%，失败可自动重新授权 |
| 同步 | 订单/商品同步成功率 | ≥99%，失败可重放 |
| 同步 | 断点续传恢复率 | ≥99.9%，游标不丢不重 |
| 同步 | 变更游标延迟（P95） | <60 秒 |
| 跨设备 | 登录后近 7 天数据可见时延（P95） | <60 秒；全量历史后台补齐 |
| 设备 | 吊销生效时延 | <60 秒 |
| 冲突 | 409 冲突留痕与可合并率 | 100% |
| 时效 | 关键数据延迟 | <5 分钟（按业务调整） |
| 稳定 | 429 比例/重试成功率 | 可观测、可控，无雪崩 |
| 安全 | 高危漏洞/越权 | 0 未修复 |
| 合规 | DPA 覆盖率 | 100% 签约卖家 |
| AI | 高风险操作人工审批率 | 100% |
| AI | 建议采纳率/人工接管率 | 有基线并持续跟踪 |
| 成本 | 单租户 AI 成本 | 有预算、告警、熔断 |
| 促销 | 批量退出成功率 | ≥99%（排除平台锁定/不支持） |
| 促销 | 退出后回读复核率 | 100% |
| 促销 | 误关闭率 | 0 |

---

## 6. 待核实

- 官方 Promotions API 对各国、各活动类型（DEAL、PRICE_DISCOUNT 及其他）的覆盖范围与 `promotion_type` 取值。
- 官方是否提供订单/消息/索赔的 Webhook/通知机制、各站点权限差异与事件覆盖；智赢的推送依赖其应用授权（`acc_order` + `mpush`），自有应用能收到哪些事件、是否需要逐账号配置回调，需实测。
- Meli API 的具体限流数值：官方合同只写“可自行决定”，第三方“1500 次/分钟”说法未获官方确认。
- ML/MP 应用拆分的最新执行细则与过渡期。
- 各国发票/税务/数据驻留的本地化要求。
- AI 训练/微调在阿根廷法及目标国数据保护法下的合法性，需法务专项意见。
- 智赢扩展后续版本是否更换选择器/接口；仅作参考，不作为依赖。
- “批量关闭折扣”在平台锁定、活动类型不支持时的官方 API 替代方案覆盖度。

---

## 7. Sources

### 官方授权与用户

- [Authentication and Authorization](https://developers.mercadolibre.com.ar/en_us/authentication-and-authorization)
- [Authorization and Token Recommendations](https://developers.mercadolibre.com.ar/en_us/authorization-and-token-recommendations)
- [Identity and Access Management – OAuth and Tokens](https://developers.mercadolivre.com.br/en_us/search-products-seller/identity-and-access-management-oauth-and-tokens)
- [Start testing（无 sandbox 说明）](https://developers.mercadolibre.com.ar/en_us/start-testing)
- [Functional Permissions](https://developers.mercadolivre.com.br/pt_br/api-docs-pt-br/permissoes-funcionais)
- [Manage your applications](https://developers.mercadolibre.com.ar/en_us/php/manage-your-applications)

### 官方商品与促销

- [Products Sync & Listings](https://developers.mercadolibre.com.ar/en_us/products-sync-listings)
- [User Products（CBT）](https://developers.mercadolibre.com.ar/en_us/user-products)
- [Traditional Campaigns（含 DELETE /seller-promotions/items）](https://developers.mercadolibre.com.sv/en_us/products-sync-listings/traditional-campaigns)
- [Price Discount（promotion_type=PRICE_DISCOUNT）](https://developers.mercadolibre.com.bo/en_us/manage-feedback/price-discountt)

### 官方订单、消息与索赔

- [Packs Management](https://developers.mercadolibre.com.ar/en_us/news/packs-management)
- [Pending Messages](https://developers.mercadolibre.com.uy/en_us/net/pending-messages)
- [Gestionar resolución de reclamos](https://developers.mercadolibre.com.bo/es_ar/gestiona-ventas/gestionar-resolucion-de-reclamos)

### 官方合同与 DPP

- [Términos y Condiciones（西语官方合同）](https://developers.mercadolibre.com.ar/es-ar-terminos-y-condiciones)
- [Developer Partner Program（官方）](https://developers.mercadolibre.com.bo/en_us/ship-products/developer-partner-program)
- [Centro de Partners México](https://centrodepartners.mercadolibre.com.mx/)
- [Centro de Partners Brasil](https://centrodepartners.mercadolivre.com.br/)
- [Bloqueo de Aplicaciones](https://developers.mercadolibre.com.co/es_ar/saldo-de-la-cuenta/bloqueo-de-aplicaciones)

### 本地证据（智赢逆向）

- [ZYing 逆向分析报告](../../research/zying/ZYing-美客多逆向分析报告.md)
- [MeliApi.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs)：OAuth、用户、刊登、订单发票、消息、索赔
- [MeliApiBuilder.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/MeliApiBuilder.resolved.cs)：从账号记录构造直连客户端
- [MeliAppInfo.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/MeliAppInfo.resolved.cs)：多应用配置（值已脱敏）
- [Meli授权详情.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs)：授权、字段保存、刷新
- [Meli_V2.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Detail.Meli_V2.resolved.cs)：刊登字段级操作
- [ZOrderMeli.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZOrderMeli.resolved.cs)：订单增量拉取
- [ZSyncMeli.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Auto.ZSyncMeli.resolved.cs)：商品同步
- [Logic.MeliLabelPrint.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/Logic.MeliLabelPrint.resolved.cs)：面单解析与打印
- [Meli消息列表.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs)：售后消息列表
- [Meli索赔列表.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli索赔列表.resolved.cs)：索赔列表
- [Meli本土Qa.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Channel.Meli本土Qa.resolved.cs)：问答列表
- [MeliMsgContainer.resolved.cs](file:///E:/美客多逆向/analysis/decrypted/user/ZYing.Dialog.MeliMsgContainer.resolved.cs)：消息发送与附件
- [ZYing.Channel.优惠活动.cs](file:///E:/美客多逆向/analysis/decompiled/user_plugin/ZYing.Channel.优惠活动.cs)：桌面端优惠活动（Lazada 专用）
- [ZYing.Dialog.D一键报名.cs](file:///E:/美客多逆向/analysis/decompiled/user_plugin/ZYing.Dialog.D一键报名.cs)：批量报名（非关闭）
- [扩展 manifest.json](file:///E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0/manifest.json)：MV3、权限、版本
- [promo_700k_780k.pretty.js](file:///E:/美客多逆向/analysis/ext_slices/promo_700k_780k.pretty.js)：批量关闭折扣的可读切片
