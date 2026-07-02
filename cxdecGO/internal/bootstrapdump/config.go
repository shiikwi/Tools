package bootstrapdump

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"

	"cxdecgo/internal/pefile"
)

func parseConfigTable(dll []byte, tableRVA uint32) (map[string][]byte, error) {
	img, err := pefile.New("bootstrap.dll", dll)
	if err != nil {
		return nil, err
	}
	off, err := img.RVAToOffset(tableRVA)
	if err != nil {
		return nil, err
	}

	result := map[string][]byte{}
	cursor := off
	for cursor < uint32(len(dll)) {
		end := cursor
		for end < uint32(len(dll)) && dll[end] != 0 {
			end++
		}

		label := string(dll[cursor:end])
		cursor = end + 1
		if label == "" {
			break
		}

		length := uint32(binary.LittleEndian.Uint16(dll[cursor:]))
		cursor += 2

		result[label] = append([]byte(nil), dll[cursor:cursor+length]...)
		cursor += length
	}
	return result, nil
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
