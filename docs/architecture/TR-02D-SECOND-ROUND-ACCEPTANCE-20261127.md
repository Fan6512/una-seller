# TR-02D 第二轮生产部署与独立质量验收（2026-11-27）

## 结论

| 项目 | 结果 |
| --- | --- |
| 生产部署 `83b505fa` | **PASS** |
| 六条人工纠错 64→70 | **PASS** |
| 186 字符差额追溯 | **PASS**（两笔 93 字符被拒重试） |
| 独立 200 条大模型准确率 | **190/200（95.0%）**，点估计达标，95% CI 91.0%–97.3% |
| 明显分类语义错误 = 0 | **FAIL**（10 条叶子级错误，0 条跨一级行业） |
| 综合判定 | **CONDITIONAL · 候选不发布、timer 不启用、全量不恢复** |

容量与安全余量：

- 完整路径全量预算 **818,169 字符**（10,482 输入），90 万上限扣除已用 27,144 后余量 54,687。
- 父级路径全量预算 **332,696 字符**，余量 540,160。
- 当前 5 万字符试验硬上限只剩 **22,856**，两种全量方案都超出该上限，需另行授权扩额，本轮未扩。

---

## 1. 生产部署

- 固定业务提交：`83b505faa15955d54cdc2e2a635334c1aac58c10`（#128 Squash Merge）
- 变更范围：5 个后端文件（`cmd/api/category_dictionary.go`、`internal/cbtcategory/review.go`、`review_provider.go`、`review_provider_test.go`、`review_test.go`）+ 文档；`backend/migrations/` 无新增
- CI：审核 HEAD `d4156e36` 的 `46018393964` PASS；合流 `46019671633` PASS（Desktop/MV3 按范围跳过）
- 构建：Docker `golang:1.26`，`-trimpath`，`vcs.revision=83b505fa…`，`vcs.modified=false`
- 二进制 SHA-256：`652920fd84cb26c6170a4ced689ca35ec028a7ea1f27943bd102e1140145467b`
- 备份：`/var/backups/una-seller/api-20261127T150457Z-pre-83b505fa/`（二进制、`api.env`、unit、词典、进度、额度）
- 回滚副本：`/opt/una-seller/una-seller-api.rollback-20261127T150457Z-83b505fa`
- 仅重启 `una-seller-api`：PID `1974968`、`NRestarts=0`、`ActiveState=active`
- 本机 `/healthz` 200、公网 `/healthz` 200、匿名 `/api/v1/collections` 401
- 启动日志：`dictionary state=loaded translated=64`；无 panic/fatal
- 未重建桌面客户端，未新增数据库迁移

## 2. 六条人工纠错（64→70）

官方 `staging.json` 快照逐条核实完整 PathNode ID/名称链，全部与交接清单的 `path_names` 一致；六条此前均无覆盖，直接追加，不替换既有条目。

| ID | 官方完整路径 | 新译名 |
| --- | --- | --- |
| CBT10236 | Antiques and Collectibles > Antiques > Seltzer Bottles | 苏打水瓶 |
| CBT2747 | Other categories > Adults > Sex Toys > Dildos & Vibrators | 假阳具与振动器 |
| CBT411889 | Tools > Garden Tools > Chainsaws > Combustion Chamber Chainsaws | 内燃式链锯 |
| CBT412112 | Other categories > Tattoo Supplies > Armrests | 纹身扶手架 |
| CBT417159 | Art, Craft and School Supplies > Sewing Notions > Sewing & Embroidery > Embroidery Frames | 刺绣绣框 |
| CBT417203 | Beauty and Personal Care > Hair Salon Listings > Hair Salon Furniture > Hair Steamers | 蒸汽美发机 |

- 备份：`/var/backups/una-seller/overrides-20261127T150540Z-pre-70.json`，SHA-256 `4964771c8d936fd7d8ee62e61cbaba5899249d51420ffa3e87f743220a74f1d2`
- 结果：`overrides.json` **70 条**，叶子 ID 与完整路径均唯一
- 生产只读计划复核：`manual_excluded=70`，证明新文件可被同一 `loadOverrides` 正常加载

热加载验证：

- API 未为验证重启：PID `1974968`、启动时间 `23:04:57`、`NRestarts=0`
- `overrides.json` mtime `23:05:40`，晚于 API 启动；Reader 按 mtime 在下次请求时重载
- `TestDictionaryResumeIncrementalUpdatesAndManualPaths`、`TestReaderBackoffAndChangedFileRecovery`、`TestDictionaryFallbackForFullWidthAndLongAncestors` 全部 PASS
- 限制：本会话无生产用户 Bearer Token，未重新执行已登录桌面客户端读取；未重启 API

