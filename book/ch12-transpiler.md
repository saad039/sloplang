# Chapter 12: Inside the Transpiler

Sloplang is not an interpreter. When you run `slop hello.slop`, the CLI reads your source, runs it through a full compilation pipeline, and executes a native binary. This chapter walks through each stage of that pipeline — lexer, parser, codegen, and runtime — with real code from the implementation. Reading this chapter will explain why certain constructs panic, why `--x` cannot mean double-negate, and how a single `.slop` file becomes an executable in under a second.

---

## 12.1 The Pipeline: .slop → .go → binary

The transpiler is a five-stage pipeline:

```
.slop source
     |  Lexer (tokenize)
     v
token stream  ([]lexer.Token)
     |  Parser (recursive descent + Pratt precedence)
     v
AST  (*parser.Program)
     |  Codegen (go/ast builder)
     v
*ast.File  (Go AST)
     |  go/format
     v
formatted Go source  ([]byte)
     |  AssembleWithRuntime (inline runtime files)
     v
single-file package main  (main.go in temp dir)
     |  go build
     v
native binary
```

The entry point in `sloplang/main.go` sequences these stages explicitly:

```go
l := lexer.New(string(source))
tokens := l.Tokenize()

p := parser.New(tokens)
program, errs := p.Parse()

gen := codegen.New()
output, err := gen.Generate(program)

assembled, err := codegen.AssembleWithRuntime(output, rtSlopValue, rtOps, rtIO)
```

The assembled source is a single `package main` file that contains both the user's translated code and the full runtime (three Go files embedded at build time via `//go:embed`). It is written to a temp directory alongside a minimal `go.mod`, compiled with `go build`, and the resulting binary is run in place.

The three embedded runtime files are:

```go
//go:embed pkg/runtime/slop_value.go
var rtSlopValue string

//go:embed pkg/runtime/ops.go
var rtOps string

//go:embed pkg/runtime/io.go
var rtIO string
```

Because the runtime is embedded in the `slop` binary itself, running a `.slop` file has no external dependency beyond a Go toolchain.

### Directory structure

```
sloplang/                          # project root
├── docs/
└── sloplang/                      # Go module root (go.mod lives here)
    ├── main.go                    # CLI: lex → parse → generate → build → run
    ├── pkg/
    │   ├── lexer/
    │   │   ├── token.go           # TokenType constants, Token struct, keywords map
    │   │   └── lexer.go           # Lexer: input → []Token
    │   ├── parser/
    │   │   ├── ast.go             # AST node types (Stmt, Expr interfaces)
    │   │   └── parser.go          # Parser: []Token → *Program
    │   ├── codegen/
    │   │   └── codegen.go         # Generator: *Program → Go source (via go/ast)
    │   └── runtime/
    │       ├── slop_value.go      # SlopValue, IsTruthy, FormatValue, StdoutWrite
    │       ├── ops.go             # Arithmetic, comparison, array operations
    │       └── io.go              # StdinRead, FileRead, FileWrite, Split, ToNum
    └── examples/
```

---

## 12.2 The Lexer — Tokenizing Source

The lexer is a single-pass, character-by-character scanner. Its state is four fields:

```go
type Lexer struct {
    input   string
    pos     int  // current position (points to current char)
    readPos int  // next reading position
    ch      byte // current char under examination
    line    int
    col     int
}
```

`NextToken()` is one large `switch` on `l.ch`. For every character that can begin a multi-character operator, the lexer peeks one character ahead before committing to a token type.

### The Token struct

```go
type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Col     int
}
```

`TokenType` is an `int` with `iota` constants covering every token in the language: `TOKEN_EOF`, `TOKEN_ILLEGAL`, literals (`TOKEN_INT`, `TOKEN_UINT`, `TOKEN_FLOAT`, `TOKEN_STRING`, `TOKEN_IDENT`), all operators, delimiters, I/O symbols, and keywords.

### Greedy multi-character matching

Every operator that shares a prefix with a shorter operator is disambiguated by peeking. The full `<` branch, for example, handles five distinct tokens:

```go
case l.ch == '<':
    if l.peekChar() == '=' {
        tok.Type = TOKEN_LTE
        tok.Literal = "<="
        l.readChar()
        l.readChar()
    } else if l.peekChar() == '-' {
        tok.Type = TOKEN_RETURN
        tok.Literal = "<-"
        l.readChar()
        l.readChar()
    } else if l.peekChar() == '<' {
        tok.Type = TOKEN_LSHIFT
        tok.Literal = "<<"
        l.readChar()
        l.readChar()
    } else if l.peekChar() == '|' {
        tok.Type = TOKEN_STDIN_READ
        tok.Literal = "<|"
        l.readChar()
        l.readChar()
    } else if l.peekChar() == '.' {
        tok.Type = TOKEN_FILE_READ
        tok.Literal = "<."
        l.readChar()
        l.readChar()
    } else {
        tok.Type = TOKEN_LT
        tok.Literal = "<"
        l.readChar()
    }
```

The `.` branch shows an additional layer of lookahead needed to distinguish `.>` from `.>>`:

```go
case l.ch == '.':
    if l.peekChar() == '>' {
        l.readChar() // consume .
        l.readChar() // consume >
        if l.ch == '>' {
            tok.Type = TOKEN_FILE_APPEND
            tok.Literal = ".>>"
            l.readChar()
        } else {
            tok.Type = TOKEN_FILE_WRITE
            tok.Literal = ".>"
        }
    } else {
        tok.Type = TOKEN_ILLEGAL
        tok.Literal = string(l.ch)
        l.readChar()
    }
```

After consuming `.` and `>`, the lexer checks whether the next character is another `>`. If it is, the token is `.>>` (file append); otherwise it is `.>` (file write). This three-character lookahead is the most complex disambiguation in the lexer.

### Why `--` is TOKEN_REMOVE and `--x` cannot double-negate

The `-` branch peeks before emitting:

```go
case l.ch == '-':
    if l.peekChar() == '-' {
        tok.Type = TOKEN_REMOVE
        tok.Literal = "--"
        l.readChar()
        l.readChar()
    } else {
        tok.Type = TOKEN_MINUS
        tok.Literal = "-"
        l.readChar()
    }
```

When the lexer sees `--x`, it greedily consumes both `-` characters as a single `TOKEN_REMOVE` token before the parser ever sees them. The parser then sees `TOKEN_REMOVE` followed by `x`, not `-` followed by `-x`. `TOKEN_REMOVE` is a binary operator (remove first occurrence), not a prefix. The syntax `--x` is therefore a parse error. To negate twice, write `-(-x)`.

### Comments and strings

`//` comments are skipped by consuming characters until newline. String literals handle four escape sequences (`\n`, `\t`, `\\`, `\"`). `readString()` returns `(string, bool)` — `false` signals EOF before a closing `"`, allowing the caller to emit `TOKEN_ILLEGAL` with `"unterminated string"`:

```go
func (l *Lexer) readString() (string, bool) {
    l.readChar() // skip opening "
    var result []byte
    for l.ch != '"' && l.ch != 0 {
        if l.ch == '\\' {
            l.readChar()
            if l.ch == 0 {
                return string(result), false  // EOF inside escape
            }
            switch l.ch {
            case 'n':  result = append(result, '\n')
            case 't':  result = append(result, '\t')
            case '\\': result = append(result, '\\')
            case '"':  result = append(result, '"')
            default:   result = append(result, '\\', l.ch)
            }
        } else {
            result = append(result, l.ch)
        }
        l.readChar()
    }
    if l.ch == 0 {
        return string(result), false  // EOF before closing "
    }
    l.readChar() // skip closing "
    return string(result), true
}
```

### Scientific notation

`readNumber()` supports scientific notation (`1.79e308`, `5e-324`). After reading the integer or decimal part, it checks for `e`/`E` and peeks ahead to verify a valid exponent before consuming. If the character after `e` is not a digit or sign, the `e` is left alone as the start of the next token (an identifier). This ensures `e` suffix on a number that isn't part of a valid exponent doesn't break tokenization.

### Keywords

After reading an identifier, `LookupIdent` checks a map:

```go
var keywords = map[string]TokenType{
    "true":  TOKEN_TRUE,
    "false": TOKEN_FALSE,
    "fn":    TOKEN_FN,
    "if":    TOKEN_IF,
    "else":  TOKEN_ELSE,
    "for":   TOKEN_FOR,
    "in":    TOKEN_IN,
    "break": TOKEN_BREAK,
    "null":  TOKEN_NULL,
}
```

If the identifier is not in the map, `TOKEN_IDENT` is returned. Keywords cannot be used as variable names.

---

## 12.3 The Parser — Building the AST

The parser is a recursive-descent parser with Pratt-style precedence for expressions. Its state:

```go
type Parser struct {
    tokens     []lexer.Token
    pos        int
    errors     []string
    arrayDepth int // tracks nesting depth inside [] brackets
}
```

`save()` and `restore()` are the full extent of the backtracking mechanism — they are just position accessors:

```go
func (p *Parser) save() int       { return p.pos }
func (p *Parser) restore(pos int) { p.pos = pos }
```

### AST interfaces

Every node is either a `Stmt` or an `Expr`:

```go
type Stmt interface {
    Node
    stmtNode()
}

type Expr interface {
    Node
    exprNode()
}
```

`Node` requires only `TokenLiteral() string`, used for debugging and error messages. The `stmtNode()` and `exprNode()` methods are unexported marker methods that enforce the type distinction at compile time — you cannot accidentally pass a statement where an expression is expected.

### Statement dispatch

`parseStatement()` dispatches on the current token type. When the current token is `TOKEN_IDENT`, additional peeking is required to distinguish the many forms an identifier can start:

- `ident {` → hashmap declaration
- `ident <<` → push statement
- `ident $` → tentative dynamic-access-set (with save/restore)
- `ident @` → tentative index-set or key-set (with save/restore)
- `ident ,` → multi-assign
- `ident =` → plain assign
- anything else → expression statement

### The `arrayDepth` counter

Bare number literals and bare `null` are forbidden outside `[]` brackets. The parser enforces this with `arrayDepth`. The counter is incremented when entering `parseArrayLiteral()` and decremented on every exit path:

```go
func (p *Parser) parseArrayLiteral() *ArrayLiteral {
    p.advance() // consume '['
    p.arrayDepth++

    al := &ArrayLiteral{}

    if p.curToken().Type == lexer.TOKEN_RBRACKET {
        p.arrayDepth--
        p.advance() // consume ']'
        return al
    }

    elem := p.parseExpression()
    if elem == nil {
        p.arrayDepth--
        return nil
    }
    // ... collect elements ...
    p.arrayDepth--
    p.advance() // consume ']'
    return al
}
```

In `parsePrimary()`, before accepting a number or null, the parser checks the counter:

```go
case lexer.TOKEN_INT, lexer.TOKEN_UINT, lexer.TOKEN_FLOAT:
    if p.arrayDepth == 0 {
        p.addError("bare number literals are not allowed outside []; use [%s] instead at line %d", p.curToken().Literal, p.curToken().Line)
        p.advance()
        return nil
    }
    return p.parseNumberLiteral()
case lexer.TOKEN_NULL:
    if p.arrayDepth == 0 {
        p.addError("bare null is not allowed outside []; use [null] instead at line %d", p.curToken().Line)
        p.advance()
        return nil
    }
    p.advance()
    return &NullLiteral{}
```

`true` and `false` are exempt — they are standalone keywords that expand to `ArrayLiteral([1])` and `ArrayLiteral([])` without needing brackets.

### Lookahead for index-set and key-set

When the parser sees `ident @`, it does not know yet whether this is an expression like `arr@0` or a statement like `arr@0 = val`. It saves its position, tentatively parses ahead, and only commits if it finds `=`:

```go
if p.peekToken().Type == lexer.TOKEN_AT {
    saved := p.save()
    savedErrors := len(p.errors)
    name := p.curToken().Literal
    p.advance() // consume ident
    p.advance() // consume @

    if p.curToken().Type == lexer.TOKEN_IDENT && p.peekToken().Type != lexer.TOKEN_LPAREN {
        // Check for static key set: ident @ ident =
        keyName := p.curToken().Literal
        p.advance() // consume key ident
        if p.curToken().Type == lexer.TOKEN_ASSIGN {
            p.advance() // consume =
            val := p.parseExpression()
            if val != nil {
                return &KeySetStmt{
                    Object: &Identifier{Name: name},
                    Key:    keyName,
                    Value:  val,
                }
            }
        }
        // Not a key-set; restore
        p.restore(saved)
        p.errors = p.errors[:savedErrors]
    }
    // ... similar for numeric index-set ...
}
```

The same pattern applies for `$` (dynamic access set). Note that `$` is checked before `@` in `parseStatement()`, because `obj$key = val` would otherwise fall through to the `@` branch.

### Postfix operators

`parsePostfix()` handles `@`, `$`, and `::` in a loop after parsing a call or primary:

```go
func (p *Parser) parsePostfix() Expr {
    expr := p.parseCall()
    // ...
    for {
        if p.curToken().Type == lexer.TOKEN_DOLLAR {
            // obj$var → DynAccessExpr
        } else if p.curToken().Type == lexer.TOKEN_AT {
            if p.curToken().Type == lexer.TOKEN_IDENT ... {
                // obj@name → KeyAccessExpr (static string key)
            } else {
                // obj@expr → IndexExpr (numeric index)
            }
        } else if p.curToken().Type == lexer.TOKEN_DCOLON {
            // obj::low::high → SliceExpr
        } else {
            break
        }
    }
    return expr
}
```

### Why `#arr` cannot appear in slice bounds

The `#` token is handled in `parseUnary()`, not in `parsePostfix()`. When the parser encounters `arr::mid::#arr`, after consuming `arr::mid::`, it calls `parsePostfixPrimary()` to read the high bound. `parsePostfixPrimary()` handles `TOKEN_INT`, `TOKEN_IDENT`, `TOKEN_LBRACKET`, `TOKEN_STRING`, and `TOKEN_LPAREN` — it does not handle `TOKEN_HASH`. The `#` character therefore starts a new unary expression at the statement level, not inside the slice. The workaround is to hoist the length: `len = #arr` then `arr::mid::len`.

### Expression precedence (low to high)

1. `||`
2. `&&`
3. `==`, `!=`, `<`, `>`, `<=`, `>=`
4. `+`, `-`, `++`, `--`, `??`, `~@`
5. `*`, `/`, `%`
6. `**` (right-associative)
7. Unary prefix: `-`, `!`, `#`, `~`, `>>`, `##`, `@@`
8. Call: `name(args...)`
9. Postfix: `$`, `@`, `::`
10. Primary: literals, identifiers, `(expr)`

---

## 12.4 Codegen — Lowering to Go

The code generator takes a `*parser.Program` and returns formatted Go source bytes. It uses the `go/ast` package to construct the Go AST programmatically rather than generating source strings directly, then calls `go/format` to pretty-print it.

```go
type Generator struct {
    declared map[string]bool // tracks variables that have been declared
    globals  map[string]bool // top-level variable names hoisted to package level
}
```

### The Generate function: two passes

`Generate()` makes two passes over the top-level statements:

**Pass 1** collects global variable names:

```go
g.globals = make(map[string]bool)
for _, s := range program.Statements {
    switch stmt := s.(type) {
    case *parser.AssignStmt:
        g.globals[stmt.Name] = true
    case *parser.HashDeclStmt:
        g.globals[stmt.Name] = true
    case *parser.MultiAssignStmt:
        for _, name := range stmt.Names {
            g.globals[name] = true
        }
    }
}
```

Then emits `var name *SlopValue` for each global and marks them as already declared, so `main()` will use `=` instead of `:=`:

```go
for name := range g.globals {
    varDecls = append(varDecls, &ast.GenDecl{
        Tok: token.VAR,
        Specs: []ast.Spec{
            &ast.ValueSpec{
                Names: []*ast.Ident{ast.NewIdent(name)},
                Type:  slopValuePtrType(),
            },
        },
    })
    g.declared[name] = true
}
```

### Go keyword sanitization

