# CBT-PUBLISH-001 · C0 契约确认单

**审批状态** · 公共 API 文档于 2026-11-23 核对。当前主账号、获批测试账号及账号专属字段均为 `UNVERIFIED` / `MANUAL_PENDING(CBT_TEST_ACCOUNT)`。真实发布、图片上传和 Windows GUI 验收分别等待 `MANUAL_PENDING(CBT_PUBLISH_AUTHORIZATION)`、`MANUAL_PENDING(CBT_IMAGE_UPLOAD_AUTHORIZATION)` 和 `MANUAL_PENDING(WINDOWS_GUI)`。

| 核对项 | 官方公开契约 | 实际账号待确认 |
| --- | --- | --- |
| 主账号、标签与子卖家 | `GET /users/me` 确认 CBT 主账号及全局 `status.sell.allow`，`GET /users/{merchant_id}` 看 `tags[]` 的 `user_product_seller` 与 `warehouse_management`。`GET /marketplace/users/{merchant_id}` 给出子卖家的 `user_id`、`site_id`、`logistic_type`，全托管指南还列出 `business_model`、`pricing_model`、`user_product`。任一子账号的 `business_model` 为 `CBT CN Fulfillment Managed` 表示全托管。缺失标签或字段时继续核实。 [用户映射] [User Products] [全托管] [卖家权限] | `UNVERIFIED` · 发布模型、启用标签、站点销售权限、Seller ID、物流及业务模式。 |
| Siteless UP / Family 创建 | 标准 UP 的单品入口是 `POST /global/items`，新 Family 批量入口是 `POST /global/user-products/families`。每个变体是独立 UP，同一 Family 共享 Family ID。已有 UP 可通过官方单独的 `POST /global/user-products/{siteless_id}` 添加站点，实际适用性待核实。全托管指南给出单品入口及 Family 更新，实际可用的 Family 创建路径需另验。Legacy 指南仅适用于仍保有旧模型资格的账号，也使用 `POST /global/items`，请求和响应不同。 [变体与家族] [全托管] [Legacy] | `UNVERIFIED` · 实际账号可用入口及变体支持。 |
| 回读与结果 | UP 用 `GET /user-products/{siteless_id}`、`GET /marketplace/user-products/{siteless_id}/mapping`、`GET /sites/{site_id}/user-products-families/{family_id}` 和 `GET /marketplace/items/{item_id}`。Legacy 可用 `GET /items/{CBT_ITEM_ID}/marketplace_items`。按实际子卖家、站点和物流核对各结果。全托管创建示例仅返回主体和 Family ID，站点结果需回读。 [变体与家族] [全托管] [刊登查询] | `UNVERIFIED` · 现有测试商品的回读权限、无返回 ID 时的稳定关联及可见延迟。 |
| 标题与描述 | 标准 UP 用 `family_name`，平台据品类和属性生成标题。同 Family 各变体应共用稳定名称。公开示例用 `description.plain_text`，概览列出站点覆盖，可随 `sites_to_sell` 创建或通过 `listing_sites` 更新，具体支持需按账号确认。Legacy 指南使用英文 `title` 和纯文本描述。仅从产品箱已保存且人工确认的 `listing_fields.titles`、可选 `listing_fields.descriptions` 或经人工选定的手工描述映射内容。 [User Products] [变体与家族] [Legacy] | `UNVERIFIED` · `en/es/pt` 可直接提交的位置、长度、站点覆盖与全托管描述支持。 |
| 价格、目标站点、销售条件 | 标准 UP 以 `sites_to_sell` 表达目的地，按 `site_id`、`logistic_type` 匹配子卖家。示例展示根级 `global_net_proceeds`，并说明 `price` 或 `global_net_proceeds` 取决于卖家计价模型，站点净回款通过 `listing_sites[].net_proceeds` 更新。Legacy 文档分别给出 `price` 和账号支持时的 `sites_to_sell[].net_proceeds`。全托管创建使用根级 `global_net_proceeds`，平台决定站点、买家价格、币种、曝光和库存。UP 库存归 Siteless 主体，站点报价独立记录。 [User Products] [变体与家族] [全托管] [Legacy] | `UNVERIFIED` · 各站点可提交价格字段、币种、精度、库存和其他销售条件，以及同一站点与物流组合支持的独立销售条件数量。 |
| 图片 | 官方上传是 `POST /pictures/items/upload`，用 multipart 字段 `file` 取得官方图片 ID。UP 创建用 `pictures[].id`。格式与大小按官方规则及目标类目确认。 [图片] [User Products] | `UNVERIFIED` · 获批账号上传权限、处理状态、图片 ID 的复用范围。 |
| 预检与错误 | `GET /categories/{id}`、`GET /categories/{id}/attributes`、`GET /marketplace/users/cap` 可检查类目、属性和站点额度。全托管文档把 `POST /global/items/validate` 标为仅 staging。发布验证错误含 `error`、`status`、`cause[]`，各 cause 可带 `cause_id/type/code/references/message`。公开示例显示顶层 200 可同时含成功和失败的 `site_items[]`。 [验证] [Legacy] [全托管] [站点结果] | `UNVERIFIED` · 受批测试环境中可用的无创建副作用预检端点、匹配当前账号的错误样本。 |

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

