# TR-02E 公共词典完整路径全量建设与验收（2026-11-28）

## 结论

| 项目 | 结果 |
| --- | --- |
| 生产部署 `c155228f` | **PASS** |
| 十条人工纠错 70→80 | **PASS** |
| 完整路径全量构建 | **完成**：10,454 候选 + 28 延期，覆盖 10,838/10,954（**98.94%**） |
| 累计额度 | **847,341 / 900,000**，余 52,659，未超限 |
| 首批 200 条 | 191/200（**95.5%**），9 条错误 |
| 最终 300 条独立样本 | 285/300（**95.0%**，95% CI 91.9%–96.9%），15 条错误 |
| 发布状态 | **未发布** `active.json`；未启用每日增量；候选仅为离线建设结果 |

判定：**构建完成，质量点估计达到 95%，但 24 条已发现错误尚未修正，不满足发布前“已知错误全部修正”要求，故不发布。** 另置信区间下界 91.9%，不能宣称总体统计认证 ≥95%。

---

## 1. 生产部署

- 固定提交：`c155228f002a07111c0d6ec8c774bb65860cc1e1`（#129 Squash Merge）
- 变更：4 个后端文件（`cmd/api/category_dictionary.go`、`review.go`、`review_build.go`、`review_build_test.go`、`review_provider.go`、`review_quota.go`）+ 设计文档；`backend/migrations/` 无新增
- CI：审核 HEAD `7513b748` 的 `46026117555` PASS；合流 `46026908679` PASS
- 构建：Docker `golang:1.26`，`-trimpath`，`vcs.revision=c155228f…`，`vcs.modified=false`
- 二进制 SHA-256：`d8db7ae32205e1902345f7ec243ab6acdbabbcb4f3edef882c5f58e7ff2e3d82`
- 备份：`/var/backups/una-seller/api-20261127T154826Z-pre-c155228f/`（二进制、`api.env`、unit、词典、进度、额度）
- 回滚副本：`/opt/una-seller/una-seller-api.rollback-20261127T154826Z-c155228f`
- 仅重启 `una-seller-api`：PID `2007640`、`NRestarts=0`、本机/公网 `/healthz` 200、匿名 401
- 启动日志：`dictionary state=loaded translated=70`；无 panic/fatal
- 未重建桌面客户端

## 2. 十条人工纠错（70→80）

按官方 `staging.json` 快照逐条核对完整 PathNode ID/名称链，十条 `path_names` 全部一致且此前无覆盖，备份后追加，不替换既有条目。

| ID | 官方完整路径 | 译名 |
| --- | --- | --- |
| CBT46006 | Babies > Baby Toys & Games > Rocking Chairs & Jumpers | 摇椅与婴儿跳跳椅 |
| CBT417048 | Cell Phones and Phones > Walkie Talkies&Radio Frequency > Accessories > Cases | 对讲机保护套 |
| CBT435188 | Musical Instruments > DJ Equipment & Accessories > DJ Accessories > Control Vinyl | 时间码黑胶唱片 |
| CBT416841 | Sports and Fitness > Badminton > Balls | 羽毛球 |
| CBT429486 | Musical Instruments > Wind Instruments > Shofar | 羊角号 |
| CBT2740 | Other categories > Adults > Sex Toys > Rings | 阴茎环 |
| CBT416798 | Health and Health Supplies > Orthopedics > Heel Cushions | 后跟垫 |
| CBT439056 | Consoles and Video Games > Console Accessories > For PlayStation > PSP > Cases, Protectors and Skins | 保护套、保护壳和机身贴 |
| CBT432092 | Sports and Fitness > Cycling > Accessories > Alarms | 自行车防盗报警器 |
| CBT414101 | Cameras and Accessories > Camera Accessories > Film Rolls > For Instant Cameras | 适用于即时成像相机 |

- 备份：`/var/backups/una-seller/overrides-20261127T150540Z-pre-80.json`
- 结果：`overrides.json` **80 条**，叶子 ID 与完整路径均唯一
- 只读计划：`manual_excluded=80`；API 未因覆盖重启（Reader 按 mtime 热加载）

## 3. 额度与预算

- 本地大模型累计 **27,144**（保留较高值，不下调）；普通翻译计数 **192,758** 未变
- 全量前 `build-plan`：10,482 输入 / **818,169 字符**，`budget_fits=true`，余 54,687，`pending=false`
- 执行预算：`-budget 900000`，累计硬上限，不按月清零

## 4. 全量构建执行

