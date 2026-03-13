# Chapter 6: Arrays

Arrays are the universal primitive in Sloplang. Every value — numbers, strings, booleans, hashmaps — lives inside an array. This chapter covers how to create arrays, access and mutate their elements, combine and query them, iterate over them, work with nested arrays, and compose common patterns like filter, map, and sum.

## 6.1 Creating Arrays

Array literals use square brackets with comma-separated elements:

```
nums   = [1, 2, 3]
empty  = []
mixed  = [1, 3.14, "hello"]
nested = [[1, 2], [3, 4]]
|> str(nums)      // [1, 2, 3]
|> "\n"
|> str(mixed)     // [1, 3.14, hello]
|> "\n"
|> str(nested)    // [[1, 2], [3, 4]]
|> "\n"
```

A few things to note:

- **Single-element arrays** like `[42]` are the standard way to hold a scalar value. Arithmetic, comparisons, and most operators expect single-element arrays when operating on scalars.
- **Mixed-type arrays** are allowed syntactically; type safety is enforced at runtime when you perform operations on them.
- **Nested arrays** are created with literal syntax — `[[1, 2], [3, 4]]` produces a two-element array whose elements are themselves arrays.
- **Empty array `[]`** is falsy. It doubles as the boolean `false` value in conditionals.

## 6.2 Indexing: `@` and `$`

Sloplang has three distinct syntaxes for element access, each with different semantics:

| Syntax   | Name                  | What it does                                              |
|----------|-----------------------|-----------------------------------------------------------|
| `arr@0`  | Literal numeric index | Returns element at position 0 (must be an integer literal) |
| `map@name` | Literal string key  | Looks up key `"name"` in a hashmap                        |
| `arr$var` | Dynamic dispatch     | If `var` is int64 → numeric index; if `var` is string → key lookup |

Here are all three forms in action:

```
arr = [10, 20, 30]
|> str(arr@0)     // [10] — literal numeric index

map{x, y} = [[1], [2]]
|> str(map@x)     // [1] — literal string key "x"
|> "\n"

i = [2]
|> str(arr$i)     // [30] — dynamic, int64 → index
|> "\n"

k = "x"
|> str(map$k)     // [1] — dynamic, string → key
|> "\n"
```

Use `@N` when the index is a compile-time constant, and `$var` when the index or key is computed at runtime.

## 6.3 Mutating Arrays: `<<`, `>>`, `~@`, index-set

### Push (`<<`)

`arr << rhs` appends all elements of `rhs` to `arr`, spreading them individually. It mutates `arr` in place.

```
arr = [1, 2]
arr << [3, 4]
|> str(arr)    // [1, 2, 3, 4]
|> "\n"
```

> **Pitfall:** `<<` spreads its right operand. `arr << [3, 4]` adds integers 3 and 4 separately — not `[3, 4]` as a single nested element. There is no operation to push a sub-array as a single nested element; nested arrays must be created with literal syntax.

### Pop (`>>`)

`>>arr` is a prefix operator that removes and returns the last element of `arr` as a single-element SlopValue. It mutates `arr` in place.

```
stack = [10, 20, 30]
last = >>stack
|> str(last)    // [30]
|> "\n"
|> str(stack)   // [10, 20]
|> "\n"
```

### Remove-at (`~@`)

`arr ~@ [n]` removes the element at index `n` and returns it. The right operand must be a bracketed single-element int64. It mutates `arr` in place.

```
items = [10, 20, 30, 40]
removed = items ~@ [1]
|> str(removed)   // [20]
|> "\n"
|> str(items)     // [10, 30, 40]
|> "\n"
```

### Index-set

Assign a new value to a specific position using `arr@N` (literal index) or `arr$var` (dynamic index) on the left-hand side:

```
arr = [10, 20, 30]
arr@1 = [99]
|> str(arr)    // [10, 99, 30]
|> "\n"
```

Dynamic index-set works the same way:

```
arr = [10, 20, 30]
i = [2]
arr$i = [77]
|> str(arr)    // [10, 20, 77]
|> "\n"
```

## 6.4 Non-Mutating Operations: `::`, `++`, `--`, `~`

### Slice (`::`)

`arr::lo::hi` returns a new array containing elements at positions `[lo, hi)` (half-open range). The source array is not modified.

```
arr = [10, 20, 30, 40, 50]
|> str(arr::1::4)    // [20, 30, 40]
|> "\n"
|> str(arr)          // [10, 20, 30, 40, 50]  — unchanged
|> "\n"
```

> **Pitfall:** `#arr` cannot appear directly in slice postfix — it starts a new unary expression rather than acting as a slice bound. Store the length in a variable first:
>
> ```
> len = #arr
> tail = arr::1::len    // OK
> // arr::1::#arr  — parse error
> ```

### Concat (`++`)

`arr1 ++ arr2` returns a new array that combines all elements of both arrays. Neither source is modified.

```
a = [1, 2, 3]
b = [4, 5]
|> str(a ++ b)    // [1, 2, 3, 4, 5]
|> "\n"
```

