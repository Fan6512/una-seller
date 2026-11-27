# OPS-01 通知队列审核记录

[Draft PR #99](https://github.com/Fan6512/una-seller/pull/99) 接续 `db56e44f`，范围为 WH-01 和 WH-02。订单和物流业务处理器由 WH-03 接入。当前处理器注册表为空，领取 SQL 仅匹配已注册主题。复审修订后，生产启动启用前必须注册当前全部接收主题的真实处理器，缺少任一处理器直接拒绝启动启用，避免接收后无人消费。隔离测试仍可构造未注册主题的积压。

## 官方来源与回调合同

- 2026-11-25 核对 Mercado Libre Global Selling 开发者文档的索引内容。页面标示更新于 2026-09-14，列出来源 IPv4 清单。另一份官方页面列出完整清单，包括末尾的 `35.245.20.104` 和 `35.186.182.146`。直接打开部分文档 URL 返回 403，本次核对采用官方页面的搜索索引。启用前应从官方页面重新确认完整清单和变更日期。 [接收通知与来源 IP](https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/receive-notifications) [完整来源清单](https://global-selling.mercadolibre.com/devsite/manage-sales-global-selling/receive-notifications)
- 官方采用 HTTP POST，要求 500 ms 内返回 HTTP 200。失败会重投，文档给出一小时窗口。订单和物流通知提供资源标识，接收后使用卖家 OAuth 对官方固定 API 路径发起 GET。 [订单与物流通知样例](https://global-selling.mercadolibre.com/devsite/receive-notifications)
- 官方要求验证通知来源和资源 URL。本次查阅的 Global Selling 文档未披露适用于这两类通知的签名或验签算法。`application_id` 只用于应用范围匹配。`X-Una-Webhook-Ingress` 是内部代理凭据。公开入口须根据可信连接来源匹配官方清单，移除客户端供给的同名头，验证通过后由代理注入私密值，Go 继续监听回环地址。若入口前置 CDN 或负载均衡器，先核对可信代理链和真实连接来源提取规则。代理层请求速率、连接并发和日志需随启用审批验证。 [官方授权安全指南](https://global-selling.mercadolibre.com/devsite/messaging-after-sale-global-selling/authorization-and-token-best-practices)

## 队列参数与故障诊断

- ACK 在 PostgreSQL 同步提交后返回，350 ms 预算涵盖读请求体及持久化。队列满、写入失败或并发槽位耗尽返回 HTTP 503。重复投递保留 HTTP 200，但仍需通过当前授权校验。未知或歧义卖家返回 403；已知唯一但断开、授权待恢复或映射未验证返回 503，不写入未经确认的回执/任务。官方重投窗口有限，超过窗口及映射完全缺失由 WH-06 漏件补偿处理。
- 投递凭据去重键含店铺、所有者、应用、通知用户、主题和资源。同一店铺父子账号对同一资源的投递合并为一个活跃任务。工作期间收到新投递时，成功提交后保留下一轮读取。已转死信的任务保留故障记录，后续不同投递可建立新任务。
- 单次领取采用 PostgreSQL `FOR UPDATE SKIP LOCKED`，租约为 45 秒。Worker 给授权、读取及业务事务最多 30 秒。fencing token 保护本地业务事务和任务状态迁移。执行前核对有效店铺连接及已验证身份，提交前重新核对通知用户、官方返回卖家、主题和资源 ID。
- 每条任务最多领取 8 次，失败后从 5 秒指数退避。上游 429 按店铺及应用对所有主题冷却，至少 60 秒并服从更长的 `Retry-After`。冷却和租约存于 PostgreSQL，进程启动后可继续检查。已领取的调用可在冷却提交前处于飞行中。第八次失败或最后一次租约过期转死信。
- 投递凭据和任务各有 20,000 条全局容量。凭据保留 7 天，死信保留 30 天。活跃任务保留到成功或进入死信，不静默删除未消费通知。启动前要求当前两个可接收主题均注册处理器；异常遗留的未注册任务、全局回执或任务容量达到 80% 均输出显式告警。应用级回执计数与全局容量计数分开。每分钟最多清理每类 500 条到期记录，并记录总量、待处理、领取中、未注册、冷却、过期租约、死信、最老任务等待时间及脱敏故障分类。
- 复审后可用下列只读查询定位积压来源。生产启用验收建议按 `receipts`、`jobs_total` 容量占比、`oldest_pending`、死信增加、过期租约和入口 503 告警。死信补偿及正式事件回放属于 WH-06。

```sql
SELECT application_id, topic, last_failure,
       count(*) FILTER (WHERE dead_at IS NULL) AS open_jobs,
       count(*) FILTER (WHERE dead_at IS NOT NULL) AS dead_jobs,
       min(pending_since) FILTER (WHERE dead_at IS NULL) AS oldest_open
FROM app.meli_notification_jobs
GROUP BY application_id, topic, last_failure
ORDER BY oldest_open NULLS LAST;
```

## 调度与复审修订

- 当前刻意采用单 Worker 串行执行、每进程每秒最多一次领取，慢任务会进一步降低吞吐。20,000 条全积压即使每条即时处理也至少需要约 5.6 小时；不承诺实时清空。WH-03 建立与手动同步共享的官方 API 限流、冷却及互斥后，再根据实测评估调度与并发，不预领取批量租约。
- 官方 `_id` 作为不透明字符串处理，仍受 16 KiB 请求体限制，只存 SHA-256；不限制字符集，不放宽 resource、application_id、user_id 或 topic 校验。
- ACK 统计使用 nearest-rank 百分位；n=96 时 p99 恰好等于最大值（第 96 个样本），不是独立尾部样本。原始压测数据是 `fdcd2ae9` 历史证据，修订后须重新执行 CI 和隔离库验收。

## 验证证据（fdcd2ae9 历史）

本机隔离 PostgreSQL 实例运行于回环地址。使用实际 HTTP 客户端并行 6 路发送请求，表中每组 96 个有效请求，计时含本地 HTTP、持久化和响应头。

| 场景 | p50 | p95 | p99 | 最大值 |
| --- | --- | --- | --- | --- |
| 同一通知并发重投 | 0.571 ms | 60.696 ms | 222.961 ms | 222.961 ms |
| 同一资源不同通知 | 3.824 ms | 6.893 ms | 9.718 ms | 9.718 ms |

数据库写锁阻塞时约 352 ms 返回 HTTP 503，未出现部分投递记录。独立停止 PostgreSQL 时返回 HTTP 503。数据库重启前的回执、租约和冷却记录保留，过期租约被新进程重新领取。测试覆盖跨店同资源隔离、竞争领取、失效 token 拒绝提交、执行期间授权撤销、指数退避、共享 429 冷却、死信和保留期限。完整 CI 运行记录见 PR。

公网来源、边缘限流及真实通知端到端 ACK 待 WH-06 联调。当前 `MELI_WEBHOOK_ENABLED` 默认为 `0`。
