package codegen

import "testing"

func TestAdv_ErrRecov_SingleUnclosedBracket(t *testing.T) {
	// GIVEN: a program with one unclosed bracket followed by a valid statement
	// WHEN: parsed
	// THEN: produces parse errors
	runE2EExpectParseErrorCount(t, "x = [1, 2\ny = [3]", 10, "")
}

func TestAdv_ErrRecov_SingleUnclosedBrace(t *testing.T) {
	// GIVEN: a program with one unclosed brace followed by a valid statement
	// WHEN: parsed
	// THEN: produces parse errors (parser cascades without recovery)
	runE2EExpectParseErrorCount(t, "m{a, b = [1, 2]\ny = [3]", 10, "")
}

func TestAdv_ErrRecov_SingleMissingEquals(t *testing.T) {
	// GIVEN: a program missing = followed by a valid statement
	// WHEN: parsed
	// THEN: produces parse errors (parser cascades without recovery)
	runE2EExpectParseErrorCount(t, "x [1, 2]\ny = [3]", 10, "")
}

func TestAdv_ErrRecov_MultipleBadStatements(t *testing.T) {
	// GIVEN: two bare-literal errors followed by valid statement
	// WHEN: parsed
	// THEN: produces parse errors, includes "bare number"
	runE2EExpectParseErrorCount(t, "x = 42\ny = null\nz = [1]", 10, "bare number")
}

func TestAdv_ErrRecov_BadTokenMidProgram(t *testing.T) {
	// GIVEN: illegal tokens mid-program with valid statements around them
	// WHEN: parsed
	// THEN: produces parse errors (parser recovers to some degree)
	runE2EExpectParseErrorCount(t, "x = [1]\n? ? ?\ny = [2]", 10, "")
}

func TestAdv_ErrRecov_UnclosedStringThenValid(t *testing.T) {
	runE2EExpectParseError(t, "x = \"hello\ny = [1]", "")
}

func TestAdv_ErrRecov_KeywordAsVarThenValid(t *testing.T) {
	// GIVEN: keyword misuse followed by valid statements
	// WHEN: parsed
	// THEN: produces parse errors (parser cascades without recovery)
	runE2EExpectParseErrorCount(t, "if = [1]\nx = [2]\n|> str(x)", 10, "")
}

func TestAdv_ErrRecov_DoubleErrorSameStmt(t *testing.T) {
	// GIVEN: missing } and duplicate key in same statement
	// WHEN: parsed
	// THEN: both reported
	runE2EExpectParseErrorCount(t, "m{a, a = [1, 2]", 10, "")
}
