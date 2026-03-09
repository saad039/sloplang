# Phase 4: Array Operators Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add full array manipulation operators — indexing, push, pop, remove, slice, concat, contains, unique, and length — so sloplang can work with arrays beyond arithmetic.

**Architecture:** Extend lexer with 8 new tokens (`@`, `#`, `<<`, `>>`, `~@`, `::`, `++`, `--`, `~`, `??`). Add 3 new AST nodes (IndexExpr, PushStmt, PopExpr). Add 10 runtime functions. Extend parser with postfix index parsing and new statement/expression forms. Codegen lowers all array ops to `sloprt.X()` calls.

**Tech Stack:** Go, `go/ast`, existing sloplang transpiler pipeline

---

## Operator Summary

| Sloplang | Name | Type | Runtime func | Example |
|----------|------|------|-------------|---------|
| `arr@0` | Index (numeric) | postfix expr | `Index(sv, idx)` | `arr@0` → element at index 0 |
| `arr@0 = val` | Index set | stmt | `IndexSet(sv, idx, val)` | `arr@0 = [99]` |
| `#arr` | Length | prefix expr | `Length(sv)` | `#arr` → `[4]` |
| `arr << [5]` | Push | stmt (mutates) | `Push(sv, val)` | appends to arr |
| `>>arr` | Pop | prefix expr (mutates) | `Pop(sv)` | removes+returns last |
| `arr ~@ 2` | Remove at | binary expr | `RemoveAt(sv, idx)` → `(removed, mutated)` | removes+returns element |
| `arr::1::3` | Slice | postfix expr | `Slice(sv, lo, hi)` | elements [1,3) |
| `[1,2] ++ [3,4]` | Concat | binary expr | `Concat(a, b)` | `[1,2,3,4]` |
| `arr -- [5]` | Remove value | binary expr | `Remove(sv, val)` | removes first occurrence |
| `~arr` | Unique | prefix expr | `Unique(sv)` | deduplicate |
| `arr ?? [5]` | Contains | binary expr | `Contains(sv, val)` | `[1]` or `[]` |

---

## Task 1: Lexer — new tokens

**Files:** `sloplang/pkg/lexer/token.go`, `sloplang/pkg/lexer/lexer.go`, `sloplang/pkg/lexer/lexer_test.go`

**New token types:**
- `TOKEN_AT` (`@`)
- `TOKEN_HASH` (`#`)
- `TOKEN_LSHIFT` (`<<`)
- `TOKEN_RSHIFT` (`>>`)
- `TOKEN_TILDE_AT` (`~@`)
- `TOKEN_DCOLON` (`::`)
- `TOKEN_CONCAT` (`++`)
- `TOKEN_REMOVE` (`--`)
- `TOKEN_TILDE` (`~`)
- `TOKEN_CONTAINS` (`??`)

**Disambiguation needed:**
- `~` vs `~@`: peek next char — if `@` emit `TOKEN_TILDE_AT`, else emit `TOKEN_TILDE`
- `+` vs `++`: peek next — if `+` emit `TOKEN_CONCAT`, else emit `TOKEN_PLUS`
- `-` vs `--`: peek next — if `-` emit `TOKEN_REMOVE`, else emit `TOKEN_MINUS`
- `<` vs `<<` vs `<=` vs `<-`: peek next — if `<` emit `TOKEN_LSHIFT`, if `=` emit `TOKEN_LTE`, if `-` emit `TOKEN_RETURN`, else emit `TOKEN_LT`
- `>` vs `>>` vs `>=`: peek next — if `>` emit `TOKEN_RSHIFT`, if `=` emit `TOKEN_GTE`, else emit `TOKEN_GT`
- `?` vs `??`: peek next — if `?` emit `TOKEN_CONTAINS`, else `TOKEN_ILLEGAL`
- `#` is always `TOKEN_HASH` (single char, no ambiguity)
- `@` is always `TOKEN_AT` (single char, no ambiguity)

