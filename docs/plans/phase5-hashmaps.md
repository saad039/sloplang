# Phase 5: Hashmaps Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add hashmap support — arrays with string keys — including declaration (`data{k1, k2} = [v1, v2]`), literal key access (`map@name`), dynamic key access (`map@$var`), key set (`map@name = val`), keys prefix (`##`), and values prefix (`@@`).

**Architecture:** Hashmaps are `SlopValue` with the existing `Keys []string` field populated parallel to `Elements`. The `@` operator already exists from Phase 4 for numeric indexing — extend it to also handle bare-word key access and `$`-prefixed dynamic key access. Add 2 new lexer tokens (`##`, `@@`), 1 new token (`$`), 3 new AST nodes, 5 new runtime functions. The parser distinguishes numeric index (`arr@0`), literal key (`map@name`), and dynamic key (`map@$var`) at the postfix parsing level.

**Tech Stack:** Go, `go/ast`, existing sloplang transpiler pipeline

---

## Design Decisions

### How `@` handles three modes

Currently `@` parses a postfix primary expression as the index. For hashmaps:

1. `arr@0` — numeric index → already works via `IndexExpr` → `sloprt.Index()`
2. `map@name` — the token after `@` is an IDENT, but NOT a variable reference — it's a literal string key. This is a `KeyAccessExpr` → `sloprt.IndexKey(sv, "name")`
3. `map@$var` — the token after `@` is `$` followed by an IDENT — evaluate the variable as a string key. This is a `DynKeyAccessExpr` → `sloprt.IndexKey(sv, var)` (where `var` is a `*SlopValue` whose first element is the string key)

**Parser strategy:** In `parsePostfix()`, after consuming `@`:
- If next is `TOKEN_DOLLAR` + `TOKEN_IDENT`: dynamic key → `DynKeyAccessExpr`
- If next is `TOKEN_IDENT` (and NOT followed by `(`): literal key → `KeyAccessExpr`
- Otherwise: numeric index → `IndexExpr` (existing)

### How hashmap declaration works

`person{name, age} = ["bob", [30]]` — the `{k1, k2}` after the identifier declares the keys. This becomes a `HashDeclStmt` that lowers to `sloprt.MapFromKeysValues([]string{"name", "age"}, val)`.

### Key set (`map@name = val` and `map@$var = val`)

At the statement level, `parseStatement()` already handles `ident @ expr = val` as `IndexSetStmt`. Extend this to also handle:
- `ident @ name = val` → `KeySetStmt` (literal key)
- `ident @ $ var = val` → `DynKeySetStmt` (dynamic key)

Both lower to `sloprt.IndexKeySet(sv, key, val)`.

---

## Task 1: Lexer — new tokens

**Files:** `sloplang/pkg/lexer/token.go`, `sloplang/pkg/lexer/lexer.go`, `sloplang/pkg/lexer/lexer_test.go`

**New token types:**
- `TOKEN_DOUBLE_HASH` (`##`)
- `TOKEN_DOUBLE_AT` (`@@`)
- `TOKEN_DOLLAR` (`$`)

**Disambiguation:**
- `#` vs `##`: peek next — if `#` emit `TOKEN_DOUBLE_HASH`, else emit `TOKEN_HASH`
- `@` vs `@@`: peek next — if `@` emit `TOKEN_DOUBLE_AT`, else emit `TOKEN_AT`
- `$` is always `TOKEN_DOLLAR` (single char, no ambiguity)

**Tests:**
- Tokenize `## @@ $` — verify each type
- Disambiguation: `# ## @ @@ $` all produce correct tokens
- Tokenize `##person` → DOUBLE_HASH, IDENT
- Tokenize `@@person` → DOUBLE_AT, IDENT
- Tokenize `map@$var` → IDENT, AT, DOLLAR, IDENT
- Tokenize `person{name, age}` → IDENT, LBRACE, IDENT, COMMA, IDENT, RBRACE

---

## Task 2: AST — new node types

**Files:** `sloplang/pkg/parser/ast.go`

**New expression nodes:**

- `KeyAccessExpr` — fields: `Object Expr`, `Key string` (covers `map@name` — literal key access)
- `DynKeyAccessExpr` — fields: `Object Expr`, `KeyVar Expr` (covers `map@$var` — dynamic key)

**New statement nodes:**

- `HashDeclStmt` — fields: `Name string`, `Keys []string`, `Value Expr` (covers `person{name, age} = [...]`)
- `KeySetStmt` — fields: `Object Expr`, `Key string`, `Value Expr` (covers `map@name = val`)
- `DynKeySetStmt` — fields: `Object Expr`, `KeyVar Expr`, `Value Expr` (covers `map@$var = val`)

