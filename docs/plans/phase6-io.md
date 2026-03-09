# Phase 6: I/O Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add file I/O operators (`<.`, `.>`, `.>>`), stdin reading (`<|`), and builtins (`split`, `to_num`) to complete the language's I/O operator set.

**Architecture:** Four new I/O tokens and AST nodes flow through the existing pipeline. Runtime I/O functions live in a new `io.go` file. `split` and `to_num` extend the builtin dispatch map in codegen. Dual-return ops (`<|`, `<.`, `to_num`) use the existing `MultiAssignStmt` + `UnpackTwo` pattern. File write/append are fire-and-forget statements.

**Tech Stack:** Go stdlib (`bufio`, `os`, `strings`, `strconv`)

---

### Task 1: Lexer — Add I/O tokens

**Files:**
- Modify: `sloplang/pkg/lexer/token.go:67-78` (add after `TOKEN_DOLLAR`, before keywords)
- Modify: `sloplang/pkg/lexer/token.go:81-136` (add to `tokenNames`)
- Modify: `sloplang/pkg/lexer/lexer.go:158-178` (extend `<` handling)
- Modify: `sloplang/pkg/lexer/lexer.go:284-302` (add `.` handling before `default`)
- Test: `sloplang/pkg/lexer/lexer_test.go`

**Step 1: Add token constants to `token.go`**

Add 4 new token types in the I/O section (after `TOKEN_DOLLAR`, before the Keywords section):

```
// I/O
TOKEN_STDIN_READ  // <|
TOKEN_FILE_READ   // <.
TOKEN_FILE_WRITE  // .>
TOKEN_FILE_APPEND // .>>
```

Add corresponding entries to `tokenNames` map:
```
TOKEN_STDIN_READ:  "STDIN_READ",
TOKEN_FILE_READ:   "FILE_READ",
TOKEN_FILE_WRITE:  "FILE_WRITE",
TOKEN_FILE_APPEND: "FILE_APPEND",
```

**Step 2: Add `<|` and `<.` to lexer `<` handling**

In `lexer.go` line 158, the `case l.ch == '<':` block currently checks `<=`, `<-`, `<<`. Add two more checks BEFORE the `else` fallthrough (order: `<=`, `<-`, `<<`, `<|`, `<.`, then `<`):

- If peek is `|`: emit `TOKEN_STDIN_READ` with literal `<|`
- If peek is `.`: emit `TOKEN_FILE_READ` with literal `<.`

**Step 3: Add `.>` and `.>>` handling to lexer**

Add a new `case l.ch == '.':` block BEFORE the `case l.ch == '"':` line (around line 285). Check:
1. If peek is `>` AND peekPeek (readPos+1) is `>`: emit `TOKEN_FILE_APPEND` with literal `.>>`, consume 3 chars
2. If peek is `>`: emit `TOKEN_FILE_WRITE` with literal `.>`, consume 2 chars
3. Else: emit `TOKEN_ILLEGAL`

Note: need a `peekCharAt(offset)` helper or just check `l.input[l.readPos+1]` with bounds check for the 3-char `.>>` case. Simplest: check peek is `>`, consume `.` and `>`, then check if NEW current char is `>` — if yes consume again for `.>>`, else it was `.>`.

**Step 4: Write lexer tests**

Test cases:
- `TestLexer_StdinRead`: tokenize `<|` → `TOKEN_STDIN_READ`
- `TestLexer_FileRead`: tokenize `<. "file.txt"` → `TOKEN_FILE_READ, TOKEN_STRING`
- `TestLexer_FileWrite`: tokenize `.> "file.txt" "data"` → `TOKEN_FILE_WRITE, TOKEN_STRING, TOKEN_STRING`
- `TestLexer_FileAppend`: tokenize `.>> "file.txt" "data"` → `TOKEN_FILE_APPEND, TOKEN_STRING, TOKEN_STRING`
- `TestLexer_IODisambiguation`: tokenize `<= <- << <| <.` → verify all 5 are correct
- `TestLexer_DotDisambiguation`: tokenize `.> .>>` → verify both correct

**Step 5: Run tests**

Run: `cd sloplang && go test ./pkg/lexer/... -v`
Expected: All pass.

**Step 6: Commit**

```bash
git add sloplang/pkg/lexer/
git commit -m "feat(lexer): add I/O tokens <|, <., .>, .>>"
```

---

### Task 2: Parser — Add I/O AST nodes and parsing

