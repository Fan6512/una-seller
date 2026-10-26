# ZYing（美客多 ERP）静态逆向分析报告

- **目标**：ZYing 1.0.1.1（Windows 桌面端，`target\ZYing`）及安装器 `target\ZInstall.exe`
- **分析日期**：2026-10-26
- **分析方式**：纯静态逆向。未运行目标程序，未连接 ZYing/MercadoLibre 服务器，未使用任何账号凭据。
- **工作区**：`E:\美客多逆向`
- **证据体系**：REA Evidence ID + 反编译源码行号 + 本地资源样本

> **脱敏声明**：本报告对 6 组内置 OAuth 应用凭据（AppId/AppSecret/RedirectUrl）、OAuth access/refresh token、登录账号等一律以 `<redacted>` 表示；仅保留算法、字段名、流程、文件位置与哈希，供复核。报告不包含可直接登录或接管账号的秘密值。

---

## 0. 结论摘要

1. **产品定位**：ZYing 是一个多平台跨境电商 ERP，美客多只是其中一个平台（站点编号 `8`），同时存在 **跨境 CBT、跨境 UP、本土** 三条美客多业务线；订单、刊登、同步、面单等核心操作由 ZYing 自有服务端代理执行，并非客户端直连 Meli 完成全部业务。
2. **双通道架构**：客户端用 **明文 TCP 私有协议**（默认端口 `7735`，美客多 2 线 `17735`）访问 ZYing 网关；同时用 **HTTPS + Bearer Token** 直连 `api.mercadolibre.com` 完成商品详情、消息、索赔、QA、图片/视频、发票、配额等辅助查询与操作。
3. **OAuth 流程**：用户在外部浏览器完成 Meli 授权，把 `code` 粘贴回客户端；客户端通过 `meli_acc.auth_token` 命令把 `code`/`appid` 交给服务端，并收到 token/userinfo 结果（换取发生在服务端为【推断】，客户端只观察到结果），再把 token 与店铺信息通过 `meli_acc.insert/update` **回传服务端保存**。客户端运行期通过 `meli_acc.detail`/`meli_acc.first` 取回 token 使用。
4. **传输安全缺陷（严重）**：ZClient 为裸 TCP，帧内包含 16 字节 UUID、16 字节 `Member.Token` 和全部业务数据；OAuth token 与店铺信息以明文传输，VPS 密码仅做 DES 可逆混淆。代码内虽存在 XOR 加解密函数，但**未被调用**。
5. **凭据硬编码（严重）**：客户端内置 6 组 Meli 应用配置，AppSecret 以字符串混淆形式存放在程序集内；登录/VPS 密码使用**固定 DES 密钥/IV**（密钥为 8 字节常量）。字符串混淆可被目标自身的解密器还原。
6. **授权模型**：许可完全由服务端 `Member` 记录驱动（`sites`、`allowsites`、`limit`、`ptime`、`level` 等），本地未发现离线签名校验；功能菜单由 `Member.AllowSite(8)` 与 `Limit.Meli` 位控制。服务端不返回许可，客户端无法正常登录使用。
7. **风控/配额**：实现了 15 分钟提前刷新 token、批量查询上限 20、429/500/502 指数式短重试、本地类目/属性缓存、刊登配额字段、Meli 错误中文化、KYC/账号限制识别等；**未发现** Meli 路径上的代理池、UA 轮换或 TLS 指纹伪装，属于“稳定性/配额控制”而非成熟反检测体系。
8. **本地数据**：没有发现本地订单/商品业务数据库；`.dat` 是类目、属性、远程地区等缓存；CefSharp 缓存中只有 1688 相关 Cookie，**没有 Meli Cookie**（Meli 授权走外部浏览器）。因此本地取证价值主要集中在配置、缓存与日志，而非账号凭据。

---

## 1. 范围、方法与证据

### 1.1 分析目标

| 项目 | 内容 |
|---|---|
| 程序目录 | `E:\美客多逆向\target\ZYing`（约 3,558 个文件，约 1.2 GB） |
| 安装器 | `E:\美客多逆向\target\ZInstall.exe` |
| 版本 | `1.0.1.1`（`version.ini`：`ProductVersion/UpdateVersion`） |
| 渠道 | `src_channel=guanwang`（官网渠道） |
| 运行框架 | .NET 6.0（`ZYing.runtimeconfig.json`：`Microsoft.NETCore.App 6.0.0` + `Microsoft.WindowsDesktop.App 6.0.0`） |
| UI | WinForms + CefSharp（存在 `libcef.dll`、`CefSharp.WinForms.dll`、`CefSharp.BrowserSubprocess.exe` 等） |

### 1.2 方法

1. `REA inspect-artifact` / `inspect-managed-artifact` 清点 PE、CLI 头、目标框架、程序集引用与哈希，生成 Evidence ID。
2. 使用 ILSpyCmd 对 `ZYing.Core.dll`、`ZYing.Work.dll`、`ZYing.User.dll`、`ZYing.UI.dll`、`ZYing.exe`、`ZDto.dll`、`ZPdf.dll` 反编译；对美客多相关 131 个 User 类型做重点阅读。
3. 解析自定义 `.dat` 容器（`analysis/scripts/dat_dump.py`），离线读取 `ZOther.dat`、`remotes.dat` 等条目。
4. 通过 pythonnet 加载目标程序集的字符串解密器，按调用点还原字符串；工作程序集还原 42,763 条、用户程序集还原 30,601 条，美客多相关 10,848 处调用点 0 缺失（`analysis/scripts/resolve_strings*.py`）。
5. 对每个结论标注【观察】【推断】【未知】，并在附录 D 给出源码行号索引。

### 1.3 约束与未做事项

- **未运行** `ZYing.exe`、`ZInstall.exe`、`CefSharp.BrowserSubprocess.exe`。
- **未连接** ZYing 网关或 MercadoLibre API；未进行抓包、重放、账号登录、token 刷新。
- **未解析** 服务端实现、数据库、密钥管理系统。
- **未使用**任何真实 AppId/AppSecret/token；报告中的凭据均脱敏。
- 未对安装器做完整解包/行为分析，仅做了 PE 与元数据层面的识别。

### 1.4 REA Evidence ID 清单

| Evidence ID | 主体 | SHA-256 | 用途 |
|---|---|---|---|
| `ev_16495473a33a1cc806acd98203792f25f3c5b1528cd3c2b746ce9dcf474636f9` | `ZYing.Core.dll` | `eade294d287844eef166642cabf2b8f72dbd5821346856aae2e58549b77e84e2` | 工具/加密/HTTP 基础设施 |
| `ev_d3246d32fb2b7f834baec22446bf737738d25cb8b122c22de7b92e1a6570823d` | `ZInstall.exe` | `847f72287ea7379b2f13ec6f06eb198ab342393615c4b818f3cc261aaa7613a8` | 安装器产物图 |
| `ev_5f0f3affb404d6bbaf6b4c8de3260cdc01092a9c219d97168566d8c43d9efa79` | `ZInstall.exe`（第二证据） | 同上 | 安装器安装行为/载荷关联 |
| `ev_c4614b9d8d78336e256caaebc0b45bd3154c61c94106575528dd6f384c62815a` | `ZYing.exe` | `ed18084439eeed731fce479bfc9f860ca4477e3be7475cee8bf5620c9e349bef` | 启动器（x86_64） |
| `ev_d98e55312a0eba5e9768a5588a2f0e277bba749ebe46e8508248aaaa8c671447` | `ZYing.Core.dll`（成员检查/加载路径） | `eade294d...b77e84e2` | Core 加载与控制流证据 |
| `ev_f7ee4e0ff6cd9f9d06f461539efa276f19f9a975ca9e22e95f9c65dd4a264b81` | `ZYing.Work.dll` | `46c409fa58db72f22aaa6ae81c1144347411c74c0f2371ab4d5713158aeb0f18` | 协议、MeliApi、Member |
| `ev_2884f51d6a15bb364cb57ad06031178f8261bb434977fdc18eb5c8b759e289f9` | `ZYing.User.dll` | `1c34feaf7cb71b55177e196b29bdeeece690f38e0239455b0aedb921fb7ab50d` | 美客多 UI/业务编排 |
| `ev_ae4916bb086fa90ef8dbd6d5b97427636218925474c615240a7d9a9507f7fec3` | `ZYing` 程序集 | — | 主程序集（4 类型/14 方法，AnyCPU） |
| `ev_e34c18631d1abbe885609005be4ffb7628fd593cf312822cbdc3a7f890ab79ac` | `ZYing.UI` 程序集 | `dbe5e8021473444db600937413b1ef373c541ccf8f91ce9de98c453e7376b816` | UI 框架 |
| `ev_ca1d88228f7d3489c787b36e260b954396f1b62c2f520c41ccb9e23cdb697d44` | `ZDto` 程序集 | `634834876e847c499d19431ef410cf6f7dc67661c88b18bb023b4c1f20817fde` | DTO/Record 序列化 |
| `ev_9de3770c53b8d4952bd013caf182d15be72258a72bfb589553d2eb8791bdd6fd` | `ZPdf` 1.0.1.4 | `5b679d73e54015e5b843c3adc352cf78fa38908895b308a287445c01efaec8c3` | 面单/PDF 处理 |

