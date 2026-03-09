# Pre-Phase-8 Syntax Refactor Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enforce bracket-only literals (reject bare numbers/booleans/null), unify variable access under `$var` syntax (replacing `@(expr)` and `@$var`), and enforce strict boolean expressions (only `[1]` and `[]` valid in boolean context).

**Architecture:** Parser changes (reject bare literals in `parsePrimary`, add `$` postfix parsing), runtime changes (`DynAccess`/`DynAccessSet` for type-dispatched access, strict `IsTruthy` that panics on non-`[1]`/`[]`), codegen updates, and mass test updates.

**Tech Stack:** Go (parser, codegen, runtime, tests)

---

### Task 1: Runtime — Add `DynAccess` and `DynAccessSet`

**Files:**
- Modify: `pkg/runtime/ops.go`
- Test: `pkg/runtime/ops_test.go`

**Step 1: Write failing tests for `DynAccess`**

Add to `pkg/runtime/ops_test.go`:

```go
func TestDynAccess_IntKey(t *testing.T) {
	arr := NewSlopValue(int64(10), int64(20), int64(30))
	idx := NewSlopValue(int64(1))
	result := DynAccess(arr, idx)
	if result.Elements[0] != int64(20) {
		t.Fatalf("expected 20, got %v", result.Elements[0])
	}
}

func TestDynAccess_StringKey(t *testing.T) {
	sv := MapFromKeysValues([]string{"name", "age"}, NewSlopValue("bob", int64(30)))
	key := NewSlopValue("name")
	result := DynAccess(sv, key)
	if result.Elements[0] != "bob" {
		t.Fatalf("expected bob, got %v", result.Elements[0])
	}
}

func TestDynAccessSet_IntKey(t *testing.T) {
	arr := NewSlopValue(int64(10), int64(20), int64(30))
	idx := NewSlopValue(int64(1))
	val := NewSlopValue(int64(99))
	DynAccessSet(arr, idx, val)
	if nested, ok := arr.Elements[1].(*SlopValue); ok {
		if nested.Elements[0] != int64(99) {
			t.Fatalf("expected 99, got %v", nested.Elements[0])
		}
	} else {
		t.Fatalf("expected *SlopValue at index 1")
	}
}

func TestDynAccessSet_StringKey(t *testing.T) {
	sv := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	key := NewSlopValue("age")
	val := NewSlopValue(int64(30))
	DynAccessSet(sv, key, val)
	if len(sv.Keys) != 2 || sv.Keys[1] != "age" {
		t.Fatalf("expected key 'age' added, got keys %v", sv.Keys)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./pkg/runtime/... -run TestDynAccess -v`
Expected: FAIL — `DynAccess` undefined

**Step 3: Implement `DynAccess` and `DynAccessSet`**

Add to `pkg/runtime/ops.go`:

```go
// DynAccess dispatches on key type: int64 → Index, string → IndexKeyStr.
func DynAccess(sv *SlopValue, key *SlopValue) *SlopValue {
	if len(key.Elements) != 1 {
		panic("sloplang: dynamic access key must be a single-element array")
	}
	switch k := key.Elements[0].(type) {
	case int64:
		return Index(sv, key)
	case string:
		return IndexKeyStr(sv, k)
	default:
		panic(fmt.Sprintf("sloplang: dynamic access key must be int64 or string, got %T", key.Elements[0]))
	}
}

// DynAccessSet dispatches on key type: int64 → IndexSet, string → IndexKeySetStr.
func DynAccessSet(sv *SlopValue, key *SlopValue, val *SlopValue) *SlopValue {
	if len(key.Elements) != 1 {
		panic("sloplang: dynamic access key must be a single-element array")
	}
	switch k := key.Elements[0].(type) {
	case int64:
		return IndexSet(sv, key, val)
	case string:
		return IndexKeySetStr(sv, k, val)
	default:
		panic(fmt.Sprintf("sloplang: dynamic access key must be int64 or string, got %T", key.Elements[0]))
	}
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./pkg/runtime/... -run TestDynAccess -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/runtime/ops.go pkg/runtime/ops_test.go
git commit -m "feat: add DynAccess/DynAccessSet runtime functions for unified $ variable access"
```

