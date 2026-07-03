package recover

import (
	"strings"
	"unicode"
)

func scanResourceTables(root string, domain string) (map[string]string, error) {
	return scanFiles(root, domain, 2<<20, addNameHint, func(data []byte) []string {
		text, ok := decodeScriptText(data)
		if !ok {
			text, ok = decodeText(data)
		}
		if !ok {
			return nil
		}
		return resourceTableHints(text)
	})
}

func resourceTableHints(text string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	rows := 0
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}

	for _, line := range strings.Split(text, "\n") {
		fields := tableFields(line)
		if len(fields) == 1 {
			name, _ := tableName(fields[0])
			if name == "" {
				continue
			}
			switch {
			case hasKnownExt(name):
				rows++
				add(name)
			case likelyImageStem(name):
				rows++
				for _, candidate := range tableImageCandidates(name) {
					add(candidate)
				}
			}
			continue
		}
		if len(fields) < 2 {
			continue
		}
		kind := tableKind(fields)
		if kind == "" {
			continue
		}
		rows++

		for i, field := range fields {
			name, hadScale := tableName(field)
			if name == "" {
				continue
			}
			if hasKnownExt(name) {
				add(name)
				continue
			}
			switch kind {
			case "image":
				if i == 0 || hadScale || likelyImageStem(name) {
					for _, candidate := range tableImageCandidates(name) {
						add(candidate)
					}
				}
			case "audio":
				if i == 0 {
					for _, candidate := range tableCandidates(name, []string{".ogg", ".opus", ".wav", ".tcw", ".ogg.sli"}) {
						add(candidate)
					}
				}
			case "video":
				if i == 0 {
					for _, candidate := range tableCandidates(name, []string{".mp4", ".webm", ".amv", ".avi", ".wmv"}) {
						add(candidate)
					}
				}
			}
		}
	}
	if rows < 2 {
		return nil
	}
	return out
}

func tableFields(line string) []string {
	line = strings.TrimSpace(strings.TrimSuffix(line, "\r"))
	if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, ";") {
		return nil
	}
	if strings.HasPrefix(line, "#") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if line == "" || !commentResourceLine(line) {
			return nil
		}
	}
	fields := strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == '\t' || r == ';' || unicode.IsSpace(r)
	})
	if len(fields) == 0 {
		return nil
	}
	return fields
}

func tableKind(fields []string) string {
	first, _ := tableName(fields[0])
	if audioStem(first) {
		return "audio"
	}
	if videoStem(first) {
		return "video"
	}
	if likelyImageStem(first) {
		return "image"
	}

	imageScore := 0
	videoScore := 0
	audioScore := 0
	for _, field := range fields {
		name, hadScale := tableName(field)
		if name == "" {
			continue
		}
		lower := strings.ToLower(name)
		if hadScale || likelyImageStem(name) || hasAnySuffix(lower, ".png", ".tlg", ".pimg", ".psd", ".psb", ".stage", ".stand", ".event") {
			imageScore++
		}
		if audioStem(name) || hasAnySuffix(lower, ".ogg", ".opus", ".wav", ".tcw", ".ogg.sli") {
			audioScore++
		}
		if videoStem(name) || hasAnySuffix(lower, ".mp4", ".webm", ".amv", ".avi", ".wmv") {
			videoScore++
		}
	}
	switch {
	case imageScore >= 2:
		return "image"
	case audioScore >= 1:
		return "audio"
	case videoScore >= 1:
		return "video"
	default:
		return ""
	}
}

func tableName(field string) (string, bool) {
	field = strings.TrimSpace(field)
	field = strings.Trim(field, "\"'[](){}")
	hadScale := false
	if left, right, ok := strings.Cut(field, ":"); ok && digitsOnly(left) {
		field = right
		hadScale = true
	}
	name, isDir, ok := cleanName(field)
	if !ok || isDir {
		return "", false
	}
	_, file := splitName(name)
	if !hasKnownExt(name) && !looksLikeStem(file) {
		return "", false
	}
	return name, hadScale
}

func tableImageCandidates(name string) []string {
	exts := []string{".png", ".tlg", ".pimg", ".psd"}
	lower := strings.ToLower(name)
	if strings.Contains(lower, "emote") || strings.Contains(lower, "motion") {
		exts = append(exts, ".psb")
	}
	return tableCandidates(name, exts)
}

func tableCandidates(name string, exts []string) []string {
	if hasKnownExt(name) {
		return []string{name}
	}
	out := make([]string, 0, len(exts))
	for _, ext := range exts {
		out = appendUnique(out, name+ext)
	}
	return out
}

func likelyImageStem(name string) bool {
	_, file := splitName(name)
	file = strings.ToLower(file)
	return strings.Contains(file, "thum_") ||
		strings.HasPrefix(file, "ev") && hasDigit(file[2:]) ||
		prefixedStem(file, "bg") ||
		prefixedStem(file, "fg") ||
		strings.HasPrefix(file, "image_") ||
		strings.HasPrefix(file, "face_")
}

func audioStem(name string) bool {
	_, file := splitName(name)
	file = strings.ToLower(file)
	return prefixedStem(file, "bgm") ||
		prefixedStem(file, "se") ||
		strings.HasPrefix(file, "sysse") ||
		strings.HasPrefix(file, "voice") ||
		strings.HasPrefix(file, "vo_")
}

func videoStem(name string) bool {
	_, file := splitName(name)
	file = strings.ToLower(file)
	return prefixedStem(file, "mov") ||
		strings.HasPrefix(file, "movie") ||
		strings.HasPrefix(file, "video") ||
		file == "op" ||
		file == "ed" ||
		strings.HasPrefix(file, "op_") ||
		strings.HasPrefix(file, "ed_")
}

func prefixedStem(file, prefix string) bool {
	if !strings.HasPrefix(file, prefix) {
		return false
	}
	if len(file) == len(prefix) {
		return true
	}
	next := file[len(prefix)]
	return next == '_' || next >= '0' && next <= '9'
}

func commentResourceLine(line string) bool {
	for _, field := range strings.FieldsFunc(line, func(r rune) bool {
		return r == ',' || r == '\t' || r == ';' || unicode.IsSpace(r)
	}) {
		name, _ := tableName(field)
		if name != "" && (hasKnownExt(name) || likelyImageStem(name) || audioStem(name) || videoStem(name)) {
			return true
		}
	}
	return false
}

func hasDigit(value string) bool {
	for _, ch := range value {
		if ch >= '0' && ch <= '9' {
			return true
		}
	}
	return false
}

func hasAnySuffix(value string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func digitsOnly(value string) bool {
	if value == "" {
		return false
	}
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}
