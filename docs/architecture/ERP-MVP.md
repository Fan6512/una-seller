# ERP 后端与 Windows / macOS 桌面客户端 MVP 开发任务书

> 项目：Una Seller — Mercado Libre AI ERP  
> 版本：v0.4（2026-11-15，采集箱/产品库/产品列表与轻量插件化规划）  
> 状态：需求基线，待实施  
> 客户端：Windows + macOS 双平台桌面应用  
> 关联规范：[Foundation](FOUNDATION.md)、[Delivery](DELIVERY.md)、[Domain and Execution](DOMAIN_AND_EXECUTION.md)、[Cross Platform](CROSS_PLATFORM.md)、[Browser Collection Plugin](../product/BROWSER-COLLECTION-PLUGIN-MVP.md)

## 1. 产品定位

Una Seller 是面向自有 Mercado Libre 跨境店铺的运营 ERP。首期先解决竞品商品采集与商品资料管理，不做大型传统 ERP，不做纯网页卖家工作台。

系统由三部分组成：
1. Windows / macOS 桌面客户端：用户日常管理界面。
2. Chrome / Edge 浏览器插件：在 Amazon 和 Mercado Libre 页面上人工选择并采集商品。
3. 一个 ERP Backend API：统一身份、业务规则、数据保存和后续平台连接。

Amazon 只作为商品研究和采集来源，**首期采集仅适配 Amazon.com 美国站**，未来也是主要来源；其他 Amazon 站点不进入本 MVP 的开发和验收范围。未来店铺授权、刊登、订单和客服运营只面向 Mercado Libre。

首期自用，单用户使用；数据库保留未来增加团队成员的基本扩展空间，但不建设商业 SaaS、多租户或完整团队权限系统。

**整体产品北极星**：采集 → AI 整理 → Mercado Libre 刊登 → 日常运营。当前 MVP 是第一段“采集箱可用”，后续具体产品库、AI、图片、刊登与已发布产品列表需求见 [产品工作流与智赢编辑窗口参考](../product/PRODUCT-WORKFLOW-AND-EDITOR.md)。本文件的第一版不做清单仅限制**当前 M0 采集闭环**，不代表否定已规划的后续功能。

## 2. 已确定的产品决策

| 项目 | 决策 |
| --- | --- |
| 桌面客户端 | Windows + macOS |
| 桌面框架 | Tauri 2 |
| 前端 | React + TypeScript strict |
| 构建工具 | Vite |
| UI | Tailwind + 少量 shadcn/ui / Base UI |
| 后端 | Go 单一 HTTP 服务，M0-03 复用 backend/cmd/api；遇到明确阻塞再评估 |
| 数据库 | 有实际持久化需求的业务数据使用 PostgreSQL |
| 本地缓存 | 只有真实离线需求出现时再引入 SQLite |
| 浏览器插件 | Chrome / Edge Manifest V3；Amazon.com 美国站优先，Mercado Libre MLM/MLB/MLA；独立适配器 |
| 开发部署 | 先本地开发测试，后续部署云服务器 |
| 用户体系 | 首期单用户；未来按真实团队需求扩展 |
| 功能组织 | 一个 Core 主题/主程序 + 构建时集成的内置业务插件；单后端与单数据库 |
| 商业化 | 当前不做支付、订阅、计费或第三方 SaaS |

本任务书遵循仓库现行 Foundation v2.1、Delivery v2.1 和 ADR-0004。不得用 Wails、Electron、NestJS 或 Ant Design 替换已经确定的 Tauri 2 + React 路线；如未来要更换技术栈，必须先提供实测依据并单独修订 ADR。

当前 M0-03 明确复用现有 Go 服务作为**唯一业务后端**。这是基于已有工程与最小开发成本的取舍，**不是声称 Go/Fastify 已做完性能对比**；出现可证实的阻塞后再评估更换（见 ADR-0001 更新）。

## 3. 跨平台支持范围

### Windows

- Windows 10 / 11 x64 为首要开发平台。
- 使用 Tauri 2 + Windows WebView2。
- 提供开发启动和可安装构建验证。

### macOS

- Apple Silicon arm64。
- Intel x64。
- 使用 Tauri 2 + WKWebView。
- 必须分别记录构建结果；有可用 Mac 设备时完成实际启动冒烟验证。

### 跨平台原则

- React / TypeScript 业务界面和 API 数据协议共用。
- 不为 Windows 和 macOS 分别编写业务客户端。
- Rust 只处理必要的系统能力，不把业务逻辑写入 Rust。
- 文件路径、系统凭据、快捷键、打印、更新等操作通过 Tauri 能力或少量系统适配处理。
- Windows 与 macOS 的用户数据目录分别使用系统推荐路径，不写死盘符或用户主目录。
- 浏览器插件与桌面操作系统无关，在两个系统的 Chrome / Edge 中使用同一套插件代码。

