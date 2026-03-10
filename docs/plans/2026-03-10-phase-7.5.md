# Phase 7.5: Syntax Strictness Refactor + Semantic E2E Tests

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Harden language semantics with strict syntax rules and verify every semantic rule with 355 exhaustive E2E tests across 9 files.

**Architecture:** Each test file covers one semantic domain. Tests use the existing `runE2E(t, source)` and `runE2EExpectPanic(t, source)` helpers in `pkg/codegen/`. Every test is a standalone function (no table-driven). Tests verify exact stdout output or that the program panics.

**Tech Stack:** Go testing, existing E2E harness (`runE2E`, `runE2EExpectPanic`, `runE2EWithStdin`)

---

## Design

### Refactors (Syntax Strictness)

- **Bracket-only literals:** Bare numbers (`x = 0`) and bare null (`x = null`) rejected outside `[]` brackets. Parser tracks `arrayDepth` counter — numbers/booleans/null only allowed when `arrayDepth > 0`.
- **Unified `$var` dynamic access:** Replaces `@$var` and `@(expr)`. Dispatches at runtime: int64 key → array index, string key → hashmap key lookup.
- **`true`/`false` keywords:** `true` → `ArrayLiteral([1])`, `false` → `ArrayLiteral([])`. Standalone (no brackets). `[true]` → `[[1]]` (nested).
- **FormatValue bracket notation:** All non-string values use brackets. Single-element strings → raw string.
- **StdoutWrite:** `fmt.Print` (no trailing newline). Explicit `"\n"` required.
- **Deep structural equality:** `==`/`!=` compare lengths, keys, and elements recursively. Ordered comparisons still require single-element.
- **IndexSet/IndexKeySetStr unwrap:** Single-element SlopValues unwrap to raw elements. Multi-element store as nested `*SlopValue`.

### Key Semantic Rules Tested

1. **Single-Element Unwrap Rule** — `IndexSet`/`IndexKeySetStr` unwrap single-element SlopValues. `Push` spreads elements. `Pop`/`RemoveAt` wrap raw elements back.
2. **Deep Structural Equality** — `==`/`!=` compare recursively. Cross-type → `[]`. Null equals null. Ordered comparisons require single-element.
3. **FormatValue Bracket Notation** — Single-element strings → raw. Everything else → brackets. `|>` uses `fmt.Print`.
4. **Boolean Strictness** — Only `[1]` truthy, `[]` falsy. Everything else panics.
5. **Null Strictness** — Panics in arithmetic, negate, boolean, logical, ordered, single-null iteration. Succeeds in `==`/`!=`, `??`, `str()`, `|>`, `#`, containment, assignment.
6. **Arithmetic Type Safety** — Element-wise, same-length, same-type. Cross-type panics. `%` rejects float. Negate converts uint64→int64.
7. **Mutation Visibility** — `Push`/`Pop`/`RemoveAt`/`IndexSet`/`IndexKeySetStr` mutate in place. `Concat`/`Slice`/`Remove`/`Unique` return new arrays. Functions receive pointers.

---

## Part A: Syntax Strictness Refactor

### Task A1: Runtime — Add `DynAccess` and `DynAccessSet`

**Files:** `pkg/runtime/ops.go`, `pkg/runtime/ops_test.go`

Add `DynAccess(sv, key)` and `DynAccessSet(sv, key, val)` — type-dispatching functions: int64 key → `Index`/`IndexSet`, string key → `IndexKeyStr`/`IndexKeySetStr`. Panic on multi-element or non-int/string keys.

### Task A2: Runtime — Strict boolean expressions

**Files:** `pkg/runtime/slop_value.go`, `pkg/runtime/ops_test.go`

Rewrite `IsTruthy()`: only `[1]` truthy, `[]` falsy. `[0]` panics with "use [] for false". Multi-element, strings, floats, null, uint64 all panic.

### Task A3: Parser — Add `$var` postfix, remove `@(expr)` and `@$var`

**Files:** `pkg/parser/parser.go`, `pkg/parser/ast.go`, `pkg/parser/parser_test.go`

- Rename `DynKeyAccessExpr` → `DynAccessExpr`, `DynKeySetStmt` → `DynAccessSetStmt`
- Add `TOKEN_DOLLAR` as postfix operator in `parsePostfix()`
- Add `$var =` statement dispatch before `@` check
- Remove `@$` and `@(expr)` paths

### Task A4: Codegen — Emit `DynAccess`/`DynAccessSet`

**Files:** `pkg/codegen/codegen.go`

Update `lowerExpr()` and `lowerStmt()` to emit `sloprt.DynAccess`/`sloprt.DynAccessSet` for `DynAccessExpr`/`DynAccessSetStmt`.

### Task A5: Parser — Reject bare literals outside `[]`

**Files:** `pkg/parser/parser.go`, `pkg/parser/parser_test.go`

Remove `TOKEN_INT/UINT/FLOAT`, `TOKEN_TRUE/FALSE`, `TOKEN_NULL` from `parsePrimary()`. Track `arrayDepth` counter — these tokens only allowed inside `parseArrayLiteral()`. `true`/`false` become standalone keywords producing `ArrayLiteral([1])`/`ArrayLiteral([])`. `parsePostfixPrimary()` retains bare number handling for `@0` and `::1::3`.

