# UNA Seller · 功能状态总台账

> **日常功能进度与需求登记的统一入口** · 首次盘点：2026-11-23  
> 核对基线：2026-11-25 main 业务基线 = 0064a2391ed0fe7aff2768053f5aa4e4b0f18e8c（#90/#49/#92/#94/#95/#96/#97 已合并）；V0.5 已发布签名测试版，负责人反馈部分实机问题，#94 修订 CI PASS；据负责人现场报告，2026-11-25 13:56:07（北京时间）生产 Go API 已升级至 `86fcb58468e353971a9a40a40b18a8aeb831f16c`，V0.5 签名客户端已发布；#95 已于 2026-11-25 15:33:11（北京时间）升级生产 Go API 至 `2f679ddad896b9361096d882e71a4ad8b20ff8b4`，生产后端 PASS；V0.6 生产迁移 0013 及 Go API `0064a239` 已升级、生产账单持久化 PASS，负责人确认实机账单持久化 PASS，其余 Windows 专项暂缓、不阻塞签名测试版；[#98](https://github.com/Fan6512/una-seller/pull/98) 版本号 Draft 待 CI 和审批。ORD-04、CBT C0 和 V0.4 Windows 历史真实验收保留原范围，其余功能未重新验收。  
> 产品范围：自用的 Mercado Libre CBT ERP；核心链路 **采集 → 工作稿/产品箱 → AI 整理 → CBT 刊登 → 已刊登管理 → 订单/财务**。Amazon 等平台只作为采集来源。  
> 本文件记录功能进度与验收，不替代 [产品定义](PRODUCT-WORKFLOW-AND-EDITOR.md)、[技术路线](../architecture/DELIVERY.md)、具体 PR/Issue 或官方 API 取证。发现记录与实际代码冲突时，先核实当前 main、PR 和真实验收证据，再更新本台账。

## 一、维护约定（保持简单）

- **固定 ID**：一个独立的用户功能一个 ID；新需求新增行，已有需求只更新原行，不反复生成新的任务书或重复 Issue。被延期或取消的条目仍保留原因与历史 PR。
- **已完成**：约定的最小功能范围已并入 main，且该范围的真实业务/操作验收有明确 PASS；不能把“CI 通过”当“业务已完成”。已完成的范围很小也要如实限定。
- **进行中**：有 Draft PR 在开发，**或者**代码已合并但真实 API、凭据、数据、Windows GUI/部署仍未验收；负责人明确批准启动开发时也可登记为进行中，但须标注实施 PR 待创建、未实现。
- **待完成**：已确认需求但尚无完成该业务目标的代码；早期只读能力、设计合同或 Mock 不等于正式实现。
- **暂缓/不做**：当前优先级之外，不主动开 PR；改变优先级须由负责人确认。
- **每条关联**：功能 ID、状态、关联 PR（合并 PR 或 Draft PR）、验收/剩余事项。PR 合并时补编号和 main 状态；真实验收通过后再移至“已完成”。涉及发布、迁移、生产变更始终单独授权。
- **新增需求时**：先在对应模块新增一行，写清最小可验收结果；确定开发时关联 PR；PR 合并和实测后更新状态。
- **普通台账直接提交 main（2026-11-24 负责人批准）**：仅限本台账的需求登记、状态、关联 PR/Issue、验收结果、经负责人确认的优先级、流程图进度同步，以及不改变合同的脱敏验收摘要归档。已有功能 PR 可随该 PR 更新台账；不再为普通进度单独建分支或 PR。
- **重要变更仍走 PR**：业务代码、重大产品需求与业务规则、架构/API 合同、数据模型/迁移、安全/权限、部署配置、CI 及仓库治理规则变更不得夹带在普通台账直接提交中；发布、迁移、生产操作仍须单独授权。
- **直接提交保护**：先读取最新 main 和事实依据，只修改相关条目与变更记录，检查差异及敏感信息；以最新 SHA 为预期父提交，使用非强制更新，遇并发或冲突重新读取并保留其他修改。保留 Git 提交历史及正常 CI，不绕过分支保护、不改保护设置；若 main 拒绝直接提交，则回退为最小文档 PR。仅在本次已授权操作中执行，不在对话结束后自动后台修改仓库。
- **优先级**：P0 = 首次真实 CBT 刊登必需；P1 = 自用稳定性与订单经营必需；P2 = 增强/非主线；暂缓 = 不抢当前资源。优先级是工作顺序，不是开发承诺。

## 二、关键事实（V0.5 代码更新至 2026-11-25，部署及真实业务沿用已有验收）

