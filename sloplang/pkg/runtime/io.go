package runtime

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var stdinScanner = bufio.NewScanner(os.Stdin)

func extractString(sv *SlopValue) string {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: expected single-element string, got %d elements", len(sv.Elements)))
	}
	s, ok := sv.Elements[0].(string)
	if !ok {
		panic(fmt.Sprintf("sloplang: expected string, got %T", sv.Elements[0]))
	}
	return s
}

// StdinRead reads one line from stdin.
// Returns (line, err) where err is [0] on success, [1] on EOF/error.
func StdinRead() (*SlopValue, *SlopValue) {
	if stdinScanner.Scan() {
		return NewSlopValue(stdinScanner.Text()), NewSlopValue(int64(0))
	}
	return NewSlopValue(""), NewSlopValue(int64(1))
}

// FileRead reads the entire contents of the file at path.
// Returns (data, err) where err is [0] on success, [1] on error.
func FileRead(path *SlopValue) (*SlopValue, *SlopValue) {
	pathStr := extractString(path)
	data, err := os.ReadFile(pathStr)
	if err != nil {
		return NewSlopValue(""), NewSlopValue(int64(1))
	}
	return NewSlopValue(string(data)), NewSlopValue(int64(0))
}

// FileWrite writes data to a file, creating or truncating it.
// Panics on error.
func FileWrite(path, data *SlopValue) {
	pathStr := extractString(path)
	dataStr := FormatValue(data)
	if err := os.WriteFile(pathStr, []byte(dataStr), 0644); err != nil {
		panic(fmt.Sprintf("sloplang: file write error: %v", err))
	}
}

// FileAppend appends data to a file, creating it if it doesn't exist.
// Panics on error.
func FileAppend(path, data *SlopValue) {
	pathStr := extractString(path)
	dataStr := FormatValue(data)
	f, err := os.OpenFile(pathStr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("sloplang: file append error: %v", err))
	}
	defer f.Close()
	if _, err := f.WriteString(dataStr); err != nil {
		panic(fmt.Sprintf("sloplang: file append error: %v", err))
	}
}

// Split splits a string by a separator.
// If sep is empty, returns the original string as-is.
func Split(sv, sep *SlopValue) *SlopValue {
	str := extractString(sv)
	sepStr := extractString(sep)
	if sepStr == "" {
		return NewSlopValue(str)
	}
	parts := strings.Split(str, sepStr)
	elems := make([]any, len(parts))
	for i, p := range parts {
		elems[i] = p
	}
	return &SlopValue{Elements: elems}
}

// Exit terminates the program with the given exit code.
// The argument must be a single-element integer array.
func Exit(sv *SlopValue) {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: exit() requires single-element int array, got %d elements", len(sv.Elements)))
	}
	code, ok := sv.Elements[0].(int64)
	if !ok {
		panic(fmt.Sprintf("sloplang: exit() requires int argument, got %T", sv.Elements[0]))
	}
	os.Exit(int(code))
}

// ToChars splits a string into an array of single-character strings.
// Panics if the argument is not a single-element string.
func ToChars(sv *SlopValue) *SlopValue {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: to_chars requires a string argument, got %d elements", len(sv.Elements)))
	}
	s, ok := sv.Elements[0].(string)
	if !ok {
		panic(fmt.Sprintf("sloplang: to_chars requires a string argument, got %T", sv.Elements[0]))
	}
	runes := []rune(s)
	elems := make([]any, len(runes))
	for i, r := range runes {
		elems[i] = string(r)
	}
	return &SlopValue{Elements: elems}
}

// ToInt converts a single-element numeric or string value to int64.
// float64 is truncated toward zero. Panics on invalid input.
func ToInt(sv *SlopValue) *SlopValue {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: to_int: requires single-element array, got %d elements", len(sv.Elements)))
	}
	switch v := sv.Elements[0].(type) {
	case int64:
		return NewSlopValue(v)
	case float64:
		return NewSlopValue(int64(v))
	case uint64:
		if v > uint64(math.MaxInt64) {
			panic("sloplang: to_int: uint64 value exceeds MaxInt64")
		}
		return NewSlopValue(int64(v))
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return NewSlopValue(i)
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return NewSlopValue(int64(f))
		}
		panic(fmt.Sprintf("sloplang: to_int: cannot convert string %q to int", v))
	default:
		panic(fmt.Sprintf("sloplang: to_int: cannot convert %T to int", v))
	}
}

// ToFloat converts a single-element numeric or string value to float64.
// Panics on invalid input.
func ToFloat(sv *SlopValue) *SlopValue {
	if len(sv.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: to_float: requires single-element array, got %d elements", len(sv.Elements)))
	}
	switch v := sv.Elements[0].(type) {
	case float64:
		return NewSlopValue(v)
	case int64:
		return NewSlopValue(float64(v))
	case uint64:
		return NewSlopValue(float64(v))
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return NewSlopValue(f)
		}
		panic(fmt.Sprintf("sloplang: to_float: cannot convert string %q to float", v))
	default:
		panic(fmt.Sprintf("sloplang: to_float: cannot convert %T to float", v))
	}
}

// FmtFloat formats a numeric value with a fixed number of decimal places.
// Returns a string. Panics if first arg is not numeric or second is not a non-negative int.
func FmtFloat(val, decimals *SlopValue) *SlopValue {
	if len(val.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: fmt_float: first argument must be numeric, got %d elements", len(val.Elements)))
	}
	if len(decimals.Elements) != 1 {
		panic(fmt.Sprintf("sloplang: fmt_float: second argument must be non-negative integer, got %d elements", len(decimals.Elements)))
	}
	d, ok := decimals.Elements[0].(int64)
	if !ok || d < 0 {
		panic("sloplang: fmt_float: second argument must be non-negative integer")
	}
	var f float64
	switch v := val.Elements[0].(type) {
	case float64:
		f = v
	case int64:
		f = float64(v)
	case uint64:
		f = float64(v)
	default:
		panic(fmt.Sprintf("sloplang: fmt_float: first argument must be numeric, got %T", v))
	}
	result := fmt.Sprintf("%.*f", int(d), f)
	return NewSlopValue(result)
}

// ToNum converts a string to a number.
// Returns (value, err) where err is [0] on success, [1] on failure.
// Tries int64 first, then float64.
func ToNum(sv *SlopValue) (*SlopValue, *SlopValue) {
	str := extractString(sv)
	if i, err := strconv.ParseInt(str, 10, 64); err == nil {
		return NewSlopValue(i), NewSlopValue(int64(0))
	}
	if f, err := strconv.ParseFloat(str, 64); err == nil {
		return NewSlopValue(f), NewSlopValue(int64(0))
	}
	return &SlopValue{}, NewSlopValue(int64(1))
}
