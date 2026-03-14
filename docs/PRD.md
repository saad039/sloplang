# Sloplang

An array-first language that transpiles to Go.

## Core Philosophy

Arrays are the universal primitive. Numbers, strings, hashmaps, and structs are all arrays. All data/container operations use pure symbolic operators (no methods, no dot notation). Control flow uses keywords (if/else, for).

## Data Model

Everything is a `SlopValue`. All values are arrays.

### Numeric Types

Numbers map to Go types based on the literal form:

| Literal | Go Type | Rule |
|---------|---------|------|
| `42`, `-10` | `int64` | Whole numbers (default) |
| `42u`, `255u` | `uint64` | `u` suffix = unsigned |
| `3.14`, `2.0` | `float64` | Has decimal point = float |

Arrays can hold any mix of types:

```
[1, 2, 3]           // all int64
[1, 3.14, "hello"]  // mixed — totally fine
[[1, 2], "abc", 42u] // nested + mixed
```

Arithmetic operators work element-wise. Each pair of elements must be the same numeric type (`[1] + [1.0]` -> runtime error), but the array itself can hold whatever.

### Literals

```
// Numbers — always in brackets, type inferred from literal
x = [42]            // int64
y = [42u]           // uint64
z = [3.14]          // float64

// Strings — the only non-bracket literal
name = "hello"

// Arrays — nested to any depth
matrix = [[1, 2], [3, 4]]

// Hashmaps — declared with {} syntax, keys are always strings
person{name, age} = ["bob", [30]]

// Empty hashmap
counts{} = []
```

### Null

`null` represents an absent/placeholder value. It is strict and **must always be inside brackets**:
- `[null] == [null]` → `[1]` (truthy), `[null] != [5]` → `[1]`
- Arithmetic, negation, truthiness checks, logical ops, ordered comparisons, and for-in iteration on null all **panic**
- `str([null])` → `"null"`, `|> [null]` prints `null`
- `#[null, null]` → `[2]` — length counts null elements normally
- `[1, null] ?? [null]` → `[1]` — contains finds null

```
x = [null]            // forward declaration
x = [null, null]      // array with null elements
```

### Booleans (Strict)

- Only `[1]` is **truthy**
- Only `[]` (empty array) is **falsy**
- `[0]` is **not a valid boolean** — panics with "use [] for false"
- Multi-element arrays in boolean context **panic**
- Strings, floats, null in boolean context **panic**
- `true` is a reserved keyword that produces `[1]`, `false` produces `[]`
- `true`/`false` are standalone (no brackets needed): `if true { }`, `x = false`
- `[true]` creates `[[1]]` (nested), `[false]` creates `[[]]` (nested) — avoid this

### Comments

```
// single-line only
```

## Operators

### Arithmetic

Element-wise, binary, same-length arrays required. NO broadcasting. Mismatched lengths are a runtime error.

| Op | Name | Example |
|----|------|---------|
| `+` | Add | `[1,2] + [3,4]` -> `[4,6]` |
| `-` | Sub | `[5,3] - [1,1]` -> `[4,2]` |
| `*` | Mul | `[2,3] * [4,5]` -> `[8,15]` |
| `/` | Div | `[10,6] / [2,3]` -> `[5,2]` |
| `%` | Mod | `[7,5] % [3,2]` -> `[1,1]` |
| `**` | Pow | `[2,3] ** [3,2]` -> `[8,9]` |
| `-` | Negate (unary) | `-[1,2,3]` -> `[-1,-2,-3]` |

### Comparisons

`==` and `!=` perform **deep structural equality** on any-size arrays and hashmaps (compares lengths, keys, and all elements recursively). `<`, `>`, `<=`, `>=` require single-element arrays only.

| Op | Example | Result |
|----|---------|--------|
| `==` | `[2] == [2]` | `[1]` (truthy) |
| `==` | `[1, 2] == [1, 2]` | `[1]` (deep equal) |
| `!=` | `[2] != [3]` | `[1]` |
| `<` | `[1] < [2]` | `[1]` |
| `>` | `[2] > [1]` | `[1]` |
| `<=` | `[1] <= [1]` | `[1]` |
| `>=` | `[2] >= [1]` | `[1]` |

Returns `[1]` (truthy) or `[]` (falsy).

### Logical

Strict truthiness: only `[1]` is truthy, only `[]` (empty array) is falsy. `[0]`, multi-element arrays, strings, floats, and null all **panic** in boolean context.

| Op | Name |
|----|------|
| `&&` | And |
| `\|\|` | Or |
| `!` | Not |

### Array / Container

| Op | Name | Usage |
|----|------|-------|
| `#` | Length (prefix) | `#arr` -> `[3]` |
| `@` | Index / Key | `arr@0` (numeric index), `map@name` (literal string key) |
| `$` | Dynamic access | `map$var` (dispatches: int→index, string→key lookup) |
| `<<` | Push | `arr << [5]` |
| `<<<` | Nested push | `arr <<< [5]` (appends as single nested element, never spreads) |
| `>>` | Pop (prefix) | `x = >>arr` (removes + returns last element) |
| `~@` | Remove at index/key | `x = arr ~@ 2` (removes + returns element at index 2), `x = map ~@ "key"` (removes + returns value for key) |
| `::` | Slice | `arr::1::4` (elements at indices 1, 2, 3) |
| `++` | Concat | `[1,2] ++ [3,4]` -> `[1,2,3,4]` |
| `--` | Remove value | `arr -- [5]` (removes first occurrence) |
| `~` | Unique (prefix) | `~arr` (deep copy, drop keys, keep first occurrences) |
| `??` | Contains | `arr ?? [5]` -> `[1]` or `[]` |
| `##` | Keys (prefix) | `##map` -> array of key strings |
| `@@` | Values (prefix) | `@@map` -> array of values |

