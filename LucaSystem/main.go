package main

import (
	"LucaSystemGO/opcode"
	"LucaSystemGO/scheme"
	"LucaSystemGO/util"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
)

func main() {
	gameName, dir, scan, old := ParseArgs()

	if scan {
		ScanOpcodes(dir, old)
		return
	}

	err := ExportSC(gameName, dir, old)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Export finish")
}

func ParseArgs() (string, string, bool, bool) {
	fs := flag.NewFlagSet("LucaSystemGO", flag.ExitOnError)
	fs.Usage = func() {
		PrintUsage(fs)
	}

	scanDir := fs.String("d", "", "scan opcodes")
	scriptDir := fs.String("e", "", "export game's script")
	old := fs.Bool("old", false, "use old extra format")
	gameFlags := make(map[string]*bool)
	for name := range scheme.Games {
		gameFlags[name] = fs.Bool(name, false, "export game's script")
	}

	fs.Parse(os.Args[1:])
	gameName := ""
	for name, value := range gameFlags {
		if *value {
			gameName = name
		}
	}
	if *scanDir != "" {
		return "", *scanDir, true, *old
	}
	return gameName, *scriptDir, false, *old
}

func ScanOpcodes(dir string, old bool) {
	entries, _ := os.ReadDir(dir)

	var opinfos []opcode.CodeLine
	var ops = make(map[byte]bool)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())
		data, _ := os.ReadFile(filePath)

		opinfo, op, err := opcode.ParseInfo(data, old)
		if err != nil {
			fmt.Printf("Skip %s\n", entry.Name())
			continue
		}

		opinfos = append(opinfos, opinfo...)
		for k, v := range op {
			ops[k] = v
		}
	}
	WriteOpInfo(opinfos, ops)
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

func PrintUsage(fs *flag.FlagSet) {
	fmt.Fprintln(fs.Output(), "Usage: LucaSystemGO.exe -d <scriptfolderpath>      scan opcodes")
	fmt.Fprintln(fs.Output(), "       LucaSystemGO.exe -gamename -e <scriptfolderpath>    export game's script")
	fmt.Fprintln(fs.Output(), "       add -old to use ScriptReader's extra format")
	fmt.Fprintln(fs.Output(), "SupportGameList:")

	names := make([]string, 0, len(scheme.Games))
	for name := range scheme.Games {
		names = append(names, name)
	}

	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(fs.Output(), " -%s\n", name)
	}
}

func ExportSC(gameName, dir string, old bool) error {
	newGameVM, ok := scheme.Games[gameName]
	if !ok {
		return fmt.Errorf("Unsupport game")
	}

	opNames, err := opcode.LoadOpcodeTxt("OPCODE.txt")
	if err != nil {
		return err
	}
	game := newGameVM()

	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		scriptPath := filepath.Join(dir, entry.Name())

		outPath := filepath.Join(dir, entry.Name()+".txt")
		outfile, _ := os.Create(outPath)

		data, _ := os.ReadFile(scriptPath)
		lines, _, _ := opcode.ParseInfo(data, old)

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
