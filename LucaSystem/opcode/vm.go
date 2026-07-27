package opcode

type VM struct {
	data []byte
	pos  int
}

func (v *VM) Reset(data []byte) {
	v.data = data
	v.pos = 0
}

func (v *VM) GetPos() int {
	return v.pos
}

func (v *VM) GetDataLen() int {
	return len(v.data)
}
