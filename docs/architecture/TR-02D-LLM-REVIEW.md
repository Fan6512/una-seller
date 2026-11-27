# TR-02D · 大模型文本翻译质量校正（第一切片）

## 当前门禁

- 百度翻译开放平台产品和定价已核：100万测试字符，源字符计费，10 QPS，指令及术语库不额外收费（https://api.fanyi.baidu.com/product/132）。
- 负责人补充开放平台官方 curl 示例：POST https://fanyi-api.baidu.com/ait/api/aiTextTranslate，JSON appid/from/to/q，Authorization: Bearer YOUR_API_KEY。已接入这个固定端点；负责人进一步提供正式参数表与成功/异常响应：显式 model_type=llm、reference 固定中文电商指令、needIntervene=0（未配置术语库，不启用）；不发送仅 NMT 支持的 tag_handling/ignore_tags；Bearer 模式不发送 sign/salt。q 按建议限制2000字符（官方上限6000），严格校验 from/to 与 trans_result/src/dst，业务错误仅输出数字错误码，不打印 error_msg。真实一条试译仍需核实，不混用智能云 OAuth。
- 已实现独立 CLI trial provider，无实际网络调用、无线上默认 provider 切换、无发布命令。通用翻译保持暂停，原192758本地预扣与136619控制台观测不下调。

## 离线命令

在生产同一服务账号及词典目录环境下，先备份，确认全量任务已停止。以下命令不调用翻译：

```sh
una-seller-api cbt-categories review -scope all
una-seller-api cbt-categories review -scope all -full-path
una-seller-api cbt-categories review -scope translated -sample 200 -seed tr02d-independent-v1 -export
una-seller-api cbt-categories review -scope missing -sample 200 -seed tr02d-independent-v2 -export
```

按一级行业轮询分层、固定seed哈希排序；不包含人工覆盖、已匹配固定术语的类目、已有候选；不安全/超长源文在计划阶段排除并报告 skipped_unsafe，固定术语排除报告 glossary_excluded。source_chars 为逐行 q 源文本字符和，不含独立 reference 指令字段及重试；此计费假设必须在首条及小规模真实试译后与控制台核对，不得以此直接推断实际账单。历史200条对照须从原验收清单取精确ID/路径，不能用新随机样本冒充。

导出JSON包含 version/rows，填入每条 candidate（简体中文叶子标签）后，`review -import /绝对路径/review.json` 严格验证版本、完整路径、原译文及人工覆盖，保存到 llm-review-candidates.json。重复同结果幂等，冲突拒绝；导入结果 total_candidates 为合并后总数，不是本次新增数。过期、冲突、空候选及受保护条目保持 fail-closed 并报告 row ID，不静默丢弃已付费候选。修复时先确保没有 review/update 运行，备份 llm-review-candidates.json 到独立文件，将旧候选中与当前 version/path/original 一致且未被人工覆盖的有效行整理为 version/rows JSON，移走旧文件并通过 -import 重新导入；不要删额度文件或改 active/staging/progress。非中文仅允许按官方 ID、名称和固定候选严格匹配的已核实专名（首例 CBT106855 / PS Vita - PlayStation Vita）；其他未翻译英文仍拒绝。不修改 active/staging/progress/overrides；未提供 apply/publish。

## 额度

只有取得当前控制台真实用量且负责人授权，才运行 `review -reconcile-usage N` 初始化/向上对账 baidu-llm-usage.json。不要用此前普通翻译的136619。缺失/损坏状态拒绝预扣；额度保护器绝对上限900000；本 PR 的真实 CLI 只允许累计50000，未提供提高上限的开关。90万是后续全量阶段目标，不是当前可用预算；全量须另行审核并扩展执行器。每次真实发送前预扣（已接线，当前未实际调用）。无月度重置，不与通用额度混用。

## 第二切片（真实验收待执行）

负责人提供的官方请求、指令字段与响应合同已实现并增加模拟测试。独立CLI provider，先1条及5批核对实际计费和响应，再200条原样本盲评与独立200条分行业验收，≥95%且无明显分类语义错误。候选独立暂存，预算充分才扩展，最终经审核另行原子发布。保留64条人工覆盖。

## 受限真实试译入口

