import os, clr, sys
TARGET = r"E:/美客多逆向/target/ZYing"
from System import AppDomain, ResolveEventHandler
from System.Reflection import Assembly
def resolve(sender,args):
    n=args.Name.split(',')[0]; p=os.path.join(TARGET,n+".dll")
    return Assembly.LoadFrom(p) if os.path.exists(p) else None
AppDomain.CurrentDomain.AssemblyResolve += ResolveEventHandler(resolve)
for dll in ["ZYing.User.dll","ZYing.Work.dll","ZYing.Core.dll"]:
    a=Assembly.LoadFrom(os.path.join(TARGET,dll))
    print("==",dll)
    for r in a.GetManifestResourceNames():
        print("  ",r)
