# PR-B：桌面采集箱

桌面采集箱通过 Go API 读取 **已经保存到 PostgreSQL 的来源商品**，不包含假数据、页面爬虫或店铺登录。

## 本地运行

1. 确认 Go 1.26、Node.js 22+、Tauri 2 所需 Rust 与系统依赖已安装。
2. 在 `backend/` 配置自己的 `DATABASE_URL` 和持久的 `UNA_SELLER_LOCAL_USER_ID`；执行 `go run ./cmd/api migrate`（仅数据库尚未迁移时），然后 `go run ./cmd/api`。
3. 在 `apps/desktop/` 执行 `npm ci`、`npm run tauri:dev`。
4. 桌面采集箱请求 `GET /api/v1/collections` 和详情 `GET /api/v1/collections/{id}`；列表如果为空会明确显示“还没有采集商品”，**不会伪造演示记录**。
5. 当前仅支持本地 API `http://127.0.0.1:8787`（或 `http://localhost:8787`）。可在 `apps/desktop/.env.local` 设置 `VITE_API_BASE_URL=http://127.0.0.1:8787`，但仅修改环境变量**不能**切换到其他端口或远端 HTTPS 地址：Tauri `connect-src` 未放行这些来源。以后确需更换地址时，必须同步精确修改 CSP，并核对 Go API 的 CORS。不要将数据库 URL、密码、服务端 owner ID 或密钥放进 Vite 环境变量。

## 数据与安全边界

- 本次仅**读取** Go API；不连接 Supabase Data API，桌面端不含数据库凭据。
- 商品图片仍从原始 HTTPS URL 加载。Tauri CSP 的 `img-src https:` 允许来源图片展示；`connect-src` 只允许 Tauri IPC、本地 8787 API 和开发模式所需的 Vite 连接，**不允许任意远端 HTTPS API**。后续确需远端 API 时，应按具体域名精确放行。
- 本地 Go 开发服务只监听 `127.0.0.1:8787`，固定本地用户不是公网身份认证，**禁止把采集 API 直接通过 Nginx 公网暴露**。
- 当 Go API 尚未部署或连接数据库未配置时，桌面会显示明确错误。
- `npm run build` 仅验证 TypeScript/Vite。没有 Windows 原生启动、macOS arm64/x64 打包和 GUI 实机记录前，必须标记 `UNVERIFIED`，不能声称双平台可用。
- 现有 Supabase 项目已经执行 PR-A Schema 迁移；**尚未验证 Windows 上 Go API 连云库以及桌面 Tauri 真机启动**。

## 下一阶段

Issue #6 / PR-C 才添加 Chrome/Edge MV3 插件，用户点击采集后调用 POST API。本 PR 不实现 POST 表单、产品编辑、AI、OAuth 或发布功能。
