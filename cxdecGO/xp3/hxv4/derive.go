package hxv4

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2s"
)

var xoptMarker = []byte{
	0x2d, 0x00, 0x2d, 0x00, 0x78, 0x00, 0x6f, 0x00, 0x70, 0x00, 0x74, 0x00,
	0x2d, 0x00, 0x2d, 0x00, 0x6e, 0x00, 0x6f, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var defaultArchiveSeed = []byte{0xce, 0xea, 0xaf, 0x2c, 0xef, 0xbe, 0xad, 0xde}

func utf16Bytes(value string) []byte {
	out := make([]byte, 0, len(value)*2)
	for _, r := range value {
		out = append(out, byte(r), byte(r>>8))
	}
	return out
}

func deriveKeyBlock(input []byte, seed uint32) []byte {
	words := make([]uint32, 8)
	state := 0x01000193 * (seed ^ 0x811c9dc5)
	for index, value := range input {
		mixed := uint32(value) ^ state
		mixed ^= mixed >> 17
		mixed *= 0xed5ad4bb
		mixed ^= mixed >> 11
		mixed *= 0xac4c1b51
		mixed ^= mixed >> 15
		mixed *= 0x31848bab
		state = mixed ^ (mixed >> 14)
		words[index%8] ^= state
	}
	prehash := make([]byte, 0, 32)
	for _, word := range words {
		prehash = append(prehash, byte(word), byte(word>>8), byte(word>>16), byte(word>>24))
	}
	sum := blake2s.Sum256(append(append([]byte{}, input...), prehash...))
	return sum[:]
}

func DeriveSharedKeys(params []byte, finalBootstrap string) (key, nonce1, material []byte, err error) {
	if len(params) < 22 {
		return nil, nil, nil, fmt.Errorf("hxv4 params truncated")
	}
	salt := make([]byte, 16)
	newHxSponge(0x90, 0x06).absorb(params).squeeze(salt)
	password := utf16Bytes(finalBootstrap)
	material = argon2.Key(password, salt, 3, 8, 1, 64)
	keyBlock := deriveKeyBlock(password, 0)
	nonceBlock := deriveKeyBlock(params, 1)
	key = make([]byte, 32)
	nonce1 = make([]byte, 24)
	for i := 0; i < 32; i++ {
		key[i] = keyBlock[i] ^ material[i]
		if i < 24 {
			nonce1[i] = nonceBlock[i] ^ material[i]
		}
	}
	return key, nonce1, material, nil
}

func DeriveNonce0(unique string, seed []byte) []byte {
	block := deriveKeyBlock(utf16Bytes(unique), 2)
	seedValue := binary.LittleEndian.Uint32(seed)
	modifier := deriveKeyBlock(seed, seedValue)
	nonce := make([]byte, 24)
	for i := 0; i < 24; i++ {
		nonce[i] = block[i] ^ modifier[i]
	}
	return nonce
}

func deriveControlWords(material []byte, mode byte) []uint32 {
	squeeze := make([]byte, 0x2000)
	newHxSponge(136, 0x1f).absorb(material).squeeze(squeeze)
	words := make([]uint32, 1024)
	for i := range words {
		chunk := append([]byte{}, squeeze[i*4:i*4+4]...)
		if mode == 2 {
			for j := 0; j < 4; j++ {
				chunk[j] ^= squeeze[0x1000+i*4+j]
			}
		}
		words[i] = binary.LittleEndian.Uint32(chunk)
	}
	return words
}

func validPermutation(data []byte) bool {
	seen := make([]bool, len(data))
	for _, value := range data {
		if int(value) >= len(data) || seen[value] {
			return false
		}
		seen[value] = true
	}
	return true
}

func DeriveProgram(params, material, nonce0 []byte) (*Program, error) {
	if len(params) < 22 {
		return nil, fmt.Errorf("hxv4 params truncated")
	}
	table := deriveControlWords(material, params[17]&1+1)
	return programFromTable(params, table, nonce0)
}

func programFromTable(params []byte, table []uint32, nonce0 []byte) (*Program, error) {
	if len(params) < 22 {
		return nil, fmt.Errorf("hxv4 params truncated")
	}
	if !validPermutation(params[0:8]) || !validPermutation(params[8:14]) || !validPermutation(params[14:17]) {
		return nil, fmt.Errorf("hxv4 params opcode maps are not permutations")
	}
	randomType := params[17] >> 7
	if randomType > 1 {
		return nil, fmt.Errorf("hxv4 params random type %d unsupported", randomType)
	}
	mask := binary.LittleEndian.Uint16(params[18:20])
	offset := binary.LittleEndian.Uint16(params[20:22])
	program := &Program{
		ContextU32: table,
		Lanes:      make([][]laneRecord, 128),
	}
	program.HolderWords[2] = binary.LittleEndian.Uint32(nonce0[0:4])
	program.HolderWords[3] = binary.LittleEndian.Uint32(nonce0[4:8])
	program.HolderWords[4] = uint32(mask)
	program.HolderWords[5] = uint32(offset)
	for lane := uint32(0); lane < 128; lane++ {
		records := buildLane(params, newLaneRng(lane, randomType))
		if records == nil {
			return nil, fmt.Errorf("hxv4 lane %d generation failed", lane)
		}
		program.Lanes[lane] = records
	}
	return program, nil
}

func ArchiveSeedCandidates(dll []byte) [][]byte {
	candidates := [][]byte{}
	add := func(seed []byte) {
		nonzero := false
		for _, value := range seed {
			if value != 0 {
				nonzero = true
			}
		}
		if !nonzero {
			return
		}
		for _, existing := range candidates {
			if bytes.Equal(existing, seed) {
				return
			}
		}
		candidates = append(candidates, append([]byte(nil), seed...))
	}
	if at := bytes.Index(dll, xoptMarker); at >= 0 && at+len(xoptMarker)+8 <= len(dll) {
		add(dll[at+len(xoptMarker) : at+len(xoptMarker)+8])
	}
	for pos := 0; pos+0x50 <= len(dll); pos++ {
		if dll[pos] != 0xc7 || dll[pos+1] != 0x45 || dll[pos+7] != 0xc7 || dll[pos+8] != 0x45 {
			continue
		}
		if dll[pos+14] != 0x0f || dll[pos+15] != 0x45 || dll[pos+16] != 0xf0 {
			continue
		}
		if dll[pos+9] != dll[pos+2]+4 {
			continue
		}
		if !uniqueKdfUse(dll[pos+17 : pos+0x50]) {
			continue
		}
		add(append(append([]byte{}, dll[pos+3:pos+7]...), dll[pos+10:pos+14]...))
	}
	add(defaultArchiveSeed)
	return candidates
}

func uniqueKdfUse(window []byte) bool {
	at := bytes.Index(window, []byte{0xff, 0x36})
	if at < 0 {
		return false
	}
	tail := window[at+2:]
	pair := bytes.Index(tail, []byte{0x6a, 0x08})
	return pair >= 0 && bytes.Contains(tail, []byte{0x56}) && bytes.Contains(tail, []byte{0xe8})
}
