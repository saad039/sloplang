package runtime

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestNewSlopValue_Int(t *testing.T) {
	sv := NewSlopValue(int64(42))
	if len(sv.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sv.Elements))
	}
	if v, ok := sv.Elements[0].(int64); !ok || v != 42 {
		t.Fatalf("expected int64(42), got %v", sv.Elements[0])
	}
}

func TestNewSlopValue_Uint(t *testing.T) {
	sv := NewSlopValue(uint64(255))
	if len(sv.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sv.Elements))
	}
	if v, ok := sv.Elements[0].(uint64); !ok || v != 255 {
		t.Fatalf("expected uint64(255), got %v", sv.Elements[0])
	}
}

func TestNewSlopValue_Float(t *testing.T) {
	sv := NewSlopValue(float64(3.14))
	if len(sv.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sv.Elements))
	}
	if v, ok := sv.Elements[0].(float64); !ok || v != 3.14 {
		t.Fatalf("expected float64(3.14), got %v", sv.Elements[0])
	}
}

func TestNewSlopValue_String(t *testing.T) {
	sv := NewSlopValue("hello")
	if len(sv.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(sv.Elements))
	}
	if v, ok := sv.Elements[0].(string); !ok || v != "hello" {
		t.Fatalf("expected string(hello), got %v", sv.Elements[0])
	}
}

func TestNewSlopValue_Array(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	if len(sv.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(sv.Elements))
	}
	for i, expected := range []int64{1, 2, 3} {
		if v, ok := sv.Elements[i].(int64); !ok || v != expected {
			t.Fatalf("element %d: expected int64(%d), got %v", i, expected, sv.Elements[i])
		}
	}
}

func TestNewSlopValue_Empty(t *testing.T) {
	sv := NewSlopValue()
	if len(sv.Elements) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(sv.Elements))
	}
}

func TestIsTruthy_NonEmpty(t *testing.T) {
	sv := NewSlopValue(int64(0))
	if !sv.IsTruthy() {
		t.Fatal("[0] should be truthy")
	}
}

func TestIsTruthy_Empty(t *testing.T) {
	sv := NewSlopValue()
	if sv.IsTruthy() {
		t.Fatal("[] should be falsy")
	}
}

func TestFormatValue_SingleInt(t *testing.T) {
	sv := NewSlopValue(int64(42))
	if got := FormatValue(sv); got != "42" {
		t.Fatalf("expected '42', got '%s'", got)
	}
}

func TestFormatValue_IntArray(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	if got := FormatValue(sv); got != "[1, 2, 3]" {
		t.Fatalf("expected '[1, 2, 3]', got '%s'", got)
	}
}

func TestFormatValue_Nested(t *testing.T) {
	inner := NewSlopValue(int64(1), int64(2))
	outer := NewSlopValue(inner, int64(3))
	if got := FormatValue(outer); got != "[[1, 2], 3]" {
		t.Fatalf("expected '[[1, 2], 3]', got '%s'", got)
	}
}

func TestFormatValue_Empty(t *testing.T) {
	sv := NewSlopValue()
	if got := FormatValue(sv); got != "[]" {
		t.Fatalf("expected '[]', got '%s'", got)
	}
}

func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestStdoutWrite_String(t *testing.T) {
	sv := NewSlopValue("hello world")
	out := captureStdout(func() { StdoutWrite(sv) })
	if out != "hello world\n" {
		t.Fatalf("expected 'hello world\\n', got %q", out)
	}
}

func TestStdoutWrite_IntArray(t *testing.T) {
	sv := NewSlopValue(int64(1), int64(2), int64(3))
	out := captureStdout(func() { StdoutWrite(sv) })
	if out != "[1, 2, 3]\n" {
		t.Fatalf("expected '[1, 2, 3]\\n', got %q", out)
	}
}