User-defined identifiers that collide with Go's 25 keywords (`func`, `var`, `return`, `range`, `map`, etc.) would produce invalid Go ASTs that crash `format.Source()`. The codegen sanitizes these with `sanitizeIdent()`, which prefixes conflicting names with `slop_`:

```go
var goKeywords = map[string]bool{
    "break": true, "case": true, "chan": true, "const": true,
    "continue": true, "default": true, "defer": true, "else": true,
    // ... all 25 Go keywords ...
}

func sanitizeIdent(name string) string {
    if goKeywords[name] {
        return "slop_" + name
    }
    return name
}
```

This is applied at every `ast.NewIdent` call site that emits a user-defined name (variable declarations, function names, parameters, for-loop variables, identifier references). Hardcoded Go names like runtime functions, `"_"`, `"main"`, and `"nil"` are not sanitized. This means a sloplang variable named `func` transparently becomes `slop_func` in the generated Go — the user never sees the conflict.

**Pass 2** lowers statements: function declarations become top-level `*ast.FuncDecl` nodes, everything else goes into `main()`.

The final `*ast.File` structure mirrors:

```go
package main

// (runtime functions are inlined by AssembleWithRuntime — no import needed)

var x *SlopValue   // hoisted top-level variables
var y *SlopValue

func userFn(a *SlopValue, b *SlopValue) *SlopValue {
    // ...
    return NewSlopValue(...)
}

func main() {
    x = NewSlopValue(int64(42))   // = not :=, because x is a package-level var
    y = NewSlopValue(int64(0))
}
```

### Variable tracking: `:=` vs `=`

`lowerStmt` for `AssignStmt` checks `declared` before choosing the assignment token:

```go
case *parser.AssignStmt:
    tok := token.DEFINE
    if g.declared[s.Name] {
        tok = token.ASSIGN
    }
    g.declared[s.Name] = true
    assign := &ast.AssignStmt{
        Lhs: []ast.Expr{ast.NewIdent(s.Name)},
        Tok: tok,
        Rhs: []ast.Expr{g.lowerExpr(s.Value)},
    }
    if tok == token.DEFINE {
        suppress := &ast.AssignStmt{
            Lhs: []ast.Expr{ast.NewIdent("_")},
            Tok: token.ASSIGN,
            Rhs: []ast.Expr{ast.NewIdent(s.Name)},
        }
        return []ast.Stmt{assign, suppress}
    }
    return []ast.Stmt{assign}
```

Every first-use `:=` is immediately followed by `_ = varName` to suppress the Go "declared and not used" error.

### Function scope isolation with global seeding

`lowerFnDecl` saves the outer `declared` map, creates a fresh one seeded with globals and parameters, lowers the body, then restores:

```go
outerDeclared := g.declared
g.declared = make(map[string]bool)
for name := range g.globals {
    g.declared[name] = true  // globals use = (write to package-level var)
}
for _, p := range fd.Params {
    g.declared[p] = true     // params are already declared
}

var bodyStmts []ast.Stmt
for _, s := range fd.Body {
    bodyStmts = append(bodyStmts, g.lowerStmt(s)...)
}

g.declared = outerDeclared
```

This gives Python-like scoping: functions can read and write top-level variables, and new names inside a function are local.

### Binary operator dispatch

`lowerExpr` maps sloplang operators to runtime function names via a map literal:

```go
opFunc := map[string]string{
    "+": "Add", "-": "Sub", "*": "Mul", "/": "Div", "%": "Mod", "**": "Pow",
    "==": "Eq", "!=": "Neq", "<": "Lt", ">": "Gt", "<=": "Lte", ">=": "Gte",
    "&&": "And", "||": "Or",
    "++": "Concat", "--": "Remove", "??": "Contains", "~@": "RemoveAt",
}
fname, ok := opFunc[e.Op]
if !ok {
    return ast.NewIdent("nil")
}
return callRuntime(fname, g.lowerExpr(e.Left), g.lowerExpr(e.Right))
```

### The `callRuntime` helper

```go
func callRuntime(funcName string, args ...ast.Expr) *ast.CallExpr {
    return &ast.CallExpr{
        Fun:  ast.NewIdent(funcName),
        Args: args,
    }
}
```