> **Pitfall:** `++` is array concatenation, not string concatenation.
> `"hello" ++ " world"` produces a 2-element array `["hello", " world"]`, not the string `"hello world"`.
> To print adjacent strings, use multiple `|>` calls:
>
> ```
> |> "hello"
> |> " world"
> |> "\n"
> ```

### Remove value (`--`)

`arr -- [val]` removes the first occurrence of `val` from the array and returns a new array. The source is not modified.

```
|> str([1, 2, 3, 2] -- [2])    // [1, 3, 2]
|> "\n"
```

Note: `--` is the TOKEN_REMOVE operator. It is not double negation — see §3.1 for the `-(-x)` pattern.

### Unique (`~`)

Prefix `~arr` returns a new array with duplicates removed, keeping the first occurrence of each value.

```
|> str(~[1, 2, 2, 3, 1])    // [1, 2, 3]
|> "\n"
```

## 6.5 Querying: `#`, `??`

### Length (`#`)

Prefix `#arr` returns the number of elements as a single-element array `[n]`.

```
|> str(#[10, 20, 30])    // [3]
|> "\n"
```

Because `#arr` returns `[n]`, you can use it directly in arithmetic or comparisons:

```
arr = [10, 20, 30]
len = #arr
|> str(len == [3])    // [1]
|> "\n"
```

### Contains (`??`)

`arr ?? [val]` returns `[1]` if `val` is present in `arr`, or `[]` if it is not. The right operand must be a single-element array.

```
|> str([10, 20, 30] ?? [20])    // [1]
|> "\n"
|> str([10, 20, 30] ?? [99])    // []
|> "\n"
```

> **Pitfall:** `arr ?? [1, 2]` panics — the right operand of `??` must be a single-element array. To check for multiple values, use separate `??` expressions or a loop.

## 6.6 Iterating Arrays

The `for item in arr { }` construct iterates over every element of `arr`. Each iteration, `item` is bound to that element wrapped in a SlopValue, so `str(item)` for a number array produces `"[10]"`, `"[20]"`, etc.

A simple sum:

```
sum = [0]
for item in [1, 2, 3, 4, 5] {
    sum = sum + item
}
|> str(sum)    // [15]
|> "\n"
```

Finding the maximum element:

```
nums = [3, 7, 2, 9, 1]
max = nums@0
for item in nums {
    if item > max {
        max = item
    }
}
|> str(max)    // [9]
|> "\n"
```

Collecting matching elements into a new array:

```
nums = [1, 2, 3, 4, 5, 6]
evens = []
for item in nums {
    if item % [2] == [0] {
        evens << item
    }
}
|> str(evens)    // [2, 4, 6]
|> "\n"
```

## 6.7 Nested Arrays and Matrices

Access a nested array with `@` twice: first to retrieve the row, then to retrieve the element within that row.

```
matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
row = matrix@1
|> str(row)       // [4, 5, 6]
|> "\n"
|> str(row@2)     // [6]
|> "\n"
```

You can chain directly for concise access:

```
matrix = [[1, 2, 3], [4, 5, 6], [7, 8, 9]]
|> str(matrix@2@0)    // [7]
|> "\n"
```

Iterating a matrix row by row:

```
matrix = [[1, 2], [3, 4], [5, 6]]
for row in matrix {
    |> str(row)
    |> "\n"
}
// [1, 2]
// [3, 4]
// [5, 6]
```

Nested arrays must be created with literal syntax. Because `<<` spreads its operand, pushing `[1, 2, 3]` with `<<` would append 1, 2, and 3 as separate elements rather than as a single nested row. Use `[[1, 2], [3, 4]]` notation for matrix literals.

## 6.8 Common Patterns and Idioms

### Filter

Build a new array containing only the elements that satisfy a condition:

```
fn filter_gt(arr, threshold) {
    result = []
    for item in arr {
        if item > threshold {
            result << item
        }
    }
    <- result
}

|> str(filter_gt([1, 3, 5, 2, 4], [2]))    // [3, 5, 4]
|> "\n"
```

### Map / Transform

Build a new array by applying a transformation to each element:

```
fn double_each(arr) {
    result = []
    for item in arr {
        result << item * [2]
    }
    <- result
}

|> str(double_each([1, 2, 3]))    // [2, 4, 6]
|> "\n"
```

### Sum / Reduce

Accumulate a result across all elements:

```
fn sum(arr) {
    acc = [0]
    for item in arr {
        acc = acc + item
    }
    <- acc
}

|> str(sum([1, 2, 3, 4, 5]))    // [15]
|> "\n"
```

### Build then Slice

Collect elements first, then take a prefix or suffix with `::`:

```
arr = [10, 20, 30, 40, 50]
len = #arr
first_three = arr::0::[3]
last_two = arr::3::len
|> str(first_three)    // [10, 20, 30]
|> "\n"
|> str(last_two)       // [40, 50]
|> "\n"
```

These four patterns — filter, map, reduce, and build-then-slice — cover the vast majority of array processing tasks in Sloplang. Combine them as needed; for example, filter first and then sum the matching elements.
