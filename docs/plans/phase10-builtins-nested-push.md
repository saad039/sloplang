# Phase 10: New Builtins + Nested Push — Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `to_chars`, `to_int`, `to_float`, `fmt_float` builtins, the `<<<` nested push operator, and document pointer-sharing semantics.

**Architecture:** Four new builtins follow the existing pattern (runtime function + codegen builtins map entry). The `<<<` operator needs a new token, AST node, codegen case, and runtime function. Documentation updates cover PRD, book, and appendices.

**Tech Stack:** Go 1.21+, `go/ast` code generation, `strconv`, `fmt`, `math`

**Spec:** `docs/superpowers/specs/2026-03-14-phase10-builtins-nested-push-design.md`

---

## Chunk 1: Builtins

### Task 1: `to_chars` — Runtime + Codegen + Tests

**Files:**
- Modify: `sloplang/pkg/runtime/io.go` (add `ToChars`)
- Modify: `sloplang/pkg/codegen/codegen.go:545` (add to builtins map)
- Modify: `sloplang/pkg/codegen/codegen.go:421` (add to `isDualReturn` — NOT needed, but verify)
- Test: `sloplang/pkg/codegen/phase10_builtins_e2e_test.go` (create)

- [ ] **Step 1: Write the E2E test file with `to_chars` tests**

Create `sloplang/pkg/codegen/phase10_builtins_e2e_test.go`:

```go
package codegen

import "testing"

// to_chars tests
func TestP10_ToChars_ASCII(t *testing.T) {
	runE2E(t, `chars = to_chars("hello")
|> str(chars)
|> "\n"`, "[h, e, l, l, o]\n")
}

func TestP10_ToChars_SingleChar(t *testing.T) {
	runE2E(t, `|> str(to_chars("x"))
|> "\n"`, "[x]\n")
}

func TestP10_ToChars_EmptyString(t *testing.T) {
	runE2E(t, `|> str(to_chars(""))
|> "\n"`, "[]\n")
}

func TestP10_ToChars_Unicode(t *testing.T) {
	runE2E(t, `|> str(to_chars("héllo"))
|> "\n"`, "[h, é, l, l, o]\n")
}

func TestP10_ToChars_FromVariable(t *testing.T) {
	runE2E(t, `s = "abc"
|> str(to_chars(s))
|> "\n"`, "[a, b, c]\n")
}

func TestP10_ToChars_PanicOnInt(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_chars([42])`, "to_chars requires a string argument")
}

func TestP10_ToChars_PanicOnNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_chars([null])`, "to_chars requires a string argument")
}

func TestP10_ToChars_PanicOnMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_chars("a" ++ "b")`, "to_chars requires a string argument")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ToChars -v -count=1`
Expected: FAIL (ToChars not defined)

- [ ] **Step 3: Implement `ToChars` in runtime**

Add to `sloplang/pkg/runtime/io.go` after `ToNum`:

```go
// ToChars splits a string into an array of single-character strings.
// Panics if the argument is not a single-element string.
func ToChars(sv *SlopValue) *SlopValue {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: to_chars requires a string argument, got %d elements", len(sv.Elements)))
	}
	s, ok := sv.Elements[0].(string)
	if !ok {
		panic(fmt.Sprintf("sloplang: to_chars requires a string argument, got %T", sv.Elements[0]))
	}
	runes := []rune(s)
	elems := make([]any, len(runes))
	for i, r := range runes {
		elems[i] = string(r)
	}
	return &SlopValue{Elements: elems}
}
```

- [ ] **Step 4: Add `to_chars` to codegen builtins map**

In `sloplang/pkg/codegen/codegen.go`, line ~545, change:

```go
builtins := map[string]string{"str": "Str", "split": "Split", "to_num": "ToNum", "exit": "Exit"}
```

to:

```go
builtins := map[string]string{"str": "Str", "split": "Split", "to_num": "ToNum", "exit": "Exit", "to_chars": "ToChars"}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ToChars -v -count=1`
Expected: PASS (all 8 tests)

- [ ] **Step 6: Commit**

```bash
git add sloplang/pkg/runtime/io.go sloplang/pkg/codegen/codegen.go sloplang/pkg/codegen/phase10_builtins_e2e_test.go
git commit -m "feat: add to_chars() builtin — string to character array"
```

---

### Task 2: `to_int` — Runtime + Codegen + Tests

**Files:**
- Modify: `sloplang/pkg/runtime/io.go` (add `ToInt`)
- Modify: `sloplang/pkg/codegen/codegen.go:545` (add to builtins map)
- Test: `sloplang/pkg/codegen/phase10_builtins_e2e_test.go` (append)

- [ ] **Step 1: Write E2E tests for `to_int`**

Append to `sloplang/pkg/codegen/phase10_builtins_e2e_test.go`:

```go
// to_int tests
func TestP10_ToInt_FromFloat(t *testing.T) {
	runE2E(t, `|> str(to_int([3.14]))
|> "\n"`, "[3]\n")
}