## 3. 历史排除 ID 与 186 字符差额

### 3.1 排除清单

合并五类真实已审查/纠错/试译 ID，去重后 **544 个**，写入 `/var/lib/una-seller/cbt-review/prior-reviewed-ids-20261127.json`：

| 来源 | 数量 | 独占贡献 |
| --- | ---: | ---: |
| 仓库首轮 CSV 200 + 被拒 CBT106855 | 201 | 0（均被其他来源覆盖） |
| 首轮普通译文 200 条抽检 `tr02-audit-200.tsv` | 200 | 166 |
| 早期 96 条旧译（progress 第 1–6 行） | 96 | 58 |
| 五批试译 80 条（progress 第 7–11 行） | 80 | 51 |
| 大模型候选 200 / 人工覆盖 70 | 200 / 70 | 0 |

说明：普通全量已产生的 3,683 条机器译文**未经过人工审查**，因此不纳入排除；否则独立样本会被压缩为“仅未译文”，无法按已有译文/未译文分层报告。该口径在报告中明确披露，不宣称样本与 3,683 条完全独立。

### 3.2 186 字符差额（精确对账）

| 阶段 | 行数 | 预扣字符 | 该轮结束累计 |
| --- | ---: | ---: | ---: |
| 探针（含首次 54001 失败） | 1+1+5 | 103+103+337 | 543 |
| 首轮 200 条（在 CBT106855 中止） | 161 | 10,815 | 11,451 |
| 补跑 33 条（再次在 CBT106855 中止） | 5 | 354 | **11,898**（与当时输出一致） |
| 临时快照补 28 条（移除 CBT106855） | 28 | 1,896 | **13,794** |

- 最终 200 条候选源字符合计 **13,505** = 440（六条探针）+ 10,815 + 354 + 1,896
- 本地累计 13,794 = 13,505 + **186** + 103
- 控制台 13,691 = 13,794 − 103 = 13,505 + 186

结论：**186 = 2 × 93**，即 `CBT106855`（PS Vita - PlayStation Vita，完整路径 93 字符）被请求两次。首轮 200 条在它处中止，补跑 33 条再次命中并中止，随后在临时快照中移除该 ID、用下一可用行替换。两次响应均无汉字、候选被拒，但 HTTP 已计费；首次 54001 鉴权失败未计费。

交接文档的 186 数值正确，但“194 条 13,065 对 13,251”的推导口径不成立：六条探针在 200 条计划中已被 `candidate_completed` 跳过，并未重复计费。实际构成为两笔被拒的 93 字符。

## 4. 独立 200 条质量验收

### 4.1 抽样

```sh
cbt-categories review -scope all -sample 200 -seed tr02d-independent-v2 \
  -full-path -exclude-ids /var/lib/una-seller/cbt-review/prior-reviewed-ids-20261127.json -trial
```

- `eligible=10,407`、`manual_excluded=70`、`glossary_excluded=8`、`candidate_completed=194`（当时）、`prior_excluded=275`
- `selected=200`、`source_chars=13,350`、`skipped_unsafe=0`
- 27 个一级行业；已有译文 73 条、未译文 127 条；单条 27–143 字符
- 与首轮 200 条候选 **ID 重叠 0**；与六条纠错 ID 重叠 0

### 4.2 执行

- `completed=200`、`attempts=200` 全部 `stored`
- `planned_source_chars=13,350`、`reserved_chars=13,350`、逐行预扣一致
- 本地大模型累计 **13,794 → 27,144**
- 无重试、无失败、无静默替换；未发布候选

### 4.3 质量结果

| 分层 | 正确 | 准确率 |
| --- | ---: | ---: |
| 已有译文 | 68/73 | 93.2% |
| 未译文 | 122/127 | 96.1% |
| **合计** | **190/200** | **95.0%** |

- 95% Wilson 置信区间约 **91.0%–97.3%**，下界低于 95%，不能统计认证“总体≥95%”
- 明显语义错误 **10 条**，全部为叶子级；跨一级行业误分类 **0 条**
- 零错误门禁 **未通过**

### 4.4 错误明细（未写入 overrides，仅建议）

