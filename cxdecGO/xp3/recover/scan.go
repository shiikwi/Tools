package recover

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

func tjsTextHints(text string) []string {
	out := quotedKnownExtNames(text)
	for _, name := range scriptAttributeHints(text) {
		out = appendUnique(out, name)
	}
	return out
}

func sceneFileHints(data []byte) []string {
	text, ok := decodeScriptText(data)
	if !ok {
		text, ok = decodeText(data)
	}
	if !ok {
		return nil
	}
	out := make([]string, 0)
	for _, name := range sceneListNames(text) {
		for _, candidate := range sceneCandidates(name) {
			out = append(out, candidate)
		}
	}
	return out
}

func decodeText(data []byte) (string, bool) {
	if len(data) >= 2 && data[0] == 0xff && data[1] == 0xfe {
		return decodeUTF16(data[2:], binary.LittleEndian), true
	}
	if len(data) >= 2 && data[0] == 0xfe && data[1] == 0xff {
		return decodeUTF16(data[2:], binary.BigEndian), true
	}
	if utf8.Valid(data) {
		return string(data), true
	}
	return "", false
}

func decodeUTF16(data []byte, order binary.ByteOrder) string {
	words := make([]uint16, 0, len(data)/2)
	for len(data) >= 2 {
		words = append(words, order.Uint16(data))
		data = data[2:]
	}
	return string(utf16.Decode(words))
}

func decodeScriptText(data []byte) (string, bool) {
	if plain, ok := descrambleScript(data); ok {
		data = plain
	}
	if len(data) >= 2 && data[0] == 0xff && data[1] == 0xfe {
		return decodeUTF16(data[2:], binary.LittleEndian), true
	}
	if len(data) >= 2 && data[0] == 0xfe && data[1] == 0xff {
		return decodeUTF16(data[2:], binary.BigEndian), true
	}
	return "", false
}

func descrambleScript(data []byte) ([]byte, bool) {
	if len(data) < 5 || data[0] != 0xfe || data[1] != 0xfe || data[3] != 0xff || data[4] != 0xfe {
		return nil, false
	}
	body := data[5:]
	switch data[2] {
	case 0:
		out := []byte{0xff, 0xfe}
		for len(body) >= 2 {
			c := binary.LittleEndian.Uint16(body)
			body = body[2:]
			if c >= 0x20 {
				c ^= ((c & 0x00fe) << 8) ^ 1
			}
			out = binary.LittleEndian.AppendUint16(out, c)
		}
		return out, true
	case 1:
		out := []byte{0xff, 0xfe}
		for len(body) >= 2 {
			c := binary.LittleEndian.Uint16(body)
			body = body[2:]
			c = ((c & 0xaaaa) >> 1) | ((c & 0x5555) << 1)
			out = binary.LittleEndian.AppendUint16(out, c)
		}
		return out, true
	case 2:
		if len(body) < 16 {
			return nil, false
		}
		reader, err := zlib.NewReader(bytes.NewReader(body[16:]))
		if err != nil {
			return nil, false
		}
		defer reader.Close()
		out, err := io.ReadAll(reader)
		return out, err == nil
	default:
		return nil, false
	}
}

