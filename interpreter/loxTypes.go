package interpreter

type LoxType byte

const (
	LOX_NONE LoxType = iota
	LOX_PRIMATIVE
	LOX_ARRAY
	LOX_CALLABLE
	LOX_FUNCTION
	LOX_CLASS
	LOX_INSTANCE
	LOX_HASHMAP
)

type LoxObject interface {
	Type() LoxType
}
