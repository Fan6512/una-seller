# TR-02 生产交付与质量复核报告（2026-11-27）

## 1. 结论

| 项目 | 结果 |
| --- | --- |
| 固定提交与联合 CI | PASS。`f644d353a3b565f5729646efca9385d43526a7c0`，#125 `a8579e98` + #126 `f644d353`，9 个后端文件，无新增迁移；[CI 45999249269](https://github.com/Fan6512/una-seller/actions/runs/45999249269) `success` |
| 生产后端升级 | PASS。SHA-256 `b09d6f30edb9387ffea19e5cc2b1aa9c47d3d9e4faf9beb131194b16e91425a7`，`vcs.revision=f644d353`，`vcs.modified=false`，仅重启 `una-seller-api`，`NRestarts=0`，健康检查 200、匿名 401 |
| 百度月额度 | PASS。`BAIDU_TRANSLATE_MONTHLY_CHAR_LIMIT=900000`，保留原计数 3,118；API 与 CLI 共用 `/var/lib/una-seller/cbt-category/baidu-usage.json` |
| 旧译复核 | PASS。96 条进度记录（94 条译文 + 2 条沿用原名），70 条保留，24 条按官方完整路径修正 |
| 五批真实试译 | PASS（接口与额度）。5/5 批、0 失败、新增 80 条、3,958 字符；10 条明确错误 + 8 条语境偏差，已追加 18 条全路径覆盖 |
| 200 条分行业质量抽检 | **FAIL**。明确语义错误 22 条，原始准确率 **89.0%**（178/200，95% CI 约 84.7%–93.3%），低于 95% 目标 |
| 全量翻译 | **已安全暂停**。进度 10,954 条中 3,779 条（3,733 条有效译文 + 46 条沿用原名，34.08%），`active.json` 未发布 |
| 每日增量任务 | **未启用**。质量门禁未通过，不创建 timer |
| 其他约束 | 未消耗百度大模型 500,000 字符测试额度；未重新构建或发布桌面客户端；未重复启动全量进程 |

**结论：基础设施、额度与断点续跑交付通过；质量门禁未通过。全量翻译与词典发布保持暂停，等待质量策略决定。**

## 2. 固定提交与代码范围

- 固定提交：`f644d353a3b565f5729646efca9385d43526a7c0`。
- #125：`a8579e98`，4 个文件（跨进程额度锁、逐次请求扣额、个人 Key 批量回归）。
- #126：`f644d353`，5 个文件（真实父级语境、保守术语、全路径覆盖、`-plan` 零调用预算）。
- 合计 9 个后端文件；`2fc79081..f644d353` 的 `backend/migrations/` 差异为空。
- CI `45999249269` 在固定提交上 `conclusion=success`。

## 3. 生产部署

| 项目 | 结果 |
| --- | --- |
| 构建 | Docker `golang:1.26`，`go1.26.11`，`-trimpath`，源码 `/tmp/una-v070-src` 检出固定提交；`vcs.revision=f644d353`，`vcs.modified=false` |
| 二进制 SHA-256 | `b09d6f30edb9387ffea19e5cc2b1aa9c47d3d9e4faf9beb131194b16e91425a7` |
| 备份 | `/var/backups/una-seller/api-20261127T124310Z-pre-f644d353/`（二进制、`api.env`、unit、词典目录） |
| 回滚副本 | `/opt/una-seller/una-seller-api.rollback-20261127T124310Z-f644d353` |
| 重启范围 | 仅 `una-seller-api`；API `MainPID=1837357`（额度重启后），`NRestarts=0`，`ActiveEnterTimestamp=2026-11-27 20:43:21 CST` |
| 健康 | 本机 `/healthz` 200，公网 `/healthz` 200，匿名 `/api/v1/collections` 401；无 panic/error 日志 |
| 迁移 | 无新增 SQL 迁移 |

## 4. 百度月额度与共用文件

