# UNA Seller #94 生产后端升级部署报告

> 日期：2026-11-25（北京时间） · 固定业务提交 `86fcb58468e353971a9a40a40b18a8aeb831f16c`
> 结论：升级完成，核心服务健康，接口与 OTA 双通道无回退。未回滚、未发布客户端、未创建新 Tag。客户端实机与功能验收由负责人人工执行。

## 1. 发布源与 CI

| 项 | 结果 |
| --- | --- |
| 目标提交 | `86fcb58468e353971a9a40a40b18a8aeb831f16c`（PR #94，父提交 `ac533fe`），MERGED 2026-11-25 05:41:26Z |
| 当前生产版本（升级前） | `1bb501920867f8d21b00f7130b94a6184262eb72`，启动于 2026-11-25 10:51:41 CST，PID 734130，`NRestarts=0` |
| main CI | 目标提交首次 run 45555320846 被后续文档推送并发取消；已重跑 attempt 2 **success**：`Go Backend` PASS（vet/test/build）、`Desktop Frontend` PASS、`Changed Files` PASS、`Browser Extension` 跳过 |
| 变更范围 | `1bb5019..86fcb58` 共 21 文件；Go 后端 5 文件（账单预览 Retry-After 暴露、限流诊断固定标签、订单读取日志），其余为桌面端/文档。**无 SQL 迁移文件，未执行 migrate** |
| 后续提交 | `89b3bcc` 仅文档，不含业务代码，未部署 |

## 2. 恢复点与备份

| 对象 | 值 |
| --- | --- |
| 复用数据库恢复点 | `/var/backups/una-seller/postgres/app-20261125T025039463305311Z.dump.gpg`，134,242 字节，SHA-256 `e700ab0107a531349b9a15e888c796bb60d9eddea94fadedb5a0629c7d896066`，`root:600`；已验证恢复点，本轮未重复完整恢复演练 |
| 上一版二进制备份（既有） | `/var/backups/una-seller/api-v05-pre-20261125T105140/una-seller-api`，SHA-256 `53ea0ae7f584ae9af6790947bcfa4e43d278a035562c12888d4681e8626c61c5` |
| 本轮升级前二进制备份 | `/var/backups/una-seller/api-20261125T055354Z-pre-86fcb58/una-seller-api`，SHA-256 `0140b549148586bfe7a6796cc0352fdf7b8228114284a7d59aa366ad7c19c30e`；同目录 `SHA256SUMS` 含三处哈希 |
| 同盘回滚副本 | `/opt/una-seller/una-seller-api.rollback-20261125T055354Z-86fcb58`，SHA-256 同上 |

回滚方式：核对备份哈希后，同目录原子换回原路径，仅重启 `una-seller-api`，复验 healthz/鉴权/订单/OTA。本轮未触发。

## 3. 洁净构建与切换

| 项 | 结果 |
| --- | --- |
| 检出 | 服务器 `/tmp/una-seller-release-86fcb58`，`git fetch --depth 1 origin 86fcb58`，`checkout --detach`，`HEAD` 一致，`git status --porcelain` 为空 |
| 构建 | Docker `golang:1.26`，`GOOS=linux GOARCH=amd64 CGO_ENABLED=0`，`-trimpath`，`GOPROXY=https://goproxy.cn,direct` |
| 产物 | ELF 64-bit LSB x86-64，静态链接；SHA-256 `dcbffb41eb3be4fbda3ecccb10657e15d7a3a8eb7fea8c4729495275e93bdd4a` |
| 构建元数据 | `vcs.revision=86fcb58468e353971a9a40a40b18a8aeb831f16c`、`vcs.time=2026-11-25T05:41:25Z`、`vcs.modified=false` |
| 切换 | 2026-11-25 05:56:07Z 同目录 `mv` 原子替换，仅 `systemctl restart una-seller-api`；PID 734130 → 784691，`NRestarts=0`，启动日志仅正常配置行 |
| 三处哈希一致 | 构建产物、`/opt/una-seller/una-seller-api`、运行进程 `/proc/784691/exe` 均为 `dcbffb41…93bdd4a` |
| 保持不变 | `api.env` `012a941c…`、`systemd unit` `68a4c433…`、Nginx vhost `c1f3983e…` 前后一致；未改 Nginx/systemd/密钥，未重启其他服务（vaultwarden 未动） |

