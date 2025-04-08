package interpreter

import "time"

type ClockFn struct{}

func (ClockFn) Call(intprt *Interpreter, args []any) (any, error) {
	return float64(time.Now().UnixMilli()) / 1000, nil
}

func (ClockFn) Arity() int {
	return 0
}

func (ClockFn) String() string {
	return "<native fn>"
}
