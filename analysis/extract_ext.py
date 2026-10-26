import os, re, json
root = r"E:/美客多逆向/analysis/zying_ext/extension/5.1.25_0/assets/js"
s = open(os.path.join(root,"content.js"), encoding="utf-8", errors="replace").read()
outdir = r"E:/美客多逆向/analysis/ext_slices"
os.makedirs(outdir, exist_ok=True)
def dump(name, start, end):
    seg = s[start:end]
    open(os.path.join(outdir,name),"w",encoding="utf-8").write(seg)
    print(name, start, end, len(seg))
# main feature region
dump("promo_700k_780k.js", 698000, 785000)
# earlier selectors region
dump("promo_595k_640k.js", 595000, 645000)
# stop activity region
dump("promo_735k_760k.js", 735000, 760000)
# campaign url region
dump("campaign_url_1560k_1575k.js", 1560000, 1575000)
# search for api endpoints
for pat in ["api.mercadolibre.com", "api.mercadolivre.com", "mercadolibre.com/", "fetch(", "XMLHttpRequest", "chrome.runtime.sendMessage", "zying.net", "api.zying"]:
    idxs = [m.start() for m in re.finditer(re.escape(pat), s)]
    print(pat, len(idxs), idxs[:10])
