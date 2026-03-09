# Phase 3: Functions + Return + Control Flow

**Goal:** Make sloplang programmable with function declarations, if/else branching, for/in iteration, infinite loops, break, return, and multi-assignment.

**Architecture:** Extend lexer with 6 keywords + 2 delimiters + 1 operator. Add 7 new AST nodes. Extend the recursive descent parser with block parsing. Functions lower to Go `*ast.FuncDecl` placed before `main()`. If/else lowers to `*ast.IfStmt`. For/in lowers to `*ast.RangeStmt` over a runtime `Iterate()` helper. Infinite `for { }` loops lower to `*ast.ForStmt` (no init/cond/post). `break` lowers to `*ast.BranchStmt` with `token.BREAK`. Return lowers to `*ast.ReturnStmt`. User-defined function calls reuse the existing `CallExpr` node but codegen distinguishes builtins from user calls.

**What already exists from Phase 2:** `(`, `)` tokens, `CallExpr` AST node + parser, `str()` builtin in codegen and runtime.

---

## Task 1: Lexer — new tokens

**Files:** `pkg/lexer/token.go`, `pkg/lexer/lexer.go`, `pkg/lexer/lexer_test.go`

**New token types:**
- `TOKEN_LBRACE` (`{`), `TOKEN_RBRACE` (`}`)
- `TOKEN_RETURN` (`<-`)
- Keywords: `TOKEN_FN`, `TOKEN_IF`, `TOKEN_ELSE`, `TOKEN_FOR`, `TOKEN_IN`, `TOKEN_BREAK`

**Key changes:**
- Add 8 new constants to `token.go` + entries in `tokenNames`
- Add 6 entries to the `keywords` map: `"fn"`, `"if"`, `"else"`, `"for"`, `"in"`, `"break"`
- In `lexer.go`, handle `{` and `}` as single-char tokens
- Disambiguate `<` vs `<=` vs `<-`: when current char is `<`, peek next — if `-` emit `TOKEN_RETURN`, if `=` emit `TOKEN_LTE`, else emit `TOKEN_LT`

**Tests:** tokenize source with all new tokens, verify disambiguation of `<` / `<=` / `<-`, verify keywords resolve correctly via `LookupIdent`, verify `{` `}` inside expressions.

---

## Task 2: AST — new node types

**Files:** `pkg/parser/ast.go`

**New statement nodes (all implement `Stmt` interface):**

- `FnDeclStmt` — fields: `Name string`, `Params []string`, `Body []Stmt`
- `IfStmt` — fields: `Condition Expr`, `Body []Stmt`, `Else []Stmt` (nil if no else)
- `ForInStmt` — fields: `VarName string`, `Iterable Expr`, `Body []Stmt`
- `ForLoopStmt` — fields: `Body []Stmt` (infinite loop, exited with `break`)
- `BreakStmt` — no fields
- `ReturnStmt` — fields: `Value Expr` (nil for bare `<-`)
- `MultiAssignStmt` — fields: `Names []string`, `Value Expr`

Each gets `stmtNode()` and `TokenLiteral()` methods following existing pattern.

---

## Task 3: Parser — block parsing + new statements

**Files:** `pkg/parser/parser.go`, `pkg/parser/parser_test.go`

**Block parsing:** Add `parseBlock()` — consumes `{`, parses statements until `}`, consumes `}`, returns `[]Stmt`. All body fields use this.

**Statement dispatch in `parseStatement()`:**
- `TOKEN_FN` → `parseFnDecl()`
- `TOKEN_IF` → `parseIfStmt()`
- `TOKEN_FOR` → `parseForStmt()` — dispatches to `parseForLoopBody()` (if next is `{`) or `parseForInBody()` (if next is IDENT)
- `TOKEN_BREAK` → `parseBreakStmt()`
- `TOKEN_RETURN` (`<-`) → `parseReturnStmt()`
- `TOKEN_IDENT` with peek at second token:
  - If next is `TOKEN_COMMA` → try `parseMultiAssign()` (lookahead: `ident, ident, ... = expr`)
  - If next is `TOKEN_ASSIGN` → `parseAssignStatement()` (existing)
  - Otherwise → fall through to expression statement (for bare function calls like `foo()`)

