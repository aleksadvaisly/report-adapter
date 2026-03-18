package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRun_JUnitToGoTest(t *testing.T) {
	input := `
<testsuites>
  <testsuite name="pytest">
    <testcase classname="test_csvmem" name="test_parse_valid" time="0.001"/>
    <testcase classname="test_csvmem" name="test_parse_empty" time="0.002">
      <failure message="AssertionError">expected 5 got 3</failure>
    </testcase>
    <testcase classname="test_csvmem" name="test_parse_skip" time="0.003">
      <skipped message="skip reason">not on this platform</skipped>
    </testcase>
  </testsuite>
</testsuites>`

	events := runGoTestConversion(t, "junit", input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	assertEvent(t, events[0], "test_csvmem", "test_parse_valid", "pass", 0.001, "")
	assertEvent(t, events[1], "test_csvmem", "test_parse_empty", "fail", 0.002, "AssertionError\nexpected 5 got 3")
	assertEvent(t, events[2], "test_csvmem", "test_parse_skip", "skip", 0.003, "skip reason\nnot on this platform")
	assertPackageEvent(t, events[3], "test_csvmem", "fail")
}

func TestRun_TRXToGoTest(t *testing.T) {
	input := `
<TestRun>
  <TestDefinitions>
    <UnitTest id="1" name="ParseValidCSV">
      <TestMethod className="CsvMem.Tests.ParserTests" />
    </UnitTest>
    <UnitTest id="2" name="ParseEmptyCSV">
      <TestMethod className="CsvMem.Tests.ParserTests" />
    </UnitTest>
    <UnitTest id="3" name="ParseSkippedCSV">
      <TestMethod className="CsvMem.Tests.ParserTests" />
    </UnitTest>
  </TestDefinitions>
  <Results>
    <UnitTestResult testId="1" testName="ParseValidCSV" outcome="Passed" duration="00:00:00.019" />
    <UnitTestResult testId="2" testName="ParseEmptyCSV" outcome="Failed" duration="00:00:00.005">
      <Output><ErrorInfo><Message>Assert.Throws failed</Message><StackTrace>at ParserTests.cs:42</StackTrace></ErrorInfo></Output>
    </UnitTestResult>
    <UnitTestResult testId="3" testName="ParseSkippedCSV" outcome="NotExecuted" duration="00:00:00.001" />
  </Results>
</TestRun>`

	events := runGoTestConversion(t, "trx", input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	assertEvent(t, events[0], "CsvMem.Tests.ParserTests", "ParseValidCSV", "pass", 0.019, "")
	assertEvent(t, events[1], "CsvMem.Tests.ParserTests", "ParseEmptyCSV", "fail", 0.005, "Assert.Throws failed\nat ParserTests.cs:42")
	assertEvent(t, events[2], "CsvMem.Tests.ParserTests", "ParseSkippedCSV", "skip", 0.001, "")
	assertPackageEvent(t, events[3], "CsvMem.Tests.ParserTests", "fail")
}

func TestRun_JestToGoTest(t *testing.T) {
	input := `{
  "testResults": [{
    "testFilePath": "/csvmem.test.js",
    "testResults": [
      {"title": "parse valid CSV", "status": "passed", "duration": 5},
      {"title": "parse empty CSV", "status": "failed", "duration": 3, "failureMessages": ["expected 5 got 3"]},
      {"title": "parse skipped CSV", "status": "pending", "duration": 1}
    ]
  }]
}`

	events := runGoTestConversion(t, "jest", input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	assertEvent(t, events[0], "/csvmem.test.js", "parse valid CSV", "pass", 0.005, "")
	assertEvent(t, events[1], "/csvmem.test.js", "parse empty CSV", "fail", 0.003, "expected 5 got 3")
	assertEvent(t, events[2], "/csvmem.test.js", "parse skipped CSV", "skip", 0.001, "")
	assertPackageEvent(t, events[3], "/csvmem.test.js", "fail")
}

func TestRun_CargoToGoTest(t *testing.T) {
	input := `
Running unittests src/lib.rs (target/debug/deps/csvmem-123)
test tests::parse_happy_path ... ok
test tests::parse_empty_csv ... FAILED
test tests::parse_ignored ... ignored

failures:

---- tests::parse_empty_csv stdout ----
thread 'tests::parse_empty_csv' panicked at src/lib.rs:42
expected 5 got 3

test result: FAILED. 1 passed; 1 failed; 1 ignored; 0 measured; 0 filtered out
`

	events := runGoTestConversion(t, "cargo", input)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
	assertEvent(t, events[0], "src/lib.rs", "tests::parse_happy_path", "pass", 0, "")
	assertEvent(t, events[1], "src/lib.rs", "tests::parse_empty_csv", "fail", 0, "thread 'tests::parse_empty_csv' panicked at src/lib.rs:42\nexpected 5 got 3")
	assertEvent(t, events[2], "src/lib.rs", "tests::parse_ignored", "skip", 0, "")
	assertPackageEvent(t, events[3], "src/lib.rs", "fail")
}

func TestRun_CoveragePyToGoCover(t *testing.T) {
	input := `{
  "files": {
    "csvmem.py": {
      "executed_lines": [1, 2, 3],
      "missing_lines": [5]
    }
  }
}`

	output := runCoverageConversion(t, "coverage-py", input)
	assertCoverageOutput(t, output, []string{
		"mode: set",
		"csvmem.py:1.1,2.1 1 1",
		"csvmem.py:2.1,3.1 1 1",
		"csvmem.py:3.1,4.1 1 1",
		"csvmem.py:5.1,6.1 1 0",
	})
}

func TestRun_CoberturaToGoCover(t *testing.T) {
	input := `
<coverage>
  <packages>
    <package name="CsvMem">
      <classes>
        <class filename="CsvMem.cs">
          <lines>
            <line number="10" hits="5" />
            <line number="11" hits="0" />
          </lines>
        </class>
      </classes>
    </package>
  </packages>
</coverage>`

	output := runCoverageConversion(t, "cobertura", input)
	assertCoverageOutput(t, output, []string{
		"mode: set",
		"CsvMem.cs:10.1,11.1 1 5",
		"CsvMem.cs:11.1,12.1 1 0",
	})
}

func TestRun_LLVMCovToGoCover(t *testing.T) {
	input := `{
  "data": [{
    "files": [{
      "filename": "src/lib.rs",
      "segments": [
        [10, 1, 5, true, true],
        [11, 1, 0, true, false],
        [13, 1, 2, true, true]
      ]
    }]
  }]
}`

	output := runCoverageConversion(t, "llvm-cov", input)
	assertCoverageOutput(t, output, []string{
		"mode: set",
		"src/lib.rs:10.1,11.1 1 5",
		"src/lib.rs:11.1,12.1 1 0",
		"src/lib.rs:12.1,13.1 1 0",
		"src/lib.rs:13.1,14.1 1 2",
	})
}

func TestRun_IstanbulToGoCover(t *testing.T) {
	input := `{
  "coverageMap": {
    "csvmem.js": {
      "l": {"1": 5, "2": 0, "3": 3}
    }
  }
}`

	output := runCoverageConversion(t, "istanbul", input)
	assertCoverageOutput(t, output, []string{
		"mode: set",
		"csvmem.js:1.1,2.1 1 5",
		"csvmem.js:2.1,3.1 1 0",
		"csvmem.js:3.1,4.1 1 3",
	})
}

func TestRun_UnsupportedCombination(t *testing.T) {
	var stdout bytes.Buffer
	err := run([]string{"--from=junit", "--to=gocover"}, strings.NewReader("<testsuite/>"), &stdout, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported coverage input format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func runGoTestConversion(t *testing.T, from, input string) []goTestEvent {
	t.Helper()

	var stdout bytes.Buffer
	if err := run([]string{"--from=" + from, "--to=gotest"}, strings.NewReader(input), &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	events := make([]goTestEvent, 0, len(lines))
	for _, line := range lines {
		var event goTestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("unmarshal event %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func runCoverageConversion(t *testing.T, from, input string) string {
	t.Helper()

	var stdout bytes.Buffer
	if err := run([]string{"--from=" + from, "--to=gocover"}, strings.NewReader(input), &stdout, &bytes.Buffer{}); err != nil {
		t.Fatalf("run failed: %v", err)
	}
	return strings.TrimSpace(stdout.String())
}

func assertEvent(t *testing.T, event goTestEvent, pkg, testName, action string, elapsed float64, output string) {
	t.Helper()

	if event.Package != pkg || event.Test != testName || event.Action != action {
		t.Fatalf("unexpected event: %+v", event)
	}
	if elapsed != 0 && event.Elapsed != elapsed {
		t.Fatalf("unexpected elapsed for %+v", event)
	}
	if output != "" && event.Output != output {
		t.Fatalf("unexpected output %q, want %q", event.Output, output)
	}
}

func assertPackageEvent(t *testing.T, event goTestEvent, pkg, action string) {
	t.Helper()

	if event.Package != pkg || event.Action != action {
		t.Fatalf("unexpected package event: %+v", event)
	}
	if event.Test != "" {
		t.Fatalf("package event must not contain test name: %+v", event)
	}
}

func assertCoverageOutput(t *testing.T, output string, expected []string) {
	t.Helper()

	actual := strings.Split(strings.TrimSpace(output), "\n")
	if len(actual) != len(expected) {
		t.Fatalf("unexpected line count\nactual: %v\nexpected: %v", actual, expected)
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Fatalf("line %d mismatch\nactual: %q\nexpected: %q", i, actual[i], expected[i])
		}
	}
}
