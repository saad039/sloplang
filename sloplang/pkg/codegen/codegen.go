package codegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strconv"
	"strings"

	"github.com/saad039/sloplang/pkg/parser"
)

// Generator produces Go source code from a sloplang AST.
type Generator struct {
	modulePath string
}

// New creates a new Generator with the given Go module path.
func New(modulePath string) *Generator {
	return &Generator{modulePath: modulePath}
}

// Generate takes a sloplang AST and returns formatted Go source code.
func (g *Generator) Generate(program *parser.Program) ([]byte, error) {
	stmts := make([]ast.Stmt, 0, len(program.Statements)*2)
	for _, s := range program.Statements {
		lowered := g.lowerStmt(s)
		stmts = append(stmts, lowered...)
	}

	mainFunc := &ast.FuncDecl{
		Name: ast.NewIdent("main"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: stmts},
	}

	importDecl := &ast.GenDecl{
		Tok: token.IMPORT,
		Specs: []ast.Spec{
			&ast.ImportSpec{
				Name: ast.NewIdent("sloprt"),
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: strconv.Quote(g.modulePath + "/pkg/runtime"),
				},
			},
		},
	}

	file := &ast.File{
		Name:  ast.NewIdent("main"),
		Decls: []ast.Decl{importDecl, mainFunc},
	}

	var buf bytes.Buffer
	fset := token.NewFileSet()
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}
	return buf.Bytes(), nil
}

func (g *Generator) lowerStmt(stmt parser.Stmt) []ast.Stmt {
	switch s := stmt.(type) {
	case *parser.AssignStmt:
		assign := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent(s.Name)},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{g.lowerExpr(s.Value)},
		}
		// Suppress Go's "declared and not used" error
		suppress := &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(s.Name)},
		}
		return []ast.Stmt{assign, suppress}
	case *parser.StdoutWriteStmt:
		return []ast.Stmt{
			&ast.ExprStmt{
				X: callSloprt("StdoutWrite", g.lowerExpr(s.Value)),
			},
		}
	default:
		return nil
	}
}

func (g *Generator) lowerExpr(expr parser.Expr) ast.Expr {
	switch e := expr.(type) {
	case *parser.ArrayLiteral:
		args := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			args[i] = g.lowerRawValue(elem)
		}
		return callSloprt("NewSlopValue", args...)
	case *parser.NumberLiteral:
		return callSloprt("NewSlopValue", g.lowerRawValue(e))
	case *parser.StringLiteral:
		return callSloprt("NewSlopValue", g.lowerRawValue(e))
	case *parser.Identifier:
		return ast.NewIdent(e.Name)
	default:
		return ast.NewIdent("nil")
	}
}

// lowerRawValue returns the raw Go expression for an element inside an array,
// without wrapping in NewSlopValue. This avoids double-wrapping.
func (g *Generator) lowerRawValue(expr parser.Expr) ast.Expr {
	switch e := expr.(type) {
	case *parser.NumberLiteral:
		return g.lowerNumberRaw(e)
	case *parser.StringLiteral:
		return &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(e.Value),
		}
	case *parser.ArrayLiteral:
		// Nested arrays become *SlopValue elements
		args := make([]ast.Expr, len(e.Elements))
		for i, elem := range e.Elements {
			args[i] = g.lowerRawValue(elem)
		}
		return callSloprt("NewSlopValue", args...)
	case *parser.Identifier:
		return ast.NewIdent(e.Name)
	default:
		return ast.NewIdent("nil")
	}
}

func (g *Generator) lowerNumberRaw(nl *parser.NumberLiteral) ast.Expr {
	switch nl.NumType {
	case parser.NumInt:
		return typeCast("int64", nl.Value)
	case parser.NumUint:
		// Strip the 'u' suffix for the Go literal
		return typeCast("uint64", strings.TrimSuffix(nl.Value, "u"))
	case parser.NumFloat:
		return typeCast("float64", nl.Value)
	default:
		return typeCast("int64", nl.Value)
	}
}

func typeCast(typeName, value string) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:  ast.NewIdent(typeName),
		Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: value}},
	}
}

func callSloprt(funcName string, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   ast.NewIdent("sloprt"),
			Sel: ast.NewIdent(funcName),
		},
		Args: args,
	}
}