正式发行前再处理 Windows 安装包签名、macOS 签名和公证；不得把尚未完成的签名/公证描述成 MVP 已通过。

## 4. 总体架构

### 桌面端

Tauri 2 + React + TypeScript + Vite，**一个统一主程序/主题、按业务模块开发内置插件**：
- Core：导航、主题、会话、API client、加载/错误/空态、简单模块注册；
- 采集箱：来源商品列表和详情；
- 后续按阶段启用产品库、AI 文案/图片、店铺刊登/发布记录、已刊登产品列表；
- 浏览器 MV3 采集扩展独立于桌面业务插件，仍只提交到统一 Backend API。

不引入运行时插件商店、单插件独立服务/数据库、任意代码装载和复杂权限平台。

现阶段复用既有 UI，不为未来插件创建空页面或空路由。

现有 Core 负责登录/连接状态、工作台及设置/插件配对；**采集箱**模块提供商品列表和内部详情页。后续业务模块分阶段注册，不重复建设另一套桌面菜单。

桌面端不直接连接 PostgreSQL，不保存 Mercado Libre OAuth Token，也不实现独立业务规则。

### 后端

一个 HTTP 业务服务：
- 身份认证。
- 插件连接和授权。
- 商品采集数据接收。
- 商品数据校验、去重和查询。
- 图片文件管理。
- 后续 Mercado Libre 官方 REST API 集成的预留边界。

先采用简单模块化单体，不拆微服务、不建额外网关。

### 数据库与文件

- PostgreSQL 保存持久化业务数据。
- 商品图片保存为文件，不把二进制塞进数据库 JSON。
- 首期保留来源图片 URL；本地缓存图片只在必要且来源允许时进行。
- 本地开发可用 Docker Compose 启动 PostgreSQL。
- 不把 PostgreSQL 打包进桌面客户端安装程序。

### 数据流

浏览器商品页面 → 浏览器插件 → ERP Backend API → PostgreSQL / 文件存储 → Windows 或 macOS 桌面客户端。

桌面客户端与浏览器插件共用同一套后端 API 和商品数据契约。

## 5. M0-03 后端实现决定

**使用 Go 作为当前唯一 ERP Backend**，复用 `backend/cmd/api` HTTP 服务及现有 `/healthz`，加入最小 PostgreSQL 连接、迁移、采集 POST/GET。当前阶段不再要求 TypeScript/Fastify 并行验证，未进行的性能对比不得宣称完成。

只有当 Go 路线出现具体工程阻塞或实测成本问题时，再通过单独决策评估更换；不同时维护两套业务后端。

只保留一个 HTTP 框架（或标准库方案）、一种数据库访问方式、一种迁移工具和必要 DTO；不叠加 ORM、Repository、CQRS、DDD、Event Sourcing 等多层抽象。

## 6. 数据库与用户模型

### 单用户首期

- 首次启动建立一个管理员账户。
- 支持登录、退出和会话失效。
- 商品记录保留创建者/所有者关联，避免将业务数据硬编码为不可区分的全局单例。
- 不开发邀请成员、团队角色、复杂 RBAC 或商业多租户。

未来增加团队成员时，再按真实需求增加成员关系和权限；不提前建立完整组织系统。

### 第一版核心数据表

| 表 | 用途 |
| --- | --- |
| users | 初始管理员账户及后续用户身份 |
| source_products | Amazon / Mercado Libre 采集商品 |
| source_product_images | 商品图片及顺序 |
| source_product_variants | 多规格变体及其属性 |

不提前创建订单、退款、广告、采购、财务、团队权限等空表。

### source_products

核心字段：
- id：ERP 内部 UUID。
- created_by_user_id：创建/采集用户。
- source_platform：amazon / mercadolibre。
- source_site：Amazon 站点或 MLM / MLB / MLA。
- source_url：商品原始链接。
- normalized_url：规范化链接，用于去重。
- title：来源商品标题。
- price：当前来源价格，使用精确数值类型。
- currency：来源货币。
- category：来源分类。
- description：详细描述。
- bullet_points：原始五点描述，按数组保存。
- collection_status：完整、部分等已保存数据的状态；提交失败不伪装为成功入库。
- collection_warnings：本次采集的缺失字段和简短诊断信息（仅有必要时保存）。
- first_collected_at：首次采集时间。
- last_collected_at：最近采集时间。