**parseFnDecl():** consume `fn`, consume ident (name), consume `(`, collect param idents separated by commas, consume `)`, call `parseBlock()` for body.

**parseIfStmt():** consume `if`, parse condition expression, call `parseBlock()` for body. If next token is `TOKEN_ELSE`, consume it, call `parseBlock()` for else body.

**parseForStmt():** consume `for`. If next token is `{`, call `parseForLoopBody()` → infinite loop. Otherwise call `parseForInBody()` → for-in loop.

**parseForLoopBody():** call `parseBlock()` for body. Returns `ForLoopStmt`.

**parseForInBody():** consume ident (loop var), expect `TOKEN_IN`, parse iterable expression, call `parseBlock()` for body. Returns `ForInStmt`.

**parseReturnStmt():** consume `<-`. If next token starts an expression (not `}` or EOF), parse expression as value. Otherwise value is nil (returns empty).

**parseMultiAssign():** collect identifiers separated by commas, consume `=`, parse expression.

**Expression statements:** When `parseStatement()` sees a `TOKEN_IDENT` followed by `TOKEN_LPAREN`, it's a bare function call. Parse as expression, wrap in a new `ExprStmt` node (add to ast.go — fields: `Expr Expr`). This handles `foo()` on its own line.

**Tests:**
- Parse `fn add(a, b) { <- a + b }` → verify `FnDeclStmt` with name, 2 params, body containing `ReturnStmt`
- Parse `fn foo() { }` → verify empty body
- Parse `if [1] { |> "yes" }` → verify `IfStmt` with no else
- Parse `if [1] { |> "yes" } else { |> "no" }` → verify else branch
- Parse `for x in items { |> str(x) }` → verify `ForInStmt` fields
- Parse `<- [1]` → verify `ReturnStmt` with value
- Parse `a, b = foo()` → verify `MultiAssignStmt` with 2 names
- Parse nested: `fn f() { if [1] { for x in [1] { |> str(x) } } }`

---

## Task 4: Runtime — Iterate helper

**Files:** `pkg/runtime/ops.go`, `pkg/runtime/ops_test.go`

**New function:** `Iterate(sv *SlopValue) []*SlopValue` — returns a slice where each element of the input array is wrapped in its own `*SlopValue`. For `*SlopValue` elements (nested arrays), they're returned as-is. This is what `for/in` ranges over.

**Hint:** `for _, elem := range sv.Elements { result = append(result, NewSlopValue(elem)) }` — but if elem is already `*SlopValue`, use it directly.

**Tests:**
- `Iterate([1, 2, 3])` → 3 `*SlopValue`s each containing one int64
- `Iterate([])` → empty slice
- `Iterate([[1,2], [3,4]])` → 2 `*SlopValue`s, each being the nested array
- `Iterate(["a", "b"])` → 2 `*SlopValue`s each containing one string

---

## Task 5: Codegen — lower new statement types

**Files:** `pkg/codegen/codegen.go`, `pkg/codegen/codegen_test.go`

**Structural change:** Currently `Generate()` puts everything in `main()`. Now function declarations must be emitted as top-level `*ast.FuncDecl` nodes, separate from `main()`. Split `lowerStmt` processing: if a statement is `FnDeclStmt`, collect it as a top-level decl. Everything else goes into `main()`.

### FnDeclStmt → *ast.FuncDecl

- Function name → `ast.NewIdent(name)`
- Each param → `*ast.Field` with name and type `*ast.SelectorExpr` for `*sloprt.SlopValue`
- Return type → `*ast.Field` with type `*sloprt.SlopValue` (single return) or tuple for multi-return
- Body → recursively lower each statement in the body
- For multi-return (functions that return `a, b`), use a return type of `(*sloprt.SlopValue, *sloprt.SlopValue)`

