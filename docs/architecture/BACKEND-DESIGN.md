# Una Seller 后端设计（Go 单体）

> 2026-11-18 · 设计基线 · 面向现有 `backend/` 与 M0→M3 路线
> 关联：[Foundation](FOUNDATION.md) · [Delivery](DELIVERY.md) · [Domain](DOMAIN_AND_EXECUTION.md) · [ERP-MVP](ERP-MVP.md) · [M0 部署](DEPLOYMENT_M0.md) · [ADR-0006 身份边界](../decisions/0006-auth-identity-boundary.md) · [云端 Go API 部署](CLOUD_GO_API_DEPLOY.md)
> 本文是目标结构与落地顺序，**不是已完成功能清单**。
> 修订：2026-11-18 第三版；`origin/main` 已含采集接口、Supabase 登录与云端部署，据此重写「代码现状」。

## 代码现状（先读）

| 位置 | 实际状态 |
| --- | --- |
| `origin/main` `6b42689`（2026-11-18） | `/healthz`；`/api/v1/collections` 与明细接口（pgx v5、私有 `app` schema、`0001` 迁移、显式 `api migrate`）；Supabase Auth 校验（`/auth/v1/user`）已接入采集路由；`Field[T]`、`migrations.Apply` 均已在代码中。仅监听 `127.0.0.1:8787`。 |
| 云端部署 | `deploy/una-seller/` 已有 systemd/Nginx 配置与 `api.env.example`，**公网部署与云库连通仍未验收**，见[云端 Go API 部署](CLOUD_GO_API_DEPLOY.md)。 |
| 尚未落地 | 订单、刊登、AI 文案/图片、对象存储与异步任务（第 8–9 章）；浏览器插件配对令牌（6.2）目前只在设计中。 |

**新表与新接口按第 7 章里程碑推进，先补迁移再改代码；不预建空表，也不在本文约束之外引入 ORM、缓存或消息队列。**

## 0. 结论

保持**一个 Go 二进制 + 一个 PostgreSQL + 现有 Nginx**；身份使用托管 Supabase Auth，不新增自建认证服务。内部按业务域分包，每个业务模块按需保留三层：

```
HTTP 层（路由/中间件/DTO/错误映射）
        ↓
领域层（校验、状态流转、合并语义、事务边界）
        ↓
存储层（pgx + 手写 SQL，不抽象 Repository）
```

PostgreSQL 同时承担持久化和轻量任务队列；Mercado Libre、AI、对象存储通过窄接口接入；不引入微服务、ORM、CQRS、Redis、消息总线。身份沿用 **ADR-0006 的 Supabase Auth**：Go 只验证 token，不自建密码/会话表；浏览器插件使用独立配对令牌。未来团队化不是“加一张成员表”就完成，见 6.6 与 7.1。

## 1. 不可违背的既有约束

来自 FOUNDATION / DELIVERY / DOMAIN / ADR，设计时先满足这些：

- 单后端、单数据库、单进程；模块化单体，不拆服务。
- 桌面客户端和浏览器插件不直连 PostgreSQL；Mercado Libre Token 只在后端。
- 金额用精确 decimal，经 API 传字符串；缺失值必须是 null，不能填 0。
- 部分采集不得覆盖已有完整数据；来源快照不得被 AI 或编辑覆盖。
- 公网 API 必须 HTTPS + 有效登录鉴权；浏览器插件使用独立凭据。
- 数据表按里程碑建立，不预建订单、团队、多租户等空表。
- 新组件必须回答：解决什么已存在问题、有没有更简单方案、运维成本与删除路径。

## 2. 运行形态与部署拓扑

```
Windows/macOS 桌面端 ─┐
Chrome/Edge 插件 ─────┼─ HTTPS ─→ Cloudflare ─→ Nginx 443 ─→ Go API 127.0.0.1:8787 ─→ Supabase PostgreSQL (TLS)
Mercado Libre OAuth ──┘                              │
                                                     └─ /healthz 公开；其余路径按鉴权放行

Go API ─→ Supabase Auth（仅验证桌面 token）
       ├→ Mercado Libre 官方 API（M1-D）
       ├→ AI Provider（M1-B）
       └→ 对象存储（M1-C）
```

- Go 进程只监听 loopback，由 systemd 以非 root 用户托管；Nginx 负责 TLS 和路径转发。
- 一个二进制，子命令：`api`（默认）、`migrate`、`version`；任务量出现后增加 `worker` 子命令，复用同一份代码。用户由 Supabase Auth 管理，Go 不提供建管理员或重置密码子命令。
- 迁移**不在启动时自动执行**；部署步骤显式 `api migrate`，失败则停止发布。
- 构建静态二进制（`CGO_ENABLED=0`），产物带版本号；保留上一版二进制用于回滚。
- 数据库连接池保持 2–5 个连接，走 Supabase Session Pooler（5432）或 Direct connection；TLS 必须开启，不关闭证书校验。
- 现有 `we-mp-rss`、Nginx 默认站点不受影响；8787 不对公网开放。

