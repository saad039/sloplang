# Sloplang Implementation Roadmap

Each phase is incremental. Every phase produces a working subset of the language with e2e tests that transpile `.slop` files, compile the Go output, run it, and verify correctness.

## Phase 1: Skeleton + Assign + Stdout

Get the pipeline working end-to-end with the smallest possible program.

**Features:**
- Go module init, project structure, CLI entry point
- Runtime: `SlopValue` struct (int64, uint64, float64, string, elements, keys), `IsTruthy()`, `StdoutWrite()`
- Lexer: number literals (int, uint `u` suffix, float), string literals, `=`, `|>`, `[`, `]`, `,`, `//` comments, identifiers
- Parser: assignment statements, array literals, string literals, `|>` stdout write
- Codegen: go/ast lowering for assignments and stdout writes

**E2E test — `hello.slop`:**
```
x = [1, 2, 3]
|> "hello world"
```
Transpile -> compile -> run -> verify stdout is `hello world`.

---

## Phase 2: Arithmetic + Comparisons + Booleans

Make numbers computable.

**Features:**
- Runtime: `Add()`, `Sub()`, `Mul()`, `Div()`, `Mod()`, `Pow()`, `Negate()` — element-wise, same-length enforced, no broadcasting. `Eq()`, `Neq()`, `Lt()`, `Gt()`, `Lte()`, `Gte()` — single-element only. Returns `[1]` or `[]`.
- Lexer: `+`, `-`, `*`, `/`, `%`, `**`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `!`, `true`, `false`
- Parser: binary expressions with precedence, unary negation/not, boolean literals
- Codegen: lower all arithmetic/comparison/logical ops to `sloprt.X()` calls

**E2E test — `math.slop`:**
```
x = [10, 20] + [3, 4]
|> str(x)
y = [5] > [3]
|> str(y)
z = -[1, 2, 3]
|> str(z)
```
Verify output: `[13, 24]`, `[1]`, `[-1, -2, -3]`.

---

## Phase 3: Functions + Return + Control Flow

Make the language programmable.

**Features:**
- Runtime: function call mechanics via `SlopFunc` type
- Lexer: `fn`, `if`, `else`, `for`, `in`, `break`, `<-`, `{`, `}`, `(`, `)`
- Parser: `fn` declarations, `if`/`else` statements, `for`/`in` loops, `for { }` infinite loops, `break`, `<-` return, function calls, multi-assignment (`a, b = call()`)
- Codegen: `*ast.FuncDecl`, `*ast.IfStmt`, `*ast.RangeStmt`, `*ast.ForStmt` (infinite loop), `*ast.BranchStmt` (break), closures as `*ast.FuncLit`
- Runtime builtins: `str()`

**E2E test — `fns.slop`:**
```
fn add(a, b) {
    <- a + b
}

result = add([3], [4])
|> str(result)

if [1] == [1] {
    |> "equal"
}

items = [10, 20, 30]
for item in items {
    |> str(item)
}
```
Verify output: `[7]`, `equal`, `10`, `20`, `30`.

---

## Phase 4: Array Operators

Full array manipulation.

**Features:**
- Runtime: `Index()`, `IndexSet()`, `Push()`, `Pop()`, `RemoveAt()`, `Length()`, `Slice()`, `Concat()`, `Remove()`, `Contains()`, `Unique()`
- Lexer: `@`, `$`, `<<`, `>>`, `~@`, `::`, `++`, `--`, `~`, `??`, `#`
- Parser: index expressions (`arr@0`), key expressions (`map@name`), dynamic key (`map@$var`), push (`<<`), pop prefix (`>>`), remove-at (`~@`), slice (`::`, `++`, `--`), unary `#`, `~`, contains `??`
- Codegen: lower all array ops to `sloprt.X()` calls

**E2E test — `arrays.slop`:**
```
arr = [10, 20, 30, 40]
|> str(arr@0)
|> str(#arr)

arr << [50]
|> str(arr)

sub = arr::1::3
|> str(sub)

combined = [1, 2] ++ [3, 4]
|> str(combined)

has = arr ?? [20]
|> str(has)

uniq = ~[1, 2, 2, 3, 1]
|> str(uniq)
```
Verify output: `10`, `[5]`, `[10, 20, 30, 40, 50]`, `[20, 30]`, `[1, 2, 3, 4]`, `[1]`, `[1, 2, 3]`.

---

## Phase 5: Hashmaps

Arrays with string keys.

**Features:**
- Runtime: `MapFromKeysValues()`, `IndexKey()`, `IndexKeySet()`, `Keys()`, `Values()`
- Lexer: `##`, `@@` (already added `{`, `}` in Phase 3)
- Parser: hashmap declaration (`data{k1, k2} = [v1, v2]`), key access (`map@name`), dynamic key (`map@$var`), key set (`map@name = val`), `##` prefix, `@@` prefix, `for k in ##map`
- Codegen: lower hashmap ops

