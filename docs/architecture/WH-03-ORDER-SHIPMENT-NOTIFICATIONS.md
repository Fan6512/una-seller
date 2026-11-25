# WH-03 订单与官方货件通知处理

基础：OPS-01 #99 已合并。生产仍默认禁用，启用、迁移、官方订阅与真实通知验收须独立审批。

## 资源合同

- 订单通知只将 resource 里的订单 ID 用作官方 GET 提示，读取 `/marketplace/orders/{id}`；不将 pack_id 当作订单 ID。保留 pack 关联。
- [官方订单文档](https://global-selling.mercadolibre.com/devsite/manage-orders-cbt)。
- [官方货件文档](https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/manage-shipments) 说明新版货件资源不再保证 order_id；货件先到时调用 `GET /shipments/{id}/orders`，要求 `X-New-Domain: true`，再逐一回读订单。货件详情继续使用 `/marketplace/shipments/{id}` 与 `x-format-new: true`。
- 本轮根据官方文档索引与现有接口实现；实际 CBT 权限、字段变体、混合卖家货件及新版关联接口未做真实取证。

## 数据与执行边界

- 模块 notifications 负责队列、租约、fencing 及事务生命周期；meli 负责授权官方回读、订单/货件解析及本地业务写入。两类处理器注册到同一个现有 Service。
- 通过已验证父子卖家映射、官方订单 seller/site 唯一匹配站点与物流模式；通知用户必须为该父卖家或匹配的子卖家。存在同一子卖家/站点多物流模式时拒绝猜测，待官方证据补齐。货件核对 sender、关联订单 seller、pack（如有）、当前 forward shipment ID、同一站点及模式。混合归属或超过 10 个唯一订单的货件保留失败任务，不截断成成功；回放/补偿属于 WH-06。
- 共用现有 orderGET 的进程内互斥、请求节奏、429 冷却。通知错误转换给 #99 持久化队列冷却。队列冷却仍是队列层的额外保护，不代表人工接口跨进程共享持久化限流；部署前核对单进程拓扑，WH-06 验证人工优先和冷却边界。
- 复用人工批量同步 upsert 和人工货件刷新解析/写入；通知不会推进历史同步游标，不修改采购、账单、SLA 或官方状态，不存原始响应。业务写入只在 worker 持有 fencing token 的事务中执行，再检查精确映射。货件先到可同事务补齐关联订单与快照；重复关联去重；保留原先 last_updated/checked_at 防旧写及货件替换保护。
- 货件通知只更新基础快照（运单号、状态、时间），保留已有详细承运商/历史；详情由原有人工刷新路径读取。订单通知只更新订单，货件独立通知负责快照。暂不新增自动账单、SLA、图片或官方写入。
- 继续单 Worker 串行，不增加并发，不新增依赖或 SQL 迁移（依赖 0014）。

## 验收门禁

新 HEAD CI + 隔离 PostgreSQL；订单先到/货件先到、同 pack 多订单、重复/乱序、同 ID 跨店、父子身份、映射改变、旧货件、并发手工同步、429 双向阻断、超时、回滚与 fencing。新增测试为候选，未独立运行。真实 CBT 接口取证、公网 ACK、来源验证、订阅、漏件补偿与启动开关由 WH-06 独立验收。默认禁用下允许开发合并审批，不把模拟结果冒充实际联调。

## 验收结果

本地门禁：Go 1.26.11 + 隔离 PostgreSQL 17.6（独立数据目录、127.0.0.1:55432、空库 `una_seller_test`）。`gofmt -l` 无输出，`go vet ./...`、`go build ./...` 通过，`go test -p 1 -count=1 ./...` 全部通过（meli 38.1s、notifications 3.1s）。新 HEAD CI：Go Backend 1m37s 通过，`internal/meli` 30.2s、`internal/notifications` 1.4s。

`TestNotificationPostgresRestart` 需要真实停库/起库，本轮在本机隔离实例按 before → 停库 → down → 起库 → after 三阶段实跑通过（停库期间 ACK 为 503，重启后租约与冷却保留、任务继续）。CI 的 persistence 阶段只在 `backend/migrations/*` 或 `internal/collections/*` 变更时触发，本次改动不会触发它，因此不以 CI 里的 SKIP 充当通过。

| 场景 | 证据 | 结果 |
| --- | --- | --- |
| 订单先到后货件到 | `TestNotificationOrderFirstThenShipmentPackPostgres` | 订单先生成且不提前写快照；货件到达后同事务补快照 |
| 货件先到 | `TestNotificationShipmentFirstAndOrderReplayPostgres` | 关联接口建单 + 同事务写快照 |
| 同 pack 多订单 | 同上 | 两笔订单同 pack 700、同 shipment 801、各自快照 |
| 重复通知 | 同上 + worker generation 用例 | 重复关联行去重；重复回放不新增行；执行中到达的新回执保留下一轮任务 |
| 乱序通知 | `TestNotificationKeepsNewerManualOrderAndProtectedData` | 旧 `last_updated` 写入被 `WHERE last_updated<=EXCLUDED` 拒绝，状态不回退 |
| 同 ID 跨店 | `TestNotificationTenancyAndRevocationPostgres` | 另一店同名订单行不变；父/子身份不符在官方读取前拒绝 |
| 站点/物流冲突 | `TestNotificationShipmentFailsClosedOnUnsafeLists`、`TestNotificationOrderScopeRejectsAmbiguity` | 站点冲突、sender 越界、同站点多物流模式均失败关闭 |
| 执行期授权撤销 | `TestNotificationTenancyAndRevocationPostgres` | 读取后断连或映射改变时不写入；错误归为 `ErrScope`（本轮修复） |
| 旧货件 | `TestNotificationStaleShipmentDoesNotReplaceSnapshot` | 订单已换 shipment 时返回 `ErrOrderShipmentChanged`，不覆盖新货件 |
| >10 笔唯一订单 | `TestNotificationShipmentFailsClosedOnUnsafeLists` | 失败关闭，不截断成部分成功，任务保留 |
| 混合卖家 | 同上 | 关联行 seller、订单 seller、sender 三方不一致即失败 |
| 空/部分/畸形响应 | 同上 | 204、206、`{"results":…}`、冲突重复行、空数组全部失败关闭且不落库 |
| 人工并发 | `TestNotificationKeepsNewerManualOrderAndProtectedData` | 采购成本、账单快照、`meli_order_syncs` 游标不变；迟到人工批次也不回退较新订单 |
| 429 双向阻断 | `TestNotificationSharesManualOrderLimit` | 通知 429 → 人工读取不再发请求；人工 429 → Worker 读取不再发请求 |
| 队列冷却边界 | `TestNotificationCooldownBoundaryPostgres` | 冷却跨 topic 保护同一店铺、不波及其他店铺、到期即释放 |
| 回滚与 fencing | `TestNotificationPostgresWorker` | 失效租约的写入整体回滚 |
| 执行超时 | `TestNotificationTimeoutKeepsLeasePostgres` | 超预算不提交、不丢任务；租约到期后同一回执被重新处理 |

官方合同核对（`GET /shipments/{id}/orders`，据官方 CBT 文档）：

- 响应顶层是数组，无 `results`/`paging` 包装；`order_id`、`pack_id` 为字符串，`seller_id` 为数字，`variation_id`/`user_product_id` 可空，每件商品一行会重复同一订单。
- 必须带 `X-New-Domain: true`；不能用货件详情的 `x-format-new: true` 代替。货件详情继续用 `/marketplace/shipments/{id}` + `x-format-new: true`。
- 2025-10-12 起货件资源不再返回 `order_id`/`external_reference`，所以必须走本接口回读。
- 204 表示该货件暂无订单（空响应体，非空数组）；404/401/500/503 见官方状态码。
- 官方文档未给出分页参数、单货件订单上限与具体 OAuth scope；10 笔上限是本服务的保守策略，不是官方限制。
- 真实 CBT 权限、字段变体与返回样例本轮未取证：无 CBT 应用凭据，且验收禁止官方写入与订阅变更。此处结论仅为文档核对 + 合成响应，不能当作实际联调。

本轮修复：`checkNotifiedOrderScope` 在映射被撤销/替换（`pgx.ErrNoRows`）时改为返回 `ErrScope`，避免把安全相关的范围拒绝记成普通 `handler_failed` 并消耗 8 次重试。

已知边界（不阻塞默认禁用合并，留给 WH-06）：

- 无“人工优先”机制。人工与 Worker 共用同一进程内互斥与 429 冷却，双方都会在对方持锁时得到 `ErrOrderBusy`（人工接口 409）。跨进程不共享限流，部署前须确认单进程拓扑。
- 队列冷却每店铺 + 应用一条，冷却下限 60 秒；上游 `Retry-After: 0` 也会被抬到 60 秒。
- 关联接口空数组（204）与 200 部分数组都只能按 8 次退避重试到死信，不能在队列内等待晚到的关联；补偿属 WH-06。200 返回的“少几笔”无法与完整列表区分，只能靠各订单自身通知或人工刷新补齐。
- 货件 `site_id` 在真实 CBT 下是否等于子站点尚未取证；若返回 `CBT`，当前站点一致性检查会失败关闭。

默认禁用不变：`MELI_WEBHOOK_ENABLED=0`，本轮未执行生产迁移、部署、官方订阅或任何官方写入。

## 公共调度收口（后续修订，待新 HEAD 验收）

以上验收结果固定于 739a0636，不是后续修订的验收结论。

- 人工优先：同进程内，后台让出给等待的人工读取；人工仅等待已经在途的后台 GET，人工之间仍保留 busy；不取消已发出的请求。多进程之间通过应用级事务 advisory lock 互斥，但不承诺跨进程人工优先；生产仍按单 API 进程验收。
- 0015 增加私有应用级读取节奏、搜索节奏与冷却表。Webhook 启用时，人工与后台均使用相同协调器；单次官方 GET（包括节奏等待）最长 15 秒，锁随 PostgreSQL 事务释放。429 统一至少冷却 60 秒，尽可能在客户端取消后仍持久化；队列领取检查应用冷却，协调器已持久化的 429 不再重复写入店铺冷却。已有店铺级冷却仅作为旧任务/非共享调用的保护。所有参与同一应用的进程必须使用协调器；协调器默认禁用时仍沿用人工旧行为。生产连接池、故障、时钟和跨进程互斥待隔离及真实拓扑复验；进程崩溃在收到 429 与数据库提交之间仍存在窗口。
- 已知暂不可用授权、正常请求竞争、204/空关联、429 释放租约并退还本次失败额度，延迟重调度；24 小时仍不可用则保留死信，后续补偿。明确官方 seller/sender 身份冲突直接死信；不明确的映射/站点仍失败关闭并受普通重试上限。普通失败及失效租约仍最多 8 次领取。
- 每个货件支持最多 30 个唯一订单的有界顺序回读，全部校验通过后一次 fenced 事务提交；新增 11 单测试。没有实现持久化分批续跑，超过 30 或长期超时仍失败关闭，需后续实现分批续跑/补偿，不得报告已完成该能力。真实 CBT 关联接口、site_id、字段、权限与大货件样本仍 NOT_RUN。
- CI persistence 路径扩展到 notifications、订单/货件/账单写入和协调模块。新增人工优先、应用级互斥/持久冷却、延迟退还额度/超龄死信、永久身份冲突及 11 单回归。

### 92f86299 复验结果

隔离 PostgreSQL 17.6（独立数据目录，`127.0.0.1:55432/una_seller_test`）+ Go 1.26.11：

- `gofmt -l` 无输出，`go vet ./...`、`go build ./...` 通过；`go test -p 1 -count=1 ./...` 全绿（meli 40.8s、notifications 3.5s）。本轮 CI 见 PR #100 checks，不以本地结果代替。
- 三阶段重启（before → 停库 → down → after）实跑通过：停库期间 ACK 503；重启后应用级 `retry_until`、店铺冷却与租约均保留，`claim` 在冷却期内拒绝领取，清除后任务继续。
- 新增 `TestSharedOrderReadCancellationReleasesLock`：节奏等待中取消不触达官方，advisory lock 随事务释放，另一进程可立即取得。
- 新增 `TestSharedOrderReadRequiresDatabase`：协调库不可用时 `orderGET` 在发请求前失败，官方调用数 0。
- 新增 `TestNotificationDeferredClaimsAreRefunded`、`TestNotificationPermanentScopeDeadLettersImmediately`、`TestNotificationElevenLinkedOrdersPostgres`、`TestNotificationLateAssociationPostgres`；普通失败仍为 8 次领取（`TestNotificationPostgresWorker`）。

复验确认的部署前置与边界（不阻塞默认禁用合并）：

- 连接池：`collections.Open` 固定 `MaxConns=3`，启用以 0015 后单次官方 GET 会占 1 条连接最长 15 秒（约占池 1/3）。生产启用前必须复核池大小与其他长查询，否则可能排队到 15 秒超时。
- 协调器只随 `MELI_WEBHOOK_ENABLED=1` 启用：任何同时服务人工订单读取、但未开 Webhook 的进程都会绕过应用级节奏与冷却。多进程部署必须统一开启。
- 空关联/204 与暂不可用任务按 1 分钟固定间隔重调度，24 小时内约 1440 次领取、每次最多 2 次官方 GET；符合“不烧重试额度”的设计，但会持续消耗配额，后续可改为指数退避。
- 时钟：`greatest()` 合并持久化时间戳，任一进程时钟超前会拉长其他进程的等待；跨机部署需 NTP 校验。
- 真实 CBT 关联接口、seller_id/site_id 实际取值、字段变体与权限本轮 **NOT_RUN**：本机与现有主机均未持有 CBT 卖家 access token（探针需 `MELI_CBT_PROBE_TOKEN`），且验收禁止官方写入、订阅变更与生产迁移，因此只有官方文档核对与合成响应，不冒充联调。

默认禁用不变，0015 仅提交迁移文件；不执行生产迁移、部署、启用、官方订阅或写入。
