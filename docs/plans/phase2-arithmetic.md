# Phase 2: Arithmetic + Comparisons + Booleans + str() Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make numbers computable — add all arithmetic, comparison, and logical operators plus `str()` builtin.

**Architecture:** Extend each pipeline stage (lexer → parser → codegen → runtime) with operator support. Parser refactored from flat `parseExpression()` to recursive descent precedence chain. All operator semantics live in runtime functions called by generated Go code.

**Tech Stack:** Go (transpiler + runtime), go/ast for codegen

---

### Task 1: Lexer — Add New Token Types

**Files:**
- Modify: `sloplang/pkg/lexer/token.go` — add 17 new token constants and their names
- Modify: `sloplang/pkg/lexer/lexer.go` — add scanning logic in `NextToken()`
- Test: `sloplang/pkg/lexer/lexer_test.go`

**Step 1: Write failing tests**

Add 6 test functions to `lexer_test.go`:
- `TestLexer_ArithmeticOperators` — tokenize `+ - * / % **`, verify each token type
- `TestLexer_ComparisonOperators` — tokenize `== != < > <= >=`
- `TestLexer_LogicalOperators` — tokenize `&& || !`
- `TestLexer_Parentheses` — tokenize `( )`
- `TestLexer_OperatorDisambiguation` — tokenize `= == ! != * ** |> || < <= > >=` in sequence, verify each is correct (this is the critical test)
- `TestLexer_FunctionCall` — tokenize `str(x)` as `IDENT LPAREN IDENT RPAREN`

Follow the existing table-driven pattern from `TestLexer_Assignment`.

**Step 2: Run tests, verify they fail**

Run: `cd sloplang && go test ./pkg/lexer/... -v`

**Step 3: Add token constants**

In `token.go`, add to the `const` iota block after `TOKEN_COMMA`:
- Arithmetic: `TOKEN_PLUS`, `TOKEN_MINUS`, `TOKEN_STAR`, `TOKEN_SLASH`, `TOKEN_PERCENT`, `TOKEN_POWER`
- Comparison: `TOKEN_EQ`, `TOKEN_NEQ`, `TOKEN_LT`, `TOKEN_GT`, `TOKEN_LTE`, `TOKEN_GTE`
- Logical: `TOKEN_AND`, `TOKEN_OR`, `TOKEN_NOT`
- Delimiters: `TOKEN_LPAREN`, `TOKEN_RPAREN`

Add all 17 entries to `tokenNames` map.

**Step 4: Add scanning logic**

In `lexer.go` `NextToken()`, replace the existing switch block. Key disambiguation logic:
- `=` → peek for `=` → `TOKEN_EQ` (`==`) or `TOKEN_ASSIGN` (`=`)
- `!` → peek for `=` → `TOKEN_NEQ` (`!=`) or `TOKEN_NOT` (`!`)
- `*` → peek for `*` → `TOKEN_POWER` (`**`) or `TOKEN_STAR` (`*`)
- `|` → peek for `>` → `TOKEN_PIPE_GT` (`|>`); peek for `|` → `TOKEN_OR` (`||`); else `ILLEGAL`
- `<` → peek for `=` → `TOKEN_LTE` (`<=`) or `TOKEN_LT` (`<`)
- `>` → peek for `=` → `TOKEN_GTE` (`>=`) or `TOKEN_GT` (`>`)
- `&` → peek for `&` → `TOKEN_AND` (`&&`)
- `/` comment check (`//`) must remain before single `/` handling
- New single-char cases: `+`, `-`, `/`, `%`, `(`, `)`

**Step 5: Run tests, verify all pass (new + existing)**

**Step 6: Commit** — `feat: lexer tokenizes arithmetic, comparison, logical operators and parens`

---

### Task 2: AST — Add BinaryExpr, UnaryExpr, CallExpr Nodes

**Files:**
- Modify: `sloplang/pkg/parser/ast.go` — append 3 new node types after `Identifier`

**Step 1: Add new node types**

Append to `ast.go`:
- `BinaryExpr` — fields: `Left Expr`, `Op string`, `Right Expr`. Implements `Expr`.
- `UnaryExpr` — fields: `Op string`, `Operand Expr`. Implements `Expr`.
- `CallExpr` — fields: `Name string`, `Args []Expr`. Implements `Expr`.

Each needs `exprNode()` and `TokenLiteral()` methods, following the existing pattern from `Identifier`.

**Step 2: Verify compilation**

Run: `cd sloplang && go build ./pkg/parser/...`

**Step 3: Commit** — `feat: add BinaryExpr, UnaryExpr, CallExpr AST nodes`

---

### Task 3: Parser — Recursive Descent Precedence Chain

**Files:**
- Modify: `sloplang/pkg/parser/parser.go` — replace `parseExpression()` with precedence chain
- Test: `sloplang/pkg/parser/parser_test.go`

