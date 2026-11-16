# 08 ILSpy 反编译结果与混淆边界（第三轮）

> 触发：用户安装 ILSpy 11.2（`D:\AI工具\ILSpy_windows_11.2.0.10050-x64`）后，继续做 IL 级分析。
> 本轮仍为**静态分析**：未运行安装程序、未运行 ERP 客户端、未抓包、未脱壳、未修改任何二进制。**RUNTIME_VERIFIED = 0 条**。
> 结论中所有 `OBSERVED` 均指"对交付物的静态观察"，不等于运行时行为。

---

## 0. 摘要（先说结论，含对 `07` 的修正）

ILSpy 确实把 `ZClient` 的完整 API 与 `ZResult` 数据契约摊开了，这部分是实质性进展。但**它没有像我上一轮估计的那样"一次关掉三个 UNKNOWN"**：这个客户端用了完整的 .NET 混淆方案，**字符串常量被加密存储**，因此命令名、URL、字段名等字面量在静态反编译里**看不到明文**。

| 上一轮（`07`）的判断 | 本轮实测 | 修正后状态 |
|---|---|---|
| ILSpy 可补 U2 的 RPC 命令名 | `采集列表` 全类仅 1 个明文字面量（`"SList"`），175 处解密调用 | **U2 命令名静态不可得**；仅结构与 API 可得 |
| ILSpy 可补 U8 枚举值 | `Enumer.Sale.Stat/Level/Grade/Flag` 不是枚举，是 `EItemCollection` | **U8 变更**：架构上无本地枚举，取值由服务端下发 |
| ILSpy 可补 U9 的 DTO | `ZResult{Command, IsErr, Message, RS}` 已确认 | **U9 部分关闭**（响应封装级，业务 DTO 仍 UNKNOWN） |

---

## 1. 工具与方法

| 项 | 内容 |
|---|---|
| 用户已装 | `D:\AI工具\ILSpy_windows_11.2.0.10050-x64\ILSpy.exe`（Avalonia GUI，**不含命令行**） |
| 本轮所用 | `ilspycmd.win-x64` **11.2.0.10050**，取自 NuGet `ilspycmd.win-x64` 包，解压到临时目录；`ilspycmd.exe` 依赖 .NET 8 运行时（本机已有 8.0.31） |
| 调用方式 | Python `subprocess` 传参（避免 Git Bash 中文参数乱码），stdout 按 **GBK** 解码后转 UTF-8 落盘 |
| 范围控制 | 只 `-t <类型>` 定向反编译指定类型，未用 `-p` 导出整个项目（17.7 MB 程序集体量不可控） |
| 产物目录 | `C:\Users\ASUS\AppData\Local\Temp\zying-rea-20261115\decomp\` |

> 踩坑记录：`ilspycmd` 输出到 stdout 时使用系统 ANSI 代码页（GBK），按 UTF-8 解码会把 `采集列表` 变成 `�ɼ��б�`。脚本化时必须按 GBK 解码。

---

## 2. 反编译成功清单

| 类型 | 所在程序集 | 行数 | 主要产出 |
|---|---|---:|---|
| `ZYing.Channel.采集列表` | `ZYing.User.dll` | 2,273 | 采集库主列表、筛选维度、`ZClient` 异步状态机 |
| `ZYing.ZClient` | `ZYing.Work.dll` | 2,435 | RPC 客户端完整 API、连接/加解密/协议方法 |
| `ZYing.ZResult` | `ZYing.Work.dll` | 105 | RPC 响应契约 |
| `Enumer.Sale` | `ZYing.Work.dll` | 735 | 采集/商品相关选项集合清单 + 2 个硬编码枚举 |

失败项：`ZYing.UI.CBase`、`Enumer.EItemCollection`（`rc 65`，类型全名不在该命名空间，未继续枚举查找——收益低）。

---

## 3. 拿到的关键结构（OBSERVED）

### 3.1 `ZYing.ZClient`：RPC 客户端全貌

```csharp
public static ZClient Create(IPGroup remote = null);
public Task<ZResult> PostAsync(Para p);
public async Task<ZResult> PostAsync(NameValueCollection nv);
public Task<ZResult> PostAsync(params Record[] arr);
public async Task<ZResult> PostAsync(RecordSet rs);
public async Task<ZResult> PostAsync(string name, bool/int/long/decimal/string value);
public async Task<ZResult> PostAsync(byte[] data);
public Task<ZResult> PostAsync();

public Record this[string name] => _ret.RS[name] ?? new Record();   // 按名字从结果集取行

