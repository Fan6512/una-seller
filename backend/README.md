# Una Seller Go API：采集箱与 Supabase Auth

Private schema `app` holds users, source products, source images and actual variants. SQL migration: `migrations/0001_source_collections.sql`. Issue #14 uses Supabase Auth to identify users; **Mercado Libre OAuth is not implemented**. Cloud Nginx/systemd deployment is prepared separately (M0-02B), but actual Ubuntu and public HTTPS verification remain pending.

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

- `GET /healthz` is available with or without a database. **All collection GET/POST routes require `Authorization: Bearer <Supabase Auth access_token>`**; missing/invalid sessions return HTTP 401. The backend verifies the token via `/auth/v1/user` and uses the verified `user.id` as the owner; `UNA_SELLER_LOCAL_USER_ID` is no longer used.
- `POST /api/v1/collections` accepts only structured JSON, returns 201 (new), 200 (recapture), or an error (not saved).
- `GET /api/v1/collections?q=...&limit=30&offset=0` and `GET /api/v1/collections/{uuid}` read persisted source data.
- `price` is a decimal **string** or null, never a JSON floating-point number; missing is distinct from null. `bullet_points` and images retain order. Variants are **only real captured attribute combinations**.
- Partial recaptures merge valid supplied fields and preserve prior data: missing fields, nulls, and empty arrays do not erase saved content.
- Complete recaptures preserve omitted fields; explicit null clears nullable scalars, and either null or [] clears `bullet_points`, top-level `images`, and `variants`. An explicitly supplied variant with `image_urls: null` or `[]` has no variant images. Do not send null to mean 'not collected'; omit the field instead.
- Before running the migration on Supabase, obtain explicit approval of this PR. The schema is private (`app`); there are no Data API grants.
- The Go server listens **only on 127.0.0.1:8787** even on Ubuntu. Cloud mode `UNA_SELLER_CLOUD=1` requires DB/Auth environment values before starting; Nginx serves public HTTPS and the exact Tauri origins have CORS access to `/healthz` and collections. **This is code readiness, not a claim that cloud HTTPS is deployed**; browser-extension pairing remains unimplemented. See [cloud deployment guide](../docs/architecture/CLOUD_GO_API_DEPLOY.md). Database passwords and service keys must remain server-side.

GitHub Actions provisions disposable PostgreSQL 17 for synthetic integration tests. It requires `TEST_DATABASE_URL` in CI, prints named test results, checks disconnect/reconnect read-back, and runs explicit seed → Docker PostgreSQL service restart → verify phases for process-restart persistence. This does **not** establish host/power-loss recovery or live Supabase connectivity.

`app` is a private schema with RLS enabled and no anon/authenticated grants or policies. The local Go backend must connect using the schema/table-owning PostgreSQL role (which bypasses RLS by default). A different non-owner role without policies cannot read the data; do not silently switch roles or expose the schema via Supabase Data API.