## 4. 接口验收

| 检查 | 回环 | 公网 | 结果 |
| --- | --- | --- | --- |
| `/healthz` | 200 | 200 | PASS |
| 匿名 `/api/v1/collections` | — | 401 | PASS |
| 匿名 `/api/v1/stores` | — | 401 | PASS |
| 匿名 `/api/v1/products` | — | 401 | PASS |
| 匿名 `/api/v1/account/service-configs` | — | 401 | PASS |
| C0 `GET .../publication-mode/probe` | — | 405（`Allow: POST`） | PASS |
| C0 匿名 `POST .../publication-mode/probe` | — | 401 | PASS |
| 已授权订单列表 `GET .../orders?site=MLB&logistic_type=remote` | — | 200，20 条，`has_more=true` | PASS |
| 已授权订单详情 `GET .../orders/{order_id}` | — | 200 | PASS |
| OTA 默认通道 `X-Una-Include-Prerelease: 0` | — | 204 + `Cache-Control: no-store` | PASS |
| OTA 测试通道 `X-Una-Include-Prerelease: 1` | — | 200 + `Cache-Control: no-store`，`version=0.5.0` | PASS |

订单读取使用客户端现有会话，Bearer 仅在进程内存中传递，未落盘、未输出、未记录。

## 5. 单笔官方账单手动读取与限流诊断

- 通过 API `POST /api/v1/stores/{store}/orders/billing-preview?site=MLB&logistic_type=remote`，仅提交一笔已有授权订单 `2000018543375722`（MLB / remote），**一次调用、无重试、无批量查询**。
- 结果 HTTP 200，`items=1`、`available=true`、`net_income_usd=null`（未推断净收益）；未输出原始官方 JSON、买家信息或 Token。
- 本次未出现 429，因此 200 响应无 `Retry-After` / `Access-Control-Expose-Headers`；不人为触发上游限流，不绕过冷却。
- Retry-After 跨源暴露机制：目标代码在 `order_rate_limited` 分支设置 `Access-Control-Expose-Headers: Retry-After` 与 `Retry-After: <秒>`；CI `Go Backend` 与服务器定点测试 `TestBillingPreviewErrorMapping` PASS，确认 429 时桌面端可读取该响应头。
- 限流诊断字段：固定标签 `source`（`local_cooldown` / `upstream`）、`resource`（`billing` 等）、`upstream_http_status`、`retry_after_seconds`；定点测试 `TestOrderReadResourceUsesFixedLabels`、`TestBillingWrappedRateLimitUsesExistingOrderCooldown` PASS。部署后日志无 rate-limited 记录，故本轮无脱敏诊断字段值可记录；代码中标签不含 URL、订单号、店铺号或 Token。

## 6. 边界与未执行

- 无新增数据库迁移，未执行 `migrate`。
- 未重复完整数据库恢复演练，仅核对既有恢复点名称/大小/哈希。
- 未批量查询官方账单；未修改 Nginx、systemd、密钥或其他服务；未发布客户端、未创建新 Tag。
- 客户端实际 OTA、安装、数据保留与 V0.5/#94 功能验收：等待负责人人工结果。

## 7. 负责人客户端人工验证点

1. 测试版开关关闭 → 检查更新应提示已是最新版（服务端默认通道 204）。
2. 打开「接收测试版更新」→ 重新检查，应出现 0.5.0（测试通道 200）。
3. 下载、安装、覆盖安装完成，界面版本显示 `0.5 测试版`，重启后保持。
4. 登录状态、店铺授权、采集工作稿、已刊登列表、订单/财务设置不丢失。
5. #94 跟进项：变体编辑、物流追踪与账单展示的实测结果符合验收口径。
6. 手动账单预览：单笔订单读取可用；如出现限流提示，界面应显示冷却时间且不连续重试、不绕过限流。
7. 异常路径：断网/超时有明确提示；正式版客户端（测试通道关闭）不应收到 0.5.0。
