# Patterns & Lessons Learned

## Planning

- **Always write `docs/plans/*.md` and `docs/plans/*.json` as the very first action** before any code. Don't defer plan file creation to "implementation time" — CLAUDE.md requires them upfront.
- **Plans must NEVER include full code blocks.** Use code hints, descriptions, and references to existing patterns instead. The implementer writes the code, the plan guides them.
- **One markdown file per phase.** Do NOT create separate design and implementation plan files. Merge the design into the top of the implementation plan.

## Documentation

- **`docs/architecture.md` must be kept up to date after every phase.** After completing a phase implementation, always review and update the architecture doc to reflect new AST nodes, runtime functions, parser changes, operators, etc. Check for outdated sections before committing.

## Project Structure

- **All Go transpiler code lives in `sloplang/` subfolder**, not the project root. The project root has `docs/`, `CLAUDE.md`, etc. The `sloplang/` subfolder has `go.mod`, `cmd/`, `pkg/`, `examples/`.

## Codegen

- **Assignment uses `:=` first time, `=` for reassignment.** Track declared variables in a `map[string]bool`. Without this, `total = total + x` inside a loop creates a new shadowed variable each iteration via `:=`, never updating the outer one.
- **Function bodies need their own scope.** Save/restore the `declared` map around `lowerFnDecl`. Function params must be pre-registered as declared so assignments to params use `=`.
- **Top-level variables are hoisted to Go package-level `var` declarations.** Without this, user-defined functions (emitted as top-level Go funcs) can't access variables declared in `main()`. A first pass collects all top-level `AssignStmt`, `HashDeclStmt`, and `MultiAssignStmt` names into a `globals` set, then emits `var x *sloprt.SlopValue` at package level. In `main()`, these use `=` (not `:=`). In `lowerFnDecl`, the function scope is seeded with globals so assignments to global names use `=` (writes to global), while new names use `:=` (creates local). This gives Python-like scoping: local-first resolution, globals accessible from any function.
- **Unary negate on number literals inside arrays must emit raw values.** `[-1]` in sloplang must produce `NewSlopValue(int64(-1))`, not `NewSlopValue(Negate(NewSlopValue(int64(1))))`. The latter creates a nested `*SlopValue`, making `str([-1])` output `[[-1]]`. Fix: in `lowerRawValue`, detect `UnaryExpr{Op:"-", Operand:NumberLiteral}` and emit `int64(-1)` directly.
- **For-in loop variable needs `_ = varName` suppressor.** Just like regular assignments, the range variable can be unused in some cases.
- **`FormatValue` always brackets non-string values.** Single-element SlopValues now format as `[42]`, not `42`. Only single-element strings remain unbracketed. This means `str([42])` returns `"[42]"`, `str([null])` returns `"[null]"`, etc. Multi-element arrays and empty arrays are unchanged.

## Lexer

- **Adding multi-char operators can break existing semantics.** When `--` was added as `TOKEN_REMOVE` in Phase 4, the existing `--[5]` (double unary negate) broke because the lexer now greedily matches `--` as one token. The fix: double negate must be written as `-(-[5])`. Any existing tests relying on the old behavior must be updated.
- **Operator disambiguation order matters.** Multi-char operators (e.g., `<<`, `>>`, `++`, `--`, `~@`, `::`, `??`) must be checked before their single-char prefixes. Always peek before emitting the single-char token.
- **String-reading helpers must detect EOF and signal errors.** `readString()` returns `(string, bool)` — `false` means EOF was hit before a closing `"`. The caller emits `TOKEN_ILLEGAL` with literal `"unterminated string"`, which the parser already rejects.

## Parser

- **Postfix operators (`@`, `::`) need a separate parsing level.** They bind tighter than binary operators but follow primary/call expressions. Insert a `parsePostfix()` layer between `parseUnary()` and `parseCall()` that loops over `@` (index) and `::` (slice).
- **Statement-level lookahead for index-set (`arr@i = val`).** Use `save()`/`restore()` to tentatively parse `ident@expr`, then check for `=`. If not found, backtrack and parse as a normal expression.
- **`$var` replaces both `@$var` and `@(expr)`.** The `$` postfix operator does type-based dispatch at runtime: int64 key → array index, string key → hashmap key lookup. This unified syntax means `arr$i` works for both numeric and string variable access. Literal numeric indices still use `@`: `arr@0`. Literal string keys still use `@`: `map@name`.
- **Statement-level lookahead must check `$` before `@`.** When parsing `ident...`, check `$` first for `DynAccessSetStmt`, then `@` for `KeySetStmt`/`IndexSetStmt`. Each failed check must restore.
- **`arrayDepth` counter tracks bracket nesting.** Numbers, booleans, and null are only allowed in `parsePrimary()` when `arrayDepth > 0`. The counter is incremented when entering `parseArrayLiteral()` and decremented on exit (including error paths).

## Runtime

- **Null (`SlopNull`) needs explicit handling in every runtime operation.** The `default:` case in type switches catches unknown types, but the error message is generic. Add explicit `case SlopNull:` with descriptive panic messages like `"cannot perform arithmetic on null"`. This also applies to `IsTruthy`, `Negate`, `deepEqual`, comparison ops, and `Iterate`.
- **Null equality is special.** `null == null` must be truthy, `null != non-null` must be truthy. Handle this before the normal type switch in `Eq`/`Neq` to avoid falling into the string/int comparison logic.

## E2E Testing

