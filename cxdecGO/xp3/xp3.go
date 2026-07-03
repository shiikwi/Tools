package xp3

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"os"
	"unicode/utf16"
)

var (
	magic     = []byte("XP3\r\n \n\x1a\x8bg\x01")
	ErrFilter = errors.New("filter")
)

type Segment struct {
	Flags        uint32
	Offset       uint64
	OriginalSize uint64
	ArchivedSize uint64
}

type Entry struct {
	Name         string
	Flags        uint32
	OriginalSize uint64
	ArchivedSize uint64
	Adler32      *uint32
	Segments     []Segment
}

type ArchiveInfo struct {
	Path                string
	Size                uint64
	HeaderIndexOffset   uint64
	ResolvedIndexOffset uint64
	IndexCompressedSize uint64
	IndexOriginalSize   uint64
	EntryCount          int
}

type Filter interface {
	Apply(data []byte, offset uint64) bool
}

func ReadArchive(path string) (ArchiveInfo, []Entry, []byte, []byte, error) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return ArchiveInfo{}, nil, nil, nil, err
	}
	indexOffset, compressedSize, originalSize, index, err := LoadIndex(blob)
	if err != nil {
		return ArchiveInfo{}, nil, nil, nil, err
	}
	entries, err := ParseEntries(index)
	if err != nil {
		return ArchiveInfo{}, nil, nil, nil, err
	}
	return ArchiveInfo{
		Path:                path,
		Size:                uint64(len(blob)),
		HeaderIndexOffset:   binary.LittleEndian.Uint64(blob[11:19]),
		ResolvedIndexOffset: indexOffset,
		IndexCompressedSize: compressedSize,
		IndexOriginalSize:   originalSize,
		EntryCount:          len(entries),
	}, entries, blob, index, nil
}

func ResolveIndexOffset(blob []byte) (uint64, error) {
	if !bytes.HasPrefix(blob, magic) {
		return 0, fmt.Errorf("invalid xp3")
	}
	offset := binary.LittleEndian.Uint64(blob[11:19])
	if offset >= uint64(len(blob)) {
		return 0, fmt.Errorf("invalid index offset")
	}
	if offset == 0x17 && len(blob) >= 0x28 {
		candidate := binary.LittleEndian.Uint64(blob[0x20:0x28])
		if candidate > 0 && candidate < uint64(len(blob)) {
			return candidate, nil
		}
	}
	return offset, nil
}

func LoadIndex(blob []byte) (uint64, uint64, uint64, []byte, error) {
	indexOffset, err := ResolveIndexOffset(blob)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	flag := blob[indexOffset]
	cursor := indexOffset + 1
	switch flag {
	case 0:
		size := binary.LittleEndian.Uint64(blob[cursor : cursor+8])
		cursor += 8
		end := cursor + size
		return indexOffset, size, size, append([]byte(nil), blob[cursor:end]...), nil
	case 1:
		compressedSize := binary.LittleEndian.Uint64(blob[cursor : cursor+8])
		originalSize := binary.LittleEndian.Uint64(blob[cursor+8 : cursor+16])
		cursor += 16
		end := cursor + compressedSize
		index, err := inflate(blob[cursor:end])
		if err != nil || uint64(len(index)) != originalSize {
			return 0, 0, 0, nil, fmt.Errorf("index decompress fail")
		}
		return indexOffset, compressedSize, originalSize, index, nil
	default:
		return 0, 0, 0, nil, fmt.Errorf("unkonwn index flag")
	}
}

