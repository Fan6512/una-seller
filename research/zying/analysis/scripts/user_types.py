import os, json, clr
TARGET = r"E:/美客多逆向/target/ZYing"
from System import AppDomain, ResolveEventHandler, Object
from System.Reflection import Assembly, BindingFlags, ReflectionTypeLoadException

def resolve(sender, args):
    name = args.Name.split(',')[0]
    p = os.path.join(TARGET, name + ".dll")
    return Assembly.LoadFrom(p) if os.path.exists(p) else None

AppDomain.CurrentDomain.AssemblyResolve += ResolveEventHandler(resolve)
asm = Assembly.LoadFrom(os.path.join(TARGET, "ZYing.User.dll"))
types = []
try:
    types = list(asm.GetTypes())
except ReflectionTypeLoadException as e:
    types = [t for t in e.Types if t is not None]
    print("partial load:", len(types))
names = []
for t in types:
    try:
        names.append((t.FullName or "", t.Assembly.GetName().Name))
    except Exception:
        pass
json.dump(names, open(r"E:/美客多逆向/analysis/user-type-names.json", "w", encoding="utf-8"), ensure_ascii=False, indent=0)
print("types:", len(names))
# check decryptor type assembly
t = asm.GetType("NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs")
print("decryptor resolved:", t)
if t: print("decryptor assembly:", t.Assembly.GetName().Name)
