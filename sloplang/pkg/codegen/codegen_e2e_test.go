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

// ==========================================
// Phase 3: Functions + Control Flow e2e tests
// ==========================================

// --- Functions: basic ---

func TestE2E_FnSingleParam(t *testing.T) {
	src := `fn double(x) {
	<- x + x
}
result = double([5])
|> str(result)`
	got := runE2E(t, src)
	if got != "10" {
		t.Fatalf("expected %q, got %q", "10", got)
	}
}

func TestE2E_FnTwoParams(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
result = add([3], [4])
|> str(result)`
	got := runE2E(t, src)
	if got != "7" {
		t.Fatalf("expected %q, got %q", "7", got)
	}
}

func TestE2E_FnNoParams(t *testing.T) {
	src := `fn greeting() {
	<- "hello"
}
|> greeting()`
	got := runE2E(t, src)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_FnCallingFn(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
fn add3(a, b, c) {
	<- add(add(a, b), c)
}
|> str(add3([1], [2], [3]))`
	got := runE2E(t, src)
	if got != "6" {
		t.Fatalf("expected %q, got %q", "6", got)
	}
}

func TestE2E_FnRecursiveCountdown(t *testing.T) {
	src := `fn countdown(n) {
	if n == [0] {
		|> "done"
		<- []
	}
	|> str(n)
	<- countdown(n - [1])
}
countdown([3])`
	got := runE2E(t, src)
	if got != "3\n2\n1\ndone" {
		t.Fatalf("expected %q, got %q", "3\n2\n1\ndone", got)
	}
}

func TestE2E_FnWithStrInBody(t *testing.T) {
	src := `fn show(x) {
	|> str(x)
	<- []
}
show([42])`
	got := runE2E(t, src)
	if got != "42" {
		t.Fatalf("expected %q, got %q", "42", got)
	}
}

func TestE2E_FnExprAsArg(t *testing.T) {
	src := `fn double(x) {
	<- x + x
}
|> str(double([1] + [2]))`
	got := runE2E(t, src)
	if got != "6" {
		t.Fatalf("expected %q, got %q", "6", got)
	}
}

