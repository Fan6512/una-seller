# Issue #57 — Windows 产品箱隔离验收

## 测试包与正式版的区别

- #57 业务代码来自已合并的 PR #59。正式版 Tauri 默认使用现有云 API，而云端还没有部署产品箱新接口；直接用旧云 API 无法验证保存、转库、读回。
- 本分支只把**非正式、未签名的 NSIS 测试构建**指向本机 http://127.0.0.1:8787，配合 Artifact 内独立的 una-seller-api.exe。正式签名发行、Updater、Nginx、生产数据库均不改变。
- 测试包与 Go API 必须来自相同 HEAD 的 [Windows Desktop Build](https://github.com/Fan6512/una-seller/actions/workflows/windows-desktop.yml) Artifact：una-seller-windows-x64。不要混用旧 #59 的云端构建。NSIS 安装程序**不包含** Go API EXE。

## 1. 本地一次性 PostgreSQL 17（或现有独立测试库）

如有 Docker Desktop，Windows PowerShell 示例：

~~~powershell
$QaPassword = [guid]::NewGuid().ToString("N")
docker run -d --name una-seller-issue57-qa -e POSTGRES_USER=postgres -e "POSTGRES_PASSWORD=$QaPassword" -e POSTGRES_DB=una_seller_issue57 -p 127.0.0.1:55457:5432 postgres:17
$env:DATABASE_URL = "postgresql://postgres:$QaPassword@127.0.0.1:55457/una_seller_issue57?sslmode=disable"
$env:SUPABASE_URL = "https://rwktwfxzfkzglxwaulcm.supabase.co"
$env:SUPABASE_PUBLISHABLE_KEY = "<你的项目公开 Publishable Key>"
$env:UNA_SELLER_CLOUD = "0"
# 在已解压的 Windows Go API EXE 所在目录执行：
.\una-seller-api.exe migrate
.\una-seller-api.exe
~~~

**必须在执行迁移前确认 DATABASE_URL 只指向上述隔离数据库**，绝不可使用生产 Supabase 数据库。环境变量优先于 %LOCALAPPDATA%/UnaSeller/api-config.json，但不要假设旧配置是安全的。不要在聊天、Issue 或 Git 中公开连接字符串及密码。

没有 Docker 的话可改用已有专用、非生产 PostgreSQL 17 测试库；如果没有隔离库，标记 BLOCKED_TEST_DATABASE，不要修改生产库。

## 2. 用独立测试账号添加三条合成采集来源

在另一个 PowerShell 窗口，从 QA 分支工作目录运行：

~~~powershell
powershell -NoProfile -File .\apps\desktop\tests\issue57-seed-qa.ps1 -IsolatedDatabaseConfirmed
~~~

脚本要求确认隔离 DB、仅向本机 8787 写入，使用隐藏密码输入向 Supabase Auth 取得测试账号会话，不向日志输出 Token。它写入英语、西语、葡语三条**合成** Amazon 来源记录；可以重复执行（同 URL 去重）。务必先让本机 Go API 连上非生产数据库且通过身份验证。**不要拿真实客户数据测试。**

## 3. Windows GUI 必验

从 QA 运行下载 Artifact，使用其中的 una-seller.exe，登录同一测试账号。测试版应用与正式版共享 Tauri app identifier；建议使用独立 Windows 测试账户，避免影响已保存的桌面会话。逐项记录 PASS / FAIL / NOT_RUN：

1. 工作台采集箱展示三条来源，图片卡片、主图点击、勾选框不误开详情；最小宽度 780px 不遮挡操作。
2. 英文、西班牙语、葡萄牙语三栏按来源语言正确预填，可独立编辑/复制，字符统计包含空格及标点。
3. 新采集不自动选 CBT 类目；编辑完英文标题后按需推荐；未配置 OAuth 时提示服务不可用，不当作真实官方推荐验收通过。
4. 人工填写采购成本、重量 g、长宽高 cm、SKU、属性及图片 URL，保存后关闭重开仍保持。
5. 勾选多条来源并通过，新增产品箱条目；已转来源只读，重复通过显示“此前已转入”而不是重复成品。
6. 产品箱独立修改标题、描述、图片和变体，保存后重启读回一致；源商品原始数据不被改写。
7. 删光图片后卡片不恢复来源缩略图；多窗口旧 revision 显示 409；失败时没有虚假的成功提示。
8. 搜索、翻页、切换标签时不携带不可见的旧勾选；未保存的表单切换时明确提示。
9. 本地 Go API 与数据库重启后再次打开程序，已保存商品仍存在。

只附脱敏 UI 截图和非敏感错误码；不可公开 Token、用户邮箱、密码、服务器地址里的秘密、客户订单信息。

## 4. 仍然独立的验收门槛

- 真实 CBT OAuth 类目推荐需要用户授权的测试店铺与受控环境另测；无店铺时的 412/503 提示通过不等于官方联调 PASS。
- NSIS 编译成功不等于 Windows GUI 真机 PASS。
- 生产 0008/0009、Go API、Nginx、正式 Windows tag/Release 必须另行备份、审核、批准；本测试 PR 不具备部署授权。
