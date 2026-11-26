# WH-06 阶段 B/C 生产升级预检记录

2026-11-26（北京时间）。目标 Go API 源码为 `3601ec3bbe508f5955832de8176984fbdc1608d4`。

## 主机与授权

核对 UNA Seller 生产主机身份，`una-seller-api` 在运行，systemd 服务与环境文件路径符合部署记录。使用运行中服务的密钥、项目原生 TokenCipher 和 AAD，从只读事务读取有效的 connected/verified 授权，确认无刷新锁。Token 在主机内解密，只用于官方 GET。未刷新或轮换。

`GET /users/me` 和 CBT 店铺映射均返回 HTTP 200。六主题使用 `GET /missed_feeds`，每主题 offset 0、limit 10，未调用本地回放命令或写入事件。结果如下。

| 主题 | HTTP | `messages` | 样本 |
| --- | --- | --- | --- |
| `marketplace_orders` | 200 | null | NOT_RUN |
| `marketplace_shipments` | 200 | 数组 | 5 条，主题与应用匹配 |
| `marketplace_questions` | 200 | null | NOT_RUN |
| `marketplace_messages` | 200 | null | NOT_RUN |
| `marketplace_claims` | 200 | null | NOT_RUN |
| `marketplace_items` | 200 | 数组 | 10 条，主题与应用匹配 |

官方文档说明应用级历史查询。当前响应形状来自实际 CBT 请求。[Global Selling 通知文档](https://global-selling.mercadolibre.com/devsite/en_us/validations-cbt/receive-notifications)

## 阻断项

候选的 `decodeMissedFeedPage` 将 JSON `messages: null` 解为 nil，随后返回无效页面错误。实际空页无法结束当前扫描或将 offset 回零。四个主题复现该响应形状，均保留真实样本 NOT_RUN。商品和货件仅核对首屏记录数量及应用、主题字段。分页、留存、实际入队和公网 ACK 尚未验证。

按本轮关键检查失败停止条件，中止备份、隔离恢复、迁移预检、生产迁移、后端部署、部署后功能验收和 Nginx 实机预检。Webhook、自动补偿、历史回放继续保持关闭，官方订阅与生产业务数据未改动。

最小修复需要明确区分显式 null 与缺失字段，并为官方空页保留回零行为及回归测试。修订候选通过 CI 和隔离验收后才能继续这次生产升级。