**Tests:**
- Tokenize `@ # << >> ~@ :: ++ -- ~ ??` — verify each token type and literal
- Disambiguation: `< << <= <- > >> >= + ++ - -- ~ ~@` all produce correct tokens
- Verify `#arr` tokenizes to `TOKEN_HASH`, `TOKEN_IDENT`
- Verify `arr@0` tokenizes to `TOKEN_IDENT`, `TOKEN_AT`, `TOKEN_INT`
- Verify `arr << [5]` tokenizes correctly
- Verify `??` in expression context tokenizes correctly

---

## Task 2: AST — new node types

**Files:** `sloplang/pkg/parser/ast.go`

**New expression nodes:**

- `IndexExpr` — fields: `Object Expr`, `Index Expr` (covers `arr@0`)
- `PopExpr` — fields: `Object Expr` (covers `>>arr`, returns removed element)
- `SliceExpr` — fields: `Object Expr`, `Low Expr`, `High Expr` (covers `arr::1::3`)

**New statement nodes:**

- `PushStmt` — fields: `Object Expr`, `Value Expr` (covers `arr << [5]`)
- `IndexSetStmt` — fields: `Object Expr`, `Index Expr`, `Value Expr` (covers `arr@0 = val`)

**Reuse existing nodes for:**
- `#arr` → `UnaryExpr{Op: "#", Operand: ...}`
- `~arr` → `UnaryExpr{Op: "~", Operand: ...}`
- `arr ?? [5]` → `BinaryExpr{Op: "??", ...}`
- `[1,2] ++ [3,4]` → `BinaryExpr{Op: "++", ...}`
- `arr -- [5]` → `BinaryExpr{Op: "--", ...}`
- `arr ~@ 2` → `BinaryExpr{Op: "~@", ...}`

Each new node gets `exprNode()`/`stmtNode()` and `TokenLiteral()` methods following existing pattern.

---

## Task 3: Parser — array operator parsing

**Files:** `sloplang/pkg/parser/parser.go`, `sloplang/pkg/parser/parser_test.go`

### Precedence changes

Current precedence (low→high): `||` → `&&` → comparison → `+`/`-` → `*`/`/`/`%` → `**` → unary → call → primary

New: Insert `++`/`--`/`??`/`~@` at the `+`/`-` level (same precedence as add/sub). Insert `::` and `@` as postfix at the call level (tighter than binary ops). `#` and `~` are prefix unary. `<<` is a statement, not expression.

### Parser changes

**`parseUnary()`:** Add `TOKEN_HASH` and `TOKEN_TILDE` as prefix unary operators alongside `-` and `!`.

**`parseAddSub()`:** Extend to also consume `TOKEN_CONCAT` (`++`), `TOKEN_REMOVE` (`--`), `TOKEN_CONTAINS` (`??`), and `TOKEN_TILDE_AT` (`~@`).

**`parseCall()`:** After parsing a call or primary, check for postfix `@` (index) and `::` (slice). Loop: if `TOKEN_AT` follows, consume it, parse index expression, wrap in `IndexExpr`. If `TOKEN_DCOLON` follows, consume it, parse low, expect another `TOKEN_DCOLON`, parse high, wrap in `SliceExpr`.

**`>>` prefix (Pop):** In `parseUnary()`, if `TOKEN_RSHIFT` is seen, consume it, parse operand, return `PopExpr{Object: operand}`.

**`<<` (Push) — statement level:** In `parseStatement()`, after parsing an assignment-like statement with `TOKEN_IDENT`, check if next token is `TOKEN_LSHIFT`. Pattern: `ident << expr`. Parse as `PushStmt{Object: Identifier, Value: expr}`.

Actually, simpler: handle `<<` in `parseStatement()` after we see an ident followed by `TOKEN_LSHIFT`:
- In `parseStatement()` case `TOKEN_IDENT`: peek second token — if `TOKEN_LSHIFT`, parse as push statement.

**Index set (`arr@0 = val`):** In `parseStatement()`, when we see `TOKEN_IDENT` followed by `TOKEN_AT`, we need to distinguish reading (`x = arr@0`) from writing (`arr@0 = val`). Strategy: parse the identifier, check for `@`, then check for `=`.
- If `TOKEN_IDENT` + `TOKEN_AT` + index + `TOKEN_ASSIGN`: parse as `IndexSetStmt`
- If `TOKEN_IDENT` + `TOKEN_AT` (no `=` after): it's an expression (falls through to expression parsing)

