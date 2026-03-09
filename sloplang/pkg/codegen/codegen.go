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
	declared   map[string]bool // tracks variables that have been declared
}

// New creates a new Generator with the given Go module path.
func New(modulePath string) *Generator {
	return &Generator{modulePath: modulePath, declared: make(map[string]bool)}
}

// Generate takes a sloplang AST and returns formatted Go source code.
func (g *Generator) Generate(program *parser.Program) ([]byte, error) {
	var fnDecls []ast.Decl
	mainStmts := make([]ast.Stmt, 0, len(program.Statements)*2)

	for _, s := range program.Statements {
		if fd, ok := s.(*parser.FnDeclStmt); ok {
			fnDecls = append(fnDecls, g.lowerFnDecl(fd))
		} else {
			lowered := g.lowerStmt(s)
			mainStmts = append(mainStmts, lowered...)
		}
	}

	mainFunc := &ast.FuncDecl{
		Name: ast.NewIdent("main"),
		Type: &ast.FuncType{Params: &ast.FieldList{}},
		Body: &ast.BlockStmt{List: mainStmts},
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

	decls := []ast.Decl{importDecl}
	decls = append(decls, fnDecls...)
	decls = append(decls, mainFunc)

	file := &ast.File{
		Name:  ast.NewIdent("main"),
		Decls: decls,
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
	case *parser.StdoutWriteStmt:
		return []ast.Stmt{
			&ast.ExprStmt{
				X: callSloprt("StdoutWrite", g.lowerExpr(s.Value)),
			},
		}
	case *parser.IfStmt:
		return []ast.Stmt{g.lowerIfStmt(s)}
	case *parser.ForInStmt:
		return []ast.Stmt{g.lowerForInStmt(s)}
	case *parser.ForLoopStmt:
		return []ast.Stmt{g.lowerForLoopStmt(s)}
	case *parser.BreakStmt:
		return []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}
	case *parser.ReturnStmt:
		return g.lowerReturnStmt(s)
	case *parser.MultiAssignStmt:
		return g.lowerMultiAssign(s)
	case *parser.PushStmt:
		return []ast.Stmt{
			&ast.ExprStmt{X: callSloprt("Push", g.lowerExpr(s.Object), g.lowerExpr(s.Value))},
		}
	case *parser.IndexSetStmt:
		return []ast.Stmt{
			&ast.ExprStmt{X: callSloprt("IndexSet", g.lowerExpr(s.Object), g.lowerExpr(s.Index), g.lowerExpr(s.Value))},
		}
	case *parser.ExprStmt:
		return []ast.Stmt{
			&ast.ExprStmt{X: g.lowerExpr(s.Expr)},
		}
	default:
		return nil
	}
}

func (g *Generator) lowerFnDecl(fd *parser.FnDeclStmt) *ast.FuncDecl {
	// Build parameter list: each param is *sloprt.SlopValue
	params := make([]*ast.Field, len(fd.Params))
	for i, p := range fd.Params {
		params[i] = &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(p)},
			Type:  slopValuePtrType(),
		}
	}

	// Save outer scope and create a new scope for the function
	outerDeclared := g.declared
	g.declared = make(map[string]bool)
	for _, p := range fd.Params {
		g.declared[p] = true // params are already declared
	}

	var bodyStmts []ast.Stmt
	for _, s := range fd.Body {
		bodyStmts = append(bodyStmts, g.lowerStmt(s)...)
	}

	// Restore outer scope
	g.declared = outerDeclared

	return &ast.FuncDecl{
		Name: ast.NewIdent(fd.Name),
		Type: &ast.FuncType{
			Params: &ast.FieldList{List: params},
			Results: &ast.FieldList{
				List: []*ast.Field{{Type: slopValuePtrType()}},
			},
		},
		Body: &ast.BlockStmt{List: bodyStmts},
	}
}

func (g *Generator) lowerIfStmt(s *parser.IfStmt) *ast.IfStmt {
	// Condition: (loweredExpr).IsTruthy()
	cond := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   g.lowerExpr(s.Condition),
			Sel: ast.NewIdent("IsTruthy"),
		},
	}

	var bodyStmts []ast.Stmt
	for _, stmt := range s.Body {
		bodyStmts = append(bodyStmts, g.lowerStmt(stmt)...)
	}

	ifStmt := &ast.IfStmt{
		Cond: cond,
		Body: &ast.BlockStmt{List: bodyStmts},
	}

	if len(s.Else) > 0 {
		var elseStmts []ast.Stmt
		for _, stmt := range s.Else {
			elseStmts = append(elseStmts, g.lowerStmt(stmt)...)
		}
		ifStmt.Else = &ast.BlockStmt{List: elseStmts}
	}

	return ifStmt
}

