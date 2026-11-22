# #57 · Una Seller 远程 Go API 受控升级 / 产品箱验收

> 状态：**PRECHECK_PARTIAL / BACKUP_NOT_VERIFIED / SERVER_NOT_VERIFIED / NO_PRODUCTION_WRITE**。
> 此文档和本分支只是升级准备，**不授权执行生产操作**。不要把 #59 代码合并或 Windows 构建成功误写成远程产品箱已验收。

## 当前已独立核对（2026-11-22）

- 主干固定待部署代码：`5bdac9660dbde39ae60d5de98c3ff1081fb94efa`（#59 Squash Merge），对应 [main CI #44979845433](https://github.com/Fan6512/una-seller/actions/runs/44979845433) **4/4 PASS**，Go/PG17 自动化集成测试已过。
- 生产 Supabase **una-seller**：PostgreSQL **17.11**、`ACTIVE_HEALTHY`；`app.source_products`、`app.collection_working_drafts` 已存在，**`app.own_products` 不存在、工作稿 `listing_fields` 不存在**。私有 `app` schema 对 `anon`、`authenticated` 不开放 USAGE。
- 已有业务资料（仅读取聚合计数，不查询个人数据）：采集商品 3、工作稿 0、订单 72、CBT 店铺连接 1。**不可清库、覆盖原订单或重建账号。**
- 该 Supabase 组织当前为 **Free**；按 [官方 Backups 文档](https://supabase.com/docs/guides/platform/backups)，不能依赖 Pro/Team 的自动日备份。**未证实任何可恢复备份**。
- 仓库的 `0008_own_products.sql` 新增独立产品箱（RLS、owner FK、source FK `ON DELETE SET NULL`），`0009_listing_editor_fields.sql` 对工作稿与产品箱加 `jsonb listing_fields`。迁移代码是**显式**执行，不会在 API 启动时自动迁移。
- 无可用 SSH 执行入口来读取生产 Ubuntu 当前二进制、现役 Nginx vhost 或备份文件，故不能声称远程 Go 已升级或已验收；当前远程 HTTP 探测也没有获得可验证响应。

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
