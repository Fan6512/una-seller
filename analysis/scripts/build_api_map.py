"""Generate meli-direct-endpoints.json and api-map.md.

Splits the client's outbound traffic into three channels: direct calls to the
official Mercado Libre REST API, ZClient command forwarding to ZYing servers, and
plain HTTP page/service calls. Tables are machine-extracted from the resolved
sources; the surrounding analysis is mine. No source bodies are emitted.
"""
import csv
import io
import json
import os
import re
import sys
from collections import defaultdict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DEC = os.path.join(ROOT, 'decrypted')
EP_OUT = os.path.join(ROOT, 'meli-direct-endpoints.json')
DOC_OUT = os.path.join(ROOT, 'api-map.md')

OBF = r'(?:(?:[\w$]+\.|[\w$]+\([^)]*\)\s*\.)?\s*(?:/\*\d+\*/)?\s*)'
METHOD_RE = re.compile(r'^\s*(?:public|private|internal|protected|static|async|\s)+'
                       r'(?:[\w<>\[\],.?]+\s+)+([A-Za-z_]\w*)\s*\(([^)]{0,160})\)', re.M)
PATH_RE = re.compile(r'"(/[a-zA-Z_][\w\-/{}\.?=&]{1,60})"')
ABS_RE = re.compile(r'"https://api\.mercadolibre\.com(/[^\s"]{1,80})"')
VERB_RE = re.compile(r'HttpMethod\.([A-Z]\w+)')
CALL_RE = re.compile(r'\b(?:Send(?:Version2)?Async|SendCoreAsync|SendRawAsync)\s*\(([^;]*?)\)\s*;')
DTO_RE = re.compile(r'(?:Deserialize|Json\.Read)<(\w+)>')
SEND_RE = re.compile(r'\b(SendVersion2Async|SendAsync|SendCoreAsync|SendRawAsync)\b')


def split_args(text):
    """Split a call's argument list on top-level commas."""
    out, depth, cur, quote = [], 0, [], None
    for ch in text:
        if quote:
            cur.append(ch)
            if ch == quote:
                quote = None
        elif ch in '"\'':
            quote = ch
            cur.append(ch)
        elif ch in '([{':
            depth += 1
            cur.append(ch)
        elif ch in ')]}':
            depth -= 1
            cur.append(ch)
        elif ch == ',' and depth == 0:
            out.append(''.join(cur))
            cur = []
        else:
            cur.append(ch)
    out.append(''.join(cur))
    return [a for a in out if a.strip()]

# Ordered from most to least specific: classify() returns the first match, so the
# token domain must come after the ones whose paths merely embed an access_token query.
DOMAIN = [
    ('售后与索赔', ['/claims', 'expected-resolutions', 'open-dispute', 'send-message', 'affects-reputation']),
    ('图片与视频', ['/pictures', '/clips']),
    ('全球商品（UP）', ['user-product', '/global', '/locals', 'user-products-families']),
    ('类目与属性', ['/categories', '/attributes', 'domain_discovery']),
    ('商品与刊登', ['/items', '/marketplace_items', 'catalog_listing', 'multiget', '/products', '/description',
                   '/shipping_options', '/visits', 'price_to_win', 'search']),
    ('站内信与问答', ['/messages', 'packs', 'moderations']),
    ('订单', ['/orders']),
    ('绩效与额度', ['performance', '/cap']),
    ('OAuth 与令牌', ['/oauth/token', 'authorization', '/users']),
]


def classify(paths):
    for name, keys in DOMAIN:
        if any(k in p for p in paths for k in keys):
            return name
    return '其它'


