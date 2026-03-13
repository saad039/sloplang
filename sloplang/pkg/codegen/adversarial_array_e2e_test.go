package codegen

import "testing"

func TestAdv_Arr_IndexNegative(t *testing.T) {
	runE2EExpectParseError(t, "x = [1,2,3]\n|> x@-1", "")
}

func TestAdv_Arr_IndexOnEmpty(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = []\n|> x@0", "index out of bounds")
}

func TestAdv_Arr_SliceNegativeStart(t *testing.T) {
	runE2EExpectParseError(t, "x = [1,2,3]\ny = x::-1::2", "")
}

func TestAdv_Arr_SliceEndBeyondLength(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = [1,2]\ny = x::0::10", "slice bounds")
}

func TestAdv_Arr_RemoveFromEmpty(t *testing.T) {
	got := runE2E(t, "x = []\ny = x -- [1]\n|> str(y)")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestAdv_Arr_DynAccessNullKey(t *testing.T) {
	runE2EExpectPanicContaining(t, "x = [1,2]\ni = [null]\n|> x$i", "SlopNull")
}

func TestAdv_Arr_SelfPush(t *testing.T) {
	got := runE2E(t, "x = [1]\nx << x\n|> str(x)")
	if got != "[1, 1]" {
		t.Fatalf("expected [1, 1], got %q", got)
	}
}

func TestAdv_Arr_NestedIndexChain(t *testing.T) {
	got := runE2E(t, "x = [[1,2],[3,4]]\ny = x@0\n|> str(y@1)")
	if got != "[2]" {
		t.Fatalf("expected [2], got %q", got)
	}
}
