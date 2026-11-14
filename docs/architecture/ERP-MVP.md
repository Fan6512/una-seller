# ERP 后端与 Windows / macOS 桌面客户端 MVP 开发任务书

> 项目：Una Seller — Mercado Libre AI ERP  
> 版本：v0.2  
> 状态：需求基线，待实施  
> 客户端：Windows + macOS 双平台桌面应用  
> 关联规范：[Foundation](FOUNDATION.md)、[Delivery](DELIVERY.md)、[Domain and Execution](DOMAIN_AND_EXECUTION.md)、[Cross Platform](CROSS_PLATFORM.md)、[Browser Collection Plugin](../product/BROWSER-COLLECTION-PLUGIN-MVP.md)

## 1. 产品定位

Una Seller 是面向自有 Mercado Libre 跨境店铺的运营 ERP。首期先解决竞品商品采集与商品资料管理，不做大型传统 ERP，不做纯网页卖家工作台。

系统由三部分组成：
1. Windows / macOS 桌面客户端：用户日常管理界面。
2. Chrome / Edge 浏览器插件：在 Amazon 和 Mercado Libre 页面上人工选择并采集商品。
3. 一个 ERP Backend API：统一身份、业务规则、数据保存和后续平台连接。

Amazon 只作为商品研究和采集来源。未来店铺授权、刊登、订单和客服运营只面向 Mercado Libre。

首期自用，单用户使用；数据库保留未来增加团队成员的基本扩展空间，但不建设商业 SaaS、多租户或完整团队权限系统。

## 2. 已确定的产品决策

| 项目 | 决策 |
| --- | --- |
| 桌面客户端 | Windows + macOS |
| 桌面框架 | Tauri 2 |
| 前端 | React + TypeScript strict |
| 构建工具 | Vite |
| UI | Tailwind + 少量 shadcn/ui / Base UI |
| 后端 | 一个 HTTP 业务服务；Go 与 TypeScript/Fastify 在 M0 做极小验证后选择 |
| 数据库 | 有实际持久化需求的业务数据使用 PostgreSQL |
| 本地缓存 | 只有真实离线需求出现时再引入 SQLite |
| 浏览器插件 | Chrome / Edge Manifest V3，独立目录和独立适配器 |
| 开发部署 | 先本地开发测试，后续部署云服务器 |
| 用户体系 | 首期单用户；未来按真实团队需求扩展 |
| 商业化 | 当前不做支付、订阅、计费或第三方 SaaS |

本任务书遵循仓库现行 Foundation v2、Delivery v2 和 ADR-0004。不得用 Wails、Electron、NestJS 或 Ant Design 替换已经确定的 Tauri 2 + React 路线；如未来要更换技术栈，必须先提供实测依据并单独修订 ADR。

后端语言尚未最终冻结。仓库现有 Go 本机健康检查服务只是 M0 的最小验证实现，不代表已经完成 Go 与 TypeScript/Fastify 的正式选型。

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

Tauri 2 + React + TypeScript + Vite：
- 登录和连接状态。
- 工作台。
- 商品采集库。
- 商品详情。
- 设置与浏览器插件绑定。

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

## 5. 后端技术验证与选择

M0 先完成一个极小的后端实现验证：
- Go：沿用仓库现有健康检查服务，验证 HTTP 路由、PostgreSQL 连接、一次迁移和一个采集接口。
- TypeScript/Fastify：实现同等范围的最小验证。
- 对比开发改动量、依赖数量、错误处理、构建/部署、后续维护成本。
- 选择一个作为唯一正式后端技术栈；另一份验证代码不进入正式业务路径。

不得因为现有健康检查代码使用 Go 就默认完成技术选型，也不得同时长期维护两套后端。

选定后只保留：
- 一个 HTTP 框架或标准库方案。
- 一种数据库访问方式。
- 一种迁移工具。
- 明确的 API DTO。

不在 M0 叠加 ORM、Repository、CQRS、DDD、Event Sourcing 等多层抽象。

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
- collection_status：完整、部分、失败等采集状态。
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
- 校验平台、站点、URL、货币、变体和图片结构。
- 使用去重键实现幂等。
- 明确返回新增、已存在并更新、部分采集或失败。
- 不把浏览器插件提交的字段当成可信输入。
- 失败信息可供插件和桌面端显示。

插件配对码短时有效且只能使用一次。插件凭据不暴露给网页页面脚本。公网部署后必须使用 HTTPS 和有效的用户鉴权。

接口 DTO 是插件与桌面端的共同契约；接口变更需同步更新两端，不允许各自维护不兼容的数据格式。

## 8. Windows / macOS 桌面客户端页面

### 8.1 导航

首期菜单：
1. 工作台。
2. 商品采集库。
3. 商品详情。
4. 设置。

商品管理、店铺管理、商品刊登、订单管理、客服中心和 AI 运营助手属于后续阶段，不在首期做完整页面。

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

### 8.3 商品采集库

支持：
- 商品缩略图。
- 商品标题。
- 来源平台。
- 来源国家/站点。
- 当前价格及货币。
- 采集时间。
- 完整/部分/失败状态。
- 标题关键词搜索。
- 来源平台、站点和采集时间筛选。
- 分页查询。

### 8.4 商品详情

分区显示：
- 基础信息：标题、原始链接、价格、货币、分类、来源。
- 图片：主图、其他图片和顺序。
- 描述：详细描述、五点描述。
- 规格：全部可获取变体、属性、价格和对应图片。
- 采集记录：首次采集时间、最近采集时间和状态。

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

## 10. 第一版不开发

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

## 11. 开发 PR 划分

### PR-A：后端技术验证与基础骨架

- 对 Go 与 TypeScript/Fastify 做最小对比验证并选定一种。
- 保留一个 HTTP 服务。
- 健康接口。
- PostgreSQL 连接与最小迁移。
- 基础配置和错误响应。
- 本地启动说明。

### PR-B：商品采集数据与 API

- source_products、图片和变体表。
- 商品采集 API。
- URL 规范化与去重。
- 数据完整性和部分采集状态。
- 商品列表、详情和删除 API。
- 最小必要测试。

### PR-C：Tauri 2 跨平台桌面客户端

- Windows/macOS 共用 Tauri 2 + React + TypeScript UI。
- 登录和 API 连接状态。
- 工作台。
- 商品采集库。
- 商品详情。
- 插件配对设置。

### PR-D：插件与 ERP 集成验收

- 浏览器插件接入统一 API。
- Amazon 和 Mercado Libre 实际页面采集。
- 采集数据入库。
- 桌面端显示新商品。
- 重复采集、部分数据和失败重试验证。

每个 PR 保持小范围，先 Draft，CI 和必要检查通过后再 Ready for Review。合并前由负责人审查；不因任务书保存到仓库就自动授权开始全部实现。

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
- Amazon 商品通过插件进入采集库。
- Mercado Libre MLM / MLB / MLA 商品通过插件进入采集库。
- 重复采集不生成重复商品。
- 插件显示正确状态。
- 后端重启后数据仍然存在。
- Windows 与 macOS 客户端登录同一后端后能读取同一份云端数据（云端阶段验收）。

## 13. 当前实施边界

本任务书定义 ERP 后端与桌面客户端 MVP，不代表全部业务功能已授权开发。

当前优先完成：
1. 后端技术选型验证。
2. 本地 PostgreSQL 与采集 API。
3. Tauri 2 Windows/macOS 桌面骨架。
4. 与独立浏览器插件的商品采集闭环。

所有新增依赖、架构抽象和基础设施必须说明实际必要性。优先少代码、少服务、可运行、可维护。
