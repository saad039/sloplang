package runtime

import (
	"fmt"
	"math"
)

func checkLengths(a, b *SlopValue) {
	if len(a.Elements) != len(b.Elements) {
		panic(fmt.Sprintf("sloplang: length mismatch: %d vs %d", len(a.Elements), len(b.Elements)))
	}
}

func binaryOp(a, b *SlopValue, op func(x, y any) any) *SlopValue {
	checkLengths(a, b)
	result := make([]any, len(a.Elements))
	for i := range a.Elements {
		if _, ok := a.Elements[i].(SlopNull); ok {
			panic("sloplang: cannot perform arithmetic on null")
		}
		if _, ok := b.Elements[i].(SlopNull); ok {
			panic("sloplang: cannot perform arithmetic on null")
		}
		result[i] = op(a.Elements[i], b.Elements[i])
	}
	return &SlopValue{Elements: result}
}

func Add(a, b *SlopValue) *SlopValue {
	return binaryOp(a, b, func(x, y any) any {
		switch xv := x.(type) {
		case int64:
			yv, ok := y.(int64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in +: int64 vs %T", y))
			}
			return xv + yv
		case uint64:
			yv, ok := y.(uint64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in +: uint64 vs %T", y))
			}
			return xv + yv
		case float64:
			yv, ok := y.(float64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in +: float64 vs %T", y))
			}
			return xv + yv
		default:
			panic(fmt.Sprintf("sloplang: unsupported type for +: %T", x))
		}
	})
}

func Sub(a, b *SlopValue) *SlopValue {
	return binaryOp(a, b, func(x, y any) any {
		switch xv := x.(type) {
		case int64:
			yv, ok := y.(int64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in -: int64 vs %T", y))
			}
			return xv - yv
		case uint64:
			yv, ok := y.(uint64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in -: uint64 vs %T", y))
			}
			return xv - yv
		case float64:
			yv, ok := y.(float64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in -: float64 vs %T", y))
			}
			return xv - yv
		default:
			panic(fmt.Sprintf("sloplang: unsupported type for -: %T", x))
		}
	})
}

func Mul(a, b *SlopValue) *SlopValue {
	return binaryOp(a, b, func(x, y any) any {
		switch xv := x.(type) {
		case int64:
			yv, ok := y.(int64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in *: int64 vs %T", y))
			}
			return xv * yv
		case uint64:
			yv, ok := y.(uint64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in *: uint64 vs %T", y))
			}
			return xv * yv
		case float64:
			yv, ok := y.(float64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in *: float64 vs %T", y))
			}
			return xv * yv
		default:
			panic(fmt.Sprintf("sloplang: unsupported type for *: %T", x))
		}
	})
}

func Div(a, b *SlopValue) *SlopValue {
	return binaryOp(a, b, func(x, y any) any {
		switch xv := x.(type) {
		case int64:
			yv, ok := y.(int64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in /: int64 vs %T", y))
			}
			if yv == 0 {
				panic("sloplang: division by zero")
			}
			return xv / yv
		case uint64:
			yv, ok := y.(uint64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in /: uint64 vs %T", y))
			}
			if yv == 0 {
				panic("sloplang: division by zero")
			}
			return xv / yv
		case float64:
			yv, ok := y.(float64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in /: float64 vs %T", y))
			}
			if yv == 0 {
				panic("sloplang: division by zero")
			}
			return xv / yv
		default:
			panic(fmt.Sprintf("sloplang: unsupported type for /: %T", x))
		}
	})
}

func Mod(a, b *SlopValue) *SlopValue {
	return binaryOp(a, b, func(x, y any) any {
		switch xv := x.(type) {
		case int64:
			yv, ok := y.(int64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in %%: int64 vs %T", y))
			}
			return xv % yv
		case uint64:
			yv, ok := y.(uint64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in %%: uint64 vs %T", y))
			}
			return xv % yv
		default:
			panic(fmt.Sprintf("sloplang: unsupported type for %%: %T", x))
		}
	})
}

