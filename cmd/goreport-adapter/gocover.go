package main

import (
	"fmt"
	"io"
	"sort"
)

func emitGoCover(w io.Writer, lines []CoverageLine) error {
	if _, err := io.WriteString(w, "mode: set\n"); err != nil {
		return err
	}

	byFile := map[string]map[int]int{}
	for _, line := range lines {
		if line.Path == "" || line.Line <= 0 {
			continue
		}
		if _, ok := byFile[line.Path]; !ok {
			byFile[line.Path] = map[int]int{}
		}
		if existing, ok := byFile[line.Path][line.Line]; !ok || line.Hits > existing {
			byFile[line.Path][line.Line] = line.Hits
		}
	}

	paths := make([]string, 0, len(byFile))
	for path := range byFile {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		lineNumbers := make([]int, 0, len(byFile[path]))
		for line := range byFile[path] {
			lineNumbers = append(lineNumbers, line)
		}
		sort.Ints(lineNumbers)
		for _, line := range lineNumbers {
			if _, err := fmt.Fprintf(w, "%s:%d.1,%d.1 1 %d\n", path, line, line+1, byFile[path][line]); err != nil {
				return err
			}
		}
	}

	return nil
}
