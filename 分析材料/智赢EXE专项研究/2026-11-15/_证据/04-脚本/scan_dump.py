# -*- coding: utf-8 -*-
"""Scan a full user-mode dump for runtime strings. Read-only, no process attach."""
import re, io, os, sys, collections

PATH = r'C:\Users\ASUS\AppData\Local\Temp\ZYing.DMP'
OUT = r'C:\Users\ASUS\AppData\Local\Temp\zying-rea-20261115\dump_scan'
CHUNK = 48 << 20
OVER = 8192

ascii_re = re.compile(r'[\x20-\x7e]{6,3000}')
wide_re = re.compile(r'[ -\~\u4e00-\u9fff\u3000-\u303f\uff00-\uffef]{4,1500}')

sale_re = re.compile(r'sale_[a-z0-9_]{2,24}')
jsonkey_re = re.compile(r'"([a-zA-Z_][a-zA-Z0-9_]{1,30})"\s*:')
cmd_re = re.compile(r'^[A-Z][A-Za-z0-9]{2,28}(?:\.[A-Za-z][A-Za-z0-9]{1,28})?$')
dot_re = re.compile(r'^[a-z][a-z0-9_]{1,10}\.[a-z][a-z0-9_]{1,18}$')
cn_re = re.compile(r'[\u4e00-\u9fff]{2,}')
url_re = re.compile(r'https?://[A-Za-z0-9\.\-/:_\?=&%#]{6,120}')

CN_KW = ('采集', '草稿', '刊登', '审核', '变体', '卖点', '描述', '图片', '站点', '类目',
         '属性', '库存', '重复', '主库', '店铺', '上架', '导入', '导出', '筛选', '定价',
         '规格', '翻译', '明细', '列表', '状态', '失败', '待', '删除', '保存', '提交', '同步')
A_KW = ('sale_', 'spider', 'Spider', 'ZClient', 'Zrpc', 'Command', '/ws', 'api', 'Api',
        'Json', 'json', 'Record', 'Draft', 'Publish', 'Market', 'Meli', 'Mercado', 'Cate',
        'Attr', 'Variant', 'Sku', 'SKU', 'Stock', 'Bullet', 'Intro', 'Brand', 'SiteId',
        'Query', 'Insert', 'Update', 'Detail', 'List', 'Token', 'http')

c_ascii = collections.Counter()
c_cmd = collections.Counter()
c_dot = collections.Counter()
c_sale = collections.Counter()
c_json = collections.Counter()
c_cn = collections.Counter()
c_cnkw = collections.Counter()
c_url = collections.Counter()
cn_ctx = collections.Counter()

def feed_ascii(s):
    for m in ascii_re.finditer(s):
        v = m.group(0).strip()
        if not v or v.isdigit():
            continue
        if any(k in v for k in A_KW):
            c_ascii[v[:160]] += 1
        if cmd_re.match(v):
            c_cmd[v] += 1
        elif dot_re.match(v):
            c_dot[v] += 1
        for f in sale_re.findall(v):
            c_sale[f] += 1
        for k in jsonkey_re.findall(v):
            c_json[k] += 1
        if 'http' in v:
            for u in url_re.findall(v):
                c_url[u] += 1

def feed_wide(s):
    for m in wide_re.finditer(s):
        v = m.group(0).strip()
        if len(v) < 4:
            continue
        if cn_re.search(v):
            c_cn[v[:160]] += 1
            hit = [k for k in CN_KW if k in v]
            if hit:
                c_cnkw[v[:160]] += 1
                for k in hit:
                    cn_ctx[k] += 1
        if any(k in v for k in A_KW):
            c_ascii[v[:160]] += 1
            for f in sale_re.findall(v):
                c_sale[f] += 1
            for k in jsonkey_re.findall(v):
                c_json[k] += 1
            if 'http' in v:
                for u in url_re.findall(v):
                    c_url[u] += 1

def main():
    total = os.path.getsize(PATH)
    os.makedirs(OUT, exist_ok=True)
    pos = 0
    n = 0
    with io.open(PATH, 'rb') as f:
        while pos < total:
            f.seek(pos)
            buf = f.read(CHUNK + OVER)
            if not buf:
                break
            try:
                feed_ascii(buf.decode('latin-1'))
            except Exception:
                pass
            try:
                feed_wide(buf.decode('utf-16le', 'ignore'))
            except Exception:
                pass
            n += 1
            sys.stderr.write('chunk %d  %d / %d MB\n' % (n, pos >> 20, total >> 20))
            sys.stderr.flush()
            pos += CHUNK

    def dump(name, ctr, minfreq=1, limit=4000):
        p = os.path.join(OUT, name)
        items = [(k, v) for k, v in ctr.items() if v >= minfreq]
        items.sort(key=lambda x: (-x[1], x[0]))
        with io.open(p, 'w', encoding='utf-8') as w:
            for k, v in items[:limit]:
                w.write('%6d\t%s\n' % (v, k))
        return len(items)

    stat = {}
    stat['sale_fields'] = dump('sale_fields.txt', c_sale)
    stat['cmd_like'] = dump('cmd_like.txt', c_cmd, 1, 3000)
    stat['dot_like'] = dump('dot_like.txt', c_dot, 1, 2000)
    stat['json_keys'] = dump('json_keys.txt', c_json, 1, 3000)
    stat['cn_ui'] = dump('cn_ui.txt', c_cnkw, 1, 3000)
    stat['cn_all'] = dump('cn_all.txt', c_cn, 2, 4000)
    stat['urls'] = dump('urls.txt', c_url, 1, 1000)
    stat['ascii_hits'] = dump('ascii_hits.txt', c_ascii, 2, 4000)

    print('=== SUMMARY ===')
    for k, v in stat.items():
        print(' %-12s %d' % (k, v))
    print('\n=== sale_* (top 60) ===')
    for k, v in c_sale.most_common(60):
        print('  %-28s %d' % (k, v))
    print('\n=== json keys (top 60) ===')
    for k, v in c_json.most_common(60):
        print('  %-30s %d' % (k, v))
    print('\n=== cmd-like (top 50) ===')
    for k, v in c_cmd.most_common(50):
        print('  %-34s %d' % (k, v))
    print('\n=== dot-like (top 40) ===')
    for k, v in c_dot.most_common(40):
        print('  %-34s %d' % (k, v))
    print('\n=== urls (top 40) ===')
    for k, v in c_url.most_common(40):
        print('  %-70s %d' % (k, v))

if __name__ == '__main__':
    main()
