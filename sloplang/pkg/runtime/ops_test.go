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

// Iterate tests

func TestIterate_IntArray(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	items := Iterate(sv)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	for i, item := range items {
		if len(item.Elements) != 1 {
			t.Fatalf("item %d: expected 1 element, got %d", i, len(item.Elements))
		}
	}
	if items[0].Elements[0].(int64) != 1 {
		t.Fatalf("expected 1, got %v", items[0].Elements[0])
	}
}

func TestIterate_Empty(t *testing.T) {
	items := Iterate(NewSlopValue())
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestIterate_Nested(t *testing.T) {
	inner1 := NewSlopValue(int64(1), int64(2))
	inner2 := NewSlopValue(int64(3), int64(4))
	sv := NewSlopValue(inner1, inner2)
	items := Iterate(sv)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	// Nested SlopValues should be returned directly
	if len(items[0].Elements) != 2 {
		t.Fatalf("expected 2 elements in first, got %d", len(items[0].Elements))
	}
}

func TestIterate_StringArray(t *testing.T) {
	sv := NewSlopValue("a", "b")
	items := Iterate(sv)
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Elements[0].(string) != "a" {
		t.Fatalf("expected 'a', got %v", items[0].Elements[0])
	}
}

// UnpackTwo tests

func TestUnpackTwo_Ints(t *testing.T) {
	sv := NewSlopValue(int64(10), int64(20))
	a, b := UnpackTwo(sv)
	if a.Elements[0].(int64) != 10 {
		t.Fatalf("expected 10, got %v", a.Elements[0])
	}
	if b.Elements[0].(int64) != 20 {
		t.Fatalf("expected 20, got %v", b.Elements[0])
	}
}

func TestUnpackTwo_Nested(t *testing.T) {
	inner1 := NewSlopValue(int64(1), int64(2))
	inner2 := NewSlopValue(int64(3), int64(4))
	sv := NewSlopValue(inner1, inner2)
	a, b := UnpackTwo(sv)
	if len(a.Elements) != 2 {
		t.Fatalf("expected 2 elements in a, got %d", len(a.Elements))
	}
	if len(b.Elements) != 2 {
		t.Fatalf("expected 2 elements in b, got %d", len(b.Elements))
	}
}

func TestUnpackTwo_PanicOnTooFew(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got nil")
		}
	}()
	UnpackTwo(NewSlopValue(int64(1)))
}

// ==========================================
// Negative / Edge Case / Boundary Tests
// ==========================================

// --- Arithmetic panic tests ---

func TestSub_LengthMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	Sub(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3)))
}

func TestSub_TypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	Sub(NewSlopValue(int64(1)), NewSlopValue(float64(1.0)))
}

func TestMul_LengthMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	Mul(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3)))
}

func TestMul_TypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	Mul(NewSlopValue(int64(1)), NewSlopValue(uint64(1)))
}

func TestDiv_LengthMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	Div(NewSlopValue(int64(10), int64(6)), NewSlopValue(int64(2)))
}

func TestDiv_TypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	Div(NewSlopValue(int64(10)), NewSlopValue(float64(2.0)))
}

func TestDiv_ByZero_Int(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on division by zero")
		}
		if s, ok := r.(string); !ok || s != "sloplang: division by zero" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	Div(NewSlopValue(int64(10)), NewSlopValue(int64(0)))
}

func TestDiv_ByZero_Uint(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on division by zero")
		}
	}()
	Div(NewSlopValue(uint64(10)), NewSlopValue(uint64(0)))
}

func TestDiv_ByZero_Float(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on division by zero")
		}
	}()
	Div(NewSlopValue(float64(10.0)), NewSlopValue(float64(0.0)))
}

func TestMod_LengthMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	Mod(NewSlopValue(int64(7)), NewSlopValue(int64(3), int64(2)))
}

func TestMod_TypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	Mod(NewSlopValue(int64(7)), NewSlopValue(uint64(3)))
}

