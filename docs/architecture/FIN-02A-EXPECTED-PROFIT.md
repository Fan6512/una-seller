# FIN-02A · 预计利润第一切片：真实样本与独立计算合同

## 2026-11-27 样本复核（生产只读，脱敏）

复用现有3笔 MLB remote、当前 paid 的账单快照；无官方 API 重读或生产写入。以下只列财务字段，不保存买家、订单、货件、交易标识。两笔 pending 的 Seller Center 净额 7.90 / 8.99 来自此前现场回执；此次独立读库只核对其算式，不声称再次登录卖家中心。

| 缓存商品额 USD | 当前支付/释放 | 货件净成本 USD | 其他扣项 USD | 算术候选 | 卖家中心现场对照 |
| --- | --- | --- | --- | --- | --- |
| 11.28 | approved / pending | 3.38（原额11.10，买家7.72） | 无 | 7.90 | 7.90 |
| 28.13 | approved / pending | 14.50（原额14.50，买家0） | CVMPI 0.51 YES；CVMPCB 1.65 NO；CVMLCB 2.48 NO | 8.99 | 8.99 |
| 8.74 | approved / released | 3.15（原额10.78，买家7.63） | CVMPI 0.16 YES；CVMPCB 0.15 NO；CVMLCB 0.78 NO | 不验证 | 未取得订单级实际结算 |

三单各自本地同货件、同Pack计数均1，三单税务数组均空、未见冲正。**本地计数1不证明官方货件/Pack完整且单订单**；空税数组不证明完整无税。两种候选算式仅匹配两个历史样本，不是全站点生产合同。NO扣项的归属、折扣/返还/账单完整性及共享货件必须继续核实。[官方 Billing](https://global-selling.mercadolibre.com/devsite/en_us/manage-orders-cbt/billing-reports-by-orders-and-packs) 中 debited_from_operation 仅表示是否从该操作扣除；[官方 Shipment](https://global-selling.mercadolibre.com/devsite/en_us/billing-reports-by-orders-and-packs/manage-shipments) 提供关联查询，但既有账号的相关接口曾返回401。不得以本地无其他订单代替官方完整性证明，也不得从平台费用单项推断实际到账。

## 本 PR 实际交付

- 新增独立 ConfirmedExpectedUSDNet（金额、明确 Seller Center 订单预计净额来源、观察时间），不复用 VerifiedUSDNet.SettledAt；ExpectedOrderProfitUSD 输出 kind=expected、净额来源/观察时间、核算汇率与利润。
- 从现有结算利润函数抽取精确公共算术，不改变已结算来源校验、舍入或利润率原定义。不重复扣除净额已经包含的费用。成本必须明确，额外成本真实为0需显式填0；汇率需有来源、核算日期且大于0。
- 纯合同仅由测试调用，**不是自动净回款计算器、未接 API/UI、无迁移**。测试的采购成本与汇率为合成数据，不是真实经营利润。绝不将候选算式写进生产或把两笔金额硬编码为真实订单数据。

## 后续接线门禁

先由现场核实两笔 Seller Center 订单作用域/净额截图或脱敏回执、费用项目归属和 NO 扣项定义，补 MLM、退款/冲正及共享货件边界（仅现有缓存/必要少量有权真实读，不重复下载全部账单）。如无可靠自动字段，优先设计最小的明确用户确认订单净额入口，不将未确认费用候选冒充已确认金额；当前不提前创建新财务表。需要明确实际采购成本、附加成本以及有日期和来源的指定核算汇率。

然后同一 FIN-02A PR 或后续小切片接单订单预计利润及 OPS-05 运营首页；展示覆盖率、归属日期和预计口径，与 FIN-02B 实际到账彻底分开。FIN-02B 仍等待交易/批次匹配、二次验证和收款流水。CBT-02 暂缓；不部署、不发版、不执行真实官方写入。
