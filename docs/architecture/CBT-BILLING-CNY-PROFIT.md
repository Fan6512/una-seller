# CBT 官方订单账单与人民币利润（独立 Draft PR）

状态：**M3-FINANCE-A / 初始只读切片**，2026-11-21。**没有生产发布许可，不可称为已经完成的利润功能。**

## 官方合同核实

Mercado Libre Global Selling 2026-06-19 更新的官方文档：

`GET /billing/integration/group/ML/order/details?order_ids=...&seller_id=...`

- Global Selling **主账号 Access Token** 请求；`seller_id` 必须是对应站点与物流模式的 Marketplace **子账号 ID**。接口会校验层级关系，错误可能是 403。不要从桌面接受任意子卖家 ID。
- **每次最多 60 个订单 ID**；官方建议分批调用、做好本地缓存、避免重复读取已处理数据和并发请求；429 必须停止并尊重重试时间。
- 账单返回 `results[].order_id`、`payment_info[].base_amount_usd`（**付款税基**）、`payment_info[].tax_details`、`details[].charge_info`、`details[].currency_info.currency_id` 等。巴西才有的 `sale_fee.net` 是 **charge amount**，不能视为卖家净收益。
- `order.total_amount` 是销售总额，刊登 `net_proceeds` 是定价模式参数，也不能代替某笔订单实际回款。
- 同一个 order 的 `details` 可能涉及多种类型/币种；没有具体实际样本证明费用、退款和税项如何形成 USD 净回款前，**禁止直接求和并展示利润**。

官方文档：https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/billing-reports-by-orders-and-packs

## 本次已实现（A1）

- Go 后端：`POST /api/v1/stores/{store}/orders/billing-preview?site=MLB&logistic_type=remote`，body `{"order_ids":["123"]}`，手动只读、1–60 笔。
- 通过现有 Supabase owner 认证、Mercado Libre OAuth 主账号与站点/物流子账号映射；**数据库再校验全部订单属于该子账号**后才查官方账单。只允许订单 ID 和安全作用域；不接收客户端自报的 seller_id，不创建虚假账单。
- 复用 Go 的 OAuth、超时、平台限流和请求串行围栏。不存原始响应；仅保留允许返回的订单号、支付状态、`taxable_base_usd`、税务类型和各收费金额及币种；过滤买家昵称、ID、地址、支付工具和其他隐私字段。
- 缺失订单账单明确 `not_available`；有账单也只标 `review_required`。**`net_income_usd` 明确为 null**。
- Go 纯计算 `EstimateProfitFromVerifiedNet` 仅接受经核实的真实 USD 净回款、手工设置的 CNY/USD 核算汇率、1688 采购成本和打包费，使用精确有理数四舍五入到人民币分；净收入≤0 时利润率为 null；不受汇率或系统所在时区影响。
- 本轮**不新增迁移、不开放前端利润、不提交收款或退款**，不会改变 #46/#48 的订单模型。

## 同一 Draft PR 后续完成的 A2 / A3（未实现，阻断 Ready/Merge）

1. 以真实 CBT owner token 只读查询一组不同站点/物流、已退款/未退款/已释放/未释放的账单；脱敏比对 Seller Center 的佣金、仓配、税费、退款和实际 USD 结算值。
2. 根据真实样本确定费用/退款去重、币种与到账口径、最终/待结算状态，增加 PG17 按 owner/store/site/logistic/child/order 的账单缓存；禁止 N+1 和跨卖家串账。
3. 在 #46/#48 合并后 rebasing：加入账户级默认打包费（初始 ¥0）和用户配置的 USD→CNY 参考汇率、订单列表「参考净收益/预计利润」及利润率。仅在真实账单已核实且采购成本齐全时展示数字，否则显示「待账单核实」「待配置汇率」「采购成本待补」。
4. 为 60 笔一批、429、权限变更、退款、空/重复账单、不同币种及算式补简短 PG 和前端测试。Windows 稳定后统一正式发布到 v0.1.5，不提前部署。
5. 正式发布前按实际版本排号新迁移；#48 SLA 使用 0006，因此财务新迁移不得重复使用 0006，避免冲突。

本 PR 是**独立财务开发入口**，在 A2/A3 完成前始终保持 Draft。不要因为已有一个纯算式就误宣称真实利润功能上线。
