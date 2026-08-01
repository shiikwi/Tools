package box

import (
	"aoiTool/script"
	"aoiTool/util"
	"encoding/binary"
	"fmt"
	"strings"
)

func ParseBox(data []byte) ([]script.TXTScript, error) {
	if util.DecodeUTF16(data[:0x10]) != "AOIMY01\x00" {
		return nil, fmt.Errorf("invalid box sig")
	}
	count := binary.BigEndian.Uint32(data[0x10:0x14])
	//0x00000000 0x14:0x18

	pos := 0x18
	var scripts []script.TXTScript
	for i := uint32(0); i < count; i++ {
		entryPos := uint32(pos) + i*0x28
		txtName := strings.TrimRight(util.DecodeUTF16(data[entryPos:entryPos+0x20]), "\x00")
		if end := strings.IndexByte(txtName, 0); end > 0 {
			txtName = txtName[:end]
		}
		offset := binary.BigEndian.Uint32(data[entryPos+0x20 : entryPos+0x24])
		size := binary.BigEndian.Uint32(data[entryPos+0x24 : entryPos+0x28])

		scripts = append(scripts, script.TXTScript{
			Name:   txtName,
			Offset: offset,
			Size:   size,
			Data:   data[offset : offset+size],
		})
	}
	return scripts, nil
}
