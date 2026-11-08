package bootstrapdump

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cxdecgo/internal/bres"
	"cxdecgo/internal/pefile"
	"cxdecgo/internal/tjs2"
)

const (
	bootstrapZlibOffset = 8
)

var (
	startupResource   = [2]pefile.ResourceName{pefile.ID(10), pefile.Text("STARTUP.TJS")}
	bootstrapResource = [2]pefile.ResourceName{pefile.ID(10), pefile.Text("BOOTSTRAP")}
	pluginResource    = [2]pefile.ResourceName{pefile.ID(10), pefile.Text("PLUGIN")}
	textResource      = [2]pefile.ResourceName{pefile.Text("TEXT"), pefile.ID(127)}
)

type Outputs struct {
	WorkDir         string `json:"work_dir"`
	BresSalt        string `json:"bres_salt"`
	StartupCipher   string `json:"startup_cipher"`
	StartupDec      string `json:"startup_dec"`
	BootstrapCipher string `json:"bootstrap_cipher"`
	BootstrapDec    string `json:"bootstrap_dec"`
	DLL             string `json:"dll"`
	PluginCipher    string `json:"plugin_cipher"`
}

type Result struct {
	Exe              string `json:"exe"`
	StartupKey       string `json:"startup_key"`
	BootstrapURL     string `json:"bootstrap_url"`
	BootstrapKey     string `json:"bootstrap_key"`
	BootstrapPrefix  string `json:"bootstrap_prefix"`
	Warning          string `json:"warning"`
	FinalBootstrap   string `json:"final_bootstrap"`
	ArchiveUniqueKey string `json:"archive_unique_key"`
}

func Run(exePath string) (*Result, error) {
	fmt.Println("==========dump bootstrap phase==========")

	absExe, err := filepath.Abs(exePath)
	if err != nil {
		return nil, err
	}
	img, err := pefile.Open(absExe)
	if err != nil {
		return nil, err
	}
	resources, err := img.Resources()
	if err != nil {
		return nil, err
	}

	startupCipher, err := pefile.FindResource(resources, startupResource[0], startupResource[1])
	if err != nil {
		return nil, err
	}
	fmt.Println("Find startup success")

	bootstrapCipher, err := pefile.FindResource(resources, bootstrapResource[0], bootstrapResource[1])
	if err != nil {
		return nil, err
	}
	fmt.Println("Find bootstrap success")

	textData, err := pefile.FindResource(resources, textResource[0], textResource[1])
	if err != nil {
		return nil, err
	}

	pluginCipher, _ := pefile.FindResource(resources, pluginResource[0], pluginResource[1])
	startupKey, err := decodeBresRoot(textData)
	if err != nil {
		return nil, err
	}

	salt, startupPlain, err := recoverSaltAndStartup(img, startupCipher, startupKey)
	if err != nil {
		return nil, err
	}

	stringsPool, err := tjs2.Strings(startupPlain)
	if err != nil {
		return nil, err
	}
	bootstrapURL, err := findBootstrapURL(stringsPool)
	if err != nil {
		return nil, err
	}
	bootstrapKey, err := bresKeyFromURL(bootstrapURL)
	if err != nil {
		return nil, err
	}
	bootstrapPlain := bres.Decrypt(bootstrapCipher, bootstrapKey, salt)
	dllBytes, err := inflateBootstrapDLL(bootstrapPlain)
	if err != nil {
		return nil, err
	}

	workDir := filepath.Join(filepath.Dir(absExe), "bootstrap_dump")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(workDir, "bootstrap.dll"), dllBytes, 0644); err != nil {
		return nil, err
	}

	params, err := findConfigEntry(dllBytes, "PARAMS")
	if err != nil {
		return nil, err
	}
	uniqueRaw, err := findConfigEntry(dllBytes, "UNIQUE")
	if err != nil {
		return nil, err
	}
	warningRaw, err := findConfigEntry(dllBytes, "WARNING")
	if err != nil {
		return nil, err
	}
	unique := decodeUTF16LE(uniqueRaw)
	warning := decodeASCII(warningRaw)
	prefix, err := findBootstrapPrefix(stringsPool)
	if err != nil {
		return nil, err
	}

	finalBootstrap := ""
	if prefix != "" {
		finalBootstrap = prefix + warning
	}

	outputs := Outputs{
		WorkDir:         workDir,
		BresSalt:        filepath.Join(workDir, "bres_salt.bin"),
		StartupCipher:   filepath.Join(workDir, "STARTUP_TJS.enc"),
		StartupDec:      filepath.Join(workDir, "STARTUP.TJS.dec"),
		BootstrapCipher: filepath.Join(workDir, "BOOTSTRAP.enc"),
		BootstrapDec:    filepath.Join(workDir, "BOOTSTRAP.dec"),
		DLL:             filepath.Join(workDir, "bootstrap.dll"),
	}
	if pluginCipher != nil {
		outputs.PluginCipher = filepath.Join(workDir, "PLUGIN.bin")
	}

	if err := writeOutputs(outputs, salt, startupCipher, startupPlain, bootstrapCipher, bootstrapPlain, dllBytes, pluginCipher); err != nil {
		return nil, err
	}

	result := &Result{
		Exe:              absExe,
		StartupKey:       startupKey,
		BootstrapURL:     bootstrapURL,
		BootstrapKey:     bootstrapKey,
		BootstrapPrefix:  prefix,
		Warning:          warning,
		FinalBootstrap:   finalBootstrap,
		ArchiveUniqueKey: unique,
	}
	if err := writeSummary(filepath.Join(workDir, "bootstrap.json"), result); err != nil {
		return nil, err
	}

	if err := writeHxv4(workDir, dllBytes, params, unique, warning, prefix, filepath.Dir(absExe)); err != nil {
		return nil, err
	}

	return result, nil
}

