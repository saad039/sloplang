# Phase 10: New Builtins + Nested Push — Design Spec

## Goal

Add 4 new builtin functions (`to_chars`, `to_int`, `to_float`, `fmt_float`), a new operator (`<<<` nested push), and document pointer-sharing semantics for mutating operators.

## Architecture

Five independent features plus one documentation task. All builtins follow the existing pattern: runtime function in Go, entry in the codegen builtins map, no parser changes needed. The `<<<` operator requires changes across lexer, parser, codegen, and runtime. Documentation updates cover the PRD, book, and appendices.

No changes to `isDualReturn()` are needed — none of the new builtins use dual-return semantics.

## Features

### 1. `to_chars(expr)` — String to character array

**Syntax:** `to_chars("hello")`, `to_chars(arr@0)`, `to_chars(myvar)`

**Semantics:**
- Accepts any expression that evaluates to a single-element string.
- Returns a `*SlopValue` where each element is a single-character string.
- Uses Go's `[]rune` iteration for correct Unicode handling.
- Empty string returns empty array `[]`.
- Panics if argument is not a single-element string.

**Examples:**
```
chars = to_chars("hello")
|> str(chars)              // [h, e, l, l, o]
|> "\n"

to_chars("")               // []
to_chars([42])             // panics: sloplang: to_chars requires a string argument
```

**Error model:** Panic on non-string input. Not dual-return.

**Implementation:**
- Runtime: `ToChars(sv *SlopValue) *SlopValue` in `io.go`
- Codegen: Add `"to_chars": "ToChars"` to builtins map
- Parser: No changes (already handles `name(expr)` as `CallExpr`)

---

### 2. `to_int(expr)` — Convert to integer

**Syntax:** `to_int([3.14])`, `to_int([42])`, `to_int("5")`

**Semantics:**
- Accepts a single-element array containing int64, float64, uint64, or string.
- Returns a single-element int64 array.
- float64 → int64: truncates toward zero (`to_int([3.7])` → `[3]`, `to_int([-2.9])` → `[-2]`).
- uint64 → int64: direct cast, panics if value > MaxInt64 with `"sloplang: to_int: uint64 value exceeds MaxInt64"`.
- string → int64: parses with `strconv.ParseInt`, panics if invalid.
- int64 → int64: no-op, returns as-is.
- Panics on null, multi-element arrays, or unparseable strings.

**Examples:**
```
|> str(to_int([3.14]))     // [3]
|> "\n"
|> str(to_int([-2.9]))     // [-2]
|> "\n"
|> str(to_int([42]))       // [42]
|> "\n"
|> str(to_int("5"))        // [5]
|> "\n"
to_int("abc")              // panics: sloplang: to_int: cannot convert string "abc" to int
to_int([null])             // panics: sloplang: to_int: cannot convert null to int
```

**Error model:** Panic on failure. Not dual-return.

**Implementation:**
- Runtime: `ToInt(sv *SlopValue) *SlopValue` in `io.go`
- Codegen: Add `"to_int": "ToInt"` to builtins map
- Parser: No changes

---

### 3. `to_float(expr)` — Convert to float

**Syntax:** `to_float([42])`, `to_float([3.14])`, `to_float("2.5")`

**Semantics:**
- Accepts a single-element array containing int64, float64, uint64, or string.
- Returns a single-element float64 array.
- int64 → float64: direct cast.
- uint64 → float64: direct cast. Note: precision loss for large uint64 values (> 2^53) is acceptable and silent.
- string → float64: parses with `strconv.ParseFloat`, panics if invalid.
- float64 → float64: no-op, returns as-is.
- Panics on null, multi-element arrays, or unparseable strings.

**Examples:**
```
|> str(to_float([42]))     // [42]  (stored as float64 internally)
|> "\n"
|> str(to_float("2.5"))   // [2.5]
|> "\n"
to_float("abc")            // panics: sloplang: to_float: cannot convert string "abc" to float
to_float([null])            // panics: sloplang: to_float: cannot convert null to float
```

**Error model:** Panic on failure. Not dual-return.

**Implementation:**
- Runtime: `ToFloat(sv *SlopValue) *SlopValue` in `io.go`
- Codegen: Add `"to_float": "ToFloat"` to builtins map
- Parser: No changes

---

### 4. `fmt_float(val, decimals)` — Float formatting with fixed decimals

**Syntax:** `fmt_float([3.14159], [2])` → `"3.14"`

**Semantics:**
- Takes two arguments: a single-element numeric array and a single-element non-negative integer array (decimal places).
- Returns a string (single-element string `*SlopValue`), not a numeric array.
- int64 input is promoted to float64 for formatting.
- Uses Go's `fmt.Sprintf("%.Nf", ...)` where N is the decimal count.
- Panics if first arg is not numeric or second arg is not a non-negative integer.

**Examples:**
```
|> fmt_float([3.14159], [2])   // 3.14
|> "\n"
|> fmt_float([42], [3])        // 42.000
|> "\n"
|> fmt_float([1.0], [0])       // 1
|> "\n"
|> fmt_float([0.1], [5])       // 0.10000
|> "\n"
fmt_float("abc", [2])          // panics: sloplang: fmt_float: first argument must be numeric
fmt_float([3.14], [-1])        // panics: sloplang: fmt_float: second argument must be non-negative integer
```

**Error model:** Panic on failure. Not dual-return.

**Implementation:**
- Runtime: `FmtFloat(val, decimals *SlopValue) *SlopValue` in `io.go`
- Codegen: Add `"fmt_float": "FmtFloat"` to builtins map (2-arg builtin, existing CallExpr handling works)
- Parser: No changes

---

