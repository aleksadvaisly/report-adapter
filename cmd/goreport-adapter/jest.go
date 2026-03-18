package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

type jestReport struct {
	TestResults []jestSuite `json:"testResults"`
}

type jestSuite struct {
	TestFilePath string           `json:"testFilePath"`
	TestResults  []jestCaseResult `json:"testResults"`
	FailureMsg   string           `json:"message"`
}

type jestCaseResult struct {
	Title           string   `json:"title"`
	FullName        string   `json:"fullName"`
	Status          string   `json:"status"`
	Duration        float64  `json:"duration"`
	FailureMessages []string `json:"failureMessages"`
}

func parseJest(data []byte) ([]TestCase, error) {
	var report jestReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parse jest json: %w", err)
	}

	results := make([]TestCase, 0)
	for _, suite := range report.TestResults {
		pkg := normalizePackagePath(suite.TestFilePath, "jest")
		for _, tc := range suite.TestResults {
			status, ok := jestStatusToAction(tc.Status)
			if !ok {
				continue
			}
			name := firstNonEmpty(tc.FullName, tc.Title)
			output := strings.TrimSpace(strings.Join(tc.FailureMessages, "\n"))
			if output == "" && status == "fail" {
				output = strings.TrimSpace(suite.FailureMsg)
			}
			results = append(results, TestCase{
				Package: pkg,
				Name:    name,
				Status:  status,
				Elapsed: tc.Duration / 1000.0,
				Output:  output,
			})
		}
	}
	return results, nil
}

func jestStatusToAction(status string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "passed":
		return "pass", true
	case "failed":
		return "fail", true
	case "pending", "skipped", "todo", "disabled":
		return "skip", true
	default:
		return "", false
	}
}

func normalizePackagePath(path, fallback string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return fallback
	}
	cleaned := filepath.ToSlash(filepath.Clean(path))
	if cleaned == "." {
		return fallback
	}
	return cleaned
}
