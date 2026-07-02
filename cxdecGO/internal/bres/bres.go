package bres

import (
	"crypto/sha3"
	"encoding/binary"
)

var chachaConst = [4]uint32{0x61707865, 0x3320646e, 0x79622d32, 0x6b206574}

func Decrypt(ciphertext []byte, pathKey string, salt []byte) []byte {
	keyWords, nonceWords, ctrBase, ctrHigh := deriveParams(pathKey, salt)
	out := make([]byte, len(ciphertext))

	var state [16]uint32
	var stream [64]byte
	for blockNum, off := uint32(0), 0; off < len(ciphertext); blockNum, off = blockNum+1, off+64 {
		state[0], state[1], state[2], state[3] = chachaConst[0], chachaConst[1], chachaConst[2], chachaConst[3]
		copy(state[4:12], keyWords[:])
		state[12] = ctrBase ^ blockNum
		state[13] = ctrHigh
		state[14] = nonceWords[0]
		state[15] = nonceWords[1]
		chacha8Block(&stream, &state)

		n := min(64, len(ciphertext)-off)
		for i := 0; i < n; i++ {
			out[off+i] = ciphertext[off+i] ^ stream[i]
		}
	}
	return out
}

func deriveParams(pathKey string, salt []byte) ([8]uint32, [2]uint32, uint32, uint32) {
	h := sha3.New384()
	h.Write(utf16LE(pathKey))
	h.Write(salt)
	digest := h.Sum(nil)

	var key [8]uint32
	for i := range key {
		key[i] = binary.LittleEndian.Uint32(digest[i*4:])
	}
	nonce := [2]uint32{
		binary.LittleEndian.Uint32(digest[32:36]),
		binary.LittleEndian.Uint32(digest[36:40]),
	}
	return key, nonce, binary.LittleEndian.Uint32(digest[40:44]), binary.LittleEndian.Uint32(digest[44:48])
}

func utf16LE(s string) []byte {
	out := make([]byte, 0, len(s)*2)
	for _, r := range s {
		if r <= 0xffff {
			out = binary.LittleEndian.AppendUint16(out, uint16(r))
			continue
		}
		r -= 0x10000
		hi := uint16(0xd800 + (r >> 10))
		lo := uint16(0xdc00 + (r & 0x3ff))
		out = binary.LittleEndian.AppendUint16(out, hi)
		out = binary.LittleEndian.AppendUint16(out, lo)
	}
	return out
}

func chacha8Block(out *[64]byte, state *[16]uint32) {
	x := *state
	for i := 0; i < 4; i++ {
		quarterRound(&x, 0, 4, 8, 12)
		quarterRound(&x, 1, 5, 9, 13)
		quarterRound(&x, 2, 6, 10, 14)
		quarterRound(&x, 3, 7, 11, 15)
		quarterRound(&x, 0, 5, 10, 15)
		quarterRound(&x, 1, 6, 11, 12)
		quarterRound(&x, 2, 7, 8, 13)
		quarterRound(&x, 3, 4, 9, 14)
	}
	for i := range x {
		x[i] += state[i]
		binary.LittleEndian.PutUint32(out[i*4:], x[i])
	}
}

func quarterRound(s *[16]uint32, a, b, c, d int) {
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

func rotl32(v uint32, n uint) uint32 {
	return (v << n) | (v >> (32 - n))
}
