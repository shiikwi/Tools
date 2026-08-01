package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ManifestEntry struct {
	Box      string   `json:"box"`
	TXTNames []string `json:"txt_name"`
}

func main() {
	paths := os.Args[1:]

	var manifest []ManifestEntry
	var outDir string
	for _, filePath := range paths {
		if filepath.Ext(filePath) == ".box" {
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
		}

	}

	if len(manifest) > 0 {
		jsonData, _ := json.MarshalIndent(manifest, "", "  ")
		os.WriteFile(filepath.Join(outDir, "Manifest.json"), append(jsonData, '\n'), 0o644)
	}
}
