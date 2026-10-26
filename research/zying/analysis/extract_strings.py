import re
from pathlib import Path

ROOT = Path(r"E:\美客多逆向\target\ZYing")
OUT = Path(r"E:\美客多逆向\analysis\strings")
OUT.mkdir(parents=True, exist_ok=True)

FILES = [
    "ZAmazon.dat", "ZEbay.dat", "ZCaptcha.dat", "ZData.dat",
    "ZForm.dat", "ZOther.dat", "remotes.dat",
    "ZYing.Core.dll", "ZYing.Work.dll", "ZYing.User.dll",
    "ZYing.UI.dll", "ZYing.dll", "ZDto.dll", "ZPdf.dll",
]

ASCII_RE = re.compile(rb"[\x20-\x7e]{4,}")
UTF16_RE = re.compile(rb"(?:[\x20-\x7e]\x00){4,}")


def emit(path: Path, data: bytes) -> int:
    items = []
    for m in ASCII_RE.finditer(data):
        s = m.group().decode("ascii", "replace")[:500]
        items.append((m.start(), "A", s))
    for m in UTF16_RE.finditer(data):
        s = m.group().decode("utf-16le", "replace")[:500]
        if len(s.strip()) >= 4:
            items.append((m.start(), "U", s))
    items.sort(key=lambda x: x[0])
    with path.open("w", encoding="utf-8", errors="replace") as f:
        for off, kind, s in items:
            f.write(f"{off:08x} {kind} {s}\n")
    return len(items)


for name in FILES:
    p = ROOT / name
    if not p.exists():
        print(f"MISSING {name}")
        continue
    out = OUT / (name + ".strings.txt")
    n = emit(out, p.read_bytes())
    print(f"{name}\t{n}\t{out.stat().st_size}")
