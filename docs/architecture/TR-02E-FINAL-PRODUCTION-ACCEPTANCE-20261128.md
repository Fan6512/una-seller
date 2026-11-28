# TR-02E 最终生产验收与发布准备（2026-11-28）

## 结论

| 项目 | 结果 |
| --- | --- |
| 生产升级 `97b31a03` | **PASS**，SHA-256 `0fcb04fe…`，`vcs.revision=97b31a03…`、`vcs.modified=false`，仅重启 `una-seller-api`，`NRestarts=0`，本机/公网 `/healthz` 200、匿名 401 |
| managed 标记保护 | **PASS**（临时目录，新二进制）：持锁后、联网前拒绝，28ms 返回，词典文件零改动 |
| 24 条人工纠错 | **PASS**，80→104，24/24 官方完整路径逐条命中，原 80 条字节级保留 |
| 历史排除清单 | **冻结 1,214 条**，SHA-256 `44538e7c…`；来源、交集与不能恢复项已披露 |
| 离线 assemble（`tr02e-final-v1`） | **PASS**，`selected=300`、与排除清单 0 重叠、104 覆盖全部生效、官方身份 0 变更 |
| 300 条复验 | **289/300（96.33%）**，11 条错误（9 条中文误译 + 2 条英文回退），全部逐条纠正并保留原始记录 |
| 最终候选 | SHA-256 `b3b9f46d71ebda0f8e72ceaac4e60a908a541861a8a1af190dc57671c089cccb`，10,928/10,954 有效中文覆盖（99.76%），26 条英文回退 |
| publish Dry-run | **PASS**，`applied=false`，`original_ok=289`、`audited=300`；active 未变、未写标记、未生成备份 |
| 正式发布 | **已执行**（2026-11-28 14:36 CST，负责人批准），`active.json` `be52b174…` → `b3b9f46d…` |

生产词典已切换为 `active.json` SHA-256 `b3b9f46d…`，中文 10,928/10,954，26 条回退英文；`llm-release-managed.json` 已写入，普通 `update` 被拒；普通翻译与词典增量 timer 保持禁用。

## 1. 生产升级

### 1.1 备份

`/var/backups/una-seller/api-20261128T053852Z-pre-97b31a03/`（33 MB，0700）：

- `una-seller-api`（`d8db7ae3…`，即 `c155228f`）、`api.env`（`699b0383…`）、`una-seller-api.service`（`68a4c433…`）
- `cbt-category/` 完整词典目录 15 个文件（含 `active.json`、`staging.json`、`overrides.json`、`progress.jsonl`、额度文件）
- `SHA256SUMS` 逐文件清单

回滚副本：`/opt/una-seller/una-seller-api.rollback-20261128T055537Z-c155228f`。

### 1.2 构建

- 源码：`https://github.com/Fan6512/una-seller` 浅克隆，HEAD `97b31a030e92387e28207b54db629770e64510bb`，工作树干净
- 构建：Docker `golang:1.26`，`go1.26.11 linux/amd64`，`CGO_ENABLED=0 go build -trimpath -buildvcs=true ./cmd/api`
- 二进制 SHA-256 **`0fcb04fe881c9cfad32a8ad01ee347ba3639a7fb4cdb836310f5a4f385de4be1`**（19,226,421 B）
- `vcs.revision=97b31a030e92387e28207b54db629770e64510bb`，`vcs.modified=false`
- 部署后复核 `/proc/<PID>/exe` 与新二进制哈希一致

### 1.3 升级结果

| 项 | 值 |
| --- | --- |
| 重启时间 | 2026-11-28 13:55:38 CST |
| MainPID | `2238246` |
| NRestarts | `0` |
| 本机 `/healthz` | 200 |
| 公网 `https://mkd.186512.xyz/healthz` | 200 |
| 匿名 `/api/v1/collections` | 401 |
| 启动日志 | `CBT category dictionary: state=loaded translated=80 version=2026-10-14T21:12:35.000Z` |
| 数据库 | 未迁移，未改表 |
| 桌面客户端 | 未重建 |

