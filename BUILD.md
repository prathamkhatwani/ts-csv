# Build and Execution Instructions

This document provides instructions for compiling, running tests, benchmarking, and integrating the Go CSV parser library.

## Prerequisites

- **Go SDK**: Version 1.18 or higher is required.
  - Download it from the [official Go website](https://go.dev/dl/).

## Building the Module

To download dependencies and verify compilation of the package, run:

```bash
go build ./...
```

## Running Unit Tests

The package includes a comprehensive test suite covering standard and custom parser options, metadata, and error handling. To run all unit tests:

```bash
go test -v ./...
```

## Running Benchmarks

To run the parser benchmarks measuring throughput and memory allocation for small, medium, and large inputs:

```bash
go test -bench="." -run="^$" -benchmem
```

## Running Fuzz Tests

To run the fuzz testing framework for validating input robustness (e.g., for 10 seconds):

```bash
go test -fuzz=FuzzParse -fuzztime=10s
```

## Integration Example

Here is a simple example of how to import and use the parser in your Go applications:

```go
package main

import (
	"fmt"
	"log"

	"github.com/prathamkhatwani/ts-csv"
)

func main() {
	// 1. Initialize parser (with default config)
	parser := csv.NewParser()

	// 2. Parse CSV input
	input := "name,age,city\nJohn,30,New York\nJane,25,San Francisco"
	rows, err := parser.Parse(input)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	// 3. Retrieve rows
	fmt.Println("Rows:")
	for _, row := range rows {
		fmt.Printf("  %v\n", row)
	}

	// 4. Retrieve as JSON (key-value maps)
	jsonData := parser.JSON()
	fmt.Println("\nJSON:")
	for _, item := range jsonData {
		fmt.Printf("  %v\n", item)
	}
}
```
