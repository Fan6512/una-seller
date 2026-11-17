# Una Seller M0 基础设施、数据库与 HTTPS 部署基线

> 决策日期：2026-10-26 · 与现行计划对齐：2026-11-16 · 状态：**选型确定，部署未验收** · 对应 [ADR-0005](../decisions/0005-m0-infrastructure.md)。
> 自用初期沿用 **一台现有 Ubuntu 服务器 + 一个 Go API + 一个 Supabase 托管 PostgreSQL 数据库**，Cloudflare/Nginx 使用现有资源。**这不是执行完成报告，也不等于本轮重新 SSH 验证**；业务阶段以 [Delivery v2.1](DELIVERY.md) 为准。

## 1. 最终选型与职责

| 层次 | 选择 | 职责与边界 |
| --- | --- | --- |
| 桌面客户端 | Tauri 2 + React/TypeScript（Win/macOS） | 仅请求经身份认证的 HTTPS API；不含 Mercado Libre Client Secret、Refresh Token 或数据库密码 |
| 域名/DNS/代理 | Cloudflare；`mkd.186512.xyz` | 承接外部 HTTPS 访问；DNS 已由用户配置，解析/代理状态需实测 |
| 反向代理/源站 TLS | 复用现有 Nginx | 为 `mkd.186512.xyz` 新建独立 server block，不覆盖其他站点，443 → Go loopback；不安装 Caddy |
| 源站证书 | **Let's Encrypt + Certbot 优先** | 用户可自行申请；有效证书及自动续期。Cloudflare Origin CA 为备选 |
| 业务后端 | Go 单体 API，Ubuntu 24.04.4 LTS，2 vCPU / 4GB | OAuth state/回调、令牌交换与刷新、官方 API 请求、数据写入 |
| 数据库 | **Supabase 托管 PostgreSQL** | M0-03 / Issue #6 的采集商品需要持久化时先建最小采集表；M1-D 才加 OAuth/店铺连接表，后续按需扩展 |
| 现有业务 | Docker `we-mp-rss` | 保留原服务；不修改端口和容器，避免 Nginx 配置互相覆盖 |

拓扑：`Windows/macOS → HTTPS mkd.186512.xyz (Cloudflare) → Nginx 443 → Go API 127.0.0.1:8787 → TLS → Supabase PostgreSQL`。Go API 再通过官方 HTTPS 调用 Mercado Libre。

## 2. 根据用户 SSH 输出确认的事实

- OS **Ubuntu 24.04.4 LTS**，不是先前口头说的 Ubuntu 22.04。
- CPU **2 vCPU**，物理内存显示 **3.6GiB**；`free -h` 的 available 约 **2.3GiB**。Swap 约 **980MiB 已用**，需观察 `vmstat` 是否持续 `si/so` 活跃，不能仅凭已用量判断不够用。
- 根盘 **59GB / 已用21GB / 剩余37GB**（约36%使用率）。
- `nginx -t` PASS；Nginx 正在监听 **0.0.0.0:80**，此前 **443/8787 未监听**；Caddy inactive。
- Docker 中已有 `we-mp-rss` 容器，`127.0.0.1:8001 -> 8001/tcp`。
- 用户报告 **Cloudflare DNS 已配置**，但尚无独立公网 DNS / HTTPS 请求的成功证据。
- Mercado Libre 开发者应用已成功创建并获得 Client ID；用户此前表示主账号 KYC 尚未完成。**App 创建成功不等于 OAuth 授权已通过**。

上述事实基于用户报告和 SSH 输出；未获得远程 SSH、Cloudflare 或 Supabase 管理权限，没有执行过 DNS 修改、证书申请或部署。

## 3. 自行申请证书：首选 Let's Encrypt

**可以自行申请免费的公网证书。** Let’s Encrypt 是公开可信 CA，Certbot 能自动续期；比仅面向 Cloudflare 代理的 Origin CA 更通用。Cloudflare Full (strict) 同时支持有效的 Let's Encrypt 证书和 Cloudflare Origin CA。

实施顺序：