func TestMod_FloatPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: mod not supported for float")
		}
	}()
	Mod(NewSlopValue(float64(7.0)), NewSlopValue(float64(3.0)))
}

func TestPow_LengthMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on length mismatch")
		}
	}()
	Pow(NewSlopValue(int64(2)), NewSlopValue(int64(3), int64(2)))
}

func TestPow_TypeMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on type mismatch")
		}
	}()
	Pow(NewSlopValue(int64(2)), NewSlopValue(float64(3.0)))
}

func TestAdd_StringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: + not supported for string")
		}
	}()
	Add(NewSlopValue("a"), NewSlopValue("b"))
}

func TestSub_StringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: - not supported for string")
		}
	}()
	Sub(NewSlopValue("a"), NewSlopValue("b"))
}

func TestMul_StringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: * not supported for string")
		}
	}()
	Mul(NewSlopValue("a"), NewSlopValue("b"))
}

func TestDiv_StringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: / not supported for string")
		}
	}()
	Div(NewSlopValue("a"), NewSlopValue("b"))
}

func TestNegate_StringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: cannot negate string")
		}
	}()
	Negate(NewSlopValue("hello"))
}

// --- Arithmetic with empty arrays ---

func TestAdd_EmptyArrays(t *testing.T) {
	result := Add(NewSlopValue(), NewSlopValue())
	if len(result.Elements) != 0 {
		t.Fatalf("expected empty result, got %d elements", len(result.Elements))
	}
}

func TestSub_EmptyArrays(t *testing.T) {
	result := Sub(NewSlopValue(), NewSlopValue())
	if len(result.Elements) != 0 {
		t.Fatalf("expected empty result, got %d elements", len(result.Elements))
	}
}

func TestNegate_EmptyArray(t *testing.T) {
	result := Negate(NewSlopValue())
	if len(result.Elements) != 0 {
		t.Fatalf("expected empty result, got %d elements", len(result.Elements))
	}
}

// --- Comparison panic tests ---

func TestNeq_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element !=")
		}
	}()
	Neq(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(1), int64(2)))
}

func TestLt_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element <")
		}
	}()
	Lt(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3), int64(4)))
}

func TestGt_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element >")
		}
	}()
	Gt(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3), int64(4)))
}

func TestLte_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element <=")
		}
	}()
	Lte(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3), int64(4)))
}

func TestGte_MultiElement_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element >=")
		}
	}()
	Gte(NewSlopValue(int64(1), int64(2)), NewSlopValue(int64(3), int64(4)))
}

func TestEq_EmptyVsNonEmpty_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: empty vs non-empty comparison")
		}
	}()
	Eq(NewSlopValue(), NewSlopValue(int64(1)))
}

func TestLt_EmptyArray_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: empty array comparison")
		}
	}()
	Lt(NewSlopValue(), NewSlopValue())
}

func TestEq_TypeMismatch_IntVsFloat_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: int64 vs float64 comparison")
		}
	}()
	Eq(NewSlopValue(int64(1)), NewSlopValue(float64(1.0)))
}

func TestEq_TypeMismatch_IntVsString_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: int64 vs string comparison")
		}
	}()
	Eq(NewSlopValue(int64(1)), NewSlopValue("1"))
}

func TestLt_TypeMismatch_IntVsUint_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic: int64 vs uint64 comparison")
		}
	}()
	Lt(NewSlopValue(int64(1)), NewSlopValue(uint64(2)))
}

// --- UnpackTwo edge cases ---

func TestUnpackTwo_EmptyPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on empty unpack")
		}
	}()
	UnpackTwo(NewSlopValue())
}

func TestUnpackTwo_ExtraElementsIgnored(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	a, b := UnpackTwo(sv)
	if a.Elements[0].(int64) != 1 {
		t.Fatalf("expected 1, got %v", a.Elements[0])
	}
	if b.Elements[0].(int64) != 2 {
		t.Fatalf("expected 2, got %v", b.Elements[0])
	}
}

