package token

//go:generate go tool stringer -type=TokenType
type TokenType int

const (
	NONE TokenType = iota

	// Single-character tokens.
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	LSQR
	RSQR

	COMMA
	DOT
	COLON

	// terminators
	SEMICOLON

	// One or two character tokens.
	BANG
	NEQ

	EQ
	EQ_EQ

	GT
	GT_EQ

	LT
	LT_EQ

	PLUS
	PLUS_EQ

	MINUS
	MINUS_EQ

	SLASH
	SLASH_EQ

	STAR
	STAR_EQ

	PERCENT
	PERCENT_EQ

	// Literals.
	IDENTIFIER
	STRING
	NUMBER

	// Keywords.
	AND
	CLASS
	ELSE
	FALSE
	FUN
	FOR
	IF
	NIL
	OR
	STATIC

	PRINT
	RETURN
	SUPER
	THIS
	TRUE
	VAR
	WHILE
	BREAK

	EOF
)

var KEYWORDS = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
	"break":  BREAK,
	"static": STATIC,
}

func LookUpKeyWord(word string) TokenType {
	if keyword, ok := KEYWORDS[word]; ok {
		return keyword
	}
	return IDENTIFIER
}