## 3. 代码结构

```
backend/
  cmd/api/main.go              # 子命令解析、启动、优雅退出
  internal/
    config/config.go           # env → Config；生产缺关键项直接失败
    db/db.go                   # pgxpool 创建、Ping、WithTx
    httpapi/                   # 只做 HTTP，不写 SQL
      router.go
      middleware.go            # recover / requestID / log / CORS / bodyLimit / auth
      errors.go                # 统一错误结构与 code 映射
      extension_handlers.go    # 配对码、插件令牌
      collection_handlers.go
      draft_handlers.go        # M1-A 再建
      product_handlers.go      # M1-A 再建
      listing_handlers.go      # M1-D/E 再建
    auth/                      # 身份边界，不自建密码体系
      supabase.go              # 调 /auth/v1/user 验证桌面 token
      extension.go             # 配对码、插件令牌、scope 检查
      store.go                 # 插件凭据 SQL
    collections/               # 采集箱
      types.go                 # 领域类型 + 校验
      service.go               # 规则变复杂时再抽；简单读可直接 handler → store
      store.go                 # SQL + 事务
    drafts/                    # M1-A 再建
    products/                  # M1-A 再建
    listings/                  # M1-D/E 再建
    meli/                      # M1-D 起
    ai/                        # M1-B 起
    storage/                   # M1-C 起
    jobs/                      # M1-B/E 起
    audit/                     # M1-D 起
  migrations/
    0001_source_collections.sql
    0002_extension_credentials.sql
    ...
    migrate.go                 # 版本表 + advisory lock + 顺序执行
```

上面的树是**目标形态，不是一次性脚手架**。当前主线只有 `cmd/api`、`internal/httpapi`、`internal/collections`、`migrations`；PR #15 额外有 `internal/httpapi/auth.go` 与 `cmd/api/config.go`。`drafts`、`products`、`listings`、`meli`、`ai`、`storage`、`jobs`、`audit` 在对应里程碑开始时再建，禁止先建空包、空接口。PR #15 的 `httpapi/auth.go` 可以先不动；出现插件令牌/scope 后，再把公共身份逻辑抽到 `auth` 包，不为目录好看做一次性搬迁。

依赖方向固定：

- `httpapi → 领域 service → store`；`httpapi` 不 import `store` 的 SQL 细节，`store` 不返回 HTTP 状态码。
- 简单只读查询允许 handler 直接调 store；出现校验、状态流转、跨表事务或外部调用时再抽 service，不为每个 handler 凑三层。
- `meli`、`ai`、`storage` 不 import 业务领域包；领域层定义自己需要的窄接口，由 `httpapi`/`main` 注入实现。
- 一个领域一个 `Store`，不是一张表一个 Store；不建 `Repository[T]`、`BaseStore`、通用 DAO。
- DTO 和领域类型只在字段确实不同时分开；不为每张表复制一套 struct。
- 共享代码在第 3 处重复时再抽包；不提前建 `pkg/`、`contracts/`、`utils/`。

## 4. 请求处理链

固定顺序：`recover → requestID → accessLog → CORS → bodyLimit → auth → handler`。

- 路由用标准库 `http.ServeMux` 的方法+通配符模式（如 `POST /api/v1/collections`、`GET /api/v1/collections/{id}`），不引入 Web 框架。
- `http.Server` 设置 `ReadHeaderTimeout 5s`、`ReadTimeout 30s`、`WriteTimeout 60s`、`IdleTimeout 120s`、`MaxHeaderBytes 64KB`。
- 收到 SIGTERM 后停止接收新请求，等待进行中请求最多 20 秒，再关闭连接池。
- panic 在中间件恢复：返回 500 + `request_id`，堆栈只进日志，不返回客户端。
- 每个请求的 context 携带 `principal`（user id、凭据类型）和 `request_id`；下游日志与外部调用复用。
- 客户端断开时取消 context；数据库和外部 HTTP 都必须接收 context。

## 5. API 契约

以下为 `/api/v1` 的目标契约；当前 `origin/main` 的 collections 错误响应仍是 `{ "error": "..." }`，切换时同步桌面与插件，不做两套并行格式。

### 5.1 基本约定

- 统一前缀 `/api/v1`，只收发 JSON，UTF-8。
- 成功响应直接返回资源；列表当前用 `{ "items": [], "limit": 30, "offset": 0 }`，数据量增长或分页抖动明显时改 `{ "items": [], "next_cursor": null }`，不为分页提前维护两套参数。
- 时间统一 RFC3339 UTC；ID 统一 UUID 字符串；不暴露数据库序列。
- 金额、价格、成本一律 decimal 字符串，数据库 `numeric(20,4)`；禁止 JSON 浮点数。
- 币种 ISO-4217 三位大写；缺失为 null。
- 未知字段默认拒绝（`DisallowUnknownFields`），契约变更走版本化，不静默接受拼写错误。

### 5.2 错误结构

