package token

type TokenType string

type Token struct {
	Type TokenType
	Literal string
}

const (
	ILLEGAL = "ILLEGAL"
	EOF= "EOF"
	
	/* Identifiers & literals */
	IDENTIFIER = "IDENTIFIER"
	INT = "INT"

	/* Operators */
	ASSIGN = "="
	PLUS = "+"
	MINUS = "-"
	BANG = "!"
	ASTERISK = "*"
	SLASH = "/"
	LT = "<"
	GT = ">"
	
	/* Delimiters */
	COMMA = ","
	SEMICOLON = ";"
	LPAREN = "("
	RPAREN = ")"
	LBRACE = "{"
	RBRACE = "}"
	
	/* Keywords */
	FUNCTION = "FUNCTION"
	LET = "LET"
	TRUE = "TRUE"
	FALSE = "FALSE"
	IF = "IF"
	ELSE = "ELSE"
	RETURN = "RETURN"

	EQ = "=="
	NOT_EQ = "!="
)

var keywords = map[string]TokenType{
	"fn": FUNCTION,
	"let": LET,
	"true": TRUE,
	"false": FALSE,
	"if": IF,
	"else": ELSE,
	"return": RETURN,
}

func LookupIdentifier(identifier string) TokenType {
		if tok, ok := keywords[identifier]; ok {
		return tok
	}
	return IDENTIFIER
}

