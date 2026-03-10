package codegen

import (
	"testing"
)

// ============================================================
// Semantic Mutation E2E Tests
// Covers IndexSet, DynAccessSet, KeySetStr, Push, Pop, RemoveAt
// ============================================================

// --- IndexSet (arr@i = val) ---

func TestSem_Mut_IndexSet_SingleInt(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@1 = [999]
|> str(arr)`)
	if got != "[100, 999, 300]" {
		t.Fatalf("expected %q, got %q", "[100, 999, 300]", got)
	}
}

func TestSem_Mut_IndexSet_MultiElement(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@1 = [1, 2]
|> str(arr)`)
	if got != "[100, [1, 2], 300]" {
		t.Fatalf("expected %q, got %q", "[100, [1, 2], 300]", got)
	}
}

func TestSem_Mut_IndexSet_String(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@0 = "hello"
|> str(arr)`)
	if got != "[hello, 200, 300]" {
		t.Fatalf("expected %q, got %q", "[hello, 200, 300]", got)
	}
}

func TestSem_Mut_IndexSet_Null(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@0 = [null]
|> str(arr)`)
	if got != "[null, 200, 300]" {
		t.Fatalf("expected %q, got %q", "[null, 200, 300]", got)
	}
}

func TestSem_Mut_IndexSet_EmptyArray(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@0 = []
|> str(arr)`)
	if got != "[[], 200, 300]" {
		t.Fatalf("expected %q, got %q", "[[], 200, 300]", got)
	}
}

func TestSem_Mut_IndexSet_Float(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@0 = [3.14]
|> str(arr)`)
	if got != "[3.14, 200, 300]" {
		t.Fatalf("expected %q, got %q", "[3.14, 200, 300]", got)
	}
}

func TestSem_Mut_IndexSet_FirstIndex(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
arr@0 = [99]
|> str(arr)`)
	if got != "[99, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[99, 20, 30]", got)
	}
}

func TestSem_Mut_IndexSet_LastIndex(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
arr@2 = [99]
|> str(arr)`)
	if got != "[10, 20, 99]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 99]", got)
	}
}

func TestSem_Mut_IndexSet_OutOfBounds(t *testing.T) {
	runE2EExpectPanic(t, `arr = [100, 200, 300]
arr@5 = [1]`)
}

func TestSem_Mut_IndexSet_NegativeIndex(t *testing.T) {
	// arr@-1 is a parse error, so use dynamic access with negative int
	runE2EExpectPanic(t, `arr = [100, 200, 300]
idx = [0] - [1]
arr$idx = [1]`)
}

func TestSem_Mut_IndexSet_ReadBack(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@1 = [5]
|> str(arr@1)`)
	if got != "[5]" {
		t.Fatalf("expected %q, got %q", "[5]", got)
	}
}

func TestSem_Mut_IndexSet_ReadBackMulti(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@1 = [1, 2]
|> str(arr@1)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Mut_IndexSet_LengthUnchanged(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@1 = [999]
|> str(#arr)`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

func TestSem_Mut_IndexSet_DoubleSet(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@0 = [5]
arr@0 = [10]
|> str(arr)`)
	if got != "[10, 200, 300]" {
		t.Fatalf("expected %q, got %q", "[10, 200, 300]", got)
	}
}

func TestSem_Mut_IndexSet_SetInLoop(t *testing.T) {
	got := runE2E(t, `arr = [0, 0, 0]
i = [0]
for x in [10, 20, 30] {
	arr$i = x
	i = i + [1]
}
|> str(arr)`)
	if got != "[10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30]", got)
	}
}

func TestSem_Mut_IndexSet_SetThenConcat(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
arr@1 = [999]
result = arr ++ [400, 500]
|> str(result)`)
	if got != "[100, 999, 300, 400, 500]" {
		t.Fatalf("expected %q, got %q", "[100, 999, 300, 400, 500]", got)
	}
}

