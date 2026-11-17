# PR-B：桌面采集箱

桌面采集箱通过 Go API 读取 **已经保存到 PostgreSQL 的来源商品**，不包含假数据、页面爬虫或店铺登录。

## 本地运行

1. 确认 Go 1.26、Node.js 22+、Tauri 2 所需 Rust 与系统依赖已安装。
2. 在 `backend/` 配置自己的 `DATABASE_URL` 和持久的 `UNA_SELLER_LOCAL_USER_ID`；执行 `go run ./cmd/api migrate`（仅数据库尚未迁移时），然后 `go run ./cmd/api`。
3. 在 `apps/desktop/` 执行 `npm ci`、`npm run tauri:dev`（仅源码开发时需要 Rust/Tauri 工具链；使用 CI 产物时不需要）。
4. 桌面采集箱请求 `GET /api/v1/collections` 和详情 `GET /api/v1/collections/{id}`；列表如果为空会明确显示“还没有采集商品”，**不会伪造演示记录**。
5. 当前仅支持本地 API `http://127.0.0.1:8787`（或 `http://localhost:8787`）。可在 `apps/desktop/.env.local` 设置 `VITE_API_BASE_URL=http://127.0.0.1:8787`，但仅修改环境变量**不能**切换到其他端口或远端 HTTPS 地址：Tauri `connect-src` 未放行这些来源。以后确需更换地址时，必须同步精确修改 CSP，并核对 Go API 的 CORS。不要将数据库 URL、密码、服务端 owner ID 或密钥放进 Vite 环境变量。

## Windows 构建产物验收（无需本机 Go/Rust 工具链）

从 [PR #13 的 Windows Desktop Build](https://github.com/Fan6512/una-seller/actions/workflows/windows-desktop.yml) 下载 `una-seller-windows-x64` Artifact，解压后使用：

- `una-seller.exe`：Tauri Windows 桌面客户端。
- `una-seller-api.exe`：独立的本地 Go API 程序；**它只在 Artifact 中，不包含在 NSIS `*-setup.exe` 安装包内**。
- `Una Seller_0.1.0_x64-setup.exe`：仅安装桌面客户端，**安装后不会自动启动或附带 Go API**。单独安装后看到“连接异常”不代表桌面程序安装失败。

测试真实商品读取时，推荐解压后使用**独立二进制组合**，而不是只装 NSIS：在 `una-seller-api.exe` 所在目录打开 PowerShell，设置当前会话的 `DATABASE_URL`（从 Supabase Connect 获取的 PostgreSQL 连接字符串）和固定 `UNA_SELLER_LOCAL_USER_ID`（UUID，不是 Supabase 项目 ID），执行 `.\una-seller-api.exe`。保持此窗口运行，再打开 `una-seller.exe` 或已安装的桌面客户端；它们通过本机 `127.0.0.1:8787` 通信。

- 不要将数据库连接字符串/密码贴到 issue、PR、截图、公开 Actions 日志，也不要写入客户端环境变量；仅本机 Go API 进程需要数据库凭据。
- 连接数据库但没有该 UUID 对应的商品时，采集箱显示空列表属于正常状态；要验证列表与详情，应使用已经持久化的商品。
- Windows 原生构建通过与 GUI 启动通过是两件事；macOS 状态另计。

## 数据与安全边界

- 本次仅**读取** Go API；不连接 Supabase Data API，桌面端不含数据库凭据。
- 商品图片仍从原始 HTTPS URL 加载。Tauri CSP 的 `img-src https:` 允许来源图片展示；`connect-src` 只允许 Tauri IPC、本地 8787 API 和开发模式所需的 Vite 连接，**不允许任意远端 HTTPS API**。后续确需远端 API 时，应按具体域名精确放行。
- 本地 Go 开发服务只监听 `127.0.0.1:8787`，固定本地用户不是公网身份认证，**禁止把采集 API 直接通过 Nginx 公网暴露**。
- 当 Go API 尚未部署或连接数据库未配置时，桌面会显示明确错误。
- `npm run build` 仅验证 TypeScript/Vite；GitHub Actions 的 Windows NSIS 构建另行验证原生可编译，不代替 GUI 真机检查。
- **2026-11-17 Windows x64 已实测 PASS**：Tauri 安装和启动、本机 Go API 连接 Supabase PostgreSQL、合成商品成功入库并在采集箱展示、重启桌面及 Go API 后仍能读取。它不代表浏览器插件已完成真实商品采集。
- **macOS arm64 / x64 均为 `UNVERIFIED`**，不宣称跨平台已验收。

## 下一阶段

Issue #6 / PR-C 才添加 Chrome/Edge MV3 插件，用户点击采集后调用 POST API。本 PR 不实现 POST 表单、产品编辑、AI、OAuth 或发布功能。
