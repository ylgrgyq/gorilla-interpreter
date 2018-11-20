package lexer

import (
	"fmt"
	"testing"
	"token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 50;
	let ten = 10;

	let add = fn(x, y) {
		x + y;
	};

	  let result = add(five, ten);

	    		"hello"
	"world"
	"wor\t\n\r\"l\\d"
	"哈哈哈"
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
		expectedColumn  int
	}{
		{token.LET, "let", 1, 1},
		{token.IDENT, "five", 1, 5},
		{token.ASSIGN, "=", 1, 10},
		{token.INT, "50", 1, 12},
		{token.SEMICOLON, ";", 1, 14},

		{token.LET, "let", 2, 2},
		{token.IDENT, "ten", 2, 6},
		{token.ASSIGN, "=", 2, 10},
		{token.INT, "10", 2, 12},
		{token.SEMICOLON, ";", 2, 14},

		{token.LET, "let", 4, 2},
		{token.IDENT, "add", 4, 6},
		{token.ASSIGN, "=", 4, 10},
		{token.FUNCTION, "fn", 4, 12},
		{token.LPAREN, "(", 4, 14},
		{token.IDENT, "x", 4, 15},
		{token.COMMA, ",", 4, 16},
		{token.IDENT, "y", 4, 18},
		{token.RPAREN, ")", 4, 19},
		{token.LBRACE, "{", 4, 21},
		{token.IDENT, "x", 5, 3},
		{token.PLUS, "+", 5, 5},
		{token.IDENT, "y", 5, 7},
		{token.SEMICOLON, ";", 5, 8},
		{token.RBRACE, "}", 6, 2},
		{token.SEMICOLON, ";", 6, 3},

		{token.LET, "let", 8, 4},
		{token.IDENT, "result", 8, 8},
		{token.ASSIGN, "=", 8, 15},
		{token.IDENT, "add", 8, 17},
		{token.LPAREN, "(", 8, 20},
		{token.IDENT, "five", 8, 21},
		{token.COMMA, ",", 8, 25},
		{token.IDENT, "ten", 8, 27},
		{token.RPAREN, ")", 8, 30},
		{token.SEMICOLON, ";", 8, 31},

		/*
			"hello"
			"world"
			"wor\t\n\r\"ld"
			"哈哈哈"
		*/
		{token.STRING, "hello", 10, 8},
		{token.STRING, "world", 11, 2},
		{token.STRING, "wor\t\n\r\"l\\d", 12, 2},
		{token.STRING, "哈哈哈", 13, 2},

		{token.EOF, "", 14, 2},
	}

	handler := func(pos token.Position, msg string) {
		panic(fmt.Sprintf("%s at line: %d, column: %d", msg, pos.Line, pos.Column))
	}
	l := New(input, handler)
	for _, test := range tests {
		testLexer(t, l, test.expectedType, test.expectedLiteral, test.expectedLine, test.expectedColumn)
	}
}

