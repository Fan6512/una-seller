# Windows / macOS 双平台规范（v1）

- 支持范围：Windows 10/11 x64 为首要开发平台，macOS Apple Silicon arm64 和 Intel x64 均为首发目标；Windows ARM64 可以留待需求与构建验证后支持。
- 技术路线：共享 React/TypeScript UI + Tauri 2 Rust 外壳；两端访问同一套 Go API；不另写 Swift/C# 业务客户端。
- 状态：目标规范，尚未完成真实构建、签名、公证或设备验收。

## 1. 跨平台边界

| 能力 | Windows | macOS | 契约 |
| --- | --- | --- | --- |
| WebView | WebView2 | WKWebView | CSS/字体/输入/滚动双端回归 |
| 应用包 | NSIS exe 或 MSI | dmg / app | 同一应用版本和数据协议 |
| 用户密钥 | Windows OS Credential API | macOS Keychain | 仅保存 Una Seller 设备会话 |
| 文件系统 | Win 路径与权限 | macOS 沙盒/权限差异 | 经受限 Rust 文件接口访问 |
| 打印面单/PDF | Windows 打印栈 | macOS 打印栈 | 共用 PrintService，系统适配 |
| 快捷键 | Ctrl | Command | 行为统一，平台键位不同 |
| 自动更新 | Tauri Updater（如启用） | Tauri Updater（如启用） | 签名和降级保护，先验证再发布 |
| 本地缓存 | SQLite | SQLite | 数据契约相同，路径/库文件独立 |

不得将 Windows 注册表、PowerShell、Win32 打印调用或绝对盘符路径直接放入 React 业务模块；所有系统差异封装在 Rust adapter / interface。

## 2. 客户端适配接口（逻辑）

- `CredentialStore`：get/set/delete ERP 设备会话；不接触 Meli refresh token。
- `FileStore`：经用户授权打开、保存、导出、临时文件清理。
- `PrintService`：PDF/面单预览、打印机选择、错误反馈；各 OS 有独立的集成测试。
- `AppUpdate`：检查版本、下载、校验签名、展示升级失败与恢复方案。
- `PlatformInfo`：OS/arch、快捷键、权限与用户可见兼容性信息。

Tauri capabilities 按窗口和命令最小授权，默认不启用 shell 执行和任意文件系统访问。敏感插件进入发布构建前要做权限审查。

## 3. 本地数据与离线

- 初版仅缓存必要商品索引、草稿、用户界面状态；离线可以阅读部分缓存和草拟，但不得展示成已同步/已发布。
- 若后续需要 SQLCipher，应先验证 Windows/macOS 双系统原生编译、密钥生命周期、升级迁移与退出后删除；不能把普通 SQLite 当加密数据库。
- 多设备换机后恢复服务端已保存的数据；未同步草稿与本地打印机设置不自动恢复。
- 远程吊销使服务端会话立即失效；已离线设备的本地擦除只能在重新联网时执行，不能作即时保证。

## 4. 发布与测试矩阵

| 阶段 | Windows x64 | macOS arm64 | macOS x64 |
| --- | --- | --- | --- |
| 每次 PR | TS/Go/Rust 静态与单测（主跑 Windows/Linux CI） | 定期构建验证 | 定期构建验证 |
| M0 入口验收 | 安装、启动、登录、API 读链路 | 构建+启动冒烟 | 构建+启动冒烟 |
| 正式发行前 | 签名安装包、卸载与更新回归 | Apple 签名/公证与实机测试 | Apple 签名/公证与实机测试 |
| 关键业务 | 表格/输入/拖拽/导出/打印 | 同样全覆盖 | 同样全覆盖 |

使用托管 macOS Runner 可以远程构建 Mac 安装包，但不能替代 Mac 实机运行、Keychain 权限、签名/公证及打印验证。
签名凭据只通过仓库 Secrets/可信签名环境注入；PR 不使用生产签名。Windows EV 证书不是强制的架构依赖。

## 5. UI 一致性

采用统一的视觉设计系统、间距和响应式规范；尊重不同系统的菜单、窗口按钮、原生字体回退与快捷键习惯。不要为追求像素完全相同而强制复制操作系统的原生控件。
