# Una Seller 文档索引

> 当前阶段：**自用优先、最小工程、仅 Mercado Libre 运营；Windows/macOS 双平台桌面客户端；浏览器插件采集 Amazon 与 Mercado Libre 竞品。**
> 2026-11-15 更新产品三区工作流、轻量内置插件、AI 编辑、多站点刊登的正式规划；浏览器采集与采集箱 MVP 继续按阶段实施。架构用于减少返工，不要求一次装完所有组件。

## 现行架构规范

1. [Foundation v2.1：第一性原理与最简技术地基](architecture/FOUNDATION.md)
2. [Delivery v2.1：M0～M3 精简开发路线](architecture/DELIVERY.md)
3. [Domain v2.1：最小平台数据与写入契约](architecture/DOMAIN_AND_EXECUTION.md)
4. [CI 与运行环境 v1：GitHub Actions、开发/部署分工与验收边界](architecture/CI_AND_RUNTIME.md)
5. [Windows/macOS 双平台规范](architecture/CROSS_PLATFORM.md)
6. [M0 基础设施与部署验收说明](architecture/DEPLOYMENT_M0.md)：早期准备资料；云端当前状态以 PR #18 的现场验收记录为准。
7. [M0-02B 云端 Go API 操作说明](architecture/CLOUD_GO_API_DEPLOY.md)：独立 systemd/Nginx 配置、部署步骤与真实环境验收门禁。
8. [后端设计（Go 单体）](architecture/BACKEND-DESIGN.md)：目录结构、分层规则、API 契约、身份与数据模型演进、任务与部署。
9. [ADR-0004：简单优先、不做付费/商业化](decisions/0004-simplicity-first.md) · [ADR-0001：M0-03 Go 实施选择](decisions/0001-language-and-modular-monolith.md) · [ADR-0005：M0 基础设施选型](decisions/0005-m0-infrastructure.md) · [全部 ADR](decisions/README.md)

## 产品工作流与商品编辑器

- [产品三区、智赢编辑窗口参考、AI 文案/白底、账号模板及刊登状态](product/PRODUCT-WORKFLOW-AND-EDITOR.md)：明确采集箱 → 产品库 → 产品列表；阶段化实现，不复制智赢源码和长表单。

## 当前 MVP 任务书

- [ERP 后端与 Windows/macOS 桌面客户端 MVP](architecture/ERP-MVP.md)：Tauri 2 桌面客户端、单一后端 API、PostgreSQL、本地开发与后续云端部署。
- [浏览器商品采集插件 MVP](product/BROWSER-COLLECTION-PLUGIN-MVP.md)：Chrome/Edge 插件、Amazon 与 Mercado Libre 卡片一键人工采集、商品详情/五点描述/多规格变体。

当前进度（截至 2026-11-18）：**M0-00 基础 CI、Go 1.26、商品采集 Go + PostgreSQL 后端与 Supabase Auth 已合并**；PR #18 云端 HTTPS、Auth、数据库读回、Windows 登录据现场执行报告 PASS。PR #19 浏览器采集插件仍为 Draft，用户已确认 Amazon 商品真实入库和 Windows 采集箱读回，v0.1.3 新版搜索卡片、高清图片/去重、视频过滤尚待实机回归。PR #20 桌面缓存与历史图片展示去重仍为 Draft，切换速度等待 Windows 验收。macOS 实机未验收；产品库、AI 文案/图片、店铺授权及刊登记录为后续阶段。历史部署说明的“待部署”是旧状态，不代表当前未上线。

## 历史规划与竞品资料

旧版商业化 SaaS 方案已从当前文件树移除，可在历史 Git 提交中查找，**不作为现行需求**。

- [智赢静态分析研究](../research/zying/README.md)：了解业务，不复用第三方私有代码或接口。

优先级：可验证的当前用户需求及平台规则 > ADR-0004 > 现行 FOUNDATION/DELIVERY/DOMAIN/CROSS_PLATFORM > 当前 MVP 任务书 > 旧 ADR > 历史规划/竞品研究。