func TestE2E_FnResultInExpr(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
|> str(add([1], [2]) + [3])`
	got := runE2E(t, src)
	if got != "6" {
		t.Fatalf("expected %q, got %q", "6", got)
	}
}

// --- Return ---

func TestE2E_ReturnValue(t *testing.T) {
	src := `fn five() {
	<- [5]
}
|> str(five())`
	got := runE2E(t, src)
	if got != "5" {
		t.Fatalf("expected %q, got %q", "5", got)
	}
}

func TestE2E_BareReturn(t *testing.T) {
	src := `fn empty() {
	<-
}
|> str(empty())`
	got := runE2E(t, src)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_EarlyReturn(t *testing.T) {
	src := `fn check(x) {
	if x == [0] {
		<- "zero"
	}
	<- "nonzero"
}
|> check([0])
|> check([5])`
	got := runE2E(t, src)
	if got != "zero\nnonzero" {
		t.Fatalf("expected %q, got %q", "zero\nnonzero", got)
	}
}

func TestE2E_ReturnFromNestedIf(t *testing.T) {
	src := `fn classify(x) {
	if x > [0] {
		<- "positive"
	}
	if x == [0] {
		<- "zero"
	}
	<- "negative"
}
|> classify([5])
|> classify([0])
|> classify(-[1])`
	got := runE2E(t, src)
	if got != "positive\nzero\nnegative" {
		t.Fatalf("expected %q, got %q", "positive\nzero\nnegative", got)
	}
}

func TestE2E_ReturnExpr(t *testing.T) {
	src := `fn sum(a, b) {
	<- a + b
}
|> str(sum([10], [20]))`
	got := runE2E(t, src)
	if got != "30" {
		t.Fatalf("expected %q, got %q", "30", got)
	}
}

// --- If/Else ---

func TestE2E_IfTruthy(t *testing.T) {
	got := runE2E(t, "if [1] { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_IfFalsy(t *testing.T) {
	got := runE2E(t, "if [] { |> \"yes\" }\n|> \"done\"")
	if got != "done" {
		t.Fatalf("expected %q, got %q", "done", got)
	}
}

func TestE2E_IfElseTruthy(t *testing.T) {
	got := runE2E(t, "if [1] { |> \"yes\" } else { |> \"no\" }")
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_IfElseFalsy(t *testing.T) {
	got := runE2E(t, "if [] { |> \"yes\" } else { |> \"no\" }")
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestE2E_IfComparison(t *testing.T) {
	got := runE2E(t, "if [1] == [1] { |> \"equal\" }")
	if got != "equal" {
		t.Fatalf("expected %q, got %q", "equal", got)
	}
}

func TestE2E_IfLogical(t *testing.T) {
	got := runE2E(t, "if [1] && [1] { |> \"both\" }")
	if got != "both" {
		t.Fatalf("expected %q, got %q", "both", got)
	}
}

func TestE2E_NestedIf(t *testing.T) {
	src := `if [1] {
	if [2] == [2] {
		|> "nested"
	}
}`
	got := runE2E(t, src)
	if got != "nested" {
		t.Fatalf("expected %q, got %q", "nested", got)
	}
}

func TestE2E_IfInsideFn(t *testing.T) {
	src := `fn abs(x) {
	if x < [0] {
		<- -x
	}
	<- x
}
|> str(abs(-[5]))
|> str(abs([3]))`
	got := runE2E(t, src)
	if got != "5\n3" {
		t.Fatalf("expected %q, got %q", "5\n3", got)
	}
}

func TestE2E_IfNotOperator(t *testing.T) {
	// [0] is truthy, !truthy = falsy → else
	got := runE2E(t, "if ![0] { |> \"not\" } else { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_IfElseMultiStmt(t *testing.T) {
	src := `if [1] == [2] {
	|> "a"
	|> "b"
} else {
	|> "c"
	|> "d"
}`
	got := runE2E(t, src)
	if got != "c\nd" {
		t.Fatalf("expected %q, got %q", "c\nd", got)
	}
}

// --- For/In ---

func TestE2E_ForInArray(t *testing.T) {
	src := `items = [1, 2, 3]
for item in items {
	|> str(item)
}`
	got := runE2E(t, src)
	if got != "1\n2\n3" {
		t.Fatalf("expected %q, got %q", "1\n2\n3", got)
	}
}

func TestE2E_ForInEmpty(t *testing.T) {
	src := `items = []
for item in items {
	|> str(item)
}
|> "done"`
	got := runE2E(t, src)
	if got != "done" {
		t.Fatalf("expected %q, got %q", "done", got)
	}
}

func TestE2E_ForInSingleElement(t *testing.T) {
	src := `for item in [42] {
	|> str(item)
}`
	got := runE2E(t, src)
	if got != "42" {
		t.Fatalf("expected %q, got %q", "42", got)
	}
}

func TestE2E_ForInComputation(t *testing.T) {
	src := `for item in [1, 2, 3] {
	|> str(item + [10])
}`
	got := runE2E(t, src)
	if got != "11\n12\n13" {
		t.Fatalf("expected %q, got %q", "11\n12\n13", got)
	}
}

func TestE2E_NestedForLoops(t *testing.T) {
	src := `for i in [1, 2] {
	for j in [10, 20] {
		|> str(i + j)
	}
}`
	got := runE2E(t, src)
	if got != "11\n21\n12\n22" {
		t.Fatalf("expected %q, got %q", "11\n21\n12\n22", got)
	}
}

func TestE2E_ForInsideFn(t *testing.T) {
	src := `fn printAll(items) {
	for item in items {
		|> str(item)
	}
	<- []
}
printAll([10, 20, 30])`
	got := runE2E(t, src)
	if got != "10\n20\n30" {
		t.Fatalf("expected %q, got %q", "10\n20\n30", got)
	}
}

func TestE2E_ForWithIfInside(t *testing.T) {
	src := `for x in [1, 2, 3, 4] {
	if x > [2] {
		|> str(x)
	}
}`
	got := runE2E(t, src)
	if got != "3\n4" {
		t.Fatalf("expected %q, got %q", "3\n4", got)
	}
}

func TestE2E_ForInStringArray(t *testing.T) {
	src := `for s in ["a", "b", "c"] {
	|> s
}`
	got := runE2E(t, src)
	if got != "a\nb\nc" {
		t.Fatalf("expected %q, got %q", "a\nb\nc", got)
	}
}

// --- Multi-assign ---

func TestE2E_MultiAssignFromFn(t *testing.T) {
	src := `fn pair() {
	<- [10, 20]
}
a, b = pair()
|> str(a)
|> str(b)`
	got := runE2E(t, src)
	if got != "10\n20" {
		t.Fatalf("expected %q, got %q", "10\n20", got)
	}
}

func TestE2E_MultiAssignUse(t *testing.T) {
	src := `fn divmod(a, b) {
	<- [a / b, a % b]
}
q, r = divmod([7], [3])
|> str(q)
|> str(r)`
	got := runE2E(t, src)
	if got != "2\n1" {
		t.Fatalf("expected %q, got %q", "2\n1", got)
	}
}

func TestE2E_MultiAssignNested(t *testing.T) {
	src := `fn pair() {
	<- [[1, 2], [3, 4]]
}
a, b = pair()
|> str(a)
|> str(b)`
	got := runE2E(t, src)
	if got != "[1, 2]\n[3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2]\n[3, 4]", got)
	}
}

func TestE2E_MultiAssignDirect(t *testing.T) {
	src := `a, b = [10, 20]
|> str(a)
|> str(b)`
	got := runE2E(t, src)
	if got != "10\n20" {
		t.Fatalf("expected %q, got %q", "10\n20", got)
	}
}

// --- Combined / Integration ---

func TestE2E_Fibonacci(t *testing.T) {
	src := `fn fib(n) {
	if n <= [1] {
		<- n
	}
	<- fib(n - [1]) + fib(n - [1] - [1])
}
for i in [0, 1, 2, 3, 4, 5, 6, 7, 8, 9] {
	|> str(fib(i))
}`
	got := runE2E(t, src)
	expected := "0\n1\n1\n2\n3\n5\n8\n13\n21\n34"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestE2E_Factorial(t *testing.T) {
	src := `fn fact(n) {
	if n <= [1] {
		<- [1]
	}
	<- n * fact(n - [1])
}
|> str(fact([5]))
|> str(fact([1]))
|> str(fact([0]))`
	got := runE2E(t, src)
	if got != "120\n1\n1" {
		t.Fatalf("expected %q, got %q", "120\n1\n1", got)
	}
}

func TestE2E_FnAsArgToFn(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
fn mul(a, b) {
	<- a * b
}
|> str(add(mul([2], [3]), [4]))`
	got := runE2E(t, src)
	if got != "10" {
		t.Fatalf("expected %q, got %q", "10", got)
	}
}

