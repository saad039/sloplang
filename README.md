# sloplang

A programming language that takes every best practice established in language design and throws them into the dustbin.

**Everything is an array.** Numbers? Arrays. Booleans? Arrays. Matrices? Arrays of arrays. Hashmaps? Arrays with string keys. Structs? Just hashmaps. Strings? Okay fine, strings are strings — but they live inside arrays too. There is one data structure and it's an array. `[42]` is a number. `[1, 2, 3]` is a list. `[[1, 2], [3, 4]]` is a matrix. `{"name": "bob", "age": [30]}` is a hashmap (which is an array). Arithmetic is element-wise — `[1] + [1]` gives you `[2]`, `[2, 3] * [4, 5]` gives you `[8, 15]`. Want to add two numbers? Wrap them in brackets first. Welcome to the slop.

## Design Philosophy

sloplang exists to stress-test LLMs. Modern language models have internalized decades of programming conventions — familiar syntax, standard operators, predictable semantics. sloplang deliberately violates all of that to see what happens when the patterns break down.

No dot notation. No methods. No `print()`. You want stdout? `|>`. You want to return from a function? `<-`. Want to read a file? `<.`. Boolean true is `[1]`, false is `[]`, and if you try to use `[0]` as false it panics and tells you to use `[]` instead.

It transpiles to Go, because of course it does.

## Quick Tour

```
// hello world
|> "hello world\n"

// everything is an array
x = [42]
y = [10, 20, 30]

// arithmetic is element-wise
|> str([2, 3] * [4, 5])        // [8, 15]
|> "\n"

// functions
fn factorial(n) {
    if n <= [1] {
        <- [1]
    }
    <- n * factorial(n - [1])
}
|> str(factorial([5]))          // [120]
|> "\n"

// arrays are the only data structure
nums = [1, 2, 3]
nums << [4]                     // push
last = >>nums                   // pop
|> str(nums ?? [2])             // contains? -> [1]
|> "\n"

// hashmaps are just arrays with keys
person{name, age} = ["bob", [30]]
|> person@name                  // bob
|> "\n"

// error handling — dual return, no exceptions
data, err = <. "file.txt"
if err != [0] {
    |> "failed\n"
}
```

## Operators

Everything is symbolic. No method calls. No `.length`. No `.push()`. Just symbols.

| Op | What it does | Example |
|----|-------------|---------|
| `+` `-` `*` `/` `%` `**` | Element-wise arithmetic | `[2] ** [3]` -> `[8]` |
| `==` `!=` `<` `>` `<=` `>=` | Comparisons (returns `[1]` or `[]`) | `[1] < [2]` -> `[1]` |
| `&&` `\|\|` `!` | Logical ops | `![]` -> `[1]` |
| `#` | Length (prefix) | `#[1,2,3]` -> `[3]` |
| `@` | Index / key access | `arr@0`, `map@name` |
| `$` | Dynamic access | `arr$i` (int->index, string->key) |
| `<<` | Push | `arr << [5]` |
| `<<<` | Nested push | `arr <<< [3,4]` -> `[..., [3, 4]]` |
| `>>` | Pop (prefix) | `x = >>arr` |
| `~@` | Remove at index | `arr ~@ [1]` |
| `::` | Slice | `arr::1::4` |
| `++` | Concat | `[1,2] ++ [3,4]` |
| `--` | Remove value | `arr -- [5]` |
| `~` | Unique (prefix) | `~[1,1,2]` -> `[1,2]` |
| `??` | Contains | `arr ?? [5]` -> `[1]` or `[]` |
| `##` | Keys (prefix) | `##map` |
| `@@` | Values (prefix) | `@@map` |
| `\|>` | Write stdout | `\|> "hello"` |
| `<\|` | Read stdin | `line = <\|` |
| `<.` | Read file | `data, err = <. "f.txt"` |
| `.>` | Write file | `.> "f.txt" data` |
| `.>>` | Append file | `.>> "f.txt" data` |
| `<-` | Return | `<- value` |

## Build

Requires Go 1.21+.

