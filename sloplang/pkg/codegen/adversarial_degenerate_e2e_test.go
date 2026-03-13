package codegen

import (
	"fmt"
	"strings"
	"testing"
)

// --- Parse Errors (2 tests) ---

func TestAdv_Degen_SinglePipe(t *testing.T) {
	runE2EExpectParseError(t, `|`, "")
}

func TestAdv_Degen_SingleQuestion(t *testing.T) {
	runE2EExpectParseError(t, `?`, "")
}

// --- Runtime — Panics (1 test) ---

func TestAdv_Degen_LargeArrayInBoolean(t *testing.T) {
	runE2EExpectPanicContaining(t, `if [1, 2, 3] { |> "yes" }`, "boolean")
}

// --- Runtime — Success (7 tests) ---

func TestAdv_Degen_ManyAssignments(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 50; i++ {
		fmt.Fprintf(&b, "a%d = [%d]\n", i, i)
	}
	b.WriteString("|> str(a50)")
	got := runE2E(t, b.String())
	if got != "[50]" {
		t.Fatalf("expected [50], got %q", got)
	}
}

func TestAdv_Degen_NestedForIn3Levels(t *testing.T) {
	got := runE2E(t, "for a in [1, 2] { for b in [3, 4] { for c in [5, 6] { |> str(a * b * c) } } }")
	if got != "[15][18][20][24][30][36][40][48]" {
		t.Fatalf("expected [15][18][20][24][30][36][40][48], got %q", got)
	}
}

func TestAdv_Degen_ChainedConcat10(t *testing.T) {
	got := runE2E(t, "|> str([1] ++ [2] ++ [3] ++ [4] ++ [5] ++ [6] ++ [7] ++ [8] ++ [9] ++ [10])")
	if got != "[1, 2, 3, 4, 5, 6, 7, 8, 9, 10]" {
		t.Fatalf("expected [1, 2, 3, 4, 5, 6, 7, 8, 9, 10], got %q", got)
	}
}

func TestAdv_Degen_DeeplyNestedStr(t *testing.T) {
	got := runE2E(t, "x = [[[[[1]]]]]\n|> str(x)")
	if got != "[[[[[1]]]]]" {
		t.Fatalf("expected [[[[[1]]]]], got %q", got)
	}
}

func TestAdv_Degen_ReassignLoop100(t *testing.T) {
	got := runE2E(t, "i = [0]\nfor { if i == [100] { break }\ni = i + [1] }\n|> str(i)")
	if got != "[100]" {
		t.Fatalf("expected [100], got %q", got)
	}
}

func TestAdv_Degen_EmptyForInLargeArray(t *testing.T) {
	got := runE2E(t, "for x in [1,2,3,4,5,6,7,8,9,10] { }\n|> \"done\"")
	if got != "done" {
		t.Fatalf("expected done, got %q", got)
	}
}

func TestAdv_Degen_ManyPipeOutputs(t *testing.T) {
	var b strings.Builder
	for i := 1; i <= 10; i++ {
		fmt.Fprintf(&b, "|> str([%d])\n", i)
	}
	got := runE2E(t, b.String())
	if got != "[1][2][3][4][5][6][7][8][9][10]" {
		t.Fatalf("expected [1][2]...[10], got %q", got)
	}
}
