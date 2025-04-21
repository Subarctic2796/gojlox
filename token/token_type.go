package token

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

func (t TokenType) String() string {
	switch t {
	case LEFT_PAREN:
		return "LEFT_PAREN"
	case RIGHT_PAREN:
		return "RIGHT_PAREN"
	case LEFT_BRACE:
		return "LEFT_BRACE"
	case RIGHT_BRACE:
		return "RIGHT_BRACE"
	case COMMA:
		return "COMMA"
	case DOT:
		return "DOT"
	case MINUS:
		return "MINUS"
	case MINUS_EQUAL:
		return "MINUS_EQUAL"
	case PLUS:
		return "PLUS"
	case PLUS_EQUAL:
		return "PLUS_EQUAL"
	case SEMICOLON:
		return "SEMICOLON"
	case SLASH:
		return "SLASH"
	case SLASH_EQUAL:
		return "SLASH_EQUAL"
	case STAR:
		return "STAR"
	case STAR_EQUAL:
		return "STAR_EQUAL"
	case BANG:
		return "BANG"
	case BANG_EQUAL:
		return "BANG_EQUAL"
	case EQUAL:
		return "EQUAL"
	case EQUAL_EQUAL:
		return "EQUAL_EQUAL"
	case GREATER:
		return "GREATER"
	case GREATER_EQUAL:
		return "GREATER_EQUAL"
	case LESS:
		return "LESS"
	case LESS_EQUAL:
		return "LESS_EQUAL"
	case IDENTIFIER:
		return "IDENTIFIER"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case AND:
		return "AND"
	case CLASS:
		return "CLASS"
	case ELSE:
		return "ELSE"
	case FALSE:
		return "FALSE"
	case FUN:
		return "FUN"
	case FOR:
		return "FOR"
	case IF:
		return "IF"
	case NIL:
		return "NIL"
	case OR:
		return "OR"
	case PRINT:
		return "PRINT"
	case RETURN:
		return "RETURN"
	case SUPER:
		return "SUPER"
	case THIS:
		return "THIS"
	case TRUE:
		return "TRUE"
	case VAR:
		return "VAR"
	case WHILE:
		return "WHILE"
	case BREAK:
		return "BREAK"
	case STATIC:
		return "STATIC"
	case EOF:
		return "EOF"
	default:
		return "UNKNOWN"
	}
}