服务端配置 BAIDU_LLM_API_KEY（必须显式配置，不自动复用通用密钥），BAIDU_LLM_APP_ID 可缺省沿用 BAIDU_TRANSLATION_APP_ID；仅 CLI 使用。凭据不进仓库。

先在控制台核实大模型实际已用字符，显式对账（不得直接猜0）。随后 `review -scope translated -sample 1 -seed tr02d-probe -trial`；通过后最多200条/次，累计5万硬上限。每行一次请求、1秒间隔、15秒超时、无重试；失败立即停且保留预扣；成功、业务失败、上下文中断均重新读取 usage_chars；无法读取则报告 null 与错误，不伪造0。候选逐条写入独立文件，绝不发布。真实响应及控制台扣额核实后才允许扩展。正式参数表已补齐；术语库具体配置及干预效果待真实验证，needIntervene 暂为0。指令明确要求保留整条路径、分隔符与层级，候选只保存叶子译名；若输出丢失分隔符则拒绝，不写脏数据。当前选取的新分层样本不等同历史200条。

## 复审取舍

保留逐条落盘与失败立即停止：首轮用于验证官方协议和计费，未验证真实返回前不能在连续异常下继续消耗额度。普通 update 须保持暂停，review 与 update 的单次文件操作共享锁，但整个试译不是全程占用 update.lock；遇锁冲突停止并报告累计扣额。后续扩大规模时再优化批量读取/落盘、可恢复行级错误及全程互斥；不增加通用 saveJSON 的目录 fsync 行为变更。

## 第二轮独立验收（保持5万累计上限）

- 将历史普通200条、五批80条、已知人工纠错、首轮大模型200条、1/5探针及已拒绝/替换行的真实 ID 汇总为 JSON 字符串数组，作为 `-exclude-ids /path/prior-ids.json`。若历史清单不完整，明确披露，不能宣称独立。既有候选、术语、人工覆盖仍自动排除。固定 seed、full-path，先离线导出计划；记录 ID 和 SHA256。使用 scope all 覆盖已有及未译类目（报告分别统计），每次最多200条；同一排除文件、seed、快照和候选状态下执行真实 trial。中断时保存原计划、已尝试 ID 和结果；不得静默替换失败行，验收以冻结的样本集为准。
- CLI JSON 增加 planned_source_chars、usage_before_chars、reserved_chars、attempts（ID、q SHA256、source_chars、reserved_chars、outcome）。正常退出、普通错误及受控中断保留报告；进程被强杀不能保证输出，须保留实际额度文件及现有候选。导入候选不等于已真实请求，不能凭候选合计推导历史请求。
- 第二轮前备份，按照首轮CSV中官方 ID 和完整路径在生产快照中核实六条修正，若任一不符暂停；在原64条 overrides 上追加（不得覆盖既有不同值）。六条原始错误仍计入首轮统计，不更改原CSV。生产执行后才登记70条。建议叶子译名见下表。候选原始文件与请求证据归档保留，不发布全量。
- 真实200条验收：按原门禁准确率≥95%、明显改变商品分类含义错误=0；不将修正过的样本混入独立评估，不自动提高5万上限。只有质量、对账与候选来源可追溯通过后，才另行全量执行器及发布审核。

| ID | 官方完整路径 | 建议人工译名 |
| --- | --- | --- |
| CBT10236 | Antiques and Collectibles > Antiques > Seltzer Bottles | 苏打水瓶 |
| CBT2747 | Other categories > Adults > Sex Toys > Dildos & Vibrators | 假阳具与振动器 |
| CBT411889 | Tools > Garden Tools > Chainsaws > Combustion Chamber Chainsaws | 内燃式链锯 |
| CBT412112 | Other categories > Tattoo Supplies > Armrests | 纹身扶手架 |
| CBT417159 | Art, Craft and School Supplies > Sewing Notions > Sewing & Embroidery > Embroidery Frames | 刺绣绣框 |
| CBT417203 | Beauty and Personal Care > Hair Salon Listings > Hair Salon Furniture > Hair Steamers | 蒸汽美发机 |

译名仅按现有官方路径及验收报告建议；内燃式链锯避免原文未明确燃料时擅自限定汽油。上表不是已写入生产 overrides 的凭证。
