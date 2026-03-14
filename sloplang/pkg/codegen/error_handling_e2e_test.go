package codegen

import "testing"

// ============================================================
// Phase 7: Error Handling Patterns
// Validates dual-return error propagation across function
// boundaries, chained errors, and runtime panic behavior.
// ============================================================

// --- 1. Basic dual-return from user functions ----------------

func TestE2E_Phase7_SafeDivSuccess(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([10], [2])
|> str(result)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[5][0]" {
		t.Fatalf("expected %q, got %q", "[5][0]", got)
	}
}

func TestE2E_Phase7_SafeDivByZero(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([10], [0])
|> str(result)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[][1]" {
		t.Fatalf("expected %q, got %q", "[][1]", got)
	}
}

func TestE2E_Phase7_RoadmapErrorsSlop(t *testing.T) {
	// Full roadmap example from Phase 7 spec
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([10], [2])
|> str(result)
|> str(err)

result2, err2 = safe_div([10], [0])
|> str(result2)
|> str(err2)

data, err3 = <. "nonexistent.txt"
|> str(err3)
`
	got := runE2E(t, src)
	// FormatValue: single-element returns raw value (5, not [5])
	expected := "[5][0][][1][1]"
	if got != expected {
		t.Fatalf("expected %q, got %q", expected, got)
	}
}

// --- 2. Error propagation across function boundaries ---------

func TestE2E_Phase7_FnWrappingFileRead_Error(t *testing.T) {
	src := `
fn read_config(path) {
    data, err = <. path
    if err != [0] {
        <- ["", [1]]
    }
    <- [data, [0]]
}

content, err = read_config("nonexistent_xyz.txt")
|> str(err)
`
	got := runE2E(t, src)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_Phase7_FnWrappingFileRead_Success(t *testing.T) {
	src := `
fn read_config(path) {
    data, err = <. path
    if err != [0] {
        <- ["", [1]]
    }
    <- [data, [0]]
}

.> "test_config.txt" "hello_config"
content, err = read_config("test_config.txt")
|> str(err)
|> content
`
	got := runE2E(t, src)
	if got != "[0]hello_config" {
		t.Fatalf("expected %q, got %q", "[0]hello_config", got)
	}
}

func TestE2E_Phase7_FnWrappingToNum_Error(t *testing.T) {
	src := `
fn parse_int(s) {
    val, err = to_num(s)
    if err != [0] {
        <- [[], [1]]
    }
    <- [val, [0]]
}

result, err = parse_int("abc")
|> str(result)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[][1]" {
		t.Fatalf("expected %q, got %q", "[][1]", got)
	}
}

func TestE2E_Phase7_FnWrappingToNum_Success(t *testing.T) {
	src := `
fn parse_int(s) {
    val, err = to_num(s)
    if err != [0] {
        <- [[], [1]]
    }
    <- [val, [0]]
}

result, err = parse_int("42")
|> str(result)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[42][0]" {
		t.Fatalf("expected %q, got %q", "[42][0]", got)
	}
}

func TestE2E_Phase7_FnCallingUserFnPropagatesError(t *testing.T) {
	src := `
fn inner(x) {
    if x == [0] {
        <- [[], [1]]
    }
    <- [x * [10], [0]]
}

fn outer(x) {
    val, err = inner(x)
    if err != [0] {
        <- [[], err]
    }
    <- [val + [1], [0]]
}

r, e = outer([0])
|> str(r)
|> str(e)
`
	got := runE2E(t, src)
	if got != "[][1]" {
		t.Fatalf("expected %q, got %q", "[][1]", got)
	}
}

// --- 3. Chained error propagation ----------------------------

func TestE2E_Phase7_ThreeLevelChain_ErrorAtBottom(t *testing.T) {
	src := `
fn step1(x) {
    if x == [0] {
        <- [[], [1]]
    }
    <- [x, [0]]
}

fn step2(x) {
    val, err = step1(x)
    if err != [0] {
        <- [[], err]
    }
    <- [val * [2], [0]]
}

fn step3(x) {
    val, err = step2(x)
    if err != [0] {
        <- [[], err]
    }
    <- [val + [100], [0]]
}

r, e = step3([0])
|> str(r)
|> str(e)
`
	got := runE2E(t, src)
	if got != "[][1]" {
		t.Fatalf("expected %q, got %q", "[][1]", got)
	}
}

func TestE2E_Phase7_ThreeLevelChain_Success(t *testing.T) {
	src := `
fn step1(x) {
    if x == [0] {
        <- [[], [1]]
    }
    <- [x, [0]]
}

fn step2(x) {
    val, err = step1(x)
    if err != [0] {
        <- [[], err]
    }
    <- [val * [2], [0]]
}

fn step3(x) {
    val, err = step2(x)
    if err != [0] {
        <- [[], err]
    }
    <- [val + [100], [0]]
}

r, e = step3([5])
|> str(r)
|> str(e)
`
	got := runE2E(t, src)
	// step1([5]) -> [5, [0]], step2: 5*2=10 -> [10, [0]], step3: 10+100=110
	if got != "[110][0]" {
		t.Fatalf("expected %q, got %q", "[110][0]", got)
	}
}

func TestE2E_Phase7_ChainPreservesErrorCode(t *testing.T) {
	src := `
fn inner(x) {
    if x < [0] {
        <- [[], [2]]
    }
    if x == [0] {
        <- [[], [3]]
    }
    <- [x, [0]]
}

fn outer(x) {
    val, err = inner(x)
    if err != [0] {
        <- [[], err]
    }
    <- [val, [0]]
}

r1, e1 = outer(-([1]))
|> str(e1)

r2, e2 = outer([0])
|> str(e2)

r3, e3 = outer([5])
|> str(e3)
`
	got := runE2E(t, src)
	// Error code 2 for negative, 3 for zero, 0 for positive
	if got != "[2][3][0]" {
		t.Fatalf("expected %q, got %q", "[2][3][0]", got)
	}
}

// --- 4. Mixed error and success paths in sequence ------------

func TestE2E_Phase7_MixedSuccessThenError(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

r1, e1 = safe_div([10], [5])
|> str(r1)
|> str(e1)
r2, e2 = safe_div([10], [0])
|> str(r2)
|> str(e2)
`
	got := runE2E(t, src)
	if got != "[2][0][][1]" {
		t.Fatalf("expected %q, got %q", "[2][0][][1]", got)
	}
}