Because `AssembleWithRuntime` inlines the runtime source directly into `package main`, there is no import prefix. Runtime functions are called by bare name (`NewSlopValue`, `Add`, `Push`, etc.).

### Dual-return detection

Some expressions return two `*SlopValue` results directly from Go: `StdinRead`, `FileRead`, and `to_num`. When these appear in a multi-assign, the codegen must not wrap them in `UnpackTwo`:

```go
func (g *Generator) isDualReturn(expr parser.Expr) bool {
    switch e := expr.(type) {
    case *parser.StdinReadExpr, *parser.FileReadExpr:
        return true
    case *parser.CallExpr:
        return e.Name == "to_num"
    }
    return false
}
```

User-defined functions that return `[result, errcode]` go through the `UnpackTwo` path (they return a single `*SlopValue` with two elements).

### Negated number literals in arrays

`[-1]` in sloplang must generate `NewSlopValue(int64(-1))`, not `NewSlopValue(Negate(NewSlopValue(int64(1))))`. The latter would create a nested `*SlopValue`, making `str([-1])` output `[[-1]]` instead of `[-1]`. `lowerRawValue` handles this:

```go
case *parser.UnaryExpr:
    if e.Op == "-" {
        if nl, ok := e.Operand.(*parser.NumberLiteral); ok {
            return g.lowerNumberRaw(&parser.NumberLiteral{
                Value:   "-" + nl.Value,
                NumType: nl.NumType,
            })
        }
    }
    return g.lowerExpr(e)
```

---

## 12.5 The Runtime — SlopValue and Operations

Every value in a running sloplang program is a `*SlopValue`. The struct is defined in `pkg/runtime/slop_value.go`:

```go
type SlopValue struct {
    Elements []any    // int64, uint64, float64, string, *SlopValue, or SlopNull
    Keys     []string // parallel to Elements for hashmaps; nil for plain arrays
}
```

`Elements` holds the raw Go values. Plain arrays have `Keys == nil`. Hashmaps have `Keys` as a parallel slice of field names. The universal constructor:

```go
func NewSlopValue(elems ...any) *SlopValue {
    return &SlopValue{Elements: elems}
}
```

### FormatValue

`FormatValue` is the function behind `str()` and `|>`. It begins with a nil guard: if `v` is `nil` (which happens when a hoisted global variable is used before assignment), it panics with `"sloplang: variable used before assignment"` instead of crashing with a SIGSEGV. After that, it has one special case: a single-element string prints without brackets. Everything else — single integers, floats, multi-element arrays, empty arrays, hashmaps — is wrapped in `[...]`:

```go
func FormatValue(v *SlopValue) string {
    if v == nil {
        panic("sloplang: variable used before assignment")
    }
    // Single-element string: print raw (no brackets)
    if len(v.Elements) == 1 {
        if s, ok := v.Elements[0].(string); ok {
            return s
        }
    }
    parts := make([]string, len(v.Elements))
    for i, elem := range v.Elements {
        parts[i] = formatElement(elem)
    }
    return "[" + strings.Join(parts, ", ") + "]"
}
```

Consequences:
- `str(["hello"])` → `hello` (string, no brackets)
- `str([42])` → `[42]` (integer, bracketed)
- `str([null])` → `[null]` (null, bracketed)
- `str([])` → `[]` (empty, bracketed)
- `str([1, 2])` → `[1, 2]` (multi-element, bracketed)

### IsTruthy

`IsTruthy` implements the strict boolean rule. Only `[1]` is truthy; only `[]` is falsy. Everything else panics with a specific message:

```go
func (sv *SlopValue) IsTruthy() bool {
    if len(sv.Elements) == 0 {
        return false
    }
    if len(sv.Elements) != 1 {
        panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got %d-element array", len(sv.Elements)))
    }
    elem := sv.Elements[0]
    if _, ok := elem.(SlopNull); ok {
        panic("sloplang: cannot use null as boolean")
    }
    i, ok := elem.(int64)
    if !ok {
        panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got single-element %T", elem))
    }
    if i == 1 {
        return true
    }
    if i == 0 {
        panic("sloplang: [0] is not a valid boolean — use [] for false")
    }
    panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got [%d]", i))
}
```

