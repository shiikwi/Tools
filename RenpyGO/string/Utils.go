package Utils

import (
	"strconv"
	"unicode"
	"unicode/utf8"
)

func RuneAt(s string, i int) (rune, int) {
	r, size := utf8.DecodeRuneInString(s[i:])
	return r, size
}

func ReadString(s string, start int, quote byte) (string, int) {
	escaped := false
	for i := start + 1; i < len(s); i++ {
		c := s[i]
		if escaped {
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == quote {
			return s[start : i+1], i + 1
		}
	}
	return s[start:], len(s)
}

func UnQuote(raw string) string {
	text, _ := strconv.Unquote(raw)
	return text
}

func IsIdentStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func IsIdentPart(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
