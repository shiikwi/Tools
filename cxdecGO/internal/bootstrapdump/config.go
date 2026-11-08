package bootstrapdump

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"
)

const configMaxValueLen = 0x4000

func findConfigEntry(dll []byte, label string) ([]byte, error) {
	pattern := append([]byte(label), 0)
	var found []byte
	start := 0
	for {
		i := bytes.Index(dll[start:], pattern)
		if i < 0 {
			break
		}
		off := start + i
		start = off + 1
		cur := off + len(pattern)
		if cur+2 > len(dll) {
			continue
		}
		length := int(binary.LittleEndian.Uint16(dll[cur:]))
		cur += 2
		if length == 0 || length > configMaxValueLen || cur+length > len(dll) {
			continue
		}
		found = append([]byte(nil), dll[cur:cur+length]...)
	}
	if found == nil {
		return nil, fmt.Errorf("config entry %q not found in bootstrap dll", label)
	}
	return found, nil
}

func requireConfig(config map[string][]byte, label string) ([]byte, error) {
	value, ok := config[label]
	if ok {
		return value, nil
	}
	labels := make([]string, 0, len(config))
	for label := range config {
		labels = append(labels, label)
	}
	return nil, fmt.Errorf("DLL config label %q was not found; available labels: %s", label, strings.Join(labels, ", "))
}

func decodeUTF16LE(data []byte) string {
	for len(data) >= 2 && data[len(data)-1] == 0 && data[len(data)-2] == 0 {
		data = data[:len(data)-2]
	}
	words := make([]uint16, len(data)/2)
	for i := range words {
		words[i] = binary.LittleEndian.Uint16(data[i*2:])
	}
	return string(utf16.Decode(words))
}

func decodeASCII(data []byte) string {
	if nul := strings.IndexByte(string(data), 0); nul >= 0 {
		data = data[:nul]
	}
	return string(data)
}
