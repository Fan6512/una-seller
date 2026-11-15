# -*- coding: utf-8 -*-
"""Pass 2: targeted runtime string mining (command names + clean CJK UI labels)."""
import re, io, os, collections

PATH = r'C:\Users\ASUS\AppData\Local\Temp\ZYing.DMP'
OUT = r'C:\Users\ASUS\AppData\Local\Temp\zying-rea-20261115\dump_scan'
CHUNK = 48 << 20
OVER = 8192

ascii_re = re.compile(r'[\x20-\x7e]{5,400}')
wide_re = re.compile(r'(?:[ -~][\x00-\xff]|[\x00-\xff][\x4e-\x9f]){3,40}')

# CamelCase with an internal capital, or dotted names
camel_re = re.compile(r'^[A-Za-z][A-Za-z0-9]{2,40}$')
dotted_re = re.compile(r'^[A-Za-z][A-Za-z0-9_]{1,24}(?:\.[A-Za-z][A-Za-z0-9_]{1,24}){1,3}$')

NOISE = re.compile(r'(?i)\.(dll|exe|pdb|json|xml|config|png|jpg|crt|crl|pak|dat|ini|txt|msu|so|wasm)$')
COMMON = set('列表图表价品单存店采集登稿审变体卖点描述类目属性规格库站翻译务明细状态删除保存提交同步图片标题关键词售货订单')

def keep_cn(s):
    if not (3 <= len(s) <= 26):
        return False
    if not re.search(r'[\u4e00-\u9fff]', s):
        return False
    if sum(1 for ch in s if ch in COMMON) < 2:
        return False
    bad = sum(1 for ch in s if '\u4e00' <= ch <= '\u9fff' and ch not in COMMON and ch not in '的是在有和了不与选择项名称版本信息数量时间价格区间管理工具其他所有新增编辑查看设置的')
    return bad <= len(s) * 0.7

c_cmd = collections.Counter()
c_cn = collections.Counter()
c_misc = collections.Counter()

def feed_ascii(s):
    for m in ascii_re.finditer(s):
        v = m.group(0).strip()
        if NOISE.search(v):
            continue
        if dotted_re.match(v):
            c_cmd[v] += 1
        elif camel_re.match(v):
            if any(c.islower() for c in v) and any(c.isupper() for c in v[1:]):
                c_cmd[v] += 1
        low = v.lower()
        if any(k in low for k in ('query', 'spider', 'cmd', 'command', 'zrpc', 'zclient')):
            c_misc[v[:120]] += 1

def feed_wide(s):
    for m in wide_re.finditer(s):
        try:
            v = m.group(0).decode('utf-16le')
        except Exception:
            continue
        v = v.strip()
        if keep_cn(v):
            c_cn[v] += 1

total = os.path.getsize(PATH)
pos = 0
with io.open(PATH, 'rb') as f:
    while pos < total:
        f.seek(pos)
        buf = f.read(CHUNK + OVER)
        if not buf:
            break
        feed_ascii(buf.decode('latin-1'))
        pos += CHUNK

os.makedirs(OUT, exist_ok=True)

def dump(name, ctr, limit):
    items = sorted(ctr.items(), key=lambda x: (-x[1], x[0]))[:limit]
    with io.open(os.path.join(OUT, name), 'w', encoding='utf-8') as w:
        for k, v in items:
            w.write('%5d\t%s\n' % (v, k))
    return len(items)

print('cmd candidates', dump('pass2_cmd.txt', c_cmd, 6000))
print('cn labels', dump('pass2_cn.txt', c_cn, 3000))
print('misc', dump('pass2_misc.txt', c_misc, 1200))

print('\n=== dotted/camelcmd sample (freq>=1, filtered) ===')
for k, v in c_cmd.most_common(120):
    print('  %-40s %d' % (k, v))
print('\n=== cn labels sample ===')
for k, v in c_cn.most_common(120):
    print('  %-26s %d' % (k, v))
