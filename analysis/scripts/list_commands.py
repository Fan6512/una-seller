import re, glob, io, collections

pat = re.compile(r'/\*(\d+)\*/"((?:[^"\\]|\\.)*)"')
cmds = collections.Counter()
for f in glob.glob(r"E:/美客多逆向/analysis/decrypted/user/*.cs"):
    s = io.open(f, encoding="utf-8", errors="replace").read()
    for off, val in pat.findall(s):
        try:
            v = bytes(val, "utf-8").decode("unicode_escape")
        except Exception:
            v = val
        if re.match(r'^[a-zA-Z][a-zA-Z0-9_]*\.[a-zA-Z0-9_]+$', v):
            cmds[v] += 1
print("distinct commands:", len(cmds))
for v, c in cmds.most_common(300):
    if v.startswith("meli"):
        print(c, v)
