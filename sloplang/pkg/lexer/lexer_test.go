package lexer

import "testing"

func TestLexer_Assignment(t *testing.T) {
	l := New(`x = [1, 2, 3]`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_LBRACKET, "["},
		{TOKEN_INT, "1"},
		{TOKEN_COMMA, ","},
		{TOKEN_INT, "2"},
		{TOKEN_COMMA, ","},
		{TOKEN_INT, "3"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token %d: expected type %s, got %s", i, exp.typ, tok.Type)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("token %d: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLexer_StdoutWrite(t *testing.T) {
	l := New(`|> "hello world"`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_PIPE_GT, "|>"},
		{TOKEN_STRING, "hello world"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token %d: expected type %s, got %s", i, exp.typ, tok.Type)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("token %d: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLexer_NumberTypes(t *testing.T) {
	l := New(`42 42u 3.14`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_INT, "42"},
		{TOKEN_UINT, "42u"},
		{TOKEN_FLOAT, "3.14"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token %d: expected type %s, got %s (literal: %q)", i, exp.typ, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("token %d: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLexer_Comments(t *testing.T) {
	l := New("x = [1] // ignore this\n|> \"hi\"")
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_LBRACKET, "["},
		{TOKEN_INT, "1"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_PIPE_GT, "|>"},
		{TOKEN_STRING, "hi"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token %d: expected type %s, got %s", i, exp.typ, tok.Type)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("token %d: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLexer_StringEscapes(t *testing.T) {
	l := New(`"hello\nworld"`)
	tok := l.NextToken()
	if tok.Type != TOKEN_STRING {
		t.Fatalf("expected STRING, got %s", tok.Type)
	}
	if tok.Literal != "hello\nworld" {
		t.Fatalf("expected 'hello\\nworld', got %q", tok.Literal)
	}
}

func TestLexer_MultiLine(t *testing.T) {
	l := New("x = [1]\n|> \"hi\"")
	// Skip to |> token
	for i := 0; i < 5; i++ {
		l.NextToken()
	}
	tok := l.NextToken() // |>
	if tok.Line != 2 {
		t.Fatalf("expected line 2, got %d", tok.Line)
	}
}

func TestLexer_Empty(t *testing.T) {
	l := New("")
	tok := l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Fatalf("expected EOF, got %s", tok.Type)
	}
}

func TestLexer_Keywords(t *testing.T) {
	l := New("true false myVar")
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_TRUE, "true"},
		{TOKEN_FALSE, "false"},
		{TOKEN_IDENT, "myVar"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token %d: expected type %s, got %s", i, exp.typ, tok.Type)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("token %d: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLexer_HelloSlop(t *testing.T) {
	l := New("x = [1, 2, 3]\n|> \"hello world\"")
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_LBRACKET, "["},
		{TOKEN_INT, "1"},
		{TOKEN_COMMA, ","},
		{TOKEN_INT, "2"},
		{TOKEN_COMMA, ","},
		{TOKEN_INT, "3"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_PIPE_GT, "|>"},
		{TOKEN_STRING, "hello world"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ {
			t.Fatalf("token %d: expected type %s, got %s", i, exp.typ, tok.Type)
		}
		if tok.Literal != exp.lit {
			t.Fatalf("token %d: expected literal %q, got %q", i, exp.lit, tok.Literal)
		}
	}
}

func TestLexer_Tokenize(t *testing.T) {
	l := New(`x = [1]`)
	tokens := l.Tokenize()
	if len(tokens) != 6 {
		t.Fatalf("expected 6 tokens, got %d", len(tokens))
	}
	if tokens[len(tokens)-1].Type != TOKEN_EOF {
		t.Fatal("last token should be EOF")
	}
}
