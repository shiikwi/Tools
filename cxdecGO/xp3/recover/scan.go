package recover

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"cxdecgo/internal/tjs2"
)

func scanTJS(root string, domain string) (map[string]string, error) {
	return scanFiles(root, domain, 0, func(data []byte) []string {
		if !bytes.HasPrefix(data, []byte("TJS2100\x00")) {
			return nil
		}
		values, err := tjs2.Strings(data)
		if err != nil {
			return nil
		}
		return values
	})
}

func scanText(root string, domain string) (map[string]string, error) {
	return scanFiles(root, domain, 4<<20, func(data []byte) []string {
		return textHints(data)
	})
}

func scanSceneList(root string, domain string) (map[string]string, error) {
	return scanFiles(root, domain, 1<<20, func(data []byte) []string {
		text, ok := decodeText(data)
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
	})
}

func textHints(data []byte) []string {
	text, ok := decodeText(data)
	if !ok {
		return nil
	}
	out := make([]string, 0)
	seen := map[string]bool{}
	for _, value := range quotedStrings(text) {
		addTextHint(&out, seen, value, true)
	}
	for _, value := range textTokens(text) {
		addTextHint(&out, seen, value, false)
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

func quotedStrings(text string) []string {
	out := make([]string, 0)
	var current strings.Builder
	inQuote := false
	escaped := false
	for _, ch := range text {
		if !inQuote {
			if ch == '"' {
				inQuote = true
				current.Reset()
			}
			continue
		}
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			if current.Len() != 0 {
				out = append(out, current.String())
			}
			current.Reset()
			inQuote = false
			continue
		}
		current.WriteRune(ch)
	}
	return out
}

func textTokens(text string) []string {
	out := make([]string, 0)
	var current strings.Builder
	for _, ch := range text {
		if ch == '/' || ch == '\\' || ch == '_' || ch == '-' || ch == '.' || ch > 127 || unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			current.WriteRune(ch)
			continue
		}
		if current.Len() != 0 && current.Len() <= 180 {
			out = append(out, current.String())
		}
		current.Reset()
	}
	if current.Len() != 0 && current.Len() <= 180 {
		out = append(out, current.String())
	}
	return out
}

func addTextHint(out *[]string, seen map[string]bool, raw string, includeParts bool) {
	name, _, ok := cleanName(raw)
	if !ok || seen[name] {
		return
	}
	seen[name] = true
	*out = append(*out, name)
	if !strings.Contains(name, "/") {
		return
	}

	parts := strings.Split(strings.Trim(name, "/"), "/")
	for i := range parts {
		suffix := strings.Join(parts[i:], "/")
		if suffix != "" && !seen[suffix] {
			seen[suffix] = true
			*out = append(*out, suffix)
		}
		if includeParts {
			dir := normalizeDir(suffix)
			if dir != "" && !seen[dir] {
				seen[dir] = true
				*out = append(*out, dir)
			}
		}
	}
	if includeParts {
		for _, part := range parts {
			if part != "" && !seen[part] {
				seen[part] = true
				*out = append(*out, part)
			}
		}
	}
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
	candidates := make([]string, 0, 3)
	if strings.HasSuffix(name, ".scn") {
		return appendSceneCandidate(candidates, name)
	}
	candidates = appendSceneCandidate(candidates, name+".scn")
	if stem, ok := strings.CutSuffix(name, ".txt"); ok {
		candidates = appendSceneCandidate(candidates, stem+".scn")
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

func scanFiles(root, domain string, maxSize int64, hints func([]byte) []string) (map[string]string, error) {
	files, err := listFiles(root)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return map[string]string{}, nil
	}

	workers := runtime.NumCPU()
	if workers > len(files) {
		workers = len(files)
	}
	jobs := make(chan string)
	results := make(chan map[string]string, workers)
	for i := 0; i < workers; i++ {
		go func() {
			local := map[string]string{}
			for path := range jobs {
				info, err := os.Stat(path)
				if err != nil || info.IsDir() || maxSize > 0 && info.Size() > maxSize {
					continue
				}
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				for _, value := range hints(data) {
					addNameHint(local, domain, value)
				}
			}
			results <- local
		}()
	}
	for _, path := range files {
		jobs <- path
	}
	close(jobs)

	out := map[string]string{}
	for i := 0; i < workers; i++ {
		for hash, name := range <-results {
			if _, ok := out[hash]; !ok {
				out[hash] = name
			}
		}
	}
	return out, nil
}

func listFiles(root string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}
