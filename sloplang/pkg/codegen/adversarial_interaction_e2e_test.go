package codegen

import "testing"

// --- Runtime — Panics (2 tests) ---

func TestAdv_Inter_UnpackSingleElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `a, b = [42]`, "unpack")
}

func TestAdv_Inter_UnpackEmpty(t *testing.T) {
	runE2EExpectPanicContaining(t, `a, b = []`, "unpack")
}

// --- Runtime — Success (12 tests) ---

func TestAdv_Inter_MapPassedToFn(t *testing.T) {
	got := runE2E(t, "m{x} = [[5]]\nfn f(m) { m@x = [10]\n<- [] }\nf(m)\n|> str(m@x)")
	if got != "[10]" {
		t.Fatalf("expected [10], got %q", got)
	}
}

func TestAdv_Inter_FnReturnsMap(t *testing.T) {
	got := runE2E(t, "fn mk() { r{a} = [[1]]\n<- r }\nm = mk()\n|> str(m@a)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Inter_LoopBuildsArrayThenLength(t *testing.T) {
	got := runE2E(t, "arr = []\nfor x in [1, 2, 3, 4, 5] { arr << x }\n|> str(#arr)")
	if got != "[5]" {
		t.Fatalf("expected [5], got %q", got)
	}
}

func TestAdv_Inter_NestedFnCalls(t *testing.T) {
	got := runE2E(t, "fn double(x) { <- x * [2] }\nfn quad(x) { <- double(double(x)) }\n|> str(quad([3]))")
	if got != "[12]" {
		t.Fatalf("expected [12], got %q", got)
	}
}

func TestAdv_Inter_GlobalMutatedByTwoFns(t *testing.T) {
	got := runE2E(t, "g = [0]\nfn inc() { g = g + [1]\n<- [] }\nfn dec() { g = g - [1]\n<- [] }\ninc()\ninc()\ndec()\n|> str(g)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Inter_MapArrayValues(t *testing.T) {
	got := runE2E(t, "m{a, b} = [[1, 2], [3, 4]]\nfor v in @@m { |> str(v) }")
	if got != "[1, 2][3, 4]" {
		t.Fatalf("expected [1, 2][3, 4], got %q", got)
	}
}

func TestAdv_Inter_ConcatThenContains(t *testing.T) {
	got := runE2E(t, "x = [1, 2] ++ [3, 4]\n|> str(x ?? [3])")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Inter_SliceThenEquality(t *testing.T) {
	got := runE2E(t, "arr = [1, 2, 3, 4, 5]\nslice = arr::1::4\n|> str(slice == [2, 3, 4])")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Inter_PopPushInLoop(t *testing.T) {
	got := runE2E(t, "arr = [1, 2, 3]\nresult = []\nfor x in [1, 2, 3] { v = >>arr\nresult << v }\n|> str(result)")
	if got != "[3, 2, 1]" {
		t.Fatalf("expected [3, 2, 1], got %q", got)
	}
}

func TestAdv_Inter_IterateMapKeys(t *testing.T) {
	// ++ is array concat — collect keys into array, then print
	got := runE2E(t, "m{name, age} = [\"bob\", [30]]\nfor k in ##m { |> k }")
	if got != "nameage" {
		t.Fatalf("expected nameage, got %q", got)
	}
}

func TestAdv_Inter_ComparisonInFnUsedInIf(t *testing.T) {
	got := runE2E(t, "fn gt(a, b) { <- a > b }\nif gt([5], [3]) { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected yes, got %q", got)
	}
}

func TestAdv_Inter_RecursiveWithAccumulator(t *testing.T) {
	src := `fn sum(arr, i, acc) {
if i == #arr { <- acc }
acc = acc + arr$i
<- sum(arr, i + [1], acc)
}
|> str(sum([1, 2, 3, 4], [0], [0]))`
	got := runE2E(t, src)
	if got != "[10]" {
		t.Fatalf("expected [10], got %q", got)
	}
}
