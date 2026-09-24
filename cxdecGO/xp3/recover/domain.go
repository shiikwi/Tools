package recover

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func resolveDomains(out map[string]string, root string, domain string, ts *targetSet, paths []string, stems []string, deep bool) {
	archives := map[string]bool{}
	entries, err := os.ReadDir(root)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.IsDir() {
			archives[strings.ToLower(e.Name())] = true
		}
	}

	asciiLower := func(s string) string {
		b := []rune(s)
		for i, r := range b {
			if r >= 'A' && r <= 'Z' {
				b[i] = r + 32
			}
		}
		return string(b)
	}

	seen := map[string]bool{}
	try := func(cand string) {
		cand = normalizeDir(cand)
		if cand == "" || seen[cand] {
			return
		}
		seen[cand] = true
		hash := pathHash(cand, domain)
		if ts.dirs[hash] {
			if _, ok := out[hash]; !ok {
				out[hash] = cand
			}
		}
	}

	prefixes := map[string]bool{}
	for _, p := range paths {
		for _, v := range []string{p, asciiLower(p)} {
			parts := strings.Split(strings.Trim(v, "/"), "/")
			if len(parts) == 0 {
				continue
			}
			head, rest := parts[0], strings.Join(parts[1:], "/")
			if archives[head] && rest != "" {
				try(rest)
			}
			try(v)
			acc := ""
			for _, seg := range parts[:len(parts)-1] {
				acc += seg + "/"
				try(acc)
				prefixes[strings.ToLower(acc)] = true
			}
		}
	}

	words := map[string]bool{}
	for _, p := range paths {
		for _, seg := range strings.Split(p, "/") {
			if seg != "" && len(seg) <= 40 && looksLikeStem(seg) {
				words[strings.ToLower(seg)] = true
			}
		}
	}
	for _, stem := range stems {
		if len(stem) <= 40 && looksLikeStem(stem) {
			words[stem] = true
		}
	}
	topDirs := map[string]bool{}
	for arc := range archives {
		de, err := os.ReadDir(filepath.Join(root, arc))
		if err != nil {
			continue
		}
		for _, d := range de {
			if d.IsDir() && !isHexStr(strings.ToLower(d.Name())) && len(d.Name()) <= 40 {
				topDirs[strings.ToLower(d.Name())] = true
			}
		}
	}
	for w := range words {
		try(w)
		for arc := range archives {
			try(arc + "/" + w)
		}
		for t := range topDirs {
			try(t + "/" + w)
		}
	}

	if !deep {
		return
	}
	pl := make([]string, 0, len(prefixes))
	for p := range prefixes {
		pl = append(pl, p)
	}
	sort.Slice(pl, func(i, j int) bool { return len(pl[i]) < len(pl[j]) })
	if len(pl) > 2000 {
		pl = pl[:2000]
	}
	wl := make([]string, 0, len(words))
	for w := range words {
		if len(w) <= 24 {
			wl = append(wl, w)
		}
	}
	sort.Strings(wl)
	tried := 0
	for _, pre := range pl {
		for _, w := range wl {
			try(pre + w)
			tried++
			if tried >= 20_000_000 {
				return
			}
		}
	}
}
