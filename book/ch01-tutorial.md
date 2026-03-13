# Chapter 1: A Tutorial Introduction

This chapter walks you through the core of Sloplang by example. Rather than
presenting a formal specification, it shows you real programs and explains what
each piece does. By the end you will have written a complete, useful program —
a word-frequency counter — and along the way you will have seen every major
feature of the language.

Sloplang is small. Its syntax is close to what you already know from scripting
languages, but with one big idea at the centre: **every value is an array**.
That single rule unifies numbers, strings, hashmaps, and compound data
structures into one coherent model. Once the rule clicks, the rest follows
naturally.

---

## 1.1 Hello, World

Every language tutorial starts here. In Sloplang, `|>` is the standard-output
write operator. It prints whatever is on its right to stdout, with no trailing
newline — you must add the newline yourself.

```
|> "hello, world\n"
```

That is the complete program. The string `"hello, world\n"` is written directly
to stdout; `\n` is the newline character. If you omit `\n`, the shell prompt
will appear on the same line as your output.

A program can contain any number of `|>` statements, and they concatenate their
output directly — there is no implicit separator.

```
|> "one"
|> " "
|> "two"
|> "\n"
// output: one two
```

---

## 1.2 Variables and the Array Model

The defining rule of Sloplang is: **every value is an array**. A single number
is a one-element array. A string is a one-element array whose element is a
string. Even the boolean constants `true` and `false` desugar to the arrays
`[1]` and `[]` respectively.

As a direct consequence, number literals must always be written inside square
brackets. Strings are the one exception — you write them with bare double
quotes, and Sloplang treats them as single-element string arrays automatically.

```
x      = [42]
pi     = [3.14]
name   = "alice"
coords = [10, 20]
```

`|>` prints its argument raw when the value is a single-element string.
For anything else you must call `str()` to convert it to a string first, and
the result will include the enclosing brackets — they are part of the
representation.

```
x    = [42]
name = "alice"

|> str(x)     // output: [42]
|> "\n"
|> name        // output: alice
|> "\n"
```

Notice that `str([42])` returns the string `"[42]"`, not `"42"`. The brackets
tell you that the value is still an array containing the integer 42. This
consistency means you never have to guess what `str()` will produce.

Multi-element arrays work the same way:

```
coords = [10, 20]
|> str(coords)
|> "\n"
// output: [10, 20]
```

---

## 1.3 Arithmetic

Arithmetic operators in Sloplang are **element-wise**: they apply the operation
to corresponding elements of two arrays. Both arrays must be the same length
and contain the same element type. Mixing integers and floats, or arrays of
different lengths, causes a runtime panic.

The basic operators are `+`, `-`, `*`, `/`, `%`, and `**` (exponentiation).
Integer division truncates toward zero, just as it does in C or Go.

```
a = [3, 4] + [1, 2]   // [4, 6]
b = [10]   - [3]       // [7]
c = [6]    * [7]       // [42]
d = [10]   / [3]       // [3]  (integer division)
e = [10]   % [3]       // [1]
f = [2]    ** [8]      // [256]
```

Because arithmetic is element-wise, you can operate on entire arrays with a
single expression:

```
xs  = [1, 2, 3]
ys  = [4, 5, 6]
sum = xs + ys

|> str(sum)
|> "\n"
// output: [5, 7, 9]
```

You can also mix scalar-style arrays with element-wise arrays, as long as
lengths match. Most of the time you will work with single-element arrays for
scalar arithmetic:

```
price    = [9]
quantity = [7]
total    = price * quantity

|> str(total)
|> "\n"
// output: [63]
```

---

## 1.4 Control Flow

Sloplang has three control-flow constructs: `if/else`, `for item in arr`, and
`for { }` (infinite loop with `break`). All three use familiar brace syntax.

### if / else

The condition of an `if` statement must evaluate to a **strict boolean**: the
array `[1]` is truthy, and the empty array `[]` is falsy. Any other value —
including `[0]`, a non-empty string, or a multi-element array — causes a
runtime panic. Comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) always
return `[1]` or `[]`, so they are safe to use directly as conditions.

```
score = [85]
if score >= [90] {
    |> "A\n"
} else {
    |> "B\n"
}
// output: B
```

There is no `else if` keyword. To chain conditions, nest a second `if` inside
the `else` block:

```
score = [72]
if score >= [90] {
    |> "A\n"
} else {
    if score >= [80] {
        |> "B\n"
    } else {
        |> "C\n"
    }
}
// output: C
```

### for-in loop

`for item in arr` iterates over every element of `arr`, binding each one to
`item` in turn. The loop variable is a SlopValue — it prints and behaves like
any other Sloplang value.

