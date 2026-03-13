# Phase 9: Adversarial Bug Fixes

## Overview

Fix all 15 bugs discovered by the adversarial test suite (currently `t.Skip("BUG: ...")`). These span 3 subsystems: lexer (5 tests), runtime (5 tests), and codegen (4 tests), plus 1 use-before-assign SIGSEGV.

## Bug Inventory

| # | Subsystem | Bug | Skipped Tests | File:Line |
|---|-----------|-----|---------------|-----------|
| 1 | Lexer | Unclosed strings consumed to EOF without error | 3 | `lexer.go:367` |
| 2 | Lexer | Scientific notation (`1.79e308`) not supported | 2 | `lexer.go:391-419` |
| 3 | Runtime | `Mod()` — no zero check for int64/uint64 | 2 | `ops.go:154,160` |
| 4 | Runtime | `Div()` — MinInt64 / -1 overflow (no check) | 1 | `ops.go:121` |
| 5 | Runtime | `Negate()` — -MinInt64 overflow (no check) | 1 | `ops.go:201` |
| 6 | Codegen | Go keywords (`func`, `var`, `return`, `range`) as var names crash `format.Source()` | 4 | `codegen.go:52` |

---

## Task 1: Lexer — Reject Unclosed Strings

**Files:**
- Modify: `sloplang/pkg/lexer/lexer.go` (`readString` method, ~line 364)
- Modify: `sloplang/pkg/codegen/adversarial_strings_e2e_test.go` (remove 2 `t.Skip`)
- Modify: `sloplang/pkg/codegen/adversarial_syntax_e2e_test.go` (remove 1 `t.Skip`)

**Root cause:** `readString()` loop exits on `l.ch == 0` (EOF) silently, returning whatever was accumulated. No error token emitted.

**Fix:** After the loop, check if `l.ch == 0`. If so, return an error token (or track an error). The simplest approach: add a bool return or check after `readString()` returns — if the current char after the call is not past the closing `"`, the string was unclosed.

**Implementation hint:** After the `for` loop (line 386), before the closing-quote readChar (line 387), check `if l.ch == 0` and emit `TOKEN_ILLEGAL` with a "unterminated string" message. This matches how the parser already handles `TOKEN_ILLEGAL`.

**Steps:**
1. In `readString()`, after the loop, check `l.ch != '"'`. If so, change the returned token to `TOKEN_ILLEGAL` with literal `"unterminated string"`.
2. The caller in `Tokenize()` that calls `readString()` needs to handle this — either by checking the returned value or by refactoring `readString` to return a `Token` instead of a `string`.
3. Remove `t.Skip(...)` from `TestAdv_Str_UnclosedStringEOF`, `TestAdv_Str_UnclosedStringBackslash`, `TestAdv_Syn_UnclosedString`.
4. Run: `go test ./pkg/lexer/... && go test ./pkg/codegen/... -run "TestAdv_Str_Unclosed|TestAdv_Syn_UnclosedString"`

---

## Task 2: Lexer — Support Scientific Notation

**Files:**
- Modify: `sloplang/pkg/lexer/lexer.go` (`readNumber` method, ~line 391)
- Modify: `sloplang/pkg/codegen/adversarial_boundaries_e2e_test.go` (remove 2 `t.Skip`)

**Root cause:** `readNumber()` handles digits, `.`, and `u` suffix but never checks for `e`/`E`. When the lexer hits `e`, it stops reading the number and treats `e308` as a separate identifier token.

**Fix:** After reading the decimal part (or integer part), check for `e`/`E`. If found, consume it, then optionally consume `+`/`-`, then consume digits. Mark as `TOKEN_FLOAT`.

**Implementation hint:** After the float branch (line 406) and before the uint check (line 409), add: if `l.ch == 'e' || l.ch == 'E'`, consume exponent. Also need to handle the case where `e` appears after an integer (e.g., `5e10` is valid float).

**Steps:**
1. In `readNumber()`, after consuming digits and optional decimal part, add exponent parsing: consume `e`/`E`, optional `+`/`-`, then digits. Set `tok.Type = TOKEN_FLOAT`.
2. Remove `t.Skip(...)` from `TestAdv_Bound_FloatMaxVal` and `TestAdv_Bound_FloatSmallestPos`.
3. Fix expected output in `TestAdv_Bound_FloatSmallestPos` if needed — Go's `fmt.Sprintf` for `5e-324` may format differently.
4. Run: `go test ./pkg/lexer/... && go test ./pkg/codegen/... -run "TestAdv_Bound_Float"`

---

## Task 3: Runtime — Mod-by-Zero Check

**Files:**
- Modify: `sloplang/pkg/runtime/ops.go` (`Mod` function, ~line 146)
- Modify: `sloplang/pkg/codegen/adversarial_boundaries_e2e_test.go` (remove 2 `t.Skip`)

**Root cause:** `Mod()` has no zero check. `xv % yv` where `yv == 0` causes a raw Go panic ("integer divide by zero") instead of a clean sloplang error.

**Fix:** Add `if yv == 0 { panic("sloplang: modulo by zero") }` before both the int64 case (line 154) and uint64 case (line 160). Mirror the pattern from `Div()`.