- `api.env` 已设置 `BAIDU_TRANSLATE_MONTHLY_CHAR_LIMIT=900000` 并重启生效。
- 修改前计数：`{"month":"2026-11","chars":3118}`，未清零。
- `CBT_CATEGORY_DICTIONARY_DIR=/var/lib/una-seller/cbt-category`；`BAIDU_TRANSLATE_USAGE_FILE` 未设置。
- CLI 全部通过 `systemd-run --uid=una-seller --gid=una-seller --property=EnvironmentFile=/etc/una-seller/api.env` 运行，与 API 使用同一服务账号、同一环境文件、同一额度文件。
- 暂停时计数：`{"month":"2026-11","chars":192758}`。
- 剩余计划 418,436 字符；若续跑，预计总额度约 611,194 字符，低于 900,000，余量约 288,806。

## 5. 旧译复核（94 条）

进度日志共 96 条记录：94 条译文 + 2 条沿用原名（`Xbox 360`、`Sanders`）。

- 保留 70 条正确译文。
- 按官方完整路径修正 24 条：`Turntables`→唱片机、`Fans`→风扇、`Space Heaters`→取暖器、`Hair Straightener Irons`→直发器、`Mixers`→厨师机、`For Blenders`→搅拌机配件、`Dryers`→干衣机、`Glues & Cutters`→胶水与切割工具、`Cutters, Knives & Blades`→切割器、刀具与刀片、`Geometry`→几何用具、`Refill Pages`→替换内页、`Writing`→书写用品、`Developing & Laboratory`→显影与实验室、`Extension Tubes`→近摄接圈、`Tablet Digitizers`→数位板、`Boomboxes`→便携式音响、`Cassette Players`→磁带播放机、`Motors`→电机、`Drawing Bords and Tables`→绘图板与绘图桌、`Electric`→电动款、`Sanders`→砂光机、`Signs`→标牌、`Pens`→笔、`Mini Drills`→微型电钻。
- 全部写入 `overrides.json`，路径哈希与官方 Dump 完全匹配；`-plan` 输入数减少量可反证 24 条路径全部命中。

## 6. 预算与五批真实试译

| 阶段 | 输入 | 批次 | 源字符 |
| --- | ---: | ---: | ---: |
| 旧译修正后、五批前 `-plan` | 10,854 | 718 | 606,157 |
| 五批后 `-plan` | 10,774 | 713 | 602,199 |

五批结果：

```json
{"official":11668,"disabled":714,"removed":0,"translated":176,"unchanged":3,
 "failed":0,"unprocessed":10775,"skipped":0,"batches":5,"published":false,"state":"paused"}
```

- 5/5 批成功，`failed=0`，新增 80 条。
- 额度 3,118 → 7,076，实际增量 3,958 字符。
- 质量：62 条可用；10 条明确错误、8 条语境偏差，18 条已追加全路径覆盖。
- 新样本原始可用率 77.5%，未达到 95% 目标。

## 7. 200 条分行业质量抽检

### 7.1 方法

- 从暂停时未参与人工校正的 3,691 条译文中抽取。
- 按 27 个一级行业分层，按存量比例分配 200 条，每个行业至少 3 条。
- 随机种子 `20261127`，仅抽取 `zh != name` 的译文，排除全部 64 条 `overrides.json` 路径。
- 判据：译文与官方完整路径的电商语义明显不符、未翻译或严重误导，计为明确语义错误；可读但不符合行业术语的偏差不计入明确错误。

### 7.2 结果

- 正确 178 条，明确语义错误 22 条。
- 原始准确率 **89.0%**（95% CI 约 84.7%–93.3%）。
- 结论：**不能认证 95% 质量目标**；按同一分布外推，未抽检的 3,491 条仍可能有约 11% 的同类错误。

### 7.3 明确错误与修正

