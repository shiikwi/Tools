package script

type TXTScript struct {
	Name   string
	Offset uint32
	Size   uint32
	Data   []byte
}

func (s *TXTScript) Decrypt() {
	cryptScript(s.Data, s.Offset)
}

func cryptScript(data []byte, offset uint32) {
	for i := range data {
		p := offset + uint32(i)
		seed := p - 0x5CC8E9D7
		state := (uint32(0xA3371629) >> ((p & 0xF) + 1)) -
			(uint32(0x5CC8E9D7) << (31 - (p & 0xF)))
		state += seed

		for shift := uint(4); shift <= 28; shift += 4 {
			nibble := (p >> shift) & 0xF
			mixed := state << (31 - nibble)
			state >>= nibble + 1
			state += mixed + seed
		}
		data[i] ^= byte(state >> (p & 0xF))
	}
}
