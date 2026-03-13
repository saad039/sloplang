package codegen

import "testing"

// --- Compile Errors (8 tests) ---

func TestAdv_Trans_VarNamedMain(t *testing.T) {
	runE2EExpectCompileError(t, "main = [1]\n|> str(main)")
}

func TestAdv_Trans_VarNamedFmt(t *testing.T) {
	runE2EExpectCompileError(t, "fmt = [1]\n|> str(fmt)")
}

func TestAdv_Trans_VarNamedSlopValue(t *testing.T) {
	runE2EExpectCompileError(t, "SlopValue = [1]\n|> str(SlopValue)")
}

func TestAdv_Trans_VarNamedNewSlopValue(t *testing.T) {
	runE2EExpectCompileError(t, "NewSlopValue = [1]\n|> str(NewSlopValue)")
}

func TestAdv_Trans_VarNamedAdd(t *testing.T) {
	runE2EExpectCompileError(t, "Add = [1]\n|> str(Add)")
}

func TestAdv_Trans_VarNamedInt64(t *testing.T) {
	runE2EExpectCompileError(t, "int64 = [1]\n|> str(int64)")
}

func TestAdv_Trans_VarNamedString(t *testing.T) {
	runE2EExpectCompileError(t, "string = [1]\n|> str(string)")
}

func TestAdv_Trans_FnNamedPanic(t *testing.T) {
	runE2EExpectCompileError(t, "fn panic() { <- [1] }\n|> str(panic())")
}

// --- Runtime — Success (4 tests) ---

func TestAdv_Trans_VarNamedValue(t *testing.T) {
	got := runE2E(t, "value = [42]\n|> str(value)")
	if got != "[42]" {
		t.Fatalf("expected [42], got %q", got)
	}
}

func TestAdv_Trans_VarNamedResult(t *testing.T) {
	got := runE2E(t, "result = [1]\n|> str(result)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Trans_ManyVarsSamePrefix(t *testing.T) {
	got := runE2E(t, "a1 = [1]\na2 = [2]\na3 = [3]\na4 = [4]\na5 = [5]\n|> str(a1 + a2 + a3 + a4 + a5)")
	if got != "[15]" {
		t.Fatalf("expected [15], got %q", got)
	}
}

func TestAdv_Trans_LongVarName(t *testing.T) {
	got := runE2E(t, "superlongvariablename = [99]\n|> str(superlongvariablename)")
	if got != "[99]" {
		t.Fatalf("expected [99], got %q", got)
	}
}