| 指标 | 经核对的情况 |
| --- | --- |
| main 业务基线 | **2f679dda**，[#90](https://github.com/Fan6512/una-seller/pull/90)（fb2b438f）、[#49](https://github.com/Fan6512/una-seller/pull/49)（954644b1）、[#92](https://github.com/Fan6512/una-seller/pull/92)（6b7c2c54）已 Squash Merge；#92 候选 CI [45532465052](https://github.com/Fan6512/una-seller/actions/runs/45532465052) 4/4 PASS，Windows QA [45532958238](https://github.com/Fan6512/una-seller/actions/runs/45532958238) PASS；合并后 main CI [45534058512](https://github.com/Fan6512/una-seller/actions/runs/45534058512) 因后续台账提交并发取消；已独立核实最终候选与 squash 业务代码树一致、台账提交仅修改文档，[45534183001](https://github.com/Fan6512/una-seller/actions/runs/45534183001) 文档范围 PASS；[#94](https://github.com/Fan6512/una-seller/pull/94)（86fcb584）已 Squash Merge，候选 [45551103481](https://github.com/Fan6512/una-seller/actions/runs/45551103481) Go/桌面/变更范围 PASS，插件按范围跳过，合并代码树一致；[#95](https://github.com/Fan6512/una-seller/pull/95)（2f679dda）已 Squash Merge，候选 CI [45558640026](https://github.com/Fan6512/una-seller/actions/runs/45558640026) PASS，业务树一致，仅保留合并前的台账更新 |
| 仍开放的代码 PR | #95 已合并；另有 [#76](https://github.com/Fan6512/una-seller/pull/76) 文档 Draft（不属本轮合并范围） |
| 分类官方原文 | 11,668；停用 714；有效待译 10,954；隔离目录已初始化 |
| 分类中文翻译 | **0/10,954**；真实翻译提供方凭据未配置，五批试译未执行 |
| AI 真实服务 | 三语标题/详情代码已合并；真实 Key、已保存测试商品与 Windows 操作验收尚缺 |
| CBT 刊登 | C0 生产只读取证 PASS，当前店铺分类 standard_up；C1 已进入开发阶段，实施 PR 待建，真实商品创建、官方图片上传及发布记录未验收 |
| CBT 测试账号 | 一个真实账号、五站点 MLB/MLM/MLC/MCO/MLA 均为 remote、UP enabled、net_proceeds；business_model 脱敏为 other。最新只读回读 2026-11-24 14:39:47（北京时间），原始枚举及卖家后台交叉核实仍待完成 |
| 桌面端 | Windows 已覆盖安装 V0.4；V0.4 签名 Pre-release `desktop-test-v0.4.0` 已发布（构建 run 45364449099，更新签名 minisign 验证 PASS）；公网 OTA 接口 PASS。覆盖安装、登录保留、检查更新按钮及测试版开关、重启保留 PASS；实际跨版本 OTA 下载/验签/安装待下一版本验收。V0.5 三个功能 PR 已合并；#92 最新 HEAD 的隔离未签名 NSIS QA 构建 PASS（安装包版本仍为 0.4.0，loopback API），不等于 V0.5 已发布或 GUI/OTA 已验收。V0.5 签名测试版已发布为 Pre-release（`desktop-test-v0.5.0`，0.5.0；构建 run 45542860330），安装包 SHA-256 `898ec6ae…`、minisign/Ed25519 验签 PASS、公网测试通道缓存刷新后返回 0.5.0、默认通道仍 204；负责人已反馈 V0.5 部分实机问题；#94 已合并且 CI PASS，#95 生产后端已升级 PASS（物流联动 GUI 待验收），下一测试版及剩余 GUI、实际跨版本 OTA 待验收。AI/词典整体 QA 和 macOS 实机仍待验收 |
| 部署 | #66 Go API 823e6056 已生产部署（据负责人报告）；0010 已迁移，0011/0012 保留；9 单真实 Shipment 后端验收 PASS。C0 已回读；V0.4 四标签及手动 C0 实测 PASS，自动重授权探测未验收；ORD-04 Windows 部分通过，#90 修订已合并、待新版实测；据负责人现场报告，#90/#49/#92 已于 2026-11-25 10:51:41（北京时间）集中升级生产 Go API 至 `1bb50192`，无新增迁移；服务正常、NRestarts=0、无回滚，构建/磁盘/运行进程哈希一致；旧二进制双备份校验、新加密数据库归档隔离恢复 PASS（直连备份挂起后改用生产会话池）；回环/公网、鉴权、C0、OTA 双通道及同机业务 PASS；#94 已于 2026-11-25 13:56:07（北京时间）生产升级至 86fcb584，main CI [45555320846](https://github.com/Fan6512/una-seller/actions/runs/45555320846) attempt 2 PASS；健康/鉴权/C0/订单/单笔账单/OTA 双通道 PASS（现场报告），无新迁移、无回滚、NRestarts=0；本轮无 429，实际限流来源 NOT_RUN；[部署报告](../architecture/GO-API-PR94-DEPLOY-REPORT-20261125.md)；#95 已于 2026-11-25 15:33:11（北京时间）升级至 `2f679dda`，合并提交 CI 重跑 PASS、三处哈希一致、NRestarts=0、无迁移/回滚，healthz/鉴权/C0/订单/单笔账单/OTA 双通道 PASS，物流候选上限/15分钟去重/终态排除与单笔官方状态比对 PASS（现场），真实 429 与失败保留现场 NOT_RUN（代码/测试 PASS）；[部署报告](../architecture/GO-API-PR95-DEPLOY-REPORT-20261125.md)；V0.5 签名测试版已发布为 Pre-release（`desktop-test-v0.5.0`），未设置 Latest、正式渠道仍为 `desktop-v0.1.1`；V0.4 已发布为 Pre-release，正式发布和真实官方写入仍需单独审批 |

## 三、已完成（限定范围内真实验收通过）

| ID | 功能及已完成的范围 | 关联 PR | 验收依据 / 不包含 |
| --- | --- | --- | --- |
| BASE-01 | Go + PostgreSQL 基础采集 API、去重、读回和 GitHub 按范围 CI | [#9](https://github.com/Fan6512/una-seller/pull/9)、[#10](https://github.com/Fan6512/una-seller/pull/10)、[#11](https://github.com/Fan6512/una-seller/pull/11)、[#51](https://github.com/Fan6512/una-seller/pull/51) | CI/PG 定点用例通过；不表示全部业务模块已上线 |
| AUTH-01 | Windows 邮箱登录、重开恢复与退出登录的基本闭环 | [#15](https://github.com/Fan6512/una-seller/pull/15) | 2026-11-17 单账号 Windows 实测；真实双账号、长期续签与 macOS 另行验收 |
| AUTH-02 | 一家 CBT 店铺 OAuth 连接、断开/重授权、真实站点映射 | [#23](https://github.com/Fan6512/una-seller/pull/23) / [Issue #21](https://github.com/Fan6512/una-seller/issues/21) | 2026-11-19 一个真实账号与五条 verified 映射 PASS；不代表商品创建、物流/账单 API 已通过 |
| COL-01 | Amazon 单商品点击采集 → Go 入库 → Windows 采集箱读取的最小闭环 | [#11](https://github.com/Fan6512/una-seller/pull/11)、[#13](https://github.com/Fan6512/una-seller/pull/13)、[#19](https://github.com/Fan6512/una-seller/pull/19) / [Issue #6](https://github.com/Fan6512/una-seller/issues/6) | 真实商品曾成功采集并读回；高清图、重复图、部分成功和复杂变体的质量验收仍列在 COL-02 |
| CBT-01 | C0 官方发布模型合同、只读探针、生产部署与现有账号模式识别 | [#69](https://github.com/Fan6512/una-seller/pull/69)、[#73](https://github.com/Fan6512/una-seller/pull/73)、[#85](https://github.com/Fan6512/una-seller/pull/85)、[#87](https://github.com/Fan6512/una-seller/pull/87) / [#50](https://github.com/Fan6512/una-seller/issues/50) | 2026-11-24 单账号真实只读 C0 PASS，standard_up；五站点证据完整，最新 14:39:47 北京时间。[PR #87 生产升级与真实 C0 验收摘要](../architecture/GO-API-PR87-DEPLOY-REPORT-20261124.md)。V0.4 实测已授权店铺、五站点、四标签及手动只读 C0 4/4 PASS；不包含原始枚举、后台独立确认、真实上传/创建权限及自动授权探测实机验收 |

## 四、进行中（代码已有但业务未验收、Draft PR 未合并，或负责人已批准启动下一阶段且实施 PR 待建）

### A. 采集、产品编辑与中文阅读

| ID | 功能 | 当前状态 | 关联 PR / Issue | 下一步或完成判据 |
| --- | --- | --- | --- | --- |
| COL-02 | 采集图片高清化、去重、视频过滤、部分成功原因与变体完整性 | 代码已有；#91 插件采集按钮布局修复已合并，负责人反馈正常；完整新版实测未完成 | [#19](https://github.com/Fan6512/una-seller/pull/19)、[#20](https://github.com/Fan6512/una-seller/pull/20)、[#91](https://github.com/Fan6512/una-seller/pull/91) / [#6](https://github.com/Fan6512/una-seller/issues/6) | 选真实复杂商品，核实图片质量、规格归属、采集和切换速度 |
| WB-01 | 采集箱整理工作稿：编辑、保存、并发冲突与重启恢复 | 代码已合并；#92 工作稿兼容修订与自动化 QA PASS，生产读取与历史数据比对 PASS（现场报告）， #94 隐藏普通历史描述（保留历史数据与超长恢复入口），保存/GUI 待验收 | [#52](https://github.com/Fan6512/una-seller/pull/52)、[#59](https://github.com/Fan6512/una-seller/pull/59)、[#92](https://github.com/Fan6512/una-seller/pull/92)、[#94](https://github.com/Fan6512/una-seller/pull/94) | 同一真实来源保存、重启读回，不覆盖原始采集 |
| WB-02 | 独立产品箱、勾选转库、来源追溯和再次编辑 | 代码已合并；#92 保存/转库字段保护回归 PASS，生产读取与历史数据比对 PASS（现场报告），实际转库/Windows GUI 待验收 | [#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65)、[#92](https://github.com/Fan6512/una-seller/pull/92) / [#57](https://github.com/Fan6512/una-seller/issues/57) | 工作稿 → 产品箱 → 编辑 → 重启读回；不得重复建产品或丢字段 |
| WB-03 | 语言标签、图片墙、变体独立编辑、重量尺寸及商品/变体 CNY 成本、USD 目标刊登收益及副图批量复制 | V0.5 #92 已合并；CI 与 Windows QA PASS，生产读取与历史数据比对 PASS（现场报告），#94 逐变体就地图片选择及独立 SKU/成本/目标收益已合并，CI PASS；新版 GUI 待验收 | [#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65)、[#92](https://github.com/Fan6512/una-seller/pull/92)、[#94](https://github.com/Fan6512/una-seller/pull/94) | 用真实多变体产品核实编辑、保存、主图、图片顺序、主图保留/副图复制、SKU 对应、独立成本及目标收益；旧客户端未变重复 SKU/无名变体可安全重排，歧义返回 422；目标刊登收益不等于实际结算 |
| CAT-01 | 官方 CBT 类目预测、分类必填属性和技术规格读取 | 代码已合并；#92 规格 ID/必填提示回归 PASS，#94 根据类目动态提示 MODEL 必填/待核实状态，CI PASS；完整客户端实测待完成 | [#31](https://github.com/Fan6512/una-seller/pull/31)、[#68](https://github.com/Fan6512/una-seller/pull/68)、[#92](https://github.com/Fan6512/una-seller/pull/92)、[#94](https://github.com/Fan6512/una-seller/pull/94) | 选择两种官方类目，核对枚举、单位、必填与切换类目后的已保存字段；SKU-06/07 真实 Family 与实际刊登 payload NOT_RUN，继续作为刊登门禁 |
| TR-01 | 订单标题、类目、属性、枚举等中文阅读（Azure/百度择一） | 代码已合并；真实翻译尚未验收 | [#67](https://github.com/Fan6512/una-seller/pull/67)、[#70](https://github.com/Fan6512/una-seller/pull/70) | 服务端配置合规凭据，测 EN/ES/PT、回退及原文/官方 ID 保留 |
| TR-02 | CBT 全量分类中文离线词典、增量更新与人工覆盖 | 代码已合并；中文 0/10,954 | [#72](https://github.com/Fan6512/una-seller/pull/72) | 先五批真实试译、核对质量与额度，随后断点续跑并记录覆盖率；不因没有密钥假报 PASS |
| AI-01 | 根据产品事实生成 EN/MX/BR 三语 SEO 标题 | 代码已合并；真实 AI 调用待验收 | [#71](https://github.com/Fan6512/una-seller/pull/71) | 使用真实保存的商品生成、分别采纳；MX/BR 不超过 60 字符且不得编造属性 |
| AI-02 | 英文卖点与详情独立采纳，按母版翻译 ES/PT，译文过期提醒 | 代码已合并；真实服务与 Windows 读回待验收 | [#71](https://github.com/Fan6512/una-seller/pull/71) | 从已保存英文描述翻译；人工核对数字/单位，手改后检测过期，重启读回 |
| DESK-01 | V0.3 编辑器、正式版/可选测试版更新通道 | 代码已合并，V0.4 Pre-release 已发布。公网 OTA 接口 PASS，V0.4 覆盖安装及检查更新等 5/5 PASS；跨版本 OTA 待验收；V0.5 #92 隔离 QA 构建 PASS，生产升级后双通道 OTA 接口复验 PASS（现场报告），V0.5 签名测试版已发布，实机跨版本 OTA 待验收 | [#29](https://github.com/Fan6512/una-seller/pull/29)、[#65](https://github.com/Fan6512/una-seller/pull/65)、[#92](https://github.com/Fan6512/una-seller/pull/92) / [#28](https://github.com/Fan6512/una-seller/issues/28) | 公网 `/api/v1/desktop/updates` 带 `X-Una-Include-Prerelease: 1` 返回 200（`0.4.0`），默认通道返回 204，两种响应均带 `Cache-Control: no-store`。现有 V0.3 无正式 Updater，已手动覆盖 V0.4；V0.5 测试版发布后默认通道 204、测试通道返回 0.5.0，两者均 `no-store`，清单安装包/签名公网下载与 SHA-256 一致；实机安装重启及实际通道选择待验收。正式 Release 需单独授权 |

### B. 已刊登管理、订单与财务

| ID | 功能 | 当前状态 | 关联 PR / Issue | 下一步或完成判据 |
| --- | --- | --- | --- | --- |
| PUB-01 | 已刊登商品只读同步、站点/SKU 展示、基础本地编辑草稿 | 代码已合并；真实性能与完整功能验收不齐 | [#27](https://github.com/Fan6512/una-seller/pull/27)、[#30](https://github.com/Fan6512/una-seller/pull/30)、[#31](https://github.com/Fan6512/una-seller/pull/31)、[#34](https://github.com/Fan6512/una-seller/pull/34) / [#26](https://github.com/Fan6512/una-seller/issues/26) | 已授权账号真实同步、详情、描述、规格；本地编辑不等于官方写回 |
| PUB-02 | 已刊登同步偶发 504、首屏慢请求 | 诊断与优化代码已有；异常根因未最终验收 | [#35](https://github.com/Fan6512/una-seller/pull/35)、[#36](https://github.com/Fan6512/una-seller/pull/36)、[#37](https://github.com/Fan6512/una-seller/pull/37) / [#38](https://github.com/Fan6512/una-seller/issues/38)、[#41](https://github.com/Fan6512/una-seller/issues/41) | 实际复现时取得 Go/代理/客户端同次脱敏分段日志再定点修复 |
| CBT-02 | **P0 · C1 首次真实刊登**：标准 UP + Net Proceeds + Remote；产品箱选择店铺及单站点，先一件无变体商品创建、官方 ID 回读与防重 | 已批准进入开发；实施 PR 待创建，尚无可用写入闭环 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50)；C0 [#85](https://github.com/Fan6512/una-seller/pull/85)、[#87](https://github.com/Fan6512/una-seller/pull/87) | 联合 CBT-03/IMG-01/PUB-03 的最小切片；开发/模拟先行。真实写入前核实原始业务枚举、后台模式及逐站点能力，并获独立批准；成功必须匹配官方 ID 与归属回读，不盲重试。多变体及多站点扩展后续实施 |
| ORD-01 | 订单只读同步、列表详情、跨站点/跨店铺搜索 | V0.4 部分通过；#90 修订已合并，生产抽测 USD 销售额、订单关联、履约筛选 PASS（现场报告）；V0.5 新订单主图按负责人意见暂通过（实际新增订单案例 NOT_RUN）；其余 GUI 按实测登记 | [#40](https://github.com/Fan6512/una-seller/pull/40)、[#43](https://github.com/Fan6512/una-seller/pull/43)、[#55](https://github.com/Fan6512/una-seller/pull/55)、[#90](https://github.com/Fan6512/una-seller/pull/90) | V0.4 同步/分页/筛选与跨店跨站搜索 PASS。#90 FIX-01/02/04：明确 USD 商品行汇总、Pack 销售编号与内部 Order ID、按需自动补图；重复疑似行待核实。生产只读核查 73 单/73 行，0 重复候选，负责人报告 73/73 Billing 美元金额一致；无回填。V0.5 实测金额、编号、补图、筛选及详情 |
| ORD-02 | 1688 采购订单号、成本、国内运单**手工关联** | V0.4 官方/采购物流分开展示 PASS；采购录入未验收 | [#46](https://github.com/Fan6512/una-seller/pull/46)、[#54](https://github.com/Fan6512/una-seller/pull/54) | 采购订单号、成本及国内运单手工保存仍待验收，确认跨店不串单；**未自动读取 1688** |
| ORD-03 | 官方发货 SLA 与单笔订单状态刷新 | V0.4 单笔订单状态刷新 PASS；#90 SLA/履约筛选修订已合并，生产已有 SLA 读取及履约筛选 PASS（现场报告），实时刷新/GUI 待验收 | [#48](https://github.com/Fan6512/una-seller/pull/48)、[#90](https://github.com/Fan6512/una-seller/pull/90) | FIX-07/08 已合并：绝对截止日期及精确倒计时，综合履约状态与服务端筛选；实时物流与 SLA 刷新会写入生产缓存，本轮 NOT_RUN；V0.5 核对实际 SLA、退款状态、单笔刷新后筛选列表一致性 |
| ORD-04 | 官方 Shipment 发货/运输/签收及运单追踪 | **生产后端 PASS；V0.4 Windows 9 PASS / 2 FAIL / 1 待验收；#90 生产已升级、已有详情只读验证 PASS（现场报告）** | [#66](https://github.com/Fan6512/una-seller/pull/66)、[#90](https://github.com/Fan6512/una-seller/pull/90)、[#94](https://github.com/Fan6512/una-seller/pull/94)、[#95](https://github.com/Fan6512/una-seller/pull/95) | 2026-11-24 生产 0010、API 823e6056、9 单真实物流与完整缓存 PASS；[脱敏验收摘要](../architecture/ORD-04-DEPLOY-REPORT-20261124.md)。V0.4 已发布为 Pre-release（[下载](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.4.0)），物流查询链接打不开、签收时间显示与后台相差 11 小时（旧版问题）；V0.5 FIX-05/06 已合并：受限网页弹窗（外部浏览器备用），时间统一北京时间；FIX-08：签收等综合履约状态。V0.5 生产已有物流详情/SLA 只读验证 PASS；实时刷新会写生产缓存，本轮 NOT_RUN；V0.5 负责人反馈实际打开浏览器新标签且查询站点不可用，物流窗口部分通过；#94 新增 Track718/17TRACK、复制运单与明确独立窗口/浏览器选择，CI PASS；第三方网站独立窗口实机兼容性、北京时间、状态及单笔刷新待复验。列表状态、运单复制、承运商、发货时间、时间轴、批量更新与缓存保留 PASS；#95 针对实机发现订单交易已同步但物流仍为旧缓存：单笔联动、每次显式订单同步后最多10笔陈旧非终态物流（15分钟间隔）且不读取历史，候选 CI [45558640026](https://github.com/Fan6512/una-seller/actions/runs/45558640026) PASS、已合并；2026-11-25 15:33:11（北京时间）生产升级至 `2f679dda` 后**生产后端 PASS**：单笔订单刷新联动官方物流（缺失快照 → `delivered` 落库）、候选列表 10 笔上限、`has_more`、15分钟去重与终态排除现场 PASS，失败保留与 429 冷却由定点测试 PASS（现场 NOT_RUN）；Windows GUI 联动提示、第三方物流窗口及时间显示待人工复验，正式 Release 待审批；2026-11-25 负责人实机反馈：一笔待发货订单恢复为正确已发货，状态更新结果 PASS，具体触发按钮未辨明；第三方网页仍打开默认地址、列表状态重复，需修复主查询入口（Track718/17TRACK 指定网页）、去掉重复状态，保留详情原始状态；不接第三方 API；[#96](https://github.com/Fan6512/una-seller/pull/96) 候选 CI 45571662671 Changed Files/桌面 PASS，本地代码复审无阻塞，负责人批准后已 Squash Merge 至 `415c782a`，Windows GUI 待验收 |
| FIN-01 | 官方 CBT 账单 A1 只读预览、费用/冲正明细 | **#49 已合并；生产账单读取抽测 PASS（现场报告）；V0.5 实机报告账单限流；#94 冷却共享、Retry-After CORS 与脱敏来源诊断已合并，生产升级及单笔账单读取 PASS（现场报告），真实429来源/冷却显示 NOT_RUN；net_income_usd=null；真实净收益与独立费用对账待验收** | [#49](https://github.com/Fan6512/una-seller/pull/49)、[#94](https://github.com/Fan6512/una-seller/pull/94) | 真实只读取证：负责人报告 73/73 账单返回；缺失/越权 200 空结果为 not_available，非 200 不猜测。仅展示已核实 USD 费用，采购 CNY；货件费用不跨订单累加，包装 429 冷却，保留原币数据。官方净收益、实际到账、退款及共享货件分摊未完成（net_income_usd=null）；A2 利润未接线。真实 Seller Center 回款核对及 Windows 验收待执行；2026-11-25 负责人实机单笔官方账单读取 PASS；账单未持久化、官方总计仍待核账。P1 补齐按授权店铺/站点/订单隔离的持久化快照、默认读缓存、显式刷新与失败保留，核实官方预计净回款字段；不得将货件净成本重复扣减或把局部小计冒充净回款；[#97](https://github.com/Fan6512/una-seller/pull/97) A2 已提交迁移 0013、单笔快照存取及桌面读取，HEAD `f16bcf3` 候选 CI [45587481720](https://github.com/Fan6512/una-seller/actions/runs/45587481720) 四项 PASS，PG 重启 seed/restart/verify 均实际 PASS；据本地验收报告，隔离 PG17.12 重复迁移、重启读回、授权/货件隔离、并发、缺失/失败保留及 429 冷却 PASS；#96 已合并，#97 已以普通 merge commit `a2570f61` 纳入 main `415c782a`，保留历史；相对旧 HEAD 仅合入两处桌面文件及台账，后端/迁移/测试工作流不变；新 HEAD `a2570f61` CI [45588642857](https://github.com/Fan6512/una-seller/actions/runs/45588642857) 四项 PASS，PG 写入/重启/读回三步实际 PASS，负责人批准后已 Squash Merge 至 `0064a239`，Windows 待验收；2026-11-25 生产迁移 0013 与后端 `0064a239` 已部署，单笔真实订单 `2000019508270206` 完成官方读取、PG 持久化、billing-cache 读回及服务重启后读回 PASS，重复缓存读取未触发官方请求；真实净回款 NOT_RUN（见 V0.6 部署报告）；2026-11-25 负责人确认账单已能持久化，实机该项 PASS，其余专项暂缓，真实净回款继续待 FIN-02 核账 |

## 五、待完成（尚无可用业务闭环）

| ID | 功能 / 最小交付目标 | 前置条件 | 参考 PR / Issue | 优先级 |
| --- | --- | --- | --- | --- |
| CBT-03 | 站点售价/币种、库存、销售条件、店铺模板，核对 CBT/UP/Family/变体/SKU 映射 | 当前账号可写模型/官方字段已确认 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| IMG-01 | 合法来源图片原件/选用版本可持续保存；上传 Mercado Libre 并取得真实 picture_id | 明确图片使用权、上传 API 和账号权限 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| PUB-03 | 刊登记录、逐站点错误定位、超时先核实、成功目标防重、失败目标安全重试 | CBT-02 发布合同落地 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0**，与 CBT-02 同一个闭环，不另搭队列平台 |
| PUB-04 | 修改已刊登价格、库存、图片、描述并读回；避免覆盖 Seller Center 人工修改 | 先有 CBT-02 稳定映射及真实可写字段 | [Delivery](../architecture/DELIVERY.md) M2 | P1 |
| IMG-02 | 商品图片去背景、白底处理、原图/候选对比及人工选用 | 先确定可用的合法修图方案；不预设官方有 AI 白底接口 | [产品定义](PRODUCT-WORKFLOW-AND-EDITOR.md) M1-C | P2 |
| AI-03 | 同一目标商品参考 3–5 条来源综合提取事实；来源记录仍独立 | AI-01/02 真实调用先验收 | [Issue #60](https://github.com/Fan6512/una-seller/issues/60) | P2 |
| FIN-02 | 区分官方预计净回款、实际结算回款及实际 USD 利润，按已核实费用/退款/汇率/采购成本计算，避免重复扣费 | FIN-01 真实账单对账、可靠汇率和采购成本 | [#49](https://github.com/Fan6512/una-seller/pull/49)（A1 已合并；A2 仅纯算式、未接线） | P1 |
| ORD-05 | 1688 授权与采购订单自动同步、厂家国内物流追踪 | 合法可靠的 1688 授权/接口验证 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | P2，主线后置 |
| ORD-06 | 货代仓库签收/转运 → Mercado Libre 官方扫描 → 买家签收的完整双物流关联 | ORD-04 真实平台 Shipment + 可用货代数据源 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | P2 |
| ORD-07 | 第三方物流 API 适配（Track718 / 17TRACK），辅助查询与轨迹缓存 | 暂缓；现阶段只保留网页入口，优先官方 Shipment；不申请 Key、不开发适配器/表，不覆盖官方状态，不纳入 V0.7 | 待独立立项 | P2（暂缓） |
| CS-01 | V0.8 售前问答客户端：问题列表、商品关联、原文/翻译、待回答筛选、人工确认并回复 | WH-04 只读数据与真实官方权限验收；发送属官方写入，须独立授权 | [Delivery](../architecture/DELIVERY.md) M3；实施 PR 待建 | V0.8 P1 |
| OPS-01 | V0.7 通知基础设施（WH-01/02）：统一 Webhook 接收、事件队列、Go Worker、去重/资源合并、可靠持久化与快速确认、限流、重试及重启恢复；六类主题按 WH-03/04/05 实施 | 待开发；沿用 `https://mkd.186512.xyz/api/v1/mercadolibre/notifications`，校验应用、资源、授权归属和请求边界，官方响应目标500ms内；不凭通知直接改业务状态，不自动官方写入；签名/IP等校验以真实官方合同为准 | [Delivery](../architecture/DELIVERY.md) M3；实施 PR 待建 | **V0.7 P0（该阶段基础设施；不替代首次 CBT 刊登 P0）** |
| WH-03 | V0.7 订单与物流通知处理器：新订单入库、Pack 关联、交易与货件状态更新，支持先收到物流通知、后补齐订单的情况 | OPS-01；复用 ORD-01/03/04 的官方解析、授权、更新时间保护与限流；订单/物流真实通知及乱序/重复回放验收 | 实施 PR 待建 | V0.7 P0 |
| WH-04 | V0.7 售前问答及售后消息只读后端：问题/商品、会话/消息原文、时间、官方待回答/未读状态，按店铺站点查询 | OPS-01；官方资源路径、卖家类型与读取权限逐类验证；不发送回复、不自动标记官方消息已读 | 实施 PR 待建 | V0.7 P1 |
| WH-05 | V0.7 投诉纠纷及站点商品通知只读后端：纠纷状态、类型、处理期限与允许操作；商品价格、库存、审核、上下架状态回读 | OPS-01；真实权限与资源归属验证；复用已刊登商品数据，不自动退款、处理纠纷或写回商品 | 实施 PR 待建 | V0.7 P1 |
| WH-06 | V0.7 通知真实联调与补偿：六类主题逐步启用，遗漏事件查询/回放、故障恢复及处理日志 | OPS-01/WH-03/04/05；先部署并确认公网回调响应、持久化、重试与安全，再经授权修改官方订阅；missed_feeds 实际 CBT 路径/权限待核实 | 实施 PR 待建 | V0.7 P0 |
| OPS-02 | V0.7 运营通知中心：默认运营首页、全局铃铛、店铺/站点筛选、待办统计与最新动态；订单/发货/SLA、问答/消息、纠纷、商品/物流/授权异常，紧急事项提醒 | 待开发；真实统计与只读详情接口，通知已读与业务已处理分离，跳转对应模块；完整业务操作 UI 后置；尚无数据的状态明确不可用，不显示虚假零值 | 实施 PR 待建 | V0.7 P1 |
| OPS-03 | V0.7 登录后店铺检查、Webhook 优先及低频订单/物流补偿、客户端缓存读取 | 待开发；初始建议订单30分钟、物流60分钟、客户端读自身 API 30秒；统一限流、顺序、手动优先、429退避；登录后增量检查且短期不重复；服务器持续处理 Webhook、低频补偿；任务不重叠，账单及完整物流历史手动；实际参数以验证为准 | 实施 PR 待建 | V0.7 P0 |
| DESK-04 | V0.7 冷启动重新登录及缓存隔离 | 待开发；保留邮箱/设置/图片缓存，ERP 会话不跨冷启动持久化；店铺授权与业务数据留在服务器，未登录不展示业务数据；完全退出后再启动需登录，最小化/切换界面无需重新登录；不保存密码，不引入离线业务库 | 实施 PR 待建 | V0.7 P1 |
| MOD-01 | V0.7 模块化边界登记与后续实施：稳定模块 ID、明确数据归属、公开内部接口/依赖、桌面导航按模块组织；内置模块与主程序一起构建 | 待开发；逐功能 PR 实施，不要求现有代码大重构，不引入动态加载/第三方 SDK/沙箱；本阶段不定义商业权益或价格 | 实施 PR 待建 | V0.7 P1 |
| CS-02 | V0.8 售后消息客户端：会话列表、双语阅读、快捷话术、人工回复 | WH-04 与真实消息权限；发送须独立批准 | 实施 PR 待建 | V0.8 P1 |
| CS-03 | V0.9 投诉纠纷管理客户端：纠纷详情、处理期限、官方允许操作、人工确认 | WH-05；真实纠纷操作权限及官方写入审批 | 实施 PR 待建 | V0.9 P1 |
| PUB-05 | 多变体 Family 与多店多站点批量刊登 | CBT-02/03、IMG-01、PUB-03 的真实单站点刊登闭环；独立任务范围、防重与逐站点结果 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | 后续 P1 |
| PUB-06 | V0.9 商品运营与异常管理：审核/下架、价格库存变化与监控 | WH-05、PUB-01/04；先只读异常和人工处理，不自动调价 | 实施 PR 待建 | V0.9 P1 |
| OPS-04 | 后续高级运营分析与自动化候选（商品监控、调价策略、运营报表） | 真实业务数据与人工流程稳定；自动官方写入另行批准 | 未立项 | P2，候选 |
| IMG-03 | 后续 AI 商品图与广告视频生成候选（IMG-02 白底先行） | 合法素材、真实服务能力、人工采用；是否复用 UnaAPI 后续决定 | 未立项 | P2，候选 |
| FIN-03 | 后续高级财务统计及利润分析候选 | FIN-01/02 真正完成净回款/退款/汇率/成本对账，不另建重复账本 | 未立项 | P2，候选 |
| CS-04 | 后续 AI 客服辅助候选：翻译、回复候选与人工采纳；自动回复不纳入当前范围 | CS-01/02、AI/TR 真实服务及隐私权限验证 | 未立项 | P2，候选 |
| MOD-02 | 后续内置应用目录/模块启用管理候选 | MOD-01 模块边界稳定后另行设计；不预设付费、权益、订阅、支付或第三方动态插件 | 未立项 | P2，候选 |
| DESK-02 | macOS arm64/x64 原生实机登录、业务 UI 和更新验收 | macOS 测试环境与已批准构建 | [CROSS_PLATFORM](../architecture/CROSS_PLATFORM.md) | P1（正式双端交付前） |
| DESK-03 | **桌面缩略图与图片加载优化·第一阶段**：内部先做（原建议 P0）已知 CDN 小图预览/点开原图及变体图片按需渲染；随后（原建议 P1）主图优先、其他懒加载、相邻最多 1–2 张预加载、失败回退重试与脱敏耗时诊断；**P2 本地有界缩略图缓存仅作为第二阶段候选，不在本期开发**；始终保留高清原始 URL 供编辑、修图与 CBT 刊登 | 独立桌面 Draft PR；白名单 URL 单测、失败/重试、性能对比和 Windows 真机验收；**不增加 Go 图片代理/对象存储/数据库迁移** | [Issue #83](https://github.com/Fan6512/una-seller/issues/83)（已立项；实现 PR 待创建）；数据库读写优化 [#79](https://github.com/Fan6512/una-seller/pull/79)、[#82](https://github.com/Fan6512/una-seller/pull/82) 属于不同范围 | **P1（全局）**（待开发；行内 P0/P1/P2 仅表示本功能内实施顺序，不阻断 CBT 首次刊登） |
| DOC-01 | 将旧 README/部分早期阶段说明与最新实际状态对齐 | 以本台账与最终实测证据为准 | [README](../../README.md)、[docs/README](../README.md) | P2，不阻断开发 |

### 5.1 模块边界登记（负责人确认方向，具体技术合同由后续 PR 审核）

仅按职责划分，**不划分免费或付费，不设计收费、权益授权或支付系统**。保持 Go + PostgreSQL、Tauri 单体部署与统一签名更新。新增功能优先归属现有模块，不因模块化拆微服务或重写已有业务。下面是职责归属，不代表已经完成目录重构或已冻结 API/表结构。

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
| 模块组织与应用入口 | MOD-01/02 | 内置模块 ID、依赖和导航/目录；不开发动态安装、收费与第三方 SDK |

## 六、下一轮执行顺序（先完整闭环，不并行扩大战线）

1. **P0 首次真实 CBT 刊登**：C0 已完成，不重复取证。先收敛 CBT-03/IMG-01 的单商品单站点字段与图片，再以 CBT-02 + PUB-03 完成标准 UP / net_proceeds / remote 的最小闭环。开发和模拟测试先行；真实图片上传及创建、迁移、生产部署需独立批准。真实发布前核实原始 business_model、卖家后台模式及目标站点权限。
2. **V0.6 修复收口**：#94/#95 及 #97 生产升级、0013 迁移、真实账单持久化 PASS；负责人确认账单实机 PASS，其他 Windows 专项暂缓、不阻塞签名测试版（未逐项实测不标 PASS）。[#98](https://github.com/Fan6512/una-seller/pull/98) 版本号 0.6.0 Draft 已创建，待 CI/合并批准、`desktop-test-v0.6.0` 签名发布审批及 OTA 验收。新订单主图历史暂通过、真实新增订单案例 NOT_RUN；不把 CI/QA 构建冒充 GUI PASS。
3. **配置真实翻译 + AI 测试服务**：用一个已保存商品验收 TR-01、AI-01、AI-02；TR-02 五批试译和离线词典推进与 C1 解耦，不阻断首次刊登。凭据只放受控环境。
4. **首次刊登完成后再扩展**：多变体 Family、多店多站点批量与 PUB-04；ORD-04 Windows 剩余验收及 FIN-01/02 真实回款对账、分摊与利润核算（#49 A1 已合并）；1688 自动化及客服完整操作 UI 后置。
5. **V0.7 自动化运营**：V0.6 收口后，按 OPS-01（WH-01/02）→ WH-03 → WH-04/05 → WH-06 → OPS-03/DESK-04 → OPS-02 推进，MOD-01 随各功能 PR 明确边界、不单独开展大重构；Go + PostgreSQL，避免额外中间件。V0.8 客服 UI（CS-01/02），V0.9 纠纷与商品运营（CS-03/PUB-06），V1.0 核心业务稳定及正式版。MOD-02 与高级运营/AI/财务候选不绑定版本、不预设收费；根据真实需求另行立项。具体架构、安全、数据模型及 API 合同由独立功能 PR 审核，不以本台账替代。CBT 首次刊登、财务核算保留独立主线。

## 七、新需求登记、完成记录范例

假设以后提出“产品箱新增某项操作”，先新增一行：

| ID | 功能 | 当前状态 | 关联 PR | 完成判据 |
| --- | --- | --- | --- | --- |
| WB-04（示例，尚未登记实际需求） | 用户描述的最小新功能 | 待完成 | — | Windows 实际点击操作、保存并读回成功 |

开始开发：在该行附 Draft PR 链接，状态改“进行中”；PR 合并：记录 PR 编号及 main 提交，**若未实测仍保持“进行中”**；实际验收 PASS：整行移入“已完成”，写日期/范围。新需求中明确的延期项移入“暂缓”，不要删除历史证据。

**每次更新只改相关行和下方变更记录**；不因一项小需求重写架构计划，也不把各 PR 的所有技术细节复制到本文件。

## 八、变更记录

| 日期 | 变更 | 依据 |
| --- | --- | --- |
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

**当前边界**：#85、#87 单账号真实 C0 已验收，C1 开发已启动、实施 PR 待建；图片上传、官方创建、逐站点发布记录、售价模板和变体真实发布均**未完成**。此图是目标流程，**不能据此发起真实 POST**。

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

    F -.-> P["V0.5 签名发布、#94/#95 生产升级 PASS；V0.6 客户端发布待执行"]
    H -.-> Q["V0.5 签名测试版发布后由负责人验收 GUI/安装/OTA；macOS 待验收"]
    O --> R["负责人真机功能验收"]

    classDef verify fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class F,G,H,L,M,O verify
~~~

**更新纪律**：绿色 CI、Windows 构建、手工 GUI、服务端部署、签名 OTA 属于不同验收项。测试版开关代码已合并，不代表当前正式签名更新渠道已经实测通过。

## 十、流程图维护方式

- 新增需求若改变业务顺序：优先修改对应章节的 Mermaid 节点与箭头，不重新制作图片；同一流程可以保留稳定功能 ID。
- 不把目标流程中的橙色/待办节点当作“已完成”；对应功能 PR 合并且真实验收通过，再更新该流程下的状态说明与上方台账。
- 新 PR 只影响某个模块时，仅维护该模块的台账行、关联 PR、流程图及下方变更记录，无须大规模重写其他图。

