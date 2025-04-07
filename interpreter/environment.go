package interpreter

import (
	"fmt"

	"github.com/Subarctic2796/gojlox/errs"
	"github.com/Subarctic2796/gojlox/token"
)

type Env struct {
	Values    map[string]any
	Enclosing *Env
}

func NewEnv() *Env {
	return &Env{make(map[string]any), nil}
}

func NewEnvWithEnclosing(enclosing *Env) *Env {
	return &Env{make(map[string]any), enclosing}
}

func (e *Env) Define(name string, val any) {
	e.Values[name] = val
}

func (e *Env) Get(name *token.Token) (any, error) {
	if val, ok := e.Values[name.Lexeme]; ok {
		return val, nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Get(name)
	}
	msg := fmt.Sprintf("Undefined variable '%s'", name.Lexeme)
	return nil, &errs.RunTimeErr{Tok: name, Msg: msg}
}

func (e *Env) Assign(name *token.Token, val any) error {
	if _, ok := e.Values[name.Lexeme]; ok {
		e.Values[name.Lexeme] = val
		return nil
	}
	if e.Enclosing != nil {
		return e.Enclosing.Assign(name, val)
	}
	msg := fmt.Sprintf("Undefined variable '%s'", name.Lexeme)
	return &errs.RunTimeErr{Tok: name, Msg: msg}
}
