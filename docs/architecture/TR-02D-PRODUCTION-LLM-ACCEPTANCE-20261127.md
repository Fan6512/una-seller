# TR-02D 生产部署与大模型质量验收报告（2026-11-27）

## 1. 结论

| 项目 | 结果 |
| --- | --- |
| 固定提交与 CI | PASS。`4c387e7cb074f68ae4a7b92315fd10764ac8f648`，7 个后端文件，无新增迁移；CI `46008869942` `success` |
| 生产部署 | PASS。`vcs.revision=4c387e7c`，`vcs.modified=false`，SHA-256 `c78fd0de…`；仅重启 API，PID `1901794`，NRestarts=0，健康 200、匿名 401 |
| 大模型配置 | PASS。`BAIDU_LLM_API_KEY` 已写入 `api.env`，未输出；APP ID 沿用普通翻译 APP ID；普通 `TRANSLATION_PROVIDER=baidu` 未改 |
| 额度对账 | 大模型控制台确认用量 0，已初始化为 `{"chars":0}`；普通翻译本地 192,758 未下调 |
| 单条真实试译 | PASS（第二次）。首次密钥返回 `54001`，本地保留 103；更新密钥后 1 条成功，累计 206 |
| 5 条小规模试译 | PASS。5/5 成功，累计 543 |
| 200 条质量对照 | **CONDITIONAL**。大模型准确率 **97.0%（194/200）**，达到 95%；但仍有 6 条叶子级语义错误，按“0 条明显分类语义错误”的严格口径未通过 |
| 人工覆盖/词典 | PASS。64 条 overrides 未变；`active.json`、`staging.json`、`progress.jsonl` 哈希未变；候选独立暂存，未发布 |
| 控制台计费 | **未完全核实**。本地大模型预扣 13,794；控制台实际扣额及 `reference` 是否额外计费需负责人核对 |
| 授权边界 | 未恢复普通全量、未发布候选词典、未启用每日增量、未提高 5 万字符上限、未重建客户端 |

**结论：部署与接口链路通过；大模型准确率达标；零分类语义错误门禁未满足，候选不可直接发布。**

## 2. 固定提交与代码范围

- 固定提交：`4c387e7cb074f68ae4a7b92315fd10764ac8f648`（#127 Squash Merge）。
- 变更：7 个 Go 文件 + `docs/architecture/TR-02D-LLM-REVIEW.md`。
- `backend/migrations/` 无差异，无新增数据库迁移。
- CI `46008869942`：`conclusion=success`，`headSha=4c387e7c…`。

## 3. 生产部署

| 项目 | 结果 |
| --- | --- |
| 构建 | Docker `golang:1.26`，`go1.26.11`，`-trimpath`；`vcs.revision=4c387e7c`，`vcs.modified=false` |
| 二进制 SHA-256 | `c78fd0de8b9cee19c3ef1c91a2147795d18e85208680778d24d630e715ea9a6a` |
| 备份 | `/var/backups/una-seller/api-20261127T134710Z-pre-4c387e7c/`（二进制、`api.env`、unit、词典、进度、普通/大模型额度） |
| 回滚副本 | `/opt/una-seller/una-seller-api.rollback-20261127T134710Z-4c387e7c` |
| 重启范围 | 仅 `una-seller-api`；`MainPID=1901794`，`NRestarts=0`，`ActiveEnterTimestamp=2026-11-27 21:47:10 CST` |
| 健康 | 本机/公网 `/healthz` 200，匿名 `/api/v1/collections` 401，无 panic |
| 词典启动日志 | `state=loaded translated=64`，64 条人工覆盖已加载 |
| 迁移 | 无 |

## 4. 大模型配置与额度

- `BAIDU_LLM_API_KEY`：已存在，长度 25，无引号/空白/CR；未打印、未提交。
- `BAIDU_LLM_APP_ID`：与 `BAIDU_TRANSLATION_APP_ID` 一致。
- `TRANSLATION_PROVIDER=baidu` 保持；普通在线翻译服务未改。
- 大模型累计额度文件：`/var/lib/una-seller/cbt-category/baidu-llm-usage.json`，独立于普通 `baidu-usage.json`。
- 按负责人确认的控制台实际用量 `0` 执行：

