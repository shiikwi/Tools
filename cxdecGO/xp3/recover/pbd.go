package recover

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"

	"golang.org/x/crypto/blake2s"
)

const pbdLZ4DictSize = 64 * 1024

const (
	xxhP1 = 2654435761
	xxhP2 = 2246822519
	xxhP3 = 3266489917
	xxhP4 = 668265263
	xxhP5 = 374761393
)

func rotl32(x uint32, r uint) uint32 { return x<<r | x>>(32-r) }

func xxh32(data []byte, seed uint32) uint32 {
	const m32 = 0xFFFFFFFF
	n := len(data)
	var h uint32
	i := 0
	if n >= 16 {
		v1 := (seed + xxhP1 + xxhP2) & m32
		v2 := (seed + xxhP2) & m32
		v3 := seed
		v4 := (seed - xxhP1) & m32
		for i+16 <= n {
			for k, v := range []*uint32{&v1, &v2, &v3, &v4} {
				lane := binary.LittleEndian.Uint32(data[i+k*4:])
				*v = (*v + lane*xxhP2) & m32
				*v = rotl32(*v, 13)
				*v = (*v * xxhP1) & m32
			}
			i += 16
		}
		h = (rotl32(v1, 1) + rotl32(v2, 7) + rotl32(v3, 12) + rotl32(v4, 18)) & m32
		for _, v := range []uint32{v1, v2, v3, v4} {
			h = (h ^ ((v*xxhP2 + xxhP3) & m32)) & m32
			h = (h*xxhP1 + xxhP4) & m32
			h ^= h >> 15
			h = (h * xxhP2) & m32
			h ^= h >> 13
			h = (h * xxhP3) & m32
			h ^= h >> 16
		}
	} else {
		h = (seed + xxhP5) & m32
	}
	h = (h + uint32(n)) & m32
	for i+4 <= n {
		h = (h + binary.LittleEndian.Uint32(data[i:])*xxhP3) & m32
		h = (rotl32(h, 17) * xxhP4) & m32
		i += 4
	}
	for i < n {
		h = (h + uint32(data[i])*xxhP5) & m32
		h = (rotl32(h, 11) * xxhP1) & m32
		i++
	}
	h ^= h >> 15
	h = (h * xxhP2) & m32
	h ^= h >> 13
	h = (h * xxhP3) & m32
	h ^= h >> 16
	return h
}

func chachaQuarter(s []uint32, a, b, c, d int) {
	s[a] += s[b]
	s[d] ^= s[a]
	s[d] = rotl32(s[d], 16)
	s[c] += s[d]
	s[b] ^= s[c]
	s[b] = rotl32(s[b], 12)
	s[a] += s[b]
	s[d] ^= s[a]
	s[d] = rotl32(s[d], 8)
	s[c] += s[d]
	s[b] ^= s[c]
	s[b] = rotl32(s[b], 7)
}

func chachaBlock(key []byte, counter uint64, nonceLo, nonceHi uint32, rounds int) [64]byte {
	var st [16]uint32
	st[0], st[1], st[2], st[3] = 0x61707865, 0x3320646e, 0x79622d32, 0x6b206574
	for i := 0; i < 8; i++ {
		st[4+i] = binary.LittleEndian.Uint32(key[i*4:])
	}
	st[12] = uint32(counter)
	st[13] = uint32(counter >> 32)
	st[14] = nonceLo
	st[15] = nonceHi
	orig := st
	for i := 0; i < rounds/2; i++ {
		chachaQuarter(st[:], 0, 4, 8, 12)
		chachaQuarter(st[:], 1, 5, 9, 13)
		chachaQuarter(st[:], 2, 6, 10, 14)
		chachaQuarter(st[:], 3, 7, 11, 15)
		chachaQuarter(st[:], 0, 5, 10, 15)
		chachaQuarter(st[:], 1, 6, 11, 12)
		chachaQuarter(st[:], 2, 7, 8, 13)
		chachaQuarter(st[:], 3, 4, 9, 14)
	}
	var out [64]byte
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(out[i*4:], st[i]+orig[i])
	}
	return out
}

func xorshift32(x uint32, fallback uint32) uint32 {
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	if x == 0 {
		return fallback
	}
	return x
}

func expandStreamBlock(base []byte, blocks int, fallback uint32) []byte {
	if blocks <= 1 {
		return base
	}
	words := make([]uint32, 0, blocks*16)
	for i := 0; i < 16; i++ {
		words = append(words, binary.LittleEndian.Uint32(base[i*4:]))
	}
	src := 0
	for i := 16; i < blocks*16; i++ {
		words = append(words, xorshift32(words[src], fallback))
		src++
	}
	out := make([]byte, len(words)*4)
	for i, w := range words {
		binary.LittleEndian.PutUint32(out[i*4:], w)
	}
	return out
}