```json
{
  "error": {
    "code": "validation_failed",
    "message": "price must be a decimal string",
    "fields": [{ "field": "price", "code": "invalid_decimal" }],
    "request_id": "0193f0c2-..."
  }
}
```

`code` 固定集合：`invalid_request`、`validation_failed`、`unauthorized`、`forbidden`、`not_found`、`conflict`、`rate_limited`、`upstream_error`、`not_ready`、`internal`。

- 客户端按 `code`/`fields` 展示，不解析英文 `message`。
- 上游错误不把 Token、请求头、完整响应体透给客户端；保留可读的字段级原因。
- 所有 5xx 日志带 `request_id`，便于按 ID 查日志。

### 5.3 缺失与显式 null

沿用 `origin/main` 已有的 `Field[T]` 三态：**字段不出现 = 本次未采集/不修改**；`null` = 明确清空；有值 = 更新。该语义同时用于采集重传和后续 PATCH 编辑，禁止用 null 表示“没采到”。

### 5.4 幂等与并发

- 采集提交：`user + platform + site + normalized_url` 唯一键 + 事务内行锁；重复提交返回 200/201 与最终记录，不新增重复行。
- 外部写操作（刊登、改价、库存）：客户端传 `Idempotency-Key`，服务端在 `publication_operations.client_request_id` 上建唯一约束；同 key 不同请求体返回 409。**该键只防止重复受理，不保证平台恰好执行一次**，不能把它当成分布式事务。
- 可编辑实体（整理工作稿、自有商品）增加 `version` 整数；更新必须带当前 `version`，条件更新命中 0 行返回 409，避免两个窗口互相覆盖。
- 超时不等于失败：外部写操作先回读平台状态，再决定重试或标记 `needs_check`；无法可靠回查的超时必须进入人工核实，不能自动重发。

### 5.5 健康检查

- `GET /healthz`：进程存活，不查数据库，可供 Nginx/Cloudflare 探活。
- `GET /readyz`：数据库 Ping + 迁移版本检查；只允许本机或内网访问，不放公网。
- 不在健康响应里返回 DSN、版本库地址、环境变量等敏感信息。

### 5.6 文档与 SDK

端点少于约 10 个时，契约由 Go 类型 + 集成测试保证；超过后再手写 `openapi.yaml` 并生成 TypeScript client。不引入注解式 Swagger 生成器，不提前搭生成流水线。

## 6. 身份与权限

首期单用户。桌面登录沿用 **ADR-0006**：Supabase Auth 负责邮箱密码、令牌签发与续签；Go 不自建 `sessions`、`password_hash`，也不签发桌面 JWT。浏览器插件使用单独配对令牌。

### 6.1 桌面登录（Supabase Auth）

- 桌面直接调用 Supabase Auth HTTP 端点（password、refresh_token、logout），不引入 `@supabase/supabase-js`，不维护第二套 Auth 客户端。
- 桌面所有 `/api/v1/*` 业务请求带 `Authorization: Bearer <Supabase access_token>`；Go 调 `GET {SUPABASE_URL}/auth/v1/user` 验证，使用返回的 `user.id` 作为 `owner_user_id`。
- 拒绝把 publishable/anon key 当 token；401 返回 `WWW-Authenticate: Bearer`；身份提供方不可达返回 503，不降级放行；跨用户资源按 404 处理，不泄露记录存在性。
- 本期不做 JWKS/本地验签；实测 `/auth/v1/user` 延迟或限流成为瓶颈后，再单独评估签名模式、JWKS 轮换与有界缓存。
- 不在 Go 建 `users.email`、`password_hash`、`sessions`；`app.users` 只保留 Supabase user UUID 与创建时间，首次写入时 `INSERT ... ON CONFLICT DO NOTHING`。
- 用户创建、禁用、密码重置都在 Supabase Dashboard 完成；Go 不提供 `create-admin`/`reset-password` 子命令，也不建邮件服务。
- 公开注册：ADR-0006 记录本地联调阶段暂未关闭；公网开放前必须在 Supabase Dashboard 实际关闭并验证，不能只在文档里写“已关闭”。

### 6.2 插件配对与令牌

| 表 | 关键列 | 用途 |
| --- | --- | --- |
| `app.extension_pairing_codes` | code_hash、user_id、expires_at、used_at、attempts | 配对码，5 分钟、一次性 |
| `app.extension_tokens` | token_hash、user_id、name、scopes、expires_at、revoked_at、last_used_at | 插件凭据，90 天、可吊销 |

