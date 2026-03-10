# sloplang

A programming language that takes every best practice established in language design and throws them into the dustbin.

**Everything is an array.** Numbers? Arrays. Booleans? Arrays. Strings? Okay fine, strings are strings. But everything else is an array. Arithmetic is element-wise. `[1] + [1]` gives you `[2]`. Want to add two numbers? Wrap them in brackets first. Welcome to the slop.

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

# build the transpiler
go build -o slop ./cmd/slop/...

# or run directly
go run ./cmd/slop/main.go <file.slop>
```

## Usage

```bash
# transpile a .slop file to Go
./slop program.slop
# -> outputs program.go

# compile and run the generated Go
go run program.go
```

The transpiler reads `<file>.slop`, generates `<file>.go` in the same directory, and prints the output path. The generated Go file imports the sloplang runtime, so make sure the module is available.

## Run the Examples

```bash
cd sloplang

# transpile an example
go run ./cmd/slop/main.go examples/hello.slop

# run the generated Go
go run examples/hello.go
```

## Tests

```bash
cd sloplang

go test ./...                           # all tests
go test ./pkg/lexer/...                 # lexer only
go test ./pkg/parser/...                # parser only
go test ./pkg/codegen/...               # codegen + e2e tests
go test ./pkg/runtime/...               # runtime only
```

## Project Structure

```
sloplang/
  cmd/slop/         # CLI entrypoint
  pkg/
    lexer/          # tokenizer
    parser/         # AST builder
    codegen/        # Go code generator
    runtime/        # SlopValue runtime (arrays, ops, I/O)
  examples/         # .slop example programs
docs/
  PRD.md            # full language spec
```

## License

Do whatever you want with it. It's slop.