**Files:**
- Modify: `sloplang/pkg/parser/ast.go` (add 4 new AST node types after line 287)
- Modify: `sloplang/pkg/parser/parser.go:33-149` (add cases to `parseStatement()`)
- Test: `sloplang/pkg/parser/parser_test.go`

**Step 1: Add AST nodes to `ast.go`**

Add these after the `DynKeySetStmt` definition (line 287):

```go
// StdinReadExpr represents: <| (reads one line from stdin)
type StdinReadExpr struct{}
func (sr *StdinReadExpr) exprNode()            {}
func (sr *StdinReadExpr) TokenLiteral() string { return "<|" }

// FileReadExpr represents: <. path (reads entire file)
type FileReadExpr struct {
    Path Expr
}
func (fr *FileReadExpr) exprNode()            {}
func (fr *FileReadExpr) TokenLiteral() string { return "<." }

// FileWriteStmt represents: .> path data
type FileWriteStmt struct {
    Path Expr
    Data Expr
}
func (fw *FileWriteStmt) stmtNode()            {}
func (fw *FileWriteStmt) TokenLiteral() string { return ".>" }

// FileAppendStmt represents: .>> path data
type FileAppendStmt struct {
    Path Expr
    Data Expr
}
func (fa *FileAppendStmt) stmtNode()            {}
func (fa *FileAppendStmt) TokenLiteral() string { return ".>>" }
```

**Step 2: Add parser cases in `parseStatement()`**

In the `switch` block of `parseStatement()`, add these cases:

For file write/append (statements, no assignment):
```go
case lexer.TOKEN_FILE_WRITE:
    return p.parseFileWriteStmt()
case lexer.TOKEN_FILE_APPEND:
    return p.parseFileAppendStmt()
```

For stdin read and file read: these appear as the RHS of a multi-assign (`a, b = <|` or `a, b = <. "path"`). The existing `parseMultiAssign()` already handles `a, b = expr` — we just need `<|` and `<.` to be parseable as expressions.

Add `TOKEN_STDIN_READ` and `TOKEN_FILE_READ` to `parsePrimary()`:
```go
case lexer.TOKEN_STDIN_READ:
    p.advance()
    return &StdinReadExpr{}
case lexer.TOKEN_FILE_READ:
    p.advance() // consume <.
    path := p.parseExpression()
    if path == nil {
        return nil
    }
    return &FileReadExpr{Path: path}
```

**Step 3: Implement parse helpers**

```go
func (p *Parser) parseFileWriteStmt() *FileWriteStmt {
    p.advance() // consume .>
    path := p.parseExpression()
    if path == nil {
        return nil
    }
    data := p.parseExpression()
    if data == nil {
        return nil
    }
    return &FileWriteStmt{Path: path, Data: data}
}

func (p *Parser) parseFileAppendStmt() *FileAppendStmt {
    p.advance() // consume .>>
    path := p.parseExpression()
    if path == nil {
        return nil
    }
    data := p.parseExpression()
    if data == nil {
        return nil
    }
    return &FileAppendStmt{Path: path, Data: data}
}
```

**Important:** `parseFileWriteStmt` parses two consecutive expressions. The first is the path, the second is the data. Since expressions don't consume `=` or statement-level tokens, this should work naturally — `parseExpression()` stops at tokens it doesn't recognize.

**Step 4: Write parser tests**

Test cases:
- `TestParser_StdinRead`: parse `line, err = <|` → MultiAssignStmt with StdinReadExpr value
- `TestParser_FileRead`: parse `data, err = <. "file.txt"` → MultiAssignStmt with FileReadExpr (Path = StringLiteral)
- `TestParser_FileReadVar`: parse `data, err = <. path` → MultiAssignStmt with FileReadExpr (Path = Identifier)
- `TestParser_FileWrite`: parse `.> "file.txt" "data"` → FileWriteStmt
- `TestParser_FileWriteVar`: parse `.> path data` → FileWriteStmt with Identifiers
- `TestParser_FileAppend`: parse `.>> "file.txt" "more"` → FileAppendStmt

**Step 5: Run tests**

Run: `cd sloplang && go test ./pkg/parser/... -v`
Expected: All pass.

**Step 6: Commit**

```bash
git add sloplang/pkg/parser/
git commit -m "feat(parser): add I/O AST nodes and parsing for <|, <., .>, .>>"
```

---

### Task 3: Runtime — I/O functions and builtins

**Files:**
- Create: `sloplang/pkg/runtime/io.go`
- Test: `sloplang/pkg/runtime/io_test.go`

**Step 1: Create `io.go` with all runtime I/O functions**

