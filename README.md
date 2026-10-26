# Una Seller — Mercado Libre AI ERP

> **当前阶段：最简自用 ERP 的开发准备。** 只做 Mercado Libre；Windows/macOS 双平台；暂不建设支付、订阅、计费及商业化 SaaS 后台。

## 先读这两份

- [开发原则与技术地基（Foundation v2）](docs/architecture/FOUNDATION.md)：第一性原理、真实问题优先、简单可用、安全不过度。
- [精简开发路线（Delivery v2）](docs/architecture/DELIVERY.md)：M0 能跑 → M1 商品可用 → M2 AI 有用 → M3 按需运营；M4 商业化冻结。

## 其他正式文档

- [Windows/macOS 双平台规范](docs/architecture/CROSS_PLATFORM.md)
- [最小领域与写操作契约](docs/architecture/DOMAIN_AND_EXECUTION.md)
- [架构决策 ADR](docs/decisions/README.md)
- [全部文档索引](docs/README.md)

## 研究资料（与产品代码隔离）

- [智赢静态研究](research/zying/README.md)
- [原合规/路线图 v3（历史）](docs/product/AI-ERP-合规接入清单与模块路线图.md)

**状态说明**：当前仓库的正式内容主要是架构规范和研究文件，M0 还没有完成。原则是先实现一个能启动、能连接、能读取真实店铺数据的程序，再逐项增加业务功能，不提前铺设复杂系统。
