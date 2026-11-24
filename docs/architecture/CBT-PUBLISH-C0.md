# CBT-PUBLISH-001 · C0 契约确认单

**审批状态** · 公共 API 文档于 2026-11-23 核对。当前主账号、获批测试账号及账号专属字段均为 `UNVERIFIED` / `MANUAL_PENDING(CBT_TEST_ACCOUNT)`。真实发布、图片上传和 Windows GUI 验收分别等待 `MANUAL_PENDING(CBT_PUBLISH_AUTHORIZATION)`、`MANUAL_PENDING(CBT_IMAGE_UPLOAD_AUTHORIZATION)` 和 `MANUAL_PENDING(WINDOWS_GUI)`。

| 核对项 | 官方公开契约 | 实际账号待确认 |
| --- | --- | --- |
| 主账号、标签与子卖家 | `GET /users/me` 读取 CBT 主账号及全局 `status.sell.allow`，`GET /users/{merchant_id}` 查 `tags[]`。`GET /marketplace/users/{merchant_id}` 读取子卖家的 `user_id`、`site_id`、`logistic_type`，响应也可包含 `business_model`、`pricing_model`、`user_product`。子卖家映射提供配置线索，站点刊登权限须单独核实。 [用户映射] [User Products] [全托管] [卖家权限] | `UNVERIFIED` · 发布模型、启用标签、站点销售权限、Seller ID、物流及业务模式。 |
| Standard UP 创建 | 确认 `user_product_seller` 已启用且账号适用 Standard UP 后，单个 Siteless UP 用 `POST /global/items`，新 Family 批量用 `POST /global/user-products/families`。每个变体是独立 UP，同 Family 共享 Family ID。发送官方 `pictures[].id`，请求省略传统 `variations[]` 和 `title`。已有 UP 增加站点时复用其 ID，公开路径为 `POST /global/user-products/{siteless_id}`。 [User Products] [变体与家族] | `UNVERIFIED` · 账号适用性、变体和增站权限。 |
| Legacy 创建 | 仅在确认账号仍有 Legacy 资格时使用。初次创建用 `POST /global/items`，添加站点用 `POST /global/items/{CBT_ITEM_ID}`。官方正逐步停用 Legacy，缺少 UP 标签时仍须确认旧模型资格。 [Legacy] | `UNVERIFIED` · 旧模型资格及创建权限。 |
| Fully Managed 创建 | 任一子卖家返回 `business_model = "CBT CN Fulfillment Managed"` 时按全托管指南选路。`user_product_seller`、`warehouse_management` 可辅助判断。指南以 `POST /global/items` 创建 Siteless UP，发送 `family_name`、CBT 类目、官方图片 ID、必需属性及根级 `global_net_proceeds`。平台分配站点、曝光、买家价格、币种和库存。请求省略 `sites_to_sell`、`price`、`listing_type_id`、`currency_id`、`available_quantity`。 [全托管] [卖家权限] | `UNVERIFIED` · 当前账号是否全托管、平台实际分配的站点及可用的 Family 创建方式。 |
| 商品成色 | Standard UP 以 `attributes[].id = ITEM_CONDITION` 表达成色。当前概览只列 `New`，`values[].id` 为 `2230284`。Legacy 指南也以该属性替代根级 `condition`。全托管指南同时要求根级 `condition` 和 `ITEM_CONDITION`，列出的根级值为 `new`、`used`，可核对的属性示例是 `New`。 [User Products] [Legacy] [全托管] | `UNVERIFIED` · 实际账号及类目的允许值，尤其全托管 `used` 是否受支持及所需属性值。 |
| 回读与结果 | UP 用 `GET /user-products/{siteless_id}`、`GET /marketplace/user-products/{siteless_id}/mapping`、`GET /sites/{site_id}/user-products-families/{family_id}` 和 `GET /marketplace/items/{item_id}`。Legacy 可用 `GET /items/{CBT_ITEM_ID}/marketplace_items`。按实际子卖家、站点和物流核对各结果。全托管创建示例仅返回主体和 Family ID，站点结果需回读。 [变体与家族] [全托管] [刊登查询] | `UNVERIFIED` · 现有测试商品的回读权限、无返回 ID 时的稳定关联及可见延迟。 |
| 标题与描述 | Standard UP 以根级 `family_name` 为家族名称，由平台生成站点标题。创建示例以根级 `description.plain_text` 提交纯文本描述。同 Family 使用同一基础名称，已产生销售时受改名限制。概览描述了 `sites_to_sell` 创建时的站点覆盖能力，更新示例使用 `listing_sites[].listing_id` 定位目标，并展示站点级 `family_name`、`description.plain_text`。Legacy 指南使用英文 `title` 和纯文本描述。全托管指南明确列出根级 `family_name` 及 Family 更新的 `common_data.description`。C 只读取产品箱已采纳保存的三语 `listing_fields.titles` 与 `listing_fields.descriptions`（正式待刊登素材）；顶层旧版 `description` 只能在用户明确审核同意后作为后备，不静默覆盖正式译文。卖点 `listing_fields.bullet_points` 保持独立，不自动拼入描述。并读取官方类目 ID、官方属性、图片、变体与站点销售条件。中文类目名仅供阅读，提交使用官方 ID 与官方属性。 [User Products] [变体与家族] [Legacy] [全托管] | `UNVERIFIED` · 当前账号新增时可接受的站点覆盖字段、标题与描述语言及长度、全托管创建描述支持。 |
| 价格、目标站点、销售条件 | Standard UP 创建以 `sites_to_sell[]` 指定 `site_id` 和 `logistic_type`。指南按计价模型使用 `price` 或 `global_net_proceeds`，创建示例展示根级 `global_net_proceeds`，站点更新示例展示 `listing_sites[].net_proceeds`。Legacy 指南列出 USD 的根级及站点 `price`，启用净回款时用 `sites_to_sell[].net_proceeds`。Fully Managed 创建用根级 `global_net_proceeds`，站点价格、币种和库存由平台决定。UP 库存归 Siteless 主体，站点回读独立记录。 [User Products] [变体与家族] [Legacy] [全托管] | `UNVERIFIED` · 实际账号创建时可写字段的层级、各站点币种及精度、库存权限及同站同物流的独立销售条件数量。 |
| 图片 | 官方上传是 `POST /pictures/items/upload`，用 multipart 字段 `file` 取得官方图片 ID。UP 创建用 `pictures[].id`。格式与大小按官方规则及目标类目确认。 [图片] [User Products] | `UNVERIFIED` · 获批账号上传权限、处理状态、图片 ID 的复用范围。 |
| 预检与错误 | `GET /categories/{id}`、`GET /categories/{id}/attributes`、`GET /marketplace/users/cap` 可检查类目、属性和站点额度。全托管文档把 `POST /global/items/validate` 标为仅 staging。发布验证错误含 `error`、`status`、`cause[]`，各 cause 可带 `cause_id/type/code/references/message`。公开示例显示顶层 200 可同时含成功和失败的 `site_items[]`。 [验证] [Legacy] [全托管] [站点结果] | `UNVERIFIED` · 获批测试环境中可用的无创建副作用预检端点、匹配当前账号的错误样本。 |