**Step 1: Write failing tests**

Add tests to `parser_test.go`:
- `TestParser_BinaryAdd` — parse `x = [1] + [2]`, assert `BinaryExpr` with `Op: "+"`
- `TestParser_Precedence_MulBeforeAdd` — parse `x = [1] + [2] * [3]`, assert outer `+`, inner right `*`
- `TestParser_UnaryNegate` — parse `x = -[1]`, assert `UnaryExpr` with `Op: "-"`
- `TestParser_UnaryNot` — parse `x = ![1]`, assert `UnaryExpr` with `Op: "!"`
- `TestParser_CallExpr` — parse `|> str(x)`, assert `CallExpr` with `Name: "str"`, 1 arg
- `TestParser_CallExprInAssign` — parse `x = str([1] + [2])`, assert `CallExpr` wrapping a `BinaryExpr` arg
- `TestParser_ParenGrouping` — parse `x = ([1] + [2]) * [3]`, assert outer `*`, inner left `+`
- `TestParser_ComparisonOps` — table-driven: all 6 comparison ops parse to correct `BinaryExpr.Op`
- `TestParser_LogicalOps` — table-driven: `&&` and `||` parse to correct `BinaryExpr.Op`
- `TestParser_LogicalPrecedence` — parse `x = [1] || [2] && [3]`, assert outer `||`, inner right `&&`
- `TestParser_DoubleUnaryNegate` — parse `x = --[5]`, assert nested `UnaryExpr`
- `TestParser_PowerOp` — parse `x = [2] ** [3]`, assert `BinaryExpr` with `Op: "**"`
- `TestParser_StdoutWriteExpr` — parse `|> [1] + [2]`, assert `StdoutWriteStmt` with `BinaryExpr` value

**Step 2: Run tests, verify they fail**

**Step 3: Implement the precedence chain**

Add a `peekToken()` helper (like `curToken()` but returns `p.tokens[p.pos+1]`).

Replace `parseExpression()` with a chain of methods, each calling the next:

```
parseExpression() → parseOr()
parseOr()         → loop on TOKEN_OR, calls parseAnd()
parseAnd()        → loop on TOKEN_AND, calls parseComparison()
parseComparison() → loop on TOKEN_EQ/NEQ/LT/GT/LTE/GTE, calls parseAddSub()
parseAddSub()     → loop on TOKEN_PLUS/MINUS, calls parseMulDivMod()
parseMulDivMod()  → loop on TOKEN_STAR/SLASH/PERCENT, calls parsePower()
parsePower()      → if TOKEN_POWER, recursive call to parsePower() (right-assoc), calls parseUnary()
parseUnary()      → if TOKEN_MINUS/NOT, recursive call to parseUnary(), else calls parseCall()
parseCall()       → if IDENT followed by LPAREN, parse args comma-separated until RPAREN → CallExpr; else calls parsePrimary()
parsePrimary()    → LPAREN expr RPAREN for grouping, plus existing cases (array, string, number, ident, bool)
```

Each binary level: get left from next-higher level, loop while current token matches, get right from next-higher level, wrap in `BinaryExpr`.

**Step 4: Run tests, verify all pass (new + existing)**

**Step 5: Commit** — `feat: recursive descent parser with operator precedence`

---

### Task 4: Runtime — Arithmetic Operations

**Files:**
- Create: `sloplang/pkg/runtime/ops.go`
- Create: `sloplang/pkg/runtime/ops_test.go`

**Step 1: Write failing tests**

Create `ops_test.go` with tests:
- `TestAdd_IntArrays` — `[1,2] + [3,4]` → `[4, 6]`
- `TestAdd_SingleElement` — `[1] + [1]` → `2`
- `TestAdd_Float` — `[3.14] + [2.86]` → `6`
- `TestAdd_Uint` — `[42u] + [8u]` → `50`
- `TestSub` — `[5,3] - [1,1]` → `[4, 2]`
- `TestMul` — `[2,3] * [4,5]` → `[8, 15]`
- `TestDiv` — `[10,6] / [2,3]` → `[5, 2]`
- `TestMod` — `[7,5] % [3,2]` → `[1, 1]`
- `TestPow` — `[2,3] ** [3,2]` → `[8, 9]`
- `TestNegate` — negate `[1,2,3]` → `[-1, -2, -3]`
- `TestNegate_Zero` — negate `[0]` → `0`
- `TestNegate_Float` — negate `[3.14]` → `-3.14`
- `TestAdd_LengthMismatch` — different length arrays → expect panic (use `defer recover`)
- `TestAdd_TypeMismatch` — `int64 + float64` → expect panic
- `TestMul_ByZero` — `[1] * [0]` → `0`

