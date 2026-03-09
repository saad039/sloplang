package lexer

// TokenType represents the type of a lexer token.
type TokenType int

const (
	TOKEN_EOF TokenType = iota
	TOKEN_ILLEGAL

	// Literals
	TOKEN_INT    // 42
	TOKEN_UINT   // 42u
	TOKEN_FLOAT  // 3.14
	TOKEN_STRING // "hello"
	TOKEN_IDENT  // variable names

	// Operators
	TOKEN_ASSIGN   // =
	TOKEN_PIPE_GT  // |>
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_COMMA    // ,

	// Arithmetic
	TOKEN_PLUS    // +
	TOKEN_MINUS   // -
	TOKEN_STAR    // *
	TOKEN_SLASH   // /
	TOKEN_PERCENT // %
	TOKEN_POWER   // **

	// Comparison
	TOKEN_EQ  // ==
	TOKEN_NEQ // !=
	TOKEN_LT  // <
	TOKEN_GT  // >
	TOKEN_LTE // <=
	TOKEN_GTE // >=

	// Logical
	TOKEN_AND // &&
	TOKEN_OR  // ||
	TOKEN_NOT // !

	// Delimiters
	TOKEN_LPAREN // (
	TOKEN_RPAREN // )
	TOKEN_LBRACE // {
	TOKEN_RBRACE // }

	// Return
	TOKEN_RETURN // <-

	// Array operators
	TOKEN_AT       // @
	TOKEN_HASH     // #
	TOKEN_LSHIFT   // <<
	TOKEN_RSHIFT   // >>
	TOKEN_TILDE_AT // ~@
	TOKEN_DCOLON   // ::
	TOKEN_CONCAT   // ++
	TOKEN_REMOVE   // --
	TOKEN_TILDE    // ~
	TOKEN_CONTAINS // ??

	// Keywords
	TOKEN_TRUE  // true
	TOKEN_FALSE // false
	TOKEN_FN    // fn
	TOKEN_IF    // if
	TOKEN_ELSE  // else
	TOKEN_FOR   // for
	TOKEN_IN    // in
	TOKEN_BREAK // break
)

var tokenNames = map[TokenType]string{
	TOKEN_EOF:      "EOF",
	TOKEN_ILLEGAL:  "ILLEGAL",
	TOKEN_INT:      "INT",
	TOKEN_UINT:     "UINT",
	TOKEN_FLOAT:    "FLOAT",
	TOKEN_STRING:   "STRING",
	TOKEN_IDENT:    "IDENT",
	TOKEN_ASSIGN:   "ASSIGN",
	TOKEN_PIPE_GT:  "PIPE_GT",
	TOKEN_LBRACKET: "LBRACKET",
	TOKEN_RBRACKET: "RBRACKET",
	TOKEN_COMMA:    "COMMA",
	TOKEN_PLUS:     "PLUS",
	TOKEN_MINUS:    "MINUS",
	TOKEN_STAR:     "STAR",
	TOKEN_SLASH:    "SLASH",
	TOKEN_PERCENT:  "PERCENT",
	TOKEN_POWER:    "POWER",
	TOKEN_EQ:       "EQ",
	TOKEN_NEQ:      "NEQ",
	TOKEN_LT:       "LT",
	TOKEN_GT:       "GT",
	TOKEN_LTE:      "LTE",
	TOKEN_GTE:      "GTE",
	TOKEN_AND:      "AND",
	TOKEN_OR:       "OR",
	TOKEN_NOT:      "NOT",
	TOKEN_LPAREN:   "LPAREN",
	TOKEN_RPAREN:   "RPAREN",
	TOKEN_LBRACE:   "LBRACE",
	TOKEN_RBRACE:   "RBRACE",
	TOKEN_RETURN:   "RETURN",
	TOKEN_AT:       "AT",
	TOKEN_HASH:     "HASH",
	TOKEN_LSHIFT:   "LSHIFT",
	TOKEN_RSHIFT:   "RSHIFT",
	TOKEN_TILDE_AT: "TILDE_AT",
	TOKEN_DCOLON:   "DCOLON",
	TOKEN_CONCAT:   "CONCAT",
	TOKEN_REMOVE:   "REMOVE",
	TOKEN_TILDE:    "TILDE",
	TOKEN_CONTAINS: "CONTAINS",
	TOKEN_TRUE:     "TRUE",
	TOKEN_FALSE:    "FALSE",
	TOKEN_FN:       "FN",
	TOKEN_IF:       "IF",
	TOKEN_ELSE:     "ELSE",
	TOKEN_FOR:      "FOR",
	TOKEN_IN:       "IN",
	TOKEN_BREAK:    "BREAK",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Token represents a single lexer token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

var keywords = map[string]TokenType{
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
	"fn":    TOKEN_FN,
	"if":    TOKEN_IF,
	"else":  TOKEN_ELSE,
	"for":   TOKEN_FOR,
	"in":    TOKEN_IN,
	"break": TOKEN_BREAK,
}

// LookupIdent returns the token type for an identifier, checking keywords first.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}
