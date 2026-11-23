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
| WB-01 | 采集箱整理工作稿：编辑、保存、并发冲突与重启恢复 | 代码已合并；最新 GUI 待验收 | [#52](https://github.com/Fan6512/una-seller/pull/52)、[#59](https://github.com/Fan6512/una-seller/pull/59) | 同一真实来源保存、重启读回，不覆盖原始采集 |
| WB-02 | 独立产品箱、勾选转库、来源追溯和再次编辑 | 代码已合并；完整 Windows 验收待完成 | [#59](https://github.com/Fan6512/una-seller/pull/59)、[#65](https://github.com/Fan6512/una-seller/pull/65) / [#57](https://github.com/Fan6512/una-seller/issues/57) | 工作稿 → 产品箱 → 编辑 → 重启读回；不得重复建产品或丢字段 |
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

## 六、下一轮执行顺序（先完整闭环，不并行扩大战线）

1. **配置真实翻译 + AI 测试服务**：先 TR-02 五批中文试译，核对成功率/费用/停机恢复；同时用一个已保存商品测 TR-01、AI-01、AI-02。凭据只放服务端。
2. **同一 main 版本的 Windows QA**：验收 WB-01/02/03、CAT-01、TR/AI 和产品字段保存读回，真实问题只做定点修复。测试包不冒充正式发行。
3. **CBT C0 真实只读取证**：已有 OAuth 不代表明确 business_model。对获批账号运行现有探针，脱敏记录可写模型与站点能力。
4. **推进首次真实 CBT 刊登**：先收敛 CBT-03/IMG-01，再作为一个最小 C1 交付 CBT-02 + PUB-03；先受控测试一件、多站点结果分别回读。**任何 POST/迁移/正式发布都需独立批准。**
5. **商品可发布后再扩展运营**：已刊登维护 PUB-04；#66 Shipment 和 #49 Billing 的真实验证与合并；1688 自动化、客服等后置。

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
| 2026-11-23 | 初始化功能总台账；盘点 main、58 个已关闭 PR、[#66](https://github.com/Fan6512/una-seller/pull/66)、[#49](https://github.com/Fan6512/una-seller/pull/49) 与公开 Issue；将功能完成和真实验收分离 | main e9b29c8f；当前官方类目 11,668/714/10,954；截至本日已提供的实测报告 |
