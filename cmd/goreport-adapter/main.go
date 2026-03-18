package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("goreport-adapter", flag.ContinueOnError)
	fs.SetOutput(stderr)

	from := fs.String("from", "", "input format")
	to := fs.String("to", "", "output format")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *from == "" || *to == "" {
		return fmt.Errorf("both --from and --to are required")
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}

	switch *to {
	case "gotest":
		tests, err := parseTests(*from, data)
		if err != nil {
			return err
		}
		return emitGoTestJSON(stdout, tests)
	case "gocover":
		lines, err := parseCoverage(*from, data)
		if err != nil {
			return err
		}
		return emitGoCover(stdout, lines)
	default:
		return fmt.Errorf("unsupported --to format %q", *to)
	}
}

func parseTests(from string, data []byte) ([]TestCase, error) {
	switch from {
	case "junit":
		return parseJUnit(data)
	case "trx":
		return parseTRX(data)
	case "jest":
		return parseJest(data)
	case "cargo":
		return parseCargo(data)
	default:
		return nil, fmt.Errorf("unsupported test input format %q", from)
	}
}

func parseCoverage(from string, data []byte) ([]CoverageLine, error) {
	switch from {
	case "coverage-py":
		return parseCoveragePy(data)
	case "cobertura":
		return parseCobertura(data)
	case "llvm-cov":
		return parseLLVMCov(data)
	case "istanbul":
		return parseIstanbul(data)
	default:
		return nil, fmt.Errorf("unsupported coverage input format %q", from)
	}
}
