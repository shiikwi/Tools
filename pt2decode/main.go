package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Pt2Decoder struct{}

func NewPt2Decoder() *Pt2Decoder {
	return &Pt2Decoder{}
}

func (pt2d *Pt2Decoder) DecodePt2(inFile string) {
	data, _ := os.ReadFile(inFile)
	header := Pt2Header{
		Version:    le16(data, 0),
		HeaderSize: le16(data, 2),
		CanvasW:    le16(data, 4),
		CanvasH:    le16(data, 6),
		Aux:        le32(data, 8),
	}
	if header.Version != 0x0100 && header.Version != 0x0101 && header.Version != 0x0102 {
		log.Fatalf("unsupported pt2version %d", header.Version)
	}

	dataOff1 := le32(data, int(header.HeaderSize+4))
	tileCount := (dataOff1 - uint32(header.HeaderSize)) / 16

	pt2Tiles := make([]Pt2Tile, 0, tileCount)
	for i := 0; i < int(tileCount); i++ {
		off := int(header.HeaderSize) + 16*i
		tile := Pt2Tile{
			Flag:       le32(data, off),
			DataOff:    le32(data, off+4),
			CompSize:   le32(data, off+8),
			TileWidth:  le16(data, off+12),
			TileHeight: le16(data, off+14),
		}

		pt2Tiles = append(pt2Tiles, tile)
	}

	tilesDecoded := make([][]byte, 0)
	for _, tile := range pt2Tiles {
		payload := data[tile.DataOff : tile.DataOff+tile.CompSize]
		outSize := int(tile.TileWidth) * int(tile.TileHeight) * 4

		var decoded []byte
		switch header.Version {
		case 0x0100:
			decoded = pt2d.DecodeLZ77(payload, int(outSize))
		case 0x0101, 0x0102:
			decoded = pt2d.DecodeLzss(payload, int(outSize))
		}

		switch header.Aux >> 24 {
		case 0x01:
			pt2d.ApplyVerticalFilter32(decoded, int(tile.TileWidth), int(tile.TileHeight))
		case 0x02:
			pt2d.ApplyVerticalFilter8(decoded, int(tile.TileWidth), int(tile.TileHeight))
		}

		tilesDecoded = append(tilesDecoded, decoded)
	}

	wd, hg, pix := pt2d.Combine(tilesDecoded, header, pt2Tiles)

	outPath := strings.TrimSuffix(inFile, filepath.Ext(inFile)) + ".png"
	pt2d.SavePNG(pix, wd, hg, outPath)
}

func (pt2d *Pt2Decoder) ApplyVerticalFilter32(buf []byte, width int, height int) {
	if height < 2 {
		return
	}
	rowBytes := width * 4
	for y := 1; y < height; y++ {
		cur := y * rowBytes
		prev := (y - 1) * rowBytes
		for x := 0; x < rowBytes; x += 4 {
			a := binary.LittleEndian.Uint32(buf[cur+x : cur+x+4])
			b := binary.LittleEndian.Uint32(buf[prev+x : prev+x+4])
			binary.LittleEndian.PutUint32(buf[cur+x:cur+x+4], a+b)
		}
	}
}

func (pt2d *Pt2Decoder) ApplyVerticalFilter8(buf []byte, rowBytes, height int) {
	//only return, not implement
	_, _, _ = buf, rowBytes, height
}

func (pt2d *Pt2Decoder) Combine(tilesdata [][]byte, header Pt2Header, pt2tiles []Pt2Tile) (int, int, []byte) {
	width := int(header.CanvasW)
	height := int(header.CanvasH)
	canvas := make([]byte, width*height*4)

	x := 0
	y := 0
	rowH := 0
	for i, tile := range pt2tiles {
		data := tilesdata[i]
		tw := int(tile.TileWidth)
		th := int(tile.TileHeight)
		copyW := min(tw, width-x)
		copyH := min(th, height-y)

		for ty := 0; ty < copyH; ty++ {
			srcOff := ty * tw * 4
			dstOff := ((y+ty)*width + x) * 4
			copy(canvas[dstOff:dstOff+copyW*4], data[srcOff:srcOff+copyW*4])
		}

		x += tw
		if th > rowH {
			rowH = th
		}
		if x >= width {
			x = 0
			y += rowH
			rowH = 0
		}
	}
	return width, height, canvas
}

func (pt2d *Pt2Decoder) SavePNG(pixels []byte, width int, height int, outPath string) {
	rgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	copy(rgba.Pix, pixels)

	file, _ := os.Create(outPath)
	defer file.Close()
	png.Encode(file, rgba)
}

func main() {
	decoder := NewPt2Decoder()
	for _, inFile := range os.Args[1:] {
		decoder.DecodePt2(inFile)
		fmt.Printf("Convert %s \n", filepath.Base(inFile))
	}
}
