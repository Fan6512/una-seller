# 接口地图：客户端直连 Meli 与 ZClient 转发的分界

证据文件：`analysis/decrypted/MeliApi.resolved.cs`（直连层）、`analysis/decrypted/ZYing_ZClient.resolved.cs`（转发层）、
`analysis/decrypted/ZYing_Meli_MeliErrorParser.resolved.cs`（错误与重试）、`analysis/command-inventory.csv`（命令站点）。
机器提取的端点清单见 `analysis/meli-direct-endpoints.json`。

## 三条出口

| 出口 | 形态 | 携带什么 | 谁负责业务规则 |
|---|---|---|---|
| 直连官方 REST | `https://api.mercadolibre.com` + 路径，`Authorization: Bearer <token>` | 单商品/单订单粒度的读写 | 美客多 |
| ZClient 命令转发 | 智赢自有二进制协议 + `cmd=<命令名>` | 业务动作（刊登、同步、拆单、核价…）与批量编排 | 智赢服务器 |
| 智赢 HTTP 服务 | `https://{meli,seller,erp,amazon,ozon,oa,hepsi}.zying.net/api` | 页面数据、配置、日志上报、三方授权跳转 | 智赢服务器 |

判读：凡是"读一个对象的当前状态"（商品详情、类目属性、索赔进度、绩效）走直连；凡是"发起一个业务动作"或"要跨账号/跨站点聚合"走命令转发。这个分界本身就是自研 ERP 的接口分层参考答案：**读官方 API、写走自己的编排服务**。

## 直连端点（按资源域分组）

动词列取自调用点：显式写出的按原样记录；写 `GET（sender 默认）` 表示调用点没传动词参数，
由 `SendAsync` 的默认值决定；带 `†` 表示没有显式动词但携带了请求体（GET + body，官方文档多要求 POST，属客户端写法）。
路径列只含字面量段，ID 与站点码由插值拼接，见 `params`。

值得注意：索赔的 `open-dispute`、`expected-resolutions/partial-refund` 这类动作端点在调用点没有显式动词，由 sender 默认成 GET（后者还额外带了一个 JSON 请求体）。动作类端点按官方约定通常是 POST，这里没有做运行时验证，只记录客户端写法；自研不要把动词交给默认值，显式写出才不会在改 sender 时静默变形。

### 售后与索赔

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `ClaimAffectsReputation` | `/affects-reputation`<br>`/marketplace/v2/claims/` | GET（sender 默认） | — |
| `ClaimOpenDispute` | `/actions/open-dispute`<br>`/marketplace/v2/claims/` | GET（sender 默认） | `ClaimInfo` |
| `ClaimPartialRefund` | `/expected-resolutions/partial-refund`<br>`/marketplace/v2/claims/` | GET（sender 默认） † | `ClaimExpectedResolutions` |
| `ClaimRefund` | `/expected-resolutions/refund`<br>`/marketplace/v2/claims/` | Post | `ClaimExpectedResolutions` |
| `DownloadClaimAttachment` | `/attachments/`<br>`/download`<br>`/marketplace/v2/claims/` | Get | — |
| `GetClaim` | `/marketplace/v2/claims/` | GET（sender 默认） | `ClaimInfo` |
| `GetClaimAttachmentUrl` | `/attachments/`<br>`/download?access_token=`<br>`/marketplace/v2/claims/` | GET（sender 默认） | — |
| `GetClaimExpectedResolutions` | `/expected-resolutions`<br>`/marketplace/v2/claims/` | GET（sender 默认） | `ClaimExpectedResolutions` |
| `GetClaimMessage` | `/marketplace/v2/claims/`<br>`/messages` | GET（sender 默认） | — |
| `GetClaimPartialRefundAvailableOffers` | `/marketplace/v2/claims/`<br>`/partial-refund/available-offers` | GET（sender 默认） | `PartialRefundAcailableOffers` |
| `GetClaimReason` | `/marketplace/v2/claims/reasons/` | GET（sender 默认） | `ClaimReason` |
| `SearchClaims` | `/marketplace/v2/claims/search?` | GET（sender 默认） | `SearchClaimsResult` |
| `SendClaimMessage` | `/actions/send-message`<br>`/marketplace/v2/claims/` | Post | — |
| `UploadClaimAttachent` | `/attachments`<br>`/marketplace/v2/claims/` | Post | — |

