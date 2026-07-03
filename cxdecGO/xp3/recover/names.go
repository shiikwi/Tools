package recover

import (
	"path"
	"strconv"
	"strings"
)

const DefaultDomain = "xp3hnp"

var knownExts = []string{
	".tjs", ".ks", ".scn", ".txt", ".csv", ".ini", ".func", ".pimg", ".pbd", ".sli", ".psb",
	".ttf", ".otf", ".ttc", ".woff", ".woff2", ".asd", ".png", ".tlg", ".tlg5", ".tlg6", ".jpg",
	".jpeg", ".bmp", ".webp", ".emf", ".stage", ".stand", ".sinfo", ".event", ".psd", ".ogg",
	".ogg.sli", ".opus", ".wav", ".tcw", ".amv", ".webm", ".mp4", ".avi", ".wmv",
}

var knownDirs = []string{
	"k2compat", "system", "sysscn", "video", "others", "rule", "sound", "bgm", "fgimage",
	"bgimage", "scenario", "scn", "image", "voice", "face", "init", "font", "sysse", "main",
	"evimage", "thum", "uipsd", "motion", "motiondx", "emote", "emotedx", "bishamon", "video2", "fgimage2",
}

func addNameHint(out map[string]string, domain, raw string) {
	name, isDir, ok := cleanName(raw)
	if !ok {
		return
	}
	if isDir {
		addDir(out, domain, name)
		return
	}

	dir, file := splitName(name)
	if dir != "" {
		addDir(out, domain, dir)
	}
	if hasExt(file) {
		addFile(out, domain, file)
		return
	}
	if strings.Contains(name, "/") {
		if looksLikeKnownDir(name) {
			addDir(out, domain, name)
		}
		if looksLikeStem(file) {
			for _, ext := range knownExts {
				addFile(out, domain, file+ext)
			}
		}
		return
	}
	if looksLikeStem(file) {
		addFile(out, domain, file)
		for _, ext := range knownExts {
			addFile(out, domain, file+ext)
		}
	}
}

func addDir(out map[string]string, domain, dir string) {
	dir = normalizeDir(dir)
	out[pathHash(dir, domain)] = dir
}

func addFile(out map[string]string, domain, file string) {
	file = normalizeFile(file)
	if file != "" && !strings.Contains(file, "/") {
		out[fileHash(file, domain)] = file
	}
}

func cleanName(raw string) (string, bool, bool) {
	value := strings.TrimSpace(raw)
	value = strings.Trim(value, "\"'")
	value = strings.Trim(value, ",;:")
	value = strings.Trim(value, "[](){}")
	value = strings.ReplaceAll(value, "\\", "/")
	if i := strings.IndexByte(value, '|'); i >= 0 {
		value = value[:i]
	}
	if value == "" || len(value) > 180 || strings.ContainsAny(value, "?<>") {
		return "", false, false
	}
	if strings.ContainsAny(value[:1], "#@&*%$") {
		return "", false, false
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return "", false, false
	}
	isDir := strings.HasSuffix(value, "/")
	value = strings.TrimPrefix(value, "./")
	value = strings.TrimPrefix(value, "/")
	value = normalizePath(value)
	if value == "" {
		return "", false, false
	}
	if isDir {
		value = normalizeDir(value)
	}
	return value, isDir, true
}

func normalizePath(value string) string {
	var parts []string
	for _, part := range strings.Split(strings.ReplaceAll(value, "\\", "/"), "/") {
		switch part {
		case "", ".":
		case "..":
			if len(parts) != 0 {
				parts = parts[:len(parts)-1]
			}
		default:
			parts = append(parts, strings.ToLower(part))
		}
	}
	return strings.Join(parts, "/")
}

func normalizeDir(value string) string {
	value = normalizePath(value)
	if value == "" {
		return ""
	}
	return value + "/"
}

func normalizeFile(value string) string {
	return strings.ToLower(path.Base(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/")))
}

func splitName(value string) (string, string) {
	value = normalizePath(value)
	i := strings.LastIndexByte(value, '/')
	if i < 0 {
		return "", value
	}
	return normalizeDir(value[:i]), value[i+1:]
}

func hasExt(file string) bool {
	i := strings.LastIndexByte(file, '.')
	return i > 0 && i+1 < len(file)
}

func looksLikeStem(file string) bool {
	if len(file) < 2 || len(file) > 80 {
		return false
	}
	for _, ch := range file {
		if ch == '_' || ch == '-' || ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch > 127 {
			continue
		}
		return false
	}
	return true
}

func looksLikeKnownDir(value string) bool {
	first := strings.Split(value, "/")[0]
	for _, dir := range knownDirs {
		if first == dir {
			return true
		}
	}
	return false
}
