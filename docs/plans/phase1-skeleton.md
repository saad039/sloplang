# Phase 1: Skeleton + Assign + Stdout

## Context

Sloplang is a brand new transpiler project with complete specs (PRD + roadmap) but zero code. Phase 1 builds the minimal end-to-end pipeline: read a `.slop` file, lex it, parse it, generate Go source via `go/ast`, compile and run it. The target program is:

```
x = [1, 2, 3]
|> "hello world"
```

Which should print `hello world` to stdout.

---

## Module Path Decision

Use `github.com/saad039/sloplang` as the Go module path. Generated Go code imports the runtime as `sloprt "github.com/saad039/sloplang/pkg/runtime"`.

---

## Implementation Steps

### Step 0: Project Scaffolding

All Go code lives in the `sloplang/` subfolder (not the project root).

- `cd sloplang/ && go mod init github.com/saad039/sloplang`
- Create directories inside `sloplang/`: `cmd/slop/`, `pkg/lexer/`, `pkg/parser/`, `pkg/codegen/`, `pkg/runtime/`, `examples/`

### Step 1: Runtime — `pkg/runtime/slop_value.go`

**SlopValue struct:**
```go
type SlopValue struct {
    Elements []any    // int64, uint64, float64, string, or *SlopValue
    Keys     []string // parallel to Elements for hashmaps; nil for plain arrays
}
```

**Functions to implement:**
- `NewSlopValue(elems ...any) *SlopValue` — generic constructor accepting raw Go values
- `(sv *SlopValue) IsTruthy() bool` — returns `len(sv.Elements) > 0`
- `StdoutWrite(v *SlopValue)` — prints single-element strings raw, otherwise prints array representation like `[1, 2, 3]`

**Key design decision:** `NewSlopValue` takes variadic `any`. For `[1, 2, 3]`, codegen produces `sloprt.NewSlopValue(int64(1), int64(2), int64(3))`. For nested arrays like `[[1,2], [3,4]]`, inner arrays are `*SlopValue` elements. This is clean and handles all cases.

**Tests:** `pkg/runtime/slop_value_test.go` — constructor variants, IsTruthy, StdoutWrite output capture.

### Step 2: Lexer Tokens — `pkg/lexer/token.go`

**Token types for Phase 1:**
`TOKEN_EOF`, `TOKEN_ILLEGAL`, `TOKEN_INT`, `TOKEN_UINT`, `TOKEN_FLOAT`, `TOKEN_STRING`, `TOKEN_IDENT`, `TOKEN_ASSIGN` (`=`), `TOKEN_PIPE_GT` (`|>`), `TOKEN_LBRACKET`, `TOKEN_RBRACKET`, `TOKEN_COMMA`, `TOKEN_TRUE`, `TOKEN_FALSE`

**Token struct:** `Type`, `Literal`, `Line`, `Col`

### Step 3: Lexer Scanner — `pkg/lexer/lexer.go`

Hand-written scanner with `NextToken()` and `Tokenize()` methods.

**Handles:**
- Whitespace/newline skipping with line tracking
- `//` comment skipping
- Number literals: digits → check for `.` (float) or `u` suffix (uint), else int
- String literals with escape sequences (`\"`, `\\`, `\n`, `\t`)
- Multi-char operators: `|>` (peek ahead from `|`)
- Single-char: `[`, `]`, `,`, `=`
- Identifiers + keyword lookup (`true`, `false`)

**Note:** Negative numbers like `-10` are NOT handled by the lexer. `-` will be a separate token (for Phase 2 unary negation).

**Tests:** `pkg/lexer/lexer_test.go` — tokenize assignments, stdout writes, number types, comments, escapes, multiline.

### Step 4: AST Definitions — `pkg/parser/ast.go`

**Interfaces:** `Node`, `Stmt`, `Expr` (with marker methods)

**Node types:**
- `Program` — `[]Stmt`
- `AssignStmt` — `Name string`, `Value Expr`
- `StdoutWriteStmt` — `Value Expr`
- `ArrayLiteral` — `Elements []Expr`
- `NumberLiteral` — `Value string`, `NumType` (Int/Uint/Float)
- `StringLiteral` — `Value string`
- `Identifier` — `Name string`

