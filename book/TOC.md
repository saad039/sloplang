# Programming in Sloplang — Table of Contents

---

## [Preface](preface.md)

- Who This Book Is For
- What Makes Sloplang Unusual
- How This Book Is Organized
- Getting Started (installation)

---

## Part I: Tutorial Introduction

### [Chapter 1: A Tutorial Introduction](ch01-tutorial.md)

- 1.1 Hello, World
- 1.2 Variables and the Array Model
- 1.3 Arithmetic
- 1.4 Control Flow — if/else, for-in, infinite loops, break
- 1.5 Functions — declaration, return, recursion
- 1.6 Arrays — push, pop, slice, contains
- 1.7 Hashmaps — declaration, key access, iteration
- 1.8 Input and Output — stdout, stdin, file read/write
- 1.9 Error Handling — dual-return, `safe_div`
- 1.10 A Complete Program — Word Frequency Counter

---

## Part II: The Language

### [Chapter 2: Types and Values](ch02-types.md)

- 2.1 The SlopValue — Everything Is an Array
- 2.2 Integer Types: int64 and uint64
- 2.3 Floating-Point: float64
- 2.4 Strings
- 2.5 Null
- 2.6 Booleans: true, false, and Strict Truthiness
- 2.7 Nested Arrays
- 2.8 The Bracket Rule

### [Chapter 3: Operators](ch03-operators.md)

- 3.1 Arithmetic Operators (+, -, *, /, %, **)
- 3.2 Comparison Operators (==, !=, <, >, <=, >=)
- 3.3 Logical Operators (&&, ||, !)
- 3.4 Operator Precedence Table
- 3.5 Element-wise Semantics
- 3.6 Type Safety in Expressions

### [Chapter 4: Control Flow](ch04-control-flow.md)

- 4.1 if and else
- 4.2 for-in Loops
- 4.3 Infinite Loops and break
- 4.4 The Return Statement: `<-`
- 4.5 Nesting and Scope

### [Chapter 5: Functions](ch05-functions.md)

- 5.1 Declaring and Calling Functions
- 5.2 Parameters and Return Values
- 5.3 Recursion
- 5.4 Variable Scope and Globals
- 5.5 Functions Are Not First-Class Values
- 5.6 Multi-Return Pattern

### [Chapter 6: Arrays](ch06-arrays.md)

- 6.1 Creating Arrays
- 6.2 Indexing: `@` (literal) and `$` (dynamic)
- 6.3 Mutating Arrays: `<<` (push), `>>` (pop), `~@` (remove-at), index-set
- 6.4 Non-Mutating Operations: `::` (slice), `++` (concat), `--` (remove), `~` (unique)
- 6.5 Querying: `#` (length), `??` (contains)
- 6.6 Iterating Arrays
- 6.7 Nested Arrays and Matrices
- 6.8 Common Patterns — filter, map, reduce, build-then-slice

### [Chapter 7: Hashmaps](ch07-hashmaps.md)

- 7.1 Declaration Syntax
- 7.2 Reading and Writing Keys
- 7.3 Dynamic Access with `$`
- 7.4 Keys and Values: `##` and `@@`
- 7.5 Iterating a Hashmap
- 7.6 Hashmaps as Structs
- 7.7 Limitations and Workarounds

### [Chapter 8: Input and Output](ch08-io.md)

- 8.1 Writing to Stdout: `|>`
- 8.2 Reading from Stdin: `<|`
- 8.3 Reading Files: `<.`
- 8.4 Writing Files: `.>` and `.>>`
- 8.5 The `str()` Builtin
- 8.6 The `split()` Builtin
- 8.7 The `to_num()` Builtin
- 8.8 Formatting Values for Output
- 8.9 No Buffering

### [Chapter 9: Error Handling](ch09-errors.md)

- 9.1 The Dual-Return Convention
- 9.2 Error Codes
- 9.3 Propagating Errors Across Functions
- 9.4 Built-in Functions That Can Fail
- 9.5 Accumulating Errors
- 9.6 When to Use `exit(code)`
- 9.7 Limitations: No Exceptions, No Stack Traces

---

## Part III: Real Programs

### [Chapter 10: Sorting and Searching](ch10-sorting-searching.md)

- 10.1 Bubble Sort
- 10.2 Merge Sort
- 10.3 Quick Sort
- 10.4 Heap Sort
- 10.5 Linear Search
- 10.6 Binary Search

### [Chapter 11: Data Structures](ch11-data-structures.md)

- 11.1 Stacks and Queues with Arrays
- 11.2 Binary Search Tree with Parallel Arrays
- 11.3 Hashmaps as Sets
- 11.4 Graphs with Adjacency Lists (BFS, DFS)

---

## Part IV: Transpiler Internals

### [Chapter 12: Inside the Transpiler](ch12-transpiler.md)

- 12.1 The Pipeline: .slop → .go → binary
- 12.2 The Lexer — Tokenizing Source
- 12.3 The Parser — Building the AST
- 12.4 Codegen — Lowering to Go
- 12.5 The Runtime — SlopValue and Operations
- 12.6 Why Certain Things Panic

---

## Appendices

### [Appendix A: Operator Reference](appendix-a-operators.md)