**Steps:**
1. Add zero check before `return xv % yv` in both int64 and uint64 cases.
2. Remove `t.Skip(...)` from `TestAdv_Bound_ModZeroInt` and `TestAdv_Bound_ModZeroUint`.
3. Run: `go test ./pkg/runtime/... && go test ./pkg/codegen/... -run "TestAdv_Bound_ModZero"`

---

## Task 4: Runtime — MinInt64 / -1 Overflow Check

**Files:**
- Modify: `sloplang/pkg/runtime/ops.go` (`Div` function int64 case, ~line 121)
- Modify: `sloplang/pkg/codegen/adversarial_boundaries_e2e_test.go` (remove 1 `t.Skip`)

**Root cause:** `math.MinInt64 / -1` overflows int64 (result would be `math.MaxInt64 + 1`). Go produces a runtime panic. Need a clean sloplang error.

**Fix:** After the existing zero check (line 118-120), add: `if xv == math.MinInt64 && yv == -1 { panic("sloplang: integer overflow: MinInt64 / -1") }`. Need `import "math"` (likely already imported for `Pow`).

**Steps:**
1. Add overflow check in `Div()` int64 case after the zero check.
2. Remove `t.Skip(...)` from `TestAdv_Bound_MinIntDivNegOne`.
3. Run: `go test ./pkg/runtime/... && go test ./pkg/codegen/... -run "TestAdv_Bound_MinIntDivNegOne"`

---

## Task 5: Runtime — Negate MinInt64 Overflow Check

**Files:**
- Modify: `sloplang/pkg/runtime/ops.go` (`Negate` function int64 case, ~line 200)
- Modify: `sloplang/pkg/codegen/adversarial_boundaries_e2e_test.go` (remove 1 `t.Skip`)

**Root cause:** `-math.MinInt64` overflows int64 (wraps to MinInt64 again). Go doesn't panic but produces wrong result. Need a clean sloplang error.

**Fix:** In the int64 case of `Negate()`, add: `if e == math.MinInt64 { panic("sloplang: cannot negate MinInt64: integer overflow") }`.

**Steps:**
1. Add overflow check in `Negate()` int64 case before `result[i] = -e`.
2. Remove `t.Skip(...)` from `TestAdv_Bound_MinIntNegate`.
3. Run: `go test ./pkg/runtime/... && go test ./pkg/codegen/... -run "TestAdv_Bound_MinIntNegate"`

---

## Task 6: Codegen — Sanitize Go Reserved Words

**Files:**
- Modify: `sloplang/pkg/codegen/codegen.go` (add identifier sanitization)
- Modify: `sloplang/pkg/codegen/adversarial_identifiers_e2e_test.go` (remove 4 `t.Skip`, change to parse error or success tests)

**Root cause:** When a sloplang variable uses a Go keyword (`func`, `var`, `return`, `range`, etc.), `ast.NewIdent(name)` emits it verbatim. `format.Source()` then crashes because the Go AST contains invalid identifiers.

**Two possible approaches:**

**Approach A — Codegen-level prefix:** Add a `sanitizeIdent(name string) string` function that checks against Go's 25 keywords and prefixes with `slop_` (e.g., `func` → `slop_func`). Apply it everywhere `ast.NewIdent` is called for user-defined names (variable declarations, function names, parameters, for-loop variables). This is transparent to the user.

**Approach B — Parser-level rejection:** Add Go keywords to the parser's reserved word check so users get a parse error like `"func" is a reserved word`. This is simpler but more restrictive.

**Recommended:** Approach A (codegen prefix). It's more permissive — users can name variables `func` or `range` in sloplang without knowing Go exists. The language shouldn't leak its implementation language's restrictions.

**Steps:**
1. Add a `goKeywords` set in codegen containing Go's 25 keywords: `break`, `case`, `chan`, `const`, `continue`, `default`, `defer`, `else`, `fallthrough`, `for`, `func`, `go`, `goto`, `if`, `import`, `interface`, `map`, `package`, `range`, `return`, `select`, `struct`, `switch`, `type`, `var`.
2. Add `sanitizeIdent(name string) string` — if name is in `goKeywords`, return `"slop_" + name`.
3. Apply `sanitizeIdent` at all `ast.NewIdent` call sites that emit user-chosen names (global vars, local vars, function names, function params, for-in loop vars).
4. Remove `t.Skip(...)` from `TestAdv_Id_GoKeyword_Func`, `_Var`, `_Return`, `_Range`.
5. Update these 4 tests to expect success (assign, print, verify output).
6. Run: `go test ./pkg/codegen/... -run "TestAdv_Id_GoKeyword"`
7. Run full suite: `go test ./...`

---

## Verification

After all 6 tasks:
1. `go test ./...` — all tests pass, 0 skipped BUG tests remaining (non-bug skips are fine).
2. `go vet ./...` — no issues.
3. `go fmt ./...` — no changes.

## Commit Plan

One commit per task (6 commits), or group by subsystem:
- `fix(lexer): reject unclosed strings at EOF`
- `feat(lexer): support scientific notation in float literals`
- `fix(runtime): add modulo-by-zero check`
- `fix(runtime): detect MinInt64/-1 overflow in division`
- `fix(runtime): detect MinInt64 overflow in negation`
- `fix(codegen): sanitize Go keywords in user identifiers`