New file `sloplang/pkg/runtime/io.go`:

```go
package runtime

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)
```

Functions to implement:

**`StdinRead()`** — returns `(*SlopValue, *SlopValue)`:
- Create a `bufio.Scanner` on `os.Stdin`, call `Scan()`
- On success: return `(NewSlopValue(scanner.Text()), NewSlopValue(int64(0)))`
- On EOF/error: return `(NewSlopValue(""), NewSlopValue(int64(1)))`

**`FileRead(path *SlopValue)`** — returns `(*SlopValue, *SlopValue)`:
- Extract string from path (must be single-element string, panic otherwise)
- `os.ReadFile(pathStr)`
- On success: return `(NewSlopValue(string(data)), NewSlopValue(int64(0)))`
- On error: return `(NewSlopValue(""), NewSlopValue(int64(1)))`

**`FileWrite(path, data *SlopValue)`** — no return:
- Extract string from path
- Format data with `FormatValue(data)`
- `os.WriteFile(pathStr, []byte(dataStr), 0644)`
- On error: panic

**`FileAppend(path, data *SlopValue)`** — no return:
- Extract string from path
- Format data with `FormatValue(data)`
- Open with `os.OpenFile(pathStr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)`
- Write, close
- On error: panic

**`Split(sv, sep *SlopValue)`** — returns `*SlopValue`:
- Extract strings from both args (single-element string, panic otherwise)
- If sep is empty string: return `NewSlopValue(str)` (original string as-is)
- Else: `strings.Split(str, sepStr)`, build SlopValue with each part as element

**`ToNum(sv *SlopValue)`** — returns `(*SlopValue, *SlopValue)`:
- Extract string from sv
- Try `strconv.ParseInt(str, 10, 64)` — if success, return `(NewSlopValue(int64(val)), NewSlopValue(int64(0)))`
- Try `strconv.ParseFloat(str, 64)` — if success, return `(NewSlopValue(float64(val)), NewSlopValue(int64(0)))`
- Return `(NewSlopValue(), NewSlopValue(int64(1)))`

Helper: `extractString(sv *SlopValue) string` — panics if not single-element string.

**Step 2: Write runtime tests**

Test `io_test.go` (skip stdin tests — hard to test in unit tests, covered in E2E):

- `TestFileReadWrite`: write file, read it back, verify content
- `TestFileAppend`: write then append, read back, verify both parts
- `TestFileReadMissing`: read nonexistent file, verify err is `[1]`
- `TestSplit_BySpace`: split `"a b c"` by `" "` → 3-element SlopValue with `"a"`, `"b"`, `"c"`
- `TestSplit_ByNewline`: split `"a\nb"` by `"\n"` → 2-element SlopValue
- `TestSplit_EmptySep`: split `"hello"` by `""` → single-element `"hello"`
- `TestSplit_NonStringPanic`: split with non-string arg → panic
- `TestToNum_Int`: `to_num("42")` → `([42], [0])`
- `TestToNum_Float`: `to_num("3.14")` → `([3.14], [0])`
- `TestToNum_Fail`: `to_num("abc")` → `([], [1])`
- `TestToNum_NegativeInt`: `to_num("-5")` → `([-5], [0])`

**Step 3: Run tests**

Run: `cd sloplang && go test ./pkg/runtime/... -v`
Expected: All pass.

**Step 4: Commit**

```bash
git add sloplang/pkg/runtime/io.go sloplang/pkg/runtime/io_test.go
git commit -m "feat(runtime): add I/O functions and builtins (split, to_num)"
```

---

### Task 4: Codegen — Lower I/O nodes and wire builtins

**Files:**
- Modify: `sloplang/pkg/codegen/codegen.go:76-160` (add cases to `lowerStmt`)
- Modify: `sloplang/pkg/codegen/codegen.go:310-385` (add cases to `lowerExpr`)
- Modify: `sloplang/pkg/codegen/codegen.go:373` (extend builtins map)
- Modify: `sloplang/pkg/codegen/codegen.go:275-299` (update `lowerMultiAssign` for dual-return)
- Test: `sloplang/pkg/codegen/codegen_test.go`

**Step 1: Add statement lowering for `.>` and `.>>`**

In `lowerStmt()`, add cases:

```go
case *parser.FileWriteStmt:
    return []ast.Stmt{
        &ast.ExprStmt{X: callSloprt("FileWrite", g.lowerExpr(s.Path), g.lowerExpr(s.Data))},
    }
case *parser.FileAppendStmt:
    return []ast.Stmt{
        &ast.ExprStmt{X: callSloprt("FileAppend", g.lowerExpr(s.Path), g.lowerExpr(s.Data))},
    }
```

