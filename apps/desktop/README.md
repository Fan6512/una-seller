# Una Seller 桌面端：登录与采集箱

桌面端使用 **Supabase Auth** 登录，通过 Go API 读取 **PostgreSQL 中当前登录用户的来源商品**；不包含假数据、页面爬虫或 Mercado Libre 店铺授权。

## 本地运行

1. 确认 Go 1.26、Node.js 22+、当前 Rust stable 和 Tauri 2 系统依赖已安装。桌面包声明 Rust 1.90 为最低版本，CI 使用 stable 并以 `--locked` 构建。
2. Windows 首次在 `%LOCALAPPDATA%\\UnaSeller\\api-config.json` 保存 `DATABASE_URL`、`SUPABASE_URL`、`SUPABASE_PUBLISHABLE_KEY`（参考[后端说明](../../backend/README.md)）；以后重新下载 Go API 不再重填。执行 `go run ./cmd/api migrate`（仅未迁移时），然后 `go run ./cmd/api`。
3. 在 Supabase Dashboard 的 **Authentication → Users** 创建/邀请一个允许登录的邮箱账号，确认邮箱和密码可用。桌面端内置的 `VITE_SUPABASE_URL` 和 `VITE_SUPABASE_PUBLISHABLE_KEY` 默认指向本仓库的 Supabase 项目，切换项目时必须同时修改构建环境与 CSP。
4. 在 `apps/desktop/` 执行 `npm ci`、`npm run tauri:dev`（仅源码开发时需要 Rust/Tauri 工具链；使用 CI 产物时不需要）；用邮箱密码登录后进入采集箱。
5. 桌面采集箱请求 `GET /api/v1/collections` 和详情 `GET /api/v1/collections/{id}`；列表如果为空会明确显示“这里还没有商品”，**不会伪造演示记录**。
6. **开发模式**（`tauri:dev`）默认访问 `http://127.0.0.1:8787`；**正式打包版本**默认访问 `https://mkd.186512.xyz`（云端上线后才可用）。可用 `VITE_API_BASE_URL` 在构建时显式覆盖；这是**构建时**选项，不是安装后动态设置。Tauri CSP 只放行本地 API 和准确的云 API 域名，更换域名仍需同步修改 CSP 与 Go CORS。不要将数据库 URL、密码、服务端 owner ID 或密钥放进 Vite 环境变量。

## Issue #14 登录说明

- Auth 后端使用已有 Supabase 项目的开源 **GoTrue / Supabase Auth**；桌面通过官方 HTTP API 实现邮箱密码登录、自动续签、恢复会话和退出登录，**不额外部署认证服务、也不创建密码表**。
- 本地测试账号已由负责人创建；当前先不阻断公开注册配置。注意 **Supabase Auth 是公网服务**，只把 Go API 放在本机并不阻止外部调用注册接口；正式分发/公网开放前须关闭注册。
- 登录会话由桌面 WebView 的本地存储保留，不包含用户密码；客户端发起业务请求时自动携带 Bearer Token。Go 服务端通过 `GET /auth/v1/user` 验证令牌并从服务端响应读取真实用户 ID，不接受客户端提交 owner UUID。
- 若已安装的是 PR #13 的老版本，老版本尚无登录页面；请下载 Issue #14 对应的新版 Windows Artifact。旧本地 API + 旧客户端组合不能作为新 Auth 版本的验收证据。
- 未对旧 UUID 对应的测试商品做迁移，避免把其他用户记录误归属到新账号。当前不开放自助注册、忘记密码入口、Google 登录或团队权限。

## Windows 构建产物验收（无需本机 Go/Rust 工具链）

