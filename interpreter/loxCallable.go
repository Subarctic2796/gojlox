package interpreter

type LoxCallable interface {
	Call(intprt *Interpreter, args []any) (any, error)
	Arity() int
}
