import struct, gzip, os, sys

TARGET = r"E:/美客多逆向/target/ZYing"
OUT = r"E:/美客多逆向/analysis/dat"
os.makedirs(OUT, exist_ok=True)

def load(path):
    data = open(path, "rb").read()
    ver, hlen = struct.unpack_from("<hh", data, 0)
    hdr = data[4:4+hlen]
    p = 0
    heads = []
    while p < len(hdr):
        nl = hdr[p]; p += 1
        name = hdr[p:p+nl].decode("utf-8", "replace"); p += nl
        start, length = struct.unpack_from("<ii", hdr, p); p += 8
        heads.append((name, start, length))
    return ver, heads, data

for fn in ["ZData.dat","ZOther.dat","ZAmazon.dat","ZEbay.dat","ZForm.dat","remotes.dat","ZCaptcha.dat"]:
    path = os.path.join(TARGET, fn)
    if not os.path.exists(path): continue
    try:
        ver, heads, data = load(path)
    except Exception as e:
        print(fn, "header parse failed:", e); continue
    print(f"== {fn} ver={ver} entries={len(heads)} size={len(data)}")
    shown = 0
    for name, start, length in heads:
        ok = data[start:start+2] == b"\x1f\x8b"
        tag = "" if ok else " NOTGZIP"
        if shown < 40:
            print(f"   {name:<48} {length:>8}{tag}")
            shown += 1
    if len(heads) > 40: print(f"   ... {len(heads)-40} more")