**E2E test — `maps.slop`:**
```
person{name, age} = ["bob", [30]]
|> person@name
|> str(person@age)

person@age = [31]
|> str(person@age)

person@email = "bob@test.com"
|> person@email

ks = ##person
|> str(ks)

vs = @@person
|> str(vs)
```
Verify output: `bob`, `[30]`, `[31]`, `bob@test.com`, `[name, age, email]`, `[bob, [31], bob@test.com]`.

---

## Phase 6: I/O (stdin + file)

Complete the I/O operator set.

**Features:**
- Runtime: `StdinRead()`, `FileRead()`, `FileWrite()`, `FileAppend()` — all unbuffered
- Lexer: `<|`, `<.`, `.>`, `.>>` (stdout `|>` already done in Phase 1)
- Parser: stdin read (`<|`), file read (`<. "path"`), file write (`.> "path" data`), file append (`.>> "path" data`), dual return for file ops
- Codegen: lower I/O ops
- Runtime builtins: `split()`, `to_num()`

**E2E test — `fileio.slop`:**
```
.> "test_out.txt" "hello file"
data, err = <. "test_out.txt"
if err != [0] {
    |> "read failed"
    <- [[], [1]]
}
|> data

.>> "test_out.txt" "\nline two"
data2, err2 = <. "test_out.txt"
|> data2
```
Verify: `test_out.txt` is created, stdout shows `hello file` then `hello file\nline two`. Clean up temp file after test.

---

## Phase 7: Error Handling Patterns

Dual return error propagation working correctly.

**Features:**
- Verify multi-assign with error codes works across function boundaries
- Runtime error codes for: mismatched array lengths in arithmetic, multi-element comparison, file not found, division by zero

**E2E test — `errors.slop`:**
```
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([10], [2])
|> str(result)
|> str(err)

result2, err2 = safe_div([10], [0])
|> str(result2)
|> str(err2)

data, err3 = <. "nonexistent.txt"
|> str(err3)
```
Verify output: `[5]`, `[0]`, `[]`, `[1]`, nonzero error code for missing file.

---

## Phase 8: Real Programs

Write actual sloplang programs and verify they produce correct output. No new features — this phase exercises everything together.

**Programs:**

**8a. Fibonacci — `examples/fibonacci.slop`:**
```
fn fib(n) {
    if n <= [1] {
        <- n
    }
    <- fib(n - [1]) + fib(n - [1] - [1])
}

i = [0]
for i in [0, 1, 2, 3, 4, 5, 6, 7, 8, 9] {
    |> str(fib(i))
}
```
Verify output: `0 1 1 2 3 5 8 13 21 34` (one per line).

**8b. Word frequency counter — `examples/wordcount.slop`:**
```
fn main() {
    .> "wc_input.txt" "the cat sat on the mat the cat"
    data, err = <. "wc_input.txt"
    if err != [0] {
        |> "cannot read file"
        <- [[], [1]]
    }

    words = split(data, " ")
    counts{} = []
    for w in words {
        if counts ?? w {
            counts@$w = counts@$w + [1]
        } else {
            counts@$w = [1]
        }
    }

    for k in ##counts {
        |> k ++ ": " ++ str(counts@$k)
    }
}
```
Verify output contains `the: [3]`, `cat: [2]`, `sat: [1]`, `on: [1]`, `mat: [1]`.

**8c. Array manipulation — `examples/array_ops.slop`:**
```
fn main() {
    data = [5, 3, 1, 4, 2, 3, 5, 1]
    |> "original: " ++ str(data)
    |> "length: " ++ str(#data)
    |> "unique: " ++ str(~data)
    |> "contains 4: " ++ str(data ?? [4])
    |> "contains 9: " ++ str(data ?? [9])
    |> "slice 2..5: " ++ str(data::2::5)

    data << [99]
    |> "after push: " ++ str(data)

    popped = >>data
    |> "popped: " ++ str(popped)
    |> "after pop: " ++ str(data)

    removed = data ~@ 0
    |> "removed at 0: " ++ str(removed)
    |> "after remove: " ++ str(data)
}
```
Verify each line of output matches expected values.

**8d. Linear search — `examples/linear_search.slop`**
Verify: finds existing element returning its index, returns error for missing element.

**8e. Binary search — `examples/binary_search.slop`**
Verify: finds element in sorted array, returns error for missing element.

**8f. Bubble sort — `examples/bubble_sort.slop`**
Verify: output is sorted ascending.

**8g. Merge sort — `examples/merge_sort.slop`**
Verify: output is sorted ascending.

**8h. Quick sort — `examples/quick_sort.slop`**
Verify: output is sorted ascending.

**8i. Heap sort — `examples/heap_sort.slop`**
Verify: output is sorted ascending.

**8j. Binary search tree — `examples/bst.slop`**
Verify: in-order traversal produces sorted output, search finds existing keys, returns empty for missing keys.

Each program is transpiled, compiled, run, and output is diffed against expected output.