func Pow(a, b *SlopValue) *SlopValue {
	return binaryOp(a, b, func(x, y any) any {
		switch xv := x.(type) {
		case int64:
			yv, ok := y.(int64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in **: int64 vs %T", y))
			}
			return int64(math.Pow(float64(xv), float64(yv)))
		case uint64:
			yv, ok := y.(uint64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in **: uint64 vs %T", y))
			}
			return uint64(math.Pow(float64(xv), float64(yv)))
		case float64:
			yv, ok := y.(float64)
			if !ok {
				panic(fmt.Sprintf("sloplang: type mismatch in **: float64 vs %T", y))
			}
			return math.Pow(xv, yv)
		default:
			panic(fmt.Sprintf("sloplang: unsupported type for **: %T", x))
		}
	})
}

func Negate(a *SlopValue) *SlopValue {
	result := make([]any, len(a.Elements))
	for i, elem := range a.Elements {
		switch e := elem.(type) {
		case SlopNull:
			panic("sloplang: cannot negate null")
		case int64:
			result[i] = -e
		case uint64:
			result[i] = -int64(e)
		case float64:
			result[i] = -e
		case *SlopValue:
			neg := Negate(e)
			if len(neg.Elements) == 1 {
				result[i] = neg.Elements[0]
			} else {
				result[i] = neg
			}
		default:
			panic(fmt.Sprintf("sloplang: cannot negate %T", elem))
		}
	}
	return &SlopValue{Elements: result}
}

// Comparison operations — single-element arrays only

func boolResult(b bool) *SlopValue {
	if b {
		return NewSlopValue(int64(1))
	}
	return NewSlopValue()
}

func checkSingleElement(a, b *SlopValue, op string) {
	if len(a.Elements) != 1 || len(b.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: %s requires single-element arrays, got lengths %d and %d", op, len(a.Elements), len(b.Elements)))
	}
}

func compareElems(a, b any) int {
	switch av := a.(type) {
	case int64:
		bv, ok := b.(int64)
		if !ok {
			panic(fmt.Sprintf("sloplang: type mismatch in comparison: int64 vs %T", b))
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
		return 0
	case uint64:
		bv, ok := b.(uint64)
		if !ok {
			panic(fmt.Sprintf("sloplang: type mismatch in comparison: uint64 vs %T", b))
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
		return 0
	case float64:
		bv, ok := b.(float64)
		if !ok {
			panic(fmt.Sprintf("sloplang: type mismatch in comparison: float64 vs %T", b))
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
		return 0
	case string:
		bv, ok := b.(string)
		if !ok {
			panic(fmt.Sprintf("sloplang: type mismatch in comparison: string vs %T", b))
		}
		if av < bv {
			return -1
		}
		if av > bv {
			return 1
		}
		return 0
	default:
		panic(fmt.Sprintf("sloplang: unsupported type for comparison: %T", a))
	}
}

func Eq(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "==")
	_, aIsNull := a.Elements[0].(SlopNull)
	_, bIsNull := b.Elements[0].(SlopNull)
	if aIsNull && bIsNull {
		return NewSlopValue(int64(1))
	}
	if aIsNull || bIsNull {
		return &SlopValue{}
	}
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) == 0)
}

func Neq(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "!=")
	_, aIsNull := a.Elements[0].(SlopNull)
	_, bIsNull := b.Elements[0].(SlopNull)
	if aIsNull && bIsNull {
		return &SlopValue{}
	}
	if aIsNull || bIsNull {
		return NewSlopValue(int64(1))
	}
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) != 0)
}

func checkNullComparison(a, b *SlopValue) {
	if _, ok := a.Elements[0].(SlopNull); ok {
		panic("sloplang: cannot compare null with ordered operators")
	}
	if _, ok := b.Elements[0].(SlopNull); ok {
		panic("sloplang: cannot compare null with ordered operators")
	}
}

func Lt(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "<")
	checkNullComparison(a, b)
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) < 0)
}

func Gt(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, ">")
	checkNullComparison(a, b)
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) > 0)
}

