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

// === Phase 2 E2E Tests ===

// Arithmetic (12)

func TestE2E_Add(t *testing.T) {
	got := runE2E(t, "x = [10, 20] + [3, 4]\n|> str(x)")
	if got != "[13, 24]" {
		t.Fatalf("expected %q, got %q", "[13, 24]", got)
	}
}

func TestE2E_Sub(t *testing.T) {
	got := runE2E(t, "x = [5, 3] - [1, 1]\n|> str(x)")
	if got != "[4, 2]" {
		t.Fatalf("expected %q, got %q", "[4, 2]", got)
	}
}

func TestE2E_Mul(t *testing.T) {
	got := runE2E(t, "x = [2, 3] * [4, 5]\n|> str(x)")
	if got != "[8, 15]" {
		t.Fatalf("expected %q, got %q", "[8, 15]", got)
	}
}

func TestE2E_Div(t *testing.T) {
	got := runE2E(t, "x = [10, 6] / [2, 3]\n|> str(x)")
	if got != "[5, 2]" {
		t.Fatalf("expected %q, got %q", "[5, 2]", got)
	}
}

func TestE2E_Mod(t *testing.T) {
	got := runE2E(t, "x = [7, 5] % [3, 2]\n|> str(x)")
	if got != "[1, 1]" {
		t.Fatalf("expected %q, got %q", "[1, 1]", got)
	}
}

func TestE2E_Pow(t *testing.T) {
	got := runE2E(t, "x = [2, 3] ** [3, 2]\n|> str(x)")
	if got != "[8, 9]" {
		t.Fatalf("expected %q, got %q", "[8, 9]", got)
	}
}

func TestE2E_AddSingleElement(t *testing.T) {
	got := runE2E(t, "x = [1] + [1]\n|> str(x)")
	if got != "2" {
		t.Fatalf("expected %q, got %q", "2", got)
	}
}

func TestE2E_SubZero(t *testing.T) {
	got := runE2E(t, "x = [0] - [0]\n|> str(x)")
	if got != "0" {
		t.Fatalf("expected %q, got %q", "0", got)
	}
}

func TestE2E_MulByZero(t *testing.T) {
	got := runE2E(t, "x = [1] * [0]\n|> str(x)")
	if got != "0" {
		t.Fatalf("expected %q, got %q", "0", got)
	}
}

func TestE2E_AddLongArrays(t *testing.T) {
	got := runE2E(t, "x = [100, 200, 300, 400, 500] + [1, 2, 3, 4, 5]\n|> str(x)")
	if got != "[101, 202, 303, 404, 505]" {
		t.Fatalf("expected %q, got %q", "[101, 202, 303, 404, 505]", got)
	}
}

func TestE2E_AddFloat(t *testing.T) {
	got := runE2E(t, "x = [3.14] + [2.86]\n|> str(x)")
	if got != "6" {
		t.Fatalf("expected %q, got %q", "6", got)
	}
}

func TestE2E_AddUint(t *testing.T) {
	got := runE2E(t, "x = [42u] + [8u]\n|> str(x)")
	if got != "50" {
		t.Fatalf("expected %q, got %q", "50", got)
	}
}

// Unary (6)

func TestE2E_Negate(t *testing.T) {
	got := runE2E(t, "z = -[1, 2, 3]\n|> str(z)")
	if got != "[-1, -2, -3]" {
		t.Fatalf("expected %q, got %q", "[-1, -2, -3]", got)
	}
}

func TestE2E_NegateZero(t *testing.T) {
	got := runE2E(t, "x = -[0]\n|> str(x)")
	if got != "0" {
		t.Fatalf("expected %q, got %q", "0", got)
	}
}

func TestE2E_NegateNegative(t *testing.T) {
	got := runE2E(t, "x = -[-5]\n|> str(x)")
	if got != "5" {
		t.Fatalf("expected %q, got %q", "5", got)
	}
}

func TestE2E_DoubleNegate(t *testing.T) {
	got := runE2E(t, "x = --[5]\n|> str(x)")
	if got != "5" {
		t.Fatalf("expected %q, got %q", "5", got)
	}
}

func TestE2E_NegateFloat(t *testing.T) {
	got := runE2E(t, "x = -[3.14]\n|> str(x)")
	if got != "-3.14" {
		t.Fatalf("expected %q, got %q", "-3.14", got)
	}
}

