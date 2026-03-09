package parser

import (
	"testing"

	"github.com/saad039/sloplang/pkg/lexer"
)

func parse(input string) (*Program, []string) {
	l := lexer.New(input)
	tokens := l.Tokenize()
	p := New(tokens)
	return p.Parse()
}

func TestParser_AssignIntArray(t *testing.T) {
	prog, errs := parse(`x = [1, 2, 3]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}

	assign, ok := prog.Statements[0].(*AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", prog.Statements[0])
	}
	if assign.Name != "x" {
		t.Fatalf("expected name 'x', got %q", assign.Name)
	}

	arr, ok := assign.Value.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", assign.Value)
	}
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	for i, expected := range []string{"1", "2", "3"} {
		nl, ok := arr.Elements[i].(*NumberLiteral)
		if !ok {
			t.Fatalf("element %d: expected NumberLiteral, got %T", i, arr.Elements[i])
		}
		if nl.Value != expected || nl.NumType != NumInt {
			t.Fatalf("element %d: expected Int(%s), got %d(%s)", i, expected, nl.NumType, nl.Value)
		}
	}
}

func TestParser_AssignString(t *testing.T) {
	prog, errs := parse(`name = "hello"`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	assign := prog.Statements[0].(*AssignStmt)
	sl, ok := assign.Value.(*StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", assign.Value)
	}
	if sl.Value != "hello" {
		t.Fatalf("expected 'hello', got %q", sl.Value)
	}
}

func TestParser_StdoutWriteString(t *testing.T) {
	prog, errs := parse(`|> "hello world"`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	sw, ok := prog.Statements[0].(*StdoutWriteStmt)
	if !ok {
		t.Fatalf("expected StdoutWriteStmt, got %T", prog.Statements[0])
	}
	sl, ok := sw.Value.(*StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", sw.Value)
	}
	if sl.Value != "hello world" {
		t.Fatalf("expected 'hello world', got %q", sl.Value)
	}
}

func TestParser_StdoutWriteIdent(t *testing.T) {
	prog, errs := parse(`|> x`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	sw := prog.Statements[0].(*StdoutWriteStmt)
	id, ok := sw.Value.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier, got %T", sw.Value)
	}
	if id.Name != "x" {
		t.Fatalf("expected 'x', got %q", id.Name)
	}
}

func TestParser_EmptyArray(t *testing.T) {
	prog, errs := parse(`x = []`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	assign := prog.Statements[0].(*AssignStmt)
	arr, ok := assign.Value.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", assign.Value)
	}
	if len(arr.Elements) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(arr.Elements))
	}
}

func TestParser_MixedArray(t *testing.T) {
	prog, errs := parse(`x = [1, 3.14, "hi"]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	assign := prog.Statements[0].(*AssignStmt)
	arr := assign.Value.(*ArrayLiteral)
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}

	if nl, ok := arr.Elements[0].(*NumberLiteral); !ok || nl.NumType != NumInt {
		t.Fatalf("element 0: expected Int, got %T", arr.Elements[0])
	}
	if nl, ok := arr.Elements[1].(*NumberLiteral); !ok || nl.NumType != NumFloat {
		t.Fatalf("element 1: expected Float, got %T", arr.Elements[1])
	}
	if _, ok := arr.Elements[2].(*StringLiteral); !ok {
		t.Fatalf("element 2: expected StringLiteral, got %T", arr.Elements[2])
	}
}

