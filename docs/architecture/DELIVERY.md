# 分阶段落地与地基验收（v1）

本路线是逐步验证，不是固定工期承诺。先自用，后 SaaS；优先获得一个可回读、可审计的真实业务闭环。

| 里程碑 | 最小交付 | 退出标准 |
| --- | --- | --- |
| M0 连接地基 | Tauri 双端构建、Go API、PostgreSQL 迁移、OpenAPI SDK、单店 OAuth、只读用户/商品 | Windows/macOS 最小冒烟 + 单店正式授权读调用 + Token 轮换/撤销/越权测试 |
| M1 商品业务 | CBT UP / site listing 身份映射、商品列表、草稿、受控修改与回读 | MX/BR/AR 各适用站点的映射验证，重复/超时请求不重复修改 |
| M2 AI 工作流 | 标题/描述/属性建议、中文→西语/葡语校验、客服草稿，人工确认 | 业务规则与字符长度校验、错误可回退、无秘密进入提示 |
| M3 完整运营 | 订单/履约/售后/促销、任务状态、增量同步/对账、跨设备 | 可恢复任务、漏通知补偿、审计完整 |
| M4 多租户 SaaS | 团队权限/计费/治理/合规扩展 | 独立租户隔离测试、服务容量与合规验收 |

## 第一个开发 PR 应完成

1. 目录骨架：apps/desktop、backend/cmd/api、backend/internal、backend/migrations、contracts。
2. 严格 TypeScript、Go 单测、Rust 检查、格式化与依赖锁文件；开发/生产配置分开。
3. PostgreSQL 第一版 schema（租户/用户/店铺连接/授权引用/外部卖家身份）。
4. OpenAPI 契约与 TypeScript SDK 自动生成；一个只读健康/示例 API 实测通过。
5. Windows 本机 dev 启动与 Windows/macOS CI 构建验证，不提交签名密钥。

## 重要的 Do Not Build

首版不建复杂微服务平台、Kafka、大型数据仓库、跨租户向量库、浏览器 DOM 代理、MCP 专用业务后端；没有真实需求先不用 Redis。
不以目录中的逆向分析产物作为产品构建依赖。

## CI 最小门槛

文档整理 PR 仅验证文档链接与完整性；进入应用代码后增加 Go build/test、TypeScript typecheck/test、Rust fmt/clippy/check、API SDK 生成一致性、PostgreSQL 迁移集成测试。macOS 构建可通过手动触发或发布分支以控制 CI 成本；正式发布前双平台实机冒烟必需。
