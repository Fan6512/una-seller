import struct, sys, binascii

M = 0xFFFFFFFF

def flsFiRVeO(P0):
    num = P0 & M
    num2 = 973202305
    num3 = 1582787682
    num4 = 1577548636
    num5 = 332884210
    num6 = (num4 * 1313243236) & 0xFFFFFFFFFFFFFFFF
    num6 |= 1
    num3 = ((num3 * num3) & 0xFFFFFFFF) % num6 & M
    num7 = (1907532890 * num4) & 0xFFFFFFFFFFFFFFFF
    if num7 == 0:
        num7 = 0xFFFFFFFFFFFFFFFF
    _ = 698203908 % num7
    num2 = (-502326134 - num3) & M
    num8 = (num3 * 183835789) & 0xFFFFFFFFFFFFFFFF
    num8 |= 1
    num4 = ((num4 * num4) & 0xFFFFFFFF) % num8 & M
    num9 = (((num5 >> 6) | ((num5 << 26) & M)) ^ num3) & M
    num10 = num9 & 0x0F0F0F0F
    num5 = (((num9 & 0xF0F0F0F0) >> 4) | ((num10 << 4) & M)) & M
    num = (num ^ ((num << 3) & M)) & M
    num = (num + num2) & M
    num = (num ^ ((num << 11) & M)) & M
    num = (num + num4) & M
    num = (num ^ (num >> 27)) & M
    num = (num + num5) & M
    return ((((num2 << 11) & M) - num3) ^ num4) - num & M

def decrypt(data, key):
    n = len(data)
    if n == 0: return b""
    rem = n % 4
    chunks = n // 4 + (1 if rem else 0)
    out = bytearray(n)
    kwords = len(key) // 4
    acc = 0
    for i in range(chunks):
        kw = i % kwords
        off = kw * 4
        kwv = (key[off+3] << 24) | (key[off+2] << 16) | (key[off+1] << 8) | key[off]
        if i == chunks - 1 and rem:
            n6 = 0
            for j in range(rem):
                if j > 0: n6 <<= 8
                n6 |= data[n - 1 - j]
            acc = (acc + kwv) & M
            acc = (acc + flsFiRVeO(acc)) & M
            x = (acc ^ n6) & M
            base = i * 4
            for k in range(rem):
                out[base + k] = (x >> (8 * k)) & 0xFF
        else:
            base = i * 4
            w = (data[base+3] << 24) | (data[base+2] << 16) | (data[base+1] << 8) | data[base]
            acc = (acc + kwv) & M
            acc = (acc + flsFiRVeO(acc)) & M
            x = (acc ^ w) & M
            out[base:base+4] = struct.pack('<I', x)
    return bytes(out)

KEY_FIXED = bytes([123,5,74,12,244,156,221,154,121,221,183,41,121,65,9,43,67,81,23,43,74,63,64,23,95,185,226,244,45,194,211,43])
IV_FIXED = bytes([117,254,41,121,65,52,9,43,221,154,12,54,68,241,68,66])

blob = open(r"E:/美客多逆向/analysis/decompiled/ZYing.Work-blob.bin","rb").read()
for name, key in [("l54SyBZvD-fixed", KEY_FIXED)]:
    out = decrypt(blob, key)
    l = struct.unpack_from('<i', out, 0)[0]
    print(name, "first len:", l)
    print("first 80 bytes:", binascii.hexlify(out[:80]).decode())
    if 0 < l < 10000 and l + 4 <= len(out):
        try:
            s = out[4:4+l].decode('utf-16le')
            print("first string:", repr(s[:200]))
        except Exception as e:
            print("decode fail", e)