### Task A6: Update all E2E tests

**Files:** `pkg/codegen/codegen_e2e_test.go`, `pkg/codegen/phase7_error_handling_test.go`

~90 changes across ~70 tests: bracket-wrap bare numbers/null, rewrite strict boolean violations, replace `@(ident)` → `$ident` and `@$var` → `$var`.

### Task A7: Update example `.slop` files

**Files:** all `examples/*.slop`

### Task A8: Update documentation

**Files:** `docs/PRD.md`, `docs/architecture.md`, `docs/patterns.md`

### Task A9: Remove dead code

**Files:** `pkg/runtime/ops.go`, `pkg/parser/parser.go`

Remove old `IndexKey()`/`IndexKeySet()` and any remaining `@$` parser paths.

---

## Part B: Semantic E2E Tests (355 tests)

## Prerequisites

**Read before starting any task:**
- `docs/PRD.md` — language spec
- `docs/patterns.md` — lessons learned (critical for correct expected output)
- `sloplang/pkg/codegen/codegen_e2e_test.go:23-80` — `runE2E` helper
- `sloplang/pkg/codegen/codegen_e2e_test.go:1439-1480` — `runE2EExpectPanic` helper

**Test helpers available (defined in `codegen_e2e_test.go`):**
- `runE2E(t, source string) string` — transpiles, compiles, runs, returns stdout
- `runE2EExpectPanic(t, source string)` — asserts non-zero exit code (runtime panic)
- `runE2EWithStdin(t, source, stdinInput string) string` — provides stdin

**Key rules for writing expected output:**
- `str([42])` → `[42]` (brackets on all non-strings)
- `str("hello")` → `hello` (raw string, no brackets)
- `|>` does NOT add a newline — `|> "a"` then `|> "b"` → `ab`
- `true` = `[1]`, `false` = `[]`
- `data@1 = [999]` unwraps to raw `999`, NOT nested `[999]`
- `data@1 = [1, 2]` stores nested `*SlopValue`
- `--` is TOKEN_REMOVE, double negate is `-(-[x])`
- `HashDeclStmt` re-declares variable — use different names per hashmap in same test

**Run command for all tasks:**
```bash
cd sloplang && go test -count=1 -run "TestSem" ./pkg/codegen/...
```

---

### Task 1: Mutation Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_mutation_e2e_test.go`

**Step 1: Write the test file**

All tests in `package codegen`. Every function name starts with `TestSem_Mut_`. Cover these cases:

**IndexSet (`arr@i = val`) — ~16 tests:**
- `_SingleInt`: `arr = [100, 200, 300]` / `arr@1 = [999]` / `str(arr)` → `[100, 999, 300]`
- `_MultiElement`: `arr@1 = [1, 2]` / `str(arr)` → `[100, [1, 2], 300]`
- `_String`: `arr@0 = "hello"` / `str(arr)` → `[hello, 200, 300]`
- `_Null`: `arr@0 = [null]` / `str(arr)` → `[null, 200, 300]`
- `_EmptyArray`: `arr@0 = []` / `str(arr)` → `[[], 200, 300]`
- `_Float`: `arr@0 = [3.14]` / `str(arr)` → `[3.14, 200, 300]`
- `_FirstIndex`: set `@0`, verify
- `_LastIndex`: set `@2` on len-3 array, verify
- `_OutOfBounds`: `arr@5 = [1]` → panic
- `_NegativeIndex`: `arr@-1 = [1]` — check if parser handles this or panics at runtime
- `_ReadBack`: `arr@1 = [5]` then `str(arr@1)` → `[5]`
- `_ReadBackMulti`: `arr@1 = [1,2]` then `str(arr@1)` → `[1, 2]`
- `_LengthUnchanged`: set `@1`, then `str(#arr)` → `[3]`
- `_DoubleSet`: set `@0 = [5]` then `@0 = [10]`, verify `[10, 200, 300]`
- `_SetInLoop`: for loop setting `arr$i` for each index
- `_SetThenConcat`: set then `++`, verify both

**DynAccessSet (`arr$var = val`) — ~6 tests:**
- `_IntKey_Single`: `idx = [1]` / `arr$idx = [999]` → unwraps
- `_IntKey_Multi`: `idx = [1]` / `arr$idx = [1, 2]` → nests
- `_StringKey_Single`: `key = "name"` / `map$key = [42]` → unwraps
- `_StringKey_Multi`: `key = "name"` / `map$key = [1, 2]` → nests
- `_FloatKey_Panics`: `key = [3.14]` / `arr$key = [1]` → panic
- `_MultiElementKey_Panics`: `key = [1, 2]` / `arr$key = [1]` → panic

**KeySetStr (`map@key = val`) — ~7 tests:**
- `_UpdateSingle`: `person@age = [31]` / `str(person@age)` → `[31]`
- `_UpdateMulti`: `person@age = [1, 2]` / `str(person@age)` → `[1, 2]`
- `_AddNewKey`: `person@email = "a@b"` / `str(person@email)` → `a@b`
- `_ValuesAfterSingleSet`: `@@person` shows raw values not nested
- `_ValuesAfterMultiSet`: `@@person` shows nested for multi-element
- `_KeysAfterAdd`: `##person` includes new key
- `_AddNewKeyInt`: `person@score = [100]` / `str(person@score)` → `[100]`