```bash
cd sloplang

# build the standalone binary (embeds runtime, no external Go dependency needed to run)
go build -o slop .

# or run directly
go run . program.slop
```

## Usage

```bash
# transpile and run a .slop file in one step
./slop program.slop
```

The CLI transpiles your `.slop` source to Go, compiles it with the embedded runtime, runs the binary, and cleans up. No Go installation needed on the target machine — the `slop` binary is fully self-contained.

## Run the Examples

```bash
cd sloplang

./slop examples/hello.slop
./slop examples/arrays.slop
./slop examples/maps.slop
./slop examples/fns.slop
./slop examples/errors.slop
./slop examples/fileio.slop
```

10 example programs covering: hello world, arithmetic, arrays, comparisons, strings, functions, hashmaps, null handling, error handling, and file I/O.

## Real Programs

10 non-trivial programs in `tests/programs/` that prove sloplang can do actual work:

- **fibonacci** — iterative Fibonacci sequence
- **wordcount** — word frequency counter
- **array_ops** — array manipulation demo
- **linear_search** / **binary_search** — search algorithms
- **bubble_sort** / **merge_sort** / **quick_sort** / **heap_sort** — sorting algorithms
- **bst** — binary search tree (insert, search, in-order traversal)

All verified by E2E tests that transpile, compile, run, and diff output.

## Tests

1,450+ tests across 5 packages:

```bash
cd sloplang

go test ./...                           # all tests
go test ./pkg/lexer/...                 # lexer only
go test ./pkg/parser/...                # parser only
go test ./pkg/codegen/...               # codegen + e2e tests
go test ./pkg/runtime/...               # runtime only
go test ./tests/programs/...            # real program e2e tests
```

The test suite includes:
- **Unit tests** — lexer tokenization, parser AST construction, runtime operations
- **E2E tests** — full pipeline (source -> lex -> parse -> codegen -> compile -> run -> verify output)
- **Semantic tests** — 355 tests covering mutation, equality, formatting, booleans, null, arithmetic, array ops, hashmaps, control flow
- **Adversarial tests** — 285 edge-case tests across syntax, types, boundaries, scoping, nesting, error recovery, identifiers, strings, hashmaps, interactions, mutation during iteration, and spec compliance
- **Snapshot tests** — panic messages verified against golden files for stable error formatting

## Project Structure

```
sloplang/
  main.go               # standalone CLI (embeds runtime via go:embed)
  pkg/
    lexer/               # tokenizer
    parser/              # recursive descent parser, AST types
    codegen/             # Go AST code generator + all E2E tests
    runtime/             # SlopValue type, ops, I/O, builtins
  tests/programs/        # 10 real .slop programs + test harness
  examples/              # 10 .slop example programs
  go.mod
docs/
  PRD.md                 # full language spec
  architecture.md        # transpiler architecture
  plans/                 # implementation plans (phases 1-9)
  patterns.md            # lessons learned
```

## Language Features

- **Array-first data model** — all values are arrays, arithmetic is element-wise
- **Strict booleans** — only `[1]` is truthy, only `[]` is falsy, everything else panics
- **Symbolic operators** — 26 operators, zero methods, zero dot notation
- **Hashmaps** — declared with `{}` syntax, keys are always strings
- **Null** — strict semantics, must be bracketed (`[null]`), panics on arithmetic/boolean use
- **Dual-return error handling** — Go-style `(result, errcode)`, no exceptions
- **File and stdin I/O** — `<.`, `.>`, `.>>`, `<|` operators
- **Scientific notation** — `[1.79e308]`, `[5e-324]` float literals
- **Go keyword safety** — sloplang variables can use Go keywords (`func`, `var`, `range`, etc.) without conflict
- **Type casting builtins** — `to_chars`, `to_int`, `to_float`, `fmt_float` for string decomposition, numeric conversion, and float formatting
- **Clean error messages** — division by zero, modulo by zero, integer overflow, type mismatches, and use-before-assign all produce readable sloplang errors instead of raw Go panics

## License

Do whatever you want with it. It's slop.
