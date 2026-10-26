# M0-02A Go API 与 HTTPS 部署

本说明只部署现有 `backend` 的 `/healthz`。当前没有 OAuth 回调、业务 API 或认证；Nginx 只能转发现有 Go handler 暴露的路径，不代表店铺授权已经完成。

## 目标配置

- 域名：`mkd.186512.xyz`
- Nginx：独立 vhost；80 仅服务 ACME HTTP-01 和 HTTPS 跳转，443 代理至 `127.0.0.1:8787`
- Go：`una-seller-api.service`，系统用户 `una-seller`，二进制 `/opt/una-seller/una-seller-api`
- Certbot webroot：`/var/www/letsencrypt`；证书续期后 reload Nginx
- Supabase 与 Mercado Libre 凭据尚未接入。systemd 预留 `/etc/una-seller/una-seller.env`，不在仓库保存环境文件或密钥。

## 服务器基线

以下是用户提供的既有环境信息，**本轮未通过 SSH 复核**；实施前按下方只读命令重新确认：

| 项目 | 既有报告 |
| --- | --- |
| OS / CPU / 内存 | Ubuntu 24.04.4 LTS；2 核；约 3.6 GiB 可用物理内存（4 GB 配置） |
| 磁盘 | 约 37 GB 可用 |
| Nginx | 已安装，80 端口监听；此前 `nginx -t` 通过；443 此前未监听 |
| 既有容器 | `we-mp-rss`，映射 `127.0.0.1:8001` |

## 先决条件与只读检查

先确认域名 A/AAAA 解析指向当前源站，Cloudflare 代理及证书验证路径可用；Cloudflare SSL/TLS 对该主机设为 **Full (strict)**。确认服务器公网地址、SSH 主机指纹和登录方式后，再执行服务器检查。不要把代理 IP 当源站地址。

```sh
uname -a
uname -m
free -h
df -h /
sudo ss -ltnp
sudo nginx -T
sudo nginx -t
sudo ufw status verbose
docker ps --format 'table {{.Names}}\t{{.Ports}}\t{{.Status}}'
command -v certbot || true
sudo certbot certificates
systemctl list-timers --all | grep -i certbot || true
```

另在云平台控制台只读确认 80/443 入站规则；不要为此启用 UFW，也不要改全 Zone SSL 策略。确认现有站点里没有 `mkd.186512.xyz` 冲突后，备份 `/etc/nginx`：

```sh
sudo cp -a /etc/nginx "/root/nginx-backup-$(date +%Y%m%d%H%M%S)"
```

本次执行记录见 [M0-02A 状态](M0-02A-STATUS.md)。

## 构建与 systemd

在构建机根据目标机 `uname -m` 选择 `GOARCH`（`x86_64` 对应 `amd64`，`aarch64` 对应 `arm64`），然后从 `backend` 目录构建：

```sh
mkdir -p dist
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o dist/una-seller-api ./cmd/api
```

将二进制和三个部署配置复制到服务器临时目录；`SSH_HOST` 必须是确认过的 SSH 主机名/IP，不可填 Cloudflare 代理 IP：

```sh
scp dist/una-seller-api deploy/systemd/una-seller-api.service deploy/nginx/mkd.186512.xyz-http.conf deploy/nginx/mkd.186512.xyz.conf deploy/certbot/reload-nginx "$SSH_HOST:/tmp/"
```

然后在服务器安装：

```sh
sudo useradd --system --home-dir /nonexistent --shell /usr/sbin/nologin una-seller
sudo install -d -o root -g una-seller -m 0750 /opt/una-seller
sudo install -o root -g una-seller -m 0750 /tmp/una-seller-api /opt/una-seller/una-seller-api
sudo install -d -o root -g root -m 0750 /etc/una-seller
sudo install -o root -g root -m 0644 /tmp/una-seller-api.service /etc/systemd/system/una-seller-api.service
sudo systemctl daemon-reload
sudo systemctl enable --now una-seller-api
systemctl status --no-pager una-seller-api
curl --fail --silent --show-error http://127.0.0.1:8787/healthz
```

若系统用户已存在，跳过 `useradd`。本服务目前无需环境变量；后续数据库/OAuth 配置应仅写入 root 可读的 `/etc/una-seller/una-seller.env`，并由对应阶段定义变量名和权限，不要提交该文件。Go 仍只监听 loopback；不要开放 8787 入站端口。

