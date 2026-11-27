# TR-02B 百度翻译接入与五批试译验收（2026-11-27）

## 1. 结论

- **接入验收 PASS**：百度通用文本翻译连通，五批真实试译 5/5 成功，0 失败、0 跳过；缓存命中不重复调用，月额度硬上限生效。
- **质量结论：有条件通过**。80 条类目译文中约 5 条语义明显错误、5 条语境偏差；中文仅作辅助显示、缺译回退英文，不阻塞默认 Provider 切换，但全量翻译前应补父级上下文或人工覆盖。
- 默认 Provider 已切换为 `baidu`；未执行全量翻译、未启用定时任务，不影响 V0.8.0 签名发布。

## 2. 代码与部署

| 项目 | 结果 |
| --- | --- |
| PR | [#123](https://github.com/Fan6512/una-seller/pull/123)，CI 四项 PASS（Go Backend 2m5s） |
| 合并提交 | `2fc79081`（squash） |
| 生产二进制 | SHA-256 `5d22e070…fcdf2`，`vcs.revision=2fc79081`，`vcs.modified=false`，go1.26.11 |
| 部署 | 2026-11-27T09:52:51Z 原子替换并仅重启 `una-seller-api`；PID 1678772，NRestarts=0，本机/公网 `/healthz` 200，无 panic |
| 备份 | `/var/backups/una-seller/api-20261127T095251Z-pre-2fc79081/`（二进制、api.env、unit）；回滚副本 `una-seller-api.rollback-20261127T095251Z-2fc79081` |
| 默认 Provider | `/etc/una-seller/api.env` 已改为 `TRANSLATION_PROVIDER=baidu`，并显式设置 `BAIDU_TRANSLATE_MONTHLY_CHAR_LIMIT=50000` |
| 凭据 | `BAIDU_TRANSLATION_APP_ID` / `BAIDU_TRANSLATION_SECRET_KEY`，仅服务端读取，未输出、未写入 Git/日志/客户端 |

## 3. 实现范围（复用为主）

- 复用统一 `TranslationService`、NFC 缓存、singleflight、离线词典、批次日志与断点续跑；未新增依赖。
- 百度语言映射：`en→en`、`es→spa`、`pt→pt`，目标 `zh`；自动检测仅用于订单标题。
- 批量：离线词典 16 条/批、合计 ≤1000 字符、1 QPS；单请求超时 3 秒，离线批次总预算 30 秒，类目属性 10 秒。
- 有限重试：仅对网络错误、HTTP 5xx、百度 `52001/52002` 重试 1 次；429、鉴权、余额、语言不支持直接冷却/失败，不盲目重试。
- 月额度：`BAIDU_TRANSLATE_MONTHLY_CHAR_LIMIT`（默认 50000，0 表示全部禁止），计数写入 `CBT_CATEGORY_DICTIONARY_DIR/baidu-usage.json`（可用 `BAIDU_TRANSLATE_USAGE_FILE` 覆盖），按北京时间月份重置；**超额在发送前拒绝**，文件损坏或不可写时 fail-closed。
- 公共额度只作用于服务级维护 Key；用户在账号管理绑定的自有 Key 不占用该额度。

## 4. 五批真实试译

执行（生产服务账号，Provider 临时覆盖为 baidu）：

```sh
cbt-categories update -max-batches 5
```

| 指标 | 结果 |
| --- | --- |
| 批次 / 请求 | 5 / 5，全部 HTTP 200 且业务成功 |
| 成功率 | **100%（5/5）**，`failed=0`、`skipped=0` |
| 处理类目 | 80 条（78 条已译 + 2 条沿用原名） |
| 输入字符 | **2,789**（计费口径，百度按源语言字符统计） |
| 输出中文 | 318 字符 |
| 剩余未处理 | 10,874 |
| 发布 | 否（部分批次不发布，`state=paused`） |
| 响应耗时 | 5 批合计约 7 秒（含 1 QPS 间隔） |

返回摘要：

```json
{"official":11668,"disabled":714,"translated":78,"unchanged":2,"failed":0,
 "unprocessed":10874,"skipped":0,"batches":5,"published":false,"state":"paused"}
```

译文样本（原文 → 中文）：

- `Antique Cash Registers` → 古董收银机；`Antique Keys & Locks` → 古董钥匙和锁；`Electric Kettles` → 电水壶；`Toasters` → 烤面包机；`Blenders` → 搅拌机；`Water Heaters` → 热水器；`VCRs` → 录像机；`Lava Lamps` → 熔岩灯。
- 语境正确示例：`Antiques > Antique Scales` → 古董秤；`Kitchen Appliances > Kitchen Scales` → 厨房秤。

质量偏差（人工复核 80 条）：

| 原文 | 百度译文 | 问题 |
| --- | --- | --- |
| Accessories > Motors | 汽车 | 应为电机/马达，语义错误 |
| Audio Antiques > Turntables | 转盘 | 应为唱机 |
| Home Appliances > Fans | 扇子 | 应为风扇 |
| Home Appliances > Space Heaters | 太空加热器 | 应为空间加热器/取暖器 |
| Sanding & Polishing > Sanders | Sanders | 未翻译，应为砂光机 |
| Audio > Boomboxes | Boombox | 未翻译，应为便携式音响 |
| Developing & Laboratory | 开发与实验室 | 相机语境应为显影与实验室 |
| Tablet Digitizers | 平板电脑数字化仪 | 应为数位板/绘图板 |
| For Blenders | 搅拌机 | 丢失“配件”语义 |
| Drawing Bords and Tables | 绘制图纸和表格 | 应为绘图板与绘图桌 |

统计：明确错误 5/80（6.3%），语境偏差 5/80；其余 70 条可用。根因是词典默认只发送叶子名称（仅同名类目带路径），同形词缺少父级上下文。

## 5. 属性名称与属性值

通过同一 Provider 真实调用 12 条典型属性（1 次请求，103 字符）：

| 原文 | 译文 | 评价 |
| --- | --- | --- |
| Brand | 品牌 | 正确 |
| Model | 模型 | 产品语境应为“型号” |
| Color | 颜色 | 正确 |
| Red / Blue | 红 / 蓝色 | 可用 |
| Material | 材料 | 正确 |
| Cotton | 棉 | 可用（棉质更佳） |
| Size | 尺寸 | 正确 |
| Pack of 2 | 2包 | 应为“2件装” |
| Power Source | 电源 | 正确 |
| Battery Powered | 电池供电 | 正确 |
| Recommended Uses | 推荐用途 | 正确 |

属性名称/值整体可用率 10/12；`Model`、`Pack of 2` 建议人工覆盖或后续上下文优化。

## 6. 缓存、续跑与额度

- **缓存命中**：同一 12 条属性连续调用两次，第一次 343ms，第二次 0s；额度仅首次 +103 字符，第二次无新增，证明未重复调用 API。
- **断点续跑**：追加 `-max-batches 1`，仅新增第 6 批 16 条（额度 +226），已译 78→94；6 条进度记录 ID 交集为 0，未重复翻译已有译文。
- **持久化**：`progress.jsonl` 6 条记录、`staging.json` 保留；服务重启后词典仍从 `active.json` 加载（当前未发布中文，`loaded_untranslated`）。
- **额度**：当前 `{"month":"2026-11","chars":3118}`；单轮全量约 492,473 字符，超过默认 50,000，需先在百度控制台确认版本/额度再调高，否则会在约 50,000 字符处停止并保持 `paused`。
- 单元测试：新增额度跨重启、超额不发请求、瞬时 503 重试一次、语言映射；`go test ./internal/httpapi -run 'TestTranslationService|TestBaidu|TestVolcengine|TestOrderTitleTranslations'` PASS。

## 7. 未执行与后续

- 未执行全量翻译、未启用每日增量、未创建 cron/timer。
- 未做客户端含中文的实机读取（当前正式词典仍 0 译文；staged 94 条待全量完成后发布）。
- 全量前建议：为同名/同形词增加父级上下文，或按本次质量清单写入 `overrides.json`；确认百度服务版本后将月上限调整到 ≥500,000。
- 火山 Provider 与 1,800,000 字符额度保护保留原状，未删除、未启用。
