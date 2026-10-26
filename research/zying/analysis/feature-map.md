# 功能地图：菜单功能 → 窗体类 → ZClient 命令 → 字段

生成方式：菜单项 key 与窗体类声明的 `Key` 常量精确关联，命令与字段取自 `command-inventory.csv`、`field-index.json`。

说明：客户端没有本地关系库，列表数据来自服务器返回的 RecordSet；因此下文的"数据对象"指服务器返回的记录集名称，而非本地表名。


## 美客多（Meli）

### 参数设置

| 功能 | 窗体类 | 命令 | 请求字段 | 返回对象 |
|---|---|---|---|---|
| Meli平台分类 | Meli分类列表 | `meli_category.refer_cbts`<br>`meli_category.refer_total`<br>`meli_category.select` | `allowvar`, `ids`, `isleaf`, `key`, `list`, `name`, `root`, `site`, `update_at`, `zh` | `list`, `root` |
| Meli运费 | Meli运费列表 | `meli_freight.insert`<br>`meli_freight.select`<br>`meli_freight.ship` | `abbr`, `list` | `list` |
| Meli分类佣金 | Meli类别佣金列表 | `meli_rate.select` | `list`, `sf_site` | `list` |
| 产品分类对照 | Meli分类对照 | `（无命令站点，纯本地/网页页）` | — | — |
| 分类国家映射 | Meli分类映射国家列表 | `meli_refer.batch`<br>`meli_refer.delete`<br>`meli_refer.select` | `abbr`, `cate`, `cateid`, `list`, `local`, `refer` | `list` |
| QA回复模板 | MeliQA回复模板 | `meli_qatemp.delete`<br>`meli_qatemp.select`<br>`meli_qatemp.stat` | `id`, `list` | `list` |

### 基本设置

| 功能 | 窗体类 | 命令 | 请求字段 | 返回对象 |
|---|---|---|---|---|
| 账户授权 | Meli授权列表 | `meli_acc.areas`<br>`meli_acc.delete`<br>`meli_acc.flag`<br>`meli_acc.flags`<br>`meli_acc.select`<br>`meli_acc.stat` | `abbr`, `acc`, `admins`, `appinfo`, `areas`, `color`, `flag`, `flags`, `id`, `list` … | `areas`, `flags`, `list` |
| 店铺绩效 | Meli绩效列表 | `meli_metrics.areas`<br>`meli_metrics.fetch`<br>`meli_metrics.select` | `acc`, `areas`, `canceled_rate`, `claims_rate`, `color`, `delayed_rate`, `list`, `quota`, `shopid`, `trade_total` … | `areas`, `list` |
| 价格调整 | Meli价格调整 | `（无命令站点，纯本地/网页页）` | — | — |
| Meli编码池 | Meli编码池 | `meli_code.clear`<br>`meli_code.delete`<br>`meli_code.select`<br>`meli_code.stat` | `code`, `list`, `stat` | — |
| 编码生成 | Meli编码 | `meli_code.insert` | `root` | — |

### 美客多本土

| 功能 | 窗体类 | 命令 | 请求字段 | 返回对象 |
|---|---|---|---|---|
| 已上传明细 | Meli本土上传明细列表 | `meli.localupdate`<br>`meli_localstat.batchstatus`<br>`meli_localstat.clean`<br>`meli_localstat.delete`<br>`meli_localstat.predict`<br>`meli_localstat.select`<br>`meli_localstat.setstat`<br>`meli_localstat.stat` | `abbr`, `acc`, `aid`, `cate`, `health`, `itemid`, `list`, `local`, `orders`, `ptime` … | `abbr`, `list`, `stat` |
| 操作日志 | MeliLocalLogList | `meli_locallog.delete`<br>`meli_locallog.select` | `acc`, `feed`, `flag`, `ierr`, `list`, `login`, `time` | `list` |
| 站内信QA | Meli本土Qa | `meli_qa.delete`<br>`meli_qa.pull`<br>`meli_qa.select`<br>`meli_qa.stat` | `acc`, `aid`, `as_text`, `as_time`, `create_at`, `id`, `ids`, `list`, `saleid`, `stat` … | `list`, `stats` |

### 美客多跨境

| 功能 | 窗体类 | 命令 | 请求字段 | 返回对象 |
|---|---|---|---|---|
| 已上传明细 | Meli上传明细列表V2 | `meli.update`<br>`meli_stat.delete`<br>`meli_stat.selectv2`<br>`meli_stat.setstat`<br>`meli_stat.statv2` | `abbr`, `acc`, `aid`, `catalog_status`, `cate`, `cbtid`, `group`, `health`, `isfill`, `itemid` … | `abbr`, `list`, `stat` |
| 产品同步 | （未定位到类） | `（无命令站点，纯本地/网页页）` | — | — |
| 操作日志 | MeliLogListV2 | `meli_log.delete_v2`<br>`meli_log.select_v2` | `acc`, `done`, `fail`, `feed`, `list`, `stat`, `time`, `type` | `list` |
| 售前QA | Meli站内信 | `meli_question.delete`<br>`meli_question.pull`<br>`meli_question.select`<br>`meli_question.stat` | `acc`, `aid`, `antime`, `id`, `ids`, `list`, `retime`, `saleid`, `stat`, `stats` | `list`, `stats` |
| 售后消息/索赔 | MeliClaimAndMessage | `（无命令站点，纯本地/网页页）` | — | — |
| 海外仓库存 | （未定位到类） | `（无命令站点，纯本地/网页页）` | — | — |