**Push (`<<`) — ~9 tests:**
- `_SingleElement`: `arr << [4]` → `[1, 2, 3, 4]`
- `_MultiElement`: `arr << [4, 5]` → `[1, 2, 3, 4, 5]`
- `_EmptyPush`: `arr << []` → `[1, 2, 3]` (unchanged)
- `_ToEmpty`: `e = []` / `e << [1]` → `[1]`
- `_Null`: `arr << [null]` → `[1, 2, 3, null]`
- `_ThenLength`: `arr << [4]` / `str(#arr)` → `[4]`
- `_MultiplePushes`: push 3 values, verify final
- `_Nested`: `arr << [[1, 2]]` → `[1, 2, 3, [1, 2]]` (spread outer)
- `_String`: `arr = ["a"]` / `arr << "b"` — verify behavior

**Pop (`>>`) — ~6 tests:**
- `_Basic`: `>>arr` on `[10, 20, 30]` → returns `[30]`, arr is `[10, 20]`
- `_SingleElement`: pop from `[42]` → returns `[42]`, arr is `[]`
- `_Empty_Panics`: pop from `[]` → panic
- `_ReturnValue`: `str(>>arr)` → `[30]`
- `_AfterIndexSet`: set `@2 = [99]` then pop → `[99]`
- `_ThenPushRoundtrip`: pop, push back, same array

**RemoveAt (`~@`) — ~6 tests:**
- `_Middle`: `arr ~@ [1]` on `[10, 20, 30, 40]` → returns `[20]`, arr is `[10, 30, 40]`
- `_First`: `arr ~@ [0]` → removes first
- `_Last`: `arr ~@ [3]` on 4-element → removes last
- `_SingleElement`: `[42] ~@ [0]` → returns `[42]`, arr is `[]`
- `_OutOfBounds_Panics`: `arr ~@ [10]` → panic
- `_ReturnValue`: `str(arr ~@ [1])` → `[20]`

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Mut" ./pkg/codegen/... -v
```
Expected: ALL PASS

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_mutation_e2e_test.go
git commit -m "test: semantic mutation E2E tests (~45 tests)"
```

---

### Task 2: Equality Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_equality_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Eq_`.

**Same-type scalar — ~6 tests:**
- `_IntEqual`: `[1] == [1]` → `[1]`
- `_IntNotEqual`: `[1] == [2]` → `[]`
- `_FloatEqual`: `[3.14] == [3.14]` → `[1]`
- `_UintEqual`: `[42u] == [42u]` → `[1]`
- `_StringEqual`: `"hello" == "hello"` — verify via variables and `==`
- `_StringNotEqual`: `"a" == "b"` → `[]`

**Cross-type (no panic, returns `[]`) — ~3 tests:**
- `_IntVsFloat`: `[1] == [1.0]` → `[]`
- `_IntVsUint`: `[1] == [1u]` → `[]`
- `_IntVsString`: compare int to string in array context → `[]`

**Multi-element deep — ~5 tests:**
- `_MultiEqual`: `[1, 2, 3] == [1, 2, 3]` → `[1]`
- `_MultiNotEqual`: `[1, 2, 3] == [1, 2, 4]` → `[]`
- `_DifferentLengths`: `[1, 2] == [1, 2, 3]` → `[]`
- `_EmptyEqual`: `[] == []` → `[1]`
- `_EmptyVsNonEmpty`: `[] == [1]` → `[]`

**Nested array — ~3 tests:**
- `_NestedEqual`: `[[1, 2], [3]] == [[1, 2], [3]]` → `[1]`
- `_NestedNotEqual`: `[[1, 2], [3]] == [[1, 2], [4]]` → `[]`
- `_NestedVsFlat`: `[[1]] == [1]` — check behavior (nested *SlopValue vs raw int64)

**Null — ~6 tests:**
- `_NullNull`: `[null] == [null]` → `[1]`
- `_NullNeqNull`: `[null] != [null]` → `[]`
- `_NullVsInt`: `[null] == [1]` → `[]`
- `_NullNeqInt`: `[null] != [1]` → `[1]`
- `_ArrayWithNull`: `[1, null] == [1, null]` → `[1]`
- `_NullOrderMatters`: `[null, 1] == [1, null]` → `[]`

**Hashmap — ~4 tests:**
- `_SameMap`: build two identical maps, compare → `[1]`
- `_DifferentValues`: same keys, different values → `[]`
- `_DifferentKeys`: different key names → `[]` (Keys arrays differ)
- `_EmptyMaps`: `[] == []` (both empty hashmap-like)

**Inequality `!=` — ~4 tests:**
- `_NeqTrue`: `[1] != [2]` → `[1]`
- `_NeqFalse`: `[1] != [1]` → `[]`
- `_NeqMulti`: `[1, 2] != [1, 3]` → `[1]`
- `_NeqEmpty`: `[] != [1]` → `[1]`

