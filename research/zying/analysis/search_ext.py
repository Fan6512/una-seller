import os, re, json, sys

root = r"E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0/assets/js"
terms = ["批量关闭", "关闭活动", "结束活动", "终止活动", "批量结束", "停止活动", "活动关闭",
         "seller-promotions", "promotions", "promociones", "campaign", "cerrar", "finalizar",
         "closeCampaign", "endCampaign", "close_campaign", "end_campaign"]
out = {}
for dirpath, _, files in os.walk(root):
    for fn in files:
        if not fn.endswith(".js"): continue
        p = os.path.join(dirpath, fn)
        try:
            s = open(p, encoding="utf-8", errors="replace").read()
        except Exception as e:
            continue
        low = s.lower()
        hits = []
        for t in terms:
            tl = t.lower()
            start = 0
            cnt = 0
            while True:
                i = low.find(tl, start)
                if i < 0: break
                cnt += 1
                start = i + len(tl)
                if cnt <= 3:
                    hits.append({"term": t, "pos": i, "ctx": s[max(0,i-180):i+220].replace("\n"," ")})
            if cnt:
                hits.append({"term": t, "count": cnt})
        if hits:
            rel = os.path.relpath(p, root)
            out[rel] = hits
open(r"E:/美客多逆向/analysis/ext_campaign_hits.json","w",encoding="utf-8").write(json.dumps(out, ensure_ascii=False, indent=1))
for f, hits in sorted(out.items()):
    print("==", f)
    for h in hits:
        if "count" in h:
            print("   ", h["term"], "TOTAL", h["count"])
        else:
            print("   ", h["term"], "@", h["pos"], ":", h["ctx"][:360])
