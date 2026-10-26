import re, io, json

pat = re.compile(r'/\*(\d+)\*/"((?:[^"\\]|\\.)*)"')
s = io.open(r"E:/美客多逆向/analysis/decrypted/MeliApi.resolved.cs", encoding="utf-8").read()
urls = []
for off, v in pat.findall(s):
    if v.startswith("/") or v.startswith("https://api.mercadolibre.com"):
        urls.append((int(off), v))
seen = set()
out = []
for off, v in sorted(urls):
    if v in seen:
        continue
    seen.add(v)
    out.append(v)
io.open(r"E:/美客多逆向/analysis/analysis-meli-endpoints.txt", "w", encoding="utf-8").write("\n".join(out))
print(len(out))
