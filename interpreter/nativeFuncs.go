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
	"push":     &ArrPushFn{},
	"delete":   &HashDelKeyFn{},
}

type ClockFn struct{}

// func (ClockFn) Call(intprt *Interpreter, args []any) (any, error) {
func (ClockFn) Call(args ...any) (any, error) {
	return float64(time.Now().UnixMilli()) / 1000, nil
}

func (ClockFn) Arity() int {
	return 0
}

func (ClockFn) String() string {
	return "<native fn clock>"
}

type StringFn struct{}

// func (StringFn) Call(intprt *Interpreter, args []any) (any, error) {
func (StringFn) Call(args ...any) (any, error) {
	return fmt.Sprint(args[1]), nil
}

func (StringFn) Arity() int {
	return 1
}

func (StringFn) String() string {
	return "<native fn string>"
}

type ParseNumFn struct{}

// func (ParseNumFn) Call(intprt *Interpreter, args []any) (any, error) {
func (ParseNumFn) Call(args ...any) (any, error) {
	if str, ok := args[1].(string); ok {
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

// func (PrintFn) Call(intprt *Interpreter, args []any) (any, error) {
func (PrintFn) Call(args ...any) (any, error) {
	fmt.Println(args[1:]...)
	return nil, nil
}

func (PrintFn) Arity() int {
	return -1
}

func (PrintFn) String() string {
	return "<native fn printf>"
}

type LenFn struct{}

// func (LenFn) Call(intprt *Interpreter, args []any) (any, error) {
func (LenFn) Call(args ...any) (any, error) {
	switch t := args[1].(type) {
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

type ArrPushFn struct{}

func (ArrPushFn) Call(args ...any) (any, error) {
	switch arr := args[1].(type) {
	case *LoxArray:
		arr.Items = append(arr.Items, args[2])
		return arr, nil
	default:
		return nil, fmt.Errorf("can only use 'push' on arrays: got '%s'", arr)
	}
}

func (ArrPushFn) Arity() int {
	return 2
}

func (ArrPushFn) String() string {
	return "<native fn push>"
}

type HashDelKeyFn struct{}

func (HashDelKeyFn) Call(args ...any) (any, error) {
	switch hm := args[1].(type) {
	case *LoxHashMap:
		key, err := hm.hashObj(args[2])
		if err != nil {
			return nil, err
		}
		delete(hm.Pairs, key)
		return nil, nil
	default:
		return nil, fmt.Errorf("can only use 'push' on arrays: got '%s'", hm)
	}
}

func (HashDelKeyFn) Arity() int {
	return 2
}

func (HashDelKeyFn) String() string {
	return "<native fn delete>"
}
