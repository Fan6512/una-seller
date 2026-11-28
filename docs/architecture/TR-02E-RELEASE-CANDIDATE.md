# TR-02E 第二切片：离线候选合并及审查样本

本切片可与生产全量翻译并行开发，不需要访问生产，也不会调用百度。固定 #129 生产部署版本不变；新功能仍须 CI/审核及单独部署或使用同版本离线副本。

全量 build-plan 必须 selected=0、pending=false、budget_fits=true。保留真实历史排除清单（包括首轮及第二轮所有 ID、旧审计与纠错）；冻结排除清单 SHA256 和 seed。示例：

```sh
una-seller-api cbt-categories review -assemble -full-path -sample 300 -seed tr02e-final-v1 -exclude-ids /absolute/path/prior-reviewed-ids.json
```

只写 llm-release-candidate.json 和 llm-release-audit.json；不发布 active、不修改 staging/progress/overrides/额度，不调用上游。按人工覆盖 > 当前已验证固定术语 > 完整路径 LLM > 官方英文处理；未经复核普通译文不作为 enabled 类目兜底。保留停用类目与历史树的官方身份。核对官方版本、完整ID/路径、父节点、状态、刊登许可；结果中所有 enabled 类目按来源计数，无法确认译文回退英文。独立样本按一级行业轮询、稳定哈希排序、排除历史已审查及人工/术语；包括延期或不安全源的英文回退。人工审查需逐条记录准确、明显误译、英文回退，并报告原始得分及覆盖率，不能把所有英文回退自动计正确。300条是建议验收量，不表示总体统计认证。

审查报告含候选快照精确 SHA256、排除 ID 集合规范排序后的 SHA256、抽样 eligible/selected 数；归档初次样本、候选、排除清单及人工评审，不能改seed反复抽样取高分。候选结构校验同时核对 Retired 历史记录。发现错误写入精确路径覆盖，经真实核实后重新合并，核对原始评审及全部修正，不篡改原始准确率。两个文件分别原子保存，并非事务；中断后必须核对哈希，禁止凭旧报告发布。当前没有 publish 命令。正式原子发布、备份回滚、实际最终验收和每日增量另行收口。
