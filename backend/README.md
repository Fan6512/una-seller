# Una Seller Go API：采集箱与 Supabase Auth

Private schema `app` holds users, source products, source images and actual variants. SQL migration: `migrations/0001_source_collections.sql`. Issue #14 uses Supabase Auth to identify users. **Issue #21 / Draft PR #23 adds Mercado Libre CBT OAuth, but live seller authorization has not been tested or merged.** Cloud infrastructure verification is tracked separately in Issue #17; this README does not establish current live deployment state.

## Windows：Go API 配置只保存一次

首次在 PowerShell 创建目录 `New-Item -ItemType Directory -Force "$env:LOCALAPPDATA\\UnaSeller"`，再用记事本创建 `%LOCALAPPDATA%\\UnaSeller\\api-config.json`：

```json
{
  "DATABASE_URL": "postgresql://<数据库连接信息>",
  "SUPABASE_URL": "https://rwktwfxzfkzglxwaulcm.supabase.co",
  "SUPABASE_PUBLISHABLE_KEY": "<项目 publishable key>"
}
```

请将 JSON 文件保存为 **UTF-8**（推荐无 BOM；程序也兼容 UTF-8 BOM）。Windows PowerShell 5.1 使用 `>` 创建文件时可能默认写入 UTF-16LE，Go API 无法解析；遇到 `invalid JSON` 时请用记事本另存为 UTF-8，不要复制含密码的文件内容到 Issue。

该文件留在 Windows 用户目录，**以后替换或升级 `una-seller-api.exe` 不用重填环境变量**。它是本地明文配置，不能加入 Git、发布 Artifact 或截图公开；修改后重启 Go API。环境变量仍优先于文件，以便继续在 CI、Linux 和服务器使用。

本地阶段负责人决定暂不关闭公开注册，**但 Supabase Auth 的注册接口仍对公网可达**，正式分发前需要限制。

## Linux/macOS/CI 继续使用环境变量

本地开发（数据库已迁移时跳过 `migrate`）：

```sh
cd backend
export DATABASE_URL='postgresql://...'
export SUPABASE_URL='https://<your-project>.supabase.co'
export SUPABASE_PUBLISHABLE_KEY='<sb_publishable_...>'
go run ./cmd/api migrate
go run ./cmd/api
```

- `GET /healthz` is available with or without a database. **All collection GET/POST/DELETE routes require `Authorization: Bearer <Supabase Auth access_token>`**; missing/invalid sessions return HTTP 401. The backend verifies the token via `/auth/v1/user` and uses the verified `user.id` as the owner; `UNA_SELLER_LOCAL_USER_ID` is no longer used.
- `POST /api/v1/collections` accepts only structured JSON, returns 201 (new), 200 (recapture), or an error (not saved).
- `GET /api/v1/collections?q=...&limit=30&offset=0` and `GET /api/v1/collections/{uuid}` read persisted source data.
- `DELETE /api/v1/collections/{uuid}` permanently deletes only the authenticated user's collected source product (204); other users' records and missing UUIDs return 404. Its source images/variants cascade; this never deletes Mercado Libre listings.
- `price` is a decimal **string** or null, never a JSON floating-point number; missing is distinct from null. `bullet_points` and images retain order. Variants are **only real captured attribute combinations**.
- Partial recaptures merge valid supplied fields and preserve prior data: missing fields, nulls, and empty arrays do not erase saved content.
- Complete recaptures preserve omitted fields; explicit null clears nullable scalars, and either null or [] clears `bullet_points`, top-level `images`, and `variants`. An explicitly supplied variant with `image_urls: null` or `[]` has no variant images. Do not send null to mean 'not collected'; omit the field instead.
- Before running the migration on Supabase, obtain explicit approval of this PR. The schema is private (`app`); there are no Data API grants.
- The Go server listens **only on 127.0.0.1:8787** even on Ubuntu. Cloud mode `UNA_SELLER_CLOUD=1` requires DB/Auth environment values before starting; Nginx serves public HTTPS and the exact Tauri origins have CORS access to `/healthz` and collections. **This is code readiness, not a claim that cloud HTTPS is deployed**; browser-extension pairing remains unimplemented. See [cloud deployment guide](../docs/architecture/CLOUD_GO_API_DEPLOY.md). Database passwords and service keys must remain server-side.

GitHub Actions provisions disposable PostgreSQL 17 for synthetic integration tests. It requires `TEST_DATABASE_URL` in CI, prints named test results, checks disconnect/reconnect read-back, and runs explicit seed → Docker PostgreSQL service restart → verify phases for process-restart persistence. This does **not** establish host/power-loss recovery or live Supabase connectivity.

`app` is a private schema with RLS enabled and no anon/authenticated grants or policies. The local Go backend must connect using the schema/table-owning PostgreSQL role (which bypasses RLS by default). A different non-owner role without policies cannot read the data; do not silently switch roles or expose the schema via Supabase Data API.

## Mercado Libre CBT 店铺连接（Issue #21 / Draft PR #23）

这是**未合并的 OAuth + Refresh/门禁切片**：Go 已实现一次性 OAuth state、可选 PKCE、官方 `/oauth/token` 和 `/users/me`、官方 CBT 主账号/子站点映射、Token 加密存储、单次轮换续签、店铺列表与断开；Tauri 已有店铺管理页面。代码与合成数据测试不代表真实卖家授权 PASS，当前 HEAD CI 状态以 GitHub 检查为准。