func TestParser_UintLiteral(t *testing.T) {
	prog, errs := parse(`x = [42u]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	assign := prog.Statements[0].(*AssignStmt)
	arr := assign.Value.(*ArrayLiteral)
	nl := arr.Elements[0].(*NumberLiteral)
	if nl.NumType != NumUint {
		t.Fatalf("expected Uint, got %v", nl.NumType)
	}
	if nl.Value != "42u" {
		t.Fatalf("expected '42u', got %q", nl.Value)
	}
}

func TestParser_MultipleStatements(t *testing.T) {
	prog, errs := parse("x = [1, 2, 3]\n|> \"hello world\"")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}

	if _, ok := prog.Statements[0].(*AssignStmt); !ok {
		t.Fatalf("statement 0: expected AssignStmt, got %T", prog.Statements[0])
	}
	if _, ok := prog.Statements[1].(*StdoutWriteStmt); !ok {
		t.Fatalf("statement 1: expected StdoutWriteStmt, got %T", prog.Statements[1])
	}
}

func TestParser_BoolTrue(t *testing.T) {
	prog, errs := parse(`x = true`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	assign := prog.Statements[0].(*AssignStmt)
	arr, ok := assign.Value.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral (true -> [1]), got %T", assign.Value)
	}
	if len(arr.Elements) != 1 {
		t.Fatalf("true should be [1], got %d elements", len(arr.Elements))
	}
}

func TestParser_BoolFalse(t *testing.T) {
	prog, errs := parse(`x = false`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	assign := prog.Statements[0].(*AssignStmt)
	arr, ok := assign.Value.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral (false -> []), got %T", assign.Value)
	}
	if len(arr.Elements) != 0 {
		t.Fatalf("false should be [], got %d elements", len(arr.Elements))
	}
}

func TestParser_Error(t *testing.T) {
	_, errs := parse(`= [1]`)
	if len(errs) == 0 {
		t.Fatal("expected parse errors, got none")
	}
}

func TestParser_BinaryAdd(t *testing.T) {
	prog, errs := parse(`x = [1] + [2]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin, ok := assign.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Op != "+" {
		t.Fatalf("expected op '+', got %q", bin.Op)
	}
}

func TestParser_Precedence_MulBeforeAdd(t *testing.T) {
	prog, errs := parse(`x = [1] + [2] * [3]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	add := assign.Value.(*BinaryExpr)
	if add.Op != "+" {
		t.Fatalf("expected outer op '+', got %q", add.Op)
	}
	mul, ok := add.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected right to be BinaryExpr, got %T", add.Right)
	}
	if mul.Op != "*" {
		t.Fatalf("expected inner op '*', got %q", mul.Op)
	}
}

func TestParser_UnaryNegate(t *testing.T) {
	prog, errs := parse(`x = -[1]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	unary, ok := assign.Value.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", assign.Value)
	}
	if unary.Op != "-" {
		t.Fatalf("expected op '-', got %q", unary.Op)
	}
}

func TestParser_UnaryNot(t *testing.T) {
	prog, errs := parse(`x = ![1]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	unary, ok := assign.Value.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", assign.Value)
	}
	if unary.Op != "!" {
		t.Fatalf("expected op '!', got %q", unary.Op)
	}
}

func TestParser_CallExpr(t *testing.T) {
	prog, errs := parse(`|> str(x)`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sw := prog.Statements[0].(*StdoutWriteStmt)
	call, ok := sw.Value.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", sw.Value)
	}
	if call.Name != "str" {
		t.Fatalf("expected name 'str', got %q", call.Name)
	}
	if len(call.Args) != 1 {
		t.Fatalf("expected 1 arg, got %d", len(call.Args))
	}
}

func TestParser_CallExprInAssign(t *testing.T) {
	prog, errs := parse(`x = str([1] + [2])`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	call, ok := assign.Value.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", assign.Value)
	}
	bin, ok := call.Args[0].(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr arg, got %T", call.Args[0])
	}
	if bin.Op != "+" {
		t.Fatalf("expected op '+', got %q", bin.Op)
	}
}

func TestParser_ParenGrouping(t *testing.T) {
	prog, errs := parse(`x = ([1] + [2]) * [3]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	mul := assign.Value.(*BinaryExpr)
	if mul.Op != "*" {
		t.Fatalf("expected outer op '*', got %q", mul.Op)
	}
	add, ok := mul.Left.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected left to be BinaryExpr, got %T", mul.Left)
	}
	if add.Op != "+" {
		t.Fatalf("expected inner op '+', got %q", add.Op)
	}
}

func TestParser_ComparisonOps(t *testing.T) {
	ops := []struct{ input, op string }{
		{`x = [1] == [2]`, "=="}, {`x = [1] != [2]`, "!="},
		{`x = [1] < [2]`, "<"}, {`x = [1] > [2]`, ">"},
		{`x = [1] <= [2]`, "<="}, {`x = [1] >= [2]`, ">="},
	}
	for _, tc := range ops {
		prog, errs := parse(tc.input)
		if len(errs) > 0 {
			t.Fatalf("input %q: unexpected errors: %v", tc.input, errs)
		}
		assign := prog.Statements[0].(*AssignStmt)
		bin := assign.Value.(*BinaryExpr)
		if bin.Op != tc.op {
			t.Fatalf("input %q: expected op %q, got %q", tc.input, tc.op, bin.Op)
		}
	}
}

func TestParser_LogicalOps(t *testing.T) {
	ops := []struct{ input, op string }{
		{`x = [1] && [2]`, "&&"}, {`x = [1] || [2]`, "||"},
	}
	for _, tc := range ops {
		prog, errs := parse(tc.input)
		if len(errs) > 0 {
			t.Fatalf("input %q: unexpected errors: %v", tc.input, errs)
		}
		assign := prog.Statements[0].(*AssignStmt)
		bin := assign.Value.(*BinaryExpr)
		if bin.Op != tc.op {
			t.Fatalf("input %q: expected op %q, got %q", tc.input, tc.op, bin.Op)
		}
	}
}

func TestParser_LogicalPrecedence(t *testing.T) {
	prog, errs := parse(`x = [1] || [2] && [3]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	or := assign.Value.(*BinaryExpr)
	if or.Op != "||" {
		t.Fatalf("expected outer op '||', got %q", or.Op)
	}
	and, ok := or.Right.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected right to be BinaryExpr, got %T", or.Right)
	}
	if and.Op != "&&" {
		t.Fatalf("expected inner op '&&', got %q", and.Op)
	}
}

func TestParser_RemoveOp(t *testing.T) {
	// With Phase 4, -- is the remove operator (binary), not double negate
	prog, errs := parse(`x = [1, 2, 3] -- [2]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin, ok := assign.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Op != "--" {
		t.Fatalf("expected op '--', got %q", bin.Op)
	}
}

func TestParser_PowerOp(t *testing.T) {
	prog, errs := parse(`x = [2] ** [3]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin := assign.Value.(*BinaryExpr)
	if bin.Op != "**" {
		t.Fatalf("expected op '**', got %q", bin.Op)
	}
}

func TestParser_StdoutWriteExpr(t *testing.T) {
	prog, errs := parse(`|> [1] + [2]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sw := prog.Statements[0].(*StdoutWriteStmt)
	bin, ok := sw.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", sw.Value)
	}
	if bin.Op != "+" {
		t.Fatalf("expected op '+', got %q", bin.Op)
	}
}

func TestParser_FnDecl(t *testing.T) {
	prog, errs := parse("fn add(a, b) { <- a + b }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn, ok := prog.Statements[0].(*FnDeclStmt)
	if !ok {
		t.Fatalf("expected FnDeclStmt, got %T", prog.Statements[0])
	}
	if fn.Name != "add" {
		t.Fatalf("expected name 'add', got %q", fn.Name)
	}
	if len(fn.Params) != 2 || fn.Params[0] != "a" || fn.Params[1] != "b" {
		t.Fatalf("expected params [a, b], got %v", fn.Params)
	}
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(fn.Body))
	}
	ret, ok := fn.Body[0].(*ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", fn.Body[0])
	}
	bin, ok := ret.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr in return, got %T", ret.Value)
	}
	if bin.Op != "+" {
		t.Fatalf("expected op '+', got %q", bin.Op)
	}
}