private unsafe async Task<ZResult> ConnectAsync(byte[] data, int trycount = 0);
private async Task<byte[]> Reading(int readed);
private void SetSerial();
private void DoHeader();  private void DoParse(byte[] back);  private void DoFinish();
private static byte[] Encrypt(byte[] arr);  private static byte[] Decrypt(byte[] arr);
private static void localhost();
public static void CheckExpired();  public static void Expired(int timeout);
public static void CloseAll();      public void Close();  public void Reset();
```

**可确认的事实**：
- 客户端与云端是**长连接 + 自定义帧**（`ConnectAsync` / `Reading(readed)` / `DoHeader` / `DoParse`），不是纯 REST。
- 请求与响应都是 `NameValueCollection` / `Record` / `RecordSet` 这类**弱类型键值容器**，没有强类型 DTO。
- `this[string name]` 说明 UI 拿到 `ZResult` 后按**字符串 key 取 `Record`** 再绑表格 —— 这是典型的"结果集驱动 UI"。
- `localhost()` 存在，与 `07` 里 `127.0.0.1:8888` 的本地监听线索一致（用途仍 UNKNOWN）。
- `CheckExpired` / `Expired` / 连接池（`IPGroup`、`GetPoolClient`、`Enqueue`）说明有连接复用与超时管理。

> 合规说明：`Encrypt` / `Decrypt` / `SetSerial` 只记录**存在与职责**，不还原算法、不提取密钥、不尝试复用。协议细节不作为本文档交付内容。

### 3.2 `ZYing.ZResult`：RPC 响应契约（U9 部分关闭）

```csharp
public class ZResult
{
    public string Command;   // 回显命令名（值在运行时才可见）
    public bool   IsErr;
    public string Message;
    public RecordSet RS;     // 结果集
    public bool   Done();
}
```

配合 `ZClient.this[string name]`，客户端取数模型是：

```text
PostAsync(...) -> ZResult{ Command, IsErr, Message, RS: RecordSet }
                -> zc["某 key"] -> Record -> 绑定表格/明细
```

**这对 Una Seller 的意义**：智赢用 `Record/RecordSet` 这种通用行集承载所有业务响应，好处是列表/明细/导出共用一套容器，坏处是**没有编译期契约、字段全靠字符串约定**。自研应改为强类型 DTO + 明确的分页元数据，不要照搬弱类型容器。

### 3.3 `Enumer.Sale`：选项来自服务端，不是本地枚举（U8 变更）

`Enumer.Sale` 是一个 `class`，其中 `Stat` / `Level` / `Grade` / `Flag` 等**全部是 `EItemCollection` 类型**（运行时集合），不是 `enum`：

```csharp
public class Sale
{
    public enum Need { 毛重 = 1 }
    public enum 特殊 { 无 = 0, 带电池 = 1, 锂电 = 10, 液膏体 = 100,
                       磁性 = 1000, 易碎 = 10000, 大件 = 100000 }

    public static readonly EItemCollection Terms, Areas, Special, Stat, Level, Grade, Flag,
        Host, HostAll, Department, Material, Gem, Metal, Age, SWidth,
        TitleKeyTypes, TitleKeyInitTypes;
}
```

**结论修正**：采集库的「状态 / 等级 / 评分 / 标记 / 站点 / 类目词」不是客户端写死的枚举，而是**服务端下发的可配置选项集**（`EItemCollection` / `EItem`）。这也解释了为什么上一轮在元数据里只能看到 `doStat/doLevel/doGrade/doFlag` 这类筛选方法，却找不到枚举定义。

> 对 Una Seller 的设计参考：采集状态这类**会随业务调整的维度**，用「服务端可选值表 + 客户端动态渲染」比硬编码枚举更耐改；但必须有稳定 code（不能只存中文 label），否则历史数据会随文案变更而失配。

### 3.4 两个硬编码枚举：商品物流/特殊属性

`Need { 毛重 = 1 }` 与 `特殊 { 无=0, 带电池=1, 锂电=10, 液膏体=100, 磁性=1000, 易碎=10000, 大件=100000 }` 是**编译期固定**的。取值是十进制位组合，支持多选叠加（如 带电池+磁性 = 1001），用于刊登时的物流/合规属性。

---

## 4. 混淆边界：静态分析到此为止（OBSERVED）

以 `ZYing.Channel.采集列表`（2,273 行）为样本统计：

| 指标 | 数值 | 含义 |
|---|---:|---|
| 明文字符串字面量 | **1**（`"SList"`） | 业务字符串几乎全部加密 |
| 解密辅助调用 `DQmUrYZ9D(...)` | 175 处 | 每次取字符串都走运行时解密 |
| `<Module>{9339e5e7-...}.m_xxx` 静态字段引用 | 228 处 | 密钥/密文分片挂在模块静态字段 |
| `switch (num)` 状态机 | 2 处 | 选择性控制流平坦化（主要在静态构造器） |
| 全程序集随机方法名占比 | 29.9%（18,389 / 61,468） | 方法级重命名 |

典型形态：

```csharp
new EItem(A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1FEC196 ^ <Module>{9339e5e7-...}.m_9ba69d50...),
          A8RedEce1ZSTaHmUDS.DQmUrYZ9D(0x1B57BFED ^ <Module>{9339e5e7-...}.m_fa144e49...))
