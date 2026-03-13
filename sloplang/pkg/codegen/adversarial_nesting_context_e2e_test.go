package codegen

import "testing"

// --- break in various contexts (5 tests) ---

func TestAdv_Nest_BreakInForBody(t *testing.T) {
	// GIVEN: break inside a for-in loop
	// WHEN: executed
	// THEN: exits loop, prints done
	got := runE2E(t, "for x in [1, 2] { break }\n|> \"done\"")
	if got != "done" {
		t.Fatalf("expected done, got %q", got)
	}
}

func TestAdv_Nest_BreakInIfInsideFor(t *testing.T) {
	// GIVEN: break inside if inside for
	// WHEN: executed
	// THEN: exits loop, prints done
	got := runE2E(t, "for x in [1, 2] { if true { break } }\n|> \"done\"")
	if got != "done" {
		t.Fatalf("expected done, got %q", got)
	}
}

func TestAdv_Nest_BreakInInfiniteLoop(t *testing.T) {
	// GIVEN: break inside infinite loop
	// WHEN: executed
	// THEN: exits loop, prints done
	got := runE2E(t, "for { break }\n|> \"done\"")
	if got != "done" {
		t.Fatalf("expected done, got %q", got)
	}
}

func TestAdv_Nest_BreakInIfOutsideLoop(t *testing.T) {
	// GIVEN: break in if outside any loop
	// WHEN: compiled
	// THEN: Go compile error
	runE2EExpectCompileError(t, `if true { break }`)
}

func TestAdv_Nest_BreakInFnOutsideLoop(t *testing.T) {
	// GIVEN: break inside function but outside loop
	// WHEN: compiled
	// THEN: Go compile error
	runE2EExpectCompileError(t, `fn f() { break }`)
}

// --- <- (return) in various contexts (3 tests) ---

func TestAdv_Nest_ReturnInFnBody(t *testing.T) {
	// GIVEN: return in function body
	// WHEN: executed
	// THEN: returns value
	got := runE2E(t, "fn f() { <- [1] }\n|> str(f())")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Nest_ReturnInIfInsideFn(t *testing.T) {
	// GIVEN: return inside if inside function
	// WHEN: executed
	// THEN: returns from if branch
	got := runE2E(t, "fn f() { if true { <- [1] }\n<- [2] }\n|> str(f())")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_Nest_ReturnInForInsideFn(t *testing.T) {
	// GIVEN: return inside for inside function
	// WHEN: executed
	// THEN: returns from first iteration
	got := runE2E(t, "fn f() { for x in [1, 2] { <- x }\n<- [0] }\n|> str(f())")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

// --- for in various contexts (2 tests) ---

func TestAdv_Nest_ForInIfBody(t *testing.T) {
	// GIVEN: for loop inside if body
	// WHEN: executed
	// THEN: loop executes normally
	got := runE2E(t, "if true { for x in [1, 2] { |> str(x) } }")
	if got != "[1][2]" {
		t.Fatalf("expected [1][2], got %q", got)
	}
}

func TestAdv_Nest_InfiniteForInFn(t *testing.T) {
	// GIVEN: infinite loop with return inside function
	// WHEN: executed
	// THEN: returns when condition met
	got := runE2E(t, "fn f() { i = [0]\nfor { i = i + [1]\nif i == [3] { <- i } } }\n|> str(f())")
	if got != "[3]" {
		t.Fatalf("expected [3], got %q", got)
	}
}

// --- fn in invalid contexts (2 tests) ---

func TestAdv_Nest_FnInsideFor(t *testing.T) {
	// GIVEN: function declaration inside for loop, called outside
	// WHEN: transpiled and compiled
	// THEN: Go compile error — f is not hoisted to top-level scope
	runE2EExpectCompileError(t, "for x in [1] { fn f() { <- [1] } }\n|> str(f())")
}

func TestAdv_Nest_FnInsideIf(t *testing.T) {
	// GIVEN: function declaration inside if body, called outside
	// WHEN: transpiled and compiled
	// THEN: Go compile error — f is not hoisted to top-level scope
	runE2EExpectCompileError(t, "if true { fn f() { <- [1] } }\n|> str(f())")
}