// --- DynAccessSet (arr$var = val) ---

func TestSem_Mut_DynAccessSet_IntKey_Single(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
idx = [1]
arr$idx = [999]
|> str(arr)`)
	if got != "[100, 999, 300]" {
		t.Fatalf("expected %q, got %q", "[100, 999, 300]", got)
	}
}

func TestSem_Mut_DynAccessSet_IntKey_Multi(t *testing.T) {
	got := runE2E(t, `arr = [100, 200, 300]
idx = [1]
arr$idx = [1, 2]
|> str(arr)`)
	if got != "[100, [1, 2], 300]" {
		t.Fatalf("expected %q, got %q", "[100, [1, 2], 300]", got)
	}
}

func TestSem_Mut_DynAccessSet_StringKey_Single(t *testing.T) {
	got := runE2E(t, `m{name} = ["bob"]
key = "name"
m$key = [42]
|> str(m@name)`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestSem_Mut_DynAccessSet_StringKey_Multi(t *testing.T) {
	got := runE2E(t, `m2{name} = ["bob"]
key = "name"
m2$key = [1, 2]
|> str(m2@name)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Mut_DynAccessSet_FloatKey_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = [100, 200, 300]
key = [3.14]
arr$key = [1]`)
}

func TestSem_Mut_DynAccessSet_MultiElementKey_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = [100, 200, 300]
key = [1, 2]
arr$key = [1]`)
}

// --- KeySetStr (map@key = val) ---

func TestSem_Mut_KeySetStr_UpdateSingle(t *testing.T) {
	got := runE2E(t, `person{name, age} = ["bob", [30]]
person@age = [31]
|> str(person@age)`)
	if got != "[31]" {
		t.Fatalf("expected %q, got %q", "[31]", got)
	}
}

func TestSem_Mut_KeySetStr_UpdateMulti(t *testing.T) {
	got := runE2E(t, `person2{name, age} = ["bob", [30]]
person2@age = [1, 2]
|> str(person2@age)`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Mut_KeySetStr_AddNewKey(t *testing.T) {
	got := runE2E(t, `person3{name, age} = ["bob", [30]]
person3@email = "a@b"
|> str(person3@email)`)
	if got != "a@b" {
		t.Fatalf("expected %q, got %q", "a@b", got)
	}
}

func TestSem_Mut_KeySetStr_ValuesAfterSingleSet(t *testing.T) {
	got := runE2E(t, `person4{name, age} = ["bob", [30]]
person4@age = [31]
|> str(@@person4)`)
	if got != "[bob, 31]" {
		t.Fatalf("expected %q, got %q", "[bob, 31]", got)
	}
}

func TestSem_Mut_KeySetStr_ValuesAfterMultiSet(t *testing.T) {
	got := runE2E(t, `person5{name, scores} = ["bob", [30]]
person5@scores = [1, 2]
|> str(@@person5)`)
	if got != "[bob, [1, 2]]" {
		t.Fatalf("expected %q, got %q", "[bob, [1, 2]]", got)
	}
}

func TestSem_Mut_KeySetStr_KeysAfterAdd(t *testing.T) {
	got := runE2E(t, `person6{name, age} = ["bob", [30]]
person6@email = "a@b"
|> str(##person6)`)
	if got != "[name, age, email]" {
		t.Fatalf("expected %q, got %q", "[name, age, email]", got)
	}
}

func TestSem_Mut_KeySetStr_AddNewKeyInt(t *testing.T) {
	got := runE2E(t, `person7{name, age} = ["bob", [30]]
person7@score = [100]
|> str(person7@score)`)
	if got != "[100]" {
		t.Fatalf("expected %q, got %q", "[100]", got)
	}
}

// --- Push (<<) ---

func TestSem_Mut_Push_SingleElement(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << [4]
|> str(arr)`)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

func TestSem_Mut_Push_MultiElement(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << [4, 5]
|> str(arr)`)
	if got != "[1, 2, 3, 4, 5]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4, 5]", got)
	}
}

