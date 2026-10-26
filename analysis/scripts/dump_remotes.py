import gzip
import struct
from pathlib import Path

path = Path(r"E:\美客多逆向\target\ZYing\remotes.dat")
data = path.read_bytes()
ver, hlen = struct.unpack_from("<hh", data, 0)
hdr = data[4:4 + hlen]
p = 0
heads = []
while p < len(hdr):
    nl = hdr[p]; p += 1
    name = hdr[p:p + nl].decode("utf-8", "replace"); p += nl
    start, length = struct.unpack_from("<ii", hdr, p); p += 8
    heads.append((name, start, length))
for name, start, length in heads:
    raw = data[start:start + length]
    if raw[:2] == b"\x1f\x8b":
        raw = gzip.decompress(raw)
    out = Path(r"E:\美客多逆向\analysis\dat") / (name + ".bin")
    out.write_bytes(raw)
    print(name, len(raw), raw[:300])
