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

func TestParser_DoubleUnaryNegate(t *testing.T) {
	prog, errs := parse(`x = --[5]`)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	assign := prog.Statements[0].(*AssignStmt)
	outer := assign.Value.(*UnaryExpr)
	if outer.Op != "-" {
		t.Fatalf("expected op '-', got %q", outer.Op)
	}
	inner, ok := outer.Operand.(*UnaryExpr)
	if !ok {
		t.Fatalf("expected inner UnaryExpr, got %T", outer.Operand)
	}
	if inner.Op != "-" {
		t.Fatalf("expected inner op '-', got %q", inner.Op)
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