CLI 与 API 一致性：`/proc/2238246/exe`、`/opt/una-seller/una-seller-api` 与 CLI 调用路径三者 SHA-256 相同；API 环境 `CBT_CATEGORY_DICTIONARY_DIR=/var/lib/una-seller/cbt-category`，CLI 通过 `systemd-run --property=EnvironmentFile=/etc/una-seller/api.env` 使用同一服务账号、同一环境文件、同一词典目录。

### 1.4 managed 标记保护（临时目录）

`/var/lib/una-seller/cbt-review/guard-test/`，含 `staging.json`、`overrides.json`、`progress.jsonl` 副本及 `llm-release-managed.json`：

- `cbt-categories update` 返回：`ordinary CBT dictionary updater disabled after managed LLM release; use the reviewed incremental workflow`
- 耗时 **28 ms**（在发起任何网络请求前拒绝）
- 三个词典文件 SHA-256 前后一致；目录仅新增 `update.lock`
- 未在生产词典目录写入标记（Dry-run 不写标记，设计如此）

### 1.5 事故与纠正（必须披露）

第一次保护测试使用 `systemd-run --setenv=CBT_CATEGORY_DICTIONARY_DIR=<临时目录>`，但 systemd 的 `EnvironmentFile=` 覆盖了 `--setenv`，进程实际指向生产词典目录且无标记，**普通 update 于 13:57:37–13:58:01 对生产目录执行了约 24 秒**：

- `progress.jsonl` 243 → 346 行（新增 103 条普通机器译文）
- 普通额度计数 192,758 → **275,291**（+82,533 字符，真实上游消耗）
- `staging.json`、`active.json`、`overrides.json` 哈希未变；未发布
- 处置：证据留存 `/var/lib/una-seller/cbt-review/incident-20261128T0558Z/`；`progress.jsonl` 已按事故前字节精确还原（SHA-256 `1c0c9070…`，前缀逐行校验一致）；**额度计数保留 275,291 不回退**，以如实反映上游消耗
- 复测改用 `env -i` 隔离并加死代理兜底，结果如上（28 ms 拒绝）

该次误跑未触及大模型额度（847,341 未变），未产生候选、未发布；但违反“不运行普通全量 update / 不修改普通额度”约束，在此明确记录，不作为 PASS 项。

## 2. 24 条人工纠错

- 现场 `overrides.json` 起始 80 条（唯一 pathKey 80）。
- 逐条从**同一 `staging.json`** 取该 ID 的官方 PathNode 链，与 24 条清单 `path_names` 全等；`status=enabled`、`listing_allowed=true`（CBT392009 为中间节点除外，见第 4 节）、叶子 ID 一致。
- 24 条互不重复，且均不在原 80 条中；追加后 **104 条、唯一 104**。
- 原 80 条 JSON 与追加前逐条相等；104 条文件 SHA-256 `72657ea6687c093d447343102afd4fd6eb4fc5275243f130af7e63b1d2ea3781`。
- `review -build-plan` 复核：`manual_excluded=104`、`candidate_completed=10,814`、`glossary_excluded=8`、`deferred_english=28`、`eligible=0`，合计 10,954。

## 3. 历史排除清单与离线组装

### 3.1 排除清单（1,214 条唯一 ID）

冻结文件：`/var/lib/una-seller/cbt-review/tr02e-final-exclusions-20261128.json`，SHA-256 **`44538e7c09ca4bacf3926f511591161fd1b482a2d8e912ee8e2d240546124231`**（15,277 B）。

