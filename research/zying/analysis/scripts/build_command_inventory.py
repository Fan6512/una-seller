"""Build command-inventory.csv from resolved (literal-restored) sources.

For every ZClient command site, recover: the command id, which server pool the
client is created against, the request parameter names, the response keys read
back, and the enclosing type/method so each row traces to a UI surface.
Dynamic commands ("prefix." + arg) are expanded from their call sites.
Output holds identifiers and locations only, never source bodies.
"""
import bisect
import csv
import io
import os
import re

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(ROOT, 'decrypted')
OUT = os.path.join(ROOT, 'command-inventory.csv')

OBF = r'(?:(?:[\w$]+\.|[\w$]+\([^)]*\)\s*\.)?\s*(?:/\*\d+\*/)?\s*)'
CMD_RE = re.compile(r'\bCommand\s*=\s*' + OBF + r'"([a-z0-9_]+(?:\.[a-z0-9_]+)*)"')
DYN_RE = re.compile(r'\bCommand\s*=\s*' + OBF + r'"([a-z0-9_]+\.?)"\s*\+\s*[\w$.\[\]()]+')
KEY_RE = re.compile(r'\[\s*' + OBF + r'"([a-zA-Z_][\w.]*)"\s*\]\s*='
                    r'|\bAdd\s*\(\s*' + OBF + r'"([a-zA-Z_][\w.]*)"')
POST_RE = re.compile(r'PostAsync\w*\s*\(\s*' + OBF + r'"([a-zA-Z_][\w.]*)"')
RESP_RE = re.compile(r'\b(?:zc|zClient|client)\s*\[\s*' + OBF + r'"([a-zA-Z_][\w.]*)"\s*\]')
KIND_RE = re.compile(r'ZClient\.(Create\w*|CreateForAI)\s*\(([^)]{0,60})\)')
DECL_RE = re.compile(r'^\s*(?:public|private|internal|protected|static|async|virtual|override|sealed|\s)+'
                     r'(?:[\w<>\[\],.?]+\s+)+([A-Za-z_]\w*)\s*\([^)]*\)\s*\{?\s*$')
TYPE_RE = re.compile(r'^\s*(?:public|private|internal|protected|\s)+'
                     r'(?:static\s+|sealed\s+|partial\s+)*class\s+([A-Za-z_][\w<>]*)')
PARA_PROP = re.compile(r'^\s*' + OBF + r'([a-zA-Z_]\w*)\s*=(?!=)')
SKIP = {'if', 'for', 'foreach', 'while', 'switch', 'using', 'lock', 'return', 'catch',
        'new', 'do', 'else', 'fixed', 'get', 'set', 'add', 'remove', 'when', 'yield', 'await'}


def index_methods(lines):
    """One pass: [(line, name, owning_type)] for every method-like declaration."""
    out, typ = [], ''
    for i, line in enumerate(lines):
        m = TYPE_RE.match(line)
        if m:
            typ = m.group(1)
            continue
        m = DECL_RE.match(line)
        if not m or m.group(1) in SKIP or '=' in line or 'throw' in line:
            continue
        out.append((i, m.group(1), typ))
    return out


def collect(lines, start, end):
    req, resp, kind, in_para = [], [], '', 0
    for i in range(start, min(end, len(lines))):
        line = lines[i]
        if not CMD_RE.search(line):
            for m in KEY_RE.finditer(line):
                n = m.group(1) or m.group(2)
                if n and n not in req:
                    req.append(n)
        for m in POST_RE.finditer(line):
            if m.group(1) not in req:
                req.append(m.group(1))
        for m in RESP_RE.finditer(line):
            if m.group(1) not in resp:
                resp.append(m.group(1))
        m = KIND_RE.search(line)
        if m and not kind:
            kind = m.group(2).strip().replace('ZClient.', '') or 'default'
        if re.search(r'new\s+Para\s*[\{\(]', line):
            in_para = 1
        elif in_para:
            if re.match(r'^\s*[};,]*\}\s*$', line):
                in_para = 0
            else:
                pm = PARA_PROP.match(line)
                if pm and pm.group(1) not in SKIP and pm.group(1) not in req:
                    req.append(pm.group(1))
    return kind, req[:16], resp[:12]


