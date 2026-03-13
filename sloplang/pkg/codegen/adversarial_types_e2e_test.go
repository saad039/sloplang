package codegen

import "testing"

// --- Runtime — Panics (4 tests) ---

func TestAdv_Type_FalseInArithmetic(t *testing.T) {
	runE2EExpectPanicContaining(t, `x = false + [1]`, "length")
}

func TestAdv_Type_ToNumNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `v, e = to_num([null])`, "Null")
}

func TestAdv_Type_BracketTrueInBoolean(t *testing.T) {
	runE2EExpectPanicContaining(t, `if [true] { |> "yes" }`, "boolean")
}

func TestAdv_Type_BracketFalseInBoolean(t *testing.T) {
	runE2EExpectPanicContaining(t, `if [false] { |> "yes" }`, "boolean")
}

// --- Runtime — Success (11 tests) ---

func TestAdv_Type_IntOverflowAdd(t *testing.T) {
	got := runE2E(t, "x = [9223372036854775807] + [1]\n|> str(x)")
	if got != "[-9223372036854775808]" {
		t.Fatalf("expected [-9223372036854775808], got %q", got)
	}
}

func TestAdv_Type_IntOverflowMul(t *testing.T) {
	got := runE2E(t, "x = [9223372036854775807] * [2]\n|> str(x)")
	if got != "[-2]" {
		t.Fatalf("expected [-2], got %q", got)
	}
}

func TestAdv_Type_UintUnderflow(t *testing.T) {
	got := runE2E(t, "x = [0u] - [1u]\n|> str(x)")
	if got != "[18446744073709551615]" {
		t.Fatalf("expected [18446744073709551615], got %q", got)
	}
}

func TestAdv_Type_FloatPrecision(t *testing.T) {
	got := runE2E(t, "x = [0.1] + [0.2]\n|> str(x)")
	// IEEE 754: 0.1 + 0.2 != 0.3
	if got == "[0.3]" {
		t.Fatalf("expected IEEE 754 imprecision, got exact [0.3]")
	}
}

func TestAdv_Type_BoolInArithmetic(t *testing.T) {
	got := runE2E(t, "x = true + true\n|> str(x)")
	if got != "[2]" {
		t.Fatalf("expected [2], got %q", got)
	}
}

func TestAdv_Type_NullInConcat(t *testing.T) {
	got := runE2E(t, "x = [null] ++ [1]\n|> str(x)")
	if got != "[null, 1]" {
		t.Fatalf("expected [null, 1], got %q", got)
	}
}

func TestAdv_Type_MixedTypeArray(t *testing.T) {
	got := runE2E(t, "x = [1, \"hello\", null]\n|> str(x)")
	if got != "[1, hello, null]" {
		t.Fatalf("expected [1, hello, null], got %q", got)
	}
}

func TestAdv_Type_EmptyStringComparison(t *testing.T) {
	got := runE2E(t, "if [\"\"] < [\"a\"] { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected yes, got %q", got)
	}
}

func TestAdv_Type_IntegerDivisionTruncates(t *testing.T) {
	got := runE2E(t, "x = [7] / [2]\n|> str(x)")
	if got != "[3]" {
		t.Fatalf("expected [3], got %q", got)
	}
}

func TestAdv_Type_MultiReturnFromNonDualFn(t *testing.T) {
	// UnpackTwo requires at least 2 elements — [42] has only 1, so it panics
	runE2EExpectPanicContaining(t, "fn f() { <- [42] }\na, b = f()\n|> str(a)", "unpack")
}

func TestAdv_Type_StdoutMultiElement(t *testing.T) {
	got := runE2E(t, `|> [1, 2, 3]`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected [1, 2, 3], got %q", got)
	}
}