### I/O

No buffering on any I/O operation.

| Op | Name | Usage |
|----|------|-------|
| `<\|` | Read stdin | `line = <\|` (reads one line from stdin) |
| `<.` | Read file | `data, err = <. "file.txt"` (reads entire file) |
| `\|>` | Write stdout (no trailing newline) | `\|> "hello world"` — use `\|> "\n"` for newlines |
| `.>` | Write file | `.> "file.txt" data` |
| `.>>` | Append file | `.>> "file.txt" data` |

### Other

| Op | Name | Usage |
|----|------|-------|
| `<-` | Return | `<- value` |
| `=` | Assign | `x = [1, 2, 3]` |

## Keywords

`fn`, `if`, `else`, `for`, `in`, `break`, `true`, `false`, `null`

## Hashmap Declaration & Access

```
// Declare — keys are implicitly strings
person{name, age} = ["bob", [30]]

// Read — bare word after @ is a literal string key
n = person@name         // "bob"
a = person@age          // [30]

// Dynamic access — $ dispatches on variable type
which = "name"
n = person$which        // "bob" (string key → key lookup)

// Numeric index access
arr = [10, 20, 30]
x = arr@0               // [10]
i = [2]
x = arr$i               // [30] (int key → index access)

// Set existing key
person@age = [31]

// Add new key
person@email = "b@b.com"
```

## Error Handling

Go-style dual return. Functions that can fail return `(result, errcode)`. Errcode `[0]` = success, nonzero = failure. No exceptions.

```
data, err = <. "file.txt"
if err != [0] {
    |> "failed"
    <- [[], [1]]
}
```

## Functions

```
fn add(a, b) {
    <- a + b
}
result = add([3], [4])    // [7]
```

Functions are **not** first-class values. They cannot be stored in variables, passed as arguments, or returned from other functions. Functions must be declared at the top level with `fn` and called by name only.

## Built-in Functions

| Function | Description |
|----------|-------------|
| `split(str, sep)` | Split string by separator, returns array of strings |
| `str(val)` | Convert value to string — single-element strings return raw string, everything else uses bracket notation (`[42]`, `[1, 2]`, `[null]`, `[]`) |
| `to_num(str)` | Convert string to number — dual return `(val, err)`, err `[0]` = success, `[1]` = failure |
| `to_chars(str)` | Split string into array of single-character strings — `to_chars("hello")` -> `["h", "e", "l", "l", "o"]`. Panics if argument is not a string |
| `to_int(val)` | Convert single-element numeric or string value to int64 — float64 is truncated toward zero. Panics on invalid input |
| `to_float(val)` | Convert single-element numeric or string value to float64. Panics on invalid input |
| `fmt_float(val, decimals)` | Format numeric value as string with fixed decimal places — `fmt_float([3.14], [2])` -> `"3.14"`. Panics on invalid input |

## Control Flow

```
// if / else
if x > [0] {
    |> "positive"
} else {
    |> "not positive"
}

// for-in loop
for item in arr {
    |> str(item)
}

// infinite loop with break
i = [0]
for {
    if i == [10] {
        break
    }
    |> str(i)
    i = i + [1]
}
```

## Not Supported

- No network I/O
- No exceptions
- No async
- No methods / dot notation
- No broadcasting in arithmetic

## Example: Word Frequency Counter

```
fn main() {
    data, err = <. "input.txt"
    if err != [0] {
        |> "cannot read file"
        <- [[], [1]]
    }

    lines = split(data, "\n")
    counts{} = []
    for line in lines {
        words = split(line, " ")
        for w in words {
            if counts ?? w {
                counts$w = counts$w + [1]
            } else {
                counts$w = [1]
            }
        }
    }

    for k in ##counts {
        |> k
        |> str(counts$k)
    }
}
```

## Transpiler

Written in Go. Pipeline: `.slop` source -> Lexer -> Parser -> Codegen -> `.go` source + runtime -> `go build` -> binary.

### Project Structure

```
sloplang/
  cmd/slop/main.go              -- CLI entry point
  pkg/lexer/token.go            -- Token type definitions
  pkg/lexer/lexer.go            -- Tokenizer
  pkg/parser/ast.go             -- AST node types
  pkg/parser/parser.go          -- Recursive descent parser
  pkg/codegen/codegen.go        -- AST -> Go source emitter
  pkg/runtime/slop_value.go     -- SlopValue type, NewSlopValue, IsTruthy, FormatValue, Str
  pkg/runtime/ops.go            -- All operations: arithmetic, comparison, logical, array, hashmap
  pkg/runtime/io.go             -- I/O: StdinRead, FileRead, FileWrite, FileAppend, Split, ToNum
  tests/programs/               -- Phase 8 real programs (.slop) + test harness
  examples/hello.slop           -- Example programs
  go.mod
```