func TestUnpackTwo_MixedTypes(t *testing.T) {
	sv := NewSlopValue(int64(42), "hello")
	a, b := UnpackTwo(sv)
	if a.Elements[0].(int64) != 42 {
		t.Fatalf("expected 42, got %v", a.Elements[0])
	}
	if b.Elements[0].(string) != "hello" {
		t.Fatalf("expected 'hello', got %v", b.Elements[0])
	}
}

// --- IsTruthy edge cases ---

func TestIsTruthy_EmptyArray(t *testing.T) {
	if NewSlopValue().IsTruthy() {
		t.Fatal("empty SlopValue should be falsy")
	}
}

func TestIsTruthy_Zero(t *testing.T) {
	if !NewSlopValue(int64(0)).IsTruthy() {
		t.Fatal("[0] should be truthy")
	}
}

func TestIsTruthy_EmptyString(t *testing.T) {
	if !NewSlopValue("").IsTruthy() {
		t.Fatal("[\"\"] should be truthy (non-empty array)")
	}
}

func TestIsTruthy_MultiElement(t *testing.T) {
	if !NewSlopValue(int64(1), int64(2)).IsTruthy() {
		t.Fatal("[1, 2] should be truthy")
	}
}

// --- FormatValue / Str edge cases ---

func TestFormatValue_NestedEmpty(t *testing.T) {
	inner := NewSlopValue()
	outer := NewSlopValue(inner)
	got := FormatValue(outer)
	if got != "[]" {
		t.Fatalf("expected '[]', got %q", got)
	}
}

func TestFormatValue_DeeplyNested(t *testing.T) {
	inner := NewSlopValue(int64(1))
	mid := NewSlopValue(inner)
	outer := NewSlopValue(mid, int64(2))
	got := FormatValue(outer)
	if got != "[1, 2]" {
		t.Fatalf("expected '[1, 2]', got %q", got)
	}
}

func TestStr_StringValue(t *testing.T) {
	result := Str(NewSlopValue("hello"))
	s := result.Elements[0].(string)
	if s != "hello" {
		t.Fatalf("expected 'hello', got %q", s)
	}
}

func TestStr_MixedArray(t *testing.T) {
	result := Str(NewSlopValue(int64(1), "hi", float64(3.14)))
	s := result.Elements[0].(string)
	if s != "[1, hi, 3.14]" {
		t.Fatalf("expected '[1, hi, 3.14]', got %q", s)
	}
}

// --- Negate edge cases ---

func TestNegate_Uint(t *testing.T) {
	result := Negate(NewSlopValue(uint64(5)))
	// uint64 negate converts to -int64
	if v, ok := result.Elements[0].(int64); !ok || v != -5 {
		t.Fatalf("expected -5 (int64), got %v (%T)", result.Elements[0], result.Elements[0])
	}
}

func TestNegate_NestedArray(t *testing.T) {
	inner := NewSlopValue(int64(1), int64(2))
	outer := NewSlopValue(inner)
	result := Negate(outer)
	// Negate on nested SlopValue should negate recursively
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
}

// --- Iterate edge cases ---

func TestIterate_SingleElement(t *testing.T) {
	sv := NewSlopValue(int64(42))
	items := Iterate(sv)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Elements[0].(int64) != 42 {
		t.Fatalf("expected 42, got %v", items[0].Elements[0])
	}
}

// ==========================================
// Phase 4: Array Operation Tests
// ==========================================

// --- Index tests ---

func TestIndex_Int(t *testing.T) {
	sv := NewSlopValue(int64(10), int64(20), int64(30))
	result := Index(sv, NewSlopValue(int64(1)))
	if result.Elements[0].(int64) != 20 {
		t.Fatalf("expected 20, got %v", result.Elements[0])
	}
}

func TestIndex_Nested(t *testing.T) {
	inner := NewSlopValue(int64(1), int64(2))
	sv := NewSlopValue(inner, int64(3))
	result := Index(sv, NewSlopValue(int64(0)))
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
}

