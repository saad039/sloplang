# Chapter 9: Error Handling

Sloplang uses a simple, predictable error model based on dual returns: functions that can fail return a value and an error code. There are no exceptions, no try-catch blocks, and no stack unwinding. Instead, you check error codes explicitly and propagate them upward when necessary.

## 9.1 The Dual-Return Convention

Functions that can fail return two values as a multi-element array: `[result, errcode]`. The caller unpacks this with the two-value assignment `a, b = fn()`.

The error code `[0]` always means success. Any nonzero error code indicates failure. The result value is typically the computed value on success, or an empty array `[]` on failure (by convention).

Why use this pattern instead of exceptions? Sloplang favors explicit error handling: the caller must decide what to do with an error at every point. This eliminates silent failures and makes the error flow visible in your code.

Here is a simple example: a safe division function that returns `[result, errcode]`:

```
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]    // failure: result=[], errcode=[1]
    }
    <- [a / b, [0]]    // success: result, errcode=[0]
}

result, err = safe_div([10], [2])
if err == [0] {
    |> str(result)    // prints: [5]
    |> "\n"
}

result, err = safe_div([10], [0])
if err != [0] {
    |> "division by zero\n"
}
```

Notice that the caller always unpacks the two values and then checks `err` to decide what to do. The pattern is: compute or receive the dual return, check if error is `[0]`, branch accordingly.

This is Go-style error handling, adapted for Sloplang's array-first model. Every function that can fail must declare its error behavior in documentation or code. Every caller must acknowledge the error and decide whether to handle it, propagate it, or abort.

## 9.2 Error Codes

Error codes are single-element arrays of integers:

- **`[0]`** always means success.
- **`[1]`** conventionally means "generic failure" — something went wrong, but no specific category.
- **`[2]`, `[3]`, etc.** are program-defined, specific failure modes.

Error codes are compared with `==` and `!=` — never used directly as a boolean condition, because `[0]` as a boolean would panic (attempting to coerce an array to true/false).

Here is an example with multiple error codes:

```
fn parse_positive(s) {
    val, err = to_num(s)
    if err != [0] {
        <- [[], [1]]    // parse failed: error code 1
    }
    if val <= [0] {
        <- [[], [2]]    // not positive: error code 2
    }
    <- [val, [0]]       // success
}

result, err = parse_positive("42")
if err == [0] {
    |> str(result)    // prints: [42]
    |> "\n"
}

result, err = parse_positive("-5")
if err == [2] {
    |> "not positive\n"
}

result, err = parse_positive("abc")
if err == [1] {
    |> "parse error\n"
}
```

Each specific error code has meaning in your program. Document them clearly so callers know what each code represents.

## 9.3 Propagating Errors Across Functions

When a function calls another function that returns an error, the pattern is: check if the error is nonzero, and if so, return it upward unchanged (or wrap it in a different code if your program distinguishes causes).

This allows errors to bubble up through multiple function layers without being silenced or transformed. Each intermediate function is responsible for deciding: stop and handle the error, transform it with additional context, or pass it through. By convention, most functions simply pass errors upward.

Here is a three-level chained propagation:

```
fn validate(x) {
    if x < [0] {
        <- [[], [2]]    // error code 2: negative
    }
    if x == [0] {
        <- [[], [3]]    // error code 3: zero
    }
    <- [x, [0]]         // success
}

fn double_safe(x) {
    v, e = validate(x)
    if e != [0] {
        <- [[], e]      // propagate error unchanged
    }
    <- [v * [2], [0]]
}

fn triple_safe(x) {
    v, e = double_safe(x)
    if e != [0] {
        <- [[], e]      // propagate error unchanged
    }
    <- [v + x, [0]]
}

tv, te = triple_safe([4])
if te == [0] {
    |> str(tv)    // prints: [12]  (4*2 + 4 = 12)
    |> "\n"
}

tv2, te2 = triple_safe([0])
if te2 != [0] {
    |> "error code: "
    |> str(te2)   // prints: error code: [3]
    |> "\n"
}
```

In this example:
- `validate` returns error `[2]` for negative input and `[3]` for zero input.
- `double_safe` calls `validate`, checks for error, and propagates it unchanged if nonzero.
- `triple_safe` calls `double_safe`, checks for error, and propagates it unchanged.
- The final caller receives the original error code from `validate`, even though it passed through two intermediate functions.

This is the essence of error propagation: each layer checks, and if an error occurred below, it stops processing and returns the error upward.

You can also transform errors at each layer if needed — for example, wrapping a low-level error in a higher-level code. The important pattern is: always check, and always decide explicitly whether to propagate, handle, or abort.

## 9.4 Built-in Functions That Can Fail

Several Sloplang built-in functions use the dual-return convention because they can fail. These are primarily I/O and parsing operations:

| Function | Success Return | Failure Return |
|----------|-----------------|-----------------|
| `<\|` (read line) | `(line_string, [0])` | `("", [1])` |
| `<. "path"` (read file) | `(data_string, [0])` | `("", [1])` |
| `to_num(s)` (parse number) | `([number], [0])` | `([], [1])` |

**Important:** `.>` (write file) and `.>>` (append file) do **not** use dual return — they panic if the operation fails. If writing may fail in your program, you will need to wrap them in defensive code or accept the panic.

The dual-return I/O functions allow you to handle read errors gracefully without crashing. Always check the error code before using the result:

```
line, err = <|
if err == [0] {
    |> "read: "
    |> line
    |> "\n"
} else {
    |> "EOF or read error\n"
}

data, ferr = <. "data.txt"
if ferr == [0] {
    |> data
    |> "\n"
} else {
    |> "could not read file\n"
}

num, nerr = to_num("42")
if nerr == [0] {
    |> "parsed: "
    |> str(num)
    |> "\n"
}
```

## 9.5 Accumulating Errors

When processing a collection, you often want to count both successes and failures, then report overall results. Use the dual-return pattern inside a loop to track both outcomes:

```
fn process(inputs) {
    sum = [0]
    fails = [0]
    for s in inputs {
        v, e = to_num(s)
        if e == [0] {
            sum = sum + v
        } else {
            fails = fails + [1]
        }
    }
    <- [sum, fails]
}

total, errors = process(["1", "2", "abc", "4"])
|> str(total)     // prints: [7]
|> "\n"
|> str(errors)    // prints: [1]
|> "\n"
```

In this example:
- `process` iterates over an array of strings.
- For each string, it calls `to_num`, which returns a value and an error code.
- If the error is `[0]` (success), it adds to the sum.
- If the error is nonzero (failure), it increments the failure counter.
- At the end, it returns both the sum and the count of failures.

This pattern is useful for batch processing, where some items may fail but you want to report overall progress and not abort on the first failure. You can also collect failing items separately if you need detailed diagnostics:

```
fn process_with_details(inputs) {
    successes = []
    failures = []
    for item in inputs {
        v, e = to_num(item)
        if e == [0] {
            successes = successes ++ [v]
        } else {
            failures = failures ++ [item]
        }
    }
    <- [[successes, failures], [0]]
}
```

In this variant, you keep the actual failing items, not just a count. The caller can then decide how to report them or retry them.

## 9.6 When to Use `exit(code)`

The `exit([code])` builtin terminates the program immediately with the given exit code. Use it when a fatal error occurs and further processing is impossible — for example, when a required configuration file cannot be read or a critical invariant has been violated.

`exit` takes a single-element integer array. By POSIX convention, `[0]` means success, and any nonzero value indicates failure. After calling `exit`, the program terminates immediately; no subsequent code executes.

```
data, err = <. "config.txt"
if err != [0] {
    |> "fatal: cannot read config.txt\n"
    exit([1])
}

// If we reach here, config.txt was read successfully
|> "config loaded\n"
|> data
|> "\n"
```

Use `exit` sparingly — only when there is no meaningful recovery. For expected failures, use the dual-return pattern instead. For example:
- If a user provides bad input, propagate the error and let the caller decide whether to retry or abort.
- If a file might not exist, use `<.` (which dual-returns) to check it gracefully.
- If the entire program depends on a precondition (like an environment variable), use `exit` only after exhausting recovery options.

`exit` is most appropriate for configuration file loads at startup, where misconfiguration should prevent the program from running at all.

## 9.7 Limitations: No Exceptions, No Stack Traces

Sloplang has no exception mechanism. There is no `try-catch`, no `recover()`, and no way to catch a runtime panic.

When a runtime error occurs — such as division by zero, an invalid boolean coercion, an index out of bounds, or a type mismatch — the program panics and aborts immediately. The Go runtime prints a panic message and stack trace to stderr, then exits with code 2.

Here are examples of runtime errors with no recovery:

```
// Index out of bounds
arr = [1, 2, 3]
|> str(arr@5)    // panics: sloplang: index out of bounds: 5 (length 3)

// Invalid boolean coercion
x = [5]
if x {           // panics: sloplang: boolean expression must be [1] or [], got [5]
    |> "yes\n"
}

// Division by zero
a = [10]
b = [0]
|> str(a / b)    // panics: sloplang: division by zero
```

When such a panic occurs, you will see output similar to:
```
panic: sloplang: index out of bounds: 5 (length 3)

goroutine 1 [running]:
...
```

The program terminates, and no recovery is possible from within Sloplang.

**Design Philosophy: Prevent, Don't Recover**

The Sloplang approach is: **design to prevent errors, not to catch them after.**

Guidelines:
- **Array bounds:** Always check array bounds before indexing. Use `if idx < #arr { ... }`.
- **Boolean checks:** Always check conditions explicitly: `if x == [0] { ... }` instead of `if x { ... }`.
- **Division:** Always check divisors: `if denom != [0] { ... }` before dividing.
- **Expected failures:** Use dual-return functions to handle them gracefully (file not found, parse error).
- **Unrecoverable errors:** Use `exit([code])` only for truly fatal situations (missing config, corrupted data).
- **Debug output:** Add `|>` statements before risky operations to understand state and catch bugs early.

Here is an example of defensive coding:

```
fn safe_access(arr, idx) {
    n = #arr
    if idx < [0] || idx >= n {
        <- [[], [1]]
    }
    <- [arr$idx, [0]]
}

result, err = safe_access([1, 2, 3], [1])
if err == [0] {
    |> "value: "
    |> str(result)
    |> "\n"
}

result2, err2 = safe_access([1, 2, 3], [10])
if err2 != [0] {
    |> "index out of bounds\n"
}
```

By being explicit and defensive in your code, you avoid surprises and panics in production. Use the dual-return pattern for expected failures, and use clear guards to prevent unexpected ones.
