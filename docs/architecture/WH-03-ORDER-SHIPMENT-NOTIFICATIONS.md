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
