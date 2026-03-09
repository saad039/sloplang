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
		return p.parseForInStmt()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStmt()
	case lexer.TOKEN_PIPE_GT:
		return p.parseStdoutWriteStatement()
	case lexer.TOKEN_IDENT:
		// Disambiguate: multi-assign (a, b = ...), single assign (a = ...), or bare expression (fn call)
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
	default:
		p.addError("unexpected token %s (%q) at line %d", p.curToken().Type, p.curToken().Literal, p.curToken().Line)
		p.advance()
		return nil
	}
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

func (p *Parser) parseForInStmt() *ForInStmt {
	p.advance() // consume 'for'

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
		p.curToken().Type == lexer.TOKEN_MINUS {
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
	return p.parseCall()
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

func (p *Parser) addError(format string, args ...any) {
	p.errors = append(p.errors, fmt.Sprintf(format, args...))
}
