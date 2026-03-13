# Preface

## Who This Book Is For

This book is for programmers with experience in at least one language who want to understand an unusual language design. You don't need to know Go—the transpiler happens to be written in Go, but you only need to use the `slop` CLI to run programs. This is not a beginner's guide; we assume you understand control flow, functions, variables, and data types. If you've written code in JavaScript, Python, Java, C, or Go, you have the background you need.

## What Makes Sloplang Unusual

Sloplang breaks from familiar language conventions in four ways:

**1. Everything is an array.** Numbers, strings, booleans—all are stored in arrays. A number is always written in brackets: `[42]`, `[3.14]`. A string literal uses double quotes—`"hello"`—and is stored as a single-element array. This is the universal data structure; there are no primitive types outside an array. Hashmaps and structs are also arrays under the hood.

**2. Booleans are strict.** Only two values are valid in a boolean context: `[1]` (true) and `[]` (false). If you try to use `[0]`, a string, a float, or any other value as a boolean, the program panics. No implicit truthy/falsy conversions.

**3. All operators are symbolic.** There are no methods and no dot notation. You don't call `arr.length()`; you use `# arr` to get the length. You don't call `map.get(key)`; you use `map @ key`. Operations are expressed through pure operator symbols, making code terse and consistent.

**4. Functions are not first-class values.** You can define functions and call them, but you cannot store a function in a variable or pass it as an argument. Functions are statements, not values.

## How This Book Is Organized

The book has four parts:

**Part I: Tutorial** walks through the core language features by example, using programs you can run.

**Part II: Language Reference** covers every operator, builtin, and language construct in detail. Chapters are organized by domain: arithmetic, arrays, hashmaps, I/O, control flow, and functions.

**Part III: Real Programs** shows complete working programs that solve practical problems, letting you see how the pieces fit together.

**Part IV: Transpiler Internals** explains how Sloplang works: the lexer, parser, code generator, and runtime. Read this if you want to modify the language or understand how it targets Go.

Appendices list all operators, builtins, known limitations, and the formal grammar.

## Getting Started

To run Sloplang programs, you need Go installed (version 1.16 or later). Download it from golang.org if you don't have it.

Clone or download the Sloplang repository. Inside the `sloplang/` directory, build the CLI:

```
go build -o slop .
```

This produces a `slop` binary. Run a program with:

```
./slop myprogram.slop
```

The CLI transpiles your code to Go, compiles it, and executes it—all in one step. The generated Go source is automatically written next to your `.slop` file with a `.go` extension (e.g., `myprogram.go`), so you can inspect the intermediate code if something goes wrong.

That's all you need to get started. Read Chapter 1 next.
