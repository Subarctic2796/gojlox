package interpreter

import (
	"fmt"
	"time"
)

type LoxNative interface {
	LoxCallable
	Name() string
}

var NativeFns = []LoxNative{
	&ClockFn{},
	&StringFn{},
	&PrintFn{},
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

func (ClockFn) Name() string {
	return "clock"
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

func (StringFn) Name() string {
	return "string"
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

func (PrintFn) Name() string {
	return "printf"
}
