package recover

import (
	"bytes"
	"encoding/binary"
	"strings"
	"unicode"
)

var tlgRefHeader = []byte("TLGref\x00")

func tlgRefHints(data []byte) []string {
	if !bytes.HasPrefix(data, tlgRefHeader) {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, 2)
	for i := len(tlgRefHeader); i+8 <= len(data); i++ {
		if data[i] == 0 || data[i+1] != 0 {
			continue
		}
		end := i + 2
		for end+1 < len(data) && !(data[end] == 0 && data[end+1] == 0) {
			end += 2
		}
		if end+1 >= len(data) || end-i < 8 {
			continue
		}
		name, isDir, ok := cleanName(decodeUTF16(data[i:end], binary.LittleEndian))
		if !ok || isDir || seen[name] || !tlgRefName(name) {
			continue
		}
		seen[name] = true
		out = append(out, name)
		i = end + 1
	}
	return out
}

func tlgRefName(name string) bool {
	for _, ch := range name {
		if unicode.IsControl(ch) {
			return false
		}
	}
	name = strings.ToLower(name)
	return strings.HasSuffix(name, ".tlg") ||
		strings.HasSuffix(name, ".tlg5") ||
		strings.HasSuffix(name, ".tlg6")
}