> Evidence ID 由 REA 报告生成；同一文件可能出现两个 ID（不同分析操作），见 `analysis/evidence_ids.txt` 与 `analysis/evidence_map.txt`。

---

## 2. 产物与技术栈

### 2.1 主程序集

| 文件 | 大小 | SHA-256 | 角色（观察/推断） |
|---|---:|---|---|
| `ZInstall.exe` | — | `847f72287ea7379b2f13ec6f06eb198ab342393615c4b818f3cc261aaa7613a8` | 安装器；REA 标记为 PE |
| `ZYing.exe` | 437,248 | `ed18084439eeed731fce479bfc9f860ca4477e3be7475cee8bf5620c9e349bef` | 启动器（x86_64） |
| `ZYing.Core.dll` | 953,856 | `eade294d287844eef166642cabf2b8f72dbd5821346856aae2e58549b77e84e2` | 基础设施：`ZClient` 依赖、`Safe`、`HttpClientEx`、`HttpAgent`、`Dat` |
| `ZYing.Work.dll` | 8,437,248 | `46c409fa58db72f22aaa6ae81c1144347411c74c0f2371ab4d5713158aeb0f18` | 业务核心：`Member`、`ZClient`、`MeliApi`、`MeliAppInfo`、`MeliApiBuilder`、`MeliErrorParser` |
| `ZYing.User.dll` | 16,710,144 | `1c34feaf7cb71b55177e196b29bdeeece690f38e0239455b0aedb921fb7ab50d` | 美客多 UI/业务编排（登录、授权、刊登、订单、消息、索赔、QA、同步、面单） |
| `ZYing.UI.dll` | 2,321,920 | `dbe5e8021473444db600937413b1ef373c541ccf8f91ce9de98c453e7376b816` | WinForms 控件/主题框架 |
| `ZDto.dll` | 4,096 | `634834876e847c499d19431ef410cf6f7dc67661c88b18bb023b4c1f20817fde` | Record/RecordSet DTO；体积很小，可能为兼容/转发程序集【推断】 |
| `ZPdf.dll` | 14,203,904 | `5b679d73e54015e5b843c3adc352cf78fa38908895b308a287445c01efaec8c3` | PDF 打印、面单合并 |

> `ZYing.Work.dll` 的类型列表可见 `Class ZYing.Member`、`Class ZYing.ZClient`、`Class ZYing.Meli.MeliApi` 等（`analysis/work-types.txt:593/667/1743`）；`ZYing.User.dll` 包含 `ZYing.Menus`、`Logic.MeliLabelPrint`、`ZYing.Auto.ZSyncMeli` 等（`analysis/user-types2.txt`）。

### 2.2 依赖与运行环境

- .NET 6.0（`net6.0`，`Microsoft.WindowsDesktop.App`）。
- CefSharp：`CefSharp.dll`、`CefSharp.Core.dll`、`CefSharp.WinForms.dll`、`CefSharp.BrowserSubprocess.exe`、`libcef.dll`、`chrome_*.pak`。
- `ZPdf.dll` 自带 PDF 处理能力；面单流程同时使用 `PdfSharp` 类型的 API（`analysis/decrypted/user/Logic.MeliLabelPrint.resolved.cs`）。
- 安装目录内存在 `appconfig.json`、`Server.ini`、`version.ini`、`Password.ini`（可选）等运行配置。

### 2.3 代码保护

| 保护 | 观察 | 评价 |
|---|---|---|
| 自定义 `.dat` 容器 | 版本号 + 头长度 + 条目表 + GZip 块 | 仅是资源打包，不是加密 |
| 字符串加密 | `EriwwTjw3(int offset)` + 每程序集密文块 + 模块字段 XOR | 可被目标自身解密器还原，无完整性保护 |
| 控制流混淆 | 部分类型/字段名被重命名，`<Module>{GUID}.m_*` 字段作密钥 | 增加阅读成本，不构成强保护 |
| 强壳/虚拟化 | 未观察到 VMProtect/Themida 类壳；程序集可直接反编译 | 保护强度低 |
| 签名校验 | 未在登录/授权路径发现本地签名校验 | 许可依赖服务端【观察+推断】 |

---

## 3. 整体架构

### 3.1 分层

```
┌──────────────────────────── ZYing.User.dll / ZYing.UI.dll ────────────────────────────┐
│ WinForms 界面 · 美客多刊登/订单/消息/索赔/QA · CefSharp 浏览器 · 本地缓存(.dat)         │
│ 业务编排：ZSyncMeli / ZOrderMeli / MeliLabelPrint / Meli授权详情 / 各 Channel 列表     │
└───────────────┬───────────────────────────────────────────────┬───────────────────────┘
                │ ZClient：裸 TCP（默认 7735 / 美客多2 17735）    │ HTTPS + Bearer
                │ 4B LastUpdate · 16B UUID · 16B Member.Token    │ api.mercadolibre.com
                ▼                                                ▼
      ZYing 服务端网关（47.x / 120.x 多节点）              MercadoLibre REST/OAuth
      账号、订单、刊登、消息、面单、同步、配额             /items /user-products /claims …
                │
                └── 服务端持有 Meli access/refresh token，可代客户端操作【推断】
```

### 3.2 双通道网络模型

| 通道 | 客户端实现 | 目标 | 数据 | 保护 |
|---|---|---|---|---|
| ZYing 私有通道 | `ZClient`（`TcpClient`/`NetworkStream`） | ZYing 网关（多 IP/端口） | 登录、会员、账号、订单、刊登、同步、面单、消息/索赔/QA 列表等 | 无 TLS、无签名；仅 UUID + `Member.Token` 校验 |
| Meli REST 通道 | `MeliApi` + `HttpClientEx` | `https://api.mercadolibre.com` | 商品/用户产品/类目/图片/视频/发票/索赔/消息/QA/配额等 | HTTPS + OAuth Bearer；`X-API-Version: 2`（部分端点） |

> **【观察】** `MeliApi.SendCoreAsync` 在每次请求前 `await CheckToken()`，再设置 `Authorization: Bearer <token>`（`MeliApi.resolved.cs:1088-1096`）。
> **【推断】** 核心业务选服务端代理而不是客户端直连，目的包括：集中管理 token、统一限流/配额、隐藏 Meli 调用细节、便于多店铺批量操作。

### 3.3 ZYing 网关节点（静态字段）

| 字段/用途 | 地址 | 端口 |
|---|---|---|
| 主线路-E 区 | `47.119.36.166` | 7735 |
| 备用线路-F 区 | `47.106.171.116` | 7735 |
| 香港线路（`appconfig.hk=true`） | `47.57.14.223` | 7735 |
| 美客多/站 | `120.78.181.98` | 7735 |
| 美客多 2 | `120.78.181.98` | 17735 |
| 其他静态字段 | `47.240.225.71`、`47.112.199.195:22`、`120.78.4.252:27735`、`120.77.69.30:27735`、`120.78.4.252:5377`、`8.217.74.159` | 见字段 |
| 默认端口 | — | 7735 |

证据：`analysis/decrypted/ZYing_ZClient.resolved.cs:66/77/226/1241-1267`。

`LoadBalancedIPGroup` 维护节点冷却与切换；`AppConfig.HK` 决定使用香港线路还是内地主线（`ZYing_ZClient.resolved.cs:1244-1245`）。

### 3.4 启动/登录/会话生命周期

1. 启动 `ZYing.exe` → 加载 .NET 6 WinForms 主界面与 CefSharp。
2. 登录：`common.login`，载荷含登录名、`Safe.Encrypt(login_pwd)`、可选 2FA；成功返回 `Member` 记录（`MLogin.resolved.cs:359-413`）。
3. 服务端下发 `Member.Token`、`sites/allowsites`、`limit`、`ptime` 等；客户端据此构建菜单与权限。
4. 每次 ZClient 请求写入 `LastUpdate`、`UUID`、`Member.Token`、命令名、载荷；响应校验 UUID/token，不匹配抛 `NoLoginEx("客户端标识符不匹配！")`（`ZYing_ZClient.resolved.cs:874-943`）。
5. 空闲连接 `CheckExpired` 在 300,000 ms 后关闭；登出广播 `common.logout`（`ZYing_Member.work.resolved.cs:794` 起）。
6. Meli token 过期前 15 分钟自动刷新（`MeliApi.CheckToken`，`MeliApi.resolved.cs:1123-1143`）。

---

## 4. 本地数据与配置

### 4.1 `.dat` 容器格式

