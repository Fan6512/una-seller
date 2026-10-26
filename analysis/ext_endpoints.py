import os, re, json
root = r"E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0/assets/js"
report = []
for fn in ["content.js","background.js","chunk-common.js","inject-meli.js","popup.js"]:
    p = os.path.join(root, fn)
    s = open(p, encoding="utf-8", errors="replace").read()
    urls = sorted(set(re.findall(r'https?://[A-Za-z0-9._~:/?#\[\]@!$&\'()*+,;=%-]{4,160}', s)))
    paths = sorted(set(re.findall(r'["\'`](/(?:listings|my|orders|messages|questions|claims|post-purchase|seller|api|campaigns|promotions)[A-Za-z0-9._~:/?#\[\]@!$&\'()*+,;=%-]{0,140})["\'`]', s)))
    fetches = []
    for m in re.finditer(r'fetch\(', s):
        fetches.append(s[max(0,m.start()-80):m.start()+220].replace("\n"," "))
    report.append({"file":fn,"urls":urls,"internal_paths":paths,"fetch_samples":fetches[:40],"fetch_count":len(fetches)})
open(r"E:/美客多逆向/analysis/ext_endpoints.json","w",encoding="utf-8").write(json.dumps(report, ensure_ascii=False, indent=1))
for r in report:
    print("==", r["file"], "fetch", r["fetch_count"], "urls", len(r["urls"]), "paths", len(r["internal_paths"]))
