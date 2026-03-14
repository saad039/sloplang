package codegen

import "testing"

// to_chars tests
func TestP10_ToChars_ASCII(t *testing.T) {
	got := runE2E(t, `chars = to_chars("hello")
|> str(chars)`)
	if got != "[h, e, l, l, o]" {
		t.Fatalf("expected [h, e, l, l, o], got %q", got)
	}
}

func TestP10_ToChars_SingleChar(t *testing.T) {
	got := runE2E(t, `|> str(to_chars("x"))`)
	if got != "x" {
		t.Fatalf("expected x, got %q", got)
	}
}

func TestP10_ToChars_EmptyString(t *testing.T) {
	got := runE2E(t, `|> str(to_chars(""))`)
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestP10_ToChars_Unicode(t *testing.T) {
	got := runE2E(t, `|> str(to_chars("héllo"))`)
	if got != "[h, é, l, l, o]" {
		t.Fatalf("expected [h, é, l, l, o], got %q", got)
	}
}

func TestP10_ToChars_FromVariable(t *testing.T) {
	got := runE2E(t, `s = "abc"
|> str(to_chars(s))`)
	if got != "[a, b, c]" {
		t.Fatalf("expected [a, b, c], got %q", got)
	}
}

func TestP10_ToChars_PanicOnInt(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_chars([42])`, "to_chars requires a string argument")
}

func TestP10_ToChars_PanicOnNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_chars([null])`, "to_chars requires a string argument")
}

func TestP10_ToChars_PanicOnMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_chars("a" ++ "b")`, "to_chars requires a string argument")
}

// to_float tests
func TestP10_ToFloat_FromInt(t *testing.T) {
	got := runE2E(t, `|> str(to_float([42]))`)
	if got != "[42]" {
		t.Fatalf("expected [42], got %q", got)
	}
}

func TestP10_ToFloat_FromFloat(t *testing.T) {
	got := runE2E(t, `|> str(to_float([3.14]))`)
	if got != "[3.14]" {
		t.Fatalf("expected [3.14], got %q", got)
	}
}

func TestP10_ToFloat_FromString(t *testing.T) {
	got := runE2E(t, `|> str(to_float("2.5"))`)
	if got != "[2.5]" {
		t.Fatalf("expected [2.5], got %q", got)
	}
}

func TestP10_ToFloat_FromZero(t *testing.T) {
	got := runE2E(t, `|> str(to_float([0]))`)
	if got != "[0]" {
		t.Fatalf("expected [0], got %q", got)
	}
}

func TestP10_ToFloat_ArithmeticAfterConversion(t *testing.T) {
	got := runE2E(t, `a = to_float([3])
b = [1.5]
|> str(a + b)`)
	if got != "[4.5]" {
		t.Fatalf("expected [4.5], got %q", got)
	}
}

func TestP10_ToFloat_PanicOnInvalidString(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_float("abc")`, "to_float: cannot convert")
}

func TestP10_ToFloat_PanicOnNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_float([null])`, "to_float: cannot convert")
}

func TestP10_ToFloat_PanicOnMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_float([1, 2])`, "to_float: requires single-element")
}

// to_int tests
func TestP10_ToInt_FromFloat(t *testing.T) {
	got := runE2E(t, `|> str(to_int([3.14]))`)
	if got != "[3]" {
		t.Fatalf("expected [3], got %q", got)
	}
}

func TestP10_ToInt_FromNegativeFloat(t *testing.T) {
	got := runE2E(t, `|> str(to_int([-2.9]))`)
	if got != "[-2]" {
		t.Fatalf("expected [-2], got %q", got)
	}
}

func TestP10_ToInt_FromInt(t *testing.T) {
	got := runE2E(t, `|> str(to_int([42]))`)
	if got != "[42]" {
		t.Fatalf("expected [42], got %q", got)
	}
}

func TestP10_ToInt_FromString(t *testing.T) {
	got := runE2E(t, `|> str(to_int("5"))`)
	if got != "[5]" {
		t.Fatalf("expected [5], got %q", got)
	}
}

func TestP10_ToInt_FromZeroFloat(t *testing.T) {
	got := runE2E(t, `|> str(to_int([0.0]))`)
	if got != "[0]" {
		t.Fatalf("expected [0], got %q", got)
	}
}

func TestP10_ToInt_PanicOnInvalidString(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_int("abc")`, "to_int: cannot convert")
}

func TestP10_ToInt_PanicOnNull(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_int([null])`, "to_int: cannot convert")
}

