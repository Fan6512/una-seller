# Una Seller 桌面端：登录与采集箱

桌面端使用 **Supabase Auth** 登录，通过 Go API 读取 **PostgreSQL 中当前登录用户的来源商品**；不包含假数据、页面爬虫或 Mercado Libre 店铺授权。

## 本地运行

1. 确认 Go 1.26、Node.js 22+、Tauri 2 所需 Rust 与系统依赖已安装。
2. Windows 首次在 `%LOCALAPPDATA%\\UnaSeller\\api-config.json` 保存 `DATABASE_URL`、`SUPABASE_URL`、`SUPABASE_PUBLISHABLE_KEY`（参考[后端说明](../../backend/README.md)）；以后重新下载 Go API 不再重填。执行 `go run ./cmd/api migrate`（仅未迁移时），然后 `go run ./cmd/api`。
3. 在 Supabase Dashboard 的 **Authentication → Users** 创建/邀请一个允许登录的邮箱账号，确认邮箱和密码可用。桌面端内置的 `VITE_SUPABASE_URL` 和 `VITE_SUPABASE_PUBLISHABLE_KEY` 默认指向本仓库的 Supabase 项目，切换项目时必须同时修改构建环境与 CSP。
4. 在 `apps/desktop/` 执行 `npm ci`、`npm run tauri:dev`（仅源码开发时需要 Rust/Tauri 工具链；使用 CI 产物时不需要）；用邮箱密码登录后进入采集箱。
5. 桌面采集箱请求 `GET /api/v1/collections` 和详情 `GET /api/v1/collections/{id}`；列表如果为空会明确显示“这里还没有商品”，**不会伪造演示记录**。
6. 当前仅支持本地 API `http://127.0.0.1:8787`（或 `http://localhost:8787`）。可在 `apps/desktop/.env.local` 设置 `VITE_API_BASE_URL=http://127.0.0.1:8787`，但仅修改环境变量**不能**切换到其他端口或远端 HTTPS 地址：Tauri `connect-src` 未放行这些来源。以后确需更换地址时，必须同步精确修改 CSP，并核对 Go API 的 CORS。不要将数据库 URL、密码、服务端 owner ID 或密钥放进 Vite 环境变量。

## Issue #14 登录说明

- Auth 后端使用已有 Supabase 项目的开源 **GoTrue / Supabase Auth**；桌面通过官方 HTTP API 实现邮箱密码登录、自动续签、恢复会话和退出登录，**不额外部署认证服务、也不创建密码表**。
- 本地测试账号已由负责人创建；当前先不阻断公开注册配置。注意 **Supabase Auth 是公网服务**，只把 Go API 放在本机并不阻止外部调用注册接口；正式分发/公网开放前须关闭注册。
- 登录会话由桌面 WebView 的本地存储保留，不包含用户密码；客户端发起业务请求时自动携带 Bearer Token。Go 服务端通过 `GET /auth/v1/user` 验证令牌并从服务端响应读取真实用户 ID，不接受客户端提交 owner UUID。
- 若已安装的是 PR #13 的老版本，老版本尚无登录页面；请下载 Issue #14 对应的新版 Windows Artifact。旧本地 API + 旧客户端组合不能作为新 Auth 版本的验收证据。
- 未对旧 UUID 对应的测试商品做迁移，避免把其他用户记录误归属到新账号。当前不开放自助注册、忘记密码入口、Google 登录或团队权限。

## Windows 构建产物验收（无需本机 Go/Rust 工具链）

从 **[本 PR #15 的 Windows Desktop Build（成功运行 #44009983090）](https://github.com/Fan6512/una-seller/actions/runs/44009983090)** 下载 `una-seller-windows-x64` Artifact，解压后使用。**不要下载 PR #13 的旧版产物，旧版没有登录功能。**

- `una-seller.exe`：Tauri Windows 桌面客户端。
- `una-seller-api.exe`：独立的本地 Go API 程序；**它只在 Artifact 中，不包含在 NSIS `*-setup.exe` 安装包内**。
- `Una Seller_0.1.0_x64-setup.exe`：仅安装桌面客户端，**安装后不会自动启动或附带 Go API**。单独安装后看到“连接异常”不代表桌面程序安装失败。

测试真实商品读取时，解压 Artifact 后，在首次创建的 `%LOCALAPPDATA%\\UnaSeller\\api-config.json` 中保存配置。在 `una-seller-api.exe` 所在目录的 PowerShell 中执行 `.\\una-seller-api.exe`，保持 API 运行，再打开 `una-seller.exe` 或已安装的客户端。**此后换新版 EXE 直接运行，不再逐次配置**；Go API 继续监听本机 `127.0.0.1:8787`。

- 不要将数据库连接字符串/密码贴到 issue、PR、截图、公开 Actions 日志，也不要写入客户端环境变量；仅本机 Go API 进程需要数据库凭据。
- **新 Supabase Auth 用户 ID 与旧本地 UUID 不同。** 以前手动创建的测试商品仍在数据库原用户下面，登录新账号后列表为空属于预期；未经授权不自动迁移或改写商品所属用户。
- Windows 原生构建通过与 GUI 启动通过是两件事；macOS 状态另计。

## 数据与安全边界

- 桌面端通过 Supabase Auth 官方接口登录/续签，仅通过 Go API 查询采集商品；不直连 PostgreSQL 或 Supabase Data API，桌面端不含数据库凭据。客户端的 Supabase publishable key 本身可公开，**数据库 URL、服务端密钥、Service Role Key 不得打包**。
- 商品图片仍从原始 HTTPS URL 加载。Tauri CSP 的 `img-src https:` 允许来源图片展示；`connect-src` 除 Tauri IPC、本地 8787 API 和开发用 Vite 连接外，**仅精确放行当前 Supabase Auth HTTPS 域名**，不允许任意远端 HTTPS API。更换项目或新增远端服务时同步调整 CSP。
- 本地 Go 服务仍只监听 `127.0.0.1:8787`；采集接口通过 Supabase Auth User API 验证 Bearer Token，后端使用验证后的用户 UUID 过滤商品。**公网部署仍须独立完成 HTTPS、域名、CSP/CORS、代理配置；此 PR 不直接公开 8787**。
- 未登录时只显示登录页，未配置 Go API / 数据库 / Auth 服务时显示明确错误；未登录、无效或过期 Token 不可读写商品。退出登录清除本地会话；已发出的短期 access token 仍按 Supabase Auth 的有效期自然失效。
- `npm run build` 仅验证 TypeScript/Vite；GitHub Actions 的 Windows NSIS 构建另行验证原生可编译，不代替 GUI 真机检查。
- **（PR #13 无登录版本）2026-11-17 Windows x64 已实测 PASS**：Tauri 安装和启动、本机 Go API 连接 Supabase PostgreSQL、合成商品成功入库并在采集箱展示、重启桌面及 Go API 后仍能读取。它不代表浏览器插件已完成真实商品采集。
- **PR #15 登录版本：Windows 实机登录/会话恢复/跨账号隔离仍为 `UNVERIFIED`**，不能沿用 PR #13 的无登录验收结论。
- 本版界面保持浅色模式，已优化小字对比度、布局与按钮反馈；自动深色模式后置，不把未实测的深色主题纳入登录验收。
- **macOS arm64 / x64 均为 `UNVERIFIED`**，不宣称跨平台已验收。

## 下一阶段

Issue #6 / PR-C 才添加 Chrome/Edge MV3 插件，用户点击采集后调用已鉴权的 POST API。当前桌面端不实现 POST 表单、产品编辑、AI、Mercado Libre OAuth 或发布功能。
