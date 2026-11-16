# Una Seller：GitHub Actions CI 与初期运行环境（v1）

> 2026-11-16 · **实施指导，尚未完成部署与 CI 运行验收**。本文件说明 CI 与环境选择，不维护第二份业务里程碑；唯一开发顺序以 [DELIVERY.md](DELIVERY.md) 为准。
>
> 当前仓库事实：公开仓库；`backend/go.mod` 为 Go 1.23；桌面端在 `apps/desktop/`，Node/npm + Vite 构建；尚无 `.github/workflows/`、PostgreSQL 业务迁移或浏览器插件代码。不得将计划写成已实现。

## 1. 运行与构建的分工

| 位置 | 职责 | 现阶段不承担 |
| --- | --- | --- |
| GitHub Actions 标准托管 Runner | PR 自动 Go 编译/单测、TS/Vite 构建；PR-A 起使用临时 PostgreSQL；PR-B 起增加 Win/macOS Tauri 原生构建；PR-C 起检查 MV3 插件 | 生产数据库、真实店铺 Token、常驻应用服务器 |
| 本地 Windows 开发电脑 | Codex 编写和审查代码；可用 Vite 预览普通 UI；需要 Windows 原生启动验收时再安装 Rust/Cargo 及 Windows 所需工具链 | 必须本机同时安装 Go、PG、Docker、Rust 的假设 |
| 2C4G Linux 云主机（**建议配置，未购置/未部署/未压测**） | 后续运行一个 Go API、一个 PostgreSQL、HTTPS 反向代理（Caddy 或现有代理二选一）；持久化业务数据与备份 | GitHub CI 构建、AI 大模型推理、复杂任务系统 |

初期可选 2 vCPU / 4GB RAM / SSD 60–80GB / Ubuntu 24.04 LTS，前提是实际负载较低、商品图片起初只存来源 URL、AI 文案/白底调用外部 API。资源是否够用依实际 CPU、RAM、I/O、存储和延迟监控判断；不宣称“保证支持多少商品或账号”。

- 云端 API 暴露公网前必须启用 HTTPS、有效 ERP 身份认证；PostgreSQL `5432` 不直接开放公网。开发数据库/CI 临时数据库与正式数据隔离；凭据通过环境变量或部署环境机密配置，不进公开仓库或构建产物。
- 后期部署可用简单 Docker Compose 管理 Go + PostgreSQL + HTTPS 反代，但**当前不提前提交生产 Compose/Dockerfile 作为既成事实**，也不覆盖仍在独立评审的 M0-02A HTTPS 方案；等云服务器及域名条件确定再选最终反代。
- PostgreSQL 数据卷持久化，备份应包含可恢复的异地副本；不能因为 Docker volume 存在就声称有备份。公网端点的账号和数据库权限按最低需要配置。
- GitHub 标准托管 Runner 可编译 Win/macOS 目标，但 **Vite 通过 ≠ Tauri 原生编译通过 ≠ 真实设备 GUI 启动通过**；无实机时标 `UNVERIFIED`。
- 前端实际 API origin、Tauri `connect-src` 与图片 `img-src` 需在 PR-B 一起对齐；允许所需 HTTPS 来源，勿因 CSP 默认值使商品图片不可见，也不要为解决图片问题全面关闭 CSP。

## 2. 公开仓库 CI：按模块出现逐步增加

