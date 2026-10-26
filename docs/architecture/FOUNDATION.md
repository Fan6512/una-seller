# Una Seller 技术架构基线（Foundation v1）

- 状态：**拟采用 / 开发前基线**，由 ADR-0001～0003 承接。
- 日期：2026-10-26。
- 产品定位：**仅面向 Mercado Libre**；Windows / macOS 桌面客户端优先，自用验证后再扩展多租户商业化。
- 当前现状：仓库主要为研究报告和分析器，**尚无可运行客户端、API、迁移或 CI**。本文件是目标，不代表完成。

## 1. 核心约束

1. 官方 Mercado Libre REST API 是平台读写的唯一产品化入口。禁止复用第三方私有网关、Cookie 自动化、反编译代码和平台页面脚本。
2. 桌面端与平台隔离：客户端仅持有 Una Seller 设备会话，**不持有 Meli access/refresh token、应用 client_secret 或服务端模型 Key**。
3. 云端 PostgreSQL 为业务事实源；客户端仅存必要的受控缓存、未提交草稿和设备设置。断网草稿不能伪装为已发布。
4. **模块化单体优先**：前期一个 Go API 进程、可选独立 Go Worker；模块按领域分界，不预设微服务、Event Bus、独立 AI 服务。
5. 正式业务写操作统一走鉴权、校验、审批（按风险）、任务/命令记录、外部执行、回读验证。
6. Go 后端与 TypeScript 桌面端以 OpenAPI 3.1 为接口契约；生成 TS SDK，不手写两份 DTO。
7. AI 只调用业务工具，不直接掌握数据库权限、平台凭据或不受约束的 HTTP/SQL 能力。
8. 兼容 Windows 与 macOS，不引入只能在 Windows 运行的 UI、打印及文件访问依赖。

## 2. 目标语言与职责

| 层级 | 选择 | 职责 | 不能做的事 |
| --- | --- | --- | --- |
| 桌面 UI | React + TypeScript (strict) + Vite | 商品/订单/客服/任务视图、表单和交互 | 不拼平台 Token；不包含业务写入规则 |
| 原生壳 | Tauri 2 + 少量 Rust | 文件/打印/密钥库/更新/窗口权限的跨平台适配 | 不复制 Go 领域逻辑 |
| 业务 API | Go（标准库 net/http + chi 可选） | 租户/权限/商品/订单/审批/API/审计 | 不暴露原始平台适配 DTO |
| 后台任务 | 同一 Go 代码库的 Worker | 同步、限流、重试、批量刊登、回读 | 不将内存队列作为唯一任务存储 |
| 服务端数据 | PostgreSQL + pgx + sqlc + goose（候选） | 事务、约束、精确金额、迁移 | 不使用浮点金额 |
| 客户端数据 | SQLite；SQLCipher 需额外验证 | 分页缓存/草稿/本地索引 | 不默认存全量买家 PII；不假设 Tauri SQL 插件自带 SQLCipher |
| AI | Go 内的独立模块 + 可替换模型适配器 | 只读查询、建议、低风险受控工具 | 不直接提交无审批高风险动作 |
| API 契约 | OpenAPI 3.1 + TS SDK 生成 | 版本化请求响应、错误码、客户端类型 | 不由数据库 ORM 对象直接充当 API DTO |

初期 Redis、消息总线、独立向量库、独立 AI 微服务、独立 Web 卖家工作台均**不属于必需依赖**。引入前以指标与 ADR 决策，不能因为“未来可能需要”而先搭建。

## 3. 逻辑架构

    Windows / macOS Desktop (Tauri + React)
                 │ HTTPS / REST (+ 可选 SSE)
                 ▼
           Go API modular monolith
       ┌─────────┼──────────────────────┐
       │Identity │Catalog / Orders      │
       │Customer │Operations / Approval │
       │AI Tools │Meli Adapter          │
       └─────────┼──────────────────────┘
                 │  PostgreSQL transaction + durable job
            Go Worker(s) ──► Mercado Libre official API
                 │
            AI Provider Adapter ──► model gateway / UnaAPI (optional)

