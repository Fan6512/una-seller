# -*- coding: utf-8 -*-
"""Diff two dumps at the string level: strings materialized only in dump2."""
import re, os, io, sys

D1 = r'C:\Users\ASUS\AppData\Local\Temp\ZYing.DMP'
D2 = r'C:\Users\ASUS\AppData\Local\Temp\ZYing (2).DMP'
OUT = r'C:\Users\ASUS\AppData\Local\Temp\zying-rea-20261115\dump_diff'
CHUNK = 32 << 20
OVER = 8192

ascii_re = re.compile(rb'[\x20-\x7e]{6,60}')
wide_re = re.compile(rb'(?:[ -~]\x00|[\x00-\xff][\x4e-\x9f]){2,30}')

def collect(path, tag):
    A, C = set(), set()
    total = os.path.getsize(path)
    pos = 0
    with io.open(path, 'rb') as f:
        while pos < total:
            f.seek(pos)
            buf = f.read(CHUNK + OVER)
            if not buf:
                break
            for m in ascii_re.finditer(buf):
                v = m.group(0)
                if b' ' in v:
                    continue
                A.add(v.decode('ascii', 'ignore'))
            for m in wide_re.finditer(buf):
                try:
                    v = m.group(0).decode('utf-16le')
                except Exception:
                    continue
                cjk = sum(1 for ch in v if '\u4e00' <= ch <= '\u9fff')
                if cjk >= 2:
                    C.add(v)
            pos += CHUNK
    io.open(os.path.join(OUT, 'a_%s.txt' % tag), 'w', encoding='utf-8').write('\n'.join(sorted(A)))
    io.open(os.path.join(OUT, 'c_%s.txt' % tag), 'w', encoding='utf-8').write('\n'.join(sorted(C)))
    print('%s ascii=%d cjk=%d' % (tag, len(A), len(C)))
    return A, C

os.makedirs(OUT, exist_ok=True)
a1, c1 = collect(D1, 'd1')
a2, c2 = collect(D2, 'd2')

new_a = sorted(a2 - a1)
new_c = sorted(c2 - c1)
print('NEW ascii=%d  NEW cjk=%d' % (len(new_a), len(new_c)))

io.open(os.path.join(OUT, 'new_ascii.txt'), 'w', encoding='utf-8').write('\n'.join(new_a))
io.open(os.path.join(OUT, 'new_cjk.txt'), 'w', encoding='utf-8').write('\n'.join(new_c))

COMMON = set('列表图表价品单存店采集登稿审变体卖点描述类目属性规格库站翻译务明细状态删除保存提交同步图片标题关键词售货订单成功失败待处理已未')
def good_cn(s):
    if not (2 <= len(s) <= 22):
        return False
    cjk = [ch for ch in s if '\u4e00' <= ch <= '\u9fff']
    if len(cjk) < 2:
        return False
    return sum(1 for ch in cjk if ch in COMMON) >= 1

gc = [s for s in new_c if good_cn(s)]
print('\n=== 新增中文（过滤后 %d 条） ===' % len(gc))
for s in gc[:200]:
    print('  ', s)

verb = re.compile(r'^(?:Query|Get|Set|Add|Update|Delete|Save|Insert|Load|List|Publish|Sync|Check|'
                  r'Submit|Create|Edit|Remove|Move|Copy|Import|Export|Search|Find|Count|Reset|'
                  r'Bind|Unbind|Login|Logout|Upload|Download|Push|Pull|Batch|Task|Report|Refresh|'
                  r'Apply|Confirm|Cancel|Spider|Sale|Product)[A-Za-z0-9]{3,40}$')
cand = [s for s in new_a if verb.match(s)]
print('\n=== 新增动词式标识符（%d 条） ===' % len(cand))
for s in cand[:200]:
    print('  ', s)
