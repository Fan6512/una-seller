"""Generate feature-map.md: menu feature -> form class -> ZClient commands -> fields.

Join basis: each form class declares a Key constant that equals the menu entry key,
so the menu tree, command inventory and per-class field index can be linked exactly.
Output is identifiers plus my own summary text; no source bodies.
"""
import csv
import io
import json
import os
import re
from collections import defaultdict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DEC = os.path.join(ROOT, 'decrypted')
CSV_IN = os.path.join(ROOT, 'command-inventory.csv')
MENU = os.path.join(ROOT, 'menu-tree.json')
FIELDS = os.path.join(ROOT, 'field-index.json')
OUT = os.path.join(ROOT, 'feature-map.md')

CLS_RE = re.compile(r'^\s*(?:public|private|internal|protected|\s)+(?:static\s+|sealed\s+|partial\s+)*class\s+([A-Za-z_][\w<>]*)', re.M)
KEY_CONST = re.compile(r'const\s+string\s+Key\s*=\s*"([^"]+)"')
SECTION_LABEL = {
    'Meli': '美客多（Meli）', 'Products': '产品与采集', 'Orders': '订单与采购',
    'Stocks': '库存与仓储', 'Moneys': '财务', 'Systems': '系统设置', 'Choose': '选品',
    'Advertising': '广告', 'Amazons': '亚马逊', 'Shopee': 'Shopee', 'Lazada': 'Lazada',
    'AliExpress': '速卖通', 'TikTok': 'TikTok', 'Temu': 'Temu', 'Shein': 'Shein',
    'Walmart': 'Walmart', 'Hepsi': 'Hepsi', 'Coupang': 'Coupang', 'B2W': 'B2W',
    'BOL': 'BOL', 'Allegro': 'Allegro', 'Linio': 'Linio', 'N11s': 'N11',
    'Wishes': 'Wish', 'EBays': 'eBay', 'Shopifys': 'Shopify', 'Service': '服务',
    'PHHGroup': '拍拍档/海外仓', 'WeoBuy': 'WeoBuy',
}
MAIN_SECTIONS = {'Meli', 'Products', 'Orders', 'Stocks', 'Moneys'}
MELI_ONLY = MAIN_SECTIONS


def load_class_keys():
    out = {}
    for dirpath, _dn, files in os.walk(DEC):
        for fn in files:
            if not fn.endswith('.cs'):
                continue
            p = os.path.join(dirpath, fn)
            raw = io.open(p, 'rb').read()
            try:
                text = raw.decode('utf-8')
            except UnicodeDecodeError:
                text = raw.decode('gbk', 'replace')
            km = KEY_CONST.search(text)
            if not km:
                continue
            cm = CLS_RE.search(text)
            if cm:
                out.setdefault(km.group(1), set()).add(cm.group(1))
    return out


