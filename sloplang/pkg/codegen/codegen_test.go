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