| 官方完整路径 | 原译 | 修正 |
| --- | --- | --- |
| Beauty and Personal Care > Beauty Treatments > Lip Plumpers > Lip Plumping Devices | 唇彩器 | 丰唇器 |
| Cameras and Accessories > Camera Accessories > Clapperboards | 挡板 | 场记板 |
| Clothing and Accessories > Fashion Accessories > Hair Accessories > Ties | 领带 | 发绳 |
| Clothing and Accessories > Footwear > Moccasins & Oxford Shoes | Moccasins和牛津鞋 | 莫卡辛鞋与牛津鞋 |
| Computers > PC Components > PC Cases & Mounts > CPU Mounts | CPU安装 | CPU散热器扣具 |
| Computers > Projector & Projection Screen > Overhead Projectors | 高射投影仪 | 高架投影仪 |
| Consoles and Video Games > Console Accessories > For Nintendo > GameCube > Coolers | 酷炫 | 散热器 |
| Consoles and Video Games > Pinballs & Arcade > Accessories & Spare Parts > Coin Pushers | 投币器 | 推币机 |
| Construction > Bathrooms & Restrooms > Toilets | 卫生间 | 马桶 |
| Health and Health Supplies > Health Care > Pharmacy > Pregnancy & Ovulation Tests | 妊娠和排卵试验 | 验孕与排卵试纸 |
| Home, Furniture and Garden > Gardens & Outdoors > Inflatable Loungers | 充气式百叶窗 | 充气躺椅 |
| Home, Furniture and Garden > Home Care & Laundry > Laundry Accessories > Clothes Pegs | 衣架 | 晾衣夹 |
| Home, Furniture and Garden > Home Decoration > Pennants & Garlands | Pennants&Garlands | 三角旗与花环 |
| Home, Furniture and Garden > Kitchen & Housewares > Cooking & Baking > Stovetop Sandwich Makers | Stovetop三明治机 | 炉灶用三明治机 |
| Industries and Offices > Gastronomy & Accommodation > Food Processing > Chocolate Tempering Machines | 巧克力回火机 | 巧克力调温机 |
| Sports and Fitness > Aerobics & Fitness > Dumbbells, Weights & Discs | 哑铃、举重和圆盘 | 哑铃、杠铃片与铃片 |
| Sports and Fitness > Cycling > Accessories > Transportation & Storage > Panniers & Rack Trunks | 盘车和货架箱 | 驮包与货架箱 |
| Sports and Fitness > Martial Arts & Boxing > Defense & Protection > Pads & Focus | 垫子与焦点 | 护具与手靶 |
| Tools > Hand Tools > Cutting & Roughing > Axes | 轴 | 斧头 |
| Tools > Hand Tools > Masonry > Plumb Bobs | 水管 | 铅锤 |
| Tools > Hand Tools > Vises | 视觉 | 台虎钳 |
| Tools > Tool Accessories > Tips & Adapters > Burrs | 毛刺 | 旋转锉 |

22 条已按官方完整路径写入 `overrides.json`，总数由 42 增至 64。修正后，本次被抽中的 200 条明确错误为 0；但这是对被抽中路径的定向修正，不等于 3,691 条总体达到 95%。

## 8. 全量任务暂停状态

| 项目 | 结果 |
| --- | --- |
| systemd unit | `tr02-full-20261127T124523Z.service` |
| 启动 / 停止 | 2026-11-27 20:45:23 CST 启动；20:50:26 CST 发送 SIGTERM 后优雅退出 |
| 最终状态 | `ActiveState=inactive`，`SubState=dead`，`Result=success`，`ExecMainStatus=0` |
| 进度 | `progress.jsonl` 243 条记录；3,779/10,954 条（3,733 条有效译文 + 46 条沿用原名），覆盖率 34.08% |
| 剩余预算 | 7,172 输入 / 483 批 / 418,436 字符（`snapshot_source=staging.json`，batch size 16） |
| 发布 | **未发布**。`active.json` 仍为 17:28 版本，SHA-256 `be52b174c59f5ee41dfa4b702b3b638198a7176d8f30dcf7aa6130d4d37fc959` |
| 额度 | `{"month":"2026-11","chars":192758}`，未重置 |
| 锁 | `update.lock` 为空文件，无运行中更新进程 |