func TestSem_Mut_Push_EmptyPush(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << []
|> str(arr)`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestSem_Mut_Push_ToEmpty(t *testing.T) {
	got := runE2E(t, `e = []
e << [1]
|> str(e)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Mut_Push_Null(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << [null]
|> str(arr)`)
	if got != "[1, 2, 3, null]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, null]", got)
	}
}

func TestSem_Mut_Push_ThenLength(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << [4]
|> str(#arr)`)
	if got != "[4]" {
		t.Fatalf("expected %q, got %q", "[4]", got)
	}
}

func TestSem_Mut_Push_MultiplePushes(t *testing.T) {
	got := runE2E(t, `arr = [1]
arr << [2]
arr << [3]
arr << [4]
|> str(arr)`)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

func TestSem_Mut_Push_Nested(t *testing.T) {
	got := runE2E(t, `arr = [1, 2, 3]
arr << [[1, 2]]
|> str(arr)`)
	if got != "[1, 2, 3, [1, 2]]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, [1, 2]]", got)
	}
}

func TestSem_Mut_Push_String(t *testing.T) {
	got := runE2E(t, `arr = ["a"]
arr << "b"
|> str(arr)`)
	if got != "[a, b]" {
		t.Fatalf("expected %q, got %q", "[a, b]", got)
	}
}

// --- Pop (>>) ---

func TestSem_Mut_Pop_Basic(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
>>arr
|> str(arr)`)
	if got != "[10, 20]" {
		t.Fatalf("expected %q, got %q", "[10, 20]", got)
	}
}

func TestSem_Mut_Pop_SingleElement(t *testing.T) {
	got := runE2E(t, `arr = [42]
>>arr
|> str(arr)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Mut_Pop_Empty_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = []
x = >>arr`)
}

func TestSem_Mut_Pop_ReturnValue(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(>>arr)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestSem_Mut_Pop_AfterIndexSet(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
arr@2 = [99]
|> str(>>arr)`)
	if got != "[99]" {
		t.Fatalf("expected %q, got %q", "[99]", got)
	}
}

func TestSem_Mut_Pop_ThenPushRoundtrip(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
x = >>arr
arr << x
|> str(arr)`)
	if got != "[10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[10, 20, 30]", got)
	}
}

// --- RemoveAt (~@) ---

func TestSem_Mut_RemoveAt_Middle(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
removed = arr ~@ [1]
|> str(removed)
|> str(arr)`)
	if got != "[20][10, 30, 40]" {
		t.Fatalf("expected %q, got %q", "[20][10, 30, 40]", got)
	}
}

func TestSem_Mut_RemoveAt_First(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
removed = arr ~@ [0]
|> str(removed)
|> str(arr)`)
	if got != "[10][20, 30, 40]" {
		t.Fatalf("expected %q, got %q", "[10][20, 30, 40]", got)
	}
}

func TestSem_Mut_RemoveAt_Last(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
removed = arr ~@ [3]
|> str(removed)
|> str(arr)`)
	if got != "[40][10, 20, 30]" {
		t.Fatalf("expected %q, got %q", "[40][10, 20, 30]", got)
	}
}

func TestSem_Mut_RemoveAt_SingleElement(t *testing.T) {
	got := runE2E(t, `arr = [42]
removed = arr ~@ [0]
|> str(removed)
|> str(arr)`)
	if got != "[42][]" {
		t.Fatalf("expected %q, got %q", "[42][]", got)
	}
}

func TestSem_Mut_RemoveAt_OutOfBounds_Panics(t *testing.T) {
	runE2EExpectPanic(t, `arr = [10, 20, 30, 40]
removed = arr ~@ [10]`)
}

func TestSem_Mut_RemoveAt_ReturnValue(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30, 40]
|> str(arr ~@ [1])`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}