**Reuse existing nodes for:**
- `##map` → `UnaryExpr{Op: "##"}` (prefix)
- `@@map` → `UnaryExpr{Op: "@@"}` (prefix)

---

## Task 3: Parser — hashmap parsing

**Files:** `sloplang/pkg/parser/parser.go`, `sloplang/pkg/parser/parser_test.go`

### Expression parsing changes

**`parsePostfix()`:** After consuming `@`, before calling `parsePostfixPrimary()`:
- If `TOKEN_DOLLAR` followed by `TOKEN_IDENT`: consume `$`, consume ident, wrap in `DynKeyAccessExpr{Object: expr, KeyVar: &Identifier{Name: ident}}`
- If `TOKEN_IDENT` and peek is NOT `TOKEN_LPAREN` (to avoid treating function calls as keys): consume ident, wrap in `KeyAccessExpr{Object: expr, Key: ident}`
- Otherwise: fall through to existing numeric index via `parsePostfixPrimary()` → `IndexExpr`

**`parseUnary()`:** Add `TOKEN_DOUBLE_HASH` and `TOKEN_DOUBLE_AT` as prefix unary operators. Emit `UnaryExpr{Op: "##"}` and `UnaryExpr{Op: "@@"}`.

### Statement parsing changes

**`parseStatement()` — ident followed by `{`**: This is hashmap declaration `name{k1, k2} = [v1, v2]`. In the `TOKEN_IDENT` case, add check: if peek is `TOKEN_LBRACE`, parse as `HashDeclStmt`.

**`parseHashDeclStmt()`:** consume ident (name), consume `{`, collect key idents separated by commas (or empty for `counts{} = []`), consume `}`, consume `=`, parse expression as value.

**`parseStatement()` — ident followed by `@`**: Extend the existing index-set lookahead to also detect:
- `ident @ name = val` → `KeySetStmt` (literal key set)
- `ident @ $ var = val` → `DynKeySetStmt` (dynamic key set)
- `ident @ number = val` → `IndexSetStmt` (existing numeric index set)

Strategy: after saving position, consume ident, consume `@`:
1. If `TOKEN_DOLLAR` + `TOKEN_IDENT` + `TOKEN_ASSIGN`: it's `DynKeySetStmt`
2. If `TOKEN_IDENT` (not followed by `(`) + `TOKEN_ASSIGN`: it's `KeySetStmt`
3. If numeric/expr + `TOKEN_ASSIGN`: it's `IndexSetStmt` (existing)
4. Otherwise: restore, fall through to expression

### Statement parsing changes — prefix `##` and `@@`

Add `TOKEN_DOUBLE_HASH` and `TOKEN_DOUBLE_AT` to the existing `TOKEN_RSHIFT, TOKEN_HASH, TOKEN_TILDE` case in `parseStatement()` that handles prefix operators as expression statements.

**Tests:**
- Parse `person{name, age} = ["bob", [30]]` → `HashDeclStmt`
- Parse `counts{} = []` → `HashDeclStmt` with empty keys
- Parse `map@name` → `KeyAccessExpr`
- Parse `map@$var` → `DynKeyAccessExpr`
- Parse `map@name = val` → `KeySetStmt`
- Parse `map@$var = val` → `DynKeySetStmt`
- Parse `##map` → `UnaryExpr{Op: "##"}`
- Parse `@@map` → `UnaryExpr{Op: "@@"}`
- Parse `for k in ##map { ... }` → `ForInStmt` with `UnaryExpr{Op: "##"}` as iterable

---

## Task 4: Runtime — hashmap functions

**Files:** `sloplang/pkg/runtime/ops.go`, `sloplang/pkg/runtime/ops_test.go`

### New functions

**`MapFromKeysValues(keys []string, vals *SlopValue) *SlopValue`**
- Creates a new SlopValue with `Keys = keys` and `Elements` from `vals.Elements`.
- If `len(keys) == 0`, returns a SlopValue with empty Keys and Elements (empty hashmap).
- If `len(keys) != len(vals.Elements)`, panic.
- The Elements are taken from vals — if an element is `*SlopValue`, use directly; if raw, use directly.

**`IndexKey(sv *SlopValue, key *SlopValue) *SlopValue`**
- `key` must be single-element string. Finds the key in `sv.Keys`, returns the corresponding element wrapped as `*SlopValue`.
- If key not found, panic with "key not found: <key>".

**`IndexKeyStr(sv *SlopValue, key string) *SlopValue`**
- Same as `IndexKey` but takes a raw Go string. Used for literal key access (`map@name`).

