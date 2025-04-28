package interpreter

type LoxCallable interface {
	// first arg is implicitly `*Interpreter`
	// so make sure to ignore it if needed
	Call(args ...any) (any, error)
	Arity() int
}
