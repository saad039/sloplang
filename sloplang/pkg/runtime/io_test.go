package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileReadWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	FileWrite(NewSlopValue(path), NewSlopValue("hello file"))

	data, err := FileRead(NewSlopValue(path))
	if extractString(data) != "hello file" {
		t.Fatalf("expected %q, got %q", "hello file", extractString(data))
	}
	if err.Elements[0].(int64) != 0 {
		t.Fatal("expected err=0")
	}
}

func TestFileAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	FileWrite(NewSlopValue(path), NewSlopValue("first"))
	FileAppend(NewSlopValue(path), NewSlopValue(" second"))

	data, err := FileRead(NewSlopValue(path))
	if extractString(data) != "first second" {
		t.Fatalf("expected %q, got %q", "first second", extractString(data))
	}
	if err.Elements[0].(int64) != 0 {
		t.Fatal("expected err=0")
	}
}

func TestFileReadMissing(t *testing.T) {
	data, err := FileRead(NewSlopValue("/nonexistent/path/file.txt"))
	if extractString(data) != "" {
		t.Fatalf("expected empty string, got %q", extractString(data))
	}
	if err.Elements[0].(int64) != 1 {
		t.Fatal("expected err=1")
	}
}

func TestFileWriteOverwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	FileWrite(NewSlopValue(path), NewSlopValue("first"))
	FileWrite(NewSlopValue(path), NewSlopValue("second"))

	got, _ := os.ReadFile(path)
	if string(got) != "second" {
		t.Fatalf("expected %q, got %q", "second", string(got))
	}
}

func TestSplit_BySpace(t *testing.T) {
	result := Split(NewSlopValue("a b c"), NewSlopValue(" "))
	if len(result.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(result.Elements))
	}
	expected := []string{"a", "b", "c"}
	for i, exp := range expected {
		if result.Elements[i].(string) != exp {
			t.Fatalf("element %d: expected %q, got %q", i, exp, result.Elements[i])
		}
	}
}

func TestSplit_ByNewline(t *testing.T) {
	result := Split(NewSlopValue("a\nb"), NewSlopValue("\n"))
	if len(result.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result.Elements))
	}
	if result.Elements[0].(string) != "a" || result.Elements[1].(string) != "b" {
		t.Fatalf("unexpected elements: %v", result.Elements)
	}
}

func TestSplit_EmptySep(t *testing.T) {
	result := Split(NewSlopValue("hello"), NewSlopValue(""))
	if len(result.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(result.Elements))
	}
	if result.Elements[0].(string) != "hello" {
		t.Fatalf("expected %q, got %q", "hello", result.Elements[0])
	}
}

func TestSplit_NonStringPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-string argument")
		}
	}()
	Split(NewSlopValue(int64(42)), NewSlopValue(" "))
}

func TestToNum_Int(t *testing.T) {
	val, err := ToNum(NewSlopValue("42"))
	if val.Elements[0].(int64) != 42 {
		t.Fatalf("expected 42, got %v", val.Elements[0])
	}
	if err.Elements[0].(int64) != 0 {
		t.Fatal("expected err=0")
	}
}

func TestToNum_Float(t *testing.T) {
	val, err := ToNum(NewSlopValue("3.14"))
	if val.Elements[0].(float64) != 3.14 {
		t.Fatalf("expected 3.14, got %v", val.Elements[0])
	}
	if err.Elements[0].(int64) != 0 {
		t.Fatal("expected err=0")
	}
}

func TestToNum_Fail(t *testing.T) {
	val, err := ToNum(NewSlopValue("abc"))
	if len(val.Elements) != 0 {
		t.Fatalf("expected empty elements, got %d", len(val.Elements))
	}
	if err.Elements[0].(int64) != 1 {
		t.Fatal("expected err=1")
	}
}

func TestToNum_NegativeInt(t *testing.T) {
	val, err := ToNum(NewSlopValue("-5"))
	if val.Elements[0].(int64) != -5 {
		t.Fatalf("expected -5, got %v", val.Elements[0])
	}
	if err.Elements[0].(int64) != 0 {
		t.Fatal("expected err=0")
	}
}
