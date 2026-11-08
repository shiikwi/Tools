package hxv4

import (
	"bytes"
	"compress/zlib"
	"cxdecgo/xp3"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"unicode/utf16"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	opAddImm = iota + 1
	opRecurse
	opAddScratch
	opMulScratch
	opScratchMinusResult
	opShlScratch
	opShrScratch
	opSubScratch
	opBitShuffle
	opSetImm
	opSetSeed
	opDec
	opInc
	opNeg
	opNot
	opTableImm
	opTableMasked
	opSubImm
	opStoreScratch
	opXorImm
	opStop
)

type Program struct {
	HolderWords [6]uint32
	ContextU32  []uint32
	Lanes       [][]laneRecord
	HXV4Key     []byte
	HXV4Nonce0  []byte
	HXV4Nonce1  []byte
}

type laneRecord struct {
	Param uint32
	Op    int
}

type Descriptor struct {
	Offset uint64
	Size   uint32
	Flags  uint16
}

type Record struct {
	DomainHash string
	FileHash   string
	Packed     uint32
	Key        uint64
}

type FilterRuntimeState struct {
	boundary0   filterBoundary
	boundary1   filterBoundary
	splitOffset uint64
	bulkKey     []byte
}

type filterBoundary struct {
	pos0  uint16
	pos1  uint16
	key   uint32
	byte0 byte
	byte1 byte
}