从 **[Windows Desktop Build 工作流](https://github.com/Fan6512/una-seller/actions/workflows/windows-desktop.yml)** 选择与本 PR 当前 HEAD（或拟验收提交）一致的**最新成功运行**，下载 `una-seller-windows-x64` Artifact，解压后使用。请核对运行详情的提交 SHA，**不要使用旧版运行或过期的固定 Run 链接**；Artifact 可能过期，必要时手动触发当前分支的构建。旧版 PR #13 产物没有登录功能，早期 PR #15 产物也不支持持久化配置。

- `una-seller.exe`：Tauri Windows 桌面客户端。
- `una-seller-api.exe`：独立的本地 Go API 程序；**它只在 Artifact 中，不包含在 NSIS `*-setup.exe` 安装包内**。
- `Una Seller_<版本>_x64-setup.exe`：安装桌面客户端。Go API 单独运行。单独安装后看到“连接异常”不代表桌面程序安装失败。

测试真实商品读取时，解压 Artifact 后，在首次创建的 `%LOCALAPPDATA%\\UnaSeller\\api-config.json` 中保存配置。在 `una-seller-api.exe` 所在目录的 PowerShell 中执行 `.\\una-seller-api.exe`，保持 API 运行，再打开 `una-seller.exe` 或已安装的客户端。**此后换新版 EXE 直接运行，不再逐次配置**；Go API 继续监听本机 `127.0.0.1:8787`。

- 不要将数据库连接字符串/密码贴到 issue、PR、截图、公开 Actions 日志，也不要写入客户端环境变量；仅本机 Go API 进程需要数据库凭据。
- **新 Supabase Auth 用户 ID 与旧本地 UUID 不同。** 以前手动创建的测试商品仍在数据库原用户下面，登录新账号后列表为空属于预期；未经授权不自动迁移或改写商品所属用户。
- Windows 原生构建通过与 GUI 启动通过是两件事；macOS 状态另计。

## Windows x64 正式更新

合并代码与正式发布准入分别验收。首次推送正式 tag 前，负责人须留存设置截图或只读配置证据，并逐项确认：

- `desktop-production` 环境已预先创建，Required reviewers 包含负责人，禁用管理员绕过，部署只允许 `desktop-v*` tag。单人维护且推 tag 与审批为同一人时关闭 Prevent self-review。该审批是同一负责人的显式批准，增加第二维护者后再启用独立审批。
- 仓库存在生效的 tag ruleset，仅匹配 `desktop-v*`。核对规则和绕过名单，确认未经授权的创建、更新和删除均被拦截。
- 从 GitHub Settings 核对实际生效配置。环境可用 `gh api repos/Fan6512/una-seller/environments/desktop-production` 读取，tag 规则可用 `gh api repos/Fan6512/una-seller/rulesets` 定位并读取详情。截图或输出仅保留非敏感设置。

通过上述检查后，按下列步骤发行：

1. 经单独授权后，将负责人保管的 Tauri Updater 公钥配置为环境变量 `TAURI_UPDATER_PUBLIC_KEY`。私钥内容存为该环境的 `TAURI_SIGNING_PRIVATE_KEY` Secret，有密码时将密码存为 `TAURI_SIGNING_PRIVATE_KEY_PASSWORD` Secret。保留离线备份，沿用同一密钥对供已安装客户端验证。
2. 从 `main` 上批准的提交创建 `desktop-vMAJOR.MINOR.PATCH` tag。递增 `tauri.conf.json`、`Cargo.toml`、`package.json` 和 `package-lock.json` 的版本，并更新 `Cargo.lock` 中的桌面包版本。环境审批后，工作流签名 NSIS 包、验证签名和版本、生成 `latest.json`，一起提交到 GitHub Release 草稿。
3. 负责人核对草稿的版本、更新说明和三个资产（安装包、签名、`latest.json`），按桌面版本递增顺序手动发布并设为 **Latest**。生产客户端从仓库的 `releases/latest/download/latest.json` 读取更新。保持 Latest 指向包含该清单的 Windows 桌面正式版。草稿失败后先检查草稿状态再重试。已发布版本的修复使用更高版本重新发行。
4. 从 0.1.0 升级时，从首个正式 Release 下载 NSIS 安装包并覆盖安装一次。之后在 **设置 → 检查更新** 查看说明和下载进度，可稍后返回设置安装。下载暂存在本次应用进程内，退出后需重新下载。安装会关闭并重启桌面客户端，商品同步结束后才开放安装。

手动构建可启用 `signed_smoke`。该构建用一次性密钥签名，Updater 仅指向本机测试端点，测试产物保留一天。

正式验收使用两个递增且由同一受管密钥签名的版本。在 Windows x64 上覆盖安装第一个版本，确认现有登录可恢复、店铺可见和本地 API 配置可读。用第一个版本在应用内检查、下载第二个版本，先选择稍后安装，返回后安装并确认重启、版本和数据保留。另检查当前已是最新版、网络异常和签名错误时的反馈。PR 和手动 CI 产物用于原生构建检查，正式在线更新由负责人批准发行后开启。

## 数据与安全边界

- 桌面端通过 Supabase Auth 官方接口登录/续签，仅通过 Go API 查询采集商品；不直连 PostgreSQL 或 Supabase Data API，桌面端不含数据库凭据。客户端的 Supabase publishable key 本身可公开，**数据库 URL、服务端密钥、Service Role Key 不得打包**。
- 商品图片仍从原始 HTTPS URL 加载。Tauri CSP 的 `img-src https:` 允许来源图片展示；`connect-src` 除 Tauri IPC、本地 8787 API 和开发用 Vite 连接外，**仅精确放行当前 Supabase Auth HTTPS 域名**，不允许任意远端 HTTPS API。更换项目或新增远端服务时同步调整 CSP。
- 本地 Go 服务仍只监听 `127.0.0.1:8787`；采集接口通过 Supabase Auth User API 验证 Bearer Token，后端使用验证后的用户 UUID 过滤商品。**公网部署仍须独立完成 HTTPS、域名、CSP/CORS、代理配置；此 PR 不直接公开 8787**。
- 未登录时只显示登录页，未配置 Go API / 数据库 / Auth 服务时显示明确错误；未登录、无效或过期 Token 不可读写商品。退出登录清除本地会话；已发出的短期 access token 仍按 Supabase Auth 的有效期自然失效。
- `npm run build` 仅验证 TypeScript/Vite；GitHub Actions 的 Windows NSIS 构建另行验证原生可编译，不代替 GUI 真机检查。
- **（PR #13 无登录版本）2026-11-17 Windows x64 已实测 PASS**：Tauri 安装和启动、本机 Go API 连接 Supabase PostgreSQL、合成商品成功入库并在采集箱展示、重启桌面及 Go API 后仍能读取。它不代表浏览器插件已完成真实商品采集。
- **PR #15 登录版本：Windows 真实邮箱登录、关闭重开自动恢复、退出后再次登录实机 `PASS`**（2026-11-17）；Go 自动化覆盖跨账号隔离 `PASS`，真实双账号/商品展示和过期续签仍 `NOT_RUN`。本地 Go API 访问 Supabase Auth 在当前网络下需要代理，云端部署尚未验收。
- 本版界面保持浅色模式，已优化小字对比度、布局与按钮反馈；自动深色模式后置，不把未实测的深色主题纳入登录验收。
- **macOS arm64 / x64 均为 `UNVERIFIED`**，不宣称跨平台已验收。

## 云端部署（M0-02B）

云端部署步骤与验收见 [云端 Go API 部署操作说明](../../docs/architecture/CLOUD_GO_API_DEPLOY.md)。**本 PR 仅准备云端运行代码和配置，未核实服务器真实部署。** 旧版 Windows Artifact 仍默认连接本地；新版正式包默认使用云域名，仅在 HTTPS、Go 服务和数据库连通后使用。客户端仍直接连接 Supabase Auth，客户端网络无法访问该域名时仍可能需要本机代理。

## 下一阶段

Issue #6 / PR-C 才添加 Chrome/Edge MV3 插件，用户点击采集后调用已鉴权的 POST API。当前桌面端不实现 POST 表单、产品编辑、AI、Mercado Libre OAuth 或发布功能。