```
for item in [1, 2, 3] {
    |> str(item)
    |> "\n"
}
// output:
// [1]
// [2]
// [3]
```

### Infinite loop with break

`for { }` runs forever until a `break` statement is reached. This is the
standard pattern for loops that do not have a predetermined iteration count.

```
i = [0]
for {
    if i == [3] { break }
    |> str(i)
    |> "\n"
    i = i + [1]
}
// output:
// [0]
// [1]
// [2]
```

---

## 1.5 Functions

Functions are declared with `fn`, followed by a name, a parameter list in
parentheses, and a braced body. The `<-` operator returns a value from a
function. Functions may only be declared at the top level — they cannot be
stored in variables or passed as arguments.

```
fn greet(name) {
    |> "hello, "
    |> name
    |> "\n"
}

greet("world")
// output: hello, world
```

Parameters receive whatever the caller passes. Return values are sent back with
`<-` and can be used in expressions or assigned to variables.

```
fn add(a, b) {
    <- a + b
}

result = add([3], [4])
|> str(result)
|> "\n"
// output: [7]
```

Functions can call themselves recursively. The classic example is factorial:

```
fn factorial(n) {
    if n <= [1] { <- [1] }
    <- n * factorial(n - [1])
}

|> str(factorial([5]))
|> "\n"
// output: [120]
```

Each recursive call reduces `n` by one until the base case `n <= [1]` is
reached, at which point `[1]` is returned and the call stack unwinds,
multiplying as it goes.

---

## 1.6 Arrays

Arrays are the universal container. You have already seen array literals like
`[1, 2, 3]`. Sloplang provides dedicated operators for indexing, pushing,
popping, and measuring arrays — all without any method-call syntax.

`arr@0` reads element 0 (zero-based index). `arr << [val]` appends every
element of the right-hand array to `arr` — `arr << [40]` pushes the integer 40.
`>>arr` is the prefix pop operator: it removes and returns the last element,
mutating `arr` in place. `#arr` returns the length as a single-element integer array.

```
arr = [10, 20, 30]
|> str(arr@0)
|> "\n"
arr << [40]
|> str(arr)
|> "\n"
last = >>arr
|> str(last)
|> "\n"
|> str(#arr)
|> "\n"
// output:
// [10]
// [10, 20, 30, 40]
// [40]
// [3]
```

After pushing 40 and then popping it, the array is back to three elements,
so `#arr` is `[3]`. Note that `<<` spreads its right operand: `arr << [40, 50]`
would push both 40 and 50.

---

## 1.7 Hashmaps

A hashmap groups named fields into a single value. The declaration syntax names
both the variable and its keys in one statement. `@key` reads a literal key,
`map@key = val` writes it, and `map$var` does dynamic lookup when the key name
is in a variable (dispatches to key lookup if `var` is a string, numeric index
if `var` is an integer).

```
person{name, age} = ["alice", [30]]
|> person@name
|> "\n"
|> str(person@age)
|> "\n"
person@age = [31]
field = "name"
|> person$field
|> "\n"
// output:
// alice
// [30]
// alice
```

String fields print raw (no brackets) when written directly with `|>`. Numeric
fields need `str()` first. `##map` returns the list of all keys as an array of
strings; `##map ?? "key"` checks whether a key exists. Iterate all keys with
`for k in ##map`.

---

## 1.8 Input and Output

Sloplang has four I/O operators:

| Operator    | Direction | Description                             |
|-------------|-----------|------------------------------------------|
| `\|>`       | Out       | Write to stdout (no trailing newline)    |
| `<\|`       | In        | Read one line from stdin                 |
| `<. "path"` | In        | Read entire file contents                |
| `.> "path"` | Out       | Write (truncate) to file                 |
| `.>> "path"`| Out       | Append to file                           |

The write operators (`|>`, `.>`, `.>>`) return nothing. The read operators
(`<|` and `<.`) return **two values** — the data and an error code — because
reads can fail. This is the **dual-return pattern**, explained in Section 1.9.

```
// Write to stdout
|> "hello\n"

// Write to file
.> "out.txt" "some data\n"

// Read from file
data, err = <. "out.txt"
if err == [0] {
    |> data
    |> "\n"
}

// Read from stdin (dual-return)
line, err = <|
if err == [0] {
    |> line
    |> "\n"
}
```

`<|` returns `(line, [0])` on success and `("", [1])` on end-of-file or
error. `<. "path"` returns `(data, [0])` on success and `("", [1])` on
error. Both follow the exact same convention.

`.>` and `.>>` panic on error rather than returning an error code — they are
intended for cases where a write failure is truly unrecoverable.