func hasKnownExt(name string) bool {
	name = strings.ToLower(name)
	for _, ext := range knownExts {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}

func quotedKnownExtNames(text string) []string {
	out := make([]string, 0)
	seen := map[string]bool{}
	for i := 0; i < len(text); i++ {
		quote := text[i]
		if quote != '"' && quote != '\'' {
			continue
		}
		var b strings.Builder
		escaped := false
		closed := false
		for i++; i < len(text); i++ {
			ch := text[i]
			if escaped {
				b.WriteByte(ch)
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				closed = true
				break
			}
			b.WriteByte(ch)
		}
		if !closed {
			break
		}
		name, isDir, ok := cleanName(b.String())
		if !ok || isDir || !hasKnownExt(name) || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func scriptAttributeHints(text string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}

	for _, line := range strings.Split(text, "\n") {
		command := scriptCommand(line)
		for key, value := range scriptAttributes(line) {
			for _, name := range scriptAttributeCandidates(command, key, value) {
				add(name)
			}
		}
	}
	return out
}

func scriptCommand(line string) string {
	line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
	if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "//") {
		return ""
	}
	line = strings.TrimLeft(line, "@[")
	line = strings.TrimSpace(line)
	end := len(line)
	for i, ch := range line {
		if unicode.IsSpace(ch) || ch == ']' {
			end = i
			break
		}
	}
	return strings.ToLower(line[:end])
}

func scriptAttributes(line string) map[string]string {
	out := map[string]string{}
	for i := 0; i < len(line); i++ {
		if line[i] != '=' {
			continue
		}
		left := i - 1
		for left >= 0 && attrChar(line[left]) {
			left--
		}
		key := strings.ToLower(line[left+1 : i])
		if key == "" {
			continue
		}
		right := i + 1
		for right < len(line) && (line[right] == ' ' || line[right] == '\t') {
			right++
		}
		if right >= len(line) {
			continue
		}
		var value string
		if line[right] == '"' || line[right] == '\'' {
			quote := line[right]
			start := right + 1
			right = start
			for right < len(line) && line[right] != quote {
				right++
			}
			value = line[start:right]
		} else {
			start := right
			for right < len(line) && !unicode.IsSpace(rune(line[right])) && line[right] != ']' && line[right] != ';' {
				right++
			}
			value = line[start:right]
		}
		out[key] = value
	}
	return out
}

func scriptAttributeCandidates(command, key, value string) []string {
	name, isDir, ok := cleanName(value)
	if !ok || isDir {
		return nil
	}
	if hasKnownExt(name) {
		return []string{name}
	}
	switch {
	case key == "rule":
		return tableCandidates(name, []string{".png", ".tlg", ".bmp"})
	case key == "uistorage":
		return tableImageCandidates(name)
	case key == "storage" || key == "file" || key == "filename" || key == "graphic":
		switch {
		case scriptAudioCommand(command):
			return tableCandidates(name, []string{".ogg", ".opus", ".wav", ".tcw", ".ogg.sli"})
		case scriptJumpCommand(command):
			return tableCandidates(name, []string{".ks", ".tjs", ".scn", ".txt"})
		case scriptImageCommand(command):
			return tableImageCandidates(name)
		}
	}
	return nil
}

func attrChar(ch byte) bool {
	return ch == '_' || ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func scriptImageCommand(command string) bool {
	return command == "bg" || command == "fg" || command == "image" ||
		command == "anim" || command == "button" || command == "glyph" ||
		command == "layopt" || command == "meswinopt" || command == "selopt" ||
		command == "trans" || strings.Contains(command, "image")
}

func scriptAudioCommand(command string) bool {
	return command == "playbgm" || command == "playse" || command == "fadebgm" ||
		command == "fadese" || command == "stopse" || strings.Contains(command, "sound")
}

func scriptJumpCommand(command string) bool {
	return command == "jump" || command == "call" || command == "rclick" ||
		command == "addsysscript" || command == "addsyshook"
}

func sceneListNames(text string) []string {
	out := make([]string, 0)
	seen := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") || strings.Contains(line, ":") {
			continue
		}
		name := strings.TrimSpace(strings.Split(line, "\t")[0])
		if i := strings.LastIndexAny(name, `/\>`); i >= 0 {
			name = name[i+1:]
		}
		lower := strings.ToLower(name)
		if !strings.HasSuffix(lower, ".txt") && !strings.HasSuffix(lower, ".scn") && !strings.HasSuffix(lower, ".ks") {
			continue
		}
		if name == "" || name == "-" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func sceneCandidates(name string) []string {
	name = normalizePath(name)
	if name == "" {
		return nil
	}
	candidates := make([]string, 0, 4)
	if strings.HasSuffix(name, ".ks") {
		candidates = appendSceneCandidate(candidates, name)
		return appendSceneCandidate(candidates, name+".scn")
	}
	if strings.HasSuffix(name, ".scn") {
		return appendSceneCandidate(candidates, name)
	}
	candidates = appendSceneCandidate(candidates, name+".scn")
	candidates = appendSceneCandidate(candidates, name+".ks")
	candidates = appendSceneCandidate(candidates, name+".ks.scn")
	if stem, ok := strings.CutSuffix(name, ".txt"); ok {
		candidates = appendSceneCandidate(candidates, stem+".scn")
		candidates = appendSceneCandidate(candidates, stem+".ks")
		candidates = appendSceneCandidate(candidates, stem+".ks.scn")
	}
	return candidates
}

func appendSceneCandidate(out []string, name string) []string {
	out = appendUnique(out, name)
	if !strings.Contains(name, "/") {
		out = appendUnique(out, "scn/"+name)
		out = appendUnique(out, "scenario/"+name)
	}
	return out
}

func appendUnique(out []string, value string) []string {
	for _, item := range out {
		if item == value {
			return out
		}
	}
	return append(out, value)
}
