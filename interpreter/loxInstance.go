package interpreter

import (
	"fmt"
	"hash/fnv"

	"github.com/Subarctic2796/gojlox/token"
)

type LoxInstance struct {
	Klass  *UserClass
	Fields map[string]any
}

func NewLoxInstance(klass *UserClass) *LoxInstance {
	return &LoxInstance{klass, make(map[string]any)}
}

func (li *LoxInstance) String() string {
	return fmt.Sprintf("%s instance", li.Klass.Name)
}

func (li *LoxInstance) Get(name *token.Token) (any, error) {
	if val, ok := li.Fields[name.Lexeme]; ok {
		return val, nil
	}
	method := li.Klass.FindMethod(name.Lexeme)
	if method != nil {
		return method.Bind(li), nil
	}
	return nil, &RunTimeErr{
		Tok: name,
		Msg: fmt.Sprintf("Undefined property '%s'", name.Lexeme),
	}
}

func (li *LoxInstance) Set(name *token.Token, val any) {
	li.Fields[name.Lexeme] = val
}

func (li *LoxInstance) Hash() uint {
	h := fnv.New64a()
	h.Write([]byte(li.Klass.Name))
	for k := range li.Klass.Methods {
		h.Write([]byte(k))
	}
	for k, v := range li.Fields {
		h.Write([]byte(k))
		h.Write(fmt.Append(nil, v))
	}
	return uint(h.Sum64())
}