- **`##` and `@@` on single-key hashmaps return single-element SlopValues.** `str()` on a single-element SlopValue prints just the value (e.g., `"item"`) not `[item]`. When writing expected output for `str(##map)` or `str(@@map)` where the map has only one key, expect no brackets.
- **HashDeclStmt re-declares the variable name with `:=`.** If you redeclare the same variable name via `m{a} = ...` then later `m{b} = ...`, Go codegen produces two `:=` for `m`, which fails compilation. Use different variable names in tests.
- **`IndexSet` and `IndexKeySetStr` unwrap single-element SlopValues.** `data@1 = [999]` stores raw `999` (int64), not a nested `*SlopValue`. Multi-element values store the whole `*SlopValue` (genuine nesting). This matches `Push` which spreads elements. Without unwrapping, `[100, 200, 300]` with `data@1 = [999]` would incorrectly produce `[100, [999], 300]` instead of `[100, 999, 300]`.

## Error Handling

- **`++` is array Concat, NOT string concatenation.** `"prefix: " ++ str(x)` produces a 2-element array `["prefix: ", "5"]` which formats as `[prefix: , 5]`, not `"prefix: 5"`. To build human-readable output, use separate `|>` calls instead.
- **User-defined functions returning `[result, errcode]` work via `UnpackTwo`.** The `isDualReturn()` check only applies to builtins (`StdinRead`, `FileRead`, `to_num`). User functions return a single `*SlopValue` containing two elements, which `UnpackTwo` destructures correctly.
- **`str()` on single-element numeric values now includes brackets.** `str([5])` outputs `[5]`. This matches the `FormatValue` bracket-all-non-strings behavior. File roundtrip patterns that relied on `str(x)` producing bare numbers (e.g., write `str([42])` to file, read back, `to_num()`) now break because `to_num("[42]")` fails.
- **`StdoutWrite` uses `fmt.Print` not `fmt.Println`.** No trailing newline per output. Multiple `|>` statements concatenate their output directly. If newlines are needed, the slop source must include explicit `"\n"` strings.
- **`==`/`!=` use deep structural equality.** `[1, 2] == [1, 2]`, `[1] == [1.0]`, and `[] == [1]` now succeed instead of panicking. Hashmaps compare both keys and values. Tests that previously used `runE2EExpectPanic` for these cases must be converted to regular `runE2E` tests. Ordered comparisons (`<`, `>`, `<=`, `>=`) still require single-element arrays.

## Syntax Strictness (Bracket-Wrapping Refactor)

- **Bare numbers outside `[]` are rejected.** `x = 0` must be `x = [0]`. This applies everywhere: assignments, arithmetic operands (`count + [1]`), comparisons (`err != [0]`), return values (`<- [0]`), and modulo (`v % [2] == [0]`).
- **Bare `null` outside `[]` is rejected.** `x = null` must be `x = [null]`. Same for `|> [null]`, comparisons (`[null] == [null]`), contains (`?? [null]`), arithmetic (`[null] + [1]`), negate (`-[null]`), conditionals (`if [null]`), not (`![null]`), comparisons (`[null] > [1]`), and iteration (`for x in [null]`).
- **`true` and `false` are standalone keywords.** `true` parses as `ArrayLiteral([1])`, `false` as `ArrayLiteral([])`. They do NOT require brackets. `[true]` creates `[[1]]` (nested), `[false]` creates `[[]]` (nested). Use `true` for truthy, `false` for falsy, or `[1]`/`[]` directly.
- **`[0]` panics in boolean context.** `IsTruthy()` only accepts `[1]` (truthy) and `[]` (falsy). Tests that previously checked `![0]` output must become `runE2EExpectPanic` tests.
- **Strings panic in boolean context.** `IsTruthy()` rejects single-element strings — tests using `if "hello"` must become panic tests.
- **When refactoring tests for stricter syntax, search for ALL bare number patterns.** A partial fix (only the tests explicitly listed) will miss dozens of other tests using bare numbers in arithmetic, comparisons, and assignments. Run the full test suite after fixing the obvious ones.

## Phase 8 — Real Programs

- **`#arr` cannot be used inside slice postfix expressions.** `arr::mid::#arr` fails because `#` is parsed as a new unary expression, not as part of the slice. Store the length first: `len = #arr` then `arr::mid::len`.
- **Vocab arrays must match modulo divisor.** If `seed % [100]` is used, the vocab array must have exactly 100 elements. Off-by-one means index-out-of-bounds at runtime.
- **`SLOP_MODULE_ROOT` env var** allows the CLI to find the sloplang module root from temp directories. The test harness sets this so `slop` can build programs in temp dirs.
- **Test template substitution pattern.** `.slop` programs use `SIZE_PLACEHOLDER` which the Go test replaces with actual sizes. `buildSource()` handles this.
- **Stale transpiled `.go` files in `examples/`** will cause `go test ./...` to fail if they reference old APIs. Delete them when changing runtime/codegen.
- **Go keyword sanitization in codegen.** User-defined identifiers that collide with Go's 25 keywords (e.g., `func`, `var`, `return`, `range`) must be prefixed with `slop_` via `sanitizeIdent()` at every `ast.NewIdent` call site that emits a user-defined name. This includes variable declarations, assignments, function declarations, parameter names, for-in loop variables, multi-assign names, identifier expression references, and user-defined function calls. Do NOT sanitize hardcoded Go names like runtime functions, type names, `"_"`, `"main"`, `"nil"`.