func TestP10_ToInt_FromNegativeFloat(t *testing.T) {
	runE2E(t, `|> str(to_int([-2.9]))
|> "\n"`, "[-2]\n")
}

func TestP10_ToInt_FromInt(t *testing.T) {
	runE2E(t, `|> str(to_int([42]))
|> "\n"`, "[42]\n")
}

func TestP10_ToInt_FromString(t *testing.T) {
	runE2E(t, `|> str(to_int("5"))
|> "\n"`, "[5]\n")
}

func TestP10_ToInt_FromZeroFloat(t *testing.T) {
	runE2E(t, `|> str(to_int([0.0]))
|> "\n"`, "[0]\n")
}

func TestP10_ToInt_PanicOnInvalidString(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_int("abc")`, "to_int: cannot convert")
}

func TestP10_ToInt_PanicOnNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_int([null])`, "to_int: cannot convert")
}

func TestP10_ToInt_PanicOnMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_int([1, 2])`, "to_int: requires single-element")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ToInt -v -count=1`
Expected: FAIL

- [ ] **Step 3: Implement `ToInt` in runtime**

Add to `sloplang/pkg/runtime/io.go`:

```go
// ToInt converts a single-element numeric or string value to int64.
// float64 is truncated toward zero. Panics on invalid input.
func ToInt(sv *SlopValue) *SlopValue {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: to_int: requires single-element array, got %d elements", len(sv.Elements)))
	}
	switch v := sv.Elements[0].(type) {
	case int64:
		return NewSlopValue(v)
	case float64:
		return NewSlopValue(int64(v))
	case uint64:
		if v > uint64(math.MaxInt64) {
			panic("sloplang: to_int: uint64 value exceeds MaxInt64")
		}
		return NewSlopValue(int64(v))
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return NewSlopValue(i)
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return NewSlopValue(int64(f))
		}
		panic(fmt.Sprintf("sloplang: to_int: cannot convert string %q to int", v))
	default:
		panic(fmt.Sprintf("sloplang: to_int: cannot convert %T to int", v))
	}
}
```

Also add `"math"` to the import block in `io.go` if not already present.

- [ ] **Step 4: Add `to_int` to codegen builtins map**

Update the builtins map to include `"to_int": "ToInt"`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ToInt -v -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add sloplang/pkg/runtime/io.go sloplang/pkg/codegen/codegen.go sloplang/pkg/codegen/phase10_builtins_e2e_test.go
git commit -m "feat: add to_int() builtin — type casting to int64"
```

---

### Task 3: `to_float` — Runtime + Codegen + Tests

**Files:**
- Modify: `sloplang/pkg/runtime/io.go` (add `ToFloat`)
- Modify: `sloplang/pkg/codegen/codegen.go:545` (add to builtins map)
- Test: `sloplang/pkg/codegen/phase10_builtins_e2e_test.go` (append)

- [ ] **Step 1: Write E2E tests for `to_float`**

Append to `sloplang/pkg/codegen/phase10_builtins_e2e_test.go`:

```go
// to_float tests
func TestP10_ToFloat_FromInt(t *testing.T) {
	runE2E(t, `|> str(to_float([42]))
|> "\n"`, "[42]\n")
}

