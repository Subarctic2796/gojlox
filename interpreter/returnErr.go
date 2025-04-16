package interpreter

import "errors"

type ReturnErr struct {
	Value any
}

func (re *ReturnErr) Error() string {
	return "Return Error"
}

var BreakErr = errors.New("Break Error")
