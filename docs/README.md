# Una Seller 文档索引

> 当前阶段：**自用优先、最小工程、仅 Mercado Libre 运营；Windows/macOS 双平台桌面客户端；浏览器插件采集 Amazon 与 Mercado Libre 竞品。**
> 2026-11-15 更新产品三区工作流、轻量内置插件、AI 编辑、多站点刊登的正式规划；浏览器采集与采集箱 MVP 继续按阶段实施。架构用于减少返工，不要求一次装完所有组件。

## 现行架构规范

1. [Foundation v2.1：第一性原理与最简技术地基](architecture/FOUNDATION.md)
2. [Delivery v2.1：M0～M3 精简开发路线](architecture/DELIVERY.md)
3. [Domain v2.1：最小平台数据与写入契约](architecture/DOMAIN_AND_EXECUTION.md)
4. [CI 与运行环境 v1：GitHub Actions、开发/部署分工与验收边界](architecture/CI_AND_RUNTIME.md)
5. [Windows/macOS 双平台规范](architecture/CROSS_PLATFORM.md)
6. [M0 基础设施与部署验收说明](architecture/DEPLOYMENT_M0.md)：既有 Ubuntu/Nginx + Supabase，部署尚未验收。
7. [M0-02B 云端 Go API 操作说明](architecture/CLOUD_GO_API_DEPLOY.md)：独立 systemd/Nginx 配置、部署步骤与真实环境验收门禁。
7. [ADR-0004：简单优先、不做付费/商业化](decisions/0004-simplicity-first.md) · [ADR-0001：M0-03 Go 实施选择](decisions/0001-language-and-modular-monolith.md) · [ADR-0005：M0 基础设施选型](decisions/0005-m0-infrastructure.md) · [全部 ADR](decisions/README.md)

## 产品工作流与商品编辑器

- [产品三区、智赢编辑窗口参考、AI 文案/白底、账号模板及刊登状态](product/PRODUCT-WORKFLOW-AND-EDITOR.md)：明确采集箱 → 产品库 → 产品列表；阶段化实现，不复制智赢源码和长表单。

## 当前 MVP 任务书

- [ERP 后端与 Windows/macOS 桌面客户端 MVP](architecture/ERP-MVP.md)：Tauri 2 桌面客户端、单一后端 API、PostgreSQL、本地开发与后续云端部署。
- [浏览器商品采集插件 MVP](product/BROWSER-COLLECTION-PLUGIN-MVP.md)：Chrome/Edge 插件、Amazon 与 Mercado Libre 卡片一键人工采集、商品详情/五点描述/多规格变体。

当前进度：**M0-00 基础 CI、Go 1.26 与 Issue #6 / PR-A 商品采集 Go + PostgreSQL 17 后端已合并**；Supabase `una-seller` 已执行 PR-A 的 `m0_03_source_collections` 迁移和 `restrict_auto_rls_event_trigger_execution` 权限收口（2026-11-16），四表 RLS 开启且空表；合成 SQL 事务验证已回滚，**Go API 云库连通与公网部署未验收**。后续实施 **PR-B 桌面采集箱 → PR-C 浏览器扩展**；Win/mac 原生构建与实机验收独立登记。产品库、AI 文案/图片、店铺模板和刊登记录为后续 M1 阶段，规划不等于已交付。

## 历史规划与竞品资料

旧版商业化 SaaS 方案已从当前文件树移除，可在历史 Git 提交中查找，**不作为现行需求**。
- [智赢静态分析研究](../research/zying/README.md)：了解业务，不复用第三方私有代码或接口。

优先级：可验证的当前用户需求及平台规则 > ADR-0004 > 现行 FOUNDATION/DELIVERY/DOMAIN/CROSS_PLATFORM > 当前 MVP 任务书 > 旧 ADR > 历史规划/竞品研究。
