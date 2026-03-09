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

func TestLexer_ArithmeticOperators(t *testing.T) {
	l := New(`+ - * / % **`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_PLUS, "+"}, {TOKEN_MINUS, "-"}, {TOKEN_STAR, "*"},
		{TOKEN_SLASH, "/"}, {TOKEN_PERCENT, "%"}, {TOKEN_POWER, "**"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_ComparisonOperators(t *testing.T) {
	l := New(`== != < > <= >=`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_EQ, "=="}, {TOKEN_NEQ, "!="}, {TOKEN_LT, "<"},
		{TOKEN_GT, ">"}, {TOKEN_LTE, "<="}, {TOKEN_GTE, ">="},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_LogicalOperators(t *testing.T) {
	l := New(`&& || !`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_AND, "&&"}, {TOKEN_OR, "||"}, {TOKEN_NOT, "!"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_Parentheses(t *testing.T) {
	l := New(`( )`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_LPAREN, "("}, {TOKEN_RPAREN, ")"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_OperatorDisambiguation(t *testing.T) {
	l := New(`= == ! != * ** |> || < <= <- << > >= >>`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_ASSIGN, "="}, {TOKEN_EQ, "=="}, {TOKEN_NOT, "!"}, {TOKEN_NEQ, "!="},
		{TOKEN_STAR, "*"}, {TOKEN_POWER, "**"}, {TOKEN_PIPE_GT, "|>"}, {TOKEN_OR, "||"},
		{TOKEN_LT, "<"}, {TOKEN_LTE, "<="}, {TOKEN_RETURN, "<-"}, {TOKEN_LSHIFT, "<<"},
		{TOKEN_GT, ">"}, {TOKEN_GTE, ">="}, {TOKEN_RSHIFT, ">>"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_FunctionCall(t *testing.T) {
	l := New(`str(x)`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "str"}, {TOKEN_LPAREN, "("}, {TOKEN_IDENT, "x"}, {TOKEN_RPAREN, ")"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_Phase3Keywords(t *testing.T) {
	l := New(`fn if else for in break`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FN, "fn"}, {TOKEN_IF, "if"}, {TOKEN_ELSE, "else"},
		{TOKEN_FOR, "for"}, {TOKEN_IN, "in"}, {TOKEN_BREAK, "break"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_Braces(t *testing.T) {
	l := New(`{ }`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_LBRACE, "{"}, {TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_ReturnOperator(t *testing.T) {
	l := New(`<- [1]`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_RETURN, "<-"}, {TOKEN_LBRACKET, "["}, {TOKEN_INT, "1"}, {TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_FnDecl(t *testing.T) {
	l := New(`fn add(a, b) { <- a + b }`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FN, "fn"}, {TOKEN_IDENT, "add"}, {TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "a"}, {TOKEN_COMMA, ","}, {TOKEN_IDENT, "b"},
		{TOKEN_RPAREN, ")"}, {TOKEN_LBRACE, "{"}, {TOKEN_RETURN, "<-"},
		{TOKEN_IDENT, "a"}, {TOKEN_PLUS, "+"}, {TOKEN_IDENT, "b"},
		{TOKEN_RBRACE, "}"}, {TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_IfElse(t *testing.T) {
	l := New(`if [1] { |> "yes" } else { |> "no" }`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IF, "if"}, {TOKEN_LBRACKET, "["}, {TOKEN_INT, "1"}, {TOKEN_RBRACKET, "]"},
		{TOKEN_LBRACE, "{"}, {TOKEN_PIPE_GT, "|>"}, {TOKEN_STRING, "yes"}, {TOKEN_RBRACE, "}"},
		{TOKEN_ELSE, "else"}, {TOKEN_LBRACE, "{"}, {TOKEN_PIPE_GT, "|>"}, {TOKEN_STRING, "no"}, {TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_InfiniteLoop(t *testing.T) {
	l := New(`for { break }`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FOR, "for"}, {TOKEN_LBRACE, "{"}, {TOKEN_BREAK, "break"}, {TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_ArrayOperators(t *testing.T) {
	l := New(`@ # << >> ~@ :: ++ -- ~ ??`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_AT, "@"}, {TOKEN_HASH, "#"}, {TOKEN_LSHIFT, "<<"},
		{TOKEN_RSHIFT, ">>"}, {TOKEN_TILDE_AT, "~@"}, {TOKEN_DCOLON, "::"},
		{TOKEN_CONCAT, "++"}, {TOKEN_REMOVE, "--"}, {TOKEN_TILDE, "~"},
		{TOKEN_CONTAINS, "??"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_ArrayOperatorDisambiguation(t *testing.T) {
	l := New(`< << <= <- > >> >= + ++ - -- ~ ~@`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_LT, "<"}, {TOKEN_LSHIFT, "<<"}, {TOKEN_LTE, "<="}, {TOKEN_RETURN, "<-"},
		{TOKEN_GT, ">"}, {TOKEN_RSHIFT, ">>"}, {TOKEN_GTE, ">="},
		{TOKEN_PLUS, "+"}, {TOKEN_CONCAT, "++"}, {TOKEN_MINUS, "-"}, {TOKEN_REMOVE, "--"},
		{TOKEN_TILDE, "~"}, {TOKEN_TILDE_AT, "~@"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_HashArr(t *testing.T) {
	l := New(`#arr`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_HASH, "#"}, {TOKEN_IDENT, "arr"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_IndexExpr(t *testing.T) {
	l := New(`arr@0`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "arr"}, {TOKEN_AT, "@"}, {TOKEN_INT, "0"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_PushExpr(t *testing.T) {
	l := New(`arr << [5]`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "arr"}, {TOKEN_LSHIFT, "<<"}, {TOKEN_LBRACKET, "["}, {TOKEN_INT, "5"}, {TOKEN_RBRACKET, "]"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_DoubleHashToken(t *testing.T) {
	l := New(`##map`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_DOUBLE_HASH, "##"}, {TOKEN_IDENT, "map"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_DoubleAtToken(t *testing.T) {
	l := New(`@@map`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_DOUBLE_AT, "@@"}, {TOKEN_IDENT, "map"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_DollarToken(t *testing.T) {
	l := New(`$key`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_DOLLAR, "$"}, {TOKEN_IDENT, "key"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_HashmapOperatorDisambiguation(t *testing.T) {
	l := New(`# ## @ @@ $`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_HASH, "#"}, {TOKEN_DOUBLE_HASH, "##"},
		{TOKEN_AT, "@"}, {TOKEN_DOUBLE_AT, "@@"},
		{TOKEN_DOLLAR, "$"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_NullKeyword(t *testing.T) {
	l := New(`null`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_NULL, "null"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_NullAssignment(t *testing.T) {
	l := New(`x = null`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_NULL, "null"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_StdinRead(t *testing.T) {
	l := New(`line, err = <|`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "line"}, {TOKEN_COMMA, ","}, {TOKEN_IDENT, "err"},
		{TOKEN_ASSIGN, "="}, {TOKEN_STDIN_READ, "<|"}, {TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_FileRead(t *testing.T) {
	l := New(`data, err = <. "file.txt"`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_IDENT, "data"}, {TOKEN_COMMA, ","}, {TOKEN_IDENT, "err"},
		{TOKEN_ASSIGN, "="}, {TOKEN_FILE_READ, "<."}, {TOKEN_STRING, "file.txt"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_FileWrite(t *testing.T) {
	l := New(`.> "file.txt" "data"`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FILE_WRITE, ".>"}, {TOKEN_STRING, "file.txt"}, {TOKEN_STRING, "data"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_FileAppend(t *testing.T) {
	l := New(`.>> "file.txt" "data"`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FILE_APPEND, ".>>"}, {TOKEN_STRING, "file.txt"}, {TOKEN_STRING, "data"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_IODisambiguation(t *testing.T) {
	l := New(`<= <- << <| <.`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_LTE, "<="}, {TOKEN_RETURN, "<-"}, {TOKEN_LSHIFT, "<<"},
		{TOKEN_STDIN_READ, "<|"}, {TOKEN_FILE_READ, "<."},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_DotDisambiguation(t *testing.T) {
	l := New(`.> .>>`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FILE_WRITE, ".>"}, {TOKEN_FILE_APPEND, ".>>"},
		{TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}

func TestLexer_ForIn(t *testing.T) {
	l := New(`for x in items { |> str(x) }`)
	expected := []struct {
		typ TokenType
		lit string
	}{
		{TOKEN_FOR, "for"}, {TOKEN_IDENT, "x"}, {TOKEN_IN, "in"}, {TOKEN_IDENT, "items"},
		{TOKEN_LBRACE, "{"}, {TOKEN_PIPE_GT, "|>"}, {TOKEN_IDENT, "str"},
		{TOKEN_LPAREN, "("}, {TOKEN_IDENT, "x"}, {TOKEN_RPAREN, ")"},
		{TOKEN_RBRACE, "}"}, {TOKEN_EOF, ""},
	}
	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.typ || tok.Literal != exp.lit {
			t.Fatalf("token %d: expected %s %q, got %s %q", i, exp.typ, exp.lit, tok.Type, tok.Literal)
		}
	}
}