func Lte(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "<=")
	checkNullComparison(a, b)
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) <= 0)
}

func Gte(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, ">=")
	checkNullComparison(a, b)
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) >= 0)
}

// Logical operations — operate on truthiness

func And(a, b *SlopValue) *SlopValue {
	if a.IsTruthy() {
		return b
	}
	return NewSlopValue()
}

func Or(a, b *SlopValue) *SlopValue {
	if a.IsTruthy() {
		return a
	}
	return b
}

func Not(a *SlopValue) *SlopValue {
	return boolResult(!a.IsTruthy())
}

// Str converts a SlopValue to its string representation.
func Str(a *SlopValue) *SlopValue {
	return NewSlopValue(FormatValue(a))
}

// Iterate returns each element of a SlopValue as its own *SlopValue.
// Used by for-in loops. Panics if the SlopValue is a single null element.
func Iterate(sv *SlopValue) []*SlopValue {
	if len(sv.Elements) == 1 {
		if _, ok := sv.Elements[0].(SlopNull); ok {
			panic("sloplang: cannot iterate over null")
		}
	}
	result := make([]*SlopValue, len(sv.Elements))
	for i, elem := range sv.Elements {
		if nested, ok := elem.(*SlopValue); ok {
			result[i] = nested
		} else {
			result[i] = NewSlopValue(elem)
		}
	}
	return result
}

