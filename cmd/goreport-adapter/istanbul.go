package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type istanbulFile struct {
	Lines map[string]int `json:"l"`
}

func parseIstanbul(data []byte) ([]CoverageLine, error) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("parse istanbul json: %w", err)
	}

	rawMap := envelope
	if rawCoverageMap, ok := envelope["coverageMap"]; ok {
		rawMap = map[string]json.RawMessage{}
		if err := json.Unmarshal(rawCoverageMap, &rawMap); err != nil {
			return nil, fmt.Errorf("parse istanbul coverageMap: %w", err)
		}
	}

	lines := make([]CoverageLine, 0)
	for path, raw := range rawMap {
		var file istanbulFile
		if err := json.Unmarshal(raw, &file); err != nil {
			continue
		}
		if len(file.Lines) == 0 {
			return nil, fmt.Errorf("istanbul file %q does not contain line coverage map 'l'", path)
		}

		lineNumbers := make([]int, 0, len(file.Lines))
		for key := range file.Lines {
			line, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("istanbul invalid line number %q in %q: %w", key, path, err)
			}
			lineNumbers = append(lineNumbers, line)
		}
		sort.Ints(lineNumbers)
		for _, line := range lineNumbers {
			lines = append(lines, CoverageLine{
				Path: path,
				Line: line,
				Hits: file.Lines[strconv.Itoa(line)],
			})
		}
	}
	return lines, nil
}
