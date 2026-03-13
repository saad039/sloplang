package codegen

import "testing"

func TestAdv_Empty_NegateEmpty(t *testing.T) {
	// GIVEN: an empty array
	// WHEN: negated with -
	// THEN: verify — may return [] or panic
	got := runE2E(t, "x = -[]\n|> str(x)")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestAdv_Empty_NotEmpty(t *testing.T) {
	// GIVEN: an empty array (falsy)
	// WHEN: ! applied
	// THEN: returns truthy and prints yes
	got := runE2E(t, "if ![] { |> \"yes\" }")
	if got != "yes" {
		t.Fatalf("expected yes, got %q", got)
	}
}

func TestAdv_Empty_LengthEmpty(t *testing.T) {
	// GIVEN: an empty array
	// WHEN: # applied
	// THEN: returns [0]
	got := runE2E(t, "|> str(#[])")
	if got != "[0]" {
		t.Fatalf("expected [0], got %q", got)
	}
}

func TestAdv_Empty_UniqueEmpty(t *testing.T) {
	// GIVEN: an empty array
	// WHEN: ~ applied
	// THEN: returns []
	got := runE2E(t, "|> str(~[])")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestAdv_Empty_ConcatEmptyLeft(t *testing.T) {
	// GIVEN: empty array on left
	// WHEN: ++ with non-empty
	// THEN: returns non-empty array
	got := runE2E(t, "|> str([] ++ [1, 2])")
	if got != "[1, 2]" {
		t.Fatalf("expected [1, 2], got %q", got)
	}
}

func TestAdv_Empty_ConcatEmptyRight(t *testing.T) {
	// GIVEN: empty array on right
	// WHEN: ++ with non-empty
	// THEN: returns non-empty array
	got := runE2E(t, "|> str([1, 2] ++ [])")
	if got != "[1, 2]" {
		t.Fatalf("expected [1, 2], got %q", got)
	}
}

func TestAdv_Empty_SliceEmptyZeroZero(t *testing.T) {
	// GIVEN: empty array
	// WHEN: sliced ::0::0
	// THEN: returns []
	got := runE2E(t, "x = []\ny = x::0::0\n|> str(y)")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestAdv_Empty_AddEmptyToEmpty(t *testing.T) {
	// GIVEN: two empty arrays
	// WHEN: added together
	// THEN: returns []
	got := runE2E(t, "|> str([] + [])")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestAdv_Empty_AddEmptyToNonEmpty(t *testing.T) {
	// GIVEN: empty array added to non-empty
	// WHEN: executed
	// THEN: runtime panics with length mismatch
	runE2EExpectPanicContaining(t, `x = [] + [1]`, "length")
}

func TestAdv_Empty_ForInEmpty(t *testing.T) {
	// GIVEN: for-in over empty array
	// WHEN: executed
	// THEN: loop body never executes
	got := runE2E(t, "for x in [] { |> \"no\" }\n|> \"done\"")
	if got != "done" {
		t.Fatalf("expected done, got %q", got)
	}
}

func TestAdv_Empty_IterateEmptyThenPush(t *testing.T) {
	// GIVEN: empty array
	// WHEN: element pushed
	// THEN: array has one element
	got := runE2E(t, "x = []\nx << [1]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}
