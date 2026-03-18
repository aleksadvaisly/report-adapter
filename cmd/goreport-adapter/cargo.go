package main

import (
	"bufio"
	"bytes"
	"regexp"
	"strings"
)

var (
	cargoTestLine    = regexp.MustCompile(`^test\s+(.+?)\s+\.\.\.\s+(ok|FAILED|ignored)$`)
	cargoRunningLine = regexp.MustCompile(`^Running\s+(?:unittests|tests)\s+(.+?)(?:\s+\(|$)`)
	cargoFailureLine = regexp.MustCompile(`^----\s+(.+?)\s+stdout\s+----$`)
)

func parseCargo(data []byte) ([]TestCase, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var results []TestCase
	indexByKey := map[string]int{}
	currentPackage := "cargo"
	currentFailureKey := ""
	var failureLines []string

	flushFailure := func() {
		if currentFailureKey == "" {
			return
		}
		if idx, ok := indexByKey[currentFailureKey]; ok {
			results[idx].Output = strings.TrimSpace(strings.Join(failureLines, "\n"))
		}
		currentFailureKey = ""
		failureLines = nil
	}

	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")

		if match := cargoRunningLine.FindStringSubmatch(line); len(match) == 2 {
			flushFailure()
			currentPackage = normalizePackagePath(match[1], "cargo")
			continue
		}

		if match := cargoFailureLine.FindStringSubmatch(line); len(match) == 2 {
			flushFailure()
			currentFailureKey = cargoKey(currentPackage, match[1])
			if _, ok := indexByKey[currentFailureKey]; !ok {
				currentFailureKey = findCargoFailureKey(indexByKey, match[1])
			}
			continue
		}

		if currentFailureKey != "" {
			if strings.HasPrefix(line, "---- ") && strings.HasSuffix(line, " stdout ----") {
				flushFailure()
				if match := cargoFailureLine.FindStringSubmatch(line); len(match) == 2 {
					currentFailureKey = cargoKey(currentPackage, match[1])
					if _, ok := indexByKey[currentFailureKey]; !ok {
						currentFailureKey = findCargoFailureKey(indexByKey, match[1])
					}
				}
				continue
			}
			if strings.HasPrefix(line, "failures:") || strings.HasPrefix(line, "test result:") {
				flushFailure()
			} else {
				failureLines = append(failureLines, line)
				continue
			}
		}

		if match := cargoTestLine.FindStringSubmatch(line); len(match) == 3 {
			status := "pass"
			switch match[2] {
			case "FAILED":
				status = "fail"
			case "ignored":
				status = "skip"
			}
			key := cargoKey(currentPackage, match[1])
			indexByKey[key] = len(results)
			results = append(results, TestCase{
				Package: currentPackage,
				Name:    match[1],
				Status:  status,
			})
		}
	}

	flushFailure()
	return results, scanner.Err()
}

func cargoKey(pkg, name string) string {
	return pkg + "\x00" + name
}

func findCargoFailureKey(indexByKey map[string]int, name string) string {
	suffix := "\x00" + name
	for key := range indexByKey {
		if strings.HasSuffix(key, suffix) {
			return key
		}
	}
	return ""
}