```

**这意味着**：
- 命令名（`spider.*` 之类）、接口 URL、字段名、中文提示文案，**静态反编译拿不到明文**。
- 继续往下只有三条路：①运行时内存取值；②模拟解密函数；③脱壳。其中 ②③ 属于绕过技术保护措施，**本轮明确不做，后续也不建议做**。
- ①（运行时 dump）是唯一既干净又合规的路径，见 §6。

> 顺带修正 `07` §4.1 的措辞：当时把 29.9% 随机名描述为"选择性混淆"，现在可以确认是**完整混淆方案**（重命名 + 字符串加密 + 控制流平坦化），只是方法/类型名中的中文业务标识被保留。

---

## 5. UNKNOWN 状态再修正（覆盖 `07` §5 的对应行）

| ID | `07` 的状态 | `08` 修正后 | 依据 |
|---|---|---|---|
| U2 | 部分 OBSERVED（结构）+ 命令名 UNKNOWN | **结构 OBSERVED（`ZClient` 全 API + `ZResult`）；命令名确认不可静态获得** | §3.1 §3.2 §4 |
| U8 | 维度名 OBSERVED / 枚举值 UNKNOWN | **架构结论 OBSERVED：无本地枚举，取值由服务端 `EItemCollection` 下发**；具体取值仍需运行时 | §3.3 |
| U9 | `Record` 容器 OBSERVED / DTO UNKNOWN | **响应封装 OBSERVED（`ZResult`）；业务 DTO 仍 UNKNOWN** | §3.2 |
| U7 | 方法名级 OBSERVED | 不变。IL 调用边仍 UNKNOWN（方法体可读，但调用目标名被加密处无法还原为命令名） | §4 |
| U1 U3 U4 U5 U6 U10–U14 | — | 不变 | `06` §5 |

---

## 6. 下一步：按性价比排序（都不需要脱壳）

| 优先级 | 动作 | 能补什么 | 风险 |
|---|---|---|---|
| **P0** | `dotnet-dump collect` + `dumpheap -type ZYing.ZResult` + `dumpobj` | **`Command` 字段的明文值**（内存里已解密），一次拿到真实命令名 | 需运行客户端（授权隔离环境）；不 hook、不 patch |
| **P1** | Process Monitor 过滤 `ZYing.exe` | U4 本地缓存真实路径、U5 `127.0.0.1:8888` 是否有监听、U6 CEF 落地位置 | 只读监控 |
| **P1** | mitmproxy / Fiddler + 自有测试账号 | U1 采集入库的 HTTPS 明文（仅 REST 部分） | 需授权环境；自定义 TCP 帧仍是密文 |
| **P2** | PerfView 采集 JIT/调用栈 | U7 运行时方法流（哪个按钮真的走了哪条链） | 只读 |
| **禁止** | 模拟 `DQmUrYZ9D` 解密、脱壳、Frida/Detours hook | —— | 属于绕过技术保护措施，不做 |

**我的判断**：`dotnet-dump` 那一步性价比最高——内存里的 `ZResult.Command` 就是明文命令名，一个 `dumpheap` 就能把 U2 从"结构已知、名字未知"推到"名字已知"，而且完全不碰保护机制。但它要求真的把客户端跑起来，这一步需要你确认环境和授权。

---

## 7. 合规复核

| 要求 | 本轮执行 |
|---|---|
| 原始 EXE / DLL 只读 | ✅ 只反编译临时目录里的副本；未修改、未回写、未重新打包 |
| 不脱壳、不绕过技术保护措施 | ✅ 识别到字符串加密后**主动停止**，未尝试还原解密逻辑 |
| 不破解许可证 / 不绕过登录 | ✅ 未涉及；`CheckExpired` / `SetSerial` 只记存在与职责 |
| 不输出密钥 / Token / 签名算法 / 协议实现 | ✅ `Encrypt`/`Decrypt` 仅列方法签名，无算法与密钥内容 |
| 不修改第三方商业软件 | ✅ 无写入 |
| 不向 GitHub 上传第三方产物 | ✅ 本轮无上传 |
| 本轮仅研究，不改业务代码 / 任务书 / 不建 PR | ✅ 仅新增报告 |

---

## 8. 本轮产出文件（临时目录 `decomp\`）

| 文件 | 内容 |
|---|---|
| `chan_caijiliebiao.cs` | `ZYing.Channel.采集列表` 反编译（2,273 行，GBK→UTF-8） |
| `work_zclient.cs` | `ZYing.ZClient`（2,435 行） |
| `work_zresult.cs` | `ZYing.ZResult`（105 行） |
| `enum_sale.cs` | `Enumer.Sale`（735 行） |

工具链：`ilspycmd.win-x64 11.2.0.10050`（NuGet RID 包解压，非安装），调用脚本走 Python subprocess。
