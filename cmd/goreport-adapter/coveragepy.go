package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

type coveragePyReport struct {
	Files map[string]coveragePyFile `json:"files"`
}

type coveragePyFile struct {
	ExecutedLines []int `json:"executed_lines"`
	MissingLines  []int `json:"missing_lines"`
}

func parseCoveragePy(data []byte) ([]CoverageLine, error) {
	var report coveragePyReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse coverage.py json: %w", err)
	}

	lines := make([]CoverageLine, 0)
	for path, file := range report.Files {
		sort.Ints(file.ExecutedLines)
		sort.Ints(file.MissingLines)
		for _, line := range file.ExecutedLines {
			lines = append(lines, CoverageLine{Path: path, Line: line, Hits: 1})
		}
		for _, line := range file.MissingLines {
			lines = append(lines, CoverageLine{Path: path, Line: line, Hits: 0})
		}
	}
	return lines, nil
}
