package token

type TokenType int

type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

const (
	ILLEGAL TokenType = iota
	EOF

	literal_start
	IDENT
	INT
	STRING
	literal_end

	infix_operators_start
	ASSIGN
	PLUS
	MINUS
	ASTERISK
	DIVIDE
	REM

	OR
	AND
	XOR
	LSHIFT
	RSHIFT

	LOR
	LAND

	LT
	LTE
	GT
	GTE

	EQ
	NOTEQ

	PLUS_ASSIGN
	MINUS_ASSIGN
	ASTERISK_ASSIGN
	DIVIDE_ASSIGN
	REM_ASSIGN

	OR_ASSIGN
	AND_ASSIGN
	XOR_ASSIGN
	LSHIFT_ASSIGN
	RSHIFT_ASSIGN
	infix_operators_end

	BANG
	INCREASE
	DECREASE

	LBRACKET
	RBRACKET

	// Delimiters
	COMMA
	SEMICOLON
	COLON

	LPAREN
	RPAREN
	LBRACE
	RBRACE

	keywords_start
	FUNCTION
	LET
	IF
	ELSE
	RETURN
	NULL
	TRUE
	FALSE
	keywords_end
)

var tokenLiteral = [...]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",

	IDENT:  "IDENT",
	INT:    "INT",
	STRING: "STRING",

	ASSIGN:   "=",
	PLUS:     "+",
	MINUS:    "-",
	ASTERISK: "*",
	DIVIDE:   "/",
	REM:      "%",

	OR:     "|",
	AND:    "&",
	XOR:    "^",
	LSHIFT: "<<",
	RSHIFT: ">>",

	LOR:  "LOR",
	LAND: "LAND",

	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">=",

	EQ:    "==",
	NOTEQ: "!=",

	PLUS_ASSIGN:     "+=",
	MINUS_ASSIGN:    "-=",
	ASTERISK_ASSIGN: "*=",
	DIVIDE_ASSIGN:   "/=",
	REM_ASSIGN:      "%=",

	OR_ASSIGN:     "|=",
	AND_ASSIGN:    "&=",
	XOR_ASSIGN:    "^=",
	LSHIFT_ASSIGN: "<<=",
	RSHIFT_ASSIGN: ">>=",

	BANG:     "!",
	INCREASE: "++",
	DECREASE: "--",

	LBRACKET: "[",
	RBRACKET: "]",

	COMMA:     ",",
	SEMICOLON: ";",
	COLON:     ":",

	LPAREN: "(",
	RPAREN: ")",
	LBRACE: "{",
	RBRACE: "}",

	FUNCTION: "fn",
	LET:      "let",
	IF:       "if",
	ELSE:     "else",
	RETURN:   "return",
	NULL:     "null",
	TRUE:     "true",
	FALSE:    "false",
}

var keywords map[string]TokenType

func GetLiteral(tp TokenType) string {
	return tokenLiteral[tp]
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}

	return IDENT
}

const (
	LOWEST_PRECEDENCE = 0
)

// Precedence get precedence of the operator tokens
func (op Token) Precedence() int {
	switch op.Type {
	case
		ASSIGN, PLUS_ASSIGN, MINUS_ASSIGN, ASTERISK_ASSIGN,
		DIVIDE_ASSIGN, REM_ASSIGN, OR_ASSIGN, AND_ASSIGN,
		XOR_ASSIGN, LSHIFT_ASSIGN, RSHIFT_ASSIGN:
		return 1
	case LOR:
		return 2
	case LAND:
		return 3
	case EQ, NOTEQ:
		return 4
	case LT, LTE, GT, GTE:
		return 5
	case PLUS, MINUS, OR, XOR:
		return 6
	case ASTERISK, DIVIDE, REM, LSHIFT, RSHIFT, AND:
		return 7
	case BANG, LPAREN, INCREASE, DECREASE:
		return 8
	case LBRACKET:
		return 9
	}
	return LOWEST_PRECEDENCE
}

func GetPrefixOperators() (ops []TokenType) {
	return []TokenType{MINUS, BANG}
}

func GetInfixOperators() (ops []TokenType) {
	for i := infix_operators_start; i < infix_operators_end; i++ {
		ops = append(ops, i)
	}
	return ops
}

func GetPostfixOperators() (ops []TokenType) {
	return []TokenType{INCREASE, DECREASE}
}

func init() {
	keywords = make(map[string]TokenType)
	for i := keywords_start; i < keywords_end; i++ {
		keywords[tokenLiteral[i]] = i
	}
}
