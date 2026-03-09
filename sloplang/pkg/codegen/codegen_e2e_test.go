package codegen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

func projectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	// codegen_e2e_test.go is in pkg/codegen/, go up 3 levels
	return filepath.Dir(filepath.Dir(filepath.Dir(filename)))
}

// runE2E transpiles source, compiles, runs, and returns stdout.
func runE2E(t *testing.T, source string) string {
	t.Helper()

	l := lexer.New(source)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}

	gen := New(modulePath)
	output, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	tmpDir := t.TempDir()
	root := projectRoot()

	goMod := fmt.Sprintf(`module test

go 1.24

require github.com/saad039/sloplang v0.0.0

replace github.com/saad039/sloplang => %s
`, root)

	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), output, 0644); err != nil {
		t.Fatalf("write main.go: %v", err)
	}

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if tidyOut, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, string(tidyOut))
	}

	buildCmd := exec.Command("go", "build", "-o", "prog", ".")
	buildCmd.Dir = tmpDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s\n\nGenerated code:\n%s", err, string(buildOut), string(output))
	}

	runCmd := exec.Command(filepath.Join(tmpDir, "prog"))
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(runOut))
	}

	return strings.TrimRight(string(runOut), "\n")
}

func TestE2E_HelloWorld(t *testing.T) {
	got := runE2E(t, `x = [1, 2, 3]
|> "hello world"
|> x`)
	expected := "hello world\n[1, 2, 3]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestE2E_UnusedVariable(t *testing.T) {
	// Variable assigned but never referenced — must still compile
	got := runE2E(t, `x = [1, 2, 3]
|> "ok"`)
	if got != "ok" {
		t.Fatalf("expected %q, got %q", "ok", got)
	}
}

func TestE2E_MultipleUnusedVariables(t *testing.T) {
	got := runE2E(t, `a = [1]
b = [2]
c = [3]
|> "done"`)
	if got != "done" {
		t.Fatalf("expected %q, got %q", "done", got)
	}
}

func TestE2E_UsedAndUnusedMixed(t *testing.T) {
	// a is unused, b is used — both must compile
	got := runE2E(t, `a = [99]
b = [42]
|> b`)
	if got != "42" {
		t.Fatalf("expected %q, got %q", "42", got)
	}
}

func TestE2E_ReassignVariable(t *testing.T) {
	// In Phase 1 each assignment is :=, so two assignments to same name
	// would fail. This test documents current behavior with distinct names.
	got := runE2E(t, `x = [1]
y = [2]
|> x
|> y`)
	if got != "1\n2" {
		t.Fatalf("expected %q, got %q", "1\n2", got)
	}
}

func TestE2E_EmptyArray(t *testing.T) {
	got := runE2E(t, `x = []
|> x`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_MixedTypes(t *testing.T) {
	got := runE2E(t, `x = [1, 3.14, "hi"]
|> x`)
	expected := "[1, 3.14, hi]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestE2E_StringOnly(t *testing.T) {
	got := runE2E(t, `|> "hello"
|> "world"`)
	if got != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", got)
	}
}

func TestE2E_UintLiteral(t *testing.T) {
	got := runE2E(t, `x = [42u]
|> x`)
	if got != "42" {
		t.Fatalf("expected %q, got %q", "42", got)
	}
}
