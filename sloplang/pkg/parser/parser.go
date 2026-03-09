package parser

import (
	"fmt"

	"github.com/saad039/sloplang/pkg/lexer"
)

// Parser produces an AST from a token slice.
type Parser struct {
	tokens []lexer.Token
	pos    int
	errors []string
}

// New creates a new Parser for the given tokens.
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse parses the token stream into a Program AST.
func (p *Parser) Parse() (*Program, []string) {
	program := &Program{}
	for p.curToken().Type != lexer.TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}
	return program, p.errors
}

func (p *Parser) parseStatement() Stmt {
	switch p.curToken().Type {
	case lexer.TOKEN_FN:
		return p.parseFnDecl()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_FOR:
		return p.parseForStmt()
	case lexer.TOKEN_BREAK:
		return p.parseBreakStmt()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_PIPE_GT:
		return p.parseStdoutWriteStatement()
	case lexer.TOKEN_FILE_WRITE:
		return p.parseFileWriteStmt()
	case lexer.TOKEN_FILE_APPEND:
		return p.parseFileAppendStmt()
	case lexer.TOKEN_IDENT:
		// Disambiguate: hashmap decl (a{...} = ...), push (a << ...), index-set/key-set (a@... = ...), multi-assign, single assign, or bare expression
		if p.peekToken().Type == lexer.TOKEN_LBRACE {
			return p.parseHashDeclStmt()
		}
		if p.peekToken().Type == lexer.TOKEN_LSHIFT {
			return p.parsePushStmt()
		}
		if p.peekToken().Type == lexer.TOKEN_AT {
			// Could be key-set (a@key = val), dyn-key-set (a@$var = val), index-set (a@i = val), or expression
			saved := p.save()
			savedErrors := len(p.errors)
			name := p.curToken().Literal
			p.advance() // consume ident
			p.advance() // consume @

			// Check for dynamic key set: ident @ $ ident =
			if p.curToken().Type == lexer.TOKEN_DOLLAR {
				p.advance() // consume $
				if p.curToken().Type == lexer.TOKEN_IDENT {
					keyVarName := p.curToken().Literal
					p.advance() // consume key ident
					if p.curToken().Type == lexer.TOKEN_ASSIGN {
						p.advance() // consume =
						val := p.parseExpression()
						if val != nil {
							return &DynKeySetStmt{
								Object: &Identifier{Name: name},
								KeyVar: &Identifier{Name: keyVarName},
								Value:  val,
							}
						}
					}
				}
				// Not a dyn-key-set; restore
				p.restore(saved)
				p.errors = p.errors[:savedErrors]
			} else if p.curToken().Type == lexer.TOKEN_IDENT && p.peekToken().Type != lexer.TOKEN_LPAREN {
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

			// If we haven't restored yet, we're still advanced past ident and @
			// Need to check if we're already restored or not
			if p.pos != saved {
				// We're still past @; try index-set: ident @ expr =
				idx := p.parsePostfixPrimary()
				if idx != nil && p.curToken().Type == lexer.TOKEN_ASSIGN {
					p.advance() // consume =
					val := p.parseExpression()
					if val != nil {
						return &IndexSetStmt{
							Object: &Identifier{Name: name},
							Index:  idx,
							Value:  val,
						}
					}
				}
				// Not an index-set; restore and parse as expression
				p.restore(saved)
				p.errors = p.errors[:savedErrors]
			}
		}
		if p.peekToken().Type == lexer.TOKEN_COMMA {
			return p.parseMultiAssign()
		}
		if p.peekToken().Type == lexer.TOKEN_ASSIGN {
			return p.parseAssignStatement()
		}
		// Bare expression statement (e.g., function call)
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		return &ExprStmt{Expr: expr}
	case lexer.TOKEN_RSHIFT, lexer.TOKEN_HASH, lexer.TOKEN_TILDE, lexer.TOKEN_DOUBLE_HASH, lexer.TOKEN_DOUBLE_AT, lexer.TOKEN_NULL:
		// Prefix operators and literals that start expression statements
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		return &ExprStmt{Expr: expr}
	default:
		p.addError("unexpected token %s (%q) at line %d", p.curToken().Type, p.curToken().Literal, p.curToken().Line)
		p.advance()
		return nil
	}
}

func (p *Parser) parsePushStmt() *PushStmt {
	name := p.curToken().Literal
	p.advance() // consume ident
	p.advance() // consume <<
	value := p.parseExpression()
	if value == nil {
		return nil
	}
	return &PushStmt{Object: &Identifier{Name: name}, Value: value}
}

func (p *Parser) parseAssignStatement() *AssignStmt {
	name := p.curToken().Literal
	p.advance() // consume identifier

	if p.curToken().Type != lexer.TOKEN_ASSIGN {
		p.addError("expected '=', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '='

	value := p.parseExpression()
	if value == nil {
		return nil
	}

	return &AssignStmt{Name: name, Value: value}
}

func (p *Parser) peekToken() lexer.Token {
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos+1]
}

func (p *Parser) parseStdoutWriteStatement() *StdoutWriteStmt {
	p.advance() // consume '|>'

	value := p.parseExpression()
	if value == nil {
		return nil
	}

	return &StdoutWriteStmt{Value: value}
}

func (p *Parser) parseBlock() []Stmt {
	if p.curToken().Type != lexer.TOKEN_LBRACE {
		p.addError("expected '{', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '{'

	var stmts []Stmt
	for p.curToken().Type != lexer.TOKEN_RBRACE && p.curToken().Type != lexer.TOKEN_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	if p.curToken().Type != lexer.TOKEN_RBRACE {
		p.addError("expected '}', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '}'
	return stmts
}

func (p *Parser) parseFnDecl() *FnDeclStmt {
	p.advance() // consume 'fn'

	if p.curToken().Type != lexer.TOKEN_IDENT {
		p.addError("expected function name, got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	name := p.curToken().Literal
	p.advance() // consume name

	if p.curToken().Type != lexer.TOKEN_LPAREN {
		p.addError("expected '(', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '('

	var params []string
	if p.curToken().Type != lexer.TOKEN_RPAREN {
		if p.curToken().Type != lexer.TOKEN_IDENT {
			p.addError("expected parameter name, got %s at line %d", p.curToken().Type, p.curToken().Line)
			return nil
		}
		params = append(params, p.curToken().Literal)
		p.advance()
		for p.curToken().Type == lexer.TOKEN_COMMA {
			p.advance() // consume ','
			if p.curToken().Type != lexer.TOKEN_IDENT {
				p.addError("expected parameter name, got %s at line %d", p.curToken().Type, p.curToken().Line)
				return nil
			}
			params = append(params, p.curToken().Literal)
			p.advance()
		}
	}

	if p.curToken().Type != lexer.TOKEN_RPAREN {
		p.addError("expected ')', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume ')'

	body := p.parseBlock()
	return &FnDeclStmt{Name: name, Params: params, Body: body}
}

func (p *Parser) parseIfStmt() *IfStmt {
	p.advance() // consume 'if'

	condition := p.parseExpression()
	if condition == nil {
		return nil
	}

	body := p.parseBlock()

	var elseBody []Stmt
	if p.curToken().Type == lexer.TOKEN_ELSE {
		p.advance() // consume 'else'
		elseBody = p.parseBlock()
	}

	return &IfStmt{Condition: condition, Body: body, Else: elseBody}
}

func (p *Parser) parseForStmt() Stmt {
	// Peek ahead: for { ... } is infinite loop, for IDENT in ... is for-in
	p.advance() // consume 'for'
	if p.curToken().Type == lexer.TOKEN_LBRACE {
		return p.parseForLoopBody()
	}
	return p.parseForInBody()
}

func (p *Parser) parseForLoopBody() *ForLoopStmt {
	body := p.parseBlock()
	return &ForLoopStmt{Body: body}
}

func (p *Parser) parseBreakStmt() *BreakStmt {
	p.advance() // consume 'break'
	return &BreakStmt{}
}

func (p *Parser) parseForInBody() *ForInStmt {

	if p.curToken().Type != lexer.TOKEN_IDENT {
		p.addError("expected loop variable, got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	varName := p.curToken().Literal
	p.advance() // consume variable name

	if p.curToken().Type != lexer.TOKEN_IN {
		p.addError("expected 'in', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume 'in'

	iterable := p.parseExpression()
	if iterable == nil {
		return nil
	}

	body := p.parseBlock()
	return &ForInStmt{VarName: varName, Iterable: iterable, Body: body}
}

func (p *Parser) parseReturnStmt() *ReturnStmt {
	p.advance() // consume '<-'

	// If next token can't start an expression, bare return
	if p.curToken().Type == lexer.TOKEN_RBRACE || p.curToken().Type == lexer.TOKEN_EOF {
		return &ReturnStmt{}
	}

	value := p.parseExpression()
	return &ReturnStmt{Value: value}
}

func (p *Parser) parseMultiAssign() *MultiAssignStmt {
	var names []string
	names = append(names, p.curToken().Literal)
	p.advance() // consume first ident

	for p.curToken().Type == lexer.TOKEN_COMMA {
		p.advance() // consume ','
		if p.curToken().Type != lexer.TOKEN_IDENT {
			p.addError("expected identifier in multi-assign, got %s at line %d", p.curToken().Type, p.curToken().Line)
			return nil
		}
		names = append(names, p.curToken().Literal)
		p.advance()
	}

	if p.curToken().Type != lexer.TOKEN_ASSIGN {
		p.addError("expected '=' in multi-assign, got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '='

	value := p.parseExpression()
	if value == nil {
		return nil
	}
	return &MultiAssignStmt{Names: names, Value: value}
}

func (p *Parser) parseHashDeclStmt() *HashDeclStmt {
	name := p.curToken().Literal
	p.advance() // consume ident
	p.advance() // consume '{'

	var keys []string
	if p.curToken().Type != lexer.TOKEN_RBRACE {
		if p.curToken().Type != lexer.TOKEN_IDENT {
			p.addError("expected key name, got %s at line %d", p.curToken().Type, p.curToken().Line)
			return nil
		}
		keys = append(keys, p.curToken().Literal)
		p.advance()
		for p.curToken().Type == lexer.TOKEN_COMMA {
			p.advance() // consume ','
			if p.curToken().Type != lexer.TOKEN_IDENT {
				p.addError("expected key name, got %s at line %d", p.curToken().Type, p.curToken().Line)
				return nil
			}
			keys = append(keys, p.curToken().Literal)
			p.advance()
		}
	}

	if p.curToken().Type != lexer.TOKEN_RBRACE {
		p.addError("expected '}', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '}'

	if p.curToken().Type != lexer.TOKEN_ASSIGN {
		p.addError("expected '=', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume '='

	value := p.parseExpression()
	if value == nil {
		return nil
	}

	return &HashDeclStmt{Name: name, Keys: keys, Value: value}
}

func (p *Parser) parseFileWriteStmt() *FileWriteStmt {
	p.advance() // consume .>
	path := p.parseExpression()
	if path == nil {
		return nil
	}
	data := p.parseExpression()
	if data == nil {
		return nil
	}
	return &FileWriteStmt{Path: path, Data: data}
}

func (p *Parser) parseFileAppendStmt() *FileAppendStmt {
	p.advance() // consume .>>
	path := p.parseExpression()
	if path == nil {
		return nil
	}
	data := p.parseExpression()
	if data == nil {
		return nil
	}
	return &FileAppendStmt{Path: path, Data: data}
}

func (p *Parser) parseExpression() Expr {
	return p.parseOr()
}

func (p *Parser) parseOr() Expr {
	left := p.parseAnd()
	if left == nil {
		return nil
	}
	for p.curToken().Type == lexer.TOKEN_OR {
		op := p.curToken().Literal
		p.advance()
		right := p.parseAnd()
		if right == nil {
			return nil
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() Expr {
	left := p.parseComparison()
	if left == nil {
		return nil
	}
	for p.curToken().Type == lexer.TOKEN_AND {
		op := p.curToken().Literal
		p.advance()
		right := p.parseComparison()
		if right == nil {
			return nil
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() Expr {
	left := p.parseAddSub()
	if left == nil {
		return nil
	}
	for p.curToken().Type == lexer.TOKEN_EQ ||
		p.curToken().Type == lexer.TOKEN_NEQ ||
		p.curToken().Type == lexer.TOKEN_LT ||
		p.curToken().Type == lexer.TOKEN_GT ||
		p.curToken().Type == lexer.TOKEN_LTE ||
		p.curToken().Type == lexer.TOKEN_GTE {
		op := p.curToken().Literal
		p.advance()
		right := p.parseAddSub()
		if right == nil {
			return nil
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseAddSub() Expr {
	left := p.parseMulDivMod()
	if left == nil {
		return nil
	}
	for p.curToken().Type == lexer.TOKEN_PLUS ||
		p.curToken().Type == lexer.TOKEN_MINUS ||
		p.curToken().Type == lexer.TOKEN_CONCAT ||
		p.curToken().Type == lexer.TOKEN_REMOVE ||
		p.curToken().Type == lexer.TOKEN_CONTAINS ||
		p.curToken().Type == lexer.TOKEN_TILDE_AT {
		op := p.curToken().Literal
		p.advance()
		right := p.parseMulDivMod()
		if right == nil {
			return nil
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parseMulDivMod() Expr {
	left := p.parsePower()
	if left == nil {
		return nil
	}
	for p.curToken().Type == lexer.TOKEN_STAR ||
		p.curToken().Type == lexer.TOKEN_SLASH ||
		p.curToken().Type == lexer.TOKEN_PERCENT {
		op := p.curToken().Literal
		p.advance()
		right := p.parsePower()
		if right == nil {
			return nil
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left
}

func (p *Parser) parsePower() Expr {
	base := p.parseUnary()
	if base == nil {
		return nil
	}
	if p.curToken().Type == lexer.TOKEN_POWER {
		op := p.curToken().Literal
		p.advance()
		exp := p.parsePower() // right-associative
		if exp == nil {
			return nil
		}
		return &BinaryExpr{Left: base, Op: op, Right: exp}
	}
	return base
}

func (p *Parser) parseUnary() Expr {
	if p.curToken().Type == lexer.TOKEN_MINUS ||
		p.curToken().Type == lexer.TOKEN_NOT {
		op := p.curToken().Literal
		p.advance()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpr{Op: op, Operand: operand}
	}
	if p.curToken().Type == lexer.TOKEN_HASH ||
		p.curToken().Type == lexer.TOKEN_TILDE {
		op := p.curToken().Literal
		p.advance()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpr{Op: op, Operand: operand}
	}
	if p.curToken().Type == lexer.TOKEN_DOUBLE_HASH ||
		p.curToken().Type == lexer.TOKEN_DOUBLE_AT {
		op := p.curToken().Literal
		p.advance()
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &UnaryExpr{Op: op, Operand: operand}
	}
	if p.curToken().Type == lexer.TOKEN_RSHIFT {
		p.advance() // consume >>
		operand := p.parseUnary()
		if operand == nil {
			return nil
		}
		return &PopExpr{Object: operand}
	}
	return p.parsePostfix()
}

// parsePostfixPrimary parses a simple expression for use as an index or slice bound.
// It handles numeric literals, identifiers, function calls, and array literals.
func (p *Parser) parsePostfixPrimary() Expr {
	switch p.curToken().Type {
	case lexer.TOKEN_INT, lexer.TOKEN_UINT, lexer.TOKEN_FLOAT:
		return p.parseNumberLiteral()
	case lexer.TOKEN_IDENT:
		if p.peekToken().Type == lexer.TOKEN_LPAREN {
			return p.parseCall() // function call as index
		}
		return p.parseIdentifier()
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral()
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_LPAREN:
		p.advance() // consume '('
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		if p.curToken().Type != lexer.TOKEN_RPAREN {
			p.addError("expected ')' in postfix expression, got %s at line %d", p.curToken().Type, p.curToken().Line)
			return nil
		}
		p.advance() // consume ')'
		return expr
	default:
		p.addError("unexpected token %s (%q) in postfix expression at line %d", p.curToken().Type, p.curToken().Literal, p.curToken().Line)
		return nil
	}
}

func (p *Parser) parsePostfix() Expr {
	expr := p.parseCall()
	if expr == nil {
		return nil
	}
	for {
		if p.curToken().Type == lexer.TOKEN_AT {
			p.advance() // consume @
			// Check for dynamic key access: @$ident
			if p.curToken().Type == lexer.TOKEN_DOLLAR && p.peekToken().Type == lexer.TOKEN_IDENT {
				p.advance() // consume $
				keyVarName := p.curToken().Literal
				p.advance() // consume ident
				expr = &DynKeyAccessExpr{Object: expr, KeyVar: &Identifier{Name: keyVarName}}
			} else if p.curToken().Type == lexer.TOKEN_IDENT && p.peekToken().Type != lexer.TOKEN_LPAREN {
				// Static key access: @ident (but not @ident( which is a function call)
				keyName := p.curToken().Literal
				p.advance() // consume ident
				expr = &KeyAccessExpr{Object: expr, Key: keyName}
			} else {
				// Numeric/expression index access
				idx := p.parsePostfixPrimary()
				if idx == nil {
					return nil
				}
				expr = &IndexExpr{Object: expr, Index: idx}
			}
		} else if p.curToken().Type == lexer.TOKEN_DCOLON {
			p.advance() // consume first ::
			low := p.parsePostfixPrimary()
			if low == nil {
				return nil
			}
			if p.curToken().Type != lexer.TOKEN_DCOLON {
				p.addError("expected '::' after slice low bound, got %s at line %d", p.curToken().Type, p.curToken().Line)
				return nil
			}
			p.advance() // consume second ::
			high := p.parsePostfixPrimary()
			if high == nil {
				return nil
			}
			expr = &SliceExpr{Object: expr, Low: low, High: high}
		} else {
			break
		}
	}
	return expr
}

func (p *Parser) parseCall() Expr {
	if p.curToken().Type == lexer.TOKEN_IDENT && p.peekToken().Type == lexer.TOKEN_LPAREN {
		name := p.curToken().Literal
		p.advance() // consume ident
		p.advance() // consume '('

		var args []Expr
		if p.curToken().Type != lexer.TOKEN_RPAREN {
			arg := p.parseExpression()
			if arg == nil {
				return nil
			}
			args = append(args, arg)
			for p.curToken().Type == lexer.TOKEN_COMMA {
				p.advance()
				arg = p.parseExpression()
				if arg == nil {
					return nil
				}
				args = append(args, arg)
			}
		}

		if p.curToken().Type != lexer.TOKEN_RPAREN {
			p.addError("expected ')', got %s at line %d", p.curToken().Type, p.curToken().Line)
			return nil
		}
		p.advance() // consume ')'
		return &CallExpr{Name: name, Args: args}
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() Expr {
	switch p.curToken().Type {
	case lexer.TOKEN_LPAREN:
		p.advance() // consume '('
		expr := p.parseExpression()
		if expr == nil {
			return nil
		}
		if p.curToken().Type != lexer.TOKEN_RPAREN {
			p.addError("expected ')', got %s at line %d", p.curToken().Type, p.curToken().Line)
			return nil
		}
		p.advance() // consume ')'
		return expr
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral()
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_INT, lexer.TOKEN_UINT, lexer.TOKEN_FLOAT:
		return p.parseNumberLiteral()
	case lexer.TOKEN_IDENT:
		return p.parseIdentifier()
	case lexer.TOKEN_TRUE, lexer.TOKEN_FALSE:
		return p.parseBoolLiteral()
	case lexer.TOKEN_NULL:
		p.advance()
		return &NullLiteral{}
	case lexer.TOKEN_STDIN_READ:
		p.advance()
		return &StdinReadExpr{}
	case lexer.TOKEN_FILE_READ:
		p.advance() // consume <.
		path := p.parseExpression()
		if path == nil {
			return nil
		}
		return &FileReadExpr{Path: path}
	default:
		p.addError("unexpected token %s (%q) in expression at line %d", p.curToken().Type, p.curToken().Literal, p.curToken().Line)
		return nil
	}
}

func (p *Parser) parseArrayLiteral() *ArrayLiteral {
	p.advance() // consume '['

	al := &ArrayLiteral{}

	if p.curToken().Type == lexer.TOKEN_RBRACKET {
		p.advance() // consume ']'
		return al
	}

	elem := p.parseExpression()
	if elem == nil {
		return nil
	}
	al.Elements = append(al.Elements, elem)

	for p.curToken().Type == lexer.TOKEN_COMMA {
		p.advance() // consume ','
		elem = p.parseExpression()
		if elem == nil {
			return nil
		}
		al.Elements = append(al.Elements, elem)
	}

	if p.curToken().Type != lexer.TOKEN_RBRACKET {
		p.addError("expected ']', got %s at line %d", p.curToken().Type, p.curToken().Line)
		return nil
	}
	p.advance() // consume ']'

	return al
}

func (p *Parser) parseStringLiteral() *StringLiteral {
	sl := &StringLiteral{Value: p.curToken().Literal}
	p.advance()
	return sl
}

func (p *Parser) parseNumberLiteral() *NumberLiteral {
	tok := p.curToken()
	nl := &NumberLiteral{Value: tok.Literal}
	switch tok.Type {
	case lexer.TOKEN_INT:
		nl.NumType = NumInt
	case lexer.TOKEN_UINT:
		nl.NumType = NumUint
	case lexer.TOKEN_FLOAT:
		nl.NumType = NumFloat
	}
	p.advance()
	return nl
}

func (p *Parser) parseIdentifier() *Identifier {
	id := &Identifier{Name: p.curToken().Literal}
	p.advance()
	return id
}

func (p *Parser) parseBoolLiteral() *ArrayLiteral {
	// true = [1], false = []
	tok := p.curToken()
	p.advance()
	if tok.Type == lexer.TOKEN_TRUE {
		return &ArrayLiteral{
			Elements: []Expr{&NumberLiteral{Value: "1", NumType: NumInt}},
		}
	}
	return &ArrayLiteral{}
}

func (p *Parser) curToken() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

func (p *Parser) peekTokenAt(offset int) lexer.Token {
	idx := p.pos + offset
	if idx >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[idx]
}

func (p *Parser) save() int       { return p.pos }
func (p *Parser) restore(pos int) { p.pos = pos }

func (p *Parser) addError(format string, args ...any) {
	p.errors = append(p.errors, fmt.Sprintf(format, args...))
}