func (g *Generator) lowerForInStmt(s *parser.ForInStmt) *ast.RangeStmt {
	// Suppress "declared and not used" for loop variable
	suppress := &ast.AssignStmt{
		Lhs: []ast.Expr{ast.NewIdent("_")},
		Tok: token.ASSIGN,
		Rhs: []ast.Expr{ast.NewIdent(s.VarName)},
	}
	bodyStmts := []ast.Stmt{suppress}
	for _, stmt := range s.Body {
		bodyStmts = append(bodyStmts, g.lowerStmt(stmt)...)
	}

	return &ast.RangeStmt{
		Key:   ast.NewIdent("_"),
		Value: ast.NewIdent(s.VarName),
		Tok:   token.DEFINE,
		X:     callSloprt("Iterate", g.lowerExpr(s.Iterable)),
		Body:  &ast.BlockStmt{List: bodyStmts},
	}
}

func (g *Generator) lowerForLoopStmt(s *parser.ForLoopStmt) *ast.ForStmt {
	var bodyStmts []ast.Stmt
	for _, stmt := range s.Body {
		bodyStmts = append(bodyStmts, g.lowerStmt(stmt)...)
	}
	return &ast.ForStmt{
		Body: &ast.BlockStmt{List: bodyStmts},
	}
}

func (g *Generator) lowerReturnStmt(s *parser.ReturnStmt) []ast.Stmt {
	if s.Value == nil {
		return []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{callSloprt("NewSlopValue")},
			},
		}
	}
	return []ast.Stmt{
		&ast.ReturnStmt{
			Results: []ast.Expr{g.lowerExpr(s.Value)},
		},
	}
}

func (g *Generator) lowerMultiAssign(s *parser.MultiAssignStmt) []ast.Stmt {
	// a, b = expr  →  a, b := sloprt.UnpackTwo(loweredExpr)
	lhs := make([]ast.Expr, len(s.Names))
	for i, name := range s.Names {
		lhs[i] = ast.NewIdent(name)
	}

	assign := &ast.AssignStmt{
		Lhs: lhs,
		Tok: token.DEFINE,
		Rhs: []ast.Expr{callSloprt("UnpackTwo", g.lowerExpr(s.Value))},
	}

	// Suppress unused variable errors
	var stmts []ast.Stmt
	stmts = append(stmts, assign)
	for _, name := range s.Names {
		stmts = append(stmts, &ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("_")},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{ast.NewIdent(name)},
		})
	}
	return stmts
}

func slopValuePtrType() *ast.StarExpr {
	return &ast.StarExpr{
		X: &ast.SelectorExpr{
			X:   ast.NewIdent("sloprt"),
			Sel: ast.NewIdent("SlopValue"),
		},
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
	case *parser.IndexExpr:
		return callSloprt("Index", g.lowerExpr(e.Object), g.lowerExpr(e.Index))
	case *parser.PopExpr:
		return callSloprt("Pop", g.lowerExpr(e.Object))
	case *parser.SliceExpr:
		return callSloprt("Slice", g.lowerExpr(e.Object), g.lowerExpr(e.Low), g.lowerExpr(e.High))
	case *parser.BinaryExpr:
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
		return callSloprt(fname, g.lowerExpr(e.Left), g.lowerExpr(e.Right))
	case *parser.UnaryExpr:
		switch e.Op {
		case "-":
			return callSloprt("Negate", g.lowerExpr(e.Operand))
		case "#":
			return callSloprt("Length", g.lowerExpr(e.Operand))
		case "~":
			return callSloprt("Unique", g.lowerExpr(e.Operand))
		default:
			return callSloprt("Not", g.lowerExpr(e.Operand))
		}
	case *parser.CallExpr:
		args := make([]ast.Expr, len(e.Args))
		for i, arg := range e.Args {
			args[i] = g.lowerExpr(arg)
		}
		builtins := map[string]string{"str": "Str"}
		if fname, ok := builtins[e.Name]; ok {
			return callSloprt(fname, args...)
		}
		// User-defined function call
		return &ast.CallExpr{
			Fun:  ast.NewIdent(e.Name),
			Args: args,
		}
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
	case *parser.BinaryExpr:
		return g.lowerExpr(e)
	case *parser.UnaryExpr:
		return g.lowerExpr(e)
	case *parser.CallExpr:
		return g.lowerExpr(e)
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
