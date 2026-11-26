# WH-06 阶段 B/C 生产升级报告

2026-11-26，北京时间。生产部署固定为 `b1dd92ffcf4fd676c379f135f51389c26bd0562f`（#111）。台账提交 `ba2f9d47` 只含文档。核对 UNA Seller 主机、systemd 单元、运行路径和服务配置后，15:46 完成数据库迁移，15:47 完成 Go API 切换。

## CI 与官方只读验证

- 合并后 CI 首轮取消。同一 SHA 第 2 次运行成功，Changed Files 和 Go Backend 通过，桌面和扩展按变更范围跳过。PG 定向回归验证显式 null 页将 offset 从 8 重置为 0，缺失字段保持拒绝。[CI 45776485335](https://github.com/Fan6512/una-seller/actions/runs/45776485335/attempts/2)
- 使用现有有效 Token，身份、CBT 店铺映射和六主题首屏请求均为 200。使用固定提交的原生解析器直接读取六个真实响应。货件 5 条、商品 10 条全部解析通过。订单、问答、消息、纠纷的显式 null 均解析为零事件，非空样本 **NOT_RUN**。未刷新 Token，未运行生产回放。真实分页和留存仍待验证。

## 备份、恢复与迁移

- 复用既有加密备份程序保存独立 app schema 快照。归档 `app-20261126T074438664942612Z.dump.gpg`，135170 字节，SHA-256 `3b211c13234b189285ae2b48b50c85ff176c7d3886a1d36babe40258eb71d1cb`。备份、旧二进制、站点配置、systemd 单元及验收回执保存在 `/var/backups/una-seller/wh06-b1dd92ff.yzPPz9jS`。
- 在无网络 PG17.11 容器恢复同一归档。用部署候选的原生 `migrate` 连续执行两次，确认 9 张通知相关表启用 RLS、0020 主键为 `(application_id, topic, credential_store_id)`、来源约束唯一。恢复容器已清理。
- 生产 PG17.11 于 15:46:22 完成原生迁移。迁移与部署后均为 73 条订单、27 条货件快照、3 条账单快照、207 条已刊登商品、1 条店铺连接。部署后该连接仍为 connected/verified、未过期、无刷新锁。收件箱、通知作业和历史回放断点均为 0 条。

## 后端切换与接口

从干净固定提交用 Go 1.26.0 构建 Linux amd64、CGO 禁用、`-trimpath` 二进制。构建信息未嵌入 Git revision，本次用构建前的干净 checkout SHA 和产物校验和固定来源。构建产物、部署文件及运行进程三处 SHA-256 一致。

`e524d7c9737258c15f05ee9f1eb71b3dfb56148e13c44fbfaa27547eb1cfcf80`

原子替换后只重启 `una-seller-api`。服务 active，单进程监听 `127.0.0.1:8787`，自动重启数为 0。以下回环与公网结果均与升级前一致。

| 检查 | HTTP |
| --- | --- |
| 健康检查 | 200 |
| 店铺授权、订单、商品、已刊登商品及账单缓存的匿名访问 | 401 |
| 货件路由 GET 方法检查 | 405 |
| OTA 默认渠道 | 204 |
| 禁用的公网 Webhook 路由 POST | 404 |

带 Supabase 客户端会话的端到端业务接口验收 **NOT_RUN**，本轮核验覆盖生产数据快照、官方授权 GET、路由与匿名鉴权边界。OTA 204 表示默认渠道无更新清单，下载安装 **NOT_RUN**。

## 入口预检与回滚

- 在生产主机用临时回环 Nginx 配置加载候选 Webhook 模板，校验官方 Global Selling 文档中的 92 个唯一 IPv4 地址。来自未授权 TCP 对端的 POST 即使伪造转发头和私有入口头，仍返回 403。生产配置 `nginx -t` 通过。临时实例已退出，真实官方来源投递和公网 ACK 压测 **NOT_RUN**。[Global Selling 通知文档](https://global-selling.mercadolibre.com/devsite/manage-sales-global-selling/receive-notifications)
- 回滚件已校验。旧 API SHA-256 为 `eb6ec75c476dcddeaddcb984f2598b0cd03547f4df9c3b4b8657ca4ba6da4e84`。如需回滚 API，从备份目录校验旧二进制，在同目录暂存后原子换回，重启该 systemd 单元并复验健康。数据库恢复须单独决策，避免覆盖备份后业务数据。本次未触发回滚。
- API 环境、systemd 单元及 Nginx 站点配置的哈希与切换前一致。Webhook、自动补偿和历史回放保持关闭。没有 Nginx reload、官方订阅变更或官方业务写入。

## 阶段 D1 生产接收器启用

2026-11-26 16:00（北京时间），在已上线的 `b1dd92ff` 启用仅 `marketplace_orders,marketplace_shipments` 的接收器和 Worker。自动补偿、历史回放及其他四主题关闭。此次只修改 API 环境与 Nginx，重启 UNA Seller API 并 reload Nginx。未重新构建或迁移。

- 官方 Global Selling 公布的 92 个 IPv4 来源进入 Nginx ACL。以实际 TCP 对端为准，拒绝其他来源，每源限流 20 次/秒、突发 40、超限返回 503。[Global Selling 通知文档](https://global-selling.mercadolibre.com/devsite/manage-sales-global-selling/receive-notifications)
- 主机内随机生成入口令牌，API 环境和 Nginx 入口 include 均由 root 以 0600 持久保存。后续维护复用原值。仅在丢失、泄露或明确批准轮换时，在服务器内部同期更新两处。
- 公网健康 200。公网未授权 POST 与伪造来源/入口头 POST 均 403。回环缺令牌 403，正确令牌的非法报文 400。运行二进制哈希保持一致，服务 active，自动重启为 0。启用回执与回滚件存于 `/var/backups/una-seller/wh06-d1.9rCqtGn5`。需暂停时关闭 API 开关、撤回 Nginx location 与限流区，保留现有令牌供恢复。
- 官方应用管理器的 UNA Seller 应用 ID 已与生产配置匹配。编辑订阅时触发账号二次验证，等待负责人通过验证。官方订阅未变更，生产接收器继续保持就绪。
- 启用后只读观察收件箱、作业、回放及发现断点均为 0，订单计数 73。真实投递、ACK、持久化、Worker 消费、去重及业务更新验收 **NOT_RUN**。后续按日常真实业务样本验证。
