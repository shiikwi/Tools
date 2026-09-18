package recover

import (
	"os"
	"strings"
)

func resolveDomains(out map[string]string, root string, domain string, ts *targetSet, paths []string) {
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
	for w := range words {
		try(w)
		for arc := range archives {
			try(arc + "/" + w)
		}
	}
}
