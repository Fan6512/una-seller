# WH-04 客服通知只读处理（候选）
 
基线：WH-03 #100 已合并。保持 `MELI_WEBHOOK_ENABLED=0`，不得修改官方订阅。候选实现，尚未独立编译或实际联调。

## 资源、模块和权限

- `marketplace_questions` → GET `/marketplace/questions/{numericID}`；`marketplace_messages` → GET `/marketplace/messages/{opaqueID}`，只接受严格路径、有限字母数字 ID；不接受任意 URL、查询参数、附件、pack 路径。参见 [官方问答](https://global-selling.mercadolibre.com/devsite/manage-questions-answers-global-selling)、[官方消息](https://global-selling.mercadolibre.com/devsite/messaging-after-sale-global-selling)、[官方通知](https://global-selling.mercadolibre.com/devsite/receive-notifications)。
- 0016 扩展现有队列约束，新增私有 RLS 客服投影，不改已有订单/货件、采购/账单、游标。客服投影以 store/child/site/topic/id 唯一；商品或 pack ID 只作关联，不当作订单号。只保留白名单状态、内容、翻译、答复、方向和时间；不保存原始 JSON、买家身份字段、附件或 Token。文本本身可能含个人信息，未来 API 必须严格 owner 校验、no-store、纯文本渲染、脱敏日志与保留策略，不得发送第三方 AI。
- 父级 seller 仅在官方站点和映射能唯一确定子账号时接受；同一子账号的多个物流模式不复制客服记录；不同子卖家不合并。消息要求响应唯一 ID、seller/pack resources 和一方为对应卖家/父账号，其他已知子账号不能冒充买家。提交前再次验证映射和 worker fence。按本地读取开始时间拒绝旧响应，不假定官方有 last_updated；不声称可完全抵御上游自身缓存回退。
- 复用 orderGET 后台优先级、应用级持久化限流、事务互斥、延迟及 429 规则；401/403 进入授权退避，不自动发消息、答复、标记已读、退款、修改商品。Model 6 文档明确限制问答及消息 API，需真实 CBT 权限证据，禁止绕过。
- 仅实现接收/回读/本地投影候选；历史会话读取、未读/已读语义、回复、附件、UI、翻译服务均不在本 PR。官方消息 GET 是否隐式改变已读状态必须取证，不能假设 GET 无副作用；若存在隐式已读且无只读方式，保持该主题不启用并调整合同。

## 门禁

CI 全量 Go/PG，0016 新装/重复/从 0014/0015 升级，接收四主题并拒绝非法路径，重复/乱序/跨租户、歧义映射、授权撤销、消息参与人冲突、状态受限清除正文，429 双向保护，真实 PG 重启。真实 CBT 仅现有授权 GET 证据，需独立审核确认不标记已读；任何官方写入、生产迁移、订阅启用须另批。

后续 WH-05 复用同一队列与协调器，不把客服、纠纷、商品混成一个业务表。OPS-02 暂不暴露客服文本，先按授权范围返回运营摘要；OPS-03/WH-06 补偿与生产接入另行验收。
