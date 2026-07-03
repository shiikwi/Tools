package unpackxp3

import (
	"cxdecgo/xp3"
	"cxdecgo/xp3/hxv4"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Run(xp3Path string) error {
	workDir := filepath.Join(filepath.Dir(xp3Path), "bootstrap_dump")
	jsonPath := filepath.Join(workDir, "hxv4.json")
	outputRoot := filepath.Join(filepath.Dir(xp3Path), "CXDECOut")

	program, err := hxv4.LoadProgram(jsonPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(outputRoot, 0755); err != nil {
		return err
	}

	if err := unpackArchive(program, xp3Path, outputRoot); err != nil {
		return err
	}

	return nil
}

func unpackArchive(program *hxv4.Program, archivePath string, outputRoot string) error {
	info, entries, blob, index, err := xp3.ReadArchive(archivePath)
	if err != nil {
		return err
	}

	var descriptor *hxv4.Descriptor
	var records []hxv4.Record
	if d, r, err := hxv4.ParseTable(blob, index, entries, program); err == nil {
		descriptor = &d
		records = r
	}

	archiveDir := filepath.Join(outputRoot, strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath)))
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return err
	}

	manifest := struct {
		Archive         xp3.ArchiveInfo  `json:"archive"`
		HXV4            *hxv4.Descriptor `json:"hxv4"`
		Records         []map[string]any `json:"records"`
		SkippedWarnings int              `json:"skipped_warnings"`
		Written         int              `json:"written"`
		Failed          int              `json:"failed"`
	}{
		Archive: info,
		HXV4:    descriptor,
		Records: make([]map[string]any, 0, len(entries)),
	}

	hasWarning := len(entries) > 0 && xp3.IsWarningEntry(0, entries[0])
	for i, entry := range entries {
		item := map[string]any{
			"index": i,
			"name":  entry.Name,
		}
		if entry.Adler32 != nil {
			item["expected_adler32"] = fmt.Sprintf("0x%08x", *entry.Adler32)
		}

		hint := firstRecord(records, i, hasWarning)
		if hint != nil {
			item["domain_hash"] = hint.DomainHash
			item["file_hash"] = hint.FileHash
		}

		if xp3.IsWarningEntry(i, entry) {
			item["status"] = "skipped_warning"
			manifest.SkippedWarnings++
			manifest.Records = append(manifest.Records, item)
			continue
		}

		data, actual, _, used, _, err := recoverEntry(program, descriptor, records, blob, i, entry, hasWarning)
		if used != nil {
			item["domain_hash"] = used.DomainHash
			item["file_hash"] = used.FileHash
			// item["filter_open_flag"] = openFlag
		}
		if err != nil {
			item["status"] = "error"
			item["error"] = err.Error()
			item["actual_adler32"] = fmt.Sprintf("0x%08x", actual)
			if errors.Is(err, xp3.ErrFilter) {
				item["status"] = "missing_filter"
			}
			manifest.Failed++
			manifest.Records = append(manifest.Records, item)
			continue
		}

		outputPath := uniquePath(filepath.Join(append([]string{archiveDir}, outputParts(i, used)...)...))
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			item["status"] = "error"
			item["error"] = err.Error()
			manifest.Failed++
			manifest.Records = append(manifest.Records, item)
			continue
		}

		item["status"] = "written"
		item["actual_adler32"] = fmt.Sprintf("0x%08x", actual)
		if entry.Adler32 != nil {
			item["adler_ok"] = actual == *entry.Adler32
		}
		manifest.Written++
		manifest.Records = append(manifest.Records, item)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(filepath.Join(archiveDir, "manifest.json"), data, 0644); err != nil {
		return err
	}
	if manifest.Failed != 0 {
		return fmt.Errorf("check unpack fail")
	}
	return nil
}

func recoverEntry(program *hxv4.Program, descriptor *hxv4.Descriptor, records []hxv4.Record, blob []byte, index int, entry xp3.Entry, hasWarning bool) ([]byte, uint32, bool, *hxv4.Record, uint16, error) {
	data, actual, applied, err := xp3.ExtractEntry(blob, entry, nil)
	if err == nil && (entry.Adler32 == nil || actual == *entry.Adler32) {
		return data, actual, applied, nil, 0, nil
	}

	firstErr := err
	for _, record := range candidateRecords(records, index, hasWarning) {
		for _, openFlag := range openFlagCandidates(descriptor, record) {
			state := hxv4.NewFilterRuntimeState(program.BuildFilterState(record.Key, openFlag))
			tryData, tryActual, tryApplied, tryErr := xp3.ExtractEntry(blob, entry, state)
			if tryErr == nil && (entry.Adler32 == nil || tryActual == *entry.Adler32) {
				return tryData, tryActual, tryApplied, record, openFlag, nil
			}
			if firstErr == nil {
				firstErr = tryErr
			}
			data = tryData
			actual = tryActual
			applied = tryApplied
		}
	}
	if firstErr == nil {
		firstErr = xp3.ErrFilter
	}
	return data, actual, applied, nil, 0, firstErr
}

func candidateRecords(records []hxv4.Record, entryIndex int, hasWarning bool) []*hxv4.Record {
	out := make([]*hxv4.Record, 0)
	for i := range records {
		if uint16(records[i].Packed>>16) != 0 {
			continue
		}
		flag := int(uint16(records[i].Packed))
		if flag == entryIndex || hasWarning && flag+1 == entryIndex {
			out = append(out, &records[i])
		}
	}
	return out
}

func firstRecord(records []hxv4.Record, entryIndex int, hasWarning bool) *hxv4.Record {
	candidates := candidateRecords(records, entryIndex, hasWarning)
	if len(candidates) == 0 {
		return nil
	}
	return candidates[0]
}

func openFlagCandidates(descriptor *hxv4.Descriptor, record *hxv4.Record) []uint16 {
	values := make([]uint16, 0, 2)
	if descriptor != nil {
		values = append(values, descriptor.Flags&1)
	}
	value := uint16(record.Packed) & 1
	if len(values) == 0 || values[0] != value {
		values = append(values, value)
	}
	return values
}

func outputParts(index int, record *hxv4.Record) []string {
	if record == nil || record.DomainHash == "" || record.FileHash == "" {
		return []string{"unknown", fmt.Sprintf("%05d", index)}
	}
	return []string{record.DomainHash, record.FileHash}
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	for i := 1; i < 100000; i++ {
		candidate := fmt.Sprintf("%s.%d", path, i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return path
}
