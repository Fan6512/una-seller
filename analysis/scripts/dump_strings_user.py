import os, sys, json, struct
import clr

TARGET = r"E:/美客多逆向/target/ZYing"
outdir = r"E:/美客多逆向/analysis/decrypted"
asm_path = os.path.join(TARGET, "ZYing.User.dll")

from System import AppDomain, ResolveEventHandler, Object, Int32
from System.Reflection import Assembly, BindingFlags

def resolve(sender, args):
    name = args.Name.split(',')[0]
    p = os.path.join(TARGET, name + ".dll")
    if os.path.exists(p):
        try:
            return Assembly.LoadFrom(p)
        except Exception:
            return None
    return None

AppDomain.CurrentDomain.AssemblyResolve += ResolveEventHandler(resolve)

asm = Assembly.LoadFrom(asm_path)
print("loaded:", asm.FullName)

t = asm.GetType("NZsfdnHVTHmTqMuAXn.kjtH3rRSqYWvB6R6Hs")
print("type:", t)
BF = BindingFlags
meth = t.GetMethod("EriwwTjw3", BF.Static | BF.NonPublic)
s0 = meth.Invoke(Object(), [Int32(0)])
print("s0:", repr(str(s0)))

field = t.GetField("LmFnbT94X", BF.Static | BF.NonPublic)
tb = bytes(field.GetValue(Object()))
orig_len = len(str(s0).encode("utf-16le"))
tb = struct.pack("<i", orig_len) + tb[4:]
open(os.path.join(outdir, "string_table_user.bin"), "wb").write(tb)
print("table bytes:", len(tb))

# User module type fields
try:
    types = list(asm.GetTypes())
except Exception as e:
    types = [x for x in e.Types if x is not None]
modt = None
for tt in types:
    try:
        f = tt.GetField("m_5b4ad7c38cba4ec7baf5f148600dc375", BF.Static | BF.NonPublic)
        if f is not None:
            modt = tt
            break
    except Exception:
        pass
print("module type:", modt)
if modt is not None:
    inst = modt.GetField("m_5b4ad7c38cba4ec7baf5f148600dc375", BF.Static | BF.NonPublic).GetValue(Object())
    vals = {}
    for f in modt.GetFields(BF.Instance | BF.NonPublic | BF.Public):
        if f.Name.startswith("m_"):
            vals[f.Name] = int(f.GetValue(inst))
    json.dump(vals, open(os.path.join(outdir, "module_fields_user.json"), "w"), indent=1)
    print("fields:", len(vals))