**Step 2: Add expression lowering for `<|` and `<.`**

In `lowerExpr()`, add cases:

```go
case *parser.StdinReadExpr:
    return callSloprt("StdinRead")
case *parser.FileReadExpr:
    return callSloprt("FileRead", g.lowerExpr(e.Path))
```

**Step 3: Update `lowerMultiAssign` for dual-return functions**

Currently `lowerMultiAssign` wraps the RHS with `UnpackTwo(...)`. But `StdinRead()`, `FileRead()`, and `ToNum()` already return two values — they don't need `UnpackTwo`.

Modify `lowerMultiAssign` to detect dual-return expressions:

```go
func (g *Generator) isDualReturn(expr parser.Expr) bool {
    switch expr.(type) {
    case *parser.StdinReadExpr, *parser.FileReadExpr:
        return true
    case *parser.CallExpr:
        ce := expr.(*parser.CallExpr)
        return ce.Name == "to_num"
    }
    return false
}
```

In `lowerMultiAssign`:
- If `g.isDualReturn(s.Value)`: emit `a, b := loweredExpr` directly (no `UnpackTwo` wrapper)
- Else: keep existing `a, b := sloprt.UnpackTwo(loweredExpr)` behavior

**Step 4: Add builtins to dispatch map**

Update the builtins map at line 373:

```go
builtins := map[string]string{"str": "Str", "split": "Split"}
```

For `to_num`, since it's dual-return, it needs special handling. When `to_num(x)` appears in a `MultiAssignStmt`, it's already handled by `isDualReturn`. But what if someone uses `to_num(x)` in a single-assign? That would be an error since it returns two values. For now, add it to builtins map:

```go
builtins := map[string]string{"str": "Str", "split": "Split", "to_num": "ToNum"}
```

The Go compiler will catch if someone tries to use a dual-return in single-value context.

**Step 5: Write codegen tests**

- `TestCodegen_FileWrite`: verify `.> "f.txt" "data"` generates `sloprt.FileWrite(...)`
- `TestCodegen_FileAppend`: verify `.>> "f.txt" "data"` generates `sloprt.FileAppend(...)`
- `TestCodegen_FileRead`: verify `d, e = <. "f.txt"` generates `sloprt.FileRead(...)` without `UnpackTwo`
- `TestCodegen_StdinRead`: verify `l, e = <|` generates `sloprt.StdinRead()` without `UnpackTwo`
- `TestCodegen_Split`: verify `split(x, y)` generates `sloprt.Split(...)`
- `TestCodegen_ToNum`: verify `v, e = to_num(x)` generates `sloprt.ToNum(...)` without `UnpackTwo`

**Step 6: Run tests**

Run: `cd sloplang && go test ./pkg/codegen/... -v -run TestCodegen`
Expected: All pass.

**Step 7: Commit**

```bash
git add sloplang/pkg/codegen/
git commit -m "feat(codegen): lower I/O AST nodes and wire split/to_num builtins"
```

---

### Task 5: E2E tests

**Files:**
- Modify: `sloplang/pkg/codegen/codegen_e2e_test.go`

**Step 1: Add E2E helper for stdin tests**

The existing `runE2E` runs the binary with no stdin. Add `runE2EWithStdin(t, source, stdinInput string)` that pipes stdin to the process:

```go
func runE2EWithStdin(t *testing.T, source, stdinInput string) string {
    // Same as runE2E but set runCmd.Stdin = strings.NewReader(stdinInput)
}
```

**Step 2: Add file I/O E2E tests**

Note: E2E tests run the compiled binary in a temp dir. File paths in sloplang source are relative to where the binary runs. The `runE2E` helper sets `runCmd.Dir = tmpDir`, so file ops will use tmpDir as CWD. This means `.> "test.txt" "hello"` writes to `tmpDir/test.txt`.

Tests:
- `TestE2E_FileWriteRead`: `.> "test.txt" "hello"` then `data, err = <. "test.txt"` then `|> data` → `hello`
- `TestE2E_FileAppend`: write then append then read → both parts present
- `TestE2E_FileReadMissing`: `data, err = <. "nofile.txt"` then `|> str(err)` → `1`
- `TestE2E_FileWriteOverwrite`: write twice to same file, read back → only second content
- `TestE2E_FileWriteVariable`: use variable as path, verify it works

**Step 3: Add stdin E2E test**

