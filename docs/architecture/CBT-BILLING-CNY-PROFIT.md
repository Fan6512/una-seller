# CBT 官方订单账单与美元利润（PR #49）

状态（2026-11-24）。已对真实 MLB 自发货样本完成订单、Pack、Shipment 费用与 Billing 的同单只读核对。实际结算回款及退款分摊仍待核账。

## 官方合同核实

Mercado Libre Global Selling 2026-06-19 更新的官方文档：

`GET /billing/integration/group/ML/order/details?order_ids=...&seller_id=...`

- Global Selling **主账号 Access Token** 请求；`seller_id` 必须是对应站点与物流模式的 Marketplace **子账号 ID**。接口会校验层级关系，错误可能是 403。不要从桌面接受任意子卖家 ID。
- **每次最多 60 个订单 ID**；官方建议分批调用、做好本地缓存、避免重复读取已处理数据和并发请求；429 必须停止并尊重重试时间。
- 账单返回 `results[].order_id`、`payment_info[].base_amount_usd`（**付款税基**）、`payment_info[].tax_details`、`details[].charge_info`、`details[].currency_info.currency_id` 等。巴西才有的 `sale_fee.net` 是 **charge amount**，不能视为卖家净收益。
- 真实样本的 `order.total_amount` 返回当地金额但币种标识为 USD。订单销售金额以明确标注 USD 的商品行计算并与 Billing 的官方美元字段核对。刊登的 `net_proceeds` 是定价参数。
- 同一个 order 的 `details` 可能涉及多种类型/币种；没有具体实际样本证明费用、退款和税项如何形成 USD 净回款前，**禁止直接求和并展示利润**。

官方文档：https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/billing-reports-by-orders-and-packs

## 本次已实现（A1）

- Go 后端：`POST /api/v1/stores/{store}/orders/billing-preview?site=MLB&logistic_type=remote`，body `{"order_ids":["123"]}`，手动只读、1–60 笔。
- 通过现有 Supabase owner 认证、Mercado Libre OAuth 主账号与站点/物流子账号映射；**数据库再校验全部订单属于该子账号**后才查官方账单。只允许订单 ID 和安全作用域；不接收客户端自报的 seller_id，不创建虚假账单。
- 复用 Go 的 OAuth、超时、平台限流和请求串行围栏。不存原始响应；仅保留允许返回的订单号、支付状态、`taxable_base_usd`、税务类型和各收费金额及币种；过滤买家昵称、ID、地址、支付工具和其他隐私字段。
- 缺失订单账单明确 `not_available`；有账单也只标 `review_required`。**`net_income_usd` 明确为 null**。
- 桌面按订单手动读取费用，并在同一订单、店铺、站点和货件未变化时复用五分钟内的结果。列表不逐行查询。费用按官方类型及原币精确归集，同一明细 ID 内容相同只计一次。桌面只对已核实币种为 USD 的官方费用展示金额，采购成本沿用 CNY。
- 单笔预览额外读取当前 Shipment 官方费用。以确认的子卖家匹配 `senders[].cost` 与 `receiver.cost_details[]`，并用 Billing 当前货件的 USD CCSI 原始扣项核对卖家运费。物流费用标为货件净成本，买卖双方运费是子项。共享货件的成本没有在订单间分摊，也没有计作订单净收益。采购、退款和冲正保留各自口径。
- Go 纯计算 `EstimateProfitFromVerifiedNet` 接收经核实的 USD 实收（金额、来源、结算时间）、USD/CNY 汇率（1 USD 换算的 CNY、来源、核算日期）、CNY 采购和附加成本。使用精确有理数，利润 USD 舍入到分，正负半分向远离零取整。利润率沿用预计利润 / 正的净收益，显示时保留一位小数。
- 前端在净回款、费用口径或核算输入未确认时显示待核账。无新迁移。

## 真实核账记录

脱敏样本商品销售额 USD 11.28，与 Billing 的美元交易及商品金额字段一致。Shipment 官方费用中买家承担 USD 7.72、当前卖家物流净成本 USD 3.38。Billing 给出卖家原始运费扣项 USD 11.10，付款状态待释放。Shipment 的 `gross_amount` 返回 USD 11.09，保留为独立字段口径。已核对的官方接口未给出可归属该订单的 USD 7.90 总净收益或实际到账，因此当前官方总计与回款继续显示待核账。

[Mercado Libre 官方货件费用](https://global-selling.mercadolibre.com/devsite/api-docs/manage-shipments)

## 真实核账门禁

1. 以真实 CBT owner token 只读查询一组不同站点/物流、已退款/未退款/已释放/未释放的账单；脱敏比对 Seller Center 的佣金、仓配、税费、退款和实际 USD 结算值。
2. 用已释放、未释放、部分退款、全额退款、费用退回、多商品和多币种各类样本确认账单覆盖、金额方向、退款分摊、费用与回款日期。以可匹配到同一订单的 Seller Center 实收为准，保留原币与对应的换算凭证。卖家或买家标识应脱敏。
3. 确认可复用的可信美元结算来源和汇率后接通公式，核对实际 CNY 附加成本。涉及多个结算日期时保留各笔汇率来源及日期。合并和发布需负责人审批。

本 PR 是**独立财务开发入口**，在 A2/A3 完成前始终保持 Draft。不要因为已有一个纯算式就误宣称真实利润功能上线。
