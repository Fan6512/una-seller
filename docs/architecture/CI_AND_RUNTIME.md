# Una Seller：GitHub Actions CI 与初期运行环境（v1）

> 2026-11-16 · **CI 与验收指导：M0-00 基础 CI 已建立并在真实 PR 上通过；数据库接入与云端部署仍未验收**。本文只定义 CI 的技术梯度与运行环境分工；唯一业务里程碑以 [DELIVERY.md](DELIVERY.md) 为准。
>
> **基础设施选型依据**：[ADR-0005](../decisions/0005-m0-infrastructure.md) 与 [M0 部署基线](DEPLOYMENT_M0.md)：沿用**已有 Ubuntu 服务器 + 现有 Nginx + 单 Go API + Supabase 托管 PostgreSQL**，不再设计自建 PG 或 Caddy 作为默认替代。服务器信息来自历史 SSH 记录；该选型**不代表 HTTPS、数据库连接或业务服务已经部署验收**。本文件仅补充公开仓库 CI、工具链与测试边界，不覆盖基础设施决策。
>
> 当前仓库事实：公开仓库；`backend/go.mod` 为 Go 1.26；桌面端在 `apps/desktop/`，Node/npm + Vite 构建；尚无 PostgreSQL 业务迁移或浏览器插件代码。不得将计划写成已实现。

## 1. 运行与构建的分工

| 位置 | 职责 | 现阶段不承担 |
| --- | --- | --- |
| GitHub Actions 标准托管 Runner | PR 自动 Go 编译/单测、TS/Vite 构建；PR-A 起使用临时 PostgreSQL；PR-B 起增加 Win/macOS Tauri 原生构建；PR-C 起检查 MV3 插件 | 生产数据库、真实店铺 Token、常驻应用服务器 |
| 本地 Windows 开发电脑 | Codex 编写和审查代码；可用 Vite 预览普通 UI；需要 Windows 原生启动验收时再安装 Rust/Cargo 及 Windows 所需工具链 | 必须本机同时安装 Go、PG、Docker、Rust 的假设 |
| **已有** Ubuntu Linux 服务器（ADR-0005 记录的历史 SSH 检查，部署状态待复核） | 现有 Nginx + 后续单 Go API；通过 TLS 访问 **Supabase 托管 PostgreSQL**，维护原有 `we-mp-rss` 服务 | 自建 PostgreSQL、Caddy、CI 编译、大型 AI 推理 |

ADR-0005 记录的**历史 SSH 观察**：Ubuntu 24.04.4 LTS、2 vCPU、约 3.6GiB RAM（当时 available 约 2.3GiB）、59GB 根盘（当时剩余约 37GB）、约 980MiB swap 已用、Nginx 80 端口在用、已有 Docker `we-mp-rss`。这些不是本轮重新 SSH 核验的当前运行指标，也**不代表 Una Seller 已部署**。部署前重新检查 `free -h`、`vmstat`、磁盘、Nginx 监听及既有服务；swap 已用并不能单凭数值判定内存不足。此配置能否支撑新 Go 服务须依实测，优先不在同机额外部署 PostgreSQL。

