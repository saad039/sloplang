# Chapter 3: Operators

Sloplang operators are the symbolic workhorses of the language. They operate element-wise on arrays, enforce strict type and length matching, and follow a clear precedence hierarchy. This chapter covers arithmetic, comparison, logical operators, their precedence rules, and the semantics that make them predictable.

## 3.1 Arithmetic Operators

Arithmetic operators perform element-wise operations on arrays of the same length and type.

**`+` (Addition)** — adds corresponding elements:
```
|> str([1, 2] + [3, 4])    // [4, 6]
|> "\n"
```

**`-` (Subtraction)** — subtracts corresponding elements:
```
|> str([10, 20] - [3, 5])   // [7, 15]
|> "\n"
```

**`*` (Multiplication)** — multiplies corresponding elements:
```
|> str([2, 3] * [4, 5])     // [8, 15]
|> "\n"
```

**`/` (Division)** — integer division (truncated toward zero) for int64 arrays. Division by zero panics with `"sloplang: division by zero"`. The edge case `MinInt64 / -1` (which overflows int64) panics with `"sloplang: integer overflow"`:
```
|> str([10] / [3])    // [3]   — integer division, truncated
|> "\n"
// [10] / [0] panics: division by zero
// [-9223372036854775808] / [-1] panics: integer overflow
```

**`%` (Modulo)** — remainder of integer division. Modulo by zero panics with `"sloplang: modulo by zero"`:
```
|> str([10] % [3])    // [1]
|> "\n"
// [10] % [0] panics: modulo by zero
```

**`**` (Power)** — exponentiation, right-associative:
```
|> str([2] ** [3])         // [8]
|> "\n"
|> str([2] ** [3] ** [2])  // [512]   — 2^(3^2) = 2^9
|> "\n"
```

**Unary `-` (Negation)** — negates all elements. Negating `MinInt64` (`-9223372036854775808`) panics with `"sloplang: cannot negate MinInt64"` because the result overflows int64:
```
|> str(-[1, 2, 3])    // [-1, -2, -3]
|> "\n"
// -[-9223372036854775808] panics: integer overflow
```

> **Pitfall:** `--x` is TOKEN_REMOVE, not double negation.
> The `--` token removes the first occurrence of a value from an array.
> Use `-(-x)` for double negation:
>
> ```
> |> str(-(-[5]))    // [5]   — correct double negate
> |> "\n"
> // --[5] applies TOKEN_REMOVE to [5], which is not what you want
> ```

## 3.2 Comparison Operators

Comparison operators produce strict boolean results: `[1]` (true) or `[]` (false).

**`==` (Equality)** — deep structural equality, works on any size array or hashmap:
```
|> str([1, 2] == [1, 2])    // [1]
|> "\n"
|> str([1, 2] == [1, 3])    // []
|> "\n"
```

**`!=` (Inequality)** — negation of equality:
```
|> str([2] != [3])    // [1]
|> "\n"
|> str([2] != [2])    // []
|> "\n"
```

**`<`, `>`, `<=`, `>=` (Ordered Comparison)** — less than, greater than, and variants. These operators require single-element arrays and panic on multi-element inputs:
```
|> str([1] < [2])         // [1]
|> "\n"
|> str([5] > [3])         // [1]
|> "\n"
|> str([2] <= [2])        // [1]
|> "\n"
|> str([1] >= [2])        // []
|> "\n"
// [1, 2] < [3, 4] would panic — ordered ops require single-element
```

## 3.3 Logical Operators

Logical operators combine strict booleans (`[1]` or `[]`) and produce strict boolean results.

**`&&` (Logical AND)** — true only if both operands are truthy:
```
|> str(true && false)    // []
|> "\n"
|> str(true && true)     // [1]
|> "\n"
```

**`||` (Logical OR)** — true if either operand is truthy:
```
|> str(true || false)    // [1]
|> "\n"
|> str(false || false)   // []
|> "\n"
```

**`!` (Logical NOT)** — negates a boolean:
```
|> str(!true)            // []
|> "\n"
|> str(!false)           // [1]
|> "\n"
```

Note that `&&` and `||` do **not** short-circuit. Both operands are always evaluated before the logical operation is applied. If either operand has side effects (e.g., a function call that prints), both side effects will occur regardless of the first operand's truthiness.

## 3.4 Operator Precedence

Sloplang defines 10 precedence levels, from lowest to highest. Higher levels bind more tightly.

| Level | Operators | Notes |
|-------|-----------|-------|
| 1 (lowest) | `\|\|` | Logical OR |
| 2 | `&&` | Logical AND |
| 3 | `==`, `!=`, `<`, `>`, `<=`, `>=` | Comparison |
| 4 | `+`, `-`, `++`, `--`, `??`, `~@` | Additive, array ops |
| 5 | `*`, `/`, `%` | Multiplicative |
| 6 | `**` | Power (right-associative) |
| 7 | `-` (unary), `!`, `#`, `~`, `>>`, `##`, `@@` | Prefix unary |
| 8 | `name(...)` | Function call |
| 9 | `$`, `@`, `::` | Postfix |
| 10 (highest) | literals, identifiers, `(expr)` | Primary |

Examples demonstrating precedence:

```
|> str([2] * [3] + [1])       // [7]  — multiplication before addition
|> "\n"
|> str([2] ** [3] ** [2])     // [512] — right-associative: 2^(3^2) = 2^9
|> "\n"
|> str([1] + [2] == [3])      // [1]  — addition before comparison
|> "\n"
|> str([10] / [2] * [3])      // [15] — left-to-right: (10/2)*3
|> "\n"
```

Use parentheses to override precedence when intent is unclear.

## 3.5 Element-wise Semantics

All arithmetic operators apply element-wise. Both operands must be arrays of the same length; Sloplang does not broadcast or reshape arrays implicitly.

```
|> str([1, 2, 3] + [10, 20, 30])   // [11, 22, 33]
|> "\n"
|> str([5, 6] * [2, 3])            // [10, 18]
|> "\n"
|> str([7, 8, 9] - [1, 2, 3])      // [6, 6, 6]
|> "\n"
// [1, 2] + [1, 2, 3] would panic — length mismatch
```

Each position in the result array is computed independently from the corresponding positions in the input arrays.

## 3.6 Type Safety in Expressions

All elements in an array must have the same type. Mixing types (int64 with float64, uint64 with int64, numeric with string) at runtime causes a panic. Type checking happens at runtime, not compile time.

```
|> str([2] * [3] + [1])    // [7]  — all int64
|> "\n"
// These would panic at runtime:
// [1] + [1.0]    — int64 vs float64
// [1u] + [1]     — uint64 vs int64
// [1] + "hello"  — numeric vs string
```

To convert between types, use the `str()` builtin to convert to strings, or `to_num()` to parse strings as numbers. Explicit type handling prevents silent coercion bugs.

