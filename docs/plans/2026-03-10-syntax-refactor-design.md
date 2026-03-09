# Pre-Phase-8 Syntax Refactor Design

## Changes

### 1. Enforce bracket-only literals
All literal values except strings must be inside `[]`. Bare numbers, booleans, and null are parser errors.

| Before (invalid after) | After (required) |
|---|---|
| `x = 0` | `x = [0]` |
| `x = null` | `x = [null]` |
| `if true { }` | `if [true] { }` |
| `err != 0` | `err != [0]` |

Strings remain bare: `x = "hello"` is valid.

### 2. Unified `$` variable access, remove `@(expr)` and `@$var`

| Syntax | Meaning |
|---|---|
| `arr@0` | Literal numeric index (read) |
| `arr@name` | Literal string key (read) |
| `arr$var` | Variable access — runtime dispatches by type: int64 → numeric index, string → key lookup |
| `arr@name = val` | Literal string key set |
| `arr$var = val` | Variable set — runtime dispatches by type |

Removed: `@(expr)`, `@$var`, `@$var =`

### 3. Strict boolean expressions

Only `[1]` and `[]` are valid in boolean context. Everything else is a runtime panic.

| Expression | Result |
|---|---|
| `[1]` | truthy |
| `[]` | falsy |
| `[0]` | **panic** (use `[]` for false) |
| `[2]`, `[-1]` | **panic** |
| `[1, 2]` | **panic** |
| `["hello"]` | **panic** |
| `[[]]` | **panic** |
| `[null]` | **panic** |

Comparisons (`==`, `!=`, `<`, `>`, `<=`, `>=`) return `[1]` or `[]` — safe.
Logical operators (`&&`, `||`, `!`) require `[1]` or `[]` inputs and produce `[1]` or `[]` — safe.
Contains (`??`) returns `[1]` or `[]` — safe.

Consequence: cannot use raw values as conditions. `if #arr { }` is a panic — use `if #arr > [0] { }` instead.
