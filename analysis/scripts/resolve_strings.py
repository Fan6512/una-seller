import json, re, os, io

base = r"E:/美客多逆向/analysis"
entries = json.load(open(base + "/decrypted/strings.json", encoding="utf-8"))
sm = {e["off"]: e["s"] for e in entries}
fields = json.load(open(base + "/decrypted/module_fields.json"))

call_re = re.compile(r"EriwwTjw3\(([^()]*)\)")
mod_re = re.compile(r"<Module>\{[^}]*\}\.m_[0-9a-f]+\.(m_[0-9a-f]+)")

def eval_lit(expr):
    expr = expr.strip()
    if not expr:
        return 0
    if expr.endswith("^"):
        expr = expr[:-1]
    expr = expr.replace("/", "//")
    v = eval(expr, {"__builtins__": {}}, {})
    return v & 0xFFFFFFFF

def csharp_str(s):
    out = s.replace("\\", "\\\\").replace('"', '\\"').replace("\r", "\r").replace("\n", "\n").replace("\t", "\t")
    return '"' + out + '"'

def resolve_file(src, dst):
    lines = io.open(src, encoding="utf-8-sig", errors="replace").read().splitlines(True)
    out = []
    mapping = []
    missing = 0
    for i, line in enumerate(lines, 1):
        if "EriwwTjw3(" not in line:
            out.append(line); continue
        def rep(m):
            nonlocal missing
            inner = m.group(1)
            mm = mod_re.search(inner)
            if not mm:
                return m.group(0)
            fname = mm.group(1)
            fv = fields.get(fname)
            if fv is None:
                return m.group(0)
            lit = inner[:mm.start()]
            off = (eval_lit(lit) ^ (fv & 0xFFFFFFFF)) & 0xFFFFFFFF
            s = sm.get(off)
            if s is None:
                missing += 1
                return "/*MISS off=%d*/%s" % (off, m.group(0))
            mapping.append((i, off, s))
            return "/*%d*/%s" % (off, csharp_str(s))
        nl = call_re.sub(rep, line)
        out.append(nl)
    io.open(dst, "w", encoding="utf-8").write("".join(out))
    print(os.path.basename(src), "resolved:", len(mapping), "missing:", missing)

for name in ["MeliApi.cs", "MeliAppInfo.cs", "MeliConsts.cs", "MeliApiBuilder.cs", "ZYing_ZClient.cs", "ZYing_Member.work.cs", "ZYing_Meli_MeliErrorParser.cs", "AmazonService_ZCaptcha.cs"]:
    src = base + "/decompiled/" + name
    if os.path.exists(src):
        resolve_file(src, base + "/decrypted/" + name.replace(".cs", ".resolved.cs"))
