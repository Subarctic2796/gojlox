package interpreter

import (
	"errors"
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
	var err error
	if !intprt.useV2 {
		_, err = intprt.executeBlock(fn.Func.Body, env)
	} else {
		_, err = intprt.executeBlock2(fn.Func.Body, env)
	}
	if err != nil {
		retVal := &ReturnErr{}
		if errors.As(err, &retVal) {
			if fn.Func.Kind == ast.FN_INIT {
				return fn.Closure.GetAt(0, "this"), nil
			}
			return retVal.Value, nil
		}
		return nil, err
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
	if len(fn.Name) == 0 {
		return "<lambda>"
	}
	return fmt.Sprintf("<fn %s>", fn.Name)
}

func (fn *UserFn) LoxFnType() ast.FnType {
	return fn.Func.Kind
}