func TestE2E_ForCallFn(t *testing.T) {
	src := `fn double(x) {
	<- x * [2]
}
for i in [1, 2, 3] {
	|> str(double(i))
}`
	got := runE2E(t, src)
	if got != "2\n4\n6" {
		t.Fatalf("expected %q, got %q", "2\n4\n6", got)
	}
}

func TestE2E_IfElseInsideForInsideFn(t *testing.T) {
	src := `fn classify(items) {
	for x in items {
		if x > [0] {
			|> "pos"
		} else {
			|> "non-pos"
		}
	}
	<- []
}
classify([1, -1, 0, 5])`
	got := runE2E(t, src)
	if got != "pos\nnon-pos\nnon-pos\npos" {
		t.Fatalf("expected %q, got %q", "pos\nnon-pos\nnon-pos\npos", got)
	}
}

func TestE2E_FnReturnIfElse(t *testing.T) {
	src := `fn max(a, b) {
	if a > b {
		<- a
	}
	<- b
}
|> str(max([5], [3]))
|> str(max([2], [8]))`
	got := runE2E(t, src)
	if got != "5\n8" {
		t.Fatalf("expected %q, got %q", "5\n8", got)
	}
}

func TestE2E_MultipleFnsCalling(t *testing.T) {
	src := `fn square(x) {
	<- x * x
}
fn sum_squares(a, b) {
	<- square(a) + square(b)
}
|> str(sum_squares([3], [4]))`
	got := runE2E(t, src)
	if got != "25" {
		t.Fatalf("expected %q, got %q", "25", got)
	}
}

