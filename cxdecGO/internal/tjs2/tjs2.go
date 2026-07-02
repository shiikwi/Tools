package tjs2

import (
	"encoding/binary"
	"fmt"
	"unicode/utf16"
)

func Strings(data []byte) ([]string, error) {
	for off := uint32(12); off+8 <= uint32(len(data)); {
		tag := string(data[off : off+4])
		size := binary.LittleEndian.Uint32(data[off+4:])
		body := data[off+8 : off+size]
		if tag == "DATA" {
			return parseDataStrings(body)
		}
		off += size
	}
	return nil, fmt.Errorf("TJS bytecode has no DATA chunk")
}

func parseDataStrings(data []byte) ([]string, error) {
	p := uint32(0)
	skipCount := func(unit uint32) {
		count := binary.LittleEndian.Uint32(data[p:])
		p += 4
		size := count * unit
		p += align4(size)
	}
	for _, unit := range []uint32{1, 2, 4, 8, 8} {
		skipCount(unit)
	}

	count := binary.LittleEndian.Uint32(data[p:])
	p += 4
	strings := make([]string, 0, count)
	for i := uint32(0); i < count; i++ {
		chars := binary.LittleEndian.Uint32(data[p:])
		p += 4
		rawSize := chars * 2
		words := make([]uint16, chars)
		for j := range words {
			words[j] = binary.LittleEndian.Uint16(data[p+uint32(j)*2:])
		}
		strings = append(strings, string(utf16.Decode(words)))
		p += align4(rawSize)
	}
	return strings, nil
}

func align4(value uint32) uint32 {
	return (value + 3) &^ 3
}
