# TR-02D · 大模型文本翻译质量校正（第一切片）

## 当前门禁

- 百度翻译开放平台产品和定价已核：100万测试字符，源字符计费，10 QPS，指令及术语库不额外收费（https://api.fanyi.baidu.com/product/132）。
- 开放平台接入页 https://fanyi-api.baidu.com/product/133 与 /doc/21 未返回完整合同。第三方实现可见 /ait/api/aiTextTranslate，但 MD5 与 Bearer 鉴权示例不一致；不猜测签名、model_type、指令字段或返回。百度智能云 OAuth 合同不能替代本账号平台。
- 此切片无网络调用、无线上 provider、无发布命令，不接触凭据。通用翻译保持暂停，原192758本地预扣与136619控制台观测不下调。

## 离线命令

在生产同一服务账号及词典目录环境下，先备份，确认全量任务已停止。以下命令不调用翻译：

```sh
una-seller-api cbt-categories review -scope all
una-seller-api cbt-categories review -scope all -full-path
una-seller-api cbt-categories review -scope translated -sample 200 -seed tr02d-independent-v1 -export
una-seller-api cbt-categories review -scope missing -sample 200 -seed tr02d-independent-v2 -export
```

按一级行业轮询分层、固定seed哈希排序；不包含人工覆盖与已有候选。source_chars 为逐行源文本字符和，不含批处理换行、指令可能的额外字符及重试；真实合同核定后重算。历史200条对照须从原验收清单取精确ID/路径，不能用新随机样本冒充。

导出JSON包含 version/rows，填入每条 candidate（简体中文叶子标签）后，`review -import /绝对路径/review.json` 严格验证版本、完整路径、原译文及人工覆盖，保存到 llm-review-candidates.json。重复同结果幂等，冲突拒绝。非中文专名暂不接受候选，留待人工复核。不修改 active/staging/progress/overrides；未提供 apply/publish。

## 额度

只有取得当前控制台真实用量且负责人授权，才运行 `review -reconcile-usage N` 初始化/向上对账 baidu-llm-usage.json。不要用此前普通翻译的136619。缺失/损坏状态拒绝预扣；累计上限900000，预扣接口支持试验上限50000，每次真实发送前预扣（后续 provider 必须接线，当前不调用）。无月度重置，不与通用额度混用。

## 第二切片（阻塞）

取得本账号开放平台官方请求示例（脱敏），确认 endpoint、鉴权、模型明确选择LLM、指令字段及响应结构、字符计费。接入独立CLI provider，先1条及5批核对实际计费和响应，再200条原样本盲评与独立200条分行业验收，≥95%且无明显分类语义错误。候选独立暂存，预算充分才扩展，最终经审核另行原子发布。保留64条人工覆盖。
