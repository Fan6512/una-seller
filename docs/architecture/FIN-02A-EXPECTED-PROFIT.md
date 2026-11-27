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


## 2026-11-27 第二轮现场核验（页面 + 生产缓存只读）

基线 `83ecaf2`。用户打开已登录 Seller Center，分别展开 MLB / MLM 订单收款详情；通过浏览器可读页面文本核验，不是截图推断。随后仅 SELECT 已有生产订单、商品行与 Billing 缓存，未刷新官方接口或修改数据库。以下只保留脱敏金额及关联结论。

| 样本 | 商品 USD | 销售佣金 USD | 包裹运费 USD | 买家运费 USD | 页面总计 USD | Billing 对照 |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| MLB C | 8.74 | 1.09 (12.5%) | 10.78 | 7.63 | 4.50 | 缓存完整算术吻合，released |
| MLM D | 23.78 | 4.76 (20%) | 5.54 | 未列出 | 13.48 | 商品行吻合；无 Billing 缓存，NOT_RUN |

- MLB C 页面销售佣金 1.09 = CVMPI 0.16 (YES) + CVMPCB 0.15 (NO) + CVMLCB 0.78 (NO)；CCSI 10.78，买家运费 7.63，净运费 3.15。缓存订单 ID 与货件 ID 均匹配，approved/released，无缓存税项或冲正。旧缓存 charges 无 sales_refs；不因此判定官方费用完整或归属已证明。
- 两页 URL `/orders/{id}/detail?marketplace-user={site}-remote` 的 id 均在相应站点生产库匹配 pack_id 而非 order_id；本地同 Pack/Shipment 各一行，不代表官方只有一笔。禁止直接拿 URL id 当内部 order_id，禁止向多订单 Pack 分摊页面总额。MLB 页面另有发票入口引用不同的内部订单 ID。
- **发现订单头金额陷阱**：两条订单记录 currency_id=USD，但 total_amount 分别为 48.07、444.22，与页面 USD 商品价不同；不得用订单头金额。现有 orderSalesAmountUSD 基于明确 USD 商品行：MLB 1×8.74，MLM 2×11.89，分别吻合页面 8.74、23.78。
- 页面收款状态均为 approved，物流均已送达。用户目前没有运输中订单，真实新单/未释放自动核算仍待验；MLM 费用、税费、退款、共享货件、SOURCE_ID 匹配仍未验证。没有原始网络响应或已证明的直接净额 API 字段。
- 页面总计 MLB 为 US$ 4,50，MLM 为 US$ 13.48；销售费用标签分别为“总销售佣金”和“销售费用”。浏览器适配器必须限定收款区块、币种、金额格式、作用域；不依赖单一翻译标签，未解析/多重候选拒绝。页面总计不等同实际到账。
- 后续 API 优先：在既有授权路径中受控读取 MLM Billing（不要读取凭据或自行重放私有端点），确认商品行、费用范围、Pack/Shipment 官方完整性与税/退款边界。普通单候选算式 = USD 商品行合计 − 已归属佣金 − 净运费；不重复扣 CCSI，未知费用/共享归属/缺失证据不出数。无可靠自动来源前保持人工确认；独立采集模块作为补充，不新增通用插件框架或运行权限。

### 受限候选算术开发

新增 `billingExpectedNetCandidate`：仅 MLB remote 正常 paid 订单；必须由上游提供官方完整 Pack/Shipment 单订单关联，不接受本地计数替代；仅识别已观察到的佣金代码及唯一 CCSI，佣金显式订单归属，拒绝重复/冲正/税项/未知费用/币种不符/运费不一致，使用 USD 商品行合计和精确分值。测试复现 4.50 的真实金额但关联标识与完整性输入为合成，不代表生产自动核验通过。未接 API/UI/确认接口，不改权限、数据库、自动采集频率，不把候选升级为已确认净额。MLM 规则待实际 Billing 样本；现有旧 MLB 缓存无 sales_refs，也不会通过新候选门禁。

