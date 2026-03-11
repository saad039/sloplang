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
	// Top-level variables are hoisted to package-level var declarations,
	// so main() uses = (not :=).
	if !strings.Contains(out, "var x") {
		t.Fatalf("expected 'var x' package-level declaration, got:\n%s", out)
	}
	if !strings.Contains(out, "x =") {
		t.Fatalf("expected 'x =' assignment, got:\n%s", out)
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

func TestCodegen_FnDecl(t *testing.T) {
	out, err := generate("fn add(a, b) { <- a + b }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "func add(") {
		t.Fatalf("expected func add(, got:\n%s", out)
	}
	if !strings.Contains(out, "*sloprt.SlopValue") {
		t.Fatalf("expected *sloprt.SlopValue param type, got:\n%s", out)
	}
	if !strings.Contains(out, "return") {
		t.Fatalf("expected return, got:\n%s", out)
	}
}

func TestCodegen_IfStmt(t *testing.T) {
	out, err := generate("x = [1]\nif x { |> \"yes\" }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, ".IsTruthy()") {
		t.Fatalf("expected .IsTruthy(), got:\n%s", out)
	}
}

func TestCodegen_IfElseStmt(t *testing.T) {
	out, err := generate("x = []\nif x { |> \"yes\" } else { |> \"no\" }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "} else {") {
		t.Fatalf("expected else block, got:\n%s", out)
	}
}

func TestCodegen_ForIn(t *testing.T) {
	out, err := generate("items = [1, 2]\nfor x in items { |> str(x) }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Iterate(") {
		t.Fatalf("expected sloprt.Iterate, got:\n%s", out)
	}
	if !strings.Contains(out, "range") {
		t.Fatalf("expected range, got:\n%s", out)
	}
}

func TestCodegen_ReturnStmt(t *testing.T) {
	out, err := generate("fn foo() { <- [1] }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "return sloprt.NewSlopValue(") {
		t.Fatalf("expected return sloprt.NewSlopValue, got:\n%s", out)
	}
}

func TestCodegen_BareReturn(t *testing.T) {
	out, err := generate("fn foo() { <- }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "return sloprt.NewSlopValue()") {
		t.Fatalf("expected return sloprt.NewSlopValue(), got:\n%s", out)
	}
}

func TestCodegen_UserCall(t *testing.T) {
	out, err := generate("fn add(a, b) { <- a + b }\nx = add([1], [2])")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "add(") {
		t.Fatalf("expected user call add(, got:\n%s", out)
	}
	if strings.Contains(out, "sloprt.add(") {
		t.Fatalf("user call should not be sloprt.add, got:\n%s", out)
	}
}

func TestCodegen_MultiAssign(t *testing.T) {
	out, err := generate("fn foo() { <- [1, 2] }\na, b = foo()")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.UnpackTwo(") {
		t.Fatalf("expected sloprt.UnpackTwo, got:\n%s", out)
	}
}

func TestCodegen_ForLoop(t *testing.T) {
	out, err := generate("for { break }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "for {") {
		t.Fatalf("expected for {, got:\n%s", out)
	}
	if !strings.Contains(out, "break") {
		t.Fatalf("expected break, got:\n%s", out)
	}
}

func TestCodegen_BreakStmt(t *testing.T) {
	out, err := generate("for { if [1] { break } }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "break") {
		t.Fatalf("expected break, got:\n%s", out)
	}
}

// ==========================================
// Phase 4: Array Operator Codegen Tests
// ==========================================

func TestCodegen_IndexExpr(t *testing.T) {
	out, err := generate("arr = [10, 20, 30]\n|> arr@0")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Index(") {
		t.Fatalf("expected sloprt.Index, got:\n%s", out)
	}
}

func TestCodegen_IndexSetStmt(t *testing.T) {
	out, err := generate("arr = [10, 20, 30]\narr@0 = [99]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.IndexSet(") {
		t.Fatalf("expected sloprt.IndexSet, got:\n%s", out)
	}
}

func TestCodegen_PushStmt(t *testing.T) {
	out, err := generate("arr = [1, 2]\narr << [3]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Push(") {
		t.Fatalf("expected sloprt.Push, got:\n%s", out)
	}
}

func TestCodegen_PopExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = >>arr")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Pop(") {
		t.Fatalf("expected sloprt.Pop, got:\n%s", out)
	}
}

func TestCodegen_SliceExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3, 4]\nx = arr::1::3")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Slice(") {
		t.Fatalf("expected sloprt.Slice, got:\n%s", out)
	}
}

