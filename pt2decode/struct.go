package main

type Pt2Header struct {
	Version    uint16
	HeaderSize uint16
	CanvasW    uint16
	CanvasH    uint16
	Aux        uint32
	Meta       [8]uint16 //0x20 headerSize has
}

type Pt2Tile struct {
	Flag       uint32
	DataOff    uint32
	CompSize   uint32
	TileWidth  uint16
	TileHeight uint16
}