GitHub 官方规则：公开仓库使用标准 GitHub-hosted Runner 免费；大型 Runner 有独立收费，Artifacts/缓存有存储额度约束。官方参考：
- [GitHub Actions billing](https://docs.github.com/en/billing/concepts/product-billing/github-actions)
- [Hosted runners](https://docs.github.com/en/actions/reference/runners/github-hosted-runners)
- [Actions limits](https://docs.github.com/en/actions/reference/limits)

### CI-0：先独立提交最小 CI PR（不改变 Issue #6）

建立 `.github/workflows/ci.yml`；`pull_request` 到 `main` 与 `push` 到 `main` 触发；默认 `permissions: contents: read`，同一 PR 的旧运行可被新提交取消；不用 `pull_request_target` 执行不可信 PR 代码，不注入生产密钥。

- **Go job / Ubuntu**：在 `backend/` 执行 `gofmt` 检查、`go vet ./...`、`go test ./...` 和 `go build ./...`；采用 `backend/go.mod` 声明的 Go 版本。当前无 `go.sum`，不要把不存在的文件设为必需缓存条件。
- **桌面前端 job / Ubuntu**：在 `apps/desktop/` 执行 `npm ci` 和 `npm run build`；使用现有 `package-lock.json`，Node 版本与项目工具链相容（建议 Node 22）。
- 只使用当前确实存在的脚本；不在仓库根目录误运行 Go 和 npm，也不提前构建不存在的浏览器插件。
- CI-0 先在自己的 Draft PR 中至少跑过一次，再讨论将稳定 job 设为 `main` 的 required checks；公开仓库无需为此预设购买 GitHub Pro。
- 安装包仅在需要验证时上传、设置短保留期；不让每个 PR 长期留存大文件。无 `.github/workflows` 时不能声称 CI 已通过。

### CI-1：Issue #6 / PR-A

Go backend job 追加临时 PostgreSQL Service（例如 PostgreSQL 16，与项目后续部署选用主版本保持兼容），使用**仅 CI 使用的合成账号与合成商品**、迁移与集成测试验证：
1. 商品写入后列表/详情读回，重启后仍存在。
2. 同一“用户 + 平台 + 来源站点 + 规范化 URL”提交不重复建商品。
3. 部分采集、空字段不清除旧的图片/卖点/完整变体。
4. 货币/金额精确处理，缺失值保留空值；错误响应不误报成功。

仅允许在 CI 临时数据库上执行建表或清空；**禁止连接云端正式数据库**。数据契约仍以 [ERP-MVP.md](ERP-MVP.md) 与 [DOMAIN_AND_EXECUTION.md](DOMAIN_AND_EXECUTION.md) 为准，不把临时测试表设计变成生产合同。

### CI-2：Issue #6 / PR-B

前端 job 保持快速反馈；新增 Windows x64 和 macOS arm64/Intel 的 Tauri 原生构建检查（用实际支持的 Runner、Rust/Tauri 系统依赖），记录失败原因。优先在桌面相关 PR 触发，避免文档改动每次跑跨平台安装包；真实 Windows/macOS GUI 启动需独立验收，不能被 CI build 取代。

### CI-3：Issue #6 / PR-C

`apps/browser-extension/` 实际建立后，增加 Chrome/Edge MV3 manifest、类型/构建、必要解析单测；只检查人工触发 Amazon.com 单商品采集的逻辑，不在 PR CI 中自动访问真实 Amazon 商品页面或测试真实店铺。

### 后续：发布与部署

当 Go API 在云端可部署且公网鉴权、域名和数据库备份都已具备后，再引入手动触发的测试部署。正式部署需要独立批准，默认 **PR CI 不自动 SSH 到生产云主机，不执行生产迁移，不处理真实 OAuth/店铺写操作**。

## 3. 不要把其他阶段的能力塞进 PR-A

- **M1-D / OAuth**：只有店铺授权真正开发时才实现 Token 刷新协调。可用 PostgreSQL 事务和行锁控制并发刷新，无需一开始引入 Redis；必须保存新 refresh token、确认官方当前失效/轮换行为，不能用过期的固定时间或“每五小时刷新”作为事实。
- **M1-E / 刊登**：先从官方账号信息判断是否启用 `user_product_seller`，据当前 User Products / Legacy 文档适配对应模型；不可把旧 `/global/items` 流程当通用路径。选择多个站点可以对应一次或多次请求，逐站点读取真实结果，HTTP 200 不等于所有站点成功。参考 [官方 User Products](https://global-selling.mercadolibre.com/devsite/manage-sales-global-selling/user-products-cbt) 与 [Legacy publishing](https://global-selling.mercadolibre.com/devsite/global-listing)。
- **M3 / 通知、订单和客服**：Webhook 需求明确后，再评估事务保存通知并异步处理；简单 PostgreSQL 持久队列可以替代早期 Redis 队列，但这是**后续设计候选**，具体 topic、签名/验证、返回时限与消息 API 必须到实施时重新核对官方资料。
- **AI**：文案/白底走外部 Provider API，候选保留并人工确认；2C4G 不承担大型图像模型计算。官方 Mercado Libre 图片质量诊断/上传与 AI 白底编辑不是同一个接口，未经验证不假设开放第三方修图 API。

## 4. 何时可以宣称完成

| 说法 | 必须有的证据 |
| --- | --- |
| 基础 CI 已建立 | `ci.yml` 在真实 PR 上运行并输出明确通过/失败记录 |
| PR-A 后端可用 | GitHub 临时 PG 集成通过 + 可重复的真实 PG 入库/读回证据 |
| PR-B Windows/macOS 可用 | 对应原生构建成功；**实际桌面启动**额外由真实 OS 设备验证 |
| API 云部署可用 | 已有用户鉴权、HTTPS、域名、隔离数据库及持久化/备份验证 |
| Mercado Libre 刊登可用 | 官方账号权限确认、实际提交、逐站点结果和平台回读证据 |

> 本文不修改 GitHub Actions 文件、业务代码或生产服务器；它仅为后续独立小 PR 明确配置及验收边界。
