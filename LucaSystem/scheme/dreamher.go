package scheme

import (
	"LucaSystemGO/opcode"
	"LucaSystemGO/util"
	"fmt"
)

type DreamHerVM struct {
	opcode.VM
}

func (g *DreamHerVM) ExportScript(name string, extra []byte, payload []byte) (string, bool, error) {
	g.Reset(payload)

	switch name {
	case "MESSAGE":
		text, err := g.MESSAGE()
		if err != nil {
			return "", false, err
		}
		return text, true, nil
	}

	return util.Bytes2HexStr(payload), false, nil
}

func (g *DreamHerVM) MESSAGE() (string, error) {
	voiceID, err := g.ReadU16()
	if err != nil {
		return "", err
	}

	jp_text, err := g.ReadLenUTF16()
	if err != nil {
		return "", err
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
			"%02X%02X%02X\n",
		voiceID,
		jp_text,
		langSplitMark,
		en_text,
		tailType, tailValue, tailZero,
	)
	return result, nil
}
