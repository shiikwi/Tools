package scheme

import (
	"LucaSystemGO/opcode"
	"LucaSystemGO/util"
	"fmt"
	"strings"
)

type usonatsuVM struct {
	opcode.VM
}

func (g *usonatsuVM) ExportScript(name string, extra []byte, payload []byte) (string, bool, error) {
	g.Reset(payload)
	var text string = ""
	var err error = nil
	switch name {
	case "MESSAGE":
		text, err = g.MESSAGE()
	case "SCENARIO":
		text, err = g.SCENARIO()
	}
	if err != nil {
		return "", false, err
	}
	if text != "" {
		return text, true, nil
	}

	return util.Bytes2HexStr(payload), false, nil
}

func (g *usonatsuVM) MESSAGE() (string, error) {
	voiceID, err := g.ReadU16()
	if err != nil {
		return "", err
	}

	jp_text, err := g.ReadLenUTF16()
	if err != nil {
		return "", err
	}

	if jp_text == "" {
		return "", nil //some script has redundant text
	}

	//0x0000 jp_text
	_, err = g.ReadU16()
	if err != nil {
		return "", err
	}

	langSplitMark, err := g.ReadU16()
	if err != nil {
		return "", err
	}

	en_text, err := g.ReadUTF82NULL()
	if err != nil {
		return "", err
	}

	cn_text, err := g.ReadLenUTF16()
	if err != nil {
		return "", err
	}

	//0x0000 cn_text
	_, err = g.ReadU16()
	if err != nil {
		return "", err
	}

	tailType, err := g.ReadByte()
	if err != nil {
		return "", err
	}
	tailValue, err := g.ReadU16()
	if err != nil {
		return "", err
	}
	tailZero, err := g.ReadU16()
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf(
		"\nVoice%d\n"+
			"%s\n"+
			"%02X\n"+
			"%s\n"+
			"%s\n"+
			"%02X%02X%02X\n",
		voiceID,
		jp_text,
		langSplitMark,
		en_text,
		cn_text,
		tailType, tailValue, tailZero,
	)

	return result, nil
}

func (g *usonatsuVM) SCENARIO() (string, error) {
	g.ReadU32()
	flag, _ := g.ReadU32()
	g.ReadU32()
	g.ReadU32()
	g.ReadByte()

	var strs []string
	var first bool = true
	var count int = 8
	if flag != 0 {
		count = 5
	}
	for i := 0; i < count; i++ {
		s, err := g.ReadUTF82NULL()
		if err != nil {
			return "", err
		}

		if first && flag == 0 {
			g.ReadByte() //skip 0x00
			first = false
		}

		strs = append(strs, s)

		if g.GetPos() >= g.GetDataLen() {
			break
		}
	}

	result := fmt.Sprintf(
		"\n%s\n",
		strings.Join(strs, ", "),
	)

	return result, nil
}
