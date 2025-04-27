package interpreter

import "fmt"

type LoxArray struct {
	Items []any
}

func (la *LoxArray) String() string {
	return fmt.Sprint(la.Items)
}

func (la *LoxArray) Iterable() {}
