package hxv4

import (
	"math/bits"
)

type hxSponge struct {
	rate   int
	pos    int
	state  [25]uint64
	domain byte
}

func newHxSponge(rate int, domain byte) *hxSponge {
	return &hxSponge{rate: rate, domain: domain}
}

func (s *hxSponge) absorb(input []byte) *hxSponge {
	off := 0
	for off < len(input) {
		count := s.rate - s.pos
		if count > len(input)-off {
			count = len(input) - off
		}
		for i := 0; i < count; i++ {
			p := s.pos + i
			s.state[p/8] ^= uint64(input[off+i]) << (8 * (p % 8))
		}
		s.pos += count
		off += count
		if s.pos == s.rate {
			keccakF1600(&s.state)
			s.pos = 0
		}
	}
	return s
}

func (s *hxSponge) squeeze(out []byte) {
	s.state[s.pos/8] ^= uint64(s.domain) << (8 * (s.pos % 8))
	last := s.rate - 1
	s.state[last/8] ^= 0x80 << (8 * (last % 8))
	keccakF1600(&s.state)
	off := 0
	for off < len(out) {
		count := s.rate
		if count > len(out)-off {
			count = len(out) - off
		}
		for i := 0; i < count; i++ {
			out[off+i] = byte(s.state[i/8] >> (8 * (i % 8)))
		}
		off += count
		if off < len(out) {
			keccakF1600(&s.state)
		}
	}
}

func keccakF1600(state *[25]uint64) {
	var roundConstants = [24]uint64{
		0x0000000000000001, 0x0000000000008082, 0x800000000000808a, 0x8000000080008000,
		0x000000000000808b, 0x0000000080000001, 0x8000000080008081, 0x8000000000008009,
		0x000000000000008a, 0x0000000000000088, 0x0000000080008009, 0x000000008000000a,
		0x000000008000808b, 0x800000000000008b, 0x8000000000008089, 0x8000000000008003,
		0x8000000000008002, 0x8000000000000080, 0x000000000000800a, 0x800000008000000a,
		0x8000000080008081, 0x8000000000008080, 0x0000000080000001, 0x8000000080008008,
	}
	var rotations = [25]uint32{0, 1, 62, 28, 27, 36, 44, 6, 55, 20, 3, 10, 43, 25, 39, 41, 45, 15, 21, 8, 18, 2, 61, 56, 14}
	for _, round := range roundConstants {
		var cross [5]uint64
		for x := 0; x < 5; x++ {
			cross[x] = state[x] ^ state[x+5] ^ state[x+10] ^ state[x+15] ^ state[x+20]
		}
		var corrected [5]uint64
		for x := 0; x < 5; x++ {
			corrected[x] = cross[(x+4)%5] ^ bits.RotateLeft64(cross[(x+1)%5], 1)
		}
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				state[x+5*y] ^= corrected[x]
			}
		}
		var rotated [25]uint64
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				rotated[y+5*((2*x+3*y)%5)] = bits.RotateLeft64(state[x+5*y], int(rotations[x+5*y]))
			}
		}
		for y := 0; y < 5; y++ {
			for x := 0; x < 5; x++ {
				state[x+5*y] = rotated[x+5*y] ^ ((^rotated[(x+1)%5+5*y]) & rotated[(x+2)%5+5*y])
			}
		}
		state[0] ^= round
	}
}