### Step 5: Parser — `pkg/parser/parser.go`

Recursive descent. Grammar:
```
program     := stmt*
stmt        := assign_stmt | stdout_stmt
assign_stmt := IDENT ASSIGN expr
stdout_stmt := PIPE_GT expr
expr        := array_lit | string_lit | number_lit | ident
array_lit   := LBRACKET (expr (COMMA expr)*)? RBRACKET
```

**Methods:** `Parse()`, `parseStatement()`, `parseAssignStatement()`, `parseStdoutWriteStatement()`, `parseExpression()`, `parseArrayLiteral()`, helpers (`curToken`, `peekToken`, `advance`, `expect`).

**Tests:** `pkg/parser/parser_test.go` — all statement/expression types, empty arrays, mixed arrays, error cases.

### Step 6: Codegen — `pkg/codegen/codegen.go`

Builds `go/ast.File` and formats with `go/format.Node`.

**Generated structure:**
```go
package main
import sloprt "github.com/saad039/sloplang/pkg/runtime"
func main() { /* transpiled statements */ }
```

**Lowering:**
- `AssignStmt` → `*ast.AssignStmt` with `token.DEFINE` (`:=`)
- `StdoutWriteStmt` → `sloprt.StdoutWrite(expr)` call
- `ArrayLiteral [1, 2, 3]` → `sloprt.NewSlopValue(int64(1), int64(2), int64(3))`
  - Uses `lowerRawValue()` for array elements to avoid double-wrapping
- `NumberLiteral` → `int64(42)` / `uint64(42)` / `float64(3.14)` cast expression
- `StringLiteral` → `*ast.BasicLit{Kind: token.STRING}` (re-quote with `strconv.Quote`)
- `Identifier` → `*ast.Ident`

**Helper:** `callSloprt(funcName, args...)` builds `sloprt.X(args...)` call expressions.

**Known pitfall:** `go/format.Node` needs valid (or zero) positions. Use `token.NoPos` everywhere.

**Tests:** `pkg/codegen/codegen_test.go` — verify generated Go source compiles, correct structure.

### Step 7: CLI — `cmd/slop/main.go`

Reads `.slop` file from `os.Args[1]`, runs lexer → parser → codegen, writes `.go` file alongside input. Reports errors at each stage.

### Step 8: E2E Test + Example

**`examples/hello.slop`:**
```
x = [1, 2, 3]
|> "hello world"
```

**E2E test** (in `pkg/codegen/codegen_e2e_test.go` or top-level):
1. Lex → parse → generate from source string
2. Write generated `.go` + `go.mod` (with `replace` directive to local module) to temp dir
3. `go build` the generated code
4. Run binary, capture stdout
5. Assert output is `hello world\n`

---

## File Summary

| File | Purpose |
|------|---------|
| `go.mod` | Module definition |
| `cmd/slop/main.go` | CLI entry point |
| `pkg/runtime/slop_value.go` | SlopValue type + StdoutWrite |
| `pkg/runtime/slop_value_test.go` | Runtime unit tests |
| `pkg/lexer/token.go` | Token type definitions |
| `pkg/lexer/lexer.go` | Hand-written scanner |
| `pkg/lexer/lexer_test.go` | Lexer unit tests |
| `pkg/parser/ast.go` | AST node types |
| `pkg/parser/parser.go` | Recursive descent parser |
| `pkg/parser/parser_test.go` | Parser unit tests |
| `pkg/codegen/codegen.go` | go/ast lowering + formatting |
| `pkg/codegen/codegen_test.go` | Codegen unit tests |
| `pkg/codegen/codegen_e2e_test.go` | E2E integration test |
| `examples/hello.slop` | Phase 1 example program |

---

## Verification

1. `go test ./pkg/runtime/...` — all runtime tests pass
2. `go test ./pkg/lexer/...` — all lexer tests pass
3. `go test ./pkg/parser/...` — all parser tests pass
4. `go test ./pkg/codegen/...` — all codegen + E2E tests pass
5. `go run ./cmd/slop/main.go examples/hello.slop` — produces `examples/hello.go`
6. `cd examples && go run hello.go` (or build+run) — prints `hello world`
7. `go vet ./...` and `go fmt ./...` — clean
