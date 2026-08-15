package opcode

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
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

		var size int = int(binary.LittleEndian.Uint16(data[off:]))
		if size < 4 {
			return nil, nil, fmt.Errorf("0x%X: Parse fail, bad size: %d", off, size)
		}
		end := off + size
		if end > len(data) {
			return nil, nil, fmt.Errorf("0x%X: Parse fail, out of range", off)
		}

		var op byte = data[off+2]
		var flag byte = data[off+3]

		var extra []byte
		cursor := off + 4
		extraSize := 0

		switch flag & 0x03 {
		case 0x00:
		case 0x01:
			extraSize = 2
		case 0x02:
			extraSize = 2
		case 0x03:
			extraSize = 4
		}
		if cursor+extraSize > end {
			return nil, nil, fmt.Errorf("0x%X: Parse fail, extra out of range", off)
		}
		if extraSize > 0 {
			extra = data[cursor : cursor+extraSize]
			cursor += extraSize
		}

		result = append(result, CodeLine{
			Offset:  off,
			Size:    int16(size),
			Opcode:  op,
			Flag:    flag,
			Extra:   extra,
			Payload: data[cursor:end],
		})

		opcodes[op] = true

		pad := size
		if pad&1 != 0 {
			pad++
		}
		off += pad
	}

	return result, opcodes, nil
}

func LoadOpcodeTxt(path string) (map[byte]string, error) {
	data, _ := os.ReadFile(path)

	names := make(map[byte]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		left, right, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			return nil, fmt.Errorf("Load opcode fail")
		}

		left = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(left)), "0x")
		value, err := strconv.ParseUint(left, 16, 8)
		if err != nil {
			return nil, fmt.Errorf("bad opcode line: %s", line)
		}

		names[byte(value)] = strings.TrimSpace(right)
	}

	return names, nil
}
