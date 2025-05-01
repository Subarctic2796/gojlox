package interpreter

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/token"
)

type ReturnErr struct {
	Value any
}

func (re *ReturnErr) Error() string {
	return "Return Error"
}

var (
	BreakErr        = errors.New("Break Error")
	RangeHashMapErr = errors.New("can't use ranges on hashmaps")
)

type RunTimeErr struct {
	Tok *token.Token
	Msg string
}

func (e *RunTimeErr) Error() string {
	return fmt.Sprintf("[RunTimeError]: %s\n[line %d] %s", e.Msg, e.Tok.Line, e.Tok)
}