### 图片与视频

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `GetClips` | `/clips`<br>`/marketplace/items/` | GET（sender 默认） | `MeliClipsInfo` |
| `GetPictureInfo` | `/pictures/` | GET（sender 默认） | `MeliPicture` |
| `UploadClips` | `/clips/upload`<br>`/marketplace/items/` | Post | `MeliClipsResponse` |
| `UploadPic` | `/pictures/items/upload` | Post | `MeliPicture` |

### 全球商品（UP）

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `CreateGlobalUP` | `/global/user-products` | Post | `CreateUpResult` |
| `CreateLocalUP` | `/global/user-products/` | Post | `CreateUpResult` |
| `GetFamilyUserProductIds` | `/sites/`<br>`/user-products-families/` | GET（sender 默认） | `FamilyUserProducts` |
| `GetGlobalUserProductId` | `/global`<br>`/user-products/` | GET（sender 默认） | — |
| `GetLocalUserProductIds` | `/locals`<br>`/user-products/` | GET（sender 默认） | — |
| `GetUserProduct` | `/user-products/` | GET（sender 默认） | `UserProduct` |
| `GetUserProductItemMapping` | `/mapping`<br>`/marketplace/user-products/` | GET（sender 默认） | — |
| `UpdateUPAttrs` | `/global/user-products/` | Put | — |
| `UpdateUPImages` | `/global/user-products/` | Put | — |
| `UpdateUPPrice` | `/global/user-products/` | Put | — |

### 类目与属性

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `CategoryPredictor` | `/marketplace/domain_discovery/search?q=` | GET（sender 默认） | — |
| `GetAllCategories` | `/categories/all`<br>`/sites/` | GET（sender 默认） | `MeliAllCategory` |
| `GetCateAttrs` | `/attributes`<br>`/categories/` | GET（sender 默认） | — |
| `GetCategory` | `/categories/` | GET（sender 默认） | `MeliCategory` |
| `GetRootCategories` | `/categories`<br>`/sites/` | GET（sender 默认） | — |

### 商品与刊登

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `CatalogListingEligibility` | `/catalog_listing_eligibility`<br>`/items/` | GET（sender 默认） | `CatalogEligibilityResponse` |
| `CatalogListingEligibility` | `/multiget/catalog_listing_eligibility?ids=` | GET（sender 默认） | — |
| `GetCatalogProduct` | `/products/` | GET（sender 默认） | `CatalogProduct` |
| `GetDescription` | `/description`<br>`/items/` | GET（sender 默认） | — |
| `GetItemIncludeAllAttrs` | `/items/` | GET（sender 默认） | `MeliItemInfo` |
| `GetItemInfo` | `/items/` | GET（sender 默认） | `MeliItemInfo` |
| `GetItemInfoBatch` | `/items?ids=` | GET（sender 默认） | — |
| `GetItemMapInfo` | `/items/`<br>`/marketplace_items` | GET（sender 默认） | `MarketItemMapInfo` |
| `GetItemVisit` | `/visits/items?ids=` | GET（sender 默认） | — |
| `GetItems` | `/items?ids=` | GET（sender 默认） | — |
| `GetMarketItemInfo` | `/marketplace/items/` | GET（sender 默认） | `MeliItemInfo` |
| `GetProductItems` | `/items`<br>`/products/` | GET（sender 默认） | `ProductItemsResult` |
| `GetWeight` | `/items/`<br>`/shipping_options/free` | GET（sender 默认） | — |
| `PriceToWin` | `/marketplace/items/`<br>`/price_to_win?siteId=` | GET（sender 默认） | `PriceToWinResult` |
| `PublishCatalog` | `/items/catalog_listings` | Post | `MeliItemInfo` |
| `SearchItemsBySeller` | `/search?`<br>`/sites/` | GET（sender 默认） | `SearchSiteItemsResult` |
| `SearchProducts` | `/marketplace/products/search?` | GET（sender 默认） | `SearchProductResponse` |
| `SerachCatalogs` | `/items/search?`<br>`/users/` | GET（sender 默认） | `SerachCatalogsResponse` |

