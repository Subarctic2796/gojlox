package interpreter

import (
	"fmt"

	"github.com/Subarctic2796/gojlox/token"
)

type Env struct {
	Enclosing *Env
	Values    map[string]any
}

func NewEnv(enclosing *Env) *Env {
	return &Env{enclosing, make(map[string]any)}
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
	return nil, &RunTimeErr{Tok: name, Msg: msg}
}

func (e *Env) GetAt(dist int, name string) any {
	return e.Ancestor(dist).Values[name]
}

func (e *Env) Ancestor(dist int) *Env {
	// goes upward
	env := e
	for range dist {
		env = env.Enclosing
	}
	return env
}

func (e *Env) Assign(name *token.Token, val any) error {
	if _, ok := e.Values[name.Lexeme]; ok {
		e.Values[name.Lexeme] = val
		return nil
	}
	// if 'a' in global and it gets updated in a local scope
	// then we have to update it in the correct scope
	if e.Enclosing != nil {
		return e.Enclosing.Assign(name, val)
	}
	msg := fmt.Sprintf("Undefined variable '%s'", name.Lexeme)
	return &RunTimeErr{Tok: name, Msg: msg}
}

func (e *Env) AssignAt(dist int, name *token.Token, val any) {
	e.Ancestor(dist).Values[name.Lexeme] = val
}
