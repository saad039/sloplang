package codegen

import "testing"

// ============================================================
// Semantic Array Operations E2E Tests
// Covers Index access, Slice, Concat, Remove, Contains, Unique, Length
// ============================================================

// --- Index access (~7 tests) ---

func TestSem_Arr_Index_First(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr@0)`)
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestSem_Arr_Index_Last(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr@2)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestSem_Arr_Index_OutOfBounds_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30]
x = arr@3`)
}

func TestSem_Arr_Index_Dynamic(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
i = [1]
|> str(arr$i)`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}

func TestSem_Arr_Index_Nested(t *testing.T) {
	got := runE2E(t, `outer = [[1, 2], [3, 4]]
|> str(outer@0)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Arr_Index_AfterPush(t *testing.T) {
	got := runE2E(t, `arr = [10, 20]
arr << [30]
|> str(arr@2)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestSem_Arr_Index_StringElement(t *testing.T) {
	got := runE2E(t, `arr = ["a", "b"]
|> arr@0`)
	if got != "a" {
		t.Fatalf("expected %q, got %q", "a", got)
	}
}

// --- Slice (~9 tests) ---

func TestSem_Arr_Slice_Basic(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40, 50]
|> str(arr::0::3)`)
	if got != "[10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30]", got)
	}
}

func TestSem_Arr_Slice_EmptyLoEqHi(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40, 50]
|> str(arr::2::2)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arr_Slice_ZeroZero(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40, 50]
|> str(arr::0::0)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arr_Slice_Full(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40, 50]
|> str(arr::0::5)`)
	if got != "[10, 20, 30, 40, 50]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30, 40, 50]", got)
	}
}

func TestSem_Arr_Slice_LastElement(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40, 50]
|> str(arr::4::5)`)
	if got != "[50]" {
		t.Fatalf("expected %q, got %q", "[50]", got)
	}
}

func TestSem_Arr_Slice_DoesNotMutate(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40, 50]
s = arr::0::3
|> str(arr)`)
	if got != "[10, 20, 30, 40, 50]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30, 40, 50]", got)
	}
}

func TestSem_Arr_Slice_HighOutOfBounds_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30, 40, 50]
x = arr::0::10`)
}

func TestSem_Arr_Slice_HiLessThanLo_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30, 40, 50]
x = arr::3::1`)
}

// --- Concat (++) (~8 tests) ---

func TestSem_Arr_Concat_Basic(t *testing.T) {
	got := runE2E(t, `|> str([1, 2] ++ [3, 4])`)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

func TestSem_Arr_Concat_LeftEmpty(t *testing.T) {
	got := runE2E(t, `|> str([] ++ [1])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Concat_RightEmpty(t *testing.T) {
	got := runE2E(t, `|> str([1] ++ [])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Concat_BothEmpty(t *testing.T) {
	got := runE2E(t, `|> str([] ++ [])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arr_Concat_DoesNotMutate(t *testing.T) {
	got := runE2E(t, `a = [1, 2]
b = [3, 4]
c = a ++ b
|> str(a)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Arr_Concat_Strings(t *testing.T) {
	got := runE2E(t, `|> str(["a"] ++ ["b"])`)
	if got != "[a, b]" {
		t.Fatalf("expected %q, got %q", "[a, b]", got)
	}
}

func TestSem_Arr_Concat_Mixed(t *testing.T) {
	got := runE2E(t, `|> str([1, "a"] ++ [null, 2])`)
	if got != "[1, a, null, 2]" {
		t.Fatalf("expected %q, got %q", "[1, a, null, 2]", got)
	}
}

func TestSem_Arr_Concat_Nested(t *testing.T) {
	got := runE2E(t, `|> str([[1]] ++ [[2]])`)
	if got != "[[1], [2]]" {
		t.Fatalf("expected %q, got %q", "[[1], [2]]", got)
	}
}

// --- Remove (--) (~5 tests) ---

func TestSem_Arr_Remove_FirstOccurrence(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3, 2, 4] -- [2])`)
	if got != "[1, 3, 2, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 3, 2, 4]", got)
	}
}

