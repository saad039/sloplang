# Sloplang Architecture

## Pipeline

```
.slop source → Lexer → Parser → Codegen → .go source + runtime → go build → binary
```

The transpiler reads a `.slop` file, tokenizes it, parses tokens into an AST, lowers the AST to `go/ast` nodes, formats them as Go source, and writes a `.go` file. The generated Go code imports `pkg/runtime` (aliased as `sloprt`) for all value operations.

## Directory Structure

```
sloplang/                          # project root
├── docs/                          # documentation
│   ├── PRD.md                     # language spec
│   ├── roadmap.md                 # implementation phases
│   ├── architecture.md            # this file
│   ├── patterns.md                # lessons learned
│   └── plans/                     # per-phase plan + JSON tracking
├── CLAUDE.md                      # project instructions
└── sloplang/                      # Go module root (go.mod lives here)
    ├── cmd/slop/main.go           # CLI entry point
    ├── pkg/
    │   ├── lexer/
    │   │   ├── token.go           # TokenType constants, Token struct, keywords map
    │   │   └── lexer.go           # Lexer: input → []Token
    │   ├── parser/
    │   │   ├── ast.go             # AST node types (Stmt, Expr interfaces)
    │   │   └── parser.go          # Parser: []Token → *Program (recursive descent)
    │   ├── codegen/
    │   │   └── codegen.go         # Generator: *Program → Go source (via go/ast)
    │   └── runtime/
    │       ├── slop_value.go      # SlopValue struct, NewSlopValue, IsTruthy, StdoutWrite, FormatValue
    │       ├── ops.go             # All operations: arithmetic, comparison, logical, array ops
    │       └── io.go              # I/O: StdinRead, FileRead, FileWrite, FileAppend, Split, ToNum
    ├── tests/
    │   └── programs/              # Phase 8 real programs (.slop) + test harness
    └── examples/                  # Example .slop programs
```

## Core Types

### SlopValue (`pkg/runtime/slop_value.go`)

The universal value type. Every sloplang value is a `SlopValue`:

```go
type SlopValue struct {
    Elements []any    // int64, uint64, float64, string, *SlopValue, or SlopNull
    Keys     []string // parallel to Elements for hashmaps; nil for plain arrays
}
```

- Strict booleans: only `[1]` is truthy, only `[]` is falsy; `[0]`, multi-element arrays, strings, and `SlopNull` all panic on truthiness check
- `NewSlopValue(elems ...any)` constructs values
- `FormatValue` renders: single-element string → raw string (e.g. `hello`), everything else → bracket notation (e.g. `[42]`, `[1, 2, 3]`, `[null]`, `[]`)
- `StdoutWrite` uses `fmt.Print` (no trailing newline) — explicit `"\n"` required in slop source
- `SlopNull struct{}` — sentinel type for null values. Panics on arithmetic, truthiness, ordered comparisons, iteration. Supports `==`/`!=` and formatting.

### Token (`pkg/lexer/token.go`)

```go
type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Col     int
}
```

Token types cover: literals (INT, UINT, FLOAT, STRING, IDENT), operators (arithmetic, comparison, logical, array), I/O (`<|`, `<.`, `.>`, `.>>`), delimiters, keywords (`fn`, `if`, `else`, `for`, `in`, `break`, `true`, `false`, `null`), and return (`<-`). `true` and `false` are reserved keywords that parse as `ArrayLiteral` producing `[1]` and `[]` respectively.

### AST Nodes (`pkg/parser/ast.go`)

Two interfaces: `Stmt` (statements) and `Expr` (expressions).

**Statements:** AssignStmt, StdoutWriteStmt, FnDeclStmt, IfStmt, ForInStmt, ForLoopStmt, BreakStmt, ReturnStmt, MultiAssignStmt, ExprStmt, PushStmt, NestPushStmt, IndexSetStmt, HashDeclStmt, KeySetStmt, DynAccessSetStmt, FileWriteStmt, FileAppendStmt

**Expressions:** ArrayLiteral, NumberLiteral, StringLiteral, NullLiteral, Identifier, BinaryExpr, UnaryExpr, CallExpr, IndexExpr, PopExpr, SliceExpr, KeyAccessExpr, DynAccessExpr, StdinReadExpr, FileReadExpr

## Lexer (`pkg/lexer/lexer.go`)

Single-pass, character-by-character tokenizer. Key design decisions:

- **Greedy multi-char matching:** `##`, `@@`, `<<<`, `<<`, `>>`, `++`, `--`, `~@`, `::`, `??`, `**`, `==`, `!=`, `<=`, `>=`, `<-`, `<|`, `<.`, `|>`, `||`, `&&`, `.>`, `.>>` are checked before their single-char prefixes. `<<<` (`TOKEN_NEST_PUSH`) is checked before `<<` to avoid greedy mismatch.
- **Keywords via map lookup:** `LookupIdent()` checks the `keywords` map, falling back to `TOKEN_IDENT`
- **Number disambiguation:** digits followed by `u` → UINT, digits with `.` → FLOAT, else INT
- **String escapes:** `\n`, `\t`, `\\`, `\"`
- **Comments:** `//` skips to EOL

## Parser (`pkg/parser/parser.go`)

Recursive descent with Pratt-style precedence for expressions.

### Precedence (low → high)

1. `||` (Or)
2. `&&` (And)
3. `==`, `!=`, `<`, `>`, `<=`, `>=` (Comparison)
4. `+`, `-`, `++`, `--`, `??`, `~@` (AddSub + array binary ops)
5. `*`, `/`, `%` (MulDivMod)
6. `**` (Power, right-associative)
7. Unary: `-`, `!`, `#`, `~`, `>>`, `##`, `@@` (prefix)
8. Call: `name(args...)`
9. Postfix: `$` (dynamic access), `@` (index / key access), `::` (slice)
10. Primary: literals, identifiers, `(expr)`

### Statement dispatch

`parseStatement()` checks the current token:
- `TOKEN_FN` → `parseFnDecl()`
- `TOKEN_IF` → `parseIfStmt()`
- `TOKEN_FOR` → `parseForStmt()` (dispatches to infinite loop or for-in)
- `TOKEN_BREAK` → `parseBreakStmt()`
- `TOKEN_RETURN` → `parseReturnStmt()`
- `TOKEN_PIPE_GT` → `parseStdoutWriteStatement()`
- `TOKEN_IDENT` → disambiguate via peek:
  - `,` → multi-assign
  - `=` → assign
  - `<<` → push statement
  - `<<<` → nested push statement
  - `$` → lookahead for dyn-access-set (`ident$var = val`)
  - `@` → lookahead for index-set / key-set vs expression
  - `{` after ident → hashmap declaration (`name{k1, k2} = [v1, v2]`)
  - else → expression statement

### Lookahead for index-set and key-set

Uses `save()`/`restore()` to tentatively parse `ident$...` or `ident@...`. If followed by `=`, commits as `DynAccessSetStmt`, `IndexSetStmt`, or `KeySetStmt`. Otherwise backtracks and parses as expression.

### Postfix operators: `$` and `@`

Two postfix operators for access:
- `map$var` → `DynAccessExpr` (dynamic access: dispatches on key type — int→index, string→key)
- `map@name` (bare identifier) → `KeyAccessExpr` (literal string key)
- `arr@0` (number literal) → `IndexExpr` (numeric index)

Bare numbers and null are rejected outside `[]` brackets. The parser tracks `arrayDepth` — these tokens are only allowed inside array literals. `true` and `false` are standalone keywords that parse as `ArrayLiteral` producing `[1]` and `[]` respectively — they do NOT require brackets. `[true]` creates `[[1]]` (nested).

## Codegen (`pkg/codegen/codegen.go`)

Lowers the sloplang AST to `go/ast` nodes, then formats with `go/format`.

### Generated structure

```go
package main

import sloprt "github.com/saad039/sloplang/pkg/runtime"

var x *sloprt.SlopValue  // hoisted top-level variables
var y *sloprt.SlopValue

func userFn1(...) *sloprt.SlopValue { ... }  // can read/write globals
func userFn2(...) *sloprt.SlopValue { ... }

func main() {
    x = ...  // uses = (not :=) since var is package-level
    y = ...
}
```

### Key design patterns

