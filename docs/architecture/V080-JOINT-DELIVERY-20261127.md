# V0.8.0 联合交付回执

日期：2026-11-27（北京时间），2026-11-28 复核 GitHub 发布资产。固定源码：`6a56734fb9671f2b863d9ef1d8dcd3de64c6d8d6`。#122 已 Squash Merge，五个版本文件均为 0.8.0。

## 后端核验

负责人要求若已升级则跳过。本轮只读确认 ubuntu@152.136.41.105、VM-0-13-ubuntu，una-seller-api 用户 una-seller，PID 1650639、active、NRestarts=0。磁盘与运行进程 SHA-256 均为 `5d11e2d92faa5a2dfa38822f63adf9e6b443c3718a39a87710b08f8ab9f42284`。Go 构建信息 linux/amd64、vcs.revision 为固定提交、vcs.modified=false。公网健康200、匿名stores401。因此跳过后端部署；未重复切换、重启或执行迁移。

既有加密备份服务于17:33:08完成，Result=success、ExecMainStatus=0；`app-20261127T093308375228157Z.dump.gpg`，328955字节，SHA-256 `83cb6621392475c5edd520aef9d0bc40efcedf1812f795ec9562c301ba62aca6`。既有脚本包含解密及pg_restore归档检查，本轮未另做隔离恢复。本轮未独立验证授权订单、商品、财务、翻译配置及通知消费；不以匿名401代替授权业务验收。

## 签名与发布

- main CI：[45971731341](https://github.com/Fan6512/una-seller/actions/runs/45971731341) success（固定提交 `6a56734f`；版本修改范围跳过 Go）。
- 签名 CI：[45974523186](https://github.com/Fan6512/una-seller/actions/runs/45974523186) success；production-environment、Windows Tauri x64 (NSIS)、Verify and stage GitHub Release draft 三个 Job 均 success。
- tag `desktop-test-v0.8.0` 指向固定提交。
- [Release](https://github.com/Fan6512/una-seller/releases/tag/desktop-test-v0.8.0)：draft=false、prerelease=true、latest=false。
- 安装包 `Una-Seller_0.8.0_x64-setup.exe`，3,210,490字节；SHA-256 `a6a7f3bc7be9a16898c7ea90fbf581c2e1a8a7be051b3dd5182bc1b6c0ffc36a`，与 GitHub asset digest 一致。
- `latest.json` 1,150字节，SHA-256 `72571c8dfc41cdf00f55bafd07982a8e387848cb5303bbbdc0c33ffe9effe527`；签名文件 `Una-Seller_0.8.0_x64-setup.exe.sig` 440字节，SHA-256 `f7d7c66270986e9002d3621b4d4b6808ed5e7d6e09053571747712b99892e3ce`。
- 工作流输出 Signature and comment signature verified，可信注释绑定 version:0.8.0。此处签名为 Tauri OTA 签名，不代表 Authenticode。
- 2026-11-27 现场回执：latest.json version、安装包URL及签名与下载资产逐项匹配；正式 OTA 通道204、no-store；测试通道缓存自然刷新后返回200、version=0.8.0，signature与URL和已核验manifest一致；正式通道再次返回204。未为刷新缓存重启服务。
- 2026-11-28 复核：GitHub Release 仍为已发布 Pre-release，安装包、manifest、签名三个资产均存在，SHA-256 / GitHub digest 一致。此项仅核验发布资产，不代表本机已安装。

## Windows 人工验收清单（NOT_RUN）

1. 从 V0.7.0 开启“接收测试版更新”，发现0.8.0；安装后版本正确、重启可用，登录、配置及本地资料保留。
2. 关闭测试版开关，不提供0.8.0；已安装0.8.0不得提示降级。
3. 店铺、订单、商品及财务列表/详情可打开，筛选范围正确，财务估算与实际到账口径清楚。
4. 已刊登商品家族图片墙和刊登视图正常；切换家族不残留上一个家族图片；检查3/5/10规格布局、窗口重排及右侧详情。
5. 翻译配置显示火山通用机器翻译，不出现已移除的豆包配置；保存、错误提示及类目批次操作正常。火山账号当前曾返回 `-403 account status abnormal: unsynchronized`，真实翻译调用需待账号恢复后再验收，不把配置 UI 通过视为真实翻译 PASS。
6. 通知状态正常，不触发通知回放。产品编辑仅检查展示，不提交真实刊登或修改官方商品。

V0.7.0 保留。未新增功能、未修改官方商品、未触发真实刊登或通知回放。真实 Windows 安装、升级、重启和 GUI 验收须由负责人在目标 Windows 设备完成；本回执不将其标记 PASS。