func ParseEntries(index []byte) ([]Entry, error) {
	entries := make([]Entry, 0)
	for cursor := 0; cursor+12 <= len(index); {
		tag := string(index[cursor : cursor+4])
		size := binary.LittleEndian.Uint64(index[cursor+4 : cursor+12])
		cursor += 12
		end := cursor + int(size)
		if tag != "File" {
			cursor = end
			continue
		}
		entry := Entry{}
		chunk := index[cursor:end]
		cursor = end
		for sub := 0; sub+12 <= len(chunk); {
			subTag := string(chunk[sub : sub+4])
			subSize := binary.LittleEndian.Uint64(chunk[sub+4 : sub+12])
			sub += 12
			subEnd := sub + int(subSize)
			body := chunk[sub:subEnd]
			sub = subEnd

			switch subTag {
			case "info":
				entry.Flags = binary.LittleEndian.Uint32(body[0:4])
				entry.OriginalSize = binary.LittleEndian.Uint64(body[4:12])
				entry.ArchivedSize = binary.LittleEndian.Uint64(body[12:20])
				nameSize := int(binary.LittleEndian.Uint16(body[20:22])) * 2
				entry.Name = decodeUTF16LE(body[22 : 22+nameSize])
			case "segm":
				for off := 0; off < len(body); off += 28 {
					entry.Segments = append(entry.Segments, Segment{
						Flags:        binary.LittleEndian.Uint32(body[off : off+4]),
						Offset:       binary.LittleEndian.Uint64(body[off+4 : off+12]),
						OriginalSize: binary.LittleEndian.Uint64(body[off+12 : off+20]),
						ArchivedSize: binary.LittleEndian.Uint64(body[off+20 : off+28]),
					})
				}
			case "adlr":
				if len(body) >= 4 {
					value := binary.LittleEndian.Uint32(body[:4])
					entry.Adler32 = &value
				}
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func ExtractEntry(blob []byte, entry Entry, filter Filter) ([]byte, uint32, bool, error) {
	if filter != nil && entry.Adler32 != nil {
		hash := adler32.New()
		if err := walkChunks(blob, entry, func(_ uint64, data []byte) error {
			_, err := hash.Write(data)
			return err
		}); err != nil {
			return nil, 0, false, err
		}
		if hash.Sum32() == *entry.Adler32 {
			filter = nil
		}
	}

	out := make([]byte, 0, entry.OriginalSize)
	filterApplied := false
	if err := walkChunks(blob, entry, func(offset uint64, data []byte) error {
		buf := append([]byte(nil), data...)
		if filter != nil {
			filterApplied = filter.Apply(buf, offset) || filterApplied
		}
		out = append(out, buf...)
		return nil
	}); err != nil {
		return nil, 0, false, err
	}
	if uint64(len(out)) != entry.OriginalSize {
		return nil, 0, false, fmt.Errorf("bad entry")
	}

	actual := adler32.Checksum(out)
	if entry.Adler32 != nil &&
		entry.Flags&0x80000000 != 0 &&
		actual != *entry.Adler32 &&
		!filterApplied {
		return out, actual, false, ErrFilter
	}
	return out, actual, filterApplied, nil
}

func IsWarningEntry(index int, entry Entry) bool {
	return index == 0 &&
		entry.Flags == 0 &&
		len(entry.Segments) == 1 &&
		entry.Segments[0].OriginalSize < entry.OriginalSize
}

func walkChunks(blob []byte, entry Entry, fn func(uint64, []byte) error) error {
	var offset uint64
	for _, seg := range entry.Segments {
		end := seg.Offset + seg.ArchivedSize
		if end > uint64(len(blob)) || end < seg.Offset {
			return fmt.Errorf("bad segment")
		}
		data := append([]byte(nil), blob[seg.Offset:end]...)
		if seg.Flags&1 != 0 {
			var err error
			data, err = inflate(data)
			if err != nil || uint64(len(data)) != seg.OriginalSize {
				return fmt.Errorf("bad segment")
			}
		}
		if err := fn(offset, data); err != nil {
			return err
		}
		offset += uint64(len(data))
	}
	return nil
}

func inflate(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func decodeUTF16LE(data []byte) string {
	words := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		words = append(words, binary.LittleEndian.Uint16(data[i:i+2]))
	}
	return string(utf16.Decode(words))
}
