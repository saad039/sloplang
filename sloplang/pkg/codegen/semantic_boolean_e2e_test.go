package codegen

import "testing"

// ============================================================
// Semantic Boolean E2E Tests
// Covers truthiness, falsiness, logical operators, true/false keywords
// ============================================================

// --- Valid booleans (~6 tests) ---

func TestSem_Bool_TruthyOne(t *testing.T) {
	got := runE2E(t, `if [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_FalsyEmpty(t *testing.T) {
	got := runE2E(t, `if [] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_TrueKeyword(t *testing.T) {
	got := runE2E(t, `if true { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_FalseKeyword(t *testing.T) {
	got := runE2E(t, `if false { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_TrueAssign(t *testing.T) {
	got := runE2E(t, `x = true
if x { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_FalseAssign(t *testing.T) {
	got := runE2E(t, `x = false
if x { } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

// --- Invalid booleans (all panic) (~10 tests) ---

func TestSem_Bool_ZeroPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [0] { }`)
}

func TestSem_Bool_TwoPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [2] { }`)
}

func TestSem_Bool_NegOnePanics(t *testing.T) {
	runE2EExpectPanic(t, `x = -[1]
if x { }`)
}

func TestSem_Bool_StringPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = "hello"
if x { }`)
}

func TestSem_Bool_EmptyStringPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = ""
if x { }`)
}

func TestSem_Bool_FloatPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [3.14] { }`)
}

func TestSem_Bool_MultiPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [1, 2] { }`)
}

func TestSem_Bool_NullPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [null] { }`)
}

func TestSem_Bool_UintZeroPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [0u] { }`)
}

func TestSem_Bool_UintOnePanics(t *testing.T) {
	runE2EExpectPanic(t, `if [1u] { }`)
}

// --- Logical operators (~11 tests) ---

func TestSem_Bool_AndBothTrue(t *testing.T) {
	got := runE2E(t, `if [1] && [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_AndLeftFalse(t *testing.T) {
	got := runE2E(t, `if [] && [1] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_AndRightFalse(t *testing.T) {
	got := runE2E(t, `if [1] && [] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_OrLeftTrue(t *testing.T) {
	got := runE2E(t, `if [1] || [] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_OrLeftFalse(t *testing.T) {
	got := runE2E(t, `if [] || [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_OrBothFalse(t *testing.T) {
	got := runE2E(t, `if [] || [] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_NotTrue(t *testing.T) {
	got := runE2E(t, `if ![1] { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_NotFalse(t *testing.T) {
	got := runE2E(t, `if ![] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_NotTrueKw(t *testing.T) {
	got := runE2E(t, `if !true { } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}

func TestSem_Bool_NotFalseKw(t *testing.T) {
	got := runE2E(t, `if !false { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_DoubleNot(t *testing.T) {
	got := runE2E(t, `if !!true { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

// --- Logical with invalid operands (panic) (~4 tests) ---

func TestSem_Bool_ZeroAndPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [0]
if x && [1] { }`)
}

func TestSem_Bool_AndRightZeroPanics(t *testing.T) {
	runE2EExpectPanic(t, `if [1] && [0] { }`)
}

func TestSem_Bool_NotZeroPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [0]
if !x { }`)
}

func TestSem_Bool_ZeroOrPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [0]
if x || [1] { }`)
}

// --- true/false keyword semantics (~8 tests) ---

func TestSem_Bool_StrTrue(t *testing.T) {
	got := runE2E(t, `|> str(true)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Bool_StrFalse(t *testing.T) {
	got := runE2E(t, `|> str(false)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Bool_TrueEqOne(t *testing.T) {
	got := runE2E(t, `if true == [1] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_FalseEqEmpty(t *testing.T) {
	got := runE2E(t, `if false == [] { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_TrueNeqFalse(t *testing.T) {
	got := runE2E(t, `if true != false { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_BracketTrue(t *testing.T) {
	got := runE2E(t, `|> str([true])`)
	if got != "[[1]]" {
		t.Fatalf("expected %q, got %q", "[[1]]", got)
	}
}

func TestSem_Bool_BracketFalse(t *testing.T) {
	got := runE2E(t, `|> str([false])`)
	if got != "[[]]" {
		t.Fatalf("expected %q, got %q", "[[]]", got)
	}
}

func TestSem_Bool_AssignThenStr(t *testing.T) {
	got := runE2E(t, `x = true
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// --- Boolean results used (~2 tests) ---

func TestSem_Bool_EqResultTruthy(t *testing.T) {
	got := runE2E(t, `r = [1] == [1]
if r { |> "yes" }`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Bool_EqResultFalsy(t *testing.T) {
	got := runE2E(t, `r = [1] == [2]
if r { |> "yes" } else { |> "no" }`)
	if got != "no" {
		t.Fatalf("expected %q, got %q", "no", got)
	}
}
