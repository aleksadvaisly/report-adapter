package main

import (
	"encoding/json"
	"io"
	"strings"
	"time"
)

const goTestReportPackage = "fmt"

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

	tick := 0
	emit := func(action, pkg, test, output string, elapsed float64) error {
		tick++
		return encoder.Encode(goTestEvent{
			Time:    base.Add(time.Duration(tick) * time.Millisecond).Format(time.RFC3339Nano),
			Action:  action,
			Package: pkg,
			Test:    test,
			Elapsed: elapsed,
			Output:  output,
		})
	}

	for _, rawTest := range tests {
		test := normalizeForGoTestReport(rawTest)
		if _, ok := summaries[test.Package]; !ok {
			summaries[test.Package] = &packageSummary{}
			order = append(order, test.Package)
			if err := emit("start", test.Package, "", "", 0); err != nil {
				return err
			}
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
		if err := emit("run", test.Package, test.Name, "", 0); err != nil {
			return err
		}
		if err := emit("output", test.Package, test.Name, "=== RUN   "+test.Name+"\n", 0); err != nil {
			return err
		}
		if test.Output != "" {
			if err := emit("output", test.Package, test.Name, test.Output+"\n", 0); err != nil {
				return err
			}
		}
		tag := "--- PASS: "
		if test.Status == "fail" {
			tag = "--- FAIL: "
		} else if test.Status == "skip" {
			tag = "--- SKIP: "
		}
		if err := emit("output", test.Package, test.Name, tag+test.Name+"\n", 0); err != nil {
			return err
		}
		if err := emit(test.Status, test.Package, test.Name, "", test.Elapsed); err != nil {
			return err
		}
	}

	for _, pkg := range order {
		summary := summaries[pkg]
		action := "pass"
		switch {
		case summary.Failed > 0:
			action = "fail"
		case summary.Passed == 0 && summary.Skipped > 0:
			action = "skip"
		}
		if err := emit(action, pkg, "", "", summary.Elapsed); err != nil {
			return err
		}
	}

	return nil
}

func normalizeForGoTestReport(test TestCase) TestCase {
	logicalPackage := strings.TrimSpace(test.Package)
	test.Package = goTestReportPackage
	if logicalPackage == "" {
		return test
	}
	prefix := logicalPackage + "/"
	if !strings.HasPrefix(test.Name, prefix) {
		test.Name = prefix + test.Name
	}
	return test
}