func LoadProgram(jsonPath string) (*Program, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}
	var payload struct {
		HolderWords []uint32 `json:"holder_words"`
		ContextU32  []uint32 `json:"context_u32"`
		Lanes       []struct {
			Records [][]int64 `json:"records"`
		} `json:"lanes"`
		HXV4Key    string `json:"hxv4_key"`
		HXV4Nonce0 string `json:"hxv4_nonce0"`
		HXV4Nonce1 string `json:"hxv4_nonce1"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	program := &Program{
		ContextU32: append([]uint32(nil), payload.ContextU32...),
		Lanes:      make([][]laneRecord, 128),
	}
	copy(program.HolderWords[:], payload.HolderWords[:6])

	program.HXV4Key, err = hex.DecodeString(payload.HXV4Key)
	if err != nil {
		return nil, err
	}
	program.HXV4Nonce0, err = hex.DecodeString(payload.HXV4Nonce0)
	if err != nil {
		return nil, err
	}
	program.HXV4Nonce1, err = hex.DecodeString(payload.HXV4Nonce1)
	if err != nil {
		return nil, err
	}

	for laneIndex, lane := range payload.Lanes {
		records := make([]laneRecord, 0, len(lane.Records))
		for _, pair := range lane.Records {
			records = append(records, laneRecord{
				Param: uint32(pair[0]),
				Op:    int(pair[1]),
			})
		}
		program.Lanes[laneIndex] = records
	}
	return program, nil
}

func ParseTable(blob []byte, index []byte, _ []xp3.Entry, program *Program) (Descriptor, []Record, error) {
	descriptor, ok, err := findDescriptor(index)
	if err != nil {
		return Descriptor{}, nil, err
	}
	if !ok {
		return Descriptor{}, nil, fmt.Errorf("no hxv4")
	}
	end := descriptor.Offset + uint64(descriptor.Size)
	decrypted, err := decryptPayload(blob[descriptor.Offset:end], descriptor.Flags, program)
	if err != nil {
		return Descriptor{}, nil, err
	}
	expectedSize := binary.LittleEndian.Uint32(decrypted[:4])
	tableBlob, err := inflate(decrypted[4:])
	if err != nil {
		return Descriptor{}, nil, fmt.Errorf("decompress hxv4 index fail")
	}
	if uint32(len(tableBlob)) != expectedSize {
		return Descriptor{}, nil, fmt.Errorf("decompress hxv4 index fail")
	}

	value, err := newTJSReader(tableBlob).readValue()
	if err != nil {
		return Descriptor{}, nil, err
	}
	root, ok := value.([]any)
	if !ok || len(root)%2 != 0 {
		return Descriptor{}, nil, fmt.Errorf("parse hxv4 table fail")
	}

	records := make([]Record, 0)
	for groupIndex := 0; groupIndex < len(root); groupIndex += 2 {
		domainHash, ok := root[groupIndex].([]byte)
		group, ok2 := root[groupIndex+1].([]any)
		if !ok || !ok2 || len(group)%2 != 0 {
			return Descriptor{}, nil, fmt.Errorf("domain hash error")
		}
		for itemIndex := 0; itemIndex < len(group); itemIndex += 2 {
			fileHash, ok := group[itemIndex].([]byte)
			pair, ok2 := group[itemIndex+1].([]any)
			if !ok || !ok2 || len(pair) != 2 {
				return Descriptor{}, nil, fmt.Errorf("file hash error")
			}
			packed, ok0 := toInt64(pair[0])
			keyValue, ok1 := toInt64(pair[1])
			if !ok0 || !ok1 {
				return Descriptor{}, nil, fmt.Errorf("file key error")
			}
			records = append(records, Record{
				DomainHash: hex.EncodeToString(domainHash),
				FileHash:   hex.EncodeToString(fileHash),
				Packed:     uint32(packed),
				Key:        uint64(keyValue),
			})
		}
	}

	return descriptor, records, nil
}

func (p *Program) BuildFilterState(key uint64, openFlag uint16) []byte {
	keyLo := uint32(key)
	keyHi := uint32(key >> 32)
	if openFlag&1 == 0 {
		keyLo ^= p.HolderWords[2]
		keyHi ^= p.HolderWords[3]
	}
	key64 := uint64(keyLo) | (uint64(keyHi) << 32)

	state := make([]byte, 48)
	binary.LittleEndian.PutUint64(state[0:8], p.Get64FromU32(keyLo))
	binary.LittleEndian.PutUint64(state[8:16], p.Get64FromU32(keyHi))
	bulkOffset := p.HolderWords[5] + (p.HolderWords[4] & uint32(key64>>16))
	binary.LittleEndian.PutUint32(state[16:20], bulkOffset)
	binary.LittleEndian.PutUint32(state[20:24], 0)

	cur := ^key64
	bitPos := -1
	out := 24
	for out < 40 {
		if bitPos < 0 {
			cur = ^p.Get64FromU32(uint32(cur))
			bitPos = 64
		} else {
			state[out] = byte(cur >> bitPos)
			out++
		}
		bitPos -= 8
	}
	state[44] = 1
	state[45] = 0
	return state
}

func (p *Program) Get64FromU32(value uint32) uint64 {
	laneIndex := value & 0x7f
	seed := value >> 7
	lo := p.evalLane(int(laneIndex), seed)
	hi := p.evalLane(int(laneIndex), ^seed)
	return uint64(lo) | (uint64(hi) << 32)
}

func (p *Program) evalLane(laneIndex int, seed uint32) uint32 {
	if laneIndex < 0 || laneIndex >= len(p.Lanes) {
		panic("bad lane")
	}
	result, _ := p.evalRecords(p.Lanes[laneIndex], 0, 0, seed, new(uint32))
	return result
}

func (p *Program) evalRecords(records []laneRecord, pc int, result uint32, seed uint32, scratch *uint32) (uint32, int) {
	for pc < len(records) {
		record := records[pc]
		pc++
		param := record.Param
		switch record.Op {
		case opStop:
			return result, pc
		case opRecurse:
			saved := *scratch
			nested := result
			result, pc = p.evalRecords(records, pc, nested, seed, scratch)
			*scratch = saved
		case opAddImm:
			result += param
		case opAddScratch:
			result += *scratch
		case opMulScratch:
			result *= *scratch
		case opScratchMinusResult:
			result = *scratch - result
		case opShlScratch:
			result <<= *scratch & 0xf
		case opShrScratch:
			result >>= *scratch & 0xf
		case opSubScratch:
			result -= *scratch
		case opBitShuffle:
			result = 2*(result&^param) | ((param >> 1) & (result >> 1))
		case opSetImm:
			result = param
		case opSetSeed:
			result = seed
		case opDec:
			result--
		case opInc:
			result++
		case opNeg:
			result = 0 - result
		case opNot:
			result = ^result
		case opTableImm:
			result = p.contextValue(int(param))
		case opTableMasked:
			result = p.contextValue(int(param & result))
		case opSubImm:
			result -= param
		case opStoreScratch:
			*scratch = result
		case opXorImm:
			result ^= param
		default:
			panic("bad drip")
		}
	}
	return result, pc
}

func (p *Program) contextValue(index int) uint32 {
	if index < 0 || index >= len(p.ContextU32) {
		panic("bad drip")
	}
	return p.ContextU32[index]
}

func NewFilterRuntimeState(seedState []byte) *FilterRuntimeState {
	nullMode := seedState[45] != 0
	state := &FilterRuntimeState{
		boundary0:   initBoundary(binary.LittleEndian.Uint64(seedState[0:8]), nullMode),
		boundary1:   initBoundary(binary.LittleEndian.Uint64(seedState[8:16]), nullMode),
		splitOffset: binary.LittleEndian.Uint64(seedState[16:24]),
	}
	if seedState[44] != 0 {
		state.bulkKey = append([]byte(nil), seedState[24:40]...)
	}
	return state
}

func (s *FilterRuntimeState) Apply(data []byte, offset uint64) bool {
	if len(data) == 0 {
		return false
	}
	end := offset + uint64(len(data))
	if len(s.bulkKey) > 0 && offset < uint64(len(s.bulkKey)) {
		overlapEnd := end
		if overlapEnd > uint64(len(s.bulkKey)) {
			overlapEnd = uint64(len(s.bulkKey))
		}
		for logical := offset; logical < overlapEnd; logical++ {
			data[logical-offset] ^= s.bulkKey[logical]
		}
	}
	if s.splitOffset <= offset {
		applyBoundary(data, offset, s.boundary1)
		return true
	}
	if s.splitOffset < end {
		firstSize := int(s.splitOffset - offset)
		applyBoundary(data[:firstSize], offset, s.boundary0)
		applyBoundary(data[firstSize:], s.splitOffset, s.boundary1)
		return true
	}
	applyBoundary(data, offset, s.boundary0)
	return true
}

func initBoundary(value uint64, nullMode bool) filterBoundary {
	pos0 := uint16(value >> 48)
	pos1 := uint16(value >> 32)
	if pos0 == pos1 {
		pos1++
	}
	keyByte := byte(value)
	byte0 := byte(value >> 8)
	byte1 := byte(value >> 16)
	if keyByte == 0 {
		if nullMode {
			keyByte = 0
		} else {
			keyByte = 0xa5
		}
	}
	if nullMode {
		byte0 = 0
		byte1 = 0
	}
	key := uint32(keyByte) * 0x01010101
	return filterBoundary{
		pos0:  pos0,
		pos1:  pos1,
		key:   key,
		byte0: byte0,
		byte1: byte1,
	}
}

func applyBoundary(data []byte, chunkStart uint64, boundary filterBoundary) {
	for i := range data {
		shift := ((chunkStart + uint64(i)) & 3) * 8
		data[i] ^= byte(boundary.key >> shift)
	}
	xorBoundaryByte(data, chunkStart, boundary.pos0, boundary.byte0)
	xorBoundaryByte(data, chunkStart, boundary.pos1, boundary.byte1)
}

func xorBoundaryByte(data []byte, chunkStart uint64, position uint16, value byte) {
	if value == 0 {
		return
	}
	pos := uint64(position)
	begin := chunkStart
	end := chunkStart + uint64(len(data))
	if pos < begin || pos >= end {
		return
	}
	data[pos-begin] ^= value
}

func FindDescriptor(index []byte) (Descriptor, bool, error) {
	return findDescriptor(index)
}

func CheckIndexPayload(payload []byte, key, nonce []byte) bool {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return false
	}
	if len(payload) < 16 {
		return false
	}
	sealed := append(append([]byte{}, payload[16:]...), payload[:16]...)
	_, err = aead.Open(nil, nonce, sealed, nil)
	return err == nil
}

func findDescriptor(index []byte) (Descriptor, bool, error) {
	for cursor := 0; cursor+12 <= len(index); {
		tag := string(index[cursor : cursor+4])
		size := binary.LittleEndian.Uint64(index[cursor+4 : cursor+12])
		cursor += 12
		end := cursor + int(size)
		body := index[cursor:end]
		cursor = end
		if tag != "Hxv4" {
			continue
		}
		return Descriptor{
			Offset: binary.LittleEndian.Uint64(body[0:8]),
			Size:   binary.LittleEndian.Uint32(body[8:12]),
			Flags:  binary.LittleEndian.Uint16(body[12:14]),
		}, true, nil
	}
	return Descriptor{}, false, nil
}

func decryptPayload(payload []byte, flags uint16, program *Program) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(program.HXV4Key)
	if err != nil {
		return nil, fmt.Errorf("init chacha20 fail")
	}
	nonce := program.HXV4Nonce0
	if flags&1 != 0 {
		nonce = program.HXV4Nonce1
	}
	sealed := make([]byte, 0, len(payload))
	sealed = append(sealed, payload[16:]...)
	sealed = append(sealed, payload[:16]...)
	out, err := aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return nil, fmt.Errorf("chacha decrypt fail")
	}
	return out, nil
}

func inflate(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

type tjsReader struct {
	data []byte
	pos  int
}

func newTJSReader(data []byte) *tjsReader {
	return &tjsReader{data: data}
}

func (r *tjsReader) readValue() (any, error) {
	tag, err := r.readU8()
	if err != nil {
		return nil, err
	}
	signedTag := int8(tag)
	switch signedTag {
	case 0, 1:
		return nil, nil
	case 2:
		return r.readString()
	case 3:
		size, err := r.readI32()
		if err != nil {
			return nil, err
		}
		return r.readBytes(int(size))
	case 4:
		value, err := r.readI64()
		if err != nil {
			return nil, err
		}
		return value, nil
	case 5:
		value, err := r.readF64()
		if err != nil {
			return nil, err
		}
		return value, nil
	case -127:
		count, err := r.readI32()
		if err != nil {
			return nil, err
		}
		values := make([]any, 0, count)
		for i := int32(0); i < count; i++ {
			value, err := r.readValue()
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
		return values, nil
	case -63:
		count, err := r.readI32()
		if err != nil {
			return nil, err
		}
		values := make(map[string]any, count)
		for i := int32(0); i < count; i++ {
			key, err := r.readString()
			if err != nil {
				return nil, err
			}
			value, err := r.readValue()
			if err != nil {
				return nil, err
			}
			values[key] = value
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unsupported TJS tag %d at 0x%x", signedTag, r.pos-1)
	}
}

func (r *tjsReader) readString() (string, error) {
	count, err := r.readI32()
	if err != nil {
		return "", err
	}
	if count < 0 {
		return "", fmt.Errorf("negative TJS string length")
	}
	raw, err := r.readBytes(int(count) * 2)
	if err != nil {
		return "", err
	}
	words := make([]uint16, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		words = append(words, binary.BigEndian.Uint16(raw[i:i+2]))
	}
	return string(utf16.Decode(words)), nil
}

func (r *tjsReader) readU8() (byte, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("truncated TJS payload")
	}
	value := r.data[r.pos]
	r.pos++
	return value, nil
}

func (r *tjsReader) readI32() (int32, error) {
	raw, err := r.readBytes(4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(raw)), nil
}

func (r *tjsReader) readI64() (int64, error) {
	raw, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(raw)), nil
}

func (r *tjsReader) readF64() (float64, error) {
	raw, err := r.readBytes(8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(raw)), nil
}

func (r *tjsReader) readBytes(size int) ([]byte, error) {
	end := r.pos + size
	if size < 0 || end < r.pos || end > len(r.data) {
		return nil, fmt.Errorf("truncated TJS payload")
	}
	value := append([]byte(nil), r.data[r.pos:end]...)
	r.pos = end
	return value, nil
}

func toInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int32:
		return int64(v), true
	case int:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}