**In boolean context — ~2 tests:**
- `_EqInIf`: `if [1, 2] == [1, 2] { |> "yes" }` → `yes`
- `_NeqInIf`: `if [1, 2] != [1, 3] { |> "yes" }` → `yes`

**After mutation — ~3 tests:**
- `_AfterIndexSet`: set `arr@0 = [5]`, compare with `[5, 2, 3]`
- `_AfterPush`: push then compare
- `_BuildTwoSeparately`: build two arrays with same ops, compare

**Ordered comparisons — ~10 tests:**
- `_LtTrue`: `[1] < [2]` → `[1]`
- `_LtFalse`: `[2] < [1]` → `[]`
- `_GtTrue`: `[2] > [1]` → `[1]`
- `_LteEqual`: `[1] <= [1]` → `[1]`
- `_GteSmaller`: `[1] >= [2]` → `[]`
- `_StringOrder`: `"a" < "b"` via variables → `[1]`
- `_FloatOrder`: `[1.5] < [2.5]` → `[1]`
- `_MultiElement_Panics`: `[1, 2] < [3, 4]` → panic
- `_NullOrdered_Panics`: `[null] < [1]` → panic
- `_TypeMismatch_Panics`: `[1] < [1.0]` → panic (int vs float in compareElems)

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Eq" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_equality_e2e_test.go
git commit -m "test: semantic equality E2E tests (~45 tests)"
```

---

### Task 3: Format Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_format_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Fmt_`.

**`str()` output — ~12 tests:**
- `_Int`: `str([42])` → `[42]`
- `_Float`: `str([3.14])` → `[3.14]`
- `_Uint`: `str([42u])` → `[42]`
- `_MultiInt`: `str([1, 2, 3])` → `[1, 2, 3]`
- `_Empty`: `str([])` → `[]`
- `_String`: `str("hello")` → `hello`
- `_Null`: `str([null])` → `[null]`
- `_Mixed`: `str([1, null, "hi"])` → `[1, null, hi]`
- `_NestedArrays`: `str([[1, 2], [3]])` → `[[1, 2], [3]]`
- `_MixedNesting`: `str([1, [2, 3]])` → `[1, [2, 3]]`
- `_True`: `str(true)` → `[1]`
- `_False`: `str(false)` → `[]`

**`|>` no trailing newline — ~5 tests:**
- `_Concatenation`: `|> "hello"` then `|> "world"` → `helloworld`
- `_ExplicitNewline`: `|> "hello"` / `|> "\n"` / `|> "world"` → `hello\nworld`
- `_IntNoNewline`: `|> str([42])` → `[42]` (no trailing newline)
- `_NullOutput`: `|> [null]` → `[null]`
- `_EmptyOutput`: `|> []` → `[]`
- `_MultiOutput`: `|> [1, 2]` → `[1, 2]`
- `_MultipleStmts`: `|> str([1])` / `|> str([2])` / `|> str([3])` → `[1][2][3]`

**FormatValue after operations — ~10 tests:**
- `_IndexAccess`: `str(arr@0)` for int element → `[10]`
- `_IndexAccessString`: `str(arr@0)` for string element → `hello`
- `_PopReturn`: `str(>>arr)` → `[30]`
- `_RemoveAtReturn`: `str(arr ~@ [1])` → `[20]`
- `_Slice`: `str(arr::0::2)` → `[10, 20]`
- `_Concat`: `str([1, 2] ++ [3, 4])` → `[1, 2, 3, 4]`
- `_Unique`: `str(~[1, 1, 2])` → `[1, 2]`
- `_ContainsTrue`: `str([1, 2, 3] ?? [2])` → `[1]`
- `_ContainsFalse`: `str([1, 2, 3] ?? [9])` → `[]`
- `_Length`: `str(#[1, 2, 3])` → `[3]`

**FormatValue with hashmaps — ~3 tests:**
- `_MapKeysSingle`: `str(##m)` on single-key map → `key` (raw string)
- `_MapValuesMulti`: `str(@@m)` on multi-key → `[v1, v2]`
- `_MapValuesAfterMutation`: set key, then `@@` reflects update

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Fmt" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_format_e2e_test.go
git commit -m "test: semantic format E2E tests (~30 tests)"
```

---

### Task 4: Boolean Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_boolean_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Bool_`.

**Valid booleans — ~6 tests:**
- `_TruthyOne`: `if [1] { |> "yes" }` → `yes`
- `_FalsyEmpty`: `if [] { |> "yes" } else { |> "no" }` → `no`
- `_TrueKeyword`: `if true { |> "yes" }` → `yes`
- `_FalseKeyword`: `if false { |> "yes" } else { |> "no" }` → `no`
- `_TrueAssign`: `x = true` / `if x { |> "yes" }` → `yes`
- `_FalseAssign`: `x = false` / `if x { } else { |> "no" }` → `no`

**Invalid booleans (all panic) — ~10 tests:**
- `_ZeroPanics`: `if [0] { }` → panic
- `_TwoPanics`: `if [2] { }` → panic
- `_NegOnePanics`: `if -[1] { }` → panic (result is `[-1]`)
- `_StringPanics`: `if "hello" { }` → panic
- `_EmptyStringPanics`: `if "" { }` → panic
- `_FloatPanics`: `if [3.14] { }` → panic
- `_MultiPanics`: `if [1, 2] { }` → panic
- `_NullPanics`: `if [null] { }` → panic
- `_UintZeroPanics`: `if [0u] { }` → panic
- `_UintOnePanics`: `if [1u] { }` → panic (uint64 not int64)

