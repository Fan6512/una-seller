# UNA Seller V0.6 生产迁移 0013 与后端升级验收报告

> 日期：2026-11-25（北京时间） · 目标提交 `0064a2391ed0fe7aff2768053f5aa4e4b0f18e8c`
> 结论：**生产迁移与后端升级 PASS**，未回滚。`billing_snapshot` 已落库，单笔真实账单完成"官方读取 → PostgreSQL 持久化 → billing-cache 读回 → 服务重启后读回"闭环；真实净回款仍待核账，未推算任何金额。

## 1. 发布源与一致性

| 项 | 结果 |
| --- | --- |
| 升级前生产版本 | `2f679ddad896b9361096d882e71a4ad8b20ff8b4`，PID 815417，`NRestarts=0`，运行二进制 SHA-256 `6270006b06ef4d4b3c0099c05c6d8df57b0111d470fa67d60d6ea58a11b8435e` |
| 目标提交 | `0064a2391ed0fe7aff2768053f5aa4e4b0f18e8c`（PR #97 squash merge），提交时间 2026-11-25T09:57:05Z |
| 候选 HEAD | `a2570f6171679f72e5ada4298ee0ef7c29991bc5`（树 `098f6cfc…`），目标树 `c2c7ad2f…`；逐文件比对 **12 个业务/测试文件 blob 完全一致**，差异仅在 `docs/` 台账 |
| 候选 CI | [45588642857](https://github.com/Fan6512/una-seller/actions/runs/45588642857) 四项 PASS |
| 合并提交 CI | 首跑 45590276316 被后续台账提交并发取消；**重跑后 Changed Files / Go Backend / Desktop Frontend / Browser Extension 全部 success**，Go Backend 的 PG seed/restart/verify 三步实际执行 |
| 变更范围 | `2f679dd..0064a239`：12 个业务/测试文件（迁移 0013、`order_billing_snapshot.go`、账单缓存路由、桌面读取、对应测试、CI 持久化步骤）+ 3 个文档文件 |

## 2. 备份与恢复点

| 对象 | 值 |
| --- | --- |
| 数据库备份 | `/var/backups/una-seller/pre-0064a23-20261125T101017Z/app-20261125T101017Z.dump.gpg`，157,089 字节，`root:600`，SHA-256 `f017b38e8dd72f7abfae9c1cfe62a7d88ddae21371a78a12a291393721c017d9` |
| 加密方式 | `pg_dump -Fc` → GPG 对称 AES256；口令文件 `/root/.una-seller-backup.pass`（0600，仅记录路径，不记录内容） |
| 可恢复性验证 | 解密后 450,163 字节、`pg_restore -l` 542 个条目；恢复到临时 Docker `postgres:17.11-bookworm`：**73 单、指纹 `507d6fbbca72763f926e178cc774b0d1`、1 店铺、18 张 app 表全部一致**；仅 3 条 Supabase `vault.*` 对象告警（不属于业务 schema），临时容器已删除 |
| 升级前二进制 | `/var/backups/una-seller/api-20261125T101108Z-pre-0064a23/una-seller-api`，SHA-256 同上 |
| 同盘回滚副本 | `/opt/una-seller/una-seller-api.rollback-20261125T101108Z-0064a23`，三处哈希一致并写入 `SHA256SUMS` |

## 3. 迁移 0013

| 项 | 结果 |
| --- | --- |
| 迁移前状态 | 订单 73 行（120 kB）、店铺 1；`pg_locks` 未授予 0、5 分钟以上长事务 0 |
| 执行方式 | 目标提交洁净检出（`HEAD=0064a239…`，`git status` 空）内 `api migrate`；`PGOPTIONS='-c lock_timeout=3000 -c statement_timeout=120000'` |
| 结果 | 18:16:20 输出 `source collections migration applied`，无重试、无超时 |
| 结构验证 | `billing_snapshot jsonb` 可空；约束 `meli_orders_billing_snapshot_check`：`billing_snapshot IS NULL OR jsonb_typeof(billing_snapshot)='object'` |
| 历史数据 | 迁移前后均 73 单、指纹 `507d6fbbca72763f926e178cc774b0d1`，店铺 1，未受损；迁移后等待锁 0 |

## 4. 构建与原子切换

| 项 | 结果 |
| --- | --- |
| 构建 | 服务器 `/tmp/una-seller-release-0064a23` 洁净检出；Docker `golang:1.26`，`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 -trimpath -buildvcs=true` |
| 产物 | 18,003,327 字节，ELF x86-64，SHA-256 `eb6ec75c476dcddeaddcb984f2598b0cd03547f4df9c3b4b8657ca4ba6da4e84` |
| 构建元数据 | `vcs.revision=0064a239…`、`vcs.time=2026-11-25T09:57:05Z`、`vcs.modified=false` |
| 切换 | 18:16:37 同目录 `mv` 原子替换，仅 `systemctl restart una-seller-api`；PID 815417 → 864371，`NRestarts=0` |
| 哈希一致 | 构建产物、`/opt/una-seller/una-seller-api`、`/proc/864371/exe` 均为 `eb6ec75c…`；验收期重启后 PID 866058 仍为同一哈希 |
| 配置未变 | `api.env` `012a941c…`、systemd unit `68a4c433…`、Nginx `c1f3983e…` 前后一致；未改 Nginx/systemd/密钥，未重启其他服务 |

## 5. 基础回归

| 检查 | 结果 |
| --- | --- |
| `/healthz` 回环 / 公网 | 200 / 200 PASS |
| 匿名店铺、订单、C0 POST | 401 / 401 / 401 PASS |
| 已授权店铺列表 | 200，1 店，5 个 MLB/MLM/MLC/MCO/MLA 子站点 PASS |
| 订单列表 / 详情 / 同步状态 | 200 / 200 / 200 PASS（列表 20 条，`has_more=true`） |
| 订单物流接口 GET | 405（按设计仅 POST 触发官方读取）PASS |
| CBT C0 probe GET / 匿名 POST | 405 / 401 PASS |
| OTA 默认通道 | 204 + `no-store` PASS |
| OTA 预发布通道 | 首次 503（GitHub 资产拉取瞬时失败，1 分钟退避），冷却后 200 返回 `version=0.5.0`，二次命中缓存 105 ms PASS |
| 服务日志 | 部署与验收窗口内 error/panic/rate-limit 计数 0 |

## 6. 财务专项验收（单笔真实订单）

订单 `2000019508270206`（MLB / remote，`paid`，USD 158.93，货件 `48507247866`）。仅一次手动官方账单读取，未批量、未制造 429。

| 验收点 | 证据 | 结论 |
| --- | --- | --- |
| 初始账单缓存 | `GET billing-cache` 返回 `reconciliation=not_cached`、`available=false`，数据库 `billing_snapshot` 为 NULL | PASS |
| 手动读取官方账单 | `POST billing-preview` 200、`Cache-Control: no-store`、耗时 3.9 s、无 429 | PASS |
| PostgreSQL 持久化 | `billing_snapshot->>'shipment_id'=48507247866`，`checked_at=2026-11-25T10:20:47.554887872Z`，4 条 charge | PASS |
| billing-cache 读回 | 200，`checked_at` 与接口返回逐字符一致，金额与条目一致 | PASS |
| 重复读取不请求官方 | 3 次缓存 GET 前后，Nginx `billing-preview` 计数保持 16 不变；官方 POST 仅 18:20:47 一条 | PASS |
| 服务重启后读回 | 18:21:11 重启（PID 864371→866058），18:21:31 缓存 GET 仍返回同一 `checked_at` 与 4 条费用 | PASS |
| 原币与精度 | 全部 charge 为 USD；金额按平台十进制字符串保存（14.5 / 0.51 / 1.65 / 2.48），无浮点改写 | PASS |
| 物流费用 | buyer 0、seller charge 14.5、seller net 14.5（CCSI detail `75459896253`） | PASS |
| 款项状态 | `approved` / `money_release_status=pending`（待释放） | PASS |
| 官方净回款 | `net_income_usd=null`，未用税基、刊登目标或"商品减运费"局部小计冒充 | 保持待核账 |

其他：验收使用负责人现有会话刷新后的令牌，仅在内存与本机临时文件传递，未写入服务器、未输出到日志；Nginx 日志中的客户端 IP 在报告中已脱敏。

## 7. NOT_RUN 与边界

- 真实 429、401/403 授权失效、部分响应、上游超时：本轮未主动制造，现场 **NOT_RUN**（代码与测试覆盖见 #97 CI）。
- 官方净回款、实际结算回款与 Seller Center 对账：**NOT_RUN**，无结算单/回款报表数据。
- 共享货件分摊、退款订单专项：本轮只验证单笔普通订单，其余 **NOT_RUN**。
- 未执行生产回滚（无阻塞故障）；未删除任何业务数据，未恢复整库覆盖。
- 未发布客户端、未创建 Tag/Release、未改官方订阅配置、未接入第三方物流 API、未做真实商品刊登。
- 生产库现有 1 条真实账单快照（上述订单），属正常业务数据，保留。

## 8. V0.6.0 签名测试版发布准备

| 项 | 状态 |
| --- | --- |
| 签名环境 | `desktop-production` 存在，含 `required_reviewers` 与分支保护；环境密钥 `TAURI_SIGNING_PRIVATE_KEY`、`TAURI_SIGNING_PRIVATE_KEY_PASSWORD`，变量 `TAURI_UPDATER_PUBLIC_KEY` 均在位 |
| 版本文件 | 当前仍为 `0.5.0`：`apps/desktop/package.json`、`src-tauri/tauri.conf.json`、`src-tauri/Cargo.toml`（`Cargo.lock`/`package-lock.json` 同步） |
| 发布路径 | `windows-desktop.yml` 支持 `workflow_dispatch`（check_only / signed_smoke）与 `desktop-test-v*` 标签构建，标签触发需 `desktop-production` 审批 |
| 待执行 | 版本号提升 PR → 负责人批准 → 打 `desktop-test-v0.6.0` → 审批签名构建。本轮未改版本、未打标签、未创建 Release |

## 9. 负责人客户端人工验证清单

1. 启动 V0.6 客户端，确认已登录且店铺 `ZYC` 状态正常（服务端已恢复 `connected`）。
2. 打开订单 `2000019508270206` 详情：应自动显示"正在读取已保存账单"，随后展示 4 条 USD 费用与"账单最近成功保存 2026-11-25 18:20"，且不触发官方请求。
3. 点击"读取官方账单"：应显示 CCSI 14.50、CVMPI 0.51、CVMPCB 1.65、CVMLCB 2.48、物流净成本 14.50、款项"待释放"；不得出现"净回款"数字。
4. 关闭详情再打开、以及完全退出客户端重进：账单仍可读取，时间戳不变，官方请求计数不增加。
5. 切换到另一笔无快照订单：应显示"待核账/未缓存"，手动读取后才出现账单。
6. 断网或超时后重试：应保留上次成功账单并提示，不得清空。
7. 若触发 429：按钮应显示冷却剩余秒数且不自动重试（本轮未制造，属观察项）。
8. 订单列表状态去重（#96）：每单只显示一个综合状态；详情仍单独显示"平台物流状态"。
9. 物流查询（#96）：Track718 `https://www.track718.com/zh-CN`、17TRACK `https://www.17track.net/zh-cn`；"独立窗口查询"应打开客户端内窗口，非 Tauri 环境应明确报错而不是静默开浏览器。
10. 采购成本为空时应显示"待补录"，实际利润显示"待核算"，不得出现推算值。
