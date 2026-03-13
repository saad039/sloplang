# Chapter 4: Control Flow

Control flow in Sloplang uses familiar keywords (`if`, `for`, `break`, `<-`) combined with the language's strict boolean semantics. Conditions must evaluate to exactly `[1]` (truthy) or `[]` (falsy); anything else causes a runtime panic. Loops work directly with array iteration or infinite loops with manual exit conditions. Return statements use the `<-` operator and can appear anywhere within a function.

---

## 4.1 if and else

Conditions in Sloplang must be either `[1]` (truthy) or `[]` (falsy). Comparison operators like `>`, `<`, `==`, and others return exactly these values, making them natural if-conditions. The `if` statement executes its body when the condition is `[1]`; the `else` branch executes when the condition is `[]`. There is no `else if` keyword—chain using `else { if ... }` instead.

```
x = [7]
if x > [5] {
    |> "big\n"
} else {
    if x > [2] {
        |> "medium\n"
    } else {
        |> "small\n"
    }
}
// output: big
```

Any value other than `[1]` or `[]` in a condition causes a runtime panic. This strictness prevents accidental misuse of truthy/falsy coercion.

---

## 4.2 for-in Loops

The `for item in arr { }` construct iterates over each element of an array. The loop variable `item` is a SlopValue wrapping each element. For numeric arrays, each element is wrapped in a single-element array—so iterating `[10, 20, 30]` gives loop variable values of `[10]`, `[20]`, and `[30]`. For string arrays, each element is the raw string itself.

```
for item in [10, 20, 30] {
    |> str(item)
    |> "\n"
}
// output:
// [10]
// [20]
// [30]
```

String iteration unwraps each element to the raw string, allowing direct output:

```
for s in ["a", "b", "c"] {
    |> s
    |> "\n"
}
// output:
// a
// b
// c
```

When iterating numbers, remember that `item` is wrapped; use `str(item)` to format it as a string.

---

## 4.3 Infinite Loops and break

An infinite loop is written as `for { }` with no condition. Use `break` inside the loop to exit it when a condition is met. The `break` statement exits the innermost enclosing loop and is typically paired with a condition checked during loop execution.

```
i = [0]
for {
    if i == [5] { break }
    |> str(i)
    |> "\n"
    i = i + [1]
}
// output:
// [0]
// [1]
// [2]
// [3]
// [4]
```

When you need explicit index access during iteration (instead of using `for-in`), use manual indexing with dynamic access (`$`) and an infinite loop:

```
arr = [10, 20, 30]
i = [0]
n = [3]
for {
    if i >= n { break }
    |> str(arr$i)
    |> "\n"
    i = i + [1]
}
// output:
// [10]
// [20]
// [30]
```

---

## 4.4 The Return Statement: `<-`

The `<-` operator returns from the current function with a value. A function may have multiple return statements in different branches—early return is an idiomatic pattern in Sloplang. Every function must have an explicit `<-` on every code path; a function that reaches the end without one will fail to compile. Using `<-` outside a function body is a compile error.

```
fn sign(x) {
    if x > [0] { <- [1] }
    if x < [0] { <- -[1] }
    <- [0]
}

|> str(sign([5]))     // [1]
|> "\n"
|> str(sign([-3]))    // [-1]
|> "\n"
|> str(sign([0]))     // [0]
|> "\n"
```

This example demonstrates early returns: if `x` is positive, the first condition triggers and returns `[1]`. If `x` is negative, the second condition triggers and returns `[-1]`. Otherwise, the final statement returns `[0]`. Multi-return and error handling patterns are covered in Chapter 5.

---

## 4.5 Nesting and Scope

All variable names declared within a function share a single flat scope. Variables assigned inside loops or if blocks are accessible throughout the entire function—there is no block-level scoping. A variable set inside a loop remains visible after the loop exits. Shadowing is not permitted; a name cannot be redeclared in an inner block.

```
fn first_even(arr) {
    result = []
    for item in arr {
        rem = item % [2]
        if rem == [0] {
            result = item
            break
        }
    }
    <- result    // visible here — same scope as loop body
}

|> str(first_even([1, 3, 4, 6]))    // [4]
|> "\n"
```

In this example, both `result` and `rem` are declared inside the loop but are part of the enclosing function scope. The variable `result` is accessed after the loop exits, which is valid because the entire function shares one flat scope. The variable `rem` is also visible after the loop, even though it is not used afterward.
