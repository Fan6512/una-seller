# UNA Seller · 功能状态总台账

> **日常功能进度与需求登记的统一入口** · 首次盘点：2026-11-23  
> 核对基线：main = e9b29c8f6f33acbc565930afd83db461e97d8cc4（已含 #70、#72、#71、#69、#73）。  
> 产品范围：自用的 Mercado Libre CBT ERP；核心链路 **采集 → 工作稿/产品箱 → AI 整理 → CBT 刊登 → 已刊登管理 → 订单/财务**。Amazon 等平台只作为采集来源。  
> 本文件记录功能进度与验收，不替代 [产品定义](PRODUCT-WORKFLOW-AND-EDITOR.md)、[技术路线](../architecture/DELIVERY.md)、具体 PR/Issue 或官方 API 取证。发现记录与实际代码冲突时，先核实当前 main、PR 和真实验收证据，再更新本台账。

## 一、维护约定（保持简单）

- **固定 ID**：一个独立的用户功能一个 ID；新需求新增行，已有需求只更新原行，不反复生成新的任务书或重复 Issue。被延期或取消的条目仍保留原因与历史 PR。
- **已完成**：约定的最小功能范围已并入 main，且该范围的真实业务/操作验收有明确 PASS；不能把“CI 通过”当“业务已完成”。已完成的范围很小也要如实限定。
- **进行中**：有 Draft PR 在开发，**或者**代码已合并但真实 API、凭据、数据、Windows GUI/部署仍未验收；两种情况在备注中写清。
- **待完成**：已确认需求但尚无完成该业务目标的代码；早期只读能力、设计合同或 Mock 不等于正式实现。
- **暂缓/不做**：当前优先级之外，不主动开 PR；改变优先级须由负责人确认。
- **每条关联**：功能 ID、状态、关联 PR（合并 PR 或 Draft PR）、验收/剩余事项。PR 合并时补编号和 main 状态；真实验收通过后再移至“已完成”。涉及发布、迁移、生产变更始终单独授权。
- **新增需求时**：先在对应模块新增一行，写清最小可验收结果；确定开发时关联 PR；PR 合并和实测后更新状态。文档变更走正常 GitHub PR，**不在对话结束后自动后台修改仓库**。
- **优先级**：P0 = 首次真实 CBT 刊登必需；P1 = 自用稳定性与订单经营必需；P2 = 增强/非主线；暂缓 = 不抢当前资源。优先级是工作顺序，不是开发承诺。

## 二、截至 2026-11-23 的关键事实

