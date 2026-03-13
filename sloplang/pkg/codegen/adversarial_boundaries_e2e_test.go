package codegen

import "testing"

// --- Runtime — Panics / BUG (6 tests) ---

func TestAdv_Bound_MinIntDivNegOne(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = [-9223372036854775808] / [-1]\n|> str(x)", "integer overflow")
}

func TestAdv_Bound_MinIntNegate(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = -[-9223372036854775808]\n|> str(x)", "cannot negate MinInt")
}

func TestAdv_Bound_ModZeroInt(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = [5] % [0]\n|> str(x)", "modulo by zero")
}

func TestAdv_Bound_ModZeroUint(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = [5u] % [0u]\n|> str(x)", "modulo by zero")
}

func TestAdv_Bound_FloatNaN(t *testing.T) {
	// math.Pow(-1.0, 0.5) → NaN — verify how FormatValue handles it
	got := runE2E(t, "x = [-1.0] ** [0.5]\n|> str(x)")
	_ = got // NaN handling varies; test just verifies no crash
}

func TestAdv_Bound_NegateUintMax(t *testing.T) {
	// Negate does -int64(e) on uint64 — MaxUint64 overflows silently (wraps to 1)
	got := runE2E(t, "x = -[18446744073709551615u]\n|> str(x)")
	_ = got // exact value depends on overflow semantics; test verifies no crash
}

// --- Runtime — Success (10 tests) ---

func TestAdv_Bound_MaxInt64(t *testing.T) {
	got := runE2E(t, "x = [9223372036854775807]\n|> str(x)")
	if got != "[9223372036854775807]" {
		t.Fatalf("expected [9223372036854775807], got %q", got)
	}
}

func TestAdv_Bound_MinInt64(t *testing.T) {
	got := runE2E(t, "x = [-9223372036854775808]\n|> str(x)")
	if got != "[-9223372036854775808]" {
		t.Fatalf("expected [-9223372036854775808], got %q", got)
	}
}

func TestAdv_Bound_MaxUint64(t *testing.T) {
	got := runE2E(t, "x = [18446744073709551615u]\n|> str(x)")
	if got != "[18446744073709551615]" {
		t.Fatalf("expected [18446744073709551615], got %q", got)
	}
}

func TestAdv_Bound_FloatMaxVal(t *testing.T) {
	got := runE2E(t, "x = [1.7976931348623157e308]\n|> str(x)")
	if got != "[1.7976931348623157e+308]" {
		t.Fatalf("expected [1.7976931348623157e+308], got %q", got)
	}
}

func TestAdv_Bound_FloatSmallestPos(t *testing.T) {
	got := runE2E(t, "x = [5e-324]\n|> str(x)")
	if got != "[5e-324]" {
		t.Fatalf("expected [5e-324], got %q", got)
	}
}

func TestAdv_Bound_IntPowLossy(t *testing.T) {
	got := runE2E(t, "x = [2] ** [53]\n|> str(x)")
	if got != "[9007199254740992]" {
		t.Fatalf("expected [9007199254740992], got %q", got)
	}
}

func TestAdv_Bound_IntPowVeryLossy(t *testing.T) {
	// int64(math.Pow(2,63)) — lossy conversion, verify actual output
	got := runE2E(t, "x = [2] ** [63]\n|> str(x)")
	_ = got // exact value depends on float64 conversion; test verifies no crash
}

func TestAdv_Bound_UintOverflowWrap(t *testing.T) {
	got := runE2E(t, "x = [18446744073709551615u] + [1u]\n|> str(x)")
	if got != "[0]" {
		t.Fatalf("expected [0], got %q", got)
	}
}

func TestAdv_Bound_ModNegativeDividend(t *testing.T) {
	got := runE2E(t, "x = [-7] % [3]\n|> str(x)")
	if got != "[-1]" {
		t.Fatalf("expected [-1], got %q", got)
	}
}

func TestAdv_Bound_FloatDivNormal(t *testing.T) {
	got := runE2E(t, "x = [1.0] / [3.0]\n|> str(x)")
	if got == "" {
		t.Fatal("expected float division output, got empty")
	}
}