```sh
review -reconcile-usage 0
```

返回 `{"chars":0}`。普通翻译计数保持 `{"month":"2026-11","chars":192758}`，未下调。

## 5. 单条与 5 条真实试译

### 5.1 首次单条失败

首次使用旧密钥执行：

```sh
review -scope translated -sample 1 -seed tr02d-probe -full-path -trial
```

返回：

```json
{"selected":1,"completed":0,"usage_chars":103}
LLM provider business error code=54001
```

`54001` 为百度鉴权/签名错误；Bearer 模式下表示 API Key 未通过鉴权。请求失败但本地预扣按设计保留 103，未生成候选。

### 5.2 更新密钥后

| 项目 | 结果 |
| --- | --- |
| 单条 | `selected=1`、`completed=1`，累计 206 |
| 5 条 | `selected=5`、`completed=5`，累计 543 |
| 单条候选 | `Electronics, Audio and Video > Audio > DJ Equipment & Accessories > Lights & Effects > Stage LED Lights` → 舞台LED灯 |
| 5 条候选 | 录音带、塑料模型制作、应答机、救生衣、摄影棚套装 |
| 鉴权 | Bearer 成功；未发送 sign/salt |
| 请求合同 | `model_type=llm`、固定中文电商 `reference`、`needIntervene=0`；代码固定，未在日志输出密钥或原文 |
| 响应校验 | `from=en`、`to=zh`、单个 `trans_result`、`src` 与请求一致；分隔符丢失则拒绝 |
| 候选文件 | `llm-review-candidates.json` 独立保存，首条 1 行，随后累计 6 行 |

本地累计从 0 增至 543；控制台实际扣额待核对。

## 6. 200 条质量对照

### 6.1 方法

```sh
review -scope translated -sample 200 -seed tr02d-quality-v1 -full-path -trial
```

- 按一级行业分层、固定 seed 哈希排序，逐条请求，1 秒间隔，无重试。
- 每行保存官方完整路径、普通译文（`original`）、大模型候选（`candidate`）。
- 评审使用随机 A/B 列位盲评，评审完成后再解码；对照表见 `TR-02D-QUALITY-200-20261127.csv`。
- 原 200 条选择包含 `CBT106855`（`PS Vita - PlayStation Vita`）。该行大模型输出无汉字，按 `validReviewRow` 的“非中文专名暂不接受候选”规则被拒绝。为保持 200 条评估量，使用同一 seed、同一分层顺序的下一可用行替换；最终为 **199 条原始 ID + 1 条同层替换**，未冒充原样本。

### 6.2 总量

| 指标 | 普通翻译 | 大模型 |
| --- | ---: | ---: |
| 正确 | 172/200（86.0%） | **194/200（97.0%）** |
| 明确语义错误 | 28 | **6** |
| 跨一级行业误分类 | 5 | **0** |
| 双方均正确 | — | 167 |
| 普通错、大模型对 | — | 27 |
| 普通对、大模型错 | — | 5 |
| 双方均错 | — | 1 |

- 200 条候选源字符合计 13,505。
- 本地大模型累计 13,794：探针与小样本 543（首次失败 103、成功单条 103、5 条 337）+ 200 条样本 13,251。
- 大模型对普通译文有差异 98 条；其中 27 条纠正了普通错误，5 条发生回退。

### 6.3 大模型 6 条语义错误

| 官方完整路径 | 普通译文 | 大模型候选 | 问题 |
| --- | --- | --- | --- |
| Antiques and Collectibles > Antiques > Seltzer Bottles | 苏打水瓶 | 矿泉水瓶 | 产品类型偏移 |
| Other categories > Adults > Sex Toys > Dildos & Vibrators | 假阳具和振动器 | 假阳刚 & 震动器 | 词义错误/错字 |
| Tools > Garden Tools > Chainsaws > Combustion Chamber Chainsaws | 燃烧室链锯 | 燃烧室电锯 | 链锯误作电锯 |
| Other categories > Tattoo Supplies > Armrests | 扶手 | 臂架 | 产品类型偏移 |
| Art, Craft and School Supplies > Sewing Notions > Sewing & Embroidery > Embroidery Frames | 刺绣框架 | 绣花机架 | 框架误作机架 |
| Beauty and Personal Care > Hair Salon Listings > Hair Salon Furniture > Hair Steamers | 蒸发器 | 蒸汽直发器 | 蒸汽美发机误作直发器 |

