# Una Seller — Mercado Libre AI ERP

> **当前阶段：最简自用 ERP 的开发准备。** 只做 Mercado Libre 运营；Windows/macOS 双平台桌面客户端；Amazon 仅作为商品研究和采集来源；暂不建设支付、订阅、计费及商业化 SaaS 后台。

## 先读这两份

- [开发原则与技术地基（Foundation v2）](docs/architecture/FOUNDATION.md)：第一性原理、真实问题优先、简单可用、安全不过度。
- [精简开发路线（Delivery v2）](docs/architecture/DELIVERY.md)：M0 能跑 → M1 商品可用 → M2 AI 有用 → M3 按需运营；M4 商业化冻结。

## 当前 MVP 开发任务书

- [ERP 后端与 Windows/macOS 桌面客户端 MVP](docs/architecture/ERP-MVP.md)
- [浏览器商品采集插件 MVP](docs/product/BROWSER-COLLECTION-PLUGIN-MVP.md)

## 其他正式文档

- [Windows/macOS 双平台规范](docs/architecture/CROSS_PLATFORM.md)
- [最小领域与写操作契约](docs/architecture/DOMAIN_AND_EXECUTION.md)
- [架构决策 ADR](docs/decisions/README.md)
- [全部文档索引](docs/README.md)

## 研究资料（与产品代码隔离）

- [智赢静态研究](research/zying/README.md)
- [原合规/路线图 v3（历史）](docs/product/AI-ERP-合规接入清单与模块路线图.md)

**状态说明**：当前仓库的正式内容主要是架构规范、M0 最小工程和研究文件。浏览器采集插件与 ERP MVP 任务书现已归档为待实施需求基线，不代表功能已经实现。原则是先实现能启动、能连接、能把真实商品采集到本地库的最小闭环，再逐项增加业务功能，不提前铺设复杂系统。

## M0-01 工程（已开始，尚未在 Win/mac 实机验收）

[运行步骤](docs/architecture/M0-01-IMPLEMENTATION.md) · [验证状态](docs/architecture/M0-01-STATUS.md)

目前 `apps/desktop/` 是 Tauri + React 最小界面，`backend/` 是**仅本机运行**的 Go 健康检查 HTTP 服务。未提供公网 ERP API、OAuth、数据库或收费功能。暂未配置 `mkd.186512.xyz`；KYC 未完成前不做真实授权。

Windows PowerShell 示例（在仓库根目录打开两个终端）：

```powershell
# 终端 A
cd backend
go test ./...
go run ./cmd/api
```

```powershell
# 终端 B
cd apps/desktop
npm install
npm run tauri:dev
```

首次运行会生成依赖锁文件；应在实际构建验证后提交。当前 PR 未新增自动 GitHub Actions，避免无谓消耗额度。