### Push

`Push` implements the `<<` operator. It appends the *elements* of `val` to `sv` — not `val` itself. This is a spread, not a nested push:

```go
func Push(sv *SlopValue, val *SlopValue) *SlopValue {
    sv.Elements = append(sv.Elements, val.Elements...)
    return sv
}
```

After `arr << [1, 2]`, `arr` gains two new elements (`1` and `2`), not one new element (`[1, 2]`).

Mutating operators (`<<`, `>>`, `~@`, `<<<`) modify the `*SlopValue` in place through pointer sharing. When a nested array is extracted via `@` or `$`, both the original and the extracted variable point to the same `*SlopValue`. This means mutations through either reference are visible from both.

### NestPush

`NestPush` implements the `<<<` operator. Unlike `Push` which spreads elements, `NestPush` appends the entire `val` as a single nested element:

```go
func NestPush(sv *SlopValue, val *SlopValue) *SlopValue {
    sv.Elements = append(sv.Elements, val)
    return sv
}
```

`arr <<< [3, 4]` makes `[3, 4]` a single nested element in `arr`, producing `[..., [3, 4]]`.

This also explains the behavior of `++` (Concat), which does the same thing as `Push` but on a new allocation:

```go
func Concat(a, b *SlopValue) *SlopValue {
    elems := make([]any, 0, len(a.Elements)+len(b.Elements))
    elems = append(elems, a.Elements...)
    elems = append(elems, b.Elements...)
    return &SlopValue{Elements: elems}
}
```

`"prefix: " ++ str(x)` — where `str(x)` returns `["5"]` — produces a two-element array `["prefix: ", "5"]` which formats as `[prefix: , 5]`, not the string `"prefix: 5"`.

---

## 12.6 Why Certain Things Panic

This section catalogs the most surprising runtime behaviors and traces each to its exact implementation.

### 1. `[0]` panics as a boolean

`IsTruthy` explicitly tests for `i == 0` and panics:

```go
if i == 0 {
    panic("sloplang: [0] is not a valid boolean — use [] for false")
}
```

This is intentional. `[0]` is a valid integer value but not a valid boolean. Use `[]` (empty array) for false, `[1]` for true, or the keywords `false` and `true`.

### 2. Strings panic as boolean

`IsTruthy` does a type assertion to `int64`. A string fails this assertion:

```go
i, ok := elem.(int64)
if !ok {
    panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got single-element %T", elem))
}
```

`%T` for a string element prints `string`, so the panic message is: `sloplang: boolean expression must be [1] or [], got single-element string`.

### 3. `--x` is TOKEN_REMOVE, not double negate

As shown in Section 12.2, the lexer greedily tokenizes `--` before the parser sees anything. The two `-` characters are consumed as a single `TOKEN_REMOVE` token. There is no second chance for the parser to reinterpret them as two separate `TOKEN_MINUS` tokens. To negate twice: `-(-x)`.

### 4. `#arr` fails inside slice postfix

`parsePostfix()` calls `parsePostfixPrimary()` to read slice bounds. `parsePostfixPrimary()` handles only: `TOKEN_INT`, `TOKEN_UINT`, `TOKEN_FLOAT`, `TOKEN_IDENT`, `TOKEN_LBRACKET`, `TOKEN_STRING`, `TOKEN_LPAREN`. `TOKEN_HASH` is not in this list. When the parser reaches `#` as a slice bound, `parsePostfixPrimary()` falls to the `default` branch and emits an error.

### 5. `++` is element concatenation, not string join

`Concat` (the `++` operator) calls `append(elems, a.Elements...)` and `append(elems, b.Elements...)`. There is no code path that checks whether elements are strings and joins them. `["a"] ++ ["b"]` produces a two-element array `["a", "b"]`, which `FormatValue` renders as `[a, b]`. To build a formatted string, use separate `|>` calls or assemble with explicit string construction.

### 6. `.>` and `.>>` panic on write error

`FileWrite` and `FileAppend` in `io.go` call `panic(...)` directly instead of returning an error:

```go
func FileWrite(path, data *SlopValue) {
    pathStr := extractString(path)
    dataStr := FormatValue(data)
    if err := os.WriteFile(pathStr, []byte(dataStr), 0644); err != nil {
        panic(fmt.Sprintf("sloplang: file write error: %v", err))
    }
}

func FileAppend(path, data *SlopValue) {
    pathStr := extractString(path)
    dataStr := FormatValue(data)
    f, err := os.OpenFile(pathStr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        panic(fmt.Sprintf("sloplang: file append error: %v", err))
    }
    defer f.Close()
    if _, err := f.WriteString(dataStr); err != nil {
        panic(fmt.Sprintf("sloplang: file append error: %v", err))
    }
}
```

Unlike `<.` (file read) and `<|` (stdin read), which return a `(*SlopValue, *SlopValue)` error pair, write operations are panic-on-failure. There is no dual-return variant for writes.

### 7. `split(str, "")` returns the original string

`Split` checks for an empty separator and returns the input unchanged:

```go
func Split(sv, sep *SlopValue) *SlopValue {
    str := extractString(sv)
    sepStr := extractString(sep)
    if sepStr == "" {
        return NewSlopValue(str)
    }
    parts := strings.Split(str, sepStr)
    elems := make([]any, len(parts))
    for i, p := range parts {
        elems[i] = p
    }
    return &SlopValue{Elements: elems}
}
```

`strings.Split("abc", "")` in Go returns `["a", "b", "c"]` (character splitting). The sloplang runtime deliberately diverges from this behavior: an empty separator is treated as "no split", returning the original string as a single-element array. This avoids surprising character decomposition.

### 8. Functions cannot be stored in variables

User-defined functions are emitted as top-level Go `func` declarations by `lowerFnDecl`. They are not `*SlopValue` objects. There is no codegen path that wraps a function in a `SlopValue`. `SlopValue.Elements` holds `int64`, `uint64`, `float64`, `string`, `*SlopValue`, and `SlopNull` — there is no function pointer case. Attempting to assign a function to a variable (`f = myFunc`) would parse as an identifier reference, which would produce a Go identifier that cannot be assigned to a `*SlopValue`.

### 9. Modulo by zero panics

`Mod()` checks for zero divisors in both int64 and uint64 cases and panics with `"sloplang: modulo by zero"`. Without this check, Go's `%` operator would produce a raw `"integer divide by zero"` panic — the same underlying CPU fault as division by zero, but with a confusing message for sloplang users.

### 10. MinInt64 / -1 panics (integer overflow)

`Div()` checks for the edge case where `math.MinInt64` is divided by `-1`. The mathematical result (`math.MaxInt64 + 1`) cannot be represented in int64. Without the check, Go would produce a raw panic. `Div()` now panics with `"sloplang: integer overflow: MinInt64 / -1"`.

### 11. Negating MinInt64 panics (integer overflow)

`Negate()` checks for `math.MinInt64` in the int64 case. `-MinInt64` overflows (silently wraps back to `MinInt64` in Go). `Negate()` now panics with `"sloplang: cannot negate MinInt64: integer overflow"`.

### 12. Using a variable before assignment panics

Top-level variables are hoisted as `var x *SlopValue` (nil by default). If code references `x` before `x = [...]` assigns to it, `FormatValue` and `binaryOp` catch the nil pointer and panic with `"sloplang: variable used before assignment"` instead of crashing with a SIGSEGV.

### 13. Unclosed strings are rejected

If a string literal reaches EOF without a closing `"`, the lexer returns `TOKEN_ILLEGAL` with `"unterminated string"`. This produces a clean parse error instead of silently consuming the rest of the file as string content.

---

The transpiler is approximately 4,000 lines of Go across four packages. Each stage is narrow in responsibility: the lexer knows nothing about AST nodes, the parser knows nothing about Go code, the codegen knows nothing about file I/O. The runtime is a pure library with no awareness of the pipeline above it. This separation makes it straightforward to add new operators, builtins, and statement forms — the pattern established in Phases 1 through 9 is: add a token, add an AST node, add a codegen case, add a runtime function.