## C0 模式识别与店铺映射边界（2026-11-24）

探针现会在脱敏 JSON 中输出 `publication_mode` 与 `mode_evidence_level`，供后续 UNA Seller 店铺模式配置使用：

- `fully_managed`：**全部**已核实子站点一致返回 `business_model = "CBT CN Fulfillment Managed"`，且主账号允许销售；混合、缺失或未知值保持 `undetermined`，并保留逐站点脱敏证据。
- `standard_up`：主账号启用 `user_product_seller`，没有全托管指示、没有矛盾的子站点 UP 标记，且账号允许销售。
- `legacy_candidate`：账号没有 UP 标签，也未返回已启用的子站点 UP；**仅是待确认的旧版候选，不能据此使用 Legacy POST**。
- `undetermined`：标签缺失、业务模式未知或矛盾、账号销售状态不明等；必须人工确认，不能按默认模型发布。

`mode_evidence_level = c0_read_only_not_publish_approved` 始终表明只是发布模型只读识别，**不等于具备创建或图片上传权限**。模式名称只用于 UNA Seller 内部适配，绝不修改 Mercado Libre 官方账号类别。没有真实账号脱敏输出前，不得将当前已连接店铺设成任何确定模式。

实际执行必须在持有有效授权、可安全刷新 OAuth Token 的受控环境中进行；不得读取 Supabase 存放的加密 Token 并尝试绕过 Go 解密，也不得把 Token 发到 GitHub Issues、聊天或日志。完成真实三次 GET 后，以官方证据将店铺按 `publication_mode` 分类，并额外核对官方可写接口权限；当前 Draft PR #85 已包含只读探针、独立私有证据表、店铺模式显示及 10 秒触发冷却；**需要独立批准迁移/部署后才能运行真实 GET 并写入真实店铺证据**。C1 发布前仍必须逐站点核对创建/图片上传权限，不能把店铺级模式当作站点级写入授权。