// deepEqual compares two element values for structural equality.
func deepEqual(a, b any) bool {
	switch av := a.(type) {
	case SlopNull:
		_, ok := b.(SlopNull)
		return ok
	case int64:
		bv, ok := b.(int64)
		return ok && av == bv
	case uint64:
		bv, ok := b.(uint64)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case *SlopValue:
		bv, ok := b.(*SlopValue)
		if !ok || len(av.Elements) != len(bv.Elements) {
			return false
		}
		for i := range av.Elements {
			if !deepEqual(av.Elements[i], bv.Elements[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// Index returns the element at the given index as a *SlopValue.
func Index(sv *SlopValue, idx *SlopValue) *SlopValue {
	if len(idx.Elements) != 1 {
		panic("sloplang: index must be a single-element int64 array")
	}
	i, ok := idx.Elements[0].(int64)
	if !ok {
		panic(fmt.Sprintf("sloplang: index must be int64, got %T", idx.Elements[0]))
	}
	if i < 0 || int(i) >= len(sv.Elements) {
		panic(fmt.Sprintf("sloplang: index out of bounds: %d (length %d)", i, len(sv.Elements)))
	}
	elem := sv.Elements[i]
	if nested, ok := elem.(*SlopValue); ok {
		return nested
	}
	return NewSlopValue(elem)
}

// IndexSet sets the element at the given index to val. Mutates sv. Returns sv.
func IndexSet(sv *SlopValue, idx *SlopValue, val *SlopValue) *SlopValue {
	if len(idx.Elements) != 1 {
		panic("sloplang: index must be a single-element int64 array")
	}
	i, ok := idx.Elements[0].(int64)
	if !ok {
		panic(fmt.Sprintf("sloplang: index must be int64, got %T", idx.Elements[0]))
	}
	if i < 0 || int(i) >= len(sv.Elements) {
		panic(fmt.Sprintf("sloplang: index out of bounds: %d (length %d)", i, len(sv.Elements)))
	}
	sv.Elements[i] = val
	return sv
}

// Length returns the number of elements as a single-element SlopValue.
func Length(sv *SlopValue) *SlopValue {
	return NewSlopValue(int64(len(sv.Elements)))
}

// Push appends val's elements to sv. Mutates sv. Returns sv.
func Push(sv *SlopValue, val *SlopValue) *SlopValue {
	sv.Elements = append(sv.Elements, val.Elements...)
	return sv
}

// Pop removes and returns the last element. Mutates sv. Panics if empty.
func Pop(sv *SlopValue) *SlopValue {
	if len(sv.Elements) == 0 {
		panic("sloplang: pop from empty array")
	}
	last := sv.Elements[len(sv.Elements)-1]
	sv.Elements = sv.Elements[:len(sv.Elements)-1]
	if nested, ok := last.(*SlopValue); ok {
		return nested
	}
	return NewSlopValue(last)
}

// RemoveAt removes the element at the given index and returns it. Mutates sv.
func RemoveAt(sv *SlopValue, idx *SlopValue) *SlopValue {
	if len(idx.Elements) != 1 {
		panic("sloplang: index must be a single-element int64 array")
	}
	i, ok := idx.Elements[0].(int64)
	if !ok {
		panic(fmt.Sprintf("sloplang: index must be int64, got %T", idx.Elements[0]))
	}
	if i < 0 || int(i) >= len(sv.Elements) {
		panic(fmt.Sprintf("sloplang: index out of bounds: %d (length %d)", i, len(sv.Elements)))
	}
	removed := sv.Elements[i]
	sv.Elements = append(sv.Elements[:i], sv.Elements[i+1:]...)
	if nested, ok := removed.(*SlopValue); ok {
		return nested
	}
	return NewSlopValue(removed)
}

// Slice returns a new SlopValue with elements from low to high (exclusive).
func Slice(sv *SlopValue, low, high *SlopValue) *SlopValue {
	if len(low.Elements) != 1 || len(high.Elements) != 1 {
		panic("sloplang: slice bounds must be single-element int64 arrays")
	}
	lo, ok1 := low.Elements[0].(int64)
	hi, ok2 := high.Elements[0].(int64)
	if !ok1 || !ok2 {
		panic("sloplang: slice bounds must be int64")
	}
	if lo < 0 || hi < lo || int(hi) > len(sv.Elements) {
		panic(fmt.Sprintf("sloplang: slice bounds out of range [%d:%d] (length %d)", lo, hi, len(sv.Elements)))
	}
	// Create a new slice to avoid aliasing
	elems := make([]any, hi-lo)
	copy(elems, sv.Elements[lo:hi])
	return &SlopValue{Elements: elems}
}

// Concat returns a NEW SlopValue with all elements of a then b. Does NOT mutate.
func Concat(a, b *SlopValue) *SlopValue {
	elems := make([]any, 0, len(a.Elements)+len(b.Elements))
	elems = append(elems, a.Elements...)
	elems = append(elems, b.Elements...)
	return &SlopValue{Elements: elems}
}

// Remove removes the first occurrence of val.Elements[0] from sv.
// Returns a new SlopValue. Uses deepEqual for comparison.
func Remove(sv *SlopValue, val *SlopValue) *SlopValue {
	if len(val.Elements) == 0 {
		// Nothing to remove; return a copy
		elems := make([]any, len(sv.Elements))
		copy(elems, sv.Elements)
		return &SlopValue{Elements: elems}
	}
	target := val.Elements[0]
	elems := make([]any, 0, len(sv.Elements))
	found := false
	for _, e := range sv.Elements {
		if !found && deepEqual(e, target) {
			found = true
			continue
		}
		elems = append(elems, e)
	}
	return &SlopValue{Elements: elems}
}

// Contains checks if val.Elements[0] is in sv.Elements.
// Returns [1] if found, [] if not.
func Contains(sv *SlopValue, val *SlopValue) *SlopValue {
	if len(val.Elements) != 1 {
		panic("sloplang: contains operand must be a single-element array")
	}
	target := val.Elements[0]
	for _, e := range sv.Elements {
		if deepEqual(e, target) {
			return NewSlopValue(int64(1))
		}
	}
	return NewSlopValue()
}

// Unique returns a new SlopValue with duplicates removed (keep first).
func Unique(sv *SlopValue) *SlopValue {
	var elems []any
	for _, e := range sv.Elements {
		found := false
		for _, existing := range elems {
			if deepEqual(e, existing) {
				found = true
				break
			}
		}
		if !found {
			elems = append(elems, e)
		}
	}
	return &SlopValue{Elements: elems}
}

// MapFromKeysValues creates a SlopValue with named keys.
// keys are the field names; vals contains the corresponding elements.
// Panics if len(keys) != len(vals.Elements) and both are non-zero.
func MapFromKeysValues(keys []string, vals *SlopValue) *SlopValue {
	if len(keys) > 0 && len(vals.Elements) > 0 && len(keys) != len(vals.Elements) {
		panic(fmt.Sprintf("sloplang: hashmap key count (%d) != value count (%d)", len(keys), len(vals.Elements)))
	}
	elems := make([]any, len(vals.Elements))
	copy(elems, vals.Elements)
	k := make([]string, len(keys))
	copy(k, keys)
	return &SlopValue{Elements: elems, Keys: k}
}

// IndexKeyStr finds key in sv.Keys and returns the corresponding element as *SlopValue.
// Panics if the key is not found.
func IndexKeyStr(sv *SlopValue, key string) *SlopValue {
	for i, k := range sv.Keys {
		if k == key {
			elem := sv.Elements[i]
			if nested, ok := elem.(*SlopValue); ok {
				return nested
			}
			return NewSlopValue(elem)
		}
	}
	panic(fmt.Sprintf("sloplang: key %q not found in hashmap", key))
}

// IndexKey extracts a string key from key (must be single-element string) and calls IndexKeyStr.
func IndexKey(sv *SlopValue, key *SlopValue) *SlopValue {
	if len(key.Elements) != 1 {
		panic("sloplang: dynamic key must be a single-element string array")
	}
	s, ok := key.Elements[0].(string)
	if !ok {
		panic(fmt.Sprintf("sloplang: dynamic key must be string, got %T", key.Elements[0]))
	}
	return IndexKeyStr(sv, s)
}

// IndexKeySetStr sets or adds a key-value pair in a hashmap. Mutates sv. Returns sv.
// If key exists, updates the corresponding element. If not, appends key and element.
func IndexKeySetStr(sv *SlopValue, key string, val *SlopValue) *SlopValue {
	for i, k := range sv.Keys {
		if k == key {
			sv.Elements[i] = val
			return sv
		}
	}
	// Key not found — append
	sv.Keys = append(sv.Keys, key)
	sv.Elements = append(sv.Elements, val)
	return sv
}

// IndexKeySet extracts a string key from key and calls IndexKeySetStr.
func IndexKeySet(sv *SlopValue, key *SlopValue, val *SlopValue) *SlopValue {
	if len(key.Elements) != 1 {
		panic("sloplang: dynamic key must be a single-element string array")
	}
	s, ok := key.Elements[0].(string)
	if !ok {
		panic(fmt.Sprintf("sloplang: dynamic key must be string, got %T", key.Elements[0]))
	}
	return IndexKeySetStr(sv, s, val)
}

// MapKeys returns a new SlopValue with each key as a string element.
func MapKeys(sv *SlopValue) *SlopValue {
	if sv.Keys == nil {
		return NewSlopValue()
	}
	elems := make([]any, len(sv.Keys))
	for i, k := range sv.Keys {
		elems[i] = k
	}
	return &SlopValue{Elements: elems}
}

// MapValues returns a new SlopValue with the same elements but no keys.
func MapValues(sv *SlopValue) *SlopValue {
	if sv.Keys == nil {
		return NewSlopValue()
	}
	elems := make([]any, len(sv.Elements))
	copy(elems, sv.Elements)
	return &SlopValue{Elements: elems}
}

// UnpackTwo destructures a SlopValue into two *SlopValues.
// Used for multi-assignment: a, b = expr
func UnpackTwo(sv *SlopValue) (*SlopValue, *SlopValue) {
	if len(sv.Elements) < 2 {
		panic(fmt.Sprintf("sloplang: unpack requires at least 2 elements, got %d", len(sv.Elements)))
	}
	wrap := func(elem any) *SlopValue {
		if nested, ok := elem.(*SlopValue); ok {
			return nested
		}
		return NewSlopValue(elem)
	}
	return wrap(sv.Elements[0]), wrap(sv.Elements[1])
}
