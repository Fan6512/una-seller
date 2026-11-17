# ADR-0006：Una Seller 登录与用户身份边界

- 日期：2026-11-17
- 状态：**Accepted（技术选型）；IMPLEMENTING（Issue #14 / PR #15）；NOT_DEPLOYED**
- 依据：[Issue #14](https://github.com/Fan6512/una-seller/issues/14)、[ADR-0004 简单优先](0004-simplicity-first.md)、[ADR-0005 基础设施](0005-m0-infrastructure.md)
- 本文仅冻结当前桌面 ERP 身份方案，不提前设计商业 SaaS、团队 RBAC 或 Mercado Libre 店铺 OAuth。

## 决策

1. **现有服务复用**：继续使用 Supabase PostgreSQL 作为唯一数据库，额外启用**现有 Supabase 项目的 Auth（开源 GoTrue）**负责邮箱密码、令牌签发和续签；不部署 Keycloak、Casdoor、Logto、独立 Node Auth 服务、密码表或另一套数据库。原有“不预设 Supabase Auth”属于 M0 初始数据库选型，已被本次明确批准的认证需求部分更新；**Edge Functions、Realtime、Storage 仍不采用**。
2. **登录模式**：先由项目管理员通过 Supabase Authentication → Users 创建受控邮箱账号，项目设置必须**关闭公开注册**；UI 不提供注册入口。第一版不做 SMTP 自建、找回密码、邮箱验证码、第三方登录、邀请与多因素登录。关闭注册需要在 Supabase Dashboard 实际操作验证，**仅隐藏前端注册按钮不算关闭**。
3. **认证校验**：Go API 统一使用 Supabase Auth 官方 `GET /auth/v1/user` 以请求中的 Access Token 直接验证用户（不会依赖公开 publishable key 自行解析 JWT）。只有认证服务器确认的用户 UUID 能作为 `owner_user_id`；**拒绝 publishable key、旧 anon key、过期/伪造 token、未登录用户**，不能把 API Key 当作 Auth Token。未鉴权接口返回 401 和 `WWW-Authenticate: Bearer`，身份提供方不可达则 503 且拒绝读写，跨用户详情按 404 处理，不泄露记录存在性。
4. **本期暂不做 JWKS**：当前 Go API 并发/规模较小，`/auth/v1/user` 的额外网络开销可接受；这样也不依赖项目究竟是 HS256 还是非对称签名。未来若实测延迟/限流成为瓶颈，再检查签名模式、JWKS 轮换和有界缓存后单独替换。**不把 `SUPABASE_JWT_SECRET` 发给客户端、写入 Git 或为了本期性能强行引入它**。
5. **会话存储（MVP 显式取舍）**：桌面 WebView2/WebKit 的本地 `localStorage` 保留短期 Access Token 与 Refresh Token，用于关闭重开后的登录恢复。它对同一系统用户可读，**不是 Windows Credential Manager / DPAPI 意义上的加密凭据仓库**。当前个人自用阶段接受此局限；不存明文密码，不记录或打印 Token，不把 Token 上传到 Git/CI。正式向陌生用户分发或处理更高敏感度业务前，单独评估系统凭据库和设备退出/撤销策略，不在此次 PR 引入 Rust keyring 插件。
6. **SDK 选择**：本期直接调用 Supabase Auth 官方 HTTP 端点（password、refresh_token、logout），不额外引入 `@supabase/supabase-js`，因为既有代码只需邮箱密码与一次并发刷新的最小范围；Go 鉴权与跨账号隔离保留现有自动测试；前端暂不新增测试框架，刷新并发、登录过期、退出清空等由 Windows 登录实机清单验证并将结果记录在 PR，未执行不得标 PASS。出现可复现回归时再按需补最小自动测试，不允许因“代码少”忽略会话正确性。日后扩展 OAuth/多端复杂登录时再评估使用 SDK，**不要同时维护两套 Auth 客户端**。
7. **跨域与公网**：Go 允许已知 Tauri 来源的 `Authorization` 预检，Tauri CSP 仅放行所选 Supabase Auth HTTPS 源。公网 Go API 需要独立完成 HTTPS、代理与地址切换，不得直接对外暴露 127.0.0.1:8787 的开发服务。商品表继续按 `owner_user_id` 过滤，数据库密码仅保留在 Go 服务端。

## 兼容和禁止事项

- PR #13 用固定 `UNA_SELLER_LOCAL_USER_ID` 写入的旧测试商品不自动归入任何新的 Supabase Auth 账号；必要时经负责人授权再单独迁移。
- Mercado Libre OAuth 是平台**店铺授权**，与 Una Seller 用户登录不同，仍按 M1-D 单独实施。
- 用户关闭登录页面后，禁止继续呈现上一账号已加载的商品列表和详情；401 错误不显示原始带有 Authorization 或 Token 的请求信息。
- 只有经受控 Windows 实机验证成功登录、读回当前账号、重启恢复、退出清空与跨账号隔离后，才能标记登录端到端 PASS；macOS 未测保持 UNVERIFIED。

## 决策理由

沿用现有托管服务减少部署和密码安全维护，不增加新的认证服务器；官方 Auth User API 避免实现 JWT 验签、HS256/JWKS 分支与密钥轮换。localStorage 是 MVP 的已知存储权衡，不是宣称零风险。对未配置公开注册限制、网络延迟与未来分发风险保留明确任务，不将其伪装成已验收。
