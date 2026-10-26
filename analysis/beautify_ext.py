import jsbeautifier, os
indir = r"E:/美客多逆向/analysis/ext_slices"
opts = jsbeautifier.default_options()
opts.indent_size = 1
opts.max_preserve_newlines = 2
opts.wrap_line_length = 0
for fn in os.listdir(indir):
    if not fn.endswith(".js"): continue
    s = open(os.path.join(indir,fn), encoding="utf-8").read()
    try:
        b = jsbeautifier.beautify(s, opts)
    except Exception as e:
        print("FAIL", fn, e); continue
    out = os.path.join(indir, fn.replace(".js",".pretty.js"))
    open(out,"w",encoding="utf-8").write(b)
    print(fn, len(s), "->", len(b), "lines", b.count("\n"))