测试及本轮 HEAD CI 以 Actions 实际结果为准；开发时未在完整仓库运行 Go 测试。仍待使用既有授权 API 取得 MLM 缓存并验证官方 Pack/Shipment 关系及费用完整性，再决定自动源接线。无运输中订单不阻断此隔离开发，不代表新单验收通过。


## 2026-11-27 开发收口：指定 Pack 账单核验入口

- `billing-probe` 是仅供受控本地运营/开发核验的单次 CLI，不是 HTTP 新接口，不启动服务、不执行迁移、不启用自动采集。现有 `PreviewOrderBilling` 保持原缓存写入行为，新增内部 `readOrderBilling(..., persist=false)` 供探针复用现有授权、费用读取、范围复查与 parser，探针不保存 Billing 快照；正常 OAuth 续期可能仍按既有机制更新令牌。
- 在已明确为 UNA Seller 的运行环境、原配置及受控权限下构建临时二进制，不替换生产程序。使用 `UNA_SELLER_BILLING_PROBE_ENABLED=1 /path/to/temporary-api billing-probe OWNER_UUID STORE_UUID SITE LOGISTIC PACK_ID`，每次只核验一笔指定 Pack，先 MLB C（页面净额4.50），后 MLM D（13.48）。真实作用域/Pack 标识仅在本地使用，不提交仓库；核验前确认主机、授权及是否允许既有 OAuth 续期，遇错误或限流停止。
- 仅从本地当前 owner/store/site/logistic/child 范围定位唯一订单，调用现有授权路径，读取前后检查 Pack 对应订单及商品行；多行拒绝。输出脱敏商品 USD、付款状态、费用及归属匹配标志、税项和运费；不输出真实交易、付款、订单、费用、Pack ID 或凭据。官方 Pack/Shipment 成员完整性仍为 not_verified，不能从本地唯一映射替代。探针未执行真实生产请求。
- 本轮 main 新增的四次提交仅修改功能台账；Draft 自身的台账修改撤回至共同祖先，以便合并时保留 main 原有全部记录及后续变更。最新专项进度在本文件和 PR 正文记录；合流后再按正常台账维护流程更新 main。
- CI 45941646276：Desktop 与 Changed Files PASS，Go 在 candidate test gofmt 门禁失败，已修正对应字段对齐。新探针与脱敏测试以及当前 HEAD 完整 CI 待实际运行确认，不将旧 CI 或静态阅读当成通过。
- 后续顺序：先用探针取得 MLM/MLB 新 Billing 脱敏证据，再通过已授权官方关系接口核验完整性和直接净额来源；不能确定时继续人工兜底，暂不启用自动 collector.meli.billing 或将候选写入预计利润接口。不修改生产，不部署。

## 2026-11-27 指定两笔 Pack 的真实探针与隔离验证

本轮从负责人指定的 `2f11d9ea98258faff240d5f73ecc09e5a00c3c40` 继续，仅使用既有 #118 分支。CI 45941875124 在 Go formatting 门禁失败，未执行 Go 测试；Desktop / Changed Files 通过。用主机上的 Go 1.26 标准 gofmt 修正候选测试后，提交 `b86386805855cdc710f948637d119ba3d4e4f328` 的 [CI 45942567609](https://github.com/Fan6512/una-seller/actions/runs/45942567609) 全部通过。本次后续修改仍需以对应新 HEAD 的 CI 为准。

执行前确认 SSH 主机 `152.136.41.105`、hostname `VM-0-13-ubuntu`、登录用户 `ubuntu`，现行服务为 `una-seller-api`，ExecStart 为 `/opt/una-seller/una-seller-api`、PID 1280054。在 `/tmp/una-fin118-20261127/backend/temporary-api` 使用该主机既有 golang:1.26 容器构建临时二进制，未替换生产程序。配置只在主机内读取，不输出凭据；复用现有 OAuth，允许正常续期。只读查询确认两笔 Pack 各对应一个当前 connected、marketplaces verified、子卖家匹配的 owner/store/site/remote 订单。真实身份和 Pack 参数仅留主机受限目录，不提交仓库。

