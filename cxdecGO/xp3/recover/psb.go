package recover

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"strings"
)

type psbHeader struct {
	version              uint16
	headerLength         uint32
	offsetNames          uint32
	offsetStrings        uint32
	offsetStringsData    uint32
	offsetChunkOffsets   uint32
	offsetChunkLengths   uint32
	offsetChunkData      uint32
	offsetEntries        uint32
	checksum             uint32
	offsetExtraOffsets   uint32
	offsetExtraLengths   uint32
	offsetExtraChunkData uint32
}

type psbKind byte

const (
	psbNull psbKind = iota
	psbString
	psbList
	psbObject
)

type psbValue struct {
	kind   psbKind
	text   string
	list   []psbValue
	object map[string]psbValue
}

type psbParser struct {
	data    []byte
	header  psbHeader
	names   []string
	strings []string
}

type psbReader struct {
	data []byte
	pos  int
}

func psbMineStrings(data []byte) ([]string, bool) {
	if !bytes.HasPrefix(data, []byte("PSB")) {
		return nil, false
	}
	root, pool, err := parsePSB(data)
	if err != nil {
		return nil, false
	}
	out := make([]string, 0, len(pool)+32)
	out = append(out, pool...)
	for _, v := range psbResourceHints(root) {
		out = appendUnique(out, v)
	}
	for _, v := range psbStringPoolHints(pool) {
		out = appendUnique(out, v)
	}
	return out, true
}

func parsePSB(data []byte) (psbValue, []string, error) {
	reader := psbReader{data: data}
	header, err := readPSBHeader(&reader)
	if err != nil {
		return psbValue{}, nil, err
	}
	if headerEncrypted(header, len(data)) {
		return psbValue{}, nil, errors.New("encrypted psb")
	}

	stringOffsets, err := readPSBArrayAt(data, header.offsetStrings)
	if err != nil {
		return psbValue{}, nil, err
	}
	strings, err := loadPSBStrings(data, header.offsetStringsData, stringOffsets)
	if err != nil {
		return psbValue{}, nil, err
	}

	var names []string
	if header.version == 1 {
		indexes, err := readPSBArrayAt(data, header.headerLength)
		if err != nil {
			return psbValue{}, nil, err
		}
		names, err = loadPSBStrings(data, header.offsetNames, indexes)
	} else {
		names, err = loadPSBTrieNames(data, header.offsetNames)
	}
	if err != nil {
		return psbValue{}, nil, err
	}

	parser := psbParser{
		data:    data,
		header:  header,
		names:   names,
		strings: strings,
	}
	entry, err := newPSBReaderAt(data, int(header.offsetEntries))
	if err != nil {
		return psbValue{}, nil, err
	}
	root, err := parser.unpack(&entry)
	return root, strings, err
}

func readPSBHeader(reader *psbReader) (psbHeader, error) {
	signature, err := reader.readBytes(4)
	if err != nil {
		return psbHeader{}, err
	}
	if !bytes.HasPrefix(signature, []byte("PSB")) {
		return psbHeader{}, errors.New("not psb")
	}
	version, err := reader.readU16()
	if err != nil {
		return psbHeader{}, err
	}
	if _, err := reader.readU16(); err != nil {
		return psbHeader{}, err
	}
	headerLength, err := reader.readU32()
	if err != nil {
		return psbHeader{}, err
	}
	offsetNames, err := reader.readU32()
	if err != nil {
		return psbHeader{}, err
	}

	header := psbHeader{
		version:      version,
		headerLength: headerLength,
		offsetNames:  offsetNames,
	}
	if offsetNames >= uint32(len(reader.data)) {
		return header, nil
	}

	fields := []*uint32{
		&header.offsetStrings,
		&header.offsetStringsData,
		&header.offsetChunkOffsets,
		&header.offsetChunkLengths,
		&header.offsetChunkData,
		&header.offsetEntries,
	}
	if version > 2 {
		fields = append(fields, &header.checksum)
	}
	if version > 3 {
		fields = append(fields, &header.offsetExtraOffsets, &header.offsetExtraLengths, &header.offsetExtraChunkData)
	}
	for _, field := range fields {
		value, err := reader.readU32()
		if err != nil {
			return psbHeader{}, err
		}
		*field = value
	}
	return header, nil
}