func TestP10_ToFloat_FromFloat(t *testing.T) {
	runE2E(t, `|> str(to_float([3.14]))
|> "\n"`, "[3.14]\n")
}

func TestP10_ToFloat_FromString(t *testing.T) {
	runE2E(t, `|> str(to_float("2.5"))
|> "\n"`, "[2.5]\n")
}

func TestP10_ToFloat_FromZero(t *testing.T) {
	runE2E(t, `|> str(to_float([0]))
|> "\n"`, "[0]\n")
}

func TestP10_ToFloat_ArithmeticAfterConversion(t *testing.T) {
	runE2E(t, `a = to_float([3])
b = [1.5]
|> str(a + b)
|> "\n"`, "[4.5]\n")
}

func TestP10_ToFloat_PanicOnInvalidString(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_float("abc")`, "to_float: cannot convert")
}

func TestP10_ToFloat_PanicOnNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_float([null])`, "to_float: cannot convert")
}

func TestP10_ToFloat_PanicOnMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_float([1, 2])`, "to_float: requires single-element")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ToFloat -v -count=1`
Expected: FAIL

- [ ] **Step 3: Implement `ToFloat` in runtime**

Add to `sloplang/pkg/runtime/io.go`:

```go
// ToFloat converts a single-element numeric or string value to float64.
// Panics on invalid input.
func ToFloat(sv *SlopValue) *SlopValue {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: to_float: requires single-element array, got %d elements", len(sv.Elements)))
	}
	switch v := sv.Elements[0].(type) {
	case float64:
		return NewSlopValue(v)
	case int64:
		return NewSlopValue(float64(v))
	case uint64:
		return NewSlopValue(float64(v))
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return NewSlopValue(f)
		}
		panic(fmt.Sprintf("sloplang: to_float: cannot convert string %q to float", v))
	default:
		panic(fmt.Sprintf("sloplang: to_float: cannot convert %T to float", v))
	}
}
```

- [ ] **Step 4: Add `to_float` to codegen builtins map**

Update the builtins map to include `"to_float": "ToFloat"`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ToFloat -v -count=1`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add sloplang/pkg/runtime/io.go sloplang/pkg/codegen/codegen.go sloplang/pkg/codegen/phase10_builtins_e2e_test.go
git commit -m "feat: add to_float() builtin — type casting to float64"
```

---

### Task 4: `fmt_float` — Runtime + Codegen + Tests

**Files:**
- Modify: `sloplang/pkg/runtime/io.go` (add `FmtFloat`)
- Modify: `sloplang/pkg/codegen/codegen.go:545` (add to builtins map)
- Test: `sloplang/pkg/codegen/phase10_builtins_e2e_test.go` (append)

- [ ] **Step 1: Write E2E tests for `fmt_float`**

Append to `sloplang/pkg/codegen/phase10_builtins_e2e_test.go`:

```go
// fmt_float tests
func TestP10_FmtFloat_TwoDecimals(t *testing.T) {
	runE2E(t, `|> fmt_float([3.14159], [2])
|> "\n"`, "3.14\n")
}

func TestP10_FmtFloat_ThreeDecimals(t *testing.T) {
	runE2E(t, `|> fmt_float([42], [3])
|> "\n"`, "42.000\n")
}

func TestP10_FmtFloat_ZeroDecimals(t *testing.T) {
	runE2E(t, `|> fmt_float([1.0], [0])
|> "\n"`, "1\n")
}

func TestP10_FmtFloat_FiveDecimals(t *testing.T) {
	runE2E(t, `|> fmt_float([0.1], [5])
|> "\n"`, "0.10000\n")
}

func TestP10_FmtFloat_IntPromoted(t *testing.T) {
	runE2E(t, `|> fmt_float([100], [2])
|> "\n"`, "100.00\n")
}

func TestP10_FmtFloat_Negative(t *testing.T) {
	runE2E(t, `|> fmt_float([-3.14], [1])
