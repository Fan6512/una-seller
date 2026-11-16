# M0-03 PR-A: local source-collection API

Private schema `app` holds users, source products, source images and actual variants. SQL migration: `migrations/0001_source_collections.sql`. No OAuth, Supabase Auth or public web endpoint is implemented here.

Against your disposable local PostgreSQL:

```sh
cd backend
export DATABASE_URL='postgresql://...'
export UNA_SELLER_LOCAL_USER_ID='a-persistent-random-uuid'
go run ./cmd/api migrate
go run ./cmd/api
```

- `GET /healthz` is available with or without a database.
- `POST /api/v1/collections` accepts only structured JSON, returns 201 (new), 200 (recapture), or an error (not saved).
- `GET /api/v1/collections?q=...&limit=30&offset=0` and `GET /api/v1/collections/{uuid}` read persisted source data.
- `price` is a decimal **string** or null, never a JSON floating-point number; missing is distinct from null. `bullet_points` and images retain order. Variants are **only real captured attribute combinations**.
- Partial recaptures merge valid supplied fields and preserve prior data: missing fields, nulls, and empty arrays do not erase saved content.
- Complete recaptures preserve omitted fields; explicit null clears nullable scalars, and either null or [] clears `bullet_points`, top-level `images`, and `variants`. An explicitly supplied variant with `image_urls: null` or `[]` has no variant images. Do not send null to mean 'not collected'; omit the field instead.
- Before running the migration on Supabase, obtain explicit approval of this PR. The schema is private (`app`); there are no Data API grants.
- The Go server listens **only on 127.0.0.1:8787**. The fixed local UUID is a temporary, server-configured owner **not real authentication**. **Never proxy `/api/v1/collections` publicly** until real user authentication and extension pairing have been implemented.

GitHub Actions provisions disposable PostgreSQL 17 for synthetic integration tests. It requires `TEST_DATABASE_URL` in CI, prints named test results, checks disconnect/reconnect read-back, and runs explicit seed → Docker PostgreSQL service restart → verify phases for process-restart persistence. This does **not** establish host/power-loss recovery or live Supabase connectivity.

`app` is a private schema with RLS enabled and no anon/authenticated grants or policies. The local Go backend must connect using the schema/table-owning PostgreSQL role (which bypasses RLS by default). A different non-owner role without policies cannot read the data; do not silently switch roles or expose the schema via Supabase Data API.