---

### Task 2: Runtime — Strict boolean expressions

**Files:**
- Modify: `pkg/runtime/slop_value.go`
- Modify: `pkg/runtime/ops_test.go`

**Step 1: Write failing tests for strict IsTruthy**

Add to `pkg/runtime/ops_test.go`:

```go
func TestIsTruthy_One(t *testing.T) {
	sv := NewSlopValue(int64(1))
	if !sv.IsTruthy() {
		t.Fatal("[1] should be truthy")
	}
}

func TestIsTruthy_Empty(t *testing.T) {
	sv := NewSlopValue()
	if sv.IsTruthy() {
		t.Fatal("[] should be falsy")
	}
}

func TestIsTruthy_Zero_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("[0] should panic in boolean context")
		}
	}()
	sv := NewSlopValue(int64(0))
	sv.IsTruthy()
}

func TestIsTruthy_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("[1, 2] should panic in boolean context")
		}
	}()
	sv := NewSlopValue(int64(1), int64(2))
	sv.IsTruthy()
}

func TestIsTruthy_String_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("[\"hello\"] should panic in boolean context")
		}
	}()
	sv := NewSlopValue("hello")
	sv.IsTruthy()
}
```

**Step 2: Run tests to verify new panic tests fail (they currently don't panic)**

Run: `go test ./pkg/runtime/... -run TestIsTruthy -v`
Expected: `TestIsTruthy_Zero_Panics`, `TestIsTruthy_MultiElement_Panics`, `TestIsTruthy_String_Panics` FAIL

**Step 3: Implement strict `IsTruthy`**

Replace `IsTruthy` in `pkg/runtime/slop_value.go`:

```go
// IsTruthy returns true only for [1] (single-element int64 with value 1).
// [] is falsy. Everything else panics.
func (sv *SlopValue) IsTruthy() bool {
	if len(sv.Elements) == 0 {
		return false
	}
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got %d-element array", len(sv.Elements)))
	}
	elem := sv.Elements[0]
	if _, ok := elem.(SlopNull); ok {
		panic("sloplang: cannot use null as boolean")
	}
	i, ok := elem.(int64)
	if !ok {
		panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got single-element %T", elem))
	}
	if i == 1 {
		return true
	}
	if i == 0 {
		panic("sloplang: [0] is not a valid boolean — use [] for false")
	}
	panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got [%d]", i))
}
```

**Step 4: Fix existing runtime tests that rely on old truthiness**

Update any runtime tests that use `[0]` as falsy or multi-element arrays as truthy.

**Step 5: Run all runtime tests**

Run: `go test ./pkg/runtime/... -v`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add pkg/runtime/slop_value.go pkg/runtime/ops_test.go
git commit -m "feat: strict boolean — only [1] is truthy, only [] is falsy, everything else panics"
```

---

### Task 3: Parser — Add `$var` postfix, remove `@(expr)` and `@$var`

**Files:**
- Modify: `pkg/parser/parser.go`
- Modify: `pkg/parser/ast.go`
- Test: `pkg/parser/parser_test.go`

**Step 1: Update AST — rename `DynKeyAccessExpr` to `DynAccessExpr`, `DynKeySetStmt` to `DynAccessSetStmt`**

In `pkg/parser/ast.go`:

- Rename `DynKeyAccessExpr` → `DynAccessExpr` (same fields: `Object Expr`, `KeyVar Expr`)
- Rename `DynKeySetStmt` → `DynAccessSetStmt` (same fields: `Object Expr`, `KeyVar Expr`, `Value Expr`)
- Update `TokenLiteral()` to return `"$"` and `"$="` respectively

**Step 2: Update `parsePostfix()` in parser.go**

Current `@` dispatch (lines 641-661):
1. `@$ident` → DynKeyAccessExpr
2. `@ident` → KeyAccessExpr
3. else → IndexExpr via parsePostfixPrimary

New dispatch — `@` and `$` are separate postfix operators:

```go
// In parsePostfix() loop:
if p.curToken().Type == lexer.TOKEN_AT {
    p.advance() // consume @
    if p.curToken().Type == lexer.TOKEN_IDENT && p.peekToken().Type != lexer.TOKEN_LPAREN {
        // @ident — static string key access
        keyName := p.curToken().Literal
        p.advance()
        expr = &KeyAccessExpr{Object: expr, Key: keyName}
    } else {
        // @number — numeric literal index
        idx := p.parsePostfixPrimary()
        if idx == nil { return nil }
        expr = &IndexExpr{Object: expr, Index: idx}
    }
} else if p.curToken().Type == lexer.TOKEN_DOLLAR {
    p.advance() // consume $
    if p.curToken().Type != lexer.TOKEN_IDENT {
        p.addError("expected identifier after $, got %s at line %d", p.curToken().Type, p.curToken().Line)
        return nil
    }
    varName := p.curToken().Literal
    p.advance()
    expr = &DynAccessExpr{Object: expr, KeyVar: &Identifier{Name: varName}}
}
```

**Step 3: Update statement-level disambiguation**

Add `if p.peekToken().Type == lexer.TOKEN_DOLLAR` check before `TOKEN_AT` for `ident$var = val` → `DynAccessSetStmt`.

Remove `TOKEN_DOLLAR` handling from inside the `TOKEN_AT` block.

**Step 4: Run parser tests**

Run: `go test ./pkg/parser/... -v`
Expected: PASS

**Step 5: Commit**

```bash
git add pkg/parser/parser.go pkg/parser/ast.go
git commit -m "feat: add $var postfix syntax, remove @(expr) and @\$var from parser"
```

---

### Task 4: Codegen — Update to emit `DynAccess`/`DynAccessSet`

**Files:**
- Modify: `pkg/codegen/codegen.go`

**Step 1: Update `lowerExpr()`**

Replace the `DynKeyAccessExpr` case:
```go
case *parser.DynAccessExpr:
    return callSloprt("DynAccess", g.lowerExpr(e.Object), g.lowerExpr(e.KeyVar))
```

**Step 2: Update `lowerStmt()`**

Replace `DynKeySetStmt` case:
```go
case *parser.DynAccessSetStmt:
    return []ast.Stmt{
        &ast.ExprStmt{X: callSloprt("DynAccessSet", g.lowerExpr(s.Object), g.lowerExpr(s.KeyVar), g.lowerExpr(s.Value))},
    }
```

**Step 3: Run codegen unit tests**

Run: `go test ./pkg/codegen/... -run TestCodegen -v`
Expected: PASS

**Step 4: Commit**

```bash
git add pkg/codegen/codegen.go
git commit -m "feat: codegen emits DynAccess/DynAccessSet for \$var syntax"
```

---

### Task 5: Parser — Reject bare literals outside `[]`

**Files:**
- Modify: `pkg/parser/parser.go`
- Test: `pkg/parser/parser_test.go`

**Step 1: Remove bare literal cases from `parsePrimary()`**

Remove `TOKEN_INT/UINT/FLOAT`, `TOKEN_TRUE/FALSE`, `TOKEN_NULL` cases from `parsePrimary()`. These tokens should ONLY be parsed inside `parseArrayLiteral()`.

**Important exception:** `parsePostfixPrimary()` still accepts bare numeric literals for `arr@0`, `arr::1::3` (index/slice bounds).

**Step 2: Write parser rejection tests**

```go
func TestParser_RejectBareNumber(t *testing.T) { ... }
func TestParser_RejectBareNull(t *testing.T) { ... }
func TestParser_RejectBareTrue(t *testing.T) { ... }
```

**Step 3: Run parser tests, fix any that use bare literals**

Run: `go test ./pkg/parser/... -v`

**Step 4: Commit**

```bash
git add pkg/parser/parser.go pkg/parser/parser_test.go
git commit -m "feat: reject bare numbers, booleans, null outside [] brackets"
```

---

### Task 6: Update all E2E tests — bare literals + strict booleans + `$var` syntax

**Files:**
- Modify: `pkg/codegen/codegen_e2e_test.go`
- Modify: `pkg/codegen/phase7_error_handling_test.go`
- Modify: `pkg/codegen/codegen_test.go`

**~58 tests** use bare numbers/null. **~19 tests** break under strict booleans. **14 occurrences** of `@(`, **13 occurrences** of `@$`.

**Bare literal fixes:**
- `sum = 0` → `sum = [0]`
- `count = 0` → `count = [0]`
- `if err == 0` → `if err == [0]`
- `if err != 0` → `if err != [0]`
- `x = null` → `x = [null]`
- `null == null` → `[null] == [null]`
- `null +` → `[null] +`

**Strict boolean fixes (~19 tests):**
- `TestE2E_ZeroIsTruthy` — rename/rewrite: `[0]` in boolean context should now be a panic test
- `TestE2E_EmptyStringIsTruthy` — `if "" { }` panics now; rewrite as panic test
- `TestE2E_EmptyArrayIsFalsy` — should still work (uses `[]`)
- `TestE2E_Phase7_EmptyArrayIsFalsy` — `if result { }` → `if #result > [0] { }`
- `TestE2E_FnMutualRecursion` — `if isEven([4]) { }` works IF `isEven` returns `[1]`/`[]`; verify
- `TestE2E_NotTruthy`, `TestE2E_NotZeroIsTruthy`, `TestE2E_NotFalsy` — `![0]` panics; rewrite
- `TestE2E_DoubleNotTruthy`, `TestE2E_DoubleNotFalsy`, `TestE2E_DoubleNotEmpty` — verify `!![1]`/`!![]` still works
- `TestE2E_NotNonEmpty`, `TestE2E_NotEmptyArray` — `![1]` → fine, `![]` → fine
- `TestE2E_IfNotOperator` — `if ![0]` → panics; rewrite
- `TestE2E_AndBothTruthy`, `TestE2E_AndRightFalsy`, `TestE2E_AndLeftFalsy` — `[1] && [1]` works
- `TestE2E_OrLeftFalsy`, `TestE2E_OrLeftTruthy`, `TestE2E_OrBothFalsy` — `[] || [1]` works
- `TestE2E_IfLogical` — `if [1] && [1]` works
- `TestE2E_IfTruthy` / `TestE2E_IfFalsy` — check what value they test with

**`$var` syntax fixes:**
- `@(idx)` → `$idx`, `@(i)` → `$i`
- `@$var` → `$var`, `@$k` → `$k`, `@$w` → `$w`
- `arr@(i + [1])` → temp variable: `tmp = i + [1]` then `arr$tmp`

**Step 1: Apply all replacements**

**Step 2: Run all E2E tests**

Run: `go test ./pkg/codegen/... -v -timeout 600s`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add pkg/codegen/
git commit -m "fix: update all E2E tests for bracketed literals, strict booleans, \$var syntax"
```

---

### Task 7: Update example .slop files

**Files:**
- Modify: all files in `examples/` directory

Update bare literals, old `@` syntax, and any non-strict boolean usage.

**Step 1: Review and update each file**

**Step 2: Commit**

```bash
git add examples/
git commit -m "fix: update example programs for new syntax"
```

---

### Task 8: Update documentation

**Files:**
- Modify: `docs/PRD.md`
- Modify: `docs/architecture.md`
- Modify: `docs/patterns.md`
- Modify: `docs/roadmap.md`

Update all references to:
- `@(expr)` → `$var`
- `@$var` → `$var`
- Bare literal examples → bracketed
- Boolean semantics: `[1]` truthy, `[]` falsy, everything else panics
- Add patterns.md entries about the syntax changes

**Step 1: Update all docs**

**Step 2: Commit**

```bash
git add docs/
git commit -m "docs: update PRD, architecture, patterns for syntax refactor"
```

---

### Task 9: Clean up dead code

**Files:**
- Modify: `pkg/runtime/ops.go` — remove `IndexKey()` and `IndexKeySet()` (replaced by `DynAccess`/`DynAccessSet`)
- Modify: `pkg/parser/parser.go` — remove any remaining old `@$` code paths

**Step 1: Remove dead functions**

**Step 2: Run all tests**

Run: `go test ./... -timeout 600s`
Expected: ALL PASS

**Step 3: Final commit**

```bash
git add -A
git commit -m "refactor: remove dead IndexKey/IndexKeySet and old @\$ parser paths"
```