```
offset 0: version   (int16, little-endian)
offset 2: headerLen (int16)
offset 4: header[headerLen]
header 每项:
  nameLen (uint8)
  name    (UTF-8)
  start   (int32)
  length  (int32)
data: 每个条目为独立 GZip 块（观察：多数以 1F 8B 开头）
```

解析脚本：`analysis/scripts/dat_dump.py`；输出：`analysis/dat/`。

### 4.2 资源清单（抽样）

| 文件 | 版本 | 条目数 | 关键条目/大小（按 .dat 头记录） | 用途 |
|---|---:|---:|---|---|
| `ZOther.dat` | 97 | 29 | `ATTRMELI-CBT` 2,353,978；`ATTRMELI-MCO` 2,092,389；`ATTRMELI-MLA` 2,298,604；`ATTRMELI-MLB` 2,263,518；`ATTRMELI-MLC` 2,082,901；`ATTRMELI-MLM` 2,435,099；`ATTRMELI-MLU` 2,020,145；各站点 `CBT/MCO/MLA/MLB/MLC/MLM/MLU` 约 200 KB | Meli 类目/属性缓存 |
| `remotes.dat` | 3 | 1 | `COMMONREMOTEAREAS` 1,989,087 | 邮编/偏远地区数据 |
| `ZData.dat` | — | — | 业务基础数据 | 通用数据缓存 |
| `ZAmazon.dat` / `ZEbay.dat` / `ZForm.dat` / `ZCaptcha.dat` | — | — | 各平台资源 | 多平台支持 |

> `ATTRMELI-*` 由 `EPropHelper.GetMeliProp(abbr)` 读取，底层 `new EPropHelper("attrMeli-" + abbr, Dat.ZOther)`；`Save()` 写回 `ZOther.dat`（`EPropHelper.work.resolved.cs:515/341/421`）。这是一种**本地类目/属性缓存**，用于减少 Meli API 调用。

### 4.3 运行配置

**`appconfig.json`（关键字段）**

| 字段 | 观察值 | 含义（推断） |
|---|---|---|
| `hk` | `false` | 是否走香港线路 |
| `usemeliserver` | `false` | 是否使用 Meli 专用服务端 |
| `useprivateserver` | `false` | 私有服务器开关 |
| `nocatalogitem` / `nooffical` / `nodepot` / `nolocal` | `false` | 上架过滤开关 |
| `onlybuy` / `onlychina` / `onlytitle` | `false` | 采集/上传过滤 |
| `replacepair` | `true` | 替换配对 |
| `parts` / `meliparts` | `"1"` / `"1"` | 分片/美客多分片 |
| `ordersize` / `salesize` | `100` / `240` | 订单/销售分页大小 |
| `alerttime` | `2026/10/26 9:05:35` | 提醒时间 |
| `notalert` | `true` | 免打扰 |
| `melipara.abbrnames` | `["墨西哥"]` | 默认站点简称 |
| `melipara.notloadcates` | `false` | 是否不加载类目 |
| `melipara.select_weight` | `false` | 是否选择重量 |
| `melipara.add_freight_enable` | `false` | 是否加运费 |
| `melipara.except_sites` | `["8","1"]` | 排除站点（8=Meli） |
| `melipara.include_lans` | `["CNY"]` | 允许币种 |

**`Server.ini`**：保存上次登录名，**明文**（示例值脱敏为 `F***1`；`MLogin.resolved.cs:203/367`）。
**`Password.ini`**：若存在则由 `MyText.Load(path)` 读取，**明文**；否则用登录记录中的 `login_pwd` 再 `Safe.Encrypt`（`MLogin.resolved.cs:359-365`）。
**`version.ini`**：`ProductVersion=1.0.1.1`、`UpdateVersion=1.0.1.1`、`src_channel=guanwang`。

### 4.4 CefSharp 缓存

- 路径：`target\ZYing\cache\temp\Cookies`（SQLite）。
- 观察到的 hostname：`.1688.com`（10 条）、`insights.1688.com`（1 条）、`.mmstat.com`（1 条）。
- **未发现 MercadoLibre 域 Cookie**。这与 Meli 授权走外部浏览器、客户端只接收 `code` 的设计一致。
- 该 SQLite 中 1688 会话 Cookie 为明文存储，属于本地敏感数据；本次分析未导出、未外传。

### 4.5 本地业务数据

- 未发现订单/商品/消息的本地数据库文件；这些数据由 ZYing 服务端保存，客户端按需分页拉取。
- 因此，仅靠本地文件无法还原店铺经营数据；需要服务端账号或抓包（本次未做）。

---

## 5. 字符串加密与协议保护

### 5.1 字符串解密机制

- 调用形态：`NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs.EriwwTjw3(<literal> ^ <Module>.m_<field>.m_<field>)`。
- 每个程序集内嵌一段密文 blob；解密器用 **自定义 32 字节 keystream + 4 字节块 XOR**（`KaqhrG2c6`/`flsFiRVeO`）还原明文字符串。
- 调用点偏移由字面量与 `<Module>{GUID}.m_<field>` 字段值 XOR 得出；模块字段相当于“密钥表”。
- 工作程序集：blob `vVrIijkpioRfkj7UNA.DwQZSGj8wTSNsxJm1B`，989,448 字节，42,763 条字符串；模块 GUID `{0e325ee4-ecc9-4c54-83c5-9fd3228c8105}`，持有字段 `m_5b4ad7c38cba4ec7baf5f148600dc375`，104 个偏移字段。
- 用户程序集：blob `4GyPb9bkfFGNKfxWD3.2huDA4q90C17H9chKH`，759,646 字节，30,601 条字符串；模块 GUID `{51cba30a-1899-472d-b922-810a0384b6c4}`，持有字段 `m_3803de857ccf44d38d28afc17036c502`，108 个偏移字段。
- 美客多相关 131 个 User 类共 10,848 处调用点，全部解析成功（0 missing）。

**评价**：这是自研的轻量字符串混淆，不是密码学保护。密钥、偏移、密文全部在本地，攻击者可直接调用目标解密器还原；也没有防篡改/完整性校验。

### 5.2 ZClient 传输协议

请求写入顺序（`ZYing_ZClient.resolved.cs:874` 起）：

```
[4B LastUpdate] [16B UUID] [16B Member.Token] [WriteName(Command)] [payload...]
```

- `Command` 为 ASCII 命令名（如 `meli.submit`），`payload` 为 `Record`/`RecordSet` 序列化数据（键值列）。
- 响应校验返回的 UUID 与 `Member.Token`；不一致抛 `NoLoginEx("客户端标识符不匹配！")`（`ZYing_ZClient.resolved.cs:943`）。
- 响应支持 GZip/Brotli 解压；`CheckExpired` 空闲 300,000 ms 关闭连接。
- `ZClient.Encrypt`/`Decrypt`（XOR key `{11..19}`）存在但**无调用点**（`ZYing_ZClient.resolved.cs:904-925`）。
- 无 TLS、无消息认证码、无请求签名、无时间戳防重放（`LastUpdate` 仅作为字段写入，未观察到校验逻辑）。

### 5.3 Meli 通道

- `HttpClientEx` 固定 UA：`Mozilla/5.0 (Windows NT 10.0; Win64; x64) ... Chrome/149.0.0.0 Safari/537.36 Edg/149.0.0.0`（`HttpClientEx.core.resolved.cs:46/64`）。
- `Accept-Encoding: gzip, deflate`；超时 60/120 秒。
- TLS 配置：`Tls | Tls11 | Tls12`，且 `ServerCertificateCustomValidationCallback => true`（接受任意证书，`HttpClientEx.core.resolved.cs:70-71`）。
- `MeliApi` 使用 `HttpClientEx.Create(120)`（`MeliApi.resolved.cs:1160`）。
- 未发现 Meli 路径使用代理；`HttpAgent` 的 132 个 UA 主要用于其他平台/采集场景（未在 MeliApi 调用链中发现）。

### 5.4 未使用的 XOR

`ZClient` 中的 XOR 加解密函数为死代码；实际协议是明文。这一点对“传输加密”判断很关键：**不能把存在加密函数等同于启用了加密**。

---

## 6. 美客多功能矩阵

### 6.1 站点与模式

| 模式 | 说明 | 证据 |
|---|---|---|
| CBT | 跨境卖家，站点标识 `CBT`；授权入口 `global-selling.mercadolibre.com` | `MeliApi.AuthUrls`；`MeliAppInfo.GetAppInfo(bool isCbt)` |
| 跨境 UP | 另一种跨境刊登/上传流程（`meli_up.*`、`meli_stat_up.*`、`meliup_sync.*`） | `ZYing.Menus` 菜单；命令目录 |
| 本土 | 本地站点店铺（MLM/MLB/MLC/MCO/MLA 等） | `meli.localsubmit`、`meli_localstat.*`、`MeliQa-本土` |

站点映射（`Enumer_Meli.resolved.cs`）：