func TestIndex_OutOfBounds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on out of bounds")
		}
	}()
	sv := NewSlopValue(int64(1), int64(2))
	Index(sv, NewSlopValue(int64(5)))
}

func TestIndex_NegativeIndex(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on negative index")
		}
	}()
	sv := NewSlopValue(int64(1))
	Index(sv, NewSlopValue(int64(-1)))
}

// --- IndexSet tests ---

func TestIndexSet_Basic(t *testing.T) {
	sv := NewSlopValue(int64(10), int64(20), int64(30))
	val := NewSlopValue(int64(99))
	result := IndexSet(sv, NewSlopValue(int64(1)), val)
	// Should mutate sv
	if result != sv {
		t.Fatal("expected IndexSet to return same pointer")
	}
	nested, ok := sv.Elements[1].(*SlopValue)
	if !ok {
		t.Fatalf("expected *SlopValue at index 1, got %T", sv.Elements[1])
	}
	if nested.Elements[0].(int64) != 99 {
		t.Fatalf("expected 99, got %v", nested.Elements[0])
	}
}

func TestIndexSet_OutOfBounds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on out of bounds")
		}
	}()
	sv := NewSlopValue(int64(1))
	IndexSet(sv, NewSlopValue(int64(5)), NewSlopValue(int64(99)))
}

// --- Length tests ---

func TestLength_Basic(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	result := Length(sv)
	if result.Elements[0].(int64) != 3 {
		t.Fatalf("expected 3, got %v", result.Elements[0])
	}
}

func TestLength_Empty(t *testing.T) {
	result := Length(NewSlopValue())
	if result.Elements[0].(int64) != 0 {
		t.Fatalf("expected 0, got %v", result.Elements[0])
	}
}

// --- Push tests ---

func TestPush_Basic(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2))
	result := Push(sv, NewSlopValue(int64(3)))
	if result != sv {
		t.Fatal("expected Push to return same pointer")
	}
	if len(sv.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(sv.Elements))
	}
	if sv.Elements[2].(int64) != 3 {
		t.Fatalf("expected 3, got %v", sv.Elements[2])
	}
}

func TestPush_MultipleElements(t *testing.T) {
	sv := NewSlopValue(int64(1))
	Push(sv, NewSlopValue(int64(2), int64(3)))
	if len(sv.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(sv.Elements))
	}
}

// --- Pop tests ---

func TestPop_Basic(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	result := Pop(sv)
	if result.Elements[0].(int64) != 3 {
		t.Fatalf("expected 3, got %v", result.Elements[0])
	}
	if len(sv.Elements) != 2 {
		t.Fatalf("expected 2 elements remaining, got %d", len(sv.Elements))
	}
}

func TestPop_Nested(t *testing.T) {
	inner := NewSlopValue(int64(10), int64(20))
	sv := NewSlopValue(int64(1), inner)
	result := Pop(sv)
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements in popped value, got %d", len(result.Elements))
	}
}

func TestPop_Empty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on pop from empty array")
		}
	}()
	Pop(NewSlopValue())
}

// --- RemoveAt tests ---

func TestRemoveAt_Basic(t *testing.T) {
	sv := NewSlopValue(int64(10), int64(20), int64(30))
	result := RemoveAt(sv, NewSlopValue(int64(1)))
	if result.Elements[0].(int64) != 20 {
		t.Fatalf("expected removed value 20, got %v", result.Elements[0])
	}
	if len(sv.Elements) != 2 {
		t.Fatalf("expected 2 elements remaining, got %d", len(sv.Elements))
	}
	if sv.Elements[0].(int64) != 10 || sv.Elements[1].(int64) != 30 {
		t.Fatalf("expected [10, 30], got %v", sv.Elements)
	}
}