Use `FormatValue()` to assert results as strings. Use `defer func() { recover() }` pattern for panic tests.

**Step 2: Run tests, verify they fail**

**Step 3: Implement arithmetic ops**

Create `ops.go` with:
- Helper `checkLengths(a, b)` — panics if `len(a.Elements) != len(b.Elements)`
- `Add(a, b)`, `Sub(a, b)`, `Mul(a, b)`, `Div(a, b)`, `Mod(a, b)`, `Pow(a, b)` — all follow the same pattern: check lengths, iterate elements, type-switch on `int64/uint64/float64`, panic on type mismatch
- `Negate(a)` — iterate elements, negate each (type-switch)
- `Div` must panic on division by zero
- `Pow` uses `math.Pow` (cast to float64, cast back)
- `Mod` only supports `int64` and `uint64` (not float)
- Each element-level helper (`addElems`, `subElems`, etc.) type-switches on `int64`, `uint64`, `float64` and panics on mismatch

**Step 4: Run tests, verify all pass**

**Step 5: Commit** — `feat: runtime arithmetic operations (Add, Sub, Mul, Div, Mod, Pow, Negate)`

---

### Task 5: Runtime — Comparison, Logical, and Str Operations

**Files:**
- Modify: `sloplang/pkg/runtime/ops.go` — append comparison/logical/str functions
- Modify: `sloplang/pkg/runtime/ops_test.go` — append tests

**Step 1: Write failing tests**

Append to `ops_test.go`:
- `TestEq_True/False`, `TestEq_String` — single-element int and string equality
- `TestNeq` — inequality
- `TestLt`, `TestGt`, `TestLte_Equal`, `TestGte_Less` — ordering comparisons
- `TestEq_MultiElement_Panics` — multi-element comparison panics
- `TestAnd_BothTruthy`, `TestAnd_LeftFalsy` — and logic
- `TestOr_LeftFalsy`, `TestOr_BothFalsy` — or logic
- `TestNot_Truthy`, `TestNot_Falsy` — not logic
- `TestStr`, `TestStr_SingleElement`, `TestStr_Empty` — str builtin

**Step 2: Run tests, verify they fail**

**Step 3: Implement**

Append to `ops.go`:
- Helper `boolResult(bool)` — returns `NewSlopValue(int64(1))` or `NewSlopValue()`
- Helper `checkSingleElement(a, b, op)` — panics if either has `len != 1`
- Helper `compareElems(a, b)` — returns `-1/0/1`, type-switches on `int64/uint64/float64/string`, panics on mismatch
- `Eq`, `Neq`, `Lt`, `Gt`, `Lte`, `Gte` — call `checkSingleElement`, then `compareElems`, return `boolResult`
- `And(a, b)` — if `a.IsTruthy()` return `b`, else return `NewSlopValue()`
- `Or(a, b)` — if `a.IsTruthy()` return `a`, else return `b`
- `Not(a)` — return `boolResult(!a.IsTruthy())`
- `Str(a)` — return `NewSlopValue(FormatValue(a))`

**Step 4: Run tests, verify all pass**

**Step 5: Commit** — `feat: runtime comparison, logical ops, and Str builtin`

---

### Task 6: Codegen — Lower BinaryExpr, UnaryExpr, CallExpr

**Files:**
- Modify: `sloplang/pkg/codegen/codegen.go` — extend `lowerExpr` and `lowerRawValue`
- Test: `sloplang/pkg/codegen/codegen_test.go`

**Step 1: Write failing tests**

Append to `codegen_test.go`:
- `TestCodegen_BinaryAdd` — verify output contains `sloprt.Add(`
- `TestCodegen_BinaryAllOps` — table-driven: all 14 binary ops map to correct `sloprt.X(` call
- `TestCodegen_UnaryNegate` — verify `sloprt.Negate(`
- `TestCodegen_UnaryNot` — verify `sloprt.Not(`
- `TestCodegen_CallStr` — verify `sloprt.Str(`

**Step 2: Run tests, verify they fail**

**Step 3: Implement**

In `lowerExpr`, add cases before `default`:
- `*parser.BinaryExpr` — use a map from op string (`+`, `-`, etc.) to runtime function name (`Add`, `Sub`, etc.), call `callSloprt(fname, lowerExpr(left), lowerExpr(right))`
- `*parser.UnaryExpr` — `-` maps to `Negate`, `!` maps to `Not`
- `*parser.CallExpr` — map of builtin names (`"str"` → `"Str"`), lower each arg with `lowerExpr`, call `callSloprt`

In `lowerRawValue`, add passthrough cases for `*parser.BinaryExpr`, `*parser.UnaryExpr`, `*parser.CallExpr` that delegate to `lowerExpr`.

**Step 4: Run tests, verify all pass**

