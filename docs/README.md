# Una Seller 文档索引

> 文档基线：2026-10-26。仓库当前仍处于研究/架构阶段；此处的目标架构不是已经实现的代码。

## 正式规范（新工程以此为准）

1. [技术架构基线](architecture/FOUNDATION.md)：模块化单体、语言分工、技术栈、边界、安全与验收。
2. [Windows/macOS 跨平台规范](architecture/CROSS_PLATFORM.md)：构建目标、原生适配、发布与双系统验证。
3. [领域模型与执行契约](architecture/DOMAIN_AND_EXECUTION.md)：CBT UP、租户隔离、官方 API、命令、队列、同步与 AI。
4. [分阶段交付路线](architecture/DELIVERY.md)：M0–M4 的可交付闭环及最小 CI。
5. [架构决策记录 ADR](decisions/README.md)：为何采用这些语言和组件；变更必须记录新 ADR。

## 旧规划及研究资料

- [早期合规与模块路线图](product/AI-ERP-合规接入清单与模块路线图.md)：2026-10-26 v3 历史规划，内容保留供核对；与正式架构冲突时以 ADR 与 FOUNDATION 为准。
- [智赢静态研究归档](../research/zying/README.md)：独立研究区，不能成为产品的代码或私有接口依赖。

## 规范优先级

官方 API 文档和适用平台规则（经复核） > 已批准 ADR > FOUNDATION / DOMAIN / CROSS_PLATFORM > DELIVERY > 历史规划与逆向研究。
正式 API 端点、权限、请求响应契约必须在真实官方文档和测试账号下验证，不由逆向结果自动推断为已支持。