不建立销量、评价、卖家信息、品牌或独立商品 ID / ASIN 字段。

### source_product_images

- source_product_id。
- image_url。
- image_order。
- image_type。
- variant_id（适用时）。
- local_path（确实缓存本地副本时）。

### source_product_variants

- source_product_id。
- attributes：规格属性 JSON。
- price：变体价格（页面提供时）。
- currency。
- 对应图片关联。

金额不使用浮点数进行最终存储或计算。来源页面没有提供的字段允许为空，不补造规格组合。

### 去重

唯一依据：
用户/所有者 + 来源平台 + 来源站点 + normalized_url。

重复采集不新增重复记录；更新最近采集时间和本次确实获取到的有效资料。不得用一次不完整采集覆盖已有完整数据。

## 7. ERP Backend API

统一 REST JSON API，首期只实现采集和桌面端必要接口。路径以 /api/v1 为前缀。

建议接口：
- GET /healthz：服务健康检查。
- POST /api/v1/auth/login：用户登录。
- POST /api/v1/auth/logout：退出。
- GET /api/v1/auth/me：当前用户。
- POST /api/v1/extension/pairing-codes：生成短时一次性插件配对码。
- POST /api/v1/extension/pair：插件绑定。
- GET /api/v1/extension/status：插件连接状态。
- POST /api/v1/collections：提交商品采集资料。
- GET /api/v1/collections：分页查询采集商品。
- GET /api/v1/collections/{id}：商品详情。
- DELETE /api/v1/collections/{id}：删除采集记录。

采集接口：
- 仅接收插件本地解析出的**结构化商品数据**，不以整页 HTML 上传/服务端解析为默认架构。
- 校验平台、站点、URL、货币、普通描述、独立五点描述、实际变体及图片关联。
- 使用去重唯一键和 PostgreSQL 事务实现一次提交的幂等新增/更新，不另建设第三方同款多阶段商业采集接口。
- 结构化请求可附带 `page_state`、`missing_fields`、`warnings`，帮助识别 Amazon.com 的正常、验证码/阻断、邮编地区限制、不可配送、缺 BuyBox、未知页面和解析失败等情况。
- **状态明确区分**：新建成功、已有更新、部分采集已保存、请求失败未保存。失败响应不能被插件当成“采集成功”。
- 本次来源页面未返回的字段不允许无条件覆盖已有完整字段；商品描述与五点描述保持独立，变体属性/价格/图片关系正确保存。
- 不把浏览器插件提交的字段当成可信输入；字段不完整应返回缺失字段与简短错误信息。
- 插件只在收到后端持久化成功的响应后更新卡片按钮状态。

插件配对码短时有效且只能使用一次。插件凭据不暴露给网页页面脚本。公网部署后必须使用 HTTPS 和有效的用户鉴权。

接口 DTO 是插件与桌面端的共同契约；接口变更需同步更新两端，不允许各自维护不兼容的数据格式。

### 与插件的五项实施变更保持同步

第一版自研插件以 Amazon.com 美国站为主要开发样本，Mercado Libre 继续覆盖 MLM、MLB、MLA。ERP Backend API 不负责访问 Amazon 或 Mercado Libre 的竞品商品网页。

1. **卡片识别与详情采集分阶段**：在浏览器内完成，服务器只接收用户点击后解析的结果，不做无人值守采集。
2. **多规格补全**：插件按需、有限地补充已确认的变体详情；服务器保存真实的属性组合、对应价格、图片及缺失标记。
3. **独立描述字段**：`description` 与 `bullet_points` 分开；A+ 图文首版不要求单独采集。
4. **页面状态与采集结果状态分开**：上游页面可标识 normal/captcha/blocked/zipRequired/deliveryUnavailable/buyboxMissing/unknown；ERP 只把实际保存成功的 complete/partial 记录列为已采集，未保存失败以错误响应处理。
5. **一次幂等入库**：只用 `POST /api/v1/collections` 传结构化字段；事务内去重与有条件更新，不拆成远端商业平台式的重复检查和二次入库。

临时用于商品详情定位的 ASIN 不必单独持久化为字段；其在规范化 URL 中出现属于 URL 的正常组成部分。平台、站点和 URL 去重规则不变。

## 8. Windows / macOS 桌面客户端页面

### 8.1 导航

**当前 M0 菜单**：
1. 工作台。
2. 采集箱（商品详情在该模块内部进入，不必占一个主菜单）。
3. 设置。

**后续 M1 菜单逐步扩展**：采集箱、产品库、产品列表；店铺/刊登记录和 AI 文案/图片可作为这些模块中的功能入口，不必每个插件都占一级导航。