**Step 5: Commit** — `feat: codegen lowers BinaryExpr, UnaryExpr, CallExpr to sloprt calls`

---

### Task 7: E2E Tests — Full 58-Test Suite

**Files:**
- Modify: `sloplang/pkg/codegen/codegen_e2e_test.go` — append 58 + 1 (math.slop roadmap) e2e tests
- Create: `sloplang/examples/math.slop`

**Step 1: Add e2e tests**

Append tests to `codegen_e2e_test.go` using the existing `runE2E` helper. Each test writes slop source, expected output string, and asserts equality. Group by category:

**Arithmetic (12 tests):** `TestE2E_Add`, `TestE2E_Sub`, `TestE2E_Mul`, `TestE2E_Div`, `TestE2E_Mod`, `TestE2E_Pow`, `TestE2E_AddSingleElement`, `TestE2E_SubZero`, `TestE2E_MulByZero`, `TestE2E_AddLongArrays`, `TestE2E_AddFloat`, `TestE2E_AddUint`

**Unary (6 tests):** `TestE2E_Negate`, `TestE2E_NegateZero`, `TestE2E_NegateNegative`, `TestE2E_DoubleNegate`, `TestE2E_NegateFloat`, `TestE2E_NotFalsy`

**Not (4 tests):** `TestE2E_NotTruthy`, `TestE2E_NotZeroIsTruthy`, `TestE2E_DoubleNotTruthy`, `TestE2E_DoubleNotFalsy`

**Comparisons (14 tests):** `TestE2E_EqTrue`, `TestE2E_EqFalse`, `TestE2E_NeqTrue`, `TestE2E_NeqFalse`, `TestE2E_LtTrue`, `TestE2E_LtFalse`, `TestE2E_GtTrue`, `TestE2E_GtFalse`, `TestE2E_LteEqual`, `TestE2E_LteLess`, `TestE2E_GteEqual`, `TestE2E_GteFail`, `TestE2E_StringEq`, `TestE2E_StringNeq`

**Logical (6 tests):** `TestE2E_AndBothTruthy`, `TestE2E_AndRightFalsy`, `TestE2E_AndLeftFalsy`, `TestE2E_OrLeftFalsy`, `TestE2E_OrLeftTruthy`, `TestE2E_OrBothFalsy`

**Precedence (8 tests):** `TestE2E_PrecedenceMulBeforeAdd`, `TestE2E_PrecedenceMulBeforeAddRight`, `TestE2E_PrecedenceParens`, `TestE2E_PrecedenceLeftAssocSub`, `TestE2E_PrecedencePower`, `TestE2E_PrecedenceArithBeforeComparison`, `TestE2E_PrecedenceComparisonBeforeLogical`, `TestE2E_PrecedenceAndBeforeOr`

**str() (4 tests):** `TestE2E_StrArray`, `TestE2E_StrSingle`, `TestE2E_StrEmpty`, `TestE2E_StrExprResult`

**Chained (4 tests):** `TestE2E_ChainedAdd`, `TestE2E_ChainedMulDiv`, `TestE2E_ChainedLogical`, `TestE2E_UnaryThenBinary`

**Roadmap (1 test):** `TestE2E_MathSlop` — the exact program from the roadmap

Each test uses `|> str(expr)` pattern to print result, asserts exact stdout match. Refer to the design doc [phase2-arithmetic-design.md](docs/plans/phase2-arithmetic-design.md) for exact input/output pairs.

**Step 2: Create `sloplang/examples/math.slop`**

The exact program from the roadmap: assign `x = [10, 20] + [3, 4]`, print, assign `y = [5] > [3]`, print, assign `z = -[1, 2, 3]`, print.

**Step 3: Run all tests**

Run: `cd sloplang && go test ./... -v -timeout 120s`

**Step 4: Commit** — `feat: 58 e2e tests for phase 2 arithmetic, comparisons, logic, str()`

---

### Task 8: Create Phase 2 JSON Tracking File

**Files:**
- Create: `docs/plans/phase2-arithmetic.json`

**Step 1: Create the JSON file**

5 categories: `lexer`, `parser`, `runtime`, `codegen`, `e2e`. Each with description, steps array, and `"passes": false`. Refer to the design doc for exact step descriptions. Follow the schema from `docs/plans/phase1-skeleton.json`.

**Step 2: Commit** — `docs: add phase 2 tracking JSON`

---

### Task 9: Final Verification and Flip Passes

**Step 1: Run full test suite**

Run: `cd sloplang && go test ./... -v -timeout 120s`

**Step 2: Run vet and fmt**

Run: `cd sloplang && go vet ./... && go fmt ./...`

**Step 3: Flip all `"passes"` to `true` in `docs/plans/phase2-arithmetic.json`**

**Step 4: Commit** — `feat: phase 2 complete — all 58 e2e tests pass`