| 来源 | 数量 | 独占贡献 |
| --- | ---: | ---: |
| 历史 544 条（`prior-reviewed-ids-20261127.json`，SHA-256 `25ffd5b4…`） | 544 | 331 |
| 首批完整 200 条（全量构建前 200 条回执，额度 27,144→41,051） | 200 | 171 |
| TR-02D 第二轮独立 200 条（`plan-200-independent-v2.json`） | 200 | 0 |
| 首轮普通译文 200 条（`TR-02D-QUALITY-200`） | 200 | 0 |
| 最终 300 条（`TR-02E-QUALITY-300`） | 300 | 0 |
| 已知已审查 309 条（300 + 首批 9 条错误 ID） | 309 | 0 |
| 24 条纠错 ID | 24 | 0 |

交集披露（不宣称完全独立）：

- 最终 300 与历史 544：**0**；与首批 200：**0**；与第二轮 200：**10**
- 首批 200 与历史 544：**13**；与最终 300：**7**
- 24 条纠错：9 条在首批 200、15 条在最终 300
- 普通全量已产生的 302 条机器译文未经人工审查，按原口径不纳入排除（已在来源说明中保留）

未能恢复的历史记录：无。544 条现场清单完整可读，五类来源与原验收报告一致。

### 3.2 离线组装（seed `tr02e-final-v1`）

```sh
una-seller-api cbt-categories review -assemble -full-path -sample 300 \
  -seed tr02e-final-v1 \
  -exclude-ids /var/lib/una-seller/cbt-review/tr02e-final-exclusions-20261128.json
```

| 指标 | 值 |
| --- | ---: |
| enabled / disabled | 10,954 / 714 |
| manual / glossary / llm / deferred / unsafe | 104 / 8 / 10,814 / 28 / 0 |
| prior_excluded | 1,105 |
| audit_eligible / audit_selected | 9,737 / **300** |
| 有效非英文覆盖（原始候选） | 10,926 / 10,954（99.74%，含术语） |

- 原始候选 SHA-256 `0845a519923747c3fa02fb2f411b3f20de66e6efdce4ccfa8f063f9922591762`
- 原始审查报告 SHA-256 `d1da436d597964b9f9ea1852e4a3e9804a49d14f725023252a57631d5f70bbe4`，其中 `snapshot_sha256` 与候选文件哈希一致
- 排除清单在工具内的规范化哈希 `ca1a5845…`（与文件哈希 `44538e7c…` 为不同口径，均已记录）
- 官方 ID、完整路径、状态、刊登许可：0 变更；`version`、`retired` 完全一致
- 28 条延期在原始候选中 `zh` 为空，回退官方英文，未冒充中文译文

## 4. 300 条质量复验

原始 300 条样本在审查前冻结（`llm-release-audit-frozen-20261128.json`），复核全程使用同一批样本，未重新抽样、未替换失败样本。

| 指标 | 值 |
| --- | ---: |
| 样本 | 300（唯一 300，与排除清单 0 重叠） |
| 原始正确 | **289 / 300 = 96.33%** |
| 错误 | **11**（9 条中文误译 + 2 条英文回退） |
| 最终正确 | 300 / 300（逐条纠正后复核） |

错误与纠正（详见 [11 条纠错证据](TR-02E-ELEVEN-CORRECTIONS-20261128.json)）：

| ID | 原标签 | 纠正 | 问题 |
| --- | --- | --- | --- |
| CBT400500 | 阴道润滑剂 | 阴道保湿剂 | moisturizer 误译为润滑剂 |
| CBT194355 | 香料烘焙混合粉 | 咸味烘焙混合粉 | savory 误译为香料 |
| CBT392009 | 适用于搅拌机和处理器 | 适用于搅拌机和料理机 | processors 误译为处理器（CPU） |
| CBT388720 | 水桶、洗脸盆和浴缸 | 水桶、洗脸盆和洗衣盆 | 洗衣场景下 tubs 误译为浴缸 |
| CBT434353 | 网线 | 网织表带 | netting 误译为网线 |
| CBT417821 | 瑜伽垫 | 禅坐垫 | zafu 误译为瑜伽垫 |
| CBT74074 | 外壳 | 机身贴膜 | skins 误译为外壳 |
| CBT408041 | 吸入器隔离器 | 吸入器储雾罐 | spacer 误译为隔离器 |
| CBT458077 | 定位臂 | 外倾角调整件 | alignment camber 误译为定位臂 |
| CBT1147 | Sega Genesis / Mega Drive | 世嘉 Genesis / Mega Drive | 英文回退，补中文厂商名 |
| CBT38188 | USB | USB 线缆 | 英文回退，补中文类别词 |

