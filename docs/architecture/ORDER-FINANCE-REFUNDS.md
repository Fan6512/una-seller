# UNA-SELLER CBT 订单状态与发货 SLA（PR #48）

状态：Draft（2026-11-21）。本 PR 仅实现单笔订单状态刷新、退款状态中文化与发货 SLA。财务设置/人民币汇率/利润计算整体推迟到经过官方账单核验的后续 PR；不混入不可用入口或 0006 财务迁移。

## 订单状态
- 用户手动更新一笔 Mercado Libre 订单，官方 `GET /marketplace/orders/{id}`。
- 事务内重新验证 owner、店铺、站点、物流类型、子卖家，按 `last_updated` 单调更新；保留 1688 采购关联和订单同步游标。
- `partially_refunded` 为「部分退款」；`refunded` 为「已退款」；`cancelled` 仅为「已取消」，不凭取消判断已退钱。
- 当前只展示销售金额（USD）；不把销售额当净回款。费用、支付、运费、税金、退款金额及结汇汇率留待官方订单级账单单独核对。

## 官方发货截止
- 以官方 `GET /marketplace/shipments/{id}/sla` 的 `expected_date` 为截止时间；识别 `on_time/delayed/early/insuficient_info`，携带 `x-format-new: true`。
- 不查已取消/申请取消和 Fulfillment；无 Shipment ID、不返回日期时不猜测。
- 新表 `app.meli_order_dispatch_sla` 通过 0006 显式迁移。按店铺/站点/物流类型/子卖家/订单+Shipment ID 隔离，换 Shipment 不继承旧 SLA。
- 手动最多串行读取本页 20 笔；遇平台限流/不可用停止，反馈成功/失败/未尝试笔数，不进行后台自动同步。
- 列表显示倒计时：≥24 小时按天，1～24 小时按小时，<1 小时为「不足1小时」，截止后「截止已过」。不足 24 小时或平台明确 `delayed` 时标红；本地截止已过不等于平台判定绩效违规。
- 客户端每 15 分钟及窗口重新聚焦时本地更新，不轮询 API。订单详情保留 `Asia/Shanghai` 绝对截止时间，数据库以 UTC 存储。

## 验证和部署边界
- 已对 MLB 单笔使用现有 owner 授权真实只读探测返回 HTTP 200、`on_time`、`expected_date=2026-11-18T01:15:42+08:00`。Seller Center 同单对照未完成；尚非生产端到端验收。
- #46 采购关联与本 PR 的读模型重叠。#46 先完成审查/合并，再基于新 main 处理 #48 重复字段与 UI，保留采购筛选、图片 JOIN、SLA 和退款刷新。
- 发布顺序：先显式执行 0006 SLA 迁移，再部署对应 Go API，最后构建/验收 Windows 客户端。不得直接对当前生产启用新二进制。

参考：https://global-selling.mercadolibre.com/devsite/en_us/support-for-integrators/manage-shipments
