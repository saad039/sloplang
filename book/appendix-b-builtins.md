# Appendix B: Built-in Functions

Sloplang provides eight built-in functions. They cannot be redefined or shadowed.

---

## `str(val)`

**Signature:** `str(val) -> string`

**Description:** Converts any `SlopValue` to its string representation.

**Returns:** Always succeeds. Returns a string.

**Formatting rules:**

| Input | Output |
|-------|--------|
| `str("hello")` | `"hello"` |
| `str([42])` | `"[42]"` |
| `str([1, 2, 3])` | `"[1, 2, 3]"` |
| `str([])` | `"[]"` |
| `str([null])` | `"[null]"` |
| `str([[1, 2], "x"])` | `"[[1, 2], x]"` |

**Notes:**

- A single-element string value prints without brackets: `str("hello")` → `hello`.
- Everything else is bracketed: integers, floats, multi-element arrays, empty arrays, null elements.
- This means file-roundtrip patterns that write `str([42])` and read back with `to_num` will fail, because `to_num("[42]")` cannot parse the brackets. There is no built-in way to format a number without brackets. Workaround: keep values as strings throughout if they will be written to files and parsed back (e.g., store `"42"` instead of `[42]`).

---

## `split(str, sep)`

**Signature:** `split(str, sep) -> array`

**Description:** Splits `str` on the separator `sep`. Returns an array of strings.

**Returns:** Always succeeds. Returns an array of strings.

**Examples:**

| Call | Result |
|------|--------|
| `split("a,b,c", ",")` | `["a", "b", "c"]` |
| `split("hello world", " ")` | `["hello", "world"]` |
| `split("abc", "")` | `"abc"` (original string, unchanged) |

**Notes:**

- If `sep` is `""` (empty string), the original string is returned unchanged. There is no built-in character iterator; sloplang does not split on individual characters.
- The result is an array of string elements, suitable for iteration with `for item in arr`.

---

## `to_num(str)`

**Signature:** `to_num(str) -> (val, err)` — dual-return

**Description:** Parses a string as a number. Tries `int64` first; if that fails, tries `float64`. Uses the multi-assignment form `val, err = to_num(s)`.

**Success return:** `(parsed_number, [0])` — the number is `int64` or `float64` depending on the input string.

**Failure return:** `([], [1])` — empty array and error code `[1]`.

**Examples:**

| Call | val | err |
|------|-----|-----|
| `to_num("3")` | `[3]` (int64) | `[0]` |
| `to_num("3.0")` | `[3.0]` (float64) | `[0]` |
| `to_num("abc")` | `[]` | `[1]` |
| `to_num("")` | `[]` | `[1]` |

**Usage pattern:**

```
val, err = to_num(s)
if err != [0] {
    |> "parse failed\n"
    <- [[], [1]]
}
// val is now a number
```

**Notes:**

- On failure the first return value is `[]` (empty array), not `[null]` and not `[0]`.
- The result type (`int64` vs `float64`) matters for downstream arithmetic. `[3]` (int64) and `[3.0]` (float64) are different types and cannot be added together without a type match.
- Check `err == [0]` (not `err != [1]`) for clarity.

---

## `exit(code)`

**Signature:** `exit(code)` — no return value

**Description:** Terminates the program immediately with the given integer exit code.

**Argument:** A single-element `int64` array, e.g., `[0]` for success or `[1]` for failure.

**Examples:**

```
exit([0])    // success
exit([1])    // generic failure
exit([2])    // specific error code
```

**Notes:**

- POSIX convention: `[0]` means success; any nonzero value means failure.
- The program stops immediately. There is no deferred cleanup in sloplang.
- `exit` does not return, so any code after it is unreachable.
- Use `exit` when you need to terminate from deep inside a function without threading return values back up through the call stack. For recoverable errors, prefer the dual-return pattern with `<-`.

---

## `to_chars(str)`

**Signature:** `to_chars(str) -> array`

**Description:** Splits a string into an array of single-character strings using Unicode rune iteration.

**Returns:** Always succeeds. Returns an array of single-character strings.

**Examples:**

| Call | Result |
|------|--------|
| `to_chars("hello")` | `["h", "e", "l", "l", "o"]` |
| `to_chars("")` | `[]` |

**Notes:**

- Panics on non-string input.
- Not dual-return.

---

## `to_int(val)`

**Signature:** `to_int(val) -> [int]`

**Description:** Converts a numeric or string value to `int64`. `float64` values are truncated toward zero.

**Returns:** Always succeeds on valid input. Returns a single-element `int64` array.

**Examples:**

| Call | Result |
|------|--------|
| `to_int([3.14])` | `[3]` |
| `to_int("5")` | `[5]` |
| `to_int([-2.9])` | `[-2]` |

**Notes:**

- Panics on invalid input.
- Not dual-return.

---

## `to_float(val)`

**Signature:** `to_float(val) -> [float]`

**Description:** Converts a numeric or string value to `float64`.

**Returns:** Always succeeds on valid input. Returns a single-element `float64` array.

**Examples:**

| Call | Result |
|------|--------|
| `to_float([42])` | `[42]` (stored as float64) |
| `to_float("2.5")` | `[2.5]` |

**Notes:**

- Panics on invalid input.
- Not dual-return.

---

## `fmt_float(val, decimals)`

**Signature:** `fmt_float(val, decimals) -> string`

**Description:** Formats a numeric value as a string with fixed decimal places.

**Returns:** Always succeeds on valid input. Returns a string.

**Examples:**

| Call | Result |
|------|--------|
| `fmt_float([3.14159], [2])` | `"3.14"` |
| `fmt_float([42], [3])` | `"42.000"` |
| `fmt_float([1.0], [0])` | `"1"` |

**Notes:**

- Panics if the first argument is not numeric or the second argument is not a non-negative integer.
- Not dual-return.
