package lexer

import (
	"monkey_kd/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	char         byte
}

func New(input string) *Lexer {
	lex := &Lexer{input: input}
	lex.readChar()
	return lex
}

func (lex *Lexer) readChar() {
	if lex.readPosition >= len(lex.input) {
		lex.char = 0
	} else {
		lex.char = lex.input[lex.readPosition]
	}
	lex.position = lex.readPosition
	lex.readPosition += 1
}

func (lex *Lexer) NextToken() token.Token {
	var tok token.Token

	lex.skipWhitespace()

	switch lex.char {
	case '=':
		if lex.peekChar() == '=' {
			char := lex.char
			lex.readChar()
			literal := string(char) + string(lex.char)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, lex.char)
		}
	case '+':
		tok = newToken(token.PLUS, lex.char)
	case '-':
		tok = newToken(token.MINUS, lex.char)
	case '!':
		if lex.peekChar() == '=' {
			char := lex.char
			lex.readChar()
			literal := string(char) + string(lex.char)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, lex.char)
		}
	case '/':
		tok = newToken(token.SLASH, lex.char)
	case '*':
		tok = newToken(token.ASTERISK, lex.char)
	case '<':
		tok = newToken(token.LT, lex.char)
	case '>':
		tok = newToken(token.GT, lex.char)
	case ';':
		tok = newToken(token.SEMICOLON, lex.char)
	case ',':
		tok = newToken(token.COMMA, lex.char)
	case '(':
		tok = newToken(token.LPAREN, lex.char)
	case ')':
		tok = newToken(token.RPAREN, lex.char)
	case '{':
		tok = newToken(token.LBRACE, lex.char)
	case '}':
		tok = newToken(token.RBRACE, lex.char)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(lex.char) {
			tok.Literal = lex.readIdentifier()
			tok.Type = token.LookupIdentifier(tok.Literal)
			return tok
		} else if isDigit(lex.char) {
			tok.Type = token.INT
			tok.Literal = lex.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, lex.char)
		}
	}

	lex.readChar()
	return tok
}

func newToken(tokenType token.TokenType, char byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(char)}
}

func (lex *Lexer) readIdentifier() string {
	position := lex.position
	for isLetter(lex.char) {
		lex.readChar()
	}
	return lex.input[position:lex.position]
}

func isLetter(char byte) bool {
	return 'a' <= char && char <= 'z' || 'A' <= char && char <= 'Z' || char == '_'
}

func (lex *Lexer) readNumber() string {
	position := lex.position
	for isDigit(lex.char) {
		lex.readChar()
	}
	return lex.input[position:lex.position]
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (lex *Lexer) skipWhitespace() {
	for lex.char == ' ' || lex.char == '\t' || lex.char == '\n' || lex.char == '\r' {
		lex.readChar()
	}
}

func (lex *Lexer) peekChar() byte {
	if lex.readPosition >= len(lex.input) {
		return 0
	} else {
		return lex.input[lex.readPosition]
	}
}