func TestRemoveAt_OutOfBounds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on out of bounds")
		}
	}()
	sv := NewSlopValue(int64(1))
	RemoveAt(sv, NewSlopValue(int64(5)))
}

// --- Slice tests ---

func TestSlice_Basic(t *testing.T) {
	sv := NewSlopValue(int64(10), int64(20), int64(30), int64(40))
	result := Slice(sv, NewSlopValue(int64(1)), NewSlopValue(int64(3)))
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
	if result.Elements[0].(int64) != 20 || result.Elements[1].(int64) != 30 {
		t.Fatalf("expected [20, 30], got %v", result.Elements)
	}
}

func TestSlice_Full(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2))
	result := Slice(sv, NewSlopValue(int64(0)), NewSlopValue(int64(2)))
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
}

func TestSlice_OutOfBounds(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on out of bounds slice")
		}
	}()
	sv := NewSlopValue(int64(1))
	Slice(sv, NewSlopValue(int64(0)), NewSlopValue(int64(5)))
}

func TestSlice_DoesNotMutate(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	result := Slice(sv, NewSlopValue(int64(0)), NewSlopValue(int64(2)))
	// Mutating the slice result should not affect original
	result.Elements[0] = int64(99)
	if sv.Elements[0].(int64) != 1 {
		t.Fatal("Slice mutated the original array")
	}
}

// --- Concat tests ---

func TestConcat_Basic(t *testing.T) {
	a := NewSlopValue(int64(1), int64(2))
	b := NewSlopValue(int64(3), int64(4))
	result := Concat(a, b)
	if len(result.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(result.Elements))
	}
	if result.Elements[0].(int64) != 1 || result.Elements[3].(int64) != 4 {
		t.Fatalf("unexpected result: %v", result.Elements)
	}
}

func TestConcat_DoesNotMutate(t *testing.T) {
	a := NewSlopValue(int64(1))
	b := NewSlopValue(int64(2))
	Concat(a, b)
	if len(a.Elements) != 1 {
		t.Fatal("Concat mutated original a")
	}
	if len(b.Elements) != 1 {
		t.Fatal("Concat mutated original b")
	}
}

func TestConcat_Empty(t *testing.T) {
	result := Concat(NewSlopValue(), NewSlopValue(int64(1)))
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
}

// --- Remove tests ---

func TestRemove_Basic(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3), int64(2))
	result := Remove(sv, NewSlopValue(int64(2)))
	if len(result.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result.Elements))
	}
	// Should remove first occurrence only
	if result.Elements[0].(int64) != 1 || result.Elements[1].(int64) != 3 || result.Elements[2].(int64) != 2 {
		t.Fatalf("expected [1, 3, 2], got %v", result.Elements)
	}
}

func TestRemove_NotFound(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2))
	result := Remove(sv, NewSlopValue(int64(99)))
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
}

func TestRemove_DoesNotMutate(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	Remove(sv, NewSlopValue(int64(2)))
	if len(sv.Elements) != 3 {
		t.Fatal("Remove mutated the original array")
	}
}

// --- Contains tests ---

func TestContains_Found(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	result := Contains(sv, NewSlopValue(int64(2)))
	if !result.IsTruthy() {
		t.Fatal("expected truthy (found)")
	}
}

func TestContains_NotFound(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	result := Contains(sv, NewSlopValue(int64(99)))
	if result.IsTruthy() {
		t.Fatal("expected falsy (not found)")
	}
}

func TestContains_String(t *testing.T) {
	sv := NewSlopValue("a", "b", "c")
	result := Contains(sv, NewSlopValue("b"))
	if !result.IsTruthy() {
		t.Fatal("expected truthy (found string)")
	}
}

// --- Unique tests ---

func TestUnique_Basic(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(1), int64(3), int64(2))
	result := Unique(sv)
	if len(result.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result.Elements))
	}
	if result.Elements[0].(int64) != 1 || result.Elements[1].(int64) != 2 || result.Elements[2].(int64) != 3 {
		t.Fatalf("expected [1, 2, 3], got %v", result.Elements)
	}
}