## 负责人只读验证

在获批的 CBT 测试账号环境运行。PowerShell 7，从仓库根目录构建到本地临时目录，使用 Go 1.26。

```powershell
$probe = Join-Path ([IO.Path]::GetTempPath()) ('una-cbt-probe-' + [guid]::NewGuid() + '.exe')
Push-Location backend
try {
    go build -o $probe ./cmd/cbt-publish-probe
    if ($LASTEXITCODE -ne 0) { throw 'probe build failed' }
    $env:MELI_CBT_PROBE_TOKEN = Read-Host -MaskInput 'Approved CBT test access token'
    & $probe
    if ($LASTEXITCODE -ne 0) { throw 'probe failed' }
} finally {
    Remove-Item Env:MELI_CBT_PROBE_TOKEN -ErrorAction SilentlyContinue
    if (Test-Path -LiteralPath $probe) { Remove-Item -LiteralPath $probe }
    Pop-Location
}
```

探针只对固定的官方 HTTPS 主机依次执行三次 GET。`GET /users/me` 核对 CBT 身份及全局销售状态。`GET /users/{merchant_id}` 读取 `user_product_seller`、`global_seller`、`warehouse_management` 标签。`GET /marketplace/users/{merchant_id}` 读取子卖家映射，输出按次运行匿名的 Seller ID、站点和已识别的配置值。`UNVERIFIED` 表示字段缺失或尚未核实，`unrecognized:<value>` 表示字段存在、未识别但符合安全枚举字符白名单（小写字母、数字、下划线，且长度受限）的配置值；无法安全显示时仅输出 `unrecognized`，`absent` 表示已返回的标签数组不包含该标签。探针结果仅用于确认身份、标签和子卖家映射，刊登权限、图片能力、预检、创建字段和商品读回需另行核实。计价枚举值如 `listing_price`、`net_proceeds` 保持原义，刊登字段映射仍需核实。 [计价枚举]

负责人在获批的 CBT 测试账号下确认主账号身份、实际发布模型及各目标子卖家，复核脱敏后提交证据。可用已有测试商品对照表中回读、SKU 查询、金额及文案字段，记录测试日期、端点、HTTP 状态、站点匹配关系和脱敏字段形状。当前仓库 Fixture 均为合成，实际账号结论保持 `UNVERIFIED`。

## 数据库迁移草案 · 待审批

两张私有 `app` 表沿用 [Go 服务的 RLS 约定](../../backend/README.md)。两表均启用 RLS，只通过 Go 服务访问。`anon`、`authenticated` 角色无授权，表不暴露给 Data API。当前 Go 数据库角色作为表所有者可绕过 RLS，所有读写、回读和审计查询必须按服务端已认证的 `owner_user_id` 隔离。提交事务须核对 `app.own_products` 和 `app.meli_store_connections` 同属该 owner，并确认当前店铺连接及目标站点映射。子身份及尝试都通过已核对归属的父身份关联。`requested_by` 来自服务端已认证用户。

| 草案表 | 最小字段及约束 |
| --- | --- |
| `cbt_publication_identities` | `id, owner_user_id, own_product_id, store_id, parent_id, scope, publication_model, variant_key, site_id, logistic_type, platform_id, platform_seller_id, creation_id, state, active_request_id, active_request_deadline, created_at, updated_at, last_verified_at`。站点行保存实际金额字段、币种、支持的库存和销售条件及平台状态。主体行保存共享库存或净回款。金额用 PostgreSQL `numeric`，平台 ID 以字符串存储。 |
| `cbt_publication_attempts` | `request_id, identity_id, requested_by, payload_sha256, submitted_at, completed_at, platform_http_status, platform_id, error_code, error_detail_redacted, outcome`。主键 `(request_id, identity_id)`。同一远端调用的多主体和站点行共享 `request_id`、执行人和精确请求体的 SHA-256。以 32 字节摘要和脱敏错误保留审计证据。 |

