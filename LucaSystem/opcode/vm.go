package opcode

type VM struct {
	data []byte
	pos  int
}

func (v *VM) Reset(data []byte) {
	v.data = data
	v.pos = 0
}
