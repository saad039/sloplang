# Appendix A: Operator Quick Reference

| Operator | Name | Syntax | Example | Result | Mutates? |
|----------|------|--------|---------|--------|----------|
| `+` | Add | `a + b` | `[1, 2] + [3, 4]` | `[4, 6]` | No |
| `-` | Subtract | `a - b` | `[5, 3] - [1, 1]` | `[4, 2]` | No |
| `-` | Negate (unary) | `-a` | `-[1, 2, 3]` | `[-1, -2, -3]` | No |
| `*` | Multiply | `a * b` | `[2, 3] * [4, 5]` | `[8, 15]` | No |
| `/` | Divide | `a / b` | `[10, 6] / [2, 3]` | `[5, 2]` | No |
| `%` | Modulo | `a % b` | `[7, 5] % [3, 2]` | `[1, 1]` | No |
| `**` | Power | `a ** b` | `[2, 3] ** [3, 2]` | `[8, 9]` | No |
| `==` | Equal | `a == b` | `[2] == [2]` | `[1]` | No |
| `!=` | Not equal | `a != b` | `[2] != [3]` | `[1]` | No |
| `<` | Less than | `a < b` | `[1] < [2]` | `[1]` | No |
| `>` | Greater than | `a > b` | `[2] > [1]` | `[1]` | No |
| `<=` | Less than or equal | `a <= b` | `[1] <= [1]` | `[1]` | No |
| `>=` | Greater than or equal | `a >= b` | `[2] >= [1]` | `[1]` | No |
| `&&` | Logical and | `a && b` | `true && true` | `[1]` | No |
| `\|\|` | Logical or | `a \|\| b` | `false \|\| true` | `[1]` | No |
| `!` | Logical not (unary) | `!a` | `!true` | `[]` | No |
| `#` | Length (unary prefix) | `#arr` | `#[10, 20, 30]` | `[3]` | No |
| `@` | Numeric index (read) | `arr@N` | `arr@0` | element at index 0 | No |
| `@` | String key (read) | `map@key` | `person@name` | value for key `"name"` | No |
| `@` | Numeric index (set) | `arr@N = val` | `arr@0 = [99]` | assigns index 0 | Yes |
| `@` | String key (set) | `map@key = val` | `person@name = "alice"` | assigns key `"name"` | Yes |
| `$` | Dynamic access (read) | `arr$var` | `arr$i` | index or key depending on type of `i` | No |
| `$` | Dynamic access (set) | `arr$var = val` | `arr$i = [99]` | assigns by index or key | Yes |
| `<<` | Push | `arr << val` | `arr << [5]` | appends `[5]` to `arr` | Yes |
| `<<<` | Nested push | `arr <<< val` | `arr <<< [5]` | appends `[5]` as nested element | Yes |
| `>>` | Pop (unary prefix) | `>>arr` | `x = >>arr` | removes and returns last element | Yes |
| `~@` | Remove at index | `arr ~@ idx` | `arr ~@ [2]` | removes and returns element at index 2 | Yes |
| `::` | Slice | `arr::start::end` | `arr::[1]::[4]` | elements at indices 1, 2, 3 | No |
| `++` | Concat | `a ++ b` | `[1, 2] ++ [3, 4]` | `[1, 2, 3, 4]` | No |
| `--` | Remove value | `arr -- val` | `arr -- [5]` | new array with first occurrence of `[5]` removed | No |
| `~` | Unique (unary prefix) | `~arr` | `~[1, 2, 1, 3]` | `[1, 2, 3]` (deep copy, first occurrences) | No |
| `??` | Contains | `arr ?? val` | `arr ?? [5]` | `[1]` if found, `[]` if not | No |
| `##` | Keys (unary prefix) | `##map` | `##person` | array of key strings | No |
| `@@` | Values (unary prefix) | `@@map` | `@@person` | array of values | No |
| `<-` | Return | `<- expr` | `<- [42]` | returns value from function | No |
| `=` | Assign | `ident = expr` | `x = [1, 2, 3]` | binds value to name | No |
| `\|>` | Write stdout | `\|> expr` | `\|> "hello\n"` | prints to stdout (no trailing newline) | No |
| `<\|` | Read stdin | `<\|` | `line, err = <\|` | reads one line from stdin; dual-return `(line, err)` | No |
| `<.` | Read file | `<. "path"` | `data, err = <. "file.txt"` | reads entire file; dual-return `(data, err)` | No |
| `.>` | Write file | `.> "path" expr` | `.> "out.txt" data` | overwrites file with value | No |
| `.>>` | Append file | `.>> "path" expr` | `.>> "out.txt" data` | appends value to file | No |

**Notes:**

- Arithmetic operators (`+`, `-`, `*`, `/`, `%`, `**`) operate element-wise. Both operands must be the same length and the same numeric type per element. No broadcasting.
- `==` and `!=` perform deep structural equality on arrays and hashmaps of any size. `<`, `>`, `<=`, `>=` require single-element arrays.
- Logical operators require strict boolean values: only `[1]` (truthy) and `[]` (falsy). `[0]`, strings, floats, and null all panic in boolean context.
- `<<` and `>>` mutate the array in place. `--` returns a new array without mutating.
- `~@` returns the removed element and mutates the array; `::`, `++`, `~` return new arrays.
- `|>` does not append a trailing newline. Use `|> "\n"` when a newline is needed.
- `.>` and `.>>` panic on write error (no dual-return). Ensure the path is writable before using them.
- `<<<` always nests: `arr <<< [3, 4]` appends `[3, 4]` as one nested element, producing `[..., [3, 4]]`. Contrast with `<<` which spreads elements individually.
