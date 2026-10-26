# 架构决策记录（ADR）

原则：**少写、只记录影响大且难回退的决定**。普通开发和组件替换不必专门立 ADR。

- [ADR-0004 第一性原理、简单优先、暂不商业化](0004-simplicity-first.md)：当前最高优先级；明确 M0～M3、M4 冻结及不过度设计。
- [ADR-0005 早期部署与证书选型](0005-m0-infrastructure.md)：Cloudflare + Nginx + Go/Ubuntu 2C4G + Supabase；Let's Encrypt 优先。
- [ADR-0001 跨平台与 Go 单体后端](0001-language-and-modular-monolith.md)：Tauri/React 已确定；M0 采用 Go API，保留事实驱动的重评估空间。
- [ADR-0002 数据、任务与渐进同步](0002-data-jobs-and-sync.md)：M0-02 使用 Supabase 托管 PostgreSQL，只建必要表与任务；Redis/同步平台延后。
- [ADR-0003 Mercado Libre 与 AI 解耦](0003-integration-and-ai-boundaries.md)：原则有效，具体实现遵循最小需求。

旧决策或历史规划与 ADR-0004 冲突时，按新的简单优先原则执行，不预设成熟商业 SaaS 规模。