def expand_dynamic(lines, meth, lo, hi):
    if not meth:
        return []
    pat = re.compile(re.escape(meth) + r'\(\s*' + OBF + r'"([a-z0-9_.]+)"')
    return [(m.group(1), i + 1) for i in range(lo, min(hi, len(lines))) for m in pat.finditer(lines[i])]


def norm_server(kind):
    """The dump of MeliApiBuilder lost some GBK literals; only three pool names exist."""
    if not kind:
        return ''
    if kind.startswith('Member.'):
        return '美客多|美客多2(按转运服务分支)'
    if kind in ('ip', 'obj2.ip'):
        return '指定服务器IP'
    if '美客多2' in kind:
        return '美客多2'
    if '美客多' in kind or any(ord(c) > 127 for c in kind):
        return '美客多'
    return kind


class Inventory:
    def __init__(self):
        self.rows, self.seen = [], {}

    def emit(self, cmd, typ, meth, rel, ln, kind, req, resp, dyn=''):
        key = (cmd, typ, meth, ln, tuple(req), tuple(resp))
        if key in self.seen:
            src = self.seen[key]['source']
            if rel not in src.split(';'):
                self.seen[key]['source'] = src + ';' + rel
            return
        row = {'command': cmd, 'server': kind, 'owner_type': typ, 'method': meth, 'source': rel,
               'line': ln, 'request_keys': '|'.join(req), 'response_keys': '|'.join(resp),
               'dynamic_prefix': dyn}
        self.seen[key] = row
        self.rows.append(row)

    def scan(self):
        for dirpath, _dn, filenames in os.walk(SRC):
            for fn in sorted(filenames):
                if not fn.endswith('.cs'):
                    continue
                path = os.path.join(dirpath, fn)
                try:
                    raw = io.open(path, 'rb').read()
                except OSError:
                    continue
                try:
                    text = raw.decode('utf-8')
                except UnicodeDecodeError:
                    text = raw.decode('gbk', 'replace')
                lines = text.splitlines()
                rel = os.path.relpath(path, SRC).replace('\\', '/')
                stem = fn[:-len('.resolved.cs')] if fn.endswith('.resolved.cs') else fn[:-3]
                index = index_methods(lines)
                starts = [s for s, _n, _t in index]
                cache = {}
                for i, line in enumerate(lines):
                    cm, dm = CMD_RE.search(line), DYN_RE.search(line)
                    if not (cm or dm):
                        continue
                    pos = bisect.bisect_right(starts, i) - 1
                    if pos < 0:
                        start, end, meth, typ = 0, i + 25, '', ''
                    else:
                        start, meth, typ = index[pos][0], index[pos][1], index[pos][2]
                        end = index[pos + 1][0] if pos + 1 < len(index) else len(lines)
                    if start not in cache:
                        cache[start] = collect(lines, start, end)
                    kind, req, resp = cache[start]
                    kind = norm_server(kind)
                    owner = typ or stem
                    if cm:
                        self.emit(cm.group(1), owner, meth, rel, i + 1, kind, req, resp)
                    else:
                        base = dm.group(1) if dm.group(1).endswith('.') else dm.group(1) + '.'
                        hits = (expand_dynamic(lines, meth, start, end)
                                or expand_dynamic(lines, meth, 0, len(lines)))
                        for lit, ln in hits:
                            self.emit(base + lit, owner, meth, rel, ln, kind, req, resp, dyn=base)

    def write(self):
        self.rows.sort(key=lambda r: (r['command'], r['source']))
        with io.open(OUT, 'w', encoding='utf-8-sig', newline='') as fh:
            w = csv.DictWriter(fh, fieldnames=['command', 'server', 'owner_type', 'method',
                                               'source', 'line', 'request_keys',
                                               'response_keys', 'dynamic_prefix'])
            w.writeheader()
            w.writerows(self.rows)


if __name__ == '__main__':
    inv = Inventory()
    inv.scan()
    inv.write()
    r = inv.rows
    print('rows:', len(r), '| unique commands:', len({x['command'] for x in r}),
          '| with request keys:', sum(1 for x in r if x['request_keys']),
          '| with response keys:', sum(1 for x in r if x['response_keys']),
          '| dynamic expanded:', sum(1 for x in r if x['dynamic_prefix']))