- 配对码 8 位 Base32，明文只在生成时返回一次；数据库存 SHA-256。插件令牌 32 字节随机数，Base64URL 编码，前缀 `una_ext_`，数据库只存 SHA-256。
- 流程：桌面已登录 → 确保 `app.users` 有该 Supabase UUID（`INSERT ... ON CONFLICT DO NOTHING`）→ `POST /api/v1/extension/pairing-codes` 生成短码 → 插件 `POST /api/v1/extension/pair` 提交短码 → 服务端校验未过期、未使用、尝试次数未超限 → 签发插件令牌。
- Scope 首版只开 `collections:read`、`collections:write`。每个路由显式检查，不靠“有令牌就能访问”；插件令牌不能调用店铺授权、刊登、账号管理或生成配对码。
- 配对入口独立限速：每 IP 10 次/分钟、全局 100 次/小时；同一配对码 5 次失败即作废。**随机错误短码不会匹配任何记录，记录上的 `attempts` 单独不够**，必须同时有 IP/全局限速。
- 插件令牌存 `chrome.storage.local`，只在 service worker 使用；内容脚本通过 `chrome.runtime.sendMessage` 转发，绝不把令牌注入页面 `window`。
- 首版至少提供吊销接口；插件令牌列表、设备命名与批量管理按需再加，不先做复杂设备管理。

### 6.3 鉴权边界

| 路由 | 桌面 Supabase token | 插件 token |
| --- | --- | --- |
| `GET /healthz` | 公开 | 公开 |
| `GET /api/v1/collections` | 允许 | `collections:read` |
| `POST /api/v1/collections` | 允许 | `collections:write` |
| `POST /api/v1/extension/pairing-codes` | 允许 | 拒绝 |
| `POST /api/v1/extension/pair` | 不需要（用配对码） | 拒绝 |
| `/api/v1/stores/*`、`/listings/*`、`/publish/*` | 允许 | 拒绝 |
| `/oauth/mercadolibre/callback` | 公开，但校验一次性 state | 拒绝 |

- 中间件先看令牌前缀：`una_ext_` 走插件令牌哈希查询；其余走 Supabase 验证。**绝不把插件令牌发给 Supabase**。
- 所有业务 SQL 的 `owner_user_id` 来自已验证的 principal，**绝不接受请求体里的 user id**。
- 未登录返回 401；凭据有效但无该操作权限返回 403；跨用户资源统一 404。
- 公网开放前必须完成：关闭 Supabase 公开注册、插件配对与 scope 检查、限速、CORS/CSP、HTTPS、备份与恢复验证；未完成前不得把 `/api/v1/collections` 经 Nginx 暴露。

### 6.4 会话存储

- ADR-0006 当前 MVP 接受桌面 `localStorage` 保存 access/refresh token，并明确它不是 Windows Credential Manager/DPAPI 意义上的凭据库；不存明文密码，不打印 Token。
- 正式向陌生用户分发或处理更高敏感数据前，再单独评估系统凭据库、设备退出与撤销策略；本期不引入 Rust keyring 插件。
- 插件令牌只存 `chrome.storage.local`，不进入页面上下文。

### 6.5 CORS

- Go API 精确允许：开发 `http://127.0.0.1:1420`、`http://localhost:1420`、`http://tauri.localhost`、`https://tauri.localhost`、`tauri://localhost`；生产 Tauri 自定义协议；插件 `chrome-extension://<固定扩展 ID>`。
- 回显允许的 Origin 并设置 `Vary: Origin`；不使用 `*`，不允许 credentials。
- 预检允许 `Authorization, Content-Type, Idempotency-Key, If-Match`。
- CORS 不是鉴权边界；插件跨域请求仍必须带令牌。
- 桌面直连 Supabase Auth 时，Tauri CSP 的 `connect-src` 需单独放行 Supabase Auth 域名；Go 的 CORS 不负责这段。

### 6.6 归属与团队边界

- `owner_user_id` 只解决当前单用户的归属与跨用户过滤，不等于团队模型。
- 团队共享还会涉及店铺访问授权、商品归属/转移、操作人记录、成员退出后的数据处理；出现真实需求时单独设计迁移，不预建组织、角色、成员表。
- 现在能做的是：每张业务表保留明确 owner、关键写操作记录 actor、权限检查集中在中间件/领域入口，避免以后散落改。

## 7. 数据模型演进

### 7.1 身份链

```
SourceProduct（采集来源快照）
   ├─ CollectionWorkingDraft（整理工作稿，可保存/恢复）
   └─ OwnProduct（自有商品草稿，一个来源可产生多个）
          ├─ OwnProductLocalization（en / es-MX / pt-BR ...）
          ├─ OwnProductVariant / OwnProductImage
          └─ PublicationOperation（一次店铺操作）
                    └─ PublicationTarget（站点目标）
                              └─ PublishedListing（官方确认后的站点商品）
```

规则：

- 五个对象各有独立身份，**不能合并成一张 `products` 表加状态列**。
- `SourceProduct` 只存来源事实；工作稿、AI 候选、自有商品都不回写它。
- 平台 item id / user product id 只出现在刊登与产品列表表，不放在 `own_products`。
- `PublishedListing` 也可以由官方只读同步产生，允许暂时没有 `OwnProduct` 或 `PublicationOperation`；不要为了外键完整而伪造本地记录。
- 采集状态、草稿准备度、发布状态、平台商品状态各用各的列/枚举，不共用一个 `status`。
- 每张业务表都带 `owner_user_id`；它只解决当前单用户的归属与越权过滤，**不等于团队共享**。团队还需要店铺访问、商品转移、操作人、成员退出处理，见 6.6。

