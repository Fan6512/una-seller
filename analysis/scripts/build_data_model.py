"""Generate record-families.json and data-model.md.

The client keeps no local relational database, so its observable data model is the
set of columns the server returns inside RecordSets. Those columns are namespaced by
entity prefix (acc_/stat_/up_/claim_/mcm_/qa_/msg_/...), which is what this extracts,
then joins with the direct-API DTO list to describe how the pieces relate.
Emits identifier inventories plus my own analysis; no source bodies.
"""
import csv
import hashlib
import io
import json
import os
import re
import sys
from collections import Counter, defaultdict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DEC = os.path.join(ROOT, 'decrypted')
FAM_OUT = os.path.join(ROOT, 'record-families.json')
DOC_OUT = os.path.join(ROOT, 'data-model.md')

GET = re.compile(r'\.(Get|GetInt|GetBool|GetDecimal|GetString)\(\s*'
                 r'(?:(?:[\w$]+\.|[\w$]+\([^)]*\)\s*\.)?\s*(?:/\*\d+\*/)?\s*)?"([a-z_][a-z0-9_]{2,42})"')
IDX = re.compile(r'\[\s*(?:(?:[\w$]+\.|[\w$]+\([^)]*\)\s*\.)?\s*(?:/\*\d+\*/)?\s*)?"([a-z_][a-z0-9_]{2,42})"\s*\]')

FAMILY_LABEL = {
    'acc': '授权账号', 'stat': '跨境已上传明细', 'up': '全球商品（User Product）',
    'claim': '售后索赔', 'mcm': '跟卖监控（Catalog Monitor）', 'qa': '售前问答（本土）',
    'q': '售前问答（跨境）', 'msg': '站内信', 'metrics': '店铺绩效', 'meli': '美客多杂项（重量/计数）',
    'log': '操作日志', 'order': '订单', 'attr': '类目属性', 'as': '自动服务状态',
    'catalog': '目录商品', 'local': '本土站点记录', 'refer': '分类对照', 'temp': '模板',
}
JOIN = ['aid', 'saleid', 'itemid', 'cbt', 'cbtid', 'localid', 'upid', 'cateid',
        'claimid', 'oid', 'site', 'abbr', 'acc', 'loginid', 'group']


