# PR #87 生产升级与真实 CBT C0 验收摘要（脱敏）

> 2026-11-24。依据负责人提供的现场部署报告、本人 Windows 手动操作反馈及随后生产数据库只读复核。仅记录可公开的版本与验收事实，不复制服务器地址、内部路径、进程号、配置文件摘要、账号 ID、凭据或原始响应。

## 生产部署

- 源码：`c066ee1ef46db46b7072f9310b51c506e8f116c7`（[PR #87](https://github.com/Fan6512/una-seller/pull/87)）。
- 现场报告：Go 1.26 Docker 构建，linux/amd64，CGO disabled，trimpath，干净源码版本；构建产物、磁盘文件及运行进程 SHA-256 一致。北京时间 14:29:40 完成切换，服务 active/running，无异常重启。
- 旧二进制双备份，哈希已核对。服务配置、Nginx、环境及其他业务未改动；未执行迁移、未触发回滚。以上是现场报告结论，并非仓库 CI 独立实测。
- 回环及公网 healthz 200；匿名店铺、采集箱、产品箱、账号配置 401；C0 GET 405（Allow POST）、匿名 POST 401。负责人已登录客户端成功触发 C0。

## 真实只读 C0

| 时间（北京时间） | 结果 | 依据 |
| --- | --- | --- |
| 13:48:14 | undetermined | 升级前历史证据 |
| 14:31:50 | standard_up | 现场部署报告、已登录 Windows 0.3.0 客户端及数据库回读 |
| 14:39:47 | standard_up | 负责人再次手动取证；随后独立生产数据库只读查询，UTC 06:39:47.217621Z |

生产证据标记 `c0_read_only_not_publish_approved`。账号允许销售、UP 标签 present、仓储标签 absent、站点映射 verified、复核时 OAuth token 有效。

| site_id | logistic_type | user_product | pricing_model | business_model（脱敏） |
| --- | --- | --- | --- | --- |
| MLB | remote | enabled | net_proceeds | other |
| MLM | remote | enabled | net_proceeds | other |
| MLC | remote | enabled | net_proceeds | other |
| MCO | remote | enabled | net_proceeds | other |
| MLA | remote | enabled | net_proceeds | other |

三次官方 GET（users/me、users/{merchant_id}、marketplace/users/{merchant_id}）只有均返回 200 且校验通过才允许落库。根据代码路径和完整证据推断 HTTP 200；未独立抓包或逐条上游打点。原始 business_model 值未保留，仅能确认非空且不等于已知精确 Managed 枚举；独立后台模式确认未执行。未导出 Token、未重新授权、未手动改库、未执行官方图片上传或商品创建。

## 验收裁决与后续

- CBT-01：单账号 C0 模式识别与生产部署 PASS。四标签 UI 及首次/重新授权自动探测的真实客户端验收仍待完成；合成测试不能替代真实授权。
- CBT-02：进入开发，实施 PR 待创建。第一切片为 standard UP + net_proceeds + remote 单商品、单站点、无变体，复用产品箱，联合 CBT-03、IMG-01、PUB-03。开发和模拟测试可推进；真实写入前需独立核验原始业务枚举、卖家后台模式、目标站点官方权限，并取得真实上传/创建批准；不得把只读分类当成写权限。
- 按既有最小工程约定避免无关迁移、重复部署及每项功能单独打包。

相关：[C0 合同](CBT-PUBLISH-C0.md) · [功能台账](../product/FEATURE-STATUS.md) · [Issue #50](https://github.com/Fan6512/una-seller/issues/50)。
