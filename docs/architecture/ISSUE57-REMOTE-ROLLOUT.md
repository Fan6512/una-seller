# #57 · Una Seller 远程 Go API 受控升级 / 产品箱验收

> 当前状态（2026-11-22）：**SERVER_BASELINE_PASS / BACKUP_VERIFIED / MIGRATION_REHEARSAL_PASS / SCHEMA_MIGRATED_PASS / GO_DEPLOYED_PASS / NGINX_ROUTED_PASS / HTTP_ANON_PASS / MANUAL_PENDING(AUTH, GUI)**。已核对 [生产部署脱敏回执](https://github.com/Fan6512/una-seller/issues/57#issuecomment-7077260112)，完整保留 [服务器与恢复回执](https://github.com/Fan6512/una-seller/issues/57#issuecomment-7074227002) 及 [隔离迁移演练回执](https://github.com/Fan6512/una-seller/issues/57#issuecomment-7075825024)。
> 本文前半部分是已实际执行过的升级操作书（保留审计）；**生产切换已完成，不应再次执行 migrate、重复替换现役 Go/Nginx 或重复恢复备份**。下一步只做授权接口、桌面客户端和真实 CBT 类目推荐验收。验证通过前不可关闭 #57 或发布 V1.0 正式版。

## 生产切换回执

| 验证项 | 状态 | 证据 |
| --- | --- | --- |
| 加密备份与隔离恢复 | PASS | PostgreSQL 17.11 独立恢复，15 张旧业务表结构和逐表行摘要一致。范围为 app schema；所有者、ACL、Supabase Auth 及异地恢复待验证。 |
| 0008/0009 迁移演练 | PASS | 从独立恢复库完整执行 0001–0009 两次。第一次核对旧数据与新增结构，第二次核对幂等。 |
| 生产结构及旧数据 | PASS | 2026-11-22 执行一次 0001–0009。16 张业务表 RLS 生效，own_products 和两处 listing_fields 存在，旧表数据摘要一致。 |
| Go / Nginx | PASS | 固定提交 5bdac966，部署二进制前后一致；Go 仅监听回环。Nginx 仅增补产品箱与类目预测转发。 |
| 匿名 HTTP | PASS | 回环及公网 healthz 200；采集、店铺、订单、产品箱与类目预测无 Token 返回 401。 |
| 历史 Windows 部分验收 | PASS | #65 旧 QA HEAD c131da1b 记录了工作稿云端保存和重启读回、三语标题、图片及 9 变体保存。 |
| 云端转库和独立产品重启读回 | MANUAL_PENDING(AUTH, GUI) | 从批准的非实盘采集来源保存工作稿，勾选转库后在产品箱编辑并重启复查。当前 #65 候选 GUI 同步复验。 |
| CBT 官方类目推荐 | MANUAL_PENDING(AUTH) | 用已授权测试店铺提交英文测试标题，记录真实候选或官方拒绝状态。 |
| 生产回退 | NOT_RUN | 备有 Go 与 Nginx 私有回退副本。切换门禁通过，实际未回退。 |

[备份与恢复](https://github.com/Fan6512/una-seller/issues/57#issuecomment-7074227002) · [两次迁移演练](https://github.com/Fan6512/una-seller/issues/57#issuecomment-7075825024) · [生产切换](https://github.com/Fan6512/una-seller/issues/57#issuecomment-7077260112) · [Windows 历史反馈](https://github.com/Fan6512/una-seller/pull/65)


## 当前已独立核对（2026-11-22）

- 主干固定待部署代码：`5bdac9660dbde39ae60d5de98c3ff1081fb94efa`（#59 Squash Merge），对应 [main CI #44979845433](https://github.com/Fan6512/una-seller/actions/runs/44979845433) **4/4 PASS**，Go/PG17 自动化集成测试已过。
- 生产 Supabase **una-seller**：PostgreSQL **17.11**、`ACTIVE_HEALTHY`；`app.source_products`、`app.collection_working_drafts` 已存在，**`app.own_products` 不存在、工作稿 `listing_fields` 不存在**。私有 `app` schema 对 `anon`、`authenticated` 不开放 USAGE。
- 已有业务资料（仅读取聚合计数，不查询个人数据）：采集商品 3、工作稿 0、订单 72、CBT 店铺连接 1。**不可清库、覆盖原订单或重建账号。**
- 该 Supabase 组织当前为 **Free**；按 [官方 Backups 文档](https://supabase.com/docs/guides/platform/backups)，不能依赖 Pro/Team 的自动日备份。**2026-11-22 已收到备份与独立恢复验证 PASS 的脱敏回执**，其证据范围是 `app` 数据与结构（15 张表、142 列、66 约束、26 索引、逐表行摘要）；所有者/ACL、Supabase Auth、异地恢复未验证。AES-256 私有归档不得进入 Git。
- 仓库的 `0008_own_products.sql` 新增独立产品箱（RLS、owner FK、source FK `ON DELETE SET NULL`），`0009_listing_editor_fields.sql` 对工作稿与产品箱加 `jsonb listing_fields`。迁移代码是**显式**执行，不会在 API 启动时自动迁移。
- 已收到 Ubuntu 执行方脱敏结果：Ubuntu 26.04 LTS x86_64，Go 与 Nginx 正常运行，Go 只监听 `127.0.0.1:8787`；公网和回环健康检查均 200，匿名采集 401、产品箱和类目推荐仍 404，`nginx -t` 通过；旧 Go / vhost / systemd 已留私有校验副本。**不等于升级已执行**。本会话不具备 Ubuntu SSH 执行通道，不可自行声称现网二进制已经切换。

## 门禁 0：Ubuntu 只读核对（不含任何写入）

在目标 Ubuntu 的 SSH 会话中，从本仓库 pinned commit 使用：

```bash
bash scripts/ops/issue57-remote-preflight.sh
sudo nginx -t
sudo systemctl status una-seller-api --no-pager
```

预检脚本只读检查 service PID/状态、8787 监听、Nginx 语法、Go binary SHA256、匿名 HTTPS 路由状态。**不得公开完整 `api.env`、`nginx -T`、生产日志、Bearer Token**。确认只有 `127.0.0.1:8787` 在监听（而不是 `0.0.0.0:8787`）；核对现役 vhost 实际文件路径、系统架构和旧版 Go 路由。若服务已经不健康、现役路径与仓库模板不一致或 SSL 异常，标记 **BLOCKED_SERVER_BASELINE**，先解决基线。

## 门禁 1：有验证的备份（不许跳过）

由拥有数据库权限的负责人在**受信任的隔离终端**使用 PostgreSQL 17 客户端执行；使用安全方式设置 `PGHOST`、`PGPORT`、`PGUSER`、`PGDATABASE`、`PGPASSWORD`、`PGSSLMODE` 等 libpq 环境，**不要把密码、DSN 写到 Git、命令历史、Issue 或 CI 输出**。核对连接对应实际 una-seller 项目；不可借用其他 Supabase 项目。下面是最小应用域备份（只备份 `app` schema），包含店铺加密 Token 和订单，文件是**高度敏感**的：

```bash
umask 077
BACKUP_DIR="$HOME/una-seller-private-backups"
mkdir -p "$BACKUP_DIR"
BACKUP_FILE="$BACKUP_DIR/app-before-issue57-$(date -u +%Y%m%dT%H%M%SZ).dump"
pg_dump --format=custom --schema=app --no-owner --no-privileges --file="$BACKUP_FILE"
test -s "$BACKUP_FILE"
pg_restore --list "$BACKUP_FILE" > "$BACKUP_FILE.list"
sha256sum "$BACKUP_FILE"
```

**强制再验恢复**：在独立、一次性的 PostgreSQL **17** 数据库，使用测试数据库的 `PG*` 环境运行 `pg_restore --no-owner --no-acl --single-transaction --dbname=<一次性测试库名称> "$BACKUP_FILE"`，对照 `app` 表结构及采集/订单/店铺的聚合计数。注意：必须在**断开生产凭据、验证测试连接目标**后才执行恢复！`pg_restore --list` 单独成功不等于数据可恢复。备份保存在私有加密存储，不能上传公开仓库和普通 Actions Artifact。Supabase Auth、其他 schema 与平台配置如需灾备，请另做全项目备份方案；本 `app` 定向备份只覆盖这轮业务迁移涉及的数据范围。

缺少可恢复备份时，**BLOCKED_BACKUP**，不得进行 `0008/0009` DDL。

## 门禁 2：准备 pin 到主干 HEAD 的新 Go 二进制（未部署）

确认代码 SHA 与 CI，使用私有临时目录编译，服务器无需全局安装 Go：

```bash
git clone --no-checkout https://github.com/Fan6512/una-seller.git /tmp/una-seller-issue57-src
git -C /tmp/una-seller-issue57-src checkout 5bdac9660dbde39ae60d5de98c3ff1081fb94efa
cd /tmp/una-seller-issue57-src
sudo docker run --rm -v "$PWD:/src" -w /src/backend golang:1.26 \
  sh -c 'go build -trimpath -o /src/una-seller-api.issue57 ./cmd/api'
sha256sum una-seller-api.issue57
```

检查 `uname -m` 与待运行平台相符、构建产物正常。**不要覆盖 `/opt/una-seller/una-seller-api`**。若已有同名临时目录，停止并核对来源，不能贸然删已有数据。不得修改其他业务容器。

## 门禁 2.5：一次性迁移演练（生产写入前）

从已经验证可恢复的同一份加密 `app` 归档，重新恢复到**隔离 PG17 空库**，用 pin 的新 Go 候选运行一次 `migrate`，核对 `app.own_products`、两处 `listing_fields`、RLS 与旧 15 张表的数据摘要，随后在这套临时库再运行一次 `migrate` 检查幂等；演练结束销毁临时测试库。

**重要：`backend/migrations.Apply` 每次都会按顺序运行 `0001`～`0009`，而非只运行 `0008/0009`**。因此这次演练特别用于核实旧 `0001-0007` 在当前有订单与店铺数据的数据库上重放时不会产生副作用；不能因新增的两份 SQL 是追加式就跳过此项。仅记录 `MIGRATION_REHEARSAL_PASS/FAIL` 和脱敏行数/摘要，不贴数据库连接串、token、原始归档或敏感日志。演练失败即终止升级，不进生产。

## 门禁 3：升级窗口（仅备份与服务器证据完成、负责人明确批准后）

1. **保存旧版本**：备份原 `/opt/una-seller/una-seller-api` 二进制、实际 systemd unit、仅 `mkd.186512.xyz` 的 Nginx vhost。保留完整权限和 sha256；**不要复制或公开含密钥的 `/etc/una-seller/api.env`**。
2. 确认短暂维护窗口，**只停止** `una-seller-api`（不动别的服务）。使用已审核 pinned HEAD 的 Go `./cmd/api migrate`，它会按 `0001`～`0009` 执行幂等 SQL。须让迁移进程使用**与现役服务相同的受保护 EnvironmentFile**，而非替换 DB DSN、清库或手动“重建”前几个 migration。可以用 systemd transient oneshot unit 加载 `/etc/una-seller/api.env`，但要先核对生产 unit、服务用户、路径与版本；不要简单 `source` 未核实的 EnvironmentFile。
3. 迁移后只读确认：`app.own_products` 新表、两处 `listing_fields` 字段、RLS、原 3/72/1 条基线数据仍在。若异常立即停止，保留原二进制，**不可尝试删除新表“回滚”**。
4. 仅在完成检查后，原位切换**一个 Go API 二进制**并 `systemctl start/restart una-seller-api`。密钥/店铺 Token 不修改。
5. Nginx 只对实际 `mkd` vhost 增加/核对 `location ^~ /api/v1/products`、`location = /api/v1/categories/predict`，按仓库 [nginx 模板](../../deploy/una-seller/nginx-mkd.conf) 补充授权头和 `no-store`；**禁止直接覆盖现网整个 vhost**，不能影响 OAuth callback 或其他域名。先 `sudo nginx -t`，成功才 `sudo systemctl reload nginx`。
6. 匿名只读冒烟：`/healthz = 200`；`/api/v1/collections`、`/api/v1/products`、`/api/v1/categories/predict?q=...` 均应 **401**（不携带 Token），不能仅凭 healthz 通过就宣称 DB/Auth 正常。
7. 用自己的受控测试账号、经批准的测试来源，完成 `GET /api/v1/products = 200`，工作稿 `PUT`→批准转库→产品箱 `PUT/GET`→重启 API→再次读回；确认保存、来源只读、幂等、跨账号隔离。**不要对真实订单、客户发货或 CBT 实盘刊登执行写入**。官方 CBT 类目推荐须测试店铺已真实授权后单独记 PASS/FAIL/NOT_RUN。

## 回退与判定

- **Go/Nginx 回退**：如果健康、登录、店铺列表、原有订单/采集箱任何一项异常，恢复上一步备份的**旧 Go 二进制与 mkd vhost**，只重启 `una-seller-api`、只 reload Nginx；重新做旧接口只读回归。不关闭其他服务。
- **DB 不能靠恢复旧 Go 自动回滚**。本轮 `0008/0009` 只做加法；Go 回退时原则上保留新表和列，避免毁掉升级后写入的数据。确需 DB 恢复要先评估备份时间点后的所有新增订单/店铺记录，明确数据恢复损失并重新审批；不自动 drop/truncate。
- 输出状态分别记录：`BACKUP_VERIFIED`、`SCHEMA_MIGRATED`、`GO_DEPLOYED`、`NGINX_ROUTED`、`HTTP_ANON_PASS`、`AUTH_PRODUCT_BOX_PASS`、`WINDOWS_GUI_PASS`、`CBT_PREDICT_PASS`。只有实际对应步骤完成才能 PASS。
- 本阶段**不创建正式桌面 Tag/Release，不升级其他容器，不直接运行真实订单同步/采购/发货**。#57 只有远程 + Windows 双端实际读写回归确认后才可关闭。
