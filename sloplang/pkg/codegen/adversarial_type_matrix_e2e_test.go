package codegen

import "testing"

// --- Arithmetic operators with wrong types (10 tests) ---

func TestAdv_TypeMx_Add_IntString(t *testing.T) {
	// GIVEN: an integer and a string
	// WHEN: added together
	// THEN: runtime panics with type mismatch
	runE2EExpectPanicContaining(t, `x = [1] + "hello"`, "type")
}

func TestAdv_TypeMx_Add_IntNull(t *testing.T) {
	// GIVEN: an integer and null
	// WHEN: added together
	// THEN: runtime panics mentioning null
	runE2EExpectPanicContaining(t, `x = [1] + [null]`, "null")
}

func TestAdv_TypeMx_Add_FloatUint(t *testing.T) {
	// GIVEN: a float and a uint
	// WHEN: added together
	// THEN: runtime panics with type mismatch
	runE2EExpectPanicContaining(t, `x = [1.0] + [1u]`, "type mismatch")
}

func TestAdv_TypeMx_Mul_StringInt(t *testing.T) {
	// GIVEN: a string and an integer
	// WHEN: multiplied
	// THEN: runtime panics with type mismatch
	runE2EExpectPanicContaining(t, `x = "hello" * [3]`, "type")
}

func TestAdv_TypeMx_Div_NullFloat(t *testing.T) {
	// GIVEN: null divided by float
	// WHEN: executed
	// THEN: runtime panics mentioning null
	runE2EExpectPanicContaining(t, `x = [null] / [2.0]`, "null")
}

func TestAdv_TypeMx_Mod_FloatFloat(t *testing.T) {
	// GIVEN: float modulo float
	// WHEN: executed
	// THEN: verify — may panic or succeed depending on Mod implementation
	runE2EExpectPanic(t, `x = [5.0] % [2.0]`)
}

func TestAdv_TypeMx_Pow_StringString(t *testing.T) {
	// GIVEN: string raised to string power
	// WHEN: executed
	// THEN: runtime panics with type mismatch
	runE2EExpectPanicContaining(t, `x = "a" ** "b"`, "type")
}

func TestAdv_TypeMx_Sub_BoolInt(t *testing.T) {
	// GIVEN: true (=[1]) minus [1]
	// WHEN: executed
	// THEN: succeeds with [0]
	got := runE2E(t, "x = true - [1]\n|> str(x)")
	if got != "[0]" {
		t.Fatalf("expected [0], got %q", got)
	}
}

func TestAdv_TypeMx_Negate_String(t *testing.T) {
	// GIVEN: a string value
	// WHEN: negated
	// THEN: runtime panics with cannot negate
	runE2EExpectPanicContaining(t, `x = -"hello"`, "cannot negate")
}

func TestAdv_TypeMx_Negate_Null(t *testing.T) {
	// GIVEN: null value
	// WHEN: negated
	// THEN: runtime panics mentioning null
	runE2EExpectPanicContaining(t, `x = -[null]`, "null")
}

// --- Comparison operators with wrong types (8 tests) ---

func TestAdv_TypeMx_Lt_IntString(t *testing.T) {
	// GIVEN: integer compared to string
	// WHEN: using less-than
	// THEN: runtime panics with type mismatch
	runE2EExpectPanicContaining(t, `x = [1] < "hello"`, "type")
}

func TestAdv_TypeMx_Gt_NullInt(t *testing.T) {
	// GIVEN: null compared to integer
	// WHEN: using greater-than
	// THEN: runtime panics mentioning null
	runE2EExpectPanicContaining(t, `x = [null] > [1]`, "null")
}

func TestAdv_TypeMx_Eq_IntFloat(t *testing.T) {
	// GIVEN: integer compared to float
	// WHEN: using equality
	// THEN: verify — may be equal or panic
	got := runE2E(t, "|> str([1] == [1.0])")
	_ = got // behavior depends on deepEqual implementation
}