**`IndexKeySet(sv *SlopValue, key *SlopValue, val *SlopValue) *SlopValue`**
- `key` must be single-element string. If key exists in `sv.Keys`, update the corresponding element. If key does NOT exist, append the key and element (grow the hashmap). Returns `sv` (mutated).

**`IndexKeySetStr(sv *SlopValue, key string, val *SlopValue) *SlopValue`**
- Same as `IndexKeySet` but takes a raw Go string. Used for literal key set (`map@name = val`).

**`MapKeys(sv *SlopValue) *SlopValue`**
- Returns a new SlopValue where each element is a string from `sv.Keys`.
- If `sv.Keys` is nil, return empty SlopValue.

**`MapValues(sv *SlopValue) *SlopValue`**
- Returns a new SlopValue with the same Elements as `sv` but no Keys.
- If `sv.Keys` is nil, return empty SlopValue.

### Tests
- `MapFromKeysValues(["name","age"], ["bob",[30]])` → SlopValue with Keys and Elements
- `MapFromKeysValues([], [])` → empty hashmap
- `MapFromKeysValues` length mismatch → panic
- `IndexKeyStr(person, "name")` → `"bob"`
- `IndexKeyStr` key not found → panic
- `IndexKey` with SlopValue key → works
- `IndexKeySetStr` existing key → updates element
- `IndexKeySetStr` new key → appends key+element
- `IndexKeySet` with SlopValue key → works
- `MapKeys(person)` → `["name", "age"]`
- `MapKeys` on plain array (nil Keys) → empty
- `MapValues(person)` → `["bob", [30]]`
- `MapValues` on plain array → empty

---

## Task 5: Codegen — lower hashmap operations

**Files:** `sloplang/pkg/codegen/codegen.go`, `sloplang/pkg/codegen/codegen_test.go`

### Statement lowering

**`HashDeclStmt`** → Two statements:
1. `name := sloprt.MapFromKeysValues([]string{"k1", "k2"}, loweredValue)`
2. `_ = name` (suppress unused)

The keys need to be emitted as a Go `[]string{...}` composite literal. Build an `ast.CompositeLit` with type `[]string` and elements as string BasicLits.

**`KeySetStmt`** → `sloprt.IndexKeySetStr(object, "key", value)` as `*ast.ExprStmt`

**`DynKeySetStmt`** → `sloprt.IndexKeySet(object, keyVar, value)` as `*ast.ExprStmt`

### Expression lowering

**`KeyAccessExpr`** → `callSloprt("IndexKeyStr", g.lowerExpr(e.Object), &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(e.Key)})`

**`DynKeyAccessExpr`** → `callSloprt("IndexKey", g.lowerExpr(e.Object), g.lowerExpr(e.KeyVar))`

**Unary ops — extend switch:**
- `"##"` → `callSloprt("MapKeys", g.lowerExpr(e.Operand))`
- `"@@"` → `callSloprt("MapValues", g.lowerExpr(e.Operand))`

### Tests
- Verify `person{name, age} = [...]` produces `sloprt.MapFromKeysValues` with `[]string{...}`
- Verify `map@name` produces `sloprt.IndexKeyStr(map, "name")`
- Verify `map@$var` produces `sloprt.IndexKey(map, var)`
- Verify `map@name = val` produces `sloprt.IndexKeySetStr`
- Verify `map@$var = val` produces `sloprt.IndexKeySet`
- Verify `##map` produces `sloprt.MapKeys`
- Verify `@@map` produces `sloprt.MapValues`

---

## Task 6: E2E tests — 40+ tests

**Files:** `sloplang/pkg/codegen/codegen_e2e_test.go`

### Hashmap declaration (5 tests)
- Declare with 2 keys: `person{name, age} = ["bob", [30]]`
- Declare with 1 key
- Declare empty hashmap: `counts{} = []`
- Declare with 3+ keys
- Declaration mismatch keys vs values → panic

### Key access — literal (5 tests)
- Access string value: `person@name` → `"bob"`
- Access nested SlopValue: `person@age` → `[30]`
- Access non-existent key → panic
- Access after set
- Multiple accesses

### Key access — dynamic (4 tests)
- Dynamic access: `which = "name"` then `person@$which` → `"bob"`
- Dynamic access with different variable
- Dynamic access in loop: `for k in ##map { |> str(map@$k) }`
- Dynamic access with computed key

### Key set — literal (4 tests)
- Set existing key: `person@age = [31]`
- Add new key: `person@email = "bob@test.com"`
- Set then verify
- Set multiple times