|> "\n"`, "-3.1\n")
}

func TestP10_FmtFloat_PanicOnString(t *testing.T) {
	runE2EExpectPanicContaining(t, `fmt_float("abc", [2])`, "fmt_float: first argument must be numeric")
}

func TestP10_FmtFloat_PanicOnNegativeDecimals(t *testing.T) {
	runE2EExpectPanicContaining(t, `fmt_float([3.14], [-1])`, "fmt_float: second argument must be non-negative integer")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_FmtFloat -v -count=1`
Expected: FAIL

- [ ] **Step 3: Implement `FmtFloat` in runtime**

Add to `sloplang/pkg/runtime/io.go`:

```go
// FmtFloat formats a numeric value with a fixed number of decimal places.
// Returns a string. Panics if first arg is not numeric or second is not a non-negative int.
func FmtFloat(val, decimals *SlopValue) *SlopValue {
	if len(val.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: fmt_float: first argument must be numeric, got %d elements", len(val.Elements)))
	}
	if len(decimals.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: fmt_float: second argument must be non-negative integer, got %d elements", len(decimals.Elements)))
	}
	d, ok := decimals.Elements[0].(int64)
	if !ok || d < 0 {
		panic("sloplang: fmt_float: second argument must be non-negative integer")
	}
	var f float64
	switch v := val.Elements[0].(type) {
	case float64:
		f = v
	case int64:
		f = float64(v)
	case uint64:
		f = float64(v)
	default:
		panic(fmt.Sprintf("sloplang: fmt_float: first argument must be numeric, got %T", v))
	}
	result := fmt.Sprintf("%.*f", int(d), f)
	return NewSlopValue(result)
}
```

- [ ] **Step 4: Add `fmt_float` to codegen builtins map**

Update the builtins map to include `"fmt_float": "FmtFloat"`.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_FmtFloat -v -count=1`
Expected: PASS

- [ ] **Step 6: Run all builtin tests together**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_ -v -count=1`
Expected: All 32 tests PASS

- [ ] **Step 7: Commit**

```bash
git add sloplang/pkg/runtime/io.go sloplang/pkg/codegen/codegen.go sloplang/pkg/codegen/phase10_builtins_e2e_test.go
git commit -m "feat: add fmt_float() builtin — float formatting with fixed decimals"
```

---

## Chunk 2: Nested Push Operator

### Task 5: `<<<` — Lexer token

**Files:**
- Modify: `sloplang/pkg/lexer/token.go` (add `TOKEN_NEST_PUSH`)
- Modify: `sloplang/pkg/lexer/lexer.go:169-173` (extend `<<` branch for `<<<`)
- Test: `sloplang/pkg/lexer/lexer_test.go` (add test)

- [ ] **Step 1: Write lexer test for `<<<`**

Add to `sloplang/pkg/lexer/lexer_test.go`:

```go
func TestLexer_NestPush(t *testing.T) {
	tokens := tokenize(t, `arr <<< [5]`)
	expectTokens(t, tokens, []TokenExpect{
		{TOKEN_IDENT, "arr"}, {TOKEN_NEST_PUSH, "<<<"}, {TOKEN_LBRACKET, "["}, {TOKEN_INT, "5"}, {TOKEN_RBRACKET, "]"},
	})
}

func TestLexer_NestPushVsLshift(t *testing.T) {
	tokens := tokenize(t, `<< <<<`)
	expectTokens(t, tokens, []TokenExpect{
		{TOKEN_LSHIFT, "<<"}, {TOKEN_NEST_PUSH, "<<<"},
	})
}

func TestLexer_LshiftStillWorks(t *testing.T) {
	tokens := tokenize(t, `arr << [5]`)
	expectTokens(t, tokens, []TokenExpect{
		{TOKEN_IDENT, "arr"}, {TOKEN_LSHIFT, "<<"}, {TOKEN_LBRACKET, "["}, {TOKEN_INT, "5"}, {TOKEN_RBRACKET, "]"},
	})
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd sloplang && go test ./pkg/lexer/ -run "TestLexer_NestPush" -v -count=1`
Expected: FAIL (TOKEN_NEST_PUSH not defined)

- [ ] **Step 3: Add `TOKEN_NEST_PUSH` to token.go**

