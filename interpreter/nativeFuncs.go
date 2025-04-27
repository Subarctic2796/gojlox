package interpreter

import (
	"fmt"
	"strconv"
	"time"
)

type NativeFn interface {
	LoxCallable
}

var NativeFns = map[string]NativeFn{
	"clock":    &ClockFn{},
	"len":      &LenFn{},
	"string":   &StringFn{},
	"printf":   &PrintFn{},
	"parseNum": &ParseNumFn{},
}

type ClockFn struct{}

func (ClockFn) Call(intprt *Interpreter, args []any) (any, error) {
	return float64(time.Now().UnixMilli()) / 1000, nil
}

func (ClockFn) Arity() int {
	return 0
}

func (ClockFn) String() string {
	return "<native fn clock>"
}

type StringFn struct{}

func (StringFn) Call(intprt *Interpreter, args []any) (any, error) {
	return fmt.Sprint(args[0]), nil
}

func (StringFn) Arity() int {
	return 1
}

func (StringFn) String() string {
	return "<native fn string>"
}

type ParseNumFn struct{}

func (ParseNumFn) Call(intprt *Interpreter, args []any) (any, error) {
	if str, ok := args[0].(string); ok {
		num, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return nil, err
		}
		return num, nil
	}
	return nil, fmt.Errorf("argument must be a string")
}

func (ParseNumFn) Arity() int {
	return 1
}

func (ParseNumFn) String() string {
	return "<native fn parseNum>"
}

type PrintFn struct{}

func (PrintFn) Call(intprt *Interpreter, args []any) (any, error) {
	fmt.Println(args...)
	return nil, nil
}

func (PrintFn) Arity() int {
	return -1
}

func (PrintFn) String() string {
	return "<native fn printf>"
}

type LenFn struct{}

func (LenFn) Call(intprt *Interpreter, args []any) (any, error) {
	switch t := args[0].(type) {
	case *LoxArray:
		return float64(len(t.Items)), nil
	case string:
		return float64(len(t)), nil
	case float64:
		return nil, fmt.Errorf("can only use 'len' on iterables: got number")
	default:
		return nil, fmt.Errorf("can only use 'len' on iterables: got %s", t)
	}
}

func (LenFn) Arity() int {
	return 1
}

func (LenFn) String() string {
	return "<native fn len>"
}