**Logical operators — ~11 tests:**
- `_AndBothTrue`: `if [1] && [1] { |> "yes" }` → `yes`
- `_AndLeftFalse`: `if [] && [1] { |> "yes" } else { |> "no" }` → `no`
- `_AndRightFalse`: `if [1] && [] { |> "yes" } else { |> "no" }` → `no`
- `_OrLeftTrue`: `if [1] || [] { |> "yes" }` → `yes`
- `_OrLeftFalse`: `if [] || [1] { |> "yes" }` → `yes`
- `_OrBothFalse`: `if [] || [] { |> "yes" } else { |> "no" }` → `no`
- `_NotTrue`: `if ![1] { |> "yes" } else { |> "no" }` → `no`
- `_NotFalse`: `if ![] { |> "yes" }` → `yes`
- `_NotTrueKw`: `if !true { } else { |> "no" }` → `no`
- `_NotFalseKw`: `if !false { |> "yes" }` → `yes`
- `_DoubleNot`: `if !!true { |> "yes" }` → `yes`

**Logical with invalid operands (panic) — ~4 tests:**
- `_ZeroAndPanics`: `[0] && [1]` in if → panic
- `_AndRightZeroPanics`: `[1] && [0]` in if → panic (left truthy, evals right)
- `_NotZeroPanics`: `![0]` in if → panic
- `_ZeroOrPanics`: `[0] || [1]` in if → panic

**`true`/`false` keyword semantics — ~8 tests:**
- `_StrTrue`: `|> str(true)` → `[1]`
- `_StrFalse`: `|> str(false)` → `[]`
- `_TrueEqOne`: `if true == [1] { |> "yes" }` → `yes`
- `_FalseEqEmpty`: `if false == [] { |> "yes" }` → `yes`
- `_TrueNeqFalse`: `if true != false { |> "yes" }` → `yes`
- `_BracketTrue`: `|> str([true])` → `[[1]]` (nested)
- `_BracketFalse`: `|> str([false])` → `[[]]` (nested)
- `_AssignThenStr`: `x = true` / `|> str(x)` → `[1]`

**Boolean results used — ~2 tests:**
- `_EqResultTruthy`: `r = [1] == [1]` / `if r { |> "yes" }` → `yes`
- `_EqResultFalsy`: `r = [1] == [2]` / `if r { |> "yes" } else { |> "no" }` → `no`

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Bool" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_boolean_e2e_test.go
git commit -m "test: semantic boolean E2E tests (~40 tests)"
```

---

### Task 5: Null Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_null_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Null_`.

**Null succeeds — ~13 tests:**
- `_Assignment`: `x = [null]` / `|> str(x)` → `[null]`
- `_StrNull`: `|> str([null])` → `[null]`
- `_StdoutNull`: `|> [null]` → `[null]`
- `_EqNullNull`: `if [null] == [null] { |> "yes" }` → `yes`
- `_NeqNullInt`: `if [null] != [1] { |> "yes" }` → `yes`
- `_IntNeqNull`: `if [1] != [null] { |> "yes" }` → `yes`
- `_NeqNullNull`: `if [null] != [null] { |> "yes" } else { |> "no" }` → `no`
- `_LengthNull`: `|> str(#[null])` → `[1]`
- `_LengthMultiNull`: `|> str(#[null, null])` → `[2]`
- `_ContainsNull`: `if [1, null, 3] ?? [null] { |> "yes" }` → `yes`
- `_IndexNull`: `arr = [null, null]` / `|> str(arr@0)` → `[null]`
- `_MixedArray`: `|> str([1, null, "hi"])` → `[1, null, hi]`
- `_HashmapValue`: `m{a} = [[null]]` / `|> str(m@a)` → `[null]`

**Null panics — arithmetic (~7 tests):**
- `_AddPanics`: `[null] + [1]` → panic
- `_SubPanics`: `[null] - [1]` → panic
- `_MulPanics`: `[null] * [1]` → panic
- `_DivPanics`: `[null] / [1]` → panic
- `_ModPanics`: `[null] % [1]` → panic
- `_PowPanics`: `[null] ** [1]` → panic
- `_RightSidePanics`: `[1] + [null]` → panic

**Null panics — unary/boolean (~5 tests):**
- `_NegatePanics`: `-[null]` → panic
- `_NotPanics`: `![null]` → panic (Not calls IsTruthy)
- `_IfPanics`: `if [null] { }` → panic
- `_AndPanics`: `[null] && [1]` → panic (IsTruthy on null)
- `_OrPanics`: `[null] || [1]` → panic

**Null panics — ordered comparison (~5 tests):**
- `_LtPanics`: `[null] < [1]` → panic
- `_GtPanics`: `[null] > [1]` → panic
- `_LtePanics`: `[null] <= [1]` → panic
- `_GtePanics`: `[null] >= [1]` → panic
- `_RightSideOrderedPanics`: `[1] < [null]` → panic

