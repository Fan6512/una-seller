import json, os, re, subprocess, sys, hashlib

base = r"E:/美客多逆向/analysis"
target = r"E:/美客多逆向/target/ZYing/ZYing.User.dll"
ilspy = os.path.join(base, "tools/ilspycmd/tools/net10.0/any/ilspycmd.dll")
outdir = os.path.join(base, "decompiled/user")
os.makedirs(outdir, exist_ok=True)
names = json.load(open(os.path.join(base, "user-meli-toplevel.json"), encoding="utf-8"))
log = open(os.path.join(base, "decompiled/user/_decompile.log"), "w", encoding="utf-8")

for i, name in enumerate(names):
    safe = re.sub(r'[\\/:*?"<>|]', "_", name)
    out = os.path.join(outdir, safe + ".cs")
    if os.path.exists(out) and os.path.getsize(out) > 0:
        continue
    p = subprocess.run(["dotnet", ilspy, "-t", name, target],
                       capture_output=True, timeout=180)
    data = p.stdout
    # ilspycmd writes console output in the OEM/ANSI codepage
    if data and b"\xef\xbb\xbf" != data[:3]:
        try:
            text = data.decode("gbk")
        except UnicodeDecodeError:
            text = data.decode("utf-8", errors="replace")
    else:
        text = data.decode("utf-8-sig", errors="replace")
    if p.returncode != 0 or not text.strip():
        log.write("FAIL %s rc=%s %s\n" % (name, p.returncode, p.stderr.decode("gbk", "replace")[:300]))
        log.flush()
        continue
    open(out, "w", encoding="utf-8").write(text)
    log.write("OK %s -> %s (%d)\n" % (name, safe, len(text)))
    log.flush()
    print(i + 1, "/", len(names), name, len(text))
log.close()
print("done")