### 站内信与问答

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `DownloadMessageAttach` | `/marketplace/messages/attachments/` | Get | — |
| `GetItemInfraction` | `/marketplace/moderations/infractions/` | GET（sender 默认） | `InfractionPageResult` |
| `GetMessage` | `/marketplace/messages/` | GET（sender 默认） | `PackMessages` |
| `GetMessageAttachmentUrl` | `/marketplace/messages/attachments/` | GET（sender 默认） | — |
| `GetPackMessages` | `/marketplace/messages/packs/` | GET（sender 默认） | `PackMessages` |
| `GetUnReadMessagees` | `/marketplace/messages/unread?` | GET（sender 默认） | `UnReadMessageResult` |
| `SendMessage` | `/marketplace/messages/packs/` | Post | `MeliMessage` |
| `UploadMessageAttach` | `/marketplace/messages/attachments?site_id=` | Post | — |

### 订单

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `GetOrderInvoice` | `/invoice`<br>`/marketplace/orders/` | Get | — |

### 绩效与额度

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `GetItemPerformance` | `/item/`<br>`/performance` | GET（sender 默认） | `ItemPerformance` |
| `GetListingsLimit` | `/marketplace/users/cap` | GET（sender 默认） | — |

### OAuth 与令牌

| 客户端方法 | 路径 | 动词 | 返回类型 |
|---|---|---|---|
| `GetAuthUrl` | `/authorization?response_type=code&client_id=` | GET（sender 默认） | — |
| `GetMarketSite` | `/marketplace/users/` | GET（sender 默认） | `MarketSiteInfo` |
| `GetMyInfo` | `/users/me` | GET（sender 默认） | `UserInfo` |
| `GetShopInfo` | `/users/` | GET（sender 默认） | `ShopInfo` |
| `GetToken` | `/oauth/token?grant_type=authorization_code&client_id=` | Post | `TokenInfo` |
| `RefreshToken` | `/oauth/token?grant_type=refresh_token&client_id=` | Post | `TokenInfo` |

## 令牌与授权

- 授权站点按销售国区分：`auth.mercadolibre.com.{ar,mx,co,br}` / `auth.mercadolibre.cl` 等，授权链接以 `/authorization?response_type=code&client_id=` 起始，说明客户端直接参与 OAuth 码流程。
- 令牌刷新在直连层完成（`RefreshToken(refreshToken, accessToken)`），换码与刷新两个动作都是对 `/oauth/token` 的 POST，只是 `grant_type` 不同；并且访问令牌里可解析出 AppId 段：客户端用它定位到具体的应用配置对象。**具体 AppId/密钥值不在本仓库记录。**
- 发送前统一走 `CheckToken()`，随后为请求挂 `BearerToken`；请求地址由 `BuildUrl()` 拼接。
- 存在三条 sender：`SendAsync`（普通 JSON 请求，51 个方法）、`SendCoreAsync`（自己拼 `HttpRequestMessage` 的低层通道，只出现在二进制进出两类：图片/视频/附件上传与附件、发票下载，7 个）、`SendVersion2Async`（7 个方法，**全部集中在全球商品 UP 一族**）。也就是说"v2 通道"不是按 API 代际划分，而是 UP 那一族资源单独走了一条发送路径。

## 错误处理与重试（可直接借鉴的部分）

- 统一的 `MeliErrorParser.TryThrow(statusCode, response)`：状态码不在 2xx **或**响应体里带 error 都算失败，即"HTTP 成功但业务失败"也会抛出。
- 归一化的错误原因标识：`too_many_requests`、`daily_reached`、`internal_error`、`validation_error`，以及绩效侧的 `historical_behaviour`、`restrictions_coliving`、`comuna_pending`、`phone_pending`。
- 可重试判定 `CanRetry()` 不只按状态码，还按**消息内容**放行，例如 "is not a cbt item"、"does not have any child for site"、"you dont have item in this site and logistic"、"Attribute id is not valid"、"Picture id …"。
  这类是官方的**最终一致性延迟**（目录/变体/物流关系还没建立），重试而不是报错才是正确行为——自研的重试策略必须区分"幂等可等待"与"真失败"，否则会把刊登队列卡死。

