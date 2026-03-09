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
	case lexer.TOKEN_IDENT:
		return p.parseAssignStatement()
	case lexer.TOKEN_PIPE_GT:
		return p.parseStdoutWriteStatement()
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

func (p *Parser) parseStdoutWriteStatement() *StdoutWriteStmt {
	p.advance() // consume '|>'

	value := p.parseExpression()
	if value == nil {
		return nil
	}

	return &StdoutWriteStmt{Value: value}
}

func (p *Parser) parseExpression() Expr {
	switch p.curToken().Type {
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral()
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_INT, lexer.TOKEN_UINT, lexer.TOKEN_FLOAT:
		return p.parseNumberLiteral()
	case lexer.TOKEN_IDENT:
		return p.parseIdentifier()
	case lexer.TOKEN_TRUE:
		return p.parseBoolLiteral()
	case lexer.TOKEN_FALSE:
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
