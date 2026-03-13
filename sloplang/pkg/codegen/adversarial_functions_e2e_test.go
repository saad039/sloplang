package codegen

import "testing"

// --- Compile Errors (4 tests) ---

func TestAdv_Fn_TooFewArgs(t *testing.T) {
	runE2EExpectCompileError(t, "fn add(a, b) { <- a + b }\nadd([1])")
}

func TestAdv_Fn_TooManyArgs(t *testing.T) {
	runE2EExpectCompileError(t, "fn add(a, b) { <- a + b }\nadd([1], [2], [3])")
}

func TestAdv_Fn_UndefinedCall(t *testing.T) {
	runE2EExpectCompileError(t, `foo()`)
}

func TestAdv_Fn_ExitNoArgs(t *testing.T) {
	runE2EExpectCompileError(t, `exit()`)
}

// --- Runtime — Panics (1 test) ---

func TestAdv_Fn_ExitMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `exit([1, 2])`, "single-element")
}

// --- Runtime — Success (1 test) ---

func TestAdv_Fn_CallBeforeDeclWithArgs(t *testing.T) {
	got := runE2E(t, "|> str(add([3], [4]))\nfn add(a, b) { <- a + b }")
	if got != "[7]" {
		t.Fatalf("expected [7], got %q", got)
	}
}