商品库草稿/编辑、AI 文案/图片、店铺授权/模板、多站点刊登和已发布商品列表是后续分阶段功能；订单/客服等更晚。首批只做真实采集箱，不在 M0 堆满空页面。

### 8.2 工作台

只展示必要信息：
- 采集商品总数。
- 今日采集数量。
- Amazon 商品数量。
- Mercado Libre 商品数量。
- 最近采集记录。
- 浏览器插件连接状态。
- 后端连接状态。

不开发复杂数据大屏。

### 8.3 采集箱（原“商品采集库”）

支持：
- 商品缩略图。
- 商品标题。
- 来源平台。
- 来源国家/站点。
- 当前价格及货币。
- 采集时间。
- 已保存商品的完整/部分采集状态；未保存成功的失败只显示采集错误通知，不作为采集箱商品记录。
- 标题关键词搜索。
- 来源平台、站点和采集时间筛选。
- 分页查询。

### 8.4 商品详情

分区显示：
- 基础信息：标题、原始链接、价格、货币、分类、来源。
- 图片：主图、其他图片和顺序。
- 描述：详细描述、五点描述。
- 规格：全部可获取变体、属性、价格和对应图片。
- 采集记录：首次采集时间、最近采集时间、完整/部分采集状态及缺失字段提示。
- 普通商品描述与五点卖点分别展示；没有来源卖点时留空，不自动拆分。
- 变体价格或图片缺失时显示为空及原因，不补造。

首期不加入 AI 改写、翻译、自动刊登或竞品分析。

### 8.5 设置

- 当前 ERP API 地址。
- 本地/云端连接状态。
- 浏览器插件配对码。
- 插件连接状态。
- 当前用户。
- 客户端版本。
- 基本诊断信息。

## 9. 本地开发与云端部署

### 本地开发

- PostgreSQL 通过 Docker Compose 启动。
- 后端在本机运行，开发阶段只监听本地地址。
- Tauri 桌面客户端连接本地后端。
- Chrome / Edge 插件连接本地后端。
- 数据库和图片目录使用持久化卷/目录。
- 本地开发数据不写入仓库，不提交凭据。

本地开发不等于把数据库和后端嵌入 Windows/macOS 安装包。

### 后续云端部署

- 一个后端业务服务部署到 Linux 云服务器。
- PostgreSQL 使用持久化存储。
- API 配置 HTTPS。
- 桌面客户端通过设置或环境配置切换 API 地址。
- 浏览器插件绑定同一个云端 API。
- Windows 和 macOS 登录后读取同一份云端已保存数据。

首期不承诺复杂离线同步。未同步的本地草稿不承诺自动跨设备恢复。

云端公网接口不能因“自用”而取消鉴权；平台 OAuth Token 只保存在后端。

## 10. 当前 M0 采集闭环不开发（M1 已有独立路线）

**下列限制仅用于当前 M0-03 / Issue #6 的首批代码 PR**。用户已确认 M1 分阶段需要产品库、AI 文案、AI 白底、店铺/站点模板、多站点上传、错误回读与已发布产品列表；不能把“首批不做”理解为“产品永远不做”。具体里程碑以 [Delivery v2.1](DELIVERY.md) 为准。

- AI 商品分析、AI 选品、AI 翻译和改写。
- Mercado Libre 商品自动刊登。
- Mercado Libre OAuth 店铺授权和多店铺操作。
- 订单与物流管理。
- 售前问答和售后消息。
- 库存、采购、财务、广告、促销。
- Amazon 店铺授权或 Amazon 卖家运营。
- 团队邀请、复杂角色权限、商业 SaaS、多租户。
- 支付、订阅、计费和商业后台。
- Redis、消息总线、微服务和复杂离线同步。

以上功能只按需进入后续里程碑，不为“未来可能需要”提前建设。

## 11. Issue #6 的开发 PR 划分（唯一当前执行顺序）

- **PR-A：后端采集入库**。复用 `backend/cmd/api`，完成 PostgreSQL 迁移、`source_products`/图片/实际变体保存、`POST /api/v1/collections` 幂等写入、列表/详情 GET；合成数据真实入库、读回、重启后可见。
- **PR-B：Tauri 跨平台采集箱**。Core + 采集箱列表/内部详情，真实 API 显示图片/普通描述/独立卖点/变体和缺失原因；Windows 原生启动验收，macOS 构建与实机状态按事实记录。
- **PR-C：Amazon.com MV3 浏览器采集**。用户点击 → 结构化采集 → Go API → 桌面采集箱可见；重复、部分和错误状态准确。以后另开范围扩展 Mercado Libre 的 PR。

