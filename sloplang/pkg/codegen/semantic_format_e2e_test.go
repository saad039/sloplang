package codegen

import "testing"

// ============================================================
// Semantic Format E2E Tests
// Covers str() output, |> no trailing newline, FormatValue after
// operations, and FormatValue with hashmaps.
// ============================================================

// --- str() output (~12 tests) ---

func TestSem_Fmt_Str_Int(t *testing.T) {
	got := runE2E(t, `|> str([42])`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestSem_Fmt_Str_Float(t *testing.T) {
	got := runE2E(t, `|> str([3.14])`)
	if got != "[3.14]" {
		t.Fatalf("expected %q, got %q", "[3.14]", got)
	}
}

func TestSem_Fmt_Str_Uint(t *testing.T) {
	got := runE2E(t, `|> str([42u])`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestSem_Fmt_Str_MultiInt(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3])`)
	if got != "[1, 2, 3]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3]", got)
	}
}

func TestSem_Fmt_Str_Empty(t *testing.T) {
	got := runE2E(t, `|> str([])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Fmt_Str_String(t *testing.T) {
	got := runE2E(t, `|> str("hello")`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestSem_Fmt_Str_Null(t *testing.T) {
	got := runE2E(t, `|> str([null])`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestSem_Fmt_Str_Mixed(t *testing.T) {
	got := runE2E(t, `|> str([1, null, "hi"])`)
	if got != "[1, null, hi]" {
		t.Fatalf("expected %q, got %q", "[1, null, hi]", got)
	}
}

func TestSem_Fmt_Str_NestedArrays(t *testing.T) {
	got := runE2E(t, `c = [[1, 2], [3]]
|> str(c)`)
	if got != "[[1, 2], [3]]" {
		t.Fatalf("expected %q, got %q", "[[1, 2], [3]]", got)
	}
}

func TestSem_Fmt_Str_MixedNesting(t *testing.T) {
	got := runE2E(t, `c = [1, [2, 3]]
|> str(c)`)
	if got != "[1, [2, 3]]" {
		t.Fatalf("expected %q, got %q", "[1, [2, 3]]", got)
	}
}

func TestSem_Fmt_Str_True(t *testing.T) {
	got := runE2E(t, `|> str(true)`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Fmt_Str_False(t *testing.T) {
	got := runE2E(t, `|> str(false)`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

// --- |> no trailing newline (~7 tests) ---

func TestSem_Fmt_Pipe_Concatenation(t *testing.T) {
	got := runE2E(t, `|> "hello"
|> "world"`)
	if got != "helloworld" {
		t.Fatalf("expected %q, got %q", "helloworld", got)
	}
}

func TestSem_Fmt_Pipe_ExplicitNewline(t *testing.T) {
	got := runE2E(t, `|> "hello"
|> "\n"
|> "world"`)
	if got != "hello\nworld" {
		t.Fatalf("expected %q, got %q", "hello\nworld", got)
	}
}

func TestSem_Fmt_Pipe_IntNoNewline(t *testing.T) {
	got := runE2E(t, `|> str([42])`)
	if got != "[42]" {
		t.Fatalf("expected %q, got %q", "[42]", got)
	}
}

func TestSem_Fmt_Pipe_NullOutput(t *testing.T) {
	got := runE2E(t, `|> str([null])`)
	if got != "[null]" {
		t.Fatalf("expected %q, got %q", "[null]", got)
	}
}

func TestSem_Fmt_Pipe_EmptyOutput(t *testing.T) {
	got := runE2E(t, `|> str([])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Fmt_Pipe_MultiOutput(t *testing.T) {
	got := runE2E(t, `|> str([1, 2])`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Fmt_Pipe_MultipleStmts(t *testing.T) {
	got := runE2E(t, `|> str([1])
|> str([2])
|> str([3])`)
	if got != "[1][2][3]" {
		t.Fatalf("expected %q, got %q", "[1][2][3]", got)
	}
}

// --- FormatValue after operations (~10 tests) ---

func TestSem_Fmt_Op_IndexAccess(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr@0)`)
	if got != "[10]" {
		t.Fatalf("expected %q, got %q", "[10]", got)
	}
}

func TestSem_Fmt_Op_IndexAccessString(t *testing.T) {
	got := runE2E(t, `arr = ["hello", "world"]
|> arr@0`)
	if got != "hello" {
		t.Fatalf("expected %q, got %q", "hello", got)
	}
}

func TestSem_Fmt_Op_PopReturn(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(>>arr)`)
	if got != "[30]" {
		t.Fatalf("expected %q, got %q", "[30]", got)
	}
}

func TestSem_Fmt_Op_RemoveAtReturn(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr ~@ [1])`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}

func TestSem_Fmt_Op_Slice(t *testing.T) {
	got := runE2E(t, `arr = [10, 20, 30]
|> str(arr::0::2)`)
	if got != "[10, 20]" {
		t.Fatalf("expected %q, got %q", "[10, 20]", got)
	}
}

func TestSem_Fmt_Op_Concat(t *testing.T) {
	got := runE2E(t, `|> str([1, 2] ++ [3, 4])`)
	if got != "[1, 2, 3, 4]" {
		t.Fatalf("expected %q, got %q", "[1, 2, 3, 4]", got)
	}
}

func TestSem_Fmt_Op_Unique(t *testing.T) {
	got := runE2E(t, `|> str(~[1, 1, 2])`)
	if got != "[1, 2]" {
		t.Fatalf("expected %q, got %q", "[1, 2]", got)
	}
}

func TestSem_Fmt_Op_ContainsTrue(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] ?? [2])`)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestSem_Fmt_Op_ContainsFalse(t *testing.T) {
	got := runE2E(t, `|> str([1, 2, 3] ?? [9])`)
	if got != "[]" {
		t.Fatalf("expected %q, got %q", "[]", got)
	}
}

func TestSem_Fmt_Op_Length(t *testing.T) {
	got := runE2E(t, `|> str(#[1, 2, 3])`)
	if got != "[3]" {
		t.Fatalf("expected %q, got %q", "[3]", got)
	}
}

// --- FormatValue with hashmaps (~3 tests) ---

func TestSem_Fmt_Map_KeysSingle(t *testing.T) {
	got := runE2E(t, `m{item} = [[5]]
|> str(##m)`)
	if got != "item" {
		t.Fatalf("expected %q, got %q", "item", got)
	}
}

func TestSem_Fmt_Map_ValuesMulti(t *testing.T) {
	got := runE2E(t, `m2{a, b} = [[1], [2]]
|> str(@@m2)`)
	// Each value is stored as *SlopValue, so MapValues returns nested SlopValues
	if got != "[[1], [2]]" {
		t.Fatalf("expected %q, got %q", "[[1], [2]]", got)
	}
}

func TestSem_Fmt_Map_ValuesAfterMutation(t *testing.T) {
	got := runE2E(t, `m3{x} = [[10]]
m3@x = [20]
|> str(@@m3)`)
	if got != "[20]" {
		t.Fatalf("expected %q, got %q", "[20]", got)
	}
}
