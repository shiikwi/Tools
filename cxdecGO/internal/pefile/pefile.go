package pefile

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf16"
)

type Section struct {
	Name    string
	VA      uint32
	Size    uint32
	Raw     uint32
	RawSize uint32
}

type ResourceName struct {
	ID     uint32
	Text   string
	IsText bool
}

func ID(id uint32) ResourceName {
	return ResourceName{ID: id}
}

func Text(text string) ResourceName {
	return ResourceName{Text: text, IsText: true}
}

func (n ResourceName) String() string {
	if n.IsText {
		return n.Text
	}
	return fmt.Sprintf("%d", n.ID)
}

func (n ResourceName) Equal(other ResourceName) bool {
	if n.IsText || other.IsText {
		return n.IsText && other.IsText && n.Text == other.Text
	}
	return n.ID == other.ID
}

type ResourceKey struct {
	Type ResourceName
	Name ResourceName
	Lang ResourceName
}

func (k ResourceKey) String() string {
	return k.Type.String() + "/" + k.Name.String() + "/" + k.Lang.String()
}

type Image struct {
	Path        string
	Data        []byte
	ImageBase   uint64
	ResourceRVA uint32
	Sections    []Section
}

func Open(path string) (*Image, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return New(path, data)
}

func New(path string, data []byte) (*Image, error) {
	pe := u32(data, 0x3c)
	if pe+24 > uint32(len(data)) || string(data[pe:pe+4]) != "PE\x00\x00" {
		return nil, fmt.Errorf("%s is not a PE file", path)
	}

	coff := pe + 4
	sectionCount := u16(data, coff+2)
	optionalSize := u16(data, coff+16)
	optional := coff + 20
	if optional+uint32(optionalSize) > uint32(len(data)) {
		return nil, fmt.Errorf("%s has a truncated optional header", path)
	}

	magic := u16(data, optional)
	var imageBase uint64
	var dataDir uint32
	switch magic {
	case 0x10b:
		imageBase = uint64(u32(data, optional+28))
		dataDir = optional + 96
	case 0x20b:
		imageBase = binary.LittleEndian.Uint64(data[optional+24:])
		dataDir = optional + 112
	default:
		return nil, fmt.Errorf("%s has unsupported PE optional header magic 0x%x", path, magic)
	}
	resourceRVA := u32(data, dataDir+16)

	sections := make([]Section, 0, sectionCount)
	sectionTable := optional + uint32(optionalSize)
	for i := uint16(0); i < sectionCount; i++ {
		off := sectionTable + uint32(i)*40
		nameBytes := data[off : off+8]
		name := string(nameBytes)
		if nul := strings.IndexByte(name, 0); nul >= 0 {
			name = name[:nul]
		}
		virtualSize := u32(data, off+8)
		rawSize := u32(data, off+16)
		size := virtualSize
		if rawSize > size {
			size = rawSize
		}
		sections = append(sections, Section{
			Name:    name,
			VA:      u32(data, off+12),
			Size:    size,
			RawSize: rawSize,
			Raw:     u32(data, off+20),
		})
	}

	return &Image{
		Path:        path,
		Data:        data,
		ImageBase:   imageBase,
		ResourceRVA: resourceRVA,
		Sections:    sections,
	}, nil
}

func (img *Image) RVAToOffset(rva uint32) (uint32, error) {
	for _, section := range img.Sections {
		if rva >= section.VA && rva < section.VA+section.Size {
			return section.Raw + (rva - section.VA), nil
		}
	}
	return 0, fmt.Errorf("RVA 0x%x is not mapped", rva)
}

func (img *Image) OffsetToRVA(offset uint32) (uint32, error) {
	for _, section := range img.Sections {
		if offset >= section.Raw && offset < section.Raw+section.RawSize {
			return section.VA + (offset - section.Raw), nil
		}
	}
	return 0, fmt.Errorf("file offset 0x%x is not mapped", offset)
}

func (img *Image) ReadRVA(rva uint32, size uint32) ([]byte, error) {
	off, err := img.RVAToOffset(rva)
	if err != nil {
		return nil, err
	}
	if uint64(off)+uint64(size) > uint64(len(img.Data)) {
		return nil, fmt.Errorf("RVA 0x%x size 0x%x exceeds %s", rva, size, img.Path)
	}
	return img.Data[off : off+size], nil
}

func (img *Image) Resources() (map[ResourceKey][]byte, error) {
	found := map[ResourceKey][]byte{}
	resourceOff, _ := img.RVAToOffset(img.ResourceRVA)

	readName := func(nameOff uint32) (ResourceName, error) {
		base := resourceOff + nameOff
		length := u16(img.Data, base)
		rawStart := base + 2
		words := make([]uint16, length)
		for i := range words {
			words[i] = u16(img.Data, rawStart+uint32(i)*2)
		}
		return Text(string(utf16.Decode(words))), nil
	}

	var walk func(directoryRel uint32, path []ResourceName) error
	walk = func(directoryRel uint32, path []ResourceName) error {
		base := resourceOff + directoryRel
		named := u16(img.Data, base+12)
		ids := u16(img.Data, base+14)
		count := uint32(named + ids)
		for i := uint32(0); i < count; i++ {
			entry := base + 16 + i*8
			nameRaw := u32(img.Data, entry)
			child := u32(img.Data, entry+4)
			var name ResourceName
			if nameRaw&0x80000000 != 0 {
				var err error
				name, err = readName(nameRaw & 0x7fffffff)
				if err != nil {
					return err
				}
			} else {
				name = ID(nameRaw)
			}

			nextPath := append(append([]ResourceName{}, path...), name)
			if child&0x80000000 != 0 {
				if err := walk(child&0x7fffffff, nextPath); err != nil {
					return err
				}
				continue
			}

			if len(nextPath) != 3 {
				continue
			}
			dataEntry := resourceOff + child
			rva := u32(img.Data, dataEntry)
			size := u32(img.Data, dataEntry+4)
			data, err := img.ReadRVA(rva, size)
			if err != nil {
				return err
			}
			copied := append([]byte(nil), data...)
			found[ResourceKey{Type: nextPath[0], Name: nextPath[1], Lang: nextPath[2]}] = copied
		}
		return nil
	}

	return found, walk(0, nil)
}

func FindResource(resources map[ResourceKey][]byte, typ ResourceName, name ResourceName) ([]byte, error) {
	for key, data := range resources {
		if key.Type.Equal(typ) && key.Name.Equal(name) {
			return data, nil
		}
	}
	keys := make([]string, 0, len(resources))
	for key := range resources {
		keys = append(keys, key.String())
	}
	sort.Strings(keys)
	if len(keys) > 12 {
		keys = keys[:12]
	}
	sample := strings.Join(keys, ", ")
	if sample == "" {
		sample = "PE has no parsed resources"
	}
	return nil, fmt.Errorf("resource %s/%s was not found; available resources include: %s", typ, name, sample)
}

func U16(data []byte, off uint32) uint16 {
	return u16(data, off)
}

func U32(data []byte, off uint32) uint32 {
	return u32(data, off)
}

func u16(data []byte, off uint32) uint16 {
	return binary.LittleEndian.Uint16(data[off:])
}

func u32(data []byte, off uint32) uint32 {
	return binary.LittleEndian.Uint32(data[off:])
}
