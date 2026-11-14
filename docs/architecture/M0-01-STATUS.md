# M0-01 当前验收记录

> 2026-10-26 初版，2026-11-15 复核更新。此记录刻意区分 **代码已提交**、**本地能实际验证** 与 **待用户环境验证**。

| 验证项 | 状态 | 证据 / 后续动作 |
| --- | --- | --- |
| Tauri 2 + React + Vite 最小骨架 | CODE_READY | 源码与构建脚本已在仓库；前端生产构建通过，Tauri 原生壳尚未启动 |
| Go 1.23 本机健康接口 | PASS (isolated Go test) | 本地使用相同 Go 源码运行 `go test ./...` 与 `go vet ./...` 通过 |
| Go 格式检查 | PASS | 使用 gofmt 格式化后写入仓库 |
| 服务端真实进程 / HTTP curl | TODO | Windows/macOS 本机执行 `go run ./cmd/api`，再 GET /healthz |
| 桌面 TypeScript 构建 | PASS (frontend build) | 2026-11-15 Windows 本机 `npm install`（71 包）+ `npm run build`（`tsc --noEmit` + `vite build`）通过；只覆盖 Web 前端，不含原生壳 |
| Windows Tauri 原生窗口 | UNVERIFIED | 需要用户 Win11 的 Rust/MSVC/WebView2 环境 |
| macOS arm64/x64 Tauri | UNVERIFIED | 需要 macOS runner/实机；不可声称已通过 |
| package-lock.json / Cargo.lock | 半完成 | `package-lock.json` 已于 2026-11-15 随 main 提交；`Cargo.lock` 仍需首次 Rust 构建生成 |
| Mercado Libre OAuth | BLOCKED | 卖家主账号未完成 KYC，没有开发者应用凭据 |
| mkd.186512.xyz HTTPS | NOT_CONFIGURED | 子域名尚未创建 DNS、证书或部署；本 PR 不修改 |

**进入下一阶段的门槛**：先完成至少 Windows 本机真实启动与 HTTP 联通检查，锁定依赖文件；再决定后端 Go 或 TypeScript。不要把未测试的原生构建当成 PASS。