SIGTERM 路径未输出最终 JSON 摘要；上述状态由 systemd 状态、进度日志、额度文件和 `-plan` 复核得出。

## 9. 热加载、客户端读取与原文回退验证

- 生产 API 未为验证而重启：`MainPID=1837357`，`NRestarts=0`，`ActiveEnterTimestamp=20:43:21 CST`；`overrides.json` 于 `20:52:12 CST` 更新，晚于进程启动。
- `cbtcategory.Reader.Load` 在每次读取时比较 `active.json` 与 `overrides.json` 的 inode、大小和 mtime；本次 64 条覆盖无需重启即可在下次请求生效。
- 定向测试（Docker `golang:1.26`，固定提交源码）PASS：
  - `TestDictionaryResumeIncrementalUpdatesAndManualPaths`：人工覆盖优先、提供方不覆盖人工译名、增量续跑。
  - `TestReaderBackoffAndChangedFileRecovery`：文件变更后热加载、坏文件保留上一有效版本并可恢复。
  - `TestDictionaryFallbackForFullWidthAndLongAncestors`、`TestSpecialSeparatorFallsBackToLeaf`：无法安全上下文时回退叶子/原文。
  - `TestCategoryAttributesEndpointAccessAndRules`：HTTP 处理器在无提供方时读取离线词典并回退官方英文。
- `active.json` 未发布新译文，未翻译路径继续回退官方英文。
- 限制：本会话没有可用的生产用户 Bearer Token，未重新执行已登录桌面客户端的真实 GUI 读取；客户端读取仅以同一二进制的 HTTP 处理器集成测试验证。未为验证重启 API。

## 10. 未执行与阻塞

1. **未续跑全量**：质量门禁 89.0% 低于 95%，继续消耗额度只会扩大未审核译文。
2. **未发布词典**：`active.json` 未变，生产主要显示英文；仅 64 条人工复核覆盖会在读取时生效。
3. **未启用每日增量**：未创建 `una-seller-cbt-dictionary.timer` 或 cron。
4. **未消耗百度大模型额度**：500,000 字符测试额度未接入、未调用。
5. **未重新构建/发布桌面客户端**。
6. **未对账百度控制台**：`baidu-usage.json` 为服务端计数，控制台账单仍需人工核对。

## 11. 建议

- **方案 A（质量优先）**：在续跑前完成 3,691 条已译内容的全量人工复核或引入可离线执行的领域术语表，把明确错误率降到 5% 以下；本轮 200 条方法可复用，但需覆盖全量而非抽样。
- **方案 B（接受辅助定位）**：负责人明确接受中文为辅助显示、约 89% 原始准确率，批准续跑剩余 7,172 条；发布后再以抽检和 overrides 逐步纠错。该方案不满足原定 95% 门槛，需书面豁免。
- 无论哪种方案，都不建议在未决策前启用每日增量或发布 `active.json`。

## 12. 证据

- 生产：`ubuntu@152.136.41.105`，`/var/lib/una-seller/cbt-category/`，备份 `/var/backups/una-seller/api-20261127T124310Z-pre-f644d353/`。
- 词典：`active.json` SHA-256 `be52b174…`；`overrides.json` SHA-256 `4964771c8d936fd7d8ee62e61cbaba5899249d51420ffa3e87f743220a74f1d2`。
- 进度：`progress.jsonl` 243 条；额度 `{"month":"2026-11","chars":192758}`。
- 抽检样本：本地 `.workbuddy/tmp/tr02-audit-200.tsv`（不入库），分层与随机种子见第 7.1 节。
- 关联：[#125](https://github.com/Fan6512/una-seller/pull/125)、[#126](https://github.com/Fan6512/una-seller/pull/126)、[CI 45999249269](https://github.com/Fan6512/una-seller/actions/runs/45999249269)。
