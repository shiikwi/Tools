package opcode

import (
	"encoding/binary"
	"fmt"
)

type CodeLine struct {
	Offset  int
	Size    int16
	Opcode  byte
	Flag    byte
	Extra   []byte
	Payload []byte
}

func ParseInfo(data []byte) ([]CodeLine, map[byte]bool, error) {
	var result []CodeLine
	opcodes := make(map[byte]bool)

	for off := 0; off < len(data); {
		if off+4 > len(data) {
			return nil, nil, fmt.Errorf("0x%X: Parse fail, out of range", off)
		}

		var size int16 = int16(binary.LittleEndian.Uint16(data[off:]))
		var op byte = data[off+2]
		var flag byte = data[off+3]

		var extra []byte
		cursor := off + 4

		switch flag & 0x03 {
		case 0x00:
		case 0x01:
			extra = data[cursor : cursor+2]
			cursor += 2
		case 0x02:
			extra = data[cursor : cursor+2]
			cursor += 2
		case 0x03:
			extra = data[cursor : cursor+4]
			cursor += 4
		}

		result = append(result, CodeLine{
			Offset:  off,
			Size:    size,
			Opcode:  op,
			Flag:    flag,
			Extra:   extra,
			Payload: data[cursor : off+int(size)],
		})

		opcodes[op] = true

		pad := int(size)
		if pad&1 != 0 {
			pad++
		}
		off += pad
	}

	return result, opcodes, nil
}
