package main

import "encoding/binary"

func le16(buf []byte, off int) uint16 {
	return binary.LittleEndian.Uint16(buf[off : off+2])
}

func le32(buf []byte, off int) uint32 {
	return binary.LittleEndian.Uint32(buf[off : off+4])
}
