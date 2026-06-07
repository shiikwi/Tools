package main

import (
	models "NaniNovelTool/struct"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ReadJson(inFile string) error {
	data, err := os.ReadFile(inFile)
	if err != nil {
		return fmt.Errorf("Failed read file: %w", err)
	}

	var scenario models.Scenario
	if err := json.Unmarshal(data, &scenario); err != nil {
		return fmt.Errorf("Failed Parse Json: %w", err)
	}

	var outPath string = filepath.Join(filepath.Dir(inFile), scenario.Name+".txt")
	file, _ := os.Create(outPath)

	defer file.Close()

	for _, line := range scenario.TextMap.IdToText.Values {
		file.WriteString(line + "\n")
	}

	fmt.Printf("Dump %s Success\n", scenario.Name)
	return nil
}

func main() {
	for _, inFile := range os.Args[1:] {
		ReadJson(inFile)
	}
}
