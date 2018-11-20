package lexer

import (
	"token"
	"unicode"
	"unicode/utf8"
)

// ErrorHandler is used to handle error encountered during parsing input string
type ErrorHandler func(pos token.Position, msg string)

// A Lexer holding the state while scanning the input string
type Lexer struct {
	input        []byte       // the byte array being scanned
	offset       int          // the offset of the character current reading
	readOffset   int          // offset after current reading character
	ch           rune         // current character
	errorHandler ErrorHandler // function to handle error

	pos token.Position // the position for current reading character in the input string

	ErrorCount int
}

// New create a Lexer to scan the input string
func New(input string, errHandler ErrorHandler) *Lexer {
	l := &Lexer{input: []byte(input), pos: token.Position{Line: 1, Column: 0}, errorHandler: errHandler}
	l.readRune()
	return l
}

func (l *Lexer) error(pos token.Position, msg string) {
	if l.errorHandler != nil {
		l.errorHandler(pos, msg)
	}
	l.ErrorCount++
}

// token0 is the default token type for current l.ch
// token1 is the returned token type when l.ch equals to '='
func (l *Lexer) switch2(ch rune, token0, token1 token.TokenType) token.Token {
	if l.ch == '=' {
		tok := newToken(token1, string([]rune{ch, l.ch}))
		l.readRune()
		return tok
	}
	return newToken(token0, string(ch))
}

func (l *Lexer) switch3(ch rune, token0, token1 token.TokenType, ch2 rune, token2 token.TokenType) token.Token {
	if l.ch == '=' {
		tok := newToken(token1, string([]rune{ch, l.ch}))
		l.readRune()
		return tok
	}

	if l.ch == ch2 {
		tok := newToken(token2, string([]rune{ch, l.ch}))
		l.readRune()
		return tok
	}

	return newToken(token0, string(ch))
}

func (l *Lexer) switch4(ch rune, token0, token1 token.TokenType, ch2 rune, token2 token.TokenType, token3 token.TokenType) token.Token {
	if l.ch == '=' {
		tok := newToken(token1, string([]rune{ch, l.ch}))
		l.readRune()
		return tok
	}

	if l.ch == ch2 {
		l.readRune()
		if l.ch == '=' {
			tok := newToken(token3, string([]rune{ch, ch2, l.ch}))
			l.readRune()
			return tok
		}

		tok := newToken(token2, string([]rune{ch, ch2}))
		return tok
	}

	return newToken(token0, string(ch))
}

// NextToken scan and return the next token parsed from the input string
func (l *Lexer) NextToken() (tok token.Token) {
	l.skipWhiteSpaces()

	pos := l.pos
	switch ch := l.ch; {
	case isLetter(ch):
		literal := l.readIdentifier()
		tok = newToken(token.LookupIdent(literal), literal)
	case isDigit(ch):
		tok = newToken(token.INT, l.readInt())
	default:
		l.readRune()
		switch ch {
		case '=':
			tok = l.switch2('=', token.ASSIGN, token.EQ)
		case '!':
			tok = l.switch2('!', token.BANG, token.NOTEQ)
		case '<':
			tok = l.switch4('<', token.LT, token.LTE, '<', token.LSHIFT, token.LSHIFT_ASSIGN)
		case '>':
			tok = l.switch4('>', token.GT, token.GTE, '>', token.RSHIFT, token.RSHIFT_ASSIGN)
		case '*':
			tok = l.switch2('*', token.ASTERISK, token.ASTERISK_ASSIGN)
		case '-':
			tok = l.switch3('-', token.MINUS, token.MINUS_ASSIGN, '-', token.DECREASE)
		case '+':
			tok = l.switch3('+', token.PLUS, token.PLUS_ASSIGN, '+', token.INCREASE)
		case '/':
			tok = l.switch2('/', token.DIVIDE, token.DIVIDE_ASSIGN)
		case '%':
			tok = l.switch2('%', token.REM, token.REM_ASSIGN)
		case '|':
			tok = l.switch3('|', token.OR, token.OR_ASSIGN, '|', token.LOR)
		case '&':
			tok = l.switch3('&', token.AND, token.AND_ASSIGN, '&', token.LAND)
		case '^':
			tok = l.switch2('^', token.XOR, token.XOR_ASSIGN)
		case ';':
			tok = newToken(token.SEMICOLON, ";")
		case '[':
			tok = newToken(token.LBRACKET, "[")
		case ']':
			tok = newToken(token.RBRACKET, "]")
		case '(':
			tok = newToken(token.LPAREN, "(")
		case ')':
			tok = newToken(token.RPAREN, ")")
		case '{':
			tok = newToken(token.LBRACE, "{")
		case '}':
			tok = newToken(token.RBRACE, "}")
		case ',':
			tok = newToken(token.COMMA, ",")
		case ':':
			tok = newToken(token.COLON, ":")
		case '"':
			tok = l.readString()
		case 0:
			tok = newToken(token.EOF, "")
		default:
			l.error(pos, "Unrecognized character")
		}
	}

	tok.Pos = pos
	return tok
}

func (l *Lexer) readString() token.Token {
	pos := l.pos
	var ret []rune
Loop:
	for {
		switch l.ch {
		case 0:
			l.error(pos, "EOF while reading string")
		case '\\':
			l.readRune()
			var nextCh rune
			switch l.ch {
			case 0:
				l.error(pos, "EOF while reading string")
			case 't':
				nextCh = '\t'
			case 'n':
				nextCh = '\n'
			case 'r':
				nextCh = '\r'
			case '\'':
				nextCh = '\''
			case '"':
				nextCh = '"'
			case '\\':
				nextCh = '\\'
			default:
				l.error(l.pos, "Unsupported escape character")
			}

			ret = append(ret, nextCh)
		case '"':
			l.readRune()
			break Loop
		default:
			ret = append(ret, l.ch)
		}
		l.readRune()
	}
	return newToken(token.STRING, string(ret))
}

func (l *Lexer) readRune() {
	size := 1
	if l.readOffset >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readOffset])
		if l.ch >= utf8.RuneSelf {
			l.ch, size = utf8.DecodeRune(l.input[l.readOffset:])

			if l.ch == utf8.RuneError {
				l.error(l.pos, "illegal utf-8 encoding")
			}
		}
	}
	l.offset = l.readOffset
	l.readOffset += size
	l.pos.AddColumn()

	if l.ch == '\n' || l.ch == '\r' {
		l.pos.AddLine()
	}
}

func (l *Lexer) skipWhiteSpaces() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readRune()
	}
}

func newToken(tokenType token.TokenType, literal string) token.Token {
	return token.Token{Type: tokenType, Literal: literal}
}

func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch == '_')
}

func (l *Lexer) readIdentifier() string {
	pos := l.offset
	for isLetter(l.ch) {
		l.readRune()
	}
	return string(l.input[pos:l.offset])
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= utf8.RuneSelf && unicode.IsDigit(ch)
}

func (l *Lexer) readInt() string {
	pos := l.offset
	l.readRune()
	for isDigit(l.ch) {
		l.readRune()
	}

	return string(l.input[pos:l.offset])
}
