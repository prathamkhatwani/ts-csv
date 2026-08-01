# ts-csv (Ported to Go)

A simple, fast, and robust CSV parser in Go, ported from the original TypeScript codebase.

[![Go Reference](https://pkg.go.dev/badge/github.com/prathamkhatwani/ts-csv.svg)](https://pkg.go.dev/github.com/prathamkhatwani/ts-csv)
[![Go Report Card](https://goreportcard.com/badge/github.com/prathamkhatwani/ts-csv)](https://goreportcard.com/report/github.com/prathamkhatwani/ts-csv)
[![Build Status](https://github.com/prathamkhatwani/ts-csv/workflows/Go/badge.svg)](https://github.com/prathamkhatwani/ts-csv/actions)
[![License](https://img.shields.io/github/license/prathamkhatwani/ts-csv.svg)](./LICENSE)

## Features

- **Rune-by-Rune Parser State Machine**: Identical character-by-character parsing state machine as the original TS version, ensuring 100% functional equivalence.
- **Custom Delimiters**: Configurable field separator, line separator, and quote characters.
- **Robust Escaping Support**: Handles quote character escaping via double-quotes or backslashes.
- **JSON Mapping**: Native method to map parsed CSV rows into key-value maps (`[]map[string]string`) using the first row as headers.
- **Native Tests & Benchmarks**: Includes a robust suite of unit tests, performance benchmarks, and automated fuzzing.

## Example

### Code Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/prathamkhatwani/ts-csv"
)

func main() {
	parser := csv.NewParser()
	rows, err := parser.Parse("a,b,c\n1,2,3\n4,5,6")
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// 1. Get raw parsed rows (equivalent to parser.rows in JS)
	fmt.Println(rows)
	// Output: [[a b c] [1 2 3] [4 5 6]]

	// 2. Get JSON object representation (equivalent to parser.json in JS)
	jsonData := parser.JSON()
	fmt.Println(jsonData)
	// Output: [map[a:1 b:2 c:3] map[a:4 b:5 c:6]]
}
```

## Setup & Execution

For detailed build, test, and optimization instructions, please see the [BUILD.md](./BUILD.md) file.

### Clone repository

```bash
git clone https://github.com/prathamkhatwani/ts-csv
```

### Running Tests

```bash
go test -v ./...
```

### Running Benchmarks

```bash
go test -bench="." -run="^$" -benchmem
```
