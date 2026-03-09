package runtime

import (
	"fmt"
	"strings"
)

// SlopValue is the universal value type in sloplang.
// All values are arrays of elements.
type SlopValue struct {
	Elements []any    // int64, uint64, float64, string, or *SlopValue
	Keys     []string // parallel to Elements for hashmaps; nil for plain arrays
}

// NewSlopValue creates a SlopValue from raw Go values.
// Accepted element types: int64, uint64, float64, string, *SlopValue.
func NewSlopValue(elems ...any) *SlopValue {
	return &SlopValue{Elements: elems}
}

// IsTruthy returns true if the SlopValue is non-empty.
// [] is falsy, everything else (including [0]) is truthy.
func (sv *SlopValue) IsTruthy() bool {
	return len(sv.Elements) > 0
}

// StdoutWrite prints a SlopValue to stdout with a trailing newline.
func StdoutWrite(v *SlopValue) {
	if len(v.Elements) == 1 {
		if s, ok := v.Elements[0].(string); ok {
			fmt.Println(s)
			return
		}
	}
	fmt.Println(FormatValue(v))
}

// FormatValue returns the string representation of a SlopValue.
func FormatValue(v *SlopValue) string {
	if len(v.Elements) == 1 {
		return formatElement(v.Elements[0])
	}
	parts := make([]string, len(v.Elements))
	for i, elem := range v.Elements {
		parts[i] = formatElement(elem)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatElement(elem any) string {
	switch e := elem.(type) {
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
