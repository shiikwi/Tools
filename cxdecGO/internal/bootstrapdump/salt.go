package bootstrapdump

import (
	"bytes"

	"cxdecgo/internal/pefile"
)

const saltSize = 0x2000

type saltCandidate struct {
	kind         string
	offset       uint32
	rva          uint32
	va           uint64
	codeRVA      uint32
	anchorOffset uint32
	anchorRVA    uint32
	anchorName   string
	salt         []byte
}

func autoSaltCandidates(img *pefile.Image) []saltCandidate {
	seen := map[uint32]bool{}
	var out []saltCandidate
	for _, c := range append(saltAssignmentCandidates(img), packedNeighborhoodCandidates(img)...) {
		if seen[c.offset] {
			continue
		}
		seen[c.offset] = true
		out = append(out, c)
	}
	return out
}

func makeSaltCandidate(img *pefile.Image, seen map[uint32]bool, c saltCandidate) (saltCandidate, bool) {
	if seen[c.offset] || c.offset+saltSize > uint32(len(img.Data)) {
		return saltCandidate{}, false
	}
	rva, err := img.OffsetToRVA(c.offset)
	if err != nil {
		return saltCandidate{}, false
	}
	c.rva = rva
	c.va = img.ImageBase + uint64(rva)
	c.salt = append([]byte(nil), img.Data[c.offset:c.offset+saltSize]...)
	if c.anchorOffset != 0 {
		if anchorRVA, err := img.OffsetToRVA(c.anchorOffset); err == nil {
			c.anchorRVA = anchorRVA
		}
	}
	seen[c.offset] = true
	return c, true
}

func saltAssignmentCandidates(img *pefile.Image) []saltCandidate {
	data := img.Data
	seen := map[uint32]bool{}
	var out []saltCandidate
	if len(data) < 20 {
		return out
	}
	for off := uint32(0); off < uint32(len(data))-20; off++ {
		if data[off] != 0xc7 || data[off+1] != 0x05 {
			continue
		}
		saltVA := uint64(pefile.U32(data, off+6))
		if saltVA < img.ImageBase {
			continue
		}
		saltRVA64 := saltVA - img.ImageBase
		if saltRVA64 > 0xffffffff {
			continue
		}
		saltOffset, err := img.RVAToOffset(uint32(saltRVA64))
		if err != nil || saltOffset+saltSize > uint32(len(data)) {
			continue
		}
		windowEnd := min(uint32(len(data))-10, off+64)
		for sizeOff := off + 10; sizeOff < windowEnd; sizeOff++ {
			if data[sizeOff] != 0xc7 || data[sizeOff+1] != 0x05 {
				continue
			}
			if pefile.U32(data, sizeOff+6) != saltSize {
				continue
			}
			codeRVA, err := img.OffsetToRVA(off)
			if err != nil {
				codeRVA = off
			}
			c, ok := makeSaltCandidate(img, seen, saltCandidate{
				kind:    "code_assignment",
				offset:  saltOffset,
				codeRVA: codeRVA,
			})
			if ok {
				out = append(out, c)
			}
			break
		}
	}
	return out
}

func packedNeighborhoodCandidates(img *pefile.Image) []saltCandidate {
	data := img.Data
	seen := map[uint32]bool{}
	var out []saltCandidate
	add := func(offset uint32, anchor uint32, name string) {
		c, ok := makeSaltCandidate(img, seen, saltCandidate{
			kind:         "packed_neighborhood",
			offset:       offset,
			anchorOffset: anchor,
			anchorName:   name,
		})
		if ok {
			out = append(out, c)
		}
	}

	for cursor := 0; ; {
		anchor := bytes.Index(data[cursor:], []byte("V2Link\x00\x00"))
		if anchor < 0 {
			break
		}
		anchor += cursor
		if anchor >= saltSize {
			add(uint32(anchor-saltSize), uint32(anchor), "V2Link-before")
		}
		cursor = anchor + 1
	}

	for cursor := 0; ; {
		anchor := bytes.Index(data[cursor:], []byte("forcedataxp3\x00"))
		if anchor < 0 {
			break
		}
		anchor += cursor
		start := (uint32(anchor+len("forcedataxp3\x00")) + 0xf) &^ 0xf
		end := min(uint32(anchor)+0x100, uint32(len(data))-saltSize)
		for offset := start; offset <= end; offset += 0x10 {
			add(offset, uint32(anchor), "forcedataxp3-near")
		}
		cursor = anchor + 1
	}
	return out
}
