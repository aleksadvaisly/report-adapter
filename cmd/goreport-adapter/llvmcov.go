package main

import (
	"encoding/json"
	"fmt"
	"sort"
)

type llvmCovReport struct {
	Data []llvmCovData `json:"data"`
}

type llvmCovData struct {
	Files []llvmCovFile `json:"files"`
}

type llvmCovFile struct {
	FileName string           `json:"filename"`
	Segments [][]llvmCovValue `json:"segments"`
}

type llvmCovValue interface{}

func parseLLVMCov(data []byte) ([]CoverageLine, error) {
	var report llvmCovReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse llvm-cov json: %w", err)
	}

	lines := make([]CoverageLine, 0)
	for _, chunk := range report.Data {
		for _, file := range chunk.Files {
			fileLines := map[int]int{}
			for i, segment := range file.Segments {
				if len(segment) < 4 {
					continue
				}
				line := intFromValue(segment[0])
				count := intFromValue(segment[2])
				hasCount, ok := boolFromValue(segment[3])
				if !ok || !hasCount || line <= 0 {
					continue
				}
				nextLine := line + 1
				if i+1 < len(file.Segments) && len(file.Segments[i+1]) > 0 {
					nextCandidate := intFromValue(file.Segments[i+1][0])
					if nextCandidate > line {
						nextLine = nextCandidate
					}
				}
				for current := line; current < nextLine; current++ {
					fileLines[current] = count
				}
			}
			lineNumbers := make([]int, 0, len(fileLines))
			for line := range fileLines {
				lineNumbers = append(lineNumbers, line)
			}
			sort.Ints(lineNumbers)
			for _, line := range lineNumbers {
				lines = append(lines, CoverageLine{
					Path: file.FileName,
					Line: line,
					Hits: fileLines[line],
				})
			}
		}
	}
	return lines, nil
}