M1 产品库/AI/店铺/刊登均由 [Delivery v2.1](DELIVERY.md) 逐阶段授权，**不得混入 M0-03**。每个 PR 保持小范围、Draft 审核后再决定合并。

## 12. MVP 验收标准

### 后端
- 选定的后端服务可本地启动。
- PostgreSQL 迁移可执行。
- 登录和基础鉴权正常。
- 插件配对正常。
- 商品采集 API 可接收和校验数据。
- URL 去重正确。
- 图片、描述、五点描述和变体完整保存。
- API 返回结构统一。

### Windows
- Tauri 2 程序能构建、启动。
- 能连接本地后端。
- 能登录并显示工作台。
- 能查看采集库、搜索和筛选。
- 能查看商品图片、描述、五点描述和规格变体。

### macOS
- arm64 与 x64 构建结果分别记录。
- 在可用 Mac 设备上完成启动冒烟验证。
- 能连接相同 API 并读取同一份已保存数据。
- 不以 Windows 构建通过代替 macOS 验收。

### 集成
- Amazon.com 美国站商品通过插件进入采集库。
- Mercado Libre MLM / MLB / MLA 商品通过插件进入采集库。
- 重复采集不生成重复商品。
- 插件显示正确状态（包括完整、部分采集和上游拦截/失败）。
- Amazon.com 多规格变体的属性、价格、对应图片能正确入库；详细描述与 Bullet Points 独立保留。
- 无法访问详情或变体数据时，不上传整页 HTML，不补造数据，也不把 API 失败误报为成功。
- 后端重启后数据仍然存在。
- Windows 与 macOS 客户端登录同一后端后能读取同一份云端数据（云端阶段验收）。

## 13. 后续商品编辑器与发布架构边界

### 来源、草稿与刊登不可混为一个状态

- 采集箱 SourceProduct：保留第三方原始事实、来源 URL、图片、卖点、实际变体；不采集销量、评价、卖家、独立 ASIN 或独立品牌列。
- 产品库 OwnProduct（M1-A 才落表）：人工补充采购/包装、SKU/库存、产品内容，AI 生成可编辑候选；来源记录不可被修改为自有商品。
- 发布目标 PublicationTarget 和发布操作（M1-E 才落表）：每个店铺×站点目标独立记录结果/错误，但**API 请求可以按官方账号模型合并多个目标站点提交**，不得预设一站点一请求；官方 UP/Item 关系按真实模型映射。
- 产品列表 PublishedListing：只展示实际成功发布并经官方回读的商品；失败和进行中位于关联刊登记录。

### 编辑页借鉴（截图功能，不复制布局）

四张智赢截图显示单窗口包含图片、类目属性、包装/成本、规格/SKU、语言标题/卖点/描述、操作历史等；Una Seller 采用**独立分区/标签**，不复制长表单。人工补充的重量、包装、采购资料不应加入浏览器插件强制采集合同；AI 不补造这些事实。

### 图片与多站点

- AI 主图白底：独立内置图片插件，原图与候选分别保存，人工确认选用。官方 AI 白底生成开放 API **尚未核实**，不得将官方上传/诊断接口冒充编辑接口。
- 账号模板：保留默认允许站点与按站点销售参数；一次勾选多个有权限的站点，由后端分别校验/发布/追踪。
- 服务端对每个目标记录 pending / processing / success / failed / needs_check；可用时按官方返回的各站点结果（如 site_items[]）和只读回查判定，整体请求成功不等于全部站点成功。失败原因需字段级可读，成功目标不能因其他目标失败而重复发布。

### 轻量插件机制

Core 的统一主题、导航、会话和 API 层共享；功能插件构建时注册页面和业务模块，后端仍是一个服务，数据库仍是一个 PostgreSQL。仅随着真实功能出现添加对应插件代码，不预先安装运行时插件系统。

完整可操作要求：[PRODUCT-WORKFLOW-AND-EDITOR.md](../product/PRODUCT-WORKFLOW-AND-EDITOR.md)。

## 14. 当前实施边界

本任务书第一阶段只定义 ERP 后端与桌面采集箱 MVP，不代表全部业务功能已授权开发。

当前优先完成：
1. 后端技术选型验证。
2. 本地 PostgreSQL 与采集 API。
3. Tauri 2 Windows/macOS 桌面骨架。
4. 与独立浏览器插件的商品采集闭环。

所有新增依赖、架构抽象和基础设施必须说明实际必要性。优先少代码、少服务、可运行、可维护。
