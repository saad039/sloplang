package codegen

import "testing"

// --- Compile Errors (3 tests) ---

func TestAdv_Scope_UndefinedVar(t *testing.T) {
	runE2EExpectCompileError(t, `|> x`)
}

func TestAdv_Scope_LocalDoesNotLeak(t *testing.T) {
	runE2EExpectCompileError(t, "fn f() { y = [5] }\nf()\n|> y")
}

func TestAdv_Scope_SameNameFnAndVar(t *testing.T) {
	runE2EExpectCompileError(t, "x = [1]\nfn x() { <- [2] }")
}

// --- Runtime — BUG (1 test) ---

func TestAdv_Scope_UseBeforeAssign(t *testing.T) {
	// GIVEN: a program that uses a variable before assigning it
	// WHEN: transpiled, compiled, and executed
	// THEN: should produce a clean panic "variable used before assignment"
	runE2EExpectPanicContaining(t, "|> x\nx = [1]", "used before assignment")
}

// --- Runtime — Success (5 tests) ---

func TestAdv_Scope_GlobalMutFromFn(t *testing.T) {
	got := runE2E(t, "x = [1]\nfn f() { x = [2]\n<- [] }\nf()\n|> str(x)")
	if got != "[2]" {
		t.Fatalf("expected [2], got %q", got)
	}
}

func TestAdv_Scope_LoopVarShadowsOuter(t *testing.T) {
	got := runE2E(t, "i = [99]\nfor i in [1, 2] { }\n|> str(i)")
	if got != "[99]" {
		t.Fatalf("expected [99], got %q", got)
	}
}

func TestAdv_Scope_FnCallBeforeDecl(t *testing.T) {
	got := runE2E(t, "|> str(f())\nfn f() { <- [42] }")
	if got != "[42]" {
		t.Fatalf("expected [42], got %q", got)
	}
}

func TestAdv_Scope_MultiAssignDuplicateVar(t *testing.T) {
	// x, x = <. "nonexistent.txt" — second x overwrites first, x gets error code [1]
	got := runE2E(t, "x, x = <. \"nonexistent.txt\"\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Scope_MutualRecursionWithBase(t *testing.T) {
	got := runE2E(t, "fn a(n) { if n == [0] { <- [0] }\n<- b(n - [1]) }\nfn b(n) { <- a(n) }\n|> str(a([5]))")
	if got != "[0]" {
		t.Fatalf("expected [0], got %q", got)
	}
}