## Nginx 与 Let's Encrypt

只在确认解析正确、80/443 可达且现有 Nginx 配置没有域名冲突后执行。`we-mp-rss` 与其他 vhost 保持原样。下面路径示例用于该域名独立站点，不覆盖 `default`。

```sh
sudo install -d -o root -g root -m 0755 /var/www/letsencrypt
if ! command -v certbot >/dev/null 2>&1; then sudo apt install certbot; fi
sudo install -o root -g root -m 0644 /tmp/mkd.186512.xyz-http.conf /etc/nginx/sites-available/mkd.186512.xyz
sudo ln -s /etc/nginx/sites-available/mkd.186512.xyz /etc/nginx/sites-enabled/mkd.186512.xyz
sudo nginx -t && sudo systemctl reload nginx
sudo certbot certonly --webroot -w /var/www/letsencrypt -d mkd.186512.xyz
```

只有 Certbot 成功签发后，才将完整配置安装到同一 `sites-available` 路径。它保留 ACME 路径、仅把本主机 HTTP 请求重定向到 HTTPS，并将请求转发给 Go。

```sh
sudo install -o root -g root -m 0644 /tmp/mkd.186512.xyz.conf /etc/nginx/sites-available/mkd.186512.xyz
sudo nginx -t && sudo systemctl reload nginx
sudo install -o root -g root -m 0755 /tmp/reload-nginx /etc/letsencrypt/renewal-hooks/deploy/10-reload-nginx
sudo systemctl enable --now certbot.timer
sudo certbot renew --dry-run
```

每次 Nginx 配置变化先通过 `nginx -t` 再 reload。若 Ubuntu 镜像中的 Certbot 未提供 `certbot.timer`，停止并按该系统安装来源配置自动续期，不创建未经验证的定时任务。HTTP-01 失败时先检查 DNS、Cloudflare/WAF 与 80 可达性；可评估 Cloudflare DNS-01，但 API Token 只能留在服务器受限凭据文件里。本流程不使用 standalone，也不停止 Nginx。

`X-Real-IP` 取自 Cloudflare 的 `CF-Connecting-IP`。它仅在请求确实经过 Cloudflare 时可信；不要基于该头做授权。若将来需要安全地识别/限制来源，先通过源站防火墙或 Cloudflare 官方 IP 段限制直接访问，再单独验证 Nginx realip 配置。

## 验收

```sh
go test ./...
go vet ./...
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/api
systemctl is-active una-seller-api
curl --fail --silent --show-error http://127.0.0.1:8787/healthz
sudo ss -ltnp | grep -E ':(80|443|8787)\b'
curl --fail --silent --show-error https://mkd.186512.xyz/healthz
openssl s_client -connect mkd.186512.xyz:443 -servername mkd.186512.xyz -verify_return_error </dev/null
sudo certbot renew --dry-run
docker inspect --format '{{.State.Status}} {{.HostConfig.NetworkMode}}' we-mp-rss
```

成功标准：Go 返回 `{"status":"ok","service":"una-seller-api"}`；8787 仅监听 `127.0.0.1`；公网上的 HTTPS 证书名称、有效期和验证链通过；Certbot dry-run 通过；原 `we-mp-rss` 容器仍运行且原 Nginx vhost 正常。缺证据的项目标记 `NOT_RUN`，不可由本地构建或 PR 状态代替。

## 回滚

部署前保存原 Nginx 配置和当前二进制。回滚时恢复该域名的旧配置/二进制，先运行 `sudo nginx -t`，通过后 reload，再检查原站点与 `we-mp-rss`。如果此前没有 Una Seller 服务，仅禁用 `una-seller-api.service` 并移除该域名 vhost；不要删除证书、其他 Nginx 配置或 Docker 数据。

## 后续边界

- M0-02B：定义 Supabase PostgreSQL 连接变量与 TLS 验证，最小化连接池；密钥只放 `/etc/una-seller/una-seller.env`。本阶段不建库、不建业务表、不连接 Supabase。
- M0-03：实现并测试 OAuth state、回调与 Token 生命周期后，才把 `/oauth/mercadolibre/callback` 接入 API；Client Secret/Token 不进入仓库或日志。已登记 Redirect URI 不等于已完成 OAuth。
