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
	chocie := os.Args[1]

	if chocie == "-d" {
		data, _ := os.ReadFile(os.Args[2])
		opinfo, ops, err := opcode.ParseInfo(data)
		if err != nil {
			fmt.Println(err)
			return
		}

		WriteOpInfo(opinfo, ops, filepath.Dir(os.Args[1]))
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

func WriteOpInfo(opinfo []opcode.CodeLine, ops map[byte]bool, dirPath string) {
	opfile, _ := os.Create(filepath.Join(dirPath, "OPCODE.txt"))
	defer opfile.Close()

	keys := slices.Sorted(maps.Keys(ops))
	for _, k := range keys {
		fmt.Fprintf(opfile, "0x%02X:\n", k)
	}

	infofile, _ := os.Create(filepath.Join(dirPath, "OPINFO.txt"))
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
	fmt.Println("Usage: LucaSystemGO.exe -d <scriptpath>      scan opcodes")
	fmt.Println("       LucaSystemGO.exe -gamename -e <scriptpath>    export game's script")
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
	script, err := os.ReadFile(os.Args[3])
	if err != nil {
		return err
	}

	opNames, err := opcode.LoadOpcodeTxt("OPCODE.txt")
	if err != nil {
		return err
	}

	outPath := filepath.Join(filepath.Dir(os.Args[3]), filepath.Base(os.Args[3])+".txt")
	oufile, _ := os.Create(outPath)

	lines, _, err := opcode.ParseInfo(script)
	game := newGameVM()

	for _, line := range lines {
		name := opNames[line.Opcode]
		if name != "" {
			text, handled, err := game.ExportScript(name, line.Extra, line.Payload)
			if err != nil {
				return fmt.Errorf("vm execute fail")
			} else if handled {
				fmt.Fprintf(oufile, "%s(%s)(%s)\n",
					name,
					util.Bytes2HexStr(line.Extra),
					text,
				)
				continue
			} else {
				fmt.Fprintf(oufile, "%s(%s)(%s)\n",
					name,
					util.Bytes2HexStr(line.Extra),
					text,
				)
				continue
			}
		}
		fmt.Fprintf(oufile, "OP_%02X(%s)(%s)\n",
			line.Opcode,
			util.Bytes2HexStr(line.Extra),
			util.Bytes2HexStr(line.Payload),
		)
	}
	return nil
}
