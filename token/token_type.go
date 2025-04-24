package token

//go:generate go tool stringer -type=TokenType
type TokenType int

const (
	NONE TokenType = iota

	// Single-character tokens.
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE

	COMMA
	DOT

	// terminators
	SEMICOLON

	// One or two character tokens.
	BANG
	BANG_EQUAL

	EQUAL
	EQUAL_EQUAL

	GREATER
	GREATER_EQUAL

	LESS
	LESS_EQUAL

	PLUS
	PLUS_EQUAL

	MINUS
	MINUS_EQUAL

	SLASH
	SLASH_EQUAL

	STAR
	STAR_EQUAL

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
