package interpreter

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/ast"
)

type LoxFn struct {
	Name    string
	Func    *ast.Lambda
	Closure *Env
	IsInit  bool
}

func NewLoxFn(name string, fn *ast.Lambda, closure *Env, isInit bool) *LoxFn {
	return &LoxFn{name, fn, closure, isInit}
}

func (fn *LoxFn) Bind(inst *LoxInstnace) *LoxFn {
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
	return fmt.Sprintf("<fn %s>", fn.Name)
}
