"""Minimal CLI metadata reader (no third-party deps).
Reads TypeDef / Field / MethodDef / MemberRef names from a managed PE.
Static, read-only. No assembly loading, no execution, no IL decompilation.
"""
import struct, sys, os, io, json

def u16(d, o): return struct.unpack_from('<H', d, o)[0]
def u32(d, o): return struct.unpack_from('<I', d, o)[0]

class Meta:
    def __init__(self, path):
        self.d = open(path, 'rb').read()
        d = self.d
        pe = u32(d, 0x3c)
        assert d[pe:pe+4] == b'PE\0\0'
        nsec = u16(d, pe+6); optsz = u16(d, pe+20)
        opt = pe + 24
        magic = u16(d, opt)
        dd = opt + (96 if magic == 0x10b else 112) + 14*8
        cli_rva, cli_size = u32(d, dd), u32(d, dd+4)
        assert cli_rva and cli_size, 'not a managed PE'
        self.secs = []
        so = opt + optsz
        for i in range(nsec):
            b = so + 40*i
            self.secs.append((u32(d, b+12), u32(d, b+8), u32(d, b+20), u32(d, b+16)))
        cli = self.off(cli_rva)
        md_rva = u32(d, cli+8)
        md = self.off(md_rva)
        assert d[md:md+4] == b'BSJB'
        vlen = u32(d, md+12)
        p = md + 16 + vlen + 2
        nstreams = u16(d, p); p += 2
        self.streams = {}
        for _ in range(nstreams):
            off, size = u32(d, p), u32(d, p+4); p += 8
            e = d.index(b'\0', p)
            name = d[p:e].decode('ascii'); p = (e + 4) & ~3
            self.streams[name] = (md + off, size)
        to, _ = self.streams['#~']
        hs = d[to+6]
        valid = struct.unpack_from('<Q', d, to+8)[0]
        p = to + 24
        self.rows = {}
        for t in range(64):
            if (valid >> t) & 1:
                self.rows[t] = u32(d, p); p += 4
        self.sw = 4 if hs & 1 else 2
        self.gw = 4 if hs & 2 else 2
        self.bw = 4 if hs & 4 else 2
        S, G, B = self.sw, self.gw, self.bw
        def T(t): return 4 if self.rows.get(t, 0) >= 65536 else 2
        def C(name):
            bits, tabs = CODED[name]
            return 4 if max(self.rows.get(t, 0) for t in tabs) >= (1 << (16-bits)) else 2
        self.schema = {
            0x00: [2, S, G, G, G],
            0x01: [C('ResolutionScope'), S, S],
            0x02: [4, S, S, C('TypeDefOrRef'), T(4), T(6)],
            0x03: [T(4)],
            0x04: [2, S, B],
            0x05: [T(6)],
            0x06: [4, 2, 2, S, B, T(8)],
            0x07: [T(8)],
            0x08: [2, 2, S],
            0x09: [T(2), C('TypeDefOrRef')],
            0x0A: [C('MemberRefParent'), S, B],
        }
        self.rs = {t: sum(schema) for t, schema in self.schema.items()}
        self.offs = {}
        cur = p
        for t in sorted(self.schema):
            if self.rows.get(t):
                self.offs[t] = cur
                cur += self.rows[t] * self.rs[t]

    def off(self, rva):
        for va, vs, ra, rs in self.secs:
            if va <= rva < va + max(vs, rs):
                return ra + (rva - va)
        raise ValueError('bad rva %x' % rva)

    def s(self, idx):
        o = self.streams['#Strings'][0] + idx
        return self.d[o:self.d.index(b'\0', o)].decode('utf-8', 'replace')

    def row(self, t, i):
        o = self.offs[t] + i * self.rs[t]
        out = []
        for w in self.schema[t]:
            out.append(int.from_bytes(self.d[o:o+w], 'little')); o += w
        return out

    def types(self):
        n = self.rows.get(2, 0)
        nm = self.rows.get(6, 0)
        out = []
        for i in range(n):
            v = self.row(2, i)
            out.append({'ns': self.s(v[2]), 'name': self.s(v[1]),
                        'flags': v[0], 'm0': v[5], 'f0': v[4]})
        for i, t in enumerate(out):
            t['m1'] = out[i+1]['m0'] if i+1 < len(out) else nm + 1
            t['f1'] = out[i+1]['f0'] if i+1 < len(out) else self.rows.get(4, 0) + 1
        return out

    def methods(self, t):
        return [(self.row(6, i)[0], self.s(self.row(6, i)[3]))
                for i in range(t['m0']-1, min(t['m1']-1, self.rows.get(6, 0)))]

    def fields(self, t):
        return [self.s(self.row(4, i)[1])
                for i in range(t['f0']-1, min(t['f1']-1, self.rows.get(4, 0)))]

    def memberrefs(self):
        out = []
        for i in range(self.rows.get(0x0A, 0)):
            v = self.row(0x0A, i)
            out.append(self.s(v[1]))
        return out

CODED = {
    'ResolutionScope': (2, [0, 0x1A, 0x23, 1]),
    'TypeDefOrRef': (2, [2, 1, 0x1B]),
    'MemberRefParent': (3, [2, 1, 0x1A, 6, 0x1B]),
}

if __name__ == '__main__':
    base = r'C:\Users\ASUS\AppData\Local\Temp\zying-rea-20261115\named'
    modes = sys.argv[1:] or ['ZYing.User.dll']
    for fn in modes:
        m = Meta(os.path.join(base, fn))
        ts = m.types()
        print('===', fn, '| TypeDef', len(ts), '| MethodDef', m.rows.get(6, 0),
              '| Field', m.rows.get(4, 0), '| MemberRef', m.rows.get(0x0A, 0))
        pat = ('Spider|Sale|Product|Draft|Publish|Market|Cate|Pic|Image|Variant|Bullet|'
               'Intro|Stock|Sku|Audit|Repeat|Duplicate|Shop|Store|Meli|Mercado|Record|ZClient|Zrpc')
        import re
        hits = [t for t in ts if re.search(pat, t['ns'] + '.' + t['name'], re.I)]
        print('--- matching types:', len(hits))
        json.dump([{'type': t['ns'] + '.' + t['name'],
                    'methods': [n for _, n in m.methods(t)],
                    'fields': m.fields(t)} for t in hits],
                  io.open(os.path.join(os.path.dirname(base), 'meta_' + fn + '.json'),
                          'w', encoding='utf-8'), ensure_ascii=False, indent=1)
        for t in hits[:12]:
            ms = m.methods(t)
            print(' ', t['ns'] + '.' + t['name'], '| methods', len(ms),
                  '| fields', len(m.fields(t)))
            print('    ', ', '.join(n for _, n in ms[:35]))
        rec = [t for t in ts if t['name'] == 'Record']
        print('--- ZYing.Record exact matches:', [t['ns'] for t in rec])
