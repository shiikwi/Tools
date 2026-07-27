package main

import (
	"LucaSystemGO/opcode"
	"LucaSystemGO/scheme"
	"LucaSystemGO/util"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		PrintUsage()
		return
	}

	chocie := os.Args[1]

	if chocie == "-d" {
		dir := os.Args[2]
		entries, _ := os.ReadDir(dir)

		var opinfos []opcode.CodeLine
		var ops = make(map[byte]bool)

		for _, entry := range entries {
			filePath := filepath.Join(dir, entry.Name())
			data, _ := os.ReadFile(filePath)

			opinfo, op, err := opcode.ParseInfo(data)
			if err != nil {
				fmt.Println(err)
				return
			}

			opinfos = append(opinfos, opinfo...)
			for k, v := range op {
				ops[k] = v
			}
		}
		WriteOpInfo(opinfos, ops)
		return
	}

	if os.Args[2] == "-e" {
		err := ExportSC()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Export finish")
	}
}

func WriteOpInfo(opinfo []opcode.CodeLine, ops map[byte]bool) {
	opfile, _ := os.Create("OPCODE.txt")
	defer opfile.Close()

	keys := slices.Sorted(maps.Keys(ops))
	for _, k := range keys {
		fmt.Fprintf(opfile, "0x%02X:\n", k)
	}

	infofile, _ := os.Create("OPINFO.txt")
	defer infofile.Close()

	for _, info := range opinfo {
		fmt.Fprintf(
			infofile,
			//Offset, Size, Opcode, Flag, Extra, Payload
			"0x%08X, 0x%04X, 0x%02X, 0x%02X, %s, %s\n",
			info.Offset,
			info.Size,
			info.Opcode,
			info.Flag,
			util.Bytes2HexStr(info.Extra),
			util.Bytes2HexStr(info.Payload),
		)
	}
}

func PrintUsage() {
	fmt.Println("Usage: LucaSystemGO.exe -d <scriptfolderpath>      scan opcodes")
	fmt.Println("       LucaSystemGO.exe -gamename -e <scriptfolderpath>    export game's script")
	fmt.Println("SupportGameList:")

	names := make([]string, 0, len(scheme.Games))
	for name := range scheme.Games {
		names = append(names, name)
	}

	sort.Strings(names)
	for _, name := range names {
		fmt.Printf(" - %s\n", name)
	}
}

func ExportSC() error {
	gameName := strings.TrimPrefix(os.Args[1], "-")
	newGameVM, ok := scheme.Games[gameName]
	if !ok {
		PrintUsage()
		return fmt.Errorf("Unsupport game")
	}

	opNames, err := opcode.LoadOpcodeTxt("OPCODE.txt")
	if err != nil {
		return err
	}
	game := newGameVM()

	dir := os.Args[3]
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		scriptPath := filepath.Join(dir, entry.Name())

		outPath := filepath.Join(dir, entry.Name()+".txt")
		outfile, _ := os.Create(outPath)

		data, _ := os.ReadFile(scriptPath)
		lines, _, _ := opcode.ParseInfo(data)

		for _, line := range lines {
			name := opNames[line.Opcode]
			if name != "" {
				text, handled, err := game.ExportScript(name, line.Extra, line.Payload)
				if err != nil {
					return fmt.Errorf("vm execute fail, %s", err)
				} else if handled {
					fmt.Fprintf(outfile, "%s(%s)(%s)\n",
						name,
						util.Bytes2HexStr(line.Extra),
						text,
					)
					continue
				} else {
					fmt.Fprintf(outfile, "%s(%s)(%s)\n",
						name,
						util.Bytes2HexStr(line.Extra),
						text,
					)
					continue
				}
			}
			fmt.Fprintf(outfile, "OP_%02X(%s)(%s)\n",
				line.Opcode,
				util.Bytes2HexStr(line.Extra),
				util.Bytes2HexStr(line.Payload),
			)
		}

		fmt.Printf("Export %s Success\n", entry.Name())
		outfile.Close()
	}

	return nil
}