For statement parsing, in `parseStatement()` when we see `TOKEN_IDENT`:
1. peek at second token:
   - `TOKEN_COMMA` → multi-assign
   - `TOKEN_ASSIGN` → assign
   - `TOKEN_LSHIFT` → push stmt
   - `TOKEN_AT` → could be index-set or expression. Need 3-token lookahead: `ident @ index =` → index set. Otherwise fall through to expression.

For the `TOKEN_AT` lookahead case: save position, try to parse `ident @ index`, if followed by `=` then it's index-set. Otherwise restore position and fall through to expression statement. Alternatively, parse `ident` as expression, then if the result is `IndexExpr` and next is `=`, convert to `IndexSetStmt`.

**Simplest approach for index set:** In `parseStatement()`, when we see `TOKEN_IDENT` + `TOKEN_AT`:
1. Save current position
2. Try parsing: consume ident, consume `@`, parse index expr
3. If next is `TOKEN_ASSIGN`: it's an `IndexSetStmt`. Consume `=`, parse value.
4. If next is NOT `=`: restore position and fall through to expression statement.

This requires a `save/restore` mechanism on the parser. Add `save()` and `restore(pos int)` methods.

**Tests:**
- Parse `arr@0` → `IndexExpr` with object=Identifier("arr"), index=NumberLiteral("0")
- Parse `arr@0 = [99]` → `IndexSetStmt`
- Parse `#arr` → `UnaryExpr{Op: "#"}`
- Parse `arr << [5]` → `PushStmt`
- Parse `>>arr` → `PopExpr`
- Parse `arr::1::3` → `SliceExpr`
- Parse `[1,2] ++ [3,4]` → `BinaryExpr{Op: "++"}`
- Parse `arr -- [5]` → `BinaryExpr{Op: "--"}`
- Parse `~arr` → `UnaryExpr{Op: "~"}`
- Parse `arr ?? [5]` → `BinaryExpr{Op: "??"}`
- Parse `arr ~@ 2` → `BinaryExpr{Op: "~@"}`
- Parse chained index: `matrix@0` (just single for now, nested arrays return `*SlopValue`)
- Parse slice followed by index: `arr::1::3` (SliceExpr)

---

## Task 4: Runtime — array operation functions

**Files:** `sloplang/pkg/runtime/ops.go`, `sloplang/pkg/runtime/ops_test.go`

### New functions

**`Index(sv *SlopValue, idx *SlopValue) *SlopValue`**
- `idx` must be single-element int64. Extract the int, index into `sv.Elements`.
- If element is `*SlopValue`, return it directly. Otherwise wrap in `NewSlopValue()`.
- Panic if index out of bounds.

**`IndexSet(sv *SlopValue, idx *SlopValue, val *SlopValue) *SlopValue`**
- Mutates `sv.Elements[idx]` to `val`. Returns `sv` (for chaining convenience).
- Panic if index out of bounds.

**`Length(sv *SlopValue) *SlopValue`**
- Returns `NewSlopValue(int64(len(sv.Elements)))`.

**`Push(sv *SlopValue, val *SlopValue) *SlopValue`**
- Appends all elements of `val` to `sv.Elements`. Returns `sv` (mutated).
- This is `sv.Elements = append(sv.Elements, val.Elements...)`.

**`Pop(sv *SlopValue) *SlopValue`**
- Removes last element, returns it wrapped as `*SlopValue`.
- Panic if empty.

**`RemoveAt(sv *SlopValue, idx *SlopValue) *SlopValue`**
- `idx` must be single-element int64. Removes element at that index from `sv.Elements`.
- Returns the removed element wrapped as `*SlopValue`.
- Mutates `sv`.
- Panic if index out of bounds.

