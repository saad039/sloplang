# Patterns & Lessons Learned

## Planning

- **Always write `docs/plans/*.md` and `docs/plans/*.json` as the very first action** before any code. Don't defer plan file creation to "implementation time" — CLAUDE.md requires them upfront.
- **Plans must NEVER include full code blocks.** Use code hints, descriptions, and references to existing patterns instead. The implementer writes the code, the plan guides them.
- **One markdown file per phase.** Do NOT create separate design and implementation plan files. Merge the design into the top of the implementation plan.

## Project Structure

- **All Go transpiler code lives in `sloplang/` subfolder**, not the project root. The project root has `docs/`, `CLAUDE.md`, etc. The `sloplang/` subfolder has `go.mod`, `cmd/`, `pkg/`, `examples/`.

## Codegen

- **Assignment uses `:=` first time, `=` for reassignment.** Track declared variables in a `map[string]bool`. Without this, `total = total + x` inside a loop creates a new shadowed variable each iteration via `:=`, never updating the outer one.
- **Function bodies need their own scope.** Save/restore the `declared` map around `lowerFnDecl`. Function params must be pre-registered as declared so assignments to params use `=`.
- **For-in loop variable needs `_ = varName` suppressor.** Just like regular assignments, the range variable can be unused in some cases.
- **`FormatValue` on single-element SlopValue returns the value directly** (e.g., `7`), not `[7]`. The roadmap's expected output for `str(result)` where result is `[7]` is just `7`, not `[7]`.