Quick-reference table of every operator: name, syntax, example, result, and whether it mutates.

### [Appendix B: Builtin Functions](appendix-b-builtins.md)

- `str(val)` — convert any value to string
- `split(str, sep)` — split string into array
- `to_num(str)` — parse string to number (dual-return)
- `exit(code)` — terminate program

### [Appendix C: Limitations and Workarounds](appendix-c-limitations.md)

1. No network I/O
2. No exceptions
3. No async or concurrency
4. `++` is array concat, not string concatenation
5. No broadcasting — array lengths must match
6. `--x` is TOKEN_REMOVE, not double-negate
7. `[0]` panics in boolean context
8. Strings panic in boolean context
9. Floats panic in boolean context
10. Null panics in arithmetic, comparisons, iteration, and booleans
11. Bare numbers outside `[]` are parse errors
12. Bare `null` outside `[]` is a parse error
13. `[true]` creates `[[1]]`, not `[1]`
14. `str([42])` → `"[42]"` breaks numeric roundtrips
15. `#arr` cannot appear in slice postfix
16. `|>` has no trailing newline
17. `.>` and `.>>` panic on error (no dual-return)
18. `split(str, "")` returns original, not characters
19. `to_num` result type affects downstream arithmetic
20. No stack traces on panic
21. Functions are not first-class values
22. Hashmaps compare keys in insertion order

### [Appendix D: Formal Grammar (EBNF)](appendix-d-grammar.md)

Complete EBNF grammar: program, statements, expressions, literals, identifiers.

---

## Quick Reference: "I'm getting a panic / error about..."

| Problem | Where to look |
|---------|---------------|
| `[0]` or `[false]` in an `if` | [Ch 2.6](ch02-types.md) — strict booleans; [App C §7](appendix-c-limitations.md) |
| String in an `if` condition | [Ch 2.6](ch02-types.md) — only `[1]`/`[]` are valid; [App C §8](appendix-c-limitations.md) |
| `--x` did something unexpected | [Ch 3.1](ch03-operators.md) — `--` is remove, use `-(-x)`; [App C §6](appendix-c-limitations.md) |
| Bare `42` or `null` won't parse | [Ch 2.8](ch02-types.md) — bracket rule; [App C §11–12](appendix-c-limitations.md) |
| `[true]` nesting to `[[1]]` | [Ch 2.6](ch02-types.md) — `true` is `[1]`, wrapping again nests; [App C §13](appendix-c-limitations.md) |
| Mismatched array lengths in `+` | [Ch 3.5](ch03-operators.md) — element-wise requires same length; [App C §5](appendix-c-limitations.md) |
| Mixed int64/float64 panic | [Ch 2.3](ch02-types.md) — no implicit casting; [App C §19](appendix-c-limitations.md) |
| `++` gave an array, not a string | [Ch 6.4](ch06-arrays.md) — `++` is array concat; [Ch 8.1](ch08-io.md); [App C §4](appendix-c-limitations.md) |
| `str([42])` wrote `"[42]"` to file | [Ch 8.5](ch08-io.md) — FormatValue brackets non-strings; [App C §14](appendix-c-limitations.md) |
| `split(s, "")` didn't split chars | [Ch 8.6](ch08-io.md) — returns original; [App C §18](appendix-c-limitations.md) |
| `to_num` returned wrong type | [Ch 8.7](ch08-io.md) — int vs float depends on input; [App C §19](appendix-c-limitations.md) |
| `.>` panicked instead of returning error | [Ch 8.4](ch08-io.md) — file writes panic, not dual-return; [App C §17](appendix-c-limitations.md) |
| Can't store a function in a variable | [Ch 5.5](ch05-functions.md) — not first-class; [App C §21](appendix-c-limitations.md) |
| `#arr` fails inside `::` slice | [Ch 12.3](ch12-transpiler.md) — parser limitation; [App C §15](appendix-c-limitations.md) |
| `@key` vs `$var` confusion | [Ch 6.2](ch06-arrays.md) — `@` is literal, `$` is dynamic |
| Hashmap `==` says not equal | [Ch 7.7](ch07-hashmaps.md) — insertion order matters; [App C §22](appendix-c-limitations.md) |
| Error handling pattern unclear | [Ch 9.1](ch09-errors.md) — dual-return convention |
| Global variable not updating | [Ch 5.4](ch05-functions.md) — scope and globals |
| How to iterate a hashmap | [Ch 7.5](ch07-hashmaps.md) — `##` for keys, `for k in` |
| Modulo by zero panic | [Ch 9.7](ch09-errors.md) — runtime error; [Ch 12.6 §9](ch12-transpiler.md) |
| MinInt64 / -1 integer overflow | [Ch 9.7](ch09-errors.md) — runtime error; [Ch 12.6 §10](ch12-transpiler.md) |
| Variable used before assignment | [Ch 9.7](ch09-errors.md) — runtime error; [Ch 12.6 §12](ch12-transpiler.md) |
| Unterminated string error | [Ch 12.6 §13](ch12-transpiler.md) — lexer rejects unclosed strings |
| No output / missing newlines | [Ch 8.1](ch08-io.md) — `\|>` has no trailing newline; [App C §16](appendix-c-limitations.md) |