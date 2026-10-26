# ADR-0001：采用 Go + TypeScript + Tauri 2 的模块化单体

- 日期：2026-10-26
- 状态：**Accepted as design baseline; implementation pending**
- 场景：Una Seller 仅面向 Mercado Libre，优先 Windows/macOS 桌面端，长期需多站点商品、订单、AI 任务和多租户运营。

## 决策

共享桌面 UI 使用 React + TypeScript (strict) + Vite，原生边界使用 Tauri 2 + 少量 Rust；服务端采用 Go 模块化单体，任务 Worker 与 HTTP API 共代码库，按需分进程；通信契约为 OpenAPI 3.1。

只允许在明确的边界处使用不同语言：React 负责交互，Rust 负责 OS 接口，Go 负责领域/授权/同步/AI 业务。禁止将业务逻辑分别复制到三端。

## 理由

Go 适合长期运行的 I/O 密集型连接、任务与平台 API 访问，运维负担较低；React 适合商品列表、批量编辑、复杂交互；Tauri 允许复用 UI 并有受限原生调用。
模块化单体比预先拆很多微服务更少网络一致性问题，且保留按领域拆分的可能。

## 备选与代价

- 全 TypeScript + Fastify/NestJS：单人研发更容易；可作为备选，但默认不采用以避免高负载 Worker 与核心业务过多共享 Node 运行时行为。没有客观性能数据前不宣称 Go 必然更快。
- Electron：组件和生态成熟，但不作为首选；如果 WebView 兼容或原生插件出现明确不可克服障碍再重评。
- 直接 Swift/C#：系统集成便利，但将导致两套 UI/业务实现。

代价是多语言工具链和 OpenAPI 类型生成；以生成 SDK、模块契约和 CI 控制。
