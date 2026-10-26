import json
import re

base = r"E:/美客多逆向/analysis"
entries = json.load(open(base + "/decrypted/strings_user.json", encoding="utf-8"))
sm = {e["off"]: e["s"] for e in entries}
fields = json.load(open(base + "/decrypted/module_fields_user.json"))
call_re = re.compile(r"EriwwTjw3\(([^()]*)\)")
mod_re = re.compile(r"<Module>\{[^}]*\}\.m_[0-9a-f]+\.(m_[0-9a-f]+)")


def eval_lit(expr):
    expr = expr.strip()
    if not expr:
        return 0
    if expr.endswith("^"):
        expr = expr[:-1]
    expr = expr.replace("/", "//")
    return eval(expr, {"__builtins__": {}}, {}) & 0xFFFFFFFF


def resolve_text(text):
    out = []
    for line in text.splitlines(True):
        if "EriwwTjw3(" not in line:
            out.append(line)
            continue

        def rep(m):
            inner = m.group(1)
            mm = mod_re.search(inner)
            if not mm:
                return m.group(0)
            fv = fields.get(mm.group(1))
            if fv is None:
                return m.group(0)
            off = (eval_lit(inner[:mm.start()]) ^ (fv & 0xFFFFFFFF)) & 0xFFFFFFFF
            s = sm.get(off)
            if s is None:
                return m.group(0)
            return '/*%d*/"%s"' % (off, s.replace("\\", "\\\\").replace('"', '\\"').replace("\n", "\\n"))

        out.append(call_re.sub(rep, line))
    return "".join(out)


for name in ["MLogin", "EGLogin"]:
    src = base + "/decrypted/" + name + ".resolved.cs"
    txt = resolve_text(open(src, encoding="utf-8").read())
    open(src, "w", encoding="utf-8").write(txt)
    print(name, "done")
