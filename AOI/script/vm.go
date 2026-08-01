package script

import (
	"aoiTool/util"
	"encoding/binary"
	"fmt"
	"strings"
)

func Disassemble(data []byte) (string, error) {
	var out strings.Builder
	pc := 0
	for pc < len(data) {
		offset := pc
		opcode, err := readByte(data, &pc)
		if err != nil {
			return "", err
		}

		args := make([]string, opcodeArgs[opcode])
		for i := range args {
			args[i], err = decodeOperand(data, &pc)
			if err != nil {
				return "", fmt.Errorf("%04X %s arg %d: %w", offset, opcodeName(opcode), i+1, err)
			}
		}

		fmt.Fprintf(&out, "◇0x%04X◇ %s(%s)\n", offset, opcodeName(opcode), strings.Join(args, ", "))
	}
	return out.String(), nil
}

func decodeOperand(data []byte, pc *int) (string, error) {
	tag, err := readByte(data, pc)
	if err != nil {
		return "", err
	}

	switch {
	case tag == 0:
		return "null", nil
	case tag&2 != 0:
		return decodeNumber(data, pc, tag)
	case tag&4 != 0:
		return decodeBool(data, pc, tag)
	case tag&8 != 0:
		return decodeString(data, pc, tag)
	case tag&1 != 0:
		text, err := readString(data, pc)
		if err != nil {
			return "", err
		}
		return "expr(" + quoteText(text) + ")", nil
	default:
		return "", fmt.Errorf("unknown tag %02X", tag)
	}
}

func decodeNumber(data []byte, pc *int, tag byte) (string, error) {
	var value string
	if tag&0x80 != 0 {
		number, err := readU32(data, pc)
		if err != nil {
			return "", err
		}
		value = fmt.Sprintf("%d", int32(number))
	} else {
		index, err := readU16(data, pc)
		if err != nil {
			return "", err
		}
		value = variableName("C", tag, index)
	}

	for tag&1 != 0 {
		tag, err := readByte(data, pc)
		if err != nil {
			return "", err
		}
		index, err := readU16(data, pc)
		if err != nil {
			return "", err
		}
		value = fmt.Sprintf("(%s+%s)", value, variableName("C", tag, index))
	}
	return value, nil
}

func decodeBool(data []byte, pc *int, tag byte) (string, error) {
	if tag&0x80 != 0 {
		value, err := readByte(data, pc)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%t", value != 0), nil
	}
	index, err := readU16(data, pc)
	if err != nil {
		return "", err
	}
	return variableName("F", tag, index), nil
}

func decodeString(data []byte, pc *int, tag byte) (string, error) {
	if tag&0x80 != 0 {
		value, err := readString(data, pc)
		if err != nil {
			return "", err
		}
		return quoteText(value), nil
	}
	index, err := readU16(data, pc)
	if err != nil {
		return "", err
	}
	return variableName("S", tag, index), nil
}

func variableName(kind string, tag byte, index uint16) string {
	switch {
	case tag&0x40 != 0:
		kind += "G"
	case tag&0x20 != 0:
		kind += "L"
	case tag&0x10 != 0:
		kind += "P"
	}
	return fmt.Sprintf("%s%d", kind, index)
}

func readByte(data []byte, pc *int) (byte, error) {
	if *pc >= len(data) {
		return 0, fmt.Errorf("unexpected end of script")
	}
	value := data[*pc]
	(*pc)++
	return value, nil
}

func readU16(data []byte, pc *int) (uint16, error) {
	if len(data)-*pc < 2 {
		return 0, fmt.Errorf("unexpected end of script")
	}
	value := binary.BigEndian.Uint16(data[*pc : *pc+2])
	*pc += 2
	return value, nil
}

func readU32(data []byte, pc *int) (uint32, error) {
	if len(data)-*pc < 4 {
		return 0, fmt.Errorf("unexpected end of script")
	}
	value := binary.BigEndian.Uint32(data[*pc : *pc+4])
	*pc += 4
	return value, nil
}

func readString(data []byte, pc *int) (string, error) {
	length, err := readByte(data, pc)
	if err != nil {
		return "", err
	}
	size := int(length) * 2
	if len(data)-*pc < size {
		return "", fmt.Errorf("unexpected end of script")
	}
	value := util.DecodeUTF16(data[*pc : *pc+size])
	*pc += size
	return value, nil
}

func quoteText(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "\r", "\\r")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return "\"" + value + "\""
}
