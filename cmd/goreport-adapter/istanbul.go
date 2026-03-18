package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

type istanbulLocation struct {
	Line int `json:"line"`
}

type istanbulStatement struct {
	Start istanbulLocation `json:"start"`
	End   istanbulLocation `json:"end"`
}

type istanbulFile struct {
	Lines        map[string]int               `json:"l"`
	StatementMap map[string]istanbulStatement `json:"statementMap"`
	Statements   map[string]int               `json:"s"`
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
		if len(file.Lines) > 0 {
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
			continue
		}

		if len(file.StatementMap) == 0 || len(file.Statements) == 0 {
			return nil, fmt.Errorf("istanbul file %q does not contain line coverage map 'l' or statement coverage map 'statementMap'/'s'", path)
		}

		coveredLines := map[int]int{}
		statementIDs := make([]int, 0, len(file.StatementMap))
		for key := range file.StatementMap {
			id, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("istanbul invalid statement id %q in %q: %w", key, path, err)
			}
			statementIDs = append(statementIDs, id)
		}
		sort.Ints(statementIDs)
		for _, id := range statementIDs {
			key := strconv.Itoa(id)
			stmt, ok := file.StatementMap[key]
			if !ok {
				return nil, fmt.Errorf("istanbul missing statementMap entry %q in %q", key, path)
			}
			hits, ok := file.Statements[key]
			if !ok {
				return nil, fmt.Errorf("istanbul missing statement hit count %q in %q", key, path)
			}
			startLine := stmt.Start.Line
			endLine := stmt.End.Line
			if startLine <= 0 || endLine <= 0 || endLine < startLine {
				return nil, fmt.Errorf("istanbul invalid statement range %q in %q: %d-%d", key, path, startLine, endLine)
			}
			for line := startLine; line <= endLine; line++ {
				if existing, ok := coveredLines[line]; !ok || hits > existing {
					coveredLines[line] = hits
				}
			}
		}

		lineNumbers := make([]int, 0, len(coveredLines))
		for line := range coveredLines {
			lineNumbers = append(lineNumbers, line)
		}
		sort.Ints(lineNumbers)
		for _, line := range lineNumbers {
			lines = append(lines, CoverageLine{
				Path: path,
				Line: line,
				Hits: coveredLines[line],
			})
		}
	}
	return lines, nil
}
