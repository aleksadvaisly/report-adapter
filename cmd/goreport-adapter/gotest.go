package main

import (
	"encoding/json"
	"io"
	"time"
)

type goTestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
	Output  string  `json:"Output,omitempty"`
}

type packageSummary struct {
	Elapsed float64
	Passed  int
	Failed  int
	Skipped int
}

func emitGoTestJSON(w io.Writer, tests []TestCase) error {
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)

	summaries := map[string]*packageSummary{}
	order := make([]string, 0)
	base := time.Now().UTC()

	for i, test := range tests {
		if _, ok := summaries[test.Package]; !ok {
			summaries[test.Package] = &packageSummary{}
			order = append(order, test.Package)
		}
		summary := summaries[test.Package]
		summary.Elapsed += test.Elapsed
		switch test.Status {
		case "fail":
			summary.Failed++
		case "skip":
			summary.Skipped++
		default:
			summary.Passed++
		}
		if err := encoder.Encode(goTestEvent{
			Time:    base.Add(time.Duration(i) * time.Millisecond).Format(time.RFC3339Nano),
			Action:  test.Status,
			Package: test.Package,
			Test:    test.Name,
			Elapsed: test.Elapsed,
			Output:  test.Output,
		}); err != nil {
			return err
		}
	}

	offset := len(tests)
	for i, pkg := range order {
		summary := summaries[pkg]
		action := "pass"
		switch {
		case summary.Failed > 0:
			action = "fail"
		case summary.Passed == 0 && summary.Skipped > 0:
			action = "skip"
		}
		if err := encoder.Encode(goTestEvent{
			Time:    base.Add(time.Duration(offset+i) * time.Millisecond).Format(time.RFC3339Nano),
			Action:  action,
			Package: pkg,
			Elapsed: summary.Elapsed,
		}); err != nil {
			return err
		}
	}

	return nil
}