- 6 条均在原一级行业内，**0 条跨一级行业误分类**。
- 但 6 条均改变了叶子商品类型；若按上一轮“Axes→轴、Vises→视觉”同口径，第二项“明显分类语义错误为 0”**未满足**。
- 5 条回退（普通正确、大模型错误）：Seltzer Bottles、Dildos & Vibrators、Combustion Chamber Chainsaws、Armrests、Embroidery Frames。

## 7. 生产状态与保护

| 项目 | 结果 |
| --- | --- |
| `active.json` | MD5 `72d3977314d99fe246dc2acca4ce3217`，未变 |
| `staging.json` | MD5 `72d3977314d99fe246dc2acca4ce3217`，未变 |
| `progress.jsonl` | MD5 `a897078e38095708950d2ab4b5f8b0e4`，未变 |
| `overrides.json` | 64 条，未变 |
| 普通额度 | `{"month":"2026-11","chars":192758}`，未变 |
| 大模型额度 | `{"chars":13794}`，未重置 |
| 候选 | 200 条，独立文件，未发布 |
| 普通全量 | 仍暂停于 3,779/10,954（34.08%） |
| 每日增量 | 未启用 |

API 未为验证重启；`MainPID=1901794`、`NRestarts=0`。大模型 trial 只写 `llm-review-candidates.json` 与 `baidu-llm-usage.json`，不触碰 `active/staging/progress/overrides`。

## 8. 计费核对

| 项目 | 本地 | 控制台 |
| --- | ---: | ---: |
| 普通翻译 | 192,758 | 负责人此前确认 136,619；按较高值保留 192,758 |
| 大模型 | **13,794** | **待负责人核对** |

- 本地预扣按逐行 `q` 源字符计，`reference` 指令字段不计入本地预扣。
- 若控制台实际等于 13,794，说明 `reference` 未额外计费；若更高，差额即 `reference`/计费口径差异，需向上对账，不能下调本地计数。
- 首条失败请求的 103 字符已保留，未退款。

## 9. 验收判定

| 条件 | 结果 |
| --- | --- |
| 大模型准确率 ≥95% | **PASS**（97.0%） |
| 明显改变分类含义的错误 = 0 | **FAIL（严格口径）**：6 条叶子级错误；跨一级行业为 0 |
| 64 条人工覆盖不变 | PASS |
| 普通词典及进度不变 | PASS |
| 控制台计费与本地预扣无异常 | **NOT_VERIFIED**：控制台读数待核对 |

**综合：CONDITIONAL。准确率达标，但零分类语义错误未满足；200 条候选只能作为人工复核输入，不得直接发布。**

## 10. 证据

- 部署备份：`/var/backups/una-seller/api-20261127T134710Z-pre-4c387e7c/`。
- 候选导入前备份：`/var/backups/una-seller/tr02d-pre-import-20261127T142025Z`。
- 候选：`/var/lib/una-seller/cbt-category/llm-review-candidates.json`（200 行）。
- 大模型额度：`/var/lib/una-seller/cbt-category/baidu-llm-usage.json`。
- 对照表：[TR-02D-QUALITY-200-20261127.csv](TR-02D-QUALITY-200-20261127.csv)。
- 关联：[#127](https://github.com/Fan6512/una-seller/pull/127)、[CI 46008869942](https://github.com/Fan6512/una-seller/actions/runs/46008869942)。

## 11. 建议

1. 先核对百度大模型控制台实际扣额；若高于 13,794，执行向上对账并记录 `reference` 计费结论。
2. 对 6 条叶子级错误做人工候选修正或补充术语/指令后再抽样；在修正前不发布候选、不扩大 5 万上限。
3. 普通全量继续暂停；每日增量仍不启用。
