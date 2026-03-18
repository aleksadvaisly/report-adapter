package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type junitDocument struct {
	XMLName   xml.Name        `xml:""`
	Name      string          `xml:"name,attr"`
	TestCases []junitTestCase `xml:"testcase"`
	Suites    []junitSuite    `xml:"testsuite"`
}

type junitSuite struct {
	Name      string          `xml:"name,attr"`
	TestCases []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	ClassName string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Time      string         `xml:"time,attr"`
	Failures  []junitMessage `xml:"failure"`
	Errors    []junitMessage `xml:"error"`
	Skipped   []junitMessage `xml:"skipped"`
	SystemOut string         `xml:"system-out"`
	SystemErr string         `xml:"system-err"`
}

type junitMessage struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",innerxml"`
}

func parseJUnit(data []byte) ([]TestCase, error) {
	var doc junitDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse junit xml: %w", err)
	}

	var suites []junitSuite
	switch doc.XMLName.Local {
	case "testsuite":
		suites = append(suites, junitSuite{Name: doc.Name, TestCases: doc.TestCases})
	default:
		suites = append(suites, doc.Suites...)
	}

	results := make([]TestCase, 0)
	for _, suite := range suites {
		for _, tc := range suite.TestCases {
			pkg := firstNonEmpty(tc.ClassName, suite.Name, "junit")
			status := "pass"
			output := strings.TrimSpace(strings.Join([]string{
				joinJUnitMessages(tc.Failures),
				joinJUnitMessages(tc.Errors),
				joinJUnitMessages(tc.Skipped),
				strings.TrimSpace(tc.SystemOut),
				strings.TrimSpace(tc.SystemErr),
			}, "\n"))

			switch {
			case len(tc.Skipped) > 0:
				status = "skip"
			case len(tc.Failures) > 0 || len(tc.Errors) > 0:
				status = "fail"
			}

			results = append(results, TestCase{
				Package: pkg,
				Name:    tc.Name,
				Status:  status,
				Elapsed: parseFloat(tc.Time),
				Output:  strings.TrimSpace(output),
			})
		}
	}
	return results, nil
}

func joinJUnitMessages(messages []junitMessage) string {
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		part := strings.TrimSpace(strings.Join([]string{
			strings.TrimSpace(msg.Message),
			stripXMLTags(msg.Text),
		}, "\n"))
		if part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, "\n")
}
