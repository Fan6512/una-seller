# TR-02F · Managed CBT 类目词典受保护增量维护

日期：2026-11-28  
状态：设计基线，等待实现 PR。  
前置：TR-02E 已正式发布，当前生产 `active.json` SHA-256 前缀 `b3b9f46d`，`llm-release-managed.json` 已存在；普通 `cbt-categories update` 已被联网前阻断。  
范围：只处理官方 CBT dump 新增/变化的类目，不重建全量词典，不修改商品/刊登数据，不创建或启用定时器。

## 1. 目标与不可破坏的生产边界

实现一条独立于旧 updater 的受保护增量流水线：

`plan → build → assemble → review/approve → publish → rollback`

- 只以 Mercado Libre 官方 CBT 完整 dump 为分类身份与结构来源。
- 对新增或完整路径变化的启用类目生成中文候选；未变化路径沿用已发布译文。
- 只写独立增量候选、清单、回执和审批材料；build/assemble 绝不写 `active.json`。
- 发布必须显式 dry-run / `-apply`，基于 active SHA、官方 dump 版本、候选 SHA 和逐条人工审批。
- 发布后保留 managed 标记；回滚后也不得解除旧 updater 保护。
- 第一阶段不创建 systemd timer、cron、后台 Worker 或自动发布。定时任务只有在人工流程经过实际运行观察并另行批准后才讨论。

**绝对禁止：**
- 不调用或放开旧 `cbt-categories update`；不删除/绕过 `llm-release-managed.json`。
- 不复用 `staging.json`、`progress.jsonl`、`llm-review-candidates.json`、`llm-build-pending.json` 等全量/旧流程状态文件。
- 不直接让普通百度机器翻译 Provider 接管词典；增量翻译使用现有独立 Baidu LLM provider 与累计额度文件。
- 不在没有审批的情况下写 active；不清零/回退任何额度；不自动重试未知结果。
- 不改数据库、不新增 SQL migration、不重建桌面客户端、不调用 Mercado Libre 写接口。

## 2. 当前实现事实

- `update.go` 的旧 Update 会下载官方 dump、用 `seed` 尽量保留相同 ID/路径译文、将进度写入 `staging.json/progress.jsonl`，并在完成时直接切换 `active.json`。这是 managed release 后被禁用的路径，不能复用为发布实现。
- `review_build.go` 的全量 LLM build 只针对当前本地 active/staging 版本，使用 `llm-review-run.lock → update.lock → baidu-llm-usage.json.lock` 的锁顺序、逐条 pending/receipt 和 900,000 字符累计硬上限。
- `review_release.go` 的全量发布要求原版本 300 条冻结审查；`validateReleaseIdentity` 和 `switchReleaseLocked` 明确禁止改变官方版本/分类身份，因此不能直接用于官方 dump 版本升级。
- `Reader.Load` 通过 active 文件 inode/大小/mtime 检测热加载；原子替换 active 后可由 Reader 自动加载，不要求重启 API。
- `overrides.json` 以完整官方路径 hash 匹配；`glossary.go` 为固定术语。两者优先级高于 LLM 候选。
- 当前生产 LLM 额度使用独立累计文件 `baidu-llm-usage.json`，硬上限 900,000；每次计划/执行必须读取现场最新计数，历史报告数值不能当作当前余额。

## 3. 新命令与文件隔离

建议 CLI 入口：

```sh
una-seller-api cbt-categories incremental plan
una-seller-api cbt-categories incremental build -max-rows 1..200 -budget N
una-seller-api cbt-categories incremental assemble
una-seller-api cbt-categories incremental publish -active-sha256 HASH -approval FILE -approval-sha256 HASH
una-seller-api cbt-categories incremental publish ... -apply
una-seller-api cbt-categories incremental rollback -active-sha256 HASH -backup-sha256 HASH
una-seller-api cbt-categories incremental rollback ... -apply
```

命令必须是独立子命令，不能 fall-through 到旧 update/review/trial 分支。参数不合法、managed 标记缺失/损坏、active 缺失/损坏、锁被占用、额度文件缺失/损坏时全部 fail-closed。

增量状态使用独立文件名（具体实现可精简，但不得与旧流程文件共用）：

