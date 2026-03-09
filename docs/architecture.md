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
    │       └── ops.go             # All operations: arithmetic, comparison, logical, array ops
    └── examples/                  # Example .slop programs
```

## Core Types

### SlopValue (`pkg/runtime/slop_value.go`)

The universal value type. Every sloplang value is a `SlopValue`:

```go
type SlopValue struct {
    Elements []any    // int64, uint64, float64, string, or *SlopValue
    Keys     []string // parallel to Elements for hashmaps; nil for plain arrays
}
```

- `[]` (empty Elements) is falsy; everything else is truthy
- `NewSlopValue(elems ...any)` constructs values
- `FormatValue` renders: single-element → raw value (e.g. `7`), multi-element → `[1, 2, 3]`

### Token (`pkg/lexer/token.go`)

```go
type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Col     int
}
```

Token types cover: literals (INT, UINT, FLOAT, STRING, IDENT), operators (arithmetic, comparison, logical, array), delimiters, keywords, and return (`<-`).

### AST Nodes (`pkg/parser/ast.go`)

Two interfaces: `Stmt` (statements) and `Expr` (expressions).

**Statements:** AssignStmt, StdoutWriteStmt, FnDeclStmt, IfStmt, ForInStmt, ForLoopStmt, BreakStmt, ReturnStmt, MultiAssignStmt, ExprStmt, PushStmt, IndexSetStmt

**Expressions:** ArrayLiteral, NumberLiteral, StringLiteral, Identifier, BinaryExpr, UnaryExpr, CallExpr, IndexExpr, PopExpr, SliceExpr

## Lexer (`pkg/lexer/lexer.go`)

Single-pass, character-by-character tokenizer. Key design decisions:

- **Greedy multi-char matching:** `<<`, `>>`, `++`, `--`, `~@`, `::`, `??`, `**`, `==`, `!=`, `<=`, `>=`, `<-`, `|>`, `||`, `&&` are checked before their single-char prefixes
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
7. Unary: `-`, `!`, `#`, `~`, `>>` (prefix)
8. Call: `name(args...)`
9. Postfix: `@` (index), `::` (slice)
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
  - `@` → lookahead for index-set (`arr@i = val`) vs expression
  - else → expression statement

### Lookahead for index-set

Uses `save()`/`restore()` to tentatively parse `ident@expr`. If followed by `=`, commits as `IndexSetStmt`. Otherwise backtracks and parses as expression.

## Codegen (`pkg/codegen/codegen.go`)

Lowers the sloplang AST to `go/ast` nodes, then formats with `go/format`.

### Generated structure

```go
package main

import sloprt "github.com/saad039/sloplang/pkg/runtime"

func userFn1(...) *sloprt.SlopValue { ... }
func userFn2(...) *sloprt.SlopValue { ... }

func main() {
    // all non-fn statements
}
```

### Key design patterns

- **Variable tracking:** `declared map[string]bool` — first use emits `:=`, subsequent uses emit `=`
- **Scope isolation:** `lowerFnDecl` saves/restores `declared`. Params pre-registered as declared.
- **Unused variable suppression:** Every `:=` declaration followed by `_ = varName`
- **Builtin dispatch:** `CallExpr` checks builtin map (`str` → `sloprt.Str`), else emits direct Go function call
- **Binary op map:** Maps sloplang operators to runtime function names (e.g., `"+"` → `"Add"`, `"++"` → `"Concat"`)
- **Unary op dispatch:** `-` → `Negate`, `!` → `Not`, `#` → `Length`, `~` → `Unique`

## Runtime (`pkg/runtime/ops.go`)

All operations are functions that take/return `*SlopValue`:

### Arithmetic (element-wise, same-length required)
`Add`, `Sub`, `Mul`, `Div`, `Mod`, `Pow`, `Negate`

### Comparison (single-element only)
`Eq`, `Neq`, `Lt`, `Gt`, `Lte`, `Gte` — return `[1]` (truthy) or `[]` (falsy)

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
| `RemoveAt(sv, idx)` | Yes | Removes+returns element at index |
| `Slice(sv, lo, hi)` | No | Returns sub-array `[lo:hi)` |
| `Concat(a, b)` | No | Returns new combined array |
| `Remove(sv, val)` | No | Returns new array without first occurrence |
| `Contains(sv, val)` | No | Returns `[1]` or `[]` |
| `Unique(sv)` | No | Returns deduplicated array |

### Helpers
- `Str(sv)` — converts to string representation
- `Iterate(sv)` — returns `[]*SlopValue` for for-in loops
- `UnpackTwo(sv)` — destructures for multi-assign
- `deepEqual(a, b)` — structural equality for Contains/Remove/Unique

## Implemented Phases

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Skeleton + Assign + Stdout | Done |
| 2 | Arithmetic + Comparisons + Booleans | Done |
| 3 | Functions + Return + Control Flow | Done |
| 4 | Array Operators | Done |
| 5 | Hashmaps | Planned |
| 6 | I/O (stdin + file) | Planned |
| 7 | Error Handling Patterns | Planned |
| 8 | Real Programs | Planned |