英文回退不计为正确中文译文，故 `CBT1147`、`CBT38188` 计入错误并已补中文标签。

纠错后 `overrides.json` **115 条**（唯一 115，SHA-256 `6b507e74…`）；前 104 条与追加前字节级一致（`72657ea6…`）。重新组装后：

- 最终候选 SHA-256 **`b3b9f46d71ebda0f8e72ceaac4e60a908a541861a8a1af190dc57671c089cccb`**
- manual 115、llm 10,805、glossary 8、deferred 26、unsafe 0，合计 10,954
- 与原始候选逐类目比对：**仅 11 条 `zh` 变化，其余字段 0 变化**

对照表：[TR-02E-FINAL-QUALITY-300-20261128.csv](TR-02E-FINAL-QUALITY-300-20261128.csv)（含原始标签、原始判定、问题说明、建议译名、最终标签）。

## 5. 发布预检

approval.json（`/var/lib/una-seller/cbt-review/tr02e-approval-20261128.json`，SHA-256 **`db540e9f…`**）：

- `candidate_sha256` = `b3b9f46d…`
- `audit_sha256` = `d1da436d…`
- `corrections_sha256` = `3c0b48c654758781e12147463b3c29b0a92ec6d99847a36104bbc520715cc60d`（24 条纠错文件）
- `reviewer` = `tr02e-final-acceptance-20261128`（正式 apply 前须由负责人确认署名）
- 300 行：`original_ok` 289 true / 11 false，`final_ok` 300 true，`final_label` 与最终候选逐条一致

Dry-run（未加 `-apply`）：

```json
{"applied":false,"before_sha256":"be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959",
 "after_sha256":"b3b9f46d71ebda0f8e72ceaac4e60a908a541861a8a1af190dc57671c089cccb",
 "backup":"llm-release-backup-be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959.json",
 "original_ok":289,"audited":300}
```

- active SHA 一致（`be52b174…`）、候选一致、审查一致、24 条纠错路径与译名全部校验通过
- 300 条原始判定齐全，289 ≥ 285；11 条未通过项均为人工覆盖且标签已变更
- 未生成 `llm-release-managed.json`、未生成备份、`active.json` 未变
- 保护机制：实测新二进制在标记存在时联网前拒绝；正式 `-apply` 会先持久化标记，此后普通 `update` 与旧二进制均不得再执行

## 6. 交付与后续

| 交付物 | 位置 |
| --- | --- |
| 生产部署回执 | 本文第 1 节 |
| 24 条纠错核验 | 本文第 2 节 |
| 历史排除清单 | [TR-02E-FINAL-EXCLUSIONS-20261128.json](TR-02E-FINAL-EXCLUSIONS-20261128.json)、[来源与交集](TR-02E-FINAL-EXCLUSIONS-PROVENANCE-20261128.json) |
| 最终候选 SHA | `b3b9f46d71ebda0f8e72ceaac4e60a908a541861a8a1af190dc57671c089cccb` |
| 300 条质量 CSV | [TR-02E-FINAL-QUALITY-300-20261128.csv](TR-02E-FINAL-QUALITY-300-20261128.csv) |
| Dry-run 结果 | 本文第 5 节 |
| 台账 | [FEATURE-STATUS.md](../product/FEATURE-STATUS.md) |

后续与门禁：