| ID | 官方路径 | 大模型译文 | 问题 | 建议 |
| --- | --- | --- | --- | --- |
| CBT46006 | Babies > Baby Toys & Games > Rocking Chairs & Jumpers | 摇椅及蹦床 | Jumpers 为婴儿跳跳椅，非蹦床 | 摇椅与婴儿跳跳椅 |
| CBT417048 | Cell Phones and Phones > Walkie Talkies&Radio Frequency > Accessories > Cases | 手机壳 | 对讲机配件应为保护套，非手机壳 | 对讲机保护套 |
| CBT435188 | Musical Instruments > DJ Equipment & Accessories > DJ Accessories > Control Vinyl | 控制台 | Control Vinyl 为时间码唱片 | 控制台唱片（时间码黑胶） |
| CBT416841 | Sports and Fitness > Badminton > Balls | 球类 | 羽毛球并非球类运动 | 羽毛球 |
| CBT429486 | Musical Instruments > Wind Instruments > Shofar | 短号 | Shofar 为羊角号，非短号 | 羊角号 |
| CBT2740 | Other categories > Adults > Sex Toys > Rings | 戒指 | 成人用品环，非首饰戒指 | 锁精环（阴茎环） |
| CBT416798 | Health and Health Supplies > Orthopedics > Heel Cushions | 鞋垫 | 后跟垫非整鞋垫（普通译文“脚跟垫”更准） | 后跟垫 |
| CBT439056 | Consoles and Video Games > Console Accessories > For PlayStation > PSP > Cases, Protectors and Skins | 保护套、保护壳和皮肤 | Skins 为机身贴/保护贴 | 保护套、保护壳和机身贴 |
| CBT432092 | Sports and Fitness > Cycling > Accessories > Alarms | 闹钟 | 自行车防盗报警器，非闹钟 | 防盗报警器 |
| CBT414101 | Cameras and Accessories > Camera Accessories > Film Rolls > For Instant Cameras | 适用于一次性相机 | 拍立得/即时成像相机，非一次性相机 | 适用于拍立得相机 |

另有若干“可理解但不够地道”的表述（如 Network Cables→网络线、Seats→座椅、Levers→杠杆、Flares>Electric→电动），未计入 10 条明确错误。

## 5. 两种方案预算预检（只读，未执行）

| 方案 | 输入 | 源字符 | 90 万上限余量（扣 27,144） |
| --- | ---: | ---: | ---: |
| 完整路径 `-full-path` | 10,482 | **818,169** | 54,687 |
| 父级路径（默认） | 10,482 | **332,696** | 540,160 |

- 两者均排除 70 条人工覆盖、8 条固定术语、394 条已完成候选
- 当前 5 万字符硬上限剩余 22,856，两种方案均需扩额；本轮未扩、未执行
- 完整路径保留完整官方语境，余量仅 6.1%，需预留失败重试与增量空间；父级路径成本约为完整路径的 40.7%

## 6. 保护项

- `active.json`、`staging.json`、`progress.jsonl` MD5 均未变（`72d39773…` / `72d39773…` / `a897078e…`）
- `overrides.json` **70 条**，仅含已授权六条
- 候选文件 **400 条**（首轮 200 + 本轮 200），独立暂存，**未发布**
- 普通翻译额度仍 `{"month":"2026-11","chars":192758}`，未重置、未下调
- 大模型本地计数 `{"chars":27144}`；未扩大 5 万上限
- 未启用每日增量 timer；未恢复普通全量；未重建桌面客户端
- 无残留 `cbt-categories` 进程

## 7. 待核对

控制台大模型实际用量：上一轮 13,691 与本地 13,794 的差额恰为 103（54001 未计费）。若本轮计费口径不变，控制台应为 **27,041**。本地保留较高值 27,144，不下调；请以控制台实际读数确认。

## 8. 后续建议

1. 授权将本轮 10 条错误写入 `overrides.json`（70→80），或先只修 10 条后再抽 200 条复验。
2. 若坚持零错误门禁，需调整提示词/术语表并重新验收；按当前 95.0% 点估计，仍不建议直接发布。
3. 全量前必须明确选择完整路径或父级路径，并单独授权扩容 5 万上限。

## 9. 证据文件

- 质量对照表：[TR-02D2-QUALITY-200-20261127.csv](TR-02D2-QUALITY-200-20261127.csv)
- 排除清单：`/var/lib/una-seller/cbt-review/prior-reviewed-ids-20261127.json`
- 抽样计划：`/var/lib/una-seller/cbt-review/plan-200-independent-v2.json`
- 试译回执：`/var/lib/una-seller/cbt-review/trial-200-v2.jsonl`
- 覆盖备份：`/var/backups/una-seller/overrides-20261127T150540Z-pre-70.json`