func TestParser_FnDeclNoParams(t *testing.T) {
	prog, errs := parse("fn foo() { |> \"hello\" }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn := prog.Statements[0].(*FnDeclStmt)
	if fn.Name != "foo" {
		t.Fatalf("expected name 'foo', got %q", fn.Name)
	}
	if len(fn.Params) != 0 {
		t.Fatalf("expected 0 params, got %d", len(fn.Params))
	}
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(fn.Body))
	}
}

func TestParser_FnDeclEmptyBody(t *testing.T) {
	prog, errs := parse("fn noop() { }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn := prog.Statements[0].(*FnDeclStmt)
	if len(fn.Body) != 0 {
		t.Fatalf("expected 0 body stmts, got %d", len(fn.Body))
	}
}

func TestParser_IfStmt(t *testing.T) {
	prog, errs := parse("if [1] { |> \"yes\" }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	ifStmt, ok := prog.Statements[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", prog.Statements[0])
	}
	if ifStmt.Condition == nil {
		t.Fatal("expected condition, got nil")
	}
	if len(ifStmt.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(ifStmt.Body))
	}
	if ifStmt.Else != nil {
		t.Fatalf("expected no else, got %d stmts", len(ifStmt.Else))
	}
}

func TestParser_IfElseStmt(t *testing.T) {
	prog, errs := parse("if [] { |> \"yes\" } else { |> \"no\" }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	ifStmt := prog.Statements[0].(*IfStmt)
	if len(ifStmt.Body) != 1 {
		t.Fatalf("expected 1 if body stmt, got %d", len(ifStmt.Body))
	}
	if len(ifStmt.Else) != 1 {
		t.Fatalf("expected 1 else stmt, got %d", len(ifStmt.Else))
	}
}

func TestParser_ForInStmt(t *testing.T) {
	prog, errs := parse("for x in items { |> str(x) }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	forStmt, ok := prog.Statements[0].(*ForInStmt)
	if !ok {
		t.Fatalf("expected ForInStmt, got %T", prog.Statements[0])
	}
	if forStmt.VarName != "x" {
		t.Fatalf("expected var 'x', got %q", forStmt.VarName)
	}
	id, ok := forStmt.Iterable.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier iterable, got %T", forStmt.Iterable)
	}
	if id.Name != "items" {
		t.Fatalf("expected 'items', got %q", id.Name)
	}
	if len(forStmt.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(forStmt.Body))
	}
}

