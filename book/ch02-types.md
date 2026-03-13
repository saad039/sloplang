# Chapter 2: Types and Values

## 2.1 The SlopValue — Everything Is an Array

In Sloplang, every value is an array. This fundamental design choice unifies the type system. At the runtime level, values are stored in a `SlopValue` struct with two fields: an `Elements` slice (containing the actual data) and a `Keys` slice (for hashmap keys). A single integer like `[42]` is a one-element array. A string, a boolean, a null — all represented as arrays.

This universality means all operations work uniformly on arrays. There are no separate scalar types at the language level; the data type is determined by what the elements contain. An element can be an int64, uint64, float64, string, null, or another `SlopValue`. This array-first design eliminates special cases and makes Sloplang's operator system consistent.

## 2.2 Integer Types: int64 and uint64

Sloplang distinguishes between signed and unsigned 64-bit integers. A decimal literal like `[42]` creates an int64; a literal with a `u` suffix like `[42u]` creates a uint64. At runtime, arithmetic operations require both operands to have identical types and lengths. Mixing int64 and uint64 will panic.

```
x = [42]
y = [42u]
|> str(x)              // [42]
|> "\n"
|> str(y)              // [42]
|> "\n"
|> str([10u] + [5u])   // [15]
|> "\n"
// [10] + [5u] would panic — int64 vs uint64 type mismatch
```

## 2.3 Floating-Point: float64

Floating-point values use the float64 type. A decimal literal with a fractional part, like `[3.14]`, creates a float64 array. Scientific notation is supported: `[1.79e308]`, `[5e-324]`, `[2.5e10]` are all valid float64 literals. The `e`/`E` may be followed by an optional `+`/`-` sign and exponent digits.

Floating-point arithmetic follows standard rules; note that Go's `%g` formatting drops trailing zeros. For example, `[3.14] + [2.86]` evaluates to `[6]` rather than `[6.0]`.

Mixing float64 with int64 or uint64 panics, just as it does when mixing integer types. Sloplang has no casting mechanism — all operands in an arithmetic expression must be the same numeric type. If your program handles both, keep them in separate variables and operate within each type (see Appendix C).

```
fa = [3.14] + [2.86]
|> str(fa)             // [6]
|> "\n"
|> str([2.5] * [2.0])  // [5]
|> "\n"
|> str([5e-324])       // [5e-324]   — scientific notation
|> "\n"
// [1] + [1.0] would panic — int64 vs float64 type mismatch
```

## 2.4 Strings

String literals are written with double quotes and do not require brackets: `"hello"` is a valid string value. Internally, strings are stored as single-element `SlopValue` arrays. When you print a single-element string using the `str()` function, it renders without brackets — the raw string appears on its own.

By contrast, `str([42])` produces `"[42]"` — the number is enclosed in brackets, making it obvious that it is an array. This distinction between the raw representation of strings and the bracketed representation of arrays is crucial for understanding Sloplang's output.

Escape sequences are supported: `\n` for newline, `\t` for tab, `\\` for a literal backslash, and `\"` for a double quote.

```
greeting = "hello"
|> greeting            // hello
|> "\n"
|> str(greeting)       // hello   (single-element string prints raw)
|> "\n"
|> str([42])           // [42]    (arrays print with brackets)
|> "\n"
```

## 2.5 Null

The null value represents the absence of data. In Sloplang, null may only appear inside brackets: `[null]`. Writing bare `null` outside of brackets is a parse error. When you convert `[null]` to a string using `str()`, the result is `"[null]"` — with brackets — not the bare string `"null"`.

Two null values are equal: `[null] == [null]` evaluates to `[1]`. You can also determine the length of an array containing nulls using the `#` operator; `#[null, null]` returns `[2]`. However, null values cannot participate in arithmetic, ordered comparisons, iteration, or boolean contexts. If you attempt any of these operations, the program will panic. Always use `== [null]` to test for null before operating on a value.

```
x = [null]
|> str(x)                   // [null]
|> "\n"
|> str([null] == [null])    // [1]
|> "\n"
// These panic:
// x + [1]
// if [null] { }
// for v in [null] { }
```

## 2.6 Booleans: true, false, and Strict Truthiness

Sloplang has two boolean keywords: `true` and `false`. The keyword `true` produces the value `[1]` (a one-element array containing int64 1), and the keyword `false` produces the value `[]` (an empty array). They are not special types; `true == [1]` is true because they are identical values.

Truthiness in Sloplang is strict: only `[1]` is truthy, and only `[]` is falsy. This strictness prevents ambiguity. The value `[0]` is not falsy — it panics with the message "use [] for false". Multi-element arrays, strings, floats, and null all panic when used in a boolean context. The language rejects ambiguous truthiness and forces you to be explicit.

A common pitfall is `[true]`, which nests the value: `[true]` creates `[[1]]`, a one-element array containing the value `[1]`. This is not the same as the boolean `true`. Avoid wrapping booleans in extra brackets.

```
// true and [1] are the same thing
|> str(true)     // [1]
|> "\n"
|> str([1])      // [1]
|> "\n"
|> str(true == [1])    // [1]  (they are equal)
|> "\n"

// false and [] are the same thing
|> str(false)    // []
|> "\n"
|> str([])       // []
|> "\n"

// Pitfall: [true] nests — NOT the same as true
c = [true]
|> str(c)        // [[1]]  ← nested array, avoid this
|> "\n"

// These panic:
// if [0] { }      — "[0] is not a valid boolean — use [] for false"
// if "yes" { }    — "boolean expression must be [1] or []"
// if [1, 2] { }   — "got 2-element array"
```

## 2.7 Nested Arrays

Arrays can contain other arrays as elements. A nested array is created by listing other arrays inside the brackets: `[[1,2],[3,4]]` is a two-element array where each element is itself a two-element array (a "matrix" structure).

Access a nested element using the `@` operator chained together. `nested@0` retrieves the first sub-array, and `(nested@0)@1` retrieves the second element of that sub-array. Parentheses disambiguate the access chain.

```
matrix = [[1,2],[3,4]]
row = matrix@0
|> str(row)         // [1, 2]
|> "\n"
|> str(row@1)       // [2]
|> "\n"
```

## 2.8 The Bracket Rule

In Sloplang, bare numbers and bare `null` outside of brackets are parse errors. You cannot write `x = 42` or `y = null`. Every numeric or null value must be wrapped in brackets: `x = [42]` and `y = [null]` are the correct forms.

The two exceptions are the keywords `true` and `false`, which automatically produce bracketed values and do not require explicit wrapping. Thus `z = true` is valid and produces the value `[1]`.

```
// Valid:
x = [42]
y = [null]
z = true

// Parse errors:
// x = 42
// y = null
```