def main():
    keys = load_class_keys()
    menu = json.load(io.open(MENU, encoding='utf-8'))
    fields = json.load(io.open(FIELDS, encoding='utf-8'))
    rows = list(csv.DictReader(io.open(CSV_IN, encoding='utf-8-sig')))
    by_class = defaultdict(list)
    for r in rows:
        by_class[r['owner_type']].append(r)

    lines = ['# 功能地图：菜单功能 → 窗体类 → ZClient 命令 → 字段',
             '',
             '生成方式：菜单项 key 与窗体类声明的 `Key` 常量精确关联，命令与字段取自 '
             '`command-inventory.csv`、`field-index.json`。',
             '',
             '说明：客户端没有本地关系库，列表数据来自服务器返回的 RecordSet；'
             '因此下文的"数据对象"指服务器返回的记录集名称，而非本地表名。',
             '']

    stats = {'features': 0, 'linked': 0, 'commands': 0}
    minor = []
    for section, groups in menu.items():
        if section not in MELI_ONLY:
            feats = [it for items in groups.values() for it in items]
            linked = sum(1 for it in feats if keys.get(it['key']) and
                         any(by_class.get(c) for c in keys[it['key']]))
            minor.append((SECTION_LABEL.get(section, section), len(feats), linked))
            continue
        lines += ['', f'## {SECTION_LABEL.get(section, section)}']
        resolved = sum(1 for items in groups.values() for it in items if keys.get(it['key']))
        if not resolved:
            lines.append('')
            lines.append('（本板块窗体尚未还原，仅登记功能项名称）：' +
                         '、'.join(it['label'] for items in groups.values() for it in items))
            continue
        for group, items in groups.items():
            lines.append('')
            lines.append(f'### {group}')
            lines.append('')
            lines.append('| 功能 | 窗体类 | 命令 | 请求字段 | 返回对象 |')
            lines.append('|---|---|---|---|---|')
            for it in items:
                stats['features'] += 1
                classes = keys.get(it['key'], set())
                cmd_rows = [r for c in classes for r in by_class.get(c, [])]
                if cmd_rows:
                    stats['linked'] += 1
                    stats['commands'] += len({r['command'] for r in cmd_rows})
                cmds = sorted({r['command'] for r in cmd_rows}) or ['（无命令站点，纯本地/网页页）']
                req = sorted({k for r in cmd_rows for k in r['request_keys'].split('|') if k})
                resp = sorted({k for r in cmd_rows for k in r['response_keys'].split('|') if k})
                flds = sorted(set(req) | {k for c in classes for k in fields.get(c, {})} )
                lines.append('| {}{} | {} | {} | {} | {} |'.format(
                    it['label'], ' ※按账号灰度' if it['gate'] else '',
                    ', '.join(sorted(classes)) or '（未定位到类）',
                    '<br>'.join(f'`{c}`' for c in cmds[:8]) + (f'<br>…共{len(cmds)}' if len(cmds) > 8 else ''),
                    ', '.join(f'`{f}`' for f in flds[:10]) + (' …' if len(flds) > 10 else '') or '—',
                    ', '.join(f'`{r}`' for r in resp[:8]) or '—'))
    lines += ['', '## 其它平台板块（仅规模，未展开）', '',
              '| 板块 | 功能项 | 有命令站点的功能项 |', '|---|---|---|']
    lines += [f'| {name} | {total} | {linked} |' for name, total, linked in sorted(minor, key=lambda x: -x[1])]
    lines += ['', '## 覆盖范围', '',
              f'- 全平台菜单共 {len(menu)} 个板块、' +
              f'{sum(len(i) for g in menu.values() for i in g.values())} 个功能项；'
              '本文件只展开美客多/产品/订单/库存/财务五个板块。',
              '- **只有美客多板块的窗体类完成还原**（菜单 key 能解析到类的 22/25 个，其中 18 个有命令站点）；'
              '产品、订单、库存、财务等板块的窗体在受 .NET Reactor 保护的模块里，尚未还原，'
              '所以对应表格里的"窗体类/命令"列为空——这是取证覆盖限制，不代表那些功能没有实现。',
              '- 展开表里的"请求字段"是同一函数窗口内出现的键名集合，可能混入读响应时用到的列名。',
              '', '## 读法', '',
              '1. 功能项与命令是**多对多**：同一命令常跨跨境/UP/本土三套页面复用，差别只在参数（如 `aid` 的口径）。',
              '2. "无命令站点"的功能并不等于没实现：它们要么是纯本地表格编辑，要么是内嵌网页/服务器侧渲染页，'
              '取证重心应转向 `api-map.md` 的直连接口。',
              '3. `※按账号灰度` 来自菜单构建里的账号判断，说明智赢按账号 ID 白名单开放新功能——'
              '自研对应做法是功能开关放在服务端下发，而不是编译进包。']
    json.dump(stats, io.open(os.path.join(ROOT, 'feature-map-stats.json'), 'w', encoding='utf-8'),
              ensure_ascii=False)
    print('features:', stats['features'], '| linked:', stats['linked'], '| commands:', stats['commands'])
    io.open(OUT, 'w', encoding='utf-8').write('\n'.join(lines) + '\n')


if __name__ == '__main__':
    main()
