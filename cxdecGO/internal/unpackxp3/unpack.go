package unpackxp3

import (
	"cxdecgo/xp3"
	"cxdecgo/xp3/hxv4"
	namerecover "cxdecgo/xp3/recover"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func Run(xp3Path string, restore bool) error {
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

	if err := unpackArchive(program, xp3Path, outputRoot, restore); err != nil {
		return err
	}

	return nil
}

func unpackArchive(program *hxv4.Program, archivePath string, outputRoot string, restore bool) error {
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
		Records: make([]map[string]any, len(entries)),
	}

	hasWarning := len(entries) > 0 && xp3.IsWarningEntry(0, entries[0])
	namesPath := filepath.Join(outputRoot, "HxNames.lst")
	names := map[string]string{}
	if restore {
		names, err = namerecover.LoadList(namesPath)
		if err != nil {
			return err
		}
	}

	if len(entries) != 0 {
		workers := runtime.NumCPU()
		if workers > len(entries) {
			workers = len(entries)
		}
		jobs := make(chan int)
		done := make(chan struct{}, workers)
		for w := 0; w < workers; w++ {
			go func() {
				for i := range jobs {
					manifest.Records[i] = unpackEntry(program, descriptor, records, blob, archiveDir, outputRoot, i, entries[i], hasWarning)
				}
				done <- struct{}{}
			}()
		}
		for i := range entries {
			jobs <- i
		}
		close(jobs)
		for w := 0; w < workers; w++ {
			<-done
		}
	}

	for _, item := range manifest.Records {
		status, _ := item["status"].(string)
		switch status {
		case "written":
			manifest.Written++
		case "skipped_warning":
			manifest.SkippedWarnings++
		case "error", "missing_filter":
			manifest.Failed++
		}
	}

	if restore {
		names, err = namerecover.Build(archiveDir, names)
		if err != nil {
			return err
		}
		if err := namerecover.SaveList(namesPath, names); err != nil {
			return err
		}
		fmt.Printf("restoreFiles restore %d files\n", restoreNames(outputRoot, archiveDir, manifest.Records, names))
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

func unpackEntry(program *hxv4.Program, descriptor *hxv4.Descriptor, records []hxv4.Record, blob []byte, archiveDir string, outputRoot string, i int, entry xp3.Entry, hasWarning bool) map[string]any {
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
		return item
	}

	data, actual, _, used, _, err := recoverEntry(program, descriptor, records, blob, i, entry, hasWarning)
	record := used
	if record == nil {
		record = hint
	}
	if record != nil {
		item["domain_hash"] = record.DomainHash
		item["file_hash"] = record.FileHash
	}
	if err != nil {
		item["status"] = "error"
		item["error"] = err.Error()
		item["actual_adler32"] = fmt.Sprintf("0x%08x", actual)
		if errors.Is(err, xp3.ErrFilter) {
			item["status"] = "missing_filter"
		}
		return item
	}

	parts := []string{"unknown", fmt.Sprintf("%05d", i)}
	if record != nil && record.DomainHash != "" && record.FileHash != "" {
		parts = []string{record.DomainHash, record.FileHash}
	}
	outputPath := filepath.Join(append([]string{archiveDir}, parts...)...)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		item["status"] = "error"
		item["error"] = err.Error()
		return item
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		item["status"] = "error"
		item["error"] = err.Error()
		return item
	}

	if rel, err := filepath.Rel(outputRoot, outputPath); err == nil {
		item["output"] = filepath.ToSlash(rel)
	}
	item["status"] = "written"
	item["actual_adler32"] = fmt.Sprintf("0x%08x", actual)
	if entry.Adler32 != nil {
		item["adler_ok"] = actual == *entry.Adler32
	}
	return item
}

func restoreNames(outputRoot, archiveDir string, records []map[string]any, names map[string]string) int {
	count := 0
	for _, item := range records {
		if item["status"] != "written" {
			continue
		}
		domainHash, _ := item["domain_hash"].(string)
		fileHash, _ := item["file_hash"].(string)
		archiveName, ok := namerecover.Resolve(names, domainHash, fileHash)
		if !ok {
			continue
		}
		srcRel, _ := item["output"].(string)
		src := filepath.Join(outputRoot, filepath.FromSlash(srcRel))
		dst := filepath.Join(archiveDir, filepath.FromSlash(archiveName))
		if src == dst {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			continue
		}
		if err := os.Rename(src, dst); err != nil {
			continue
		}
		rel, err := filepath.Rel(outputRoot, dst)
		if err == nil {
			item["output"] = filepath.ToSlash(rel)
		}
		count++
	}
	return count
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