func TestOperatorToken(t *testing.T) {
	input := `! -/*5;
	5 < 10 > 5;

	if (5 < 10) {
		return true;
	} else {
		return false;
	}

	10 == 9
	10 != 9
	
	plusplus++;
	minusminus--
	--minus
	++plus;

	array[15]
	{a:1, b:2}

	5 <= 10 >= 5;
	a -= 10
	a += 10
	a /= 10
	a *= 10
	a % 10
	a %= 10
	5 || 10
	5 && 10
	5 & 10
	5 | 10
	5 ^ 10
	a |= 5
	b &= 5
	b ^= 10
	a << 5
	a >> 10
	a <<= 5
	a >>= 10
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
		expectedLine    int
		expectedColumn  int
	}{

		// ! -/*5;
		{token.BANG, "!", 1, 1},
		{token.MINUS, "-", 1, 3},
		{token.DIVIDE, "/", 1, 4},
		{token.ASTERISK, "*", 1, 5},
		{token.INT, "5", 1, 6},
		{token.SEMICOLON, ";", 1, 7},

		// 5 < 10 > 5;
		{token.INT, "5", 2, 2},
		{token.LT, "<", 2, 4},
		{token.INT, "10", 2, 6},
		{token.GT, ">", 2, 9},
		{token.INT, "5", 2, 11},
		{token.SEMICOLON, ";", 2, 12},

		/*
			if (5 < 10) {
				return true;
			} else {
				return false;
			}
		*/
		{token.IF, "if", 4, 2},
		{token.LPAREN, "(", 4, 5},
		{token.INT, "5", 4, 6},
		{token.LT, "<", 4, 8},
		{token.INT, "10", 4, 10},
		{token.RPAREN, ")", 4, 12},
		{token.LBRACE, "{", 4, 14},
		{token.RETURN, "return", 5, 3},
		{token.TRUE, "true", 5, 10},
		{token.SEMICOLON, ";", 5, 14},
		{token.RBRACE, "}", 6, 2},
		{token.ELSE, "else", 6, 4},
		{token.LBRACE, "{", 6, 9},
		{token.RETURN, "return", 7, 3},
		{token.FALSE, "false", 7, 10},
		{token.SEMICOLON, ";", 7, 15},
		{token.RBRACE, "}", 8, 2},

		/*
			10 == 9
			10 != 9
		*/
		{token.INT, "10", 10, 2},
		{token.EQ, "==", 10, 5},
		{token.INT, "9", 10, 8},
		{token.INT, "10", 11, 2},
		{token.NOTEQ, "!=", 11, 5},
		{token.INT, "9", 11, 8},
		/*
			plusplus++;
			minusminus--
			--minus
			++plus;
		*/
		{token.IDENT, "plusplus", 13, 2},
		{token.INCREASE, "++", 13, 10},
		{token.SEMICOLON, ";", 13, 12},
		{token.IDENT, "minusminus", 14, 2},
		{token.DECREASE, "--", 14, 12},
		{token.DECREASE, "--", 15, 2},
		{token.IDENT, "minus", 15, 4},
		{token.INCREASE, "++", 16, 2},
		{token.IDENT, "plus", 16, 4},
		{token.SEMICOLON, ";", 16, 8},

		/*
			array[15]
		*/
		{token.IDENT, "array", 18, 2},
		{token.LBRACKET, "[", 18, 7},
		{token.INT, "15", 18, 8},
		{token.RBRACKET, "]", 18, 10},

		/*
			{a:1, b:2}
		*/
		{token.LBRACE, "{", 19, 2},
		{token.IDENT, "a", 19, 3},
		{token.COLON, ":", 19, 4},
		{token.INT, "1", 19, 5},
		{token.COMMA, ",", 19, 6},
		{token.IDENT, "b", 19, 8},
		{token.COLON, ":", 19, 9},
		{token.INT, "2", 19, 10},
		{token.RBRACE, "}", 19, 11},

		/*
			5 <= 10 >= 5;
			a -= 10
			a += 10
			a /= 10
			a *= 10
			a % 10
			a %= 10
			5 || 10
			5 && 10
			5 & 10
			5 | 10
			5 ^ 10
			a |= 5
			b &= 5
			b ^= 10
			a << 5
			a >> 10
			a <<= 5
			a >>= 10
		*/

		{token.INT, "5", 21, 2},
		{token.LTE, "<=", 21, 4},
		{token.INT, "10", 21, 7},
		{token.GTE, ">=", 21, 10},
		{token.INT, "5", 21, 13},
		{token.SEMICOLON, ";", 21, 14},

		{token.IDENT, "a", 22, 2},
		{token.MINUS_ASSIGN, "-=", 22, 4},
		{token.INT, "10", 22, 7},
		{token.IDENT, "a", 23, 2},
		{token.PLUS_ASSIGN, "+=", 23, 4},
		{token.INT, "10", 23, 7},
		{token.IDENT, "a", 24, 2},
		{token.DIVIDE_ASSIGN, "/=", 24, 4},
		{token.INT, "10", 24, 7},
		{token.IDENT, "a", 25, 2},
		{token.ASTERISK_ASSIGN, "*=", 25, 4},
		{token.INT, "10", 25, 7},
		{token.IDENT, "a", 26, 2},
		{token.REM, "%", 26, 4},
		{token.INT, "10", 26, 6},
		{token.IDENT, "a", 27, 2},
		{token.REM_ASSIGN, "%=", 27, 4},
		{token.INT, "10", 27, 7},
		{token.INT, "5", 28, 2},
		{token.LOR, "||", 28, 4},
		{token.INT, "10", 28, 7},
		{token.INT, "5", 29, 2},
		{token.LAND, "&&", 29, 4},
		{token.INT, "10", 29, 7},

		{token.INT, "5", 30, 2},
		{token.AND, "&", 30, 4},
		{token.INT, "10", 30, 6},
		{token.INT, "5", 31, 2},
		{token.OR, "|", 31, 4},
		{token.INT, "10", 31, 6},

		{token.INT, "5", 32, 2},
		{token.XOR, "^", 32, 4},
		{token.INT, "10", 32, 6},
		{token.IDENT, "a", 33, 2},
		{token.OR_ASSIGN, "|=", 33, 4},
		{token.INT, "5", 33, 7},
		{token.IDENT, "b", 34, 2},
		{token.AND_ASSIGN, "&=", 34, 4},
		{token.INT, "5", 34, 7},

		{token.IDENT, "b", 35, 2},
		{token.XOR_ASSIGN, "^=", 35, 4},
		{token.INT, "10", 35, 7},
		{token.IDENT, "a", 36, 2},
		{token.LSHIFT, "<<", 36, 4},
		{token.INT, "5", 36, 7},
		{token.IDENT, "a", 37, 2},
		{token.RSHIFT, ">>", 37, 4},
		{token.INT, "10", 37, 7},
		{token.IDENT, "a", 38, 2},
		{token.LSHIFT_ASSIGN, "<<=", 38, 4},
		{token.INT, "5", 38, 8},
		{token.IDENT, "a", 39, 2},
		{token.RSHIFT_ASSIGN, ">>=", 39, 4},
		{token.INT, "10", 39, 8},

		{token.EOF, "", 40, 2},
	}

	handler := func(pos token.Position, msg string) {
		panic(fmt.Sprintf("%s at line: %d, column: %d", msg, pos.Line, pos.Column))
	}
	l := New(input, handler)
	for _, test := range tests {
		testLexer(t, l, test.expectedType, test.expectedLiteral, test.expectedLine, test.expectedColumn)
	}
}

func TestLexerError(t *testing.T) {
	tests := []struct {
		input    string
		errorMsg string
	}{
		{"hello你好", "Unrecognized character at line: 1, column: 6"},
		{`hello * 1;
		"哈哈哈哈`, "EOF while reading string at line: 2, column: 4"},
		{`a + b;
		c + d;
		 "哈哈哈哈\`, "EOF while reading string at line: 3, column: 5"},
		{`"哈哈哈\x哈\"`, "Unsupported escape character at line: 1, column: 6"},
	}

	handler := func(pos token.Position, msg string) {
		panic(fmt.Sprintf("%s at line: %d, column: %d", msg, pos.Line, pos.Column))
	}

	for _, test := range tests {
		l := New(test.input, handler)
		testError(t, l, test.errorMsg)
	}
}

