# TR-02D · 大模型文本翻译质量校正（第一切片）

## 当前门禁

- 百度翻译开放平台产品和定价已核：100万测试字符，源字符计费，10 QPS，指令及术语库不额外收费（https://api.fanyi.baidu.com/product/132）。
- 负责人补充开放平台官方 curl 示例：POST https://fanyi-api.baidu.com/ait/api/aiTextTranslate，JSON appid/from/to/q，Authorization: Bearer YOUR_API_KEY。已接入这个固定端点；返回结构暂按公开实现的 trans_result/src/dst 严格校验，真实一条试译核实；未确认的指令/术语库字段不发送，不混用智能云 OAuth。
- 已实现独立 CLI trial provider，无实际网络调用、无线上默认 provider 切换、无发布命令。通用翻译保持暂停，原192758本地预扣与136619控制台观测不下调。

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

只有取得当前控制台真实用量且负责人授权，才运行 `review -reconcile-usage N` 初始化/向上对账 baidu-llm-usage.json。不要用此前普通翻译的136619。缺失/损坏状态拒绝预扣；累计上限900000，预扣接口支持试验上限50000，每次真实发送前预扣（已接线，当前未实际调用）。无月度重置，不与通用额度混用。

## 第二切片（阻塞）

取得本账号开放平台官方请求示例（脱敏），确认 endpoint、鉴权、模型明确选择LLM、指令字段及响应结构、字符计费。接入独立CLI provider，先1条及5批核对实际计费和响应，再200条原样本盲评与独立200条分行业验收，≥95%且无明显分类语义错误。候选独立暂存，预算充分才扩展，最终经审核另行原子发布。保留64条人工覆盖。

## 受限真实试译入口

服务端配置 BAIDU_LLM_API_KEY（必须显式配置，不自动复用通用密钥），BAIDU_LLM_APP_ID 可缺省沿用 BAIDU_TRANSLATION_APP_ID；仅 CLI 使用。凭据不进仓库。

先在控制台核实大模型实际已用字符，显式对账（不得直接猜0）。随后 `review -scope translated -sample 1 -seed tr02d-probe -trial`；通过后最多200条/次，累计5万硬上限。每行一次请求、1秒间隔、15秒超时、无重试；失败立即停且保留预扣。候选逐条写入独立文件，绝不发布。真实响应及控制台扣额核实后才允许扩展。完整官方参数表（尤其翻译指令、术语库、返回示例）仍需补齐，禁止凭猜测添加字段。当前选取的新分层样本不等同历史200条。
