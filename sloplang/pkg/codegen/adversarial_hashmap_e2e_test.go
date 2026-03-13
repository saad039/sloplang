package codegen

import "testing"

// --- Parser Rejection (9 tests) ---

func TestAdv_Map_DuplicateKeys(t *testing.T) {
	runE2EExpectParseError(t, `m{a, b, a} = [1, 2, 3]`, "duplicate key")
}

func TestAdv_Map_DuplicateKeys_Adjacent(t *testing.T) {
	runE2EExpectParseError(t, `m{x, x} = [1, 2]`, "duplicate key")
}

func TestAdv_Map_NumericKey(t *testing.T) {
	runE2EExpectParseError(t, `m{1, 2} = [1, 2]`, "expected key name")
}

func TestAdv_Map_KeywordAsKey(t *testing.T) {
	runE2EExpectParseError(t, `m{if, else} = [1, 2]`, "")
}

func TestAdv_Map_EmptyKeyName(t *testing.T) {
	runE2EExpectParseError(t, `m{, b} = [1, 2]`, "")
}

func TestAdv_Map_TrailingComma(t *testing.T) {
	runE2EExpectParseError(t, `m{a, b,} = [1, 2]`, "")
}

func TestAdv_Map_MissingEquals(t *testing.T) {
	runE2EExpectParseError(t, `m{a, b} [1, 2]`, "expected '='")
}

func TestAdv_Map_MissingValue(t *testing.T) {
	runE2EExpectParseError(t, `m{a, b} =`, "")
}

func TestAdv_Map_UnclosedBrace(t *testing.T) {
	runE2EExpectParseError(t, `m{a, b = [1, 2]`, "expected '}'")
}

// --- Runtime (4 tests) ---

func TestAdv_Map_KeyValueMismatch_MoreKeys(t *testing.T) {
	runE2EExpectPanicContaining(t, `m{a, b, c} = [1, 2]`, "key count")
}

func TestAdv_Map_KeyValueMismatch_MoreValues(t *testing.T) {
	runE2EExpectPanicContaining(t, `m{a} = [1, 2, 3]`, "key count")
}

func TestAdv_Map_DynAccess_IntOnMap(t *testing.T) {
	// Dynamic access with int key on map — dispatches int→index, but map has no numeric keys
	// Verify: may succeed (returning nil/zero) or panic depending on DynAccess impl
	got := runE2E(t, "m{a} = [[1]]\ni = [0]\n|> str(m$i)")
	_ = got // behavior depends on DynAccess implementation for int on map
}

func TestAdv_Map_ReDeclareSameVar(t *testing.T) {
	// Re-declaring same variable as hashmap — Go allows var redeclaration
	// Verify: second declaration overwrites first
	got := runE2E(t, "m{a} = [[1]]\nm{b} = [[2]]\n|> str(m@b)")
	if got != "[2]" {
		t.Fatalf("expected [2], got %q", got)
	}
}