1. 确认 Cloudflare 里 `mkd` A 记录指向目标服务器公网 IP，橙云代理已启用；从外部验证解析。检查云平台**安全组**的 80/443 入站、主机防火墙 `ufw status`（不要为此直接启用 UFW，防止影响 SSH）。
2. 备份 Nginx 相关现有配置，检查 `sudo nginx -T` 中现有的 `server_name`，确认没有与 `mkd.186512.xyz` 重复。先在 `sites-available` 新增**独立**该域名的 HTTP(80) 站点，并启用相应软链接；绝不覆盖默认站点或现有 `we-mp-rss` 配置。
3. 每次修改后先 `sudo nginx -t`；通过后才 `sudo systemctl reload nginx`。证书申请方式建议 **Certbot Nginx 插件**：在确认 80 端口可以从公网通过 Cloudflare 正常访问新域名、没有挑战拦截后运行 `sudo certbot --nginx -d mkd.186512.xyz`。**插件会修改匹配的 Nginx 虚拟主机**，事先备份并核对生成的 diff。
4. 如果 HTTP-01 被代理规则/WAF/重定向阻碍，改用 **DNS-01**（Cloudflare TXT 校验/自动化 DNS 插件），不要停止当前占用 80 的 Nginx。尤其不要使用会抢占 80 端口的 `certbot --standalone`。
5. 证书签发后验证 `sudo certbot certificates`、`sudo certbot renew --dry-run`、Nginx 443 监听及源站证书链。确认源站端口 443 对 Cloudflare 可达。
6. Cloudflare SSL/TLS 对 **`mkd.186512.xyz`** 使用 **Full (strict)**。若其他主机尚不支持 HTTPS，不贸然修改整个 Zone 的 TLS 策略；在账号支持的情况下使用主机名限定的 Configuration Rule，或先逐个确认其他源站证书。
7. `curl -i https://mkd.186512.xyz/` 检查公网证书与预期返回内容。只有 DNS、TLS、443 和目标主机同时通过才可宣布 HTTPS ready。

备选：Cloudflare `SSL/TLS → Origin Server` 签发 Origin CA，放到源站 Nginx 使用，仍选 Full (strict)；**此证书通常不被绕过 Cloudflare 的普通浏览器直接信任**，故本项目优先 Let's Encrypt。不要把私钥上传到 GitHub 或聊天记录。

