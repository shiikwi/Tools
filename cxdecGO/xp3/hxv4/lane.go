package hxv4

import (
	"math/bits"
)

type laneRng struct {
	s0         uint64
	s1         uint64
	randomType byte
}

func newLaneRng(lane uint32, randomType byte) *laneRng {
	const gamma = 0x9e3779b97f4a7c15
	base := uint64(^lane)<<32 | uint64(lane)
	return &laneRng{
		s0:         splitmix64Final(base + gamma),
		s1:         splitmix64Final(base + gamma + gamma),
		randomType: randomType,
	}
}

func splitmix64Final(value uint64) uint64 {
	value = (value ^ (value >> 30)) * 0xbf58476d1ce4e5b9
	value = (value ^ (value >> 27)) * 0x94d049bb133111eb
	return value ^ (value >> 31)
}

func (r *laneRng) nextU32() uint32 {
	return uint32(r.nextU64())
}

func (r *laneRng) nextU64() uint64 {
	if r.randomType == 0 {
		result := bits.RotateLeft64(r.s0+r.s1, 17) + r.s0
		r.s1 ^= r.s0
		r.s0 = bits.RotateLeft64(r.s0, 49) ^ r.s1 ^ (r.s1 << 21)
		r.s1 = bits.RotateLeft64(r.s1, 28)
		return result
	}
	result := bits.RotateLeft64(r.s0*5, 7) * 9
	r.s1 ^= r.s0
	r.s0 = bits.RotateLeft64(r.s0, 24) ^ r.s1 ^ (r.s1 << 16)
	r.s1 = bits.RotateLeft64(r.s1, 37)
	return result
}

type laneGen struct {
	params []byte
	rng    *laneRng
	length int
	out    []laneRecord
}

func (g *laneGen) addLength(size int) bool {
	if g.length+size > 128 {
		return false
	}
	g.length += size
	return true
}

func (g *laneGen) emit(op int, param uint32) {
	g.out = append(g.out, laneRecord{Param: param, Op: op})
}

func (g *laneGen) genChild(depth int) bool {
	if g.rng.nextU32()&1 != 0 {
		return g.genBinary(depth)
	}
	return g.genUnary(depth)
}

func (g *laneGen) genBinary(depth int) bool {
	if depth <= 1 {
		return g.genLeaf()
	}
	if !g.addLength(1) {
		return false
	}
	g.emit(opRecurse, 0)
	if !g.genChild(depth - 1) {
		return false
	}
	if !g.addLength(2) {
		return false
	}
	g.emit(opStoreScratch, 0)
	if !g.genChild(depth - 1) {
		return false
	}
	var op int
	switch g.rng.nextU32() % 6 {
	case uint32(g.params[8]):
		op = opAddScratch
		if !g.addLength(2) {
			return false
		}
	case uint32(g.params[9]):
		op = opSubScratch
		if !g.addLength(2) {
			return false
		}
	case uint32(g.params[10]):
		op = opScratchMinusResult
		if !g.addLength(4) {
			return false
		}
	case uint32(g.params[11]):
		op = opMulScratch
		if !g.addLength(3) {
			return false
		}
	case uint32(g.params[12]):
		op = opShlScratch
		if !g.addLength(9) {
			return false
		}
	case uint32(g.params[13]):
		op = opShrScratch
		if !g.addLength(9) {
			return false
		}
	default:
		return false
	}
	if !g.addLength(1) {
		return false
	}
	g.emit(op, 0)
	g.emit(opStop, 0)
	return true
}

func (g *laneGen) genUnary(depth int) bool {
	if depth <= 1 {
		return g.genLeaf()
	}
	if !g.genChild(depth - 1) {
		return false
	}
	switch g.rng.nextU32() & 7 {
	case uint32(g.params[0]):
		if !g.addLength(2) {
			return false
		}
		g.emit(opNot, 0)
	case uint32(g.params[1]):
		if !g.addLength(2) {
			return false
		}
		g.emit(opNeg, 0)
	case uint32(g.params[2]):
		if !g.addLength(1) {
			return false
		}
		g.emit(opInc, 0)
	case uint32(g.params[3]):
		if !g.addLength(1) {
			return false
		}
		g.emit(opDec, 0)
	case uint32(g.params[4]):
		if !g.addLength(21) {
			return false
		}
		g.emit(opBitShuffle, 0xaaaaaaaa)
	case uint32(g.params[5]):
		if !g.addLength(1) {
			return false
		}
		param := g.rng.nextU32()
		if !g.addLength(4) {
			return false
		}
		g.emit(opXorImm, param)
	case uint32(g.params[6]):
		add := g.rng.nextU32()&1 != 0
		if !g.addLength(1) {
			return false
		}
		param := g.rng.nextU32()
		if !g.addLength(4) {
			return false
		}
		if add {
			g.emit(opAddImm, param)
		} else {
			g.emit(opSubImm, param)
		}
	case uint32(g.params[7]):
		if !g.addLength(13) {
			return false
		}
		g.emit(opTableMasked, 1023)
	default:
		return false
	}
	return true
}

func (g *laneGen) genLeaf() bool {
	switch g.rng.nextU32() % 3 {
	case uint32(g.params[14]):
		if !g.addLength(1) {
			return false
		}
		param := g.rng.nextU32()
		if !g.addLength(4) {
			return false
		}
		g.emit(opSetImm, param)
	case uint32(g.params[15]):
		if !g.addLength(2) {
			return false
		}
		g.emit(opSetSeed, 0)
	case uint32(g.params[16]):
		if !g.addLength(7) {
			return false
		}
		param := g.rng.nextU32() & 0x3ff
		if !g.addLength(4) {
			return false
		}
		g.emit(opTableImm, param)
	default:
		return false
	}
	return true
}

func buildLane(params []byte, rng *laneRng) []laneRecord {
	for depth := 5; depth >= 1; depth-- {
		gen := &laneGen{params: params, rng: rng}
		if !gen.addLength(9) {
			continue
		}
		if !gen.genBinary(depth) {
			continue
		}
		if !gen.addLength(6) {
			continue
		}
		gen.emit(opStop, 0)
		return gen.out
	}
	return nil
}
