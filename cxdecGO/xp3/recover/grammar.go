package recover

import (
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type grammarSeg struct {
	alpha string
	digW  int
}

type grammarTpl struct {
	segs  []grammarSeg
	vals  [][]int
	exts  map[string]bool
	space uint64
}

var grammarRuns = regexp.MustCompile(`(\d+|[^0-9]+)`)

const (
	grammarTplSpaceCap = 20_000_000
	grammarTriesCap    = 100_000_000
)

func sweepGrammar(verified map[string]string, out map[string]string, domain string, ts *targetSet) {
	templates := map[string]*grammarTpl{}
	for _, name := range verified {
		dot := strings.LastIndexByte(name, '.')
		if dot <= 0 || dot == len(name)-1 {
			continue
		}
		low := strings.ToLower(name)
		stem, ext := low[:dot], low[dot:]
		var segs []grammarSeg
		var key strings.Builder
		var vals []int
		for _, run := range grammarRuns.FindAllString(stem, -1) {
			if isDigitRun(run) {
				v, _ := strconv.Atoi(run)
				vals = append(vals, v)
				segs = append(segs, grammarSeg{digW: len(run)})
				key.WriteString("<D" + strconv.Itoa(len(run)) + ">")
			} else {
				segs = append(segs, grammarSeg{alpha: run})
				key.WriteString(run)
			}
		}
		if len(vals) == 0 || len(vals) > 4 {
			continue
		}
		tpl, ok := templates[key.String()]
		if !ok {
			tpl = &grammarTpl{segs: segs, exts: map[string]bool{}}
			templates[key.String()] = tpl
		}
		tpl.exts[ext] = true
		tpl.vals = append(tpl.vals, vals)
	}

	ordered := make([]*grammarTpl, 0, len(templates))
	for _, tpl := range templates {
		_, space := grammarRanges(tpl.vals)
		tpl.space = space
		ordered = append(ordered, tpl)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].space < ordered[j].space })

	var tried atomic.Int64
	for _, tpl := range ordered {
		if tpl.space > grammarTplSpaceCap {
			continue
		}
		ranges, _ := grammarRanges(tpl.vals)
		for ext := range tpl.exts {
			grammarEnumerate(tpl.segs, ranges, ext, out, domain, ts, &tried)
			if tried.Load() >= grammarTriesCap {
				return
			}
		}
	}
}

func grammarSeedStems(stems []string, out map[string]string, domain string, ts *targetSet) {
	sort.Slice(stems, func(i, j int) bool {
		if len(stems[i]) != len(stems[j]) {
			return len(stems[i]) < len(stems[j])
		}
		return stems[i] < stems[j]
	})
	var tried, hits atomic.Int64
	var mu sync.Mutex
	try := func(name string) {
		hash := fileHash(name, domain)
		if !ts.files[hash] {
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if _, known := out[hash]; known {
			return
		}
		out[hash] = name
		hits.Add(1)
	}

	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	var next atomic.Int64
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := int(next.Add(1)) - 1
				if i >= len(stems) || tried.Load() >= grammarTriesCap {
					return
				}
				stem := stems[i]
				if len(stem) < 2 || len(stem) > 8 {
					continue
				}
				before := hits.Load()
				grammarStemDigits(stem, "", try, &tried)
				if hits.Load() > before {
					for c := byte('a'); c <= 'z'; c++ {
						grammarStemDigits(stem, string(c), try, &tried)
						if tried.Load() >= grammarTriesCap {
							return
						}
					}
				}
			}
		}()
	}
	wg.Wait()
}

func grammarStemDigits(stem, suffix string, try func(string), tried *atomic.Int64) {
	for _, ext := range []string{".ogg", ".opus", ".png", ".tlg"} {
		for width := 1; width <= 4; width++ {
			maxn := 1
			for i := 0; i < width; i++ {
				maxn *= 10
			}
			for n := 0; n < maxn; n++ {
				digits := strconv.Itoa(n)
				for len(digits) < width {
					digits = "0" + digits
				}
				tried.Add(1)
				try(stem + digits + suffix + ext)
				if tried.Load() >= grammarTriesCap {
					return
				}
			}
		}
	}
}

func grammarRanges(vals [][]int) ([][2]int, uint64) {
	ndig := len(vals[0])
	ranges := make([][2]int, ndig)
	space := uint64(1)
	for pos := 0; pos < ndig; pos++ {
		lo, hi := 1<<31-1, 0
		for _, v := range vals {
			if len(v) != ndig {
				continue
			}
			if v[pos] < lo {
				lo = v[pos]
			}
			if v[pos] > hi {
				hi = v[pos]
			}
		}
		if lo > hi {
			lo, hi = 0, 9
		}
		span := (hi-lo)*3/2 + 500
		if span < 9 {
			span = 9
		}
		if span > 50000 {
			span = 50000
		}
		low := lo - span/3
		if low < 0 {
			low = 0
		}
		ranges[pos] = [2]int{low, hi + span}
		space *= uint64(ranges[pos][1] - ranges[pos][0] + 1)
	}
	return ranges, space
}

func grammarEnumerate(segs []grammarSeg, ranges [][2]int, ext string, out map[string]string, domain string, ts *targetSet, tried *atomic.Int64) {
	var rec func(segPos, digPos int, prefix []byte)
	rec = func(segPos, digPos int, prefix []byte) {
		if tried.Load() >= grammarTriesCap {
			return
		}
		if segPos == len(segs) {
			tried.Add(1)
			addFileTarget(out, domain, string(prefix)+ext, ts)
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

func isDigitRun(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}