func TestUnique_NoDuplicates(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	result := Unique(sv)
	if len(result.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result.Elements))
	}
}

func TestUnique_Empty(t *testing.T) {
	result := Unique(NewSlopValue())
	if len(result.Elements) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(result.Elements))
	}
}

func TestUnique_DoesNotMutate(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(1))
	Unique(sv)
	if len(sv.Elements) != 2 {
		t.Fatal("Unique mutated the original array")
	}
}

// --- deepEqual tests ---

func TestDeepEqual_Ints(t *testing.T) {
	if !deepEqual(int64(5), int64(5)) {
		t.Fatal("expected true")
	}
	if deepEqual(int64(5), int64(6)) {
		t.Fatal("expected false")
	}
}

func TestDeepEqual_Strings(t *testing.T) {
	if !deepEqual("abc", "abc") {
		t.Fatal("expected true")
	}
	if deepEqual("abc", "def") {
		t.Fatal("expected false")
	}
}

func TestDeepEqual_DifferentTypes(t *testing.T) {
	if deepEqual(int64(1), "1") {
		t.Fatal("expected false for different types")
	}
}

func TestDeepEqual_NestedSlopValue(t *testing.T) {
	a := NewSlopValue(int64(1), int64(2))
	b := NewSlopValue(int64(1), int64(2))
	c := NewSlopValue(int64(1), int64(3))
	if !deepEqual(a, b) {
		t.Fatal("expected true for equal nested values")
	}
	if deepEqual(a, c) {
		t.Fatal("expected false for different nested values")
	}
}

// ==========================================
// Phase 5: Hashmap Tests
// ==========================================

// --- MapFromKeysValues tests ---

func TestMapFromKeysValues_Basic(t *testing.T) {
	vals := NewSlopValue("bob", NewSlopValue(int64(30)))
	m := MapFromKeysValues([]string{"name", "age"}, vals)
	if len(m.Keys) != 2 || m.Keys[0] != "name" || m.Keys[1] != "age" {
		t.Fatalf("expected keys [name, age], got %v", m.Keys)
	}
	if len(m.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(m.Elements))
	}
	if m.Elements[0].(string) != "bob" {
		t.Fatalf("expected 'bob', got %v", m.Elements[0])
	}
}

func TestMapFromKeysValues_Empty(t *testing.T) {
	m := MapFromKeysValues([]string{}, NewSlopValue())
	if len(m.Keys) != 0 || len(m.Elements) != 0 {
		t.Fatalf("expected empty hashmap, got keys=%v elements=%v", m.Keys, m.Elements)
	}
}

func TestMapFromKeysValues_Mismatch(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on key/value count mismatch")
		}
	}()
	MapFromKeysValues([]string{"a", "b"}, NewSlopValue("x"))
}

// --- IndexKeyStr tests ---

func TestIndexKeyStr_Found(t *testing.T) {
	m := MapFromKeysValues([]string{"name", "age"}, NewSlopValue("bob", NewSlopValue(int64(30))))
	result := IndexKeyStr(m, "name")
	if len(result.Elements) != 1 || result.Elements[0].(string) != "bob" {
		t.Fatalf("expected 'bob', got %v", result)
	}
}

func TestIndexKeyStr_FoundNested(t *testing.T) {
	inner := NewSlopValue(int64(30))
	m := MapFromKeysValues([]string{"name", "age"}, NewSlopValue("bob", inner))
	result := IndexKeyStr(m, "age")
	if result != inner {
		t.Fatalf("expected same *SlopValue, got different one")
	}
}

func TestIndexKeyStr_NotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on key not found")
		}
	}()
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	IndexKeyStr(m, "missing")
}

// --- IndexKey tests ---

func TestIndexKey_Basic(t *testing.T) {
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	result := IndexKey(m, NewSlopValue("name"))
	if result.Elements[0].(string) != "bob" {
		t.Fatalf("expected 'bob', got %v", result.Elements[0])
	}
}

