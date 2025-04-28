package interpreter

type LoxIterable interface {
	IndexGet(index any) (any, error)
	IndexSet(index any, value any) error
}
