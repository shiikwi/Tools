package recover

import (
	"bytes"
	"compress/zlib"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"cxdecgo/internal/tjs2"
)

var mediaMagics = [][]byte{
	[]byte("OggS"), {0x89, 'P', 'N', 'G'}, []byte("TLG5\x00"), []byte("TLG6\x00"),
	[]byte("RIFF"), []byte("ftyp"), {0x8B, 'P', 'S', 'D'}, []byte("ID3"),
	[]byte("PK\x03\x04"), {0x1f, 0x8b}, []byte("7z\xbc\xaf"),
	[]byte("TLGref\x00"), []byte("TLG0\x00"), []byte("TLGX\x00"),
}

func isMediaMagic(data []byte) bool {
	for _, m := range mediaMagics {
		if bytes.HasPrefix(data, m) {
			return true
		}
	}
	return false
}

var asciiRun = regexp.MustCompile(`[\x20-\x7e]{3,}`)

func isShapeWord(w uint16) bool {
	if w < 0x20 || w == 0x7f {
		return false
	}
	if w >= 0x80 && w <= 0x9f {
		return false
	}
	if w >= 0xd800 && w <= 0xdfff {
		return false
	}
	return true
}

func harvestRuns(data []byte) []string {
	out := make([]string, 0, 64)
	for _, m := range asciiRun.FindAll(data, -1) {
		out = append(out, string(m))
		if len(out) >= 20000 {
			return out
		}
	}
	n := len(data) / 2
	words := make([]uint16, n)
	for i := 0; i < n; i++ {
		words[i] = uint16(data[i*2]) | uint16(data[i*2+1])<<8
	}
	cur := make([]uint16, 0, 64)
	flush := func() {
		if len(cur) >= 3 {
			out = append(out, string(utf16.Decode(cur)))
		}
		cur = cur[:0]
	}
	for _, w := range words {
		if isShapeWord(w) {
			cur = append(cur, w)
		} else {
			flush()
			if len(out) >= 20000 {
				return out
			}
		}
	}
	flush()
	return out
}

var utf8Token = regexp.MustCompile(`[\w./\-\x{3000}-\x{30ff}\x{4e00}-\x{9fff}\x{ff00}-\x{ffef}]+`)

var shapeSplit = regexp.MustCompile(`[^\w./\-\x{3000}-\x{30ff}\x{4e00}-\x{9fff}\x{ff00}-\x{ffef}]+`)

func splitShape(text string) []string {
	return shapeSplit.Split(text, -1)
}

func extractStrings(data []byte) []string {
	if len(data) >= 8 && bytes.HasPrefix(data, []byte("mdf\x00")) {
		reader, err := zlib.NewReader(bytes.NewReader(data[8:]))
		if err != nil {
			return nil
		}
		inflated, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return nil
		}
		out := extractStrings(inflated)
		for _, match := range utf8Token.FindAll(inflated, -1) {
			out = appendUnique(out, string(match))
		}
		return out
	}
	if len(data) >= 8 && bytes.HasPrefix(data, []byte("TJS2100\x00")) {
		values, err := tjs2.Strings(data)
		if err != nil {
			return nil
		}
		return values
	}
	if len(data) >= 3 && bytes.HasPrefix(data, []byte("PSB")) {
		if values, ok := psbMineStrings(data); ok {
			return values
		}
		return nil
	}
	if isMediaMagic(data) {
		return nil
	}
	if plain, ok := descrambleScript(data); ok {
		text := string(plain)
		out := harvestRuns(plain)
		for _, v := range tjsTextHints(text) {
			out = appendUnique(out, v)
		}
		for _, v := range sceneFileHints(plain) {
			out = appendUnique(out, v)
		}
		for _, v := range tableFileHints(plain) {
			out = appendUnique(out, v)
		}
		return out
	}
	if len(data) >= 2 && (bytes.HasPrefix(data, []byte{0xff, 0xfe}) || bytes.HasPrefix(data, []byte{0xfe, 0xff})) {
		return harvestRuns(data)
	}
	if utf8.Valid(data) {
		text := string(data)
		out := make([]string, 0, 64)
		for _, m := range utf8Token.FindAllString(text, -1) {
			out = append(out, m)
		}
		for _, v := range sceneFileHints(data) {
			out = appendUnique(out, v)
		}
		for _, v := range tableFileHints(data) {
			out = appendUnique(out, v)
		}
		return out
	}
	if len(data) > 256<<10 {
		return nil
	}
	out := harvestRuns(data)
	for _, v := range tlgRefHints(data) {
		out = appendUnique(out, v)
	}
	return out
}

