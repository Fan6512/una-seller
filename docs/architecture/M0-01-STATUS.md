# M0-01 当前验收记录

> 2026-10-26。此记录刻意区分 **代码已提交**、**本地能实际验证** 与 **待用户环境验证**。

| 验证项 | 状态 | 证据 / 后续动作 |
| --- | --- | --- |
| Tauri 2 + React + Vite 最小骨架 | CODE_READY | 源码及 npm 构建脚本已创建，npm 包尚未在本环境安装 |
| Go 1.23 本机健康接口 | PASS (isolated Go test) | 本地使用相同 Go 源码运行 `go test ./...` 与 `go vet ./...` 通过 |
| Go 格式检查 | PASS | 使用 gofmt 格式化后写入仓库 |
| 服务端真实进程 / HTTP curl | TODO | Windows/macOS 本机执行 `go run ./cmd/api`，再 GET /healthz |
| 桌面 TypeScript 构建 | UNVERIFIED | 当前工具环境 npm registry 不可达，未能 `npm install` |
| Windows Tauri 原生窗口 | UNVERIFIED | 需要用户 Win11 的 Rust/MSVC/WebView2 环境 |
| macOS arm64/x64 Tauri | UNVERIFIED | 需要 macOS runner/实机；不可声称已通过 |
| package-lock.json / Cargo.lock | PENDING | 首次在线 npm install / Rust Cargo 构建后生成并在后续提交固定 |
| Mercado Libre OAuth | BLOCKED | 卖家主账号未完成 KYC，没有开发者应用凭据 |
| mkd.186512.xyz HTTPS | NOT_CONFIGURED | 子域名尚未创建 DNS、证书或部署；本 PR 不修改 |

**进入下一阶段的门槛**：先完成至少 Windows 本机真实启动与 HTTP 联通检查，锁定依赖文件；再决定后端 Go 或 TypeScript。不要把未测试的原生构建当成 PASS。
