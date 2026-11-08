package bootstrapdump

import (
	"compress/zlib"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cxdecgo/xp3/hxv4"
)

type hxv4Payload struct {
	Version    int      `json:"version"`
	Module     string   `json:"source_module"`
	Key        string   `json:"hxv4_key"`
	Nonce0     string   `json:"hxv4_nonce0"`
	Nonce1     string   `json:"hxv4_nonce1"`
	Holder     []uint32 `json:"holder_words"`
	Context    []uint32 `json:"context_u32"`
	CallbackRv uint32   `json:"callback_rva_base"`
	Lanes      []struct {
		Index   int       `json:"index"`
		Records [][]int64 `json:"records"`
	} `json:"lanes"`
}

func writeHxv4(workDir string, dll, params []byte, unique, warning, prefix, gameDir string) error {
	fmt.Println("==========gen Hxv4 phase==========")
	payload, flags, err := probeSpecialPayload(gameDir)
	if err != nil {
		return err
	}
	program, err := deriveProgram(dll, params, unique, warning, prefix, payload, flags)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(hxv4PayloadOf(program), "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(workDir, "hxv4.json"), data, 0644); err != nil {
		return err
	}
	fmt.Println("write hxv4.json success")
	return nil
}

func deriveProgram(dll, params []byte, unique, warning, prefix string, payload []byte, flags uint16) (*hxv4.Program, error) {
	prefixes := []string{prefix}
	if prefix == "" || prefix != unique {
		prefixes = append(prefixes, unique)
	}
	seedCandidates := hxv4.ArchiveSeedCandidates(dll)
	var lastErr error = fmt.Errorf("hxv4 derivation had no candidates")
	for _, candidate := range prefixes {
		key, nonce1, material, err := hxv4.DeriveSharedKeys(params, candidate+warning)
		if err != nil {
			lastErr = err
			continue
		}
		for _, seed := range seedCandidates {
			nonce0 := hxv4.DeriveNonce0(unique, seed)
			nonce := nonce0
			if flags&1 != 0 {
				nonce = nonce1
			}
			if !hxv4.CheckIndexPayload(payload, key, nonce) {
				lastErr = fmt.Errorf("hxv4 index authentication failed for all candidates")
				continue
			}
			program, err := hxv4.DeriveProgram(params, material, nonce0)
			if err != nil {
				return nil, err
			}
			program.HXV4Key = key
			program.HXV4Nonce0 = nonce0
			program.HXV4Nonce1 = nonce1
			return program, nil
		}
	}
	return nil, lastErr
}

func probeSpecialPayload(gameDir string) ([]byte, uint16, error) {
	matches, err := filepath.Glob(filepath.Join(gameDir, "*.xp3"))
	if err != nil || len(matches) == 0 {
		return nil, 0, fmt.Errorf("no xp3 archive found for hxv4 validation")
	}
	sort.Slice(matches, func(i, j int) bool {
		left, right := fileSize(matches[i]), fileSize(matches[j])
		return left < right
	})
	var lastErr error = fmt.Errorf("no hxv4 descriptor found in any archive")
	for _, path := range matches {
		payload, flags, err := probeArchive(path)
		if err == nil {
			return payload, flags, nil
		}
		lastErr = err
	}
	return nil, 0, lastErr
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func probeArchive(path string) ([]byte, uint16, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()
	head := make([]byte, 0x28)
	if _, err := io.ReadFull(file, head); err != nil {
		return nil, 0, err
	}
	indexOffset := binary.LittleEndian.Uint64(head[11:19])
	if indexOffset == 0x17 {
		candidate := binary.LittleEndian.Uint64(head[0x20:0x28])
		if candidate > 0 && candidate < uint64(fileSize(path)) {
			indexOffset = candidate
		}
	}
	meta := make([]byte, 17)
	if _, err := file.ReadAt(meta, int64(indexOffset)); err != nil {
		return nil, 0, err
	}
	var index []byte
	switch meta[0] {
	case 0:
		size := binary.LittleEndian.Uint64(meta[1:9])
		index = make([]byte, size)
		if _, err := file.ReadAt(index, int64(indexOffset)+17); err != nil {
			return nil, 0, err
		}
	case 1:
		compressed := binary.LittleEndian.Uint64(meta[1:9])
		raw := make([]byte, compressed)
		if _, err := file.ReadAt(raw, int64(indexOffset)+17); err != nil {
			return nil, 0, err
		}
		reader, err := zlib.NewReader(strings.NewReader(string(raw)))
		if err != nil {
			return nil, 0, err
		}
		index, err = io.ReadAll(reader)
		if err != nil {
			return nil, 0, err
		}
		reader.Close()
	default:
		return nil, 0, fmt.Errorf("unknown index flag %d in %s", meta[0], path)
	}
	descriptor, found, err := hxv4.FindDescriptor(index)
	if err != nil || !found {
		return nil, 0, fmt.Errorf("no hxv4 descriptor in %s", path)
	}
	payload := make([]byte, descriptor.Size)
	if _, err := file.ReadAt(payload, int64(descriptor.Offset)); err != nil {
		return nil, 0, err
	}
	return payload, descriptor.Flags, nil
}

func hxv4PayloadOf(program *hxv4.Program) *hxv4Payload {
	out := &hxv4Payload{
		Version:    1,
		Module:     "bootstrap.dll",
		Key:        hexEncode(program.HXV4Key),
		Nonce0:     hexEncode(program.HXV4Nonce0),
		Nonce1:     hexEncode(program.HXV4Nonce1),
		Holder:     program.HolderWords[:],
		Context:    program.ContextU32,
		CallbackRv: 0,
	}
	for index, records := range program.Lanes {
		lane := struct {
			Index   int       `json:"index"`
			Records [][]int64 `json:"records"`
		}{Index: index, Records: make([][]int64, 0, len(records))}
		for _, record := range records {
			lane.Records = append(lane.Records, []int64{int64(record.Param), int64(record.Op)})
		}
		out.Lanes = append(out.Lanes, lane)
	}
	return out
}

func hexEncode(data []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 0, len(data)*2)
	for _, value := range data {
		out = append(out, digits[value>>4], digits[value&0xf])
	}
	return string(out)
}
