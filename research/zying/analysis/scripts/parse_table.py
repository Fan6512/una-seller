import struct, json, os, io, re

outdir = r"E:/美客多逆向/analysis/decrypted"
tb = open(os.path.join(outdir, "string_table.bin"), "rb").read()
entries = []
pos = 0
n = len(tb)
while pos + 4 <= n:
    l = struct.unpack_from("<i", tb, pos)[0]
    if l < 0 or l % 2 != 0 or pos + 4 + l > n:
        print("bad entry at", pos, "len", l); break
    try:
        s = tb[pos+4:pos+4+l].decode("utf-16le")
    except Exception as e:
        print("decode fail at", pos, e); break
    entries.append({"off": pos, "len": l, "s": s})
    pos += 4 + l
print("entries:", len(entries), "consumed:", pos, "of", n)
json.dump(entries, open(os.path.join(outdir, "strings.json"), "w", encoding="utf-8"), ensure_ascii=False, indent=0)
with io.open(os.path.join(outdir, "all_strings.txt"), "w", encoding="utf-8") as f:
    for e in entries:
        f.write("%d\t%s\n" % (e["off"], e["s"].replace("\n", "\n").replace("\r", "\r")))
print("sample:")
for e in entries[:8]:
    print(e["off"], repr(e["s"][:100]))