**Null panics — iteration (~1 test):**
- `_IterateSinglePanics`: `for x in [null] { }` → panic

**Null edge cases — ~5 tests:**
- `_MultiNullIterates`: `for x in [null, null] { |> str(x) }` → `[null][null]`
- `_UniqueNull`: `|> str(~[null, null, 1])` → `[null, 1]`
- `_RemoveNull`: `|> str([1, null, 3] -- [null])` → `[1, 3]`
- `_ConcatNull`: `|> str([1, null] ++ [null, 2])` → `[1, null, null, 2]`
- `_NullInMixedIteration`: `for x in [null, 1] { |> str(x) }` → `[null][1]`

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Null" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_null_e2e_test.go
git commit -m "test: semantic null E2E tests (~35 tests)"
```

---

### Task 6: Arithmetic Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_arithmetic_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Arith_`.

**Type consistency — ~7 tests:**
- `_IntAdd`: `|> str([1] + [2])` → `[3]`
- `_FloatAdd`: `|> str([1.5] + [2.5])` → `[4]`
- `_UintAdd`: `|> str([1u] + [2u])` → `[3]`
- `_IntFloatPanics`: `[1] + [1.0]` → panic
- `_IntUintPanics`: `[1] + [1u]` → panic
- `_FloatUintPanics`: `[1.0] + [1u]` → panic
- `_StringAddPanics`: `"a" + "b"` — verify panics (string unsupported)

**Length mismatch — ~4 tests:**
- `_LongPlusShort`: `[1, 2] + [1]` → panic
- `_ShortPlusLong`: `[1] + [1, 2]` → panic
- `_EmptyPlusEmpty`: `|> str([] + [])` → `[]`
- `_EmptyPlusOne`: `[] + [1]` → panic

**Multi-element element-wise — ~6 tests:**
- `_AddMulti`: `|> str([1, 2, 3] + [4, 5, 6])` → `[5, 7, 9]`
- `_SubMulti`: `|> str([10, 20] - [3, 7])` → `[7, 13]`
- `_MulMulti`: `|> str([2, 3] * [4, 5])` → `[8, 15]`
- `_DivMulti`: `|> str([10, 6] / [2, 3])` → `[5, 2]`
- `_ModMulti`: `|> str([7, 5] % [3, 2])` → `[1, 1]`
- `_PowMulti`: `|> str([2, 3] ** [3, 2])` → `[8, 9]`

**Division by zero — ~4 tests:**
- `_DivByZeroInt`: `[10] / [0]` → panic
- `_DivByZeroUint`: `[10u] / [0u]` → panic
- `_DivByZeroFloat`: `[10.0] / [0.0]` → panic
- `_DivByZeroSecondElem`: `[10, 20] / [5, 0]` → panic

**Negate — ~8 tests:**
- `_Int`: `|> str(-[1])` → `[-1]`
- `_Multi`: `|> str(-[1, 2, 3])` → `[-1, -2, -3]`
- `_Zero`: `|> str(-[0])` → `[0]`
- `_Float`: `|> str(-[3.14])` → `[-3.14]`
- `_DoubleNegate`: `|> str(-(-[5]))` → `[5]` (note: `-(-x)` syntax)
- `_Uint`: `|> str(-[1u])` → `[-1]` (uint→int conversion)
- `_Empty`: `|> str(-[])` → `[]`
- `_StringPanics`: `-"hello"` → panic

**Mod edge cases — ~2 tests:**
- `_BasicMod`: `|> str([7] % [3])` → `[1]`
- `_FloatModPanics`: `[7.0] % [3.0]` → panic

**Pow edge cases — ~3 tests:**
- `_ZeroExp`: `|> str([2] ** [0])` → `[1]`
- `_ZeroBase`: `|> str([0] ** [0])` → `[1]`
- `_NegExp`: `|> str([2] ** -([1]))` → `[0]` (int truncation)

**Precedence — ~3 tests:**
- `_MulBeforeAdd`: `|> str([1] + [2] * [3])` → `[7]`
- `_ParensOverride`: `|> str(([1] + [2]) * [3])` → `[9]`
- `_NegInExpr`: `|> str(-[1] + [2])` → `[1]`

**Type mismatch for all ops — ~5 tests:**
- `_SubTypeMismatch`: `[1] - [1.0]` → panic
- `_MulTypeMismatch`: `[1] * [1u]` → panic
- `_DivTypeMismatch`: `[1.0] / [1u]` → panic
- `_ModTypeMismatch`: `[1] % [1u]` → panic
- `_PowTypeMismatch`: `[1] ** [1.0]` → panic

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Arith" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_arithmetic_e2e_test.go
git commit -m "test: semantic arithmetic E2E tests (~40 tests)"
```

---

### Task 7: Array Operations Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_array_ops_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Arr_`.

**Index access — ~7 tests:**
- `_First`: `arr@0` on `[10, 20, 30]` → `[10]`
- `_Last`: `arr@2` → `[30]`
- `_OutOfBounds_Panics`: `arr@3` → panic
- `_Dynamic`: `i = [1]` / `arr$i` → `[20]`
- `_Nested`: `[[1,2],[3,4]]@0` → `[1, 2]`
- `_AfterPush`: push then access new index
- `_StringElement`: `["a", "b"]@0` → `a` (raw string)

