# Benchmark Report

This document reports the performance characteristics of the Go CSV parser implementation under different input sizes.

## Benchmark Environment

- **OS**: Windows (11/10)
- **Architecture**: amd64
- **CPU**: AMD Ryzen 7 8845HS w/ Radeon 780M Graphics (16 logical CPUs)
- **Go version**: go1.21.6

## Performance Results

Benchmarks were executed using:
```bash
go test -bench="." -run="^$" -benchmem
```

| Benchmark Name | Iterations | Time per Op (ns) | Bytes per Op (B) | Allocations per Op |
| :--- | :--- | :--- | :--- | :--- |
| `BenchmarkParseSmall` | 613,704 | 1,861 ns/op | 572 B/op | 29 allocs/op |
| `BenchmarkParseMedium` | 21,542 | 58,481 ns/op | 22,544 B/op | 918 allocs/op |
| `BenchmarkParseLarge` | 2,464 | 522,428 ns/op | 238,514 B/op | 9,022 allocs/op |

## Analysis

1. **Scalability**: Performance scales linearly with the size of the input. A small 3-row string takes ~1.8 microseconds, whereas a large 1000-row string takes ~0.52 milliseconds.
2. **Allocations**: The parser processes inputs rune-by-rune, dynamically constructing cells, rows, and slices. Memory allocations are proportional to the number of cells/fields, which is expected for a general-purpose, custom-delimited parser state machine.
