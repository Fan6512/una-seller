# TR-02E · 显式原子发布与回滚（第三切片）

负责人已批准开发工具，未批准正式发布。#130 的生产离线组装/104条人工覆盖/最终300条复验仍须提供真实回执。先冻结原始300条的 llm-release-audit.json 与 llm-release-candidate.json（原始候选另存 baseline），归档完整历史排除清单及其哈希。独立原始得分必须≥285/300。新发现错误全部人工精确路径纠正，重新 assemble 生成最终候选，但保留原始300条评审与原始候选，**不要以重抽的样本替换原审查**。如分数不足，停止发布。无需百度调用。

## 审核文件

由负责人审阅最终候选及评审记录后批准确切 approval.json SHA256。审批文件包含 candidate_sha256（最终候选精确字节）、audit_sha256（冻结原始审查）、corrections_sha256（已批准24条纠错JSON）、reviewer、rows（300行，每条 id、original_ok 布尔、final_ok 布尔、final_label）。不得自动将英文回退计正确；人工记录专名与无法确认类别。示例行：{"id":"CBT123","original_ok":false,"final_ok":true,"final_label":"人工修正标签"}。实际填写完整300条，不使用示例ID。历史排除清单须包括原544个、首批完整200个、最终300个、其他已检查/纠错ID，现场保留来源；工具验证清单哈希与无重叠，不代替人员对历史来源完整性的确认。

## 命令

```sh
una-seller-api cbt-categories publish -active-sha256 ACTIVE -approval /path/approval.json -approval-sha256 APPROVED_SHA -audit /path/frozen-audit.json -baseline /path/frozen-baseline.json -exclude-ids /path/frozen-exclusions.json -corrections /path/TR-02E-TWENTY-FOUR-CORRECTIONS-20261128.json
```

默认 dry-run，不写 active/backup。必须另获正式发布授权后在**同一数据目录、同一审核材料**使用 -apply。工具在双锁下复算候选、校验当前active、官方身份/刊登许可/Retired、冻结baseline/审查/排除哈希、300个原始判断（至少285正确）、全部最终判断、原错误已人工修正、24条已知错误对应的完整路径和译名；审查后非人工变动阻断。备份当前 active 精确字节为 llm-release-backup-<sha>.json，fsync 后原子替换 active 并 fsync 目录。以服务账户运行，目录权限保持私有。若 rename 后 fsync 错误，先读回 active hash 确认状态，不盲目重试。无数据库/网络/上游调用。

回滚也默认 dry-run，需明确当前 active 和备份哈希（同官方版本、树、刊登权限），另行 -apply：

```sh
una-seller-api cbt-categories rollback -active-sha256 CURRENT -backup-sha256 PREVIOUS
```

回滚只恢复 active.json，保留人工覆盖、源数据、额度及审查证据；不恢复普通翻译。发布后检查 Reader 热加载、真实客户端类目中文/官方ID不变、匿名认证及API健康、回退英文。**普通 update 命令仍可能发布旧 staging**，务必保持普通翻译和 timer 禁用；每日增量另行开发，不能直接启用现有 updater。本切片未实际部署或发布，需 CI/真实离线验收及单独批准。
