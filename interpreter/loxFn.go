package interpreter

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/ast"
)

type LoxFn struct {
	Func    *ast.Function
	Closure *Env
	IsInit  bool
}

func NewLoxFn(fn *ast.Function, closure *Env, isInit bool) *LoxFn {
	return &LoxFn{fn, closure, isInit}
}

func (fn *LoxFn) Bind(inst *LoxInstnace) *LoxFn {
	env := NewEnvWithEnclosing(fn.Closure)
	env.Define("this", inst)
	return &LoxFn{fn.Func, env, fn.IsInit}
}

func (fn *LoxFn) Call(intprt *Interpreter, args []any) (any, error) {
	env := NewEnvWithEnclosing(fn.Closure)
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

func (fn LoxFn) String() string {
	return fmt.Sprintf("<fn %s>", fn.Func.Name.Lexeme)
}
