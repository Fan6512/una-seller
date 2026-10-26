import os, json
import clr
TARGET = r"E:/美客多逆向/target/ZYing"
outdir = r"E:/美客多逆向/analysis/decrypted"
from System import AppDomain, ResolveEventHandler, Object
from System.Reflection import Assembly, BindingFlags

def resolve(s, a):
    n = a.Name.split(',')[0]
    p = os.path.join(TARGET, n + ".dll")
    if os.path.exists(p):
        try:
            return Assembly.LoadFrom(p)
        except Exception:
            return None
    return None

AppDomain.CurrentDomain.AssemblyResolve += ResolveEventHandler(resolve)
asm = Assembly.LoadFrom(os.path.join(TARGET, "ZYing.User.dll"))
BF = BindingFlags
t = asm.GetType("<Module>{51cba30a-1899-472d-b922-810a0384b6c4}")
holder = t.GetField("m_3803de857ccf44d38d28afc17036c502", BF.Static | BF.NonPublic)
inst = holder.GetValue(Object())
vals = {}
for f in t.GetFields(BF.Instance | BF.NonPublic | BF.Public):
    if f.Name.startswith("m_"):
        vals[f.Name] = int(f.GetValue(inst))
json.dump(vals, open(os.path.join(outdir, "module_fields_user.json"), "w"), indent=1)
print("fields:", len(vals))