func TestIndexKey_NonStringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on non-string key")
		}
	}()
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	IndexKey(m, NewSlopValue(int64(1)))
}

func TestIndexKey_MultiElementPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on multi-element key")
		}
	}()
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	IndexKey(m, NewSlopValue("a", "b"))
}

// --- IndexKeySetStr tests ---

func TestIndexKeySetStr_Update(t *testing.T) {
	m := MapFromKeysValues([]string{"name", "age"}, NewSlopValue("bob", NewSlopValue(int64(30))))
	newVal := NewSlopValue(int64(31))
	result := IndexKeySetStr(m, "age", newVal)
	if result != m {
		t.Fatal("expected same pointer returned")
	}
	// Check the element was updated
	if m.Elements[1].(*SlopValue) != newVal {
		t.Fatalf("expected updated value, got %v", m.Elements[1])
	}
}

func TestIndexKeySetStr_AddNew(t *testing.T) {
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	newVal := NewSlopValue("bob@test.com")
	IndexKeySetStr(m, "email", newVal)
	if len(m.Keys) != 2 || m.Keys[1] != "email" {
		t.Fatalf("expected keys [name, email], got %v", m.Keys)
	}
	if len(m.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(m.Elements))
	}
	if m.Elements[1].(*SlopValue) != newVal {
		t.Fatalf("expected new value at index 1")
	}
}

// --- IndexKeySet tests ---

func TestIndexKeySet_Basic(t *testing.T) {
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	newVal := NewSlopValue("alice")
	result := IndexKeySet(m, NewSlopValue("name"), newVal)
	if result != m {
		t.Fatal("expected same pointer returned")
	}
	if m.Elements[0].(*SlopValue) != newVal {
		t.Fatalf("expected updated value")
	}
}

func TestIndexKeySet_NonStringPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on non-string key")
		}
	}()
	m := MapFromKeysValues([]string{"name"}, NewSlopValue("bob"))
	IndexKeySet(m, NewSlopValue(int64(1)), NewSlopValue("alice"))
}

// --- MapKeys tests ---

func TestMapKeys_Basic(t *testing.T) {
	m := MapFromKeysValues([]string{"name", "age"}, NewSlopValue("bob", NewSlopValue(int64(30))))
	result := MapKeys(m)
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
	if result.Elements[0].(string) != "name" || result.Elements[1].(string) != "age" {
		t.Fatalf("expected [name, age], got %v", result.Elements)
	}
}

func TestMapKeys_NilKeys(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2))
	result := MapKeys(sv)
	if len(result.Elements) != 0 {
		t.Fatalf("expected empty result for plain array, got %d elements", len(result.Elements))
	}
}

func TestMapKeys_Empty(t *testing.T) {
	m := MapFromKeysValues([]string{}, NewSlopValue())
	result := MapKeys(m)
	if len(result.Elements) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(result.Elements))
	}
}

// --- MapValues tests ---

func TestMapValues_Basic(t *testing.T) {
	inner := NewSlopValue(int64(30))
	m := MapFromKeysValues([]string{"name", "age"}, NewSlopValue("bob", inner))
	result := MapValues(m)
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
	if result.Keys != nil {
		t.Fatal("expected nil keys on result")
	}
	if result.Elements[0].(string) != "bob" {
		t.Fatalf("expected 'bob', got %v", result.Elements[0])
	}
}

func TestMapValues_NilKeys(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2))
	result := MapValues(sv)
	if len(result.Elements) != 0 {
		t.Fatalf("expected empty result for plain array, got %d elements", len(result.Elements))
	}
}

func TestMapValues_Empty(t *testing.T) {
	m := MapFromKeysValues([]string{}, NewSlopValue())
	result := MapValues(m)
	if len(result.Elements) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(result.Elements))
	}
}

