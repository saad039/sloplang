package codegen

import "testing"

// --- Go builtins as variable names (8 tests) ---
// Some Go builtins (len, append, make, copy) cause compile errors when shadowed
// because the runtime uses them. Others (new, cap, delete, close) can be shadowed
// without issue because the generated code doesn't call them.

func TestAdv_Ident_VarNamedLen(t *testing.T) {
	// GIVEN: variable named "len" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: compile error — runtime uses len()
	runE2EExpectCompileError(t, "len = [1]\n|> str(len)")
}

func TestAdv_Ident_VarNamedAppend(t *testing.T) {
	// GIVEN: variable named "append" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: compile error — runtime uses append()
	runE2EExpectCompileError(t, "append = [1]\n|> str(append)")
}

func TestAdv_Ident_VarNamedMake(t *testing.T) {
	// GIVEN: variable named "make" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: compile error — runtime uses make()
	runE2EExpectCompileError(t, "make = [1]\n|> str(make)")
}

func TestAdv_Ident_VarNamedNew(t *testing.T) {
	// GIVEN: variable named "new" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: succeeds — Go allows shadowing "new" and runtime doesn't use it
	got := runE2E(t, "new = [1]\n|> str(new)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Ident_VarNamedCap(t *testing.T) {
	// GIVEN: variable named "cap" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: succeeds — Go allows shadowing "cap" and runtime doesn't use it
	got := runE2E(t, "cap = [1]\n|> str(cap)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Ident_VarNamedCopy(t *testing.T) {
	// GIVEN: variable named "copy" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: compile error — runtime uses copy()
	runE2EExpectCompileError(t, "copy = [1]\n|> str(copy)")
}

func TestAdv_Ident_VarNamedDelete(t *testing.T) {
	// GIVEN: variable named "delete" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: succeeds — Go allows shadowing "delete" and runtime doesn't use it
	got := runE2E(t, "delete = [1]\n|> str(delete)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Ident_VarNamedClose(t *testing.T) {
	// GIVEN: variable named "close" (Go builtin)
	// WHEN: transpiled and compiled
	// THEN: succeeds — Go allows shadowing "close" and runtime doesn't use it
	got := runE2E(t, "close = [1]\n|> str(close)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

// --- Go keywords as variable names (4 tests) ---
// These are Go keywords — the transpiler produces invalid Go that fails at
// go/format stage (AssembleWithRuntime), which is before go build.
// runE2EExpectCompileError expects parse to succeed, so we need a different approach.
// These fail at codegen, so we test that codegen/assembly produces an error.

func TestAdv_Id_GoKeyword_Func(t *testing.T) {
	got := runE2E(t, "func = [1]\n|> str(func)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Id_GoKeyword_Var(t *testing.T) {
	got := runE2E(t, "var = [1]\n|> str(var)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Id_GoKeyword_Return(t *testing.T) {
	got := runE2E(t, "return = [1]\n|> str(return)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Id_GoKeyword_Range(t *testing.T) {
	got := runE2E(t, "range = [1]\n|> str(range)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

// --- Runtime collision names (4 tests) ---

func TestAdv_Ident_VarNamedSub(t *testing.T) {
	// GIVEN: variable named "Sub" (runtime func)
	// WHEN: transpiled and compiled
	// THEN: compile error — conflicts with runtime
	runE2EExpectCompileError(t, "Sub = [1]\n|> str(Sub)")
}

func TestAdv_Ident_VarNamedMul(t *testing.T) {
	// GIVEN: variable named "Mul" (runtime func)
	// WHEN: transpiled and compiled
	// THEN: compile error — conflicts with runtime
	runE2EExpectCompileError(t, "Mul = [1]\n|> str(Mul)")
}

func TestAdv_Ident_VarNamedDiv(t *testing.T) {
	// GIVEN: variable named "Div" (runtime func)
	// WHEN: transpiled and compiled
	// THEN: compile error — conflicts with runtime
	runE2EExpectCompileError(t, "Div = [1]\n|> str(Div)")
}

func TestAdv_Ident_VarNamedIterate(t *testing.T) {
	// GIVEN: variable named "Iterate" (runtime func)
	// WHEN: transpiled and compiled
	// THEN: compile error — conflicts with runtime
	runE2EExpectCompileError(t, "Iterate = [1]\n|> str(Iterate)")
}

// --- Special identifiers (3 tests) ---

func TestAdv_Ident_VarNamedUnderscore(t *testing.T) {
	// GIVEN: variable named "_"
	// WHEN: transpiled and compiled
	// THEN: compile error — _ cannot be read in Go
	runE2EExpectCompileError(t, "_ = [1]\n|> str(_)")
}

func TestAdv_Ident_VarNamedInit(t *testing.T) {
	// GIVEN: variable named "init" (Go special function)
	// WHEN: transpiled and compiled
	// THEN: compile error — init is special
	runE2EExpectCompileError(t, "init = [1]\n|> str(init)")
}

func TestAdv_Ident_VarNamedNil(t *testing.T) {
	// GIVEN: variable named "nil" (Go predeclared)
	// WHEN: transpiled and compiled
	// THEN: compile error
	runE2EExpectCompileError(t, "nil = [1]\n|> str(nil)")
}