func TestParser_ReturnStmt(t *testing.T) {
	prog, errs := parse("fn foo() { <- [1] }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn := prog.Statements[0].(*FnDeclStmt)
	ret, ok := fn.Body[0].(*ReturnStmt)
	if !ok {
		t.Fatalf("expected ReturnStmt, got %T", fn.Body[0])
	}
	if ret.Value == nil {
		t.Fatal("expected return value, got nil")
	}
}

func TestParser_BareReturn(t *testing.T) {
	prog, errs := parse("fn foo() { <- }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn := prog.Statements[0].(*FnDeclStmt)
	ret := fn.Body[0].(*ReturnStmt)
	if ret.Value != nil {
		t.Fatalf("expected nil return value, got %T", ret.Value)
	}
}

func TestParser_MultiAssign(t *testing.T) {
	prog, errs := parse("a, b = foo()")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	ma, ok := prog.Statements[0].(*MultiAssignStmt)
	if !ok {
		t.Fatalf("expected MultiAssignStmt, got %T", prog.Statements[0])
	}
	if len(ma.Names) != 2 || ma.Names[0] != "a" || ma.Names[1] != "b" {
		t.Fatalf("expected names [a, b], got %v", ma.Names)
	}
	call, ok := ma.Value.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", ma.Value)
	}
	if call.Name != "foo" {
		t.Fatalf("expected 'foo', got %q", call.Name)
	}
}

func TestParser_ExprStmt(t *testing.T) {
	prog, errs := parse("foo()")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	es, ok := prog.Statements[0].(*ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", prog.Statements[0])
	}
	call, ok := es.Expr.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", es.Expr)
	}
	if call.Name != "foo" {
		t.Fatalf("expected 'foo', got %q", call.Name)
	}
}

func TestParser_NestedBlocks(t *testing.T) {
	prog, errs := parse("fn f() { if [1] { for x in [1] { |> str(x) } } }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn := prog.Statements[0].(*FnDeclStmt)
	if len(fn.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(fn.Body))
	}
	ifStmt := fn.Body[0].(*IfStmt)
	if len(ifStmt.Body) != 1 {
		t.Fatalf("expected 1 if body stmt, got %d", len(ifStmt.Body))
	}
	forStmt := ifStmt.Body[0].(*ForInStmt)
	if len(forStmt.Body) != 1 {
		t.Fatalf("expected 1 for body stmt, got %d", len(forStmt.Body))
	}
}

func TestParser_ForLoop(t *testing.T) {
	prog, errs := parse("for { |> \"loop\" }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	forLoop, ok := prog.Statements[0].(*ForLoopStmt)
	if !ok {
		t.Fatalf("expected ForLoopStmt, got %T", prog.Statements[0])
	}
	if len(forLoop.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(forLoop.Body))
	}
}

