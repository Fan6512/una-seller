from pathlib import Path

src = Path(r"E:\美客多逆向\analysis\work-linemeli.gbk")
dst = Path(r"E:\美客多逆向\analysis\decrypted\Work_MLimit_LineMeli.cs")
dst.write_text(src.read_bytes().decode("gbk", "replace"), encoding="utf-8")
print(dst, dst.stat().st_size)