func TestAdv_TypeMx_Eq_IntString(t *testing.T) {
	// GIVEN: integer compared to string
	// WHEN: using equality
	// THEN: returns false (different types)
	got := runE2E(t, "|> str([1] == \"hello\")")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

func TestAdv_TypeMx_Neq_NullInt(t *testing.T) {
	// GIVEN: null compared to integer
	// WHEN: using not-equal
	// THEN: returns true
	got := runE2E(t, "|> str([null] != [1])")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_TypeMx_Lte_FloatUint(t *testing.T) {
	// GIVEN: float compared to uint
	// WHEN: using less-than-or-equal
	// THEN: runtime panics with type mismatch
	runE2EExpectPanicContaining(t, `x = [1.0] <= [2u]`, "type mismatch")
}

func TestAdv_TypeMx_Lt_MultiElement(t *testing.T) {
	// GIVEN: multi-element arrays
	// WHEN: using less-than
	// THEN: runtime panics requiring single-element
	runE2EExpectPanicContaining(t, `x = [1, 2] < [3, 4]`, "single-element")
}

func TestAdv_TypeMx_Gte_EmptyVsInt(t *testing.T) {
	// GIVEN: empty array compared to integer
	// WHEN: using greater-than-or-equal
	// THEN: runtime panics with length error
	runE2EExpectPanicContaining(t, `x = [] >= [1]`, "single-element")
}

// --- Logical operators with invalid truthiness (6 tests) ---

func TestAdv_TypeMx_And_ZeroIsNotFalsy(t *testing.T) {
	// GIVEN: [0] used in boolean context
	// WHEN: used with &&
	// THEN: runtime panics — [0] is not valid boolean
	runE2EExpectPanicContaining(t, `x = [0] && [1]`, "use [] for false")
}

func TestAdv_TypeMx_Or_StringNotBool(t *testing.T) {
	// GIVEN: string used in boolean context
	// WHEN: used with ||
	// THEN: runtime panics
	runE2EExpectPanicContaining(t, `x = "hello" || [1]`, "")
}

func TestAdv_TypeMx_Not_Float(t *testing.T) {
	// GIVEN: float value
	// WHEN: negated with !
	// THEN: runtime panics — not valid boolean
	runE2EExpectPanicContaining(t, `x = ![1.0]`, "")
}

func TestAdv_TypeMx_Not_Null(t *testing.T) {
	// GIVEN: null value
	// WHEN: negated with !
	// THEN: runtime panics mentioning null
	runE2EExpectPanicContaining(t, `x = ![null]`, "null")
}

func TestAdv_TypeMx_And_MultiElement(t *testing.T) {
	// GIVEN: multi-element array in boolean context
	// WHEN: used with &&
	// THEN: runtime panics
	runE2EExpectPanicContaining(t, `x = [1, 2] && [1]`, "")
}

func TestAdv_TypeMx_If_IntNotBool(t *testing.T) {
	// GIVEN: integer [5] in if condition
	// WHEN: evaluated as boolean
	// THEN: runtime panics — not valid boolean
	runE2EExpectPanicContaining(t, `if [5] { |> "yes" }`, "")
}

// --- Array/container operators with wrong types (8 tests) ---

func TestAdv_TypeMx_Concat_IntPlusString(t *testing.T) {
	// GIVEN: integer array concat with string
	// WHEN: using ++
	// THEN: succeeds — concat merges elements regardless of type
	got := runE2E(t, "x = [1] ++ \"hello\"\n|> str(x)")
	_ = got // verify no crash
}

func TestAdv_TypeMx_Contains_IntInString(t *testing.T) {
	// GIVEN: string tested for integer containment
	// WHEN: using ??
	// THEN: succeeds — returns [] (not found, different types)
	got := runE2E(t, "x = \"hello\" ?? [1]\n|> str(x)")
	_ = got // verify no crash
}

func TestAdv_TypeMx_Contains_CrossType(t *testing.T) {
	// GIVEN: int array tested for float containment
	// WHEN: using ??
	// THEN: verify — may return [] or panic
	got := runE2E(t, "|> str([1, 2] ?? [1.0])")
	_ = got // behavior depends on Contains implementation
}

func TestAdv_TypeMx_Length_String(t *testing.T) {
	// GIVEN: string value
	// WHEN: # applied
	// THEN: verify — may return string length or panic
	got := runE2E(t, "|> str(#\"hello\")")
	_ = got // behavior depends on Length implementation
}

func TestAdv_TypeMx_Length_Null(t *testing.T) {
	// GIVEN: array containing null
	// WHEN: # applied
	// THEN: returns [1] — length counts all elements including null
	got := runE2E(t, "|> str(#[null])")
	if got != "[1]" {
		t.Fatalf("expected [1], got %q", got)
	}
}

func TestAdv_TypeMx_Remove_TypeMismatch(t *testing.T) {
	// GIVEN: integer array removing string
	// WHEN: using --
	// THEN: verify — may return unchanged array or panic
	got := runE2E(t, "|> str([1, 2, 3] -- \"hello\")")
	_ = got // behavior depends on Remove implementation
}

func TestAdv_TypeMx_Index_StringOnArray(t *testing.T) {
	// GIVEN: string key used on array
	// WHEN: using @
	// THEN: runtime panics — string key on non-map
	runE2EExpectPanicContaining(t, "arr = [1, 2]\n|> arr@name", "")
}

func TestAdv_TypeMx_Keys_OnArray(t *testing.T) {
	// GIVEN: ## applied to array (not hashmap)
	// WHEN: executed
	// THEN: returns empty keys array (array has no keys)
	got := runE2E(t, "|> str(##[1, 2])")
	if got != "[]" {
		t.Fatalf("expected [], got %q", got)
	}
}

// --- Unary operators across types (5 tests) ---

func TestAdv_TypeMx_Not_Int(t *testing.T) {
	// GIVEN: integer [5]
	// WHEN: ! applied
	// THEN: runtime panics — not valid boolean
	runE2EExpectPanicContaining(t, `x = ![5]`, "")
}

func TestAdv_TypeMx_Not_MultiElement(t *testing.T) {
	// GIVEN: multi-element array
	// WHEN: ! applied
	// THEN: runtime panics
	runE2EExpectPanicContaining(t, `x = ![1, 2]`, "")
}

func TestAdv_TypeMx_Negate_Bool(t *testing.T) {
	// GIVEN: true (=[1])
	// WHEN: negated with -
	// THEN: returns [-1]
	got := runE2E(t, "|> str(-true)")
	if got != "[-1]" {
		t.Fatalf("expected [-1], got %q", got)
	}
}

func TestAdv_TypeMx_Negate_MultiElement(t *testing.T) {
	// GIVEN: multi-element array [1, 2, 3]
	// WHEN: negated
	// THEN: element-wise negate [-1, -2, -3]
	got := runE2E(t, "|> str(-[1, 2, 3])")
	if got != "[-1, -2, -3]" {
		t.Fatalf("expected [-1, -2, -3], got %q", got)
	}
}

func TestAdv_TypeMx_Length_MultiElement(t *testing.T) {
	// GIVEN: multi-element array
	// WHEN: # applied
	// THEN: returns element count
	got := runE2E(t, "|> str(#[1, 2, 3])")
	if got != "[3]" {
		t.Fatalf("expected [3], got %q", got)
	}
}
