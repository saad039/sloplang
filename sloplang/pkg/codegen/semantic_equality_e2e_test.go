package codegen

import "testing"

// ============================================================
// Semantic Equality E2E Tests
// Covers ==, !=, <, >, <=, >= with deep structural equality
// ============================================================

// --- Same-type scalar (~6 tests) ---

func TestSem_Eq_IntEqual(t *testing.T) {
	got := runE2E(t, `x = [1] == [1]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_IntNotEqual(t *testing.T) {
	got := runE2E(t, `x = [1] == [2]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_FloatEqual(t *testing.T) {
	got := runE2E(t, `x = [3.14] == [3.14]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_UintEqual(t *testing.T) {
	got := runE2E(t, `x = [42u] == [42u]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_StringEqual(t *testing.T) {
	got := runE2E(t, `a = "hello"
b = "hello"
x = a == b
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_StringNotEqual(t *testing.T) {
	got := runE2E(t, `a = "a"
b = "b"
x = a == b
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Cross-type (returns [], no panic) (~3 tests) ---

func TestSem_Eq_IntVsFloat(t *testing.T) {
	got := runE2E(t, `x = [1] == [1.0]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_IntVsUint(t *testing.T) {
	got := runE2E(t, `x = [1] == [1u]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_IntVsString(t *testing.T) {
	got := runE2E(t, `a = [1]
b = "1"
x = a == b
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Multi-element deep (~5 tests) ---

func TestSem_Eq_MultiEqual(t *testing.T) {
	got := runE2E(t, `x = [1, 2, 3] == [1, 2, 3]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_MultiNotEqual(t *testing.T) {
	got := runE2E(t, `x = [1, 2, 3] == [1, 2, 4]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_DifferentLengths(t *testing.T) {
	got := runE2E(t, `x = [1, 2] == [1, 2, 3]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_EmptyEqual(t *testing.T) {
	got := runE2E(t, `x = [] == []
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_EmptyVsNonEmpty(t *testing.T) {
	got := runE2E(t, `x = [] == [1]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Nested array (~3 tests) ---

func TestSem_Eq_NestedEqual(t *testing.T) {
	got := runE2E(t, `a = [[1, 2], [3, 4]]
b = [[1, 2], [3, 4]]
x = a == b
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_NestedNotEqual(t *testing.T) {
	got := runE2E(t, `a = [[1, 2], [3, 4]]
b = [[1, 2], [3, 5]]
x = a == b
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_NestedVsFlat(t *testing.T) {
	got := runE2E(t, `a = [[1]]
b = [1]
x = a == b
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Null (~6 tests) ---

func TestSem_Eq_NullNull(t *testing.T) {
	got := runE2E(t, `x = [null] == [null]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_NullNeqNull(t *testing.T) {
	got := runE2E(t, `x = [null] != [null]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_NullVsInt(t *testing.T) {
	got := runE2E(t, `x = [null] == [1]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_NullNeqInt(t *testing.T) {
	got := runE2E(t, `x = [null] != [1]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_ArrayWithNull(t *testing.T) {
	got := runE2E(t, `x = [1, null] == [1, null]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_NullOrderMatters(t *testing.T) {
	got := runE2E(t, `x = [null, 1] == [1, null]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- Hashmap (~4 tests) ---

func TestSem_Eq_SameMap(t *testing.T) {
	got := runE2E(t, `m1{name, age} = ["bob", [30]]
m2{name, age} = ["bob", [30]]
x = m1 == m2
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_DifferentValues(t *testing.T) {
	got := runE2E(t, `m3{name, age} = ["bob", [30]]
m4{name, age} = ["bob", [31]]
x = m3 == m4
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_DifferentKeys(t *testing.T) {
	got := runE2E(t, `m5{name, age} = ["bob", [30]]
m6{name, score} = ["bob", [30]]
x = m5 == m6
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_EmptyMaps(t *testing.T) {
	got := runE2E(t, `x = [] == []
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// --- Inequality != (~4 tests) ---

func TestSem_Eq_NeqTrue(t *testing.T) {
	got := runE2E(t, `x = [1] != [2]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_NeqFalse(t *testing.T) {
	got := runE2E(t, `x = [1] != [1]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_NeqMulti(t *testing.T) {
	got := runE2E(t, `x = [1, 2] != [1, 3]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_NeqEmpty(t *testing.T) {
	got := runE2E(t, `x = [] != [1]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// --- In boolean context (~2 tests) ---

func TestSem_Eq_EqInIf(t *testing.T) {
	got := runE2E(t, `if [1, 2] == [1, 2] {
	|> "yes"
}`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

func TestSem_Eq_NeqInIf(t *testing.T) {
	got := runE2E(t, `if [1, 2] != [1, 3] {
	|> "yes"
}`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

// --- After mutation (~3 tests) ---

func TestSem_Eq_AfterIndexSet(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr@0 = [5]
expected = [5, 2, 3]
x = arr == expected
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_AfterPush(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr << [3]
expected = [1, 2, 3]
x = arr == expected
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_BuildTwoSeparately(t *testing.T) {
	got := runE2E(t, `a = [1, 2]
a << [3]
b = [1, 2]
b << [3]
x = a == b
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// --- Ordered comparisons (~10 tests) ---

func TestSem_Eq_LtTrue(t *testing.T) {
	got := runE2E(t, `x = [1] < [2]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_LtFalse(t *testing.T) {
	got := runE2E(t, `x = [2] < [1]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_GtTrue(t *testing.T) {
	got := runE2E(t, `x = [2] > [1]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_LteEqual(t *testing.T) {
	got := runE2E(t, `x = [1] <= [1]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_GteSmaller(t *testing.T) {
	got := runE2E(t, `x = [1] >= [2]
|> str(x)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Eq_StringOrder(t *testing.T) {
	got := runE2E(t, `a = "a"
b = "b"
x = a < b
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_FloatOrder(t *testing.T) {
	got := runE2E(t, `x = [1.5] < [2.5]
|> str(x)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Eq_MultiElement_Panics(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] < [3, 4]
|> str(x)`)
}

func TestSem_Eq_NullOrdered_Panics(t *testing.T) {
	runE2EExpectPanic(t, `x = [null] < [1]
|> str(x)`)
}

func TestSem_Eq_TypeMismatch_Panics(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] < [1.0]
|> str(x)`)
}
