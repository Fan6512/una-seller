"""Extract the ZYing menu tree (platform section -> group -> feature entry) from the resolved menu source.

Scope matters: the same local variable names (eItemCollection, ...4) are reused in
every menu-building method, so entries and group labels must be matched inside one
method body only. Each entry is (form-key, display-label, optional gating).
Emits keys, labels and gating identities only.
"""
import io
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(ROOT, 'decrypted', 'ZYing_Menus.resolved.cs')
OUT = os.path.join(ROOT, 'menu-tree.json')

OBF = r'(?:(?:[\w$]+\.|[\w$]+\([^)]*\)\s*\.)?\s*(?:/\*\d+\*/)?\s*)'
LIT = r'"((?:[^"\\]|\\.)*)"'
METHOD_RE = re.compile(r'static\s+EItemCollection\s+([A-Za-z_]\w*)')
ADD_RE = re.compile(r'(eItemCollection\d*)\.Add\(\s*new EItem\(\s*' + OBF + LIT + r'\s*,\s*' + OBF + LIT)
BUILD_RE = re.compile(r'(_[A-Za-z_]\w*)\s*=\s*new EItemCollection\(([^;]{0,6000})\)')
ITEM3_RE = re.compile(r'new EItem\(\s*' + OBF + LIT + r'\s*,\s*' + OBF + LIT + r'\s*,\s*(eItemCollection\d*)\)')
GATE_RE = re.compile(r'Member\.UserID\.In\([\d,\s]+\)|IsVip|IsTest')


def load(path):
    raw = io.open(path, 'rb').read()
    try:
        return raw.decode('utf-8')
    except UnicodeDecodeError:
        return raw.decode('gbk', 'replace')


def main():
    text = load(SRC)
    marks = [(m.group(1), m.start()) for m in METHOD_RE.finditer(text)]
    marks.append(('<end>', len(text)))

    tree = {}
    for (name, start), (_n2, stop) in zip(marks, marks[1:]):
        body = text[start:stop]
        adds = [(m.group(1), m.group(2), m.group(3), m.start()) for m in ADD_RE.finditer(body)]
        builds = [(m.group(2), m.start()) for m in BUILD_RE.finditer(body)]
        for blob, bpos in builds:
            for gm in ITEM3_RE.finditer(blob):
                label, coll = gm.group(1), gm.group(3)
                # entries assigned to this collection are those added before this build
                items = []
                for c, key, label2, apos in adds:
                    if c != coll or apos > bpos:
                        continue
                    if key in {i['key'] for i in items}:
                        continue
                    gate = ''
                    gpos = next((bp for _b, bp in builds if bp > apos), None)
                    near = body[max(0, apos - 240):apos]
                    if gpos is not None and near.count('if (') > near.count('}'):
                        g = GATE_RE.findall(near)
                        gate = g[-1] if g else 'conditional'
                    items.append({'key': key, 'label': label2, 'gate': gate})
                if items:
                    bucket = tree.setdefault(name, {}).setdefault(label, {})
                    bucket.update({i['key']: i for i in items})
        for c, key, label, _p in adds:
            if c not in {m.group(3) for blob, _ in builds for m in ITEM3_RE.finditer(blob)}:
                bucket = tree.setdefault(name, {}).setdefault('(ungrouped)', {})
                bucket.setdefault(key, {'key': key, 'label': label, 'gate': ''})

    out = {sec: {grp: list(items.values()) for grp, items in sorted(groups.items())}
           for sec, groups in sorted(tree.items())}
    json.dump(out, io.open(OUT, 'w', encoding='utf-8'), ensure_ascii=False, indent=1)
    print('sections:', len(out))
    for sec in out:
        n = sum(len(v) for v in out[sec].values())
        print(f'  {sec:16s} groups={len(out[sec]):2d} entries={n}')


if __name__ == '__main__':
    sys.exit(main())
