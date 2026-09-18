package recover

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func derivePBD(out map[string]string, root string, domain string, ts *targetSet) {
	filepath.WalkDir(root, func(p string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		base := entry.Name()
		if len(base) != 64 || !isHexStr(base) {
			return nil
		}
		name, ok := out[base]
		if !ok {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".pbd" && ext != ".sinfo" {
			return nil
		}
		info, err := entry.Info()
		if err != nil || info.Size() > 8<<20 {
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return nil
		}
		if ext == ".sinfo" {
			if plain, ok := descrambleScript(data); ok {
				data = plain
			}
		}
		strs, ints, _ := pbdDecode(data)
		if ext == ".sinfo" && len(ints) == 0 {
			for _, raw := range harvestRuns(data) {
				strs = append(strs, raw)
			}
		}
		container := strings.TrimSuffix(name, filepath.Ext(name))
		for _, iv := range ints {
			for _, cand := range []string{
				container + "_" + strconv.Itoa(iv),
				container + "_" + pad4(iv),
			} {
				addFileTarget(out, domain, cand+".tlg", ts)
				addFileTarget(out, domain, cand+".png", ts)
				addFileTarget(out, domain, cand+filepath.Ext(name), ts)
			}
		}
		for _, s := range strs {
			s = strings.TrimSpace(s)
			if s == "" || len(s) > 40 || strings.ContainsAny(s, "/\\") {
				continue
			}
			if !looksLikeStem(s) {
				continue
			}
			addFileTarget(out, domain, container+"_"+s+".tlg", ts)
			addFileTarget(out, domain, container+"_"+s+".png", ts)
		}
		return nil
	})
}

func pad4(v int) string {
	s := strconv.Itoa(v)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}
