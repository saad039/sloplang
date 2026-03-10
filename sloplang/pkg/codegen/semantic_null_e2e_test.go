package codegen

import "testing"

// ============================================================
// Semantic Null E2E Tests
// Covers null assignment, str(), equality, arithmetic panics,
// boolean panics, ordered comparison panics, iteration, and
// edge cases with array operators.
// ============================================================

// --- Null succeeds (~13 tests) ---

func TestSem_Null_Assignment(t *testing.T) {
	got := runE2E(t, `x = [null]
|> str(x)`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestSem_Null_StrNull(t *testing.T) {
	got := runE2E(t, `|> str([null])`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestSem_Null_StdoutNull(t *testing.T) {
	got := runE2E(t, `|> str([null])`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestSem_Null_EqNullNull(t *testing.T) {
	got := runE2E(t, `if [null] == [null] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Null_NeqNullInt(t *testing.T) {
	got := runE2E(t, `if [null] != [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Null_IntNeqNull(t *testing.T) {
	got := runE2E(t, `if [1] != [null] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Null_NeqNullNull(t *testing.T) {
	got := runE2E(t, `if [null] != [null] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Null_LengthNull(t *testing.T) {
	got := runE2E(t, `|> str(#[null])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Null_LengthMultiNull(t *testing.T) {
	got := runE2E(t, `|> str(#[null, null])`)
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestSem_Null_ContainsNull(t *testing.T) {
	got := runE2E(t, `arr = [1, null, 3]
if arr ?? [null] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Null_IndexNull(t *testing.T) {
	got := runE2E(t, `arr = [null, null]
|> str(arr@0)`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestSem_Null_MixedArray(t *testing.T) {
	got := runE2E(t, `|> str([1, null, "hi"])`)
	if got != "[1, null, hi]" {
		t.Fatalf("expected %q, got %q", "[1, null, hi]", got)
	}
}

func TestSem_Null_HashmapValue(t *testing.T) {
	got := runE2E(t, `m{a} = [[null]]
|> str(m@a)`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

// --- Null panics — arithmetic (~7 tests) ---

func TestSem_Null_AddPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = x + [1]`)
}

func TestSem_Null_SubPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = x - [1]`)
}

func TestSem_Null_MulPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = x * [1]`)
}

func TestSem_Null_DivPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = x / [1]`)
}

func TestSem_Null_ModPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = x % [1]`)
}

func TestSem_Null_PowPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = x ** [1]`)
}

func TestSem_Null_RightSidePanics(t *testing.T) {
	runE2EExpectPanic(t, `y = [1] + [null]`)
}

// --- Null panics — unary/boolean (~5 tests) ---

func TestSem_Null_NegatePanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
y = -x`)
}

func TestSem_Null_NotPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if !x { }`)
}

func TestSem_Null_IfPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [null] { }`)
}

func TestSem_Null_AndPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if x && [1] { }`)
}

func TestSem_Null_OrPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if x || [1] { }`)
}

// --- Null panics — ordered comparison (~5 tests) ---

func TestSem_Null_LtPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if x < [1] { }`)
}

func TestSem_Null_GtPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if x > [1] { }`)
}

func TestSem_Null_LtePanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if x <= [1] { }`)
}

func TestSem_Null_GtePanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null]
if x >= [1] { }`)
}

func TestSem_Null_RightSideOrderedPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [1] < [null] { }`)
}

// --- Null panics — iteration (~1 test) ---

func TestSem_Null_IterateSinglePanics(t *testing.T) {
	runE2EExpectPanic(t, `for x in [null] { }`)
}

// --- Null edge cases (~5 tests) ---

func TestSem_Null_MultiNullIterates(t *testing.T) {
	got := runE2E(t, `for x in [null, null] {
|> str(x)
}`)
	if got != "[null][null]" {
		t.Fatalf("expected %q, got %q", "[null][null]", got)
	}
}

func TestSem_Null_UniqueNull(t *testing.T) {
	got := runE2E(t, `|> str(~[null, null, 1])`)
	if got != "[null, 1]" {
		t.Fatalf("expected %q, got %q", "[null, 1]", got)
	}
}

func TestSem_Null_RemoveNull(t *testing.T) {
	got := runE2E(t, `|> str([1, null, 3] -- [null])`)
	if got != "[1, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 3]", got)
	}
}

func TestSem_Null_ConcatNull(t *testing.T) {
	got := runE2E(t, `|> str([1, null] ++ [null, 2])`)
	if got != "[1, null, null, 2]" {
		t.Fatalf("expected %q, got %q", "[1, null, null, 2]", got)
	}
}

func TestSem_Null_NullInMixedIteration(t *testing.T) {
	got := runE2E(t, `for x in [null, 1] {
|> str(x)
}`)
	if got != "[null][1]" {
		t.Fatalf("expected %q, got %q", "[null][1]", got)
	}
}
