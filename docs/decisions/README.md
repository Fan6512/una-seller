# 架构决策记录（ADR）

状态仅描述工程选择，不等于已经交付。

- [ADR-0001 语言与模块化单体](0001-language-and-modular-monolith.md)：确定 Go + TypeScript + Tauri/Rust 以及替代方案。
- [ADR-0002 数据、队列与跨设备策略](0002-data-jobs-and-sync.md)：PostgreSQL 事实源、按需缓存、持久任务及同步边界。
- [ADR-0003 Mercado Libre 与 AI 解耦](0003-integration-and-ai-boundaries.md)：平台官方接口、模型工具权限与独立商用边界。

新架构变更新增 ADR，而非直接覆盖历史决定；标注状态、日期、理由、代价、迁移和验收证据。
