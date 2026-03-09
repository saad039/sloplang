package runtime

import "testing"

// Arithmetic tests

func TestAdd_IntArrays(t *testing.T) {
	result := Add(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3), int64(4)))
	if got := FormatValue(result); got != "[4, 6]" {
		t.Fatalf("expected [4, 6], got %s", got)
	}
}

func TestAdd_SingleElement(t *testing.T) {
	result := Add(NewSlopValue(int64(1)), NewSlopValue(int64(1)))
	if got := FormatValue(result); got != "2" {
		t.Fatalf("expected 2, got %s", got)
	}
}

func TestAdd_Float(t *testing.T) {
	result := Add(NewSlopValue(float64(3.14)), NewSlopValue(float64(2.86)))
	if got := FormatValue(result); got != "6" {
		t.Fatalf("expected 6, got %s", got)
	}
}

func TestAdd_Uint(t *testing.T) {
	result := Add(NewSlopValue(uint64(42)), NewSlopValue(uint64(8)))
	if got := FormatValue(result); got != "50" {
		t.Fatalf("expected 50, got %s", got)
	}
}

func TestSub(t *testing.T) {
	result := Sub(NewSlopValue(int64(5), int64(3)), NewSlopValue(int64(1), int64(1)))
	if got := FormatValue(result); got != "[4, 2]" {
		t.Fatalf("expected [4, 2], got %s", got)
	}
}

func TestMul(t *testing.T) {
	result := Mul(NewSlopValue(int64(2), int64(3)), NewSlopValue(int64(4), int64(5)))
	if got := FormatValue(result); got != "[8, 15]" {
		t.Fatalf("expected [8, 15], got %s", got)
	}
}

func TestDiv(t *testing.T) {
	result := Div(NewSlopValue(int64(10), int64(6)), NewSlopValue(int64(2), int64(3)))
	if got := FormatValue(result); got != "[5, 2]" {
		t.Fatalf("expected [5, 2], got %s", got)
	}
}

func TestMod(t *testing.T) {
	result := Mod(NewSlopValue(int64(7), int64(5)), NewSlopValue(int64(3), int64(2)))
	if got := FormatValue(result); got != "[1, 1]" {
		t.Fatalf("expected [1, 1], got %s", got)
	}
}

func TestPow(t *testing.T) {
	result := Pow(NewSlopValue(int64(2), int64(3)), NewSlopValue(int64(3), int64(2)))
	if got := FormatValue(result); got != "[8, 9]" {
		t.Fatalf("expected [8, 9], got %s", got)
	}
}

func TestNegate(t *testing.T) {
	result := Negate(NewSlopValue(int64(1), int64(2), int64(3)))
	if got := FormatValue(result); got != "[-1, -2, -3]" {
		t.Fatalf("expected [-1, -2, -3], got %s", got)
	}
}

func TestNegate_Zero(t *testing.T) {
	result := Negate(NewSlopValue(int64(0)))
	if got := FormatValue(result); got != "0" {
		t.Fatalf("expected 0, got %s", got)
	}
}

func TestNegate_Float(t *testing.T) {
	result := Negate(NewSlopValue(float64(3.14)))
	if got := FormatValue(result); got != "-3.14" {
		t.Fatalf("expected -3.14, got %s", got)
	}
}

func TestAdd_LengthMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	Add(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3)))
}

func TestAdd_TypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	Add(NewSlopValue(int64(1)), NewSlopValue(float64(1.0)))
}

func TestMul_ByZero(t *testing.T) {
	result := Mul(NewSlopValue(int64(1)), NewSlopValue(int64(0)))
	if got := FormatValue(result); got != "0" {
		t.Fatalf("expected 0, got %s", got)
	}
}

// Comparison tests

func TestEq_True(t *testing.T) {
	result := Eq(NewSlopValue(int64(2)), NewSlopValue(int64(2)))
	if got := FormatValue(result); got != "1" {
		t.Fatalf("expected 1, got %s", got)
	}
}