### 7.2 按里程碑建表

| 阶段 | 新增表 | 说明 |
| --- | --- | --- |
| 当前 M0（`origin/main`） | `users`、`source_products`、`source_product_images`、`source_product_variants` | 已合并；`app` 私有 schema。本地 `e10b4b0` 尚无这些表 |
| 公网开放前 | `extension_pairing_codes`、`extension_tokens` | 插件凭据；桌面身份由 Supabase Auth 提供，不建 sessions/password 表 |
| M1-A | `collection_working_drafts`、`own_products`、`own_product_variants`、`own_product_images`、`own_product_localizations` | 工作稿每来源一份；自有商品可多份 |
| M1-B | `ai_runs`、`jobs` | 生成审计与用量；异步生成从这时开始用任务表 |
| M1-C | `image_assets`（及必要的 derivative 行） | 原图不可变；派生物单独记录 |
| M1-D | `store_connections`、`store_templates`、`oauth_states` | Token 加密存储；只读同步游标 |
| M1-E | `publication_operations`、`publication_targets`、`published_listings` | 请求级与站点级状态分开；复用已有 `jobs` |
| M2 | 复用 `publication_operations`，增加 `kind`；按需加同步状态 | 改价/库存/内容更新不另起系统 |
| M3 | `webhook_events`、订单/消息表 | 有真实通知需求再建 |

### 7.3 关键字段约束

- 金额 `numeric(20,4)`；币种 `char(3)` 或带 CHECK 的 text；未知为 NULL。
- 时间全部 `timestamptz`，默认 `now()`，应用层统一 UTC。
- 可编辑表加 `version bigint not null default 1`、`created_at`、`updated_at`。
- 工作稿正文用 `payload jsonb` + `payload_version`，服务端按版本做 schema 校验；需要检索/排序的字段才提列。
- `own_product_images` 存 `role`（main/additional/variant）、`sort_order`、`source_image_id`、`asset_id`（M1-C 起）、`selected_derivative_id`；不覆盖原图。
- `published_listings` 唯一键 `(store_connection_id, site_id, platform_item_id)`；`own_product_id` 与 `operation_id` 可空，官方只读同步的旧商品不要求本地草稿或刊登操作。只读同步只 upsert，不因某次未返回就删除本地记录，用 `last_synced_at` 标记。
- `publication_targets` 唯一键 `(operation_id, store_connection_id, site_id)`；当前产品模型是一次操作一个账号，因此再加约束保证 target 的店铺与 operation 的店铺一致。把店铺放进唯一键成本为零，可防止未来跨店铺批量时同一站点撞键；若真做跨店铺批次，则把 operation 定义为批次。保存平台 item id、permalink、字段级错误、`last_checked_at`。
- `jobs` 用 `status`（pending/processing/needs_check/done/failed/dead）+ `run_at` + `attempts` + `dedupe_key` + `lock_token` + `locked_until`；部分索引只扫 pending。`dedupe_key` 用部分唯一索引，只约束 pending/processing，任务完成后允许同 key 再入队。`lock_token` 是 fencing token，防止崩溃前的旧执行者覆盖新执行者结果。

### 7.4 删除与保留

- 采集记录删除：`source_product_images`、`source_product_variants` 级联；一旦存在工作稿或自有商品，改为 409 提示先处理关联，不级联删除下游。
- 自有商品删除：刊登记录保留（审计），`own_product_id` 置空或保留历史引用。
- 产品列表与刊登记录不随商品删除而消失；它们是平台事实与操作历史。
- `audit_events` 追加写，不更新不删除；按需做保留期清理。

## 8. 外部集成

### 8.1 Mercado Libre

一个 `meli.Client` 包，先覆盖认证、请求、错误转换、基本限流；功能变多再按商品/订单分文件，不建 Adapter + Gateway + Repository 三层。

- OAuth：`oauth_states` 表存 state 哈希 + user_id + 过期时间；回调校验一次性 state，用 code 换 Token，再调 `/users/me` 验证身份。
- Token 存储：AES-256-GCM 加密后入库，列带 `key_version`；密钥只在服务器 env，不进 Git、不进日志。数据库被备份或误导出时，明文 Token 不直接泄露。
- 刷新：用短事务抢占刷新租约（`refresh_lease_until` + 随机 `refresh_token_lock`），提交后再调用平台，刷新请求 15 秒超时；不要在长事务里等外部 HTTP。成功后单条 UPDATE 保存加密 Token；`invalid_grant` 标记 `reauth_required`，不无限重试。
- 刷新失败窗口：平台已轮换 Token、本地保存失败或网络超时，都不能当作“再刷一次就好”。保存失败或超时后标记 `refresh_unknown`，停止自动刷新，向前端返回可读错误；下次用户操作或定时只读校验确认状态，确认失效则要求重新授权。Token 是否轮换、旧 Refresh Token 是否立即失效，实施时按平台当期文档和真实响应核实，不硬编码固定小时数。
- 并发：同一连接只允许一个刷新执行者；其他请求等待租约释放后复用新 Token，不各自发起刷新。
- 重试：GET 遇 429/5xx 最多 2 次，指数退避 + jitter，尊重 `Retry-After`；POST 不盲目重试，超时先回读，无法确认则 `needs_check`。
- 错误：统一 `APIError{Status, Code, Message, Causes[]}`；把 `cause` 映射到编辑器字段，原始响应脱敏后可折叠展示。
- 限流：按连接做简单令牌桶；收到 429 时尊重 `Retry-After`，不把限流当成可立即重试的普通错误。
- 只调用官方 REST API；不做 DOM 自动化、不引用第三方私有协议。

