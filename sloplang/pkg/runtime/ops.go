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
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) == 0)
}

func Neq(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "!=")
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) != 0)
}

func Lt(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "<")
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) < 0)
}

func Gt(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, ">")
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) > 0)
}

func Lte(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, "<=")
	return boolResult(compareElems(a.Elements[0], b.Elements[0]) <= 0)
}

func Gte(a, b *SlopValue) *SlopValue {
	checkSingleElement(a, b, ">=")
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
// Used by for-in loops.
func Iterate(sv *SlopValue) []*SlopValue {
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
