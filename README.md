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
  - `scripts/`：本次分析使用的脚本。

## 有意不提交（仅本地保留）

- `target/`：目标程序二进制，约 1.2 GB，属第三方专有文件。
- `analysis/decrypted/`、`analysis/decompiled/`、`analysis/strings/`、`analysis/dat/`、`analysis/zying_ext/`、`analysis/ext_slices/`：逐字反编译输出、解密字符串与资源容器。
- `analysis/plugin_search.txt`、`analysis/meli_terms*`、`analysis/meli_partner*`：含第三方会话 Cookie 样本，或第三方页面整篇转载。
- `analysis/tools/`：本地工具副本。

报告中的 `file:///E:/美客多逆向/...` 证据链接指向本机工作区，不随仓库分发；仓库只保留结论与派生清单。

## 使用边界

本仓库内容仅用于理解既有系统的功能形态，作为自研 ERP 的设计参考。业务实现一律走美客多官方 REST API：不复用私有协议、不复制第三方代码、不做页面自动化。
