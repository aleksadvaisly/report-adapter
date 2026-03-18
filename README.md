# report-adapter

`report-adapter` to adapter formatów raportów testów i coverage.
Konwertuje natywne formaty z różnych ekosystemów do formatów narzędzi Go:

- testy -> `go test -json` (NDJSON)
- coverage -> `coverage.out` (Go cover profile)

Cel v1 jest prosty: ujednolicić wejście, a rendering HTML/SVG zostawić istniejącym narzędziom Go.

## Status

Aktualnie zaimplementowane jest podejście v1: adapter formatów wejściowych do `go test -json` i `coverage.out`.

Obsługiwane wejścia:

- testy: `junit`, `trx`, `jest`, `cargo`
- coverage: `coverage-py`, `cobertura`, `llvm-cov`, `istanbul`

Obsługiwane wyjścia:

- `gotest` -> `go test -json`
- `gocover` -> `coverage.out`

## Instalacja

```bash
go install github.com/aleksadvaisly/report-adapter/cmd/goreport-adapter@latest
go install github.com/vakenbolt/go-test-report@latest
go install github.com/nikolaydubina/go-cover-treemap@latest
```

## Użycie

### Testy -> `go test -json`

```bash
cat pytest-report.xml | goreport-adapter --from=junit --to=gotest > test-report.json
cat results.trx       | goreport-adapter --from=trx --to=gotest > test-report.json
cat jest-report.json  | goreport-adapter --from=jest --to=gotest > test-report.json
cargo test 2>&1       | goreport-adapter --from=cargo --to=gotest > test-report.json
```

Potem HTML:

```bash
cat test-report.json | go-test-report -o test-report.html
```

### Coverage -> `coverage.out`

```bash
cat coverage-py.json       | goreport-adapter --from=coverage-py --to=gocover > coverage.out
cat coverage.cobertura.xml | goreport-adapter --from=cobertura --to=gocover > coverage.out
cat llvm-cov-report.json   | goreport-adapter --from=llvm-cov --to=gocover > coverage.out
cat jest-report.json       | goreport-adapter --from=istanbul --to=gocover > coverage.out
```

Potem raport coverage:

```bash
go-cover-treemap -coverprofile=coverage.out > treemap.svg
```

## Format wejściowy i wyjściowy

### `--from`

| Flaga | Format | Typ |
|---|---|---|
| `junit` | JUnit XML | testy |
| `trx` | Visual Studio TRX XML | testy |
| `jest` | Jest JSON | testy |
| `cargo` | `cargo test` stdout | testy |
| `coverage-py` | coverage.py JSON | coverage |
| `cobertura` | Cobertura XML | coverage |
| `llvm-cov` | llvm-cov export JSON | coverage |
| `istanbul` | Istanbul/Jest coverageMap | coverage |

### `--to`

| Flaga | Format |
|---|---|
| `gotest` | `go test -json` |
| `gocover` | `coverage.out` |

## Ograniczenia

- `coverage.out` jest emitowane w uproszczonej formie liniowej: jeden wpis na linię, `stmt_count=1`.
- Parser `cargo` opiera się na wzorcach tekstowych ze `stdout`, więc jest bardziej wrażliwy na warianty outputu niż parsery XML/JSON.
- Parser Istanbul wymaga mapy pokrycia linii `l`.

## Development

Uruchomienie testów:

```bash
go test ./...
```

Główna implementacja znajduje się w:

- `cmd/goreport-adapter/main.go`
- `cmd/goreport-adapter/*.go`
