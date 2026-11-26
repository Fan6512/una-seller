# OPS-03 phase A — known-resource compensation (candidate)

Disabled by default. Same durable queue and shared GET coordination, no cursor mutation. At most eight hints per sweep, once per minute in the single approved API process. Paid known orders 30m, non-final known shipments 60m, deduplicated by shipment. Exact current mapping required, admission binds original store/owner/parent and suppresses outstanding jobs, recent receipts/dead letters, global/store cooldown or backlog >=32. 0018 records source separately. No billing, procurement, historical scan or official writes.

Discovery of orders absent from the database remains pending: independent persisted search windows/cursor, overlap and bounded paging. Do not reuse the manual historical cursor. WH-06 missed-feed contract and coverage remain separate. Acceptance: tests, restart, queue contention, revoked scope, 429, source provenance, 30/60m boundaries and load with pool 6 before enable. Multi-process sweep budget is not yet supported (admission still deduplicates/caps backlog).

## 2026-11-26 acceptance

- PASS: bounded known-order/shipment sweep, 30/60 minute gate, inbox/job dedup, source provenance, store/child/site revalidation, shared read coordination suppression, 32-job backlog cap and new-process persistence. `TestCompensationCooldownBacklogAndRestart` covers cooldown, backlog and restart on PostgreSQL; existing `TestKnownResourceCompensation` covers 429 coordination, dedup and terminal shipment exclusion.
- FAIL: orders absent from the local database are never discovered, and there is no persisted search window/cursor with bounded paging or overlap. Manual priority is only the in-process foreground gate; it was not measured under concurrent compensation.
- NOT_RUN: real official search/GET budget, multi-process sweep, production 429, load at pool 6.