### 8.2 AI

- `ai.Generator` 窄接口：输入商品事实 + 目标语言 + 任务类型，输出结构化候选；实现可为 OpenAI 兼容端点或 UnaAPI 网关。
- 超时 60 秒；生成耗时超过请求预算时改为 `jobs` 异步任务，前端轮询结果，不用长连接硬撑。
- 输入只包含必要商品事实，不含平台 Token、买家信息、数据库内容；来源描述视为不可信文本，提示词明确分隔，禁止模型调用工具或写平台。
- 输出按 JSON schema 校验；候选单独保存，人工按字段采纳；AI 不直接写 `SourceProduct`，也不直接刊登。
- `ai_runs` 记录 provider、model、prompt_version、input_hash、用量、状态；按日预算上限，超限返回可读错误而不是静默降级。
- 图片白底是独立能力（M1-C），与文案生成不共用“万能 AI 服务”；官方是否开放白底 API 未确认前不依赖。

### 8.3 图片与对象存储

- M0/M1-A 只保存来源图片 URL，桌面端直接加载；CSP 的 `img-src` 按实际 CDN 白名单配置。
- M1-C 首次需要自有图片时引入对象存储（优先 Cloudflare R2 等 S3 兼容服务；若考虑 Supabase Storage，先更新 ADR-0006 的 Storage 边界），定义 3 个方法的 `storage.Store`：`Put`、`Delete`、`SignedURL`。
- 数据库只存 `storage_key`、`sha256`、`mime`、`width/height`、`bytes`；不把二进制塞进 PostgreSQL。
- 原图不可变；白底/裁剪结果是新 asset 或 derivative，记录 provider 与创建时间；人工选用后写 `selected_derivative_id`。
- 后端如需下载来源图片，限制协议为 HTTPS、域名白名单、重定向次数、响应大小，并阻止解析到内网地址，避免 SSRF。

## 9. 异步任务

触发条件（满足其一才引入）：

- 外部写操作需要重试/回读（刊登、改价、库存）。
- 单次处理超过约 10 秒（AI 文案、白底、批量同步）。
- 需要断点恢复或定时执行（店铺只读同步）。

设计：

- `jobs` 表列：`id`、`kind`、`payload jsonb`、`status`、`run_at`、`check_at`、`attempts`、`max_attempts`、`dedupe_key`、`lock_token uuid`、`locked_until timestamptz`、`last_error`、`created_at`、`updated_at`。状态：`pending/processing/needs_check/done/failed/dead`。
- 领取用短事务：`UPDATE ... SET status='processing', lock_token=$2, locked_until=now()+lease, attempts=attempts+1 WHERE id = (SELECT id FROM jobs WHERE status='pending' AND run_at<=now() ORDER BY run_at FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING ...`，立即提交，再执行外部调用。**不要在整个任务执行期间持有数据库事务。**
- 完成时条件更新：`WHERE id=$1 AND lock_token=$2`。命中 0 行说明租约已被别人接管，旧执行者必须丢弃结果，不能覆盖新执行者。
- 进程崩溃后租约到期，另一个 worker 可重新领取；长任务每 30 秒心跳延长 `locked_until`，租约默认 2 分钟。心跳与完成都必须带 `WHERE id=$1 AND lock_token=$2` 条件；心跳失败立即停止执行，避免两个执行者同时写。`lock_token` 是 fencing token，解决“旧执行者晚到写入”的竞态。
- 外部写入结果不明确（超时、连接中断、平台 5xx 无法判断是否已执行）时，任务转 `needs_check` 并设 `check_at`；由回读任务确认成功/失败，**不能直接改回 pending 重发**。
- 每个 handler 必须可重入：先查业务状态与操作记录，再决定是否执行；外部写操作先回读。
- 失败退避：1m、5m、30m、2h；超过 `max_attempts` 进 `dead`，保留错误供人工处理。
- 单进程内 runner 默认开启（`WORKER_ENABLED=true`）；需要隔离时用同一二进制的 `worker` 子命令单独跑，代码不分叉。
- 轮询 1–2 秒；单用户规模足够。出现真实积压/延迟证据再评估 River/Redis，不在现在安装。
- 优雅退出：停止领取新任务，等待当前任务最多 30 秒；仍未完成则清空自己的 `lock_token` 并保留 `processing`，让租约自然过期后被重新领取，避免两个执行者同时写。

