# Una Seller 文档索引

> 当前阶段：**自用优先、最小工程、仅 Mercado Libre，不开发支付或商业化功能。**
> 2026-10-26 更新为精简路线。架构是帮助减少返工的指导，不是要求一次装完所有组件的施工清单。

## 现行规范

1. [Foundation v2：第一性原理与最简技术地基](architecture/FOUNDATION.md)
2. [Delivery v2：M0～M3 精简开发路线](architecture/DELIVERY.md)
3. [Domain v2：最小平台数据与写入契约](architecture/DOMAIN_AND_EXECUTION.md)
4. [Windows/macOS 双平台规范](architecture/CROSS_PLATFORM.md)
5. [M0 基础设施、DNS/HTTPS/证书和 Supabase 部署基线](architecture/DEPLOYMENT_M0.md)
6. [ADR-0004：简单优先、不做付费/商业化](decisions/0004-simplicity-first.md) · [ADR-0005：M0 初期部署选型](decisions/0005-m0-infrastructure.md) · [全部 ADR](decisions/README.md)

## 历史规划与竞品资料

- [原合规/模块路线图 v3](product/AI-ERP-合规接入清单与模块路线图.md)：研究时期面向多租户 SaaS 的历史材料，**不作为当前开发需求**。
- [智赢静态分析研究](../research/zying/README.md)：了解业务，不复用第三方私有代码或接口。

优先级：可验证的当前用户需求及平台规则 > ADR-0004 > 现行 FOUNDATION/DELIVERY/DOMAIN/CROSS_PLATFORM > 旧 ADR > 历史规划/竞品研究。
