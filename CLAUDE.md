# Sloplang

## Project Overview

An array-first programming language that transpiles to Go. Arrays are the universal primitive — numbers, strings, hashmaps, and structs are all arrays. All data/container operations use pure symbolic operators (no methods, no dot notation). Control flow uses keywords.

**Tech Stack:** Go (transpiler + runtime)

**Language spec:** `docs/PRD.md`

---

## Build Commands

```bash
go build -o slop .                     # Build the transpiler (standalone binary)
./slop <file.slop>                     # Transpile + compile + run a .slop file
go run . <file.slop>                   # Same, without building first
```

---

## Lint, Format & Type Check

```bash
go fmt ./...                    # Format Go code
go vet ./...                    # Static analysis
```

---

## Testing

```bash
go test ./...                                   # All tests
go test ./pkg/lexer/...                         # Lexer tests only
go test ./pkg/parser/...                        # Parser tests only
go test ./pkg/codegen/...                       # Codegen tests only
go test ./pkg/runtime/...                       # Runtime tests only
go test -run TestName ./pkg/lexer/...           # Single test by name
```

**E2E testing:** Transpile a `.slop` file, compile the Go output, run it, diff stdout against expected output.

## General Guidelines
- All of the documentation lives inside `docs` directory.
- The PRD is located at `docs/PRD.md`.
- For each mistake, wrong approach and assumption you make, record it in `docs/patterns.md`. ALWAYS READ IT IN START.
  - The COST of not recording them is NOT learning from your experience. It leads to errors and frustration on the HUMAN development end.
- Whenever unsure about a library or framework, use the `context7` MCP server to search and fetch its documentation.
- The development roadmap is available at `docs/roadmap.md`. You MUST never update it WITHOUT asking the user first.

## Feature Implementation Guide
- Always plan before implementing a feature. Ask the user to explicitly improve the plan before you start writing code.
  - Small and explicit changes can be made without plan mode.
- Always write your plan to the `docs/plans/` directory with an appropriate file name.
- For each plan, you must create a JSON file corresponding to the plan file. It must contain all of the todo tasks and their status. The file must be a JSON array of feature objects following this schema:
  ```json
  [
    {
      "category": "lexer",
      "description": "lexer correctly tokenizes all multi-char operators",
      "steps": [
        "Tokenize source containing <<, >>, ::, ++, --, ??, ##, @@, ~@, <|, |>, <., .>, .>>, **, <-",
        "Verify each token type is correct",
        "Verify single-char operators are not confused with multi-char prefixes",
        "Verify // comments are skipped",
        "Verify string literals with escapes are handled"
      ],
      "passes": false
    },
    {
      "category": "parser",
      "description": "parser produces correct AST for nested expressions with operator precedence",
      "steps": [
        "Parse [1, 2] + [3, 4] and verify BinaryOp(Add) with two ArrayLit children",
        "Parse arr@0 and verify IndexExpr with numeric index",
        "Parse map@name and verify IndexExpr with literal string key",
        "Parse if/else and verify IfStmt with correct branches",
        "Parse fn declaration and verify FnDecl with params and body"
      ],
      "passes": false
    },
    {
      "category": "codegen",
      "description": "codegen produces valid Go source via go/ast lowering",
      "steps": [
        "Generate Go source for a simple assignment + stdout write",
        "Verify output compiles with go build",
        "Verify go/format produces correctly formatted output",
        "Verify runtime imports are included",
        "Verify function declarations lower to *ast.FuncDecl"
      ],
      "passes": false
    },
    {
      "category": "runtime",
      "description": "SlopValue arithmetic enforces same-length arrays and rejects mismatched types",
      "steps": [
        "Add two int64 arrays of same length and verify element-wise result",
        "Add two arrays of different lengths and verify runtime panic",
        "Add int64 array to float64 array and verify runtime panic",
        "Verify unary negate works on all numeric types",
        "Verify comparison operators reject multi-element arrays"
      ],
      "passes": false
    },
    {
      "category": "e2e",
      "description": "hello world program transpiles, compiles, and runs correctly",
      "steps": [
        "Write hello.slop with assignment and stdout write",
        "Run transpiler to produce .go output",
        "Compile the .go output with go build",
        "Run the binary and capture stdout",
        "Verify stdout matches expected output"
      ],
      "passes": false
    }
  ]
  ```
- After implementing a plan and verifying it, you must update the corresponding JSON file's passes from `false` to `true`. This is a MUST. Without it, no plan is considered **DONE**.
  **Rules for the feature list:**
  - Use JSON format, not Markdown.
  - You MUST ALWAYS flip `passes` from `false` to `true` after verifying end-to-end. Never remove, rewrite, or weaken existing test entries — doing so risks hiding broken functionality.


## Commit Guidelines
- Always commit after completing a single feature and verifying its functionality.
- Conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`
- Focused, atomic commits