- 云端 API 暴露公网前必须启用 HTTPS、有效 ERP 身份认证；**数据库采用 Supabase 托管 PostgreSQL**，仅 Go 服务持有其 TLS 连接凭据，不让桌面端直连，不公开本机数据库端口。GitHub CI 的临时 PostgreSQL 与 Supabase 正式/测试项目隔离；凭据不进公开仓库或构建产物。
- 部署依据 [ADR-0005](../decisions/0005-m0-infrastructure.md)，采用 **Cloudflare `mkd.186512.xyz` + 现有 Nginx + Go 进程（候选 systemd）+ Supabase**；[PR #4](https://github.com/Fan6512/una-seller/pull/4) 的 Nginx/HTTPS 配置仍是独立代码候选。不得覆盖现有 Nginx、`we-mp-rss`，也不得将 Caddy、自建 PostgreSQL 或 Compose 视为默认方案；部署与证书均需单独验证。
- Supabase 的持久化与备份能力、免费计划限制、恢复责任需要在项目实际创建后核实；**托管数据库不等于已经具备可恢复的备份**。PR-B 需要独立可达的真实测试数据库及 Go API，具体前置条件见下文 CI-2。
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
- **桌面前端 job / Ubuntu**：在 `apps/desktop/` 执行 `npm ci` 和 `npm run build`；使用现有 `package-lock.json`，配置 Node **`22.x`（实际版本不得低于 `22.12`）**，满足 Vite 7 的最低版本要求。参考 [Vite 7 官方迁移说明](https://v7.vite.dev/guide/migration)。
- 只使用当前确实存在的脚本；不在仓库根目录误运行 Go 和 npm，也不提前构建不存在的浏览器插件。
- CI-0 先在自己的 Draft PR 中至少跑过一次，再讨论将稳定 job 设为 `main` 的 required checks；公开仓库无需为此预设购买 GitHub Pro。
- 安装包仅在需要验证时上传、设置短保留期；不让每个 PR 长期留存大文件。无 `.github/workflows` 时不能声称 CI 已通过。

### CI-1：Issue #6 / PR-A

Go backend job 追加 GitHub-hosted Ubuntu 的临时 PostgreSQL Service（**先按 PostgreSQL 16 配置**；Supabase 项目建成后若主版本不同再对齐），使用**仅 CI 使用的合成账号与合成商品**、迁移与集成测试验证：
1. 商品写入后列表/详情读回，**断开数据库连接、重新建立连接后仍能读回**；这并不等于数据库进程或机器重启持久性已验收。
2. 同一“用户 + 平台 + 来源站点 + 规范化 URL”提交不重复建商品。
3. 部分采集、空字段不清除旧的图片/卖点/完整变体。
4. 货币/金额精确处理，缺失值保留空值；错误响应不误报成功。

仅允许在 CI 临时数据库上执行建表或清空；**禁止连接 Supabase 正式数据库**。GitHub Service Container 由 job 管理并在结束时销毁，CI-1 **只负责断连重连读回**；真实进程/数据库重启持久性另在本地 Docker 或获批准的受控环境完成，未执行则标 `NOT_RUN`，不得因 CI 绿而声称“重启通过”。数据契约仍以 [ERP-MVP.md](ERP-MVP.md) 与 [DOMAIN_AND_EXECUTION.md](DOMAIN_AND_EXECUTION.md) 为准。

### CI-2：Issue #6 / PR-B

前端 job 保持快速反馈；新增 Windows x64 与 macOS arm64 / Intel 的 Tauri 原生构建，按实际支持的 Runner 与 Rust/Tauri 依赖配置。建议起始标签：`windows-latest`、`macos-15`（arm64）和 **`macos-15-intel`（x86_64）**；`macos-13` 已在 2025-12-04 退役。Intel 标准 Runner 预计在 2027 年 macOS 15 镜像退役时结束支持，真正实施时再核对官方可用标签。Intel 初期只做 unsigned build，不要求签名/公证。参考 [GitHub Runner 退役公告](https://github.blog/changelog/2025-09-19-github-actions-macos-13-runner-image-is-closing-down/)。

**PR-B 的强制前置**：至少一套经明确批准、真实可达的开发/测试 PostgreSQL（优先**独立 Supabase 测试项目/数据库**，也可本机 PG）及连接它的 Go API，能够预先写入合成商品。CI-1 的临时 PostgreSQL 不会跨 job 保留，**不能充当 PR-B Windows 实机验收的数据源**。若尚无可用测试库、API 地址与最小身份/凭据配置，就将“原生客户端显示真实持久化记录”标 `BLOCKED`，不以 Mock/Vite 或 Actions 构建冒充通过；测试库启用和公网测试 API 部署均需单独审批、不得默认接生产。

优先在桌面相关 PR 触发原生构建，避免纯文档变更重复耗时；真实 Windows/macOS GUI 启动需独立验收，不能被 CI build 取代。

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
| PR-A 后端可用 | GitHub 临时 PG 集成通过 + 合成商品新增/去重/断连重连读回证据；**重启持久性另验**，未运行不得声称通过 |
| PR-B Windows/macOS 可用 | 对应原生构建成功 + **有可达测试 PG 的 Go API** + 真实设备启动、显示已持久化商品；未实测的设备/架构单独标 `UNVERIFIED` |
| API 云部署可用 | 已有用户鉴权、HTTPS、域名、隔离数据库及持久化/备份验证 |
| Mercado Libre 刊登可用 | 官方账号权限确认、实际提交、逐站点结果和平台回读证据 |

> 本文不修改 GitHub Actions 文件、业务代码或生产服务器；它仅为后续独立小 PR 明确配置及验收边界。
