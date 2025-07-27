package token

import "fmt"

type Token struct {
	Kind    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func NewToken(kind TokenType, lexeme string, lit any, line int) Token {
	return Token{kind, lexeme, lit, line}
}

func (t Token) String() string {
	return fmt.Sprintf("%s %s %v", t.Kind, t.Lexeme, t.Literal)
}