def main():
    text = io.open(os.path.join(DEC, 'MeliApi.resolved.cs'), 'rb').read()
    text = text.decode('utf-8', 'replace')
    lines = text.splitlines()
    sigs = []
    for i, line in enumerate(lines):
        # Only a declaration sits at one tab of indentation in ILSpy output; statements are
        # deeper, and `await SendAsync(url, HttpMethod.Put, x)` otherwise reads as a
        # signature and cuts the window before the verb it carries.
        if not line.startswith('\t') or line.startswith('\t\t'):
            continue
        m = METHOD_RE.match(line)
        if not m or line.lstrip().startswith(('return', 'throw', 'new ', 'await ')):
            continue
        head = line[:m.start(1) + len(m.group(1))]
        if 'new ' in head or 'Exception(' in line:
            continue
        sigs.append((i, m.group(1), m.group(2)))
    sigs.append((len(lines), '<end>', ''))

    eps = []
    for (start, name, args), (stop, _n2, _a2) in zip(sigs, sigs[1:]):
        body = '\n'.join(lines[start:stop])
        paths = sorted({m.group(1) for m in PATH_RE.finditer(body)}
                       | {m.group(1) for m in ABS_RE.finditer(body)})
        if not paths or name in ('CheckToken', 'ByteArrayContent', 'Exception', 'ErrorHandler'):
            continue
        verbs = sorted({m.group(1) for m in VERB_RE.finditer(body)})
        implicit_body = False
        for call in CALL_RE.finditer(body):
            parts = split_args(call.group(1))
            if len(parts) >= 3 and parts[1].strip() == 'null' \
                    and parts[2].strip() not in ('null', 'true', 'false'):
                implicit_body = True
        eps.append({
            'method': name,
            'params': re.sub(r'\s+', ' ', args)[:120],
            'domain': classify(paths),
            'paths': paths,
            'verbs': verbs or ['Get(default)'],
            'body_without_verb': implicit_body,
            'dto': sorted({m.group(1) for m in DTO_RE.finditer(body)}),
            'sender': sorted({m.group(1) for m in SEND_RE.finditer(body)}),
        })
    json.dump(eps, io.open(EP_OUT, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)

    by_domain = defaultdict(list)
    for e in eps:
        by_domain[e['domain']].append(e)

    cmds = [r['command'] for r in
            csv.DictReader(io.open(os.path.join(ROOT, 'command-inventory.csv'), encoding='utf-8-sig'))]

    out = [
        '# 接口地图：客户端直连 Meli 与 ZClient 转发的分界',
        '',
        '证据文件：`analysis/decrypted/MeliApi.resolved.cs`（直连层）、`analysis/decrypted/ZYing_ZClient.resolved.cs`（转发层）、',
        '`analysis/decrypted/ZYing_Meli_MeliErrorParser.resolved.cs`（错误与重试）、`analysis/command-inventory.csv`（命令站点）。',
        '机器提取的端点清单见 `analysis/meli-direct-endpoints.json`。',
        '',
        '## 三条出口',
        '',
        '| 出口 | 形态 | 携带什么 | 谁负责业务规则 |',
        '|---|---|---|---|',
        '| 直连官方 REST | `https://api.mercadolibre.com` + 路径，`Authorization: Bearer <token>` | 单商品/单订单粒度的读写 | 美客多 |',
        '| ZClient 命令转发 | 智赢自有二进制协议 + `cmd=<命令名>` | 业务动作（刊登、同步、拆单、核价…）与批量编排 | 智赢服务器 |',
        '| 智赢 HTTP 服务 | `https://{meli,seller,erp,amazon,ozon,oa,hepsi}.zying.net/api` | 页面数据、配置、日志上报、三方授权跳转 | 智赢服务器 |',
        '',
        '判读：凡是"读一个对象的当前状态"（商品详情、类目属性、索赔进度、绩效）走直连；'
        '凡是"发起一个业务动作"或"要跨账号/跨站点聚合"走命令转发。'
        '这个分界本身就是自研 ERP 的接口分层参考答案：**读官方 API、写走自己的编排服务**。',
        '',
        '## 直连端点（按资源域分组）',
        '',
        '动词列取自调用点：显式写出的按原样记录；写 `GET（sender 默认）` 表示调用点没传动词参数，',
        '由 `SendAsync` 的默认值决定；带 `†` 表示没有显式动词但携带了请求体（GET + body，官方文档多要求 POST，属客户端写法）。',
        '路径列只含字面量段，ID 与站点码由插值拼接，见 `params`。',
        '',
        '值得注意：索赔的 `open-dispute`、`expected-resolutions/partial-refund` 这类动作端点在调用点没有显式动词，'
        '由 sender 默认成 GET（后者还额外带了一个 JSON 请求体）。动作类端点按官方约定通常是 POST，'
        '这里没有做运行时验证，只记录客户端写法；自研不要把动词交给默认值，显式写出才不会在改 sender 时静默变形。',
        '',
    ]
    for dom in [n for n, _k in DOMAIN] + ['其它']:
        rows = by_domain.get(dom)
        if not rows:
            continue
        out += [f'### {dom}', '', '| 客户端方法 | 路径 | 动词 | 返回类型 |', '|---|---|---|---|']
        for e in sorted(rows, key=lambda x: x['method']):
            verb = '/'.join(e['verbs']) if e['verbs'] != ['Get(default)'] else 'GET（sender 默认）'
            if e['body_without_verb']:
                verb += ' †'
            out.append('| `{}` | {} | {} | {} |'.format(
                e['method'],
                '<br>'.join(f'`{p}`' for p in e['paths'][:3]) + (' …' if len(e['paths']) > 3 else ''),
                verb,
                ', '.join(f'`{d}`' for d in e['dto'][:2]) or '—'))
        out.append('')

    out += [
        '## 令牌与授权',
        '',
        '- 授权站点按销售国区分：`auth.mercadolibre.com.{ar,mx,co,br}` / `auth.mercadolibre.cl` 等，'
        '授权链接以 `/authorization?response_type=code&client_id=` 起始，说明客户端直接参与 OAuth 码流程。',
        '- 令牌刷新在直连层完成（`RefreshToken(refreshToken, accessToken)`），换码与刷新两个动作都是对 `/oauth/token` 的 POST，'
        '只是 `grant_type` 不同；并且访问令牌里可解析出 AppId 段：客户端用它定位到具体的应用配置对象。'
        '**具体 AppId/密钥值不在本仓库记录。**',
        '- 发送前统一走 `CheckToken()`，随后为请求挂 `BearerToken`；请求地址由 `BuildUrl()` 拼接。',
        '- 存在三条 sender：`SendAsync`（普通 JSON 请求，51 个方法）、`SendCoreAsync`（自己拼 `HttpRequestMessage` 的低层通道，'
        '只出现在二进制进出两类：图片/视频/附件上传与附件、发票下载，7 个）、`SendVersion2Async`（7 个方法，'
        '**全部集中在全球商品 UP 一族**）。也就是说"v2 通道"不是按 API 代际划分，而是 UP 那一族资源单独走了一条发送路径。',
        '',
        '## 错误处理与重试（可直接借鉴的部分）',
        '',
        '- 统一的 `MeliErrorParser.TryThrow(statusCode, response)`：状态码不在 2xx **或**响应体里带 error 都算失败，'
        '即"HTTP 成功但业务失败"也会抛出。',
        '- 归一化的错误原因标识：`too_many_requests`、`daily_reached`、`internal_error`、`validation_error`，'
        '以及绩效侧的 `historical_behaviour`、`restrictions_coliving`、`comuna_pending`、`phone_pending`。',
        '- 可重试判定 `CanRetry()` 不只按状态码，还按**消息内容**放行，例如 "is not a cbt item"、'
        '"does not have any child for site"、"you dont have item in this site and logistic"、"Attribute id is not valid"、"Picture id …"。',
        '  这类是官方的**最终一致性延迟**（目录/变体/物流关系还没建立），重试而不是报错才是正确行为——'
        '自研的重试策略必须区分"幂等可等待"与"真失败"，否则会把刊登队列卡死。',
        '',
        '### 已观测到的退避实现',
        '',
        '| 位置 | 触发 | 策略 | 备注 |',
        '|---|---|---|---|',
        '| 图片上传 | 状态码 429 / 500 / 502 | 递增退避（约 600ms × 第几次重试）后递归重试 | 把网关 5xx 与限流同等对待 |',
        '| 全球商品读取 | 捕获到 429 状态的 HTTP 异常 | 固定等待约 500ms 后递归重试，默认上限 2 次 | 批量枚举 ID 时调用方把上限提到 5 |',
        '| 家庭系列商品 ID 读取 | 同上 | 同上，默认上限 3 次 | 重试耗尽**返回空值而不抛异常** |',
        '| 命令通道清理后 | 执行完清理类命令 | 固定等待 1 秒再重新取统计 | 客户端主动避让，避免连续写后读 |',
        '| 翻译类循环 | 每条译文之间 | 约 100ms 一条 | 逐条节流，不并发打满 |',
        '| 净收益估算任务 | 每条之间 | 约 10ms 让出 UI 线程 | 主要目的是界面不冻结 |',
        '',
        '判读：智赢对限流的处理是**局部、各自实现**的——上传一套、读 UP 一套、命令通道又一套，'
        '并且重试耗尽后静默返回空值，错误不会上浮。自研应改成**集中式限流器**：'
        '按账号 × 端点维护令牌桶与 429 冷却窗口，重试次数与退避曲线统一配置，'
        '耗尽后进入可观测的失败队列（而不是返回 null），否则批量任务会"看起来跑完了但少了一半数据"。',
        '',
        '### 服务器寻址方式',
        '',
        '- 命令通道的服务器是**编译进客户端的服务器组**：按平台与用途各建一组（美客多主组、UP 组、本地服务组、'
        'Ozon、Wildberries 等），组内地址取自写死的 IP 字面量，另支持按传入地址临时建组；'
        '**没有看到从服务器下发端点的通道**（具体地址不在本仓库记录）。',
        '- 含义：客户端不依赖 DNS 也能连上业务服务器，抗域名封锁；代价是换服务器必须发版。'
        '自研应把端点下发交给服务端配置接口，客户端只保留一个引导地址，同时保留"引导地址不可用时"的备用清单。',
        '',
        '## ZClient 命令通道的事实',
        '',
        f'- 命令清单 {len(cmds)} 个调用点、{len(set(cmds))} 个不同命令名、'
        f'{len({c.split(".")[0] for c in cmds})} 个域前缀，命令名形如 `<域>.<动作>`；'
        '同一动作在跨境/UP/本土三套页面里靠**参数**区分而不是靠不同命令。',
        '- 请求参数两种写法：键值表（字段名即参数名）与一个强类型参数对象（属性名即参数名），'
        '最终都编码成二进制体；响应统一是记录集（`Record`/`RecordSet`），带 `list`、`stat`、`MaxCount` 之类的约定键。',
        '- 分页由服务器返回的记录集自带信息驱动，客户端不保存 SQL：整个 `decrypted` 目录里没有出现任何 '
        '`create table`/`select … from` 语句，**客户端侧没有本地关系库**。',
        '',
        '## 浏览器扩展侧',
        '',
        '- 扩展复用同一套 `CmdHandler?cmd=` 命令契约（`mpush.select`、`shipper.shippers`、`ali1688.product`、`syslog.insert`），'
        '因此它是"同一账号体系的第二前端"，不是独立系统。',
        '- 扩展内联了美客多商品页 URL 前缀（`articulo.mercadolibre.*`）与 1688/淘宝货源页，用于读取页面上下文；'
        '这属于第三方页面自动化，**自研不照做**，改由官方 API 取同一份数据。',
        '',
        '## 限制',
        '',
        '- 以上均为静态观察：没有运行时抓包，未验证任何一条接口在真实账号下的实际响应。',
        '- 直连端点表来自 `MeliApi` 一个类；`ZYing.Work.dll` 内可能还有别的直连实现（该模块受 .NET Reactor 保护，尚未还原）。',
        '- 方法到路径的关联是**同函数窗口**内提取的，插值字符串拼接的动态段（如 `id`、站点码）不在路径里，需读 `params` 列理解。',
    ]
    io.open(DOC_OUT, 'w', encoding='utf-8').write('\n'.join(out) + '\n')
    print('endpoints:', len(eps), '| domains:', {e['domain'] for e in eps})


if __name__ == '__main__':
    sys.exit(main())
