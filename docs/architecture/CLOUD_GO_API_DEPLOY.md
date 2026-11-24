# M0-02B：Una Seller 云端 Go API 最小部署操作说明

> 2026-11-17 · 对应 [Issue #17](https://github.com/Fan6512/una-seller/issues/17) · **准备/待服务器验收（SERVER_E2E_BLOCKED）**
>
> 仓库配置可以通过 CI 验证；**以下命令没有在现有 Ubuntu 主机上执行过**。只有 SSH 真机、HTTPS 与 Windows 真实登录后才允许标记 `CLOUD_DEPLOYED / PASS`。

## 目标与范围
- 现有 Ubuntu 服务器（不改变同机其他服务）+ 现有 Nginx + Cloudflare `mkd.186512.xyz` + 一个 Go API + 已有 Supabase Auth/PostgreSQL。
- 本地 `tauri:dev` 仍默认 `http://127.0.0.1:8787`，正式 Windows 打包默认 `https://mkd.186512.xyz`；`VITE_API_BASE_URL` 是编译时覆盖选项。**部署完成前不要把新包作为可用的云端客户端**。
- Go 服务在云端仍只监听 `127.0.0.1:8787`，Nginx 暴露 HTTPS；仅允许 `/healthz` 和 `/api/v1/collections...` 路由。使用已验证的 Supabase Auth Access Token；不得直接开放 8787。
- 不建新数据库、不引入 Redis、Caddy、K8s 或自动 SSH 部署；**绝不修改或重启其他业务容器**。

## 1. 先只读检查服务器，不要覆盖配置
在 Ubuntu 的 SSH 终端执行，检查输出后再决定是否继续：
```sh
uname -a
docker --version
sudo nginx -t
sudo ss -lntp | grep -E ':(80|443|8787)\b' || true
sudo nginx -T 2>/dev/null | grep -E 'server_name .*mkd\.186512\.xyz' || true
getent ahosts mkd.186512.xyz || true
curl -sS -o /dev/null -w 'Supabase Auth HTTP=%{http_code}\n' --max-time 10 https://rwktwfxzfkzglxwaulcm.supabase.co/auth/v1/user
```
对 Auth 的无凭据请求只检查是否获得 HTTP 响应；`000`/连接重置为网络阻断。已有 `mkd` 站点或 8787 端口占用时先停止操作并核对归属。不要公开完整 `nginx -T`、服务环境文件或密钥。

另在 Cloudflare 检查 `mkd` 的 DNS 指向目标主机、安全组允许 80/443、代理及 TLS 模式。不要全局修改其它域名的 SSL/TLS 策略。

**准备公网启用前的账号门禁**：本项目当前仅有一个负责人账号、尚未对外注册。进入 Supabase Dashboard → Authentication 的登录提供商/注册设置，确认**关闭新的公开账号注册**，再开放云端 Go API。现有账号的邮箱密码登录应继续可用并在客户端复测。这项设置在本地开发阶段已被允许暂缓，但**公网部署前不能继续当作已完成**；不要通过删除现有用户来达到限制注册的目的。

## 2. 从仓库构建 Linux amd64 Go 程序（不要求主机安装 Go）
服务器已有 Docker，可**仅用 Docker 临时编译**，不影响现有业务容器：
```sh
git clone --depth 1 --branch feat/m0-02b-cloud-go-api-20261117 https://github.com/Fan6512/una-seller.git /tmp/una-seller-cloud
cd /tmp/una-seller-cloud
sudo docker run --rm -v "$PWD:/src" -w /src/backend golang:1.26 sh -c 'go build -trimpath -o /src/una-seller-api ./cmd/api'
sudo install -d -m 0755 /opt/una-seller
sudo install -m 0755 una-seller-api /opt/una-seller/una-seller-api
```
适用于 `uname -m` 是 `x86_64` 的现有 Ubuntu；若架构不同，必须改成对应编译目标，不要强行运行。PR 合并后，后续更新应改为按已审核的 `main` 提交 SHA 检出，不长期依赖 Draft 分支。

## 3. 使用 systemd 管理 Go，私有配置只存服务器
```sh
id una-seller >/dev/null 2>&1 || sudo useradd --system --create-home --home-dir /var/lib/una-seller --shell /usr/sbin/nologin una-seller
sudo install -d -m 0700 /etc/una-seller
sudo install -m 0600 deploy/una-seller/api.env.example /etc/una-seller/api.env
sudoedit /etc/una-seller/api.env
sudo install -m 0644 deploy/una-seller/una-seller-api.service /etc/systemd/system/una-seller-api.service
sudo systemctl daemon-reload
sudo systemctl enable --now una-seller-api
sudo systemctl is-active una-seller-api
```
在 `sudoedit` 中填入服务器专用的 `DATABASE_URL`、`SUPABASE_URL`、`SUPABASE_PUBLISHABLE_KEY`。不要复制到 GitHub、日志或桌面程序。

**P0 / Issue #75：用户自行绑定云端 AI、翻译 Key（PR #77，独立授权后才部署）**

先审核/合并代码并确认生产备份，再单独批准执行 `backend/cmd/api/main.go` 对应的 `api migrate`（新表为私有的 `app.user_service_credentials`，版本文件 `0011`；未合并的 PR #66 使用 `0010`，两者的迁移在最终 main 合流后统一核对）。**不得在未授权时运行 migrate、重启生产 API 或试真实付费调用。**

`deploy/una-seller/api.env.example` 同步新增：
- `UNA_SELLER_SERVICE_CREDENTIALS_KEY`：在服务器内独立生成的随机 32 字节 Base64 主密钥（可用 `openssl rand -base64 32`），绝不复用 Mercado Libre Token 加密密钥；保管备份，权限仅限服务进程；留空会禁用所有账号 Key 管理功能，启动时明确 WARNING，客户端查询返回 503。
- `UNA_SELLER_SERVICE_CREDENTIALS_KEY_ID=v1`：每次正式换主密钥必须同时换 ID（例如 `v2`）。旧记录不会被自动删除或不安全降级；旧账号需重新绑定 Key，或者由负责人另行审批全量重加密流程。客户端看到“需重新绑定”的提示。严禁无计划更换或丢失主密钥。
- `AI_ALLOWED_BASE_URLS`：服务器允许的 OpenAI-compatible HTTPS API 地址，多个用英文逗号分隔，如 `https://api.deepseek.com/v1`。官方 `https://api.openai.com/v1` 内置；尾部单个 `/` 可规范化，非法条目被跳过并记录通用 warning，不允许任意域名代理。部署时确认实际模型可用及受支持的 JSON 模式。
- `TRANSLATION_PROVIDER`、`AI_API_KEY` 等原部署级变量仍供兼容及维护用；**启用用户 Key 模式后，未绑定的用户不能自动消费共享 Key 配额**。离线 CBT 类目全量翻译仍由独立运维配置运行，不使用用户的付费 Key。

按此前审批的 HTTPS/Go 服务流程重新加载配置后，先进行带 Bearer 的账号配置 GET（仅返回状态、不返回密钥），再由负责人手动测试一个已保存商品的云 AI；不得直接启动全词典批量翻译。操作日志仅记录 owner/kind/provider/绑定或删除事件，不能收集 Key 或用户输入。完成后仍需 Windows 集中 GUI 验收。

Supabase 数据库已建立 `app` schema，无需重新执行 migrate。连接串从 Supabase Dashboard → Connect 获取：服务器支持 IPv6 可选 Direct `:5432`；仅 IPv4 使用 **Session Pooler `:5432`**，均启用 TLS 校验（例如 `sslmode=verify-full`）。不要把 HTTP 代理用于 PostgreSQL 连接；若服务器访问 Auth 失败再单独核查服务器出站网络。

**本机验证：**
```sh
curl -i http://127.0.0.1:8787/healthz
curl -i http://127.0.0.1:8787/api/v1/collections
```
预期分别 **200** 和 **401**（第二条故意不带 Token）。若 `503` 或启动失败，先看 `sudo systemctl status una-seller-api` 与 `sudo journalctl -u una-seller-api -n 50 --no-pager`（分享前删去私密信息）。`UNA_SELLER_CLOUD=1` 会拒绝缺少 DB/Auth 环境配置的启动。

## 4. 为独立域名取得证书，随后启用 Nginx 反代
**先确认** `mkd.186512.xyz` 的 DNS 与 80/443 安全组、以及服务器没有同名 `server_name`。在初次签证书前，可以先为该域名创建一个只在 80 返回 404 的独立 Nginx vhost，运行 `sudo nginx -t` 后再重载；不能覆盖已有业务配置。使用 `sudo certbot --nginx -d mkd.186512.xyz` 签发 Let's Encrypt 证书；如果 Cloudflare 阻断 HTTP-01，则改用 DNS-01，不停止 Nginx。

**证书实际签发后**，对 Certbot 生成的同名站点文件做备份，再用 `deploy/una-seller/nginx-mkd.conf` 模板替换/合并**仅该域名**的 vhost 内容：
```sh
sudo cp /etc/nginx/sites-available/mkd.186512.xyz.conf /etc/nginx/sites-available/mkd.186512.xyz.conf.pre-una-seller
sudo install -m 0644 deploy/una-seller/nginx-mkd.conf /etc/nginx/sites-available/mkd.186512.xyz.conf
sudo nginx -t && sudo systemctl reload nginx
sudo certbot renew --dry-run
```
以上路径是假设 vhost 使用 `sites-available/mkd.186512.xyz.conf`；**必须先核对实际文件路径及 Certbot 编辑结果**再执行，不能盲目复制。确认证书文件路径与模板一致，Cloudflare 到源站的 TLS 使用 Full (strict)。模板的未知路径返回 404，敏感列表加 `Cache-Control: no-store`。

## 5. 真正的外部验收（所有检查都需对准同一部署版本）
```sh
curl -i https://mkd.186512.xyz/healthz
curl -i https://mkd.186512.xyz/api/v1/collections
```
预期 **200 / 401**。然后从新的 Windows 构建登录账号，检查右上角“服务正常”、采集箱空列表响应 **200**；不需要为了本轮验收创建测试商品或第二个账号。该登录后请求通过云端 API 获取商品，不能只凭 `/healthz` 200 宣称 DB/Auth 已通。

- 无 Token 不能读写；已验证登录后能看到属于自己的列表；Go 8787 不监听公网地址。
- 不影响既有 RSS 站点，Cloudflare 不缓存含认证的响应。
- macOS 与浏览器插件独立后续验收，不能冒充已通过。

**限制：桌面应用依然直接访问 Supabase Auth。** 即使 Go API 已云端化，当前 Windows 网络若无法访问 Supabase Auth，登录/续签仍可能需要本机代理；本期不另做登录代理。

## 5.1 Auth 校验快路径（PR #81，待审核部署）

> 以下为 **PR #81 代码行为**，并非已经在 Ubuntu 服务器完成部署或真实性能验收；以实际部署 SHA 为准。

- **仅 4 类只读端点**启用快路径：`GET /api/v1/collections`、`GET /api/v1/collections/{uuid}`、`GET /api/v1/products`、`GET /api/v1/products/{uuid}`。UUID 必须通过现有 `ValidUserID` 校验。其他路径即使使用 GET（包括 `/working-draft`、`/cost`、店铺绑定/授权类接口）仍走 `/auth/v1/user` 实时验证；所有 POST、PUT、DELETE 等写操作也不走快路径。新路由不会自动进入白名单。
- 对 **HS256** Access Token：后端仍先以 Supabase `/auth/v1/user` 的成功响应确定身份，仅对上述只读端点缓存该 Token 的成功结果，最多 **30 秒**，且不会超过 Access Token 的 `exp`。缓存键为 Token 的 SHA-256 摘要、单进程内存有界，不持有 JWT 签名共享密钥；失效或失败不缓存。**残留风险**：若 Supabase 账号被删除或远端已拒绝某个 Token，快路径中仍已缓存的成功结果可能在余下 ≤30 秒内继续允许读取；写操作/其他敏感接口每次请求仍访问 Auth。普通客户端退出会清除本地凭据，但不能据此宣称已签发的 Access Token 立即全局失效。
- 对 **ES256 / RS256** Access Token：使用 Supabase 的公开 JWKS 验签并校验 `iss`、`aud`、`role`、`exp`、`iat`、`sub`。公钥缓存最多 10 分钟；未知 `kid` 触发受限刷新。允许 30 秒时钟偏差仅用于签发时间等校验，过期时间仍严格拒绝。空/异常 JWKS 不覆盖已有密钥集；JWKS 故障期间失败关闭，最多 30 秒负缓存限制重试洪峰。
- **排查办法**：先在本机已登录客户端内仅检查 Access Token *header* 的 `alg` / 是否有 `kid`；不要把完整 Access Token、Refresh Token、Authorization Header、签名密钥、完整账号会话截图写入日志或提交 PR。出现只读接口 `503` 时，检查 JWKS 网络与时钟/NTP；写接口 `503` 则优先检查远端 `/auth/v1/user` 可达性。必要时回滚该 PR 对应后端二进制，恢复全部接口远程验证，不需要修改 Supabase 密钥或清理数据库。
- 性能验收以部署后客户端 RTT / API Auth 耗时 / 数据库耗时 **P95** 为准。CI PASS 不能代替实机延迟与真实签名算法确认。

## 6. 回滚
Go 服务可执行文件更新前备份上一版；需要回滚时恢复对应二进制并 `sudo systemctl restart una-seller-api`。Nginx 每次修改前备份新建的 `mkd` vhost，`nginx -t` 失败绝不 reload；已有其他站点保持原样。不要清空 Supabase 数据库或修改其他 Docker 容器。

**完成状态判定：** 仓库 CI 通过 = `CODE_READY`；Ubuntu 服务器进程、DNS、TLS、客户端 Bearer 读回均通过 = `CLOUD_DEPLOYED / PASS`。目前只提交代码与部署模板，尚无后者的实测证据。
