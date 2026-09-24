package recover

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
)

type manifestRecord struct {
	Index      int    `json:"index"`
	DomainHash string `json:"domain_hash"`
	FileHash   string `json:"file_hash"`
	Output     string `json:"output"`
	Status     string `json:"status"`
}

var manifestDigit = regexp.MustCompile(`(\d+|[^0-9]+)`)

func interpolateNeighbors(outputRoot string, out map[string]string, domain string, ts *targetSet) {
	entries, err := os.ReadDir(outputRoot)
	if err != nil {
		return
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(outputRoot, e.Name(), "manifest.json")
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var manifest struct {
			Records []manifestRecord `json:"records"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}
		byDomain := map[string][]manifestRecord{}
		for _, r := range manifest.Records {
			h := lower(r.FileHash)
			if h == "" || r.Status != "written" {
				continue
			}
			d := lower(r.DomainHash)
			byDomain[d] = append(byDomain[d], r)
		}
		for _, recs := range byDomain {
			sort.Slice(recs, func(i, j int) bool { return recs[i].Index < recs[j].Index })
			for i := 0; i < len(recs); i++ {
				if _, ok := out[lower(recs[i].FileHash)]; ok {
					continue
				}
				lo, hi := i-1, i+1
				for lo >= 0 && out[lower(recs[lo].FileHash)] == "" {
					lo--
				}
				for hi < len(recs) && out[lower(recs[hi].FileHash)] == "" {
					hi++
				}
				if lo < 0 || hi >= len(recs) {
					continue
				}
				grammarGap(out, domain, ts,
					out[lower(recs[lo].FileHash)],
					out[lower(recs[hi].FileHash)],
					hi-lo-1)
			}
		}
	}
}

func grammarGap(out map[string]string, domain string, ts *targetSet, lo, hi string, gap int) {
	if gap > 64 {
		return
	}
	lost, _ := splitName(lo)
	hied, _ := splitName(hi)
	if lost != hied {
		lost = hied
	}
	loStem, loExt := splitExt(lost)
	hiStem, hiExt := splitExt(hied)
	if loExt != hiExt {
		return
	}
	loRuns := manifestDigit.FindAllString(strings.ToLower(loStem), -1)
	hiRuns := manifestDigit.FindAllString(strings.ToLower(hiStem), -1)
	if len(loRuns) != len(hiRuns) {
		return
	}
	var segs []grammarSeg
	var loVals, hiVals []int
	var key strings.Builder
	loParts := manifestDigit.FindAllString(strings.ToLower(loStem), -1)
	hiParts := manifestDigit.FindAllString(strings.ToLower(hiStem), -1)
	for k := range loParts {
		a, b := loParts[k], hiParts[k]
		if a == b {
			segs = append(segs, grammarSeg{alpha: a})
			key.WriteString(a)
			continue
		}
		if !isDigitRun(a) || !isDigitRun(b) || len(a) != len(b) {
			return
		}
		av, _ := strconv.Atoi(a)
		bv, _ := strconv.Atoi(b)
		if bv <= av {
			return
		}
		loVals = append(loVals, av)
		hiVals = append(hiVals, bv)
		segs = append(segs, grammarSeg{digW: len(a)})
		key.WriteString("<D" + strconv.Itoa(len(a)) + ">")
	}
	if len(loVals) == 0 || len(loVals) > 2 {
		return
	}
	ranges := make([][2]int, len(loVals))
	space := uint64(1)
	for k := range loVals {
		ranges[k] = [2]int{loVals[k], hiVals[k]}
		space *= uint64(hiVals[k] - loVals[k] + 1)
		if space > 1_000_000 {
			return
		}
	}
	var tried atomic.Int64
	var rec func(segPos, digPos int, prefix []byte)
	rec = func(segPos, digPos int, prefix []byte) {
		if tried.Load() >= grammarTriesCap {
			return
		}
		if segPos == len(segs) {
			tried.Add(1)
			addFileTarget(out, domain, string(prefix)+loExt, ts)
			return
		}
		seg := segs[segPos]
		if seg.digW == 0 {
			rec(segPos+1, digPos, append(prefix, seg.alpha...))
			return
		}
		for n := ranges[digPos][0]; n <= ranges[digPos][1]; n++ {
			digits := strconv.Itoa(n)
			for len(digits) < seg.digW {
				digits = "0" + digits
			}
			rec(segPos+1, digPos+1, append(prefix, digits...))
		}
	}
	rec(0, 0, nil)
}

func splitExt(name string) (string, string) {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[:i], name[i:]
		}
	}
	return name, ""
}