func TestE2E_SumArray(t *testing.T) {
	src := `fn sum(arr) {
	total = [0]
	for x in arr {
		total = total + x
	}
	<- total
}
|> str(sum([1, 2, 3, 4, 5]))`
	got := runE2E(t, src)
	if got != "15" {
		t.Fatalf("expected %q, got %q", "15", got)
	}
}

func TestE2E_BareFnCall(t *testing.T) {
	src := `fn greet() {
	|> "hello"
	<- []
}
greet()`
	got := runE2E(t, src)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_DeepNesting(t *testing.T) {
	src := `fn deep(x) {
	if x > [0] {
		for i in [1] {
			if x == [1] {
				<- "one"
			}
		}
	}
	<- "other"
}
|> deep([1])
|> deep([0])`
	got := runE2E(t, src)
	if got != "one\nother" {
		t.Fatalf("expected %q, got %q", "one\nother", got)
	}
}

// --- Edge cases ---

func TestE2E_ParamShadowsGlobal(t *testing.T) {
	src := `x = [100]
fn foo(x) {
	<- x + [1]
}
|> str(foo([5]))
|> str(x)`
	got := runE2E(t, src)
	if got != "6\n100" {
		t.Fatalf("expected %q, got %q", "6\n100", got)
	}
}

func TestE2E_FnAllStatements(t *testing.T) {
	src := `fn all_statements(x) {
	y = x + [1]
	if y > [5] {
		|> "big"
	}
	for i in [1, 2] {
		|> str(i)
	}
	<- y
}
|> str(all_statements([10]))`
	got := runE2E(t, src)
	if got != "big\n1\n2\n11" {
		t.Fatalf("expected %q, got %q", "big\n1\n2\n11", got)
	}
}

func TestE2E_FnThreeParams(t *testing.T) {
	src := `fn clamp(x, lo, hi) {
	if x < lo {
		<- lo
	}
	if x > hi {
		<- hi
	}
	<- x
}
|> str(clamp([5], [0], [10]))
|> str(clamp(-[1], [0], [10]))
|> str(clamp([15], [0], [10]))`
	got := runE2E(t, src)
	if got != "5\n0\n10" {
		t.Fatalf("expected %q, got %q", "5\n0\n10", got)
	}
}

func TestE2E_IfTrueKeyword(t *testing.T) {
	got := runE2E(t, "if true { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_IfFalseKeyword(t *testing.T) {
	got := runE2E(t, "if false { |> \"yes\" } else { |> \"no\" }")
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

// Roadmap e2e — fns.slop

func TestE2E_FnsSlop(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}

result = add([3], [4])
|> str(result)

if [1] == [1] {
	|> "equal"
}

items = [10, 20, 30]
for item in items {
	|> str(item)
}`
	got := runE2E(t, src)
	expected := "7\nequal\n10\n20\n30"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}
