# Chapter 8: Input and Output

Sloplang provides four I/O primitives: stdout writes, stdin reads, file reads, and file writes. All I/O is unbuffered and immediate. Reading from stdin and files uses a dual-return pattern that signals success or failure; writing operations panic on error with no way to catch them.

## 8.1 Writing to Stdout: `|>`

The `|>` operator writes a value to stdout without a trailing newline.

```
|> "hello"
|> " "
|> "world"
|> "\n"
// output: hello world
```

To terminate lines, you must explicitly include `"\n"`:

```
|> "line 1"
|> "\n"
|> "line 2"
|> "\n"
```

> **Pitfall:** `|>` writes no trailing newline. Always include `"\n"` at the end of each line you want terminated.

> **Pitfall:** `++` cannot build strings. `"count: " ++ str([42])` produces a 2-element array `["count: ", "[42]"]`, not a concatenated string. Use separate `|>` calls instead:
> ```
> |> "count: "
> |> str([42])
> |> "\n"
> ```

## 8.2 Reading from Stdin: `<|`

The `<|` operator reads one line from stdin. It uses a dual-return pattern: the first return value is the line content, and the second signals success or error.

- Success: `(line_string, [0])`
- EOF or error: `("", [1])` — the error value is EMPTY STRING `""`, NOT null

Typical usage:

```
line, err = <|
if err == [0] {
    |> line
    |> "\n"
} else {
    |> "EOF\n"
}
```

Reading in a loop until EOF:

```
line, err = <|
for {
    if err != [0] { break }
    |> line
    |> "\n"
    line, err = <|
}
```

## 8.3 Reading Files: `<.`

The `<.` operator reads the entire contents of a file. It uses dual-return: the first return value is the file contents, and the second signals success or error.

- Success: `(file_contents_string, [0])`
- Error: `("", [1])` — empty string, NOT null

```
data, err = <. "input.txt"
if err != [0] {
    |> "cannot read file\n"
} else {
    |> data
    |> "\n"
}
```

## 8.4 Writing Files: `.>` and `.>>`

The `.>` operator writes data to a file (truncate if it exists, create if it doesn't):

```
.> "out.txt" "first line\n"
```

The `.>>` operator appends data to a file (create if it doesn't exist):

```
.>> "out.txt" "second line\n"
```

**Important:** Both `.>` and `.>>` do NOT dual-return. They PANIC on error. There is no way to catch a write failure in sloplang — design programs to write to paths that are known to be writable.

## 8.5 The `str()` Builtin

The `str()` builtin converts a value to its string representation. The rule is:

- Single-element strings → raw (no brackets): the string content is printed directly
- Everything else → bracket notation: `[42]`, `[null]`, `[]`, `[1, 2, 3]`

```
|> str("hello")    // hello   (raw string)
|> "\n"
|> str([42])       // [42]   (bracketed)
|> "\n"
|> str([null])     // [null]  (bracketed)
|> "\n"
|> str([])         // []
|> "\n"
|> str([1, 2, 3])  // [1, 2, 3]
|> "\n"
```

> **Pitfall:** Writing a number to a file then reading it back breaks `to_num`:
> ```
> x = [42]
> .> "num.txt" str(x)      // writes "[42]" (with brackets)
> data, derr = <. "num.txt"
> n, err = to_num(data)    // FAILS — to_num("[42]") cannot parse "[42]"
> ```
> There is no built-in way to format a number without brackets. Workaround: keep values as strings throughout if they will be written to files and parsed back. For example, store `"42"` (a string) instead of `[42]` (a number), and use `to_num` only when you need to do arithmetic.

## 8.6 The `split()` Builtin

The `split(str, sep)` builtin splits a string on a separator and returns an array of substrings.

```
parts = split("a,b,c", ",")
|> str(parts)    // [a, b, c]
|> "\n"

words = split("hello world foo", " ")
|> str(words)    // [hello, world, foo]
|> "\n"
```

> **Pitfall:** `split(str, "")` with an empty separator does NOT split into characters. It returns the original string unchanged.
> ```
> weird = split("hello", "")
> |> str(weird)    // hello  (original string returned)
> |> "\n"
> ```

## 8.7 The `to_num()` Builtin

The `to_num(str_val)` builtin parses a string to a number. It uses dual-return: the first return value is the parsed number, and the second signals success or error.

- Success: `([n], [0])` — tries int64 first, then float64
- Failure: `([], [1])` — failure value is `[]` (empty array), NOT `[null]`, NOT `[0]`

The type matters: `"3"` parses as int64, while `"3.0"` parses as float64. Type affects downstream arithmetic.

```
n, err = to_num("42")
if err == [0] {
    |> str(n)    // [42]  (int64)
    |> "\n"
}

f, ferr = to_num("3.14")
if ferr == [0] {
    |> str(f)    // [3.14]  (float64)
    |> "\n"
}

bad, berr = to_num("xyz")
if berr != [0] {
    |> "parse failed\n"
    // bad == []  (empty array, not null)
}
```

## 8.8 Formatting Values for Output

All output in sloplang is governed by a single formatting rule, applied by `|>`, `str()`, and file write operators (`.>` and `.>>`):

- Single-element strings: printed raw — the string content appears directly
- Everything else: printed with brackets — `[42]`, `[[1, 2], [3, 4]]`, `[null]`, `[1, 2, 3]`

This is why `|> "alice"` prints `alice` (raw), but `|> str([42])` prints `[42]` (bracketed). Understanding this rule prevents surprises when your output includes numbers, arrays, or null values.

## 8.9 No Buffering

All I/O in sloplang is unbuffered:

- Every `|>` call immediately writes to stdout
- Every `.>>` call immediately appends to disk
- No flush operation exists — output appears as it is written
- This matters in long-running programs: output appears progressively as operations complete, not batched at program exit

This design simplifies I/O semantics and makes debugging easier: what you see in the output stream is exactly what happened in the order it happened.
