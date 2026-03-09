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
- **For-in loop variable needs `_ = varName` suppressor.** Just like regular assignments, the range variable can be unused in some cases.
- **`FormatValue` on single-element SlopValue returns the value directly** (e.g., `7`), not `[7]`. The roadmap's expected output for `str(result)` where result is `[7]` is just `7`, not `[7]`.

## Lexer

- **Adding multi-char operators can break existing semantics.** When `--` was added as `TOKEN_REMOVE` in Phase 4, the existing `--[5]` (double unary negate) broke because the lexer now greedily matches `--` as one token. The fix: double negate must be written as `-(-[5])`. Any existing tests relying on the old behavior must be updated.
- **Operator disambiguation order matters.** Multi-char operators (e.g., `<<`, `>>`, `++`, `--`, `~@`, `::`, `??`) must be checked before their single-char prefixes. Always peek before emitting the single-char token.

## Parser

- **Postfix operators (`@`, `::`) need a separate parsing level.** They bind tighter than binary operators but follow primary/call expressions. Insert a `parsePostfix()` layer between `parseUnary()` and `parseCall()` that loops over `@` (index) and `::` (slice).
- **Statement-level lookahead for index-set (`arr@i = val`).** Use `save()`/`restore()` to tentatively parse `ident@expr`, then check for `=`. If not found, backtrack and parse as a normal expression.