## 10. 配置、日志与运维

### 10.1 配置

`internal/config` 从环境变量读取，生产缺关键项直接启动失败：

| 变量 | 说明 |
| --- | --- |
| `APP_ENV` | `dev` / `prod` |
| `ADDR` | 默认 `127.0.0.1:8787` |
| `DATABASE_URL` | 必填；TLS 连接 |
| `PUBLIC_BASE_URL` | OAuth 回调与对外链接使用 |
| `CORS_ORIGINS` | 逗号分隔的精确来源 |
| `SUPABASE_URL` | 项目 URL；Go 调 `/auth/v1/user` 验证桌面 token |
| `SUPABASE_PUBLISHABLE_KEY` | Supabase publishable key；不是用户 token，也不是 service role key |
| `TOKEN_ENC_KEY` | 32 字节 Base64；加密平台 Token |
| `MELI_CLIENT_ID` / `MELI_CLIENT_SECRET` / `MELI_REDIRECT_URI` | M1-D 起必填 |
| `AI_BASE_URL` / `AI_API_KEY` / `AI_MODEL` | M1-B 起必填 |
| `WORKER_ENABLED` | 默认 false，M1-B 起按需开启 |
| `JOB_LEASE` | 默认 2m；任务租约时长 |
| `LOG_LEVEL` | `info` / `debug` |

- 本地 `.env` 只用于开发且被 gitignore；服务器用 systemd `EnvironmentFile`，权限 600。
- 启动日志只打印 `APP_ENV`、监听地址、是否连接数据库，不打印 DSN、密钥、Token。
- 配置结构体显式字段，不用 `map[string]string` 到处传。

### 10.2 日志

- `log/slog` JSON 输出到 stdout，由 journald 收集。
- 请求日志字段：`request_id`、`method`、`path`、`status`、`duration_ms`、`user_id`、`error_code`。
- 默认不记录请求体、商品标题/描述、Token、Cookie、Authorization；排查时用 ID 关联数据库。
- 外部调用日志记录目标域名、状态码、耗时、重试次数，不记录完整响应体。

### 10.3 迁移

- `migrations/` 下按序号命名，只增不改；执行记录写入 `app.schema_migrations`（version、checksum、applied_at）。
- 执行迁移命令时获取 `pg_advisory_lock`，防止并发执行；每个文件在事务内执行，失败回滚并停止发布。API 启动本身不自动迁移。
- 采用 expand/contract：先加可空列/新表 → 发布兼容代码 → 回填 → 再加约束；保证上一版二进制在迁移后仍能运行，便于回滚。
- 生产迁移前先备份；破坏性变更（删列/改类型）拆成两次发布。

### 10.4 部署与回滚

1. **验证产物**：CI 构建 `api` 并输出版本号与 SHA-256；服务器下载后先校验 checksum，再执行 `api version`，确认不是截断/错架构的二进制。
2. **备份**：备份当前 release 的二进制、systemd 配置与数据库；记录当前版本号，作为回滚点。
3. **验证迁移兼容性**：在 staging 副本或维护窗口内先跑 `api migrate`，用**上一版二进制**对迁移后的库做启动和读取冒烟；确认 expand/contract 成立。不能只靠约定，也不能在生产直接试。
4. **放置新版本**：上传到 `/opt/una-seller/releases/<version>/api`，设置属主与 `0755`，不要直接覆盖正在运行的 `current/api`。
5. **执行迁移**：运行 `releases/<version>/api migrate`；失败则停止，不动当前进程。
6. **原子切换**：创建临时符号链接指向新 release，再用 `mv -T` 原子替换 `current`，然后 `systemctl restart una-seller-api`；不要用“先删后传”的覆盖方式。
7. **就绪检查**：检查 `/healthz`、`/readyz`、一次登录后读取采集列表、一次写入；观察 5 分钟日志与错误率。全部通过才算发布完成。
8. **回滚**：把 `current` 切回上一版本并重启；因迁移已按 expand/contract 验证，旧二进制应能继续读取。若迁移是破坏性的，只能用第 2 步备份恢复，不能靠切二进制。
9. Nginx 只新增独立 server block；每次 `nginx -t` 通过后再 reload，不覆盖现有站点与 `we-mp-rss`。
10. 保留最近 3 个 release 目录；确认新版本稳定后再清理旧产物，不在发布当时删。

### 10.5 备份与恢复

- Supabase 自带备份能力需实际核实（免费计划可能暂停/容量受限），不能假定等于可恢复。
- 增加每日 `pg_dump` 到异地对象存储，保留 7–30 天；至少每季度做一次恢复演练。
- 明确 RPO/RTO 并记录；备份文件加密，访问凭据与数据库凭据分开管理。

### 10.6 观测

单用户阶段不引入 Prometheus/Grafana。`/healthz`、`/readyz`、结构化日志、数据库慢查询日志足够。出现真实性能问题再加指标，不提前部署监控栈。

## 11. 测试策略