func recoverSaltAndStartup(img *pefile.Image, startupCipher []byte, startupKey string) ([]byte, []byte, error) {
	candidates := autoSaltCandidates(img)
	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("could not locate bres salt")
	}
	for _, candidate := range candidates {
		plain := bres.Decrypt(startupCipher, startupKey, candidate.salt)
		if bytes.HasPrefix(plain, []byte("TJS2100\x00")) {
			fmt.Println("decrypt startup success")
			return candidate.salt, plain, nil
		}
	}
	return nil, nil, fmt.Errorf("no auto salt candidate decrypted STARTUP.TJS")
}

func decodeBresRoot(textData []byte) (string, error) {
	text := decodeUTF16LE(textData)
	text = strings.TrimRight(text, "\x00")
	const marker = "bres://./"
	if !strings.HasPrefix(text, marker) {
		return "", fmt.Errorf("unexpected bres root: %q", text)
	}
	return strings.Trim(text[len(marker):], "/"), nil
}

func findBootstrapURL(stringsPool []string) (string, error) {
	for _, value := range stringsPool {
		if strings.HasPrefix(value, "bres://./") && strings.HasSuffix(strings.ToLower(value), "/bootstrap") {
			return value, nil
		}
	}
	return "", fmt.Errorf("could not find bootstrap bres URL in STARTUP.TJS strings")
}

func bresKeyFromURL(url string) (string, error) {
	const marker = "bres://./"
	if !strings.HasPrefix(url, marker) {
		return "", fmt.Errorf("not a local bres URL: %q", url)
	}
	rest := url[len(marker):]
	return strings.SplitN(rest, "/", 2)[0], nil
}

func findBootstrapPrefix(stringsPool []string) (string, error) {
	var candidates []string
	seen := map[string]bool{}
	for i, value := range stringsPool {
		if !strings.EqualFold(value, "startup.tjs") || i < 2 || stringsPool[i-1] != "" {
			continue
		}
		prefix := stringsPool[i-2]
		if prefix == "" || seen[prefix] {
			continue
		}
		seen[prefix] = true
		candidates = append(candidates, prefix)
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	for i, value := range stringsPool {
		if !strings.EqualFold(value, "startup.tjs") {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			prefix := stringsPool[j]
			if prefix == "" || strings.EqualFold(prefix, "object") || strings.EqualFold(prefix, "startup.tjs") {
				continue
			}
			if seen[prefix] {
				break
			}
			seen[prefix] = true
			candidates = append(candidates, prefix)
			break
		}
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	return "", fmt.Errorf("failed find BootstrapPrefix")
}

func inflateBootstrapDLL(bootstrapPlain []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(bootstrapPlain[bootstrapZlibOffset:]))
	if err != nil {
		return nil, fmt.Errorf("zlib init fail: %w", err)
	}
	defer reader.Close()
	dll, err := io.ReadAll(reader)
	if !bytes.HasPrefix(dll, []byte("MZ")) {
		return nil, fmt.Errorf("decompress BOOTSTRAP fail")
	}
	fmt.Println("dump BootstrapDLL success")
	return dll, nil
}

func writeOutputs(outputs Outputs, salt, startupCipher, startupPlain, _, _, dllBytes, pluginCipher []byte) error {
	files := map[string][]byte{
		outputs.BresSalt:      salt,
		outputs.StartupCipher: startupCipher,
		outputs.StartupDec:    startupPlain,
		// outputs.BootstrapCipher: bootstrapCipher,
		// outputs.BootstrapDec:    bootstrapPlain,
		outputs.DLL: dllBytes,
	}
	if pluginCipher != nil && outputs.PluginCipher != "" {
		files[outputs.PluginCipher] = pluginCipher
	}
	for path, data := range files {
		if data != nil {
			if err := os.WriteFile(path, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeSummary(path string, result *Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}