**`Slice(sv *SlopValue, low *SlopValue, high *SlopValue) *SlopValue`**
- `low` and `high` must be single-element int64. Returns new `SlopValue` with `sv.Elements[lo:hi]`.
- Panic if out of bounds.

**`Concat(a *SlopValue, b *SlopValue) *SlopValue`**
- Returns new `SlopValue` with all elements of `a` followed by all elements of `b`.
- Does NOT mutate either.

**`Remove(sv *SlopValue, val *SlopValue) *SlopValue`**
- Removes first occurrence of `val.Elements[0]` from `sv.Elements`.
- Returns new `SlopValue` with the element removed.
- If not found, returns copy unchanged.

**`Contains(sv *SlopValue, val *SlopValue) *SlopValue`**
- `val` must be single-element. Checks if any element in `sv` equals `val.Elements[0]`.
- Returns `[1]` (truthy) or `[]` (falsy).
- Uses deep equality for `*SlopValue` elements.

**`Unique(sv *SlopValue) *SlopValue`**
- Returns new `SlopValue` with duplicate elements removed (keep first occurrence).
- Uses same equality check as Contains.

### Tests for each function

- `Index([10,20,30], [0])` → `[10]`
- `Index([10,20,30], [2])` → `[30]`
- `Index` out of bounds → panic
- `Index` on nested: `Index([[1,2],[3,4]], [0])` → `[1,2]`
- `IndexSet([10,20,30], [1], [99])` → sv now `[10,99,30]`
- `IndexSet` out of bounds → panic
- `Length([1,2,3])` → `[3]`
- `Length([])` → `[0]`
- `Push([1,2], [3])` → sv now `[1,2,3]`
- `Push([], [1])` → sv now `[1]`
- `Pop([1,2,3])` → returns `[3]`, sv now `[1,2]`
- `Pop([])` → panic
- `RemoveAt([10,20,30], [1])` → returns `[20]`, sv now `[10,30]`
- `RemoveAt` out of bounds → panic
- `Slice([10,20,30,40], [1], [3])` → `[20,30]`
- `Slice` out of bounds → panic
- `Concat([1,2], [3,4])` → `[1,2,3,4]`
- `Concat([], [1])` → `[1]`
- `Concat([1], [])` → `[1]`
- `Remove([1,2,3,2], [2])` → `[1,3,2]` (removes first only)
- `Remove([1,2,3], [5])` → `[1,2,3]` (not found, unchanged)
- `Contains([1,2,3], [2])` → `[1]`
- `Contains([1,2,3], [5])` → `[]`
- `Contains([], [1])` → `[]`
- `Unique([1,2,2,3,1])` → `[1,2,3]`
- `Unique([])` → `[]`
- `Unique(["a","b","a"])` → `["a","b"]`

---

## Task 5: Codegen — lower array operators

**Files:** `sloplang/pkg/codegen/codegen.go`, `sloplang/pkg/codegen/codegen_test.go`

### Statement lowering

**`PushStmt`** → `sloprt.Push(object, value)` as `*ast.ExprStmt`

**`IndexSetStmt`** → `sloprt.IndexSet(object, index, value)` as `*ast.ExprStmt`

### Expression lowering

**`IndexExpr`** → `sloprt.Index(object, index)` call

**`PopExpr`** → `sloprt.Pop(object)` call

**`SliceExpr`** → `sloprt.Slice(object, low, high)` call

**Binary ops — extend `opFunc` map:**
- `"++"` → `"Concat"`
- `"--"` → `"Remove"`
- `"??"` → `"Contains"`
- `"~@"` → `"RemoveAt"`

**Unary ops — extend unary handling:**
- `"#"` → `sloprt.Length(operand)`
- `"~"` → `sloprt.Unique(operand)`

### Tests
- Verify `arr@0` produces `sloprt.Index(arr, ...)` call
- Verify `#arr` produces `sloprt.Length(arr)` call
- Verify `arr << [5]` produces `sloprt.Push(arr, ...)` call
- Verify `>>arr` produces `sloprt.Pop(arr)` call
- Verify `arr::1::3` produces `sloprt.Slice(arr, ...)` call
- Verify `++` produces `sloprt.Concat(...)` call
- Verify `--` produces `sloprt.Remove(...)` call
- Verify `??` produces `sloprt.Contains(...)` call
- Verify `~` produces `sloprt.Unique(...)` call
- Verify `~@` produces `sloprt.RemoveAt(...)` call