func testError(t *testing.T, l *Lexer, expectErrorMsg string) {
	defer func() {
		if r := recover(); r != nil {
			if r != expectErrorMsg {
				t.Errorf("%q - token Literal wrong. expected %q, got %q", l.lineAtPosition(), expectErrorMsg, r)
			}
		}
	}()
	for {
		tk := l.NextToken()
		if tk.Type == token.EOF {
			t.Fatalf("%q - token type wrong. missing ILLEGAL", l.lineAtPosition())
		}
	}
}
func testLexer(t *testing.T, l *Lexer, expectedType token.TokenType, expectedLiteral string, expectedLine, expectedColumn int) {
	tk := l.NextToken()

	if tk.Type != expectedType {
		t.Fatalf("%q - token type wrong. expected=%q, got=%q",
			l.lineAtPosition(), expectedType, tk.Type)
	}

	if tk.Literal != expectedLiteral {
		t.Fatalf("%q - token literal wrong. expected=%q, got=%q",
			l.lineAtPosition(), expectedLiteral, tk.Literal)
	}

	if tk.Pos.Line != expectedLine {
		t.Fatalf("%q - token line wrong. expected=%d, got=%d",
			l.lineAtPosition(), expectedLine, tk.Pos.Line)
	}

	if tk.Pos.Column != expectedColumn {
		t.Fatalf("%q - token column wrong. expect token %q with literal %q at %d, got=%d",
			l.lineAtPosition(), expectedType, expectedLiteral, expectedColumn, tk.Pos.Column)
	}
}

func (l *Lexer) lineAtPosition() string {
	start := l.offset
	if start == len(l.input) {
		start--
	}

	for start > 0 && l.input[start] != '\n' && l.input[start] != '\r' {
		start--
	}

	end := l.offset
	for end < len(l.input) && l.input[end] != '\n' && l.input[end] != '\r' {
		end++
	}
	return string(l.input[start:end])
}