- `llm-incremental-source.json`：通过官方 MD5/Created 校验后解析的只读官方 snapshot。
- `llm-incremental-manifest.json`：active SHA、官方 version、source SHA、增删改分类 ID、待翻译 ID、来源字符数和预算预检。
- `llm-incremental-candidates.json`：按 ID 保存的 LLM 候选与 source SHA。
- `llm-incremental-deferred.json`：安全回退英文的 ID/source SHA。
- `llm-incremental-pending.json` / `llm-incremental-receipts.jsonl`：逐条请求意图与 fsync 回执。
- `llm-incremental-candidate.json`：完整候选 snapshot。
- `llm-incremental-review.json`：本轮所有新增/变化启用类目的逐条审查清单。
- `llm-incremental-approval.json`：人工审批材料；审批文件由操作者确认后提供，不自动伪造 reviewer。

文件需限制大小、校验 JSON/UTF-8、原子写入；pending/receipt 尾部损坏时停止，归档核查后才可继续。

## 4. 增量比较规则

以锁定读取的当前 active 为基线，以本次官方 dump 为目标：

| 类别 | 处理 |
| --- | --- |
| 新 ID | 作为新增；若 enabled 且无人工覆盖/术语，则需要翻译 |
| 相同 ID、完整 `PathNode(ID, Name)` 链完全一致 | 保留 active 的 `Zh`，不重复消耗额度 |
| 相同 ID，但叶子名、祖先 ID 或祖先名称导致完整路径变化 | 视为变化；不得沿用旧 `Zh`，重新按新完整路径翻译 |
| 仅 status / listing_allowed / has_children 等官方元数据变化，完整路径不变 | 更新官方元数据，保留原译文 |
| enabled 状态新开启且当前无译文 | 纳入新增待翻译 |
| disabled 类目 | 不请求翻译；官方状态和刊登权限必须原样保留 |
| active 中已不存在的 ID | 追加到 `Retired`，保留旧路径与译文；不作为新翻译请求 |
| 完全相同的官方版本（MD5 + Created） | no-op，不请求翻译、不发布 |

补充约束：
- 官方 dump 的 gzip MD5、Created、完整 JSON/父子路径、ID、状态和 listing_allowed 必须通过现有 `fetchDump/readDump/parseDump` 校验。
- 官方版本早于 active 必须拒绝；同一 Created 但 MD5 不同应作为异常停止并人工调查，不猜测版本先后。
- `overrides.json` 完整路径覆盖优先；`applyCategoryGlossary` 固定术语其次；只有剩余新增/变化 enabled 类目才调用 LLM。
- 对未变化路径，保留 active 中已发布的译文或空值；不要顺手重试历史 26 条英文回退，避免每日重复消耗。历史延期重试应另设显式命令/批准。
- 输出 manifest 必须列明 added / changed_path / metadata_changed / removed / unchanged / newly_enabled / translation_required 数量及 ID，不能只报总数。

## 5. 锁、额度和请求纪律

锁顺序必须与现有全量 LLM build / publish 一致：

1. `llm-review-run.lock`
2. `update.lock`
3. 发送请求时由 provider 获取 `baidu-llm-usage.json.lock`

要求：
- 增量入口必须确认 `llm-release-managed.json` 是普通文件且 JSON 明确为 `managed:true`；缺失或损坏都拒绝执行。绝不删除该标记。
- 同时检查全量 `llm-build-pending.json` 与增量 pending；任一未解决即停止，不自动重发。
- 使用现有 `llmReviewProvider`、`BAIDU_LLM_APP_ID`、`BAIDU_LLM_API_KEY` 和共享累计额度；不使用普通机器翻译的月额度文件。
- 每次 build 先估算**全部剩余增量**完整路径字符数，再读取累计用量并确认 `usage + remaining <= budget <= 900000`。不满足则在第一条请求前停止。
- 每条请求前持久化 pending；请求后保存候选或 deferred，追加并 fsync receipt，最后才删除 pending。额度只增不减；未知结果必须人工对账并归档。
- 每条一次请求、沿用现有 1 秒间隔和 15 秒超时，不自动重试；只允许每次最多 200 行，重复运行只处理尚未完成行。
- plan 只取官方 dump 并估算，不调用翻译 API；build 是唯一会消耗 LLM 额度的命令，需单独授权。

## 6. 候选组装与逐条审批

assemble 从冻结 source + active 基线 + manifest + candidates + deferred + overrides + glossary 构造完整 snapshot：

