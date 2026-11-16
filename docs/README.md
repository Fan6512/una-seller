# Una Seller 文档索引

> 当前阶段：**自用优先、最小工程、仅 Mercado Libre 运营；Windows/macOS 双平台桌面客户端；浏览器插件采集 Amazon 与 Mercado Libre 竞品。**
> 2026-11-15 更新产品三区工作流、轻量内置插件、AI 编辑、多站点刊登的正式规划；浏览器采集与采集箱 MVP 继续按阶段实施。架构用于减少返工，不要求一次装完所有组件。

## 现行架构规范

1. [Foundation v2.1：第一性原理与最简技术地基](architecture/FOUNDATION.md)
2. [Delivery v2.1：M0～M3 精简开发路线](architecture/DELIVERY.md)
3. [Domain v2.1：最小平台数据与写入契约](architecture/DOMAIN_AND_EXECUTION.md)
4. [CI 与运行环境 v1：公开 GitHub Actions、2C4G、构建/验收边界](architecture/CI_AND_RUNTIME.md)
5. [Windows/macOS 双平台规范](architecture/CROSS_PLATFORM.md)
6. [ADR-0004：简单优先、不做付费/商业化](decisions/0004-simplicity-first.md) · [ADR-0001：M0-03 Go 实施选择](decisions/0001-language-and-modular-monolith.md) · [全部 ADR](decisions/README.md)

## 产品工作流与商品编辑器

- [产品三区、智赢编辑窗口参考、AI 文案/白底、账号模板及刊登状态](product/PRODUCT-WORKFLOW-AND-EDITOR.md)：明确采集箱 → 产品库 → 产品列表；阶段化实现，不复制智赢源码和长表单。

## 当前 MVP 任务书

- [ERP 后端与 Windows/macOS 桌面客户端 MVP](architecture/ERP-MVP.md)：Tauri 2 桌面客户端、单一后端 API、PostgreSQL、本地开发与后续云端部署。
- [浏览器商品采集插件 MVP](product/BROWSER-COLLECTION-PLUGIN-MVP.md)：Chrome/Edge 插件、Amazon 与 Mercado Libre 卡片一键人工采集、商品详情/五点描述/多规格变体。

当前执行顺序是 **先建最小 CI Draft PR，再实施 M0-03 的 PR-A 后端 → PR-B 桌面 → PR-C 浏览器扩展**。PR-A 添加临时 PostgreSQL 集成测试；Win/mac 原生构建和插件测试跟随对应阶段。产品库、AI 文案/图片、店铺模板、发布记录和已刊登产品列表列入后续 M1，文档规划不等于已完成或已授权全部同时开发。

## 历史规划与竞品资料

- [原合规/模块路线图 v3](product/AI-ERP-合规接入清单与模块路线图.md)：研究时期面向多租户 SaaS 的历史材料，**不作为当前开发需求**。
- [智赢静态分析研究](../research/zying/README.md)：了解业务，不复用第三方私有代码或接口。

优先级：可验证的当前用户需求及平台规则 > ADR-0004 > 现行 FOUNDATION/DELIVERY/DOMAIN/CROSS_PLATFORM > 当前 MVP 任务书 > 旧 ADR > 历史规划/竞品研究。