func TestSem_Arr_Remove_NotFound(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] -- [9])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestSem_Arr_Remove_EmptyVal(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] -- [])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestSem_Arr_Remove_DoesNotMutate(t *testing.T) {
	got := runE2E(t, `a = [1, 2, 3]
b = a -- [2]
|> str(a)`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestSem_Arr_Remove_Null(t *testing.T) {
	got := runE2E(t, `|> str([1, null, 3] -- [null])`)
	if got != "[1, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 3]", got)
	}
}

// --- Contains (??) (~6 tests) ---

func TestSem_Arr_Contains_Found(t *testing.T) {
	got := runE2E(t, `|> str([10, 20, 30] ?? [20])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Contains_NotFound(t *testing.T) {
	got := runE2E(t, `|> str([10, 20, 30] ?? [99])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arr_Contains_EmptyArray(t *testing.T) {
	got := runE2E(t, `|> str([] ?? [1])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arr_Contains_NullFound(t *testing.T) {
	got := runE2E(t, `arr = [null, 1]
|> str(arr ?? [null])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Contains_MultiOperand_Panics(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] ?? [1, 2]`)
}

func TestSem_Arr_Contains_InConditional(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3, 4, 5]
if arr ?? [5] {
	|> "yes"
}`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}

// --- Unique (~) (~8 tests) ---

func TestSem_Arr_Unique_Basic(t *testing.T) {
	got := runE2E(t, `|> str(~[1, 2, 2, 3, 1])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestSem_Arr_Unique_Empty(t *testing.T) {
	got := runE2E(t, `|> str(~[])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Arr_Unique_Single(t *testing.T) {
	got := runE2E(t, `|> str(~[1])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Unique_AllSame(t *testing.T) {
	got := runE2E(t, `|> str(~[1, 1, 1])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Unique_DoesNotMutate(t *testing.T) {
	got := runE2E(t, `a = [1, 2, 2]
b = ~a
|> str(a)`)
	if got != "[1, 2, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 2]", got)
	}
}

func TestSem_Arr_Unique_NullDedup(t *testing.T) {
	got := runE2E(t, `|> str(~[null, 1, null])`)
	if got != "[null, 1]" {
		t.Fatalf("expected %q, got %q", "[null, 1]", got)
	}
}

func TestSem_Arr_Unique_MixedTypes(t *testing.T) {
	got := runE2E(t, `|> str(~[1, "1", 1])`)
	if got != `[1, 1]` {
		t.Fatalf("expected %q, got %q", `[1, 1]`, got)
	}
}

func TestSem_Arr_Unique_Strings(t *testing.T) {
	got := runE2E(t, `|> str(~["a", "b", "a"])`)
	if got != "[a, b]" {
		t.Fatalf("expected %q, got %q", "[a, b]", got)
	}
}

// --- Length (#) (~6 tests) ---

func TestSem_Arr_Length_Basic(t *testing.T) {
	got := runE2E(t, `|> str(#[1, 2, 3])`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Arr_Length_Empty(t *testing.T) {
	got := runE2E(t, `|> str(#[])`)
	if got != "[0]" {
		t.Fatalf("expected %q, got %q", "[0]", got)
	}
}

func TestSem_Arr_Length_Null(t *testing.T) {
	got := runE2E(t, `|> str(#[null])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Arr_Length_Nested(t *testing.T) {
	got := runE2E(t, `|> str(#[1, [2, 3]])`)
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestSem_Arr_Length_AfterPush(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr << [3]
|> str(#arr)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Arr_Length_InExpression(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
if #arr == [3] {
	|> "yes"
}`)
	if got != "yes" {
		t.Fatalf("expected %q, got %q", "yes", got)
	}
}
