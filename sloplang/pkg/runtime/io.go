package runtime

import (
	"bufio"
	"fmt"
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