**Slice — ~9 tests:**
- `_Basic`: `arr::0::3` → `[10, 20, 30]`
- `_EmptyLoEqHi`: `arr::2::2` → `[]`
- `_ZeroZero`: `arr::0::0` → `[]`
- `_Full`: `arr::0::5` → full copy
- `_LastElement`: `arr::4::5` → `[50]`
- `_DoesNotMutate`: slice then verify original unchanged
- `_HighOutOfBounds_Panics`: `arr::0::10` → panic
- `_NegativeLow_Panics`: check if `-1` can be passed (parser may reject)
- `_HiLessThanLo_Panics`: `arr::3::1` → panic

**Concat (`++`) — ~8 tests:**
- `_Basic`: `[1, 2] ++ [3, 4]` → `[1, 2, 3, 4]`
- `_LeftEmpty`: `[] ++ [1]` → `[1]`
- `_RightEmpty`: `[1] ++ []` → `[1]`
- `_BothEmpty`: `[] ++ []` → `[]`
- `_DoesNotMutate`: concat then verify originals unchanged
- `_Strings`: `["a"] ++ ["b"]` → `[a, b]`
- `_Mixed`: `[1, "a"] ++ [null, 2]` → `[1, a, null, 2]`
- `_Nested`: `[[1]] ++ [[2]]` → `[[1], [2]]`

**Remove (`--`) — ~5 tests:**
- `_FirstOccurrence`: `[1, 2, 3, 2, 4] -- [2]` → `[1, 3, 2, 4]`
- `_NotFound`: `[1, 2, 3] -- [9]` → `[1, 2, 3]`
- `_EmptyVal`: `[1, 2, 3] -- []` → `[1, 2, 3]`
- `_DoesNotMutate`: remove then verify original unchanged
- `_RemoveNull`: `[1, null, 3] -- [null]` → `[1, 3]`

**Contains (`??`) — ~6 tests:**
- `_Found`: `[10, 20, 30] ?? [20]` → `[1]`
- `_NotFound`: `[10, 20, 30] ?? [99]` → `[]`
- `_EmptyArray`: `[] ?? [1]` → `[]`
- `_NullFound`: `[null, 1] ?? [null]` → `[1]`
- `_MultiOperand_Panics`: `[1, 2] ?? [1, 2]` → panic
- `_InConditional`: `if arr ?? [5] { |> "yes" }` — verify boolean use

**Unique (`~`) — ~6 tests:**
- `_Basic`: `~[1, 2, 2, 3, 1]` → `[1, 2, 3]`
- `_Empty`: `~[]` → `[]`
- `_Single`: `~[1]` → `[1]`
- `_AllSame`: `~[1, 1, 1]` → `[1]`
- `_DoesNotMutate`: unique then verify original unchanged
- `_NullDedup`: `~[null, 1, null]` → `[null, 1]`
- `_MixedTypes`: `~[1, "1", 1]` → `[1, 1]` (int64 vs string not equal, both kept)
- `_Strings`: `~["a", "b", "a"]` → `[a, b]`

**Length (`#`) — ~5 tests:**
- `_Basic`: `#[1, 2, 3]` → `[3]`
- `_Empty`: `#[]` → `[0]`
- `_Null`: `#[null]` → `[1]`
- `_Nested`: `#[1, [2, 3]]` → `[2]`
- `_AfterPush`: push then length
- `_InExpression`: `if #arr == [3] { |> "yes" }` → `yes`

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Arr" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_array_ops_e2e_test.go
git commit -m "test: semantic array ops E2E tests (~45 tests)"
```

---

### Task 8: Hashmap Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_hashmap_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Map_`.

**Declaration & access — ~5 tests:**
- `_BasicDecl`: `person{name, age} = ["bob", [30]]` / `|> person@name` → `bob` / `|> str(person@age)` → `[30]`
- `_EmptyMap`: `m{} = []` / `|> str(##m)` → `[]`
- `_SingleKey`: `m{x} = [[5]]` / `|> str(m@x)` → `[5]`
- `_KeyNotFound_Panics`: `person@missing` → panic
- `_StringValue`: `m{a} = ["hello"]` / `|> m@a` → `hello`

**Key set — ~5 tests:**
- `_UpdateExisting`: `person@age = [31]` / `str(person@age)` → `[31]`
- `_AddNew`: `person@email = "a@b"` / `str(person@email)` → `a@b`
- `_SingleUnwraps`: after `person@age = [31]`, `@@person` shows `31` not `[31]`
- `_MultiNests`: `person@scores = [1, 2]` / `str(person@scores)` → `[1, 2]`
- `_AddNewInt`: `person@score = [100]` / `str(person@score)` → `[100]`

**`##` and `@@` — ~6 tests:**
- `_Keys`: `##person` → `[name, age]`
- `_Values`: `@@person` after set shows raw values
- `_KeysAfterAdd`: adding key shows in `##`
- `_ValuesAfterUpdate`: update then `@@` reflects change
- `_KeysOnNonMap`: `##arr` on plain array → `[]`
- `_ValuesOnNonMap`: `@@arr` on plain array → `[]`