func addFileTarget(out map[string]string, domain string, file string, ts *targetSet) {
	file = normalizeFile(file)
	if file == "" || strings.Contains(file, "/") {
		return
	}
	hash := fileHash(file, domain)
	if !ts.files[hash] {
		return
	}
	if _, ok := out[hash]; !ok {
		out[hash] = file
	}
}

func addDirTarget(out map[string]string, domain string, dir string, ts *targetSet) {
	dir = normalizeDir(dir)
	if dir == "" {
		return
	}
	hash := pathHash(dir, domain)
	if !ts.dirs[hash] {
		return
	}
	if _, ok := out[hash]; !ok {
		out[hash] = dir
	}
}

func addNameHintT(out map[string]string, domain string, raw string, ts *targetSet) {
	name, isDir, ok := cleanName(raw)
	if !ok {
		return
	}
	if isDir {
		addDirTarget(out, domain, name, ts)
		return
	}
	dir, file := splitName(name)
	if dir != "" {
		addDirTarget(out, domain, dir, ts)
	}
	if hasExt(file) {
		addFileTarget(out, domain, file, ts)
		addFileTarget(out, domain, file+".scn", ts)
		return
	}
	if strings.Contains(name, "/") {
		if looksLikeKnownDir(name) {
			addDirTarget(out, domain, name, ts)
		}
		if looksLikeStem(file) {
			for _, ext := range knownExts {
				addFileTarget(out, domain, file+ext, ts)
			}
		}
		return
	}
	if looksLikeStem(file) {
		addFileTarget(out, domain, file, ts)
		for _, ext := range knownExts {
			addFileTarget(out, domain, file+ext, ts)
		}
	}
}

type mineResult struct {
	paths []string
	stems []string
}

func mineShape(out map[string]string, root string, domain string, ts *targetSet) mineResult {
	var res mineResult
	seen := map[string]bool{}
	const seenCap = 3_000_000
	filepath.WalkDir(root, func(p string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() > 64<<20 {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		raws := extractStrings(data)
		for _, raw := range raws {
			if len(seen) >= seenCap {
				break
			}
			for _, tok := range append([]string{raw}, splitShape(raw)...) {
				if len(tok) < 2 || len(tok) > 300 {
					continue
				}
				if seen[tok] {
					continue
				}
				seen[tok] = true
				name, isDir, ok := cleanName(tok)
				if !ok {
					continue
				}
				if isDir {
					addDirTarget(out, domain, name, ts)
					continue
				}
				if strings.Contains(name, "/") {
					dir, file := splitName(name)
					if dir != "" {
						addDirTarget(out, domain, dir, ts)
						res.paths = append(res.paths, dir)
					}
					addNameHintT(out, domain, name, ts)
					if looksLikeStem(file) {
						res.paths = append(res.paths, name)
					}
					continue
				}
				if looksLikeStem(name) || hasExt(name) {
					addNameHintT(out, domain, name, ts)
					if looksLikeStem(name) && !hasExt(name) {
						res.stems = append(res.stems, normalizeFile(name))
					}
				}
			}
		}
		return nil
	})
	unique := res.stems[:0]
	seenStem := map[string]bool{}
	for _, s := range res.stems {
		if !seenStem[s] {
			seenStem[s] = true
			unique = append(unique, s)
		}
	}
	res.stems = unique
	return res
}