var pbdCryptParams = map[uint16][2]int{
	1: {8, 0x10}, 2: {12, 8}, 3: {20, 4}, 4: {8, 1}, 5: {12, 1}, 6: {20, 1},
}

func crypt4s0(data []byte, seed uint32, crypt uint16, iv []byte) ([]byte, error) {
	params, ok := pbdCryptParams[crypt]
	if crypt != 0 && !ok {
		return nil, errors.New("pbd: unsupported crypt field")
	}
	if crypt == 0 {
		return data, nil
	}
	rounds, expandBlocks := params[0], params[1]
	mac, err := blake2s.New256(seedBytes(seed))
	if err != nil {
		return nil, err
	}
	mac.Write(iv)
	key := mac.Sum(nil)
	nonceLo := xxh32(iv, seed)
	nonceHi := seed
	fallback := nonceHi ^ nonceLo
	if fallback == 0 {
		fallback = seed
	}
	if fallback == 0 {
		fallback = 0xFFFFFFFF
	}
	out := make([]byte, len(data))
	off, counter := 0, uint64(0)
	for off < len(data) {
		base := chachaBlock(key, counter, nonceLo, nonceHi, rounds)
		counter++
		stream := expandStreamBlock(base[:], expandBlocks, fallback)
		take := len(stream)
		if take > len(data)-off {
			take = len(data) - off
		}
		for i := 0; i < take; i++ {
			out[off+i] = data[off+i] ^ stream[i]
		}
		off += take
	}
	return out, nil
}

func seedBytes(seed uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, seed)
	return b
}

func lz4Block(src []byte, dict []byte) ([]byte, error) {
	out := make([]byte, 0, pbdLZ4DictSize)
	i := 0
	for i < len(src) {
		token := src[i]
		i++
		lit := int(token >> 4)
		if lit == 15 {
			for {
				if i >= len(src) {
					return nil, errors.New("pbd: lz4 literal length truncated")
				}
				b := src[i]
				i++
				lit += int(b)
				if b != 255 {
					break
				}
			}
		}
		if i+lit > len(src) {
			return nil, errors.New("pbd: lz4 literals truncated")
		}
		out = append(out, src[i:i+lit]...)
		i += lit
		if i >= len(src) {
			break
		}
		if i+2 > len(src) {
			return nil, errors.New("pbd: lz4 offset truncated")
		}
		offset := int(binary.LittleEndian.Uint16(src[i:]))
		i += 2
		if offset == 0 || offset > len(out)+len(dict) {
			return nil, errors.New("pbd: lz4 offset out of range")
		}
		ml := int(token&15) + 4
		if token&15 == 15 {
			for {
				if i >= len(src) {
					return nil, errors.New("pbd: lz4 match length truncated")
				}
				b := src[i]
				i++
				ml += int(b)
				if b != 255 {
					break
				}
			}
		}
		start := len(out) - offset
		for k := 0; k < ml; k++ {
			idx := start + k
			if idx >= 0 {
				out = append(out, out[idx])
			} else {
				out = append(out, dict[len(dict)+idx])
			}
		}
	}
	return out, nil
}

func lz4Frames(data []byte) ([]byte, error) {
	var out []byte
	pos := 0
	for pos < len(data) {
		if pos+2 > len(data) {
			return out, errors.New("pbd: lz4 frame size truncated")
		}
		size := int(binary.LittleEndian.Uint16(data[pos:]))
		pos += 2
		if size == 0 {
			return out, nil
		}
		if pos+size > len(data) {
			return out, errors.New("pbd: lz4 frame truncated")
		}
		dictStart := len(out) - pbdLZ4DictSize
		if dictStart < 0 {
			dictStart = 0
		}
		block, err := lz4Block(data[pos:pos+size], out[dictStart:])
		if err != nil {
			return out, err
		}
		out = append(out, block...)
		pos += size
	}
	return out, nil
}

type pbdChecker interface {
	byteForType(t byte) byte
}

type ns0Checker struct{ seed uint32 }

func pbdRound(b []byte) {
	a := b[0] ^ (b[0] * 2)
	x := a >> 2
	x ^= b[2]
	x = x >> 3
	x ^= b[2]
	x ^= a
	b[0], b[1], b[2] = b[1], b[2], x
}

func (c *ns0Checker) byteForType(t byte) byte {
	b := seedBytes(c.seed)
	if t == 0 {
		return b[2]
	}
	pbdRound(b)
	c.seed = binary.LittleEndian.Uint32(b)
	return b[2]
}

type s0Checker struct{ b0, b1, b2 byte }

