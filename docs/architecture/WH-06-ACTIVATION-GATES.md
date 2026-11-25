# WH-06 phase A — activation preparation (candidate, no production changes)

Default remains off. Enabled process requires explicit reviewed topic set, matching worker registration, required migration tables/column, and private proxy-to-Go token. Disabled known topic is 503 (never false ACK). Compensation requires both order and shipment handlers. No automatic official subscription writes. Nginx example explicit 404; separate inactive template allows only reviewed current official sources, deny-all fallback and overwritten private header. Never deploy template directly without verified includes and http-level rate limit zone. Do not store token or real IP allowlist in Git. With CDN/trusted proxy validate real-IP topology or do not enable.

## Mandatory owner-approved production gates

- Confirm exact UNA Seller host 152.136.41.105, service /opt/una-seller, candidate commit; verify same binary hash and one API process; preflight encrypted backup/restore and PG migrations 0014–0018 in isolation; review any later migration. Separate migration/deploy approval. Increase pool to six only after measuring capacity. 
- Establish current official app account, six topic contracts and permissions with bounded read-only checks. In particular messages GET may implicitly mark read; do not enable messages without proving safe behavior. Unknown 403/401 do not guess scope or mutate account. Verify item/claim site IDs and timestamps. Start with orders/shipments only, verify delivery then extend explicit set after approval. 
- Verify current official source IP evidence from official docs/support and actual transport topology; reject spoofed forwarding/header, malformed path, bad methods and rate spikes; Nginx syntax/rollback tested on nonproduction. Measure public ACK p95/p99 and worst case below 500ms with valid source, offline503/restart, actual retry and queue health. 
- Missing-feed contract is NOT implemented: verify actual CBT missed_feeds GET URL, authorization, pagination, retention and identifiers, read-only sample. Implement bounded replay through same validated durable inbox with source provenance and no cursor advancement before persistence; never treat unsupported API as empty result. Complete missing-new-order OPS-03 discovery and compensation tests. 
- Only after separate approval modify official app topic subscription; reconcile callback configured and observed; rollback disables subscription/edge, preserve queued data and avoid deleting receipts. Report all results PASS/FAIL/NOT_RUN, no raw buyer text, token, secrets or IP logs.

No production migration, Nginx install, enable, deployment or official writes executed by this PR.

## 2026-11-26 read-only endpoint check

- PASS: `GET /missed_feeds?app_id=$APP_ID&limit=10` returns 200 for the current CBT app; default page size is 10 and `topic` filtering is accepted. `marketplace_questions` returned 0; the first page contained `marketplace_item_competition` and `marketplace_items`, with callback codes 404/0 because the receiver is disabled.
- BLOCKER: missed-feed `resource` is not the live notification shape. Observed `/marketplace/resource/items/{itemId}/` with trailing slash and a UUID `_id`, while live `marketplace_items` uses `/marketplace/items/{id}`. The messages example uses a bare message ID. Replay must normalize per topic and re-read through the allowlisted GET, never trust `request.data`. Unknown topics such as `marketplace_item_competition` must stay rejected.
- FAIL: safe replay is not implemented, so no cursor advance, overlap or partial-page behavior can be tested.
- NOT_RUN: Nginx syntax/rollback on a nonproduction host, real source-IP allowlist, public ACK p95/p99, retry and queue-health under edge 503.