### Key set — dynamic (3 tests)
- Dynamic set: `key = "age"` then `map@$key = [31]`
- Dynamic set new key
- Dynamic set in loop

### Keys prefix `##` (4 tests)
- `##person` → array of key strings
- `##` on empty hashmap → empty
- `##` in for-in: `for k in ##map { ... }`
- `##` length: `#(##map)`

### Values prefix `@@` (3 tests)
- `@@person` → array of values
- `@@` on empty hashmap → empty
- `@@` used in expression

### Combined / integration (6+ tests)
- maps.slop roadmap example (full)
- Hashmap + for-in iteration over keys
- Hashmap inside function
- Function returning hashmap
- Hashmap with array values, index into value
- Count pattern: empty hashmap, check contains, increment/set

### Edge cases (4+ tests)
- Key with underscore: `data{my_key} = ["val"]`
- Single key hashmap
- Reassign entire hashmap
- Use hashmap value in arithmetic: `person@age + [1]`

---

## Task 7: Roadmap e2e — maps.slop

**Files:** `sloplang/examples/maps.slop`

Create the example from the roadmap:
```
person{name, age} = ["bob", [30]]
|> person@name
|> str(person@age)

person@age = [31]
|> str(person@age)

person@email = "bob@test.com"
|> person@email

ks = ##person
|> str(ks)

vs = @@person
|> str(vs)
```

Expected output: `bob`, `30`, `31`, `bob@test.com`, `[name, age, email]`, `[bob, 31, bob@test.com]`.

Note: `person@name` is a single-element string SlopValue → `StdoutWrite` prints `bob`. `str(person@age)` where age is `[30]` → FormatValue returns `"30"` (single element) → StdoutWrite prints `30`. After `person@email = "bob@test.com"`, the email value is a single-element string SlopValue → printed as `bob@test.com`. Keys and values format as arrays.

Wait — the roadmap expected output says `[name, age, email]` and `[bob, [31], bob@test.com]`. Let me re-examine. `MapKeys` returns elements that are strings, so FormatValue on a 3-element SlopValue with string elements gives `[name, age, email]`. `MapValues` returns a SlopValue with elements: string `"bob"`, `*SlopValue{int64(31)}`, string `"bob@test.com"`. FormatValue: `[bob, 31, bob@test.com]`. The roadmap says `[bob, [31], bob@test.com]` — that would require `[31]` format. But FormatValue on `*SlopValue{int64(31)}` recursively calls FormatValue which returns `"31"` (single element). So actual output is `[bob, 31, bob@test.com]`, NOT `[bob, [31], bob@test.com]`.

Verify this passes as an e2e test with the corrected expected output.

---

## Task 8: Null value — lexer, parser, AST

**Files:** `sloplang/pkg/lexer/token.go`, `sloplang/pkg/lexer/lexer.go`, `sloplang/pkg/parser/ast.go`, `sloplang/pkg/parser/parser.go`, `sloplang/pkg/lexer/lexer_test.go`, `sloplang/pkg/parser/parser_test.go`

### Design

`null` is a keyword that represents an absent/placeholder value. It is strict: panics on arithmetic, truthiness, logical ops, ordered comparisons, and for-in iteration. Only `==`/`!=` and formatting (`str`/`|>`) work with it.

**Representation:** Go type `SlopNull struct{}` with singleton stored in `SlopValue.Elements` alongside `int64`, `uint64`, `float64`, `string`, `*SlopValue`.

**Semantics:**

| Operation | Behavior |
|-----------|----------|
| `x = null` | Forward declaration, assigns `[SlopNull{}]` |
| `[null, null, null]` | Array with 3 null elements |
| `null == null` | `[1]` (truthy) |
| `null != [5]` | `[1]` (truthy) |
| `null + [1]` | **panic** |
| `-null` | **panic** |
| `if null { }` | **panic** — null is neither truthy nor falsy |
| `!null` | **panic** |
| `&&` / `||` with null | **panic** |
| `<`, `>`, `<=`, `>=` with null | **panic** |
| `str(null)` | `"null"` |
| `|> null` | prints `null` |
| `#[null, null]` | `[2]` — length counts null elements normally |
| `[1, null] ?? null` | `[1]` — contains finds null via deepEqual |
| `for x in null` | **panic** — cannot iterate null |

### Lexer changes
- Add `TOKEN_NULL` to token types
- Add `"null"` to keywords map → `TOKEN_NULL`

### Parser changes
- Add `NullLiteral{}` AST node (implements `Expr`, no fields)
- `parsePrimary()`: `TOKEN_NULL` → `NullLiteral{}`

