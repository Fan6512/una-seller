# UNA Seller · 功能状态总台账

> **日常功能进度与需求登记的统一入口** · 首次盘点：2026-11-23  
> 核对基线：2026-11-28 main `d2b39a6d`（TR-02E 与 V0.8.0 交付回执已归档；本次仅更新台账）；生产 API `97b31a03`（词典已发布 `b3b9f46d…`）；V0.8.0 固定源码 `6a56734f` 已签名并发布 Pre-release，测试 OTA 已验证；Windows 实机验收仍 NOT_RUN。旧基线与证据保留在 [整理前台账](https://github.com/Fan6512/una-seller/blob/abe3eb2b7c77caa589daafb61b1c42ac09c5da7d/docs/product/FEATURE-STATUS.md) 和各行历史记录。
> 产品范围：自用的 Mercado Libre CBT ERP；核心链路 **采集 → 工作稿/产品箱 → AI 整理 → CBT 刊登 → 已刊登管理 → 订单/财务**。Amazon 等平台只作为采集来源。  
> 本文件记录功能进度与验收，不替代 [产品定义](PRODUCT-WORKFLOW-AND-EDITOR.md)、[技术路线](../architecture/DELIVERY.md)、具体 PR/Issue 或官方 API 取证。发现记录与实际代码冲突时，先核实当前 main、PR 和真实验收证据，再更新本台账。

## 一、维护约定（保持简单）

- **固定 ID**：一个独立的用户功能一个 ID；新需求新增行，已有需求只更新原行，不反复生成新的任务书或重复 Issue。被延期或取消的条目仍保留原因与历史 PR。
- **已完成**：约定的最小功能范围已并入 main，且该范围的真实业务/操作验收有明确 PASS；不能把“CI 通过”当“业务已完成”。已完成的范围很小也要如实限定。
- **待部署 / 待发布**：相关代码已合并、开发收口，但当前目标环境或签名客户端尚未交付；交付后根据实际剩余验收迁入待实测。
- **待实测 / 待联调**：已有代码或交付路径，真实数据、服务凭据、账号权限、设备、GUI 或自然样本仍有缺口；只在发现实际缺陷时定点修复。
- **待开发**：业务目标仍缺实现；已开始的功能在原行标注 Draft PR、实施中或证据门禁，不另外建重复行。若只完成一部分，明确未完成子范围。
- **暂缓 / 依赖阻塞**：负责人明确暂停或已明确后置到前置闭环，不主动开 PR；恢复须符合原门禁与负责人决定。
- **候选**：未立项的增强方向，不承诺当前版本交付。
- **每条关联**：功能 ID、当前状态/下一步、关联 PR、证据、原优先级（若有）。每个 ID 只在一个当前类别；旧行中的历史阶段记录（含旧待开发/未部署/Draft 等）是当时证据，不代表当前状态，以当前状态列和最新部署报告为准。PR 合并不等于上线，真实验收通过后才关闭约定范围。涉及发布、迁移、生产变更始终单独授权。
- **新增需求时**：先在对应模块新增一行，写清最小可验收结果；确定开发时关联 PR；PR 合并和实测后更新状态。
- **普通台账直接提交 main（2026-11-24 负责人批准）**：仅限本台账的需求登记、状态、关联 PR/Issue、验收结果、经负责人确认的优先级、流程图进度同步，以及不改变合同的脱敏验收摘要归档。已有功能 PR 可随该 PR 更新台账；不再为普通进度单独建分支或 PR。
- **重要变更仍走 PR**：业务代码、重大产品需求与业务规则、架构/API 合同、数据模型/迁移、安全/权限、部署配置、CI 及仓库治理规则变更不得夹带在普通台账直接提交中；发布、迁移、生产操作仍须单独授权。
- **直接提交保护**：先读取最新 main 和事实依据，只修改相关条目与变更记录，检查差异及敏感信息；以最新 SHA 为预期父提交，使用非强制更新，遇并发或冲突重新读取并保留其他修改。保留 Git 提交历史及正常 CI，不绕过分支保护、不改保护设置；若 main 拒绝直接提交，则回退为最小文档 PR。仅在本次已授权操作中执行，不在对话结束后自动后台修改仓库。
- **优先级**：P0 = 首次真实 CBT 刊登必需；P1 = 自用稳定性与订单经营必需；P2 = 增强/非主线；暂缓 = 不抢当前资源。优先级是工作顺序，不是开发承诺。

## 二、关键事实与当前状态索引（2026-11-28）

| 指标 | 当前已核实或最新登记 |
| --- | --- |
| main | `d2b39a6d`（V0.8.0交付回执归档；本次台账更新为文档提交）；业务基线 #131 `20404c6a` 已部署生产，PR CI 46111706636 / Main CI 46112541157 PASS；生产 API `97b31a03` |
| CI | [联合 CI 45957718570 attempt2](https://github.com/Fan6512/una-seller/actions/runs/45957718570/attempts/2) PASS（业务基线 1f697fc）；#119 [CI 45958003380](https://github.com/Fan6512/una-seller/actions/runs/45958003380) PASS；[Windows 发布 CI 45960420514](https://github.com/Fan6512/una-seller/actions/runs/45960420514) PASS（76cefd1d）。此前取消记录保留于历史；不将发布 CI 等同 GUI 验收；#120 审核 HEAD `5870d3cb` [CI 45965821184](https://github.com/Fan6512/una-seller/actions/runs/45965821184) 与 [Windows 构建 45966663613](https://github.com/Fan6512/una-seller/actions/runs/45966663613) PASS，合并后 CI 待核；#121 审核 HEAD `61813971` [合流 CI 45969191526](https://github.com/Fan6512/una-seller/actions/runs/45969191526) 与 [Windows 构建 45969658126](https://github.com/Fan6512/una-seller/actions/runs/45969658126) PASS；合并后固定业务基线 `0fe79e02` [联合 CI 45970530199 attempt2](https://github.com/Fan6512/una-seller/actions/runs/45970530199/attempts/2) 四项 PASS；#122 [CI 45971124751](https://github.com/Fan6512/una-seller/actions/runs/45971124751) Desktop/Changed Files PASS，后端/MV3 按范围跳过；版本合并后 CI 待核；#125 新 HEAD 3236e22a、#126 新 HEAD eddb0a91 CI PASS；[合流 CI 45999249269](https://github.com/Fan6512/una-seller/actions/runs/45999249269) Go Backend/Changed Files PASS，Desktop/MV3 按范围跳过，#125 合并后独立 CI 被后续合并取消；#127 审核 HEAD 046463f9 [CI 46007939609](https://github.com/Fan6512/una-seller/actions/runs/46007939609) PASS，合流 [CI 46008869942](https://github.com/Fan6512/una-seller/actions/runs/46008869942) Go Backend/Changed Files PASS，Desktop/MV3 按范围跳过；#128 HEAD d4156e36 [CI 46018393964](https://github.com/Fan6512/una-seller/actions/runs/46018393964) PASS，合流 [CI 46019671633](https://github.com/Fan6512/una-seller/actions/runs/46019671633) Go Backend/Changed Files PASS，Desktop/MV3按范围跳过；V0.8.0 固定源码 `6a56734f`：[main CI 45971731341](https://github.com/Fan6512/una-seller/actions/runs/45971731341) success；[签名/Windows NSIS/Release staging CI 45974523186](https://github.com/Fan6512/una-seller/actions/runs/45974523186) success。CI/构建不替代 Windows GUI 实机验收 |
| 开放 PR | 当前无开放 PR；原文档 PR #124 的交付回执已归档至 main，重复 PR 已关闭；#120–#131 业务/交付相关 PR 均已合并 |
| 生产 | 2026-11-28 13:55Z 已升级 `97b31a03`（SHA-256 `0fcb04fe881c9cfad32a8ad01ee347ba3639a7fb4cdb836310f5a4f385de4be1`，vcs.revision=97b31a03…、vcs.modified=false，PID 2238246、NRestarts=0、本机/公网 `/healthz` 200、匿名 401；备份 `api-20261128T053852Z-pre-97b31a03` 含二进制/api.env/unit/完整词典；无新增迁移；未重建桌面）。14:36 发布词典后重启加载：PID 2252694、NRestarts=0、启动 `translated=10927`。managed 标记在临时目录 28ms、生产目录 129ms 均实测联网前拒绝。见 [TR-02E 最终验收](../architecture/TR-02E-FINAL-PRODUCTION-ACCEPTANCE-20261128.md)。2026-11-27 09:27Z 实际升级后端至 `6a56734f`（SHA-256 `5d11e2d9…`，go1.26.11，`vcs.modified=false`，PID 1650639，NRestarts=0，本机/公网 `/healthz` 200，匿名 401，无 panic）；升级前 `1f697fc8` 已备份至 `/var/backups/una-seller/api-20261127T092726Z-pre-6a56734f/` 并留回滚副本；`1f697fc8..6a56734f` 无新增 SQL 迁移；新增 `CBT_CATEGORY_DICTIONARY_DIR=/var/lib/una-seller/cbt-category`。此前 #115–#118 统一升级、FX/Billing 启用及 0021–0025 以原回执为准；历史汇率仍按计划回填。2026-11-27 09:52Z 再次升级至 `2fc79081`（TR-02B，SHA `5d22e070…`，PID 1678772，NRestarts=0，`/healthz` 200，备份 `api-20261127T095251Z-pre-2fc79081`），默认 Provider 已切 `baidu`。其他四个主题及历史回放未获新增启用授权。2026-11-27 12:43Z 已实际升级固定提交 `f644d353`，SHA-256 `b09d6f30edb9387ffea19e5cc2b1aa9c47d3d9e4faf9beb131194b16e91425a7`，vcs.modified=false，已独立 SSH 复核；备份 `api-20261127T124310Z-pre-f644d353` 含词典、配置及二进制，保留回滚；无新增迁移。20:43 设置 90 万额度并重启，最终 PID 1837357、NRestarts=0、本机/公网健康 200、匿名 401；TR-02 全量任务于 20:50 安全暂停，active 未发布。；据2026-11-27 13:47Z现场回执已升级 `4c387e7c`，SHA-256 `c78fd0de8b9cee19c3ef1c91a2147795d18e85208680778d24d630e715ea9a6a`，vcs.modified=false，PID1901794、NRestarts=0、本机/公网健康200、匿名401；备份 `api-20261127T134710Z-pre-4c387e7c`，启动词典 translated=64（人工覆盖），非全量发布；无新增迁移。已独立 SSH 复核；21:47 启动日志 translated=64。大模型 1/5/200 条受限试译完成，候选 200 条未发布。2026-11-27 15:04Z 升级 `83b505fa`（TR-02D 第二轮，SHA-256 `652920fd84cb26c6170a4ced689ca35ec028a7ea1f27943bd102e1140145467b`，vcs.modified=false，PID 1974968、NRestarts=0、本机/公网健康 200、匿名 401，备份 `api-20261127T150457Z-pre-83b505fa`，无新增迁移）；23:05 追加六条覆盖至 70 条，未重启 API，热加载按 mtime 生效。2026-11-27 15:48Z 已升级 `c155228f`（TR-02E，SHA-256 `d8db7ae32205e1902345f7ec243ab6acdbabbcb4f3edef882c5f58e7ff2e3d82`，vcs.modified=false，PID 2007640、NRestarts=0、本机/公网健康 200、匿名 401，备份 `api-20261127T154826Z-pre-c155228f`，无新增迁移）；23:05 起十条纠错追加至 80 条，未重启 API。TR-02E 全量构建于 2026-11-28 02:10Z 完成，候选 10,854 条、延期 28 条、回执 10,480 条，额度 847,341，未发布。 |
| Webhook | 真实货件持久入库/消费 PASS，原公网/上游503（351ms）未改判；#117隔离暖连接 PASS、冷连接仍503，ACK 修复据现场回执已部署，待上线后自然样本（原503不改判） |
| 自动补偿 | 30/60分钟节律已真实复核，五站点断点推进；不重开调度开发 |
| 桌面 | [V0.8.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.8.0) 已发布（Pre-release，非正式 Latest）；安装包 SHA-256 `a6a7f3bc7be9a16898c7ea90fbf581c2e1a8a7be051b3dd5182bc1b6c0ffc36a` 与 GitHub digest 一致；`latest.json`、签名资产存在；Tauri OTA 签名核验通过（非 Authenticode）；测试通道200、正式通道204（no-store）；Windows 实机安装升级、重启及 GUI 验收 NOT_RUN，见 [V0.8.0交付回执](../architecture/V080-JOINT-DELIVERY-20261127.md) |
| 词典 / AI | 官方 Dump 2026-10-14T21:12:35Z（11,668 条、停用 714、有效 10,954）。`active.json` 于 2026-11-28 14:36 CST **正式发布**为 `b3b9f46d…`（发布前 `be52b174…` 已备份），中文 **10,928/10,954**（启动状态计数 10,927，CBT106855 译名与英文名相同），26 条回退英文。`overrides.json` **115 条**（TR-02C 64 + TR-02D 六条 + TR-02E 十条 + 首批/最终 24 条 + 本轮复验 11 条）。TR-02E 最终离线候选 **10,928 / 10,954 有效中文（99.76%，含 8 条术语）**，候选 SHA-256 `b3b9f46d…`：115 人工 + 10,805 大模型 + 8 术语 + 26 英文延期。300 条独立复验 **289/300（96.33%）**，11 条错误全部逐条纠正并保留原始记录；publish 经负责人批准 **已 apply**（`applied=true`，`original_ok=289`、`audited=300`），已写入 `llm-release-managed.json`，普通 `update` 实测被拒。大模型本地 **847,341 / 900,000**，余 52,659；普通额度因 13:57Z 一次误跑升至 **275,291**（已在验收报告披露，不回退）。详见 [TR-02E 最终验收](../architecture/TR-02E-FINAL-PRODUCTION-ACCEPTANCE-20261128.md)。 |
| CBT | C0 单店五站点只读 PASS，standard_up/net_proceeds/remote；原始业务枚举及真实写权限另核。PUB-05A 代码合并，CBT-02 暂缓，未真实上传/创建 | 
| 双账号 | 2026-11-27 核心实机 PASS（AUTH-01/02）：两账号各一家 UP 店（负责人确认）、各五条 verified 映射；生产库原账号订单/刊登/采集 73/207/4，新账号 14/50/1。API 主动越权、长期 Token 续签、公开注册关闭状态仍待核实；不阻塞 FIN-02A。 |
| 财务 | 73订单、3账单缓存及0实际采购成本为11-27 00:21历史只读基线，不冒充实时。两笔MLB待释放预计净额7.90/8.99已有现场样本；实际结算订单映射、到账、成本及汇率缺证据。FIN-02A #118 已合并，隔离真实 MLB/MLM 自动估算 USD4.50/13.48 及日汇率锁定6.6 技术验收 PASS；据现场回执服务端升级及 FX/Billing 启用已完成，历史汇率回填中；Windows 利润口径与真实成本录入待验；FIN-02B实际结算待开发 |
 
整理仅基于仓库与已登记验收证据，不重新查询生产数据库或官方服务，不将过往阶段的历史记录当成当前运行状态。

## 三、已完成（限定范围内真实验收通过）

只关闭已验收的最小范围。扩展功能及未验收场景继续由对应 ID 跟踪，不自动标 PASS。

| ID | 功能 | 当前状态 / 下一步 | 关联证据及历史范围 | 原优先级 |
| --- | --- | --- | --- | --- |
| BASE-01 | Go + PostgreSQL 基础采集 API、去重、读回和 GitHub 按范围 CI | 基础采集 API / 入库读回及 CI 已通过。 | 关联：[#9](https://github.com/Fan6512/una-seller/pull/9)、[#10](https://github.com/Fan6512/una-seller/pull/10)、[#11](https://github.com/Fan6512/una-seller/pull/11)、[#51](https://github.com/Fan6512/una-seller/pull/51)；已验收范围：CI/PG 定点用例通过；不表示全部业务模块已上线 | 原阶段未单列 |
| AUTH-01 | Windows 邮箱登录、重开恢复与退出登录的基本闭环 | 单账号 Windows 登录/退出及重开恢复、真实双账号切换与核心业务数据隔离 PASS；API 主动跨账号读写越权、长期 Token 续签、公开注册关闭状态仍待专项验证，macOS 见 DESK-02，冷启动见 DESK-04。 | 关联：[#15](https://github.com/Fan6512/una-seller/pull/15)；已验收范围：2026-11-17 单账号 Windows 实测；2026-11-27 负责人实机确认：新账号独立登录、已刊登商品同步、插件单品采集、退出并切回原账号后订单/商品/采集箱恢复均 PASS；生产库两轮只读复核见本日变更记录。仅关闭核心业务验收，不视为全部安全验收 | 原阶段未单列 |
| AUTH-02 | 一家 CBT 店铺 OAuth 连接、断开/重授权、真实站点映射 | 两家独立真实 CBT 店铺、各五条 verified 授权映射已核实；原店铺连接未受影响。 | 关联：[#23](https://github.com/Fan6512/una-seller/pull/23) / [Issue #21](https://github.com/Fan6512/una-seller/issues/21)；已验收范围：2026-11-19 一个真实账号与五条 verified 映射 PASS；2026-11-27 新增第二个真实账号及第二家 UP 店（店铺类型由负责人确认），两账号各一条 connected 连接、各五条 verified 映射；同一卖家跨用户重复归属为 0。新账号断开/重授权及 Token 自动续签未专项实测，不代表真实创建或完整物流/账单对账已通过 | 原阶段未单列 |
| COL-01 | Amazon 单商品点击采集 → Go 入库 → Windows 采集箱读取的最小闭环 | 单商品采集至 Windows 采集箱最小闭环已通过；质量与复杂变体见 COL-02。 | 关联：[#11](https://github.com/Fan6512/una-seller/pull/11)、[#13](https://github.com/Fan6512/una-seller/pull/13)、[#19](https://github.com/Fan6512/una-seller/pull/19) / [Issue #6](https://github.com/Fan6512/una-seller/issues/6)；已验收范围：真实商品曾成功采集并读回；高清图、重复图、部分成功和复杂变体的质量验收仍列在 COL-02；2026-11-27 负责人实机使用 V0.1.4 浏览器插件在新账号采集 1 件，生产库新账号采集箱 0→1、原账号保持 4；切回原账号读回正常。 [插件构建](https://github.com/Fan6512/una-seller/actions/runs/45857160861) | 原阶段未单列 |
| CBT-01 | C0 官方发布模型合同、只读探针、生产部署与现有账号模式识别 | 单账号真实 C0 只读与手动探针已通过；真实创建见 CBT-02。 | 关联：[#69](https://github.com/Fan6512/una-seller/pull/69)、[#73](https://github.com/Fan6512/una-seller/pull/73)、[#85](https://github.com/Fan6512/una-seller/pull/85)、[#87](https://github.com/Fan6512/una-seller/pull/87) / [#50](https://github.com/Fan6512/una-seller/issues/50)；已验收范围：2026-11-24 单账号真实只读 C0 PASS，standard_up；五站点证据完整，最新 14:39:47 北京时间。[PR #87 生产升级与真实 C0 验收摘要](../architecture/GO-API-PR87-DEPLOY-REPORT-20261124.md)。V0.4 实测已授权店铺、五站点、四标签及手动只读 C0 4/4 PASS；不包含原始枚举、后台独立确认、真实上传/创建权限及自动授权探测实机验收 | 原阶段未单列 |
| OPS-01 | V0.7 通知基础设施（WH-01/02）：统一 Webhook 接收、事件队列、Go Worker、去重/资源合并、可靠持久化与快速确认、限流、重试及重启恢复；六类主题按 WH-03/04/05 实施 | 统一接收、持久队列及 Worker 基础设施已部署，0014–0020 已迁移，真实货件持久收件与消费已有证据。仅关闭基础设施开发及该有限范围；公网200 ACK 及其余主题由 WH-06 跟踪。 | 原范围/前置及阶段记录：[#99](https://github.com/Fan6512/una-seller/pull/99) 已 Squash Merge 至 `5874f7c5`，审核 HEAD `481e974e`，WH-01/02 代码完成；旧 HEAD `fdcd2ae9` 复审/验收通过，后续复审修订已提交，新 HEAD [CI 45634604451](https://github.com/Fan6512/una-seller/actions/runs/45634604451) 四项 PASS，Go 格式/vet/test/build、PG 真停机 503 与重启读回均实际 PASS，负责人批准后已合并；[CI 45625829215](https://github.com/Fan6512/una-seller/actions/runs/45625829215) 四项 PASS，PG 停机/重启读回实际 PASS；据隔离 PG17.12 报告，迁移 0014、新装/重复/旧回执升级、并发去重、租约/fencing、授权隔离、重试/429/死信/容量与保留 PASS，回环 HTTP 最大 ACK 重投 223ms/不同通知 9.72ms、故障 503；订单/货件处理器已由 WH-03 #100 合并注册，仍默认禁用；公网来源校验、边缘限流、真实 ACK、订阅及补偿待 WH-06，未执行生产迁移/启用。复审修订：已知暂不可用授权返回 503、未知/歧义 403；启用前强制注册全部接收主题处理器，未注册积压或全局容量达 80% 告警；保守单 Worker 调度，WH-03 共享限流后评估；不透明通知 ID、应用级统计、百分位及 CI 命名修正，新回归 CI PASS。沿用 `https://mkd.186512.xyz/api/v1/mercadolibre/notifications`，不凭通知直接改业务状态，不自动官方写入；关联与历史证据：[审核记录](../architecture/OPS-01-WEBHOOK-QUEUE-REVIEW.md)；[Delivery](../architecture/DELIVERY.md) M3 | **V0.7 P0（该阶段基础设施；不替代首次 CBT 刊登 P0）** |
| OPS-03 | V0.7 登录后店铺检查、Webhook 优先及低频订单/物流补偿、客户端缓存读取 | D2 已生产启用；首轮、后续30/60分钟节律及五站点断点复核 PASS。真实漏单发现、429 恢复及 Windows 缓存交互仍待自然样本/实测，不重开调度开发；归 WH-06/实机验收跟踪。 | 原范围/前置及阶段记录：代码已合并、生产启用预检；订单30分钟、物流60分钟、客户端读自身 API 30秒；复用 WH-03 公共协调器（人工优先、应用级持久化冷却/互斥候选，待验收），统一顺序与 429 退避；登录后增量检查且短期不重复；服务器持续处理 Webhook、低频补偿；任务不重叠，账单及完整物流历史手动；实际参数以验证为准；关联与历史证据：[#104](https://github.com/Fan6512/una-seller/pull/104) 已 Squash Merge `49176630`（批准 HEAD `13204403`），[CI 45762678351](https://github.com/Fan6512/una-seller/actions/runs/45762678351) 四项全绿、PG 真停机恢复 PASS；漏单搜索/断点/统一队列、generation、客户端缓存刷新及夹具修复保留；真实 CBT 搜索及 Windows 交互 NOT_RUN，20:36远程只读预检：1有效授权/5映射、73订单、68到期订单/20去重到期货件、收件/活动/死信/发现/回放/冷却均0；部署版已有每分钟受限调度、30/60分钟、来源区分及共享协调。D2 已启用，首轮生产验收 PASS，不将补偿当作真实推送；20:40:39 生产启用，20:51现场验收53笔订单与20货件，五站点首窗完成；20:54远程独立只读复核73条 compensation、活动/死信/冷却/回放0、五站点revision=2，下一轮均按30分钟排程。11-27 00:21-00:23远程复核：五站点持续推进（四站revision16、一站revision14到期），订单490次/68资源、重复422最短1858.006秒；货件38次/20资源、重复18最短3601.265秒；00:21活动/死信/冷却均0，回放0。00:38 复核订单511次/68资源、货件38次/20资源，五站点均revision16，队列、死信、冷却为0。观察到30/60分钟节律 PASS；真实新订单发现入库及实际429恢复仍待样本；共享订单更新时间可由货件写回推进，不能将到期0解释为每笔均已订单GET，后续按实际需要核查订单独立刷新时间。[生产回执](../architecture/WH-06-PRODUCTION-DEPLOY-20261126.md) | V0.7 P0 |
TR-02E | 公共 CBT 词典完整路径全量构建、独立质量复验及受保护正式发布 | 2026-11-28 已完成正式发布：独立 300 条复验 289/300（96.33%），11 条错误全部逐条纠正；最终中文 10,928/10,954（99.76%），26 条英文延期；负责人批准 apply，active 切换至 `b3b9f46d…`，115 条人工覆盖及 managed 标记生效；普通 updater 联网前阻断 PASS。每日增量维护不属于本完成范围，继续由 TR-02 跟踪。 | 关联：[PR #129](https://github.com/Fan6512/una-seller/pull/129)、[#130](https://github.com/Fan6512/una-seller/pull/130)、[#131](https://github.com/Fan6512/una-seller/pull/131)；[最终生产验收](../architecture/TR-02E-FINAL-PRODUCTION-ACCEPTANCE-20261128.md)；[发布回执](../architecture/TR-02E-FINAL-PRODUCTION-ACCEPTANCE-20261128.md)。负责人2026-11-27批准完整路径方案及调整后质量门禁：原始样本及最终独立样本准确率≥95%，官方ID/路径/刊登权限100%不变，已知错误全部修正，人工覆盖优先，未确认回退官方英文；首轮及第二轮合计384/400（96%），历史严格零错误门禁FAIL不改判。当前大模型本地847341/900000、余52659，普通192758；控制台实际读数待核，只向上对账。最终独立300条复验；[发布前交接](../architecture/TR-02E-FINAL-AUDIT-HANDOFF-20261128.md)，候选合并后另行开发原子发布/回滚、Windows真实验收及每日增量；不自动发布、不改变普通在线翻译。[设计](../architecture/TR-02E-FULL-BUILD.md) | P0
| TR-02B | 百度通用文本翻译接入：统一接口/缓存/批处理、语言映射、月额度硬上限、默认 Provider 切换 | #123 已合并 `2fc79081` 并部署；五批真实试译 5/5、0 失败、2,789 字符，staged 94/10,954；缓存命中、断点续跑与额度持久化 PASS；质量 5/80 明确错误，全量前需上下文或 overrides。未全量、未配 cron。 | 关联：[PR #123](https://github.com/Fan6512/una-seller/pull/123)、[TR-02B 验收](../architecture/TR-02B-BAIDU-ACCEPTANCE-20261127.md)；保留 Azure/火山 Provider；公共维护 Key 与用户 Key 额度隔离；默认 50,000 字符/月，超额发送前拒绝 | P1 |

## 四、已合并，待交付或验收

### A. 待部署 / 待发布

V0.8.0 联合交付及签名 Pre-release 已完成；当前没有已合并但尚未部署/发布的条目。正式 Release 仍需另行批准；Windows 实机验收列在下一节。

### B. 待实测 / 待联调

已有代码或交付路径；先按剩余验收项复验，只有实际缺陷才定点修复。缺凭据/样本/设备时记阻塞，不假报通过。

| ID | 功能 | 当前状态 / 下一步 | 关联证据及历史范围 | 原优先级 |
| --- | --- | --- | --- | --- |
DESK-01 | 桌面版本交付、正式版/可选测试版 OTA 更新通道 | V0.8.0 已完成签名 Pre-release 与测试 OTA 交付：tag `desktop-test-v0.8.0`，固定源码 `6a56734f`；main CI [45971731341](https://github.com/Fan6512/una-seller/actions/runs/45971731341)、签名/Windows NSIS/Release staging CI [45974523186](https://github.com/Fan6512/una-seller/actions/runs/45974523186) PASS；测试通道 200、正式通道 204（均 no-store）。**Windows 实机验收 NOT_RUN**。下一步负责人从 V0.7.0 开启测试版更新升级至 0.8.0，重启后核对登录/配置/本地资料保留；关闭测试版开关确认不再提供测试版且已安装 0.8.0 不降级；再验收店铺/订单/商品/财务、通知和翻译设置。不得把 CI/Release 存在视为 GUI PASS。 | 原阶段记录：代码已合并，V0.4 Pre-release 已发布。公网 OTA 接口 PASS，V0.4 覆盖安装及检查更新等 5/5 PASS；跨版本 OTA 待验收；V0.5 #92 隔离 QA 构建 PASS，生产升级后双通道 OTA 接口复验 PASS（现场报告），V0.5 签名测试版已发布，实机跨版本 OTA 待验收；关联：[#29](https://github.com/Fan6512/una-seller/pull/29)、[#65](https://github.com/Fan6512/una-seller/pull/65)、[#92](https://github.com/Fan6512/una-seller/pull/92) / [#28](https://github.com/Fan6512/una-seller/issues/28)；验收判据：公网 `/api/v1/desktop/updates` 带 `X-Una-Include-Prerelease: 1` 返回 200（`0.4.0`），默认通道返回 204，两种响应均带 `Cache-Control: no-store`。现有 V0.3 无正式 Updater，已手动覆盖 V0.4；V0.5 测试版发布后默认通道 204、测试通道返回 0.5.0，两者均 `no-store`，清单安装包/签名公网下载与 SHA-256 一致；实机安装重启及实际通道选择待验收。正式 Release 需单独授权；[V0.7.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.7.0)；[Windows 发布 CI 45960420514](https://github.com/Fan6512/una-seller/actions/runs/45960420514) PASS | 原阶段未单列
PUB-07 | 已刊登商品双视图：家族固定尺寸卡片＋独立整行规格展开区；刊登独立主图墙＋按需右侧详情 | V0.8.0 已交付测试版，Windows GUI 实机验收 NOT_RUN。负责人核对家族卡片固定尺寸、单家族整行展开、3/5/10 规格换行、刊登独立主图墙、右侧详情、窗口重排、家族/店铺/站点/分页切换不残留旧内容及未保存草稿保护；仅检查展示，不触发官方写入。 | 家族默认收起，仅同一时间展开一个家族；点击另一家族规格入口自动收起旧组，展开区插在该家族所在行下方，3/5/10 个规格随可用宽度换行，其他卡片尺寸不被拉伸。官方 Family ID 优先、Siteless Family ID 其次，无关联独立展示；本阶段沿用现有站点/物流模式与分页，明确当前页归组而非全店完整家族。刊登视图每条官方刊登单独主图与标题，不自动展开详情；点击图片右侧打开共享详情，左侧按宽度重排，点击其他商品原地更新，关闭恢复全宽；家族展开状态不因详情开关丢失；切换店铺/站点/分页清空选择，保留既有本地草稿语义，保护未保存编辑。主图等比完整显示、懒加载、左右独立滚动；不新增平台写入、同步、迁移或插件平台。最小验收：3/5/10 规格、单家族展开、窗口重排、详情关闭/切换及隔离；全店跨页归组后续独立评估。；详情中文化沿 TR-01 同一 #120 实施，复用现有翻译服务，不翻译用户填写内容；因此本 PR 新增小范围 Go 元数据取词改动，交付需后端与客户端版本配合，不新增迁移或官方写入；[最终 CI 45965821184](https://github.com/Fan6512/una-seller/actions/runs/45965821184)、[Windows NSIS 45966663613](https://github.com/Fan6512/una-seller/actions/runs/45966663613) PASS；负责人 2026-11-27 明确批准合并；与 #121 合流保留 TR-01 元数据取词。 | 本期 P1（负责人批准推进）
TR-03 | 火山引擎通用机器翻译（AK/SK） | 已随 `6a56734f`/`2fc79081` 部署并包含在 V0.8.0 客户端；Windows GUI 配置验收 NOT_RUN，需确认选项显示火山通用翻译、不出现豆包、保存与错误提示正常。真实接口此前返回 `-403 account status abnormal: unsynchronized`，账号状态恢复前不重复真实调用、不标记试译 PASS。 | 负责人取消豆包机器翻译（非持续免费），删除适配器、账号选项、CLI及环境变量，不新增迁移；保留 Azure、百度。火山固定 TranslateText 2020-06-01、HMAC-SHA256 AK/SK、16条/5000字符、业务状态及结果数量校验、失败回退、独立维护凭据。类目属性在线读取限时10秒，单请求8秒，后台词典批次30秒，订单读取原4秒；大量枚举仍可能部分回退，公共离线覆盖另行推进。401/403两分钟冷却及脱敏日志，429/业务限流冷却，测试连接提示检查凭据/服务/余额。先五批真实试译并核对免费额度与计费，批准后全量；无月额度硬限制、无生产调用/部署。#120 已合并，已同步 main 776ba5a4，回读确认元数据取词函数与 #120 原样一致，仅原范围12文件差异；合流 [CI 45969191526](https://github.com/Fan6512/una-seller/actions/runs/45969191526) 与 [Windows NSIS 45969658126](https://github.com/Fan6512/una-seller/actions/runs/45969658126) PASS；负责人明确批准合并；真实接口和 GUI 待验。 | P1
| PUB-05A | 单店单站点 UP 多变体 Family 核心 | #115 已合并；据现场回执联合服务端升级及 V0.7.0 发布已完成，待负责人验证多规格编辑及刊登准备界面；CBT-02 暂缓，禁止真实刊登。 | 原范围/前置及阶段记录：CBT-02 前置；稳定内部变体身份、SKU/库存/成本/净收益、PARENT_PK/CHILD_PK、官方图片ID、Family/UP/站点ID映射及防重；复用现有产品箱，单/多规格共用刊登模块；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50#issuecomment-7192167415)，[#115](https://github.com/Fan6512/una-seller/pull/115) 已 Squash Merge [`1d8cb563`](https://github.com/Fan6512/una-seller/commit/1d8cb5639fbd101a592e24e0f9613bc8b4c5a122)（审核版 `4447e47`，合并前仅同步 main 台账）；A1/A2/A3 统一单/多规格、双标题策略、稳定身份及库存、产品箱接线、官方候选解析及逐规格回读、请求持久化防重已实现；0021/0022 新增私有刊登及恢复记录，受控恢复保留原尝试和审计，明确未创建才释放并使旧修订失效；Go 50s/Nginx专用60s、历史类目过滤已修复；[CI 45846191304](https://github.com/Fan6512/una-seller/actions/runs/45846191304) 四项 PASS，[最终复审](https://github.com/Fan6512/una-seller/pull/115#issuecomment-7197379339) PASS，负责人已批准并合并，模拟开发阶段结束；合并后 CI 第 3 次运行四项通过。加密备份隔离恢复、0021/0022 原生迁移两次及权限核对通过，候选 Nginx 60/20 秒路由验证通过。[夜间预检](../architecture/NIGHT-PREFLIGHT-20261127.md)。生产迁移、部署、真实图片上传及创建仍为 NOT_RUN；[标题策略补充](https://github.com/Fan6512/una-seller/pull/115#issuecomment-7192939051)：复用已有 platform_auto/localized，根级英文 Family Name 与目标站点自定义 family_name 分离，站点覆盖先模拟及真实回读验收，不以 UP 标识限制多语、不静默降级；2026-11-27 负责人转述本地联合交付回执（V070-JOINT-DELIVERY-20261127.md，本次未直接读取原文件） | P0 |
| OPS-05 | 运营首页真实每日经营总览：今日销售额、今日利润，后续按需增加订单量/客单价与趋势 | 销售额 #116、预计利润 #118 已合并且据现场回执已上线；V0.7.0 已发布，待手工核对北京日、店铺/站点销售额、利润及覆盖率，人工核实与账单估算不重复；实际结算归 FIN-02B。 | 原范围/前置及阶段记录：订单模块提供按北京时间自然日、店铺/站点授权范围的本地销售聚合；明确支付/取消/退款、订单去重及币种口径，禁止跨币种直接相加、禁止平台实时逐单请求。财务模块 FIN-02 提供经核实净回款、成本、汇率与利润聚合，区分预计/已确认、覆盖率及归属日期；缺失显示不可用而非零。不以订单总额、刊登净收益或不完整账单推算利润；运营中心仅只读消费稳定接口，不复制财务账本；关联与历史证据：[#116](https://github.com/Fan6512/una-seller/pull/116) 已 Squash Merge [`ffc90614`](https://github.com/Fan6512/una-seller/commit/ffc90614acabb93f6db2374b6a9987ea8e5aadf5)（最终审核 `fd746b2`），[最终复审](https://github.com/Fan6512/una-seller/pull/116#issuecomment-7198677425) PASS，非阻断收口修复日期格式、汇总不可用提示及午夜测试，明确创建日且当前paid；[最终 CI 45925033420](https://github.com/Fan6512/una-seller/actions/runs/45925033420) PASS，负责人已批准合并。缓存按北京时间订单创建日、授权店铺及站点核算明确 USD 商品额。跨午夜隐藏旧数据。历史真实订单带交易日及两站点抽查通过，今日空缓存已标明平台覆盖未确认。[CI 45860271541](https://github.com/Fan6512/una-seller/actions/runs/45860271541) Go/PG、桌面和 Changed Files 通过，扩展按范围跳过。Windows 手工验收及生产接线待执行。[回执](../architecture/NIGHT-PREFLIGHT-20261127.md)；[联合 CI 45957718570 attempt2](https://github.com/Fan6512/una-seller/actions/runs/45957718570/attempts/2) PASS；2026-11-27 负责人转述本地联合交付回执（V070-JOINT-DELIVERY-20261127.md，本次未直接读取原文件） | 当前业务重点：先真实销售聚合；利润依赖 FIN-02 对账，暂不做趋势图/复杂报表 |
| FIN-02A | 预计利润：订单级预计净回款、采购与独立额外成本、北京订单日固定汇率 | #118 已合并；据现场回执服务端升级、FX/Billing 启用完成，历史汇率回填中；V0.7.0 待负责人实机核对估算风险、观察时间、日汇率锁定、实际成本录入、详情/首页一致性及保存失败保留输入。 | 隔离真实MLB/MLM自动估算USD4.50/13.48、汇率6.6锁定、独立成本与历史日汇总、限流恢复、重启与幂等迁移通过；人工确认优先、账单估算独立标注、Shipment 401风险保留、缺成本/汇率不造数。collector.fx / collector.meli.billing 默认关闭；[专项回执](../architecture/FIN-02A-EXPECTED-PROFIT.md)，[最终 CI 45957315802](https://github.com/Fan6512/una-seller/actions/runs/45957315802)。真实成本/实际结算不在验收范围。；2026-11-27 负责人转述本地联合交付回执（V070-JOINT-DELIVERY-20261127.md，本次未直接读取原文件） | P1 |
| WH-06 | V0.7 通知真实联调与补偿：六类主题逐步启用，遗漏事件查询/回放、故障恢复及处理日志 | #117 ACK 修复已合并，隔离及 CI PASS，据现场回执已随联合后端上线；自然公网 HTTP200<500ms、订单通知、去重、真实429仍待样本；原503不改判，冷连接风险保留。其余四主题与回放仍独立审批。 | 原范围/前置及阶段记录：OPS-01/WH-03/04/05；先核对真实来源、ACK、持久化和重试，再经授权逐主题启用及修改订阅；关联与历史证据：[#106](https://github.com/Fan6512/una-seller/pull/106) 阶段 A 已合并，断点按应用、主题、授权店铺隔离。[#111](https://github.com/Fan6512/una-seller/pull/111) 修复显式 null 空页。固定提交 `b1dd92ff` 的 [合并后 CI](https://github.com/Fan6512/una-seller/actions/runs/45776485335/attempts/2) 通过，官方六主题首屏解析和隔离 PG17.11 备份恢复、原生迁移两次均通过。2026-11-26 生产已应用 0014–0020 并部署该 Go API，健康、进程哈希、现有业务计数和 Nginx 实机预检通过。Webhook、自动补偿、历史回放保持关闭，官方订阅未变更。真实公网 ACK、实际订阅投递及客户端授权会话验收待推进。2026-11-26 15:55（北京时间）负责人批准推进 D1（仅订单、货件通知）；生产 Supabase 只读复核 9 张通知表均存在且启用 RLS、1 个有效授权、收件箱/作业/回放及发现断点均为 0。16:00 已启用生产 API 和 Nginx，仅订单与货件。健康和来源限制通过。官方应用编辑要求负责人二次验证，订阅未变更；实际通知接收、ACK、持久化、Worker 消费、去重和业务更新 NOT_RUN。自动补偿、历史回放与其他四主题关闭。负责人已核对官方应用设置（未独立验证保存结果）；2026-11-26 19:47（北京）生产只读收件箱0/作业0/订单73/回放0，真实投递仍未证实。20:16（北京）再次生产只读：收件箱0/作业0/订单73/回放0；本地无16:00后新建订单及官方订单更新时间（仅1条本地更新时间变化），不能排除漏单；仓库入口模板 access_log off，实际生产日志覆盖待现场核实。下一步仅检查实际入口/Nginx错误及脱敏ACK日志、订单和货件 missed_feeds 首屏及16:00后状态/耗时/URL匹配，不重复修改订阅、不伪造通知、不执行回放；20:21（北京）现场代理报告：生产入口、健康、DNS直连及官方订单/货件 missed_feeds 首屏查询 PASS；订单 null/0、货件5条均早于16:00；Inbox0/Jobs0/Orders73/Replay0；Nginx access_log off 已确认，16:00:05/06 两条 ACL 403 未归因（需核对启用时负向烟测），真实通知端到端 NOT_RUN。20:29现场报告五字段独立脱敏日志上线 PASS，仅 reload Nginx；配置轮转 PASS（首次实际轮转待观察），403高置信对应负向烟测但未逐条确认；负责人已批准日志验收，D1转真实事件观察，仍 NOT_RUN；后续 D2 仅订单/货件自动补偿，按现有30/60分钟、观察真实队列/429/恢复，D2于20:40已启用且首轮PASS；D1 2026-11-26 23:57:25.198（北京）首条 marketplace_shipments webhook 收据实际入库，快照 checked_at 23:57:25.970、订单缓存更新23:57:28.246，状态 shipped，匹配1订单、无残留作业/同资源后续补偿；11-27 00:23数据库侧真实货件消费与业务更新复核 PASS；11-27 复核对应五字段日志：HTTP/上游均 503，均耗时 351ms，D1 HTTP 200 ACK 门禁未通过。01:14负责人要求直接推进；[#117](https://github.com/Fan6512/una-seller/pull/117) 已 Squash Merge [`dc431e74`](https://github.com/Fan6512/una-seller/commit/dc431e74a9914ac5ebd4710310f7a22ebe4bc8ae)（隔离实测 `6339e85d`，最终审核 `58fbc371`），内部350→400ms、合并同步提交设置与准入锁SQL往返、实时持久化分阶段脱敏诊断，新增隔离PG准入锁超时/重试/去重覆盖，保持持久提交后才200及原授权/容量/并发；[CI 45866279728](https://github.com/Fan6512/una-seller/actions/runs/45866279728)PASS；[最终 CI 45916458457](https://github.com/Fan6512/una-seller/actions/runs/45916458457) Go/Changed PASS，PG停机503与重启读回PASS，[现场验收方案](../architecture/WH-06-ACK-LATENCY.md)，实时授权、容量去重及收件/作业合并为一次参数化SQL，保留独立准入锁及同步提交。隔离约59ms RTT现场正常12/12最大239.2ms、重复12/12最大239.7ms、两路8/8最大362ms，锁失败无入库且解锁重试PASS；冷连接仍503。[最终复审](https://github.com/Fan6512/una-seller/pull/117#issuecomment-7207330049) PASS；非阻断收口仅细化pgx超时/SQLSTATE57014脱敏分类，[最终 CI 45923741701](https://github.com/Fan6512/una-seller/actions/runs/45923741701) PASS；负责人批准后已合并，正式部署待批准；联合#115/116/117构建、隔离迁移权限预检现场PASS；生产ACK仍未通过，不认定原503具体根因。订单主题、真实重复投递和 429 恢复仍待样本。[夜间回执](../architecture/NIGHT-PREFLIGHT-20261127.md)。[部署报告](../architecture/WH-06-PRODUCTION-DEPLOY-20261126.md)；2026-11-27 负责人转述本地联合交付回执（V070-JOINT-DELIVERY-20261127.md，本次未直接读取原文件） | V0.7 P0 |
| OPS-02 | V0.7 运营通知中心：默认运营首页、全局铃铛、店铺/站点筛选、待办统计与最新动态；订单/发货/SLA、问答/消息、纠纷、商品/物流/授权异常，紧急事项提醒 | #105 已合并并随 V0.7.0 交付，待真实待办/消息及负责人 Windows GUI 核对；不重复开发。 | 原范围/前置及阶段记录：待开发；真实统计与只读详情接口，通知已读与业务已处理分离，跳转对应模块；完整业务操作 UI 后置；尚无数据的状态明确不可用，不显示虚假零值；关联与历史证据：[#105](https://github.com/Fan6512/una-seller/pull/105) 已 Squash Merge `59472ef3`（批准 HEAD `9bc1eabd`），CI 45765732968 Go/Desktop/Changed PASS；默认运营首页、待办动态及筛选、问答未回答待办、Pack 会话后续回复清待办、新消息重开；外部回复需同步，真实消息/Windows NOT_RUN，统一联调；[V0.7.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.7.0) | V0.7 P1 |
| DESK-05 | 桌面统一状态动效与体验：出现、消失、按压、选择、数字、图片、加载/成功/失败、弹窗 | #114 已合并，V0.7.0 签名测试版已发布；原实机反馈及最后一轮复验豁免保留，日常使用遇实际问题再定点修复。 | 原范围/前置及阶段记录：状态审计、共享节奏/曲线、减少动态效果、无布局动画/逐行动画；Windows WebView2 真机性能、窄宽屏及深色主题待验收；关联与历史证据：[#102](https://github.com/Fan6512/una-seller/pull/102) 原 Draft `819ec606` 已关闭 已同步 main 并解决 App.tsx 冲突，保留 OPS-02 和订单可见性；纠正一级页面条件卸载文档；[CI 45782202089](https://github.com/Fan6512/una-seller/actions/runs/45782202089) 四项 PASS，历史候选/基线 Windows NSIS 构建 PASS（非新 HEAD）；深色语义主题缺失，Windows 实机状态/帧率/深色 NOT_RUN，待真实验收；[#114](https://github.com/Fan6512/una-seller/pull/114) 已 Squash Merge `de6a7fc3`；代码 HEAD `773cf671` [CI 45813109362](https://github.com/Fan6512/una-seller/actions/runs/45813109362) 四项 PASS，最终 `b521a4e7` 仅同步台账；统一 Windows QA `30fbee91` 负责人确认8项通过并反馈分批旧提示，最终补丁已修，最后一轮实机 NOT_RUN（负责人明确豁免），未发布 Release，暂停客户端 UI 优化；[V0.7.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.7.0) | V0.7 P1 |
| DESK-06 | **桌面全局操作反馈（Ding）**：右侧业务内容区顶部居中单胶囊、物理弹簧原地形变、统一 success/error/warning/info/loading 五态及 `notify`/`batch`/`task` 公共入口；**Batch & Counter 与 Lifecycle & State 均为本功能 P0**：同一批次计数/真实进度/成功失败跳过汇总，异步准备→执行→异常→安全重试→终态原地演进；连续同类操作合并，独立任务/店铺隔离，旧回调不得覆盖新提示 | #114 已合并，V0.7.0 签名测试版已发布；原实机反馈及最后一轮复验豁免保留，日常使用遇实际问题再定点修复。 | 原范围/前置及阶段记录：代码已实现；固定 Ding React/Core 0.0.9（MIT 及真实锁文件已提交，最终实机复验已豁免）；通过轻量适配层隔离第三方 API，业务只调用 notify/batch/task，Ding 仅渲染；上游更新走独立依赖 PR 与现有 CI、必要实机验收，不自动合并、不复制源码、不新增更新服务，与 DESK-05 #102 共享主题/减弱动画规范但独立实施；使用单胶囊 replace 模式，后台任务真实状态与重要错误仍保留在业务页面/结果，未知进度不造百分比，重试/忽略/撤销仅在真实业务支持时提供（未知刊登结果禁止盲重试）。先覆盖订单同步、复制运单、工作稿保存及已有采集/批量图片场景，新增业务统一复用；提示音可选默认关闭（本功能 P1）；不新增后端/API/数据库/队列/历史消息中心，不替代 OPS-02，不改表单校验及危险操作确认。验收：现有 CI/构建和最小必要 Windows 实测，连续操作不堆叠、批次进度与结果真实、生命周期切换、并发不串任务且重要错误不丢失；关联与历史证据：[Ding](https://github.com/YusufCeng1z/ding)；相关 [#102](https://github.com/Fan6512/una-seller/pull/102)；[#112](https://github.com/Fan6512/una-seller/pull/112) 原 Draft `57c705ee` 已关闭，CI [45791652228](https://github.com/Fan6512/una-seller/actions/runs/45791652228) PASS；旧 `351ec5fc` Windows QA [45788612243](https://github.com/Fan6512/una-seller/actions/runs/45788612243) 实机 FAIL（卡顿、旧提示重复、遮挡顶栏）；源码修复已提交，最终复验已豁免；[#114](https://github.com/Fan6512/una-seller/pull/114) 已 Squash Merge `de6a7fc3`；代码 HEAD `773cf671` [CI 45813109362](https://github.com/Fan6512/una-seller/actions/runs/45813109362) 四项 PASS，最终 `b521a4e7` 仅同步台账；统一 Windows QA `30fbee91` 负责人确认8项通过并反馈分批旧提示，最终补丁已修，最后一轮实机 NOT_RUN（负责人明确豁免），未发布 Release，暂停客户端 UI 优化；[V0.7.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.7.0) | **P1（全局）**；内部 P0 两模式同时交付，不阻断首次 CBT 刊登，版本待排期 |
| DESK-07 | 桌面主导航改为全高左侧竖栏：运营中心、工作台、已刊登商品、订单管理、售后管理；左下个人中心，右邻设置图标；默认运营中心 | #114 已合并，V0.7.0 签名测试版已发布；原实机反馈及最后一轮复验豁免保留，日常使用遇实际问题再定点修复。 | 原范围/前置及阶段记录：Apple HIG Sidebar 风格、统一18px线性图标/38px行高/184px宽、浅蓝灰选中态、紧凑窄屏；删除整条顶部重复状态栏，正常静默，实际服务/店铺连接异常非阻塞弹窗，每次异常可关闭、恢复后新异常再提醒，保留原健康检查和店铺管理入口，不把主动断开或普通列表错误误报为故障；运营中心上总览/中待办/下动态，今日金额依赖 OPS-05/FIN-02。售后仅真实未开放占位，不提前实现 CS-01/02/03；独立壳层、不新增依赖/玻璃特效；与 #102/#112 合流保留动效与反馈隔离，Ding 在右侧内容区居中、不再依赖已删除的 topbar；关联与历史证据：[#113](https://github.com/Fan6512/una-seller/pull/113) 原 Draft `ea6b33bc` 已关闭，侧栏、状态区与运营首页布局代码已提交；独立 CI [45796202107](https://github.com/Fan6512/una-seller/actions/runs/45796202107) PASS，未实现今日真实金额；[#114](https://github.com/Fan6512/una-seller/pull/114) 已 Squash Merge `de6a7fc3`；代码 HEAD `773cf671` [CI 45813109362](https://github.com/Fan6512/una-seller/actions/runs/45813109362) 四项 PASS，最终 `b521a4e7` 仅同步台账；统一 Windows QA `30fbee91` 负责人确认8项通过并反馈分批旧提示，最终补丁已修，最后一轮实机 NOT_RUN（负责人明确豁免），未发布 Release，暂停客户端 UI 优化；[V0.7.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.7.0) | V0.7 P1 |
| COL-02 | 采集图片高清化、去重、视频过滤、部分成功原因与变体完整性 | 复杂商品采集质量、高清图、规格归属及速度待完整实测。 | 原阶段记录：代码已有；#91 插件采集按钮布局修复已合并，负责人反馈正常；完整新版实测未完成；关联：[#19](https://github.com/Fan6512/una-seller/pull/19)、[#20](https://github.com/Fan6512/una-seller/pull/20)、[#91](https://github.com/Fan6512/una-seller/pull/91) / [#6](https://github.com/Fan6512/una-seller/issues/6)；验收判据：选真实复杂商品，核实图片质量、规格归属、采集和切换速度 | 原阶段未单列 |
| WB-01 | 采集箱整理工作稿：编辑、保存、并发冲突与重启恢复 | 新版工作稿保存及重启读回 GUI 待验收，保留已有通过范围。；保留 #76 及 Issue #75 的 2026-11-23 负责人核心采集→整理→保存→重启读回 PASS；仅对后续变更范围定点验收，不重复旧版通过项 | 原阶段记录：代码已合并；#92 工作稿兼容修订与自动化 QA PASS，生产读取与历史数据比对 PASS（现场报告）， #94 隐藏普通历史描述（保留历史数据与超长恢复入口），保存/GUI 待验收；关联：[#52](https://github.com/Fan6512/una-seller/pull/52)、[#59](https://github.com/Fan6512/una-seller/pull/59)、[#92](https://github.com/Fan6512/una-seller/pull/92)、[#94](https://github.com/Fan6512/una-seller/pull/94)；验收判据：同一真实来源保存、重启读回，不覆盖原始采集 | 原阶段未单列 |
| WB-02 | 独立产品箱、勾选转库、来源追溯和再次编辑 | 转库、再次编辑和重启读回的新版实机验收。；保留 #76 及 Issue #75 的 2026-11-23 负责人转库→编辑→重启读回 PASS；仅验收后续变更与实际缺陷 | 原阶段记录：代码已合并；#92 保存/转库字段保护回归 PASS，生产读取与历史数据比对 PASS（现场报告），实际转库/Windows GUI 待验收；关联：[#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65)、[#92](https://github.com/Fan6512/una-seller/pull/92) / [#57](https://github.com/Fan6512/una-seller/issues/57)；验收判据：工作稿 → 产品箱 → 编辑 → 重启读回；不得重复建产品或丢字段 | 原阶段未单列 |
| WB-03 | 语言标签、图片墙、变体独立编辑、重量尺寸及商品/变体 CNY 成本、USD 目标刊登收益及副图批量复制 | 真实多规格商品的主副图、独立 SKU/成本/目标收益等新版 GUI 核对。 | 原阶段记录：V0.5 #92 已合并；CI 与 Windows QA PASS，生产读取与历史数据比对 PASS（现场报告），#94 逐变体就地图片选择及独立 SKU/成本/目标收益已合并，CI PASS；新版 GUI 待验收；关联：[#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65)、[#92](https://github.com/Fan6512/una-seller/pull/92)、[#94](https://github.com/Fan6512/una-seller/pull/94)；验收判据：用真实多变体产品核实编辑、保存、主图、图片顺序、主图保留/副图复制、SKU 对应、独立成本及目标收益；旧客户端未变重复 SKU/无名变体可安全重排，歧义返回 422；目标刊登收益不等于实际结算 | 原阶段未单列 |
| CAT-01 | 官方 CBT 类目预测、分类必填属性和技术规格读取 | 两种官方类目、枚举、单位及切换后历史字段核对；真实刊登属 CBT-02。 | 原阶段记录：代码已合并；#92 规格 ID/必填提示回归 PASS，#94 根据类目动态提示 MODEL 必填/待核实状态，CI PASS；完整客户端实测待完成；关联：[#31](https://github.com/Fan6512/una-seller/pull/31)、[#68](https://github.com/Fan6512/una-seller/pull/68)、[#92](https://github.com/Fan6512/una-seller/pull/92)、[#94](https://github.com/Fan6512/una-seller/pull/94)；验收判据：选择两种官方类目，核对枚举、单位、必填与切换类目后的已保存字段；SKU-06/07 真实 Family 与实际刊登 payload NOT_RUN，继续作为刊登门禁 | 原阶段未单列 |
| ACC-01 | 设置 → 账号管理：登录邮箱、修改密码及必要再验证 | #78 已合并；真实 Supabase 改密/再认证及 Windows GUI 尚无 PASS 证据，集中实测 | [Issue #75](https://github.com/Fan6512/una-seller/issues/75)、[#78](https://github.com/Fan6512/una-seller/pull/78)（合并 `fe287414`）；沿用 Supabase Auth，不保存密码，改密后清除会话；原 [#76](https://github.com/Fan6512/una-seller/pull/76) 需求登记 | 原 P0（集中 QA 前），不阻塞已批准的联合升级 |
| CRED-01 | 用户云端 AI/翻译服务绑定：保存、替换、删除、显式测试自有 Key | #77/#78 已合并；0011 已生产迁移，真实服务端启用、用户 Key 绑定及小样本调用仍待确认，不假报 PASS | [Issue #75](https://github.com/Fan6512/una-seller/issues/75)、[#77](https://github.com/Fan6512/una-seller/pull/77)（合并 `ee3fd5b3`）、[#78](https://github.com/Fan6512/una-seller/pull/78)（合并 `fe287414`）；owner 隔离、私有加密存储、受控 HTTPS 与按账号调用，不回传原 Key；原 [#76](https://github.com/Fan6512/una-seller/pull/76) 需求登记；不擅自付费批量试译 | 原 P0（真实 AI 验证前） |
| TR-01 | 订单标题、类目、属性、枚举等中文阅读（Azure/百度/火山通用择一） | 默认 Provider 已切 `baidu` 并随 `2fc79081` 部署；百度真实调用可用，EN/ES/PT 中文阅读接口仍待登录实机验证。 | 原阶段记录：代码已合并；真实翻译尚未验收；关联：[#67](https://github.com/Fan6512/una-seller/pull/67)、[#70](https://github.com/Fan6512/una-seller/pull/70)；验收判据：服务端配置合规凭据，测 EN/ES/PT、回退及原文/官方 ID 保留；2026-11-27 负责人补充已刊登详情中文阅读：分类路径双语：官方原文为主（CBT 英文优先，站点原文可能为西/葡语），中文为下方辅助参考；属性组/字段名称、官方枚举选项中文优先；品牌/型号/SKU/商品编码、自由文本、标题、描述及保存值保留原文。#120 已合并 `84efeabd`（审核 `5870d3cb`） 复用现有账号翻译服务、补足站点元数据及少量常用术语本地回退；未配置翻译或缺译时保留官方原文，不伪造中文覆盖率。原始 ID、枚举值、草稿及官方提交合同不变；候选 CI 与 Windows NSIS 构建 PASS；合并后 CI、生产交付及真实西/葡语和服务配置待验收 | 原阶段未单列 |
TR-02 | CBT 全量分类中文离线词典、增量更新与人工覆盖 | TR-02E 完整路径词典已正式发布：`active.json`=`b3b9f46d…`，中文 10,928/10,954（99.76%），115 条人工覆盖、26 条回退英文；普通 updater 的 managed 保护已实测联网前阻断。**尚未完成：每日增量维护流程**，timer 保持关闭。下一步单独实现只处理新增/变更官方类目、额度预检、staging、差异与质量检查、受控原子发布及回滚；旧 updater 不得覆盖正式词典，验收及单独批准前不启用 timer。 | 原阶段记录：代码已合并；中文 0/10,954；关联：[#72](https://github.com/Fan6512/una-seller/pull/72)；验收判据：先五批真实试译、核对质量与额度，随后断点续跑并记录覆盖率；不因没有密钥假报 PASS；2026-11-27 负责人重申类目全量下载、后台预翻译、增量更新；选择/推荐时直接查已发布词典，官方英文为主中文辅助，缺译回退原文，不在该交互中临时翻译类目路径。现有 CBT 预测仍用官方推荐接口与词典 Localize，不能把离线译名当作官方动态属性/可刊登权限；生产同一服务账号使用独立维护翻译凭据（不消费用户 Key），核对 active/staging/progress 及定时任务、实际覆盖率和无运行时类目翻译调用。各站点类目路径的共享离线覆盖为后续扩展，#120 当前站点详情路径在线翻译不等于 CBT 推荐路径，需单独收口，属性枚举暂不做全量预取。 | 原阶段未单列
| TR-02A | 公共词典月额度保护（百度优先） | [#125](https://github.com/Fan6512/una-seller/pull/125) `a8579e98` 已合并，固定 `f644d353` 联合 CI PASS，已部署；90 万上限已设置，原计数 3,118 保留，API/CLI 同一 `baidu-usage.json` 已复核；五批扣额 3,958，暂停计数 192,758。负责人核对控制台实耗191,828，较本地192,758少930；保留高值，当前对账PASS，最终运行验收待办；火山额度另待开发。 | 负责人确认百度通用剩余997211字符，部署验收后设置 BAIDU_TRANSLATE_MONTHLY_CHAR_LIMIT=900000；代码通用默认50000保留，避免标准版超额。已有 baidu-usage.json 不清零，旧重试未计量需与控制台对账；不自动切换或重试超额请求，不新增依赖/迁移；火山额度另待开发。 | P0（TR-02 全量前置） |
| TR-02C | 公共类目电商语境、术语及预算预检 | [#126](https://github.com/Fan6512/una-seller/pull/126) `f644d353` 已合并并部署；父级语境、保守术语与全路径覆盖、`-plan` 验证通过；旧 96 条、五批 80 条及抽检 200 条复核完成，共 64 条全路径覆盖。200 条抽检 89.0%，未达 95%，全量暂停，不标记完成。 | 全量前备份并复核已试译94条，正确译文保留，错误/歧义按实际官方完整路径校对；不全局替换 Motors/Fans/Turntables 等歧义词。先按真实目录重算语境输入字符，核对额度，再五批试译与原有样本对照；质量及额度通过后才全量与发布，保留官方原文及ID。大模型文本测试额度剩余约 98.6 万字符；质量校正已由 TR-02D 调用并完成 200 条对照。 | P0（TR-02 全量前置） |
| TR-02D | 百度大模型文本翻译：公共电商词典质量校正 | [#127](https://github.com/Fan6512/una-seller/pull/127) `4c387e7c`、[#128](https://github.com/Fan6512/una-seller/pull/128) `83b505fa` 均已合并并部署；首轮 1/5/200 条与第二轮独立 200 条受限试译完成；候选 400 条独立暂存。独立样本 190/200（95.0%，CI 91.0%–97.3%），已有译文 68/73、未译文 122/127；10 条叶子级语义错误（0 跨一级行业），零错误门禁未过。六条授权纠错已写入，overrides 70 条；186 字符差额追溯为 CBT106855 两次 93 字符被拒计费。本地大模型 27,144，控制台预计 27,041；全量预算完整路径 818,169 / 父级 332,696。未发布候选、未扩 5 万上限、timer 未启用。 负责人2026-11-27批准后续TR-02E：10条纠错、调整未来质量门禁、完整路径全量与90万受限执行器开发，保留最终真实发布验收；历史严格零错误门禁不改判。 | 负责人确认大模型文本翻译剩余100万测试字符、领域文本翻译50万；独立累计90万为未来绝对上界，本版本仅开放累计5万试验硬上限，不按月重置、不自动切换；负责人提供[百度开放平台接入文档](https://fanyi-api.baidu.com/product/133)官方curl：固定POST /ait/api/aiTextTranslate，Bearer + JSON appid/from/to/q；独立 BAIDU_LLM_API_KEY，APP ID 可复用。首轮每条独立请求、1秒间隔、无重试，累计5万硬限制；官方返回及指令字段已实现；术语库未配置暂不开启，真实响应与本轮 q 字符计费已核对，不混用百度智能云或文档接口。复用现有锁/批处理/进度，独立候选校正文件，64条人工覆盖优先。先200条原样本盲评对照，再200条独立分层样本≥95%且无明显分类语义错误；预算足够才校正历史及补余量，独立验收后原子发布，timer仍暂停。2026-11-27 已部署并完成两轮受限试译；第二轮详见 [报告](../architecture/TR-02D-SECOND-ROUND-ACCEPTANCE-20261127.md)。 | P0（TR-02 质量门禁前置） |
| AI-01 | 根据产品事实生成 EN/MX/BR 三语 SEO 标题 | 真实 AI 三语标题、字数与人工采纳验证。 | 原阶段记录：代码已合并；真实 AI 调用待验收；关联：[#71](https://github.com/Fan6512/una-seller/pull/71)；验收判据：使用真实保存的商品生成、分别采纳；MX/BR 不超过 60 字符且不得编造属性 | 原阶段未单列 |
| AI-02 | 英文卖点与详情独立采纳，按母版翻译 ES/PT，译文过期提醒 | 真实描述生成/翻译、过期提示和重启读回验证。 | 原阶段记录：代码已合并；真实服务与 Windows 读回待验收；关联：[#71](https://github.com/Fan6512/una-seller/pull/71)；验收判据：从已保存英文描述翻译；人工核对数字/单位，手改后检测过期，重启读回 | 原阶段未单列 |
| PUB-01 | 已刊登商品只读同步、站点/SKU 展示、基础本地编辑草稿 | 已刊登同步、详情/描述/规格完整性与性能验收。 | 原阶段记录：代码已合并；真实性能与完整功能验收不齐；关联：[#27](https://github.com/Fan6512/una-seller/pull/27)、[#30](https://github.com/Fan6512/una-seller/pull/30)、[#31](https://github.com/Fan6512/una-seller/pull/31)、[#34](https://github.com/Fan6512/una-seller/pull/34) / [#26](https://github.com/Fan6512/una-seller/issues/26)；验收判据：已授权账号真实同步、详情、描述、规格；本地编辑不等于官方写回；2026-11-27 新账号实际商品同步 PASS，生产库新账号已刊登缓存 0→50、原账号保持 207；未证明全量覆盖、详情/描述/规格完整性 | 原阶段未单列 |
| PUB-02 | 已刊登同步偶发 504、首屏慢请求 | 偶发504只在复现时采集同请求分段证据并修复，不先重构。 | 原阶段记录：诊断与优化代码已有；异常根因未最终验收；关联：[#35](https://github.com/Fan6512/una-seller/pull/35)、[#36](https://github.com/Fan6512/una-seller/pull/36)、[#37](https://github.com/Fan6512/una-seller/pull/37) / [#38](https://github.com/Fan6512/una-seller/issues/38)、[#41](https://github.com/Fan6512/una-seller/issues/41)；验收判据：实际复现时取得 Go/代理/客户端同次脱敏分段日志再定点修复 | 原阶段未单列 |
| ORD-01 | 订单只读同步、列表详情、跨站点/跨店铺搜索 | 新版金额、编号、图片、筛选及实际新增订单 GUI 验收；历史通过范围保留。 | 原阶段记录：V0.4 部分通过；#90 修订已合并，生产抽测 USD 销售额、订单关联、履约筛选 PASS（现场报告）；V0.5 新订单主图按负责人意见暂通过（实际新增订单案例 NOT_RUN）；其余 GUI 按实测登记；关联：[#40](https://github.com/Fan6512/una-seller/pull/40)、[#43](https://github.com/Fan6512/una-seller/pull/43)、[#55](https://github.com/Fan6512/una-seller/pull/55)、[#90](https://github.com/Fan6512/una-seller/pull/90)；验收判据：V0.4 同步/分页/筛选与跨店跨站搜索 PASS。#90 FIX-01/02/04：明确 USD 商品行汇总、Pack 销售编号与内部 Order ID、按需自动补图；重复疑似行待核实。生产只读核查 73 单/73 行，0 重复候选，负责人报告 73/73 Billing 美元金额一致；无回填。V0.5 实测金额、编号、补图、筛选及详情；2026-11-27 双账号订单读取与切换 PASS；生产库新账号订单 5→14、原账号保持 73，未验证新增订单全部详情及具体同步来源 | 原阶段未单列 |
| ORD-02 | 1688 采购订单号、成本、国内运单**手工关联** | 1688 采购号、CNY成本和国内运单手工录入保存/隔离验收。 | 原阶段记录：V0.4 官方/采购物流分开展示 PASS；采购录入未验收；关联：[#46](https://github.com/Fan6512/una-seller/pull/46)、[#54](https://github.com/Fan6512/una-seller/pull/54)；验收判据：采购订单号、成本及国内运单手工保存仍待验收，确认跨店不串单；**未自动读取 1688** | 原阶段未单列 |
| ORD-03 | 官方发货 SLA 与单笔订单状态刷新 | 实际 SLA、退款状态及单笔刷新后列表一致性。 | 原阶段记录：V0.4 单笔订单状态刷新 PASS；#90 SLA/履约筛选修订已合并，生产已有 SLA 读取及履约筛选 PASS（现场报告），实时刷新/GUI 待验收；关联：[#48](https://github.com/Fan6512/una-seller/pull/48)、[#90](https://github.com/Fan6512/una-seller/pull/90)；验收判据：FIX-07/08 已合并：绝对截止日期及精确倒计时，综合履约状态与服务端筛选；实时物流与 SLA 刷新会写入生产缓存，本轮 NOT_RUN；V0.5 核对实际 SLA、退款状态、单笔刷新后筛选列表一致性 | 原阶段未单列 |
| ORD-04 | 官方 Shipment 发货/运输/签收及运单追踪 | 新版官方查询链接、签收时间、物流刷新等 GUI 剩余验收。 | 原阶段记录：**生产后端 PASS；V0.4 Windows 9 PASS / 2 FAIL / 1 待验收；#90 生产已升级、已有详情只读验证 PASS（现场报告）**；关联：[#66](https://github.com/Fan6512/una-seller/pull/66)、[#90](https://github.com/Fan6512/una-seller/pull/90)、[#94](https://github.com/Fan6512/una-seller/pull/94)、[#95](https://github.com/Fan6512/una-seller/pull/95)；验收判据：2026-11-24 生产 0010、API 823e6056、9 单真实物流与完整缓存 PASS；[脱敏验收摘要](../architecture/ORD-04-DEPLOY-REPORT-20261124.md)。V0.4 已发布为 Pre-release（[下载](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.4.0)），物流查询链接打不开、签收时间显示与后台相差 11 小时（旧版问题）；V0.5 FIX-05/06 已合并：受限网页弹窗（外部浏览器备用），时间统一北京时间；FIX-08：签收等综合履约状态。V0.5 生产已有物流详情/SLA 只读验证 PASS；实时刷新会写生产缓存，本轮 NOT_RUN；V0.5 负责人反馈实际打开浏览器新标签且查询站点不可用，物流窗口部分通过；#94 新增 Track718/17TRACK、复制运单与明确独立窗口/浏览器选择，CI PASS；第三方网站独立窗口实机兼容性、北京时间、状态及单笔刷新待复验。列表状态、运单复制、承运商、发货时间、时间轴、批量更新与缓存保留 PASS；#95 针对实机发现订单交易已同步但物流仍为旧缓存：单笔联动、每次显式订单同步后最多10笔陈旧非终态物流（15分钟间隔）且不读取历史，候选 CI [45558640026](https://github.com/Fan6512/una-seller/actions/runs/45558640026) PASS、已合并；2026-11-25 15:33:11（北京时间）生产升级至 `2f679dda` 后**生产后端 PASS**：单笔订单刷新联动官方物流（缺失快照 → `delivered` 落库）、候选列表 10 笔上限、`has_more`、15分钟去重与终态排除现场 PASS，失败保留与 429 冷却由定点测试 PASS（现场 NOT_RUN）；Windows GUI 联动提示、第三方物流窗口及时间显示待人工复验，正式 Release 待审批；2026-11-25 负责人实机反馈：一笔待发货订单恢复为正确已发货，状态更新结果 PASS，具体触发按钮未辨明；第三方网页仍打开默认地址、列表状态重复，需修复主查询入口（Track718/17TRACK 指定网页）、去掉重复状态，保留详情原始状态；不接第三方 API；[#96](https://github.com/Fan6512/una-seller/pull/96) 候选 CI 45571662671 Changed Files/桌面 PASS，本地代码复审无阻塞，负责人批准后已 Squash Merge 至 `415c782a`，Windows GUI 待验收 | 原阶段未单列 |
| FIN-01 | 官方 CBT 账单 A1 只读预览、费用/冲正明细 | #97/0013 生产账单持久化、单笔缓存及重启读回已 PASS，负责人实机该项 PASS；余下账单字段、退款/共享费用对账及真实429验证；利润接线归 FIN-02。 | 原阶段记录：**#49 已合并；生产账单读取抽测 PASS（现场报告）；V0.5 实机报告账单限流；#94 冷却共享、Retry-After CORS 与脱敏来源诊断已合并，生产升级及单笔账单读取 PASS（现场报告），真实429来源/冷却显示 NOT_RUN；net_income_usd=null；真实净收益与独立费用对账待验收**；关联：[#49](https://github.com/Fan6512/una-seller/pull/49)、[#94](https://github.com/Fan6512/una-seller/pull/94)；验收判据：真实只读取证：负责人报告 73/73 账单返回；缺失/越权 200 空结果为 not_available，非 200 不猜测。仅展示已核实 USD 费用，采购 CNY；货件费用不跨订单累加，包装 429 冷却，保留原币数据。官方净收益、实际到账、退款及共享货件分摊未完成（net_income_usd=null）；A2 利润未接线。真实 Seller Center 回款核对及 Windows 验收待执行；2026-11-25 负责人实机单笔官方账单读取 PASS；账单未持久化、官方总计仍待核账。P1 补齐按授权店铺/站点/订单隔离的持久化快照、默认读缓存、显式刷新与失败保留，核实官方预计净回款字段；不得将货件净成本重复扣减或把局部小计冒充净回款；[#97](https://github.com/Fan6512/una-seller/pull/97) A2 已提交迁移 0013、单笔快照存取及桌面读取，HEAD `f16bcf3` 候选 CI [45587481720](https://github.com/Fan6512/una-seller/actions/runs/45587481720) 四项 PASS，PG 重启 seed/restart/verify 均实际 PASS；据本地验收报告，隔离 PG17.12 重复迁移、重启读回、授权/货件隔离、并发、缺失/失败保留及 429 冷却 PASS；#96 已合并，#97 已以普通 merge commit `a2570f61` 纳入 main `415c782a`，保留历史；相对旧 HEAD 仅合入两处桌面文件及台账，后端/迁移/测试工作流不变；新 HEAD `a2570f61` CI [45588642857](https://github.com/Fan6512/una-seller/actions/runs/45588642857) 四项 PASS，PG 写入/重启/读回三步实际 PASS，负责人批准后已 Squash Merge 至 `0064a239`，Windows 待验收；2026-11-25 生产迁移 0013 与后端 `0064a239` 已部署，单笔真实订单 `2000019508270206` 完成官方读取、PG 持久化、billing-cache 读回及服务重启后读回 PASS，重复缓存读取未触发官方请求；真实净回款 NOT_RUN（见 V0.6 部署报告）；2026-11-25 负责人确认账单已能持久化，实机该项 PASS，其余专项暂缓，真实净回款继续待 FIN-02 核账；2026-11-27 新账号账单快照 0→1、原账号保持 3，仅证明缓存存在，不代表金额、净回款或退款对账通过 | 原阶段未单列 |
| WH-03 | V0.7 订单与物流通知处理器：新订单入库、Pack 关联、交易与货件状态更新，支持货件先到 | 0014/0015 已部署，订单/货件主题已启用；货件真实消费与缓存更新已核实，订单自然通知、多订单归属、共享限流实际429仍待样本；ACK 归 WH-06。 | 原范围/前置及阶段记录：OPS-01；复用 ORD-01/03/04 的解析、授权、更新时间保护和人工/后台公共协调器；精确站点/物流模式/子卖家归属；关联与历史证据：[#100](https://github.com/Fan6512/una-seller/pull/100) 已 Squash Merge 至 `32067017`，审核 HEAD `76906f4`，最终复审 APPROVE，负责人批准合并（默认禁用）；合并后 CI 待复验。[CI 45662044591](https://github.com/Fan6512/una-seller/actions/runs/45662044591) 四项、PG 真停机 ACK 503/重启恢复实际 PASS；据本地隔离 PG17.6 报告全量测试及边界变异验证 PASS。0015 公共协调、人工优先、延迟退还额度、指数退避及身份冲突死信；真实 CBT GET 货件 sender_id=0、无 site_id、含 order_id，订单 context.site/子卖家，新版关联接口 401（具体原因未证实）；无权限/空关联兜底、同范围已知订单及晚到订单快照，30 单 PASS、31 单整笔回滚、跨范围不写。按负责人反馈保留 30 单上限、取消分批续跑；未知关联/本地滞后及真实多订单仍需 OPS-03/WH-06 补偿联调；连接池压测/调整、0014/0015 迁移、真实公网、来源校验和订阅为启用门禁；索引/极端死锁后续优化。未生产迁移/部署/启用 | V0.7 P0 |
| WH-04 | V0.7 售前问答及售后消息只读后端：问题/商品、会话/消息原文、时间、官方待回答/未读状态，按店铺站点查询 | 后端已随生产基线部署，问答/消息主题未启用；真实消息与未答问答取证后另行审批启用，不自动回复或标记官方已读。 | 原范围/前置及阶段记录：OPS-01；官方资源路径、卖家类型与读取权限逐类验证；不发送回复、不自动标记官方消息已读；关联与历史证据：[#101](https://github.com/Fan6512/una-seller/pull/101) 已 Squash Merge 至 `8b8aa1a6`，审核 HEAD `e855583`，负责人批准；合并后 CI 待复验；[CI 45736784362](https://github.com/Fan6512/una-seller/actions/runs/45736784362) Go/PG 真停机重启 PASS，Desktop/MV3 SKIP；显式兼容数字/字符串消息 ID，保留数字与归属校验（旧 json.Number 也接受合法带引号数字，未证实原 P1 故障）；真实问答 GET 200，消息真实投影/已读变化及 UNANSWERED 样本 NOT_RUN，消息主题禁止启用，生产门禁未验 | V0.7 P1 |
| WH-05 | V0.7 投诉纠纷及站点商品通知只读后端：纠纷状态、类型、处理期限与允许操作；商品价格、库存、审核、上下架状态回读 | 后端已随生产基线部署，纠纷/商品主题未启用；真实开放纠纷与商品通知验证后另行批准启用，不自动执行官方动作。 | 原范围/前置及阶段记录：OPS-01；真实权限与资源归属验证；复用已刊登商品数据，不自动退款、处理纠纷或写回商品；关联与历史证据：[#103](https://github.com/Fan6512/una-seller/pull/103) 已 Squash Merge 至 `1f69d14d`，审核 HEAD `4d59dcdb`，负责人批准；合并后 CI 待复验；[CI 45740159214](https://github.com/Fan6512/una-seller/actions/runs/45740159214) Go/PG 真停机重启 PASS，Desktop/MV3 SKIP；真实 sender/receiver 投诉与站点商品 GET 已取证，卖家动作隔离/mandatory 期限、商品缺失保留/明确清空/零值/旧快照回归已验；0016/0017 幂等及本地 PG17 重启据报告 PASS，真实开放投诉期限/动作 NOT_RUN；main 仅台账差异，无冲突。默认禁用，不部署、不迁移、不启用 | V0.7 P1 |
| DESK-02 | macOS arm64/x64 原生实机登录、业务 UI 和更新验收 | 待 macOS arm64/x64 真实设备及获批构建，登录、业务 UI、OTA 验收；不把 Windows 结果算 macOS PASS。 | 原范围/前置及阶段记录：macOS 测试环境与已批准构建；关联与历史证据：[CROSS_PLATFORM](../architecture/CROSS_PLATFORM.md) | P1（正式双端交付前） |

## 五、尚待开发、暂缓与候选

### A. 待开发

只登记未完成的实施范围；有前置依赖时先满足依赖，原优先级保留，不代表现在全部启动。

| ID | 功能 | 当前状态 / 下一步 | 关联证据及历史范围 | 原优先级 |
| --- | --- | --- | --- | --- |
| FIN-02 | 区分官方预计净回款、实际结算回款及实际 USD 利润，按已核实费用/退款/汇率/采购成本计算，避免重复扣费 | FIN-02A 已联合交付（见待实测/联调）；剩余 FIN-02B 实际结算：订单至付款批次、收款流水、退款/共享费用、实际汇率及二次验证仍待取证，未开发接线。 | 历史开发记录：当前业务重点：预计利润 FIN-02A 先固定两笔 MLB USD7.90/8.99 的脱敏预计净回款字段、退款与共享费用完整性，再实现预计净回款及成本/汇率接线；实际利润 FIN-02B 仍缺订单至付款批次、收款流水及二次验证，单独保留证据门禁。[#118](https://github.com/Fan6512/una-seller/pull/118) Draft `ec3f1ba8` 第一切片已提交：独立 ConfirmedExpectedUSDNet/ExpectedOrderProfitUSD，复用精确金额算术，不伪装实收；3笔现有生产快照脱敏只读复核，两笔 pending 算式 11.28−3.38=7.90、28.13−14.50−0.51−1.65−2.48=8.99 与现场净额吻合，但 NO 扣项、官方共享归属及费用/税完整性尚未核实。原纯预计合同已推进运行接线：0024 增加订单级可选卖家中心净额/额外成本人工确认，完整 owner/店铺/子卖家范围与证据版本防串/防旧确认；缺失及退款不算利润，订单详情、OPS-05 今日人民币预计利润汇总及覆盖率已提交。自动净额可靠来源仍待核，不要求逐单录入；原候选 CI 45929353526 PASS。负责人确认按北京时间创建日锁定 USD/CNY、截取一位小数不四舍五入；新增私有0023日汇率表、可选 Go 自动采集、历史日补齐、首次成功不覆盖、CNY利润精确计算；采用前一已公布交易日ECB参考价，真实源与隔离迁移待验，生产未迁移/未启用；[核账证据](../architecture/FIN-02A-EXPECTED-PROFIT.md)。[CI 45930400368](https://github.com/Fan6512/una-seller/actions/runs/45930400368) PASS。接线候选 45931869086 Desktop PASS、Go格式 FAIL，已修订；[接线 CI 45932096096](https://github.com/Fan6512/una-seller/actions/runs/45932096096) PASS。自动净额继续取证：按官方 Billing 文档扩充已授权账单白名单 payment_id、sales_info.order_id/operation_id/transaction_amount_usd、shipping_info.pack_id，保留共享费用归属，不把交易额冒作净回款；无新迁移/额外官方请求。[最新 CI 45932642492](https://github.com/Fan6512/una-seller/actions/runs/45932642492) in_progress。下一步现场核对 Seller Center 直接订单净额字段及 MLM/共享/退款样本，并用 payment_id/operation_id 对照已有资金报告 SOURCE_ID；生产授权浏览器/SSH 不在当前会话可用范围；隔离迁移/真实汇率源和 Windows GUI 待验；不把候选公式推广到全部订单。；原范围/前置及阶段记录：FIN-01 真实账单对账、可靠汇率和采购成本；关联与历史证据：[#49](https://github.com/Fan6512/una-seller/pull/49)（A1 已合并；A2 仅纯算式、未接线）；11-27 00:21生产只读73订单、3账单缓存、0实际采购成本录入，实际净回款无已验证来源；[核账方案](../architecture/CBT-BILLING-CNY-PROFIT.md)补充 Global Selling Payment Reports 与收款流水交叉核对、订单级归属/冲正/实际汇率验证，优先验证实际样本后接线；本轮核对 3 笔 MLB 快照（1 released、2 pending）与实际账户的 11 月 4 日自动回款报告，批次金额匹配，缺订单级映射和到账流水。实际成本和汇率缺失，证据门禁未通过；明确独立预计利润（经验证的预计净回款）与实际利润（订单级已结算净回款），各自独立成本、汇率、日期及覆盖率；11-27 负责人转述现场核账：两笔MLB待释放订单卖家中心净额USD7.90/8.99，与既有账单商品额和扣项一致，预计净回款取得初步订单级样本；下一步固定订单作用域、账单字段及费用/退款完整性、确认成本与汇率，单独推进预计利润最小切片，不把样本推广到全部订单；实际结算仍缺订单至批次映射、到账流水和二次验证 | P1 |
| DESK-04 | V0.7 冷启动重新登录及缓存隔离 | 原需求未实施；按负责人既有裁决后置，待排期。 | 原范围/前置及阶段记录：待开发；保留邮箱/设置/图片缓存，ERP 会话不跨冷启动持久化；店铺授权与业务数据留在服务器，未登录不展示业务数据；完全退出后再启动需登录，最小化/切换界面无需重新登录；不保存密码，不引入离线业务库；关联与历史证据：实施 PR 待建 | V0.7 P1 |
| DESK-03 | **桌面缩略图与图片加载优化·第一阶段**：内部先做（原建议 P0）已知 CDN 小图预览/点开原图及变体图片按需渲染；随后（原建议 P1）主图优先、其他懒加载、相邻最多 1–2 张预加载、失败回退重试与脱敏耗时诊断；**P2 本地有界缩略图缓存仅作为第二阶段候选，不在本期开发**；始终保留高清原始 URL 供编辑、修图与 CBT 刊登 | Issue #83 已立项，实施 PR 待建；按需小图/懒加载等，不新增图片代理或迁移。 | 原范围/前置及阶段记录：独立桌面 Draft PR；白名单 URL 单测、失败/重试、性能对比和 Windows 真机验收；**不增加 Go 图片代理/对象存储/数据库迁移**；关联与历史证据：[Issue #83](https://github.com/Fan6512/una-seller/issues/83)（已立项；实现 PR 待创建）；数据库读写优化 [#79](https://github.com/Fan6512/una-seller/pull/79)、[#82](https://github.com/Fan6512/una-seller/pull/82) 属于不同范围 | **P1（全局）**（待开发；行内 P0/P1/P2 仅表示本功能内实施顺序，不阻断 CBT 首次刊登） |
| ORD-05 | 1688 授权与采购订单自动同步、厂家国内物流追踪 | 待合法可靠1688授权与接口取证后自动采购同步，原P2主线后置。 | 原范围/前置及阶段记录：合法可靠的 1688 授权/接口验证；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | P2，主线后置 |
| ORD-06 | 货代仓库签收/转运 → Mercado Libre 官方扫描 → 买家签收的完整双物流关联 | 待可用货代/国内物流数据源；完整双物流关联未实现。 | 原范围/前置及阶段记录：ORD-04 真实平台 Shipment + 可用货代数据源；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | P2 |
| CS-01 | V0.8 售前问答客户端：问题列表、商品关联、原文/翻译、待回答筛选、人工确认并回复 | V0.8，依赖 WH-04 真实权限；人工回复官方写入单独审批。 | 原范围/前置及阶段记录：WH-04 只读数据与真实官方权限验收；发送属官方写入，须独立授权；关联与历史证据：[Delivery](../architecture/DELIVERY.md) M3；实施 PR 待建 | V0.8 P1 |
| CS-02 | V0.8 售后消息客户端：会话列表、双语阅读、快捷话术、人工回复 | V0.8，依赖 WH-04；人工消息回复单独审批。 | 原范围/前置及阶段记录：WH-04 与真实消息权限；发送须独立批准；关联与历史证据：实施 PR 待建 | V0.8 P1 |
| CS-03 | V0.9 投诉纠纷管理客户端：纠纷详情、处理期限、官方允许操作、人工确认 | V0.9，依赖 WH-05；真实官方操作另行审批。 | 原范围/前置及阶段记录：WH-05；真实纠纷操作权限及官方写入审批；关联与历史证据：实施 PR 待建 | V0.9 P1 |
| PUB-06 | V0.9 商品运营与异常管理：审核/下架、价格库存变化与监控 | V0.9，依赖 WH-05/PUB-01/04，先只读异常与人工处理。 | 原范围/前置及阶段记录：WH-05、PUB-01/04；先只读异常和人工处理，不自动调价；关联与历史证据：实施 PR 待建 | V0.9 P1 |
| IMG-02 | 商品图片去背景、白底处理、原图/候选对比及人工选用 | 合法修图方案确认后独立实施，原P2。 | 原范围/前置及阶段记录：先确定可用的合法修图方案；不预设官方有 AI 白底接口；关联与历史证据：[产品定义](PRODUCT-WORKFLOW-AND-EDITOR.md) M1-C | P2 |
| AI-03 | 同一目标商品参考 3–5 条来源综合提取事实；来源记录仍独立 | 待 AI-01/02 真实调用验收，原P2。 | 原范围/前置及阶段记录：AI-01/02 真实调用先验收；关联与历史证据：[Issue #60](https://github.com/Fan6512/una-seller/issues/60) | P2 |
| MOD-01 | V0.7 模块化边界登记与后续实施：稳定模块 ID、明确数据归属、公开内部接口/依赖、桌面导航按模块组织；内置模块与主程序一起构建 | 模块边界已登记，逐功能 PR 落地，不单独大重构或做收费/动态插件。 | 原范围/前置及阶段记录：待开发；逐功能 PR 实施，不要求现有代码大重构，不引入动态加载/第三方 SDK/沙箱；本阶段不定义商业权益或价格；关联与历史证据：[#118](https://github.com/Fan6512/una-seller/pull/118) 首个内置 `collector.fx`（`internal/collectors/fx`）拆分外部传输/生命周期，财务拥有锁定日、校验/截取、存储和利润。保留独立启用开关；不引入通用调度框架、动态加载、插件市场或额外管理 UI；第二个真实采集器出现时再抽公共接口 | V0.7 P1 |
| DOC-01 | 将旧 README/部分早期阶段说明与最新实际状态对齐 | 本次总台账已整理；其他 README/早期说明仍待按范围同步，P2。 | 原范围/前置及阶段记录：以本台账与最终实测证据为准；关联与历史证据：[README](../../README.md)、[docs/README](../README.md) | P2，不阻断开发 |

### B. 暂缓 / 依赖阻塞

负责人已暂停的事项或明确后置到真实刊登闭环之后。未获恢复批准，不为它们重复创建实施 PR 或真实写入。

| ID | 功能 | 当前状态 / 下一步 | 关联证据及历史范围 | 原优先级 |
| --- | --- | --- | --- | --- |
| CBT-02 | **P0 · C1 首次真实刊登**：标准 UP + Net Proceeds + Remote；产品箱选择店铺及单站点，先具备单规格及多规格统一刊登能力，真实 Family 至少两规格、官方 ID 回读与防重 | 负责人尚无真实可销售商品，首次真实刊登 NOT_RUN；待商品准备并确认恢复。 | 原阶段记录：PUB-05A #115 已合并；负责人暂缓 CBT-02（尚无准备好的真实可销售商品）；首次真实刊登 NOT_RUN；关联：[Issue #50](https://github.com/Fan6512/una-seller/issues/50)；C0 [#85](https://github.com/Fan6512/una-seller/pull/85)、[#87](https://github.com/Fan6512/una-seller/pull/87)；验收判据：恢复条件：负责人确认商品已准备好；期间不专门进行刊登预检、迁移部署、官方图片上传或批准登记；联合 CBT-03/IMG-01/PUB-03 的最小切片；先核对合并后 CI、生产部署/备份/0021与0022隔离迁移、Nginx专用60s、候选商品与站点授权；真实写入前核实原始业务枚举、后台模式及逐站点能力，并获独立批准；成功必须匹配官方 ID 与归属回读，不盲重试。PUB-05A 作为前置 P0，单规格可选烟测，多站点批量后置；不等待 WH-06 真实通知 | 原阶段未单列 |
| CBT-03 | 站点售价/币种、库存、销售条件、店铺模板，核对 CBT/UP/Family/变体/SKU 映射 | #115 已实现最小单站点库存、价格/目标净收益及部分映射；站点字段/店铺模板剩余范围随 CBT-02 恢复，不重做核心。 | 原范围/前置及阶段记录：当前账号可写模型/官方字段已确认；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| IMG-01 | 合法来源图片原件/选用版本可持续保存；上传 Mercado Libre 并取得真实 picture_id | 随 CBT-02 暂缓真实图片上传、来源保存与官方 picture_id 联调。 | 原范围/前置及阶段记录：明确图片使用权、上传 API 和账号权限；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| PUB-03 | 刊登记录、逐站点错误定位、超时先核实、成功目标防重、失败目标安全重试 | #115 已实现尝试持久化、防重、受控恢复与回读核心；真实逐目标错误和安全重试验收随 CBT-02 恢复。 | 原范围/前置及阶段记录：CBT-02 发布合同落地；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0**，与 CBT-02 同一个闭环，不另搭队列平台 |
| PUB-04 | 修改已刊登价格、库存、图片、描述并读回；避免覆盖 Seller Center 人工修改 | 待 CBT-02 稳定映射及真实可写字段后实施；不提前官方修改。 | 原范围/前置及阶段记录：先有 CBT-02 稳定映射及真实可写字段；关联与历史证据：[Delivery](../architecture/DELIVERY.md) M2 | P1 |
| PUB-05B | 多店多站点批量刊登 | 明确后置，待单店单站点真实闭环。 | 原范围/前置及阶段记录：PUB-05A 与 CBT-02 真实单站点闭环后扩展；独立范围、防重和逐站点结果；关联与历史证据：[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | 后续 P1 |
| ORD-07 | 第三方物流 API 适配（Track718 / 17TRACK），辅助查询与轨迹缓存 | 已明确暂缓第三方物流 API；保留官方查询及网页入口。 | 原范围/前置及阶段记录：暂缓；现阶段只保留网页入口，优先官方 Shipment；不申请 Key、不开发适配器/表，不覆盖官方状态，不纳入 V0.7；关联与历史证据：待独立立项 | P2（暂缓） |

### C. 候选（未立项）

仅记录方向，无本期交付承诺，不自动开启开发。

| ID | 功能 | 当前状态 / 下一步 | 关联证据及历史范围 | 原优先级 |
| --- | --- | --- | --- | --- |
| OPS-04 | 后续高级运营分析与自动化候选（商品监控、调价策略、运营报表） | 高级运营分析与自动化候选。 | 原范围/前置及阶段记录：真实业务数据与人工流程稳定；自动官方写入另行批准；关联与历史证据：未立项 | P2，候选 |
| IMG-03 | 后续 AI 商品图与广告视频生成候选（IMG-02 白底先行） | AI商品图与广告视频候选。 | 原范围/前置及阶段记录：合法素材、真实服务能力、人工采用；是否复用 UnaAPI 后续决定；关联与历史证据：未立项 | P2，候选 |
| FIN-03 | 后续高级财务统计及利润分析候选 | 高级财务统计候选。 | 原范围/前置及阶段记录：FIN-01/02 真正完成净回款/退款/汇率/成本对账，不另建重复账本；关联与历史证据：未立项 | P2，候选 |
| CS-04 | 后续 AI 客服辅助候选：翻译、回复候选与人工采纳；自动回复不纳入当前范围 | AI客服辅助候选，保留人工采纳。 | 原范围/前置及阶段记录：CS-01/02、AI/TR 真实服务及隐私权限验证；关联与历史证据：未立项 | P2，候选 |
| MOD-02 | 后续内置应用目录/模块启用管理候选 | 内置应用目录/启用管理候选，不预设收费或第三方动态插件。 | 原范围/前置及阶段记录：MOD-01 模块边界稳定后另行设计；不预设付费、权益、订阅、支付或第三方动态插件；关联与历史证据：未立项 | P2，候选 |

### 5.1 模块边界登记（负责人确认方向，具体技术合同由后续 PR 审核）

仅按职责划分，**不划分免费或付费，不设计收费、权益授权或支付系统**。保持 Go + PostgreSQL、Tauri 单体部署与统一签名更新。新增功能优先归属现有模块，不因模块化拆微服务或重写已有业务。下面是职责归属，不代表已经完成目录重构或已冻结 API/表结构。

模块化优先、按需插件化：新增业务优先通过内置模块的稳定接口协作，明确数据所有者、依赖方向和错误处理；公共能力由核心或共享层统一提供，业务模块不直接耦合第三方实现，也不复制其他模块的数据、凭据和状态。插件化不等于每个功能独立安装或部署；当前保留单体构建、统一升级，不提前引入动态加载、插件 SDK、沙箱、微服务或收费体系。每个功能 PR 明确模块归属及跨模块调用边界，只有真实需求出现时才扩展接口。

| 模块 | 职责 / 对应功能 ID | 边界 |
| --- | --- | --- |
| 核心与平台连接 | BASE、AUTH、DESK；店铺 OAuth、站点映射、认证、更新、缓存隔离 | 统一身份与店铺授权；业务模块不自行持有/复制授权 Token |
| 采集与产品 | COL、WB、CAT | 来源、产品事实、工作稿及类目；不负责官方刊登写入 |
| 翻译与 AI 文案 | TR、AI | 翻译与事实约束候选；人工采纳后由产品模块保存 |
| 图片与媒体 | IMG、DESK-03 | 原始图片、预览/修图候选、媒体能力；官方图片上传与刊登协作 |
| 刊登与商品运营 | CBT、PUB | 官方发布、站点映射、回读及已刊登商品状态；统一发布防重 |
| 订单与官方物流 | ORD-01/03/04/07、WH-03 | 交易、Pack、Shipment、SLA 与履约；不把采购物流当成官方发货 |
| 采购与国内物流 | ORD-02/05/06 | 采购关联、成本来源、国内/货代物流；与官方订单引用关联 |
| 财务 | FIN | 官方账单、确认回款、退款/费用、汇率、利润；不伪造净回款 |
| 售前问答 | WH-04 问答、CS-01 | 商品问题及人工答复 |
| 售后消息 | WH-04 消息、CS-02 | 会话、消息及人工答复 |
| 投诉纠纷 | WH-05 纠纷、CS-03 | 纠纷、期限及官方允许操作；退款实际结算仍由财务确认 |
| 通知与后台同步 | OPS-01/03、WH-06 | 统一回调、队列、任务、限流、补偿；通知处理器调用各业务模块，不另建重复业务库 |
| 运营中心 | OPS-02/04 | 聚合各模块真实待办/统计/动态、已读与跳转；不替代业务源状态 |
| 桌面公共交互与反馈 | DESK-05/06 | 共享动效与操作反馈接口；Ding 为可替换渲染适配器，不拥有业务状态、不直接执行业务操作、不替代 OPS-02；由调用方负责业务结果及错误 |
| 模块组织与应用入口 | MOD-01/02 | 内置模块 ID、依赖和导航/目录；不开发动态安装、收费与第三方 SDK |

## 六、下一轮执行顺序（当前有效，旧裁决仅作历史）

1. **V0.8.0 客户端交付已完成；Windows 实机验收为当前第一优先**：负责人从 V0.7.0 开启接收测试版更新并升级到 0.8.0，验证重启、登录/配置/本地资料保留；关闭测试版开关确认稳定通道不提供测试版且不降级；验收店铺、订单、商品、财务、PUB-07 家族/刊登双视图、TR-03 翻译设置、通知状态。产品编辑仅检查展示，不真实刊登、不修改官方商品、不触发通知回放。开发代理负责 Release/签名/CI/OTA/后端技术核验，负责人执行本机 GUI；未测不记 PASS。
2. **TR-02 每日增量维护保护流程**：TR-02E 全量词典已发布，managed 保护已生效；下一步开发只处理新增/变更官方类目的受保护增量流程，包含额度预检、staging、差异/质量检查、人工批准、原子发布与回滚。普通旧 updater 不得覆盖正式词典；timer 在独立验收与批准前保持关闭。
3. **FIN-02A / OPS-05 运行验收**：核对北京时间订单日、销售额、预计利润、店铺/站点范围、覆盖率、固定汇率、真实成本录入及详情/首页一致性；历史汇率按计划回填。不得把账单估算冒充实际到账。
4. **FIN-02B 实际结算独立推进**：先取证订单—付款批次—收款流水映射、退款/共享费用归属和实际汇率，再决定最小开发切片；没有二次来源验证不接线、不冒充实际利润。
5. **WH-06 / OPS-03 / OPS-02 自然样本观察**：核验真实公网 ACK、订单/货件消费、重复去重、429 恢复和漏单；不伪造通知、不回放、不扩大主题，不重开已收口的补偿开发。
6. **CBT-02 暂缓**：负责人准备真实可销售商品并明确恢复后，再联合 CBT-03/IMG-01/PUB-03 完成单店单站点最小真实刊登闭环，逐项取得真实图片上传及官方创建授权；不重做 PUB-05A，多店多站点批量后置。

上述是已批准方向的当前整理，不新增功能或改变正式优先级。原 V0.6/V0.7/V0.8/V0.9 计划和历史决策见 [整理前台账](https://github.com/Fan6512/una-seller/blob/abe3eb2b7c77caa589daafb61b1c42ac09c5da7d/docs/product/FEATURE-STATUS.md)、下方变更记录；与本节冲突时以本节最新状态为准。

## 七、新需求登记、完成记录范例

新增需求在对应类别增加唯一 ID；已有功能只更新原行。示例（非实际需求）：

| ID | 功能 | 当前状态 / 下一步 | 关联证据及历史范围 | 原优先级 |
| --- | --- | --- | --- | --- |
| WB-04（示例） | 产品箱新增某项操作 | 待开发；Windows 实际点击、保存并读回成功 | 实施 PR 待建 | 待确认 |

开始开发在原行注明 Draft PR 和当前进度；合并后按剩余事项移至待部署或待实测；真实验收 PASS 才移至已完成（标注范围）。暂缓或候选保留原因与历史证据。每次只改相关行及必要的当前摘要和变更记录，避免重复登记。

## 八、变更记录

| 日期 | 变更 | 依据 |
| 2026-11-28 | V0.8.0 客户端交付复核：固定源码 `6a56734f` 的签名 Pre-release `desktop-test-v0.8.0` 已发布；main CI、签名 CI、Windows NSIS 与 Release staging PASS；安装包 SHA-256 与 GitHub digest 一致，测试 OTA 200、正式 OTA 204（no-store）。PUB-07/TR-03 移入待实测，DESK-01 Windows 实机验收明确 NOT_RUN；归档交付回执并关闭重复文档 PR #124。 | [V0.8.0交付回执](../architecture/V080-JOINT-DELIVERY-20261127.md)、[Release](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.8.0)、[签名 CI](https://github.com/Fan6512/una-seller/actions/runs/45974523186) |
| --- | --- | --- |
| 2026-11-28 | #131 HEAD `5ea9223f` PR CI 46111706636 PASS，已解除Draft并Squash合并 `20404c6a`；发布/回滚前持久化managed标记，普通updater联网前阻断，回滚后保护保留。main CI 46112137829 PASS；未部署、未发布。 | [#131](https://github.com/Fan6512/una-seller/pull/131) |
| 2026-11-28 | 负责人批准#130合并、离线组装及最终300条复验，并开发原子发布回滚工具；#130 e3e3fa45 已合并、合流CI 46111106831 PASS，3个后端文件一致；Draft #131 59b66c63 已提交哈希绑定审批、双锁、原始样本及人工修正门禁、备份/回滚。生产104条覆盖回执未取得，离线组装未执行，未授权正式发布或timer。 | [#130](https://github.com/Fan6512/una-seller/pull/130)、[#131](https://github.com/Fan6512/una-seller/pull/131) |
| 2026-11-28 | 负责人批准TR-02E 24条纠错（80→104）；精确路径清单及已知历史ID归档，现场应用与独立300条复验待执行。#130 2d79c828 CI 46108576244 PASS，Draft待审批；未发布、未新增调用。 | [交接](../architecture/TR-02E-FINAL-AUDIT-HANDOFF-20261128.md) |
| 2026-11-28 | TR-02E 全量：`c155228f` 部署；十条纠错 70→80；42 批完成 10,454 新候选 + 28 延期 + 10,480 回执，覆盖 98.94%；额度 847,341/900,000；首批 200 条 191/200（95.5%）、最终 300 条 285/300（95.0%，CI 91.9%–96.9%）、15 条错误；未发布、timer 未启用，发布前修正 24 条。 | [报告](../architecture/TR-02E-FULL-BUILD-ACCEPTANCE-20261128.md)、[质量表](../architecture/TR-02E-QUALITY-300-20261128.csv) |
| 2026-11-27 | #129 已解除Draft并合并 c155228f，审核及合流 CI PASS，6个后端文件与审核版本一致、无迁移；负责人已批准部署、10条纠错、完整路径全量建设，仍待生产执行。提前提交 Draft #130 1a467037 离线候选合并与300条独立审查样本，未发布、不启用timer。 | [#129](https://github.com/Fan6512/una-seller/pull/129)、[合流CI](https://github.com/Fan6512/una-seller/actions/runs/46026908679)、[#130](https://github.com/Fan6512/una-seller/pull/130) |
| 2026-11-27 | 负责人批准 TR-02E 十条纠错（70→80）、未来质量门禁调整、完整路径全量及90万受限执行器开发；Draft #129 df4a41d4 已提交全量候选构建、预算/断点/未知结果保护及纠错清单，历史严格零错误门禁不改判，最终真实验收/发布仍单独门禁。未部署或调用。 | [#129](https://github.com/Fan6512/una-seller/pull/129) |
| 2026-11-27 | #128 已合并83b505fa，合流CI 46019671633 PASS；已知201历史ID、6条纠错清单及CSV字符核算归档，差186待现场追溯，早期历史ID须补齐。待生产部署、覆盖及第二轮200条验收/全量预算预检；未调用、未扩额、不发布、不启用timer。 | [交接](../architecture/TR-02D-SECOND-ROUND-HANDOFF-20261127.md) |
| 2026-11-27 | TR-02D 继续收口：Draft #128 d4156e36，PS Vita 精确专名例外、历史样本ID显式排除、逐行源文哈希/预扣/结果回执及六条修正清单；保持5万累计上限及原质量门禁。254字符历史差额待现场核实，未应用生产覆盖、未调用新样本、未发布。 | [#128](https://github.com/Fan6512/una-seller/pull/128)、[CI 46018393964](https://github.com/Fan6512/una-seller/actions/runs/46018393964) |
| 2026-11-27 | 负责人核对百度控制台：大模型13,691、本地13,794，差103恰等于首次54001失败预扣；普通191,828、本地192,758，差930。两者保留较高本地计数，本轮对账PASS；结果支持reference未额外计费，不代表所有未来请求；200条候选源字符13,505与本轮预扣13,251差254仍需按ID/请求记录核对，质量严格门禁不改判。 | 负责人控制台读数；[TR-02D 验收](../architecture/TR-02D-PRODUCTION-LLM-ACCEPTANCE-20261127.md) |
| 2026-11-27 | TR-02D 第二轮：`83b505fa` 部署；六条纠错 64→70；历史排除 544 ID；186 字符追溯为 CBT106855 两次 93；独立 200 条 LLM 190/200（95.0%，CI 91.0%–97.3%），10 条叶子级错误（0 跨一级行业），零错误门禁未过；完整路径预算 818,169、父级 332,696；本地 LLM 27,144；未发布/未扩额/timer 未启用。 | [报告](../architecture/TR-02D-SECOND-ROUND-ACCEPTANCE-20261127.md)、[质量表](../architecture/TR-02D2-QUALITY-200-20261127.csv) |
| 2026-11-27 | TR-02D 生产部署与受限试译：#127 `4c387e7c` 部署；1/5/200 条真实试译；LLM 194/200（97.0%）、普通 172/200（86.0%）；6 条叶子级语义错误（0 跨一级行业），严格零错误门禁未过；64 覆盖/词典/进度未变；本地 LLM 13,794，控制台待核对；未发布、未扩 5 万、timer 未启用。 | [报告](../architecture/TR-02D-PRODUCTION-LLM-ACCEPTANCE-20261127.md)、[CI](https://github.com/Fan6512/una-seller/actions/runs/46008869942) |
| 2026-11-27 | #127 已解除Draft并Squash Merge `4c387e7c`，合流CI PASS；据现场回执生产升级成功，64条人工覆盖已加载，离线单条103字符/200条13444字符预算已验证。受限试译因缺独立大模型Key与真实控制台用量N阻塞，未初始化LLM计数、未调用；普通翻译本地192758保留，不续跑、不发布、不启用timer。 | [#127](https://github.com/Fan6512/una-seller/pull/127)、[CI](https://github.com/Fan6512/una-seller/actions/runs/46008869942)、现场回执 |
| 2026-11-27 | #127 独立复审：f3f2e03c CI PASS；采纳P1额度报告、P2候选行ID与额度上限澄清及低成本CLI改进，推送046463f9。首轮仍硬限制5万，90万仅后续全量目标；不静默删除候选、不自动续跑、不消耗真实额度，保持Draft待新CI。 | [#127](https://github.com/Fan6512/una-seller/pull/127) |
| 2026-11-27 | 负责人补齐大模型官方输入/输出合同；#127 明确 llm 模型、reference 电商指令、needIntervene=0、Bearer 无sign/salt、2000字符及严格源/目标和错误码校验，补模拟测试；旧HEAD CI PASS，新HEAD f3f2e03c 待CI。未真实调用或部署，术语库暂不启用，首次仍需控制台累计额度对账。 | [#127](https://github.com/Fan6512/una-seller/pull/127) |
| 2026-11-27 | 负责人授权先行TR-02D并提供官方Bearer curl；创建Draft #127，HEAD 626c01ba，离线样本/候选/累计额度及显式受限试译已实现；不修改原普通词典/默认provider、不发布，真实响应与可选指令字段待核。未实际调用或部署，CI待验。 | [#127](https://github.com/Fan6512/una-seller/pull/127) |
| 2026-11-27 | 负责人确认百度大模型文本翻译100万、领域文本翻译50万剩余免费额度；TR-02 质量89%阻塞，登记 TR-02D 质量校正建议，先验证开放平台正式接口合同与小规模样本，不修改默认Provider、不续跑旧任务、不消耗额度。 | 负责人现场回执；[百度产品](https://fanyi-api.baidu.com/product/13)、[定价](https://api.fanyi.baidu.com/product/132) |
| 2026-11-27 | TR-02 生产收尾：`f644d353` 已部署并独立复核；90 万额度保留原 3,118 计数；全量任务安全暂停于 3,779/10,954（34.08%），`active.json` 未发布；200 条分行业抽检 178/200（89.0%），22 条明确错误已按完整路径写入 overrides；质量未达 95%，未启用每日增量，未消耗大模型额度，未重建客户端。 | [报告](../architecture/TR-02-PRODUCTION-DELIVERY-20261127.md)、[联合 CI](https://github.com/Fan6512/una-seller/actions/runs/45999249269) |
| 2026-11-27 | #125→#126 已 Squash Merge（a8579e98→f644d353），联合 CI 45999249269 PASS；据现场回执已生产部署、启用90万字符额度且保留计数，五批真实试译0失败，旧96条复核24条修正、新80条原始可用62/80（77.5%），10错误8偏差已人工覆盖；全量 unit 最后观察 active、额度99,860；质量目标95%尚未满足，最终发布/覆盖/独立质量抽检/客户端及timer待验，不重复启动，不将接口成功视为质量PASS。 | [联合 CI](https://github.com/Fan6512/una-seller/actions/runs/45999249269)、负责人现场回执 |
| 2026-11-27 | 二轮复审采纳 #125 个人Key已知源语言批量退化P1与轻量日志/注释修复，#126 完整词典术语更新早退、预算快照和批次假设、锁错误提示；分别推送3236e22a、eddb0a91，旧HEAD Windows构建均已PASS，新HEAD CI待验，保留原PR及Ready状态。未合并、部署或试译；生产仍需共用同一绝对usage路径，额度对账、94条复核、预算及五批复验后再全量。 | [#125](https://github.com/Fan6512/una-seller/pull/125)、[#126](https://github.com/Fan6512/una-seller/pull/126) |
| 2026-11-27 | #125/#126 Go Backend、Changed Files 均 PASS，Desktop/MV3 按范围跳过；代码复核后已转 Ready，等待批准依次合并 #125→#126。未修改生产、未扩大额度、未调用真实接口；仍需真实94条路径复核与 -plan 预算、样本复验。 | [#125 CI](https://github.com/Fan6512/una-seller/actions/runs/45978738415)、[#126 CI](https://github.com/Fan6512/una-seller/actions/runs/45978890002) |
| 2026-11-27 | 负责人确认百度通用文本剩余997211字符、大模型测试额度500000；批准 TR-02A/02C/02。创建 #125 跨进程持久额度与逐次请求扣额、#126 真实父级语境/保守术语/零调用预算预检；两者 Draft，待 CI/复审。生产90万额度尚未修改，先复核94条旧译与重算预算再受控试译，未执行全量。 | [#125](https://github.com/Fan6512/una-seller/pull/125)、[#126](https://github.com/Fan6512/una-seller/pull/126) |
| 2026-11-27 | TR-02B 百度翻译接入并部署 `2fc79081`：五批真实试译 5/5、0 失败、2,789 字符，staged 94/10,954；缓存命中、续跑、月额度（默认 50,000）PASS；质量 5/80 明确错误；默认 Provider 已切 `baidu`；未全量、未配 cron。 | [PR #123](https://github.com/Fan6512/una-seller/pull/123)、[TR-02B 验收](../architecture/TR-02B-BAIDU-ACCEPTANCE-20261127.md) |
| 2026-11-27 | TR-02 生产接入：后端升级 `6a56734f`，新增词典目录并发布官方 Dump（11,668/714/10,954），中文 0/10,954；五批真实试译 BLOCKED（火山 `-403 account status abnormal: unsynchronized`，1/5 批、0 译文），未切换供应商/未全量/未配 cron；登记月 180 万字符额度保护 TR-02A 待开发。 | [TR-02 验收](../architecture/TR-02-CBT-DICTIONARY-ACCEPTANCE-20261127.md) |
| 2026-11-27 | 负责人明确批准 #122 Squash Merge、生产后端升级及 desktop-test-v0.8.0 签名构建/OTA；已合并 6a56734f，回读五个版本文件均为0.8.0；联合业务CI四项 PASS、版本PR CI PASS、无新增SQL迁移。固定发布候选6a56734f，合并后版本CI待核；当前连接不具备生产SSH、Tag/Release写入能力，尚未部署/打Tag/签名/发布/真实翻译；由受权本地代理接续，保留既有签名审批。 | [#122](https://github.com/Fan6512/una-seller/pull/122)、[版本CI](https://github.com/Fan6512/una-seller/actions/runs/45971124751) |
| 2026-11-27 | #120/#121 联合业务基线 0fe79e02 CI attempt2 四项 PASS，确认无新增 SQL 迁移；负责人批准 V0.8.0（现有签名和 OTA 规则仅接受 0.X.0，原 0.7.0 不覆盖）；创建 #122 同步五个版本文件，候选 eb99f0d4，CI 待验。未执行生产部署、真实翻译、签名或发布。 | [联合 CI](https://github.com/Fan6512/una-seller/actions/runs/45970530199/attempts/2)、[#122](https://github.com/Fan6512/una-seller/pull/122) |
| 2026-11-27 | 负责人批准 #121 Squash Merge，已合并 0fe79e02；审核 HEAD 61813971 合流 CI 四项及 Windows NSIS PASS。TR-03 移入已合并待交付，TR-01/02 更新；#120/#121 待联合 CI、部署、真实试译与 V0.7.1 测试版发布（须另行批准），未执行生产操作。 | [#121](https://github.com/Fan6512/una-seller/pull/121)、[合并提交](https://github.com/Fan6512/una-seller/commit/0fe79e0226ad4696f84b9bd4aa186226b2f6844e) |
| 2026-11-27 | #121 合流 HEAD 61813971 CI 45969191526 四项 PASS，复核保留 #120 官方元数据筛选及火山16条批次、10/30秒预算、鉴权冷却；main 后续仅台账文档变化，无冲突；解除 Draft 待 Windows NSIS 和明确合并批准。未真实调用、未部署。 | [#121](https://github.com/Fan6512/una-seller/pull/121)、[合流 CI](https://github.com/Fan6512/una-seller/actions/runs/45969191526) |
| 2026-11-27 | #121 移除豆包候选 a25eccdd CI 四项 PASS；同步已含 #120 的 main 776ba5a4，合流 HEAD 61813971，唯一重叠 translation.go 的 #120 元数据取词函数逐字保留，同时保留火山批次与超时策略；与 main 差异仍12文件，无合并冲突；新 CI、独立复审及真实接口待验，保持 Draft。 | [#121](https://github.com/Fan6512/una-seller/pull/121)、[旧候选 CI](https://github.com/Fan6512/una-seller/actions/runs/45968653458) |
| 2026-11-27 | 负责人明确批准 #120，已 Squash Merge 84efeabd；PUB-07 移入已合并待交付，TR-01 对应更新；审核及 Windows NSIS PASS 不等于 GUI 实机验收。#121 单独待复审及合流，未部署/发布。 | [#120](https://github.com/Fan6512/una-seller/pull/120)、[合并提交](https://github.com/Fan6512/una-seller/commit/84efeabd269df36b2c5ee81e90fec5bede51fbc0) |
| 2026-11-27 | #120 5870d3cb CI、独立复审与 Windows NSIS 构建 PASS，Ready 待合并，GUI 未验；#121 按负责人要求移除豆包，保留火山通用、Azure、百度，a25eccdd 修订在线类目超时及鉴权/限流，新增81条真实 handler mock 验证；新 CI 待验，未部署/真实翻译。 | [#120](https://github.com/Fan6512/una-seller/pull/120)、[#121](https://github.com/Fan6512/una-seller/pull/121) |
| 2026-11-27 | #120 独立复审后收口：生产导航 guard、挂载卸载复位、同步/刷新确认与选择门禁；现有 DOM fixture 验证确认取消、移除与草稿回读；展开焦点、翻页模式保留及枚举兜底。旧4709f000 CI PASS，新候选5870d3cb 待 CI/复审，保持 Draft；权限失效强制关闭不保留未手动保存内容。 | [#120](https://github.com/Fan6512/una-seller/pull/120) |
| 2026-11-27 | 负责人补充接入火山通用机器翻译；#121 原分支新增 AK/SK 签名独立 provider、账号设置与后台16条批次支持，保留豆包/Azure/百度；TR-02 优先通用翻译，先五批核对质量/计费。旧豆包 CI PASS，新候选39ac6417 待 CI/真实接口验收，未部署或翻译生产词典。 | [#121](https://github.com/Fan6512/una-seller/pull/121) |
| 2026-11-27 | 负责人批准接入火山豆包机器翻译：TR-03 独立 Draft #121，复用账号加密凭据与 TR-02 离线字典，16条批次、失败原文回退和定点测试；未使用真实 Key 发请求、未部署、未启用全量翻译；Token计费与火山通用翻译免费字符政策不同。 | [#121](https://github.com/Fan6512/una-seller/pull/121)，候选 988fdaa5 |
| 2026-11-27 | 负责人重申 TR-02 类目离线预翻译：现有 #72 已具备官方 Dump、版本校验、增量复用、断点续跑和热加载，CBT 预测/选择读词典，不现场翻译类目路径。生产覆盖率、维护凭据及 cron 待核实，先五批真实试译，再续跑；各站点路径离线覆盖另行收口，不改变动态属性官方规则或 #120 布局验收。 | 负责人确认；[现有维护说明](../../backend/README.md)、[#72](https://github.com/Fan6512/una-seller/pull/72) |
| 2026-11-27 | #120 复核：417e59b2 桌面 CI 旧浏览器断言要求中文主路径，和新要求冲突；4709f000 修正主/辅路径真实 DOM 验收并验证中文枚举保留原始保存值、品牌自由文本不翻译；最新 CI 与独立审核待验，未合并/部署。 | [#120](https://github.com/Fan6512/una-seller/pull/120)、[旧候选 CI](https://github.com/Fan6512/una-seller/actions/runs/45963178625) |
| 2026-11-27 | 负责人明确分类路径原文优先、中文辅助：CBT 官方英文在上、中文在下；站点路径保留实际官方语言，不伪造官方英文；其他属性标签仍中文优先。#120 原分支更新显示及回归测试，候选 417e59b2，未合并/部署。 | [#120](https://github.com/Fan6512/una-seller/pull/120) |
| 2026-11-27 | 负责人补充已刊登详情中文化，更新 TR-01/PUB-07；#120 原分支补充站点类目路径、属性组/字段、官方选项中文显示，保护用户原始文本及标识符，缺译原文回退。旧布局 CI PASS；新候选 f559513b 待 CI/独立审核、真实语言样本；未合并、部署、发布。 | [#120](https://github.com/Fan6512/una-seller/pull/120)、[旧布局 CI](https://github.com/Fan6512/una-seller/actions/runs/45962204992) |
| 2026-11-27 | PUB-07 首版 UI 候选已提交 Draft #120：整行展开与实际宽度换列、独立刊登图片墙、详情按需挂载及草稿离开确认；新增一项现有 editorUx 布局回归。仅当前页归组，CI/Windows/独立审核待验；未合并、未部署、未发布。 | [#120](https://github.com/Fan6512/una-seller/pull/120)，候选 `202de201` |
| 2026-11-27 | 负责人确认 PUB-07 双视图：家族固定尺寸卡片、单家族整行规格展开；刊登独立主图墙、点击后右侧详情、关闭恢复全宽。独立桌面 UI 开发，不改变同步/官方写入权限、现有发布及 FIN-02A 验收顺序。 | 负责人 2026-11-27 交互设计确认 |
| 2026-11-27 | V0.7.0 联合交付：#115–#118 已合并，联合 CI attempt2 PASS，#119 版本同步合并 76cefd1d，签名 Pre-release 已发布。负责人报告服务端升级、FX/Billing 启用及 OTA 发布完成；历史汇率仍回填。将本批条目由待部署/发布移至待实测/联调，Windows 人工验收由负责人完成，开发代理负责技术核验，不标记客户端 PASS；CBT-02 不真实刊登、FIN-02B 待开发。 | [联合 CI 45957718570 attempt2](https://github.com/Fan6512/una-seller/actions/runs/45957718570/attempts/2)；[V0.7.0 签名测试版](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.7.0)；[Windows 发布 CI 45960420514](https://github.com/Fan6512/una-seller/actions/runs/45960420514)；2026-11-27 负责人转述本地联合交付回执（V070-JOINT-DELIVERY-20261127.md，本次未直接读取原文件） |
| 2026-11-27 | 负责人批准 #118 合并，FIN-02A Squash Merge `1f697fc`；最终候选 `a34dcc4` 四项 CI PASS，预计利润开发收口，移入待部署/发布；OPS-05 同步预计利润接线，FIN-02B 保留待开发。未生产迁移、部署、启用采集或发布客户端。 | [#118](https://github.com/Fan6512/una-seller/pull/118)；[CI 45957315802](https://github.com/Fan6512/una-seller/actions/runs/45957315802)；负责人批准 |
| 2026-11-27 | AUTH-01/02 真实双账号核心业务验收 PASS：负责人确认新账号商品同步、浏览器插件单品采集、切回原账号后订单/刊登/采集箱恢复均正常。生产库两轮只读核对：两账号各一连接、各五条 verified 映射；原账号订单/刊登/采集/工作稿/产品库/账单快照保持 73/207/4/3/1/3；新账号分别 5→14、0→50、0→1、0→0、0→0、0→1。新账号首轮 5/5 订单存在物流快照；通知任务 5→0，不能单凭此判定消费成功。两轮重复店铺归属、通知任务/收件/订单发现、工作稿/产品库来源归属异常均 0；anon/authenticated 无 app schema USAGE。主动 API 越权读写、长期 Token 续签、公开注册关闭状态未实测；不将快照存在性视为完整对账，不修改生产数据，不改变 FIN-02A 排期。 | 负责人 2026-11-27 Windows/浏览器实机反馈；同日生产 Supabase PG17.11 两轮只读查询；[插件 V0.1.4 main 构建](https://github.com/Fan6512/una-seller/actions/runs/45857160861) |
| 2026-11-27 | FIN-02A #118 接线 CI 45932096096 PASS；自动净额取证补 Billing 付款/交易/费用订单与Pack关联字段（不保留买家数据、不推断净回款），待现场同订单直接净额字段、MLM/退款/共享与资金报告关联样本。最新 HEAD ec3f1ba8，CI in_progress；未生产变更 | [#118](https://github.com/Fan6512/una-seller/pull/118) |
| 2026-11-27 | 负责人确认内置采集插件方向；#118 新增 collector.fx 独立传输/生命周期及 0024 订单确认、人民币利润读写和首页覆盖率。原汇率 CI PASS，接线首轮桌面 PASS、Go格式失败已修订；最新 CI in_progress。自动订单净额来源仍未核实；人工确认只是可选兜底，不将未知算0。未生产变更 | [#118](https://github.com/Fan6512/una-seller/pull/118) |
| 2026-11-27 | FIN-02A 新增人民币日度核算规则：北京时间订单创建日统一固定汇率、截取一位小数；#118 43f7f7d9 新增0023及可选自动采集、历史日期补齐、CNY精确计算；未生产启用、未接运行UI，提供方实测与预计净回款证据仍待核实 | [#118](https://github.com/Fan6512/una-seller/pull/118) |
| 2026-11-27 | 负责人批准推进 FIN-02A，#118 第一切片提交：两笔预计净额的算术复核及独立预计利润纯合同，未接运行接口、不自动推算未核实订单、不增加迁移或生产操作；实际结算独立待证 | [#118](https://github.com/Fan6512/una-seller/pull/118)、[样本回执](../architecture/FIN-02A-EXPECTED-PROFIT.md) |
| 2026-11-27 | 负责人批准收口剩余 PR：#76 同步最新分类台账后仅补 ACC-01/CRED-01 及早期工作台证据，CI 45928354039 PASS，Squash Merge 7005a33c；#98 基于当前 main 保留五个版本文件六处 0.5.0→0.6.0，CI 45928454082 PASS，Squash Merge 9aaaa0e4；无开放 PR，62 个功能 ID，正式统一部署与签名发行仍独立审批 | [#76](https://github.com/Fan6512/una-seller/pull/76)、[#98](https://github.com/Fan6512/una-seller/pull/98) |
| 2026-11-27 | #76 历史文档分支同步最新台账，补回 ACC-01/CRED-01（#77/#78 已合并，真实账号与凭据待实测），保留 11-23 工作台核心实机证据；不恢复旧刊登顺序，不修改业务代码或审批 | [#76](https://github.com/Fan6512/una-seller/pull/76)、[Issue #75](https://github.com/Fan6512/una-seller/issues/75) |
| 2026-11-27 | 负责人要求分类整理：按已完成、待部署/发布、待实测、待开发、暂缓、候选六类重排；保留所有功能 ID、原范围/优先级、历史证据、流程图及模块边界。更新 #115/116/117 已合并、联合部署待批、CBT-02暂停、FIN-02A/B 证据阶段及当前执行顺序。不新增开发或生产操作 | [整理前台账](https://github.com/Fan6512/una-seller/blob/abe3eb2b7c77caa589daafb61b1c42ac09c5da7d/docs/product/FEATURE-STATUS.md)；负责人授权 |
| 2026-11-26 | 负责人批准 #105 Squash Merge `59472ef3`，OPS-02 收口。#106 重基 `1e199f96`，保留 #104/#105 新代码；候选 `7b729254` 新增0020 来源/单页回放 checkpoint、严格历史资源归一化、原队列事务与断点、服务端手动命令，默认禁用；CI 45768614663 in_progress，真实官方/Nginx/ACK 待验，保持 Draft。未部署/迁移/启用/订阅/刷新Token | [#105](https://github.com/Fan6512/una-seller/pull/105)、[#106](https://github.com/Fan6512/una-seller/pull/106) |
| 2026-11-26 | 负责人确认新增 DESK-06 桌面全局操作反馈（Ding）：单胶囊、统一五态及公共入口，Batch & Counter 与 Lifecycle & State 同为本功能 P0，真实进度、部分成功/失败汇总、安全重试及并发隔离；提示音默认关闭，保留业务错误和任务结果，与 DESK-05/#102 及 OPS-02/#105 分离。全局 P1、待开发、独立实施 PR 待建；仅登记需求，未修改代码、未部署或发布 | 负责人需求；[Ding](https://github.com/YusufCeng1z/ding) |
| 2026-11-26 | #105 `9bc1eabd` 精确 HEAD CI 45765732968 Changed Files/Go/Desktop PASS；复核问答 UNANSWERED 与 Pack 会话待回复规则、已回复清待办/迟到旧消息不重开/新消息按会话重开及归属，main 仅台账差异且 mergeable=true。最小有效测试收口，真实消息同步及 Windows 实机 NOT_RUN，待负责人批准合并；未生产操作 | [#105](https://github.com/Fan6512/una-seller/pull/105)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45765732968) |
| 2026-11-26 | 用户确认回复过的消息算已处理，新买家消息才重新提示；#105 `9bc1eabd` 按 Pack 精确归属与官方创建时间派生待回复会话及动态，沿用已有 PG 测试，不新增表/队列，不标官方已读；上一 HEAD CI PASS，当前待验，真实消息及 Windows NOT_RUN，保持 Draft | [#105](https://github.com/Fan6512/una-seller/pull/105) |
| 2026-11-26 | 负责人批准 #104 Squash Merge `49176630`，OPS-03 收口，真实 CBT/Windows 联调后置。#105 重基 `6239b729` 并推进 `4f006054` 运营首页、铃铛、摘要/待办/动态与筛选，无新表/队列；沿用已有测试补最小归属断言；CI 45764620023 运行中，保持 Draft。后续采用最小有效测试，不提前扩展复杂测试。未部署/迁移/启用/订阅 | [#104](https://github.com/Fan6512/una-seller/pull/104)、[#105](https://github.com/Fan6512/una-seller/pull/105) |
| 2026-11-26 | #104 `889cd8e` CI Go FAIL，五个 discovery 守卫暴露现有夹具泄漏；源码定位数据库故障测试关闭共享夹具池且清理吞错。新 `13204403` 独立故障池、owner 级清理错误显性化，保留 generation 与加强 E2E；四文件无运行时/迁移改动，新 CI/PG 待复验；真实 CBT/Windows NOT_RUN，保持 Draft，未合并/部署/迁移/启用 | [#104](https://github.com/Fan6512/una-seller/pull/104)、[失败 CI](https://github.com/Fan6512/una-seller/actions/runs/45750091142) |
| 2026-11-26 | #104 复审：b3930bc CI 四项与 PG 真停机恢复 PASS；拒绝测试全库删除其他店铺，改为检测残留 fail-fast；E2E 加入真实搜索适配/共享协调/详情503重试，修复共享队列旧 generation 失败或过期租约误终止新通知。新 `889cd8e9` 六文件、无新迁移，CI 45750091142 运行中；真实 CBT token 过期及桌面交互 NOT_RUN，保持 Draft。未合并/部署/迁移/启用/订阅 | [#104](https://github.com/Fan6512/una-seller/pull/104)、[新 CI](https://github.com/Fan6512/una-seller/actions/runs/45750091142) |
| 2026-11-26 | #104 OPS-03 完整补偿候选 `9487cd68`：新增0019 独立搜索进度，固定创建时间窗口/重叠分页、持久页 ID/断点、共享队列背压与限流、待付款补查及桌面闲置首页本地轮询。15文件增量，当前 CI 45746002509 运行中，真实 CBT、PG 隔离与 UI 验收待完成，保持 Draft；WH-06 仍负责 missed-feed 回放，未合并/部署/迁移/启用 | [#104](https://github.com/Fan6512/una-seller/pull/104)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45746002509) |
| 2026-11-26 | 负责人批准 #103 Squash Merge，合并提交 `1f69d14d`，WH-05 本阶段代码收口，真实 opened 投诉 NOT_RUN。#104 按实际 Git 祖先 c8a7ddd6 重放 8 文件至新 main，HEAD `79ba4c1b`，保留 #101/#103 修复；当前功能缺失新订单发现及持久化搜索窗口，保持 Draft、待开发及新 CI；#105/#106 尚待重基。本轮未部署、迁移、启用或订阅 | [#103](https://github.com/Fan6512/una-seller/pull/103)、[#104](https://github.com/Fan6512/una-seller/pull/104) |
| 2026-11-26 | #103 最终复审 `4d59dcdb` 限定代码 APPROVE，待负责人合并授权；精确 HEAD CI Go/PG 真停机重启 PASS；卖家动作及 mandatory 期限、商品缺失/零值/旧响应、迁移约束已核对，本地重复迁移及定点/重启据报告 PASS；真实 opened 投诉 NOT_RUN；main 仅台账变化，无冲突。停止扩大 PR；#104 后续重基并补齐新订单发现、持久搜索窗口。未合并/部署/迁移/启用/订阅 | [#103](https://github.com/Fan6512/una-seller/pull/103)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45740159214)；本地复审报告 |
| 2026-11-26 | #103 收口复审：45ea18b8 重基无代码冲突，追加 `4d59dcdb` 商品 sub_status 缺失保留、明确空数组清空及 PG 边界测试，保留卖家动作归属及强制期限修订；最新 CI 待复验，真实 opened 投诉 NOT_RUN，未合并/部署/迁移/启用 | [#103](https://github.com/Fan6512/una-seller/pull/103) |
| 2026-11-26 | 负责人批准 #101 Squash Merge，合并提交 `8b8aa1a6`；WH-04 本阶段代码收口，真实消息及未回答问答 NOT_RUN，消息主题禁止启用。#103 重放 12 文件到新 main，HEAD `45ea18b8`，保留 #101 ID 兼容及 #103 卖家动作/强制期限修复，当前 CI 待复验，保持 Draft。#104–#106 尚未重基；未部署、迁移、启用、订阅 | [#101](https://github.com/Fan6512/una-seller/pull/101)、[#103](https://github.com/Fan6512/una-seller/pull/103) |
| 2026-11-26 | #101 最终复审 `e855583` 限定代码 APPROVE（待合并授权）；精确 CI Go/PG 真停机重启 PASS；兼容性加固字符串/数字 ID，纠正原 P1 未证实的判断；真实消息及 UNANSWERED 样本 NOT_RUN，禁止启用消息主题；#103 须重基后验收。未合并、部署、迁移、启用或订阅 | [#101](https://github.com/Fan6512/una-seller/pull/101)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45736784362)；本地验收报告及最小 Go 实验 |
| 2026-11-26 | 本地综合验收：六 PR 原验收 HEAD CI PASS；#101 限定代码 APPROVE 待授权（消息真实样本缺失禁止启用）；#103 真实投诉 sender/receiver 修复后复审发现动作归属及 mandatory 期限缺陷，原分支追加 `e8dda3d7` 待新 CI；#104 缺新订单发现/窗口、#105 缺动态待办筛选、#106 缺安全回放、#102 缺深色主题及实机测量，均未完成。#104/#105 的实际祖先落后上游修订，须逐一重基；保持全部 Draft，未合并/部署/迁移/启用/订阅/发布 | 本地 V0.7 综合验收报告；[#101](https://github.com/Fan6512/una-seller/pull/101)、[#103](https://github.com/Fan6512/una-seller/pull/103)、[修复](https://github.com/Fan6512/una-seller/commit/e8dda3d7fd11c7409bea2850e24b2f99ed9624c0) |
| 2026-11-26 | 负责人批准并行准备后续代码：WH-04 #101 Go/PG CI PASS；独立桌面动效 #102 CI 失败作业重跑全绿；WH-05 #103 → OPS-03 #104 → OPS-02 #105 → WH-06 #106 为依赖叠加候选，现有 CI 仅 main 目标，需逐一重基/重跑验收；缺失新订单发现、真实客服/纠纷/商品、missed_feeds、公网、深色主题与 Windows 实测未完成。全部 Draft，未合并、部署、迁移、启用、订阅或发版 | [#101](https://github.com/Fan6512/una-seller/pull/101)、[#102](https://github.com/Fan6512/una-seller/pull/102)、[#103](https://github.com/Fan6512/una-seller/pull/103)、[#104](https://github.com/Fan6512/una-seller/pull/104)、[#105](https://github.com/Fan6512/una-seller/pull/105)、[#106](https://github.com/Fan6512/una-seller/pull/106) |
| 2026-11-26 | 负责人批准 #100 Squash Merge，合并提交 `32067017`，候选 `76906f4` CI 四项与 PG 真停机/重启 PASS；WH-03 代码合并完成，Webhook 默认禁用，合并后 CI 待复验。下一步 WH-04 复用现有协调器；0014/0015 生产迁移、连接池容量、OPS-03 补偿、公网安全与订阅仍为启用门禁，本轮未部署或启用 | [#100](https://github.com/Fan6512/una-seller/pull/100)、[合并提交](https://github.com/Fan6512/una-seller/commit/320670173b3f2993cc938c746aac5a13aed9f783) |
| 2026-11-26 | #100 最终复审 `76906f4` APPROVE（默认禁用，待合并授权）；CI 四项/PG 真停机及重启实际 PASS，本地隔离测试及变异验证据报告 PASS；30/31 边界与跨范围测试已核对；main 基线后仅台账变化。停止扩大 PR，保留 30 单；未知关联及本地滞后补偿、连接池容量、生产迁移及公网/订阅为启用前门禁；未合并、部署或启用 | [#100](https://github.com/Fan6512/una-seller/pull/100)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45662044591)；本地报告 |
| 2026-11-26 | #100 真实 CBT 只读取证：货件 sender_id=0、无 site_id、含 order_id，订单 context.site；关联接口 401；旧 `9af6b09` CI 四项/PG 重启 PASS。复审直接修订 `be973233` 仅空关联/无权限兜底、已知同货件订单更新及晚到订单继承，待新 CI/PG。负责人确认实际多为 5–10 笔，保留 30 单硬上限、不做分批续跑；未知关联补偿、连接池启用门禁待完成。保持 Draft，未生产迁移/部署/订阅 | [#100](https://github.com/Fan6512/una-seller/pull/100)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45653803441)；负责人反馈及本地取证报告 |
| 2026-11-25 | 负责人要求先修公共地基，#100 原分支 `de3fe28d` + `92f86299` 新增 0015 应用级持久化协调、人工优先、延迟退还额度、明确身份冲突死信、晚到补查、有界 30 单、CI persistence 门禁及回归；新 HEAD 待验收，旧 CI/隔离报告仅历史；分批持久续跑、部分关联补偿和真实 CBT 取证未完成，保持 Draft 暂缓合并，未生产迁移/部署/启用 | 负责人决策；[#100](https://github.com/Fan6512/una-seller/pull/100) |
| 2026-11-25 | WH-03 #100 `739a0636` 候选 CI success（Changed Files/Go 实际 PASS，Desktop/MV3/PG restart SKIP），据隔离 PG17.6 本地报告全量及真停机重启 PASS；复审建议默认禁用合并、待负责人批准。人工优先、共享冷却/忙碌重试、真实 CBT 货件字段和关联权限、晚到/部分/超上限补偿及 CI persistence 路径门禁待后续；无生产迁移、部署、启用或订阅改动 | [#100](https://github.com/Fan6512/una-seller/pull/100)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45642046950)；本地验收报告 |
| 2026-11-25 | WH-03 [#100](https://github.com/Fan6512/una-seller/pull/100) Draft `08b96583`：基于 #99 注册订单/货件处理器，复用订单及货件写入、现有 orderGET；新版货件关联订单接口支持货件先到；精确归属、保留采购/账单/游标、时间防护及事务 fencing，新增模拟与隔离 PG 回归待验证。无新迁移/依赖；未部署、启用或改订阅 | 负责人推进授权；[WH-03 设计](https://github.com/Fan6512/una-seller/blob/08b96583d4521db5fb2e61bfd48bde938dca14b6/docs/architecture/WH-03-ORDER-SHIPMENT-NOTIFICATIONS.md) |
| 2026-11-25 | 负责人批准 #99 Squash Merge，已合并至 `5874f7c5`；候选 `481e974e` CI 四项 PASS（PG 停机 503/重启读回实际执行），WH-01/02 基础设施已进入 main、默认禁用。下一开发切片 WH-03 订单/货件处理器与既有共享限流；0014 生产迁移、部署、真实公网安全/通知/订阅仍未执行 | [#99](https://github.com/Fan6512/una-seller/pull/99)、[CI 45634604451](https://github.com/Fan6512/una-seller/actions/runs/45634604451)；负责人批准 |
| 2026-11-25 | #99 修订 HEAD `481e974e` CI 四项 PASS，Go 格式/vet/test/build、PG 持久化/停机 503/重启读回实际 PASS；main 自 PR 基线仅台账变化，可合并，保持 Draft 待负责人批准。未注册处理器前不得启用，生产迁移/部署/公网通知及订阅未执行 | [#99](https://github.com/Fan6512/una-seller/pull/99)、[CI 45634604451](https://github.com/Fan6512/una-seller/actions/runs/45634604451) |
| 2026-11-25 | OPS-01 #99 原分支复审修订 `481e974e`：暂不可用授权 503、未知/歧义 403；全部主题处理器启动门禁、积压/容量告警；保守调度文档化；不透明 ID、应用回执计数、百分位及 CI 命名修正，新增回归。旧 `fdcd2ae9` CI/隔离验收仅历史证据；新 HEAD 待 CI，保持 Draft，未生产迁移/部署/启用 | 负责人要求直接修复；[#99](https://github.com/Fan6512/una-seller/pull/99) |
| 2026-11-25 | OPS-01 #99 HEAD `fdcd2ae9` WH-01/02 代码完成，CI 四项 PASS、PG 真停机重启读回 PASS；据隔离报告回环 HTTP 最大 ACK 223ms/9.72ms、数据库故障 503。复审建议默认禁用状态合并、待负责人批准；WH-03 官方回读与共享限流待实现，WH-06 公网安全/真实通知/补偿待验收，未部署/迁移/启用/改订阅 | [#99](https://github.com/Fan6512/una-seller/pull/99)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45625829215)；开发验收报告 |
| 2026-11-25 | 负责人确认客户端后续统一 GitHub Actions 自动构建，先推进 Webhook；#98 版本 PR 候选 CI PASS、待批准；OPS-01 WH-01 #99 Draft `db56e44f` 已提交默认关闭的订单/货件通知持久化入口及迁移 0014，尚未验证 CI/隔离数据库/真实来源和延迟；WH-02 Worker、资源合并、重试及其他主题后续开发。未迁移、部署、修改官方订阅或发布 | 负责人决策；[#98](https://github.com/Fan6512/una-seller/pull/98)、[#99](https://github.com/Fan6512/una-seller/pull/99) |
| 2026-11-25 | 负责人确认实机账单持久化 PASS，其他专项暂缓、不阻塞测试版；#98 `db7d2aae` 已创建版本号 0.6.0 Draft（五个版本文件，+6/-6），待 CI、合并批准及单独签名发布审批，未打 Tag/未建 Release；真实净回款仍待 FIN-02 | 负责人实机反馈；[#98](https://github.com/Fan6512/una-seller/pull/98) |
| 2026-11-25 | V0.6 生产迁移 0013 与后端升级至 `0064a239`：合并提交 CI 重跑四项 PASS（含 PG 重启三步）；数据库备份 157,089 B、SHA `f017b38e…`，恢复演练 73 单指纹与 18 张表一致；迁移 `lock_timeout=3s` 18:16:20 成功，`billing_snapshot` 与 CHECK 校验通过、历史 73 单指纹不变；二进制 `eb6ec75c…`（`vcs.modified=false`）三处哈希一致，PID 815417→864371→866058、`NRestarts=0`；基础回归与 OTA 双通道 PASS（预发布首次 503 为 GitHub 瞬时失败，冷却后 200）；单笔订单 `2000019508270206` 官方读取→PG 持久化→billing-cache 读回→重启读回 PASS，3 次缓存读取未增加官方请求；真实净回款、真实 429、授权失效 NOT_RUN；V0.6.0 版本未改、未打 Tag/未建 Release | [脱敏部署报告](../architecture/PRODUCTION-V06-MIGRATION-DEPLOY-REPORT-20261125.md)；负责人批准 |
| 2026-11-25 | 负责人批准 #97 Squash Merge，已合并至 `0064a239`；候选 `a2570f61` CI 四项 PASS，隔离 PG 迁移及重启持久化验收 PASS；生产迁移 0013、后端升级、Windows 实机及 V0.6 发布仍待审批和执行，真实净回款 NOT_RUN | [#97](https://github.com/Fan6512/una-seller/pull/97)、[CI 45588642857](https://github.com/Fan6512/una-seller/actions/runs/45588642857)；负责人批准 |
| 2026-11-25 | #97 合流后 HEAD `a2570f61` CI 四项 PASS；PG17 写入/重启/读回三步均实际 PASS，GitHub 可合并，保持 Draft 待负责人批准；生产迁移/部署/客户端发布未执行，真实净回款 NOT_RUN | [#97](https://github.com/Fan6512/una-seller/pull/97)、[CI 45588642857](https://github.com/Fan6512/una-seller/actions/runs/45588642857) |
| 2026-11-25 | 负责人批准 #96 Squash Merge，已合并 `415c782a`；#97 原分支以 merge commit `a2570f61` 更新至该 main 基线，保留历史、不强推，核对两处共享桌面文件完整保留两边改动；新 HEAD CI 待确认，#97 保持 Draft，生产迁移/部署/客户端发布未执行，真实净回款 NOT_RUN | [#96](https://github.com/Fan6512/una-seller/pull/96)、[#97](https://github.com/Fan6512/una-seller/pull/97)；负责人批准 |
| 2026-11-25 | #96 `9b316722` 本地复审无阻塞，候选 CI PASS，待负责人批准合并；#97 `f16bcf3` 最新 CI 四项 PASS（PG 写入/重启/读回实际执行），据本地报告隔离 PG17.12 迁移及持久化、授权隔离、并发、冷却与失败保留 PASS；待 #96 合并后更新基线复验；真实净回款 NOT_RUN，Windows GUI、生产迁移、部署及发布均未执行 | [#96](https://github.com/Fan6512/una-seller/pull/96)、[#97](https://github.com/Fan6512/una-seller/pull/97)、[CI](https://github.com/Fan6512/una-seller/actions/runs/45587481720)；本地验收报告 |
| 2026-11-25 | #96 候选桌面/变更范围 CI PASS；FIN-01 A2 #97 独立 Draft 已提交单笔账单持久化与客户端本地读回代码，待 CI、隔离 PostgreSQL、Windows 及真实财务核账；不合并、不迁移、不部署，不改变 CBT C1 P0 | [#96](https://github.com/Fan6512/una-seller/pull/96)、[#97](https://github.com/Fan6512/una-seller/pull/97) |
| 2026-11-25 | 负责人实机反馈：#94 官方账单读取 PASS，持久化及官方总计/回款/利润未完成；#95 一笔订单状态由待发货更新为已发货 PASS，触发按钮未确认；第三方网页入口、订单列表状态去重待修。FIN-01/02 P1，ORD-04 UI 修复 P1，新增 ORD-07 P2 暂缓、不纳入 V0.7；不改变 CBT C1 P0。 | 负责人 2026-11-25 实机反馈及优先级确认 |
| 2026-11-25 | #95 生产后端升级 15:33:11（北京时间）至 `2f679dda`：目标提交 CI 重跑 PASS、候选与合并树后端一致、无新增迁移；二进制三处哈希一致、NRestarts=0、无回滚；healthz/鉴权/C0/订单/单笔账单/OTA 双通道 PASS；#95 物流专项单笔官方状态比对、10 笔上限/15分钟去重/终态排除现场 PASS，失败保留与真实 429 现场 NOT_RUN（代码/测试 PASS）。Windows GUI、V0.6 客户端发布与实机验收待执行，不标记 V0.6 完成 | [#95](https://github.com/Fan6512/una-seller/pull/95)；[CI 45564522734](https://github.com/Fan6512/una-seller/actions/runs/45564522734)；[脱敏部署报告](../architecture/GO-API-PR95-DEPLOY-REPORT-20261125.md)；负责人批准 |
| 2026-11-25 | 负责人确认只明确业务模块边界，暂不区分免费/付费；补全 V0.7 六类通知、运营中心、同步、登录安全及模块化任务，登记 V0.8/0.9 客服纠纷商品 UI、批量刊登和后续候选。维持原有 CBT/财务主线与 V0.6 交付范围；本次仅需求台账，不改代码/技术合同/官方订阅 | 负责人 2026-11-25 决策 |
| 2026-11-25 | #95 订单物流联动同步已 Squash Merge，候选 CI PASS，业务代码树与候选一致，保留已有台账更新；无新增 SQL 迁移，生产升级、V0.6 客户端签名发布及实机验收待执行 | [#95](https://github.com/Fan6512/una-seller/pull/95) 2f679dda；[CI](https://github.com/Fan6512/una-seller/actions/runs/45558640026)；负责人推进授权 |
| 2026-11-25 | #94 生产升级 13:56:07（北京时间）至 86fcb584，main CI attempt 2 PASS；服务/鉴权/订单/单笔账单/OTA PASS（现场报告），无迁移及回滚，真实429未发生。#95 候选 CI PASS、Draft。负责人确认 V0.7 六类 Webhook、运营中心、低频补偿同步和启动登录安全，沿用现有通知 Callback URL；未实施 V0.7、未更改官方订阅、未发布 V0.6 | [部署报告](../architecture/GO-API-PR94-DEPLOY-REPORT-20261125.md)；[#95](https://github.com/Fan6512/una-seller/pull/95)；[45558640026](https://github.com/Fan6512/una-seller/actions/runs/45558640026)；负责人确认 |
| 2026-11-25 | V0.5 实机反馈修订 #94 合并：逐变体独立编辑/图片、旧描述隐藏、MODEL 动态必填提示、Track718/17TRACK 查询、账单共享冷却/Retry-After CORS/脱敏来源诊断；最终候选 CI PASS，合并树一致。新订单主图负责人暂通过但真实新单 NOT_RUN；账单限流来源及物流独立窗口待复验。未部署 #94、未发布 V0.6 | [#94](https://github.com/Fan6512/una-seller/pull/94) 86fcb584；[CI](https://github.com/Fan6512/una-seller/actions/runs/45551103481)；负责人推进授权 |
| 2026-11-25 | V0.5 Windows 签名测试版发布为 Pre-release（`desktop-test-v0.5.0`，0.5.0）：PR #93 合并提交 `c55b322` 后 main CI PASS，`desktop-production` 审批通过，签名构建 run 45542860330 PASS；安装包 SHA-256 `898ec6ae…`、minisign/Ed25519 验签 PASS、Key ID 与 V0.4 一致，`latest.json` 与五处版本文件一致；公网默认通道 204、测试通道缓存 TTL 刷新后返回 0.5.0，均 `no-store`；未设置 Latest、未部署 Go API、未重建未签名 QA 包；实机 OTA 与功能待负责人验收 | [#93](https://github.com/Fan6512/una-seller/pull/93)；[构建 run](https://github.com/Fan6512/una-seller/actions/runs/45542860330)；[Release](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.5.0)；[发布报告](../architecture/DESKTOP-V05-TEST-RELEASE-REPORT-20261125.md) |
| 2026-11-25 | V0.5 生产 Go API 10:51:41（北京时间）切换至 `1bb50192`：恢复点、服务与哈希、鉴权/C0/OTA、同机业务、订单与账单只读抽测、工作台读取及历史数据比对 PASS；NRestarts=0、无回滚、无新增迁移；实时物流/SLA 刷新、独立费用对账、旧客户端全流程、GUI/跨版本 OTA NOT_RUN。签名测试版待授权 | 负责人现场部署报告；脱敏完整报告本地提交 `bcc5751` 尚未推送，仓库部署 SHA 已独立核对 |
| 2026-11-25 | V0.5 功能开发收口：#90 订单修订、#49 A1 官方账单、#92 工作台及 Windows 洁净构建全部合入 main；候选 CI 和 #92 Windows QA PASS；未部署生产、未发布 V0.5、未人工验收 GUI/OTA。真实 Family/刊登 payload、净收益/实际到账/分摊及 A2 保持待完成 | [#90](https://github.com/Fan6512/una-seller/pull/90) fb2b438f；[#49](https://github.com/Fan6512/una-seller/pull/49) 954644b1；[#92](https://github.com/Fan6512/una-seller/pull/92) 6b7c2c54；[Windows QA](https://github.com/Fan6512/una-seller/actions/runs/45532958238)；负责人批准 |
| 2026-11-24 | 负责人批准 V0.5 FIX-01～08：美元销售额、后台销售编号、净收益及详情费用、新订单主图自动补全、官方物流受限网页弹窗、北京时间、精确 SLA、综合履约状态及服务端筛选。#90 已开始 FIX-06/07/08（候选 CI 待确认），FIX-01/02/03 待真实只读取证，FIX-04/05 待实现；财务沿 #49，未合并/部署/发布 | [#90](https://github.com/Fan6512/una-seller/pull/90)、[#49](https://github.com/Fan6512/una-seller/pull/49)；负责人确认 |
| 2026-11-24 | V0.4 Windows 部分实测：67 项中 22 PASS、4 FAIL、41 待验收。DESK-01 五项通过，实际跨版本 OTA 待测；ORD-04 9/2/1，ORD-01～03 4/2/3，CBT-01 四项通过；COL-02、WB-01～03、PUB-01/02、CAT-01、TR-01、AI-01/02 未验收。订单金额与后台编号对照优先（本轮 P0）；物流链接、签收时区、缺图补全、SLA 展示 P1；净收益沿 FIN-01/02 推进，不伪造。保持全局 CBT 首次刊登 P0，不擅自开发或发布 | 负责人 2026-11-24 V0.4 实机报告；[#66](https://github.com/Fan6512/una-seller/pull/66)、[#89](https://github.com/Fan6512/una-seller/pull/89) |
| 2026-11-24 | DESK-01 公网 OTA 接口 PASS，两种通道均返回 `Cache-Control: no-store`。Windows 实机在线升级待验收 | [Nginx 路由模板](../../deploy/una-seller/nginx-mkd.conf) 与生产验收。`nginx -t`、reload 和公网复测通过，测试版 200（`0.4.0`）、默认通道 204，Go API PID 一致 |
| 2026-11-24 | V0.4 Windows 签名测试版发布为 Pre-release（`desktop-test-v0.4.0`，0.4.0）；安装包 SHA-256 `0ab9c688…`，更新签名 minisign 验证 PASS；Windows 实机与 OTA 待验收；发布时公网更新接口为 Nginx 404 | [#89](https://github.com/Fan6512/una-seller/pull/89)；构建 run 45364449099；[Release](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.4.0) |
| 2026-11-24 | ORD-04 生产后端验收 PASS（0010、9 单真实物流与缓存）；批准 V0.4 签名测试版，OTA/GUI 待验收，未正式发布 | [#66](https://github.com/Fan6512/una-seller/pull/66)；[脱敏验收摘要](../architecture/ORD-04-DEPLOY-REPORT-20261124.md) |
| 2026-11-24 | 负责人批准普通台账直接提交 main；重要设计、合同、代码与治理变更仍走 PR，保留并发保护及正常 CI | PR #88 负责人决策 |
| 2026-11-24 | CBT-01 单账号真实 C0 生产验收完成；CBT-02 P0 进入开发，实施 PR 待建；同步更新证据边界与流程 | [#87](https://github.com/Fan6512/una-seller/pull/87)；[PR #87 生产升级与真实 C0 验收摘要](../architecture/GO-API-PR87-DEPLOY-REPORT-20261124.md)；生产数据库最新回读 14:39:47 北京时间 |
| 2026-11-24 | 新增 DESK-03 桌面端图片预览及加载速度优化（P1，待开发）；第一阶段限定前端低分辨率预览、失败回退、按需加载/有限预取、变体延迟渲染和脱敏耗时测量，保留原始高清 URL；与 #79/#82 数据库优化分离 | [Issue #83](https://github.com/Fan6512/una-seller/issues/83)；2026-11-24 负责人需求 |
| 2026-11-23 | 初始化功能总台账；盘点 main、58 个已关闭 PR、[#66](https://github.com/Fan6512/una-seller/pull/66)、[#49](https://github.com/Fan6512/una-seller/pull/49) 与公开 Issue；将功能完成和真实验收分离 | main e9b29c8f；当前官方类目 11,668/714/10,954；截至本日已提供的实测报告 |


## 九、业务流程图（Mermaid，可随功能与 PR 更新）

> 使用上下结构的 Mermaid flowchart TD；GitHub 直接渲染，原始代码可复制修改。**这是目标业务流程而非全部已上线功能的声明。** 业务节点旁的功能 ID 对应本台账；每张图下说明目前哪些节点仍待真实验收或开发。实线代表业务主线，虚线代表附带字段、证据与关联信息。

### 9.1 总体业务主流程

~~~mermaid
flowchart TD
    A["外部商品来源（Amazon 等）"] --> B["浏览器插件采集"]
    B --> C["UNA Seller 采集箱"]
    C --> D["整理工作稿：补充真实商品资料"]
    D --> E["AI 文案与多语言候选"]
    E --> F["人工审核并保存"]
    F --> G["独立产品箱"]
    G --> H["CBT 类目 / 属性 / 图片 / 站点资料确认"]
    H --> I["选择店铺及获准站点"]
    I --> J["CBT 发布前校验"]
    J --> K["官方创建与逐站点结果回读"]
    K --> L["已刊登商品管理"]
    L --> M["订单同步"]
    M --> N["1688 采购与订单关联"]
    N --> O["国内物流 → 货代 → 官方物流"]
    O --> P["买家签收"]
    P --> Q["账单、回款与利润核算"]

    D -.-> D1["规格 / 图片 / 成本 / 尺寸重量 / 备注"]
    E -.-> E1["EN / ES-MX / PT-BR"]
    K -.-> K1["success / failed / needs_check"]
    O -.-> O1["国内运单与官方运单分开记录"]

    classDef done fill:#e6f5e9,color:#245b32,stroke:#63a773
    classDef planned fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class B,C,D,G,M done
    class J,K,N,O,Q planned
~~~

**当前边界**：采集、工作稿、产品箱、订单读取有基础代码；AI/翻译待真实凭据验收；J、K 的真实 C1 刊登、完整物流与利润闭环仍待完成。

### 9.2 采集 → 整理工作稿 → 产品箱（COL / WB）

~~~mermaid
flowchart TD
    A["来源平台商品详情"] --> B["浏览器插件采集"]
    B --> C{"图片、价格、规格等是否取全？"}
    C -->|完整| D["采集箱保存来源快照"]
    C -->|部分成功| E["采集箱保存已取得数据并提示缺项"]
    E --> D
    D --> F["打开整理工作稿"]
    F --> G["人工补充或修改商品事实"]
    G --> H["可选：调用 AI 生成候选文案"]
    H --> I["人工选择采纳内容"]
    I --> J["保存工作稿与 revision"]
    J --> K{"确认转入产品箱？"}
    K -->|否| F
    K -->|是| L["复制生成独立 OwnProduct"]
    L --> M["产品箱编辑与保存"]
    M --> N["关闭并重新打开：确认读回"]

    D -.-> R["来源标题 / 图片 / 变体 / 原始描述"]
    G -.-> S["图片排序、主图、SKU、成本、重量尺寸、属性"]
    L -.-> T["来源快照不被覆盖；重复转库不重复建产品"]

    classDef verify fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class C,E,N verify
~~~

**当前边界**：COL-01 的基本采集曾实测；COL-02 的高清图片和完整变体、WB-01～03 的 #92 新版保存/转库、变体成本与副图复制已合并，Windows 自动化 QA PASS；真实 GUI 保存/转库读回仍须验收。

### 9.3 CBT 类目与中文阅读翻译（CAT / TR）

~~~mermaid
flowchart TD
    A["读取 Mercado Libre 官方 CBT 类目 Dump"] --> B["离线词典保存官方 ID / 英文路径"]
    B --> C["筛选有效类目与停用类目"]
    C --> D["配置 Azure 或百度翻译服务"]
    D --> E["分批中文翻译并记录进度"]
    E --> F{"该条译文有效吗？"}
    F -->|是| G["保存对应类目 ID 的中文路径"]
    F -->|无法安全翻译| H["skipped：保留官方英文"]
    F -->|外部调用失败| I["failed：保留进度，等待重试"]
    G --> J["发布有效词典供客户端读取"]
    H --> J
    I --> E
    J --> K["工作稿 / 产品箱中文分类路径"]
    K --> L["选择官方类目和属性并保存官方 ID"]

    M["订单标题 / 动态属性 / 枚举原文"] --> N["按需中文阅读翻译"]
    N --> O["界面优先中文，保留原文"]
    O --> L

    classDef pending fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class D,E,F,G pending
~~~

**当前边界**：#70、#72 实现了代码；官方原文 11,668 条、停用 714 条、**有效 10,954 条的中文译名目前仍为 0**。中文阅读翻译不改变官方 ID/属性值，不等于买家文案本地化。

### 9.4 AI 三语标题、卖点与详情（AI / TR）

~~~mermaid
flowchart TD
    A["工作稿 / 产品箱已保存的商品事实"] --> B["核对材质、尺寸、规格与实际功能"]
    B --> C["AI 生成 EN / ES-MX / PT-BR 标题候选"]
    C --> D["人工检查准确性及站点字符限制"]
    D --> E["逐语言采纳并保存标题"]

    B --> F["AI 生成英文卖点与英文详情"]
    F --> G["卖点与描述独立采纳"]
    G --> H["保存英文母版详情"]
    H --> I["从已保存英文母版翻译 ES-MX / PT-BR 详情"]
    I --> J["检查数字、单位、型号与歧义"]
    J --> K{"译文是否可采纳？"}
    K -->|否| L["保留候选，人工修改或重新生成"]
    L --> J
    K -->|是| M["分别保存西语 / 葡语详情及来源摘要"]
    M --> N{"以后英文母版发生变化？"}
    N -->|是| O["提示现有译文可能过期，人工复核"]
    N -->|否| P["保留已采纳的三语文案"]
    O --> I

    C -.-> C1["SEO 本地化，不逐字机翻标题"]
    G -.-> G1["不自动覆盖已保存卖点或详情"]
    J -.-> J1["如 5.5 与 5,5；千分位歧义需人工核对"]

    classDef verify fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class C,F,I,J verify
~~~

**当前边界**：#71 代码已合并，真实 AI Key、真实已保存商品和 Windows GUI 生成/采纳/重开读回尚未验收；中文阅读翻译和买家向 ES/PT 文案是不同业务链路。

### 9.5 产品箱 → CBT C1 正式刊登（CBT / IMG / PUB）

~~~mermaid
flowchart TD
    A["产品箱选择待刊登商品"] --> B["选择已授权店铺及 CBT 站点"]
    B --> C["C0 只读探针核验账号发布模型"]
    C --> D["根据账号能力确定官方创建接口"]
    D --> E["按站点确认售价、币种、SKU 与变体"]
    E --> F["检查类目、属性、三语标题和描述"]
    F --> G["处理并上传商品图片，取得官方图片 ID"]
    G --> H{"发布前官方必填校验通过？"}
    H -->|否| I["定位字段错误，返回产品箱修改"]
    I --> E
    H -->|是| J["人工确认发布"]
    J --> K["服务端创建本地发布意图与 creation_id"]
    K --> L["调用 CBT C1 官方创建接口"]
    L --> M{"官方创建结果是否明确？"}
    M -->|确定成功| N["逐目标站点回读成功对象"]
    M -->|明确失败| O["记录失败站点、原因与字段"]
    M -->|超时或结果未知| P["needs_check / pending_verify"]
    P --> Q["先只读查询官方结果，禁止盲目重发"]
    Q --> M
    N --> R["已刊登商品列表关联真实官方 ID"]
    O --> S["仅修正及重试未成功的目标站点"]

    C -.-> T["#85 + #87：单账号真实 C0 PASS；standard_up"]
    L -.-> U["C1 开发启动；真实 POST 须单独批准"]

    classDef planned fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class D,E,G,H,J,K,L,M,N,O,P,Q,R,S planned
~~~

**当前边界**：#85、#87 单账号真实 C0 已验收，PUB-05A #115 已合并，单/多规格核心、持久尝试与回读防重已有代码，联合部署待批准；CBT-02 暂缓，图片上传、官方创建、真实逐目标回读及剩余售价模板尚未验收或完成。此图是目标流程，**不能据此发起真实 POST**。

### 9.6 已刊登商品同步与后续维护（PUB）

~~~mermaid
flowchart TD
    A["已经存在的 Mercado Libre 官方商品"] --> C["UNA Seller 只读同步已刊登商品"]
    B["由 C1 新创建的官方商品"] --> C
    C --> D["按账号 / 站点 / SKU 查看商品列表"]
    D --> E["读取详情、分类、属性、图片与描述"]
    E --> F{"需要修改商品？"}
    F -->|否| G["继续只读同步 / 查询状态"]
    F -->|是| H["保存本地编辑草稿"]
    H --> I["检查账号权限与官方可写字段"]
    I --> J["人工确认待修改内容"]
    J --> K["调用官方对应更新接口"]
    K --> L["回读实际更新结果"]
    L --> M{"官方是否确认成功？"}
    M -->|成功| N["更新本地发布快照"]
    M -->|失败| O["保留原值与错误原因，人工修正"]
    M -->|未知| P["needs_check：先核实，禁止重复写入"]

    D -.-> Q["只读同步/本地草稿已有代码"]
    K -.-> R["官方商品价格/库存/图片/文案写回尚未完成"]

    classDef planned fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class I,J,K,L,M,N,O,P planned
~~~

### 9.7 订单采购 → 双物流 → 签收（ORD）

~~~mermaid
flowchart TD
    A["买家在 Mercado Libre 下单"] --> B["UNA Seller 同步订单"]
    B --> C["卖家在 1688 采购"]
    C --> D["厂家发货"]
    D --> E["国内快递运输"]
    E --> F["货代仓库签收"]
    F --> G["货代转运 / 交接"]
    G --> H["Mercado Libre 官方扫描发货"]
    H --> I["官方物流配送"]
    I --> J["买家签收"]
    J --> K["订单结算与回款"]

    C -.-> L["1688 采购订单号"]
    E -.-> M["国内运单号 / 物流状态"]
    H -.-> N["Mercado Libre 官方运单号"]
    I -.-> O["官方配送状态"]

    K --> P["账单 / 利润核算"]
    J --> Q{"是否发生售后？"}
    Q -->|是| R["退款 / 退货 / 纠纷"]
    Q -->|否| S["订单完成"]

    classDef key fill:#e9f2ff,color:#163d73,stroke:#6a9cda
    classDef complete fill:#e6f5e9,color:#245b32,stroke:#63a773
    class F,H key
    class J complete
~~~

**当前边界**：#40/#43/#55 的订单读取和搜索、#46/#54 的**人工采购关联**已合并；#66 官方 Shipment 已合并，生产后端 PASS，V0.4 GUI 部分通过，金额/编号、图片、物流链接及时区等修订已由 依据现场报告 #90 已生产升级且抽测 PASS，实时物流/SLA 刷新及 V0.5 实机待复验；1688 自动授权/采购物流及货代自动跟踪未形成完整链路。图中的厂家、货代、买家节点描述卖家真实履约流程，不表示系统已经自动感知全部状态。

### 9.8 账单、利润与售后（FIN / CS）

~~~mermaid
flowchart TD
    A["Mercado Libre 订单及官方结算数据"] --> B["读取官方账单明细"]
    B --> C["核对费用 / 税项 / 退款 / 冲正"]
    C --> D["匹配同一订单的真实净回款"]
    D --> E["读取 1688 采购成本及其他可信成本"]
    E --> F["获取有来源和日期的汇率"]
    F --> G{"数据是否已完整核对？"}
    G -->|否| H["显示待核账，不虚构净收益或利润"]
    H --> I["补充回款、成本或汇率证据"]
    I --> G
    G -->|是| J["计算 USD 利润与利润率"]
    J --> K["订单利润卡片 / 经营统计"]

    A --> L{"是否发生退款或争议？"}
    L -->|是| M["读取官方退款 / 纠纷状态"]
    M --> N["核实退款金额、费用冲正和分摊"]
    N --> C
    L -->|否| G

    B -.-> O["#49：A1 已生产部署，账单读取抽测 PASS；独立对账及 GUI 待验收"]
    J -.-> P["真实净回款和利润自动归集尚未完成"]

    classDef verify fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class B,C,D,E,F,G,J,K,M,N verify
~~~

### 9.9 Windows / macOS 客户端交付与在线更新（DESK）

~~~mermaid
flowchart TD
    A["功能代码合入 main"] --> B["GitHub 按范围运行 CI"]
    B --> C{"相关检查通过？"}
    C -->|否| D["修复并重新提交 PR"]
    D --> B
    C -->|是| E["构建候选客户端"]
    E --> F["自动化 QA / 后端版本与部署核对"]
    F --> G{"候选构建与必要门禁通过？"}
    G -->|否| D
    G -->|是| H["经审批创建签名 Release"]
    H --> I{"用户开启接收测试版？"}
    I -->|否| J["仅检测正式 Release"]
    I -->|是| K["可检测正式版与合规的预发布版"]
    J --> L["Tauri 下载并验证签名"]
    K --> L
    L --> M{"验签和版本检查通过？"}
    M -->|否| N["拒绝安装，保留当前版本"]
    M -->|是| O["用户确认更新、安装并重启"]

    F -.-> P["V0.5 签名已发布；#115/116/117 已合并、联合升级待批准"]
    H -.-> Q["新版签名测试版待批准；GUI/安装/OTA、macOS 独立验收"]
    O --> R["负责人真机功能验收"]

    classDef verify fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class F,G,H,L,M,O verify
~~~

**更新纪律**：绿色 CI、Windows 构建、手工 GUI、服务端部署、签名 OTA 属于不同验收项。测试版开关代码已合并，不代表当前正式签名更新渠道已经实测通过。

## 十、流程图维护方式

- 新增需求若改变业务顺序：优先修改对应章节的 Mermaid 节点与箭头，不重新制作图片；同一流程可以保留稳定功能 ID。
- 不把目标流程中的橙色/待办节点当作“已完成”；对应功能 PR 合并且真实验收通过，再更新该流程下的状态说明与上方台账。
- 新 PR 只影响某个模块时，仅维护该模块的台账行、关联 PR、流程图及下方变更记录，无须大规模重写其他图。



### DESK-06 依赖维护

Ding React 与 Core 均固定 npm 版本 0.0.9，MIT 许可随构建分发。业务仅依赖 `notify` / `batch` / `task`，独立渲染适配层承接 Ding。后续优先复用 Dependabot；升级时核实上游接口和许可、提交真实 npm 锁文件、独立 PR 经现有 CI 及必要 Windows WebView2 验收后人工合并，不复制上游源码，不自动合并。


### 历史裁决：2026-11-26 负责人业务优先级收口（以第六节为当前执行顺序）

DESK-05/06/07 已由 #114 合并，本阶段停止客户端视觉优化、不重复构建或验收。业务优先：WH-06 D1 真实联调 → OPS-03/D2 订单/货件自动补偿；并行不阻塞 CBT-02 首次单站点真实刊登最小闭环（开发/模拟先行，官方写入须明确批准）；OPS-05 先真实销售额，FIN-02 核实净回款/退款/费用/汇率/成本再接真实利润。只做必要的业务 UI，不新增通用插件平台、服务或复杂测试。DESK-04 等非阻塞客户端事项后置。


### 历史裁决：2026-11-26 20:36 业务推进裁决

Webhook 为主路径、低频自动补偿兜底。顺序：OPS-03 D2 生产启用及真实联调 → PUB-05A 多变体核心 → CBT-02 首次真实刊登（不等待 D1）→ OPS-05 真实销售额 → FIN-02 真实利润；PUB-05B 多店多站点批量后置。继续冻结非必要客户端视觉优化。生产启用与官方刊登写入必须分别核对并取得授权。


### 历史裁决：2026-11-27 00:18 负责人调整顺序

PUB-05A 已合并，CBT-02 因真实可销售商品尚未准备好而暂停，不做真实刊登准备、不要求临时测试商品。OPS-05 真实销售额优先，随后 FIN-02 真实利润；WH-06 货件真实样本可独立核对 ACK，OPS-03 继续正常生产运行。待商品准备好再恢复 CBT-02，所有既有生产迁移/部署及官方写入授权门槛保留；后续可合并后端升级窗口，具体部署另行批准。

### 2026-11-28 TR-02E 最终生产验收（本阶段更新）

- 生产升级 `97b31a03`（#131 已合并）：SHA-256 `0fcb04fe…`、`vcs.modified=false`、仅重启 `una-seller-api`、`NRestarts=0`、本机/公网健康 200、匿名 401；备份 `api-20261128T053852Z-pre-97b31a03` 含二进制/api.env/unit/完整词典；无迁移、未重建桌面。
- managed 保护：临时目录实测新二进制在持锁后、联网前拒绝普通 `update`（28ms），词典文件零改动。
- 24 条纠错 80→104，逐条以 `staging.json` 官方完整路径核验；原 80 条字节级保留。
- 历史排除清单 1,214 条冻结（SHA-256 `44538e7c…`），来源与交集已披露：最终 300 与历史 544、首批 200 均为 0 重叠，与 TR-02D 第二轮 200 重叠 10。
- 离线 assemble（seed `tr02e-final-v1`）：`selected=300`、与排除 0 重叠、104 覆盖全部生效、官方身份 0 变更；原始候选 `0845a519…`、审查报告 `d1da436d…`。
- 300 条复验 289/300（96.33%），11 条错误（9 中文误译 + 2 英文回退）逐条纠正，覆盖 115 条；最终候选 `b3b9f46d…`，10,928/10,954（99.76%），26 条英文延期。
- publish Dry-run PASS（`applied=false`、`original_ok=289`、`audited=300`），active 未变、未写标记；**正式发布等待负责人单独批准**。
- 事故披露：13:57Z 保护测试因 `EnvironmentFile` 覆盖 `--setenv` 误指向生产目录，普通 update 运行约 24 秒，新增 103 条 progress、普通额度 192,758→275,291；progress 已按字节还原，额度不回退。详见验收报告第 1.5 节。

### 2026-11-28 14:36 正式发布回执（负责人已批准）

- 批准后在**同一数据目录、同一审核材料**执行 `cbt-categories publish … -apply`：`applied=true`，active `be52b174…` → **`b3b9f46d…`**，`original_ok=289`、`audited=300`。
- 备份 `llm-release-backup-be52b174…json`（哈希 `be52b174…`，4,022,647 B）；已写 `llm-release-managed.json`（`{"managed":true}`）。
- 生产实测普通 `update` 被拒（129ms，含 sudo/systemd 开销），普通额度保持 275,291 未变。
- 14:36:56 CST 重启 `una-seller-api` 验证加载：PID 2252694、NRestarts=0、本机/公网 `/healthz` 200、匿名 401、启动日志 `state=loaded translated=10927`（10,928 条非空译名中 1 条与英文名相同）。
- 未重建桌面客户端、未改数据库、未启用词典增量 timer（仅原有数据库备份 timer）。
- 回滚材料就绪：`cbt-categories rollback -active-sha256 b3b9f46d… -backup-sha256 be52b174…`（默认 dry-run，另行批准才 `-apply`）。
