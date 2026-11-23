# Una Seller — Mercado Libre AI ERP

> **当前阶段：最简自用 ERP 的开发准备。** 只做 Mercado Libre 运营；Windows/macOS 双平台桌面客户端；Amazon 仅作为商品研究和采集来源；暂不建设支付、订阅、计费及商业化 SaaS 后台。

## 先读这两份

- [开发原则与技术地基（Foundation v2.1）](docs/architecture/FOUNDATION.md)：第一性原理、真实问题优先、简单可用、安全不过度。
- [精简开发路线（Delivery v2.1）](docs/architecture/DELIVERY.md)：先基础 CI → M0 采集箱 → M1 草稿/AI/刊登 → M2 商品维护 → M3 按需运营；M4 商业化冻结。

## 最新功能进度

**[功能状态总台账：已完成 / 进行中 / 待完成 + 对应 PR](docs/product/FEATURE-STATUS.md)**。这是日常进度和新需求登记入口；以下早期进度文字仅保留历史上下文，不作为当前验收结论。

## 产品目标及开发路线

**采集箱 → AI 整理 → 产品库 → Mercado Libre 多站点刊登 → 已发布产品列表 → 日常运营**。桌面端保持一个统一主题/主程序，内置业务插件按需开发，仍只用一个后端与数据库。

- [商品工作流、智赢编辑器参考和插件规划](docs/product/PRODUCT-WORKFLOW-AND-EDITOR.md)
- [Delivery v2.1 分阶段开发路线](docs/architecture/DELIVERY.md)

## 当前 MVP 开发任务书

- [ERP 后端与 Windows/macOS 桌面客户端 MVP](docs/architecture/ERP-MVP.md)
- [浏览器商品采集插件 MVP](docs/product/BROWSER-COLLECTION-PLUGIN-MVP.md)

## 其他正式文档

- [CI 与初期运行环境](docs/architecture/CI_AND_RUNTIME.md)：公开仓库的 Actions 分阶段启用、现有 Ubuntu 服务器 + Supabase 托管库、本机工具链分工。
- [M0 基础设施与部署验收](docs/architecture/DEPLOYMENT_M0.md)：已有 Ubuntu/Nginx + Supabase 选型、HTTPS 步骤与尚未验收清单。
- [Windows/macOS 双平台规范](docs/architecture/CROSS_PLATFORM.md)
- [最小领域与写操作契约](docs/architecture/DOMAIN_AND_EXECUTION.md)
- [架构决策 ADR](docs/decisions/README.md)
- [全部文档索引](docs/README.md)

## 研究资料（与产品代码隔离）

- [智赢静态研究](research/zying/README.md)

**当前交付状态（2026-11-16）**：M0-00 基础 CI 与 Go 1.26 已合并；Issue #6 / PR-A 的采集 POST/GET、去重、图片、变体和 PostgreSQL 17 集成测试已合并。Supabase `una-seller` 已执行四张业务表迁移，并完成可回滚的 SQL 合成数据验证；Go API 连接云库、Windows/macOS Tauri 实机与浏览器插件尚未验收。合并历史请查看 GitHub PR #9、#10、#11；后续任务以 Issue #6 为准。

## M0-01 工程（基础工程；尚未在 Win/mac 实机验收）

[运行步骤](docs/architecture/M0-01-IMPLEMENTATION.md) · [验证状态](docs/architecture/M0-01-STATUS.md)

目前 `backend/` 具有本地 Go 健康接口和 PostgreSQL 采集 POST/GET；仅供本地开发，不具备公网认证，不得将采集 API 对外发布。桌面端尚无已验收的真实采集箱 GUI；新的 [PR-B 桌面采集箱](https://github.com/Fan6512/una-seller/pull/13) 正在独立开发。暂未验收 `mkd.186512.xyz` 的 DNS/HTTPS，OAuth 和收费未开发。

Windows PowerShell 示例（在仓库根目录打开两个终端）：

```powershell
# 终端 A
cd backend
go test ./...
# 先在本地单独设置 DATABASE_URL 与 UNA_SELLER_LOCAL_USER_ID
# 全新数据库才运行 go run ./cmd/api migrate；现有 Supabase 已做迁移
go run ./cmd/api
```

```powershell
# 终端 B
cd apps/desktop
npm install
npm run tauri:dev
```

**M0-00 基础 CI 与 PR-A PostgreSQL 集成测试已建立并在真实 PR 上通过。** 桌面原生构建、Windows/macOS 真机验收和 MV3 插件仍属后续任务；不得把 Vite 编译当作桌面端验收。
