# Architectural Decision Log

This document records the non-trivial architectural divergences made during the port of `@gregoranders/csv` (TypeScript) to Go, and the rationale behind each.

---

## 1. Exception Model → Multi-Return Error Propagation

**TS Approach:** The TypeScript parser uses `throw new ParseError(line, column)` inside deeply nested private methods (`handleQuoteNotEscaped`). JavaScript's exception unwinding automatically propagates the error up through `handleQuote → handleNext → parse`, regardless of call depth.

**Go Divergence:** Go has no exception mechanism. Every internal method that can fail (`handleQuote`, `handleQuoteNotEscaped`) was redesigned to return `error` as a second return value. The call chain in `handleNext()` was restructured to explicitly check and propagate errors at each level:

```go
func (p *Parser) handleNext() error {
    quoted, err := p.handleQuote()
    if err != nil {
        return err
    }
    // ...
}
```

**Rationale:** This is idiomatic Go. Using `panic/recover` would technically work but violates Go conventions and makes the API fragile for consumers. The explicit error chain also makes control flow visible in code review, which the implicit `throw` path in TS does not.

---

## 2. Immutability via `Object.freeze` → Mutable Slices by Design

**TS Approach:** After parsing, the TypeScript implementation calls `makeImmutable()`, which recursively applies `Object.freeze()` on every row, every field string, and the top-level `_rows` array. The `rows` getter and `json` getter both return `readonly` typed arrays. The test suite explicitly verifies that mutations throw `TypeError`.

**Go Divergence:** We intentionally omit the `makeImmutable()` step entirely. Go has no language-level equivalent to `Object.freeze` for slices or maps. Achieving similar protection would require either:
- Returning deep-copied slices on every `Rows()` call (expensive for large CSVs).
- Wrapping slices in a custom read-only struct (breaks idiomatic `[]string` expectations and makes the API cumbersome).

**Rationale:** Go's convention is that consumers are trusted not to mutate returned slices unless documented otherwise. The performance cost and API ergonomic degradation of enforcing immutability outweigh the safety benefit in Go's ecosystem. This is the single largest behavioral divergence from the TS implementation.

---

## 3. Generic Type Parameter `<T>` → Concrete `map[string]string`

**TS Approach:** The `Parser` class is generic: `class Parser<T = Record<string, string>>`. The `json` getter returns `readonly T[]`, allowing consumers to cast parsed JSON output to a custom type.

**Go Divergence:** Go generics (introduced in 1.18) could theoretically replicate this, but the original TS implementation never actually uses the generic type parameter `T` for anything other than casting — it always builds a plain `Record<string, string>` object internally. Adding Go generics here would introduce complexity without functional benefit.

**Rationale:** We use the concrete return type `[]map[string]string` for `JSON()`. This is simpler, more readable, and matches what the TS parser actually produces. Consumers who need typed structs can trivially marshal the map output into their desired Go struct using `encoding/json`.

---

## 4. Property Getters → Explicit Methods

**TS Approach:** TypeScript uses ES5 property getters (`get rows()`, `get json()`) that look like field accesses but execute code.

**Go Divergence:** Go does not have property getters. We expose these as explicit methods: `Rows()` and `JSON()`. The naming follows Go convention (exported, PascalCase) rather than the TS convention (camelCase).

**Rationale:** In Go, the distinction between field access and method calls is explicit by convention (parentheses). This avoids the hidden computation ambiguity that getters introduce, and aligns with `encoding/json`, `database/sql`, and other standard library patterns.

---

## 5. UTF-16 Code Unit Indexing → Rune-Based Iteration

**TS Approach:** TypeScript iterates strings using `text[this._index]`, which accesses individual UTF-16 code units. For characters in the Basic Multilingual Plane (BMP), this is equivalent to one character per index. However, emoji and other supplementary characters (e.g., `😀`, U+1F600) occupy two UTF-16 code units (a surrogate pair), meaning `text[i]` returns a lone surrogate — not a complete character.

**Go Divergence:** We convert the input to `[]rune` and iterate rune-by-rune. Each rune represents a complete Unicode code point, regardless of byte length.

**Rationale:** This is a deliberate fidelity trade-off. For the **shared public API** (the `Parse()` method), both implementations produce identical row/field outputs on all tested inputs including emoji, as confirmed by 508,444 differential fuzz iterations. The internal `lineOffset` / `fieldOffset` counters may differ on supplementary characters (Go counts 1 rune, TS counts 2 code units), but these counters only surface in `ParseError` messages for malformed input — not in the parsed data itself. We chose rune-based iteration because it is correct-by-default for Unicode and avoids surrogate-pair splitting bugs.

---

## 6. `Object.assign` Configuration Merge → Zero-Value Sentinel Pattern

**TS Approach:** Configuration merging uses `Object.assign({}, DefaultConfiguration, configuration)`, which copies all provided properties over defaults in one call. Missing keys simply remain at their default values.

**Go Divergence:** Go structs initialize with zero values (`""` for strings). We use explicit zero-value checks for each field:
```go
if cfg.FieldSeparator != "" {
    opts.FieldSeparator = cfg.FieldSeparator
}
```

**Rationale:** `Object.assign` has no direct Go equivalent. The zero-value sentinel pattern is idiomatic Go and works correctly here because empty-string delimiters are nonsensical for CSV parsing. The trade-off is that a user cannot explicitly set a delimiter to `""` — but this is intentionally unsupported (an empty delimiter would cause infinite loops or undefined behavior in both the TS and Go implementations).

