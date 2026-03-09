package lexer

// Lexer tokenizes sloplang source code.
type Lexer struct {
	input   string
	pos     int  // current position (points to current char)
	readPos int  // next reading position
	ch      byte // current char under examination
	line    int
	col     int
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPos]
	}
	l.pos = l.readPos
	l.readPos++
	l.col++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) {
		return 0
	}
	return l.input[l.readPos]
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := Token{Line: l.line, Col: l.col}

	switch {
	case l.ch == 0:
		tok.Type = TOKEN_EOF
		tok.Literal = ""

	case l.ch == '/' && l.peekChar() == '/':
		l.skipComment()
		return l.NextToken()

	case l.ch == '/':
		tok.Type = TOKEN_SLASH
		tok.Literal = "/"
		l.readChar()

	case l.ch == '[':
		tok.Type = TOKEN_LBRACKET
		tok.Literal = "["
		l.readChar()

	case l.ch == ']':
		tok.Type = TOKEN_RBRACKET
		tok.Literal = "]"
		l.readChar()

	case l.ch == ',':
		tok.Type = TOKEN_COMMA
		tok.Literal = ","
		l.readChar()

	case l.ch == '(':
		tok.Type = TOKEN_LPAREN
		tok.Literal = "("
		l.readChar()

	case l.ch == ')':
		tok.Type = TOKEN_RPAREN
		tok.Literal = ")"
		l.readChar()

	case l.ch == '+':
		if l.peekChar() == '+' {
			tok.Type = TOKEN_CONCAT
			tok.Literal = "++"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_PLUS
			tok.Literal = "+"
			l.readChar()
		}

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

	case l.ch == '*':
		if l.peekChar() == '*' {
			tok.Type = TOKEN_POWER
			tok.Literal = "**"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_STAR
			tok.Literal = "*"
			l.readChar()
		}

	case l.ch == '%':
		tok.Type = TOKEN_PERCENT
		tok.Literal = "%"
		l.readChar()

	case l.ch == '=':
		if l.peekChar() == '=' {
			tok.Type = TOKEN_EQ
			tok.Literal = "=="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_ASSIGN
			tok.Literal = "="
			l.readChar()
		}

	case l.ch == '!':
		if l.peekChar() == '=' {
			tok.Type = TOKEN_NEQ
			tok.Literal = "!="
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_NOT
			tok.Literal = "!"
			l.readChar()
		}

	case l.ch == '{':
		tok.Type = TOKEN_LBRACE
		tok.Literal = "{"
		l.readChar()

	case l.ch == '}':
		tok.Type = TOKEN_RBRACE
		tok.Literal = "}"
		l.readChar()

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

	case l.ch == '>':
		if l.peekChar() == '=' {
			tok.Type = TOKEN_GTE
			tok.Literal = ">="
			l.readChar()
			l.readChar()
		} else if l.peekChar() == '>' {
			tok.Type = TOKEN_RSHIFT
			tok.Literal = ">>"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_GT
			tok.Literal = ">"
			l.readChar()
		}

	case l.ch == '&' && l.peekChar() == '&':
		tok.Type = TOKEN_AND
		tok.Literal = "&&"
		l.readChar()
		l.readChar()

	case l.ch == '|':
		if l.peekChar() == '>' {
			tok.Type = TOKEN_PIPE_GT
			tok.Literal = "|>"
			l.readChar()
			l.readChar()
		} else if l.peekChar() == '|' {
			tok.Type = TOKEN_OR
			tok.Literal = "||"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_ILLEGAL
			tok.Literal = string(l.ch)
			l.readChar()
		}

	case l.ch == '@':
		if l.peekChar() == '@' {
			tok.Type = TOKEN_DOUBLE_AT
			tok.Literal = "@@"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_AT
			tok.Literal = "@"
			l.readChar()
		}

	case l.ch == '#':
		if l.peekChar() == '#' {
			tok.Type = TOKEN_DOUBLE_HASH
			tok.Literal = "##"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_HASH
			tok.Literal = "#"
			l.readChar()
		}

	case l.ch == '$':
		tok.Type = TOKEN_DOLLAR
		tok.Literal = "$"
		l.readChar()

	case l.ch == '~':
		if l.peekChar() == '@' {
			tok.Type = TOKEN_TILDE_AT
			tok.Literal = "~@"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_TILDE
			tok.Literal = "~"
			l.readChar()
		}

	case l.ch == '?':
		if l.peekChar() == '?' {
			tok.Type = TOKEN_CONTAINS
			tok.Literal = "??"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_ILLEGAL
			tok.Literal = string(l.ch)
			l.readChar()
		}

	case l.ch == ':':
		if l.peekChar() == ':' {
			tok.Type = TOKEN_DCOLON
			tok.Literal = "::"
			l.readChar()
			l.readChar()
		} else {
			tok.Type = TOKEN_ILLEGAL
			tok.Literal = string(l.ch)
			l.readChar()
		}

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

	case l.ch == '"':
		tok.Type = TOKEN_STRING
		tok.Literal = l.readString()

	case isDigit(l.ch):
		return l.readNumber()

	case isLetter(l.ch):
		literal := l.readIdentifier()
		tok.Type = LookupIdent(literal)
		tok.Literal = literal
		return tok

	default:
		tok.Type = TOKEN_ILLEGAL
		tok.Literal = string(l.ch)
		l.readChar()
	}

	return tok
}

// Tokenize returns all tokens from the input.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' || l.ch == '\n' {
		if l.ch == '\n' {
			l.line++
			l.col = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) readString() string {
	l.readChar() // skip opening "
	var result []byte
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			default:
				result = append(result, '\\', l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}
	l.readChar() // skip closing "
	return string(result)
}

func (l *Lexer) readNumber() Token {
	tok := Token{Line: l.line, Col: l.col}
	startPos := l.pos

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
		tok.Type = TOKEN_FLOAT
		tok.Literal = l.input[startPos:l.pos]
		return tok
	}

	if l.ch == 'u' {
		tok.Type = TOKEN_UINT
		tok.Literal = l.input[startPos:l.pos] + "u"
		l.readChar() // consume 'u'
		return tok
	}

	tok.Type = TOKEN_INT
	tok.Literal = l.input[startPos:l.pos]
	return tok
}

func (l *Lexer) readIdentifier() string {
	startPos := l.pos
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPos:l.pos]
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}
