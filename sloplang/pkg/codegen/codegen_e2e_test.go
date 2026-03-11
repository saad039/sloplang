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
	runCmd.Dir = tmpDir
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run failed: %v\n%s", err, string(runOut))
	}

	return strings.TrimRight(string(runOut), "\n")
}

// runE2EWithStdin is like runE2E but pipes stdinInput to the process's stdin.
func runE2EWithStdin(t *testing.T, source, stdinInput string) string {
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
	runCmd.Dir = tmpDir
	runCmd.Stdin = strings.NewReader(stdinInput)
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
	expected := "hello world[1, 2, 3]"
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
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestE2E_ReassignVariable(t *testing.T) {
	// In Phase 1 each assignment is :=, so two assignments to same name
	// would fail. This test documents current behavior with distinct names.
	got := runE2E(t, `x = [1]
y = [2]
|> x
|> y`)
	if got != "[1][2]" {
		t.Fatalf("expected %q, got %q", "[1][2]", got)
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
	if got != "helloworld" {
		t.Fatalf("expected %q, got %q", "helloworld", got)
	}
}

func TestE2E_UintLiteral(t *testing.T) {
	got := runE2E(t, `x = [42u]
|> x`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
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
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestE2E_SubZero(t *testing.T) {
	got := runE2E(t, "x = [0] - [0]\n|> str(x)")
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_MulByZero(t *testing.T) {
	got := runE2E(t, "x = [1] * [0]\n|> str(x)")
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
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
	if got != "[6]" {
		t.Fatalf("expected %q, got %q", "[6]", got)
	}
}

func TestE2E_AddUint(t *testing.T) {
	got := runE2E(t, "x = [42u] + [8u]\n|> str(x)")
	if got != "[50]" {
		t.Fatalf("expected %q, got %q", "[50]", got)
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
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_NegateNegative(t *testing.T) {
	got := runE2E(t, "x = -[-5]\n|> str(x)")
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestE2E_DoubleNegate(t *testing.T) {
	// With Phase 4, -- is the remove operator. Use -(-[5]) for double negate.
	got := runE2E(t, "x = -(-[5])\n|> str(x)")
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestE2E_NegateFloat(t *testing.T) {
	got := runE2E(t, "x = -[3.14]\n|> str(x)")
	if got != "[-3.14]" {
		t.Fatalf("expected %q, got %q", "[-3.14]", got)
	}
}

func TestE2E_NotFalsy(t *testing.T) {
	got := runE2E(t, "x = ![]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	// [0] is not a valid boolean in strict mode; ![0] panics
	runE2EExpectPanic(t, "x = ![0]")
}

func TestE2E_DoubleNotTruthy(t *testing.T) {
	got := runE2E(t, "x = !![1]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_LteLess(t *testing.T) {
	got := runE2E(t, "x = [1] <= [2]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_GteEqual(t *testing.T) {
	got := runE2E(t, "x = [2] >= [2]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_StringNeq(t *testing.T) {
	got := runE2E(t, "x = \"abc\" != \"def\"\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// Logical (6)

func TestE2E_AndBothTruthy(t *testing.T) {
	got := runE2E(t, "x = [1] && [1]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_OrLeftTruthy(t *testing.T) {
	got := runE2E(t, "x = [1] || []\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[7]" {
		t.Fatalf("expected %q, got %q", "[7]", got)
	}
}

func TestE2E_PrecedenceMulBeforeAddRight(t *testing.T) {
	got := runE2E(t, "x = [1] + [2] * [3]\n|> str(x)")
	if got != "[7]" {
		t.Fatalf("expected %q, got %q", "[7]", got)
	}
}

func TestE2E_PrecedenceParens(t *testing.T) {
	got := runE2E(t, "x = ([1] + [2]) * [3]\n|> str(x)")
	if got != "[9]" {
		t.Fatalf("expected %q, got %q", "[9]", got)
	}
}

func TestE2E_PrecedenceLeftAssocSub(t *testing.T) {
	got := runE2E(t, "x = [10] - [2] - [3]\n|> str(x)")
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestE2E_PrecedencePower(t *testing.T) {
	got := runE2E(t, "x = [2] ** [3]\n|> str(x)")
	if got != "[8]" {
		t.Fatalf("expected %q, got %q", "[8]", got)
	}
}

func TestE2E_PrecedenceArithBeforeComparison(t *testing.T) {
	got := runE2E(t, "x = [1] + [2] > [2]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_PrecedenceComparisonBeforeLogical(t *testing.T) {
	got := runE2E(t, "x = [1] > [0] && [2] > [1]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_PrecedenceAndBeforeOr(t *testing.T) {
	got := runE2E(t, "x = [1] > [0] || [] && []\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
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
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
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
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

// Chained (4)

func TestE2E_ChainedAdd(t *testing.T) {
	got := runE2E(t, "x = [1] + [2] + [3] + [4]\n|> str(x)")
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestE2E_ChainedMulDiv(t *testing.T) {
	got := runE2E(t, "x = [10] / [2] * [3]\n|> str(x)")
	if got != "[15]" {
		t.Fatalf("expected %q, got %q", "[15]", got)
	}
}

func TestE2E_ChainedLogical(t *testing.T) {
	got := runE2E(t, "x = [1] == [1] && [2] == [2] && [3] == [3]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_UnaryThenBinary(t *testing.T) {
	got := runE2E(t, "x = -[1] + [2]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// Roadmap e2e — math.slop

func TestE2E_MathSlop(t *testing.T) {
	got := runE2E(t, "x = [10, 20] + [3, 4]\n|> str(x)\ny = [5] > [3]\n|> str(y)\nz = -[1, 2, 3]\n|> str(z)")
	expected := "[13, 24][1][-1, -2, -3]"
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
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestE2E_FnTwoParams(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
result = add([3], [4])
|> str(result)`
	got := runE2E(t, src)
	if got != "[7]" {
		t.Fatalf("expected %q, got %q", "[7]", got)
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
	if got != "[6]" {
		t.Fatalf("expected %q, got %q", "[6]", got)
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
	if got != "[3][2][1]done" {
		t.Fatalf("expected %q, got %q", "[3][2][1]done", got)
	}
}

func TestE2E_FnWithStrInBody(t *testing.T) {
	src := `fn show(x) {
	|> str(x)
	<- []
}
show([42])`
	got := runE2E(t, src)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestE2E_FnExprAsArg(t *testing.T) {
	src := `fn double(x) {
	<- x + x
}
|> str(double([1] + [2]))`
	got := runE2E(t, src)
	if got != "[6]" {
		t.Fatalf("expected %q, got %q", "[6]", got)
	}
}

func TestE2E_FnResultInExpr(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
|> str(add([1], [2]) + [3])`
	got := runE2E(t, src)
	if got != "[6]" {
		t.Fatalf("expected %q, got %q", "[6]", got)
	}
}

// --- Return ---

func TestE2E_ReturnValue(t *testing.T) {
	src := `fn five() {
	<- [5]
}
|> str(five())`
	got := runE2E(t, src)
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
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
	if got != "zerononzero" {
		t.Fatalf("expected %q, got %q", "zerononzero", got)
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
	if got != "positivezeronegative" {
		t.Fatalf("expected %q, got %q", "positivezeronegative", got)
	}
}

func TestE2E_ReturnExpr(t *testing.T) {
	src := `fn sum(a, b) {
	<- a + b
}
|> str(sum([10], [20]))`
	got := runE2E(t, src)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
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
	if got != "[5][3]" {
		t.Fatalf("expected %q, got %q", "[5][3]", got)
	}
}

func TestE2E_IfNotOperator(t *testing.T) {
	// [0] is not a valid boolean in strict mode; ![0] panics
	runE2EExpectPanic(t, `if ![0] { |> "not" } else { |> "yes" }`)
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
	if got != "cd" {
		t.Fatalf("expected %q, got %q", "cd", got)
	}
}

// --- For/In ---

func TestE2E_ForInArray(t *testing.T) {
	src := `items = [1, 2, 3]
for item in items {
	|> str(item)
}`
	got := runE2E(t, src)
	if got != "[1][2][3]" {
		t.Fatalf("expected %q, got %q", "[1][2][3]", got)
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
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestE2E_ForInComputation(t *testing.T) {
	src := `for item in [1, 2, 3] {
	|> str(item + [10])
}`
	got := runE2E(t, src)
	if got != "[11][12][13]" {
		t.Fatalf("expected %q, got %q", "[11][12][13]", got)
	}
}

func TestE2E_NestedForLoops(t *testing.T) {
	src := `for i in [1, 2] {
	for j in [10, 20] {
		|> str(i + j)
	}
}`
	got := runE2E(t, src)
	if got != "[11][21][12][22]" {
		t.Fatalf("expected %q, got %q", "[11][21][12][22]", got)
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
	if got != "[10][20][30]" {
		t.Fatalf("expected %q, got %q", "[10][20][30]", got)
	}
}

func TestE2E_ForWithIfInside(t *testing.T) {
	src := `for x in [1, 2, 3, 4] {
	if x > [2] {
		|> str(x)
	}
}`
	got := runE2E(t, src)
	if got != "[3][4]" {
		t.Fatalf("expected %q, got %q", "[3][4]", got)
	}
}

func TestE2E_ForInStringArray(t *testing.T) {
	src := `for s in ["a", "b", "c"] {
	|> s
}`
	got := runE2E(t, src)
	if got != "abc" {
		t.Fatalf("expected %q, got %q", "abc", got)
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
	if got != "[10][20]" {
		t.Fatalf("expected %q, got %q", "[10][20]", got)
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
	if got != "[2][1]" {
		t.Fatalf("expected %q, got %q", "[2][1]", got)
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
	if got != "[1, 2][3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2][3, 4]", got)
	}
}

func TestE2E_MultiAssignDirect(t *testing.T) {
	src := `a, b = [10, 20]
|> str(a)
|> str(b)`
	got := runE2E(t, src)
	if got != "[10][20]" {
		t.Fatalf("expected %q, got %q", "[10][20]", got)
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
	expected := "[0][1][1][2][3][5][8][13][21][34]"
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
	if got != "[120][1][1]" {
		t.Fatalf("expected %q, got %q", "[120][1][1]", got)
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
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
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
	if got != "[2][4][6]" {
		t.Fatalf("expected %q, got %q", "[2][4][6]", got)
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
	if got != "posnon-posnon-pospos" {
		t.Fatalf("expected %q, got %q", "posnon-posnon-pospos", got)
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
	if got != "[5][8]" {
		t.Fatalf("expected %q, got %q", "[5][8]", got)
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
	if got != "[25]" {
		t.Fatalf("expected %q, got %q", "[25]", got)
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
	if got != "[15]" {
		t.Fatalf("expected %q, got %q", "[15]", got)
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
	if got != "oneother" {
		t.Fatalf("expected %q, got %q", "oneother", got)
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
	if got != "[6][100]" {
		t.Fatalf("expected %q, got %q", "[6][100]", got)
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
	if got != "big[1][2][11]" {
		t.Fatalf("expected %q, got %q", "big[1][2][11]", got)
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
	if got != "[5][0][10]" {
		t.Fatalf("expected %q, got %q", "[5][0][10]", got)
	}
}

func TestE2E_IfTrueKeyword(t *testing.T) {
	// bare true is rejected; use [1] for truthy
	got := runE2E(t, `if [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_IfFalseKeyword(t *testing.T) {
	// bare false is rejected; use [] for falsy
	got := runE2E(t, `if [] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

// --- Infinite loop + break ---

func TestE2E_ForLoopBreakImmediate(t *testing.T) {
	src := `for {
	|> "once"
	break
}
|> "done"`
	got := runE2E(t, src)
	if got != "oncedone" {
		t.Fatalf("expected %q, got %q", "oncedone", got)
	}
}

func TestE2E_ForLoopCounterBreak(t *testing.T) {
	src := `i = [0]
for {
	if i == [3] {
		break
	}
	|> str(i)
	i = i + [1]
}`
	got := runE2E(t, src)
	if got != "[0][1][2]" {
		t.Fatalf("expected %q, got %q", "[0][1][2]", got)
	}
}

func TestE2E_ForLoopInsideFn(t *testing.T) {
	src := `fn countTo(n) {
	i = [0]
	for {
		if i == n {
			break
		}
		|> str(i)
		i = i + [1]
	}
	<- n
}
countTo([4])`
	got := runE2E(t, src)
	if got != "[0][1][2][3]" {
		t.Fatalf("expected %q, got %q", "[0][1][2][3]", got)
	}
}

func TestE2E_ForLoopAccumulate(t *testing.T) {
	src := `fn sumTo(n) {
	total = [0]
	i = [1]
	for {
		if i > n {
			break
		}
		total = total + i
		i = i + [1]
	}
	<- total
}
|> str(sumTo([10]))`
	got := runE2E(t, src)
	if got != "[55]" {
		t.Fatalf("expected %q, got %q", "[55]", got)
	}
}

func TestE2E_ForLoopNestedBreak(t *testing.T) {
	src := `for {
	for {
		|> "inner"
		break
	}
	|> "outer"
	break
}`
	got := runE2E(t, src)
	if got != "innerouter" {
		t.Fatalf("expected %q, got %q", "innerouter", got)
	}
}

func TestE2E_ForLoopWithForIn(t *testing.T) {
	src := `count = [0]
for {
	for x in [1, 2, 3] {
		count = count + x
	}
	break
}
|> str(count)`
	got := runE2E(t, src)
	if got != "[6]" {
		t.Fatalf("expected %q, got %q", "[6]", got)
	}
}

func TestE2E_BreakInForIn(t *testing.T) {
	// break also works inside for-in loops
	src := `for x in [1, 2, 3, 4, 5] {
	if x == [3] {
		break
	}
	|> str(x)
}`
	got := runE2E(t, src)
	if got != "[1][2]" {
		t.Fatalf("expected %q, got %q", "[1][2]", got)
	}
}

func TestE2E_ForLoopFibonacci(t *testing.T) {
	// Fibonacci using infinite loop instead of recursion
	src := `fn fib_loop(n) {
	if n <= [1] {
		<- n
	}
	a = [0]
	b = [1]
	i = [2]
	for {
		if i > n {
			break
		}
		temp = b
		b = a + b
		a = temp
		i = i + [1]
	}
	<- b
}
|> str(fib_loop([0]))
|> str(fib_loop([1]))
|> str(fib_loop([5]))
|> str(fib_loop([10]))`
	got := runE2E(t, src)
	if got != "[0][1][5][55]" {
		t.Fatalf("expected %q, got %q", "[0][1][5][55]", got)
	}
}

// ==========================================
// Negative / Edge Case / Boundary E2E Tests
// ==========================================

// --- Runtime panics (programs that should crash) ---

func runE2EExpectPanic(t *testing.T, source string) {
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
	err = runCmd.Run()
	if err == nil {
		t.Fatal("expected program to panic/exit non-zero, but it succeeded")
	}
}

func TestE2E_PanicAddLengthMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] + [3]`)
}

func TestE2E_PanicAddTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] + [1.0]`)
}

func TestE2E_PanicSubLengthMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] - [3]`)
}

func TestE2E_PanicMulTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] * [1u]`)
}

func TestE2E_PanicDivByZeroInt(t *testing.T) {
	runE2EExpectPanic(t, `x = [10] / [0]`)
}

func TestE2E_PanicDivByZeroFloat(t *testing.T) {
	runE2EExpectPanic(t, `x = [10.0] / [0.0]`)
}

func TestE2E_PanicDivByZeroUint(t *testing.T) {
	runE2EExpectPanic(t, `x = [10u] / [0u]`)
}

func TestE2E_PanicModLengthMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [7] % [3, 2]`)
}

func TestE2E_PanicEqMultiElement(t *testing.T) {
	// Multi-element comparison no longer panics — it succeeds
	_ = runE2E(t, `x = [1, 2] == [1, 2]`)
}

func TestE2E_PanicLtMultiElement(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] < [3, 4]`)
}

func TestE2E_PanicGtMultiElement(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] > [3, 4]`)
}

func TestE2E_PanicEqTypeMismatch(t *testing.T) {
	// Type mismatch comparison no longer panics — it succeeds
	_ = runE2E(t, `x = [1] == [1.0]`)
}

func TestE2E_PanicEqEmptyVsNonEmpty(t *testing.T) {
	// Empty vs non-empty comparison no longer panics — it succeeds
	_ = runE2E(t, `x = [] == [1]`)
}

func TestE2E_PanicUnpackTwoTooFew(t *testing.T) {
	runE2EExpectPanic(t, `a, b = [1]`)
}

func TestE2E_PanicUnpackTwoEmpty(t *testing.T) {
	runE2EExpectPanic(t, `a, b = []`)
}

// --- Boundary: empty and zero values ---

func TestE2E_EmptyArrayIsFalsy(t *testing.T) {
	got := runE2E(t, `if [] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestE2E_ZeroIsTruthy(t *testing.T) {
	// [0] is not a valid boolean in strict mode; panics
	runE2EExpectPanic(t, `if [0] { |> "yes" } else { |> "no" }`)
}

func TestE2E_EmptyStringIsTruthy(t *testing.T) {
	// strings are not valid booleans in strict mode; panics
	runE2EExpectPanic(t, `if "" { |> "yes" } else { |> "no" }`)
}

func TestE2E_AddEmptyArrays(t *testing.T) {
	got := runE2E(t, `x = [] + []
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_NegateEmptyArray(t *testing.T) {
	got := runE2E(t, `x = -[]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_NotEmptyArray(t *testing.T) {
	// ![] → truthy ([1])
	got := runE2E(t, `|> str(![])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_NotNonEmpty(t *testing.T) {
	// ![1] → falsy ([])
	got := runE2E(t, `|> str(![1])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_DoubleNotEmpty(t *testing.T) {
	// !![] → falsy
	got := runE2E(t, `|> str(!![])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Variable reassignment edge cases ---

func TestE2E_ReassignSameVar(t *testing.T) {
	got := runE2E(t, `x = [1]
x = [2]
x = [3]
|> str(x)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_ReassignInLoop(t *testing.T) {
	got := runE2E(t, `x = [0]
for i in [1, 2, 3] {
	x = x + i
}
|> str(x)`)
	if got != "[6]" {
		t.Fatalf("expected %q, got %q", "[6]", got)
	}
}

func TestE2E_ReassignInNestedLoops(t *testing.T) {
	got := runE2E(t, `x = [0]
for i in [1, 2] {
	for j in [10, 20] {
		x = x + j
	}
}
|> str(x)`)
	if got != "[60]" {
		t.Fatalf("expected %q, got %q", "[60]", got)
	}
}

// --- Function edge cases ---

func TestE2E_FnReturnsEmptyArray(t *testing.T) {
	got := runE2E(t, `fn empty() {
	<- []
}
|> str(empty())`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_FnReturnString(t *testing.T) {
	got := runE2E(t, `fn greet() {
	<- "hello"
}
|> greet()`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_FnMutualRecursion(t *testing.T) {
	src := `fn isEven(n) {
	if n == [0] { <- [1] }
	<- isOdd(n - [1])
}
fn isOdd(n) {
	if n == [0] { <- [] }
	<- isEven(n - [1])
}
if isEven([4]) { |> "even" } else { |> "odd" }
if isEven([3]) { |> "even" } else { |> "odd" }`
	got := runE2E(t, src)
	if got != "evenodd" {
		t.Fatalf("expected %q, got %q", "evenodd", got)
	}
}

func TestE2E_FnManyParams(t *testing.T) {
	src := `fn sum5(a, b, c, d, e) {
	<- a + b + c + d + e
}
|> str(sum5([1], [2], [3], [4], [5]))`
	got := runE2E(t, src)
	if got != "[15]" {
		t.Fatalf("expected %q, got %q", "[15]", got)
	}
}

func TestE2E_FnScopeIsolation(t *testing.T) {
	// With hoisting, top-level variables are package-level globals visible to functions.
	got := runE2E(t, `x = [100]
fn getX() {
	<- x
}
|> str(getX())`)
	if got != "[100]" {
		t.Fatalf("expected [100], got %q", got)
	}
}

// --- Control flow edge cases ---

func TestE2E_IfElseChain(t *testing.T) {
	src := `fn classify(x) {
	if x > [100] {
		<- "big"
	} else {
		if x > [10] {
			<- "medium"
		} else {
			if x > [0] {
				<- "small"
			} else {
				<- "zero-or-neg"
			}
		}
	}
}
|> classify([200])
|> classify([50])
|> classify([5])
|> classify([0])`
	got := runE2E(t, src)
	if got != "bigmediumsmallzero-or-neg" {
		t.Fatalf("expected %q, got %q", "bigmediumsmallzero-or-neg", got)
	}
}

func TestE2E_ForInNoBody(t *testing.T) {
	// Iterating but doing nothing — must compile
	got := runE2E(t, `for x in [1, 2, 3] { }
|> "done"`)
	if got != "done" {
		t.Fatalf("expected %q, got %q", "done", got)
	}
}

func TestE2E_ForLoopEmptyBody(t *testing.T) {
	// Infinite loop with immediate break, empty-ish body
	got := runE2E(t, `for { break }
|> "done"`)
	if got != "done" {
		t.Fatalf("expected %q, got %q", "done", got)
	}
}

func TestE2E_BreakFromForInEarly(t *testing.T) {
	// Break from first iteration
	got := runE2E(t, `for x in [10, 20, 30] {
	|> str(x)
	break
}`)
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestE2E_NestedBreakOnlyBreaksInner(t *testing.T) {
	src := `count = [0]
for i in [1, 2, 3] {
	for j in [10, 20, 30] {
		if j == [20] {
			break
		}
		count = count + [1]
	}
	count = count + [100]
}
|> str(count)`
	got := runE2E(t, src)
	// Each outer iteration: inner does 1 iter (j=10) then breaks at j=20
	// So: 3 outer * (1 inner + 100) = 303
	if got != "[303]" {
		t.Fatalf("expected %q, got %q", "[303]", got)
	}
}

// --- Large / stress ---

func TestE2E_LargeArray(t *testing.T) {
	// Build a 50-element array
	elems := make([]string, 50)
	for i := range elems {
		elems[i] = fmt.Sprintf("%d", i+1)
	}
	src := fmt.Sprintf("x = [%s]\n|> str(x)", strings.Join(elems, ", "))
	got := runE2E(t, src)
	expected := fmt.Sprintf("[%s]", strings.Join(elems, ", "))
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestE2E_LoopManyIterations(t *testing.T) {
	src := `i = [0]
for {
	if i == [100] { break }
	i = i + [1]
}
|> str(i)`
	got := runE2E(t, src)
	if got != "[100]" {
		t.Fatalf("expected %q, got %q", "[100]", got)
	}
}

// --- String edge cases ---

func TestE2E_EmptyStringOutput(t *testing.T) {
	got := runE2E(t, `|> ""`)
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestE2E_StringWithEscapes(t *testing.T) {
	got := runE2E(t, `|> "hello\tworld"`)
	if got != "hello\tworld" {
		t.Fatalf("expected %q, got %q", "hello\tworld", got)
	}
}

func TestE2E_StringWithNewline(t *testing.T) {
	got := runE2E(t, `|> "line1\nline2"`)
	if got != "line1\nline2" {
		t.Fatalf("expected %q, got %q", "line1\nline2", got)
	}
}

// --- Expression edge cases ---

func TestE2E_NestedParens(t *testing.T) {
	got := runE2E(t, `|> str((([1] + [2]) * [3]))`)
	if got != "[9]" {
		t.Fatalf("expected %q, got %q", "[9]", got)
	}
}

func TestE2E_NegativeResult(t *testing.T) {
	got := runE2E(t, `|> str([1] - [5])`)
	if got != "[-4]" {
		t.Fatalf("expected %q, got %q", "[-4]", got)
	}
}

func TestE2E_PowerOfZero(t *testing.T) {
	got := runE2E(t, `|> str([5] ** [0])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ZeroToThePowerZero(t *testing.T) {
	got := runE2E(t, `|> str([0] ** [0])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_MultiAssignExtraElements(t *testing.T) {
	// Unpack from 3-element array — only takes first 2
	got := runE2E(t, `a, b = [10, 20, 30]
|> str(a)
|> str(b)`)
	if got != "[10][20]" {
		t.Fatalf("expected %q, got %q", "[10][20]", got)
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
	expected := "[7]equal[10][20][30]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

// ==========================================
// Phase 4: Array Operators E2E Tests
// ==========================================

// --- Index (8 tests) ---

func TestE2E_IndexFirst(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr@0)`)
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestE2E_IndexLast(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr@2)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestE2E_IndexMiddle(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr@1)`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}

func TestE2E_IndexNested(t *testing.T) {
	got := runE2E(t, `m = [[1, 2], [3, 4]]
|> str(m@0)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestE2E_IndexFnResult(t *testing.T) {
	got := runE2E(t, `fn f() {
	<- [10, 20]
}
|> str(f()@1)`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}

func TestE2E_IndexSet(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
arr@1 = [99]
|> str(arr)`)
	if got != "[10, 99, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 99, 30]", got)
	}
}

func TestE2E_IndexSetVarIndex(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
idx = [2]
arr$idx = [99]
|> str(arr)`)
	if got != "[10, 20, 99]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 99]", got)
	}
}

func TestE2E_PanicIndexOutOfBounds(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30]
x = arr@5`)
}

// --- Length (4 tests) ---

func TestE2E_LengthBasic(t *testing.T) {
	got := runE2E(t, `|> str(#[1, 2, 3])`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_LengthEmpty(t *testing.T) {
	got := runE2E(t, `|> str(#[])`)
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_LengthInExpr(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
if #arr == [3] {
	|> "correct"
}`)
	if got != "correct" {
		t.Fatalf("expected %q, got %q", "correct", got)
	}
}

func TestE2E_LengthOfConcat(t *testing.T) {
	got := runE2E(t, `|> str(#([1, 2] ++ [3]))`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

// --- Push (4 tests) ---

func TestE2E_PushBasic(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << [5]
|> str(arr)`)
	if got != "[1, 2, 3, 5]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 5]", got)
	}
}

func TestE2E_PushToEmpty(t *testing.T) {
	got := runE2E(t, `arr = []
arr << [1]
|> str(arr)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_PushThenLength(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr << [3]
|> str(#arr)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_PushInsideLoop(t *testing.T) {
	got := runE2E(t, `arr = []
for x in [10, 20, 30] {
	arr << x
}
|> str(arr)`)
	if got != "[10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30]", got)
	}
}

// --- Pop (4 tests) ---

func TestE2E_PopReturnsLast(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
x = >>arr
|> str(x)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_PopModifiesArray(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
x = >>arr
|> str(arr)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestE2E_PopMultiple(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
a = >>arr
b = >>arr
|> str(a)
|> str(b)
|> str(arr)`)
	if got != "[3][2][1]" {
		t.Fatalf("expected %q, got %q", "[3][2][1]", got)
	}
}

func TestE2E_PanicPopEmpty(t *testing.T) {
	runE2EExpectPanic(t, `arr = []
x = >>arr`)
}

// --- RemoveAt (4 tests) ---

func TestE2E_RemoveAtFirst(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
removed = arr ~@ [0]
|> str(removed)
|> str(arr)`)
	if got != "[10][20, 30]" {
		t.Fatalf("expected %q, got %q", "[10][20, 30]", got)
	}
}

func TestE2E_RemoveAtMiddle(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
removed = arr ~@ [1]
|> str(removed)
|> str(arr)`)
	if got != "[20][10, 30]" {
		t.Fatalf("expected %q, got %q", "[20][10, 30]", got)
	}
}

func TestE2E_RemoveAtLast(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
removed = arr ~@ [2]
|> str(removed)
|> str(arr)`)
	if got != "[30][10, 20]" {
		t.Fatalf("expected %q, got %q", "[30][10, 20]", got)
	}
}

func TestE2E_PanicRemoveAtOutOfBounds(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30]
removed = arr ~@ [5]`)
}

// --- Slice (6 tests) ---

func TestE2E_SliceBasic(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
|> str(arr::1::3)`)
	if got != "[20, 30]" {
		t.Fatalf("expected %q, got %q", "[20, 30]", got)
	}
}

func TestE2E_SliceFromStart(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
|> str(arr::0::2)`)
	if got != "[10, 20]" {
		t.Fatalf("expected %q, got %q", "[10, 20]", got)
	}
}

func TestE2E_SliceToEnd(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
|> str(arr::2::4)`)
	if got != "[30, 40]" {
		t.Fatalf("expected %q, got %q", "[30, 40]", got)
	}
}

func TestE2E_SliceFull(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
|> str(arr::0::4)`)
	if got != "[10, 20, 30, 40]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30, 40]", got)
	}
}

func TestE2E_SliceEmpty(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
|> str(arr::2::2)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_PanicSliceOutOfBounds(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30, 40]
x = arr::0::10`)
}

// --- Concat (4 tests) ---

func TestE2E_ConcatBasic(t *testing.T) {
	got := runE2E(t, `|> str([1, 2] ++ [3, 4])`)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

func TestE2E_ConcatWithEmpty(t *testing.T) {
	got := runE2E(t, `|> str([] ++ [1])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ConcatStrings(t *testing.T) {
	got := runE2E(t, `|> str(["a"] ++ ["b"])`)
	if got != "[a, b]" {
		t.Fatalf("expected %q, got %q", "[a, b]", got)
	}
}

func TestE2E_ConcatChained(t *testing.T) {
	got := runE2E(t, `|> str(([1] ++ [2]) ++ [3])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

// --- Remove (3 tests) ---

func TestE2E_RemoveExisting(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] -- [2])`)
	if got != "[1, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 3]", got)
	}
}

func TestE2E_RemoveNonExistent(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] -- [5])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestE2E_RemoveDuplicate(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3, 2] -- [2])`)
	if got != "[1, 3, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 3, 2]", got)
	}
}

// --- Contains (4 tests) ---

func TestE2E_ContainsExisting(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] ?? [2])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ContainsNonExistent(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] ?? [5])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_ContainsInIf(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
if arr ?? [2] {
	|> "found"
}`)
	if got != "found" {
		t.Fatalf("expected %q, got %q", "found", got)
	}
}

func TestE2E_ContainsOnEmpty(t *testing.T) {
	got := runE2E(t, `|> str([] ?? [1])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Unique (3 tests) ---

func TestE2E_UniqueWithDups(t *testing.T) {
	got := runE2E(t, `|> str(~[1, 2, 2, 3, 1])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestE2E_UniqueEmpty(t *testing.T) {
	got := runE2E(t, `|> str(~[])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_UniqueAlreadyUnique(t *testing.T) {
	got := runE2E(t, `|> str(~[1, 2, 3])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

// --- Combined / Integration (10 tests) ---

func TestE2E_PushThenIterate(t *testing.T) {
	got := runE2E(t, `arr = []
arr << [1]
arr << [2]
arr << [3]
for x in arr {
	|> str(x)
}`)
	if got != "[1][2][3]" {
		t.Fatalf("expected %q, got %q", "[1][2][3]", got)
	}
}

func TestE2E_FilterPattern(t *testing.T) {
	got := runE2E(t, `src = [1, 2, 3, 4, 5]
evens = []
for x in src {
	if x % [2] == [0] {
		evens << x
	}
}
|> str(evens)`)
	if got != "[2, 4]" {
		t.Fatalf("expected %q, got %q", "[2, 4]", got)
	}
}

func TestE2E_FnReturnsArrayCallerIndexes(t *testing.T) {
	got := runE2E(t, `fn pair() {
	<- [100, 200]
}
|> str(pair()@0)
|> str(pair()@1)`)
	if got != "[100][200]" {
		t.Fatalf("expected %q, got %q", "[100][200]", got)
	}
}

func TestE2E_SliceThenConcat(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
left = arr::0::2
right = arr::2::4
combined = left ++ right
|> str(combined)`)
	if got != "[10, 20, 30, 40]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30, 40]", got)
	}
}

func TestE2E_AccumulatePattern(t *testing.T) {
	got := runE2E(t, `result = []
i = [1]
for {
	if i > [5] {
		break
	}
	result << i
	i = i + [1]
}
|> str(result)`)
	if got != "[1, 2, 3, 4, 5]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4, 5]", got)
	}
}

func TestE2E_FnWithArrayManip(t *testing.T) {
	got := runE2E(t, `fn reverse(arr) {
	result = []
	i = #arr - [1]
	for {
		if i < [0] {
			break
		}
		result << arr$i
		i = i - [1]
	}
	<- result
}
|> str(reverse([1, 2, 3]))`)
	if got != "[3, 2, 1]" {
		t.Fatalf("expected %q, got %q", "[3, 2, 1]", got)
	}
}

func TestE2E_IndexSetThenRead(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
arr@0 = [99]
arr@2 = [77]
|> str(arr@0)
|> str(arr@1)
|> str(arr@2)`)
	if got != "[99][20][77]" {
		t.Fatalf("expected %q, got %q", "[99][20][77]", got)
	}
}

func TestE2E_ContainsInFilterLoop(t *testing.T) {
	got := runE2E(t, `allowed = [2, 4, 6]
src = [1, 2, 3, 4, 5, 6]
filtered = []
for x in src {
	if allowed ?? x {
		filtered << x
	}
}
|> str(filtered)`)
	if got != "[2, 4, 6]" {
		t.Fatalf("expected %q, got %q", "[2, 4, 6]", got)
	}
}

func TestE2E_PopInLoop(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
result = []
for {
	if #arr == [0] {
		break
	}
	x = >>arr
	result << x
}
|> str(result)`)
	if got != "[3, 2, 1]" {
		t.Fatalf("expected %q, got %q", "[3, 2, 1]", got)
	}
}

func TestE2E_UniqueAfterConcat(t *testing.T) {
	got := runE2E(t, `a = [1, 2, 3]
b = [2, 3, 4]
combined = a ++ b
|> str(~combined)`)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

// --- arrays.slop example (roadmap) ---

func TestE2E_ArraysSlopExample(t *testing.T) {
	src := `arr = [10, 20, 30, 40]
|> str(arr@0)
|> str(#arr)

arr << [50]
|> str(arr)

sub = arr::1::3
|> str(sub)

combined = [1, 2] ++ [3, 4]
|> str(combined)

has = arr ?? [20]
|> str(has)

uniq = ~[1, 2, 2, 3, 1]
|> str(uniq)`
	got := runE2E(t, src)
	expected := "[10][4][10, 20, 30, 40, 50][20, 30][1, 2, 3, 4][1][1, 2, 3]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

// === Phase 5 E2E Tests — Hashmaps ===

// --- Hashmap declaration (5 tests) ---

func TestE2E_HashDecl_TwoKeys(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> person@name
|> str(person@age)`)
	if got != "bob[30]" {
		t.Fatalf("expected %q, got %q", "bob[30]", got)
	}
}

func TestE2E_HashDecl_OneKey(t *testing.T) {
	got := runE2E(t, `data{x} = [[42]]
|> str(data@x)`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestE2E_HashDecl_Empty(t *testing.T) {
	got := runE2E(t, `counts{} = []
|> str(##counts)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_HashDecl_ThreeKeys(t *testing.T) {
	got := runE2E(t, `rgb{r, g, b} = [[255], [128], [0]]
|> str(rgb@r)
|> str(rgb@g)
|> str(rgb@b)`)
	if got != "[255][128][0]" {
		t.Fatalf("expected %q, got %q", "[255][128][0]", got)
	}
}

func TestE2E_HashDecl_MismatchPanic(t *testing.T) {
	runE2EExpectPanic(t, `bad{a, b, c} = [[1], [2]]`)
}

// --- Key access — literal (5 tests) ---

func TestE2E_KeyAccessLit_StringVal(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> person@name`)
	if got != "bob" {
		t.Fatalf("expected %q, got %q", "bob", got)
	}
}

func TestE2E_KeyAccessLit_NestedSlopValue(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> str(person@age)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestE2E_KeyAccessLit_AfterKeySet(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
person@age = [31]
|> str(person@age)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestE2E_KeyAccessLit_MultipleAccesses(t *testing.T) {
	got := runE2E(t, `config{host, port, debug} = ["localhost", [8080], [1]]
|> config@host
|> str(config@port)
|> str(config@debug)`)
	if got != "localhost[8080][1]" {
		t.Fatalf("expected %q, got %q", "localhost[8080][1]", got)
	}
}

func TestE2E_KeyAccessLit_NonExistentPanic(t *testing.T) {
	runE2EExpectPanic(t, `person{name, age} = ["bob", [30]]
|> person@email`)
}

// --- Key access — dynamic (4 tests) ---

func TestE2E_KeyAccessDyn_Basic(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
which = "name"
|> person$which`)
	if got != "bob" {
		t.Fatalf("expected %q, got %q", "bob", got)
	}
}

func TestE2E_KeyAccessDyn_DifferentKeys(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
k1 = "name"
k2 = "age"
|> person$k1
|> str(person$k2)`)
	if got != "bob[30]" {
		t.Fatalf("expected %q, got %q", "bob[30]", got)
	}
}

func TestE2E_KeyAccessDyn_InLoop(t *testing.T) {
	got := runE2E(t, `data{x, y} = [[10], [20]]
for k in ##data {
	|> str(data$k)
}`)
	if got != "[10][20]" {
		t.Fatalf("expected %q, got %q", "[10][20]", got)
	}
}

func TestE2E_KeyAccessDyn_AfterDynSet(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
key = "age"
person$key = [31]
|> str(person$key)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

// --- Key set — literal (4 tests) ---

func TestE2E_KeySetLit_Existing(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
person@age = [31]
|> str(person@age)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestE2E_KeySetLit_AddNew(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
person@email = "bob@test.com"
|> person@email`)
	if got != "bob@test.com" {
		t.Fatalf("expected %q, got %q", "bob@test.com", got)
	}
}

func TestE2E_KeySetLit_SetThenReadMultiple(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
person@age = [31]
person@email = "bob@test.com"
|> person@name
|> str(person@age)
|> person@email`)
	if got != "bob[31]bob@test.com" {
		t.Fatalf("expected %q, got %q", "bob[31]bob@test.com", got)
	}
}

func TestE2E_KeySetLit_InIfBlock(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
if [1] == [1] {
	person@age = [99]
}
|> str(person@age)`)
	if got != "[99]" {
		t.Fatalf("expected %q, got %q", "[99]", got)
	}
}

// --- Key set — dynamic (3 tests) ---

func TestE2E_KeySetDyn_Basic(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
key = "age"
person$key = [31]
|> str(person@age)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestE2E_KeySetDyn_AddNewKey(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
key = "email"
person$key = "new@test.com"
|> person@email`)
	if got != "new@test.com" {
		t.Fatalf("expected %q, got %q", "new@test.com", got)
	}
}

func TestE2E_KeySetDyn_InLoopBuildMap(t *testing.T) {
	got := runE2E(t, `result{} = []
keys = ["a", "b", "c"]
vals = [[1], [2], [3]]
i = [0]
for k in keys {
	result$k = vals$i
	i = i + [1]
}
|> str(##result)
|> str(@@result)`)
	if got != "[a, b, c][1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[a, b, c][1, 2, 3]", got)
	}
}

// --- Keys prefix ## (4 tests) ---

func TestE2E_MapKeys_Basic(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> str(##person)`)
	if got != "[name, age]" {
		t.Fatalf("expected %q, got %q", "[name, age]", got)
	}
}

func TestE2E_MapKeys_Empty(t *testing.T) {
	got := runE2E(t, `empty{} = []
|> str(##empty)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_MapKeys_IterateKeys(t *testing.T) {
	got := runE2E(t, `data{x, y} = [[10], [20]]
for k in ##data {
	|> k
}`)
	if got != "xy" {
		t.Fatalf("expected %q, got %q", "xy", got)
	}
}

func TestE2E_MapKeys_Length(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> str(#(##person))`)
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

// --- Values prefix @@ (3 tests) ---

func TestE2E_MapValues_Basic(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> str(@@person)`)
	if got != "[bob, [30]]" {
		t.Fatalf("expected %q, got %q", "[bob, [30]]", got)
	}
}

func TestE2E_MapValues_Empty(t *testing.T) {
	got := runE2E(t, `empty{} = []
|> str(@@empty)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestE2E_MapValues_ForIn(t *testing.T) {
	got := runE2E(t, `data{a, b, c} = [[1], [2], [3]]
for v in @@data {
	|> str(v)
}`)
	if got != "[1][2][3]" {
		t.Fatalf("expected %q, got %q", "[1][2][3]", got)
	}
}

// --- Combined / integration (7 tests) ---

func TestE2E_MapsSlopExample(t *testing.T) {
	src := `person{name, age} = ["bob", [30]]
|> person@name
|> str(person@age)

person@age = [31]
|> str(person@age)

person@email = "bob@test.com"
|> person@email

ks = ##person
|> str(ks)

vs = @@person
|> str(vs)`
	got := runE2E(t, src)
	expected := "bob[30][31]bob@test.com[name, age, email][bob, 31, bob@test.com]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

func TestE2E_HashInsideFn(t *testing.T) {
	got := runE2E(t, `fn makePoint(x, y) {
	pt{x, y} = [x, y]
	<- pt
}
p = makePoint([10], [20])
|> str(p)`)
	// makePoint returns the hashmap; str() on it shows values
	// Note: SlopValue with keys formats as an array of its elements
	if got != "[[10], [20]]" {
		t.Fatalf("expected %q, got %q", "[[10], [20]]", got)
	}
}

func TestE2E_FnReturnsHashCallerAccesses(t *testing.T) {
	got := runE2E(t, `fn makePerson(n, a) {
	p{name, age} = [n, a]
	<- p
}
person = makePerson("alice", [25])
|> person@name
|> str(person@age)`)
	if got != "alice[25]" {
		t.Fatalf("expected %q, got %q", "alice[25]", got)
	}
}

func TestE2E_CountPattern(t *testing.T) {
	got := runE2E(t, `counts{} = []
items = ["a", "b", "a", "c", "b", "a"]
for item in items {
	found = [0]
	for k in ##counts {
		if k == item {
			found = [1]
			counts$k = counts$k + [1]
		}
	}
	if found == [0] {
		counts$item = [1]
	}
}
|> str(counts@a)
|> str(counts@b)
|> str(counts@c)`)
	if got != "[3][2][1]" {
		t.Fatalf("expected %q, got %q", "[3][2][1]", got)
	}
}

func TestE2E_HashWithArrayValues(t *testing.T) {
	got := runE2E(t, `data{nums, labels} = [[1, 2, 3], ["x", "y", "z"]]
|> str(data@nums)
|> str(data@labels)`)
	if got != "[1, 2, 3][x, y, z]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3][x, y, z]", got)
	}
}

func TestE2E_KeyFromFnResultDynAccess(t *testing.T) {
	got := runE2E(t, `fn getKey() {
	<- "age"
}
person{name, age} = ["bob", [30]]
k = getKey()
|> str(person$k)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestE2E_HashMultipleOpsSequence(t *testing.T) {
	got := runE2E(t, `config{host, port} = ["localhost", [8080]]
|> config@host
config@port = [9090]
config@debug = [1]
|> str(config@port)
|> str(config@debug)
|> str(##config)
|> str(@@config)`)
	if got != "localhost[9090][1][host, port, debug][localhost, 9090, 1]" {
		t.Fatalf("expected %q, got %q", "localhost[9090][1][host, port, debug][localhost, 9090, 1]", got)
	}
}

// --- Edge cases (5 tests) ---

func TestE2E_HashKey_Underscore(t *testing.T) {
	got := runE2E(t, `data{my_key} = ["val"]
|> data@my_key`)
	if got != "val" {
		t.Fatalf("expected %q, got %q", "val", got)
	}
}

func TestE2E_HashSingleKeyOps(t *testing.T) {
	got := runE2E(t, `box{item} = [[5]]
|> str(box@item)
box@item = [10]
|> str(box@item)
|> str(##box)
|> str(@@box)`)
	// ## and @@ on single-key map return single-element SlopValue
	// str() on single-element → just the value, no brackets
	if got != "[5][10]item[10]" {
		t.Fatalf("expected %q, got %q", "[5][10]item[10]", got)
	}
}

func TestE2E_HashReassignVariable(t *testing.T) {
	// Use different variable names since re-declaring same name isn't supported
	got := runE2E(t, `m1{a} = [[1]]
|> str(m1@a)
m2{b} = [[2]]
|> str(m2@b)`)
	if got != "[1][2]" {
		t.Fatalf("expected %q, got %q", "[1][2]", got)
	}
}

func TestE2E_HashValueInArithmetic(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
|> str(person@age + [1])`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestE2E_HashValueComparison(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
if person@age > [18] {
	|> "adult"
} else {
	|> "minor"
}`)
	if got != "adult" {
		t.Fatalf("expected %q, got %q", "adult", got)
	}
}

// ==========================================
// Null Value E2E Tests
// ==========================================

func TestE2E_NullAssignStr(t *testing.T) {
	got := runE2E(t, `x = [null]
|> str(x)`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestE2E_NullStdoutWrite(t *testing.T) {
	got := runE2E(t, `|> [null]`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestE2E_NullArrayFormat(t *testing.T) {
	got := runE2E(t, `x = [null, null, null]
|> str(x)`)
	if got != "[null, null, null]" {
		t.Fatalf("expected %q, got %q", "[null, null, null]", got)
	}
}

func TestE2E_NullEqNull(t *testing.T) {
	got := runE2E(t, `if [null] == [null] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_NullNeqValue(t *testing.T) {
	got := runE2E(t, `if [null] != [5] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestE2E_ValueEqNull(t *testing.T) {
	got := runE2E(t, `if [1] == [null] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestE2E_NullLength(t *testing.T) {
	got := runE2E(t, `|> str(#[null, null])`)
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestE2E_NullContains(t *testing.T) {
	got := runE2E(t, `if [1, null] ?? [null] { |> "found" }`)
	if got != "found" {
		t.Fatalf("expected %q, got %q", "found", got)
	}
}

// --- Null panic tests ---

func TestE2E_NullAddPanic(t *testing.T) {
	runE2EExpectPanic(t, `x = [null] + [1]`)
}

func TestE2E_NullNegatePanic(t *testing.T) {
	runE2EExpectPanic(t, `x = -[null]`)
}

func TestE2E_NullIfPanic(t *testing.T) {
	runE2EExpectPanic(t, `if [null] { |> "nope" }`)
}

func TestE2E_NullNotPanic(t *testing.T) {
	runE2EExpectPanic(t, `x = ![null]`)
}

func TestE2E_NullGtPanic(t *testing.T) {
	runE2EExpectPanic(t, `x = [null] > [1]`)
}

func TestE2E_NullIteratePanic(t *testing.T) {
	runE2EExpectPanic(t, `for x in [null] { |> "nope" }`)
}

// --- Null non-panic edge cases ---

func TestE2E_NullInHashmapValue(t *testing.T) {
	got := runE2E(t, `m{a} = [null]
|> str(m@a)`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestE2E_NullForwardDecl(t *testing.T) {
	got := runE2E(t, `x = [null]
x = [42]
|> str(x)`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

// --- Phase 6: I/O tests ---

func TestE2E_FileWriteRead(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "hello"
data, err = <. "test.txt"
|> data`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_FileAppend(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "first"
.>> "test.txt" " second"
data, err = <. "test.txt"
|> data`)
	if got != "first second" {
		t.Fatalf("expected %q, got %q", "first second", got)
	}
}

func TestE2E_FileReadMissing(t *testing.T) {
	got := runE2E(t, `data, err = <. "nofile.txt"
|> str(err)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_FileWriteOverwrite(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "first"
.> "test.txt" "second"
data, err = <. "test.txt"
|> data`)
	if got != "second" {
		t.Fatalf("expected %q, got %q", "second", got)
	}
}

func TestE2E_FileWriteVariable(t *testing.T) {
	got := runE2E(t, `path = "myfile.txt"
.> path "variable path"
data, err = <. path
|> data`)
	if got != "variable path" {
		t.Fatalf("expected %q, got %q", "variable path", got)
	}
}

func TestE2E_StdinRead(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
|> line`, "hello\n")
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_StdinReadEOF(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
|> str(err)`, "")
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_SplitBySpace(t *testing.T) {
	got := runE2E(t, `words = split("a b c", " ")
|> str(words)`)
	if got != "[a, b, c]" {
		t.Fatalf("expected %q, got %q", "[a, b, c]", got)
	}
}

func TestE2E_SplitByNewline(t *testing.T) {
	got := runE2E(t, `lines = split("a\nb", "\n")
|> str(lines)`)
	if got != "[a, b]" {
		t.Fatalf("expected %q, got %q", "[a, b]", got)
	}
}

func TestE2E_SplitEmptySep(t *testing.T) {
	got := runE2E(t, `x = split("hello", "")
|> str(x)`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_ToNumInt(t *testing.T) {
	got := runE2E(t, `val, err = to_num("42")
|> str(val)`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestE2E_ToNumFloat(t *testing.T) {
	got := runE2E(t, `val, err = to_num("3.14")
|> str(val)`)
	if got != "[3.14]" {
		t.Fatalf("expected %q, got %q", "[3.14]", got)
	}
}

func TestE2E_ToNumFail(t *testing.T) {
	got := runE2E(t, `val, err = to_num("abc")
|> str(err)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_WriteSplitToNum(t *testing.T) {
	got := runE2E(t, `.> "nums.txt" "10 20 30"
data, err = <. "nums.txt"
parts = split(data, " ")
sum = [0]
for p in parts {
	val, verr = to_num(p)
	sum = sum + val
}
|> str(sum)`)
	if got != "[60]" {
		t.Fatalf("expected %q, got %q", "[60]", got)
	}
}

// --- Phase 6: Edge case / boundary tests ---

// File I/O edge cases

func TestE2E_FileWriteEmptyString(t *testing.T) {
	got := runE2E(t, `.> "test.txt" ""
data, err = <. "test.txt"
|> str(#data)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_FileWriteEmptyStringContent(t *testing.T) {
	got := runE2E(t, `.> "test.txt" ""
data, err = <. "test.txt"
|> data
|> "done"`)
	if got != "done" {
		t.Fatalf("expected %q, got %q", "done", got)
	}
}

func TestE2E_FileReadSuccessErrIsZero(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "x"
data, err = <. "test.txt"
|> str(err)`)
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_FileReadMissingDataIsEmpty(t *testing.T) {
	got := runE2E(t, `data, err = <. "nope.txt"
|> data
|> "end"`)
	if got != "end" {
		t.Fatalf("expected %q, got %q", "end", got)
	}
}

func TestE2E_FileAppendCreatesFile(t *testing.T) {
	got := runE2E(t, `.>> "new.txt" "created"
data, err = <. "new.txt"
|> data`)
	if got != "created" {
		t.Fatalf("expected %q, got %q", "created", got)
	}
}

func TestE2E_FileAppendMultiple(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "a"
.>> "test.txt" "b"
.>> "test.txt" "c"
data, err = <. "test.txt"
|> data`)
	if got != "abc" {
		t.Fatalf("expected %q, got %q", "abc", got)
	}
}

func TestE2E_FileWriteNewlineContent(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "line1\nline2\nline3"
data, err = <. "test.txt"
lines = split(data, "\n")
|> str(#lines)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_FileWriteArrayFormatted(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
.> "test.txt" str(arr)
data, err = <. "test.txt"
|> data`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestE2E_FileWriteNumericFormatted(t *testing.T) {
	got := runE2E(t, `x = [42]
.> "test.txt" str(x)
data, err = <. "test.txt"
|> data`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestE2E_FileReadInIf(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "exists"
data, err = <. "test.txt"
if err == [0] {
	|> data
} else {
	|> "fail"
}`)
	if got != "exists" {
		t.Fatalf("expected %q, got %q", "exists", got)
	}
}

func TestE2E_FileReadMissingInIf(t *testing.T) {
	got := runE2E(t, `data, err = <. "nope.txt"
if err != [0] {
	|> "not found"
} else {
	|> data
}`)
	if got != "not found" {
		t.Fatalf("expected %q, got %q", "not found", got)
	}
}

func TestE2E_FileIOInFunction(t *testing.T) {
	got := runE2E(t, `fn save(path, content) {
	.> path content
	<- [0]
}
fn load(path) {
	data, err = <. path
	<- data
}
save("test.txt", "from fn")
result = load("test.txt")
|> result`)
	if got != "from fn" {
		t.Fatalf("expected %q, got %q", "from fn", got)
	}
}

func TestE2E_FileIOInLoop(t *testing.T) {
	got := runE2E(t, `names = ["a.txt", "b.txt", "c.txt"]
for name in names {
	.> name name
}
result = []
for name in names {
	data, err = <. name
	result << data
}
|> str(result)`)
	if got != "[a.txt, b.txt, c.txt]" {
		t.Fatalf("expected %q, got %q", "[a.txt, b.txt, c.txt]", got)
	}
}

func TestE2E_FileWritePathFromFnResult(t *testing.T) {
	got := runE2E(t, `fn getpath() {
	<- "dynamic.txt"
}
.> getpath() "content"
data, err = <. getpath()
|> data`)
	if got != "content" {
		t.Fatalf("expected %q, got %q", "content", got)
	}
}

func TestE2E_FileReadThenSplitLines(t *testing.T) {
	got := runE2E(t, `.> "multi.txt" "alpha\nbeta\ngamma"
data, err = <. "multi.txt"
lines = split(data, "\n")
for line in lines {
	|> line
}`)
	if got != "alphabetagamma" {
		t.Fatalf("expected %q, got %q", "alphabetagamma", got)
	}
}

// Stdin edge cases

func TestE2E_StdinReadMultipleLines(t *testing.T) {
	got := runE2EWithStdin(t, `line1, err1 = <|
line2, err2 = <|
|> line1
|> line2`, "first\nsecond\n")
	if got != "firstsecond" {
		t.Fatalf("expected %q, got %q", "firstsecond", got)
	}
}

func TestE2E_StdinReadThenEOF(t *testing.T) {
	got := runE2EWithStdin(t, `line1, err1 = <|
line2, err2 = <|
|> str(err1)
|> str(err2)`, "only\n")
	if got != "[0][1]" {
		t.Fatalf("expected %q, got %q", "[0][1]", got)
	}
}

func TestE2E_StdinReadEmptyLine(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
|> str(#line)
|> str(err)`, "\n")
	if got != "[1][0]" {
		t.Fatalf("expected %q, got %q", "[1][0]", got)
	}
}

func TestE2E_StdinReadInLoop(t *testing.T) {
	got := runE2EWithStdin(t, `result = []
count = [0]
for {
	line, err = <|
	if err != [0] {
		break
	}
	result << line
	count = count + [1]
}
|> str(count)
|> str(result)`, "a\nb\nc\n")
	if got != "[3][a, b, c]" {
		t.Fatalf("expected %q, got %q", "[3][a, b, c]", got)
	}
}

func TestE2E_StdinReadWithWhitespace(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
|> line`, "  spaces  \n")
	if got != "  spaces  " {
		t.Fatalf("expected %q, got %q", "  spaces  ", got)
	}
}

// Split edge cases

func TestE2E_SplitSingleChar(t *testing.T) {
	got := runE2E(t, `parts = split("a,b,c", ",")
|> str(parts)`)
	if got != "[a, b, c]" {
		t.Fatalf("expected %q, got %q", "[a, b, c]", got)
	}
}

func TestE2E_SplitNoMatch(t *testing.T) {
	got := runE2E(t, `parts = split("hello", ",")
|> str(parts)`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestE2E_SplitTrailingSep(t *testing.T) {
	got := runE2E(t, `parts = split("a,b,", ",")
|> str(#parts)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_SplitLeadingSep(t *testing.T) {
	got := runE2E(t, `parts = split(",a,b", ",")
|> str(#parts)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_SplitConsecutiveSeps(t *testing.T) {
	got := runE2E(t, `parts = split("a,,b", ",")
|> str(#parts)
|> str(parts@0)
|> str(parts@2)`)
	if got != "[3]ab" {
		t.Fatalf("expected %q, got %q", "[3]ab", got)
	}
}

func TestE2E_SplitMultiCharSep(t *testing.T) {
	got := runE2E(t, `parts = split("a::b::c", "::")
|> str(parts)`)
	if got != "[a, b, c]" {
		t.Fatalf("expected %q, got %q", "[a, b, c]", got)
	}
}

func TestE2E_SplitResultIterate(t *testing.T) {
	got := runE2E(t, `parts = split("x-y-z", "-")
for p in parts {
	|> p
}`)
	if got != "xyz" {
		t.Fatalf("expected %q, got %q", "xyz", got)
	}
}

func TestE2E_SplitResultIndex(t *testing.T) {
	got := runE2E(t, `parts = split("a:b:c", ":")
|> parts@0
|> parts@2`)
	if got != "ac" {
		t.Fatalf("expected %q, got %q", "ac", got)
	}
}

func TestE2E_SplitResultLength(t *testing.T) {
	got := runE2E(t, `parts = split("one two three four", " ")
|> str(#parts)`)
	if got != "[4]" {
		t.Fatalf("expected %q, got %q", "[4]", got)
	}
}

func TestE2E_SplitThenContains(t *testing.T) {
	got := runE2E(t, `parts = split("cat,dog,bird", ",")
if parts ?? "dog" {
	|> "found"
} else {
	|> "nope"
}`)
	if got != "found" {
		t.Fatalf("expected %q, got %q", "found", got)
	}
}

func TestE2E_SplitFromVariable(t *testing.T) {
	got := runE2E(t, `s = "1-2-3"
sep = "-"
parts = split(s, sep)
|> str(parts)`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

// to_num edge cases

func TestE2E_ToNumNegativeInt(t *testing.T) {
	got := runE2E(t, `val, err = to_num("-99")
|> str(val)
|> str(err)`)
	if got != "[-99][0]" {
		t.Fatalf("expected %q, got %q", "[-99][0]", got)
	}
}

func TestE2E_ToNumZero(t *testing.T) {
	got := runE2E(t, `val, err = to_num("0")
|> str(val)
|> str(err)`)
	if got != "[0][0]" {
		t.Fatalf("expected %q, got %q", "[0][0]", got)
	}
}

func TestE2E_ToNumNegativeFloat(t *testing.T) {
	got := runE2E(t, `val, err = to_num("-2.5")
|> str(val)
|> str(err)`)
	if got != "[-2.5][0]" {
		t.Fatalf("expected %q, got %q", "[-2.5][0]", got)
	}
}

func TestE2E_ToNumEmptyString(t *testing.T) {
	got := runE2E(t, `val, err = to_num("")
|> str(err)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ToNumWhitespace(t *testing.T) {
	got := runE2E(t, `val, err = to_num(" 42 ")
|> str(err)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ToNumLargeInt(t *testing.T) {
	got := runE2E(t, `val, err = to_num("9999999999")
|> str(val)`)
	if got != "[9999999999]" {
		t.Fatalf("expected %q, got %q", "[9999999999]", got)
	}
}

func TestE2E_ToNumScientificNotation(t *testing.T) {
	got := runE2E(t, `val, err = to_num("1e3")
|> str(val)
|> str(err)`)
	if got != "[1000][0]" {
		t.Fatalf("expected %q, got %q", "[1000][0]", got)
	}
}

func TestE2E_ToNumErrCheckBeforeUse(t *testing.T) {
	got := runE2E(t, `val, err = to_num("bad")
if err != [0] {
	|> "parse error"
} else {
	|> str(val)
}`)
	if got != "parse error" {
		t.Fatalf("expected %q, got %q", "parse error", got)
	}
}

func TestE2E_ToNumInLoop(t *testing.T) {
	got := runE2E(t, `strs = ["10", "20", "30"]
sum = [0]
for s in strs {
	v, e = to_num(s)
	sum = sum + v
}
|> str(sum)`)
	if got != "[60]" {
		t.Fatalf("expected %q, got %q", "[60]", got)
	}
}

func TestE2E_ToNumIntPreferredOverFloat(t *testing.T) {
	// "42" should parse as int64, not float64
	got := runE2E(t, `val, err = to_num("42")
x = val + [8]
|> str(x)`)
	if got != "[50]" {
		t.Fatalf("expected %q, got %q", "[50]", got)
	}
}

// Combined pipeline edge cases

func TestE2E_StdinSplitToNum(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
parts = split(line, " ")
sum = [0]
for p in parts {
	v, e = to_num(p)
	sum = sum + v
}
|> str(sum)`, "5 10 15\n")
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestE2E_ReadSplitFilterWrite(t *testing.T) {
	got := runE2E(t, `.> "input.txt" "1,2,3,4,5"
data, err = <. "input.txt"
parts = split(data, ",")
evens = []
for p in parts {
	v, e = to_num(p)
	if v % [2] == [0] {
		evens << p
	}
}
|> str(evens)`)
	if got != "[2, 4]" {
		t.Fatalf("expected %q, got %q", "[2, 4]", got)
	}
}

func TestE2E_FileAppendInLoop(t *testing.T) {
	got := runE2E(t, `nums = [1, 2, 3]
.> "log.txt" ""
for n in nums {
	.>> "log.txt" str(n)
}
data, err = <. "log.txt"
|> data`)
	if got != "[1][2][3]" {
		t.Fatalf("expected %q, got %q", "[1][2][3]", got)
	}
}

func TestE2E_StdinReadToFileWriteRoundtrip(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
.> "captured.txt" line
data, rerr = <. "captured.txt"
|> data`, "captured input\n")
	if got != "captured input" {
		t.Fatalf("expected %q, got %q", "captured input", got)
	}
}

func TestE2E_SplitThenJoinViaAppend(t *testing.T) {
	got := runE2E(t, `parts = split("a-b-c", "-")
.> "joined.txt" ""
first = [1]
for p in parts {
	if first == [1] {
		.> "joined.txt" p
		first = [0]
	} else {
		.>> "joined.txt" ","
		.>> "joined.txt" p
	}
}
data, err = <. "joined.txt"
|> data`)
	if got != "a,b,c" {
		t.Fatalf("expected %q, got %q", "a,b,c", got)
	}
}

func TestE2E_ToNumThenArithmetic(t *testing.T) {
	got := runE2E(t, `a, e1 = to_num("10")
b, e2 = to_num("3")
|> str(a + b)
|> str(a - b)
|> str(a * b)
|> str(a / b)`)
	if got != "[13][7][30][3]" {
		t.Fatalf("expected %q, got %q", "[13][7][30][3]", got)
	}
}

func TestE2E_FileWriteSpecialChars(t *testing.T) {
	got := runE2E(t, `.> "test.txt" "tab\there\nnewline"
data, err = <. "test.txt"
|> data`)
	if got != "tab\there\nnewline" {
		t.Fatalf("expected %q, got %q", "tab\there\nnewline", got)
	}
}

// === Phase 6: Additional boundary and edge-case tests ===

// --- File I/O boundary tests ---

func TestE2E_FileWriteReadUnicode(t *testing.T) {
	got := runE2E(t, `.> "uni.txt" "hello 世界 🌍"
data, err = <. "uni.txt"
|> data`)
	if got != "hello 世界 🌍" {
		t.Fatalf("expected %q, got %q", "hello 世界 🌍", got)
	}
}

func TestE2E_FileWriteOnlyWhitespace(t *testing.T) {
	got := runE2E(t, `.> "ws.txt" "   \t\n  "
data, err = <. "ws.txt"
|> str(#data)`)
	// data is a single-element string, length is 1
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_FileMultipleRapidWrites(t *testing.T) {
	got := runE2E(t, `.> "f.txt" "one"
.> "f.txt" "two"
.> "f.txt" "three"
.> "f.txt" "four"
.> "f.txt" "five"
data, err = <. "f.txt"
|> data`)
	if got != "five" {
		t.Fatalf("expected %q, got %q", "five", got)
	}
}

func TestE2E_FileAppendNoInitialWrite(t *testing.T) {
	// .>> on a nonexistent file should create it
	got := runE2E(t, `.>> "fresh.txt" "line1"
.>> "fresh.txt" "line2"
data, err = <. "fresh.txt"
|> data`)
	if got != "line1line2" {
		t.Fatalf("expected %q, got %q", "line1line2", got)
	}
}

func TestE2E_FileReadSameFileTwice(t *testing.T) {
	got := runE2E(t, `.> "dup.txt" "content"
d1, e1 = <. "dup.txt"
d2, e2 = <. "dup.txt"
|> d1
|> d2
|> str(e1)
|> str(e2)`)
	if got != "contentcontent[0][0]" {
		t.Fatalf("expected %q, got %q", "contentcontent[0][0]", got)
	}
}

func TestE2E_FileReadAfterAppendModifiesContent(t *testing.T) {
	got := runE2E(t, `.> "rw.txt" "start"
d1, e1 = <. "rw.txt"
.>> "rw.txt" "+more"
d2, e2 = <. "rw.txt"
|> d1
|> d2`)
	if got != "startstart+more" {
		t.Fatalf("expected %q, got %q", "startstart+more", got)
	}
}

func TestE2E_FileWriteReadLongContent(t *testing.T) {
	// Build a long string via repeated concat
	got := runE2E(t, `s = "abcdefghij"
long = s
i = [0]
for {
	if i == [9] { break }
	long = long ++ s
	i = i + [1]
}
.> "long.txt" str(long)
data, err = <. "long.txt"
|> str(#data)`)
	// 1 element string, length is 1
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_FileReadMultipleMissing(t *testing.T) {
	got := runE2E(t, `d1, e1 = <. "nope1.txt"
d2, e2 = <. "nope2.txt"
d3, e3 = <. "nope3.txt"
|> str(e1)
|> str(e2)
|> str(e3)`)
	if got != "[1][1][1]" {
		t.Fatalf("expected %q, got %q", "[1][1][1]", got)
	}
}

func TestE2E_FileWriteEmptyThenAppend(t *testing.T) {
	got := runE2E(t, `.> "ea.txt" ""
.>> "ea.txt" "appended"
data, err = <. "ea.txt"
|> data`)
	if got != "appended" {
		t.Fatalf("expected %q, got %q", "appended", got)
	}
}

func TestE2E_FileAppendNewlineDelimited(t *testing.T) {
	got := runE2E(t, `.> "lines.txt" "line1"
.>> "lines.txt" "\nline2"
.>> "lines.txt" "\nline3"
data, err = <. "lines.txt"
parts = split(data, "\n")
|> str(#parts)
|> parts@0
|> parts@2`)
	if got != "[3]line1line3" {
		t.Fatalf("expected %q, got %q", "[3]line1line3", got)
	}
}

func TestE2E_FileReadErrThenSuccessWrite(t *testing.T) {
	// Read a missing file (err=1), then write it, then read again (err=0)
	got := runE2E(t, `d1, e1 = <. "late.txt"
.> "late.txt" "now exists"
d2, e2 = <. "late.txt"
|> str(e1)
|> str(e2)
|> d2`)
	if got != "[1][0]now exists" {
		t.Fatalf("expected %q, got %q", "[1][0]now exists", got)
	}
}

func TestE2E_FileWritePathFromConcat(t *testing.T) {
	got := runE2E(t, `base = "data"
ext = ".txt"
path = base ++ ext
.> str(path) "concat path"
data, err = <. str(path)
|> data`)
	if got != "concat path" {
		t.Fatalf("expected %q, got %q", "concat path", got)
	}
}

func TestE2E_FileIOPreservesExactContent(t *testing.T) {
	// Verify no trailing newline is added by file write
	got := runE2E(t, `.> "exact.txt" "no-newline"
data, err = <. "exact.txt"
if data == "no-newline" {
	|> "exact"
} else {
	|> "modified"
}`)
	if got != "exact" {
		t.Fatalf("expected %q, got %q", "exact", got)
	}
}

// --- Stdin boundary tests ---

func TestE2E_StdinReadNoTrailingNewline(t *testing.T) {
	// Input without trailing newline — scanner should still read the line
	got := runE2EWithStdin(t, `line, err = <|
|> line
|> str(err)`, "no-newline")
	if got != "no-newline[0]" {
		t.Fatalf("expected %q, got %q", "no-newline[0]", got)
	}
}

func TestE2E_StdinReadUnicode(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
|> line`, "café résumé\n")
	if got != "café résumé" {
		t.Fatalf("expected %q, got %q", "café résumé", got)
	}
}

func TestE2E_StdinReadMultipleEOF(t *testing.T) {
	// After EOF, subsequent reads should also return err=1
	got := runE2EWithStdin(t, `l1, e1 = <|
l2, e2 = <|
l3, e3 = <|
|> str(e1)
|> str(e2)
|> str(e3)`, "only\n")
	if got != "[0][1][1]" {
		t.Fatalf("expected %q, got %q", "[0][1][1]", got)
	}
}

func TestE2E_StdinReadLongLine(t *testing.T) {
	// 500 chars is well within scanner buffer
	long := ""
	for i := 0; i < 500; i++ {
		long += "x"
	}
	got := runE2EWithStdin(t, `line, err = <|
|> str(#line)`, long+"\n")
	if got != "[1]" {
		// line is a single-element string, #line = 1
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_StdinReadTabsAndSpaces(t *testing.T) {
	got := runE2EWithStdin(t, `line, err = <|
|> line`, "\t  \t  \n")
	if got != "\t  \t  " {
		t.Fatalf("expected %q, got %q", "\t  \t  ", got)
	}
}

func TestE2E_StdinReadExhaustedThenFileRead(t *testing.T) {
	// After stdin is exhausted, file operations should still work
	got := runE2EWithStdin(t, `l1, e1 = <|
l2, e2 = <|
.> "out.txt" l1
data, rerr = <. "out.txt"
|> data
|> str(e2)`, "hello\n")
	if got != "hello[1]" {
		t.Fatalf("expected %q, got %q", "hello[1]", got)
	}
}

func TestE2E_StdinReadEmptyInput(t *testing.T) {
	// Completely empty stdin — immediate EOF
	got := runE2EWithStdin(t, `line, err = <|
|> str(err)
|> line
|> "end"`, "")
	if got != "[1]end" {
		t.Fatalf("expected %q, got %q", "[1]end", got)
	}
}

func TestE2E_StdinReadManyLines(t *testing.T) {
	got := runE2EWithStdin(t, `count = [0]
for {
	line, err = <|
	if err != [0] { break }
	count = count + [1]
}
|> str(count)`, "a\nb\nc\nd\ne\nf\ng\nh\ni\nj\n")
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

// --- Split boundary tests ---

func TestE2E_SplitEmptyString(t *testing.T) {
	// Splitting empty string by any separator
	// strings.Split("", ",") returns [""] — 1 element (empty string)
	got := runE2E(t, `parts = split("", ",")
|> str(#parts)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_SplitSepEqualsString(t *testing.T) {
	// Separator equals the entire string
	got := runE2E(t, `parts = split("abc", "abc")
|> str(#parts)`)
	// strings.Split("abc", "abc") = ["", ""] — 2 empty strings
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestE2E_SplitSingleCharString(t *testing.T) {
	got := runE2E(t, `parts = split("x", ",")
|> str(parts)`)
	// No match, returns ["x"]
	if got != "x" {
		t.Fatalf("expected %q, got %q", "x", got)
	}
}

func TestE2E_SplitOnEveryChar(t *testing.T) {
	// Using single-char separator on string made of that char
	got := runE2E(t, `parts = split("---", "-")
|> str(#parts)`)
	// strings.Split("---", "-") = ["", "", "", ""] — 4 empty strings
	if got != "[4]" {
		t.Fatalf("expected %q, got %q", "[4]", got)
	}
}

func TestE2E_SplitLongSeparator(t *testing.T) {
	got := runE2E(t, `parts = split("helloworldhello", "world")
|> str(parts)`)
	if got != "[hello, hello]" {
		t.Fatalf("expected %q, got %q", "[hello, hello]", got)
	}
}

func TestE2E_SplitWhitespaceSep(t *testing.T) {
	// Tab separator
	got := runE2E(t, `parts = split("a\tb\tc", "\t")
|> str(parts)`)
	if got != "[a, b, c]" {
		t.Fatalf("expected %q, got %q", "[a, b, c]", got)
	}
}

func TestE2E_SplitResultArithmetic(t *testing.T) {
	// Split then to_num each part and do arithmetic
	got := runE2E(t, `parts = split("100,200,300", ",")
v1, e1 = to_num(parts@0)
v2, e2 = to_num(parts@1)
v3, e3 = to_num(parts@2)
|> str(v1 + v2 + v3)`)
	if got != "[600]" {
		t.Fatalf("expected %q, got %q", "[600]", got)
	}
}

func TestE2E_SplitThenIndex(t *testing.T) {
	got := runE2E(t, `parts = split("zero,one,two,three,four", ",")
|> parts@0
|> parts@4
|> str(#parts)`)
	if got != "zerofour[5]" {
		t.Fatalf("expected %q, got %q", "zerofour[5]", got)
	}
}

func TestE2E_SplitNewlineOnly(t *testing.T) {
	// String that is just newlines
	got := runE2E(t, `parts = split("\n\n", "\n")
|> str(#parts)`)
	// strings.Split("\n\n", "\n") = ["", "", ""] — 3 elements
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

// --- to_num boundary tests ---

func TestE2E_ToNumPlusSign(t *testing.T) {
	// "+42" — ParseInt handles this
	got := runE2E(t, `val, err = to_num("+42")
|> str(val)
|> str(err)`)
	if got != "[42][0]" {
		t.Fatalf("expected %q, got %q", "[42][0]", got)
	}
}

func TestE2E_ToNumHexString(t *testing.T) {
	// "0x10" — ParseInt base 10 rejects this, ParseFloat also rejects
	got := runE2E(t, `val, err = to_num("0x10")
|> str(err)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ToNumLeadingZeros(t *testing.T) {
	got := runE2E(t, `val, err = to_num("007")
|> str(val)
|> str(err)`)
	if got != "[7][0]" {
		t.Fatalf("expected %q, got %q", "[7][0]", got)
	}
}

func TestE2E_ToNumDecimalOnly(t *testing.T) {
	// ".5" — ParseInt rejects, ParseFloat accepts
	got := runE2E(t, `val, err = to_num(".5")
|> str(val)
|> str(err)`)
	if got != "[0.5][0]" {
		t.Fatalf("expected %q, got %q", "[0.5][0]", got)
	}
}

func TestE2E_ToNumTrailingDecimal(t *testing.T) {
	// "5." — ParseInt rejects, ParseFloat accepts as 5.0
	got := runE2E(t, `val, err = to_num("5.")
|> str(val)
|> str(err)`)
	if got != "[5][0]" {
		t.Fatalf("expected %q, got %q", "[5][0]", got)
	}
}

func TestE2E_ToNumJustDot(t *testing.T) {
	got := runE2E(t, `val, err = to_num(".")
|> str(err)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_ToNumSpecialStrings(t *testing.T) {
	// "NaN", "Inf" — these should fail since they're not standard numbers
	// Actually ParseFloat accepts them, so let's test that
	got := runE2E(t, `v1, e1 = to_num("NaN")
v2, e2 = to_num("Inf")
v3, e3 = to_num("-Inf")
|> str(e1)
|> str(e2)
|> str(e3)`)
	// ParseFloat accepts NaN, +Inf, -Inf — all succeed with err=0
	if got != "[0][0][0]" {
		t.Fatalf("expected %q, got %q", "[0][0][0]", got)
	}
}

func TestE2E_ToNumVeryLargeInt(t *testing.T) {
	// Near int64 max: 9223372036854775807
	got := runE2E(t, `val, err = to_num("9223372036854775807")
|> str(err)`)
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_ToNumOverflowInt(t *testing.T) {
	// int64 overflow, but ParseFloat can still handle it
	got := runE2E(t, `val, err = to_num("9223372036854775808")
|> str(err)`)
	// ParseInt fails, ParseFloat succeeds
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_ToNumSmallNegativeFloat(t *testing.T) {
	got := runE2E(t, `val, err = to_num("-0.001")
|> str(val)
|> str(err)`)
	if got != "[-0.001][0]" {
		t.Fatalf("expected %q, got %q", "[-0.001][0]", got)
	}
}

func TestE2E_ToNumMixedValidInvalid(t *testing.T) {
	// Loop through mixed valid/invalid, count successes and failures
	got := runE2E(t, `inputs = ["1", "two", "3", "four", "5"]
ok = [0]
fail = [0]
for s in inputs {
	v, e = to_num(s)
	if e == [0] {
		ok = ok + [1]
	} else {
		fail = fail + [1]
	}
}
|> str(ok)
|> str(fail)`)
	if got != "[3][2]" {
		t.Fatalf("expected %q, got %q", "[3][2]", got)
	}
}

func TestE2E_ToNumFloatPrecision(t *testing.T) {
	got := runE2E(t, `val, err = to_num("0.1")
result = val + val + val
|> str(err)`)
	// Just verify it parses successfully
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestE2E_ToNumThenCompare(t *testing.T) {
	got := runE2E(t, `a, e1 = to_num("10")
b, e2 = to_num("20")
if a < b {
	|> "less"
} else {
	|> "not less"
}`)
	if got != "less" {
		t.Fatalf("expected %q, got %q", "less", got)
	}
}

// --- Combined pipeline boundary tests ---

func TestE2E_ReadMissingSkipProcessing(t *testing.T) {
	// Read missing file, check err, skip processing
	got := runE2E(t, `data, err = <. "missing.txt"
if err != [0] {
	|> "skipped"
} else {
	parts = split(data, ",")
	|> str(parts)
}`)
	if got != "skipped" {
		t.Fatalf("expected %q, got %q", "skipped", got)
	}
}

func TestE2E_StdinSplitFilterToNum(t *testing.T) {
	// Read stdin, split, filter only numeric, sum
	got := runE2EWithStdin(t, `line, err = <|
parts = split(line, ",")
sum = [0]
for p in parts {
	v, e = to_num(p)
	if e == [0] {
		sum = sum + v
	}
}
|> str(sum)`, "10,abc,20,def,30\n")
	if got != "[60]" {
		t.Fatalf("expected %q, got %q", "[60]", got)
	}
}

func TestE2E_FileRoundtripNumeric(t *testing.T) {
	// str([42]) now returns "[42]" which to_num cannot parse,
	// so val is empty and val * [2] panics on length mismatch.
	runE2EExpectPanic(t, `x = [42]
.> "num.txt" str(x)
data, err = <. "num.txt"
val, verr = to_num(data)
result = val * [2]
|> str(result)`)
}

func TestE2E_FileSplitToNumSum(t *testing.T) {
	// Write CSV, read, split, convert each, sum
	got := runE2E(t, `.> "csv.txt" "5,10,15,20,25"
data, err = <. "csv.txt"
parts = split(data, ",")
sum = [0]
for p in parts {
	v, e = to_num(p)
	sum = sum + v
}
|> str(sum)`)
	if got != "[75]" {
		t.Fatalf("expected %q, got %q", "[75]", got)
	}
}

func TestE2E_StdinToFileToStdout(t *testing.T) {
	// Full pipeline: stdin → file → read → stdout
	got := runE2EWithStdin(t, `line, err = <|
.> "pipe.txt" line
data, rerr = <. "pipe.txt"
|> data`, "piped data\n")
	if got != "piped data" {
		t.Fatalf("expected %q, got %q", "piped data", got)
	}
}

func TestE2E_SplitToNumWriteResults(t *testing.T) {
	// Split, convert, write each result to a file
	got := runE2E(t, `parts = split("1,2,3", ",")
.> "results.txt" ""
for p in parts {
	v, e = to_num(p)
	doubled = v * [2]
	.>> "results.txt" str(doubled)
	.>> "results.txt" ","
}
data, err = <. "results.txt"
|> data`)
	if got != "[2],[4],[6]," {
		t.Fatalf("expected %q, got %q", "[2],[4],[6],", got)
	}
}

func TestE2E_MultipleStdinLinesSplitEach(t *testing.T) {
	got := runE2EWithStdin(t, `total = [0]
for {
	line, err = <|
	if err != [0] { break }
	parts = split(line, " ")
	for p in parts {
		v, e = to_num(p)
		if e == [0] {
			total = total + v
		}
	}
}
|> str(total)`, "1 2 3\n4 5 6\n")
	if got != "[21]" {
		t.Fatalf("expected %q, got %q", "[21]", got)
	}
}

func TestE2E_FileAppendFromStdinLoop(t *testing.T) {
	got := runE2EWithStdin(t, `.> "log.txt" ""
for {
	line, err = <|
	if err != [0] { break }
	.>> "log.txt" line
	.>> "log.txt" "\n"
}
data, rerr = <. "log.txt"
lines = split(data, "\n")
// Last element is empty due to trailing \n
|> str(#lines - [1])`, "alpha\nbeta\ngamma\n")
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestE2E_ChainedFileOperations(t *testing.T) {
	// Write to file A, read A, transform, write to file B, read B
	got := runE2E(t, `.> "a.txt" "10,20,30"
data, e1 = <. "a.txt"
parts = split(data, ",")
result = []
for p in parts {
	v, e = to_num(p)
	result << v + [1]
}
.> "b.txt" str(result)
data2, e2 = <. "b.txt"
|> data2`)
	if got != "[11, 21, 31]" {
		t.Fatalf("expected %q, got %q", "[11, 21, 31]", got)
	}
}

func TestE2E_ToNumFailValIsEmpty(t *testing.T) {
	// When to_num fails, val should have 0 elements
	got := runE2E(t, `val, err = to_num("not_a_number")
|> str(#val)
|> str(err)`)
	if got != "[0][1]" {
		t.Fatalf("expected %q, got %q", "[0][1]", got)
	}
}

func TestE2E_SplitThenAppendToArray(t *testing.T) {
	got := runE2E(t, `parts = split("a:b:c", ":")
all = []
for p in parts {
	all << p
}
all << "d"
|> str(all)`)
	if got != "[a, b, c, d]" {
		t.Fatalf("expected %q, got %q", "[a, b, c, d]", got)
	}
}

func TestE2E_FileWriteReadInConditionalBranches(t *testing.T) {
	got := runE2E(t, `flag = [1]
if flag == [1] {
	.> "cond.txt" "branch-true"
} else {
	.> "cond.txt" "branch-false"
}
data, err = <. "cond.txt"
|> data`)
	if got != "branch-true" {
		t.Fatalf("expected %q, got %q", "branch-true", got)
	}
}

func TestE2E_SplitEmptySepIdentity(t *testing.T) {
	// Empty separator with various strings returns original
	got := runE2E(t, `a = split("abc", "")
b = split("", "")
c = split("x y z", "")
|> str(a)
|> str(b)
|> str(c)`)
	if got != "abcx y z" {
		t.Fatalf("expected %q, got %q", "abcx y z", got)
	}
}

func TestE2E_StdinReadInFunction(t *testing.T) {
	got := runE2EWithStdin(t, `fn readLine() {
	line, err = <|
	<- line
}
result = readLine()
|> result`, "from-fn\n")
	if got != "from-fn" {
		t.Fatalf("expected %q, got %q", "from-fn", got)
	}
}

func TestE2E_ToNumZeroFloat(t *testing.T) {
	got := runE2E(t, `val, err = to_num("0.0")
|> str(val)
|> str(err)`)
	if got != "[0][0]" {
		t.Fatalf("expected %q, got %q", "[0][0]", got)
	}
}

func TestE2E_FileWriteHashmapFormatted(t *testing.T) {
	got := runE2E(t, `m{name, age} = ["alice", [30]]
.> "map.txt" str(@@m)
data, err = <. "map.txt"
|> data`)
	if got != "[alice, [30]]" {
		t.Fatalf("expected %q, got %q", "[alice, [30]]", got)
	}
}

func TestE2E_FileReadWriteWithHashmap(t *testing.T) {
	got := runE2E(t, `config{host, port} = ["localhost", "8080"]
.> "config.txt" config@host
.>> "config.txt" ":"
.>> "config.txt" config@port
data, err = <. "config.txt"
|> data`)
	if got != "localhost:8080" {
		t.Fatalf("expected %q, got %q", "localhost:8080", got)
	}
}