- 单元测试：URL 规范化、金额/币种校验、三态合并、状态流转、Token 加密、错误映射、限流、Supabase verifier 的 401/503、插件 scope 与配对限速。
- 集成测试：真实 PostgreSQL（CI 临时库）验证迁移从零、升级、事务回滚、并发去重、部分采集不覆盖、`version` 冲突、任务租约过期后的接管与旧执行者写入被拒绝。
- 外部服务：`httptest.Server` 模拟 Meli 的 401 刷新、429、5xx、超时、`site_items[]` 混合结果，以及 Supabase `/auth/v1/user` 的成功/失效/不可达；不访问真实店铺。
- 契约测试：关键 DTO 用 golden JSON 固定字段名与 null 语义；OpenAPI 出现后再加 schema 校验。
- 不 mock pgx；数据库行为用真实 PG 验证。单元测试可 mock 窄接口（`meli`、`ai`、`storage`）。
- CI 运行 `gofmt`、`go vet`、`go test ./...`；集成测试必须显式设置 `TEST_DATABASE_URL`，缺失时失败而不是静默跳过。

## 12. 分阶段落地

| 阶段 | 后端工作 | 完成标志 |
| --- | --- | --- |
| 当前 M0（同步后） | collections 的 PG 迁移、幂等写入、读回；把 PR #15 的 Supabase Auth 验证并入主线 | 部分采集不覆盖旧数据；重启后读回；未登录 401，跨用户 404 |
| M0 PR-C | 采集插件 → API → 桌面闭环；Amazon.com 单商品 | 入库、去重、桌面可见 |
| 公网开放前 | 插件配对/令牌/scope/限速；Supabase Auth 跨账号隔离验证；统一错误与 request id；迁移 runner v2；CORS/CSP；HTTPS；关闭 Supabase 公开注册；备份恢复演练 | 公网只开必要路径；插件只能提交采集；未授权写全部拒绝 |
| M1-A | 工作稿、自有商品、变体/图片/多语言表；`version` 并发控制；删除关联规则 | 关闭重启后工作稿可恢复；来源快照不被覆盖 |
| M1-B | AI 适配器、`ai_runs`、异步生成任务与轮询 | 候选可按字段采纳；失败可见；不自动刊登 |
| M1-C | 对象存储、`image_assets`、白底任务、签名 URL | 原图/候选/选用可追溯；不覆盖原图 |
| M1-D | OAuth、Token 加密、店铺/模板、只读同步、`published_listings` | 旧商品无需自有商品也能出现在产品列表 |
| M1-E | 刊登操作/目标、逐站点结果、回读、幂等键 | MLB 成功/MLM 失败等混合结果准确；成功目标不重复发布 |
| M2 | 改价/库存/内容更新复用操作表；批量任务；同步调度 | 每次操作可说明是否生效 |
| M3 | Webhook 入库、订单/消息表、worker 扩展 | 通知可重放、可追踪；不丢事件 |

## 13. 明确不做

- 微服务、API Gateway、服务网格、独立 AI 服务、独立身份服务。
- GORM/ent 等 ORM、通用 Repository 接口、CQRS、事件溯源、领域事件总线。
- Redis 做缓存/队列/分布式锁；当前规模 PostgreSQL 足够（认证由 Supabase Auth 管理，不需要 Redis 会话）。
- 一张 `products` 表用可空字段承载采集、草稿、刊登三种身份。
- 用 JSONB 替代所有关系列；只对真正动态的属性/工作稿正文使用 JSONB。
- 启动时自动迁移、无版本记录的 `IF NOT EXISTS` 堆叠。
- 金额浮点、时间本地字符串、缺失值填 0。
- 把 Mercado Libre Token 放进桌面端、插件、日志或 AI 提示词。
- 另建 `users.password_hash`、`sessions` 或独立认证服务；身份按 ADR-0006 用 Supabase Auth，Go 只验证 token。
- 把桌面 `localStorage` 说成系统凭据库；它只是 ADR-0006 记录的 MVP 取舍。
- 在插件凭据、scope 检查、限速、Supabase 公开注册限制未完成前，把采集 API 暴露到公网。
- 为未来团队/多租户预建组织、角色、权限表。

## 14. 待确认的三个实现选择

身份不是开放项：ADR-0006 已选 Supabase Auth，Go 只做 token 验证；若未来要换，需单独 ADR，不能并行维护两套认证。

1. **迁移 runner**：建议在 `origin/main` 的 `migrations.Apply` 上升级为版本表 + advisory lock + 顺序执行（约 80 行，无新依赖）；若团队更倾向成熟工具，再评估 goose，但只选一种。
2. **平台 Token 加密**：建议应用层 AES-256-GCM（密钥在 env，列带 `key_version`），不单纯依赖 Supabase 静态加密与备份权限。
3. **对象存储**：M1-C 开工前用真实图片量、出口流量和价格对比 Cloudflare R2 与其他 S3 兼容服务；Supabase Storage 需先更新 ADR-0006 的 Storage 边界。在此之前不建存储表和上传接口。
