# ADR-0005：M0 初期基础设施与 HTTPS 证书选型

- 日期：2026-10-26
- 状态：**Accepted（技术选型）；NOT_DEPLOYED（部署未完成）**
- 阶段对齐：2026-11-16，按照已合并的 Foundation v2.1 / Delivery v2.1 明确 M0-03 采集箱、M1-D OAuth；不更改当时 SSH 记录
- 依据：用户明确选择现有 Ubuntu 2C4G 服务器、Cloudflare 子域名、Supabase 托管 PostgreSQL，并希望自行申请 HTTPS 证书；服务器诊断和 Nginx 状态由用户 SSH 输出确认。
- 优先级：服从 [ADR-0004](0004-simplicity-first.md) 的简单优先原则；此 ADR 对旧的 Go/TypeScript 待定和本地部署 PostgreSQL 默认方案构成更新。

## 决策

1. **桌面**：Tauri 2 + React + TypeScript，Windows/macOS；客户端通过 HTTPS 请求业务 API，**不直连 Supabase 数据库**。
2. **后端**：一个 Go API 服务，运行于现有 **Ubuntu 24.04.4 LTS / 2 vCPU / 4GB**。运行时仅监听回环接口，由 Nginx 转发；不增设第二个 Web 网关。
3. **公网边界**：Cloudflare 管理 `mkd.186512.xyz` DNS/代理，**复用现有 Nginx**（当前占用 80），为该主机新增独立 TLS 虚拟主机监听 443。正式 TLS 优先使用用户自行通过 Let's Encrypt / Certbot 申请、自动续期的证书；Cloudflare 使用 **Full (strict)**。Cloudflare Origin CA 是备选。
4. **数据库**：前期 **Supabase 托管 PostgreSQL**，在 **M0-03 / Issue #6 的 PR-A 首次需要真实采集数据持久化时**按需接入，M1-D 才增加 OAuth/店铺授权数据；只建立当期确需的数据表与迁移。Go 服务使用 PostgreSQL 协议，不预装额外 Supabase 服务。IPv4-only 服务器可使用官方 Session Pooler（5432），连接信息取自实际 Supabase 项目并启用 TLS 验证；凭据仅在 Go 服务端。
5. **应用回调**：计划使用 `https://mkd.186512.xyz/oauth/mercadolibre/callback`，**尚未核实是否已在 Mercado Libre 开发者后台登记**，也未部署回调接口；实施 OAuth 前需确认并匹配真实登记 URI。Mercado Libre 开发者应用已创建，Client ID 已取得；不把 Client Secret、Token、证书私钥写入仓库或终端日志。
6. **保护现有服务**：同机 Nginx 和 Docker 容器 `we-mp-rss` 保持正常；任何配置变更先备份、执行 `nginx -t` 后才 reload。仅在本子域名处理 TLS/重定向，不擅自改全站加密模式。
7. **不做**：Redis、Kafka、集群、Kubernetes、收费系统、业务多租户、Supabase Auth/Realtime/Edge Functions 的预设集成。

## 理由及可逆性

M0 主要是有限店铺、低并发调用，现有 2C4G 足够作为起点；托管 PostgreSQL 减少服务器数据库维护。Nginx 已存在且正在使用，重复安装 Caddy 反而复杂。Go 标准库已具备最小健康端点，用户明确选择继续沿用 Go；并未作 Go 与 Node 的正式压测对比。PostgreSQL 是可迁移的标准接口，未来扩容或自建可迁移而无需重做业务层。

## 与现行开发阶段的对齐（2026-11-16）

本 ADR **只决定运行基础设施**，不重定产品计划：以现行 [Delivery v2.1](../architecture/DELIVERY.md) 和 [Issue #6](https://github.com/Fan6512/una-seller/issues/6) 为开发顺序。**M0-03 是采集箱 Go API/桌面/MV3 闭环**；涉及真实授权的 Mercado Libre OAuth、店铺连接与只读同步属于 **M1-D**。证书与独立域名部署属于 M0-02A/B 的候选工作，不应成为本地/CI 采集 API 开发的隐性阻塞。

用户此前 SSH 观察到现有 Ubuntu、Nginx 和 `we-mp-rss`，不表示本轮重新 SSH 实测，也不表示 HTTPS、Supabase 或 OAuth 已验收；上线前必须单独核验。保留原有 Foundation v2.1、Delivery v2.1、ADR-0001 和 ADR-0002，不用旧版本覆盖产品三区/AI/插件规划。

## 实施边界与状态

**已知事实**：用户报告 Cloudflare DNS 已配置；服务器实测 Ubuntu 24.04.4、2 核、约 3.6GiB 内存、约 37GB 可用磁盘、Nginx 在 80 监听且 `nginx -t` PASS、443 当时未监听；存在 Docker `we-mp-rss`。

**尚未完成**：DNS 公网解析复核、Let's Encrypt 证书签发/自动续期、Nginx 443、安全组、Go API 公网部署、Supabase 项目创建/连接、真实 OAuth 与 Token 管理。不能把技术选型视为上线通过。

实施手册、验收点及官方文档链接见 [M0 部署基线](../architecture/DEPLOYMENT_M0.md)。