func TestIterate_MixedTypes(t *testing.T) {
	sv := NewSlopValue(int64(1), "hello", float64(3.14))
	items := Iterate(sv)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[1].Elements[0].(string) != "hello" {
		t.Fatalf("expected 'hello', got %v", items[1].Elements[0])
	}
}

// ==========================================
// Null Value Tests
// ==========================================

func TestFormatValue_Null(t *testing.T) {
	got := FormatValue(NewSlopValue(SlopNull{}))
	if got != "null" {
		t.Fatalf("expected 'null', got %q", got)
	}
}

func TestFormatValue_MixedWithNull(t *testing.T) {
	got := FormatValue(NewSlopValue(int64(1), SlopNull{}, "hi"))
	if got != "[1, null, hi]" {
		t.Fatalf("expected '[1, null, hi]', got %q", got)
	}
}

func TestIsTruthy_Null(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on null truthiness check")
		}
		if s, ok := r.(string); !ok || s != "sloplang: cannot use null as boolean" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	NewSlopValue(SlopNull{}).IsTruthy()
}

func TestEq_NullNull(t *testing.T) {
	result := Eq(NewSlopValue(SlopNull{}), NewSlopValue(SlopNull{}))
	if len(result.Elements) == 0 {
		t.Fatal("expected truthy (null == null), got falsy")
	}
}

func TestEq_NullNonNull(t *testing.T) {
	result := Eq(NewSlopValue(SlopNull{}), NewSlopValue(int64(1)))
	if len(result.Elements) != 0 {
		t.Fatal("expected falsy (null != non-null), got truthy")
	}
}

func TestNeq_NullNonNull(t *testing.T) {
	result := Neq(NewSlopValue(SlopNull{}), NewSlopValue(int64(5)))
	if len(result.Elements) == 0 {
		t.Fatal("expected truthy (null != non-null), got falsy")
	}
}

func TestLt_WithNull(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on null ordered comparison")
		}
		if s, ok := r.(string); !ok || s != "sloplang: cannot compare null with ordered operators" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	Lt(NewSlopValue(SlopNull{}), NewSlopValue(int64(1)))
}

func TestAdd_WithNull(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on arithmetic with null")
		}
		if s, ok := r.(string); !ok || s != "sloplang: cannot perform arithmetic on null" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	Add(NewSlopValue(SlopNull{}), NewSlopValue(int64(1)))
}

func TestNegate_Null(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on negating null")
		}
		if s, ok := r.(string); !ok || s != "sloplang: cannot negate null" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	Negate(NewSlopValue(SlopNull{}))
}

func TestDeepEqual_NullNull(t *testing.T) {
	if !deepEqual(SlopNull{}, SlopNull{}) {
		t.Fatal("expected true: null == null")
	}
}

func TestDeepEqual_NullNonNull(t *testing.T) {
	if deepEqual(SlopNull{}, int64(1)) {
		t.Fatal("expected false: null != int64")
	}
	if deepEqual(int64(1), SlopNull{}) {
		t.Fatal("expected false: int64 != null")
	}
}

func TestContains_Null(t *testing.T) {
	sv := NewSlopValue(int64(1), SlopNull{})
	result := Contains(sv, NewSlopValue(SlopNull{}))
	if len(result.Elements) == 0 {
		t.Fatal("expected truthy: array contains null")
	}
}

func TestIterate_Null(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic on iterating null")
		}
		if s, ok := r.(string); !ok || s != "sloplang: cannot iterate over null" {
			t.Fatalf("unexpected panic message: %v", r)
		}
	}()
	Iterate(NewSlopValue(SlopNull{}))
}

func TestIterate_ArrayWithNullElement(t *testing.T) {
	// [1, null, 3] should iterate fine — null is just an element
	sv := NewSlopValue(int64(1), SlopNull{}, int64(3))
	items := Iterate(sv)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if _, ok := items[1].Elements[0].(SlopNull); !ok {
		t.Fatalf("expected SlopNull at index 1, got %T", items[1].Elements[0])
	}
}