Go 模块通过明确的应用服务接口调用；不允许领域 A 直接写领域 B 的表。HTTP Handler 只做入参校验、鉴权和响应映射；外部 API DTO 在 integration 层转换为领域模型。

## 4. 建议组件（初版锁定/候选）

- UI：shadcn/ui + Base UI（可访问性组件）、Tailwind CSS + CSS variables、Lucide、TanStack Router、TanStack Query、TanStack Table + Virtual、React Hook Form + Zod、Zustand（只限 UI 状态）。
- 需要尽早自建统一 `DataGrid`、`EntityEditor`、`TaskProgress` 三类业务组件；列宽、排序、批量选择、键盘操作与失败明细统一。
- Go：pgx、sqlc、goose、slog、OpenTelemetry、标准 testing；chi 仅在标准路由不够时增加。
- 持久任务：优先评估 PostgreSQL 事务队列（例如 River，**需复核授权、运维与功能范围**）；Redis 仅在缓存/协调/吞吐有实测需求后引入。
- 测试：Go 单测和 PostgreSQL 集成测试、TS 类型测试、React 组件测试、Windows/macOS Tauri 构建及冒烟。

版本策略：首批可运行骨架确定后固定 Go/Node/Rust/Tauri 版本、依赖锁文件和构建镜像；升级通过单独 PR 验证，不在此文档臆定尚未测试的精确版本。

## 5. 数据与安全不变量

- 表含明确的 `tenant_id` 或显式租户归属关系，服务端每次读取按当前 membership 及店铺授权判定，不信任客户端传来的租户 ID。
- 外部账号、站点 Seller、UP / family / local item 分离建模，不用单个 `item_id` 兼任不同平台实体。
- 金额使用 PostgreSQL NUMERIC 或显式精度整数 + currency，禁止浮点数直接用于支付/售价/费用。
- Token Vault 服务端加密存储，密钥与密文分离；OAuth state、刷新锁和 refresh token 轮换测试为上线门槛。
- 日志/错误/崩溃报告/模型请求不记录 token、buyer PII 或可恢复凭据。调试标记在发行构建中关闭。
- 外部商品描述、买家消息、第三方文档视为不可信输入；模型建议经过结构校验、领域规则与审批后才能进入命令系统。
- 含买家 PII 的本地缓存最小化、可清理、可审计；离线设备无法保证远程即时擦除。

## 6. 阶段及架构扩展原则

- M0: Windows/macOS 可构建的 Tauri 骨架 + Go API + PostgreSQL 迁移 + 单店 OAuth 读链路。
- M1: CBT UP 与 MX/BR/AR 本地商品映射、列表、草稿、可回读的受控写操作。
- M2: AI 标题/翻译/属性建议 + 客服草稿，全部人工确认，模型独立替换。
- M3: 订单/履约/促销/报表/跨设备、持久任务与对账。
- M4: 多租户商业化、高级权限、区域化与可观测的水平扩展。

先按照领域接口分层；仅在实测负载、故障隔离、交付独立性需要时拆服务。不为了追求“纯微服务”复制事件系统、账务或 OAuth 服务。

## 7. 建议的未来代码布局（尚未创建应用代码）

    apps/desktop/                 Tauri + React
    backend/cmd/api/              Go HTTP 入口
    backend/cmd/worker/           Go Worker 入口
    backend/internal/identity/
    backend/internal/catalog/
    backend/internal/orders/
    backend/internal/operations/
    backend/internal/ai/
    backend/internal/integration/mercadolibre/
    backend/migrations/
    contracts/openapi.yaml
    docs/architecture/            本规范
    docs/decisions/               ADR
    docs/product/                 产品/历史计划
    research/zying/               竞品研究（仅供参考）

不创建空壳应用工程；第一轮代码 PR 按 M0 落地目录、测试与 CI。

## 8. 变更流程

架构变化须提交 ADR，说明理由、替代方案、迁移与回退，引用可运行证据。新依赖必须评估维护状态、许可证、安全、Win/mac 支持和打包体积。
