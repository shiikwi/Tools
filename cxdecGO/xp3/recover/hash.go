package recover

import (
	"encoding/hex"
	"math/bits"
	"unicode/utf16"

	"golang.org/x/crypto/blake2s"
)

func pathHash(path, domain string) string {
	data := hashInput(path, domain)
	v0 := uint64(0x736f6d6570736575)
	v1 := uint64(0x646f72616e646f6d)
	v2 := uint64(0x6c7967656e657261)
	v3 := uint64(0x7465646279746573)

	off := 0
	for ; off+8 <= len(data); off += 8 {
		m := uint64(data[off]) | uint64(data[off+1])<<8 | uint64(data[off+2])<<16 | uint64(data[off+3])<<24 |
			uint64(data[off+4])<<32 | uint64(data[off+5])<<40 | uint64(data[off+6])<<48 | uint64(data[off+7])<<56
		v3 ^= m
		sipRound(&v0, &v1, &v2, &v3)
		sipRound(&v0, &v1, &v2, &v3)
		v0 ^= m
	}

	tail := uint64(len(data)&0xff) << 56
	for i, b := range data[off:] {
		tail |= uint64(b) << (8 * i)
	}
	v3 ^= tail
	sipRound(&v0, &v1, &v2, &v3)
	sipRound(&v0, &v1, &v2, &v3)
	v0 ^= tail

	v2 ^= 0xff
	for i := 0; i < 4; i++ {
		sipRound(&v0, &v1, &v2, &v3)
	}
	sum := v0 ^ v1 ^ v2 ^ v3
	out := []byte{
		byte(sum), byte(sum >> 8), byte(sum >> 16), byte(sum >> 24),
		byte(sum >> 32), byte(sum >> 40), byte(sum >> 48), byte(sum >> 56),
	}
	return hex.EncodeToString(out)
}

func fileHash(file, domain string) string {
	sum := blake2s.Sum256(hashInput(file, domain))
	return hex.EncodeToString(sum[:])
}

func hashInput(value, domain string) []byte {
	out := utf16le(value)
	if domain != "" {
		out = append(out, utf16le(domain)...)
	}
	return out
}

func utf16le(value string) []byte {
	words := utf16.Encode([]rune(value))
	out := make([]byte, 0, len(words)*2)
	for _, word := range words {
		out = append(out, byte(word), byte(word>>8))
	}
	return out
}

func sipRound(v0, v1, v2, v3 *uint64) {
	*v0 += *v1
	*v1 = bits.RotateLeft64(*v1, 13)
	*v1 ^= *v0
	*v0 = bits.RotateLeft64(*v0, 32)

	*v2 += *v3
	*v3 = bits.RotateLeft64(*v3, 16)
	*v3 ^= *v2

	*v0 += *v3
	*v3 = bits.RotateLeft64(*v3, 21)
	*v3 ^= *v0

	*v2 += *v1
	*v1 = bits.RotateLeft64(*v1, 17)
	*v1 ^= *v2
	*v2 = bits.RotateLeft64(*v2, 32)
}

func FileHashHex(name string) string { return fileHash(name, DefaultDomain) }

func PathHashHex(path string) string { return pathHash(path, DefaultDomain) }
