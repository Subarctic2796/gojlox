package interpreter

import (
	"fmt"

	"github.com/Subarctic2796/gojlox/token"
)

type LoxCallable interface {
	Call(intprt *Interpreter, args []any) (any, error)
	Arity() int
}

type LoxClass struct {
	Name       string
	SuperClass *LoxClass
	Methods    map[string]*UserFn
}

func NewLoxClass(name string, superclass *LoxClass, methods map[string]*UserFn) *LoxClass {
	return &LoxClass{name, superclass, methods}
}

func (lc *LoxClass) FindMethod(name string) *UserFn {
	if val, ok := lc.Methods[name]; ok {
		return val
	}
	if lc.SuperClass != nil {
		return lc.SuperClass.FindMethod(name)
	}
	return nil
}

func (lc *LoxClass) String() string {
	return fmt.Sprint(lc.Name)
}

func (lc *LoxClass) Call(intprt *Interpreter, args []any) (any, error) {
	inst := NewLoxInstance(lc)
	init := lc.FindMethod("init")
	if init != nil {
		_, _ = init.Bind(inst).Call(intprt, args)
	}
	return inst, nil
}

func (lc *LoxClass) Arity() int {
	init := lc.FindMethod("init")
	if init == nil {
		return 0
	}
	return init.Arity()
}

type LoxInstance struct {
	Klass  *LoxClass
	Fields map[string]any
}

func NewLoxInstance(klass *LoxClass) *LoxInstance {
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
