# Patterns & Lessons Learned

## Planning

- **Always write `docs/plans/*.md` and `docs/plans/*.json` as the very first action** before any code. Don't defer plan file creation to "implementation time" — CLAUDE.md requires them upfront.

## Project Structure

- **All Go transpiler code lives in `sloplang/` subfolder**, not the project root. The project root has `docs/`, `CLAUDE.md`, etc. The `sloplang/` subfolder has `go.mod`, `cmd/`, `pkg/`, `examples/`.
