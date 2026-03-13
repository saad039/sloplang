package codegen

import "testing"

// --- Functions are not first-class (3 tests) ---

func TestAdv_Spec_FnNotFirstClass_AssignToVar(t *testing.T) {
	// GIVEN: a function assigned to a variable
	// WHEN: compiled
	// THEN: compile error — functions are not first-class
	runE2EExpectCompileError(t, "fn f() { <- [1] }\ng = f")
}

func TestAdv_Spec_FnNotFirstClass_PassAsArg(t *testing.T) {
	// GIVEN: a function passed as argument
	// WHEN: compiled
	// THEN: compile error
	runE2EExpectCompileError(t, "fn f() { <- [1] }\nfn g(x) { <- x }\ng(f)")
}

func TestAdv_Spec_FnNotFirstClass_ReturnFn(t *testing.T) {
	// GIVEN: a function returned from another function
	// WHEN: compiled
	// THEN: compile error
	runE2EExpectCompileError(t, "fn f() { <- [1] }\nfn g() { <- f }\ng()")
}

// --- Boolean strictness (4 tests) ---

func TestAdv_Spec_ZeroNotFalsy_ExactMsg(t *testing.T) {
	// GIVEN: [0] used in if condition
	// WHEN: executed
	// THEN: panics with "use [] for false"
	runE2EExpectPanicContaining(t, `if [0] { |> "yes" }`, "use [] for false")
}

func TestAdv_Spec_StringNotBool(t *testing.T) {
	// GIVEN: string used in if condition
	// WHEN: executed
	// THEN: panics — string not valid boolean
	runE2EExpectPanicContaining(t, `if "hello" { |> "yes" }`, "")
}

func TestAdv_Spec_FloatNotBool(t *testing.T) {
	// GIVEN: float used in if condition
	// WHEN: executed
	// THEN: panics — float not valid boolean
	runE2EExpectPanicContaining(t, `if [1.0] { |> "yes" }`, "")
}

func TestAdv_Spec_NullNotBool(t *testing.T) {
	// GIVEN: null used in if condition
	// WHEN: executed
	// THEN: panics mentioning null
	runE2EExpectPanicContaining(t, `if [null] { |> "yes" }`, "null")
}

// --- str() formatting contract (1 test) ---

func TestAdv_Spec_StrMultiString(t *testing.T) {
	// GIVEN: multi-element string array
	// WHEN: str() applied
	// THEN: bracketed comma-separated output
	got := runE2E(t, "|> str([\"a\", \"b\"])")
	if got != "[a, b]" {
		t.Fatalf("expected [a, b], got %q", got)
	}
}

// --- Slice semantics (2 tests) ---

func TestAdv_Spec_SliceInclusiveExclusive(t *testing.T) {
	// GIVEN: array sliced with ::start::end
	// WHEN: executed
	// THEN: start-inclusive, end-exclusive
	got := runE2E(t, "x = [10, 20, 30, 40, 50]\n|> str(x::1::4)")
	if got != "[20, 30, 40]" {
		t.Fatalf("expected [20, 30, 40], got %q", got)
	}
}

func TestAdv_Spec_SliceFullRange(t *testing.T) {
	// GIVEN: array sliced from 0 to length
	// WHEN: executed
	// THEN: returns copy of full array
	got := runE2E(t, "x = [1, 2, 3]\nn = #x\n|> str(x::0::n)")
	if got != "[1, 2, 3]" {
		t.Fatalf("expected [1, 2, 3], got %q", got)
	}
}

// --- Stdout has no trailing newline (2 tests) ---

func TestAdv_Spec_StdoutNoNewline(t *testing.T) {
	// GIVEN: two stdout writes
	// WHEN: executed
	// THEN: output is concatenated with no newline between
	got := runE2E(t, "|> \"a\"\n|> \"b\"")
	if got != "ab" {
		t.Fatalf("expected ab, got %q", got)
	}
}

func TestAdv_Spec_ExplicitNewline(t *testing.T) {
	// GIVEN: stdout write with \n
	// WHEN: executed
	// THEN: newline only where explicit
	got := runE2E(t, `|> "a\nb"`)
	if got != "a\nb" {
		t.Fatalf("expected a\\nb, got %q", got)
	}
}

// --- Remove operator (1 test) ---

func TestAdv_Spec_RemoveFirstOccurrenceOnly(t *testing.T) {
	// GIVEN: array with multiple occurrences of value
	// WHEN: -- applied
	// THEN: only first occurrence removed
	got := runE2E(t, "x = [1, 2, 1, 3, 1]\ny = x -- [1]\n|> str(y)")
	if got != "[2, 1, 3, 1]" {
		t.Fatalf("expected [2, 1, 3, 1], got %q", got)
	}
}

// --- ++ is array concat, not string concat (1 test) ---

func TestAdv_Spec_ConcatNotStringConcat(t *testing.T) {
	// GIVEN: string ++ str(number)
	// WHEN: executed
	// THEN: produces two-element array, not string concatenation
	got := runE2E(t, "x = \"hello\" ++ str([5])\n|> str(x)")
	if got != "[hello, [5]]" {
		t.Fatalf("expected [hello, [5]], got %q", got)
	}
}

// --- Snapshot tests (6 tests) ---

func TestAdv_Snap_LengthMismatch(t *testing.T) {
	// GIVEN: addition of arrays with different lengths
	// WHEN: executed
	// THEN: panic message clearly states both lengths
	runE2EExpectPanicSnapshot(t, `x = [1, 2] + [3]`, "length_mismatch.snapshot")
}

func TestAdv_Snap_TypeMismatch(t *testing.T) {
	// GIVEN: addition of int and float
	// WHEN: executed
	// THEN: panic message names both types
	runE2EExpectPanicSnapshot(t, `x = [1] + [1.0]`, "type_mismatch.snapshot")
}

func TestAdv_Snap_IndexOutOfRange(t *testing.T) {
	// GIVEN: index beyond array length
	// WHEN: executed
	// THEN: panic shows index and array length
	runE2EExpectPanicSnapshot(t, "x = [1, 2]\n|> x@5", "index_out_of_range.snapshot")
}

func TestAdv_Snap_ZeroNotFalsy(t *testing.T) {
	// GIVEN: [0] used as boolean
	// WHEN: executed
	// THEN: panic includes "use [] for false" suggestion
	runE2EExpectPanicSnapshot(t, `if [0] { }`, "zero_not_falsy.snapshot")
}

func TestAdv_Snap_KeyNotFound(t *testing.T) {
	// GIVEN: accessing non-existent hashmap key
	// WHEN: executed
	// THEN: panic shows which key was missing
	runE2EExpectPanicSnapshot(t, "m{a} = [[1]]\n|> m@b", "key_not_found.snapshot")
}

func TestAdv_Snap_DivByZero(t *testing.T) {
	// GIVEN: division by zero
	// WHEN: executed
	// THEN: clean sloplang error message
	runE2EExpectPanicSnapshot(t, `x = [1] / [0]`, "div_by_zero.snapshot")
}
