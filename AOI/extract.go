package main

import (
	"aoiTool/box"
	"fmt"
	"os"
	"path/filepath"
)

func ExtractBox(boxPath, outDir string) ([]string, error) {
	data, err := os.ReadFile(boxPath)
	if err != nil {
		return nil, err
	}
	scripts, err := box.ParseBox(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", boxPath, err)
	}

	var names []string
	for _, sc := range scripts {
		names = append(names, sc.Name)
		sc.Decrypt()

		relPath := filepath.Join(outDir, sc.Name)
		if err := os.WriteFile(relPath, sc.Data, 0o644); err != nil {
			return nil, err
		}
	}

	return names, nil
}
