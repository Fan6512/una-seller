# 智赢（ZYing）保护壳与构建指纹识别报告

取证时间：2026-10-26 · 工具：Detect It Easy 3.21（`diec.exe`，命令行）· 方式：纯静态，未运行目标程序
原始机器可读结果：`analysis/die_<模块名>.json`，人工汇总：`analysis/die_summary.txt`
说明：DIE 的判定属于特征匹配（`type` 为 `linker` / `compiler` / `library` / `protector` / `debug data` 等），带 `(Heur)` 前缀的是启发式猜测，可信度低于确定性条目。

## 1. 核心结论

| 结论 | 强度 | 影响 |
|---|---|---|
| `ZYing.User.dll` 与 `ZYing.Work.dll` 被 **.NET Reactor 6.x** 处理，启用了 Control Flow 与 Anti-ILDASM | 确定性（protector 条目） | 之前"仅自定义字符串混淆、无商用保护壳"的判断是**错的**，已在正文更正 |
| 这两个模块还伴随通用加密段与反分析启发式命中 | 启发式 | 直接 IL 反编译会失败或产出不可读控制流 |
| `ZYing.Core.dll`、`ZYing.UI.dll`、`ZDto.dll`、`ZPdf.dll` **没有** protector 命中 | 确定性（无相关条目） | 优先从这几个模块取证，性价比最高 |
| 全部模块仍保留 **codeview 调试记录**（PDB 路径） | 确定性 | 原始类型名/命名空间未剥净，符号可用于结构还原 |
| 主程序 `ZYing.exe` 是 PE64 原生 apphost（.NET 宿主加载器），不是托管 IL | 确定性 | 业务逻辑不在 exe 里，逆向重心应放在 dll |

## 2. 逐模块识别

### `ZYing.exe`（PE64，425 KB）
原生可执行文件，链接器与编译器特征指向 MSVC 14.29 / Visual Studio 2019 16.x，并命中 .NET Core 宿主加载器特征。调试数据里除了 codeview 还有 `vc_feature` 与 `pogo`（增量链接/PGO 残留）。这就是 .NET 单文件发布产生的应用宿主，真正的托管代码在同目录 dll 里。

### `ZYing.Core.dll`（PE32，932 KB）
无壳。特征命中 Newtonsoft.Json 与 **OpenAI SDK**。CLI 头为 il-only，目标框架是 netstandard2.0（见 PDB 路径）。
→ 这是本次最有价值的单点：核心库把 OpenAI 客户端**编进了自己的共享库**，说明 AI 能力不止存在于浏览器扩展，桌面进程侧同样有出网调用的实现。该模块尚未在工作区完成反编译。

### `ZYing.UI.dll`（PE32，2.2 MB）
无壳。CLR 4.0 元数据 + codeview；另有一条许可相关启发式命中（启发式，不作为结论）。WinForms 控件与页面装配主要在这里。

### `ZYing.User.dll`（PE32，16 MB）与 `ZYing.Work.dll`（PE32，8 MB）
两者特征几乎一致：Newtonsoft.Json、.NET Core 元数据、**.NET Reactor 6.x（Control Flow + Anti-ILDASM）**，加上"通用加密""混淆""反分析"三条启发式。`ZYing.Work.dll` 另有一条"窃取类行为"的启发式命中——这是商用保护壳常见的误报（控制流混淆 + 字符串加密 + 反调试探针的形状与 stealermeta 检测器重叠），**不作为恶意判定**，也没有任何运行时证据支持。
→ 之前用 `rea` 托管静态通道分析 `ZYing.User.dll` 时，元数据检视在表示层预算处中止并返回 `coverage.state: partial`（见 `analysis/rea_user_identity.json`）；现在有了机理解释：不是尺寸问题，是保护壳改写后的元数据形状让解析器提前退出。

### `ZDto.dll`（PE32，4 KB）
极小的纯数据/契约程序集，无壳。适合用来做字段与 DTO 结构的对照表。

### `ZPdf.dll`（PE32，13.9 MB）
无壳，但有一条混淆启发式；体积主要来自内嵌字体/资源。业务价值低。

## 3. 构建与源码布局（来自 codeview 记录）

调试记录里的 PDB 路径揭示了原始工程结构：

- `ZYing.Core.dll` → `…\sources\ZYing\ZCore\obj\RELEASE\netstandard2.0\ZYing.Core.pdb`
- `ZYing.UI.dll` → `…\sources\ZYing\ZUI\obj\RELEASE\net6.0-windows\ZYing.UI.pdb`

可读出三件事：解决方案根目录名 `ZYing`，工程目录按 `ZCore` / `ZUI` 分层；核心库面向 netstandard2.0（可跨 .NET Framework 与 .NET 6 复用），UI 层锁 `net6.0-windows`（WinForms 专用 TFM）；构建为 RELEASE 配置且未strip调试信息。

## 4. 对取证的直接影响

1. **去混淆工具选择要换**：面对 .NET Reactor 6.x，通用 de4dot 分支对新版控制流还原与字符串解密覆盖不足，应使用针对该壳的 DeDRunner/NETReactorSlayer 一类专用工具；在此之前，`ZYing.User.dll` / `ZYing.Work.dll` 的"已解密源码"只应视为部分还原，凡涉及这两个模块的行为结论都需标注 coverage 限制。
2. **取证优先级重排**：`ZYing.Core.dll`（无壳 + OpenAI SDK）> `ZDto.dll`（无壳 + 契约）> `ZYing.UI.dll`（无壳 + 界面装配）> 两个被壳模块。前三个基本能白盒读通，已足够支撑自研 ERP 所需的命令、字段与调用形态证据。
3. **瘦客户端结论不变、且更强**：被保护的两个模块体积最大、逻辑最重，但它们暴露给外界的仍然只是 `ZClient` 命令名与参数集合（详见 `analysis/command_groups.txt` 与后续 `command-inventory.csv`）。保护壳藏住的是实现细节，而自研真正需要的接口形状已经在明面模块里。
4. **PDB 可用性**：codeview 未剥，意味着可以在合法范围内用符号名重建模块/类型清单，而不必靠混淆后的名字猜测。

## 5. 未做 / 未知

- 未运行目标程序，因此**未验证** .NET Reactor 的运行时解密路径、反调试触发条件与网络校验行为。
- 未做去混淆尝试（需另行安装专用工具，且解出的仍是第三方代码，只可参考不可复制）。
- `ZYing.Core.dll` 里的 OpenAI 客户端具体配置（端点、模型、是否走智赢中转）尚未反编译确认；扩展侧已确认是直连+自带 Key，桌面侧是否一致属未知。
- DIE 未对第三方库（如 ZPdf 内嵌字体）做进一步分离识别。
