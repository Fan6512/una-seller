# una-seller — Mercado Libre AI ERP

> **阶段：架构准备 / 未开始应用实现。** 项目范围仅 Mercado Libre。客户端目标：Windows + macOS（Intel / Apple Silicon）；不提供独立 Web 卖家工作台。

## 从这里开始

- **[正式架构与技术栈](docs/architecture/FOUNDATION.md)**：Go 模块化单体 + Tauri 2/React/TypeScript + PostgreSQL + OpenAPI。
- **[跨平台规范](docs/architecture/CROSS_PLATFORM.md)**：构建、系统适配、双平台验收。
- **[领域模型与任务执行契约](docs/architecture/DOMAIN_AND_EXECUTION.md)**：CBT UP、订单、任务、同步、AI 权限。
- **[交付路线](docs/architecture/DELIVERY.md)**：M0→M4，由单店实际业务闭环演进。
- **[架构 ADR](docs/decisions/README.md)**：语言、数据库、队列、平台与 AI 的关键决策。
- **[文档导航](docs/README.md)**：全部正式与历史资料。

## 研究资料（与产品代码隔离）

- [智赢静态研究说明](research/zying/README.md)
- [智赢美客多逆向分析报告](research/zying/ZYing-美客多逆向分析报告.md)
- [合规接入与模块路线图（历史 v3）](docs/product/AI-ERP-合规接入清单与模块路线图.md)
- [研究生成器与派生数据](research/zying/analysis/)：原 `analysis/` 目录整体迁移，生成器依赖的内部路径不变。

## 本轮仓库整理

只移动研究资料并新增工程规范，**没有创建可运行 ERP 代码、没有执行真实 Meli 请求、没有重跑静态分析**；研究中的原始绝对路径与机器生成数据仍需独立复核。

## 合规边界

仅通过 Mercado Libre 官方 API 构建自己的 ERP；竞品研究不作为复用私有协议、二进制、反编译源码、OAuth 凭据或 DOM 自动化的依据。实施以官方文档、目标账号授权结果和适用法律为准。
