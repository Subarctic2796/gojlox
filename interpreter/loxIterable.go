package interpreter

type LoxIterable interface {
	IndexGet(index any) (any, error)
	IndexRange(startIndex, stopIndex any) (any, error)
	IndexSet(index any, value any) error
}