- **Variable hoisting:** Top-level `AssignStmt`, `HashDeclStmt`, and `MultiAssignStmt` names are collected in a first pass and emitted as `var name *sloprt.SlopValue` at package level. `main()` uses `=` for these.
- **Variable tracking:** `declared map[string]bool` — first use emits `:=`, subsequent uses emit `=`
- **Scope isolation with global seeding:** `lowerFnDecl` saves/restores `declared`. The new scope is seeded with globals (so assignments to global names use `=`) and params (pre-registered as declared). New names inside functions use `:=` (local).
- **Unused variable suppression:** Every `:=` declaration followed by `_ = varName`
- **Builtin dispatch:** `CallExpr` checks builtin map (`str` → `sloprt.Str`, `split` → `sloprt.Split`, `to_num` → `sloprt.ToNum`, `to_chars` → `sloprt.ToChars`, `to_int` → `sloprt.ToInt`, `to_float` → `sloprt.ToFloat`, `fmt_float` → `sloprt.FmtFloat`), else emits direct Go function call
- **Dual-return detection:** `isDualReturn()` identifies expressions that return two values (StdinReadExpr, FileReadExpr, `to_num` calls) — these skip `UnpackTwo` wrapping in `lowerMultiAssign`
- **Binary op map:** Maps sloplang operators to runtime function names (e.g., `"+"` → `"Add"`, `"++"` → `"Concat"`)
- **Unary op dispatch:** `-` → `Negate`, `!` → `Not`, `#` → `Length`, `~` → `Unique`, `##` → `MapKeys`, `@@` → `MapValues`
- **Hashmap declaration:** `HashDeclStmt` lowers to `sloprt.MapFromKeysValues` with a `[]string` composite literal for keys
- **Key access/set:** `KeyAccessExpr`/`KeySetStmt` lower to `sloprt.IndexKeyStr`/`IndexKeySetStr`; `DynAccessExpr`/`DynAccessSetStmt` lower to `sloprt.DynAccess`/`DynAccessSet` (type-dispatching runtime functions)
- **Null literal:** `NullLiteral` lowers to `sloprt.NewSlopValue(sloprt.SlopNull{})`

## Runtime (`pkg/runtime/ops.go`)

All operations are functions that take/return `*SlopValue`:

### Arithmetic (element-wise, same-length required)
`Add`, `Sub`, `Mul`, `Div`, `Mod`, `Pow`, `Negate`

### Comparison (single-element only)
`Eq`, `Neq` — deep structural equality on any-size arrays and hashmaps (compares lengths, keys, and elements recursively); `Lt`, `Gt`, `Lte`, `Gte` — single-element only. All return `[1]` (truthy) or `[]` (falsy)

### Logical
`And`, `Or`, `Not` — operate on truthiness

### Array operations
| Function | Mutates? | Description |
|----------|----------|-------------|
| `Index(sv, idx)` | No | Returns element at index |
| `IndexSet(sv, idx, val)` | Yes | Sets element at index |
| `Length(sv)` | No | Returns `[len]` |
| `Push(sv, val)` | Yes | Appends val's elements |
| `Pop(sv)` | Yes | Removes+returns last element |
| `RemoveAt(sv, idx)` | Yes | int: removes+returns element at index; string: removes+returns value for hashmap key |
| `Slice(sv, lo, hi)` | No | Returns sub-array `[lo:hi)` |
| `Concat(a, b)` | No | Returns new combined array |
| `Remove(sv, val)` | No | Returns new array without first occurrence |
| `Contains(sv, val)` | No | Returns `[1]` or `[]` |
| `Unique(sv)` | No | Returns deduplicated array |
| `NestPush(sv, val)` | Yes | Appends val as a single nested element (no spread) |

### Dynamic access (type-dispatching)
| Function | Description |
|----------|-------------|
| `DynAccess(sv, key)` | If key is int64 → `Index`, if string → `IndexKeyStr` |
| `DynAccessSet(sv, key, val)` | If key is int64 → `IndexSet`, if string → `IndexKeySetStr` |

### Hashmap operations
| Function | Mutates? | Description |
|----------|----------|-------------|
| `MapFromKeysValues(keys, vals)` | N/A | Creates hashmap with parallel Keys and Elements |
| `IndexKeyStr(sv, key)` | No | Returns value for literal string key |
| `IndexKey(sv, key)` | No | Returns value for dynamic key (SlopValue) |
| `IndexKeySetStr(sv, key, val)` | Yes | Sets value for literal string key (appends if new) |
| `IndexKeySet(sv, key, val)` | Yes | Sets value for dynamic key (SlopValue) |
| `MapKeys(sv)` | No | Returns array of key strings |
| `MapValues(sv)` | No | Returns array of values (Keys stripped) |

### I/O operations (`pkg/runtime/io.go`)
| Function | Returns | Description |
|----------|---------|-------------|
| `StdinRead()` | `(*SlopValue, *SlopValue)` | Reads one line from stdin; returns `(line, [0])` or `("", [1])` |
| `FileRead(path)` | `(*SlopValue, *SlopValue)` | Reads entire file; returns `(data, [0])` or `("", [1])` |
| `FileWrite(path, data)` | void | Writes data to file (truncates); panics on error |
| `FileAppend(path, data)` | void | Appends data to file; panics on error |
| `Split(sv, sep)` | `*SlopValue` | Splits string by separator; empty sep returns original |
| `ToNum(sv)` | `(*SlopValue, *SlopValue)` | Parses string to int64/float64; returns `(val, [0])` or `([], [1])` |

