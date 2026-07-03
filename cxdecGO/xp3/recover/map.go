package recover

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Finder func(root string, domain string) (map[string]string, error)

var finders = []struct {
	name string
	fn   Finder
}{
	{"defaultNames", defaultNames},
	{"scanSceneList", scanSceneList},
	{"scanTJS", scanTJS},
	{"scanResourceTables", scanResourceTables},
	{"scanPSB", scanPSB},
}

func Build(root string, base map[string]string) (map[string]string, error) {
	out := map[string]string{}
	for hash, name := range base {
		out[strings.ToLower(hash)] = name
	}

	var firstErr error
	for _, finder := range finders {
		before := fileNameCount(out)
		names, err := finder.fn(root, DefaultDomain)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		for hash, name := range names {
			if _, ok := out[hash]; !ok {
				out[hash] = name
			}
		}
		fmt.Printf("%s get %d file names\n", finder.name, fileNameCount(out)-before)
	}
	return out, firstErr
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
		hash, name, ok := strings.Cut(line, ":")
		if !ok || hash == "" || name == "" {
			continue
		}
		out[strings.ToLower(hash)] = name
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
