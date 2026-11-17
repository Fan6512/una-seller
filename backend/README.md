# Una Seller Go API：采集箱与 Supabase Auth

Private schema `app` holds users, source products, source images and actual variants. SQL migration: `migrations/0001_source_collections.sql`. Issue #14 uses Supabase Auth to identify users; **Mercado Libre OAuth and public deployment are not implemented**.

本地开发（数据库已迁移时跳过 `migrate`；Supabase 项目中先创建可登录的受控用户并关闭公开注册）：

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
- The Go server still listens **only on 127.0.0.1:8787**. Public HTTPS, reverse proxy, CSP/CORS for the eventual domain, and browser-extension pairing are **not yet deployed**; do not expose this local listener directly to the Internet. Database passwords and service keys must remain server-side.

GitHub Actions provisions disposable PostgreSQL 17 for synthetic integration tests. It requires `TEST_DATABASE_URL` in CI, prints named test results, checks disconnect/reconnect read-back, and runs explicit seed → Docker PostgreSQL service restart → verify phases for process-restart persistence. This does **not** establish host/power-loss recovery or live Supabase connectivity.

`app` is a private schema with RLS enabled and no anon/authenticated grants or policies. The local Go backend must connect using the schema/table-owning PostgreSQL role (which bypasses RLS by default). A different non-owner role without policies cannot read the data; do not silently switch roles or expose the schema via Supabase Data API.