| 站点 | 国家/域名 | 币种 |
|---|---|---|
| MLB | 巴西 `produto.mercadolivre.com.br` | BRL |
| MLM | 墨西哥 `articulo.mercadolibre.com.mx` | MXN |
| MLC | 智利 `articulo.mercadolibre.cl` | CLP |
| MCO | 哥伦比亚 `articulo.mercadolibre.com.co` | COP |
| MLA | 阿根廷 `articulo.mercadolibre.com.ar` | ARS |
| MLU | 乌拉圭 `articulo.mercadolibre.com.uy` | UYU |

授权 URL 仅覆盖 `CBT/MLM/MCO/MLB/MLC/MLA`（`MeliApi.resolved.cs:1145-1160`）；`MLU` 在枚举中存在，但授权入口未在 `AuthUrls` 中观察到；UI 对 MLU 有“已不支持乌拉圭市场”的确认提示（`ZYing.Detail.Meli授权详情.resolved.cs:1391` 附近）。**【观察/未知】** MLU 可能仅作为历史数据/站点映射保留。

### 6.2 授权与账户管理

- 入口：`Meli授权详情`（`ZYing.Detail.Meli授权详情.resolved.cs`，3,138 行）。
- `Btn登录_Click`：`Tool.Start(MeliApi.GetAuthUrl(siteId, "client_" + Environment.TickCount))`，用系统浏览器打开授权页（`:699-729`）。
- `Btn复制链接_Click`：复制授权 URL，便于在远程/VPS 浏览器中完成授权（`:755`）。
- `BtnAuth`：把用户粘贴的 `code` 与 `appid` 发给服务端：
  ```csharp
  zc.Command = "meli_acc.auth_token";
  nv["code"] = text;
  nv["appid"] = appInfo.AppId;
  await zc.PostAsync(nv);
  Record record = zc["root"];
  TokenInfo token = JsonConvert.DeserializeObject<TokenInfo>(record.Get("token"));
  UserInfo user = JsonConvert.DeserializeObject<UserInfo>(record.Get("userinfo"));
  ```
  （`:779-861`）
- CBT 额外解析 `marketsite` 为 marketplaces/logistics；本土设置站点/币种。重新授权时会拒绝店铺 ID 变化。
- `FreshBtn_Click`：若过期时间 < 当前 + 15 分钟，调用 `MeliApi.RefreshToken` 刷新，再 `GetMyInfo()`、`GetMarketSite()`、`GetListingsLimit()`，填充配额/总商品数/刊登类型（`:955-1010`）。
- `BtnClick`（确认提交）：组装 `Record` 并发送 `meli_acc.insert` / `meli_acc.update`（`:1269-1475`），字段包括：
  `acc_app`、`acc_password`（`Safe.Encrypt`）、`acc_shopid`、`acc_shop`、`acc_name`、`acc_expire`、`acc_access`、`acc_fresh`、`acc_email`、`acc_corp`、`acc_siteid`、`acc_cur`、`selectedabbrs`、`groupabbrs`、`siteinfo`、`acc_iscbt`、`acc_isfull`、`acc_siteuserid`、`acc_flag`、`acc_quota`、`acc_totalitems`、`acc_listingtype`、`acc_tags`、`acc_id`/`acc_group`。
- **关键观察**：`acc_access`/`acc_fresh` 直接以明文进入 ZClient 载荷；只有 VPS 密码 `acc_password` 做了 DES 混淆。

### 6.3 刊登/上传

- 入口：`Meli新版上传Detail`（2,882 行）、`Meli_Detail`、`Meli_UP_Detail`、`Meli本土上传Detail`。
- 类目选择：`meli_kind.select`（`:929`）。
- 提交：`BtnSubmit_Click` 取 `GetRecord()`（继承自基类，未在本次反编译范围内），追加业务字段后发送 `meli.submit`（`:1081/1235`）：
  - `accs`、`grades` 或显式商品 ID；
  - 可选 `start`/`end`；
  - `extra`（属性 JSON）、`stock`、`size_id`；
  - `net_proceeds_enabled`、`add_freight_enable`、`except_sites`、`include_lans`。
- 本地参数 `AppConfig.MeliSubmitPara` 持久化 `abbrnames`、`notloadcates`、`select_weight`、`add_freight_enable`、`except_sites`、`include_lans`（`:625/813/1269`）。
- 校验：标题前后缀不能含 `"./."`；重量/尺寸替换阈值（重量 ≥10g、边长 ≥2cm）；至少选择一个站点；可选商品等级过滤（来自 UI 校验逻辑）。
- 上传任务类型（`Enumer.Meli.MeliTaskType`）：上架、更新、删除、更新图片、更新价格、更新库存、更新重量尺寸、上架（UP）、下架、更新属性、更新标题、更新变体、获取产品。

### 6.4 同步、订单、面单

| 功能 | 命令 | 关键载荷 | 证据 |
|---|---|---|---|
| 产品同步 | `meli.sync` | `aid`、`startTime`、`endTime` | `ZYing.Auto.ZSyncMeli.resolved.cs:171-195` |
| 订单接收 | `meli.order` | `aid`、`startTime`、`status`（all/paid/cancelled/invalid） | `ZYing.Auto.ZOrderMeli.resolved.cs:165-189` |
| 面单打印（单个） | `meli.label` | `oid` | `Logic.MeliLabelPrint.resolved.cs:59-61` |
| 面单打印（批量） | `meli.label` | `oid`、`pick`（是否拣货单） | `ZYing.Dialog.打印Meli.resolved.cs:501-561` |
| 订单拆分 | `meli.order_split` / `meli.order_split_skus` | 订单/SKU | `ZYing.Dialog.Meli订单拆分` |
| 发票 | `meli.invoice` | 订单 | `ZYing.Detail.OP_` / `OP美客多` |

- 面单返回 base64 PDF：`zc.Message`（`Logic.MeliLabelPrint`）或 `base64` 列（批量）；本地用 PdfSharp 截取第一页，或叠加拣货单/包裹表（`SaveLableBytes`，`:569-695`），批量用 `PdfPrint.MergePdfs` 合并（`:379`）。
- 面单线路选择：`Member.转运服务 ? ZClient.美客多2 : ZClient.美客多`（`Logic.MeliLabelPrint.resolved.cs:57`）。

### 6.5 消息、索赔、QA

| 模块 | 命令 | 说明 |
|---|---|---|
| 站内信/消息列表 | `meli_message.` + action、`meli_message.fetch`、`meli_message.detail`、`meli_message.send_message`、`meli_message.save_trans_message` | 列表、拉取、详情、回复、保存翻译 |
| 索赔 | `meli_claim.` + action、`meli_claim.fetch`、`meli_claim.detail`、`meli_claim.refund`、`meli_claim.partial_refund`、`meli_claim.open_dispute`、`meli_claim.set_remark`、`meli_claim.send_message`、`meli_claim.save_trans_message`、`meli_claim.get_partial_refund_offers` | 退款、部分退款、开争议、备注、沟通、报价 |
| 售前 QA | `meli_question.pull`、`meli_question.delete`、`meli_question.detail`、`meli_question.reply`、`meli_question.memo`、`meli_question.noreply_count` | 跨境 QA |
| 本土 QA | `meli_qa.pull`、`meli_qa.delete`、`meli_qa.detail`、`meli_qa.reply`、`meli_qa.save`、`meli_qa.memo` | 本土站内信 QA |
| QA 模板 | `meli_qatemp.select/stat/insert/detail/delete` | 回复模板 |

证据：`ZYing.Channel.Meli消息列表.resolved.cs:573/739`、`ZYing.Channel.Meli索赔列表.resolved.cs:563/851`、`ZYing.Channel.Meli站内信.resolved.cs:503/665`、`ZYing.Channel.Meli本土Qa.resolved.cs:289/625`、`ZYing.Channel.MeliQA_.resolved.cs:373/439`。

### 6.6 价格、配额、绩效、跟卖、编码

| 能力 | 命令/端点 | 说明 |
|---|---|---|
| 价格调整 | `meli_stat.adjust_price`、`meli.stat.updatesale`、`meli_netproceeds.*` | 批量调价、净收益估算 |
| 配额/刊登限制 | `meli_acc` + `/marketplace/users/cap` | 账号配额、总商品数、刊登类型 |
| 店铺绩效 | `meli_metrics.fetch/detail` | 绩效列表/详情 |
| 违规/侵权 | `meli_stat.get_infraction`、`/marketplace/moderations/infractions/` | 违规查询 |
| 跟卖监控 | `meli_catalog_monitor.*` | 添加监控、调价、暂停、删除、详情 |
| 编码池/GTIN | `meli_code.*`、`Gtin` 菜单、`users.loadUPC` | 编码生成与池管理 |
| 类目/属性 | `meli_category.*`、`meli_kind.*`、`ATTRMELI-*` 缓存 | 类目选择、属性更新、参考 CBT |
| 图片/视频 | `meli_picture.*`、`meli_clips.*` | 图片上传、视频管理 |
| 运费/分类佣金 | `meli_freight.*`、`meli_rate.select` | 运费、佣金 |
| 映射/对照 | `meli_refer.*`、`meli_catemap.insert`、`meli_category.refer_*` | 分类映射、国家映射 |
| 日志 | `meli_log.*`、`meli_log_up.*`、`meli_locallog.*` | 操作日志/UP 日志/本土日志 |

