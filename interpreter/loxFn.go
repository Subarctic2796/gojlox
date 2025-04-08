package interpreter

import (
	"errors"
	"fmt"

	"github.com/Subarctic2796/gojlox/ast"
)

type LoxFn struct {
	Func    *ast.Function
	Closure *Env
}

func NewLoxFn(fn *ast.Function, closure *Env) *LoxFn {
	return &LoxFn{fn, closure}
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
			return retVal.Value, nil
		}
		return nil, err
	}
	return nil, nil
}

func (fn *LoxFn) Arity() int {
	return len(fn.Func.Params)
}

func (fn LoxFn) String() string {
	return fmt.Sprintf("<fn %s>", fn.Func.Name.Lexeme)
}
