package codegen

import "testing"

// --- Parse Errors (2 tests) ---

func TestAdv_Str_UnclosedStringEOF(t *testing.T) {
	runE2EExpectParseError(t, `x = "hello`, "")
}

func TestAdv_Str_UnclosedStringBackslash(t *testing.T) {
	runE2EExpectParseError(t, `x = "hello\`, "")
}

// --- Runtime — Panics (1 test) ---

func TestAdv_Str_SplitNonString(t *testing.T) {
	runE2EExpectPanicContaining(t, `|> str(split([1, 2], " "))`, "single-element string")
}

// --- Runtime — Success (11 tests) ---

func TestAdv_Str_EscapeNewline(t *testing.T) {
	got := runE2E(t, `|> "hello\nworld"`)
	if got != "hello\nworld" {
		t.Fatalf("expected hello\\nworld, got %q", got)
	}
}

func TestAdv_Str_EscapeTab(t *testing.T) {
	got := runE2E(t, `|> "a\tb"`)
	if got != "a\tb" {
		t.Fatalf("expected a\\tb, got %q", got)
	}
}

func TestAdv_Str_EscapeBackslash(t *testing.T) {
	got := runE2E(t, `|> "a\\b"`)
	if got != `a\b` {
		t.Fatalf("expected a\\b, got %q", got)
	}
}

func TestAdv_Str_EscapeQuote(t *testing.T) {
	got := runE2E(t, `|> "say \"hi\""`)
	if got != `say "hi"` {
		t.Fatalf("expected say \"hi\", got %q", got)
	}
}

func TestAdv_Str_UnknownEscape(t *testing.T) {
	got := runE2E(t, `|> "a\zb"`)
	if got != `a\zb` {
		t.Fatalf("expected a\\zb, got %q", got)
	}
}

func TestAdv_Str_EmptyString(t *testing.T) {
	got := runE2E(t, `x = ""
|> x
|> "end"`)
	if got != "end" {
		t.Fatalf("expected end, got %q", got)
	}
}

func TestAdv_Str_StringWithBrackets(t *testing.T) {
	got := runE2E(t, "x = \"[1, 2]\"\n|> x")
	if got != "[1, 2]" {
		t.Fatalf("expected [1, 2], got %q", got)
	}
}

func TestAdv_Str_EqualityCaseSensitive(t *testing.T) {
	got := runE2E(t, "if \"Hello\" == \"hello\" { |> \"same\" } else { |> \"diff\" }")
	if got != "diff" {
		t.Fatalf("expected diff, got %q", got)
	}
}

func TestAdv_Str_OnlySpaces(t *testing.T) {
	got := runE2E(t, "x = \"   \"\n|> x")
	if got != "   " {
		t.Fatalf("expected 3 spaces, got %q", got)
	}
}

func TestAdv_Str_AllEscapesCombined(t *testing.T) {
	got := runE2E(t, `|> "a\nb\tc\\d\"e"`)
	if got != "a\nb\tc\\d\"e" {
		t.Fatalf("expected combined escapes, got %q", got)
	}
}

func TestAdv_Str_StringContainsInArray(t *testing.T) {
	got := runE2E(t, "arr = [\"hello\", \"world\"]\n|> str(arr ?? [\"hello\"])")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}