In `sloplang/pkg/lexer/token.go`, add after `TOKEN_DOLLAR`:

```go
TOKEN_NEST_PUSH // <<<
```

Add to `tokenNames` map:

```go
TOKEN_NEST_PUSH: "NEST_PUSH",
```

- [ ] **Step 4: Modify lexer `<` branch to handle `<<<`**

In `sloplang/pkg/lexer/lexer.go`, replace the `<<` branch (lines 169-173):

```go
} else if l.peekChar() == '<' {
	tok.Type = TOKEN_LSHIFT
	tok.Literal = "<<"
	l.readChar()
	l.readChar()
}
```

with:

```go
} else if l.peekChar() == '<' {
	l.readChar() // consume first <
	l.readChar() // consume second <, l.ch is now char after <<
	if l.ch == '<' {
		tok.Type = TOKEN_NEST_PUSH
		tok.Literal = "<<<"
		l.readChar() // consume third <
	} else {
		tok.Type = TOKEN_LSHIFT
		tok.Literal = "<<"
	}
}
```

- [ ] **Step 5: Run lexer tests to verify they pass**

Run: `cd sloplang && go test ./pkg/lexer/ -v -count=1`
Expected: ALL PASS (new tests + all existing tests including LSHIFT tests)

- [ ] **Step 6: Commit**

```bash
git add sloplang/pkg/lexer/token.go sloplang/pkg/lexer/lexer.go sloplang/pkg/lexer/lexer_test.go
git commit -m "feat: add TOKEN_NEST_PUSH (<<<) to lexer"
```

---

### Task 6: `<<<` — Parser AST node

**Files:**
- Modify: `sloplang/pkg/parser/ast.go` (add `NestPushStmt`)
- Modify: `sloplang/pkg/parser/parser.go:57-58` (add `TOKEN_NEST_PUSH` branch)

- [ ] **Step 1: Add `NestPushStmt` to ast.go**

Add after `PushStmt` (line ~228) in `sloplang/pkg/parser/ast.go`:

```go
// NestPushStmt represents: object <<< value (nested push)
type NestPushStmt struct {
	Object Expr
	Value  Expr
}

func (ns *NestPushStmt) stmtNode()            {}
func (ns *NestPushStmt) TokenLiteral() string { return "<<<" }
```

- [ ] **Step 2: Add `parseNestPushStmt` and wire it in parser.go**

In `sloplang/pkg/parser/parser.go`, add after the `parsePushStmt` function:

```go
func (p *Parser) parseNestPushStmt() *NestPushStmt {
	name := p.curToken().Literal
	p.advance() // consume ident
	p.advance() // consume <<<
	value := p.parseExpression()
	if value == nil {
		return nil
	}
	return &NestPushStmt{
		Object: &Identifier{Name: name},
		Value:  value,
	}
}
```

In the `TOKEN_IDENT` case of `parseStatement()`, add after the `TOKEN_LSHIFT` check (line ~58):

```go
if p.peekToken().Type == lexer.TOKEN_NEST_PUSH {
	return p.parseNestPushStmt()
}
```

- [ ] **Step 3: Run parser tests to make sure nothing breaks**

Run: `cd sloplang && go test ./pkg/parser/ -v -count=1`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add sloplang/pkg/parser/ast.go sloplang/pkg/parser/parser.go
git commit -m "feat: add NestPushStmt AST node and parser support for <<<"
```

---

### Task 7: `<<<` — Codegen + Runtime + E2E Tests

**Files:**
- Modify: `sloplang/pkg/runtime/ops.go` (add `NestPush`)
- Modify: `sloplang/pkg/codegen/codegen.go:238-241` (add `NestPushStmt` case)
- Test: `sloplang/pkg/codegen/phase10_builtins_e2e_test.go` (append)

- [ ] **Step 1: Write E2E tests for `<<<`**

Append to `sloplang/pkg/codegen/phase10_builtins_e2e_test.go`:

```go
// <<< (nested push) tests
func TestP10_NestPush_MultiElement(t *testing.T) {
	runE2E(t, `arr = [1, 2]
arr <<< [3, 4]
|> str(arr)
|> "\n"`, "[1, 2, [3, 4]]\n")
}