func TestP10_ToInt_PanicOnMultiElement(t *testing.T) {
	runE2EExpectPanicContaining(t, `to_int([1, 2])`, "to_int: requires single-element")
}

// fmt_float tests
func TestP10_FmtFloat_TwoDecimals(t *testing.T) {
	got := runE2E(t, `|> fmt_float([3.14159], [2])
|> "\n"`)
	if got != "3.14" {
		t.Fatalf("expected 3.14, got %q", got)
	}
}

func TestP10_FmtFloat_ThreeDecimals(t *testing.T) {
	got := runE2E(t, `|> fmt_float([42], [3])
|> "\n"`)
	if got != "42.000" {
		t.Fatalf("expected 42.000, got %q", got)
	}
}

func TestP10_FmtFloat_ZeroDecimals(t *testing.T) {
	got := runE2E(t, `|> fmt_float([1.0], [0])
|> "\n"`)
	if got != "1" {
		t.Fatalf("expected 1, got %q", got)
	}
}

func TestP10_FmtFloat_FiveDecimals(t *testing.T) {
	got := runE2E(t, `|> fmt_float([0.1], [5])
|> "\n"`)
	if got != "0.10000" {
		t.Fatalf("expected 0.10000, got %q", got)
	}
}

func TestP10_FmtFloat_IntPromoted(t *testing.T) {
	got := runE2E(t, `|> fmt_float([100], [2])
|> "\n"`)
	if got != "100.00" {
		t.Fatalf("expected 100.00, got %q", got)
	}
}

func TestP10_FmtFloat_Negative(t *testing.T) {
	got := runE2E(t, `|> fmt_float([-3.14], [1])
|> "\n"`)
	if got != "-3.1" {
		t.Fatalf("expected -3.1, got %q", got)
	}
}

func TestP10_FmtFloat_PanicOnString(t *testing.T) {
	runE2EExpectPanicContaining(t, `fmt_float("abc", [2])`, "fmt_float: first argument must be numeric")
}

func TestP10_FmtFloat_PanicOnNegativeDecimals(t *testing.T) {
	runE2EExpectPanicContaining(t, `fmt_float([3.14], [-1])`, "fmt_float: second argument must be non-negative integer")
}

// <<< (nested push) tests
func TestP10_NestPush_MultiElement(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr <<< [3, 4]
|> str(arr)`)
	if got != "[1, 2, [3, 4]]" {
		t.Fatalf("expected [1, 2, [3, 4]], got %q", got)
	}
}

func TestP10_NestPush_SingleElement(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr <<< [42]
|> str(arr)`)
	if got != "[1, 2, [42]]" {
		t.Fatalf("expected [1, 2, [42]], got %q", got)
	}
}

func TestP10_NestPush_String(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr <<< "hello"
|> str(arr)`)
	if got != "[1, 2, hello]" {
		t.Fatalf("expected [1, 2, hello], got %q", got)
	}
}

func TestP10_NestPush_StringArray(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr <<< ["saad", "is"]
|> str(arr)`)
	if got != "[1, 2, [saad, is]]" {
		t.Fatalf("expected [1, 2, [saad, is]], got %q", got)
	}
}

func TestP10_NestPush_EmptyIntoEmpty(t *testing.T) {
	got := runE2E(t, `arr = []
arr <<< [1]
|> str(arr)`)
	if got != "[[1]]" {
		t.Fatalf("expected [[1]], got %q", got)
	}
}

func TestP10_NestPush_Multiple(t *testing.T) {
	got := runE2E(t, `arr = [1, 2]
arr <<< [3, 4]
arr <<< [42]
arr <<< "hello"
|> str(arr)`)
	if got != "[1, 2, [3, 4], [42], hello]" {
		t.Fatalf("expected [1, 2, [3, 4], [42], hello], got %q", got)
	}
}

func TestP10_NestPush_SpreadVsNest(t *testing.T) {
	got := runE2E(t, `a = [1]
a << [2, 3]
|> str(a)
|> "\n"
b = [1]
b <<< [2, 3]
|> str(b)`)
	if got != "[1, 2, 3]\n[1, [2, 3]]" {
		t.Fatalf("expected [1, 2, 3]\\n[1, [2, 3]], got %q", got)
	}
}

func TestP10_NestPush_EmptyArray(t *testing.T) {
	got := runE2E(t, `arr = [1]
arr <<< []
|> str(arr)`)
	if got != "[1, []]" {
		t.Fatalf("expected [1, []], got %q", got)
	}
}