两笔分别执行 `UNA_SELLER_BILLING_PROBE_ENABLED=1 temporary-api billing-probe OWNER_UUID STORE_UUID SITE remote PACK_ID`。现有授权 Billing / shipment costs 均读取成功，结果如下：

| 样本 | 明确 USD 商品行 | 官方佣金明细 USD | 官方 CCSI / 买家运费 / 净运费 USD | 对照页面净额 USD |
| --- | ---: | --- | --- | ---: |
| MLB C | 8.74 | CVMPCB 0.15 NO；CVMPI 0.16 YES；CVMLCB 0.78 NO | 10.78 / 7.63 / 3.15 | 4.50 |
| MLM D | 2×11.89 = 23.78 | CVMPI 0.42 YES；CV 4.34 YES | 5.54 / 0 / 5.54 | 13.48 |

每条费用均有唯一 detail ID、一条 sales_info，order_id 匹配当前订单、shipping_info.pack_id 匹配目标 Pack，存在 operation_id；没有观察到冲正或税项。MLB approved/released；MLM approved/pending。`8.74−1.09−3.15=4.50` 与 `23.78−4.76−5.54=13.48` 均吻合页面。前后两笔 billing_snapshot 的数据库摘要完全相同；尤其 MLM 仍没有账单缓存。未保存真实确认、未写入财务账单缓存。付款释放状态不证明实际到账，费用引用匹配也不单独证明所有应计费用已完整。

### 官方关系和直接来源的边界

