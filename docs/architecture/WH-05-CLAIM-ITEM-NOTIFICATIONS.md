# WH-05 投诉及商品通知（候选，依赖 WH-04 #101）

- 官方通知资源 `/marketplace/claims/{id}` → 当前详细读取文档 `GET /marketplace/v2/claims/{id}`；`/marketplace/items/{siteItemId}` → `GET /marketplace/items/{id}`。仅允许 numeric claim ID 和 3 字母站点前缀+数字 item ID，不接受 CBT 父商品或任意路径。参见 [通知](https://global-selling.mercadolibre.com/devsite/receive-notifications)、[投诉](https://global-selling.mercadolibre.com/devsite/manage-claims)、[商品](https://global-selling.mercadolibre.com/devsite/global-listing)。
- 0017 在 0016 之上扩展六主题队列约束，调整 0016 的重复应用保护以免回装旧窄约束；新增独立 RLS 投诉、商品状态投影。归属由官方 claim seller player / item seller+site 和当前父子映射证明；多卖家/缺失或歧义拒绝。投诉 resource_type 与 related_id 分开，不把支付/货件/pack 猜成订单。状态及官方更新时间保护，不保存玩家信息、原始 JSON 或买家消息。商品数量缺失保留已知值，0 不是缺失。
- 复用共享读取协调器和队列 fencing。无官方回复、申诉、退款、退货、暂停、改价、改库存或其他写入。不把状态投影直接覆盖人工同步的刊登详情、编辑稿；OPS-02 负责展示通知摘要，真实全量刊登补偿以后处理。
- Model 6 投诉 API 可能被拒；官方文档的通知 URL 与详情 URL 版本不同，真实 GET 权限/字段及无副作用须验证。接收六主题并注册真实候选 handler，默认禁用。
- 回归：路径白名单/重复迁移/持久入队、商品和投诉投影、旧响应、跨站点、授权撤销、刊登缓存保护。新 HEAD CI/PG、真实 CBT 只读取证、权限失败恢复、范围与 fencing 端到端待验收。先审核/合并 #101，再将本 PR rebase 或移至 main，重新运行 CI。禁止生产迁移/部署/订阅。

## 2026-11-26 收口复审

#101 已合并，WH-05 已重基到 main。真实 CBT 只读证据据本地报告：shipment 投诉为 sender/receiver，14 条 closed、0 条 opened；站点商品 200，含 price/status/sub_status/quantity。已验证的父子映射独立于通知用户识别唯一卖家参与人；仅提取该参与人的动作名（格式白名单，不是执行授权），只有 mandatory=true 的有效期限参与最早强制期限。真实开放投诉的期限和动作仍 NOT_RUN。

商品投影省略/null 的 sub_status 保留已知值，明确空数组清空；价格和数量缺失保留，明确零值更新。保留值属于上次已知值，并不代表所有字段均在同一时间重新取得；不得将缺失数据或无币种的价格投影直接解释成实时可售数量/结算金额。新增 PG 测试覆盖省略字段、明确清空、零值、旧快照和超长审核状态。当前 HEAD CI 待验收。Webhook 默认关闭，不合并、不迁移、不部署、不启用。
