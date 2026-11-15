# -*- coding: utf-8 -*-
"""Pass 3 on the SECOND dump: JSON payloads, Command field, verb-prefixed identifiers."""
import re, os, io, collections, hashlib

P2 = r'C:\Users\ASUS\AppData\Local\Temp\ZYing (2).DMP'
OUT = r'C:\Users\ASUS\AppData\Local\Temp\zying-rea-20261115\dump_scan2'
CHUNK = 32 << 20
OVER = 16384

pats = {
    'cmd_field': re.compile(rb'(?:"Command"|\x00C\x00o\x00m\x00m\x00a\x00n\x00d\x00\x00?)'),
    'sale_key': re.compile(rb'"sale_|\x00"\x00s\x00a\x00l\x00e\x00_\x00|\x00"\x00s\x00a\x00l\x00e\x00_'),
    'json_like': re.compile(rb'\{"[a-zA-Z_]{2,24}":'),
}

verb_re = re.compile(
    r'^(?:Query|Get|Set|Add|Update|Delete|Save|Insert|Load|List|Publish|Sync|Check|Submit|'
    r'Create|Edit|Remove|Move|Copy|Import|Export|Search|Find|Count|Reset|Bind|Unbind|Login|'
    r'Logout|Upload|Download|Push|Pull|Batch|Task|Report|Refresh|Apply|Confirm|Cancel)'
    r'[A-Z][A-Za-z0-9]{2,40}$')
ascii_re = re.compile(rb'[\x20-\x7e]{4,200}')

c_verb = collections.Counter()
wins = []
seen = set()

def grab(buf, pos, half=1400):
    a = max(0, pos - half)
    b = min(len(buf), pos + half)
    w = buf[a:b]
    h = hashlib.md5(w).hexdigest()
    if h in seen:
        return
    seen.add(h)
    wins.append(w)

total = os.path.getsize(P2)
pos = 0
with io.open(P2, 'rb') as f:
    while pos < total:
        f.seek(pos)
        buf = f.read(CHUNK + OVER)
        if not buf:
            break
        for name, pat in pats.items():
            n = 0
            for m in pat.finditer(buf):
                grab(buf, m.start())
                n += 1
                if n > 400:
                    break
        for m in ascii_re.finditer(buf):
            try:
                v = m.group(0).decode('ascii')
            except Exception:
                continue
            if verb_re.match(v):
                c_verb[v] += 1
        pos += CHUNK

os.makedirs(OUT, exist_ok=True)
print('windows captured', len(wins))
print('verb-prefixed identifiers', len(c_verb))

# 从窗口里提取 JSON 对象 / 键值
json_objs = []
kv = collections.Counter()
for w in wins:
    for enc in ('latin-1', 'utf-16le'):
        try:
            s = w.decode(enc, 'ignore')
        except Exception:
            continue
        for m in re.finditer(r'\{[^{}\x00]{20,3000}\}', s):
            t = m.group(0)
            if '"sale_' in t or '"Command"' in t or '"name"' in t:
                json_objs.append(t)
        for m in re.finditer(r'"(Command|cmd|method|action)"\s*:\s*"([^"\x00]{1,60})"', s):
            kv[(m.group(1), m.group(2))] += 1

uniq = {}
for j in json_objs:
    uniq[hashlib.md5(j.encode('utf-8','ignore')).hexdigest()] = j
print('unique json objects', len(uniq))
print('command kv pairs', len(kv))

with io.open(os.path.join(OUT, 'json_payloads.txt'), 'w', encoding='utf-8') as f:
    for j in list(uniq.values())[:400]:
        f.write(j + '\n---\n')
with io.open(os.path.join(OUT, 'verb_ids.txt'), 'w', encoding='utf-8') as f:
    for k, v in sorted(c_verb.items(), key=lambda x: (-x[1], x[0])):
        f.write('%5d\t%s\n' % (v, k))

print('\n=== Command 键值 ===')
for (k, v), n in kv.most_common(40):
    print('  %-10s = %-40s %d' % (k, v, n))

print('\n=== verb 前缀标识符（前 120） ===')
for k, v in c_verb.most_common(120):
    print('  %-42s %d' % (k, v))

print('\n=== JSON 样本（前 8 个，各截 700 字） ===')
for j in list(uniq.values())[:8]:
    print('---')
    print(j[:700])
