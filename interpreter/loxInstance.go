package interpreter

import (
	"fmt"

	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

type LoxInstnace struct {
	Klass  *LoxClass
	Fields map[string]any
}

func NewLoxInstance(klass *LoxClass) *LoxInstnace {
	return &LoxInstnace{klass, make(map[string]any)}
}

func (li *LoxInstnace) String() string {
	return fmt.Sprintf("%s instance", li.Klass.Name)
}

func (li *LoxInstnace) Get(name *token.Token) (any, error) {
	if val, ok := li.Fields[name.Lexeme]; ok {
		return val, nil
	}
	method := li.Klass.FindMethod(name.Lexeme)
	if method != nil {
		return method.Bind(li), nil
	}
	return nil, &errs.RunTimeErr{
		Tok: name,
		Msg: fmt.Sprintf("Undefined property '%s'", name.Lexeme),
	}
}

func (li *LoxInstnace) Set(name *token.Token, val any) {
	li.Fields[name.Lexeme] = val
}