func newS0Checker(seed uint32) *s0Checker {
	return &s0Checker{
		b0: byte(seed>>24) ^ byte(seed),
		b1: byte(seed >> 8),
		b2: byte(seed >> 16),
	}
}

func (c *s0Checker) byteForType(t byte) byte {
	if t == 0 {
		return c.b2
	}
	tv := c.b0 ^ (c.b0 << 1)
	c.b0, c.b1 = c.b1, c.b2
	c.b2 = c.b2 ^ (c.b2 >> 3) ^ tv ^ (tv >> 5)
	return c.b2
}

func pbdParse(data []byte, pos int, checker pbdChecker, depth int, strs *[]string, ints *[]int) (int, bool) {
	const maxDepth = 256
	const maxItems = 1_000_000
	if depth > maxDepth || pos+2 > len(data) {
		return pos, false
	}
	t := data[pos]
	if data[pos+1] != checker.byteForType(t) {
		return pos, false
	}
	pos += 2
	switch t {
	case 0:
		return pos, true
	case 2:
		if pos+4 > len(data) {
			return pos, false
		}
		units := int(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		if units > 1<<22 || pos+units*2 > len(data) {
			return pos, false
		}
		words := make([]uint16, units)
		for i := 0; i < units; i++ {
			words[i] = binary.LittleEndian.Uint16(data[pos+i*2:])
		}
		*strs = append(*strs, string(utf16.Decode(words)))
		return pos + units*2, true
	case 4:
		if pos+8 > len(data) {
			return pos, false
		}
		v := int64(binary.LittleEndian.Uint64(data[pos:]))
		if v >= 0 && v <= 1_000_000 {
			*ints = append(*ints, int(v))
		}
		return pos + 8, true
	case 5:
		return pos + 8, true
	case 0x81:
		if pos+4 > len(data) {
			return pos, false
		}
		count := int(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		if count > maxItems {
			return pos, false
		}
		for i := 0; i < count; i++ {
			var ok bool
			pos, ok = pbdParse(data, pos, checker, depth+1, strs, ints)
			if !ok {
				return pos, false
			}
		}
		return pos, true
	case 0xc1:
		if pos+4 > len(data) {
			return pos, false
		}
		count := int(binary.LittleEndian.Uint32(data[pos:]))
		pos += 4
		if count > maxItems {
			return pos, false
		}
		for i := 0; i < count; i++ {
			if pos+4 > len(data) {
				return pos, false
			}
			units := int(binary.LittleEndian.Uint32(data[pos:]))
			pos += 4
			if units > 1<<22 || pos+units*2 > len(data) {
				return pos, false
			}
			words := make([]uint16, units)
			for j := 0; j < units; j++ {
				words[j] = binary.LittleEndian.Uint16(data[pos+j*2:])
			}
			*strs = append(*strs, string(utf16.Decode(words)))
			pos += units * 2
			var ok bool
			pos, ok = pbdParse(data, pos, checker, depth+1, strs, ints)
			if !ok {
				return pos, false
			}
		}
		return pos, true
	}
	return pos, false
}

func Crypt4S0Test(data []byte, seed uint32, crypt uint16, iv []byte) ([]byte, error) {
	return crypt4s0(data, seed, crypt, iv)
}

func LZ4FramesTest(data []byte) ([]byte, error) { return lz4Frames(data) }

func pbdDecode(data []byte) ([]string, []int, bool) {
	if len(data) < 20 {
		return nil, nil, false
	}
	magic := string(data[:8])
	seed := binary.LittleEndian.Uint32(data[8:])
	crypt := binary.LittleEndian.Uint16(data[12:])
	switch magic {
	case "TJS/ns0\x00":
		body := data[16:]
		if len(body) < 4 {
			return nil, nil, false
		}
		checker := &ns0Checker{seed: seed}
		var strs []string
		var ints []int
		if _, ok := pbdParse(body, 0, checker, 0, &strs, &ints); !ok {
			return strs, ints, false
		}
		return strs, ints, true
	case "TJS/4s0\x00":
		ivLen := int(binary.LittleEndian.Uint16(data[14:]))
		if 16+ivLen > len(data) {
			return nil, nil, false
		}
		iv := data[16 : 16+ivLen]
		framed, err := crypt4s0(data[16+ivLen:], seed, crypt, iv)
		if err != nil {
			return nil, nil, false
		}
		plain, err := lz4Frames(framed)
		if err != nil || len(plain) < 4 {
			return nil, nil, false
		}
		checker := newS0Checker(seed)
		var strs []string
		var ints []int
		if _, ok := pbdParse(plain[:len(plain)-4], 0, checker, 0, &strs, &ints); !ok {
			return strs, ints, false
		}
		return strs, ints, true
	}
	return nil, nil, false
}