### Tests
- Tokenize `null` → `TOKEN_NULL`
- Tokenize `null` in context: `x = null` → IDENT, ASSIGN, NULL
- Parse `null` → `NullLiteral{}`
- Parse `[null, null]` → `ArrayLiteral` with 2 `NullLiteral` elements
- Parse `x = null` → `AssignStmt` with `NullLiteral` value

---

## Task 9: Null value — runtime

**Files:** `sloplang/pkg/runtime/slop_value.go`, `sloplang/pkg/runtime/ops.go`, `sloplang/pkg/runtime/ops_test.go`

### Changes
- Add `type SlopNull struct{}` in `slop_value.go`
- `FormatValue`: `case SlopNull` → `"null"`
- `IsTruthy`: `SlopNull{}` element → panic("cannot use null as boolean")
- `Eq`: both `SlopNull{}` → equal; one null one non-null → not equal
- `Neq`: inverse of Eq
- `Lt`/`Gt`/`Lte`/`Gte`: null element → panic
- All arithmetic (`Add`/`Sub`/`Mul`/`Div`/`Mod`/`Pow`): null element → panic
- `Negate`: null element → panic
- `deepEqual`: `SlopNull{}` matches `SlopNull{}`
- `Iterate`: SlopValue with null element → panic

### Tests
- `FormatValue` on `[SlopNull{}]` → `"null"`
- `FormatValue` on `[1, SlopNull{}, "hi"]` → `[1, null, hi]`
- `IsTruthy` on null → panic
- `Eq(null, null)` → truthy
- `Eq(null, [1])` → falsy
- `Neq(null, [5])` → truthy
- `Lt` with null → panic
- `Add` with null → panic
- `Negate` null → panic
- `deepEqual(SlopNull{}, SlopNull{})` → true
- `Contains([1, null], null)` → truthy

---

## Task 10: Null value — codegen + E2E tests

**Files:** `sloplang/pkg/codegen/codegen.go`, `sloplang/pkg/codegen/codegen_test.go`, `sloplang/pkg/codegen/codegen_e2e_test.go`

### Codegen changes
- `NullLiteral` lowers to `sloprt.NewSlopValue(sloprt.SlopNull{})`

### E2E tests (15+ tests)
- `x = null` then `|> str(x)` → `null`
- `|> null` → `null`
- `x = [null, null, null]` then `|> str(x)` → `[null, null, null]`
- `null == null` → truthy, print "yes"
- `null != [5]` → truthy, print "yes"
- `[1] == null` → falsy, print "no"
- `#[null, null]` → `2`
- `[1, null] ?? null` → truthy
- `null + [1]` → panic
- `-null` → panic
- `if null { }` → panic
- `!null` → panic
- `null > [1]` → panic
- `for x in null` → panic
- Null in hashmap value: `m{a} = [null]`, `|> str(m@a)` → `null`
- Forward declaration: `x = null`, then `x = [42]`, `|> str(x)` → `42`

---

## Task 11: Comprehensive examples

**Files:** `sloplang/examples/math.slop`, `sloplang/examples/fns.slop`, `sloplang/examples/arrays.slop`, `sloplang/examples/maps.slop`, `sloplang/examples/null.slop`

### `math.slop` — expand to show full picture
- int64, uint64, float64 literals
- All 6 arithmetic ops (+, -, *, /, %, **)
- Unary negation on multi-element arrays
- All 6 comparisons
- Logical ops (&&, ||, !)
- true/false keywords

### `fns.slop` — expand to show full control flow
- Function declaration and return
- if/else branching
- for-in loop
- Infinite for loop with break
- Multi-assign (a, b = expr)
- Nested function calls

### `arrays.slop` — expand to show every array operator
- All 11 ops: @, #, <<, >>, ~@, ::, ++, --, ~, ??
- Index-set (arr@(i) = val)
- Parenthesized index with variable
- Chained operations

### `maps.slop` — expand to show all hashmap features
- Declaration (multi-key, single-key, empty)
- Literal key access/set
- Dynamic key access/set ($var)
- ## keys, @@ values
- Numeric index on hashmap (map@(0))
- Iteration over ##map

### `null.slop` (new) — null value examples
- Forward declaration with null
- Null in arrays
- null == null, null != [5]
- str(null), |> null
- #[null, null]
- [1, null] ?? null

---

## Task 12: Final verification + update docs

- Run `go test ./...` — all tests pass
- Run `go vet ./...` — clean
- Run `go fmt ./...` — clean
- Update `docs/architecture.md` with null value + updated examples
- Update `phase5-hashmaps.json` — flip new entries to `true`
- Commit
