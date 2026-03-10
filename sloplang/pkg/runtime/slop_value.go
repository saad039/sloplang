package runtime

import (
	"fmt"
	"strings"
)

// SlopNull represents the null value in sloplang.
type SlopNull struct{}

// SlopValue is the universal value type in sloplang.
// All values are arrays of elements.
type SlopValue struct {
	Elements []any    // int64, uint64, float64, string, *SlopValue, or SlopNull
	Keys     []string // parallel to Elements for hashmaps; nil for plain arrays
}

// NewSlopValue creates a SlopValue from raw Go values.
// Accepted element types: int64, uint64, float64, string, *SlopValue.
func NewSlopValue(elems ...any) *SlopValue {
	return &SlopValue{Elements: elems}
}

// IsTruthy returns true only for [1] (single-element int64 with value 1).
// [] is falsy. Everything else panics — strict boolean semantics.
func (sv *SlopValue) IsTruthy() bool {
	if len(sv.Elements) == 0 {
		return false
	}
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got %d-element array", len(sv.Elements)))
	}
	elem := sv.Elements[0]
	if _, ok := elem.(SlopNull); ok {
		panic("sloplang: cannot use null as boolean")
	}
	i, ok := elem.(int64)
	if !ok {
		panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got single-element %T", elem))
	}
	if i == 1 {
		return true
	}
	if i == 0 {
		panic("sloplang: [0] is not a valid boolean — use [] for false")
	}
	panic(fmt.Sprintf("sloplang: boolean expression must be [1] or [], got [%d]", i))
}

// StdoutWrite prints a SlopValue to stdout without a trailing newline.
// Use explicit "\n" in the value for newlines.
func StdoutWrite(v *SlopValue) {
	fmt.Print(FormatValue(v))
}

// FormatValue returns the string representation of a SlopValue.
// Single-element strings print without brackets; everything else uses bracket notation.
func FormatValue(v *SlopValue) string {
	// Single-element string: print raw (no brackets)
	if len(v.Elements) == 1 {
		if s, ok := v.Elements[0].(string); ok {
			return s
		}
	}
	parts := make([]string, len(v.Elements))
	for i, elem := range v.Elements {
		parts[i] = formatElement(elem)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatElement(elem any) string {
	switch e := elem.(type) {
	case SlopNull:
		return "null"
	case int64:
		return fmt.Sprintf("%d", e)
	case uint64:
		return fmt.Sprintf("%d", e)
	case float64:
		return fmt.Sprintf("%g", e)
	case string:
		return e
	case *SlopValue:
		return FormatValue(e)
	default:
		return fmt.Sprintf("%v", e)
	}
}
