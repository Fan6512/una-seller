import json
import re

p = r"C:\Users\Administrator\.qoder\projects\E-------\54c7135f-df13-42e1-bbf6-5f0d44a86b11.jsonl"
pat = re.compile(r"ev_[0-9a-f]{64}")
name_pat = re.compile(r'"name"\s*:\s*"([^"]+)"')
asm_pat = re.compile(r"assembly:\s*([^\n\\]+)")
seen = {}


def walk(x):
    if isinstance(x, dict):
        for v in x.values():
            walk(v)
    elif isinstance(x, list):
        for v in x:
            walk(v)
    elif isinstance(x, str):
        for m in pat.finditer(x):
            ev = m.group(0)
            if ev in seen:
                continue
            tail = x[m.end():m.end() + 3000]
            name = name_pat.search(tail)
            asm = asm_pat.search(tail)
            seen[ev] = (
                name.group(1) if name else "",
                asm.group(1).strip() if asm else "",
                tail[:300].replace("\\n", " "),
            )


for line in open(p, encoding="utf-8", errors="replace"):
    if "ev_" not in line:
        continue
    try:
        o = json.loads(line)
    except Exception:
        continue
    walk(o)

print(len(seen))
with open(r"E:\美客多逆向\analysis\evidence_map.txt", "w", encoding="utf-8") as f:
    for ev, (n, a, c) in seen.items():
        line = f"{ev}\n  name={n}\n  asm={a}\n  ctx={c}\n\n"
        f.write(line)
        print(ev, "|", n, "|", a)