| 指标 | 经核对的情况 |
| --- | --- |
| main | **e9b29c8f**，#73 已 Squash Merge；本次主干相关 Go/Changed Files 检查成功，桌面与 MV3 按范围跳过 |
| 仍开放的代码 PR | **[#66](https://github.com/Fan6512/una-seller/pull/66)** 官方 Shipment（Draft，含尚未合并的 0010 迁移）；**[#49](https://github.com/Fan6512/una-seller/pull/49)** 订单账单预览（Draft） |
| 分类官方原文 | 11,668；停用 714；有效待译 10,954；隔离目录已初始化 |
| 分类中文翻译 | **0/10,954**；真实翻译提供方凭据未配置，五批试译未执行 |
| AI 真实服务 | 三语标题/详情代码已合并；真实 Key、已保存测试商品与 Windows 操作验收尚缺 |
| CBT 刊登 | C0 只读合同及探针已合并；C1 真实商品创建、官方图片上传、发布记录未实现 |
| CBT 测试账号 | 单个 CBT 主账号的 OAuth 与五条站点映射曾实测成功；**不能替代** C0 对当前账号实际 business_model、可写模型/字段的只读取证 |
| 桌面端 | Windows 已有历史真实登录/采集测试；最新 main 的 V0.3 + AI/词典整体 Windows QA 未完成；macOS 实机未验收 |
| 部署 | GitHub 合并不代表云端 Go API、数据库、客户端及签名发行同步完成；生产迁移和真实刊登需单独审批 |

## 三、已完成（限定范围内真实验收通过）

| ID | 功能及已完成的范围 | 关联 PR | 验收依据 / 不包含 |
| --- | --- | --- | --- |
| BASE-01 | Go + PostgreSQL 基础采集 API、去重、读回和 GitHub 按范围 CI | [#9](https://github.com/Fan6512/una-seller/pull/9)、[#10](https://github.com/Fan6512/una-seller/pull/10)、[#11](https://github.com/Fan6512/una-seller/pull/11)、[#51](https://github.com/Fan6512/una-seller/pull/51) | CI/PG 定点用例通过；不表示全部业务模块已上线 |
| AUTH-01 | Windows 邮箱登录、重开恢复与退出登录的基本闭环 | [#15](https://github.com/Fan6512/una-seller/pull/15) | 2026-11-17 单账号 Windows 实测；真实双账号、长期续签与 macOS 另行验收 |
| AUTH-02 | 一家 CBT 店铺 OAuth 连接、断开/重授权、真实站点映射 | [#23](https://github.com/Fan6512/una-seller/pull/23) / [Issue #21](https://github.com/Fan6512/una-seller/issues/21) | 2026-11-19 一个真实账号与五条 verified 映射 PASS；不代表商品创建、物流/账单 API 已通过 |
| COL-01 | Amazon 单商品点击采集 → Go 入库 → Windows 采集箱读取的最小闭环 | [#11](https://github.com/Fan6512/una-seller/pull/11)、[#13](https://github.com/Fan6512/una-seller/pull/13)、[#19](https://github.com/Fan6512/una-seller/pull/19) / [Issue #6](https://github.com/Fan6512/una-seller/issues/6) | 真实商品曾成功采集并读回；高清图、重复图、部分成功和复杂变体的质量验收仍列在 COL-02 |

## 四、进行中（代码已有但业务未验收，或 Draft PR 未合并）

### A. 采集、产品编辑与中文阅读

| ID | 功能 | 当前状态 | 关联 PR / Issue | 下一步或完成判据 |
| --- | --- | --- | --- | --- |
| COL-02 | 采集图片高清化、去重、视频过滤、部分成功原因与变体完整性 | 代码已有；新版实测未完成 | [#19](https://github.com/Fan6512/una-seller/pull/19)、[#20](https://github.com/Fan6512/una-seller/pull/20) / [#6](https://github.com/Fan6512/una-seller/issues/6) | 选真实复杂商品，核实图片质量、规格归属、采集和切换速度 |
| WB-01 | 采集箱整理工作稿：编辑、保存、并发冲突与重启恢复 | 核心实机流程 OWNER PASS（2026-11-23）；边界用例仍保留 CI 证据 | [#52](https://github.com/Fan6512/una-seller/pull/52)、[#59](https://github.com/Fan6512/una-seller/pull/59) | 负责人已确认真实采集→整理→保存→重启读回；不重复安装复测。保留原始快照与 CAS 的定向回归 |
| WB-02 | 独立产品箱、勾选转库、来源追溯和再次编辑 | 转库→编辑→重启读回 OWNER PASS（2026-11-23）；批量/并发等边界仍以定向测试为准 | [#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65) / [#57](https://github.com/Fan6512/una-seller/issues/57) | 核心工作流已由负责人实测；仅产品箱发布前字段扩充与实际问题做定点回归，不重新执行整套验收 |
| WB-03 | 语言标签、图片墙、变体 SKU 表、重量尺寸及采购成本编辑 | 代码已合并；最终 UI 验收待完成 | [#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65) | 用真实多变体产品核实编辑、保存、主图、图片顺序和 SKU 对应 |
| CAT-01 | 官方 CBT 类目预测、分类必填属性和技术规格读取 | 代码已合并；真实类目/客户端回归待完成 | [#31](https://github.com/Fan6512/una-seller/pull/31)、[#68](https://github.com/Fan6512/una-seller/pull/68) | 选择两种官方类目，核对枚举、单位、必填与切换类目后的已保存字段 |
| TR-01 | 订单标题、类目、属性、枚举等中文阅读（Azure/百度择一） | 代码已合并；真实翻译尚未验收 | [#67](https://github.com/Fan6512/una-seller/pull/67)、[#70](https://github.com/Fan6512/una-seller/pull/70) | 服务端配置合规凭据，测 EN/ES/PT、回退及原文/官方 ID 保留 |
| TR-02 | CBT 全量分类中文离线词典、增量更新与人工覆盖 | 代码已合并；中文 0/10,954 | [#72](https://github.com/Fan6512/una-seller/pull/72) | 先五批真实试译、核对质量与额度，随后断点续跑并记录覆盖率；不因没有密钥假报 PASS |
| AI-01 | 根据产品事实生成 EN/MX/BR 三语 SEO 标题 | 代码已合并；真实 AI 调用待验收 | [#71](https://github.com/Fan6512/una-seller/pull/71) | 使用真实保存的商品生成、分别采纳；MX/BR 不超过 60 字符且不得编造属性 |
| AI-02 | 英文卖点与详情独立采纳，按母版翻译 ES/PT，译文过期提醒 | 代码已合并；真实服务与 Windows 读回待验收 | [#71](https://github.com/Fan6512/una-seller/pull/71) | 从已保存英文描述翻译；人工核对数字/单位，手改后检测过期，重启读回 |
| DESK-01 | V0.3 编辑器、正式版/可选测试版更新通道 | 代码已合并；签名 OTA/新客户端真实验收待完成 | [#29](https://github.com/Fan6512/una-seller/pull/29)、[#65](https://github.com/Fan6512/una-seller/pull/65) / [#28](https://github.com/Fan6512/una-seller/issues/28) | 最新版 Windows QA、实际验签更新、测试版 ON/OFF；正式 Release 需单独授权 |

### B. 已刊登管理、订单与财务

| ID | 功能 | 当前状态 | 关联 PR / Issue | 下一步或完成判据 |
| --- | --- | --- | --- | --- |
| PUB-01 | 已刊登商品只读同步、站点/SKU 展示、基础本地编辑草稿 | 代码已合并；真实性能与完整功能验收不齐 | [#27](https://github.com/Fan6512/una-seller/pull/27)、[#30](https://github.com/Fan6512/una-seller/pull/30)、[#31](https://github.com/Fan6512/una-seller/pull/31)、[#34](https://github.com/Fan6512/una-seller/pull/34) / [#26](https://github.com/Fan6512/una-seller/issues/26) | 已授权账号真实同步、详情、描述、规格；本地编辑不等于官方写回 |
| PUB-02 | 已刊登同步偶发 504、首屏慢请求 | 诊断与优化代码已有；异常根因未最终验收 | [#35](https://github.com/Fan6512/una-seller/pull/35)、[#36](https://github.com/Fan6512/una-seller/pull/36)、[#37](https://github.com/Fan6512/una-seller/pull/37) / [#38](https://github.com/Fan6512/una-seller/issues/38)、[#41](https://github.com/Fan6512/una-seller/issues/41) | 实际复现时取得 Go/代理/客户端同次脱敏分段日志再定点修复 |
| CBT-01 | C0 官方发布模型合同及**只读探针** | 代码和合同已合并；真实 CBT 账号能力证据待补 | [#69](https://github.com/Fan6512/una-seller/pull/69)、[#73](https://github.com/Fan6512/una-seller/pull/73) | 获批账号执行三类 GET，确认 business_model、logistic_type、站点与发布模型；**不能据 C0 自动进入真实 POST** |
| ORD-01 | 订单只读同步、列表详情、跨站点/跨店铺搜索 | 代码已合并；当前真实 Windows/账号闭环待复核 | [#40](https://github.com/Fan6512/una-seller/pull/40)、[#43](https://github.com/Fan6512/una-seller/pull/43)、[#55](https://github.com/Fan6512/una-seller/pull/55) | 真实店铺同步、Order ID 精确搜索和店铺隔离；仅搜索已同步订单 |
| ORD-02 | 1688 采购订单号、成本、国内运单**手工关联** | 代码已合并；新版 GUI 与真实订单核对待完成 | [#46](https://github.com/Fan6512/una-seller/pull/46)、[#54](https://github.com/Fan6512/una-seller/pull/54) | 1688 采购记录与官方订单分开展示，确认跨店不串单；**未自动读取 1688** |
| ORD-03 | 官方发货 SLA 与单笔订单状态刷新 | 代码已合并；真实 Seller Center 对照仍不完整 | [#48](https://github.com/Fan6512/una-seller/pull/48) | 真实订单时限及退款状态按官方回读，不把 Shipment ID 当运单 |
| ORD-04 | 官方 Shipment 发货/运输/签收及运单追踪 | **Draft PR 开发中；未合并** | [#66](https://github.com/Fan6512/una-seller/pull/66) | 合流当前 main，真实 Shipment 只读回读、Windows 测试；0010 迁移和部署需审批 |
| FIN-01 | 官方 CBT 账单只读预览、费用/冲正/退款明细 | **Draft PR 开发中；未合并** | [#49](https://github.com/Fan6512/una-seller/pull/49) | 合流当前 main，真实账单与 Seller Center 对账、Windows 验收；未核实净回款标注“待核账” |

## 五、待完成（尚无可用业务闭环）

| ID | 功能 / 最小交付目标 | 前置条件 | 参考 PR / Issue | 优先级 |
| --- | --- | --- | --- | --- |
| ACC-01 | **设置 → 账号管理：**登录邮箱、修改密码与必要再验证；不增加独立账号系统 | 现有 Supabase Auth 登录和本人会话 | [Issue #75](https://github.com/Fan6512/una-seller/issues/75) | **P0**，集中 Windows QA 前交付 |
| CRED-01 | **用户云端 AI/翻译服务绑定：**保存/替换/删除/测试用户自己的 AI Key、Azure 或百度 Key，服务端加密、按认证 owner 动态选择服务 | 私有加密凭据存储、owner scoped API、受控 HTTPS、限流及密钥不回传；与 #66 的 0010 迁移协调 | [Issue #75](https://github.com/Fan6512/una-seller/issues/75) | **P0**，AI 真机验证前交付 |
| CBT-02 | **C1 真实刊登：**产品箱选择店铺/获准站点、官方模型创建、逐站点发布记录与安全回读重试 | CBT-01 实际账号取证；独立审批发布写入 | C0 [#69](https://github.com/Fan6512/una-seller/pull/69)、[#73](https://github.com/Fan6512/una-seller/pull/73)；[Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| CBT-03 | 站点售价/币种、库存、销售条件、店铺模板，核对 CBT/UP/Family/变体/SKU 映射 | 当前账号可写模型/官方字段已确认 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| IMG-01 | 合法来源图片原件/选用版本可持续保存；上传 Mercado Libre 并取得真实 picture_id | 明确图片使用权、上传 API 和账号权限 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0** |
| PUB-03 | 刊登记录、逐站点错误定位、超时先核实、成功目标防重、失败目标安全重试 | CBT-02 发布合同落地 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | **P0**，与 CBT-02 同一个闭环，不另搭队列平台 |
| PUB-04 | 修改已刊登价格、库存、图片、描述并读回；避免覆盖 Seller Center 人工修改 | 先有 CBT-02 稳定映射及真实可写字段 | [Delivery](../architecture/DELIVERY.md) M2 | P1 |
| IMG-02 | 商品图片去背景、白底处理、原图/候选对比及人工选用 | 先确定可用的合法修图方案；不预设官方有 AI 白底接口 | [产品定义](PRODUCT-WORKFLOW-AND-EDITOR.md) M1-C | P2 |
| AI-03 | 同一目标商品参考 3–5 条来源综合提取事实；来源记录仍独立 | AI-01/02 真实调用先验收 | [Issue #60](https://github.com/Fan6512/una-seller/issues/60) | P2 |
| FIN-02 | 按真实已确认净回款 + 费用/退款/汇率/采购成本计算 USD 利润 | FIN-01 真实账单对账、可靠汇率和采购成本 | [#49](https://github.com/Fan6512/una-seller/pull/49)（目前只有预览/算式） | P1 |
| ORD-05 | 1688 授权与采购订单自动同步、厂家国内物流追踪 | 合法可靠的 1688 授权/接口验证 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | P2，主线后置 |
| ORD-06 | 货代仓库签收/转运 → Mercado Libre 官方扫描 → 买家签收的完整双物流关联 | ORD-04 真实平台 Shipment + 可用货代数据源 | [Issue #50](https://github.com/Fan6512/una-seller/issues/50) | P2 |
| CS-01 | 售前 Questions、售后 Messages、退款/争议在 ERP 内查询和处理 | 有真实运营需求与官方权限 | [Delivery](../architecture/DELIVERY.md) M3 | P2 |
| OPS-01 | Webhook/通知、必要自动更新与异常待办 | 真实业务流程和官方 Topics/权限验证 | [Delivery](../architecture/DELIVERY.md) M3 | P2 |
| DESK-02 | macOS arm64/x64 原生实机登录、业务 UI 和更新验收 | macOS 测试环境与已批准构建 | [CROSS_PLATFORM](../architecture/CROSS_PLATFORM.md) | P1（正式双端交付前） |
| DOC-01 | 将旧 README/部分早期阶段说明与最新实际状态对齐 | 以本台账与最终实测证据为准 | [README](../../README.md)、[docs/README](../README.md) | P2，不阻断开发 |

## 六、下一轮执行顺序（2026-11-23 负责人最新决策：先实现，最后一次集中 Windows QA）

1. **不重测已经实机确认的核心工作台**：负责人已确认真实采集、整理工作稿、产品箱转库、编辑、重启读回。后续仅做变更范围的 Go/PG17、TS/Mock 定向回归；高清图/复杂变体等尚未报告 PASS 的细项继续追踪。
2. **先建设账号云端服务**：完成 ACC-01/CRED-01，设置中增加修改密码和用户自有 AI/中文翻译 Key；Go 服务器由启动级单一环境凭据调整为每个登录用户按请求取回加密 Key。服务端真实测试已保存产品的三语文案，失败不得覆盖草稿。 [Issue #75](https://github.com/Fan6512/una-seller/issues/75)。
3. **并行推进 C0 只读账号取证**：使用已合并探针，对获批账号确认 Standard UP/Legacy/Fully Managed、业务模式、站点和计价权限。未明确资格前不能进行真实商品 POST。
4. **补齐完整的刊登前置流程并打通 C1**：产品箱已经保存的多语言文案、官方分类/属性/变体/图片、站点报价与销售条件完整性核验；官方图片上传并记录 picture_id；仅在获批模型下完成单 SKU、单站点商品创建、结果回读与安全重试。各站点状态、错误原因及成功对象 ID 均要持久化。扩大到多站点/变体前先验证核心创建链路。
5. **最后统一 Windows 客户端构建及 GUI 综合验收**：已对齐的云端 Go API、数据库迁移（均需独立审批）和客户端功能一次性以同一候选版本安装验收，重点验收账号管理、真实 AI 调用、完整预检和首次刊登回读。不要每合并一个 PR 就要求人工重装 Windows。
6. **运营能力后置**：已刊登修改写回、#66 Shipment、#49 Billing、1688 自动化、客服等分别推进。离线 10,954 条词典翻译可先做小批次试译；不得成为单 SKU 刊登的阻塞项，也不得静默消耗某账号 Key。

**PR 规则不变：**开发使用小型 Draft PR，Go/PG17 和 Desktop 定向 CI 必须保持，实际云端有授权的小样本可独立探测；不在本轮反复触发完整 Windows 发行。生产迁移、真实商品创建、签名发行均需负责人单独批准。

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
| 2026-11-23 | 负责人补充核心采集→整理→转库→编辑→重启读回已实机验证；新增 ACC-01/CRED-01；调整为刊登前功能/后端全部对齐后再集中 Windows QA，真实 AI 与 C1 为下一轮关键验收 | [Issue #75](https://github.com/Fan6512/una-seller/issues/75)、[Issue #50](https://github.com/Fan6512/una-seller/issues/50) |
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

**当前边界**：COL-01 的基本采集曾实测；COL-02 的高清图片和完整变体、WB-01～03 的新版 Windows 保存/转库读回仍须验收。

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

    C -.-> T["#69 + #73：C0 已合并，但真实账号 GET 尚未验收"]
    L -.-> U["C1 尚未开发；真实 POST 须单独批准"]

    classDef planned fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class D,E,G,H,J,K,L,M,N,O,P,Q,R,S planned
~~~

**当前边界**：#69、#73 仅完成 C0 代码/合同；C1、图片上传、逐站点发布记录、售价模板和变体真实发布均**未完成**。此图是目标流程，**不能据此发起真实 POST**。

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

**当前边界**：#40/#43/#55 的订单读取和搜索、#46/#54 的**人工采购关联**已合并；#66 官方 Shipment 仍 Draft；1688 自动授权/采购物流及货代自动跟踪未形成完整链路。图中的厂家、货代、买家节点描述卖家真实履约流程，不表示系统已经自动感知全部状态。

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

    B -.-> O["#49：账单只读预览仍为 Draft"]
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
    E --> F["Windows GUI 真机验收"]
    F --> G{"功能与后端版本匹配、验收通过？"}
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

    F -.-> P["当前 V0.3 整体 Windows GUI 待验收"]
    H -.-> Q["真实签名 OTA / macOS 实机尚未验收"]

    classDef verify fill:#fff5e5,color:#7a4d00,stroke:#e3ae56
    class F,G,H,L,M,O verify
~~~

**更新纪律**：绿色 CI、Windows 构建、手工 GUI、服务端部署、签名 OTA 属于不同验收项。测试版开关代码已合并，不代表当前正式签名更新渠道已经实测通过。

## 十、流程图维护方式

- 新增需求若改变业务顺序：优先修改对应章节的 Mermaid 节点与箭头，不重新制作图片；同一流程可以保留稳定功能 ID。
- 不把目标流程中的橙色/待办节点当作“已完成”；对应功能 PR 合并且真实验收通过，再更新该流程下的状态说明与上方台账。
- 新 PR 只影响某个模块时，仅维护该模块的台账行、关联 PR、流程图及下方变更记录，无须大规模重写其他图。