1. ~~正式发布须负责人单独批准~~：已于 2026-11-28 14:36 CST 取得批准并执行，见第 7 节。
2. `-apply` 必须在同一数据目录、同一审核材料下执行，仅替换 `active.json`；随后核对 Reader 热加载、真实客户端类目中文与官方 ID、匿名 401 与健康检查。
3. 回滚只能使用新二进制的 `cbt-categories rollback`，不得使用旧二进制执行 `update`。
4. 每日增量仍需单独开发，不得直接启用现有 updater 或 timer。
5. 置信区间：289/300 点估计 96.33%，95% Wilson 区间约 93.6%–97.9%，仍为样本口径，不宣称总体统计认证。

## 7. 正式发布执行回执（2026-11-28 14:36 CST）

负责人于本会话明确批准后执行。审核材料与 Dry-run 完全一致，未修改任何文件：approval.json `db540e9f…`、候选 `b3b9f46d…`、冻结审查 `d1da436d…`、基线 `0845a519…`、排除清单 `44538e7c…`（工具口径 `ca1a5845…`）、24 条纠错 `3c0b48c6…`。

```sh
una-seller-api cbt-categories publish \
  -active-sha256 be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959 \
  -approval /var/lib/una-seller/cbt-review/tr02e-approval-20261128.json \
  -approval-sha256 db540e9f4564d08a9995f5d8af82005b07aaa16a6a96e4a19e955b32b809b93d \
  -audit /var/lib/una-seller/cbt-review/llm-release-audit-frozen-20261128.json \
  -baseline /var/lib/una-seller/cbt-review/llm-release-baseline-20261128.json \
  -exclude-ids /var/lib/una-seller/cbt-review/tr02e-final-exclusions-20261128.json \
  -corrections /var/lib/una-seller/cbt-review/TR-02E-TWENTY-FOUR-CORRECTIONS-20261128.json \
  -apply
```

```json
{"applied":true,
 "before_sha256":"be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959",
 "after_sha256":"b3b9f46d71ebda0f8e72ceaac4e60a908a541861a8a1af190dc57671c089cccb",
 "backup":"llm-release-backup-be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959.json",
 "original_ok":289,"audited":300}
```

发布后核验：

| 项 | 结果 |
| --- | --- |
| `active.json` | SHA-256 `b3b9f46d…`（与候选逐字节一致） |
| 备份 | `llm-release-backup-be52b174…json`，哈希 `be52b174…`，4,022,647 B |
| managed 标记 | 已写入 `{"managed":true}`（17 B） |
| 普通 `update` | 生产目录实测拒绝：`ordinary CBT dictionary updater disabled after managed LLM release`，129 ms（含 sudo/systemd 开销），普通额度保持 275,291 未变 |
| 服务 | 14:36:56 CST 重启加载，PID 2252694、`NRestarts=0`、`ActiveState=active` |
| 健康 | 本机 `/healthz` 200、公网 `/healthz` 200、匿名 `/api/v1/collections` 401 |
| 词典加载 | `state=loaded translated=10927 version=2026-10-14T21:12:35.000Z` |
| 未被改动 | 未重建桌面客户端、未改数据库、未启用词典增量 timer（仅原有 `una-seller-db-backup.timer`） |

计数说明：候选内非空中文标签 10,928 条；启动状态按 `label != 官方英文名` 统计为 10,927，差额 1 条为 `CBT106855 PS Vita - PlayStation Vita`（译名与英文名相同）。26 条延期为官方英文回退。

回滚（默认 dry-run，需另行批准才 `-apply`）：

```sh
una-seller-api cbt-categories rollback \
  -active-sha256 b3b9f46d71ebda0f8e72ceaac4e60a908a541861a8a1af190dc57671c089cccb \
  -backup-sha256 be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959
```

未覆盖项：Reader 热加载与真实客户端类目中文显示需已登录 Bearer Token，本会话无生产用户令牌，未做客户端实机读取；已通过重启加载证明发布文件可被正常解析，热加载路径由同一 mtime 变更逻辑驱动。