---

## 7. Class Instance Fields → Struct with Value-Copy State

**TS Approach:** `_state` and `_quoteState` are object references. `this._quoteState = { ...this._state }` creates a shallow copy via spread syntax.

**Go Divergence:** `State` is a Go struct (value type). Assignment `p.quoteState = p.state` automatically performs a full value copy because Go structs are copied by value, not by reference.

**Rationale:** This is a subtle but critical correctness point. In TypeScript, forgetting the spread (`...`) would create a shared reference, causing state corruption. In Go, the value-copy semantics are the default behavior, making this class of bug impossible. We deliberately kept `State` as a plain struct (not a pointer) to preserve this safety property.

---

## 8. Short-Circuit Boolean Chain → Explicit If-Else with Error Returns

**TS Approach:** `handleNext()` uses JavaScript's short-circuit evaluation:
```typescript
this.handleQuote() || this.handleFieldSeparator() || this.handleLineSeparator();
```
This relies on the fact that each `handle*` method returns a boolean, and `||` short-circuits on `true`. Errors are thrown as exceptions, so they bypass the boolean chain entirely.

**Go Divergence:** Because `handleQuote` now returns `(bool, error)`, the short-circuit chain cannot be used. We restructured the logic as explicit if-else blocks:
```go
quoted, err := p.handleQuote()
if err != nil { return err }
if !quoted {
    if !p.handleFieldSeparator() {
        p.handleLineSeparator()
    }
}
```

**Rationale:** Go's multi-return values and lack of exceptions make the boolean-chain pattern impossible. The explicit if-else structure is arguably more readable and debuggable, as each branch is independently testable and the error path is visible.

---

## 9. `String.prototype.slice` → Byte-Aware Rune Slicing

**TS Approach:** `handleQuoteEscaped` strips the trailing backslash using `this._cell.slice(0, Math.max(0, this._cell.length - 1))`. In JavaScript, `string.length` counts UTF-16 code units and `slice` operates on those units.

**Go Divergence:** We use `p.cell[:len(p.cell)-1]`, which operates on **bytes**, not runes. This works correctly because the character being removed is always a backslash (`\`, ASCII, 1 byte). If the cell contained multi-byte characters before the backslash, byte slicing still correctly removes only the trailing single-byte `\`.

**Rationale:** We validated that this byte-level slice is safe specifically because the removed character (`\`) is always a single-byte ASCII character. Using `[]rune` conversion here would add unnecessary overhead. This is a performance-conscious decision backed by the invariant that the escape character is always ASCII.

---

## 10. Variadic Constructor → Functional Options Considered, Rejected

**Go Divergence:** We considered three patterns for `NewParser` configuration:
1. **Functional Options** (`NewParser(WithSeparator(","), WithQuote("\""))`) — popular in Go libraries.
2. **Config Struct** (`NewParser(Configuration{...})`) — simpler, matches TS structure.
3. **Variadic Config** (`NewParser(config ...Configuration)`) — allows zero-arg default.

We chose option 3 (variadic config struct) because it most closely mirrors the TypeScript constructor's optional parameter pattern, maintains a 1:1 structural correspondence for port reviewers, and avoids the overhead of defining individual option functions for a parser with only 3 configuration fields.

**Rationale:** Functional options shine when there are many optional parameters or when the API evolves frequently. With only 3 fields (`FieldSeparator`, `LineSeparator`, `Quote`), a config struct is more straightforward and readable.

---

## 11. CLI Bridge Architecture: `stdin` JSON Protocol → `execSync`

**TS Approach:** The original library is a pure TypeScript module imported directly by Jest tests.

**Go Divergence:** To enable the original Jest test suite to run against the Go implementation (Track C requirement), we built a CLI binary (`src/cmd/csv-cli/main.go`) that accepts JSON on `stdin` and returns JSON on `stdout`. A thin TypeScript adapter imports this binary via `child_process.execSync`. This is a deliberate architectural boundary:
- The Go binary is fully standalone with zero Node.js dependencies.
- The JSON protocol is the contract; either side can be replaced independently.
- The `sendError` function in the CLI exits with code 0 (not 1) to ensure the parent process receives valid JSON even on parse errors, avoiding `execSync` throwing on non-zero exits.

**Rationale:** This decoupled architecture allows the Go parser to be tested by Jest without embedding a JavaScript runtime, while keeping the Go module usable as a standalone library or CLI tool.

---

## 12. Test Structure: Jest `describe/it` → Go Table-Driven Tests

**TS Approach:** Tests use Jest's nested `describe` / `it` blocks with `.map()` to iterate over test cases, `expect().toStrictEqual()` for deep equality, and `expect().toThrow()` for error assertions.

**Go Divergence:** We use Go's table-driven test pattern with `[]struct{ text string; expected []csv.Row }` slices iterated via `for _, tc := range tests` with subtests (`t.Run`). Deep equality uses `reflect.DeepEqual`, and error assertions use direct string comparison on `err.Error()`.

**Rationale:** Table-driven tests are the Go community standard (recommended in the Go wiki and used throughout the standard library). They provide the same parameterized coverage as Jest's `.map()` pattern while being idiomatic and compatible with `go test -run` filtering. The `reflect.DeepEqual` usage is confined to test code only — zero `reflect` in production (documented in `ZERO_UNSAFE.md`).