探针固定走官方 HTTPS 主机，依次发送三次 GET，并限制超时及响应大小。输出仅含公开能力值与每次运行内的子卖家别名。未知值显示 `UNVERIFIED`，`absent` 表示返回的标签数组缺少该标签。实际 Seller ID 由负责人在获批环境中核对。遇到 `UNVERIFIED`，在同一获批环境查证原字段并记录结论。促销文档的账号计价枚举含 `listing_price` 和 `net_proceeds`。探针保留遇到的原始已知枚举，刊登字段映射仍需核实。 [计价枚举] 现有测试刊登可用于对照上表回读、SKU 查询、金额及文案字段。记录端点、HTTP 状态、匹配关系和去标识字段形状。负责人复核脱敏后再交付 Fixture。仓库中的测试数据均为合成。

## 数据库迁移草案 · 待审批

两张私有 `app` 表。复用 `app.own_products`、`app.meli_store_connections` 和 `app.users`，按产品与店铺归属校验 `owner_user_id`。

| 草案表 | 最小字段及约束 |
| --- | --- |
| `cbt_publication_identities` | `id, owner_user_id, own_product_id, store_id, parent_id, scope, publication_model, variant_key, site_id, logistic_type, platform_id, platform_seller_id, creation_id, state, active_request_id, active_request_deadline`。站点行记录实际金额字段与币种、支持的库存和销售条件、平台状态。主体行保存共享库存或净回款。金额用 PostgreSQL `numeric`，平台 ID 以字符串存储。 |
| `cbt_publication_attempts` | `request_id, identity_id, requested_by, payload_sha256, submitted_at, completed_at, platform_http_status, platform_id, error_code, error_detail_redacted, outcome`。主键 `(request_id, identity_id)`。同一远端调用的多主体/站点行共享 `request_id`、执行人和精确请求体的 SHA-256。审计保留各次结果，错误仅存已脱敏内容。 |

- **身份键草案** · Family 使用 `(owner_user_id, own_product_id, store_id)` 的 Family 作用域唯一键。Siteless 或 Legacy 主体使用同一组三字段加稳定的 `variant_key`。用单品固定键表示无变体，现有变体的稳定键来源待审批。站点使用 `(parent_id, site_id, logistic_type)` 的站点作用域唯一键。若同一站点和物流组合支持并需要多个独立销售条件，加入经核实的额外身份维度。平台 ID 落定后以 `(store_id, scope, platform_id)` 约束已知身份。模型切换需要核对旧主体和新主体的归属关系。
- **提交和回读** · 新身份初始为 `not_attempted`。短事务按固定顺序锁定受影响的 Family、主体和站点身份，检查正在执行或待核实的身份，写入 `submitting`、截止时间和尝试行后提交事务，再调用远端。回写时匹配当前 `active_request_id`，按身份分别更新。平台确认 ID 与匹配结果后进入 `success`。明确的拒绝记 `failed`。超时、502、断线、超过执行截止时间仍无结果和缺少可信结果记 `pending_verify`，先按已知 ID 回读。官方刊登搜索提供 `seller_sku` 查询，须验证当前模型能保存并完整查回稳定 SKU。识别不唯一、无结果或关联未核实时保持 `pending_verify`，等待人工核实。发布权限启用后，仅在确认原请求未创建目标时允许用户明确重试。已确认存在的 UP 主体需保留，站点补发复用该主体及经批准的站点入口。 [SKU 查询]
- **草案审批点** · 变体键、同站同物流销售条件的基数、全托管结果的站点来源、金额精度和 SKU 回读均为 `UNVERIFIED`。批准后才确定 DDL 和正式迁移编号。当前迁移应用清单仍到 `0009`，`0010` 由 #66 使用。

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