---

## 1.9 Error Handling

Sloplang does not have exceptions. Instead it uses the **dual-return pattern**:
a function that can fail returns two values, conventionally `[result, errcode]`.
An error code of `[0]` means success; anything else means failure. The caller
checks the error code before using the result.

You implement this in your own functions by returning a two-element array with
`<-`:

```
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([10], [2])
if err == [0] {
    |> str(result)
    |> "\n"
}
// output: [5]
```

A note on the condition `if b == [0]`: this is valid because `==` is a
comparison operator that always returns `[1]` or `[]`. The condition is not
`[0]` itself — it is the result of the equality test, which is a proper
boolean. The panic from `IsTruthy` only fires when `[0]` is used *directly*
as a condition.

The built-in `to_num(s)` converts a string to a number and follows the same
convention:

```
// Builtin dual-return
n, nerr = to_num("42")
if nerr == [0] {
    |> str(n)
    |> "\n"
}
// output: [42]
```

If the string cannot be parsed as a number, `to_num` returns `([], [1])`.

You can propagate errors across function boundaries by forwarding the error code
back to the caller, allowing chains of fallible operations without deep nesting:

```
fn parse_int(s) {
    val, err = to_num(s)
    if err != [0] {
        <- [[], [1]]
    }
    <- [val, [0]]
}

x, xerr = parse_int("not a number")
if xerr != [0] {
    |> "parse failed\n"
}
// output: parse failed
```

---

## 1.10 A Complete Program — Word Frequency Counter

To close the chapter we will build a complete program that reads lines from
stdin, splits each line into words, counts how many times each word appears,
and prints the results. It uses nearly every feature covered in this chapter:
hashmaps, dual-return I/O, for-in loops, and the `split` builtin.

```
// Read all lines from stdin
counts{} = []
line, err = <|
for {
    if err != [0] { break }
    words = split(line, " ")
    for w in words {
        if ##counts ?? w {
            counts$w = counts$w + [1]
        } else {
            counts$w = [1]
        }
    }
    line, err = <|
}
for k in ##counts {
    |> k
    |> ": "
    |> str(counts$k)
    |> "\n"
}
```

Let's walk through it piece by piece.

### `counts{} = []`

This declares an empty hashmap named `counts`. The `{}` with no keys inside
creates a hashmap with no initial fields; `[]` is the empty array and serves
as the placeholder value. After this line, `counts` is ready to have keys
inserted dynamically.

### Reading lines in a loop

```
line, err = <|
for {
    if err != [0] { break }
    ...
    line, err = <|
}
```

The program reads the first line before entering the loop, then re-reads at
the bottom of each iteration. When `<|` reaches end-of-file it returns
`("", [1])`, so `err` becomes `[1]` and the `if err != [0]` check at the top
breaks out of the loop. This is the standard read-process-loop pattern in
Sloplang.

### Splitting and counting

```
words = split(line, " ")
for w in words {
    if ##counts ?? w {
        counts$w = counts$w + [1]
    } else {
        counts$w = [1]
    }
}
```

`split(line, " ")` breaks the line on single spaces and returns an array of
strings. The for-in loop binds each word to `w`. `##counts` gives the list of
all current keys in the hashmap; `##counts ?? w` checks whether `w` is already
a key. If it is, we increment the count using dynamic access `counts$w`. If it
is not, we set the count to `[1]` to record the first occurrence.

### Printing the results

```
for k in ##counts {
    |> k
    |> ": "
    |> str(counts$k)
    |> "\n"
}
```

`##counts` again gives the array of all keys. We iterate over them, printing
each key (a raw string, so no `str()` needed), the separator `": "`, and the
count (a number, so `str()` is required to convert it for output).

Given the input:

```
the cat sat on the mat
the cat sat
```

The output might look like (order may vary):

```
the: [3]
cat: [2]
sat: [2]
on: [1]
mat: [1]
```

---

## Summary

You have now seen all of Sloplang's core features:

- **Array model** — every value is an array; numbers require brackets
- **Arithmetic** — element-wise, same-length, same-type
- **Control flow** — `if/else`, `for item in arr`, `for { break }`
- **Functions** — `fn`, `<-` return, recursion
- **Array operators** — `@` index, `<<` push, `>>` pop, `#` length
- **Hashmaps** — `{}` declaration, `@key` literal access, `$var` dynamic access, `##` keys, `??` contains
- **I/O** — `|>` stdout, `<|` stdin, `<.` file read, `.>` file write
- **Error handling** — dual-return pattern with `[result, errcode]`

The next chapter covers the complete operator reference, including slicing,
set-like operations, and the full I/O model.