func headerEncrypted(header psbHeader, fileLen int) bool {
	return header.headerLength > 72 ||
		header.offsetNames == 0 ||
		header.version > 1 && header.headerLength != header.offsetNames && header.headerLength != 0 ||
		int(header.offsetNames) >= fileLen
}

func loadPSBStrings(data []byte, base uint32, offsets []uint32) ([]string, error) {
	out := make([]string, 0, len(offsets))
	for _, offset := range offsets {
		start, ok := checkedPSBAdd(int(base), int(offset))
		if !ok {
			return nil, errors.New("psb offset overflow")
		}
		value, err := readPSBZeroString(data, start)
		if err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

func loadPSBTrieNames(data []byte, offset uint32) ([]string, error) {
	reader, err := newPSBReaderAt(data, int(offset))
	if err != nil {
		return nil, err
	}
	charset, err := reader.readTypedArray()
	if err != nil {
		return nil, err
	}
	namesData, err := reader.readTypedArray()
	if err != nil {
		return nil, err
	}
	nameIndexes, err := reader.readTypedArray()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(nameIndexes))
	for _, index := range nameIndexes {
		current := int(index)
		bytes := make([]byte, 0, 32)
		for {
			if current < 0 || current >= len(namesData) {
				return nil, errors.New("psb name index")
			}
			chr := namesData[current]
			if chr == 0 {
				break
			}
			codeIndex := int(chr)
			if codeIndex < 0 || codeIndex >= len(namesData) {
				return nil, errors.New("psb name code")
			}
			code := namesData[codeIndex]
			if int(code) >= len(charset) {
				return nil, errors.New("psb name charset")
			}
			real := chr - charset[code]
			bytes = append(bytes, byte(real))
			current = int(code)
		}
		for left, right := 0, len(bytes)-1; left < right; left, right = left+1, right-1 {
			bytes[left], bytes[right] = bytes[right], bytes[left]
		}
		names = append(names, string(bytes))
	}
	return names, nil
}

func (p *psbParser) unpack(reader *psbReader) (psbValue, error) {
	typeByte, err := reader.readU8()
	if err != nil {
		return psbValue{}, err
	}
	switch {
	case typeByte <= 0x04:
		return psbValue{kind: psbNull}, nil
	case typeByte >= 0x05 && typeByte <= 0x0c:
		_, err := reader.readBytes(int(typeByte - 0x04))
		return psbValue{kind: psbNull}, err
	case typeByte >= 0x0d && typeByte <= 0x14:
		_, err := reader.readArray(typeByte - 0x0d + 1)
		return psbValue{kind: psbNull}, err
	case typeByte >= 0x15 && typeByte <= 0x18:
		index, err := reader.readCompactU32(typeByte - 0x15 + 1)
		if err != nil {
			return psbValue{}, err
		}
		if int(index) >= len(p.strings) {
			return psbValue{}, errors.New("psb string index")
		}
		return psbValue{kind: psbString, text: p.strings[index]}, nil
	case typeByte >= 0x19 && typeByte <= 0x1c:
		_, err := reader.readCompactU32(typeByte - 0x19 + 1)
		return psbValue{kind: psbNull}, err
	case typeByte == 0x1d:
		return psbValue{kind: psbNull}, nil
	case typeByte == 0x1e:
		_, err := reader.readU32()
		return psbValue{kind: psbNull}, err
	case typeByte == 0x1f:
		_, err := reader.readU64()
		return psbValue{kind: psbNull}, err
	case typeByte == 0x20:
		return p.loadList(reader)
	case typeByte == 0x21 && p.header.version == 1:
		return p.loadObjectV1(reader)
	case typeByte == 0x21:
		return p.loadObject(reader)
	case typeByte >= 0x22 && typeByte <= 0x25:
		_, err := reader.readCompactU32(typeByte - 0x22 + 1)
		return psbValue{kind: psbNull}, err
	default:
		return psbValue{}, errors.New("psb type")
	}
}

func (p *psbParser) loadObject(reader *psbReader) (psbValue, error) {
	nameType, err := reader.readU8()
	if err != nil {
		return psbValue{}, err
	}
	names, err := reader.readArray(psbArrayCountSize(nameType))
	if err != nil {
		return psbValue{}, err
	}
	offsetType, err := reader.readU8()
	if err != nil {
		return psbValue{}, err
	}
	offsets, err := reader.readArray(psbArrayCountSize(offsetType))
	if err != nil {
		return psbValue{}, err
	}

	base := reader.pos
	object := make(map[string]psbValue, len(names))
	for i, nameIndex := range names {
		if i >= len(offsets) || int(nameIndex) >= len(p.names) {
			continue
		}
		childPos, ok := checkedPSBAdd(base, int(offsets[i]))
		if !ok {
			continue
		}
		child, err := newPSBReaderAt(p.data, childPos)
		if err != nil {
			continue
		}
		value, err := p.unpack(&child)
		if err != nil {
			continue
		}
		object[p.names[nameIndex]] = value
	}
	return psbValue{kind: psbObject, object: object}, nil
}

func (p *psbParser) loadObjectV1(reader *psbReader) (psbValue, error) {
	offsetType, err := reader.readU8()
	if err != nil {
		return psbValue{}, err
	}
	offsets, err := reader.readArray(psbArrayCountSize(offsetType))
	if err != nil {
		return psbValue{}, err
	}

	base := reader.pos
	object := make(map[string]psbValue, len(offsets))
	for _, offset := range offsets {
		childPos, ok := checkedPSBAdd(base, int(offset))
		if !ok {
			continue
		}
		child, err := newPSBReaderAt(p.data, childPos)
		if err != nil {
			continue
		}
		typeByte, err := child.readU8()
		if err != nil || typeByte < 0x05 || typeByte > 0x08 {
			continue
		}
		nameIndex, err := child.readCompactU32(typeByte - 0x04)
		if err != nil || int(nameIndex) >= len(p.names) {
			continue
		}
		value, err := p.unpack(&child)
		if err != nil {
			continue
		}
		object[p.names[nameIndex]] = value
	}
	return psbValue{kind: psbObject, object: object}, nil
}

func (p *psbParser) loadList(reader *psbReader) (psbValue, error) {
	offsetType, err := reader.readU8()
	if err != nil {
		return psbValue{}, err
	}
	offsets, err := reader.readArray(psbArrayCountSize(offsetType))
	if err != nil {
		return psbValue{}, err
	}

	base := reader.pos
	values := make([]psbValue, 0, len(offsets))
	for _, offset := range offsets {
		childPos, ok := checkedPSBAdd(base, int(offset))
		if !ok {
			continue
		}
		child, err := newPSBReaderAt(p.data, childPos)
		if err != nil {
			continue
		}
		value, err := p.unpack(&child)
		if err == nil {
			values = append(values, value)
		}
	}
	return psbValue{kind: psbList, list: values}, nil
}

func newPSBReaderAt(data []byte, pos int) (psbReader, error) {
	if pos < 0 || pos > len(data) {
		return psbReader{}, errors.New("psb offset")
	}
	return psbReader{data: data, pos: pos}, nil
}

func (r *psbReader) readU8() (byte, error) {
	data, err := r.readBytes(1)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}

func (r *psbReader) readU16() (uint16, error) {
	data, err := r.readBytes(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(data), nil
}

func (r *psbReader) readU32() (uint32, error) {
	data, err := r.readBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

func (r *psbReader) readU64() (uint64, error) {
	data, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(data), nil
}

func (r *psbReader) readBytes(size int) ([]byte, error) {
	if size < 0 || r.pos > len(r.data)-size {
		return nil, errors.New("psb eof")
	}
	out := r.data[r.pos : r.pos+size]
	r.pos += size
	return out, nil
}

func (r *psbReader) readCompactU32(size byte) (uint32, error) {
	if size > 4 {
		return 0, errors.New("psb int")
	}
	data, err := r.readBytes(int(size))
	if err != nil {
		return 0, err
	}
	var out uint32
	for i, value := range data {
		out |= uint32(value) << (8 * i)
	}
	return out, nil
}

func (r *psbReader) readTypedArray() ([]uint32, error) {
	typeByte, err := r.readU8()
	if err != nil {
		return nil, err
	}
	return r.readArray(psbArrayCountSize(typeByte))
}

func (r *psbReader) readArray(countSize byte) ([]uint32, error) {
	if countSize > 8 {
		return nil, errors.New("psb array")
	}
	countBytes, err := r.readBytes(int(countSize))
	if err != nil {
		return nil, err
	}
	count := compactUint64(countBytes)
	if count > uint64(math.MaxInt32) {
		return nil, errors.New("psb array count")
	}
	entryType, err := r.readU8()
	if err != nil {
		return nil, err
	}
	if entryType < 0x0c {
		return nil, errors.New("psb array type")
	}
	entrySize := entryType - 0x0c
	if entrySize > 4 {
		return nil, errors.New("psb array item")
	}
	if entrySize == 0 {
		if count == 0 {
			return nil, nil
		}
		return nil, errors.New("psb zero array")
	}
	byteLen := int(count) * int(entrySize)
	data, err := r.readBytes(byteLen)
	if err != nil {
		return nil, err
	}
	out := make([]uint32, 0, count)
	for i := 0; i < len(data); i += int(entrySize) {
		out = append(out, compactUint32(data[i:i+int(entrySize)]))
	}
	return out, nil
}

func readPSBArrayAt(data []byte, offset uint32) ([]uint32, error) {
	reader, err := newPSBReaderAt(data, int(offset))
	if err != nil {
		return nil, err
	}
	return reader.readTypedArray()
}

func psbArrayCountSize(typeByte byte) byte {
	if typeByte < 0x0d || typeByte > 0x14 {
		return 9
	}
	return typeByte - 0x0d + 1
}

func compactUint32(data []byte) uint32 {
	var out uint32
	for i, value := range data {
		out |= uint32(value) << (8 * i)
	}
	return out
}

func compactUint64(data []byte) uint64 {
	var out uint64
	for i, value := range data {
		out |= uint64(value) << (8 * i)
	}
	return out
}

func checkedPSBAdd(left, right int) (int, bool) {
	if right > 0 && left > math.MaxInt-right {
		return 0, false
	}
	return left + right, true
}

func readPSBZeroString(data []byte, start int) (string, error) {
	if start < 0 || start >= len(data) {
		return "", errors.New("psb string")
	}
	end := bytes.IndexByte(data[start:], 0)
	if end < 0 {
		return "", errors.New("psb string end")
	}
	return string(data[start : start+end]), nil
}

func psbResourceHints(root psbValue) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	add := func(raw string, exts []string, dirs []string) {
		for _, name := range psbCandidates(raw, exts, dirs) {
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}

	var walk func(psbValue, []string, string)
	walk = func(value psbValue, path []string, command string) {
		switch value.kind {
		case psbString:
			if name, _, ok := cleanName(value.text); ok && hasKnownExt(name) {
				add(name, nil, nil)
			}
		case psbList:
			for _, child := range value.list {
				walk(child, path, command)
			}
		case psbObject:
			if name, ok := psbObjectString(value.object, "name"); ok {
				command = strings.ToLower(name)
			}
			psbObjectHints(value.object, path, command, add)
			for key, child := range value.object {
				walk(child, append(path, key), command)
			}
		}
	}
	walk(root, nil, "")
	return out
}

func psbObjectHints(object map[string]psbValue, path []string, command string, add func(string, []string, []string)) {
	if imageFile, ok := object["imageFile"]; ok && imageFile.kind == psbObject {
		if file, ok := psbObjectString(imageFile.object, "file"); ok {
			add(file, psbImageExts, psbImageDirs)
		}
	}
	if stand, ok := object["stand"]; ok && stand.kind == psbObject {
		if file, ok := psbObjectString(stand.object, "file"); ok {
			add(file, []string{".stand"}, []string{"fgimage", "fgimage2"})
		}
	}
	if file, ok := psbObjectString(object, "file"); ok {
		switch {
		case psbLastPath(path, "imageFile"):
			add(file, psbImageExts, psbImageDirs)
		case psbLastPath(path, "stand"):
			add(file, []string{".stand"}, []string{"fgimage", "fgimage2"})
		case psbImageCommand(command):
			add(file, psbImageExts, psbImageDirs)
		case psbAudioCommand(command):
			add(file, psbAudioExts, psbAudioDirs)
		}
	}
	if filename, ok := psbObjectString(object, "filename"); ok {
		switch {
		case psbAudioCommand(command):
			add(filename, psbAudioExts, psbAudioDirs)
		case psbImageCommand(command):
			add(filename, psbImageExts, psbImageDirs)
		case psbVideoCommand(command):
			add(filename, psbVideoExts, []string{"video", "video2"})
		default:
			add(filename, nil, nil)
		}
	}
	if rule, ok := psbObjectString(object, "rule"); ok {
		add(rule, []string{".tlg", ".png", ".bmp"}, []string{"rule"})
	}
}

var psbImageExts = []string{".tlg", ".png", ".pimg", ".psb", ".stage", ".event", ".jpg", ".jpeg", ".webp", ".bmp"}
var psbAudioExts = []string{".ogg", ".ogg.sli", ".wav", ".opus", ".tcw"}
var psbVideoExts = []string{".mp4", ".webm", ".avi", ".wmv", ".amv"}
var psbImageDirs = []string{"image", "bgimage", "fgimage", "fgimage2", "evimage", "face", "thum", "rule"}
var psbAudioDirs = []string{"bgm", "sound", "se", "sysse", "voice"}

func psbCandidates(raw string, exts []string, dirs []string) []string {
	name, isDir, ok := cleanName(raw)
	if !ok || isDir {
		return nil
	}
	if hasKnownExt(name) {
		return []string{name}
	}
	if len(exts) == 0 {
		return nil
	}
	out := make([]string, 0, len(exts)*(len(dirs)+1))
	hasDir := strings.Contains(name, "/")
	for _, ext := range exts {
		candidate := name + ext
		out = appendUnique(out, candidate)
		if hasDir {
			continue
		}
		for _, dir := range dirs {
			out = appendUnique(out, dir+"/"+candidate)
		}
	}
	return out
}

func psbStringPoolHints(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0)
	add := func(name string) {
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	for _, value := range values {
		for _, name := range psbStringCandidates(value) {
			add(name)
		}
	}
	return out
}

func psbStringCandidates(raw string) []string {
	name, isDir, ok := cleanName(raw)
	if !ok || isDir {
		return nil
	}
	if hasKnownExt(name) {
		return []string{name}
	}
	_, file := splitName(name)
	lower := strings.ToLower(file)
	switch {
	case strings.HasPrefix(lower, "map") && hasDigit(lower[3:]):
		return tableCandidates(name, []string{".png"})
	case audioStem(name) || voiceLineStem(lower):
		return tableCandidates(name, psbAudioExts)
	case strings.HasPrefix(lower, "cg_"):
		out := psbPoolImageCandidates(name)
		if trimmed := strings.TrimPrefix(name, "cg_"); trimmed != name {
			for _, candidate := range psbPoolImageCandidates(trimmed) {
				out = appendUnique(out, candidate)
			}
		}
		return out
	case likelyImageStem(name) || psbImageStem(lower):
		return psbPoolImageCandidates(name)
	default:
		return nil
	}
}

func psbPoolImageCandidates(name string) []string {
	return tableCandidates(name, []string{".png", ".pimg", ".tlg"})
}

func psbImageStem(file string) bool {
	return strings.HasPrefix(file, "sd") && hasDigit(file) ||
		strings.HasPrefix(file, "lay_") ||
		strings.HasPrefix(file, "grade_") ||
		strings.HasPrefix(file, "sakura") && hasDigit(file) ||
		strings.HasPrefix(file, "facemask_") ||
		strings.HasPrefix(file, "cine") && hasDigit(file)
}

func voiceLineStem(file string) bool {
	if len(file) < 9 || file[0] < 'a' || file[0] > 'z' || file[3] != '_' {
		return false
	}
	if file[1] < '0' || file[1] > '9' || file[2] < '0' || file[2] > '9' {
		return false
	}
	for i := 4; i < len(file); i++ {
		if file[i] < '0' || file[i] > '9' {
			return false
		}
	}
	return true
}

func psbObjectString(object map[string]psbValue, key string) (string, bool) {
	value, ok := object[key]
	if !ok || value.kind != psbString || strings.TrimSpace(value.text) == "" {
		return "", false
	}
	return value.text, true
}

func psbLastPath(path []string, want string) bool {
	return len(path) != 0 && path[len(path)-1] == want
}

func psbAudioCommand(command string) bool {
	return command == "bgm" || command == "se" || command == "sysse" ||
		command == "voice" || strings.Contains(command, "sound")
}

func psbImageCommand(command string) bool {
	return command == "image" || command == "cg" || command == "bg" ||
		command == "fg" || strings.Contains(command, "image")
}

func psbVideoCommand(command string) bool {
	return command == "movie" || command == "video" || strings.Contains(command, "movie")
}

func ParsePSBAllStrings(data []byte) ([]string, error) {
	root, pool, err := parsePSB(data)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(pool)+16)
	out = append(out, pool...)
	var walk func(v psbValue)
	walk = func(v psbValue) {
		if v.kind == psbString {
			out = append(out, v.text)
		}
		for _, item := range v.list {
			walk(item)
		}
		for _, item := range v.object {
			walk(item)
		}
	}
	walk(root)
	return out, nil
}
