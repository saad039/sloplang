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
