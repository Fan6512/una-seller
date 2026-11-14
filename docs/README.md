# Una Seller 文档索引

> 当前阶段：**自用优先、最小工程、仅 Mercado Libre 运营；Windows/macOS 双平台桌面客户端；浏览器插件采集 Amazon 与 Mercado Libre 竞品。**
> 2026-11-15 增加浏览器采集插件与 ERP MVP 任务书。架构用于减少返工，不要求一次装完所有组件。

## 现行架构规范

1. [Foundation v2：第一性原理与最简技术地基](architecture/FOUNDATION.md)
2. [Delivery v2：M0～M3 精简开发路线](architecture/DELIVERY.md)
3. [Domain v2：最小平台数据与写入契约](architecture/DOMAIN_AND_EXECUTION.md)
4. [Windows/macOS 双平台规范](architecture/CROSS_PLATFORM.md)
5. [ADR-0004：简单优先、不做付费/商业化](decisions/0004-simplicity-first.md) 与 [全部 ADR](decisions/README.md)

## 当前 MVP 任务书

- [ERP 后端与 Windows/macOS 桌面客户端 MVP](architecture/ERP-MVP.md)：Tauri 2 桌面客户端、单一后端 API、PostgreSQL、本地开发与后续云端部署。
- [浏览器商品采集插件 MVP](product/BROWSER-COLLECTION-PLUGIN-MVP.md)：Chrome/Edge 插件、Amazon 与 Mercado Libre 卡片一键人工采集、商品详情/五点描述/多规格变体。

两份任务书是待实施需求基线，不代表代码已完成，也不自动授权一次性开发所有后续业务功能。

## 历史规划与竞品资料

- [原合规/模块路线图 v3](product/AI-ERP-合规接入清单与模块路线图.md)：研究时期面向多租户 SaaS 的历史材料，**不作为当前开发需求**。
- [智赢静态分析研究](../research/zying/README.md)：了解业务，不复用第三方私有代码或接口。

优先级：可验证的当前用户需求及平台规则 > ADR-0004 > 现行 FOUNDATION/DELIVERY/DOMAIN/CROSS_PLATFORM > 当前 MVP 任务书 > 旧 ADR > 历史规划/竞品研究。