**Design decision on return type:** All functions return `*sloprt.SlopValue`. For multi-assign to work, functions that do dual-return need to return `(*sloprt.SlopValue, *sloprt.SlopValue)`. Simple approach: always generate functions with dual return `(*sloprt.SlopValue, *sloprt.SlopValue)`, and single-return `<- x` lowers to `return x, nil`. The caller can then do `a, b := fn()` or `a := fn()` — but Go doesn't allow ignoring a return value like that.

**Better approach:** Inspect the function body to determine return arity. If any `ReturnStmt` has a multi-value pattern (or if any call site uses `MultiAssignStmt`), generate dual return. Otherwise single return. For Phase 3 simplicity: **all user functions return a single `*sloprt.SlopValue`**. Multi-assign `a, b = call()` is handled by the function returning a 2-element `*SlopValue`, and the multi-assign codegen destructures it: `a = result.Elements[0].(*SlopValue)`, `b = result.Elements[1].(*SlopValue)`. This keeps Go function signatures uniform.

### IfStmt → *ast.IfStmt

- Condition: `sloprt.IsTruthy(expr)` — but `IsTruthy` is a method on `*SlopValue`. So condition becomes `expr.IsTruthy()` → in go/ast: `*ast.CallExpr` on the lowered condition expression with selector `IsTruthy`.
- Body: lower each statement in body block
- Else: if present, the `Else` field is an `*ast.BlockStmt` containing lowered else statements

**Wait — `IsTruthy()` is a method, not a function.** Codegen needs to emit `(loweredExpr).IsTruthy()`. Hint: `&ast.CallExpr{Fun: &ast.SelectorExpr{X: loweredCondition, Sel: ast.NewIdent("IsTruthy")}}`.

### ForInStmt → *ast.RangeStmt

- Range over `sloprt.Iterate(loweredIterable)` — this returns `[]*SlopValue`
- Key: `ast.NewIdent("_")` (discard index)
- Value: `ast.NewIdent(varName)`
- Body: lower each statement
- Tok: `token.DEFINE` (`:=`)

### ReturnStmt → *ast.ReturnStmt

- If value is non-nil: `&ast.ReturnStmt{Results: []ast.Expr{loweredValue}}`
- If value is nil: `&ast.ReturnStmt{Results: []ast.Expr{callSloprt("NewSlopValue")}}` — return empty array `[]`

### MultiAssignStmt → destructured assignment

Lower `a, b = call()` to:
```
tmp := loweredExpr
a := tmp.Elements[0].(*sloprt.SlopValue)   // or wrap raw element
b := tmp.Elements[1].(*sloprt.SlopValue)
```

Hint: use `ast.TypeAssertExpr` for the cast, `ast.IndexExpr` for the element access.

Alternatively simpler: add a runtime helper `UnpackTwo(sv *SlopValue) (*SlopValue, *SlopValue)` that returns elements 0 and 1 as `*SlopValue`. Then codegen emits `a, b := sloprt.UnpackTwo(loweredExpr)`.

### User-defined function calls

Currently `CallExpr` codegen checks a `builtins` map and only handles `str`. Extend: if the function name is not in the builtins map, emit a direct Go function call `name(loweredArgs...)`. The Go function was already emitted as a top-level `*ast.FuncDecl`.

### ExprStmt

If added for bare function calls, simply lower: `&ast.ExprStmt{X: g.lowerExpr(s.Expr)}`.

**Tests:**
- Verify `fn add(a, b) { <- a + b }` produces `func add(a *sloprt.SlopValue, b *sloprt.SlopValue) *sloprt.SlopValue`
- Verify `if` produces `ast.IfStmt` with `.IsTruthy()` call
- Verify `for x in items` produces range over `sloprt.Iterate`
- Verify `<- x` produces `return x`
- Verify user call `add(x, y)` produces `add(x, y)` (not `sloprt.Add`)
- Verify `a, b = foo()` produces `sloprt.UnpackTwo` call

---

## Task 6: Runtime — UnpackTwo helper