### 跨境UP

| 功能 | 窗体类 | 命令 | 请求字段 | 返回对象 |
|---|---|---|---|---|
| 已传明细 | Meli_UP上传列表 | `meli_stat_up.delete`<br>`meli_stat_up.select`<br>`meli_stat_up.setstat`<br>`meli_stat_up.stat`<br>`meli_up.update` | `acc`, `aid`, `catalog_status`, `cate`, `cbtid`, `group`, `health`, `itemid`, `list`, `local` … | `list`, `sites`, `stat` |
| 产品同步 | （未定位到类） | `（无命令站点，纯本地/网页页）` | — | — |
| 操作日志 | MeliUP日志列表 | `meli_log_up.delete`<br>`meli_log_up.select` | `acc`, `done`, `fail`, `feed`, `list`, `stat`, `time`, `type` | `list` |
| 图片上传 | Meli图片上传 | `（无命令站点，纯本地/网页页）` | `buypage`, `id`, `list`, `meliurl`, `msg`, `pic`, `pics` | — |
| 跟卖监控 | Meli跟卖监控列表 | `meli_catalog_monitor.remove`<br>`meli_catalog_monitor.select`<br>`meli_catalog_monitor.stat` | `acc`, `catalog_status`, `create`, `itemid`, `list`, `ptime`, `rank`, `stat`, `upid`, `winnerid` | `list`, `stat` |

## 财务

（本板块窗体尚未还原，仅登记功能项名称）：智能生成、运单费核对、充值审核、软件余额、软件费用、软件充值、月租费、店铺返点、代缴记录

## 订单与采购

（本板块窗体尚未还原，仅登记功能项名称）：订单列表、订单关联列表、头程单、订单全局搜索、退货服务表、统计图表、库存列表、套装设置、批量采购、1688采购、我的采购单、我的运单、国外退货单、转运监控、速卖通交接单、Shein发货单、转运服务表、创建无主件、认领无主件、下单器1688

## 产品与采集

（本板块窗体尚未还原，仅登记功能项名称）：供应商、调整分销价格、已代理的产品、未代理产品、分销出单、选品方法、选品数据库、已选库、搜索词排名、选品大数据、1688货源、单个采集、批量采集、产品批量导入、产品批量替换、产品图片、产品分类、分类映射、变体对照、侵权词替换、防侵权、标题模板、1688核价

## 库存与仓储

（本板块窗体尚未还原，仅登记功能项名称）：已代理的产品、未代理产品、分销出单、供应商、调整分销价格、套装设置、库存预购、国内退回单

## 其它平台板块（仅规模，未展开）

| 板块 | 功能项 | 有命令站点的功能项 |
|---|---|---|
| 亚马逊 | 30 | 1 |
| 选品 | 23 | 0 |
| 系统设置 | 19 | 0 |
| Lazada | 12 | 1 |
| Walmart | 12 | 0 |
| 广告 | 9 | 0 |
| Shopee | 9 | 1 |
| Hepsi | 8 | 0 |
| Coupang | 6 | 0 |
| eBay | 6 | 0 |
| Linio | 6 | 1 |
| Temu | 6 | 0 |
| Shopify | 5 | 0 |
| B2W | 4 | 1 |
| Wish | 4 | 1 |
| 速卖通 | 3 | 0 |
| Allegro | 3 | 0 |
| N11 | 3 | 0 |
| 服务 | 3 | 0 |
| TikTok | 3 | 0 |
| 拍拍档/海外仓 | 2 | 1 |
| Shein | 2 | 1 |
| BOL | 1 | 0 |
| WeoBuy | 1 | 0 |

## 覆盖范围

- 全平台菜单共 29 个板块、265 个功能项；本文件只展开美客多/产品/订单/库存/财务五个板块。
- **只有美客多板块的窗体类完成还原**（菜单 key 能解析到类的 22/25 个，其中 18 个有命令站点）；产品、订单、库存、财务等板块的窗体在受 .NET Reactor 保护的模块里，尚未还原，所以对应表格里的"窗体类/命令"列为空——这是取证覆盖限制，不代表那些功能没有实现。
- 展开表里的"请求字段"是同一函数窗口内出现的键名集合，可能混入读响应时用到的列名。

## 读法

1. 功能项与命令是**多对多**：同一命令常跨跨境/UP/本土三套页面复用，差别只在参数（如 `aid` 的口径）。
2. "无命令站点"的功能并不等于没实现：它们要么是纯本地表格编辑，要么是内嵌网页/服务器侧渲染页，取证重心应转向 `api-map.md` 的直连接口。
3. `※按账号灰度` 来自菜单构建里的账号判断，说明智赢按账号 ID 白名单开放新功能——自研对应做法是功能开关放在服务端下发，而不是编译进包。
