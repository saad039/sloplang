package codegen

import (
	"strings"
	"testing"

	"github.com/saad039/sloplang/pkg/lexer"
	"github.com/saad039/sloplang/pkg/parser"
)

func generate(input string) (string, error) {
	l := lexer.New(input)
	tokens := l.Tokenize()
	p := parser.New(tokens)
	prog, errs := p.Parse()
	if len(errs) > 0 {
		return "", nil
	}
	gen := New()
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
	if !strings.Contains(out, "NewSlopValue") {
		t.Fatalf("expected NewSlopValue call, got:\n%s", out)
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
	if !strings.Contains(out, "StdoutWrite") {
		t.Fatalf("expected StdoutWrite, got:\n%s", out)
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
	if !strings.Contains(out, "package main") {
		t.Fatalf("expected package main (second check), got:\n%s", out)
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
	if !strings.Contains(out, "NewSlopValue()") {
		t.Fatalf("expected NewSlopValue() with no args, got:\n%s", out)
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
	if !strings.Contains(out, "StdoutWrite(x)") {
		t.Fatalf("expected StdoutWrite(x), got:\n%s", out)
	}
}

func TestCodegen_BinaryAllOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x = [1] + [2]`, "Add("},
		{`x = [1] - [2]`, "Sub("},
		{`x = [1] * [2]`, "Mul("},
		{`x = [1] / [2]`, "Div("},
		{`x = [1] % [2]`, "Mod("},
		{`x = [1] ** [2]`, "Pow("},
		{`x = [1] == [2]`, "Eq("},
		{`x = [1] != [2]`, "Neq("},
		{`x = [1] < [2]`, "Lt("},
		{`x = [1] > [2]`, "Gt("},
		{`x = [1] <= [2]`, "Lte("},
		{`x = [1] >= [2]`, "Gte("},
		{`x = [1] && [2]`, "And("},
		{`x = [1] || [2]`, "Or("},
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
	if !strings.Contains(out, "Negate(") {
		t.Fatalf("expected Negate, got:\n%s", out)
	}
}

func TestCodegen_UnaryNot(t *testing.T) {
	out, err := generate(`x = ![1]`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Not(") {
		t.Fatalf("expected Not, got:\n%s", out)
	}
}

func TestCodegen_CallStr(t *testing.T) {
	out, err := generate(`|> str([1])`)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Str(") {
		t.Fatalf("expected Str, got:\n%s", out)
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
	if !strings.Contains(out, "*SlopValue") {
		t.Fatalf("expected *SlopValue param type, got:\n%s", out)
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
	if !strings.Contains(out, "Iterate(") {
		t.Fatalf("expected Iterate, got:\n%s", out)
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
	if !strings.Contains(out, "return NewSlopValue(") {
		t.Fatalf("expected return NewSlopValue, got:\n%s", out)
	}
}

func TestCodegen_BareReturn(t *testing.T) {
	out, err := generate("fn foo() { <- }")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "return NewSlopValue()") {
		t.Fatalf("expected return NewSlopValue(), got:\n%s", out)
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
	if !strings.Contains(out, "UnpackTwo(") {
		t.Fatalf("expected UnpackTwo, got:\n%s", out)
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
	if !strings.Contains(out, "Index(") {
		t.Fatalf("expected Index, got:\n%s", out)
	}
}

func TestCodegen_IndexSetStmt(t *testing.T) {
	out, err := generate("arr = [10, 20, 30]\narr@0 = [99]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "IndexSet(") {
		t.Fatalf("expected IndexSet, got:\n%s", out)
	}
}

func TestCodegen_PushStmt(t *testing.T) {
	out, err := generate("arr = [1, 2]\narr << [3]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Push(") {
		t.Fatalf("expected Push, got:\n%s", out)
	}
}

func TestCodegen_PopExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = >>arr")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Pop(") {
		t.Fatalf("expected Pop, got:\n%s", out)
	}
}

func TestCodegen_SliceExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3, 4]\nx = arr::1::3")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Slice(") {
		t.Fatalf("expected Slice, got:\n%s", out)
	}
}

func TestCodegen_ConcatExpr(t *testing.T) {
	out, err := generate("x = [1, 2] ++ [3, 4]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Concat(") {
		t.Fatalf("expected Concat, got:\n%s", out)
	}
}

func TestCodegen_RemoveExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = arr -- [2]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Remove(") {
		t.Fatalf("expected Remove, got:\n%s", out)
	}
}

func TestCodegen_ContainsExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = arr ?? [2]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Contains(") {
		t.Fatalf("expected Contains, got:\n%s", out)
	}
}

func TestCodegen_RemoveAtExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = arr ~@ [1]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "RemoveAt(") {
		t.Fatalf("expected RemoveAt, got:\n%s", out)
	}
}

func TestCodegen_LengthExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 3]\nx = #arr")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Length(") {
		t.Fatalf("expected Length, got:\n%s", out)
	}
}

func TestCodegen_UniqueExpr(t *testing.T) {
	out, err := generate("arr = [1, 2, 1]\nx = ~arr")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "Unique(") {
		t.Fatalf("expected Unique, got:\n%s", out)
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
	if !strings.Contains(out, "MapFromKeysValues(") {
		t.Fatalf("expected MapFromKeysValues, got:\n%s", out)
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
	if !strings.Contains(out, "MapFromKeysValues(") {
		t.Fatalf("expected MapFromKeysValues, got:\n%s", out)
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
	if !strings.Contains(out, "IndexKeyStr(") {
		t.Fatalf("expected IndexKeyStr, got:\n%s", out)
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
	if !strings.Contains(out, "DynAccess(") {
		t.Fatalf("expected DynAccess, got:\n%s", out)
	}
}

func TestCodegen_KeySetStmt(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nperson@name = [\"alice\"]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "IndexKeySetStr(") {
		t.Fatalf("expected IndexKeySetStr, got:\n%s", out)
	}
}

func TestCodegen_DynAccessSetStmt(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nk = \"name\"\nperson$k = [\"alice\"]")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "DynAccessSet(") {
		t.Fatalf("expected DynAccessSet, got:\n%s", out)
	}
}

func TestCodegen_MapKeysExpr(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nx = ##person")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "MapKeys(") {
		t.Fatalf("expected MapKeys, got:\n%s", out)
	}
}

func TestCodegen_MapValuesExpr(t *testing.T) {
	out, err := generate("person{name} = [\"bob\"]\nx = @@person")
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}
	if !strings.Contains(out, "MapValues(") {
		t.Fatalf("expected MapValues, got:\n%s", out)
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
	if !strings.Contains(out, "SlopNull{}") {
		t.Fatalf("expected SlopNull{}, got:\n%s", out)
	}
}
