import re, json, io, os

SRC = 'zying_ext/extension/5.1.25_0/assets/js/chunk-common.js'
s = open(SRC, encoding='utf-8').read()
start = s.index('let o={deepseek:')
end = s.index('protocol:"openai"}},i=', start) + len('protocol:"openai"')
seg = s[start:end]
pat = re.compile(r'([a-z]+):\{name:"(.*?)"(?:,baseURL:"(.*?)")?(?:,apiKeyUrl:"(.*?)")?(?:,defaultModel:"(.*?)")?(?:,models:\[(.*?)\])?(?:,protocol:"(.*?)")?\}')
rows = []
for m in pat.finditer(seg):
    rows.append({
        "provider": m.group(1),
        "display_name": m.group(2),
        "base_url": m.group(3) or "",
        "api_key_portal": m.group(4) or "",
        "default_model": m.group(5) or "",
        "model_list": [x.strip('"') for x in (m.group(6) or "").split(',') if x.strip('"')],
        "protocol": m.group(7) or "",
    })
json.dump(rows, io.open('ai_providers.json', 'w', encoding='utf-8'), ensure_ascii=False, indent=2)
for r in rows:
    print(r["provider"], "|", r["base_url"], "|", r["default_model"], "|", ",".join(r["model_list"]))
