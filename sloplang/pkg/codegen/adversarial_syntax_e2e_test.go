package codegen

import "testing"

// --- Parse Errors (18 tests) ---

func TestAdv_Syn_KeywordAsVar_If(t *testing.T) {
	runE2EExpectParseError(t, `if = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_Fn(t *testing.T) {
	runE2EExpectParseError(t, `fn = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_For(t *testing.T) {
	runE2EExpectParseError(t, `for = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_True(t *testing.T) {
	runE2EExpectParseError(t, `true = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_False(t *testing.T) {
	runE2EExpectParseError(t, `false = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_Null(t *testing.T) {
	runE2EExpectParseError(t, `null = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_Break(t *testing.T) {
	runE2EExpectParseError(t, `break = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_In(t *testing.T) {
	runE2EExpectParseError(t, `in = [1]`, "")
}

func TestAdv_Syn_KeywordAsVar_Else(t *testing.T) {
	runE2EExpectParseError(t, `else = [1]`, "")
}

func TestAdv_Syn_TrailingCommaInArray(t *testing.T) {
	runE2EExpectParseError(t, `x = [1, 2,]`, "")
}

func TestAdv_Syn_LeadingCommaInArray(t *testing.T) {
	runE2EExpectParseError(t, `x = [, 1, 2]`, "")
}

func TestAdv_Syn_DoubleCommaInArray(t *testing.T) {
	runE2EExpectParseError(t, `x = [1,, 2]`, "")
}

func TestAdv_Syn_UnclosedArray(t *testing.T) {
	runE2EExpectParseError(t, `x = [1, 2`, "")
}

func TestAdv_Syn_UnclosedString(t *testing.T) {
	runE2EExpectParseError(t, `x = "hello`, "")
}

func TestAdv_Syn_FnKeywordAsParam(t *testing.T) {
	runE2EExpectParseError(t, `fn f(if) { <- [1] }`, "expected parameter name")
}

func TestAdv_Syn_AssignToLiteral(t *testing.T) {
	runE2EExpectParseError(t, `[1] = [2]`, "")
}

func TestAdv_Syn_BareNumberAssign(t *testing.T) {
	runE2EExpectParseError(t, `x = 42`, "bare number")
}

func TestAdv_Syn_BareNullAssign(t *testing.T) {
	runE2EExpectParseError(t, `x = null`, "")
}

// --- Compile Errors (5 tests) ---

func TestAdv_Syn_BreakOutsideLoop(t *testing.T) {
	runE2EExpectCompileError(t, `break`)
}

func TestAdv_Syn_ReturnOutsideFunction(t *testing.T) {
	runE2EExpectCompileError(t, `<- [1]`)
}

func TestAdv_Syn_EmptyFnBody(t *testing.T) {
	runE2EExpectCompileError(t, `fn f() { }`)
}

func TestAdv_Syn_FnDuplicateParams(t *testing.T) {
	runE2EExpectCompileError(t, `fn f(x, x) { <- x }`)
}

func TestAdv_Syn_NestedFnDecl(t *testing.T) {
	runE2EExpectCompileError(t, "fn f() { fn g() { <- [1] }\n<- g() }")
}

// --- Valid Programs (8 tests) ---

func TestAdv_Syn_EmptyProgram(t *testing.T) {
	got := runE2E(t, ``)
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestAdv_Syn_OnlyComment(t *testing.T) {
	got := runE2E(t, `// just a comment`)
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestAdv_Syn_OnlyWhitespace(t *testing.T) {
	got := runE2E(t, "   \n\n")
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestAdv_Syn_EmptyIfBody(t *testing.T) {
	got := runE2E(t, `if true { }`)
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestAdv_Syn_EmptyForBody(t *testing.T) {
	got := runE2E(t, `for x in [1, 2] { }`)
	if got != "" {
		t.Fatalf("expected empty output, got %q", got)
	}
}

func TestAdv_Syn_TwoStatementsOneLine(t *testing.T) {
	got := runE2E(t, "x = [1]\n|> str(x)")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Syn_DeeplyNestedArrays(t *testing.T) {
	got := runE2E(t, "x = [[[[[[[[[[[1]]]]]]]]]]]\n|> str(x)")
	if got != "[[[[[[[[[[[1]]]]]]]]]]]" {
		t.Fatalf("expected deeply nested, got %q", got)
	}
}

func TestAdv_Syn_CommentInExpression(t *testing.T) {
	got := runE2E(t, "x = [1] +\n// comment\n[2]\n|> str(x)")
	if got != "[3]" {
		t.Fatalf("expected [3], got %q", got)
	}
}