- 入口：`review -build -full-path -budget 900000 -sample 200 -seed tr02e-v1`
- 首批 200 条完成并验收后，按序连续执行；共 42 次批次调用（第 42 次 `selected=0` 正常收尾）
- 结果：
  - 候选文件 **10,854 条**（含首两轮 400 条，去重唯一）
  - 本轮新增候选 **10,454 条**
  - 延期 `llm-build-deferred.json` **28 条**（官方英文回退）
  - 逐条回执 `llm-build-receipts.jsonl` **10,480 条** = 10,454 候选 + 26 自动延期
  - 最终 `build-plan`：`eligible=0`、`candidate_completed=10,838`、`deferred_english=28`、`manual_excluded=80`、`glossary_excluded=8`
- 覆盖率：**10,838 / 10,954 = 98.94%**；加 80 条人工覆盖为 99.67%；28 条延期（0.26%）发布时回退官方英文
- 额度：27,144 → **847,341**，实际消耗 **820,197**；距 90 万上限余 **52,659**
- 全程未发布、未启用 timer、未改普通 Provider

注：10,480 条回执与 28 条延期的差额 2 条为人工延期（CBT6662、CBT375063），其 pending 已归档，未生成 build 回执。

## 5. 异常与 pending 恢复

| 类型 | 次数 | 处理 |
| --- | ---: | --- |
| `30000` 业务错误（瞬时） | 15 次自动 + 若干人工 | 归档 pending、保留预扣、原批重试；最终全部越过 |
| 分隔符未保留 | 2 行（CBT6662、CBT47393） | CBT47393 重试成功；CBT6662 连续两次失败后转延期 |
| 响应源文不匹配 | 1 行（CBT375063） | 重试仍失败，转延期 |
| 人工停止/中断 | 2 次 | pending 归档后从原批次恢复 |

- 自动重试仅限 `30000` 且无候选落盘，最多 5 次/批、总计 200 次；每次均复制 pending 至 `loop/pending-archive/` 并记录 ID、源字符、预扣
- 结束时 **无 pending 残留**；`active/staging/progress` 未变
- 延期集中在主机平台专名（Xbox 360、PS3/PS4/PS5、Switch、Wii、Game Boy 等）及少量缩写/特殊源文，符合“非中文专名回退英文”的设计

## 6. 质量验收

### 6.1 首批 200 条

- 200/200 成功，额度 27,144 → 41,051
- 人工复核 **191/200（95.5%）**，9 条明显错误
- 达到 ≥95% 门禁，据此继续剩余批次

首批错误：

| ID | 路径 | 译文 | 建议 |
| --- | --- | --- | --- |
| CBT430680 | Sports and Fitness > Martial Arts & Boxing > Clothing > Shorts & Bermuda Shorts | 短裤与短裙 | 短裤与百慕大短裤 |
| CBT457948 | Electronics, Audio and Video > Projectors & Screens > Electric Screens | 电子屏幕 | 电动投影幕 |
| CBT66807 | Cell Phones and Phones > Tariffers and Cabins > Cabins | 机舱 | 电话亭 |
| CBT417127 | Clothing and Accessories > Fashion Accessories > Glasses & Accessories > Frames & non-Rx glasses | 镜框及非近视眼镜 | 镜框及非处方眼镜 |
| CBT455352 | Home, Furniture and Garden > Gardens & Outdoors > Swimming Pools > Pool Replacement Parts > Plugs | 插头 | 排水塞 |
| CBT435256 | Health and Health Supplies > Health Care > Anti-Snoring Devices > Mouth Guards | 口腔护理 | 防鼾牙套 |
| CBT457116 | Industries and Offices > Farming > Livestock Production > Tools and Accessories > Cattle Herders | 养牛户 | 赶牛器具 |
| CBT388931 | Vehicle Accessories > Motorcycle Replacement Parts > Chassis > Commands > Other Commands | 其他指令 | 其他操纵件 |
| CBT438144 | Appliances > Air Conditioning > Space Heaters > Wood Heaters | 燃气取暖器 | 燃木取暖器 |

### 6.2 最终 300 条独立样本

- 抽样：从本轮 10,381 条可抽候选（排除此前 544 个已检查/纠错 ID）按 27 个一级行业分层，固定 seed `20261128`，抽样前冻结
- 构成：已有译文 90 条、未译文 210 条
- 复核：**285/300（95.0%）**；95% Wilson CI **91.9%–96.9%**
- 分层：已有译文 85/90（94.4%）、未译文 200/210（95.2%）
- 明显语义错误 **15 条**；未发现跨一级行业误分类

