# TR-02 公共 CBT 类目词典生产接入与五批试译验收（2026-11-27）

## 1. 结论

- 生产后端已升级到 V0.8.0 固定候选 `6a56734f`，公共词典目录已接入并在线，官方分类 Dump 已发布。
- **五批真实试译 BLOCKED（未通过）**：火山通用机器翻译返回业务错误 `-403 account status abnormal: unsynchronized`，仅发出 1/5 批，成功 0 批、译文 0 条、中文覆盖率 0/10,954。
- 未切换其他供应商，未执行全量翻译，未配置定时增量任务；本次结果不阻塞 V0.8.0 签名发布。

> 后续（2026-11-27）：已改用百度 Provider 完成五批真实试译并切换默认 Provider，火山账号问题保留备用。见 [TR-02B 验收](TR-02B-BAIDU-ACCEPTANCE-20261127.md)。

## 2. 生产环境与配置

| 项目 | 结果 |
| --- | --- |
| 主机 / 服务 | `152.136.41.105`，`una-seller-api`（用户 `una-seller`） |
| 升级前 | `1f697fc8`，二进制 SHA-256 `41ca9a60…d99c6`，PID 1602721 |
| 升级后 | `6a56734f`（V0.8.0 固定候选），SHA-256 `5d11e2d9…f42284`，PID 1650639，NRestarts=0 |
| 构建 | 洁净检出 + Docker `golang:1.26`，`vcs.modified=false`，`vcs.time=2026-11-27T09:12:08Z`，go1.26.11 |
| 备份 | `/var/backups/una-seller/api-20261127T092726Z-pre-6a56734f/`（二进制、api.env、unit）；`/opt/una-seller/una-seller-api.rollback-20261127T092726Z-6a56734f` |
| 健康 | 本机与公网 `/healthz` 200，匿名 `/api/v1/stores` 401，启动后无 panic/fatal，Nginx/systemd 未改 |
| 翻译配置 | `TRANSLATION_PROVIDER=volcengine`；AK/SK 已配置，未输出明文；签名校验通过（见 §4），说明 AK/SK 对可用且签名实现正确 |
| 词典配置 | 新增 `CBT_CATEGORY_DICTIONARY_DIR=/var/lib/una-seller/cbt-category`；目录 0700，属主 `una-seller` |
| 升级差异 | `1f697fc8..6a56734f` 后端 11 文件，无新增 SQL 迁移 |

## 3. 官方分类与当前覆盖率

| 指标 | 值 |
| --- | --- |
| 官方 Dump | `created=2026-10-14T21:12:35Z`，`md5=f5acbe4aca4abfe99819ab22c2e7956f` |
| 官方类目 | 11,668 |
| 停用 / 已移除 | 714 / 0 |
| 有效可译 | 10,954 |
| 已译 / 沿用原名 / 未处理 / 跳过 | 0 / 0 / 10,954 / 0 |
| 中文覆盖率 | **0.00%（0/10,954）** |
| active.json | SHA-256 `be52b174…37fc959`，4,022,647 B |
| API 启动日志 | `state=loaded_untranslated translated=0 version=2026-10-14T21:12:35.000Z` |
| 定时任务 | 未配置（`una-seller` 与 root crontab、systemd timer 均无 cbt-categories） |

全量规模（按现有 16 条/批与 1001 字符上限计算）：10,953 个唯一输入、492,473 字符、710 批；单轮全量低于月 180 万字符额度。

首次以 `TRANSLATION_PROVIDER=disabled` 拉取并发布官方版本，返回：
`{"official":11668,"disabled":714,"removed":0,"translated":0,"unprocessed":10954,"published":true,"state":"provider_disabled"}`。

## 4. 五批试译结果

执行命令（生产服务账号，未输出凭据）：

```sh
una-seller-api cbt-categories update -max-batches 5
```

返回：

```json
{"version":{"md5":"f5acbe4aca4abfe99819ab22c2e7956f","created":"2026-10-14T21:12:35.000Z"},
 "official":11668,"disabled":714,"removed":0,"translated":0,"unchanged":0,
 "failed":16,"unprocessed":10938,"skipped":0,"batches":1,"published":false,"state":"paused"}
```

| 指标 | 结果 |
| --- | --- |
| 计划批次 / 实际发出 | 5 / 1（首个 Provider 级失败后按设计停止） |
| Provider 请求总数 | 3（1 个 16 条批次 + 2 个单条诊断请求） |
| 成功请求 / 成功率 | 0 / **0%** |
| 翻译字符数 | **0**（失败批次实际输入 16 条、558 字符，未产生成功计费） |
| 失败类目 | 16 条，含 `Electronics, Audio and Video`、`Lava Lamps`、`TVs`、`VCRs`、`Xbox 360` 等 |
| 译文样本 | 无（未取得任何译文） |
| 发布 | 否；`active.json` 未改写，`staging.json` 与 `active.json` 字节一致 |

Provider 返回（HTTP 200，业务错误）：`Code=-403`，`Message=account status abnormal: unsynchronized`。两次单条诊断请求结果一致，排除瞬时抖动。签名被服务端接受，故不是 AK/SK 格式或签名算法问题，而是**火山账号状态异常/未同步**；需负责人在火山控制台核实实名认证、机器翻译服务开通及账户状态后再重试。应用侧按既有策略进入 120 秒冷却并停止，未自动切换供应商。

## 5. 持久化、续跑、去重与读取

- **持久化 PASS**：`progress.jsonl` 已落盘 1 条记录（版本、16 个 failed ID），`staging.json`、`active.json` 保留且未发布；服务重启后重新加载词典成功。
- **断点续跑 / 重复翻译保护：代码级 PASS，运行时多批 NOT_RUN**（Provider 阻塞）。`cbtcategory` 全量单测通过（0.970s），覆盖：同版本不调用 Provider、续跑不重复已完成行、增量只译新增名称、人工覆盖不消耗额度、未发布结果不泄漏、失败项单独重试、单写者锁与损坏文件回退。
- **客户端读取**：API 以同一 `Reader` 加载词典，启动日志为 `loaded_untranslated`；桌面端读取 `path_zh` 且仅在长度匹配时展示，缺译回退官方英文。因当前无中文译文，含中文的实机读取 **NOT_RUN**。
- 注意：16 条失败项按现有排序位于队尾，完整跑完前不会重试；部分批次运行不会发布，符合现有“完成才发布”语义。

## 6. 待开发登记：公共词典月额度保护

- 目标：火山通用翻译每月 180 万字符持久化额度，达到阈值即暂停并保留进度，禁止自动切换供应商。
- 最小实现：目录内 `usage.json` 记录 `{month, chars}`；每次成功请求按输入字符累加并 `fsync`；跨月重置；达到 1,800,000 时返回 `state=quota_protected` 并停止；不引入令牌桶、分布式计数或新依赖。
- 状态：已登记为待开发，建议在全量翻译前完成。

## 7. 后续

1. 负责人处理火山账号状态（实名/服务开通/账户同步），无需更换 AK/SK。
2. 重跑 `cbt-categories update -max-batches 5`，取得首批真实译文并核对控制台计费与免费额度。
3. 五批验收通过后再执行全量，并配置每日增量（同一服务账号、同一目录）。
4. 全量前落地 §6 额度保护；翻译验收与 V0.8.0 签名发布解耦。
