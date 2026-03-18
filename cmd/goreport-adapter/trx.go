package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type trxRun struct {
	Results     []trxUnitTestResult `xml:"Results>UnitTestResult"`
	Definitions []trxUnitTest       `xml:"TestDefinitions>UnitTest"`
}

type trxUnitTest struct {
	ID         string        `xml:"id,attr"`
	Name       string        `xml:"name,attr"`
	TestMethod trxTestMethod `xml:"TestMethod"`
}

type trxTestMethod struct {
	ClassName string `xml:"className,attr"`
}

type trxUnitTestResult struct {
	TestName string    `xml:"testName,attr"`
	Outcome  string    `xml:"outcome,attr"`
	Duration string    `xml:"duration,attr"`
	TestID   string    `xml:"testId,attr"`
	Output   trxOutput `xml:"Output"`
}

type trxOutput struct {
	StdOut    string       `xml:"StdOut"`
	ErrorInfo trxErrorInfo `xml:"ErrorInfo"`
}

type trxErrorInfo struct {
	Message    string `xml:"Message"`
	StackTrace string `xml:"StackTrace"`
}

func parseTRX(data []byte) ([]TestCase, error) {
	var run trxRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parse trx xml: %w", err)
	}

	classByID := make(map[string]string, len(run.Definitions))
	for _, def := range run.Definitions {
		if def.ID != "" && def.TestMethod.ClassName != "" {
			classByID[def.ID] = def.TestMethod.ClassName
		}
	}

	results := make([]TestCase, 0, len(run.Results))
	for _, item := range run.Results {
		status, ok := trxOutcomeToStatus(item.Outcome)
		if !ok {
			continue
		}
		output := strings.TrimSpace(strings.Join([]string{
			strings.TrimSpace(item.Output.StdOut),
			strings.TrimSpace(item.Output.ErrorInfo.Message),
			strings.TrimSpace(item.Output.ErrorInfo.StackTrace),
		}, "\n"))
		results = append(results, TestCase{
			Package: firstNonEmpty(classByID[item.TestID], "trx"),
			Name:    firstNonEmpty(item.TestName, item.TestID),
			Status:  status,
			Elapsed: parseTRXDuration(item.Duration),
			Output:  output,
		})
	}
	return results, nil
}

func trxOutcomeToStatus(outcome string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(outcome)) {
	case "passed":
		return "pass", true
	case "failed", "error", "timeout", "aborted":
		return "fail", true
	case "notexecuted", "skipped", "warning", "inconclusive":
		return "skip", true
	default:
		return "", false
	}
}
