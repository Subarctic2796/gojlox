package interpreter

type LoxIterable interface {
	Get(index int) (any, error)
	Set(index int) (any, error)
}
