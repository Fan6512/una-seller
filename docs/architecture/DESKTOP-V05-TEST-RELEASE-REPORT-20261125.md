# V0.5.0 Windows 签名测试版发布报告

> 日期：2026-11-25（北京时间） · 范围：签名测试版 Pre-release，不设置 Latest，不改变正式版渠道
> 结论：PR #93 合并后 main CI 通过；Tag 与生产环境审批沿用 V0.4 流程；签名安装包、`.sig`、`latest.json` 全部核验通过；Release 已发布为 Pre-release。客户端实际 OTA 与功能验收由负责人人工执行。

## 1. 发布对象与门禁

| 项 | 结果 |
| --- | --- |
| 合并 PR | [#93](https://github.com/Fan6512/una-seller/pull/93) `chore(desktop): prepare V0.5 Windows test release`，MERGED 2026-11-25 03:41:38Z |
| 固定发布提交 | `c55b32271ce1f35b08d3339bd86e28d226fee745`，合并后 `origin/main` HEAD 一致 |
| 合并后 main CI | [run 45541730745](https://github.com/Fan6512/una-seller/actions/runs/45541730745) success：`Desktop Frontend` PASS、`Changed Files` PASS；`Go Backend` / `Browser Extension` 因本 PR 只改 5 个版本文件被范围化 CI 跳过，符合 PR 声明的无业务/后端/依赖变更 |
| 测试 Tag | `desktop-test-v0.5.0`，轻量 Tag，指向上述固定提交，已推送 origin |
| 签名构建 | [run 45542860330](https://github.com/Fan6512/una-seller/actions/runs/45542860330) success：`production-environment` 门禁 PASS，`Windows Tauri x64 (NSIS)` 签名构建 PASS，`Verify and stage GitHub Release draft` PASS |
| 环境审批 | `desktop-production`（required reviewer `Fan6512`，`can_admins_bypass=false`）；本次部署 `8047848525` 已批准，与 V0.4 同一审批环境 |
| 签名密钥 | 更新公钥 Key ID `c69c6a13a682d18b`，与 V0.4 安装包签名 Key ID 一致，沿用 V0.4 生产签名密钥对 |
| Release | [desktop-test-v0.5.0](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.5.0)，`draft=false`、`prerelease=true`，发布于 2026-11-25 04:02:18Z |
| 正式渠道 | `releases/latest` 仍为 `desktop-v0.1.1`（`prerelease=false`），未设置 Latest，未发布正式版 |

## 2. 安装包、签名、清单与版本一致性

| 资产 | SHA-256 | 结果 |
| --- | --- | --- |
| `Una-Seller_0.5.0_x64-setup.exe` | `898ec6ae2e2b951bb36d7e0bc9f7d201c58a5f64b05f876978204c4c0f3e6c9c` | 与 GitHub Release 资产 digest 一致；公网下载 HTTP 200，大小 3,182,870 字节 |
| `Una-Seller_0.5.0_x64-setup.exe.sig` | `70e4edcbdd0aadf2f13ef685d81fe56ff4e85fcac410160981fa8028fdda4e74` | 与 Release digest 一致；公网下载 HTTP 200 |
| `latest.json` | `b3c920370381deffc4bc2dcf5879436839031322e5c9a44134232bb29596a38f` | 与 Release digest 一致；公网下载 HTTP 200 |

- 签名独立核验：解码 `updater.pub.b64` 得到 Ed25519 公钥，按 minisign 规则对安装包做 BLAKE2b-512 预哈希后 `crypto.verify` **PASS**；CI 内 `minisign -Vm` 亦 PASS。
- 签名 trusted comment：`timestamp:1795579155`、`file:Una Seller_0.5.0_x64-setup.exe`、`version:0.5.0`，版本绑定正确。
- `latest.json`：`version=0.5.0`；`platforms.windows-x86_64.url` 指向本 Tag 安装包；`signature` 字段与 `.sig` 资产逐字一致。
- 五处版本文件全部为 `0.5.0`：`tauri.conf.json`、`package.json`、`package-lock.json`、`Cargo.toml`、`Cargo.lock`，与 Tag 及签名版本一致。

## 3. 公网 OTA 双通道

端点：`https://mkd.186512.xyz/api/v1/desktop/updates`，通道由请求头 `X-Una-Include-Prerelease` 选择；Go API 负责候选筛选与清单校验，Tauri 客户端负责密码学验签。

- 发布前基线：默认通道 `0` → 204；测试通道 `1` → 200（`0.4.0`）；两者均 `Cache-Control: no-store`。
- 发布后：默认通道 `0` → 204、`Cache-Control: no-store`，行为不变；安装包、`.sig`、`latest.json` 公网下载均为 200 且 SHA-256 与 Release 一致。
- 发布后测试通道复测结果见下方「最终复测」；Go API 按通道缓存 10 分钟，发布后旧通道缓存会随 TTL 自然过期刷新。
- 未重新部署 Go API，未重复构建未签名 QA 包。

### 最终复测（2026-11-25 04:08–04:09Z）

- 默认通道 `X-Una-Include-Prerelease: 0`：HTTP **204 No Content**、`Cache-Control: no-store`、空响应体，发布前行为不变。
- 测试通道 `X-Una-Include-Prerelease: 1`：HTTP **200 OK**、`Cache-Control: no-store`、`version=0.5.0`，`platforms.windows-x86_64.url` 指向 `desktop-test-v0.5.0` 安装包。
- 按 OTA 清单 URL 下载安装包与 `.sig`：HTTP 200，SHA-256 分别为 `898ec6ae…` / `70e4edcb…`，与 Release 资产一致；`latest.json` 公网下载 SHA-256 `b3c92037…`，一致。
- 发布后初期测试通道仍返回缓存中的 `0.4.0`；Go API 按通道缓存 10 分钟，未重启/未重新部署，TTL 过期后于 04:08:42Z 自动刷新为 `0.5.0`。
- 下载复测中出现一次瞬时连接错误，重试后 HTTP 200 且哈希一致，未修改 Release 资产。
- `releases/latest` 仍为 `desktop-v0.1.1`，正式渠道未受影响。

## 4. 负责人人工验证清单（客户端）

1. 现有 V0.4 客户端保持「接收测试版更新」关闭，点「检查更新」：应提示已是最新版（服务端 204）。
2. 打开「接收测试版更新」，重新检查更新：应出现 `0.5.0` 更新信息。
3. 下载并安装：观察进度、安装器启动、UAC 提示、覆盖安装是否完成。
4. 安装后界面版本显示为 `0.5 测试版`，重启后仍为该版本。
5. 登录状态与本地数据保留：账号、店铺授权、采集工作稿、已刊登列表、订单/财务设置不丢失。
6. 测试版开关重启后保持，已安装的 `0.5.0` 不再反复提示更新。
7. 异常路径：网络中断/超时有明确提示且不崩溃；如可模拟签名错误的清单，客户端应拒绝安装。
8. V0.5 业务变更抽测：订单 FIX-01/02/04–08、官方 CBT 账单预览、工作台变体 SKU 与共用副图、Amazon 详情布局修复，按负责人既有验收口径逐项确认。
9. 再次确认未误入正式渠道：正式版客户端（测试通道关闭）不应收到 `0.5.0`。

## 5. 未执行 / 待验收

- 未发布正式版、未设置 Latest、未改变正式版渠道。
- 未重新部署 Go API，未重复构建未签名 QA 安装包。
- 客户端实机跨版本 OTA 下载/验签/安装、重启与数据保留、业务功能验收：等待负责人人工结果。
