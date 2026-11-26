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
- 复用 Go 的 OAuth、超时、平台限流和请求串行围栏。不存原始响应；仅保留允许返回的订单号、支付状态、`taxable_base_usd`、税务类型和各收费金额及币种；过滤买家昵称、ID、地址、支付工具、非结构化 `transaction_detail` 自由文本和其他隐私字段。
- 缺失订单账单明确 `not_available`；有账单也只标 `review_required`。**`net_income_usd` 明确为 null**。
- 桌面按订单手动读取费用，并在同一订单、店铺、站点和货件未变化时复用五分钟内的结果。列表不逐行查询。费用按官方类型及原币精确归集，同一明细 ID 内容相同只计一次。桌面只展示已核实币种为 USD 的官方费用小计及明细；其他币种的费用、币种标签及占位行均不展示，仅在存在未核实美元金额时统一提示。后端仍保留允许返回的原币费用字段，不猜测换算；采购成本沿用 CNY。
- 单笔预览额外读取当前 Shipment 官方费用。以确认的子卖家匹配 `senders[].cost` 与 `receiver.cost_details[]`，并用 Billing 当前货件的 USD CCSI 原始扣项核对卖家运费。物流费用标为货件净成本，买卖双方运费是子项。共享货件的成本没有在订单间分摊，也没有计作订单净收益。采购、退款和冲正保留各自口径。
- **A2 前置纯算式（仅测试，未接入任何运行接口或 UI，不属于 A1 已上线功能）：** `EstimateProfitFromVerifiedNet` 接收经核实的 USD 实收（金额、来源、结算时间）、USD/CNY 汇率（1 USD 换算的 CNY、来源、核算日期）、CNY 采购和附加成本。使用精确有理数，利润 USD 舍入到分，正负半分向远离零取整。利润率沿用预计利润 / 正的净收益，显示时保留一位小数。
- 前端在净回款、费用口径或核算输入未确认时显示待核账。无新迁移。

## 真实核账记录

脱敏样本商品销售额 USD 11.28，与 Billing 的美元交易及商品金额字段一致。Shipment 官方费用中买家承担 USD 7.72、当前卖家物流净成本 USD 3.38。Billing 给出卖家原始运费扣项 USD 11.10，付款状态待释放。Shipment 的 `gross_amount` 返回 USD 11.09，保留为独立字段口径。已核对的官方接口未给出可归属该订单的 USD 7.90 总净收益或实际到账，因此当前官方总计与回款继续显示待核账。

