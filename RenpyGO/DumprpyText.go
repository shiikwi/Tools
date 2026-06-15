package main

import (
	Utils "RenpyGO/string"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenIdent tokenKind = iota
	tokenString
	tokenColon
	tokenOther
)

type token struct {
	kind tokenKind
	text string
	raw  string
}

type textItem struct {
	line    int
	kind    string
	speaker string
	text    string
}

type Extractor struct {
	commands map[string]bool
}

func NewExtractor() *Extractor {
	return &Extractor{
		commands: commandWords,
	}
}

var commandWords = map[string]bool{
	"$":         true,
	"call":      true,
	"default":   true,
	"define":    true,
	"elif":      true,
	"else":      true,
	"for":       true,
	"hide":      true,
	"if":        true,
	"image":     true,
	"init":      true,
	"jump":      true,
	"label":     true,
	"menu":      true,
	"pause":     true,
	"play":      true,
	"python":    true,
	"queue":     true,
	"return":    true,
	"scene":     true,
	"screen":    true,
	"show":      true,
	"stop":      true,
	"transform": true,
	"voice":     true,
	"while":     true,
	"with":      true,
}

func (e *Extractor) dumpText(inFile string) []textItem {
	f, _ := os.Open(inFile)
	defer f.Close()

	reader := bufio.NewReader(f)
	var items []textItem
	lineNo := 0
	for {
		line, err := reader.ReadString('\n')
		if len(line) == 0 && err != nil {
			if err == io.EOF {
				break
			}
			return nil
		}

		lineNo++
		line = strings.TrimRight(line, "\r\n")
		if item, ok := e.ExtractLine(line, lineNo); ok {
			items = append(items, item)
		}
	}
	return items
}

func (e *Extractor) ExtractLine(line string, lineNo int) (textItem, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return textItem{}, false
	}
	tokens := lex(trimmed)
	if tokens[0].kind == tokenString {
		if len(tokens) > 1 && tokens[1].kind == tokenColon {
			return textItem{line: lineNo, kind: "menu", text: tokens[0].text}, true
		}
		return textItem{line: lineNo, kind: "narrator", text: tokens[0].text}, true
	}

	first := tokens[0].text
	if first == "old" || first == "new" || first == "extend" {
		if s, ok := firstStr(tokens[1:]); ok {
			return textItem{line: lineNo, kind: first, text: s.text}, true
		}
		return textItem{}, false
	}

	if e.commands[first] {
		return textItem{}, false
	}

	stringIndex := -1
	for i, tok := range tokens {
		if tok.kind == tokenString {
			stringIndex = i
			break
		}
	}

	if stringIndex < 0 {
		return textItem{}, false
	}

	for _, tok := range tokens[:stringIndex] {
		if tok.kind != tokenIdent {
			return textItem{}, false
		}
	}

	return textItem{
		line:    lineNo,
		kind:    "say",
		speaker: first,
		text:    tokens[stringIndex].text,
	}, true
}

func lex(s string) []token {
	var tokens []token
	for i := 0; i < len(s); {
		r, size := Utils.RuneAt(s, i)
		if unicode.IsSpace(r) {
			i += size
			continue
		}

		if r == '#' {
			break
		}

		if r == ':' {
			tokens = append(tokens, token{kind: tokenColon, text: ":"})
			i += size
			continue
		}

		if r == '"' || r == '\'' {
			raw, next := Utils.ReadString(s, i, byte(r))
			text := Utils.UnQuote(raw)
			tokens = append(tokens, token{kind: tokenString, text: text, raw: raw})
			i = next
			continue
		}

		if Utils.IsIdentStart(r) {
			start := i
			i += size
			for i < len(s) {
				r, size = Utils.RuneAt(s, i)
				if !Utils.IsIdentPart(r) {
					break
				}
				i += size
			}
			tokens = append(tokens, token{kind: tokenIdent, text: s[start:i]})
			continue
		}

		tokens = append(tokens, token{kind: tokenOther, text: string(r)})
		i += size
	}
	return tokens
}

func firstStr(tokens []token) (token, bool) {
	for _, tok := range tokens {
		if tok.kind == tokenString {
			return tok, true
		}
	}
	return token{}, false
}

func WriteTxt(items []textItem, outPath string, ifdumpMeta bool) {
	var out *os.File
	out, _ = os.Create(outPath)
	defer out.Close()

	w := bufio.NewWriter(out)
	for _, item := range items {
		if ifdumpMeta {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", item.line, item.kind, item.speaker, item.text)
		} else {
			fmt.Fprintln(w, item.text)
		}
	}
	w.Flush()
}

func main() {
	rpyExtract := NewExtractor()
	for _, inFile := range os.Args[1:] {
		var outPath string = filepath.Join(filepath.Dir(inFile), filepath.Base(inFile)+".txt")
		items := rpyExtract.dumpText(inFile)
		WriteTxt(items, outPath, true) //change false to write text only
	}
}