func TestP10_NestPush_SingleElement(t *testing.T) {
	runE2E(t, `arr = [1, 2]
arr <<< [42]
|> str(arr)
|> "\n"`, "[1, 2, [42]]\n")
}

func TestP10_NestPush_String(t *testing.T) {
	runE2E(t, `arr = [1, 2]
arr <<< "hello"
|> str(arr)
|> "\n"`, "[1, 2, [hello]]\n")
}

func TestP10_NestPush_StringArray(t *testing.T) {
	runE2E(t, `arr = [1, 2]
arr <<< ["saad", "is"]
|> str(arr)
|> "\n"`, "[1, 2, [saad, is]]\n")
}

func TestP10_NestPush_EmptyIntoEmpty(t *testing.T) {
	runE2E(t, `arr = []
arr <<< [1]
|> str(arr)
|> "\n"`, "[[1]]\n")
}

func TestP10_NestPush_Multiple(t *testing.T) {
	runE2E(t, `arr = [1, 2]
arr <<< [3, 4]
arr <<< [42]
arr <<< "hello"
|> str(arr)
|> "\n"`, "[1, 2, [3, 4], [42], [hello]]\n")
}

func TestP10_NestPush_SpreadVsNest(t *testing.T) {
	// Verify << still spreads and <<< nests
	runE2E(t, `a = [1]
a << [2, 3]
|> str(a)
|> "\n"
b = [1]
b <<< [2, 3]
|> str(b)
|> "\n"`, "[1, 2, 3]\n[1, [2, 3]]\n")
}

