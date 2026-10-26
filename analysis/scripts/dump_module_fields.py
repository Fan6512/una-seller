import os, json, clr
TARGET = r"E:/美客多逆向/target/ZYing"
out = r"E:/美客多逆向/analysis/decrypted/module_fields.json"
from System import AppDomain, ResolveEventHandler, Object, Type
from System.Reflection import Assembly, BindingFlags

def resolve(sender, args):
    name = args.Name.split(',')[0]
    p = os.path.join(TARGET, name + ".dll")
    return Assembly.LoadFrom(p) if os.path.exists(p) else None

AppDomain.CurrentDomain.AssemblyResolve += ResolveEventHandler(resolve)
asm = Assembly.LoadFrom(os.path.join(TARGET, "ZYing.Work.dll"))
BF = BindingFlags
name = "<Module>{0e325ee4-ecc9-4c54-83c5-9fd3228c8105}"
t = asm.GetType(name)
print("type:", t)
inst = t.GetField("m_5b4ad7c38cba4ec7baf5f148600dc375", BF.Static | BF.NonPublic).GetValue(Object())
vals = {}
for f in t.GetFields(BF.Instance | BF.NonPublic | BF.Public):
    if f.Name.startswith("m_"):
        vals[f.Name] = int(f.GetValue(inst))
json.dump(vals, open(out, "w"), indent=1)
print("fields:", len(vals))
print(list(vals.items())[:3])
