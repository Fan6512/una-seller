# UNA Seller PR #95 生产后端升级部署报告

> 日期：2026-11-25（北京时间） · 固定业务提交 `2f679ddad896b9361096d882e71a4ad8b20ff8b4`
> 结论：**生产后端 PASS**，服务健康、核心接口与 OTA 双通道无回退、无回滚。Windows GUI / V0.6 客户端实机验收待负责人人工执行；未发布 V0.6、未创建新 Tag。真实 429 与物流刷新失败场景本轮未触发，标记 NOT_RUN。

## 1. 发布源、CI 与一致性

| 项 | 结果 |
| --- | --- |
| 目标提交 | `2f679ddad896b9361096d882e71a4ad8b20ff8b4`（PR #95 merge commit，父提交 `e3ff970`），MERGED 2026-11-25 06:58:09Z |
| 升级前生产版本 | `86fcb58468e353971a9a40a40b18a8aeb831f16c`，启动 2026-11-25 13:56:08 CST，PID 784691，`NRestarts=0`；运行二进制 SHA-256 `dcbffb41…93bdd4a`，`vcs.modified=false` |
| 候选 CI | [45558640026](https://github.com/Fan6512/una-seller/actions/runs/45558640026)（PR head `c1397cd`）success：`Changed Files`、`Go Backend`、`Desktop Frontend` PASS，`Browser Extension` 按范围跳过 |
| 候选/目标一致性 | `c1397cd..2f679dd` 仅 `docs/product/FEATURE-STATUS.md` 不同；**backend 代码树完全一致**，候选 PASS 可覆盖目标业务代码 |
| 合并提交 CI | 首次 run 45564522734 被后续文档推送并发取消；重跑 **success**：`Changed Files`、`Go Backend`（vet/test/build）、`Desktop Frontend` PASS，Extension 跳过 |
| 目标之后的 main | `d320b48`、`7c88829` 均为文档提交，未纳入本次部署 |
| 变更范围 | `86fcb58..2f679dd`：Go 后端 3 文件（`orders.go` 新增候选列表路由、`order_shipment.go` 候选查询、`order_shipment_test.go`），桌面端 3 文件，文档 2 文件。**无 SQL 迁移文件，未执行 migrate** |

## 2. 升级前状态、恢复点与备份

| 对象 | 值 |
| --- | --- |
| 服务与磁盘 | `una-seller-api` active，PID 784691，`NRestarts=0`；`/`、`/var/backups` 可用 47G；回环/公网 healthz 均 200 |
| 复用数据库恢复点 | `/var/backups/una-seller/postgres/app-20261125T025039463305311Z.dump.gpg`，SHA-256 `e700ab0107a531349b9a15e888c796bb60d9eddea94fadedb5a0629c7d896066`，沿用 #94 已验证恢复点；未重复恢复演练 |
| 本轮升级前备份 | `/var/backups/una-seller/api-20261125T072016Z-pre-2f679dd/una-seller-api`，SHA-256 `dcbffb41eb3be4fbda3ecccb10657e15d7a3a8eb7fea8c4729495275e93bdd4a` |
| 同盘回滚副本 | `/opt/una-seller/una-seller-api.rollback-20261125T072016Z-2f679dd`，SHA-256 同上；备份目录 `SHA256SUMS` 记录三处哈希并已复算一致 |

回滚方式：核对备份哈希后同目录原子换回原路径，仅重启 `una-seller-api`，复验 healthz/鉴权/订单/OTA。本轮未触发。

## 3. 洁净构建与原子切换

| 项 | 结果 |
| --- | --- |
| 检出 | 服务器 `/tmp/una-seller-release-2f679dd`，`git init` + `git fetch --depth 1 origin 2f679dd…` + `checkout --detach`；`HEAD` 等于目标 SHA，`git status --porcelain` 为空。首次 fetch 因网络瞬时卡顿超时终止（未触碰生产），重试完成 |
| 构建 | Docker `golang:1.26`，`GOOS=linux GOARCH=amd64 CGO_ENABLED=0 -trimpath`，`GOPROXY=https://goproxy.cn,direct` |
| 产物 | ELF 64-bit LSB x86-64，静态链接；SHA-256 `6270006b06ef4d4b3c0099c05c6d8df57b0111d470fa67d60d6ea58a11b8435e` |
| 构建元数据 | `vcs.revision=2f679ddad896b9361096d882e71a4ad8b20ff8b4`、`vcs.time=2026-11-25T06:58:09Z`、`vcs.modified=false` |
| 切换 | 2026-11-25 07:33:11Z 同目录 `mv` 原子替换，仅 `systemctl restart una-seller-api`；PID 784691 → 815417，`NRestarts=0`，启动日志仅常规配置行 |
| 三处哈希一致 | 构建产物、`/opt/una-seller/una-seller-api`、运行进程 `/proc/815417/exe` 均为 `6270006b…8435e` |
| 保持不变 | `api.env` `012a941c…`、`systemd unit` `68a4c433…`、Nginx vhost `c1f3983e…` 前后一致；未改 Nginx/systemd/密钥，未重启其他服务（vaultwarden 未动） |

## 4. 基础生产回归

| 检查 | 结果 |
| --- | --- |
| `/healthz` 回环 / 公网 | 200 / 200 PASS |
| 匿名 `/api/v1/collections`、`/api/v1/stores` | 401 / 401 PASS |
| 匿名店铺范围内订单列表 | 401 PASS（`/api/v1/orders` 本身无此路由返回 404，符合预期） |
| C0 `GET .../publication-mode/probe` / 匿名 `POST` | 405 / 401 PASS |
| 已授权店铺列表、订单列表、订单详情、同步状态 | 200 / 200 / 200 / 200 PASS |
| 单笔官方账单 `POST .../orders/billing-preview` | 200、`items=1`、`Cache-Control: no-store` PASS；无 429 |
| OTA 默认通道 `X-Una-Include-Prerelease: 0` | 204 + `Cache-Control: no-store` PASS |
| OTA 测试通道 `X-Una-Include-Prerelease: 1` | 200 + `Cache-Control: no-store`，`version=0.5.0` PASS |

订单/账单读取使用客户端现有会话，Bearer 仅在进程内存中传递，未落盘、未输出、未记录；未输出原始敏感响应或买家信息。

## 5. #95 专项验收（单店 MLB / remote，未做批量官方查询）

| 验收点 | 现场/证据 | 结论 |
| --- | --- | --- |
| 单笔订单刷新同步读取官方物流 | 订单 `2000018701509380`：刷新前 `shipment_snapshot` 为空 → `POST .../orders/{id}/refresh` 200 → `POST .../orders/{id}/shipment` 200；官方状态 `delivered`，回读详情持久化同一 `checked_at` | **PASS** |
| 已知物流滞后订单与官方状态比对 | 上述订单交易状态 `paid` 但物流快照缺失；官方返回 `delivered`，比对后落库一致 | **PASS**（1 笔） |
| 手动同步后补充陈旧物流 | `GET .../orders/shipment-refresh-candidates` 200，返回 10 笔陈旧/未匹配订单；刷新其中 1 笔后该订单从候选列表消失 | **PASS**（服务端；GUI 联动待人工） |
| 单轮最多 10 笔 | 现场候选响应恰好 10 条且 `has_more=true`；定点测试用 20 条 fixture 断言 `len=10, has_more=true` | **PASS** |
| 15 分钟内不重复刷新 | 刚刷新订单立即不再出现在候选；定点测试中 `checked_at` 相差 1 分钟的 fixture 被排除 | **PASS** |
| 已匹配终态不参与刷新 | 该店 19 笔终态快照订单均未进入候选；定点测试终态 `delivered`/`cancelled` fixture 被排除 | **PASS** |
| 刷新失败保留订单与物流缓存 | 未注入上游故障；定点测试覆盖失败不覆盖已有快照 | 代码/测试 **PASS**，现场 **NOT_RUN** |
| 429 遵守冷却、不连续重试 | 本轮无 429，部署后日志无 rate-limited 记录；定点测试确认上游 429 在读取详情前停止、概要读取不绕过冷却，客户端 `shouldStopLogisticsBatch` 对 429/502/504/断网停止 | 代码/测试 **PASS**，现场 **NOT_RUN** |

定点测试：服务器目标提交洁净目录下 `go test ./internal/meli -run '^(TestShipmentRefreshCandidatesScopeAndFreshness|TestRefreshOrderShipmentOfficialSnapshotAndScope)$' -count=1` PASS；合并提交 CI 的 Desktop Frontend 项含 `tests/ordersScope.test.mjs` PASS（本地临时 worktree 缺 `node_modules`，未重复执行该项）。本轮未主动制造限流或上游故障，未批量查询官方账单。

## 6. 边界与未执行

- 无新增数据库迁移，未执行 `migrate`；未重复完整数据库恢复演练。
- 未修改 Nginx、systemd、环境变量或密钥；未重启其他服务。
- 未发布 V0.6 客户端、未创建新 Tag、未扩大到 V0.7 开发。
- 未批量查询官方账单；未泄露 Token、买家信息或原始敏感响应。
- **生产后端 PASS ≠ 整个 V0.6 完成**：Windows GUI、跨版本 OTA、单笔联动提示、第三方物流窗口与时间显示仍待负责人人工验收。

## 7. 负责人客户端人工验证点

1. 单笔订单「刷新订单」后，提示应包含官方物流概要已更新；若物流读取失败，提示物流未更新但订单数据保留。
2. 手动「同步订单」后，提示应追加物流概要更新笔数；陈旧非终态订单被补充刷新，单轮不超过 10 笔，仍有更多时提示后续再同步。
3. 对已知物流滞后订单，界面显示的官方状态与后台一致；15 分钟内重复同步不应再次刷新同一记录；已签收/取消终态不反复刷新。
4. 物流链接在独立窗口/外部浏览器打开；运单复制、承运商、北京时间显示与状态时间轴符合验收口径。
5. 若出现 429/超时：界面显示冷却提示，不连续重试、不绕过限流；已有订单与物流缓存不丢失。
6. 既有 V0.5 项：测试版开关关闭时正式通道不收到 0.5.0；开启后可发现/安装 0.5.0；登录、店铺授权、工作稿、已刊登、订单/财务数据不丢失；账单单笔预览可用。
