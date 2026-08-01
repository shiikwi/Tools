package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ManifestEntry struct {
	Box      string   `json:"box"`
	TXTNames []string `json:"txt_name"`
}

func main() {
	paths := os.Args[1:]
	defer func() {
		fmt.Println("Press any key exit...")
		fmt.Scanln()
	}()

	var manifest []ManifestEntry
	var outDir string
	for _, filePath := range paths {
		switch strings.ToLower(filepath.Ext(filePath)) {
		case ".box":
			outDir = filepath.Join(filepath.Dir(filePath), "extBox")
			os.MkdirAll(outDir, 0o755)
			txtNames, err := ExtractBox(filePath, outDir)
			if err != nil {
				fmt.Println(err)
			}

			manifest = append(manifest, ManifestEntry{
				Box:      filepath.Base(filePath),
				TXTNames: txtNames,
			})

			fmt.Printf("Extract %s\n", filepath.Base(filePath))
		case ".txt":
			if err := DisassembleFile(filePath); err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Disassemble %s\n", filepath.Base(filePath))
		}
	}

	if len(manifest) > 0 {
		jsonData, _ := json.MarshalIndent(manifest, "", "  ")
		os.WriteFile(filepath.Join(outDir, "Manifest.json"), append(jsonData, '\n'), 0o644)
	}
}
