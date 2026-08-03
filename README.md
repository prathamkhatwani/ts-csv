# ts-csv → Go Port

> **Go port of [`@gregoranders/csv`](https://github.com/gregoranders/ts-csv)** — a simple CSV parser, ported from TypeScript to idiomatic Go.

[![License](https://img.shields.io/github/license/prathamkhatwani/ts-csv.svg)](https://github.com/prathamkhatwani/ts-csv/blob/trial/LICENSE)

## Original Repository

This project is a complete port of the original TypeScript CSV parser by [@gregoranders](https://github.com/gregoranders):

- **Upstream Repo:** [https://github.com/gregoranders/ts-csv](https://github.com/gregoranders/ts-csv)
- **Upstream Version Ported:** `v0.0.13`
- **Original License:** MIT

---

## Quick Start

### Prerequisites

- **Go** 1.18 or higher ([download](https://go.dev/dl/))

### Build (One Step)

```bash
go build ./...
```

This compiles the core parser library and the CLI binary.

### Build the CLI Binary

```bash
go build -o bin/csv-cli ./src/cmd/csv-cli
```

### Run Tests

```bash
go test -v ./tests/port/...
```

### Run Benchmarks

```bash
go test -bench=. -benchmem ./tests/port/...
```

### Run Fuzz Tests

```bash
go test -fuzz=FuzzParse -fuzztime=10s ./tests/port/...
```

---

## Usage

```go
package main

import (
    "fmt"
    "log"

    csv "github.com/prathamkhatwani/ts-csv/src"
)

func main() {
    parser := csv.NewParser()

    rows, err := parser.Parse("name,age,city\nAlice,30,NYC\nBob,25,LA")
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }

    fmt.Println("Rows:", rows)
    fmt.Println("JSON:", parser.JSON())
}
```

**Output:**
```
Rows: [[name age city] [Alice 30 NYC] [Bob 25 LA]]
JSON: [map[age:30 city:NYC name:Alice] map[age:25 city:LA name:Bob]]
```

### Custom Delimiters

```go
parser := csv.NewParser(csv.Configuration{
    FieldSeparator: ";",
    LineSeparator:  "\t",
    Quote:          "'",
})
rows, _ := parser.Parse("a;b;c\t1;2;3")
```

---

## Project Structure

```
├── src/
│   ├── csv.go                  # Core parser (state machine)
│   └── cmd/csv-cli/main.go     # CLI binary (JSON over stdin/stdout)
├── tests/
│   └── port/
│       ├── csv_test.go          # Ported test suite (all original tests)
│       └── csv_bench_test.go    # Benchmark suite (small/medium/large)
├── fuzz/
│   ├── harness.go               # Go native fuzz harness
│   ├── worker/main.go           # IPC worker for differential fuzzing
│   ├── differential_fuzzer.js   # Differential fuzz harness (TS vs Go)
│   ├── differential_fuzz.log    # 65s fuzz log (508K inputs, 0 divergences)
│   └── ts_csv_bundled.js        # Bundled original TS parser for comparison
├── bench/
│   ├── methodology.md           # Benchmark report with methodology
│   └── results.json             # Machine-readable benchmark data
├── go.mod                       # Go module definition
├── BUILD.md                     # Build & run instructions
├── DECISIONS.md                 # 12 architectural divergences with rationale
├── BUG_REPORT.md                # Upstream bugs found via differential testing
├── DIFFERENTIAL_FUZZING.md      # 65s differential fuzz survivor report
├── ZERO_UNSAFE.md               # Zero unsafe audit report
└── README.md                    # This file
```

---

## Verification & Reports

| Deliverable | File | Status |
|---|---|---|
| Public repo with port | This repository | ✅ |
| One-step build command | [`BUILD.md`](./BUILD.md) — `go build ./...` | ✅ |
| Original test suite passing | [`tests/port/csv_test.go`](./tests/port/csv_test.go) | ✅ All pass |
| Differential fuzz harness | [`DIFFERENTIAL_FUZZING.md`](./DIFFERENTIAL_FUZZING.md) — 508K inputs, 0 divergences | ✅ |
| Decision log | [`DECISIONS.md`](./DECISIONS.md) — 12 architectural divergences | ✅ |
| Benchmark report | [`bench/methodology.md`](./bench/methodology.md) | ✅ |
| Bug report (upstream) | [`BUG_REPORT.md`](./BUG_REPORT.md) — 9 latent bugs found | ✅ |
| Zero unsafe audit | [`ZERO_UNSAFE.md`](./ZERO_UNSAFE.md) | ✅ |

---

## License

[MIT](./LICENSE) — Original work by [Gregor Anders](https://github.com/gregoranders), Go port by [Pratham Khatwani](https://github.com/prathamkhatwani).