func TestParser_BreakStmt(t *testing.T) {
	prog, errs := parse("for { break }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	forLoop := prog.Statements[0].(*ForLoopStmt)
	bs, ok := forLoop.Body[0].(*BreakStmt)
	if !ok {
		t.Fatalf("expected BreakStmt, got %T", forLoop.Body[0])
	}
	_ = bs
}

func TestParser_ForLoopWithBreak(t *testing.T) {
	prog, errs := parse(`for { if [1] { break } |> "hi" }`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	forLoop := prog.Statements[0].(*ForLoopStmt)
	if len(forLoop.Body) != 2 {
		t.Fatalf("expected 2 body stmts, got %d", len(forLoop.Body))
	}
	ifStmt, ok := forLoop.Body[0].(*IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", forLoop.Body[0])
	}
	_, ok = ifStmt.Body[0].(*BreakStmt)
	if !ok {
		t.Fatalf("expected BreakStmt inside if, got %T", ifStmt.Body[0])
	}
}

// ==========================================
// Negative / Edge Case / Boundary Tests
// ==========================================

func TestParser_Error_MissingCloseBracket(t *testing.T) {
	_, errs := parse(`x = [1, 2`)
	if len(errs) == 0 {
		t.Fatal("expected parse error for unclosed bracket")
	}
}

func TestParser_Error_MissingCloseParen(t *testing.T) {
	_, errs := parse(`x = str(y`)
	if len(errs) == 0 {
		t.Fatal("expected parse error for unclosed paren")
	}
}

func TestParser_Error_MissingFnName(t *testing.T) {
	_, errs := parse(`fn (a, b) { <- a }`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: fn missing name")
	}
}

func TestParser_Error_MissingFnOpenParen(t *testing.T) {
	_, errs := parse(`fn foo a, b) { <- a }`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: fn missing (")
	}
}

func TestParser_Error_MissingFnCloseParen(t *testing.T) {
	_, errs := parse(`fn foo(a, b { <- a }`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: fn missing )")
	}
}

func TestParser_Error_MissingIfBrace(t *testing.T) {
	_, errs := parse(`if [1] |> "yes" }`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: if missing {")
	}
}

func TestParser_Error_MissingForInKeyword(t *testing.T) {
	_, errs := parse(`for x items { |> str(x) }`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: for missing 'in'")
	}
}

func TestParser_Error_ForInMissingVar(t *testing.T) {
	_, errs := parse(`for [1] in items { }`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: for-in non-ident but not {")
	}
}

func TestParser_Error_MultiAssignMissingEquals(t *testing.T) {
	_, errs := parse(`a, b foo()`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: multi-assign missing =")
	}
}

func TestParser_Error_UnexpectedToken(t *testing.T) {
	_, errs := parse(`] [1]`)
	if len(errs) == 0 {
		t.Fatal("expected parse error: unexpected ]")
	}
}

func TestParser_Error_AssignMissingValue(t *testing.T) {
	_, errs := parse(`x = `)
	if len(errs) == 0 {
		t.Fatal("expected parse error: assign missing value")
	}
}

func TestParser_BreakOutsideLoop(t *testing.T) {
	// Parser doesn't enforce break context — it parses fine.
	// Go compiler catches break outside loop.
	prog, errs := parse("break")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	_, ok := prog.Statements[0].(*BreakStmt)
	if !ok {
		t.Fatalf("expected BreakStmt, got %T", prog.Statements[0])
	}
}

func TestParser_EmptyForLoop(t *testing.T) {
	prog, errs := parse("for { }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fl := prog.Statements[0].(*ForLoopStmt)
	if len(fl.Body) != 0 {
		t.Fatalf("expected 0 body stmts, got %d", len(fl.Body))
	}
}

func TestParser_NestedForLoops(t *testing.T) {
	prog, errs := parse("for { for { break } }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	outer := prog.Statements[0].(*ForLoopStmt)
	inner, ok := outer.Body[0].(*ForLoopStmt)
	if !ok {
		t.Fatalf("expected nested ForLoopStmt, got %T", outer.Body[0])
	}
	_, ok = inner.Body[0].(*BreakStmt)
	if !ok {
		t.Fatalf("expected BreakStmt in inner loop, got %T", inner.Body[0])
	}
}

func TestParser_ForLoopInsideFn(t *testing.T) {
	prog, errs := parse("fn f() { for { break } }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	fn := prog.Statements[0].(*FnDeclStmt)
	fl, ok := fn.Body[0].(*ForLoopStmt)
	if !ok {
		t.Fatalf("expected ForLoopStmt, got %T", fn.Body[0])
	}
	if len(fl.Body) != 1 {
		t.Fatalf("expected 1 body stmt, got %d", len(fl.Body))
	}
}

// ==========================================
// Phase 4: Array Operator Tests
// ==========================================

func TestParser_IndexExpr(t *testing.T) {
	prog, errs := parse(`|> arr@0`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sw := prog.Statements[0].(*StdoutWriteStmt)
	idx, ok := sw.Value.(*IndexExpr)
	if !ok {
		t.Fatalf("expected IndexExpr, got %T", sw.Value)
	}
	obj, ok := idx.Object.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier object, got %T", idx.Object)
	}
	if obj.Name != "arr" {
		t.Fatalf("expected 'arr', got %q", obj.Name)
	}
	num, ok := idx.Index.(*NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral index, got %T", idx.Index)
	}
	if num.Value != "0" {
		t.Fatalf("expected '0', got %q", num.Value)
	}
}

func TestParser_IndexSetStmt(t *testing.T) {
	prog, errs := parse(`arr@0 = [99]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	is, ok := prog.Statements[0].(*IndexSetStmt)
	if !ok {
		t.Fatalf("expected IndexSetStmt, got %T", prog.Statements[0])
	}
	obj := is.Object.(*Identifier)
	if obj.Name != "arr" {
		t.Fatalf("expected 'arr', got %q", obj.Name)
	}
	num := is.Index.(*NumberLiteral)
	if num.Value != "0" {
		t.Fatalf("expected '0', got %q", num.Value)
	}
	arr, ok := is.Value.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral value, got %T", is.Value)
	}
	if len(arr.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr.Elements))
	}
}

func TestParser_HashLength(t *testing.T) {
	prog, errs := parse(`x = #arr`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	unary, ok := assign.Value.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", assign.Value)
	}
	if unary.Op != "#" {
		t.Fatalf("expected op '#', got %q", unary.Op)
	}
}

func TestParser_PushStmt(t *testing.T) {
	prog, errs := parse(`arr << [5]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	ps, ok := prog.Statements[0].(*PushStmt)
	if !ok {
		t.Fatalf("expected PushStmt, got %T", prog.Statements[0])
	}
	obj := ps.Object.(*Identifier)
	if obj.Name != "arr" {
		t.Fatalf("expected 'arr', got %q", obj.Name)
	}
}

func TestParser_PopExpr(t *testing.T) {
	prog, errs := parse(`x = >>arr`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	pop, ok := assign.Value.(*PopExpr)
	if !ok {
		t.Fatalf("expected PopExpr, got %T", assign.Value)
	}
	obj := pop.Object.(*Identifier)
	if obj.Name != "arr" {
		t.Fatalf("expected 'arr', got %q", obj.Name)
	}
}

func TestParser_SliceExpr(t *testing.T) {
	prog, errs := parse(`x = arr::1::3`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	slice, ok := assign.Value.(*SliceExpr)
	if !ok {
		t.Fatalf("expected SliceExpr, got %T", assign.Value)
	}
	obj := slice.Object.(*Identifier)
	if obj.Name != "arr" {
		t.Fatalf("expected 'arr', got %q", obj.Name)
	}
	low := slice.Low.(*NumberLiteral)
	if low.Value != "1" {
		t.Fatalf("expected low '1', got %q", low.Value)
	}
	high := slice.High.(*NumberLiteral)
	if high.Value != "3" {
		t.Fatalf("expected high '3', got %q", high.Value)
	}
}

func TestParser_ConcatExpr(t *testing.T) {
	prog, errs := parse(`x = [1, 2] ++ [3, 4]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin, ok := assign.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Op != "++" {
		t.Fatalf("expected op '++', got %q", bin.Op)
	}
}

func TestParser_RemoveBinaryExpr(t *testing.T) {
	prog, errs := parse(`x = arr -- [5]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin, ok := assign.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Op != "--" {
		t.Fatalf("expected op '--', got %q", bin.Op)
	}
}

func TestParser_UniqueExpr(t *testing.T) {
	prog, errs := parse(`x = ~arr`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	unary, ok := assign.Value.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", assign.Value)
	}
	if unary.Op != "~" {
		t.Fatalf("expected op '~', got %q", unary.Op)
	}
}

func TestParser_ContainsExpr(t *testing.T) {
	prog, errs := parse(`x = arr ?? [5]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin, ok := assign.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Op != "??" {
		t.Fatalf("expected op '??', got %q", bin.Op)
	}
}

func TestParser_RemoveAtExpr(t *testing.T) {
	prog, errs := parse(`x = arr ~@ [2]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	bin, ok := assign.Value.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", assign.Value)
	}
	if bin.Op != "~@" {
		t.Fatalf("expected op '~@', got %q", bin.Op)
	}
}

func TestParser_IfWithComparison(t *testing.T) {
	prog, errs := parse("if [1] == [1] { |> \"equal\" }")
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	ifStmt := prog.Statements[0].(*IfStmt)
	bin, ok := ifStmt.Condition.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr condition, got %T", ifStmt.Condition)
	}
	if bin.Op != "==" {
		t.Fatalf("expected op '==', got %q", bin.Op)
	}
}
