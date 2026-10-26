from pathlib import Path

base = Path(r"E:\美客多逆向\analysis")
for name in ["MLogin", "EGLogin"]:
    src = base / (name + ".gbk")
    dst = base / "decrypted" / (name + ".resolved.cs")
    dst.write_text(src.read_bytes().decode("gbk", "replace"), encoding="utf-8")
    print(dst, dst.stat().st_size)
