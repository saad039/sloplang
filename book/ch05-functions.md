# Chapter 5: Functions

Functions in Sloplang let you encapsulate logic into reusable blocks. They are always declared at the top level using the `fn` keyword, called by name, and can return values using `<-`. Functions cannot be stored in variables or passed as arguments — they are statements, not first-class values. This chapter covers declaration, parameters, return values, recursion, scope, and the multi-return pattern for error handling.

## 5.1 Declaring and Calling Functions

Function declaration uses the `fn` keyword and must occur at the top level of your program (outside any other function):

```
fn add(a, b) {
    <- a + b
}

|> str(add([3], [4]))    // [7]
|> "\n"
```

Here, `add` takes two parameters (`a` and `b`) and returns their sum. The `<-` operator immediately exits the function and returns the value. When you call `add([3], [4])`, the arguments are bound to the parameters, and the function body executes.

Function calls use positional arguments and are always called by name:

```
fn greet(name) {
    |> "Hello, "
    |> name
    |> "\n"
}

greet("Alice")    // prints: Hello, Alice
```

Parameters are local variables in the function scope — they exist only during the function call and shadow any global variables with the same name.

## 5.2 Parameters and Return Values

Parameters are assigned to local variables at call time. The `<-` operator is the only way to return a value from a function; it immediately exits and returns the given value. Every function must have an explicit `<-` on every code path; a function that reaches the end without one will fail to compile.

Multiple `<-` statements in different branches is idiomatic and provides early exit:

```
fn clamp(val, lo, hi) {
    if val < lo { <- lo }
    if val > hi { <- hi }
    <- val
}

|> str(clamp([5], [1], [10]))     // [5]
|> "\n"
|> str(clamp([-1], [1], [10]))    // [1]
|> "\n"
|> str(clamp([15], [1], [10]))    // [10]
|> "\n"
```

In this example, `clamp` returns early if the value falls outside the bounds, otherwise returns the value itself. Each branch has its own `<-`, making the intent clear.

## 5.3 Recursion

Functions can call themselves recursively. Sloplang has no tail-call optimization, so deep recursion will eventually overflow the stack. Use recursion for problems with naturally bounded depth, such as factorial or small Fibonacci numbers.

Here is factorial:

```
fn factorial(n) {
    if n <= [1] { <- [1] }
    <- n * factorial(n - [1])
}

|> str(factorial([5]))    // [120]
|> "\n"
|> str(factorial([10]))   // [3628800]
|> "\n"
```

And Fibonacci:

```
fn fib(n) {
    if n <= [1] { <- n }
    <- fib(n - [1]) + fib(n - [2])
}

|> str(fib([10]))    // [55]
|> "\n"
```

Note that `fib` has exponential time complexity — it is only practical for small values of `n` (typically `n <= 30`). For larger inputs, use iteration or memoization.

## 5.4 Variable Scope and Globals

Top-level assignments (outside any function) create package-level global variables. All functions can read and write these globals. Parameters shadow globals with the same name when inside a function.

Here is an example of reading and writing globals:

```
counter = [0]    // global

fn increment() {
    counter = counter + [1]    // writes global
}

increment()
increment()
|> str(counter)    // [2]
|> "\n"
```

Each call to `increment()` modifies the global `counter`. After two calls, `counter` is `[2]`.

Parameters shadow globals:

```
x = [100]    // global

fn show(x) {
    |> str(x)    // local x, not global
    |> "\n"
}
show([5])        // [5]
|> str(x)        // [100]  — global unchanged
|> "\n"
```

Inside `show`, the parameter `x` (bound to `[5]`) shadows the global `x` (which is `[100]`). The parameter is used, and the global is unaffected.

## 5.5 Functions Are Not First-Class Values

Functions in Sloplang are not first-class values: they cannot be stored in variables, passed as arguments, or returned from other functions. Functions must be declared at the top level and called by name only.

This does not work:

```
// f = add            — no function references
// apply(add, x, y)   — cannot pass functions
```

If you need parameterized behavior, define separate named functions and use a string-based dispatcher:

```
fn do_add(a, b) { <- a + b }
fn do_mul(a, b) { <- a * b }

fn apply_op(op, a, b) {
    if op == "add" { <- do_add(a, b) }
    if op == "mul" { <- do_mul(a, b) }
    <- []
}

|> str(apply_op("add", [3], [4]))    // [7]
|> "\n"
|> str(apply_op("mul", [3], [4]))    // [12]
|> "\n"
```

Here, `apply_op` manually dispatches to `do_add` or `do_mul` based on the string `op`. This pattern is suitable when the set of operations is small and known in advance.

## 5.6 Multi-Return Pattern

To return multiple values from a function, return a multi-element array and unpack it at the call site:

```
fn divmod(x, y) {
    <- [x / y, x % y]
}

q, r = divmod([7], [3])
|> str(q)    // [2]
|> "\n"
|> str(r)    // [1]
|> "\n"
```

The function `divmod` returns a two-element array `[quotient, remainder]`, which is unpacked into `q` and `r` at the call site.

This pattern extends to error handling. By convention, the last element is an error code: `[0]` means success, and nonzero means failure. This is the standard error-handling pattern in Sloplang:

```
fn safe_parse(s) {
    val, err = to_num(s)
    if err != [0] {
        <- [[], [1]]
    }
    <- [val, [0]]
}

result, err = safe_parse("42")
if err == [0] {
    |> str(result)    // [42]
    |> "\n"
}
```

Here, `safe_parse` calls the builtin `to_num`, which returns `(value, errcode)`. If the error code is nonzero, `safe_parse` returns an empty array and error code `[1]`. Otherwise, it returns the parsed value and success code `[0]`. The caller unpacks both values and checks the error code before using the result.

This pattern is covered in depth in Chapter 9 (Error Handling).