- **唯一约束** · 主键 `id`，服务端在首次创建意图时生成随机 UUID `creation_id`，客户端不得指定；按 `(owner_user_id, creation_id)` 唯一，安全重试沿用原值。该 ID 仅用于本地关联，不代表平台支持服务端幂等。按 `scope` 建部分唯一索引，Family 键为 `(owner_user_id, own_product_id, store_id)`，Siteless 或 Legacy 主体键为同一组字段加稳定 `variant_key`。无变体使用固定键，现有变体的键来源待审批。站点键为 `(parent_id, site_id, logistic_type)`，若确认同站同物流需要多个独立销售条件，再加对应身份维度。已关联的 `platform_id` 以 `(store_id, scope, platform_id)` 局部唯一。尝试表对 `outcome = 'submitting'` 的 `identity_id` 局部唯一。模型切换前须核对旧主体和新主体的归属关系。
- **时间戳与查询索引** · 所有时间用 `timestamptz`。`created_at` 和 `submitted_at` 由写入事务记录，`updated_at` 随状态或条件更新，`completed_at` 记录尝试结果，`last_verified_at` 记录匹配回读。按 `(identity_id, submitted_at DESC)` 查尝试历史，按 `(owner_user_id, active_request_deadline)` 对 `submitting` 行建部分索引供恢复查询。按 owner、产品、店铺读取 Family 或主体时复用相应作用域唯一索引的前缀，站点按 `parent_id` 查。
- **提交和回读** · 在短事务按唯一键建立初始 `not_attempted` 身份，按固定顺序锁定受影响的 Family、主体和站点，拒绝同一身份并发或待核实的再次提交。写入 `submitting`、请求标识、截止时间和尝试行，提交事务后才调用远端。回写或恢复时以当前状态和 `active_request_id` 为条件，保证过期回包无法覆盖更新结果。平台确认 ID 与匹配结果后进入 `success`，明确拒绝记 `failed`。超时、502、断线、截止仍无结果或缺少可信结果记 `pending_verify`，先按已知 ID 回读。官方搜索提供 `seller_sku`，须验证实际模型能保存并完整查回稳定 SKU。匹配有歧义、空结果或未核实归属时保持 `pending_verify`。确认原请求未创建目标后，可将待核实状态更新为 `failed` 并记录核实依据，之后才接受用户明确重试。站点重试按模型复用已关联的 Siteless UP 或 Legacy CBT 主体及经批准的增站入口。 [SKU 查询]
- **范围和审批** · C1 包括创建、结果回读及安全重试。已刊登商品的暂停和关闭在 C1 范围之外。账号模型、变体键、同站同物流销售条件基数、全托管站点来源、金额精度和 SKU 回读均为 `UNVERIFIED`。收到脱敏账号证据并获负责人批准后才确定正式 DDL 和迁移编号。当前迁移应用清单仍到 `0009`，`0010` 由 #66 使用。

### 多站点结果预览 · 合成数据

示例为标准 UP 的假设响应。一次提交共用已确认主体 `U1000`，三个目标分别展示结果。

| 范围 | 证据 | 显示结果 |
| --- | --- | --- |
| Siteless UP | 可信响应返回 `U1000` | `success` |
| MLM / remote | 返回匹配子卖家和刊登 ID `MLM1001` | `success` |
| MLB / remote | 站点返回明确拒绝及 `item.max_limit_publish` | `failed` |
| MCO / remote | 响应缺少该目标结果 | `pending_verify` |

[用户映射]: https://global-selling.mercadolibre.com/devsite/en_us/support-for-integrators/items-and-searches-global-selling
[User Products]: https://global-selling.mercadolibre.com/devsite/user-products-cbt
[变体与家族]: https://global-selling.mercadolibre.com/devsite/en_us/pricing-reference/price-per-variation-cbt
[全托管]: https://global-selling.mercadolibre.com/devsite/fully-managed-product-publishing
[Legacy]: https://global-selling.mercadolibre.com/devsite/global-listing
[刊登查询]: https://global-selling.mercadolibre.com/devsite/items-and-searches-global-selling
[图片]: https://global-selling.mercadolibre.com/devsite/manage-sales-global-selling/pictures
[验证]: https://global-selling.mercadolibre.com/devsite/validations-cbt
[站点结果]: https://global-selling.mercadolibre.com/devsite/en_us/user-products-cbt/create-an-item-with-size-chart
[SKU 查询]: https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/items-and-searches-global-selling
[计价枚举]: https://global-selling.mercadolibre.com/devsite/en_us/size-chart-validation/manage-promotions-gs
[卖家权限]: https://global-selling.mercadolibre.com/devsite/en_us/price-per-variation-cbt/fully-managed
