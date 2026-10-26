"""Harvest observable field/schema identifiers per UI module.

The client keeps no local relational database: grids are server RecordSet
objects, so the practical data model is the set of field names the UI indexes
or reads back. This walks the resolved sources and groups those names by the
module that touches them, which feeds data-model.md.
Emits identifier lists and counts only.
"""
import io
import json
import os
import re
import sys
from collections import Counter, defaultdict

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(ROOT, 'decrypted')
OUT = os.path.join(ROOT, 'field-index.json')

IDX_RE = re.compile(r'\[\s*(?:(?:[\w$]+\.|[\w$]+\([^)]*\)\s*\.)?\s*(?:/\*\d+\*/)?\s*)?"([a-zA-Z_][\w.]{1,40})"\s*\]')
COL_RE = re.compile(r'(?:Columns|Fields)\[\s*(?:[\w$]+\.|)\s*(?:/\*\d+\*/)?\s*"([a-zA-Z_][\w.]{1,40})"')


def load(path):
    raw = io.open(path, 'rb').read()
    try:
        return raw.decode('utf-8')
    except UnicodeDecodeError:
        return raw.decode('gbk', 'replace')


CLS_RE = re.compile(r'^\s*(?:public|private|internal|protected|\s)+'
                    r'(?:static\s+|sealed\s+|partial\s+)*class\s+([A-Za-z_][\w<>]*)', re.M)


def module_of(text, rel):
    """Group by the declared type name so twin dumps of one class merge."""
    m = CLS_RE.search(text)
    if m:
        return m.group(1)
    stem = os.path.basename(rel)[:-3].replace('.resolved', '')
    return stem[len('ZYing.'):] if stem.startswith('ZYing.') else stem


def main():
    per_module = defaultdict(Counter)
    for dirpath, _dn, filenames in os.walk(SRC):
        for fn in sorted(filenames):
            if not fn.endswith('.cs'):
                continue
            path = os.path.join(dirpath, fn)
            rel = os.path.relpath(path, SRC).replace('\\', '/')
            text = load(path)
            keys = set()
            for rx in (IDX_RE, COL_RE):
                for m in rx.finditer(text):
                    keys.add(m.group(1))
            mod = module_of(text, rel)
            for k in keys:
                per_module[mod][k] += 1

    out = {mod: dict(sorted(c.items(), key=lambda kv: -kv[1])) for mod, c in sorted(per_module.items())}
    json.dump(out, io.open(OUT, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
    total = sum(len(v) for v in out.values())
    print('modules:', len(out), '| distinct (module,field) pairs:', total)
    for mod in sorted(out, key=lambda m: -len(out[m]))[:12]:
        print(f'{len(out[mod]):4d}  {mod}')


if __name__ == '__main__':
    sys.exit(main())
