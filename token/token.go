package token

import "fmt"

type Token struct {
	Kind    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func (t Token) String() string {
	return fmt.Sprintf("%s %s %v", t.Kind, t.Lexeme, t.Literal)
}
