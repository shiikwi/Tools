package main

import (
	"LucaSystemGO/opcode"
	"LucaSystemGO/util"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

func main() {
	data, _ := os.ReadFile(os.Args[1])
	opinfo, ops, err := opcode.ParseInfo(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	WriteOpInfo(opinfo, ops, filepath.Dir(os.Args[1]))
}

func WriteOpInfo(opinfo []opcode.CodeLine, ops map[byte]bool, dirPath string) {
	opfile, _ := os.Create(filepath.Join(dirPath, "OPCODE.txt"))
	defer opfile.Close()

	keys := slices.Sorted(maps.Keys(ops))
	for k := range keys {
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