- 官方 version、category 数量、所有 ID、Name、ParentID、Depth、HasChildren、ListingAllowed、Status、完整 Path 必须与冻结官方 source 完全一致。
- Retired 必须等于 active 历史 Retired 加本轮真实移除项，不得静默丢历史。
- 未变化路径的 `Zh` 必须与 active 一致；新/变化路径不得继承旧路径译文。
- manual override > glossary > 新 LLM candidate > 官方英文（空 `Zh`）；不使用普通未复核译文兜底。
- 所有新/变化且需要翻译的 enabled 行都进入 review 文件，包含 ID、完整路径、原文、source SHA、候选中文、来源（manual/glossary/LLM/deferred）。
- 第一阶段要求**本轮全部新增/变化译文逐条人工确认**，不以小样本 95% 代替逐条审核；无译文回退官方英文必须明确显示并单独确认。
- 审批文件绑定 active SHA、官方 version/MD5、source SHA、manifest SHA、candidate SHA、review SHA、reviewer，以及每个 delta ID 的 final label / accepted 判定；缺项、重复、过期、哈希不一致均拒绝。

## 7. 发布与回滚

publish 必须在 `llm-review-run.lock → update.lock` 下重新读取全部证据：

- active SHA 必须仍等于审批基线；source、manifest、candidate、review、approval 哈希全部匹配。
- 再次校验候选与官方 source 的完整身份一致性、未变化路径译文保留、变化路径译文来源及逐条审批。
- 默认 dry-run：只返回 before/after SHA、版本、delta 计数和 backup 名称，不写 active、backup 或 managed 标记。
- `-apply` 才能执行：先按当前 active SHA 创建内容寻址备份，再用同目录临时文件 + fsync + rename + 目录 fsync 原子替换 active。已有 managed 标记必须保持存在且内容不被删除。
- 发布失败不得删除 managed 标记；失败后停止并留证据。
- rollback 需要 expected active SHA + backup SHA + 显式 `-apply`；验证备份 SHA 与 snapshot 可解析，先备份当前 active，再原子切回。rollback 不解除 managed 标记，不调用旧 updater。
- active 替换后通过 Reader 热加载；不得为词典切换无必要地重启 API。生产部署另行批准。

## 8. 第一阶段验收（保持简单、定点）

最小 Go 测试：

1. managed 缺失/损坏时增量命令 fail-closed；旧 update 仍在联网前拒绝。
2. 相同版本 no-op；新 ID 进入待翻译；完整路径改变重新翻译；未变路径保留原译文；仅元数据变化不重新翻译；移除项进入 Retired。
3. overrides / glossary 优先，disabled 不请求翻译；历史空译文不在每日增量自动重试。
4. 完整剩余额度预检在首个请求前阻断；共享累计计数只增不减；pending、receipt 中断后阻断重发。
5. build/assemble 不修改 active、staging、progress、overrides 或普通额度。
6. active/source/candidate/approval 任一哈希变化均阻止 publish；dry-run 零写入；apply 原子切换并留备份；rollback 恢复原 bytes 且 managed 标记仍在。
7. Reader 能热加载新 active；官方 ID / path / status / listing_allowed 全部与 source 完全一致。

只跑本包定向测试与现有 Go Backend CI；不引入大规模测试矩阵、不新增迁移、不启用 timer。

## 9. 实施与上线分阶段

- **A · 实现 PR**：新增独立 incremental 子命令、数据结构、候选/审批/发布/回滚和上述定点测试。保持 Draft 直至 CI 与代码复审通过。
- **B · 生产只读部署**：合并后仅升级 API 二进制；无 migration、无桌面构建、无 timer。核对旧 update 仍阻断、managed 标记未变、健康 200 / 匿名 401；只运行 incremental plan，检查官方版本、delta 计数与额度估算，不请求翻译、不写 active。
- **C · 首轮真实增量**：先把 plan 输出、实际 LLM 累计余额、预估字符与预计成本交负责人确认；获批后 build。逐条审查全部 delta，dry-run 复核后再单独批准 apply。发布后核对 active SHA、Reader 热加载、官方身份和备份。
- **D · 定时任务（本期不做）**：至少完成若干次人工增量、回滚演练和额度对账后，再单独立项。第一版定时任务最多自动 plan/build 生成候选并报告，不自动 apply；timer 需独立审批与开关。

**交付完成定义**：实现 PR 合并并通过 CI；生产只读 plan 通过；首轮 build/review/dry-run/apply 有完整哈希和额度收据；旧 updater 持续阻断；回滚演练 PASS。以上完成前 TR-02 不标记整体完成。