- 服务器私有配置：`MELI_CLIENT_ID`、`MELI_CLIENT_SECRET`、`MELI_REDIRECT_URI`、`MELI_TOKEN_ENCRYPTION_KEY`，如 DevCenter 启用 PKCE，另设 `MELI_PKCE_ENABLED=1`。具体占位配置见 `deploy/una-seller/api.env.example`；Secret、加密密钥不得放入桌面构建、Git 或公开日志。缺少完整配置时，授权 API 返回 503。
- 在 DevCenter 登记的 HTTPS Callback URI 与 `MELI_REDIRECT_URI` **必须完全相同**：`https://<your-owned-api-domain>/api/v1/mercadolibre/oauth/callback`；路径由服务端固定，不能增加随用户变化的 query。
- 服务器代理模板 `deploy/una-seller/nginx-mkd.conf` 已包含店铺 API 和 Callback，Callback 关闭 Nginx access log，避免 code/state 被日志收集。**仓库模板更新不等于已更改真实服务器**。
- 显式执行 `go run ./cmd/api migrate`（包含 `0002_meli_store_connections.sql`）时必须先核对已有数据库及备份/审批，不做自动迁移。
- `POST /api/v1/stores/oauth/start`（Supabase Bearer）：返回官方 `authorization_url`；`GET /api/v1/mercadolibre/oauth/callback`（一次性 state）：交换 Token 并绑定官方卖家身份；`GET /api/v1/stores`（Supabase Bearer）：仅列出本人店铺；`POST /api/v1/stores/{uuid}/disconnect`：仅断开本人店铺；`POST /api/v1/stores/{uuid}/connection-check`：仅本人触发 Token 状态检查及必要的自动续签，返回去敏状态，不含 Token。上述 API 不返回 access/refresh token。
- **CBT 账号模型**：`meli_store_connections.seller_id` 是 CBT 主账号的 global user_id，一账号只有一套 Token。由 `GET /marketplace/users/{user_id}` 得到 `marketplaces[]`（站点 site_id + 子账号 user_id + logistic_type）并保存在同一行 JSONB；失败时状态为 `unverified`，不是猜测五站全开。只有 verified 的明确站点才能通过 `RequireSiteConnection`。
- **Token 自动续签与统一门禁已写入 Go**：`RequireStoreConnection` / `RequireSiteConnection` 只在服务端供后续订单和刊登复用；进程内 singleflight + PostgreSQL 短事务行锁/刷新租约，**不在数据库事务里等待外部 HTTP**。平台返回 `invalid_grant` 标记 `reauth_required`；无法确定续签是否成功标记 `refresh_unknown`，拒绝盲目重试旧 Refresh Token。
- 当前仅实现连接状态/站点映射和服务端门禁组件，**不等于真实 Orders/刊登 API 已经接入**。手动强制续签仅用于服务端受控测试，不向普通桌面 UI 暴露不受限的强制轮换按钮。
- Windows 真实 OAuth 操作、DevCenter 权限、线上 Callback、第二卖家店铺以及 macOS 运行尚未验收；请如实记 NOT_RUN/UNVERIFIED。


## CBT 分类中文词典

在 API 与执行更新命令的服务账号环境中设置同一个绝对目录 `CBT_CATEGORY_DICTIONARY_DIR`。首次载入只需要目录写权限和访问[官方 CBT Dump](https://global-selling.mercadolibre.com/devsite/category-dump-global-selling)的网络连接。

```sh
cd backend
export CBT_CATEGORY_DICTIONARY_DIR=/absolute/private/una-seller/cbt-category
TRANSLATION_PROVIDER=disabled go run ./cmd/api cbt-categories update
go run ./cmd/api cbt-categories update -max-batches 50
```

第一条更新读取、校验并发布官方分类数据，供缺少翻译配置时显示原文。第二条沿用配置的 Azure 或百度 Provider，最多处理 50 个批次。省略 `-max-batches` 时持续处理，直到完成或 Provider 暂停。部署后可用 `una-seller-api cbt-categories update` 作为 cron 入口，由同一服务账号运行。同一目录仅允许一个更新进程。锁随文件句柄关闭释放，退出后目录中可保留空的 `update.lock`。

- 命令优先检查官方时间和 MD5。相同版本已完成时跳过翻译。新版本继续使用 ID、名称及完整路径都匹配的译文。运行中断后执行相同命令，从已保存的批次续跑。
- Provider 已配置且更新暂停时，API 继续读取有效词典。关闭 Provider 时，验证通过的新官方版本可原子发布，未译名称显示原文。停用和已移除类目供历史阅读，已知不可上市的类目从新预测候选排除。
- 命令输出 JSON。 `official` 是当前 Dump 的类目数，`disabled` 和 `removed` 分别是停用与已移除数量，`translated` 是已有译名数量，`unchanged` 是确认沿用原名数量，`failed` 与 `unprocessed` 分别是失败待重试、尚未处理数量；`skipped` 是确定无法构造安全翻译输入而暂时保留官方英文的数量，`skip_samples` 最多提供 20 条类目 ID 与原因。`skipped` 不阻止其他已翻译类目发布，后续同版本更新仍可尝试补译，失败的外部翻译请求不会记作 `skipped`。查看 `published` 和 `state` 判断本次执行是否发布，或是否仅保存了进度。
- 需要人工调整时，在该目录写入 UTF-8 的 `overrides.json`，按下例保留官方 ID 与完整路径。新版本改名或移动时旧人工译名留在文件中，只有路径仍匹配时才生效。

```json
[
  {
    "path": [
      { "id": "CBT100", "name": "Home" },
      { "id": "CBT101", "name": "Other" }
    ],
    "zh": "其他家居用品"
  }
]
```

分类预测与编辑器读取服务端词典。已刊登站点商品只有在本地已同步并验证 CBT 父商品归属时，才附带 CBT 类目路径。属性和技术规格继续读取现有官方规则。机器及人工中文名称只供阅读，客户端保存官方分类 ID 和原文。
