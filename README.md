# una-seller — 美客多 ERP 前期调研

本仓库保存对智赢（ZYing）美客多 ERP 的**静态逆向分析结论**，以及自研 AI ERP 的**合规接入清单与模块路线图**。

## 目录内容

- `ZYing-美客多逆向分析报告.md`：证据化逆向报告（架构、模块、ZClient 网关、授权、同步、风控）。
- `AI-ERP-合规接入清单与模块路线图.md`：美客多官方接入要求 + 新 ERP 的模块与阶段路线。
- `analysis/`：派生清单与分析脚本，包括
  - `command_groups.txt`、`analysis-meli-endpoints.txt`：ZClient 命令分组与端点；
  - `*-types.txt`、`user-type-names.json`、`resources.txt`：类型与资源清单；
  - `ext_endpoints.json`、`ext_campaign_report.txt`：官方接口直连与活动相关发现；
  - `managed-*.json`、`rea_user_identity.json`、`inspect-installer.json`：REA 静态检查证据；
  - `die-*.json`、`die_summary.txt`、`die-protection-report.md`：Detect It Easy 识别结果与保护壳/构建指纹结论；
  - `ai_providers.json`、`ai-gateway-findings.md`：AI 供应商注册表与批量核价编排协议的行为级观察；
  - `command-inventory.csv`：237 个 ZClient 命令调用点（193 个命令名、38 个域前缀）连同服务器池、窗体类、方法、行号与请求/返回字段；
  - `feature-map.md` + `menu-tree.json` + `field-index.json`：菜单功能 → 窗体类 → 命令 → 字段；29 个板块 265 个功能项，其中美客多板块 25 项有 18 项已落到具体命令；
  - `api-map.md` + `meli-direct-endpoints.json`：68 个直连官方 REST 端点（按资源域分组），以及直连/命令转发/自有 HTTP 三条出口的分界、令牌链路、限流与退避实证；
  - `data-model.md` + `record-families.json`：客户端无本地关系库这一结论的证据、15 个记录列前缀族、贯穿的连接键与 31 个直连 DTO；
  - `scripts/`：本次分析使用的脚本（`build_*.py` 为上述清单的生成器，可重复执行）。

## 有意不提交（仅本地保留）

- `target/`：目标程序二进制，约 1.2 GB，属第三方专有文件。
- `analysis/decrypted/`、`analysis/decompiled/`、`analysis/strings/`、`analysis/dat/`、`analysis/zying_ext/`、`analysis/ext_slices/`：逐字反编译输出、解密字符串与资源容器。
- `analysis/plugin_search.txt`、`analysis/meli_terms*`、`analysis/meli_partner*`：含第三方会话 Cookie 样本，或第三方页面整篇转载。
- `analysis/tools/`：本地工具副本。

报告中的 `file:///E:/美客多逆向/...` 证据链接指向本机工作区，不随仓库分发；仓库只保留结论与派生清单。

## 使用边界

本仓库内容仅用于理解既有系统的功能形态，作为自研 ERP 的设计参考。业务实现一律走美客多官方 REST API：不复用私有协议、不复制第三方代码、不做页面自动化。