func TestP10_NestPush_EmptyArray(t *testing.T) {
	runE2E(t, `arr = [1]
arr <<< []
|> str(arr)
|> "\n"`, "[1, []]\n")
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_NestPush -v -count=1`
Expected: FAIL (NestPush not defined)

- [ ] **Step 3: Implement `NestPush` in runtime**

Add to `sloplang/pkg/runtime/ops.go` after `Push`:

```go
// NestPush appends val as a single nested element to sv. Mutates sv.
// Unlike Push which spreads elements, NestPush always nests.
func NestPush(sv *SlopValue, val *SlopValue) *SlopValue {
	sv.Elements = append(sv.Elements, val)
	return sv
}
```

- [ ] **Step 4: Add `NestPushStmt` case to codegen**

In `sloplang/pkg/codegen/codegen.go`, add after the `PushStmt` case (line ~241):

```go
case *parser.NestPushStmt:
	return []ast.Stmt{
		&ast.ExprStmt{X: callRuntime("NestPush", g.lowerExpr(s.Object), g.lowerExpr(s.Value))},
	}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd sloplang && go test ./pkg/codegen/ -run TestP10_NestPush -v -count=1`
Expected: ALL PASS (8 tests)

- [ ] **Step 6: Run full test suite**

Run: `cd sloplang && go test ./... -count=1`
Expected: ALL PASS (1,450+ existing + ~40 new = ~1,490+ tests, 0 failures)

- [ ] **Step 7: Commit**

```bash
git add sloplang/pkg/runtime/ops.go sloplang/pkg/codegen/codegen.go sloplang/pkg/codegen/phase10_builtins_e2e_test.go
git commit -m "feat: add <<< (nested push) operator — appends value as single nested element"
```

---

## Chunk 3: Documentation Updates

### Task 8: Update PRD

**Files:**
- Modify: `docs/PRD.md`

- [ ] **Step 1: Add `<<<` to operator table**

Find the operator table in `docs/PRD.md` and add after the `<<` row:

```
| `<<<` | Nested push | `arr <<< [5]` (appends as nested element) |
```

- [ ] **Step 2: Add new builtins to builtins section**

Find the builtins section in `docs/PRD.md` and add entries for `to_chars`, `to_int`, `to_float`, `fmt_float`.

- [ ] **Step 3: Commit**

```bash
git add docs/PRD.md
git commit -m "docs: add phase 10 features to PRD — new builtins + <<< operator"
```

---

### Task 9: Update Book — Appendix A (Operators) + Appendix B (Builtins)

**Files:**
- Modify: `book/appendix-a-operators.md`
- Modify: `book/appendix-b-builtins.md`

- [ ] **Step 1: Add `<<<` to Appendix A operator table**

Add after the `<<` row in `book/appendix-a-operators.md`:

```
| `<<<` | Nested push | `arr <<< val` | `arr <<< [5]` | appends `[5]` as nested element | Yes |
```

Add to notes section:

```
- `<<<` always nests: `arr <<< [3, 4]` appends `[3, 4]` as one element, producing `[..., [3, 4]]`. Contrast with `<<` which spreads.
```

- [ ] **Step 2: Add new builtins to Appendix B**

Add sections for `to_chars`, `to_int`, `to_float`, `fmt_float` to `book/appendix-b-builtins.md`, following the existing format (signature, description, examples table, notes).

- [ ] **Step 3: Commit**

```bash
git add book/appendix-a-operators.md book/appendix-b-builtins.md
git commit -m "docs: update book appendices with phase 10 builtins and <<< operator"
```

---

### Task 10: Update Book — ch06-arrays.md (pointer-sharing) + ch12-transpiler.md

**Files:**
- Modify: `book/ch06-arrays.md`
- Modify: `book/ch12-transpiler.md`

- [ ] **Step 1: Add pointer-sharing section to ch06-arrays.md**

Add a new section (e.g., 6.9) titled "Aliasing and Pointer Sharing" that documents:
- Extracting a nested array via `@` or `$` gives a reference to the same `*SlopValue`
- Mutating the extracted value with `<<`, `>>`, `~@`, or `<<<` also mutates the original
- Include the `matrix = [[1, 2], [3, 4]]; row = matrix@0; row << [99]` example

- [ ] **Step 2: Add `<<<` section to ch06-arrays.md**

Add documentation for `<<<` in the mutating arrays section (6.3), alongside `<<`.

- [ ] **Step 3: Fix incorrect statement in ch12-transpiler.md**

At line ~315, change the statement "arrays are values and mutations are returned, not applied in place behind a reference" to accurately describe the runtime model: mutating operators (`<<`, `>>`, `~@`, `<<<`) modify the `*SlopValue` in place, and extracted sub-arrays share the same pointer.

- [ ] **Step 4: Add `<<<` to ch12-transpiler.md runtime section**

Add `NestPush` documentation alongside the existing `Push` documentation (~line 740-748).

- [ ] **Step 5: Commit**

```bash
git add book/ch06-arrays.md book/ch12-transpiler.md
git commit -m "docs: document pointer-sharing semantics and <<< in book chapters 6 and 12"
```

---

### Task 11: Update README + architecture docs

**Files:**
- Modify: `README.md`
- Modify: `docs/architecture.md`
- Modify: `docs/patterns.md`

- [ ] **Step 1: Add new features to README operator table and language features**

Add `<<<` to the operators table in `README.md`. Add `to_chars`, `to_int`, `to_float`, `fmt_float` mention to the Language Features section.

- [ ] **Step 2: Update architecture.md**

Add phase 10 to the architecture doc with new tokens, AST nodes, runtime functions, and builtins.

- [ ] **Step 3: Update patterns.md if any lessons arise**

Record any patterns/lessons learned during implementation.

- [ ] **Step 4: Commit**

```bash
git add README.md docs/architecture.md docs/patterns.md
git commit -m "docs: update README, architecture, and patterns for phase 10"
```

---

### Task 12: Update plan JSON and verify

**Files:**
- Modify: `docs/plans/phase10-builtins-nested-push.json`

- [ ] **Step 1: Flip all `passes` to `true` in the JSON**

After verifying all tests pass and all documentation is updated, update each entry's `passes` field from `false` to `true`.

- [ ] **Step 2: Final full test run**

Run: `cd sloplang && go test ./... -count=1`
Expected: ALL PASS, 0 failures, 0 skipped

- [ ] **Step 3: Commit**

```bash
git add docs/plans/phase10-builtins-nested-push.json
git commit -m "feat: phase 10 complete — mark all tasks as passing"
```
