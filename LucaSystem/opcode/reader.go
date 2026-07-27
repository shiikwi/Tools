package opcode

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

func (v *VM) ReadU8() (byte, error) {
	if v.pos+1 > len(v.data) {
		return 0, fmt.Errorf("ReadU8 out of range")
	}

	value := v.data[v.pos]
	v.pos++
	return value, nil
}

func (v *VM) ReadU16() (uint16, error) {
	if v.pos+2 > len(v.data) {
		return 0, fmt.Errorf("ReadU16 out of range")
	}

	value := binary.LittleEndian.Uint16(v.data[v.pos:])
	v.pos += 2
	return value, nil
}

func (v *VM) ReadU32() (uint32, error) {
	if v.pos+4 > len(v.data) {
		return 0, fmt.Errorf("ReadU32 out of range")
	}

	value := binary.LittleEndian.Uint32(v.data[v.pos:])
	v.pos += 4
	return value, nil
}

func (v *VM) ReadByte() (byte, error) {
	if v.pos+1 > len(v.data) {
		return 0, fmt.Errorf("ReadByte out of range")
	}

	value := v.data[v.pos]
	v.pos++
	return value, nil
}

func (v *VM) ReadBytes(size int) ([]byte, error) {
	if v.pos+size > len(v.data) {
		return nil, fmt.Errorf("ReadBytes out of range")
	}

	value := v.data[v.pos : v.pos+size]
	v.pos += size
	return value, nil
}

func (v *VM) ReadLenUTF16() (string, error) {
	len, err := v.ReadU16()
	if err != nil {
		return "", err
	}

	data, err := v.ReadBytes((int(len)) * 2)
	if err != nil {
		return "", err
	}

	words := make([]uint16, len)
	for i := range words {
		words[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return string(utf16.Decode(words)), nil
}

func (v *VM) ReadLenUTF8() (string, error) {
	len, err := v.ReadU16()
	if err != nil {
		return "", err
	}

	data, err := v.ReadBytes((int(len)))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (v *VM) ReadUTF82NULL() (string, error) {
	var buf []byte
	for {
		b, err := v.ReadByte()
		if err != nil {
			return "", err
		}

		if b == 0x00 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf), nil
}

func (v *VM) ReadUTF162NULL() (string, error) {
	var buf []uint16
	for {
		b, err := v.ReadU16()
		if err != nil {
			return "", err
		}

		if b == 0x0000 {
			break
		}
		buf = append(buf, b)
	}
	return string(utf16.Decode(buf)), nil
}