func TestE2E_NotFalsy(t *testing.T) {
	got := runE2E(t, "x = ![]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

// Not (4)

func TestE2E_NotTruthy(t *testing.T) {
	got := runE2E(t, "x = ![1]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_NotZeroIsTruthy(t *testing.T) {
	got := runE2E(t, "x = ![0]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_DoubleNotTruthy(t *testing.T) {
	got := runE2E(t, "x = !![1]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_DoubleNotFalsy(t *testing.T) {
	got := runE2E(t, "x = !![]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// Comparisons (14)

func TestE2E_EqTrue(t *testing.T) {
	got := runE2E(t, "x = [2] == [2]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_EqFalse(t *testing.T) {
	got := runE2E(t, "x = [2] == [3]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_NeqTrue(t *testing.T) {
	got := runE2E(t, "x = [2] != [3]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_NeqFalse(t *testing.T) {
	got := runE2E(t, "x = [2] != [2]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_LtTrue(t *testing.T) {
	got := runE2E(t, "x = [1] < [2]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_LtFalse(t *testing.T) {
	got := runE2E(t, "x = [2] < [1]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_GtTrue(t *testing.T) {
	got := runE2E(t, "x = [2] > [1]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_GtFalse(t *testing.T) {
	got := runE2E(t, "x = [1] > [2]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_LteEqual(t *testing.T) {
	got := runE2E(t, "x = [1] <= [1]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_LteLess(t *testing.T) {
	got := runE2E(t, "x = [1] <= [2]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_GteEqual(t *testing.T) {
	got := runE2E(t, "x = [2] >= [2]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_GteFail(t *testing.T) {
	got := runE2E(t, "x = [1] >= [2]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_StringEq(t *testing.T) {
	got := runE2E(t, "x = \"abc\" == \"abc\"\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_StringNeq(t *testing.T) {
	got := runE2E(t, "x = \"abc\" != \"def\"\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

// Logical (6)

func TestE2E_AndBothTruthy(t *testing.T) {
	got := runE2E(t, "x = [1] && [1]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_AndRightFalsy(t *testing.T) {
	got := runE2E(t, "x = [1] && []\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_AndLeftFalsy(t *testing.T) {
	got := runE2E(t, "x = [] && [1]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_OrLeftFalsy(t *testing.T) {
	got := runE2E(t, "x = [] || [1]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_OrLeftTruthy(t *testing.T) {
	got := runE2E(t, "x = [1] || []\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_OrBothFalsy(t *testing.T) {
	got := runE2E(t, "x = [] || []\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// Precedence (8)

func TestE2E_PrecedenceMulBeforeAdd(t *testing.T) {
	got := runE2E(t, "x = [2] * [3] + [1]\n|> str(x)")
	if got != "7" {
		t.Fatalf("expected %q, got %q", "7", got)
	}
}

func TestE2E_PrecedenceMulBeforeAddRight(t *testing.T) {
	got := runE2E(t, "x = [1] + [2] * [3]\n|> str(x)")
	if got != "7" {
		t.Fatalf("expected %q, got %q", "7", got)
	}
}

func TestE2E_PrecedenceParens(t *testing.T) {
	got := runE2E(t, "x = ([1] + [2]) * [3]\n|> str(x)")
	if got != "9" {
		t.Fatalf("expected %q, got %q", "9", got)
	}
}

func TestE2E_PrecedenceLeftAssocSub(t *testing.T) {
	got := runE2E(t, "x = [10] - [2] - [3]\n|> str(x)")
	if got != "5" {
		t.Fatalf("expected %q, got %q", "5", got)
	}
}

func TestE2E_PrecedencePower(t *testing.T) {
	got := runE2E(t, "x = [2] ** [3]\n|> str(x)")
	if got != "8" {
		t.Fatalf("expected %q, got %q", "8", got)
	}
}

func TestE2E_PrecedenceArithBeforeComparison(t *testing.T) {
	got := runE2E(t, "x = [1] + [2] > [2]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_PrecedenceComparisonBeforeLogical(t *testing.T) {
	got := runE2E(t, "x = [1] > [0] && [2] > [1]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_PrecedenceAndBeforeOr(t *testing.T) {
	got := runE2E(t, "x = [1] > [0] || [] && []\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

// str() builtin (4)

func TestE2E_StrArray(t *testing.T) {
	got := runE2E(t, "|> str([1, 2, 3])")
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestE2E_StrSingle(t *testing.T) {
	got := runE2E(t, "|> str([42])")
	if got != "42" {
		t.Fatalf("expected %q, got %q", "42", got)
	}
}

func TestE2E_StrEmpty(t *testing.T) {
	got := runE2E(t, "|> str([])")
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_StrExprResult(t *testing.T) {
	got := runE2E(t, "|> str([1] + [2])")
	if got != "3" {
		t.Fatalf("expected %q, got %q", "3", got)
	}
}

// Chained (4)

func TestE2E_ChainedAdd(t *testing.T) {
	got := runE2E(t, "x = [1] + [2] + [3] + [4]\n|> str(x)")
	if got != "10" {
		t.Fatalf("expected %q, got %q", "10", got)
	}
}

func TestE2E_ChainedMulDiv(t *testing.T) {
	got := runE2E(t, "x = [10] / [2] * [3]\n|> str(x)")
	if got != "15" {
		t.Fatalf("expected %q, got %q", "15", got)
	}
}

func TestE2E_ChainedLogical(t *testing.T) {
	got := runE2E(t, "x = [1] == [1] && [2] == [2] && [3] == [3]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

func TestE2E_UnaryThenBinary(t *testing.T) {
	got := runE2E(t, "x = -[1] + [2]\n|> str(x)")
	if got != "1" {
		t.Fatalf("expected %q, got %q", "1", got)
	}
}

// Roadmap e2e — math.slop

func TestE2E_MathSlop(t *testing.T) {
	got := runE2E(t, "x = [10, 20] + [3, 4]\n|> str(x)\ny = [5] > [3]\n|> str(y)\nz = -[1, 2, 3]\n|> str(z)")
	expected := "[13, 24]\n1\n[-1, -2, -3]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
