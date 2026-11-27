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


## 2026-11-27 负责人新增人民币核算规则

- 以订单创建日的北京时间自然日锁定同一 USD/CNY 核算汇率，只保留整数位及第一位小数（正数向下截取，6.6899→6.6、6.999→6.9、7.0→7.0），不是四舍五入，不以刷新日/收款日汇率重算历史订单。成本或退款证据变化仍可改变利润，只有汇率固定。
- 实施选择：每日开盘核算参考采用上一已公布交易日的 ECB 参考价（Frankfurter v2 固定 USD/CNY、providers=ECB、date=订单北京时间日期减一天），避免当天尚未公布时先用旧价后又重估；周末/节假日接受之前不超过7个自然日的观测值；历史补单使用相同日期规则。不是实时市场价、银行结汇价或法定报税价。记录实际观测日期、完整原始值、锁定值、来源、获取时间。数据源 [v2 API](https://frankfurter.dev/) 已查公开文档，生产主机连通性及实际返回尚未验证。
- 新增私有 app.daily_usd_cny_rates（原生迁移0023，不执行生产）；一个核算日一行，首次有效记录 INSERT ON CONFLICT DO NOTHING，重启、重复采集、源修订均不覆盖。RLS、无公开角色授权。无订单身份、凭据或业务金额发送给汇率提供方。
- 同一 Go API 进程独立轻量循环（UNA_SELLER_DAILY_FX_ENABLED=1 才启用），启动检查及每30分钟检查今日及最多一个历史缺失日期，已存在不请求；单请求8秒、每轮25秒、禁止跳转；异常停止当轮、下轮重试，不使用硬编码汇率、不影响订单或Webhook。不需要新服务/Redis/客户端逐单请求。现阶段先验证提供方可用和隔离迁移；之后随联合部署授权启用。
- 新增按订单创建日期读取锁定汇率的内部函数，以及预计人民币利润输出：已确认预计USD净额×锁定汇率−实际采购及额外CNY成本，直接精确算至分，不将已舍入USD利润乘回人民币。缺净回款/成本/汇率时仍不出数。当前尚未接订单详情及运营首页，FIN-02B 不使用此参考价冒充实际结汇。


## 接线切片与内置采集边界（本次更新）

- collector.fx 是编译时内置插件，固定外部 HTTP 传输、生命周期与重试在 internal/collectors/fx；财务仍负责日期选择、原始响应验证、截取、持久化和读回。沿用 UNA_SELLER_DAILY_FX_ENABLED，关闭不影响已锁定日期。无通用调度框架、动态插件或设置 UI；第二个真实采集器出现后再提取公共接口。
- 官方 Billing 公开字段仍不足以证明所有订单净回款，尤其 NO 扣项、税务和 Pack/Shipment 分摊；不把两个样本算式写成通用自动公式。独立预计利润接线先提供可选的人工确认来源（仅核对 Seller Center 同一订单 USD 预计净额之后才使用），未核实订单不要求填写、不自动猜值；后续验证的自动来源使用同一计算合同。
- 0024 在既有私有 app.meli_orders 增加一列 expected_profit_confirmation，完整订单主键/现有 RLS 不变。已连接且已验证 owner/店铺/站点/物流/唯一子卖家才能单订单读取或显式确认/清除；不向官方写入。有效 paid 账单为前置，服务器记录来源、确认时间、净额和显式额外成本。确认绑定订单财务快照与上一确认版本的 SHA256；退款/状态、金额、订单 last_updated、货件/Pack、账单内容变化会使旧确认不可用；仅账单读取时间变化不失效。图片变化也可能触发保守重核。未证实的共享货件净额不能以货件合计确认。
- GET/POST /api/v1/stores/{store}/orders/{order}/expected-profit 使用既有认证、订单范围和本地缓存；POST 必须显式 confirmed=true 和 revision，禁止自动重试；缺成本/日汇率/有效确认返回空利润。订单详情显示净额、固定汇率、采购/额外成本、人民币预计利润及原因；实际利润保持待核账。
- OPS-05 首页在既有 sales 查询内关联采购和日汇率，使用同一计算逻辑、按北京创建日和店铺/站点，累计每单已舍入人民币利润，显示已核算/当前 paid 数；无可用订单为空，退款及未知不算0，截断隐藏金额。不是完整平台覆盖或实际回款。
- 仅隔离合成确认测试、权限/冲突/失效及汇总检查；不录入任何生产订单确认、不调用真实账单、无生产迁移/部署。真实自动净回款来源与现场数据仍是待核项，不将本切片称为全自动净额。


## 自动净额来源继续核查（2026-11-27）

官方 [Billing Reports by Orders and Packs](https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/billing-reports-by-orders-and-packs) 字段定义明确：payment_id 是付款标识；sales_info.operation_id / order_id 是费用关联的交易/订单标识；transaction_amount_usd 是交易额而非净回款；base_amount_usd 是税基，sale_fee.net 是收费净额。shipping_info.pack_id 与多条 sales_info 有助识别共享费用，但不提供可直接信任的分摊。官方 [Reconciliation](https://sellers.mercadolibre.com/learning-center/nota/everything-you-need-to-know-about-reconciliation) 说明非直接从交易扣除的费用仍可能从 Mercado Pago 余额扣除；NO 不是免费，也不能默认全部计入本笔订单。官方 [Packs](https://global-selling.mercadolibre.com/devsite/category-dump-global-selling/packs) 与 [Shipments](https://global-selling.mercadolibre.com/devsite/api-docs/manage-shipments) 提供关联订单查询。已知实际 release CSV 的 PURCHASE_ID 为空，先前只尝试订单/Pack/货件编号；应另核 payment_id/operation_id 与 SOURCE_ID，不把批次总额当订单净额。

本次扩充既有 Billing 白名单，仅保留 payment_id、每条费用 sales_info 的 order_id/operation_id/transaction_amount_usd、pack_id；保留多订单关联，不把多条交易额累加成净回款；重复费用 ID 内容冲突仍拒绝。无 PII/原始响应、无新迁移、无额外官方请求，不改变预计利润确认/汇总的门禁。已有缓存不会凭空得到新字段，需现场代表样本受控读取。

现场继续核验（优先已登录 Seller Center 的订单详情真实显示及其正常页面读取响应，不猜私有端点、不复制凭据、不自动登录/绕过验证）：
1. 同一真实完整作用域下，获取两笔已对照 MLB 样本及至少一笔 MLM 正常单；若有退款/共享 Pack，额外各取一笔。先复用既有回执及缓存，必要 Billing 按子卖家合批、<=60，遇429停，不能逐单轮询。核实页面是否有直接的 USD 订单级预计净额字段、字段路径、净额与费用/退款状态、币种和主体归属。
2. 若直接字段可经官方公开授权 API 或用户已授权页面的正常可读数据取得，优先采用明确来源。若只有页面字段，先做独立浏览器采集插件的只读样本合同，不照搬私有 API、不持久化登录 Cookie；自动化是否支持需进一步审查。如果是货件合计、局部税基或当前刊登净收益则拒绝。
3. 对照 Billing payment_id / operation_id 与已存在资金报告 SOURCE_ID，按订单/子卖家/币种/交易身份去重；只记录脱敏字段路径、关联个数、金额和差异。不把已释放当实际收款。按官方 Pack 与货件关联列表验证共享性，401/403 或无数据仍未知。
4. 若直接字段不可用，只有代表样本证明费用、折扣、税、退款、共享归属及完整性时，才将相应版本/站点/物流的计算规则接入自动源；其它订单继续待核实。不可用时保持当前人工确认兜底，不能用已知两笔算式全量推广。

本轮只完成文档与 parser 证据准备，真实页面/官方响应尚未取得，自动预计净回款仍未完成；保持 Draft。

## 自动采集现场复核回执（2026-11-27，证据不足）

本轮代码基线 `ec3f1ba8d64978ad659c7db15758e0603f165afc`，原分支 `codex/fin02a-expected-profit`，PR #118 保持 Draft。此回执仅补充取证状态，不代表自动净回款实现或验收成功。

### 本轮实际观察

- 可用浏览器清单仅有 ChatGPT 对话和 GitHub 功能台账页，内置浏览器无标签页；没有可读取的 Seller Center 或 Mercado Pago 页面。因此没有订单详情的正常页面请求/响应，也没有可记录的净额 JSON 路径。未尝试自动登录、复制 Cookie/令牌或绕过二次验证。
- 原分支与本工作区未发现资金报告 CSV；用户 Downloads 顶层 CSV 仅为广告报告，未读取其内容。历史回执引用的资金报告 SHA-256 为 `bc7f5f77df7bd7567cc47e4fd19cb14c4eecdbe7e71598d5e88902059e8a196c`，本轮未取得对应文件或真实 Billing 标识输入。没有执行真实 `payment_id` / `operation_id` 与 `SOURCE_ID` 的集合匹配；结果是 NOT_RUN，不是“零匹配”。未扫描浏览器配置、凭据目录或连接生产服务器。
- 官方 Billing 文档直接读取返回 HTTP 403；公开搜索索引仍返回该文档字段定义。索引不是本账户响应，不用于证明账户接口可用或样本净额。未调用真实订单/Billing API，未触发缓存刷新或生产写入。

| 样本 | 已有金额与币种 | 本轮字段路径、主体及金额核验 | 状态 |
| --- | --- | --- | --- |
| MLB 历史样本 A | USD 7.90，历史现场报告 | 未取得页面响应；完整订单/子卖家/物流作用域未独立复核 | BLOCKED_ENV |
| MLB 历史样本 B | USD 8.99，历史现场报告 | 未取得页面响应；完整订单/子卖家/物流作用域未独立复核 | BLOCKED_ENV |
| MLM 正常订单 | 未取得 | 无真实页面或响应 | NOT_RUN |
| 真实退款订单 | 历史缓存回执提及部分退款，不代表本轮可读 | 退款、冲正及净额变化未核实 | NOT_RUN |
| 真实共享货件订单 | 历史缓存回执提及共享货件，不代表官方完整归属已验证 | 订单列表、货件总额与单订单净额未核实 | NOT_RUN |

### 官方 API 与报告关联边界

公开 [Billing 字段定义](https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/billing-reports-by-orders-and-packs) 提供 `results[].payment_info[].payment_id`、`results[].details[].sales_info[].operation_id` 及同项 `order_id`。这些是关联候选标识，不是已证实的订单预计净额；交易额、税基与费用净额也不能替代净回款。本轮没有验证任何官方公开授权 API 能返回本账户两笔页面净额，不能据文档缺少字段就断言全部官方 API 均不支持。

拿到原表及真实 Billing 样本后，分别将上述两种标识作为字符串与 `SOURCE_ID` 精确比较，保留来源路径，不按金额/日期模糊认领。按店铺、子卖家、站点、订单及报告口径核对币种，统计每个标识关联订单数、报告行数和冲正/费用类型；多行或多订单关系先解释再归属。历史报告未见独立币种列，需从报告所属账户与页面确认币种，不能因样本为 USD 就给所有报告行默认 USD。即使 ID 命中，也只证明关联候选，仍需核实报告金额口径与预计/已结算时间差。

### 现有 MV3 插件可行性评估（静态检查）

检查 `apps/browser-extension/manifest.json`、`background.js` 和现有预计利润 HTTP 入口：插件为 MV3，content script 仅匹配 Amazon；后台也以 Amazon 来源白名单限制采集消息，现有保存目标为商品采集箱。当前没有 Seller Center 适配器，不能直接复用商品采集请求保存财务数据。

[Chrome 官方 content script 文档](https://developer.chrome.com/docs/extensions/develop/concepts/content-scripts) 支持从当前页面 DOM 读取并向后台传递数据；默认隔离世界无法直接读取页面 JavaScript 变量，也不自动获得页面网络响应正文。故“只读当前已打开订单 DOM”在平台机制上可行，但具体选择器、金额标签、币种与订单作用域仍缺真实样本，正常请求响应的可读性也尚未验证。

后续最小方案是用户在已登录的单订单详情显式触发采集，使用精确站点/路径权限或 `activeTab` 加 `scripting` 临时授权，仅提取经过验证的订单标识、子卖家/站点/物流、USD 净额、观察时间及来源版本。缺任何归属字段、退款语义不明、页面变化或仅有货件合计时拒绝出数。先输出本地脱敏回执；不读取/导出平台凭据，不重放私有端点，不自动翻页或批量请求。

现有人工确认接口要求 `confirmed=true`、`revision` 并携带额外成本，来源语义是人工确认；自动适配器不能伪造该确认。真实证据通过后，接入独立已验证来源并保留现有身份/店铺/订单范围与失效门禁，继续复用预计利润、北京时间日度固定汇率及首页汇总，不新建一套计算逻辑。本轮不新增权限、不安装插件、不改运行代码或数据库。

解除阻塞所需输入：可读取的已登录卖家中心样本详情及其正常响应，和上述哈希对应的原报告（或含真实候选 ID 的受控本地输入）。脱敏回执只保留样本代号、字段路径、币种、金额及关联个数，不提交订单/交易 ID、账户资料或原始响应。当前停止于证据门禁，自动净额仍未实现，不修改生产环境。

本轮验证：`git diff --check` 通过，仅此文档变更；尝试运行既有 `TestExpectedProfitConfirmedNet`，当前 shell 未找到 `go`，测试未执行。没有代码改动，不将文档核对或历史 CI 当作本轮真实订单验收。