### 6.7 菜单与权限

`ZYing.Menus.resolved.cs` 中与 Meli 相关的主要菜单：

- **美客多跨境**：跨境站刊登（`SubmitNew`）、已上传明细（`melistatv2`）、产品同步（`meliitem_sync`）、操作日志（`MeliLogListV2`）、售前 QA（`MeliQa`）、售后消息/索赔（`Meli_Claim_Message`）、海外仓库存（`Full_Stocks`）。
- **跨境 UP**：跨境 UP 刊登（`Submit_UP`）、已传明细（`melistat_up`）、产品同步（`meliup_sync`）、操作日志（`Meli_UP_LogList`）、图片上传（`meli_pic_upload`）、跟卖监控（`catalog_monitor`）。
- **美客多本土**：本土站刊登（`LocalSubmit`）、已上传明细（`LocalMelistat`）、操作日志（`MeliLocalLog`）、站内信 QA（`MeliQa-本土`）。
- **基本设置**：账户授权（`Acc`）、店铺绩效（`meliacc_metrics`）、价格调整（`MeliPrice`）、Meli 编码池（`MeliCode`）、编码生成（`Gtin`）。
- **参数设置**：Meli 平台分类（`MeliCatagory`）、Meli 运费、Meli 分类佣金、产品分类对照（`MeliMap`）、采集分类对照（仅 UserID `1117/1801/1051`）、分类国家映射（仅 `1117/1051`）、QA 回复模板（`MeliQa模板`）。
- **订单**：美客多订单（`OrderMeli`），条件 `Member.AllowSite(8) && Member.Limit.Meli.允许查看`。

**Meli 权限位**（`Work_MLimit_LineMeli.cs:1352` 起）：

| 位 | 属性 | 含义 |
|---:|---|---|
| 0 | 允许查看 | 菜单/列表可见 |
| 1 | 帐户授权 | 可添加/编辑 Meli 账号 |
| 2 | 上传产品 | 可刊登 |
| 3 | 修改明细 | 可修改已上传商品 |
| 4 | 价格调整 | 可调价 |
| 5 | GTIN 编码 | 可使用编码池/GTIN |

构造函数：`Hide = !Member.AllowSite(8)`。

> **【观察】** 个别参数菜单对 UserID `1117/1801/1051` 开放，属于硬编码的内部/测试账号白名单；普通用户不可见。这反映作者在正式版中保留了内部运营入口。

---

## 7. 网络 API 与协议细节

### 7.1 OAuth 流程（时序）

```
用户                 ZYing 客户端                     ZYing 服务端                 MercadoLibre
 |  选择站点/点击登录     |                                |                           |
 |--------------------->| MeliApi.GetAuthUrl(site,state)  |                           |
 |  外部浏览器打开授权页   |                                |                           |
 |=======================================================>|  (Meli 授权页)             |
 |  授权后回调到 vendor 域名，复制 code                   |                           |
 |--------------------->| BtnAuth: meli_acc.auth_token    |                           |
 |                      |------------------------------->|  POST /oauth/token         |
 |                      |                                |-------------------------->|
 |                      |<-- root.token/userinfo/marketsite|<--------------------------|
 |                      | BtnClick: meli_acc.insert/update|                           |
 |                      |------------------------------->|  保存 token/店铺信息        |
 |                      | 运行期 meli_acc.detail/first     |                           |
 |                      |------------------------------->|  返回 token                |
 |                      | MeliApi + Bearer -------------->|-------------------------->|
```

关键点：

- `GetAuthUrl` 拼接 `<authHost>/authorization?response_type=code&client_id=<AppId>&redirect_uri=<escaped>&state=<escaped>`（`MeliApi.resolved.cs:57-75`）。
- `state` 使用 `"client_" + Environment.TickCount`，可预测性较高【观察】。
- `GetToken`/`RefreshToken` 在客户端也有实现，使用内置 `AppSecret`（`MeliApi.resolved.cs:76-121`）；但 `GetToken` 在本次反编译的 131 个 User 类中**未发现调用点**，实际主路径是服务端 `meli_acc.auth_token`【观察/未知】。
- 服务端返回 `token`（JSON：access/refresh/expires）、`userinfo`、`marketsite`；客户端只做反序列化与展示。**【推断】** 授权码换 token 的调用发生在服务端，因为客户端只需提交 `code` + `appid` 即可拿到成品 token；客户端虽保留 `MeliApi.GetToken`，但未发现生产调用点。
- 自动刷新（`CheckToken`）只更新内存字段，不回写服务端；用户在授权详情手动刷新后，仍需点击“确认提交”才会持久化【观察+推断】。这意味着 refresh token 轮换后，服务端可能保留旧 refresh token，存在后续刷新失败的风险【推断】。

### 7.2 REST 端点清单（69 个字符串字面量）

> 以下为反编译结果中的端点字面量（`analysis/analysis-meli-endpoints.txt`），已去掉重复；`https://api.mercadolibre.com` 为 `MeliApi.BuildUrl` 的基址。方括号为用途推断，非官方命名。

```
/authorization?response_type=code&client_id=
https://api.mercadolibre.com
/products/
/items/
/catalog_listing_eligibility
/multiget/catalog_listing_eligibility?ids=
/marketplace/domain_discovery/search?q=
/marketplace/v2/claims/
/affects-reputation
/actions/open-dispute
/expected-resolutions/partial-refund
/expected-resolutions/refund
/global/user-products
/global/user-products/
/attachments/
/download
/marketplace/messages/attachments/
/sites/
/categories/all
/categories/
/attributes
/download?access_token=
/expected-resolutions
/messages
/partial-refund/available-offers
/marketplace/v2/claims/reasons/
/marketplace/items/
/clips
/description
/user-products-families/
/user-products/
/global
/items?ids=
/marketplace/moderations/infractions/
https://api.mercadolibre.com/items/
/marketplace_items
/item/
/performance
/visits/items?ids=
/marketplace/users/cap
/locals
/marketplace/users/
/marketplace/messages/
/users/me
/marketplace/orders/
/invoice
/marketplace/messages/packs/
/pictures/
/items
/categories
/users/
https://api.mercadolibre.com/oauth/token?grant_type=authorization_code&client_id=
/marketplace/messages/unread?
/marketplace/user-products/
/mapping
/shipping_options/free
/price_to_win?siteId=
/items/catalog_listings
https://api.mercadolibre.com/oauth/token?grant_type=refresh_token&client_id=
/marketplace/v2/claims/search?
/search?
/marketplace/products/search?
/actions/send-message
/items/search?
/attachments
/clips/upload
/marketplace/messages/attachments?site_id=
/pictures/items/upload
```

**观察到的调用特征**：

- 单商品详情：`/items/{id}?include_attributes=all`；批量：`/items?ids=...&include_attributes=all`，超过 20 个直接抛异常（`MeliApi.resolved.cs:560-590`）。
- 用户产品（UP）：`/marketplace/user-products/...`、`/global/user-products/...`，部分请求加 `X-API-Version: 2`（`MeliApi.resolved.cs:1072-1085`）。
- 配额：`/marketplace/users/cap`（`MeliApi.resolved.cs:791-799`）。
- 索赔：`/marketplace/v2/claims/search?`、`/marketplace/v2/claims/{id}`、`/actions/open-dispute`、`/expected-resolutions/refund`、`/expected-resolutions/partial-refund`、`/partial-refund/available-offers`。
- 消息：`/marketplace/messages/`、`/marketplace/messages/packs/`、`/marketplace/messages/attachments/`、`/actions/send-message`、`/marketplace/messages/unread?`。
- 图片/视频：`/pictures/items/upload`、`/pictures/`、`/clips`、`/clips/upload`。
- 发票：`/invoice`；绩效：`/performance`；访问量：`/visits/items?ids=`；违规：`/marketplace/moderations/infractions/`。

### 7.3 请求头、批量与重试

| 机制 | 观察 |
|---|---|
| 认证 | `Authorization: Bearer <access_token>`（`MeliApi.resolved.cs:1094`） |
| 版本头 | `X-API-Version: 2`（`MeliApi.resolved.cs:1080`） |
| UA | 固定 Chrome/149 + Edge/149（`HttpClientEx.core.resolved.cs:46/64`） |
| 批量 | `/items?ids=` 最多 20 个 ID（`:560/588`） |
| 图片上传重试 | `retry` 从 0 开始，`retry > 3` 抛错；仅 429/500/502 重试；延迟 `600 * retry` ms（`:433-490`） |
| 用户产品重试 | 初始 `retry=2`，429 时 `Task.Delay(500)` 后递归（`:802-818`） |
| 家庭/变体重试 | 初始 `retry=3`，429 时 500ms（`:846-862`） |
| 超时 | `HttpClientEx.Create(60/120)` 秒 |

