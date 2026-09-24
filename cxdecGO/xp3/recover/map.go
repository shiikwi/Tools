package recover

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type targetSet struct {
	files map[string]bool
	dirs  map[string]bool
}

func isHexStr(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return len(s) > 0
}

func collectTargets(root string) *targetSet {
	ts := &targetSet{files: map[string]bool{}, dirs: map[string]bool{}}
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		base := d.Name()
		if d.IsDir() {
			if len(base) == 16 && isHexStr(base) {
				ts.dirs[base] = true
			}
			return nil
		}
		if len(base) == 64 && isHexStr(base) {
			ts.files[base] = true
		}
		return nil
	})
	return ts
}

func matchedTargets(ts *targetSet, out map[string]string) int {
	n := 0
	for h := range ts.files {
		if _, ok := out[h]; ok {
			n++
		}
	}
	for h := range ts.dirs {
		if _, ok := out[h]; ok {
			n++
		}
	}
	return n
}

func Build(root string, base map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for hash, name := range base {
		out[strings.ToLower(hash)] = name
	}
	ts := collectTargets(root)
	total := len(ts.files) + len(ts.dirs)

	var firstErr error
	for round := 1; round <= 5; round++ {
		before := matchedTargets(ts, out)
		if names, err := defaultNames(root, DefaultDomain); err == nil {
			for hash, name := range names {
				if _, ok := out[hash]; !ok {
					out[hash] = name
				}
			}
		}
		res := mineShape(out, root, DefaultDomain, ts)
		fmt.Printf("mineShape match %d names\n", matchedTargets(ts, out))
		resolveDomains(out, root, DefaultDomain, ts, res.paths, res.stems, round >= 2)
		fmt.Printf("resolveDomains match %d names\n", matchedTargets(ts, out))
		derivePBD(out, root, DefaultDomain, ts)
		fmt.Printf("derivePBD match %d names\n", matchedTargets(ts, out))
		verified := map[string]string{}
		for h := range ts.files {
			if name, ok := out[h]; ok {
				verified[h] = name
			}
		}
		sweepGrammar(verified, out, DefaultDomain, ts)
		grammarSeedStems(res.stems, out, DefaultDomain, ts)
		interpolateNeighbors(root, out, DefaultDomain, ts)
		now := matchedTargets(ts, out)
		fmt.Printf("round %d: %d/%d names matched\n", round, now, total)
		if now == before {
			break
		}
	}
	return out, firstErr
}

func RestoreDirs(outputRoot string, names map[string]string) int {
	var dirs []string
	filepath.WalkDir(outputRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() || p == outputRoot {
			return nil
		}
		base := d.Name()
		if len(base) == 16 && isHexStr(base) {
			dirs = append(dirs, p)
		}
		return nil
	})
	applied := 0
	for _, dir := range dirs {
		name, ok := names[filepath.Base(dir)]
		if !ok {
			continue
		}
		rel := strings.Trim(name, "/")
		if rel == "" {
			continue
		}
		dst := filepath.Join(filepath.Dir(dir), filepath.FromSlash(rel))
		if dst == dir {
			continue
		}
		os.MkdirAll(filepath.Dir(dst), 0755)
		if _, err := os.Stat(dst); err == nil {
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				src := filepath.Join(dir, e.Name())
				d2 := filepath.Join(dst, e.Name())
				if _, err := os.Stat(d2); err == nil {
					continue
				}
				os.Rename(src, d2)
			}
			os.Remove(dir)
			applied++
			continue
		}
		if err := os.Rename(dir, dst); err == nil {
			applied++
		}
	}
	return applied
}

func Resolve(names map[string]string, domainHash, fileHash string) (string, bool) {
	domainHash = strings.ToLower(domainHash)
	fileHash = strings.ToLower(fileHash)
	dir, ok := names[domainHash]
	if !ok {
		return "", false
	}
	file, ok := names[fileHash]
	if !ok {
		return "", false
	}
	return dir + file, true
}

func ResolvePartial(names map[string]string, domainHash, fileHash string) (string, bool, bool) {
	domainHash = strings.ToLower(domainHash)
	fileHash = strings.ToLower(fileHash)

	dir, ok := names[domainHash]
	if !ok {
		return "", false, false
	}
	file, full := names[fileHash]
	if !full {
		file = fileHash
	}
	return dir + file, true, full
}

func defaultNames(_ string, domain string) (map[string]string, error) {
	out := map[string]string{}
	addDir(out, domain, "")
	addDir(out, domain, "data/")
	for _, dir := range knownDirs {
		addDir(out, domain, dir+"/")
	}
	for _, name := range []string{"!scnlist.txt", "scn/!scnlist.txt", "scenario/!scnlist.txt"} {
		addNameHint(out, domain, name)
	}
	return out, nil
}

func LoadList(path string) (map[string]string, error) {
	out := map[string]string{}
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return out, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields := strings.Split(line, ":")
		name := fields[len(fields)-1]
		if name == "" {
			continue
		}
		if len(fields) >= 4 {
			if h := strings.ToLower(fields[1]); h != "" {
				out[h] = name
			}
			if strings.Contains(name, "/") {
				if h := strings.ToLower(fields[0]); h != "" {
					out[h] = name[:strings.LastIndex(name, "/")+1]
				}
			}
			continue
		}
		hash := strings.ToLower(fields[0])
		if hash == "" {
			continue
		}
		out[hash] = name
	}
	return out, scanner.Err()
}

func SaveList(path string, names map[string]string) error {
	keys := make([]string, 0, len(names))
	for hash := range names {
		keys = append(keys, hash)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, hash := range keys {
		b.WriteString(strings.ToUpper(hash))
		b.WriteByte(':')
		b.WriteString(names[hash])
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func fileNameCount(names map[string]string) int {
	count := 0
	for hash := range names {
		if len(hash) == 64 {
			count++
		}
	}
	return count
}
