# report-adapter

<p align="center"><img src="docs/goadapter4.png"></p>

Your test framework produces reports in its own format. Your coverage tool writes yet another.
`goreport-adapter` converts all of them into Go's standard formats so you can use one set of visualization tools across every language.

## Getting Started

Install the adapter and the rendering tools it pairs with:

```bash
go install github.com/aleksadvaisly/report-adapter/cmd/goreport-adapter@latest
go install github.com/vakenbolt/go-test-report@latest
go install github.com/nikolaydubina/go-cover-treemap@latest
```

The adapter reads from stdin and writes to stdout. Pick an input format with `--from` and an output format with `--to`:

```bash
cat pytest-report.xml | goreport-adapter --from=junit --to=gotest > test-report.json
```

That one command turned a JUnit XML file into Go's native test JSON. Pipe it into `go-test-report` for an HTML report:

```bash
cat test-report.json | go-test-report -o test-report.html
```

## Test Reports

The adapter normalizes test results from four frameworks into `go test -json` (NDJSON). Each produces the same output format - pass, fail, skip events with timing and captured output:

```bash
cat pytest-report.xml | goreport-adapter --from=junit --to=gotest > test-report.json
cat results.trx       | goreport-adapter --from=trx --to=gotest > test-report.json
cat jest-report.json  | goreport-adapter --from=jest --to=gotest > test-report.json
cargo test 2>&1       | goreport-adapter --from=cargo --to=gotest > test-report.json
```

The `cargo` parser reads directly from `cargo test` stdout - no intermediate file needed.

## Coverage Reports

Coverage data from four formats converts into Go's `coverage.out` profile:

```bash
cat coverage-py.json       | goreport-adapter --from=coverage-py --to=gocover > coverage.out
cat coverage.cobertura.xml | goreport-adapter --from=cobertura --to=gocover > coverage.out
cat llvm-cov-report.json   | goreport-adapter --from=llvm-cov --to=gocover > coverage.out
cat jest-report.json       | goreport-adapter --from=istanbul --to=gocover > coverage.out
```

Istanbul support covers both direct line maps (`l` field) and statement maps with hit counts.

Once you have `coverage.out`, render it as an SVG treemap:

```bash
go-cover-treemap -coverprofile=coverage.out > treemap.svg
```

## CI Integration

A typical pipeline step converts native reports and generates HTML/SVG artifacts:

```bash
pytest --junitxml=report.xml
cat report.xml | goreport-adapter --from=junit --to=gotest | go-test-report -o tests.html

coverage json -o cov.json
cat cov.json | goreport-adapter --from=coverage-py --to=gocover > coverage.out
go-cover-treemap -coverprofile=coverage.out > treemap.svg
```

The adapter has no external dependencies - it is a single static binary built on Go's stdlib.

## Limitations

The adapter stabilizes the `Package` field to a safe Go package name for `go-test-report` compatibility. Original source context is preserved in test names.

Coverage output uses a simplified line-based form: one entry per line with `stmt_count=1`. The `cargo` parser relies on stdout text patterns, making it more sensitive to output variations than the XML/JSON parsers.

## Development

```bash
go test ./...
```