| ID | 路径 | 译文 | 问题 | 建议 |
| --- | --- | --- | --- | --- |
| CBT437354 | Vehicle Accessories > Performance > Engine > Distribution | 分配 | 应为配气系统 | 配气系统 |
| CBT456699 | Vehicle Accessories > Heavy Line Replacement Parts > HVAC > Air Conditioners > AC Receivers Dryers | 空调接收器 烘干机 | 应为储液干燥器 | 空调储液干燥器 |
| CBT458782 | Vehicle Accessories > Car & Truck Replacement Parts > Engine > Pumps > Primer Pumps | 底漆泵 | 应为手油泵 | 手油泵 |
| CBT456993 | Vehicle Accessories > Heavy Line Replacement Parts > HVAC > Commands | 指令 | 应为控制面板 | 空调控制面板 |
| CBT456960 | Vehicle Accessories > Heavy Line Replacement Parts > Cabin Parts > Switches > Of Windshield Wiper | 挡风玻璃雨刷 | 缺“开关” | 雨刮开关 |
| CBT445771 | Vehicle Accessories > Scheduled Service > Motor > Fuel Bombs Change | 燃油弹更换 | 应为燃油泵更换 | 燃油泵更换 |
| CBT75214 | Vehicle Accessories > Car & Truck Replacement Parts > Interior Parts > Lighters | 点火器 | 应为点烟器 | 点烟器 |
| CBT457293 | Vehicle Accessories > Heavy Line Replacement Parts > Steering & Suspension > CV Axles and Joints > Tripod Joints | 三脚架接头 | 应为三球销万向节 | 三球销万向节 |
| CBT436113 | Home, Furniture and Garden > Home Care & Laundry > Laundry Accessories > Buckets, Washbowls & Tubs | 桶、洗脸盆和浴缸 | Tubs 应为洗衣盆 | 桶、洗脸盆和洗衣盆 |
| CBT114885 | Sports and Fitness > Handball > Balls | 球类 | 应为手球 | 手球 |
| CBT1334 | Sports and Fitness > Voleyball > Balls | 球类 | 应为排球 | 排球 |
| CBT1345 | Sports and Fitness > Golf > Balls | 球类 | 应为高尔夫球 | 高尔夫球 |
| CBT412728 | Computers > Printing > Spare Parts > For Toners > Toner Drums | 墨盒 | 应为感光鼓 | 感光鼓 |
| CBT3560 | Computers > Printing > Printing Supplies > Toner Cartridges | 墨盒 | 应为碳粉盒 | 碳粉盒 |
| CBT455066 | Appliances > Cooking > Spare Parts & Accessories > For Ranges & Ovens > Resistances | 电阻 | 应为加热管 | 加热管 |

两轮样本合计 476/500（95.2%），但这是两个独立样本的合计，不作为总体保证。

## 7. 保护项

- `active.json`、`staging.json` MD5 均未变（`72d39773…`），中文仍 **0/10,954 已发布**
- `progress.jsonl` 未变（`a897078e…`）
- `overrides.json` **80 条**
- 候选 10,854 条独立暂存，**未发布**
- 普通翻译额度 `{"month":"2026-11","chars":192758}` 未重置、未下调
- 大模型计数 `{"chars":847341}`，未扩 90 万上限
- 未启用每日增量；未恢复普通全量翻译；未重建桌面客户端
- 无残留 build 进程、无 pending

## 8. 待核对

- 控制台大模型实际用量：上一轮 13,691 与本地 13,794 相差 103（54001 未计费）。本轮含 15 次 `30000` 重试与 3 次特殊重试，控制台可能低于本地；本地保留 **847,341**，不下调，请以控制台读数确认。
- 发布前须将两轮共 **24 条**错误写入 overrides（80→104）或等价修正，并重新抽 300 条复验；本轮未获该授权。

## 9. 证据文件

- 质量对照表：[TR-02E-QUALITY-300-20261128.csv](TR-02E-QUALITY-300-20261128.csv)
- 抽样 ID 与 TSV：`.workbuddy/tmp/tr02e-final-300-ids.json`、`.workbuddy/tmp/tr02e-final-300.tsv`
- 构建回执：`/var/lib/una-seller/cbt-category/llm-build-receipts.jsonl`
- 延期清单：`/var/lib/una-seller/cbt-category/llm-build-deferred.json`
- 批次输出：`/var/lib/una-seller/cbt-review/loop/batch-*.json`
- pending 归档：`/var/lib/una-seller/cbt-review/loop/pending-archive/`、`/var/backups/una-seller/tr02e-pending-*`
- 覆盖备份：`/var/backups/una-seller/overrides-20261127T150540Z-pre-80.json`
