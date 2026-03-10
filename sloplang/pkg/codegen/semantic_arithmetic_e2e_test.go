package codegen

import "testing"

// ============================================================
// Semantic Arithmetic E2E Tests
// Covers type consistency, length mismatch, element-wise ops,
// division by zero, negate, mod/pow edge cases, precedence,
// and type mismatch across all arithmetic operators.
// ============================================================

// --- Type consistency (~7 tests) ---

func TestSem_Arith_IntAdd(t *testing.T) {
	got := runE2E(t, `|> str([1] + [2])`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Arith_FloatAdd(t *testing.T) {
	got := runE2E(t, `|> str([1.5] + [2.5])`)
	if got != "[4]" {
		t.Fatalf("expected %q, got %q", "[4]", got)
	}
}

func TestSem_Arith_UintAdd(t *testing.T) {
	got := runE2E(t, `|> str([1u] + [2u])`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Arith_IntFloatPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] + [1.0]`)
}

func TestSem_Arith_IntUintPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] + [1u]`)
}

func TestSem_Arith_FloatUintPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [1.0] + [1u]`)
}

func TestSem_Arith_StringAddPanics(t *testing.T) {
	runE2EExpectPanic(t, `a = "a"
b = "b"
x = a + b`)
}

// --- Length mismatch (~4 tests) ---

func TestSem_Arith_LongPlusShort(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] + [1]`)
}

func TestSem_Arith_ShortPlusLong(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] + [1, 2]`)
}

func TestSem_Arith_EmptyPlusEmpty(t *testing.T) {
	got := runE2E(t, `|> str([] + [])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arith_EmptyPlusOne(t *testing.T) {
	runE2EExpectPanic(t, `x = [] + [1]`)
}

// --- Multi-element element-wise (~6 tests) ---

func TestSem_Arith_AddMulti(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] + [4, 5, 6])`)
	if got != "[5, 7, 9]" {
		t.Fatalf("expected %q, got %q", "[5, 7, 9]", got)
	}
}

func TestSem_Arith_SubMulti(t *testing.T) {
	got := runE2E(t, `|> str([10, 20] - [3, 7])`)
	if got != "[7, 13]" {
		t.Fatalf("expected %q, got %q", "[7, 13]", got)
	}
}

func TestSem_Arith_MulMulti(t *testing.T) {
	got := runE2E(t, `|> str([2, 3] * [4, 5])`)
	if got != "[8, 15]" {
		t.Fatalf("expected %q, got %q", "[8, 15]", got)
	}
}

func TestSem_Arith_DivMulti(t *testing.T) {
	got := runE2E(t, `|> str([10, 6] / [2, 3])`)
	if got != "[5, 2]" {
		t.Fatalf("expected %q, got %q", "[5, 2]", got)
	}
}

func TestSem_Arith_ModMulti(t *testing.T) {
	got := runE2E(t, `|> str([7, 5] % [3, 2])`)
	if got != "[1, 1]" {
		t.Fatalf("expected %q, got %q", "[1, 1]", got)
	}
}

func TestSem_Arith_PowMulti(t *testing.T) {
	got := runE2E(t, `|> str([2, 3] ** [3, 2])`)
	if got != "[8, 9]" {
		t.Fatalf("expected %q, got %q", "[8, 9]", got)
	}
}

// --- Division by zero (~4 tests) ---

func TestSem_Arith_DivByZeroInt(t *testing.T) {
	runE2EExpectPanic(t, `x = [10] / [0]`)
}

func TestSem_Arith_DivByZeroUint(t *testing.T) {
	runE2EExpectPanic(t, `x = [10u] / [0u]`)
}

func TestSem_Arith_DivByZeroFloat(t *testing.T) {
	runE2EExpectPanic(t, `x = [10.0] / [0.0]`)
}

func TestSem_Arith_DivByZeroSecondElem(t *testing.T) {
	runE2EExpectPanic(t, `x = [10, 20] / [5, 0]`)
}

// --- Negate (~8 tests) ---

func TestSem_Arith_Negate_Int(t *testing.T) {
	got := runE2E(t, `|> str(-[1])`)
	if got != "[-1]" {
		t.Fatalf("expected %q, got %q", "[-1]", got)
	}
}

func TestSem_Arith_Negate_Multi(t *testing.T) {
	got := runE2E(t, `|> str(-[1, 2, 3])`)
	if got != "[-1, -2, -3]" {
		t.Fatalf("expected %q, got %q", "[-1, -2, -3]", got)
	}
}

func TestSem_Arith_Negate_Zero(t *testing.T) {
	got := runE2E(t, `|> str(-[0])`)
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestSem_Arith_Negate_Float(t *testing.T) {
	got := runE2E(t, `|> str(-[3.14])`)
	if got != "[-3.14]" {
		t.Fatalf("expected %q, got %q", "[-3.14]", got)
	}
}

func TestSem_Arith_Negate_DoubleNegate(t *testing.T) {
	got := runE2E(t, `|> str(-(-[5]))`)
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestSem_Arith_Negate_Uint(t *testing.T) {
	got := runE2E(t, `|> str(-[1u])`)
	if got != "[-1]" {
		t.Fatalf("expected %q, got %q", "[-1]", got)
	}
}

func TestSem_Arith_Negate_Empty(t *testing.T) {
	got := runE2E(t, `|> str(-[])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arith_Negate_StringPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = "hello"
y = -x`)
}

// --- Mod edge cases (~2 tests) ---

func TestSem_Arith_BasicMod(t *testing.T) {
	got := runE2E(t, `|> str([7] % [3])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arith_FloatModPanics(t *testing.T) {
	runE2EExpectPanic(t, `x = [7.0] % [3.0]`)
}

// --- Pow edge cases (~3 tests) ---

func TestSem_Arith_Pow_ZeroExp(t *testing.T) {
	got := runE2E(t, `|> str([2] ** [0])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arith_Pow_ZeroBase(t *testing.T) {
	got := runE2E(t, `|> str([0] ** [0])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arith_Pow_NegExp(t *testing.T) {
	got := runE2E(t, `x = -[1]
|> str([2] ** x)`)
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

// --- Precedence (~3 tests) ---

func TestSem_Arith_MulBeforeAdd(t *testing.T) {
	got := runE2E(t, `|> str([1] + [2] * [3])`)
	if got != "[7]" {
		t.Fatalf("expected %q, got %q", "[7]", got)
	}
}

func TestSem_Arith_ParensOverride(t *testing.T) {
	got := runE2E(t, `|> str(([1] + [2]) * [3])`)
	if got != "[9]" {
		t.Fatalf("expected %q, got %q", "[9]", got)
	}
}

func TestSem_Arith_NegInExpr(t *testing.T) {
	got := runE2E(t, `|> str(-[1] + [2])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// --- Type mismatch for all ops (~5 tests) ---

func TestSem_Arith_SubTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] - [1.0]`)
}

func TestSem_Arith_MulTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] * [1u]`)
}

func TestSem_Arith_DivTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1.0] / [1u]`)
}

func TestSem_Arith_ModTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] % [1u]`)
}

func TestSem_Arith_PowTypeMismatch(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] ** [1.0]`)
}