### 5. `<<<` — Nested push operator

**Syntax:** `arr <<< expr`

**Semantics:**
- Appends the right-hand side as a single nested element, never spreading.
- Mutates `arr` in place (same as `<<`).
- Everything goes as nested — no unwrapping, no special-casing single elements.
- Only works with an identifier on the left side (same constraint as `<<`).

**Comparison with `<<`:**

| Operation | `<<` (spread) | `<<<` (nest) |
|-----------|--------------|--------------|
| `[1,2] _ [3,4]` | `[1, 2, 3, 4]` | `[1, 2, [3, 4]]` |
| `[1,2] _ [42]` | `[1, 2, 42]` | `[1, 2, [42]]` |
| `[1,2] _ "hi"` | `[1, 2, hi]` | `[1, 2, [hi]]` |
| `[] _ [1]` | `[1]` | `[[1]]` |

**Full example:**
```
arr = [1, 2]

arr <<< [3, 4]
|> str(arr)          // [1, 2, [3, 4]]
|> "\n"

arr <<< [42]
|> str(arr)          // [1, 2, [3, 4], [42]]
|> "\n"

arr <<< "hello"
|> str(arr)          // [1, 2, [3, 4], [42], [hello]]
|> "\n"

arr <<< ["saad", "is"]
|> str(arr)          // [1, 2, [3, 4], [42], [hello], [saad, is]]
|> "\n"
```

**Implementation:**

- **Lexer:** Add `TOKEN_NEST_PUSH` (`<<<`). In the `<` disambiguation chain, modify the existing `<<` branch. Currently, after detecting `peekChar() == '<'`, the code does two `readChar()` calls and emits `TOKEN_LSHIFT`. Change this to: after the two `readChar()` calls, check `l.ch` (NOT `peekChar()` — the second `readChar` already advanced past the second `<`). If `l.ch == '<'`, consume it with one more `readChar()` and emit `TOKEN_NEST_PUSH`; otherwise emit `TOKEN_LSHIFT` as before. The `<<<` token is greedy — `arr<<<expr` is always nested push. If `<< <expr>` is intended (e.g., `arr << <|`), a space is required between `<<` and `<`.

- **Parser:** Add `NestPushStmt{Object Expr, Value Expr}` AST node. In `parseStatement()`, when current token is `TOKEN_IDENT` and peek is `TOKEN_NEST_PUSH`, parse like `PushStmt` but produce `NestPushStmt`.

- **Codegen:** Lower `NestPushStmt` to `NestPush(obj, val)` runtime call (same pattern as `PushStmt` → `Push`).

- **Runtime:** `NestPush(sv, val *SlopValue) *SlopValue` in `ops.go`:
  ```go
  func NestPush(sv *SlopValue, val *SlopValue) *SlopValue {
      sv.Elements = append(sv.Elements, val)
      return sv
  }
  ```

---

### 6. Document pointer-sharing semantics for mutating operators

**Problem:** The book currently does not document that extracting a nested array via `@` or `$` gives a reference to the same `*SlopValue`. Mutations through `<<`, `>>`, `~@`, or `<<<` on the extracted value also affect the original.

**Example of undocumented behavior:**
```
matrix = [[1, 2], [3, 4]]
row = matrix@0
row << [99]
// matrix@0 is now [1, 2, 99] — mutation through row affected matrix
```

**Documentation updates needed:**
- **Book ch06-arrays.md:** Add a section on pointer-sharing / aliasing for mutating operators.
- **Book ch12-transpiler.md:** Correct the statement at line 315 ("arrays are values and mutations are returned, not applied in place behind a reference") to accurately describe the runtime model.
- **Appendix A:** Add `<<<` to the operator reference table.
- **Appendix B:** Add `to_chars`, `to_int`, `to_float`, `fmt_float` to the builtins reference.
- **PRD (docs/PRD.md):** Add `<<<` to the operator table, add new builtins to the builtins section.

---

## Testing Strategy

Each feature gets:
- **Runtime unit tests** in `ops_test.go` or `io_test.go` — test the Go function directly.
- **E2E tests** in codegen — full pipeline (source → lex → parse → codegen → compile → run → verify output).
- **Adversarial/edge-case tests** — invalid inputs, panics, boundary conditions.

**Test categories per feature:**

| Feature | Happy path | Edge cases | Panic cases |
|---------|-----------|------------|-------------|
| `to_chars` | ASCII, Unicode, variable input | empty string, single char | non-string input, multi-element array, null |
| `to_int` | float→int, string→int, int→int | MaxInt64, MinInt64, zero, negative float | null, multi-element, invalid string, uint64 overflow |
| `to_float` | int→float, string→float, float→float | very large ints, zero | null, multi-element, invalid string |
| `fmt_float` | various decimal counts, int input | 0 decimals, large decimals | non-numeric input, negative decimals, non-int decimals |
| `<<<` | multi-element, single-element, string, empty | nested `<<<`, mixed types | n/a (same panic surface as `<<`) |

**Existing test updates:** Any tests that rely on the absence of `<<<` or the builtin name table may need updates. The `<<<` lexer change must not break `<<` tokenization — verify with existing `<<` tests. Additionally, test greedy lexing edge cases: `<<<` followed by various characters, and `<< <|` (push + stdin read with space) to ensure correct disambiguation.

---

## Caveat: `arr@N <<<` not supported

Like `<<`, the `<<<` operator only works with an identifier on the left side. `arr@3 <<< [...]` is a parse error. Workaround:

```
inner = arr@3
inner <<< ["saad", "is"]
// arr@3 is also mutated (pointer sharing)
```

This is consistent with the existing `<<` limitation and is not in scope for this phase.
