# Appendix C: Limitations and Workarounds

This appendix lists known limitations of sloplang and practical workarounds for each.

---

### 1. No network I/O

**Limitation:** Sloplang has no built-in socket, HTTP, or network primitives.

**Workaround:** Write data to a file, invoke an external tool (e.g., `curl`) via a shell wrapper, and read the result back with `<.`.

---

### 2. No exceptions

**Limitation:** There is no `try/catch` or exception mechanism. Runtime panics (e.g., type mismatch, null in arithmetic) terminate the program.

**Workaround:** Use the dual-return pattern. Functions that can fail return `(result, errcode)`. Check `err != [0]` before using the result. Builtins `to_num` and `<.` already follow this pattern.

---

### 3. No async or concurrency

**Limitation:** Sloplang is strictly sequential. There are no goroutines, threads, futures, or event loops.

**Workaround:** Structure programs sequentially. For workloads requiring parallelism, run multiple sloplang programs as separate processes and coordinate via files.

---

### 4. `++` is array concat, not string concatenation

**Limitation:** `"prefix: " ++ str(x)` does not produce a joined string. It produces a two-element array `["prefix: ", str(x)]`, which formats as `[prefix: , 5]`, not `"prefix: 5"`.

**Workaround:** Use multiple `|>` calls to print adjacent strings:
```
|> "prefix: "
|> str(x)
|> "\n"
```

---

### 5. No broadcasting — array lengths must match for arithmetic

**Limitation:** `[1, 2, 3] + [10]` panics at runtime. Arithmetic operators require both operands to be the same length.

**Workaround:** Ensure arrays have the same length before operating on them. Build same-length arrays explicitly, or operate element-by-element inside a loop.

---

### 6. `--x` is TOKEN_REMOVE, not double-negate

**Limitation:** The lexer greedily tokenizes `--` as the remove-value operator (`TOKEN_REMOVE`). Writing `--[5]` does not double-negate; it is a parse error or semantic error.

**Workaround:** Use `-(-x)` for double negation:
```
result = -(-[5])    // [5]
```

---

### 7. `[0]` panics in boolean context

**Limitation:** `IsTruthy` only accepts `[1]` (truthy) and `[]` (falsy). `[0]` panics with "use [] for false".

**Workaround:** Use `[]` for false and `[1]` (or `true`) for true. Never use `[0]` as a boolean. If you have a numeric zero that you want to test, compare it explicitly: `x == [0]`.

---

### 8. Strings panic in boolean context

**Limitation:** A string value such as `"hello"` panics when used in a boolean position (`if "hello"`, `&& "ok"`, etc.).

**Workaround:** Compare explicitly:
```
if s != "" {
    // s is non-empty
}
```

---

### 9. Floats panic in boolean context

**Limitation:** Float values such as `[3.14]` panic when used in a boolean position.

**Workaround:** Compare with a float literal:
```
if f != [0.0] {
    // f is nonzero
}
```

---

### 10. Null panics in arithmetic, ordered comparisons, iteration, and boolean context

**Limitation:** `[null]` cannot be added, subtracted, compared with `<`/`>`, negated, used as a loop target, or evaluated for truthiness. All of these operations panic.

**Workaround:** Check for null with `==` before use:
```
if val == [null] {
    |> "not set\n"
} else {
    |> str(val + [1])
}
```

---

### 11. Bare numbers outside `[]` are parse errors

**Limitation:** Writing `x = 42` or `y = 0` is rejected by the parser. Number literals are only valid inside array brackets.

**Workaround:** Always bracket numbers: `x = [42]`, `y = [0]`, `if count > [0]`.

---

### 12. Bare `null` outside `[]` is a parse error

**Limitation:** Writing `x = null` is rejected. `null` is only valid inside array brackets.

**Workaround:** Always bracket null: `x = [null]`.

---

### 13. `[true]` creates `[[1]]`, not `[1]`