### 已观测到的退避实现

| 位置 | 触发 | 策略 | 备注 |
|---|---|---|---|
| 图片上传 | 状态码 429 / 500 / 502 | 递增退避（约 600ms × 第几次重试）后递归重试 | 把网关 5xx 与限流同等对待 |
| 全球商品读取 | 捕获到 429 状态的 HTTP 异常 | 固定等待约 500ms 后递归重试，默认上限 2 次 | 批量枚举 ID 时调用方把上限提到 5 |
| 家庭系列商品 ID 读取 | 同上 | 同上，默认上限 3 次 | 重试耗尽**返回空值而不抛异常** |
| 命令通道清理后 | 执行完清理类命令 | 固定等待 1 秒再重新取统计 | 客户端主动避让，避免连续写后读 |
| 翻译类循环 | 每条译文之间 | 约 100ms 一条 | 逐条节流，不并发打满 |
| 净收益估算任务 | 每条之间 | 约 10ms 让出 UI 线程 | 主要目的是界面不冻结 |

判读：智赢对限流的处理是**局部、各自实现**的——上传一套、读 UP 一套、命令通道又一套，并且重试耗尽后静默返回空值，错误不会上浮。自研应改成**集中式限流器**：按账号 × 端点维护令牌桶与 429 冷却窗口，重试次数与退避曲线统一配置，耗尽后进入可观测的失败队列（而不是返回 null），否则批量任务会"看起来跑完了但少了一半数据"。

### 服务器寻址方式

- 命令通道的服务器是**编译进客户端的服务器组**：按平台与用途各建一组（美客多主组、UP 组、本地服务组、Ozon、Wildberries 等），组内地址取自写死的 IP 字面量，另支持按传入地址临时建组；**没有看到从服务器下发端点的通道**（具体地址不在本仓库记录）。
- 含义：客户端不依赖 DNS 也能连上业务服务器，抗域名封锁；代价是换服务器必须发版。自研应把端点下发交给服务端配置接口，客户端只保留一个引导地址，同时保留"引导地址不可用时"的备用清单。

## ZClient 命令通道的事实

- 命令清单 237 个调用点、193 个不同命令名、38 个域前缀，命令名形如 `<域>.<动作>`；同一动作在跨境/UP/本土三套页面里靠**参数**区分而不是靠不同命令。
- 请求参数两种写法：键值表（字段名即参数名）与一个强类型参数对象（属性名即参数名），最终都编码成二进制体；响应统一是记录集（`Record`/`RecordSet`），带 `list`、`stat`、`MaxCount` 之类的约定键。
- 分页由服务器返回的记录集自带信息驱动，客户端不保存 SQL：整个 `decrypted` 目录里没有出现任何 `create table`/`select … from` 语句，**客户端侧没有本地关系库**。

## 浏览器扩展侧

- 扩展复用同一套 `CmdHandler?cmd=` 命令契约（`mpush.select`、`shipper.shippers`、`ali1688.product`、`syslog.insert`），因此它是"同一账号体系的第二前端"，不是独立系统。
- 扩展内联了美客多商品页 URL 前缀（`articulo.mercadolibre.*`）与 1688/淘宝货源页，用于读取页面上下文；这属于第三方页面自动化，**自研不照做**，改由官方 API 取同一份数据。

## 限制

- 以上均为静态观察：没有运行时抓包，未验证任何一条接口在真实账号下的实际响应。
- 直连端点表来自 `MeliApi` 一个类；`ZYing.Work.dll` 内可能还有别的直连实现（该模块受 .NET Reactor 保护，尚未还原）。
- 方法到路径的关联是**同函数窗口**内提取的，插值字符串拼接的动态段（如 `id`、站点码）不在路径里，需读 `params` 列理解。
