package interpreter

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/ast"
	"github.com/Subarctic2796/gojlox/token"
)

type LoxCallable interface {
	Call(intprt *Interpreter, args []any) (any, error)
	Arity() int
}

type LoxFn struct {
	Name    string
	Func    *ast.Lambda
	Closure *Env
	IsInit  bool
}

func NewLoxFn(name string, fn *ast.Lambda, closure *Env, isInit bool) *LoxFn {
	return &LoxFn{name, fn, closure, isInit}
}

func (fn *LoxFn) Bind(inst *LoxInstance) *LoxFn {
	env := NewEnv(fn.Closure)
	env.Define("this", inst)
	return &LoxFn{fn.Name, fn.Func, env, fn.IsInit}
}

func (fn *LoxFn) Call(intprt *Interpreter, args []any) (any, error) {
	env := NewEnv(fn.Closure)
	for i, param := range fn.Func.Params {
		env.Define(param.Lexeme, args[i])
	}
	_, err := intprt.executeBlock(fn.Func.Body, env)
	if err != nil {
		retVal := &ReturnErr{}
		if errors.As(err, &retVal) {
			if fn.IsInit {
				return fn.Closure.GetAt(0, "this"), nil
			}
			return retVal.Value, nil
		}
		return nil, err
	}
	if fn.IsInit {
		return fn.Closure.GetAt(0, "this"), nil
	}
	return nil, nil
}

func (fn *LoxFn) Arity() int {
	return len(fn.Func.Params)
}

func (fn *LoxFn) String() string {
	if len(fn.Name) == 0 {
		return "<lambda>"
	}
	return fmt.Sprintf("<fn %s>", fn.Name)
}

type LoxClass struct {
	Name       string
	SuperClass *LoxClass
	Methods    map[string]*LoxFn
}

func NewLoxClass(name string, superclass *LoxClass, methods map[string]*LoxFn) *LoxClass {
	return &LoxClass{name, superclass, methods}
}

func (lc *LoxClass) FindMethod(name string) *LoxFn {
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
		init.Bind(inst).Call(intprt, args)
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