- 复用同一 Service 授权，按照[官方 Pack 文档](https://global-selling.mercadolibre.com/devsite/manage-sales-global-selling/packs)请求 `/marketplace/orders/pack/{pack}`：两笔均 HTTP 200，Pack ID、shipment.id 均匹配，各只有一个目标 order_id。这是官方 Pack 单订单证据，取代此前的本地单行推测。
- 按[官方 Shipment 关联文档](https://global-selling.mercadolibre.com/devsite/api-docs/manage-shipments)请求 `/shipments/{shipment}/orders`，使用既有传输层的 `X-New-Domain: true`：两笔均 HTTP 401。未尝试其它身份、替代私有接口或绕过授权。**官方 Shipment 成员完整性仍未知**，不能把 Pack 单订单推成 Shipment 单订单。
- `/marketplace/orders/{order}` 两笔均 HTTP 200；临时只读诊断在响应结构中检查 `net_received_amount`、`net_amount`、`net_amount_usd`、`net_income_usd`，未出现这些字段。没有保存或输出原始响应、买家数据或订单标识。这只说明检查的字段/资源未找到直接来源，不证明所有公开 API 或正常 Seller Center 页面响应均无直接净额。未取得新的 Seller Center 网络响应或资金报告。
- [Billing 官方使用边界](https://global-selling.mercadolibre.com/devsite/en_us/deals-gs/billing-reports-by-orders-and-packs)要求财务核对、按子卖家合批最多60单、避免重取已处理订单。不能为实时预计利润逐单轮询 Billing；本次是负责人明确指定的两笔一次性核验，后续代码验证使用合成服务，不重复下载真实账单。

### 本次开发、验证及继续条件

- 探针增加脱敏 Pack / Shipment HTTP 状态、成员行数和匹配标志；仅两者都成功且明确归属单订单时，关系字段才为 `single_order_verified`。401/403/404 保留未知，429/网络/解析错误返回失败，关系读取不自动重试。读取结束复核父授权、子卖家、Pack 映射、货件、状态和商品行。该关系标志不确认费用完整，更不是已确认净额。
- 受限候选增加本次真实观察到的 MLM remote CVMPI + CV 代码及 YES 模式，不向 MLB 放开 CV；继续要求官方 Pack 和 Shipment 完整单订单证据。测试的标识及完整成员输入为合成，两笔真实 Shipment 401 均不能通过候选门禁。未调用确认入口，没有把候选写入利润查询或首页。
- 隔离 PostgreSQL 17.11 位于本机回环端口55438、独立容器 `una-fin118-pg-20261127`；不使用生产数据库。专项测试覆盖两站点候选、未知/共享成员拒绝、探针脱敏、异主拒绝、200/401/429、无账单缓存写入、日汇率首次锁定、预计利润精确算术，以及已有确认权限/CAS/失效/首页汇总。
- 在隔离库创建合成日汇率、订单、账单和确认哨兵，再完整执行两次迁移、重启 PostgreSQL 后读回：日汇率6.6及原始6.6899、账单和确认均保留；两张表 RLS 启用，FX 对 anon/authenticated/service_role 无数据权限；错误的四舍五入汇率和非对象确认均触发约束拒绝。未新增迁移、未在生产执行0023/0024。回退运行代码可保留新增表列，未演练删除真实财务数据。

当前状态：**真实取证及受限候选推进；自动采集与自动预计利润来源接线尚未完成**。阻塞项是 Shipment 成员接口401，以及尚无已验证的直接净额/完整费用来源合同。现有人工确认和财务算术继续是唯一运行路径。独立 `collector.meli.billing` 的自动批处理、存储和利润来源身份仍须等证据门禁满足后再实现；本轮不新增空壳插件、调度器或未经证实的自动来源。下一步只需解决既有授权下官方 Shipment 关系的可读性，或取得正常已授权页面的明确订单级 USD 净额来源；之后在同一 Draft PR 接入独立已验证来源，并复用现有利润计算、日汇率和范围/CAS门禁。保持 Draft，不合并、不部署、不启用生产自动采集。

## 2026-11-27 分级预计利润及默认关闭的自动账单采集

负责人允许 Shipment 401 不再独占阻塞经营估算，但未知不得标为已验证。严格候选和人工确认合同继续保留。新增独立 `billing_estimate` 来源及 `shipment_membership_unknown` 风险：官方 Pack 必须唯一指向当前订单和货件；每项 USD 费用（含 CCSI）须有唯一明细、明确同一订单/Pack/operation 引用，已验证的两站费用代码、付款状态及运费算式一致。已知共享货件、退款、未知费用、税项、冲正、缺失明细拒绝计算。仅 Shipment 成员 HTTP401/403 允许标记未知后估算；200 空或不匹配、404 等不作为未知许可。即使成员完整，账单估算仍不是平台确认或实际到账。

- 0025 在私有现有订单表追加独立账单估算、采集尝试及明确额外人民币成本列，增加 Pack/Shipment 查询索引。预计利润读接口使用同一证据和 CAS，已核实净额优先；订单证据变化或出现已知本地共享关系时停止使用估算。额外成本可独立填写（实际为零须显式0），不会生成净额确认；订单详情显示来源和风险，运营首页分别累计已核实与账单估算，未知成本/汇率不出利润。人工确认与实际结算合同不合并。
- 新内置 `collector.meli.billing` 默认关闭，仅在 `UNA_SELLER_BILLING_COLLECTOR_ENABLED=1` 且已有 OAuth 配置时启动。当前保守参数：每2小时一轮，最近60天且创建至少1小时的 paid MLB/MLM remote 普通单，每轮一个已授权店铺/子卖家范围、最多10单，最多3次账单及关系阶段尝试；复用现有请求优先级/限流和 DB advisory lock，不依赖浏览器或新服务。关闭不删除历史数据。
- 优先复用适用的既有 Billing 缓存；否则按同一授权范围一次合批读取。将可用的脱敏 Billing 明细持久化后再逐货件核对成本和官方关系。429/网络失败停止当轮，已保存的明细可在后续轮次继续关系核验，不重复取已处理账单。无估算条件的成功观察标为待核，不以本地单行冒充官方完整性；订单证据变化时旧估算和尝试失效。崩溃在远端读取与本地提交之间仍可能导致受限重试，不宣称跨系统 exactly-once。
- 新增纯算术/风险/独立成本及 PG+合成服务批处理测试（两笔合批、关系429后复用、401估算、独立成本、首页分开汇总、已知共享失效）。无新增真实 API 调用或生产操作；未取得补充 Shipment items / orders shipments 真实结果。0025 尚待隔离迁移、最终 CI 和实际业务验收；新采集器未部署或启用。首次本轮 CI 45944661176：桌面及 Changed Files PASS，Go 格式门禁失败，已修正；完整 CI 结果以最新 HEAD 为准。

## 2026-11-27 最终技术审核回执（基线 6f7cca79）

本节为当前验收状态，前文保留各阶段取证历史。核验代码基线为 `6f7cca791f638d50ffc6b98e675b83d27fb3e366`；[CI 45947166110](https://github.com/Fan6512/una-seller/actions/runs/45947166110) 已完成且成功：Go formatting/vet/test/build、PostgreSQL 重启门禁、Desktop 及 Changed Files 均通过；未受影响的扩展按规则跳过。没有实际 CI 失败项需要修复，本轮仅补充核验记录，不修改运行代码或扩大测试框架。

### 环境及执行边界

- 再次确认 UNA Seller 主机 `152.136.41.105` / `VM-0-13-ubuntu`，生产服务 PID `1280054`、二进制 SHA256 `e524d7c9737258c15f05ee9f1eb71b3dfb56148e13c44fbfaa27547eb1cfcf80`。临时源码、API 和核验脚本位于 `/tmp/una-fin118-final-6f7cca79`，使用主机已有 golang:1.26 构建。
- 独立 Docker 网络 `una-fin118-final`、PostgreSQL 17.11 容器 `una-fin118-final-pg`、独立数据库 `una_fin118_final`，仅回环端口55439。临时 API 使用自身容器内的8787，不发布端口、不占用生产8787。仅在临时 API 中启用 Billing/FX，关闭通知与补偿任务；生产配置未修改。
- 生产库只读复制两笔必要订单财务字段及其唯一已连接/已验证授权范围。订单从隔离库空 Billing 缓存开始；不复制买家数据或全店账单。复用当时有效期超过15分钟的既有授权，在主机内部处理加密凭据和配置，不输出、不提交凭据；未触发克隆授权续期。
- 为统计官方请求，只在临时源码的既有 HTTP GET 入口添加一行资源类别日志，不记录 URL、ID、请求头或响应体；算法、授权和调度逻辑保持基线。原始基线临时 API 亦已构建。补充断言沿用既有 Go 测试机制，临时文件只在主机验收目录，不加入仓库测试框架。

### 自动采集、限流与进程重启 — PASS

| 样本 | 自动估算 USD | 来源 / 风险 | 最终账单 / 关系尝试次数 | 最终状态 |
| --- | ---: | --- | --- | --- |
| MLB C | 4.50 | billing_estimate / shipment_membership_unknown | 1 / 1 | done |
| MLM D | 13.48 | billing_estimate / shipment_membership_unknown | 2 / 1 | done |

第一临时进程完成 MLB：Billing、shipment costs、Pack、Shipment 成员各一次；Pack 匹配、Shipment 401 后保留风险并生成独立估算，没有转为人工确认或实际结算。重新启动进程后选择下一站点 MLM，首次 Billing 遇到 HTTP400 包装的 `[REPORTING_CORE_API][ML_DETAILS_BY_ORDER][429]`，既有识别逻辑转换为限流并停止当轮，持久化 pending、账单尝试1，不写估算。

实际等待110秒，超过上游60秒冷却后，**仅在隔离库提前 pending 的本地两小时调度期限**以加速验收；未修改生产参数、未重置尝试次数，也未声称真实等待了两小时。再次启动临时 API 后 MLM 恢复成功，读取其 Billing/costs/Pack/Shipment 各一次，得到13.48及401风险。完成两站点后再次启动临时 API，官方资源类别请求数为0；成本录入、重复迁移及数据库重启后再启动一次，官方请求仍为0。done 状态和尝试次数保持不变，证明已处理观察在这些重启中没有重复请求。

此外，在同一隔离 PostgreSQL 上运行既有 `TestBillingCollectorBatchResumeAndProfit`：同一子卖家两笔一次合批，关系阶段429后复用持久化 Billing、后续轮次不重取、401风险、独立成本及共享失效均通过。真实两笔属于不同站点，按当前“一轮一个范围”策略分别请求，不能将跨站点请求合为一个账单批次。

### 独立成本、HTTP 与首页口径 — PASS

隔离库明确录入以下**合成采购及额外成本**，只用于验证计算，不代表这两笔真实经营成本或利润：

| 样本 | 锁定汇率 | 合成采购 CNY | 独立额外成本 CNY | 详情估算利润 CNY | 同北京订单日、站点的首页查询汇总 CNY |
| --- | ---: | ---: | ---: | ---: | ---: |
| MLB C | 6.6 | 20.00 | 2.50 | 7.20 | 7.20 |
| MLM D | 6.6 | 55.00 | 5.00 | 28.97 | 28.97 |

缺成本时利润为空；独立写额外成本后仍为 `billing_estimate`，保留风险且 `has_confirmation=false`，旧 revision 更新被拒绝。实际 HTTP handler 的详情 GET、`extra_cost_only=true` POST 均200；未显式确认成本或混入净额确认字段的请求均400。此 HTTP 核验使用隔离测试身份验证器，不代表真实桌面登录验收，也没有对生产身份验证作旁路修改。

真实订单创建日分别为北京2026-11-06、2026-11-13：首页一致性使用同一生产 `operationsSales` 查询及对应历史日期检查，未修改订单创建时间或产品时钟。当前日首页 HTTP 查询正确排除两笔历史订单，不能把历史样本利润冒充今日汇总。已核实利润计数和金额保持0/空，账单估算计入独立累计栏，不混入核实来源。

### 真实汇率与订单日锁定 — PASS

临时 API 的 collector.fx 已从真实固定 Frankfurter v2 ECB USD/CNY 提供方取得并在隔离库保存：

| 北京核算日 | 官方观测日 | 原始 USD/CNY | 首次锁定值 |
| --- | --- | ---: | ---: |
| 2026-11-06 | 2026-11-05 | 6.6684 | 6.6 |
| 2026-11-13 | 2026-11-12 | 6.6683 | 6.6 |
| 2026-11-27 | 2026-11-26 | 6.6602 | 6.6 |

三个日期均采用前一已公布交易日并截取一位小数。对两笔订单日尝试再次保存不同合成报价7.123/7.1后，整条已锁定记录的摘要（含来源和获取时间）不变。重启前后调用同一刷新逻辑且使用拒绝网络的计数传输，缺失日期数为0、提供方请求数为0；详情继续按原创建日读取6.6。仅日期发送到公共汇率提供方，未发送订单或授权资料。

### 0023—0025 隔离迁移与审核结论

重复执行迁移并实际重启隔离 PostgreSQL，既有 `TestBillingSnapshotPostgresRestartPersistence` 的 before/after 均通过：Billing 快照、0025估算/尝试/独立成本保留，RLS继续启用；非对象估算/尝试、负成本及三位小数成本被约束拒绝。重启后两笔真实采集估算、合成成本、详情/历史日汇总及首次锁定汇率均再次核对一致。精简专项还覆盖风险拒绝、来源优先级、日界与截取、首次锁定、人工确认/CAS/失效；均通过。

结论：**基线6f7cca79的自动采集、分级估算、恢复、成本接线及订单日汇率技术验收通过，可在原 Draft #118 提交代码审核。** Shipment 官方成员仍为401，风险标识必须保留；未验证平台直接净额字段、实际结算或真实经营成本，不据此提升为“已核实净额”或实际利润。未运行 Windows GUI 人工验收，不宣称跨系统 exactly-once。两小时自然调度由隔离加速及现有生命周期覆盖，本次未实等两小时。

收尾只读复核：两笔生产 Billing 缓存摘要与验收前一致，生产仍无0023表及0025列；服务 active，PID与二进制摘要均未变化。已移除全部本轮临时 API/数据库容器、隔离网络和数据库卷，并删除临时凭据配置及作用域文件；保留受限临时源码和脱敏运行日志供核对。

本轮没有执行生产迁移、替换程序、部署或开启生产自动采集。审核通过与生产上线授权分别处理；保持 Draft，不合并。