**Files:** `pkg/runtime/ops.go`, `pkg/runtime/ops_test.go`

**New function:** `UnpackTwo(sv *SlopValue) (*SlopValue, *SlopValue)` — takes a `*SlopValue` with at least 2 elements, returns elements 0 and 1 each wrapped as `*SlopValue`. If element is already `*SlopValue`, use directly. If raw (int64, string, etc.), wrap in `NewSlopValue()`. Panic if fewer than 2 elements.

**Tests:** unpack `[1, 2]`, unpack `[[1,2], [3,4]]`, panic on `[1]`.

---

## Task 7: E2E tests — 50+ tests

**Files:** `pkg/codegen/codegen_e2e_test.go`

**Categories and test cases:**

### Functions — basic (8 tests)
- `fn` with single param, call it, verify output
- `fn` with two params, add them, verify output
- `fn` with no params, returns constant
- `fn` calling another `fn`
- Recursive `fn` (factorial or countdown)
- `fn` with str() inside body
- `fn` with expression as argument: `add([1] + [2], [3])`
- `fn` result used in expression: `add([1], [2]) + [3]`

### Return (5 tests)
- `<-` with value
- `<-` bare (no value) → returns `[]`
- Early return: if condition, return early, else continue
- Return from nested if
- Return expression: `<- a + b`

### If/else (10 tests)
- `if` truthy (non-empty array) → body runs
- `if` falsy (empty array) → body skipped
- `if`/`else` truthy → if branch
- `if`/`else` falsy → else branch
- `if` with comparison: `if [1] == [1]`
- `if` with logical: `if [1] && [1]`
- Nested `if` inside `if`
- `if` inside `fn` body
- `if` with `!` (not): `if ![0]` is truthy (since `[0]` is truthy, `!` makes it falsy → else runs)
- `if` `else` with multiple statements in each branch

### For/in (8 tests)
- Iterate over `[1, 2, 3]`, print each
- Iterate over empty array → no output
- Iterate over single element
- Iterate with computation: `item + [1]` inside loop
- Nested for loops
- For inside fn
- For with if inside
- For over string array: `["a", "b", "c"]`

### Multi-assign (4 tests)
- `a, b = fn()` where fn returns 2-element array
- Use both returned values
- Multi-assign with nested SlopValues
- Multi-assign from non-function expression: `a, b = [10, 20]` (unpack directly)

### Combined / integration (10 tests)
- Fibonacci: `fn fib(n) { if n <= [1] { <- n } <- fib(n - [1]) + fib(n - [1] - [1]) }` — verify first 10 values
- Factorial recursive
- fn that takes result of another fn as arg
- for loop calling fn on each element
- if/else inside for inside fn
- fn returning result of if/else (ternary-like pattern)
- Multiple fns calling each other
- fn with for loop accumulating: sum array elements
- Bare function call statement (no assignment): `print_stuff()`
- Nested function: fn defined, called from another fn, with if inside

### Edge cases (5+ tests)
- fn with same param name as global variable (shadowing)
- for loop variable doesn't leak outside loop scope (or does — Go semantics)
- deeply nested: fn → if → for → if → return
- fn with all statement types: assign, if, for, return, stdout write
- fns.slop roadmap test

**Total: 50+ tests**

---

## Task 8: Roadmap e2e — fns.slop

**Files:** `sloplang/examples/fns.slop`

Create the example file from the roadmap:
```
fn add(a, b) {
    <- a + b
}

result = add([3], [4])
|> str(result)

if [1] == [1] {
    |> "equal"
}

items = [10, 20, 30]
for item in items {
    |> str(item)
}
```

Expected output: `[7]`, `equal`, `10`, `20`, `30` — each on its own line.

Verify this passes as an e2e test.

---

## Task 9: Final verification + flip passes

- Run `go test ./...` — all tests pass
- Run `go vet ./...` — clean
- Run `go fmt ./...` — clean
- Flip all `"passes": false` to `true` in `phase3-functions-control-flow.json`
- Commit
