package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

// ============================================================
// Semantic Control Flow E2E Tests
// Covers if/else, for-in, infinite loop + break, functions,
// variable scoping, and combined patterns
// ============================================================

// --- If/else (~7 tests) ---

func TestSem_Flow_TrueBranch(t *testing.T) {
	got := runE2E(t, `if true { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Flow_ElseBranch(t *testing.T) {
	got := runE2E(t, `if false { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Flow_NestedIf(t *testing.T) {
	got := runE2E(t, `if true { if true { |> "deep" } }`)
	if got != "deep" {
		t.Fatalf("expected %q, got %q", "deep", got)
	}
}

func TestSem_Flow_WithComparison(t *testing.T) {
	got := runE2E(t, `if [1] > [0] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Flow_WithLogical(t *testing.T) {
	got := runE2E(t, `if [1] && [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Flow_WithNot(t *testing.T) {
	got := runE2E(t, `if !false { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Flow_ChainedElse(t *testing.T) {
	got := runE2E(t, `if false { } else { if true { |> "second" } }`)
	if got != "second" {
		t.Fatalf("expected %q, got %q", "second", got)
	}
}

// --- For-in (~6 tests) ---

func TestSem_Flow_ForIn_Basic(t *testing.T) {
	got := runE2E(t, `for x in [1, 2, 3] { |> str(x) }`)
	if got != "[1][2][3]" {
		t.Fatalf("expected %q, got %q", "[1][2][3]", got)
	}
}

func TestSem_Flow_ForIn_EmptyArray(t *testing.T) {
	got := runE2E(t, `for x in [] { |> "never" }`)
	if got != "" {
		t.Fatalf("expected %q, got %q", "", got)
	}
}

func TestSem_Flow_ForIn_Strings(t *testing.T) {
	got := runE2E(t, `for x in ["a", "b"] { |> x }`)
	if got != "ab" {
		t.Fatalf("expected %q, got %q", "ab", got)
	}
}

func TestSem_Flow_ForIn_Nested(t *testing.T) {
	src := `for i in [1, 2] {
	for j in [10, 20] {
		|> str(i)
		|> str(j)
	}
}`
	got := runE2E(t, src)
	if got != "[1][10][1][20][2][10][2][20]" {
		t.Fatalf("expected %q, got %q", "[1][10][1][20][2][10][2][20]", got)
	}
}

func TestSem_Flow_ForIn_WithBreak(t *testing.T) {
	src := `for x in [1, 2, 3, 4] {
	if x == [3] { break }
	|> str(x)
}`
	got := runE2E(t, src)
	if got != "[1][2]" {
		t.Fatalf("expected %q, got %q", "[1][2]", got)
	}
}

func TestSem_Flow_ForIn_WithMutation(t *testing.T) {
	src := `result = []
for x in [10, 20, 30] {
	result << x
}
|> str(result)`
	got := runE2E(t, src)
	if got != "[10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30]", got)
	}
}

// --- Infinite loop + break (~3 tests) ---

func TestSem_Flow_InfLoop_BasicCounter(t *testing.T) {
	src := `i = [0]
for {
	if i == [5] { break }
	i = i + [1]
}
|> str(i)`
	got := runE2E(t, src)
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestSem_Flow_InfLoop_NestedBreak(t *testing.T) {
	// Outer iterates [1,2], inner counts to 3 then breaks
	src := `result = []
for outer in [1, 2] {
	count = [0]
	for {
		if count == [3] { break }
		count = count + [1]
	}
	result << outer
	result << count
}
|> str(result)`
	got := runE2E(t, src)
	if got != "[1, 3, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 3, 2, 3]", got)
	}
}

func TestSem_Flow_InfLoop_BreakAfterMutation(t *testing.T) {
	src := `arr = []
for {
	if #arr == [3] { break }
	arr << [1]
}
|> str(arr)`
	got := runE2E(t, src)
	if got != "[1, 1, 1]" {
		t.Fatalf("expected %q, got %q", "[1, 1, 1]", got)
	}
}

// --- Functions (~8 tests) ---

func TestSem_Flow_Fn_ReturnValue(t *testing.T) {
	src := `fn add(a, b) {
	<- a + b
}
|> str(add([1], [2]))`
	got := runE2E(t, src)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Flow_Fn_EarlyReturn(t *testing.T) {
	src := `fn f(x) {
	if x == [0] { <- [99] }
	<- x
}
|> str(f([0]))
|> str(f([5]))`
	got := runE2E(t, src)
	if got != "[99][5]" {
		t.Fatalf("expected %q, got %q", "[99][5]", got)
	}
}

func TestSem_Flow_Fn_Recursive(t *testing.T) {
	src := `fn fact(n) {
	if n == [0] { <- [1] }
	<- n * fact(n - [1])
}
|> str(fact([5]))`
	got := runE2E(t, src)
	if got != "[120]" {
		t.Fatalf("expected %q, got %q", "[120]", got)
	}
}

func TestSem_Flow_Fn_MutatesArray(t *testing.T) {
	src := `arr = [1, 2, 3]
fn push_val(a, v) {
	a << v
	<- a
}
push_val(arr, [4])
|> str(arr)`
	got := runE2E(t, src)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

func TestSem_Flow_Fn_LocalScope(t *testing.T) {
	// Function bodies have their own scope. x = [99] inside fn creates
	// a new local, doesn't modify outer x.
	// BUT: in Go codegen, fn is a top-level function — it can't even SEE
	// outer x. So we pass x as a param and verify the outer x is unchanged.
	src := `x = [1]
fn f(x) {
	x = [99]
	<- x
}
|> str(f([50]))
|> str(x)`
	got := runE2E(t, src)
	if got != "[99][1]" {
		t.Fatalf("expected %q, got %q", "[99][1]", got)
	}
}

func TestSem_Flow_Fn_MultiAssign(t *testing.T) {
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

func TestSem_Flow_Fn_NestedCalls(t *testing.T) {
	src := `fn double(x) {
	<- x * [2]
}
fn quad(x) {
	<- double(double(x))
}
|> str(quad([3]))`
	got := runE2E(t, src)
	if got != "[12]" {
		t.Fatalf("expected %q, got %q", "[12]", got)
	}
}

func TestSem_Flow_Fn_NoReturn(t *testing.T) {
	// A function with no explicit <- causes a Go compile error (missing return).
	// Verify the generated code fails to build.
	src := `fn noop() {
	|> "side"
}
noop()`
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	gen := New()
	output, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	sv, ops, io := loadRuntimeFiles(t)
	assembled, err := AssembleWithRuntime(output, sv, ops, io)
	if err != nil {
		t.Fatalf("AssembleWithRuntime: %v", err)
	}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module sloprun\n\ngo 1.24\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), assembled, 0644)
	buildCmd := exec.Command("go", "build", "-o", "prog", ".")
	buildCmd.Dir = tmpDir
	_, buildErr := buildCmd.CombinedOutput()
	if buildErr == nil {
		t.Fatal("expected compile error: fn without return should fail to build")
	}
}

// --- Variable scoping (~3 tests) ---

func TestSem_Flow_Var_Reassignment(t *testing.T) {
	src := `x = [1]
x = [2]
|> str(x)`
	got := runE2E(t, src)
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestSem_Flow_Var_ParamShadow(t *testing.T) {
	// Param x shadows outer x; outer x unmodified after call
	src := `x = [10]
fn f(x) {
	<- x + [1]
}
|> str(f([5]))
|> str(x)`
	got := runE2E(t, src)
	if got != "[6][10]" {
		t.Fatalf("expected %q, got %q", "[6][10]", got)
	}
}

func TestSem_Flow_Var_LoopVar(t *testing.T) {
	// Loop variable is scoped inside the for-range block in Go codegen,
	// so accessing it outside the loop causes a compile error.
	src := `for i in [1, 2, 3] { }
|> str(i)`
	l := lexer.New(src)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	gen := New()
	output, err := gen.Generate(prog)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	sv, ops, io := loadRuntimeFiles(t)
	assembled, err := AssembleWithRuntime(output, sv, ops, io)
	if err != nil {
		t.Fatalf("AssembleWithRuntime: %v", err)
	}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module sloprun\n\ngo 1.24\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "main.go"), assembled, 0644)
	buildCmd := exec.Command("go", "build", "-o", "prog", ".")
	buildCmd.Dir = tmpDir
	_, buildErr := buildCmd.CombinedOutput()
	if buildErr == nil {
		t.Fatal("expected compile error: loop variable should not be accessible outside loop")
	}
}

// --- Combined patterns (~5 tests) ---

func TestSem_Flow_Combined_BuildArrayInLoop(t *testing.T) {
	src := `result = []
for i in [1, 2, 3, 4, 5] {
	result << i
}
|> str(result)`
	got := runE2E(t, src)
	if got != "[1, 2, 3, 4, 5]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4, 5]", got)
	}
}

func TestSem_Flow_Combined_ErrorPropagation(t *testing.T) {
	src := `fn safe_div(a, b) {
	if b == [0] { <- [0, 1] }
	<- [a / b, 0]
}
result, err = safe_div([10], [0])
|> str(result)
|> str(err)`
	got := runE2E(t, src)
	if got != "[0][1]" {
		t.Fatalf("expected %q, got %q", "[0][1]", got)
	}
}

func TestSem_Flow_Combined_NestedFnCalls(t *testing.T) {
	src := `fn a(x) { <- x + [1] }
fn b(x) { <- a(x) + [1] }
fn c(x) { <- b(x) + [1] }
|> str(c([0]))`
	got := runE2E(t, src)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Flow_Combined_ArrayOpsInLoop(t *testing.T) {
	// Build array by setting elements via $ in a loop
	src := `arr = [0, 0, 0]
i = [0]
for x in [10, 20, 30] {
	arr$i = x
	i = i + [1]
}
|> str(arr)`
	got := runE2E(t, src)
	if got != "[10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30]", got)
	}
}

func TestSem_Flow_Combined_HashmapInLoop(t *testing.T) {
	// Build hashmap by setting keys in loop via $
	src := `m{} = []
keys = ["a", "b", "c"]
vals = [[10], [20], [30]]
i = [0]
for k in keys {
	m$k = vals$i
	i = i + [1]
}
|> str(##m)
|> str(@@m)`
	got := runE2E(t, src)
	if got != "[a, b, c][10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[a, b, c][10, 20, 30]", got)
	}
}