### 7.4 ZClient 命令规模

本次对全部反编译 `.cs` 扫描：

- `.Command =` 赋值点：**288 处**；
- 去重后的命令字符串：**184 个**；
  - 其中 **15 个是动态前缀**（如 `meli_acc.`、`meli_claim.`、`meli_message.`、`meli_qa.`、`meli_stat.` 等，运行时拼接 action）；
  - 完整静态命令 **169 个**；
- 美客多相关：**174 个**（含 15 个动态前缀）；通用/其他：**10 个**（`common.login/logout`、`logins.*`、`sale.batch`、`spider.save_cost`、`users.*`）。

> 与早前统计的差异：本次以“所有 `.resolved.cs` + 动态前缀”重新统计，数字以本节为准。

### 7.5 关键命令载荷

| 命令 | 载荷/返回 | 证据 |
|---|---|---|
| `common.login` | 登录名、`login_pwd`（DES）、2FA code；返回 `Member` 记录 | `MLogin.resolved.cs:359-413` |
| `logins.auth` | 登录二次验证 code | `MLogin.resolved.cs:511` |
| `logins.check` / `logins.delete` | 登录绑定查询/删除 | `ZYing.Channel._` / `EGLogin` |
| `meli_acc.auth_token` | `code`、`appid` → `root.token`、`root.userinfo`、`root.marketsite` | `ZYing.Detail.Meli授权详情.resolved.cs:817-861` |
| `meli_acc.insert/update` | 全部 `acc_*` 字段（见 6.2） | `:1475` |
| `meli_acc.detail` / `first` | `id`/首个账号 → Meli 账号 + token | `MeliApiBuilder.resolved.cs:30-48` |
| `meli_acc.get_appinfo` | 服务端下发自定义应用配置 | `MeliAppInfo.resolved.cs:106-118` |
| `meli.submit` | 上传记录 + `accs/grades/extra/stock/size_id/...` | `ZYing.Detail.Meli新版上传Detail.resolved.cs:1081-1235` |
| `meli.sync` | `aid`、`startTime`、`endTime` | `ZYing.Auto.ZSyncMeli.resolved.cs:173-195` |
| `meli.order` | `aid`、`startTime`、`status` | `ZYing.Auto.ZOrderMeli.resolved.cs:167-189` |
| `meli.label` | `oid`、`pick` → base64 PDF | `Logic.MeliLabelPrint.resolved.cs:59-61`；`ZYing.Dialog.打印Meli.resolved.cs:501-561` |
| `meli_message.*` | 会话/消息 ID、内容 | `ZYing.Channel.Meli消息列表.resolved.cs:573/739` |
| `meli_claim.*` | 索赔 ID、退款金额/原因、消息 | `ZYing.Channel.Meli索赔列表.resolved.cs:563/851` |
| `meli_question.*` / `meli_qa.*` | 问题 ID、回复内容、备注 | `ZYing.Channel.Meli站内信.resolved.cs:503/665`；`Meli本土Qa.resolved.cs:289/625` |

> 大多数命令的具体响应字段由 `Record`/`RecordSet` 动态解析，未逐一还原 schema；本报告只标注客户端实际读取的字段。

---

## 8. 授权与许可机制

### 8.1 `Member` 模型

`Member.Parse(Record re)`（`ZYing_Member.work.resolved.cs:693-768`）从服务端记录读取：

| 字段 | 用途 |
|---|---|
| `id`、`token`、`userid`、`name`、`corp`、`people`、`role`、`rolename` | 身份与会话 |
| `sites`、`allowsites` | 可用平台与允许平台；`AllowSite(8)` 判断 Meli |
| `limit` | 各平台权限位（`MLimit`，含 `Meli`） |
| `ptime` | 许可/账号到期时间 |
| `level`、`type`、`auth` | 会员等级/类型/认证状态 |
| `sect`、`config`、`pub`、`lans` | 部门、配置、语言 |
| `access`、`secret`、`bappkey/bsecret/btoken`、`alikey/alisecret`、`youdaokey/youdaosecret`、`fnkey/fnsecret` | 服务端下发的第三方平台密钥（具体值未在本次样本中观察） |
| `depot`、`shipper`、`reward`、`percent`、`profit`、`label_fee`、`exam_*`、`scantype`、`instead`、`batch`、`edit`、`own`、`self`、`negtive`、`bad`、`for` | 仓储/结算/考试/扫描/编辑等业务开关 |

- `SupportSites()` 返回支持的站点列表；Meli 为 `8`（`ZYing_Member.work.resolved.cs:775` 起）。
- `AllowSite`/`ContainSite` 是菜单和业务逻辑的主要闸门（`:646-690`）。
- `Logout` 使用 `common.logout`，并向各 IPGroup 广播（`:794` 起）。

### 8.2 许可判断

- **服务端驱动**：登录成功才返回 `Member`；`ptime` 到期、`level` 不足、`sites` 不含 8、`limit` 位为 false 都会导致功能不可用。
- **本地无签名**：未发现对 `Member` 记录的本地签名/证书校验；但本地伪造 `Member` 只能改变 UI 显示，服务端命令仍会校验 `Member.Token` 与账号权限【推断】。
- **登录绑定**：`logins.check` / `logins.delete` 用于查询/删除登录绑定；2FA 通过 `logins.auth` 或本地数字码完成。
- **风险**：由于核心操作由服务端执行，客户端逆向无法直接“离线破解”账号配额；但可提取已授权 token 直连 Meli（需服务端返回 token），这是架构层面的暴露面【推断】。

### 8.3 内置 Meli 应用配置

`MeliAppInfo` 内置 6 组配置（`MeliAppInfo.resolved.cs:46-101`）：

| 属性 | 用途（推断） |
|---|---|
| `ZYing` | 默认跨境应用 |
| `PuHui` | 普汇/合作方应用 |
| `JuYing` | 聚鹰/合作方应用 |
| `Cohere` | Cohere 应用（特定会员条件下用于 CBT） |
| `网店ZYing`（混淆名） | 本土/网店应用 |
| `网店`（混淆名，历史） | 旧应用，保留 |

- 每组含 `AppId`、`AppSecret`、`RedirectUrl`，均以字符串混淆存放；报告不展示具体值。
- `Load()` 通过 `meli_acc.get_appinfo` 允许服务端下发**自定义应用**（`SetSelfApp` 读取 `appid/appsecret/appname/redicrturl`，`:151-160`），说明运营方可以按租户切换 OAuth 应用。
- `GetAppInfo(bool isCbt)`：CBT 优先自定义应用，否则按会员条件选 `Cohere` 或 `ZYing`；本土固定使用“网店ZYing”【观察】。

**影响**：AppSecret 存在客户端意味着任何拿到样本的人都能提取；若这些应用仍有效，可能被用于伪造授权请求、刷新 token 或消耗应用配额。报告不提供利用步骤。

---

## 9. 风控与反爬策略

### 9.1 机制清单

| # | 机制 | 证据 | 性质 |
|---:|---|---|---|
| 1 | Token 提前 15 分钟刷新 | `MeliApi.CheckToken`（`:1123`） | 稳定性 |
| 2 | 批量查询上限 20 | `GetItems`（`:560/588`） | 配额/稳定性 |
| 3 | 429/500/502 重试 + 递增延迟 | `UploadPic`（600ms×retry）；UP/变体 429→500ms | 限流适配 |
| 4 | 仅对白名单错误重试 | `MeliErrorParser.CanRetry`（`:34-60`） | 避免无效重试 |
| 5 | 本地类目/属性缓存 | `ATTRMELI-*` + `EPropHelper` | 降低 API 调用 |
| 6 | 刊登/上架配额字段 | `acc_quota`、`acc_totalitems`、`acc_listingtype`、`/marketplace/users/cap` | 配额管理 |
| 7 | 每日上传/刊登限制识别 | `MeliErrorParser` 的 `daily_reached`、listing limit 翻译 | 平台限制 |
| 8 | KYC/账号限制识别 | `User is unable to list`、`restrictions_coliving`、`phone_pending` 等翻译 | 合规/风控 |
| 9 | Meli 错误中文化 | `MeliErrorParser.Translates`（337 行） | 运营效率 |
| 10 | 固定浏览器 UA | `HttpClientEx` Chrome/149 + Edge/149 | 基础伪装 |
| 11 | 关闭证书校验 | `ServerCertificateCustomValidationCallback => true` | **安全弱点**，不是反爬 |
| 12 | 无代理/Meli | MeliApi/HttpClientEx 未发现 Proxy 配置 | 未做 IP 轮换 |
| 13 | 132 个浏览器 UA | `HttpAgent` | 可能用于其他平台采集；Meli 路径未使用 |
| 14 | 任务队列/节流 | 服务端 `meli.sync`/`meli.order` 分页、`ordersize/salesize` | 批量处理 |

