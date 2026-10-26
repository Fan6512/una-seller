# M0-01 最小桌面与健康接口（开发分支验收）

> 阶段：M0-01。仅构建本地工程，**不是已能登录或连接真实美客多的 ERP**。KYC 未完成、无 App ID，`mkd.186512.xyz` DNS/HTTPS 尚未配置。

## 做了什么

- `apps/desktop/`：Tauri 2 + React/TypeScript + Vite 基础窗口；UI 显示本机 API 实际在线/离线状态并可重试；没有假数据模拟已连接。
- `backend/`：Go 1.23 标准库 HTTP，仅监听 `127.0.0.1:8787`，仅 GET `/healthz`；没有公网服务和写接口。
- 没有安装数据库、Redis、消息队列、身份/计费服务。没有修改 `mkd.186512.xyz` 的 DNS 或服务器。
- 不创建 Actions workflow：目前先靠本地命令验证，避免消耗 CI 额度。

## Windows 11 开发命令（需预先安装 Node.js、Go、Rust、MSVC C++ Build Tools、WebView2）

第一终端：

```powershell
cd backend
go test ./...
go run ./cmd/api
```

第二终端：

```powershell
cd apps/desktop
npm install
npm run tauri:dev
```

首次执行 `npm install` 会产生 `package-lock.json`；首次 Tauri 构建会产生 `src-tauri/Cargo.lock`。应在环境验证完成后将锁文件单独提交以确保复现。运行 `npm run tauri:dev` 会自动用 `app-icon.svg` 生成各平台图标文件。

只看前端也可执行 `npm run dev`，随后在浏览器打开 `http://127.0.0.1:1420`。

## macOS 验证

预装 Xcode Command Line Tools、Rust、Node.js、Go 后执行同样的后端与桌面命令：

```bash
(cd backend && go test ./... && go run ./cmd/api) &
(cd apps/desktop && npm install && npm run tauri:dev)
```

使用 macOS arm64 与 Intel 构建时分别执行 `npm run tauri:build`，确认 `.app/.dmg`、授权权限与 WebView 行为。**没有 Mac 实机或 macOS Runner 验证记录，不得标记跨平台 PASS。**

## 验收

1. `go test ./...` PASS；`curl http://127.0.0.1:8787/healthz` 返回 `{"status":"ok","service":"una-seller-api"}`。
2. 前端成功获取 API 后显示“连接正常”；停止后端再点击“重新检测”应显示“尚未连接”。
3. Windows 的 `npm run tauri:dev` 真正启动窗口；macOS 的构建和启动单独验证；不能用普通浏览器截图代替 Tauri 实机。
4. `npm run build` TS 检查通过，依赖锁文件生成并在后续 PR 固定。
5. KYC 与 OAuth 真实联调保持 **BLOCKED**，不要谎称已完成。

## 当前技术选择

Go 标准库做本机服务，是验证最短闭环的可执行候选，**不表示已做完 Go vs TypeScript 后端的选型**。正式后端选型将在 M0-02 前用维护成本与验证结果决定。该本机服务不能直接作为公网服务部署，正式部署需要 HTTPS、ERP 身份认证和数据库。

本阶段故意不加入 OpenAPI、PostgreSQL、任意 AI 框架、复杂 UI Kit；等有实际业务需要再加。