def main():
    keys, per_class = Counter(), defaultdict(set)
    seen = set()
    for dirpath, _dn, files in os.walk(DEC):
        for fn in sorted(files):
            if not fn.endswith('.cs'):
                continue
            path = os.path.join(dirpath, fn)
            raw = io.open(path, 'rb').read()
            # The dump contains twin files per class (one named by the internal type,
            # one by the Chinese form title) with byte-identical bodies; counting both
            # would double every read site.
            digest = hashlib.sha1(raw).hexdigest()
            if digest in seen:
                continue
            seen.add(digest)
            text = raw.decode('utf-8', 'replace')
            cls = re.search(r'class\s+([A-Za-z_][\w<>]*)', text)
            cls = cls.group(1) if cls else fn.replace('.resolved.cs', '')[:-3]
            for rx in (GET, IDX):
                for m in rx.finditer(text):
                    k = m.group(2) if rx is GET else m.group(1)
                    keys[k] += 1
                    per_class[cls].add(k)

    fams = defaultdict(Counter)
    for k, c in keys.items():
        if '_' in k:
            fams[k.split('_', 1)[0]][k] = c

    payload = {
        'total_distinct_keys': len(keys),
        'families': {a: sorted(f) for a, f in
                     sorted(fams.items(), key=lambda x: -len(x[1])) if len(f) >= 3},
        'top_keys': [k for k, _c in keys.most_common(40)],
        'classes_scanned': len(per_class),
    }
    json.dump(payload, io.open(FAM_OUT, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)

    eps = json.load(io.open(os.path.join(ROOT, 'meli-direct-endpoints.json'), encoding='utf-8'))
    dtos = sorted({d for e in eps for d in e['dto']})
    rows = list(csv.DictReader(io.open(os.path.join(ROOT, 'command-inventory.csv'), encoding='utf-8-sig')))
    resp_envelope = Counter(k for r in rows for k in r['response_keys'].split('|') if k)

    out = [
        '# 数据模型：智赢客户端看到的数据形状',
        '',
        '## 最关键的一条结论',
        '',
        '**客户端没有本地关系型数据库。** 整个反编译集合里没有出现 `create table` / `select … from` / ORM 映射，',
        '列表数据全部是服务器返回的**记录集（RecordSet）**，列名由服务器决定。',
        f'客户端读到的不同列名共 {len(keys)} 个，其中按实体前缀成族的有 {len(payload["families"])} 族。',
        '',
        '这意味着：智赢的"数据模型"其实是**服务器视图的投影**。前缀就是视图归属——同一条商品记录在',
        '跨境已上传明细里叫 `stat_itemid`，在全球商品里叫 `up_itemid`，在跟卖监控里叫 `mcm_itemid`，',
        '客户端把它们统一读进 `_itemid` 这样的本地变量。**自研不要复刻这种设计**：它把 JOIN 逻辑压进了',
        '每条 SQL 视图，字段语义随页面漂移，跨功能复用只能靠服务器再写一个视图。',
        '',
        '## 记录列前缀族（服务器视图归属）',
        '',
        '| 前缀 | 归属实体 | 列数 | 代表列 |',
        '|---|---|---|---|',
    ]
    for a, cols in payload['families'].items():
        out.append(f'| `{a}_` | {FAMILY_LABEL.get(a, "（未归类）")} | {len(cols)} | ' +
                   ', '.join(f'`{c}`' for c in cols[:6]) + (' …' if len(cols) > 6 else '') + ' |')

    out += [
        '',
        '## 贯穿各族的连接键',
        '',
        '| 键 | 含义（据使用位置判读） | 出现频次 |',
        '|---|---|---|',
    ]
    for k in JOIN:
        if k in keys:
            out.append(f'| `{k}` | {JOIN_NOTE[k]} | {keys[k]} |')

    out += [
        '',
        '规律：**账号维度用 `aid`，行维度用 `saleid`/`localid`，平台对象维度用 `itemid`/`cbtid`/`upid`**，',
        '站点维度用 `abbr`（站点缩写）+ `site`。三者构成客户端所有列表页的过滤条件，也正好是自研系统里',
        '`account_id` / `listing_id` / `platform_item_id` / `site` 这套主键的外向映射。',
        '',
        '## 记录集信封约定',
        '',
        '命令返回不是裸数组，而是带元信息的信封。高频返回键：' +
        ', '.join(f'`{k}`({c})' for k, c in resp_envelope.most_common(8)) + '。',
        '其中 `list` 是数据体，`root` 是树形结构的父节点，`stat`/`stats` 是汇总口径，`abbr`/`areas`/`flags` 是',
        '字典表。分页信息随记录集一起返回（记录对象上可直接取总条数），客户端不自己算总数。',
        '',
        '## 直连官方 API 的对象模型',
        '',
        f'与上面不同，`MeliApi` 直连层使用强类型 DTO 反序列化，共 {len(dtos)} 个：',
        '',
        '`' + '`, `'.join(dtos) + '`',
        '',
        '这些 DTO 与官方资源同名同形（商品、类目、索赔、消息包、绩效、用户产品），'
        '是**自研服务端真正该建模的部分**；前缀族属于智赢私有视图，不必继承。',
        '',
        '## 建议的落库形态（对照结论）',
        '',
        '| 智赢可见形态 | 自研对应 | 理由 |',
        '|---|---|---|',
        '| 服务器视图前缀族 | 规范化表 + 视图只在 API 层拼装 | 字段语义不该随页面变化 |',
        '| 记录集分页元信息 | 表内 `updated_at` + 游标 | 换设备后要能续传，而不是重拉 |',
        '| `aid`/`saleid`/`itemid` 三键 | `account_id`/`listing_id`/`platform_item_id` | 一一对应，且是同步与幂等的锚点 |',
        '| 无本地库 | 客户端本地缓存（快照 + 变更游标） | 智赢每次都要回源，弱网与冷启动体验差；'
        '自研把服务器当唯一真相、客户端只作缓存与出站队列 |',
        '',
        '## 限制',
        '',
        f'- 覆盖范围是已还原的 {payload["classes_scanned"]} 个类型，集中在美客多相关窗体；'
        '受 .NET Reactor 保护的模块尚未还原，非美客多板块的列族没有统计进来。',
        '- 列名到实体的归属依据是**前缀 + 使用位置**，属推断，不是服务器 schema 的直接观测。',
        '- 频次是"该列名在已还原代码里被读取的次数"，按文件内容去重后统计（反编译产物里每个窗体存在两份同名不同文件名的副本），'
        '不代表数据量或重要性。',
        '- 全部为静态观察，未验证任何返回结构在真实账号下的实际形态。',
    ]
    io.open(DOC_OUT, 'w', encoding='utf-8').write('\n'.join(out) + '\n')
    print('distinct keys:', len(keys), '| families>=3:', len(payload['families']),
          '| classes:', payload['classes_scanned'])


JOIN_NOTE = {
    'aid': '智赢侧的授权账号主键（不是美客多 user_id，后者另有字段）',
    'saleid': '上传/刊登记录行主键，跨境与本土共用同一命名',
    'itemid': '美客多商品 ID（MLM/MLA… 站点对象）',
    'cbt': '跨境商品标识',
    'cbtid': '跨境商品 ID，与 `itemid` 常成对出现',
    'localid': '本土站点的行主键',
    'upid': '全球商品（User Product）ID',
    'cateid': '类目 ID',
    'claimid': '索赔单 ID',
    'oid': '订单 ID',
    'site': '站点码（MLM/MLA/…）',
    'abbr': '站点缩写，用于字典与筛选',
    'acc': '账号显示名（列表展示用，非主键）',
    'loginid': '登录会话/设备标识，同步链路相关',
    'group': '分组键，用于账号或记录分组',
}


if __name__ == '__main__':
    sys.exit(main())