[Mercado Libre 官方货件费用](https://global-selling.mercadolibre.com/devsite/api-docs/manage-shipments)

## 真实核账门禁

1. 以真实 CBT owner token 只读查询一组不同站点/物流、已退款/未退款/已释放/未释放的账单；脱敏比对 Seller Center 的佣金、仓配、税费、退款和实际 USD 结算值。
2. 用已释放、未释放、部分退款、全额退款、费用退回、多商品和多币种各类样本确认账单覆盖、金额方向、退款分摊、费用与回款日期。以可匹配到同一订单的 Seller Center 实收为准，保留原币与对应的换算凭证。卖家或买家标识应脱敏。
3. 确认可复用的可信美元结算来源和汇率后接通公式，核对实际 CNY 附加成本。涉及多个结算日期时保留各笔汇率来源及日期。合并和发布需负责人审批。

本 PR 仅交付 A1 只读费用预览，A2/A3 另行验收。A1 可以在代码审核及负责人批准后独立合并；不要因为已有纯算式就宣称真实利润功能上线。上游 HTTP 404 的具体语义尚未验证，不将接口不存在或权限错误猜作无账单；官方 200 且缺少对应账单时返回 not_available。


## FIN-02 2026-11-27 核账实施门禁（真实利润优先，不依赖刊登）

生产只读基线（北京时间00:21）：73笔订单，68 paid、2 partially_refunded、3 cancelled；仅3笔账单快照（MLB paid），已填实际采购成本0笔。当前两有订单站点为MLB（57笔）与MLM（16笔）。未核实订单级实际USD净回款来源，不能显示实际利润。

### 真实来源核实

- 官方 [Billing Reports by Orders and Packs](https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/billing-reports-by-orders-and-packs)：base_amount_usd 为税基，sale_fee.net 为巴西收费净额，money_release_status/date 表示释放状态/日期，均不单独证明订单实际实收。现有 billing-cache 安全复用；不重复批量读全部73单，先按子卖家、站点、物流选择最少量代表性订单，最多60/次官方请求、遇429停止。
- 官方 [All about receiving payments](https://sellers.mercadolibre.com/learning-center/news/all-about-receiving-payments/) 指向 Global Selling Payment reports 中的付款/转账报告；优先核实当前账号真实可下载的报告及字段，若仅批次金额不得强分摊到订单。官方 [Reconciliation](https://sellers.mercadolibre.com/learning-center/nota/everything-you-need-to-know-about-reconciliation) 介绍费用/退款关联及付款报告核对；注意费用付款报告与卖家收款报告可能并非同一口径。Mercado Pago Releases report 仅在当前 CBT 账户真实可用且币种/归属符合时采用；不假设本地 Mercado Pago 的报告/API 可跨 CBT 主账号复用。公开文档的搜索索引支持这些入口，真实账号权限/报告列未验证。
- 先从当前卖家中心已存在的报告导出只读样本，与实际USD收款账户流水匹配；没有完成结算或无可用报表时标明缺样，不伪造回款或自动生成/触发转账。运营日销售（订单创建日）、官方释放日、实际结算日及收款换汇日分别保存，不混作今日收入。

### 最小核账样本与实现条件

1. 现有已缓存 MLB paid 账单先复用；再选择 MLM 正常订单、MLB 部分退款、已取消/已退款（真实存在才取）、已释放与未释放。核对同一完整 owner/store/site/logistic/seller/order 作用域、Pack/Shipment 共享费用归属；报告中使用稳定订单/交易/费用/冲正/批次 ID 匹配、去重，保留原币及明确 USD 金额，避免重复扣减已净额费用。
2. 核对支付报告实际结算订单净额与收款流水。只有可靠订单级金额或可核实分摊才允许作为 VerifiedUSDNet；共享费用、退款冲正及分次回款未核实时不可用。避免在 Git、日志及聊天贴买家姓名、地址、收款账号、原始报表或授权凭证；仅提交脱敏字段定义、计数、核账示例和哈希。
3. 实际采购成本用当前订单采购关联；明确每单真实CNY采购和国内额外成本（真实为0需显式确认，未填不算0）。区分真实收款工具换汇凭据与会计/运营参考汇率，记录 CNY/USD、来源、日期、用途；不默认今日汇率或汇率7，不将参考汇率标作实际换汇。
4. 证据通过后在独立 FIN-02 Draft PR 复用 EstimateProfitFromVerifiedNet。先单订单实际净回款与成本/汇率，再按结算归属日期、店铺、站点聚合；严格区分预计利润与实际核算利润，覆盖不足显示待核实/已核实小计及覆盖率。多个结算日/退款批次需要按真实分摊与各自汇率核算；尚无证据不提前设计不可逆财务账本或大规模新迁移。不要重复扣除已经反映在净回款中的平台费用或退款。
5. 最小测试：正常订单、部分退款/冲正、未结算/缺成本/缺汇率不可用、费用不重复扣减及跨店归属。首次实际核账样本由负责人核对，再接入运营首页实际利润，不扩大报表或视觉优化。当前 FIN-02 为证据准备，尚无运行时真实利润接线。

## 2026-11-27 夜间真实核账补记

Global Selling 实际账户可打开 Mercado Pago 的资金释放报告。已有 11 月 4 日 `automatic by payout` CSV，`reserve_for_payout` 金额与同一窗口的活动列表转账金额一致。该报告提供 `SOURCE_ID`、交易类别、净借贷、费用和税项，`PURCHASE_ID` 为空。抽查的订单、Pack 与货件未匹配报告编号。转账详情需要二次验证，收款到账流水、可靠的订单到交易/批次映射和实际 CNY 成本仍缺。3 笔 MLB 已付账单快照为 1 笔释放、2 笔待释放，均待核账。继续按上述实施门禁复用纯算式，生产无实际利润接线。[夜间核账回执](NIGHT-PREFLIGHT-20261127.md)


## 2026-11-27 夜间核账后续：预计与实际利润双口径

[夜间回执](NIGHT-PREFLIGHT-20261127.md) 已确认实际账户的 11 月 4 日 automatic by payout CSV，批次扣账与活动列表金额相符，但 PURCHASE_ID 为空、抽查来源编号无法匹配缓存订单/pack/货件，转账详情需二次验证，未取得收款流水。3 个 MLB 账单 1 released/2 pending，仍为 review_required。未获订单级净回款与真实成本/汇率证据。

- 预计利润服务日常经营：经核实的订单级预计 USD 净回款减实际采购及国内额外 CNY 成本按明确核算汇率折算。预计来源须经过 Seller Center 或官方字段逐单对照，刊登 net_proceeds、订单总额、税基、局部佣金均不是预计净回款。未释放不自动等同零，缺失成本不填零。
- 实际核算利润服务对账：可信实际结算 USD 净回款减实际成本，使用实际换汇凭据或明确标记的会计核算汇率。区分到账/释放/订单创建日；同一笔订单跨批次、退款冲正和共享费用需证实分摊，不能将批次总额按销售额猜分。
- 预计和实际利润不混加；运营首页标注归属日期、数据来源及覆盖率，未知显示不可用。先取得至少一个可独立追溯的预计/实际订单级样本，再决定最小存储和读接口；不为未证实字段提前建立账本或制造数值。现有 EstimateProfitFromVerifiedNet 只用于已核实净回款的计算核心，预计口径需明确独立类型/来源/状态，不能伪装已结算。实际二次验证、真实收款流水和用户实际成本由负责人提供或授权，期间保持 CBT-02 暂停。
