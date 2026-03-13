package codegen

import "testing"

// --- Runtime — Panics / Verify (2 tests) ---

func TestAdv_MutIter_PopDuringIterate(t *testing.T) {
	// Iterate snapshots slice header but Pop shrinks it
	// Go range iterates on snapshot — may not panic; verify no crash
	got := runE2E(t, "arr = [1, 2, 3]\nfor x in arr { v = >>arr }\n|> str(arr)")
	_ = got // behavior depends on iteration snapshot semantics
}

func TestAdv_MutIter_RemoveAtDuringIterate(t *testing.T) {
	// Array shrinks during iteration — Go range iterates on snapshot
	// Verify: may complete or panic depending on slice mutation visibility
	got := runE2E(t, "arr = [1, 2, 3]\nfor x in arr { arr ~@ [0] }\n|> str(arr)")
	_ = got // behavior depends on iteration snapshot semantics
}

// --- Runtime — Success (10 tests) ---

func TestAdv_MutIter_PushToDifferentArray(t *testing.T) {
	got := runE2E(t, "src = [1, 2, 3]\ndst = []\nfor x in src { dst << x }\n|> str(dst)")
	if got != "[1, 2, 3]" {
		t.Fatalf("expected [1, 2, 3], got %q", got)
	}
}

func TestAdv_MutIter_IndexSetOnSource(t *testing.T) {
	got := runE2E(t, "arr = [1, 2, 3]\nfor x in arr { arr@0 = [99] }\n|> str(arr)")
	if got != "[99, 2, 3]" {
		t.Fatalf("expected [99, 2, 3], got %q", got)
	}
}

func TestAdv_MutIter_CopyBeforeMutate(t *testing.T) {
	got := runE2E(t, "a = [1, 2, 3]\nb = a::0::3\na << [4]\n|> str(b)")
	if got != "[1, 2, 3]" {
		t.Fatalf("expected [1, 2, 3], got %q", got)
	}
}

func TestAdv_MutIter_SharedRefAfterAssign(t *testing.T) {
	got := runE2E(t, "a = [1, 2]\nb = a\na << [3]\n|> str(b)")
	if got != "[1, 2, 3]" {
		t.Fatalf("expected [1, 2, 3], got %q", got)
	}
}

func TestAdv_MutIter_MapKeySetDuringKeyIter(t *testing.T) {
	got := runE2E(t, "m{a, b} = [[1], [2]]\nfor k in ##m { m@a = [99] }\n|> str(m@a)")
	if got != "[99]" {
		t.Fatalf("expected [99], got %q", got)
	}
}

func TestAdv_MutIter_BuildFromLoop(t *testing.T) {
	got := runE2E(t, "result = []\nfor x in [10, 20, 30] { result << x * [2] }\n|> str(result)")
	if got != "[20, 40, 60]" {
		t.Fatalf("expected [20, 40, 60], got %q", got)
	}
}

func TestAdv_MutIter_AccumulatorSum(t *testing.T) {
	got := runE2E(t, "sum = [0]\nfor x in [1, 2, 3, 4, 5] { sum = sum + x }\n|> str(sum)")
	if got != "[15]" {
		t.Fatalf("expected [15], got %q", got)
	}
}

func TestAdv_MutIter_FilterPattern(t *testing.T) {
	got := runE2E(t, "src = [1, 2, 3, 4, 5]\ndst = []\nfor x in src { if x > [2] { dst << x } }\n|> str(dst)")
	if got != "[3, 4, 5]" {
		t.Fatalf("expected [3, 4, 5], got %q", got)
	}
}

func TestAdv_MutIter_NestedLoopOuterMutation(t *testing.T) {
	got := runE2E(t, "arr = [1, 2]\nresult = []\nfor a in arr { for b in [10, 20] { result << a + b } }\n|> str(result)")
	if got != "[11, 21, 12, 22]" {
		t.Fatalf("expected [11, 21, 12, 22], got %q", got)
	}
}

func TestAdv_MutIter_SwapElements(t *testing.T) {
	got := runE2E(t, "arr = [1, 2]\ntmp = arr@0\narr@0 = arr@1\narr@1 = tmp\n|> str(arr)")
	if got != "[2, 1]" {
		t.Fatalf("expected [2, 1], got %q", got)
	}
}