**Dynamic access — ~4 tests:**
- `_StringKeyRead`: `key = "name"` / `person$key` → `bob`
- `_IntKeyOnArray`: `i = [0]` / `arr$i` → array index
- `_StringKeyWrite`: `key = "new"` / `person$key = [42]` → adds key
- `_IntKeyOnMapPanics`: int key on hashmap (what happens? — test it)

**Hashmap iteration — ~2 tests:**
- `_IterateKeys`: `for k in ##map { |> k }` → prints each key
- `_IterateValues`: `for v in @@map { |> str(v) }` → prints each value

**Hashmap in functions — ~4 tests:**
- `_PassToFn`: pass map to function, access key inside
- `_ReturnFromFn`: build map in function, return, access outside
- `_MutateInFn`: modify map inside function, visible outside (pointer)
- `_BuildInLoop`: loop building map with `$`, verify keys/values after

**Hashmap equality — ~2 tests:**
- `_SameStructure`: two identical maps built separately → `==` returns `[1]`
- `_DifferentValues`: same keys, different vals → `==` returns `[]`

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Map" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_hashmap_e2e_test.go
git commit -m "test: semantic hashmap E2E tests (~30 tests)"
```

---

### Task 9: Control Flow Semantics

**Files:**
- Create: `sloplang/pkg/codegen/semantic_control_flow_e2e_test.go`

**Step 1: Write the test file**

All function names start with `TestSem_Flow_`.

**If/else — ~7 tests:**
- `_TrueBranch`: `if true { |> "yes" }` → `yes`
- `_ElseBranch`: `if false { |> "yes" } else { |> "no" }` → `no`
- `_NestedIf`: `if true { if true { |> "deep" } }` → `deep`
- `_WithComparison`: `if [1] > [0] { |> "yes" }` → `yes`
- `_WithLogical`: `if [1] && [1] { |> "yes" }` → `yes`
- `_WithNot`: `if !false { |> "yes" }` → `yes`
- `_ChainedElse`: `if false { } else { if true { |> "second" } }` → `second`

**For-in — ~6 tests:**
- `_Basic`: `for x in [1, 2, 3] { |> str(x) }` → `[1][2][3]`
- `_EmptyArray`: `for x in [] { |> "never" }` → ``
- `_Strings`: `for x in ["a", "b"] { |> x }` → `ab`
- `_Nested`: two nested for loops, verify output
- `_WithBreak`: `for x in [1,2,3,4] { if x == [3] { break } |> str(x) }` → `[1][2]`
- `_WithMutation`: push in loop, verify after

**Infinite loop + break — ~3 tests:**
- `_BasicCounter`: `i = [0]` / `for { if i == [5] { break } i = i + [1] }` / `|> str(i)` → `[5]`
- `_NestedBreak`: break from inner only, outer continues
- `_BreakAfterMutation`: modify array in loop, break, verify

**Functions — ~8 tests:**
- `_ReturnValue`: `fn add(a, b) { <- a + b }` / `add([1], [2])` → `[3]`
- `_EarlyReturn`: `fn f(x) { if x == [0] { <- [99] } <- x }` / test both paths
- `_Recursive`: factorial or sum
- `_MutatesArray`: fn pushes to array param, verify outside
- `_LocalScope`: fn local var doesn't leak to caller
- `_MultiAssign`: `a, b = fn_returning_pair()` / verify both
- `_NoReturn`: fn with no explicit return (what happens?)
- `_NestedCalls`: fn calling fn

**Variable scoping — ~3 tests:**
- `_Reassignment`: `x = [1]` / `x = [2]` / `str(x)` → `[2]`
- `_ParamShadow`: outer var `x`, fn param `x`, verify independence
- `_LoopVar`: for loop var scope

**Combined patterns — ~5 tests:**
- `_BuildArrayInLoop`: push in loop, verify final array
- `_ErrorPropagation`: fn returns error, caller checks and branches
- `_NestedFnCalls`: 3-level function chain
- `_ArrayOpsInLoop`: index/set inside loop body
- `_HashmapInLoop`: build hashmap via `$` in loop

**Step 2: Run tests**

```bash
cd sloplang && go test -count=1 -run "TestSem_Flow" ./pkg/codegen/... -v
```

**Step 3: Commit**

```bash
git add sloplang/pkg/codegen/semantic_control_flow_e2e_test.go
git commit -m "test: semantic control flow E2E tests (~35 tests)"
```

---

### Task 10: Final Verification & JSON Tracking

**Step 1: Run all semantic tests together**

```bash
cd sloplang && go test -count=1 -run "TestSem" ./pkg/codegen/... -v 2>&1 | tail -20
```
Expected: ALL PASS

**Step 2: Run the FULL test suite (including all existing tests)**

```bash
cd sloplang && go test -count=1 ./... 2>&1
```
Expected: ALL packages PASS

**Step 3: Create JSON tracking file**

Create `docs/plans/2026-03-10-semantic-e2e-tests.json` with entries for each file.

**Step 4: Commit all**

```bash
git add docs/plans/2026-03-10-semantic-e2e-tests.json
git commit -m "test: complete semantic E2E test suite tracking"
```