### 9.2 评估

- 这些措施主要是**保证大批量刊登/订单同步时的稳定性、避免触发平台限流、降低 API 调用量**，并非成熟的“反检测”体系。
- 未观察到：代理池/IP 轮换、请求指纹伪造（TLS/HTTP2）、验证码自动处理、行为模拟、随机化请求间隔（除固定 500/600ms 重试外）。
- `HttpAgent` 的 132 个 UA 与 `AmazonService_ZCaptcha` 表明产品在其他平台可能有采集/验证码相关能力；本次未发现其用于 Meli 主路径。
- **合规声明**：以上仅用于防御性安全评估与授权测试记录；不提供绕过平台风控、批量注册、虚假交易或规避封禁的方法。

---

## 10. 安全发现

### 10.1 风险总表

| 级别 | 发现 | 证据 | 影响 |
|---|---|---|---|
| **严重** | OAuth access/refresh token 经裸 TCP 传输 | `ZClient.GetPost`（`:874`）；`meli_acc.insert/update`；`acc_access/acc_fresh` 明文 | 同一网络路径上的被动监听者可获取 token，冒充店铺调用 Meli API |
| **严重** | 6 组 Meli AppSecret 硬编码在客户端 | `MeliAppInfo.resolved.cs:46-101`；`MeliApi.GetToken/RefreshToken` 使用 AppSecret | 样本泄露即可提取应用密钥；可能被用于伪造授权/刷新 token、消耗配额 |
| **高** | 登录/VPS 密码使用固定 DES 密钥/IV | `Safe.core.resolved.cs:32-34/506-543`（key=`19750630`） | 可逆混淆，任何拿到样本的人都能解密 |
| **高** | Meli HTTPS 证书校验被关闭 | `HttpClientEx.core.resolved.cs:70-71` | 中间人可截获 Bearer token 与业务数据 |
| **高** | `Server.ini` 明文保存登录名；`Password.ini` 明文密码回退 | `MLogin.resolved.cs:203/359-367` | 本地取证/共享机器泄露账号 |
| **中** | 自动刷新后的 token 不回写服务端 | `CheckToken`（`:1123-1143`）只改内存；`BtnFresh` 需手动保存 | refresh token 轮换后服务端旧 token 失效，可能导致刷新失败 |
| **中** | CefSharp Cookie SQLite 明文保存 1688 会话 | `cache/temp/Cookies` | 本地会话被其他进程/取证工具读取 |
| **中** | 固定 UA + TLS 1.0/1.1/1.2 | `HttpClientEx` | 指纹稳定，易被平台关联；旧 TLS 协议安全性差 |
| **低** | `state` 使用 TickCount | `MeliApi.GetAuthUrl` 调用点 | CSRF 防护弱（但授权码仍需用户粘贴，实际风险有限） |
| **低** | 字符串混淆可逆 | 5.1 节 | 提高逆向成本有限，不构成保密 |

### 10.2 重点说明

1. **Token 明文传输是架构级问题**：即使 Meli 侧是 HTTPS，token 在“客户端 → ZYing 网关”这一段仍是明文。若攻击者能访问该网络路径（公共 Wi-Fi、VPS 宿主、旁路镜像），即可同时获得 `Member.Token` 与 Meli token。代码中的 XOR 函数是死代码，不能提供保护。
2. **AppSecret 客户端化是供应链风险**：6 组应用凭据共享给所有客户端；一旦某一版本泄露，影响的是应用级配额与信任，而非单个店铺。报告未验证这些应用当前是否仍有效。
3. **DES 不是加密而是编码**：固定 key/IV + DES-ECB/CBC（实现使用 `DESCryptoServiceProvider` + 固定 IV），没有随机盐/KDF；`Safe.Encrypt` 的结果可直接用同一程序集解密。
4. **TLS 校验关闭**：`ServerCertificateCustomValidationCallback => true` 使 HTTPS 的中间人防护失效；结合固定 UA，攻击者可在受控网络中进行透明代理。
5. **自动刷新不回写**：`CheckToken` 刷新后只更新 `_token/_refresh/_expire`；`RefreshToken` 不调用任何 `meli_acc` 命令。用户手动刷新后也必须点“确认提交”才写回。此设计可能导致服务端保存的 refresh token 与 Meli 最新值不一致。
6. **权限绕过面**：本地 `Member`/菜单可被 patch，但服务端命令仍校验 `Member.Token` 与账号权限；真正的高价值目标是 token 与 AppSecret，而非 UI 开关【推断】。

---

## 11. 未知项与后续验证

| # | 未知项 | 说明 | 建议（仅授权场景） |
|---:|---|---|---|
| 1 | 服务端实现与 token 存储方式 | 未获取服务端代码/数据库；无法确认 token 是否加密存储、是否有访问审计 | 若获授权，做服务端配置与日志审查 |
| 2 | 每个命令的响应 schema | 客户端按 `Record` 动态读取，未逐字段还原 | 用测试账号做受控抓包（需另行授权） |
| 3 | `MeliApi.GetToken` 的实际调用方 | 在 131 个 User 类中未发现调用；可能在未反编译类型/历史代码中 | 扩大反编译范围或运行时断点 |
| 4 | `GetRecord()` 基类 | 上传表单的公共字段来自继承基类，未在本次类型集合中 | 反编译基类/父类型 |
| 5 | 安装器行为 | `ZInstall.exe` 仅做 PE/元数据识别，未展开安装脚本/释放文件 | 在隔离环境做安装监控（需授权） |
| 6 | MLU 支持状态 | 枚举与 UI 提示存在冲突 | 查服务端站点配置或产品文档 |
| 7 | 自动刷新后的 token 生命周期 | 静态代码显示不回写；实际服务端是否接受旧 refresh token 未验证 | 受控环境测试（需授权） |
| 8 | `HttpAgent` 132 UA 的用途 | 未在 Meli 路径发现调用；可能用于 1688/其他平台采集 | 追踪调用链 |
| 9 | 日志是否记录 token | `TextLogger` 存在于上传失败等路径；未确认是否记录请求头/token | 检查日志文件样本 |
| 10 | 本地是否还有其他隐藏数据 | 未发现业务 DB；可能有加密日志/临时文件未覆盖 | 全目录熵扫描/文件类型普查 |

---

## 附录 A：Meli 端点字面量（69）

见 7.2 节代码块；原始文件：`analysis/analysis-meli-endpoints.txt`。

## 附录 B：ZClient 命令目录（184 个字符串，含 15 个动态前缀）

> 说明：带 `.` 结尾的条目为动态前缀（运行时拼接 action）；其余为完整静态命令。分组按首段前缀。