func TestCodegen_ConcatExpr(t *testing.T) {
	out, err := generate("x = [1, 2] ++ [3, 4]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Concat(") {
		t.Fatalf("expected sloprt.Concat, got:\n%s", out)
	}
}

func TestCodegen_RemoveExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = arr -- [2]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Remove(") {
		t.Fatalf("expected sloprt.Remove, got:\n%s", out)
	}
}

func TestCodegen_ContainsExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = arr ?? [2]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Contains(") {
		t.Fatalf("expected sloprt.Contains, got:\n%s", out)
	}
}

func TestCodegen_RemoveAtExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = arr ~@ [1]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.RemoveAt(") {
		t.Fatalf("expected sloprt.RemoveAt, got:\n%s", out)
	}
}

func TestCodegen_LengthExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = #arr")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Length(") {
		t.Fatalf("expected sloprt.Length, got:\n%s", out)
	}
}

func TestCodegen_UniqueExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 1]\nx = ~arr")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.Unique(") {
		t.Fatalf("expected sloprt.Unique, got:\n%s", out)
	}
}

func TestCodegen_ExprStmt(t *testing.T) {
	out, err := generate("fn foo() { |> \"hi\" }\nfoo()")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "foo()") {
		t.Fatalf("expected foo() call, got:\n%s", out)
	}
}

// ==========================================
// Phase 5: Hashmap Codegen Tests
// ==========================================

func TestCodegen_HashDeclStmt(t *testing.T) {
	out, err := generate(`person{name, age} = ["bob", [30]]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.MapFromKeysValues(") {
		t.Fatalf("expected sloprt.MapFromKeysValues, got:\n%s", out)
	}
	if !strings.Contains(out, `[]string{"name", "age"}`) {
		t.Fatalf("expected keys literal, got:\n%s", out)
	}
}

func TestCodegen_HashDeclStmt_EmptyKeys(t *testing.T) {
	out, err := generate(`counts{} = []`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.MapFromKeysValues(") {
		t.Fatalf("expected sloprt.MapFromKeysValues, got:\n%s", out)
	}
	if !strings.Contains(out, "[]string{}") {
		t.Fatalf("expected empty keys literal, got:\n%s", out)
	}
}

func TestCodegen_KeyAccessExpr(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\n|> person@name")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.IndexKeyStr(") {
		t.Fatalf("expected sloprt.IndexKeyStr, got:\n%s", out)
	}
	if !strings.Contains(out, `"name"`) {
		t.Fatalf("expected string key, got:\n%s", out)
	}
}

func TestCodegen_DynAccessExpr(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nk = \"name\"\n|> person$k")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.DynAccess(") {
		t.Fatalf("expected sloprt.DynAccess, got:\n%s", out)
	}
}

func TestCodegen_KeySetStmt(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nperson@name = [\"alice\"]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.IndexKeySetStr(") {
		t.Fatalf("expected sloprt.IndexKeySetStr, got:\n%s", out)
	}
}

func TestCodegen_DynAccessSetStmt(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nk = \"name\"\nperson$k = [\"alice\"]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.DynAccessSet(") {
		t.Fatalf("expected sloprt.DynAccessSet, got:\n%s", out)
	}
}

func TestCodegen_MapKeysExpr(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nx = ##person")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.MapKeys(") {
		t.Fatalf("expected sloprt.MapKeys, got:\n%s", out)
	}
}

func TestCodegen_MapValuesExpr(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nx = @@person")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.MapValues(") {
		t.Fatalf("expected sloprt.MapValues, got:\n%s", out)
	}
}

// ==========================================
// Null Literal Codegen Tests
// ==========================================

func TestCodegen_NullLiteral(t *testing.T) {
	out, err := generate(`x = [null]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "sloprt.SlopNull{}") {
		t.Fatalf("expected sloprt.SlopNull{}, got:\n%s", out)
	}
}
