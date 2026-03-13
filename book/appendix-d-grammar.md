# Appendix D: Formal Grammar

This appendix gives a complete EBNF grammar for sloplang. Each rule is written as:

```
rule = ... ;
```

`{ x }` means zero or more repetitions of `x`. `[ x ]` means `x` is optional. `( a | b )` means a choice between `a` and `b`. Terminal strings are in double quotes.

---

## Program

```ebnf
program         = { statement } ;
```

---

## Statements

```ebnf
statement       = assign-stmt
               | multi-assign-stmt
               | hashmap-decl-stmt
               | stdout-write-stmt
               | file-write-stmt
               | file-append-stmt
               | push-stmt
               | index-set-stmt
               | key-set-stmt
               | dyn-set-stmt
               | fn-decl-stmt
               | if-stmt
               | for-in-stmt
               | for-loop-stmt
               | break-stmt
               | return-stmt
               | expr-stmt
               ;

assign-stmt       = identifier "=" expr ;
multi-assign-stmt = identifier "," identifier "=" expr ;
hashmap-decl-stmt = identifier "{" [ identifier { "," identifier } ] "}" "=" expr ;
stdout-write-stmt = "|>" expr ;
file-write-stmt   = ".>" expr expr ;
file-append-stmt  = ".>>" expr expr ;
push-stmt         = identifier "<<" expr ;
index-set-stmt    = identifier "@" number-literal "=" expr ;
key-set-stmt      = identifier "@" identifier "=" expr ;
dyn-set-stmt      = identifier "$" identifier "=" expr ;
fn-decl-stmt      = "fn" identifier "(" [ identifier { "," identifier } ] ")" block ;
if-stmt           = "if" expr block [ "else" block ] ;
for-in-stmt       = "for" identifier "in" expr block ;
for-loop-stmt     = "for" block ;
break-stmt        = "break" ;
return-stmt       = "<-" expr ;
expr-stmt         = expr ;

block             = "{" { statement } "}" ;
```

---

## Expressions

Expressions are listed from lowest to highest precedence.

```ebnf
expr            = or-expr ;

or-expr         = and-expr { "||" and-expr } ;

and-expr        = cmp-expr { "&&" cmp-expr } ;

cmp-expr        = add-expr { ( "==" | "!=" | "<" | ">" | "<=" | ">=" ) add-expr } ;

add-expr        = mul-expr { ( "+" | "-" | "++" | "--" | "??" | "~@" ) mul-expr } ;

mul-expr        = pow-expr { ( "*" | "/" | "%" ) pow-expr } ;

pow-expr        = unary-expr [ "**" pow-expr ] ;    (* right-associative *)

unary-expr      = ( "-" | "!" | "#" | "~" | ">>" | "##" | "@@" ) unary-expr
               | postfix-expr
               ;

postfix-expr    = call-expr { postfix-op } ;

postfix-op      = "@" number-literal
               | "@" identifier
               | "$" identifier
               | "::" postfix-primary "::" postfix-primary
               ;

postfix-primary = number-literal | identifier | array-literal | string-literal | "(" expr ")" | identifier "(" [ expr { "," expr } ] ")" ;

call-expr       = primary-expr
               | identifier "(" [ expr { "," expr } ] ")"
               ;
```

---

## Primary Expressions

```ebnf
primary-expr    = identifier
               | array-literal
               | string-literal
               | "true"
               | "false"
               | "(" expr ")"
               | stdin-read
               | file-read
               ;

stdin-read      = "<|" ;
file-read       = "<." expr ;
```

---

## Literals

```ebnf
array-literal   = "[" [ array-element { "," array-element } ] "]" ;

array-element   = number-literal
               | "null"
               | string-literal
               | expr
               ;

number-literal  = integer-literal | uint-literal | float-literal ;

integer-literal = digit { digit } ;

uint-literal    = digit { digit } "u" ;

float-literal   = digit { digit } "." { digit } ;

string-literal  = '"' { string-char } '"' ;

string-char     = any-char-except-quote-and-backslash
               | "\\" ( "n" | "t" | "\\" | '"' )
               ;
```

---

## Identifiers and Digits

```ebnf
identifier      = letter { letter | digit | "_" } ;

letter          = "a" | "b" | ... | "z" | "A" | "B" | ... | "Z" ;

digit           = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9" ;
```

---

## Comments

```ebnf
line-comment    = "//" { any-char-except-newline } newline ;
```

Comments are discarded by the lexer and do not appear in the grammar above.

---

## Notes

**The bracket rule:** Number literals (`integer-literal`, `uint-literal`, `float-literal`) and `null` are only valid as `array-element` inside an `array-literal`. They cannot appear as standalone expressions. Writing `x = 42` or `x = null` is a parse error; the correct forms are `x = [42]` and `x = [null]`.

**Boolean keywords:** `true` and `false` are standalone keywords that expand to `[1]` and `[]` respectively. They do not require brackets and should not be placed inside `[]` (which would create nested values `[[1]]` and `[[]]`).

**Operator precedence summary (high to low):**

| Level | Operators |
|-------|-----------|
| Postfix | `@N`, `@key`, `$var`, `::` |
| Unary prefix | `-`, `!`, `#`, `~`, `>>`, `##`, `@@` |
| Power | `**` (right-associative) |
| Multiplicative | `*`, `/`, `%` |
| Additive / array | `+`, `-`, `++`, `--`, `??`, `~@` |
| Comparison | `==`, `!=`, `<`, `>`, `<=`, `>=` |
| Logical and | `&&` |
| Logical or | `\|\|` |

**Statement disambiguation:** The parser uses lookahead to distinguish statement forms that begin with an identifier. It checks for `$` (dynamic set), `@` (index/key set), and `<<` (push) before falling back to expression-statement or plain assignment.

**Dual-return assignments:** `multi-assign-stmt` handles the `val, err = expr` form used by `to_num`, `<.`, and `<|`. The right-hand side must produce a two-element value.
