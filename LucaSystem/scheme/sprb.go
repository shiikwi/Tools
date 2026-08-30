package scheme

import (
	"LucaSystemGO/opcode"
	"LucaSystemGO/util"
	"fmt"
)

type SPRBVM struct {
	opcode.VM
}

func (g *SPRBVM) ExportScript(name string, extra []byte, payload []byte) (string, bool, error) {
	g.Reset(payload)
	var text string = ""
	var err error = nil
	switch name {
	case "MESSAGE":
		text, err = g.MESSAGE()
	case "SELECT":
		text, err = g.SELECT()
	}
	if err != nil {
		return "", false, err
	}
	if text != "" {
		return text, true, nil
	}

	return util.Bytes2HexStr(payload), false, nil
}

func (g *SPRBVM) MESSAGE() (string, error) {
	voiceID, err := g.ReadU16()
	if err != nil {
		return "", err
	}

	jp_text, err := g.ReadLenUTF16()
	if err != nil {
		return "", err
	}

	var en_text string
	var cn_text string
	if jp_text != "" {
		if _, err = g.ReadU16(); err != nil {
			return "", err
		}
		if _, err = g.ReadU16(); err != nil {
			return "", err
		}

		en_text, err = g.ReadUTF82NULL()
		if err != nil {
			return "", err
		}

		cn_text, err = g.ReadLenUTF16()
		if err != nil {
			return "", err
		}
	}

	result := fmt.Sprintf(
		"\nVoice%d\n"+
			"%s\n"+
			"%s\n"+
			"%s\n",
		voiceID,
		jp_text,
		en_text,
		cn_text)

	return result, nil
}

func (g *SPRBVM) SELECT() (string, error) {
	g.ReadU16()
	g.ReadU16()
	g.ReadU16()
	g.ReadU16()

	jp_text, err := g.ReadLenUTF16()
	if err != nil {
		return "", err
	}

	var en_text string
	var cn_text string
	if jp_text != "" {
		en_text, err = g.ReadLenUTF16()
		if err != nil {
			return "", err
		}

		cn_text, err = g.ReadLenUTF16()
		if err != nil {
			return "", err
		}
	}

	result := fmt.Sprintf(
		"\n%s\n"+
			"%s\n"+
			"%s\n",
		jp_text,
		en_text,
		cn_text,
	)

	return result, nil
}