- `TestE2E_StdinRead`: `line, err = <|` then `|> line` with stdin `"hello\n"` → `hello`
- `TestE2E_StdinReadEOF`: `line, err = <|` with empty stdin → err is `1`

**Step 4: Add split E2E tests**

- `TestE2E_SplitBySpace`: `words = split("a b c", " ")` then `|> str(words)` → `[a, b, c]`
- `TestE2E_SplitByNewline`: `lines = split("a\nb", "\n")` then `|> str(lines)` → `[a, b]`
- `TestE2E_SplitEmptySep`: `x = split("hello", "")` then `|> str(x)` → `hello`

**Step 5: Add to_num E2E tests**

- `TestE2E_ToNumInt`: `val, err = to_num("42")` then `|> str(val)` → `42`
- `TestE2E_ToNumFloat`: `val, err = to_num("3.14")` then `|> str(val)` → `3.14`
- `TestE2E_ToNumFail`: `val, err = to_num("abc")` then `|> str(err)` → `1`

**Step 6: Add combined pipeline test**

- `TestE2E_WriteSplitToNum`: write `"10 20 30"` to file, read back, split by space, iterate and to_num each, sum them, verify output

**Step 7: Run all E2E tests**

Run: `cd sloplang && go test ./pkg/codegen/... -v -run TestE2E`
Expected: All pass.

**Step 8: Commit**

```bash
git add sloplang/pkg/codegen/codegen_e2e_test.go
git commit -m "test(e2e): add E2E tests for file I/O, stdin, split, to_num"
```

---

### Task 6: Example file and docs update

**Files:**
- Create: `sloplang/examples/fileio.slop`
- Modify: `docs/architecture.md`
- Modify: `docs/plans/phase6-io.json` (flip all passes to true)

**Step 1: Create `fileio.slop` example**

```
// Write to file
.> "demo.txt" "hello file"

// Read from file
data, err = <. "demo.txt"
if err != [0] {
    |> "read failed"
}
|> data

// Append to file
.>> "demo.txt" "\nline two"

// Read again
data2, err2 = <. "demo.txt"
|> data2

// Split
words = split("the cat sat on the mat", " ")
|> str(words)

// to_num
val, nerr = to_num("42")
|> str(val)

badval, berr = to_num("abc")
|> str(berr)
```

Verify: `cd sloplang && go run ./cmd/slop/main.go examples/fileio.slop && go run examples/fileio.go`

**Step 2: Update `docs/architecture.md`**

- Add new token types to Token section
- Add new AST nodes to the lists
- Add I/O operations table to Runtime section
- Add `split`, `to_num` to builtins list
- Update Phase 6 status from "Planned" to "Done"
- Add `io.go` to directory structure

**Step 3: Flip all JSON passes to true**

Update `docs/plans/phase6-io.json` — set all `"passes": true`.

**Step 4: Commit**

```bash
git add sloplang/examples/fileio.slop docs/architecture.md docs/plans/phase6-io.json
git commit -m "feat: complete Phase 6 I/O — examples, docs, tracking"
```

---

### Important Notes for Implementer

1. **Lexer `.` vs float disambiguation:** The `.` character also appears in float literals (`3.14`), but floats are parsed in `readNumber()` which is reached via the `isDigit(l.ch)` path — not the `.` path. So `.>` and `.>>` won't conflict with floats.

2. **`parseFileWriteStmt` parses two expressions:** Make sure `parseExpression()` stops at the right boundary. Since `.> "path" "data"` has two string literals, the first `parseExpression()` will consume just `"path"` (a primary expression) and stop when it sees `"data"` (which can't continue the expression).

3. **Dual-return in `lowerMultiAssign`:** The critical change is skipping `UnpackTwo` for I/O expressions and `to_num`. The runtime functions directly return `(*SlopValue, *SlopValue)` instead of a packed `*SlopValue` that needs unpacking.

4. **Stdin E2E testing:** The `runE2EWithStdin` helper must pipe stdin before starting the process. Use `cmd.Stdin = strings.NewReader(input)`.

5. **File E2E test isolation:** Each E2E test runs in its own temp dir (via `t.TempDir()`). File ops use relative paths, so they'll be isolated. But verify that `runCmd.Dir = tmpDir` is set so the binary runs from the temp dir.

6. **`FormatValue` for file write data:** When writing `data` to a file, the runtime should use `FormatValue(data)` to convert the SlopValue to a string. This means `.> "f.txt" [1, 2, 3]` writes `[1, 2, 3]` to the file, and `.> "f.txt" "hello"` writes `hello`.
