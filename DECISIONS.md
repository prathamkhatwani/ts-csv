# Architectural Decisions & Implementation Trade-Offs

This document outlines the architectural decisions and design trade-offs made during the porting of the `@gregoranders/csv` library from TypeScript to Go.

## 1. Porting Strategy & API Mapping
- **TypeScript Getters as Go Methods**: The TypeScript class uses JavaScript getters for `rows` and `json`. In Go, we map these to standard methods `Rows()` and `JSON()`.
- **Constructor Options**: We implement optional parameters in the Go constructor `NewParser(config ...Configuration)` to mimic TypeScript's default constructor parameters and override behaviors.

## 2. Handling Immutability
- **TypeScript Behavior**: The TS parser calls `Object.freeze` on all parsed cells, rows, and the JSON objects to guarantee immutability.
- **Go Trade-Off**: Go does not support language-level immutability for slices or maps without custom wrappers (which degrade developer ergonomics and performance). We made the decision to return standard Go slices (`[]Row`) and maps (`[]map[string]string`) as they are idiomatic, easy to use, and highly performant.

## 3. String & Rune Representation
- **Index Alignment**: TypeScript indexes strings by 16-bit UTF-16 code units. To ensure precise alignment of index counting, offsets, and line tracking, we convert the Go string input into a slice of runes (`[]rune`) inside `Parse()`.
- **Rune-by-Rune Processing**: Processing the input rune-by-rune matches the TS character-by-character parser state machine perfectly, avoiding discrepancies with multi-byte Unicode characters.

## 4. Error Propagation
- **Exceptions to Idiomatic Errors**: Instead of throwing exceptions (`throw new ParseError`), we return errors explicitly from our internal state machine helpers and propagate them up to the public `Parse()` method, which returns `([]Row, error)`.
- **Custom Error Type**: A custom `ParseError` struct implements Go's `error` interface to match the exact string format of the TS counterpart (`Invalid CSV at line:col`).

## 5. Testing & Fuzzing
- **Test Coverage**: We ported all valid and invalid unit test suites from `index.spec.ts` to `csv_test.go` using the `testing` package.
- **Fuzz Testing**: Added native Go fuzzing (`csv_fuzz_test.go`) to continuously test parser stability against malformed or random inputs. Note that in some Windows environments, Go's fuzzing subprocess cleanup can encounter file locking ("Access is denied" when deleting temporary executables), which is an OS/environment issue rather than a code defect.