```
common: common.login, common.logout
logins: logins.auth, logins.check, logins.delete

meli: meli.batch, meli.invoice, meli.label, meli.localsubmit, meli.localupdate,
      meli.order, meli.order_split, meli.order_split_skus, meli.submit, meli.sync,
      meli.syncAccProduct, meli.update

meli_acc: meli_acc., meli_acc.appinfo, meli_acc.auth_token, meli_acc.del_appinfo,
          meli_acc.delete, meli_acc.detail, meli_acc.first, meli_acc.flag,
          meli_acc.get_appinfo, meli_acc.stat

meli_catalog_monitor: meli_catalog_monitor., meli_catalog_monitor.add_monitors,
          meli_catalog_monitor.adjust_price, meli_catalog_monitor.async,
          meli_catalog_monitor.delete_item, meli_catalog_monitor.detail,
          meli_catalog_monitor.pause_active, meli_catalog_monitor.remove

meli_category: meli_category.attr_values, meli_category.detail, meli_category.refer_cbt,
          meli_category.refer_cbts, meli_category.refer_total, meli_category.select,
          meli_category.update, meli_category.update_attrs, meli_category.updateattr

meli_catemap: meli_catemap.insert

meli_claim: meli_claim., meli_claim.delete, meli_claim.detail, meli_claim.fetch,
          meli_claim.fetch_claim, meli_claim.get_partial_refund_offers,
          meli_claim.open_dispute, meli_claim.partial_refund, meli_claim.refund,
          meli_claim.save_trans_message, meli_claim.send_message, meli_claim.set_remark

meli_clips: meli_clips.delete, meli_clips.fetch, meli_clips.load, meli_clips.save,
          meli_clips.select

meli_code: meli_code., meli_code.clear, meli_code.delete, meli_code.insert

meli_extra: meli_extra., meli_extra.delete, meli_extra.detail, meli_extra.insert,
          meli_extra.update

meli_freight: meli_freight.insert, meli_freight.select, meli_freight.ship
meli_item_best: meli_item_best.delete
meli_kind: meli_kind.pop, meli_kind.select

meli_locallog: meli_locallog., meli_locallog.delete, meli_locallog.detail

meli_localstat: meli_localstat., meli_localstat.active, meli_localstat.attr,
          meli_localstat.batchDelete, meli_localstat.batchstatus, meli_localstat.clean,
          meli_localstat.delete, meli_localstat.detail, meli_localstat.predict,
          meli_localstat.setstat, meli_localstat.title

meli_log: meli_log., meli_log.delete, meli_log.delete_v2, meli_log.detail,
          meli_log.detail_v2, meli_log.exe

meli_log_up: meli_log_up., meli_log_up.delete, meli_log_up.detail

meli_message: meli_message., meli_message.detail, meli_message.fetch,
          meli_message.save_trans_message, meli_message.send_message

meli_metrics: meli_metrics., meli_metrics.detail, meli_metrics.fetch

meli_netproceeds: meli_netproceeds.select, meli_netproceeds.set_netproceeds,
          meli_netproceeds.set_zero, meli_netproceeds.update

meli_picture: meli_picture.save_pics, meli_picture.select

meli_qa: meli_qa., meli_qa.delete, meli_qa.detail, meli_qa.memo, meli_qa.pull,
          meli_qa.reply, meli_qa.save

meli_qatemp: meli_qatemp.delete, meli_qatemp.detail, meli_qatemp.insert,
          meli_qatemp.select, meli_qatemp.stat

meli_question: meli_question., meli_question.delete, meli_question.detail,
          meli_question.memo, meli_question.noreply_count, meli_question.pull,
          meli_question.reply

meli_rate: meli_rate.select

meli_refer: meli_refer.batch, meli_refer.delete, meli_refer.detail, meli_refer.offer,
          meli_refer.select, meli_refer.update

meli_size: meli_size.update
meli_sizechart: meli_sizechart.delete, meli_sizechart.fetch, meli_sizechart.select

meli_stat: meli_stat., meli_stat.active, meli_stat.adjust_price, meli_stat.attr,
          meli_stat.batchDelete, meli_stat.change_cate, meli_stat.changesale,
          meli_stat.clean, meli_stat.delete, meli_stat.detail, meli_stat.export,
          meli_stat.get_infraction, meli_stat.intro, meli_stat.predict,
          meli_stat.publish, meli_stat.setstat, meli_stat.title, meli_stat.updatesale

meli_stat_up: meli_stat_up., meli_stat_up.active, meli_stat_up.adjust_price,
          meli_stat_up.change_cate, meli_stat_up.delete, meli_stat_up.detail,
          meli_stat_up.get_infraction, meli_stat_up.intro, meli_stat_up.publish,
          meli_stat_up.setstat, meli_stat_up.title, meli_stat_up.update_attrs

meli_sync: meli_sync.sync_items
meli_up: meli_up.submit, meli_up.update
meli_update: meli_update.batch_ids
meliup_sync: meliup_sync.sync_items

sale: sale.batch
spider: spider.save_cost
users: users.export, users.loadUPC, users.meli
```

## 附录 C：`.dat` 容器与条目

- 格式：见 4.1；解析器：`analysis/scripts/dat_dump.py`；输出：`analysis/dat/`。
- `ZOther.dat`（v97，29 项）：Meli 类目/属性缓存（`ATTRMELI-CBT/MCO/MLA/MLB/MLC/MLM/MLU`），以及 `CBT/MCO/MLA/MLB/MLC/MLM/MLU` 等站点数据。
- `remotes.dat`（v3，1 项）：`COMMONREMOTEAREAS`，约 1.99 MB，邮编/偏远地区数据。
- `ZData.dat`、`ZAmazon.dat`、`ZEbay.dat`、`ZForm.dat`、`ZCaptcha.dat`：其他平台/表单/验证码资源（未逐项展开）。
- 这些资源均为**本地缓存/静态数据**，未发现账号 token 或私钥。

## 附录 D：关键源码行号索引

| 主题 | 文件:行 | 说明 |
|---|---|---|
| ZClient 默认端口 | `analysis/decrypted/ZYing_ZClient.resolved.cs:66/77` | `Port = 7735` |
| ZClient 网关节点 | 同上 `:1241-1267` | 主/备/香港/美客多 2 等 |
| ZClient 请求帧 | 同上 `:874` | `GetPost` |
| ZClient 响应校验 | 同上 `:943` | UUID/token 不匹配抛 `NoLoginEx` |
| ZClient 死 XOR | 同上 `:904-925` | `Encrypt`/`Decrypt` 未调用 |
| Meli 授权 URL | `analysis/decrypted/MeliApi.resolved.cs:57` | `GetAuthUrl` |
| Meli 换 token | 同上 `:76` | `GetToken`（客户端实现） |
| Meli 刷新 token | 同上 `:99` | `RefreshToken` |
| Meli 请求头/认证 | 同上 `:1088-1096` | `SendCoreAsync` + Bearer |
| Meli 版本头 | 同上 `:1072-1085` | `X-API-Version: 2` |
| Meli 自动刷新 | 同上 `:1123-1143` | 15 分钟提前刷新 |
| Meli 批量上限 | 同上 `:560/588` | >20 抛异常 |
| Meli 图片重试 | 同上 `:433-490` | 429/500/502，600ms×retry |
| Meli UP 重试 | 同上 `:802-862` | 429→500ms |
| Meli 授权站点 | 同上 `:1145-1160` | CBT/MLM/MCO/MLB/MLC/MLA |
| 会员解析 | `analysis/decrypted/ZYing_Member.work.resolved.cs:693-768` | `Parse` |
| 会员站点闸门 | 同上 `:646-690` | `AllowSite`/`ContainSite` |
| 会员登出 | 同上 `:794` | `common.logout` |
| Meli 权限位 | `analysis/decrypted/Work_MLimit_LineMeli.cs:1352` | 0-5 位 |
| 固定 DES | `analysis/decrypted/Safe.core.resolved.cs:32-34/506-543` | key/IV + `Encrypt/Decrypt` |
| HTTP UA/TLS | `analysis/decrypted/HttpClientEx.core.resolved.cs:46/64/70-71` | 固定 UA、TLS1.0-1.2、跳过证书校验 |
| 登录/Server.ini | `analysis/decrypted/MLogin.resolved.cs:203/359-413` | 明文登录名、密码 DES、`common.login` |
| 2FA | 同上 `:511` | `logins.auth` |
| 应用配置 | `analysis/decrypted/MeliAppInfo.resolved.cs:46-101/106-135/151` | 6 组应用、服务端下发、选择逻辑 |
| API Builder | `analysis/decrypted/MeliApiBuilder.resolved.cs:14-48` | `meli_acc.detail/first` |
| 错误重试白名单 | `analysis/decrypted/ZYing_Meli_MeliErrorParser.resolved.cs:34-60` | `CanRetry` |
| 属性缓存 | `analysis/decrypted/EPropHelper.work.resolved.cs:515/341/421` | `attrMeli-*` |
| 授权流程 | `analysis/decrypted/user/ZYing.Detail.Meli授权详情.resolved.cs:699/729/779/817/955/1269/1475` | 浏览器授权、`meli_acc.auth_token`、刷新、保存 |
| 产品同步 | `analysis/decrypted/user/ZYing.Auto.ZSyncMeli.resolved.cs:171-195` | `meli.sync` |
| 订单接收 | `analysis/decrypted/user/ZYing.Auto.ZOrderMeli.resolved.cs:165-189` | `meli.order` |
| 面单 | `analysis/decrypted/user/Logic.MeliLabelPrint.resolved.cs:57-61`；`ZYing.Dialog.打印Meli.resolved.cs:501-695` | `meli.label`、base64 PDF、合并 |
| 消息/索赔/QA | `analysis/decrypted/user/ZYing.Channel.Meli消息列表.resolved.cs:573/739` 等 | 动态命令 |
| 菜单 | `analysis/decrypted/ZYing_Menus.resolved.cs:1739-1841/3463` | Meli 菜单与权限 |

---

### 附：分析产物位置

- 反编译源码：`analysis/decompiled/`（按类型平铺）
- 字符串还原源码：`analysis/decrypted/`（工作程序集）与 `analysis/decrypted/user/`（用户程序集）
- 资源解析：`analysis/dat/`
- 端点清单：`analysis/analysis-meli-endpoints.txt`
- 命令分组：`analysis/command_groups.txt`
- 类型/资源/证据清单：`analysis/work-types.txt`、`analysis/user-types2.txt`、`analysis/resources.txt`、`analysis/evidence_ids.txt`、`analysis/evidence_map.txt`
- 解析脚本：`analysis/scripts/dat_dump.py`、`resolve_strings.py`、`resolve_strings_user.py`、`map_evidence.py`

> 本报告仅用于授权范围内的安全评估与逆向研究。报告中的服务器地址、端口、文件路径用于证据复核；请勿用于未授权的访问、账号接管或平台风控规避。
