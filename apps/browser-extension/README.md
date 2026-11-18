# Una Seller 浏览器采集扩展 — M0-03 / PR-C 第一切片

- 适用 **Chrome 102+（Manifest V3）**。Edge 实机兼容性待验证。在 `chrome://extensions` / `edge://extensions` 开启开发者模式，选择“加载已解压的扩展程序”，选中此目录。
- 点击扩展图标，使用**现有 Una Seller / Supabase Auth 邮箱密码**登录一次。公开注册关闭不影响既有账号登录。密码仅用于一次身份请求，不写入插件存储；登录会话保存在扩展本地隔离存储（设置为 `TRUSTED_CONTEXTS`），不会暴露给 Amazon 页面。
- 打开 `https://www.amazon.com/` 搜索或类目列表。页面识别到有效商品卡片后显示 **采集** 按钮；详情页也可点按钮。**没有用户点击就不会请求商品详情或发送采集数据**。
- 单击按钮后，扩展后台请求 Amazon.com 商品详情（不切换当前标签页），内容脚本解析标题、页面显示价格、图片、分类、普通描述、Bullet Points；向已鉴权的 `POST /api/v1/collections` 提交结构化 JSON。仅后端返回 200/201 并有保存记录 ID 时显示“已采集/部分成功”；后端 202 显示“待核验”，失败显示“重新采集”。
- 当前不从 Amazon 组合/猜测变体，遇到规格选择会标记 `variants_not_exhaustive` 和部分成功。更完整的页面状态、变体价格与图片关系作为下一切片。遇验证码、登录墙、403/429、网络错误停止，用户手动解决后再次点击。
- 扩展仅授权访问 Amazon.com、Una Seller 公开业务 API 和 Supabase Auth。扩展不调用浏览器 Cookie API。后台请求 Amazon 商品详情使用 `credentials: "include"`，浏览器可能随请求向 Amazon 发送该站点的 Cookie，具体取决于浏览器的 Cookie 策略。扩展不读取浏览历史、卖家账号、评价或销量，不自动批量采集，也不把网页 HTML 发送到业务 API。
- 公开业务 API 地址和 Supabase publishable key 可用于扩展代码。服务器私有 IP、内部主机名、服务端密钥、Supabase service-role key 和真实访问或刷新 Token 不得提交。
- 无法将会话存储限制为 `TRUSTED_CONTEXTS` 时，扩展会提示错误并停止登录、读取或使用会话。
- 需要云端 Go API 部署包含 **受控 Chrome extension Origin CORS** 的当前 PR 代码；源站还使用旧 Go 二进制时，插件提交可能被拒绝。**禁止直接在 CI 部署生产**，完成代码合并与现场授权后才更新 Go 二进制。
- 桌面端采集箱当前列表可查看新记录。跨账号数据不可见由 Go API 按经过 Supabase 验证的用户 ID 隔离。
- 安装后的真实 Amazon.com 页面、不同地区/邮编、多规格及 Win/macOS Chrome/Edge 实机运行均须实测，本 PR 的语法/单测不等于这些验收已经 PASS。

代码职责：`amazon-core.js` 只解析公开 DOM 字段；`content.js` 只注入按钮并在点击后发送消息；`background.js` 持有身份会话并请求 ERP API；`popup.*` 只用于登录与退出。无第三方运行依赖或构建服务器。
