package codegen

import "testing"

// --- Runtime — Success (15 tests) ---

func TestAdv_Prec_PowBeforeMul(t *testing.T) {
	got := runE2E(t, `|> str([2] ** [3] * [4])`)
	if got != "[32]" {
		t.Fatalf("expected [32], got %q", got)
	}
}

func TestAdv_Prec_PowRightAssoc(t *testing.T) {
	got := runE2E(t, `|> str([2] ** [3] ** [2])`)
	if got != "[512]" {
		t.Fatalf("expected [512], got %q", got)
	}
}

func TestAdv_Prec_UnaryMinusVsBinaryMinus(t *testing.T) {
	got := runE2E(t, "x = [5]\n|> str(-x - [2])")
	if got != "[-7]" {
		t.Fatalf("expected [-7], got %q", got)
	}
}

func TestAdv_Prec_LengthInExpr(t *testing.T) {
	got := runE2E(t, "arr = [1, 2, 3]\n|> str(#arr + [10])")
	if got != "[13]" {
		t.Fatalf("expected [13], got %q", got)
	}
}

func TestAdv_Prec_IndexInExpr(t *testing.T) {
	got := runE2E(t, "arr = [10, 20, 30]\n|> str(arr@1 + [5])")
	if got != "[25]" {
		t.Fatalf("expected [25], got %q", got)
	}
}

func TestAdv_Prec_DynAccessInExpr(t *testing.T) {
	got := runE2E(t, "arr = [10, 20]\ni = [1]\n|> str(arr$i * [3])")
	if got != "[60]" {
		t.Fatalf("expected [60], got %q", got)
	}
}

func TestAdv_Prec_ComparisonVsLogical(t *testing.T) {
	got := runE2E(t, "x = [1] < [2] && [3] > [1]\nif x { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected yes, got %q", got)
	}
}

func TestAdv_Prec_NotVsComparison(t *testing.T) {
	got := runE2E(t, "x = !([1] == [2])\nif x { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected yes, got %q", got)
	}
}

func TestAdv_Prec_MixedArithPow(t *testing.T) {
	got := runE2E(t, `|> str([2] + [3] ** [2] * [4])`)
	if got != "[38]" {
		t.Fatalf("expected [38], got %q", got)
	}
}

func TestAdv_Prec_ContainsVsLogical(t *testing.T) {
	got := runE2E(t, "if [1, 2, 3] ?? [2] && [4, 5] ?? [5] { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected yes, got %q", got)
	}
}

func TestAdv_Prec_NegatePow(t *testing.T) {
	got := runE2E(t, `|> str(-[2] ** [3])`)
	if got != "[-8]" {
		t.Fatalf("expected [-8], got %q", got)
	}
}

func TestAdv_Prec_ChainedAdd(t *testing.T) {
	got := runE2E(t, `|> str([1] + [2] + [3] + [4])`)
	if got != "[10]" {
		t.Fatalf("expected [10], got %q", got)
	}
}

func TestAdv_Prec_ParensOverridePow(t *testing.T) {
	got := runE2E(t, `|> str(([2] + [3]) ** [2])`)
	if got != "[25]" {
		t.Fatalf("expected [25], got %q", got)
	}
}

func TestAdv_Prec_ConcatVsAdd(t *testing.T) {
	got := runE2E(t, `|> str([1] ++ [2] ++ [3])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected [1, 2, 3], got %q", got)
	}
}

func TestAdv_Prec_NegateWithIndex(t *testing.T) {
	got := runE2E(t, "arr = [10, 20]\n|> str(-arr@0)")
	if got != "[-10]" {
		t.Fatalf("expected [-10], got %q", got)
	}
}

// --- Runtime — Verify (2 tests) ---

func TestAdv_Prec_ChainedComparison(t *testing.T) {
	// [1] < [2] yields [1], then [1] < [3] yields [1] — should succeed
	got := runE2E(t, "x = [1] < [2] < [3]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Prec_ComparisonResultInArith(t *testing.T) {
	// [1] < [2] → [1], then [1] + [5] → [6]
	got := runE2E(t, "x = ([1] < [2]) + [5]\n|> str(x)")
	if got != "[6]" {
		t.Fatalf("expected [6], got %q", got)
	}
}

// --- Parse Error (1 test) ---

func TestAdv_Prec_PrefixBinaryOp(t *testing.T) {
	runE2EExpectParseError(t, `x = + [1]`, "")
}