### Type Casting & Formatting
| Function | Returns | Description |
|----------|---------|-------------|
| `ToChars(sv)` | `*SlopValue` | Decomposes a string into an array of single-character strings |
| `ToInt(sv)` | `*SlopValue` | Converts float64 to int64 (truncates); panics if not a float |
| `ToFloat(sv)` | `*SlopValue` | Converts int64 to float64; panics if not an int |
| `FmtFloat(sv, fmt)` | `*SlopValue` | Formats a float with a format string (e.g., `"%.2f"`) |

### Helpers
- `Str(sv)` — converts to string representation
- `Iterate(sv)` — returns `[]*SlopValue` for for-in loops
- `UnpackTwo(sv)` — destructures for multi-assign
- `deepEqual(a, b)` — structural equality for Contains/Remove/Unique

## Testing (`pkg/codegen/*_e2e_test.go`)

### E2E Test Harness

- `runE2E(t, source)` — transpiles `.slop` source → compiles Go → runs binary → returns stdout
- `runE2EExpectPanic(t, source)` — asserts non-zero exit code (runtime panic)
- `runE2EWithStdin(t, source, stdin)` — provides stdin via pipe

### Phase 8 Program Tests (`tests/programs/programs_test.go`)

10 self-contained `.slop` programs tested at multiple sizes via template substitution (`SIZE_PLACEHOLDER`):

| Program | Sizes | Description |
|---------|-------|-------------|
| fibonacci | 0,1,5,10,20 | Recursive fib |
| wordcount | 0,1,5,10,100,10K,100K | LCG-sampled 100-word vocab |
| array_ops_demo | 0,1,5,10,100,10K,100K | Push/contains/remove-at/index-set |
| linear_search | 0,1,5,10,100,10K,100K | Sequential scan, dual-return |
| binary_search | 0,1,5,10,100,10K,100K | With embedded merge sort |
| bubble_sort | 0,1,5,10,100 | In-place swap (O(n²) limits sizes) |
| merge_sort | 0,1,5,10,100,10K,100K | Recursive split/merge |
| quick_sort | 0,1,5,10,100,10K,100K | Functional partition |
| heap_sort | 0,1,5,10,100,10K,100K | In-place heapify |
| bst | 0,1,5,10,100,10K | Parallel-array BST with iterative inorder |

Each test builds the CLI, substitutes size, runs the `.slop` program, reads `results.txt`, and diffs against Go-computed expected output.

### Semantic E2E Test Suite (Phase 7.5)

355 tests across 9 domain files, each testing a specific semantic rule through the full pipeline:

| File | Domain | Tests |
|------|--------|-------|
| `semantic_mutation_e2e_test.go` | IndexSet, DynAccessSet, KeySetStr, Push, Pop, RemoveAt | 50 |
| `semantic_equality_e2e_test.go` | Deep equality, cross-type, null, hashmap, ordered comparisons | 46 |
| `semantic_format_e2e_test.go` | str(), \|> no newline, FormatValue after ops, hashmaps | 32 |
| `semantic_boolean_e2e_test.go` | IsTruthy strictness, logical ops, true/false keywords | 41 |
| `semantic_null_e2e_test.go` | Null succeeds/panics in every operator context | 36 |
| `semantic_arithmetic_e2e_test.go` | Type safety, element-wise, div-by-zero, negate, precedence | 42 |
| `semantic_array_ops_e2e_test.go` | Index, slice, concat, remove, contains, unique, length | 48 |
| `semantic_hashmap_e2e_test.go` | Declaration, key access/set, ##/@@, dynamic access, functions | 28 |
| `semantic_control_flow_e2e_test.go` | If/else, for-in, infinite loop, functions, scoping, combined | 32 |

Test naming convention: `TestSem_<Domain>_<Case>` (e.g., `TestSem_Mut_IndexSet_SingleInt`).

## Implemented Phases

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Skeleton + Assign + Stdout | Done |
| 2 | Arithmetic + Comparisons + Booleans | Done |
| 3 | Functions + Return + Control Flow | Done |
| 4 | Array Operators | Done |
| 5 | Hashmaps | Done |
| 6 | I/O (stdin + file) | Done |
| 7 | Error Handling Patterns | Done |
| 7.5 | Syntax Strictness Refactor + Semantic E2E Tests (355 tests) | Done |
| 8 | Real Programs (10 programs + E2E tests) | Done |
| 10 | Nested Push (`<<<`) + Type Casting Builtins (`to_chars`, `to_int`, `to_float`, `fmt_float`) | Done |