参考：
- [Ubuntu：获取 TLS 证书](https://ubuntu.com/server/docs/how-to/security/obtain-tls-certificates/)
- [Cloudflare：Full (strict)](https://developers.cloudflare.com/ssl/origin-configuration/ssl-modes/full-strict/)

## 4. Nginx/Go 最小上线边界

- `mkd.186512.xyz` 在 Nginx **单独匹配 server_name**；HTTP 80 用于 ACME 校验/必要的 HTTPS 跳转；HTTPS 443 转发到 Go `127.0.0.1:8787`。
- 初期只对外开放必要路径：`/healthz` 可返回不含凭据的简短健康响应；`/oauth/mercadolibre/callback` 必须等 OAuth handler 真正实现后再接入授权流程；其他 API 未完成身份认证前不得放行敏感读写。
- Go 进程应由最小 `systemd` 单元托管，非 root 运行，可重启且日志不包含 Secret/授权码/Token；配置只在服务器安全环境里，不写入 Git。
- M0-01 的 `backend/cmd/api/main.go` 是**开发版本**，只提供本地 `/healthz`，绝不可因为 Nginx 反向代理已经配置就宣称 OAuth 功能完成。部署验证可先用健康接口打通请求链，再开发正式回调。
- 浏览器跨域策略、Tauri CSP、客户端 API 基址应随正式域名配置，不能把测试期 `http://127.0.0.1:8787` 当公网生产地址；改动应与对应的后续 M0-02 代码一起测试。

## 5. Supabase 托管 PostgreSQL 的最简使用法

- **M0-01 健康检查暂不需要建库**；M0-03 / Issue #6 的 PR-A 开始采集商品写入时，按需准备 Supabase PostgreSQL 或 CI 临时 PostgreSQL，使用最小采集表及 SQL 迁移；真正店铺 Token 的存储与授权迁移留在 M1-D，不能在 PR-A 预建授权表。
- Go 服务端使用标准 PostgreSQL 连接；**Issue #14 已批准复用现有 Supabase Auth 进行 Una Seller 登录，实施边界见 [ADR-0006](../decisions/0006-auth-identity-boundary.md)**；仍不引入 Edge Functions、Realtime、Storage。桌面端不得持有数据库 URL、密码、Service Role Key。
- 在 Ubuntu 服务器上检查 IPv6 出站能力。**IPv6 可用**优先使用 Supabase Dashboard 提供的 Direct connection；**仅 IPv4** 时使用官方 Shared Session Pooler（端口 **5432**，非 serverless 常用 Transaction Pooler **6543**），连接字符串必须从 Supabase Dashboard 的 Connect 复制，开启 TLS 校验，不猜测主机名或密码。
- 持久化客户端连接池先保持低并发，例如按实际调用量配置 2–5 个连接，而非部署 Redis 或多个实例；出现连接耗尽时先观测。
- Token、Refresh Token 应在服务端受控、必要时加密存储；备份和恢复责任需明确。**免费计划可能有容量/暂停限制**，不能当成正式 SLA 或无需备份的长期生产保障。
- 官方说明：[Supabase 连接 PostgreSQL](https://supabase.com/docs/guides/database/connecting-to-postgres)。

## 6. Mercado Libre OAuth 与凭据边界

- **计划使用的 OAuth 回调 URI**：`https://mkd.186512.xyz/oauth/mercadolibre/callback`。**是否已在 Mercado Libre 应用后台实际登记仍需核对**；OAuth 实施时重定向 URI 必须与后台登记值完全一致。
- 用户已创建开发者应用并拿到 Client ID，**Client Secret 不放入项目源码、客户端或文档**；在后端安全配置中注入。不要在 URL 查询参数或日志里暴露敏感 Token。
- 实现必要的授权状态 `state` 生成/一次性校验及到期失效、授权码交换/一次性使用、Token 安全保存、刷新与解绑、`/users/me` 真实只读验证。
- HTTPS 的根路径返回 200 只是 TLS 连通性验收；只有真实回调 handler、正确权限和实际 OAuth 测试通过后，才能标记 **店铺连接 PASS**。
- Notification configuration 暂不启用，后续确有订单/消息通知需求再考虑 Webhook；不要提前配置不存在的接收服务。

## 7. 分段验收与禁止提前宣告完成

| 阶段 | 验收目标 | 现在 |
| --- | --- | --- |
| M0-01 | Windows 实机 Tauri 启动、Go 本机健康检查；macOS 双架构至少构建/启动记录 | **PR #2 已于 2026-11-14 合并**；Windows/macOS 实机启动仍未验收 |
| M0-02a | Cloudflare DNS 公网解析 + Let's Encrypt 自动续期 + 独立 Nginx 443；旧 RSS 正常 | DNS 由用户配置；其余未验收 |
| M0-02b | Go API 经 `https://mkd.186512.xyz/healthz` 正确响应；与真实 Supabase 连接需要另行验证 | 尚未部署/连库 |
| **M0-03 / Issue #6** | 采集 API、持久化/去重/读回、桌面采集箱、Amazon.com 单商品 MV3 采集闭环 | 尚未完成真实采集数据链路；与公网 HTTPS 部署独立验收 |
| **M1-D** | Mercado Libre OAuth 回调、state、Token 安全保存、`/users/me`、刷新/解绑与既有商品只读同步 | 开发者应用已创建；真实授权未测，KYC 可能影响权限 |

## 8. 必须避免的操作

- 不在现有 Nginx 80/443 上盲目覆盖或关闭默认站点，不停止 `we-mp-rss`，不因为签证书而停止 Nginx。
- 不未经检查就开启 UFW、不把 Go 8787 公网监听、不在 Git 中记录 Supabase DB 密码、Cloudflare API Token、证书私钥、Mercado Libre Client Secret。
- 不在未知其他站点兼容性时全局切换 Cloudflare SSL/TLS 模式；不在源站证书未验证时宣布 HTTPS 完成。
- 不在现有服务器上**默认自建 PostgreSQL**；CI 可用临时 PostgreSQL，本机 PostgreSQL 仅在有真实测试需要时选用。Redis、队列、K8s 等不因未来需求提前部署。
