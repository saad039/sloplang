# Phase 7: Error Handling Patterns

## Overview

Verification and test hardening phase. The dual-return error propagation infrastructure is already working (via `UnpackTwo` for user functions, direct dual-return for builtins). This phase adds comprehensive E2E tests to make the error handling contract concrete and prevent regressions.

No runtime or codegen changes required.

## Key Findings from Verification

- User functions returning `<- [result, errcode]` work correctly via `UnpackTwo` path
- Error propagation across function boundaries works (wrapping `<.`, `to_num`, or other user fns)
- Chained propagation works (fn A calls fn B, propagates B's error)
- Roadmap expected output uses `[5]` but `FormatValue` returns `5` for single-element values (known pattern)

## Test Categories

### 1. Basic dual-return from user functions
- `safe_div(a, b)` — success and division-by-zero paths
- Return `[result, [0]]` on success, `[[], [1]]` on failure

### 2. Error propagation across function boundaries
- Function wrapping `<.` (file read) and re-returning error
- Function wrapping `to_num` and re-returning error
- Function calling another user function and propagating its error

### 3. Chained error propagation
- Multi-level chains: `step3` → `step2` → `step1`, error at any level propagates up
- Verify error code preserved through chain

### 4. Mixed error/success paths
- Same function called with valid and invalid inputs in sequence
- Verify each call independently returns correct result/error

### 5. File I/O errors
- Read nonexistent file → error code `[1]`
- Write then read → success with correct data
- Read after write verifying round-trip

### 6. to_num parse failures
- Valid integer string → success
- Valid float string → success
- Invalid string → error code `[1]`
- Empty string → error code `[1]`

### 7. Division by zero guarded
- User-level check before division prevents runtime panic
- Returns error code instead of crashing

### 8. Empty/null edge cases
- Returning `[]` as error result
- Error code comparison with `[0]`
- Empty array vs non-empty distinction in error checking

### 9. Error codes in control flow
- Using error codes in if/else branching
- Accumulating errors in a loop (process list of items, count failures)

### 10. Panic tests (unguarded errors)
- Unguarded division by zero still panics
- Type mismatch in arithmetic still panics
- Array length mismatch still panics
- These confirm panics are NOT silently swallowed