func TestEq_False(t *testing.T) {
	result := Eq(NewSlopValue(int64(2)), NewSlopValue(int64(3)))
	if got := FormatValue(result); got != "[]" {
		t.Fatalf("expected [], got %s", got)
	}
}

func TestEq_String(t *testing.T) {
	result := Eq(NewSlopValue("abc"), NewSlopValue("abc"))
	if got := FormatValue(result); got != "1" {
		t.Fatalf("expected 1, got %s", got)
	}
}

func TestNeq(t *testing.T) {
	result := Neq(NewSlopValue(int64(2)), NewSlopValue(int64(3)))
	if got := FormatValue(result); got != "1" {
		t.Fatalf("expected 1, got %s", got)
	}
}

func TestLt(t *testing.T) {
	result := Lt(NewSlopValue(int64(1)), NewSlopValue(int64(2)))
	if got := FormatValue(result); got != "1" {
		t.Fatalf("expected 1, got %s", got)
	}
}

func TestGt(t *testing.T) {
	result := Gt(NewSlopValue(int64(2)), NewSlopValue(int64(1)))
	if got := FormatValue(result); got != "1" {
		t.Fatalf("expected 1, got %s", got)
	}
}

func TestLte_Equal(t *testing.T) {
	result := Lte(NewSlopValue(int64(1)), NewSlopValue(int64(1)))
	if got := FormatValue(result); got != "1" {
		t.Fatalf("expected 1, got %s", got)
	}
}

func TestGte_Less(t *testing.T) {
	result := Gte(NewSlopValue(int64(1)), NewSlopValue(int64(2)))
	if got := FormatValue(result); got != "[]" {
		t.Fatalf("expected [], got %s", got)
	}
}

func TestEq_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element comparison")
		}
	}()
	Eq(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(1), int64(2)))
}

// Logical tests

func TestAnd_BothTruthy(t *testing.T) {
	result := And(NewSlopValue(int64(1)), NewSlopValue(int64(1)))
	if !result.IsTruthy() {
		t.Fatal("expected truthy")
	}
}

func TestAnd_LeftFalsy(t *testing.T) {
	result := And(NewSlopValue(), NewSlopValue(int64(1)))
	if result.IsTruthy() {
		t.Fatal("expected falsy")
	}
}

func TestOr_LeftFalsy(t *testing.T) {
	result := Or(NewSlopValue(), NewSlopValue(int64(1)))
	if !result.IsTruthy() {
		t.Fatal("expected truthy")
	}
}

func TestOr_BothFalsy(t *testing.T) {
	result := Or(NewSlopValue(), NewSlopValue())
	if result.IsTruthy() {
		t.Fatal("expected falsy")
	}
}

func TestNot_Truthy(t *testing.T) {
	result := Not(NewSlopValue(int64(1)))
	if result.IsTruthy() {
		t.Fatal("expected falsy")
	}
}

func TestNot_Falsy(t *testing.T) {
	result := Not(NewSlopValue())
	if !result.IsTruthy() {
		t.Fatal("expected truthy")
	}
}

// Str tests

func TestStr(t *testing.T) {
	result := Str(NewSlopValue(int64(1), int64(2), int64(3)))
	s, ok := result.Elements[0].(string)
	if !ok || s != "[1, 2, 3]" {
		t.Fatalf("expected '[1, 2, 3]', got %v", result.Elements[0])
	}
}

func TestStr_SingleElement(t *testing.T) {
	result := Str(NewSlopValue(int64(42)))
	s := result.Elements[0].(string)
	if s != "42" {
		t.Fatalf("expected '42', got %q", s)
	}
}

func TestStr_Empty(t *testing.T) {
	result := Str(NewSlopValue())
	s := result.Elements[0].(string)
	if s != "[]" {
		t.Fatalf("expected '[]', got %q", s)
	}
}