**Limitation:** `true` inside an array literal creates a nested value `[[1]]`, not the flat `[1]` you might expect. Similarly `[false]` creates `[[]]`.

**Workaround:** Use `true` and `false` as standalone expressions only. Do not place them inside `[]`. Use `[1]` and `[]` directly when you need boolean elements inside a larger array.

---

### 14. `str([42])` → `"[42]"` breaks numeric file roundtrip

**Limitation:** `str([42])` produces `"[42]"` (with brackets). Writing that to a file and reading it back, then calling `to_num("[42]")`, fails because `to_num` cannot parse the brackets.

**Workaround:** There is no built-in way to format a number without brackets (`str(val@0)` still produces `"[42]"` because `@0` re-wraps the element). Keep values as strings throughout if they will be written to files and parsed back. For example, store `"42"` (a string) instead of `[42]` (a number), and only use `to_num` when arithmetic is needed.

---

### 15. `#arr` cannot appear in slice postfix expressions

**Limitation:** `arr::mid::#arr` fails. The `#` is tokenized as a new unary prefix, not as part of the slice bounds.

**Workaround:** Store the length in a variable first:
```
len = #arr
result = arr::mid::len
```

---

### 16. `|>` has no trailing newline

**Limitation:** `|> "hello"` prints `hello` without a newline. Multiple `|>` statements run together on the same line.

**Workaround:** Append `"\n"` explicitly to any output that should end a line:
```
|> "hello\n"
```

---

### 17. `.>` and `.>>` panic on error (no dual-return)

**Limitation:** If the file path is unwritable or the directory does not exist, `.>` and `.>>` panic. There is no way to catch the error.

**Workaround:** Ensure the output path is writable before writing. If the path depends on user input, validate it first (e.g., attempt a small write and check with `<.`).

---

### 18. `split(str, "")` returns the original string, not a character array

**Limitation:** `split("abc", "")` returns `"abc"` unchanged. There is no built-in way to iterate over the individual characters of a string.

**Workaround:** There is no built-in character iterator in sloplang. Redesign algorithms to work on whole tokens (words, delimited fields) rather than individual characters.

---

### 19. `to_num` result type affects downstream arithmetic

**Limitation:** `to_num("3")` returns `int64`; `to_num("3.0")` returns `float64`. Adding an `int64` result to a `float64` operand panics with a type mismatch.

**Workaround:** Be explicit about the expected type. If the input could be either, normalize it: parse and then apply a float conversion via multiplication if needed, or ensure all operands in an arithmetic expression are the same type.

---

### 20. No stack traces on panic

**Limitation:** When a runtime panic occurs (null arithmetic, boolean type error, index out of bounds, etc.), the program terminates with a brief message and no stack trace visible to sloplang code.

**Workaround:** Add `|>` debug output before risky operations to narrow down which line is panicking:
```
|> "about to index\n"
x = arr@5
|> "indexed ok\n"
```

---

### 21. Functions are not first-class values

**Limitation:** Functions cannot be stored in variables, passed as arguments, or returned from other functions. There are no closures, lambdas, or higher-order functions.

**Workaround:** Dispatch manually using string keys in a hashmap, or write separate named functions and use `if`/`else` chains to select between them:
```
if op == "add" {
    result = add(a, b)
} else {
    result = sub(a, b)
}
```

---

### 22. No command-line argument access

**Limitation:** There is no built-in way to access command-line arguments. The `slop` CLI takes a source file but does not pass remaining arguments to the running program. There is no `os.Args` equivalent.

**Workaround:** Read input from stdin (`<|`) or from a configuration file (`<. "config.txt"`).

---

### 23. Hashmaps compare keys in insertion order for `==`

**Limitation:** Two hashmaps with the same keys and values but inserted in different orders may not compare as equal with `==`, because key comparison walks the key array in insertion order.

**Workaround:** Normalize insertion order when building hashmaps that will be compared. Always insert keys in the same deterministic sequence (e.g., alphabetical) so that structural equality is reliable.