func TestE2E_Phase7_MixedErrorThenSuccess(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

r1, e1 = safe_div([10], [0])
|> str(r1)
|> str(e1)
r2, e2 = safe_div([10], [5])
|> str(r2)
|> str(e2)
`
	got := runE2E(t, src)
	if got != "[][1][2][0]" {
		t.Fatalf("expected %q, got %q", "[][1][2][0]", got)
	}
}

func TestE2E_Phase7_AlternatingSuccessFailure(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

r1, e1 = safe_div([12], [3])
|> str(r1)
r2, e2 = safe_div([12], [0])
|> str(r2)
r3, e3 = safe_div([12], [4])
|> str(r3)
r4, e4 = safe_div([12], [0])
|> str(r4)
`
	got := runE2E(t, src)
	if got != "[4][][3][]" {
		t.Fatalf("expected %q, got %q", "[4][][3][]", got)
	}
}

// --- 5. File I/O error handling patterns ---------------------

func TestE2E_Phase7_FileReadNonexistent(t *testing.T) {
	src := `
data, err = <. "does_not_exist_12345.txt"
|> str(err)
`
	got := runE2E(t, src)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_Phase7_FileWriteThenReadRoundtrip(t *testing.T) {
	src := `
.> "roundtrip.txt" "hello roundtrip"
data, err = <. "roundtrip.txt"
|> str(err)
|> data
`
	got := runE2E(t, src)
	if got != "[0]hello roundtrip" {
		t.Fatalf("expected %q, got %q", "[0]hello roundtrip", got)
	}
}

func TestE2E_Phase7_FileReadErrorDataIsEmpty(t *testing.T) {
	src := `
data, err = <. "nope_nope_nope.txt"
|> str(#data)
`
	got := runE2E(t, src)
	// Empty string "" has length 1 (single string element)
	// Actually: FileRead returns NewSlopValue("") which is a SlopValue with one element ("")
	// #data = [1]
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

// --- 6. to_num parse failure error handling ------------------

func TestE2E_Phase7_ToNumValidInt(t *testing.T) {
	src := `
val, err = to_num("123")
|> str(val)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[123][0]" {
		t.Fatalf("expected %q, got %q", "[123][0]", got)
	}
}

func TestE2E_Phase7_ToNumValidFloat(t *testing.T) {
	src := `
val, err = to_num("3.14")
|> str(val)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[3.14][0]" {
		t.Fatalf("expected %q, got %q", "[3.14][0]", got)
	}
}

func TestE2E_Phase7_ToNumInvalidString(t *testing.T) {
	src := `
val, err = to_num("not_a_number")
|> str(val)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[][1]" {
		t.Fatalf("expected %q, got %q", "[][1]", got)
	}
}

func TestE2E_Phase7_ToNumEmptyString(t *testing.T) {
	src := `
val, err = to_num("")
|> str(val)
|> str(err)
`
	got := runE2E(t, src)
	if got != "[][1]" {
		t.Fatalf("expected %q, got %q", "[][1]", got)
	}
}

// --- 7. Division by zero: guarded vs unguarded ---------------

func TestE2E_Phase7_DivByZeroGuarded(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([42], [0])
|> str(err)
`
	got := runE2E(t, src)
	if got != "[1]" {
		t.Fatalf("expected %q, got %q", "[1]", got)
	}
}

func TestE2E_Phase7_DivByZeroUnguardedPanics(t *testing.T) {
	src := `
x = [10] / [0]
|> str(x)
`
	runE2EExpectPanic(t, src)
}

// --- 8. Empty array and edge cases in error patterns ---------

func TestE2E_Phase7_EmptyArrayIsFalsy(t *testing.T) {
	src := `
fn maybe_fail() {
    <- [[], [1]]
}

result, err = maybe_fail()
if result {
    |> "truthy"
} else {
    |> "falsy"
}
`
	got := runE2E(t, src)
	if got != "falsy" {
		t.Fatalf("expected %q, got %q", "falsy", got)
	}
}

func TestE2E_Phase7_ErrorCodeZeroMeansSuccess(t *testing.T) {
	src := `
fn succeed() {
    <- [[42], [0]]
}

result, err = succeed()
if err == [0] {
    |> "success"
} else {
    |> "failure"
}
`
	got := runE2E(t, src)
	if got != "success" {
		t.Fatalf("expected %q, got %q", "success", got)
	}
}

func TestE2E_Phase7_ErrorCodeNonzeroMeansFailure(t *testing.T) {
	src := `
fn fail() {
    <- [[], [1]]
}

result, err = fail()
if err != [0] {
    |> "failure detected"
} else {
    |> "success"
}
`
	got := runE2E(t, src)
	if got != "failure detected" {
		t.Fatalf("expected %q, got %q", "failure detected", got)
	}
}

// --- 9. Error codes driving control flow ---------------------

func TestE2E_Phase7_ErrorCodeDrivesIfElse(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

result, err = safe_div([20], [4])
if err == [0] {
    |> "success"
    |> str(result)
} else {
    |> "error"
}
`
	got := runE2E(t, src)
	if got != "success[5]" {
		t.Fatalf("expected %q, got %q", "success[5]", got)
	}
}

func TestE2E_Phase7_LoopCountingFailures(t *testing.T) {
	src := `
fn safe_div(a, b) {
    if b == [0] {
        <- [[], [1]]
    }
    <- [a / b, [0]]
}

divisors = [2, 0, 5, 0, 1]
failures = [0]
for d in divisors {
    r, err = safe_div([10], d)
    if err != [0] {
        failures = failures + [1]
    }
}
|> str(failures)
`
	got := runE2E(t, src)
	if got != "[2]" {
		t.Fatalf("expected %q, got %q", "[2]", got)
	}
}

func TestE2E_Phase7_EarlyReturnOnFirstError(t *testing.T) {
	src := `
fn process_item(x) {
    if x < [0] {
        <- [[], [1]]
    }
    <- [x * [2], [0]]
}

fn process_all() {
    items = [3, 5, -([1]), 8]
    results = []
    for item in items {
        val, err = process_item(item)
        if err != [0] {
            |> "error encountered"
            <- [results, [1]]
        }
        results << val
    }
    <- [results, [0]]
}

out, err = process_all()
|> str(err)
|> str(out)
`
	got := runE2E(t, src)
	if got != "error encountered[1][6, 10]" {
		t.Fatalf("expected %q, got %q", "error encountered[1][6, 10]", got)
	}
}

// --- 10. Runtime panics for unguarded errors remain intact ----

func TestE2E_Phase7_PanicDivByZeroUnguarded(t *testing.T) {
	runE2EExpectPanic(t, `x = [10] / [0]`)
}

func TestE2E_Phase7_PanicTypeMismatchArithmetic(t *testing.T) {
	runE2EExpectPanic(t, `x = [1] + [1.0]`)
}

func TestE2E_Phase7_PanicLengthMismatchArithmetic(t *testing.T) {
	runE2EExpectPanic(t, `x = [1, 2] + [3]`)
}

func TestE2E_Phase7_PanicMultiElementComparison(t *testing.T) {
	// Multi-element comparison no longer panics — it succeeds
	_ = runE2E(t, `x = [1, 2] == [1, 2]`)
}
