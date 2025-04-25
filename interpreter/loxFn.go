package interpreter

import (
	"fmt"

	"github.com/Subarctic2796/gojlox/ast"
)

type LoxFn interface {
	LoxFnType() ast.FnType
}

type UserFn struct {
	Name    string
	Func    *ast.Lambda
	Closure *Env
}

func NewUserFn(name string, fn *ast.Lambda, closure *Env) *UserFn {
	return &UserFn{name, fn, closure}
}

func (fn *UserFn) Bind(inst *LoxInstance) *UserFn {
	env := NewEnv(fn.Closure)
	env.Define("this", inst)
	return &UserFn{fn.Name, fn.Func, env}
}

func (fn *UserFn) Call(intprt *Interpreter, args []any) (any, error) {
	env := NewEnv(fn.Closure)
	for i, param := range fn.Func.Params {
		env.Define(param.Lexeme, args[i])
	}
	_, err := intprt.executeBlock(fn.Func.Body, env)
	if err != nil {
		switch e := err.(type) {
		case *ReturnErr:
			if fn.Func.Kind == ast.FN_INIT {
				return fn.Closure.GetAt(0, "this"), nil
			}
			return e.Value, nil
		default:
			return nil, err
		}
	}
	if fn.Func.Kind == ast.FN_INIT {
		return fn.Closure.GetAt(0, "this"), nil
	}
	return nil, nil
}

func (fn *UserFn) Arity() int {
	return len(fn.Func.Params)
}

func (fn *UserFn) String() string {
	switch fn.Func.Kind {
	case ast.FN_LAMBDA:
		return "<lambda>"
	case ast.FN_FUNC:
		return fmt.Sprintf("<fn %s>", fn.Name)
	case ast.FN_STATIC:
		return fmt.Sprintf("<static fn %s>", fn.Name)
	case ast.FN_METHOD:
		return fmt.Sprintf("<method fn %s>", fn.Name)
	case ast.FN_INIT:
		return fmt.Sprintf("<init fn %s>", fn.Name)
	case ast.FN_NATIVE:
		panic("[unreachable] can't have a user defined function that is native")
	default:
		panic("[unreachable] can't have a function that is of type none")
	}
}

func (fn *UserFn) LoxFnType() ast.FnType {
	return fn.Func.Kind
}
