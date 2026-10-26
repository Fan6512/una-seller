import sqlite3
from pathlib import Path

p = Path(r"E:\美客多逆向\target\ZYing\cache\temp\Cookies")
con = sqlite3.connect(f"file:{p}?mode=ro", uri=True)
cur = con.cursor()
print(cur.execute("SELECT name FROM sqlite_master WHERE type='table'").fetchall())
try:
    rows = cur.execute("SELECT host_key, COUNT(*), MIN(expires_utc), MAX(expires_utc) FROM cookies GROUP BY host_key ORDER BY 2 DESC LIMIT 40").fetchall()
    for r in rows:
        print(r)
except Exception as e:
    print("ERR", e)
con.close()
