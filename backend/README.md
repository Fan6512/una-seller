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
- Partial recaptures merge valid supplied fields and preserve prior data; complete recaptures can replace explicitly supplied arrays and clear explicit nulls.
- Before running the migration on Supabase, obtain explicit approval of this PR. The schema is private (`app`); there are no Data API grants.
- The Go server listens **only on 127.0.0.1:8787**. The fixed local UUID is a temporary, server-configured owner **not real authentication**. **Never proxy `/api/v1/collections` publicly** until real user authentication and extension pairing have been implemented.

GitHub Actions provisions disposable PostgreSQL 17 for the synthetic integration test and validates disconnect/reconnect read-back. It does **not** establish database-process restart persistence or live Supabase connectivity.
