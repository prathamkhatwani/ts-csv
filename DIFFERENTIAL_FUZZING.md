# Differential Fuzz Survivor Report

## Executive Summary

This repository achieves **Differential Fuzz Survivor (+5 points)** status by running continuous differential fuzz testing between the original `@gregoranders/csv` TypeScript parser and the ported Go implementation for **65 continuous seconds** with **zero divergences** across **508,444 randomly generated inputs**.

---

## Fuzzing Results

| Metric | Result | Target / Requirement | Status |
|---|---|---|---|
| **Duration** | **65.00 continuous seconds** | ≥ 60 seconds | ✅ PASSED |
| **Total Inputs Fuzzed** | **508,444 inputs** | — | ✅ EXCEEDED |
| **Matches** | **508,444** | — | ✅ 100% Match |
| **Divergences** | **0** | **0** | ✅ PERFECT |
| **Fuzz Log File** | [`fuzz/differential_fuzz.log`](file:///c:/Users/Admin/OneDrive/Desktop/csvparser/fuzz/differential_fuzz.log) | Published log | ✅ PUBLISHED |

---

## Methodology & Architecture

1. **High-Performance Worker Process:** A persistent Go worker binary (`bin/fuzz-worker.exe`) reads line-delimited Base64 JSON payloads over `stdin` and writes JSON responses to `stdout`. This eliminates process instantiation overhead and enables **7,800+ differential iterations per second**.
2. **Dual Execution Harness:** The Node.js differential harness (`fuzz/differential_fuzzer.js`) feeds identical random inputs to:
   - Original TypeScript parser (`@gregoranders/csv` v0.0.13)
   - Ported Go parser (`github.com/prathamkhatwani/ts-csv`)
3. **Random Input Generator:** Generates a diverse mix of:
   - Random character strings (alphanumeric, punctuation, symbols).
   - Structured CSVs with random row/column counts, quotes, delimiters, newlines.
   - Escaped quotes (`\"` and `""`).
   - UTF-8 multi-byte characters, Chinese characters, and Emojis.
   - Unclosed quotes, trailing separators, and empty strings.
4. **Equivalence Checking:**
   - On valid CSV inputs: Confirms `JSON.stringify(tsRows) === JSON.stringify(goRows)`.
   - On malformed CSV inputs: Confirms both parsers reject the input and throw `ParseError`.

---

## Fuzzing Log Summary

```
======================================================================
DIFFERENTIAL FUZZ SURVIVOR HARNESS
Running continuous fuzzing against TS & Go parsers for 65s...
======================================================================
[1s / 65s]  Iterations: 8000   | Matches: 8000   | Divergences: 0
[10s / 65s] Iterations: 76000  | Matches: 76000  | Divergences: 0
[20s / 65s] Iterations: 146500 | Matches: 146500 | Divergences: 0
[30s / 65s] Iterations: 210500 | Matches: 210500 | Divergences: 0
[40s / 65s] Iterations: 288500 | Matches: 288500 | Divergences: 0
[50s / 65s] Iterations: 379500 | Matches: 379500 | Divergences: 0
[60s / 65s] Iterations: 471000 | Matches: 471000 | Divergences: 0
[65s / 65s] Iterations: 508444 | Matches: 508444 | Divergences: 0

======================================================================
FUZZING COMPLETE
Elapsed Time:    65.00 seconds (>= 60s requirement Met)
Total Inputs:    508444
Matches:         508444
Divergences:     0
Fuzz Log Saved:  fuzz/differential_fuzz.log
======================================================================

 SUCCESS: ZERO DIVERGENCES ACHIEVED! Differential Fuzz Survivor Qualified.
```

---

## How to Reproduce

To re-run the 65-second differential fuzzer:

```bash
# 1. Build the Go worker binary
go build -o bin/fuzz-worker.exe ./fuzz/worker/main.go

# 2. Run the differential fuzzer
node ./fuzz/differential_fuzzer.js
```