---

## Task 6: E2E tests — 50+ tests

**Files:** `sloplang/pkg/codegen/codegen_e2e_test.go`

### Index (8 tests)
- Index first element: `arr@0`
- Index last element: `arr@2` on 3-element array
- Index on variable index: `i = [1]` then `arr@i` — **wait**: `arr@0` uses a literal int, but `arr@i` uses a variable. The `@` always takes the next token as index expression. So `arr@i` should work if the parser parses the expression after `@`. Actually looking at PRD: `arr@0` is numeric index, `map@name` is literal key, `map@$var` is dynamic key. For Phase 4 (arrays only, no hashmaps), we only need numeric index `arr@0` and `arr@$var` (dynamic). But `$` is Phase 5. For now, just `arr@expr` where expr is a number literal or variable.
- Index on nested array
- Index into result of function call
- Index set: `arr@1 = [99]`, verify change
- Index out of bounds → panic

### Length (4 tests)
- `#arr` on non-empty
- `#arr` on empty
- `#` on result of expression: `#(arr ++ [5])`
- `#` used in comparison: `if #arr == [3]`

### Push (4 tests)
- Push single element
- Push to empty array
- Push then verify length
- Push inside loop

### Pop (4 tests)
- Pop returns last element
- Pop modifies array
- Pop in assignment: `x = >>arr`
- Pop empty → panic

### RemoveAt (4 tests)
- Remove first element
- Remove middle element
- Remove last element
- RemoveAt out of bounds → panic

### Slice (6 tests)
- Basic slice: `arr::1::3`
- Slice from start: `arr::0::2`
- Slice to end: `arr::2::4` on 4-element array
- Full slice: `arr::0::4`
- Empty slice: `arr::2::2`
- Slice out of bounds → panic

### Concat (4 tests)
- Basic concat: `[1,2] ++ [3,4]`
- Concat with empty: `[] ++ [1]` and `[1] ++ []`
- Concat strings: `["a"] ++ ["b"]`
- Concat result used in expression

### Remove (3 tests)
- Remove existing value
- Remove non-existent value (no change)
- Remove from array with duplicates (first only)

### Contains (4 tests)
- Contains existing → `[1]`
- Contains non-existent → `[]`
- Contains in if condition
- Contains on empty array

### Unique (3 tests)
- Unique on array with duplicates
- Unique on empty
- Unique on already-unique array

### Combined / integration (6+ tests)
- `arrays.slop` roadmap example
- Push then iterate
- Filter pattern: for-in + if contains + concat
- Sort-like: repeated index + comparison
- Chained operations: slice then concat
- Function returning array, caller indexes into result

---

## Task 7: Roadmap e2e — arrays.slop

**Files:** `sloplang/examples/arrays.slop`

Create the example from the roadmap:
```
arr = [10, 20, 30, 40]
|> str(arr@0)
|> str(#arr)

arr << [50]
|> str(arr)

sub = arr::1::3
|> str(sub)

combined = [1, 2] ++ [3, 4]
|> str(combined)

has = arr ?? [20]
|> str(has)

uniq = ~[1, 2, 2, 3, 1]
|> str(uniq)
```

Expected output: `10`, `4`, `[10, 20, 30, 40, 50]`, `[20, 30]`, `[1, 2, 3, 4]`, `1`, `[1, 2, 3]`.

Note: `#arr` returns `[4]` → `str([4])` → `"4"` (single-element FormatValue). `arr ?? [20]` returns `[1]` → `str([1])` → `"1"`.

Verify this passes as an e2e test.

---

## Task 8: Final verification + flip passes

- Run `go test ./...` — all tests pass
- Run `go vet ./...` — clean
- Run `go fmt ./...` — clean
- Flip all `"passes": false` to `true` in `phase4-array-operators.json`
- Commit
