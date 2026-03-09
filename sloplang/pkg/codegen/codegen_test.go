package codegen

import (
	"strings"
	"testing"

	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

const modulePath = "github.com/saad039/sloplang"

func generate(input string) (string, error) {
	l := lexer.New(input)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, errs := p.Parse()
	if len(errs) > 0 {
		return "", nil
	}
	gen := New(modulePath)
	output, err := gen.Generate(prog)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func TestCodegen_Assignment(t *testing.T) {
	out, err := generate(`x = [1, 2, 3]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "x :=") {
		t.Fatalf("expected ':=' assignment, got:\n%s", out)
	}
	if !strings.Contains(out, "sloprt.NewSlopValue") {
		t.Fatalf("expected sloprt.NewSlopValue call, got:\n%s", out)
	}
	if !strings.Contains(out, "int64(1)") {
		t.Fatalf("expected int64(1), got:\n%s", out)
	}
}

func TestCodegen_StdoutWrite(t *testing.T) {
	out, err := generate(`|> "hello world"`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.StdoutWrite") {
		t.Fatalf("expected sloprt.StdoutWrite, got:\n%s", out)
	}
	if !strings.Contains(out, `"hello world"`) {
		t.Fatalf("expected string literal, got:\n%s", out)
	}
}

func TestCodegen_FullProgram(t *testing.T) {
	out, err := generate("x = [1, 2, 3]\n|> \"hello world\"")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "package main") {
		t.Fatalf("expected 'package main', got:\n%s", out)
	}
	if !strings.Contains(out, `sloprt "github.com/saad039/sloplang/pkg/runtime"`) {
		t.Fatalf("expected runtime import, got:\n%s", out)
	}
	if !strings.Contains(out, "func main()") {
		t.Fatalf("expected func main(), got:\n%s", out)
	}
}

func TestCodegen_EmptyArray(t *testing.T) {
	out, err := generate(`x = []`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.NewSlopValue()") {
		t.Fatalf("expected sloprt.NewSlopValue() with no args, got:\n%s", out)
	}
}

func TestCodegen_NumberTypes(t *testing.T) {
	out, err := generate(`x = [42u]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "uint64(42)") {
		t.Fatalf("expected uint64(42), got:\n%s", out)
	}

	out, err = generate(`x = [3.14]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "float64(3.14)") {
		t.Fatalf("expected float64(3.14), got:\n%s", out)
	}
}

func TestCodegen_StdoutWriteIdent(t *testing.T) {
	out, err := generate("x = [1]\n|> x")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.StdoutWrite(x)") {
		t.Fatalf("expected sloprt.StdoutWrite(x), got:\n%s", out)
	}
}

func TestCodegen_BinaryAllOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x = [1] + [2]`, "sloprt.Add("},
		{`x = [1] - [2]`, "sloprt.Sub("},
		{`x = [1] * [2]`, "sloprt.Mul("},
		{`x = [1] / [2]`, "sloprt.Div("},
		{`x = [1] % [2]`, "sloprt.Mod("},
		{`x = [1] ** [2]`, "sloprt.Pow("},
		{`x = [1] == [2]`, "sloprt.Eq("},
		{`x = [1] != [2]`, "sloprt.Neq("},
		{`x = [1] < [2]`, "sloprt.Lt("},
		{`x = [1] > [2]`, "sloprt.Gt("},
		{`x = [1] <= [2]`, "sloprt.Lte("},
		{`x = [1] >= [2]`, "sloprt.Gte("},
		{`x = [1] && [2]`, "sloprt.And("},
		{`x = [1] || [2]`, "sloprt.Or("},
	}
	for _, tc := range tests {
		out, err := generate(tc.input)
		if err != nil {
			t.Fatalf("input %q: codegen error: %v", tc.input, err)
		}
		if !strings.Contains(out, tc.expected) {
			t.Fatalf("input %q: expected %q, got:\n%s", tc.input, tc.expected, out)
		}
	}
}

func TestCodegen_UnaryNegate(t *testing.T) {
	out, err := generate(`x = -[1]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Negate(") {
		t.Fatalf("expected sloprt.Negate, got:\n%s", out)
	}
}

func TestCodegen_UnaryNot(t *testing.T) {
	out, err := generate(`x = ![1]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Not(") {
		t.Fatalf("expected sloprt.Not, got:\n%s", out)
	}
}

func TestCodegen_CallStr(t *testing.T) {
	out, err := generate(`|> str([1])`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Str(") {
		t.Fatalf("expected sloprt.Str, got:\n%s", out)
	}
}
