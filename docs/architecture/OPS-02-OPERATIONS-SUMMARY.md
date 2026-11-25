# OPS-02 phase A — local operations summary (candidate)

Owner-only GET /api/v1/stores/operations-summary, no-store, current verified marketplace mapping, limited to 100 stores, no official GET and no raw message text. Counts are cache observations (including incoming available messages over 7d, NOT unread). Disconnected stores show unavailable, not zero. Schema not yet migrated returns 503. Desktop page polls local API every 30s while mounted/visible, aborts stale request, retains last snapshot on errors and displays Beijing time. No marking read, message/claim write, automatic AI, or official actions. Next: tenant isolation/HTTP method/503 tests, actual desktop validation, notification activity list, filters and actionable task state after official contracts. This is the summary foundation, not complete operations center.

## 2026-11-26 acceptance

- PASS: owner-scoped query, no-store, GET-only, empty owner, disconnected/revoked store hiding cached numbers, and per-metric accuracy are covered by `TestOperationsSummaryOwnerScopedAndAccuratePostgres` (two owners, two stores, excluded states). Desktop is not mounted when settings are open and aborts stale requests.
- FAIL: the requested latest-activity list, actionable to-do state and filters are not implemented. Only seven cached counters are returned. There is no HTTP handler test for 405/503 or a desktop test for the page.
- NOT_RUN: real desktop rendering and 30s refresh against a deployed API.
